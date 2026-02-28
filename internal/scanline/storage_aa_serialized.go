// Package scanline provides serialized scanline storage support for the AGG rendering pipeline.
// This file implements the serialized_scanlines_adaptor_aa class from AGG's agg_scanline_storage_aa.h
package scanline

import (
	"encoding/binary"
	"math"

	"agg_go/internal/basics"
)

// SerializedScanlinesAdaptorAA provides access to serialized scanline data.
// This corresponds to AGG's serialized_scanlines_adaptor_aa<T> template class.
type SerializedScanlinesAdaptorAA[T any] struct {
	data []byte // Serialized data buffer
	end  int    // End position in data
	ptr  int    // Current position in data
	dx   int    // X offset for coordinate transformation
	dy   int    // Y offset for coordinate transformation
	minX int    // Minimum X coordinate (with offset applied)
	minY int    // Minimum Y coordinate (with offset applied)
	maxX int    // Maximum X coordinate (with offset applied)
	maxY int    // Maximum Y coordinate (with offset applied)
}

// SerializedEmbeddedScanline provides iteration over serialized scanline data.
type SerializedEmbeddedScanline[T any] struct {
	data     []byte // Serialized data buffer
	ptr      int    // Current position in data buffer (points to first span header)
	end      int    // End position for current serialized scanline
	y        int    // Y coordinate of scanline
	numSpans int    // Number of spans in scanline
	dx       int    // X offset for coordinate transformation
}

// SerializedEmbeddedScanlineIterator provides iteration over spans in serialized data.
type SerializedEmbeddedScanlineIterator[T any] struct {
	data      []byte            // Serialized data buffer
	ptr       int               // Current position in data buffer
	end       int               // End position for current serialized scanline
	span      SerializedSpan[T] // Current span data
	dx        int               // X offset for coordinate transformation
	remaining int               // Number of spans remaining (including current)
	valid     bool              // Whether current span is valid
}

// SerializedSpan represents a span within serialized scanline data.
type SerializedSpan[T any] struct {
	X      basics.Int32 // Starting X coordinate (with offset applied)
	Len    basics.Int32 // Length (if negative, it's a solid span)
	Covers []T          // Coverage data (points into serialized buffer)
}

// NewSerializedScanlinesAdaptorAA creates a new adaptor for serialized scanline data.
func NewSerializedScanlinesAdaptorAA[T any](data []byte, size int, dx, dy float64) *SerializedScanlinesAdaptorAA[T] {
	clampedSize := clampSerializedSize(data, size)
	return &SerializedScanlinesAdaptorAA[T]{
		data: data,
		end:  clampedSize,
		ptr:  0,
		dx:   basics.IRound(dx),
		dy:   basics.IRound(dy),
		minX: math.MaxInt32,
		minY: math.MaxInt32,
		maxX: math.MinInt32,
		maxY: math.MinInt32,
	}
}

// NewSerializedScanlinesAdaptorAAEmpty creates an empty adaptor.
func NewSerializedScanlinesAdaptorAAEmpty[T any]() *SerializedScanlinesAdaptorAA[T] {
	return &SerializedScanlinesAdaptorAA[T]{
		minX: math.MaxInt32,
		minY: math.MaxInt32,
		maxX: math.MinInt32,
		maxY: math.MinInt32,
	}
}

// Init initializes the adaptor with new data and offsets.
func (s *SerializedScanlinesAdaptorAA[T]) Init(data []byte, size int, dx, dy float64) {
	clampedSize := clampSerializedSize(data, size)
	s.data = data
	s.end = clampedSize
	s.ptr = 0
	s.dx = basics.IRound(dx)
	s.dy = basics.IRound(dy)
	s.minX = math.MaxInt32
	s.minY = math.MaxInt32
	s.maxX = math.MinInt32
	s.maxY = math.MinInt32
}

