package color

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestRGB16(t *testing.T) {
	// Test NewRGB16
	rgb := NewRGB16[Linear](32768, 49152, 16384)
	if rgb.R != 32768 || rgb.G != 49152 || rgb.B != 16384 {
		t.Errorf("NewRGB16 failed: got %v, want {32768, 49152, 16384}", rgb)
	}

	// Test Clear
	rgb.Clear()
	if rgb.R != 0 || rgb.G != 0 || rgb.B != 0 {
		t.Errorf("Clear failed: got %v, want {0, 0, 0}", rgb)
	}

	// Test ConvertToRGB
	rgb = NewRGB16[Linear](65535, 32768, 0)
	frgb := rgb.ConvertToRGB()
	expected := RGB{R: 1.0, G: 0.5, B: 0.0}
	if math.Abs(frgb.R-expected.R) > 1e-6 ||
		math.Abs(frgb.G-expected.G) > 1e-3 || // Allow slightly more error for 16-bit
		math.Abs(frgb.B-expected.B) > 1e-6 {
		t.Errorf("ConvertToRGB failed: got %v, want %v", frgb, expected)
	}

	// Test ConvertFromRGB16
	frgb = RGB{R: 1.0, G: 0.5, B: 0.0}
	rgb = ConvertFromRGB16[Linear](frgb)
	if rgb.R != 65535 || rgb.G != 32768 || rgb.B != 0 {
		t.Errorf("ConvertFromRGB16 failed: got %v, want {65535, 32768, 0}", rgb)
	}

	// Test ConvertToRGBA
	rgba := rgb.ConvertToRGBA()
	if math.Abs(rgba.A-1.0) > 1e-6 {
		t.Errorf("ConvertToRGBA alpha failed: got %f, want 1.0", rgba.A)
	}

	// Test ToRGBA16
	rgb = NewRGB16[Linear](25600, 38400, 51200)
	rgba16 := rgb.ToRGBA16()
	if rgba16.R != 25600 || rgba16.G != 38400 || rgba16.B != 51200 || rgba16.A != 65535 {
		t.Errorf("ToRGBA16 failed: got %v, want {25600, 38400, 51200, 65535}", rgba16)
	}

	// Test Add
	rgb1 := NewRGB16[Linear](25600, 25600, 25600)
	rgb2 := NewRGB16[Linear](12800, 19200, 6400)
	sum := rgb1.Add(rgb2)
	if sum.R != 38400 || sum.G != 44800 || sum.B != 32000 {
		t.Errorf("Add failed: got %v, want {38400, 44800, 32000}", sum)
	}

	// Test Add with overflow
	rgb1 = NewRGB16[Linear](51200, 51200, 51200)
	rgb2 = NewRGB16[Linear](25600, 25600, 25600)
	sum = rgb1.Add(rgb2)
	if sum.R != 65535 || sum.G != 65535 || sum.B != 65535 {
		t.Errorf("Add overflow failed: got %v, want {65535, 65535, 65535}", sum)
	}

	// Test Scale
	rgb = NewRGB16[Linear](25600, 38400, 51200)
	scaled := rgb.Scale(1.5)
	if scaled.R != 38400 || scaled.G != 57600 || scaled.B != 65535 {
		t.Errorf("Scale failed: got %v, want {38400, 57600, 65535}", scaled)
	}

	// Test Gradient
	rgb1 = NewRGB16[Linear](0, 0, 0)
	rgb2 = NewRGB16[Linear](65535, 65535, 65535)
	mid := rgb1.Gradient(rgb2, 32768) // ~50% interpolation
	// Should be approximately halfway
	if mid.R < 30000 || mid.R > 35535 || mid.G < 30000 || mid.G > 35535 || mid.B < 30000 || mid.B > 35535 {
		t.Errorf("Gradient failed: got %v, expected around {32767, 32767, 32767}", mid)
	}

	// Test IsBlack and IsWhite
	black := NewRGB16[Linear](0, 0, 0)
	white := NewRGB16[Linear](65535, 65535, 65535)
	gray := NewRGB16[Linear](32768, 32768, 32768)

	if !black.IsBlack() {
		t.Error("IsBlack failed for black color")
	}
	if !white.IsWhite() {
		t.Error("IsWhite failed for white color")
	}
	if gray.IsBlack() || gray.IsWhite() {
		t.Error("IsBlack/IsWhite failed for gray color")
	}

	// Test Luminance
	// Test with pure colors
	red := NewRGB16[Linear](65535, 0, 0)
	green := NewRGB16[Linear](0, 65535, 0)
	blue := NewRGB16[Linear](0, 0, 65535)

	redLum := red.Luminance()
	greenLum := green.Luminance()
	blueLum := blue.Luminance()

	// Green should have the highest luminance, blue the lowest
	if greenLum <= redLum || redLum <= blueLum {
		t.Errorf("Luminance calculation incorrect: R=%d, G=%d, B=%d", redLum, greenLum, blueLum)
	}

	// Test luminance for white should be close to 65535
	whiteLum := white.Luminance()
	if whiteLum < 65000 { // Allow some error due to integer arithmetic
		t.Errorf("White luminance too low: got %d, expected ~65535", whiteLum)
	}

	// Test luminance for black should be 0
	blackLum := black.Luminance()
	if blackLum != 0 {
		t.Errorf("Black luminance incorrect: got %d, expected 0", blackLum)
	}
}

