//go:build darwin

// Package app is Kit's shell: a menu bar app without a Dock icon, its
// settings window and menu bar menu, the settings of the app itself
// (language, appearance, accent color), and the features it hosts, each a
// tool with its own settings pages.
package app

import (
	"log"
	"path/filepath"
	"slices"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

// App is Kit. Add its features, then Run it.
type App struct {
	features []Feature
	pages    []Page

	store    *store.Store
	settings *store.Section[Settings]
	window   *mygo.Window
	tray     *mygo.Tray
	keys     keys
	tap      *darwin.EventTap
	updater  updater
	quitting bool

	// Read and written on the main thread: by the views and in
	// Window.Update.
	page         string // the settings page shown
	uiLanguage   string // "" follows the system
	appearance   int    // index in appearances, bound to its control
	accent       string // the accent color's id, bound to its select
	openAtLogin  bool   // as the system has it, bound to its switch
	checkAuto    bool   // bound to the switch of Settings.CheckUpdates
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
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { a.quitting = true })
	mygo.App.WhenReady(a.ready)
	return mygo.App.Run()
}

func (a *App) ready() {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	a.store, err = store.Open(filepath.Join(dir, "settings.json"))
	if err != nil {
		log.Printf("settings: %v; using defaults", err)
	}
	if !a.store.Existed() {
		migrateFromVoice(dir, a.store)
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
	a.window.OnBlur(func() {
		for _, f := range a.features {
			if b, ok := f.(SettingsBlurred); ok {
				b.SettingsBlurred()
			}
		}
	})
	a.window.OnClose(func(e *mygo.CloseEvent) {
		if !a.quitting {
			e.PreventDefault()
			a.window.Hide()
		}
	})

	host := &Host{
		Store:        a.store,
		DataDir:      dir,
		Settings:     a.window,
		ShowSettings: a.showSettings,
		HandleKeys:   func(h darwin.KeyHandler) { a.keys = append(a.keys, h) },
	}
	for _, f := range a.features {
		if err := f.Start(host); err != nil {
			log.Printf("feature: %v", err)
		}
	}
	a.pages = a.collectPages()
	a.tap = &darwin.EventTap{Handler: a.keys}

	// MyGo refreshes the light or dark look of the windows that exist when
	// the appearance is set, so set it once the features' are made too.
	applyAppearance(s.Appearance)

	a.setupTray()
	mygo.App.OnLocaleChanged(func(string) {
		a.window.SetTitle(locales.SettingsTitle())
		a.localizeTray()
	})

	// Ask for Accessibility once; macOS shows its own dialog.
	darwin.AccessibilityTrusted(true)
	go a.watchPermissions()
	go a.watchUpdates()
	if a.needsSetup() || !darwin.AccessibilityTrusted(false) {
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
			a.page = page
		}
		// The user may have changed it in System Settings meanwhile.
		a.openAtLogin = mygo.App.OpenAtLogin()
	})
	a.window.Show()
	a.window.Focus()
}
