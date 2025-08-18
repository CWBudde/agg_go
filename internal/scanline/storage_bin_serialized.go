// Package scanline provides scanline storage containers for the AGG rendering pipeline.
// This file implements the serialized_scanlines_adaptor_bin class from AGG's agg_scanline_storage_bin.h
package scanline

import (
	"encoding/binary"
	"math"

	"agg_go/internal/basics"
)

// EmbeddedScanlineSerial is an embedded scanline that reads from serialized binary data.
// This corresponds to the embedded_scanline class in AGG's serialized_scanlines_adaptor_bin.
type EmbeddedScanlineSerial struct {
	ptr      []byte // Current position in serialized data
	y        int    // Y coordinate of current scanline
	numSpans int    // Number of spans in current scanline
	dx       int    // X offset for coordinate transformation
}

// EmbeddedScanlineSerialIterator provides iteration over spans in serialized data.
// This corresponds to the const_iterator class in AGG's embedded_scanline.
type EmbeddedScanlineSerialIterator struct {
	ptr  []byte     // Current position in span data
	span SpanSerial // Current span
	dx   int        // X offset for coordinate transformation
}

// SpanSerial represents a span read from serialized data.
// This corresponds to the span struct in AGG's const_iterator.
type SpanSerial struct {
	X   basics.Int32 // Starting X coordinate
	Len basics.Int32 // Length of the span
}

// NewEmbeddedScanlineSerial creates a new embedded scanline for serialized data.
func NewEmbeddedScanlineSerial() *EmbeddedScanlineSerial {
	return &EmbeddedScanlineSerial{
		ptr:      nil,
		y:        0,
		numSpans: 0,
		dx:       0,
	}
}

// Reset is provided for interface compatibility but does nothing.
func (sl *EmbeddedScanlineSerial) Reset(_, _ int) {}

// NumSpans returns the number of spans in the current scanline.
func (sl *EmbeddedScanlineSerial) NumSpans() int {
	return sl.numSpans
}

// Y returns the Y coordinate of the current scanline.
func (sl *EmbeddedScanlineSerial) Y() int {
	return sl.y
}

// Begin returns an iterator for the spans in this scanline.
func (sl *EmbeddedScanlineSerial) Begin() *EmbeddedScanlineSerialIterator {
	it := &EmbeddedScanlineSerialIterator{
		ptr: sl.ptr,
		dx:  sl.dx,
	}

	// Read first span
	it.span.X = basics.Int32(it.readInt32()) + basics.Int32(it.dx)
	it.span.Len = basics.Int32(it.readInt32())

	return it
}

// Init initializes the embedded scanline from serialized data at the current position.
func (sl *EmbeddedScanlineSerial) Init(ptr []byte, dx, dy int) {
	sl.ptr = ptr
	sl.y = sl.readInt32() + dy
	sl.numSpans = sl.readInt32()
	sl.dx = dx
}

// readInt32 reads a 32-bit integer from the current position and advances the pointer.
func (sl *EmbeddedScanlineSerial) readInt32() int {
	if len(sl.ptr) < 4 {
		return 0
	}
	val := int(binary.LittleEndian.Uint32(sl.ptr[:4]))
	sl.ptr = sl.ptr[4:]
	return val
}

// Span returns the current span data.
func (it *EmbeddedScanlineSerialIterator) Span() SpanSerial {
	return it.span
}

// Next advances to the next span.
func (it *EmbeddedScanlineSerialIterator) Next() {
	it.span.X = basics.Int32(it.readInt32()) + basics.Int32(it.dx)
	it.span.Len = basics.Int32(it.readInt32())
}

// readInt32 reads a 32-bit integer from the current position and advances the pointer.
func (it *EmbeddedScanlineSerialIterator) readInt32() int {
	if len(it.ptr) < 4 {
		return 0
	}
	val := int(binary.LittleEndian.Uint32(it.ptr[:4]))
	it.ptr = it.ptr[4:]
	return val
}

// SerializedScanlinesAdaptorBin provides access to serialized binary scanline data.
// It can read scanline data that was serialized by ScanlineStorageBin.
// This corresponds to AGG's serialized_scanlines_adaptor_bin class.
type SerializedScanlinesAdaptorBin struct {
	data []byte // Serialized data buffer
	end  []byte // End of data buffer (for bounds checking)
	ptr  []byte // Current read position
	dx   int    // X coordinate offset
	dy   int    // Y coordinate offset
	minX int    // Minimum X coordinate
	minY int    // Minimum Y coordinate
	maxX int    // Maximum X coordinate
	maxY int    // Maximum Y coordinate
}

