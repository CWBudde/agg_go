package pixfmt

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func TestGammaNone(t *testing.T) {
	gamma := NewGammaNone()

	// Test identity function
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0, 2.0}
	for _, x := range testCases {
		result := gamma.Apply(x)
		if result != x {
			t.Errorf("GammaNone.Apply(%f) = %f, expected %f", x, result, x)
		}
	}
}

func TestGammaPower(t *testing.T) {
	// Test default constructor (gamma = 1.0)
	gamma := &GammaPower{gamma: 1.0}
	if gamma.Gamma() != 1.0 {
		t.Errorf("Default gamma should be 1.0, got %f", gamma.Gamma())
	}

	// Test with gamma = 1.0 (should behave like identity)
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, x := range testCases {
		result := gamma.Apply(x)
		if math.Abs(result-x) > epsilon {
			t.Errorf("GammaPower(1.0).Apply(%f) = %f, expected %f", x, result, x)
		}
	}

	// Test with gamma = 2.0 (square function)
	gamma = NewGammaPower(2.0)
	if gamma.Gamma() != 2.0 {
		t.Errorf("Gamma should be 2.0, got %f", gamma.Gamma())
	}

	result := gamma.Apply(0.5)
	expected := 0.25
	if math.Abs(result-expected) > epsilon {
		t.Errorf("GammaPower(2.0).Apply(0.5) = %f, expected %f", result, expected)
	}

	// Test with gamma = 0.5 (square root function)
	gamma.SetGamma(0.5)
	result = gamma.Apply(0.25)
	expected = 0.5
	if math.Abs(result-expected) > epsilon {
		t.Errorf("GammaPower(0.5).Apply(0.25) = %f, expected %f", result, expected)
	}

	// Test edge cases
	result = gamma.Apply(0.0)
	if result != 0.0 {
		t.Errorf("GammaPower.Apply(0.0) = %f, expected 0.0", result)
	}

	result = gamma.Apply(1.0)
	if result != 1.0 {
		t.Errorf("GammaPower.Apply(1.0) = %f, expected 1.0", result)
	}
}

func TestGammaThreshold(t *testing.T) {
	// Test default threshold (0.5)
	gamma := &GammaThreshold{threshold: 0.5}
	if gamma.Threshold() != 0.5 {
		t.Errorf("Default threshold should be 0.5, got %f", gamma.Threshold())
	}

	// Test values below threshold
	result := gamma.Apply(0.3)
	if result != 0.0 {
		t.Errorf("GammaThreshold.Apply(0.3) = %f, expected 0.0", result)
	}

	result = gamma.Apply(0.49999)
	if result != 0.0 {
		t.Errorf("GammaThreshold.Apply(0.49999) = %f, expected 0.0", result)
	}

	// Test values at or above threshold
	result = gamma.Apply(0.5)
	if result != 1.0 {
		t.Errorf("GammaThreshold.Apply(0.5) = %f, expected 1.0", result)
	}

	result = gamma.Apply(0.7)
	if result != 1.0 {
		t.Errorf("GammaThreshold.Apply(0.7) = %f, expected 1.0", result)
	}

	result = gamma.Apply(1.0)
	if result != 1.0 {
		t.Errorf("GammaThreshold.Apply(1.0) = %f, expected 1.0", result)
	}

	// Test custom threshold
	gamma = NewGammaThreshold(0.75)
	gamma.SetThreshold(0.75)

	result = gamma.Apply(0.7)
	if result != 0.0 {
		t.Errorf("GammaThreshold(0.75).Apply(0.7) = %f, expected 0.0", result)
	}

	result = gamma.Apply(0.8)
	if result != 1.0 {
		t.Errorf("GammaThreshold(0.75).Apply(0.8) = %f, expected 1.0", result)
	}
}

