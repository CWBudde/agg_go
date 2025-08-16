package array

import (
	"fmt"
	"unsafe"
)

// BlockScale represents the block scaling constants for PodBVector.
type BlockScale struct {
	Shift int // Block shift (S parameter from C++)
	Size  int // Block size = 1 << Shift
	Mask  int // Block mask = Size - 1
}

// NewBlockScale creates a new block scale with the specified shift.
func NewBlockScale(shift int) BlockScale {
	size := 1 << shift
	return BlockScale{
		Shift: shift,
		Size:  size,
		Mask:  size - 1,
	}
}

// DefaultBlockScale provides the default block scale (shift = 6, size = 64).
var DefaultBlockScale = NewBlockScale(6)

// PodBVector is a block-based vector similar to std::deque.
// It doesn't reallocate memory but instead uses blocks of data of power-of-two size.
// The data is NOT contiguous in memory, so only indexed access is valid.
// This is equivalent to AGG's pod_bvector<T, S> template class.
type PodBVector[T any] struct {
	scale       BlockScale
	size        int
	numBlocks   int
	maxBlocks   int
	blocks      [][]T
	blockPtrInc int
}

// NewPodBVector creates a new block vector with default block scale.
func NewPodBVector[T any]() *PodBVector[T] {
	return NewPodBVectorWithScale[T](DefaultBlockScale)
}

// NewPodBVectorWithScale creates a new block vector with the specified block scale.
func NewPodBVectorWithScale[T any](scale BlockScale) *PodBVector[T] {
	return &PodBVector[T]{
		scale:       scale,
		size:        0,
		numBlocks:   0,
		maxBlocks:   0,
		blocks:      nil,
		blockPtrInc: scale.Size,
	}
}

// NewPodBVectorWithIncrement creates a new block vector with custom block pointer increment.
func NewPodBVectorWithIncrement[T any](scale BlockScale, blockPtrInc int) *PodBVector[T] {
	return &PodBVector[T]{
		scale:       scale,
		size:        0,
		numBlocks:   0,
		maxBlocks:   0,
		blocks:      nil,
		blockPtrInc: blockPtrInc,
	}
}

// NewPodBVectorCopy creates a new block vector as a copy of another.
func NewPodBVectorCopy[T any](other *PodBVector[T]) *PodBVector[T] {
	if other == nil {
		return NewPodBVector[T]()
	}

	bv := &PodBVector[T]{
		scale:       other.scale,
		size:        other.size,
		numBlocks:   other.numBlocks,
		maxBlocks:   other.maxBlocks,
		blockPtrInc: other.blockPtrInc,
	}

	if other.maxBlocks > 0 {
		bv.blocks = make([][]T, other.maxBlocks)

		for i := 0; i < other.numBlocks; i++ {
			bv.blocks[i] = make([]T, other.scale.Size)
			copy(bv.blocks[i], other.blocks[i])
		}
	}

	return bv
}

// Size returns the number of elements.
func (bv *PodBVector[T]) Size() int {
	return bv.size
}

// Capacity returns the current capacity (number of allocated blocks * block size).
func (bv *PodBVector[T]) Capacity() int {
	return bv.numBlocks * bv.scale.Size
}

// ByteSize returns the size in bytes.
func (bv *PodBVector[T]) ByteSize() int {
	if bv.size == 0 {
		return 0
	}
	var dummy T
	return bv.size * int(unsafe.Sizeof(dummy))
}

// At returns the element at the specified index with bounds checking.
func (bv *PodBVector[T]) At(i int) T {
	if i < 0 || i >= bv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, bv.size))
	}
	blockIdx := i >> bv.scale.Shift
	elemIdx := i & bv.scale.Mask
	return bv.blocks[blockIdx][elemIdx]
}

// Set sets the element at the specified index with bounds checking.
func (bv *PodBVector[T]) Set(i int, v T) {
	if i < 0 || i >= bv.size {
		panic(fmt.Sprintf("index %d out of bounds [0, %d)", i, bv.size))
	}
	blockIdx := i >> bv.scale.Shift
	elemIdx := i & bv.scale.Mask
	bv.blocks[blockIdx][elemIdx] = v
}

// ValueAt returns the element at the specified index (unsafe, may panic).
func (bv *PodBVector[T]) ValueAt(i int) T {
	blockIdx := i >> bv.scale.Shift
	elemIdx := i & bv.scale.Mask
	return bv.blocks[blockIdx][elemIdx]
}

// allocateBlock allocates a new block at the specified index.
func (bv *PodBVector[T]) allocateBlock(nb int) {
	if nb >= bv.maxBlocks {
		// Need to grow the block pointer array
		newMaxBlocks := bv.maxBlocks + bv.blockPtrInc
		newBlocks := make([][]T, newMaxBlocks)

		if bv.blocks != nil {
			copy(newBlocks[:bv.numBlocks], bv.blocks[:bv.numBlocks])
		}

		bv.blocks = newBlocks
		bv.maxBlocks = newMaxBlocks
	}

	bv.blocks[nb] = make([]T, bv.scale.Size)
	bv.numBlocks++
}

// dataPtr returns a pointer to the location where the next element should be stored.
func (bv *PodBVector[T]) dataPtr() *T {
	blockIdx := bv.size >> bv.scale.Shift
	if blockIdx >= bv.numBlocks {
		bv.allocateBlock(blockIdx)
	}
	elemIdx := bv.size & bv.scale.Mask
	return &bv.blocks[blockIdx][elemIdx]
}

// Add appends an element to the vector.
func (bv *PodBVector[T]) Add(val T) {
	*bv.dataPtr() = val
	bv.size++
}

