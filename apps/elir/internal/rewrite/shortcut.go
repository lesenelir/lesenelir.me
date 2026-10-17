package rewrite

// macOS virtual key codes (Carbon kVK_*) and CGEventFlags bits.
const (
	keyR        = 15 // kVK_ANSI_R
	keyT        = 17 // kVK_ANSI_T
	keyEscape   = 53
	flagShift   = 0x20000
	flagControl = 0x40000
	flagOption  = 0x80000
	flagCommand = 0x100000
)

// shortcuts takes ⌥T and ⌥R from the keyboard of every app, and Esc while
// a rewrite runs. Its methods run on the event tap's thread and must return
// quickly: fire and cancel start or stop work elsewhere.
type shortcuts struct {
	fire   func(Mode)
	cancel func()
	busy   func() bool

	down map[int64]bool // keys whose key down was ours, so their key up is too
}

func (s *shortcuts) FlagsChanged(int64, uint64) bool { return false }

func (s *shortcuts) KeyDown(keycode int64, flags uint64, repeat bool) bool {
	mods := flags & (flagShift | flagControl | flagOption | flagCommand)
	switch {
	case keycode == keyEscape && mods == 0 && s.busy != nil && s.busy():
		s.take(keycode)
		if !repeat && s.cancel != nil {
			s.cancel()
		}
		return true
	case (keycode == keyT || keycode == keyR) && mods == flagOption:
		s.take(keycode)
		if !repeat && s.fire != nil {
			mode := Translate
			if keycode == keyR {
				mode = Polish
			}
			s.fire(mode)
		}
		return true
	}
	return false
}

func (s *shortcuts) KeyUp(keycode int64) bool {
	if s.down[keycode] {
		delete(s.down, keycode)
		return true
	}
	return false
}

func (s *shortcuts) take(keycode int64) {
	if s.down == nil {
		s.down = map[int64]bool{}
	}
	s.down[keycode] = true
}
