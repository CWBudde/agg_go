// Package scanline provides scanline storage containers for the AGG rendering pipeline.
// This file implements the scanline_storage_bin class from AGG's agg_scanline_storage_bin.h
package scanline

import (
	"encoding/binary"
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// SpanDataBin represents a span of binary pixels without coverage information.
// This corresponds to the span_data struct in AGG's scanline_storage_bin class.
type SpanDataBin struct {
	X   basics.Int32 // Starting X coordinate
	Len basics.Int32 // Length of the span
}

// ScanlineDataBin represents metadata for a binary scanline.
// This corresponds to the scanline_data struct in AGG's scanline_storage_bin class.
type ScanlineDataBin struct {
	Y         int // Y coordinate of the scanline
	NumSpans  int // Number of spans in this scanline
	StartSpan int // Index of first span in spans array
}

// EmbeddedScanlineBin is an embedded scanline that references data in ScanlineStorageBin.
// This corresponds to the embedded_scanline class in AGG's scanline_storage_bin.
type EmbeddedScanlineBin struct {
	storage     *ScanlineStorageBin // Reference to parent storage
	scanline    ScanlineDataBin     // Current scanline metadata
	scanlineIdx int                 // Index of current scanline
}

// EmbeddedScanlineBinIterator provides iteration over spans in an embedded scanline.
// This corresponds to the const_iterator class in AGG's embedded_scanline.
type EmbeddedScanlineBinIterator struct {
	storage *ScanlineStorageBin // Reference to parent storage
	spanIdx int                 // Current span index
	span    SpanDataBin         // Current span data
}

// NewEmbeddedScanlineBin creates a new embedded scanline for the given storage.
func NewEmbeddedScanlineBin(storage *ScanlineStorageBin) *EmbeddedScanlineBin {
	sl := &EmbeddedScanlineBin{
		storage: storage,
	}
	sl.Setup(0)
	return sl
}

// Reset is provided for interface compatibility but does nothing for embedded scanlines.
func (sl *EmbeddedScanlineBin) Reset(_, _ int) {}

// NumSpans returns the number of spans in the current scanline.
func (sl *EmbeddedScanlineBin) NumSpans() int {
	return sl.scanline.NumSpans
}

// Y returns the Y coordinate of the current scanline.
func (sl *EmbeddedScanlineBin) Y() int {
	return sl.scanline.Y
}

// Begin returns an iterator for the spans in this scanline.
func (sl *EmbeddedScanlineBin) Begin() *EmbeddedScanlineBinIterator {
	return &EmbeddedScanlineBinIterator{
		storage: sl.storage,
		spanIdx: sl.scanline.StartSpan,
		span:    sl.storage.SpanByIndex(sl.scanline.StartSpan),
	}
}

// Setup configures the embedded scanline to reference the specified scanline index.
func (sl *EmbeddedScanlineBin) Setup(scanlineIdx int) {
	sl.scanlineIdx = scanlineIdx
	sl.scanline = sl.storage.ScanlineByIndex(scanlineIdx)
}

// Span returns the current span data.
func (it *EmbeddedScanlineBinIterator) Span() SpanDataBin {
	return it.span
}

// Next advances to the next span.
// This corresponds to AGG's operator++() for the const_iterator.
// Note: AGG doesn't return a value from operator++, it just advances the iterator.
// The caller is responsible for checking bounds via the parent scanline's num_spans.
func (it *EmbeddedScanlineBinIterator) Next() {
	it.spanIdx++
	it.span = it.storage.SpanByIndex(it.spanIdx)
}

// ScanlineStorageBin is a storage container for binary scanlines.
// It stores scanlines without anti-aliasing information, only indicating
// which pixels are covered. This corresponds to AGG's scanline_storage_bin class.
type ScanlineStorageBin struct {
	spans        *array.PodBVector[SpanDataBin]     // Storage for span data
	scanlines    *array.PodBVector[ScanlineDataBin] // Storage for scanline metadata
	fakeSpan     SpanDataBin                        // Sentinel span for out-of-bounds access
	fakeScanline ScanlineDataBin                    // Sentinel scanline for out-of-bounds access
	minX         int                                // Minimum X coordinate
	minY         int                                // Minimum Y coordinate
	maxX         int                                // Maximum X coordinate
	maxY         int                                // Maximum Y coordinate
	curScanline  int                                // Current scanline index for iteration
}

// NewScanlineStorageBin creates a new binary scanline storage container.
func NewScanlineStorageBin() *ScanlineStorageBin {
	// Use block increment size of 256-2 = 254 to match AGG default exactly
	// AGG uses pod_bvector<span_data, 10> which gives 1024 elements per block, but with increment of 254
	// We need to replicate this behavior with our block vector implementation
	spansScale := array.NewBlockScale(10)   // 1024 elements per block to match AGG's pod_bvector<span_data, 10>
	scanlineScale := array.NewBlockScale(8) // 256 elements per block to match AGG's pod_bvector<scanline_data, 8>

	storage := &ScanlineStorageBin{
		spans:       array.NewPodBVectorWithScale[SpanDataBin](spansScale),
		scanlines:   array.NewPodBVectorWithScale[ScanlineDataBin](scanlineScale),
		minX:        math.MaxInt32,
		minY:        math.MaxInt32,
		maxX:        math.MinInt32,
		maxY:        math.MinInt32,
		curScanline: 0,
	}

	// Initialize fake span and scanline
	storage.fakeScanline.Y = 0
	storage.fakeScanline.NumSpans = 0
	storage.fakeScanline.StartSpan = 0
	storage.fakeSpan.X = 0
	storage.fakeSpan.Len = 0

	return storage
}

// Prepare clears the storage and resets bounds for new rendering.
// This corresponds to AGG's prepare() method.
func (s *ScanlineStorageBin) Prepare() {
	s.scanlines.RemoveAll()
	s.spans.RemoveAll()
	s.minX = math.MaxInt32
	s.minY = math.MaxInt32
	s.maxX = math.MinInt32
	s.maxY = math.MinInt32
	s.curScanline = 0
}

// Render processes a scanline and stores its span data.
// This corresponds to AGG's template render() method.
func (s *ScanlineStorageBin) Render(sl ScanlineInterface) {
	var slThis ScanlineDataBin

	y := sl.Y()
	if y < s.minY {
		s.minY = y
	}
	if y > s.maxY {
		s.maxY = y
	}

	slThis.Y = y
	slThis.NumSpans = sl.NumSpans()
	slThis.StartSpan = s.spans.Size()

	// Iterate over spans in the input scanline
	iterator := sl.Begin()
	numSpans := slThis.NumSpans

	for i := 0; i < numSpans; i++ {
		spanInfo := iterator.GetSpan()

		sp := SpanDataBin{
			X:   basics.Int32(spanInfo.X),
			Len: basics.Int32(int(math.Abs(float64(spanInfo.Len)))),
		}
		s.spans.Add(sp)

		x1 := int(sp.X)
		x2 := int(sp.X + sp.Len - 1)
		if x1 < s.minX {
			s.minX = x1
		}
		if x2 > s.maxX {
			s.maxX = x2
		}

		if i < numSpans-1 {
			iterator.Next()
		}
	}

	s.scanlines.Add(slThis)
}

// RenderBinScanline processes a binary scanline and stores its span data.
// This is a specialized version for binary scanlines.
func (s *ScanlineStorageBin) RenderBinScanline(sl *ScanlineBin) {
	var slThis ScanlineDataBin

	y := sl.Y()
	if y < s.minY {
		s.minY = y
	}
	if y > s.maxY {
		s.maxY = y
	}

	slThis.Y = y
	slThis.NumSpans = sl.NumSpans()
	slThis.StartSpan = s.spans.Size()

	// Get spans from binary scanline
	spans := sl.Spans()
	for _, binSpan := range spans {
		sp := SpanDataBin{
			X:   basics.Int32(binSpan.X),
			Len: basics.Int32(binSpan.Len),
		}
		s.spans.Add(sp)

		x1 := int(sp.X)
		x2 := int(sp.X + sp.Len - 1)
		if x1 < s.minX {
			s.minX = x1
		}
		if x2 > s.maxX {
			s.maxX = x2
		}
	}

	s.scanlines.Add(slThis)
}

// MinX returns the minimum X coordinate across all stored spans.
func (s *ScanlineStorageBin) MinX() int {
	return s.minX
}

// MinY returns the minimum Y coordinate across all stored scanlines.
func (s *ScanlineStorageBin) MinY() int {
	return s.minY
}

// MaxX returns the maximum X coordinate across all stored spans.
func (s *ScanlineStorageBin) MaxX() int {
	return s.maxX
}

// MaxY returns the maximum Y coordinate across all stored scanlines.
func (s *ScanlineStorageBin) MaxY() int {
	return s.maxY
}

// RewindScanlines resets the iterator to the beginning and returns true if there are scanlines.
// This corresponds to AGG's rewind_scanlines() method.
func (s *ScanlineStorageBin) RewindScanlines() bool {
	s.curScanline = 0
	return s.scanlines.Size() > 0
}

// SweepScanline fills the provided scanline with the next stored scanline data.
// This corresponds to AGG's template sweep_scanline() method.
func (s *ScanlineStorageBin) SweepScanline(sl ScanlineInterface) bool {
	sl.ResetSpans()

	for {
		if s.curScanline >= s.scanlines.Size() {
			return false
		}

		slThis := s.scanlines.At(s.curScanline)

		numSpans := slThis.NumSpans
		spanIdx := slThis.StartSpan

		for i := 0; i < numSpans; i++ {
			sp := s.spans.At(spanIdx + i)
			sl.AddSpan(int(sp.X), int(sp.Len), basics.CoverFull)
		}

		s.curScanline++
		if sl.NumSpans() > 0 {
			sl.Finalize(slThis.Y)
			break
		}
	}

	return true
}

// SweepEmbeddedScanline fills the provided embedded scanline with the next stored scanline.
// This is a specialization for embedded_scanline.
func (s *ScanlineStorageBin) SweepEmbeddedScanline(sl *EmbeddedScanlineBin) bool {
	for {
		if s.curScanline >= s.scanlines.Size() {
			return false
		}

		sl.Setup(s.curScanline)
		s.curScanline++

		if sl.NumSpans() > 0 {
			break
		}
	}

	return true
}

// ByteSize calculates the number of bytes needed to serialize this storage.
// This corresponds to AGG's byte_size() method.
func (s *ScanlineStorageBin) ByteSize() int {
	size := 4 * 4 // min_x, min_y, max_x, max_y (4 bytes each)

	for i := 0; i < s.scanlines.Size(); i++ {
		slThis := s.scanlines.At(i)
		size += 4 * 2                   // Y, num_spans (4 bytes each)
		size += slThis.NumSpans * 4 * 2 // X, span_len for each span (4 bytes each)
	}

	return size
}

// WriteInt32 writes a 32-bit integer to the byte slice at the specified offset.
// This corresponds to AGG's write_int32() static method.
func WriteInt32(dst []byte, offset int, val basics.Int32) {
	binary.LittleEndian.PutUint32(dst[offset:], uint32(val))
}

// Serialize writes the storage data to a byte slice in AGG's binary format.
// This corresponds to AGG's serialize() method.
func (s *ScanlineStorageBin) Serialize(data []byte) {
	offset := 0

	// Write bounds
	WriteInt32(data, offset, basics.Int32(s.minX))
	offset += 4
	WriteInt32(data, offset, basics.Int32(s.minY))
	offset += 4
	WriteInt32(data, offset, basics.Int32(s.maxX))
	offset += 4
	WriteInt32(data, offset, basics.Int32(s.maxY))
	offset += 4

	// Write scanlines
	for i := 0; i < s.scanlines.Size(); i++ {
		slThis := s.scanlines.At(i)

		WriteInt32(data, offset, basics.Int32(slThis.Y))
		offset += 4

		WriteInt32(data, offset, basics.Int32(slThis.NumSpans))
		offset += 4

		// Write spans for this scanline
		spanIdx := slThis.StartSpan
		for j := 0; j < slThis.NumSpans; j++ {
			sp := s.spans.At(spanIdx + j)

			WriteInt32(data, offset, sp.X)
			offset += 4

			WriteInt32(data, offset, sp.Len)
			offset += 4
		}
	}
}

// ScanlineByIndex returns the scanline at the specified index.
// Returns a fake scanline if the index is out of bounds.
// This corresponds to AGG's scanline_by_index() method.
func (s *ScanlineStorageBin) ScanlineByIndex(i int) ScanlineDataBin {
	if i < s.scanlines.Size() {
		return s.scanlines.At(i)
	}
	return s.fakeScanline
}

// SpanByIndex returns the span at the specified index.
// Returns a fake span if the index is out of bounds.
// This corresponds to AGG's span_by_index() method.
func (s *ScanlineStorageBin) SpanByIndex(i int) SpanDataBin {
	if i < s.spans.Size() {
		return s.spans.At(i)
	}
	return s.fakeSpan
}

// Compile-time interface compliance checks
// Ensure ScanlineStorageBin implements the RasterizerInterface from boolean_algebra.go
var _ interface {
	RewindScanlines() bool
	// SweepScanline method signature varies by implementation, but basic bounds methods must match
	MinX() int
	MinY() int
	MaxX() int
	MaxY() int
} = (*ScanlineStorageBin)(nil)

// Ensure EmbeddedScanlineBin implements core scanline interface methods
var _ interface {
	Y() int
	NumSpans() int
	Reset(int, int)
} = (*EmbeddedScanlineBin)(nil)
