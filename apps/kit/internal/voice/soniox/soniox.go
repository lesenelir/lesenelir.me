// Package soniox streams a recording to Soniox's real-time speech-to-text
// while it is being made, so that the text is all but ready when the user
// stops talking. See https://soniox.com/docs/stt/api-reference/websocket-api.
package soniox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// DefaultURL is Soniox's real-time endpoint.
const DefaultURL = "wss://stt-rt.soniox.com/transcribe-websocket"

// DefaultModel is the real-time model the settings start with.
const DefaultModel = "stt-rt-v5"

// ErrNoKey is returned when no API key is set.
var ErrNoKey = errors.New("soniox: no API key")

// Config is one session's setup.
type Config struct {
	Model string
	// LanguageHints are ISO 639-1 codes the speech is likely in.
	LanguageHints []string
	// Terms are names and words to spell as written.
	Terms []string
}

// Client opens sessions. It is safe for concurrent use.
type Client struct {
	URL string
	// Key returns the API key; it is read for every session.
	Key func() string
	// HTTP dials the connection; nil is http.DefaultClient.
	HTTP *http.Client
}

// Error is an error Soniox reported.
type Error struct {
	Code    int    `json:"error_code"`
	Type    string `json:"error_type"`
	Message string `json:"error_message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("soniox: %d %s: %s", e.Code, e.Type, e.Message)
}

// Session is one recording streamed to Soniox. Write sends audio as it
// comes; Finish waits for the text; Close drops it. Its methods are safe
// for concurrent use.
type Session struct {
	cancel context.CancelFunc
	audio  chan []byte

	mu       sync.Mutex
	finished bool // no more audio
	text     strings.Builder

	done chan struct{} // closed once the text is final or failed
	err  error         // set before done closes
}

// audioQueue holds the audio recorded while the connection is being made:
// 100 ms chunks, so a few minutes of them.
const audioQueue = 4096

// Open starts a session. It returns at once: the connection is made in the
// background, and audio written before it is ready waits in a queue. The
// session ends when ctx is done.
func (c *Client) Open(ctx context.Context, cfg Config) *Session {
	ctx, cancel := context.WithCancel(ctx)
	s := &Session{cancel: cancel, audio: make(chan []byte, audioQueue), done: make(chan struct{})}
	go s.run(ctx, c, cfg)
	return s
}

// Write queues 16 kHz mono 16-bit PCM. It never blocks, so the recorder's
// callback may call it; audio after Finish or past a full queue is dropped.
func (s *Session) Write(pcm []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished {
		return
	}
	select {
	case s.audio <- pcm:
	default:
	}
}

// Finish ends the audio and returns the transcript once Soniox has
// finalized it.
func (s *Session) Finish(ctx context.Context) (string, error) {
	s.mu.Lock()
	if !s.finished {
		s.finished = true
		close(s.audio)
	}
	s.mu.Unlock()
	select {
	case <-s.done:
	case <-ctx.Done():
		s.Close()
		return "", ctx.Err()
	}
	if s.err != nil {
		return "", s.err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.TrimSpace(s.text.String()), nil
}

// Close drops the session.
func (s *Session) Close() {
	s.mu.Lock()
	if !s.finished {
		s.finished = true
		close(s.audio)
	}
	s.mu.Unlock()
	s.cancel()
}

func (s *Session) fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.done:
	default:
		s.err = err
		close(s.done)
	}
}

func (s *Session) succeed() { s.fail(nil) }

type startRequest struct {
	Model         string          `json:"model"`
	AudioFormat   string          `json:"audio_format"`
	SampleRate    int             `json:"sample_rate"`
	NumChannels   int             `json:"num_channels"`
	LanguageHints []string        `json:"language_hints,omitempty"`
	Context       *sessionContext `json:"context,omitempty"`
}

// sessionContext is what Soniox knows of the speech before it hears it.
type sessionContext struct {
	Terms []string `json:"terms,omitempty"`
}

type token struct {
	Text    string `json:"text"`
	IsFinal bool   `json:"is_final"`
}

type response struct {
	Tokens   []token `json:"tokens"`
	Finished bool    `json:"finished"`
	Error
}

// finMarker is the token that ends a finalization.
const finMarker = "<fin>"

func (s *Session) run(ctx context.Context, c *Client, cfg Config) {
	defer s.cancel()
	key := ""
	if c.Key != nil {
		key = c.Key()
	}
	if key == "" {
		s.fail(ErrNoKey)
		return
	}
	url := c.URL
	if url == "" {
		url = DefaultURL
	}
	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
		HTTPClient: c.HTTP,
		HTTPHeader: http.Header{"Authorization": {"Bearer " + key}},
	})
	if err != nil {
		s.fail(fmt.Errorf("soniox: connect: %w", err))
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1 << 20)

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}
	start := startRequest{
		Model:         model,
		AudioFormat:   "s16le",
		SampleRate:    16000,
		NumChannels:   1,
		LanguageHints: cfg.LanguageHints,
	}
	if len(cfg.Terms) > 0 {
		start.Context = &sessionContext{Terms: cfg.Terms}
	}
	b, _ := json.Marshal(start)
	if err := conn.Write(ctx, websocket.MessageText, b); err != nil {
		s.fail(fmt.Errorf("soniox: start: %w", err))
		return
	}

	go s.read(ctx, conn)

	for pcm := range s.audio {
		if err := conn.Write(ctx, websocket.MessageBinary, pcm); err != nil {
			s.fail(fmt.Errorf("soniox: send audio: %w", err))
			return
		}
	}
	// The audio has ended: have Soniox finalize what it holds now rather
	// than wait for more, then end the stream.
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"finalize"}`)); err != nil {
		s.fail(fmt.Errorf("soniox: finalize: %w", err))
		return
	}
	if err := conn.Write(ctx, websocket.MessageText, []byte{}); err != nil {
		s.fail(fmt.Errorf("soniox: end: %w", err))
		return
	}
	// Wait for the reader, which ends the session.
	select {
	case <-s.done:
	case <-ctx.Done():
		s.fail(ctx.Err())
	case <-time.After(30 * time.Second):
		s.fail(errors.New("soniox: no final transcript"))
	}
	conn.Close(websocket.StatusNormalClosure, "")
}

func (s *Session) read(ctx context.Context, conn *websocket.Conn) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			s.fail(fmt.Errorf("soniox: read: %w", err))
			return
		}
		var res response
		if err := json.Unmarshal(data, &res); err != nil {
			s.fail(fmt.Errorf("soniox: bad response: %w", err))
			return
		}
		if res.Code != 0 {
			e := res.Error
			s.fail(&e)
			return
		}
		fin := false
		s.mu.Lock()
		for _, t := range res.Tokens {
			switch {
			case t.Text == finMarker:
				fin = true
			case t.IsFinal:
				s.text.WriteString(t.Text)
			}
		}
		s.mu.Unlock()
		if fin || res.Finished {
			s.succeed()
			return
		}
	}
}
