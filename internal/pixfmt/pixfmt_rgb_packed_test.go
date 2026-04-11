package pixfmt

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/renderer"
)

// Compile-time interface satisfaction checks.
var (
	_ renderer.PixelFormat[color.RGBA8[color.Linear]] = (*PixFmtRGB555[blender.BlenderRGB555])(nil)
	_ renderer.PixelFormat[color.RGBA8[color.Linear]] = (*PixFmtRGB565[blender.BlenderRGB565])(nil)
	_ renderer.PixelFormat[color.RGBA8[color.Linear]] = (*PixFmtBGR555[blender.BlenderBGR555])(nil)
	_ renderer.PixelFormat[color.RGBA8[color.Linear]] = (*PixFmtBGR565[blender.BlenderBGR565])(nil)
)

// Test RGB555 packing/unpacking accuracy
func TestRGB555PackUnpack(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b basics.Int8u
	}{
		{"Black", 0, 0, 0},
		{"White", 255, 255, 255},
		{"Red", 255, 0, 0},
		{"Green", 0, 255, 0},
		{"Blue", 0, 0, 255},
		{"Gray", 128, 128, 128},
		{"Custom", 123, 89, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack and unpack
			packed := blender.MakePixel555(tt.r, tt.g, tt.b)
			ur, ug, ub := blender.UnpackPixel555(packed)

			// Check that unpacked values are within precision limits
			// RGB555 has 5 bits per component, so we lose 3 bits of precision
			if !within555Precision(tt.r, ur) {
				t.Errorf("Red precision loss too high: original=%d, unpacked=%d", tt.r, ur)
			}
			if !within555Precision(tt.g, ug) {
				t.Errorf("Green precision loss too high: original=%d, unpacked=%d", tt.g, ug)
			}
			if !within555Precision(tt.b, ub) {
				t.Errorf("Blue precision loss too high: original=%d, unpacked=%d", tt.b, ub)
			}

			// Verify the unused bit is set
			if packed&0x8000 == 0 {
				t.Error("Unused bit should be set in RGB555")
			}
		})
	}
}

// Test RGB565 packing/unpacking accuracy
func TestRGB565PackUnpack(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b basics.Int8u
	}{
		{"Black", 0, 0, 0},
		{"White", 255, 255, 255},
		{"Red", 255, 0, 0},
		{"Green", 0, 255, 0},
		{"Blue", 0, 0, 255},
		{"Gray", 128, 128, 128},
		{"Custom", 123, 89, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pack and unpack
			packed := blender.MakePixel565(tt.r, tt.g, tt.b)
			ur, ug, ub := blender.UnpackPixel565(packed)

			// Check that unpacked values are within precision limits
			// RGB565: Red/Blue have 5 bits (lose 3), Green has 6 bits (lose 2)
			if !within565RedBluePrecision(tt.r, ur) {
				t.Errorf("Red precision loss too high: original=%d, unpacked=%d", tt.r, ur)
			}
			if !within565GreenPrecision(tt.g, ug) {
				t.Errorf("Green precision loss too high: original=%d, unpacked=%d", tt.g, ug)
			}
			if !within565RedBluePrecision(tt.b, ub) {
				t.Errorf("Blue precision loss too high: original=%d, unpacked=%d", tt.b, ub)
			}
		})
	}
}

// Test BGR555 packing/unpacking
func TestBGR555PackUnpack(t *testing.T) {
	r, g, b := basics.Int8u(123), basics.Int8u(89), basics.Int8u(200)

	packed := MakePixelBGR555(r, g, b)
	ur, ug, ub := UnpackPixelBGR555(packed)

	if !within555Precision(r, ur) {
		t.Errorf("Red precision loss too high: original=%d, unpacked=%d", r, ur)
	}
	if !within555Precision(g, ug) {
		t.Errorf("Green precision loss too high: original=%d, unpacked=%d", g, ug)
	}
	if !within555Precision(b, ub) {
		t.Errorf("Blue precision loss too high: original=%d, unpacked=%d", b, ub)
	}
}

// Test BGR565 packing/unpacking
func TestBGR565PackUnpack(t *testing.T) {
	r, g, b := basics.Int8u(123), basics.Int8u(89), basics.Int8u(200)

	packed := MakePixelBGR565(r, g, b)
	ur, ug, ub := UnpackPixelBGR565(packed)

	if !within565RedBluePrecision(r, ur) {
		t.Errorf("Red precision loss too high: original=%d, unpacked=%d", r, ur)
	}
	if !within565GreenPrecision(g, ug) {
		t.Errorf("Green precision loss too high: original=%d, unpacked=%d", g, ug)
	}
	if !within565RedBluePrecision(b, ub) {
		t.Errorf("Blue precision loss too high: original=%d, unpacked=%d", b, ub)
	}
}

