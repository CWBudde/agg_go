package array

import (
	"testing"
	"unsafe"
)

func TestBlockAllocatorBasic(t *testing.T) {
	ba := NewBlockAllocator(1024)

	// Test initial state
	stats := ba.Stats()
	if stats.NumBlocks != 0 {
		t.Errorf("Expected 0 blocks initially, got %d", stats.NumBlocks)
	}

	// Test basic allocation
	data := ba.AllocateBytes(100)
	if len(data) != 100 {
		t.Errorf("Expected 100 bytes, got %d", len(data))
	}

	stats = ba.Stats()
	if stats.NumBlocks != 1 {
		t.Errorf("Expected 1 block after allocation, got %d", stats.NumBlocks)
	}
	if stats.BlockSize != 1024 {
		t.Errorf("Expected block size 1024, got %d", stats.BlockSize)
	}
}

func TestBlockAllocatorMultipleAllocations(t *testing.T) {
	ba := NewBlockAllocator(256)

	// Allocate multiple small chunks
	chunks := make([][]byte, 0)
	for i := 0; i < 10; i++ {
		chunk := ba.AllocateBytes(20)
		if len(chunk) != 20 {
			t.Errorf("Allocation %d: expected 20 bytes, got %d", i, len(chunk))
		}
		chunks = append(chunks, chunk)

		// Write test data
		for j := range chunk {
			chunk[j] = byte(i)
		}
	}

	// Verify data integrity
	for i, chunk := range chunks {
		for _, b := range chunk {
			if b != byte(i) {
				t.Errorf("Data corruption in chunk %d", i)
				break
			}
		}
	}

	stats := ba.Stats()
	if stats.NumBlocks != 1 {
		t.Errorf("Expected 1 block for small allocations, got %d", stats.NumBlocks)
	}
}

func TestBlockAllocatorLargeAllocation(t *testing.T) {
	ba := NewBlockAllocator(256)

	// Allocate something larger than block size
	data := ba.AllocateBytes(512)
	if len(data) != 512 {
		t.Errorf("Expected 512 bytes, got %d", len(data))
	}

	stats := ba.Stats()
	if stats.NumBlocks != 1 {
		t.Errorf("Expected 1 block for large allocation, got %d", stats.NumBlocks)
	}
	if stats.TotalAllocated < 512 {
		t.Errorf("Expected at least 512 bytes allocated, got %d", stats.TotalAllocated)
	}
}

func TestBlockAllocatorAlignment(t *testing.T) {
	ba := NewBlockAllocator(1024)

	// Test different alignments
	alignments := []int{1, 2, 4, 8, 16, 32}

	for _, align := range alignments {
		data := ba.AllocateAligned(64, align)
		if len(data) != 64 {
			t.Errorf("Aligned allocation: expected 64 bytes, got %d", len(data))
		}

		// Check alignment
		addr := uintptr(unsafe.Pointer(&data[0]))
		if addr%uintptr(align) != 0 {
			t.Errorf("Allocation not aligned to %d bytes: address %x", align, addr)
		}
	}
}

func TestBlockAllocatorZeroSize(t *testing.T) {
	ba := NewBlockAllocator(256)

	data := ba.AllocateBytes(0)
	if data != nil {
		t.Error("Expected nil for zero-size allocation")
	}

	stats := ba.Stats()
	if stats.NumBlocks != 0 {
		t.Errorf("Expected 0 blocks for zero allocation, got %d", stats.NumBlocks)
	}
}

func TestBlockAllocatorRemoveAll(t *testing.T) {
	ba := NewBlockAllocator(256)

	// Allocate several chunks
	for i := 0; i < 5; i++ {
		ba.AllocateBytes(100)
	}

	stats := ba.Stats()
	if stats.NumBlocks == 0 {
		t.Error("Expected blocks to be allocated")
	}

	// Remove all
	ba.RemoveAll()

	stats = ba.Stats()
	if stats.NumBlocks != 0 {
		t.Errorf("Expected 0 blocks after RemoveAll, got %d", stats.NumBlocks)
	}
	if stats.TotalAllocated != 0 {
		t.Errorf("Expected 0 total allocated after RemoveAll, got %d", stats.TotalAllocated)
	}
}

func TestBlockAllocatorWithIncrement(t *testing.T) {
	ba := NewBlockAllocatorWithIncrement(256, 64)

	// Allocate enough to force block array growth
	for i := 0; i < 100; i++ {
		ba.AllocateBytes(300) // Each allocation creates a new block
	}

	stats := ba.Stats()
	if stats.NumBlocks != 100 {
		t.Errorf("Expected 100 blocks, got %d", stats.NumBlocks)
	}
}

