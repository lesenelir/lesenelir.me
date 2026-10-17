//go:build darwin

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
}
