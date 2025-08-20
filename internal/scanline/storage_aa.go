// Package scanline provides scanline storage containers for the AGG rendering pipeline.
// This file implements the scanline_cell_storage and scanline_storage_aa classes from AGG's agg_scanline_storage_aa.h
package scanline

import (
	"math"
	"unsafe"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// ScanlineInterface defines the interface for scanline containers.
// This is a local copy to avoid circular imports. Using basics.Int8u for simplicity.
type ScanlineInterface interface {
	Y() int
	NumSpans() int
	Begin() ScanlineIterator
	// Optional methods for modifying scanlines
	ResetSpans()
	AddSpan(x, len int, cover basics.Int8u)
	AddCells(x, len int, covers []basics.Int8u)
	Finalize(y int)
}

// ScanlineIterator provides iteration over spans in a scanline.
type ScanlineIterator interface {
	GetSpan() SpanInfo
	Next() bool
}

// SpanInfo represents span data for iteration.
type SpanInfo struct {
	X      int
	Len    int
	Covers []basics.Int8u
}

// ExtraSpan represents a dynamically allocated span of cells.
// This corresponds to the extra_span struct in AGG's scanline_cell_storage.
type ExtraSpan[T any] struct {
	Len int // Number of cells in this span
	Ptr []T // Slice containing the cells
}

// ScanlineCellStorage is a storage container for scanline cell data.
// It uses block-based storage for efficiency with fallback to dynamic allocation
// for spans that don't fit in the main storage blocks.
//
// This is equivalent to AGG's scanline_cell_storage<T> template class.
type ScanlineCellStorage[T any] struct {
	cells        *array.PodBVector[T]            // Primary block-based storage
	extraStorage *array.PodBVector[ExtraSpan[T]] // Overflow storage for large spans (pod_bvector<extra_span, 6>)
	allocator    basics.PodAllocator[T]          // Allocator for extra spans
}

// NewScanlineCellStorage creates a new scanline cell storage container.
// The initial block size is set to match AGG's default: pod_bvector<T, 12> with increment 128-2.
func NewScanlineCellStorage[T any]() *ScanlineCellStorage[T] {
	// Create PodBVector with custom block size to match C++: pod_bvector<T, 12>
	// The 12 is the block scale (2^12 = 4096), and 128-2 = 126 is the block increment
	cellStorage := array.NewPodBVectorWithIncrement[T](array.NewBlockScale(12), 126)

	return &ScanlineCellStorage[T]{
		cells:        cellStorage,
		extraStorage: array.NewPodBVectorWithScale[ExtraSpan[T]](array.NewBlockScale(6)), // pod_bvector<extra_span, 6>
		allocator:    basics.NewPodAllocator[T](),
	}
}

// NewScanlineCellStorageCopy creates a new storage as a copy of another.
// This implements the copy constructor behavior from AGG.
func NewScanlineCellStorageCopy[T any](other *ScanlineCellStorage[T]) *ScanlineCellStorage[T] {
	if other == nil {
		return NewScanlineCellStorage[T]()
	}

	storage := &ScanlineCellStorage[T]{
		cells:        array.NewPodBVectorCopy(other.cells),
		extraStorage: array.NewPodBVectorWithScale[ExtraSpan[T]](array.NewBlockScale(6)),
		allocator:    basics.NewPodAllocator[T](),
	}

	storage.copyExtraStorage(other)
	return storage
}

// RemoveAll clears all stored data and releases extra storage.
// This corresponds to AGG's remove_all() method.
func (s *ScanlineCellStorage[T]) RemoveAll() {
	// Clear extra storage and deallocate pod_allocator blocks (matching C++ behavior)
	for i := s.extraStorage.Size() - 1; i >= 0; i-- {
		extraSpan := s.extraStorage.At(i)
		if extraSpan.Ptr != nil {
			s.allocator.Deallocate(extraSpan.Ptr, extraSpan.Len)
		}
	}
	s.extraStorage.RemoveAll()

	// Clear the main cell storage
	s.cells.RemoveAll()
}

// AddCells adds a sequence of cells to the storage.
// Returns the index where the cells were stored. Positive indices indicate
// storage in the main cells array, negative indices indicate extra storage.
// This corresponds to AGG's add_cells() method.
func (s *ScanlineCellStorage[T]) AddCells(cells []T, numCells int) int {
	if numCells <= 0 || len(cells) < numCells {
		return -1
	}

	// Try to allocate in the main block storage first
	idx := s.cells.AllocateContinuousBlock(numCells)
	if idx >= 0 {
		// Success - copy data to main storage
		for i := 0; i < numCells; i++ {
			s.cells.Set(idx+i, cells[i])
		}
		return idx
	}

	// Main storage couldn't accommodate - use extra storage
	extraSpan := ExtraSpan[T]{
		Len: numCells,
		Ptr: s.allocator.Allocate(numCells),
	}

	// Copy the data
	copy(extraSpan.Ptr, cells[:numCells])

	// Add to extra storage
	s.extraStorage.Add(extraSpan)

	// Return negative index indicating extra storage
	return -s.extraStorage.Size()
}

// Get returns a pointer to the cells at the specified index.
// Positive indices access the main storage, negative indices access extra storage.
// Returns nil if the index is invalid.
// This corresponds to AGG's operator[] const overload.
func (s *ScanlineCellStorage[T]) Get(idx int) []T {
	if idx >= 0 {
		// Positive index - access main storage
		if idx >= s.cells.Size() {
			return nil
		}
		// For positive indices, we need to return a slice that points to the
		// data at the specified index. Since PodBVector doesn't expose
		// direct slice access, we'll need to create a temporary slice.
		// In practice, this is used to access the start of a continuous block.
		return s.getBlockSlice(idx)
	}

	// Negative index - access extra storage
	extraIdx := -idx - 1
	if extraIdx >= s.extraStorage.Size() {
		return nil
	}

	return s.extraStorage.At(extraIdx).Ptr
}

// getBlockSlice returns a slice starting at the given index in the main storage.
// This is a helper method to work with PodBVector's indexed access.
func (s *ScanlineCellStorage[T]) getBlockSlice(startIdx int) []T {
	// Since we can't get direct slice access to PodBVector, we'll create
	// a slice that captures the continuous block that was allocated.
	// We'll determine the size by looking at available space in the current block.
	remaining := s.cells.Size() - startIdx
	if remaining <= 0 {
		return nil
	}

	// Create a slice with the data from the PodBVector
	result := make([]T, remaining)
	for i := 0; i < remaining; i++ {
		result[i] = s.cells.At(startIdx + i)
	}
	return result
}

// GetMutable returns a mutable reference to cells at the specified index.
// This corresponds to AGG's operator[] non-const overload.
func (s *ScanlineCellStorage[T]) GetMutable(idx int) []T {
	return s.Get(idx) // In Go, slices are already mutable references
}

// Assign copies data from another ScanlineCellStorage instance.
// This corresponds to AGG's operator= method.
func (s *ScanlineCellStorage[T]) Assign(other *ScanlineCellStorage[T]) {
	if other == nil || s == other {
		return
	}

	// Clear existing data
	s.RemoveAll()

	// Copy the main cell storage
	s.cells = array.NewPodBVectorCopy(other.cells)

	// Copy extra storage
	s.copyExtraStorage(other)
}

// copyExtraStorage copies extra storage from another instance.
// This is a private helper method used by the copy constructor.
func (s *ScanlineCellStorage[T]) copyExtraStorage(other *ScanlineCellStorage[T]) {
	if other == nil || other.extraStorage.Size() == 0 {
		return
	}

	// Deep copy each extra span
	for i := 0; i < other.extraStorage.Size(); i++ {
		srcSpan := other.extraStorage.At(i)
		dstSpan := ExtraSpan[T]{
			Len: srcSpan.Len,
			Ptr: s.allocator.Allocate(srcSpan.Len),
		}
		copy(dstSpan.Ptr, srcSpan.Ptr)
		s.extraStorage.Add(dstSpan)
	}
}

// SpanData represents a span within a scanline for storage.
// This corresponds to the span_data struct in AGG's scanline_storage_aa.
type SpanData struct {
	X        basics.Int32 // Starting X coordinate
	Len      basics.Int32 // Length (if negative, it's a solid span)
	CoversID int          // Index of cells in the ScanlineCellStorage
}

// ScanlineData represents metadata for a complete scanline.
// This corresponds to the scanline_data struct in AGG's scanline_storage_aa.
type ScanlineData struct {
	Y         int // Y coordinate of the scanline
	NumSpans  int // Number of spans in this scanline
	StartSpan int // Starting index in the spans array
}

// ScanlineStorageAA is a storage container for anti-aliased scanlines.
// It stores complete scanlines with coverage data for later rendering.
// This corresponds to AGG's scanline_storage_aa<T> template class.
type ScanlineStorageAA[T any] struct {
	covers       *ScanlineCellStorage[T]         // Storage for coverage data
	spans        *array.PodBVector[SpanData]     // Storage for span data
	scanlines    *array.PodBVector[ScanlineData] // Storage for scanline metadata
	fakeSpan     SpanData                        // Fallback span for invalid indices
	fakeScanline ScanlineData                    // Fallback scanline for invalid indices
	minX         int                             // Minimum X coordinate
	minY         int                             // Minimum Y coordinate
	maxX         int                             // Maximum X coordinate
	maxY         int                             // Maximum Y coordinate
	curScanline  int                             // Current scanline index for iteration
}

// NewScanlineStorageAA creates a new anti-aliased scanline storage container.
func NewScanlineStorageAA[T any]() *ScanlineStorageAA[T] {
	// Create storage components with initial block sizes matching AGG
	spans := array.NewPodBVectorWithIncrement[SpanData](array.NewBlockScale(10), 256-2) // pod_bvector<span_data, 10>
	scanlines := array.NewPodBVectorWithScale[ScanlineData](array.NewBlockScale(8))     // pod_bvector<scanline_data, 8>

	storage := &ScanlineStorageAA[T]{
		covers:      NewScanlineCellStorage[T](),
		spans:       spans,
		scanlines:   scanlines,
		minX:        math.MaxInt32,
		minY:        math.MaxInt32,
		maxX:        math.MinInt32,
		maxY:        math.MinInt32,
		curScanline: 0,
	}

	// Initialize fake data
	storage.fakeScanline.Y = 0
	storage.fakeScanline.NumSpans = 0
	storage.fakeScanline.StartSpan = 0
	storage.fakeSpan.X = 0
	storage.fakeSpan.Len = 0
	storage.fakeSpan.CoversID = 0

	return storage
}

// Prepare clears all stored data and resets bounds for new rendering.
// This implements the renderer interface method.
func (s *ScanlineStorageAA[T]) Prepare() {
	s.covers.RemoveAll()
	s.scanlines.RemoveAll()
	s.spans.RemoveAll()
	s.minX = math.MaxInt32
	s.minY = math.MaxInt32
	s.maxX = math.MinInt32
	s.maxY = math.MinInt32
	s.curScanline = 0
}

// Render stores a scanline for later rendering.
// This implements the renderer interface method.
func (s *ScanlineStorageAA[T]) Render(sl ScanlineInterface) {
	var slData ScanlineData

	y := sl.Y()
	if y < s.minY {
		s.minY = y
	}
	if y > s.maxY {
		s.maxY = y
	}

	slData.Y = y
	slData.NumSpans = sl.NumSpans()
	slData.StartSpan = s.spans.Size()

	// Iterate through spans in the scanline
	spanIter := sl.Begin()
	numSpans := slData.NumSpans

	for i := 0; i < numSpans; i++ {
		span := spanIter.GetSpan()

		var sp SpanData
		sp.X = basics.Int32(span.X)
		sp.Len = basics.Int32(span.Len)

		// Calculate actual length for bounds checking
		length := span.Len
		if length < 0 {
			length = -length
		}

		// Store the coverage data using proper type conversion
		var coversT []T
		if length > 0 && len(span.Covers) > 0 {
			// Use the minimum of length and available covers
			actualLength := length
			if len(span.Covers) < length {
				actualLength = len(span.Covers)
			}
			coversT = make([]T, actualLength)
			// Use unsafe conversion for proper type handling (matching C++ memcpy)
			if actualLength > 0 {
				srcPtr := unsafe.Pointer(&span.Covers[0])
				dstPtr := unsafe.Pointer(&coversT[0])
				srcSize := actualLength * int(unsafe.Sizeof(span.Covers[0]))
				dstSize := actualLength * int(unsafe.Sizeof(coversT[0]))
				// Only copy if sizes match (same underlying type)
				if srcSize == dstSize {
					for i := 0; i < srcSize; i++ {
						*(*byte)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i))) =
							*(*byte)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i)))
					}
				}
			}
		}
		sp.CoversID = s.covers.AddCells(coversT, len(coversT))
		s.spans.Add(sp)

		// Update bounds
		x1 := span.X
		x2 := span.X + length - 1
		if x1 < s.minX {
			s.minX = x1
		}
		if x2 > s.maxX {
			s.maxX = x2
		}

		if !spanIter.Next() {
			break
		}
	}

	s.scanlines.Add(slData)
}

