//go:build darwin || windows

package app

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// The settings window is laid out as macOS's System Settings: a sidebar of
// pages over the window's sidebar material, under a hidden title bar whose
// traffic lights sit over the sidebar, and a page of grouped rows.
const (
	settingsWidth  = 760
	settingsHeight = 540
	sidebarWidth   = 200
	headerHeight   = 52
)

func (a *App) view(c *ui.Context) {
	a.systemAccent = c.Theme().Accent
	c.SetTheme(withAccent(c.Theme(), a.accent))
	t := c.Theme()
	m := locales.In(c.Locale())
	pal := widget.PaletteFor(t)
	vibrant := c.Vibrancy()
	if vibrant {
		c.Root().Background(ui.Transparent)
	}
	page := a.current()
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		side := ui.Column(c).Width(sidebarWidth).Shrink(0)
		if !vibrant {
			side.Background(t.Surface)
		}
		side.Children(func() { a.sidebar(c, m) })
		ui.Box(c).Width(1).Background(pal.Border)
		// The page scrolls under its title bar, and fades out under the
		// bar's scroll edge, as in macOS 26's own windows.
		ui.Box(c).Grow(1).MinWidth(0).Background(pal.Pane).Clip().Children(func() {
			ui.Scroll(c.Key("page-"+page.ID)).Fill().ScrollbarInsets(headerHeight, 0, 0).Children(func() {
				ui.Column(c).Padding(headerHeight+8, 28, 28, 28).Gap(22).Children(func() {
					page.View(c, m, pal)
					if a.saveErr != nil {
						ui.Text(c, m.SaveFailed(a.saveErr.Error())).FontSize(12).TextColor(t.Danger)
					}
				})
			})
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(headerHeight + 10).PassThrough().
				Material(glass.ScrollEdge{Background: pal.Pane})
			// The bar drags the window, as a title bar does.
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(headerHeight).DragWindow()
			ui.Row(c).Absolute().Top(0).Left(28).Height(headerHeight).AlignItems(ui.Center).DragWindow().Children(func() {
				ui.Text(c, page.Title(m)).FontSize(15).Bold().SingleLine()
			})
		})
	})
}

// current is the page shown, General until the user picks another.
func (a *App) current() Page {
	for _, p := range a.pages {
		if p.ID == a.page {
			return p
		}
	}
	if len(a.pages) == 0 {
		a.pages = []Page{a.generalPage()}
	}
	a.page = a.pages[0].ID
	return a.pages[0]
}

func (a *App) sidebar(c *ui.Context, m locales.Messages) {
	// The traffic lights sit here.
	ui.Box(c).Height(headerHeight).Shrink(0).DragWindow()
	ui.Column(c).Gap(2).Padding(0, 10).Role(ui.RoleList).Label(m.SettingsTitle()).Children(func() {
		setup := false
		for _, p := range a.pages {
			// What the app does and what the user keeps, then what is set
			// up once.
			if p.Setup && !setup {
				setup = true
				ui.Box(c).Height(12)
			}
			a.navItem(c, p.ID, p.Icon, p.Title(m))
		}
	})
}

// navItem is a page in the sidebar: the one shown filled with the accent
// color, as macOS's sidebars show it, and a faint accent behind the one
// hovered.
func (a *App) navItem(c *ui.Context, page string, icon *ui.SVG, label string) {
	t := c.Theme()
	hover := float32(0.08)
	if t.Dark {
		hover = 0.14
	}
	chosen := a.page == page
	item := ui.ButtonBase(c.Key("nav-"+page)).Gap(t.Space(2)).Padding(t.Space(1.5), t.Space(2)).
		Radius(t.Radius).AlignItems(ui.Center).Label(label).Selected(chosen)
	iconColor, textColor := t.Accent, t.Text
	switch {
	case chosen:
		item.Background(t.Accent)
		iconColor, textColor = t.AccentText, t.AccentText
	case item.Hovered():
		item.Background(ui.RGBA(t.Accent.R, t.Accent.G, t.Accent.B, hover))
	}
	item.Children(func() {
		ui.Icon(c, icon).FontSize(16).TextColor(iconColor)
		ui.Text(c, label).Grow(1).SingleLine().TextColor(textColor)
	})
	if item.Clicked() {
		a.page = page
	}
}
