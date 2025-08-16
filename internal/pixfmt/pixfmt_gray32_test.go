package pixfmt

import (
	"math"
	"testing"

	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

func TestGray32PixelType(t *testing.T) {
	p := &Gray32PixelType{}
	p.Set(0.5)

	if math.Abs(float64(p.V-0.5)) > 0.001 {
		t.Errorf("Expected V=0.5, got V=%f", p.V)
	}
}

func TestPixFmtGray32Basic(t *testing.T) {
	// Create a test buffer
	width, height := 100, 50
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4) // 4 bytes per pixel

	// Create pixel format
	pf := NewPixFmtGray32(rbuf)

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

func TestPixFmtGray32CopyPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Test copy pixel
	gray := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)
	pf.CopyPixel(5, 5, gray)

	// Check that pixel was set
	retrievedGray := pf.GetPixel(5, 5)
	if math.Abs(float64(retrievedGray.V-0.5)) > 0.001 {
		t.Errorf("CopyPixel failed: expected V=0.5, got V=%f", retrievedGray.V)
	}
}

func TestPixFmtGray32BlendPixel(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Set background
	bgGray := color.NewGray32WithAlpha[color.Linear](0.25, 1.0)
	pf.CopyPixel(5, 5, bgGray)

	// Blend with another color
	blendGray := color.NewGray32WithAlpha[color.Linear](0.75, 0.5) // 50% alpha
	pf.BlendPixel(5, 5, blendGray, 1.0)                            // Full coverage

	// Result should be between 0.25 and 0.75
	result := pf.GetPixel(5, 5)
	if result.V <= 0.25 || result.V >= 0.75 {
		t.Errorf("BlendPixel failed: expected blended value between 0.25-0.75, got %f", result.V)
	}
}

func TestPixFmtGray32Lines(t *testing.T) {
	width, height := 20, 20
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	gray := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)

	// Test horizontal line
	pf.CopyHline(5, 10, 15, gray)
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if math.Abs(float64(pixel.V-0.5)) > 0.001 {
			t.Errorf("CopyHline failed at (%d, 10): expected V=0.5, got V=%f", x, pixel.V)
		}
	}

	// Test vertical line
	pf.CopyVline(10, 5, 15, gray)
	for y := 5; y <= 15; y++ {
		pixel := pf.GetPixel(10, y)
		if math.Abs(float64(pixel.V-0.5)) > 0.001 {
			t.Errorf("CopyVline failed at (10, %d): expected V=0.5, got V=%f", y, pixel.V)
		}
	}
}

func TestPixFmtGray32BlendHline(t *testing.T) {
	width, height := 20, 20
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Fill with background
	bg := color.NewGray32WithAlpha[color.Linear](0.125, 1.0)
	pf.Fill(bg)

	// Blend a line
	blendColor := color.NewGray32WithAlpha[color.Linear](0.875, 0.5)
	pf.BlendHline(5, 10, 15, blendColor, 1.0)

	// Check that line was blended
	for x := 5; x <= 15; x++ {
		pixel := pf.GetPixel(x, 10)
		if pixel.V <= 0.125 || pixel.V >= 0.875 {
			t.Errorf("BlendHline failed at (%d, 10): expected blended value, got V=%f", x, pixel.V)
		}
	}
}

func TestPixFmtGray32Rectangle(t *testing.T) {
	width, height := 20, 20
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	gray := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)

	// Test copy rectangle
	pf.CopyBar(5, 5, 15, 15, gray)

	// Check rectangle corners and center
	checkPoints := []struct{ x, y int }{
		{5, 5}, {15, 5}, {5, 15}, {15, 15}, {10, 10},
	}

	for _, pt := range checkPoints {
		pixel := pf.GetPixel(pt.x, pt.y)
		if math.Abs(float64(pixel.V-0.5)) > 0.001 {
			t.Errorf("CopyBar failed at (%d, %d): expected V=0.5, got V=%f", pt.x, pt.y, pixel.V)
		}
	}

	// Check that outside rectangle is not affected
	outsidePixel := pf.GetPixel(3, 3)
	if math.Abs(float64(outsidePixel.V)) > 0.001 {
		t.Errorf("CopyBar affected outside pixel at (3, 3): expected V=0, got V=%f", outsidePixel.V)
	}
}

