package secret

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsSecretsEncryptedAndPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secrets.json")
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	s := New("test-key", "ELIR_TEST_UNUSED_KEY")
	s.Load(st)
	for _, value := range []string{"first-test-secret", "second-test-secret"} {
		if err := s.Set(value); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil || bytes.Contains(data, []byte(value)) {
			t.Fatalf("plaintext secret or read failure: %v", err)
		}
		reloaded, err := OpenStore(path)
		if err != nil {
			t.Fatal(err)
		}
		copy := New("test-key", "ELIR_TEST_UNUSED_KEY")
		copy.Load(reloaded)
		if copy.Get() != value {
			t.Fatal("saved secret did not round-trip")
		}
	}
	if err := s.Set(""); err != nil {
		t.Fatal(err)
	}
	if s.Get() != "" {
		t.Fatal("removed secret remains")
	}
}

func TestWindowsCorruptSecretsAreMovedAside(t *testing.T) {
	data, err := protectData([]byte("secret"), true)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 1
	if _, err := protectData(data, false); err == nil {
		t.Fatal("accepted corrupt encrypted data")
	}
	path := filepath.Join(t.TempDir(), "secrets.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	st, err := OpenStore(path)
	if err == nil {
		t.Fatal("opened corrupt encrypted store")
	}
	aside, _ := filepath.Glob(path + ".unreadable-*")
	if len(aside) != 1 {
		t.Fatalf("corrupt store not kept aside: %v", aside)
	}
	if kept, err := os.ReadFile(aside[0]); err != nil || !bytes.Equal(data, kept) {
		t.Fatalf("corrupt original changed: %v", err)
	}
	if err := st.put("new", "saved"); err != nil {
		t.Fatalf("a new key cannot be saved: %v", err)
	}
}
