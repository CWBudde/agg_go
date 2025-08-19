package span

import (
	"math"
	"testing"

	"agg_go/internal/color"
	"agg_go/internal/transform"
)

// Test gradient shape functions
func TestGradientShapeFunctions(t *testing.T) {
	t.Run("GradientLinearX", func(t *testing.T) {
		g := GradientLinearX{}

		// Test that it returns X coordinate
		if g.Calculate(10, 20, 100) != 10 {
			t.Errorf("LinearX: got %d, want 10", g.Calculate(10, 20, 100))
		}
		if g.Calculate(-5, 15, 100) != -5 {
			t.Errorf("LinearX negative: got %d, want -5", g.Calculate(-5, 15, 100))
		}
	})

	t.Run("GradientLinearY", func(t *testing.T) {
		g := GradientLinearY{}

		// Test that it returns Y coordinate
		if g.Calculate(10, 20, 100) != 20 {
			t.Errorf("LinearY: got %d, want 20", g.Calculate(10, 20, 100))
		}
		if g.Calculate(15, -7, 100) != -7 {
			t.Errorf("LinearY negative: got %d, want -7", g.Calculate(15, -7, 100))
		}
	})

	t.Run("GradientRadial", func(t *testing.T) {
		g := GradientRadial{}

		// Test basic radial gradient (Pythagorean theorem)
		result := g.Calculate(3, 4, 100)
		expected := 5 // sqrt(3^2 + 4^2) = 5
		if result != expected {
			t.Errorf("Radial (3,4): got %d, want %d", result, expected)
		}

		// Test origin
		if g.Calculate(0, 0, 100) != 0 {
			t.Errorf("Radial origin: got %d, want 0", g.Calculate(0, 0, 100))
		}
	})

	t.Run("GradientRadialDouble", func(t *testing.T) {
		g := GradientRadialDouble{}

		// Test with precise floating point calculation
		result := g.Calculate(3, 4, 100)
		if result != 5 {
			t.Errorf("RadialDouble (3,4): got %d, want 5", result)
		}
	})

	t.Run("GradientDiamond", func(t *testing.T) {
		g := GradientDiamond{}

		// Diamond gradient returns max of abs(x), abs(y)
		tests := []struct {
			x, y, expected int
		}{
			{3, 4, 4},   // max(3, 4) = 4
			{-5, 2, 5},  // max(5, 2) = 5
			{-3, -7, 7}, // max(3, 7) = 7
			{0, 0, 0},   // max(0, 0) = 0
		}

		for _, tt := range tests {
			result := g.Calculate(tt.x, tt.y, 100)
			if result != tt.expected {
				t.Errorf("Diamond (%d,%d): got %d, want %d", tt.x, tt.y, result, tt.expected)
			}
		}
	})

	t.Run("GradientConic", func(t *testing.T) {
		g := GradientConic{}

		// Test some known angles
		// At (1, 0), angle is 0, so result should be 0
		result1 := g.Calculate(1, 0, 100)
		if result1 != 0 {
			t.Errorf("Conic (1,0): got %d, want 0", result1)
		}

		// At (0, 1), angle is π/2, so result should be around d2/2
		result2 := g.Calculate(0, 1, 100)
		expected2 := 50 // (π/2) * 100 / π = 50
		if absInt(result2-expected2) > 2 {
			t.Errorf("Conic (0,1): got %d, want ~%d", result2, expected2)
		}
	})
}

func TestGradientRadialFocus(t *testing.T) {
	t.Run("BasicFocusGradient", func(t *testing.T) {
		g := NewGradientRadialFocus(100.0, 25.0, 25.0)

		// Check accessors
		if math.Abs(g.Radius()-100.0) > 0.1 {
			t.Errorf("Radius: got %f, want 100.0", g.Radius())
		}
		if math.Abs(g.FocusX()-25.0) > 0.1 {
			t.Errorf("FocusX: got %f, want 25.0", g.FocusX())
		}
		if math.Abs(g.FocusY()-25.0) > 0.1 {
			t.Errorf("FocusY: got %f, want 25.0", g.FocusY())
		}

		// Test calculation (should return some reasonable value)
		result := g.Calculate(0, 0, 1000)
		if result < 0 || result > 10000 {
			t.Errorf("Focus gradient result out of range: got %d", result)
		}
	})

	t.Run("DegenerateCase", func(t *testing.T) {
		// Test with focus on the circle boundary (should be handled gracefully)
		g := NewGradientRadialFocus(100.0, 100.0, 0.0)

		result := g.Calculate(0, 0, 1000)
		if result < 0 {
			t.Errorf("Degenerate focus gradient should not return negative values: got %d", result)
		}
	})
}

