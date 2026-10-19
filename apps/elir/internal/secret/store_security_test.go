//go:build darwin || windows

package secret

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUnreadableSecretStoreIsMovedAside(t *testing.T) {
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
			if _, ok := st.lookup("old"); ok {
				t.Fatal("published a partially decoded key")
			}
			aside, _ := filepath.Glob(path + ".unreadable-*")
			if len(aside) != 1 {
				t.Fatalf("unreadable store not kept aside: %v", aside)
			}
			if kept, err := os.ReadFile(aside[0]); err != nil || !bytes.Equal(before, kept) {
				t.Fatalf("unreadable store changed: %v", err)
			}
			if err := st.put("new", "saved"); err != nil {
				t.Fatalf("a new key cannot be saved: %v", err)
			}
			reopened, err := OpenStore(path)
			if err != nil {
				t.Fatal(err)
			}
			if v, _ := reopened.lookup("new"); v != "saved" {
				t.Fatal("new key not saved")
			}
		})
	}
}

func TestStoreThatCannotBeReadIsNotReplaced(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.json")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	st, err := OpenStore(path)
	if err == nil {
		t.Fatal("read a directory as the store")
	}
	if err := st.put("new", "must-not-save"); err == nil {
		t.Fatal("replaced a store it could not read")
	}
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		t.Fatalf("store moved or replaced: %v", err)
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
