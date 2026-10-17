//go:build darwin

package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

var iconPermissions = widget.Icon(`<path d="M8 1.5l5 2v4c0 3.2-2.2 5.6-5 7-2.8-1.4-5-3.8-5-7v-4z"/><path d="M5.8 8l1.6 1.6L10.4 6.6"/>`)

// permissions are the system's privacy permissions the tools use.
type permissions struct {
	accessibility bool
	microphone    darwin.MicStatus
}

// watchPermissions keeps the Permissions page current, as the user changes
// them in System Settings.
func (a *App) watchPermissions() {
	for {
		p := permissions{
			accessibility: darwin.AccessibilityTrusted(false),
			microphone:    darwin.Microphone(),
		}
		a.window.Update(func() { a.perms = p })
		time.Sleep(2 * time.Second)
	}
}

// permissionsPage shows the system's permissions, then each feature's part.
func (a *App) permissionsPage() Page {
	return Page{
		ID:    "permissions",
		Icon:  iconPermissions,
		Title: locales.Messages.NavPermissions,
		Setup: true,
		View: func(c *ui.Context, m locales.Messages, pal widget.Palette) {
			t := c.Theme()
			widget.Group(c, pal, "", func(g *widget.Rows) {
				g.Row(c, m.PermAccessibility(), m.PermAccessibilityNote(), func() {
					if a.perms.accessibility {
						widget.Badge(c, m.StatusAllowed(), t.Success)
						return
					}
					widget.Badge(c, m.StatusNeeded(), t.Danger)
					if ui.Button(c, m.OpenSettings()).Clicked() {
						darwin.OpenAccessibilitySettings()
					}
				})
				g.Row(c, m.PermMicrophone(), micText(m, a.perms.microphone), func() {
					switch a.perms.microphone {
					case darwin.MicAuthorized:
						widget.Badge(c, m.StatusAllowed(), t.Success)
					case darwin.MicDenied, darwin.MicRestricted:
						widget.Badge(c, m.StatusNeeded(), t.Danger)
						if ui.Button(c, m.OpenSettings()).Clicked() {
							darwin.OpenMicrophoneSettings()
						}
					default:
						widget.Badge(c, m.StatusWaiting(), t.TextMuted)
					}
				})
				for _, f := range a.features {
					if p, ok := f.(PermissionsView); ok {
						p.PermissionRows(c, m, pal, g)
					}
				}
			})
			for _, f := range a.features {
				if p, ok := f.(PermissionsView); ok {
					p.PermissionTips(c, m, pal)
				}
			}
		},
	}
}

func micText(m locales.Messages, s darwin.MicStatus) string {
	switch s {
	case darwin.MicAuthorized:
		return m.MicAllowed()
	case darwin.MicDenied, darwin.MicRestricted:
		return m.MicDenied()
	}
	return m.MicAsk()
}
