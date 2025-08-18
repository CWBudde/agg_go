// Package buffer provides rendering buffer implementations for AGG.
// This package handles the low-level memory management for pixel data.
package buffer

import (
	"unsafe"

	"agg_go/internal/basics"
)

// RenderingBuffer provides access to a pixel buffer with configurable stride.
// This is equivalent to AGG's row_accessor template class.
// Stride represents bytes per row, not elements per row.
type RenderingBuffer[T any] struct {
	buf    []T
	start  []T
	width  int
	height int
	stride int // Number of bytes per row (can be negative for bottom-up)
}

// NewRenderingBuffer creates a new rendering buffer.
func NewRenderingBuffer[T any]() *RenderingBuffer[T] {
	return &RenderingBuffer[T]{}
}

// NewRenderingBufferWithData creates a rendering buffer with existing data.
func NewRenderingBufferWithData[T any](buf []T, width, height, stride int) *RenderingBuffer[T] {
	rb := &RenderingBuffer[T]{}
	rb.Attach(buf, width, height, stride)
	return rb
}

// Attach attaches a buffer to the rendering buffer.
// stride is in bytes per row and can be negative for bottom-up organization.
func (rb *RenderingBuffer[T]) Attach(buf []T, width, height, stride int) {
	rb.buf = buf
	rb.width = width
	rb.height = height
	rb.stride = stride

	if stride < 0 {
		// Negative stride means bottom-up organization
		// Calculate offset in elements: (-stride * (height-1)) / sizeof(T)
		elementSize := int(unsafe.Sizeof(*new(T)))
		byteOffset := (-stride) * (height - 1)
		elementOffset := byteOffset / elementSize
		if len(buf) > elementOffset {
			rb.start = buf[elementOffset:]
		} else {
			rb.start = buf
		}
	} else {
		rb.start = buf
	}
}

// Buf returns the raw buffer data.
func (rb *RenderingBuffer[T]) Buf() []T {
	return rb.buf
}

// Width returns the buffer width in pixels.
func (rb *RenderingBuffer[T]) Width() int {
	return rb.width
}

// Height returns the buffer height in pixels.
func (rb *RenderingBuffer[T]) Height() int {
	return rb.height
}

// Stride returns the buffer stride (bytes per row).
func (rb *RenderingBuffer[T]) Stride() int {
	return rb.stride
}

// StrideAbs returns the absolute value of the stride.
func (rb *RenderingBuffer[T]) StrideAbs() int {
	if rb.stride < 0 {
		return -rb.stride
	}
	return rb.stride
}

// RowPtr returns a slice to the beginning of the specified row with length checking.
// x is the starting x coordinate, y is the row, length is the number of elements needed.
func (rb *RenderingBuffer[T]) RowPtr(x, y int, length int) []T {
	if y < 0 || y >= rb.height {
		return nil
	}

	// Convert byte stride to element stride
	elementSize := int(unsafe.Sizeof(*new(T)))
	elementStride := rb.stride / elementSize

	// Calculate row start in elements
	rowOffset := y * elementStride
	if rowOffset < 0 || rowOffset >= len(rb.start) {
		return nil
	}

	// Add x offset
	rowStart := rowOffset + x
	if x < 0 || rowStart < 0 || rowStart >= len(rb.start) {
		return nil
	}

	// Calculate end position
	end := rowStart + length
	if end > len(rb.start) {
		end = len(rb.start)
	}

	return rb.start[rowStart:end]
}

// Row returns a slice for the entire row.
func (rb *RenderingBuffer[T]) Row(y int) []T {
	if y < 0 || y >= rb.height {
		return nil
	}

	// Convert byte stride to element stride
	elementSize := int(unsafe.Sizeof(*new(T)))
	elementStride := rb.stride / elementSize

	// Calculate row start in elements
	rowOffset := y * elementStride
	if rowOffset < 0 || rowOffset >= len(rb.start) {
		return nil
	}

	// Calculate the absolute element stride for the row length
	absElementStride := elementStride
	if absElementStride < 0 {
		absElementStride = -absElementStride
	}

	// End is row start + absolute stride (in elements)
	end := rowOffset + absElementStride
	if end > len(rb.start) {
		end = len(rb.start)
	}

	return rb.start[rowOffset:end]
}

// RowData returns row information for the specified row.
// This is equivalent to AGG's row_data with x1=0, x2=width-1.
func (rb *RenderingBuffer[T]) RowData(y int) basics.ConstRowInfo[T] {
	return basics.ConstRowInfo[T]{
		X1:  0,
		X2:  rb.width - 1,
		Ptr: rb.Row(y),
	}
}

// CopyFrom copies data from another rendering buffer.
func (rb *RenderingBuffer[T]) CopyFrom(src *RenderingBuffer[T]) {
	if src == nil {
		return
	}

	minHeight := basics.IMin(rb.height, src.height)
	minWidth := basics.IMin(rb.width, src.width)

	for y := 0; y < minHeight; y++ {
		srcRow := src.Row(y)
		dstRow := rb.Row(y)

		if srcRow == nil || dstRow == nil {
			continue
		}

		copyLen := basics.IMin(len(srcRow), len(dstRow))
		copyLen = basics.IMin(copyLen, minWidth)

		copy(dstRow[:copyLen], srcRow[:copyLen])
	}
}

