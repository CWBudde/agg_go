package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

func TestPixFmtRGB24(t *testing.T) {
	// Create a 4x4 RGB24 buffer (48 bytes total)
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)

	// Create RGB24 pixel format
	pixfmt := NewPixFmtRGB24(rbuf)

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width() failed: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.Height() != height {
		t.Errorf("Height() failed: got %d, want %d", pixfmt.Height(), height)
	}
	if pixfmt.PixWidth() != 3 {
		t.Errorf("PixWidth() failed: got %d, want 3", pixfmt.PixWidth())
	}

	// Test pixel operations
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	green := color.RGB8Linear{R: 0, G: 255, B: 0}
	blue := color.RGB8Linear{R: 0, G: 0, B: 255}

	// Test CopyPixel
	pixfmt.CopyPixel(0, 0, red)
	pixfmt.CopyPixel(1, 0, green)
	pixfmt.CopyPixel(2, 0, blue)

	// Test GetPixel
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("GetPixel(0,0) failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = pixfmt.GetPixel(1, 0)
	if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
		t.Errorf("GetPixel(1,0) failed: got {%d, %d, %d}, want {0, 255, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = pixfmt.GetPixel(2, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 255 {
		t.Errorf("GetPixel(2,0) failed: got {%d, %d, %d}, want {0, 0, 255}", pixel.R, pixel.G, pixel.B)
	}

	// Test bounds checking
	pixel = pixfmt.GetPixel(-1, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("GetPixel(-1,0) bounds check failed: got {%d, %d, %d}, want {0, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = pixfmt.GetPixel(width, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("GetPixel(width,0) bounds check failed: got {%d, %d, %d}, want {0, 0, 0}", pixel.R, pixel.G, pixel.B)
	}
}

func TestPixFmtRGB24BlendPixel(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	// Set initial pixel to gray
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.CopyPixel(0, 0, gray)

	// Blend red with full alpha and coverage
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	pixfmt.BlendPixel(0, 0, red, 255, 255)

	// Should be red now
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("BlendPixel full alpha failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	// Reset and blend with half alpha
	pixfmt.CopyPixel(1, 0, gray)
	pixfmt.BlendPixel(1, 0, red, 128, 255)

	pixel = pixfmt.GetPixel(1, 0)
	// Should be somewhere between gray and red
	if pixel.R <= 128 || pixel.R >= 255 {
		t.Errorf("BlendPixel half alpha failed: red component %d should be between 128 and 255", pixel.R)
	}
	if pixel.G >= 128 {
		t.Errorf("BlendPixel half alpha failed: green component %d should be less than 128", pixel.G)
	}
}

func TestPixFmtRGB24Lines(t *testing.T) {
	width, height := 8, 8
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	red := color.RGB8Linear{R: 255, G: 0, B: 0}

	// Test CopyHline
	pixfmt.CopyHline(2, 1, 5, red)
	for x := 2; x <= 5; x++ {
		pixel := pixfmt.GetPixel(x, 1)
		if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
			t.Errorf("CopyHline failed at (%d,1): got {%d, %d, %d}, want {255, 0, 0}", x, pixel.R, pixel.G, pixel.B)
		}
	}

	// Test CopyVline
	pixfmt.CopyVline(3, 2, 5, red)
	for y := 2; y <= 5; y++ {
		pixel := pixfmt.GetPixel(3, y)
		if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
			t.Errorf("CopyVline failed at (3,%d): got {%d, %d, %d}, want {255, 0, 0}", y, pixel.R, pixel.G, pixel.B)
		}
	}

	// Test BlendHline with gray background
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.CopyHline(0, 7, 7, gray)           // Fill row with gray
	pixfmt.BlendHline(2, 7, 5, red, 128, 255) // Blend red with half alpha

	for x := 2; x <= 5; x++ {
		pixel := pixfmt.GetPixel(x, 7)
		if pixel.R <= 128 || pixel.R >= 255 {
			t.Errorf("BlendHline failed at (%d,7): red component %d should be between 128 and 255", x, pixel.R)
		}
	}
}

