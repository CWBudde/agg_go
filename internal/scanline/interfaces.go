package scanline

// Scanline is the unified scanline contract used by both the rasterizer (writer
// side) and the renderer (reader side). In C++ AGG, the scanline type is passed
// as a template parameter and implicitly satisfies both roles. In Go, this
// single interface captures the full contract.
//
// Writer methods (called by the rasterizer during SweepScanline):
//
//	ResetSpans, AddCell, AddSpan, Finalize
//
// Reader methods (called by the renderer after SweepScanline):
//
//	Y, NumSpans, BeginIterator
//
// Setup method (called before rendering begins):
//
//	Reset
type Scanline interface {
	// --- Writer side (rasterizer fills) ---

	// ResetSpans clears all accumulated span data for reuse.
	ResetSpans()

	// AddCell adds a single cell at position x with the given coverage.
	AddCell(x int, cover uint)

	// AddSpan adds a horizontal run of length pixels at position x with
	// uniform coverage.
	AddSpan(x, length int, cover uint)

	// Finalize marks the scanline complete at the given y coordinate.
	Finalize(y int)

	// --- Reader side (renderer reads) ---

	// Y returns the current scanline's Y coordinate.
	Y() int

	// NumSpans returns the number of spans in this scanline.
	NumSpans() int

	// BeginIterator returns an iterator to the first span.
	BeginIterator() ScanlineIterator

	// --- Setup ---

	// Reset prepares the scanline for a new rendering pass within the
	// specified X range.
	Reset(minX, maxX int)
}

// sliceIterP8 adapts a []SpanP8 slice to ScanlineIterator.
type sliceIterP8 struct {
	spans []SpanP8
	idx   int
}

func (it *sliceIterP8) GetSpan() SpanInfo {
	s := it.spans[it.idx]
	return SpanInfo{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}

func (it *sliceIterP8) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// sliceIterU8 adapts a []Span slice (ScanlineU8) to ScanlineIterator.
type sliceIterU8 struct {
	spans []Span
	idx   int
}

func (it *sliceIterU8) GetSpan() SpanInfo {
	s := it.spans[it.idx]
	return SpanInfo{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}

func (it *sliceIterU8) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// sliceIterBin adapts a []SpanBin slice to ScanlineIterator.
type sliceIterBin struct {
	spans []SpanBin
	idx   int
}

func (it *sliceIterBin) GetSpan() SpanInfo {
	s := it.spans[it.idx]
	return SpanInfo{X: int(s.X), Len: int(s.Len), Covers: nil}
}

func (it *sliceIterBin) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// sliceIter32P8 adapts a []Span32P8 slice to ScanlineIterator.
type sliceIter32P8 struct {
	spans []Span32P8
	idx   int
}

func (it *sliceIter32P8) GetSpan() SpanInfo {
	s := it.spans[it.idx]
	return SpanInfo{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}

func (it *sliceIter32P8) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}

// sliceIter32U8 adapts a []Span32U8 slice to ScanlineIterator.
type sliceIter32U8 struct {
	spans []Span32U8
	idx   int
}

func (it *sliceIter32U8) GetSpan() SpanInfo {
	s := it.spans[it.idx]
	return SpanInfo{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}

func (it *sliceIter32U8) Next() bool {
	it.idx++
	return it.idx < len(it.spans)
}
