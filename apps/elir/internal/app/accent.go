//go:build darwin

package app

import (
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// accentChoice is an accent color the settings offer, in Apple's system
// colors for the light and the dark appearance. The zero choice follows
// the color chosen in macOS.
type accentChoice struct {
	id          string
	light, dark ui.Color
}

var accentChoices = []accentChoice{
	{id: ""},
	{"blue", ui.Hex("#007aff"), ui.Hex("#0a84ff")},
	{"green", ui.Hex("#34c759"), ui.Hex("#30d158")},
	{"red", ui.Hex("#ff3b30"), ui.Hex("#ff453a")},
	{"yellow", ui.Hex("#ffcc00"), ui.Hex("#ffd60a")},
	{"purple", ui.Hex("#af52de"), ui.Hex("#bf5af2")},
	{"orange", ui.Hex("#ff9500"), ui.Hex("#ff9f0a")},
	{"pink", ui.Hex("#ff2d55"), ui.Hex("#ff375f")},
	{"brown", ui.Hex("#a2845e"), ui.Hex("#ac8e68")},
}

func accentName(m locales.Messages, id string) string {
	switch id {
	case "blue":
		return m.AccentBlue()
	case "green":
		return m.AccentGreen()
	case "red":
		return m.AccentRed()
	case "yellow":
		return m.AccentYellow()
	case "purple":
		return m.AccentPurple()
	case "orange":
		return m.AccentOrange()
	case "pink":
		return m.AccentPink()
	case "brown":
		return m.AccentBrown()
	}
	return m.AccentSystem()
}

// accentColor returns the color of a choice in the appearance, and false
// for the system's.
func accentColor(id string, dark bool) (ui.Color, bool) {
	for _, a := range accentChoices {
		if a.id == id && id != "" {
			if dark {
				return a.dark, true
			}
			return a.light, true
		}
	}
	return ui.Color{}, false
}

// withAccent returns the theme with the accent color chosen in the
// settings, shaded as MyGo shades the system's: hovered, pressed, the
// selection, the focus ring, and text on it black where white would not
// stand out.
func withAccent(t *ui.Theme, id string) *ui.Theme {
	a, ok := accentColor(id, t.Dark)
	if !ok {
		return t
	}
	th := *t
	black, white := ui.Hex("#000000"), ui.Hex("#ffffff")
	th.Accent = a
	if th.Dark {
		th.AccentHover, th.AccentPressed = a.Mix(white, 0.2), a.Mix(black, 0.15)
		th.Selection, th.Focus = a.Alpha(0.4), a.Mix(white, 0.2).Alpha(0.6)
	} else {
		th.AccentHover, th.AccentPressed = a.Mix(black, 0.12), a.Mix(black, 0.25)
		th.Selection, th.Focus = a.Alpha(0.25), a.Alpha(0.55)
	}
	th.AccentText = white
	if luminance(a) > 165 {
		th.AccentText = ui.Hex("#18181b")
	}
	return &th
}

func luminance(c ui.Color) float32 {
	return 0.299*float32(c.R) + 0.587*float32(c.G) + 0.114*float32(c.B)
}

// accentSelect picks the accent color, each choice shown with its swatch.
// system is the accent macOS has, for the "System" swatch.
func (a *App) accentSelect(c *ui.Context, m locales.Messages, system ui.Color) {
	t := c.Theme()
	ids := make([]string, len(accentChoices))
	for i, choice := range accentChoices {
		ids[i] = choice.id
	}
	swatch := func(id string) {
		col, ok := accentColor(id, t.Dark)
		if !ok {
			col = system
		}
		ui.Box(c).Size(12, 12).Radius(6).Background(col).Border(1, ui.RGBA(0, 0, 0, 0.12))
	}
	widget.Dropdown(c.Key("accent"), m.FieldAccent(), &a.accent, ids, func(id string) string { return accentName(m, id) }, swatch, 180, func() {
		a.update(func(s *Settings) { s.AccentColor = a.accent })
	})
}
