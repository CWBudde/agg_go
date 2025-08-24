package gamma

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestGammaLUTIdentity(t *testing.T) {
	lut := NewGammaLUT8()

	// Test identity mapping (gamma = 1.0)
	for i := 0; i < 256; i++ {
		input := basics.Int8u(i)
		output := lut.Dir(input)
		inverse := lut.Inv(output)

		// For identity, output should equal input
		if output != input {
			t.Errorf("Identity gamma failed: Dir(%d) = %d, expected %d", input, output, input)
		}

		// Inverse should also equal input
		if inverse != input {
			t.Errorf("Identity gamma inverse failed: Inv(%d) = %d, expected %d", output, inverse, input)
		}
	}
}

func TestGammaLUTBasicGamma(t *testing.T) {
	testCases := []struct {
		gamma     float64
		input     basics.Int8u
		expected  basics.Int8u
		tolerance basics.Int8u
	}{
		{2.2, 128, 56, 2},  // Mid-gray with gamma 2.2: (128/255)^2.2 * 255 ≈ 56
		{2.2, 64, 12, 2},   // Dark gray with gamma 2.2: (64/255)^2.2 * 255 ≈ 12
		{2.2, 192, 137, 2}, // Light gray with gamma 2.2: (192/255)^2.2 * 255 ≈ 137
		{0.5, 64, 128, 2},  // Inverse gamma effect: sqrt(64/255) * 255 ≈ 128
		{0.5, 128, 181, 2}, // Inverse gamma effect: sqrt(128/255) * 255 ≈ 181
	}

	for _, tc := range testCases {
		lut := NewGammaLUT8WithGamma(tc.gamma)
		result := lut.Dir(tc.input)

		diff := int(result) - int(tc.expected)
		if diff < 0 {
			diff = -diff
		}

		if basics.Int8u(diff) > tc.tolerance {
			t.Errorf("Gamma %.1f: Dir(%d) = %d, expected %d±%d",
				tc.gamma, tc.input, result, tc.expected, tc.tolerance)
		}
	}
}

func TestGammaLUTStandardValues(t *testing.T) {
	// Test standard sRGB gamma 2.2
	lut := NewGammaLUT8WithGamma(2.2)

	// Test boundary values
	if lut.Dir(0) != 0 {
		t.Errorf("Dir(0) should be 0, got %d", lut.Dir(0))
	}

	if lut.Dir(255) != 255 {
		t.Errorf("Dir(255) should be 255, got %d", lut.Dir(255))
	}

	if lut.Inv(0) != 0 {
		t.Errorf("Inv(0) should be 0, got %d", lut.Inv(0))
	}

	if lut.Inv(255) != 255 {
		t.Errorf("Inv(255) should be 255, got %d", lut.Inv(255))
	}
}

func TestGammaLUTRoundTrip(t *testing.T) {
	gammaValues := []float64{0.5, 1.0, 1.5, 2.0, 2.2, 2.4}

	for _, gamma := range gammaValues {
		lut := NewGammaLUT8WithGamma(gamma)

		errorCount := 0
		maxAllowedErrors := 50 // Allow some errors due to quantization

		for i := 16; i < 256; i++ { // Skip very small values that lose precision
			input := basics.Int8u(i)
			corrected := lut.Dir(input)
			recovered := lut.Inv(corrected)

			// Allow small rounding errors
			diff := int(recovered) - int(input)
			if diff < 0 {
				diff = -diff
			}

			tolerance := 3 // Allow ±3 for quantization errors
			if gamma == 1.0 {
				tolerance = 1 // Identity should be more precise
			}

			if diff > tolerance {
				errorCount++
				if errorCount <= 5 { // Report only first few failures
					t.Logf("Round-trip error for gamma %.1f: %d -> %d -> %d (diff: %d)",
						gamma, input, corrected, recovered, diff)
				}
			}
		}

		if errorCount > maxAllowedErrors {
			t.Errorf("Too many round-trip errors for gamma %.1f: %d errors (max allowed: %d)",
				gamma, errorCount, maxAllowedErrors)
		}
	}
}

