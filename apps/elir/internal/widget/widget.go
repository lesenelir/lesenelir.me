// Package widget is the settings window's building blocks, shared by the
// app's pages and its tools': grouped rows as macOS's System Settings shows
// them, keycaps, status badges, a glass dropdown, quiet icon buttons and
// empty states, in colors over the theme's.
package widget

import "github.com/egoist/mygo/ui"

// Icon parses a 16×16 line icon drawn in the text's color.
func Icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	IconCheck   = Icon(`<path d="M3.5 8.5l3 3 6-7"/>`)
	IconChevron = Icon(`<path d="M4.5 6.5L8 10l3.5-3.5"/>`)
	IconCopy    = Icon(`<rect x="5.5" y="5.5" width="8" height="8" rx="1.5"/><path d="M10.5 5.5V4A1.5 1.5 0 0 0 9 2.5H4A1.5 1.5 0 0 0 2.5 4v5A1.5 1.5 0 0 0 4 10.5h1.5"/>`)
	IconTrash   = Icon(`<path d="M2.5 4.5h11M6.5 4.5V3a1 1 0 0 1 1-1h1a1 1 0 0 1 1 1v1.5M4 4.5l.7 8.6a1 1 0 0 0 1 .9h4.6a1 1 0 0 0 1-.9l.7-8.6"/>`)
)

// Palette is the settings window's colors over the theme's.
type Palette struct {
	Pane, Card, Border, Muted ui.Color
}

// PaletteFor returns the palette of the theme's appearance.
func PaletteFor(t *ui.Theme) Palette {
	if t.Dark {
		return Palette{
			Pane:   ui.Hex("#1c1c1e"),
			Card:   ui.RGBA(255, 255, 255, 0.05),
			Border: ui.RGBA(255, 255, 255, 0.08),
			Muted:  t.TextMuted,
		}
	}
	return Palette{
		Pane:   ui.Hex("#f5f5f7"),
		Card:   ui.Hex("#ffffff"),
		Border: ui.RGBA(0, 0, 0, 0.08),
		Muted:  t.TextMuted,
	}
}

// Rows builds the rows of a group, with a hairline between them.
type Rows struct {
	pal Palette
	n   int
}

// Group is a titled card of rows, as System Settings groups its settings.
func Group(c *ui.Context, pal Palette, title string, body func(g *Rows)) {
	ui.Column(c).Gap(8).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(12).Bold().TextColor(pal.Muted).Padding(0, 4)
		}
		ui.Column(c).Radius(10).Background(pal.Card).Border(1, pal.Border).Clip().Children(func() {
			body(&Rows{pal: pal})
		})
	})
}

func (g *Rows) separator(c *ui.Context) {
	if g.n > 0 {
		ui.Box(c).Height(1).Background(g.pal.Border).Margin(0, 0, 0, 14)
	}
	g.n++
}

// Row is a label, with a note under it, and its control on the right.
func (g *Rows) Row(c *ui.Context, label, note string, control func()) {
	g.separator(c)
	ui.Row(c).MinHeight(44).Padding(9, 14).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(2).Children(func() {
			ui.Text(c, label)
			if note != "" {
				ui.Text(c, note).FontSize(11.5).TextColor(g.pal.Muted)
			}
		})
		if control != nil {
			ui.Row(c).Shrink(0).Gap(8).AlignItems(ui.Center).Children(control)
		}
	})
}

// Block is a row that is all content.
func (g *Rows) Block(c *ui.Context, content func()) {
	g.separator(c)
	ui.Column(c).Padding(12, 14).Children(content)
}

// Keycaps shows a shortcut as keys.
func Keycaps(c *ui.Context, pal Palette, keys ...string) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		for _, k := range keys {
			ui.Text(c, k).FontSize(12).Padding(2, 8).Radius(5).Border(1, pal.Border).
				Background(pal.Card).TextColor(pal.Muted).SingleLine()
		}
	})
}

// Badge shows a status in its color.
func Badge(c *ui.Context, text string, color ui.Color) {
	ui.Badge(c, text).Background(color.Alpha(0.14)).TextColor(color)
}

// IconButton is a quiet button showing only an icon, which its tooltip and
// label name.
func IconButton(c *ui.Context, icon *ui.SVG, label string, pal Palette) ui.Element {
	t := c.Theme()
	b := ui.ButtonBase(c).Size(28, 26).Radius(6).Center().Label(label).Tooltip(label)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() { ui.Icon(c, icon).FontSize(15).TextColor(pal.Muted) })
	return b
}

// EmptyState fills an empty page: an icon, if any, over a line.
func EmptyState(c *ui.Context, pal Palette, icon *ui.SVG, text string) {
	ui.Column(c).AlignItems(ui.Center).Gap(10).Padding(40, 0).Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(32).TextColor(pal.Muted)
		}
		ui.Text(c, text).FontSize(12.5).TextColor(pal.Muted).TextAlign(ui.Center)
	})
}

// Note is a quiet line under a page's groups.
func Note(c *ui.Context, pal Palette, text string) {
	ui.Text(c, text).FontSize(11.5).TextColor(pal.Muted)
}
