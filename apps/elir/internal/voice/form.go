//go:build darwin || windows

package voice

import (
	"slices"
	"strings"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/secret"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/config"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice/prompt"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
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
	sonioxLangs []string
	asrModel    string
	asrLang     string
	llmModel    string
	providers   []string
	timeout     float64 // seconds
	targets     []bool  // by prompt.Languages
	defaultTo   string  // language labels, as the selects show them
	zhTo        string
	enTo        string
	remember    bool
	dictionary  []string
	showTimings bool
	whileRec    int // index in whileRecording
	saveHistory bool
}

// whileRecording are the choices for the computer's sound while recording,
// as the settings keep them, in the order the control shows them.
var whileRecording = []string{"mute", "lower", "none"}

// asrServices are the speech-to-text services, as the settings keep them,
// in the order the segmented control shows them.
var asrServices = []string{"soniox", "openrouter"}

func formFrom(s config.Settings, keys apiKeys) settingsForm {
	f := settingsForm{
		apiKey:      keys.openrouter,
		sonioxKey:   keys.soniox,
		asrService:  max(0, slices.Index(asrServices, s.ASRService)),
		sonioxModel: s.SonioxModel,
		sonioxLangs: slices.Clone(s.SonioxLanguages),
		asrModel:    s.ASRModel,
		asrLang:     s.ASRLanguage,
		llmModel:    s.LLMModel,
		providers:   slices.Clone(s.Providers),
		timeout:     s.Timeout().Seconds(),
		targets:     make([]bool, len(prompt.Languages)),
		defaultTo:   prompt.LanguageFor(s.DefaultTarget).Label,
		zhTo:        prompt.LanguageFor(s.TargetFor("zh")).Label,
		enTo:        prompt.LanguageFor(s.TargetFor("en")).Label,
		remember:    s.RememberTarget,
		dictionary:  slices.Clone(s.Dictionary),
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

// cleanList trims the items of a token field, dropping empty and repeated
// ones.
func cleanList(items []string) []string {
	var out []string
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" && !slices.Contains(out, item) {
			out = append(out, item)
		}
	}
	return out
}

// apply copies the form into the settings.
func (f settingsForm) apply(s *config.Settings) {
	timeout := f.timeout
	if timeout <= 0 {
		timeout = config.Defaults().LLMTimeout
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
	s.SonioxLanguages = cleanList(f.sonioxLangs)
	s.ASRModel = strings.TrimSpace(f.asrModel)
	s.ASRLanguage = strings.TrimSpace(f.asrLang)
	s.LLMModel = strings.TrimSpace(f.llmModel)
	s.Providers = cleanList(f.providers)
	s.LLMTimeout = timeout
	s.Targets = targets
	s.DefaultTarget = defaultTo
	if s.Rules == nil {
		s.Rules = map[string]string{}
	}
	s.Rules["zh"] = codeFor(f.zhTo)
	s.Rules["en"] = codeFor(f.enTo)
	s.RememberTarget = f.remember
	s.Dictionary = cleanList(f.dictionary)
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
func (v *Voice) commit() {
	if err := v.settings.Update(v.form.apply); err != nil {
		v.saved = saveResult{saveFailed, err}
		return
	}
	if v.saved.state == saveFailed {
		v.saved = saveResult{}
	}
}

// commitKey writes a key to the Keychain once typing pauses.
func (v *Voice) commitKey(sec *secret.Secret, value string) {
	sec.SetLater(value, func(err error) {
		v.host.Settings.Update(func() { v.keySaved(err) })
	})
}

// flushKeys writes the keys still waiting for typing to pause.
func (v *Voice) flushKeys() {
	v.openrouterKey.Flush()
	v.sonioxKey.Flush()
}

// keySaved shows, or clears, a key that failed to save.
func (v *Voice) keySaved(err error) {
	switch {
	case err != nil:
		v.saved = saveResult{saveKeyFailed, err}
	case v.saved.state == saveKeyFailed:
		v.saved = saveResult{}
	}
}