func TestPixFmtRGB24Bar(t *testing.T) {
	width, height := 8, 8
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	blue := color.RGB8Linear{R: 0, G: 0, B: 255}

	// Test CopyBar
	pixfmt.CopyBar(2, 2, 5, 5, blue)
	for y := 2; y <= 5; y++ {
		for x := 2; x <= 5; x++ {
			pixel := pixfmt.GetPixel(x, y)
			if pixel.R != 0 || pixel.G != 0 || pixel.B != 255 {
				t.Errorf("CopyBar failed at (%d,%d): got {%d, %d, %d}, want {0, 0, 255}", x, y, pixel.R, pixel.G, pixel.B)
			}
		}
	}

	// Test BlendBar
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.Clear(gray)                          // Fill entire buffer with gray
	pixfmt.BlendBar(1, 1, 3, 3, blue, 128, 255) // Blend blue with half alpha

	for y := 1; y <= 3; y++ {
		for x := 1; x <= 3; x++ {
			pixel := pixfmt.GetPixel(x, y)
			if pixel.B <= 128 || pixel.B >= 255 {
				t.Errorf("BlendBar failed at (%d,%d): blue component %d should be between 128 and 255", x, y, pixel.B)
			}
		}
	}
}

func TestPixFmtRGB24Spans(t *testing.T) {
	width, height := 8, 8
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	green := color.RGB8Linear{R: 0, G: 255, B: 0}

	// Test BlendSolidHspan with uniform coverage
	pixfmt.BlendSolidHspan(1, 1, 4, green, 255, nil)
	for x := 1; x < 5; x++ {
		pixel := pixfmt.GetPixel(x, 1)
		if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
			t.Errorf("BlendSolidHspan uniform failed at (%d,1): got {%d, %d, %d}, want {0, 255, 0}", x, pixel.R, pixel.G, pixel.B)
		}
	}

	// Test BlendSolidHspan with varying coverage
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.CopyHline(0, 2, 7, gray) // Fill row with gray
	covers := []basics.Int8u{255, 128, 64, 32}
	pixfmt.BlendSolidHspan(2, 2, 4, green, 255, covers)

	// First pixel should be fully green
	pixel := pixfmt.GetPixel(2, 2)
	if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
		t.Errorf("BlendSolidHspan varying coverage pixel 0 failed: got {%d, %d, %d}, want {0, 255, 0}", pixel.R, pixel.G, pixel.B)
	}

	// Other pixels should be partially blended
	for i := 1; i < 4; i++ {
		pixel = pixfmt.GetPixel(2+i, 2)
		if pixel.G <= 128 || pixel.G >= 255 {
			t.Errorf("BlendSolidHspan varying coverage pixel %d failed: green component %d should be between 128 and 255", i, pixel.G)
		}
	}

	// Test BlendSolidVspan
	pixfmt.Clear(gray)
	pixfmt.BlendSolidVspan(3, 1, 4, green, 255, nil)
	for y := 1; y < 5; y++ {
		pixel := pixfmt.GetPixel(3, y)
		if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
			t.Errorf("BlendSolidVspan failed at (3,%d): got {%d, %d, %d}, want {0, 255, 0}", y, pixel.R, pixel.G, pixel.B)
		}
	}
}

func TestPixFmtRGB24Clear(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	// Fill with some random data first
	for i := range bufData {
		bufData[i] = basics.Int8u(i % 256)
	}

	// Clear with red
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	pixfmt.Clear(red)

	// Check all pixels are red
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pixfmt.GetPixel(x, y)
			if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
				t.Errorf("Clear failed at (%d,%d): got {%d, %d, %d}, want {255, 0, 0}", x, y, pixel.R, pixel.G, pixel.B)
			}
		}
	}

	// Test Fill (should be same as Clear)
	blue := color.RGB8Linear{R: 0, G: 0, B: 255}
	pixfmt.Fill(blue)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := pixfmt.GetPixel(x, y)
			if pixel.R != 0 || pixel.G != 0 || pixel.B != 255 {
				t.Errorf("Fill failed at (%d,%d): got {%d, %d, %d}, want {0, 0, 255}", x, y, pixel.R, pixel.G, pixel.B)
			}
		}
	}
}