func TestGammaFunctions(t *testing.T) {
	tests := []struct {
		name      string
		function  GammaFunction
		input     float64
		expected  float64
		tolerance float64
	}{
		{"GammaNone identity", NewGammaNone(), 0.5, 0.5, 0.001},
		{"GammaNone zero", NewGammaNone(), 0.0, 0.0, 0.001},
		{"GammaNone one", NewGammaNone(), 1.0, 1.0, 0.001},

		{"GammaPower 2.2", NewGammaPower(2.2), 0.5, math.Pow(0.5, 2.2), 0.001},
		{"GammaPower 1.0", NewGammaPower(1.0), 0.5, 0.5, 0.001},

		{"GammaThreshold below", NewGammaThreshold(0.5), 0.3, 0.0, 0.001},
		{"GammaThreshold above", NewGammaThreshold(0.5), 0.7, 1.0, 0.001},
		{"GammaThreshold exact", NewGammaThreshold(0.5), 0.5, 1.0, 0.001},

		{"GammaLinear below", NewGammaLinear(0.2, 0.8), 0.1, 0.0, 0.001},
		{"GammaLinear above", NewGammaLinear(0.2, 0.8), 0.9, 1.0, 0.001},
		{"GammaLinear middle", NewGammaLinear(0.2, 0.8), 0.5, 0.5, 0.001},

		{"GammaMultiply normal", NewGammaMultiply(1.5), 0.5, 0.75, 0.001},
		{"GammaMultiply clamp", NewGammaMultiply(1.5), 0.8, 1.0, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function.Apply(tt.input)
			diff := math.Abs(result - tt.expected)
			if diff > tt.tolerance {
				t.Errorf("Apply(%.3f) = %.6f, expected %.6f (diff: %.6f)",
					tt.input, result, tt.expected, diff)
			}
		})
	}
}

func TestSRGBLUTFloat(t *testing.T) {
	lut := NewSRGBLUTFloat()

	// Test known sRGB values
	testCases := []struct {
		srgb      basics.Int8u
		linear    float32
		tolerance float32
	}{
		{0, 0.0, 0.001},
		{255, 1.0, 0.001},
		{128, 0.2159, 0.002}, // Mid-gray in linear space (sRGB 128 ≈ 0.2159 linear)
		{64, 0.0513, 0.002},  // Dark gray
		{192, 0.5276, 0.002}, // Light gray
	}

	for _, tc := range testCases {
		result := lut.Dir(tc.srgb)
		diff := result - tc.linear
		if diff < 0 {
			diff = -diff
		}

		if diff > tc.tolerance {
			t.Errorf("sRGB to linear: Dir(%d) = %.4f, expected %.4f",
				tc.srgb, result, tc.linear)
		}

		// Test round-trip
		recovered := lut.Inv(result)
		if recovered != tc.srgb {
			// Allow ±1 for rounding in binary search
			srgbDiff := int(recovered) - int(tc.srgb)
			if srgbDiff < 0 {
				srgbDiff = -srgbDiff
			}
			if srgbDiff > 1 {
				t.Errorf("sRGB round-trip failed: %d -> %.4f -> %d",
					tc.srgb, result, recovered)
			}
		}
	}
}

func TestSRGBLUT16(t *testing.T) {
	lut := NewSRGBLUT16()

	// Test boundary values
	if lut.Dir(0) != 0 {
		t.Errorf("Dir(0) should be 0, got %d", lut.Dir(0))
	}

	if lut.Dir(255) != 65535 {
		t.Errorf("Dir(255) should be 65535, got %d", lut.Dir(255))
	}

	// Test mid-range value
	result := lut.Dir(128)
	expected := basics.Int16u(65535.0*SRGBToLinear(128.0/255.0) + 0.5)
	tolerance := basics.Int16u(100) // Allow some tolerance for rounding

	diff := int(result) - int(expected)
	if diff < 0 {
		diff = -diff
	}

	if basics.Int16u(diff) > tolerance {
		t.Errorf("sRGB16 Dir(128) = %d, expected %d±%d", result, expected, tolerance)
	}
}

func TestSRGBLUT8(t *testing.T) {
	lut := NewSRGBLUT8()

	errorCount := 0
	maxAllowedErrors := 80 // Allow errors due to sRGB quantization

	// Test round-trip for 8-bit sRGB
	for i := 0; i < 256; i++ {
		input := basics.Int8u(i)
		linear := lut.Dir(input)
		recovered := lut.Inv(linear)

		// Allow small differences due to sRGB curve quantization
		diff := int(input) - int(recovered)
		if diff < 0 {
			diff = -diff
		}

		if diff > 3 { // Allow ±3 for sRGB quantization errors
			errorCount++
			if errorCount <= 5 {
				t.Logf("sRGB8 round-trip error: %d -> %d -> %d (diff: %d)", input, linear, recovered, diff)
			}
		}
	}

	if errorCount > maxAllowedErrors {
		t.Errorf("Too many sRGB8 round-trip errors: %d (max allowed: %d)", errorCount, maxAllowedErrors)
	}
}