// Test RGB555 pixel format functionality
func TestPixFmtRGB555Basic(t *testing.T) {
	// Create a small buffer
	width, height := 4, 4
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pixfmt := NewPixFmtRGB555(rbuf, blender.BlenderRGB555{})

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width mismatch: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.Height() != height {
		t.Errorf("Height mismatch: got %d, want %d", pixfmt.Height(), height)
	}
	if pixfmt.PixWidth() != 2 {
		t.Errorf("PixWidth should be 2 bytes for 16-bit format")
	}

	// Test pixel copy and retrieval
	testColor := color.RGBA8[color.Linear]{R: 255, G: 128, B: 64, A: 255}
	pixfmt.CopyPixel(1, 1, testColor)

	retrieved := pixfmt.Pixel(1, 1)

	// Check within precision limits
	if !within555Precision(testColor.R, retrieved.R) {
		t.Errorf("Red precision loss: expected ~%d, got %d", testColor.R, retrieved.R)
	}
	if !within555Precision(testColor.G, retrieved.G) {
		t.Errorf("Green precision loss: expected ~%d, got %d", testColor.G, retrieved.G)
	}
	if !within555Precision(testColor.B, retrieved.B) {
		t.Errorf("Blue precision loss: expected ~%d, got %d", testColor.B, retrieved.B)
	}
}

// Test RGB565 pixel format functionality
func TestPixFmtRGB565Basic(t *testing.T) {
	// Create a small buffer
	width, height := 4, 4
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pixfmt := NewPixFmtRGB565(rbuf, blender.BlenderRGB565{})

	// Test basic properties
	if pixfmt.Width() != width {
		t.Errorf("Width mismatch: got %d, want %d", pixfmt.Width(), width)
	}
	if pixfmt.Height() != height {
		t.Errorf("Height mismatch: got %d, want %d", pixfmt.Height(), height)
	}
	if pixfmt.PixWidth() != 2 {
		t.Errorf("PixWidth should be 2 bytes for 16-bit format")
	}

	// Test pixel copy and retrieval
	testColor := color.RGBA8[color.Linear]{R: 255, G: 128, B: 64, A: 255}
	pixfmt.CopyPixel(1, 1, testColor)

	retrieved := pixfmt.Pixel(1, 1)

	// Check within precision limits
	if !within565RedBluePrecision(testColor.R, retrieved.R) {
		t.Errorf("Red precision loss: expected ~%d, got %d", testColor.R, retrieved.R)
	}
	if !within565GreenPrecision(testColor.G, retrieved.G) {
		t.Errorf("Green precision loss: expected ~%d, got %d", testColor.G, retrieved.G)
	}
	if !within565RedBluePrecision(testColor.B, retrieved.B) {
		t.Errorf("Blue precision loss: expected ~%d, got %d", testColor.B, retrieved.B)
	}
}

// Test RGB555 blending
func TestBlenderRGB555(t *testing.T) {
	bl := blender.BlenderRGB555{}

	// Start with a known pixel
	pixel := blender.MakePixel555(100, 150, 200)

	// Blend with 50% alpha, no coverage modulation
	bl.BlendPix(&pixel, 200, 100, 50, 128, 255)

	// Extract the result
	r, g, b := blender.UnpackPixel555(pixel)

	// Should be approximately halfway between original and new values
	// Due to precision loss, we check within reasonable bounds
	expectedR := (100 + 200) / 2 // ~150
	expectedG := (150 + 100) / 2 // ~125
	expectedB := (200 + 50) / 2  // ~125

	if !within555Precision(basics.Int8u(expectedR), r) {
		t.Errorf("Blended red not as expected: got %d, expected ~%d", r, expectedR)
	}
	if !within555Precision(basics.Int8u(expectedG), g) {
		t.Errorf("Blended green not as expected: got %d, expected ~%d", g, expectedG)
	}
	if !within555Precision(basics.Int8u(expectedB), b) {
		t.Errorf("Blended blue not as expected: got %d, expected ~%d", b, expectedB)
	}
}

// Test RGB565 blending
func TestBlenderRGB565(t *testing.T) {
	bl := blender.BlenderRGB565{}

	// Start with a known pixel
	pixel := blender.MakePixel565(100, 150, 200)

	// Blend with 50% alpha, no coverage modulation
	bl.BlendPix(&pixel, 200, 100, 50, 128, 255)

	// Extract the result
	r, g, b := blender.UnpackPixel565(pixel)

	// Should be approximately halfway between original and new values
	expectedR := (100 + 200) / 2 // ~150
	expectedG := (150 + 100) / 2 // ~125
	expectedB := (200 + 50) / 2  // ~125

	if !within565RedBluePrecision(basics.Int8u(expectedR), r) {
		t.Errorf("Blended red not as expected: got %d, expected ~%d", r, expectedR)
	}
	if !within565GreenPrecision(basics.Int8u(expectedG), g) {
		t.Errorf("Blended green not as expected: got %d, expected ~%d", g, expectedG)
	}
	if !within565RedBluePrecision(basics.Int8u(expectedB), b) {
		t.Errorf("Blended blue not as expected: got %d, expected ~%d", b, expectedB)
	}
}

