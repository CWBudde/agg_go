package basics

import (
	"fmt"
	"testing"
)

func TestBitsetIterator_BasicIteration(t *testing.T) {
	// Test data: 10101010 (0xAA)
	data := []byte{0xAA}
	iter := NewBitsetIterator(data, 0)

	expected := []uint{1, 0, 1, 0, 1, 0, 1, 0}

	for i, exp := range expected {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d", i)
		}

		bit := iter.Bit()
		if bit != exp {
			t.Errorf("Bit %d: expected %d, got %d", i, exp, bit)
		}

		iter.Next()
	}

	if iter.HasNext() {
		t.Error("Iterator should have ended")
	}
}

func TestBitsetIterator_WithOffset(t *testing.T) {
	// Test data: 11110000 (0xF0)
	data := []byte{0xF0}

	tests := []struct {
		offset   uint
		expected []uint
	}{
		{0, []uint{1, 1, 1, 1, 0, 0, 0, 0}},
		{2, []uint{1, 1, 0, 0, 0, 0}},
		{4, []uint{0, 0, 0, 0}},
		{6, []uint{0, 0}},
		{7, []uint{0}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("offset_%d", test.offset), func(t *testing.T) {
			iter := NewBitsetIterator(data, test.offset)

			for i, exp := range test.expected {
				if !iter.HasNext() {
					t.Fatalf("Iterator ended prematurely at bit %d (offset %d)", i, test.offset)
				}

				bit := iter.Bit()
				if bit != exp {
					t.Errorf("Offset %d, bit %d: expected %d, got %d", test.offset, i, exp, bit)
				}

				iter.Next()
			}

			if iter.HasNext() {
				t.Errorf("Iterator should have ended (offset %d)", test.offset)
			}
		})
	}
}

func TestBitsetIterator_MultipleBytesWithOffset(t *testing.T) {
	// Test data: 11110000 01010101 (0xF0, 0x55)
	data := []byte{0xF0, 0x55}

	// Start at bit 6 (2nd bit from end of first byte)
	iter := NewBitsetIterator(data, 6)

	// Should get: 00 (from first byte) + 01010101 (from second byte)
	expected := []uint{0, 0, 0, 1, 0, 1, 0, 1, 0, 1}

	for i, exp := range expected {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d", i)
		}

		bit := iter.Bit()
		if bit != exp {
			t.Errorf("Bit %d: expected %d, got %d", i, exp, bit)
		}

		iter.Next()
	}
}

func TestBitsetIterator_ByteBoundary(t *testing.T) {
	// Test crossing byte boundaries
	data := []byte{0xFF, 0x00, 0xAA}
	iter := NewBitsetIterator(data, 0)

	// First byte: all 1s
	for i := 0; i < 8; i++ {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d", i)
		}

		bit := iter.Bit()
		if bit != 1 {
			t.Errorf("First byte, bit %d: expected 1, got %d", i, bit)
		}

		iter.Next()
	}

	// Second byte: all 0s
	for i := 0; i < 8; i++ {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d (second byte)", i)
		}

		bit := iter.Bit()
		if bit != 0 {
			t.Errorf("Second byte, bit %d: expected 0, got %d", i, bit)
		}

		iter.Next()
	}

	// Third byte: 10101010
	expected := []uint{1, 0, 1, 0, 1, 0, 1, 0}
	for i, exp := range expected {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d (third byte)", i)
		}

		bit := iter.Bit()
		if bit != exp {
			t.Errorf("Third byte, bit %d: expected %d, got %d", i, exp, bit)
		}

		iter.Next()
	}
}

func TestBitsetIterator_EmptyData(t *testing.T) {
	iter := NewBitsetIterator(nil, 0)

	if iter.HasNext() {
		t.Error("Empty iterator should not have next")
	}

	bit := iter.Bit()
	if bit != 0 {
		t.Errorf("Empty iterator bit should be 0, got %d", bit)
	}

	// Should not panic
	iter.Next()
}

