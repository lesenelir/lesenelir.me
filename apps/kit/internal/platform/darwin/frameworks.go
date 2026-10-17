//go:build darwin

// Package darwin is the macOS half of the app that MyGo does not cover:
// the global keyboard event tap, microphone capture with AudioQueue, the
// output volume, pasting with a synthesized Cmd+V and the privacy
// permissions behind them. It knows nothing of the tools that use it.
// It calls the system frameworks through purego, without cgo, as MyGo does.
package darwin

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	loadOnce sync.Once
	loadErr  error
)

// CoreFoundation.
var (
	cfRunLoopGetMain              func() uintptr
	cfRunLoopAddSource            func(rl, source, mode uintptr)
	cfRunLoopRemoveSource         func(rl, source, mode uintptr)
	cfMachPortCreateRunLoopSource func(alloc, port uintptr, order int) uintptr
	cfMachPortInvalidate          func(port uintptr)
	cfRelease                     func(ref uintptr)
	cfDictionaryCreate            func(alloc uintptr, keys, values *uintptr, n int, keyCallbacks, valueCallbacks uintptr) uintptr

	kCFRunLoopCommonModes           uintptr
	kCFBooleanTrue                  uintptr
	kCFTypeDictionaryKeyCallBacks   uintptr // address of the struct
	kCFTypeDictionaryValueCallBacks uintptr // address of the struct
)

// CoreGraphics.
var (
	cgEventTapCreate            func(tap, place, options uint32, mask uint64, callback, info uintptr) uintptr
	cgEventTapEnable            func(tap uintptr, enable bool)
	cgEventGetFlags             func(event uintptr) uint64
	cgEventGetIntegerValueField func(event uintptr, field uint32) int64
	cgEventSetIntegerValueField func(event uintptr, field uint32, value int64)
	cgEventCreateKeyboardEvent  func(source uintptr, keycode uint16, down bool) uintptr
	cgEventSetFlags             func(event uintptr, flags uint64)
	cgEventPost                 func(tap uint32, event uintptr)
	cgEventSourceCreate         func(state int32) uintptr

	cgPreflightScreenCaptureAccess func() bool
	cgRequestScreenCaptureAccess   func() bool
)

// ApplicationServices (HIServices).
var (
	axIsProcessTrusted            func() bool
	axIsProcessTrustedWithOptions func(options uintptr) bool
	kAXTrustedCheckOptionPrompt   uintptr
)

// AudioToolbox.
var (
	audioQueueNewInput       func(format *streamDescription, callback, userData, runLoop, runLoopMode uintptr, flags uint32, queue *uintptr) int32
	audioQueueAllocateBuffer func(queue uintptr, size uint32, buffer *uintptr) int32
	audioQueueEnqueueBuffer  func(queue, buffer uintptr, packets uint32, descriptions uintptr) int32
	audioQueueStart          func(queue, startTime uintptr) int32
	audioQueueStop           func(queue uintptr, immediate bool) int32
	audioQueueDispose        func(queue uintptr, immediate bool) int32
)

// AVFoundation, for the microphone permission.
var avMediaTypeAudio uintptr

func load() error {
	loadOnce.Do(func() { loadErr = loadFrameworks() })
	return loadErr
}

