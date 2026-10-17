//go:build darwin

package app

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"

	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

func (a *App) setupTray() {
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

// trayMenu is the menu bar icon's menu, in the language the app shows,
// led by the update found, if any. Main thread only.
func (a *App) trayMenu() *mygo.Menu {
	var items []*mygo.MenuItem
	if a.upd.state == updateAvailable {
		items = append(items,
			&mygo.MenuItem{Label: locales.TrayUpdate(a.upd.rel.Version()), Click: func(*mygo.MenuItem, *mygo.Window) { a.showSettings("general") }},
			mygo.Separator())
	}
	return mygo.NewMenu(append(items, []*mygo.MenuItem{
		{Label: locales.TraySettings(), Click: func(*mygo.MenuItem, *mygo.Window) { a.showSettings("") }},
		mygo.Separator(),
		{Label: locales.TrayQuit(), Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	}...))
}

// localizeTray says the tray's words again, after the language changes or
// an update is found.
func (a *App) localizeTray() {
	if a.tray == nil {
		return
	}
	a.tray.SetToolTip(locales.TrayTooltip())
	a.tray.SetMenu(a.trayMenu())
}

// trayIcon draws a toolbox, black on transparent for macOS to tint.
func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	ink := image.NewUniform(color.NRGBA{A: 255})
	for _, r := range []image.Rectangle{
		image.Rect(12, 5, 20, 7),   // top of the handle
		image.Rect(12, 7, 14, 11),  // left of the handle
		image.Rect(18, 7, 20, 11),  // right of the handle
		image.Rect(4, 11, 28, 17),  // lid
		image.Rect(4, 19, 28, 27),  // body
		image.Rect(14, 15, 18, 21), // latch, across the seam
	} {
		draw.Draw(img, r, ink, image.Point{}, draw.Src)
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}
