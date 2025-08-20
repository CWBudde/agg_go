package color

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestRGBA8Arithmetic(t *testing.T) {
	// Test RGBA8Multiply
	result := RGBA8Multiply(128, 128)
	expected := basics.Int8u(64) // 128*128/255 ≈ 64
	if result != expected {
		t.Errorf("RGBA8Multiply(128, 128) = %d, expected %d", result, expected)
	}

	// Test with edge cases
	if RGBA8Multiply(0, 255) != 0 {
		t.Error("RGBA8Multiply(0, 255) should be 0")
	}
	if RGBA8Multiply(255, 255) != 255 {
		t.Error("RGBA8Multiply(255, 255) should be 255")
	}
}

func TestRGBA8Lerp(t *testing.T) {
	// Test linear interpolation
	result := RGBA8Lerp(0, 255, 128) // 50% between 0 and 255
	expected := basics.Int8u(128)
	if result < expected-1 || result > expected+1 {
		t.Errorf("RGBA8Lerp(0, 255, 128) = %d, expected ~%d", result, expected)
	}

	// Test edge cases
	if RGBA8Lerp(100, 200, 0) != 100 {
		t.Error("RGBA8Lerp with alpha 0 should return first value")
	}
	if RGBA8Lerp(100, 200, 255) != 200 {
		t.Error("RGBA8Lerp with alpha 255 should return second value")
	}
}

func TestRGBA8Prelerp(t *testing.T) {
	// Test premultiplied lerp
	result := RGBA8Prelerp(100, 50, 128)
	expected := 100 + 50 - RGBA8Multiply(100, 128)
	if result != expected {
		t.Errorf("RGBA8Prelerp(100, 50, 128) = %d, expected %d", result, expected)
	}
}

func TestRGBA8Methods(t *testing.T) {
	c := NewRGBA8[Linear](100, 150, 200, 255)

	// Test IsOpaque
	if !c.IsOpaque() {
		t.Error("Color with alpha 255 should be opaque")
	}

	// Test IsTransparent
	c.A = 0
	if !c.IsTransparent() {
		t.Error("Color with alpha 0 should be transparent")
	}

	// Test Opacity
	c.Opacity(0.5)
	expected := basics.Int8u(128) // 50% of 255
	if c.A < expected-1 || c.A > expected+1 {
		t.Errorf("Opacity(0.5) set alpha to %d, expected ~%d", c.A, expected)
	}

	// Test GetOpacity
	opacity := c.GetOpacity()
	if opacity < 0.49 || opacity > 0.51 {
		t.Errorf("GetOpacity() = %.3f, expected ~0.5", opacity)
	}
}