func TestGammaLinear(t *testing.T) {
	// Test default range [0, 1]
	gamma := &GammaLinear{start: 0.0, end: 1.0}
	if gamma.Start() != 0.0 || gamma.End() != 1.0 {
		t.Errorf("Default range should be [0,1], got [%f,%f]", gamma.Start(), gamma.End())
	}

	// Test identity behavior with default range
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, x := range testCases {
		result := gamma.Apply(x)
		if math.Abs(result-x) > epsilon {
			t.Errorf("GammaLinear[0,1].Apply(%f) = %f, expected %f", x, result, x)
		}
	}

	// Test clamping below start
	result := gamma.Apply(-0.5)
	if result != 0.0 {
		t.Errorf("GammaLinear.Apply(-0.5) = %f, expected 0.0", result)
	}

	// Test clamping above end
	result = gamma.Apply(1.5)
	if result != 1.0 {
		t.Errorf("GammaLinear.Apply(1.5) = %f, expected 1.0", result)
	}

	// Test custom range [0.2, 0.8]
	gamma = NewGammaLinear(0.2, 0.8)
	gamma.Set(0.2, 0.8)

	result = gamma.Apply(0.1)
	if result != 0.0 {
		t.Errorf("GammaLinear[0.2,0.8].Apply(0.1) = %f, expected 0.0", result)
	}

	result = gamma.Apply(0.2)
	if result != 0.0 {
		t.Errorf("GammaLinear[0.2,0.8].Apply(0.2) = %f, expected 0.0", result)
	}

	result = gamma.Apply(0.5)
	expected := 0.5 // (0.5 - 0.2) / (0.8 - 0.2) = 0.3 / 0.6 = 0.5
	if math.Abs(result-expected) > epsilon {
		t.Errorf("GammaLinear[0.2,0.8].Apply(0.5) = %f, expected %f", result, expected)
	}

	result = gamma.Apply(0.8)
	if result != 1.0 {
		t.Errorf("GammaLinear[0.2,0.8].Apply(0.8) = %f, expected 1.0", result)
	}

	result = gamma.Apply(0.9)
	if result != 1.0 {
		t.Errorf("GammaLinear[0.2,0.8].Apply(0.9) = %f, expected 1.0", result)
	}

	// Test individual setters
	gamma.SetStart(0.1)
	gamma.SetEnd(0.9)
	if gamma.Start() != 0.1 || gamma.End() != 0.9 {
		t.Errorf("After SetStart/SetEnd, range should be [0.1,0.9], got [%f,%f]", gamma.Start(), gamma.End())
	}
}

func TestGammaMultiply(t *testing.T) {
	// Test default multiplier (1.0)
	gamma := &GammaMultiply{multiplier: 1.0}
	if gamma.Value() != 1.0 {
		t.Errorf("Default multiplier should be 1.0, got %f", gamma.Value())
	}

	// Test identity behavior with multiplier = 1.0
	testCases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, x := range testCases {
		result := gamma.Apply(x)
		if math.Abs(result-x) > epsilon {
			t.Errorf("GammaMultiply(1.0).Apply(%f) = %f, expected %f", x, result, x)
		}
	}

	// Test multiplication without clamping
	gamma = NewGammaMultiply(2.0)
	gamma.SetValue(2.0)
	if gamma.Value() != 2.0 {
		t.Errorf("Multiplier should be 2.0, got %f", gamma.Value())
	}

	result := gamma.Apply(0.25)
	expected := 0.5
	if math.Abs(result-expected) > epsilon {
		t.Errorf("GammaMultiply(2.0).Apply(0.25) = %f, expected %f", result, expected)
	}

	// Test clamping to 1.0
	result = gamma.Apply(0.75)
	expected = 1.0 // 0.75 * 2.0 = 1.5, clamped to 1.0
	if result != expected {
		t.Errorf("GammaMultiply(2.0).Apply(0.75) = %f, expected %f (clamped)", result, expected)
	}

	result = gamma.Apply(1.0)
	expected = 1.0
	if result != expected {
		t.Errorf("GammaMultiply(2.0).Apply(1.0) = %f, expected %f", result, expected)
	}

	// Test with multiplier < 1.0
	gamma.SetValue(0.5)
	result = gamma.Apply(0.8)
	expected = 0.4
	if math.Abs(result-expected) > epsilon {
		t.Errorf("GammaMultiply(0.5).Apply(0.8) = %f, expected %f", result, expected)
	}

	// Test edge case: zero multiplication
	result = gamma.Apply(0.0)
	if result != 0.0 {
		t.Errorf("GammaMultiply.Apply(0.0) = %f, expected 0.0", result)
	}
}

