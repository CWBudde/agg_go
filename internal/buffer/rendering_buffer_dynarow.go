// Package buffer provides rendering buffer implementations for AGG.
// This file implements the dynamic row allocation rendering buffer.
package buffer

import (
	"agg_go/internal/basics"
)

// RenderingBufferDynarow provides a rendering buffer with dynamic allocation of rows.
// The rows are allocated as needed when requesting for RowPtr().
// The class automatically calculates min_x and max_x for each row.
// Generally it's more efficient to use this class as a temporary buffer
// for rendering a few lines and then to blend it with another buffer.
// This is equivalent to AGG's rendering_buffer_dynarow class.
type RenderingBufferDynarow struct {
	rows      []basics.RowInfo[basics.Int8u] // Row information with dynamic allocation
	width     int                            // Width in pixels
	height    int                            // Height in pixels
	byteWidth int                            // Width in bytes per row
}

// NewRenderingBufferDynarow creates a new dynamic rendering buffer.
func NewRenderingBufferDynarow() *RenderingBufferDynarow {
	return &RenderingBufferDynarow{}
}

// NewRenderingBufferDynarowWithSize creates a new dynamic rendering buffer with specified dimensions.
func NewRenderingBufferDynarowWithSize(width, height, byteWidth int) *RenderingBufferDynarow {
	rb := &RenderingBufferDynarow{
		width:     width,
		height:    height,
		byteWidth: byteWidth,
		rows:      make([]basics.RowInfo[basics.Int8u], height),
	}
	// Initialize all rows as empty (nil pointers with invalid bounds)
	for i := range rb.rows {
		rb.rows[i] = basics.RowInfo[basics.Int8u]{
			X1:  -1,
			X2:  -1,
			Ptr: nil,
		}
	}
	return rb
}

// Init allocates and clears the buffer with specified dimensions.
// This deallocates any existing row data and reinitializes the buffer.
func (rb *RenderingBufferDynarow) Init(width, height, byteWidth int) {
	// Deallocate existing rows (Go's GC will handle the cleanup)
	// In C++ this would explicitly deallocate memory

	if width > 0 && height > 0 {
		rb.width = width
		rb.height = height
		rb.byteWidth = byteWidth
		rb.rows = make([]basics.RowInfo[basics.Int8u], height)

		// Initialize all rows as empty
		for i := range rb.rows {
			rb.rows[i] = basics.RowInfo[basics.Int8u]{
				X1:  -1,
				X2:  -1,
				Ptr: nil,
			}
		}
	} else {
		rb.width = 0
		rb.height = 0
		rb.byteWidth = 0
		rb.rows = nil
	}
}

// Width returns the buffer width in pixels.
func (rb *RenderingBufferDynarow) Width() int {
	return rb.width
}

// Height returns the buffer height in pixels.
func (rb *RenderingBufferDynarow) Height() int {
	return rb.height
}

// ByteWidth returns the buffer width in bytes.
func (rb *RenderingBufferDynarow) ByteWidth() int {
	return rb.byteWidth
}

// RowPtr returns a pointer to the beginning of the specified row with dynamic allocation.
// Memory for the row is allocated as needed and x1/x2 bounds are tracked automatically.
// This is the main function used for rendering.
func (rb *RenderingBufferDynarow) RowPtr(x, y int, length int) []basics.Int8u {
	if y < 0 || y >= rb.height {
		return nil
	}

	row := &rb.rows[y]
	x2 := x + length - 1

	if row.Ptr != nil {
		// Row already allocated, expand bounds if necessary
		if x < row.X1 {
			row.X1 = x
		}
		if x2 > row.X2 {
			row.X2 = x2
		}
	} else {
		// Allocate new row
		row.Ptr = make([]basics.Int8u, rb.byteWidth)
		row.X1 = x
		row.X2 = x2
		// Initialize to zero (Go slices are zero-initialized by default)
	}

	// Return slice starting from x position
	if x >= len(row.Ptr) {
		return nil
	}

	end := x + length
	if end > len(row.Ptr) {
		end = len(row.Ptr)
	}

	return row.Ptr[x:end]
}

// RowPtrConst returns a const pointer to the row (read-only access).
func (rb *RenderingBufferDynarow) RowPtrConst(y int) []basics.Int8u {
	if y < 0 || y >= rb.height {
		return nil
	}
	return rb.rows[y].Ptr
}

// RowPtrMutable returns a mutable pointer to the row, allocating if necessary.
// Returns the entire row buffer (byteWidth elements).
func (rb *RenderingBufferDynarow) RowPtrMutable(y int) []basics.Int8u {
	// Allocate the row if needed, then return the full row
	rb.RowPtr(0, y, rb.width) // This ensures allocation
	if y < 0 || y >= rb.height {
		return nil
	}
	return rb.rows[y].Ptr
}

// RowData returns row information for the specified row.
func (rb *RenderingBufferDynarow) RowData(y int) basics.RowInfo[basics.Int8u] {
	if y < 0 || y >= rb.height {
		return basics.RowInfo[basics.Int8u]{X1: -1, X2: -1, Ptr: nil}
	}
	return rb.rows[y]
}

// IsRowAllocated returns true if the specified row has been allocated.
func (rb *RenderingBufferDynarow) IsRowAllocated(y int) bool {
	if y < 0 || y >= rb.height {
		return false
	}
	return rb.rows[y].Ptr != nil
}

// GetAllocatedBounds returns the actual allocated bounds for a row.
// Returns (-1, -1) if row is not allocated.
func (rb *RenderingBufferDynarow) GetAllocatedBounds(y int) (x1, x2 int) {
	if y < 0 || y >= rb.height || rb.rows[y].Ptr == nil {
		return -1, -1
	}
	return rb.rows[y].X1, rb.rows[y].X2
}

// Clear deallocates all rows and resets the buffer to empty state.
func (rb *RenderingBufferDynarow) Clear() {
	for i := range rb.rows {
		rb.rows[i] = basics.RowInfo[basics.Int8u]{
			X1:  -1,
			X2:  -1,
			Ptr: nil,
		}
	}
}
