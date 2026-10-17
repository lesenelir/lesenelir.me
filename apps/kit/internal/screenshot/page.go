//go:build darwin

package screenshot

import (
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

var iconScreenshot = widget.Icon(`<path d="M2.5 5.5v-2a1 1 0 0 1 1-1h2M10.5 2.5h2a1 1 0 0 1 1 1v2M13.5 10.5v2a1 1 0 0 1-1 1h-2M5.5 13.5h-2a1 1 0 0 1-1-1v-2"/><circle cx="8" cy="8" r="2.2"/>`)

// shotBatch is how many screenshots the page shows before "Show more".
const shotBatch = 30

// shotMotion grows screenshots in as they come and collapses them as they
// go, the others sliding to their places.
var shotMotion = ui.ElementTransition{
	Enter: &ui.Motion{Collapse: true},
	Exit:  &ui.Motion{Collapse: true},
}

// Pages is the tool's one settings page: the shortcut, how long
// screenshots are kept, and the screenshots.
func (s *Screenshot) Pages() []app.Page {
	return []app.Page{{ID: "screenshot", Icon: iconScreenshot, Title: locales.Messages.NavScreenshot, View: s.page}}
}

// keepLabel names a choice of how long screenshots are kept.
func keepLabel(m locales.Messages, days int) string {
	if days <= 0 {
		return m.KeepForever()
	}
	return m.KeepDays(days)
}

func (s *Screenshot) page(c *ui.Context, m locales.Messages, pal widget.Palette) {
	t := c.Theme()
	widget.Group(c, pal, m.SectionCapture(), func(g *widget.Rows) {
		g.Row(c, m.ShortcutCapture(), m.CaptureNote(), func() { widget.Keycaps(c, pal, "⌘", "⇧", "2") })
		g.Row(c, m.ShortcutWindow(), "", func() { widget.Keycaps(c, pal, "space") })
		g.Row(c, m.ShortcutCancel(), "", func() { widget.Keycaps(c, pal, "esc") })
	})
	// A terminal's ⌘V pastes text alone; Claude Code reads an image
	// itself on ⌃V.
	widget.Group(c, pal, m.SectionPaste(), func(g *widget.Rows) {
		g.Row(c, m.PasteApps(), "", func() { widget.Keycaps(c, pal, "⌘", "V") })
		g.Row(c, m.PasteTerminals(), m.PasteTerminalsNote(), func() { widget.Keycaps(c, pal, "⌃", "V") })
	})
	widget.Group(c, pal, m.SectionScreenshots(), func(g *widget.Rows) {
		g.Row(c, m.FieldKeep(), m.KeepNote(), func() {
			widget.Dropdown(c.Key("keep"), m.FieldKeep(), &s.keepDays, keepChoices,
				func(days int) string { return keepLabel(m, days) }, nil, 140, s.setKeep)
		})
		g.Row(c, m.ScreenshotUsage(len(s.shots), widget.FormatSize(Size(s.shots))), "", func() {
			if ui.Button(c.Key("clear-screenshots"), m.ButtonClearHistory()).Disabled(len(s.shots) == 0).Clicked() {
				s.confirm = true
			}
		})
	})
	if ui.AlertDialog(c, &s.confirm, m.ClearScreenshotsTitle(), m.ClearScreenshotsMessage(), m.ButtonCancel(), m.ButtonClearHistory()) == 1 {
		s.clear()
	}
	if s.err != nil {
		ui.Text(c, m.ScreenshotFailed(s.err.Error())).FontSize(12).TextColor(t.Danger)
	}
	if len(s.shots) == 0 {
		widget.EmptyState(c, pal, iconScreenshot, m.ScreenshotsEmpty())
		return
	}
	shown := min(len(s.shots), max(s.shown, shotBatch))
	var deleted *Shot
	ui.Grid(c).Columns(3).GapX(12).GapY(12).Children(func() {
		for _, shot := range s.shots[:shown] {
			if s.card(c.Key("s-"+shot.Path), m, pal, shot) {
				deleted = &shot
			}
		}
	})
	if deleted != nil {
		s.delete(c, m, *deleted)
	}
	if shown < len(s.shots) {
		if ui.Button(c.Key("show-more"), m.ShowMore()).Clicked() {
			s.shown = shown + shotBatch
		}
	}
}

// card shows a screenshot and reports whether the user deleted it.
func (s *Screenshot) card(c *ui.Context, m locales.Messages, pal widget.Palette, shot Shot) (deleted bool) {
	copyIt := func() {
		if err := copyShot(shot.Path); err != nil {
			s.err = err
			return
		}
		c.Toast(m.Copied())
	}
	card := ui.Column(c).Radius(10).Background(pal.Card).Border(1, pal.Border).Padding(6).Gap(6).
		MinWidth(0).Transition(shotMotion)
	card.ContextMenu(func(menu *ui.Menu) {
		if menu.Item(m.ButtonCopy()).Chosen() {
			copyIt()
		}
		if menu.Item(m.ShowInFinder()).Chosen() {
			mygo.Shell.ShowItemInFolder(shot.Path)
		}
		menu.Separator()
		if menu.Item(m.ButtonDelete()).Chosen() {
			deleted = true
		}
	})
	card.Children(func() {
		img := ui.Box(c.Key("image")).AspectRatio(16.0 / 10).Radius(6).Clip().Background(pal.Border).
			Cursor(ui.CursorPointer).Label(m.ScreenshotThumbnail()).Tooltip(m.ThumbnailHint())
		img.DragData(fileData(shot.Path))
		img.Children(func() {
			if thumb := s.thumb(shot.Path); thumb != nil {
				ui.Image(c, thumb).Fill().Fit(ui.Cover)
			}
		})
		if img.Clicked() {
			_ = mygo.Shell.OpenPath(shot.Path)
		}
		ui.Row(c).Gap(2).AlignItems(ui.Center).Padding(0, 0, 0, 4).Children(func() {
			ui.Text(c, widget.FormatTime(c.Locale(), shot.Time)).FontSize(11.5).TextColor(pal.Muted).
				SingleLine().Grow(1).MinWidth(0)
			if widget.IconButton(c.Key("copy"), widget.IconCopy, m.ButtonCopy(), pal).Clicked() {
				copyIt()
			}
			if widget.IconButton(c.Key("delete"), widget.IconTrash, m.ButtonDelete(), pal).Clicked() {
				deleted = true
			}
		})
	})
	return deleted
}

// delete deletes a screenshot, and offers to put it back.
func (s *Screenshot) delete(c *ui.Context, m locales.Messages, shot Shot) {
	data, err := os.ReadFile(shot.Path)
	if err == nil {
		err = s.lib.Delete(shot.Path)
	}
	if err != nil {
		s.err = err
		return
	}
	s.remove(shot.Path)
	c.ToastAction(m.Deleted(), m.Undo(), func() {
		if err := s.lib.Restore(shot, data); err != nil {
			s.err = err
			return
		}
		s.reload()
	})
}

func (s *Screenshot) clear() {
	if err := s.lib.Clear(); err != nil {
		s.err = err
	}
	s.reload()
	s.shown = 0
	if s.pop.thumb != nil {
		s.hidePopup()
	}
}
