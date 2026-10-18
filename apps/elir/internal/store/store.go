// Package store keeps the user's settings in one JSON file in the app's data
// directory, a section per part of the app: "app" for the shell, "voice" for
// voice input, and one for each tool added later. Each part reads and writes
// only its own section, as a Go type of its own. API keys are not here: they
// live in the Keychain.
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sync"
)

// Store is the settings file. It is safe for concurrent use.
type Store struct {
	path    string
	existed bool

	mu       sync.Mutex
	sections map[string]json.RawMessage
	loadErr  error
}

// Open reads the settings at path. A missing file is an empty store; a
// corrupt or unreadable one uses defaults and rejects writes until reopened.
func Open(path string) (*Store, error) {
	st := &Store{path: path, sections: map[string]json.RawMessage{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	st.existed = true
	if err != nil {
		st.loadErr = err
		return st, err
	}
	if err := json.Unmarshal(data, &st.sections); err != nil {
		st.sections = map[string]json.RawMessage{}
		st.loadErr = err
		return st, err
	}
	if st.sections == nil {
		st.sections = map[string]json.RawMessage{}
	}
	return st, nil
}

// Existed reports whether the file was there when the store opened, which
// a new install's is not.
func (st *Store) Existed() bool { return st.existed }

// Raw returns a section as it is stored, nil when there is none.
func (st *Store) Raw(key string) json.RawMessage {
	st.mu.Lock()
	defer st.mu.Unlock()
	return bytes.Clone(st.sections[key])
}

// SetRaw replaces a section and saves the file.
func (st *Store) SetRaw(key string, raw json.RawMessage) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.loadErr != nil {
		return st.loadErr
	}
	candidate := maps.Clone(st.sections)
	candidate[key] = bytes.Clone(raw)
	if err := st.save(candidate); err != nil {
		return err
	}
	st.sections = candidate
	return nil
}

// save writes the file, through a temporary one so that a crash cannot
// leave it half written. The caller holds mu.
func (st *Store) save(candidate map[string]json.RawMessage) error {
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o755); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}

// Section is one part's settings, of type T. It is safe for concurrent
// use.
type Section[T any] struct {
	st    *Store
	key   string
	clone func(T) T

	mu    sync.Mutex
	value T
}

// NewSection returns the section key of st: what the file holds over
// defaults, so that settings added later get their defaults. clone copies
// a value deeply, so that Get hands out values no one else changes; nil
// copies it as Go assigns it, enough for T without slices or maps. A
// section the file cannot be read into keeps its defaults, with the error.
func NewSection[T any](st *Store, key string, defaults T, clone func(T) T) (*Section[T], error) {
	if clone == nil {
		clone = func(v T) T { return v }
	}
	s := &Section[T]{st: st, key: key, clone: clone, value: clone(defaults)}
	var err error
	if raw := st.Raw(key); raw != nil {
		v := clone(defaults)
		if err = json.Unmarshal(raw, &v); err == nil {
			s.value = v
		}
	}
	return s, err
}

// Get returns a copy of the section.
func (s *Section[T]) Get() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clone(s.value)
}

// Update changes the section and saves the file.
func (s *Section[T]) Update(fn func(*T)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	candidate := s.clone(s.value)
	fn(&candidate)
	raw, err := json.Marshal(candidate)
	if err != nil {
		return err
	}
	if err := s.st.SetRaw(s.key, raw); err != nil {
		return err
	}
	s.value = candidate
	return nil
}
