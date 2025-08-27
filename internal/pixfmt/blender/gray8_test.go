package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestBlenderGray(t *testing.T) {
	blender := BlenderGray8[color.Linear]{}

	// Test blending with full coverage
	dst := basics.Int8u(100)
	cv := basics.Int8u(200)
	alpha := basics.Int8u(128) // 50% alpha
	cover := basics.Int8u(255) // Full coverage

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be between 100 and 200
	if dst <= 100 || dst >= 200 {
		t.Errorf("BlendPix result %d should be between 100 and 200", dst)
	}
}

func TestBlenderGrayPre(t *testing.T) {
	blender := BlenderGray8Pre[color.Linear]{}

	// Test premultiplied blending
	dst := basics.Int8u(100)
	cv := basics.Int8u(200)
	alpha := basics.Int8u(128) // 50% alpha
	cover := basics.Int8u(255) // Full coverage

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be different from non-premultiplied
	// (exact value depends on the premultiplied calculation)
	if dst == 100 {
		t.Error("BlendPix should have changed the destination value")
	}
}

func TestBlendGrayPixel(t *testing.T) {
	dst := basics.Int8u(50)
	src := color.NewGray8WithAlpha[color.Linear](150, 200)
	cover := basics.Int8u(255)
	blender := BlenderGray8[color.Linear]{}

	BlendGrayPixel(&dst, src, cover, blender)

	// Should blend 50 towards 150 with 200/255 alpha
	if dst <= 50 || dst >= 150 {
		t.Errorf("BlendGrayPixel result %d should be between 50 and 150", dst)
	}
}

func TestBlendGrayPixelZeroAlpha(t *testing.T) {
	original := basics.Int8u(100)
	dst := original
	src := color.NewGray8WithAlpha[color.Linear](200, 0) // Zero alpha
	cover := basics.Int8u(255)
	blender := BlenderGray8[color.Linear]{}

	BlendGrayPixel(&dst, src, cover, blender)

	// Should not change destination when alpha is 0
	if dst != original {
		t.Errorf("BlendGrayPixel with zero alpha should not change dst: expected %d, got %d", original, dst)
	}
}

func TestCopyGrayPixel(t *testing.T) {
	dst := basics.Int8u(100)
	src := color.NewGray8WithAlpha[color.Linear](200, 150)

	CopyGrayPixel(&dst, src)

	// Should copy the value regardless of alpha
	if dst != 200 {
		t.Errorf("CopyGrayPixel should copy src.V: expected 200, got %d", dst)
	}
}

func TestBlendGrayHline(t *testing.T) {
	// Create a test row
	dst := make([]basics.Int8u, 10)
	for i := range dst {
		dst[i] = 50 // Background color
	}

	src := color.NewGray8WithAlpha[color.Linear](150, 200)
	blender := BlenderGray8[color.Linear]{}

	// Blend horizontal line
	x := 2
	length := 5
	BlendGrayHline(dst, x, length, src, nil, blender)

	// Check affected pixels
	for i := x; i < x+length; i++ {
		if dst[i] <= 50 || dst[i] >= 150 {
			t.Errorf("BlendGrayHline pixel %d should be between 50 and 150, got %d", i, dst[i])
		}
	}

	// Check unaffected pixels
	if dst[0] != 50 || dst[1] != 50 || dst[7] != 50 {
		t.Error("BlendGrayHline should not affect pixels outside the range")
	}
}

func TestBlendGrayHlineWithCovers(t *testing.T) {
	// Create a test row
	dst := make([]basics.Int8u, 10)
	for i := range dst {
		dst[i] = 50 // Background color
	}

	src := color.NewGray8WithAlpha[color.Linear](150, 255) // Full alpha
	covers := []basics.Int8u{255, 200, 100, 50, 0}         // Varying coverage
	blender := BlenderGray8[color.Linear]{}

	// Blend horizontal line with varying coverage
	x := 2
	length := len(covers)
	BlendGrayHline(dst, x, length, src, covers, blender)

	// Check that higher coverage results in values closer to source
	if dst[x] <= dst[x+1] { // Full coverage should give higher value than partial
		t.Error("Higher coverage should result in values closer to source")
	}
	if dst[x+4] != 50 { // Zero coverage should leave pixel unchanged
		t.Errorf("Zero coverage should leave pixel unchanged: expected 50, got %d", dst[x+4])
	}
}

func TestCopyGrayHline(t *testing.T) {
	// Create a test row
	dst := make([]basics.Int8u, 10)
	for i := range dst {
		dst[i] = 50 // Background color
	}

	src := color.NewGray8WithAlpha[color.Linear](200, 100) // Alpha should be ignored

	// Copy horizontal line
	x := 2
	length := 5
	CopyGrayHline(dst, x, length, src)

	// Check that all affected pixels are exactly the source value
	for i := x; i < x+length; i++ {
		if dst[i] != 200 {
			t.Errorf("CopyGrayHline pixel %d should be 200, got %d", i, dst[i])
		}
	}

	// Check unaffected pixels
	if dst[0] != 50 || dst[1] != 50 || dst[7] != 50 {
		t.Error("CopyGrayHline should not affect pixels outside the range")
	}
}

func TestFillGraySpan(t *testing.T) {
	// Create a test row
	dst := make([]basics.Int8u, 10)
	for i := range dst {
		dst[i] = 50 // Background color
	}

	src := color.NewGray8WithAlpha[color.Linear](175, 128)

	// Fill span
	x := 3
	length := 4
	FillGraySpan(dst, x, length, src)

	// Check filled pixels
	for i := x; i < x+length; i++ {
		if dst[i] != 175 {
			t.Errorf("FillGraySpan pixel %d should be 175, got %d", i, dst[i])
		}
	}

	// Check unfilled pixels
	if dst[0] != 50 || dst[2] != 50 || dst[7] != 50 {
		t.Error("FillGraySpan should not affect pixels outside the range")
	}
}

func TestBlenderGrayPremultiplied(t *testing.T) {
	blender := BlenderGray8Pre[color.Linear]{}

	// Test with values that demonstrate premultiplied behavior
	dst := basics.Int8u(0)     // Start with black
	cv := basics.Int8u(100)    // Premultiplied color value
	alpha := basics.Int8u(100) // Alpha value
	cover := basics.Int8u(255) // Full coverage

	blender.BlendPix(&dst, cv, alpha, cover)

	// With premultiplied blending, the math is different
	// dst = prelerp(dst, cv*cover, alpha*cover)
	expectedCV := color.Gray8Multiply(cv, cover)
	expectedAlpha := color.Gray8Multiply(alpha, cover)
	expected := color.Gray8Prelerp(0, expectedCV, expectedAlpha)

	if dst != expected {
		t.Errorf("Premultiplied blending: expected %d, got %d", expected, dst)
	}
}
