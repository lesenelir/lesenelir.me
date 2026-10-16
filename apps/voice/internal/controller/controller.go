// Package controller runs a dictation or translation from the hotkey to the
// pasted text: record, transcribe, rewrite, paste. One goroutine owns the
// state and handles hotkey actions and pipeline results in turn, so that
// Esc can cancel any stage and nothing is pasted after it.
package controller

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/audio"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/hotkey"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

// State is the stage of the current session.
type State int

const (
	Idle State = iota
	Recording
	Transcribing
	Rewriting
	Pasting
)

func (s State) String() string {
	return [...]string{"idle", "recording", "transcribing", "rewriting", "pasting"}[s]
}

// MinRecording is the shortest recording worth sending: shorter ones are
// taps of Fn that changed the user's mind.
const MinRecording = 300 * time.Millisecond

// Recorder captures the microphone as 16 kHz mono 16-bit PCM.
type Recorder interface {
	Start() error
	// Stop ends the recording and returns it.
	Stop() ([]byte, error)
	// Cancel ends the recording and drops it.
	Cancel()
}

// Transcriber turns a recording, 16 kHz mono 16-bit PCM, into text.
type Transcriber interface {
	Transcribe(ctx context.Context, pcm []byte) (string, error)
}

// Rewriter cleans up or translates a transcript.
type Rewriter interface {
	Rewrite(ctx context.Context, r prompt.Request) (string, error)
}

// Paster puts text where the user's cursor is.
type Paster interface {
	Paste(text string) error
}

// Deps are what the controller drives.
type Deps struct {
	Recorder    Recorder
	Transcriber Transcriber
	Rewriter    Rewriter
	Paster      Paster
	Settings    func() store.Settings
	// Capture tells the hotkey detector which keys to take: Esc while a
	// session can be cancelled, Tab while a translation records.
	Capture func(esc, tab bool)
	// OnChange receives every new status, on the controller's goroutine.
	OnChange func(Status)
	// OnTargetPicked receives a target chosen with Tab or on the voice bar.
	OnTargetPicked func(target string)
}

// Status is what the voice bar and the settings show.
type Status struct {
	State State
	Mode  prompt.Mode
	// Target is the translation target code while translating.
	Target string
	// Note tells what went wrong or what was done instead, NoteNone when
	// all went to plan; Err is the error behind it, if any.
	Note Note
	Err  error
	// Text is the last text pasted.
	Text string
	// Since is when the current stage began.
	Since time.Time
	// Timings are how long the last session's requests took.
	Timings Timings
}

// Timings are the durations of a session's stages after recording.
type Timings struct {
	Transcribe, Rewrite time.Duration
}

// Note is what the voice bar tells the user after a session. The UI words
// it in the app's language.
type Note int

const (
	NoteNone Note = iota
	NoteTooShort
	NoteSilent
	NoteCancelled
	NoteNothing
	NoteRawPasted
	NoteNoKey
	NoteASRFailed
	NotePasteFailed
	NoteMicFailed
)

// Controller is safe for concurrent use.
type Controller struct {
	d     Deps
	inbox chan any

	mu     sync.Mutex
	status Status

	// Owned by the run goroutine.
	session int
	cancel  context.CancelFunc
	manual  bool
	timings Timings
}

type action struct {
	a      hotkey.Action
	target string // for setTarget
}

type transcribed struct {
	session int
	text    string
	err     error
}

type rewritten struct {
	session int
	text    string
	note    Note
}

// New starts a controller; it stops when ctx is done.
func New(ctx context.Context, d Deps) *Controller {
	if d.Capture == nil {
		d.Capture = func(bool, bool) {}
	}
	if d.OnChange == nil {
		d.OnChange = func(Status) {}
	}
	if d.OnTargetPicked == nil {
		d.OnTargetPicked = func(string) {}
	}
	c := &Controller{d: d, inbox: make(chan any, 16)}
	go c.run(ctx)
	return c
}

// Handle queues a hotkey action. It never blocks, so the event tap may
// call it; an action that finds the queue full is dropped.
func (c *Controller) Handle(a hotkey.Action) {
	select {
	case c.inbox <- action{a: a}:
	default:
	}
}

// SetTarget picks the translation target of the current recording, as a
// click on the voice bar does.
func (c *Controller) SetTarget(target string) {
	select {
	case c.inbox <- action{target: target}:
	default:
	}
}

// Status returns the current status.
func (c *Controller) Status() Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status
}

func (c *Controller) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if c.cancel != nil {
				c.cancel()
			}
			return
		case m := <-c.inbox:
			switch m := m.(type) {
			case action:
				if m.target != "" {
					c.pickTarget(m.target)
				} else {
					c.onAction(ctx, m.a)
				}
			case transcribed:
				c.onTranscribed(ctx, m)
			case rewritten:
				c.onRewritten(m)
			}
		}
	}
}

func (c *Controller) set(fn func(*Status)) {
	c.mu.Lock()
	fn(&c.status)
	s := c.status
	c.mu.Unlock()
	switch s.State {
	case Recording:
		c.d.Capture(true, s.Mode == prompt.Translate)
	case Transcribing, Rewriting:
		c.d.Capture(true, false)
	default:
		c.d.Capture(false, false)
	}
	c.d.OnChange(s)
}

