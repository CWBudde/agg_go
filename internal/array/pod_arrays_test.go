package array

import (
	"testing"
)

func TestPodArrayAdaptor(t *testing.T) {
	// Create test data
	data := []int{1, 2, 3, 4, 5}
	adaptor := NewPodArrayAdaptor(data)

	// Test size
	if adaptor.Size() != 5 {
		t.Errorf("Expected size 5, got %d", adaptor.Size())
	}

	// Test element access
	for i := 0; i < adaptor.Size(); i++ {
		if adaptor.At(i) != data[i] {
			t.Errorf("At(%d): expected %d, got %d", i, data[i], adaptor.At(i))
		}
		if adaptor.ValueAt(i) != data[i] {
			t.Errorf("ValueAt(%d): expected %d, got %d", i, data[i], adaptor.ValueAt(i))
		}
	}

	// Test element modification
	adaptor.Set(2, 99)
	if adaptor.At(2) != 99 {
		t.Errorf("Set(2, 99): expected 99, got %d", adaptor.At(2))
	}
	if data[2] != 99 {
		t.Errorf("Expected underlying data to be modified, got %d", data[2])
	}

	// Test bounds checking
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for out-of-bounds access")
		}
	}()
	adaptor.At(10)
}

func TestPodAutoArray(t *testing.T) {
	// Test creation
	arr := NewPodAutoArray[int](5)

	if arr.Size() != 5 {
		t.Errorf("Expected size 5, got %d", arr.Size())
	}

	// Test element access and modification
	for i := 0; i < arr.Size(); i++ {
		arr.Set(i, i*10)
	}

	for i := 0; i < arr.Size(); i++ {
		expected := i * 10
		if arr.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, arr.At(i))
		}
	}

	// Test creation from data
	sourceData := []int{10, 20, 30}
	arr2 := NewPodAutoArrayFrom(sourceData)

	if arr2.Size() != 3 {
		t.Errorf("Expected size 3, got %d", arr2.Size())
	}

	for i := 0; i < arr2.Size(); i++ {
		if arr2.At(i) != sourceData[i] {
			t.Errorf("At(%d): expected %d, got %d", i, sourceData[i], arr2.At(i))
		}
	}

	// Test assignment
	newData := []int{100, 200}
	arr2.Assign(newData)

	// Should have copied first 2 elements and zeroed the rest
	if arr2.At(0) != 100 || arr2.At(1) != 200 || arr2.At(2) != 0 {
		t.Errorf("Assign failed: got [%d, %d, %d]", arr2.At(0), arr2.At(1), arr2.At(2))
	}

	// Test Data() method
	data := arr2.Data()
	if len(data) != 3 {
		t.Errorf("Data(): expected length 3, got %d", len(data))
	}
}

func TestPodAutoVector(t *testing.T) {
	vec := NewPodAutoVector[int](10)

	// Test initial state
	if vec.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", vec.Size())
	}
	if vec.Capacity() != 10 {
		t.Errorf("Expected capacity 10, got %d", vec.Capacity())
	}

	// Test adding elements
	for i := 0; i < 5; i++ {
		vec.Add(i * 2)
	}

	if vec.Size() != 5 {
		t.Errorf("Expected size 5, got %d", vec.Size())
	}

	for i := 0; i < vec.Size(); i++ {
		expected := i * 2
		if vec.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, vec.At(i))
		}
	}

	// Test PushBack
	vec.PushBack(100)
	if vec.Size() != 6 || vec.At(5) != 100 {
		t.Error("PushBack failed")
	}

	// Test IncSize
	vec.IncSize(2)
	if vec.Size() != 8 {
		t.Errorf("IncSize: expected size 8, got %d", vec.Size())
	}

	// Test Clear
	vec.Clear()
	if vec.Size() != 0 {
		t.Errorf("Clear: expected size 0, got %d", vec.Size())
	}

	// Test capacity overflow
	vec = NewPodAutoVector[int](2)
	vec.Add(1)
	vec.Add(2)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for capacity overflow")
		}
	}()
	vec.Add(3)
}

