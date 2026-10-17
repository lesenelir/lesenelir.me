//go:build darwin

// Kit is a personal toolbox for macOS: a menu bar app whose tools share one
// settings window. Its first tool is voice input; more come as features of
// their own under internal/.
package main

import (
	"log"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/voice"
)

func main() {
	kit := app.New()
	kit.Add(voice.New())
	if err := kit.Run(); err != nil {
		log.Fatal(err)
	}
}