func (c *Controller) onAction(ctx context.Context, a hotkey.Action) {
	st := c.Status()
	switch st.State {
	case Idle:
		switch a {
		case hotkey.Dictate:
			c.start(prompt.Dictate)
		case hotkey.Translate:
			c.start(prompt.Translate)
		}
	case Recording:
		switch a {
		case hotkey.Dictate, hotkey.Translate:
			c.stop(ctx)
		case hotkey.Cancel:
			c.d.Recorder.Cancel()
			c.abort(NoteCancelled, nil)
		case hotkey.CycleTarget:
			if st.Mode == prompt.Translate {
				c.pickTarget(c.d.Settings().NextTarget(st.Target))
			}
		}
	case Transcribing, Rewriting:
		if a == hotkey.Cancel {
			c.abort(NoteCancelled, nil)
		}
	}
}

func (c *Controller) start(mode prompt.Mode) {
	if err := c.d.Recorder.Start(); err != nil {
		c.set(func(s *Status) { s.State, s.Note, s.Err = Idle, NoteMicFailed, err })
		return
	}
	c.session++
	c.manual = false
	target := ""
	if mode == prompt.Translate {
		target = c.d.Settings().InitialTarget()
	}
	c.timings = Timings{}
	c.set(func(s *Status) {
		s.State, s.Mode, s.Target, s.Note, s.Err = Recording, mode, target, NoteNone, nil
		s.Since, s.Timings = time.Now(), Timings{}
	})
}

func (c *Controller) pickTarget(target string) {
	st := c.Status()
	if st.State != Recording || st.Mode != prompt.Translate {
		return
	}
	c.manual = true
	c.set(func(s *Status) { s.Target = target })
	c.d.OnTargetPicked(target)
}

func (c *Controller) stop(ctx context.Context) {
	pcm, err := c.d.Recorder.Stop()
	switch {
	case err != nil:
		c.abort(NoteMicFailed, err)
		return
	case audio.Duration(pcm) < MinRecording:
		c.abort(NoteTooShort, nil)
		return
	case audio.Silent(pcm):
		c.abort(NoteSilent, nil)
		return
	}
	run, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	session := c.session
	c.set(func(s *Status) { s.State, s.Since = Transcribing, time.Now() })
	go func() {
		text, err := c.d.Transcriber.Transcribe(run, pcm)
		c.send(ctx, transcribed{session: session, text: text, err: err})
	}()
}

func (c *Controller) onTranscribed(ctx context.Context, m transcribed) {
	st := c.Status()
	if m.session != c.session || st.State != Transcribing {
		return
	}
	c.timings.Transcribe = time.Since(st.Since)
	switch {
	case errors.Is(m.err, ErrNoKey):
		c.abort(NoteNoKey, nil)
		return
	case m.err != nil:
		c.abort(NoteASRFailed, m.err)
		return
	case m.text == "":
		c.abort(NoteNothing, nil)
		return
	}
	settings := c.d.Settings()
	req := prompt.Request{Mode: st.Mode, Transcript: m.text, Dictionary: settings.Dictionary}
	if st.Mode == prompt.Translate {
		req.Target = st.Target
		if !c.manual {
			req.Target = settings.TargetFor(prompt.DetectLanguage(m.text))
		}
	}
	run, cancel := context.WithTimeout(ctx, settings.Timeout())
	c.cancel = cancel
	c.set(func(s *Status) { s.State, s.Target, s.Since = Rewriting, req.Target, time.Now() })
	go func() {
		defer cancel()
		out, err := c.d.Rewriter.Rewrite(run, req)
		out = prompt.Clean(out)
		r := rewritten{session: m.session, text: out}
		if err != nil || !prompt.Plausible(req, out) {
			r.text, r.note = m.text, NoteRawPasted
		}
		c.send(ctx, r)
	}()
}

func (c *Controller) onRewritten(m rewritten) {
	st := c.Status()
	if m.session != c.session || st.State != Rewriting {
		return
	}
	c.timings.Rewrite = time.Since(st.Since)
	c.cancel = nil
	c.set(func(s *Status) { s.State, s.Since = Pasting, time.Now() })
	note, err := m.note, c.d.Paster.Paste(m.text)
	if err != nil {
		note = NotePasteFailed
	}
	c.set(func(s *Status) {
		s.State, s.Text, s.Note, s.Err, s.Timings = Idle, m.text, note, err, c.timings
	})
}

// abort ends the session: the pipeline's results are ignored from now on.
func (c *Controller) abort(note Note, err error) {
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
	c.session++
	c.set(func(s *Status) { s.State, s.Note, s.Err = Idle, note, err })
}

// ErrNoKey is what a Transcriber returns, or wraps, without an API key.
var ErrNoKey = errors.New("no API key")

// send delivers a pipeline result, unless the controller has stopped.
func (c *Controller) send(ctx context.Context, m any) {
	select {
	case c.inbox <- m:
	case <-ctx.Done():
	}
}
