package agg

import (
	"image/color"
	"testing"
)

// The framebuffer stores straight (non-premultiplied) alpha, so the exported
// Go image must be straight as well. *image.RGBA would read R=255 at A=77 as
// an invalid premultiplied colour.
func TestImageExportKeepsStraightAlpha(t *testing.T) {
	img := NewImage([]byte{255, 0, 0, 77, 0, 255, 0, 255}, 2, 1, 8)
	want := []color.NRGBA{{255, 0, 0, 77}, {0, 255, 0, 255}}

	goImg := img.ToGoImage()
	for x, w := range want {
		if got := goImg.NRGBAAt(x, 0); got != w {
			t.Errorf("ToGoImage pixel %d = %v, want %v", x, got, w)
		}
	}

	std, err := img.ToStandardImage()
	if err != nil {
		t.Fatal(err)
	}
	for x, w := range want {
		if got := color.NRGBAModel.Convert(std.At(x, 0)); got != w {
			t.Errorf("ToStandardImage pixel %d = %v, want %v", x, got, w)
		}
	}
}

// A translucent comp-op result reaches the exported image unchanged: xor of a
// 0.7-alpha source over opaque red keeps the straight red at alpha 1-0.7.
func TestContextXorExportsStraightColor(t *testing.T) {
	ctx := NewContext(4, 4)
	ctx.Clear(Red)
	ctx.SetBlendMode(BlendXor)
	ctx.SetColor(RGBA(0, 0.3, 1, 0.7))
	ctx.FillRectangle(0, 0, 4, 4)

	goImg := ctx.GetImage().ToGoImage()
	if got, want := goImg.NRGBAAt(1, 1), (color.NRGBA{255, 0, 0, 77}); got != want {
		t.Fatalf("xor pixel = %v, want %v", got, want)
	}
}
