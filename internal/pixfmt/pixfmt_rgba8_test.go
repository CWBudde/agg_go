package pixfmt

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/simd"
)

type grayRowSource struct {
	row []basics.Int8u
}

func (s *grayRowSource) RowData(y int) []basics.Int8u {
	if y != 0 {
		return nil
	}
	return s.row
}

func (s *grayRowSource) Width() int {
	return len(s.row)
}

func (s *grayRowSource) Height() int {
	return 1
}

func TestRGBAPixelType(t *testing.T) {
	p := &RGBAPixelType{}
	p.Set(128, 64, 192, 255)

	if p.R != 128 || p.G != 64 || p.B != 192 || p.A != 255 {
		t.Errorf("Set failed: expected (128,64,192,255), got (%d,%d,%d,%d)", p.R, p.G, p.B, p.A)
	}

	c := p.GetColor()
	if c.R != 128 || c.G != 64 || c.B != 192 || c.A != 255 {
		t.Errorf("GetColor failed: expected (128,64,192,255), got (%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
}

func TestPixFmtRGBA32Basic(t *testing.T) {
	// Create a test buffer
	width, height := 100, 50
	buf := make([]basics.Int8u, width*height*4) // 4 bytes per pixel
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)

	// Create pixel format
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// Test basic properties
	if pf.Width() != width {
		t.Errorf("Width() expected %d, got %d", width, pf.Width())
	}
	if pf.Height() != height {
		t.Errorf("Height() expected %d, got %d", height, pf.Height())
	}
	if pf.PixWidth() != 4 {
		t.Errorf("PixWidth() expected 4, got %d", pf.PixWidth())
	}
}

func TestPixFmtRGBA32CopyPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// Test copy pixel
	rgba := color.NewRGBA8[color.Linear](128, 64, 192, 255)
	pf.CopyPixel(5, 5, rgba)

	// Check that pixel was set
	retrieved := pf.GetPixel(5, 5)
	if retrieved.R != 128 || retrieved.G != 64 || retrieved.B != 192 || retrieved.A != 255 {
		t.Errorf("CopyPixel failed: expected (128,64,192,255), got (%d,%d,%d,%d)",
			retrieved.R, retrieved.G, retrieved.B, retrieved.A)
	}
}

func TestPixFmtRGBA32BlendPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// Set background
	bg := color.NewRGBA8[color.Linear](100, 100, 100, 255)
	pf.CopyPixel(5, 5, bg)

	// Blend with another color
	blend := color.NewRGBA8[color.Linear](200, 150, 50, 128) // 50% alpha
	pf.BlendPixel(5, 5, blend, 255)                          // Full coverage

	// Result should be between background and blend colors
	result := pf.GetPixel(5, 5)
	if result.R <= 100 || result.R >= 200 {
		t.Errorf("BlendPixel red failed: expected between 100-200, got %d", result.R)
	}
	if result.G <= 100 || result.G >= 150 {
		t.Errorf("BlendPixel green failed: expected between 100-150, got %d", result.G)
	}
	if result.B >= 100 { // Should be darker since blend is 50
		t.Errorf("BlendPixel blue failed: expected less than 100, got %d", result.B)
	}
}

func TestPixFmtRGBA32Lines(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	rgba := color.NewRGBA8[color.Linear](128, 64, 192, 255)

	// Test horizontal line
	pf.CopyHline(5, 10, 15, rgba)
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if pixel.R != 128 || pixel.G != 64 || pixel.B != 192 || pixel.A != 255 {
			t.Errorf("CopyHline failed at (%d, 10): expected (128,64,192,255), got (%d,%d,%d,%d)",
				x, pixel.R, pixel.G, pixel.B, pixel.A)
		}
	}

	// Test vertical line
	pf.CopyVline(10, 5, 15, rgba)
	for y := 5; y <= 15; y++ {
		pixel := pf.GetPixel(10, y)
		if pixel.R != 128 || pixel.G != 64 || pixel.B != 192 || pixel.A != 255 {
			t.Errorf("CopyVline failed at (10, %d): expected (128,64,192,255), got (%d,%d,%d,%d)",
				y, pixel.R, pixel.G, pixel.B, pixel.A)
		}
	}
}