// MinX returns the minimum X coordinate of all stored scanlines.
func (s *ScanlineStorageAA[T]) MinX() int {
	return s.minX
}

// MinY returns the minimum Y coordinate of all stored scanlines.
func (s *ScanlineStorageAA[T]) MinY() int {
	return s.minY
}

// MaxX returns the maximum X coordinate of all stored scanlines.
func (s *ScanlineStorageAA[T]) MaxX() int {
	return s.maxX
}

// MaxY returns the maximum Y coordinate of all stored scanlines.
func (s *ScanlineStorageAA[T]) MaxY() int {
	return s.maxY
}

// RewindScanlines prepares for scanline iteration and returns true if there are scanlines.
func (s *ScanlineStorageAA[T]) RewindScanlines() bool {
	s.curScanline = 0
	return s.scanlines.Size() > 0
}

// SweepScanline fills the provided scanline with the next stored scanline data.
// Returns true if a scanline was filled, false if iteration is complete.
func (s *ScanlineStorageAA[T]) SweepScanline(sl ScanlineInterface) bool {
	// Continue until we find a scanline with spans or reach the end
	for {
		if s.curScanline >= s.scanlines.Size() {
			return false
		}

		slThis := s.scanlines.At(s.curScanline)

		numSpans := slThis.NumSpans
		spanIdx := slThis.StartSpan

		// Reset scanline before adding spans
		sl.ResetSpans()

		// Add all spans to the scanline
		for i := 0; i < numSpans; i++ {
			sp := s.spans.At(spanIdx + i)
			covers := s.covers.Get(sp.CoversID)

			if sp.Len < 0 {
				// Solid span - use first coverage value with proper type conversion
				if len(covers) > 0 {
					// Use unsafe conversion for proper type handling
					var cover basics.Int8u
					if unsafe.Sizeof(covers[0]) == unsafe.Sizeof(cover) {
						srcPtr := unsafe.Pointer(&covers[0])
						dstPtr := unsafe.Pointer(&cover)
						*(*byte)(dstPtr) = *(*byte)(srcPtr)
					}
					sl.AddSpan(int(sp.X), int(-sp.Len), cover)
				}
			} else {
				// Coverage span - convert []T to []basics.Int8u with proper type handling
				coversInt8u := make([]basics.Int8u, len(covers))
				if len(covers) > 0 && unsafe.Sizeof(covers[0]) == unsafe.Sizeof(coversInt8u[0]) {
					// Direct memory copy for compatible types
					srcPtr := unsafe.Pointer(&covers[0])
					dstPtr := unsafe.Pointer(&coversInt8u[0])
					size := len(covers) * int(unsafe.Sizeof(coversInt8u[0]))
					for i := 0; i < size; i++ {
						*(*byte)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i))) =
							*(*byte)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i)))
					}
				}
				sl.AddCells(int(sp.X), int(sp.Len), coversInt8u)
			}
		}

		s.curScanline++

		// If scanline has spans, finalize and return
		if sl.NumSpans() > 0 {
			sl.Finalize(slThis.Y)
			break
		}
	}

	return true
}

