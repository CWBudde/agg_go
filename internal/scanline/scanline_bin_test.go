package scanline

import (
	"testing"
)

// Tests for ScanlineBin (16-bit coordinates)

func TestNewScanlineBin(t *testing.T) {
	sl := NewScanlineBin()
	if sl == nil {
		t.Fatal("NewScanlineBin() returned nil")
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be 0x7FFFFFF0, got %d", sl.lastX)
	}
}

func TestScanlineBinReset(t *testing.T) {
	sl := NewScanlineBin()

	minX, maxX := 10, 100
	sl.Reset(minX, maxX)

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be reset to 0x7FFFFFF0, got %d", sl.lastX)
	}

	if sl.curSpan != 0 {
		t.Errorf("Expected curSpan to be 0, got %d", sl.curSpan)
	}

	// Verify array is properly sized
	expectedSize := maxX - minX + 3
	if sl.spans.Size() < expectedSize {
		t.Errorf("Expected spans array size >= %d, got %d", expectedSize, sl.spans.Size())
	}
}

func TestScanlineBinAddCell(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add a single cell - coverage value should be ignored
	sl.AddCell(10, 999) // Coverage value 999 should be ignored

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span in iterator, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Expected span X to be 10, got %d", span.X)
	}
	if span.Len != 1 {
		t.Errorf("Expected span Len to be 1, got %d", span.Len)
	}
}

func TestScanlineBinConsecutiveCells(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add consecutive cells - should merge into one span
	sl.AddCell(10, 100)
	sl.AddCell(11, 200) // Different coverage values should be ignored
	sl.AddCell(12, 50)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span (merged), got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	span := spans[0]
	if span.X != 10 {
		t.Errorf("Expected span X to be 10, got %d", span.X)
	}
	if span.Len != 3 {
		t.Errorf("Expected span Len to be 3, got %d", span.Len)
	}
}

func TestScanlineBinNonConsecutiveCells(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add non-consecutive cells - should create separate spans
	sl.AddCell(10, 100)
	sl.AddCell(15, 150) // Gap of 4 pixels

	if sl.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	if len(spans) != 2 {
		t.Fatalf("Expected 2 spans in iterator, got %d", len(spans))
	}

	// First span
	if spans[0].X != 10 || spans[0].Len != 1 {
		t.Errorf("First span: expected (X:10, Len:1), got (X:%d, Len:%d)",
			spans[0].X, spans[0].Len)
	}

	// Second span
	if spans[1].X != 15 || spans[1].Len != 1 {
		t.Errorf("Second span: expected (X:15, Len:1), got (X:%d, Len:%d)",
			spans[1].X, spans[1].Len)
	}
}

func TestScanlineBinAddSpan(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add a span - coverage value should be ignored
	sl.AddSpan(30, 5, 255)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	span := spans[0]
	if span.X != 30 {
		t.Errorf("Expected span X to be 30, got %d", span.X)
	}
	if span.Len != 5 {
		t.Errorf("Expected span Len to be 5, got %d", span.Len)
	}
}

func TestScanlineBinAddCells(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// AddCells should work like AddSpan - covers array is ignored
	var covers []CoverType = nil
	sl.AddCells(20, 4, covers)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	span := spans[0]
	if span.X != 20 {
		t.Errorf("Expected span X to be 20, got %d", span.X)
	}
	if span.Len != 4 {
		t.Errorf("Expected span Len to be 4, got %d", span.Len)
	}
}

func TestScanlineBinMixedOperations(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Mix different addition methods
	sl.AddCell(10, 100)
	sl.AddCell(11, 110) // Should merge

	sl.AddSpan(20, 3, 200) // Separate span

	sl.AddCells(30, 2, nil) // Another separate span

	if sl.NumSpans() != 3 {
		t.Errorf("Expected 3 spans, got %d", sl.NumSpans())
	}

	spans := sl.Begin()

	// First span (merged cells)
	if spans[0].X != 10 || spans[0].Len != 2 {
		t.Errorf("First span: expected (X:10, Len:2), got (X:%d, Len:%d)",
			spans[0].X, spans[0].Len)
	}

	// Second span
	if spans[1].X != 20 || spans[1].Len != 3 {
		t.Errorf("Second span: expected (X:20, Len:3), got (X:%d, Len:%d)",
			spans[1].X, spans[1].Len)
	}

	// Third span
	if spans[2].X != 30 || spans[2].Len != 2 {
		t.Errorf("Third span: expected (X:30, Len:2), got (X:%d, Len:%d)",
			spans[2].X, spans[2].Len)
	}
}

