package agg

import (
	"bytes"
	"image"
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

// A premultiplied image is demultiplied on export, so ToGoImage and SaveToPNG
// show the straight colour instead of the darker premultiplied bytes.
func TestImageExportDemultipliesPremultipliedImage(t *testing.T) {
	img := NewImage([]byte{100, 50, 25, 128, 0, 255, 0, 255}, 2, 1, 8)
	if err := img.Premultiply(); err != nil {
		t.Fatal(err)
	}
	if got := img.AlphaMode(); got != AlphaPremultiplied {
		t.Fatalf("AlphaMode after Premultiply = %v, want AlphaPremultiplied", got)
	}
	if got, want := img.Data[:4], []byte{50, 25, 12, 128}; !bytes.Equal(got, want) {
		t.Fatalf("premultiplied bytes = %v, want %v", got, want)
	}
	// Same rounding as Image.Demultiply.
	want := []color.NRGBA{{99, 49, 23, 128}, {0, 255, 0, 255}}

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
	if got, want := img.Data[:4], []byte{50, 25, 12, 128}; !bytes.Equal(got, want) {
		t.Errorf("export modified the image: %v, want %v", got, want)
	}

	if err := img.Demultiply(); err != nil {
		t.Fatal(err)
	}
	if got := img.AlphaMode(); got != AlphaStraight {
		t.Fatalf("AlphaMode after Demultiply = %v, want AlphaStraight", got)
	}
	if got := img.ToGoImage().NRGBAAt(0, 0); got != want[0] {
		t.Errorf("ToGoImage after Demultiply = %v, want %v", got, want[0])
	}
}

// Images written through an API that declares the destination premultiplied
// remember that, so their export is demultiplied too.
func TestImageAlphaModeFollowsDeclaredDestination(t *testing.T) {
	src := NewImage([]byte{50, 25, 12, 128}, 1, 1, 4)
	dst := NewImage(make([]byte, 4), 1, 1, 4)
	err := CompositeImage(dst, src, Rect{X1: 0, Y1: 0, X2: 1, Y2: 1}, PointI{}, CompositeOptions{
		BlendMode: BlendSrc,
		Opacity:   1,
		AlphaMode: AlphaPremultiplied,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := dst.AlphaMode(); got != AlphaPremultiplied {
		t.Fatalf("CompositeImage destination AlphaMode = %v, want AlphaPremultiplied", got)
	}
	if got, want := dst.ToGoImage().NRGBAAt(0, 0), (color.NRGBA{99, 49, 23, 128}); got != want {
		t.Errorf("exported composite pixel = %v, want %v", got, want)
	}

	affineDst := NewImage(make([]byte, 4), 1, 1, 4)
	err = DrawImageAffine(affineDst, src, Rect{X1: 0, Y1: 0, X2: 1, Y2: 1}, NewTransformationsFromValues(1, 0, 0, 1, 0, 0), ImageTransformOptions{
		SourceAlpha:      AlphaPremultiplied,
		DestinationAlpha: AlphaPremultiplied,
		Opacity:          1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := affineDst.AlphaMode(); got != AlphaPremultiplied {
		t.Fatalf("DrawImageAffine destination AlphaMode = %v, want AlphaPremultiplied", got)
	}

	if err := dst.SetAlphaMode(AlphaStraight); err != nil {
		t.Fatal(err)
	}
	if got := dst.AlphaMode(); got != AlphaStraight {
		t.Fatalf("AlphaMode after SetAlphaMode = %v, want AlphaStraight", got)
	}
	if err := dst.SetAlphaMode(AlphaMode(7)); err == nil {
		t.Error("SetAlphaMode accepted an invalid mode")
	}
}

// Go's color.Color.RGBA() is premultiplied; importing must store straight
// bytes, so a translucent NRGBA image round-trips unchanged.
func TestNewImageFromStandardImageKeepsStraightAlpha(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	src.SetNRGBA(0, 0, color.NRGBA{200, 100, 50, 128})
	src.SetNRGBA(1, 0, color.NRGBA{10, 20, 30, 0})

	img, err := NewImageFromStandardImage(src)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := img.Data[:4], []byte{200, 100, 50, 128}; !bytes.Equal(got, want) {
		t.Errorf("imported bytes = %v, want %v", got, want)
	}
	if got, want := img.ToGoImage().NRGBAAt(0, 0), src.NRGBAAt(0, 0); got != want {
		t.Errorf("round trip = %v, want %v", got, want)
	}
}
