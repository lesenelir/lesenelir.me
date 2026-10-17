package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type appSettings struct {
	Theme string `json:"theme"`
}

type toolSettings struct {
	Model string   `json:"model"`
	Tags  []string `json:"tags"`
}

func cloneTool(t toolSettings) toolSettings {
	t.Tags = slices.Clone(t.Tags)
	return t
}

func TestSectionsPersistApart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.json")
	st, err := Open(path)
	if err != nil || st.Existed() {
		t.Fatalf("new store: existed %v, err %v", st.Existed(), err)
	}
	app, _ := NewSection(st, "app", appSettings{Theme: "system"}, nil)
	tool, _ := NewSection(st, "tool", toolSettings{Model: "a", Tags: []string{"x"}}, cloneTool)
	if err := app.Update(func(s *appSettings) { s.Theme = "dark" }); err != nil {
		t.Fatal(err)
	}
	if err := tool.Update(func(s *toolSettings) { s.Tags = append(s.Tags, "y") }); err != nil {
		t.Fatal(err)
	}

	again, err := Open(path)
	if err != nil || !again.Existed() {
		t.Fatalf("reopened: existed %v, err %v", again.Existed(), err)
	}
	app2, _ := NewSection(again, "app", appSettings{Theme: "system"}, nil)
	tool2, _ := NewSection(again, "tool", toolSettings{Model: "a"}, cloneTool)
	if got := app2.Get().Theme; got != "dark" {
		t.Errorf("theme %q", got)
	}
	if got := tool2.Get(); got.Model != "a" || !slices.Equal(got.Tags, []string{"x", "y"}) {
		t.Errorf("tool %+v", got)
	}
}

// A setting the file does not have yet, as one added in a later version,
// keeps its default.
func TestNewFieldsGetDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	_ = os.WriteFile(path, []byte(`{"tool":{"tags":["kept"]}}`), 0o644)
	st, _ := Open(path)
	tool, err := NewSection(st, "tool", toolSettings{Model: "default"}, cloneTool)
	if err != nil {
		t.Fatal(err)
	}
	if got := tool.Get(); got.Model != "default" || !slices.Equal(got.Tags, []string{"kept"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestGetIsACopy(t *testing.T) {
	st, _ := Open(filepath.Join(t.TempDir(), "s.json"))
	tool, _ := NewSection(st, "tool", toolSettings{Tags: []string{"x"}}, cloneTool)
	v := tool.Get()
	v.Tags[0] = "changed"
	if got := tool.Get().Tags[0]; got != "x" {
		t.Fatalf("Get shares state: %q", got)
	}
}

func TestCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	_ = os.WriteFile(path, []byte("{"), 0o644)
	st, err := Open(path)
	if err == nil {
		t.Fatal("no error for a corrupt file")
	}
	app, _ := NewSection(st, "app", appSettings{Theme: "system"}, nil)
	if app.Get().Theme != "system" {
		t.Fatal("no defaults after a corrupt file")
	}
}

func TestBadSectionKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	_ = os.WriteFile(path, []byte(`{"app":{"theme":42}}`), 0o644)
	st, _ := Open(path)
	app, err := NewSection(st, "app", appSettings{Theme: "system"}, nil)
	if err == nil || app.Get().Theme != "system" {
		t.Fatalf("theme %q, err %v", app.Get().Theme, err)
	}
}

func TestSetRaw(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	st, _ := Open(path)
	if err := st.SetRaw("app", json.RawMessage(`{"theme":"light"}`)); err != nil {
		t.Fatal(err)
	}
	again, _ := Open(path)
	app, _ := NewSection(again, "app", appSettings{}, nil)
	if app.Get().Theme != "light" {
		t.Fatalf("theme %q", app.Get().Theme)
	}
}
