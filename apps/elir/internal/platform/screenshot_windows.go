package platform

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

var gdi32 = systemDLL("gdi32.dll")
var (
	shotGetDC           = user32.MustFindProc("GetDC")
	shotReleaseDC       = user32.MustFindProc("ReleaseDC")
	shotCreateDC        = gdi32.MustFindProc("CreateCompatibleDC")
	shotDeleteDC        = gdi32.MustFindProc("DeleteDC")
	shotCreateDIB       = gdi32.MustFindProc("CreateDIBSection")
	shotSelectObject    = gdi32.MustFindProc("SelectObject")
	shotDeleteObject    = gdi32.MustFindProc("DeleteObject")
	shotBitBlt          = gdi32.MustFindProc("BitBlt")
	shotFlush           = gdi32.MustFindProc("GdiFlush")
	shotEnumMonitors    = user32.MustFindProc("EnumDisplayMonitors")
	shotSetDPI, _       = user32.FindProc("SetThreadDpiAwarenessContext")
	shotGetDPI, _       = user32.FindProc("GetThreadDpiAwarenessContext")
	shotDPIAwareness, _ = user32.FindProc("GetAwarenessFromDpiAwarenessContext")
	shotMoveFile        = kernel32.NewProc("MoveFileExW")
)

type shotRect struct{ Left, Top, Right, Bottom int32 }

func (r shotRect) image() image.Rectangle {
	return image.Rectangle{Min: image.Pt(int(r.Left), int(r.Top)), Max: image.Pt(int(r.Right), int(r.Bottom))}
}

var shotIDs atomic.Uint64
var shotMonitorSessions sync.Map
var shotMonitorCallback = syscall.NewCallback(func(_, _ uintptr, rect, id uintptr) uintptr {
	if v, ok := shotMonitorSessions.Load(id); ok {
		p := v.(*[]image.Rectangle)
		*p = append(*p, (*shotRect)(native(rect)).image())
		return 1
	}
	return 0
})

func screenshotMonitors() ([]image.Rectangle, error) {
	var monitors []image.Rectangle
	id := uintptr(shotIDs.Add(1))
	shotMonitorSessions.Store(id, &monitors)
	defer shotMonitorSessions.Delete(id)
	r, _, err := shotEnumMonitors.Call(0, 0, shotMonitorCallback, id)
	if r == 0 || len(monitors) == 0 {
		return nil, fmt.Errorf("enumerate displays: %v", err)
	}
	for _, r := range monitors {
		if r.Empty() {
			return nil, errors.New("enumerate displays: invalid monitor bounds")
		}
	}
	// Enumeration order is not part of the display layout contract.
	slices.SortFunc(monitors, func(a, b image.Rectangle) int {
		for _, pair := range [][2]int{{a.Min.X, b.Min.X}, {a.Min.Y, b.Min.Y}, {a.Max.X, b.Max.X}, {a.Max.Y, b.Max.Y}} {
			if pair[0] < pair[1] {
				return -1
			}
			if pair[0] > pair[1] {
				return 1
			}
		}
		return 0
	})
	return monitors, nil
}

// screenshotSize bounds native allocation and all subsequent Go pixel indexing.
func screenshotSize(r image.Rectangle) (stride, size int, err error) {
	for _, coordinate := range []int{r.Min.X, r.Min.Y, r.Max.X, r.Max.Y} {
		if int64(coordinate) < -1<<31 || int64(coordinate) > 1<<31-1 {
			return 0, 0, errors.New("desktop coordinate exceeds Win32 range")
		}
	}
	w, h := int64(r.Max.X)-int64(r.Min.X), int64(r.Max.Y)-int64(r.Min.Y)
	const limit = 512 << 20
	if w <= 0 || h <= 0 || w > 1<<31-1 || h > 1<<31-1 || w > limit/4 || h > limit/(w*4) {
		return 0, 0, fmt.Errorf("desktop dimensions exceed 512 MiB snapshot budget: %v", r)
	}
	return int(w * 4), int(w * 4 * h), nil
}

func screenshotSelection(r, desktop image.Rectangle, monitors []image.Rectangle) image.Rectangle {
	r = r.Intersect(desktop)
	if r.Dx() <= 1 || r.Dy() <= 1 {
		return image.Rectangle{}
	}
	for _, monitor := range monitors {
		if r.Overlaps(monitor) {
			return r
		}
	}
	return image.Rectangle{}
}

type screenshotDIB struct {
	dc, bitmap, previous uintptr
	pixels               []byte
	bounds               image.Rectangle
	stride               int
}

