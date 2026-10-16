package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/hotkey"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

// speech is one second of audio that is not silence.
var speech = func() []byte {
	b := make([]byte, 32000)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}()

type fakeRecorder struct {
	pcm       []byte
	startErr  error
	cancelled bool
}

// Start hands the whole recording over as one chunk, as if it was spoken
// at once.
func (r *fakeRecorder) Start(onChunk func([]byte)) error {
	if r.startErr == nil {
		onChunk(r.pcm)
	}
	return r.startErr
}
func (r *fakeRecorder) Stop() ([]byte, error) { return r.pcm, nil }
func (r *fakeRecorder) Cancel()               { r.cancelled = true }

type fakeASR struct {
	text  string
	err   error
	block chan struct{}
	delay time.Duration
	calls atomic.Int32
}

func (a *fakeASR) Transcribe(ctx context.Context, pcm []byte) (string, error) {
	a.calls.Add(1)
	time.Sleep(a.delay)
	if a.block != nil {
		select {
		case <-a.block:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if len(pcm) != len(speech) {
		return "", fmt.Errorf("recording of %d bytes", len(pcm))
	}
	return a.text, a.err
}

type fakeLLM struct {
	mu  sync.Mutex
	got []prompt.Request
	out func(prompt.Request) (string, error)
}

func (l *fakeLLM) Rewrite(ctx context.Context, r prompt.Request) (string, error) {
	l.mu.Lock()
	l.got = append(l.got, r)
	l.mu.Unlock()
	return l.out(r)
}

type fakePaster struct {
	mu   sync.Mutex
	text []string
}

func (p *fakePaster) Paste(text string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.text = append(p.text, text)
	return nil
}

func (p *fakePaster) pasted() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.text...)
}

type harness struct {
	c       *Controller
	rec     *fakeRecorder
	asr     *fakeASR
	llm     *fakeLLM
	paster  *fakePaster
	changes chan Status
	capture chan [2]bool
	picked  chan string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h := &harness{
		rec:    &fakeRecorder{pcm: speech},
		asr:    &fakeASR{text: "嗯我们周三，不对，周四开会"},
		llm:    &fakeLLM{out: func(r prompt.Request) (string, error) { return "我们周四开会。", nil }},
		paster: &fakePaster{},
		// Big enough that the controller never waits on the test.
		changes: make(chan Status, 64),
		capture: make(chan [2]bool, 64),
		picked:  make(chan string, 64),
	}
	h.c = New(ctx, Deps{
		Recorder:       h.rec,
		Transcriber:    h.asr,
		Rewriter:       h.llm,
		Paster:         h.paster,
		Settings:       store.Defaults,
		Capture:        func(esc, tab bool) { h.capture <- [2]bool{esc, tab} },
		OnChange:       func(s Status) { h.changes <- s },
		OnTargetPicked: func(target string) { h.picked <- target },
	})
	return h
}

// until waits for a status in the given state and returns it.
func (h *harness) until(t *testing.T, state State) Status {
	t.Helper()
	timeout := time.After(2 * time.Second)
	for {
		select {
		case s := <-h.changes:
			if s.State == state {
				return s
			}
		case <-timeout:
			t.Fatalf("no %v status; now %+v", state, h.c.Status())
		}
	}
}

func TestDictation(t *testing.T) {
	h := newHarness(t)
	h.asr.delay = 5 * time.Millisecond
	h.c.Handle(hotkey.Dictate)
	if s := h.until(t, Recording); s.Mode != prompt.Dictate {
		t.Fatalf("recording %+v", s)
	}
	if got := <-h.capture; got != [2]bool{true, false} {
		t.Fatalf("capture while dictating %v", got)
	}
	h.c.Handle(hotkey.Dictate)
	h.until(t, Transcribing)
	h.until(t, Rewriting)
	s := h.until(t, Idle)
	if s.Text != "我们周四开会。" || s.Note != NoteNone {
		t.Fatalf("done %+v", s)
	}
	if s.Timings.Transcribe < 5*time.Millisecond || s.Timings.Rewrite <= 0 {
		t.Fatalf("timings %+v", s.Timings)
	}
	if got := h.paster.pasted(); len(got) != 1 || got[0] != "我们周四开会。" {
		t.Fatalf("pasted %q", got)
	}
	if r := h.llm.got[0]; r.Mode != prompt.Dictate || r.Transcript != "嗯我们周三，不对，周四开会" {
		t.Fatalf("rewrite request %+v", r)
	}
}

