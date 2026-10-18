package screenshot

import (
	"errors"
	"image"
	"math"

	"golang.org/x/image/draw"
)

type displaySnapshot struct {
	bounds image.Rectangle
	image  image.Image
}

// composeSelection maps display points to pixels, retaining the highest source
// density across mixed-scale displays and black pixels in display gaps.
func composeSelection(area image.Rectangle, shots []displaySnapshot) (*image.NRGBA, error) {
	if area.Empty() {
		return nil, nil
	}
	scale := float64(1)
	covered := image.Rectangle{}
	for _, s := range shots {
		if s.bounds.Overlaps(area) {
			covered = covered.Union(s.bounds.Intersect(area))
			scale = math.Max(scale, float64(s.image.Bounds().Dx())/float64(s.bounds.Dx()))
		}
	}
	if covered.Empty() {
		return nil, nil
	}
	area = area.Intersect(covered)
	w, h := int(math.Ceil(float64(area.Dx())*scale)), int(math.Ceil(float64(area.Dy())*scale))
	const limit = 256 << 20
	if w <= 0 || h <= 0 || w > limit/4 || h > limit/(w*4) {
		return nil, errors.New("screenshot selection exceeds 256 MiB image budget")
	}
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 3; i < len(out.Pix); i += 4 {
		out.Pix[i] = 255
	}
	for _, s := range shots {
		r := area.Intersect(s.bounds)
		if r.Empty() {
			continue
		}
		src := s.image.Bounds()
		sx, sy := float64(src.Dx())/float64(s.bounds.Dx()), float64(src.Dy())/float64(s.bounds.Dy())
		a := image.Rect(src.Min.X+int(float64(r.Min.X-s.bounds.Min.X)*sx), src.Min.Y+int(float64(r.Min.Y-s.bounds.Min.Y)*sy), src.Min.X+int(float64(r.Max.X-s.bounds.Min.X)*sx), src.Min.Y+int(float64(r.Max.Y-s.bounds.Min.Y)*sy))
		d := image.Rect(int(float64(r.Min.X-area.Min.X)*scale), int(float64(r.Min.Y-area.Min.Y)*scale), int(math.Ceil(float64(r.Max.X-area.Min.X)*scale)), int(math.Ceil(float64(r.Max.Y-area.Min.Y)*scale)))
		draw.NearestNeighbor.Scale(out, d, s.image, a, draw.Src, nil)
	}
	return out, nil
}
