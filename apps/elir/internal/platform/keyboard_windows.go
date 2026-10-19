package platform

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/egoist/mygo"
)

var user32 = systemDLL("user32.dll")
var (
	setHook       = user32.MustFindProc("SetWindowsHookExW")
	unhook        = user32.MustFindProc("UnhookWindowsHookEx")
	nextHook      = user32.MustFindProc("CallNextHookEx")
	asyncKey      = user32.MustFindProc("GetAsyncKeyState")
	sendInput     = user32.MustFindProc("SendInput")
	getMessage    = user32.MustFindProc("GetMessageW")
	postThreadMsg = user32.MustFindProc("PostThreadMessageW")
	currentThread = kernel32.NewProc("GetCurrentThreadId")
)

var ErrNotTrusted = errors.New("Windows blocked keyboard input; use an app with the same privilege level")

// maskKey is an unassigned virtual key. Typed between an Alt or Win press and
// its release, it keeps the app in front from taking a lone Alt for its menu
// bar, or a lone Win for the Start menu, when the key in between was ours.
const maskKey = 0xe8

// keyBudget is how long the hook waits for the main thread to handle a key.
// Every key of the session waits on the hook meanwhile, and Windows removes a
// hook that overruns LowLevelHooksTimeout without telling it; a key handled
// late reaches the app in front.
const keyBudget = 200 * time.Millisecond

// KeyHandler receives the normalized key codes used by the existing detectors.
// Calls run on MyGo's main thread.
type KeyHandler interface {
	FlagsChanged(int64, uint64) bool
	KeyDown(int64, uint64, bool) bool
	KeyUp(int64) bool
}

// EventTap is a low-level keyboard hook. The hook lives on a thread of its
// own, which does nothing but wait for keys, so that a busy main thread
// neither stalls typing in every app nor gets the hook removed.
type EventTap struct {
	Handler KeyHandler

	mu       sync.Mutex
	thread   uintptr       // the hook thread's ID, 0 when stopped
	stopped  chan struct{} // closed when the hook thread ends
	callback uintptr
	keys     chan func() // keys on their way to the main thread, in order

	// Main thread only.
	down map[uint32]bool
	fn   bool
}

func (t *EventTap) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running() {
		return nil
	}
	mygo.RunOnMain(func() {
		t.down = make(map[uint32]bool)
		t.fn = false
		for _, k := range []uint32{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0x5b, 0x5c} {
			v, _, _ := asyncKey.Call(uintptr(k))
			t.down[k] = v&0x8000 != 0
		}
	})
	if t.callback == 0 {
		t.callback = syscall.NewCallback(t.handle)
	}
	if t.keys == nil {
		t.keys = make(chan func(), 256)
		go func() {
			for k := range t.keys {
				mygo.RunOnMain(k)
			}
		}()
	}
	started := make(chan error, 1)
	t.stopped = make(chan struct{})
	go t.run(started)
	return <-started
}

// run installs the hook and waits for keys until Stop.
func (t *EventTap) run(started chan<- error) {
	// The thread is never unlocked, so it ends with this goroutine.
	runtime.LockOSThread()
	defer close(t.stopped)
	hook, _, err := setHook.Call(13, t.callback, 0, 0) // WH_KEYBOARD_LL
	if hook == 0 {
		started <- fmt.Errorf("keyboard hook: %w", err)
		return
	}
	defer unhook.Call(hook)
	t.thread, _, _ = currentThread.Call()
	started <- nil
	var msg [7]uintptr // MSG
	for {
		if r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg[0])), 0, 0, 0); int32(r) <= 0 {
			return
		}
	}
}

func (t *EventTap) running() bool {
	if t.thread == 0 {
		return false
	}
	select {
	case <-t.stopped:
		return false
	default:
		return true
	}
}

func (t *EventTap) Running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running()
}

func (t *EventTap) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.thread == 0 {
		return
	}
	postThreadMsg.Call(t.thread, 0x12, 0, 0) // WM_QUIT
	<-t.stopped
	t.thread = 0
}

