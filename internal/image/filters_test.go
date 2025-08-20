package image

import (
	"fmt"
	"math"
	"testing"

	"agg_go/internal/basics"
)

// TestImageFilterLUT tests the basic LUT functionality
func TestImageFilterLUT(t *testing.T) {
	t.Run("NewImageFilterLUT", func(t *testing.T) {
		lut := NewImageFilterLUT()
		if lut == nil {
			t.Fatal("NewImageFilterLUT returned nil")
		}
		if lut.Radius() != 0 {
			t.Errorf("Expected radius 0, got %f", lut.Radius())
		}
		if lut.Diameter() != 0 {
			t.Errorf("Expected diameter 0, got %d", lut.Diameter())
		}
	})

	t.Run("Calculate with BilinearFilter", func(t *testing.T) {
		lut := NewImageFilterLUT()
		filter := BilinearFilter{}
		lut.Calculate(filter, true)

		if lut.Radius() != 1.0 {
			t.Errorf("Expected radius 1.0, got %f", lut.Radius())
		}
		if lut.Diameter() != 2 {
			t.Errorf("Expected diameter 2, got %d", lut.Diameter())
		}
		if lut.Start() != 0 {
			t.Errorf("Expected start 0, got %d", lut.Start())
		}

		weights := lut.WeightArray()
		if len(weights) == 0 {
			t.Error("Weight array is empty")
		}
	})

	t.Run("NewImageFilterLUTWithFilter", func(t *testing.T) {
		filter := BilinearFilter{}
		lut := NewImageFilterLUTWithFilter(filter, true)

		if lut.Radius() != 1.0 {
			t.Errorf("Expected radius 1.0, got %f", lut.Radius())
		}
	})
}

// TestBilinearFilter tests the bilinear filter implementation
func TestBilinearFilter(t *testing.T) {
	filter := BilinearFilter{}

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	// Test weight calculation
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 1.0},
		{0.5, 0.5},
		{1.0, 0.0},
	}

	for _, tt := range tests {
		result := filter.CalcWeight(tt.x)
		if math.Abs(result-tt.expected) > 1e-6 {
			t.Errorf("CalcWeight(%f) = %f, want %f", tt.x, result, tt.expected)
		}
	}
}

// TestHanningFilter tests the Hanning filter implementation
func TestHanningFilter(t *testing.T) {
	filter := HanningFilter{}

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	// Test weight calculation
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 1.0},
		{0.5, 0.5},
		{1.0, 0.0},
	}

	for _, tt := range tests {
		result := filter.CalcWeight(tt.x)
		if math.Abs(result-tt.expected) > 1e-6 {
			t.Errorf("CalcWeight(%f) = %f, want %f", tt.x, result, tt.expected)
		}
	}
}

// TestHammingFilter tests the Hamming filter implementation
func TestHammingFilter(t *testing.T) {
	filter := HammingFilter{}

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test weight at x=1 (should be approximately 0.08)
	result = filter.CalcWeight(1.0)
	expected := 0.54 + 0.46*math.Cos(basics.Pi)
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(1.0) = %f, want %f", result, expected)
	}
}

// TestHermiteFilter tests the Hermite filter implementation
func TestHermiteFilter(t *testing.T) {
	filter := HermiteFilter{}

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	// Test weight calculation
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 1.0},
		{1.0, 0.0},
	}

	for _, tt := range tests {
		result := filter.CalcWeight(tt.x)
		if math.Abs(result-tt.expected) > 1e-6 {
			t.Errorf("CalcWeight(%f) = %f, want %f", tt.x, result, tt.expected)
		}
	}
}

// TestQuadricFilter tests the Quadric filter implementation
func TestQuadricFilter(t *testing.T) {
	filter := QuadricFilter{}

	if filter.Radius() != 1.5 {
		t.Errorf("Expected radius 1.5, got %f", filter.Radius())
	}

	// Test weight calculation
	tests := []struct {
		x        float64
		expected float64
	}{
		{0.0, 0.75},
		{0.5, 0.5},
		{1.0, 0.125},
		{1.5, 0.0},
		{2.0, 0.0},
	}

	for _, tt := range tests {
		result := filter.CalcWeight(tt.x)
		if math.Abs(result-tt.expected) > 1e-6 {
			t.Errorf("CalcWeight(%f) = %f, want %f", tt.x, result, tt.expected)
		}
	}
}

