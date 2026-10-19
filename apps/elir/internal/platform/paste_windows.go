package platform

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
)

type Paster struct{}

var (
	readClipboard  = mygo.Clipboard.Read
	writeClipboard = mygo.Clipboard.Write
	injectControl  = pressControl
)

func (Paster) Paste(text string) error {
	saved, err := readClipboard()
	if err != nil {
		return err
	}
	keep := len(saved.Items()) > 0
	// This temporary transport must not become a Windows history/cloud entry.
	data := transfer.New(transfer.NewItem(
		transfer.Bytes(transfer.Text, []byte(text)),
		transfer.Bytes("ExcludeClipboardContentFromMonitorProcessing", []byte{0}),
	))
	if err := writeClipboard(data); err != nil {
		return err
	}
	if err := injectControl('V'); err != nil {
		return err
	}
	time.AfterFunc(500*time.Millisecond, func() {
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

func CopySelection() error { return pressControl('C') }

type MicStatus int

const (
	MicNotDetermined MicStatus = iota
	MicRestricted
	MicDenied
	MicAuthorized
)

// Windows desktop apps have no per-app Accessibility/Screen Recording grant.
func AccessibilityTrusted(bool) bool { return true }
func ScreenCaptureAllowed() bool     { return true }
func RequestScreenCapture() bool     { return true }
func OpenAccessibilitySettings()     { _ = mygo.Shell.OpenExternal("ms-settings:easeofaccess-keyboard") }
func OpenScreenRecordingSettings()   { _ = mygo.Shell.OpenExternal("ms-settings:privacy-screenshots") }
func OpenMicrophoneSettings()        { _ = mygo.Shell.OpenExternal("ms-settings:privacy-microphone") }
func OpenKeyboardSettings()          { OpenAccessibilitySettings() }

// Permission is established by opening the capture device, not a fabricated
// macOS authorization result. The recorder publishes a denied result on failure.
func Microphone() MicStatus { return MicStatus(microphoneStatus.Load()) }
