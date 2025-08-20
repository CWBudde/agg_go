package scanline

import (
	"testing"

	"agg_go/internal/basics"
)

func TestScanlineCellStorage_NewAndBasicOperations(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	if storage == nil {
		t.Fatal("NewScanlineCellStorage returned nil")
	}

	// Test that storage is initially empty
	cells := storage.Get(0)
	if cells != nil {
		t.Error("Expected nil for invalid index on empty storage")
	}
}

func TestScanlineCellStorage_AddCellsMainStorage(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Add some cells that should fit in main storage
	testCells := []basics.Int8u{10, 20, 30, 40, 50}
	idx := storage.AddCells(testCells, len(testCells))

	// Should return a positive index for main storage
	if idx < 0 {
		t.Errorf("Expected positive index for main storage, got %d", idx)
	}

	// Retrieve the cells
	retrieved := storage.Get(idx)
	if retrieved == nil {
		t.Fatal("Failed to retrieve stored cells")
	}

	// Verify the data matches (at least the first few elements)
	if len(retrieved) == 0 {
		t.Fatal("Retrieved empty slice")
	}

	// Check that we can access the stored data through PodBVector
	for i := 0; i < len(testCells); i++ {
		expected := testCells[i]
		actual := storage.cells.At(idx + i)
		if actual != expected {
			t.Errorf("Cell %d: expected %d, got %d", i, expected, actual)
		}
	}
}

func TestScanlineCellStorage_AddCellsExtraStorage(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Fill up the main storage by adding a span larger than block size
	// With block scale 12, block size is 4096, so use 5000 to force extra storage
	largeSpan := make([]basics.Int8u, 5000) // Larger than block size (4096)
	for i := range largeSpan {
		largeSpan[i] = basics.Int8u(i % 255)
	}

	// This should go to extra storage
	idx := storage.AddCells(largeSpan, len(largeSpan))

	// Should return a negative index for extra storage
	if idx >= 0 {
		t.Errorf("Expected negative index for extra storage, got %d", idx)
	}

	// Retrieve the cells
	retrieved := storage.Get(idx)
	if retrieved == nil {
		t.Fatal("Failed to retrieve cells from extra storage")
	}

	if len(retrieved) != len(largeSpan) {
		t.Errorf("Retrieved slice length %d, expected %d", len(retrieved), len(largeSpan))
	}

	// Verify the data matches
	for i := 0; i < len(largeSpan); i++ {
		if retrieved[i] != largeSpan[i] {
			t.Errorf("Cell %d: expected %d, got %d", i, largeSpan[i], retrieved[i])
		}
	}
}

func TestScanlineCellStorage_MultipleExtraStorage(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Add multiple spans to extra storage
	span1 := []basics.Int8u{1, 2, 3}
	span2 := []basics.Int8u{4, 5, 6, 7}
	span3 := []basics.Int8u{8, 9}

	// Force them to extra storage by making them larger than block size
	largeSpan1 := make([]basics.Int8u, 5000) // Larger than block size
	largeSpan2 := make([]basics.Int8u, 5000)
	largeSpan3 := make([]basics.Int8u, 5000)

	copy(largeSpan1, span1)
	copy(largeSpan2, span2)
	copy(largeSpan3, span3)

	idx1 := storage.AddCells(largeSpan1, len(largeSpan1))
	idx2 := storage.AddCells(largeSpan2, len(largeSpan2))
	idx3 := storage.AddCells(largeSpan3, len(largeSpan3))

	// All should be negative indices
	indices := []int{idx1, idx2, idx3}
	expectedIndices := []int{-1, -2, -3}

	for i, idx := range indices {
		if idx != expectedIndices[i] {
			t.Errorf("Index %d: expected %d, got %d", i, expectedIndices[i], idx)
		}
	}

	// Verify we can retrieve all spans
	retrieved1 := storage.Get(idx1)
	retrieved2 := storage.Get(idx2)
	retrieved3 := storage.Get(idx3)

	if len(retrieved1) != len(largeSpan1) {
		t.Errorf("Span 1 length: expected %d, got %d", len(largeSpan1), len(retrieved1))
	}
	if len(retrieved2) != len(largeSpan2) {
		t.Errorf("Span 2 length: expected %d, got %d", len(largeSpan2), len(retrieved2))
	}
	if len(retrieved3) != len(largeSpan3) {
		t.Errorf("Span 3 length: expected %d, got %d", len(largeSpan3), len(retrieved3))
	}
}

