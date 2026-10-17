//go:build darwin

package main

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

// dropdown is a select whose menu is a pane of Liquid Glass over the page:
// the cards under it show through, frosted, as macOS 26's menus do. The
// choice is checked, the option the pointer or the arrows are on is filled
// with the accent. lead, if set, draws before a label, as a swatch.
func dropdown[T comparable](c *ui.Context, selected *T, options []T, label func(T) string, lead func(T), width float32, onChange func()) {
	t := c.Theme()
	sel := ui.SelectBase(c, selected)
	sel.Trigger.Width(width).Gap(8).Padding(5, 10).Radius(t.Radius).Background(t.Surface).
		Border(1, t.Border).AlignItems(ui.Center).Children(func() {
		if lead != nil {
			lead(*selected)
		}
		ui.Text(c, label(*selected)).Grow(1).SingleLine()
		ui.Icon(c, iconChevron).FontSize(14).TextColor(t.TextMuted)
	})
	if onChange != nil {
		sel.Trigger.OnChange(onChange)
	}
	sel.Popup(func(panel ui.Element) {
		panel.Margin(6, 0, 0, 0).Padding(5).Radius(14).Width(width).
			Material(glass.Glass{}).Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.18))
		for _, option := range options {
			item := sel.Item(option).Gap(6).Padding(4, 8).Radius(8).AlignItems(ui.Center)
			text := t.Text
			if item.Highlighted() {
				item.Background(t.Accent)
				text = t.AccentText
			}
			item.Children(func() {
				check := ui.Box(c).Size(14, 14).Center()
				if option == *selected {
					check.Children(func() { ui.Icon(c, iconCheck).FontSize(13).TextColor(text) })
				}
				if lead != nil {
					lead(option)
				}
				ui.Text(c, label(option)).Grow(1).SingleLine().TextColor(text)
			})
		}
	})
}
