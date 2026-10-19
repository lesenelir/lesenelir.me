package platform

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
)

type guid struct {
	A    uint32
	B, C uint16
	D    [8]byte
}

var (
	ole32         = systemDLL("ole32.dll")
	coInit        = ole32.MustFindProc("CoInitializeEx")
	coUninit      = ole32.MustFindProc("CoUninitialize")
	coCreate      = ole32.MustFindProc("CoCreateInstance")
	coFree        = ole32.MustFindProc("CoTaskMemFree")
	clsEnumerator = guid{0xbcde0395, 0xe52f, 0x467c, [8]byte{0x8e, 0x3d, 0xc4, 0x57, 0x92, 0x91, 0x69, 0x2e}}
	iidEnumerator = guid{0xa95664d2, 0x9614, 0x4f35, [8]byte{0xa7, 0x46, 0xde, 0x8d, 0xb6, 0x36, 0x17, 0xe6}}
	iidVolume     = guid{0x5cdf2c82, 0x841e, 0x4546, [8]byte{0x97, 0x22, 0x0c, 0xf7, 0x40, 0x78, 0x22, 0x9a}}
)

// ErrOutputGone is RestoreOutput's error for an endpoint that is no longer
// there or no longer plugged in: there is nothing left to undo.
var ErrOutputGone = errors.New("the output device is gone")

func native(p uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&p)) }
func method(obj uintptr, index int) uintptr {
	vt := *(*uintptr)(native(obj))
	return *(*uintptr)(unsafe.Add(native(vt), index*int(unsafe.Sizeof(uintptr(0)))))
}

//go:uintptrescapes
func comCall(obj uintptr, index int, args ...uintptr) uintptr {
	a := append([]uintptr{obj}, args...)
	r, _, _ := syscall.SyscallN(method(obj, index), a...)
	return r
}

func comError(op string, hr uintptr) error {
	if int32(hr) >= 0 {
		return nil
	}
	return fmt.Errorf("%s: HRESULT %#08x", op, uint32(hr))
}

type OutputState struct {
	Kind   string  `json:"kind"`
	Device string  `json:"device"`
	Volume float32 `json:"volume"`
	Set    float32 `json:"set"`
}

// withOutput keeps COM initialization and every interface call on one thread.
// An explicit endpoint ID restores the original device even if defaults changed.
func withOutput(id string, fn func(uintptr, string) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := coInit.Call(0, 0)
	if uint32(hr) == 0x80010106 { // RPC_E_CHANGED_MODE: MyGo's main thread is STA.
		hr, _, _ = coInit.Call(0, 2) // COINIT_APARTMENTTHREADED; normally S_FALSE.
	}
	if err := comError("CoInitializeEx", hr); err != nil {
		return err
	}
	defer coUninit.Call()
	var enumerator uintptr
	hr, _, _ = coCreate.Call(uintptr(unsafe.Pointer(&clsEnumerator)), 0, 0x17, uintptr(unsafe.Pointer(&iidEnumerator)), uintptr(unsafe.Pointer(&enumerator)))
	if err := comError("MMDeviceEnumerator", hr); err != nil {
		return err
	}
	defer comCall(enumerator, 2)
	var device uintptr
	if id == "" {
		hr = comCall(enumerator, 4, 0, 0, uintptr(unsafe.Pointer(&device)))
	} else {
		p, err := syscall.UTF16PtrFromString(id)
		if err != nil {
			return err
		}
		hr = comCall(enumerator, 5, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&device)))
		if uint32(hr) == 0x80070490 { // E_NOTFOUND
			return ErrOutputGone
		}
	}
	if err := comError("audio endpoint", hr); err != nil {
		return err
	}
	defer comCall(device, 2)
	if id != "" {
		var state uint32
		if err := comError("endpoint state", comCall(device, 6, uintptr(unsafe.Pointer(&state)))); err != nil {
			return err
		}
		if state != 1 { // DEVICE_STATE_ACTIVE
			return ErrOutputGone
		}
	}
	var p *uint16
	if err := comError("endpoint ID", comCall(device, 5, uintptr(unsafe.Pointer(&p)))); err != nil {
		return err
	}
	defer coFree.Call(uintptr(unsafe.Pointer(p)))
	n := 0
	for *(*uint16)(unsafe.Add(unsafe.Pointer(p), n*2)) != 0 {
		n++
	}
	id = syscall.UTF16ToString(unsafe.Slice(p, n))
	var volume uintptr
	if err := comError("endpoint volume", comCall(device, 3, uintptr(unsafe.Pointer(&iidVolume)), 0x17, 0, uintptr(unsafe.Pointer(&volume)))); err != nil {
		return err
	}
	defer comCall(volume, 2)
	return fn(volume, id)
}

func outputVolume(v uintptr) (float32, error) {
	var level float32
	err := comError("GetMasterVolumeLevelScalar", comCall(v, 9, uintptr(unsafe.Pointer(&level))))
	return level, err
}

func setOutputVolume(v uintptr, level float32) error {
	// purego supplies the correct floating-point ABI on amd64 and arm64.
	var set func(uintptr, float32, uintptr) int32
	purego.RegisterFunc(&set, method(v, 7))
	return comError("SetMasterVolumeLevelScalar", uintptr(set(v, level, 0)))
}

func QuietOutput(level float32) (state OutputState, err error) {
	err = withOutput("", func(v uintptr, id string) error {
		var muted int32
		if err := comError("GetMute", comCall(v, 15, uintptr(unsafe.Pointer(&muted)))); err != nil {
			return err
		}
		if muted != 0 {
			return nil
		}
		if level <= 0 {
			if err := comError("SetMute", comCall(v, 14, 1, 0)); err != nil {
				return err
			}
			state = OutputState{Kind: "mute", Device: id}
			return nil
		}
		volume, err := outputVolume(v)
		if err != nil {
			return err
		}
		target := volume * min(max(level, 0), 1)
		if volume-target < 0.01 {
			return nil
		}
		if err := setOutputVolume(v, target); err != nil {
			return err
		}
		state = OutputState{Kind: "volume", Device: id, Volume: volume, Set: target}
		return nil
	})
	return
}

func RestoreOutput(s OutputState) error {
	if s.Kind == "" {
		return nil
	}
	return withOutput(s.Device, func(v uintptr, _ string) error {
		if s.Kind == "mute" {
			var muted int32
			if err := comError("GetMute", comCall(v, 15, uintptr(unsafe.Pointer(&muted)))); err != nil {
				return err
			}
			if muted == 0 {
				return nil
			}
			return comError("SetMute", comCall(v, 14, 0, 0))
		}
		volume, err := outputVolume(v)
		if err != nil || math.Abs(float64(volume-s.Set)) > 0.01 {
			return err
		}
		return setOutputVolume(v, s.Volume)
	})
}