func TestGradientAdaptors(t *testing.T) {
	baseGrad := GradientLinearX{}

	t.Run("RepeatAdaptor", func(t *testing.T) {
		repeatGrad := NewGradientRepeatAdaptor(baseGrad)

		// Test normal case
		if repeatGrad.Calculate(5, 0, 10) != 5 {
			t.Errorf("Repeat normal: got %d, want 5", repeatGrad.Calculate(5, 0, 10))
		}

		// Test wrap-around
		result := repeatGrad.Calculate(15, 0, 10)
		if result != 5 { // 15 % 10 = 5
			t.Errorf("Repeat wrap: got %d, want 5", result)
		}

		// Test negative wrap
		result = repeatGrad.Calculate(-3, 0, 10)
		if result != 7 { // -3 % 10 = -3, then -3 + 10 = 7
			t.Errorf("Repeat negative: got %d, want 7", result)
		}
	})

	t.Run("ReflectAdaptor", func(t *testing.T) {
		reflectGrad := NewGradientReflectAdaptor(baseGrad)

		// Test normal case
		if reflectGrad.Calculate(5, 0, 10) != 5 {
			t.Errorf("Reflect normal: got %d, want 5", reflectGrad.Calculate(5, 0, 10))
		}

		// Test reflection
		result := reflectGrad.Calculate(15, 0, 10)
		if result != 5 { // 15 % 20 = 15, then 20 - 15 = 5
			t.Errorf("Reflect wrap: got %d, want 5", result)
		}
	})
}

func TestGradientLinearColor(t *testing.T) {
	// Create test colors
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}

	t.Run("BasicColorGradient", func(t *testing.T) {
		colorGrad := NewGradientLinearColorRGBA8(black, white, 256)

		if colorGrad.Size() != 256 {
			t.Errorf("Color gradient size: got %d, want 256", colorGrad.Size())
		}

		// Test start color
		start := colorGrad.ColorAt(0)
		if start.R != 0 || start.G != 0 || start.B != 0 {
			t.Errorf("Start color: got (%d,%d,%d), want (0,0,0)", start.R, start.G, start.B)
		}

		// Test end color
		end := colorGrad.ColorAt(255)
		if end.R != 255 || end.G != 255 || end.B != 255 {
			t.Errorf("End color: got (%d,%d,%d), want (255,255,255)", end.R, end.G, end.B)
		}

		// Test middle color (should be gray)
		mid := colorGrad.ColorAt(127)
		expected := uint8(127) // Approximate middle value
		tolerance := uint8(2)
		if abs(int(mid.R)-int(expected)) > int(tolerance) {
			t.Errorf("Mid color R: got %d, want ~%d", mid.R, expected)
		}
	})

	t.Run("CustomSize", func(t *testing.T) {
		colorGrad := NewGradientLinearColorRGBA8(black, white, 100)

		if colorGrad.Size() != 100 {
			t.Errorf("Custom size: got %d, want 100", colorGrad.Size())
		}

		// Test that indices work correctly
		end := colorGrad.ColorAt(99)
		if end.R != 255 || end.G != 255 || end.B != 255 {
			t.Errorf("Custom size end color: got (%d,%d,%d), want (255,255,255)", end.R, end.G, end.B)
		}
	})

	t.Run("SetColors", func(t *testing.T) {
		colorGrad := NewGradientLinearColorRGBA8(black, white, 256)

		// Change to red-blue gradient
		red := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}
		blue := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255}
		colorGrad.SetColors(red, blue, 128)

		if colorGrad.Size() != 128 {
			t.Errorf("After SetColors size: got %d, want 128", colorGrad.Size())
		}

		start := colorGrad.ColorAt(0)
		if start.R != 255 || start.G != 0 || start.B != 0 {
			t.Errorf("After SetColors start: got (%d,%d,%d), want (255,0,0)", start.R, start.G, start.B)
		}

		end := colorGrad.ColorAt(127)
		if end.R != 0 || end.G != 0 || end.B != 255 {
			t.Errorf("After SetColors end: got (%d,%d,%d), want (0,0,255)", end.R, end.G, end.B)
		}
	})
}