func (d *screenshotDIB) close() error {
	var err error
	if d.previous != 0 {
		r, _, e := shotSelectObject.Call(d.dc, d.previous)
		if r == 0 || r == ^uintptr(0) {
			err = fmt.Errorf("restore bitmap: %v", e)
		}
		d.previous = 0
	}
	// Deleting the DC also deselects a bitmap if restoration failed.
	if d.dc != 0 {
		r, _, e := shotDeleteDC.Call(d.dc)
		if r == 0 {
			err = errors.Join(err, fmt.Errorf("delete capture DC: %v", e))
		}
		d.dc = 0
	}
	if d.bitmap != 0 {
		r, _, e := shotDeleteObject.Call(d.bitmap)
		if r == 0 {
			err = errors.Join(err, fmt.Errorf("delete capture bitmap: %v", e))
		}
		d.bitmap = 0
	}
	d.pixels = nil
	return err
}

func captureDesktop(monitors []image.Rectangle) (d *screenshotDIB, err error) {
	var bounds image.Rectangle
	for _, r := range monitors {
		bounds = bounds.Union(r)
	}
	if _, _, err := screenshotSize(bounds); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && d != nil {
			err = errors.Join(err, d.close())
			d = nil
		}
	}()
	screen, _, e := shotGetDC.Call(0)
	if screen == 0 {
		return d, fmt.Errorf("capture desktop GetDC: %v", e)
	}
	defer func() {
		r, _, e := shotReleaseDC.Call(0, screen)
		if r == 0 {
			err = errors.Join(err, fmt.Errorf("release desktop DC: %v", e))
		}
	}()
	d, err = newScreenshotDIB(screen, bounds)
	if err != nil {
		return nil, err
	}
	for _, r := range monitors {
		p := r.Min.Sub(d.bounds.Min)
		ok, _, e := shotBitBlt.Call(d.dc, uintptr(p.X), uintptr(p.Y), uintptr(r.Dx()), uintptr(r.Dy()), screen, uintptr(r.Min.X), uintptr(r.Min.Y), 0x40cc0020) // SRCCOPY | CAPTUREBLT
		if ok == 0 {
			return d, fmt.Errorf("capture desktop BitBlt: %v", e)
		}
	}
	if r, _, e := shotFlush.Call(); r == 0 {
		return d, fmt.Errorf("capture desktop GdiFlush: %v", e)
	}
	return d, nil
}

