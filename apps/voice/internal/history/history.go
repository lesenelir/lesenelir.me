// Package history keeps the text of past dictations and translations in a
// local file, one JSON object per line, newest last. It keeps no audio.
package history

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"
)

// MaxEntries is how many entries the file keeps; older ones are dropped.
const MaxEntries = 1000

// Entry is one pasted result.
type Entry struct {
	ID   string    `json:"id"`
	Time time.Time `json:"time"`
	// Mode is "dictate" or "translate"; Target the translation's language.
	Mode   string `json:"mode"`
	Target string `json:"target,omitempty"`
	// Transcript is what speech to text heard; Text what was pasted.
	Transcript string `json:"transcript"`
	Text       string `json:"text"`
	// TranscribeMS and RewriteMS are how long the requests took.
	TranscribeMS int64 `json:"transcribeMs,omitempty"`
	RewriteMS    int64 `json:"rewriteMs,omitempty"`
}

// Store is the history file. It is safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
	// count is the number of entries in the file, -1 until read.
	count int
}

// Open returns the history kept at path; the file is made on the first Add.
func Open(path string) *Store { return &Store{path: path, count: -1} }

// Add appends an entry, giving it an ID and time if it has none, and drops
// the oldest beyond MaxEntries.
func (s *Store) Add(e Entry) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	if e.ID == "" {
		e.ID = strconv.FormatInt(e.Time.UnixNano(), 36)
	}
	line, err := json.Marshal(e)
	if err != nil {
		return e, err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return e, err
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return e, err
	}
	_, err = f.Write(append(line, '\n'))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return e, err
	}
	if s.count >= 0 {
		s.count++
	}
	if s.count < 0 || s.count > MaxEntries {
		entries, err := s.read()
		if err != nil {
			return e, err
		}
		if len(entries) > MaxEntries {
			if err := s.write(entries[len(entries)-MaxEntries:]); err != nil {
				return e, err
			}
		}
	}
	return e, nil
}

// List returns the entries, newest first.
func (s *Store) List() ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.read()
	slices.Reverse(entries)
	return entries, err
}

// Delete removes the entry with the ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.read()
	if err != nil {
		return err
	}
	kept := slices.DeleteFunc(entries, func(e Entry) bool { return e.ID == id })
	return s.write(kept)
}

// Restore puts back an entry Delete took out, in its place by time.
func (s *Store) Restore(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := s.read()
	if err != nil {
		return err
	}
	if slices.ContainsFunc(entries, func(x Entry) bool { return x.ID == e.ID }) {
		return nil
	}
	i, _ := slices.BinarySearchFunc(entries, e, func(a, b Entry) int { return a.Time.Compare(b.Time) })
	return s.write(slices.Insert(entries, i, e))
}

// Clear deletes the history file.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	s.count = 0
	return nil
}

// Size is the history file's size in bytes.
func (s *Store) Size() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, err := os.Stat(s.path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// read returns the entries oldest first, skipping lines it cannot read.
func (s *Store) read() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		s.count = 0
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		var e Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.ID != "" {
			entries = append(entries, e)
		}
	}
	s.count = len(entries)
	return entries, sc.Err()
}

// write replaces the file with the entries, oldest first.
func (s *Store) write(entries []Entry) error {
	if len(entries) == 0 {
		s.count = 0
		if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	var buf bytes.Buffer
	for _, e := range entries {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return err
	}
	s.count = len(entries)
	return nil
}