// SweepEmbeddedScanline provides specialized sweep for embedded scanlines.
// Returns true if a scanline was found, false if iteration is complete.
func (s *ScanlineStorageAA[T]) SweepEmbeddedScanline(sl *EmbeddedScanline[T]) bool {
	for {
		if s.curScanline >= s.scanlines.Size() {
			return false
		}
		sl.Init(s, s.curScanline)
		s.curScanline++
		if sl.NumSpans() > 0 {
			return true
		}
	}
}

// ScanlineByIndex returns scanline data by index with bounds checking.
func (s *ScanlineStorageAA[T]) ScanlineByIndex(i int) ScanlineData {
	if i < s.scanlines.Size() {
		return s.scanlines.At(i)
	}
	return s.fakeScanline
}

// SpanByIndex returns span data by index with bounds checking.
func (s *ScanlineStorageAA[T]) SpanByIndex(i int) SpanData {
	if i < s.spans.Size() {
		return s.spans.At(i)
	}
	return s.fakeSpan
}

// CoversByIndex returns coverage data by index.
func (s *ScanlineStorageAA[T]) CoversByIndex(i int) []T {
	return s.covers.Get(i)
}

// writeInt32 writes a 32-bit integer in little-endian byte order.
// This matches AGG's write_int32 static method.
func writeInt32(dst []byte, val basics.Int32) {
	dst[0] = byte(val)
	dst[1] = byte(val >> 8)
	dst[2] = byte(val >> 16)
	dst[3] = byte(val >> 24)
}