func TestScanlineBinFinalize(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)
	sl.AddCell(10, 128)

	sl.Finalize(42)

	if sl.Y() != 42 {
		t.Errorf("Expected Y coordinate to be 42, got %d", sl.Y())
	}
}

func TestScanlineBinResetSpans(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)
	sl.AddCell(10, 128)
	sl.AddCell(11, 129)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span before reset, got %d", sl.NumSpans())
	}

	sl.ResetSpans()

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans after reset, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be reset to 0x7FFFFFF0, got %d", sl.lastX)
	}

	// Should be able to add new spans after reset
	sl.AddCell(20, 200)
	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span after adding new cell, got %d", sl.NumSpans())
	}
}

// Additional edge case tests for proper AGG compliance

func TestScanlineBinEmptyBegin(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Empty scanline should return nil from Begin()
	spans := sl.Begin()
	if spans != nil {
		t.Errorf("Expected nil from Begin() for empty scanline, got %v", spans)
	}
}

func TestScanlineBinAGGIndexConvention(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add a single cell
	sl.AddCell(10, 128)

	// Check that span is stored at index 1 (AGG convention)
	if sl.curSpan != 1 {
		t.Errorf("Expected curSpan to be 1 (AGG convention), got %d", sl.curSpan)
	}

	spans := sl.Begin()
	if len(spans) != 1 {
		t.Errorf("Expected 1 span from Begin(), got %d", len(spans))
	}

	if spans[0].X != 10 {
		t.Errorf("Expected span X to be 10, got %d", spans[0].X)
	}
}

func TestScanlineBinSpanIndexZeroUnused(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 100)

	// Add some spans
	sl.AddCell(5, 100)
	sl.AddCell(10, 100) // Non-consecutive, creates new span
	sl.AddCell(11, 100) // Consecutive, extends span

	if sl.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", sl.NumSpans())
	}

	// Verify AGG convention: index 0 should be unused
	// curSpan should be 2 (pointing to last span)
	if sl.curSpan != 2 {
		t.Errorf("Expected curSpan to be 2, got %d", sl.curSpan)
	}

	spans := sl.Begin()
	if len(spans) != 2 {
		t.Errorf("Expected 2 spans from Begin(), got %d", len(spans))
	}

	// First span
	if spans[0].X != 5 || spans[0].Len != 1 {
		t.Errorf("First span: expected (X:5, Len:1), got (X:%d, Len:%d)", spans[0].X, spans[0].Len)
	}

	// Second span (merged)
	if spans[1].X != 10 || spans[1].Len != 2 {
		t.Errorf("Second span: expected (X:10, Len:2), got (X:%d, Len:%d)", spans[1].X, spans[1].Len)
	}
}

func TestScanlineBinLargeSpanGrowth(t *testing.T) {
	sl := NewScanlineBin()
	sl.Reset(0, 1000)

	// Add many non-consecutive spans to test array growth
	for i := 0; i < 50; i++ {
		sl.AddCell(i*10, 128) // Each cell is 10 units apart
	}

	if sl.NumSpans() != 50 {
		t.Errorf("Expected 50 spans, got %d", sl.NumSpans())
	}

	spans := sl.Begin()
	if len(spans) != 50 {
		t.Errorf("Expected 50 spans from Begin(), got %d", len(spans))
	}

	// Check a few spans
	if spans[0].X != 0 {
		t.Errorf("First span X: expected 0, got %d", spans[0].X)
	}
	if spans[49].X != 490 {
		t.Errorf("Last span X: expected 490, got %d", spans[49].X)
	}
}

// Tests for Scanline32Bin (32-bit coordinates)

