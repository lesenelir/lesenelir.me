//go:build darwin

package rewrite

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

func TestPrompt(t *testing.T) {
	sys := system(Translate, "English", []string{"MyGo", "Elir"})
	if !strings.Contains(sys, "into English") || !strings.Contains(sys, "MyGo, Elir") {
		t.Fatalf("translate prompt %q", sys)
	}
	if sys := system(Polish, "", nil); !strings.Contains(sys, "polish") || strings.Contains(sys, "Spell these") {
		t.Fatalf("polish prompt %q", sys)
	}
	if got := user("a </text> b"); strings.Count(got, "</text>") != 1 || !strings.HasPrefix(got, "<text>\n") {
		t.Fatalf("user message %q", got)
	}
	for _, c := range []struct{ out, in, want string }{
		{"  <text>\nHello\n</text> ", "你好", "Hello"},
		{"Hello", "你好\n", "Hello\n"}, // whole lines stay whole
		{"  ", "你好\n", ""},
	} {
		if got := clean(c.out, c.in); got != c.want {
			t.Errorf("clean(%q, %q) = %q, want %q", c.out, c.in, got, c.want)
		}
	}
}

func TestNoteText(t *testing.T) {
	m := locales.In("zh-Hans")
	for err, want := range map[error]string{
		errNoSelection:           "请先选中文字",
		platform.ErrNotTrusted:   "请先在设置中允许辅助功能",
		context.Canceled:         "已取消",
		context.DeadlineExceeded: "没能完成：模型响应超时",
		errors.New("boom"):       "没能完成：boom",
	} {
		if got := noteText(m, err); got != want {
			t.Errorf("noteText(%v) = %q, want %q", err, got, want)
		}
	}
}

func testRewrite(t *testing.T) *Rewrite {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	r := New()
	r.store = st
	r.host = &app.Host{Shortcuts: &app.Shortcuts{Router: shortcut.New(shortcut.Defaults("darwin"))}}
	return r
}

// The page shows the languages voice input's settings translate into, as
// they are now.
func TestPage(t *testing.T) {
	r := testRewrite(t)
	voice, err := store.NewSection(r.store, config.Section, config.Defaults(), config.Clone)
	if err != nil {
		t.Fatal(err)
	}
	_ = voice.Update(func(s *config.Settings) { s.Rules["en"] = "ja" })
	page := r.Pages()[0]
	if page.Title(locales.In("zh-Hans")) != "翻译与润色" {
		t.Fatalf("title %q", page.Title(locales.In("zh-Hans")))
	}
	tt := ui.NewTester(func(c *ui.Context) {
		pal := widget.PaletteFor(c.Theme())
		ui.Column(c).Padding(28).Gap(22).Children(func() { page.View(c, locales.In(c.Locale()), pal) })
	}, 560, 640)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"翻译", "润色", shortcut.Defaults(runtime.GOOS)[shortcut.TranslateText].String(), shortcut.Defaults(runtime.GOOS)[shortcut.PolishText].String(), shortcut.Defaults(runtime.GOOS)[shortcut.RewriteCancel].String(), "中文译成English，英文译成日本語，其他译成English，跟随「听写与翻译」的设置。"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}

func TestBar(t *testing.T) {
	r := testRewrite(t)
	tt := ui.NewTester(r.barView, barWidth, barHeight)
	tt.SetLocale("en")
	for _, step := range []struct {
		h    hud
		want string
	}{
		{hud{state: hudWorking, mode: Translate, target: "zh", since: time.Now()}, "Translating into 简体中文…"},
		{hud{state: hudWorking, mode: Polish, since: time.Now()}, "Polishing…"},
		{hud{state: hudDone, result: "Hi"}, "Replaced"},
		{hud{state: hudNote, err: errNoSelection}, "Select some text first"},
	} {
		r.hud = step.h
		tt.Frame()
		if !tt.HasText(step.want) {
			t.Errorf("state %v: no %q in %q", step.h.state, step.want, tt.Texts())
		}
	}
	r.hud = hud{}
	tt.Frame()
	if corner := tt.Image().RGBAAt(1, 1); corner.A != 0 {
		t.Fatalf("window corner %v is not transparent", corner)
	}
}