func loadFrameworks() error {
	open := func(name string) (uintptr, error) {
		path := "/System/Library/Frameworks/" + name + ".framework/" + name
		h, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return 0, fmt.Errorf("darwin: open %s: %w", name, err)
		}
		return h, nil
	}
	cf, err := open("CoreFoundation")
	if err != nil {
		return err
	}
	cg, err := open("CoreGraphics")
	if err != nil {
		return err
	}
	as, err := open("ApplicationServices")
	if err != nil {
		return err
	}
	at, err := open("AudioToolbox")
	if err != nil {
		return err
	}
	av, err := open("AVFoundation")
	if err != nil {
		return err
	}

	purego.RegisterLibFunc(&cfRunLoopGetMain, cf, "CFRunLoopGetMain")
	purego.RegisterLibFunc(&cfRunLoopAddSource, cf, "CFRunLoopAddSource")
	purego.RegisterLibFunc(&cfRunLoopRemoveSource, cf, "CFRunLoopRemoveSource")
	purego.RegisterLibFunc(&cfMachPortCreateRunLoopSource, cf, "CFMachPortCreateRunLoopSource")
	purego.RegisterLibFunc(&cfMachPortInvalidate, cf, "CFMachPortInvalidate")
	purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
	purego.RegisterLibFunc(&cfDictionaryCreate, cf, "CFDictionaryCreate")

	purego.RegisterLibFunc(&cgEventTapCreate, cg, "CGEventTapCreate")
	purego.RegisterLibFunc(&cgEventTapEnable, cg, "CGEventTapEnable")
	purego.RegisterLibFunc(&cgEventGetFlags, cg, "CGEventGetFlags")
	purego.RegisterLibFunc(&cgEventGetIntegerValueField, cg, "CGEventGetIntegerValueField")
	purego.RegisterLibFunc(&cgEventSetIntegerValueField, cg, "CGEventSetIntegerValueField")
	purego.RegisterLibFunc(&cgEventCreateKeyboardEvent, cg, "CGEventCreateKeyboardEvent")
	purego.RegisterLibFunc(&cgEventSetFlags, cg, "CGEventSetFlags")
	purego.RegisterLibFunc(&cgEventPost, cg, "CGEventPost")
	purego.RegisterLibFunc(&cgEventSourceCreate, cg, "CGEventSourceCreate")
	purego.RegisterLibFunc(&cgPreflightScreenCaptureAccess, cg, "CGPreflightScreenCaptureAccess")
	purego.RegisterLibFunc(&cgRequestScreenCaptureAccess, cg, "CGRequestScreenCaptureAccess")

	purego.RegisterLibFunc(&axIsProcessTrusted, as, "AXIsProcessTrusted")
	purego.RegisterLibFunc(&axIsProcessTrustedWithOptions, as, "AXIsProcessTrustedWithOptions")

	purego.RegisterLibFunc(&audioQueueNewInput, at, "AudioQueueNewInput")
	purego.RegisterLibFunc(&audioQueueAllocateBuffer, at, "AudioQueueAllocateBuffer")
	purego.RegisterLibFunc(&audioQueueEnqueueBuffer, at, "AudioQueueEnqueueBuffer")
	purego.RegisterLibFunc(&audioQueueStart, at, "AudioQueueStart")
	purego.RegisterLibFunc(&audioQueueStop, at, "AudioQueueStop")
	purego.RegisterLibFunc(&audioQueueDispose, at, "AudioQueueDispose")

	vars := []struct {
		lib   uintptr
		name  string
		value *uintptr
		deref bool
	}{
		{cf, "kCFRunLoopCommonModes", &kCFRunLoopCommonModes, true},
		{cf, "kCFBooleanTrue", &kCFBooleanTrue, true},
		{cf, "kCFTypeDictionaryKeyCallBacks", &kCFTypeDictionaryKeyCallBacks, false},
		{cf, "kCFTypeDictionaryValueCallBacks", &kCFTypeDictionaryValueCallBacks, false},
		{as, "kAXTrustedCheckOptionPrompt", &kAXTrustedCheckOptionPrompt, true},
		{av, "AVMediaTypeAudio", &avMediaTypeAudio, true},
	}
	for _, v := range vars {
		addr, err := purego.Dlsym(v.lib, v.name)
		if err != nil {
			return fmt.Errorf("darwin: %s: %w", v.name, err)
		}
		if v.deref {
			addr = readPointer(addr)
		}
		*v.value = addr
	}
	return nil
}

// readPointer reads the pointer stored at a C address.
func readPointer(addr uintptr) uintptr {
	return **(**uintptr)(unsafe.Pointer(&addr))
}

// bytesAt views n bytes of C memory.
func bytesAt(addr uintptr, n int) []byte {
	return unsafe.Slice(*(**byte)(unsafe.Pointer(&addr)), n)
}