func TestNewScanline32Bin(t *testing.T) {
	sl := NewScanline32Bin()
	if sl == nil {
		t.Fatal("NewScanline32Bin() returned nil")
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be 0x7FFFFFF0, got %d", sl.lastX)
	}
}

func TestScanline32BinReset(t *testing.T) {
	sl := NewScanline32Bin()

	// Reset parameters are ignored for scanline32_bin
	sl.Reset(999, 9999)

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be reset to 0x7FFFFFF0, got %d", sl.lastX)
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans after reset, got %d", sl.NumSpans())
	}
}

func TestScanline32BinAddCell(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add a single cell - coverage value should be ignored
	sl.AddCell(10000, 999)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10000 {
		t.Errorf("Expected span X to be 10000, got %d", span.X)
	}
	if span.Len != 1 {
		t.Errorf("Expected span Len to be 1, got %d", span.Len)
	}
}

func TestScanline32BinConsecutiveCells(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add consecutive cells - should merge into one span
	sl.AddCell(50000, 100)
	sl.AddCell(50001, 200)
	sl.AddCell(50002, 50)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span (merged), got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	span := spans[0]
	if span.X != 50000 {
		t.Errorf("Expected span X to be 50000, got %d", span.X)
	}
	if span.Len != 3 {
		t.Errorf("Expected span Len to be 3, got %d", span.Len)
	}
}

func TestScanline32BinAddSpan(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add a span with large coordinates
	sl.AddSpan(100000, 500, 255)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	span := spans[0]
	if span.X != 100000 {
		t.Errorf("Expected span X to be 100000, got %d", span.X)
	}
	if span.Len != 500 {
		t.Errorf("Expected span Len to be 500, got %d", span.Len)
	}
}

func TestScanline32BinIterator(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add multiple spans
	sl.AddSpan(1000, 10, 0)
	sl.AddSpan(2000, 20, 0)
	sl.AddSpan(3000, 30, 0)

	if sl.NumSpans() != 3 {
		t.Errorf("Expected 3 spans, got %d", sl.NumSpans())
	}

	// Test iterator
	it := sl.Begin()

	// First span
	if !it.HasMore() {
		t.Fatal("Iterator should have first span")
	}
	span := it.Span()
	if span.X != 1000 || span.Len != 10 {
		t.Errorf("First span: expected (X:1000, Len:10), got (X:%d, Len:%d)",
			span.X, span.Len)
	}

	// Second span
	if !it.Next() {
		t.Fatal("Iterator should have second span")
	}
	span = it.Span()
	if span.X != 2000 || span.Len != 20 {
		t.Errorf("Second span: expected (X:2000, Len:20), got (X:%d, Len:%d)",
			span.X, span.Len)
	}

	// Third span
	if !it.Next() {
		t.Fatal("Iterator should have third span")
	}
	span = it.Span()
	if span.X != 3000 || span.Len != 30 {
		t.Errorf("Third span: expected (X:3000, Len:30), got (X:%d, Len:%d)",
			span.X, span.Len)
	}

	// No more spans
	if it.Next() {
		t.Error("Iterator should not have more spans")
	}
}

func TestScanline32BinMergeSpans(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add span, then consecutive cells that should merge
	sl.AddSpan(1000, 5, 0)
	sl.AddCell(1005, 0) // Should merge with previous span
	sl.AddCell(1006, 0) // Should continue merging

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 merged span, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if spans[0].X != 1000 || spans[0].Len != 7 {
		t.Errorf("Expected merged span (X:1000, Len:7), got (X:%d, Len:%d)",
			spans[0].X, spans[0].Len)
	}
}

func TestScanline32BinResetSpans(t *testing.T) {
	sl := NewScanline32Bin()
	sl.Reset(0, 0)

	// Add some spans
	sl.AddSpan(1000, 10, 0)
	sl.AddSpan(2000, 20, 0)

	if sl.NumSpans() != 2 {
		t.Errorf("Expected 2 spans before reset, got %d", sl.NumSpans())
	}

	sl.ResetSpans()

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans after reset, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be reset to 0x7FFFFFF0, got %d", sl.lastX)
	}
}
