package platform

import (
	"context"
	"errors"
	"fmt"
	"image"
	"slices"
	"sync"
	"syscall"
	"unsafe"
)

var (
	shotRegisterClass    = user32.MustFindProc("RegisterClassExW")
	shotCreateWindow     = user32.MustFindProc("CreateWindowExW")
	shotDefWindow        = user32.MustFindProc("DefWindowProcW")
	shotDestroyWindow    = user32.MustFindProc("DestroyWindow")
	shotShowWindow       = user32.MustFindProc("ShowWindow")
	shotSetForeground    = user32.MustFindProc("SetForegroundWindow")
	shotGetForeground    = user32.MustFindProc("GetForegroundWindow")
	shotSetFocus         = user32.MustFindProc("SetFocus")
	shotGetFocus         = user32.MustFindProc("GetFocus")
	shotSetCapture       = user32.MustFindProc("SetCapture")
	shotGetCapture       = user32.MustFindProc("GetCapture")
	shotReleaseCapture   = user32.MustFindProc("ReleaseCapture")
	shotCursorPos        = user32.MustFindProc("GetCursorPos")
	shotSetTimer         = user32.MustFindProc("SetTimer")
	shotKillTimer        = user32.MustFindProc("KillTimer")
	shotGetMessage       = user32.MustFindProc("GetMessageW")
	shotTranslateMessage = user32.MustFindProc("TranslateMessage")
	shotDispatchMessage  = user32.MustFindProc("DispatchMessageW")
	shotBeginPaint       = user32.MustFindProc("BeginPaint")
	shotEndPaint         = user32.MustFindProc("EndPaint")
	shotInvalidate       = user32.MustFindProc("InvalidateRect")
	shotUpdateWindow     = user32.MustFindProc("UpdateWindow")
	shotLoadCursor       = user32.MustFindProc("LoadCursorW")
	shotFrameRect        = user32.MustFindProc("FrameRect")
	shotStockObject      = gdi32.MustFindProc("GetStockObject")
	shotBrushColor       = gdi32.MustFindProc("SetDCBrushColor")
	shotWindowAt         = user32.MustFindProc("WindowFromPoint")
	shotAncestor         = user32.MustFindProc("GetAncestor")
	shotWindowRect       = user32.MustFindProc("GetWindowRect")
	shotDWM, _           = systemDLL("dwmapi.dll").FindProc("DwmGetWindowAttribute")
)

type shotPoint struct{ X, Y int32 }

func (p shotPoint) image() image.Point { return image.Pt(int(p.X), int(p.Y)) }

type screenshotOverlay struct {
	ctx                                  context.Context
	d                                    *screenshotDIB
	monitors                             []image.Rectangle
	hwnd, timer                          uintptr
	start                                image.Point
	selection                            image.Rectangle
	dragging, activated, finishing, done bool
	err                                  error
}

// The registries contain Go references; Windows retains only an integer token.
var shotCreating sync.Map
var shotWindows sync.Map
var shotClassOnce sync.Once
var shotClassErr error
var shotClassName = syscall.StringToUTF16Ptr("Elir.NativeScreenshot.1")
var shotWindowCallback = syscall.NewCallback(screenshotWindowProc)

func registerScreenshotClass() error {
	shotClassOnce.Do(func() {
		cursor, _, err := shotLoadCursor.Call(0, 32515) // IDC_CROSS
		if cursor == 0 {
			shotClassErr = fmt.Errorf("load selection cursor: %v", err)
			return
		}
		module, _, err := kernel32.NewProc("GetModuleHandleW").Call(0)
		if module == 0 {
			shotClassErr = fmt.Errorf("get selection module: %v", err)
			return
		}
		class := struct {
			Size, Style                        uint32
			Proc                               uintptr
			ClassExtra, WindowExtra            int32
			Instance, Icon, Cursor, Background uintptr
			Menu, Name                         *uint16
			SmallIcon                          uintptr
		}{Proc: shotWindowCallback, Instance: module, Cursor: cursor, Name: shotClassName}
		class.Size = uint32(unsafe.Sizeof(class))
		r, _, err := shotRegisterClass.Call(uintptr(unsafe.Pointer(&class)))
		if r == 0 {
			shotClassErr = fmt.Errorf("register selection window: %v", err)
		}
	})
	return shotClassErr
}

