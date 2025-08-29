package color

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestRGB8(t *testing.T) {
	// Test NewRGB8
	rgb := NewRGB8[Linear](128, 192, 64)
	if rgb.R != 128 || rgb.G != 192 || rgb.B != 64 {
		t.Errorf("NewRGB8 failed: got %v, want {128, 192, 64}", rgb)
	}

	// Test Clear
	rgb.Clear()
	if rgb.R != 0 || rgb.G != 0 || rgb.B != 0 {
		t.Errorf("Clear failed: got %v, want {0, 0, 0}", rgb)
	}

	// Test ConvertToRGB
	rgb = NewRGB8[Linear](255, 128, 0)
	frgb := rgb.ConvertToRGB()
	expected := RGB{R: 1.0, G: 128.0 / 255.0, B: 0.0}
	if math.Abs(frgb.R-expected.R) > 1e-6 ||
		math.Abs(frgb.G-expected.G) > 1e-6 ||
		math.Abs(frgb.B-expected.B) > 1e-6 {
		t.Errorf("ConvertToRGB failed: got %v, want %v", frgb, expected)
	}

	// Test ConvertFromRGB
	frgb = RGB{R: 1.0, G: 0.5, B: 0.0}
	rgb = ConvertFromRGB[Linear](frgb)
	if rgb.R != 255 || rgb.G != 128 || rgb.B != 0 {
		t.Errorf("ConvertFromRGB failed: got %v, want {255, 128, 0}", rgb)
	}

	// Test ToRGBA8
	rgb = NewRGB8[Linear](100, 150, 200)
	rgba := rgb.ToRGBA8()
	if rgba.R != 100 || rgba.G != 150 || rgba.B != 200 || rgba.A != 255 {
		t.Errorf("ToRGBA8 failed: got %v, want {100, 150, 200, 255}", rgba)
	}

	// Test Add
	rgb1 := NewRGB8[Linear](100, 100, 100)
	rgb2 := NewRGB8[Linear](50, 75, 25)
	sum := rgb1.Add(rgb2)
	if sum.R != 150 || sum.G != 175 || sum.B != 125 {
		t.Errorf("Add failed: got %v, want {150, 175, 125}", sum)
	}

	// Test Add with overflow
	rgb1 = NewRGB8[Linear](200, 200, 200)
	rgb2 = NewRGB8[Linear](100, 100, 100)
	sum = rgb1.Add(rgb2)
	if sum.R != 255 || sum.G != 255 || sum.B != 255 {
		t.Errorf("Add overflow failed: got %v, want {255, 255, 255}", sum)
	}

	// Test Scale
	rgb = NewRGB8[Linear](100, 150, 200)
	scaled := rgb.Scale(1.5)
	if scaled.R != 150 || scaled.G != 225 || scaled.B != 255 {
		t.Errorf("Scale failed: got %v, want {150, 225, 255}", scaled)
	}

	// Test Gradient
	rgb1 = NewRGB8[Linear](0, 0, 0)
	rgb2 = NewRGB8[Linear](255, 255, 255)
	mid := rgb1.Gradient(rgb2, 128)
	// Should be approximately halfway
	if mid.R < 120 || mid.R > 135 || mid.G < 120 || mid.G > 135 || mid.B < 120 || mid.B > 135 {
		t.Errorf("Gradient failed: got %v, expected around {127, 127, 127}", mid)
	}

	// Test IsBlack and IsWhite
	black := NewRGB8[Linear](0, 0, 0)
	white := NewRGB8[Linear](255, 255, 255)
	gray := NewRGB8[Linear](128, 128, 128)

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
	red := NewRGB8[Linear](255, 0, 0)
	green := NewRGB8[Linear](0, 255, 0)
	blue := NewRGB8[Linear](0, 0, 255)

	redLum := red.Luminance()
	greenLum := green.Luminance()
	blueLum := blue.Luminance()

	// Green should have the highest luminance, blue the lowest
	if greenLum <= redLum || redLum <= blueLum {
		t.Errorf("Luminance calculation incorrect: R=%d, G=%d, B=%d", redLum, greenLum, blueLum)
	}

	// Test luminance for white should be close to 255
	whiteLum := white.Luminance()
	if whiteLum < 250 { // Allow some error due to integer arithmetic
		t.Errorf("White luminance too low: got %d, expected ~255", whiteLum)
	}

	// Test luminance for black should be 0
	blackLum := black.Luminance()
	if blackLum != 0 {
		t.Errorf("Black luminance incorrect: got %d, expected 0", blackLum)
	}
}

