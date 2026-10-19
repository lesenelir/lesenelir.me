//go:build darwin || windows

package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// Feature is a tool in Elir, as voice input: what it does in the background
// and the pages it adds to the settings window. The app starts it once it
// is ready, on the main thread.
type Feature interface {
	// Start sets the feature up: its settings, windows, hotkeys.
	Start(h *Host) error
	// Pages are the feature's settings pages, in the order the sidebar
	// lists them.
	Pages() []Page
}

// Optional parts of a Feature, which the app calls when it has them.
type (
	// Stopper releases background resources before quit, off the main thread.
	Stopper interface{ Stop() }
	// NeedsSetup reports whether the feature lacks something to work, as
	// an API key: the settings window then opens at launch.
	NeedsSetup interface{ NeedsSetup() bool }
	// SettingsShown is told the settings window is about to show, on the
	// main thread, to read its fields afresh.
	SettingsShown interface{ SettingsShown() }
	// SettingsBlurred is told the settings window lost the focus, on the
	// main thread, to save what is still being typed.
	SettingsBlurred interface{ SettingsBlurred() }
	// PermissionRows adds the rows of the permissions the feature needs to
	// the Permissions page's group, as the microphone. The page shows
	// afresh every two seconds, so the rows can read the permissions as
	// they draw.
	PermissionRows interface {
		PermissionRows(c *ui.Context, m locales.Messages, pal widget.Palette, g *widget.Rows)
	}
	// PermissionTips adds groups of tips under the Permissions page's
	// group, as a system setting the feature works better with.
	PermissionTips interface {
		PermissionTips(c *ui.Context, m locales.Messages, pal widget.Palette)
	}
)

// Page is a settings page in the sidebar.
type Page struct {
	ID    string
	Icon  *ui.SVG
	Title func(m locales.Messages) string
	// Setup puts the page with those set up once, as Models and
	// Permissions, below the ones used every day.
	Setup bool
	View  func(c *ui.Context, m locales.Messages, pal widget.Palette)
}

// Host is what the app gives a feature.
type Host struct {
	// Store is the settings file, of which a feature keeps a section.
	Store *store.Store
	// Secrets is the file of API keys, which only the user can read.
	Secrets *secret.Store
	// DataDir is the app's data directory; a feature keeps its files in a
	// subdirectory named after it.
	DataDir string
	// Settings is the settings window, for Update and Invalidate from the
	// feature's goroutines.
	Settings *mygo.Window
	// Tray owns the tools' nonactivating overlays on Windows.
	Tray *mygo.Tray
	// ShowSettings opens the settings window at a page, "" for the one
	// shown last.
	ShowSettings func(page string)
	// Shortcuts owns the shared bindings, editor and event router.
	Shortcuts *Shortcuts
}

func (h *Host) Shortcut(c *ui.Context, a shortcut.Action) { h.Shortcuts.Control(c, a) }
