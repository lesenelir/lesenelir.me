// Package hotkey turns raw macOS keyboard events into the app's actions: a
// tap of Fn dictates, a tap of Fn with Left Shift translates, Esc cancels and
// Tab cycles the translation target. It holds no platform code, so the rules
// are tested on any machine; the event tap feeds it.
package hotkey

import "sync/atomic"

// Action is what a key event asks the app to do.
type Action int

const (
	None Action = iota
	// Dictate starts dictation, or finishes the current recording.
	Dictate
	// Translate starts translation, or finishes the current recording.
	Translate
	// Cancel drops the current recording or request.
	Cancel
	// CycleTarget moves to the next translation target language.
	CycleTarget
)

func (a Action) String() string {
	switch a {
	case Dictate:
		return "dictate"
	case Translate:
		return "translate"
	case Cancel:
		return "cancel"
	case CycleTarget:
		return "cycle-target"
	}
	return "none"
}

// macOS virtual key codes (Carbon kVK_*) and CGEventFlags bits.
const (
	KeyFn         = 63
	KeyGlobe      = 179
	KeyLeftShift  = 56
	KeyEscape     = 53
	KeyTab        = 48
	FlagFn        = 0x800000
	FlagLeftShift = 0x2 // NX_DEVICELSHIFTKEYMASK
)

// Detector decides, event by event, which action a key press means and
// whether the event should be kept from the app in front. It runs on the
// thread of the event tap; Capture may be called from any goroutine.
type Detector struct {
	fnDown     bool
	lshiftDown bool
	// dirty marks an Fn press joined by another key: Fn+F5, Fn+←, Fn+Cmd…
	// Those belong to the system and must not start a recording.
	dirty bool
	// withShift marks an Fn press joined by Left Shift.
	withShift bool
	// swallowed remembers swallowed key downs, so that their key ups are
	// swallowed too even after capture ends.
	swallowed map[int64]bool

	captureEsc atomic.Bool
	captureTab atomic.Bool
}

// Capture sets which keys the detector takes for itself: Esc while a
// recording or request can be cancelled, Tab while a translation records.
func (d *Detector) Capture(esc, tab bool) {
	d.captureEsc.Store(esc)
	d.captureTab.Store(tab)
}

// FlagsChanged handles a modifier key going down or up.
func (d *Detector) FlagsChanged(keycode int64, flags uint64) Action {
	switch keycode {
	case KeyFn, KeyGlobe:
		down := flags&FlagFn != 0
		switch {
		case down && !d.fnDown:
			d.fnDown = true
			d.dirty = false
			d.withShift = d.lshiftDown
		case !down && d.fnDown:
			d.fnDown = false
			if d.dirty {
				return None
			}
			if d.withShift {
				return Translate
			}
			return Dictate
		}
	case KeyLeftShift:
		d.lshiftDown = flags&FlagLeftShift != 0
		if d.fnDown && d.lshiftDown {
			d.withShift = true
		}
	default:
		// Command, Option, Control or Right Shift with Fn is someone
		// else's shortcut.
		if d.fnDown {
			d.dirty = true
		}
	}
	return None
}

// KeyDown handles a key press. It reports the action and whether to
// swallow the event.
func (d *Detector) KeyDown(keycode int64, repeat bool) (Action, bool) {
	if d.fnDown {
		d.dirty = true
		return None, false
	}
	switch {
	case keycode == KeyEscape && d.captureEsc.Load():
		d.swallow(keycode)
		if repeat {
			return None, true
		}
		return Cancel, true
	case keycode == KeyTab && d.captureTab.Load():
		d.swallow(keycode)
		if repeat {
			return None, true
		}
		return CycleTarget, true
	}
	return None, false
}

// KeyUp reports whether to swallow a key release.
func (d *Detector) KeyUp(keycode int64) bool {
	if d.swallowed[keycode] {
		delete(d.swallowed, keycode)
		return true
	}
	return false
}

func (d *Detector) swallow(keycode int64) {
	if d.swallowed == nil {
		d.swallowed = map[int64]bool{}
	}
	d.swallowed[keycode] = true
}