func TestSRGBToLinear(t *testing.T) {
	// Test the breakpoint at 0.04045
	result := SRGBToLinear(0.04045)
	expected := 0.04045 / 12.92
	if math.Abs(result-expected) > epsilon {
		t.Errorf("SRGBToLinear(0.04045) = %f, expected %f", result, expected)
	}

	// Test linear region (below threshold)
	result = SRGBToLinear(0.03)
	expected = 0.03 / 12.92
	if math.Abs(result-expected) > epsilon {
		t.Errorf("SRGBToLinear(0.03) = %f, expected %f", result, expected)
	}

	// Test power region (above threshold)
	result = SRGBToLinear(0.5)
	expected = math.Pow((0.5+0.055)/1.055, 2.4)
	if math.Abs(result-expected) > epsilon {
		t.Errorf("SRGBToLinear(0.5) = %f, expected %f", result, expected)
	}

	// Test edge cases
	result = SRGBToLinear(0.0)
	if result != 0.0 {
		t.Errorf("SRGBToLinear(0.0) = %f, expected 0.0", result)
	}

	result = SRGBToLinear(1.0)
	expected = math.Pow((1.0+0.055)/1.055, 2.4)
	if math.Abs(result-expected) > epsilon {
		t.Errorf("SRGBToLinear(1.0) = %f, expected %f", result, expected)
	}
}

func TestLinearToSRGB(t *testing.T) {
	// Test the breakpoint at 0.0031308
	result := LinearToSRGB(0.0031308)
	expected := 0.0031308 * 12.92
	if math.Abs(result-expected) > epsilon {
		t.Errorf("LinearToSRGB(0.0031308) = %f, expected %f", result, expected)
	}

	// Test linear region (below threshold)
	result = LinearToSRGB(0.002)
	expected = 0.002 * 12.92
	if math.Abs(result-expected) > epsilon {
		t.Errorf("LinearToSRGB(0.002) = %f, expected %f", result, expected)
	}

	// Test power region (above threshold)
	result = LinearToSRGB(0.5)
	expected = 1.055*math.Pow(0.5, 1.0/2.4) - 0.055
	if math.Abs(result-expected) > epsilon {
		t.Errorf("LinearToSRGB(0.5) = %f, expected %f", result, expected)
	}

	// Test edge cases
	result = LinearToSRGB(0.0)
	if result != 0.0 {
		t.Errorf("LinearToSRGB(0.0) = %f, expected 0.0", result)
	}

	result = LinearToSRGB(1.0)
	expected = 1.055*math.Pow(1.0, 1.0/2.4) - 0.055
	if math.Abs(result-expected) > epsilon {
		t.Errorf("LinearToSRGB(1.0) = %f, expected %f", result, expected)
	}
}

func TestSRGBRoundTrip(t *testing.T) {
	// Test that sRGB -> Linear -> sRGB is identity (within epsilon)
	// Note: avoiding exact boundary point 0.04045 due to floating-point precision
	testValues := []float64{0.0, 0.03, 0.041, 0.1, 0.5, 0.8, 1.0}

	for _, x := range testValues {
		linear := SRGBToLinear(x)
		srgb := LinearToSRGB(linear)
		if math.Abs(srgb-x) > epsilon {
			t.Errorf("sRGB round trip failed: %f -> %f -> %f", x, linear, srgb)
		}
	}
}

func TestLinearRoundTrip(t *testing.T) {
	// Test that Linear -> sRGB -> Linear is identity (within epsilon)
	testValues := []float64{0.0, 0.002, 0.0031308, 0.1, 0.5, 0.8, 1.0}

	for _, x := range testValues {
		srgb := LinearToSRGB(x)
		linear := SRGBToLinear(srgb)
		if math.Abs(linear-x) > epsilon {
			t.Errorf("Linear round trip failed: %f -> %f -> %f", x, srgb, linear)
		}
	}
}

// Test that all gamma functions implement the GammaFunction interface
func TestGammaFunctionInterface(t *testing.T) {
	var _ GammaFunction = &GammaNone{}
	var _ GammaFunction = &GammaPower{}
	var _ GammaFunction = &GammaThreshold{}
	var _ GammaFunction = &GammaLinear{}
	var _ GammaFunction = &GammaMultiply{}
}

// Benchmark tests
func BenchmarkGammaNone(b *testing.B) {
	gamma := NewGammaNone()
	for i := 0; i < b.N; i++ {
		gamma.Apply(0.5)
	}
}

func BenchmarkGammaPower(b *testing.B) {
	gamma := NewGammaPower(2.2)
	for i := 0; i < b.N; i++ {
		gamma.Apply(0.5)
	}
}

func BenchmarkSRGBToLinear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SRGBToLinear(0.5)
	}
}

func BenchmarkLinearToSRGB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LinearToSRGB(0.5)
	}
}
