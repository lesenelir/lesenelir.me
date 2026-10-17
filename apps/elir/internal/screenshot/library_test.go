package screenshot

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLibrary(t *testing.T) {
	l, err := OpenLibrary(filepath.Join(t.TempDir(), "screenshot"))
	if err != nil {
		t.Fatal(err)
	}
	if shots, err := l.List(); err != nil || len(shots) != 0 {
		t.Fatalf("empty library: %v, %v", shots, err)
	}
	now := time.Date(2026, 10, 17, 14, 32, 5, 0, time.Local)
	add := func(age time.Duration) string {
		path := l.NewPath(now.Add(-age))
		if err := os.WriteFile(path, []byte("png"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, now.Add(-age), now.Add(-age)); err != nil {
			t.Fatal(err)
		}
		return path
	}
	newest := add(0)
	if filepath.Base(newest) != "Screenshot 2026-10-17 at 14.32.05.png" {
		t.Fatalf("named %q", filepath.Base(newest))
	}
	if second := add(0); filepath.Base(second) != "Screenshot 2026-10-17 at 14.32.05 (2).png" {
		t.Fatalf("the second of the same second named %q", filepath.Base(second))
	}
	old := add(8 * 24 * time.Hour)
	_ = os.WriteFile(filepath.Join(l.dir, "notes.txt"), nil, 0o644)

	shots, err := l.List()
	if err != nil || len(shots) != 3 || shots[2].Path != old || Size(shots) != 9 {
		t.Fatalf("list %+v, %v", shots, err)
	}
	if n, err := l.Prune(0, now); n != 0 || err != nil {
		t.Fatalf("kept forever, pruned %d, %v", n, err)
	}
	if n, err := l.Prune(7*24*time.Hour, now); n != 1 || err != nil {
		t.Fatalf("pruned %d, %v", n, err)
	}

	shots, _ = l.List()
	if err := l.Delete(shots[0].Path); err != nil {
		t.Fatal(err)
	}
	if err := l.Restore(shots[0], []byte("png")); err != nil {
		t.Fatal(err)
	}
	if again, _ := l.List(); len(again) != 2 || !again[0].Time.Equal(shots[0].Time) {
		t.Fatalf("after restore %+v", again)
	}
	if err := l.Delete("/etc/hosts"); err == nil {
		t.Fatal("deleted a file outside the library")
	}
	if err := l.Clear(); err != nil {
		t.Fatal(err)
	}
	if shots, _ := l.List(); len(shots) != 0 {
		t.Fatalf("after clear %+v", shots)
	}
	if _, err := os.Stat(filepath.Join(l.dir, "notes.txt")); err != nil {
		t.Fatal("clear deleted a file that is not a screenshot")
	}
}

func TestShrink(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 400, 100))
	for x := range 400 {
		for y := range 100 {
			c := color.NRGBA{255, 255, 255, 255}
			if x%2 == 0 {
				c = color.NRGBA{0, 0, 0, 255}
			}
			src.SetNRGBA(x, y, c)
		}
	}
	got := Shrink(src, 200, 200)
	if b := got.Bounds(); b.Dx() != 200 || b.Dy() != 50 {
		t.Fatalf("shrunk to %v", b)
	}
	// Black and white columns average to grey.
	if c := got.(*image.RGBA).RGBAAt(10, 10); c.R < 120 || c.R > 135 || c.A != 255 {
		t.Fatalf("pixel %v", c)
	}
	if small := image.NewRGBA(image.Rect(0, 0, 10, 10)); Shrink(small, 200, 200) != image.Image(small) {
		t.Fatal("a small image was copied")
	}
}
