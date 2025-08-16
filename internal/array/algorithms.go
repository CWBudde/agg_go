package array

// No imports needed - algorithms are self-contained

// QuickSortThreshold defines the threshold below which insertion sort is used.
const QuickSortThreshold = 9

// SwapElements swaps two elements.
// This is equivalent to AGG's swap_elements<T> function.
func SwapElements[T any](a, b *T) {
	temp := *a
	*a = *b
	*b = temp
}

// LessFunc is a function type for comparison operations.
type LessFunc[T any] func(a, b T) bool

// EqualFunc is a function type for equality operations.
type EqualFunc[T any] func(a, b T) bool

// QuickSort performs hybrid quicksort with insertion sort for small arrays.
// This is equivalent to AGG's quick_sort<Array, Less> function.
func QuickSort[T any](arr ArrayInterface[T], less LessFunc[T]) {
	if arr.Size() < 2 {
		return
	}

	// Convert to slice for easier manipulation
	data := make([]T, arr.Size())
	for i := 0; i < arr.Size(); i++ {
		data[i] = arr.At(i)
	}

	// Perform quicksort on slice
	quickSortSlice(data, less)

	// Copy back to array
	for i := 0; i < arr.Size(); i++ {
		arr.Set(i, data[i])
	}
}

// QuickSortSlice performs quicksort on a regular Go slice.
func QuickSortSlice[T any](arr []T, less LessFunc[T]) {
	quickSortSlice(arr, less)
}

// quickSortSlice is the internal implementation of quicksort.
func quickSortSlice[T any](arr []T, less LessFunc[T]) {
	if len(arr) < 2 {
		return
	}

	type stackFrame struct {
		base  int
		limit int
	}

	stack := make([]stackFrame, 0, 80)
	base := 0
	limit := len(arr)

	for {
		length := limit - base

		if length > QuickSortThreshold {
			// Use quicksort for larger subarrays
			pivot := base + length/2
			SwapElements(&arr[base], &arr[pivot])

			i := base + 1
			j := limit - 1

			// Ensure arr[j] <= arr[i] <= arr[base]
			if less(arr[j], arr[i]) {
				SwapElements(&arr[i], &arr[j])
			}
			if less(arr[base], arr[i]) {
				SwapElements(&arr[base], &arr[i])
			}
			if less(arr[j], arr[base]) {
				SwapElements(&arr[j], &arr[base])
			}

			// Partition
			for {
				for i++; less(arr[i], arr[base]); i++ {
				}
				for j--; less(arr[base], arr[j]); j-- {
				}

				if i > j {
					break
				}

				SwapElements(&arr[i], &arr[j])
			}

			SwapElements(&arr[base], &arr[j])

			// Push the larger subarray onto stack
			if j-base > limit-i {
				stack = append(stack, stackFrame{base, j})
				base = i
			} else {
				stack = append(stack, stackFrame{i, limit})
				limit = j
			}
		} else {
			// Use insertion sort for small subarrays
			for i := base + 1; i < limit; i++ {
				j := i
				for j > base && less(arr[j], arr[j-1]) {
					SwapElements(&arr[j], &arr[j-1])
					j--
				}
			}

			// Pop from stack
			if len(stack) > 0 {
				frame := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				base = frame.base
				limit = frame.limit
			} else {
				break
			}
		}
	}
}

// RemoveDuplicates removes duplicates from a sorted array.
// Returns the number of remaining elements.
// This is equivalent to AGG's remove_duplicates<Array, Equal> function.
func RemoveDuplicates[T any](arr ArrayInterface[T], equal EqualFunc[T]) int {
	if arr.Size() < 2 {
		return arr.Size()
	}

	j := 1
	for i := 1; i < arr.Size(); i++ {
		if !equal(arr.At(i), arr.At(i-1)) {
			arr.Set(j, arr.At(i))
			j++
		}
	}

	return j
}

// RemoveDuplicatesSlice removes duplicates from a sorted slice.
func RemoveDuplicatesSlice[T any](arr []T, equal EqualFunc[T]) int {
	if len(arr) < 2 {
		return len(arr)
	}

	j := 1
	for i := 1; i < len(arr); i++ {
		if !equal(arr[i], arr[i-1]) {
			arr[j] = arr[i]
			j++
		}
	}

	return j
}

// InvertContainer reverses the elements in an array.
// This is equivalent to AGG's invert_container<Array> function.
func InvertContainer[T any](arr ArrayInterface[T]) {
	i := 0
	j := arr.Size() - 1
	for i < j {
		a := arr.At(i)
		b := arr.At(j)
		arr.Set(i, b)
		arr.Set(j, a)
		i++
		j--
	}
}

// InvertSlice reverses the elements in a slice.
func InvertSlice[T any](arr []T) {
	i := 0
	j := len(arr) - 1
	for i < j {
		SwapElements(&arr[i], &arr[j])
		i++
		j--
	}
}

// BinarySearchPos finds the position where a value should be inserted in a sorted array.
// This is equivalent to AGG's binary_search_pos<Array, Value, Less> function.
func BinarySearchPos[T any](arr ArrayInterface[T], val T, less LessFunc[T]) int {
	if arr.Size() == 0 {
		return 0
	}

	beg := 0
	end := arr.Size() - 1

	if less(val, arr.At(0)) {
		return 0
	}
	if less(arr.At(end), val) {
		return end + 1
	}

	for end-beg > 1 {
		mid := (end + beg) >> 1
		if less(val, arr.At(mid)) {
			end = mid
		} else {
			beg = mid
		}
	}

	return end
}

