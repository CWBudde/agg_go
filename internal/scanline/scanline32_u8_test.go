package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

func TestScanline32U8_NewScanline32U8(t *testing.T) {
	sl := NewScanline32U8()
	if sl == nil {
		t.Fatal("NewScanline32U8() returned nil")
	}

	if sl.Y() != 0 {
		t.Errorf("Expected Y() = 0, got %d", sl.Y())
	}

	if sl.NumSpans() != 0 {
		t.Errorf("Expected NumSpans() = 0, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX to be 0x7FFFFFF0, got %d", sl.lastX)
	}
}

func TestScanline32U8_Reset(t *testing.T) {
	sl := NewScanline32U8()

	// Test basic reset
	sl.Reset(10, 100)
	if sl.minX != 10 {
		t.Errorf("Expected minX = 10, got %d", sl.minX)
	}
	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX = 0x7FFFFFF0, got %d", sl.lastX)
	}

	// Test that covers array is sized correctly
	expectedSize := 100 - 10 + 2 // 92
	if sl.covers.Size() < expectedSize {
		t.Errorf("Expected covers size >= %d, got %d", expectedSize, sl.covers.Size())
	}
}

func TestScanline32U8_AddCell(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 100)

	// Add first cell
	sl.AddCell(10, 255)

	if sl.NumSpans() != 1 {
		t.Fatalf("Expected 1 span after AddCell, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span in slice, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Expected span X = 10, got %d", span.X)
	}
	if span.Len != 1 {
		t.Errorf("Expected span Len = 1, got %d", span.Len)
	}

	// Check coverage value
	if span.Covers[0] != 255 {
		t.Errorf("Expected coverage = 255, got %d", span.Covers[0])
	}

	// Add adjacent cell (should extend span)
	sl.AddCell(11, 200)

	if sl.NumSpans() != 1 {
		t.Fatalf("Expected 1 span after extending, got %d", sl.NumSpans())
	}

	spans = sl.Spans()
	span = spans[0]
	if span.Len != 2 {
		t.Errorf("Expected extended span Len = 2, got %d", span.Len)
	}

	// Add non-adjacent cell (should create new span)
	sl.AddCell(15, 100)

	if sl.NumSpans() != 2 {
		t.Fatalf("Expected 2 spans after non-adjacent add, got %d", sl.NumSpans())
	}
}

func TestScanline32U8_AddCells(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 100)

	covers := []basics.Int8u{255, 200, 150, 100}
	sl.AddCells(10, len(covers), covers)

	if sl.NumSpans() != 1 {
		t.Fatalf("Expected 1 span after AddCells, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	span := spans[0]

	if span.X != 10 {
		t.Errorf("Expected span X = 10, got %d", span.X)
	}
	if span.Len != 4 {
		t.Errorf("Expected span Len = 4, got %d", span.Len)
	}

	// Check coverage values
	for i, expectedCover := range covers {
		if span.Covers[i] != expectedCover {
			t.Errorf("Expected coverage[%d] = %d, got %d", i, expectedCover, span.Covers[i])
		}
	}
}

func TestScanline32U8_AddSpan(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 100)

	// Add solid span
	sl.AddSpan(20, 5, 128)

	if sl.NumSpans() != 1 {
		t.Fatalf("Expected 1 span after AddSpan, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	span := spans[0]

	if span.X != 20 {
		t.Errorf("Expected span X = 20, got %d", span.X)
	}
	if span.Len != 5 {
		t.Errorf("Expected span Len = 5, got %d", span.Len)
	}

	// Check that all coverage values are the same
	for i := 0; i < 5; i++ {
		if span.Covers[i] != 128 {
			t.Errorf("Expected coverage[%d] = 128, got %d", i, span.Covers[i])
		}
	}
}

func TestScanline32U8_Finalize(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 100)

	// Test finalize sets Y coordinate
	sl.Finalize(42)
	if sl.Y() != 42 {
		t.Errorf("Expected Y() = 42 after Finalize, got %d", sl.Y())
	}
}

func TestScanline32U8_ResetSpans(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 100)

	// Add some spans
	sl.AddCell(10, 255)
	sl.AddCell(20, 200)

	if sl.NumSpans() != 2 {
		t.Fatalf("Expected 2 spans before reset, got %d", sl.NumSpans())
	}

	// Reset spans
	sl.ResetSpans()

	if sl.NumSpans() != 0 {
		t.Errorf("Expected 0 spans after ResetSpans, got %d", sl.NumSpans())
	}

	if sl.lastX != 0x7FFFFFF0 {
		t.Errorf("Expected lastX reset to 0x7FFFFFF0, got %d", sl.lastX)
	}
}

func TestScanline32U8_MixedOperations(t *testing.T) {
	sl := NewScanline32U8()
	sl.Reset(0, 200)

	// Test mixed operations: cell, span, cells
	sl.AddCell(10, 255)
	sl.AddSpan(11, 3, 200)                       // Should extend first span
	sl.AddCells(14, 2, []basics.Int8u{150, 100}) // Should extend further

	if sl.NumSpans() != 1 {
		t.Fatalf("Expected 1 span after mixed adjacent operations, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	span := spans[0]

	if span.X != 10 {
		t.Errorf("Expected span X = 10, got %d", span.X)
	}
	if span.Len != 6 { // 1 + 3 + 2
		t.Errorf("Expected span Len = 6, got %d", span.Len)
	}

	// Check coverage pattern
	expectedCovers := []basics.Int8u{255, 200, 200, 200, 150, 100}
	for i, expected := range expectedCovers {
		if span.Covers[i] != expected {
			t.Errorf("Expected coverage[%d] = %d, got %d", i, expected, span.Covers[i])
		}
	}

	// Add non-adjacent span
	sl.AddSpan(50, 2, 50)

	if sl.NumSpans() != 2 {
		t.Fatalf("Expected 2 spans after non-adjacent add, got %d", sl.NumSpans())
	}
}

func TestScanline32U8_LargeCoordinates(t *testing.T) {
	sl := NewScanline32U8()

	// Test with large 32-bit coordinates
	minX := 1000000
	maxX := 1000100
	sl.Reset(minX, maxX)

	sl.AddCell(minX+10, 255)
	sl.AddSpan(minX+50, 5, 128)

	if sl.NumSpans() != 2 {
		t.Fatalf("Expected 2 spans with large coordinates, got %d", sl.NumSpans())
	}

	spans := sl.Spans()

	if spans[0].X != Coord32Type(minX+10) {
		t.Errorf("Expected first span X = %d, got %d", minX+10, spans[0].X)
	}
	if spans[1].X != Coord32Type(minX+50) {
		t.Errorf("Expected second span X = %d, got %d", minX+50, spans[1].X)
	}
}
