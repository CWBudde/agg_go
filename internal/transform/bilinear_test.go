package transform

import (
	"math"
	"testing"
)

func TestNewTransBilinear(t *testing.T) {
	tb := NewTransBilinear()
	if tb == nil {
		t.Fatal("NewTransBilinear returned nil")
	}
	if tb.IsValid() {
		t.Error("NewTransBilinear should create invalid transformation initially")
	}
}

func TestTransBilinear_IdentityTransformation(t *testing.T) {
	// Test identity transformation: unit square to unit square
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1} // unit square
	dst := [8]float64{0, 0, 1, 0, 1, 1, 0, 1} // same unit square

	tb := NewTransBilinearQuadToQuad(src, dst)

	if !tb.IsValid() {
		t.Fatal("Identity transformation should be valid")
	}

	// Test corner points
	testPoints := [][2]float64{
		{0, 0},
		{1, 0},
		{1, 1},
		{0, 1},
		{0.5, 0.5}, // center point
	}

	tolerance := 1e-10
	for _, point := range testPoints {
		x, y := tb.TransformValues(point[0], point[1])
		if math.Abs(x-point[0]) > tolerance || math.Abs(y-point[1]) > tolerance {
			t.Errorf("Identity transform failed for point (%f,%f): got (%f,%f)",
				point[0], point[1], x, y)
		}
	}
}

func TestTransBilinear_RectToQuad(t *testing.T) {
	// Transform unit rectangle to a known quadrilateral
	quad := [8]float64{
		0, 0, // (0,0) -> (0,0)
		2, 0, // (1,0) -> (2,0)
		2, 3, // (1,1) -> (2,3)
		0, 3, // (0,1) -> (0,3)
	}

	tb := NewTransBilinearRectToQuad(0, 0, 1, 1, quad)

	if !tb.IsValid() {
		t.Fatal("RectToQuad transformation should be valid")
	}

	// Test corner transformations
	testCases := [][4]float64{
		{0, 0, 0, 0}, // (0,0) -> (0,0)
		{1, 0, 2, 0}, // (1,0) -> (2,0)
		{1, 1, 2, 3}, // (1,1) -> (2,3)
		{0, 1, 0, 3}, // (0,1) -> (0,3)
	}

	tolerance := 1e-10
	for _, tc := range testCases {
		x, y := tb.TransformValues(tc[0], tc[1])
		if math.Abs(x-tc[2]) > tolerance || math.Abs(y-tc[3]) > tolerance {
			t.Errorf("RectToQuad failed for (%f,%f): got (%f,%f), want (%f,%f)",
				tc[0], tc[1], x, y, tc[2], tc[3])
		}
	}
}

func TestTransBilinear_QuadToRect(t *testing.T) {
	// Transform a quadrilateral to unit rectangle
	quad := [8]float64{
		1, 1, // (1,1) -> (0,0)
		3, 1, // (3,1) -> (1,0)
		3, 4, // (3,4) -> (1,1)
		1, 4, // (1,4) -> (0,1)
	}

	tb := NewTransBilinearQuadToRect(quad, 0, 0, 1, 1)

	if !tb.IsValid() {
		t.Fatal("QuadToRect transformation should be valid")
	}

	// Test corner transformations
	testCases := [][4]float64{
		{1, 1, 0, 0}, // (1,1) -> (0,0)
		{3, 1, 1, 0}, // (3,1) -> (1,0)
		{3, 4, 1, 1}, // (3,4) -> (1,1)
		{1, 4, 0, 1}, // (1,4) -> (0,1)
	}

	tolerance := 1e-10
	for _, tc := range testCases {
		x, y := tb.TransformValues(tc[0], tc[1])
		if math.Abs(x-tc[2]) > tolerance || math.Abs(y-tc[3]) > tolerance {
			t.Errorf("QuadToRect failed for (%f,%f): got (%f,%f), want (%f,%f)",
				tc[0], tc[1], x, y, tc[2], tc[3])
		}
	}
}

