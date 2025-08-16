package array

import (
	"fmt"
	"unsafe"

	"agg_go/internal/basics"
)

// PodArrayAdaptor wraps an existing slice with AGG-compatible interface.
// This is equivalent to AGG's pod_array_adaptor<T> template class.
type PodArrayAdaptor[T any] struct {
	array []T
	size  int
}

// NewPodArrayAdaptor creates a new array adaptor wrapping the provided slice.
func NewPodArrayAdaptor[T any](array []T) *PodArrayAdaptor[T] {
	return &PodArrayAdaptor[T]{
		array: array,
		size:  len(array),
	}
}

// Size returns the number of elements in the array.
func (pa *PodArrayAdaptor[T]) Size() int {
	return pa.size
}

// At returns the element at the specified index with bounds checking.
func (pa *PodArrayAdaptor[T]) At(i int) T {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	return pa.array[i]
}

// Set sets the element at the specified index with bounds checking.
func (pa *PodArrayAdaptor[T]) Set(i int, v T) {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	pa.array[i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (pa *PodArrayAdaptor[T]) ValueAt(i int) T {
	return pa.array[i]
}

// PodAutoArray is a fixed-size array with compile-time size.
// This is equivalent to AGG's pod_auto_array<T, Size> template class.
// In Go, we simulate the compile-time size with a runtime size parameter.
type PodAutoArray[T any] struct {
	array []T
	size  int
}

// NewPodAutoArray creates a new fixed-size array with the specified size.
func NewPodAutoArray[T any](size int) *PodAutoArray[T] {
	return &PodAutoArray[T]{
		array: make([]T, size),
		size:  size,
	}
}

// NewPodAutoArrayFrom creates a new fixed-size array initialized from a slice.
func NewPodAutoArrayFrom[T any](data []T) *PodAutoArray[T] {
	arr := &PodAutoArray[T]{
		array: make([]T, len(data)),
		size:  len(data),
	}
	copy(arr.array, data)
	return arr
}

// Size returns the fixed size of the array.
func (pa *PodAutoArray[T]) Size() int {
	return pa.size
}

// At returns the element at the specified index with bounds checking.
func (pa *PodAutoArray[T]) At(i int) T {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	return pa.array[i]
}

// Set sets the element at the specified index with bounds checking.
func (pa *PodAutoArray[T]) Set(i int, v T) {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	pa.array[i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (pa *PodAutoArray[T]) ValueAt(i int) T {
	return pa.array[i]
}

// Assign copies data from the provided slice.
func (pa *PodAutoArray[T]) Assign(data []T) {
	copyLen := basics.IMin(len(data), pa.size)
	copy(pa.array[:copyLen], data[:copyLen])

	// Zero out remaining elements if data is shorter
	var zero T
	for i := copyLen; i < pa.size; i++ {
		pa.array[i] = zero
	}
}

// Data returns the underlying slice (read-only access).
func (pa *PodAutoArray[T]) Data() []T {
	return pa.array[:pa.size]
}

// PodAutoVector is a fixed-capacity vector with dynamic size.
// This is equivalent to AGG's pod_auto_vector<T, Size> template class.
type PodAutoVector[T any] struct {
	array []T
	size  int
	cap   int
}

// NewPodAutoVector creates a new auto vector with the specified capacity.
func NewPodAutoVector[T any](capacity int) *PodAutoVector[T] {
	return &PodAutoVector[T]{
		array: make([]T, capacity),
		size:  0,
		cap:   capacity,
	}
}

// Size returns the current number of elements.
func (pv *PodAutoVector[T]) Size() int {
	return pv.size
}

// Capacity returns the maximum capacity.
func (pv *PodAutoVector[T]) Capacity() int {
	return pv.cap
}

// At returns the element at the specified index with bounds checking.
func (pv *PodAutoVector[T]) At(i int) T {
	if i < 0 || i >= pv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pv.size))
	}
	return pv.array[i]
}

// Set sets the element at the specified index with bounds checking.
func (pv *PodAutoVector[T]) Set(i int, v T) {
	if i < 0 || i >= pv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pv.size))
	}
	pv.array[i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (pv *PodAutoVector[T]) ValueAt(i int) T {
	return pv.array[i]
}

// Add appends an element to the vector.
func (pv *PodAutoVector[T]) Add(v T) {
	if pv.size >= pv.cap {
		panic(fmt.Sprintf("capacity exceeded: %d >= %d", pv.size, pv.cap))
	}
	pv.array[pv.size] = v
	pv.size++
}

// PushBack appends an element to the vector (equivalent to Add).
func (pv *PodAutoVector[T]) PushBack(v T) {
	pv.Add(v)
}

// IncSize increases the size by the specified amount.
func (pv *PodAutoVector[T]) IncSize(size int) {
	newSize := pv.size + size
	if newSize > pv.cap {
		panic(fmt.Sprintf("capacity exceeded: %d > %d", newSize, pv.cap))
	}
	pv.size = newSize
}

// RemoveAll clears all elements.
func (pv *PodAutoVector[T]) RemoveAll() {
	pv.size = 0
}

// Clear clears all elements (equivalent to RemoveAll).
func (pv *PodAutoVector[T]) Clear() {
	pv.size = 0
}

// Data returns a slice view of the current elements.
func (pv *PodAutoVector[T]) Data() []T {
	return pv.array[:pv.size]
}

// PodArray is a dynamic array with explicit memory management.
// This is equivalent to AGG's pod_array<T> template class.
type PodArray[T any] struct {
	array []T
	size  int
}

// NewPodArray creates a new dynamic array.
func NewPodArray[T any]() *PodArray[T] {
	return &PodArray[T]{
		array: nil,
		size:  0,
	}
}

// NewPodArrayWithSize creates a new dynamic array with the specified size.
func NewPodArrayWithSize[T any](size int) *PodArray[T] {
	var array []T
	if size > 0 {
		array = make([]T, size)
	}
	return &PodArray[T]{
		array: array,
		size:  size,
	}
}

// NewPodArrayCopy creates a new dynamic array as a copy of another.
func NewPodArrayCopy[T any](other *PodArray[T]) *PodArray[T] {
	if other == nil || other.size == 0 {
		return NewPodArray[T]()
	}

	arr := &PodArray[T]{
		array: make([]T, other.size),
		size:  other.size,
	}
	copy(arr.array, other.array[:other.size])
	return arr
}

// Size returns the number of elements.
func (pa *PodArray[T]) Size() int {
	return pa.size
}

// At returns the element at the specified index with bounds checking.
func (pa *PodArray[T]) At(i int) T {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	return pa.array[i]
}

// Set sets the element at the specified index with bounds checking.
func (pa *PodArray[T]) Set(i int, v T) {
	if i < 0 || i >= pa.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pa.size))
	}
	pa.array[i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (pa *PodArray[T]) ValueAt(i int) T {
	return pa.array[i]
}

// Resize changes the size of the array, reallocating if necessary.
func (pa *PodArray[T]) Resize(size int) {
	if size == pa.size {
		return
	}

	if size == 0 {
		pa.array = nil
		pa.size = 0
		return
	}

	newArray := make([]T, size)
	if pa.array != nil {
		copyLen := basics.IMin(pa.size, size)
		copy(newArray[:copyLen], pa.array[:copyLen])
	}

	pa.array = newArray
	pa.size = size
}

// Data returns the underlying slice.
func (pa *PodArray[T]) Data() []T {
	if pa.array == nil {
		return nil
	}
	return pa.array[:pa.size]
}

// Assign copies data from another PodArray.
func (pa *PodArray[T]) Assign(other *PodArray[T]) {
	if other == nil {
		pa.Resize(0)
		return
	}

	pa.Resize(other.size)
	if other.size > 0 {
		copy(pa.array[:pa.size], other.array[:other.size])
	}
}

// PodVector is a growable vector with automatic capacity management.
// This is equivalent to AGG's pod_vector<T> template class.
type PodVector[T any] struct {
	array    []T
	size     int
	capacity int
}

// NewPodVector creates a new growable vector.
func NewPodVector[T any]() *PodVector[T] {
	return &PodVector[T]{
		array:    nil,
		size:     0,
		capacity: 0,
	}
}

// NewPodVectorWithCapacity creates a new vector with the specified capacity.
func NewPodVectorWithCapacity[T any](cap int, extraTail int) *PodVector[T] {
	totalCap := cap + extraTail
	var array []T
	if totalCap > 0 {
		array = make([]T, totalCap)
	}
	return &PodVector[T]{
		array:    array,
		size:     0,
		capacity: totalCap,
	}
}

// NewPodVectorCopy creates a new vector as a copy of another.
func NewPodVectorCopy[T any](other *PodVector[T]) *PodVector[T] {
	if other == nil {
		return NewPodVector[T]()
	}

	pv := &PodVector[T]{
		size:     other.size,
		capacity: other.capacity,
	}

	if other.capacity > 0 {
		pv.array = make([]T, other.capacity)
		copy(pv.array[:other.size], other.array[:other.size])
	}

	return pv
}

// Size returns the number of elements.
func (pv *PodVector[T]) Size() int {
	return pv.size
}

// Capacity returns the current capacity.
func (pv *PodVector[T]) Capacity() int {
	return pv.capacity
}

// ByteSize returns the size in bytes.
func (pv *PodVector[T]) ByteSize() int {
	if pv.size == 0 {
		return 0
	}
	var dummy T
	return pv.size * int(unsafe.Sizeof(dummy))
}

// At returns the element at the specified index with bounds checking.
func (pv *PodVector[T]) At(i int) T {
	if i < 0 || i >= pv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pv.size))
	}
	return pv.array[i]
}

