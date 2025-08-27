package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

// Helper: create an RGBA64 rendering buffer
func createRGBA64Buffer(width, height int) *buffer.RenderingBufferU8 {
	bufData := make([]basics.Int8u, width*height*8) // 8 bytes per pixel (4 * 2 bytes)
	return buffer.NewRenderingBufferU8WithData(bufData, width, height, width*8)
}

func TestPixFmtRGBA64Linear_Construction(t *testing.T) {
	rbuf := createRGBA64Buffer(100, 100)
	pf := NewPixFmtRGBA64Linear(rbuf)

	if pf.Width() != 100 {
		t.Errorf("Expected width 100, got %d", pf.Width())
	}
	if pf.Height() != 100 {
		t.Errorf("Expected height 100, got %d", pf.Height())
	}
}

func TestPixFmtRGBA64_ByteOrderRGBA(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := buffer.RowU8(rbuf, 5)
	ptr := 5 * 8 // 5 pixels * 8 bytes

	// RGBA, little-endian
	if row[ptr+0] != 0x34 || row[ptr+1] != 0x12 {
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x78 || row[ptr+3] != 0x56 {
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0xBC || row[ptr+5] != 0x9A {
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xF0 || row[ptr+7] != 0xDE {
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderARGB(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtARGB64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := buffer.RowU8(rbuf, 5)
	ptr := 5 * 8

	// ARGB, little-endian
	if row[ptr+0] != 0xF0 || row[ptr+1] != 0xDE {
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x34 || row[ptr+3] != 0x12 {
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x78 || row[ptr+5] != 0x56 {
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xBC || row[ptr+7] != 0x9A {
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderABGR(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtABGR64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := buffer.RowU8(rbuf, 5)
	ptr := 5 * 8

	// ABGR, little-endian
	if row[ptr+0] != 0xF0 || row[ptr+1] != 0xDE {
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0xBC || row[ptr+3] != 0x9A {
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x78 || row[ptr+5] != 0x56 {
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0x34 || row[ptr+7] != 0x12 {
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_ByteOrderBGRA(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtBGRA64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyPixel(5, 5, testColor)

	row := buffer.RowU8(rbuf, 5)
	ptr := 5 * 8

	// BGRA, little-endian
	if row[ptr+0] != 0xBC || row[ptr+1] != 0x9A {
		t.Errorf("B channel incorrect: got %02x%02x, expected BC9A", row[ptr+1], row[ptr+0])
	}
	if row[ptr+2] != 0x78 || row[ptr+3] != 0x56 {
		t.Errorf("G channel incorrect: got %02x%02x, expected 7856", row[ptr+3], row[ptr+2])
	}
	if row[ptr+4] != 0x34 || row[ptr+5] != 0x12 {
		t.Errorf("R channel incorrect: got %02x%02x, expected 3412", row[ptr+5], row[ptr+4])
	}
	if row[ptr+6] != 0xF0 || row[ptr+7] != 0xDE {
		t.Errorf("A channel incorrect: got %02x%02x, expected F0DE", row[ptr+7], row[ptr+6])
	}
}

func TestPixFmtRGBA64_PixelReadBack(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

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
		{"RGBA", func(r *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtRGBA64Linear(r)
		}},
		{"ARGB", func(r *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtARGB64Linear(r)
		}},
		{"ABGR", func(r *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtABGR64Linear(r)
		}},
		{"BGRA", func(r *buffer.RenderingBufferU8) interface {
			CopyPixel(x, y int, c color.RGBA16[color.Linear])
			Pixel(x, y int) color.RGBA16[color.Linear]
		} {
			return NewPixFmtBGRA64Linear(r)
		}},
	}

	for _, ord := range orders {
		t.Run(ord.name, func(t *testing.T) {
			rbuf := createRGBA64Buffer(10, 10)
			pf := ord.pf(rbuf)

			testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
			pf.CopyPixel(3, 4, testColor)
			readColor := pf.Pixel(3, 4)

			if readColor.R != testColor.R || readColor.G != testColor.G ||
				readColor.B != testColor.B || readColor.A != testColor.A {
				t.Errorf("Read back color doesn't match for %s. Expected %+v, got %+v",
					ord.name, testColor, readColor)
			}
		})
	}
}

func TestPixFmtRGBA64_CopyHline(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}
	pf.CopyHline(2, 5, 5, testColor)

	for x := 2; x < 7; x++ {
		readColor := pf.Pixel(x, 5)
		if readColor.R != testColor.R || readColor.G != testColor.G ||
			readColor.B != testColor.B || readColor.A != testColor.A {
			t.Errorf("Pixel at (%d,5) doesn't match. Expected %+v, got %+v", x, testColor, readColor)
		}
	}
}

func TestPixFmtRGBA64_BlendPixel(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyPixel(5, 5, bg)

	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	pf.BlendPixel(5, 5, fg, 128)

	result := pf.Pixel(5, 5)
	if result.R <= bg.R || result.R >= fg.R {
		t.Errorf("Red not blended: got %04x, expect between %04x and %04x", result.R, bg.R, fg.R)
	}
	if result.G >= bg.G {
		t.Errorf("Green not blended down from background: got %04x, bg %04x", result.G, bg.G)
	}
}

func TestPixFmtRGBA64_BlendHline(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyHline(0, 5, 10, bg)

	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	pf.BlendHline(2, 5, 5, fg, 128)

	for x := 2; x < 7; x++ {
		result := pf.Pixel(x, 5)
		if result.R <= bg.R || result.R >= fg.R {
			t.Errorf("Pixel (%d,5) red not blended: got %04x", x, result.R)
		}
	}

	result0 := pf.Pixel(0, 5)
	if result0.R != bg.R || result0.G != bg.G || result0.B != bg.B {
		t.Errorf("Pixel (0,5) should remain background")
	}
}

func TestPixFmtRGBA64_BlendSolidHspan(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyHline(0, 5, 10, bg)

	fg := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0xFFFF}
	covers := []basics.Int8u{255, 192, 128, 64, 0}
	pf.BlendSolidHspan(2, 5, 5, fg, covers)

	r2 := pf.Pixel(2, 5) // 100%
	r3 := pf.Pixel(3, 5) // 75%
	r4 := pf.Pixel(4, 5) // 50%
	r6 := pf.Pixel(6, 5) // untouched

	if !(r2.R > r3.R && r3.R > r4.R) {
		t.Errorf("Coverage blending wrong. R: %04x > %04x > %04x expected", r2.R, r3.R, r4.R)
	}
	if r6.R != bg.R || r6.G != bg.G || r6.B != bg.B {
		t.Errorf("Zero coverage pixel should remain unchanged")
	}
}

func TestPixFmtRGBA64_BoundaryConditions(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	// Out of bounds writes shouldn't panic
	pf.CopyPixel(-1, 5, testColor)
	pf.CopyPixel(5, -1, testColor)
	pf.CopyPixel(10, 5, testColor)
	pf.CopyPixel(5, 10, testColor)

	// Out of bounds read should be zero
	zero := pf.Pixel(-1, 5)
	if zero.R != 0 || zero.G != 0 || zero.B != 0 || zero.A != 0 {
		t.Errorf("Out-of-bounds read should be zero, got %+v", zero)
	}
}

func TestPixFmtRGBA64_TransparentBlending(t *testing.T) {
	rbuf := createRGBA64Buffer(10, 10)
	pf := NewPixFmtRGBA64Linear(rbuf)

	bg := color.RGBA16[color.Linear]{R: 0x8000, G: 0x8000, B: 0x8000, A: 0xFFFF}
	pf.CopyPixel(5, 5, bg)

	transparent := color.RGBA16[color.Linear]{R: 0xFFFF, G: 0x0000, B: 0x0000, A: 0x0000}
	pf.BlendPixel(5, 5, transparent, 255)

	result := pf.Pixel(5, 5)
	if result != bg {
		t.Errorf("Background should remain unchanged when blending transparent color")
	}
}

// Benchmarks

func BenchmarkPixFmtRGBA64_CopyPixel(b *testing.B) {
	rbuf := createRGBA64Buffer(1000, 1000)
	pf := NewPixFmtRGBA64Linear(rbuf)
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.CopyPixel(i%1000, (i/1000)%1000, testColor)
	}
}

func BenchmarkPixFmtRGBA64_BlendPixel(b *testing.B) {
	rbuf := createRGBA64Buffer(1000, 1000)
	pf := NewPixFmtRGBA64Linear(rbuf)
	testColor := color.RGBA16[color.Linear]{R: 0x1234, G: 0x5678, B: 0x9ABC, A: 0xDEF0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.BlendPixel(i%1000, (i/1000)%1000, testColor, 128)
	}
}
