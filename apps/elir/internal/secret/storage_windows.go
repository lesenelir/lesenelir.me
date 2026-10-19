package secret

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

var crypt32 = syscall.NewLazyDLL("crypt32.dll")
var protect = crypt32.NewProc("CryptProtectData")
var unprotect = crypt32.NewProc("CryptUnprotectData")
var localFree = syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")

type dataBlob struct {
	Size uint32
	Data *byte
}

// DPAPI binds the file contents to the current Windows user's credentials.
// POSIX mode bits alone do not protect a secret on Windows.
func protectData(data []byte, encrypt bool) ([]byte, error) {
	var in, out dataBlob
	in.Size = uint32(len(data))
	if len(data) != 0 {
		in.Data = &data[0]
	}
	proc := unprotect
	if encrypt {
		proc = protect
	}
	r, _, err := proc.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, 1, uintptr(unsafe.Pointer(&out)))
	runtime.KeepAlive(data)
	if r == 0 {
		return nil, fmt.Errorf("DPAPI: %w", err)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

// decodeSecretFile decrypts the file's JSON.
func decodeSecretFile(data []byte) ([]byte, error) { return protectData(data, false) }

func writeSecretFile(path string, data []byte) error {
	encrypted, err := protectData(data, true)
	if err != nil {
		return err
	}
	return os.WriteFile(path, encrypted, 0600)
}

func fromKeychain(string) string { return "" }
