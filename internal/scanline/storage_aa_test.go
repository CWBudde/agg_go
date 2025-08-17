package scanline

import (
	"agg_go/internal/basics"
	"testing"
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

	// Fill up the main storage by adding many large spans
	// The PodBVector has limited continuous block size
	largeSpan := make([]basics.Int8u, 100) // Larger than typical block size
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

	// Force them to extra storage by making them large enough
	largeSpan1 := make([]basics.Int8u, 100)
	largeSpan2 := make([]basics.Int8u, 100)
	largeSpan3 := make([]basics.Int8u, 100)

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