func TestPixFmtRGBA32Rectangle(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	rgba := color.NewRGBA8[color.Linear](200, 100, 50, 255)

	// Test filled rectangle
	pf.CopyBar(5, 5, 10, 10, rgba)

	// Check corners
	corners := [][2]int{{5, 5}, {10, 5}, {5, 10}, {10, 10}}
	for _, corner := range corners {
		pixel := pf.GetPixel(corner[0], corner[1])
		if pixel.R != 200 || pixel.G != 100 || pixel.B != 50 || pixel.A != 255 {
			t.Errorf("CopyBar failed at (%d, %d): expected (200,100,50,255), got (%d,%d,%d,%d)",
				corner[0], corner[1], pixel.R, pixel.G, pixel.B, pixel.A)
		}
	}

	// Check that pixels outside rectangle are unchanged (should be 0)
	pixel := pf.GetPixel(4, 4)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 || pixel.A != 0 {
		t.Errorf("CopyBar affected pixel outside rectangle at (4, 4): got (%d,%d,%d,%d)",
			pixel.R, pixel.G, pixel.B, pixel.A)
	}
}

func TestPixFmtRGBA32Spans(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	rgba := color.NewRGBA8[color.Linear](150, 200, 100, 255)

	// Test horizontal span with varying coverage
	covers := []basics.Int8u{255, 200, 150, 100, 50}
	pf.BlendSolidHspan(5, 10, len(covers), rgba, covers)

	// Check that coverage affects the result
	for i, cover := range covers {
		pixel := pf.GetPixel(5+i, 10)
		// With higher coverage, we should get values closer to source color
		if cover == 255 && pixel.R != 150 {
			t.Errorf("BlendSolidHspan with full coverage should give R=150, got R=%d", pixel.R)
		}
		if cover == 50 && pixel.R >= 150 {
			t.Errorf("BlendSolidHspan with low coverage should give R<150, got R=%d", pixel.R)
		}
	}
}

func TestPixFmtRGBA32Clear(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height*4)
	// Initialize buffer with non-zero values
	for i := range buf {
		buf[i] = 100
	}

	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// Clear with a specific color
	clearColor := color.NewRGBA8[color.Linear](50, 100, 150, 200)
	pf.Clear(clearColor)

	// Check that all pixels are now the clear color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pf.GetPixel(x, y)
			if pixel.R != 50 || pixel.G != 100 || pixel.B != 150 || pixel.A != 200 {
				t.Errorf("Clear failed at (%d, %d): expected (50,100,150,200), got (%d,%d,%d,%d)",
					x, y, pixel.R, pixel.G, pixel.B, pixel.A)
			}
		}
	}
}

func TestPixFmtRGBA32CopyHlineSIMDDispatch(t *testing.T) {
	t.Cleanup(simd.ResetDetection)
	simd.SetForcedFeatures(simd.Features{
		Architecture: runtimeGOARCHForSIMDTest(),
		HasAVX2:      runtimeGOARCHForSIMDTest() == "amd64",
		HasSSE2:      runtimeGOARCHForSIMDTest() == "amd64",
		HasNEON:      runtimeGOARCHForSIMDTest() == "arm64",
	})

	width, height := 8, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](9, 8, 7, 6)
	pf.CopyHline(0, 0, width, c)

	for i := 0; i < width; i++ {
		p := i * 4
		if got := buf[p : p+4]; got[0] != 9 || got[1] != 8 || got[2] != 7 || got[3] != 6 {
			t.Fatalf("pixel %d = %v, want [9 8 7 6]", i, got)
		}
	}
}

func runtimeGOARCHForSIMDTest() string {
	return simd.DetectFeatures().Architecture
}

