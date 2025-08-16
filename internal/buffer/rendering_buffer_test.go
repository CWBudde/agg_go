package buffer

import (
	"testing"

	"agg_go/internal/basics"
)

// Test basic rendering buffer functionality
func TestRenderingBufferBasic(t *testing.T) {
	width, height := 10, 5
	stride := width * 4 // 4 bytes per pixel (RGBA)
	buf := make([]basics.Int8u, height*stride)

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	// Test accessors
	if rb.Width() != width {
		t.Errorf("Width() expected %d, got %d", width, rb.Width())
	}
	if rb.Height() != height {
		t.Errorf("Height() expected %d, got %d", height, rb.Height())
	}
	if rb.Stride() != stride {
		t.Errorf("Stride() expected %d, got %d", stride, rb.Stride())
	}
	if rb.StrideAbs() != stride {
		t.Errorf("StrideAbs() expected %d, got %d", stride, rb.StrideAbs())
	}
}

// Test negative stride (bottom-up buffer organization)
func TestRenderingBufferNegativeStride(t *testing.T) {
	width, height := 10, 5
	stride := -(width * 4) // Negative stride
	buf := make([]basics.Int8u, height*(-stride))

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	if rb.Stride() != stride {
		t.Errorf("Stride() expected %d, got %d", stride, rb.Stride())
	}
	if rb.StrideAbs() != -stride {
		t.Errorf("StrideAbs() expected %d, got %d", -stride, rb.StrideAbs())
	}

	// Should be able to access rows
	row := rb.Row(0)
	if row == nil {
		t.Error("Row(0) should not be nil for negative stride")
	}
}

// Test row access
func TestRenderingBufferRowAccess(t *testing.T) {
	width, height := 8, 4
	stride := width * 1 // 1 byte per pixel
	buf := make([]basics.Int8u, height*stride)

	// Fill buffer with test pattern
	for i := range buf {
		buf[i] = basics.Int8u(i % 256)
	}

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	// Test row access
	for y := 0; y < height; y++ {
		row := rb.Row(y)
		if row == nil {
			t.Errorf("Row(%d) should not be nil", y)
			continue
		}
		if len(row) != width {
			t.Errorf("Row(%d) length expected %d, got %d", y, width, len(row))
		}

		// Check first element of each row
		expectedFirstElement := basics.Int8u((y * stride) % 256)
		if row[0] != expectedFirstElement {
			t.Errorf("Row(%d)[0] expected %d, got %d", y, expectedFirstElement, row[0])
		}
	}
}

// Test bounds checking
func TestRenderingBufferBounds(t *testing.T) {
	width, height := 5, 3
	stride := width * 2
	buf := make([]basics.Int8u, height*stride)

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	// Test out-of-bounds access
	if row := rb.Row(-1); row != nil {
		t.Error("Row(-1) should return nil")
	}
	if row := rb.Row(height); row != nil {
		t.Error("Row(height) should return nil")
	}

	// Test RowPtr bounds
	if rowPtr := rb.RowPtr(-1, 0, 1); rowPtr != nil {
		t.Error("RowPtr with negative x should return nil")
	}
	if rowPtr := rb.RowPtr(0, -1, 1); rowPtr != nil {
		t.Error("RowPtr with negative y should return nil")
	}
	if rowPtr := rb.RowPtr(0, height, 1); rowPtr != nil {
		t.Error("RowPtr with y >= height should return nil")
	}
}

// Test RowData functionality
func TestRenderingBufferRowData(t *testing.T) {
	width, height := 6, 2
	stride := width * 1
	buf := make([]basics.Int8u, height*stride)

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	rowData := rb.RowData(0)
	if rowData.X1 != 0 {
		t.Errorf("RowData.X1 expected 0, got %d", rowData.X1)
	}
	if rowData.X2 != width-1 {
		t.Errorf("RowData.X2 expected %d, got %d", width-1, rowData.X2)
	}
	if rowData.Ptr == nil {
		t.Error("RowData.Ptr should not be nil")
	}
	if len(rowData.Ptr) != width {
		t.Errorf("RowData.Ptr length expected %d, got %d", width, len(rowData.Ptr))
	}
}

