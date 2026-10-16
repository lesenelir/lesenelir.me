// Package store keeps the user's settings in a JSON file in the app's data
// directory. The API key is not here: it lives in the Keychain.
package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// Settings are what the user can change.
type Settings struct {
	// UILanguage is the language of the app's interface, as a BCP 47 tag
	// ("en", "zh-Hans"); "" follows the system.
	UILanguage string `json:"uiLanguage"`

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
		ASRModel:      "qwen/qwen3-asr-1.7b",
		LLMModel:      "openai/gpt-oss-120b",
		Providers:     []string{"cerebras", "groq"},
		LLMTimeout:    5,
		Targets:       []string{"en", "zh", "ja"},
		DefaultTarget: "en",
		Rules:         map[string]string{"zh": "en", "en": "zh"},
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

// Store holds the settings and saves every change. It is safe for
// concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
	s    Settings
}

// Open loads the settings at path, or the defaults when there are none.
func Open(path string) (*Store, error) {
	st := &Store{path: path, s: Defaults()}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(data, &st.s); err != nil {
		return &Store{path: path, s: Defaults()}, err
	}
	return st, nil
}

// Get returns a copy of the settings.
func (st *Store) Get() Settings {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.s
	s.Providers = slices.Clone(s.Providers)
	s.Targets = slices.Clone(s.Targets)
	s.Dictionary = slices.Clone(s.Dictionary)
	s.Rules = maps.Clone(s.Rules)
	return s
}

// Update changes the settings and saves them.
func (st *Store) Update(fn func(*Settings)) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	fn(&st.s)
	data, err := json.MarshalIndent(st.s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o755); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}