func TestSRGBConversionFunctions(t *testing.T) {
	// Test individual sRGB conversion functions
	testCases := []struct {
		srgb      float64
		linear    float64
		tolerance float64
	}{
		{0.0, 0.0, 0.001},
		{1.0, 1.0, 0.001},
		{0.5, 0.2140, 0.001},
		{0.04045, 0.04045 / 12.92, 0.001}, // Breakpoint
	}

	for _, tc := range testCases {
		// Test sRGB to linear
		result := SRGBToLinear(tc.srgb)
		diff := math.Abs(result - tc.linear)
		if diff > tc.tolerance {
			t.Errorf("SRGBToLinear(%.4f) = %.6f, expected %.6f",
				tc.srgb, result, tc.linear)
		}

		// Test linear to sRGB
		result = LinearToSRGB(tc.linear)
		diff = math.Abs(result - tc.srgb)
		if diff > tc.tolerance {
			t.Errorf("LinearToSRGB(%.6f) = %.4f, expected %.4f",
				tc.linear, result, tc.srgb)
		}
	}
}

func TestSRGBConvTypes(t *testing.T) {
	// Test SRGBConvFloat
	convFloat := SRGBConvFloat{}

	linear := convFloat.RGBFromSRGB(128)
	expected := float32(SRGBToLinear(128.0 / 255.0))
	tolerance := float32(0.001)

	diff := linear - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("SRGBConvFloat.RGBFromSRGB(128) = %.6f, expected %.6f",
			linear, expected)
	}

	// Test round-trip
	recovered := convFloat.RGBToSRGB(linear)
	if recovered != 128 {
		// Allow ±1 for rounding
		diff := int(recovered) - 128
		if diff < 0 {
			diff = -diff
		}
		if diff > 1 {
			t.Errorf("SRGBConvFloat round-trip failed: 128 -> %.6f -> %d",
				linear, recovered)
		}
	}

	// Test alpha conversion (should be identity for linear space)
	alpha := convFloat.AlphaFromSRGB(200)
	expectedAlpha := float32(200) / 255.0
	if math.Abs(float64(alpha-expectedAlpha)) > 0.001 {
		t.Errorf("AlphaFromSRGB(200) = %.6f, expected %.6f", alpha, expectedAlpha)
	}
}

// Benchmark tests
func BenchmarkGammaLUT8Dir(b *testing.B) {
	lut := NewGammaLUT8WithGamma(2.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lut.Dir(basics.Int8u(i & 255))
	}
}

func BenchmarkGammaLUT8Inv(b *testing.B) {
	lut := NewGammaLUT8WithGamma(2.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lut.Inv(basics.Int8u(i & 255))
	}
}

func BenchmarkSRGBLUTFloat(b *testing.B) {
	lut := NewSRGBLUTFloat()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lut.Dir(basics.Int8u(i & 255))
	}
}

func BenchmarkSRGBLUTFloatInv(b *testing.B) {
	lut := NewSRGBLUTFloat()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := float32(i&255) / 255.0
		_ = lut.Inv(v)
	}
}

func BenchmarkSRGBConversionDirect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := float64(i&255) / 255.0
		_ = SRGBToLinear(v)
	}
}

// Test memory usage and performance for larger bit depths
func TestGammaLUT16Bit(t *testing.T) {
	lut := NewGammaLUTWithShifts[basics.Int8u, basics.Int16u](8, 16)
	lut.SetGamma(2.2)

	// Test some known values
	result := lut.Dir(128)
	if result == 0 {
		t.Error("16-bit gamma LUT should produce non-zero output for input 128")
	}

	// Test round-trip within tolerance
	recovered := lut.Inv(result)
	diff := int(recovered) - 128
	if diff < 0 {
		diff = -diff
	}
	if diff > 1 {
		t.Errorf("16-bit gamma round-trip failed: 128 -> %d -> %d", result, recovered)
	}
}

// Test the AGG bridge compatibility
func TestAGGGammaLUTBridge(t *testing.T) {
	agg := NewAGGGammaLUT(2.2)

	// Test basic functionality
	if agg.Gamma() != 2.2 {
		t.Errorf("Expected gamma 2.2, got %.2f", agg.Gamma())
	}

	// Test gamma correction
	input := basics.Int8u(128)
	output := agg.Dir(input)
	expected := basics.Int8u(56) // Known value for gamma 2.2
	if output != expected {
		t.Errorf("AGG bridge Dir(%d) = %d, expected %d", input, output, expected)
	}

	// Test round-trip
	recovered := agg.Inv(output)
	if recovered != input {
		t.Errorf("AGG bridge round-trip failed: %d -> %d -> %d", input, output, recovered)
	}

	// Test gamma change
	agg.SetGamma(1.0)
	if agg.Gamma() != 1.0 {
		t.Errorf("Expected gamma 1.0 after SetGamma, got %.2f", agg.Gamma())
	}

	// Identity test
	identityOutput := agg.Dir(input)
	if identityOutput != input {
		t.Errorf("Identity gamma failed: Dir(%d) = %d, expected %d", input, identityOutput, input)
	}
}