// Test Clear functionality
func TestRenderingBufferClear(t *testing.T) {
	width, height := 4, 3
	stride := width * 1
	buf := make([]basics.Int8u, height*stride)

	// Fill with non-zero values
	for i := range buf {
		buf[i] = 42
	}

	rb := NewRenderingBuffer[basics.Int8u]()
	rb.Attach(buf, width, height, stride)

	// Clear with specific value
	rb.Clear(123)

	// Check that rows contain the clear value
	for y := 0; y < height; y++ {
		row := rb.Row(y)
		for x := 0; x < width; x++ {
			if row[x] != 123 {
				t.Errorf("After Clear(123), row[%d][%d] expected 123, got %d", y, x, row[x])
			}
		}
	}

	// Test ClearZero
	rb.ClearZero()
	for y := 0; y < height; y++ {
		row := rb.Row(y)
		for x := 0; x < width; x++ {
			if row[x] != 0 {
				t.Errorf("After ClearZero(), row[%d][%d] expected 0, got %d", y, x, row[x])
			}
		}
	}
}

// Test CopyFrom functionality
func TestRenderingBufferCopyFrom(t *testing.T) {
	width, height := 3, 2
	stride := width * 1

	// Source buffer
	srcBuf := make([]basics.Int8u, height*stride)
	for i := range srcBuf {
		srcBuf[i] = basics.Int8u(i + 10)
	}
	srcRb := NewRenderingBufferU8WithData(srcBuf, width, height, stride)

	// Destination buffer
	dstBuf := make([]basics.Int8u, height*stride)
	dstRb := NewRenderingBufferU8WithData(dstBuf, width, height, stride)

	// Copy from source to destination
	dstRb.CopyFrom(srcRb)

	// Verify copy
	for y := 0; y < height; y++ {
		srcRow := srcRb.Row(y)
		dstRow := dstRb.Row(y)
		for x := 0; x < width; x++ {
			if dstRow[x] != srcRow[x] {
				t.Errorf("CopyFrom failed at [%d][%d]: expected %d, got %d", y, x, srcRow[x], dstRow[x])
			}
		}
	}
}

// Test rendering buffer cache
func TestRenderingBufferCache(t *testing.T) {
	width, height := 5, 4
	stride := width * 1
	buf := make([]basics.Int8u, height*stride)

	rbc := NewRenderingBufferCache[basics.Int8u]()
	rbc.Attach(buf, width, height, stride)

	// Test that cache was built
	rows := rbc.Rows()
	if len(rows) != height {
		t.Errorf("Cached rows length expected %d, got %d", height, len(rows))
	}

	// Test cached row access
	for y := 0; y < height; y++ {
		row := rbc.Row(y)
		if row == nil {
			t.Errorf("Cached Row(%d) should not be nil", y)
		}
		if len(row) != width {
			t.Errorf("Cached Row(%d) length expected %d, got %d", y, width, len(row))
		}
	}

	// Test cached RowData
	rowData := rbc.RowData(0)
	if rowData.X1 != 0 || rowData.X2 != width-1 {
		t.Errorf("Cached RowData bounds expected (0, %d), got (%d, %d)", width-1, rowData.X1, rowData.X2)
	}
}

// Test type aliases
func TestRenderingBufferU8(t *testing.T) {
	width, height := 3, 2
	stride := width * 1
	buf := make([]basics.Int8u, height*stride)

	rb := NewRenderingBufferU8WithData(buf, width, height, stride)

	if rb.Width() != width {
		t.Errorf("U8 buffer Width() expected %d, got %d", width, rb.Width())
	}

	// Test convenience function
	row := RowU8(rb, 0)
	if row == nil {
		t.Error("RowU8 should not return nil")
	}
	if len(row) != width {
		t.Errorf("RowU8 length expected %d, got %d", width, len(row))
	}
}