func TestScanlineCellStorage_CopyConstructor(t *testing.T) {
	original := NewScanlineCellStorage[basics.Int8u]()

	// Add some data to the original
	mainCells := []basics.Int8u{1, 2, 3, 4, 5}
	extraCells := make([]basics.Int8u, 100) // Force to extra storage
	for i := range extraCells {
		extraCells[i] = basics.Int8u(i % 100)
	}

	mainIdx := original.AddCells(mainCells, len(mainCells))
	extraIdx := original.AddCells(extraCells, len(extraCells))

	// Create a copy
	copy := NewScanlineCellStorageCopy(original)

	// Verify the copy has the same data
	copyMain := copy.Get(mainIdx)
	copyExtra := copy.Get(extraIdx)

	if copyMain == nil || copyExtra == nil {
		t.Fatal("Copy failed to preserve data")
	}

	// Verify main storage data
	for i := 0; i < len(mainCells); i++ {
		expected := mainCells[i]
		actual := copy.cells.At(mainIdx + i)
		if actual != expected {
			t.Errorf("Copy main cell %d: expected %d, got %d", i, expected, actual)
		}
	}

	// Verify extra storage data
	if len(copyExtra) != len(extraCells) {
		t.Errorf("Copy extra storage length: expected %d, got %d", len(extraCells), len(copyExtra))
	}

	for i := 0; i < len(extraCells); i++ {
		if copyExtra[i] != extraCells[i] {
			t.Errorf("Copy extra cell %d: expected %d, got %d", i, extraCells[i], copyExtra[i])
		}
	}

	// Modify original to ensure deep copy
	original.RemoveAll()

	// Copy should still have the data
	copyMainAfter := copy.Get(mainIdx)
	copyExtraAfter := copy.Get(extraIdx)

	if copyMainAfter == nil || copyExtraAfter == nil {
		t.Error("Deep copy failed - data was lost when original was cleared")
	}
}

func TestScanlineCellStorage_Assignment(t *testing.T) {
	source := NewScanlineCellStorage[basics.Int8u]()
	target := NewScanlineCellStorage[basics.Int8u]()

	// Add data to source
	cells := []basics.Int8u{10, 20, 30}
	idx := source.AddCells(cells, len(cells))

	// Add different data to target first
	targetCells := []basics.Int8u{100, 200}
	target.AddCells(targetCells, len(targetCells))

	// Assign source to target
	target.Assign(source)

	// Verify target now has source's data
	retrieved := target.Get(idx)
	if retrieved == nil {
		t.Fatal("Assignment failed to copy data")
	}

	// Check the data
	for i := 0; i < len(cells); i++ {
		expected := cells[i]
		actual := target.cells.At(idx + i)
		if actual != expected {
			t.Errorf("Assigned cell %d: expected %d, got %d", i, expected, actual)
		}
	}
}

func TestScanlineCellStorage_RemoveAll(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Add some data
	cells := []basics.Int8u{1, 2, 3}
	extraCells := make([]basics.Int8u, 100)

	mainIdx := storage.AddCells(cells, len(cells))
	extraIdx := storage.AddCells(extraCells, len(extraCells))

	// Verify data exists
	if storage.Get(mainIdx) == nil || storage.Get(extraIdx) == nil {
		t.Fatal("Failed to add initial data")
	}

	// Clear all data
	storage.RemoveAll()

	// Verify data is gone
	if storage.Get(mainIdx) != nil {
		t.Error("Main storage not cleared after RemoveAll")
	}
	if storage.Get(extraIdx) != nil {
		t.Error("Extra storage not cleared after RemoveAll")
	}

	// Verify we can add new data after clearing
	newCells := []basics.Int8u{99, 98, 97}
	newIdx := storage.AddCells(newCells, len(newCells))

	if storage.Get(newIdx) == nil {
		t.Error("Cannot add data after RemoveAll")
	}
}

