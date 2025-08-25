package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// Helper function to create a rendering buffer for RGBA64 testing
func createRGBA64Buffer(width, height int) *buffer.RenderingBufferU8 {
	bufData := make([]basics.Int8u, width*height*8) // 8 bytes per pixel (4 channels * 2 bytes each)
	return buffer.NewRenderingBufferU8WithData(bufData, width, height, width*8)
}

func TestPixFmtRGBA64Linear_Construction(t *testing.T) {
	buf := createRGBA64Buffer(100, 100)
	pf := NewPixFmtRGBA64Linear(buf)

	if pf.Width() != 100 {
		t.Errorf("Expected width 100, got %d", pf.Width())
	}
	if pf.Height() != 100 {
		t.Errorf("Expected height 100, got %d", pf.Height())
	}
}

func TestPixFmtRGBA64_ByteOrderRGBA(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	// Test RGBA order
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := pf.RowData(5)
	ptr := 5 * 8 // 5 pixels * 8 bytes per pixel

	// Check RGBA order: R=0x1234, G=0x5678, B=0x9ABC, A=0xDEF0
	// Little-endian: low byte first
	if row[ptr+0] != 0x34 || row[ptr+1] != 0x12 { // R channel
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x78 || row[ptr+3] != 0x56 { // G channel
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0xBC || row[ptr+5] != 0x9A { // B channel
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xF0 || row[ptr+7] != 0xDE { // A channel
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderARGB(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtARGB64Linear(buf)

	// Test ARGB order
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := pf.RowData(5)
	ptr := 5 * 8 // 5 pixels * 8 bytes per pixel

	// Check ARGB order: A=0xDEF0, R=0x1234, G=0x5678, B=0x9ABC
	// Little-endian: low byte first
	if row[ptr+0] != 0xF0 || row[ptr+1] != 0xDE { // A channel (position 0)
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x34 || row[ptr+3] != 0x12 { // R channel (position 1)
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x78 || row[ptr+5] != 0x56 { // G channel (position 2)
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xBC || row[ptr+7] != 0x9A { // B channel (position 3)
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderABGR(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtABGR64Linear(buf)

	// Test ABGR order
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := pf.RowData(5)
	ptr := 5 * 8 // 5 pixels * 8 bytes per pixel

	// Check ABGR order: A=0xDEF0, B=0x9ABC, G=0x5678, R=0x1234
	// Little-endian: low byte first
	if row[ptr+0] != 0xF0 || row[ptr+1] != 0xDE { // A channel (position 0)
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0xBC || row[ptr+3] != 0x9A { // B channel (position 1)
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x78 || row[ptr+5] != 0x56 { // G channel (position 2)
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0x34 || row[ptr+7] != 0x12 { // R channel (position 3)
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderBGRA(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtBGRA64Linear(buf)

	// Test BGRA order
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := pf.RowData(5)
	ptr := 5 * 8 // 5 pixels * 8 bytes per pixel

	// Check BGRA order: B=0x9ABC, G=0x5678, R=0x1234, A=0xDEF0
	// Little-endian: low byte first
	if row[ptr+0] != 0xBC || row[ptr+1] != 0x9A { // B channel (position 0)
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x78 || row[ptr+3] != 0x56 { // G channel (position 1)
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x34 || row[ptr+5] != 0x12 { // R channel (position 2)
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xF0 || row[ptr+7] != 0xDE { // A channel (position 3)
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_PixelReadBack(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(3, 4, testColor)

	readColor := pf.Pixel(3, 4)
	if readColor.R != testColor.R || readColor.G != testColor.G ||
		readColor.B != testColor.B || readColor.A != testColor.A {
		t.Errorf("Read back color doesn't match. Expected %+v, got %+v", testColor, readColor)
	}
}

func TestPixFmtRGBA64_PixelReadBackDifferentOrders(t *testing.T) {
	orders := []struct {
		name string
		pf   func(*buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		}
	}{
		{"RGBA", func(buf *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtRGBA64Linear(buf)
		}},
		{"ARGB", func(buf *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtARGB64Linear(buf)
		}},
		{"ABGR", func(buf *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtABGR64Linear(buf)
		}},
		{"BGRA", func(buf *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtBGRA64Linear(buf)
		}},
	}

	for _, order := range orders {
		t.Run(order.name, func(t *testing.T) {
			buf := createRGBA64Buffer(10, 10)
			pf := order.pf(buf)

			testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
			pf.CopyPixel(3, 4, testColor)

			readColor := pf.Pixel(3, 4)
			if readColor.R != testColor.R || readColor.G != testColor.G ||
				readColor.B != testColor.B || readColor.A != testColor.A {
				t.Errorf("Read back color doesn't match for %s order. Expected %+v, got %+v",
					order.name, testColor, readColor)
			}
		})
	}
}

func TestPixFmtRGBA64_CopyHline(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyHline(2, 5, 5, testColor) // Copy 5 pixels starting at (2,5)

	// Check all 5 pixels
	for x := 2; x < 7; x++ {
		readColor := pf.Pixel(x, 5)
		if readColor.R != testColor.R || readColor.G != testColor.G ||
			readColor.B != testColor.B || readColor.A != testColor.A {
			t.Errorf("Pixel at (%d,5) doesn't match. Expected %+v, got %+v", x, testColor, readColor)
		}
	}
}

func TestPixFmtRGBA64_BlendPixel(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	// Set background
	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyPixel(5, 5, bg)

	// Blend with 50% coverage
	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	pf.BlendPixel(5, 5, fg, 128) // 50% coverage (128/255)

	result := pf.Pixel(5, 5)

	// Result should be somewhere between background and foreground
	if result.R <= bg.R || result.R >= fg.R {
		t.Errorf("Red channel not properly blended: got %04x, expected between %04x and %04x",
			result.R, bg.R, fg.R)
	}
	if result.G >= bg.G { // Green should be reduced from background
		t.Errorf("Green channel not properly blended: got %04x, expected less than %04x",
			result.G, bg.G)
	}
}

func TestPixFmtRGBA64_BlendHline(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	// Set background
	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyHline(0, 5, 10, bg) // Fill entire row with background

	// Blend horizontal line
	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	pf.BlendHline(2, 5, 5, fg, 128) // Blend 5 pixels with 50% coverage

	// Check blended pixels
	for x := 2; x < 7; x++ {
		result := pf.Pixel(x, 5)
		if result.R <= bg.R || result.R >= fg.R {
			t.Errorf("Pixel at (%d,5) red channel not properly blended: got %04x", x, result.R)
		}
	}

	// Check non-blended pixels remain background
	result0 := pf.Pixel(0, 5)
	if result0.R != bg.R || result0.G != bg.G || result0.B != bg.B {
		t.Errorf("Pixel at (0,5) should remain background color")
	}
}

func TestPixFmtRGBA64_BlendSolidHspan(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	// Set background
	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyHline(0, 5, 10, bg)

	// Blend with varying coverage
	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	covers := []basics.Int8u{255, 192, 128, 64, 0} // Full, 75%, 50%, 25%, 0%
	pf.BlendSolidHspan(2, 5, 5, fg, covers)

	// Check that different coverage produces different results
	result2 := pf.Pixel(2, 5) // 100% coverage
	result3 := pf.Pixel(3, 5) // 75% coverage
	result4 := pf.Pixel(4, 5) // 50% coverage
	result6 := pf.Pixel(6, 5) // 0% coverage (should be unchanged)

	if result2.R <= result3.R || result3.R <= result4.R {
		t.Errorf("Coverage blending not working correctly. R values: %04x > %04x > %04x",
			result2.R, result3.R, result4.R)
	}

	if result6.R != bg.R || result6.G != bg.G || result6.B != bg.B {
		t.Errorf("Zero coverage pixel should remain unchanged")
	}
}

func TestPixFmtRGBA64_BoundaryConditions(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	// Test out of bounds - should not crash
	pf.CopyPixel(-1, 5, testColor)
	pf.CopyPixel(5, -1, testColor)
	pf.CopyPixel(10, 5, testColor)
	pf.CopyPixel(5, 10, testColor)

	// Test reading out of bounds - should return zero color
	zeroColor := pf.Pixel(-1, 5)
	if zeroColor.R != 0 || zeroColor.G != 0 || zeroColor.B != 0 || zeroColor.A != 0 {
		t.Errorf("Out of bounds read should return zero color, got %+v", zeroColor)
	}
}

func TestPixFmtRGBA64_TransparentBlending(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	// Set background
	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyPixel(5, 5, bg)

	// Try to blend transparent color
	transparent := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0x0000}
	pf.BlendPixel(5, 5, transparent, 255)

	result := pf.Pixel(5, 5)
	// Should remain unchanged since source is transparent
	if result.R != bg.R || result.G != bg.G || result.B != bg.B || result.A != bg.A {
		t.Errorf("Background should remain unchanged when blending transparent color")
	}
}

func TestPixFmtRGBA64_MakePix(t *testing.T) {
	buf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(buf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(3, 4, testColor)

	pix := pf.MakePix(3, 4)
	if len(pix) != 4 {
		t.Errorf("Expected 4 components, got %d", len(pix))
	}
	if pix[0] != testColor.R || pix[1] != testColor.G || pix[2] != testColor.B || pix[3] != testColor.A {
		t.Errorf("MakePix returned wrong values. Expected [%04x,%04x,%04x,%04x], got [%04x,%04x,%04x,%04x]",
			testColor.R, testColor.G, testColor.B, testColor.A,
			pix[0], pix[1], pix[2], pix[3])
	}
}

// Benchmark tests
func BenchmarkPixFmtRGBA64_CopyPixel(b *testing.B) {
	buf := createRGBA64Buffer(1000, 1000)
	pf := NewPixFmtRGBA64Linear(buf)
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.CopyPixel(i%1000, (i/1000)%1000, testColor)
	}
}

func BenchmarkPixFmtRGBA64_BlendPixel(b *testing.B) {
	buf := createRGBA64Buffer(1000, 1000)
	pf := NewPixFmtRGBA64Linear(buf)
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.BlendPixel(i%1000, (i/1000)%1000, testColor, 128)
	}
}