// ByteSize calculates the total size in bytes needed to serialize all stored scanlines.
// This corresponds to AGG's byte_size() method.
func (s *ScanlineStorageAA[T]) ByteSize() int {
	var size int
	// Size for min_x, min_y, max_x, max_y (4 int32 values)
	size = 4 * 4 // 4 bytes per int32 * 4 values

	// Calculate size for each scanline
	for i := 0; i < s.scanlines.Size(); i++ {
		// Size for scanline header: scanline_size, Y, num_spans (3 int32 values)
		size += 3 * 4

		slThis := s.scanlines.At(i)
		numSpans := slThis.NumSpans
		spanIdx := slThis.StartSpan

		// Calculate size for each span
		for j := 0; j < numSpans; j++ {
			if spanIdx+j >= s.spans.Size() {
				break // Safety check to prevent index out of bounds
			}

			sp := s.spans.At(spanIdx + j)

			// Size for span header: X, span_len (2 int32 values)
			size += 2 * 4

			// Size for coverage data
			if sp.Len < 0 {
				// Solid span - single coverage value
				size += int(unsafe.Sizeof(*new(T)))
			} else {
				// Coverage array - multiple coverage values
				size += int(sp.Len) * int(unsafe.Sizeof(*new(T)))
			}
		}
	}

	return size
}