func TestPixFmtRGBA32Bounds(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	rgba := color.NewRGBA8[color.Linear](128, 128, 128, 255)

	// Test out-of-bounds operations (should not crash)
	pf.CopyPixel(-1, -1, rgba)
	pf.CopyPixel(width, height, rgba)
	pf.BlendPixel(-1, -1, rgba, 255)
	pf.BlendPixel(width, height, rgba, 255)

	// These operations should be safe and not affect valid pixels
	pf.CopyPixel(0, 0, rgba)
	pixel := pf.GetPixel(0, 0)
	if pixel.R != 128 || pixel.G != 128 || pixel.B != 128 || pixel.A != 255 {
		t.Error("Valid pixel operation failed after out-of-bounds tests")
	}
}

func TestPixFmtRGBA32ConcreteTypes(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)

	// Test that all concrete types can be created
	_ = NewPixFmtRGBA32[color.Linear](rbuf)
	_ = NewPixFmtARGB32[color.Linear](rbuf)
	_ = NewPixFmtBGRA32[color.Linear](rbuf)
	_ = NewPixFmtABGR32[color.Linear](rbuf)

	// These should compile without errors
}

func TestPixFmtRGBA32CopyFromOverlapSameRow(t *testing.T) {
	width, height := 5, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	srcColors := []color.RGBA8[color.Linear]{
		{R: 1, A: 255},
		{R: 2, A: 255},
		{R: 3, A: 255},
		{R: 4, A: 255},
		{R: 5, A: 255},
	}
	for x, c := range srcColors {
		pf.CopyPixel(x, 0, c)
	}

	pf.CopyFrom(pf, 1, 0, 0, 0, 4)

	for x, want := range []basics.Int8u{1, 1, 2, 3, 4} {
		if got := pf.GetPixel(x, 0).R; got != want {
			t.Fatalf("CopyFrom overlap mismatch at x=%d: got %d want %d", x, got, want)
		}
	}
}

func TestPixFmtRGBA32BlendFromOverlapSameRow(t *testing.T) {
	width, height := 5, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	srcColors := []color.RGBA8[color.Linear]{
		{R: 1, A: 255},
		{R: 2, A: 255},
		{R: 3, A: 255},
		{R: 4, A: 255},
		{R: 5, A: 255},
	}
	for x, c := range srcColors {
		pf.CopyPixel(x, 0, c)
	}

	pf.BlendFrom(pf, 1, 0, 0, 0, 4, basics.CoverFull)

	for x, want := range []basics.Int8u{1, 1, 2, 3, 4} {
		if got := pf.GetPixel(x, 0).R; got != want {
			t.Fatalf("BlendFrom overlap mismatch at x=%d: got %d want %d", x, got, want)
		}
	}
}

func TestPixFmtRGBA32BlendFromColor(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)
	src := &grayRowSource{row: []basics.Int8u{0, 128, 255}}
	c := color.NewRGBA8[color.Linear](200, 40, 20, 255)

	pf.BlendFromColor(src, c, 0, 0, 0, 0, 3, basics.CoverFull)

	if got := pf.GetPixel(0, 0); got.R != 0 || got.G != 0 || got.B != 0 || got.A != 0 {
		t.Fatalf("expected zero-coverage pixel to stay empty, got %+v", got)
	}
	mid := pf.GetPixel(1, 0)
	if mid.R == 0 || mid.R >= c.R {
		t.Fatalf("expected partial blend at x=1, got %+v", mid)
	}
	if got := pf.GetPixel(2, 0); got != c {
		t.Fatalf("expected full-coverage pixel to match source color, got %+v want %+v", got, c)
	}
}

func TestRGBAPixelTypeSetColor(t *testing.T) {
	p := &RGBAPixelType{}
	c := color.NewRGBA8[color.Linear](10, 20, 30, 40)
	p.SetColor(c)
	if p.R != 10 || p.G != 20 || p.B != 30 || p.A != 40 {
		t.Errorf("SetColor failed: got (%d,%d,%d,%d)", p.R, p.G, p.B, p.A)
	}
}

