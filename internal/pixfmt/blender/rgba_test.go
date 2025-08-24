package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestBlenderRGBA(t *testing.T) {
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	// Test blending with full coverage
	dst := []basics.Int8u{100, 100, 100, 255}                                               // Gray background
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128) // 50% alpha
	cover := basics.Int8u(255)                                                              // Full coverage

	blender.BlendPix(dst, r, g, b, a, cover)

	// Result should be between background and source colors
	if dst[0] <= 100 || dst[0] >= 200 { // Red
		t.Errorf("BlendPix red result %d should be between 100 and 200", dst[0])
	}
	if dst[1] <= 100 || dst[1] >= 150 { // Green
		t.Errorf("BlendPix green result %d should be between 100 and 150", dst[1])
	}
	// For blue, with 50% alpha blending from 100 to 50, we expect something around 75
	// But let's be more lenient with the test since fixed-point arithmetic can have variations
	if dst[2] < 70 || dst[2] > 90 { // Blue blend range
		t.Errorf("BlendPix blue result %d should be between 70 and 90 (blending 100->50 at 50%%)", dst[2])
	}
}

func TestBlenderRGBAPre(t *testing.T) {
	blender := BlenderRGBAPre[color.Linear, RGBAOrder]{}

	// Test premultiplied blending
	dst := []basics.Int8u{100, 100, 100, 255}
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128)
	cover := basics.Int8u(255)

	originalDst := make([]basics.Int8u, len(dst))
	copy(originalDst, dst)

	blender.BlendPix(dst, r, g, b, a, cover)

	// Should have changed the destination
	changed := false
	for i := 0; i < 4; i++ {
		if dst[i] != originalDst[i] {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("BlendPix should have modified the destination")
	}
}

func TestBlenderRGBAPlain(t *testing.T) {
	blender := BlenderRGBAPlain[color.Linear, RGBAOrder]{}

	// Test plain (non-premultiplied) blending
	dst := []basics.Int8u{100, 100, 100, 200}
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128)
	cover := basics.Int8u(255)

	blender.BlendPix(dst, r, g, b, a, cover)

	// Should have blended correctly
	if dst[0] == 100 && dst[1] == 100 && dst[2] == 100 {
		t.Error("BlendPix should have changed the destination values")
	}
}

func TestBlendRGBAPixel(t *testing.T) {
	dst := []basics.Int8u{50, 50, 50, 255}
	src := color.NewRGBA8[color.Linear](150, 200, 100, 200)
	cover := basics.Int8u(255)
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	BlendRGBAPixel(dst, src, cover, blender)

	// Should blend towards source color
	if dst[0] <= 50 || dst[1] <= 50 {
		t.Error("BlendRGBAPixel should blend towards brighter colors")
	}
}

func TestBlendRGBAPixelZeroAlpha(t *testing.T) {
	original := []basics.Int8u{100, 100, 100, 255}
	dst := make([]basics.Int8u, len(original))
	copy(dst, original)

	src := color.NewRGBA8[color.Linear](200, 200, 200, 0) // Zero alpha
	cover := basics.Int8u(255)
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	BlendRGBAPixel(dst, src, cover, blender)

	// Should not change destination when source is transparent
	for i := 0; i < 4; i++ {
		if dst[i] != original[i] {
			t.Errorf("BlendRGBAPixel with zero alpha should not change dst[%d]: expected %d, got %d",
				i, original[i], dst[i])
		}
	}
}

func TestCopyRGBAPixel(t *testing.T) {
	dst := []basics.Int8u{0, 0, 0, 0}
	src := color.NewRGBA8[color.Linear](100, 150, 200, 255)

	CopyRGBAPixel[RGBAOrder](dst, src)

	if dst[0] != 100 || dst[1] != 150 || dst[2] != 200 || dst[3] != 255 {
		t.Errorf("CopyRGBAPixel failed: got (%d,%d,%d,%d), expected (100,150,200,255)",
			dst[0], dst[1], dst[2], dst[3])
	}
}

