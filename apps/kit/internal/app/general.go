//go:build darwin

package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

var iconGeneral = widget.Icon(`<circle cx="8" cy="8" r="2.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4"/>`)

// generalPage is the app's own settings: its language and its look.
func (a *App) generalPage() Page {
	return Page{
		ID:    "general",
		Icon:  iconGeneral,
		Title: locales.Messages.NavGeneral,
		View: func(c *ui.Context, m locales.Messages, pal widget.Palette) {
			widget.Group(c, pal, "", func(g *widget.Rows) {
				g.Row(c, m.FieldLanguage(), "", func() {
					widget.Dropdown(c.Key("ui-language"), m.FieldLanguage(), &a.uiLanguage, uiLanguages,
						func(tag string) string { return languageLabel(m, tag) }, nil, 180, a.applyUILanguage)
				})
				g.Row(c, m.FieldAppearance(), "", func() {
					ui.Segmented(c.Key("appearance"), &a.appearance,
						m.AppearanceSystem(), m.AppearanceLight(), m.AppearanceDark()).
						Label(m.FieldAppearance()).OnChange(func() { a.setAppearance(a.appearance) })
				})
				g.Row(c, m.FieldAccent(), "", func() { a.accentSelect(c, m, a.systemAccent) })
			})
			widget.Group(c, pal, "", func(g *widget.Rows) {
				g.Row(c, m.FieldOpenAtLogin(), m.OpenAtLoginNote(), func() {
					ui.Switch(c.Key("open-at-login"), &a.openAtLogin).Label(m.FieldOpenAtLogin()).OnChange(a.setOpenAtLogin)
				})
			})
		},
	}
}

// update changes the app's settings and keeps them; a failure shows under
// the page.
func (a *App) update(fn func(*Settings)) {
	a.saveErr = a.settings.Update(fn)
}

// setOpenAtLogin makes the app start when the user logs in, or not, as
// its switch says. It is the system's setting, which the user can also
// change in System Settings → General → Login Items, so the app keeps no
// copy of it.
func (a *App) setOpenAtLogin() {
	a.saveErr = mygo.App.SetOpenAtLogin(a.openAtLogin)
	a.openAtLogin = mygo.App.OpenAtLogin()
}

// appearances are the appearance choices, in the order the segmented
// control shows them, as the settings keep them.
var appearances = []string{"", "light", "dark"}

// applyAppearance makes the app light, dark or the system's.
func applyAppearance(appearance string) {
	source := mygo.ThemeSystem
	switch appearance {
	case "light":
		source = mygo.ThemeLight
	case "dark":
		source = mygo.ThemeDark
	}
	mygo.Theme.SetSource(source)
}

// setAppearance switches the appearance at once and keeps the choice.
func (a *App) setAppearance(i int) {
	appearance := appearances[i]
	a.update(func(s *Settings) { s.Appearance = appearance })
	applyAppearance(appearance)
}

// uiLanguages are the interface languages, "" for the system's.
var uiLanguages = []string{"", "en", "zh-Hans"}

// languageLabel names an interface language in itself, or the system's.
func languageLabel(m locales.Messages, tag string) string {
	switch tag {
	case "en":
		return "English"
	case "zh-Hans":
		return "简体中文"
	}
	return m.LanguageSystem()
}

// applyUILanguage switches the interface to the language chosen, at once,
// and keeps the choice.
func (a *App) applyUILanguage() {
	tag := a.uiLanguage
	a.update(func(s *Settings) { s.UILanguage = tag })
	mygo.App.SetLocale(tag)
}