func TestRGBA8PremultiplyDemultiply(t *testing.T) {
	original := NewRGBA8[Linear](128, 64, 192, 128) // 50% alpha
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

	// Should be close to original (some rounding error expected)
	tolerance := basics.Int8u(2)
	if absInt8u(c.R, original.R) > tolerance ||
		absInt8u(c.G, original.G) > tolerance ||
		absInt8u(c.B, original.B) > tolerance {
		t.Errorf("Demultiply didn't restore original: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA8Gradient(t *testing.T) {
	c1 := NewRGBA8[Linear](0, 0, 0, 255)       // Black
	c2 := NewRGBA8[Linear](255, 255, 255, 255) // White

	// 50% gradient should be gray
	mid := c1.Gradient(c2, 128)
	expected := basics.Int8u(128)
	tolerance := basics.Int8u(2)

	if absInt8u(mid.R, expected) > tolerance ||
		absInt8u(mid.G, expected) > tolerance ||
		absInt8u(mid.B, expected) > tolerance {
		t.Errorf("Gradient midpoint: got (%d,%d,%d), expected (~%d,~%d,~%d)",
			mid.R, mid.G, mid.B, expected, expected, expected)
	}
}

func TestRGBA8Add(t *testing.T) {
	c1 := NewRGBA8[Linear](100, 50, 75, 200)
	c2 := NewRGBA8[Linear](50, 100, 25, 55)

	sum := c1.Add(c2)

	if sum.R != 150 || sum.G != 150 || sum.B != 100 || sum.A != 255 {
		t.Errorf("Add result: got (%d,%d,%d,%d), expected (150,150,100,255)",
			sum.R, sum.G, sum.B, sum.A)
	}
}

func TestRGBA8Scale(t *testing.T) {
	c := NewRGBA8[Linear](100, 150, 200, 255)
	scaled := c.Scale(0.5)

	if scaled.R != 50 || scaled.G != 75 || scaled.B != 100 {
		t.Errorf("Scale(0.5) result: got (%d,%d,%d), expected (50,75,100)",
			scaled.R, scaled.G, scaled.B)
	}
}

func TestRGBA8ConversionsFromToRGBA(t *testing.T) {
	// Test conversion from floating-point
	rgba := NewRGBA(0.5, 0.25, 0.75, 0.8)
	rgba8 := ConvertFromRGBA[Linear](rgba)

	expectedR := basics.Int8u(128) // 0.5*255 + 0.5
	expectedG := basics.Int8u(64)  // 0.25*255 + 0.5
	expectedB := basics.Int8u(191) // 0.75*255 + 0.5 = 191.75 -> 191
	expectedA := basics.Int8u(204) // 0.8*255 + 0.5

	if rgba8.R != expectedR || rgba8.G != expectedG ||
		rgba8.B != expectedB || rgba8.A != expectedA {
		t.Errorf("ConvertFromRGBA result: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			rgba8.R, rgba8.G, rgba8.B, rgba8.A,
			expectedR, expectedG, expectedB, expectedA)
	}

	// Test conversion back to floating-point
	rgbaBack := rgba8.ConvertToRGBA()
	tolerance := 0.01

	if abs64(rgbaBack.R, rgba.R) > tolerance ||
		abs64(rgbaBack.G, rgba.G) > tolerance ||
		abs64(rgbaBack.B, rgba.B) > tolerance ||
		abs64(rgbaBack.A, rgba.A) > tolerance {
		t.Errorf("ConvertToRGBA roundtrip error: got (%.3f,%.3f,%.3f,%.3f), expected (%.3f,%.3f,%.3f,%.3f)",
			rgbaBack.R, rgbaBack.G, rgbaBack.B, rgbaBack.A,
			rgba.R, rgba.G, rgba.B, rgba.A)
	}
}

func TestRGBA8CommonTypes(t *testing.T) {
	// Test that type aliases work correctly
	var linear RGBA8Linear
	var srgb RGBA8SRGB
	var srgba SRGBA8

	linear = NewRGBA8[Linear](128, 128, 128, 255)
	srgb = NewRGBA8[SRGB](128, 128, 128, 255)
	srgba = NewRGBA8[SRGB](128, 128, 128, 255)

	if linear.R != 128 || srgb.R != 128 || srgba.R != 128 {
		t.Error("Type aliases should work correctly")
	}
}

// Test implementations of gamma interfaces
type testGammaLUT struct {
	gamma float64
}

func newTestGammaLUT(gamma float64) *testGammaLUT {
	return &testGammaLUT{gamma: gamma}
}

func (g *testGammaLUT) Dir(v basics.Int8u) basics.Int8u {
	// Apply gamma correction using proper power function
	normalized := float64(v) / 255.0
	if normalized <= 0 {
		return 0
	}
	corrected := math.Pow(normalized, g.gamma)
	if corrected > 1.0 {
		corrected = 1.0
	}
	return basics.Int8u(corrected*255 + 0.5)
}

func (g *testGammaLUT) Inv(v basics.Int8u) basics.Int8u {
	// Apply inverse gamma correction (gamma^-1)
	normalized := float64(v) / 255.0
	if normalized <= 0 {
		return 0
	}
	corrected := math.Pow(normalized, 1.0/g.gamma)
	if corrected > 1.0 {
		corrected = 1.0
	}
	return basics.Int8u(corrected*255 + 0.5)
}

type testGammaFunc struct {
	gamma float64
}

func newTestGammaFunc(gamma float64) *testGammaFunc {
	return &testGammaFunc{gamma: gamma}
}

func (g *testGammaFunc) DirFloat(v float32) float32 {
	if v <= 0 {
		return 0
	}
	result := math.Pow(float64(v), g.gamma)
	if result > 1.0 {
		result = 1.0
	}
	return float32(result)
}

func (g *testGammaFunc) InvFloat(v float32) float32 {
	if v <= 0 {
		return 0
	}
	result := math.Pow(float64(v), 1.0/g.gamma)
	if result > 1.0 {
		result = 1.0
	}
	return float32(result)
}

func TestRGBA8ApplyGamma(t *testing.T) {
	gamma := newTestGammaLUT(2.0) // Square gamma for testing

	// Test ApplyGammaDir
	c := NewRGBA8[Linear](128, 64, 192, 255)
	original := c
	c.ApplyGammaDir(gamma)

	// RGB values should change, alpha should remain the same
	if c.A != original.A {
		t.Errorf("ApplyGammaDir changed alpha: got %d, expected %d", c.A, original.A)
	}

	// For gamma=2.0, values should be reduced (darker)
	if c.R >= original.R || c.G >= original.G || c.B >= original.B {
		t.Errorf("ApplyGammaDir with gamma=2.0 should reduce values: got (%d,%d,%d), original (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}

	// Test ApplyGammaInv
	c.ApplyGammaInv(gamma)

	// Values should be approximately back to original (some rounding error expected)
	tolerance := basics.Int8u(5) // Allow some tolerance for rounding
	if absInt8u(c.R, original.R) > tolerance ||
		absInt8u(c.G, original.G) > tolerance ||
		absInt8u(c.B, original.B) > tolerance {
		t.Errorf("ApplyGammaInv didn't restore original: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA8ApplyGammaEdgeCases(t *testing.T) {
	gamma := newTestGammaLUT(2.0)

	// Test with all zeros
	c := NewRGBA8[Linear](0, 0, 0, 128)
	c.ApplyGammaDir(gamma)
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 128 {
		t.Errorf("Gamma on zeros: got (%d,%d,%d,%d), expected (0,0,0,128)", c.R, c.G, c.B, c.A)
	}

	// Test with max values
	c = NewRGBA8[Linear](255, 255, 255, 255)
	c.ApplyGammaDir(gamma)
	if c.R != 255 || c.G != 255 || c.B != 255 || c.A != 255 {
		t.Errorf("Gamma on max values: got (%d,%d,%d,%d), expected (255,255,255,255)", c.R, c.G, c.B, c.A)
	}
}

func TestRGBA16ApplyGamma(t *testing.T) {
	gamma := newTestGammaLUT(2.0)

	// Test ApplyGammaDir
	c := NewRGBA16[Linear](32768, 16384, 49152, 65535) // ~50%, ~25%, ~75%, 100%
	original := c
	c.ApplyGammaDir(gamma)

	// Alpha should remain unchanged
	if c.A != original.A {
		t.Errorf("ApplyGammaDir changed alpha: got %d, expected %d", c.A, original.A)
	}

	// RGB values should change
	if c.R == original.R && c.G == original.G && c.B == original.B {
		t.Error("ApplyGammaDir should change RGB values")
	}

	// Test ApplyGammaInv
	c.ApplyGammaInv(gamma)

	// Values should be approximately back to original
	tolerance := basics.Int16u(1000) // Allow tolerance for 8-bit conversion rounding
	if abs16u(c.R, original.R) > tolerance ||
		abs16u(c.G, original.G) > tolerance ||
		abs16u(c.B, original.B) > tolerance {
		t.Errorf("ApplyGammaInv didn't restore original within tolerance: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA32ApplyGamma(t *testing.T) {
	gammaFunc := newTestGammaFunc(2.0)

	// Test ApplyGammaDir
	c := NewRGBA32[Linear](0.5, 0.25, 0.75, 1.0)
	original := c
	c.ApplyGammaDir(gammaFunc)

	// Alpha should remain unchanged
	if c.A != original.A {
		t.Errorf("ApplyGammaDir changed alpha: got %.3f, expected %.3f", c.A, original.A)
	}

	// For gamma=2.0, values should be reduced (squared)
	if c.R >= original.R || c.G >= original.G || c.B >= original.B {
		t.Errorf("ApplyGammaDir with gamma=2.0 should reduce values: got (%.3f,%.3f,%.3f), original (%.3f,%.3f,%.3f)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}

	// Test ApplyGammaInv
	c.ApplyGammaInv(gammaFunc)

	// Values should be approximately back to original
	tolerance := float32(0.1) // Allow some tolerance
	if abs32(c.R, original.R) > tolerance ||
		abs32(c.G, original.G) > tolerance ||
		abs32(c.B, original.B) > tolerance {
		t.Errorf("ApplyGammaInv didn't restore original: got (%.3f,%.3f,%.3f), expected (%.3f,%.3f,%.3f)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestGammaIdentity(t *testing.T) {
	identityGamma := newTestGammaLUT(1.0) // Identity gamma

	c := NewRGBA8[Linear](128, 64, 192, 255)
	original := c

	c.ApplyGammaDir(identityGamma)
	if c.R != original.R || c.G != original.G || c.B != original.B || c.A != original.A {
		t.Errorf("Identity gamma should not change values: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			c.R, c.G, c.B, c.A, original.R, original.G, original.B, original.A)
	}

	c.ApplyGammaInv(identityGamma)
	if c.R != original.R || c.G != original.G || c.B != original.B || c.A != original.A {
		t.Errorf("Identity inverse gamma should not change values: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			c.R, c.G, c.B, c.A, original.R, original.G, original.B, original.A)
	}
}

// Helper functions for tests
func absInt8u(a, b basics.Int8u) basics.Int8u {
	if a > b {
		return a - b
	}
	return b - a
}

func abs16u(a, b basics.Int16u) basics.Int16u {
	if a > b {
		return a - b
	}
	return b - a
}

func abs32(a, b float32) float32 {
	if a > b {
		return a - b
	}
	return b - a
}

func abs64(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
