package scanline

import "agg_go/internal/basics"

// ScanlineHitTest is a simple scanline implementation for hit testing.
// It checks if a specific X coordinate is covered by any rendered spans.
type ScanlineHitTest struct {
	x   int  // X coordinate to test for hit
	hit bool // Whether the test coordinate was hit
}

// NewScanlineHitTest creates a new hit test scanline for the given X coordinate
func NewScanlineHitTest(x int) *ScanlineHitTest {
	return &ScanlineHitTest{
		x:   x,
		hit: false,
	}
}

// ResetSpans resets the scanline for a new test, clearing the hit flag
func (s *ScanlineHitTest) ResetSpans() {
	s.hit = false
}

// AddCell adds a single cell to the scanline.
// If the cell's X coordinate matches our test coordinate, mark as hit.
func (s *ScanlineHitTest) AddCell(x int, cover basics.Int8u) {
	if x == s.x && cover > 0 {
		s.hit = true
	}
}

// AddSpan adds a span of consecutive pixels to the scanline.
// If our test X coordinate falls within this span, mark as hit.
func (s *ScanlineHitTest) AddSpan(x, len int, cover basics.Int8u) {
	if cover > 0 && s.x >= x && s.x < x+len {
		s.hit = true
	}
}

// AddCells adds multiple cells with individual coverage values.
// If our test X coordinate matches any cell position with coverage, mark as hit.
func (s *ScanlineHitTest) AddCells(x, len int, covers []basics.Int8u) {
	for i := 0; i < len; i++ {
		if covers[i] > 0 && x+i == s.x {
			s.hit = true
			return
		}
	}
}

// Finalize completes the scanline processing (no-op for hit test)
func (s *ScanlineHitTest) Finalize(y int) {
	// No operation needed for hit testing
}

// Hit returns whether the test coordinate was hit by any rendered span
func (s *ScanlineHitTest) Hit() bool {
	return s.hit
}

// NumSpans returns the number of spans (always 0 or 1 for hit test)
func (s *ScanlineHitTest) NumSpans() int {
	if s.hit {
		return 1
	}
	return 0
}

// Y returns the Y coordinate of this scanline (not used for hit testing)
func (s *ScanlineHitTest) Y() int {
	return 0
}

// Begin returns an iterator for this scanline (empty for hit test)
func (s *ScanlineHitTest) Begin() ScanlineIterator {
	return &HitTestIterator{}
}

// HitTestIterator is a dummy iterator for hit test scanlines
type HitTestIterator struct{}

// GetSpan returns an empty span info
func (h *HitTestIterator) GetSpan() SpanInfo {
	return SpanInfo{}
}

// Next always returns false as hit test has no spans to iterate
func (h *HitTestIterator) Next() bool {
	return false
}

// ScanlineInterface implementation check
var _ ScanlineInterface = (*ScanlineHitTest)(nil)
