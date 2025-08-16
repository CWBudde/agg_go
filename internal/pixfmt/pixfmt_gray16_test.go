package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

func TestGray16PixelType(t *testing.T) {
	p := &Gray16PixelType{}
	p.Set(0x8000)

	if p.V != 0x8000 {
		t.Errorf("Expected V=0x8000, got V=0x%X", p.V)
	}
}

func TestPixFmtGray16Basic(t *testing.T) {
	// Create a test buffer
	width, height := 100, 50
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2) // 2 bytes per pixel

	// Create pixel format
	pf := NewPixFmtGray16(rbuf)

	// Test basic properties
	if pf.Width() != width {
		t.Errorf("Width() expected %d, got %d", width, pf.Width())
	}
	if pf.Height() != height {
		t.Errorf("Height() expected %d, got %d", height, pf.Height())
	}
	if pf.PixWidth() != 2 {
		t.Errorf("PixWidth() expected 2, got %d", pf.PixWidth())
	}
}

func TestPixFmtGray16CopyPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	// Test copy pixel
	gray := color.NewGray16WithAlpha[color.Linear](0x8000, 0xFFFF)
	pf.CopyPixel(5, 5, gray)

	// Check that pixel was set
	retrievedGray := pf.GetPixel(5, 5)
	if retrievedGray.V != 0x8000 {
		t.Errorf("CopyPixel failed: expected V=0x8000, got V=0x%X", retrievedGray.V)
	}
}

func TestPixFmtGray16BlendPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	// Set background
	bgGray := color.NewGray16WithAlpha[color.Linear](0x4000, 0xFFFF)
	pf.CopyPixel(5, 5, bgGray)

	// Blend with another color
	blendGray := color.NewGray16WithAlpha[color.Linear](0xC000, 0x8000) // 50% alpha
	pf.BlendPixel(5, 5, blendGray, 0xFFFF)                              // Full coverage

	// Result should be between 0x4000 and 0xC000
	result := pf.GetPixel(5, 5)
	if result.V <= 0x4000 || result.V >= 0xC000 {
		t.Errorf("BlendPixel failed: expected blended value between 0x4000-0xC000, got 0x%X", result.V)
	}
}

func TestPixFmtGray16Lines(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	gray := color.NewGray16WithAlpha[color.Linear](0x8000, 0xFFFF)

	// Test horizontal line
	pf.CopyHline(5, 10, 15, gray)
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if pixel.V != 0x8000 {
			t.Errorf("CopyHline failed at (%d, 10): expected V=0x8000, got V=0x%X", x, pixel.V)
		}
	}

	// Test vertical line
	pf.CopyVline(10, 5, 15, gray)
	for y := 5; y <= 15; y++ {
		pixel := pf.GetPixel(10, y)
		if pixel.V != 0x8000 {
			t.Errorf("CopyVline failed at (10, %d): expected V=0x8000, got V=0x%X", y, pixel.V)
		}
	}
}

func TestPixFmtGray16BlendHline(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	// Fill with background
	bg := color.NewGray16WithAlpha[color.Linear](0x2000, 0xFFFF)
	pf.Fill(bg)

	// Blend a line
	blendColor := color.NewGray16WithAlpha[color.Linear](0xE000, 0x8000)
	pf.BlendHline(5, 10, 15, blendColor, 0xFFFF)

	// Check that line was blended
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if pixel.V <= 0x2000 || pixel.V >= 0xE000 {
			t.Errorf("BlendHline failed at (%d, 10): expected blended value, got V=0x%X", x, pixel.V)
		}
	}
}

func TestPixFmtGray16Rectangle(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	gray := color.NewGray16WithAlpha[color.Linear](0x8000, 0xFFFF)

	// Test copy rectangle
	pf.CopyBar(5, 5, 15, 15, gray)

	// Check rectangle corners and center
	checkPoints := []struct{ x, y int }{
		{5, 5}, {15, 5}, {5, 15}, {15, 15}, {10, 10},
	}

	for _, pt := range checkPoints {
		pixel := pf.GetPixel(pt.x, pt.y)
		if pixel.V != 0x8000 {
			t.Errorf("CopyBar failed at (%d, %d): expected V=0x8000, got V=0x%X", pt.x, pt.y, pixel.V)
		}
	}

	// Check that outside rectangle is not affected
	outsidePixel := pf.GetPixel(3, 3)
	if outsidePixel.V != 0 {
		t.Errorf("CopyBar affected outside pixel at (3, 3): expected V=0, got V=0x%X", outsidePixel.V)
	}
}

func TestPixFmtGray16SolidSpan(t *testing.T) {
	width, height := 20, 20
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	// Fill with background
	bg := color.NewGray16WithAlpha[color.Linear](0x2000, 0xFFFF)
	pf.Fill(bg)

	// Create coverage array
	covers := make([]basics.Int16u, 10)
	for i := range covers {
		covers[i] = basics.Int16u((i + 1) * 0x1999) // Increasing coverage
	}

	// Blend solid span
	blendColor := color.NewGray16WithAlpha[color.Linear](0xE000, 0xFFFF)
	pf.BlendSolidHspan(5, 10, 10, blendColor, covers)

	// Check that pixels have different values based on coverage
	lastValue := basics.Int16u(0)
	for i := 0; i < 10; i++ {
		pixel := pf.GetPixel(5+i, 10)
		if i > 0 && pixel.V <= lastValue {
			t.Errorf("BlendSolidHspan failed: pixel %d (V=0x%X) should be brighter than previous (V=0x%X)",
				i, pixel.V, lastValue)
		}
		lastValue = pixel.V
	}
}

func TestPixFmtGray16Clear(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	// Set some pixels first
	gray := color.NewGray16WithAlpha[color.Linear](0x8000, 0xFFFF)
	pf.CopyPixel(5, 5, gray)

	// Clear with different color
	clearColor := color.NewGray16WithAlpha[color.Linear](0x4000, 0xFFFF)
	pf.Clear(clearColor)

	// Check that all pixels are set to clear color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pf.GetPixel(x, y)
			if pixel.V != 0x4000 {
				t.Errorf("Clear failed at (%d, %d): expected V=0x4000, got V=0x%X", x, y, pixel.V)
			}
		}
	}
}

func TestPixFmtGray16BoundsChecking(t *testing.T) {
	width, height := 10, 10
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pf := NewPixFmtGray16(rbuf)

	gray := color.NewGray16WithAlpha[color.Linear](0x8000, 0xFFFF)

	// Test out of bounds operations (should not crash)
	pf.CopyPixel(-1, -1, gray)
	pf.CopyPixel(width, height, gray)
	pf.BlendPixel(-5, -5, gray, 0xFFFF)
	pf.BlendPixel(width+5, height+5, gray, 0xFFFF)

	// Get out of bounds pixel should return zero value
	outPixel := pf.GetPixel(-1, -1)
	if outPixel.V != 0 || outPixel.A != 0 {
		t.Errorf("Out of bounds GetPixel should return zero value, got V=0x%X, A=0x%X", outPixel.V, outPixel.A)
	}
}
