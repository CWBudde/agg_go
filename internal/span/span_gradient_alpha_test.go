package span

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/transform"
)

// Test alpha functions
func TestAlphaFunctions(t *testing.T) {
	t.Run("GradientAlphaLinear", func(t *testing.T) {
		alphaFunc := NewGradientAlphaLinear(0, 255, 256)

		if alphaFunc.Size() != 256 {
			t.Errorf("Size: got %d, want 256", alphaFunc.Size())
		}

		// Test start value
		if alphaFunc.AlphaAt(0) != 0 {
			t.Errorf("Start alpha: got %d, want 0", alphaFunc.AlphaAt(0))
		}

		// Test end value
		if alphaFunc.AlphaAt(255) != 255 {
			t.Errorf("End alpha: got %d, want 255", alphaFunc.AlphaAt(255))
		}

		// Test middle value (should be around 127)
		mid := alphaFunc.AlphaAt(128)
		if absAlpha(int(mid)-128) > 2 {
			t.Errorf("Middle alpha: got %d, want ~128", mid)
		}

		// Test out of bounds
		if alphaFunc.AlphaAt(-1) != 0 {
			t.Errorf("Negative index: got %d, want 0", alphaFunc.AlphaAt(-1))
		}
		if alphaFunc.AlphaAt(300) != 255 {
			t.Errorf("Oversized index: got %d, want 255", alphaFunc.AlphaAt(300))
		}
	})

	t.Run("GradientAlphaX", func(t *testing.T) {
		alphaFunc := NewGradientAlphaX(256)

		if alphaFunc.Size() != 256 {
			t.Errorf("Size: got %d, want 256", alphaFunc.Size())
		}

		// Test identity function
		tests := []int{0, 50, 100, 200, 255}
		for _, test := range tests {
			result := alphaFunc.AlphaAt(test)
			if int(result) != test {
				t.Errorf("Identity[%d]: got %d, want %d", test, result, test)
			}
		}

		// Test bounds
		if alphaFunc.AlphaAt(-1) != 0 {
			t.Errorf("Negative index: got %d, want 0", alphaFunc.AlphaAt(-1))
		}
		if alphaFunc.AlphaAt(256) != 255 {
			t.Errorf("Oversized index: got %d, want 255", alphaFunc.AlphaAt(256))
		}
	})

	t.Run("GradientAlphaOneMinusX", func(t *testing.T) {
		alphaFunc := NewGradientAlphaOneMinusX(256)

		if alphaFunc.Size() != 256 {
			t.Errorf("Size: got %d, want 256", alphaFunc.Size())
		}

		// Test inverse function
		tests := []struct {
			input, expected int
		}{
			{0, 255},
			{50, 205},
			{100, 155},
			{200, 55},
			{255, 0},
		}

		for _, test := range tests {
			result := alphaFunc.AlphaAt(test.input)
			if int(result) != test.expected {
				t.Errorf("OneMinusX[%d]: got %d, want %d", test.input, result, test.expected)
			}
		}

		// Test bounds
		if alphaFunc.AlphaAt(-1) != 255 {
			t.Errorf("Negative index: got %d, want 255", alphaFunc.AlphaAt(-1))
		}
	})

	t.Run("GradientAlphaLUT", func(t *testing.T) {
		values := []basics.Int8u{10, 50, 100, 150, 200}
		alphaFunc := NewGradientAlphaLUT(values)

		if alphaFunc.Size() != 5 {
			t.Errorf("Size: got %d, want 5", alphaFunc.Size())
		}

		// Test stored values
		for i, expected := range values {
			result := alphaFunc.AlphaAt(i)
			if result != expected {
				t.Errorf("LUT[%d]: got %d, want %d", i, result, expected)
			}
		}

		// Test modification
		alphaFunc.SetAlphaAt(2, 75)
		if alphaFunc.AlphaAt(2) != 75 {
			t.Errorf("After modification: got %d, want 75", alphaFunc.AlphaAt(2))
		}

		// Test out of bounds
		if alphaFunc.AlphaAt(-1) != values[0] {
			t.Errorf("Negative index: got %d, want %d", alphaFunc.AlphaAt(-1), values[0])
		}
		if alphaFunc.AlphaAt(10) != values[4] {
			t.Errorf("Oversized index: got %d, want %d", alphaFunc.AlphaAt(10), values[4])
		}

		// Test values copy
		copied := alphaFunc.Values()
		if len(copied) != len(values) {
			t.Errorf("Values copy length: got %d, want %d", len(copied), len(values))
		}
	})

	t.Run("SetValues", func(t *testing.T) {
		alphaFunc := NewGradientAlphaLinear(0, 100, 100)
		alphaFunc.SetValues(50, 200, 200)

		if alphaFunc.Size() != 200 {
			t.Errorf("After SetValues size: got %d, want 200", alphaFunc.Size())
		}
		if alphaFunc.AlphaAt(0) != 50 {
			t.Errorf("After SetValues start: got %d, want 50", alphaFunc.AlphaAt(0))
		}
		if alphaFunc.AlphaAt(199) != 200 {
			t.Errorf("After SetValues end: got %d, want 200", alphaFunc.AlphaAt(199))
		}
	})
}

