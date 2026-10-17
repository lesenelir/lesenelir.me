package hotkey

import "testing"

const (
	shiftFlags = 0x20000 | FlagLeftShift
	cmdFlags   = 0x100000
)

func TestFnTapDictates(t *testing.T) {
	var d Detector
	if a := d.FlagsChanged(KeyFn, FlagFn); a != None {
		t.Fatalf("Fn down = %v, want none", a)
	}
	if a := d.FlagsChanged(KeyFn, 0); a != Dictate {
		t.Fatalf("Fn up = %v, want dictate", a)
	}
}

func TestGlobeKeyDictates(t *testing.T) {
	var d Detector
	d.FlagsChanged(KeyGlobe, FlagFn)
	if a := d.FlagsChanged(KeyGlobe, 0); a != Dictate {
		t.Fatalf("Globe up = %v, want dictate", a)
	}
}

func TestFnWithLeftShiftTranslates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		steps func(d *Detector) Action
	}{
		{"shift first", func(d *Detector) Action {
			d.FlagsChanged(KeyLeftShift, shiftFlags)
			d.FlagsChanged(KeyFn, shiftFlags|FlagFn)
			d.FlagsChanged(KeyLeftShift, FlagFn)
			return d.FlagsChanged(KeyFn, 0)
		}},
		{"fn first", func(d *Detector) Action {
			d.FlagsChanged(KeyFn, FlagFn)
			d.FlagsChanged(KeyLeftShift, shiftFlags|FlagFn)
			return d.FlagsChanged(KeyFn, shiftFlags)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var d Detector
			if a := tc.steps(&d); a != Translate {
				t.Fatalf("got %v, want translate", a)
			}
		})
	}
}

func TestRightShiftWithFnIsIgnored(t *testing.T) {
	var d Detector
	d.FlagsChanged(KeyFn, FlagFn)
	d.FlagsChanged(60, 0x20000|0x4|FlagFn) // Right Shift
	if a := d.FlagsChanged(KeyFn, 0x20000|0x4); a != None {
		t.Fatalf("Fn + Right Shift = %v, want none", a)
	}
}

func TestFnShortcutsAreIgnored(t *testing.T) {
	var d Detector
	d.FlagsChanged(KeyFn, FlagFn)
	if a, swallow := d.KeyDown(123, false); a != None || swallow { // ←
		t.Fatalf("Fn+← down = %v, %v", a, swallow)
	}
	if a := d.FlagsChanged(KeyFn, 0); a != None {
		t.Fatalf("Fn up after Fn+← = %v, want none", a)
	}

	d.FlagsChanged(KeyFn, FlagFn)
	d.FlagsChanged(55, cmdFlags|FlagFn) // Command
	if a := d.FlagsChanged(KeyFn, cmdFlags); a != None {
		t.Fatalf("Fn up after Fn+Cmd = %v, want none", a)
	}

	// The next clean tap works again.
	d.FlagsChanged(55, 0)
	d.FlagsChanged(KeyFn, FlagFn)
	if a := d.FlagsChanged(KeyFn, 0); a != Dictate {
		t.Fatalf("clean tap = %v, want dictate", a)
	}
}

func TestEscCancelsOnlyWhileCaptured(t *testing.T) {
	var d Detector
	if a, swallow := d.KeyDown(KeyEscape, false); a != None || swallow {
		t.Fatalf("Esc without capture = %v, %v", a, swallow)
	}
	if d.KeyUp(KeyEscape) {
		t.Fatal("Esc up without capture was swallowed")
	}

	d.Capture(true, false)
	if a, swallow := d.KeyDown(KeyEscape, false); a != Cancel || !swallow {
		t.Fatalf("captured Esc = %v, %v", a, swallow)
	}
	d.Capture(false, false)
	if !d.KeyUp(KeyEscape) {
		t.Fatal("the key up of a swallowed Esc reached the app")
	}
}

func TestTabCyclesOnlyWhileTranslating(t *testing.T) {
	var d Detector
	d.Capture(true, false)
	if a, swallow := d.KeyDown(KeyTab, false); a != None || swallow {
		t.Fatalf("Tab while dictating = %v, %v", a, swallow)
	}
	d.Capture(true, true)
	if a, swallow := d.KeyDown(KeyTab, false); a != CycleTarget || !swallow {
		t.Fatalf("Tab while translating = %v, %v", a, swallow)
	}
	if a, swallow := d.KeyDown(KeyTab, true); a != None || !swallow {
		t.Fatalf("repeated Tab = %v, %v", a, swallow)
	}
}
