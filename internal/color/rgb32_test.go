package color

import (
	"math"
	"testing"
)

func TestRGB32(t *testing.T) {
	// Test NewRGB32
	rgb := NewRGB32[Linear](0.5, 0.75, 0.25)
	if rgb.R != 0.5 || rgb.G != 0.75 || rgb.B != 0.25 {
		t.Errorf("NewRGB32 failed: got %v, want {0.5, 0.75, 0.25}", rgb)
	}

	// Test Clear
	rgb.Clear()
	if rgb.R != 0.0 || rgb.G != 0.0 || rgb.B != 0.0 {
		t.Errorf("Clear failed: got %v, want {0, 0, 0}", rgb)
	}

	// Test ConvertToRGB
	rgb = NewRGB32[Linear](1.0, 0.5, 0.0)
	frgb := rgb.ConvertToRGB()
	expected := RGB{R: 1.0, G: 0.5, B: 0.0}
	if math.Abs(frgb.R-expected.R) > 1e-6 ||
		math.Abs(frgb.G-expected.G) > 1e-6 ||
		math.Abs(frgb.B-expected.B) > 1e-6 {
		t.Errorf("ConvertToRGB failed: got %v, want %v", frgb, expected)
	}

	// Test ConvertFromRGB32
	frgb = RGB{R: 1.0, G: 0.5, B: 0.0}
	rgb = ConvertFromRGB32[Linear](frgb)
	if math.Abs(float64(rgb.R)-1.0) > 1e-6 ||
		math.Abs(float64(rgb.G)-0.5) > 1e-6 ||
		math.Abs(float64(rgb.B)-0.0) > 1e-6 {
		t.Errorf("ConvertFromRGB32 failed: got %v, want {1.0, 0.5, 0.0}", rgb)
	}

	// Test ConvertToRGBA
	rgba := rgb.ConvertToRGBA()
	if math.Abs(rgba.A-1.0) > 1e-6 {
		t.Errorf("ConvertToRGBA alpha failed: got %f, want 1.0", rgba.A)
	}

	// Test ToRGBA32
	rgb = NewRGB32[Linear](0.4, 0.6, 0.8)
	rgba32 := rgb.ToRGBA32()
	if math.Abs(float64(rgba32.R)-0.4) > 1e-6 ||
		math.Abs(float64(rgba32.G)-0.6) > 1e-6 ||
		math.Abs(float64(rgba32.B)-0.8) > 1e-6 ||
		math.Abs(float64(rgba32.A)-1.0) > 1e-6 {
		t.Errorf("ToRGBA32 failed: got %v, want {0.4, 0.6, 0.8, 1.0}", rgba32)
	}

	// Test Add
	rgb1 := NewRGB32[Linear](0.4, 0.4, 0.4)
	rgb2 := NewRGB32[Linear](0.2, 0.3, 0.1)
	sum := rgb1.Add(rgb2)
	if math.Abs(float64(sum.R)-0.6) > 1e-6 ||
		math.Abs(float64(sum.G)-0.7) > 1e-6 ||
		math.Abs(float64(sum.B)-0.5) > 1e-6 {
		t.Errorf("Add failed: got %v, want {0.6, 0.7, 0.5}", sum)
	}

	// Test Scale
	rgb = NewRGB32[Linear](0.4, 0.6, 0.8)
	scaled := rgb.Scale(1.5)
	if math.Abs(float64(scaled.R)-0.6) > 1e-6 ||
		math.Abs(float64(scaled.G)-0.9) > 1e-6 ||
		math.Abs(float64(scaled.B)-1.2) > 1e-6 {
		t.Errorf("Scale failed: got %v, want {0.6, 0.9, 1.2}", scaled)
	}

	// Test Gradient
	rgb1 = NewRGB32[Linear](0.0, 0.0, 0.0)
	rgb2 = NewRGB32[Linear](1.0, 1.0, 1.0)
	mid := rgb1.Gradient(rgb2, 0.5)
	if math.Abs(float64(mid.R)-0.5) > 1e-6 ||
		math.Abs(float64(mid.G)-0.5) > 1e-6 ||
		math.Abs(float64(mid.B)-0.5) > 1e-6 {
		t.Errorf("Gradient failed: got %v, expected {0.5, 0.5, 0.5}", mid)
	}

	// Test IsBlack and IsWhite
	black := NewRGB32[Linear](0.0, 0.0, 0.0)
	white := NewRGB32[Linear](1.0, 1.0, 1.0)
	gray := NewRGB32[Linear](0.5, 0.5, 0.5)

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
	red := NewRGB32[Linear](1.0, 0.0, 0.0)
	green := NewRGB32[Linear](0.0, 1.0, 0.0)
	blue := NewRGB32[Linear](0.0, 0.0, 1.0)

	redLum := red.Luminance()
	greenLum := green.Luminance()
	blueLum := blue.Luminance()

	// Green should have the highest luminance, blue the lowest
	if greenLum <= redLum || redLum <= blueLum {
		t.Errorf("Luminance calculation incorrect: R=%f, G=%f, B=%f", redLum, greenLum, blueLum)
	}

	// Test luminance for white should be close to 1.0
	whiteLum := white.Luminance()
	if math.Abs(float64(whiteLum)-1.0) > 1e-6 {
		t.Errorf("White luminance incorrect: got %f, expected 1.0", whiteLum)
	}

	// Test luminance for black should be 0
	blackLum := black.Luminance()
	if math.Abs(float64(blackLum)) > 1e-6 {
		t.Errorf("Black luminance incorrect: got %f, expected 0.0", blackLum)
	}
}

