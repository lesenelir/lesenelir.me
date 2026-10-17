package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenMissingGivesDefaults(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := st.Get(); got.ASRModel != Defaults().ASRModel || got.DefaultTarget != "en" {
		t.Fatalf("got %+v", got)
	}
}

func TestUpdatePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.json")
	st, _ := Open(path)
	if err := st.Update(func(s *Settings) { s.LLMModel = "anthropic/claude-haiku-5.5" }); err != nil {
		t.Fatal(err)
	}
	again, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := again.Get().LLMModel; got != "anthropic/claude-haiku-5.5" {
		t.Fatalf("LLMModel = %q", got)
	}
}

func TestCorruptFileFallsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	_ = os.WriteFile(path, []byte("{"), 0o644)
	st, err := Open(path)
	if err == nil {
		t.Fatal("no error for a corrupt file")
	}
	if st.Get().ASRModel == "" {
		t.Fatal("corrupt file left no defaults")
	}
}

func TestGetIsACopy(t *testing.T) {
	st, _ := Open(filepath.Join(t.TempDir(), "s.json"))
	s := st.Get()
	s.Rules["zh"] = "ja"
	s.Targets[0] = "ko"
	if got := st.Get(); got.Rules["zh"] != "en" || got.Targets[0] != "en" {
		t.Fatalf("Get shares state: %+v", got)
	}
}

func TestTargets(t *testing.T) {
	s := Defaults()
	if got := s.TargetFor("zh"); got != "en" {
		t.Errorf("zh → %q", got)
	}
	if got := s.TargetFor("en"); got != "zh" {
		t.Errorf("en → %q", got)
	}
	if got := s.TargetFor("fr"); got != "en" {
		t.Errorf("fr → %q, want the default", got)
	}
	if got := s.NextTarget("en"); got != "zh" {
		t.Errorf("after en %q", got)
	}
	if got := s.NextTarget("ja"); got != "en" {
		t.Errorf("after ja %q", got)
	}
	if got := s.NextTarget("ko"); got != "en" {
		t.Errorf("after an unlisted target %q", got)
	}
	s.RememberTarget, s.LastTarget = true, "ja"
	if got := s.InitialTarget(); got != "ja" {
		t.Errorf("initial %q", got)
	}
}

func TestReasoningEffort(t *testing.T) {
	s := Defaults()
	if got := s.ReasoningEffort(); got != "low" {
		t.Errorf("gpt-oss effort %q", got)
	}
	s.LLMModel = "anthropic/claude-haiku-5.5"
	if got := s.ReasoningEffort(); got != "none" {
		t.Errorf("haiku effort %q", got)
	}
	s.Reasoning = "minimal"
	if got := s.ReasoningEffort(); got != "minimal" {
		t.Errorf("explicit effort %q", got)
	}
}
