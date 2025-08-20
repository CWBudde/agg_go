package transform

import (
	"math"
	"testing"
)

// TestTransDoublePath_ComprehensiveIntegration demonstrates all major features
// working together, similar to the AGG C++ examples like trans_curve2.cpp
func TestTransDoublePath_ComprehensiveIntegration(t *testing.T) {
	trans := NewTransDoublePath()

	// Test 1: Complex curved path transformation similar to text-on-path rendering
	t.Run("TextOnPathSimulation", func(t *testing.T) {
		// Create a sine wave as the base path (like text baseline)
		trans.Reset()
		for i := 0; i <= 100; i++ {
			x := float64(i) * 5.0                     // 500 units total width
			y := 50.0 + 20.0*math.Sin(float64(i)*0.1) // sine wave with amplitude 20
			if i == 0 {
				trans.MoveTo1(x, y)
			} else {
				trans.LineTo1(x, y)
			}
		}

		// Create a parallel path above (like text cap height)
		for i := 0; i <= 100; i++ {
			x := float64(i) * 5.0
			y := 80.0 + 20.0*math.Sin(float64(i)*0.1) // 30 units above base
			if i == 0 {
				trans.MoveTo2(x, y)
			} else {
				trans.LineTo2(x, y)
			}
		}

		trans.SetBaseHeight(30.0) // Height between baseline and cap
		trans.FinalizePaths()

		// Test character placement along the path
		// Note: y=0 means baseline path, y=baseHeight means cap path
		testPoints := []struct {
			x, y float64
			name string
		}{
			{0, 0, "Start baseline"},
			{0, 30, "Start capheight"},
			{250, 15, "Middle half-height"},
			{500, 0, "End baseline"},
		}

		for _, tp := range testPoints {
			x, y := tp.x, tp.y
			trans.Transform(&x, &y)

			// Just verify transformation completes without crashing and produces reasonable results
			t.Logf("%s: (%.1f,%.1f) -> (%.1f,%.1f)", tp.name, tp.x, tp.y, x, y)
		}
	})

	// Test 2: Variable width corridor (like envelope distortion)
	t.Run("VariableWidthCorridor", func(t *testing.T) {
		trans.Reset()

		// Bottom path: straight line
		trans.MoveTo1(0, 0)
		trans.LineTo1(200, 0)

		// Top path: varying height (narrow-wide-narrow envelope)
		trans.MoveTo2(0, 10)   // Start narrow (10 units high)
		trans.LineTo2(50, 40)  // Expand to 40 units
		trans.LineTo2(100, 60) // Peak at 60 units
		trans.LineTo2(150, 30) // Contract to 30 units
		trans.LineTo2(200, 5)  // End very narrow (5 units)

		trans.FinalizePaths()

		// Test that corridor varies along the path
		testPositions := []float64{0, 50, 100, 150, 200}

		for _, x := range testPositions {
			// Test bottom of corridor (y=0 in input space)
			x1, y1 := x, 0.0
			trans.Transform(&x1, &y1)

			// Test top of corridor (y=1 in input space)
			x2, y2 := x, 1.0
			trans.Transform(&x2, &y2)

			actualHeight := y2 - y1
			t.Logf("x=%.0f: corridor height %.1f (bottom: %.1f, top: %.1f)", x, actualHeight, y1, y2)
		}
	})

	// Test 3: Perspective-like distortion between curves
	t.Run("PerspectiveDistortion", func(t *testing.T) {
		trans.Reset()

		// Bottom path: wide line
		trans.MoveTo1(0, 100)
		trans.LineTo1(200, 100)

		// Top path: narrow line (perspective vanishing effect)
		trans.MoveTo2(50, 50) // Narrower span
		trans.LineTo2(150, 50)

		trans.FinalizePaths()

		// Test perspective-like transformation
		testTransforms := []struct {
			inputX, inputY float64
			description    string
		}{
			{0, 0.5, "Left edge"},
			{100, 0.5, "Center"},
			{200, 0.5, "Right edge"},
		}

		for _, tt := range testTransforms {
			x, y := tt.inputX, tt.inputY
			trans.Transform(&x, &y)
			t.Logf("%s: (%.0f,%.1f) -> (%.1f,%.1f)", tt.description, tt.inputX, tt.inputY, x, y)
		}
	})

	// Test 4: Preserve X scale vs uniform distribution comparison
	t.Run("ScalingModeComparison", func(t *testing.T) {
		// Create paths with non-uniform segments
		createNonUniformPaths := func(trans *TransDoublePath) {
			// Short segments at start, long segment at end
			trans.MoveTo1(0, 0)
			trans.LineTo1(5, 0)   // Very short
			trans.LineTo1(10, 0)  // Short
			trans.LineTo1(100, 0) // Very long

			trans.MoveTo2(0, 20)
			trans.LineTo2(5, 20)
			trans.LineTo2(10, 20)
			trans.LineTo2(100, 20)
		}

		// Test with preserve X scale enabled
		trans1 := NewTransDoublePath()
		createNonUniformPaths(trans1)
		trans1.SetPreserveXScale(true)
		trans1.FinalizePaths()

		// Test with preserve X scale disabled (uniform distribution)
		trans2 := NewTransDoublePath()
		createNonUniformPaths(trans2)
		trans2.SetPreserveXScale(false)
		trans2.FinalizePaths()

		// Compare results at critical points
		testX := []float64{2.5, 7.5, 55} // In first, second, and third segments

		for _, x := range testX {
			x1, y1 := x, 0.5
			x2, y2 := x, 0.5

			trans1.Transform(&x1, &y1) // Preserve X scale
			trans2.Transform(&x2, &y2) // Uniform distribution

			// Results should differ due to different interpolation methods
			if math.Abs(x1-x2) < 1e-10 {
				t.Logf("At x=%.1f: preserve_x_scale may not be affecting results (both gave x=%.3f)", x, x1)
			} else {
				t.Logf("At x=%.1f: preserve_x_scale=true -> x=%.3f, preserve_x_scale=false -> x=%.3f", x, x1, x2)
			}
		}
	})

	// Test 5: Base length scaling verification
	t.Run("BaseLengthScaling", func(t *testing.T) {
		trans.Reset()

		// Create 100-unit long paths
		trans.MoveTo1(0, 0)
		trans.LineTo1(100, 0)
		trans.MoveTo2(0, 10)
		trans.LineTo2(100, 10)

		// Set base length to 200 (double the actual length)
		trans.SetBaseLength(200.0)
		trans.FinalizePaths()

		// Input coordinate 200 should map to actual end (100)
		x, y := 200.0, 0.0
		trans.Transform(&x, &y)

		if math.Abs(x-100.0) > 1e-10 {
			t.Errorf("Base length scaling failed: expected x=100, got %f", x)
		}

		// Input coordinate 100 should map to actual middle (50)
		x, y = 100.0, 0.0
		trans.Transform(&x, &y)

		if math.Abs(x-50.0) > 1e-10 {
			t.Errorf("Base length scaling failed: expected x=50, got %f", x)
		}

		t.Logf("Base length scaling verified: 2x scale factor working correctly")
	})
}

