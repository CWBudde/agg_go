package gamma

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestNewGammaSpline(t *testing.T) {
	gs := NewGammaSpline()
	if gs == nil {
		t.Fatal("NewGammaSpline returned nil")
	}

	// Check default identity curve
	if !gs.IsIdentity(0.01) {
		t.Error("Default gamma spline should be identity")
	}

	// Test that gamma table is properly initialized
	gamma := gs.Gamma()
	if len(gamma) != 256 {
		t.Errorf("Expected gamma table length 256, got %d", len(gamma))
	}

	// Identity gamma should have gamma[i] ≈ i
	for i := 0; i < 256; i++ {
		expected := uint8(i)
		actual := gamma[i]
		if abs(int(actual)-int(expected)) > 1 { // Allow 1 unit tolerance for rounding
			t.Errorf("Identity gamma[%d]: expected ~%d, got %d", i, expected, actual)
		}
	}
}

func TestGammaSplineValues(t *testing.T) {
	gs := NewGammaSpline()

	tests := []struct {
		name                string
		kx1, ky1, kx2, ky2  float64
		expectedConstraints bool
	}{
		{"Identity curve", 1.0, 1.0, 1.0, 1.0, true},
		{"Bright curve", 0.5, 1.5, 0.5, 1.5, true},
		{"Dark curve", 1.5, 0.5, 1.5, 0.5, true},
		{"Edge case low", 0.002, 0.002, 0.002, 0.002, true},
		{"Edge case high", 1.999, 1.999, 1.999, 1.999, true},
		{"Out of bounds - should clamp", -1.0, 3.0, -0.5, 2.5, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs.Values(test.kx1, test.ky1, test.kx2, test.ky2)

			// Verify we can retrieve values
			gotKx1, gotKy1, gotKx2, gotKy2 := gs.GetValues()

			// Values should be within valid range after clamping (allow small epsilon)
			epsilon := 0.0001
			if gotKx1 < 0.001-epsilon || gotKx1 > 1.999+epsilon {
				t.Errorf("kx1 out of range: %f", gotKx1)
			}
			if gotKy1 < 0.001-epsilon || gotKy1 > 1.999+epsilon {
				t.Errorf("ky1 out of range: %f", gotKy1)
			}
			if gotKx2 < 0.001-epsilon || gotKx2 > 1.999+epsilon {
				t.Errorf("kx2 out of range: %f", gotKx2)
			}
			if gotKy2 < 0.001-epsilon || gotKy2 > 1.999+epsilon {
				t.Errorf("ky2 out of range: %f", gotKy2)
			}

			// For identity curve, check that output matches input
			if test.name == "Identity curve" {
				for i := 0; i <= 10; i++ {
					x := float64(i) / 10.0
					y := gs.Y(x)
					if math.Abs(y-x) > 0.01 {
						t.Errorf("Identity curve Y(%f) = %f, expected ~%f", x, y, x)
					}
				}
			}
		})
	}
}

func TestGammaSplineY(t *testing.T) {
	gs := NewGammaSpline()

	// Test boundary conditions
	tests := []struct {
		input     float64
		expected  float64
		tolerance float64
	}{
		{0.0, 0.0, 0.001},  // Start point
		{1.0, 1.0, 0.001},  // End point
		{-1.0, 0.0, 0.001}, // Below range - should clamp to 0
		{2.0, 1.0, 0.001},  // Above range - should clamp to 1
		{0.5, 0.5, 0.1},    // Middle point for identity
	}

	for _, test := range tests {
		result := gs.Y(test.input)
		if math.Abs(result-test.expected) > test.tolerance {
			t.Errorf("Y(%f) = %f, expected %f ± %f",
				test.input, result, test.expected, test.tolerance)
		}
	}

	// Test monotonicity for identity curve
	prev := gs.Y(0.0)
	for i := 1; i <= 100; i++ {
		x := float64(i) / 100.0
		current := gs.Y(x)
		if current < prev-0.001 { // Allow small numerical errors
			t.Errorf("Gamma curve should be monotonic: Y(%f) = %f < Y(%f) = %f",
				x, current, x-0.01, prev)
		}
		prev = current
	}
}

