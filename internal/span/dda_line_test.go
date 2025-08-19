package span

import (
	"testing"
)

func TestGouraudDDAInterpolator(t *testing.T) {
	tests := []struct {
		name     string
		y1, y2   int
		count    uint
		expected []int
	}{
		{
			name:     "Simple interpolation",
			y1:       0,
			y2:       100,
			count:    5,
			expected: []int{0, 20, 40, 60, 80},
		},
		{
			name:     "Negative range",
			y1:       100,
			y2:       0,
			count:    4,
			expected: []int{100, 75, 50, 25},
		},
		{
			name:     "Single step",
			y1:       10,
			y2:       20,
			count:    1,
			expected: []int{10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewGouraudDDAInterpolator(tt.y1, tt.y2, tt.count, 10)

			for i, expected := range tt.expected {
				actual := interp.Y()
				if actual != expected {
					t.Errorf("Step %d: expected Y=%d, got Y=%d", i, expected, actual)
				}
				interp.Inc()
			}
		})
	}
}

func TestGouraudDDAInterpolatorZeroCount(t *testing.T) {
	// Test zero count handling
	interp := NewGouraudDDAInterpolator(0, 100, 0, 8)
	if interp.Y() != 0 {
		t.Errorf("Expected Y=0 for zero count, got Y=%d", interp.Y())
	}
}

func TestGouraudDDAInterpolatorOperations(t *testing.T) {
	interp := NewGouraudDDAInterpolator(0, 256, 16, 8)

	// Test increment
	y1 := interp.Y()
	interp.Inc()
	y2 := interp.Y()
	if y2 <= y1 {
		t.Errorf("Inc() should increase Y value: %d -> %d", y1, y2)
	}

	// Test decrement
	interp.Dec()
	y3 := interp.Y()
	if y3 != y1 {
		t.Errorf("Dec() should restore Y value: expected %d, got %d", y1, y3)
	}

	// Test add
	interp.Add(4)
	interp.Sub(4)
	y5 := interp.Y()
	if y5 != y1 {
		t.Errorf("Add/Sub should be reversible: expected %d, got %d", y1, y5)
	}

	// Test DY accessor
	dy := interp.DY()
	if dy < 0 {
		t.Errorf("DY should be non-negative initially, got %d", dy)
	}
}

func TestGouraudDda2LineInterpolator(t *testing.T) {
	tests := []struct {
		name     string
		y1, y2   int
		count    int
		expected []int
	}{
		{
			name:     "Basic interpolation",
			y1:       0,
			y2:       10,
			count:    5,
			expected: []int{0, 2, 4, 6, 8},
		},
		{
			name:     "With remainder",
			y1:       0,
			y2:       7,
			count:    3,
			expected: []int{0, 2, 4},
		},
		{
			name:     "Negative delta",
			y1:       10,
			y2:       0,
			count:    5,
			expected: []int{10, 8, 6, 4, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := NewGouraudDda2Interpolator(tt.y1, tt.y2, tt.count)

			for i, expected := range tt.expected {
				actual := interp.Y()
				if actual != expected {
					t.Errorf("Step %d: expected Y=%d, got Y=%d", i, expected, actual)
				}
				interp.Inc()
			}
		})
	}
}

func TestGouraudDda2LineInterpolatorZeroCount(t *testing.T) {
	// Test zero count handling
	interp := NewGouraudDda2Interpolator(0, 100, 0)
	if interp.Y() != 0 {
		t.Errorf("Expected Y=0 for zero count, got Y=%d", interp.Y())
	}
}

func TestGouraudDda2LineInterpolatorBackward(t *testing.T) {
	interp := NewGouraudDda2InterpolatorBackward(0, 10, 5)
	expected := []int{0, 2, 4, 6, 8}

	for i, exp := range expected {
		actual := interp.Y()
		if actual != exp {
			t.Errorf("Step %d: expected Y=%d, got Y=%d", i, exp, actual)
		}
		interp.Inc()
	}
}

func TestGouraudDda2LineInterpolatorSimple(t *testing.T) {
	interp := NewGouraudDda2InterpolatorSimple(20, 4)
	expected := []int{0, 5, 10, 15}

	for i, exp := range expected {
		actual := interp.Y()
		if actual != exp {
			t.Errorf("Step %d: expected Y=%d, got Y=%d", i, exp, actual)
		}
		interp.Inc()
	}
}

func TestGouraudDda2LineInterpolatorOperations(t *testing.T) {
	interp := NewGouraudDda2Interpolator(0, 100, 10)

	// Test increment/decrement
	y1 := interp.Y()
	interp.Inc()
	y2 := interp.Y()
	interp.Dec()
	y3 := interp.Y()

	if y2 <= y1 {
		t.Errorf("Inc() should increase Y: %d -> %d", y1, y2)
	}
	if y3 != y1 {
		t.Errorf("Dec() should restore Y: expected %d, got %d", y1, y3)
	}

	// Test accessors
	mod := interp.Mod()
	rem := interp.Rem()

	if rem < 0 {
		t.Errorf("Rem should be non-negative, got %d", rem)
	}

	// Mod can be negative due to adjustment
	_ = mod // Just verify it doesn't panic
}

func TestGouraudDda2LineInterpolatorSaveLoad(t *testing.T) {
	interp := NewGouraudDda2Interpolator(0, 100, 10)

	// Advance a few steps
	interp.Inc()
	interp.Inc()
	interp.Inc()

	// Save state
	saved := interp.Save()
	y1 := interp.Y()

	// Continue advancing
	interp.Inc()
	interp.Inc()

	// Restore state
	interp.Load(saved)
	y2 := interp.Y()

	if y1 != y2 {
		t.Errorf("Save/Load should restore state: expected Y=%d, got Y=%d", y1, y2)
	}
}

func BenchmarkGouraudDDAInterpolator(b *testing.B) {
	interp := NewGouraudDDAInterpolator(0, 1000000, 1000, 14)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Inc()
		_ = interp.Y()
	}
}

func BenchmarkGouraudDda2LineInterpolator(b *testing.B) {
	interp := NewGouraudDda2Interpolator(0, 1000000, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Inc()
		_ = interp.Y()
	}
}