// Serialize writes all stored scanline data to a byte buffer.
// The data is written in AGG's serialization format for cross-platform compatibility.
// This corresponds to AGG's serialize() method.
func (s *ScanlineStorageAA[T]) Serialize(data []byte) {
	if len(data) < s.ByteSize() {
		return // Not enough space
	}

	offset := 0

	// Write bounds (min_x, min_y, max_x, max_y)
	writeInt32(data[offset:], basics.Int32(s.minX))
	offset += 4
	writeInt32(data[offset:], basics.Int32(s.minY))
	offset += 4
	writeInt32(data[offset:], basics.Int32(s.maxX))
	offset += 4
	writeInt32(data[offset:], basics.Int32(s.maxY))
	offset += 4

	// Write each scanline
	for i := 0; i < s.scanlines.Size(); i++ {
		slThis := s.scanlines.At(i)

		// Remember position for scanline size
		sizePos := offset
		offset += 4 // Reserve space for scanline size

		// Write Y coordinate
		writeInt32(data[offset:], basics.Int32(slThis.Y))
		offset += 4

		// Write number of spans
		writeInt32(data[offset:], basics.Int32(slThis.NumSpans))
		offset += 4

		// Write each span
		numSpans := slThis.NumSpans
		spanIdx := slThis.StartSpan

		for j := 0; j < numSpans; j++ {
			if spanIdx+j >= s.spans.Size() {
				break // Safety check to prevent index out of bounds
			}

			sp := s.spans.At(spanIdx + j)
			covers := s.covers.Get(sp.CoversID)

			// Write span X coordinate
			writeInt32(data[offset:], sp.X)
			offset += 4

			// Write span length
			writeInt32(data[offset:], sp.Len)
			offset += 4

			// Write coverage data
			if sp.Len < 0 {
				// Solid span - write single coverage value
				if len(covers) > 0 {
					// Use memcpy equivalent with unsafe.Pointer for proper type T handling
					coverSize := int(unsafe.Sizeof(covers[0]))
					if offset+coverSize <= len(data) {
						// Direct memory copy using unsafe pointers (matching C++ std::memcpy)
						srcPtr := unsafe.Pointer(&covers[0])
						dstPtr := unsafe.Pointer(&data[offset])
						for i := 0; i < coverSize; i++ {
							*(*byte)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i))) =
								*(*byte)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i)))
						}
						offset += coverSize
					}
				}
			} else {
				// Coverage array - write all coverage values
				actualLen := int(sp.Len)
				if len(covers) < actualLen {
					actualLen = len(covers)
				}

				coverSize := int(unsafe.Sizeof(covers[0]))
				totalSize := actualLen * coverSize
				if offset+totalSize <= len(data) {
					// Bulk memory copy using unsafe pointers (matching C++ std::memcpy)
					if actualLen > 0 {
						srcPtr := unsafe.Pointer(&covers[0])
						dstPtr := unsafe.Pointer(&data[offset])
						for i := 0; i < totalSize; i++ {
							*(*byte)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i))) =
								*(*byte)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i)))
						}
						offset += totalSize
					}
				}
			}
		}

		// Write scanline size at the reserved position
		scanlineSize := offset - sizePos
		writeInt32(data[sizePos:], basics.Int32(scanlineSize))
	}
}

