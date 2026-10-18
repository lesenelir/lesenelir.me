//go:build darwin || windows

package app

import (
	"encoding/json"
	"errors"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

type Shortcuts struct {
	*shortcut.Router
	config  shortcut.Config
	store   *store.Store
	window  *mygo.Window
	editing shortcut.Action
	sides   bool
	saving  bool
	err     error
	loadErr error
	Running func() bool
}

func newShortcuts(st *store.Store, w *mygo.Window) *Shortcuts {
	config, b, err := shortcut.Decode(st.Raw("shortcuts"), runtime.GOOS)
	s := &Shortcuts{Router: shortcut.New(b), config: config, store: st, window: w, loadErr: err}
	if w != nil {
		s.Changed = w.Invalidate
	}
	return s
}
func (s *Shortcuts) CancelEdit() {
	if s == nil {
		return
	}
	s.End()
	s.editing = ""
	s.err = nil
}
func (s *Shortcuts) Control(c *ui.Context, a shortcut.Action) {
	if s == nil {
		return
	}
	m := locales.In(c.Locale())
	ui.Column(c.Key(string(a))).MaxWidth(320).Gap(6).Children(func() {
		if s.editing != a {
			ui.Button(c.Key("edit"), s.Bindings[a].String()).Label(shortcutName(m, a)).Disabled(s.saving).OnClick(func() { s.Begin(); s.editing = a; s.err = nil; s.sides = s.Bindings[a].HasSides() })
			if s.loadErr != nil {
				ui.Text(c, m.ShortcutLoadError(s.loadErr.Error())).FontSize(11).TextColor(c.Theme().Danger)
			}
			return
		}
		ui.Text(c, m.ShortcutRecord()).Bold()
		ui.Text(c, m.ShortcutPhysical()).FontSize(11)
		if s.Running != nil && !s.Running() {
			ui.Text(c, m.ShortcutUnavailable()).FontSize(11).TextColor(c.Theme().Danger)
		}
		candidate := s.Candidate
		if !s.sides {
			candidate = candidate.AnySide()
		}
		if candidate.Key != "" {
			ui.Text(c, candidate.String()).Bold()
		}
		ui.Checkbox(c.Key("sides"), &s.sides, m.ShortcutSides())
		b := shortcut.Clone(s.Bindings)
		b[a] = candidate
		err := shortcut.Validate(b, runtime.GOOS)
		if s.Candidate.Key != "" && err != nil {
			ui.Text(c, shortcutError(m, err)).FontSize(11).TextColor(c.Theme().Danger)
		}
		if a != shortcut.Window && a != shortcut.ScreenshotCancel && shortcut.Printable(candidate) {
			ui.Text(c, m.ShortcutOccupies()).FontSize(11).TextColor(c.Theme().Danger)
		}
		if s.err != nil {
			ui.Text(c, m.SaveFailed(s.err.Error())).FontSize(11).TextColor(c.Theme().Danger)
		}
		ui.Row(c).Gap(6).Children(func() {
			ui.Button(c.Key("save"), m.ShortcutSave()).Disabled(s.saving || err != nil || s.Candidate.Key == "").OnClick(func() { s.save(b) })
			ui.Button(c.Key("cancel"), m.ButtonCancel()).Disabled(s.saving).OnClick(s.CancelEdit)
			ui.Button(c.Key("reset"), m.ShortcutReset()).Disabled(s.saving).OnClick(func() { s.Candidate = shortcut.Defaults(runtime.GOOS)[a]; s.sides = s.Candidate.HasSides() })
		})
	})
}
func (s *Shortcuts) save(b shortcut.Bindings) {
	if s.saving {
		return
	}
	if err := shortcut.Validate(b, runtime.GOOS); err != nil {
		s.err = err
		return
	}
	config := s.config
	config.Set(runtime.GOOS, b)
	raw, err := json.Marshal(config)
	if err != nil {
		s.err = err
		return
	}
	s.saving = true
	go func() {
		err := s.store.SetRaw("shortcuts", raw)
		s.window.Update(func() {
			s.saving = false
			if err != nil {
				s.err = err
				return
			}
			s.config = config
			s.Set(b)
			s.loadErr = nil
			s.CancelEdit()
		})
	}()
}

func (s *Shortcuts) Reset() {
	s.Router.Reset()
	s.editing = ""
	s.err = nil
	if s.Changed != nil {
		s.Changed()
	}
}

func shortcutName(m locales.Messages, a shortcut.Action) string {
	switch a {
	case shortcut.Dictate:
		return m.ShortcutDictate()
	case shortcut.Translate:
		return m.ShortcutTranslate()
	case shortcut.VoiceCancel:
		return m.NavVoice() + " · " + m.ShortcutCancel()
	case shortcut.CycleTarget:
		return m.ShortcutSwitch()
	case shortcut.TranslateText:
		return m.ShortcutTranslateText()
	case shortcut.PolishText:
		return m.ShortcutPolishText()
	case shortcut.RewriteCancel:
		return m.NavRewrite() + " · " + m.ShortcutCancel()
	case shortcut.Capture:
		return m.ShortcutCapture()
	case shortcut.Window:
		return m.ShortcutWindow()
	case shortcut.ScreenshotCancel:
		return m.NavScreenshot() + " · " + m.ShortcutCancel()
	}
	return string(a)
}
func shortcutError(m locales.Messages, err error) string {
	var conflict *shortcut.ConflictError
	if errors.As(err, &conflict) {
		return m.ShortcutConflict(shortcutName(m, conflict.Action), shortcutName(m, conflict.Other))
	}
	return err.Error()
}
