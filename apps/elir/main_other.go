//go:build !darwin && !windows

package main

import "log"

// Windows support is planned: it needs its own hotkey hook, recorder and
// paster under internal/platform.
func main() {
	log.Fatal("Elir supports macOS and Windows only")
}