func TestAlphaWrappers(t *testing.T) {
	t.Run("RGBA8AlphaWrapper", func(t *testing.T) {
		rgba := &color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 128}
		wrapper := NewRGBA8AlphaWrapper(rgba)

		// Test get alpha
		if wrapper.GetAlpha() != 128 {
			t.Errorf("GetAlpha: got %d, want 128", wrapper.GetAlpha())
		}

		// Test set alpha
		wrapper.SetAlpha(200)
		if wrapper.GetAlpha() != 200 {
			t.Errorf("After SetAlpha: got %d, want 200", wrapper.GetAlpha())
		}
		if rgba.A != 200 {
			t.Errorf("Original color alpha not updated: got %d, want 200", rgba.A)
		}

		// RGB values should be unchanged
		if rgba.R != 100 || rgba.G != 150 || rgba.B != 200 {
			t.Errorf("RGB values changed: got (%d,%d,%d), want (100,150,200)", rgba.R, rgba.G, rgba.B)
		}
	})

	t.Run("Gray8AlphaWrapper", func(t *testing.T) {
		gray := &color.Gray8[color.Linear]{V: 100, A: 128}
		wrapper := NewGray8AlphaWrapper(gray)

		// Test get alpha
		if wrapper.GetAlpha() != 128 {
			t.Errorf("GetAlpha: got %d, want 128", wrapper.GetAlpha())
		}

		// Test set alpha
		wrapper.SetAlpha(200)
		if wrapper.GetAlpha() != 200 {
			t.Errorf("After SetAlpha: got %d, want 200", wrapper.GetAlpha())
		}
		if gray.A != 200 {
			t.Errorf("Original color alpha not updated: got %d, want 200", gray.A)
		}

		// Value should be unchanged
		if gray.V != 100 {
			t.Errorf("Gray value changed: got %d, want 100", gray.V)
		}
	})
}

