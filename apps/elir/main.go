//go:build darwin || windows

// Elir is a personal toolbox for macOS: a menu bar app whose tools share one
// settings window. Its tools are voice input, translating and polishing
// selected text, and screenshots; more come as features of their own under
// internal/.
package main

import (
	"log"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/rewrite"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/screenshot"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/voice"
)

func main() {
	elir := app.New()
	elir.Add(voice.New())
	elir.Add(rewrite.New())
	elir.Add(screenshot.New())
	if err := elir.Run(); err != nil {
		log.Fatal(err)
	}
}
