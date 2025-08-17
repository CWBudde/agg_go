package scanline

import (
	"agg_go/internal/basics"
	"testing"
)

// MockBinScanline is a mock implementation of ScanlineInterface for testing.
type MockBinScanline struct {
	y       int
	spans   []MockBinSpan
	spanIdx int
}

// MockBinSpan represents a span for testing.
// This matches SpanInfo structure for easy conversion.
type MockBinSpan SpanInfo

// MockBinScanlineIterator provides iteration over MockBinSpan.
type MockBinScanlineIterator struct {
	spans   []MockBinSpan
	spanIdx int
}

func NewMockBinScanline(y int) *MockBinScanline {
	return &MockBinScanline{
		y:       y,
		spans:   make([]MockBinSpan, 0),
		spanIdx: 0,
	}
}

func (sl *MockBinScanline) Y() int {
	return sl.y
}

func (sl *MockBinScanline) NumSpans() int {
	return len(sl.spans)
}

func (sl *MockBinScanline) Begin() ScanlineIterator {
	return &MockBinScanlineIterator{
		spans:   sl.spans,
		spanIdx: 0,
	}
}

func (sl *MockBinScanline) ResetSpans() {
	sl.spans = sl.spans[:0]
	sl.spanIdx = 0
}

func (sl *MockBinScanline) AddSpan(x, length int, cover basics.Int8u) {
	sl.spans = append(sl.spans, MockBinSpan(SpanInfo{
		X:      x,
		Len:    length,
		Covers: []basics.Int8u{cover},
	}))
}

func (sl *MockBinScanline) AddCells(x, length int, covers []basics.Int8u) {
	sl.spans = append(sl.spans, MockBinSpan(SpanInfo{
		X:      x,
		Len:    length,
		Covers: covers,
	}))
}

func (sl *MockBinScanline) Finalize(y int) {
	sl.y = y
}

func (it *MockBinScanlineIterator) GetSpan() SpanInfo {
	if it.spanIdx >= len(it.spans) {
		return SpanInfo{}
	}
	span := it.spans[it.spanIdx]
	return SpanInfo(span)
}

func (it *MockBinScanlineIterator) Next() bool {
	it.spanIdx++
	return it.spanIdx < len(it.spans)
}

func TestScanlineStorageBinBasicOperations(t *testing.T) {
	storage := NewScanlineStorageBin()

	// Test initial state
	if storage.MinX() != 2147483647 || storage.MaxX() != -2147483648 {
		t.Errorf("Initial bounds should be at extreme values, got MinX=%d, MaxX=%d", storage.MinX(), storage.MaxX())
	}

	if storage.RewindScanlines() {
		t.Error("Empty storage should not have scanlines to rewind")
	}

	// Test prepare
	storage.Prepare()
	if !storage.RewindScanlines() == false {
		t.Error("After prepare, storage should be empty")
	}
}

func TestScanlineStorageBinRender(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Create a mock scanline with some spans
	scanline := NewMockBinScanline(10)
	scanline.AddSpan(5, 10, basics.CoverFull)
	scanline.AddSpan(20, 5, basics.CoverFull)

	// Render the scanline
	storage.Render(scanline)

	// Check bounds
	if storage.MinX() != 5 || storage.MaxX() != 24 {
		t.Errorf("Expected bounds X: [5, 24], got [%d, %d]", storage.MinX(), storage.MaxX())
	}

	if storage.MinY() != 10 || storage.MaxY() != 10 {
		t.Errorf("Expected bounds Y: [10, 10], got [%d, %d]", storage.MinY(), storage.MaxY())
	}

	// Test that we can rewind
	if !storage.RewindScanlines() {
		t.Error("Storage should have scanlines after rendering")
	}
}

func TestScanlineStorageBinMultipleScanlines(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Add multiple scanlines
	for y := 0; y < 3; y++ {
		scanline := NewMockBinScanline(y * 10)
		scanline.AddSpan(y*5, 10, basics.CoverFull)
		storage.Render(scanline)
	}

	// Check bounds include all scanlines
	if storage.MinY() != 0 || storage.MaxY() != 20 {
		t.Errorf("Expected Y bounds [0, 20], got [%d, %d]", storage.MinY(), storage.MaxY())
	}

	if storage.MinX() != 0 || storage.MaxX() != 19 {
		t.Errorf("Expected X bounds [0, 19], got [%d, %d]", storage.MinX(), storage.MaxX())
	}
}

