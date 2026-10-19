package screenshot

import (
	"runtime"
	"testing"
)

func TestShortcut(t *testing.T) {
	command, other := uint64(flagCommand), uint64(flagControl)
	if runtime.GOOS == "windows" {
		command, other = flagControl, flagCommand
	}
	fired := 0
	s := &shortcut{fire: func() { fired++ }}
	for _, flags := range []uint64{command, flagShift, command | flagShift | flagOption, command | flagShift | other} {
		if s.KeyDown(key2, flags, false) {
			t.Errorf("took 2 with flags %#x", flags)
		}
	}
	if s.KeyUp(key2) {
		t.Error("took a key up of 2 it did not take down")
	}
	// Caps Lock and the other bits of the flags do not matter.
	if !s.KeyDown(key2, command|flagShift|0x10000|0x100, false) || !s.KeyDown(key2, command|flagShift, true) {
		t.Fatal("let ⌘⇧2 through")
	}
	if !s.KeyUp(key2) || s.KeyUp(key2) {
		t.Fatal("key up of ⌘⇧2")
	}
	if fired != 1 {
		t.Fatalf("fired %d times, want once for a held key", fired)
	}
	if s.KeyDown(20, command|flagShift, false) {
		t.Fatal("took ⌘⇧3")
	}
}