// Test bounds checking
func TestPackedFormatBounds(t *testing.T) {
	width, height := 4, 4
	buf := make([]basics.Int16u, width*height)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, width, height, width*2)
	pixfmt := NewPixFmtRGB555(rbuf, blender.BlenderRGB555{})

	// Test out-of-bounds access (should not panic)
	testColor := color.RGBA8[color.Linear]{R: 255, G: 128, B: 64, A: 255}

	// These should not panic
	pixfmt.CopyPixel(-1, 0, testColor)
	pixfmt.CopyPixel(0, -1, testColor)
	pixfmt.CopyPixel(width, 0, testColor)
	pixfmt.CopyPixel(0, height, testColor)

	retrieved := pixfmt.Pixel(-1, 0)
	if retrieved.R != 0 || retrieved.G != 0 || retrieved.B != 0 {
		t.Error("Out-of-bounds Pixel should return zero")
	}
}

// Test color conversion functions
func TestPackedColorConversion(t *testing.T) {
	testColor := color.RGBA8[color.Linear]{R: 123, G: 89, B: 200, A: 255}

	// RGB555
	pixel555 := MakePixel555(testColor.R, testColor.G, testColor.B)
	converted555 := blender.MakeColorRGB555(pixel555)
	if !within555Precision(testColor.R, converted555.R) ||
		!within555Precision(testColor.G, converted555.G) ||
		!within555Precision(testColor.B, converted555.B) {
		t.Error("RGB555 color conversion failed precision check")
	}

	// RGB565
	pixel565 := MakePixel565(testColor.R, testColor.G, testColor.B)
	converted565 := blender.MakeColorRGB565(pixel565)
	if !within565RedBluePrecision(testColor.R, converted565.R) ||
		!within565GreenPrecision(testColor.G, converted565.G) ||
		!within565RedBluePrecision(testColor.B, converted565.B) {
		t.Error("RGB565 color conversion failed precision check")
	}

	// BGR555
	pixelBGR555 := MakePixelBGR555(testColor.R, testColor.G, testColor.B)
	convertedBGR555 := blender.MakeColorBGR555(pixelBGR555)
	if !within555Precision(testColor.R, convertedBGR555.R) ||
		!within555Precision(testColor.G, convertedBGR555.G) ||
		!within555Precision(testColor.B, convertedBGR555.B) {
		t.Error("BGR555 color conversion failed precision check")
	}

	// BGR565
	pixelBGR565 := MakePixelBGR565(testColor.R, testColor.G, testColor.B)
	convertedBGR565 := blender.MakeColorBGR565(pixelBGR565)
	if !within565RedBluePrecision(testColor.R, convertedBGR565.R) ||
		!within565GreenPrecision(testColor.G, convertedBGR565.G) ||
		!within565RedBluePrecision(testColor.B, convertedBGR565.B) {
		t.Error("BGR565 color conversion failed precision check")
	}
}

// Helper functions to check precision within acceptable limits

// RGB555 has 5 bits per component, so precision loss is up to 7 (2^3 - 1)
func within555Precision(original, unpacked basics.Int8u) bool {
	diff := int(original) - int(unpacked)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 7
}

// RGB565 red/blue have 5 bits, precision loss up to 7
func within565RedBluePrecision(original, unpacked basics.Int8u) bool {
	diff := int(original) - int(unpacked)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 7
}

// RGB565 green has 6 bits, precision loss up to 3 (2^2 - 1)
func within565GreenPrecision(original, unpacked basics.Int8u) bool {
	diff := int(original) - int(unpacked)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 3
}

// Benchmark tests for performance comparison
func BenchmarkRGB555Pack(b *testing.B) {
	r, g, bl := basics.Int8u(123), basics.Int8u(89), basics.Int8u(200)

	for i := 0; i < b.N; i++ {
		_ = MakePixel555(r, g, bl)
	}
}

func BenchmarkRGB555Unpack(b *testing.B) {
	pixel := MakePixel555(123, 89, 200)

	for i := 0; i < b.N; i++ {
		_, _, _ = UnpackPixel555(pixel)
	}
}

func BenchmarkRGB565Pack(b *testing.B) {
	r, g, bl := basics.Int8u(123), basics.Int8u(89), basics.Int8u(200)

	for i := 0; i < b.N; i++ {
		_ = MakePixel565(r, g, bl)
	}
}