func TestSpanGradient(t *testing.T) {
	t.Run("BasicLinearSpanGradient", func(t *testing.T) {
		// Set up components
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
		white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}

		spanGrad := NewLinearGradientRGBA8(interp, black, white, 0.0, 100.0, 256)

		// Test span generation
		span := make([]color.RGBA8[color.Linear], 10)
		spanGrad.Generate(span, 0, 0, 10)

		// Verify that colors were generated
		allBlack := true
		for _, c := range span {
			if c.R != 0 || c.G != 0 || c.B != 0 {
				allBlack = false
				break
			}
		}

		if allBlack {
			t.Error("Span gradient should generate varying colors, but all were black")
		}

		// Verify alpha is preserved
		for i, c := range span {
			if c.A != 255 {
				t.Errorf("Span[%d] alpha: got %d, want 255", i, c.A)
			}
		}
	})

	t.Run("RadialSpanGradient", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		red := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 255}
		blue := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: 255}

		spanGrad := NewRadialGradientRGBA8(interp, red, blue, 0.0, 50.0, 128)

		// Test span generation
		span := make([]color.RGBA8[color.Linear], 5)
		spanGrad.Generate(span, 10, 10, 5)

		// Verify colors were generated and vary
		firstColor := span[0]
		hasVariation := false
		for i := 1; i < len(span); i++ {
			if span[i] != firstColor {
				hasVariation = true
				break
			}
		}

		if !hasVariation {
			t.Error("Radial gradient should produce color variation across span")
		}
	})

	t.Run("GradientAccessors", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
		white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}

		spanGrad := NewLinearGradientRGBA8(interp, black, white, 10.0, 90.0, 256)

		// Test distance accessors
		if math.Abs(spanGrad.D1()-10.0) > 0.1 {
			t.Errorf("D1: got %f, want 10.0", spanGrad.D1())
		}
		if math.Abs(spanGrad.D2()-90.0) > 0.1 {
			t.Errorf("D2: got %f, want 90.0", spanGrad.D2())
		}

		// Test setters
		spanGrad.SetD1(5.0)
		spanGrad.SetD2(95.0)

		if math.Abs(spanGrad.D1()-5.0) > 0.1 {
			t.Errorf("After SetD1: got %f, want 5.0", spanGrad.D1())
		}
		if math.Abs(spanGrad.D2()-95.0) > 0.1 {
			t.Errorf("After SetD2: got %f, want 95.0", spanGrad.D2())
		}
	})
}

func TestGradientConstants(t *testing.T) {
	if GradientSubpixelShift != 4 {
		t.Errorf("GradientSubpixelShift: got %d, want 4", GradientSubpixelShift)
	}
	if GradientSubpixelScale != 16 {
		t.Errorf("GradientSubpixelScale: got %d, want 16", GradientSubpixelScale)
	}
	if GradientSubpixelMask != 15 {
		t.Errorf("GradientSubpixelMask: got %d, want 15", GradientSubpixelMask)
	}
}

// Benchmarks
func BenchmarkGradientShapeFunctions(b *testing.B) {
	b.Run("LinearX", func(b *testing.B) {
		g := GradientLinearX{}
		for i := 0; i < b.N; i++ {
			g.Calculate(i, i, 1000)
		}
	})

	b.Run("Radial", func(b *testing.B) {
		g := GradientRadial{}
		for i := 0; i < b.N; i++ {
			g.Calculate(i%100, i%100, 1000)
		}
	})

	b.Run("RadialFocus", func(b *testing.B) {
		g := NewGradientRadialFocus(100.0, 25.0, 25.0)
		for i := 0; i < b.N; i++ {
			g.Calculate(i%200-100, i%200-100, 1000)
		}
	})

	b.Run("Diamond", func(b *testing.B) {
		g := GradientDiamond{}
		for i := 0; i < b.N; i++ {
			g.Calculate(i%200-100, i%200-100, 1000)
		}
	})
}

func BenchmarkGradientColorFunction(b *testing.B) {
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	colorGrad := NewGradientLinearColorRGBA8(black, white, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		colorGrad.ColorAt(i % 256)
	}
}

func BenchmarkSpanGradientGenerate(b *testing.B) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)

	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}

	spanGrad := NewLinearGradientRGBA8(interp, black, white, 0.0, 100.0, 256)
	span := make([]color.RGBA8[color.Linear], 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spanGrad.Generate(span, i%100, i%100, 100)
	}
}

// Helper function for integer absolute value (avoid duplicate with interpolator_linear_test.go)
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
