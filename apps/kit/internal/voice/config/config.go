// Package config is voice input's settings: the "voice" section of the
// settings file. The API keys are not here: they live in the Keychain.
package config

import (
	"maps"
	"slices"
	"strings"
	"time"
)

// Section is the settings file's section for voice input.
const Section = "voice"

// Settings are what the user can change about voice input.
type Settings struct {
	// HideTimings keeps how long transcription and cleanup took off the
	// voice bar.
	HideTimings bool `json:"hideTimings"`
	// NoHistory keeps pasted text out of the local history.
	NoHistory bool `json:"noHistory"`
	// WhileRecording is what happens to the computer's own sound while the
	// microphone records, so that music and videos stay out of the text:
	// "mute", "lower" or "none".
	WhileRecording string `json:"whileRecording"`

	// ASRService is "soniox", to transcribe while recording, or
	// "openrouter", to transcribe the recording after it ends. Without a
	// Soniox API key, or when Soniox fails, OpenRouter does it.
	ASRService string `json:"asrService"`
	// SonioxModel is the real-time model; SonioxLanguages hint at the
	// languages spoken, as ISO 639-1 codes.
	SonioxModel     string   `json:"sonioxModel"`
	SonioxLanguages []string `json:"sonioxLanguages"`

	// ASRModel is the OpenRouter speech-to-text model; ASRLanguage an
	// ISO 639-1 hint, "" to detect it.
	ASRModel    string `json:"asrModel"`
	ASRLanguage string `json:"asrLanguage"`
	// LLMModel cleans up and translates; Providers are tried first, in
	// order. Reasoning is the effort, "" to pick one for the model.
	LLMModel   string   `json:"llmModel"`
	Providers  []string `json:"providers"`
	Reasoning  string   `json:"reasoning"`
	LLMTimeout float64  `json:"llmTimeoutSeconds"`

	// Targets are the translation targets, in the order Tab cycles them.
	Targets       []string `json:"targets"`
	DefaultTarget string   `json:"defaultTarget"`
	// Rules map a spoken language to the target it translates into.
	Rules map[string]string `json:"rules"`
	// RememberTarget starts a translation in the last target picked on
	// the voice bar, rather than by the rules.
	RememberTarget bool   `json:"rememberTarget"`
	LastTarget     string `json:"lastTarget"`

	// Dictionary holds names and terms to spell as written.
	Dictionary []string `json:"dictionary"`
}

// Defaults are the settings of a new install.
func Defaults() Settings {
	return Settings{
		WhileRecording:  "mute",
		ASRService:      "soniox",
		SonioxModel:     "stt-rt-v5",
		SonioxLanguages: []string{"zh", "en"},
		ASRModel:        "qwen/qwen3-asr-1.7b",
		LLMModel:        "openai/gpt-oss-120b",
		Providers:       []string{"cerebras", "groq"},
		LLMTimeout:      5,
		Targets:         []string{"en", "zh", "ja"},
		DefaultTarget:   "en",
		Rules:           map[string]string{"zh": "en", "en": "zh"},
	}
}

// ReasoningEffort returns the reasoning effort to request. gpt-oss cannot
// turn reasoning off, so it gets the lowest; others get none, since every
// reasoning token delays the text.
func (s Settings) ReasoningEffort() string {
	if s.Reasoning != "" {
		return s.Reasoning
	}
	if strings.HasPrefix(s.LLMModel, "openai/gpt-oss") {
		return "low"
	}
	return "none"
}

// Timeout is how long a rewrite may take before the raw transcript is
// pasted instead.
func (s Settings) Timeout() time.Duration {
	if s.LLMTimeout <= 0 {
		return 5 * time.Second
	}
	return time.Duration(s.LLMTimeout * float64(time.Second))
}

// InitialTarget is the target a translation shows when it starts.
func (s Settings) InitialTarget() string {
	if s.RememberTarget && s.LastTarget != "" {
		return s.LastTarget
	}
	return s.DefaultTarget
}

// TargetFor returns the target for speech in the language spoken.
func (s Settings) TargetFor(spoken string) string {
	if t, ok := s.Rules[spoken]; ok && t != "" {
		return t
	}
	return s.DefaultTarget
}

// NextTarget returns the target after current, as Tab moves.
func (s Settings) NextTarget(current string) string {
	if len(s.Targets) == 0 {
		return current
	}
	i := slices.Index(s.Targets, current)
	return s.Targets[(i+1)%len(s.Targets)]
}

// Clone copies the settings deeply, as the store hands them out.
func Clone(s Settings) Settings {
	s.Providers = slices.Clone(s.Providers)
	s.SonioxLanguages = slices.Clone(s.SonioxLanguages)
	s.Targets = slices.Clone(s.Targets)
	s.Dictionary = slices.Clone(s.Dictionary)
	s.Rules = maps.Clone(s.Rules)
	return s
}
