package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

func TestGrayPixelType(t *testing.T) {
	p := &GrayPixelType{}
	p.Set(128)

	if p.V != 128 {
		t.Errorf("Expected V=128, got V=%d", p.V)
	}
}

func TestPixFmtGray8Basic(t *testing.T) {
	// Create a test buffer
	width, height := 100, 50
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)

	// Create pixel format
	pf := NewPixFmtGray8(rbuf)

	// Test basic properties
	if pf.Width() != width {
		t.Errorf("Width() expected %d, got %d", width, pf.Width())
	}
	if pf.Height() != height {
		t.Errorf("Height() expected %d, got %d", height, pf.Height())
	}
	if pf.PixWidth() != 1 {
		t.Errorf("PixWidth() expected 1, got %d", pf.PixWidth())
	}
}

func TestPixFmtGray8CopyPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	// Test copy pixel
	gray := color.NewGray8WithAlpha[color.Linear](128, 255)
	pf.CopyPixel(5, 5, gray)

	// Check that pixel was set
	retrievedGray := pf.GetPixel(5, 5)
	if retrievedGray.V != 128 {
		t.Errorf("CopyPixel failed: expected V=128, got V=%d", retrievedGray.V)
	}
}

func TestPixFmtGray8BlendPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	// Set background
	bgGray := color.NewGray8WithAlpha[color.Linear](100, 255)
	pf.CopyPixel(5, 5, bgGray)

	// Blend with another color
	blendGray := color.NewGray8WithAlpha[color.Linear](200, 128) // 50% alpha
	pf.BlendPixel(5, 5, blendGray, 255)                          // Full coverage

	// Result should be between 100 and 200
	result := pf.GetPixel(5, 5)
	if result.V <= 100 || result.V >= 200 {
		t.Errorf("BlendPixel failed: expected blended value between 100-200, got %d", result.V)
	}
}

func TestPixFmtGray8Lines(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	gray := color.NewGray8WithAlpha[color.Linear](128, 255)

	// Test horizontal line
	pf.CopyHline(5, 10, 15, gray)
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if pixel.V != 128 {
			t.Errorf("CopyHline failed at (%d, 10): expected V=128, got V=%d", x, pixel.V)
		}
	}

	// Test vertical line
	pf.CopyVline(10, 5, 15, gray)
	for y := 5; y <= 15; y++ {
		pixel := pf.GetPixel(10, y)
		if pixel.V != 128 {
			t.Errorf("CopyVline failed at (10, %d): expected V=128, got V=%d", y, pixel.V)
		}
	}
}

func TestPixFmtGray8Rectangle(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	gray := color.NewGray8WithAlpha[color.Linear](200, 255)

	// Test filled rectangle
	pf.CopyBar(5, 5, 10, 10, gray)

	// Check corners
	corners := [][2]int{{5, 5}, {10, 5}, {5, 10}, {10, 10}}
	for _, corner := range corners {
		pixel := pf.GetPixel(corner[0], corner[1])
		if pixel.V != 200 {
			t.Errorf("CopyBar failed at (%d, %d): expected V=200, got V=%d",
				corner[0], corner[1], pixel.V)
		}
	}

	// Check that pixels outside rectangle are unchanged (should be 0)
	pixel := pf.GetPixel(4, 4)
	if pixel.V != 0 {
		t.Errorf("CopyBar affected pixel outside rectangle at (4, 4): got V=%d", pixel.V)
	}
}

func TestPixFmtGray8Spans(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	gray := color.NewGray8WithAlpha[color.Linear](150, 255)

	// Test horizontal span with varying coverage
	covers := []basics.Int8u{255, 200, 150, 100, 50}
	pf.BlendSolidHspan(5, 10, len(covers), gray, covers)

	// Check that coverage affects the result
	for i, cover := range covers {
		pixel := pf.GetPixel(5+i, 10)
		// With higher coverage, we should get values closer to 150
		// With lower coverage, closer to 0 (background)
		if cover == 255 && pixel.V != 150 {
			t.Errorf("BlendSolidHspan with full coverage should give V=150, got V=%d", pixel.V)
		}
		if cover == 50 && pixel.V >= 150 {
			t.Errorf("BlendSolidHspan with low coverage should give V<150, got V=%d", pixel.V)
		}
	}
}

func TestPixFmtGray8Clear(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height)
	// Initialize buffer with non-zero values
	for i := range buf {
		buf[i] = 100
	}

	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	// Clear with a specific color
	clearGray := color.NewGray8WithAlpha[color.Linear](50, 255)
	pf.Clear(clearGray)

	// Check that all pixels are now the clear color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pf.GetPixel(x, y)
			if pixel.V != 50 {
				t.Errorf("Clear failed at (%d, %d): expected V=50, got V=%d", x, y, pixel.V)
			}
		}
	}
}

