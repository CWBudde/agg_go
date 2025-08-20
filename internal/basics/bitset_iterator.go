package basics

import "math/bits"

// BitsetIterator provides efficient iteration over set bits in bitmask data structures.
// This is primarily used in font rendering for processing monochrome bitmaps.
type BitsetIterator struct {
	bits       []byte
	mask       byte
	origBits   []byte // Original slice for Begin() method
	origOffset uint   // Original offset for Begin() method
	position   uint   // Current bit position (0-based)
}

// NewBitsetIterator creates a new bitset iterator starting at the specified bit offset.
// The offset parameter specifies which bit to start from (0-based indexing).
func NewBitsetIterator(bits []byte, offset uint) *BitsetIterator {
	if len(bits) == 0 {
		return &BitsetIterator{
			bits:       nil,
			mask:       0,
			origBits:   bits,
			origOffset: offset,
			position:   offset,
		}
	}

	byteOffset := offset >> 3 // offset / 8
	bitOffset := offset & 7   // offset % 8

	// Ensure we don't go past the end of the slice
	if int(byteOffset) >= len(bits) {
		return &BitsetIterator{
			bits:       nil,
			mask:       0,
			origBits:   bits,
			origOffset: offset,
			position:   offset,
		}
	}

	return &BitsetIterator{
		bits:       bits[byteOffset:],
		mask:       0x80 >> bitOffset, // Start with MSB and shift right
		origBits:   bits,
		origOffset: offset,
		position:   offset,
	}
}

// Next advances the iterator to the next bit position.
func (b *BitsetIterator) Next() {
	if len(b.bits) == 0 {
		return
	}

	b.position++
	b.mask >>= 1
	if b.mask == 0 {
		// Move to next byte
		b.bits = b.bits[1:]
		b.mask = 0x80
	}
}

// Bit returns whether the current bit is set (non-zero).
// Returns 0 if the bit is not set, non-zero if it is set.
func (b *BitsetIterator) Bit() uint {
	if len(b.bits) == 0 {
		return 0
	}

	if b.bits[0]&b.mask != 0 {
		return 1
	}
	return 0
}

// HasNext returns true if there are more bits to iterate over.
func (b *BitsetIterator) HasNext() bool {
	return len(b.bits) > 0
}

// Begin resets the iterator to the starting position.
func (b *BitsetIterator) Begin() {
	if len(b.origBits) == 0 {
		b.bits = nil
		b.mask = 0
		b.position = b.origOffset
		return
	}

	byteOffset := b.origOffset >> 3 // offset / 8
	bitOffset := b.origOffset & 7   // offset % 8

	// Ensure we don't go past the end of the slice
	if int(byteOffset) >= len(b.origBits) {
		b.bits = nil
		b.mask = 0
		b.position = b.origOffset
		return
	}

	b.bits = b.origBits[byteOffset:]
	b.mask = 0x80 >> bitOffset
	b.position = b.origOffset
}

// Current returns the current bit position (0-based index).
func (b *BitsetIterator) Current() uint {
	return b.position
}

// Done returns true if iteration is complete (inverse of HasNext).
func (b *BitsetIterator) Done() bool {
	return !b.HasNext()
}

// FindNextSetBit advances the iterator to the next set bit.
// Returns true if a set bit was found, false if no more set bits exist.
func (b *BitsetIterator) FindNextSetBit() bool {
	for b.HasNext() {
		if b.Bit() != 0 {
			return true
		}
		b.Next()
	}
	return false
}

// CountSetBits counts the number of set bits from current position to end.
// This method does not advance the iterator position.
func (b *BitsetIterator) CountSetBits() uint {
	count := uint(0)

	if len(b.bits) == 0 {
		return 0
	}

	// Count bits in current partial byte
	currentByte := b.bits[0]
	currentMask := b.mask
	for currentMask != 0 {
		if currentByte&currentMask != 0 {
			count++
		}
		currentMask >>= 1
	}

	// Count bits in remaining full bytes using optimized bit counting
	for i := 1; i < len(b.bits); i++ {
		count += uint(bits.OnesCount8(b.bits[i]))
	}

	return count
}

// AdvanceToNextSetBit is like FindNextSetBit but advances past the found bit.
// Returns true if a set bit was found and skipped, false if no more set bits exist.
func (b *BitsetIterator) AdvanceToNextSetBit() bool {
	if b.FindNextSetBit() {
		b.Next()
		return true
	}
	return false
}