func TestPixFmtBGR24(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)

	// Create BGR24 pixel format
	pixfmt := NewPixFmtBGR24(rbuf)

	// Test that RGB values are stored in BGR order
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	pixfmt.CopyPixel(0, 0, red)

	// In BGR format, red should be at index 2
	if bufData[0] != 0 || bufData[1] != 0 || bufData[2] != 255 {
		t.Errorf("BGR24 order failed: buffer [%d, %d, %d] should be [0, 0, 255] for red color",
			bufData[0], bufData[1], bufData[2])
	}

	// Verify GetPixel still returns the correct RGB values
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("BGR24 GetPixel failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}
}

func TestPixFmtRGB24CopyFrom(t *testing.T) {
	width, height := 4, 4

	// Create source format
	srcData := make([]basics.Int8u, width*height*3)
	srcBuf := buffer.NewRenderingBufferU8WithData(srcData, width, height, width*3)
	src := NewPixFmtRGB24(srcBuf)

	// Fill source with a pattern
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	green := color.RGB8Linear{R: 0, G: 255, B: 0}
	src.CopyPixel(0, 0, red)
	src.CopyPixel(1, 0, green)
	src.CopyPixel(0, 1, green)
	src.CopyPixel(1, 1, red)

	// Create destination format
	dstData := make([]basics.Int8u, width*height*3)
	dstBuf := buffer.NewRenderingBufferU8WithData(dstData, width, height, width*3)
	dst := NewPixFmtRGB24(dstBuf)

	// Copy 2x2 region
	dst.CopyFrom(src, 0, 0, 0, 0, 2, 2)

	// Verify copy
	pixel := dst.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("CopyFrom (0,0) failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = dst.GetPixel(1, 0)
	if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
		t.Errorf("CopyFrom (1,0) failed: got {%d, %d, %d}, want {0, 255, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = dst.GetPixel(0, 1)
	if pixel.R != 0 || pixel.G != 255 || pixel.B != 0 {
		t.Errorf("CopyFrom (0,1) failed: got {%d, %d, %d}, want {0, 255, 0}", pixel.R, pixel.G, pixel.B)
	}

	pixel = dst.GetPixel(1, 1)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("CopyFrom (1,1) failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}
}

func TestPixFmtRGB24BlendPixelRGBA(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)

	// Set initial pixel to gray
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.CopyPixel(0, 0, gray)

	// Blend RGBA pixel (alpha becomes the blending alpha)
	rgba := color.RGBA8Linear{R: 255, G: 0, B: 0, A: 128}
	pixfmt.BlendPixelRGBA(0, 0, rgba, 255)

	// Should be blended based on RGBA's alpha
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R <= 128 || pixel.R >= 255 {
		t.Errorf("BlendPixelRGBA failed: red component %d should be between 128 and 255", pixel.R)
	}
	if pixel.G >= 128 {
		t.Errorf("BlendPixelRGBA failed: green component %d should be less than 128", pixel.G)
	}
}

// Benchmark tests
func BenchmarkPixFmtRGB24CopyPixel(b *testing.B) {
	width, height := 100, 100
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)
	color := color.RGB8Linear{R: 255, G: 128, B: 64}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % width
		y := (i / width) % height
		pixfmt.CopyPixel(x, y, color)
	}
}

func BenchmarkPixFmtRGB24BlendPixel(b *testing.B) {
	width, height := 100, 100
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)
	color := color.RGB8Linear{R: 255, G: 128, B: 64}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % width
		y := (i / width) % height
		pixfmt.BlendPixel(x, y, color, 255, 255)
	}
}

