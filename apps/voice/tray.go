//go:build darwin

package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"

	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/voice/locales"
)

func (a *app) setupTray() {
	var err error
	a.tray, err = mygo.NewTray(mygo.TrayOptions{
		Icon:           trayIcon(),
		IconIsTemplate: true,
		ToolTip:        locales.TrayTooltip(),
		Menu:           a.trayMenu(),
	})
	if err != nil {
		log.Printf("tray: %v", err)
		return
	}
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) { a.tray.Destroy() })
}

// trayMenu is the menu bar icon's menu, in the language the app shows.
func (a *app) trayMenu() *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Label: locales.TraySettings(), Click: func(*mygo.MenuItem, *mygo.Window) { a.showSettings() }},
		mygo.Separator(),
		{Label: locales.TrayQuit(), Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
}

// localizeTray says the tray's words again after the language changes.
func (a *app) localizeTray() {
	if a.tray == nil {
		return
	}
	a.tray.SetToolTip(locales.TrayTooltip())
	a.tray.SetMenu(a.trayMenu())
}

// trayIcon draws a microphone, black on transparent for macOS to tint.
func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	ink := image.NewUniform(color.NRGBA{A: 255})
	for _, r := range []image.Rectangle{
		image.Rect(12, 3, 20, 19),  // capsule
		image.Rect(7, 13, 9, 20),   // left arm of the holder
		image.Rect(23, 13, 25, 20), // right arm
		image.Rect(9, 20, 23, 22),  // bottom of the holder
		image.Rect(15, 22, 17, 27), // stem
		image.Rect(10, 27, 22, 29), // base
	} {
		draw.Draw(img, r, ink, image.Point{}, draw.Src)
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}
