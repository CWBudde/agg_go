package scanline

import (
	"github.com/MeKo-Christian/agg_go/internal/array"
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// CoverType is the per-pixel coverage type used by anti-aliased scanlines.
type CoverType = basics.Int8u

// Coord16Type is the 16-bit coordinate type used by ScanlineU8 spans.
type Coord16Type = basics.Int16

// Span is the Go equivalent of AGG's scanline_u8::span.
type Span struct {
	X      Coord16Type // Starting X coordinate
	Len    Coord16Type // Length of the span
	Covers []CoverType // Pointer to coverage values array
}

// ScanlineU8 is the Go equivalent of AGG's scanline_u8. It stores one row as
// explicit spans plus one cover byte per covered pixel, which makes iteration
// simple at the cost of more memory than the packed variants.
type ScanlineU8 struct {
	minX    int                        // Minimum X coordinate for current scanline
	lastX   int                        // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y       int                        // Y coordinate of current scanline
	covers  *array.PodArray[CoverType] // Coverage values array
	spans   *array.PodArray[Span]      // Spans array
	curSpan int                        // Index of current span being built
}

// NewScanlineU8 creates an unpacked AA scanline container.
func NewScanlineU8() *ScanlineU8 {
	return &ScanlineU8{
		minX:    0,
		lastX:   0x7FFFFFF0, // Sentinel value indicating no previous X
		curSpan: 0,
		covers:  array.NewPodArray[CoverType](),
		spans:   array.NewPodArray[Span](),
	}
}

// Reset prepares the scanline for a new row. As in AGG, callers must then add
// cells/spans with monotonically increasing x coordinates.
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

// AddCell adds one covered pixel. x must not go backwards within the row.
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

// AddCells adds a run of per-pixel covers. x must not go backwards within the row.
func (sl *ScanlineU8) AddCells(x, length int, covers []CoverType) {
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

// AddSpan adds a solid-coverage run. x must not go backwards within the row.
func (sl *ScanlineU8) AddSpan(x, length int, cover uint) {
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

// Finalize records the row y after all spans have been accumulated.
func (sl *ScanlineU8) Finalize(y int) {
	sl.y = y
}

// ResetSpans clears the accumulated row while reusing the allocated buffers.
func (sl *ScanlineU8) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value
	sl.curSpan = 0
}

// Y returns the row coordinate.
func (sl *ScanlineU8) Y() int {
	return sl.y
}

// NumSpans returns the number of accumulated spans.
func (sl *ScanlineU8) NumSpans() int {
	return sl.curSpan
}

// Begin returns the span slice starting at index 1, preserving AGG's sentinel
// convention that slot 0 is unused.
func (sl *ScanlineU8) Begin() []Span {
	if sl.curSpan == 0 {
		return nil
	}
	return sl.spans.Data()[1 : sl.curSpan+1]
}

// Spans is a Go-friendly alias for Begin.
func (sl *ScanlineU8) Spans() []Span {
	return sl.Begin()
}

// BeginIterator returns an iterator over the spans, satisfying the unified
// Scanline interface.
func (sl *ScanlineU8) BeginIterator() ScanlineIterator {
	spans := sl.Begin()
	if len(spans) == 0 {
		return &sliceIterU8{}
	}
	return &sliceIterU8{spans: spans}
}
