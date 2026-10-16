//go:build darwin

package darwin

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// audioAddress is AudioObjectPropertyAddress.
type audioAddress struct{ Selector, Scope, Element uint32 }

const (
	audioSystemObject               = 1
	propDefaultOutputDevice         = 0x644F7574 // 'dOut'
	propMute                        = 0x6D757465 // 'mute'
	propVirtualMainVolume           = 0x766D7663 // 'vmvc', AudioHardwareService
	scopeGlobal                     = 0x676C6F62 // 'glob'
	scopeOutput                     = 0x6F757470 // 'outp'
	elementMain                     = 0
	volumeTolerance         float32 = 0.01
)

var (
	coreAudioOnce sync.Once
	coreAudioErr  error

	audioObjectGetPropertyData     func(object uint32, address *audioAddress, qualifierSize uint32, qualifier unsafe.Pointer, size *uint32, data unsafe.Pointer) int32
	audioObjectSetPropertyData     func(object uint32, address *audioAddress, qualifierSize uint32, qualifier unsafe.Pointer, size uint32, data unsafe.Pointer) int32
	audioObjectHasProperty         func(object uint32, address *audioAddress) bool
	audioObjectIsPropertySettable  func(object uint32, address *audioAddress, settable *uint8) int32
	hardwareServiceGetPropertyData func(object uint32, address *audioAddress, qualifierSize uint32, qualifier unsafe.Pointer, size *uint32, data unsafe.Pointer) int32
	hardwareServiceSetPropertyData func(object uint32, address *audioAddress, qualifierSize uint32, qualifier unsafe.Pointer, size uint32, data unsafe.Pointer) int32
)

func loadCoreAudio() error {
	coreAudioOnce.Do(func() {
		if err := load(); err != nil {
			coreAudioErr = err
			return
		}
		ca, err := purego.Dlopen("/System/Library/Frameworks/CoreAudio.framework/CoreAudio", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			coreAudioErr = fmt.Errorf("darwin: open CoreAudio: %w", err)
			return
		}
		at, err := purego.Dlopen("/System/Library/Frameworks/AudioToolbox.framework/AudioToolbox", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			coreAudioErr = fmt.Errorf("darwin: open AudioToolbox: %w", err)
			return
		}
		purego.RegisterLibFunc(&audioObjectGetPropertyData, ca, "AudioObjectGetPropertyData")
		purego.RegisterLibFunc(&audioObjectSetPropertyData, ca, "AudioObjectSetPropertyData")
		purego.RegisterLibFunc(&audioObjectHasProperty, ca, "AudioObjectHasProperty")
		purego.RegisterLibFunc(&audioObjectIsPropertySettable, ca, "AudioObjectIsPropertySettable")
		purego.RegisterLibFunc(&hardwareServiceGetPropertyData, at, "AudioHardwareServiceGetPropertyData")
		purego.RegisterLibFunc(&hardwareServiceSetPropertyData, at, "AudioHardwareServiceSetPropertyData")
	})
	return coreAudioErr
}

// OutputState is how to undo QuietOutput: unmute the device, or put its
// volume back. A zero Kind means there is nothing to undo. It is plain
// data, so that it can be saved and undone after a crash.
type OutputState struct {
	Kind   string  `json:"kind"` // "mute", "volume" or ""
	Device uint32  `json:"device"`
	Volume float32 `json:"volume"` // the volume to put back
	Set    float32 `json:"set"`    // the volume QuietOutput set
}

// QuietOutput mutes the default output device, with level 0, or lowers its
// volume to level times what it is. It returns how to undo it. A device
// that cannot be muted has its volume set to 0 instead.
func QuietOutput(level float32) (OutputState, error) {
	if err := loadCoreAudio(); err != nil {
		return OutputState{}, err
	}
	device, err := defaultOutputDevice()
	if err != nil {
		return OutputState{}, err
	}
	if level <= 0 && muteSettable(device) {
		muted, err := getMute(device)
		if err != nil {
			return OutputState{}, err
		}
		if muted {
			return OutputState{}, nil // already silent: nothing to undo
		}
		if err := setMute(device, true); err != nil {
			return OutputState{}, err
		}
		return OutputState{Kind: "mute", Device: device}, nil
	}
	volume, err := getVolume(device)
	if err != nil {
		return OutputState{}, err
	}
	target := volume * max(0, level)
	if volume <= target+volumeTolerance {
		return OutputState{}, nil // already as quiet
	}
	if err := setVolume(device, target); err != nil {
		return OutputState{}, err
	}
	return OutputState{Kind: "volume", Device: device, Volume: volume, Set: target}, nil
}