// EmbeddedScanline provides efficient iteration over stored scanlines without
// copying span data. This corresponds to AGG's embedded_scanline class.
type EmbeddedScanline[T any] struct {
	storage      *ScanlineStorageAA[T] // Reference to the storage
	scanlineData ScanlineData          // Current scanline metadata
	scanlineIdx  int                   // Current scanline index
}

// EmbeddedScanlineIterator provides iteration over spans in an embedded scanline.
// This corresponds to AGG's embedded_scanline::const_iterator class.
type EmbeddedScanlineIterator[T any] struct {
	storage     *ScanlineStorageAA[T] // Reference to the storage
	spanIdx     int                   // Current span index
	span        EmbeddedSpan[T]       // Current span data
	numSpans    int                   // Total number of spans
	currentSpan int                   // Current span counter for bounds checking
}

// EmbeddedSpan represents a span within an embedded scanline.
type EmbeddedSpan[T any] struct {
	X      basics.Int32 // Starting X coordinate
	Len    basics.Int32 // Length (if negative, it's a solid span)
	Covers []T          // Coverage data
}

// NewEmbeddedScanline creates a new embedded scanline for the given storage.
func NewEmbeddedScanline[T any](storage *ScanlineStorageAA[T]) *EmbeddedScanline[T] {
	return &EmbeddedScanline[T]{
		storage: storage,
	}
}

// Reset is a no-op for embedded scanlines (interface compatibility).
func (e *EmbeddedScanline[T]) Reset(minX, maxX int) {
	// No operation needed for embedded scanlines
}

// NumSpans returns the number of spans in the current scanline.
func (e *EmbeddedScanline[T]) NumSpans() int {
	return e.scanlineData.NumSpans
}

// Y returns the Y coordinate of the current scanline.
func (e *EmbeddedScanline[T]) Y() int {
	return e.scanlineData.Y
}

// Begin returns an iterator to the first span in the scanline.
func (e *EmbeddedScanline[T]) Begin() *EmbeddedScanlineIterator[T] {
	iter := &EmbeddedScanlineIterator[T]{
		storage:     e.storage,
		spanIdx:     e.scanlineData.StartSpan,
		numSpans:    e.scanlineData.NumSpans,
		currentSpan: 0,
	}
	if e.scanlineData.NumSpans > 0 {
		iter.initSpan()
	}
	return iter
}

// Init initializes the embedded scanline with data from the specified index.
func (e *EmbeddedScanline[T]) Init(storage *ScanlineStorageAA[T], scanlineIdx int) {
	e.storage = storage
	e.scanlineIdx = scanlineIdx
	e.scanlineData = storage.ScanlineByIndex(scanlineIdx)
}

// GetSpan returns the current span data.
func (it *EmbeddedScanlineIterator[T]) GetSpan() EmbeddedSpan[T] {
	return it.span
}

// Next advances to the next span and returns true if valid.
// This implements the C++ operator++() behavior.
func (it *EmbeddedScanlineIterator[T]) Next() bool {
	it.currentSpan++
	if it.currentSpan >= it.numSpans {
		return false // No more spans
	}
	it.spanIdx++
	it.initSpan()
	return true
}

// initSpan initializes the current span data.
func (it *EmbeddedScanlineIterator[T]) initSpan() {
	// Bounds check to match C++ safety
	if it.spanIdx >= 0 && it.currentSpan < it.numSpans {
		s := it.storage.SpanByIndex(it.spanIdx)
		it.span.X = s.X
		it.span.Len = s.Len
		it.span.Covers = it.storage.CoversByIndex(s.CoversID)
	} else {
		// Initialize with safe defaults for out-of-bounds access
		it.span.X = 0
		it.span.Len = 0
		it.span.Covers = nil
	}
}

// Concrete type aliases for common usage matching AGG's typedefs
type (
	ScanlineStorageAA8  = ScanlineStorageAA[basics.Int8u]  // 8-bit coverage
	ScanlineStorageAA16 = ScanlineStorageAA[basics.Int16u] // 16-bit coverage
	ScanlineStorageAA32 = ScanlineStorageAA[basics.Int32u] // 32-bit coverage
)