// Set sets the element at the specified index with bounds checking.
func (pv *PodVector[T]) Set(i int, v T) {
	if i < 0 || i >= pv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, pv.size))
	}
	pv.array[i] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (pv *PodVector[T]) ValueAt(i int) T {
	return pv.array[i]
}

// SetCapacity sets new capacity. All data is lost, size is set to zero.
func (pv *PodVector[T]) SetCapacity(cap int, extraTail int) {
	pv.size = 0
	newCap := cap + extraTail

	if newCap > pv.capacity {
		pv.capacity = newCap
		if pv.capacity > 0 {
			pv.array = make([]T, pv.capacity)
		} else {
			pv.array = nil
		}
	}
}

// Allocate allocates n elements. All data is lost, but elements can be accessed.
func (pv *PodVector[T]) Allocate(size int, extraTail int) {
	pv.SetCapacity(size, extraTail)
	pv.size = size
}

// Resize changes the size while keeping existing content.
func (pv *PodVector[T]) Resize(newSize int) {
	if newSize > pv.size {
		if newSize > pv.capacity {
			// Need to grow capacity
			newArray := make([]T, newSize)
			if pv.array != nil {
				copy(newArray[:pv.size], pv.array[:pv.size])
			}
			pv.array = newArray
			pv.capacity = newSize
		}
		pv.size = newSize
	} else {
		pv.size = newSize
	}
}

