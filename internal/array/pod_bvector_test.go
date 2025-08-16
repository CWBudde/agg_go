package array

import (
	"testing"
)

func TestBlockScale(t *testing.T) {
	scale := NewBlockScale(6)

	if scale.Shift != 6 {
		t.Errorf("Expected shift 6, got %d", scale.Shift)
	}
	if scale.Size != 64 {
		t.Errorf("Expected size 64, got %d", scale.Size)
	}
	if scale.Mask != 63 {
		t.Errorf("Expected mask 63, got %d", scale.Mask)
	}

	// Test default scale
	if DefaultBlockScale.Shift != 6 {
		t.Errorf("Expected default shift 6, got %d", DefaultBlockScale.Shift)
	}
}

func TestPodBVectorBasic(t *testing.T) {
	bv := NewPodBVector[int]()

	// Test initial state
	if bv.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", bv.Size())
	}
	if bv.Capacity() != 0 {
		t.Errorf("Expected initial capacity 0, got %d", bv.Capacity())
	}

	// Test adding elements
	for i := 0; i < 100; i++ {
		bv.Add(i * 2)
	}

	if bv.Size() != 100 {
		t.Errorf("Expected size 100, got %d", bv.Size())
	}

	// Verify all elements
	for i := 0; i < bv.Size(); i++ {
		expected := i * 2
		if bv.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, bv.At(i))
		}
		if bv.ValueAt(i) != expected {
			t.Errorf("ValueAt(%d): expected %d, got %d", i, expected, bv.ValueAt(i))
		}
	}

	// Test element modification
	bv.Set(50, 999)
	if bv.At(50) != 999 {
		t.Errorf("Set(50, 999): expected 999, got %d", bv.At(50))
	}
}

func TestPodBVectorWithCustomScale(t *testing.T) {
	scale := NewBlockScale(4) // Block size = 16
	bv := NewPodBVectorWithScale[int](scale)

	// Add elements across multiple blocks
	for i := 0; i < 50; i++ {
		bv.Add(i)
	}

	if bv.Size() != 50 {
		t.Errorf("Expected size 50, got %d", bv.Size())
	}

	// Should have at least 4 blocks (50 / 16 = 3.125, rounded up)
	expectedBlocks := (50 + scale.Size - 1) / scale.Size
	if bv.numBlocks < expectedBlocks {
		t.Errorf("Expected at least %d blocks, got %d", expectedBlocks, bv.numBlocks)
	}
}

func TestPodBVectorPushBack(t *testing.T) {
	bv := NewPodBVector[string]()

	testStrings := []string{"hello", "world", "test", "data"}
	for _, s := range testStrings {
		bv.PushBack(s)
	}

	if bv.Size() != len(testStrings) {
		t.Errorf("Expected size %d, got %d", len(testStrings), bv.Size())
	}

	for i, expected := range testStrings {
		if bv.At(i) != expected {
			t.Errorf("At(%d): expected %s, got %s", i, expected, bv.At(i))
		}
	}
}

func TestPodBVectorModifyLast(t *testing.T) {
	bv := NewPodBVector[int]()

	bv.Add(1)
	bv.Add(2)
	bv.Add(3)

	// Modify last element
	bv.ModifyLast(99)

	if bv.Size() != 3 {
		t.Errorf("Expected size 3, got %d", bv.Size())
	}
	if bv.At(2) != 99 {
		t.Errorf("Expected last element to be 99, got %d", bv.At(2))
	}
}

func TestPodBVectorRemoveLast(t *testing.T) {
	bv := NewPodBVector[int]()

	for i := 0; i < 5; i++ {
		bv.Add(i)
	}

	bv.RemoveLast()

	if bv.Size() != 4 {
		t.Errorf("Expected size 4, got %d", bv.Size())
	}

	// Remove from empty should not crash
	empty := NewPodBVector[int]()
	empty.RemoveLast()
	if empty.Size() != 0 {
		t.Error("RemoveLast on empty vector should keep size 0")
	}
}

