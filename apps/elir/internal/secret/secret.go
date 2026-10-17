//go:build darwin

// Package secret keeps API keys in the Keychain, or takes them from an
// environment variable in development.
package secret

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// TypingDelay is how long typing must pause before SetLater writes a key,
// so that the Keychain is not written on every keystroke.
const TypingDelay = 600 * time.Millisecond

// Secret is one API key. Get is safe from any goroutine; SetLater and
// Flush are for the main thread, where the settings' fields change.
type Secret struct {
	name string // the Keychain account
	env  string // the environment variable read when the Keychain has none

	mu    sync.RWMutex
	value string

	pending *time.Timer // a write waiting for typing to pause
	next    string      // what it writes
	done    func(error) // told when it is written
}

// New returns the secret kept under name in the Keychain, or in the
// environment variable env.
func New(name, env string) *Secret { return &Secret{name: name, env: env} }

// Name is the secret's Keychain account.
func (s *Secret) Name() string { return s.name }

// Load reads the secret.
func (s *Secret) Load() {
	b, err := mygo.Secrets.Get(s.name)
	if err != nil && !errors.Is(err, mygo.ErrSecretNotFound) {
		log.Printf("keychain %s: %v", s.name, err)
	}
	v := string(b)
	if v == "" {
		v = os.Getenv(s.env)
	}
	s.mu.Lock()
	s.value = v
	s.mu.Unlock()
}

// Get returns the secret, "" when there is none.
func (s *Secret) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// Set writes the secret to the Keychain, or removes it there when empty.
func (s *Secret) Set(v string) error {
	v = strings.TrimSpace(v)
	var err error
	if v == "" {
		err = mygo.Secrets.Delete(s.name)
	} else {
		err = mygo.Secrets.Set(s.name, []byte(v))
	}
	if err != nil {
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
