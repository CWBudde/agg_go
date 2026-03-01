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
	// If we get here without panic, the gamma function was applied successfully
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

	// Test master alpha affects rendering
	agg2d.SetMasterAlpha(0.5)
	agg2d.FillColor(Color{255, 0, 0, 255}) // Red with full alpha

	// Render a simple rectangle to test alpha application
	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	// The master alpha should be applied during rendering
	// We don't check the exact pixel values here, but verify no crash occurs
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
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create a buffer and attach it
	buf := make([]uint8, 200*200*4) // 200x200 RGBA
	agg2d.Attach(buf, 200, 200, 200*4)

	// Create a simple test image
	testImageData := make([]uint8, 400) // 10x10x4

	// Fill test image with red pixels
	for i := 0; i < 400; i += 4 {
		testImageData[i] = 255   // R
		testImageData[i+1] = 0   // G
		testImageData[i+2] = 0   // B
		testImageData[i+3] = 255 // A
	}

	testImage := &Image{
		width:  10,
		height: 10,
		Data:   testImageData,
	}
	testImage.Attach(testImageData, 10, 10, 40)

	// Test without path - should fall back to regular transform
	err := agg2d.TransformImagePath(testImage, 0, 0, 10, 10, 50, 50, 60, 60)
	if err != nil {
		t.Errorf("TransformImagePath without path should not error: %v", err)
	}

	// Add a simple rectangular path
	agg2d.ResetPath()
	agg2d.Rectangle(45, 45, 65, 65)

	// Test with path - should apply bounding box clipping
	err = agg2d.TransformImagePath(testImage, 0, 0, 10, 10, 50, 50, 60, 60)
	if err != nil {
		t.Errorf("TransformImagePath with path should not error: %v", err)
	}

	// Test path-based parallelogram transformation
	parallelogram := []float64{50, 50, 60, 50, 55, 65}
	err = agg2d.TransformImagePathParallelogram(testImage, 0, 0, 10, 10, parallelogram)
	if err != nil {
		t.Errorf("TransformImagePathParallelogram should not error: %v", err)
	}

	// Test simple versions
	err = agg2d.TransformImagePathSimple(testImage, 50, 50, 60, 60)
	if err != nil {
		t.Errorf("TransformImagePathSimple should not error: %v", err)
	}

	err = agg2d.TransformImagePathParallelogramSimple(testImage, parallelogram)
	if err != nil {
		t.Errorf("TransformImagePathParallelogramSimple should not error: %v", err)
	}
}

func TestRenderingIntegration(t *testing.T) {
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create a buffer and attach it
	buf := make([]uint8, 100*100*4) // 100x100 RGBA
	agg2d.Attach(buf, 100, 100, 100*4)

	// Test combined features: gamma + master alpha + blend mode
	agg2d.SetAntiAliasGamma(1.8)
	agg2d.SetMasterAlpha(0.7)
	agg2d.SetBlendMode(BlendMultiply)

	// Set colors
	agg2d.FillColor(Color{128, 255, 128, 255})
	agg2d.LineColor(Color{255, 128, 128, 255})
	agg2d.LineWidth(2.0)

	// Render shapes to test integration
	agg2d.Rectangle(10, 10, 50, 50)
	agg2d.DrawPath(FillOnly)

	agg2d.Rectangle(20, 20, 60, 60)
	agg2d.DrawPath(StrokeOnly)

	// Draw a circle with path
	agg2d.ResetPath()
	agg2d.Ellipse(70, 70, 15, 15) // Circle is an ellipse with equal radii
	agg2d.DrawPath(FillAndStroke)

	// If we reach here without panicking, the integration works
}

// Helper function to test if two colors are approximately equal (for floating point comparisons)
func colorsApproxEqual(c1, c2 Color, tolerance uint8) bool {
	return math.Abs(float64(c1[0])-float64(c2[0])) <= float64(tolerance) &&
		math.Abs(float64(c1[1])-float64(c2[1])) <= float64(tolerance) &&
		math.Abs(float64(c1[2])-float64(c2[2])) <= float64(tolerance) &&
		math.Abs(float64(c1[3])-float64(c2[3])) <= float64(tolerance)
}
