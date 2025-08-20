package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

// MockScanline is a test implementation of ScanlineInterface
type MockScanline struct {
	y          int
	spans      []MockSpan
	resetCalls int
}

type MockSpan struct {
	X      int
	Len    int
	Covers []basics.Int8u
}

type MockScanlineIterator struct {
	spans []MockSpan
	index int
}

func NewMockScanline(y int) *MockScanline {
	return &MockScanline{
		y:     y,
		spans: make([]MockSpan, 0),
	}
}

func (m *MockScanline) Y() int {
	return m.y
}

func (m *MockScanline) NumSpans() int {
	return len(m.spans)
}

func (m *MockScanline) Begin() ScanlineIterator {
	return &MockScanlineIterator{
		spans: m.spans,
		index: 0,
	}
}

func (m *MockScanline) ResetSpans() {
	m.resetCalls++
	m.spans = m.spans[:0] // Clear spans but keep capacity
}

func (m *MockScanline) AddSpan(x, length int, cover basics.Int8u) {
	// For solid spans, create a single-element covers array
	covers := []basics.Int8u{cover}
	m.spans = append(m.spans, MockSpan{
		X:      x,
		Len:    -length, // Negative length indicates solid span
		Covers: covers,
	})
}

func (m *MockScanline) AddCells(x, length int, covers []basics.Int8u) {
	m.spans = append(m.spans, MockSpan{
		X:      x,
		Len:    length,
		Covers: covers,
	})
}

func (m *MockScanline) Finalize(y int) {
	m.y = y
}

func (mi *MockScanlineIterator) GetSpan() SpanInfo {
	if mi.index >= len(mi.spans) {
		return SpanInfo{}
	}
	span := mi.spans[mi.index]
	return SpanInfo(span)
}

func (mi *MockScanlineIterator) Next() bool {
	mi.index++
	return mi.index < len(mi.spans)
}

func TestScanlineStorageAA_NewAndBasicOperations(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	if storage == nil {
		t.Fatal("NewScanlineStorageAA returned nil")
	}

	// Test initial bounds (should be at extremes)
	if storage.MinX() != 2147483647 || storage.MinY() != 2147483647 {
		t.Errorf("Initial min bounds incorrect: got MinX=%d, MinY=%d", storage.MinX(), storage.MinY())
	}

	if storage.MaxX() != -2147483648 || storage.MaxY() != -2147483648 {
		t.Errorf("Initial max bounds incorrect: got MaxX=%d, MaxY=%d", storage.MaxX(), storage.MaxY())
	}

	// Test that RewindScanlines returns false when empty
	if storage.RewindScanlines() {
		t.Error("RewindScanlines should return false for empty storage")
	}
}

func TestScanlineStorageAA_StoreSingleScanline(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Create a mock scanline with test data
	mockScanline := NewMockScanline(100)

	// Add some spans to the mock scanline
	mockScanline.spans = append(mockScanline.spans, MockSpan{
		X:      10,
		Len:    5,
		Covers: []basics.Int8u{255, 200, 150, 100, 50},
	})
	mockScanline.spans = append(mockScanline.spans, MockSpan{
		X:      20,
		Len:    -3,                  // Solid span
		Covers: []basics.Int8u{128}, // Single coverage value for solid span
	})

	// Store the scanline
	storage.Render(mockScanline)

	// Check bounds were updated correctly
	expectedMinX := 10
	expectedMaxX := 22 // 20 + 3 - 1
	expectedMinY := 100
	expectedMaxY := 100

	if storage.MinX() != expectedMinX {
		t.Errorf("MinX: expected %d, got %d", expectedMinX, storage.MinX())
	}
	if storage.MaxX() != expectedMaxX {
		t.Errorf("MaxX: expected %d, got %d", expectedMaxX, storage.MaxX())
	}
	if storage.MinY() != expectedMinY {
		t.Errorf("MinY: expected %d, got %d", expectedMinY, storage.MinY())
	}
	if storage.MaxY() != expectedMaxY {
		t.Errorf("MaxY: expected %d, got %d", expectedMaxY, storage.MaxY())
	}

	// Test that RewindScanlines now returns true
	if !storage.RewindScanlines() {
		t.Error("RewindScanlines should return true after storing scanlines")
	}
}