func selectScreenshot(ctx context.Context, d *screenshotDIB, monitors []image.Rectangle) (selection image.Rectangle, err error) {
	if err := ctx.Err(); err != nil {
		return selection, err
	}
	if err := registerScreenshotClass(); err != nil {
		return selection, err
	}
	s := &screenshotOverlay{ctx: ctx, d: d, monitors: monitors}
	id := uintptr(shotIDs.Add(1))
	shotCreating.Store(id, s)
	defer shotCreating.Delete(id)
	defer func() {
		err = errors.Join(err, s.close())
		if err != nil {
			selection = image.Rectangle{}
		}
	}()
	module, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	v := d.bounds
	hwnd, _, e := shotCreateWindow.Call(0x88, uintptr(unsafe.Pointer(shotClassName)), 0, 0x80000000, uintptr(v.Min.X), uintptr(v.Min.Y), uintptr(v.Dx()), uintptr(v.Dy()), 0, 0, module, id) // TOPMOST | TOOLWINDOW, WS_POPUP
	if hwnd == 0 {
		return selection, fmt.Errorf("create selection window: %v", e)
	}
	s.hwnd = hwnd
	s.timer, _, e = shotSetTimer.Call(hwnd, 1, 50, 0)
	if s.timer == 0 {
		return selection, fmt.Errorf("create selection timer: %v", e)
	}
	shotShowWindow.Call(hwnd, 5)
	shotSetForeground.Call(hwnd)
	shotSetFocus.Call(hwnd)
	foreground, _, _ := shotGetForeground.Call()
	focus, _, _ := shotGetFocus.Call()
	if foreground != hwnd || focus != hwnd {
		return selection, errors.New("activate selection window: keyboard focus unavailable")
	}
	s.activated = true
	shotUpdateWindow.Call(hwnd)
	s.checkDisplays()
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
		if int32(r) == -1 {
			s.finish(image.Rectangle{}, fmt.Errorf("selection GetMessage: %v", e))
			break
		}
		if r == 0 {
			s.finish(image.Rectangle{}, nil)
			break
		}
		shotTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		shotDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
	return s.selection, s.err
}

func (s *screenshotOverlay) finish(r image.Rectangle, err error) {
	if s.done {
		return
	}
	s.done, s.finishing = true, true
	s.selection, s.err = r, err
}

func (s *screenshotOverlay) close() error {
	s.finishing = true
	var err error
	if s.hwnd == 0 {
		return nil
	}
	if s.timer != 0 {
		if r, _, e := shotKillTimer.Call(s.hwnd, s.timer); r == 0 {
			err = fmt.Errorf("kill selection timer: %v", e)
		}
		s.timer = 0
	}
	if h, _, _ := shotGetCapture.Call(); h == s.hwnd {
		if r, _, e := shotReleaseCapture.Call(); r == 0 {
			err = errors.Join(err, fmt.Errorf("release selection mouse: %v", e))
		}
	}
	hwnd := s.hwnd
	if r, _, e := shotDestroyWindow.Call(hwnd); r == 0 {
		err = errors.Join(err, fmt.Errorf("destroy selection window: %v", e))
	}
	return err
}

func (s *screenshotOverlay) cursor() (image.Point, bool) {
	var p shotPoint
	r, _, err := shotCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if r == 0 {
		s.finish(image.Rectangle{}, fmt.Errorf("selection cursor: %v", err))
		return image.Point{}, false
	}
	return p.image(), true
}

func (s *screenshotOverlay) checkDisplays() {
	monitors, err := screenshotMonitors()
	if err != nil {
		s.finish(image.Rectangle{}, err)
		return
	}
	if !slices.Equal(monitors, s.monitors) {
		s.finish(image.Rectangle{}, errors.New("display configuration changed; retry screenshot"))
	}
}

func (s *screenshotOverlay) windowSelection() {
	p, ok := s.cursor()
	if !ok {
		return
	}
	// shortcut: live window bounds can differ from frozen pixels; snapshot the window tree if this must be eliminated.
	s.finishing = true
	shotShowWindow.Call(s.hwnd, 0)
	packed := uintptr(uint64(uint32(p.X)) | uint64(uint32(p.Y))<<32)
	hwnd, _, _ := shotWindowAt.Call(packed)
	hwnd, _, _ = shotAncestor.Call(hwnd, 2) // GA_ROOT
	if hwnd == 0 {
		s.finish(image.Rectangle{}, nil)
		return
	}
	var rect shotRect
	if shotDWM != nil {
		hr, _, _ := shotDWM.Call(hwnd, 9, uintptr(unsafe.Pointer(&rect)), unsafe.Sizeof(rect))
		if int32(hr) >= 0 && rect.Right > rect.Left && rect.Bottom > rect.Top {
			if r := screenshotSelection(rect.image(), s.d.bounds, s.monitors); !r.Empty() {
				s.finish(r, nil)
				return
			}
		}
	}
	r, _, err := shotWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if r == 0 {
		s.finish(image.Rectangle{}, fmt.Errorf("selection GetWindowRect: %v", err))
		return
	}
	if rect.Right <= rect.Left || rect.Bottom <= rect.Top {
		s.finish(image.Rectangle{}, nil)
		return
	}
	s.finish(screenshotSelection(rect.image(), s.d.bounds, s.monitors), nil)
}

