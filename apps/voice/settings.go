//go:build darwin

package main

import (
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

// settingsForm holds the settings window's fields until Save.
type settingsForm struct {
	apiKey     string
	asrModel   string
	asrLang    string
	llmModel   string
	providers  string
	timeout    string
	targets    []bool // by prompt.Languages
	defaultTo  string // language labels, as the selects show them
	zhTo       string
	enTo       string
	remember   bool
	dictionary string
}

type permissions struct {
	accessibility bool
	microphone    darwin.MicStatus
	tapRunning    bool
}

func formFrom(s store.Settings, key string) settingsForm {
	f := settingsForm{
		apiKey:     key,
		asrModel:   s.ASRModel,
		asrLang:    s.ASRLanguage,
		llmModel:   s.LLMModel,
		providers:  strings.Join(s.Providers, ", "),
		timeout:    strconv.FormatFloat(s.LLMTimeout, 'f', -1, 64),
		targets:    make([]bool, len(prompt.Languages)),
		defaultTo:  prompt.LanguageFor(s.DefaultTarget).Label,
		zhTo:       prompt.LanguageFor(s.TargetFor("zh")).Label,
		enTo:       prompt.LanguageFor(s.TargetFor("en")).Label,
		remember:   s.RememberTarget,
		dictionary: strings.Join(s.Dictionary, "\n"),
	}
	for i, l := range prompt.Languages {
		for _, t := range s.Targets {
			if t == l.Code {
				f.targets[i] = true
			}
		}
	}
	return f
}

func codeFor(label string) string {
	for _, l := range prompt.Languages {
		if l.Label == label {
			return l.Code
		}
	}
	return "en"
}

func splitList(s string, sep func(rune) bool) []string {
	var out []string
	for _, part := range strings.FieldsFunc(s, sep) {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// apply copies the form into the settings.
func (f settingsForm) apply(s *store.Settings) {
	timeout, err := strconv.ParseFloat(strings.TrimSpace(f.timeout), 64)
	if err != nil || timeout <= 0 {
		timeout = store.Defaults().LLMTimeout
	}
	var targets []string
	for i, on := range f.targets {
		if on {
			targets = append(targets, prompt.Languages[i].Code)
		}
	}
	defaultTo := codeFor(f.defaultTo)
	if len(targets) == 0 {
		targets = []string{defaultTo}
	}
	s.ASRModel = strings.TrimSpace(f.asrModel)
	s.ASRLanguage = strings.TrimSpace(f.asrLang)
	s.LLMModel = strings.TrimSpace(f.llmModel)
	s.Providers = splitList(f.providers, func(r rune) bool { return r == ',' || r == ' ' })
	s.LLMTimeout = timeout
	s.Targets = targets
	s.DefaultTarget = defaultTo
	if s.Rules == nil {
		s.Rules = map[string]string{}
	}
	s.Rules["zh"] = codeFor(f.zhTo)
	s.Rules["en"] = codeFor(f.enTo)
	s.RememberTarget = f.remember
	s.Dictionary = splitList(f.dictionary, func(r rune) bool { return r == '\n' })
}

// saveResult is the outcome of the last Save, shown beside the button.
type saveResult struct {
	state saveState
	err   error
}

type saveState int

const (
	saveNone saveState = iota
	saveDone
	saveKeyFailed
	saveFailed
)

func (r saveResult) text(m locales.Messages) string {
	switch r.state {
	case saveDone:
		return m.Saved()
	case saveKeyFailed:
		return m.SaveKeyFailed(r.err.Error())
	case saveFailed:
		return m.SaveFailed(r.err.Error())
	}
	return ""
}

// save writes the form to the store and the key to the Keychain.
func (a *app) save() {
	if err := a.setKey(strings.TrimSpace(a.form.apiKey)); err != nil {
		a.saved = saveResult{saveKeyFailed, err}
		return
	}
	if err := a.store.Update(a.form.apply); err != nil {
		a.saved = saveResult{saveFailed, err}
		return
	}
	a.form = formFrom(a.store.Get(), a.apiKey())
	a.saved = saveResult{state: saveDone}
}

// appearances are the appearance choices, in the order the segmented
// control shows them, as the settings keep them.
var appearances = []string{"", "light", "dark"}

// applyAppearance makes the app light, dark or the system's.
func applyAppearance(appearance string) {
	source := mygo.ThemeSystem
	switch appearance {
	case "light":
		source = mygo.ThemeLight
	case "dark":
		source = mygo.ThemeDark
	}
	mygo.Theme.SetSource(source)
}

// setAppearance switches the appearance at once and keeps the choice.
func (a *app) setAppearance(i int) {
	appearance := appearances[i]
	_ = a.store.Update(func(s *store.Settings) { s.Appearance = appearance })
	applyAppearance(appearance)
}

// uiLanguages are the interface languages, "" for the system's.
var uiLanguages = []string{"", "en", "zh-Hans"}

func uiLanguageLabels(m locales.Messages) []string {
	return []string{m.LanguageSystem(), "English", "简体中文"}
}

// setUILanguage switches the interface at once and keeps the choice.
func (a *app) setUILanguage(tag string) {
	if tag == a.uiLanguage {
		return
	}
	a.uiLanguage = tag
	_ = a.store.Update(func(s *store.Settings) { s.UILanguage = tag })
	mygo.App.SetLocale(tag)
}

func (a *app) languageSelect(c *ui.Context, m locales.Messages) {
	labels := uiLanguageLabels(m)
	// The label of the system option changes with the language; pick the
	// current choice again when the bound label no longer matches one.
	if !slices.Contains(labels, a.uiLanguageLabel) {
		i := max(0, slices.Index(uiLanguages, a.uiLanguage))
		a.uiLanguageLabel = labels[i]
	}
	ui.Select(c.Key("ui-language"), &a.uiLanguageLabel, labels).OnChange(func() {
		if i := slices.Index(labels, a.uiLanguageLabel); i >= 0 {
			a.setUILanguage(uiLanguages[i])
		}
	})
}

func (a *app) settingsView(c *ui.Context) {
	t := c.Theme()
	m := locales.In(c.Locale())
	labels := make([]string, len(prompt.Languages))
	for i, l := range prompt.Languages {
		labels[i] = l.Label
	}
	ui.Scroll(c).Fill().Children(func() {
		ui.Column(c).Padding(24).Gap(20).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				ui.Text(c, "Voice").FontSize(24).Bold()
				ui.Text(c, m.Tagline()).FontSize(12).TextColor(t.TextMuted)
			})

			section(c, m.SectionGeneral(), func() {
				ui.Form(c, func() {
					ui.Field(c, m.FieldLanguage(), func() { a.languageSelect(c, m) })
					ui.Field(c, m.FieldAppearance(), func() {
						ui.Segmented(c.Key("appearance"), &a.appearance,
							m.AppearanceSystem(), m.AppearanceLight(), m.AppearanceDark()).
							OnChange(func() { a.setAppearance(a.appearance) })
					})
				})
			})

			section(c, m.SectionPermissions(), func() {
				permRow(c, m.PermAccessibility(), check(a.perms.accessibility), m.PermAccessibilityNote(), m.OpenSettings(), darwin.OpenAccessibilitySettings)
				permRow(c, m.PermMicrophone(), micMark(a.perms.microphone), micText(m, a.perms.microphone), m.OpenSettings(), darwin.OpenMicrophoneSettings)
				permRow(c, m.PermHotkeys(), check(a.perms.tapRunning), m.PermHotkeysNote(), "", nil)
				permRow(c, m.PermGlobe(), markInfo, m.PermGlobeNote(), m.OpenKeyboardSettings(), darwin.OpenKeyboardSettings)
			})

			section(c, m.SectionOpenRouter(), func() {
				ui.Form(c, func() {
					ui.Field(c, m.FieldApiKey(), func() {
						ui.TextInput(c, &a.form.apiKey).Password().Placeholder("sk-or-…")
					}).Description(m.ApiKeyNote())
					ui.Field(c, m.FieldAsrModel(), func() {
						ui.TextInput(c, &a.form.asrModel).Placeholder("qwen/qwen3-asr-1.7b")
					}).Description(m.AsrModelNote())
					ui.Field(c, m.FieldAsrLanguage(), func() {
						ui.TextInput(c, &a.form.asrLang).Placeholder(m.AsrLanguagePlaceholder())
					})
					ui.Field(c, m.FieldLlmModel(), func() {
						ui.TextInput(c, &a.form.llmModel).Placeholder("openai/gpt-oss-120b")
					}).Description(m.LlmModelNote())
					ui.Field(c, m.FieldProviders(), func() {
						ui.TextInput(c, &a.form.providers).Placeholder("cerebras, groq")
					})
					ui.Field(c, m.FieldTimeout(), func() {
						ui.TextInput(c, &a.form.timeout).Width(80)
					}).Description(m.TimeoutNote())
				})
			})

			section(c, m.SectionTranslation(), func() {
				ui.Form(c, func() {
					ui.Field(c, m.FieldChineseTo(), func() { ui.Select(c.Key("zh"), &a.form.zhTo, labels) })
					ui.Field(c, m.FieldEnglishTo(), func() { ui.Select(c.Key("en"), &a.form.enTo, labels) })
					ui.Field(c, m.FieldOtherTo(), func() { ui.Select(c.Key("default"), &a.form.defaultTo, labels) })
					ui.Field(c, m.FieldRemember(), func() { ui.Switch(c, &a.form.remember) }).
						Description(m.RememberNote())
				})
				ui.Text(c, m.TargetsNote()).FontSize(12).TextColor(t.TextMuted)
				ui.Row(c).Wrap().Gap(12).Children(func() {
					for i, l := range prompt.Languages {
						ui.Checkbox(c.Key("target-"+l.Code), &a.form.targets[i], l.Label)
					}
				})
			})

			section(c, m.SectionDictionary(), func() {
				ui.TextArea(c, &a.form.dictionary).Height(96).Placeholder(m.DictionaryPlaceholder())
			})

			if s, note := a.status, noteText(m, a.status); s.Text != "" || note != "" {
				section(c, m.SectionLast(), func() {
					if note != "" {
						ui.Text(c, note).FontSize(12).TextColor(t.TextMuted)
					}
					if s.Text != "" {
						ui.Text(c, s.Text).Selectable()
					}
				})
			}

			ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.saved.text(m)).FontSize(12).TextColor(t.TextMuted).Grow(1)
				if ui.Button(c, m.ButtonRevert()).Clicked() {
					a.form = formFrom(a.store.Get(), a.apiKey())
					a.saved = saveResult{}
				}
				if ui.PrimaryButton(c, m.ButtonSave()).Clicked() {
					a.save()
				}
			})
		})
	})
}

