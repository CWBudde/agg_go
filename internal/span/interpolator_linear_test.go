package span

import (
	"math"
	"testing"

	"agg_go/internal/transform"
)

func TestDda2LineInterpolator(t *testing.T) {
	t.Run("BasicInterpolation", func(t *testing.T) {
		// Test interpolation from 0 to 100 over 10 steps
		dda := NewDda2LineInterpolator(0, 100, 10)

		if dda.Y() != 0 {
			t.Errorf("Initial Y: got %d, want 0", dda.Y())
		}

		// Step through and check values
		values := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90}
		for i, expected := range values {
			if dda.Y() != expected {
				t.Errorf("Step %d: got %d, want %d", i, dda.Y(), expected)
			}
			if i < len(values)-1 {
				dda.Inc()
			}
		}
	})

	t.Run("NegativeSlope", func(t *testing.T) {
		// Test interpolation from 100 to 0 over 10 steps
		dda := NewDda2LineInterpolator(100, 0, 10)

		if dda.Y() != 100 {
			t.Errorf("Initial Y: got %d, want 100", dda.Y())
		}

		// Check that values decrease
		prev := dda.Y()
		for i := 0; i < 5; i++ {
			dda.Inc()
			curr := dda.Y()
			if curr >= prev {
				t.Errorf("Step %d: value should decrease, got %d (prev %d)", i, curr, prev)
			}
			prev = curr
		}
	})

	t.Run("ZeroCount", func(t *testing.T) {
		// Test with zero count (should default to 1)
		dda := NewDda2LineInterpolator(10, 20, 0)
		if dda.Y() != 10 {
			t.Errorf("With zero count: got %d, want 10", dda.Y())
		}
	})
}

func TestSpanInterpolatorLinear(t *testing.T) {
	t.Run("IdentityTransform", func(t *testing.T) {
		// Test with identity transformation
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Begin interpolation at (0, 0) for 10 pixels
		interp.Begin(0, 0, 10)

		// Check initial coordinates
		x, y := interp.Coordinates()
		if x != 0 || y != 0 {
			t.Errorf("Initial coordinates: got (%d, %d), want (0, 0)", x, y)
		}

		// Step through and verify coordinates advance
		prevX := x
		for i := 0; i < 5; i++ {
			interp.Next()
			x, y = interp.Coordinates()
			if x <= prevX {
				t.Errorf("X coordinate should increase, got %d (prev %d)", x, prevX)
			}
			prevX = x
		}
	})

	t.Run("ScaleTransform", func(t *testing.T) {
		// Test with 2x scale transformation
		trans := transform.NewTransAffine()
		trans.ScaleXY(2.0, 2.0)
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Begin interpolation at (1, 1) for 4 pixels
		interp.Begin(1, 1, 4)

		// Check that coordinates are scaled
		x, y := interp.Coordinates()
		// With 2x scale, (1,1) should become approximately (2,2) in subpixel units
		expectedX := int(2.0 * 256) // 2.0 * subpixel_scale
		expectedY := int(2.0 * 256)

		tolerance := 10 // Allow some tolerance for rounding
		if abs(x-expectedX) > tolerance || abs(y-expectedY) > tolerance {
			t.Errorf("Scaled coordinates: got (%d, %d), want (~%d, ~%d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("TranslationTransform", func(t *testing.T) {
		// Test with translation transformation
		trans := transform.NewTransAffine()
		trans.Translate(10.0, 20.0)
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Begin interpolation at (0, 0) for 2 pixels
		interp.Begin(0, 0, 2)

		// Check that coordinates are translated
		x, y := interp.Coordinates()
		expectedX := int(10.0 * 256) // 10.0 * subpixel_scale
		expectedY := int(20.0 * 256) // 20.0 * subpixel_scale

		tolerance := 10
		if abs(x-expectedX) > tolerance || abs(y-expectedY) > tolerance {
			t.Errorf("Translated coordinates: got (%d, %d), want (~%d, ~%d)", x, y, expectedX, expectedY)
		}
	})

	t.Run("ResynchronizeTest", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Begin interpolation
		interp.Begin(0, 0, 10)

		// Advance a few steps
		for i := 0; i < 3; i++ {
			interp.Next()
		}

		// Resynchronize should update interpolation (basic functionality test)
		interp.Resynchronize(5, 5, 5)

		// Test that we can continue advancing after resync
		beforeX, beforeY := interp.Coordinates()
		interp.Next()
		afterX, afterY := interp.Coordinates()

		// Just verify that coordinates change when we advance
		if beforeX == afterX && beforeY == afterY {
			t.Error("Coordinates should change when advancing after resynchronization")
		}
	})

	t.Run("SubpixelShift", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		if interp.SubpixelShift() != 8 {
			t.Errorf("Default subpixel shift: got %d, want 8", interp.SubpixelShift())
		}

		// Test custom subpixel shift
		interp2 := NewSpanInterpolatorLinear(trans, 6)
		if interp2.SubpixelShift() != 6 {
			t.Errorf("Custom subpixel shift: got %d, want 6", interp2.SubpixelShift())
		}
	})
}

func TestSpanInterpolatorInterface(t *testing.T) {
	// Test that our implementation satisfies the interface
	trans := transform.NewTransAffine()
	var interp SpanInterpolatorInterface = NewSpanInterpolatorLinearDefault(trans)

	// Test interface methods
	interp.Begin(0, 0, 5)
	x, y := interp.Coordinates()
	if x != 0 || y != 0 {
		t.Errorf("Interface coordinates: got (%d, %d), want (0, 0)", x, y)
	}

	interp.Next()
	x2, y2 := interp.Coordinates()
	if x2 == x && y2 == y {
		t.Error("Interface Next() should advance coordinates")
	}

	if interp.SubpixelShift() != 8 {
		t.Errorf("Interface SubpixelShift: got %d, want 8", interp.SubpixelShift())
	}
}

func BenchmarkDda2LineInterpolator(b *testing.B) {
	dda := NewDda2LineInterpolator(0, 1000000, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dda.Inc()
		_ = dda.Y()
	}
}

func BenchmarkSpanInterpolatorLinear(b *testing.B) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)

	b.Run("Begin", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			interp.Begin(0, 0, 1000)
		}
	})

	b.Run("NextAndCoordinates", func(b *testing.B) {
		interp.Begin(0, 0, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			interp.Next()
			_, _ = interp.Coordinates()
		}
	})
}

func BenchmarkTransformOperations(b *testing.B) {
	trans := transform.NewTransAffine()
	trans.Rotate(math.Pi / 4) // 45 degrees
	trans.ScaleXY(2.0, 1.5)
	trans.Translate(100, 200)

	interp := NewSpanInterpolatorLinearDefault(trans)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Begin(float64(i%100), float64(i%100), 10)
		for j := 0; j < 10; j++ {
			interp.Next()
			_, _ = interp.Coordinates()
		}
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
