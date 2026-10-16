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
)

func (a *app) setupTray() {
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Label: "设置…", Click: func(*mygo.MenuItem, *mygo.Window) { a.showSettings() }},
		mygo.Separator(),
		{Label: "退出 Voice", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
	var err error
	a.tray, err = mygo.NewTray(mygo.TrayOptions{
		Icon:           trayIcon(),
		IconIsTemplate: true,
		ToolTip:        "Voice — Fn 听写，Fn + 左 Shift 翻译",
		Menu:           menu,
	})
	if err != nil {
		log.Printf("tray: %v", err)
		return
	}
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) { a.tray.Destroy() })
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
