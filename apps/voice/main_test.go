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

func TestOverlayShowsState(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("zh-Hans")

	a.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt.Frame()
	if !tt.HasText("esc") || tt.HasText("→ English") {
		t.Errorf("dictation bar %q", tt.Texts())
	}

	a.status = controller.Status{State: controller.Recording, Mode: prompt.Translate, Target: "en"}
	tt.Frame()
	if !tt.HasText("→ English") {
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
	a.status = controller.Status{State: controller.Transcribing}
	tt.Frame()
	if !tt.HasText("Transcribing…") {
		t.Errorf("English bar %q", tt.Texts())
	}
	a.status = controller.Status{State: controller.Idle, Note: controller.NotePasteFailed, Err: darwin.ErrNotTrusted}
	tt.Frame()
	if !tt.HasText("Couldn't paste without Accessibility; the text is on the clipboard") {
		t.Errorf("English note %q", tt.Texts())
	}
}

// The capsule is dark over any appearance, and the window around it is
// transparent, with no shadow of its own to outline it.
func TestOverlayCapsule(t *testing.T) {
	a := testApp(t)
	a.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	// Just inside the capsule's top edge, at the middle.
	inside := func() (lum int, alpha uint8) {
		px := tt.Image().RGBAAt(overlayWidth/2, overlayHeight-18-capsuleHeight+5)
		return (int(px.R) + int(px.G) + int(px.B)) / 3, px.A
	}
	for _, dark := range []bool{false, true} {
		tt.SetDark(dark)
		if lum, alpha := inside(); lum > 60 || alpha < 200 {
			t.Errorf("dark=%v: capsule luminance %d alpha %d", dark, lum, alpha)
		}
		if corner := tt.Image().RGBAAt(1, 1); corner.A != 0 {
			t.Errorf("dark=%v: window corner %v is not transparent", dark, corner)
		}
	}
}

func TestSettingsPages(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), "")
	tt := ui.NewTester(a.settingsView, settingsWidth, settingsHeight)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"通用", "权限", "模型", "翻译", "词典", "界面语言", "外观", "快捷键", "fn", "esc"} {
		if !tt.HasText(want) {
			t.Errorf("general page: no %q in %q", want, tt.Texts())
		}
	}
	for _, step := range []struct{ item, want string }{
		{"权限", "辅助功能"},
		{"模型", "API Key"},
		{"翻译", "说中文时译成"},
		{"词典", "词典里的名字和术语会按原样拼写。"},
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
	a.form = formFrom(a.store.Get(), "")
	tt := ui.NewTester(a.settingsView, settingsWidth, settingsHeight)
	tt.SetLocale("en")
	for _, want := range []string{"General", "Language", "System default", "Appearance", "Light", "Dark", "Shortcuts"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	tt.SetLocale("zh-Hans")
	if !tt.HasText("跟随系统") || !tt.HasText("界面语言") || tt.HasText("System default") {
		t.Errorf("Chinese settings %q", tt.Texts())
	}
}

// Settings save as they change: there is no Save button.
func TestSettingsSaveAsTheyChange(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), "")
	a.page = pageModels
	tt := ui.NewTester(a.settingsView, settingsWidth, settingsHeight)
	tt.SetLocale("en")
	// The timeout's field sits at the right of its row, level with the
	// middle of the label and the note under it.
	note, ok := tt.Find("After this, the raw transcript is pasted.")
	if !ok {
		t.Fatalf("no timeout row in %q", tt.Texts())
	}
	tt.ClickAt(settingsWidth-28-14-30, note.Y+note.H/2-8)
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("8")
	tt.Frame()
	if got := a.store.Get().LLMTimeout; got != 8 {
		t.Fatalf("timeout saved as %v, want 8", got)
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
