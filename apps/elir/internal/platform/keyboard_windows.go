package platform

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

var user32 = systemDLL("user32.dll")
var (
	setHook   = user32.MustFindProc("SetWindowsHookExW")
	unhook    = user32.MustFindProc("UnhookWindowsHookEx")
	nextHook  = user32.MustFindProc("CallNextHookEx")
	asyncKey  = user32.MustFindProc("GetAsyncKeyState")
	sendInput = user32.MustFindProc("SendInput")
)

var ErrNotTrusted = errors.New("Windows blocked keyboard input; use an app with the same privilege level")

// KeyHandler receives the normalized key codes used by the existing detectors.
// Calls run on MyGo's main thread, where the hook is installed.
type KeyHandler interface {
	FlagsChanged(int64, uint64) bool
	KeyDown(int64, uint64, bool) bool
	KeyUp(int64) bool
}

type EventTap struct {
	Handler  KeyHandler
	hook     uintptr
	callback uintptr
	down     map[uint32]bool
}

func (t *EventTap) Start() error {
	var err error
	mygo.RunOnMain(func() {
		if t.hook != 0 {
			return
		}
		if h, ok := t.Handler.(interface{ Reset() }); ok {
			h.Reset()
		}
		t.down = make(map[uint32]bool)
		for _, k := range []uint32{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0x5b, 0x5c} {
			v, _, _ := asyncKey.Call(uintptr(k))
			t.down[k] = v&0x8000 != 0
		}
		if t.callback == 0 {
			t.callback = syscall.NewCallback(t.handle)
		}
		t.hook, _, err = setHook.Call(13, t.callback, 0, 0) // WH_KEYBOARD_LL
		if t.hook != 0 {
			err = nil
		} else {
			err = fmt.Errorf("keyboard hook: %w", err)
		}
	})
	return err
}

func (t *EventTap) Running() bool {
	var running bool
	mygo.RunOnMain(func() { running = t.hook != 0 })
	return running
}

func (t *EventTap) Stop() {
	mygo.RunOnMain(func() {
		if t.hook != 0 {
			unhook.Call(t.hook)
			t.hook = 0
			if h, ok := t.Handler.(interface{ Reset() }); ok {
				h.Reset()
			}
		}
	})
}

func (t *EventTap) handle(code, message, data uintptr) uintptr {
	if int32(code) >= 0 && t.Handler != nil {
		e := (*struct {
			Key, Scan, Flags, Time uint32
			Extra                  uintptr
		})(native(data))
		// Ignore injected events, including our own Ctrl+C / Ctrl+V.
		if e.Flags&0x10 == 0 {
			down := message == 0x100 || message == 0x104
			vk := shortcut.WindowsVK(e.Key, e.Scan, e.Flags&1 != 0)
			if t.key(vk, down) {
				return 1
			}
		}
	}
	r, _, _ := nextHook.Call(0, code, message, data)
	return r
}

func (t *EventTap) key(vk uint32, down bool) bool {
	repeat := t.down[vk]
	t.down[vk] = down
	flags := t.flags()
	if k, ok := modifierKey(vk); ok {
		return t.Handler.FlagsChanged(k, flags)
	}
	key := normalizedKey(vk)
	if down {
		return t.Handler.KeyDown(key, flags, repeat)
	}
	return t.Handler.KeyUp(key)
}

func (t *EventTap) flags() uint64 {
	var f uint64
	if t.down[0xa0] {
		f |= 0x20000 | 2
	}
	if t.down[0xa1] {
		f |= 0x20000 | 4
	}
	if t.down[0xa2] {
		f |= 0x40000 | 1
	}
	if t.down[0xa3] {
		f |= 0x40000 | 0x2000
	}
	if t.down[0xa4] {
		f |= 0x80000 | 0x20
	}
	if t.down[0xa5] {
		f |= 0x80000 | 0x40
	}
	if t.down[0x5b] {
		f |= 0x100000 | 8
	}
	if t.down[0x5c] {
		f |= 0x100000 | 16
	}
	return f
}

func modifierKey(vk uint32) (int64, bool) {
	switch vk {
	case 0xa0:
		return 56, true
	case 0xa1:
		return 60, true
	case 0xa2:
		return 59, true
	case 0xa3:
		return 62, true
	case 0xa4:
		return 58, true
	case 0xa5:
		return 61, true
	case 0x5b:
		return 55, true
	case 0x5c:
		return 54, true
	}
	return 0, false
}

func normalizedKey(vk uint32) int64 { return shortcut.WindowsKey(vk) }

// INPUT's union has pointer alignment on both supported 64-bit architectures.
type keyboardInput struct {
	Type        uint32
	_           uint32
	Key, Scan   uint16
	Flags, Time uint32
	_           uint32
	Extra       uintptr
	_           [8]byte
}

func pressControl(key uint16) error {
	// Release held modifiers in the injected stream so Alt+T cannot become
	// Alt+Ctrl+C. Restore them after the copy/paste chord.
	var events []keyboardInput
	var held []uint16
	for _, k := range []uint16{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0x5b, 0x5c} {
		v, _, _ := asyncKey.Call(uintptr(k))
		if v&0x8000 != 0 {
			held = append(held, k)
			events = append(events, keyboardInput{Type: 1, Key: k, Flags: 2})
		}
	}
	events = append(events, keyboardInput{Type: 1, Key: 0x11}, keyboardInput{Type: 1, Key: key}, keyboardInput{Type: 1, Key: key, Flags: 2}, keyboardInput{Type: 1, Key: 0x11, Flags: 2})
	for _, k := range held {
		events = append(events, keyboardInput{Type: 1, Key: k})
	}
	n, _, _ := sendInput.Call(uintptr(len(events)), uintptr(unsafe.Pointer(&events[0])), unsafe.Sizeof(events[0]))
	if n != uintptr(len(events)) {
		return ErrNotTrusted
	}
	return nil
}
