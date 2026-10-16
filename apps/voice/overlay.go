//go:build darwin

package main

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
)

const (
	overlayWidth  = 360
	overlayHeight = 56
	// overlayMargin keeps the voice bar above the Dock.
	overlayMargin = 72
	// messageTime is how long a message stays on the voice bar.
	messageTime = 2500 * time.Millisecond
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
		case s.Message != "":
			if !a.overlay.IsVisible() {
				a.placeOverlay()
				a.overlay.ShowInactive()
			}
			time.AfterFunc(messageTime, func() {
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
	s := a.status
	bg := ui.RGBA(24, 24, 27, 0.92)
	text := ui.Hex("#fafafa")
	muted := ui.Hex("#a1a1aa")
	ui.Row(c).Fill().Padding(0, 16).Gap(12).AlignItems(ui.Center).Radius(overlayHeight / 2).Background(bg).Children(func() {
		switch s.State {
		case controller.Recording:
			ui.Box(c).Size(10, 10).Radius(5).Background(t.Danger)
			a.modeChip(c, s, text)
			a.levelMeter(c, text)
			ui.Spacer(c)
			hint := "Fn 完成 · Esc 取消"
			if s.Mode == prompt.Translate {
				hint = "Tab 换语言 · Esc 取消"
			}
			ui.Text(c, hint).FontSize(11).TextColor(muted).NoWrap()
		case controller.Transcribing, controller.Rewriting, controller.Pasting:
			ui.Spinner(c).Size(16, 16)
			ui.Text(c, stageText(s)).FontSize(13).TextColor(text).NoWrap()
			ui.Spacer(c)
			ui.Text(c, "Esc 取消").FontSize(11).TextColor(muted).NoWrap()
		default:
			ui.Text(c, s.Message).FontSize(13).TextColor(text).MaxLines(1).Ellipsis("…").Grow(1)
		}
	})
}

func (a *app) modeChip(c *ui.Context, s controller.Status, text ui.Color) {
	if s.Mode != prompt.Translate {
		ui.Text(c, "听写").FontSize(13).Bold().TextColor(text).NoWrap()
		return
	}
	chip := ui.Row(c).Gap(4).Padding(3, 8).Radius(6).Background(ui.RGBA(255, 255, 255, 0.12)).
		Cursor(ui.CursorPointer).Tooltip("点击切换目标语言")
	chip.Children(func() {
		ui.Text(c, "翻译 → "+prompt.LanguageFor(s.Target).Label).FontSize(13).Bold().TextColor(text).NoWrap()
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

func stageText(s controller.Status) string {
	switch {
	case s.State == controller.Transcribing:
		return "转写中…"
	case s.State == controller.Rewriting && s.Mode == prompt.Translate:
		return "翻译成 " + prompt.LanguageFor(s.Target).Label + "…"
	case s.State == controller.Rewriting:
		return "整理中…"
	}
	return "粘贴中…"
}