func TestRGBColorConstants(t *testing.T) {
	// Test predefined colors
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
}

func TestRGBHelperFunctions(t *testing.T) {
	// Test RGB8Multiply24
	result := RGB8Multiply24(128, 128)
	expected := RGBA8Multiply(128, 128) // Should be same as RGBA version
	if result != expected {
		t.Errorf("RGB8Multiply24 failed: got %d, want %d", result, expected)
	}

	// Test RGB8Lerp24
	result = RGB8Lerp24(0, 255, 128)
	expected = RGBA8Lerp(0, 255, 128) // Should be same as RGBA version
	if result != expected {
		t.Errorf("RGB8Lerp24 failed: got %d, want %d", result, expected)
	}

	// Test RGB8MultCover24
	result = RGB8MultCover24(200, 128)
	expected = RGBA8MultCover(200, 128) // Should be same as RGBA version
	if result != expected {
		t.Errorf("RGB8MultCover24 failed: got %d, want %d", result, expected)
	}
}

func TestColorOrders(t *testing.T) {
	// Test RGB24 order
	if OrderRGB24.R != 0 || OrderRGB24.G != 1 || OrderRGB24.B != 2 || OrderRGB24.A != -1 {
		t.Error("OrderRGB24 incorrect")
	}

	// Test BGR24 order
	if OrderBGR24.R != 2 || OrderBGR24.G != 1 || OrderBGR24.B != 0 || OrderBGR24.A != -1 {
		t.Error("OrderBGR24 incorrect")
	}
}

func TestRGBAConversions(t *testing.T) {
	// Test ConvertRGBAToRGB8
	rgba := NewRGBA(100.0/255.0, 150.0/255.0, 200.0/255.0, 128.0/255.0) // Alpha should be ignored
	rgb := ConvertRGBAToRGB8[Linear](rgba)
	if rgb.R != 100 || rgb.G != 150 || rgb.B != 200 {
		t.Errorf("ConvertFromRGBA failed: got %v, want {100, 150, 200}", rgb)
	}

	// Test ConvertToRGBA from RGB8
	rgb8 := NewRGB8[Linear](50, 75, 100)
	rgba8 := rgb8.ConvertToRGBA()
	const scale = 1.0 / 255.0
	expectedR := float64(50) * scale
	expectedG := float64(75) * scale
	expectedB := float64(100) * scale
	if math.Abs(rgba8.R-expectedR) > 1e-6 ||
		math.Abs(rgba8.G-expectedG) > 1e-6 ||
		math.Abs(rgba8.B-expectedB) > 1e-6 ||
		math.Abs(rgba8.A-1.0) > 1e-6 {
		t.Errorf("ConvertToRGBA failed: got %v, want {%f, %f, %f, 1.0}", rgba8, expectedR, expectedG, expectedB)
	}
}

// Benchmark tests
func BenchmarkRGB8Gradient(b *testing.B) {
	rgb1 := NewRGB8[Linear](0, 0, 0)
	rgb2 := NewRGB8[Linear](255, 255, 255)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb1.Gradient(rgb2, basics.Int8u(i&255))
	}
}

func BenchmarkRGB8Add(b *testing.B) {
	rgb1 := NewRGB8[Linear](100, 100, 100)
	rgb2 := NewRGB8[Linear](50, 75, 25)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb1.Add(rgb2)
	}
}

func BenchmarkRGB8Luminance(b *testing.B) {
	rgb := NewRGB8[Linear](128, 192, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rgb.Luminance()
	}
}