func TestScanlineStorageAA_StoreMultipleScanlines(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Store multiple scanlines
	for y := 50; y <= 52; y++ {
		mockScanline := NewMockScanline(y)
		mockScanline.spans = append(mockScanline.spans, MockSpan{
			X:      y, // Use Y as X for variety
			Len:    2,
			Covers: []basics.Int8u{255, 128},
		})
		storage.Render(mockScanline)
	}

	// Check bounds
	expectedMinX := 50
	expectedMaxX := 53 // 52 + 2 - 1 = 53
	expectedMinY := 50
	expectedMaxY := 52

	if storage.MinX() != expectedMinX {
		t.Errorf("MinX: expected %d, got %d", expectedMinX, storage.MinX())
	}
	if storage.MaxX() != expectedMaxX {
		t.Errorf("MaxX: expected %d, got %d", expectedMaxX, storage.MaxX())
	}
	if storage.MinY() != expectedMinY {
		t.Errorf("MinY: expected %d, got %d", expectedMinY, storage.MinY())
	}
	if storage.MaxY() != expectedMaxY {
		t.Errorf("MaxY: expected %d, got %d", expectedMaxY, storage.MaxY())
	}
}

func TestScanlineStorageAA_SweepScanline(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Store a scanline
	originalScanline := NewMockScanline(200)
	originalScanline.spans = append(originalScanline.spans, MockSpan{
		X:      100,
		Len:    3,
		Covers: []basics.Int8u{255, 200, 150},
	})
	storage.Render(originalScanline)

	// Prepare for sweep
	storage.RewindScanlines()

	// Create a target scanline for sweeping
	targetScanline := NewMockScanline(0)

	// Sweep the scanline
	result := storage.SweepScanline(targetScanline)

	if !result {
		t.Error("SweepScanline should return true when data is available")
	}

	// Check that the target scanline was populated correctly
	if targetScanline.Y() != 200 {
		t.Errorf("Target scanline Y: expected 200, got %d", targetScanline.Y())
	}

	if targetScanline.NumSpans() != 1 {
		t.Errorf("Target scanline spans: expected 1, got %d", targetScanline.NumSpans())
	}

	if targetScanline.resetCalls != 1 {
		t.Errorf("ResetSpans should have been called once, called %d times", targetScanline.resetCalls)
	}

	// Test that second sweep returns false (no more data)
	result = storage.SweepScanline(targetScanline)
	if result {
		t.Error("Second SweepScanline should return false")
	}
}

func TestScanlineStorageAA_EmbeddedScanline(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Store a scanline
	mockScanline := NewMockScanline(300)
	mockScanline.spans = append(mockScanline.spans, MockSpan{
		X:      50,
		Len:    2,
		Covers: []basics.Int8u{255, 128},
	})
	mockScanline.spans = append(mockScanline.spans, MockSpan{
		X:      60,
		Len:    -4, // Solid span
		Covers: []basics.Int8u{200},
	})
	storage.Render(mockScanline)

	// Create embedded scanline
	embedded := NewEmbeddedScanline(storage)

	// Test sweep
	storage.RewindScanlines()
	result := storage.SweepEmbeddedScanline(embedded)

	if !result {
		t.Error("SweepEmbeddedScanline should return true when data is available")
	}

	if embedded.Y() != 300 {
		t.Errorf("Embedded scanline Y: expected 300, got %d", embedded.Y())
	}

	if embedded.NumSpans() != 2 {
		t.Errorf("Embedded scanline spans: expected 2, got %d", embedded.NumSpans())
	}

	// Test iteration
	iter := embedded.Begin()
	if iter == nil {
		t.Fatal("Begin() returned nil iterator")
	}

	// Check first span
	span1 := iter.GetSpan()
	if span1.X != 50 {
		t.Errorf("First span X: expected 50, got %d", span1.X)
	}
	if span1.Len != 2 {
		t.Errorf("First span Len: expected 2, got %d", span1.Len)
	}

	// Advance to next span
	iter.Next()
	span2 := iter.GetSpan()
	if span2.X != 60 {
		t.Errorf("Second span X: expected 60, got %d", span2.X)
	}
	if span2.Len != -4 {
		t.Errorf("Second span Len: expected -4, got %d", span2.Len)
	}
}

