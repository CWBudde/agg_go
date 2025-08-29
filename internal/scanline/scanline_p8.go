// Package scanline provides scanline containers for the AGG rendering pipeline.
// This file implements the packed scanline container that stores horizontal
// spans with coverage values in a more memory-efficient packed format.
package scanline

import (
	"unsafe"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// SpanP8 represents a packed horizontal span of pixels.
// This corresponds to the span struct in AGG's scanline_p8 class.
// If Len is negative, it's a solid span where all pixels have the same coverage value.
type SpanP8 struct {
	X      basics.Int32  // Starting X coordinate (32-bit)
	Len    basics.Int32  // Length of span (negative = solid span with single cover value)
	Covers *basics.Int8u // Pointer to coverage values in the coverage array
}

// ScanlineP8 is a packed scanline container class.
// This class is used to transfer data from a scanline rasterizer
// to the rendering buffer. Unlike ScanlineU8, it uses a more compact
// representation where solid spans (all pixels with same coverage)
// are stored with negative length values.
//
// This is equivalent to AGG's scanline_p8 class.
type ScanlineP8 struct {
	lastX     int                        // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y         int                        // Y coordinate of current scanline
	covers    *array.PodArray[CoverType] // Coverage values array
	coverPtr  *CoverType                 // Current position in covers array
	spans     *array.PodArray[SpanP8]    // Spans array
	curSpan   *SpanP8                    // Pointer to current span being built
	spanIndex int                        // Index of current span
}

// NewScanlineP8 creates a new packed scanline container.
func NewScanlineP8() *ScanlineP8 {
	sl := &ScanlineP8{
		lastX:  0x7FFFFFF0, // Sentinel value indicating no previous X
		covers: array.NewPodArray[CoverType](),
		spans:  array.NewPodArray[SpanP8](),
	}
	return sl
}

// Reset prepares the scanline for a new row between min_x and max_x coordinates.
// This method must be called before adding any cells or spans to a new scanline.
func (sl *ScanlineP8) Reset(minX, maxX int) {
	maxLen := maxX - minX + 3 // Extra space for safety

	// Resize arrays if needed to accommodate the scanline width
	if maxLen > sl.spans.Size() {
		sl.spans.Resize(maxLen)
		sl.covers.Resize(maxLen)
	}

	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value

	// Set up coverage pointer at the beginning of the array
	if sl.covers.Size() > 0 {
		coverData := sl.covers.Data()
		sl.coverPtr = &coverData[0]
	}

	// Set up the first span (index 0 is used as a sentinel with len=0)
	if sl.spans.Size() > 0 {
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[0]
		sl.curSpan.Len = 0
		sl.spanIndex = 0
	}
}

// AddCell adds a single cell with coverage value to the scanline.
// X coordinates must be provided in increasing order.
func (sl *ScanlineP8) AddCell(x int, cover uint) {
	// Store the coverage value
	*sl.coverPtr = CoverType(cover)

	if x == sl.lastX+1 && sl.curSpan.Len > 0 {
		// Extend current span (non-solid span)
		sl.curSpan.Len++
	} else {
		// Start new span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curSpan.Covers = sl.coverPtr
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = 1
	}

	sl.lastX = x
	// Move coverage pointer forward
	coverData := sl.covers.Data()
	coverPtrIndex := int(uintptr(unsafe.Pointer(sl.coverPtr))-uintptr(unsafe.Pointer(&coverData[0]))) / int(unsafe.Sizeof(CoverType(0)))
	sl.coverPtr = &coverData[coverPtrIndex+1]
}

// AddCells adds multiple cells with individual coverage values to the scanline.
// X coordinates must be provided in increasing order.
func (sl *ScanlineP8) AddCells(x int, length int, covers []CoverType) {
	// Copy coverage values to our internal array
	coverData := sl.covers.Data()
	coverPtrIndex := int(uintptr(unsafe.Pointer(sl.coverPtr))-uintptr(unsafe.Pointer(&coverData[0]))) / int(unsafe.Sizeof(CoverType(0)))

	// Copy the coverage values
	for i := 0; i < length; i++ {
		coverData[coverPtrIndex+i] = covers[i]
	}

	if x == sl.lastX+1 && sl.curSpan.Len > 0 {
		// Extend current span (non-solid span)
		sl.curSpan.Len += basics.Int32(length)
	} else {
		// Start new span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curSpan.Covers = sl.coverPtr
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = basics.Int32(length)
	}

	// Move coverage pointer forward by length
	sl.coverPtr = &coverData[coverPtrIndex+length]
	sl.lastX = x + length - 1
}

// AddSpan adds a span of pixels all with the same coverage value.
// This creates a "solid" span with negative length for efficiency.
// X coordinates must be provided in increasing order.
func (sl *ScanlineP8) AddSpan(x int, length int, cover uint) {
	coverData := sl.covers.Data()

	// Check if we can merge with the previous solid span
	if x == sl.lastX+1 &&
		sl.curSpan.Len < 0 &&
		sl.curSpan.Covers != nil &&
		*sl.curSpan.Covers == CoverType(cover) {
		// Extend the existing solid span
		sl.curSpan.Len -= basics.Int32(length)
	} else {
		// Store the single coverage value
		*sl.coverPtr = CoverType(cover)

		// Start new solid span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curSpan.Covers = sl.coverPtr
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = -basics.Int32(length) // Negative indicates solid span

		// Move coverage pointer forward by 1 (only one value stored for solid spans)
		coverPtrIndex := int(uintptr(unsafe.Pointer(sl.coverPtr))-uintptr(unsafe.Pointer(&coverData[0]))) / int(unsafe.Sizeof(CoverType(0)))
		sl.coverPtr = &coverData[coverPtrIndex+1]
	}

	sl.lastX = x + length - 1
}

// Finalize finalizes the scanline and sets its Y coordinate.
// This should be called after all cells/spans have been added.
func (sl *ScanlineP8) Finalize(y int) {
	sl.y = y
}

// ResetSpans prepares the scanline for accumulating a new set of spans.
// This should be called after rendering the current scanline.
func (sl *ScanlineP8) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value

	// Reset coverage pointer to beginning
	if sl.covers.Size() > 0 {
		coverData := sl.covers.Data()
		sl.coverPtr = &coverData[0]
	}

	// Reset span pointer to beginning
	if sl.spans.Size() > 0 {
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[0]
		sl.curSpan.Len = 0
		sl.spanIndex = 0
	}
}

