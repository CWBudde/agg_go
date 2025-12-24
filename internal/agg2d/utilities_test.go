package agg2d

import (
	"math"
	"testing"

	"agg_go/internal/transform"
)

func TestMathematicalConstants(t *testing.T) {
	if Pi != math.Pi {
		t.Errorf("Pi = %v, want %v", Pi, math.Pi)
	}
}

func TestDegRadConversions(t *testing.T) {
	tests := []struct {
		degrees float64
		radians float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{270, 3 * math.Pi / 2},
		{360, 2 * math.Pi},
		{-90, -math.Pi / 2},
		{45, math.Pi / 4},
	}

	for _, test := range tests {
		t.Run("Deg2Rad", func(t *testing.T) {
			result := Deg2Rad(test.degrees)
			if math.Abs(result-test.radians) > 1e-10 {
				t.Errorf("Deg2Rad(%v) = %v, want %v", test.degrees, result, test.radians)
			}
		})

		t.Run("Rad2Deg", func(t *testing.T) {
			result := Rad2Deg(test.radians)
			if math.Abs(result-test.degrees) > 1e-10 {
				t.Errorf("Rad2Deg(%v) = %v, want %v", test.radians, result, test.degrees)
			}
		})
	}
}

func TestAlignPoint(t *testing.T) {
	agg2d := NewAgg2D()

	tests := []struct {
		x, y                 float64
		expectedX, expectedY float64
	}{
		{1.2, 3.7, 1.5, 3.5},
		{0.0, 0.0, 0.5, 0.5},
		{-1.8, -2.3, -1.5, -2.5},
		{5.5, 7.5, 5.5, 7.5},
	}

	for _, test := range tests {
		t.Run("AlignPoint", func(t *testing.T) {
			x, y := test.x, test.y
			agg2d.AlignPoint(&x, &y)

			if math.Abs(x-test.expectedX) > 1e-10 {
				t.Errorf("AlignPoint x: got %v, want %v", x, test.expectedX)
			}
			if math.Abs(y-test.expectedY) > 1e-10 {
				t.Errorf("AlignPoint y: got %v, want %v", y, test.expectedY)
			}
		})
	}

	// Test with nil pointers
	agg2d.AlignPoint(nil, nil) // Should not panic
}

func TestInBox(t *testing.T) {
	agg2d := NewAgg2D()
	agg2d.transform = transform.NewTransAffine() // Identity transform
	agg2d.clipBox = struct{ X1, Y1, X2, Y2 float64 }{0, 0, 100, 100}

	tests := []struct {
		x, y     float64
		expected bool
	}{
		{50, 50, true},   // Inside
		{0, 0, true},     // On boundary
		{100, 100, true}, // On boundary
		{-10, 50, false}, // Outside left
		{110, 50, false}, // Outside right
		{50, -10, false}, // Outside top
		{50, 110, false}, // Outside bottom
	}

	for _, test := range tests {
		t.Run("InBox", func(t *testing.T) {
			result := agg2d.InBox(test.x, test.y)
			if result != test.expected {
				t.Errorf("InBox(%v, %v) = %v, want %v", test.x, test.y, result, test.expected)
			}
		})
	}
}

func TestWorldScreenScalarConversion(t *testing.T) {
	agg2d := NewAgg2D()
	agg2d.transform = transform.NewTransAffine()
	agg2d.transform.ScaleXY(2.0, 2.0) // 2x scale

	// Test world to screen
	worldScalar := 10.0
	screenScalar := agg2d.WorldToScreenScalar(worldScalar)
	expectedScreen := 20.0 // 10 * 2

	if math.Abs(screenScalar-expectedScreen) > 1e-10 {
		t.Errorf("WorldToScreenScalar(%v) = %v, want %v", worldScalar, screenScalar, expectedScreen)
	}

	// Test screen to world
	backToWorld := agg2d.ScreenToWorldScalar(screenScalar)
	if math.Abs(backToWorld-worldScalar) > 1e-10 {
		t.Errorf("ScreenToWorldScalar(%v) = %v, want %v", screenScalar, backToWorld, worldScalar)
	}
}

