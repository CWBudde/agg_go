package blender

import (
	"math"
	"testing"

	"agg_go/internal/color"
)

func TestBlenderGray32BlendPix(t *testing.T) {
	blender := BlenderGray32Linear{}
	dst := float32(0.25)
	cv := float32(0.75)
	alpha := float32(0.5)
	cover := float32(1.0)

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be between original dst and cv
	if dst <= 0.25 || dst >= 0.75 {
		t.Errorf("BlendPix failed: expected blended value between 0.25-0.75, got %f", dst)
	}
}

func TestBlenderGray32PreBlendPix(t *testing.T) {
	blender := BlenderGray32PreLinear{}
	dst := float32(0.25)
	cv := float32(0.75)
	alpha := float32(0.5)
	cover := float32(1.0)

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be different from original dst
	if math.Abs(float64(dst-0.25)) < 0.001 {
		t.Errorf("BlendPix failed: no blending occurred, result still %f", dst)
	}
}

func TestGray32Lerp(t *testing.T) {
	p := float32(0.25)
	q := float32(0.75)
	a := float32(0.5) // halfway
	result := Gray32Lerp(p, q, a)
	expected := float32(0.5) // Should be halfway between p and q

	if math.Abs(float64(result-expected)) > 0.001 {
		t.Errorf("Gray32Lerp failed: expected %f, got %f", expected, result)
	}
}

func TestGray32Prelerp(t *testing.T) {
	p := float32(0.5)
	q := float32(0.25)
	a := float32(0.5)
	result := Gray32Prelerp(p, q, a)

	// Prelerp formula: p + q - p*a
	// With these values: 0.5 + 0.25 - 0.5*0.5 = 0.5 + 0.25 - 0.25 = 0.5
	// So result equals p, which is expected behavior
	expected := float32(0.5)
	if math.Abs(float64(result-expected)) > 0.001 {
		t.Errorf("Gray32Prelerp failed: expected %f, got %f", expected, result)
	}
}

func TestBlendGray32Pixel(t *testing.T) {
	dst := float32(0.25)
	src := color.NewGray32WithAlpha[color.Linear](0.75, 0.5)
	cover := float32(1.0)
	blender := BlenderGray32[color.Linear]{}

	original := dst
	BlendGray32Pixel(&dst, src, cover, blender)

	if math.Abs(float64(dst-original)) < 0.001 {
		t.Errorf("BlendGray32Pixel failed: no blending occurred")
	}
}

func TestCopyGray32Pixel(t *testing.T) {
	dst := float32(0.25)
	src := color.NewGray32[color.Linear](0.75)

	CopyGray32Pixel(&dst, src)

	if math.Abs(float64(dst-0.75)) > 0.001 {
		t.Errorf("CopyGray32Pixel failed: expected 0.75, got %f", dst)
	}
}

func TestBlendGray32Hline(t *testing.T) {
	dst := make([]float32, 10)
	for i := range dst {
		dst[i] = 0.25 // Initialize with some value
	}

	src := color.NewGray32WithAlpha[color.Linear](0.75, 0.5)
	blender := BlenderGray32[color.Linear]{}

	// Test without covers
	BlendGray32Hline(dst, 2, 5, src, nil, blender)

	// Check that pixels 2-6 were modified
	for i := 2; i < 7; i++ {
		if math.Abs(float64(dst[i]-0.25)) < 0.001 {
			t.Errorf("BlendGray32Hline failed: pixel %d was not blended", i)
		}
	}

	// Check that other pixels were not modified
	if math.Abs(float64(dst[0]-0.25)) > 0.001 ||
		math.Abs(float64(dst[1]-0.25)) > 0.001 ||
		math.Abs(float64(dst[7]-0.25)) > 0.001 {
		t.Errorf("BlendGray32Hline failed: unexpected pixels were modified")
	}
}

func TestBlendGray32HlineWithCovers(t *testing.T) {
	dst := make([]float32, 5)
	for i := range dst {
		dst[i] = 0.25
	}

	src := color.NewGray32WithAlpha[color.Linear](0.75, 1.0)
	covers := []float32{0.0, 0.5, 1.0, 0.25, 0.0}
	blender := BlenderGray32[color.Linear]{}

	BlendGray32Hline(dst, 0, 5, src, covers, blender)

	// Check that only pixels with non-zero coverage were modified
	if math.Abs(float64(dst[0]-0.25)) > 0.001 { // cover = 0.0, should not change
		t.Errorf("BlendGray32Hline with covers failed: pixel 0 should not have changed")
	}
	if math.Abs(float64(dst[1]-0.25)) < 0.001 { // cover = 0.5, should change
		t.Errorf("BlendGray32Hline with covers failed: pixel 1 should have changed")
	}
	if math.Abs(float64(dst[2]-0.25)) < 0.001 { // cover = 1.0, should change
		t.Errorf("BlendGray32Hline with covers failed: pixel 2 should have changed")
	}
	if math.Abs(float64(dst[4]-0.25)) > 0.001 { // cover = 0.0, should not change
		t.Errorf("BlendGray32Hline with covers failed: pixel 4 should not have changed")
	}
}

func TestCopyGray32Hline(t *testing.T) {
	dst := make([]float32, 10)
	src := color.NewGray32[color.Linear](0.75)

	CopyGray32Hline(dst, 2, 5, src)

	// Check that pixels 2-6 were set
	for i := 2; i < 7; i++ {
		if math.Abs(float64(dst[i]-0.75)) > 0.001 {
			t.Errorf("CopyGray32Hline failed: pixel %d = %f, expected 0.75", i, dst[i])
		}
	}

	// Check that other pixels remain zero
	if math.Abs(float64(dst[0])) > 0.001 ||
		math.Abs(float64(dst[1])) > 0.001 ||
		math.Abs(float64(dst[7])) > 0.001 {
		t.Errorf("CopyGray32Hline failed: unexpected pixels were modified")
	}
}

func TestBlenderGray32ZeroAlpha(t *testing.T) {
	dst := float32(0.5)
	src := color.NewGray32WithAlpha[color.Linear](0.75, 0.0) // Zero alpha
	cover := float32(1.0)
	blender := BlenderGray32[color.Linear]{}

	original := dst
	BlendGray32Pixel(&dst, src, cover, blender)

	// Should not blend when alpha is zero
	if math.Abs(float64(dst-original)) > 0.001 {
		t.Errorf("BlendGray32Pixel with zero alpha failed: should not blend, got %f", dst)
	}
}