func TestAllocateType(t *testing.T) {
	ba := NewBlockAllocator(1024)

	// Test allocating a struct
	type TestStruct struct {
		A int32
		B float64
		C bool
	}

	ptr := AllocateType[TestStruct](ba)
	if ptr == nil {
		t.Error("Expected non-nil pointer")
	}

	// Test the allocated memory
	ptr.A = 42
	ptr.B = 3.14
	ptr.C = true

	if ptr.A != 42 || ptr.B != 3.14 || ptr.C != true {
		t.Error("Allocated memory not working correctly")
	}

	// Check alignment
	addr := uintptr(unsafe.Pointer(ptr))
	expectedAlign := unsafe.Alignof(*ptr)
	if addr%expectedAlign != 0 {
		t.Errorf("Struct not properly aligned: address %x, expected alignment %d", addr, expectedAlign)
	}
}

func TestAllocateTypeSlice(t *testing.T) {
	ba := NewBlockAllocator(1024)

	slice := AllocateTypeSlice[int32](ba, 10)
	if len(slice) != 10 {
		t.Errorf("Expected slice length 10, got %d", len(slice))
	}

	// Test the slice
	for i := range slice {
		slice[i] = int32(i * 2)
	}

	for i, expected := range slice {
		if expected != int32(i*2) {
			t.Errorf("Slice[%d]: expected %d, got %d", i, i*2, expected)
		}
	}

	// Test zero-length slice
	emptySlice := AllocateTypeSlice[int](ba, 0)
	if emptySlice != nil {
		t.Error("Expected nil for zero-length slice")
	}
}

func TestPooledBlockAllocator(t *testing.T) {
	pool := NewPooledBlockAllocator(256, 3)

	// Get allocators from pool
	allocators := make([]*BlockAllocator, 0)
	for i := 0; i < 5; i++ {
		ba := pool.Get()
		if ba == nil {
			t.Error("Expected non-nil allocator from pool")
		}
		allocators = append(allocators, ba)

		// Use the allocator
		ba.AllocateBytes(100)
	}

	// Return allocators to pool
	for _, ba := range allocators {
		pool.Put(ba)
	}

	// Pool should only keep maxPool allocators
	if len(pool.allocators) > 3 {
		t.Errorf("Pool should not exceed maxPool size, got %d", len(pool.allocators))
	}

	// Get allocator again - should be reused
	ba := pool.Get()
	if ba == nil {
		t.Error("Expected reused allocator from pool")
	}

	// Should be clean
	stats := ba.Stats()
	if stats.NumBlocks != 0 {
		t.Error("Reused allocator should be clean")
	}
}

func TestPooledBlockAllocatorClear(t *testing.T) {
	pool := NewPooledBlockAllocator(256, 5)

	// Add allocators to pool
	allocators := make([]*BlockAllocator, 3)
	for i := 0; i < 3; i++ {
		ba := pool.Get()
		ba.AllocateBytes(100)
		allocators[i] = ba
	}

	// Put them all back
	for _, ba := range allocators {
		pool.Put(ba)
	}

	if len(pool.allocators) != 3 {
		t.Errorf("Expected 3 allocators in pool, got %d", len(pool.allocators))
	}

	pool.Clear()

	if len(pool.allocators) != 0 {
		t.Errorf("Expected 0 allocators after clear, got %d", len(pool.allocators))
	}
}

func TestBlockAllocatorStats(t *testing.T) {
	ba := NewBlockAllocator(512)

	// Test initial stats
	stats := ba.Stats()
	expected := "BlockAllocator{blocks: 0, allocated: 0 bytes, used: 0 bytes, rest: 0 bytes, block_size: 512}"
	if stats.String() != expected {
		t.Errorf("Initial stats string incorrect: got %s", stats.String())
	}

	// Allocate some memory
	ba.AllocateBytes(100)
	ba.AllocateBytes(200)

	stats = ba.Stats()
	if stats.NumBlocks != 1 {
		t.Errorf("Expected 1 block, got %d", stats.NumBlocks)
	}
	if stats.TotalAllocated != 512 {
		t.Errorf("Expected 512 bytes allocated, got %d", stats.TotalAllocated)
	}
	if stats.TotalUsed != 300 {
		t.Errorf("Expected 300 bytes used, got %d", stats.TotalUsed)
	}
	if stats.CurrentRest != 212 {
		t.Errorf("Expected 212 bytes rest, got %d", stats.CurrentRest)
	}
}
