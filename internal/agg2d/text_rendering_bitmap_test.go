package agg2d

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/font"
	isc "github.com/cwbudde/agg_go/internal/scanline"
)

func pixelAt(buf []byte, width, x, y int) (r, g, b, a uint8) {
	idx := (y*width + x) * 4
	return buf[idx], buf[idx+1], buf[idx+2], buf[idx+3]
}

type testScanlineU8StorageWrapper struct {
	sl *isc.ScanlineU8
}

func (w testScanlineU8StorageWrapper) Y() int        { return w.sl.Y() }
func (w testScanlineU8StorageWrapper) NumSpans() int { return w.sl.NumSpans() }
func (w testScanlineU8StorageWrapper) ResetSpans()   { w.sl.ResetSpans() }
func (w testScanlineU8StorageWrapper) AddSpan(x, length int, cover basics.Int8u) {
	w.sl.AddSpan(x, length, uint(cover))
}

func (w testScanlineU8StorageWrapper) AddCells(x, length int, covers []basics.Int8u) {
	for i := 0; i < length && i < len(covers); i++ {
		w.sl.AddCell(x+i, uint(covers[i]))
	}
}
func (w testScanlineU8StorageWrapper) Finalize(y int) { w.sl.Finalize(y) }
func (w testScanlineU8StorageWrapper) Begin() isc.ScanlineIterator {
	return w.sl.BeginIterator()
}

func makeSerializedGray8Adaptor(x0, y0 int, rows [][]uint8) *font.SerializedScanlinesAdaptorAA {
	storage := isc.NewScanlineStorageAA[basics.Int8u]()
	sl := isc.NewScanlineU8()
	storage.Prepare()

	maxWidth := 0
	for _, row := range rows {
		if len(row) > maxWidth {
			maxWidth = len(row)
		}
	}
	sl.Reset(x0, x0+maxWidth)

	for rowIdx, row := range rows {
		sl.ResetSpans()
		for col, cov := range row {
			if cov != 0 {
				sl.AddCell(x0+col, uint(cov))
			}
		}
		if sl.NumSpans() > 0 {
			sl.Finalize(y0 + rowIdx)
			storage.Render(testScanlineU8StorageWrapper{sl: sl})
		}
	}

	data := make([]byte, storage.ByteSize())
	storage.Serialize(data)
	bounds := basics.Rect[int]{X1: storage.MinX(), Y1: storage.MinY(), X2: storage.MaxX() + 1, Y2: storage.MaxY() + 1}
	return font.NewSerializedScanlinesAdaptorAA(data, bounds)
}

func makeSerializedMonoAdaptor(x0, y0 int, width int, bits [][]bool) *font.SerializedScanlinesAdaptorBin {
	storage := isc.NewScanlineStorageBin()
	sl := isc.NewScanlineBin()
	storage.Prepare()
	sl.Reset(x0, x0+width)

	for rowIdx, row := range bits {
		sl.ResetSpans()
		for col := 0; col < width && col < len(row); col++ {
			if row[col] {
				sl.AddCell(x0+col, 0)
			}
		}
		if sl.NumSpans() > 0 {
			sl.Finalize(y0 + rowIdx)
			storage.RenderBinScanline(sl)
		}
	}

	data := make([]byte, storage.ByteSize())
	storage.Serialize(data)
	bounds := basics.Rect[int]{X1: storage.MinX(), Y1: storage.MinY(), X2: storage.MaxX() + 1, Y2: storage.MaxY() + 1}
	return font.NewSerializedScanlinesAdaptorBin(data, bounds)
}

func TestRenderGlyphScanlinesGray8UsesCoverage(t *testing.T) {
	agg2d := NewAgg2D()
	width, height := 12, 8
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.ClearAll(Color{0, 0, 0, 0})
	agg2d.FillColor(Color{255, 0, 0, 255})

	adaptor := makeSerializedGray8Adaptor(2, 3, [][]uint8{
		{255, 0, 255},
		{0, 255, 0},
	})

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

	adaptor := makeSerializedMonoAdaptor(4, 2, 8, [][]bool{
		{true, false, true, false, true, false, true, false},
	})

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

func TestRenderGlyphScanlinesUsesGeneralBlendMode(t *testing.T) {
	agg2d := NewAgg2D()
	width, height := 4, 4
	buf := make([]byte, width*height*4)
	agg2d.Attach(buf, width, height, width*4)
	agg2d.FillColor(Color{255, 0, 0, 255})

	adaptor := makeSerializedGray8Adaptor(1, 1, [][]uint8{{255}})
	glyph := &font.GlyphCache{DataType: font.GlyphDataGray8}

	agg2d.SetBlendMode(BlendAlpha)
	agg2d.ClearAll(Color{0, 255, 0, 255})
	agg2d.renderGlyphScanlines(adaptor, glyph, 0, 0)
	r, g, b, a := pixelAt(buf, width, 1, 1)
	if r != 255 || g != 0 || b != 0 || a != 255 {
		t.Fatalf("alpha blend expected pure source red, got rgba(%d,%d,%d,%d)", r, g, b, a)
	}

	agg2d.SetBlendMode(BlendMultiply)
	agg2d.ClearAll(Color{0, 255, 0, 255})
	agg2d.renderGlyphScanlines(adaptor, glyph, 0, 0)
	r, g, b, a = pixelAt(buf, width, 1, 1)
	if r == 255 && g == 0 && b == 0 {
		t.Fatalf("multiply blend should not match pure source red, got rgba(%d,%d,%d,%d)", r, g, b, a)
	}
}