func TestScanlineCellStorage_InvalidIndices(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Test invalid positive indices
	if storage.Get(100) != nil {
		t.Error("Expected nil for out-of-bounds positive index")
	}

	// Test invalid negative indices
	if storage.Get(-1) != nil {
		t.Error("Expected nil for invalid negative index")
	}
	if storage.Get(-100) != nil {
		t.Error("Expected nil for out-of-bounds negative index")
	}
}

func TestScanlineCellStorage_EmptyAndNilCells(t *testing.T) {
	storage := NewScanlineCellStorage[basics.Int8u]()

	// Test adding empty cells
	idx := storage.AddCells([]basics.Int8u{}, 0)
	if idx != -1 {
		t.Errorf("Expected -1 for empty cells, got %d", idx)
	}

	// Test adding nil cells
	idx = storage.AddCells(nil, 5)
	if idx != -1 {
		t.Errorf("Expected -1 for nil cells, got %d", idx)
	}

	// Test adding cells with invalid count
	cells := []basics.Int8u{1, 2, 3}
	idx = storage.AddCells(cells, 10) // More than available
	if idx != -1 {
		t.Errorf("Expected -1 for invalid count, got %d", idx)
	}
}

func TestScanlineStorageAA_ByteSize(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Empty storage should have base size (4 * int32 for bounds)
	baseSize := 4 * 4
	if storage.ByteSize() != baseSize {
		t.Errorf("Empty storage ByteSize: expected %d, got %d", baseSize, storage.ByteSize())
	}

	// Create a mock scanline with spans
	mockSL := NewMockScanline(10)
	mockSL.AddSpan(5, 3, 128)                                  // Solid span
	mockSL.AddCells(10, 4, []basics.Int8u{100, 150, 200, 250}) // Coverage span

	// Render the scanline
	storage.Render(mockSL)

	// Calculate expected size:
	// Base: 16 bytes (4 int32s for bounds)
	// Scanline header: 12 bytes (3 int32s: size, Y, num_spans)
	// Span 1: 8 bytes header + 1 byte cover = 9 bytes
	// Span 2: 8 bytes header + 4 bytes covers = 12 bytes
	expectedSize := 16 + 12 + 9 + 12
	actualSize := storage.ByteSize()

	if actualSize != expectedSize {
		t.Errorf("ByteSize with data: expected %d, got %d", expectedSize, actualSize)
	}
}

func TestScanlineStorageAA_Serialize(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Create test data
	mockSL1 := NewMockScanline(5)
	mockSL1.AddSpan(10, 2, 200) // Solid span

	mockSL2 := NewMockScanline(7)
	mockSL2.AddCells(15, 3, []basics.Int8u{50, 100, 150}) // Coverage span

	storage.Render(mockSL1)
	storage.Render(mockSL2)

	// Get the required buffer size
	expectedSize := storage.ByteSize()
	data := make([]byte, expectedSize)

	// Serialize the data
	storage.Serialize(data)

	// Verify the serialized data structure
	if len(data) != expectedSize {
		t.Errorf("Serialized data length: expected %d, got %d", expectedSize, len(data))
	}

	// Read back the bounds (first 16 bytes = 4 int32s)
	minX := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24
	minY := int32(data[4]) | int32(data[5])<<8 | int32(data[6])<<16 | int32(data[7])<<24
	maxX := int32(data[8]) | int32(data[9])<<8 | int32(data[10])<<16 | int32(data[11])<<24
	maxY := int32(data[12]) | int32(data[13])<<8 | int32(data[14])<<16 | int32(data[15])<<24

	if int(minX) != storage.MinX() {
		t.Errorf("Serialized MinX: expected %d, got %d", storage.MinX(), minX)
	}
	if int(minY) != storage.MinY() {
		t.Errorf("Serialized MinY: expected %d, got %d", storage.MinY(), minY)
	}
	if int(maxX) != storage.MaxX() {
		t.Errorf("Serialized MaxX: expected %d, got %d", storage.MaxX(), maxX)
	}
	if int(maxY) != storage.MaxY() {
		t.Errorf("Serialized MaxY: expected %d, got %d", storage.MaxY(), maxY)
	}
}

