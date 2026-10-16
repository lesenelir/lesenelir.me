//go:build darwin

package darwin

import (
	"testing"
	"unsafe"
)

// The frameworks load and every function and constant resolves. Nothing
// here asks the user for a permission.
func TestFrameworksLoad(t *testing.T) {
	if err := load(); err != nil {
		t.Fatal(err)
	}
	for name, v := range map[string]uintptr{
		"kCFRunLoopCommonModes":           kCFRunLoopCommonModes,
		"kCFBooleanTrue":                  kCFBooleanTrue,
		"kCFTypeDictionaryKeyCallBacks":   kCFTypeDictionaryKeyCallBacks,
		"kCFTypeDictionaryValueCallBacks": kCFTypeDictionaryValueCallBacks,
		"kAXTrustedCheckOptionPrompt":     kAXTrustedCheckOptionPrompt,
		"AVMediaTypeAudio":                avMediaTypeAudio,
	} {
		if v == 0 {
			t.Errorf("%s is nil", name)
		}
	}
	_ = AccessibilityTrusted(false)
	if s := Microphone(); s < MicNotDetermined || s > MicAuthorized {
		t.Errorf("microphone status %d", s)
	}
}

func TestStructLayouts(t *testing.T) {
	if got := unsafe.Sizeof(streamDescription{}); got != 40 {
		t.Errorf("AudioStreamBasicDescription is %d bytes, want 40", got)
	}
	var b queueBuffer
	for name, got := range map[string]uintptr{
		"mAudioData":         unsafe.Offsetof(b.Data),
		"mAudioDataByteSize": unsafe.Offsetof(b.Size),
		"mUserData":          unsafe.Offsetof(b.UserData),
	} {
		want := map[string]uintptr{"mAudioData": 8, "mAudioDataByteSize": 16, "mUserData": 24}[name]
		if got != want {
			t.Errorf("%s at %d, want %d", name, got, want)
		}
	}
}

func TestDictionaryCreate(t *testing.T) {
	if err := load(); err != nil {
		t.Fatal(err)
	}
	keys := []uintptr{kAXTrustedCheckOptionPrompt}
	values := []uintptr{kCFBooleanTrue}
	d := cfDictionaryCreate(0, &keys[0], &values[0], 1, kCFTypeDictionaryKeyCallBacks, kCFTypeDictionaryValueCallBacks)
	if d == 0 {
		t.Fatal("CFDictionaryCreate returned NULL")
	}
	cfRelease(d)
}
