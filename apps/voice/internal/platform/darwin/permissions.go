//go:build darwin

package darwin

import (
	"github.com/ebitengine/purego/objc"
	"github.com/egoist/mygo"
)

// MicStatus is AVAuthorizationStatus.
type MicStatus int

const (
	MicNotDetermined MicStatus = iota
	MicRestricted
	MicDenied
	MicAuthorized
)

// AccessibilityTrusted reports whether the app may watch the keyboard and
// press keys. With prompt, macOS shows its dialog pointing the user to the
// setting, once per launch.
func AccessibilityTrusted(prompt bool) bool {
	if load() != nil {
		return false
	}
	if !prompt {
		return axIsProcessTrusted()
	}
	keys := []uintptr{kAXTrustedCheckOptionPrompt}
	values := []uintptr{kCFBooleanTrue}
	opts := cfDictionaryCreate(0, &keys[0], &values[0], 1, kCFTypeDictionaryKeyCallBacks, kCFTypeDictionaryValueCallBacks)
	defer cfRelease(opts)
	return axIsProcessTrustedWithOptions(opts)
}

// Microphone reports the microphone permission.
func Microphone() MicStatus {
	if load() != nil {
		return MicNotDetermined
	}
	device := objc.ID(objc.GetClass("AVCaptureDevice"))
	if device == 0 {
		return MicNotDetermined
	}
	return MicStatus(objc.Send[int](device, objc.RegisterName("authorizationStatusForMediaType:"), objc.ID(avMediaTypeAudio)))
}

// OpenAccessibilitySettings opens Privacy & Security → Accessibility.
func OpenAccessibilitySettings() {
	_ = mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")
}

// OpenMicrophoneSettings opens Privacy & Security → Microphone.
func OpenMicrophoneSettings() {
	_ = mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone")
}

// OpenKeyboardSettings opens Keyboard, where "Press 🌐 key to" should be
// "Do Nothing" so that Fn does not also switch the input source.
func OpenKeyboardSettings() {
	_ = mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.Keyboard-Settings.extension")
}
