package basics

import (
	"math/bits"
)

// BitsetIteratorOptimized provides highly optimized iteration over bitsets using 64-bit words.
// This implementation uses hardware bit scan instructions and word-level operations for
// maximum performance when dealing with sparse or large bitsets.
type BitsetIteratorOptimized struct {
	data        []byte   // Original data
	words       []uint64 // 64-bit words for fast processing
	currentWord uint64   // Current word being processed
	wordIndex   int      // Index of current word
	bitIndex    uint     // Bit index within current word
	position    uint     // Global bit position
	origOffset  uint     // Original starting offset
	totalBits   uint     // Total bits available
}

// NewBitsetIteratorOptimized creates a new optimized bitset iterator.
// This version processes data in 64-bit chunks for improved performance.
func NewBitsetIteratorOptimized(bits []byte, offset uint) *BitsetIteratorOptimized {
	if len(bits) == 0 {
		return &BitsetIteratorOptimized{
			data:       bits,
			origOffset: offset,
			position:   offset,
		}
	}

	// Calculate total available bits
	totalBits := uint(len(bits) * 8)
	if offset >= totalBits {
		return &BitsetIteratorOptimized{
			data:       bits,
			origOffset: offset,
			position:   offset,
			totalBits:  totalBits,
		}
	}

	iter := &BitsetIteratorOptimized{
		data:       bits,
		origOffset: offset,
		position:   offset,
		totalBits:  totalBits,
	}

	iter.convertToWords()
	iter.seekToPosition(offset)
	return iter
}

// convertToWords converts the byte slice to 64-bit words for efficient processing.
// We need to maintain the same bit ordering as the byte-based iterator.
func (b *BitsetIteratorOptimized) convertToWords() {
	if len(b.data) == 0 {
		return
	}

	// Pad data to multiple of 8 bytes for clean word conversion
	wordCount := (len(b.data) + 7) / 8
	b.words = make([]uint64, wordCount)

	for i := 0; i < wordCount; i++ {
		startByte := i * 8
		endByte := startByte + 8
		if endByte > len(b.data) {
			endByte = len(b.data)
		}

		// Build word byte by byte, maintaining MSB-first ordering within bytes
		var word uint64
		for j := startByte; j < endByte; j++ {
			byteVal := b.data[j]
			// Each byte is placed in the word with its bits in MSB-first order
			// Byte 0 occupies bits 0-7, byte 1 occupies bits 8-15, etc.
			bitOffset := (j - startByte) * 8

			// Reverse the bits within the byte to match MSB-first expectation
			reversedByte := uint64(0)
			for k := uint(0); k < 8; k++ {
				if byteVal&(0x80>>k) != 0 {
					reversedByte |= uint64(1) << (uint(bitOffset) + k)
				}
			}
			word |= reversedByte
		}
		b.words[i] = word
	}
}

// seekToPosition positions the iterator at the specified bit offset.
func (b *BitsetIteratorOptimized) seekToPosition(offset uint) {
	if len(b.words) == 0 || offset >= b.totalBits {
		b.currentWord = 0
		b.wordIndex = len(b.words)
		b.bitIndex = 64
		b.position = offset
		return
	}

	b.wordIndex = int(offset / 64)
	b.bitIndex = offset % 64
	b.position = offset

	if b.wordIndex < len(b.words) {
		// Load current word and mask off bits before our position
		b.currentWord = b.words[b.wordIndex]
		if b.bitIndex > 0 {
			// Clear bits before our starting position
			mask := ^((uint64(1) << b.bitIndex) - 1)
			b.currentWord &= mask
		}
	}
}

// HasNext returns true if there are more bits to iterate over.
func (b *BitsetIteratorOptimized) HasNext() bool {
	return b.position < b.totalBits
}

// Done returns true if iteration is complete.
func (b *BitsetIteratorOptimized) Done() bool {
	return !b.HasNext()
}

// Current returns the current bit position (0-based index).
func (b *BitsetIteratorOptimized) Current() uint {
	return b.position
}