func BenchmarkPixFmtRGB24Clear(b *testing.B) {
	width, height := 100, 100
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24(rbuf)
	color := color.RGB8Linear{R: 255, G: 128, B: 64}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pixfmt.Clear(color)
	}
}

// ==============================================================================
// RGBX32 (RGB with padding byte) Tests
// ==============================================================================

func TestPixFmtRGBX32(t *testing.T) {
	// Create a 4x4 RGBX32 buffer (64 bytes total: 4*4*4)
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*4)

	// Create RGBX32 pixel format
	pixfmt := NewPixFmtRGBX32Linear(rbuf)

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width() failed: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.Height() != height {
		t.Errorf("Height() failed: got %d, want %d", pixfmt.Height(), height)
	}
	if pixfmt.PixWidth() != 4 {
		t.Errorf("PixWidth() failed: got %d, want 4", pixfmt.PixWidth())
	}

	// Test pixel operations
	red := color.RGB8Linear{R: 255, G: 0, B: 0}

	// Set padding byte to a known value
	bufData[3] = 99 // Padding byte for first pixel

	// Test CopyPixel
	pixfmt.CopyPixel(0, 0, red)

	// Verify RGB values are set correctly
	if bufData[0] != 255 || bufData[1] != 0 || bufData[2] != 0 {
		t.Errorf("RGBX32 CopyPixel failed: buffer [%d, %d, %d, %d] should have RGB [255, 0, 0]",
			bufData[0], bufData[1], bufData[2], bufData[3])
	}

	// Verify padding byte is unchanged
	if bufData[3] != 99 {
		t.Errorf("RGBX32 CopyPixel modified padding byte: got %d, want 99", bufData[3])
	}

	// Test GetPixel
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("RGBX32 GetPixel failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}
}

func TestPixFmtXRGB32(t *testing.T) {
	// Create a 4x4 XRGB32 buffer (64 bytes total: 4*4*4)
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*4)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*4)

	// Create XRGB32 pixel format
	pixfmt := NewPixFmtXRGB32Pre[color.Linear](rbuf)

	// Test pixel operations
	red := color.RGB8Linear{R: 255, G: 0, B: 0}

	// Set padding byte to a known value
	bufData[0] = 99 // Padding byte for first pixel (at beginning)

	// Test CopyPixel
	pixfmt.CopyPixel(0, 0, red)

	// Verify RGB values are set correctly (at offsets 1, 2, 3)
	if bufData[1] != 255 || bufData[2] != 0 || bufData[3] != 0 {
		t.Errorf("XRGB32 CopyPixel failed: buffer [%d, %d, %d, %d] should have RGB at [1,2,3] = [255, 0, 0]",
			bufData[0], bufData[1], bufData[2], bufData[3])
	}

	// Verify padding byte is unchanged
	if bufData[0] != 99 {
		t.Errorf("XRGB32 CopyPixel modified padding byte: got %d, want 99", bufData[0])
	}

	// Test GetPixel
	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("XRGB32 GetPixel failed: got {%d, %d, %d}, want {255, 0, 0}", pixel.R, pixel.G, pixel.B)
	}
}

// ==============================================================================
// Premultiplied RGB Tests
// ==============================================================================