func TestPixFmtRGBA32Pixel(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)
	c := color.NewRGBA8[color.Linear](7, 8, 9, 255)
	pf.CopyPixel(0, 0, c)
	if got := pf.Pixel(0, 0); got != c {
		t.Errorf("Pixel alias mismatch: got %+v want %+v", got, c)
	}
}

func TestPixFmtRGBA32BlendHlineOpaqueFull(t *testing.T) {
	// opaque + CoverFull → direct copy fast path
	width, height := 8, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](200, 100, 50, 255)
	pf.BlendHline(1, 0, 5, c, basics.CoverFull)

	for x := 1; x < 6; x++ {
		if got := pf.GetPixel(x, 0); got != c {
			t.Errorf("BlendHline opaque at x=%d: got %+v want %+v", x, got, c)
		}
	}
	// pixels outside the span unchanged
	if got := pf.GetPixel(0, 0); got.R != 0 {
		t.Errorf("pixel before span should be zero")
	}
}

func TestPixFmtRGBA32BlendHlinePartialAlpha(t *testing.T) {
	width, height := 4, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// pre-fill with opaque white
	white := color.NewRGBA8[color.Linear](255, 255, 255, 255)
	pf.CopyHline(0, 0, width, white)

	src := color.NewRGBA8[color.Linear](0, 0, 0, 128) // semi-transparent black
	pf.BlendHline(0, 0, width, src, basics.CoverFull)

	for x := 0; x < width; x++ {
		p := pf.GetPixel(x, 0)
		if p.R >= 255 || p.R == 0 {
			t.Errorf("BlendHline partial alpha at x=%d: red=%d should be between 0 and 255", x, p.R)
		}
	}
}

func TestPixFmtRGBA32BlendHlineTransparent(t *testing.T) {
	// transparent color → no-op
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 4, 1, 4*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)
	transparent := color.NewRGBA8[color.Linear](255, 0, 0, 0)
	pf.BlendHline(0, 0, 4, transparent, 255)
	for x := 0; x < 4; x++ {
		if got := pf.GetPixel(x, 0); got.R != 0 {
			t.Errorf("transparent BlendHline should not modify pixel at x=%d", x)
		}
	}
}

func TestPixFmtRGBA32BlendVline(t *testing.T) {
	width, height := 1, 8
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](0, 255, 0, 255)
	pf.BlendVline(0, 2, 4, c, basics.CoverFull)

	for y := 2; y <= 4; y++ {
		if got := pf.GetPixel(0, y); got != c {
			t.Errorf("BlendVline at y=%d: got %+v want %+v", y, got, c)
		}
	}
	// before span
	if got := pf.GetPixel(0, 1); got.G != 0 {
		t.Errorf("pixel before BlendVline span should be zero")
	}
}

func TestPixFmtRGBA32BlendBar(t *testing.T) {
	width, height := 8, 8
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// pre-fill white
	white := color.NewRGBA8[color.Linear](255, 255, 255, 255)
	pf.CopyBar(0, 0, width-1, height-1, white)

	src := color.NewRGBA8[color.Linear](0, 0, 255, 128)
	pf.BlendBar(2, 2, 5, 5, src, basics.CoverFull)

	for y := 2; y <= 5; y++ {
		for x := 2; x <= 5; x++ {
			p := pf.GetPixel(x, y)
			if p.B <= 128 {
				t.Errorf("BlendBar at (%d,%d): blue=%d should be > 128 after blending", x, y, p.B)
			}
		}
	}
}

