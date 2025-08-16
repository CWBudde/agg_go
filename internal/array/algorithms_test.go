package array

import (
	"reflect"
	"testing"
)

func TestSwapElements(t *testing.T) {
	a := 10
	b := 20

	SwapElements(&a, &b)

	if a != 20 || b != 10 {
		t.Errorf("SwapElements failed: a=%d, b=%d", a, b)
	}
}

func TestQuickSortSlice(t *testing.T) {
	// Test with integers
	data := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	QuickSortSlice(data, IntLess)

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("QuickSort failed: got %v, expected %v", data, expected)
	}

	// Test with strings
	strings := []string{"banana", "apple", "cherry", "date"}
	expectedStrings := []string{"apple", "banana", "cherry", "date"}

	QuickSortSlice(strings, func(a, b string) bool { return a < b })

	if !reflect.DeepEqual(strings, expectedStrings) {
		t.Errorf("QuickSort strings failed: got %v, expected %v", strings, expectedStrings)
	}
}

func TestQuickSortArray(t *testing.T) {
	vec := NewPodVector[int]()
	data := []int{5, 2, 8, 1, 9, 3, 7, 4, 6}
	for _, v := range data {
		vec.Add(v)
	}

	QuickSort[int](vec, IntLess)

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i, exp := range expected {
		if vec.At(i) != exp {
			t.Errorf("QuickSort array failed at %d: got %d, expected %d", i, vec.At(i), exp)
		}
	}
}

func TestQuickSortEdgeCases(t *testing.T) {
	// Empty slice
	empty := []int{}
	QuickSortSlice(empty, IntLess)
	if len(empty) != 0 {
		t.Error("Empty slice should remain empty")
	}

	// Single element
	single := []int{42}
	QuickSortSlice(single, IntLess)
	if len(single) != 1 || single[0] != 42 {
		t.Error("Single element slice should remain unchanged")
	}

	// Already sorted
	sorted := []int{1, 2, 3, 4, 5}
	QuickSortSlice(sorted, IntLess)
	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(sorted, expected) {
		t.Error("Already sorted slice should remain sorted")
	}

	// Reverse sorted
	reverse := []int{5, 4, 3, 2, 1}
	QuickSortSlice(reverse, IntLess)
	expected = []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(reverse, expected) {
		t.Error("Reverse sorted slice should be sorted correctly")
	}

	// Duplicates
	duplicates := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	QuickSortSlice(duplicates, IntLess)
	expected = []int{1, 1, 2, 3, 3, 4, 5, 5, 6, 9}
	if !reflect.DeepEqual(duplicates, expected) {
		t.Errorf("Duplicates not sorted correctly: got %v, expected %v", duplicates, expected)
	}
}

func TestRemoveDuplicatesSlice(t *testing.T) {
	// Test with sorted slice containing duplicates
	data := []int{1, 1, 2, 3, 3, 3, 4, 5, 5}
	newLen := RemoveDuplicatesSlice(data, IntEqual)

	expected := []int{1, 2, 3, 4, 5}
	if newLen != len(expected) {
		t.Errorf("RemoveDuplicates length: got %d, expected %d", newLen, len(expected))
	}

	for i := 0; i < newLen; i++ {
		if data[i] != expected[i] {
			t.Errorf("RemoveDuplicates[%d]: got %d, expected %d", i, data[i], expected[i])
		}
	}

	// Test with no duplicates
	noDups := []int{1, 2, 3, 4, 5}
	newLen = RemoveDuplicatesSlice(noDups, IntEqual)
	if newLen != 5 {
		t.Errorf("No duplicates: expected length 5, got %d", newLen)
	}

	// Test with all same elements
	allSame := []int{3, 3, 3, 3, 3}
	newLen = RemoveDuplicatesSlice(allSame, IntEqual)
	if newLen != 1 {
		t.Errorf("All same: expected length 1, got %d", newLen)
	}
	if allSame[0] != 3 {
		t.Errorf("All same: expected first element 3, got %d", allSame[0])
	}
}

