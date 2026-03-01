//go:build freetype

package freetype2

import (
	"agg_go/internal/basics"
	"agg_go/internal/fonts"
	"agg_go/internal/scanline"
)

// gray8AdaptorWrapper is a thin Go adapter from the scanline package's
// serialized AA adaptor to the separate internal/fonts fman interface.
// It remains necessary because scanline exposes concrete serialized-scanline
// iteration APIs while internal/fonts expects a generic adaptor/span interface.
type gray8AdaptorWrapper struct {
	adaptor          *scanline.SerializedScanlinesAdaptorAA[uint8]
	embeddedScanline *scanline.SerializedEmbeddedScanline[uint8]
	bounds           basics.Rect[int]
}

// newGray8AdaptorWrapper creates a new gray8AdaptorWrapper wrapping the given scanline adaptor.
func newGray8AdaptorWrapper(adaptor *scanline.SerializedScanlinesAdaptorAA[uint8]) *gray8AdaptorWrapper {
	return &gray8AdaptorWrapper{
		adaptor:          adaptor,
		embeddedScanline: scanline.NewSerializedEmbeddedScanline[uint8](),
	}
}

// InitGlyph initializes the adaptor with glyph data and position.
func (w *gray8AdaptorWrapper) InitGlyph(data []byte, dataSize uint32, x, y float64) {
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
func (w *gray8AdaptorWrapper) Bounds() basics.Rect[int] {
	return w.bounds
}

// Rewind prepares the adaptor for scanline iteration.
func (w *gray8AdaptorWrapper) Rewind(pathID uint) {
	if w.adaptor != nil {
		w.adaptor.RewindScanlines()
	}
}

// SweepScanline returns the next scanline.
func (w *gray8AdaptorWrapper) SweepScanline() bool {
	if w.adaptor != nil && w.embeddedScanline != nil {
		return w.adaptor.SweepSerializedEmbeddedScanline(w.embeddedScanline)
	}
	return false
}

// NumSpans returns the number of spans in current scanline.
func (w *gray8AdaptorWrapper) NumSpans() uint {
	if w.embeddedScanline != nil {
		return uint(w.embeddedScanline.NumSpans())
	}
	return 0
}

// Begin returns iterator for the first span.
func (w *gray8AdaptorWrapper) Begin() fonts.Gray8SpanIterator {
	if w.embeddedScanline != nil {
		iter := w.embeddedScanline.Begin()
		return &gray8SpanIteratorWrapper{iter: iter}
	}
	return &gray8SpanIteratorWrapper{}
}

// gray8SpanIteratorWrapper adapts the scanline span iterator to the Fman interface.
type gray8SpanIteratorWrapper struct {
	iter *scanline.SerializedEmbeddedScanlineIterator[uint8]
}

// Next advances to the next span.
func (w *gray8SpanIteratorWrapper) Next() {
	if w.iter != nil {
		w.iter.Next()
	}
}

// IsValid returns true if the iterator is at a valid span.
func (w *gray8SpanIteratorWrapper) IsValid() bool {
	if w.iter != nil {
		return w.iter.IsValid()
	}
	return false
}

// X returns the starting X coordinate of current span.
func (w *gray8SpanIteratorWrapper) X() int {
	if w.iter != nil {
		return w.iter.X()
	}
	return 0
}

// Len returns the length of current span.
func (w *gray8SpanIteratorWrapper) Len() int {
	if w.iter != nil {
		return w.iter.Len()
	}
	return 0
}

// Covers returns the coverage array for current span.
func (w *gray8SpanIteratorWrapper) Covers() []uint8 {
	if w.iter != nil {
		return w.iter.Covers()
	}
	return nil
}

// monoAdaptorWrapper is a thin Go adapter from the scanline package's
// serialized binary adaptor to the separate internal/fonts fman interface.
// It remains necessary because scanline exposes concrete serialized-scanline
// iteration APIs while internal/fonts expects a generic adaptor/span interface.
type monoAdaptorWrapper struct {
	adaptor          *scanline.SerializedScanlinesAdaptorBin
	embeddedScanline *scanline.EmbeddedScanlineSerial
	bounds           basics.Rect[int]
}

// newMonoAdaptorWrapper creates a new monoAdaptorWrapper wrapping the given scanline adaptor.
func newMonoAdaptorWrapper(adaptor *scanline.SerializedScanlinesAdaptorBin) *monoAdaptorWrapper {
	return &monoAdaptorWrapper{
		adaptor:          adaptor,
		embeddedScanline: scanline.NewEmbeddedScanlineSerial(),
	}
}

// InitGlyph initializes the adaptor with glyph data and position.
func (w *monoAdaptorWrapper) InitGlyph(data []byte, dataSize uint32, x, y float64) {
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
func (w *monoAdaptorWrapper) Bounds() basics.Rect[int] {
	return w.bounds
}

// Rewind prepares the adaptor for scanline iteration.
func (w *monoAdaptorWrapper) Rewind(pathID uint) {
	if w.adaptor != nil {
		w.adaptor.RewindScanlines()
	}
}

// SweepScanline returns the next scanline.
func (w *monoAdaptorWrapper) SweepScanline() bool {
	if w.adaptor != nil && w.embeddedScanline != nil {
		return w.adaptor.SweepEmbeddedScanline(w.embeddedScanline)
	}
	return false
}

// NumSpans returns the number of spans in current scanline.
func (w *monoAdaptorWrapper) NumSpans() uint {
	if w.embeddedScanline != nil {
		return uint(w.embeddedScanline.NumSpans())
	}
	return 0
}

// Begin returns iterator for the first span.
func (w *monoAdaptorWrapper) Begin() fonts.MonoSpanIterator {
	if w.embeddedScanline != nil {
		iter := w.embeddedScanline.Begin()
		return &monoSpanIteratorWrapper{iter: iter}
	}
	return &monoSpanIteratorWrapper{}
}

// monoSpanIteratorWrapper adapts the scanline span iterator to the Fman interface.
type monoSpanIteratorWrapper struct {
	iter *scanline.EmbeddedScanlineSerialIterator
}

// Next advances to the next span.
func (w *monoSpanIteratorWrapper) Next() {
	if w.iter != nil {
		w.iter.Next()
	}
}

// IsValid returns true if the iterator is at a valid span.
func (w *monoSpanIteratorWrapper) IsValid() bool {
	if w.iter != nil {
		return w.iter.IsValid()
	}
	return false
}

// X returns the starting X coordinate of current span.
func (w *monoSpanIteratorWrapper) X() int {
	if w.iter != nil {
		return w.iter.X()
	}
	return 0
}

// Len returns the length of current span.
func (w *monoSpanIteratorWrapper) Len() int {
	if w.iter != nil {
		return w.iter.Len()
	}
	return 0
}