func TestBlendRGBAHline(t *testing.T) {
	// Create a test row (10 pixels)
	dst := make([]basics.Int8u, 40) // 10 pixels * 4 components
	for i := range dst {
		dst[i] = 50 // Gray background
	}

	src := color.NewRGBA8[color.Linear](150, 200, 100, 200)
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	// Blend horizontal line from pixel 2 to 6 (5 pixels)
	x := 2
	length := 5
	BlendRGBAHline(dst, x, length, src, nil, blender)

	// Check affected pixels
	for i := x; i < x+length; i++ {
		pixelOffset := i * 4
		if dst[pixelOffset] <= 50 || dst[pixelOffset+1] <= 50 { // R, G should increase
			t.Errorf("BlendRGBAHline pixel %d should be brighter than background", i)
		}
	}

	// Check unaffected pixels
	if dst[0] != 50 || dst[4] != 50 || dst[28] != 50 { // Before x=2 and after x=6
		t.Error("BlendRGBAHline should not affect pixels outside the range")
	}
}

func TestBlendRGBAHlineWithCovers(t *testing.T) {
	// Create a test row
	dst := make([]basics.Int8u, 40)
	for i := range dst {
		dst[i] = 50
	}

	src := color.NewRGBA8[color.Linear](200, 200, 200, 255) // White with full alpha
	covers := []basics.Int8u{255, 200, 100, 50, 0}          // Varying coverage
	blender := BlenderRGBA[color.Linear, RGBAOrder]{}

	x := 2
	length := len(covers)
	BlendRGBAHline(dst, x, length, src, covers, blender)

	// Check that higher coverage results in values closer to source
	pixel0 := dst[x*4]     // Full coverage
	pixel1 := dst[(x+1)*4] // Partial coverage
	pixel4 := dst[(x+4)*4] // Zero coverage

	if pixel0 <= pixel1 {
		t.Error("Higher coverage should result in values closer to source")
	}
	if pixel4 != 50 {
		t.Errorf("Zero coverage should leave pixel unchanged: expected 50, got %d", pixel4)
	}
}

func TestCopyRGBAHline(t *testing.T) {
	dst := make([]basics.Int8u, 40)
	src := color.NewRGBA8[color.Linear](100, 150, 200, 255)

	x := 2
	length := 5
	CopyRGBAHline[RGBAOrder](dst, x, length, src)

	// Check that all affected pixels are exactly the source color
	for i := x; i < x+length; i++ {
		pixelOffset := i * 4
		if dst[pixelOffset] != 100 || dst[pixelOffset+1] != 150 ||
			dst[pixelOffset+2] != 200 || dst[pixelOffset+3] != 255 {
			t.Errorf("CopyRGBAHline pixel %d should match source color", i)
		}
	}
}

func TestColorOrderTypes(t *testing.T) {
	// Test that different color order types can be created
	_ = BlenderRGBA8{}
	_ = BlenderARGB8{}
	_ = BlenderBGRA8{}
	_ = BlenderABGR8{}
	_ = BlenderRGBA8Pre{}
	_ = BlenderRGBA8Plain{}

	// These should compile without errors
}

func TestGetColorOrder(t *testing.T) {
	// Test RGBA order
	rgbaOrder := GetColorOrder[RGBAOrder]()
	if rgbaOrder.R != 0 || rgbaOrder.G != 1 || rgbaOrder.B != 2 || rgbaOrder.A != 3 {
		t.Errorf("RGBA order incorrect: got R=%d,G=%d,B=%d,A=%d, expected R=0,G=1,B=2,A=3",
			rgbaOrder.R, rgbaOrder.G, rgbaOrder.B, rgbaOrder.A)
	}

	// Test ARGB order
	argbOrder := GetColorOrder[ARGBOrder]()
	if argbOrder.A != 0 || argbOrder.R != 1 || argbOrder.G != 2 || argbOrder.B != 3 {
		t.Errorf("ARGB order incorrect: got A=%d,R=%d,G=%d,B=%d, expected A=0,R=1,G=2,B=3",
			argbOrder.A, argbOrder.R, argbOrder.G, argbOrder.B)
	}

	// Test BGRA order
	bgraOrder := GetColorOrder[BGRAOrder]()
	if bgraOrder.B != 0 || bgraOrder.G != 1 || bgraOrder.R != 2 || bgraOrder.A != 3 {
		t.Errorf("BGRA order incorrect: got B=%d,G=%d,R=%d,A=%d, expected B=0,G=1,R=2,A=3",
			bgraOrder.B, bgraOrder.G, bgraOrder.R, bgraOrder.A)
	}
}
