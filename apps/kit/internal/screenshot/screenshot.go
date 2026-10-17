//go:build darwin

// Package screenshot is Kit's screenshot tool. ⌘⇧2 runs macOS's own area
// and window capture; the screenshot goes to the clipboard, ready to paste
// into a chat, and shows as a thumbnail in the corner of the screen, which
// drags into any app. Screenshots are kept in Kit's data, never on the
// Desktop, for as long as the settings say.
package screenshot

import (
	"bytes"
	"context"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/kit/internal/app"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/platform/darwin"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/store"
	"github.com/lesenelir/lesenelir.me/apps/kit/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/kit/locales"
)

// section is the settings file's section for screenshots.
const section = "screenshot"

// Settings are what the user can change about screenshots.
type Settings struct {
	// KeepDays is how many days screenshots are kept; 0 keeps them until
	// the user deletes them.
	KeepDays int `json:"keepDays"`
}

func defaults() Settings { return Settings{KeepDays: 7} }

// keep is how long screenshots are kept, 0 for ever.
func (s Settings) keep() time.Duration {
	return time.Duration(max(0, s.KeepDays)) * 24 * time.Hour
}

// keepChoices are the choices of KeepDays, in the order the select shows
// them.
var keepChoices = []int{1, 7, 30, 0}

const (
	// thumbWidth and thumbHeight bound a thumbnail in pixels: twice the
	// size it shows at, for Retina displays.
	thumbWidth  = 2 * popupThumbWidth
	thumbHeight = 2 * popupThumbHeight
	// pruneEvery is how often screenshots past their time are deleted.
	pruneEvery = time.Hour
)

// Screenshot is the screenshot tool, a feature of the app.
type Screenshot struct {
	host     *app.Host
	settings *store.Section[Settings]
	lib      *Library
	keys     *shortcut
	popup    *mygo.Window
	busy     atomic.Bool // a capture runs
	load     chan string // thumbnails to make, by path

	// Read and written on the main thread: by the views and in
	// Window.Update.
	shots    []Shot // newest first
	thumbs   map[string]*ui.Bitmap
	loading  map[string]bool
	shown    int
	keepDays int // bound to its select
	confirm  bool
	err      error
	pop      popup
}

var _ interface {
	app.Feature
	app.SettingsShown
	app.PermissionRows
} = (*Screenshot)(nil)

// New returns the screenshot tool, to Add to the app.
func New() *Screenshot {
	s := &Screenshot{
		load:    make(chan string, 64),
		thumbs:  map[string]*ui.Bitmap{},
		loading: map[string]bool{},
	}
	s.keys = &shortcut{fire: s.take}
	return s
}

// Start reads the settings and the screenshots kept, deletes those past
// their time, makes the thumbnail window and takes ⌘⇧2.
func (s *Screenshot) Start(h *app.Host) error {
	s.host = h
	var err error
	s.settings, err = store.NewSection(h.Store, section, defaults(), nil)
	if err != nil {
		log.Printf("screenshot settings: %v; using defaults", err)
	}
	s.keepDays = s.settings.Get().KeepDays
	if s.lib, err = OpenLibrary(filepath.Join(h.DataDir, section)); err != nil {
		return err
	}
	s.reload()
	s.popup = s.newPopup()
	go s.loadThumbs()
	go func() {
		for {
			s.prune()
			time.Sleep(pruneEvery)
		}
	}()
	h.HandleKeys(s.keys)
	return nil
}

// SettingsShown lists the screenshots afresh, in case the files changed.
func (s *Screenshot) SettingsShown() { s.reload() }

// take captures a screenshot, unless one is being captured. It runs on the
// event tap's thread, so the capture runs on a goroutine of its own.
func (s *Screenshot) take() {
	if !s.busy.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer s.busy.Store(false)
		// Without the permission, macOS captures the desktop alone.
		if !darwin.ScreenCaptureAllowed() {
			darwin.RequestScreenCapture()
			s.host.ShowSettings("permissions")
			return
		}
		shot, thumb, err := s.capture()
		if err != nil {
			log.Printf("screenshot: %v", err)
		}
		s.host.Settings.Update(func() {
			s.err = err
			if thumb == nil {
				return
			}
			s.shots = slices.Insert(s.shots, 0, shot)
			s.thumbs[shot.Path] = thumb
			s.showPopup(shot, thumb)
		})
	}()
}

