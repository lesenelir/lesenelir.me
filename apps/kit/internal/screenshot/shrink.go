package screenshot

import (
	"image"
	"image/color"
)

// Shrink scales img down to fit in w×h pixels, each new pixel the average
// of those it covers, for a sharp thumbnail of a large screenshot. An image
// that fits already is returned as it is.
func Shrink(img image.Image, w, h int) image.Image {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw <= w && sh <= h || sw == 0 || sh == 0 {
		return img
	}
	scale := min(float64(w)/float64(sw), float64(h)/float64(sh))
	dw, dh := max(1, int(float64(sw)*scale+0.5)), max(1, int(float64(sh)*scale+0.5))
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	pixel := pixelReader(img)
	for dy := range dh {
		y0, y1 := b.Min.Y+dy*sh/dh, b.Min.Y+(dy+1)*sh/dh
		for dx := range dw {
			x0, x1 := b.Min.X+dx*sw/dw, b.Min.X+(dx+1)*sw/dw
			var r, g, bl, a, n uint64
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					pr, pg, pb, pa := pixel(x, y)
					r, g, bl, a, n = r+pr, g+pg, bl+pb, a+pa, n+1
				}
			}
			i := dst.PixOffset(dx, dy)
			dst.Pix[i+0] = uint8(r / n)
			dst.Pix[i+1] = uint8(g / n)
			dst.Pix[i+2] = uint8(bl / n)
			dst.Pix[i+3] = uint8(a / n)
		}
	}
	return dst
}

// pixelReader reads a pixel as premultiplied 8-bit RGBA, straight from the
// pixels of the types PNG decodes screenshots to, which At would box.
func pixelReader(img image.Image) func(x, y int) (r, g, b, a uint64) {
	switch m := img.(type) {
	case *image.RGBA:
		return func(x, y int) (r, g, b, a uint64) {
			p := m.Pix[m.PixOffset(x, y):]
			return uint64(p[0]), uint64(p[1]), uint64(p[2]), uint64(p[3])
		}
	case *image.NRGBA:
		return func(x, y int) (r, g, b, a uint64) {
			p := m.Pix[m.PixOffset(x, y):]
			a = uint64(p[3])
			return uint64(p[0]) * a / 255, uint64(p[1]) * a / 255, uint64(p[2]) * a / 255, a
		}
	}
	return func(x, y int) (r, g, b, a uint64) {
		c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
		return uint64(c.R), uint64(c.G), uint64(c.B), uint64(c.A)
	}
}
