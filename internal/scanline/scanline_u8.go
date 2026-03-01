// Package scanline provides scanline containers for the AGG rendering pipeline.
// This package implements unpacked scanline containers that store horizontal
// spans with coverage values for anti-aliased rendering.
package scanline

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// CoverType represents coverage values for anti-aliasing (0-255)
type CoverType = basics.Int8u

// Coord16Type represents 16-bit coordinate values for scanline_u8
type Coord16Type = basics.Int16

// Span represents a horizontal span of pixels with coverage values.
// This corresponds to the span struct in AGG's scanline_u8 class.
type Span struct {
	X      Coord16Type // Starting X coordinate
	Len    Coord16Type // Length of the span
	Covers []CoverType // Pointer to coverage values array
}

// ScanlineU8 is an unpacked scanline container class.
// This class is used to transfer data from a scanline rasterizer
// to the rendering buffer. It stores information of horizontal spans
// to render into a pixel-map buffer. Each span has starting X, length,
// and an array of bytes that determine the cover-values for each pixel.
//
// This is equivalent to AGG's scanline_u8 class.
type ScanlineU8 struct {
	minX    int                        // Minimum X coordinate for current scanline
	lastX   int                        // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y       int                        // Y coordinate of current scanline
	covers  *array.PodArray[CoverType] // Coverage values array
	spans   *array.PodArray[Span]      // Spans array
	curSpan int                        // Index of current span being built
}

// NewScanlineU8 creates a new scanline container.
func NewScanlineU8() *ScanlineU8 {
	return &ScanlineU8{
		minX:    0,
		lastX:   0x7FFFFFF0, // Sentinel value indicating no previous X
		curSpan: 0,
		covers:  array.NewPodArray[CoverType](),
		spans:   array.NewPodArray[Span](),
	}
}

// Reset prepares the scanline for a new row between min_x and max_x coordinates.
// This method must be called before adding any cells or spans to a new scanline.
func (sl *ScanlineU8) Reset(minX, maxX int) {
	maxLen := maxX - minX + 2

	// Resize arrays if needed to accommodate the scanline width
	if maxLen > sl.covers.Size() {
		sl.covers.Resize(maxLen)
		sl.spans.Resize(maxLen)
	}

	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.minX = minX
	sl.curSpan = 0
}

// AddCell adds a single cell with coverage value to the scanline.
// X coordinates must be provided in increasing order.
func (sl *ScanlineU8) AddCell(x int, cover uint) {
	x -= sl.minX
	if x < 0 || x >= sl.covers.Size() {
		return
	}
	sl.covers.Set(x, CoverType(cover))

	if x == sl.lastX+1 {
		// Extend current span
		currentSpan := sl.spans.ValueAt(sl.curSpan)
		currentSpan.Len++
		sl.spans.Set(sl.curSpan, currentSpan)
	} else {
		// Start new span
		sl.curSpan++
		newSpan := Span{
			X:      Coord16Type(x + sl.minX),
			Len:    1,
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		if sl.curSpan >= sl.spans.Size() {
			sl.spans.Resize(sl.spans.Size() + 1)
		}
		sl.spans.Set(sl.curSpan, newSpan)
	}
	sl.lastX = x
}

// AddCells adds multiple cells with individual coverage values to the scanline.
// X coordinates must be provided in increasing order.
func (sl *ScanlineU8) AddCells(x int, length int, covers []CoverType) {
	x -= sl.minX
	if x < 0 {
		diff := -x
		if diff >= length {
			return
		}
		x = 0
		length -= diff
		covers = covers[diff:]
	}

	if x+length > sl.covers.Size() {
		length = sl.covers.Size() - x
	}

	if length <= 0 {
		return
	}

	// Copy coverage values to our internal array
	coverData := sl.covers.Data()
	copy(coverData[x:x+length], covers[:length])

	if x == sl.lastX+1 {
		// Extend current span
		currentSpan := sl.spans.ValueAt(sl.curSpan)
		currentSpan.Len += Coord16Type(length)
		sl.spans.Set(sl.curSpan, currentSpan)
	} else {
		// Start new span
		sl.curSpan++
		newSpan := Span{
			X:      Coord16Type(x + sl.minX),
			Len:    Coord16Type(length),
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		if sl.curSpan >= sl.spans.Size() {
			sl.spans.Resize(sl.spans.Size() + 1)
		}
		sl.spans.Set(sl.curSpan, newSpan)
	}
	sl.lastX = x + length - 1
}

// AddSpan adds a span of pixels all with the same coverage value.
// X coordinates must be provided in increasing order.
func (sl *ScanlineU8) AddSpan(x int, length int, cover uint) {
	x -= sl.minX
	if x < 0 {
		diff := -x
		if diff >= length {
			return
		}
		x = 0
		length -= diff
	}

	if x+length > sl.covers.Size() {
		length = sl.covers.Size() - x
	}

	if length <= 0 {
		return
	}

	// Fill coverage values with the same value
	coverData := sl.covers.Data()
	for i := 0; i < length; i++ {
		coverData[x+i] = CoverType(cover)
	}

	if x == sl.lastX+1 {
		// Extend current span
		currentSpan := sl.spans.ValueAt(sl.curSpan)
		currentSpan.Len += Coord16Type(length)
		sl.spans.Set(sl.curSpan, currentSpan)
	} else {
		// Start new span
		sl.curSpan++
		newSpan := Span{
			X:      Coord16Type(x + sl.minX),
			Len:    Coord16Type(length),
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		if sl.curSpan >= sl.spans.Size() {
			sl.spans.Resize(sl.spans.Size() + 1)
		}
		sl.spans.Set(sl.curSpan, newSpan)
	}
	sl.lastX = x + length - 1
}

// Finalize finalizes the scanline and sets its Y coordinate.
// This should be called after all cells/spans have been added.
func (sl *ScanlineU8) Finalize(y int) {
	sl.y = y
}

// ResetSpans prepares the scanline for accumulating a new set of spans.
// This should be called after rendering the current scanline.
func (sl *ScanlineU8) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.curSpan = 0
}

// Y returns the Y coordinate of the current scanline.
func (sl *ScanlineU8) Y() int {
	return sl.y
}

// NumSpans returns the number of spans in the current scanline.
// This is guaranteed to be greater than 0 if any cells/spans were added.
func (sl *ScanlineU8) NumSpans() int {
	return sl.curSpan
}

// Begin returns an iterator (slice) to the spans.
// The returned slice starts from index 1, as index 0 is unused (following AGG convention).
func (sl *ScanlineU8) Begin() []Span {
	if sl.curSpan == 0 {
		return nil
	}
	return sl.spans.Data()[1 : sl.curSpan+1]
}

// Spans returns all valid spans as a slice for iteration.
// This is a Go-idiomatic way to iterate over spans.
func (sl *ScanlineU8) Spans() []Span {
	return sl.Begin()
}
