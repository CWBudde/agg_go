package agg2d

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/font"
)

func pixelAt(buf []byte, width, x, y int) (r, g, b, a uint8) {
	idx := (y*width + x) * 4
	return buf[idx], buf[idx+1], buf[idx+2], buf[idx+3]
}

func TestRenderGlyphScanlinesGray8UsesCoverage(t *testing.T) {
	agg2d := NewAgg2D()
	width, height := 12, 8
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.ClearAll(Color{0, 0, 0, 0})
	agg2d.FillColor(Color{255, 0, 0, 255})

	// 3x2 glyph bitmap coverage, row-major.
	// Row 0: 255,0,255
	// Row 1: 0,255,0
	data := []byte{
		255, 0, 255,
		0, 255, 0,
	}
	bounds := basics.Rect[int]{X1: 2, Y1: 3, X2: 5, Y2: 5}
	adaptor := font.NewSerializedScanlinesAdaptorAA(data, bounds)

	glyph := &font.GlyphCache{DataType: font.GlyphDataGray8}
	agg2d.renderGlyphScanlines(adaptor, glyph, 0, 0)

	// Covered pixels should be written.
	_, _, _, a := pixelAt(buf, width, 2, 3)
	if a == 0 {
		t.Fatalf("expected covered pixel at (2,3)")
	}
	_, _, _, a = pixelAt(buf, width, 4, 3)
	if a == 0 {
		t.Fatalf("expected covered pixel at (4,3)")
	}
	_, _, _, a = pixelAt(buf, width, 3, 4)
	if a == 0 {
		t.Fatalf("expected covered pixel at (3,4)")
	}

	// Zero-coverage pixels must remain untouched (would fail with rectangle fallback).
	_, _, _, a = pixelAt(buf, width, 3, 3)
	if a != 0 {
		t.Fatalf("expected zero-coverage pixel at (3,3) to remain transparent, got alpha=%d", a)
	}
	_, _, _, a = pixelAt(buf, width, 2, 4)
	if a != 0 {
		t.Fatalf("expected zero-coverage pixel at (2,4) to remain transparent, got alpha=%d", a)
	}
}

func TestRenderGlyphScanlinesMonoDecodesBits(t *testing.T) {
	agg2d := NewAgg2D()
	width, height := 16, 6
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.ClearAll(Color{0, 0, 0, 0})
	agg2d.FillColor(Color{0, 255, 0, 255})

	// 8x1 monochrome row, 0b10101010 (MSB-first).
	data := []byte{0xAA}
	bounds := basics.Rect[int]{X1: 4, Y1: 2, X2: 12, Y2: 3}
	adaptor := font.NewSerializedScanlinesAdaptorBin(data, bounds)

	glyph := &font.GlyphCache{DataType: font.GlyphDataMono}
	agg2d.renderGlyphScanlines(adaptor, glyph, 0, 0)

	// Set bits: columns 0,2,4,6.
	setX := []int{4, 6, 8, 10}
	for _, x := range setX {
		_, _, _, a := pixelAt(buf, width, x, 2)
		if a == 0 {
			t.Fatalf("expected set mono bit at x=%d", x)
		}
	}

	// Clear bits: columns 1,3,5,7.
	clearX := []int{5, 7, 9, 11}
	for _, x := range clearX {
		_, _, _, a := pixelAt(buf, width, x, 2)
		if a != 0 {
			t.Fatalf("expected clear mono bit at x=%d, got alpha=%d", x, a)
		}
	}
}