func TestPixFmtGray32SolidSpan(t *testing.T) {
	width, height := 20, 20
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Fill with background
	bg := color.NewGray32WithAlpha[color.Linear](0.125, 1.0)
	pf.Fill(bg)

	// Create coverage array
	covers := make([]float32, 10)
	for i := range covers {
		covers[i] = float32(i+1) / 10.0 // Increasing coverage from 0.1 to 1.0
	}

	// Blend solid span
	blendColor := color.NewGray32WithAlpha[color.Linear](0.875, 1.0)
	pf.BlendSolidHspan(5, 10, 10, blendColor, covers)

	// Check that pixels have different values based on coverage
	lastValue := float32(0)
	for i := 0; i < 10; i++ {
		pixel := pf.GetPixel(5+i, 10)
		if i > 0 && pixel.V <= lastValue {
			t.Errorf("BlendSolidHspan failed: pixel %d (V=%f) should be brighter than previous (V=%f)",
				i, pixel.V, lastValue)
		}
		lastValue = pixel.V
	}
}

func TestPixFmtGray32Clear(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Set some pixels first
	gray := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)
	pf.CopyPixel(5, 5, gray)

	// Clear with different color
	clearColor := color.NewGray32WithAlpha[color.Linear](0.25, 1.0)
	pf.Clear(clearColor)

	// Check that all pixels are set to clear color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pf.GetPixel(x, y)
			if math.Abs(float64(pixel.V-0.25)) > 0.001 {
				t.Errorf("Clear failed at (%d, %d): expected V=0.25, got V=%f", x, y, pixel.V)
			}
		}
	}
}

func TestPixFmtGray32BoundsChecking(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	gray := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)

	// Test out of bounds operations (should not crash)
	pf.CopyPixel(-1, -1, gray)
	pf.CopyPixel(width, height, gray)
	pf.BlendPixel(-5, -5, gray, 1.0)
	pf.BlendPixel(width+5, height+5, gray, 1.0)

	// Get out of bounds pixel should return zero value
	outPixel := pf.GetPixel(-1, -1)
	if math.Abs(float64(outPixel.V)) > 0.001 || math.Abs(float64(outPixel.A)) > 0.001 {
		t.Errorf("Out of bounds GetPixel should return zero value, got V=%f, A=%f", outPixel.V, outPixel.A)
	}
}

func TestPixFmtGray32ZeroAlpha(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Set background
	bg := color.NewGray32WithAlpha[color.Linear](0.5, 1.0)
	pf.CopyPixel(5, 5, bg)

	// Try to blend with zero alpha color
	zeroAlpha := color.NewGray32WithAlpha[color.Linear](0.9, 0.0)
	pf.BlendPixel(5, 5, zeroAlpha, 1.0)

	// Pixel should remain unchanged
	result := pf.GetPixel(5, 5)
	if math.Abs(float64(result.V-0.5)) > 0.001 {
		t.Errorf("BlendPixel with zero alpha should not change pixel: expected V=0.5, got V=%f", result.V)
	}
}

func TestPixFmtGray32PartialCoverage(t *testing.T) {
	width, height := 10, 10
	buf := make([]float32, width*height)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, width, height, width*4)
	pf := NewPixFmtGray32(rbuf)

	// Set background
	bg := color.NewGray32WithAlpha[color.Linear](0.2, 1.0)
	pf.CopyPixel(5, 5, bg)

	// Blend with partial coverage
	blendColor := color.NewGray32WithAlpha[color.Linear](0.8, 1.0)
	pf.BlendPixel(5, 5, blendColor, 0.5) // 50% coverage

	// Result should be closer to background than to blend color
	result := pf.GetPixel(5, 5)
	expectedApprox := float32(0.2 + (0.8-0.2)*0.5)        // Linear interpolation with 50% coverage
	if math.Abs(float64(result.V-expectedApprox)) > 0.1 { // Allow some tolerance
		t.Errorf("BlendPixel with partial coverage failed: expected ~%f, got %f", expectedApprox, result.V)
	}
}
