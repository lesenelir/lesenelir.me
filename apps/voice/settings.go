//go:build darwin

package main

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/controller"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/prompt"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/soniox"
	"github.com/lesenelir/lesenelir.me/apps/voice/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

// The settings window is laid out as macOS's System Settings: a sidebar of
// pages over the window's sidebar material, under a hidden title bar whose
// traffic lights sit over the sidebar, and a page of grouped rows.
const (
	settingsWidth  = 760
	settingsHeight = 540
	sidebarWidth   = 200
	headerHeight   = 52
	// keyDelay is how long typing pauses before the API key is written to
	// the Keychain.
	keyDelay = 600 * time.Millisecond
)

// Settings pages, by sidebar item ID.
const (
	pageGeneral     = "general"
	pagePermissions = "permissions"
	pageModels      = "models"
	pageTranslation = "translation"
	pageDictionary  = "dictionary"
)

func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	iconGeneral     = icon(`<circle cx="8" cy="8" r="2.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4"/>`)
	iconPermissions = icon(`<path d="M8 1.5l5 2v4c0 3.2-2.2 5.6-5 7-2.8-1.4-5-3.8-5-7v-4z"/><path d="M5.8 8l1.6 1.6L10.4 6.6"/>`)
	iconModels      = icon(`<rect x="3.5" y="3.5" width="9" height="9" rx="2"/><path d="M6 1.5v2M10 1.5v2M6 12.5v2M10 12.5v2M1.5 6h2M1.5 10h2M12.5 6h2M12.5 10h2"/>`)
	iconTranslation = icon(`<path d="M2 3.5h7M5.5 2v1.5M3.5 3.5c.6 2.4 2.4 4.2 4.5 5M7.5 3.5c-.6 2.4-2.4 4.2-4.5 5"/><path d="M8.5 14l2.5-6 2.5 6M9.3 12h3.4"/>`)
	iconDictionary  = icon(`<path d="M3 2.5h7.5a1.5 1.5 0 0 1 1.5 1.5v10H4.5A1.5 1.5 0 0 1 3 12.5z"/><path d="M3 12.5A1.5 1.5 0 0 1 4.5 11H12"/><path d="M6 5.5h3.5"/>`)
)

// apiKeys are the keys the settings show.
type apiKeys struct {
	openrouter, soniox string
}

// settingsForm holds the settings window's fields, which are saved as they
// change.
type settingsForm struct {
	apiKey      string // OpenRouter's
	sonioxKey   string
	asrService  int // index in asrServices
	sonioxModel string
	sonioxLangs string
	asrModel    string
	asrLang     string
	llmModel    string
	providers   string
	timeout     string
	targets     []bool // by prompt.Languages
	defaultTo   string // language labels, as the selects show them
	zhTo        string
	enTo        string
	remember    bool
	dictionary  string
	showTimings bool
	whileRec    int // index in whileRecording
	saveHistory bool
}

type permissions struct {
	accessibility bool
	microphone    darwin.MicStatus
	tapRunning    bool
}

// whileRecording are the choices for the computer's sound while recording,
// as the settings keep them, in the order the control shows them.
var whileRecording = []string{"mute", "lower", "none"}

// asrServices are the speech-to-text services, as the settings keep them,
// in the order the segmented control shows them.
var asrServices = []string{"soniox", "openrouter"}

