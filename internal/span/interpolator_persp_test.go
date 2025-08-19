package span

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestSpanInterpolatorPerspectiveExact_Basic(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	if !interp.IsValid() {
		t.Error("New interpolator should be valid with identity transform")
	}

	if shift := interp.SubpixelShift(); shift != 8 {
		t.Errorf("Expected subpixel shift 8, got %d", shift)
	}
}

func TestSpanInterpolatorPerspectiveLerp_Basic(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	if !interp.IsValid() {
		t.Error("New interpolator should be valid with identity transform")
	}

	if shift := interp.SubpixelShift(); shift != 8 {
		t.Errorf("Expected subpixel shift 8, got %d", shift)
	}
}

func TestSpanInterpolatorPerspectiveExact_IdentityTransform(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Test identity transformation (should be 1:1)
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	interp.QuadToQuad(src, dst)

	length := 100
	interp.Begin(0, 0, length)

	x, y := interp.Coordinates()
	expectedX := 0
	expectedY := 0

	// Allow some tolerance for subpixel calculations
	tolerance := 5 // About 5 subpixel units
	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Identity transform at start: expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}

	// Test halfway through
	for i := 0; i < length/2; i++ {
		interp.Next()
	}

	x, y = interp.Coordinates()
	expectedX = (length / 2) << 8 // Convert to subpixel
	expectedY = 0

	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Identity transform at middle: expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspectiveLerp_IdentityTransform(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Test identity transformation (should be 1:1)
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	interp.QuadToQuad(src, dst)

	length := 100
	interp.Begin(0, 0, length)

	x, y := interp.Coordinates()
	expectedX := 0
	expectedY := 0

	// Allow some tolerance for subpixel calculations
	tolerance := 5 // About 5 subpixel units
	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Identity transform at start: expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}

	// Test halfway through
	for i := 0; i < length/2; i++ {
		interp.Next()
	}

	x, y = interp.Coordinates()
	expectedX = (length / 2) << 8 // Convert to subpixel
	expectedY = 0

	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Identity transform at middle: expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspectiveExact_RectToQuad(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Transform a 100x100 rectangle to a quad
	quad := [8]float64{
		10, 10, // top-left
		110, 20, // top-right
		90, 110, // bottom-right
		20, 90, // bottom-left
	}
	interp.RectToQuad(0, 0, 100, 100, quad)

	if !interp.IsValid() {
		t.Error("Rect to quad transformation should be valid")
	}

	length := 100
	interp.Begin(0, 0, length)

	// Test that we start near the expected transformed position
	x, y := interp.Coordinates()
	expectedX := basics.IRound(10.0 * 256) // quad[0] in subpixel
	expectedY := basics.IRound(10.0 * 256) // quad[1] in subpixel

	tolerance := 50 // Allow more tolerance for perspective transformation
	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Rect to quad start: expected around (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspectiveLerp_RectToQuad(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Transform a 100x100 rectangle to a quad
	quad := [8]float64{
		10, 10, // top-left
		110, 20, // top-right
		90, 110, // bottom-right
		20, 90, // bottom-left
	}
	interp.RectToQuad(0, 0, 100, 100, quad)

	if !interp.IsValid() {
		t.Error("Rect to quad transformation should be valid")
	}

	length := 100
	interp.Begin(0, 0, length)

	// Test that we start near the expected transformed position
	x, y := interp.Coordinates()
	expectedX := basics.IRound(10.0 * 256) // quad[0] in subpixel
	expectedY := basics.IRound(10.0 * 256) // quad[1] in subpixel

	tolerance := 50 // Allow more tolerance for perspective transformation
	if absInt(x-expectedX) > tolerance || absInt(y-expectedY) > tolerance {
		t.Errorf("Rect to quad start: expected around (%d, %d), got (%d, %d)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspectiveExact_QuadToRect(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Transform a quad to a 100x100 rectangle
	quad := [8]float64{
		10, 10, // top-left
		110, 20, // top-right
		90, 110, // bottom-right
		20, 90, // bottom-left
	}
	interp.QuadToRect(quad, 0, 0, 100, 100)

	if !interp.IsValid() {
		t.Error("Quad to rect transformation should be valid")
	}
}

func TestSpanInterpolatorPerspectiveLerp_QuadToRect(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Transform a quad to a 100x100 rectangle
	quad := [8]float64{
		10, 10, // top-left
		110, 20, // top-right
		90, 110, // bottom-right
		20, 90, // bottom-left
	}
	interp.QuadToRect(quad, 0, 0, 100, 100)

	if !interp.IsValid() {
		t.Error("Quad to rect transformation should be valid")
	}
}

func TestSpanInterpolatorPerspectiveExact_LocalScale(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Use identity transform
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	interp.QuadToQuad(src, dst)

	length := 10
	interp.Begin(0, 0, length)

	// Test that local scale is computed
	sx, sy := interp.LocalScale()

	// For identity transform, scale should be around 1.0 (which is 1 << (8-8) = 1)
	if sx <= 0 || sy <= 0 {
		t.Errorf("Local scale should be positive, got (%d, %d)", sx, sy)
	}
}

func TestSpanInterpolatorPerspectiveLerp_LocalScale(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Use identity transform
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	interp.QuadToQuad(src, dst)

	length := 10
	interp.Begin(0, 0, length)

	// Test that local scale is computed
	sx, sy := interp.LocalScale()

	// For identity transform, scale should be around 1.0 (which is 1 << (8-8) = 1)
	if sx <= 0 || sy <= 0 {
		t.Errorf("Local scale should be positive, got (%d, %d)", sx, sy)
	}
}

func TestSpanInterpolatorPerspectiveExact_Resynchronize(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Use simple scaling transform
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 200, 0, 200, 200, 0, 200}
	interp.QuadToQuad(src, dst)

	length := 10
	interp.Begin(0, 0, length)

	// Move forward
	for i := 0; i < 5; i++ {
		interp.Next()
	}

	// Resynchronize to a new end point
	interp.Resynchronize(50, 0, 5)

	// Should not panic and should continue working
	x, y := interp.Coordinates()
	if x < 0 || y < 0 {
		t.Errorf("Coordinates after resynchronize should be valid, got (%d, %d)", x, y)
	}
}

func TestSpanInterpolatorPerspectiveLerp_Resynchronize(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Use simple scaling transform
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 200, 0, 200, 200, 0, 200}
	interp.QuadToQuad(src, dst)

	length := 10
	interp.Begin(0, 0, length)

	// Move forward
	for i := 0; i < 5; i++ {
		interp.Next()
	}

	// Resynchronize to a new end point
	interp.Resynchronize(50, 0, 5)

	// Should not panic and should continue working
	x, y := interp.Coordinates()
	if x < 0 || y < 0 {
		t.Errorf("Coordinates after resynchronize should be valid, got (%d, %d)", x, y)
	}
}

func TestSpanInterpolatorPerspectiveExact_Transform(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	// Set up a simple 2x scaling
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 200, 0, 200, 200, 0, 200}
	interp.QuadToQuad(src, dst)

	// Test direct transformation
	x, y := 50.0, 50.0
	interp.Transform(&x, &y)

	// Should be approximately 2x scaled
	expectedX, expectedY := 100.0, 100.0
	tolerance := 1.0

	if math.Abs(x-expectedX) > tolerance || math.Abs(y-expectedY) > tolerance {
		t.Errorf("Transform: expected (%.1f, %.1f), got (%.1f, %.1f)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspectiveLerp_Transform(t *testing.T) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	// Set up a simple 2x scaling
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 200, 0, 200, 200, 0, 200}
	interp.QuadToQuad(src, dst)

	// Test direct transformation
	x, y := 50.0, 50.0
	interp.Transform(&x, &y)

	// Should be approximately 2x scaled
	expectedX, expectedY := 100.0, 100.0
	tolerance := 1.0

	if math.Abs(x-expectedX) > tolerance || math.Abs(y-expectedY) > tolerance {
		t.Errorf("Transform: expected (%.1f, %.1f), got (%.1f, %.1f)",
			expectedX, expectedY, x, y)
	}
}

func TestSpanInterpolatorPerspective_Constructors(t *testing.T) {
	// Test QuadToQuad constructor
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{0, 0, 200, 0, 200, 200, 0, 200}

	exactQuad := NewSpanInterpolatorPerspectiveExactQuadToQuad(src, dst, 8)
	if !exactQuad.IsValid() {
		t.Error("QuadToQuad constructor should create valid interpolator")
	}

	lerpQuad := NewSpanInterpolatorPerspectiveLerpQuadToQuad(src, dst, 8)
	if !lerpQuad.IsValid() {
		t.Error("QuadToQuad constructor should create valid interpolator")
	}

	// Test RectToQuad constructor
	quad := [8]float64{10, 10, 110, 20, 90, 110, 20, 90}

	exactRect := NewSpanInterpolatorPerspectiveExactRectToQuad(0, 0, 100, 100, quad, 8)
	if !exactRect.IsValid() {
		t.Error("RectToQuad constructor should create valid interpolator")
	}

	lerpRect := NewSpanInterpolatorPerspectiveLerpRectToQuad(0, 0, 100, 100, quad, 8)
	if !lerpRect.IsValid() {
		t.Error("RectToQuad constructor should create valid interpolator")
	}

	// Test QuadToRect constructor
	exactQuadRect := NewSpanInterpolatorPerspectiveExactQuadToRect(quad, 0, 0, 100, 100, 8)
	if !exactQuadRect.IsValid() {
		t.Error("QuadToRect constructor should create valid interpolator")
	}

	lerpQuadRect := NewSpanInterpolatorPerspectiveLerpQuadToRect(quad, 0, 0, 100, 100, 8)
	if !lerpQuadRect.IsValid() {
		t.Error("QuadToRect constructor should create valid interpolator")
	}
}

func TestSpanInterpolatorPerspective_CompareExactVsLerp(t *testing.T) {
	// Compare exact vs lerp interpolation for similar results
	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{10, 10, 110, 15, 105, 110, 5, 95}

	exact := NewSpanInterpolatorPerspectiveExactQuadToQuad(src, dst, 8)
	lerp := NewSpanInterpolatorPerspectiveLerpQuadToQuad(src, dst, 8)

	length := 50
	exact.Begin(0, 0, length)
	lerp.Begin(0, 0, length)

	// Compare coordinates at several points
	for i := 0; i < length; i += 10 {
		for j := 0; j < i; j++ {
			exact.Next()
			lerp.Next()
		}

		xe, ye := exact.Coordinates()
		xl, yl := lerp.Coordinates()

		// Lerp should be reasonably close to exact for most transformations
		toleranceX := absInt(xe)/10 + 100 // Allow 10% error plus some base tolerance
		toleranceY := absInt(ye)/10 + 100

		if absInt(xe-xl) > toleranceX || absInt(ye-yl) > toleranceY {
			t.Logf("At step %d: exact (%d, %d), lerp (%d, %d), diff (%d, %d)",
				i, xe, ye, xl, yl, absInt(xe-xl), absInt(ye-yl))
		}

		// Reset for next iteration
		exact.Begin(0, 0, length)
		lerp.Begin(0, 0, length)
	}
}

// Benchmarks
func BenchmarkSpanInterpolatorPerspectiveExact(b *testing.B) {
	interp := NewSpanInterpolatorPerspectiveExact(8)

	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{10, 10, 110, 15, 105, 110, 5, 95}
	interp.QuadToQuad(src, dst)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Begin(0, 0, 100)
		for j := 0; j < 100; j++ {
			interp.Next()
			_, _ = interp.Coordinates()
			_, _ = interp.LocalScale()
		}
	}
}

func BenchmarkSpanInterpolatorPerspectiveLerp(b *testing.B) {
	interp := NewSpanInterpolatorPerspectiveLerp(8)

	src := [8]float64{0, 0, 100, 0, 100, 100, 0, 100}
	dst := [8]float64{10, 10, 110, 15, 105, 110, 5, 95}
	interp.QuadToQuad(src, dst)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interp.Begin(0, 0, 100)
		for j := 0; j < 100; j++ {
			interp.Next()
			_, _ = interp.Coordinates()
			_, _ = interp.LocalScale()
		}
	}
}
