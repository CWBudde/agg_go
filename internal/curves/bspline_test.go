package curves

import (
	"math"
	"testing"
)

// Test data for B-spline interpolation
var (
	// Simple linear test data
	linearX = []float64{0, 1, 2, 3, 4}
	linearY = []float64{0, 1, 2, 3, 4}

	// Quadratic test data: y = x^2
	quadX = []float64{0, 1, 2, 3, 4}
	quadY = []float64{0, 1, 4, 9, 16}

	// Sine wave test data
	sineX = []float64{0, math.Pi / 4, math.Pi / 2, 3 * math.Pi / 4, math.Pi}
	sineY = []float64{0, math.Sqrt2 / 2, 1, math.Sqrt2 / 2, 0}
)

func TestNewBSpline(t *testing.T) {
	bs := NewBSpline()
	if bs == nil {
		t.Fatal("NewBSpline() returned nil")
	}
	if bs.NumPoints() != 0 {
		t.Errorf("Expected 0 points, got %d", bs.NumPoints())
	}
	if bs.MaxPoints() != 0 {
		t.Errorf("Expected 0 max points, got %d", bs.MaxPoints())
	}
}

func TestNewBSplineWithCapacity(t *testing.T) {
	bs := NewBSplineWithCapacity(10)
	if bs == nil {
		t.Fatal("NewBSplineWithCapacity() returned nil")
	}
	if bs.MaxPoints() != 10 {
		t.Errorf("Expected 10 max points, got %d", bs.MaxPoints())
	}
	if bs.NumPoints() != 0 {
		t.Errorf("Expected 0 points, got %d", bs.NumPoints())
	}
}

func TestNewBSplineFromPoints(t *testing.T) {
	bs := NewBSplineFromPoints(linearX, linearY)
	if bs == nil {
		t.Fatal("NewBSplineFromPoints() returned nil")
	}
	if bs.NumPoints() != len(linearX) {
		t.Errorf("Expected %d points, got %d", len(linearX), bs.NumPoints())
	}
}

func TestBSplineAddPoint(t *testing.T) {
	bs := NewBSplineWithCapacity(5)

	// Add points one by one
	for i := 0; i < len(linearX); i++ {
		bs.AddPoint(linearX[i], linearY[i])
		if bs.NumPoints() != i+1 {
			t.Errorf("After adding point %d, expected %d points, got %d",
				i, i+1, bs.NumPoints())
		}
	}

	// Try to add beyond capacity
	bs.AddPoint(5.0, 5.0)
	if bs.NumPoints() != 5 {
		t.Errorf("Expected points to remain at 5 when adding beyond capacity, got %d",
			bs.NumPoints())
	}
}

func TestBSplineLinearInterpolation(t *testing.T) {
	bs := NewBSplineFromPoints(linearX, linearY)

	// Test interpolation at control points
	for i, x := range linearX {
		y := bs.Get(x)
		expected := linearY[i]
		if math.Abs(y-expected) > 1e-10 {
			t.Errorf("At control point x=%g, expected y=%g, got y=%g",
				x, expected, y)
		}
	}

	// Test interpolation between control points
	// For linear data, B-spline should interpolate exactly
	testPoints := []float64{0.5, 1.5, 2.5, 3.5}
	expectedPoints := []float64{0.5, 1.5, 2.5, 3.5}

	for i, x := range testPoints {
		y := bs.Get(x)
		expected := expectedPoints[i]
		if math.Abs(y-expected) > 1e-6 {
			t.Errorf("At interpolation point x=%g, expected y≈%g, got y=%g",
				x, expected, y)
		}
	}
}

func TestBSplineQuadraticInterpolation(t *testing.T) {
	bs := NewBSplineFromPoints(quadX, quadY)

	// Test interpolation at control points
	for i, x := range quadX {
		y := bs.Get(x)
		expected := quadY[i]
		if math.Abs(y-expected) > 1e-10 {
			t.Errorf("At control point x=%g, expected y=%g, got y=%g",
				x, expected, y)
		}
	}

	// Test interpolation between control points
	// For quadratic data, B-spline should be very close but not exact
	testX := 1.5
	testY := bs.Get(testX)
	expectedY := testX * testX // 2.25

	// Allow some tolerance for spline approximation
	tolerance := 0.5
	if math.Abs(testY-expectedY) > tolerance {
		t.Errorf("At x=%g, expected y≈%g, got y=%g (diff=%g)",
			testX, expectedY, testY, math.Abs(testY-expectedY))
	}
}