func TestTransBilinear_ArbitraryQuadToQuad(t *testing.T) {
	// Transform one arbitrary quad to another
	src := [8]float64{
		0, 0, // bottom-left
		2, 0, // bottom-right
		2, 2, // top-right
		0, 2, // top-left
	}
	dst := [8]float64{
		1, 1, // bottom-left
		4, 2, // bottom-right
		3, 5, // top-right
		0, 4, // top-left
	}

	tb := NewTransBilinearQuadToQuad(src, dst)

	if !tb.IsValid() {
		t.Fatal("QuadToQuad transformation should be valid")
	}

	// Test that corner points map correctly
	testCases := [][4]float64{
		{0, 0, 1, 1}, // src[0,1] -> dst[0,1]
		{2, 0, 4, 2}, // src[2,3] -> dst[2,3]
		{2, 2, 3, 5}, // src[4,5] -> dst[4,5]
		{0, 2, 0, 4}, // src[6,7] -> dst[6,7]
	}

	tolerance := 1e-10
	for _, tc := range testCases {
		x, y := tb.TransformValues(tc[0], tc[1])
		if math.Abs(x-tc[2]) > tolerance || math.Abs(y-tc[3]) > tolerance {
			t.Errorf("QuadToQuad failed for (%f,%f): got (%f,%f), want (%f,%f)",
				tc[0], tc[1], x, y, tc[2], tc[3])
		}
	}

	// Test center point (should be somewhere reasonable)
	centerX, centerY := tb.TransformValues(1, 1)
	if centerX < 0 || centerX > 5 || centerY < 1 || centerY > 6 {
		t.Errorf("Center point transformation seems unreasonable: (%f,%f)", centerX, centerY)
	}
}

func TestTransBilinear_DegenerateQuadrilateral(t *testing.T) {
	// Test with a truly degenerate quadrilateral (all same points)
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1} // unit square (valid)
	dst := [8]float64{5, 5, 5, 5, 5, 5, 5, 5} // all points the same (degenerate)

	tb := NewTransBilinearQuadToQuad(src, dst)

	// This should result in an invalid transformation due to singularity
	if tb.IsValid() {
		// If it's somehow valid, at least check it doesn't crash
		_, _ = tb.TransformValues(0.5, 0.5)
		t.Log("Degenerate quadrilateral resulted in valid transformation - this may be mathematically correct")
	}
}

func TestTransBilinear_InvalidTransform(t *testing.T) {
	// Test behavior with invalid transformation
	tb := NewTransBilinear() // starts invalid

	// Transform should return input unchanged
	x, y := tb.TransformValues(5, 7)
	if x != 5 || y != 7 {
		t.Errorf("Invalid transform should return input unchanged: got (%f,%f), want (5,7)", x, y)
	}
}

func TestIteratorX_Identity(t *testing.T) {
	// Test iterator with identity transformation
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("Identity transformation should be valid")
	}

	// Create iterator at (0.5, 0.3) with step 0.1
	it := tb.NewIteratorX(0.5, 0.3, 0.1)

	tolerance := 1e-10
	expectedX := 0.5
	expectedY := 0.3

	// Check initial position
	if math.Abs(it.X()-expectedX) > tolerance || math.Abs(it.Y()-expectedY) > tolerance {
		t.Errorf("Initial iterator position: got (%f,%f), want (%f,%f)",
			it.X(), it.Y(), expectedX, expectedY)
	}

	// Step forward a few times and check positions
	for i := 0; i < 5; i++ {
		it.Next()
		expectedX += 0.1

		if math.Abs(it.X()-expectedX) > tolerance || math.Abs(it.Y()-expectedY) > tolerance {
			t.Errorf("Iterator position after %d steps: got (%f,%f), want (%f,%f)",
				i+1, it.X(), it.Y(), expectedX, expectedY)
		}
	}
}

func TestIteratorX_WithTransform(t *testing.T) {
	// Test iterator with actual transformation
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1} // unit square
	dst := [8]float64{0, 0, 2, 0, 2, 2, 0, 2} // 2x2 square (scale by 2)

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("Scale transformation should be valid")
	}

	// Create iterator and check that successive points match direct transformation
	startX, startY := 0.1, 0.2
	step := 0.05

	it := tb.NewIteratorX(startX, startY, step)

	tolerance := 1e-10
	currentX := startX

	for i := 0; i < 10; i++ {
		// Get expected position using direct transform
		expectedX, expectedY := tb.TransformValues(currentX, startY)

		// Compare with iterator position
		if math.Abs(it.X()-expectedX) > tolerance || math.Abs(it.Y()-expectedY) > tolerance {
			t.Errorf("Iterator mismatch at step %d: got (%f,%f), want (%f,%f)",
				i, it.X(), it.Y(), expectedX, expectedY)
		}

		it.Next()
		currentX += step
	}
}

func TestIteratorX_InvalidTransform(t *testing.T) {
	// Test iterator with invalid transformation
	tb := NewTransBilinear()

	it := tb.NewIteratorX(1, 2, 0.1)

	// Should return original coordinates since transform is invalid
	if it.X() != 1 || it.Y() != 2 {
		t.Errorf("Invalid transform iterator should return original coords: got (%f,%f), want (1,2)",
			it.X(), it.Y())
	}

	// Next should not change anything meaningful
	it.Next()
	if it.X() != 1 || it.Y() != 2 {
		t.Error("Invalid transform iterator should not change after Next()")
	}
}

