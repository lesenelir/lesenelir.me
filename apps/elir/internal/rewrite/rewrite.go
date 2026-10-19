//go:build darwin || windows

// Package rewrite is Elir's tool for selected text. Select text in any
// app and press ⌥T to translate it or ⌥R to polish it: the result replaces
// the selection, which ⌘Z undoes. Translation follows the languages of
// voice input's settings, and both use its chat model and OpenRouter key.
package rewrite

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
)

const (
	// copyWait is how long the app in front has to put the selection on
	// the clipboard after Cmd+C.
	copyWait = 600 * time.Millisecond
	// minTimeout is the least a rewrite waits for the model: voice input's
	// timeout is made for a sentence spoken, not a page selected.
	minTimeout = 30 * time.Second
)

// Errors a rewrite ends with, which the bar words.
var (
	errNoSelection = errors.New("rewrite: nothing is selected")
	errNoKey       = errors.New("rewrite: no OpenRouter API key")
)

// Rewrite is the tool for selected text, a feature of the app.
type Rewrite struct {
	store  *store.Store
	key    *secret.Secret
	client *openrouter.Client
	host   *app.Host
	bar    *mygo.Window

	busy   atomic.Bool
	mu     sync.Mutex
	cancel context.CancelFunc // the rewrite running

	// Read and written on the main thread: by the views and in
	// Window.Update.
	hud hud
}

var _ app.Feature = (*Rewrite)(nil)

// New returns the tool for selected text, to Add to the app.
func New() *Rewrite {
	r := &Rewrite{key: secret.New("openrouter-api-key", "OPENROUTER_API_KEY")}
	return r
}

// Start reads the API key, makes the bar and takes ⌥T, ⌥R and Esc.
func (r *Rewrite) Start(h *app.Host) error {
	r.host = h
	r.store = h.Store
	r.key.Load(h.Secrets)
	r.client = &openrouter.Client{Key: r.key.Get, HTTP: &http.Client{Timeout: 2 * minTimeout}}
	r.bar = widget.NewOverlay(h.Tray, "Elir", barWidth, barHeight, r.barView)
	h.Shortcuts.Register(shortcut.TranslateText, nil, func() { r.start(Translate) })
	h.Shortcuts.Register(shortcut.PolishText, nil, func() { r.start(Polish) })
	h.Shortcuts.Register(shortcut.RewriteCancel, r.busy.Load, r.stop)
	return nil
}

// voiceSettings reads voice input's settings afresh, as its pages change
// them: the model, the languages to translate into, the dictionary.
func (r *Rewrite) voiceSettings() config.Settings {
	s := config.Defaults()
	if raw := r.store.Raw(config.Section); raw != nil {
		if err := json.Unmarshal(raw, &s); err != nil {
			return config.Defaults()
		}
	}
	return s
}

// start rewrites the selection, unless a rewrite runs. It runs on the
// event tap's thread, so the work runs on a goroutine of its own.
func (r *Rewrite) start(mode Mode) {
	if !r.busy.CompareAndSwap(false, true) {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.cancel = cancel
	r.mu.Unlock()
	go func() {
		defer r.busy.Store(false)
		defer cancel()
		r.run(ctx, mode)
	}()
}

// stop cancels the rewrite running, as Esc does.
func (r *Rewrite) stop() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (r *Rewrite) run(ctx context.Context, mode Mode) {
	text, err := copySelection()
	if err == nil && r.key.Get() == "" {
		err = errNoKey
	}
	if err != nil {
		r.show(hud{state: hudNote, err: err})
		return
	}
	s := r.voiceSettings()
	target := ""
	if mode == Translate {
		target = s.TargetFor(prompt.DetectLanguage(text))
	}
	r.show(hud{state: hudWorking, mode: mode, target: target, since: time.Now()})

	out, err := r.rewrite(ctx, mode, text, target, s)
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		r.show(hud{state: hudNote, err: context.Canceled})
		return
	case err != nil:
		log.Printf("rewrite: %v", err)
		r.show(hud{state: hudNote, err: err})
		return
	}
	if err := (platform.Paster{}).Paste(out); err != nil {
		r.show(hud{state: hudNote, err: err})
		return
	}
	r.show(hud{state: hudDone, mode: mode, result: out})
}

// rewrite asks the chat model of voice input's settings to translate the
// text into target, or to polish it.
func (r *Rewrite) rewrite(ctx context.Context, mode Mode, text, target string, s config.Settings) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, max(s.Timeout(), minTimeout))
	defer cancel()
	start := time.Now()
	out, err := r.client.Chat(ctx, openrouter.ChatRequest{
		Model: s.LLMModel,
		Messages: []openrouter.Message{
			{Role: "system", Content: system(mode, prompt.LanguageFor(target).Name, s.Dictionary)},
			{Role: "user", Content: user(text)},
		},
		Temperature: 0.2,
		MaxTokens:   maxTokens(text),
		Reasoning:   s.ReasoningEffort(),
		Providers:   s.Providers,
	})
	log.Printf("rewrite %s: %v, err=%v", s.LLMModel, time.Since(start).Round(time.Millisecond), err)
	if err != nil {
		return "", err
	}
	out = clean(out, text)
	if strings.TrimSpace(out) == "" {
		return "", errors.New("the model answered nothing")
	}
	return out, nil
}

// copySelection copies what is selected in the app in front, and puts the
// clipboard back as it was.
func copySelection() (string, error) {
	saved, err := mygo.Clipboard.Read()
	if err != nil {
		return "", err
	}
	keep := len(saved.Items()) > 0
	restore := func() {
		if keep {
			_ = mygo.Clipboard.Write(saved)
		} else {
			mygo.Clipboard.Clear()
		}
	}
	if err := mygo.Clipboard.Write(transfer.Data{}); err != nil {
		return "", err
	}
	if err := platform.CopySelection(); err != nil {
		restore()
		return "", err
	}
	var text string
	for deadline := time.Now().Add(copyWait); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if text = mygo.Clipboard.ReadText(); text != "" {
			break
		}
	}
	restore()
	if strings.TrimSpace(text) == "" {
		return "", errNoSelection
	}
	return text, nil
}
