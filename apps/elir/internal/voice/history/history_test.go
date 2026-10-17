package history

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddListDeleteClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "history.jsonl")
	s := Open(path)
	if entries, err := s.List(); err != nil || len(entries) != 0 {
		t.Fatalf("empty history: %v, %v", entries, err)
	}
	base := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	for i, text := range []string{"第一条", "second", "第三条"} {
		if _, err := s.Add(Entry{Time: base.Add(time.Duration(i) * time.Minute), Mode: "dictate", Text: text}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := s.List()
	if err != nil || len(entries) != 3 || entries[0].Text != "第三条" || entries[2].Text != "第一条" {
		t.Fatalf("newest first: %+v, %v", entries, err)
	}
	if s.Size() == 0 {
		t.Fatal("no size")
	}

	// A fresh store reads what was written.
	again := Open(path)
	if got, _ := again.List(); len(got) != 3 {
		t.Fatalf("reopened: %d entries", len(got))
	}

	if err := s.Delete(entries[1].ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.List(); len(got) != 2 || got[0].Text != "第三条" || got[1].Text != "第一条" {
		t.Fatalf("after delete: %+v", got)
	}

	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("the file is still there after Clear")
	}
	if got, _ := s.List(); len(got) != 0 || s.Size() != 0 {
		t.Fatalf("after clear: %+v", got)
	}
	if err := s.Clear(); err != nil {
		t.Fatalf("clearing twice: %v", err)
	}
}

func TestKeepsTheNewest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	s := Open(path)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range MaxEntries + 5 {
		if _, err := s.Add(Entry{Time: base.Add(time.Duration(i) * time.Second), Text: fmt.Sprint(i)}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := Open(path).List()
	if err != nil || len(entries) != MaxEntries {
		t.Fatalf("%d entries, %v", len(entries), err)
	}
	if entries[0].Text != fmt.Sprint(MaxEntries+4) || entries[len(entries)-1].Text != "5" {
		t.Fatalf("kept %s … %s", entries[0].Text, entries[len(entries)-1].Text)
	}
}

func TestSkipsBrokenLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	_ = os.WriteFile(path, []byte("{broken\n"+`{"id":"a","text":"ok"}`+"\n"), 0o600)
	entries, err := Open(path).List()
	if err != nil || len(entries) != 1 || entries[0].Text != "ok" {
		t.Fatalf("%+v, %v", entries, err)
	}
}

func TestRestore(t *testing.T) {
	s := Open(filepath.Join(t.TempDir(), "history.jsonl"))
	base := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	var middle Entry
	for i, text := range []string{"a", "b", "c"} {
		e, _ := s.Add(Entry{Time: base.Add(time.Duration(i) * time.Minute), Text: text})
		if text == "b" {
			middle = e
		}
	}
	_ = s.Delete(middle.ID)
	if err := s.Restore(middle); err != nil {
		t.Fatal(err)
	}
	_ = s.Restore(middle) // twice: kept once
	got, _ := s.List()
	if len(got) != 3 || got[0].Text != "c" || got[1].Text != "b" || got[2].Text != "a" {
		t.Fatalf("after restore %+v", got)
	}
}
