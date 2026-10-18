//go:build darwin || windows

package screenshot

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"

	"github.com/lesenelir/lesenelir.me/apps/elir/internal/widget"
	"github.com/lesenelir/lesenelir.me/apps/elir/locales"
)

// After a capture its thumbnail shows in the bottom right corner of the
// screen, as macOS's own does: drag it into an app, or click it to open
// the screenshot. It goes after a while, unless the pointer is on it.
const (
	popupThumbWidth  = 200
	popupThumbHeight = 140
	// popupMargin leaves room around the thumbnail for its shadow.
	popupMargin = 16
	// popupInset keeps the window off the screen's edges.
	popupInset = 8
	popupTime  = 5 * time.Second
)

var (
	popupBorder = ui.RGBA(255, 255, 255, 0.25)
	popupShadow = ui.RGBA(0, 0, 0, 0.35)
)

// popup is the thumbnail shown.
type popup struct {
	shot    Shot
	thumb   *ui.Bitmap
	w, h    float32 // the size it shows at, in DIPs
	hovered bool
	seq     int // counts the thumbnails shown, so a stale timer hides none
}

func (s *Screenshot) newPopup() *mygo.Window {
	return widget.NewOverlay(s.host.Tray, "Screenshot", popupThumbWidth+2*popupMargin, popupThumbHeight+2*popupMargin, s.popupView)
}

// fit returns the size a w×h pixel thumbnail shows at, within the popup's
// bounds and at most half its pixels, as the thumbnail is made for 2×.
func fit(w, h int) (float32, float32) {
	fw, fh := float32(w)/2, float32(h)/2
	scale := min(1, popupThumbWidth/fw, popupThumbHeight/fh)
	return max(1, fw*scale), max(1, fh*scale)
}

// showPopup shows a screenshot's thumbnail in the corner of the display the
// pointer is on, in place of the one shown. Main thread only.
func (s *Screenshot) showPopup(shot Shot, thumb *ui.Bitmap) {
	w, h := fit(thumb.Size())
	s.pop = popup{shot: shot, thumb: thumb, w: w, h: h, seq: s.pop.seq + 1}
	if s.popup == nil {
		return
	}
	ww, wh := int(w)+2*popupMargin, int(h)+2*popupMargin
	area := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint()).WorkArea
	s.popup.SetSize(ww, wh)
	s.popup.SetPosition(area.X+area.Width-ww-popupInset, area.Y+area.Height-wh-popupInset)
	s.popup.ShowInactive()
	s.popup.Invalidate()
	s.hideLater(s.pop.seq)
}

// hideLater hides the thumbnail after a while, or later while the pointer
// is on it.
func (s *Screenshot) hideLater(seq int) {
	time.AfterFunc(popupTime, func() {
		s.popup.Update(func() {
			switch {
			case s.pop.seq != seq:
			case s.pop.hovered:
				s.hideLater(seq)
			default:
				s.hidePopup()
			}
		})
	})
}

// hidePopup hides the thumbnail. Main thread only.
func (s *Screenshot) hidePopup() {
	s.pop = popup{seq: s.pop.seq}
	if s.popup != nil {
		s.popup.Hide()
	}
}

func (s *Screenshot) popupView(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	p := s.pop
	if p.thumb == nil {
		return
	}
	m := locales.In(c.Locale())
	ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		// The shadow is the card's, the rounded corners clip the image
		// inside it, as a clip would cut the shadow off.
		card := ui.Box(c.Key("thumb")).Size(p.w, p.h).Radius(8).Border(1, popupBorder).
			Shadow(0, 4, 14, 0, popupShadow).Cursor(ui.CursorPointer).Label(m.ScreenshotThumbnail()).
			Tooltip(m.ThumbnailHint())
		card.DragData(fileData(p.shot.Path), transfer.DragOptions{
			Operations: transfer.Copy,
			Done: func(r transfer.Result) {
				if !r.Canceled && r.Operation != transfer.None {
					s.hidePopup()
				}
			},
		})
		card.Children(func() {
			ui.Box(c).Fill().Radius(8).Clip().Children(func() { ui.Image(c, p.thumb).Fill().Fit(ui.Cover) })
		})
		s.pop.hovered = card.Hovered()
		if card.Clicked() {
			_ = mygo.Shell.OpenPath(p.shot.Path)
			s.hidePopup()
		}
	})
}
