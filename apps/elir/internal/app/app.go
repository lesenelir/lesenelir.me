//go:build darwin || windows

// Package app is Elir's shell: a menu bar app without a Dock icon, its
// settings window and menu bar menu, the settings of the app itself
// (language, appearance, accent color), and the features it hosts, each a
// tool with its own settings pages.
package app

import (
	"log"
	"path/filepath"
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// App is Elir. Add its features, then Run it.
type App struct {
	features []Feature
	pages    []Page

	store    *store.Store
	settings *store.Section[Settings]
	window   *mygo.Window
	tray     *mygo.Tray
	keys     *Shortcuts
	tap      *platform.EventTap
	updater  updater
	quitting bool
	stopDone chan struct{}

	// Read and written on the main thread: by the views and in
	// Window.Update.
	page         string // the settings page shown
	uiLanguage   string // "" follows the system
	appearance   int    // index in appearances, bound to its control
	accent       string // the accent color's id, bound to its select
	openAtLogin  bool   // as the system has it, bound to its switch
	checkAuto    bool   // bound to the switch of Settings.CheckUpdates
	showInDock   bool   // bound to the switch of Settings.ShowInDock
	upd          updates
	systemAccent ui.Color // macOS's accent, for the "System" swatch
	perms        permissions
	saveErr      error
}

// New returns the app, without features.
func New() *App { return &App{updater: mygoUpdater{}} }

// Add adds a feature, before Run. Features show in the sidebar in the
// order they are added.
func (a *App) Add(f Feature) { a.features = append(a.features, f) }

// Run runs the app until it quits. A second launch shows the first one's
// settings instead.
func (a *App) Run() error {
	if !mygo.App.RequestSingleInstanceLock() {
		return nil
	}
	mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnSecondInstance(func([]string, string) { a.showSettings("") })
	// A click on the Dock icon, when Elir has one, opens the settings.
	mygo.App.OnActivate(func(hasVisibleWindows bool) {
		if a.showInDock && !hasVisibleWindows {
			a.showSettings("")
		}
	})
	mygo.App.OnBeforeQuit(a.beforeQuit)
	mygo.App.WhenReady(a.ready)
	return mygo.App.Run()
}

// stopTimeout is how long quitting waits for the tools to stop.
var stopTimeout = 3 * time.Second

func (a *App) beforeQuit(e *mygo.QuitEvent) {
	if a.stopDone != nil {
		select {
		case <-a.stopDone:
			a.quitting = true
			return
		default:
			e.PreventDefault()
			return
		}
	}
	e.PreventDefault()
	a.stopDone = make(chan struct{})
	// macOS recorder cancellation needs the UI loop to keep serving RunOnMain.
	go func() {
		stopped := make(chan struct{})
		go func() {
			for _, f := range a.features {
				if s, ok := f.(Stopper); ok {
					s.Stop()
				}
			}
			close(stopped)
		}()
		// A tool stuck stopping must not keep the app from quitting; the
		// next launch undoes what a recording left quieted.
		select {
		case <-stopped:
		case <-time.After(stopTimeout):
			log.Printf("quit: the tools did not stop within %v", stopTimeout)
		}
		close(a.stopDone)
		mygo.App.Quit()
	}()
}

func (a *App) ready() {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	if ok, err := migrateFromKit(dir); err != nil {
		log.Printf("migrate from Kit: %v", err)
	} else if ok {
		log.Printf("migrated from Kit: %s", dir)
	}
	secrets, err := secret.OpenStore(filepath.Join(dir, "secrets.json"))
	if err != nil {
		log.Printf("secrets: %v", err)
	}
	a.store, err = store.Open(filepath.Join(dir, "settings.json"))
	if err != nil {
		log.Printf("settings: %v; using defaults", err)
	}
	if !a.store.Existed() {
		migrateFromVoice(dir, a.store, secrets)
	}
	a.settings, err = store.NewSection(a.store, settingsSection, defaultSettings(), nil)
	if err != nil {
		log.Printf("settings %q: %v", settingsSection, err)
	}
	s := a.settings.Get()
	a.uiLanguage = s.UILanguage
	mygo.App.SetLocale(a.uiLanguage)
	a.appearance = max(0, slices.Index(appearances, s.Appearance))
	a.accent = s.AccentColor
	a.openAtLogin = mygo.App.OpenAtLogin()
	a.checkAuto = s.CheckUpdates
	a.showInDock = s.ShowInDock

	a.window = mygo.NewWindow(mygo.WindowOptions{
		Title:     locales.SettingsTitle(),
		Hidden:    true,
		Width:     settingsWidth,
		Height:    settingsHeight,
		MinWidth:  640,
		MinHeight: 420,
		StateKey:  "settings-window",
		// The view draws the whole window, the traffic lights over its
		// sidebar, so no title bar of another color sits over it.
		TitleBarStyle: mygo.TitleBarHiddenInset,
		Vibrancy:      mygo.VibrancySidebar,
		Content:       ui.View(a.view),
	})
	a.keys = newShortcuts(a.store, a.window)
	a.window.OnBlur(func() {
		a.keys.CancelEdit()
		for _, f := range a.features {
			if b, ok := f.(SettingsBlurred); ok {
				b.SettingsBlurred()
			}
		}
	})
	a.window.OnClose(func(e *mygo.CloseEvent) {
		a.keys.CancelEdit()
		if !a.quitting {
			e.PreventDefault()
			a.window.Hide()
		}
	})

	a.setupTray()
	host := &Host{
		Store:        a.store,
		Secrets:      secrets,
		DataDir:      dir,
		Settings:     a.window,
		Tray:         a.tray,
		ShowSettings: a.showSettings,
		Shortcuts:    a.keys,
	}
	for _, f := range a.features {
		if err := f.Start(host); err != nil {
			log.Printf("feature: %v", err)
		}
	}
	a.pages = a.collectPages()
	a.tap = &platform.EventTap{Handler: a.keys}
	a.keys.Running = a.tap.Running

	// MyGo refreshes the light or dark look of the windows that exist when
	// the appearance is set, so set it once the features' are made too.
	applyAppearance(s.Appearance)
	applyDock(s.ShowInDock)

	mygo.App.OnLocaleChanged(func(string) {
		a.window.SetTitle(locales.SettingsTitle())
		a.localizeTray()
	})

	// Ask for Accessibility once; macOS shows its own dialog.
	platform.AccessibilityTrusted(true)
	go a.watchPermissions()
	go a.watchUpdates()
	if a.needsSetup() || !platform.AccessibilityTrusted(false) {
		a.showSettings("")
	}
}

// collectPages lists the sidebar's pages: General, the features' pages
// used every day, then those set up once, then Permissions.
func (a *App) collectPages() []Page {
	pages := []Page{a.generalPage()}
	var setup []Page
	for _, f := range a.features {
		for _, p := range f.Pages() {
			if p.Setup {
				setup = append(setup, p)
			} else {
				pages = append(pages, p)
			}
		}
	}
	return append(append(pages, setup...), a.permissionsPage())
}

func (a *App) needsSetup() bool {
	for _, f := range a.features {
		if n, ok := f.(NeedsSetup); ok && n.NeedsSetup() {
			return true
		}
	}
	return false
}

// showSettings opens the settings window at a page, "" for the last one.
func (a *App) showSettings(page string) {
	a.window.Update(func() {
		for _, f := range a.features {
			if s, ok := f.(SettingsShown); ok {
				s.SettingsShown()
			}
		}
		if page != "" {
			a.keys.CancelEdit()
			a.page = page
		}
		// The user may have changed it in System Settings meanwhile.
		a.openAtLogin = mygo.App.OpenAtLogin()
	})
	a.window.Show()
	a.window.Focus()
}
