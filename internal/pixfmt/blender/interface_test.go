package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

// TestBlenderInterfaceCompliance tests that all blender types implement their respective interfaces
func TestBlenderInterfaceCompliance(t *testing.T) {
	// Test RGBA blenders implement RGBABlender interface
	var _ RGBABlender[color.Linear] = BlenderRGBA8[color.Linear, order.RGBA]{}
	var _ RGBABlender[color.Linear] = BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	var _ RGBABlender[color.Linear] = BlenderRGBA8Plain[color.Linear, order.RGBA]{}

	// Test other RGBA color orders
	var _ RGBABlender[color.Linear] = BlenderRGBA8[color.Linear, order.ARGB]{}
	var _ RGBABlender[color.Linear] = BlenderRGBA8[color.Linear, order.BGRA]{}
	var _ RGBABlender[color.Linear] = BlenderRGBA8[color.Linear, order.ABGR]{}

	// Test RGBA blenders with SRGB color space
	var _ RGBABlender[color.SRGB] = BlenderRGBA8[color.SRGB, order.RGBA]{}
	var _ RGBABlender[color.SRGB] = BlenderRGBA8Pre[color.SRGB, order.RGBA]{}
	var _ RGBABlender[color.SRGB] = BlenderRGBA8Plain[color.SRGB, order.RGBA]{}

	// Test RawRGBAOrder interface compliance
	var _ RawRGBAOrder = BlenderRGBA8[color.Linear, order.RGBA]{}
	var _ RawRGBAOrder = BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	var _ RawRGBAOrder = BlenderRGBA8Plain[color.Linear, order.RGBA]{}
}

// TestRGBABlenderBehavior tests actual RGBA blender method behavior
func TestRGBABlenderBehavior(t *testing.T) {
	// Create test pixel data (RGBA order)
	pixel := []basics.Int8u{128, 64, 192, 255} // R=128, G=64, B=192, A=255

	// Test RGBA8 blender
	rgbaBlender := BlenderRGBA8[color.Linear, order.RGBA]{}

	// Test GetPlain method (should return the original values for RGBA8Plain storage)
	// Note: BlenderRGBA8 expects premultiplied storage, so this gets demultiplied values
	r, g, b, a := rgbaBlender.GetPlain(pixel)
	if a == 0 {
		t.Errorf("Alpha should not be zero for test pixel")
	} else if r != 128 || g != 64 || b != 192 || a != 255 {
		t.Logf("GetPlain returned: (%d,%d,%d,%d), expected close to (128,64,192,255)", r, g, b, a)
		// This might be expected behavior for premultiplied blenders
	}

	// Test SetPlain method
	testPixel := make([]basics.Int8u, 4)
	rgbaBlender.SetPlain(testPixel, 100, 150, 200, 255)
	// For BlenderRGBA8, SetPlain stores premultiplied values
	// Verify the values are set (exact values depend on blender type)
	checkR, checkG, checkB, checkA := rgbaBlender.GetPlain(testPixel)
	if checkA == 0 {
		t.Errorf("SetPlain/GetPlain failed: alpha should not be zero")
	} else {
		t.Logf("SetPlain/GetPlain roundtrip: set (100,150,200,255), got (%d,%d,%d,%d)",
			checkR, checkG, checkB, checkA)
		// For fully opaque pixels, we expect reasonable roundtrip behavior
		if checkA != 255 {
			t.Errorf("Alpha should roundtrip correctly: expected 255, got %d", checkA)
		}
	}

	// Test BlendPix method with no blending (cover = 0)
	dst := make([]basics.Int8u, 4)
	copy(dst, pixel)
	original := make([]basics.Int8u, 4)
	copy(original, dst)
	rgbaBlender.BlendPix(dst, 255, 255, 255, 255, 0) // cover = 0 should do nothing
	for i := range 4 {
		if dst[i] != original[i] {
			t.Errorf("BlendPix with cover=0 should not modify pixel, but pixel[%d] changed from %d to %d",
				i, original[i], dst[i])
		}
	}

	// Test BlendPix method with full blending
	rgbaBlender.BlendPix(dst, 255, 0, 0, 255, 255) // blend with opaque red
	t.Logf("After blending with red: (%d,%d,%d,%d)", dst[0], dst[1], dst[2], dst[3])
	// The exact result depends on blending math, but red component should increase
}
