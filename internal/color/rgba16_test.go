package color

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/gamma"
)

// Test arithmetic functions
func TestRGBA16Arithmetic(t *testing.T) {
	// Test RGBA16Multiply
	result := RGBA16Multiply(32768, 32768)
	expected := basics.Int16u(16384) // 32768*32768/65535 ≈ 16384
	tolerance := basics.Int16u(10)   // Allow some tolerance for rounding
	if abs16u(result, expected) > tolerance {
		t.Errorf("RGBA16Multiply(32768, 32768) = %d, expected ~%d", result, expected)
	}

	// Test with edge cases
	if RGBA16Multiply(0, 65535) != 0 {
		t.Error("RGBA16Multiply(0, 65535) should be 0")
	}
	if RGBA16Multiply(65535, 65535) != 65535 {
		t.Error("RGBA16Multiply(65535, 65535) should be 65535")
	}
}

func TestRGBA16Lerp(t *testing.T) {
	// Test linear interpolation
	result := RGBA16Lerp(0, 65535, 32768) // ~50% between 0 and 65535
	expected := basics.Int16u(32768)
	tolerance := basics.Int16u(100)
	if abs16u(result, expected) > tolerance {
		t.Errorf("RGBA16Lerp(0, 65535, 32768) = %d, expected ~%d", result, expected)
	}

	// Test edge cases
	if RGBA16Lerp(10000, 20000, 0) != 10000 {
		t.Error("RGBA16Lerp with alpha 0 should return first value")
	}
	if RGBA16Lerp(10000, 20000, 65535) != 20000 {
		t.Error("RGBA16Lerp with alpha 65535 should return second value")
	}
}

func TestRGBA16Prelerp(t *testing.T) {
	// Test premultiplied lerp - this should be same as Lerp for RGBA16
	result := RGBA16Prelerp(10000, 5000, 32768)
	expected := RGBA16Lerp(10000, 5000, 32768)
	if result != expected {
		t.Errorf("RGBA16Prelerp(10000, 5000, 32768) = %d, expected %d", result, expected)
	}
}

func TestRGBA16MultCover(t *testing.T) {
	// Test coverage multiplication
	result := RGBA16MultCover(32768, 32768)
	expected := RGBA16Multiply(32768, 32768)
	if result != expected {
		t.Errorf("RGBA16MultCover should behave same as RGBA16Multiply")
	}
}

