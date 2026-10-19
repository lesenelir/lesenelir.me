//go:build darwin || windows

// Package voice is Elir's voice input. Tap Fn, speak, tap Fn again: the
// speech is transcribed, cleaned up and pasted where the cursor is. Fn +
// Left Shift does the same but translates; Esc cancels. Speech goes to
// Soniox as it is spoken, or to OpenRouter once it ends; the rewrite goes
// to OpenRouter.
package voice

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/controller"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/history"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/hotkey"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/soniox"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// Voice is voice input, a feature of the app.
type Voice struct {
	host     *app.Host
	settings *store.Section[config.Settings]
	ctrl     *controller.Controller
	rec      *platform.Recorder
	keys     *hotkeys
	overlay  *mygo.Window

	openrouterKey *secret.Secret
	sonioxKey     *secret.Secret
	history       *history.Store

	// Read and written on the main thread: by the views and in
	// Window.Update.
	status  controller.Status
	hideSeq int
	form    settingsForm
	saved   saveResult
	hist    historyView
	levels  [waveBars]float32 // the voice bar's waveform, oldest first
	levelAt time.Time         // when levels last moved
}

var _ interface {
	app.Feature
	app.NeedsSetup
	app.SettingsShown
	app.SettingsBlurred
	app.PermissionRows
	app.PermissionTips
} = (*Voice)(nil)

// New returns voice input, to Add to the app.
func New() *Voice {
	return &Voice{
		rec:           &platform.Recorder{},
		keys:          &hotkeys{},
		openrouterKey: secret.New("openrouter-api-key", "OPENROUTER_API_KEY"),
		sonioxKey:     secret.New("soniox-api-key", "SONIOX_API_KEY"),
	}
}

// Start reads the settings, history and keys, makes the voice bar, and
// watches Fn.
func (v *Voice) Start(h *app.Host) error {
	v.host = h
	dir := filepath.Join(h.DataDir, "voice")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var err error
	v.settings, err = store.NewSection(h.Store, config.Section, config.Defaults(), config.Clone)
	if err != nil {
		log.Printf("voice settings: %v; using defaults", err)
	}
	v.history = history.Open(filepath.Join(dir, "history.jsonl"))
	v.loadHistory()
	quiet := &quieter{settings: v.settings.Get, path: filepath.Join(dir, "output-to-restore.json")}
	quiet.recover()
	v.openrouterKey.Load(h.Secrets)
	v.sonioxKey.Load(h.Secrets)

	client := &openrouter.Client{Key: v.openrouterKey.Get, HTTP: &http.Client{Timeout: 60 * time.Second}}
	sonioxClient := &soniox.Client{Key: v.sonioxKey.Get}
	v.ctrl = controller.New(context.Background(), controller.Deps{
		Recorder:    v.rec,
		Streamer:    &streamer{client: sonioxClient, settings: v.settings.Get, key: v.sonioxKey.Get},
		Transcriber: &transcriber{client: client, settings: v.settings.Get},
		Rewriter:    rewriter{client: client, settings: v.settings.Get},
		Paster:      platform.Paster{},
		Settings:    v.settings.Get,
		Capture:     v.keys.detector.Capture,
		OnChange:    v.onStatus,
		Recording:   quiet.Recording,
		OnResult:    v.onResult,
		OnTargetPicked: func(target string) {
			if v.settings.Get().RememberTarget {
				_ = v.settings.Update(func(s *config.Settings) { s.LastTarget = target })
			}
		},
	})
	v.keys.handle = v.ctrl.Handle

	v.overlay = mygo.NewWindow(mygo.WindowOptions{
		Title:       "Voice",
		Panel:       true,
		Hidden:      true,
		Frameless:   true,
		Transparent: true,
		// The capsule draws its own shadow; the window's would outline the
		// whole transparent rectangle.
		DisableShadow: true,
		Width:         overlayWidth,
		Height:        overlayHeight,
		Content:       ui.View(v.overlayView),
	})
	v.overlay.SetIgnoreMouseEvents(false)

	v.form = formFrom(v.settings.Get(), v.apiKeys())
	h.HandleKeys(v.keys)
	return nil
}

// NeedsSetup reports whether the OpenRouter key, which cleanup and
// translation need, is missing.
func (v *Voice) NeedsSetup() bool { return v.openrouterKey.Get() == "" }

// SettingsShown reads the fields afresh.
func (v *Voice) SettingsShown() {
	v.flushKeys()
	v.form = formFrom(v.settings.Get(), v.apiKeys())
}

// SettingsBlurred saves the keys still being typed.
func (v *Voice) SettingsBlurred() { v.flushKeys() }

// PermissionRows shows the microphone permission.
func (v *Voice) PermissionRows(c *ui.Context, m locales.Messages, pal widget.Palette, g *widget.Rows) {
	t := c.Theme()
	mic := platform.Microphone()
	g.Row(c, m.PermMicrophone(), micText(m, mic), func() {
		switch mic {
		case platform.MicAuthorized:
			widget.Badge(c, m.StatusAllowed(), t.Success)
		case platform.MicDenied, platform.MicRestricted:
			widget.Badge(c, m.StatusNeeded(), t.Danger)
			if ui.Button(c, m.OpenSettings()).Clicked() {
				platform.OpenMicrophoneSettings()
			}
		default:
			widget.Badge(c, m.StatusWaiting(), t.TextMuted)
		}
	})
}

func micText(m locales.Messages, s platform.MicStatus) string {
	switch s {
	case platform.MicAuthorized:
		return m.MicAllowed()
	case platform.MicDenied, platform.MicRestricted:
		return m.MicDenied(runtime.GOOS)
	}
	return m.MicAsk(runtime.GOOS)
}

// PermissionTips tells how to keep the Globe key from macOS, which would
// otherwise open the emoji picker or start its own dictation on Fn.
func (v *Voice) PermissionTips(c *ui.Context, m locales.Messages, pal widget.Palette) {
	if runtime.GOOS == "windows" {
		return
	}
	widget.Group(c, pal, m.StatusTip(), func(g *widget.Rows) {
		g.Row(c, m.PermGlobe(), m.PermGlobeNote(), func() {
			if ui.Button(c, m.OpenKeyboardSettings()).Clicked() {
				platform.OpenKeyboardSettings()
			}
		})
	})
}

// apiKeys are the keys, as the settings show them.
func (v *Voice) apiKeys() apiKeys {
	return apiKeys{openrouter: v.openrouterKey.Get(), soniox: v.sonioxKey.Get()}
}

// hotkeys turns the event tap's key events into the controller's actions.
type hotkeys struct {
	detector hotkey.Detector
	handle   func(hotkey.Action)
}

func (k *hotkeys) FlagsChanged(keycode int64, flags uint64) bool {
	k.do(k.detector.FlagsChanged(keycode, flags))
	return false
}

func (k *hotkeys) KeyDown(keycode int64, _ uint64, repeat bool) bool {
	action, swallow := k.detector.KeyDown(keycode, repeat)
	k.do(action)
	return swallow
}

func (k *hotkeys) KeyUp(keycode int64) bool { return k.detector.KeyUp(keycode) }

func (k *hotkeys) do(a hotkey.Action) {
	if a != hotkey.None && k.handle != nil {
		k.handle(a)
	}
}