func TestBitsetIterator_OutOfBounds(t *testing.T) {
	data := []byte{0xFF}

	// Offset beyond data
	iter := NewBitsetIterator(data, 10)

	if iter.HasNext() {
		t.Error("Out of bounds iterator should not have next")
	}

	bit := iter.Bit()
	if bit != 0 {
		t.Errorf("Out of bounds iterator bit should be 0, got %d", bit)
	}
}

func TestBitsetIterator_SingleBit(t *testing.T) {
	data := []byte{0x80} // Only MSB set

	tests := []struct {
		offset   uint
		expected uint
	}{
		{0, 1}, // MSB
		{1, 0}, // Second bit
		{7, 0}, // LSB
	}

	for _, test := range tests {
		iter := NewBitsetIterator(data, test.offset)

		if !iter.HasNext() && test.expected == 1 {
			t.Errorf("Iterator should have next for offset %d", test.offset)
		}

		bit := iter.Bit()
		if bit != test.expected {
			t.Errorf("Offset %d: expected %d, got %d", test.offset, test.expected, bit)
		}
	}
}

func TestBitsetIterator_AllPatterns(t *testing.T) {
	patterns := []struct {
		name     string
		data     byte
		expected []uint
	}{
		{"all_zeros", 0x00, []uint{0, 0, 0, 0, 0, 0, 0, 0}},
		{"all_ones", 0xFF, []uint{1, 1, 1, 1, 1, 1, 1, 1}},
		{"alternating_1", 0xAA, []uint{1, 0, 1, 0, 1, 0, 1, 0}},
		{"alternating_2", 0x55, []uint{0, 1, 0, 1, 0, 1, 0, 1}},
		{"first_half", 0xF0, []uint{1, 1, 1, 1, 0, 0, 0, 0}},
		{"second_half", 0x0F, []uint{0, 0, 0, 0, 1, 1, 1, 1}},
	}

	for _, pattern := range patterns {
		t.Run(pattern.name, func(t *testing.T) {
			data := []byte{pattern.data}
			iter := NewBitsetIterator(data, 0)

			for i, exp := range pattern.expected {
				if !iter.HasNext() {
					t.Fatalf("Iterator ended prematurely at bit %d", i)
				}

				bit := iter.Bit()
				if bit != exp {
					t.Errorf("Bit %d: expected %d, got %d", i, exp, bit)
				}

				iter.Next()
			}
		})
	}
}

func BenchmarkBitsetIterator_SimpleByte(b *testing.B) {
	data := []byte{0xAA}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIterator(data, 0)
		for iter.HasNext() {
			_ = iter.Bit()
			iter.Next()
		}
	}
}