// BinarySearchPosSlice finds the position where a value should be inserted in a sorted slice.
func BinarySearchPosSlice[T any](arr []T, val T, less LessFunc[T]) int {
	if len(arr) == 0 {
		return 0
	}

	beg := 0
	end := len(arr) - 1

	if less(val, arr[0]) {
		return 0
	}
	if less(arr[end], val) {
		return len(arr)
	}

	for end-beg > 1 {
		mid := (end + beg) >> 1
		if less(val, arr[mid]) {
			end = mid
		} else {
			beg = mid
		}
	}

	return end
}

// RangeAdaptor provides a view into a subset of an array.
// This is equivalent to AGG's range_adaptor<Array> template class.
type RangeAdaptor[T any] struct {
	array ArrayInterface[T]
	start int
	size  int
}

// NewRangeAdaptor creates a new range adaptor.
func NewRangeAdaptor[T any](array ArrayInterface[T], start, size int) *RangeAdaptor[T] {
	// Clamp start and size to valid bounds
	if start < 0 {
		start = 0
	}
	if start >= array.Size() {
		start = array.Size()
		size = 0
	} else if start+size > array.Size() {
		size = array.Size() - start
	}
	if size < 0 {
		size = 0
	}

	return &RangeAdaptor[T]{
		array: array,
		start: start,
		size:  size,
	}
}

// Size returns the size of the range.
func (ra *RangeAdaptor[T]) Size() int {
	return ra.size
}

// At returns the element at the specified index with bounds checking.
func (ra *RangeAdaptor[T]) At(i int) T {
	if i < 0 || i >= ra.size {
		panic("index out of bounds")
	}
	return ra.array.At(ra.start + i)
}

// Set sets the element at the specified index with bounds checking.
func (ra *RangeAdaptor[T]) Set(i int, v T) {
	if i < 0 || i >= ra.size {
		panic("index out of bounds")
	}
	ra.array.Set(ra.start+i, v)
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (ra *RangeAdaptor[T]) ValueAt(i int) T {
	return ra.array.ValueAt(ra.start + i)
}

// SliceRangeAdaptor provides a view into a subset of a slice.
type SliceRangeAdaptor[T any] struct {
	slice []T
	start int
	size  int
}

// NewSliceRangeAdaptor creates a new slice range adaptor.
func NewSliceRangeAdaptor[T any](slice []T, start, size int) *SliceRangeAdaptor[T] {
	// Clamp start and size to valid bounds
	if start < 0 {
		start = 0
	}
	if start >= len(slice) {
		start = len(slice)
		size = 0
	} else if start+size > len(slice) {
		size = len(slice) - start
	}
	if size < 0 {
		size = 0
	}

	return &SliceRangeAdaptor[T]{
		slice: slice,
		start: start,
		size:  size,
	}
}

// Size returns the size of the range.
func (sra *SliceRangeAdaptor[T]) Size() int {
	return sra.size
}

// At returns the element at the specified index with bounds checking.
func (sra *SliceRangeAdaptor[T]) At(i int) T {
	if i < 0 || i >= sra.size {
		panic("index out of bounds")
	}
	return sra.slice[sra.start+i]
}

// Set sets the element at the specified index with bounds checking.
func (sra *SliceRangeAdaptor[T]) Set(i int, v T) {
	if i < 0 || i >= sra.size {
		panic("index out of bounds")
	}
	sra.slice[sra.start+i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (sra *SliceRangeAdaptor[T]) ValueAt(i int) T {
	return sra.slice[sra.start+i]
}

// Data returns the underlying slice view.
func (sra *SliceRangeAdaptor[T]) Data() []T {
	return sra.slice[sra.start : sra.start+sra.size]
}

// Utility functions for working with slices

// FindInSlice finds the first occurrence of a value in a slice.
func FindInSlice[T any](slice []T, val T, equal EqualFunc[T]) int {
	for i, elem := range slice {
		if equal(elem, val) {
			return i
		}
	}
	return -1
}

// ContainsInSlice checks if a slice contains a value.
func ContainsInSlice[T any](slice []T, val T, equal EqualFunc[T]) bool {
	return FindInSlice(slice, val, equal) >= 0
}

// IsSortedSlice checks if a slice is sorted according to the given comparison function.
func IsSortedSlice[T any](slice []T, less LessFunc[T]) bool {
	for i := 1; i < len(slice); i++ {
		if less(slice[i], slice[i-1]) {
			return false
		}
	}
	return true
}

// UniqueSlice removes consecutive duplicates from a slice (modifies in place).
// Returns the new length. The slice should be sorted first for best results.
func UniqueSlice[T any](slice []T, equal EqualFunc[T]) int {
	if len(slice) <= 1 {
		return len(slice)
	}

	writeIdx := 1
	for readIdx := 1; readIdx < len(slice); readIdx++ {
		if !equal(slice[readIdx], slice[writeIdx-1]) {
			slice[writeIdx] = slice[readIdx]
			writeIdx++
		}
	}

	return writeIdx
}
