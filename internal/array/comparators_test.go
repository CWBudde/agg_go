package array

import (
	"testing"
)

func TestBasicComparators(t *testing.T) {
	// Test IntLess and IntGreater
	if !IntLess(5, 10) {
		t.Error("IntLess(5, 10) should be true")
	}
	if IntLess(10, 5) {
		t.Error("IntLess(10, 5) should be false")
	}
	if IntLess(5, 5) {
		t.Error("IntLess(5, 5) should be false")
	}

	if !IntGreater(10, 5) {
		t.Error("IntGreater(10, 5) should be true")
	}
	if IntGreater(5, 10) {
		t.Error("IntGreater(5, 10) should be false")
	}

	// Test UnsignedLess and UnsignedGreater
	if !UnsignedLess(uint(5), uint(10)) {
		t.Error("UnsignedLess(5, 10) should be true")
	}
	if UnsignedGreater(uint(5), uint(10)) {
		t.Error("UnsignedGreater(5, 10) should be false")
	}
}

func TestTypedComparators(t *testing.T) {
	// Test various typed comparators
	if !Int32Less(5, 10) {
		t.Error("Int32Less(5, 10) should be true")
	}
	if !Int64Less(100, 200) {
		t.Error("Int64Less(100, 200) should be true")
	}
	if !Float32Less(1.5, 2.5) {
		t.Error("Float32Less(1.5, 2.5) should be true")
	}
	if !Float64Less(1.5, 2.5) {
		t.Error("Float64Less(1.5, 2.5) should be true")
	}
}

func TestEqualityComparators(t *testing.T) {
	// Test integer equality
	if !IntEqual(5, 5) {
		t.Error("IntEqual(5, 5) should be true")
	}
	if IntEqual(5, 10) {
		t.Error("IntEqual(5, 10) should be false")
	}

	// Test unsigned equality
	if !UnsignedEqual(uint(5), uint(5)) {
		t.Error("UnsignedEqual(5, 5) should be true")
	}

	// Test float equality
	if !Float32Equal(1.5, 1.5) {
		t.Error("Float32Equal(1.5, 1.5) should be true")
	}
	if !Float64Equal(2.5, 2.5) {
		t.Error("Float64Equal(2.5, 2.5) should be true")
	}
}

func TestFloatEpsilonComparators(t *testing.T) {
	// Test float32 epsilon comparison
	if !Float32EqualEps(1.0, 1.000001, 0.001) {
		t.Error("Float32EqualEps should be true within epsilon")
	}
	if Float32EqualEps(1.0, 1.1, 0.001) {
		t.Error("Float32EqualEps should be false outside epsilon")
	}

	// Test float64 epsilon comparison
	if !Float64EqualEps(1.0, 1.000000001, 0.001) {
		t.Error("Float64EqualEps should be true within epsilon")
	}
	if Float64EqualEps(1.0, 1.1, 0.001) {
		t.Error("Float64EqualEps should be false outside epsilon")
	}
}

func TestComparatorInterfaces(t *testing.T) {
	// Test IntComparator
	intComp := IntComparator{}
	if !intComp.Less(5, 10) {
		t.Error("IntComparator.Less(5, 10) should be true")
	}
	if !intComp.Equal(5, 5) {
		t.Error("IntComparator.Equal(5, 5) should be true")
	}
	if intComp.Equal(5, 10) {
		t.Error("IntComparator.Equal(5, 10) should be false")
	}

	// Test UintComparator
	uintComp := UintComparator{}
	if !uintComp.Less(uint(5), uint(10)) {
		t.Error("UintComparator.Less(5, 10) should be true")
	}
	if !uintComp.Equal(uint(5), uint(5)) {
		t.Error("UintComparator.Equal(5, 5) should be true")
	}

	// Test Float64Comparator
	floatComp := NewFloat64Comparator(0.001)
	if !floatComp.Less(1.5, 2.5) {
		t.Error("Float64Comparator.Less(1.5, 2.5) should be true")
	}
	if !floatComp.Equal(1.0, 1.0005) {
		t.Error("Float64Comparator.Equal should be true within epsilon")
	}
	if floatComp.Equal(1.0, 1.1) {
		t.Error("Float64Comparator.Equal should be false outside epsilon")
	}

	// Test StringComparator
	stringComp := StringComparator{}
	if !stringComp.Less("apple", "banana") {
		t.Error("StringComparator.Less should work lexicographically")
	}
	if !stringComp.Equal("hello", "hello") {
		t.Error("StringComparator.Equal should work for identical strings")
	}
}