func TestScanlineStorageAA_Prepare(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Store some data
	mockScanline := NewMockScanline(100)
	mockScanline.spans = append(mockScanline.spans, MockSpan{
		X:      10,
		Len:    2,
		Covers: []basics.Int8u{255, 128},
	})
	storage.Render(mockScanline)

	// Verify data was stored
	if !storage.RewindScanlines() {
		t.Error("Storage should have data before Prepare")
	}

	// Clear with Prepare
	storage.Prepare()

	// Verify data was cleared
	if storage.RewindScanlines() {
		t.Error("Storage should be empty after Prepare")
	}

	// Check bounds were reset
	if storage.MinX() != 2147483647 || storage.MinY() != 2147483647 {
		t.Error("Bounds should be reset after Prepare")
	}
}

func TestScanlineStorageAA_ConcreteTypes(t *testing.T) {
	// Test that concrete type aliases work
	storage8 := NewScanlineStorageAA[basics.Int8u]()
	storage16 := NewScanlineStorageAA[basics.Int16u]()
	storage32 := NewScanlineStorageAA[basics.Int32u]()

	if storage8 == nil || storage16 == nil || storage32 == nil {
		t.Error("Concrete type constructors should work")
	}

	// Test type aliases
	var _ = storage8
	var _ = storage16
	var _ = storage32
}

func TestScanlineStorageAA_BoundsChecking(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Test bounds checking for empty storage
	scanlineData := storage.ScanlineByIndex(0)
	if scanlineData.Y != 0 || scanlineData.NumSpans != 0 {
		t.Error("ScanlineByIndex should return fake data for invalid index")
	}

	spanData := storage.SpanByIndex(0)
	if spanData.X != 0 || spanData.Len != 0 {
		t.Error("SpanByIndex should return fake data for invalid index")
	}

	covers := storage.CoversByIndex(0)
	if covers != nil {
		t.Error("CoversByIndex should return nil for invalid index")
	}
}

func TestScanlineStorageAA_LargeScanlines(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Create a scanline with many spans
	mockScanline := NewMockScanline(1000)

	// Add 100 spans
	for i := 0; i < 100; i++ {
		covers := make([]basics.Int8u, 10)
		for j := range covers {
			covers[j] = basics.Int8u(i + j)
		}
		mockScanline.spans = append(mockScanline.spans, MockSpan{
			X:      i * 20,
			Len:    10,
			Covers: covers,
		})
	}

	storage.Render(mockScanline)

	// Test bounds
	expectedMinX := 0
	expectedMaxX := 99*20 + 10 - 1 // Last span starts at 99*20, length 10
	if storage.MinX() != expectedMinX {
		t.Errorf("MinX: expected %d, got %d", expectedMinX, storage.MinX())
	}
	if storage.MaxX() != expectedMaxX {
		t.Errorf("MaxX: expected %d, got %d", expectedMaxX, storage.MaxX())
	}

	// Test sweep
	storage.RewindScanlines()
	targetScanline := NewMockScanline(0)

	result := storage.SweepScanline(targetScanline)
	if !result {
		t.Error("SweepScanline should handle large scanlines")
	}

	if targetScanline.NumSpans() != 100 {
		t.Errorf("Target should have 100 spans, got %d", targetScanline.NumSpans())
	}
}