// RewindScanlines prepares for scanline iteration.
func (s *SerializedScanlinesAdaptorAA[T]) RewindScanlines() bool {
	s.ptr = 0
	if s.ptr < s.end {
		s.minX = int(s.readInt32()) + s.dx
		s.minY = int(s.readInt32()) + s.dy
		s.maxX = int(s.readInt32()) + s.dx
		s.maxY = int(s.readInt32()) + s.dy
	}
	return s.ptr < s.end
}

// MinX returns the minimum X coordinate of all scanlines.
func (s *SerializedScanlinesAdaptorAA[T]) MinX() int {
	return s.minX
}

// MinY returns the minimum Y coordinate of all scanlines.
func (s *SerializedScanlinesAdaptorAA[T]) MinY() int {
	return s.minY
}

// MaxX returns the maximum X coordinate of all scanlines.
func (s *SerializedScanlinesAdaptorAA[T]) MaxX() int {
	return s.maxX
}

// MaxY returns the maximum Y coordinate of all scanlines.
func (s *SerializedScanlinesAdaptorAA[T]) MaxY() int {
	return s.maxY
}

// SweepScanline fills the provided scanline with the next serialized scanline data.
func (s *SerializedScanlinesAdaptorAA[T]) SweepScanline(sl ScanlineInterface) bool {
	for {
		if s.ptr >= s.end {
			return false
		}

		_ = s.readInt32() // Skip scanline size in bytes
		y := int(s.readInt32()) + s.dy
		numSpans := int(s.readInt32())

		sl.ResetSpans()

		for i := 0; i < numSpans; i++ {
			x := int(s.readInt32()) + s.dx
			length := int(s.readInt32())

			if length < 0 {
				// Solid span
				cover, ok := s.readCover()
				if !ok {
					return false
				}
				sl.AddSpan(x, -length, coverToInt8u(cover))
			} else {
				// Coverage span
				covers, ok := s.readCovers(length)
				if !ok {
					return false
				}
				coversInt8u := make([]basics.Int8u, len(covers))
				for i, c := range covers {
					coversInt8u[i] = coverToInt8u(c)
				}
				sl.AddCells(x, length, coversInt8u)
			}
		}

		if sl.NumSpans() > 0 {
			sl.Finalize(y)
			break
		}
	}
	return true
}

// SweepSerializedEmbeddedScanline provides specialized sweep for serialized embedded scanlines.
func (s *SerializedScanlinesAdaptorAA[T]) SweepSerializedEmbeddedScanline(sl *SerializedEmbeddedScanline[T]) bool {
	for {
		if s.ptr >= s.end {
			return false
		}

		byteSize := int(s.readInt32U())
		scanlineEnd := s.ptr + byteSize - 4 // subtract already consumed size field
		if scanlineEnd > s.end {
			scanlineEnd = s.end
		}
		sl.Init(s.data, s.ptr, scanlineEnd, s.dx, s.dy)
		s.ptr += byteSize - 4 // Advance past the scanline data (minus the size we already read)

		if sl.NumSpans() > 0 {
			return true
		}
	}
}

// readInt32 reads a 32-bit signed integer from the current position.
func (s *SerializedScanlinesAdaptorAA[T]) readInt32() basics.Int32 {
	if s.ptr+4 > s.end {
		return 0
	}
	val := binary.LittleEndian.Uint32(s.data[s.ptr : s.ptr+4])
	s.ptr += 4
	return basics.Int32(val)
}

// readInt32U reads a 32-bit unsigned integer from the current position.
func (s *SerializedScanlinesAdaptorAA[T]) readInt32U() basics.Int32u {
	if s.ptr+4 > s.end {
		return 0
	}
	val := binary.LittleEndian.Uint32(s.data[s.ptr : s.ptr+4])
	s.ptr += 4
	return basics.Int32u(val)
}

