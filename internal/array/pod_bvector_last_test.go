package array

import (
	"testing"
)

func TestPodBVectorLast(t *testing.T) {
	bv := NewPodBVector[int]()

	// Test Last() on empty vector
	if last := bv.Last(); last != nil {
		t.Error("Last() on empty vector should return nil")
	}

	// Add single element
	bv.Add(42)
	if last := bv.Last(); last == nil {
		t.Error("Last() returned nil for non-empty vector")
	} else if *last != 42 {
		t.Errorf("Expected *Last() = 42, got %d", *last)
	}

	// Add more elements
	bv.Add(100)
	bv.Add(200)

	if last := bv.Last(); last == nil {
		t.Error("Last() returned nil")
	} else if *last != 200 {
		t.Errorf("Expected *Last() = 200, got %d", *last)
	}

	// Test modification through Last()
	*bv.Last() = 999
	if bv.ValueAt(bv.Size()-1) != 999 {
		t.Error("Modifying through Last() didn't work")
	}

	// Test after RemoveLast
	bv.RemoveLast()
	if last := bv.Last(); last == nil {
		t.Error("Last() returned nil after RemoveLast")
	} else if *last != 100 {
		t.Errorf("Expected *Last() = 100 after RemoveLast, got %d", *last)
	}

	// Test clear
	bv.Clear()
	if last := bv.Last(); last != nil {
		t.Error("Last() should return nil after Clear()")
	}
}

func TestPodBVectorLastDifferentBlockSizes(t *testing.T) {
	// Test with small block size
	scale := NewBlockScale(2) // Block size = 4
	bv := NewPodBVectorWithScale[int](scale)

	// Add enough elements to span multiple blocks
	for i := 0; i < 10; i++ {
		bv.Add(i * 10)
	}

	if last := bv.Last(); last == nil {
		t.Error("Last() returned nil")
	} else if *last != 90 {
		t.Errorf("Expected *Last() = 90, got %d", *last)
	}

	// Test that Last() works across block boundaries
	for i := bv.Size() - 1; i >= 0; i-- {
		if last := bv.Last(); last == nil {
			t.Errorf("Last() returned nil at size %d", bv.Size())
		} else if *last != (bv.Size()-1)*10 {
			t.Errorf("Expected *Last() = %d, got %d", (bv.Size()-1)*10, *last)
		}
		bv.RemoveLast()
	}

	// Should be empty now
	if last := bv.Last(); last != nil {
		t.Error("Last() should return nil for empty vector")
	}
}
