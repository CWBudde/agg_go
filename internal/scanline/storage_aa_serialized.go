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
	ptr      int // Current position in data buffer
	y        int // Y coordinate of scanline
	numSpans int // Number of spans in scanline
	dx       int // X offset for coordinate transformation
}

// SerializedEmbeddedScanlineIterator provides iteration over spans in serialized data.
type SerializedEmbeddedScanlineIterator[T any] struct {
	ptr  int               // Current position in data buffer
	span SerializedSpan[T] // Current span data
	dx   int               // X offset for coordinate transformation
}

// SerializedSpan represents a span within serialized scanline data.
type SerializedSpan[T any] struct {
	X      basics.Int32 // Starting X coordinate (with offset applied)
	Len    basics.Int32 // Length (if negative, it's a solid span)
	Covers []T          // Coverage data (points into serialized buffer)
}

// NewSerializedScanlinesAdaptorAA creates a new adaptor for serialized scanline data.
func NewSerializedScanlinesAdaptorAA[T any](data []byte, size int, dx, dy float64) *SerializedScanlinesAdaptorAA[T] {
	return &SerializedScanlinesAdaptorAA[T]{
		data: data,
		end:  size,
		ptr:  0,
		dx:   int(dx + 0.5), // Round to nearest integer
		dy:   int(dy + 0.5),
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
	s.data = data
	s.end = size
	s.ptr = 0
	s.dx = int(dx + 0.5)
	s.dy = int(dy + 0.5)
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
				cover := s.readCover().(basics.Int8u)
				sl.AddSpan(x, -length, cover)
			} else {
				// Coverage span
				covers := s.readCovers(length)
				coversInt8u := make([]basics.Int8u, len(covers))
				for i, c := range covers {
					coversInt8u[i] = c.(basics.Int8u)
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
		sl.Init(s.data, s.ptr, s.dx, s.dy)
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
// This is simplified to work with basics.Int8u
func (s *SerializedScanlinesAdaptorAA[T]) readCover() any {
	if s.ptr+1 > s.end {
		return basics.Int8u(0)
	}
	val := s.data[s.ptr]
	s.ptr++
	return basics.Int8u(val)
}

// readCovers reads an array of coverage values from the current position.
// This is simplified to work with basics.Int8u
func (s *SerializedScanlinesAdaptorAA[T]) readCovers(count int) []any {
	if s.ptr+count > s.end {
		return nil
	}

	result := make([]any, count)
	for i := 0; i < count; i++ {
		result[i] = basics.Int8u(s.data[s.ptr+i])
	}
	s.ptr += count
	return result
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
		ptr: e.ptr,
		dx:  e.dx,
	}
	iter.initSpan()
	return iter
}

// Init initializes the embedded scanline with serialized data.
func (e *SerializedEmbeddedScanline[T]) Init(data []byte, ptr, dx, dy int) {
	e.ptr = ptr
	e.dx = dx

	// Read Y coordinate and number of spans
	if ptr+8 <= len(data) {
		e.y = int(binary.LittleEndian.Uint32(data[ptr:ptr+4])) + dy
		e.numSpans = int(binary.LittleEndian.Uint32(data[ptr+4 : ptr+8]))
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
	// Advance past current span data
	// For simplicity, assume T is basics.Int8u (1 byte)
	if it.span.Len < 0 {
		it.ptr += 1 // Size of one coverage value
	} else {
		it.ptr += int(it.span.Len) // Size of coverage array
	}
	it.initSpan()
	return true // In this implementation, we assume bounds are checked externally
}

// initSpan initializes the current span data from serialized buffer.
func (it *SerializedEmbeddedScanlineIterator[T]) initSpan() {
	// This is a simplified implementation that assumes the data buffer contains valid data
	// In a real implementation, we'd need proper bounds checking and error handling

	// For now, this is a placeholder implementation
	it.span.X = 0
	it.span.Len = 0
	it.span.Covers = nil
}

// Concrete type aliases for common usage
type (
	SerializedScanlinesAdaptorAA8  = SerializedScanlinesAdaptorAA[basics.Int8u]  // 8-bit coverage
	SerializedScanlinesAdaptorAA16 = SerializedScanlinesAdaptorAA[basics.Int16u] // 16-bit coverage
	SerializedScanlinesAdaptorAA32 = SerializedScanlinesAdaptorAA[basics.Int32u] // 32-bit coverage
)