// readCover reads a single coverage value from the current position.
func (s *SerializedScanlinesAdaptorAA[T]) readCover() (T, bool) {
	coverSize := serializedCoverSize[T]()
	if s.ptr+coverSize > s.end {
		var zero T
		return zero, false
	}
	val := deserializeCover[T](s.data[s.ptr : s.ptr+coverSize])
	s.ptr += coverSize
	return val, true
}

// readCovers reads an array of coverage values from the current position.
func (s *SerializedScanlinesAdaptorAA[T]) readCovers(count int) ([]T, bool) {
	if count <= 0 {
		return nil, true
	}

	coverSize := serializedCoverSize[T]()
	byteCount := count * coverSize
	if s.ptr+byteCount > s.end {
		return nil, false
	}

	result := make([]T, count)
	for i := 0; i < count; i++ {
		start := s.ptr + i*coverSize
		result[i] = deserializeCover[T](s.data[start : start+coverSize])
	}
	s.ptr += byteCount
	return result, true
}

// SerializedEmbeddedScanline methods

// NewSerializedEmbeddedScanline creates a new embedded scanline for serialized data.
func NewSerializedEmbeddedScanline[T any]() *SerializedEmbeddedScanline[T] {
	return &SerializedEmbeddedScanline[T]{}
}

// Reset is a no-op for serialized embedded scanlines.
func (e *SerializedEmbeddedScanline[T]) Reset(minX, maxX int) {
	// No operation needed
}

// NumSpans returns the number of spans in the current scanline.
func (e *SerializedEmbeddedScanline[T]) NumSpans() int {
	return e.numSpans
}

// Y returns the Y coordinate of the current scanline.
func (e *SerializedEmbeddedScanline[T]) Y() int {
	return e.y
}

// Begin returns an iterator to the first span in the serialized scanline.
func (e *SerializedEmbeddedScanline[T]) Begin() *SerializedEmbeddedScanlineIterator[T] {
	iter := &SerializedEmbeddedScanlineIterator[T]{
		data:      e.data,
		ptr:       e.ptr,
		end:       e.end,
		dx:        e.dx,
		remaining: e.numSpans,
		valid:     false,
	}
	if e.numSpans > 0 {
		iter.initSpan()
	}
	return iter
}

// Init initializes the embedded scanline with serialized data.
func (e *SerializedEmbeddedScanline[T]) Init(data []byte, ptr, end, dx, dy int) {
	e.data = data
	e.ptr = ptr
	e.end = end
	e.dx = dx
	e.y = 0
	e.numSpans = 0

	// Read Y coordinate and number of spans
	if ptr+8 <= len(data) && ptr+8 <= end {
		e.y = int(int32(binary.LittleEndian.Uint32(data[ptr:ptr+4]))) + dy
		e.numSpans = int(int32(binary.LittleEndian.Uint32(data[ptr+4 : ptr+8])))
		if e.numSpans < 0 {
			e.numSpans = 0
		}
		e.ptr += 8 // Move past Y and numSpans
	}
}

// SerializedEmbeddedScanlineIterator methods

// GetSpan returns the current span data.
func (it *SerializedEmbeddedScanlineIterator[T]) GetSpan() SerializedSpan[T] {
	return it.span
}

// Next advances to the next span.
func (it *SerializedEmbeddedScanlineIterator[T]) Next() bool {
	if !it.valid {
		return false
	}

	it.remaining--
	if it.remaining <= 0 {
		it.valid = false
		it.span = SerializedSpan[T]{}
		return false
	}

	it.initSpan()
	return it.valid
}

