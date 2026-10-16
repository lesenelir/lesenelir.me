//go:build darwin

package darwin

import (
	"time"

	"github.com/egoist/mygo"
)

const (
	cgHIDEventTap               = 0
	cgEventSourceStateHIDSystem = 1
	cgEventFlagMaskCommand      = 0x100000
	keyV                        = 9 // kVK_ANSI_V
)

// restoreDelay is how long the pasted text stays on the clipboard before
// what was there comes back: the app in front reads it when it handles
// Cmd+V, which slow apps do some time after the key arrives.
const restoreDelay = 500 * time.Millisecond

// Paster puts text at the cursor of the app in front: it places the text on
// the clipboard, presses Cmd+V and puts the clipboard back as it was.
type Paster struct{}

// Paste needs the Accessibility permission to press Cmd+V. Without it the
// text stays on the clipboard for the user to paste.
func (Paster) Paste(text string) error {
	if err := load(); err != nil {
		return err
	}
	saved, err := mygo.Clipboard.Read()
	keep := err == nil && len(saved.Items()) > 0
	mygo.Clipboard.WriteText(text)
	if !axIsProcessTrusted() {
		return ErrNotTrusted
	}
	pressCmdV()
	time.AfterFunc(restoreDelay, func() {
		// The user may have copied something meanwhile; leave it be.
		if mygo.Clipboard.ReadText() != text {
			return
		}
		if keep {
			_ = mygo.Clipboard.Write(saved)
		} else {
			mygo.Clipboard.Clear()
		}
	})
	return nil
}

func pressCmdV() {
	source := cgEventSourceCreate(cgEventSourceStateHIDSystem)
	if source != 0 {
		defer cfRelease(source)
	}
	for _, down := range []bool{true, false} {
		ev := cgEventCreateKeyboardEvent(source, keyV, down)
		if ev == 0 {
			continue
		}
		// Only Command: a Shift still held from Fn + Left Shift must not
		// turn this into Cmd+Shift+V.
		cgEventSetFlags(ev, cgEventFlagMaskCommand)
		cgEventSetIntegerValueField(ev, fieldEventSourceUserData, syntheticTag)
		cgEventPost(cgHIDEventTap, ev)
		cfRelease(ev)
	}
}
