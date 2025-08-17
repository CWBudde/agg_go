// Package scanline provides scanline containers for the AGG rendering pipeline.
// This file implements the 32-bit unpacked scanline container that stores horizontal
// spans with coverage values using 32-bit coordinates for larger coordinate ranges.
package scanline

import (
	"agg_go/internal/array"
)

// Span32U8 represents a horizontal span of pixels with 32-bit coordinates.
// This corresponds to the span struct in AGG's scanline32_u8 class.
// Unlike Span32P8, this is always unpacked - each pixel has its own coverage value.
type Span32U8 struct {
	X      Coord32Type // Starting X coordinate (32-bit)
	Len    Coord32Type // Length of the span
	Covers []CoverType // Slice pointing to coverage values in the coverage array
}

// Scanline32U8 is a 32-bit unpacked scanline container class.
// This class is used to transfer data from a scanline rasterizer
// to the rendering buffer. It stores information of horizontal spans
// to render into a pixel-map buffer. Each span has starting X, length,
// and an array of bytes that determine the cover-values for each pixel.
// Uses 32-bit coordinates for larger coordinate ranges.
//
// This is equivalent to AGG's scanline32_u8 class.
type Scanline32U8 struct {
	minX   int                         // Minimum X coordinate for current scanline
	lastX  int                         // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y      int                         // Y coordinate of current scanline
	covers *array.PodArray[CoverType]  // Coverage values array
	spans  *array.PodBVector[Span32U8] // Spans array using block vector
}

// NewScanline32U8 creates a new 32-bit unpacked scanline container.
func NewScanline32U8() *Scanline32U8 {
	return &Scanline32U8{
		minX:   0,
		lastX:  0x7FFFFFF0, // Sentinel value indicating no previous X
		covers: array.NewPodArray[CoverType](),
		spans:  array.NewPodBVector[Span32U8](),
	}
}

// Reset prepares the scanline for a new row between min_x and max_x coordinates.
// This method must be called before adding any cells or spans to a new scanline.
func (sl *Scanline32U8) Reset(minX, maxX int) {
	maxLen := maxX - minX + 2

	// Resize coverage array if needed to accommodate the scanline width
	if maxLen > sl.covers.Size() {
		sl.covers.Resize(maxLen)
	}

	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.minX = minX
	sl.spans.RemoveAll() // Clear spans using PodBVector method
}

// AddCell adds a single cell with coverage value to the scanline.
// X coordinates must be provided in increasing order.
func (sl *Scanline32U8) AddCell(x int, cover uint) {
	x -= sl.minX
	sl.covers.Set(x, CoverType(cover))

	if x == sl.lastX+1 {
		// Extend current span
		lastSpan := sl.spans.Last()
		if lastSpan != nil {
			lastSpan.Len++
		}
	} else {
		// Start new span
		newSpan := Span32U8{
			X:      Coord32Type(x + sl.minX),
			Len:    1,
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		sl.spans.Add(newSpan)
	}
	sl.lastX = x
}

// AddCells adds multiple cells with individual coverage values to the scanline.
// X coordinates must be provided in increasing order.
func (sl *Scanline32U8) AddCells(x int, length int, covers []CoverType) {
	x -= sl.minX

	// Copy coverage values to our internal array
	coverData := sl.covers.Data()
	copy(coverData[x:x+length], covers[:length])

	if x == sl.lastX+1 {
		// Extend current span
		lastSpan := sl.spans.Last()
		if lastSpan != nil {
			lastSpan.Len += Coord32Type(length)
		}
	} else {
		// Start new span
		newSpan := Span32U8{
			X:      Coord32Type(x + sl.minX),
			Len:    Coord32Type(length),
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		sl.spans.Add(newSpan)
	}
	sl.lastX = x + length - 1
}

// AddSpan adds a span of pixels all with the same coverage value.
// X coordinates must be provided in increasing order.
func (sl *Scanline32U8) AddSpan(x int, length int, cover uint) {
	x -= sl.minX

	// Fill coverage values with the same value
	coverData := sl.covers.Data()
	for i := 0; i < length; i++ {
		coverData[x+i] = CoverType(cover)
	}

	if x == sl.lastX+1 {
		// Extend current span
		lastSpan := sl.spans.Last()
		if lastSpan != nil {
			lastSpan.Len += Coord32Type(length)
		}
	} else {
		// Start new span
		newSpan := Span32U8{
			X:      Coord32Type(x + sl.minX),
			Len:    Coord32Type(length),
			Covers: sl.covers.Data()[x:], // Slice starting at position x
		}
		sl.spans.Add(newSpan)
	}
	sl.lastX = x + length - 1
}

// Finalize finalizes the scanline and sets its Y coordinate.
// This should be called after all cells/spans have been added.
func (sl *Scanline32U8) Finalize(y int) {
	sl.y = y
}

// ResetSpans prepares the scanline for accumulating a new set of spans.
// This should be called after rendering the current scanline.
func (sl *Scanline32U8) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.spans.RemoveAll()  // Clear spans using PodBVector method
}

// Y returns the Y coordinate of the current scanline.
func (sl *Scanline32U8) Y() int {
	return sl.y
}

// NumSpans returns the number of spans in the current scanline.
// This is guaranteed to be greater than 0 if any cells/spans were added.
func (sl *Scanline32U8) NumSpans() int {
	return sl.spans.Size()
}

// Begin returns an iterator (slice) to the spans.
// For compatibility with AGG's C++ interface, this returns all spans.
func (sl *Scanline32U8) Begin() []Span32U8 {
	if sl.spans.Size() == 0 {
		return nil
	}

	// Create a slice from the block vector
	spans := make([]Span32U8, sl.spans.Size())
	for i := 0; i < sl.spans.Size(); i++ {
		spans[i] = sl.spans.ValueAt(i)
	}
	return spans
}

// Spans returns all valid spans as a slice for iteration.
// This is a Go-idiomatic way to iterate over spans.
func (sl *Scanline32U8) Spans() []Span32U8 {
	return sl.Begin()
}