// initSpan initializes the current span data from serialized buffer.
func (it *SerializedEmbeddedScanlineIterator[T]) initSpan() {
	if it.remaining <= 0 || it.ptr+8 > it.end || it.ptr+8 > len(it.data) {
		it.valid = false
		it.span = SerializedSpan[T]{}
		return
	}

	x := int(int32(binary.LittleEndian.Uint32(it.data[it.ptr:it.ptr+4]))) + it.dx
	length := int(int32(binary.LittleEndian.Uint32(it.data[it.ptr+4 : it.ptr+8])))
	it.ptr += 8

	coverSize := serializedCoverSize[T]()
	if length < 0 {
		if it.ptr+coverSize > it.end || it.ptr+coverSize > len(it.data) {
			it.valid = false
			it.span = SerializedSpan[T]{}
			return
		}

		cover := deserializeCover[T](it.data[it.ptr : it.ptr+coverSize])
		it.ptr += coverSize
		it.span = SerializedSpan[T]{
			X:      basics.Int32(x),
			Len:    basics.Int32(length),
			Covers: []T{cover},
		}
		it.valid = true
		return
	}

	byteCount := length * coverSize
	if byteCount < 0 || it.ptr+byteCount > it.end || it.ptr+byteCount > len(it.data) {
		it.valid = false
		it.span = SerializedSpan[T]{}
		return
	}

	covers := make([]T, length)
	for i := 0; i < length; i++ {
		start := it.ptr + i*coverSize
		covers[i] = deserializeCover[T](it.data[start : start+coverSize])
	}
	it.ptr += byteCount

	it.span = SerializedSpan[T]{
		X:      basics.Int32(x),
		Len:    basics.Int32(length),
		Covers: covers,
	}
	it.valid = true
}

// IsValid reports whether the iterator currently points to a valid span.
func (it *SerializedEmbeddedScanlineIterator[T]) IsValid() bool {
	return it.valid
}

// X returns the current span X coordinate.
func (it *SerializedEmbeddedScanlineIterator[T]) X() int {
	if !it.valid {
		return 0
	}
	return int(it.span.X)
}

// Len returns the current span length (negative means a solid span).
func (it *SerializedEmbeddedScanlineIterator[T]) Len() int {
	if !it.valid {
		return 0
	}
	return int(it.span.Len)
}

// Covers returns the current span coverage values.
func (it *SerializedEmbeddedScanlineIterator[T]) Covers() []T {
	if !it.valid {
		return nil
	}
	return it.span.Covers
}

func clampSerializedSize(data []byte, size int) int {
	if size < 0 {
		return 0
	}
	if size > len(data) {
		return len(data)
	}
	return size
}

func serializedCoverSize[T any]() int {
	var zero T
	switch any(zero).(type) {
	case uint8:
		return 1
	case uint16:
		return 2
	case uint32:
		return 4
	default:
		// AGG serialized adaptor supports uint8/uint16/uint32 cover types.
		// Default to 1 byte to remain backward-compatible with existing 8-bit paths.
		return 1
	}
}

func deserializeCover[T any](data []byte) T {
	var zero T
	switch any(zero).(type) {
	case uint8:
		if len(data) >= 1 {
			return any(data[0]).(T)
		}
	case uint16:
		if len(data) >= 2 {
			return any(binary.LittleEndian.Uint16(data[:2])).(T)
		}
	case uint32:
		if len(data) >= 4 {
			return any(binary.LittleEndian.Uint32(data[:4])).(T)
		}
	}
	return zero
}

func coverToInt8u[T any](cover T) basics.Int8u {
	switch v := any(cover).(type) {
	case uint8:
		return basics.Int8u(v)
	case uint16:
		return basics.Int8u(v)
	case uint32:
		return basics.Int8u(v)
	default:
		return 0
	}
}

// Concrete type aliases for common usage
type (
	SerializedScanlinesAdaptorAA8  = SerializedScanlinesAdaptorAA[basics.Int8u]  // 8-bit coverage
	SerializedScanlinesAdaptorAA16 = SerializedScanlinesAdaptorAA[basics.Int16u] // 16-bit coverage
	SerializedScanlinesAdaptorAA32 = SerializedScanlinesAdaptorAA[basics.Int32u] // 32-bit coverage
)