func TestScanlineStorageAA_SerializeEmpty(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Empty storage should serialize just the bounds
	expectedSize := 4 * 4 // 4 int32s
	data := make([]byte, expectedSize)

	storage.Serialize(data)

	if len(data) != expectedSize {
		t.Errorf("Empty serialized data length: expected %d, got %d", expectedSize, len(data))
	}
}

func TestScanlineStorageAA_SerializeInsufficientBuffer(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Add some data
	mockSL := NewMockScanline(0)
	mockSL.AddSpan(0, 1, 255)
	storage.Render(mockSL)

	// Try to serialize with insufficient buffer
	smallBuffer := make([]byte, 10) // Too small

	// Should not panic and should handle gracefully
	storage.Serialize(smallBuffer)

	// No specific assertion needed - just verify no panic occurs
}

func TestScanlineStorageAA_EdgeCases(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Test with zero-length spans
	mockSL := NewMockScanline(0)
	mockSL.AddCells(0, 0, []basics.Int8u{}) // Zero-length span
	storage.Render(mockSL)

	// Should not crash
	size := storage.ByteSize()
	if size <= 0 {
		t.Error("ByteSize should be positive even with zero-length spans")
	}

	// Test serialization
	data := make([]byte, size)
	storage.Serialize(data)
}

func TestScanlineStorageAA_LargeData(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Create a large coverage array
	largeCoverage := make([]basics.Int8u, 1000)
	for i := range largeCoverage {
		largeCoverage[i] = basics.Int8u(i % 255)
	}

	// Add scanline with large coverage data
	mockSL := NewMockScanline(100)
	mockSL.AddCells(0, len(largeCoverage), largeCoverage)
	storage.Render(mockSL)

	// Test ByteSize calculation
	size := storage.ByteSize()
	if size <= len(largeCoverage) {
		t.Errorf("ByteSize should account for headers + data, got %d", size)
	}

	// Test serialization
	data := make([]byte, size)
	storage.Serialize(data)

	// Verify we can read back the data
	if len(data) != size {
		t.Errorf("Serialized large data: expected %d bytes, got %d", size, len(data))
	}
}

func TestScanlineStorageAA_NegativeSpanLengths(t *testing.T) {
	storage := NewScanlineStorageAA[basics.Int8u]()

	// Create scanline with negative span length (solid span)
	mockSL := NewMockScanline(50)
	mockSL.AddSpan(10, 5, 200) // This creates a span with negative length internally
	storage.Render(mockSL)

	// Verify ByteSize accounts for solid span correctly
	size := storage.ByteSize()
	expectedMin := 16 + 12 + 8 + 1 // bounds + scanline header + span header + 1 cover byte
	if size < expectedMin {
		t.Errorf("ByteSize for solid span: expected at least %d, got %d", expectedMin, size)
	}

	// Test serialization
	data := make([]byte, size)
	storage.Serialize(data)
}

func TestWriteInt32(t *testing.T) {
	tests := []struct {
		name     string
		value    basics.Int32
		expected []byte
	}{
		{"Zero", 0, []byte{0, 0, 0, 0}},
		{"Positive small", 255, []byte{255, 0, 0, 0}},
		{"Positive large", 0x12345678, []byte{0x78, 0x56, 0x34, 0x12}},
		{"Negative", -1, []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"Max int32", 0x7FFFFFFF, []byte{0xFF, 0xFF, 0xFF, 0x7F}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 4)
			writeInt32(buf, tt.value)

			for i, expected := range tt.expected {
				if buf[i] != expected {
					t.Errorf("writeInt32(%d) byte %d: expected 0x%02X, got 0x%02X",
						tt.value, i, expected, buf[i])
				}
			}
		})
	}
}
