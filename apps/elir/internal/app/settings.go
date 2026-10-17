//go:build darwin || windows

package app

// settingsSection is the settings file's section for the app itself.
const settingsSection = "app"

// Settings are what the user can change about the app itself, rather than
// one of its tools.
type Settings struct {
	// UILanguage is the language of the app's interface, as a BCP 47 tag
	// ("en", "zh-Hans"); "" follows the system.
	UILanguage string `json:"uiLanguage"`
	// Appearance is "light", "dark", or "" to follow the system.
	Appearance string `json:"appearance"`
	// AccentColor is the accent color's name ("blue", "green"…), or "" to
	// follow the one chosen in macOS.
	AccentColor string `json:"accentColor"`
	// ShowInDock gives Elir a Dock icon, and a place in ⌘Tab, besides its
	// menu bar icon.
	ShowInDock bool `json:"showInDock"`
	// CheckUpdates checks for a new version a while after launch and
	// then daily.
	CheckUpdates bool `json:"checkUpdates"`
}

// defaultSettings are the settings of a new install, and of those the
// file lacks.
func defaultSettings() Settings { return Settings{CheckUpdates: true} }