func section(c *ui.Context, title string, body func()) {
	ui.Column(c).Gap(10).Children(func() {
		ui.Text(c, title).FontSize(15).Bold()
		body()
	})
}

type mark int

const (
	markOK mark = iota
	markMissing
	markInfo
)

func check(ok bool) mark {
	if ok {
		return markOK
	}
	return markMissing
}

func permRow(c *ui.Context, name string, m mark, note, action string, open func()) {
	t := c.Theme()
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		symbol, color := "✓", t.Success
		switch m {
		case markMissing:
			symbol, color = "✗", t.Danger
		case markInfo:
			symbol, color = "i", t.TextMuted
		}
		ui.Text(c, symbol).Bold().TextColor(color).Width(14)
		ui.Text(c, name).Width(96)
		ui.Text(c, note).FontSize(12).TextColor(t.TextMuted).Grow(1)
		if open != nil && ui.Button(c, action).Clicked() {
			open()
		}
	})
}

func micMark(s darwin.MicStatus) mark {
	switch s {
	case darwin.MicAuthorized:
		return markOK
	case darwin.MicDenied, darwin.MicRestricted:
		return markMissing
	}
	return markInfo
}

func micText(m locales.Messages, s darwin.MicStatus) string {
	switch s {
	case darwin.MicAuthorized:
		return m.MicAllowed()
	case darwin.MicDenied, darwin.MicRestricted:
		return m.MicDenied()
	}
	return m.MicAsk()
}
