//go:build darwin || windows

package rewrite

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	shortcuts "github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// The bar is a dark capsule at the bottom of the screen, as voice input's:
// it says what runs, then that the selection was replaced, with a button
// to copy the result where it could not be, as in a web page.
const (
	barWidth  = 640
	barHeight = 76
	// barMargin keeps the bar above the Dock.
	barMargin     = 56
	capsuleHeight = 40
	doneTime      = 3 * time.Second
	noteTime      = 2500 * time.Millisecond
)

var (
	capsuleFill   = ui.RGBA(20, 20, 22, 0.94)
	capsuleBorder = ui.RGBA(255, 255, 255, 0.1)
	capsuleShadow = ui.RGBA(0, 0, 0, 0.35)
	barText       = ui.RGBA(255, 255, 255, 0.92)
	barMuted      = ui.RGBA(255, 255, 255, 0.5)
	doneGreen     = ui.Hex("#32d74b")
)

type hudState int

const (
	hudHidden hudState = iota
	hudWorking
	hudDone
	hudNote
)

// hud is what the bar shows.
type hud struct {
	state   hudState
	mode    Mode
	target  string // the language translated into
	since   time.Time
	result  string
	err     error // a note's
	copied  bool
	hovered bool
	seq     int // counts what the bar showed, so a stale timer hides none
}

// show shows h on the bar, from any goroutine.
func (r *Rewrite) show(h hud) {
	r.bar.Update(func() {
		h.seq = r.hud.seq + 1
		r.hud = h
		if !r.bar.IsVisible() {
			area := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint()).WorkArea
			r.bar.SetPosition(area.X+(area.Width-barWidth)/2, area.Y+area.Height-barHeight-barMargin)
			r.bar.ShowInactive()
		}
		switch h.state {
		case hudDone:
			r.hideLater(h.seq, doneTime)
		case hudNote:
			r.hideLater(h.seq, noteTime)
		}
	})
}

// hideLater hides the bar after a while, or later while the pointer is on
// it.
func (r *Rewrite) hideLater(seq int, after time.Duration) {
	time.AfterFunc(after, func() {
		r.bar.Update(func() {
			switch {
			case r.hud.seq != seq:
			case r.hud.hovered:
				r.hideLater(seq, after)
			default:
				r.hud = hud{seq: seq}
				r.bar.Hide()
			}
		})
	})
}

func (r *Rewrite) barView(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	h := &r.hud
	if h.state == hudHidden {
		return
	}
	m := locales.In(c.Locale())
	ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.End).Padding(0, 0, 18, 0).Children(func() {
		capsule := ui.Row(c.Key("capsule")).Height(capsuleHeight).Padding(0, 14).Gap(10).AlignItems(ui.Center).
			Radius(capsuleHeight/2).Background(capsuleFill).Border(1, capsuleBorder).
			Shadow(0, 6, 18, 0, capsuleShadow)
		capsule.Children(func() {
			switch h.state {
			case hudWorking:
				ui.Spinner(c).TextColor(barText).Label(workingText(m, *h))
				ui.Text(c, workingText(m, *h)).FontSize(12.5).TextColor(barText).NoWrap()
				ui.Text(c, formatDuration(time.Since(h.since))).FontSize(12.5).TextColor(barMuted).
					FontFeatures("tnum").NoWrap()
				c.After(100 * time.Millisecond)
				keycap(c, r.host.Shortcuts.Bindings[shortcuts.RewriteCancel].String())
			case hudDone:
				ui.Icon(c, widget.IconCheck).Size(14, 14).TextColor(doneGreen)
				ui.Text(c, m.RewriteReplaced()).FontSize(12.5).TextColor(barText).NoWrap()
				label := m.ButtonCopy()
				if h.copied {
					label = m.Copied()
				}
				chip := ui.Row(c.Key("copy")).Padding(2, 8).Radius(6).Background(ui.RGBA(255, 255, 255, 0.12)).
					Cursor(ui.CursorPointer).Label(label)
				chip.Children(func() { ui.Text(c, label).FontSize(12).Bold().TextColor(barText).NoWrap() })
				if chip.Clicked() {
					mygo.Clipboard.WriteText(h.result)
					h.copied = true
				}
			default:
				ui.Text(c, noteText(m, h.err)).FontSize(12.5).TextColor(barText).MaxWidth(barWidth - 60).
					MaxLines(1).Ellipsis("…")
			}
		})
		h.hovered = capsule.Hovered()
	})
}

func keycap(c *ui.Context, key string) {
	ui.Text(c, key).FontSize(10.5).Padding(1, 6).Radius(5).Border(1, ui.RGBA(255, 255, 255, 0.2)).
		TextColor(barMuted).SingleLine()
}

// workingText says what runs.
func workingText(m locales.Messages, h hud) string {
	if h.mode == Translate {
		return m.StageTranslating(prompt.LanguageFor(h.target).Label)
	}
	return m.StagePolishing()
}

// noteText words how a rewrite ended without a result.
func noteText(m locales.Messages, err error) string {
	var api *openrouter.Error
	switch {
	case errors.Is(err, errNoSelection):
		return m.NoteNoSelection()
	case errors.Is(err, errNoKey), errors.Is(err, openrouter.ErrNoKey):
		return m.NoteNoKey()
	case runtime.GOOS == "darwin" && errors.Is(err, platform.ErrNotTrusted):
		return m.NoteNeedsAccessibility()
	case errors.Is(err, context.Canceled):
		return m.NoteCancelled()
	case errors.Is(err, context.DeadlineExceeded):
		return m.RewriteFailed(m.RewriteTimedOut())
	case errors.As(err, &api):
		return m.RewriteFailed(api.Error())
	case err != nil:
		return m.RewriteFailed(err.Error())
	}
	return ""
}

// formatDuration writes a duration as seconds: 0.8s, 1.2s, 14s.
func formatDuration(d time.Duration) string {
	if d < 10*time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}
