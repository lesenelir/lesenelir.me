//go:build !darwin

package main

import "log"

// Windows support is planned: it needs its own hotkey hook, recorder and
// paster under internal/platform.
func main() {
	log.Fatal("Elir supports macOS only for now")
}