func TestSpanGradientAlpha(t *testing.T) {
	t.Run("BasicAlphaSpanGradient", func(t *testing.T) {
		// Set up components
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		spanAlphaGrad := NewLinearAlphaGradientRGBA8[*SpanInterpolatorLinear[*transform.TransAffine], color.Linear](interp, 0, 255, 0.0, 100.0, 256)

		// Create span of colors with initial alpha values
		span := make([]RGBA8AlphaWrapper[color.Linear], 10)
		for i := range span {
			rgba := &color.RGBA8[color.Linear]{R: 100, G: 150, B: 200, A: 64} // Initial alpha 64
			span[i] = NewRGBA8AlphaWrapper(rgba)
		}

		// Generate alpha gradient
		spanAlphaGrad.Generate(span, 0, 0, 10)

		// Check that alpha values were modified
		alphaChanged := false
		for i := range span {
			if span[i].GetAlpha() != 64 {
				alphaChanged = true
				break
			}
		}

		if !alphaChanged {
			t.Error("Alpha gradient should modify alpha values")
		}

		// Check that RGB values are preserved
		for i := range span {
			if span[i].Color.R != 100 || span[i].Color.G != 150 || span[i].Color.B != 200 {
				t.Errorf("RGB values changed at index %d: got (%d,%d,%d), want (100,150,200)",
					i, span[i].Color.R, span[i].Color.G, span[i].Color.B)
			}
		}
	})

	t.Run("RadialAlphaSpanGradient", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		spanAlphaGrad := NewRadialAlphaGradientRGBA8[*SpanInterpolatorLinear[*transform.TransAffine], color.Linear](interp, 255, 0, 0.0, 50.0, 128)

		// Create span of colors
		span := make([]RGBA8AlphaWrapper[color.Linear], 5)
		for i := range span {
			rgba := &color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: 128}
			span[i] = NewRGBA8AlphaWrapper(rgba)
		}

		// Generate radial alpha gradient
		spanAlphaGrad.Generate(span, 10, 10, 5)

		// Verify alpha values vary
		firstAlpha := span[0].GetAlpha()
		hasVariation := false
		for i := 1; i < len(span); i++ {
			if span[i].GetAlpha() != firstAlpha {
				hasVariation = true
				break
			}
		}

		if !hasVariation {
			t.Error("Radial alpha gradient should produce alpha variation across span")
		}
	})

	t.Run("AlphaGradientAccessors", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		spanAlphaGrad := NewLinearAlphaGradientRGBA8[*SpanInterpolatorLinear[*transform.TransAffine], color.Linear](interp, 0, 255, 10.0, 90.0, 256)

		// Test distance accessors
		if absAlpha(int(spanAlphaGrad.D1()*10)-100) > 1 {
			t.Errorf("D1: got %f, want ~10.0", spanAlphaGrad.D1())
		}
		if absAlpha(int(spanAlphaGrad.D2()*10)-900) > 1 {
			t.Errorf("D2: got %f, want ~90.0", spanAlphaGrad.D2())
		}

		// Test setters
		spanAlphaGrad.SetD1(5.0)
		spanAlphaGrad.SetD2(95.0)

		if absAlpha(int(spanAlphaGrad.D1()*10)-50) > 1 {
			t.Errorf("After SetD1: got %f, want ~5.0", spanAlphaGrad.D1())
		}
		if absAlpha(int(spanAlphaGrad.D2()*10)-950) > 1 {
			t.Errorf("After SetD2: got %f, want ~95.0", spanAlphaGrad.D2())
		}
	})

	t.Run("AlphaGradientWithDifferentShapes", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Test with diamond gradient
		diamondGrad := GradientDiamond{}
		alphaFunc := NewGradientAlphaLinear(0, 255, 100)

		spanAlphaGrad := NewSpanGradientAlpha[RGBA8AlphaWrapper[color.Linear], *SpanInterpolatorLinear[*transform.TransAffine], GradientDiamond, *GradientAlphaLinear](
			interp, diamondGrad, alphaFunc, 0.0, 50.0)

		// Create span
		span := make([]RGBA8AlphaWrapper[color.Linear], 5)
		for i := range span {
			rgba := &color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: 128}
			span[i] = NewRGBA8AlphaWrapper(rgba)
		}

		// Generate gradient
		spanAlphaGrad.Generate(span, 0, 0, 5)

		// Should complete without error and modify alpha values
		for i := range span {
			if span[i].Color.R != 0 || span[i].Color.G != 255 || span[i].Color.B != 0 {
				t.Errorf("RGB values changed at index %d", i)
			}
		}
	})

	t.Run("AlphaGradientWithCustomLUT", func(t *testing.T) {
		trans := transform.NewTransAffine()
		interp := NewSpanInterpolatorLinearDefault(trans)

		// Create custom alpha LUT
		customAlphas := []basics.Int8u{255, 200, 150, 100, 50, 25, 0}
		alphaLUT := NewGradientAlphaLUT(customAlphas)

		gradientFunc := GradientLinearX{}
		spanAlphaGrad := NewSpanGradientAlpha[RGBA8AlphaWrapper[color.Linear], *SpanInterpolatorLinear[*transform.TransAffine], GradientLinearX, *GradientAlphaLUT](
			interp, gradientFunc, alphaLUT, 0.0, float64(len(customAlphas)-1))

		// Create span
		span := make([]RGBA8AlphaWrapper[color.Linear], len(customAlphas))
		for i := range span {
			rgba := &color.RGBA8[color.Linear]{R: 100, G: 100, B: 100, A: 128}
			span[i] = NewRGBA8AlphaWrapper(rgba)
		}

		// Generate gradient (should use LUT values)
		spanAlphaGrad.Generate(span, 0, 0, len(customAlphas))

		// Verify some alpha values correspond to LUT (allowing for interpolation/mapping)
		hasExpectedValues := false
		for i := range span {
			alpha := span[i].GetAlpha()
			for _, expectedAlpha := range customAlphas {
				if alpha == expectedAlpha {
					hasExpectedValues = true
					break
				}
			}
		}

		if !hasExpectedValues {
			t.Error("Custom LUT should produce some matching alpha values")
		}
	})
}