// PushBack appends an element to the vector (equivalent to Add).
func (bv *PodBVector[T]) PushBack(val T) {
	bv.Add(val)
}

// ModifyLast removes the last element and adds a new one.
func (bv *PodBVector[T]) ModifyLast(val T) {
	if bv.size > 0 {
		bv.RemoveLast()
	}
	bv.Add(val)
}

// RemoveLast removes the last element.
func (bv *PodBVector[T]) RemoveLast() {
	if bv.size > 0 {
		bv.size--
	}
}

// RemoveAll clears all elements but keeps allocated blocks.
func (bv *PodBVector[T]) RemoveAll() {
	bv.size = 0
}

// Clear clears all elements (equivalent to RemoveAll).
func (bv *PodBVector[T]) Clear() {
	bv.size = 0
}

// FreeAll deallocates all blocks and resets the vector.
func (bv *PodBVector[T]) FreeAll() {
	bv.FreeTail(0)
}

// FreeTail deallocates blocks beyond the specified size.
func (bv *PodBVector[T]) FreeTail(size int) {
	if size < bv.size {
		neededBlocks := (size + bv.scale.Mask) >> bv.scale.Shift

		// Free unnecessary blocks
		for bv.numBlocks > neededBlocks {
			bv.numBlocks--
			bv.blocks[bv.numBlocks] = nil // Let GC handle deallocation
		}

		// If no blocks needed, free the block pointer array too
		if bv.numBlocks == 0 {
			bv.blocks = nil
			bv.maxBlocks = 0
		}

		bv.size = size
	}
}

// AllocateContinuousBlock allocates a continuous block of elements.
// Returns the starting index if successful, -1 if impossible.
func (bv *PodBVector[T]) AllocateContinuousBlock(numElements int) int {
	if numElements >= bv.scale.Size {
		return -1 // Impossible to allocate
	}

	// Ensure we have at least one block
	bv.dataPtr()

	rest := bv.scale.Size - (bv.size & bv.scale.Mask)

	if numElements <= rest {
		// The rest of the current block is sufficient
		index := bv.size
		bv.size += numElements
		return index
	}

	// Move to the next block
	bv.size += rest
	bv.dataPtr() // Allocate the new block

	index := bv.size
	bv.size += numElements
	return index
}

// AddArray adds multiple elements from a slice.
func (bv *PodBVector[T]) AddArray(elements []T) {
	for _, elem := range elements {
		bv.Add(elem)
	}
}

// AddData adds elements from a data accessor (function that provides elements).
func (bv *PodBVector[T]) AddData(accessor func() (T, bool)) {
	for {
		if elem, ok := accessor(); ok {
			bv.Add(elem)
		} else {
			break
		}
	}
}

// CutAt reduces the size to the specified value if it's smaller.
func (bv *PodBVector[T]) CutAt(size int) {
	if size < bv.size {
		bv.size = size
	}
}

// Serialize writes the vector data to the provided byte slice.
func (bv *PodBVector[T]) Serialize(ptr []byte) {
	if bv.size == 0 {
		return
	}

	var dummy T
	elementSize := int(unsafe.Sizeof(dummy))

	for i := 0; i < bv.size; i++ {
		elem := bv.ValueAt(i)
		srcBytes := (*[1024]byte)(unsafe.Pointer(&elem))[:elementSize:elementSize]
		copy(ptr[i*elementSize:], srcBytes)
	}
}

// Deserialize reads vector data from the provided byte slice.
func (bv *PodBVector[T]) Deserialize(data []byte) {
	bv.RemoveAll()

	if len(data) == 0 {
		return
	}

	var dummy T
	elementSize := int(unsafe.Sizeof(dummy))
	numElements := len(data) / elementSize

	for i := 0; i < numElements; i++ {
		ptr := bv.dataPtr()
		dstBytes := (*[1024]byte)(unsafe.Pointer(ptr))[:elementSize:elementSize]
		copy(dstBytes, data[i*elementSize:])
		bv.size++
	}
}

// DeserializeAt replaces or adds elements starting from the specified position.
func (bv *PodBVector[T]) DeserializeAt(start int, emptyVal T, data []byte) {
	// Fill with empty values up to start position
	for bv.size < start {
		bv.Add(emptyVal)
	}

	if len(data) == 0 {
		return
	}

	var dummy T
	elementSize := int(unsafe.Sizeof(dummy))
	numElements := len(data) / elementSize

	for i := 0; i < numElements; i++ {
		pos := start + i

		if pos < bv.size {
			// Replace existing element
			var elem T
			dstBytes := (*[1024]byte)(unsafe.Pointer(&elem))[:elementSize:elementSize]
			copy(dstBytes, data[i*elementSize:])
			bv.Set(pos, elem)
		} else {
			// Add new element
			ptr := bv.dataPtr()
			dstBytes := (*[1024]byte)(unsafe.Pointer(ptr))[:elementSize:elementSize]
			copy(dstBytes, data[i*elementSize:])
			bv.size++
		}
	}
}

// Assign copies data from another PodBVector.
func (bv *PodBVector[T]) Assign(other *PodBVector[T]) {
	if other == nil {
		bv.RemoveAll()
		return
	}

	// Ensure we have enough blocks
	neededBlocks := other.numBlocks
	for bv.numBlocks < neededBlocks {
		bv.allocateBlock(bv.numBlocks)
	}

	// Copy the data block by block
	for i := 0; i < other.numBlocks; i++ {
		copy(bv.blocks[i], other.blocks[i])
	}

	bv.size = other.size
}
