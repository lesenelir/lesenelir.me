package screenshot

// macOS virtual key codes (Carbon kVK_*) and CGEventFlags bits.
const (
	key2        = 19 // kVK_ANSI_2: the key of 2 on any layout
	flagShift   = 0x20000
	flagControl = 0x40000
	flagOption  = 0x80000
	flagCommand = 0x100000
)

// shortcut takes ⌘⇧2 from the keyboard of every app for fire. Its methods
// run on the event tap's thread and must return quickly.
type shortcut struct {
	fire func()
	down bool // ⌘⇧2 went down here, so its key up is ours too
}

func (s *shortcut) FlagsChanged(int64, uint64) bool { return false }

func (s *shortcut) KeyDown(keycode int64, flags uint64, repeat bool) bool {
	mods := flags & (flagShift | flagControl | flagOption | flagCommand)
	if keycode != key2 || mods != flagCommand|flagShift {
		return false
	}
	s.down = true
	if !repeat && s.fire != nil {
		s.fire()
	}
	return true
}

func (s *shortcut) KeyUp(keycode int64) bool {
	if keycode == key2 && s.down {
		s.down = false
		return true
	}
	return false
}
