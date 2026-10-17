//go:build darwin

package app

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

// tool is a feature with a page used every day, one set up once, and a
// part of the Permissions page.
type tool struct{}

func (tool) Start(*Host) error { return nil }

func (tool) Pages() []Page {
	text := func(s string) func(*ui.Context, locales.Messages, widget.Palette) {
		return func(c *ui.Context, _ locales.Messages, _ widget.Palette) { ui.Text(c, s) }
	}
	return []Page{
		{ID: "setup", Title: func(locales.Messages) string { return "Tool Setup" }, Setup: true, View: text("set up once")},
		{ID: "daily", Title: func(locales.Messages) string { return "Tool" }, View: text("used every day")},
	}
}

func (tool) PermissionRows(c *ui.Context, _ locales.Messages, _ widget.Palette, g *widget.Rows) {
	g.Row(c, "Tool permission", "", nil)
}

func (tool) PermissionTips(c *ui.Context, _ locales.Messages, pal widget.Palette) {
	widget.Group(c, pal, "Tool tip", func(g *widget.Rows) { g.Row(c, "Do this", "", nil) })
}

func testApp(t *testing.T, features ...Feature) *App {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	a := &App{store: st, features: features}
	if a.settings, err = store.NewSection(st, settingsSection, Settings{}, nil); err != nil {
		t.Fatal(err)
	}
	a.pages = a.collectPages()
	a.tap = &darwin.EventTap{Handler: a.keys}
	return a
}

// General comes first, then the features' pages used every day, then
// those set up once, then Permissions.
func TestPages(t *testing.T) {
	a := testApp(t, tool{})
	var ids []string
	for _, p := range a.pages {
		ids = append(ids, p.ID)
	}
	if want := []string{"general", "daily", "setup", "permissions"}; !slices.Equal(ids, want) {
		t.Fatalf("pages %q, want %q", ids, want)
	}
	tt := ui.NewTester(a.view, settingsWidth, settingsHeight)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"通用", "Tool", "Tool Setup", "权限", "界面语言", "外观", "强调色"} {
		if !tt.HasText(want) {
			t.Errorf("general page: no %q in %q", want, tt.Texts())
		}
	}
	for _, step := range []struct{ item, want string }{
		{"Tool", "used every day"},
		{"Tool Setup", "set up once"},
		{"权限", "辅助功能"},
		{"权限", "Tool permission"},
		{"权限", "按键监听"},
		{"权限", "Tool tip"},
	} {
		if err := tt.Click(step.item); err != nil {
			t.Fatal(err)
		}
		if !tt.HasText(step.want) {
			t.Errorf("%s page: no %q in %q", step.item, step.want, tt.Texts())
		}
	}
}

func TestSettingsFollowLanguage(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, settingsWidth, settingsHeight)
	tt.SetLocale("en")
	for _, want := range []string{"General", "Language", "System default", "Appearance", "Light", "Dark", "Accent color"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	tt.SetLocale("zh-Hans")
	if !tt.HasText("跟随系统") || !tt.HasText("界面语言") || tt.HasText("System default") {
		t.Errorf("Chinese settings %q", tt.Texts())
	}
}

func TestWithAccent(t *testing.T) {
	base := ui.LightTheme()
	if got := withAccent(base, ""); got != base {
		t.Fatal("the system choice changed the theme")
	}
	blue := withAccent(base, "blue")
	if blue.Accent != ui.Hex("#007aff") || blue.AccentText != ui.Hex("#ffffff") {
		t.Fatalf("blue: accent %v text %v", blue.Accent, blue.AccentText)
	}
	if base.Accent == blue.Accent {
		t.Fatal("withAccent changed the theme it was given")
	}
	if y := withAccent(base, "yellow"); y.AccentText == ui.Hex("#ffffff") {
		t.Fatal("white text on yellow")
	}
	if d := withAccent(ui.DarkTheme(), "blue"); d.Accent != ui.Hex("#0a84ff") {
		t.Fatalf("dark blue %v", d.Accent)
	}
}

