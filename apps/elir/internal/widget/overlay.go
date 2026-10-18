//go:build darwin || windows

package widget

import (
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// NewOverlay makes a hidden, nonactivating window whose view draws its shadow.
// Callers position it before showing it, independently of the tray's position.
func NewOverlay(tray *mygo.Tray, title string, width, height int, view func(*ui.Context)) *mygo.Window {
	if runtime.GOOS == "windows" && tray != nil {
		// shortcut: MyGo 0.5.5 needs flyouts for material-free Windows alpha; remove when transparent panels work.
		w := mygo.NewFlyout(mygo.FlyoutOptions{
			Tray: tray, Hidden: true, Width: width, Height: height, Content: ui.View(view),
		})
		w.SetTitle(title)
		return w
	}
	return mygo.NewWindow(mygo.WindowOptions{
		Title: title, Panel: true, Hidden: true, Frameless: true,
		Transparent: true, DisableShadow: true,
		Width: width, Height: height, Content: ui.View(view),
	})
}
