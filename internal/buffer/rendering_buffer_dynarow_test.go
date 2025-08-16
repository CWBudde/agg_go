package buffer

import (
	"testing"
)

// Test basic dynamic row buffer functionality
func TestRenderingBufferDynarowBasic(t *testing.T) {
	width, height, byteWidth := 10, 5, 40 // 40 bytes per row

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Test accessors
	if rb.Width() != width {
		t.Errorf("Width() expected %d, got %d", width, rb.Width())
	}
	if rb.Height() != height {
		t.Errorf("Height() expected %d, got %d", height, rb.Height())
	}
	if rb.ByteWidth() != byteWidth {
		t.Errorf("ByteWidth() expected %d, got %d", byteWidth, rb.ByteWidth())
	}

	// Initially no rows should be allocated
	for y := 0; y < height; y++ {
		if rb.IsRowAllocated(y) {
			t.Errorf("Row %d should not be allocated initially", y)
		}
	}
}

// Test dynamic row allocation
func TestRenderingBufferDynarowAllocation(t *testing.T) {
	width, height, byteWidth := 8, 4, 32

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Access a row - should trigger allocation
	rowPtr := rb.RowPtr(2, 1, 5) // x=2, y=1, length=5
	if rowPtr == nil {
		t.Error("RowPtr should not return nil for valid parameters")
	}
	if len(rowPtr) != 5 {
		t.Errorf("RowPtr length expected 5, got %d", len(rowPtr))
	}

	// Row should now be allocated
	if !rb.IsRowAllocated(1) {
		t.Error("Row 1 should be allocated after RowPtr access")
	}

	// Check bounds
	x1, x2 := rb.GetAllocatedBounds(1)
	if x1 != 2 || x2 != 6 { // x=2, length=5, so x2 = 2+5-1 = 6
		t.Errorf("Allocated bounds expected (2, 6), got (%d, %d)", x1, x2)
	}

	// Other rows should still not be allocated
	for y := 0; y < height; y++ {
		if y != 1 && rb.IsRowAllocated(y) {
			t.Errorf("Row %d should not be allocated", y)
		}
	}
}

// Test bounds expansion
func TestRenderingBufferDynarowBoundsExpansion(t *testing.T) {
	width, height, byteWidth := 10, 3, 30

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// First access
	rb.RowPtr(3, 0, 4) // x=3, y=0, length=4 -> x1=3, x2=6
	x1, x2 := rb.GetAllocatedBounds(0)
	if x1 != 3 || x2 != 6 {
		t.Errorf("First access bounds expected (3, 6), got (%d, %d)", x1, x2)
	}

	// Second access - expand to the left
	rb.RowPtr(1, 0, 3) // x=1, y=0, length=3 -> x1=1, x2=6 (x1 expands)
	x1, x2 = rb.GetAllocatedBounds(0)
	if x1 != 1 || x2 != 6 {
		t.Errorf("Left expansion bounds expected (1, 6), got (%d, %d)", x1, x2)
	}

	// Third access - expand to the right
	rb.RowPtr(5, 0, 3) // x=5, y=0, length=3 -> x1=1, x2=7 (x2 expands)
	x1, x2 = rb.GetAllocatedBounds(0)
	if x1 != 1 || x2 != 7 {
		t.Errorf("Right expansion bounds expected (1, 7), got (%d, %d)", x1, x2)
	}
}

// Test RowData functionality
func TestRenderingBufferDynarowRowData(t *testing.T) {
	width, height, byteWidth := 6, 4, 24

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Unallocated row should return invalid bounds
	rowData := rb.RowData(0)
	if rowData.X1 != -1 || rowData.X2 != -1 || rowData.Ptr != nil {
		t.Errorf("Unallocated row data expected (-1, -1, nil), got (%d, %d, %v)",
			rowData.X1, rowData.X2, rowData.Ptr != nil)
	}

	// Allocate a row
	rb.RowPtr(2, 0, 3)

	// Now RowData should return valid information
	rowData = rb.RowData(0)
	if rowData.X1 != 2 || rowData.X2 != 4 {
		t.Errorf("Allocated row data bounds expected (2, 4), got (%d, %d)", rowData.X1, rowData.X2)
	}
	if rowData.Ptr == nil {
		t.Error("Allocated row data Ptr should not be nil")
	}
	if len(rowData.Ptr) != byteWidth {
		t.Errorf("Allocated row data Ptr length expected %d, got %d", byteWidth, len(rowData.Ptr))
	}
}

// Test row pointer methods
func TestRenderingBufferDynarowRowPointers(t *testing.T) {
	width, height, byteWidth := 5, 3, 20

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Test const access to unallocated row
	constPtr := rb.RowPtrConst(0)
	if constPtr != nil {
		t.Error("RowPtrConst should return nil for unallocated row")
	}

	// Test mutable access - should allocate
	mutPtr := rb.RowPtrMutable(0)
	if mutPtr == nil {
		t.Error("RowPtrMutable should not return nil")
	}
	if len(mutPtr) != byteWidth {
		t.Errorf("RowPtrMutable length expected %d, got %d", byteWidth, len(mutPtr))
	}

	// Row should now be allocated
	if !rb.IsRowAllocated(0) {
		t.Error("Row 0 should be allocated after RowPtrMutable")
	}

	// Const access should now work
	constPtr = rb.RowPtrConst(0)
	if constPtr == nil {
		t.Error("RowPtrConst should not return nil for allocated row")
	}

	// Should be the same underlying data
	if len(constPtr) != len(mutPtr) {
		t.Error("Const and mutable pointers should have same length")
	}
}

