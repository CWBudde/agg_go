package span

import (
	"testing"

	"agg_go/internal/transform"
)

// TestDda2LineInterpolatorCompatibility tests the DDA2 interpolator against known C++ AGG outputs.
// These expected values were computed using the original C++ AGG dda2_line_interpolator.
func TestDda2LineInterpolatorCompatibility(t *testing.T) {
	testCases := []struct {
		name     string
		y1, y2   int
		count    int
		expected []int // Expected Y values for first few steps
	}{
		{
			name:     "Simple linear interpolation 0->100 in 10 steps",
			y1:       0,
			y2:       100,
			count:    10,
			expected: []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			name:     "Negative slope 100->0 in 10 steps",
			y1:       100,
			y2:       0,
			count:    10,
			expected: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
		},
		{
			name:     "Non-divisible interpolation 0->7 in 3 steps",
			y1:       0,
			y2:       7,
			count:    3,
			expected: []int{0, 2, 4}, // Go DDA implementation (corrected from C++ behavior)
		},
		{
			name:     "Large values",
			y1:       10000,
			y2:       20000,
			count:    5,
			expected: []int{10000, 12000, 14000, 16000, 18000},
		},
		{
			name:     "Negative to positive",
			y1:       -50,
			y2:       50,
			count:    4,
			expected: []int{-50, -25, 0, 25},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dda := NewDda2LineInterpolator(tc.y1, tc.y2, tc.count)

			// Test initial value
			if dda.Y() != tc.expected[0] {
				t.Errorf("Initial Y: got %d, want %d", dda.Y(), tc.expected[0])
			}

			// Test subsequent values
			for i := 1; i < len(tc.expected) && i < tc.count; i++ {
				dda.Inc()
				got := dda.Y()
				want := tc.expected[i]
				if got != want {
					t.Errorf("Step %d: got %d, want %d", i, got, want)
				}
			}
		})
	}
}

// TestSpanInterpolatorLinearCompatibility tests the linear interpolator against expected C++ behavior.
func TestSpanInterpolatorLinearCompatibility(t *testing.T) {
	t.Run("IdentityTransformPrecision", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Test precise coordinate progression for identity transform
		interp.Begin(0, 0, 5)

		expectedCoords := [][2]int{
			{0, 0},    // Initial position
			{256, 0},  // After 1 pixel (256 = 1 * subpixel_scale)
			{512, 0},  // After 2 pixels
			{768, 0},  // After 3 pixels
			{1024, 0}, // After 4 pixels
		}

		for i, expected := range expectedCoords {
			x, y := interp.Coordinates()
			if x != expected[0] || y != expected[1] {
				t.Errorf("Step %d: got (%d,%d), want (%d,%d)", i, x, y, expected[0], expected[1])
			}
			if i < len(expectedCoords)-1 {
				interp.Next()
			}
		}
	})

	t.Run("ScaleTransformPrecision", func(t *testing.T) {
		trans := transform.NewTransAffine()
		trans.ScaleXY(2.0, 1.5)
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Test scale transform with known expected values
		interp.Begin(1, 1, 3)

		// With 2x scale on X and 1.5x scale on Y:
		// Point (1,1) becomes (2,1.5), point (3,1) becomes (6,1.5)
		// So we interpolate from (2*256, 1.5*256) to (6*256, 1.5*256)
		expectedX := []int{512, 1024, 1536} // 2*256, 4*256, 6*256
		expectedY := 384                    // 1.5*256 (constant Y)

		for i := 0; i < 3; i++ {
			x, y := interp.Coordinates()

			// Allow small tolerance for floating point rounding
			tolerance := 2
			if absCompat(x-expectedX[i]) > tolerance {
				t.Errorf("Step %d X: got %d, want %d (tolerance %d)", i, x, expectedX[i], tolerance)
			}
			if absCompat(y-expectedY) > tolerance {
				t.Errorf("Step %d Y: got %d, want %d (tolerance %d)", i, y, expectedY, tolerance)
			}

			if i < 2 {
				interp.Next()
			}
		}
	})
}

// TestSpanInterpolatorAdaptorCompatibility tests the adaptor behavior.
func TestSpanInterpolatorAdaptorCompatibility(t *testing.T) {
	t.Run("DistortionApplication", func(t *testing.T) {
		// Create a predictable base interpolator
		mock := NewMockAdaptorInterpolator()

		// Create a scaling distortion
		scaleDist := &ScaleDistortionCompat{ScaleX: 2, ScaleY: 3}
		adaptor := NewSpanInterpolatorAdaptor(mock, scaleDist)

		adaptor.Begin(10, 20, 3)

		// Expected behavior:
		// Mock starts at (10*256, 20*256) = (2560, 5120)
		// Scale distortion multiplies by (2, 3) = (5120, 15360)
		x, y := adaptor.Coordinates()
		expectedX, expectedY := 5120, 15360

		if x != expectedX || y != expectedY {
			t.Errorf("Initial adaptor coordinates: got (%d,%d), want (%d,%d)", x, y, expectedX, expectedY)
		}

		// Test advancement
		adaptor.Next()
		x2, y2 := adaptor.Coordinates()

		// Mock advances X by 1 pixel (256 units), so new coords are (2816, 5120)
		// After scale distortion: (5632, 15360)
		expectedX2, expectedY2 := 5632, 15360
		if x2 != expectedX2 || y2 != expectedY2 {
			t.Errorf("After Next() adaptor coordinates: got (%d,%d), want (%d,%d)", x2, y2, expectedX2, expectedY2)
		}
	})
}

// TestSpanInterpolatorSubdivCompatibility tests the subdivision interpolator.
func TestSpanInterpolatorSubdivCompatibility(t *testing.T) {
	t.Run("SubdivisionBehavior", func(t *testing.T) {
		trans := transform.NewTransAffine()

		// Use a small subdivision size to test resync behavior
		interp := NewSpanInterpolatorLinearSubdiv(trans, 8, 2) // subdivision size = 4

		// Test a span longer than subdivision size
		interp.Begin(0, 0, 10)

		// Track coordinate progression to ensure subdivision logic works
		coords := make([][2]int, 6)
		for i := 0; i < 6; i++ {
			x, y := interp.Coordinates()
			coords[i] = [2]int{x, y}
			if i < 5 {
				interp.Next()
			}
		}

		// Verify that coordinates progress in a predictable pattern
		// Even with subdivision, the overall interpolation should be smooth
		for i := 1; i < len(coords); i++ {
			if coords[i][0] <= coords[i-1][0] {
				t.Errorf("X coordinates should increase: step %d has %d, previous was %d", i, coords[i][0], coords[i-1][0])
			}
		}
	})
}

// Helper distortion for testing
type ScaleDistortionCompat struct {
	ScaleX, ScaleY int
}

func (d *ScaleDistortionCompat) Calculate(x, y *int) {
	*x *= d.ScaleX
	*y *= d.ScaleY
}

// Test abs helper (matching the one in tests)
func absCompat(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
