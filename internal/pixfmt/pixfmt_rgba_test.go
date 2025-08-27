package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

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
