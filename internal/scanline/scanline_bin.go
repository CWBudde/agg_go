// Package scanline provides scanline containers for the AGG rendering pipeline.
package scanline

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// SpanBin represents a binary horizontal span without coverage values.
// This is used when only pixel presence matters, not anti-aliasing coverage.
// Corresponds to the span struct in AGG's scanline_bin class.
type SpanBin struct {
	X   basics.Int16 // Starting X coordinate
	Len basics.Int16 // Length of the span
}

// ScanlineBin is a binary scanline container which supports the interface
// used in the rasterizer. This is used when anti-aliasing is not needed.
// It only stores which pixels are covered, not their coverage values.
//
// This is equivalent to AGG's scanline_bin class.
// IMPORTANT: Following AGG convention, spans are stored starting at index 1 (index 0 is unused).
type ScanlineBin struct {
	lastX   int                      // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y       int                      // Y coordinate of current scanline
	spans   *array.PodArray[SpanBin] // Spans array (index 0 unused, spans start at index 1)
	curSpan int                      // Index of current span being built (starts at 1)
}

// NewScanlineBin creates a new binary scanline container.
func NewScanlineBin() *ScanlineBin {
	spans := array.NewPodArray[SpanBin]()
	// Ensure we have space for the dummy element at index 0
	if spans.Size() == 0 {
		spans.Resize(1)
	}
	return &ScanlineBin{
		lastX:   0x7FFFFFF0, // Sentinel value indicating no previous X
		curSpan: 0,          // Will start at index 1 when first span is added
		spans:   spans,
	}
}

// Reset prepares the scanline for a new row between min_x and max_x coordinates.
func (sl *ScanlineBin) Reset(minX, maxX int) {
	maxLen := maxX - minX + 3
	// Ensure we have space for at least the dummy element at index 0 plus the calculated spans
	if maxLen+1 > sl.spans.Size() {
		sl.spans.Resize(maxLen + 1)
	}
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.curSpan = 0        // Will start at index 1 when first span is added
}

// AddCell adds a single cell to the scanline. The coverage value is ignored.
func (sl *ScanlineBin) AddCell(x int, _ uint) {
	if x == sl.lastX+1 && sl.curSpan > 0 {
		// Extend current span
		currentSpan := sl.spans.ValueAt(sl.curSpan)
		currentSpan.Len++
		sl.spans.Set(sl.curSpan, currentSpan)
	} else {
		// Start new span - AGG convention: spans start at index 1
		sl.curSpan++
		newSpan := SpanBin{
			X:   basics.Int16(x),
			Len: 1,
		}
		// Ensure we have space for the new span
		if sl.curSpan >= sl.spans.Size() {
			sl.spans.Resize(sl.spans.Size() + 10) // Grow in chunks for efficiency
		}
		sl.spans.Set(sl.curSpan, newSpan)
	}
	sl.lastX = x
}

// AddSpan adds a span of pixels to the scanline. The coverage value is ignored.
func (sl *ScanlineBin) AddSpan(x int, length int, _ uint) {
	if x == sl.lastX+1 && sl.curSpan > 0 {
		// Extend current span
		currentSpan := sl.spans.ValueAt(sl.curSpan)
		currentSpan.Len += basics.Int16(length)
		sl.spans.Set(sl.curSpan, currentSpan)
	} else {
		// Start new span - AGG convention: spans start at index 1
		sl.curSpan++
		newSpan := SpanBin{
			X:   basics.Int16(x),
			Len: basics.Int16(length),
		}
		// Ensure we have space for the new span
		if sl.curSpan >= sl.spans.Size() {
			sl.spans.Resize(sl.spans.Size() + 10) // Grow in chunks for efficiency
		}
		sl.spans.Set(sl.curSpan, newSpan)
	}
	sl.lastX = x + length - 1
}

// AddCells adds multiple cells to the scanline. The covers pointer is ignored.
func (sl *ScanlineBin) AddCells(x int, length int, _ []CoverType) {
	sl.AddSpan(x, length, 0)
}

// Finalize finalizes the scanline and sets its Y coordinate.
func (sl *ScanlineBin) Finalize(y int) {
	sl.y = y
}

// ResetSpans prepares the scanline for accumulating a new set of spans.
func (sl *ScanlineBin) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.curSpan = 0        // Will start at index 1 when first span is added
}

// Y returns the Y coordinate of the current scanline.
func (sl *ScanlineBin) Y() int {
	return sl.y
}

// NumSpans returns the number of spans in the current scanline.
// Following AGG convention: spans start at index 1, so curSpan is the count.
func (sl *ScanlineBin) NumSpans() int {
	return sl.curSpan // curSpan is already the count since we start from 1
}

// Begin returns an iterator (slice) to the spans.
// The returned slice starts from index 1, as index 0 is unused (following AGG convention).
func (sl *ScanlineBin) Begin() []SpanBin {
	if sl.curSpan == 0 {
		return nil
	}
	// AGG stores spans starting at index 1, return slice from 1 to curSpan (inclusive)
	return sl.spans.Data()[1 : sl.curSpan+1]
}

