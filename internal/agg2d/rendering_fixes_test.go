// Package agg2d rendering fixes tests
// This file tests the newly implemented rendering features: gamma correction, master alpha, and blend modes
package agg2d

import (
	"math"
	"testing"

	"agg_go/internal/pixfmt/blender"
)

func TestGammaCorrection(t *testing.T) {
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create a buffer and attach it
	buf := make([]uint8, 100*100*4) // 100x100 RGBA
	agg2d.Attach(buf, 100, 100, 100*4)

	// Test default gamma (should be 1.0)
	if gamma := agg2d.GetAntiAliasGamma(); gamma != 1.0 {
		t.Errorf("Expected default gamma to be 1.0, got %f", gamma)
	}

	// Test setting gamma
	testGamma := 2.2
	agg2d.SetAntiAliasGamma(testGamma)
	if gamma := agg2d.GetAntiAliasGamma(); gamma != testGamma {
		t.Errorf("Expected gamma to be %f, got %f", testGamma, gamma)
	}

	// Test gamma bounds
	agg2d.SetAntiAliasGamma(-1.0) // Should clamp to minimum
	if gamma := agg2d.GetAntiAliasGamma(); gamma <= 0 {
		t.Error("Gamma should not be zero or negative")
	}

	// Test gamma correction is applied to rasterizer
	agg2d.SetAntiAliasGamma(1.5)
	agg2d.updateRasterizerGamma()
	if got := agg2d.rasterizer.ApplyGamma(64); got == 64 {
		t.Fatalf("expected non-identity gamma mapping after updateRasterizerGamma, got %d", got)
	}
}

func TestAttachResetsRasterizerGamma(t *testing.T) {
	agg2d := NewAgg2D()

	buf := make([]uint8, 100*100*4)
	agg2d.Attach(buf, 100, 100, 100*4)

	agg2d.SetAntiAliasGamma(2.0)
	if got := agg2d.rasterizer.ApplyGamma(128); got == 128 {
		t.Fatalf("expected non-linear gamma after SetAntiAliasGamma, got identity result %d", got)
	}

	agg2d.Attach(buf, 100, 100, 100*4)
	if gamma := agg2d.GetAntiAliasGamma(); gamma != 1.0 {
		t.Fatalf("expected attach to reset stored gamma to 1.0, got %v", gamma)
	}
	if got := agg2d.rasterizer.ApplyGamma(128); got != 128 {
		t.Fatalf("expected attach to reset rasterizer gamma to identity, got %d", got)
	}
}

func TestMasterAlpha(t *testing.T) {
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create a buffer and attach it
	buf := make([]uint8, 100*100*4) // 100x100 RGBA
	agg2d.Attach(buf, 100, 100, 100*4)

	// Test default master alpha (should be 1.0)
	if alpha := agg2d.GetMasterAlpha(); alpha != 1.0 {
		t.Errorf("Expected default master alpha to be 1.0, got %f", alpha)
	}

	// Test setting master alpha
	testAlpha := 0.5
	agg2d.SetMasterAlpha(testAlpha)
	if alpha := agg2d.GetMasterAlpha(); alpha != testAlpha {
		t.Errorf("Expected master alpha to be %f, got %f", testAlpha, alpha)
	}

	// Test alpha bounds
	agg2d.SetMasterAlpha(-1.0)
	if alpha := agg2d.GetMasterAlpha(); alpha != 0.0 {
		t.Errorf("Expected clamped alpha to be 0.0, got %f", alpha)
	}

	agg2d.SetMasterAlpha(2.0)
	if alpha := agg2d.GetMasterAlpha(); alpha != 1.0 {
		t.Errorf("Expected clamped alpha to be 1.0, got %f", alpha)
	}

	// Test master alpha affects rendering — verify alpha attenuation.
	agg2d.SetMasterAlpha(0.25)
	agg2d.FillColor(Color{255, 0, 0, 255}) // Red with full alpha
	// Clear to white first.
	agg2d.ClearAllRGBA(255, 255, 255, 255)

	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	// Interior pixel: master alpha scales coverage via the gamma table.
	// With src-over compositing on a white background, the output alpha
	// stays 255 but the red channel should be attenuated (blended with white).
	r, g, b, _ := pixelAt(buf, 100, 30, 30)
	if r == 255 && g == 255 && b == 255 {
		t.Fatal("master alpha 0.25: interior pixel is still white, expected rendering")
	}
	if r == 0 {
		t.Fatal("master alpha 0.25: interior pixel has zero red, expected red fill")
	}
	// Green/blue channels should be > 0 because the red is blended with white background.
	if g == 0 && b == 0 {
		t.Fatal("master alpha 0.25: expected partial blend with white, but g=b=0 indicates full opacity")
	}
}

