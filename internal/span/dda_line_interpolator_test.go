package span

import "testing"

func TestDDALineInterpolatorBasic(t *testing.T) {
	// Test linear interpolation from 0 to 100 over 10 steps
	dda := NewDDALineInterpolator(0, 100, 10, 8) // Using 8-bit precision

	expectedValues := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90}

	for i, expected := range expectedValues {
		actual := dda.Y()
		// Allow small tolerance for fixed-point arithmetic
		if actual < expected-2 || actual > expected+2 {
			t.Errorf("Step %d: expected ~%d, got %d", i, expected, actual)
		}
		dda.Inc()
	}
}

func TestDDALineInterpolatorHighPrecision(t *testing.T) {
	// Test with 14-bit precision (typical for color interpolation)
	dda := NewDDALineInterpolator(0, 255, 256, 14)

	// Check initial value
	if dda.Y() != 0 {
		t.Errorf("Expected initial value 0, got %d", dda.Y())
	}

	// Advance halfway (128 steps)
	for i := 0; i < 128; i++ {
		dda.Inc()
	}

	val := dda.Y()
	// Should be approximately 127-128
	if val < 126 || val > 129 {
		t.Errorf("Expected value around 127-128 at halfway point, got %d", val)
	}
}

func TestDDALineInterpolatorDecrement(t *testing.T) {
	dda := NewDDALineInterpolator(0, 100, 10, 8)

	// Advance several steps
	for i := 0; i < 5; i++ {
		dda.Inc()
	}
	valForward := dda.Y()

	// Step back one
	dda.Dec()
	valBack := dda.Y()

	if valBack >= valForward {
		t.Errorf("Dec should decrease value: %d >= %d", valBack, valForward)
	}
}

func TestDDALineInterpolatorAddSub(t *testing.T) {
	dda1 := NewDDALineInterpolator(0, 100, 10, 8)
	dda2 := NewDDALineInterpolator(0, 100, 10, 8)

	// Advance one by single steps
	for i := 0; i < 3; i++ {
		dda1.Inc()
	}
	val1 := dda1.Y()

	// Advance the other by Add(3)
	dda2.Add(3)
	val2 := dda2.Y()

	// Should be the same (within tolerance)
	if val1 != val2 {
		t.Errorf("Inc() x3 and Add(3) should give same result: %d vs %d", val1, val2)
	}

	// Test Sub
	dda2.Sub(1)
	val3 := dda2.Y()

	if val3 >= val2 {
		t.Errorf("Sub should decrease value: %d >= %d", val3, val2)
	}
}

func TestDDALineInterpolatorZeroCount(t *testing.T) {
	// Test with zero count (should be treated as 1)
	dda := NewDDALineInterpolator(10, 20, 0, 8)

	// Should not crash and should start at initial value
	if dda.Y() != 10 {
		t.Errorf("Expected initial value 10, got %d", dda.Y())
	}

	// Inc should work
	dda.Inc()
	val := dda.Y()
	if val <= 10 {
		t.Errorf("Expected value to increase after Inc(), got %d", val)
	}
}

func TestDDALineInterpolatorNegativeRange(t *testing.T) {
	// Test interpolation from 100 to 0 (decreasing)
	dda := NewDDALineInterpolator(100, 0, 10, 8)

	initial := dda.Y()
	if initial != 100 {
		t.Errorf("Expected initial value 100, got %d", initial)
	}

	// Advance halfway
	for i := 0; i < 5; i++ {
		dda.Inc()
	}

	mid := dda.Y()
	if mid > initial {
		t.Errorf("Expected decreasing values: %d should be < %d", mid, initial)
	}
}

func TestDDALineInterpolatorFractionShift(t *testing.T) {
	// Test different fraction shifts
	shifts := []int{4, 8, 12, 14, 16}

	for _, shift := range shifts {
		dda := NewDDALineInterpolator(0, 1000, 100, shift)

		// Should not crash with different precision levels
		for i := 0; i < 10; i++ {
			dda.Inc()
		}

		val := dda.Y()
		if val < 0 || val > 1000 {
			t.Errorf("With shift %d, got unreasonable value %d", shift, val)
		}
	}
}

func TestDDALineInterpolatorDYAccess(t *testing.T) {
	dda := NewDDALineInterpolator(0, 100, 10, 8)

	// Initial DY should be 0
	if dda.DY() != 0 {
		t.Errorf("Expected initial DY 0, got %d", dda.DY())
	}

	// After Inc, DY should change
	dda.Inc()
	if dda.DY() == 0 {
		t.Errorf("Expected DY to change after Inc()")
	}
}

// Benchmark tests
func BenchmarkDDALineInterpolatorInc(b *testing.B) {
	dda := NewDDALineInterpolator(0, 255, 256, 14)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dda.Inc()
	}
}

func BenchmarkDDALineInterpolatorY(b *testing.B) {
	dda := NewDDALineInterpolator(0, 255, 256, 14)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dda.Y()
	}
}
