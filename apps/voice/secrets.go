//go:build darwin

package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// secret is an API key kept in the Keychain, or taken from an environment
// variable in development.
type secret struct {
	name string // the Keychain account
	env  string // the environment variable read when the Keychain has none

	mu    sync.RWMutex
	value string

	// timer is the pending write of a key being typed; main thread only.
	timer *time.Timer
}

func newSecret(name, env string) *secret { return &secret{name: name, env: env} }

func (s *secret) load() {
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

// Get returns the key, "" when there is none. It is safe from any
// goroutine.
func (s *secret) Get() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// set writes the key to the Keychain, or removes it there when empty.
func (s *secret) set(v string) error {
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