func TestNoFillNoLine(t *testing.T) {
	agg2d := NewAgg2D()

	// Set initial colors
	agg2d.FillColor(Red)
	agg2d.LineColor(Blue)

	// Test NoFill
	agg2d.NoFill()
	fillColor := agg2d.GetFillColor()
	expectedTransparent := Color{0, 0, 0, 0}
	if fillColor != expectedTransparent {
		t.Errorf("NoFill: expected %v, got %v", expectedTransparent, fillColor)
	}
	if agg2d.FillGradientFlag() != Solid {
		t.Errorf("NoFill: expected Solid gradient flag")
	}

	// Test NoLine
	agg2d.NoLine()
	lineColor := agg2d.GetLineColor()
	if lineColor != expectedTransparent {
		t.Errorf("NoLine: expected %v, got %v", expectedTransparent, lineColor)
	}
	if agg2d.LineGradientFlag() != Solid {
		t.Errorf("NoLine: expected Solid gradient flag")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{5.0, 0.0, 10.0, 5.0},   // Within range
		{-5.0, 0.0, 10.0, 0.0},  // Below min
		{15.0, 0.0, 10.0, 10.0}, // Above max
		{0.0, 0.0, 10.0, 0.0},   // At min
		{10.0, 0.0, 10.0, 10.0}, // At max
	}

	for _, test := range tests {
		t.Run("Clamp", func(t *testing.T) {
			result := Clamp(test.value, test.min, test.max)
			if result != test.expected {
				t.Errorf("Clamp(%v, %v, %v) = %v, want %v",
					test.value, test.min, test.max, result, test.expected)
			}
		})
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		value    float64
		expected uint8
	}{
		{128.0, 128},
		{-10.0, 0},
		{300.0, 255},
		{0.0, 0},
		{255.0, 255},
		{127.5, 127},
	}

	for _, test := range tests {
		t.Run("ClampByte", func(t *testing.T) {
			result := ClampByte(test.value)
			if result != test.expected {
				t.Errorf("ClampByte(%v) = %v, want %v", test.value, result, test.expected)
			}
		})
	}
}

func TestLinearInterpolate(t *testing.T) {
	tests := []struct {
		a, b, t, expected float64
	}{
		{0.0, 10.0, 0.5, 5.0},
		{0.0, 10.0, 0.0, 0.0},
		{0.0, 10.0, 1.0, 10.0},
		{10.0, 20.0, 0.25, 12.5},
	}

	for _, test := range tests {
		t.Run("LinearInterpolate", func(t *testing.T) {
			result := LinearInterpolate(test.a, test.b, test.t)
			if math.Abs(result-test.expected) > 1e-10 {
				t.Errorf("LinearInterpolate(%v, %v, %v) = %v, want %v",
					test.a, test.b, test.t, result, test.expected)
			}
		})
	}
}

func TestColorInterpolate(t *testing.T) {
	c1 := Color{0, 0, 0, 255}
	c2 := Color{255, 255, 255, 255}

	// Test midpoint interpolation
	result := ColorInterpolate(c1, c2, 0.5)
	expected := Color{127, 127, 127, 255}
	if result != expected {
		t.Errorf("ColorInterpolate midpoint: expected %v, got %v", expected, result)
	}

	// Test boundary conditions
	result = ColorInterpolate(c1, c2, 0.0)
	if result != c1 {
		t.Errorf("ColorInterpolate t=0: expected %v, got %v", c1, result)
	}

	result = ColorInterpolate(c1, c2, 1.0)
	if result != c2 {
		t.Errorf("ColorInterpolate t=1: expected %v, got %v", c2, result)
	}

	// Test clamping
	result = ColorInterpolate(c1, c2, -0.5)
	if result != c1 {
		t.Errorf("ColorInterpolate t<0: expected %v, got %v", c1, result)
	}

	result = ColorInterpolate(c1, c2, 1.5)
	if result != c2 {
		t.Errorf("ColorInterpolate t>1: expected %v, got %v", c2, result)
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		x1, y1, x2, y2, expected float64
	}{
		{0, 0, 3, 4, 5.0},   // 3-4-5 triangle
		{0, 0, 0, 0, 0.0},   // Same point
		{1, 1, 4, 5, 5.0},   // Another 3-4-5
		{-1, -1, 2, 3, 5.0}, // Negative coordinates
	}

	for _, test := range tests {
		t.Run("Distance", func(t *testing.T) {
			result := Distance(test.x1, test.y1, test.x2, test.y2)
			if math.Abs(result-test.expected) > 1e-10 {
				t.Errorf("Distance(%v, %v, %v, %v) = %v, want %v",
					test.x1, test.y1, test.x2, test.y2, result, test.expected)
			}
		})
	}
}

