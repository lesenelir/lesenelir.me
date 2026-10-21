package platform

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	foregroundWindow = user32.MustFindProc("GetForegroundWindow")
	windowProcess    = user32.MustFindProc("GetWindowThreadProcessId")
	openProcess      = kernel32.NewProc("OpenProcess")
	processImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

// FrontApp returns the executable name of the app whose window is in
// front, such as WeChat.exe, "" when there is none or it cannot be read.
func FrontApp() string {
	window, _, _ := foregroundWindow.Call()
	if window == 0 {
		return ""
	}
	var pid uint32
	windowProcess.Call(window, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}
	const processQueryLimitedInformation = 0x1000
	process, _, _ := openProcess.Call(processQueryLimitedInformation, 0, uintptr(pid))
	if process == 0 {
		return ""
	}
	defer closeHandle.Call(process)
	var buffer [1024]uint16
	size := uint32(len(buffer))
	if ok, _, _ := processImageName.Call(process, 0, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size))); ok == 0 {
		return ""
	}
	return filepath.Base(syscall.UTF16ToString(buffer[:size]))
}
