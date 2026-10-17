//go:build darwin

package screenshot

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// pageWidth is the width of a page in the settings window, right of its
// sidebar.
const pageWidth = 560

func testScreenshot(t *testing.T) *Screenshot {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := New()
	s.host = &app.Host{Store: st, DataDir: dir}
	if s.settings, err = store.NewSection(st, section, defaults(), nil); err != nil {
		t.Fatal(err)
	}
	s.keepDays = s.settings.Get().KeepDays
	if s.lib, err = OpenLibrary(filepath.Join(dir, section)); err != nil {
		t.Fatal(err)
	}
	return s
}

// addShot keeps a screenshot taken age ago.
func addShot(t *testing.T, s *Screenshot, age time.Duration) Shot {
	t.Helper()
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 64, 40)))
	when := time.Now().Add(-age)
	path := s.lib.NewPath(when)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(path, when, when)
	return Shot{Path: path, Time: when, Size: int64(b.Len())}
}

// pageView shows the tool's page as the settings window does, on its pane.
func pageView(s *Screenshot) func(c *ui.Context) {
	page := s.Pages()[0]
	return func(c *ui.Context) {
		pal := widget.PaletteFor(c.Theme())
		ui.Box(c).Fill().Background(pal.Pane).Children(func() {
			ui.Scroll(c.Key("page")).Fill().Children(func() {
				ui.Column(c).Padding(28).Gap(22).Children(func() {
					page.View(c, locales.In(c.Locale()), pal)
				})
			})
		})
	}
}

func TestPageEmpty(t *testing.T) {
	s := testScreenshot(t)
	if title := s.Pages()[0].Title(locales.In("zh-Hans")); title != "截图" {
		t.Fatalf("title %q", title)
	}
	tt := ui.NewTester(pageView(s), pageWidth, 900)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"截取区域或窗口", "⌘", "2", "截取指针下方的窗口", "粘贴到 Claude Code 等终端", "⌃", "截图保留", "7 天", "共 0 张 · 0 B", "还没有截图。使用上方截图快捷键截一张。"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}

func TestPageListsDeletesAndUndoes(t *testing.T) {
	s := testScreenshot(t)
	addShot(t, s, time.Hour)
	newest := addShot(t, s, time.Minute)
	s.reload()
	tt := ui.NewTester(pageView(s), pageWidth, 900)
	tt.SetLocale("en")
	usage := "2 screenshots · " + widget.FormatSize(Size(s.shots))
	if !tt.HasText(usage) {
		t.Fatalf("no %q in %q", usage, tt.Texts())
	}
	if len(s.shots) != 2 || s.shots[0].Path != newest.Path {
		t.Fatalf("shots %+v", s.shots)
	}
	// The newest comes first: delete it from its own button.
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, err := os.Stat(newest.Path); !os.IsNotExist(err) {
		t.Fatalf("the file is still there: %v", err)
	}
	if len(s.shots) != 1 || !tt.HasText("Deleted") {
		t.Fatalf("after delete: %+v %q", s.shots, tt.Texts())
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	// Put back with the time it was taken, it keeps its place.
	if len(s.shots) != 2 || s.shots[0].Path != newest.Path || s.shots[0].Time.Sub(newest.Time).Abs() > time.Millisecond {
		t.Fatalf("after undo: %+v", s.shots)
	}
}

func TestClearAll(t *testing.T) {
	s := testScreenshot(t)
	addShot(t, s, time.Minute)
	s.reload()
	tt := ui.NewTester(pageView(s), pageWidth, 900)
	tt.SetLocale("zh-Hans")
	if err := tt.Click("清除全部"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("删除全部截图？") {
		t.Fatalf("no confirmation: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if shots, _ := s.lib.List(); len(shots) != 0 || len(s.shots) != 0 {
		t.Fatalf("left %+v", shots)
	}
}

// Choosing a shorter time keeps it, and deletes what is past it at once.
func TestKeep(t *testing.T) {
	s := testScreenshot(t)
	addShot(t, s, 3*24*time.Hour)
	addShot(t, s, time.Hour)
	s.reload()
	s.keepDays = 1
	s.setKeep()
	if got := s.settings.Get().KeepDays; got != 1 {
		t.Fatalf("kept %d days", got)
	}
	// setKeep prunes on a goroutine; prune here to wait for it.
	if n, err := s.lib.Prune(s.settings.Get().keep(), time.Now()); err != nil || n > 1 {
		t.Fatalf("pruned %d, %v", n, err)
	}
	if shots, _ := s.lib.List(); len(shots) != 1 {
		t.Fatalf("left %d screenshots", len(shots))
	}
	if (Settings{}).keep() != 0 || (Settings{KeepDays: 7}).keep() != 7*24*time.Hour {
		t.Fatal("keep")
	}
	m := locales.In("en")
	if keepLabel(m, 1) != "1 day" || keepLabel(m, 30) != "30 days" || keepLabel(m, 0) != "Forever" {
		t.Fatalf("labels %q %q %q", keepLabel(m, 1), keepLabel(m, 30), keepLabel(m, 0))
	}
}

func TestPopupFit(t *testing.T) {
	for _, c := range []struct {
		w, h   int
		fw, fh float32
	}{
		{400, 280, 200, 140},
		{400, 100, 200, 50},
		{100, 280, 50, 140},
		{60, 40, 30, 20}, // small screenshots show at their size
	} {
		if fw, fh := fit(c.w, c.h); fw != c.fw || fh != c.fh {
			t.Errorf("fit(%d, %d) = %v×%v, want %v×%v", c.w, c.h, fw, fh, c.fw, c.fh)
		}
	}
}

func TestPopupView(t *testing.T) {
	s := testScreenshot(t)
	shot := addShot(t, s, 0)
	data, _ := os.ReadFile(shot.Path)
	thumb, err := thumbnail(data)
	if err != nil {
		t.Fatal(err)
	}
	s.showPopup(shot, thumb) // no window in tests: the view alone
	tt := ui.NewTester(s.popupView, popupThumbWidth+2*popupMargin, popupThumbHeight+2*popupMargin)
	if corner := tt.Image().RGBAAt(1, 1); corner.A != 0 {
		t.Fatalf("window corner %v is not transparent", corner)
	}
	s.hidePopup()
	if s.pop.thumb != nil {
		t.Fatal("still shows a thumbnail")
	}
}
