package color

import (
	"testing"

	"agg_go/internal/basics"
)

// Test constructor and basic methods
func TestRGBA32Methods(t *testing.T) {
	c := NewRGBA32[Linear](0.4, 0.6, 0.8, 1.0)

	// Test IsOpaque
	if !c.IsOpaque() {
		t.Error("Color with alpha 1.0 should be opaque")
	}

	// Test IsTransparent
	c.A = 0
	if !c.IsTransparent() {
		t.Error("Color with alpha 0 should be transparent")
	}

	// Test Opacity
	c.Opacity(0.5)
	expected := float32(0.5)
	if abs32(c.A, expected) > 0.001 {
		t.Errorf("Opacity(0.5) set alpha to %f, expected %f", c.A, expected)
	}

	// Test GetOpacity
	opacity := c.GetOpacity()
	if abs64(opacity, 0.5) > 0.001 {
		t.Errorf("GetOpacity() = %.3f, expected 0.5", opacity)
	}

	// Test Clear
	c.Clear()
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 0 {
		t.Errorf("Clear() failed: got (%f,%f,%f,%f)", c.R, c.G, c.B, c.A)
	}

	// Test Transparent
	c = NewRGBA32[Linear](0.4, 0.6, 0.8, 1.0)
	c.Transparent()
	if c.R != 0.4 || c.G != 0.6 || c.B != 0.8 || c.A != 0 {
		t.Errorf("Transparent() should only clear alpha: got (%f,%f,%f,%f)", c.R, c.G, c.B, c.A)
	}
}

func TestRGBA32PremultiplyDemultiply(t *testing.T) {
	original := NewRGBA32[Linear](0.8, 0.4, 0.6, 0.5) // 50% alpha
	c := original

	// Test premultiplication
	c.Premultiply()

	// Values should be reduced proportionally to alpha
	if c.R > original.R || c.G > original.G || c.B > original.B {
		t.Error("Premultiplication should reduce RGB values")
	}
	if c.A != original.A {
		t.Error("Premultiplication should not change alpha")
	}

	// Test demultiplication
	c.Demultiply()

	// Should be close to original (some floating-point error expected)
	tolerance := float32(0.001)
	if abs32(c.R, original.R) > tolerance ||
		abs32(c.G, original.G) > tolerance ||
		abs32(c.B, original.B) > tolerance {
		t.Errorf("Demultiply didn't restore original: got (%f,%f,%f), expected (%f,%f,%f)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA32PremultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	c := NewRGBA32[Linear](0.8, 0.4, 0.6, 0.0)
	c.Premultiply()
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("Premultiply with A=0 should set RGB to 0, got (%f,%f,%f)", c.R, c.G, c.B)
	}

	// Test with alpha = 1.0 (should not change RGB)
	c = NewRGBA32[Linear](0.8, 0.4, 0.6, 1.0)
	originalR, originalG, originalB := c.R, c.G, c.B
	c.Premultiply()
	if c.R != originalR || c.G != originalG || c.B != originalB {
		t.Errorf("Premultiply with A=1.0 should not change RGB: original=(%f,%f,%f), got=(%f,%f,%f)",
			originalR, originalG, originalB, c.R, c.G, c.B)
	}

	// Test with very small alpha
	c = NewRGBA32[Linear](0.8, 0.4, 0.6, 0.01)
	c.Premultiply()
	if c.R >= 0.8 || c.G >= 0.4 || c.B >= 0.6 {
		t.Errorf("Premultiply with very small A should significantly reduce RGB, got (%f,%f,%f)", c.R, c.G, c.B)
	}
}

func TestRGBA32DemultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	c := NewRGBA32[Linear](0.8, 0.4, 0.6, 0.0)
	c.Demultiply()
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("Demultiply with A=0 should set RGB to 0, got (%f,%f,%f)", c.R, c.G, c.B)
	}

	// Test with alpha = 1.0 (should not change RGB significantly)
	c = NewRGBA32[Linear](0.8, 0.4, 0.6, 1.0)
	originalR, originalG, originalB := c.R, c.G, c.B
	c.Demultiply()
	tolerance := float32(0.001)
	if abs32(c.R, originalR) > tolerance ||
		abs32(c.G, originalG) > tolerance ||
		abs32(c.B, originalB) > tolerance {
		t.Errorf("Demultiply with A=1.0 should barely change RGB: original=(%f,%f,%f), got=(%f,%f,%f)",
			originalR, originalG, originalB, c.R, c.G, c.B)
	}
}

