package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestBlenderGray16BlendPix(t *testing.T) {
	blender := BlenderGray16Linear{}
	dst := basics.Int16u(0x4000)   // ~0.25 in 16-bit
	cv := basics.Int16u(0xC000)    // ~0.75 in 16-bit
	alpha := basics.Int16u(0x8000) // 0.5 alpha
	cover := basics.Int16u(0xFFFF) // Full coverage

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be between original dst and cv
	if dst <= 0x4000 || dst >= 0xC000 {
		t.Errorf("BlendPix failed: expected blended value between 0x4000-0xC000, got 0x%X", dst)
	}
}

func TestBlenderGray16PreBlendPix(t *testing.T) {
	blender := BlenderGray16PreLinear{}
	dst := basics.Int16u(0x4000)
	cv := basics.Int16u(0xC000)
	alpha := basics.Int16u(0x8000)
	cover := basics.Int16u(0xFFFF)

	blender.BlendPix(&dst, cv, alpha, cover)

	// Result should be different from dst (some blending should occur)
	if dst == 0x4000 {
		t.Errorf("BlendPix failed: no blending occurred, result still 0x%X", dst)
	}
}

func TestGray16Multiply(t *testing.T) {
	// Test basic multiplication
	a := basics.Int16u(0x8000) // 0.5
	b := basics.Int16u(0x8000) // 0.5
	result := Gray16Multiply(a, b)
	expected := basics.Int16u(0x4000) // 0.25

	// Allow some tolerance due to fixed-point arithmetic
	if result < expected-100 || result > expected+100 {
		t.Errorf("Gray16Multiply failed: expected ~0x%X, got 0x%X", expected, result)
	}
}

func TestGray16Lerp(t *testing.T) {
	p := basics.Int16u(0x2000) // 0.125
	q := basics.Int16u(0xE000) // 0.875
	a := basics.Int16u(0x8000) // 0.5 (halfway)
	result := Gray16Lerp(p, q, a)
	expected := basics.Int16u(0x8000) // Should be halfway between p and q

	// Allow some tolerance
	if result < expected-100 || result > expected+100 {
		t.Errorf("Gray16Lerp failed: expected ~0x%X, got 0x%X", expected, result)
	}
}

func TestGray16Prelerp(t *testing.T) {
	p := basics.Int16u(0x8000)
	q := basics.Int16u(0x4000)
	a := basics.Int16u(0x8000)
	result := Gray16Prelerp(p, q, a)

	// Prelerp formula: p + q - multiply(p, a)
	// With these values: 0x8000 + 0x4000 - multiply(0x8000, 0x8000)
	// = 0x8000 + 0x4000 - 0x4000 = 0x8000
	// So result equals p, which is expected behavior
	expected := basics.Int16u(0x8000)
	if result != expected {
		t.Errorf("Gray16Prelerp failed: expected 0x%X, got 0x%X", expected, result)
	}
}

func TestBlendGray16Pixel(t *testing.T) {
	dst := basics.Int16u(0x4000)
	src := color.NewGray16WithAlpha[color.Linear](0xC000, 0x8000)
	cover := basics.Int16u(0xFFFF)
	blender := BlenderGray16[color.Linear]{}

	original := dst
	BlendGray16Pixel(&dst, src, cover, blender)

	if dst == original {
		t.Errorf("BlendGray16Pixel failed: no blending occurred")
	}
}

func TestCopyGray16Pixel(t *testing.T) {
	dst := basics.Int16u(0x4000)
	src := color.NewGray16[color.Linear](0xC000)

	CopyGray16Pixel(&dst, src)

	if dst != 0xC000 {
		t.Errorf("CopyGray16Pixel failed: expected 0xC000, got 0x%X", dst)
	}
}

func TestBlendGray16Hline(t *testing.T) {
	dst := make([]basics.Int16u, 10)
	for i := range dst {
		dst[i] = 0x4000 // Initialize with some value
	}

	src := color.NewGray16WithAlpha[color.Linear](0xC000, 0x8000)
	blender := BlenderGray16[color.Linear]{}

	// Test without covers
	BlendGray16Hline(dst, 2, 5, src, nil, blender)

	// Check that pixels 2-6 were modified
	for i := 2; i < 7; i++ {
		if dst[i] == 0x4000 {
			t.Errorf("BlendGray16Hline failed: pixel %d was not blended", i)
		}
	}

	// Check that other pixels were not modified
	if dst[0] != 0x4000 || dst[1] != 0x4000 || dst[7] != 0x4000 {
		t.Errorf("BlendGray16Hline failed: unexpected pixels were modified")
	}
}

func TestCopyGray16Hline(t *testing.T) {
	dst := make([]basics.Int16u, 10)
	src := color.NewGray16[color.Linear](0xC000)

	CopyGray16Hline(dst, 2, 5, src)

	// Check that pixels 2-6 were set
	for i := 2; i < 7; i++ {
		if dst[i] != 0xC000 {
			t.Errorf("CopyGray16Hline failed: pixel %d = 0x%X, expected 0xC000", i, dst[i])
		}
	}

	// Check that other pixels remain zero
	if dst[0] != 0 || dst[1] != 0 || dst[7] != 0 {
		t.Errorf("CopyGray16Hline failed: unexpected pixels were modified")
	}
}