func TestScanlineStorageBinSweepScanline(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Add a scanline
	inputScanline := NewMockBinScanline(15)
	inputScanline.AddSpan(100, 50, basics.CoverFull)
	storage.Render(inputScanline)

	// Rewind and sweep
	if !storage.RewindScanlines() {
		t.Error("Should be able to rewind")
	}

	outputScanline := NewMockBinScanline(0)
	if !storage.SweepScanline(outputScanline) {
		t.Error("Should be able to sweep scanline")
	}

	// Check the swept scanline
	if outputScanline.Y() != 15 {
		t.Errorf("Expected Y=15, got %d", outputScanline.Y())
	}

	if outputScanline.NumSpans() != 1 {
		t.Errorf("Expected 1 span, got %d", outputScanline.NumSpans())
	}

	span := SpanInfo(outputScanline.spans[0])
	if span.X != 100 || span.Len != 50 {
		t.Errorf("Expected span [100, 50], got [%d, %d]", span.X, span.Len)
	}

	// Try to sweep again - should return false
	if storage.SweepScanline(outputScanline) {
		t.Error("Second sweep should return false")
	}
}

func TestScanlineStorageBinEmbeddedScanline(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Add some data
	inputScanline := NewMockBinScanline(25)
	inputScanline.AddSpan(10, 5, basics.CoverFull)
	inputScanline.AddSpan(20, 3, basics.CoverFull)
	storage.Render(inputScanline)

	// Test embedded scanline
	embedded := NewEmbeddedScanlineBin(storage)

	if !storage.RewindScanlines() {
		t.Error("Should be able to rewind")
	}

	if !storage.SweepEmbeddedScanline(embedded) {
		t.Error("Should be able to sweep embedded scanline")
	}

	if embedded.Y() != 25 {
		t.Errorf("Expected Y=25, got %d", embedded.Y())
	}

	if embedded.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", embedded.NumSpans())
	}

	// Test iterator
	iterator := embedded.Begin()
	span1 := iterator.Span()
	if span1.X != 10 || span1.Len != 5 {
		t.Errorf("Expected first span [10, 5], got [%d, %d]", span1.X, span1.Len)
	}

	iterator.Next()
	span2 := iterator.Span()
	if span2.X != 20 || span2.Len != 3 {
		t.Errorf("Expected second span [20, 3], got [%d, %d]", span2.X, span2.Len)
	}
}

func TestScanlineStorageBinSerialization(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Add test data
	scanline1 := NewMockBinScanline(10)
	scanline1.AddSpan(5, 10, basics.CoverFull)
	storage.Render(scanline1)

	scanline2 := NewMockBinScanline(20)
	scanline2.AddSpan(15, 8, basics.CoverFull)
	scanline2.AddSpan(30, 5, basics.CoverFull)
	storage.Render(scanline2)

	// Test byte size calculation
	expectedSize := 16 + // bounds (4 int32s)
		8 + 8 + // scanline1: Y + num_spans + 1 span (2 int32s)
		8 + 16 // scanline2: Y + num_spans + 2 spans (4 int32s)

	actualSize := storage.ByteSize()
	if actualSize != expectedSize {
		t.Errorf("Expected byte size %d, got %d", expectedSize, actualSize)
	}

	// Test serialization
	data := make([]byte, actualSize)
	storage.Serialize(data)

	// Verify the serialized data has the correct size
	if len(data) != actualSize {
		t.Errorf("Serialized data size mismatch: expected %d, got %d", actualSize, len(data))
	}
}

func TestScanlineStorageBinRenderBinScanline(t *testing.T) {
	storage := NewScanlineStorageBin()
	storage.Prepare()

	// Create a binary scanline
	binScanline := NewScanlineBin()
	binScanline.Reset(0, 100)
	binScanline.AddSpan(10, 20, 0) // Cover value ignored for binary
	binScanline.AddSpan(40, 15, 0)
	binScanline.Finalize(30)

	// Render it
	storage.RenderBinScanline(binScanline)

	// Check bounds
	if storage.MinX() != 10 || storage.MaxX() != 54 {
		t.Errorf("Expected X bounds [10, 54], got [%d, %d]", storage.MinX(), storage.MaxX())
	}

	if storage.MinY() != 30 || storage.MaxY() != 30 {
		t.Errorf("Expected Y bounds [30, 30], got [%d, %d]", storage.MinY(), storage.MaxY())
	}

	// Verify we can sweep it back
	storage.RewindScanlines()
	outputScanline := NewMockBinScanline(0)
	if !storage.SweepScanline(outputScanline) {
		t.Error("Should be able to sweep the rendered binary scanline")
	}

	if outputScanline.Y() != 30 {
		t.Errorf("Expected Y=30, got %d", outputScanline.Y())
	}

	if outputScanline.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", outputScanline.NumSpans())
	}
}