func formFrom(s store.Settings, keys apiKeys) settingsForm {
	f := settingsForm{
		apiKey:      keys.openrouter,
		sonioxKey:   keys.soniox,
		asrService:  max(0, slices.Index(asrServices, s.ASRService)),
		sonioxModel: s.SonioxModel,
		sonioxLangs: strings.Join(s.SonioxLanguages, ", "),
		asrModel:    s.ASRModel,
		asrLang:     s.ASRLanguage,
		llmModel:    s.LLMModel,
		providers:   strings.Join(s.Providers, ", "),
		timeout:     strconv.FormatFloat(s.LLMTimeout, 'f', -1, 64),
		targets:     make([]bool, len(prompt.Languages)),
		defaultTo:   prompt.LanguageFor(s.DefaultTarget).Label,
		zhTo:        prompt.LanguageFor(s.TargetFor("zh")).Label,
		enTo:        prompt.LanguageFor(s.TargetFor("en")).Label,
		remember:    s.RememberTarget,
		dictionary:  strings.Join(s.Dictionary, "\n"),
		showTimings: !s.HideTimings,
		whileRec:    max(0, slices.Index(whileRecording, s.WhileRecording)),
		saveHistory: !s.NoHistory,
	}
	for i, l := range prompt.Languages {
		f.targets[i] = slices.Contains(s.Targets, l.Code)
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
	s.ASRService = asrServices[f.asrService]
	s.SonioxModel = strings.TrimSpace(f.sonioxModel)
	s.SonioxLanguages = splitList(f.sonioxLangs, func(r rune) bool { return r == ',' || r == ' ' })
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
	s.HideTimings = !f.showTimings
	s.WhileRecording = whileRecording[f.whileRec]
	s.NoHistory = !f.saveHistory
}

// saveResult is the outcome of the last save that failed, if any.
type saveResult struct {
	state saveState
	err   error
}

type saveState int

const (
	saveNone saveState = iota
	saveKeyFailed
	saveFailed
)

func (r saveResult) text(m locales.Messages) string {
	switch r.state {
	case saveKeyFailed:
		return m.SaveKeyFailed(r.err.Error())
	case saveFailed:
		return m.SaveFailed(r.err.Error())
	}
	return ""
}

// commit saves the form as it changes.
func (a *app) commit() {
	if err := a.store.Update(a.form.apply); err != nil {
		a.saved = saveResult{saveFailed, err}
		return
	}
	if a.saved.state == saveFailed {
		a.saved = saveResult{}
	}
}

// commitKey writes a key to the Keychain once typing pauses.
func (a *app) commitKey(sec *secret, value string) {
	if sec.timer != nil {
		sec.timer.Stop()
	}
	sec.timer = time.AfterFunc(keyDelay, func() { a.storeKey(sec, value) })
}

// flushKeys writes the keys still waiting for typing to pause.
func (a *app) flushKeys() {
	for _, k := range []struct {
		sec   *secret
		value string
	}{{a.openrouterKey, a.form.apiKey}, {a.sonioxKey, a.form.sonioxKey}} {
		if k.sec.timer != nil && k.sec.timer.Stop() {
			a.storeKey(k.sec, k.value)
		}
	}
}

func (a *app) storeKey(sec *secret, value string) {
	err := sec.set(value)
	a.settings.Update(func() {
		switch {
		case err != nil:
			a.saved = saveResult{saveKeyFailed, err}
		case a.saved.state == saveKeyFailed:
			a.saved = saveResult{}
		}
	})
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

// palette is the settings window's colors over the theme's.
type palette struct {
	pane, card, border, muted ui.Color
}

func settingsPalette(t *ui.Theme) palette {
	if t.Dark {
		return palette{
			pane:   ui.Hex("#1c1c1e"),
			card:   ui.RGBA(255, 255, 255, 0.05),
			border: ui.RGBA(255, 255, 255, 0.08),
			muted:  t.TextMuted,
		}
	}
	return palette{
		pane:   ui.Hex("#f5f5f7"),
		card:   ui.Hex("#ffffff"),
		border: ui.RGBA(0, 0, 0, 0.08),
		muted:  t.TextMuted,
	}
}

func (a *app) settingsView(c *ui.Context) {
	systemAccent := c.Theme().Accent
	c.SetTheme(withAccent(c.Theme(), a.accent))
	t := c.Theme()
	m := locales.In(c.Locale())
	pal := settingsPalette(t)
	vibrant := c.Vibrancy()
	if vibrant {
		c.Root().Background(ui.Transparent)
	}
	if a.page == "" {
		a.page = pageGeneral
	}
	a.systemAccent = systemAccent
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		side := ui.Column(c).Width(sidebarWidth).Shrink(0)
		if !vibrant {
			side.Background(t.Surface)
		}
		side.Children(func() { a.settingsSidebar(c, m) })
		ui.Box(c).Width(1).Background(pal.border)
		ui.Column(c).Grow(1).MinWidth(0).Background(pal.pane).Children(func() {
			bar := c.TitleBar()
			ui.Row(c).Height(headerHeight).Shrink(0).AlignItems(ui.Center).
				Padding(0, bar.Right+24, 0, 28).DragWindow().Children(func() {
				ui.Text(c, pageTitle(m, a.page)).FontSize(15).Bold().SingleLine()
			})
			ui.Scroll(c.Key("page-" + a.page)).Grow(1).Children(func() {
				ui.Column(c).Padding(4, 28, 28, 28).Gap(22).Children(func() {
					switch a.page {
					case pageHistory:
						a.historyPage(c, m, pal)
					case pagePermissions:
						a.permissionsPage(c, m, pal)
					case pageModels:
						a.modelsPage(c, m, pal)
					case pageTranslation:
						a.translationPage(c, m, pal)
					case pageDictionary:
						a.dictionaryPage(c, m, pal)
					default:
						a.generalPage(c, m, pal)
					}
					if note := a.saved.text(m); note != "" {
						ui.Text(c, note).FontSize(12).TextColor(t.Danger)
					}
				})
			})
		})
	})
}