func TestReverseComparator(t *testing.T) {
	intComp := IntComparator{}
	reverseComp := NewReverseComparator(intComp)

	// Reverse should invert the less comparison
	if !reverseComp.Less(10, 5) {
		t.Error("ReverseComparator.Less(10, 5) should be true (reversed)")
	}
	if reverseComp.Less(5, 10) {
		t.Error("ReverseComparator.Less(5, 10) should be false (reversed)")
	}

	// Equal should remain the same
	if !reverseComp.Equal(5, 5) {
		t.Error("ReverseComparator.Equal should not change")
	}
}

func TestFunctionCreators(t *testing.T) {
	intComp := IntComparator{}

	lessFunc := LessFuncFromComparator(intComp)
	equalFunc := EqualFuncFromComparator(intComp)

	if !lessFunc(5, 10) {
		t.Error("LessFunc from comparator should work")
	}
	if !equalFunc(5, 5) {
		t.Error("EqualFunc from comparator should work")
	}
}

func TestGlobalComparators(t *testing.T) {
	// Test that global comparators are initialized
	if DefaultIntComparator.Less(10, 5) {
		t.Error("DefaultIntComparator should work correctly")
	}
	if DefaultUintComparator.Less(uint(10), uint(5)) {
		t.Error("DefaultUintComparator should work correctly")
	}
	if DefaultFloat64Comparator.Less(10.0, 5.0) {
		t.Error("DefaultFloat64Comparator should work correctly")
	}
	if DefaultStringComparator.Less("zebra", "apple") {
		t.Error("DefaultStringComparator should work correctly")
	}
}

func TestComparatorConvenienceFunctions(t *testing.T) {
	// Test SortInts
	ints := []int{3, 1, 4, 1, 5}
	SortInts(ints)
	expected := []int{1, 1, 3, 4, 5}
	for i, v := range expected {
		if ints[i] != v {
			t.Errorf("SortInts failed at index %d: got %d, expected %d", i, ints[i], v)
		}
	}

	// Test SortIntsDescending
	ints2 := []int{3, 1, 4, 1, 5}
	SortIntsDescending(ints2)
	expectedDesc := []int{5, 4, 3, 1, 1}
	for i, v := range expectedDesc {
		if ints2[i] != v {
			t.Errorf("SortIntsDescending failed at index %d: got %d, expected %d", i, ints2[i], v)
		}
	}

	// Test SortUints
	uints := []uint{3, 1, 4, 1, 5}
	SortUints(uints)
	expectedUints := []uint{1, 1, 3, 4, 5}
	for i, v := range expectedUints {
		if uints[i] != v {
			t.Errorf("SortUints failed at index %d: got %d, expected %d", i, uints[i], v)
		}
	}

	// Test SortFloat64s
	floats := []float64{3.1, 1.2, 4.3, 1.4, 5.5}
	SortFloat64s(floats)
	if floats[0] != 1.2 || floats[1] != 1.4 || floats[4] != 5.5 {
		t.Error("SortFloat64s failed")
	}

	// Test SortStrings
	strings := []string{"cherry", "apple", "banana"}
	SortStrings(strings)
	expectedStrings := []string{"apple", "banana", "cherry"}
	for i, v := range expectedStrings {
		if strings[i] != v {
			t.Errorf("SortStrings failed at index %d: got %s, expected %s", i, strings[i], v)
		}
	}
}

func TestIsSortedFunctions(t *testing.T) {
	// Test IsSortedInts
	sortedInts := []int{1, 2, 3, 4, 5}
	if !IsSortedInts(sortedInts) {
		t.Error("IsSortedInts should return true for sorted slice")
	}

	unsortedInts := []int{1, 3, 2, 4, 5}
	if IsSortedInts(unsortedInts) {
		t.Error("IsSortedInts should return false for unsorted slice")
	}

	// Test IsSortedUints
	sortedUints := []uint{1, 2, 3, 4, 5}
	if !IsSortedUints(sortedUints) {
		t.Error("IsSortedUints should return true for sorted slice")
	}

	// Test IsSortedFloat64s
	sortedFloats := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	if !IsSortedFloat64s(sortedFloats) {
		t.Error("IsSortedFloat64s should return true for sorted slice")
	}

	// Test IsSortedStrings
	sortedStrings := []string{"apple", "banana", "cherry"}
	if !IsSortedStrings(sortedStrings) {
		t.Error("IsSortedStrings should return true for sorted slice")
	}

	unsortedStrings := []string{"banana", "apple", "cherry"}
	if IsSortedStrings(unsortedStrings) {
		t.Error("IsSortedStrings should return false for unsorted slice")
	}
}