func TestAlphaGradientPrepare(t *testing.T) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)
	spanAlphaGrad := NewLinearAlphaGradientRGBA8[*SpanInterpolatorLinear[*transform.TransAffine], color.Linear](interp, 0, 255, 0.0, 100.0, 256)

	// Prepare should not panic (it's a no-op)
	spanAlphaGrad.Prepare()
}

// Benchmarks
func BenchmarkAlphaFunctions(b *testing.B) {
	b.Run("GradientAlphaLinear", func(b *testing.B) {
		alphaFunc := NewGradientAlphaLinear(0, 255, 256)
		for i := 0; i < b.N; i++ {
			alphaFunc.AlphaAt(i % 256)
		}
	})

	b.Run("GradientAlphaX", func(b *testing.B) {
		alphaFunc := NewGradientAlphaX(256)
		for i := 0; i < b.N; i++ {
			alphaFunc.AlphaAt(i % 256)
		}
	})

	b.Run("GradientAlphaOneMinusX", func(b *testing.B) {
		alphaFunc := NewGradientAlphaOneMinusX(256)
		for i := 0; i < b.N; i++ {
			alphaFunc.AlphaAt(i % 256)
		}
	})

	b.Run("GradientAlphaLUT", func(b *testing.B) {
		values := make([]basics.Int8u, 256)
		for i := range values {
			values[i] = basics.Int8u(i)
		}
		alphaFunc := NewGradientAlphaLUT(values)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			alphaFunc.AlphaAt(i % 256)
		}
	})
}

func BenchmarkSpanGradientAlphaGenerate(b *testing.B) {
	trans := transform.NewTransAffine()
	interp := NewSpanInterpolatorLinearDefault(trans)

	spanAlphaGrad := NewLinearAlphaGradientRGBA8[*SpanInterpolatorLinear[*transform.TransAffine], color.Linear](interp, 0, 255, 0.0, 100.0, 256)

	span := make([]RGBA8AlphaWrapper[color.Linear], 100)
	for i := range span {
		rgba := &color.RGBA8[color.Linear]{R: 128, G: 128, B: 128, A: 128}
		span[i] = NewRGBA8AlphaWrapper(rgba)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spanAlphaGrad.Generate(span, i%100, i%100, 100)
	}
}

func BenchmarkAlphaWrappers(b *testing.B) {
	rgba := &color.RGBA8[color.Linear]{R: 128, G: 128, B: 128, A: 128}
	wrapper := NewRGBA8AlphaWrapper(rgba)

	b.Run("SetAlpha", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			wrapper.SetAlpha(basics.Int8u(i % 256))
		}
	})

	b.Run("GetAlpha", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = wrapper.GetAlpha()
		}
	})
}

// Helper function for integer absolute value (avoid duplicate with other test files)
func absAlpha(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