func TestMasterAlphaAffectsRasterizerGamma(t *testing.T) {
	agg2d := NewAgg2D()

	buf := make([]uint8, 100*100*4)
	agg2d.Attach(buf, 100, 100, 100*4)

	agg2d.SetMasterAlpha(0.5)
	got := agg2d.rasterizer.ApplyGamma(255)
	if got != 127 {
		t.Fatalf("expected master alpha to scale full coverage to 127, got %d", got)
	}

	agg2d.SetAntiAliasGamma(2.0)
	got = agg2d.rasterizer.ApplyGamma(64)
	want := uint8(0.5 * math.Pow(float64(64)/255.0, 0.5) * 255.0)
	if got != want {
		t.Fatalf("expected combined master alpha/gamma coverage %d, got %d", want, got)
	}
}

func TestBlendModes(t *testing.T) {
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create a buffer and attach it
	buf := make([]uint8, 100*100*4) // 100x100 RGBA
	agg2d.Attach(buf, 100, 100, 100*4)

	// Test default blend mode
	if mode := agg2d.GetBlendMode(); mode != BlendAlpha {
		t.Errorf("Expected default blend mode to be BlendAlpha, got %d", mode)
	}

	// Test setting various blend modes
	testModes := []BlendMode{
		BlendClear, BlendSrc, BlendDst, BlendSrcOver, BlendDstOver,
		BlendSrcIn, BlendDstIn, BlendSrcOut, BlendDstOut,
		BlendSrcAtop, BlendDstAtop, BlendXor, BlendAdd,
		BlendMultiply, BlendScreen, BlendOverlay,
		BlendDarken, BlendLighten, BlendColorDodge, BlendColorBurn,
		BlendHardLight, BlendSoftLight, BlendDifference, BlendExclusion,
	}

	for _, mode := range testModes {
		agg2d.SetBlendMode(mode)
		if currentMode := agg2d.GetBlendMode(); currentMode != mode {
			t.Errorf("Expected blend mode %d, got %d", mode, currentMode)
		}

		// Test blend mode conversion
		compOp := blendModeToCompOp(mode)
		if compOp < 0 {
			t.Errorf("Invalid CompOp conversion for blend mode %d", mode)
		}
	}

	// Test image blend modes
	agg2d.SetImageBlendMode(BlendMultiply)
	if mode := agg2d.GetImageBlendMode(); mode != BlendMultiply {
		t.Errorf("Expected image blend mode to be BlendMultiply, got %d", mode)
	}

	// Test image blend color
	testColor := Color{128, 64, 192, 255}
	agg2d.SetImageBlendColor(testColor)
	if color := agg2d.GetImageBlendColor(); color != testColor {
		t.Errorf("Expected image blend color %v, got %v", testColor, color)
	}
}

func TestBlendModeConversion(t *testing.T) {
	// Test specific blend mode to CompOp conversions
	tests := []struct {
		blendMode BlendMode
		expected  blender.CompOp
	}{
		{BlendAlpha, blender.CompOpSrcOver},
		{BlendClear, blender.CompOpClear},
		{BlendSrc, blender.CompOpSrc},
		{BlendAdd, blender.CompOpPlus},
		{BlendMultiply, blender.CompOpMultiply},
		{BlendScreen, blender.CompOpScreen},
	}

	for _, test := range tests {
		result := blendModeToCompOp(test.blendMode)
		if result != test.expected {
			t.Errorf("blendModeToCompOp(%d) = %d, expected %d", test.blendMode, result, test.expected)
		}
	}
}