func TestRGB16Conversions(t *testing.T) {
	// Test ConvertRGBAToRGB16
	rgba := NewRGBA(0.4, 0.6, 0.8, 0.5) // Alpha should be ignored
	rgb := ConvertRGBAToRGB16[Linear](rgba)
	expectedR := basics.Int16u(26214) // 0.4*65535 + 0.5 ≈ 26214
	expectedG := basics.Int16u(39321) // 0.6*65535 + 0.5 ≈ 39321
	expectedB := basics.Int16u(52428) // 0.8*65535 + 0.5 ≈ 52428
	if rgb.R != expectedR || rgb.G != expectedG || rgb.B != expectedB {
		t.Errorf("ConvertRGBAToRGB16 failed: got %v, want {%d, %d, %d}", rgb, expectedR, expectedG, expectedB)
	}

	// Test ConvertToRGBA from RGB16
	rgb16 := NewRGB16[Linear](13107, 19661, 26214) // ~0.2, 0.3, 0.4 in 16-bit
	rgba64 := rgb16.ConvertToRGBA()
	expectedRf := float64(13107) / 65535.0
	expectedGf := float64(19661) / 65535.0
	expectedBf := float64(26214) / 65535.0
	if math.Abs(rgba64.R-expectedRf) > 1e-4 ||
		math.Abs(rgba64.G-expectedGf) > 1e-4 ||
		math.Abs(rgba64.B-expectedBf) > 1e-4 ||
		math.Abs(rgba64.A-1.0) > 1e-6 {
		t.Errorf("ConvertToRGBA failed: got %v, want {%f, %f, %f, 1.0}", rgba64, expectedRf, expectedGf, expectedBf)
	}
}

