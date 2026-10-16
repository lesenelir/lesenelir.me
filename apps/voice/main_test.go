//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/soniox"
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
	a.form = formFrom(a.store.Get(), apiKeys{})
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
	a.form = formFrom(a.store.Get(), apiKeys{})
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
	a.form = formFrom(a.store.Get(), apiKeys{})
	a.page = pageModels
	// Tall enough that the timeout row shows without scrolling.
	tt := ui.NewTester(a.settingsView, settingsWidth, 1000)
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
	f := formFrom(s, apiKeys{openrouter: "sk-or-1"})
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

func TestOverlayTimings(t *testing.T) {
	a := testApp(t)
	tt := ui.NewTester(a.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("zh-Hans")
	a.status = controller.Status{State: controller.Rewriting, Since: time.Now().Add(-1200 * time.Millisecond)}
	tt.Frame()
	if !tt.HasText("整理中…") || !tt.HasText("1.2s") {
		t.Errorf("rewriting bar %q", tt.Texts())
	}
	a.status = controller.Status{
		State:   controller.Idle,
		Text:    "我们周四开会。",
		Timings: controller.Timings{Transcribe: 1340 * time.Millisecond, Rewrite: 620 * time.Millisecond},
	}
	tt.Frame()
	if !tt.HasText("转写 1.3s · 整理 0.6s") {
		t.Errorf("done bar %q", tt.Texts())
	}
	if !showsTimings(a.status, false) || showsTimings(a.status, true) {
		t.Error("the setting does not decide whether timings show")
	}
	if showsTimings(controller.Status{State: controller.Idle, Note: controller.NoteCancelled}, false) {
		t.Error("a cancelled session shows timings")
	}
}

func TestFormatDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		80 * time.Millisecond:   "0.1s",
		1240 * time.Millisecond: "1.2s",
		14 * time.Second:        "14s",
	} {
		if got := formatDuration(d); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestSettingsTimingsSwitch(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), apiKeys{})
	if !a.form.showTimings {
		t.Fatal("timings are hidden by default")
	}
	tt := ui.NewTester(a.settingsView, settingsWidth, settingsHeight)
	tt.SetLocale("zh-Hans")
	if !tt.HasText("在语音条上显示耗时") {
		t.Fatalf("no timings switch in %q", tt.Texts())
	}
	s := store.Defaults()
	a.form.showTimings = false
	a.form.apply(&s)
	if !s.HideTimings {
		t.Fatal("turning the switch off kept timings shown")
	}
}

