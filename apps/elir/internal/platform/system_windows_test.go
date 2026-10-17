package platform

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestSystemPathsIgnoreEnvironment(t *testing.T) {
	actual := SystemDirectory()
	t.Setenv("SystemRoot", t.TempDir())
	t.Setenv("WINDIR", t.TempDir())
	if !filepath.IsAbs(actual) || !strings.EqualFold(filepath.Base(actual), "System32") {
		t.Fatalf("unexpected system directory %q", actual)
	}
	for _, name := range []string{"user32.dll", "winmm.dll", "ole32.dll"} {
		dll := systemDLL(name)
		defer dll.Release()
		if dll.Name != filepath.Join(actual, name) {
			t.Fatalf("untrusted DLL path %q", dll.Name)
		}
		var loaded [32768]uint16
		n, _, err := kernel32.NewProc("GetModuleFileNameW").Call(uintptr(dll.Handle), uintptr(unsafe.Pointer(&loaded[0])), uintptr(len(loaded)))
		if n == 0 || !strings.EqualFold(syscall.UTF16ToString(loaded[:n]), dll.Name) {
			t.Fatalf("unexpected loaded DLL: %v", err)
		}
	}
}