func TestRGBA32Gradient(t *testing.T) {
	c1 := NewRGBA32[Linear](0.0, 0.0, 0.0, 1.0) // Black
	c2 := NewRGBA32[Linear](1.0, 1.0, 1.0, 1.0) // White

	// 50% gradient should be gray
	mid := c1.Gradient(c2, 0.5)
	expected := float32(0.5)
	tolerance := float32(0.001)

	if abs32(mid.R, expected) > tolerance ||
		abs32(mid.G, expected) > tolerance ||
		abs32(mid.B, expected) > tolerance {
		t.Errorf("Gradient midpoint: got (%f,%f,%f), expected (~%f,~%f,~%f)",
			mid.R, mid.G, mid.B, expected, expected, expected)
	}

	// Test endpoints
	start := c1.Gradient(c2, 0.0)
	if abs32(start.R, c1.R) > tolerance || abs32(start.G, c1.G) > tolerance ||
		abs32(start.B, c1.B) > tolerance || abs32(start.A, c1.A) > tolerance {
		t.Errorf("Gradient at k=0.0 should return first color")
	}

	end := c1.Gradient(c2, 1.0)
	if abs32(end.R, c2.R) > tolerance || abs32(end.G, c2.G) > tolerance ||
		abs32(end.B, c2.B) > tolerance || abs32(end.A, c2.A) > tolerance {
		t.Errorf("Gradient at k=1.0 should return second color")
	}
}

func TestRGBA32Add(t *testing.T) {
	c1 := NewRGBA32[Linear](0.4, 0.2, 0.3, 0.8)
	c2 := NewRGBA32[Linear](0.2, 0.4, 0.1, 0.2)

	sum := c1.Add(c2)

	expectedR := float32(0.6) // 0.4 + 0.2
	expectedG := float32(0.6) // 0.2 + 0.4
	expectedB := float32(0.4) // 0.3 + 0.1
	expectedA := float32(1.0) // 0.8 + 0.2

	tolerance := float32(0.001)
	if abs32(sum.R, expectedR) > tolerance || abs32(sum.G, expectedG) > tolerance ||
		abs32(sum.B, expectedB) > tolerance || abs32(sum.A, expectedA) > tolerance {
		t.Errorf("Add result: got (%f,%f,%f,%f), expected (%f,%f,%f,%f)",
			sum.R, sum.G, sum.B, sum.A, expectedR, expectedG, expectedB, expectedA)
	}

	// Test that values don't clamp beyond reasonable floating point values
	c1 = NewRGBA32[Linear](0.8, 0.8, 0.8, 0.8)
	c2 = NewRGBA32[Linear](0.8, 0.8, 0.8, 0.8)
	sum = c1.Add(c2)
	// Results will be 1.6 for each component (no automatic clamping in Add)
	if sum.R != 1.6 || sum.G != 1.6 || sum.B != 1.6 || sum.A != 1.6 {
		t.Errorf("Add should not automatically clamp: got (%f,%f,%f,%f)", sum.R, sum.G, sum.B, sum.A)
	}
}

func TestRGBA32AddWithCover(t *testing.T) {
	c := NewRGBA32[Linear](0.4, 0.4, 0.4, 0.4)
	c2 := NewRGBA32[Linear](0.6, 0.6, 0.6, 0.6)

	// Test with full coverage (255)
	c1 := c
	c1.AddWithCover(c2, 255)
	expected := c.Add(c2)
	tolerance := float32(0.001)
	if abs32(c1.R, expected.R) > tolerance || abs32(c1.G, expected.G) > tolerance ||
		abs32(c1.B, expected.B) > tolerance || abs32(c1.A, expected.A) > tolerance {
		t.Errorf("AddWithCover(255) should behave same as Add")
	}

	// Test with partial coverage
	c1 = c
	original := c1
	c1.AddWithCover(c2, 128) // ~50% coverage
	// Values should increase but not as much as full coverage
	if c1.R <= original.R || c1.G <= original.G || c1.B <= original.B || c1.A <= original.A {
		t.Error("AddWithCover should increase component values")
	}
	fullAdd := original.Add(c2)
	if c1.R >= fullAdd.R || c1.G >= fullAdd.G || c1.B >= fullAdd.B || c1.A >= fullAdd.A {
		t.Error("AddWithCover partial should be less than full add")
	}

	// Test with zero coverage
	c1 = c
	original = c1
	c1.AddWithCover(c2, 0)
	if abs32(c1.R, original.R) > tolerance || abs32(c1.G, original.G) > tolerance ||
		abs32(c1.B, original.B) > tolerance || abs32(c1.A, original.A) > tolerance {
		t.Error("AddWithCover(0) should not change the color")
	}

	// Test AddWithCover clamping behavior
	c1 = NewRGBA32[Linear](0.8, 0.8, 0.8, 0.8)
	c2 = NewRGBA32[Linear](0.8, 0.8, 0.8, 0.8)
	c1.AddWithCover(c2, 255) // Full coverage
	// AddWithCover should clamp to 1.0 unlike Add
	if c1.R != 1.0 || c1.G != 1.0 || c1.B != 1.0 || c1.A != 1.0 {
		t.Errorf("AddWithCover should clamp to 1.0: got (%f,%f,%f,%f)", c1.R, c1.G, c1.B, c1.A)
	}
}