func (a *app) settingsSidebar(c *ui.Context, m locales.Messages) {
	// The traffic lights sit here.
	ui.Box(c).Height(headerHeight).Shrink(0).DragWindow()
	ui.Column(c).Gap(2).Padding(0, 10).Role(ui.RoleList).Label(m.SettingsTitle()).Children(func() {
		a.navItem(c, pageGeneral, iconGeneral, m.NavGeneral())
		a.navItem(c, pageHistory, iconHistory, m.NavHistory())
		a.navItem(c, pagePermissions, iconPermissions, m.NavPermissions())
		a.navItem(c, pageModels, iconModels, m.NavModels())
		a.navItem(c, pageTranslation, iconTranslation, m.NavTranslation())
		a.navItem(c, pageDictionary, iconDictionary, m.NavDictionary())
	})
}

// navItem is a page in the sidebar: the one shown filled with the accent
// color, as macOS's sidebars show it, and a faint accent behind the one
// hovered.
func (a *app) navItem(c *ui.Context, page string, icon *ui.SVG, label string) {
	t := c.Theme()
	hover := float32(0.08)
	if t.Dark {
		hover = 0.14
	}
	chosen := a.page == page
	item := ui.ButtonBase(c.Key("nav-"+page)).Gap(t.Space(2)).Padding(t.Space(1.5), t.Space(2)).
		Radius(t.Radius).AlignItems(ui.Center).Label(label).Selected(chosen)
	iconColor, textColor := t.Accent, t.Text
	switch {
	case chosen:
		item.Background(t.Accent)
		iconColor, textColor = t.AccentText, t.AccentText
	case item.Hovered():
		item.Background(ui.RGBA(t.Accent.R, t.Accent.G, t.Accent.B, hover))
	}
	item.Children(func() {
		ui.Icon(c, icon).FontSize(16).TextColor(iconColor)
		ui.Text(c, label).Grow(1).SingleLine().TextColor(textColor)
	})
	if item.Clicked() {
		a.page = page
	}
}

func pageTitle(m locales.Messages, page string) string {
	switch page {
	case pageHistory:
		return m.NavHistory()
	case pagePermissions:
		return m.NavPermissions()
	case pageModels:
		return m.NavModels()
	case pageTranslation:
		return m.NavTranslation()
	case pageDictionary:
		return m.NavDictionary()
	}
	return m.NavGeneral()
}