func TestTranslationFollowsRules(t *testing.T) {
	for _, tc := range []struct{ spoken, target string }{
		{"我们明天发布新版本", "en"},
		{"We ship the new version tomorrow", "zh"},
	} {
		h := newHarness(t)
		h.asr.text = tc.spoken
		h.llm.out = func(prompt.Request) (string, error) { return "译文", nil }
		h.c.Handle(hotkey.Translate)
		if s := h.until(t, Recording); s.Mode != prompt.Translate || s.Target != "en" {
			t.Fatalf("recording %+v", s)
		}
		h.c.Handle(hotkey.Dictate) // Fn finishes either mode
		h.until(t, Idle)
		if r := h.llm.got[0]; r.Target != tc.target {
			t.Errorf("%q translated into %q, want %q", tc.spoken, r.Target, tc.target)
		}
	}
}

func TestTabPicksTargetOverRules(t *testing.T) {
	h := newHarness(t)
	h.asr.text = "我们明天发布新版本"
	h.llm.out = func(prompt.Request) (string, error) { return "明日リリースします", nil }
	h.c.Handle(hotkey.Translate)
	h.until(t, Recording)
	h.c.Handle(hotkey.CycleTarget) // en → zh
	h.c.Handle(hotkey.CycleTarget) // zh → ja
	h.c.Handle(hotkey.Translate)
	h.until(t, Idle)
	if r := h.llm.got[0]; r.Target != "ja" {
		t.Fatalf("target %q, want the one picked", r.Target)
	}
	if first, second := <-h.picked, <-h.picked; first != "zh" || second != "ja" {
		t.Fatalf("picked %q then %q", first, second)
	}
}

func TestTabIgnoredWhileDictating(t *testing.T) {
	h := newHarness(t)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.CycleTarget)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Idle)
	if r := h.llm.got[0]; r.Mode != prompt.Dictate || r.Target != "" {
		t.Fatalf("request %+v", r)
	}
}

func TestEscWhileRecording(t *testing.T) {
	h := newHarness(t)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Cancel)
	if s := h.until(t, Idle); s.Note != NoteCancelled {
		t.Fatalf("status %+v", s)
	}
	if !h.rec.cancelled {
		t.Fatal("recorder not cancelled")
	}
	if len(h.paster.pasted()) != 0 {
		t.Fatal("pasted after Esc")
	}
}

func TestEscWhileTranscribingPastesNothing(t *testing.T) {
	h := newHarness(t)
	h.asr.block = make(chan struct{})
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Transcribing)
	h.c.Handle(hotkey.Cancel)
	if s := h.until(t, Idle); s.Note != NoteCancelled {
		t.Fatalf("status %+v", s)
	}
	// The cancelled session's transcript arrives late and is ignored; a new
	// session works.
	close(h.asr.block)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Idle)
	if got := h.paster.pasted(); len(got) != 1 {
		t.Fatalf("pasted %q", got)
	}
}

func TestRewriteFailurePastesTranscript(t *testing.T) {
	h := newHarness(t)
	h.llm.out = func(prompt.Request) (string, error) { return "", errors.New("502") }
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	s := h.until(t, Idle)
	if s.Note != NoteRawPasted || s.Text != h.asr.text {
		t.Fatalf("status %+v", s)
	}
}

func TestRewriteThatAnswersIsRejected(t *testing.T) {
	h := newHarness(t)
	h.asr.text = "帮我写一封请假邮件"
	h.llm.out = func(prompt.Request) (string, error) {
		return "尊敬的领导：您好！因家中有事需要处理，我想在本周五请假一天，工作已与同事交接妥当，期间如有急事可随时电话联系我。望批准。此致敬礼。", nil
	}
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	if s := h.until(t, Idle); s.Text != "帮我写一封请假邮件" {
		t.Fatalf("pasted %q", s.Text)
	}
}

