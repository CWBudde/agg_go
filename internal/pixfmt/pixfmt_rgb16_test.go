package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// ==============================================================================
// RGB48 (16-bit per channel) Tests
// ==============================================================================

func TestPixFmtRGB48Linear(t *testing.T) {
	// Create a 4x4 RGB48 buffer (48 bytes total: 4*4*3*2)
	width, height := 4, 4
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)

	// Create RGB48 pixel format
	pixfmt := NewPixFmtRGB48Linear(rbuf)

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width() failed: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.Height() != height {
		t.Errorf("Height() failed: got %d, want %d", pixfmt.Height(), height)
	}
	if pixfmt.PixWidth() != 6 {
		t.Errorf("PixWidth() failed: got %d, want 6", pixfmt.PixWidth())
	}

	// Test pixel operations with 16-bit values
	red := color.RGB16Linear{R: 65535, G: 0, B: 0}   // Full red
	green := color.RGB16Linear{R: 0, G: 65535, B: 0} // Full green
	blue := color.RGB16Linear{R: 0, G: 0, B: 65535}  // Full blue

	// Test CopyPixel
	pixfmt.CopyPixel(0, 0, red)
	pixfmt.CopyPixel(1, 0, green)
	pixfmt.CopyPixel(2, 0, blue)

	// Test GetPixel
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 65535 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("GetPixel(0,0) failed: got {%d, %d, %d}, want {65535, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = pixfmt.GetPixel(1, 0)
	if pixel.R != 0 || pixel.G != 65535 || pixel.B != 0 {
		t.Errorf("GetPixel(1,0) failed: got {%d, %d, %d}, want {0, 65535, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = pixfmt.GetPixel(2, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 65535 {
		t.Errorf("GetPixel(2,0) failed: got {%d, %d, %d}, want {0, 0, 65535}", pixel.R, pixel.G, pixel.B)
	}

	// Test bounds checking
	pixel = pixfmt.GetPixel(-1, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("GetPixel(-1,0) bounds check failed: got {%d, %d, %d}, want {0, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	// Test blending with 16-bit values
	gray := color.RGB16Linear{R: 32768, G: 32768, B: 32768} // Mid gray
	pixfmt.CopyPixel(0, 1, gray)
	pixfmt.BlendPixel(0, 1, red, 32768, 65535) // Blend with half alpha

	pixel = pixfmt.GetPixel(0, 1)
	if pixel.R <= 32768 || pixel.R >= 65535 {
		t.Errorf("BlendPixel failed: red component %d should be between 32768 and 65535", pixel.R)
	}
}

func TestPixFmtRGB48CopyHline(t *testing.T) {
	width, height := 8, 2
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	c := color.RGB16Linear{R: 60000, G: 30000, B: 10000}
	pf.CopyHline(1, 0, 5, c)

	for x := 1; x <= 5; x++ {
		if got := pf.GetPixel(x, 0); got != c {
			t.Errorf("CopyHline x=%d: got %+v want %+v", x, got, c)
		}
	}
	if pf.GetPixel(0, 0).R != 0 {
		t.Error("pixel before CopyHline span should be zero")
	}
}

func TestPixFmtRGB48BlendHline(t *testing.T) {
	width, height := 4, 1
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	// fill with gray
	gray := color.RGB16Linear{R: 32768, G: 32768, B: 32768}
	pf.CopyHline(0, 0, width, gray)

	red := color.RGB16Linear{R: 65535, G: 0, B: 0}
	pf.BlendHline(0, 0, width, red, 32768, 65535)

	for x := 0; x < width; x++ {
		p := pf.GetPixel(x, 0)
		if p.R <= 32768 || p.R >= 65535 {
			t.Errorf("BlendHline x=%d: R=%d should be between 32768 and 65535", x, p.R)
		}
	}
}

func TestPixFmtRGB48CopyVline(t *testing.T) {
	width, height := 2, 8
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	c := color.RGB16Linear{R: 50000, G: 25000, B: 5000}
	pf.CopyVline(0, 2, 5, c)

	for y := 2; y <= 5; y++ {
		if got := pf.GetPixel(0, y); got != c {
			t.Errorf("CopyVline y=%d: got %+v want %+v", y, got, c)
		}
	}
}

func TestPixFmtRGB48BlendVline(t *testing.T) {
	width, height := 1, 4
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	gray := color.RGB16Linear{R: 32768, G: 32768, B: 32768}
	pf.CopyVline(0, 0, height-1, gray)

	blue := color.RGB16Linear{R: 0, G: 0, B: 65535}
	pf.BlendVline(0, 0, height, blue, 65535, 65535)

	for y := 0; y < height; y++ {
		p := pf.GetPixel(0, y)
		if p.B != 65535 {
			t.Errorf("BlendVline full alpha y=%d: B=%d want 65535", y, p.B)
		}
	}
}

func TestPixFmtRGB48CopyBar(t *testing.T) {
	width, height := 6, 6
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	c := color.RGB16Linear{R: 40000, G: 20000, B: 10000}
	pf.CopyBar(1, 1, 4, 4, c)

	for y := 1; y <= 4; y++ {
		for x := 1; x <= 4; x++ {
			if got := pf.GetPixel(x, y); got != c {
				t.Errorf("CopyBar (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}
	if pf.GetPixel(0, 0).R != 0 {
		t.Error("pixel outside CopyBar should be zero")
	}
}

func TestPixFmtRGB48BlendBar(t *testing.T) {
	width, height := 6, 6
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	gray := color.RGB16Linear{R: 32768, G: 32768, B: 32768}
	pf.CopyBar(0, 0, width-1, height-1, gray)

	red := color.RGB16Linear{R: 65535, G: 0, B: 0}
	pf.BlendBar(1, 1, 4, 4, red, 65535, 65535)

	for y := 1; y <= 4; y++ {
		for x := 1; x <= 4; x++ {
			got := pf.GetPixel(x, y)
			// 16-bit fixed-point blending may have ±1 rounding; allow tolerance
			if got.R < 65534 || got.G > 1 || got.B > 1 {
				t.Errorf("BlendBar full alpha (%d,%d): got %+v want approx %+v", x, y, got, red)
			}
		}
	}
}

func TestPixFmtRGB48BlendSolidHspan(t *testing.T) {
	width, height := 4, 1
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	gray := color.RGB16Linear{R: 32768, G: 32768, B: 32768}
	pf.CopyHline(0, 0, width, gray)

	red := color.RGB16Linear{R: 65535, G: 0, B: 0}
	covers := []basics.Int16u{65535, 32768, 16384, 0}
	pf.BlendSolidHspan(0, 0, 4, red, 65535, covers)

	// full cover → full red
	if pf.GetPixel(0, 0).R != 65535 {
		t.Errorf("BlendSolidHspan cover=65535: expected R=65535, got %d", pf.GetPixel(0, 0).R)
	}
	// zero cover → unchanged
	if pf.GetPixel(3, 0).R != 32768 {
		t.Errorf("BlendSolidHspan cover=0: expected R=32768, got %d", pf.GetPixel(3, 0).R)
	}

	// nil covers → uniform full coverage
	bufData2 := make([]basics.Int16u, 4*3)
	rbuf2 := buffer.NewRenderingBufferU16WithData(bufData2, 4, 1, 4*3*2)
	pf2 := NewPixFmtRGB48Linear(rbuf2)
	pf2.BlendSolidHspan(0, 0, 4, red, 65535, nil)
	for x := 0; x < 4; x++ {
		if pf2.GetPixel(x, 0).R != 65535 {
			t.Errorf("BlendSolidHspan nil covers x=%d: expected R=65535, got %d", x, pf2.GetPixel(x, 0).R)
		}
	}
}

func TestPixFmtRGB48BlendSolidVspan(t *testing.T) {
	width, height := 1, 4
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	green := color.RGB16Linear{R: 0, G: 65535, B: 0}
	pf.BlendSolidVspan(0, 0, 4, green, 65535, nil)

	for y := 0; y < 4; y++ {
		if pf.GetPixel(0, y).G != 65535 {
			t.Errorf("BlendSolidVspan nil covers y=%d: expected G=65535, got %d", y, pf.GetPixel(0, y).G)
		}
	}
}

func TestPixFmtRGB48CopyColorHspan(t *testing.T) {
	bufData := make([]basics.Int16u, 4*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 4, 1, 4*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	colors := []color.RGB16Linear{
		{R: 1000}, {R: 2000}, {R: 3000}, {R: 4000},
	}
	pf.CopyColorHspan(0, 0, 4, colors)

	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("CopyColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}
}

func TestPixFmtRGB48BlendColorHspan(t *testing.T) {
	bufData := make([]basics.Int16u, 4*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 4, 1, 4*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	colors := []color.RGB16Linear{
		{R: 65535}, {R: 30000}, {R: 10000}, {R: 50000},
	}
	pf.BlendColorHspan(0, 0, 4, colors, nil, 65535, 65535)

	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("BlendColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}
}

func TestPixFmtRGB48CopyColorVspan(t *testing.T) {
	bufData := make([]basics.Int16u, 4*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 4, 3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	colors := []color.RGB16Linear{
		{G: 1000}, {G: 2000}, {G: 3000}, {G: 4000},
	}
	pf.CopyColorVspan(0, 0, 4, colors)

	for y, want := range colors {
		if got := pf.GetPixel(0, y); got != want {
			t.Errorf("CopyColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtRGB48BlendColorVspan(t *testing.T) {
	bufData := make([]basics.Int16u, 4*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 4, 3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	colors := []color.RGB16Linear{
		{B: 65535}, {B: 30000}, {B: 10000}, {B: 5000},
	}
	pf.BlendColorVspan(0, 0, 4, colors, nil, 65535, 65535)

	for y, want := range colors {
		if got := pf.GetPixel(0, y); got != want {
			t.Errorf("BlendColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtRGB48FillAndClear(t *testing.T) {
	bufData := make([]basics.Int16u, 4*4*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 4, 4, 4*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)

	c := color.RGB16Linear{R: 12345, G: 23456, B: 34567}
	pf.Fill(c)

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if got := pf.GetPixel(x, y); got != c {
				t.Errorf("Fill (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}

	pf.Clear(color.RGB16Linear{})
	if pf.GetPixel(0, 0).R != 0 {
		t.Error("Clear should zero all pixels")
	}
}

func TestPixFmtRGB48PremultiplyNoOp(t *testing.T) {
	// Premultiply/Demultiply are no-ops for RGB formats
	bufData := make([]basics.Int16u, 1*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 1, 3*2)
	pf := NewPixFmtRGB48Linear(rbuf)
	c := color.RGB16Linear{R: 1000, G: 2000, B: 3000}
	pf.CopyPixel(0, 0, c)
	pf.Premultiply()
	pf.Demultiply()
	if got := pf.GetPixel(0, 0); got != c {
		t.Errorf("Premultiply/Demultiply should be no-ops for RGB: got %+v want %+v", got, c)
	}
}

func TestPixFmtRGB48ApplyGamma(t *testing.T) {
	bufData := make([]basics.Int16u, 1*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 1, 3*2)
	pf := NewPixFmtRGB48Linear(rbuf)
	c := color.RGB16Linear{R: 1000, G: 2000, B: 3000}
	pf.CopyPixel(0, 0, c)

	// double each channel
	double := func(v basics.Int16u) basics.Int16u { return v * 2 }
	pf.ApplyGammaDir(double)
	got := pf.GetPixel(0, 0)
	if got.R != 2000 || got.G != 4000 || got.B != 6000 {
		t.Errorf("ApplyGammaDir double: got %+v", got)
	}

	pf.ApplyGammaInv(func(v basics.Int16u) basics.Int16u { return v / 2 })
	got = pf.GetPixel(0, 0)
	if got.R != 1000 || got.G != 2000 || got.B != 3000 {
		t.Errorf("ApplyGammaInv half: got %+v", got)
	}
}

func TestPixFmtRGB48Pixel(t *testing.T) {
	bufData := make([]basics.Int16u, 1*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 1, 3*2)
	pf := NewPixFmtRGB48Linear(rbuf)
	c := color.RGB16Linear{R: 5000, G: 10000, B: 15000}
	pf.CopyPixel(0, 0, c)
	if got := pf.Pixel(0, 0); got != c {
		t.Errorf("Pixel alias: got %+v want %+v", got, c)
	}
}

func TestPixFmtRGB48Clear(t *testing.T) {
	bufData := make([]basics.Int16u, 2*2*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 2, 2, 2*3*2)
	pf := NewPixFmtRGB48Linear(rbuf)
	pf.CopyPixel(0, 0, color.RGB16Linear{R: 100, G: 200, B: 300})
	pf.Clear(color.RGB16Linear{})
	if got := pf.GetPixel(0, 0); got.R != 0 || got.G != 0 || got.B != 0 {
		t.Errorf("Clear did not zero pixel: %+v", got)
	}
}

func TestPixFmtRGB48Constructors(t *testing.T) {
	bufData := make([]basics.Int16u, 1*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, 1, 1, 3*2)
	_ = NewPixFmtBGR48Linear(rbuf)
	_ = NewPixFmtRGB48SRGB(rbuf)
	_ = NewPixFmtBGR48SRGB(rbuf)
	_ = NewPixFmtRGB48Pre(rbuf)
	_ = NewPixFmtBGR48Pre(rbuf)
	_ = NewPixFmtRGB48PreSRGB(rbuf)
	_ = NewPixFmtBGR48PreSRGB(rbuf)
}

func TestPixFmtRGB48CopyFromOverlappingVerticalRegion(t *testing.T) {
	width, height := 3, 4
	bufData := make([]basics.Int16u, width*height*3)
	rbuf := buffer.NewRenderingBufferU16WithData(bufData, width, height, width*3*2)
	pixfmt := NewPixFmtRGB48Linear(rbuf)

	rows := []color.RGB16Linear{
		{R: 1000, G: 0, B: 0},
		{R: 2000, G: 0, B: 0},
		{R: 3000, G: 0, B: 0},
		{R: 4000, G: 0, B: 0},
	}
	for y, c := range rows {
		for x := 0; x < width; x++ {
			pixfmt.CopyPixel(x, y, c)
		}
	}

	pixfmt.CopyFrom(pixfmt, 0, 0, 0, 1, width, 3)

	for x := 0; x < width; x++ {
		if got := pixfmt.GetPixel(x, 1); got.R != 1000 {
			t.Fatalf("row 1 pixel %d red = %d, want 1000", x, got.R)
		}
		if got := pixfmt.GetPixel(x, 2); got.R != 2000 {
			t.Fatalf("row 2 pixel %d red = %d, want 2000", x, got.R)
		}
		if got := pixfmt.GetPixel(x, 3); got.R != 3000 {
			t.Fatalf("row 3 pixel %d red = %d, want 3000", x, got.R)
		}
	}
}