func BenchmarkBitsetIterator_LargeData(b *testing.B) {
	// Create 1KB of test data
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA // Alternating pattern
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIterator(data, 0)
		count := 0
		for iter.HasNext() {
			if iter.Bit() != 0 {
				count++
			}
			iter.Next()
		}
		// Should have counted 4096 set bits (1024 * 4)
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

// Benchmarks for new functionality

func BenchmarkBitsetIterator_FindNextSetBit_Dense(b *testing.B) {
	// Dense data (50% set bits)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA // 50% density
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIterator(data, 0)
		count := 0
		for iter.FindNextSetBit() {
			count++
			iter.Next()
		}
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

func BenchmarkBitsetIterator_FindNextSetBit_Sparse(b *testing.B) {
	// Sparse data (1 bit per 64 bits)
	data := make([]byte, 1024)
	for i := 0; i < len(data); i += 8 {
		data[i] = 0x01 // Only 1 bit per 64
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIterator(data, 0)
		count := 0
		for iter.FindNextSetBit() {
			count++
			iter.Next()
		}
		// Should find 128 set bits (1024/8)
		if count != 128 {
			b.Fatalf("Expected 128 set bits, got %d", count)
		}
	}
}

func BenchmarkBitsetIterator_CountSetBits(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIterator(data, 0)
		count := iter.CountSetBits()
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

// Benchmarks for BitsetIteratorOptimized

func BenchmarkBitsetIteratorOptimized_BasicIteration(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIteratorOptimized(data, 0)
		count := 0
		for iter.HasNext() {
			if iter.Bit() != 0 {
				count++
			}
			iter.Next()
		}
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

func BenchmarkBitsetIteratorOptimized_FindNextSetBit_Dense(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIteratorOptimized(data, 0)
		count := 0
		for iter.FindNextSetBit() {
			count++
			iter.Next()
		}
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

func BenchmarkBitsetIteratorOptimized_FindNextSetBit_Sparse(b *testing.B) {
	// Very sparse data (1 bit per 512 bits)
	data := make([]byte, 1024)
	for i := 0; i < len(data); i += 64 {
		data[i] = 0x01
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIteratorOptimized(data, 0)
		count := 0
		for iter.FindNextSetBit() {
			count++
			iter.Next()
		}
		// Should find 16 set bits (1024/64)
		if count != 16 {
			b.Fatalf("Expected 16 set bits, got %d", count)
		}
	}
}

func BenchmarkBitsetIteratorOptimized_CountSetBits(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter := NewBitsetIteratorOptimized(data, 0)
		count := iter.CountSetBits()
		if count != 4096 {
			b.Fatalf("Expected 4096 set bits, got %d", count)
		}
	}
}

// Comparison benchmarks

func BenchmarkBitsetIterator_vs_Optimized_Dense(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xFF // Dense data
	}

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIterator(data, 0)
			count := 0
			for iter.FindNextSetBit() {
				count++
				iter.Next()
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIteratorOptimized(data, 0)
			count := 0
			for iter.FindNextSetBit() {
				count++
				iter.Next()
			}
		}
	})
}

func BenchmarkBitsetIterator_vs_Optimized_Sparse(b *testing.B) {
	data := make([]byte, 1024)
	for i := 0; i < len(data); i += 128 {
		data[i] = 0x01 // Very sparse
	}

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIterator(data, 0)
			count := 0
			for iter.FindNextSetBit() {
				count++
				iter.Next()
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIteratorOptimized(data, 0)
			count := 0
			for iter.FindNextSetBit() {
				count++
				iter.Next()
			}
		}
	})
}

func BenchmarkBitsetIterator_vs_Optimized_CountSetBits(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xAA
	}

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIterator(data, 0)
			_ = iter.CountSetBits()
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			iter := NewBitsetIteratorOptimized(data, 0)
			_ = iter.CountSetBits()
		}
	})
}

func TestBitsetIterator_RealWorldPattern(t *testing.T) {
	// Simulate a simple monospace font bitmap pattern
	// This could represent a small portion of a character like 'H'
	fontData := []byte{
		0x81, // 10000001 - left and right edges
		0x81, // 10000001 - left and right edges
		0xFF, // 11111111 - horizontal bar
		0x81, // 10000001 - left and right edges
	}

	iter := NewBitsetIterator(fontData, 0)
	bitCount := 0
	totalBits := 0

	for iter.HasNext() {
		if iter.Bit() != 0 {
			bitCount++
		}
		totalBits++
		iter.Next()
	}

	// Should have 32 total bits and 14 set bits
	// 0x81 (10000001) = 2 bits, 0xFF (11111111) = 8 bits
	// Total: 2 + 2 + 8 + 2 = 14 set bits
	if totalBits != 32 {
		t.Errorf("Expected 32 total bits, got %d", totalBits)
	}

	if bitCount != 14 {
		t.Errorf("Expected 14 set bits, got %d", bitCount)
	}
}

// Tests for new BitsetIterator methods

func TestBitsetIterator_Begin(t *testing.T) {
	data := []byte{0xAA, 0x55} // 10101010 01010101
	iter := NewBitsetIterator(data, 3)

	// Advance a few positions
	iter.Next()
	iter.Next()
	iter.Next()

	origPos := iter.Current()
	if origPos != 6 {
		t.Errorf("Expected position 6, got %d", origPos)
	}

	// Reset to beginning
	iter.Begin()

	if iter.Current() != 3 {
		t.Errorf("Expected position 3 after Begin(), got %d", iter.Current())
	}

	// Should be at original starting bit
	expectedBit := uint(0) // bit 3 of 0xAA (10101010) is 0
	if iter.Bit() != expectedBit {
		t.Errorf("Expected bit %d after Begin(), got %d", expectedBit, iter.Bit())
	}
}

