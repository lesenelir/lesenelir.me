//go:build darwin

// Package secret keeps API keys in a file of the app's data that only the
// user can read, or takes them from an environment variable in
// development.
//
// They were in the Keychain until Elir 0.2.3. A self-signed app has no Team
// ID, so the Keychain binds its items to the exact build that wrote them,
// and asked for the login password after every update. A key not yet in
// the file is read from the Keychain once, which may ask one last time,
// and kept in the file from then on.
package secret

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// TypingDelay is how long typing must pause before SetLater writes a key,
// so that the file is not written on every keystroke.
const TypingDelay = 600 * time.Millisecond

// Store is the file of the keys: a JSON object of each key by its name,
// readable and writable only by the user. It is safe for concurrent use.
type Store struct {
	path string

	mu     sync.Mutex
	values map[string]string // a name present with "" is a key removed
}

// OpenStore reads the keys at path. A missing file is an empty store.
func OpenStore(path string) (*Store, error) {
	st := &Store{path: path, values: map[string]string{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	return st, json.Unmarshal(data, &st.values)
}

// lookup returns a key, and whether the file has it, even as removed.
func (st *Store) lookup(name string) (string, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	v, ok := st.values[name]
	return v, ok
}

// put keeps a key, "" to remove it, and saves the file.
func (st *Store) put(name, value string) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.values[name] = value
	data, err := json.MarshalIndent(st.values, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o700); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}

// Import keeps a key from elsewhere, unless the file has it already, and
// reports whether it did.
func (st *Store) Import(name, value string) (bool, error) {
	if _, ok := st.lookup(name); ok || value == "" {
		return false, nil
	}
	return true, st.put(name, value)
}

// Secret is one API key. Get is safe from any goroutine; SetLater and
// Flush are for the main thread, where the settings' fields change.
type Secret struct {
	name string // its name in the file, and its Keychain account before
	env  string // the environment variable read when the file has none

	mu    sync.RWMutex
	st    *Store
	value string

	pending *time.Timer // a write waiting for typing to pause
	next    string      // what it writes
	done    func(error) // told when it is written
}

// New returns the secret kept under name, or in the environment variable
// env.
func New(name, env string) *Secret { return &Secret{name: name, env: env} }

// Name is the secret's name in the file.
func (s *Secret) Name() string { return s.name }

// Load reads the secret from st, which keeps it from then on. A secret
// st has never had is read from the Keychain, once.
func (s *Secret) Load(st *Store) {
	v, ok := st.lookup(s.name)
	if !ok {
		v = fromKeychain(s.name)
		if v != "" {
			if err := st.put(s.name, v); err != nil {
				log.Printf("secret %s: %v", s.name, err)
			}
		}
	}
	if v == "" {
		v = os.Getenv(s.env)
	}
	s.mu.Lock()
	s.st, s.value = st, v
	s.mu.Unlock()
}

// fromKeychain reads a key the app kept in the Keychain before, "" when
// there is none.
func fromKeychain(name string) string {
	b, err := mygo.Secrets.Get(name)
	if err != nil && !errors.Is(err, mygo.ErrSecretNotFound) {
		log.Printf("keychain %s: %v", name, err)
	}
	return strings.TrimSpace(string(b))
}

// Get returns the secret, "" when there is none. It reads the store, so
// that secrets of the same name, as two tools have for one API key, agree.
func (s *Secret) Get() string {
	s.mu.RLock()
	st, v := s.st, s.value
	s.mu.RUnlock()
	if st != nil {
		if stored, ok := st.lookup(s.name); ok {
			if stored == "" {
				return os.Getenv(s.env)
			}
			return stored
		}
	}
	return v
}

// Set keeps the secret, or removes it when empty.
func (s *Secret) Set(v string) error {
	v = strings.TrimSpace(v)
	s.mu.RLock()
	st := s.st
	s.mu.RUnlock()
	if st == nil {
		return errors.New("secret: not loaded")
	}
	if err := st.put(s.name, v); err != nil {
		return err
	}
	s.mu.Lock()
	s.value = v
	s.mu.Unlock()
	return nil
}

// SetLater writes v once typing pauses for TypingDelay, then tells done,
// on a goroutine of its own. A later call replaces it.
func (s *Secret) SetLater(v string, done func(error)) {
	if s.pending != nil {
		s.pending.Stop()
	}
	s.next, s.done = v, done
	s.pending = time.AfterFunc(TypingDelay, func() { done(s.Set(v)) })
}

// Flush writes at once a value SetLater is still waiting to write.
func (s *Secret) Flush() {
	if s.pending != nil && s.pending.Stop() {
		s.pending = nil
		s.done(s.Set(s.next))
	}
}