// capture takes a screenshot into the library and copies it. A cancelled
// capture returns no thumbnail and no error.
func (s *Screenshot) capture() (Shot, *ui.Bitmap, error) {
	path := s.lib.NewPath(time.Now())
	ok, err := capture(context.Background(), path)
	if !ok || err != nil {
		return Shot{}, nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Shot{}, nil, err
	}
	if err := copyPNG(data); err != nil {
		return Shot{}, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Shot{}, nil, err
	}
	shot := Shot{Path: path, Time: info.ModTime(), Size: info.Size()}
	thumb, err := thumbnail(data)
	if err != nil {
		return Shot{}, nil, err
	}
	return shot, thumb, nil
}

// copyPNG puts a PNG on the clipboard, as ⌃⌘⇧4 does.
func copyPNG(data []byte) error {
	return mygo.Clipboard.Write(transfer.New(transfer.NewItem(transfer.Bytes(transfer.PNG, data))))
}

// copyShot copies a kept screenshot.
func copyShot(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return copyPNG(data)
}

// fileData offers a screenshot's file to a drag: chats and documents take
// it as an image.
func fileData(path string) transfer.Data {
	d, err := transfer.FileData(path)
	if err != nil {
		log.Printf("screenshot drag: %v", err)
	}
	return d
}

// thumbnail decodes a screenshot into a small bitmap to show.
func thumbnail(data []byte) (*ui.Bitmap, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ui.NewBitmap(Shrink(img, thumbWidth, thumbHeight)), nil
}

// thumb returns a screenshot's thumbnail, nil until it is made.
func (s *Screenshot) thumb(path string) *ui.Bitmap {
	if b, ok := s.thumbs[path]; ok || s.loading[path] {
		return b
	}
	select {
	case s.load <- path:
		s.loading[path] = true
	default: // asked for again on a later frame
	}
	return nil
}

// loadThumbs makes the thumbnails the page asks for, one at a time, for
// a long history not to take every core.
func (s *Screenshot) loadThumbs() {
	for path := range s.load {
		var thumb *ui.Bitmap
		data, err := os.ReadFile(path)
		if err == nil {
			thumb, err = thumbnail(data)
		}
		if err != nil {
			log.Printf("screenshot thumbnail: %v", err)
		}
		s.host.Settings.Update(func() {
			delete(s.loading, path)
			s.thumbs[path] = thumb // nil for one that failed: not retried
		})
	}
}

// reload lists the screenshots from the disk, on the main thread.
func (s *Screenshot) reload() {
	shots, err := s.lib.List()
	if err != nil {
		s.err = err
		return
	}
	s.setShots(shots)
}

// setShots shows shots, forgetting the thumbnails of those gone.
func (s *Screenshot) setShots(shots []Shot) {
	s.shots = shots
	for path := range s.thumbs {
		if !slices.ContainsFunc(shots, func(x Shot) bool { return x.Path == path }) {
			delete(s.thumbs, path)
		}
	}
}

// prune deletes the screenshots past their time. It runs on any goroutine.
func (s *Screenshot) prune() {
	n, err := s.lib.Prune(s.settings.Get().keep(), time.Now())
	if err != nil {
		log.Printf("screenshot prune: %v", err)
	}
	if n == 0 {
		return
	}
	shots, err := s.lib.List()
	s.host.Settings.Update(func() {
		if err != nil {
			s.err = err
			return
		}
		s.setShots(shots)
	})
}

// setKeep keeps the choice of how long to keep screenshots, and deletes
// those past it at once.
func (s *Screenshot) setKeep() {
	days := s.keepDays
	s.err = s.settings.Update(func(x *Settings) { x.KeepDays = days })
	go s.prune()
}

// remove takes a screenshot off the list.
func (s *Screenshot) remove(path string) {
	s.shots = slices.DeleteFunc(s.shots, func(x Shot) bool { return x.Path == path })
	delete(s.thumbs, path)
}

// PermissionRows shows the Screen Recording permission, without which
// macOS captures the desktop but no window.
func (s *Screenshot) PermissionRows(c *ui.Context, m locales.Messages, pal widget.Palette, g *widget.Rows) {
	t := c.Theme()
	g.Row(c, m.PermScreenRecording(), m.PermScreenRecordingNote(), func() {
		if darwin.ScreenCaptureAllowed() {
			widget.Badge(c, m.StatusAllowed(), t.Success)
			return
		}
		widget.Badge(c, m.StatusNeeded(), t.Danger)
		if ui.Button(c.Key("screen-recording"), m.OpenSettings()).Clicked() {
			darwin.RequestScreenCapture()
			darwin.OpenScreenRecordingSettings()
		}
	})
}
