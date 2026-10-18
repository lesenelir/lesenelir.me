package screenshot

import (
	"image"
	"image/color"
	"testing"
)

func TestComposeMixedScaleAndNegativeOrigin(t *testing.T) {
	a, b := image.NewNRGBA(image.Rect(0, 0, 200, 100)), image.NewNRGBA(image.Rect(0, 0, 200, 200))
	for i := 0; i < len(a.Pix); i += 4 {
		a.Pix[i], a.Pix[i+3] = 255, 255
	}
	for i := 0; i < len(b.Pix); i += 4 {
		b.Pix[i+2], b.Pix[i+3] = 255, 255
	}
	shots := []displaySnapshot{{image.Rect(-200, 0, 0, 100), a}, {image.Rect(20, 0, 120, 100), b}}
	got, err := composeSelection(image.Rect(-50, 0, 70, 50), shots)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds() != image.Rect(0, 0, 240, 100) {
		t.Fatal(got.Bounds())
	}
	for _, c := range []struct {
		x    int
		want color.NRGBA
	}{{1, color.NRGBA{R: 255, A: 255}}, {110, color.NRGBA{A: 255}}, {200, color.NRGBA{B: 255, A: 255}}} {
		if got.NRGBAAt(c.x, 20) != c.want {
			t.Fatalf("pixel %d: %v", c.x, got.NRGBAAt(c.x, 20))
		}
	}
	if got, err := composeSelection(image.Rectangle{}, shots); got != nil || err != nil {
		t.Fatal("empty selection")
	}
	if got, err := composeSelection(image.Rect(1, 0, 10, 10), shots); got != nil || err != nil {
		t.Fatal("display gap selection")
	}
}
