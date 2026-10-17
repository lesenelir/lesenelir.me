//go:build darwin

package secret

import (
	"os"
	"path/filepath"
	"testing"
)

// Keys live in a file only the user can read. Every name here is in the
// file before Load, so Load never reads the Keychain.
func TestStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Elir", "secrets.json")
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := st.Import("openrouter-api-key", "sk-or-1"); !ok || err != nil {
		t.Fatalf("import: %v, %v", ok, err)
	}
	if ok, _ := st.Import("openrouter-api-key", "sk-or-2"); ok {
		t.Fatal("import replaced a key the file has")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode %v, %v", info.Mode().Perm(), err)
	}

	key := New("openrouter-api-key", "TEST_NO_SUCH_ENV")
	key.Load(st)
	if key.Get() != "sk-or-1" {
		t.Fatalf("loaded %q", key.Get())
	}
	if err := key.Set("  sk-or-3 "); err != nil || key.Get() != "sk-or-3" {
		t.Fatalf("set: %q, %v", key.Get(), err)
	}
	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := reopened.lookup("openrouter-api-key"); v != "sk-or-3" {
		t.Fatalf("saved %q", v)
	}

	// A key removed stays removed: no Keychain, the environment's in
	// development.
	if err := key.Set(""); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_ELIR_KEY", "from-env")
	after, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	again := New("openrouter-api-key", "TEST_ELIR_KEY")
	again.Load(after)
	if again.Get() != "from-env" {
		t.Fatalf("removed key loaded as %q", again.Get())
	}
}

func TestSetBeforeLoad(t *testing.T) {
	if err := New("x", "").Set("v"); err == nil {
		t.Fatal("set a key with nowhere to keep it")
	}
}

// Two secrets of one name, as two tools hold for one API key, agree.
func TestSharedSecret(t *testing.T) {
	st, err := OpenStore(filepath.Join(t.TempDir(), "secrets.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = st.Import("openrouter-api-key", "sk-1")
	a, b := New("openrouter-api-key", ""), New("openrouter-api-key", "")
	a.Load(st)
	b.Load(st)
	if err := a.Set("sk-2"); err != nil {
		t.Fatal(err)
	}
	if b.Get() != "sk-2" {
		t.Fatalf("the other secret has %q", b.Get())
	}
}