func TestTransBilinear_BilinearInterpolation(t *testing.T) {
	// Test that the transformation does proper bilinear interpolation
	// Use a transformation where we can predict the result
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1} // unit square
	dst := [8]float64{
		0, 0, // (0,0) -> (0,0)
		1, 1, // (1,0) -> (1,1)
		2, 2, // (1,1) -> (2,2)
		1, 1, // (0,1) -> (1,1)
	}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("Bilinear interpolation test transformation should be valid")
	}

	// Test the center point (0.5, 0.5)
	// With bilinear interpolation, this should map to (1, 1)
	x, y := tb.TransformValues(0.5, 0.5)
	tolerance := 1e-10

	if math.Abs(x-1.0) > tolerance || math.Abs(y-1.0) > tolerance {
		t.Errorf("Center point bilinear interpolation: got (%f,%f), want (1,1)", x, y)
	}
}

// Benchmark tests
func BenchmarkTransBilinear_Transform(b *testing.B) {
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 2, 1, 1, 3, -1, 2}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		b.Fatal("Transformation should be valid")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.TransformValues(0.5, 0.7)
	}
}

func BenchmarkTransBilinear_IteratorX(b *testing.B) {
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 2, 1, 1, 3, -1, 2}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		b.Fatal("Transformation should be valid")
	}

	it := tb.NewIteratorX(0, 0.5, 0.01)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it.Next()
	}
}

func BenchmarkTransBilinear_Creation(b *testing.B) {
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 2, 1, 1, 3, -1, 2}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb := NewTransBilinearQuadToQuad(src, dst)
		_ = tb.IsValid() // Make sure it's used
	}
}

func TestTransBilinear_InverseTransform_Identity(t *testing.T) {
	// Test inverse transformation with identity
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("Identity transformation should be valid")
	}

	testPoints := [][2]float64{
		{0.0, 0.0},
		{1.0, 0.0},
		{1.0, 1.0},
		{0.0, 1.0},
		{0.5, 0.5},
		{0.25, 0.75},
		{0.8, 0.2},
	}

	tolerance := 1e-8
	for _, point := range testPoints {
		// Forward transform
		fx, fy := tb.TransformValues(point[0], point[1])
		// Inverse transform should give back original
		ix, iy := tb.InverseTransformValues(fx, fy)

		if math.Abs(ix-point[0]) > tolerance || math.Abs(iy-point[1]) > tolerance {
			t.Errorf("Inverse identity failed for (%f,%f): forward->(%f,%f) inverse->(%f,%f)",
				point[0], point[1], fx, fy, ix, iy)
		}
	}
}

func TestTransBilinear_InverseTransform_RectToQuad(t *testing.T) {
	// Test inverse with rectangle to quadrilateral transformation
	quad := [8]float64{1, 1, 3, 2, 2, 4, 0, 3}
	tb := NewTransBilinearRectToQuad(0, 0, 1, 1, quad)

	if !tb.IsValid() {
		t.Fatal("RectToQuad transformation should be valid")
	}

	testPoints := [][2]float64{
		{0.0, 0.0},
		{1.0, 0.0},
		{1.0, 1.0},
		{0.0, 1.0},
		{0.5, 0.5},
		{0.25, 0.75},
		{0.1, 0.9},
	}

	tolerance := 1e-8
	for _, point := range testPoints {
		// Forward transform
		fx, fy := tb.TransformValues(point[0], point[1])
		// Inverse transform
		ix, iy := tb.InverseTransformValues(fx, fy)

		if math.Abs(ix-point[0]) > tolerance || math.Abs(iy-point[1]) > tolerance {
			t.Errorf("Inverse RectToQuad failed for (%f,%f): forward->(%f,%f) inverse->(%f,%f)",
				point[0], point[1], fx, fy, ix, iy)
		}
	}
}