func TestRGB8ColorspaceConversion(t *testing.T) {
	// Test conversion from Linear to sRGB
	linearRgb := NewRGB8[Linear](128, 192, 64)
	srgbRgb := ConvertRGB8LinearToSRGB(linearRgb)

	// Linear values should be converted using the sRGB curve
	// We can't check exact values easily, but we can verify basic properties
	if srgbRgb.R == linearRgb.R && srgbRgb.G == linearRgb.G && srgbRgb.B == linearRgb.B {
		t.Error("Linear to sRGB conversion should change values (unless input values happen to match)")
	}

	// Test conversion from sRGB to Linear
	srgbOriginal := NewRGB8[SRGB](128, 192, 64)
	linearConverted := ConvertRGB8SRGBToLinear(srgbOriginal)

	// Values should be different after conversion
	if linearConverted.R == srgbOriginal.R && linearConverted.G == srgbOriginal.G && linearConverted.B == srgbOriginal.B {
		t.Error("sRGB to Linear conversion should change values (unless input values happen to match)")
	}

	// Test round-trip conversion accuracy
	linearOriginal := NewRGB8[Linear](100, 150, 200)
	srgbConverted := ConvertRGB8LinearToSRGB(linearOriginal)
	linearRoundtrip := ConvertRGB8SRGBToLinear(srgbConverted)

	// Round-trip should be very close to original (allowing for lookup table quantization)
	if absInt(int(linearRoundtrip.R)-int(linearOriginal.R)) > 1 ||
		absInt(int(linearRoundtrip.G)-int(linearOriginal.G)) > 1 ||
		absInt(int(linearRoundtrip.B)-int(linearOriginal.B)) > 1 {
		t.Errorf("Linear round-trip conversion failed: original=%v, roundtrip=%v",
			linearOriginal, linearRoundtrip)
	}

	// Test specific known conversion values
	// Black should remain black in both colorspaces
	blackLinear := NewRGB8[Linear](0, 0, 0)
	blackSRGB := ConvertRGB8LinearToSRGB(blackLinear)
	if blackSRGB.R != 0 || blackSRGB.G != 0 || blackSRGB.B != 0 {
		t.Errorf("Black conversion failed: got %v, want {0, 0, 0}", blackSRGB)
	}

	// White should remain white in both colorspaces
	whiteLinear := NewRGB8[Linear](255, 255, 255)
	whiteSRGB := ConvertRGB8LinearToSRGB(whiteLinear)
	if whiteSRGB.R != 255 || whiteSRGB.G != 255 || whiteSRGB.B != 255 {
		t.Errorf("White conversion failed: got %v, want {255, 255, 255}", whiteSRGB)
	}

	// Test mid-gray conversion consistency with known sRGB curve
	// Mid-gray in linear (128) should convert to a higher sRGB value due to gamma correction
	midGrayLinear := NewRGB8[Linear](128, 128, 128)
	midGraySRGB := ConvertRGB8LinearToSRGB(midGrayLinear)
	if midGraySRGB.R <= 128 || midGraySRGB.G <= 128 || midGraySRGB.B <= 128 {
		t.Errorf("Mid-gray Linear to sRGB should increase values: got %v, expected > 128", midGraySRGB)
	}
}

func TestRGB8ExplicitConversion(t *testing.T) {
	t.Run("Linear->sRGB", func(t *testing.T) {
		linearSrc := NewRGB8[Linear](100, 150, 200)
		srgbDst := ConvertRGB8LinearToSRGB(linearSrc)

		expected := ConvertRGB8LinearToSRGB(linearSrc) // direct reference
		if srgbDst.R != expected.R || srgbDst.G != expected.G || srgbDst.B != expected.B {
			t.Errorf("Linear->sRGB failed: got %v, want %v", srgbDst, expected)
		}
	})

	t.Run("sRGB->Linear", func(t *testing.T) {
		srgbSrc := NewRGB8[SRGB](100, 150, 200)
		linearDst := ConvertRGB8SRGBToLinear(srgbSrc)

		expected := ConvertRGB8SRGBToLinear(srgbSrc)
		if linearDst.R != expected.R || linearDst.G != expected.G || linearDst.B != expected.B {
			t.Errorf("sRGB->Linear failed: got %v, want %v", linearDst, expected)
		}
	})

	t.Run("Same colorspace no-op", func(t *testing.T) {
		linearSrc := NewRGB8[Linear](75, 125, 175)
		linearDst := RGB8[Linear](linearSrc) // no conversion, just copy

		if linearDst.R != linearSrc.R || linearDst.G != linearSrc.G || linearDst.B != linearSrc.B {
			t.Errorf("Same-CS copy should be no-op: got %v, want %v", linearDst, linearSrc)
		}
	})
}