func TestRGBA32Scale(t *testing.T) {
	c := NewRGBA32[Linear](0.4, 0.6, 0.8, 1.0)
	scaled := c.Scale(0.5)

	expectedR := float32(0.2) // 0.4 * 0.5
	expectedG := float32(0.3) // 0.6 * 0.5
	expectedB := float32(0.4) // 0.8 * 0.5
	expectedA := float32(0.5) // 1.0 * 0.5

	tolerance := float32(0.001)
	if abs32(scaled.R, expectedR) > tolerance || abs32(scaled.G, expectedG) > tolerance ||
		abs32(scaled.B, expectedB) > tolerance || abs32(scaled.A, expectedA) > tolerance {
		t.Errorf("Scale(0.5) result: got (%f,%f,%f,%f), expected (%f,%f,%f,%f)",
			scaled.R, scaled.G, scaled.B, scaled.A, expectedR, expectedG, expectedB, expectedA)
	}

	// Test scaling by values > 1
	scaled2 := c.Scale(2.0)
	if scaled2.R != 0.8 || scaled2.G != 1.2 || scaled2.B != 1.6 || scaled2.A != 2.0 {
		t.Errorf("Scale(2.0) should not automatically clamp: got (%f,%f,%f,%f)",
			scaled2.R, scaled2.G, scaled2.B, scaled2.A)
	}
}

func TestRGBA32ConversionsFromToRGBA(t *testing.T) {
	// Test conversion from floating-point RGBA
	rgba := NewRGBA(0.5, 0.25, 0.75, 0.8)
	rgba32 := ConvertFromRGBA32[Linear](rgba)

	expectedR := float32(0.5)
	expectedG := float32(0.25)
	expectedB := float32(0.75)
	expectedA := float32(0.8)

	tolerance := float32(0.001)
	if abs32(rgba32.R, expectedR) > tolerance ||
		abs32(rgba32.G, expectedG) > tolerance ||
		abs32(rgba32.B, expectedB) > tolerance ||
		abs32(rgba32.A, expectedA) > tolerance {
		t.Errorf("ConvertFromRGBA32 result: got (%f,%f,%f,%f), expected (%f,%f,%f,%f)",
			rgba32.R, rgba32.G, rgba32.B, rgba32.A,
			expectedR, expectedG, expectedB, expectedA)
	}

	// Test conversion back to floating-point RGBA
	rgbaBack := rgba32.ConvertToRGBA()
	tolerance64 := 0.001

	if abs64(rgbaBack.R, rgba.R) > tolerance64 ||
		abs64(rgbaBack.G, rgba.G) > tolerance64 ||
		abs64(rgbaBack.B, rgba.B) > tolerance64 ||
		abs64(rgbaBack.A, rgba.A) > tolerance64 {
		t.Errorf("ConvertToRGBA roundtrip error: got (%.3f,%.3f,%.3f,%.3f), expected (%.3f,%.3f,%.3f,%.3f)",
			rgbaBack.R, rgbaBack.G, rgbaBack.B, rgbaBack.A,
			rgba.R, rgba.G, rgba.B, rgba.A)
	}
}

func TestRGBA32CommonTypes(t *testing.T) {
	// Test that type aliases work correctly
	var linear RGBA32Linear
	var srgb RGBA32SRGB

	linear = NewRGBA32[Linear](0.5, 0.5, 0.5, 1.0)
	srgb = NewRGBA32[SRGB](0.5, 0.5, 0.5, 1.0)

	if linear.R != 0.5 || srgb.R != 0.5 {
		t.Error("Type aliases should work correctly")
	}
}