func screenshotWindowProc(hwnd, message, wparam, lparam uintptr) uintptr {
	if message == 0x81 { // WM_NCCREATE: lpCreateParams is the first CREATESTRUCT field.
		id := *(*uintptr)(native(lparam))
		if v, ok := shotCreating.Load(id); ok {
			s := v.(*screenshotOverlay)
			s.hwnd = hwnd
			shotWindows.Store(hwnd, s)
		}
	}
	value, ok := shotWindows.Load(hwnd)
	if !ok {
		r, _, _ := shotDefWindow.Call(hwnd, message, wparam, lparam)
		return r
	}
	s := value.(*screenshotOverlay)
	if message == 0x82 { // WM_NCDESTROY
		shotWindows.Delete(hwnd)
		s.hwnd = 0
		if !s.finishing {
			s.finish(image.Rectangle{}, nil)
		}
	}
	switch message {
	case 0xf: // WM_PAINT, including paints synchronously sent by ShowWindow.
		s.paint()
		return 0
	case 0x14: // WM_ERASEBKGND
		return 1
	}
	if !s.done && !s.finishing {
		switch message {
		case 0x201: // WM_LBUTTONDOWN
			if p, ok := s.cursor(); ok {
				s.start, s.dragging = p, true
				shotSetCapture.Call(hwnd)
				if h, _, _ := shotGetCapture.Call(); h != hwnd {
					s.finish(image.Rectangle{}, errors.New("capture selection mouse failed"))
				}
			}
		case 0x200: // WM_MOUSEMOVE
			if s.dragging {
				if p, ok := s.cursor(); ok {
					s.selection = image.Rectangle{Min: s.start, Max: p}.Canon()
					if r, _, e := shotInvalidate.Call(hwnd, 0, 0); r == 0 {
						s.finish(image.Rectangle{}, fmt.Errorf("invalidate selection: %v", e))
					}
				}
			}
		case 0x202: // WM_LBUTTONUP
			if s.dragging {
				if p, ok := s.cursor(); ok {
					s.finish(screenshotSelection(image.Rectangle{Min: s.start, Max: p}.Canon(), s.d.bounds, s.monitors), nil)
				}
			}
		case 0x100: // WM_KEYDOWN
			if wparam == 0x1b {
				s.finish(image.Rectangle{}, nil)
			}
			if wparam == 0x20 && lparam&(1<<30) == 0 {
				s.windowSelection()
			}
		case 0x10, 0x1f: // WM_CLOSE / WM_CANCELMODE
			s.finish(image.Rectangle{}, nil)
		case 0x8: // WM_KILLFOCUS
			if s.activated {
				s.finish(image.Rectangle{}, nil)
			}
		case 0x6: // WM_ACTIVATE
			if s.activated && wparam&0xffff == 0 {
				s.finish(image.Rectangle{}, nil)
			}
		case 0x215: // WM_CAPTURECHANGED
			if s.dragging && lparam != hwnd {
				s.finish(image.Rectangle{}, nil)
			}
		case 0x7e, 0x2e0: // WM_DISPLAYCHANGE / WM_DPICHANGED
			s.finish(image.Rectangle{}, errors.New("display configuration changed; retry screenshot"))
		case 0x113: // WM_TIMER
			if err := s.ctx.Err(); err != nil {
				s.finish(image.Rectangle{}, err)
			} else {
				s.checkDisplays()
			}
		}
	}
	// Never let DefWindowProc close a window or beep on handled input.
	switch message {
	case 0x10, 0x100, 0x201, 0x202, 0x200:
		return 0
	}
	r, _, _ := shotDefWindow.Call(hwnd, message, wparam, lparam)
	return r
}

func (s *screenshotOverlay) paint() {
	var ps struct {
		DC              uintptr
		Erase           int32
		Rect            shotRect
		Restore, Update int32
		Reserved        [32]byte
	}
	dc, _, e := shotBeginPaint.Call(s.hwnd, uintptr(unsafe.Pointer(&ps)))
	if dc == 0 {
		s.finish(image.Rectangle{}, fmt.Errorf("paint selection BeginPaint: %v", e))
		return
	}
	defer func() {
		if r, _, e := shotEndPaint.Call(s.hwnd, uintptr(unsafe.Pointer(&ps))); r == 0 {
			s.finish(image.Rectangle{}, fmt.Errorf("paint selection EndPaint: %v", e))
		}
	}()
	r := ps.Rect.image()
	if !r.Empty() {
		if ok, _, e := shotBitBlt.Call(dc, uintptr(r.Min.X), uintptr(r.Min.Y), uintptr(r.Dx()), uintptr(r.Dy()), s.d.dc, uintptr(r.Min.X), uintptr(r.Min.Y), 0xcc0020); ok == 0 {
			s.finish(image.Rectangle{}, fmt.Errorf("paint selection BitBlt: %v", e))
			return
		}
	}
	if !s.selection.Empty() {
		r := s.selection.Sub(s.d.bounds.Min)
		rect := shotRect{int32(r.Min.X), int32(r.Min.Y), int32(r.Max.X), int32(r.Max.Y)}
		brush, _, _ := shotStockObject.Call(18) // DC_BRUSH, owned by Windows.
		previous, _, _ := shotBrushColor.Call(dc, 0x0000ff)
		defer shotBrushColor.Call(dc, previous)
		if ok, _, e := shotFrameRect.Call(dc, uintptr(unsafe.Pointer(&rect)), brush); ok == 0 {
			s.finish(image.Rectangle{}, fmt.Errorf("paint selection FrameRect: %v", e))
		}
	}
}