// RestoreOutput undoes QuietOutput, unless the user has changed the mute or
// the volume since: then theirs stands.
func RestoreOutput(s OutputState) error {
	if s.Kind == "" {
		return nil
	}
	if err := loadCoreAudio(); err != nil {
		return err
	}
	switch s.Kind {
	case "mute":
		if muted, err := getMute(s.Device); err != nil || !muted {
			return err
		}
		return setMute(s.Device, false)
	case "volume":
		v, err := getVolume(s.Device)
		if err != nil || math.Abs(float64(v-s.Set)) > float64(volumeTolerance) {
			return err
		}
		return setVolume(s.Device, s.Volume)
	}
	return fmt.Errorf("darwin: unknown output state %q", s.Kind)
}

func defaultOutputDevice() (uint32, error) {
	var device uint32
	size := uint32(unsafe.Sizeof(device))
	addr := audioAddress{propDefaultOutputDevice, scopeGlobal, elementMain}
	if st := audioObjectGetPropertyData(audioSystemObject, &addr, 0, nil, &size, unsafe.Pointer(&device)); st != 0 {
		return 0, osStatus("default output device", st)
	}
	if device == 0 {
		return 0, fmt.Errorf("darwin: no default output device")
	}
	return device, nil
}

func muteAddress() *audioAddress { return &audioAddress{propMute, scopeOutput, elementMain} }

func muteSettable(device uint32) bool {
	if !audioObjectHasProperty(device, muteAddress()) {
		return false
	}
	var settable uint8
	return audioObjectIsPropertySettable(device, muteAddress(), &settable) == 0 && settable != 0
}

func getMute(device uint32) (bool, error) {
	var muted uint32
	size := uint32(unsafe.Sizeof(muted))
	if st := audioObjectGetPropertyData(device, muteAddress(), 0, nil, &size, unsafe.Pointer(&muted)); st != 0 {
		return false, osStatus("get mute", st)
	}
	return muted != 0, nil
}

func setMute(device uint32, mute bool) error {
	var v uint32
	if mute {
		v = 1
	}
	if st := audioObjectSetPropertyData(device, muteAddress(), 0, nil, uint32(unsafe.Sizeof(v)), unsafe.Pointer(&v)); st != 0 {
		return osStatus("set mute", st)
	}
	return nil
}

func volumeAddress() *audioAddress {
	return &audioAddress{propVirtualMainVolume, scopeOutput, elementMain}
}

func getVolume(device uint32) (float32, error) {
	var v float32
	size := uint32(unsafe.Sizeof(v))
	if st := hardwareServiceGetPropertyData(device, volumeAddress(), 0, nil, &size, unsafe.Pointer(&v)); st != 0 {
		return 0, osStatus("get volume", st)
	}
	return v, nil
}

func setVolume(device uint32, v float32) error {
	v = min(1, max(0, v))
	if st := hardwareServiceSetPropertyData(device, volumeAddress(), 0, nil, uint32(unsafe.Sizeof(v)), unsafe.Pointer(&v)); st != 0 {
		return osStatus("set volume", st)
	}
	return nil
}

// OutputStatus reads the default output device's mute and volume, without
// changing them.
func OutputStatus() (device uint32, muted bool, volume float32, err error) {
	if err = loadCoreAudio(); err != nil {
		return
	}
	if device, err = defaultOutputDevice(); err != nil {
		return
	}
	if muteSettable(device) {
		if muted, err = getMute(device); err != nil {
			return
		}
	}
	volume, err = getVolume(device)
	return
}