func TestScanlineStorageBinBoundsAccess(t *testing.T) {
	storage := NewScanlineStorageBin()

	// Test out-of-bounds access returns fake data
	fakeScanline := storage.ScanlineByIndex(999)
	if fakeScanline.Y != 0 || fakeScanline.NumSpans != 0 {
		t.Error("Out-of-bounds scanline access should return fake scanline")
	}

	fakeSpan := storage.SpanByIndex(999)
	if fakeSpan.X != 0 || fakeSpan.Len != 0 {
		t.Error("Out-of-bounds span access should return fake span")
	}

	// Add some data and test valid access
	scanline := NewMockBinScanline(42)
	scanline.AddSpan(100, 200, basics.CoverFull)
	storage.Render(scanline)

	validScanline := storage.ScanlineByIndex(0)
	if validScanline.Y != 42 {
		t.Errorf("Expected Y=42, got %d", validScanline.Y)
	}

	validSpan := storage.SpanByIndex(0)
	if validSpan.X != 100 || validSpan.Len != 200 {
		t.Errorf("Expected span [100, 200], got [%d, %d]", validSpan.X, validSpan.Len)
	}
}

func TestSerializedScanlinesAdaptorBinBasic(t *testing.T) {
	// Create a storage, add data, and serialize it
	storage := NewScanlineStorageBin()
	storage.Prepare()

	scanline := NewMockBinScanline(15)
	scanline.AddSpan(50, 25, basics.CoverFull)
	storage.Render(scanline)

	// Serialize the data
	size := storage.ByteSize()
	data := make([]byte, size)
	storage.Serialize(data)

	// Create adaptor and test
	adaptor := NewSerializedScanlinesAdaptorBinWithData(data, 0.0, 0.0)

	if !adaptor.RewindScanlines() {
		t.Error("Should be able to rewind serialized data")
	}

	// Check bounds
	if adaptor.MinX() != 50 || adaptor.MaxX() != 74 {
		t.Errorf("Expected X bounds [50, 74], got [%d, %d]", adaptor.MinX(), adaptor.MaxX())
	}

	if adaptor.MinY() != 15 || adaptor.MaxY() != 15 {
		t.Errorf("Expected Y bounds [15, 15], got [%d, %d]", adaptor.MinY(), adaptor.MaxY())
	}
}

func TestSerializedScanlinesAdaptorBinSweep(t *testing.T) {
	// Create storage with multiple scanlines
	storage := NewScanlineStorageBin()
	storage.Prepare()

	for i := 0; i < 3; i++ {
		scanline := NewMockBinScanline(i * 10)
		scanline.AddSpan(i*20, 10, basics.CoverFull)
		storage.Render(scanline)
	}

	// Serialize
	size := storage.ByteSize()
	data := make([]byte, size)
	storage.Serialize(data)

	// Create adaptor with offset
	adaptor := NewSerializedScanlinesAdaptorBinWithData(data, 5.0, 3.0)
	adaptor.RewindScanlines()

	// Sweep first scanline
	outputScanline := NewMockBinScanline(0)
	if !adaptor.SweepScanline(outputScanline) {
		t.Error("Should be able to sweep first scanline")
	}

	// Check offset is applied
	if outputScanline.Y() != 3 { // 0 + 3 offset
		t.Errorf("Expected Y=3, got %d", outputScanline.Y())
	}

	if len(outputScanline.spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(outputScanline.spans))
	}

	span := SpanInfo(outputScanline.spans[0])
	if span.X != 5 { // 0 + 5 offset
		t.Errorf("Expected X=5, got %d", span.X)
	}
}

func TestSerializedScanlinesAdaptorBinEmbedded(t *testing.T) {
	// Create and serialize some data
	storage := NewScanlineStorageBin()
	storage.Prepare()

	scanline := NewMockBinScanline(100)
	scanline.AddSpan(10, 5, basics.CoverFull)
	scanline.AddSpan(20, 8, basics.CoverFull)
	storage.Render(scanline)

	size := storage.ByteSize()
	data := make([]byte, size)
	storage.Serialize(data)

	// Test embedded scanline adaptor
	adaptor := NewSerializedScanlinesAdaptorBinWithData(data, 0.0, 0.0)
	adaptor.RewindScanlines()

	embedded := NewEmbeddedScanlineSerial()
	if !adaptor.SweepEmbeddedScanline(embedded) {
		t.Error("Should be able to sweep embedded scanline")
	}

	if embedded.Y() != 100 {
		t.Errorf("Expected Y=100, got %d", embedded.Y())
	}

	if embedded.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", embedded.NumSpans())
	}

	// Test iterator
	iterator := embedded.Begin()
	span1 := iterator.Span()
	if span1.X != 10 || span1.Len != 5 {
		t.Errorf("Expected span [10, 5], got [%d, %d]", span1.X, span1.Len)
	}

	iterator.Next()
	span2 := iterator.Span()
	if span2.X != 20 || span2.Len != 8 {
		t.Errorf("Expected span [20, 8], got [%d, %d]", span2.X, span2.Len)
	}
}
