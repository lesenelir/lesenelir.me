package platform

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Go itself restricts kernel32.dll to System32. Bootstrap all other paths from
// the OS API, rather than the caller's PATH or SystemRoot environment variables.
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var systemDirectory = func() string {
	var buffer [32768]uint16
	n, _, err := kernel32.NewProc("GetSystemDirectoryW").Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if n == 0 || n >= uintptr(len(buffer)) {
		panic(fmt.Sprintf("GetSystemDirectoryW: %v", err))
	}
	return syscall.UTF16ToString(buffer[:n])
}()

// SystemDirectory is the trusted OS directory, independent of user environment.
func SystemDirectory() string { return systemDirectory }

func systemDLL(name string) *syscall.DLL {
	path := filepath.Join(systemDirectory, name)
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
	}
	// Restrict both this DLL and its dependencies; an absolute path alone
	// only constrains the requested DLL. Keep the handle for the process lifetime.
	handle, _, err := kernel32.NewProc("LoadLibraryExW").Call(uintptr(unsafe.Pointer(p)), 0, 0x800) // LOAD_LIBRARY_SEARCH_SYSTEM32
	if handle == 0 {
		panic(fmt.Sprintf("load system DLL %s: %v", name, err))
	}
	return &syscall.DLL{Name: path, Handle: syscall.Handle(handle)}
}
