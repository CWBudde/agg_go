//go:build freetype

package freetype2

import (
	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/scanline"
)

// Gray8AdaptorWrapper is a thin Go adapter from the scanline package's
// serialized AA adaptor to the separate internal/fonts fman interface.
// It is an intentional package-boundary wrapper, not a direct AGG type.
type Gray8AdaptorWrapper struct {
	adaptor          *scanline.SerializedScanlinesAdaptorAA[uint8]
	embeddedScanline *scanline.SerializedEmbeddedScanline[uint8]
	bounds           basics.Rect[int]
}

// NewGray8AdaptorWrapper creates a new Gray8AdaptorWrapper wrapping the given scanline adaptor.
func NewGray8AdaptorWrapper(adaptor *scanline.SerializedScanlinesAdaptorAA[uint8]) *Gray8AdaptorWrapper {
	return &Gray8AdaptorWrapper{
		adaptor:          adaptor,
		embeddedScanline: scanline.NewSerializedEmbeddedScanline[uint8](),
	}
}

// InitGlyph initializes the adaptor with glyph data and position.
func (w *Gray8AdaptorWrapper) InitGlyph(data []byte, dataSize uint32, x, y float64) {
	if w.adaptor == nil || data == nil {
		return
	}

	// Initialize the underlying adaptor with the glyph data
	w.adaptor.Init(data, int(dataSize), x, y)

	// Rewind to calculate bounds
	w.adaptor.RewindScanlines()

	// Calculate bounds based on the adaptor's bounds methods
	w.bounds = basics.Rect[int]{
		X1: w.adaptor.MinX(),
		Y1: w.adaptor.MinY(),
		X2: w.adaptor.MaxX(),
		Y2: w.adaptor.MaxY(),
	}
}

// Bounds returns the bounding rectangle of the glyph.
func (w *Gray8AdaptorWrapper) Bounds() basics.Rect[int] {
	return w.bounds
}

// Rewind prepares the adaptor for scanline iteration.
func (w *Gray8AdaptorWrapper) Rewind(pathID uint) {
	if w.adaptor != nil {
		w.adaptor.RewindScanlines()
	}
}

// SweepScanline returns the next scanline.
func (w *Gray8AdaptorWrapper) SweepScanline() bool {
	if w.adaptor != nil && w.embeddedScanline != nil {
		return w.adaptor.SweepSerializedEmbeddedScanline(w.embeddedScanline)
	}
	return false
}

// NumSpans returns the number of spans in current scanline.
func (w *Gray8AdaptorWrapper) NumSpans() uint {
	if w.embeddedScanline != nil {
		return uint(w.embeddedScanline.NumSpans())
	}
	return 0
}

// Begin returns iterator for the first span.
func (w *Gray8AdaptorWrapper) Begin() fonts.Gray8SpanIterator {
	if w.embeddedScanline != nil {
		iter := w.embeddedScanline.Begin()
		return &Gray8SpanIteratorWrapper{iter: iter}
	}
	return &Gray8SpanIteratorWrapper{}
}

// Gray8SpanIteratorWrapper adapts the scanline span iterator to the Fman interface.
type Gray8SpanIteratorWrapper struct {
	iter *scanline.SerializedEmbeddedScanlineIterator[uint8]
}

// Next advances to the next span.
func (w *Gray8SpanIteratorWrapper) Next() {
	if w.iter != nil {
		w.iter.Next()
	}
}

// IsValid returns true if the iterator is at a valid span.
func (w *Gray8SpanIteratorWrapper) IsValid() bool {
	if w.iter != nil {
		return w.iter.IsValid()
	}
	return false
}

// X returns the starting X coordinate of current span.
func (w *Gray8SpanIteratorWrapper) X() int {
	if w.iter != nil {
		return w.iter.X()
	}
	return 0
}

// Len returns the length of current span.
func (w *Gray8SpanIteratorWrapper) Len() int {
	if w.iter != nil {
		return w.iter.Len()
	}
	return 0
}