func TestPixFmtRGBA32BlendSolidVspan(t *testing.T) {
	width, height := 1, 6
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](0, 200, 100, 255)

	// nil covers → uniform full coverage
	pf.BlendSolidVspan(0, 0, height, c, nil)
	for y := 0; y < height; y++ {
		if got := pf.GetPixel(0, y); got != c {
			t.Errorf("BlendSolidVspan(nil) at y=%d: got %+v want %+v", y, got, c)
		}
	}

	// varying covers
	buf2 := make([]basics.Int8u, 1*4*4)
	rbuf2 := buffer.NewRenderingBufferU8WithData(buf2, 1, 4, 1*4)
	pf2 := NewPixFmtRGBA32[color.Linear](rbuf2)
	covers := []basics.Int8u{255, 128, 64, 0}
	pf2.BlendSolidVspan(0, 0, 4, c, covers)

	// full cover → full color
	if got := pf2.GetPixel(0, 0); got != c {
		t.Errorf("BlendSolidVspan cover=255 should produce full color")
	}
	// zero cover → unchanged (black)
	if got := pf2.GetPixel(0, 3); got.G != 0 {
		t.Errorf("BlendSolidVspan cover=0 should leave pixel unchanged")
	}
}

func TestPixFmtRGBA32CopyColorHspan(t *testing.T) {
	width, height := 4, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	colors := []color.RGBA8[color.Linear]{
		color.NewRGBA8[color.Linear](10, 0, 0, 255),
		color.NewRGBA8[color.Linear](20, 0, 0, 255),
		color.NewRGBA8[color.Linear](30, 0, 0, 255),
		color.NewRGBA8[color.Linear](40, 0, 0, 255),
	}
	pf.CopyColorHspan(0, 0, 4, colors)

	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("CopyColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}
}

func TestPixFmtRGBA32CopyColorVspan(t *testing.T) {
	width, height := 1, 4
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	colors := []color.RGBA8[color.Linear]{
		color.NewRGBA8[color.Linear](10, 0, 0, 255),
		color.NewRGBA8[color.Linear](20, 0, 0, 255),
		color.NewRGBA8[color.Linear](30, 0, 0, 255),
		color.NewRGBA8[color.Linear](40, 0, 0, 255),
	}
	pf.CopyColorVspan(0, 0, 4, colors)

	for y, want := range colors {
		if got := pf.GetPixel(0, y); got != want {
			t.Errorf("CopyColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtRGBA32BlendColorHspan(t *testing.T) {
	width, height := 4, 1
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	colors := []color.RGBA8[color.Linear]{
		color.NewRGBA8[color.Linear](255, 0, 0, 255),
		color.NewRGBA8[color.Linear](0, 255, 0, 255),
		color.NewRGBA8[color.Linear](0, 0, 255, 255),
		color.NewRGBA8[color.Linear](100, 100, 100, 255),
	}
	// blend with full cover, no covers slice
	pf.BlendColorHspan(0, 0, 4, colors, nil, basics.CoverFull)

	for x, want := range colors {
		if got := pf.GetPixel(x, 0); got != want {
			t.Errorf("BlendColorHspan x=%d: got %+v want %+v", x, got, want)
		}
	}

	// blend with varying covers
	buf2 := make([]basics.Int8u, 2*4)
	rbuf2 := buffer.NewRenderingBufferU8WithData(buf2, 2, 1, 2*4)
	pf2 := NewPixFmtRGBA32[color.Linear](rbuf2)
	covers2 := []basics.Int8u{255, 0}
	pf2.BlendColorHspan(0, 0, 2, colors[:2], covers2, basics.CoverFull)
	if got := pf2.GetPixel(0, 0); got != colors[0] {
		t.Errorf("BlendColorHspan cover=255 x=0: got %+v want %+v", got, colors[0])
	}
	if got := pf2.GetPixel(1, 0); got.R != 0 || got.G != 0 {
		t.Errorf("BlendColorHspan cover=0 x=1: should be black/unchanged, got %+v", got)
	}
}

func TestPixFmtRGBA32BlendColorVspan(t *testing.T) {
	width, height := 1, 3
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	colors := []color.RGBA8[color.Linear]{
		color.NewRGBA8[color.Linear](255, 0, 0, 255),
		color.NewRGBA8[color.Linear](0, 255, 0, 255),
		color.NewRGBA8[color.Linear](0, 0, 255, 255),
	}
	pf.BlendColorVspan(0, 0, 3, colors, nil, basics.CoverFull)

	for y, want := range colors {
		if got := pf.GetPixel(0, y); got != want {
			t.Errorf("BlendColorVspan y=%d: got %+v want %+v", y, got, want)
		}
	}
}

func TestPixFmtRGBA32Fill(t *testing.T) {
	width, height := 4, 4
	buf := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](10, 20, 30, 255)
	pf.Fill(c)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if got := pf.GetPixel(x, y); got != c {
				t.Errorf("Fill at (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}
}

func TestPixFmtRGBA32Premultiply(t *testing.T) {
	buf := make([]basics.Int8u, 1*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// pixel with R=200, A=128 → premultiplied R ≈ 100
	pf.CopyPixel(0, 0, color.NewRGBA8[color.Linear](200, 0, 0, 128))
	pf.Premultiply()

	got := pf.GetPixel(0, 0)
	// R should be ~100 (200 * 128/255 ≈ 100)
	if got.R > 105 || got.R < 95 {
		t.Errorf("Premultiply: expected R≈100, got %d", got.R)
	}
}

func TestPixFmtRGBA32Demultiply(t *testing.T) {
	buf := make([]basics.Int8u, 1*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	// Start with premultiplied: R=100, A=128 → demultiplied R≈199
	pf.CopyPixel(0, 0, color.NewRGBA8[color.Linear](100, 0, 0, 128))
	pf.Demultiply()

	got := pf.GetPixel(0, 0)
	if got.R < 195 || got.R > 204 {
		t.Errorf("Demultiply: expected R≈199, got %d", got.R)
	}
}

func TestPixFmtRGBA32DemultiplyZeroAlpha(t *testing.T) {
	buf := make([]basics.Int8u, 1*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	pf.CopyPixel(0, 0, color.NewRGBA8[color.Linear](100, 50, 25, 0))
	pf.Demultiply() // zero alpha → R,G,B should be zeroed

	got := pf.GetPixel(0, 0)
	if got.R != 0 || got.G != 0 || got.B != 0 {
		t.Errorf("Demultiply zero alpha: expected RGB=0, got %+v", got)
	}
}

func TestPixFmtRGBA32ApplyGammaDir(t *testing.T) {
	buf := make([]basics.Int8u, 1*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	pf.CopyPixel(0, 0, color.NewRGBA8[color.Linear](100, 150, 200, 255))

	// identity gamma → no change
	identity := func(v basics.Int8u) basics.Int8u { return v }
	pf.ApplyGammaDir(identity)
	if got := pf.GetPixel(0, 0); got.R != 100 || got.G != 150 || got.B != 200 || got.A != 255 {
		t.Errorf("identity gamma changed pixel: %+v", got)
	}

	// double gamma → clamp at 255
	double := func(v basics.Int8u) basics.Int8u {
		if v > 127 {
			return 255
		}
		return v * 2
	}
	pf.ApplyGammaDir(double)
	got := pf.GetPixel(0, 0)
	if got.R != 200 || got.G != 255 || got.B != 255 {
		t.Errorf("double gamma failed: got R=%d G=%d B=%d", got.R, got.G, got.B)
	}
	// alpha should be unchanged
	if got.A != 255 {
		t.Errorf("gamma modified alpha: got %d", got.A)
	}
}

func TestPixFmtRGBA32ApplyGammaInv(t *testing.T) {
	buf := make([]basics.Int8u, 1*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)
	pf.CopyPixel(0, 0, color.NewRGBA8[color.Linear](50, 100, 150, 128))

	// ApplyGammaInv calls ApplyGammaDir, so just verify it runs without panic
	pf.ApplyGammaInv(func(v basics.Int8u) basics.Int8u { return v })
	if got := pf.GetPixel(0, 0); got.A != 128 {
		t.Errorf("ApplyGammaInv modified alpha: got %d", got.A)
	}
}

func TestPixFmtRGBA32PreConstructors(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)

	_ = NewPixFmtRGBA32Pre[color.Linear](rbuf)
	_ = NewPixFmtBGRA32Pre[color.Linear](rbuf)
	_ = NewPixFmtARGB32Pre[color.Linear](rbuf)
	_ = NewPixFmtABGR32Pre[color.Linear](rbuf)
}

func TestPixFmtRGBA32PlainConstructors(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)

	_ = NewPixFmtRGBA32Plain[color.Linear](rbuf)
	_ = NewPixFmtBGRA32Plain[color.Linear](rbuf)
	_ = NewPixFmtARGB32Plain[color.Linear](rbuf)
	_ = NewPixFmtABGR32Plain[color.Linear](rbuf)
}

func TestPixFmtRGBA32LinearConstructors(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)

	_ = NewPixFmtRGBA32Linear(rbuf)
	_ = NewPixFmtBGRA32Linear(rbuf)
	_ = NewPixFmtARGB32Linear(rbuf)
	_ = NewPixFmtABGR32Linear(rbuf)
	_ = NewPixFmtRGBA32PreLinear(rbuf)
	_ = NewPixFmtBGRA32PreLinear(rbuf)
	_ = NewPixFmtARGB32PreLinear(rbuf)
	_ = NewPixFmtABGR32PreLinear(rbuf)
	_ = NewPixFmtRGBA32PlainLinear(rbuf)
	_ = NewPixFmtBGRA32PlainLinear(rbuf)
	_ = NewPixFmtARGB32PlainLinear(rbuf)
	_ = NewPixFmtABGR32PlainLinear(rbuf)
}

func TestPixFmtRGBA32BlendHlinePre(t *testing.T) {
	// Test BlendHline with a premultiplied blender
	buf := make([]basics.Int8u, 8*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 8, 1, 8*4)
	pf := NewPixFmtRGBA32Pre[color.Linear](rbuf)

	c := color.NewRGBA8[color.Linear](128, 64, 32, 128)
	pf.BlendHline(0, 0, 4, c, basics.CoverFull)

	// After blending onto black background, pixels should have non-zero values
	for x := 0; x < 4; x++ {
		p := pf.GetPixel(x, 0)
		if p.R == 0 && p.G == 0 && p.B == 0 && p.A == 0 {
			t.Errorf("BlendHline Pre at x=%d: pixel is still zero after blend", x)
		}
	}
}

func TestPixFmtRGBA32RowData(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)

	row := pf.RowData(0)
	if len(row) == 0 {
		t.Error("RowData(0) should return non-empty slice")
	}
	if pf.RowData(-1) != nil {
		t.Error("RowData(-1) should return nil")
	}
	if pf.RowData(1) != nil {
		t.Error("RowData(out of bounds) should return nil")
	}
}

func TestPixFmtRGBA32BlendFromLUT(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtRGBA32[color.Linear](rbuf)
	src := &grayRowSource{row: []basics.Int8u{0, 2, 1}}
	lut := make([]color.RGBA8[color.Linear], 3)
	lut[0] = color.NewRGBA8[color.Linear](10, 20, 30, 255)
	lut[1] = color.NewRGBA8[color.Linear](40, 50, 60, 255)
	lut[2] = color.NewRGBA8[color.Linear](70, 80, 90, 255)

	pf.BlendFromLUT(src, lut, 0, 0, 0, 0, 3, basics.CoverFull)

	if got := pf.GetPixel(0, 0); got != lut[0] {
		t.Fatalf("lut mismatch at x=0: got %+v want %+v", got, lut[0])
	}
	if got := pf.GetPixel(1, 0); got != lut[2] {
		t.Fatalf("lut mismatch at x=1: got %+v want %+v", got, lut[2])
	}
	if got := pf.GetPixel(2, 0); got != lut[1] {
		t.Fatalf("lut mismatch at x=2: got %+v want %+v", got, lut[1])
	}
}