func TestBitsetIterator_Current(t *testing.T) {
	data := []byte{0xFF}
	iter := NewBitsetIterator(data, 2)

	positions := []uint{2, 3, 4, 5}
	for _, expectedPos := range positions {
		if iter.Current() != expectedPos {
			t.Errorf("Expected position %d, got %d", expectedPos, iter.Current())
		}
		iter.Next()
	}
}

func TestBitsetIterator_Done(t *testing.T) {
	data := []byte{0x80} // Only MSB set
	iter := NewBitsetIterator(data, 0)

	// Should not be done initially
	if iter.Done() {
		t.Error("Iterator should not be done initially")
	}

	// Advance through all bits
	for iter.HasNext() {
		iter.Next()
	}

	// Should be done now
	if !iter.Done() {
		t.Error("Iterator should be done after advancing through all bits")
	}
}

func TestBitsetIterator_FindNextSetBit(t *testing.T) {
	data := []byte{0x88} // 10001000 - bits 0 and 4 are set
	iter := NewBitsetIterator(data, 0)

	// First set bit should be at position 0
	found := iter.FindNextSetBit()
	if !found {
		t.Fatal("Should have found first set bit")
	}
	if iter.Current() != 0 {
		t.Errorf("Expected position 0, got %d", iter.Current())
	}
	if iter.Bit() == 0 {
		t.Error("Expected bit to be set")
	}

	// Move past first set bit and find next
	iter.Next()
	found = iter.FindNextSetBit()
	if !found {
		t.Fatal("Should have found second set bit")
	}
	if iter.Current() != 4 {
		t.Errorf("Expected position 4, got %d", iter.Current())
	}
	if iter.Bit() == 0 {
		t.Error("Expected bit to be set")
	}

	// No more set bits after this
	iter.Next()
	found = iter.FindNextSetBit()
	if found {
		t.Error("Should not have found more set bits")
	}
}

func TestBitsetIterator_CountSetBits(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		offset   uint
		expected uint
	}{
		{
			name:     "all_ones",
			data:     []byte{0xFF},
			offset:   0,
			expected: 8,
		},
		{
			name:     "half_ones",
			data:     []byte{0xF0},
			offset:   0,
			expected: 4,
		},
		{
			name:     "with_offset",
			data:     []byte{0xFF},
			offset:   4,
			expected: 4,
		},
		{
			name:     "multiple_bytes",
			data:     []byte{0xFF, 0x0F},
			offset:   0,
			expected: 12,
		},
		{
			name:     "sparse_pattern",
			data:     []byte{0xAA, 0x55}, // 10101010 01010101
			offset:   0,
			expected: 8,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iter := NewBitsetIterator(test.data, test.offset)
			count := iter.CountSetBits()
			if count != test.expected {
				t.Errorf("Expected %d set bits, got %d", test.expected, count)
			}

			// Verify iterator position hasn't changed
			if iter.Current() != test.offset {
				t.Errorf("Iterator position changed: expected %d, got %d", test.offset, iter.Current())
			}
		})
	}
}

func TestBitsetIterator_AdvanceToNextSetBit(t *testing.T) {
	data := []byte{0x88} // 10001000 - bits 0 and 4 are set
	iter := NewBitsetIterator(data, 0)

	// First set bit is at position 0, should advance to position 1
	found := iter.AdvanceToNextSetBit()
	if !found {
		t.Fatal("Should have found and advanced past first set bit")
	}
	if iter.Current() != 1 {
		t.Errorf("Expected position 1 after advance, got %d", iter.Current())
	}

	// Next set bit is at position 4, should advance to position 5
	found = iter.AdvanceToNextSetBit()
	if !found {
		t.Fatal("Should have found and advanced past second set bit")
	}
	if iter.Current() != 5 {
		t.Errorf("Expected position 5 after advance, got %d", iter.Current())
	}

	// No more set bits
	found = iter.AdvanceToNextSetBit()
	if found {
		t.Error("Should not have found more set bits")
	}
}

// Tests for BitsetIteratorOptimized

