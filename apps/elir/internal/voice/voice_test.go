//go:build darwin

package voice

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

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/openrouter"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/controller"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/history"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/hotkey"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/soniox"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// pageWidth is the width of a page in the settings window, right of its
// sidebar.
const pageWidth = 560

func testVoice(t *testing.T) *Voice {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	v := New()
	v.host = &app.Host{Store: st}
	v.settings, err = store.NewSection(st, config.Section, config.Defaults(), config.Clone)
	if err != nil {
		t.Fatal(err)
	}
	v.ctrl = controller.New(ctx, controller.Deps{Settings: v.settings.Get})
	v.history = history.Open(filepath.Join(t.TempDir(), "history.jsonl"))
	v.form = formFrom(v.settings.Get(), apiKeys{})
	return v
}

// pageView shows one of the feature's pages as the settings window does,
// on its pane.
func pageView(t *testing.T, v *Voice, id string) func(c *ui.Context) {
	t.Helper()
	i := slices.IndexFunc(v.Pages(), func(p app.Page) bool { return p.ID == id })
	if i < 0 {
		t.Fatalf("no page %q", id)
	}
	page := v.Pages()[i]
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

func TestPages(t *testing.T) {
	v := testVoice(t)
	m := locales.In("zh-Hans")
	var titles []string
	for _, p := range v.Pages() {
		titles = append(titles, p.Title(m))
	}
	if want := []string{"听写与翻译", "历史", "词典", "模型"}; !slices.Equal(titles, want) {
		t.Fatalf("pages %q, want %q", titles, want)
	}
	for _, step := range []struct{ page, want string }{
		{"models", "API Key"},
		{"voice", "说中文时译成"},
		{"dictionary", "词典里的名字和术语会按原样拼写。"},
	} {
		tt := ui.NewTester(pageView(t, v, step.page), pageWidth, 900)
		tt.SetLocale("zh-Hans")
		if !tt.HasText(step.want) {
			t.Errorf("%s page: no %q in %q", step.page, step.want, tt.Texts())
		}
	}
}

// The hotkeys turn Fn into an action for the controller, and keep the
// event from no app.
func TestHotkeys(t *testing.T) {
	var got []string
	k := &hotkeys{handle: func(a hotkey.Action) { got = append(got, a.String()) }}
	if k.FlagsChanged(hotkey.KeyFn, hotkey.FlagFn) || k.FlagsChanged(hotkey.KeyFn, 0) {
		t.Fatal("swallowed Fn")
	}
	k.detector.Capture(true, false)
	if !k.KeyDown(hotkey.KeyEscape, 0, false) || !k.KeyUp(hotkey.KeyEscape) {
		t.Fatal("let a captured Esc through")
	}
	if want := []string{"dictate", "cancel"}; !slices.Equal(got, want) {
		t.Fatalf("actions %q, want %q", got, want)
	}
}

func TestOverlayShowsState(t *testing.T) {
	v := testVoice(t)
	tt := ui.NewTester(v.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("zh-Hans")

	v.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt.Frame()
	if !tt.HasText("esc") || tt.HasText("→ English") {
		t.Errorf("dictation bar %q", tt.Texts())
	}

	v.status = controller.Status{State: controller.Recording, Mode: prompt.Translate, Target: "en"}
	tt.Frame()
	if !tt.HasText("→ English") {
		t.Errorf("translation bar %q", tt.Texts())
	}

	v.status = controller.Status{State: controller.Rewriting, Mode: prompt.Translate, Target: "zh"}
	tt.Frame()
	if !tt.HasText("翻译成 简体中文…") {
		t.Errorf("rewriting bar %q", tt.Texts())
	}

	v.status = controller.Status{State: controller.Idle, Note: controller.NoteCancelled}
	tt.Frame()
	if !tt.HasText("已取消") {
		t.Errorf("note bar %q", tt.Texts())
	}
}

func TestOverlayFollowsLanguage(t *testing.T) {
	v := testVoice(t)
	tt := ui.NewTester(v.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("en")
	v.status = controller.Status{State: controller.Transcribing}
	tt.Frame()
	if !tt.HasText("Transcribing…") {
		t.Errorf("English bar %q", tt.Texts())
	}
	v.status = controller.Status{State: controller.Idle, Note: controller.NotePasteFailed, Err: darwin.ErrNotTrusted}
	tt.Frame()
	if !tt.HasText("Couldn't paste without Accessibility; the text is on the clipboard") {
		t.Errorf("English note %q", tt.Texts())
	}
}

// The capsule is dark over any appearance, and the window around it is
// transparent, with no shadow of its own to outline it.
func TestOverlayCapsule(t *testing.T) {
	v := testVoice(t)
	v.status = controller.Status{State: controller.Recording, Mode: prompt.Dictate}
	tt := ui.NewTester(v.overlayView, overlayWidth, overlayHeight)
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

// Settings save as they change: there is no Save button.
func TestSettingsSaveAsTheyChange(t *testing.T) {
	v := testVoice(t)
	// Tall enough that the timeout row shows without scrolling.
	tt := ui.NewTester(pageView(t, v, "models"), pageWidth, 1000)
	tt.SetLocale("en")
	// The timeout's field sits at the right of its row, level with the
	// middle of the label and the note under it.
	note, ok := tt.Find("After this, the raw transcript is pasted.")
	if !ok {
		t.Fatalf("no timeout row in %q", tt.Texts())
	}
	// The number input has its stepper on the right: click its text, left
	// of it.
	tt.ClickAt(pageWidth-28-14-110, note.Y+note.H/2-8)
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("8")
	tt.Frame()
	if got := v.settings.Get().LLMTimeout; got != 8 {
		t.Fatalf("timeout saved as %v, want 8", got)
	}
}

func TestFormRoundTrip(t *testing.T) {
	s := config.Defaults()
	f := formFrom(s, apiKeys{openrouter: "sk-or-1"})
	if f.zhTo != "English" || f.enTo != "简体中文" || !slices.Equal(f.providers, []string{"cerebras", "groq"}) || f.timeout != 5 {
		t.Fatalf("form %+v", f)
	}
	f.llmModel = " anthropic/claude-haiku-5.5 "
	f.enTo = "日本語"
	f.providers = []string{"groq", " cerebras ", "", "together", "groq"}
	f.timeout = 0
	f.dictionary = []string{"MyGo", "", " OpenRouter "}
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
	if s.LLMTimeout != config.Defaults().LLMTimeout {
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
	v := testVoice(t)
	tt := ui.NewTester(v.overlayView, overlayWidth, overlayHeight)
	tt.SetLocale("zh-Hans")
	v.status = controller.Status{State: controller.Rewriting, Since: time.Now().Add(-1200 * time.Millisecond)}
	tt.Frame()
	if !tt.HasText("整理中…") || !tt.HasText("1.2s") {
		t.Errorf("rewriting bar %q", tt.Texts())
	}
	v.status = controller.Status{
		State:   controller.Idle,
		Text:    "我们周四开会。",
		Timings: controller.Timings{Transcribe: 1340 * time.Millisecond, Rewrite: 620 * time.Millisecond},
	}
	tt.Frame()
	if !tt.HasText("转写 1.3s · 整理 0.6s") {
		t.Errorf("done bar %q", tt.Texts())
	}
	if !showsTimings(v.status, false) || showsTimings(v.status, true) {
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
	v := testVoice(t)
	if !v.form.showTimings {
		t.Fatal("timings are hidden by default")
	}
	tt := ui.NewTester(pageView(t, v, "voice"), pageWidth, 1000)
	tt.SetLocale("zh-Hans")
	if !tt.HasText("在语音条上显示耗时") {
		t.Fatalf("no timings switch in %q", tt.Texts())
	}
	s := config.Defaults()
	v.form.showTimings = false
	v.form.apply(&s)
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
	v := testVoice(t)
	tr := &transcriber{client: &openrouter.Client{BaseURL: srv.URL, Key: func() string { return "sk" }}, settings: v.settings.Get}
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
	v := testVoice(t)
	tr := &transcriber{client: &openrouter.Client{BaseURL: srv.URL, Key: func() string { return "sk" }}, settings: v.settings.Get}
	pcm := make([]byte, 64000) // 2 s of silence compresses to almost nothing
	if _, err := tr.Transcribe(context.Background(), pcm); err != nil {
		t.Fatal(err)
	}
	if size == 0 || size >= len(pcm)/4 {
		t.Fatalf("uploaded %d bytes for %d of PCM", size, len(pcm))
	}
}

func TestModelsPageFollowsService(t *testing.T) {
	v := testVoice(t)
	tt := ui.NewTester(pageView(t, v, "models"), pageWidth, 1000)
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
	if got := v.settings.Get().ASRService; got != "openrouter" {
		t.Fatalf("service saved as %q", got)
	}
	if tt.HasText("你会说的语言代码") || !tt.HasText("转写语言提示") {
		t.Errorf("OpenRouter: %q", tt.Texts())
	}
}

func TestFormSonioxFields(t *testing.T) {
	s := config.Defaults()
	f := formFrom(s, apiKeys{soniox: "sx"})
	if f.asrService != 0 || !slices.Equal(f.sonioxLangs, []string{"zh", "en"}) || f.sonioxModel != "stt-rt-v5" || f.sonioxKey != "sx" {
		t.Fatalf("form %+v", f)
	}
	f.asrService = 1
	f.sonioxLangs = []string{"zh", "en", "ja"}
	f.apply(&s)
	if s.ASRService != "openrouter" || !slices.Equal(s.SonioxLanguages, []string{"zh", "en", "ja"}) {
		t.Fatalf("settings %+v", s)
	}
}

func TestStreamerNeedsSonioxAndAKey(t *testing.T) {
	v := testVoice(t)
	key := ""
	s := &streamer{client: &soniox.Client{URL: "ws://127.0.0.1:1"}, settings: v.settings.Get, key: func() string { return key }}
	if s.Open(context.Background()) != nil {
		t.Fatal("streams without a key")
	}
	key = "sx"
	st := s.Open(context.Background())
	if st == nil {
		t.Fatal("no stream with Soniox and a key")
	}
	st.Close()
	_ = v.settings.Update(func(x *config.Settings) { x.ASRService = "openrouter" })
	if s.Open(context.Background()) != nil {
		t.Fatal("streams with OpenRouter picked")
	}
}

// With "none" the computer's sound is left alone and nothing is saved.
func TestQuieterLeaveAlone(t *testing.T) {
	v := testVoice(t)
	_ = v.settings.Update(func(s *config.Settings) { s.WhileRecording = "none" })
	q := &quieter{settings: v.settings.Get, path: filepath.Join(t.TempDir(), "restore.json")}
	q.Recording(true)
	if _, err := os.Stat(q.path); !os.IsNotExist(err) {
		t.Fatalf("saved a restore file: %v", err)
	}
	q.Recording(false)
}

// What a crashed run left is undone, and the file goes either way. The
// device here does not exist, so nothing real changes.
func TestQuieterRecovers(t *testing.T) {
	v := testVoice(t)
	path := filepath.Join(t.TempDir(), "restore.json")
	for _, content := range []string{
		`{"kind":"volume","device":4294967000,"volume":0.5,"set":0.1}`,
		`{`,
	} {
		_ = os.WriteFile(path, []byte(content), 0o644)
		q := &quieter{settings: v.settings.Get, path: path}
		q.recover()
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s: restore file kept", content)
		}
	}
	(&quieter{settings: v.settings.Get, path: path}).recover() // no file: nothing to do
}

func TestWhileRecordingSetting(t *testing.T) {
	s := config.Defaults()
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
	v := testVoice(t)
	tt := ui.NewTester(pageView(t, v, "voice"), pageWidth, 900)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"录音时的其他声音", "静音", "调低", "不处理"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
}

func TestHistoryPage(t *testing.T) {
	v := testVoice(t)
	for _, e := range []history.Entry{
		{Mode: "dictate", Transcript: "嗯我们周三，不对，周四开会", Text: "我们周四开会。"},
		{Mode: "translate", Target: "en", Transcript: "明天发布", Text: "We ship tomorrow."},
	} {
		if _, err := v.history.Add(e); err != nil {
			t.Fatal(err)
		}
	}
	v.loadHistory()
	tt := ui.NewTester(pageView(t, v, "history"), pageWidth, 900)
	tt.SetLocale("zh-Hans")
	for _, want := range []string{"历史", "保存历史", "共 2 条", "We ship tomorrow.", "翻译 → English", "我们周四开会。", "听写", "原始转写"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}

	// Delete the newest: the translation.
	if err := tt.Click("删除"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tt.HasText("We ship tomorrow.") || !tt.HasText("共 1 条") {
		t.Fatalf("after delete: %q", tt.Texts())
	}
	if got, _ := v.history.List(); len(got) != 1 || got[0].Text != "我们周四开会。" {
		t.Fatalf("file after delete: %+v", got)
	}

	// Clear all asks first; Enter picks the alert's default, clearing.
	if err := tt.Click("清除全部"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("清除全部历史？") {
		t.Fatalf("no confirmation: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if got, _ := v.history.List(); len(got) != 0 || v.history.Size() != 0 {
		t.Fatalf("history left after clearing: %+v", got)
	}
	if !tt.HasText("还没有记录。听写或翻译的文字会显示在这里。") {
		t.Fatalf("after clear: %q", tt.Texts())
	}
}

func TestHistoryOff(t *testing.T) {
	v := testVoice(t)
	_ = v.settings.Update(func(s *config.Settings) { s.NoHistory = true })
	v.onResult(controller.Result{Text: "secret"})
	if got, _ := v.history.List(); len(got) != 0 {
		t.Fatalf("kept %+v with history off", got)
	}
	f := formFrom(v.settings.Get(), apiKeys{})
	if f.saveHistory {
		t.Fatal("form shows history on")
	}
}

func TestHistorySearchDeleteUndo(t *testing.T) {
	v := testVoice(t)
	for _, text := range []string{"我们周四开会。", "We ship tomorrow."} {
		if _, err := v.history.Add(history.Entry{Mode: "dictate", Text: text}); err != nil {
			t.Fatal(err)
		}
	}
	v.loadHistory()
	tt := ui.NewTester(pageView(t, v, "history"), pageWidth, 900)
	tt.SetLocale("zh-Hans")

	// Search narrows the list.
	v.hist.query = "ship"
	tt.Frame()
	if !tt.HasText("We ship tomorrow.") || tt.HasText("我们周四开会。") {
		t.Fatalf("search: %q", tt.Texts())
	}
	v.hist.query = "nothing like it"
	tt.Frame()
	if !tt.HasText("没有匹配的记录") {
		t.Fatalf("no match: %q", tt.Texts())
	}
	v.hist.query = ""
	tt.Frame()

	// Delete from the right-click menu, then undo from the toast. The
	// entry's selectable text keeps its own copy menu; the rest of the card,
	// as its badge, opens the entry's. The newest entry comes first.
	if err := tt.RightClick("听写"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("删除"); err != nil {
		t.Fatalf("%v; menu %q", err, tt.Menu())
	}
	tt.Frame()
	if got, _ := v.history.List(); len(got) != 1 {
		t.Fatalf("after delete: %+v", got)
	}
	if !tt.HasText("已删除") {
		t.Fatalf("no toast: %q", tt.Texts())
	}
	if err := tt.Click("撤销"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got, _ := v.history.List(); len(got) != 2 || got[0].Text != "We ship tomorrow." {
		t.Fatalf("after undo: %+v", got)
	}
	if !tt.HasText("We ship tomorrow.") {
		t.Fatalf("the page did not get the entry back: %q", tt.Texts())
	}
}
