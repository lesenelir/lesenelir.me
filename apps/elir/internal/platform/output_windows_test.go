package platform

import (
	"runtime"
	"strings"
	"testing"
)

func TestWithOutputPreservesCOMApartment(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode int
	}{
		{"uninitialized", -1}, {"MTA", 0}, {"STA", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			ownsCOM := false
			defer func() {
				if ownsCOM {
					coUninit.Call()
				}
			}()
			if tc.mode >= 0 {
				hr, _, _ := coInit.Call(0, uintptr(tc.mode))
				ownsCOM = int32(hr) >= 0
				if uint32(hr) != 0 {
					t.Fatalf("initial apartment HRESULT %#x", uint32(hr))
				}
			}
			for range 2 {
				// An invalid endpoint reaches COM but never reads or changes an
				// actual audio device. A missing COM class is also harmless here.
				err := withOutput("elir-invalid-test-endpoint", func(uintptr, string) error {
					t.Fatal("invalid endpoint reached volume callback")
					return nil
				})
				if err == nil || !(strings.HasPrefix(err.Error(), "audio endpoint:") || strings.HasPrefix(err.Error(), "MMDeviceEnumerator:")) {
					t.Fatalf("expected endpoint failure after COM initialization, got %v", err)
				}
			}
			mode, want := uintptr(2), uint32(0)
			if ownsCOM {
				mode, want = uintptr(tc.mode), 1
			} // Existing apartment: S_FALSE.
			hr, _, _ := coInit.Call(0, mode)
			if int32(hr) >= 0 {
				coUninit.Call()
			}
			if uint32(hr) != want {
				t.Fatalf("caller's apartment changed: HRESULT %#x, want %#x", uint32(hr), want)
			}
			if ownsCOM {
				coUninit.Call()
				ownsCOM = false
			}
			// Once the caller releases its reference, switching modes must work.
			// A leaked S_FALSE reference would instead yield RPC_E_CHANGED_MODE.
			if mode == 2 {
				mode = 0
			} else {
				mode = 2
			}
			hr, _, _ = coInit.Call(0, mode)
			if int32(hr) >= 0 {
				coUninit.Call()
			}
			if uint32(hr) != 0 {
				t.Fatalf("COM initialization leaked: HRESULT %#x", uint32(hr))
			}
		})
	}
}