func TestRGB16EdgeCases(t *testing.T) {
	// Test maximum values
	maxRgb := NewRGB16[Linear](65535, 65535, 65535)

	// Add should clamp at maximum
	sum := maxRgb.Add(NewRGB16[Linear](1, 1, 1))
	if sum.R != 65535 || sum.G != 65535 || sum.B != 65535 {
		t.Errorf("Add should clamp at maximum: got %v, want {65535, 65535, 65535}", sum)
	}

	// Scale should clamp overflow
	scaled := maxRgb.Scale(1.1)
	if scaled.R != 65535 || scaled.G != 65535 || scaled.B != 65535 {
		t.Errorf("Scale should clamp overflow: got %v, want {65535, 65535, 65535}", scaled)
	}

	// Test zero values
	zero := NewRGB16[Linear](0, 0, 0)
	if !zero.IsBlack() {
		t.Error("Zero values should be considered black")
	}

	// Test Gradient boundary cases
	black := NewRGB16[Linear](0, 0, 0)
	white := NewRGB16[Linear](65535, 65535, 65535)

	// k=0 should return first color
	grad0 := black.Gradient(white, 0)
	if grad0.R != 0 || grad0.G != 0 || grad0.B != 0 {
		t.Errorf("Gradient with k=0 should return first color: got %v", grad0)
	}

	// k=65535 should return second color
	grad1 := black.Gradient(white, 65535)
	if grad1.R != 65535 || grad1.G != 65535 || grad1.B != 65535 {
		t.Errorf("Gradient with k=65535 should return second color: got %v", grad1)
	}
}

func TestRGB16ColorspaceConversion(t *testing.T) {
	// Test conversion from Linear to sRGB using the conversion functions
	linearRgb := NewRGB16[Linear](32768, 49152, 16384)

	// Convert to RGBA32 for colorspace conversion (since RGB16 doesn't have direct colorspace conversion)
	rgba32 := RGBA32[Linear]{
		R: float32(linearRgb.R) / 65535.0,
		G: float32(linearRgb.G) / 65535.0,
		B: float32(linearRgb.B) / 65535.0,
		A: 1.0,
	}
	srgbRgba32 := ConvertRGBA32LinearToSRGB(rgba32)

	// Convert back to RGB16
	convertedSRGB := NewRGB16[SRGB](
		uint16(srgbRgba32.R*65535+0.5),
		uint16(srgbRgba32.G*65535+0.5),
		uint16(srgbRgba32.B*65535+0.5),
	)

	// Values should be different after gamma correction
	if convertedSRGB.R == linearRgb.R && convertedSRGB.G == linearRgb.G && convertedSRGB.B == linearRgb.B {
		t.Error("Linear to sRGB conversion should change values")
	}

	// Test specific known conversion values
	// Black should remain black in both colorspaces
	blackLinear := NewRGB16[Linear](0, 0, 0)
	if !blackLinear.IsBlack() {
		t.Error("Black should remain black regardless of colorspace")
	}

	// White should remain white in both colorspaces
	whiteLinear := NewRGB16[Linear](65535, 65535, 65535)
	if !whiteLinear.IsWhite() {
		t.Error("White should remain white regardless of colorspace")
	}
}

func TestRGB16HelperFunctions(t *testing.T) {
	// Test RGB16Lerp function directly
	result := RGB16Lerp(0, 65535, 32768) // 50% between 0 and 65535
	expected := uint16(32767)            // Should be approximately halfway
	if result < expected-1 || result > expected+1 {
		t.Errorf("RGB16Lerp failed: got %d, expected ~%d", result, expected)
	}

	// Test endpoints
	if RGB16Lerp(100, 200, 0) != 100 {
		t.Error("RGB16Lerp with alpha 0 should return first value")
	}
	if RGB16Lerp(100, 200, 65535) != 200 {
		t.Error("RGB16Lerp with alpha 65535 should return second value")
	}

	// Test RGB16Prelerp (premultiplied interpolation)
	prelerpResult := RGB16Prelerp(16384, 49152, 32768)
	// This should be similar to regular lerp for this case
	if prelerpResult < 30000 || prelerpResult > 35000 {
		t.Errorf("RGB16Prelerp result seems out of range: got %d", prelerpResult)
	}

	// Test RGB16MultCover
	multResult := RGB16MultCover(32768, 32768) // 50% * 50%
	expectedMult := uint16(16384)              // Should be approximately 25%
	if multResult < expectedMult-1000 || multResult > expectedMult+1000 {
		t.Errorf("RGB16MultCover failed: got %d, expected ~%d", multResult, expectedMult)
	}
}

