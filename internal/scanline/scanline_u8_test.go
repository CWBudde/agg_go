package scanline

import (
	"testing"
)

func TestNewScanlineU8(t *testing.T) {
	sl := NewScanlineU8()
	if sl == nil {
		t.Fatal("NewScanlineU8() returned nil")
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be 0x7FFFFFF0, got %d", sl.lastX)
	}
}

func TestReset(t *testing.T) {
	sl := NewScanlineU8()

	// Test reset with a reasonable range
	minX, maxX := 10, 100
	sl.Reset(minX, maxX)

	if sl.minX != minX {
		t.Errorf("Expected minX to be %d, got %d", minX, sl.minX)
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be reset to 0x7FFFFFF0, got %d", sl.lastX)
	}

	if sl.curSpan != 0 {
		t.Errorf("Expected curSpan to be 0, got %d", sl.curSpan)
	}

	// Verify arrays are properly sized
	expectedSize := maxX - minX + 2
	if sl.covers.Size() < expectedSize {
		t.Errorf("Expected covers array size >= %d, got %d", expectedSize, sl.covers.Size())
	}
	if sl.spans.Size() < expectedSize {
		t.Errorf("Expected spans array size >= %d, got %d", expectedSize, sl.spans.Size())
	}
}

func TestAddCell(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Add a single cell
	sl.AddCell(10, 128)

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
	if span.Covers[0] != 128 {
		t.Errorf("Expected first cover value to be 128, got %d", span.Covers[0])
	}
}

func TestAddConsecutiveCells(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Add consecutive cells - should merge into one span
	sl.AddCell(10, 100)
	sl.AddCell(11, 150)
	sl.AddCell(12, 200)

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
	if span.Covers[0] != 100 || span.Covers[1] != 150 || span.Covers[2] != 200 {
		t.Errorf("Expected cover values [100, 150, 200], got [%d, %d, %d]",
			span.Covers[0], span.Covers[1], span.Covers[2])
	}
}

func TestAddNonConsecutiveCells(t *testing.T) {
	sl := NewScanlineU8()
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
	span1 := spans[0]
	if span1.X != 10 || span1.Len != 1 || span1.Covers[0] != 100 {
		t.Errorf("First span: expected (X:10, Len:1, Cover:100), got (X:%d, Len:%d, Cover:%d)",
			span1.X, span1.Len, span1.Covers[0])
	}

	// Second span
	span2 := spans[1]
	if span2.X != 15 || span2.Len != 1 || span2.Covers[0] != 150 {
		t.Errorf("Second span: expected (X:15, Len:1, Cover:150), got (X:%d, Len:%d, Cover:%d)",
			span2.X, span2.Len, span2.Covers[0])
	}
}

func TestAddCells(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Add multiple cells at once
	covers := []CoverType{100, 150, 200, 250}
	sl.AddCells(20, len(covers), covers)

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

	for i, expectedCover := range covers {
		if span.Covers[i] != expectedCover {
			t.Errorf("Expected cover[%d] to be %d, got %d", i, expectedCover, span.Covers[i])
		}
	}
}

func TestAddSpan(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Add a span with uniform coverage
	sl.AddSpan(30, 5, 180)

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

	// Check all coverage values are the same
	for i := 0; i < 5; i++ {
		if span.Covers[i] != 180 {
			t.Errorf("Expected cover[%d] to be 180, got %d", i, span.Covers[i])
		}
	}
}

func TestMixedOperations(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Mix different addition methods
	sl.AddCell(10, 100)
	sl.AddCell(11, 110) // Should merge with previous

	sl.AddSpan(20, 3, 200) // Separate span

	covers := []CoverType{250, 240}
	sl.AddCells(30, len(covers), covers) // Another separate span

	if sl.NumSpans() != 3 {
		t.Errorf("Expected 3 spans, got %d", sl.NumSpans())
	}

	spans := sl.Begin()

	// First span (merged cells)
	if spans[0].X != 10 || spans[0].Len != 2 {
		t.Errorf("First span: expected (X:10, Len:2), got (X:%d, Len:%d)", spans[0].X, spans[0].Len)
	}

	// Second span (uniform)
	if spans[1].X != 20 || spans[1].Len != 3 {
		t.Errorf("Second span: expected (X:20, Len:3), got (X:%d, Len:%d)", spans[1].X, spans[1].Len)
	}

	// Third span (from cells)
	if spans[2].X != 30 || spans[2].Len != 2 {
		t.Errorf("Third span: expected (X:30, Len:2), got (X:%d, Len:%d)", spans[2].X, spans[2].Len)
	}
}

func TestFinalize(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)
	sl.AddCell(10, 128)

	// Test finalize
	sl.Finalize(42)

	if sl.Y() != 42 {
		t.Errorf("Expected Y coordinate to be 42, got %d", sl.Y())
	}
}

func TestResetSpans(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)
	sl.AddCell(10, 128)
	sl.AddCell(11, 129)

	if sl.NumSpans() != 1 {
		t.Errorf("Expected 1 span before reset, got %d", sl.NumSpans())
	}

	// Reset spans
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

func TestEmptySpans(t *testing.T) {
	sl := NewScanlineU8()
	sl.Reset(0, 100)

	// Test with no spans added
	spans := sl.Begin()
	if spans != nil {
		t.Errorf("Expected nil spans when no spans added, got %v", spans)
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans, got %d", sl.NumSpans())
	}
}

func TestMinXOffset(t *testing.T) {
	sl := NewScanlineU8()

	// Test with non-zero minX
	minX := 50
	sl.Reset(minX, 150)

	// Add cell at absolute coordinate 60
	sl.AddCell(60, 128)

	spans := sl.Begin()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	// The span should have absolute X coordinate, not relative
	if spans[0].X != 60 {
		t.Errorf("Expected span X to be 60 (absolute coordinate), got %d", spans[0].X)
	}
}

func TestLargeCoordinates(t *testing.T) {
	sl := NewScanlineU8()

	// Test with larger coordinates near int16 limits
	minX := 30000
	maxX := 32000
	sl.Reset(minX, maxX)

	sl.AddCell(31000, 255)

	spans := sl.Begin()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	if spans[0].X != 31000 {
		t.Errorf("Expected span X to be 31000, got %d", spans[0].X)
	}
}
