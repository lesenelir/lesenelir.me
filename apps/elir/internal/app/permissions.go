//go:build darwin || windows

package app

import (
	"log"
	"runtime"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

var iconPermissions = widget.Icon(`<path d="M8 1.5l5 2v4c0 3.2-2.2 5.6-5 7-2.8-1.4-5-3.8-5-7v-4z"/><path d="M5.8 8l1.6 1.6L10.4 6.6"/>`)

// permissions are what the app itself needs: Accessibility, for the
// hotkeys of every tool and for pasting.
type permissions struct {
	accessibility bool
	tapRunning    bool
}

// watchPermissions starts the event tap as soon as Accessibility is
// allowed, and keeps the Permissions page current as the user changes the
// permissions in System Settings.
func (a *App) watchPermissions() {
	for {
		trusted := platform.AccessibilityTrusted(false)
		running := a.tap.Running()
		if !running && trusted {
			if err := a.tap.Start(); err != nil {
				log.Printf("event tap: %v", err)
			} else {
				running = true
				log.Print("event tap: watching the hotkeys")
			}
		}
		p := permissions{accessibility: trusted, tapRunning: running}
		a.window.Update(func() { a.perms = p })
		time.Sleep(2 * time.Second)
	}
}

// permissionsPage shows Accessibility and whether the hotkeys run, then
// the permissions of each feature, then their tips.
func (a *App) permissionsPage() Page {
	return Page{
		ID:    "permissions",
		Icon:  iconPermissions,
		Title: locales.Messages.NavPermissions,
		Setup: true,
		View: func(c *ui.Context, m locales.Messages, pal widget.Palette) {
			t := c.Theme()
			widget.Group(c, pal, "", func(g *widget.Rows) {
				if runtime.GOOS == "darwin" {
					g.Row(c, m.PermAccessibility(), m.PermAccessibilityNote(), func() {
						if a.perms.accessibility {
							widget.Badge(c, m.StatusAllowed(), t.Success)
							return
						}
						widget.Badge(c, m.StatusNeeded(), t.Danger)
						if ui.Button(c, m.OpenSettings()).Clicked() {
							platform.OpenAccessibilitySettings()
						}
					})
				}
				g.Row(c, m.PermHotkeys(), m.PermHotkeysNote(), func() {
					if a.perms.tapRunning {
						widget.Badge(c, m.StatusRunning(), t.Success)
					} else {
						widget.Badge(c, m.StatusWaiting(), t.TextMuted)
					}
				})
				for _, f := range a.features {
					if p, ok := f.(PermissionRows); ok {
						p.PermissionRows(c, m, pal, g)
					}
				}
			})
			for _, f := range a.features {
				if p, ok := f.(PermissionTips); ok {
					p.PermissionTips(c, m, pal)
				}
			}
		},
	}
}