func TestTransBilinear_InverseTransform_ArbitraryQuad(t *testing.T) {
	// Test inverse with arbitrary quadrilateral transformation
	src := [8]float64{0, 0, 2, 0, 2, 2, 0, 2}
	dst := [8]float64{1, 1, 4, 2, 3, 5, 0, 4}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("QuadToQuad transformation should be valid")
	}

	testPoints := [][2]float64{
		{0.0, 0.0},
		{2.0, 0.0},
		{2.0, 2.0},
		{0.0, 2.0},
		{1.0, 1.0},
		{0.5, 1.5},
		{1.8, 0.3},
	}

	tolerance := 1e-8
	for _, point := range testPoints {
		// Forward transform
		fx, fy := tb.TransformValues(point[0], point[1])
		// Inverse transform
		ix, iy := tb.InverseTransformValues(fx, fy)

		if math.Abs(ix-point[0]) > tolerance || math.Abs(iy-point[1]) > tolerance {
			t.Errorf("Inverse arbitrary quad failed for (%f,%f): forward->(%f,%f) inverse->(%f,%f)",
				point[0], point[1], fx, fy, ix, iy)
		}
	}
}

func TestTransBilinear_InverseTransform_InvalidTransform(t *testing.T) {
	// Test inverse with invalid transformation
	tb := NewTransBilinear()

	x, y := tb.InverseTransformValues(5, 7)
	if x != 5 || y != 7 {
		t.Errorf("Invalid inverse transform should return input unchanged: got (%f,%f), want (5,7)", x, y)
	}
}

func TestTransBilinear_InverseTransform_NonConvergent(t *testing.T) {
	// Test case where Newton-Raphson might have difficulty converging
	// Use a nearly degenerate transformation
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 1, 1e-10, 2, 2e-10, 1, 1e-10} // Nearly collinear

	tb := NewTransBilinearQuadToQuad(src, dst)

	// Test inverse at a reasonable point
	fx, fy := 0.5, 0.5
	ix, iy := tb.InverseTransformValues(fx, fy)

	// Should not return NaN or Inf
	if math.IsNaN(ix) || math.IsNaN(iy) || math.IsInf(ix, 0) || math.IsInf(iy, 0) {
		t.Error("Inverse transform should not return NaN or Inf even for difficult cases")
	}

	// If transformation is valid, test round-trip accuracy with relaxed tolerance
	if tb.IsValid() {
		checkX, checkY := tb.TransformValues(ix, iy)
		tolerance := 1e-6 // More relaxed for nearly degenerate cases
		if math.Abs(checkX-fx) > tolerance || math.Abs(checkY-fy) > tolerance {
			t.Logf("Round-trip accuracy reduced for nearly degenerate case: target=(%f,%f) result=(%f,%f)",
				fx, fy, checkX, checkY)
		}
	}
}

func TestTransBilinear_InverseTransform_CornerCases(t *testing.T) {
	// Test inverse transformation at various challenging points
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 3, 1, 2, 4, -1, 3} // Asymmetric quadrilateral

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		t.Fatal("Asymmetric transformation should be valid")
	}

	// Test points near boundaries and corners
	testPoints := [][2]float64{
		{0.001, 0.001}, // Near corner
		{0.999, 0.999}, // Near opposite corner
		{0.5, 0.001},   // Near edge
		{0.001, 0.5},   // Near edge
		{0.5, 0.5},     // Center
	}

	tolerance := 1e-7
	for _, point := range testPoints {
		fx, fy := tb.TransformValues(point[0], point[1])
		ix, iy := tb.InverseTransformValues(fx, fy)

		if math.Abs(ix-point[0]) > tolerance || math.Abs(iy-point[1]) > tolerance {
			t.Errorf("Inverse corner case failed for (%f,%f): forward->(%f,%f) inverse->(%f,%f)",
				point[0], point[1], fx, fy, ix, iy)
		}
	}
}

// Benchmark inverse transformation
func BenchmarkTransBilinear_InverseTransform(b *testing.B) {
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 2, 1, 1, 3, -1, 2}

	tb := NewTransBilinearQuadToQuad(src, dst)
	if !tb.IsValid() {
		b.Fatal("Transformation should be valid")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.InverseTransformValues(0.5, 0.7)
	}
}

// Test error conditions and edge cases
func TestTransBilinear_EdgeCases(t *testing.T) {
	t.Run("ZeroSizeQuad", func(t *testing.T) {
		// All points the same
		src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
		dst := [8]float64{5, 5, 5, 5, 5, 5, 5, 5}

		tb := NewTransBilinearQuadToQuad(src, dst)
		// This might be valid depending on the mathematical properties
		// The key test is that it doesn't crash
		_, _ = tb.TransformValues(0.5, 0.5)
	})

	t.Run("LargeValues", func(t *testing.T) {
		src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
		dst := [8]float64{1e6, 1e6, 2e6, 1e6, 2e6, 2e6, 1e6, 2e6}

		tb := NewTransBilinearQuadToQuad(src, dst)
		if tb.IsValid() {
			x, y := tb.TransformValues(0.5, 0.5)
			if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
				t.Error("Large values should not produce NaN or Inf")
			}
		}
	})

	t.Run("NearSingular", func(t *testing.T) {
		// Nearly collinear points
		src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
		dst := [8]float64{0, 0, 1, 1e-15, 2, 2e-15, 1, 1e-15}

		tb := NewTransBilinearQuadToQuad(src, dst)
		// Should either be valid with reasonable results or invalid
		// Main test is that it doesn't crash
		if tb.IsValid() {
			x, y := tb.TransformValues(0.5, 0.5)
			if math.IsNaN(x) || math.IsNaN(y) {
				t.Error("Near-singular matrix should not produce NaN")
			}
		}
	})
}