// Test bounds checking
func TestRenderingBufferDynarowBounds(t *testing.T) {
	width, height, byteWidth := 4, 2, 16

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Test out-of-bounds access
	if rowPtr := rb.RowPtr(0, -1, 1); rowPtr != nil {
		t.Error("RowPtr with negative y should return nil")
	}
	if rowPtr := rb.RowPtr(0, height, 1); rowPtr != nil {
		t.Error("RowPtr with y >= height should return nil")
	}

	// Test bounds checking for unallocated rows
	x1, x2 := rb.GetAllocatedBounds(-1)
	if x1 != -1 || x2 != -1 {
		t.Errorf("GetAllocatedBounds for invalid row should return (-1, -1), got (%d, %d)", x1, x2)
	}

	x1, x2 = rb.GetAllocatedBounds(height)
	if x1 != -1 || x2 != -1 {
		t.Errorf("GetAllocatedBounds for invalid row should return (-1, -1), got (%d, %d)", x1, x2)
	}
}

// Test Init functionality
func TestRenderingBufferDynarowInit(t *testing.T) {
	rb := NewRenderingBufferDynarow()

	// Initial state
	if rb.Width() != 0 || rb.Height() != 0 || rb.ByteWidth() != 0 {
		t.Error("New buffer should have zero dimensions")
	}

	// Initialize with dimensions
	width, height, byteWidth := 6, 3, 24
	rb.Init(width, height, byteWidth)

	if rb.Width() != width || rb.Height() != height || rb.ByteWidth() != byteWidth {
		t.Errorf("After Init, dimensions expected (%d, %d, %d), got (%d, %d, %d)",
			width, height, byteWidth, rb.Width(), rb.Height(), rb.ByteWidth())
	}

	// Allocate some rows
	rb.RowPtr(1, 0, 2)
	rb.RowPtr(2, 1, 3)

	if !rb.IsRowAllocated(0) || !rb.IsRowAllocated(1) {
		t.Error("Rows should be allocated")
	}

	// Re-initialize - should clear all allocations
	rb.Init(4, 2, 16)

	if rb.Width() != 4 || rb.Height() != 2 || rb.ByteWidth() != 16 {
		t.Error("Re-initialization failed")
	}

	// All rows should be deallocated
	for y := 0; y < rb.Height(); y++ {
		if rb.IsRowAllocated(y) {
			t.Errorf("Row %d should not be allocated after re-initialization", y)
		}
	}

	// Test clearing (Init with zero dimensions)
	rb.Init(0, 0, 0)
	if rb.Width() != 0 || rb.Height() != 0 || rb.ByteWidth() != 0 {
		t.Error("Init with zero dimensions should clear buffer")
	}
}

// Test Clear functionality
func TestRenderingBufferDynarowClear(t *testing.T) {
	width, height, byteWidth := 5, 3, 20

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Allocate some rows
	rb.RowPtr(1, 0, 3)
	rb.RowPtr(2, 1, 2)
	rb.RowPtr(0, 2, 4)

	// Verify rows are allocated
	for y := 0; y < height; y++ {
		if !rb.IsRowAllocated(y) {
			t.Errorf("Row %d should be allocated", y)
		}
	}

	// Clear all rows
	rb.Clear()

	// Verify all rows are deallocated
	for y := 0; y < height; y++ {
		if rb.IsRowAllocated(y) {
			t.Errorf("Row %d should not be allocated after Clear()", y)
		}
	}

	// Bounds should be invalid
	for y := 0; y < height; y++ {
		x1, x2 := rb.GetAllocatedBounds(y)
		if x1 != -1 || x2 != -1 {
			t.Errorf("Row %d bounds should be (-1, -1) after Clear(), got (%d, %d)", y, x1, x2)
		}
	}
}

// Test large allocations and edge cases
func TestRenderingBufferDynarowEdgeCases(t *testing.T) {
	width, height, byteWidth := 100, 50, 400

	rb := NewRenderingBufferDynarowWithSize(width, height, byteWidth)

	// Test large allocation
	rowPtr := rb.RowPtr(10, 25, 80)
	if rowPtr == nil {
		t.Error("Large allocation should succeed")
	}
	if len(rowPtr) != 80 {
		t.Errorf("Large allocation length expected 80, got %d", len(rowPtr))
	}

	// Test allocation beyond byte width
	rowPtr = rb.RowPtr(350, 30, 100) // x=350 is beyond byteWidth=400
	if rowPtr == nil {
		t.Error("Allocation beyond byte width should still succeed")
	}
	// Length should be truncated
	if len(rowPtr) > 50 { // 400 - 350 = 50
		t.Errorf("Allocation beyond byte width should be truncated, got length %d", len(rowPtr))
	}

	// Test zero length allocation
	rowPtr = rb.RowPtr(0, 0, 0)
	if rowPtr == nil {
		t.Error("Zero length allocation should succeed but return empty slice")
	}
	if len(rowPtr) != 0 {
		t.Errorf("Zero length allocation should return empty slice, got length %d", len(rowPtr))
	}
}
