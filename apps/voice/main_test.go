//go:build darwin

package main

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
)

func testApp(t *testing.T) *app {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return &app{
		store: st,
		rec:   &darwin.Recorder{},
		ctrl:  controller.New(ctx, controller.Deps{Settings: st.Get}),
	}
}

func TestOverlayShowsMode(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("zh-Hans")

	a.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt.Frame()
	if !tt.HasText("听写") || !tt.HasText("Fn 完成 · Esc 取消") {
		t.Errorf("dictation bar %q", tt.Texts())
	}

	a.status = controller.Status{State: controller.Recording, Mode: prompt.Translate, Target: "en"}
	tt.Frame()
	if !tt.HasText("翻译 → English") {
		t.Errorf("translation bar %q", tt.Texts())
	}

	a.status = controller.Status{State: controller.Rewriting, Mode: prompt.Translate, Target: "zh"}
	tt.Frame()
	if !tt.HasText("翻译成 简体中文…") {
		t.Errorf("rewriting bar %q", tt.Texts())
	}

	a.status = controller.Status{State: controller.Idle, Note: controller.NoteCancelled}
	tt.Frame()
	if !tt.HasText("已取消") {
		t.Errorf("note bar %q", tt.Texts())
	}
}

func TestOverlayFollowsLanguage(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("en")
	a.status = controller.Status{State: controller.Recording, Mode: prompt.Translate, Target: "zh"}
	tt.Frame()
	if !tt.HasText("Translate → 简体中文") || !tt.HasText("Tab to switch · Esc to cancel") {
		t.Errorf("English bar %q", tt.Texts())
	}
	a.status = controller.Status{State: controller.Idle, Note: controller.NotePasteFailed, Err: darwin.ErrNotTrusted}
	tt.Frame()
	if !tt.HasText("Couldn't paste without Accessibility; the text is on the clipboard") {
		t.Errorf("English note %q", tt.Texts())
	}
}

func TestSettingsFollowLanguage(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), "")
	tt := ui.NewTester(a.settingsView, 560, 720)
	tt.SetLocale("en")
	for _, want := range []string{"General", "Language", "System default", "Permissions", "Save"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	tt.SetLocale("zh-Hans")
	if !tt.HasText("跟随系统") || !tt.HasText("界面语言") || tt.HasText("System default") {
		t.Errorf("Chinese settings %q", tt.Texts())
	}
}

func TestSettingsViewRenders(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), "")
	tt := ui.NewTester(a.settingsView, 560, 720)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"权限", "OpenRouter", "翻译", "词典", "保存"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}

func TestFormRoundTrip(t *testing.T) {
	s := store.Defaults()
	f := formFrom(s, "sk-or-1")
	if f.zhTo != "English" || f.enTo != "简体中文" || f.providers != "cerebras, groq" {
		t.Fatalf("form %+v", f)
	}
	f.llmModel = " anthropic/claude-haiku-5.5 "
	f.enTo = "日本語"
	f.providers = "groq,cerebras  together"
	f.timeout = "oops"
	f.dictionary = "MyGo\n\n OpenRouter \n"
	f.targets = make([]bool, len(prompt.Languages)) // none ticked
	f.apply(&s)

	if s.LLMModel != "anthropic/claude-haiku-5.5" {
		t.Errorf("model %q", s.LLMModel)
	}
	if s.Rules["en"] != "ja" || s.Rules["zh"] != "en" {
		t.Errorf("rules %v", s.Rules)
	}
	if !slices.Equal(s.Providers, []string{"groq", "cerebras", "together"}) {
		t.Errorf("providers %q", s.Providers)
	}
	if s.LLMTimeout != store.Defaults().LLMTimeout {
		t.Errorf("timeout %v", s.LLMTimeout)
	}
	if !slices.Equal(s.Dictionary, []string{"MyGo", "OpenRouter"}) {
		t.Errorf("dictionary %q", s.Dictionary)
	}
	if !slices.Equal(s.Targets, []string{"en"}) {
		t.Errorf("no targets ticked should keep the default: %q", s.Targets)
	}
}

func TestSettingsShowAppearance(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), "")
	tt := ui.NewTester(a.settingsView, 560, 720)
	tt.SetLocale("en")
	for _, want := range []string{"Appearance", "System", "Light", "Dark"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"外观", "浅色", "深色"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}

func TestOverlayFollowsAppearance(t *testing.T) {
	a := testApp(t)
	a.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	// A point on the bar's background, between the red dot and the label.
	luminance := func() int {
		px := tt.Image().RGBAAt(40, overlayHeight/2-12)
		return (int(px.R) + int(px.G) + int(px.B)) / 3
	}
	tt.SetDark(false)
	light := luminance()
	tt.SetDark(true)
	dark := luminance()
	if light < 200 || dark > 60 {
		t.Fatalf("background luminance light %d, dark %d", light, dark)
	}
}
