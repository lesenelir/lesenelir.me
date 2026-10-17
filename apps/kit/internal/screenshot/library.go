package screenshot

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Shot is a screenshot kept on disk.
type Shot struct {
	Path string
	Time time.Time // when it was taken: the file's modification time
	Size int64
}

// Library is the screenshots Kit keeps, a PNG file each in a directory of
// its data: never on the Desktop.
type Library struct {
	dir string
}

// OpenLibrary returns the library in dir, which it makes if need be.
func OpenLibrary(dir string) (*Library, error) {
	return &Library{dir: dir}, os.MkdirAll(dir, 0o755)
}

// NewPath names the file of a screenshot taken at t, as macOS names its
// own, numbered when one of that second exists already.
func (l *Library) NewPath(t time.Time) string {
	base := "Screenshot " + t.Format("2006-01-02") + " at " + t.Format("15.04.05")
	path := filepath.Join(l.dir, base+".png")
	for n := 2; ; n++ {
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return path
		}
		path = filepath.Join(l.dir, fmt.Sprintf("%s (%d).png", base, n))
	}
}

// List returns the screenshots, newest first.
func (l *Library) List() ([]Shot, error) {
	entries, err := os.ReadDir(l.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var shots []Shot
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".png") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue // gone meanwhile
		}
		shots = append(shots, Shot{Path: filepath.Join(l.dir, e.Name()), Time: info.ModTime(), Size: info.Size()})
	}
	slices.SortFunc(shots, func(a, b Shot) int { return b.Time.Compare(a.Time) })
	return shots, nil
}

// Prune deletes the screenshots older than keep, none when keep is 0, and
// reports how many it deleted.
func (l *Library) Prune(keep time.Duration, now time.Time) (int, error) {
	if keep <= 0 {
		return 0, nil
	}
	shots, err := l.List()
	if err != nil {
		return 0, err
	}
	n := 0
	var errs []error
	for _, s := range shots {
		if now.Sub(s.Time) > keep {
			if err := os.Remove(s.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				errs = append(errs, err)
				continue
			}
			n++
		}
	}
	return n, errors.Join(errs...)
}

// Delete deletes a screenshot of the library.
func (l *Library) Delete(path string) error {
	if filepath.Dir(path) != l.dir {
		return fmt.Errorf("screenshot: %s is not in the library", path)
	}
	err := os.Remove(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// Restore puts back a deleted screenshot, with the time it was taken.
func (l *Library) Restore(s Shot, data []byte) error {
	if filepath.Dir(s.Path) != l.dir {
		return fmt.Errorf("screenshot: %s is not in the library", s.Path)
	}
	if err := os.WriteFile(s.Path, data, 0o644); err != nil {
		return err
	}
	return os.Chtimes(s.Path, s.Time, s.Time)
}

// Clear deletes every screenshot.
func (l *Library) Clear() error {
	shots, err := l.List()
	if err != nil {
		return err
	}
	var errs []error
	for _, s := range shots {
		if err := os.Remove(s.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Size is how much space the screenshots take.
func Size(shots []Shot) int64 {
	var n int64
	for _, s := range shots {
		n += s.Size
	}
	return n
}
