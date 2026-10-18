//go:build darwin

package screenshot

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/platform"
	"github.com/lesenelir/lesenelir.me/apps/elir/internal/shortcut"
)

// capture owns selection input; screencapture only takes non-interactive
// snapshots before any overlay is shown, so configured keys never inject input.
func capture(ctx context.Context, path string, b shortcut.Bindings, router *shortcut.Router) (bool, error) {
	displays := mygo.Screen.Displays()
	if len(displays) == 0 {
		return false, errors.New("screenshot: no displays")
	}
	windows, err := platform.ScreenshotWindows()
	if err != nil {
		return false, err
	}
	dir, err := os.MkdirTemp("", ".elir-selection-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(dir)
	var shots []displaySnapshot
	var pixels int64
	for i, d := range displays {
		r := image.Rect(d.Bounds.X, d.Bounds.Y, d.Bounds.X+d.Bounds.Width, d.Bounds.Y+d.Bounds.Height)
		file := filepath.Join(dir, fmt.Sprintf("display-%d.png", i))
		rect := fmt.Sprintf("%d,%d,%d,%d", r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		jobCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		output, err := exec.CommandContext(jobCtx, "/usr/sbin/screencapture", "-x", "-R"+rect, file).CombinedOutput()
		cancel()
		if err != nil {
			return false, fmt.Errorf("screenshot snapshot: %w: %s", err, output)
		}
		f, err := os.Open(file)
		if err != nil {
			return false, err
		}
		config, err := png.DecodeConfig(f)
		if err != nil {
			f.Close()
			return false, err
		}
		pixels += int64(config.Width) * int64(config.Height)
		if config.Width <= 0 || config.Height <= 0 || pixels > 64<<20 {
			f.Close()
			return false, errors.New("screenshot snapshots exceed 256 MiB budget")
		}
		if _, err = f.Seek(0, 0); err != nil {
			f.Close()
			return false, err
		}
		img, decodeErr := png.Decode(f)
		closeErr := f.Close()
		if err := errors.Join(decodeErr, closeErr); err != nil {
			return false, err
		}
		shots = append(shots, displaySnapshot{bounds: r, image: img})
	}
	if !slices.Equal(displays, mygo.Screen.Displays()) {
		return false, errors.New("display configuration changed; retry screenshot")
	}
	area, err := selectMacScreenshot(ctx, displays, shots, windows, b, router)
	if err != nil || area.Empty() {
		return false, err
	}
	img, err := composeSelection(area, shots)
	if err != nil || img == nil {
		return false, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return false, err
	}
	err = png.Encode(f, img)
	err = errors.Join(err, f.Close(), ctx.Err())
	if err != nil {
		os.Remove(path)
		return false, err
	}
	return true, nil
}

func selectMacScreenshot(ctx context.Context, displays []mygo.Display, shots []displaySnapshot, windows []image.Rectangle, b shortcut.Bindings, router *shortcut.Router) (area image.Rectangle, err error) {
	type result struct {
		area image.Rectangle
		err  error
	}
	done := make(chan result, 1)
	var overlays []*mygo.Window
	var start image.Point
	var selected image.Rectangle
	dragging, windowMode, finished, armed := false, false, false, false
	var selectionErr error
	finish := func(r image.Rectangle) {
		if !finished {
			finished = true
			done <- result{r, selectionErr}
		}
	}
	pointer := func() image.Point { p := mygo.Screen.CursorScreenPoint(); return image.Pt(p.X, p.Y) }
	target := func(p image.Point) image.Rectangle {
		for _, r := range windows {
			if p.In(r) {
				return r
			}
		}
		return image.Rectangle{}
	}
	redraw := func() {
		for _, w := range overlays {
			w.Invalidate()
		}
	}
	var offDisplays func()
	mygo.RunOnMain(func() {
		offDisplays = mygo.Screen.OnDisplaysChanged(func() {
			selectionErr = errors.New("display configuration changed; retry screenshot")
			finish(image.Rectangle{})
		})
		router.Selection = func(key string, m shortcut.Mods) bool {
			if b[shortcut.ScreenshotCancel].Match(key, m) {
				finish(image.Rectangle{})
				return true
			}
			if b[shortcut.Window].Match(key, m) {
				windowMode = !windowMode
				dragging = false
				selected = image.Rectangle{}
				if windowMode {
					selected = target(pointer())
				}
				redraw()
				return true
			}
			return false
		}
		for i, d := range displays {
			bounds := shots[i].bounds
			bitmap := ui.NewBitmap(shots[i].image)
			view := func(c *ui.Context) {
				ui.Box(c.Key("selection")).Fill().Cursor(ui.CursorCrosshair).HandleInput(func(e ui.InputEvent) bool {
					if finished {
						return true
					}
					p := image.Pt(bounds.Min.X+int(math.Round(float64(e.X))), bounds.Min.Y+int(math.Round(float64(e.Y))))
					switch e.Kind {
					case ui.InputPointerDown:
						if e.Button != 0 {
							finish(image.Rectangle{})
							return true
						}
						if windowMode {
							finish(target(p))
							return true
						}
						start = p
						dragging = true
						selected = image.Rectangle{}
						return true
					case ui.InputPointerMove:
						if windowMode {
							selected = target(p)
						} else if dragging {
							selected = image.Rectangle{Min: start, Max: p}.Canon()
						}
						redraw()
						return true
					case ui.InputPointerUp:
						if dragging {
							dragging = false
							r := image.Rectangle{Min: start, Max: p}.Canon()
							if r.Dx() > 1 && r.Dy() > 1 {
								finish(r)
							} else {
								selected = image.Rectangle{}
								redraw()
							}
						}
						return true
					}
					return false
				}).Children(func() {
					ui.Image(c, bitmap).Fill().Fit(ui.Contain).PassThrough()
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Background(ui.RGBA(0, 0, 0, 0.15)).PassThrough()
					r := selected.Intersect(bounds)
					if !r.Empty() {
						ui.Box(c).Absolute().Left(float32(r.Min.X-bounds.Min.X)).Top(float32(r.Min.Y-bounds.Min.Y)).Width(float32(r.Dx())).Height(float32(r.Dy())).Border(2, ui.Hex("#ffffff")).PassThrough()
					}
				})
			}
			w := mygo.NewWindow(mygo.WindowOptions{Title: "Elir Screenshot", Hidden: true, Panel: true, Frameless: true, DisableShadow: true, DisableResize: true, AlwaysOnTop: true, Width: d.Bounds.Width, Height: d.Bounds.Height, Content: ui.View(view)})
			w.SetBounds(d.Bounds)
			overlays = append(overlays, w)
			w.OnClose(func(*mygo.CloseEvent) { finish(image.Rectangle{}) })
			w.OnBlur(func() {
				w.Update(func() {
					if !armed || finished {
						return
					}
					for _, other := range overlays {
						if other.IsFocused() {
							return
						}
					}
					finish(image.Rectangle{})
				})
			})
		}
		for _, w := range overlays {
			w.ShowInactive()
		}
		p := pointer()
		for i, s := range shots {
			if p.In(s.bounds) {
				overlays[i].Focus()
				break
			}
		}
		armed = true
	})
	select {
	case result := <-done:
		area, err = result.area, result.err
	case <-ctx.Done():
		err = ctx.Err()
	}
	mygo.RunOnMain(func() {
		armed = false
		router.Selection = nil
		if offDisplays != nil {
			offDisplays()
		}
		for _, w := range overlays {
			w.Close()
		}
		overlays = nil
	})
	return area, err
}