func TestAngle(t *testing.T) {
	tests := []struct {
		x1, y1, x2, y2, expected float64
	}{
		{0, 0, 1, 0, 0.0},           // East
		{0, 0, 0, 1, math.Pi / 2},   // North
		{0, 0, -1, 0, math.Pi},      // West
		{0, 0, 0, -1, -math.Pi / 2}, // South
		{0, 0, 1, 1, math.Pi / 4},   // Northeast
	}

	for _, test := range tests {
		t.Run("Angle", func(t *testing.T) {
			result := Angle(test.x1, test.y1, test.x2, test.y2)
			if math.Abs(result-test.expected) > 1e-10 {
				t.Errorf("Angle(%v, %v, %v, %v) = %v, want %v",
					test.x1, test.y1, test.x2, test.y2, result, test.expected)
			}
		})
	}
}

func TestNormalizeAngle(t *testing.T) {
	tests := []struct {
		angle, expected float64
	}{
		{0.0, 0.0},
		{math.Pi, math.Pi},
		{2 * math.Pi, 0.0},
		{3 * math.Pi, math.Pi},
		{-math.Pi, math.Pi},
		{-2 * math.Pi, 0.0},
	}

	for _, test := range tests {
		t.Run("NormalizeAngle", func(t *testing.T) {
			result := NormalizeAngle(test.angle)
			if math.Abs(result-test.expected) > 1e-10 {
				t.Errorf("NormalizeAngle(%v) = %v, want %v", test.angle, result, test.expected)
			}
		})
	}
}

func TestFloatingPointComparison(t *testing.T) {
	// Test IsZero
	if !IsZero(0.0) {
		t.Errorf("IsZero(0.0) should be true")
	}
	if !IsZero(1e-15) {
		t.Errorf("IsZero(1e-15) should be true")
	}
	if IsZero(1e-5) {
		t.Errorf("IsZero(1e-5) should be false")
	}

	// Test IsEqual
	if !IsEqual(1.0, 1.0) {
		t.Errorf("IsEqual(1.0, 1.0) should be true")
	}
	if !IsEqual(1.0, 1.0+1e-15) {
		t.Errorf("IsEqual(1.0, 1.0+1e-15) should be true")
	}
	if IsEqual(1.0, 1.1) {
		t.Errorf("IsEqual(1.0, 1.1) should be false")
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		value    float64
		expected int
	}{
		{5.0, 1},
		{-3.0, -1},
		{0.0, 0},
		{1e-15, 1}, // Very small positive is still positive
		{-1e-15, -1}, // Very small negative is still negative
	}

	for _, test := range tests {
		t.Run("Sign", func(t *testing.T) {
			result := Sign(test.value)
			if result != test.expected {
				t.Errorf("Sign(%v) = %v, want %v", test.value, result, test.expected)
			}
		})
	}
}

// Benchmarks
func BenchmarkDeg2Rad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Deg2Rad(90.0)
	}
}

func BenchmarkDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Distance(0, 0, 3, 4)
	}
}

func BenchmarkColorInterpolate(b *testing.B) {
	c1 := Color{0, 0, 0, 255}
	c2 := Color{255, 255, 255, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ColorInterpolate(c1, c2, 0.5)
	}
}