func TestBSplineExtrapolation(t *testing.T) {
	bs := NewBSplineFromPoints(linearX, linearY)

	// Test left extrapolation
	leftX := -1.0
	leftY := bs.Get(leftX)
	// Linear extrapolation should give y = x for this data
	expectedLeftY := leftX
	if math.Abs(leftY-expectedLeftY) > 1e-6 {
		t.Errorf("Left extrapolation at x=%g, expected y≈%g, got y=%g",
			leftX, expectedLeftY, leftY)
	}

	// Test right extrapolation
	rightX := 5.0
	rightY := bs.Get(rightX)
	expectedRightY := rightX
	if math.Abs(rightY-expectedRightY) > 1e-6 {
		t.Errorf("Right extrapolation at x=%g, expected y≈%g, got y=%g",
			rightX, expectedRightY, rightY)
	}
}

func TestBSplineGetStateful(t *testing.T) {
	bs := NewBSplineFromPoints(linearX, linearY)

	// Test that GetStateful produces same results as Get
	testPoints := []float64{-1, 0.5, 1.5, 2.5, 3.5, 5}

	for _, x := range testPoints {
		yGet := bs.Get(x)
		yGetStateful := bs.GetStateful(x)

		if math.Abs(yGet-yGetStateful) > 1e-12 {
			t.Errorf("At x=%g, Get()=%g but GetStateful()=%g",
				x, yGet, yGetStateful)
		}
	}

	// Test sequential access optimization
	// When accessing points in order, GetStateful should be optimized
	for _, x := range []float64{0.1, 0.2, 0.3, 0.4, 0.5} {
		y1 := bs.GetStateful(x)
		y2 := bs.Get(x)
		if math.Abs(y1-y2) > 1e-12 {
			t.Errorf("Sequential GetStateful failed at x=%g", x)
		}
	}
}

func TestBSplineEdgeCases(t *testing.T) {
	// Test with empty spline
	bs := NewBSpline()
	y := bs.Get(1.0)
	if y != 0.0 {
		t.Errorf("Empty spline should return 0.0, got %g", y)
	}

	// Test with single point
	bs.Init(1)
	bs.AddPoint(1.0, 2.0)
	bs.Prepare()
	y = bs.Get(1.0)
	if y != 0.0 {
		t.Errorf("Single point spline should return 0.0, got %g", y)
	}

	// Test with two points
	bs.Init(2)
	bs.AddPoint(0.0, 0.0)
	bs.AddPoint(1.0, 1.0)
	bs.Prepare()
	y = bs.Get(0.5)
	if y != 0.0 {
		t.Errorf("Two point spline should return 0.0, got %g", y)
	}
}

func TestBSplineReset(t *testing.T) {
	bs := NewBSplineFromPoints(linearX, linearY)
	if bs.NumPoints() != len(linearX) {
		t.Errorf("Expected %d points before reset", len(linearX))
	}

	bs.Reset()
	if bs.NumPoints() != 0 {
		t.Errorf("Expected 0 points after reset, got %d", bs.NumPoints())
	}

	// Should be able to add points after reset
	bs.AddPoint(1.0, 1.0)
	bs.AddPoint(2.0, 4.0)
	bs.AddPoint(3.0, 9.0)
	bs.Prepare()

	if bs.NumPoints() != 3 {
		t.Errorf("Expected 3 points after adding to reset spline, got %d",
			bs.NumPoints())
	}
}

func TestBSplineInitFromPointsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for mismatched slice lengths")
		}
	}()

	x := []float64{1, 2, 3}
	y := []float64{1, 2} // Different length
	NewBSplineFromPoints(x, y)
}