// Bit returns whether the current bit is set.
func (b *BitsetIteratorOptimized) Bit() uint {
	if !b.HasNext() {
		return 0
	}

	// Check if current bit is set
	if b.wordIndex < len(b.words) {
		bitMask := uint64(1) << b.bitIndex
		if b.words[b.wordIndex]&bitMask != 0 {
			return 1
		}
	}
	return 0
}

// Next advances the iterator to the next bit position.
func (b *BitsetIteratorOptimized) Next() {
	if !b.HasNext() {
		return
	}

	b.position++
	b.bitIndex++

	if b.bitIndex >= 64 {
		// Move to next word
		b.wordIndex++
		b.bitIndex = 0
		if b.wordIndex < len(b.words) {
			b.currentWord = b.words[b.wordIndex]
		}
	}
}

// Begin resets the iterator to the starting position.
func (b *BitsetIteratorOptimized) Begin() {
	b.seekToPosition(b.origOffset)
}

// FindNextSetBit efficiently finds the next set bit using hardware bit scan.
// Returns true if a set bit was found, false otherwise.
func (b *BitsetIteratorOptimized) FindNextSetBit() bool {
	for b.wordIndex < len(b.words) {
		// Mask current word to only consider bits from current position
		word := b.currentWord
		if b.bitIndex > 0 {
			mask := ^((uint64(1) << b.bitIndex) - 1)
			word &= mask
		}

		if word != 0 {
			// Found set bits in current word
			trailingZeros := bits.TrailingZeros64(word)
			b.bitIndex = uint(trailingZeros)
			b.position = uint(b.wordIndex*64) + b.bitIndex

			if b.position < b.totalBits {
				return true
			}
		}

		// Move to next word
		b.wordIndex++
		b.bitIndex = 0
		if b.wordIndex < len(b.words) {
			b.currentWord = b.words[b.wordIndex]
		}
	}

	// No more set bits found
	b.position = b.totalBits
	return false
}

// AdvanceToNextSetBit finds and advances past the next set bit.
func (b *BitsetIteratorOptimized) AdvanceToNextSetBit() bool {
	if b.FindNextSetBit() {
		b.Next()
		return true
	}
	return false
}

// CountSetBits efficiently counts set bits from current position using popcount.
func (b *BitsetIteratorOptimized) CountSetBits() uint {
	count := uint(0)

	if b.wordIndex >= len(b.words) {
		return 0
	}

	// Count bits in current partial word
	word := b.currentWord
	if b.bitIndex > 0 {
		mask := ^((uint64(1) << b.bitIndex) - 1)
		word &= mask
	}

	// Mask out bits beyond total length
	bitsInCurrentWord := 64 - b.bitIndex
	maxBitsFromPosition := b.totalBits - b.position
	if bitsInCurrentWord > maxBitsFromPosition {
		rightShift := bitsInCurrentWord - maxBitsFromPosition
		word >>= rightShift
		word <<= rightShift
	}

	count += uint(bits.OnesCount64(word))

	// Count bits in remaining full words
	for i := b.wordIndex + 1; i < len(b.words); i++ {
		wordBits := b.words[i]

		// Handle last word that might be partial
		if i == len(b.words)-1 {
			remainingBits := b.totalBits % 64
			if remainingBits > 0 {
				mask := (uint64(1) << remainingBits) - 1
				wordBits &= mask
			}
		}

		count += uint(bits.OnesCount64(wordBits))
	}

	return count
}

// SkipZeros efficiently skips over zero bits until a set bit or end is reached.
// This is particularly efficient for sparse bitsets.
func (b *BitsetIteratorOptimized) SkipZeros() bool {
	return b.FindNextSetBit()
}

// CountTrailingZeros counts consecutive zero bits from current position.
func (b *BitsetIteratorOptimized) CountTrailingZeros() uint {
	count := uint(0)
	savedPos := b.position
	savedWordIndex := b.wordIndex
	savedBitIndex := b.bitIndex

	for b.HasNext() {
		if b.Bit() != 0 {
			break
		}
		count++
		b.Next()
	}

	// Restore position
	b.position = savedPos
	b.wordIndex = savedWordIndex
	b.bitIndex = savedBitIndex

	return count
}
