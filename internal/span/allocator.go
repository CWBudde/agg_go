// Package span provides span allocation and generation functionality for AGG.
// This package implements the span allocator and generator interfaces used
// by scanline renderers for anti-aliased rendering with varying colors.
package span

// SpanAllocator provides basic span allocation functionality.
// This is a simple implementation of the SpanAllocatorInterface
// that allocates color arrays for scanline rendering.
type SpanAllocator[C SpanColorType] struct {
	buffer []C // Reusable buffer for color allocation
}

// NewSpanAllocator creates a new span allocator.
func NewSpanAllocator[C SpanColorType]() *SpanAllocator[C] {
	return &SpanAllocator[C]{
		buffer: make([]C, 0, 256), // Start with reasonable capacity
	}
}

// Allocate allocates an array of colors for the given length.
// Returns a slice that can hold 'len' color values.
// The returned slice is valid until the next call to Allocate.
func (sa *SpanAllocator[C]) Allocate(length int) []C {
	// Ensure buffer has enough capacity
	if cap(sa.buffer) < length {
		sa.buffer = make([]C, length, length*2)
	} else {
		sa.buffer = sa.buffer[:length]
	}

	// Clear the buffer (set all elements to zero value)
	var zero C
	for i := range sa.buffer {
		sa.buffer[i] = zero
	}

	return sa.buffer
}