func TestPixFmtGray8Bounds(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	gray := color.NewGray8WithAlpha[color.Linear](128, 255)

	// Test out-of-bounds operations (should not crash)
	pf.CopyPixel(-1, -1, gray)
	pf.CopyPixel(width, height, gray)
	pf.BlendPixel(-1, -1, gray, 255)
	pf.BlendPixel(width, height, gray, 255)

	// These operations should be safe and not affect valid pixels
	pf.CopyPixel(0, 0, gray)
	pixel := pf.GetPixel(0, 0)
	if pixel.V != 128 {
		t.Errorf("Valid pixel operation failed after out-of-bounds tests")
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test ClampX
	if ClampX(-5, 10) != 0 {
		t.Error("ClampX should clamp negative values to 0")
	}
	if ClampX(15, 10) != 9 {
		t.Error("ClampX should clamp values >= width to width-1")
	}
	if ClampX(5, 10) != 5 {
		t.Error("ClampX should return valid values unchanged")
	}

	// Test ClampY
	if ClampY(-5, 10) != 0 {
		t.Error("ClampY should clamp negative values to 0")
	}
	if ClampY(15, 10) != 9 {
		t.Error("ClampY should clamp values >= height to height-1")
	}

	// Test InBounds
	if !InBounds(5, 5, 10, 10) {
		t.Error("InBounds should return true for valid coordinates")
	}
	if InBounds(-1, 5, 10, 10) {
		t.Error("InBounds should return false for negative x")
	}
	if InBounds(5, -1, 10, 10) {
		t.Error("InBounds should return false for negative y")
	}
	if InBounds(10, 5, 10, 10) {
		t.Error("InBounds should return false for x >= width")
	}
	if InBounds(5, 10, 10, 10) {
		t.Error("InBounds should return false for y >= height")
	}

	// Test Min/Max
	if Min(5, 10) != 5 {
		t.Error("Min should return smaller value")
	}
	if Max(5, 10) != 10 {
		t.Error("Max should return larger value")
	}
}

func TestPixFmtGray8CopyFromOverlapSameRow(t *testing.T) {
	width, height := 5, 1
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)
	pf := NewPixFmtGray8(rbuf)

	for x, v := range []basics.Int8u{1, 2, 3, 4, 5} {
		pf.CopyPixel(x, 0, color.Gray8[color.Linear]{V: v, A: 255})
	}

	pf.CopyFrom(pf, 1, 0, 0, 0, 4)

	for x, want := range []basics.Int8u{1, 1, 2, 3, 4} {
		if got := pf.GetPixel(x, 0).V; got != want {
			t.Fatalf("CopyFrom overlap mismatch at x=%d: got %d want %d", x, got, want)
		}
	}
}

func TestPixFmtGray8BlendFromColor(t *testing.T) {
	buf := make([]basics.Int8u, 3)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3)
	pf := NewPixFmtGray8(rbuf)
	src := &grayRowSource{row: []basics.Int8u{0, 128, 255}}
	c := color.NewGray8WithAlpha[color.Linear](200, 255)

	pf.BlendFromColor(src, c, 0, 0, 0, 0, 3, basics.CoverFull)

	if got := pf.GetPixel(0, 0).V; got != 0 {
		t.Fatalf("expected zero-coverage pixel to stay empty, got %d", got)
	}
	mid := pf.GetPixel(1, 0).V
	if mid == 0 || mid >= c.V {
		t.Fatalf("expected partial blend at x=1, got %d", mid)
	}
	if got := pf.GetPixel(2, 0).V; got != c.V {
		t.Fatalf("expected full-coverage pixel to match source gray, got %d want %d", got, c.V)
	}
}

func TestPixFmtGray8BlendFromLUT(t *testing.T) {
	buf := make([]basics.Int8u, 3)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3)
	pf := NewPixFmtGray8(rbuf)
	src := &grayRowSource{row: []basics.Int8u{0, 2, 1}}
	lut := []color.Gray8[color.Linear]{
		{V: 10, A: 255},
		{V: 40, A: 255},
		{V: 70, A: 255},
	}

	pf.BlendFromLUT(src, lut, 0, 0, 0, 0, 3, basics.CoverFull)

	if got := pf.GetPixel(0, 0).V; got != lut[0].V {
		t.Fatalf("lut mismatch at x=0: got %d want %d", got, lut[0].V)
	}
	if got := pf.GetPixel(1, 0).V; got != lut[2].V {
		t.Fatalf("lut mismatch at x=1: got %d want %d", got, lut[2].V)
	}
	if got := pf.GetPixel(2, 0).V; got != lut[1].V {
		t.Fatalf("lut mismatch at x=2: got %d want %d", got, lut[1].V)
	}
}