// Clear fills the buffer with the specified value.
// This matches AGG's clear(T value) method.
func (rb *RenderingBuffer[T]) Clear(value T) {
	// Clear stride_abs() elements per row, not the entire buffer
	elementSize := int(unsafe.Sizeof(*new(T)))
	elementStride := rb.StrideAbs() / elementSize

	for y := 0; y < rb.height; y++ {
		row := rb.Row(y)
		if row == nil {
			continue
		}
		// Fill up to elementStride or row length, whichever is smaller
		fillLen := elementStride
		if fillLen > len(row) {
			fillLen = len(row)
		}
		for x := 0; x < fillLen; x++ {
			row[x] = value
		}
	}
}

// ClearZero fills the buffer with the zero value for convenience.
func (rb *RenderingBuffer[T]) ClearZero() {
	var zero T
	rb.Clear(zero)
}

// RenderingBufferCache provides cached row pointers for faster access.
// This is equivalent to AGG's row_ptr_cache template class.
type RenderingBufferCache[T any] struct {
	RenderingBuffer[T]
	rows [][]T
}

// NewRenderingBufferCache creates a new rendering buffer with row caching.
func NewRenderingBufferCache[T any]() *RenderingBufferCache[T] {
	return &RenderingBufferCache[T]{}
}

// Attach attaches a buffer and builds the row cache.
func (rbc *RenderingBufferCache[T]) Attach(buf []T, width, height, stride int) {
	rbc.RenderingBuffer.Attach(buf, width, height, stride)
	rbc.buildRowCache()
}

// buildRowCache builds the cache of row pointers.
func (rbc *RenderingBufferCache[T]) buildRowCache() {
	rbc.rows = make([][]T, rbc.height)
	for y := 0; y < rbc.height; y++ {
		rbc.rows[y] = rbc.RenderingBuffer.Row(y)
	}
}

// Row returns a cached row slice.
func (rbc *RenderingBufferCache[T]) Row(y int) []T {
	if y < 0 || y >= len(rbc.rows) {
		return nil
	}
	return rbc.rows[y]
}

// RowData returns cached row information for the specified row.
func (rbc *RenderingBufferCache[T]) RowData(y int) basics.ConstRowInfo[T] {
	return basics.ConstRowInfo[T]{
		X1:  0,
		X2:  rbc.width - 1,
		Ptr: rbc.Row(y),
	}
}

// Rows returns all cached row pointers.
func (rbc *RenderingBufferCache[T]) Rows() [][]T {
	return rbc.rows
}

// Concrete rendering buffer type for uint8 data (common case)
type RenderingBufferU8 = RenderingBuffer[basics.Int8u]

// NewRenderingBufferU8 creates a new uint8 rendering buffer
func NewRenderingBufferU8() *RenderingBufferU8 {
	return NewRenderingBuffer[basics.Int8u]()
}

// NewRenderingBufferU8WithData creates a uint8 rendering buffer with existing data
func NewRenderingBufferU8WithData(buf []basics.Int8u, width, height, stride int) *RenderingBufferU8 {
	return NewRenderingBufferWithData(buf, width, height, stride)
}

// RowU8 returns a uint8 row slice - convenience function for pixel formats
func RowU8(rb *RenderingBufferU8, y int) []basics.Int8u {
	return rb.Row(y)
}

// Concrete rendering buffer type for uint16 data
type RenderingBufferU16 = RenderingBuffer[basics.Int16u]

// NewRenderingBufferU16 creates a new uint16 rendering buffer
func NewRenderingBufferU16() *RenderingBufferU16 {
	return NewRenderingBuffer[basics.Int16u]()
}

// NewRenderingBufferU16WithData creates a uint16 rendering buffer with existing data
func NewRenderingBufferU16WithData(buf []basics.Int16u, width, height, stride int) *RenderingBufferU16 {
	return NewRenderingBufferWithData(buf, width, height, stride)
}

// RowU16 returns a uint16 row slice - convenience function for pixel formats
func RowU16(rb *RenderingBufferU16, y int) []basics.Int16u {
	return rb.Row(y)
}

// Concrete rendering buffer type for uint32 data
type RenderingBufferU32 = RenderingBuffer[basics.Int32u]

// NewRenderingBufferU32 creates a new uint32 rendering buffer
func NewRenderingBufferU32() *RenderingBufferU32 {
	return NewRenderingBuffer[basics.Int32u]()
}

// NewRenderingBufferU32WithData creates a uint32 rendering buffer with existing data
func NewRenderingBufferU32WithData(buf []basics.Int32u, width, height, stride int) *RenderingBufferU32 {
	return NewRenderingBufferWithData(buf, width, height, stride)
}

// RowU32 returns a uint32 row slice - convenience function for pixel formats
func RowU32(rb *RenderingBufferU32, y int) []basics.Int32u {
	return rb.Row(y)
}

// Concrete rendering buffer type for float32 data
type RenderingBufferF32 = RenderingBuffer[float32]

// NewRenderingBufferF32 creates a new float32 rendering buffer
func NewRenderingBufferF32() *RenderingBufferF32 {
	return NewRenderingBuffer[float32]()
}

// NewRenderingBufferF32WithData creates a float32 rendering buffer with existing data
func NewRenderingBufferF32WithData(buf []float32, width, height, stride int) *RenderingBufferF32 {
	return NewRenderingBufferWithData(buf, width, height, stride)
}

// RowF32 returns a float32 row slice - convenience function for pixel formats
func RowF32(rb *RenderingBufferF32, y int) []float32 {
	return rb.Row(y)
}