func (a *app) generalPage(c *ui.Context, m locales.Messages, pal palette) {
	group(c, pal, "", func(g *rows) {
		g.row(c, m.FieldLanguage(), "", func() { a.languageSelect(c, m) })
		g.row(c, m.FieldAppearance(), "", func() {
			ui.Segmented(c.Key("appearance"), &a.appearance,
				m.AppearanceSystem(), m.AppearanceLight(), m.AppearanceDark()).
				OnChange(func() { a.setAppearance(a.appearance) })
		})
		g.row(c, m.FieldAccent(), "", func() { a.accentSelect(c, m, a.systemAccent) })
		g.row(c, m.FieldWhileRecording(), m.WhileRecordingNote(), func() {
			ui.Segmented(c.Key("while-recording"), &a.form.whileRec,
				m.WhileRecordingMute(), m.WhileRecordingLower(), m.WhileRecordingNone()).OnChange(a.commit)
		})
		g.row(c, m.FieldShowTimings(), m.ShowTimingsNote(), func() {
			ui.Switch(c.Key("show-timings"), &a.form.showTimings).Label(m.FieldShowTimings()).OnChange(a.commit)
		})
	})
	group(c, pal, m.SectionShortcuts(), func(g *rows) {
		g.row(c, m.ShortcutDictate(), "", func() { keycaps(c, pal, "fn") })
		g.row(c, m.ShortcutTranslate(), "", func() { keycaps(c, pal, "fn", m.KeyLeftShift()) })
		g.row(c, m.ShortcutCancel(), "", func() { keycaps(c, pal, "esc") })
		g.row(c, m.ShortcutSwitch(), "", func() { keycaps(c, pal, "tab") })
	})
	if s, note := a.status, noteText(m, a.status); s.Text != "" || note != "" {
		group(c, pal, m.SectionLast(), func(g *rows) {
			g.block(c, func() {
				ui.Column(c).Gap(6).Children(func() {
					if note != "" {
						ui.Text(c, note).FontSize(12).TextColor(pal.muted)
					}
					if s.Timings != (controller.Timings{}) {
						ui.Text(c, lastTimingsText(m, s)).FontSize(12).TextColor(pal.muted)
					}
					if s.Text != "" {
						ui.Text(c, s.Text).Selectable()
					}
				})
			})
		})
	}
}

func (a *app) languageSelect(c *ui.Context, m locales.Messages) {
	labels := uiLanguageLabels(m)
	// The label of the system option changes with the language; pick the
	// current choice again when the bound label no longer matches one.
	if !slices.Contains(labels, a.uiLanguageLabel) {
		i := max(0, slices.Index(uiLanguages, a.uiLanguage))
		a.uiLanguageLabel = labels[i]
	}
	ui.Select(c.Key("ui-language"), &a.uiLanguageLabel, labels).Width(180).OnChange(func() {
		if i := slices.Index(labels, a.uiLanguageLabel); i >= 0 {
			a.setUILanguage(uiLanguages[i])
		}
	})
}

func (a *app) permissionsPage(c *ui.Context, m locales.Messages, pal palette) {
	t := c.Theme()
	group(c, pal, "", func(g *rows) {
		g.row(c, m.PermAccessibility(), m.PermAccessibilityNote(), func() {
			if a.perms.accessibility {
				pill(c, m.StatusAllowed(), t.Success)
				return
			}
			pill(c, m.StatusNeeded(), t.Danger)
			if ui.Button(c, m.OpenSettings()).Clicked() {
				darwin.OpenAccessibilitySettings()
			}
		})
		g.row(c, m.PermMicrophone(), micText(m, a.perms.microphone), func() {
			switch a.perms.microphone {
			case darwin.MicAuthorized:
				pill(c, m.StatusAllowed(), t.Success)
			case darwin.MicDenied, darwin.MicRestricted:
				pill(c, m.StatusNeeded(), t.Danger)
				if ui.Button(c, m.OpenSettings()).Clicked() {
					darwin.OpenMicrophoneSettings()
				}
			default:
				pill(c, m.StatusWaiting(), t.TextMuted)
			}
		})
		g.row(c, m.PermHotkeys(), m.PermHotkeysNote(), func() {
			if a.perms.tapRunning {
				pill(c, m.StatusRunning(), t.Success)
			} else {
				pill(c, m.StatusWaiting(), t.TextMuted)
			}
		})
	})
	group(c, pal, m.StatusTip(), func(g *rows) {
		g.row(c, m.PermGlobe(), m.PermGlobeNote(), func() {
			if ui.Button(c, m.OpenKeyboardSettings()).Clicked() {
				darwin.OpenKeyboardSettings()
			}
		})
	})
}

