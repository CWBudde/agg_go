package scanline

import (
	"github.com/MeKo-Christian/agg_go/internal/array"
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// SpanP8 is the Go equivalent of AGG's scanline_p8::span. A negative Len marks
// a solid span that stores only one cover value.
type SpanP8 struct {
	X      basics.Int32   // Starting X coordinate (32-bit)
	Len    basics.Int32   // Length of span (negative = solid span with single cover value)
	Covers []basics.Int8u // Coverage values in the coverage array
}

// ScanlineP8 is the Go equivalent of AGG's scanline_p8. It uses the packed
// solid-span encoding to reduce cover storage compared with ScanlineU8.
type ScanlineP8 struct {
	lastX     int                        // Last X coordinate processed (sentinel: 0x7FFFFFF0)
	y         int                        // Y coordinate of current scanline
	covers    *array.PodArray[CoverType] // Coverage values array
	coverIdx  int                        // Next write position in covers array
	spans     *array.PodArray[SpanP8]    // Spans array
	curSpan   *SpanP8                    // Pointer to current span being built
	curStart  int                        // Start index of the current span's covers
	spanIndex int                        // Index of current span
}

// NewScanlineP8 creates a packed AA scanline container.
func NewScanlineP8() *ScanlineP8 {
	sl := &ScanlineP8{
		lastX:  0x7FFFFFF0, // Sentinel value indicating no previous X
		covers: array.NewPodArray[CoverType](),
		spans:  array.NewPodArray[SpanP8](),
	}
	return sl
}

// Reset prepares the scanline for a new row.
func (sl *ScanlineP8) Reset(minX, maxX int) {
	maxLen := maxX - minX + 3 // Extra space for safety

	// Resize arrays if needed to accommodate the scanline width
	if maxLen > sl.spans.Size() {
		sl.spans.Resize(maxLen)
		sl.covers.Resize(maxLen)
	}

	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value

	sl.coverIdx = 0
	sl.curStart = 0

	// Set up the first span (index 0 is used as a sentinel with len=0)
	if sl.spans.Size() > 0 {
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[0]
		sl.curSpan.Len = 0
		sl.spanIndex = 0
	}
}

// AddCell adds one covered pixel. x must not go backwards within the row.
func (sl *ScanlineP8) AddCell(x int, cover uint) {
	coverData := sl.covers.Data()

	// Store the coverage value
	coverData[sl.coverIdx] = CoverType(cover)

	if x == sl.lastX+1 && sl.curSpan.Len > 0 {
		// Extend current span (non-solid span)
		sl.curSpan.Len++
		sl.curSpan.Covers = coverData[sl.curStart : sl.coverIdx+1]
	} else {
		// Start new span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curStart = sl.coverIdx
		sl.curSpan.Covers = coverData[sl.curStart : sl.curStart+1]
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = 1
	}

	sl.lastX = x
	sl.coverIdx++
}

// AddCells adds a run of per-pixel covers.
func (sl *ScanlineP8) AddCells(x, length int, covers []CoverType) {
	// Copy coverage values to our internal array
	coverData := sl.covers.Data()

	// Copy the coverage values
	for i := 0; i < length; i++ {
		coverData[sl.coverIdx+i] = covers[i]
	}

	if x == sl.lastX+1 && sl.curSpan.Len > 0 {
		// Extend current span (non-solid span)
		sl.curSpan.Len += basics.Int32(length)
		sl.curSpan.Covers = coverData[sl.curStart : sl.coverIdx+length]
	} else {
		// Start new span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curStart = sl.coverIdx
		sl.curSpan.Covers = coverData[sl.curStart : sl.curStart+length]
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = basics.Int32(length)
	}

	sl.coverIdx += length
	sl.lastX = x + length - 1
}

// AddSpan adds a solid-coverage run, reusing the packed negative-length encoding
// from AGG's scanline_p8.
func (sl *ScanlineP8) AddSpan(x, length int, cover uint) {
	coverData := sl.covers.Data()

	// Check if we can merge with the previous solid span
	if x == sl.lastX+1 &&
		sl.curSpan.Len < 0 &&
		len(sl.curSpan.Covers) > 0 &&
		sl.curSpan.Covers[0] == CoverType(cover) {
		// Extend the existing solid span
		sl.curSpan.Len -= basics.Int32(length)
	} else {
		// Store the single coverage value
		coverData[sl.coverIdx] = CoverType(cover)

		// Start new solid span
		sl.spanIndex++
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[sl.spanIndex]
		sl.curStart = sl.coverIdx
		sl.curSpan.Covers = coverData[sl.curStart : sl.curStart+1]
		sl.curSpan.X = basics.Int32(x)
		sl.curSpan.Len = -basics.Int32(length) // Negative indicates solid span

		// Move coverage position forward by 1 (only one value stored for solid spans)
		sl.coverIdx++
	}

	sl.lastX = x + length - 1
}

// Finalize records the row y after accumulation.
func (sl *ScanlineP8) Finalize(y int) {
	sl.y = y
}

// ResetSpans clears the accumulated row while reusing buffers.
func (sl *ScanlineP8) ResetSpans() {
	sl.lastX = 0x7FFFFFF0 // Reset to sentinel value

	sl.coverIdx = 0
	sl.curStart = 0

	// Reset span pointer to beginning
	if sl.spans.Size() > 0 {
		spanData := sl.spans.Data()
		sl.curSpan = &spanData[0]
		sl.curSpan.Len = 0
		sl.spanIndex = 0
	}
}

// Y returns the row coordinate.
func (sl *ScanlineP8) Y() int {
	return sl.y
}

// NumSpans returns the number of accumulated spans.
func (sl *ScanlineP8) NumSpans() int {
	return sl.spanIndex
}

// Begin returns the span slice starting at index 1, preserving AGG's sentinel
// convention that slot 0 is unused.
func (sl *ScanlineP8) Begin() []SpanP8 {
	if sl.spanIndex == 0 {
		return nil
	}
	spanData := sl.spans.Data()
	return spanData[1 : sl.spanIndex+1]
}

// Spans is a Go-friendly alias for Begin.
func (sl *ScanlineP8) Spans() []SpanP8 {
	return sl.Begin()
}

// IsSolid reports whether the span uses the packed solid-span encoding.
func (span *SpanP8) IsSolid() bool {
	return span.Len < 0
}

// ActualLen returns the absolute pixel count represented by Len.
func (span *SpanP8) ActualLen() int {
	if span.Len < 0 {
		return int(-span.Len)
	}
	return int(span.Len)
}

// GetCovers returns the stored cover slice. Solid spans expose the one-value
// backing slice used by the packed representation.
func (span *SpanP8) GetCovers() []CoverType {
	if len(span.Covers) == 0 {
		return nil
	}

	length := span.ActualLen()
	if length == 0 {
		return nil
	}

	if length > len(span.Covers) {
		length = len(span.Covers)
	}

	return span.Covers[:length]
}
