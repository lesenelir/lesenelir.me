//go:build ignore

// Draws the menu bar icon, the head of the app icon's character: hair with
// bangs and a cowlick, the outline of the face, two eyes. It is a template
// image, black on transparent, which macOS tints for the menu bar:
//
//	go run design/icon/menubar.go internal/app/trayicon.png 32
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
)

// inside reports whether the point (x, y) of the 32×32 design is ink.
func inside(x, y float64) bool {
	const cx, cy, r, ring = 16, 18.5, 11.2, 2.3
	d := math.Hypot(x-cx, y-cy)
	// The cowlick: a curl above the head.
	if dc := math.Hypot(x-21, y-7.8); math.Abs(dc-4.6) < 1.15 && x < 21 && y < 7.8 {
		return true
	}
	if d > r {
		return false
	}
	// The hair: the top of the head, down to scalloped bangs.
	bangs := 13.6 + 3.6*math.Abs(math.Sin(math.Pi*(x-5)/7.35))
	if y < bangs {
		return true
	}
	// The face: an outline, and two eyes.
	if d > r-ring {
		return true
	}
	for _, ex := range []float64{12, 20} {
		if math.Hypot(x-ex, y-21.6) < 1.75 {
			return true
		}
	}
	return false
}

func main() {
	size, _ := strconv.Atoi(os.Args[2])
	scale := float64(size) / 32
	const ss = 8
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			n := 0
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					if inside((float64(px)+(float64(sx)+0.5)/ss)/scale, (float64(py)+(float64(sy)+0.5)/ss)/scale) {
						n++
					}
				}
			}
			img.SetNRGBA(px, py, color.NRGBA{A: uint8(255 * n / (ss * ss))})
		}
	}
	f, _ := os.Create(os.Args[1])
	png.Encode(f, img)
	f.Close()
}