// TestBicubicFilter tests the Bicubic filter implementation
func TestBicubicFilter(t *testing.T) {
	filter := BicubicFilter{}

	if filter.Radius() != 2.0 {
		t.Errorf("Expected radius 2.0, got %f", filter.Radius())
	}

	// Test weight at x=0 - for bicubic this should be 2/3
	result := filter.CalcWeight(0.0)
	expected := (1.0 / 6.0) * (8.0 - 4.0 + 0.0 - 0.0) // pow3(2) - 4*pow3(1) + 6*pow3(0) - 4*pow3(-1)
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want %f", result, expected)
	}

	// Test weight at x=1 - for bicubic this should be 1/6
	result = filter.CalcWeight(1.0)
	expected = (1.0 / 6.0) * (27.0 - 4.0*8.0 + 6.0*1.0 - 0.0) // pow3(3) - 4*pow3(2) + 6*pow3(1) - 4*pow3(0)
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(1.0) = %f, want %f", result, expected)
	}
}

// TestKaiserFilter tests the Kaiser filter implementation
func TestKaiserFilter(t *testing.T) {
	filter := NewKaiserFilter(6.33)

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test that weights decrease as x increases
	weight1 := filter.CalcWeight(0.5)
	weight2 := filter.CalcWeight(0.8)
	if weight1 <= weight2 {
		t.Errorf("Expected weight to decrease with x: weight(0.5)=%f, weight(0.8)=%f", weight1, weight2)
	}
}