func TestPathBasedImageTransformations(t *testing.T) {
	width := 200
	stride := width * 4

	mkImage := func() *Image {
		data := make([]uint8, 10*10*4)
		for i := 0; i < len(data); i += 4 {
			data[i], data[i+1], data[i+2], data[i+3] = 255, 0, 0, 255 // red
		}
		img := &Image{width: 10, height: 10, Data: data}
		img.Attach(data, 10, 10, 40)
		return img
	}

	t.Run("no_path_no_output", func(t *testing.T) {
		// TransformImagePath without a user path rasterizes nothing
		// because it renders only the current user path.
		ctx := NewAgg2D()
		buf := make([]uint8, width*width*4)
		ctx.Attach(buf, width, width, stride)
		err := ctx.TransformImagePath(mkImage(), 0, 0, 10, 10, 50, 50, 60, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Buffer should remain untouched (all zeroes).
		for i := range buf {
			if buf[i] != 0 {
				t.Fatal("expected no output without a path, but pixels were modified")
			}
		}
	})

	t.Run("with_path_renders", func(t *testing.T) {
		ctx := NewAgg2D()
		buf := make([]uint8, width*width*4)
		for i := 0; i < len(buf); i += 4 {
			buf[i], buf[i+1], buf[i+2], buf[i+3] = 255, 255, 255, 255
		}
		ctx.Attach(buf, width, width, stride)

		// Define a clipping path then render.
		ctx.ResetPath()
		ctx.Rectangle(45, 45, 65, 65)
		err := ctx.TransformImagePath(mkImage(), 0, 0, 10, 10, 50, 50, 60, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have rendered red-ish pixels in the target region.
		if !hasNonWhiteIn(buf, width, 50, 50, 60, 60) {
			t.Fatal("TransformImagePath with path should render visible pixels")
		}
		// Exterior should remain white.
		r, g, b, a := pixelAt(buf, width, 5, 5)
		if !(r == 255 && g == 255 && b == 255 && a == 255) {
			t.Fatalf("exterior pixel = (%d,%d,%d,%d), want white", r, g, b, a)
		}
	})

	t.Run("parallelogram", func(t *testing.T) {
		ctx := NewAgg2D()
		buf := make([]uint8, width*width*4)
		ctx.Attach(buf, width, width, stride)
		ctx.ResetPath()
		ctx.Rectangle(45, 45, 65, 65)
		parallelogram := []float64{50, 50, 60, 50, 55, 65}
		err := ctx.TransformImagePathParallelogram(mkImage(), 0, 0, 10, 10, parallelogram)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("simple_variants", func(t *testing.T) {
		ctx := NewAgg2D()
		buf := make([]uint8, width*width*4)
		ctx.Attach(buf, width, width, stride)
		ctx.ResetPath()
		ctx.Rectangle(45, 45, 65, 65)
		if err := ctx.TransformImagePathSimple(mkImage(), 50, 50, 60, 60); err != nil {
			t.Fatalf("TransformImagePathSimple: %v", err)
		}
		parallelogram := []float64{50, 50, 60, 50, 55, 65}
		if err := ctx.TransformImagePathParallelogramSimple(mkImage(), parallelogram); err != nil {
			t.Fatalf("TransformImagePathParallelogramSimple: %v", err)
		}
	})
}

func TestRenderingIntegration(t *testing.T) {
	agg2d := NewAgg2D()

	width := 100
	buf := make([]uint8, width*width*4)
	// Start with white background.
	for i := 0; i < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = 255, 255, 255, 255
	}
	agg2d.Attach(buf, width, width, width*4)

	// Combined features: gamma + master alpha + blend mode.
	agg2d.SetAntiAliasGamma(1.8)
	agg2d.SetMasterAlpha(0.7)
	agg2d.SetBlendMode(BlendMultiply)

	// Green-ish fill, red-ish stroke.
	agg2d.FillColor(Color{128, 255, 128, 255})
	agg2d.LineColor(Color{255, 128, 128, 255})
	agg2d.LineWidth(2.0)

	// Filled rectangle.
	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	// Stroked rectangle.
	agg2d.Rectangle(20, 20, 60, 60)
	agg2d.DrawPath(StrokeOnly)

	// Circle.
	agg2d.ResetPath()
	agg2d.Ellipse(70, 70, 15, 15)
	agg2d.DrawPath(FillAndStroke)

	// Verify: filled rect interior should show green-dominant multiply blend.
	_, g, _, a := pixelAt(buf, width, 30, 30)
	if a == 0 {
		t.Fatal("filled rect interior has zero alpha")
	}
	if g < 128 {
		t.Fatalf("filled rect interior green = %d, expected >= 128 for green fill", g)
	}

	// Verify: stroke edges near (20,20) should be non-white.
	if !hasNonWhiteIn(buf, width, 19, 19, 22, 22) {
		t.Fatal("stroked rect should produce visible stroke pixels near (20,20)")
	}

	// Verify: circle center should be non-white.
	if !hasNonWhiteIn(buf, width, 68, 68, 73, 73) {
		t.Fatal("circle center should be non-white")
	}

	// Verify: exterior should remain white.
	r, g2, b, a2 := pixelAt(buf, width, 95, 5)
	if !(r == 255 && g2 == 255 && b == 255 && a2 == 255) {
		t.Fatalf("exterior pixel = (%d,%d,%d,%d), want white", r, g2, b, a2)
	}
}

// Helper function to test if two colors are approximately equal (for floating point comparisons)
func colorsApproxEqual(c1, c2 Color, tolerance uint8) bool {
	return math.Abs(float64(c1[0])-float64(c2[0])) <= float64(tolerance) &&
		math.Abs(float64(c1[1])-float64(c2[1])) <= float64(tolerance) &&
		math.Abs(float64(c1[2])-float64(c2[2])) <= float64(tolerance) &&
		math.Abs(float64(c1[3])-float64(c2[3])) <= float64(tolerance)
}
