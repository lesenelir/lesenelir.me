package widget

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Opt in: this shows native windows briefly and samples their desktop pixels.
func TestMain(m *testing.M) {
	if os.Getenv("ELIR_OVERLAY_DESKTOP_TEST") != "1" {
		os.Exit(m.Run())
	}
	mygo.App.OnWindowAllClosed(func() {})
	code := make(chan int, 1)
	mygo.App.WhenReady(func() {
		go func() {
			code <- m.Run()
			mygo.App.Quit()
		}()
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(<-code)
}

func TestOverlayDesktop(t *testing.T) {
	if os.Getenv("ELIR_OVERLAY_DESKTOP_TEST") != "1" {
		t.Skip("set ELIR_OVERLAY_DESKTOP_TEST=1 in a desktop session")
	}
	tray, err := mygo.NewTray(mygo.TrayOptions{ToolTip: "Elir overlay test"})
	if err != nil {
		t.Fatal(err)
	}
	defer tray.Destroy()
	view := func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Column(c).Fill().Center().Children(func() {
			ui.Box(c).Size(100, 40).Radius(20).Background(ui.Hex("#141416"))
		})
	}
	// A known background distinguishes transparent pixels from black pixels.
	back := mygo.NewWindow(mygo.WindowOptions{
		Panel: true, Hidden: true, Frameless: true, DisableShadow: true,
		Width: 700, Height: 240, Content: ui.View(func(c *ui.Context) {
			c.Root().Background(ui.Hex("#2a699a"))
		}),
	})
	defer back.Destroy()
	area := mygo.Screen.PrimaryDisplay().WorkArea
	x, y := area.X+(area.Width-700)/2, area.Y+(area.Height-240)/2
	back.SetPosition(x, y)
	back.ShowInactive()
	w := NewOverlay(tray, "Elir overlay test", 640, 76, view)
	defer w.Destroy()
	if !w.IsFlyout() || w.IsVisible() || w.HasShadow() || !w.IsAlwaysOnTop() {
		t.Fatal("overlay must be a hidden, topmost flyout without a system shadow")
	}
	user := syscall.NewLazyDLL("user32.dll")
	style, _, _ := user.NewProc("GetWindowLongPtrW").Call(w.NativeHandle(), ^uintptr(19)) // GWL_EXSTYLE = -20
	const wantStyle = 0x00200000 | 0x08000000 | 0x00000080                                // NOREDIRECTIONBITMAP, NOACTIVATE, TOOLWINDOW
	if style&wantStyle != wantStyle {
		t.Fatalf("transparent, nonactivating tool window styles missing: %#x", style)
	}
	for _, size := range [][2]int{{640, 76}, {232, 172}, {640, 76}} {
		w.SetSize(size[0], size[1])
		w.SetPosition(x+24, y+24)
		before, _, _ := user.NewProc("GetForegroundWindow").Call()
		w.ShowInactive()
		after, _, _ := user.NewProc("GetForegroundWindow").Call()
		if before != after || !w.IsVisible() {
			t.Fatal("showing the overlay changed the foreground window or failed")
		}
		b := w.Bounds()
		if b.X != x+24 || b.Y != y+24 || b.Width != size[0] || b.Height != size[1] {
			t.Fatalf("caller's geometry was lost: %+v", b)
		}
		// Wait for both windows to render, including after hide/resize/show.
		deadline := time.Now().Add(5 * time.Second)
		for {
			corner := overlayDesktopPixel(t, w, 8, 8)
			center := overlayDesktopPixel(t, w, size[0]/2, size[1]/2)
			if corner == 0x9a692a && center == 0x161414 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("desktop pixels: corner=%#x want background %#x, center=%#x want capsule %#x", corner, 0x9a692a, center, 0x161414)
			}
			time.Sleep(30 * time.Millisecond)
		}
		t.Logf("%dx%d: desktop shows through the corner, capsule stays opaque, foreground preserved", size[0], size[1])
		w.Hide()
		if w.IsVisible() {
			t.Fatal("overlay did not hide")
		}
	}
}

// Read an actual desktop pixel, rather than an offscreen renderer's alpha.
func overlayDesktopPixel(t *testing.T, w *mygo.Window, x, y int) uintptr {
	t.Helper()
	user := syscall.NewLazyDLL("user32.dll")
	var origin struct{ x, y int32 }
	if ok, _, err := user.NewProc("ClientToScreen").Call(w.NativeHandle(), uintptr(unsafe.Pointer(&origin))); ok == 0 {
		t.Fatal(err)
	}
	// Client geometry is in physical pixels; MyGo's size is in DIPs.
	dpi, _, _ := user.NewProc("GetDpiForWindow").Call(w.NativeHandle())
	dc, _, err := user.NewProc("GetDC").Call(0)
	if dc == 0 {
		t.Fatal(err)
	}
	defer user.NewProc("ReleaseDC").Call(0, dc)
	pixel, _, _ := syscall.NewLazyDLL("gdi32.dll").NewProc("GetPixel").Call(dc,
		uintptr(origin.x)+uintptr(x)*dpi/96, uintptr(origin.y)+uintptr(y)*dpi/96)
	return pixel
}
