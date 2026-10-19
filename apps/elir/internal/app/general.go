//go:build darwin || windows

package app

import (
	"fmt"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
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
				g.Row(c, m.FieldOpenAtLogin(), m.OpenAtLoginNote(runtime.GOOS), func() {
					ui.Switch(c.Key("open-at-login"), &a.openAtLogin).Label(m.FieldOpenAtLogin()).OnChange(a.setOpenAtLogin)
				})
				if runtime.GOOS == "darwin" {
					g.Row(c, m.FieldShowInDock(), m.ShowInDockNote(), func() {
						ui.Switch(c.Key("show-in-dock"), &a.showInDock).Label(m.FieldShowInDock()).OnChange(a.setShowInDock)
					})
				}
			})
			a.updatesGroup(c, m, pal)
		},
	}
}

// update changes the app's settings and keeps them; a failure shows under
// the page.
func (a *App) update(fn func(*Settings)) {
	a.saveErr = a.settings.Update(fn)
}

// applyDock shows or hides Elir's Dock icon; its menu bar icon stays.
func applyDock(show bool) {
	if show {
		mygo.App.SetActivationPolicy(mygo.ActivationPolicyRegular)
	} else {
		mygo.App.SetActivationPolicy(mygo.ActivationPolicyAccessory)
	}
}

// setShowInDock keeps the choice and applies it at once. macOS deactivates
// an app that leaves the Dock, so the settings window comes back to the
// front.
func (a *App) setShowInDock() {
	show := a.showInDock
	a.update(func(s *Settings) { s.ShowInDock = show })
	applyDock(show)
	if a.window != nil {
		mygo.App.Focus()
		a.window.Show()
		a.window.Focus()
	}
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

// updatesGroup shows the version, and checks for and installs updates.
func (a *App) updatesGroup(c *ui.Context, m locales.Messages, pal widget.Palette) {
	u := &a.upd
	widget.Group(c, pal, m.SectionUpdates(), func(g *widget.Rows) {
		g.Row(c, m.AppVersion(mygo.App.Version()), a.updateNote(m), func() {
			switch {
			case u.state == updateAvailable, u.state == updateFailed && u.rel != nil:
				if ui.PrimaryButton(c.Key("install-update"), m.ButtonInstallUpdate()).Clicked() {
					u.state = updateAvailable
					a.installUpdate()
				}
			case u.state == updateInstalling:
				ui.Progress(c, u.progress).Width(120).Label(m.ButtonInstallUpdate())
			case u.state == updateChecking:
				ui.Spinner(c).Label(m.UpdateChecking())
			default:
				if ui.Button(c.Key("check-updates"), m.ButtonCheckUpdates()).Disabled(!a.updater.Enabled()).Clicked() {
					a.checkUpdates(true)
				}
			}
		})
		if u.rel != nil && u.rel.Notes() != "" {
			g.Block(c, func() {
				ui.Collapsible(c.Key("whats-new"), m.WhatsNew(), &u.notesOn, func() {
					markdown.View(c, u.rel.Notes(), nil)
				})
			})
		}
		g.Row(c, m.FieldCheckUpdates(), m.CheckUpdatesNote(), func() {
			ui.Switch(c.Key("check-updates-auto"), &a.checkAuto).Label(m.FieldCheckUpdates()).
				Disabled(!a.updater.Enabled()).OnChange(a.setCheckUpdates)
		})
	})
}

// updateNote says where updating stands, under the version.
func (a *App) updateNote(m locales.Messages) string {
	if runtime.GOOS == "windows" {
		return m.UpdatesManual()
	}
	u := a.upd
	switch {
	case !a.updater.Enabled() && mygo.IsDev():
		return m.UpdatesDev()
	case !a.updater.Enabled():
		return m.UpdatesUnwritable()
	case u.state == updateChecking:
		return m.UpdateChecking()
	case u.state == updateLatest && u.manual:
		return m.UpdateLatest()
	case u.state == updateAvailable:
		return m.UpdateAvailable(u.rel.Version())
	case u.state == updateInstalling:
		return m.UpdateDownloading(fmt.Sprintf("%.0f", 100*u.progress))
	case u.state == updateFailed && u.rel != nil:
		return m.UpdateFailed(u.err.Error())
	case u.state == updateFailed:
		return m.UpdateCheckFailed(u.err.Error())
	}
	return ""
}
