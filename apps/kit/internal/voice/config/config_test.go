package config

import "testing"

func TestCloneIsDeep(t *testing.T) {
	s := Defaults()
	c := Clone(s)
	c.Rules["zh"] = "ja"
	c.Targets[0] = "ko"
	c.Providers[0] = "x"
	if s.Rules["zh"] != "en" || s.Targets[0] != "en" || s.Providers[0] != "cerebras" {
		t.Fatalf("Clone shares state: %+v", s)
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
