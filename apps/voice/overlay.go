//go:build darwin

package main

import (
	"errors"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

const (
	overlayWidth  = 360
	overlayHeight = 56
	// overlayMargin keeps the voice bar above the Dock.
	overlayMargin = 72
	// noteTime is how long a note stays on the voice bar.
	noteTime = 2500 * time.Millisecond
)

// barWeights shape the level meter into a little waveform.
var barWeights = []float32{0.35, 0.6, 0.85, 1, 0.8, 0.55, 0.3}

// onStatus shows the controller's status on the voice bar. It runs on the
// controller's goroutine.
func (a *app) onStatus(s controller.Status) {
	a.overlay.Update(func() {
		a.status = s
		a.hideSeq++
		seq := a.hideSeq
		switch {
		case s.State != controller.Idle:
			if !a.overlay.IsVisible() {
				a.placeOverlay()
				a.overlay.ShowInactive()
			}
		case s.Note != controller.NoteNone:
			if !a.overlay.IsVisible() {
				a.placeOverlay()
				a.overlay.ShowInactive()
			}
			time.AfterFunc(noteTime, func() {
				a.overlay.Update(func() {
					if a.hideSeq == seq {
						a.overlay.Hide()
					}
				})
			})
		default:
			a.overlay.Hide()
		}
	})
	// The settings window shows the last result.
	a.settings.Invalidate()
}

// placeOverlay puts the voice bar at the bottom center of the display the
// pointer is on.
func (a *app) placeOverlay() {
	area := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint()).WorkArea
	a.overlay.SetPosition(area.X+(area.Width-overlayWidth)/2, area.Y+area.Height-overlayHeight-overlayMargin)
}

func (a *app) overlayView(c *ui.Context) {
	t := c.Theme()
	m := locales.In(c.Locale())
	s := a.status
	bg := ui.RGBA(24, 24, 27, 0.92)
	text := ui.Hex("#fafafa")
	muted := ui.Hex("#a1a1aa")
	ui.Row(c).Fill().Padding(0, 16).Gap(12).AlignItems(ui.Center).Radius(overlayHeight / 2).Background(bg).Children(func() {
		switch s.State {
		case controller.Recording:
			ui.Box(c).Size(10, 10).Radius(5).Background(t.Danger)
			a.modeChip(c, m, s, text)
			a.levelMeter(c, text)
			ui.Spacer(c)
			hint := m.HintDictate()
			if s.Mode == prompt.Translate {
				hint = m.HintTranslate()
			}
			ui.Text(c, hint).FontSize(11).TextColor(muted).NoWrap()
		case controller.Transcribing, controller.Rewriting, controller.Pasting:
			ui.Spinner(c).Size(16, 16)
			ui.Text(c, stageText(m, s)).FontSize(13).TextColor(text).NoWrap()
			ui.Spacer(c)
			ui.Text(c, m.HintCancel()).FontSize(11).TextColor(muted).NoWrap()
		default:
			ui.Text(c, noteText(m, s)).FontSize(13).TextColor(text).MaxLines(1).Ellipsis("…").Grow(1)
		}
	})
}

func (a *app) modeChip(c *ui.Context, m locales.Messages, s controller.Status, text ui.Color) {
	if s.Mode != prompt.Translate {
		ui.Text(c, m.ModeDictate()).FontSize(13).Bold().TextColor(text).NoWrap()
		return
	}
	chip := ui.Row(c).Gap(4).Padding(3, 8).Radius(6).Background(ui.RGBA(255, 255, 255, 0.12)).
		Cursor(ui.CursorPointer).Tooltip(m.SwitchLanguageTooltip())
	chip.Children(func() {
		ui.Text(c, m.ModeTranslate(prompt.LanguageFor(s.Target).Label)).FontSize(13).Bold().TextColor(text).NoWrap()
	})
	if chip.Clicked() {
		a.ctrl.SetTarget(a.store.Get().NextTarget(s.Target))
	}
}

func (a *app) levelMeter(c *ui.Context, color ui.Color) {
	level := a.rec.Level()
	ui.Row(c).Gap(3).AlignItems(ui.Center).Height(24).Children(func() {
		for _, w := range barWeights {
			h := 3 + 21*level*w
			ui.Box(c).Width(3).Height(h).Radius(1.5).Background(color)
		}
	})
	// Keep the meter moving while recording.
	c.After(50 * time.Millisecond)
}

func stageText(m locales.Messages, s controller.Status) string {
	switch {
	case s.State == controller.Transcribing:
		return m.StageTranscribing()
	case s.State == controller.Rewriting && s.Mode == prompt.Translate:
		return m.StageTranslating(prompt.LanguageFor(s.Target).Label)
	case s.State == controller.Rewriting:
		return m.StageCleaning()
	}
	return m.StagePasting()
}

// noteText words the status's note, "" when there is none.
func noteText(m locales.Messages, s controller.Status) string {
	detail := ""
	if s.Err != nil {
		detail = s.Err.Error()
	}
	switch s.Note {
	case controller.NoteTooShort:
		return m.NoteTooShort()
	case controller.NoteSilent:
		return m.NoteSilent()
	case controller.NoteCancelled:
		return m.NoteCancelled()
	case controller.NoteNothing:
		return m.NoteNothing()
	case controller.NoteRawPasted:
		return m.NoteRawPasted()
	case controller.NoteNoKey:
		return m.NoteNoKey()
	case controller.NoteASRFailed:
		return m.NoteAsrFailed(detail)
	case controller.NotePasteFailed:
		if errors.Is(s.Err, darwin.ErrNotTrusted) {
			return m.NotePasteNeedsAccessibility()
		}
		return m.NotePasteFailed(detail)
	case controller.NoteMicFailed:
		return m.NoteMicFailed(detail)
	}
	return ""
}
