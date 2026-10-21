//go:build darwin || windows

package voice

import (
	"runtime"

	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	shortcuts "github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/controller"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/soniox"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

var (
	iconVoice      = widget.Icon(`<path d="M2.5 7v2M5 4.5v7M8 2.5v11M11 4.5v7M13.5 7v2"/>`)
	iconDictionary = widget.Icon(`<path d="M3 2.5h7.5a1.5 1.5 0 0 1 1.5 1.5v10H4.5A1.5 1.5 0 0 1 3 12.5z"/><path d="M3 12.5A1.5 1.5 0 0 1 4.5 11H12"/><path d="M6 5.5h3.5"/>`)
	iconModels     = widget.Icon(`<rect x="3.5" y="3.5" width="9" height="9" rx="2"/><path d="M6 1.5v2M10 1.5v2M6 12.5v2M10 12.5v2M1.5 6h2M1.5 10h2M12.5 6h2M12.5 10h2"/>`)
)

// Pages are voice input's settings pages: its two modes and the history
// and dictionary used every day, then the models set up once.
func (v *Voice) Pages() []app.Page {
	return []app.Page{
		{ID: "voice", Icon: iconVoice, Title: locales.Messages.NavVoice, View: v.page(v.voicePage)},
		{ID: "history", Icon: iconHistory, Title: locales.Messages.NavHistory, View: v.page(v.historyPage)},
		{ID: "dictionary", Icon: iconDictionary, Title: locales.Messages.NavDictionary, View: v.page(v.dictionaryPage)},
		{ID: "models", Icon: iconModels, Title: locales.Messages.NavModels, Setup: true, View: v.page(v.modelsPage)},
	}
}

// page shows a page, with the last save that failed under it.
func (v *Voice) page(view func(c *ui.Context, m locales.Messages, pal widget.Palette)) func(c *ui.Context, m locales.Messages, pal widget.Palette) {
	return func(c *ui.Context, m locales.Messages, pal widget.Palette) {
		view(c, m, pal)
		if note := v.saved.text(m); note != "" {
			ui.Text(c, note).FontSize(12).TextColor(c.Theme().Danger)
		}
	}
}

// Suggestions for the model and provider fields: OpenRouter's IDs that suit
// voice input, which the user may change for any other.
var (
	asrModelSuggestions = []string{
		"qwen/qwen3-asr-1.7b", "openai/gpt-4o-mini-transcribe", "openai/gpt-4o-transcribe",
		"microsoft/mai-transcribe-2", "elevenlabs/scribe-v2", "openai/whisper-large-v3-turbo",
	}
	llmModelSuggestions = []string{
		"openai/gpt-oss-120b", "anthropic/claude-haiku-5.5", "google/gemini-3.1-flash-lite", "openai/gpt-oss-20b",
	}
	providerSuggestions = []string{"cerebras", "groq", "together", "deepinfra", "fireworks", "sambanova"}
)

// languageCodes are the codes of the languages the app knows.
func languageCodes() []string {
	codes := make([]string, len(prompt.Languages))
	for i, l := range prompt.Languages {
		codes[i] = l.Code
	}
	return codes
}

func (v *Voice) modelsPage(c *ui.Context, m locales.Messages, pal widget.Palette) {
	const fieldWidth = 240
	streaming := asrServices[v.form.asrService] == "soniox"
	widget.Group(c, pal, m.SectionSpeech(), func(g *widget.Rows) {
		g.Row(c, m.FieldAsrService(), m.AsrServiceNote(), func() {
			ui.Segmented(c.Key("asr-service"), &v.form.asrService, "Soniox", "OpenRouter").
				Label(m.FieldAsrService()).OnChange(v.commit)
		})
	})
	if streaming {
		widget.Group(c, pal, "Soniox", func(g *widget.Rows) {
			g.Row(c, m.FieldSonioxKey(), m.SonioxKeyNote(runtime.GOOS), func() {
				ui.TextInput(c.Key("soniox-key"), &v.form.sonioxKey).Password().Placeholder("Soniox API key").
					Label("Soniox " + m.FieldSonioxKey()).Width(fieldWidth).
					OnChange(func() { v.commitKey(v.sonioxKey, v.form.sonioxKey) })
			})
			g.Row(c, m.FieldSonioxModel(), "", func() {
				ui.Autocomplete(c.Key("soniox-model"), &v.form.sonioxModel, []string{soniox.DefaultModel}).
					Placeholder(soniox.DefaultModel).Label("Soniox " + m.FieldSonioxModel()).Width(fieldWidth).OnChange(v.commit)
			})
			g.Row(c, m.FieldSonioxLanguages(), m.SonioxLanguagesNote(), func() {
				ui.TokenField(c.Key("soniox-langs"), &v.form.sonioxLangs, languageCodes()).
					Label(m.FieldSonioxLanguages()).Width(fieldWidth).OnChange(v.commit)
			})
		})
	}
	widget.Group(c, pal, m.SectionOpenRouter(), func(g *widget.Rows) {
		g.Row(c, m.FieldApiKey(), m.OpenRouterKeyNote(runtime.GOOS), func() {
			ui.TextInput(c.Key("api-key"), &v.form.apiKey).Password().Placeholder("sk-or-…").
				Label("OpenRouter " + m.FieldApiKey()).Width(fieldWidth).
				OnChange(func() { v.commitKey(v.openrouterKey, v.form.apiKey) })
		})
		note := ""
		if streaming {
			note = m.AsrFallbackNote()
		}
		g.Row(c, m.FieldAsrModel(), note, func() {
			ui.Autocomplete(c.Key("asr-model"), &v.form.asrModel, asrModelSuggestions).
				Placeholder(asrModelSuggestions[0]).Label(m.FieldAsrModel()).Width(fieldWidth).OnChange(v.commit)
		})
		if !streaming {
			g.Row(c, m.FieldAsrLanguage(), m.AsrLanguagePlaceholder(), func() {
				ui.Autocomplete(c.Key("asr-lang"), &v.form.asrLang, languageCodes()).Placeholder("zh, en…").
					Label(m.FieldAsrLanguage()).Width(fieldWidth).OnChange(v.commit)
			})
		}
	})
	widget.Group(c, pal, m.SectionRewrite(), func(g *widget.Rows) {
		g.Row(c, m.FieldLlmModel(), "", func() {
			ui.Autocomplete(c.Key("llm-model"), &v.form.llmModel, llmModelSuggestions).
				Placeholder(llmModelSuggestions[0]).Label(m.FieldLlmModel()).Width(fieldWidth).OnChange(v.commit)
		})
		g.Row(c, m.FieldProviders(), m.ProvidersNote(), func() {
			ui.TokenField(c.Key("providers"), &v.form.providers, providerSuggestions).
				Label(m.FieldProviders()).Width(fieldWidth).OnChange(v.commit)
		})
		g.Row(c, m.FieldTimeout(), m.TimeoutNote(), func() {
			ui.NumberInput(c.Key("timeout"), &v.form.timeout, 1, 60, 0.5).Label(m.FieldTimeout()).OnChange(v.commit)
		})
	})
	widget.Note(c, pal, m.AutoSaveNote())
}

// voicePage is voice input in its two modes: dictation, and
// translation, which differs only in its key and in what the rewrite
// does, with what both share below them.
func (v *Voice) voicePage(c *ui.Context, m locales.Messages, pal widget.Palette) {
	labels := make([]string, len(prompt.Languages))
	for i, l := range prompt.Languages {
		labels[i] = l.Label
	}
	widget.Group(c, pal, m.SectionDictation(), func(g *widget.Rows) {
		g.Row(c, m.ShortcutDictate(), "", func() { v.host.Shortcut(c, shortcuts.Dictate) })
		g.Row(c, m.ShortcutCancel(), m.CancelNote(), func() { v.host.Shortcut(c, shortcuts.VoiceCancel) })
		g.Row(c, m.FieldAppFormat(), m.AppFormatNote(), func() {
			ui.Switch(c.Key("app-format"), &v.form.appFormat).Label(m.FieldAppFormat()).OnChange(v.commit)
		})
	})
	widget.Group(c, pal, m.SectionTranslate(), func(g *widget.Rows) {
		g.Row(c, m.ShortcutTranslate(), m.TranslateNote(), func() { v.host.Shortcut(c, shortcuts.Translate) })
		g.Row(c, m.FieldChineseTo(), "", func() {
			widget.Dropdown(c.Key("zh"), m.FieldChineseTo(), &v.form.zhTo, labels, widget.Same, nil, 160, v.commit)
		})
		g.Row(c, m.FieldEnglishTo(), "", func() {
			widget.Dropdown(c.Key("en"), m.FieldEnglishTo(), &v.form.enTo, labels, widget.Same, nil, 160, v.commit)
		})
		g.Row(c, m.FieldOtherTo(), "", func() {
			widget.Dropdown(c.Key("default"), m.FieldOtherTo(), &v.form.defaultTo, labels, widget.Same, nil, 160, v.commit)
		})
		g.Row(c, m.ShortcutSwitch(), m.TargetsTabNote(), func() { v.host.Shortcut(c, shortcuts.CycleTarget) })
		g.Block(c, func() {
			ui.Grid(c).Columns(3).GapX(16).GapY(12).Label(m.SectionTargets()).Children(func() {
				for i, l := range prompt.Languages {
					ui.Checkbox(c.Key("target-"+l.Code), &v.form.targets[i], l.Label).OnChange(v.commit)
				}
			})
		})
		g.Row(c, m.FieldRemember(), m.RememberNote(), func() {
			ui.Switch(c.Key("remember"), &v.form.remember).Label(m.FieldRemember()).OnChange(v.commit)
		})
	})
	widget.Group(c, pal, m.SectionVoiceBar(), func(g *widget.Rows) {
		g.Row(c, m.FieldWhileRecording(), m.WhileRecordingNote(), func() {
			ui.Segmented(c.Key("while-recording"), &v.form.whileRec,
				m.WhileRecordingMute(), m.WhileRecordingLower(), m.WhileRecordingNone()).
				Label(m.FieldWhileRecording()).OnChange(v.commit)
		})
		g.Row(c, m.FieldShowTimings(), m.ShowTimingsNote(), func() {
			ui.Switch(c.Key("show-timings"), &v.form.showTimings).Label(m.FieldShowTimings()).OnChange(v.commit)
		})
	})
	if s, note := v.status, noteText(m, v.status); s.Text != "" || note != "" {
		widget.Group(c, pal, m.SectionLast(), func(g *widget.Rows) {
			g.Block(c, func() {
				ui.Column(c).Gap(6).Children(func() {
					if note != "" {
						ui.Text(c, note).FontSize(12).TextColor(pal.Muted)
					}
					if s.Timings != (controller.Timings{}) {
						ui.Text(c, lastTimingsText(m, s)).FontSize(12).TextColor(pal.Muted)
					}
					if s.Text != "" {
						ui.Text(c, s.Text).Selectable()
					}
				})
			})
		})
	}
}

func (v *Voice) dictionaryPage(c *ui.Context, m locales.Messages, pal widget.Palette) {
	widget.Group(c, pal, "", func(g *widget.Rows) {
		g.Block(c, func() {
			ui.TokenField(c.Key("dictionary"), &v.form.dictionary, nil).Label(m.NavDictionary()).
				Placeholder(m.DictionaryAdd()).MinHeight(120).OnChange(v.commit)
		})
	})
	widget.Note(c, pal, m.DictionaryNote())
}

// lastTimingsText says how long the last session's requests took.
func lastTimingsText(m locales.Messages, s controller.Status) string {
	asr, llm := formatDuration(s.Timings.Transcribe), formatDuration(s.Timings.Rewrite)
	if s.Mode == prompt.Translate {
		return m.LastTimingsTranslate(asr, llm)
	}
	return m.LastTimingsDictate(asr, llm)
}
