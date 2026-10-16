//go:build darwin

package darwin

import (
	"errors"
	"sync"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/voice/internal/hotkey"
)

const (
	cgSessionEventTap       = 1
	cgHeadInsertEventTap    = 0
	cgEventTapOptionDefault = 0

	cgEventKeyDown      = 10
	cgEventKeyUp        = 11
	cgEventFlagsChanged = 12
	// The system disables a tap whose callback is slow, or on secure
	// input, and tells it with these event types.
	cgEventTapDisabledByTimeout   = 0xFFFFFFFE
	cgEventTapDisabledByUserInput = 0xFFFFFFFF

	fieldKeyboardEventAutorepeat = 8
	fieldKeyboardEventKeycode    = 9
	fieldEventSourceUserData     = 42

	// syntheticTag marks the events this app posts, so its tap lets them by.
	syntheticTag = 0x766F696365 // "voice"
)

// ErrNotTrusted means the app may not watch the keyboard or press keys
// yet: the user has to allow it under Privacy & Security → Accessibility.
var ErrNotTrusted = errors.New("darwin: accessibility permission needed")

// EventTap watches the keyboard of every app for the hotkeys. There is one
// per process; its callback runs on the main thread.
type EventTap struct {
	Detector hotkey.Detector
	// OnAction receives the detector's actions on the main thread; it must
	// return quickly.
	OnAction func(hotkey.Action)

	tap    uintptr
	source uintptr
}

var (
	tapCallbackOnce sync.Once
	tapCallback     uintptr
	// activeTap is the tap the callback serves; only the main thread
	// touches it.
	activeTap *EventTap
)

// Start installs the tap. It fails with ErrNotTrusted until the user
// allows Accessibility, after which Start can be called again.
func (t *EventTap) Start() error {
	if err := load(); err != nil {
		return err
	}
	tapCallbackOnce.Do(func() { tapCallback = purego.NewCallback(onTapEvent) })
	var err error
	mygo.RunOnMain(func() {
		if t.tap != 0 {
			return
		}
		mask := uint64(1)<<cgEventKeyDown | uint64(1)<<cgEventKeyUp | uint64(1)<<cgEventFlagsChanged
		tap := cgEventTapCreate(cgSessionEventTap, cgHeadInsertEventTap, cgEventTapOptionDefault, mask, tapCallback, 0)
		if tap == 0 {
			err = ErrNotTrusted
			return
		}
		t.tap = tap
		t.source = cfMachPortCreateRunLoopSource(0, tap, 0)
		cfRunLoopAddSource(cfRunLoopGetMain(), t.source, kCFRunLoopCommonModes)
		cgEventTapEnable(tap, true)
		activeTap = t
	})
	return err
}

// Running reports whether the tap is installed.
func (t *EventTap) Running() bool {
	running := false
	mygo.RunOnMain(func() { running = t.tap != 0 })
	return running
}

// Stop removes the tap.
func (t *EventTap) Stop() {
	mygo.RunOnMain(func() {
		if t.tap == 0 {
			return
		}
		cgEventTapEnable(t.tap, false)
		cfRunLoopRemoveSource(cfRunLoopGetMain(), t.source, kCFRunLoopCommonModes)
		cfMachPortInvalidate(t.tap)
		cfRelease(t.source)
		cfRelease(t.tap)
		t.tap, t.source = 0, 0
		if activeTap == t {
			activeTap = nil
		}
	})
}

// onTapEvent is the CGEventTapCallBack. Returning 0 swallows the event.
func onTapEvent(proxy, eventType, event, info uintptr) uintptr {
	t := activeTap
	if t == nil {
		return event
	}
	switch typ := uint32(eventType); typ {
	case cgEventTapDisabledByTimeout, cgEventTapDisabledByUserInput:
		cgEventTapEnable(t.tap, true)
		return event
	case cgEventKeyDown, cgEventKeyUp, cgEventFlagsChanged:
		if cgEventGetIntegerValueField(event, fieldEventSourceUserData) == syntheticTag {
			return event
		}
		keycode := cgEventGetIntegerValueField(event, fieldKeyboardEventKeycode)
		var action hotkey.Action
		swallow := false
		switch typ {
		case cgEventFlagsChanged:
			action = t.Detector.FlagsChanged(keycode, cgEventGetFlags(event))
		case cgEventKeyDown:
			repeat := cgEventGetIntegerValueField(event, fieldKeyboardEventAutorepeat) != 0
			action, swallow = t.Detector.KeyDown(keycode, repeat)
		case cgEventKeyUp:
			swallow = t.Detector.KeyUp(keycode)
		}
		if action != hotkey.None && t.OnAction != nil {
			t.OnAction(action)
		}
		if swallow {
			return 0
		}
	}
	return event
}