func TestShortAndSilentRecordings(t *testing.T) {
	for _, tc := range []struct {
		pcm  []byte
		want Note
	}{
		{make([]byte, 3200), NoteTooShort},
		{make([]byte, 32000), NoteSilent},
	} {
		h := newHarness(t)
		h.rec.pcm = tc.pcm
		h.c.Handle(hotkey.Dictate)
		h.until(t, Recording)
		h.c.Handle(hotkey.Dictate)
		if s := h.until(t, Idle); s.Note != tc.want {
			t.Errorf("note %v, want %v", s.Note, tc.want)
		}
	}
}

func TestNoKey(t *testing.T) {
	h := newHarness(t)
	h.asr.err = fmt.Errorf("transcribe: %w", ErrNoKey)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	if s := h.until(t, Idle); s.Note != NoteNoKey {
		t.Fatalf("note %v", s.Note)
	}
}

func TestMicFailure(t *testing.T) {
	h := newHarness(t)
	h.rec.startErr = errors.New("no input device")
	h.c.Handle(hotkey.Dictate)
	if s := h.until(t, Idle); s.Note != NoteMicFailed || s.Err == nil || s.Err.Error() != "no input device" {
		t.Fatalf("status %+v", s)
	}
}

type fakeStream struct {
	text    string
	err     error
	mu      sync.Mutex
	written int
	closed  bool
}

func (s *fakeStream) Write(pcm []byte) {
	s.mu.Lock()
	s.written += len(pcm)
	s.mu.Unlock()
}

func (s *fakeStream) Finish(ctx context.Context) (string, error) { return s.text, s.err }

func (s *fakeStream) Close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}

type fakeStreamer struct{ stream *fakeStream }

func (f fakeStreamer) Open(context.Context) Stream {
	if f.stream == nil {
		return nil
	}
	return f.stream
}

func newStreamingHarness(t *testing.T, stream *fakeStream) *harness {
	t.Helper()
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h.c = New(ctx, Deps{
		Recorder:    h.rec,
		Streamer:    fakeStreamer{stream},
		Transcriber: h.asr,
		Rewriter:    h.llm,
		Paster:      h.paster,
		Settings:    store.Defaults,
		OnChange:    func(s Status) { h.changes <- s },
	})
	return h
}

func TestStreamingDictation(t *testing.T) {
	stream := &fakeStream{text: "嗯我们周四开会"}
	h := newStreamingHarness(t, stream)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	s := h.until(t, Idle)
	if s.Note != NoteNone || s.Text != "我们周四开会。" {
		t.Fatalf("status %+v", s)
	}
	if n := h.asr.calls.Load(); n != 0 {
		t.Fatalf("the whole recording was transcribed too (%d calls)", n)
	}
	if h.llm.got[0].Transcript != "嗯我们周四开会" {
		t.Fatalf("rewrote %q", h.llm.got[0].Transcript)
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if stream.written != len(speech) {
		t.Fatalf("streamed %d bytes, recorded %d", stream.written, len(speech))
	}
}

func TestStreamFailureFallsBack(t *testing.T) {
	h := newStreamingHarness(t, &fakeStream{err: errors.New("soniox: connect: refused")})
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	s := h.until(t, Idle)
	if s.Note != NoteStreamFallback || s.Text != "我们周四开会。" {
		t.Fatalf("status %+v", s)
	}
	if n := h.asr.calls.Load(); n != 1 {
		t.Fatalf("whole-recording transcriptions %d, want 1", n)
	}
}

func TestEscClosesStream(t *testing.T) {
	stream := &fakeStream{text: "x"}
	h := newStreamingHarness(t, stream)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Cancel)
	h.until(t, Idle)
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if !stream.closed {
		t.Fatal("Esc left the stream open")
	}
}

func TestStreamingOffUsesTranscriber(t *testing.T) {
	h := newStreamingHarness(t, nil)
	h.c.Handle(hotkey.Dictate)
	h.until(t, Recording)
	h.c.Handle(hotkey.Dictate)
	if s := h.until(t, Idle); s.Note != NoteNone {
		t.Fatalf("status %+v", s)
	}
	if n := h.asr.calls.Load(); n != 1 {
		t.Fatalf("transcriptions %d, want 1", n)
	}
}