func TestRGB16Constants(t *testing.T) {
	// Test predefined color constants (these are RGB8 constants but should exist)
	if RGB8Black.R != 0 || RGB8Black.G != 0 || RGB8Black.B != 0 {
		t.Error("RGB8Black constant incorrect")
	}
	if RGB8White.R != 255 || RGB8White.G != 255 || RGB8White.B != 255 {
		t.Error("RGB8White constant incorrect")
	}
	if RGB8Red.R != 255 || RGB8Red.G != 0 || RGB8Red.B != 0 {
		t.Error("RGB8Red constant incorrect")
	}
	if RGB8Green.R != 0 || RGB8Green.G != 255 || RGB8Green.B != 0 {
		t.Error("RGB8Green constant incorrect")
	}
	if RGB8Blue.R != 0 || RGB8Blue.G != 0 || RGB8Blue.B != 255 {
		t.Error("RGB8Blue constant incorrect")
	}
	if RGB8Cyan.R != 0 || RGB8Cyan.G != 255 || RGB8Cyan.B != 255 {
		t.Error("RGB8Cyan constant incorrect")
	}
	if RGB8Magenta.R != 255 || RGB8Magenta.G != 0 || RGB8Magenta.B != 255 {
		t.Error("RGB8Magenta constant incorrect")
	}
	if RGB8Yellow.R != 255 || RGB8Yellow.G != 255 || RGB8Yellow.B != 0 {
		t.Error("RGB8Yellow constant incorrect")
	}
}

func TestRGB16ColorOrders(t *testing.T) {
	// Test RGB24 and BGR24 order
	if OrderRGB24.R != 0 || OrderRGB24.G != 1 || OrderRGB24.B != 2 || OrderRGB24.A != -1 {
		t.Error("OrderRGB24 incorrect")
	}
	if OrderBGR24.R != 2 || OrderBGR24.G != 1 || OrderBGR24.B != 0 || OrderBGR24.A != -1 {
		t.Error("OrderBGR24 incorrect")
	}
}

func TestRGB16ExplicitConversion(t *testing.T) {
	t.Run("Convert method", func(t *testing.T) {
		originalRGB := NewRGB16[Linear](25600, 38400, 51200)
		convertedRGB := originalRGB.Convert()

		if convertedRGB.R != originalRGB.R || convertedRGB.G != originalRGB.G || convertedRGB.B != originalRGB.B {
			t.Errorf("Convert should return identical values: got %v, want %v", convertedRGB, originalRGB)
		}
	})

	t.Run("Round-trip RGB conversion", func(t *testing.T) {
		original := NewRGB16[Linear](25600, 38400, 51200)

		// Convert to float and back
		rgbFloat := original.ConvertToRGB()
		roundTrip := ConvertFromRGB16[Linear](rgbFloat)

		// Allow small error due to float precision
		if absUint16(roundTrip.R, original.R) > 1 ||
			absUint16(roundTrip.G, original.G) > 1 ||
			absUint16(roundTrip.B, original.B) > 1 {
			t.Errorf("Round-trip conversion failed: original=%v, roundtrip=%v", original, roundTrip)
		}
	})
}

// Helper function for absolute difference between uint16 values
func absUint16(a, b uint16) uint16 {
	if a > b {
		return a - b
	}
	return b - a
}

// Benchmark tests
func BenchmarkRGB16Gradient(b *testing.B) {
	rgb1 := NewRGB16[Linear](0, 0, 0)
	rgb2 := NewRGB16[Linear](65535, 65535, 65535)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := uint16(i & 65535)
		_ = rgb1.Gradient(rgb2, k)
	}
}

func BenchmarkRGB16Add(b *testing.B) {
	rgb1 := NewRGB16[Linear](25600, 25600, 25600)
	rgb2 := NewRGB16[Linear](12800, 19200, 6400)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb1.Add(rgb2)
	}
}

func BenchmarkRGB16Luminance(b *testing.B) {
	rgb := NewRGB16[Linear](32768, 49152, 16384)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb.Luminance()
	}
}