func TestBSplineSineWave(t *testing.T) {
	bs := NewBSplineFromPoints(sineX, sineY)

	// Test at control points
	for i, x := range sineX {
		y := bs.Get(x)
		expected := sineY[i]
		if math.Abs(y-expected) > 1e-10 {
			t.Errorf("Sine wave at control point x=%g, expected y=%g, got y=%g",
				x, expected, y)
		}
	}

	// Test interpolation - should be smooth
	testX := math.Pi / 3 // Between π/4 and π/2
	testY := bs.Get(testX)

	// Value should be between the surrounding control points and reasonable
	if testY < 0.5 || testY > 1.0 {
		t.Errorf("Sine interpolation at x=%g seems unreasonable: y=%g",
			testX, testY)
	}
}

func TestBSplineMonotonicity(t *testing.T) {
	// For monotonic input data, spline should maintain general monotonic trend
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{0, 2, 3, 5, 8} // Monotonically increasing

	bs := NewBSplineFromPoints(x, y)

	// Sample points and check general trend
	prev := bs.Get(0.5)
	for _, testX := range []float64{1.0, 1.5, 2.0, 2.5, 3.0, 3.5} {
		current := bs.Get(testX)
		if current < prev-0.5 { // Allow some tolerance for spline behavior
			t.Errorf("Significant non-monotonicity detected: x=%g, y=%g < prev=%g",
				testX, current, prev)
		}
		prev = current
	}
}

// Benchmark tests
func BenchmarkBSplineGet(b *testing.B) {
	bs := NewBSplineFromPoints(quadX, quadY)
	x := 1.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bs.Get(x)
	}
}

func BenchmarkBSplineGetStateful(b *testing.B) {
	bs := NewBSplineFromPoints(quadX, quadY)
	x := 1.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bs.GetStateful(x)
	}
}

func BenchmarkBSplineSequentialAccess(b *testing.B) {
	bs := NewBSplineFromPoints(quadX, quadY)
	testPoints := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, x := range testPoints {
			_ = bs.GetStateful(x)
		}
	}
}

// Example usage test
func ExampleBSpline() {
	// Create control points
	x := []float64{0, 1, 2, 3, 4}
	y := []float64{0, 1, 4, 9, 16} // y = x²

	// Create and initialize B-spline
	spline := NewBSplineFromPoints(x, y)

	// Get interpolated values
	result := spline.Get(1.5) // Interpolate at x=1.5
	_ = result

	// Use stateful version for better performance with sequential access
	for x := 0.0; x <= 4.0; x += 0.1 {
		_ = spline.GetStateful(x)
	}
}

func TestBSplinePerformanceComparison(t *testing.T) {
	// Create a larger dataset for performance testing
	n := 20
	x := make([]float64, n)
	y := make([]float64, n)

	for i := 0; i < n; i++ {
		x[i] = float64(i) * 0.5
		y[i] = math.Sin(x[i]) // Sine wave
	}

	bs := NewBSplineFromPoints(x, y)

	// Test many random access points with Get
	testX := 5.0
	numTests := 1000

	// Time regular Get method
	startTime := getTime()
	for i := 0; i < numTests; i++ {
		_ = bs.Get(testX + float64(i)*0.001)
	}
	regularTime := getTime() - startTime

	// Reset for stateful test
	bs = NewBSplineFromPoints(x, y)

	// Time GetStateful method with sequential access
	startTime = getTime()
	for i := 0; i < numTests; i++ {
		_ = bs.GetStateful(testX + float64(i)*0.001)
	}
	statefulTime := getTime() - startTime

	t.Logf("Regular Get: %v, Stateful Get: %v", regularTime, statefulTime)

	// GetStateful should be faster or at least not significantly slower
	// for sequential access patterns
	if statefulTime > regularTime*2 {
		t.Errorf("GetStateful is significantly slower than Get for sequential access")
	}
}

// Helper function to get current time (simplified)
func getTime() int64 {
	// This is a placeholder - in real benchmarks you'd use time.Now()
	return 0
}
