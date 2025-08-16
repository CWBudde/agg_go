package array

import (
	"fmt"
	"unsafe"
)

// Block represents a single memory block in the allocator.
type Block struct {
	data []byte
	size int
}

// BlockAllocator is an allocator for arbitrary POD data.
// Most useful in different cache systems for efficient memory allocations.
// Memory is allocated with blocks of fixed size. If required size exceeds
// the block size, the allocator creates a new block of the required size.
// This is equivalent to AGG's block_allocator class.
type BlockAllocator struct {
	blockSize     int
	blockPtrInc   int
	numBlocks     int
	maxBlocks     int
	blocks        []Block
	bufPtr        []byte
	rest          int
	currentOffset int
}

// NewBlockAllocator creates a new block allocator with the specified block size.
func NewBlockAllocator(blockSize int) *BlockAllocator {
	return NewBlockAllocatorWithIncrement(blockSize, 256-8)
}

// NewBlockAllocatorWithIncrement creates a new block allocator with custom increment.
func NewBlockAllocatorWithIncrement(blockSize, blockPtrInc int) *BlockAllocator {
	return &BlockAllocator{
		blockSize:     blockSize,
		blockPtrInc:   blockPtrInc,
		numBlocks:     0,
		maxBlocks:     0,
		blocks:        nil,
		bufPtr:        nil,
		rest:          0,
		currentOffset: 0,
	}
}

// RemoveAll deallocates all blocks and resets the allocator.
func (ba *BlockAllocator) RemoveAll() {
	// In Go, we just need to clear references and let GC handle deallocation
	ba.blocks = nil
	ba.numBlocks = 0
	ba.maxBlocks = 0
	ba.bufPtr = nil
	ba.rest = 0
	ba.currentOffset = 0
}

// allocateBlock allocates a new block of at least the specified size.
func (ba *BlockAllocator) allocateBlock(size int) {
	if size < ba.blockSize {
		size = ba.blockSize
	}

	if ba.numBlocks >= ba.maxBlocks {
		// Need to grow the block array
		newMaxBlocks := ba.maxBlocks + ba.blockPtrInc
		newBlocks := make([]Block, newMaxBlocks)

		if ba.blocks != nil {
			copy(newBlocks[:ba.numBlocks], ba.blocks[:ba.numBlocks])
		}

		ba.blocks = newBlocks
		ba.maxBlocks = newMaxBlocks
	}

	// Allocate the new block
	data := make([]byte, size)
	ba.blocks[ba.numBlocks] = Block{
		data: data,
		size: size,
	}

	ba.bufPtr = data
	ba.currentOffset = 0
	ba.numBlocks++
	ba.rest = size
}

// Allocate allocates memory of the specified size with optional alignment.
// Returns a byte slice pointing to the allocated memory.
func (ba *BlockAllocator) Allocate(size int, alignment int) []byte {
	if size == 0 {
		return nil
	}

	if alignment < 1 {
		alignment = 1
	}

	if size <= ba.rest {
		// Check if we can allocate from current block
		ptr := ba.bufPtr[ba.currentOffset:]

		if alignment > 1 {
			// Calculate alignment offset
			currentAddr := uintptr(unsafe.Pointer(&ptr[0]))
			aligned := (currentAddr + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1)
			alignOffset := int(aligned - currentAddr)

			totalSize := size + alignOffset
			if totalSize <= ba.rest {
				ba.rest -= totalSize
				ba.currentOffset += totalSize
				return ptr[alignOffset : alignOffset+size]
			}

			// Not enough space in current block, allocate new one
			ba.allocateBlock(size + alignment - 1)
			return ba.Allocate(size, alignment)
		}

		// No alignment needed
		ba.rest -= size
		ba.currentOffset += size
		return ptr[:size]
	}

	// Not enough space in current block
	ba.allocateBlock(size + alignment - 1)
	return ba.Allocate(size, alignment)
}

