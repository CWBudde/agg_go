// Package buffer provides rendering buffer implementations for AGG.
// This file implements row pointer caching for efficient pattern access.
package buffer

// RowPtrCache provides cached row access for efficient pattern rendering.
// This is equivalent to AGG's row_ptr_cache template class.
type RowPtrCache[T any] struct {
	buf    []T
	rows   [][]T
	width  int
	height int
	stride int
}

// NewRowPtrCache creates a new row pointer cache.
func NewRowPtrCache[T any]() *RowPtrCache[T] {
	return &RowPtrCache[T]{}
}

// Attach attaches a buffer and builds the row cache.
// stride may be negative for bottom-up buffers (matching C++ AGG row_ptr_cache).
func (rpc *RowPtrCache[T]) Attach(buf []T, width, height, stride int) {
	rpc.buf = buf
	rpc.width = width
	rpc.height = height
	rpc.stride = stride
	rpc.rows = make([][]T, height)

	// C++ AGG: row_ptr starts at buf (positive stride) or
	// buf - (height-1)*stride (negative stride), then advances by stride.
	absStride := stride
	if absStride < 0 {
		absStride = -absStride
	}

	for y := range height {
		var rowOffset int
		if stride >= 0 {
			rowOffset = y * stride
		} else {
			rowOffset = (height - 1 - y) * absStride
		}
		if rowOffset >= 0 && rowOffset < len(buf) {
			end := rowOffset + width
			if end > len(buf) {
				end = len(buf)
			}
			rpc.rows[y] = buf[rowOffset:end]
		}
	}
}

// RowPtr returns a pointer to the specified row.
func (rpc *RowPtrCache[T]) RowPtr(y int) []T {
	if y < 0 || y >= rpc.height {
		return nil
	}
	return rpc.rows[y]
}

// Rows returns all row pointers for use with filters.
func (rpc *RowPtrCache[T]) Rows() [][]T {
	return rpc.rows
}

// Width returns the cache width.
func (rpc *RowPtrCache[T]) Width() int {
	return rpc.width
}

// Height returns the cache height.
func (rpc *RowPtrCache[T]) Height() int {
	return rpc.height
}

// Stride returns the cache stride.
func (rpc *RowPtrCache[T]) Stride() int {
	return rpc.stride
}