func TestRemoveDuplicatesArray(t *testing.T) {
	vec := NewPodVector[int]()
	data := []int{1, 1, 2, 3, 3, 3, 4, 5, 5}
	for _, v := range data {
		vec.Add(v)
	}

	newLen := RemoveDuplicates[int](vec, IntEqual)

	expected := []int{1, 2, 3, 4, 5}
	if newLen != len(expected) {
		t.Errorf("RemoveDuplicates array length: got %d, expected %d", newLen, len(expected))
	}

	for i := 0; i < newLen; i++ {
		if vec.At(i) != expected[i] {
			t.Errorf("RemoveDuplicates array[%d]: got %d, expected %d", i, vec.At(i), expected[i])
		}
	}
}

func TestInvertSlice(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	InvertSlice(data)

	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("InvertSlice failed: got %v, expected %v", data, expected)
	}

	// Test even length
	even := []int{1, 2, 3, 4}
	InvertSlice(even)
	expectedEven := []int{4, 3, 2, 1}
	if !reflect.DeepEqual(even, expectedEven) {
		t.Errorf("InvertSlice even failed: got %v, expected %v", even, expectedEven)
	}

	// Test single element
	single := []int{42}
	InvertSlice(single)
	if single[0] != 42 {
		t.Error("Single element should remain unchanged")
	}

	// Test empty
	empty := []int{}
	InvertSlice(empty)
	if len(empty) != 0 {
		t.Error("Empty slice should remain empty")
	}
}

func TestInvertContainer(t *testing.T) {
	vec := NewPodVector[int]()
	data := []int{1, 2, 3, 4, 5}
	for _, v := range data {
		vec.Add(v)
	}

	InvertContainer[int](vec)

	expected := []int{5, 4, 3, 2, 1}
	for i, exp := range expected {
		if vec.At(i) != exp {
			t.Errorf("InvertContainer[%d]: got %d, expected %d", i, vec.At(i), exp)
		}
	}
}

func TestBinarySearchPosSlice(t *testing.T) {
	data := []int{1, 3, 5, 7, 9, 11, 13}

	tests := []struct {
		value    int
		expected int
	}{
		{0, 0},  // Before first
		{1, 1},  // Equal to first
		{2, 1},  // Between first and second
		{7, 4},  // Equal to middle
		{8, 4},  // Between middle elements
		{13, 6}, // Equal to last
		{15, 7}, // After last
	}

	for _, test := range tests {
		pos := BinarySearchPosSlice(data, test.value, IntLess)
		if pos != test.expected {
			t.Errorf("BinarySearchPos(%d): got %d, expected %d", test.value, pos, test.expected)
		}
	}

	// Test empty slice
	empty := []int{}
	if BinarySearchPosSlice(empty, 5, IntLess) != 0 {
		t.Error("Binary search in empty slice should return 0")
	}

	// Test single element
	single := []int{5}
	if BinarySearchPosSlice(single, 3, IntLess) != 0 {
		t.Error("Binary search before single element should return 0")
	}
	if BinarySearchPosSlice(single, 7, IntLess) != 1 {
		t.Error("Binary search after single element should return 1")
	}
}

func TestBinarySearchPosArray(t *testing.T) {
	vec := NewPodVector[int]()
	data := []int{1, 3, 5, 7, 9, 11, 13}
	for _, v := range data {
		vec.Add(v)
	}

	pos := BinarySearchPos[int](vec, 6, IntLess)
	if pos != 3 {
		t.Errorf("BinarySearchPos array: got %d, expected 3", pos)
	}
}

func TestRangeAdaptor(t *testing.T) {
	vec := NewPodVector[int]()
	for i := 0; i < 10; i++ {
		vec.Add(i * 10)
	}

	// Create range adaptor for middle section
	ra := NewRangeAdaptor[int](vec, 3, 4)

	if ra.Size() != 4 {
		t.Errorf("RangeAdaptor size: got %d, expected 4", ra.Size())
	}

	// Test element access
	for i := 0; i < ra.Size(); i++ {
		expected := (i + 3) * 10
		if ra.At(i) != expected {
			t.Errorf("RangeAdaptor At(%d): got %d, expected %d", i, ra.At(i), expected)
		}
		if ra.ValueAt(i) != expected {
			t.Errorf("RangeAdaptor ValueAt(%d): got %d, expected %d", i, ra.ValueAt(i), expected)
		}
	}

	// Test element modification
	ra.Set(1, 999)
	if vec.At(4) != 999 {
		t.Error("RangeAdaptor Set should modify underlying array")
	}

	// Test bounds clamping
	ra2 := NewRangeAdaptor[int](vec, 8, 10) // Should clamp to size 2
	if ra2.Size() != 2 {
		t.Errorf("RangeAdaptor bounds clamp: got size %d, expected 2", ra2.Size())
	}

	// Test out-of-bounds bounds clamping
	ra3 := NewRangeAdaptor[int](vec, 15, 5) // Start beyond array
	if ra3.Size() != 0 {
		t.Errorf("RangeAdaptor out-of-bounds: got size %d, expected 0", ra3.Size())
	}
}

