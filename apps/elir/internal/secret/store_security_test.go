//go:build darwin || windows

package secret

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUnreadableSecretStoreCannotBeOverwritten(t *testing.T) {
	for _, content := range []string{"not a secret file", "null", `{"old":"value","bad":1}`} {
		t.Run(content, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "secrets.json")
			if err := writeSecretFile(path, []byte(content)); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			st, err := OpenStore(path)
			if err == nil {
				t.Fatal("accepted invalid secret store")
			}
			if err := st.put("new", "must-not-save"); err == nil {
				t.Fatal("overwrote failed store")
			}
			if _, ok := st.lookup("old"); ok {
				t.Fatal("published a partially decoded key")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("original store changed: %v", err)
			}
		})
	}
}

func TestFailedSecretWriteKeepsActiveValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.json")
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.put("key", "saved"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".tmp", 0700); err != nil {
		t.Fatal(err)
	}
	if err := st.put("key", "unsaved"); err == nil {
		t.Fatal("expected write failure")
	}
	if value, _ := st.lookup("key"); value != "saved" {
		t.Fatal("failed write changed the active key")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("failed write changed stored key: %v", err)
	}
}