func TestPodBVectorClear(t *testing.T) {
	bv := NewPodBVector[int]()

	for i := 0; i < 100; i++ {
		bv.Add(i)
	}

	bv.Clear()
	if bv.Size() != 0 {
		t.Errorf("Clear: expected size 0, got %d", bv.Size())
	}

	// Should be able to add after clear
	bv.Add(42)
	if bv.Size() != 1 || bv.At(0) != 42 {
		t.Error("Failed to add after clear")
	}
}

func TestPodBVectorRemoveAll(t *testing.T) {
	bv := NewPodBVector[int]()

	for i := 0; i < 100; i++ {
		bv.Add(i)
	}

	bv.RemoveAll()
	if bv.Size() != 0 {
		t.Errorf("RemoveAll: expected size 0, got %d", bv.Size())
	}
}

func TestPodBVectorFreeTail(t *testing.T) {
	bv := NewPodBVector[int]()

	// Add enough elements to span multiple blocks
	for i := 0; i < 200; i++ {
		bv.Add(i)
	}

	originalBlocks := bv.numBlocks

	// Free tail to smaller size
	bv.FreeTail(50)

	if bv.Size() != 50 {
		t.Errorf("FreeTail: expected size 50, got %d", bv.Size())
	}

	// Should have fewer blocks now
	if bv.numBlocks >= originalBlocks {
		t.Errorf("Expected fewer blocks after FreeTail, had %d, now %d", originalBlocks, bv.numBlocks)
	}

	// Verify remaining elements
	for i := 0; i < bv.Size(); i++ {
		if bv.At(i) != i {
			t.Errorf("After FreeTail, At(%d): expected %d, got %d", i, i, bv.At(i))
		}
	}
}

func TestPodBVectorFreeAll(t *testing.T) {
	bv := NewPodBVector[int]()

	for i := 0; i < 100; i++ {
		bv.Add(i)
	}

	bv.FreeAll()

	if bv.Size() != 0 {
		t.Errorf("FreeAll: expected size 0, got %d", bv.Size())
	}
	if bv.numBlocks != 0 {
		t.Errorf("FreeAll: expected 0 blocks, got %d", bv.numBlocks)
	}
}

func TestPodBVectorAllocateContinuousBlock(t *testing.T) {
	scale := NewBlockScale(4) // Block size = 16
	bv := NewPodBVectorWithScale[int](scale)

	// Add some elements
	for i := 0; i < 10; i++ {
		bv.Add(i)
	}

	// Try to allocate a continuous block that fits in remaining space
	remainingInBlock := scale.Size - (bv.Size() & scale.Mask)
	if remainingInBlock > 1 {
		idx := bv.AllocateContinuousBlock(remainingInBlock - 1)
		if idx == -1 {
			t.Error("Should be able to allocate block that fits")
		}
		if idx != 10 {
			t.Errorf("Expected allocation at index 10, got %d", idx)
		}
	}

	// Try to allocate a block larger than block size
	idx := bv.AllocateContinuousBlock(scale.Size + 1)
	if idx != -1 {
		t.Error("Should not be able to allocate block larger than block size")
	}
}

func TestPodBVectorAddArray(t *testing.T) {
	bv := NewPodBVector[int]()

	data := []int{10, 20, 30, 40, 50}
	bv.AddArray(data)

	if bv.Size() != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), bv.Size())
	}

	for i, expected := range data {
		if bv.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, bv.At(i))
		}
	}
}

func TestPodBVectorAddData(t *testing.T) {
	bv := NewPodBVector[int]()

	data := []int{1, 2, 3, 4, 5}
	idx := 0

	accessor := func() (int, bool) {
		if idx < len(data) {
			val := data[idx]
			idx++
			return val, true
		}
		return 0, false
	}

	bv.AddData(accessor)

	if bv.Size() != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), bv.Size())
	}

	for i, expected := range data {
		if bv.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, bv.At(i))
		}
	}
}