func TestPodArray(t *testing.T) {
	// Test empty creation
	arr := NewPodArray[int]()
	if arr.Size() != 0 {
		t.Errorf("Expected size 0, got %d", arr.Size())
	}

	// Test creation with size
	arr = NewPodArrayWithSize[int](5)
	if arr.Size() != 5 {
		t.Errorf("Expected size 5, got %d", arr.Size())
	}

	// Test element access and modification
	for i := 0; i < arr.Size(); i++ {
		arr.Set(i, i+100)
	}

	for i := 0; i < arr.Size(); i++ {
		expected := i + 100
		if arr.At(i) != expected {
			t.Errorf("At(%d): expected %d, got %d", i, expected, arr.At(i))
		}
	}

	// Test resize
	arr.Resize(10)
	if arr.Size() != 10 {
		t.Errorf("Resize: expected size 10, got %d", arr.Size())
	}

	// Check that original data is preserved
	for i := 0; i < 5; i++ {
		expected := i + 100
		if arr.At(i) != expected {
			t.Errorf("After resize, At(%d): expected %d, got %d", i, expected, arr.At(i))
		}
	}

	// Test copy constructor
	arr2 := NewPodArrayCopy(arr)
	if arr2.Size() != arr.Size() {
		t.Errorf("Copy: expected size %d, got %d", arr.Size(), arr2.Size())
	}

	for i := 0; i < arr.Size(); i++ {
		if arr2.At(i) != arr.At(i) {
			t.Errorf("Copy: At(%d): expected %d, got %d", i, arr.At(i), arr2.At(i))
		}
	}

	// Test assignment
	arr3 := NewPodArray[int]()
	arr3.Assign(arr)

	if arr3.Size() != arr.Size() {
		t.Errorf("Assign: expected size %d, got %d", arr.Size(), arr3.Size())
	}

	// Test Data() method
	data := arr.Data()
	if len(data) != arr.Size() {
		t.Errorf("Data(): expected length %d, got %d", arr.Size(), len(data))
	}
}

func TestPodVector(t *testing.T) {
	vec := NewPodVector[int]()

	// Test initial state
	if vec.Size() != 0 || vec.Capacity() != 0 {
		t.Error("Expected empty vector")
	}

	// Test adding elements (auto-grow)
	for i := 0; i < 20; i++ {
		vec.Add(i)
	}

	if vec.Size() != 20 {
		t.Errorf("Expected size 20, got %d", vec.Size())
	}

	for i := 0; i < vec.Size(); i++ {
		if vec.At(i) != i {
			t.Errorf("At(%d): expected %d, got %d", i, i, vec.At(i))
		}
	}

	// Test with initial capacity
	vec2 := NewPodVectorWithCapacity[int](10, 5)
	if vec2.Capacity() != 15 {
		t.Errorf("Expected capacity 15, got %d", vec2.Capacity())
	}

	// Test allocate
	vec2.Allocate(8, 2)
	if vec2.Size() != 8 {
		t.Errorf("Allocate: expected size 8, got size %d", vec2.Size())
	}
	// Capacity should be at least 10 (might be higher if it was already higher)
	if vec2.Capacity() < 10 {
		t.Errorf("Allocate: expected capacity at least 10, got %d", vec2.Capacity())
	}

	// Test resize
	vec.Resize(30)
	if vec.Size() != 30 {
		t.Errorf("Resize: expected size 30, got %d", vec.Size())
	}

	// Test InsertAt
	vec.InsertAt(5, 999)
	if vec.At(5) != 999 {
		t.Errorf("InsertAt: expected 999 at position 5, got %d", vec.At(5))
	}

	// Test CutAt
	vec.CutAt(10)
	if vec.Size() != 10 {
		t.Errorf("CutAt: expected size 10, got %d", vec.Size())
	}

	// Test Zero
	vec.Zero()
	for i := 0; i < vec.Size(); i++ {
		if vec.At(i) != 0 {
			t.Errorf("Zero: expected 0 at position %d, got %d", i, vec.At(i))
		}
	}

	// Test copy constructor
	vec3 := NewPodVectorCopy(vec)
	if vec3.Size() != vec.Size() || vec3.Capacity() != vec.Capacity() {
		t.Error("Copy constructor failed")
	}

	// Test ByteSize
	if vec.ByteSize() != vec.Size()*8 { // int is 8 bytes on 64-bit
		t.Errorf("ByteSize: expected %d, got %d", vec.Size()*8, vec.ByteSize())
	}
}

func TestPodVectorSerialization(t *testing.T) {
	vec := NewPodVector[int32]()

	// Add test data
	testData := []int32{1, 2, 3, 4, 5}
	for _, val := range testData {
		vec.Add(val)
	}

	// Test serialization
	buffer := make([]byte, vec.ByteSize())
	vec.Serialize(buffer)

	// Test deserialization
	vec2 := NewPodVector[int32]()
	vec2.Deserialize(buffer)

	if vec2.Size() != vec.Size() {
		t.Errorf("Deserialize: expected size %d, got %d", vec.Size(), vec2.Size())
	}

	for i := 0; i < vec.Size(); i++ {
		if vec2.At(i) != vec.At(i) {
			t.Errorf("Deserialize: At(%d): expected %d, got %d", i, vec.At(i), vec2.At(i))
		}
	}
}

func TestPodVectorBounds(t *testing.T) {
	vec := NewPodVectorWithCapacity[int](5, 0)
	vec.Add(10)
	vec.Add(20)

	// Test valid access
	if vec.At(0) != 10 || vec.At(1) != 20 {
		t.Error("Valid access failed")
	}

	// Test out-of-bounds access
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for out-of-bounds access")
		}
	}()
	vec.At(5)
}