// Covers returns the coverage array for current span.
func (w *Gray8SpanIteratorWrapper) Covers() []uint8 {
	if w.iter != nil {
		return w.iter.Covers()
	}
	return nil
}

// MonoAdaptorWrapper is a thin Go adapter from the scanline package's
// serialized binary adaptor to the separate internal/fonts fman interface.
// It is an intentional package-boundary wrapper, not a direct AGG type.
type MonoAdaptorWrapper struct {
	adaptor          *scanline.SerializedScanlinesAdaptorBin
	embeddedScanline *scanline.EmbeddedScanlineSerial
	bounds           basics.Rect[int]
}

// NewMonoAdaptorWrapper creates a new MonoAdaptorWrapper wrapping the given scanline adaptor.
func NewMonoAdaptorWrapper(adaptor *scanline.SerializedScanlinesAdaptorBin) *MonoAdaptorWrapper {
	return &MonoAdaptorWrapper{
		adaptor:          adaptor,
		embeddedScanline: scanline.NewEmbeddedScanlineSerial(),
	}
}

// InitGlyph initializes the adaptor with glyph data and position.
func (w *MonoAdaptorWrapper) InitGlyph(data []byte, dataSize uint32, x, y float64) {
	if w.adaptor == nil || data == nil {
		return
	}

	// Initialize the underlying adaptor with the glyph data
	w.adaptor.Init(data, x, y)

	// Rewind to calculate bounds
	w.adaptor.RewindScanlines()

	// Calculate bounds based on the adaptor's bounds methods
	w.bounds = basics.Rect[int]{
		X1: w.adaptor.MinX(),
		Y1: w.adaptor.MinY(),
		X2: w.adaptor.MaxX(),
		Y2: w.adaptor.MaxY(),
	}
}

// Bounds returns the bounding rectangle of the glyph.
func (w *MonoAdaptorWrapper) Bounds() basics.Rect[int] {
	return w.bounds
}

// Rewind prepares the adaptor for scanline iteration.
func (w *MonoAdaptorWrapper) Rewind(pathID uint) {
	if w.adaptor != nil {
		w.adaptor.RewindScanlines()
	}
}

// SweepScanline returns the next scanline.
func (w *MonoAdaptorWrapper) SweepScanline() bool {
	if w.adaptor != nil && w.embeddedScanline != nil {
		return w.adaptor.SweepEmbeddedScanline(w.embeddedScanline)
	}
	return false
}

// NumSpans returns the number of spans in current scanline.
func (w *MonoAdaptorWrapper) NumSpans() uint {
	if w.embeddedScanline != nil {
		return uint(w.embeddedScanline.NumSpans())
	}
	return 0
}

// Begin returns iterator for the first span.
func (w *MonoAdaptorWrapper) Begin() fonts.MonoSpanIterator {
	if w.embeddedScanline != nil {
		iter := w.embeddedScanline.Begin()
		return &MonoSpanIteratorWrapper{iter: iter}
	}
	return &MonoSpanIteratorWrapper{}
}

// MonoSpanIteratorWrapper adapts the scanline span iterator to the Fman interface.
type MonoSpanIteratorWrapper struct {
	iter *scanline.EmbeddedScanlineSerialIterator
}

// Next advances to the next span.
func (w *MonoSpanIteratorWrapper) Next() {
	if w.iter != nil {
		w.iter.Next()
	}
}

// IsValid returns true if the iterator is at a valid span.
func (w *MonoSpanIteratorWrapper) IsValid() bool {
	if w.iter != nil {
		return w.iter.IsValid()
	}
	return false
}

// X returns the starting X coordinate of current span.
func (w *MonoSpanIteratorWrapper) X() int {
	if w.iter != nil {
		return w.iter.X()
	}
	return 0
}

// Len returns the length of current span.
func (w *MonoSpanIteratorWrapper) Len() int {
	if w.iter != nil {
		return w.iter.Len()
	}
	return 0
}