func TestGammaSplineApplyGamma(t *testing.T) {
	gs := NewGammaSpline()

	// Test 8-bit gamma application
	for i := 0; i < 256; i++ {
		input := uint8(i)
		output := gs.ApplyGamma(input)

		// For identity curve, output should be close to input
		if abs(int(output)-int(input)) > 1 {
			t.Errorf("ApplyGamma(%d) = %d, expected ~%d for identity curve",
				input, output, input)
		}
	}

	// Test floating-point gamma application
	for i := 0; i <= 10; i++ {
		input := float64(i) / 10.0
		output := gs.ApplyGammaFloat(input)

		// For identity curve, output should equal input
		if math.Abs(output-input) > 0.01 {
			t.Errorf("ApplyGammaFloat(%f) = %f, expected ~%f for identity curve",
				input, output, input)
		}
	}
}

func TestGammaSplineBox(t *testing.T) {
	gs := NewGammaSpline()

	// Set rendering box
	x1, y1, x2, y2 := 10.0, 20.0, 100.0, 200.0
	gs.Box(x1, y1, x2, y2)

	// Generate a few vertices to test box bounds
	gs.Rewind(0)

	// First vertex should be at start of box
	x, y, cmd := gs.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", cmd)
	}
	if x != x1 || y != y1 {
		t.Errorf("First vertex should be (%f, %f), got (%f, %f)", x1, y1, x, y)
	}

	// Generate several more vertices
	for i := 0; i < 10; i++ {
		x, y, cmd = gs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		// Vertices should be within box bounds
		if x < x1 || x > x2 || y < y1 || y > y2 {
			t.Errorf("Vertex (%f, %f) outside box [%f,%f,%f,%f]",
				x, y, x1, y1, x2, y2)
		}
	}
}

func TestGammaSplineVertexGeneration(t *testing.T) {
	gs := NewGammaSpline()
	gs.Box(0, 0, 100, 100)

	// Test vertex generation
	gs.Rewind(0)

	vertexCount := 0
	hasMoveTo := false
	hasLineTo := false

	for {
		x, y, cmd := gs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		vertexCount++
		if cmd == basics.PathCmdMoveTo {
			hasMoveTo = true
		} else if cmd == basics.PathCmdLineTo {
			hasLineTo = true
		}

		// Coordinates should be reasonable
		if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
			t.Errorf("Invalid vertex coordinates: (%f, %f)", x, y)
		}

		// Prevent infinite loops
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated - possible infinite loop")
		}
	}

	if vertexCount == 0 {
		t.Error("No vertices generated")
	}
	if !hasMoveTo {
		t.Error("No MoveTo command found")
	}
	if !hasLineTo {
		t.Error("No LineTo commands found")
	}
}

func TestGammaSplineGetCurvePoints(t *testing.T) {
	gs := NewGammaSpline()

	tests := []struct {
		numPoints int
		shouldErr bool
	}{
		{0, true},    // Invalid
		{-1, true},   // Invalid
		{1, false},   // Single point
		{10, false},  // Normal case
		{100, false}, // Many points
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			xPoints, yPoints := gs.GetCurvePoints(test.numPoints)

			if test.shouldErr {
				if xPoints != nil || yPoints != nil {
					t.Error("Expected nil for invalid numPoints")
				}
				return
			}

			if len(xPoints) != test.numPoints || len(yPoints) != test.numPoints {
				t.Errorf("Expected %d points, got %d x-points and %d y-points",
					test.numPoints, len(xPoints), len(yPoints))
			}

			// Check that points are in ascending X order and valid range
			for i := 0; i < len(xPoints); i++ {
				if xPoints[i] < 0.0 || xPoints[i] > 1.0 {
					t.Errorf("X point %d out of range [0,1]: %f", i, xPoints[i])
				}
				if yPoints[i] < 0.0 || yPoints[i] > 1.0 {
					t.Errorf("Y point %d out of range [0,1]: %f", i, yPoints[i])
				}

				if i > 0 && xPoints[i] < xPoints[i-1] {
					t.Errorf("X points not ascending: %f >= %f", xPoints[i-1], xPoints[i])
				}
			}
		})
	}
}

func TestGammaSplineIsIdentity(t *testing.T) {
	gs := NewGammaSpline()

	// Default should be identity
	if !gs.IsIdentity(0.01) {
		t.Error("Default curve should be identity")
	}

	// Set a non-identity curve
	gs.Values(0.5, 1.5, 0.5, 1.5) // Bright curve
	if gs.IsIdentity(0.01) {
		t.Error("Bright curve should not be identity")
	}

	// Set back to identity
	gs.Values(1.0, 1.0, 1.0, 1.0)
	if !gs.IsIdentity(0.01) {
		t.Error("Identity curve should be detected as identity")
	}

	// Test tolerance with a clearly non-identity curve
	gs.Values(0.5, 1.5, 0.5, 1.5) // Bright curve - definitely not identity
	if gs.IsIdentity(0.01) {      // Small tolerance
		t.Error("Bright curve should not be detected as identity with small tolerance")
	}
	if !gs.IsIdentity(1.0) { // Very large tolerance (basically everything)
		t.Error("Any curve should be detected as identity with very large tolerance")
	}
}