func newScreenshotDIB(screen uintptr, bounds image.Rectangle) (d *screenshotDIB, err error) {
	stride, size, err := screenshotSize(bounds)
	if err != nil {
		return nil, err
	}
	d = &screenshotDIB{bounds: bounds, stride: stride}
	defer func() {
		if err != nil {
			err = errors.Join(err, d.close())
			d = nil
		}
	}()
	var e error
	d.dc, _, e = shotCreateDC.Call(screen)
	if d.dc == 0 {
		return d, fmt.Errorf("capture desktop CreateCompatibleDC: %v", e)
	}
	info := struct {
		Size                   uint32
		Width, Height          int32
		Planes, Bits           uint16
		Compression, ImageSize uint32
		X, Y                   int32
		Used, Important        uint32
		Colors                 [1]uint32
	}{Size: 40, Width: int32(d.bounds.Dx()), Height: -int32(d.bounds.Dy()), Planes: 1, Bits: 32}
	var bits unsafe.Pointer
	d.bitmap, _, e = shotCreateDIB.Call(screen, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if d.bitmap == 0 || bits == nil {
		return d, fmt.Errorf("capture desktop CreateDIBSection: %v", e)
	}
	d.pixels = unsafe.Slice((*byte)(bits), size)
	clear(d.pixels)
	d.previous, _, e = shotSelectObject.Call(d.dc, d.bitmap)
	if d.previous == 0 || d.previous == ^uintptr(0) {
		d.previous = 0
		return d, fmt.Errorf("capture desktop SelectObject: %v", e)
	}
	return d, nil
}

func (d *screenshotDIB) crop(r image.Rectangle) (*image.NRGBA, error) {
	if r.Empty() || !r.In(d.bounds) {
		return nil, errors.New("crop screenshot: invalid bounds")
	}
	if _, _, err := screenshotSize(r); err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	local := r.Sub(d.bounds.Min)
	for y := 0; y < r.Dy(); y++ {
		src := d.pixels[(local.Min.Y+y)*d.stride+local.Min.X*4:][:r.Dx()*4]
		dst := img.Pix[y*img.Stride:][:r.Dx()*4]
		for x := 0; x < len(src); x += 4 {
			dst[x], dst[x+1], dst[x+2], dst[x+3] = src[x+2], src[x+1], src[x], 255
		}
	}
	return img, nil
}

func screenshotDPI() (restore func() error, err error) {
	if shotSetDPI == nil || shotGetDPI == nil || shotDPIAwareness == nil {
		// Without thread APIs, GetProcessDpiAwareness proves the inherited default.
		if shotSetDPI == nil && shotGetDPI == nil {
			dll, e := loadSystemDLL("shcore.dll")
			if e != nil {
				return nil, fmt.Errorf("screenshot DPI: %w", e)
			}
			defer dll.Release()
			query, e := dll.FindProc("GetProcessDpiAwareness")
			if e == nil {
				var awareness int32
				hr, _, _ := query.Call(0, uintptr(unsafe.Pointer(&awareness)))
				if int32(hr) >= 0 && awareness == 2 {
					return func() error { return nil }, nil
				}
			}
		}
		return nil, errors.New("screenshot DPI: per-monitor thread awareness unavailable")
	}
	old, _, e := shotSetDPI.Call(^uintptr(3)) // PER_MONITOR_AWARE_V2 (-4)
	if old == 0 {
		old, _, e = shotSetDPI.Call(^uintptr(2)) // PER_MONITOR_AWARE (-3)
	}
	if old == 0 {
		return nil, fmt.Errorf("screenshot DPI: set per-monitor awareness: %v", e)
	}
	restore = func() error {
		r, _, e := shotSetDPI.Call(old)
		if r == 0 {
			return fmt.Errorf("restore screenshot DPI: %v", e)
		}
		return nil
	}
	current, _, _ := shotGetDPI.Call()
	awareness, _, _ := shotDPIAwareness.Call(current)
	if int32(awareness) != 2 {
		return restore, errors.New("screenshot DPI: thread is not per-monitor aware")
	}
	return restore, nil
}

// CaptureScreenshot runs on a worker goroutine, independently of MyGo's UI.
func CaptureScreenshot(ctx context.Context, path string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	img, err := captureScreenshotImage(ctx)
	if err != nil || img == nil {
		return false, err
	}
	if err := writeScreenshotPNG(ctx, path, img, png.Encode, moveScreenshot); err != nil {
		return false, err
	}
	return true, nil
}

func captureScreenshotImage(ctx context.Context) (img *image.NRGBA, err error) {
	runtime.LockOSThread()
	restore, dpiErr := screenshotDPI()
	defer func() {
		if restore != nil {
			if e := restore(); e != nil {
				err = errors.Join(err, e)
				img = nil
				// Keep this thread locked: Go retires it when the worker exits.
				return
			}
		}
		runtime.UnlockOSThread()
	}()
	if dpiErr != nil {
		return nil, dpiErr
	}
	monitors, err := screenshotMonitors()
	if err != nil {
		return nil, err
	}
	d, err := captureDesktop(monitors)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := d.close(); e != nil {
			err = errors.Join(err, e)
			img = nil
		}
	}()
	r, err := selectScreenshot(ctx, d, monitors)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if r.Empty() {
		return nil, nil
	}
	current, err := screenshotMonitors()
	if err != nil {
		return nil, err
	}
	if !slices.Equal(current, monitors) {
		return nil, errors.New("display configuration changed; retry screenshot")
	}
	return d.crop(r)
}

func moveScreenshot(from, to string) error {
	src, err := syscall.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	dst, err := syscall.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	r, _, err := shotMoveFile.Call(uintptr(unsafe.Pointer(src)), uintptr(unsafe.Pointer(dst)), 0) // No REPLACE_EXISTING.
	if r == 0 {
		return err
	}
	return nil
}

func writeScreenshotPNG(ctx context.Context, path string, img image.Image, encode func(io.Writer, image.Image) error, move func(string, string) error) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".elir-capture-*.tmp")
	if err != nil {
		return fmt.Errorf("write png: %w", err)
	}
	temp := f.Name()
	committed := false
	defer func() {
		if committed {
			return
		}
		if e := os.Remove(temp); e != nil && !errors.Is(e, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove temporary png: %w", e))
		}
	}()
	err = encode(f, img)
	err = errors.Join(err, f.Close())
	if err != nil {
		return fmt.Errorf("write png: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := move(temp, path); err != nil {
		return fmt.Errorf("commit png: %w", err)
	}
	committed = true
	return nil
}
