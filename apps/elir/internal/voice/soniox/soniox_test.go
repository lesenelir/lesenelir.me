package soniox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// fakeSoniox answers as Soniox does: it reads the start request and the
// audio, sends a non-final token while audio comes, and on finalize the
// final tokens and <fin>.
type fakeSoniox struct {
	t       *testing.T
	delay   time.Duration // before accepting, as a slow handshake
	reply   []token       // final tokens to send on finalize
	failure string        // an error response to send instead

	mu    sync.Mutex
	start startRequest
	audio int
	auth  string
}

func (f *fakeSoniox) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(f.delay)
	f.mu.Lock()
	f.auth = r.Header.Get("Authorization")
	f.mu.Unlock()
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		f.t.Error(err)
		return
	}
	defer conn.CloseNow()
	ctx := r.Context()
	_, data, err := conn.Read(ctx)
	if err != nil {
		return
	}
	f.mu.Lock()
	_ = json.Unmarshal(data, &f.start)
	f.mu.Unlock()
	if f.failure != "" {
		_ = conn.Write(ctx, websocket.MessageText, []byte(f.failure))
		return
	}
	sentPartial := false
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if typ == websocket.MessageBinary {
			f.mu.Lock()
			f.audio += len(data)
			f.mu.Unlock()
			if !sentPartial {
				sentPartial = true
				_ = conn.Write(ctx, websocket.MessageText, []byte(`{"tokens":[{"text":"我们","is_final":false}]}`))
			}
			continue
		}
		if string(data) == `{"type":"finalize"}` {
			toks := append(append([]token{}, f.reply...), token{Text: finMarker, IsFinal: true})
			b, _ := json.Marshal(response{Tokens: toks})
			_ = conn.Write(ctx, websocket.MessageText, b)
		}
	}
}

func newServer(t *testing.T, f *fakeSoniox) *Client {
	t.Helper()
	f.t = t
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return &Client{URL: "ws" + strings.TrimPrefix(srv.URL, "http"), Key: func() string { return "sk-soniox" }}
}

func TestStream(t *testing.T) {
	f := &fakeSoniox{
		delay: 100 * time.Millisecond,
		reply: []token{{Text: "我们", IsFinal: true}, {Text: "周四开会", IsFinal: true}, {Text: " with the React team", IsFinal: true}},
	}
	c := newServer(t, f)
	s := c.Open(context.Background(), Config{Model: "stt-rt-v5", LanguageHints: []string{"zh", "en"}, Terms: []string{"MyGo"}})
	// Written at once, before the slow handshake ends: they wait in the
	// queue rather than being lost.
	for range 10 {
		s.Write(make([]byte, 3200))
	}
	text, err := s.Finish(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if text != "我们周四开会 with the React team" {
		t.Fatalf("text %q", text)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.auth != "Bearer sk-soniox" {
		t.Errorf("Authorization %q", f.auth)
	}
	if f.audio != 32000 {
		t.Errorf("server got %d bytes of audio, want 32000", f.audio)
	}
	st := f.start
	if st.Model != "stt-rt-v5" || st.AudioFormat != "s16le" || st.SampleRate != 16000 || st.NumChannels != 1 {
		t.Errorf("start %+v", st)
	}
	if strings.Join(st.LanguageHints, ",") != "zh,en" || st.Context == nil || st.Context.Terms[0] != "MyGo" {
		t.Errorf("hints %v context %+v", st.LanguageHints, st.Context)
	}
}

func TestStreamError(t *testing.T) {
	c := newServer(t, &fakeSoniox{
		failure: `{"tokens":[],"error_code":401,"error_type":"unauthenticated","error_message":"bad key"}`,
	})
	s := c.Open(context.Background(), Config{})
	s.Write(make([]byte, 3200))
	_, err := s.Finish(context.Background())
	var e *Error
	if !errors.As(err, &e) || e.Code != 401 || e.Type != "unauthenticated" {
		t.Fatalf("err = %v", err)
	}
}

func TestNoKey(t *testing.T) {
	c := &Client{Key: func() string { return "" }}
	if _, err := c.Open(context.Background(), Config{}).Finish(context.Background()); !errors.Is(err, ErrNoKey) {
		t.Fatalf("err = %v", err)
	}
}

func TestUnreachable(t *testing.T) {
	c := &Client{URL: "ws://127.0.0.1:1/transcribe-websocket", Key: func() string { return "k" }}
	s := c.Open(context.Background(), Config{})
	s.Write(make([]byte, 3200))
	if _, err := s.Finish(context.Background()); err == nil {
		t.Fatal("no error for an unreachable server")
	}
}

func TestCloseDropsTheSession(t *testing.T) {
	c := newServer(t, &fakeSoniox{reply: []token{{Text: "x", IsFinal: true}}})
	s := c.Open(context.Background(), Config{})
	s.Write(make([]byte, 3200))
	s.Close()
	s.Write(make([]byte, 3200)) // after Close: dropped, no panic
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := s.Finish(ctx); err == nil {
		t.Fatal("a closed session finished")
	}
}