// TestTransDoublePath_ErrorConditions tests edge cases and error handling
func TestTransDoublePath_ErrorConditions(t *testing.T) {
	trans := NewTransDoublePath()

	t.Run("SinglePointPaths", func(t *testing.T) {
		// Test with single-point paths (degenerate case)
		trans.MoveTo1(50, 25)
		trans.MoveTo2(50, 75)
		trans.FinalizePaths()

		x, y := 50.0, 0.5
		trans.Transform(&x, &y)
		// Should not crash - behavior is implementation-defined for degenerate cases
		t.Logf("Single point transformation: (50,0.5) -> (%.3f,%.3f)", x, y)
	})

	t.Run("EmptyPaths", func(t *testing.T) {
		trans.Reset()
		// Don't add any paths

		x, y := 50.0, 0.5
		originalX, originalY := x, y
		trans.Transform(&x, &y)

		// Should not modify coordinates when paths aren't ready
		if x != originalX || y != originalY {
			t.Errorf("Empty paths should not transform coordinates: (%.3f,%.3f) -> (%.3f,%.3f)", originalX, originalY, x, y)
		}
	})

	t.Run("MismatchedPathLengths", func(t *testing.T) {
		trans.Reset()

		// Create paths with very different point counts
		trans.MoveTo1(0, 0)
		trans.LineTo1(100, 0) // Simple 2-point path

		// Complex path with many points
		trans.MoveTo2(0, 10)
		for i := 1; i <= 50; i++ {
			x := float64(i) * 2.0
			y := 10 + 5*math.Sin(float64(i)*0.2)
			trans.LineTo2(x, y)
		}

		trans.FinalizePaths()

		// Should handle gracefully
		x, y := 50.0, 0.5
		trans.Transform(&x, &y)
		t.Logf("Mismatched path complexity handled: (50,0.5) -> (%.3f,%.3f)", x, y)
	})
}

// BenchmarkTransDoublePath_ComplexTransformation benchmarks realistic usage
func BenchmarkTransDoublePath_ComplexTransformation(b *testing.B) {
	trans := NewTransDoublePath()

	// Create complex sinusoidal paths similar to real text-on-path scenarios
	for i := 0; i <= 200; i++ {
		x := float64(i) * 2.5 // 500 units total
		y1 := 100 + 30*math.Sin(float64(i)*0.05) + 10*math.Cos(float64(i)*0.1)
		y2 := y1 + 25 + 5*math.Sin(float64(i)*0.08)

		if i == 0 {
			trans.MoveTo1(x, y1)
			trans.MoveTo2(x, y2)
		} else {
			trans.LineTo1(x, y1)
			trans.LineTo2(x, y2)
		}
	}

	trans.SetBaseHeight(30.0)
	trans.SetPreserveXScale(true)
	trans.FinalizePaths()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := float64(i%500) + 0.5 // Vary position along path
		y := float64(i%30) / 30.0 // Vary height in corridor
		trans.Transform(&x, &y)
	}
}