// Y returns the Y coordinate of the current scanline.
func (sl *ScanlineP8) Y() int {
	return sl.y
}

// NumSpans returns the number of spans in the current scanline.
// This is guaranteed to be greater than 0 if any cells/spans were added.
func (sl *ScanlineP8) NumSpans() int {
	return sl.spanIndex
}

// Begin returns an iterator (slice) to the spans.
// The returned slice starts from index 1, as index 0 is unused (following AGG convention).
func (sl *ScanlineP8) Begin() []SpanP8 {
	if sl.spanIndex == 0 {
		return nil
	}
	spanData := sl.spans.Data()
	return spanData[1 : sl.spanIndex+1]
}

// Spans returns all valid spans as a slice for iteration.
// This is a Go-idiomatic way to iterate over spans.
func (sl *ScanlineP8) Spans() []SpanP8 {
	return sl.Begin()
}

// IsSolid returns true if the span is solid (all pixels have same coverage).
// This is indicated by a negative length value.
func (span *SpanP8) IsSolid() bool {
	return span.Len < 0
}

// ActualLen returns the actual length of the span (absolute value of Len).
func (span *SpanP8) ActualLen() int {
	if span.Len < 0 {
		return int(-span.Len)
	}
	return int(span.Len)
}

// GetCovers returns the coverage values for this span.
// For solid spans, this returns a slice with a single repeated value.
// For non-solid spans, this returns the actual coverage array slice.
func (span *SpanP8) GetCovers() []CoverType {
	if span.Covers == nil {
		return nil
	}

	length := span.ActualLen()
	if length == 0 {
		return nil
	}

	// Create a slice from the pointer
	// Note: In a real implementation, we'd need access to the full coverage array
	// For now, we'll use unsafe pointer arithmetic
	return (*[1 << 20]CoverType)(unsafe.Pointer(span.Covers))[:length:length]
}