func TestPodBVectorCutAt(t *testing.T) {
	bv := NewPodBVector[int]()

	for i := 0; i < 20; i++ {
		bv.Add(i)
	}

	bv.CutAt(10)
	if bv.Size() != 10 {
		t.Errorf("CutAt: expected size 10, got %d", bv.Size())
	}

	// CutAt with larger size should not change anything
	bv.CutAt(15)
	if bv.Size() != 10 {
		t.Errorf("CutAt with larger size: expected size 10, got %d", bv.Size())
	}
}

func TestPodBVectorCopy(t *testing.T) {
	bv1 := NewPodBVector[int]()

	for i := 0; i < 50; i++ {
		bv1.Add(i * 3)
	}

	bv2 := NewPodBVectorCopy(bv1)

	if bv2.Size() != bv1.Size() {
		t.Errorf("Copy: expected size %d, got %d", bv1.Size(), bv2.Size())
	}

	for i := 0; i < bv1.Size(); i++ {
		if bv2.At(i) != bv1.At(i) {
			t.Errorf("Copy: At(%d): expected %d, got %d", i, bv1.At(i), bv2.At(i))
		}
	}

	// Modify original, copy should be unchanged
	bv1.Set(10, 999)
	if bv2.At(10) == 999 {
		t.Error("Copy should be independent of original")
	}
}

func TestPodBVectorAssign(t *testing.T) {
	bv1 := NewPodBVector[int]()
	bv2 := NewPodBVector[int]()

	for i := 0; i < 30; i++ {
		bv1.Add(i + 100)
	}

	bv2.Assign(bv1)

	if bv2.Size() != bv1.Size() {
		t.Errorf("Assign: expected size %d, got %d", bv1.Size(), bv2.Size())
	}

	for i := 0; i < bv1.Size(); i++ {
		if bv2.At(i) != bv1.At(i) {
			t.Errorf("Assign: At(%d): expected %d, got %d", i, bv1.At(i), bv2.At(i))
		}
	}
}

func TestPodBVectorSerialization(t *testing.T) {
	bv := NewPodBVector[int32]()

	testData := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, val := range testData {
		bv.Add(val)
	}

	// Test serialization
	buffer := make([]byte, bv.ByteSize())
	bv.Serialize(buffer)

	// Test deserialization
	bv2 := NewPodBVector[int32]()
	bv2.Deserialize(buffer)

	if bv2.Size() != bv.Size() {
		t.Errorf("Deserialize: expected size %d, got %d", bv.Size(), bv2.Size())
	}

	for i := 0; i < bv.Size(); i++ {
		if bv2.At(i) != bv.At(i) {
			t.Errorf("Deserialize: At(%d): expected %d, got %d", i, bv.At(i), bv2.At(i))
		}
	}
}

func TestPodBVectorDeserializeAt(t *testing.T) {
	bv := NewPodBVector[int32]()

	// Add initial data
	for i := int32(0); i < 5; i++ {
		bv.Add(i)
	}

	// Create serialization data
	newData := []int32{100, 200, 300}
	buffer := make([]byte, len(newData)*4) // 4 bytes per int32
	for i, val := range newData {
		buffer[i*4] = byte(val)
		buffer[i*4+1] = byte(val >> 8)
		buffer[i*4+2] = byte(val >> 16)
		buffer[i*4+3] = byte(val >> 24)
	}

	// Deserialize at position 3
	bv.DeserializeAt(3, -1, buffer)

	// Should have grown to accommodate
	expectedSize := 3 + len(newData)
	if bv.Size() < expectedSize {
		t.Errorf("DeserializeAt: expected size at least %d, got %d", expectedSize, bv.Size())
	}

	// Check that positions 0-2 are preserved
	for i := 0; i < 3; i++ {
		if i < 5 && bv.At(i) != int32(i) {
			t.Errorf("DeserializeAt: original data at %d should be preserved", i)
		}
	}
}

func TestPodBVectorBounds(t *testing.T) {
	bv := NewPodBVector[int]()
	bv.Add(10)
	bv.Add(20)

	// Test valid access
	if bv.At(0) != 10 || bv.At(1) != 20 {
		t.Error("Valid access failed")
	}

	// Test out-of-bounds access
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for out-of-bounds access")
		}
	}()
	bv.At(5)
}
