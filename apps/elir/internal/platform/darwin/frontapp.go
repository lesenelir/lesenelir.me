//go:build darwin

package darwin

import (
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

// FrontApp returns the bundle identifier of the app in front, "" when there
// is none or it has none. It may be called from any goroutine.
func FrontApp() string {
	workspace := objc.ID(objc.GetClass("NSWorkspace"))
	pools := objc.ID(objc.GetClass("NSAutoreleasePool"))
	if workspace == 0 || pools == 0 {
		return ""
	}
	// The objects below are autoreleased, and a goroutine's thread has no
	// pool of its own.
	pool := pools.Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	app := workspace.Send(objc.RegisterName("sharedWorkspace")).Send(objc.RegisterName("frontmostApplication"))
	if app == 0 {
		return ""
	}
	id := app.Send(objc.RegisterName("bundleIdentifier"))
	if id == 0 {
		return ""
	}
	return goString(objc.Send[*byte](id, objc.RegisterName("UTF8String")))
}

// goString copies a C string.
func goString(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}
