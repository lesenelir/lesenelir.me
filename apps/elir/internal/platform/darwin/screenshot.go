//go:build darwin

package darwin

import (
	"errors"
	"image"
	"math"
	"os"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var windowInfoOnce sync.Once
var windowInfoErr error
var windowList func(uint32, uint32) uintptr
var arrayCount func(uintptr) int64
var arrayValue func(uintptr, int64) uintptr
var dictionaryValue func(uintptr, uintptr) uintptr
var numberValue func(uintptr, int32, *int64) bool
var rectValue func(uintptr, *windowRect) bool
var windowBoundsKey, windowLayerKey, windowPIDKey uintptr

type windowRect struct{ X, Y, W, H float64 }

// ScreenshotWindows snapshots front-to-back window bounds before overlays exist.
func ScreenshotWindows() ([]image.Rectangle, error) {
	if err := load(); err != nil {
		return nil, err
	}
	windowInfoOnce.Do(func() {
		cg, err := purego.Dlopen("/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			windowInfoErr = err
			return
		}
		cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			windowInfoErr = err
			return
		}
		purego.RegisterLibFunc(&windowList, cg, "CGWindowListCopyWindowInfo")
		purego.RegisterLibFunc(&arrayCount, cf, "CFArrayGetCount")
		purego.RegisterLibFunc(&arrayValue, cf, "CFArrayGetValueAtIndex")
		purego.RegisterLibFunc(&dictionaryValue, cf, "CFDictionaryGetValue")
		purego.RegisterLibFunc(&numberValue, cf, "CFNumberGetValue")
		purego.RegisterLibFunc(&rectValue, cg, "CGRectMakeWithDictionaryRepresentation")
		for _, key := range []struct {
			name string
			out  *uintptr
		}{{"kCGWindowBounds", &windowBoundsKey}, {"kCGWindowLayer", &windowLayerKey}, {"kCGWindowOwnerPID", &windowPIDKey}} {
			addr, err := purego.Dlsym(cg, key.name)
			if err != nil {
				windowInfoErr = err
				return
			}
			*key.out = *(*uintptr)(unsafe.Pointer(addr))
		}
	})
	if windowInfoErr != nil {
		return nil, windowInfoErr
	}
	list := windowList(1|16, 0) // On-screen, excluding desktop elements.
	if list == 0 {
		return nil, errors.New("screenshot: window list unavailable")
	}
	defer cfRelease(list)
	var result []image.Rectangle
	for i, n := int64(0), arrayCount(list); i < n; i++ {
		dict := arrayValue(list, i)
		layer, pid := int64(0), int64(0)
		l, p := dictionaryValue(dict, windowLayerKey), dictionaryValue(dict, windowPIDKey)
		if l == 0 || p == 0 || !numberValue(l, 4, &layer) || !numberValue(p, 4, &pid) || layer != 0 || pid == int64(os.Getpid()) {
			continue
		}
		b := dictionaryValue(dict, windowBoundsKey)
		var r windowRect
		if b == 0 || !rectValue(b, &r) || r.W <= 0 || r.H <= 0 {
			continue
		}
		result = append(result, image.Rect(int(math.Floor(r.X)), int(math.Floor(r.Y)), int(math.Ceil(r.X+r.W)), int(math.Ceil(r.Y+r.H))))
	}
	return result, nil
}