// AllocateBytes allocates memory for the specified number of bytes.
func (ba *BlockAllocator) AllocateBytes(size int) []byte {
	return ba.Allocate(size, 1)
}

// AllocateAligned allocates memory with the specified alignment.
func (ba *BlockAllocator) AllocateAligned(size int, alignment int) []byte {
	return ba.Allocate(size, alignment)
}

// AllocateType allocates memory for the specified type T.
func AllocateType[T any](ba *BlockAllocator) *T {
	var dummy T
	size := int(unsafe.Sizeof(dummy))
	alignment := int(unsafe.Alignof(dummy))

	bytes := ba.Allocate(size, alignment)
	if len(bytes) < size {
		return nil
	}

	return (*T)(unsafe.Pointer(&bytes[0]))
}

// AllocateTypeSlice allocates memory for a slice of the specified type T.
func AllocateTypeSlice[T any](ba *BlockAllocator, count int) []T {
	if count <= 0 {
		return nil
	}

	var dummy T
	size := int(unsafe.Sizeof(dummy)) * count
	alignment := int(unsafe.Alignof(dummy))

	bytes := ba.Allocate(size, alignment)
	if len(bytes) < size {
		return nil
	}

	// Convert byte slice to typed slice
	return unsafe.Slice((*T)(unsafe.Pointer(&bytes[0])), count)
}

// Stats returns allocation statistics.
func (ba *BlockAllocator) Stats() BlockAllocatorStats {
	totalAllocated := 0
	totalUsed := ba.currentOffset

	for i := 0; i < ba.numBlocks; i++ {
		totalAllocated += ba.blocks[i].size
	}

	return BlockAllocatorStats{
		NumBlocks:      ba.numBlocks,
		TotalAllocated: totalAllocated,
		TotalUsed:      totalUsed,
		CurrentRest:    ba.rest,
		BlockSize:      ba.blockSize,
	}
}

// BlockAllocatorStats provides information about allocator usage.
type BlockAllocatorStats struct {
	NumBlocks      int // Number of allocated blocks
	TotalAllocated int // Total bytes allocated
	TotalUsed      int // Total bytes used in current block
	CurrentRest    int // Remaining bytes in current block
	BlockSize      int // Default block size
}

// String returns a string representation of the stats.
func (stats BlockAllocatorStats) String() string {
	return fmt.Sprintf("BlockAllocator{blocks: %d, allocated: %d bytes, used: %d bytes, rest: %d bytes, block_size: %d}",
		stats.NumBlocks, stats.TotalAllocated, stats.TotalUsed, stats.CurrentRest, stats.BlockSize)
}

// PooledBlockAllocator provides a pool of block allocators for reuse.
type PooledBlockAllocator struct {
	allocators []*BlockAllocator
	blockSize  int
	maxPool    int
}

// NewPooledBlockAllocator creates a new pooled block allocator.
func NewPooledBlockAllocator(blockSize, maxPool int) *PooledBlockAllocator {
	return &PooledBlockAllocator{
		allocators: make([]*BlockAllocator, 0, maxPool),
		blockSize:  blockSize,
		maxPool:    maxPool,
	}
}

// Get returns an allocator from the pool or creates a new one.
func (pba *PooledBlockAllocator) Get() *BlockAllocator {
	if len(pba.allocators) > 0 {
		// Pop from pool
		idx := len(pba.allocators) - 1
		allocator := pba.allocators[idx]
		pba.allocators = pba.allocators[:idx]
		return allocator
	}

	// Create new allocator
	return NewBlockAllocator(pba.blockSize)
}

// Put returns an allocator to the pool for reuse.
func (pba *PooledBlockAllocator) Put(allocator *BlockAllocator) {
	if len(pba.allocators) < pba.maxPool {
		allocator.RemoveAll() // Clean up for reuse
		pba.allocators = append(pba.allocators, allocator)
	}
	// Otherwise, let it be garbage collected
}

// Clear removes all allocators from the pool.
func (pba *PooledBlockAllocator) Clear() {
	pba.allocators = pba.allocators[:0]
}