func TestPixFmtRGB24Pre(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24Pre(rbuf)

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width() failed: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.PixWidth() != 3 {
		t.Errorf("PixWidth() failed: got %d, want 3", pixfmt.PixWidth())
	}

	// Set initial pixel to gray
	gray := color.RGB8Linear{R: 128, G: 128, B: 128}
	pixfmt.CopyPixel(0, 0, gray)

	// Blend with premultiplied semantics
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	pixfmt.BlendPixel(0, 0, red, 128, 255)

	// Should use premultiplied blending
	pixel := pixfmt.GetPixel(0, 0)
	// With C++ implementation: prelerp(128, 255, 128) = 128 + 255 - multiply(128, 128) = 128 + 255 - 64 = 319 (wraps to 63)
	// This matches the original AGG behavior but is counter-intuitive for RGB blending
	// Expected result is 63 (due to arithmetic overflow wrapping)
	if pixel.R != 63 {
		t.Errorf("Premultiplied blending failed: red component %d should be 63 (matching C++ AGG behavior)", pixel.R)
	}
	// Test that green and blue are also affected correctly
	// When blending red (255,0,0) with alpha 128 onto gray (128,128,128):
	// prelerp(128, 0, 128) = 128 + 0 - multiply(128, 128) = 128 + 0 - 64 = 64
	if pixel.G != 64 || pixel.B != 64 {
		t.Errorf("Premultiplied blending failed: G and B components should be 64, got G=%d, B=%d", pixel.G, pixel.B)
	}
}

// Additional comprehensive tests for premultiplied blending
func TestPixFmtRGB24Pre_ComprehensiveBlending(t *testing.T) {
	width, height := 4, 4
	bufData := make([]basics.Int8u, width*height*3)
	rbuf := buffer.NewRenderingBufferU8WithData(bufData, width, height, width*3)
	pixfmt := NewPixFmtRGB24Pre(rbuf)

	// Test case 1: Blend with zero alpha (should have no effect)
	white := color.RGB8Linear{R: 255, G: 255, B: 255}
	pixfmt.CopyPixel(0, 0, white)
	red := color.RGB8Linear{R: 255, G: 0, B: 0}
	pixfmt.BlendPixel(0, 0, red, 0, 255) // Zero alpha

	pixel := pixfmt.GetPixel(0, 0)
	if pixel.R != 255 || pixel.G != 255 || pixel.B != 255 {
		t.Errorf("Zero alpha blend failed: got {%d, %d, %d}, want {255, 255, 255}", pixel.R, pixel.G, pixel.B)
	}

	// Test case 2: Blend with full alpha (should completely replace)
	black := color.RGB8Linear{R: 0, G: 0, B: 0}
	pixfmt.CopyPixel(1, 0, white)
	pixfmt.BlendPixel(1, 0, black, 255, 255) // Full alpha

	pixel = pixfmt.GetPixel(1, 0)
	if pixel.R != 0 || pixel.G != 0 || pixel.B != 0 {
		t.Errorf("Full alpha blend failed: got {%d, %d, %d}, want {0, 0, 0}", pixel.R, pixel.G, pixel.B)
	}

	// Test case 3: Partial coverage
	pixfmt.CopyPixel(2, 0, white)
	pixfmt.BlendPixel(2, 0, black, 255, 128) // Full alpha, half coverage

	pixel = pixfmt.GetPixel(2, 0)
	// prelerp(255, 0, 128) = 255 + 0 - multiply(255, 128) = 255 + 0 - 128 = 127
	if pixel.R != 127 || pixel.G != 127 || pixel.B != 127 {
		t.Errorf("Half coverage blend failed: got {%d, %d, %d}, want {127, 127, 127}", pixel.R, pixel.G, pixel.B)
	}

	// Test case 4: Various alpha values
	testAlphas := []basics.Int8u{64, 128, 192}
	expectedResults := []basics.Int8u{191, 127, 63} // prelerp(255, 0, alpha)

	for i, alpha := range testAlphas {
		pixfmt.CopyPixel(3, i, white)
		pixfmt.BlendPixel(3, i, black, alpha, 255)
		pixel = pixfmt.GetPixel(3, i)
		expected := expectedResults[i]
		if pixel.R != expected || pixel.G != expected || pixel.B != expected {
			t.Errorf("Alpha %d blend failed: got {%d, %d, %d}, want {%d, %d, %d}",
				alpha, pixel.R, pixel.G, pixel.B, expected, expected, expected)
		}
	}
}
