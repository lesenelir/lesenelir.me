//go:build darwin

// Voice is a menu bar app for voice input. Tap Fn, speak, tap Fn again: the
// speech is transcribed, cleaned up and pasted where the cursor is. Fn +
// Left Shift does the same but translates; Esc cancels. Speech to text and
// the rewrite both go through OpenRouter.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

// secretKey names the OpenRouter API key in the Keychain.
const secretKey = "openrouter-api-key"

type app struct {
	store *store.Store
	ctrl  *controller.Controller
	rec   *darwin.Recorder
	tap   *darwin.EventTap

	overlay  *mygo.Window
	settings *mygo.Window
	tray     *mygo.Tray
	quitting bool

	keyMu sync.RWMutex
	key   string

	// Read and written on the main thread: by the views and in
	// Window.Update.
	status          controller.Status
	hideSeq         int
	form            settingsForm
	perms           permissions
	saved           saveResult
	uiLanguage      string // "" follows the system
	uiLanguageLabel string // bound to the language select
}

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	a := &app{rec: &darwin.Recorder{}, tap: &darwin.EventTap{}}
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnSecondInstance(func([]string, string) { a.showSettings() })
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { a.quitting = true })
	mygo.App.WhenReady(a.ready)
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *app) ready() {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	a.store, err = store.Open(filepath.Join(dir, "settings.json"))
	if err != nil {
		log.Printf("settings: %v; using defaults", err)
	}
	a.loadKey()
	a.uiLanguage = a.store.Get().UILanguage
	mygo.App.SetLocale(a.uiLanguage)

	client := &openrouter.Client{Key: a.apiKey, HTTP: &http.Client{Timeout: 60 * time.Second}}
	a.ctrl = controller.New(context.Background(), controller.Deps{
		Recorder:    a.rec,
		Transcriber: transcriber{client: client, store: a.store},
		Rewriter:    rewriter{client: client, store: a.store},
		Paster:      darwin.Paster{},
		Settings:    a.store.Get,
		Capture:     a.tap.Detector.Capture,
		OnChange:    a.onStatus,
		OnTargetPicked: func(target string) {
			if a.store.Get().RememberTarget {
				_ = a.store.Update(func(s *store.Settings) { s.LastTarget = target })
			}
		},
	})
	a.tap.OnAction = a.ctrl.Handle

	a.overlay = mygo.NewWindow(mygo.WindowOptions{
		Title:       "Voice",
		Panel:       true,
		Hidden:      true,
		Frameless:   true,
		Transparent: true,
		Width:       overlayWidth,
		Height:      overlayHeight,
		Content:     ui.View(a.overlayView),
	})
	a.overlay.SetIgnoreMouseEvents(false)

	a.form = formFrom(a.store.Get(), a.apiKey())
	a.settings = mygo.NewWindow(mygo.WindowOptions{
		Title:     locales.SettingsTitle(),
		Hidden:    true,
		Width:     560,
		Height:    720,
		MinWidth:  480,
		MinHeight: 480,
		StateKey:  "settings",
		Content:   ui.View(a.settingsView),
	})
	a.settings.OnClose(func(e *mygo.CloseEvent) {
		if !a.quitting {
			e.PreventDefault()
			a.settings.Hide()
		}
	})

	a.setupTray()
	mygo.App.OnLocaleChanged(func(string) {
		a.settings.SetTitle(locales.SettingsTitle())
		a.localizeTray()
	})

	// Ask for Accessibility once; macOS shows its own dialog.
	darwin.AccessibilityTrusted(true)
	go a.watchPermissions()
	if a.apiKey() == "" || !darwin.AccessibilityTrusted(false) {
		a.showSettings()
	}
}

// watchPermissions starts the event tap as soon as Accessibility is
// granted, and keeps the settings' permission rows current.
func (a *app) watchPermissions() {
	for {
		running := a.tap.Running()
		if !running && darwin.AccessibilityTrusted(false) {
			if err := a.tap.Start(); err != nil {
				log.Printf("event tap: %v", err)
			} else {
				running = true
				log.Print("event tap: watching Fn")
			}
		}
		p := permissions{
			accessibility: darwin.AccessibilityTrusted(false),
			microphone:    darwin.Microphone(),
			tapRunning:    running,
		}
		a.settings.Update(func() { a.perms = p })
		time.Sleep(2 * time.Second)
	}
}

func (a *app) loadKey() {
	key, err := mygo.Secrets.Get(secretKey)
	if err != nil && !errors.Is(err, mygo.ErrSecretNotFound) {
		log.Printf("keychain: %v", err)
	}
	k := string(key)
	if k == "" {
		k = os.Getenv("OPENROUTER_API_KEY")
	}
	a.keyMu.Lock()
	a.key = k
	a.keyMu.Unlock()
}

func (a *app) apiKey() string {
	a.keyMu.RLock()
	defer a.keyMu.RUnlock()
	return a.key
}

func (a *app) setKey(k string) error {
	if k == "" {
		if err := mygo.Secrets.Delete(secretKey); err != nil {
			return err
		}
	} else if err := mygo.Secrets.Set(secretKey, []byte(k)); err != nil {
		return err
	}
	a.keyMu.Lock()
	a.key = k
	a.keyMu.Unlock()
	return nil
}

func (a *app) showSettings() {
	a.settings.Update(func() { a.form = formFrom(a.store.Get(), a.apiKey()) })
	a.settings.Show()
	a.settings.Focus()
}
