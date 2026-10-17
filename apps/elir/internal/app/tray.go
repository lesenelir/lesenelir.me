//go:build darwin

package app

import (
	_ "embed"
	"log"

	"github.com/egoist/mygo"

	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

func (a *App) setupTray() {
	var err error
	a.tray, err = mygo.NewTray(mygo.TrayOptions{
		Icon:           trayIcon,
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

// trayIcon is the head of the app icon's character, black on transparent
// for macOS to tint: 16 points, drawn at 32 pixels for Retina by
// design/icon/menubar.go.
//
//go:embed trayicon.png
var trayIcon []byte