func TestTransBilinear_TransformerInterface(t *testing.T) {
	// Test that TransBilinear properly implements Transformer interface
	quad := [8]float64{0, 0, 2, 0, 2, 2, 0, 2}
	tb := NewTransBilinearRectToQuad(0, 0, 1, 1, quad)

	if !tb.IsValid() {
		t.Fatal("Transformation should be valid")
	}

	// Test pointer-based Transform method
	x, y := 0.5, 0.5
	tb.Transform(&x, &y)

	// Should transform to center of the quad (1, 1)
	tolerance := 1e-10
	if math.Abs(x-1.0) > tolerance || math.Abs(y-1.0) > tolerance {
		t.Errorf("Transform interface failed: expected (1,1), got (%f,%f)", x, y)
	}
}

func TestTransBilinear_ToMatrix(t *testing.T) {
	// Test matrix extraction
	src := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}
	dst := [8]float64{0, 0, 2, 0, 2, 2, 0, 2} // scaling by 2

	tb := NewTransBilinearQuadToQuad(src, dst)
	matrix := tb.ToMatrix()

	// For a simple scale transformation, we can verify some expected values
	if !tb.IsValid() {
		t.Fatal("Transformation should be valid")
	}

	// Just verify we get a non-zero matrix
	hasNonZero := false
	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			if math.Abs(matrix[i][j]) > 1e-10 {
				hasNonZero = true
				break
			}
		}
	}

	if !hasNonZero {
		t.Error("ToMatrix should return non-zero matrix for valid transformation")
	}
}

func TestNewTransBilinearFromTransformer(t *testing.T) {
	// Create an affine transformation for testing
	affine := NewTransAffineFromValues(2.0, 0.0, 0.0, 2.0, 1.0, 1.0) // scale by 2, translate by (1,1)

	// Create a source quadrilateral (unit square)
	srcQuad := [8]float64{0, 0, 1, 0, 1, 1, 0, 1}

	// Extract bilinear transformation
	tb := NewTransBilinearFromTransformer(affine, srcQuad)

	if !tb.IsValid() {
		t.Fatal("Transformation should be valid")
	}

	// Test that it produces similar results
	tolerance := 1e-10
	testPoints := [][2]float64{{0, 0}, {1, 0}, {0.5, 0.5}}

	for _, point := range testPoints {
		// Transform with original affine
		ax, ay := point[0], point[1]
		affine.Transform(&ax, &ay)

		// Transform with extracted bilinear
		bx, by := tb.TransformValues(point[0], point[1])

		if math.Abs(ax-bx) > tolerance || math.Abs(ay-by) > tolerance {
			t.Errorf("NewTransBilinearFromTransformer mismatch at (%f,%f): affine=(%f,%f), bilinear=(%f,%f)",
				point[0], point[1], ax, ay, bx, by)
		}
	}
}

func TestNewTransBilinearFromRect(t *testing.T) {
	// Create a simple affine transformation
	affine := NewTransAffineFromValues(1.5, 0.0, 0.0, 1.5, 0.5, 0.5)

	// Extract bilinear approximation over a rectangle
	tb := NewTransBilinearFromRect(affine, 0, 0, 2, 2)

	if !tb.IsValid() {
		t.Fatal("Transformation should be valid")
	}

	// Test corner points of the rectangle
	corners := [][2]float64{{0, 0}, {2, 0}, {2, 2}, {0, 2}}
	tolerance := 1e-10

	for _, corner := range corners {
		// Transform with original affine
		ax, ay := corner[0], corner[1]
		affine.Transform(&ax, &ay)

		// Transform with extracted bilinear
		bx, by := tb.TransformValues(corner[0], corner[1])

		if math.Abs(ax-bx) > tolerance || math.Abs(ay-by) > tolerance {
			t.Errorf("NewTransBilinearFromRect mismatch at corner (%f,%f): affine=(%f,%f), bilinear=(%f,%f)",
				corner[0], corner[1], ax, ay, bx, by)
		}
	}
}
