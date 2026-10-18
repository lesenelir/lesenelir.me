package platform

import (
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
	"unsafe"
)

func TestScreenshotGeometry(t *testing.T) {
	monitors := []image.Rectangle{image.Rect(-200, -100, 0, 100), image.Rect(0, 0, 300, 200)}
	desktop := monitors[0].Union(monitors[1])
	for _, tc := range []struct {
		name     string
		from, to image.Point
		want     image.Rectangle
	}{
		{"negative", image.Pt(-180, -80), image.Pt(-20, 80), image.Rect(-180, -80, -20, 80)},
		{"reverse x", image.Pt(-20, -80), image.Pt(-180, 80), image.Rect(-180, -80, -20, 80)},
		{"reverse y", image.Pt(-180, 80), image.Pt(-20, -80), image.Rect(-180, -80, -20, 80)},
		{"reverse both", image.Pt(-20, 80), image.Pt(-180, -80), image.Rect(-180, -80, -20, 80)},
		{"cross screen", image.Pt(-150, -50), image.Pt(100, 100), image.Rect(-150, -50, 100, 100)},
		{"partial offscreen", image.Pt(-300, -200), image.Pt(100, 100), image.Rect(-200, -100, 100, 100)},
		{"hole", image.Pt(10, -90), image.Pt(80, -10), image.Rectangle{}},
		{"outside", image.Pt(400, 400), image.Pt(500, 500), image.Rectangle{}},
		{"one pixel", image.Pt(10, 10), image.Pt(11, 30), image.Rectangle{}},
		{"zero", image.Pt(10, 10), image.Pt(10, 10), image.Rectangle{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := image.Rectangle{Min: tc.from, Max: tc.to}.Canon()
			if got := screenshotSelection(r, desktop, monitors); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
	for _, r := range []image.Rectangle{
		{}, {Min: image.Pt(20, 20), Max: image.Pt(10, 10)}, image.Rect(0, 0, 1<<30, 1<<30),
		image.Rect(-1<<31, 0, 1<<31-1, 100), image.Rect(0, 0, 16384, 8193),
		image.Rect(-1<<62, 0, 1<<62, 1),
	} {
		if _, _, err := screenshotSize(r); err == nil {
			t.Fatalf("accepted invalid dimensions %v", r)
		}
	}
	if stride, size, err := screenshotSize(image.Rect(-8192, -4096, 8192, 4096)); err != nil || stride != 65536 || size != 512<<20 {
		t.Fatalf("budget boundary: %d, %d, %v", stride, size, err)
	}
}

func TestScreenshotPixelsAndPNG(t *testing.T) {
	// A top-down synthetic DIB with row padding and undefined alpha bytes.
	d := &screenshotDIB{bounds: image.Rect(-2, -1, 1, 1), stride: 16, pixels: []byte{
		0, 0, 255, 0, 0, 255, 0, 17, 255, 0, 0, 255, 99, 99, 99, 99,
		0, 0, 0, 0, 255, 255, 255, 0, 1, 2, 3, 0, 99, 99, 99, 99,
	}}
	img, err := d.crop(image.Rect(-1, -1, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	want := []color.NRGBA{{0, 255, 0, 255}, {0, 0, 255, 255}, {255, 255, 255, 255}, {3, 2, 1, 255}}
	path := filepath.Join(t.TempDir(), "中文 space ' image.png")
	if err := writeScreenshotPNG(context.Background(), path, img, png.Encode, moveScreenshot); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds() != image.Rect(0, 0, 2, 2) {
		t.Fatal(decoded.Bounds())
	}
	for i, c := range want {
		got := color.NRGBAModel.Convert(decoded.At(i%2, i/2)).(color.NRGBA)
		if got != c {
			t.Fatalf("pixel %d: %v, want %v", i, got, c)
		}
	}
	black, err := d.crop(image.Rect(-2, 0, -1, 1))
	if err != nil || black.NRGBAAt(0, 0) != (color.NRGBA{A: 255}) {
		t.Fatalf("black hole: %v, %v", black, err)
	}
	if _, err := d.crop(image.Rect(-3, 0, -1, 1)); err == nil {
		t.Fatal("accepted out-of-bounds crop")
	}
}

func TestScreenshotFileContract(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	sentinel := errors.New("injected failure")
	for _, name := range []string{"success", "pre-cancel", "cancel-before-commit", "encode", "write", "close", "move", "exists", "late-cancel"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "中文 ' screen.png")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			encode := png.Encode
			move := moveScreenshot
			switch name {
			case "pre-cancel":
				cancel()
			case "cancel-before-commit":
				encode = func(w io.Writer, img image.Image) error { err := png.Encode(w, img); cancel(); return err }
			case "encode":
				encode = func(w io.Writer, _ image.Image) error { _, _ = w.Write([]byte("partial")); return sentinel }
			case "write":
				encode = func(w io.Writer, img image.Image) error { _ = w.(*os.File).Close(); return png.Encode(w, img) }
			case "close":
				encode = func(w io.Writer, _ image.Image) error { return w.(*os.File).Close() }
			case "move":
				move = func(_, _ string) error { return sentinel }
			case "exists":
				if err := os.WriteFile(path, []byte("keep original"), 0600); err != nil {
					t.Fatal(err)
				}
			case "late-cancel":
				move = func(from, to string) error { err := moveScreenshot(from, to); cancel(); return err }
			}
			err := writeScreenshotPNG(ctx, path, img, encode, move)
			if name == "success" || name == "late-cancel" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Fatal("missing error")
				}
				if (name == "pre-cancel" || name == "cancel-before-commit") && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if (name == "encode" || name == "move") && !errors.Is(err, sentinel) {
					t.Fatal(err)
				}
				if name == "exists" && !errors.Is(err, os.ErrExist) {
					t.Fatalf("expected file conflict, got %v", err)
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			wantFiles := 0
			if name == "exists" || name == "success" || name == "late-cancel" {
				wantFiles = 1
			}
			if len(entries) != wantFiles {
				t.Fatalf("temporary files leaked: %v", entries)
			}
			if name == "exists" {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != "keep original" {
					t.Fatalf("overwrote existing target: %q, %v", data, err)
				}
			}
		})
	}
	if err := writeScreenshotPNG(context.Background(), filepath.Join(t.TempDir(), "absent", "image.png"), img, png.Encode, moveScreenshot); err == nil {
		t.Fatal("missing CreateTemp error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ok, err := CaptureScreenshot(ctx, filepath.Join(t.TempDir(), "unused.png")); ok || !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-cancel: %v %v", ok, err)
	}
}

func TestScreenshotTerminalState(t *testing.T) {
	accepted := image.Rect(-10, -10, 20, 20)
	for _, message := range []uintptr{0x8, 0x6, 0x215, 0x10, 0x1f, 0x7e, 0x2e0} {
		s := &screenshotOverlay{activated: true, dragging: true}
		shotWindows.Store(uintptr(0), s)
		screenshotWindowProc(0, message, 0, 1)
		shotWindows.Delete(uintptr(0))
		if !s.done || !s.selection.Empty() {
			t.Fatalf("message %#x did not terminate: %+v", message, s)
		}
		if (message == 0x7e || message == 0x2e0) && s.err == nil {
			t.Fatal("display change without error")
		}
	}
	for _, message := range []uintptr{0x8, 0x6, 0x215} {
		s := &screenshotOverlay{activated: true, dragging: true}
		s.finish(accepted, nil)
		shotWindows.Store(uintptr(0), s)
		screenshotWindowProc(0, message, 0, 1)
		shotWindows.Delete(uintptr(0))
		s.finish(image.Rectangle{}, errors.New("late error"))
		if s.selection != accepted || s.err != nil {
			t.Fatal("terminal result overwritten")
		}
	}
	// Hiding for Space must suppress its own focus/capture notifications.
	s := &screenshotOverlay{finishing: true, activated: true, dragging: true}
	shotWindows.Store(uintptr(0), s)
	for _, message := range []uintptr{0x8, 0x6, 0x215} {
		screenshotWindowProc(0, message, 0, 1)
	}
	shotWindows.Delete(uintptr(0))
	if s.done {
		t.Fatal("self-hide cancelled window selection")
	}
	s.finish(accepted, nil)
	if s.selection != accepted {
		t.Fatal("Space could not finish")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s = &screenshotOverlay{ctx: ctx}
	shotWindows.Store(uintptr(0), s)
	screenshotWindowProc(0, 0x113, 1, 0)
	shotWindows.Delete(uintptr(0))
	if !errors.Is(s.err, context.Canceled) {
		t.Fatal(s.err)
	}
	// A fresh session is independent of the previous terminal result.
	s = &screenshotOverlay{}
	shotWindows.Store(uintptr(0), s)
	screenshotWindowProc(0, 0x100, 0x1b, 0)
	shotWindows.Delete(uintptr(0))
	if !s.done || s.err != nil {
		t.Fatal("Esc failed")
	}
	if err := (&screenshotDIB{}).close(); err != nil {
		t.Fatal(err)
	}
	if err := (&screenshotOverlay{}).close(); err != nil {
		t.Fatal(err)
	}
}

func TestScreenshotDPI(t *testing.T) {
	runtime.LockOSThread()
	unlock := true
	defer func() {
		if unlock {
			runtime.UnlockOSThread()
		}
	}()
	if shotGetDPI == nil {
		t.Skip("thread DPI queries unavailable")
	}
	before, _, _ := shotGetDPI.Call()
	restore, err := screenshotDPI()
	if restore != nil {
		defer func() {
			if err := restore(); err != nil {
				unlock = false
				t.Error(err)
				return
			}
			after, _, _ := shotGetDPI.Call()
			equal, _ := user32.FindProc("AreDpiAwarenessContextsEqual")
			if equal == nil {
				t.Error("DPI comparison unavailable")
				return
			}
			if same, _, _ := equal.Call(before, after); same == 0 {
				t.Error("DPI context not restored")
			}
		}()
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestScreenshotQueuedMousePosition(t *testing.T) {
	// Queued button-up must use its historical 32-bit screen point even if the
	// live pointer has moved elsewhere, including negative and >16-bit positions.
	bounds := image.Rect(-40000, -1000, 4000, 2000)
	for _, points := range [][2]image.Point{
		{image.Pt(-39000, 20), image.Pt(-35000, 50)},
		{image.Pt(-35000, 50), image.Pt(-39000, 20)},
		{image.Pt(1640, 620), image.Pt(2240, 860)},
		{image.Pt(2240, 860), image.Pt(1640, 620)},
	} {
		s := &screenshotOverlay{d: &screenshotDIB{bounds: bounds}, monitors: []image.Rectangle{bounds}, start: points[0], pointer: points[1], dragging: true}
		shotWindows.Store(uintptr(0), s)
		screenshotWindowProc(0, 0x202, 0, 0) // lParam is deliberately unrelated.
		shotWindows.Delete(uintptr(0))
		want := (image.Rectangle{Min: points[0], Max: points[1]}).Canon()
		if !s.done || s.err != nil || s.selection != want {
			t.Fatalf("queued point lost: %v, want %v, err %v", s.selection, want, s.err)
		}
	}
}

func TestScreenshotUnavailableDPIQuery(t *testing.T) {
	old := shotDPIAwareness
	shotDPIAwareness = nil
	defer func() { shotDPIAwareness = old }()
	if restore, err := screenshotDPI(); restore != nil || err == nil {
		t.Fatalf("unprovable DPI allowed: %v", err)
	}
}

func TestScreenshotNativeDIBLifetime(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	for range 100 {
		// No GetDC or desktop reads: create an empty, offscreen native bitmap.
		d, err := newScreenshotDIB(0, image.Rect(-2, -1, 2, 2))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range d.pixels {
			if b != 0 {
				t.Fatal("uninitialized snapshot pixels")
			}
		}
		d.pixels[0], d.pixels[2] = 1, 254
		img, err := d.crop(d.bounds)
		if err != nil || img.NRGBAAt(0, 0) != (color.NRGBA{R: 254, B: 1, A: 255}) {
			t.Fatalf("native DIB pixels: %v %v", img, err)
		}
		if err := d.close(); err != nil {
			t.Fatal(err)
		}
		if err := d.close(); err != nil {
			t.Fatal("non-idempotent cleanup", err)
		}
	}
	// Partial allocation (DC acquired, bitmap not acquired) also owns its DC.
	dc, _, err := shotCreateDC.Call(0)
	if dc == 0 {
		t.Fatal(err)
	}
	if err := (&screenshotDIB{dc: dc}).close(); err != nil {
		t.Fatal(err)
	}
	if d, err := newScreenshotDIB(^uintptr(0), image.Rect(0, 0, 2, 2)); d != nil || err == nil {
		t.Fatalf("invalid DC succeeded: %v %v", d, err)
	}
}

func TestScreenshotHiddenWindowLifecycle(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := registerScreenshotClass(); err != nil {
		t.Fatal(err)
	}
	d, err := newScreenshotDIB(0, image.Rect(0, 0, 16, 16))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := d.close(); err != nil {
			t.Error(err)
		}
	}()
	resources := user32.MustFindProc("GetGuiResources")
	gdiBefore, _, _ := resources.Call(^uintptr(0), 0)
	userBefore, _, _ := resources.Call(^uintptr(0), 1)
	module, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	post := user32.MustFindProc("PostMessageW")
	isWindow := user32.MustFindProc("IsWindow")
	for i := range 100 {
		message := []uintptr{0x10, 0x100, 0x113, 0x7e, 0x2e0}[i%5]
		func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			s := &screenshotOverlay{ctx: ctx, d: d}
			id := uintptr(shotIDs.Add(1))
			shotCreating.Store(id, s)
			defer shotCreating.Delete(id)
			hwnd, _, e := shotCreateWindow.Call(0x88, uintptr(unsafe.Pointer(shotClassName)), 0, 0x80000000, 0, 0, 16, 16, 0, 0, module, id)
			if hwnd == 0 {
				t.Fatal(e)
			}
			defer func() {
				if err := s.close(); err != nil {
					t.Error(err)
				}
				if _, exists := shotWindows.Load(hwnd); exists {
					t.Error("HWND registry leak")
				}
				if exists, _, _ := isWindow.Call(hwnd); exists != 0 {
					t.Error("selection HWND survived cleanup")
				}
			}()
			s.timer, _, e = shotSetTimer.Call(hwnd, 1, 50, 0)
			if s.timer == 0 {
				t.Fatal(e)
			}
			cancel()
			if message == 0x2e0 {
				// WM_DPICHANGED contains a pointer and cannot be posted asynchronously.
				user32.MustFindProc("SendMessageW").Call(hwnd, message, 0, 0)
			} else if ok, _, e := post.Call(hwnd, message, 0x1b, 0); ok == 0 {
				t.Fatal(e)
			}
			var msg struct {
				Window         uintptr
				Message        uint32
				WParam, LParam uintptr
				Time           uint32
				Point          shotPoint
				Private        uint32
			}
			for !s.done {
				r, _, e := shotGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
				if r == 0 || int32(r) == -1 {
					t.Fatalf("hidden message loop: %v", e)
				}
				shotDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
			}
			if message == 0x113 && !errors.Is(s.err, context.Canceled) {
				t.Fatal(s.err)
			}
			if (message == 0x7e || message == 0x2e0) && s.err == nil {
				t.Fatal("display change accepted")
			}
		}()
	}
	gdiAfter, _, _ := resources.Call(^uintptr(0), 0)
	userAfter, _, _ := resources.Call(^uintptr(0), 1)
	t.Logf("100 hidden sessions: GDI %d -> %d, USER %d -> %d", gdiBefore, gdiAfter, userBefore, userAfter)
	// Process-wide USER totals fluctuate; assert each owned HWND above instead.
	if gdiAfter > gdiBefore {
		t.Fatal("native GDI resources grew")
	}
}

// Explicit opt-in only: captures the real desktop into native memory, shows the
// overlay, then cancels; never writes an image or touches the clipboard.
func TestScreenshotDesktopCancellation(t *testing.T) {
	if os.Getenv("ELIR_SCREENSHOT_DESKTOP_TEST") != "1" {
		t.Skip("requires an interactive Windows desktop")
	}
	var version struct {
		Size, Major, Minor, Build, Platform uint32
		Name                                [128]uint16
	}
	version.Size = uint32(unsafe.Sizeof(version))
	ntdll := systemDLL("ntdll.dll")
	defer ntdll.Release()
	if p, e := ntdll.FindProc("RtlGetVersion"); e == nil {
		p.Call(uintptr(unsafe.Pointer(&version)))
	}
	t.Logf("Windows %d.%d build %d", version.Major, version.Minor, version.Build)
	runtime.LockOSThread()
	restore, err := screenshotDPI()
	if err != nil {
		t.Fatal(err)
	}
	monitors, err := screenshotMonitors()
	if err != nil {
		t.Fatal(err)
	}
	getDPI, _ := user32.FindProc("GetDpiForSystem")
	if getDPI != nil {
		dpi, _, _ := getDPI.Call()
		t.Logf("system DPI %d (not per-monitor calibration)", dpi)
	}
	t.Logf("physical display bounds: %v", monitors)
	if err := restore(); err != nil {
		t.Fatal(err)
	}
	runtime.UnlockOSThread()
	for range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		path := filepath.Join(t.TempDir(), "must-not-exist.png")
		ok, err := CaptureScreenshot(ctx, path)
		cancel()
		if ok || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("desktop cancellation: %v %v", ok, err)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("cancelled capture left a file")
		}
	}
	for _, registry := range []*sync.Map{&shotCreating, &shotWindows, &shotMonitorSessions} {
		registry.Range(func(key, _ any) bool { t.Errorf("session registry leaked token %v", key); return true })
	}
}
