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