func TestAccentSelect(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.view, settingsWidth, settingsHeight)
	tt.SetLocale("en")
	row, ok := tt.Find("Accent color")
	if !ok {
		t.Fatalf("no accent row in %q", tt.Texts())
	}
	// The select sits at the right of its row.
	tt.ClickAt(settingsWidth-28-14-90, row.Y+row.H/2)
	tt.Frame()
	for _, want := range []string{"Blue", "Green", "Yellow", "Brown"} {
		if !tt.HasText(want) {
			t.Fatalf("the menu lacks %q: %q", want, tt.Texts())
		}
	}
	if err := tt.Click("Purple"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.accent != "purple" || a.settings.Get().AccentColor != "purple" {
		t.Fatalf("chose %q, saved %q", a.accent, a.settings.Get().AccentColor)
	}
}

// Voice's flat settings split into the app's section and voice input's,
// and its history moves under voice/.
func TestMigrateFiles(t *testing.T) {
	root := t.TempDir()
	old, dir := filepath.Join(root, "Voice Dev"), filepath.Join(root, "Kit Dev")
	if got := voiceDataDir(dir); got != old {
		t.Fatalf("Voice's data in %q, want %q", got, old)
	}
	st, err := store.Open(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := migrateFiles(old, dir, st); ok || err != nil {
		t.Fatalf("no Voice: migrated %v, %v", ok, err)
	}

	_ = os.MkdirAll(old, 0o755)
	_ = os.WriteFile(filepath.Join(old, "settings.json"), []byte(`{"uiLanguage":"zh-Hans","accentColor":"green","llmModel":"m"}`), 0o644)
	_ = os.WriteFile(filepath.Join(old, "history.jsonl"), []byte("{}\n"), 0o644)
	if ok, err := migrateFiles(old, dir, st); !ok || err != nil {
		t.Fatalf("migrated %v, %v", ok, err)
	}
	s, err := store.NewSection(st, settingsSection, Settings{}, nil)
	if err != nil || s.Get() != (Settings{UILanguage: "zh-Hans", AccentColor: "green"}) {
		t.Fatalf("app settings %+v, %v", s.Get(), err)
	}
	if raw := st.Raw(voiceSection); string(raw) != `{"uiLanguage":"zh-Hans","accentColor":"green","llmModel":"m"}` {
		t.Fatalf("voice settings %s", raw)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "voice", "history.jsonl")); err != nil || string(b) != "{}\n" {
		t.Fatalf("history %q, %v", b, err)
	}
	// Reopened, the file holds both sections.
	if st, err := store.Open(filepath.Join(dir, "settings.json")); err != nil || st.Raw(voiceSection) == nil || st.Raw(settingsSection) == nil {
		t.Fatalf("saved file: %v", err)
	}
}

// recorder is a key handler that swallows one key.
type recorder struct {
	swallow int64
	seen    []int64
}

func (r *recorder) FlagsChanged(keycode int64, _ uint64) bool {
	r.seen = append(r.seen, keycode)
	return keycode == r.swallow
}

func (r *recorder) KeyDown(keycode int64, _ uint64, _ bool) bool {
	r.seen = append(r.seen, keycode)
	return keycode == r.swallow
}

func (r *recorder) KeyUp(keycode int64) bool {
	r.seen = append(r.seen, keycode)
	return keycode == r.swallow
}

// Every feature sees every key; one that swallows it keeps it from the app
// in front.
func TestKeys(t *testing.T) {
	a, b := &recorder{swallow: 19}, &recorder{swallow: 53}
	k := keys{a, b}
	if k.KeyDown(1, 0, false) || !k.KeyDown(19, 0, false) || !k.KeyUp(53) || !k.FlagsChanged(19, 0) {
		t.Fatal("swallowed the wrong keys")
	}
	if want := []int64{1, 19, 53, 19}; !slices.Equal(a.seen, want) || !slices.Equal(b.seen, want) {
		t.Fatalf("seen %v and %v, want %v", a.seen, b.seen, want)
	}
}