func TestRGBA32BoundaryValues(t *testing.T) {
	// Test with minimum values
	c := NewRGBA32[Linear](0.0, 0.0, 0.0, 0.0)
	if !c.IsTransparent() {
		t.Error("Color with all zeros should be transparent")
	}
	if c.IsOpaque() {
		t.Error("Color with A=0 should not be opaque")
	}

	// Test with maximum values
	c = NewRGBA32[Linear](1.0, 1.0, 1.0, 1.0)
	if c.IsTransparent() {
		t.Error("Color with A=1.0 should not be transparent")
	}
	if !c.IsOpaque() {
		t.Error("Color with A=1.0 should be opaque")
	}

	// Test with boundary alpha values
	c = NewRGBA32[Linear](0.5, 0.5, 0.5, 0.001)
	if c.IsTransparent() || c.IsOpaque() {
		t.Error("Color with A=0.001 should be neither transparent nor opaque")
	}

	c = NewRGBA32[Linear](0.5, 0.5, 0.5, 0.999)
	if c.IsTransparent() || c.IsOpaque() {
		t.Error("Color with A=0.999 should be neither transparent nor opaque")
	}
}

func TestRGBA32OpacityClamp(t *testing.T) {
	c := NewRGBA32[Linear](0.5, 0.5, 0.5, 0.5)

	// Test negative opacity
	c.Opacity(-0.1)
	if c.A != 0.0 {
		t.Errorf("Opacity(-0.1) should clamp to 0.0, got %f", c.A)
	}

	// Test opacity > 1.0
	c.Opacity(1.1)
	if c.A != 1.0 {
		t.Errorf("Opacity(1.1) should clamp to 1.0, got %f", c.A)
	}

	// Test normal opacity
	c.Opacity(0.25)
	expected := float32(0.25)
	tolerance := float32(0.001)
	if abs32(c.A, expected) > tolerance {
		t.Errorf("Opacity(0.25) expected %f, got %f", expected, c.A)
	}
}

func TestRGBA32FloatingPointPrecision(t *testing.T) {
	// Test that operations maintain reasonable floating-point precision
	c1 := NewRGBA32[Linear](0.1, 0.2, 0.3, 0.4)
	c2 := NewRGBA32[Linear](0.5, 0.6, 0.7, 0.8)

	// Test multiple operations in sequence
	result := c1.Add(c2)
	result = result.Scale(0.5)
	result = result.Gradient(c1, 0.5)

	// Should still be finite values
	if result.R != result.R || result.G != result.G || result.B != result.B || result.A != result.A {
		t.Error("Operations resulted in NaN values")
	}

	// Should be reasonable values
	if result.R < -10.0 || result.R > 10.0 ||
		result.G < -10.0 || result.G > 10.0 ||
		result.B < -10.0 || result.B > 10.0 ||
		result.A < -10.0 || result.A > 10.0 {
		t.Errorf("Operations resulted in unreasonable values: (%f,%f,%f,%f)", result.R, result.G, result.B, result.A)
	}
}

func TestRGBA32ComprehensiveRoundTrip(t *testing.T) {
	// Test multiple round trips with various values
	testValues := []struct{ r, g, b, a float32 }{
		{0.0, 0.0, 0.0, 0.0},
		{1.0, 1.0, 1.0, 1.0},
		{0.5, 0.5, 0.5, 0.5},
		{0.25, 0.75, 0.5, 0.625},
		{0.001, 0.999, 0.5, 0.25},
	}

	for _, tv := range testValues {
		original := NewRGBA32[Linear](tv.r, tv.g, tv.b, tv.a)

		// Round trip: RGBA32 -> RGBA -> RGBA32
		rgba := original.ConvertToRGBA()
		recovered := ConvertFromRGBA32[Linear](rgba)

		tolerance := float32(0.001)
		if abs32(recovered.R, original.R) > tolerance ||
			abs32(recovered.G, original.G) > tolerance ||
			abs32(recovered.B, original.B) > tolerance ||
			abs32(recovered.A, original.A) > tolerance {
			t.Errorf("Round trip drift too large: orig=(%f,%f,%f,%f) recovered=(%f,%f,%f,%f)",
				original.R, original.G, original.B, original.A,
				recovered.R, recovered.G, recovered.B, recovered.A)
		}
	}
}