func TestRGB32Conversions(t *testing.T) {
	// Test ConvertRGBAToRGB32
	rgba := NewRGBA(0.4, 0.6, 0.8, 0.5) // Alpha should be ignored
	rgb := ConvertRGBAToRGB32[Linear](rgba)
	if math.Abs(float64(rgb.R)-0.4) > 1e-6 ||
		math.Abs(float64(rgb.G)-0.6) > 1e-6 ||
		math.Abs(float64(rgb.B)-0.8) > 1e-6 {
		t.Errorf("ConvertRGBAToRGB32 failed: got %v, want {0.4, 0.6, 0.8}", rgb)
	}

	// Test ConvertToRGBA from RGB32
	rgb32 := NewRGB32[Linear](0.2, 0.3, 0.4)
	rgba64 := rgb32.ConvertToRGBA()
	if math.Abs(rgba64.R-0.2) > 1e-6 ||
		math.Abs(rgba64.G-0.3) > 1e-6 ||
		math.Abs(rgba64.B-0.4) > 1e-6 ||
		math.Abs(rgba64.A-1.0) > 1e-6 {
		t.Errorf("ConvertToRGBA failed: got %v, want {0.2, 0.3, 0.4, 1.0}", rgba64)
	}
}

func TestRGB32EdgeCases(t *testing.T) {
	// Test extreme values
	extreme := NewRGB32[Linear](-0.5, 1.5, 0.0)

	// Add should work with negative values
	positive := NewRGB32[Linear](0.7, 0.2, 0.3)
	sum := extreme.Add(positive)
	if math.Abs(float64(sum.R)-0.2) > 1e-6 ||
		math.Abs(float64(sum.G)-1.7) > 1e-6 ||
		math.Abs(float64(sum.B)-0.3) > 1e-6 {
		t.Errorf("Add with negative values failed: got %v, want {0.2, 1.7, 0.3}", sum)
	}

	// Scale should work with negative multiplier
	scaled := positive.Scale(-1.0)
	if math.Abs(float64(scaled.R)+0.7) > 1e-6 ||
		math.Abs(float64(scaled.G)+0.2) > 1e-6 ||
		math.Abs(float64(scaled.B)+0.3) > 1e-6 {
		t.Errorf("Scale with negative multiplier failed: got %v, want {-0.7, -0.2, -0.3}", scaled)
	}

	// Test epsilon boundary cases for IsBlack and IsWhite
	almostBlack := NewRGB32[Linear](1e-7, 1e-7, 1e-7)
	almostWhite := NewRGB32[Linear](0.999999, 0.999999, 0.999999)

	if !almostBlack.IsBlack() {
		t.Error("Almost black should be considered black")
	}
	if !almostWhite.IsWhite() {
		t.Error("Almost white should be considered white")
	}
}

func TestRGB32ColorspaceConversion(t *testing.T) {
	// Test conversion from Linear to sRGB using the conversion functions
	linearRgb := NewRGB32[Linear](0.5, 0.75, 0.25)
	srgbRgb := ConvertRGBA32LinearToSRGB(linearRgb.ToRGBA32())

	// Convert back to RGB32
	convertedSRGB := NewRGB32[SRGB](srgbRgb.R, srgbRgb.G, srgbRgb.B)

	// Values should be different after gamma correction
	if math.Abs(float64(convertedSRGB.R)-float64(linearRgb.R)) < 0.01 &&
		math.Abs(float64(convertedSRGB.G)-float64(linearRgb.G)) < 0.01 &&
		math.Abs(float64(convertedSRGB.B)-float64(linearRgb.B)) < 0.01 {
		t.Error("Linear to sRGB conversion should change values significantly")
	}
}

func TestRGB32HelperFunctions(t *testing.T) {
	// Test RGB helper functions that are part of rgb32.go

	// Test RGB16Lerp
	result := RGB16Lerp(0, 65535, 32768) // 50% between 0 and 65535
	expected := uint16(32767)            // Should be approximately halfway
	if result < expected-1 || result > expected+1 {
		t.Errorf("RGB16Lerp failed: got %d, expected ~%d", result, expected)
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

func TestRGB32Constants(t *testing.T) {
	// Test predefined color constants
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

func TestRGB32ColorOrders(t *testing.T) {
	// Test RGB24 and BGR24 order
	if OrderRGB24.R != 0 || OrderRGB24.G != 1 || OrderRGB24.B != 2 || OrderRGB24.A != -1 {
		t.Error("OrderRGB24 incorrect")
	}
	if OrderBGR24.R != 2 || OrderBGR24.G != 1 || OrderBGR24.B != 0 || OrderBGR24.A != -1 {
		t.Error("OrderBGR24 incorrect")
	}
}

// Benchmark tests
func BenchmarkRGB32Gradient(b *testing.B) {
	rgb1 := NewRGB32[Linear](0.0, 0.0, 0.0)
	rgb2 := NewRGB32[Linear](1.0, 1.0, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := float32(i&255) / 255.0
		_ = rgb1.Gradient(rgb2, k)
	}
}

func BenchmarkRGB32Add(b *testing.B) {
	rgb1 := NewRGB32[Linear](0.4, 0.4, 0.4)
	rgb2 := NewRGB32[Linear](0.2, 0.3, 0.1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb1.Add(rgb2)
	}
}

func BenchmarkRGB32Luminance(b *testing.B) {
	rgb := NewRGB32[Linear](0.5, 0.75, 0.25)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb.Luminance()
	}
}