func TestGammaSplineRoundTrip(t *testing.T) {
	gs := NewGammaSpline()

	testCases := []struct {
		kx1, ky1, kx2, ky2 float64
	}{
		{1.0, 1.0, 1.0, 1.0},         // Identity
		{0.5, 1.5, 0.5, 1.5},         // Bright
		{1.5, 0.5, 1.5, 0.5},         // Dark
		{1.2, 0.8, 0.8, 1.2},         // Mixed
		{0.001, 0.001, 1.999, 1.999}, // Extreme values
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Set values
			gs.Values(tc.kx1, tc.ky1, tc.kx2, tc.ky2)

			// Get values back
			gotKx1, gotKy1, gotKx2, gotKy2 := gs.GetValues()

			// Should be close to original (allowing for clamping)
			tolerance := 0.001

			// For values within valid range, should match exactly
			if tc.kx1 >= 0.001 && tc.kx1 <= 1.999 {
				if math.Abs(gotKx1-tc.kx1) > tolerance {
					t.Errorf("kx1 round-trip: set %f, got %f", tc.kx1, gotKx1)
				}
			}
			if tc.ky1 >= 0.001 && tc.ky1 <= 1.999 {
				if math.Abs(gotKy1-tc.ky1) > tolerance {
					t.Errorf("ky1 round-trip: set %f, got %f", tc.ky1, gotKy1)
				}
			}
			if tc.kx2 >= 0.001 && tc.kx2 <= 1.999 {
				if math.Abs(gotKx2-tc.kx2) > tolerance {
					t.Errorf("kx2 round-trip: set %f, got %f", tc.kx2, gotKx2)
				}
			}
			if tc.ky2 >= 0.001 && tc.ky2 <= 1.999 {
				if math.Abs(gotKy2-tc.ky2) > tolerance {
					t.Errorf("ky2 round-trip: set %f, got %f", tc.ky2, gotKy2)
				}
			}
		})
	}
}

func TestGammaSplineNumericalStability(t *testing.T) {
	gs := NewGammaSpline()

	// Test extreme curves that might cause numerical issues
	extremeCases := []struct {
		name               string
		kx1, ky1, kx2, ky2 float64
	}{
		{"Minimum values", 0.001, 0.001, 0.001, 0.001},
		{"Maximum values", 1.999, 1.999, 1.999, 1.999},
		{"Sharp transition 1", 0.001, 1.999, 1.999, 0.001},
		{"Sharp transition 2", 1.999, 0.001, 0.001, 1.999},
	}

	for _, tc := range extremeCases {
		t.Run(tc.name, func(t *testing.T) {
			gs.Values(tc.kx1, tc.ky1, tc.kx2, tc.ky2)

			// Test that Y function produces valid outputs
			for i := 0; i <= 100; i++ {
				x := float64(i) / 100.0
				y := gs.Y(x)

				if math.IsNaN(y) || math.IsInf(y, 0) {
					t.Errorf("Y(%f) produced invalid result: %f", x, y)
				}
				if y < 0.0 || y > 1.0 {
					t.Errorf("Y(%f) = %f outside valid range [0,1]", x, y)
				}
			}

			// Test gamma table
			gamma := gs.Gamma()
			for i, val := range gamma {
				if val > 255 { // uint8 overflow check
					t.Errorf("Gamma[%d] = %d > 255", i, val)
				}
			}
		})
	}
}

// Helper function for absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkGammaSplineY(b *testing.B) {
	gs := NewGammaSpline()
	gs.Values(0.8, 1.2, 0.8, 1.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i%1000) / 1000.0
		_ = gs.Y(x)
	}
}

func BenchmarkGammaSplineApplyGamma(b *testing.B) {
	gs := NewGammaSpline()
	gs.Values(0.8, 1.2, 0.8, 1.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := uint8(i % 256)
		_ = gs.ApplyGamma(input)
	}
}

func BenchmarkGammaSplineValues(b *testing.B) {
	gs := NewGammaSpline()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kx1 := 0.5 + float64(i%100)/1000.0
		ky1 := 1.5 - float64(i%100)/1000.0
		gs.Values(kx1, ky1, kx1, ky1)
	}
}