// handle runs on the hook thread.
func (t *EventTap) handle(code, message, data uintptr) uintptr {
	if int32(code) >= 0 && t.Handler != nil {
		e := (*struct {
			Key, Scan, Flags, Time uint32
			Extra                  uintptr
		})(native(data))
		// Ignore injected events, including our own Ctrl+C / Ctrl+V.
		if e.Flags&0x10 == 0 {
			down := message == 0x100 || message == 0x104
			if t.forward(e.Key, down) {
				return 1
			}
		}
	}
	r, _, _ := nextHook.Call(0, code, message, data)
	return r
}

// forward has the main thread handle a key, and reports whether to swallow
// it. A key the main thread is too busy for passes to the app in front; the
// handlers still see it, in order, so that they know which keys are down.
func (t *EventTap) forward(vk uint32, down bool) bool {
	type result struct{ swallow, mask bool }
	reply := make(chan result, 1)
	select {
	case t.keys <- func() {
		swallow := t.key(vk, down)
		reply <- result{swallow, swallow && down && t.flags()&(0x80000|0x100000) != 0}
	}:
	default:
		return false
	}
	select {
	case r := <-reply:
		if r.mask {
			// The app has seen Alt or Win go down but not this key.
			sendKeys(maskKeys())
		}
		return r.swallow
	case <-time.After(keyBudget):
		return false
	}
}

func (t *EventTap) key(vk uint32, down bool) bool {
	repeat := t.down[vk]
	t.down[vk] = down
	flags := t.flags()
	if vk == 0x77 { // F8 replaces the Mac's Fn key.
		if down && !repeat && flags&(0x40000|0x80000|0x100000|4) == 0 {
			t.fn = true
		}
		if t.fn {
			if down {
				flags |= 0x800000
			}
			t.Handler.FlagsChanged(63, flags)
			if !down {
				t.fn = false
			}
			return true
		}
	}
	if t.fn {
		flags |= 0x800000
	}
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
	if t.down[0xa2] || t.down[0xa3] {
		f |= 0x40000
	}
	if t.down[0xa4] || t.down[0xa5] {
		f |= 0x80000
	}
	if t.down[0x5b] || t.down[0x5c] {
		f |= 0x100000
	}
	return f
}

func modifierKey(vk uint32) (int64, bool) {
	switch vk {
	case 0xa0:
		return 56, true
	case 0xa1:
		return 60, true
	case 0xa2, 0xa3:
		return 59, true
	case 0xa4, 0xa5:
		return 58, true
	case 0x5b, 0x5c:
		return 55, true
	}
	return 0, false
}

func normalizedKey(vk uint32) int64 {
	switch vk {
	case 0x1b:
		return 53
	case 9:
		return 48
	case 'T':
		return 17
	case 'R':
		return 15
	case '2':
		return 19
	}
	return int64(vk) + 256
}

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

// maskKeys press and release maskKey.
func maskKeys() []keyboardInput {
	return []keyboardInput{{Type: 1, Key: maskKey}, {Type: 1, Key: maskKey, Flags: 2}}
}

func sendKeys(events []keyboardInput) bool {
	n, _, _ := sendInput.Call(uintptr(len(events)), uintptr(unsafe.Pointer(&events[0])), unsafe.Sizeof(events[0]))
	return n == uintptr(len(events))
}

func pressControl(key uint16) error {
	// Release held modifiers in the injected stream so Alt+T cannot become
	// Alt+Ctrl+C. Restore them after the copy/paste chord. The mask keys keep
	// either Alt release, ours or the user's, from opening the app's menu.
	mask := maskKeys()
	var events []keyboardInput
	var held []uint16
	for _, k := range []uint16{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0x5b, 0x5c} {
		v, _, _ := asyncKey.Call(uintptr(k))
		if v&0x8000 != 0 {
			held = append(held, k)
		}
	}
	if len(held) > 0 {
		events = append(events, mask...)
	}
	for _, k := range held {
		events = append(events, keyboardInput{Type: 1, Key: k, Flags: 2})
	}
	events = append(events, keyboardInput{Type: 1, Key: 0x11}, keyboardInput{Type: 1, Key: key}, keyboardInput{Type: 1, Key: key, Flags: 2}, keyboardInput{Type: 1, Key: 0x11, Flags: 2})
	for _, k := range held {
		events = append(events, keyboardInput{Type: 1, Key: k})
	}
	if len(held) > 0 {
		events = append(events, mask...)
	}
	if !sendKeys(events) {
		return ErrNotTrusted
	}
	return nil
}