func (a *app) modelsPage(c *ui.Context, m locales.Messages, pal palette) {
	const fieldWidth = 240
	streaming := asrServices[a.form.asrService] == "soniox"
	group(c, pal, m.SectionSpeech(), func(g *rows) {
		g.row(c, m.FieldAsrService(), m.AsrServiceNote(), func() {
			ui.Segmented(c.Key("asr-service"), &a.form.asrService, "Soniox", "OpenRouter").OnChange(a.commit)
		})
	})
	if streaming {
		group(c, pal, "Soniox", func(g *rows) {
			g.row(c, m.FieldSonioxKey(), m.SonioxKeyNote(), func() {
				ui.TextInput(c.Key("soniox-key"), &a.form.sonioxKey).Password().Placeholder("Soniox API key").
					Width(fieldWidth).OnChange(func() { a.commitKey(a.sonioxKey, a.form.sonioxKey) })
			})
			g.row(c, m.FieldSonioxModel(), "", func() {
				ui.TextInput(c.Key("soniox-model"), &a.form.sonioxModel).Placeholder(soniox.DefaultModel).
					Width(fieldWidth).OnChange(a.commit)
			})
			g.row(c, m.FieldSonioxLanguages(), m.SonioxLanguagesNote(), func() {
				ui.TextInput(c.Key("soniox-langs"), &a.form.sonioxLangs).Placeholder("zh, en").
					Width(fieldWidth).OnChange(a.commit)
			})
		})
	}
	group(c, pal, m.SectionOpenRouter(), func(g *rows) {
		g.row(c, m.FieldApiKey(), m.OpenRouterKeyNote(), func() {
			ui.TextInput(c.Key("api-key"), &a.form.apiKey).Password().Placeholder("sk-or-…").
				Width(fieldWidth).OnChange(func() { a.commitKey(a.openrouterKey, a.form.apiKey) })
		})
		note := m.AsrModelNote()
		if streaming {
			note = m.AsrFallbackNote()
		}
		g.row(c, m.FieldAsrModel(), note, func() {
			ui.TextInput(c.Key("asr-model"), &a.form.asrModel).Placeholder("qwen/qwen3-asr-1.7b").
				Width(fieldWidth).OnChange(a.commit)
		})
		if !streaming {
			g.row(c, m.FieldAsrLanguage(), m.AsrLanguagePlaceholder(), func() {
				ui.TextInput(c.Key("asr-lang"), &a.form.asrLang).Placeholder("zh, en…").
					Width(fieldWidth).OnChange(a.commit)
			})
		}
	})
	group(c, pal, m.SectionRewrite(), func(g *rows) {
		g.row(c, m.FieldLlmModel(), m.LlmModelNote(), func() {
			ui.TextInput(c.Key("llm-model"), &a.form.llmModel).Placeholder("openai/gpt-oss-120b").
				Width(fieldWidth).OnChange(a.commit)
		})
		g.row(c, m.FieldProviders(), "", func() {
			ui.TextInput(c.Key("providers"), &a.form.providers).Placeholder("cerebras, groq").
				Width(fieldWidth).OnChange(a.commit)
		})
		g.row(c, m.FieldTimeout(), m.TimeoutNote(), func() {
			ui.TextInput(c.Key("timeout"), &a.form.timeout).Width(72).OnChange(a.commit)
		})
	})
	ui.Text(c, m.AutoSaveNote()).FontSize(11.5).TextColor(pal.muted)
}