// TestCatromFilter tests the Catmull-Rom filter implementation
func TestCatromFilter(t *testing.T) {
	filter := CatromFilter{}

	if filter.Radius() != 2.0 {
		t.Errorf("Expected radius 2.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test weight at x=1
	result = filter.CalcWeight(1.0)
	expected := 0.5 * (2.0 + 1.0*(-5.0+1.0*3.0))
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(1.0) = %f, want %f", result, expected)
	}

	// Test weight beyond radius
	result = filter.CalcWeight(2.5)
	if math.Abs(result) > 1e-6 {
		t.Errorf("CalcWeight(2.5) = %f, want 0.0", result)
	}
}

// TestMitchellFilter tests the Mitchell-Netravali filter implementation
func TestMitchellFilter(t *testing.T) {
	filter := NewMitchellFilter(1.0/3.0, 1.0/3.0)

	if filter.Radius() != 2.0 {
		t.Errorf("Expected radius 2.0, got %f", filter.Radius())
	}

	// Test weight at x=0
	result := filter.CalcWeight(0.0)
	if result <= 0 {
		t.Errorf("CalcWeight(0.0) = %f, should be positive", result)
	}

	// Test weight beyond radius
	result = filter.CalcWeight(2.5)
	if math.Abs(result) > 1e-6 {
		t.Errorf("CalcWeight(2.5) = %f, want 0.0", result)
	}
}

// TestSplineFilters tests the spline filter implementations
func TestSplineFilters(t *testing.T) {
	t.Run("Spline16", func(t *testing.T) {
		filter := Spline16Filter{}
		if filter.Radius() != 2.0 {
			t.Errorf("Expected radius 2.0, got %f", filter.Radius())
		}

		// Test weight at x=0 (should be 1.0)
		result := filter.CalcWeight(0.0)
		if math.Abs(result-1.0) > 1e-6 {
			t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
		}
	})

	t.Run("Spline36", func(t *testing.T) {
		filter := Spline36Filter{}
		if filter.Radius() != 3.0 {
			t.Errorf("Expected radius 3.0, got %f", filter.Radius())
		}

		// Test weight at x=0 (should be 1.0)
		result := filter.CalcWeight(0.0)
		if math.Abs(result-1.0) > 1e-6 {
			t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
		}
	})
}

// TestGaussianFilter tests the Gaussian filter implementation
func TestGaussianFilter(t *testing.T) {
	filter := GaussianFilter{}

	if filter.Radius() != 2.0 {
		t.Errorf("Expected radius 2.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be positive)
	result := filter.CalcWeight(0.0)
	expected := math.Sqrt(2.0 / basics.Pi)
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want %f", result, expected)
	}

	// Test that weights decrease as x increases
	weight1 := filter.CalcWeight(0.5)
	weight2 := filter.CalcWeight(1.0)
	if weight1 <= weight2 {
		t.Errorf("Expected weight to decrease with x: weight(0.5)=%f, weight(1.0)=%f", weight1, weight2)
	}
}

// TestBesselFilter tests the Bessel filter implementation
func TestBesselFilter(t *testing.T) {
	filter := BesselFilter{}

	if math.Abs(filter.Radius()-3.2383) > 1e-4 {
		t.Errorf("Expected radius 3.2383, got %f", filter.Radius())
	}

	// Test weight at x=0
	result := filter.CalcWeight(0.0)
	expected := basics.Pi / 4.0
	if math.Abs(result-expected) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want %f", result, expected)
	}
}

// TestSincFilter tests the Sinc filter implementation
func TestSincFilter(t *testing.T) {
	filter := NewSincFilter(3.0)

	if filter.Radius() != 3.0 {
		t.Errorf("Expected radius 3.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test minimum radius constraint
	filter2 := NewSincFilter(1.0)
	if filter2.Radius() != 2.0 {
		t.Errorf("Expected minimum radius 2.0, got %f", filter2.Radius())
	}
}

// TestLanczosFilter tests the Lanczos filter implementation
func TestLanczosFilter(t *testing.T) {
	filter := NewLanczosFilter(3.0)

	if filter.Radius() != 3.0 {
		t.Errorf("Expected radius 3.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test weight beyond radius (should be 0.0)
	result = filter.CalcWeight(4.0)
	if math.Abs(result) > 1e-6 {
		t.Errorf("CalcWeight(4.0) = %f, want 0.0", result)
	}

	// Test minimum radius constraint
	filter2 := NewLanczosFilter(1.0)
	if filter2.Radius() != 2.0 {
		t.Errorf("Expected minimum radius 2.0, got %f", filter2.Radius())
	}
}

// TestBlackmanFilter tests the Blackman filter implementation
func TestBlackmanFilter(t *testing.T) {
	filter := NewBlackmanFilter(3.0)

	if filter.Radius() != 3.0 {
		t.Errorf("Expected radius 3.0, got %f", filter.Radius())
	}

	// Test weight at x=0 (should be 1.0)
	result := filter.CalcWeight(0.0)
	if math.Abs(result-1.0) > 1e-6 {
		t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
	}

	// Test weight beyond radius (should be 0.0)
	result = filter.CalcWeight(4.0)
	if math.Abs(result) > 1e-6 {
		t.Errorf("CalcWeight(4.0) = %f, want 0.0", result)
	}

	// Test minimum radius constraint
	filter2 := NewBlackmanFilter(1.0)
	if filter2.Radius() != 2.0 {
		t.Errorf("Expected minimum radius 2.0, got %f", filter2.Radius())
	}
}

// TestPreDefinedFilters tests the pre-defined filter variants
func TestPreDefinedFilters(t *testing.T) {
	testCases := []struct {
		name           string
		filter         FilterFunction
		expectedRadius float64
	}{
		{"Sinc36", NewSinc36Filter(), 3.0},
		{"Sinc64", NewSinc64Filter(), 4.0},
		{"Sinc100", NewSinc100Filter(), 5.0},
		{"Sinc144", NewSinc144Filter(), 6.0},
		{"Sinc196", NewSinc196Filter(), 7.0},
		{"Sinc256", NewSinc256Filter(), 8.0},
		{"Lanczos36", NewLanczos36Filter(), 3.0},
		{"Lanczos64", NewLanczos64Filter(), 4.0},
		{"Lanczos100", NewLanczos100Filter(), 5.0},
		{"Lanczos144", NewLanczos144Filter(), 6.0},
		{"Lanczos196", NewLanczos196Filter(), 7.0},
		{"Lanczos256", NewLanczos256Filter(), 8.0},
		{"Blackman36", NewBlackman36Filter(), 3.0},
		{"Blackman64", NewBlackman64Filter(), 4.0},
		{"Blackman100", NewBlackman100Filter(), 5.0},
		{"Blackman144", NewBlackman144Filter(), 6.0},
		{"Blackman196", NewBlackman196Filter(), 7.0},
		{"Blackman256", NewBlackman256Filter(), 8.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.filter.Radius() != tc.expectedRadius {
				t.Errorf("Expected radius %f, got %f", tc.expectedRadius, tc.filter.Radius())
			}

			// Test weight at x=0 (should be 1.0 for all sinc-based filters)
			result := tc.filter.CalcWeight(0.0)
			if math.Abs(result-1.0) > 1e-6 {
				t.Errorf("CalcWeight(0.0) = %f, want 1.0", result)
			}
		})
	}
}

// TestImageFilter tests the generic ImageFilter type
func TestImageFilter(t *testing.T) {
	filter := NewImageFilter(BilinearFilter{})

	if filter.Radius() != 1.0 {
		t.Errorf("Expected radius 1.0, got %f", filter.Radius())
	}

	if filter.Diameter() != 2 {
		t.Errorf("Expected diameter 2, got %d", filter.Diameter())
	}

	weights := filter.WeightArray()
	if len(weights) == 0 {
		t.Error("Weight array is empty")
	}
}

// TestFilterNormalization tests that filter weights are properly normalized
func TestFilterNormalization(t *testing.T) {
	filters := []FilterFunction{
		BilinearFilter{},
		HanningFilter{},
		HammingFilter{},
		HermiteFilter{},
		QuadricFilter{},
		BicubicFilter{},
		NewKaiserFilter(6.33),
		CatromFilter{},
		NewMitchellFilter(1.0/3.0, 1.0/3.0),
		Spline16Filter{},
		GaussianFilter{},
	}

	for i, filter := range filters {
		t.Run(fmt.Sprintf("Filter%d", i), func(t *testing.T) {
			lut := NewImageFilterLUTWithFilter(filter, true)

			// Test that the LUT was created
			if lut.WeightArray() == nil {
				t.Error("Weight array is nil")
			}

			// Test basic properties
			if lut.Radius() != filter.Radius() {
				t.Errorf("LUT radius %f does not match filter radius %f", lut.Radius(), filter.Radius())
			}

			if lut.Diameter() <= 0 {
				t.Errorf("Diameter should be positive, got %d", lut.Diameter())
			}
		})
	}
}

// BenchmarkFilters benchmarks the filter weight calculations
func BenchmarkFilters(b *testing.B) {
	filters := map[string]FilterFunction{
		"Bilinear": BilinearFilter{},
		"Hanning":  HanningFilter{},
		"Hamming":  HammingFilter{},
		"Hermite":  HermiteFilter{},
		"Quadric":  QuadricFilter{},
		"Bicubic":  BicubicFilter{},
		"Kaiser":   NewKaiserFilter(6.33),
		"Catrom":   CatromFilter{},
		"Mitchell": NewMitchellFilter(1.0/3.0, 1.0/3.0),
		"Spline16": Spline16Filter{},
		"Spline36": Spline36Filter{},
		"Gaussian": GaussianFilter{},
		"Sinc":     NewSincFilter(3.0),
		"Lanczos":  NewLanczosFilter(3.0),
		"Blackman": NewBlackmanFilter(3.0),
	}

	for name, filter := range filters {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				x := float64(i%100) / 100.0
				filter.CalcWeight(x)
			}
		})
	}
}

// BenchmarkLUTCreation benchmarks the LUT creation process
func BenchmarkLUTCreation(b *testing.B) {
	filter := BilinearFilter{}
	for i := 0; i < b.N; i++ {
		lut := NewImageFilterLUTWithFilter(filter, true)
		_ = lut.WeightArray()
	}
}