// Spans returns all valid spans as a slice for iteration.
// This is a Go-idiomatic way to iterate over spans.
func (sl *ScanlineBin) Spans() []SpanBin {
	return sl.Begin()
}

// =============================================================scanline32_bin

// Span32Bin represents a binary horizontal span with 32-bit coordinates.
// This is used for larger coordinate spaces.
type Span32Bin struct {
	X   basics.Int32 // Starting X coordinate
	Len basics.Int32 // Length of the span
}

// NewSpan32Bin creates a new 32-bit span with the specified coordinates.
func NewSpan32Bin(x, length basics.Int32) Span32Bin {
	return Span32Bin{X: x, Len: length}
}

// Scanline32Bin is a binary scanline container with 32-bit coordinates.
// It uses a block vector for efficient memory management of spans.
//
// This is equivalent to AGG's scanline32_bin class.
type Scanline32Bin struct {
	lastX int                          // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y     int                          // Y coordinate of current scanline
	spans *array.PodBVector[Span32Bin] // Dynamic spans using block vector
}

// NewScanline32Bin creates a new 32-bit binary scanline container.
func NewScanline32Bin() *Scanline32Bin {
	// Use shift of 4 (block size = 16) for span storage
	scale := array.NewBlockScale(4)
	return &Scanline32Bin{
		lastX: 0x7FFFFFF0, // Sentinel value indicating no previous X
		spans: array.NewPodBVectorWithScale[Span32Bin](scale),
	}
}

// Reset prepares the scanline for a new row. Parameters are ignored for compatibility.
func (sl *Scanline32Bin) Reset(_, _ int) {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.spans.RemoveAll()
}

// AddCell adds a single cell to the scanline. The coverage value is ignored.
func (sl *Scanline32Bin) AddCell(x int, _ uint) {
	if x == sl.lastX+1 && sl.spans.Size() > 0 {
		// Extend last span
		lastIdx := sl.spans.Size() - 1
		lastSpan := sl.spans.At(lastIdx)
		lastSpan.Len++
		sl.spans.Set(lastIdx, lastSpan)
	} else {
		// Add new span
		sl.spans.Add(NewSpan32Bin(basics.Int32(x), 1))
	}
	sl.lastX = x
}

// AddSpan adds a span of pixels to the scanline. The coverage value is ignored.
func (sl *Scanline32Bin) AddSpan(x int, length int, _ uint) {
	if x == sl.lastX+1 && sl.spans.Size() > 0 {
		// Extend last span
		lastIdx := sl.spans.Size() - 1
		lastSpan := sl.spans.At(lastIdx)
		lastSpan.Len += basics.Int32(length)
		sl.spans.Set(lastIdx, lastSpan)
	} else {
		// Add new span
		sl.spans.Add(NewSpan32Bin(basics.Int32(x), basics.Int32(length)))
	}
	sl.lastX = x + length - 1
}

// AddCells adds multiple cells to the scanline. The covers pointer is ignored.
func (sl *Scanline32Bin) AddCells(x int, length int, _ []CoverType) {
	sl.AddSpan(x, length, 0)
}

// Finalize finalizes the scanline and sets its Y coordinate.
func (sl *Scanline32Bin) Finalize(y int) {
	sl.y = y
}

// ResetSpans prepares the scanline for accumulating a new set of spans.
func (sl *Scanline32Bin) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.spans.RemoveAll()
}

// Y returns the Y coordinate of the current scanline.
func (sl *Scanline32Bin) Y() int {
	return sl.y
}

// NumSpans returns the number of spans in the current scanline.
func (sl *Scanline32Bin) NumSpans() int {
	return sl.spans.Size()
}

// Begin returns an iterator for the spans.
// Since PodBVector stores spans sequentially, we can iterate directly.
func (sl *Scanline32Bin) Begin() *Scanline32BinIterator {
	return &Scanline32BinIterator{
		spans:   sl.spans,
		spanIdx: 0,
	}
}

// Spans returns all spans as a slice for iteration.
// This is a Go-idiomatic way to iterate over spans.
func (sl *Scanline32Bin) Spans() []Span32Bin {
	result := make([]Span32Bin, sl.spans.Size())
	for i := 0; i < sl.spans.Size(); i++ {
		result[i] = sl.spans.At(i)
	}
	return result
}

// Scanline32BinIterator provides iteration over spans in a Scanline32Bin.
// This corresponds to the const_iterator class in AGG's scanline32_bin.
type Scanline32BinIterator struct {
	spans   *array.PodBVector[Span32Bin]
	spanIdx int
}

// Next advances to the next span and returns true if there are more spans.
func (it *Scanline32BinIterator) Next() bool {
	it.spanIdx++
	return it.spanIdx < it.spans.Size()
}

// Span returns the current span.
func (it *Scanline32BinIterator) Span() Span32Bin {
	if it.spanIdx >= it.spans.Size() {
		panic("iterator out of bounds")
	}
	return it.spans.At(it.spanIdx)
}

// HasMore returns true if there are more spans to iterate.
func (it *Scanline32BinIterator) HasMore() bool {
	return it.spanIdx < it.spans.Size()
}
