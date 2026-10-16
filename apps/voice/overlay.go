//go:build darwin

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

// The voice bar is a dark capsule at the bottom of the screen, as macOS's
// own HUDs: dark in either appearance, so that it reads over any window.
// The window is larger than the capsule to leave room for its shadow.
const (
	overlayWidth  = 340
	overlayHeight = 76
	// overlayMargin keeps the voice bar above the Dock.
	overlayMargin = 56
	capsuleHeight = 40
	// noteTime is how long a note stays on the voice bar, timingsTime how
	// long the timings of a session that went well do.
	noteTime    = 2500 * time.Millisecond
	timingsTime = 1800 * time.Millisecond

	waveBars = 20
	// waveStep is how often the waveform moves on by a bar.
	waveStep = 45 * time.Millisecond
)

var (
	capsuleFill   = ui.RGBA(20, 20, 22, 0.94)
	capsuleBorder = ui.RGBA(255, 255, 255, 0.1)
	capsuleShadow = ui.RGBA(0, 0, 0, 0.35)
	barText       = ui.RGBA(255, 255, 255, 0.92)
	barMuted      = ui.RGBA(255, 255, 255, 0.5)
	recordingRed  = ui.Hex("#ff453a")
	doneGreen     = ui.Hex("#32d74b")

	iconCheck = icon(`<path d="M3.5 8.5l3 3 6-7"/>`)
)

// showsTimings reports whether the voice bar shows the timings of a session
// that went well.
func showsTimings(s controller.Status, hidden bool) bool {
	return !hidden && s.State == controller.Idle && s.Note == controller.NoteNone && s.Timings != (controller.Timings{})
}

// formatDuration writes a duration as seconds: 0.8s, 1.2s, 14s.
func formatDuration(d time.Duration) string {
	if d < 10*time.Second {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.0fs", d.Seconds())
}

// timingsText says how long a session's requests took, on the voice bar.
func timingsText(m locales.Messages, s controller.Status) string {
	asr, llm := formatDuration(s.Timings.Transcribe), formatDuration(s.Timings.Rewrite)
	if s.Mode == prompt.Translate {
		return m.TimingsTranslate(asr, llm)
	}
	return m.TimingsDictate(asr, llm)
}

// onStatus shows the controller's status on the voice bar. It runs on the
// controller's goroutine.
func (a *app) onStatus(s controller.Status) {
	a.overlay.Update(func() {
		if s.State == controller.Recording && a.status.State != controller.Recording {
			a.levels = [waveBars]float32{}
		}
		a.status = s
		a.hideSeq++
		seq := a.hideSeq
		switch {
		case s.State != controller.Idle:
			if !a.overlay.IsVisible() {
				a.placeOverlay()
				a.overlay.ShowInactive()
			}
		case s.Note != controller.NoteNone || showsTimings(s, a.store.Get().HideTimings):
			if !a.overlay.IsVisible() {
				a.placeOverlay()
				a.overlay.ShowInactive()
			}
			linger := noteTime
			if s.Note == controller.NoteNone {
				linger = timingsTime
			}
			time.AfterFunc(linger, func() {
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
	c.Root().Background(ui.Transparent)
	m := locales.In(c.Locale())
	s := a.status
	// The capsule sits at the bottom center, as wide as what it shows.
	ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.End).Padding(0, 0, 18, 0).Children(func() {
		capsule := ui.Row(c).Height(capsuleHeight).Padding(0, 14).Gap(10).AlignItems(ui.Center).
			Radius(capsuleHeight/2).Background(capsuleFill).Border(1, capsuleBorder).
			Shadow(0, 6, 18, 0, capsuleShadow)
		capsule.Children(func() {
			switch s.State {
			case controller.Recording:
				pulsingDot(c)
				if s.Mode == prompt.Translate {
					a.targetChip(c, m, s)
				}
				a.waveform(c)
				overlayKeycap(c, "esc")
			case controller.Transcribing, controller.Rewriting, controller.Pasting:
				thinkingDots(c)
				ui.Text(c, stageText(m, s)).FontSize(12.5).TextColor(barText).NoWrap()
				if !s.Since.IsZero() {
					ui.Text(c, formatDuration(time.Since(s.Since))).FontSize(12.5).TextColor(barMuted).
						FontFeatures("tnum").NoWrap()
					c.After(100 * time.Millisecond)
				}
			case controller.Idle:
				if s.Note == controller.NoteNone {
					ui.Icon(c, iconCheck).Size(14, 14).TextColor(doneGreen)
					ui.Text(c, timingsText(m, s)).FontSize(12.5).TextColor(barText).NoWrap()
					return
				}
				fallthrough
			default:
				ui.Text(c, noteText(m, s)).FontSize(12.5).TextColor(barText).MaxWidth(280).
					MaxLines(1).Ellipsis("…")
			}
		})
	})
}

// pulsingDot is the red recording light, breathing.
func pulsingDot(c *ui.Context) {
	dot := ui.Box(c.Key("rec-dot")).Size(8, 8).Radius(4).Background(recordingRed)
	p := dot.Loop("pulse", 1400*time.Millisecond, ui.EaseInOut)
	dot.Opacity(0.45 + 0.55*triangle(p))
}

// thinkingDots are three dots lighting up in turn while a request runs.
func thinkingDots(c *ui.Context) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		for i := range 3 {
			d := ui.Box(c.Key(i)).Size(5, 5).Radius(2.5).Background(barText)
			p := d.Loop("thinking", 900*time.Millisecond, ui.Linear)
			phase := p - float32(i)/3
			if phase < 0 {
				phase++
			}
			d.Opacity(0.25 + 0.75*triangle(phase))
		}
	})
}

// triangle rises from 0 to 1 and falls back as p goes from 0 to 1.
func triangle(p float32) float32 {
	if p < 0.5 {
		return p * 2
	}
	return 2 - p*2
}

// waveform scrolls the loudness of the last second from right to left.
func (a *app) waveform(c *ui.Context) {
	if now := time.Now(); now.Sub(a.levelAt) >= waveStep {
		copy(a.levels[:], a.levels[1:])
		a.levels[waveBars-1] = a.rec.Level()
		a.levelAt = now
	}
	ui.Row(c).Gap(2).AlignItems(ui.Center).Height(22).Children(func() {
		for i, level := range a.levels {
			// Older bars fade a little, so the newest sound reads first.
			fade := 0.45 + 0.55*float32(i)/float32(waveBars-1)
			ui.Box(c.Key(i)).Width(2.5).Height(3 + 19*level).Radius(1.25).
				Background(ui.RGBA(255, 255, 255, fade))
		}
	})
	c.After(waveStep)
}

// targetChip shows the translation target; a click moves to the next one.
func (a *app) targetChip(c *ui.Context, m locales.Messages, s controller.Status) {
	chip := ui.Row(c.Key("target")).Padding(2, 8).Radius(6).Background(ui.RGBA(255, 255, 255, 0.12)).
		Cursor(ui.CursorPointer).Tooltip(m.SwitchLanguageTooltip())
	chip.Children(func() {
		ui.Text(c, "→ "+prompt.LanguageFor(s.Target).Label).FontSize(12).Bold().TextColor(barText).NoWrap()
	})
	if chip.Clicked() {
		a.ctrl.SetTarget(a.store.Get().NextTarget(s.Target))
	}
}

func overlayKeycap(c *ui.Context, key string) {
	ui.Text(c, key).FontSize(10.5).Padding(1, 6).Radius(5).Border(1, ui.RGBA(255, 255, 255, 0.2)).
		TextColor(barMuted).SingleLine()
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