func (a *app) translationPage(c *ui.Context, m locales.Messages, pal palette) {
	labels := make([]string, len(prompt.Languages))
	for i, l := range prompt.Languages {
		labels[i] = l.Label
	}
	group(c, pal, "", func(g *rows) {
		g.row(c, m.FieldChineseTo(), "", func() {
			ui.Select(c.Key("zh"), &a.form.zhTo, labels).Width(160).OnChange(a.commit)
		})
		g.row(c, m.FieldEnglishTo(), "", func() {
			ui.Select(c.Key("en"), &a.form.enTo, labels).Width(160).OnChange(a.commit)
		})
		g.row(c, m.FieldOtherTo(), "", func() {
			ui.Select(c.Key("default"), &a.form.defaultTo, labels).Width(160).OnChange(a.commit)
		})
		g.row(c, m.FieldRemember(), m.RememberNote(), func() {
			ui.Switch(c.Key("remember"), &a.form.remember).Label(m.FieldRemember()).OnChange(a.commit)
		})
	})
	group(c, pal, m.SectionTargets(), func(g *rows) {
		g.block(c, func() {
			ui.Grid(c).Columns(3).GapX(16).GapY(12).Children(func() {
				for i, l := range prompt.Languages {
					ui.Checkbox(c.Key("target-"+l.Code), &a.form.targets[i], l.Label).OnChange(a.commit)
				}
			})
		})
	})
}

func (a *app) dictionaryPage(c *ui.Context, m locales.Messages, pal palette) {
	group(c, pal, "", func(g *rows) {
		g.block(c, func() {
			ui.TextArea(c.Key("dictionary"), &a.form.dictionary).Height(240).
				Placeholder(m.DictionaryPlaceholder()).OnChange(a.commit)
		})
	})
	ui.Text(c, m.DictionaryNote()).FontSize(11.5).TextColor(pal.muted)
}

// rows builds the rows of a group, with a hairline between them.
type rows struct {
	pal palette
	n   int
}

// group is a titled card of rows, as System Settings groups its settings.
func group(c *ui.Context, pal palette, title string, body func(g *rows)) {
	ui.Column(c).Gap(8).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(12).Bold().TextColor(pal.muted).Padding(0, 4)
		}
		ui.Column(c).Radius(10).Background(pal.card).Border(1, pal.border).Clip().Children(func() {
			body(&rows{pal: pal})
		})
	})
}

func (g *rows) separator(c *ui.Context) {
	if g.n > 0 {
		ui.Box(c).Height(1).Background(g.pal.border).Margin(0, 0, 0, 14)
	}
	g.n++
}

// row is a label, with a note under it, and its control on the right.
func (g *rows) row(c *ui.Context, label, note string, control func()) {
	g.separator(c)
	ui.Row(c).MinHeight(44).Padding(9, 14).Gap(16).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(2).Children(func() {
			ui.Text(c, label)
			if note != "" {
				ui.Text(c, note).FontSize(11.5).TextColor(g.pal.muted)
			}
		})
		if control != nil {
			ui.Row(c).Shrink(0).Gap(8).AlignItems(ui.Center).Children(control)
		}
	})
}

// block is a row that is all content.
func (g *rows) block(c *ui.Context, content func()) {
	g.separator(c)
	ui.Column(c).Padding(12, 14).Children(content)
}

// keycaps shows a shortcut as keys.
func keycaps(c *ui.Context, pal palette, keys ...string) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		for _, k := range keys {
			ui.Text(c, k).FontSize(12).Padding(2, 8).Radius(5).Border(1, pal.border).
				Background(pal.card).TextColor(pal.muted).SingleLine()
		}
	})
}

// pill shows a status in its color.
func pill(c *ui.Context, text string, color ui.Color) {
	ui.Text(c, text).FontSize(11.5).Padding(2, 9).Radius(999).
		Background(ui.RGBA(color.R, color.G, color.B, 0.14)).TextColor(color).SingleLine()
}

// lastTimingsText says how long the last session's requests took.
func lastTimingsText(m locales.Messages, s controller.Status) string {
	asr, llm := formatDuration(s.Timings.Transcribe), formatDuration(s.Timings.Rewrite)
	if s.Mode == prompt.Translate {
		return m.LastTimingsTranslate(asr, llm)
	}
	return m.LastTimingsDictate(asr, llm)
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
