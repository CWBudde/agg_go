package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

func TestScanlineP8_BasicOperations(t *testing.T) {
	sl := NewScanlineP8()

	// Test initial state
	if sl.Y() != 0 {
		t.Errorf("Initial Y should be 0, got %d", sl.Y())
	}
	if sl.NumSpans() != 0 {
		t.Errorf("Initial NumSpans should be 0, got %d", sl.NumSpans())
	}

	// Test reset
	sl.Reset(0, 100)

	// Add some cells
	sl.AddCell(10, 128)
	sl.AddCell(11, 200)
	sl.AddCell(12, 255)

	// Finalize
	sl.Finalize(5)

	// Check Y coordinate
	if sl.Y() != 5 {
		t.Errorf("Y after finalize should be 5, got %d", sl.Y())
	}

	// Check number of spans (should be 1 since cells are consecutive)
	if sl.NumSpans() != 1 {
		t.Errorf("NumSpans should be 1, got %d", sl.NumSpans())
	}

	// Check span details
	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Span X should be 10, got %d", span.X)
	}
	if span.Len != 3 {
		t.Errorf("Span length should be 3, got %d", span.Len)
	}
	if span.IsSolid() {
		t.Error("Span should not be solid")
	}
}

func TestScanlineP8_NonConsecutiveCells(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add non-consecutive cells
	sl.AddCell(10, 128)
	sl.AddCell(20, 200)
	sl.AddCell(30, 255)

	sl.Finalize(10)

	// Should have 3 separate spans
	if sl.NumSpans() != 3 {
		t.Errorf("NumSpans should be 3, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 3 {
		t.Fatalf("Should have 3 spans, got %d", len(spans))
	}

	// Check each span
	expectedX := []basics.Int32{10, 20, 30}
	for i, span := range spans {
		if span.X != expectedX[i] {
			t.Errorf("Span %d: X should be %d, got %d", i, expectedX[i], span.X)
		}
		if span.Len != 1 {
			t.Errorf("Span %d: length should be 1, got %d", i, span.Len)
		}
	}
}

func TestScanlineP8_SolidSpans(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add a solid span
	sl.AddSpan(10, 5, 128)

	sl.Finalize(15)

	// Should have 1 solid span
	if sl.NumSpans() != 1 {
		t.Errorf("NumSpans should be 1, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Span X should be 10, got %d", span.X)
	}
	if span.Len != -5 {
		t.Errorf("Span length should be -5 (solid), got %d", span.Len)
	}
	if !span.IsSolid() {
		t.Error("Span should be solid")
	}
	if span.ActualLen() != 5 {
		t.Errorf("Actual length should be 5, got %d", span.ActualLen())
	}
}

func TestScanlineP8_MergeSolidSpans(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add consecutive solid spans with same coverage
	sl.AddSpan(10, 5, 128)
	sl.AddSpan(15, 3, 128)

	sl.Finalize(20)

	// Should have merged into 1 solid span
	if sl.NumSpans() != 1 {
		t.Errorf("NumSpans should be 1 (merged), got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Span X should be 10, got %d", span.X)
	}
	if span.Len != -8 {
		t.Errorf("Span length should be -8 (merged solid), got %d", span.Len)
	}
	if span.ActualLen() != 8 {
		t.Errorf("Actual length should be 8, got %d", span.ActualLen())
	}
}

func TestScanlineP8_NoMergeDifferentCoverage(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add consecutive solid spans with different coverage
	sl.AddSpan(10, 5, 128)
	sl.AddSpan(15, 3, 200) // Different coverage

	sl.Finalize(25)

	// Should NOT merge - different coverage values
	if sl.NumSpans() != 2 {
		t.Errorf("NumSpans should be 2 (not merged), got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 2 {
		t.Fatalf("Should have 2 spans, got %d", len(spans))
	}
}

func TestScanlineP8_AddCells(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add multiple cells at once
	covers := []basics.Int8u{100, 150, 200, 250}
	sl.AddCells(10, 4, covers)

	sl.Finalize(30)

	// Should have 1 span with 4 cells
	if sl.NumSpans() != 1 {
		t.Errorf("NumSpans should be 1, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.X != 10 {
		t.Errorf("Span X should be 10, got %d", span.X)
	}
	if span.Len != 4 {
		t.Errorf("Span length should be 4, got %d", span.Len)
	}
	if span.IsSolid() {
		t.Error("Span should not be solid")
	}
}

func TestScanlineP8_MixedOperations(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Mix of different operations
	sl.AddCell(5, 100)
	sl.AddCell(6, 110)
	sl.AddSpan(10, 5, 128) // Solid span
	sl.AddCell(20, 200)
	covers := []basics.Int8u{210, 220, 230}
	sl.AddCells(21, 3, covers)
	sl.AddSpan(30, 2, 255) // Another solid span

	sl.Finalize(35)

	// Should have 4 spans total
	expectedSpans := 4
	if sl.NumSpans() != expectedSpans {
		t.Errorf("NumSpans should be %d, got %d", expectedSpans, sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != expectedSpans {
		t.Fatalf("Should have %d spans, got %d", expectedSpans, len(spans))
	}

	// Verify span properties
	// Span 0: cells at 5-6
	if spans[0].X != 5 || spans[0].Len != 2 {
		t.Errorf("Span 0: expected X=5, Len=2, got X=%d, Len=%d", spans[0].X, spans[0].Len)
	}

	// Span 1: solid span at 10-14
	if spans[1].X != 10 || spans[1].Len != -5 {
		t.Errorf("Span 1: expected X=10, Len=-5, got X=%d, Len=%d", spans[1].X, spans[1].Len)
	}

	// Span 2: cells at 20-23
	if spans[2].X != 20 || spans[2].Len != 4 {
		t.Errorf("Span 2: expected X=20, Len=4, got X=%d, Len=%d", spans[2].X, spans[2].Len)
	}

	// Span 3: solid span at 30-31
	if spans[3].X != 30 || spans[3].Len != -2 {
		t.Errorf("Span 3: expected X=30, Len=-2, got X=%d, Len=%d", spans[3].X, spans[3].Len)
	}
}

func TestScanlineP8_ResetSpans(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Add some data
	sl.AddCell(10, 128)
	sl.AddSpan(20, 5, 200)
	sl.Finalize(40)

	// Verify data exists
	if sl.NumSpans() != 2 {
		t.Errorf("Should have 2 spans before reset, got %d", sl.NumSpans())
	}

	// Reset spans
	sl.ResetSpans()

	// Should be ready for new data
	if sl.NumSpans() != 0 {
		t.Errorf("NumSpans should be 0 after reset, got %d", sl.NumSpans())
	}

	// Add new data
	sl.AddCell(5, 100)
	sl.Finalize(45)

	// Should have new data
	if sl.NumSpans() != 1 {
		t.Errorf("Should have 1 span after adding new data, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span, got %d", len(spans))
	}
	if spans[0].X != 5 {
		t.Errorf("New span X should be 5, got %d", spans[0].X)
	}
}

func TestScanlineP8_EmptyScanline(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 100)

	// Don't add any cells/spans
	sl.Finalize(50)

	// Should have no spans
	if sl.NumSpans() != 0 {
		t.Errorf("Empty scanline should have 0 spans, got %d", sl.NumSpans())
	}

	spans := sl.Spans()
	if spans != nil {
		t.Error("Empty scanline should return nil spans")
	}
}

func TestScanlineP8_LargeScanline(t *testing.T) {
	sl := NewScanlineP8()
	sl.Reset(0, 1000)

	// Add many cells
	for x := 0; x < 1000; x += 10 {
		sl.AddCell(x, uint(x%256))
	}

	sl.Finalize(60)

	// Should have 100 spans (one for each non-consecutive cell)
	if sl.NumSpans() != 100 {
		t.Errorf("Should have 100 spans, got %d", sl.NumSpans())
	}
}

func TestScanlineP8_BoundaryConditions(t *testing.T) {
	sl := NewScanlineP8()

	// Test with minimum range
	sl.Reset(0, 0)
	sl.AddCell(0, 255)
	sl.Finalize(70)

	if sl.NumSpans() != 1 {
		t.Errorf("Should handle minimum range, got %d spans", sl.NumSpans())
	}

	// Test with large coordinates
	sl.Reset(10000, 10100)
	sl.AddCell(10050, 128)
	sl.Finalize(80)

	spans := sl.Spans()
	if len(spans) != 1 {
		t.Fatalf("Should have 1 span with large coordinates, got %d", len(spans))
	}
	if spans[0].X != 10050 {
		t.Errorf("Span X should be 10050, got %d", spans[0].X)
	}
}