func BenchmarkRGB565Unpack(b *testing.B) {
	pixel := MakePixel565(123, 89, 200)

	for i := 0; i < b.N; i++ {
		_, _, _ = UnpackPixel565(pixel)
	}
}

func BenchmarkBlenderRGB555(b *testing.B) {
	bl := blender.BlenderRGB555{}
	pixel := blender.MakePixel555(100, 150, 200)

	for i := 0; i < b.N; i++ {
		bl.BlendPix(&pixel, 200, 100, 50, 128, 255)
	}
}

func BenchmarkBlenderRGB565(b *testing.B) {
	bl := blender.BlenderRGB565{}
	pixel := blender.MakePixel565(100, 150, 200)

	for i := 0; i < b.N; i++ {
		bl.BlendPix(&pixel, 200, 100, 50, 128, 255)
	}
}

func BenchmarkRGB565BlendSolidHspan(b *testing.B) {
	w, h := 1000, 1000
	pf, _ := makeRGB565Buf(w, h)
	c := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 128}
	covers := make([]basics.Int8u, w)
	for i := range covers {
		covers[i] = 128 // half coverage
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < h; y++ {
			pf.BlendSolidHspan(0, y, w, c, covers)
		}
	}
}

// --- Contract tests ---

func makeRGB565Buf(w, h int) (*PixFmtRGB565[blender.BlenderRGB565], []basics.Int16u) {
	buf := make([]basics.Int16u, w*h)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, w, h, w*2)
	return NewPixFmtRGB565(rbuf, blender.BlenderRGB565{}), buf
}

func makeRGB555Buf(w, h int) (*PixFmtRGB555[blender.BlenderRGB555], []basics.Int16u) {
	buf := make([]basics.Int16u, w*h)
	rbuf := buffer.NewRenderingBufferU16WithData(buf, w, h, w*2)
	return NewPixFmtRGB555(rbuf, blender.BlenderRGB555{}), buf
}

func TestPixFmtRGB555Clear(t *testing.T) {
	pf, buf := makeRGB555Buf(8, 4)
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	pf.Clear(white)

	expected := MakePixel555(255, 255, 255)
	for i, pix := range buf {
		if pix != expected {
			t.Errorf("pixel %d: got 0x%04X, want 0x%04X", i, pix, expected)
			break
		}
	}
}

func TestPixFmtRGB555CopyHline(t *testing.T) {
	pf, buf := makeRGB555Buf(8, 4)
	pf.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	red := color.RGBA8[color.Linear]{R: 248, G: 0, B: 0, A: 255}
	pf.CopyHline(1, 0, 5, red)

	// Pixels 1–5 in row 0 should be red.
	expectedRed := MakePixel555(248, 0, 0)
	for x := 1; x <= 5; x++ {
		if buf[x] != expectedRed {
			t.Errorf("x=%d: got 0x%04X, want 0x%04X", x, buf[x], expectedRed)
		}
	}
	// Pixels 0 and 6 should still be black.
	if buf[0] != MakePixel555(0, 0, 0) {
		t.Errorf("x=0 should be black, got 0x%04X", buf[0])
	}
	if buf[6] != MakePixel555(0, 0, 0) {
		t.Errorf("x=6 should be black, got 0x%04X", buf[6])
	}
}

func TestPixFmtRGB555BlendSolidHspan_FullCover(t *testing.T) {
	pf, _ := makeRGB555Buf(8, 4)
	pf.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})

	c := color.RGBA8[color.Linear]{R: 248, G: 0, B: 0, A: 255}
	covers := []basics.Int8u{255, 255, 255}
	pf.BlendSolidHspan(2, 1, 3, c, covers)

	// With opaque color + full coverage, blend ≈ copy.
	for x := 2; x <= 4; x++ {
		got := pf.Pixel(x, 1)
		if !within555Precision(248, got.R) || got.G != 0 || got.B != 0 {
			t.Errorf("x=%d: got {%d,%d,%d}, want ~{248,0,0}", x, got.R, got.G, got.B)
		}
	}
}

func TestPixFmtRGB555BlendPixel_ZeroCoverNoOp(t *testing.T) {
	pf, _ := makeRGB555Buf(4, 4)
	orig := color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 255}
	pf.CopyPixel(2, 2, orig)

	// Blending with cover=0 must not modify the pixel.
	pf.BlendPixel(2, 2, color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}, 0)
	got := pf.Pixel(2, 2)
	if !within555Precision(orig.R, got.R) || !within555Precision(orig.G, got.G) || !within555Precision(orig.B, got.B) {
		t.Errorf("BlendPixel with cover=0 modified pixel: got {%d,%d,%d}", got.R, got.G, got.B)
	}
}