func TestSliceRangeAdaptor(t *testing.T) {
	data := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90}

	sra := NewSliceRangeAdaptor(data, 2, 5)

	if sra.Size() != 5 {
		t.Errorf("SliceRangeAdaptor size: got %d, expected 5", sra.Size())
	}

	// Test element access
	for i := 0; i < sra.Size(); i++ {
		expected := (i + 2) * 10
		if sra.At(i) != expected {
			t.Errorf("SliceRangeAdaptor At(%d): got %d, expected %d", i, sra.At(i), expected)
		}
	}

	// Test Data method
	rangeData := sra.Data()
	expected := []int{20, 30, 40, 50, 60}
	if !reflect.DeepEqual(rangeData, expected) {
		t.Errorf("SliceRangeAdaptor Data: got %v, expected %v", rangeData, expected)
	}
}

func TestUtilityFunctions(t *testing.T) {
	data := []int{1, 3, 5, 7, 9}

	// Test FindInSlice
	idx := FindInSlice(data, 5, IntEqual)
	if idx != 2 {
		t.Errorf("FindInSlice: got %d, expected 2", idx)
	}

	idx = FindInSlice(data, 6, IntEqual)
	if idx != -1 {
		t.Errorf("FindInSlice not found: got %d, expected -1", idx)
	}

	// Test ContainsInSlice
	if !ContainsInSlice(data, 7, IntEqual) {
		t.Error("ContainsInSlice should return true for existing element")
	}

	if ContainsInSlice(data, 8, IntEqual) {
		t.Error("ContainsInSlice should return false for non-existing element")
	}

	// Test IsSortedSlice
	if !IsSortedSlice(data, IntLess) {
		t.Error("IsSortedSlice should return true for sorted slice")
	}

	unsorted := []int{1, 5, 3, 7, 9}
	if IsSortedSlice(unsorted, IntLess) {
		t.Error("IsSortedSlice should return false for unsorted slice")
	}

	// Test UniqueSlice
	duplicates := []int{1, 1, 2, 3, 3, 4, 4, 4, 5}
	newLen := UniqueSlice(duplicates, IntEqual)

	expected := []int{1, 2, 3, 4, 5}
	if newLen != len(expected) {
		t.Errorf("UniqueSlice length: got %d, expected %d", newLen, len(expected))
	}

	for i := 0; i < newLen; i++ {
		if duplicates[i] != expected[i] {
			t.Errorf("UniqueSlice[%d]: got %d, expected %d", i, duplicates[i], expected[i])
		}
	}
}

func TestConvenienceSortFunctions(t *testing.T) {
	// Test SortInts
	ints := []int{5, 2, 8, 1, 9}
	SortInts(ints)
	expectedInts := []int{1, 2, 5, 8, 9}
	if !reflect.DeepEqual(ints, expectedInts) {
		t.Errorf("SortInts: got %v, expected %v", ints, expectedInts)
	}

	// Test SortIntsDescending
	ints2 := []int{5, 2, 8, 1, 9}
	SortIntsDescending(ints2)
	expectedDesc := []int{9, 8, 5, 2, 1}
	if !reflect.DeepEqual(ints2, expectedDesc) {
		t.Errorf("SortIntsDescending: got %v, expected %v", ints2, expectedDesc)
	}

	// Test SortStrings
	strings := []string{"banana", "apple", "cherry"}
	SortStrings(strings)
	expectedStrings := []string{"apple", "banana", "cherry"}
	if !reflect.DeepEqual(strings, expectedStrings) {
		t.Errorf("SortStrings: got %v, expected %v", strings, expectedStrings)
	}

	// Test IsSortedInts
	if !IsSortedInts(ints) {
		t.Error("IsSortedInts should return true for sorted slice")
	}

	unsorted := []int{3, 1, 4}
	if IsSortedInts(unsorted) {
		t.Error("IsSortedInts should return false for unsorted slice")
	}
}