func TestBitsetIteratorOptimized_BasicFunctionality(t *testing.T) {
	data := []byte{0xAA, 0x55} // 10101010 01010101
	iter := NewBitsetIteratorOptimized(data, 0)

	expected := []uint{1, 0, 1, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 1, 0, 1}

	for i, exp := range expected {
		if !iter.HasNext() {
			t.Fatalf("Iterator ended prematurely at bit %d", i)
		}

		bit := iter.Bit()
		if bit != exp {
			t.Errorf("Bit %d: expected %d, got %d", i, exp, bit)
		}

		if iter.Current() != uint(i) {
			t.Errorf("Position %d: expected %d, got %d", i, i, iter.Current())
		}

		iter.Next()
	}
}

func TestBitsetIteratorOptimized_FindNextSetBit(t *testing.T) {
	// Create data with set bits at positions 5, 13, 67 (crossing word boundaries)
	data := make([]byte, 16)
	// Bit 5 means: byte 0, bit 5 (counting from MSB): 0x04 (00000100)
	data[0] = 0x04 // bit 5 in first byte
	// Bit 13 means: byte 1, bit 5 (13 % 8 = 5): 0x04 (00000100)
	data[1] = 0x04 // bit 13 (5 + 8)
	// Bit 67 means: byte 8, bit 3 (67 % 8 = 3): 0x10 (00010000)
	data[8] = 0x10 // bit 67 (3 + 64)

	iter := NewBitsetIteratorOptimized(data, 0)

	// Find first set bit
	found := iter.FindNextSetBit()
	if !found {
		t.Fatal("Should have found first set bit")
	}
	if iter.Current() != 5 {
		t.Errorf("Expected position 5, got %d", iter.Current())
	}

	// Advance and find next
	iter.Next()
	found = iter.FindNextSetBit()
	if !found {
		t.Fatal("Should have found second set bit")
	}
	if iter.Current() != 13 {
		t.Errorf("Expected position 13, got %d", iter.Current())
	}

	// Advance and find next (crosses word boundary)
	iter.Next()
	found = iter.FindNextSetBit()
	if !found {
		t.Fatal("Should have found third set bit")
	}
	if iter.Current() != 67 {
		t.Errorf("Expected position 67, got %d", iter.Current())
	}
}

func TestBitsetIteratorOptimized_CountSetBits(t *testing.T) {
	// Create pattern with known number of set bits
	data := []byte{0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00}
	iter := NewBitsetIteratorOptimized(data, 0)

	count := iter.CountSetBits()
	expected := uint(32) // 4 bytes * 8 bits = 32 set bits
	if count != expected {
		t.Errorf("Expected %d set bits, got %d", expected, count)
	}
}

func TestBitsetIteratorOptimized_LargeData(t *testing.T) {
	// Test with data larger than 64 bits
	size := 1024 // 8192 bits
	data := make([]byte, size)

	// Set every 8th bit
	expectedCount := uint(0)
	for i := 0; i < size; i++ {
		data[i] = 0x80 // Only MSB set in each byte
		expectedCount++
	}

	iter := NewBitsetIteratorOptimized(data, 0)
	count := iter.CountSetBits()

	if count != expectedCount {
		t.Errorf("Expected %d set bits, got %d", expectedCount, count)
	}
}

func TestBitsetIteratorOptimized_Begin(t *testing.T) {
	data := []byte{0xAA, 0x55}
	iter := NewBitsetIteratorOptimized(data, 3)

	// Advance and then reset
	iter.Next()
	iter.Next()
	iter.Begin()

	if iter.Current() != 3 {
		t.Errorf("Expected position 3 after Begin(), got %d", iter.Current())
	}
}

func TestBitsetIteratorOptimized_EmptyData(t *testing.T) {
	iter := NewBitsetIteratorOptimized(nil, 0)

	if iter.HasNext() {
		t.Error("Empty iterator should not have next")
	}

	if !iter.Done() {
		t.Error("Empty iterator should be done")
	}

	if iter.CountSetBits() != 0 {
		t.Error("Empty iterator should have 0 set bits")
	}
}
