//go:build darwin || windows

package app

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

func TestShortcutEditorLifetimeAndKeys(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	s := newShortcuts(st, nil)
	visible := true
	tt := ui.NewTester(func(c *ui.Context) {
		if visible {
			s.Control(c, shortcut.Dictate)
		}
	}, 560, 460)
	tt.SetLocale("en")
	if err := tt.Click(shortcutName(locales.In("en"), shortcut.Dictate)); err != nil {
		t.Fatal(err)
	}
	if !s.Editing || !tt.HasText("Press a shortcut") {
		t.Fatal(tt.Texts())
	}
	for _, key := range []int64{53, 48, 36, 101} {
		s.KeyDown(key, 0, false)
		s.KeyUp(key)
		tt.Frame()
		tt.Frame()
		if s.Candidate.Key != shortcut.Key(key) || !tt.HasText(s.Candidate.String()) {
			t.Fatal(tt.Texts())
		}
	}
	s.err = errors.New("disk full")
	tt.Frame()
	if !strings.Contains(strings.Join(tt.Texts(), " "), "disk full") {
		t.Fatal(tt.Texts())
	}
	visible = false
	s.CancelEdit()
	tt.Frame()
	tt.Frame()
	visible = true
	tt.Frame()
	if s.Editing || tt.HasText("Press a shortcut") {
		t.Fatal("hidden editor remained active")
	}
	if err := tt.Click(shortcutName(locales.In("en"), shortcut.Dictate)); err != nil {
		t.Fatal(err)
	}
	s.KeyDown(101, 0, false)
	s.Reset()
	s.KeyUp(101)
	tt.Frame()
	if s.Editing || tt.HasText("Press a shortcut") {
		t.Fatal("tap reset left stale editor")
	}
}
func TestRestoreDefaultChecksWholeTable(t *testing.T) {
	st, _ := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	s := newShortcuts(st, nil)
	b := shortcut.Defaults(runtime.GOOS)
	original := b[shortcut.Dictate]
	b[shortcut.Dictate] = shortcut.Binding{Key: "F9"}
	// Translate may take the voice key on both systems, Fn included.
	b[shortcut.Translate] = original
	s.Set(b)
	tt := ui.NewTester(func(c *ui.Context) { s.Control(c, shortcut.Dictate) }, 560, 460)
	tt.SetLocale("en")
	if err := tt.Click(shortcutName(locales.In("en"), shortcut.Dictate)); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Restore default"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(tt.Texts(), " "), "conflicts") {
		t.Fatal(tt.Texts())
	}
	if s.Bindings[shortcut.Dictate].Key != "F9" {
		t.Fatal("reset published before validation")
	}
}