func TestRGBA32PremultiplyDemultiplyRoundTrip(t *testing.T) {
	cases := []struct{ r, g, b, a float32 }{
		{0.0, 0.0, 0.0, 0.0},
		{1.0, 1.0, 1.0, 0.0},
		{0.0, 0.0, 0.0, 1.0},
		{1.0, 1.0, 1.0, 1.0},
		{0.8, 0.4, 0.6, 0.001},
		{0.75, 0.5, 0.25, 0.01},
		{0.6, 0.3, 0.9, 0.5},
	}

	for _, c := range cases {
		color := NewRGBA32[Linear](c.r, c.g, c.b, c.a)
		original := color
		color.Premultiply()
		color.Demultiply()

		if c.a == 0.0 {
			// With zero alpha, RGB should be zero after demultiply
			if color.R != 0.0 || color.G != 0.0 || color.B != 0.0 {
				t.Fatalf("A=0 should force RGB=0 after demultiply, got (%f,%f,%f)", color.R, color.G, color.B)
			}
			continue
		}

		// For very small alpha values, some precision loss is expected but should still be reasonable
		if c.a <= 0.01 {
			// With very small alpha, some precision loss is expected but values should be finite
			if color.R != color.R || color.G != color.G || color.B != color.B {
				t.Fatalf("Very small alpha resulted in NaN: (%f,%f,%f) with A=%f", color.R, color.G, color.B, c.a)
			}
			continue
		}

		// For reasonable alpha values, check round-trip accuracy
		tolerance := float32(0.01) // Reasonable tolerance for floating-point operations
		if abs32(color.R, original.R) > tolerance ||
			abs32(color.G, original.G) > tolerance ||
			abs32(color.B, original.B) > tolerance {
			t.Errorf("Round-trip drift too large for RGB: orig=(%f,%f,%f) back=(%f,%f,%f) (A=%f)",
				original.R, original.G, original.B, color.R, color.G, color.B, c.a)
		}
		if abs32(color.A, original.A) > tolerance {
			t.Errorf("Alpha changed on round-trip: orig=%f back=%f", original.A, color.A)
		}
	}
}

func TestRGBA32GradientEdgeCases(t *testing.T) {
	c1 := NewRGBA32[Linear](0.1, 0.2, 0.3, 0.4)
	c2 := NewRGBA32[Linear](0.9, 0.8, 0.7, 0.6)

	// Test various gradient positions
	testK := []float32{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, k := range testK {
		result := c1.Gradient(c2, k)

		// Results should be finite
		if result.R != result.R || result.G != result.G ||
			result.B != result.B || result.A != result.A {
			t.Errorf("Gradient at k=%f resulted in NaN", k)
		}

		// Results should be within reasonable bounds for interpolation
		minR, maxR := c1.R, c2.R
		if minR > maxR {
			minR, maxR = maxR, minR
		}
		if result.R < minR-0.001 || result.R > maxR+0.001 {
			t.Errorf("Gradient R component out of bounds at k=%f: got %f, bounds [%f,%f]", k, result.R, minR, maxR)
		}
	}
}

func TestRGBA32AddWithCoverEdgeCases(t *testing.T) {
	c := NewRGBA32[Linear](0.2, 0.3, 0.4, 0.5)
	c2 := NewRGBA32[Linear](0.6, 0.7, 0.8, 0.9)

	// Test with opaque second color and full coverage
	c2Opaque := NewRGBA32[Linear](0.6, 0.7, 0.8, 1.0)
	cTest := c
	cTest.AddWithCover(c2Opaque, 255)

	// Should replace the color when second color is opaque and coverage is full
	tolerance := float32(0.001)
	if abs32(cTest.R, c2Opaque.R) > tolerance ||
		abs32(cTest.G, c2Opaque.G) > tolerance ||
		abs32(cTest.B, c2Opaque.B) > tolerance ||
		abs32(cTest.A, c2Opaque.A) > tolerance {
		t.Errorf("AddWithCover with opaque color and full coverage should replace color")
	}

	// Test coverage edge cases
	coverageValues := []basics.Int8u{0, 1, 127, 128, 254, 255}
	for _, cover := range coverageValues {
		cTest := c
		original := cTest
		cTest.AddWithCover(c2, cover)

		// Results should be finite
		if cTest.R != cTest.R || cTest.G != cTest.G ||
			cTest.B != cTest.B || cTest.A != cTest.A {
			t.Errorf("AddWithCover with cover=%d resulted in NaN", cover)
		}

		// With zero coverage, should not change
		if cover == 0 {
			if abs32(cTest.R, original.R) > tolerance ||
				abs32(cTest.G, original.G) > tolerance ||
				abs32(cTest.B, original.B) > tolerance ||
				abs32(cTest.A, original.A) > tolerance {
				t.Errorf("AddWithCover with cover=0 should not change color")
			}
		}
	}
}