// NewSerializedScanlinesAdaptorBin creates a new adaptor for reading serialized data.
func NewSerializedScanlinesAdaptorBin() *SerializedScanlinesAdaptorBin {
	return &SerializedScanlinesAdaptorBin{
		minX: math.MaxInt32,
		minY: math.MaxInt32,
		maxX: math.MinInt32,
		maxY: math.MinInt32,
	}
}

// NewSerializedScanlinesAdaptorBinWithData creates a new adaptor with data and offsets.
func NewSerializedScanlinesAdaptorBinWithData(data []byte, dx, dy float64) *SerializedScanlinesAdaptorBin {
	adaptor := &SerializedScanlinesAdaptorBin{
		data: data,
		end:  data[len(data):], // Points to end of slice
		ptr:  data,
		dx:   int(dx + 0.5), // Round to nearest integer
		dy:   int(dy + 0.5), // Round to nearest integer
		minX: math.MaxInt32,
		minY: math.MaxInt32,
		maxX: math.MinInt32,
		maxY: math.MinInt32,
	}
	return adaptor
}

// Init initializes the adaptor with new data and offsets.
func (a *SerializedScanlinesAdaptorBin) Init(data []byte, dx, dy float64) {
	a.data = data
	a.end = data[len(data):]
	a.ptr = data
	a.dx = int(dx + 0.5) // Round to nearest integer
	a.dy = int(dy + 0.5) // Round to nearest integer
	a.minX = math.MaxInt32
	a.minY = math.MaxInt32
	a.maxX = math.MinInt32
	a.maxY = math.MinInt32
}

// readInt32 reads a 32-bit integer from the current position and advances the pointer.
func (a *SerializedScanlinesAdaptorBin) readInt32() int {
	if len(a.ptr) < 4 {
		return 0
	}
	val := int(binary.LittleEndian.Uint32(a.ptr[:4]))
	a.ptr = a.ptr[4:]
	return val
}

// RewindScanlines resets the iterator to the beginning and reads bounds information.
// This corresponds to AGG's rewind_scanlines() method.
func (a *SerializedScanlinesAdaptorBin) RewindScanlines() bool {
	a.ptr = a.data
	if len(a.data) == 0 {
		return false
	}

	if len(a.ptr) >= 16 { // Need at least 4 int32s for bounds
		a.minX = a.readInt32() + a.dx
		a.minY = a.readInt32() + a.dy
		a.maxX = a.readInt32() + a.dx
		a.maxY = a.readInt32() + a.dy
	}

	return len(a.data) > 0
}

// MinX returns the minimum X coordinate.
func (a *SerializedScanlinesAdaptorBin) MinX() int {
	return a.minX
}

// MinY returns the minimum Y coordinate.
func (a *SerializedScanlinesAdaptorBin) MinY() int {
	return a.minY
}

// MaxX returns the maximum X coordinate.
func (a *SerializedScanlinesAdaptorBin) MaxX() int {
	return a.maxX
}

// MaxY returns the maximum Y coordinate.
func (a *SerializedScanlinesAdaptorBin) MaxY() int {
	return a.maxY
}

// SweepScanline reads the next scanline from serialized data into the provided scanline.
// This corresponds to AGG's template sweep_scanline() method.
func (a *SerializedScanlinesAdaptorBin) SweepScanline(sl ScanlineInterface) bool {
	sl.ResetSpans()

	for {
		if len(a.ptr) < 8 { // Need at least 2 int32s for Y and num_spans
			return false
		}

		y := a.readInt32() + a.dy
		numSpans := a.readInt32()

		if len(a.ptr) < numSpans*8 { // Need 2 int32s per span
			return false
		}

		for i := 0; i < numSpans; i++ {
			x := a.readInt32() + a.dx
			length := a.readInt32()

			if length < 0 {
				length = -length
			}
			sl.AddSpan(x, length, basics.CoverFull)
		}

		if sl.NumSpans() > 0 {
			sl.Finalize(y)
			break
		}
	}

	return true
}

// SweepEmbeddedScanline reads the next scanline into an embedded scanline.
// This is a specialization for embedded_scanline.
func (a *SerializedScanlinesAdaptorBin) SweepEmbeddedScanline(sl *EmbeddedScanlineSerial) bool {
	for {
		if len(a.ptr) < 8 { // Need at least 2 int32s
			return false
		}

		// Initialize the embedded scanline at current position
		sl.Init(a.ptr, a.dx, a.dy)

		// Skip past this scanline's data
		y := a.readInt32()        // Y coordinate
		numSpans := a.readInt32() // Number of spans

		_ = y // Suppress unused variable warning

		if len(a.ptr) < numSpans*8 { // Check we have enough data for all spans
			return false
		}

		// Skip span data
		a.ptr = a.ptr[numSpans*8:]

		if sl.NumSpans() > 0 {
			break
		}
	}

	return true
}