// A model that refuses FLAC gets the WAV every model takes, from then on
// directly.
func TestTranscriberFallsBackToWAV(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
			calls = append(calls, "flac")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":{"code":400,"message":"unsupported audio format"}}`))
			return
		}
		var body struct {
			InputAudio struct{ Format string } `json:"input_audio"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		calls = append(calls, body.InputAudio.Format)
		_, _ = w.Write([]byte(`{"text":"你好"}`))
	}))
	t.Cleanup(srv.Close)
	a := testApp(t)
	tr := &transcriber{client: &openrouter.Client{BaseURL: srv.URL, Key: func() string { return "sk" }}, store: a.store}
	pcm := make([]byte, 32000)
	for i := range pcm {
		pcm[i] = byte(i * 7)
	}
	for range 2 {
		if text, err := tr.Transcribe(context.Background(), pcm); err != nil || text != "你好" {
			t.Fatalf("got %q, %v", text, err)
		}
	}
	if want := []string{"flac", "wav", "wav"}; !slices.Equal(calls, want) {
		t.Fatalf("requests %q, want %q", calls, want)
	}
}

func TestTranscriberUploadsFLAC(t *testing.T) {
	var size int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, h, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(f)
		size = len(data)
		if h.Filename != "audio.flac" || string(data[:4]) != "fLaC" {
			t.Errorf("uploaded %q starting %q", h.Filename, data[:4])
		}
		_, _ = w.Write([]byte(`{"text":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	a := testApp(t)
	tr := &transcriber{client: &openrouter.Client{BaseURL: srv.URL, Key: func() string { return "sk" }}, store: a.store}
	pcm := make([]byte, 64000) // 2 s of silence compresses to almost nothing
	if _, err := tr.Transcribe(context.Background(), pcm); err != nil {
		t.Fatal(err)
	}
	if size == 0 || size >= len(pcm)/4 {
		t.Fatalf("uploaded %d bytes for %d of PCM", size, len(pcm))
	}
}

func TestModelsPageFollowsService(t *testing.T) {
	a := testApp(t)
	a.form = formFrom(a.store.Get(), apiKeys{})
	a.page = pageModels
	tt := ui.NewTester(a.settingsView, settingsWidth, settingsHeight)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"转写服务", "Soniox", "语言提示", "Soniox 不可用时改用它。"} {
		if !tt.HasText(want) {
			t.Errorf("Soniox: no %q in %q", want, tt.Texts())
		}
	}
	if err := tt.Click("OpenRouter"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := a.store.Get().ASRService; got != "openrouter" {
		t.Fatalf("service saved as %q", got)
	}
	if tt.HasText("你会说的语言代码") || !tt.HasText("转写语言提示") {
		t.Errorf("OpenRouter: %q", tt.Texts())
	}
}

func TestFormSonioxFields(t *testing.T) {
	s := store.Defaults()
	f := formFrom(s, apiKeys{soniox: "sx"})
	if f.asrService != 0 || f.sonioxLangs != "zh, en" || f.sonioxModel != "stt-rt-v5" || f.sonioxKey != "sx" {
		t.Fatalf("form %+v", f)
	}
	f.asrService = 1
	f.sonioxLangs = "zh,en, ja"
	f.apply(&s)
	if s.ASRService != "openrouter" || !slices.Equal(s.SonioxLanguages, []string{"zh", "en", "ja"}) {
		t.Fatalf("settings %+v", s)
	}
}

func TestStreamerNeedsSonioxAndAKey(t *testing.T) {
	a := testApp(t)
	key := ""
	s := &streamer{client: &soniox.Client{URL: "ws://127.0.0.1:1"}, store: a.store, key: func() string { return key }}
	if s.Open(context.Background()) != nil {
		t.Fatal("streams without a key")
	}
	key = "sx"
	st := s.Open(context.Background())
	if st == nil {
		t.Fatal("no stream with Soniox and a key")
	}
	st.Close()
	_ = a.store.Update(func(x *store.Settings) { x.ASRService = "openrouter" })
	if s.Open(context.Background()) != nil {
		t.Fatal("streams with OpenRouter picked")
	}
}

// With "none" the computer's sound is left alone and nothing is saved.
func TestQuieterLeaveAlone(t *testing.T) {
	a := testApp(t)
	_ = a.store.Update(func(s *store.Settings) { s.WhileRecording = "none" })
	q := &quieter{store: a.store, path: filepath.Join(t.TempDir(), "restore.json")}
	q.Recording(true)
	if _, err := os.Stat(q.path); !os.IsNotExist(err) {
		t.Fatalf("saved a restore file: %v", err)
	}
	q.Recording(false)
}

// What a crashed run left is undone, and the file goes either way. The
// device here does not exist, so nothing real changes.
func TestQuieterRecovers(t *testing.T) {
	a := testApp(t)
	path := filepath.Join(t.TempDir(), "restore.json")
	for _, content := range []string{
		`{"kind":"volume","device":4294967000,"volume":0.5,"set":0.1}`,
		`{`,
	} {
		_ = os.WriteFile(path, []byte(content), 0o644)
		q := &quieter{store: a.store, path: path}
		q.recover()
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s: restore file kept", content)
		}
	}
	(&quieter{store: a.store, path: path}).recover() // no file: nothing to do
}

func TestWhileRecordingSetting(t *testing.T) {
	s := store.Defaults()
	if s.WhileRecording != "mute" {
		t.Fatalf("default %q", s.WhileRecording)
	}
	f := formFrom(s, apiKeys{})
	if f.whileRec != 0 {
		t.Fatalf("form index %d", f.whileRec)
	}
	f.whileRec = 1
	f.apply(&s)
	if s.WhileRecording != "lower" {
		t.Fatalf("saved %q", s.WhileRecording)
	}
	a := testApp(t)
	a.form = formFrom(a.store.Get(), apiKeys{})
	tt := ui.NewTester(a.settingsView, settingsWidth, 900)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"录音时的其他声音", "静音", "调低", "不处理"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}
