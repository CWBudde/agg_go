package span

import (
	"testing"
)

func TestSpanAllocator_Allocate(t *testing.T) {
	alloc := NewSpanAllocator()

	t.Run("basic allocation", func(t *testing.T) {
		colors := alloc.Allocate(10)
		if len(colors) != 10 {
			t.Errorf("Expected length 10, got %d", len(colors))
		}

		// All elements should be nil initially
		for i, color := range colors {
			if color != nil {
				t.Errorf("Expected nil at index %d, got %v", i, color)
			}
		}
	})

	t.Run("reallocation with larger size", func(t *testing.T) {
		colors1 := alloc.Allocate(5)
		colors2 := alloc.Allocate(20)

		if len(colors2) != 20 {
			t.Errorf("Expected length 20, got %d", len(colors2))
		}

		// Previous allocation should be invalidated
		_ = colors1 // Just to avoid unused variable warning
	})

	t.Run("reallocation with smaller size", func(t *testing.T) {
		alloc.Allocate(50) // Grow the buffer
		colors := alloc.Allocate(10)

		if len(colors) != 10 {
			t.Errorf("Expected length 10, got %d", len(colors))
		}
	})
}

func TestSpanAllocator_ZeroLength(t *testing.T) {
	alloc := NewSpanAllocator()
	colors := alloc.Allocate(0)

	if len(colors) != 0 {
		t.Errorf("Expected length 0, got %d", len(colors))
	}
}