func TestRGBA16Methods(t *testing.T) {
	c := NewRGBA16[Linear](25600, 38400, 51200, 65535)

	// Test IsOpaque
	if !c.IsOpaque() {
		t.Error("Color with alpha 65535 should be opaque")
	}

	// Test IsTransparent
	c.A = 0
	if !c.IsTransparent() {
		t.Error("Color with alpha 0 should be transparent")
	}

	// Test Opacity
	c.Opacity(0.5)
	expected := basics.Int16u(32768) // 50% of 65535
	tolerance := basics.Int16u(100)
	if abs16u(c.A, expected) > tolerance {
		t.Errorf("Opacity(0.5) set alpha to %d, expected ~%d", c.A, expected)
	}

	// Test GetOpacity
	opacity := c.GetOpacity()
	if abs64(opacity, 0.5) > 0.01 {
		t.Errorf("GetOpacity() = %.3f, expected ~0.5", opacity)
	}

	// Test Clear
	c.Clear()
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 0 {
		t.Errorf("Clear() failed: got (%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}

	// Test Transparent
	c = NewRGBA16[Linear](25600, 38400, 51200, 65535)
	c.Transparent()
	if c.R != 25600 || c.G != 38400 || c.B != 51200 || c.A != 0 {
		t.Errorf("Transparent() should only clear alpha: got (%d,%d,%d,%d)", c.R, c.G, c.B, c.A)
	}
}

func TestRGBA16PremultiplyDemultiply(t *testing.T) {
	original := NewRGBA16[Linear](32768, 16384, 49152, 32768) // 50% alpha
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
	tolerance := basics.Int16u(500) // Higher tolerance for 16-bit
	if abs16u(c.R, original.R) > tolerance ||
		abs16u(c.G, original.G) > tolerance ||
		abs16u(c.B, original.B) > tolerance {
		t.Errorf("Demultiply didn't restore original: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA16PremultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	c := NewRGBA16[Linear](32768, 16384, 49152, 0)
	c.Premultiply()
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("Premultiply with A=0 should set RGB to 0, got (%d,%d,%d)", c.R, c.G, c.B)
	}

	// Test with alpha = 65535 (should not change RGB)
	c = NewRGBA16[Linear](32768, 16384, 49152, 65535)
	originalR, originalG, originalB := c.R, c.G, c.B
	c.Premultiply()
	if c.R != originalR || c.G != originalG || c.B != originalB {
		t.Errorf("Premultiply with A=65535 should not change RGB: original=(%d,%d,%d), got=(%d,%d,%d)",
			originalR, originalG, originalB, c.R, c.G, c.B)
	}

	// Test with very small alpha
	c = NewRGBA16[Linear](32768, 16384, 49152, 100)
	c.Premultiply()
	if c.R >= 32768 || c.G >= 16384 || c.B >= 49152 {
		t.Errorf("Premultiply with very small A should significantly reduce RGB, got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestRGBA16DemultiplyEdgeCases(t *testing.T) {
	// Test with zero alpha
	c := NewRGBA16[Linear](32768, 16384, 49152, 0)
	c.Demultiply()
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("Demultiply with A=0 should set RGB to 0, got (%d,%d,%d)", c.R, c.G, c.B)
	}

	// Test with alpha = 65535 (should not change RGB significantly)
	c = NewRGBA16[Linear](32768, 16384, 49152, 65535)
	originalR, originalG, originalB := c.R, c.G, c.B
	c.Demultiply()
	tolerance := basics.Int16u(10)
	if abs16u(c.R, originalR) > tolerance ||
		abs16u(c.G, originalG) > tolerance ||
		abs16u(c.B, originalB) > tolerance {
		t.Errorf("Demultiply with A=65535 should barely change RGB: original=(%d,%d,%d), got=(%d,%d,%d)",
			originalR, originalG, originalB, c.R, c.G, c.B)
	}
}

func TestRGBA16Gradient(t *testing.T) {
	c1 := NewRGBA16[Linear](0, 0, 0, 65535)             // Black
	c2 := NewRGBA16[Linear](65535, 65535, 65535, 65535) // White

	// 50% gradient should be gray
	mid := c1.Gradient(c2, 32768)
	expected := basics.Int16u(32768)
	tolerance := basics.Int16u(500)

	if abs16u(mid.R, expected) > tolerance ||
		abs16u(mid.G, expected) > tolerance ||
		abs16u(mid.B, expected) > tolerance {
		t.Errorf("Gradient midpoint: got (%d,%d,%d), expected (~%d,~%d,~%d)",
			mid.R, mid.G, mid.B, expected, expected, expected)
	}

	// Test endpoints
	start := c1.Gradient(c2, 0)
	if start.R != c1.R || start.G != c1.G || start.B != c1.B || start.A != c1.A {
		t.Errorf("Gradient at k=0 should return first color")
	}

	end := c1.Gradient(c2, 65535)
	if end.R != c2.R || end.G != c2.G || end.B != c2.B || end.A != c2.A {
		t.Errorf("Gradient at k=65535 should return second color")
	}
}

func TestRGBA16Add(t *testing.T) {
	c1 := NewRGBA16[Linear](25600, 12800, 19200, 51200)
	c2 := NewRGBA16[Linear](12800, 25600, 6400, 14080)

	sum := c1.Add(c2)

	expectedR := basics.Int16u(38400) // 25600 + 12800
	expectedG := basics.Int16u(38400) // 12800 + 25600
	expectedB := basics.Int16u(25600) // 19200 + 6400
	expectedA := basics.Int16u(65280) // 51200 + 14080, should be 65280

	if sum.R != expectedR || sum.G != expectedG || sum.B != expectedB || sum.A != expectedA {
		t.Errorf("Add result: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			sum.R, sum.G, sum.B, sum.A, expectedR, expectedG, expectedB, expectedA)
	}

	// Test overflow clamping
	c1 = NewRGBA16[Linear](50000, 50000, 50000, 50000)
	c2 = NewRGBA16[Linear](50000, 50000, 50000, 50000)
	sum = c1.Add(c2)
	if sum.R != 65535 || sum.G != 65535 || sum.B != 65535 || sum.A != 65535 {
		t.Errorf("Add overflow should clamp to 65535: got (%d,%d,%d,%d)", sum.R, sum.G, sum.B, sum.A)
	}
}

func TestRGBA16AddWithCover(t *testing.T) {
	c := NewRGBA16[Linear](25600, 25600, 25600, 25600)
	c2 := NewRGBA16[Linear](38400, 38400, 38400, 38400)

	// Test with full coverage (255)
	c1 := c
	c1.AddWithCover(c2, 255)
	expected := c.Add(c2)
	if c1.R != expected.R || c1.G != expected.G || c1.B != expected.B || c1.A != expected.A {
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
	if c1.R != original.R || c1.G != original.G || c1.B != original.B || c1.A != original.A {
		t.Error("AddWithCover(0) should not change the color")
	}
}

func TestRGBA16ConversionsFromToRGBA(t *testing.T) {
	// Test conversion from floating-point
	rgba := NewRGBA(0.5, 0.25, 0.75, 0.8)
	rgba16 := ConvertFromRGBA16[Linear](rgba)

	expectedR := basics.Int16u(32768) // 0.5*65535 + 0.5
	expectedG := basics.Int16u(16384) // 0.25*65535 + 0.5
	expectedB := basics.Int16u(49151) // 0.75*65535 + 0.5
	expectedA := basics.Int16u(52428) // 0.8*65535 + 0.5

	tolerance := basics.Int16u(10)
	if abs16u(rgba16.R, expectedR) > tolerance ||
		abs16u(rgba16.G, expectedG) > tolerance ||
		abs16u(rgba16.B, expectedB) > tolerance ||
		abs16u(rgba16.A, expectedA) > tolerance {
		t.Errorf("ConvertFromRGBA16 result: got (%d,%d,%d,%d), expected (%d,%d,%d,%d)",
			rgba16.R, rgba16.G, rgba16.B, rgba16.A,
			expectedR, expectedG, expectedB, expectedA)
	}

	// Test conversion back to floating-point
	rgbaBack := rgba16.ConvertToRGBA()
	tolerance64 := 0.01

	if abs64(rgbaBack.R, rgba.R) > tolerance64 ||
		abs64(rgbaBack.G, rgba.G) > tolerance64 ||
		abs64(rgbaBack.B, rgba.B) > tolerance64 ||
		abs64(rgbaBack.A, rgba.A) > tolerance64 {
		t.Errorf("ConvertToRGBA roundtrip error: got (%.3f,%.3f,%.3f,%.3f), expected (%.3f,%.3f,%.3f,%.3f)",
			rgbaBack.R, rgbaBack.G, rgbaBack.B, rgbaBack.A,
			rgba.R, rgba.G, rgba.B, rgba.A)
	}
}

func TestRGBA16CommonTypes(t *testing.T) {
	// Test that type aliases work correctly
	var linear RGBA16Linear
	var srgb RGBA16SRGB

	linear = NewRGBA16[Linear](32768, 32768, 32768, 65535)
	srgb = NewRGBA16[SRGB](32768, 32768, 32768, 65535)

	if linear.R != 32768 || srgb.R != 32768 {
		t.Error("Type aliases should work correctly")
	}
}

func TestRGBA16BoundaryValues(t *testing.T) {
	// Test with minimum values
	c := NewRGBA16[Linear](0, 0, 0, 0)
	if !c.IsTransparent() {
		t.Error("Color with all zeros should be transparent")
	}
	if c.IsOpaque() {
		t.Error("Color with A=0 should not be opaque")
	}

	// Test with maximum values
	c = NewRGBA16[Linear](65535, 65535, 65535, 65535)
	if c.IsTransparent() {
		t.Error("Color with A=65535 should not be transparent")
	}
	if !c.IsOpaque() {
		t.Error("Color with A=65535 should be opaque")
	}

	// Test with boundary alpha values
	c = NewRGBA16[Linear](32768, 32768, 32768, 1)
	if c.IsTransparent() || c.IsOpaque() {
		t.Error("Color with A=1 should be neither transparent nor opaque")
	}

	c = NewRGBA16[Linear](32768, 32768, 32768, 65534)
	if c.IsTransparent() || c.IsOpaque() {
		t.Error("Color with A=65534 should be neither transparent nor opaque")
	}
}

func TestRGBA16OpacityClamp(t *testing.T) {
	c := NewRGBA16[Linear](32768, 32768, 32768, 32768)

	// Test negative opacity
	c.Opacity(-0.1)
	if c.A != 0 {
		t.Errorf("Opacity(-0.1) should clamp to 0, got %d", c.A)
	}

	// Test opacity > 1.0
	c.Opacity(1.1)
	if c.A != 65535 {
		t.Errorf("Opacity(1.1) should clamp to 65535, got %d", c.A)
	}

	// Test normal opacity
	c.Opacity(0.25)
	expected := basics.Int16u(16384) // 0.25 * 65535
	tolerance := basics.Int16u(10)
	if abs16u(c.A, expected) > tolerance {
		t.Errorf("Opacity(0.25) expected ~%d, got %d", expected, c.A)
	}
}

func TestRGBA16ArithmeticProperties(t *testing.T) {
	// Test multiply properties
	for a := basics.Int16u(0); a < 65535; a += 6553 { // Test every ~10%
		if RGBA16Multiply(a, 0) != 0 {
			t.Fatalf("a*0 != 0 for a=%d", a)
		}
		if RGBA16Multiply(a, 65535) != a {
			t.Fatalf("a*65535 != a for a=%d, got %d", a, RGBA16Multiply(a, 65535))
		}

		for b := basics.Int16u(0); b < 65535; b += 13107 { // Test fewer combinations
			if RGBA16Multiply(a, b) != RGBA16Multiply(b, a) {
				t.Fatalf("commutativity broken: a=%d b=%d", a, b)
			}
		}
	}
}

func TestRGBA16LerpEndpoints(t *testing.T) {
	if RGBA16Lerp(1000, 2000, 0) != 1000 {
		t.Fatal("a=0 should return p")
	}
	if RGBA16Lerp(1000, 2000, 65535) != 2000 {
		t.Fatal("a=65535 should return q")
	}

	// Test p>q case
	result := RGBA16Lerp(2000, 1000, 32768) // 50%
	expected := basics.Int16u(1500)
	tolerance := basics.Int16u(50)
	if abs16u(result, expected) > tolerance {
		t.Fatalf("p>q 50%% expected ~%d, got %d", expected, result)
	}
}

func TestRGBA16ComprehensiveRoundTrip(t *testing.T) {
	// Test multiple round trips with various values
	testValues := []struct{ r, g, b, a basics.Int16u }{
		{0, 0, 0, 0},
		{65535, 65535, 65535, 65535},
		{32768, 32768, 32768, 32768},
		{16384, 49151, 32768, 40960},
		{1, 65534, 32767, 16383},
	}

	for _, tv := range testValues {
		original := NewRGBA16[Linear](tv.r, tv.g, tv.b, tv.a)

		// Round trip: RGBA16 -> RGBA -> RGBA16
		rgba := original.ConvertToRGBA()
		recovered := ConvertFromRGBA16[Linear](rgba)

		tolerance := basics.Int16u(50) // Allow reasonable tolerance
		if abs16u(recovered.R, original.R) > tolerance ||
			abs16u(recovered.G, original.G) > tolerance ||
			abs16u(recovered.B, original.B) > tolerance ||
			abs16u(recovered.A, original.A) > tolerance {
			t.Errorf("Round trip drift too large: orig=(%d,%d,%d,%d) recovered=(%d,%d,%d,%d)",
				original.R, original.G, original.B, original.A,
				recovered.R, recovered.G, recovered.B, recovered.A)
		}
	}
}

func TestRGBA16PremultiplyDemultiplyRoundTrip(t *testing.T) {
	cases := []struct{ r, g, b, a basics.Int16u }{
		{0, 0, 0, 0},
		{65535, 65535, 65535, 0},
		{0, 0, 0, 65535},
		{65535, 65535, 65535, 65535},
		{32768, 16384, 49152, 1},
		{51200, 25600, 38400, 100},
		{40960, 20480, 30720, 32768},
	}

	for _, c := range cases {
		color := NewRGBA16[Linear](c.r, c.g, c.b, c.a)
		original := color
		color.Premultiply()
		color.Demultiply()

		if c.a == 0 {
			// With zero alpha, RGB should be zero after demultiply
			if color.R != 0 || color.G != 0 || color.B != 0 {
				t.Fatalf("A=0 should force RGB=0 after demultiply, got (%d,%d,%d)", color.R, color.G, color.B)
			}
			continue
		}

		// For very small alpha values, precision loss is expected
		if c.a <= 100 {
			// With very small alpha, precision loss is expected and acceptable
			continue
		}

		// For reasonable alpha values, check round-trip accuracy
		tolerance := basics.Int16u(500) // Higher tolerance for 16-bit operations
		if abs16u(color.R, original.R) > tolerance ||
			abs16u(color.G, original.G) > tolerance ||
			abs16u(color.B, original.B) > tolerance {
			t.Errorf("Round-trip drift too large for RGB: orig=(%d,%d,%d) back=(%d,%d,%d) (A=%d)",
				original.R, original.G, original.B, color.R, color.G, color.B, c.a)
		}
		if color.A != original.A {
			t.Errorf("Alpha changed on round-trip: orig=%d back=%d", original.A, color.A)
		}
	}
}

func TestRGBA16ApplyGamma(t *testing.T) {
	lut := gamma.NewGammaLUT16WithGamma(2.0)

	c := NewRGBA16[Linear](32768, 16384, 49152, 65535)
	original := c

	c.ApplyGammaDir(lut)

	if c.A != original.A {
		t.Errorf("ApplyGammaDir changed alpha: got %d, expected %d", c.A, original.A)
	}
	if c.R == original.R && c.G == original.G && c.B == original.B {
		t.Error("ApplyGammaDir should change RGB values")
	}

	// For gamma=2.0, values should be reduced (darker)
	if c.R >= original.R || c.G >= original.G || c.B >= original.B {
		t.Errorf("ApplyGammaDir with gamma=2.0 should reduce values: got (%d,%d,%d), original (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}

	c.ApplyGammaInv(lut)

	tolerance := basics.Int16u(1000)
	if abs16u(c.R, original.R) > tolerance ||
		abs16u(c.G, original.G) > tolerance ||
		abs16u(c.B, original.B) > tolerance {
		t.Errorf("ApplyGammaInv didn't restore original within tolerance: got (%d,%d,%d), expected (%d,%d,%d)",
			c.R, c.G, c.B, original.R, original.G, original.B)
	}
}

func TestRGBA16ApplyGammaEdgeCases(t *testing.T) {
	lut := gamma.NewGammaLUT16WithGamma(2.0)

	// Test with all zeros
	c := NewRGBA16[Linear](0, 0, 0, 32768)
	c.ApplyGammaDir(lut)
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 32768 {
		t.Errorf("Gamma on zeros: got (%d,%d,%d,%d), expected (0,0,0,32768)", c.R, c.G, c.B, c.A)
	}

	// Test with max values
	c = NewRGBA16[Linear](65535, 65535, 65535, 65535)
	c.ApplyGammaDir(lut)
	if c.R != 65535 || c.G != 65535 || c.B != 65535 || c.A != 65535 {
		t.Errorf("Gamma on max values: got (%d,%d,%d,%d), expected (65535,65535,65535,65535)", c.R, c.G, c.B, c.A)
	}
}