// Zero fills the vector with zero values.
func (pv *PodVector[T]) Zero() {
	var zero T
	for i := 0; i < pv.size; i++ {
		pv.array[i] = zero
	}
}

// Add appends an element to the vector.
func (pv *PodVector[T]) Add(v T) {
	if pv.size >= pv.capacity {
		// Auto-grow capacity
		newCap := pv.capacity * 2
		if newCap == 0 {
			newCap = 16 // Initial capacity
		}
		pv.Resize(pv.size + 1) // This will handle capacity growth
		pv.array[pv.size-1] = v
	} else {
		pv.array[pv.size] = v
		pv.size++
	}
}

// PushBack appends an element to the vector (equivalent to Add).
func (pv *PodVector[T]) PushBack(v T) {
	pv.Add(v)
}

// InsertAt inserts an element at the specified position.
func (pv *PodVector[T]) InsertAt(pos int, val T) {
	if pos >= pv.size {
		// Insert at end or beyond
		for pv.size < pos {
			var zero T
			pv.Add(zero)
		}
		pv.Add(val)
	} else {
		// Insert in middle - need to grow first
		pv.Add(val) // This ensures capacity
		// Move elements to make room
		copy(pv.array[pos+1:pv.size], pv.array[pos:pv.size-1])
		pv.array[pos] = val
	}
}

// IncSize increases the size by the specified amount.
func (pv *PodVector[T]) IncSize(size int) {
	pv.Resize(pv.size + size)
}

// RemoveAll clears all elements.
func (pv *PodVector[T]) RemoveAll() {
	pv.size = 0
}

// Clear clears all elements (equivalent to RemoveAll).
func (pv *PodVector[T]) Clear() {
	pv.size = 0
}

// CutAt reduces the size to the specified value if it's smaller.
func (pv *PodVector[T]) CutAt(num int) {
	if num < pv.size {
		pv.size = num
	}
}

// Data returns the underlying slice.
func (pv *PodVector[T]) Data() []T {
	if pv.array == nil {
		return nil
	}
	return pv.array[:pv.size]
}

// Serialize writes the vector data to the provided byte slice.
func (pv *PodVector[T]) Serialize(ptr []byte) {
	if pv.size == 0 {
		return
	}

	var dummy T
	elementSize := int(unsafe.Sizeof(dummy))

	for i := 0; i < pv.size; i++ {
		srcBytes := (*[1024]byte)(unsafe.Pointer(&pv.array[i]))[:elementSize:elementSize]
		copy(ptr[i*elementSize:], srcBytes)
	}
}

// Deserialize reads vector data from the provided byte slice.
func (pv *PodVector[T]) Deserialize(data []byte) {
	if len(data) == 0 {
		pv.RemoveAll()
		return
	}

	var dummy T
	elementSize := int(unsafe.Sizeof(dummy))
	numElements := len(data) / elementSize

	pv.Allocate(numElements, 0)

	for i := 0; i < numElements; i++ {
		dstBytes := (*[1024]byte)(unsafe.Pointer(&pv.array[i]))[:elementSize:elementSize]
		copy(dstBytes, data[i*elementSize:])
	}
}

// Assign copies data from another PodVector.
func (pv *PodVector[T]) Assign(other *PodVector[T]) {
	if other == nil {
		pv.RemoveAll()
		return
	}

	pv.Allocate(other.size, 0)
	if other.size > 0 {
		copy(pv.array[:pv.size], other.array[:other.size])
	}
}
