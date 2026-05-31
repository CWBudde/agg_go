package agg2d

import (
	"image"
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
)

const imgFloatEps = 1e-6

func feq(a, b float32) bool {
	return math.Abs(float64(a)-float64(b)) <= imgFloatEps
}

func feqTol(a, b, tol float32) bool {
	return math.Abs(float64(a)-float64(b)) <= float64(tol)
}

func wantFloatPixel(t *testing.T, img *ImageFloat, x, y int, r, g, b, a float32) {
	t.Helper()
	c := img.GetPixel(x, y)
	if !feq(c.R, r) || !feq(c.G, g) || !feq(c.B, b) || !feq(c.A, a) {
		t.Fatalf("pixel(%d,%d) = {%v,%v,%v,%v}, want {%v,%v,%v,%v}",
			x, y, c.R, c.G, c.B, c.A, r, g, b, a)
	}
}

func TestImageFloatBasicProps(t *testing.T) {
	img := NewImageFloatEmpty(4, 3)
	if img.Width() != 4 || img.Height() != 3 {
		t.Fatalf("Width/Height = %d/%d, want 4/3", img.Width(), img.Height())
	}
	// Stride is in bytes: 4 pixels * 4 channels * 4 bytes = 64.
	if img.Stride() != 64 {
		t.Fatalf("Stride = %d, want 64", img.Stride())
	}
	if !img.IsAttached() {
		t.Fatal("IsAttached() = false, want true")
	}
	if len(img.Data) != 4*3*4 {
		t.Fatalf("len(Data) = %d, want %d", len(img.Data), 4*3*4)
	}
}

func TestImageFloatGetSetPixelRoundTrip(t *testing.T) {
	img := NewImageFloatEmpty(3, 2)
	img.SetPixel(1, 1, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8))
	wantFloatPixel(t, img, 1, 1, 0.2, 0.4, 0.6, 0.8)
	// neighbors untouched
	wantFloatPixel(t, img, 0, 1, 0, 0, 0, 0)
	wantFloatPixel(t, img, 2, 0, 0, 0, 0, 0)
}

func TestImageFloatGetPixelOutOfBounds(t *testing.T) {
	img := NewImageFloatEmpty(2, 2)
	// must not panic and return transparent black
	wantFloatPixel(t, img, -1, 0, 0, 0, 0, 0)
	wantFloatPixel(t, img, 0, 5, 0, 0, 0, 0)
	img.SetPixel(5, 5, color.NewRGBA32[color.Linear](1, 1, 1, 1)) // no-op, no panic
}

func TestImageFloatPremultiplyDemultiply(t *testing.T) {
	img := NewImageFloatEmpty(1, 1)
	img.SetPixel(0, 0, color.NewRGBA32[color.Linear](1.0, 0.5, 0.25, 0.5))

	img.Premultiply()
	// straight {1,0.5,0.25,0.5} -> premul {0.5,0.25,0.125,0.5}
	wantFloatPixel(t, img, 0, 0, 0.5, 0.25, 0.125, 0.5)

	img.Demultiply()
	// back to straight
	wantFloatPixel(t, img, 0, 0, 1.0, 0.5, 0.25, 0.5)
}

func TestImageFloatToFromNRGBA64RoundTrip(t *testing.T) {
	img := NewImageFloatEmpty(2, 1)
	img.SetPixel(0, 0, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8))
	img.SetPixel(1, 0, color.NewRGBA32[color.Linear](1.0, 0.0, 0.5, 1.0))

	nrgba := img.ToNRGBA64()
	if nrgba.Bounds().Dx() != 2 || nrgba.Bounds().Dy() != 1 {
		t.Fatalf("NRGBA64 bounds = %v, want 2x1", nrgba.Bounds())
	}
	// 16-bit straight: 0.2*65535 ~ 13107
	c := nrgba.NRGBA64At(0, 0)
	if c.R != uint16(math.Round(0.2*65535)) || c.A != uint16(math.Round(0.8*65535)) {
		t.Fatalf("NRGBA64At(0,0) = %+v, R/A unexpected", c)
	}

	back := NewImageFloatFromNRGBA64(nrgba)
	const tol = 1.0 / 65535.0
	bc := back.GetPixel(0, 0)
	if !feqTol(bc.R, 0.2, tol) || !feqTol(bc.G, 0.4, tol) || !feqTol(bc.B, 0.6, tol) || !feqTol(bc.A, 0.8, tol) {
		t.Fatalf("round-trip pixel(0,0) = %+v, want ~{0.2,0.4,0.6,0.8}", bc)
	}
}

func TestImageFloatToRGBAPremultiplies(t *testing.T) {
	img := NewImageFloatEmpty(1, 1)
	// straight color, 50% alpha
	img.SetPixel(0, 0, color.NewRGBA32[color.Linear](1.0, 0.5, 0.0, 0.5))

	rgba := img.ToRGBA()
	// Go's image.RGBA is alpha-premultiplied:
	// a8 = round(0.5*255) = 128
	// r8 = round(1.0*0.5*255) = 128, g8 = round(0.5*0.5*255) = 64, b8 = 0
	got := rgba.RGBAAt(0, 0)
	if absDiff(got.R, 128) > 1 || absDiff(got.G, 64) > 1 || absDiff(got.B, 0) > 1 || absDiff(got.A, 128) > 1 {
		t.Fatalf("RGBAAt(0,0) = %+v, want ~{128,64,0,128}", got)
	}
}

func TestImageFloatFromRGBADemultiplies(t *testing.T) {
	rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
	// premultiplied pixel {128,64,0,128}
	rgba.SetRGBA(0, 0, stdcolor.RGBA{R: 128, G: 64, B: 0, A: 128})

	img := NewImageFloatFromRGBA(rgba)
	c := img.GetPixel(0, 0)
	// straight: a = 128/255 ~ 0.502; r = (128/255)/(128/255) = 1.0; g = 64/128 = 0.5; b = 0
	const tol = 0.01
	if !feqTol(c.R, 1.0, tol) || !feqTol(c.G, 0.5, tol) || !feqTol(c.B, 0.0, tol) || !feqTol(c.A, 128.0/255.0, tol) {
		t.Fatalf("FromRGBA pixel(0,0) = %+v, want ~{1.0,0.5,0,0.502}", c)
	}
}

func TestImageFloatToFromImage8RoundTrip(t *testing.T) {
	img := NewImageFloatEmpty(2, 1)
	img.SetPixel(0, 0, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8))
	img.SetPixel(1, 0, color.NewRGBA32[color.Linear](1.0, 0.0, 1.0, 1.0))

	img8 := img.ToImage8()
	if img8.Width() != 2 || img8.Height() != 1 {
		t.Fatalf("Image8 dims = %dx%d, want 2x1", img8.Width(), img8.Height())
	}
	// straight 8-bit, no premultiply: 0.2*255 ~ 51
	px := img8.GetPixel(0, 0)
	if absDiff(px[0], 51) > 1 || absDiff(px[3], 204) > 1 {
		t.Fatalf("Image8 GetPixel(0,0) = %v, want ~{51,102,153,204}", px)
	}

	back := NewImageFloatFromImage8(img8)
	const tol = 1.0 / 255.0
	bc := back.GetPixel(0, 0)
	if !feqTol(bc.R, 0.2, tol) || !feqTol(bc.G, 0.4, tol) || !feqTol(bc.B, 0.6, tol) || !feqTol(bc.A, 0.8, tol) {
		t.Fatalf("round-trip pixel(0,0) = %+v, want ~{0.2,0.4,0.6,0.8}", bc)
	}
}

func absDiff(a, b uint8) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}
