package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

func TestBlenderRGB(t *testing.T) {
	blender := BlenderRGB8[color.Linear, order.RGB]{}

	// Test BlendPix with full opacity
	dst := []basics.Int8u{100, 150, 200}       // RGB destination
	blender.BlendPix(dst, 255, 0, 0, 255, 255) // Blend red with full alpha and coverage

	// Should be close to red
	if dst[0] != 255 || dst[1] != 0 || dst[2] != 0 {
		t.Errorf("BlendPix full opacity failed: got [%d, %d, %d], want [255, 0, 0]", dst[0], dst[1], dst[2])
	}

	// Test BlendPix with half opacity
	dst = []basics.Int8u{100, 150, 200}
	blender.BlendPix(dst, 255, 0, 0, 128, 255) // Blend red with half alpha

	// Should be somewhere between original and red
	if dst[0] <= 100 {
		t.Errorf("BlendPix half opacity failed: red component %d should be greater than 100", dst[0])
	}
	if dst[1] >= 150 {
		t.Errorf("BlendPix half opacity failed: green component %d should be less than 150", dst[1])
	}
	if dst[2] >= 200 {
		t.Errorf("BlendPix half opacity failed: blue component %d should be less than 200", dst[2])
	}

	// Test BlendPix with zero alpha
	dst = []basics.Int8u{100, 150, 200}
	original := []basics.Int8u{100, 150, 200}
	blender.BlendPix(dst, 255, 0, 0, 0, 255) // Blend with zero alpha

	// Should remain unchanged
	if dst[0] != original[0] || dst[1] != original[1] || dst[2] != original[2] {
		t.Errorf("BlendPix zero alpha failed: got [%d, %d, %d], want [%d, %d, %d]",
			dst[0], dst[1], dst[2], original[0], original[1], original[2])
	}

	// Test BlendPix with half coverage
	dst = []basics.Int8u{100, 150, 200}
	blender.BlendPix(dst, 255, 0, 0, 255, 128) // Blend red with full alpha, half coverage

	// Coverage should reduce the blending effect
	if dst[0] <= 100 {
		t.Errorf("BlendPix half coverage failed: red component %d should be greater than 100", dst[0])
	}
}

func TestBlenderRGBPre(t *testing.T) {
	blender := BlenderRGBPre[color.Linear, order.RGB]{}

	// Test BlendPix with premultiplied colors
	dst := []basics.Int8u{100, 150, 200}
	blender.BlendPix(dst, 128, 0, 0, 128, 255) // Premultiplied red (50% alpha)

	// Should blend using premultiplied arithmetic
	if dst[0] <= 100 {
		t.Errorf("BlendPixPre failed: red component %d should be greater than 100", dst[0])
	}
	if dst[1] >= 150 {
		t.Errorf("BlendPixPre failed: green component %d should be less than 150", dst[1])
	}
	if dst[2] >= 200 {
		t.Errorf("BlendPixPre failed: blue component %d should be less than 200", dst[2])
	}

	// Test with zero alpha
	dst = []basics.Int8u{100, 150, 200}
	original := []basics.Int8u{100, 150, 200}
	blender.BlendPix(dst, 255, 0, 0, 0, 255)

	// Should remain unchanged
	if dst[0] != original[0] || dst[1] != original[1] || dst[2] != original[2] {
		t.Errorf("BlendPixPre zero alpha failed: got [%d, %d, %d], want [%d, %d, %d]",
			dst[0], dst[1], dst[2], original[0], original[1], original[2])
	}
}

func TestBlenderBGR(t *testing.T) {
	blender := BlenderRGB8[color.Linear, order.BGR]{}

	// Test BGR color order
	dst := []basics.Int8u{100, 150, 200}       // BGR destination (B=100, G=150, R=200)
	blender.BlendPix(dst, 255, 0, 0, 255, 255) // Blend red with full alpha

	// In BGR order, red goes to index 2
	if dst[2] != 255 || dst[1] != 0 || dst[0] != 0 {
		t.Errorf("BlenderBGR failed: got [%d, %d, %d], want [0, 0, 255] (BGR order)", dst[0], dst[1], dst[2])
	}
}

func TestRGBOrderIndices(t *testing.T) {
	var rgb order.RGB
	if rgb.IdxR() != 0 || rgb.IdxG() != 1 || rgb.IdxB() != 2 {
		t.Errorf("RGB indices wrong: R=%d G=%d B=%d", rgb.IdxR(), rgb.IdxG(), rgb.IdxB())
	}
	var bgr order.BGR
	if bgr.IdxR() != 2 || bgr.IdxG() != 1 || bgr.IdxB() != 0 {
		t.Errorf("BGR indices wrong: R=%d G=%d B=%d", bgr.IdxR(), bgr.IdxG(), bgr.IdxB())
	}
}

func TestBlendRGBPixel(t *testing.T) {
	dst := []basics.Int8u{100, 150, 200}
	src := color.RGB8[color.Linear]{R: 255, G: 0, B: 0}
	bl := BlenderRGB8[color.Linear, order.RGB]{}

	BlendRGBPixel[BlenderRGB8[color.Linear, order.RGB], color.Linear, order.RGB](dst, src, 255, 255, bl)

	// Should blend to red
	if dst[0] != 255 || dst[1] != 0 || dst[2] != 0 {
		t.Errorf("BlendRGBPixel failed: got [%d, %d, %d], want [255, 0, 0]", dst[0], dst[1], dst[2])
	}
}

func TestCopyRGBPixel(t *testing.T) {
	dst := []basics.Int8u{100, 150, 200}
	src := color.RGB8[color.Linear]{R: 255, G: 128, B: 64}

	CopyRGBPixel[color.Linear, order.RGB](dst, src)

	if dst[0] != 255 || dst[1] != 128 || dst[2] != 64 {
		t.Errorf("CopyRGBPixel failed: got [%d, %d, %d], want [255, 128, 64]", dst[0], dst[1], dst[2])
	}
}

func TestBlendRGBHline(t *testing.T) {
	// Create a buffer for 3 pixels (9 bytes total)
	dst := make([]basics.Int8u, 9)
	for i := range dst {
		dst[i] = 100 // Initialize with gray
	}

	src := color.RGB8[color.Linear]{R: 255, G: 0, B: 0}
	bl := BlenderRGB8[color.Linear, order.RGB]{}

	// Blend 3 pixels starting at x=0
	BlendRGBHline[BlenderRGB8[color.Linear, order.RGB], color.Linear, order.RGB](
		dst, 0, 3, src, 255, nil, bl,
	)

	// Check all pixels became red
	for i := 0; i < 3; i++ {
		offset := i * 3
		if dst[offset] != 255 || dst[offset+1] != 0 || dst[offset+2] != 0 {
			t.Errorf("BlendRGBHline pixel %d failed: got [%d, %d, %d], want [255, 0, 0]",
				i, dst[offset], dst[offset+1], dst[offset+2])
		}
	}

	// Test with varying coverage
	dst = make([]basics.Int8u, 9)
	for i := range dst {
		dst[i] = 100
	}

	covers := []basics.Int8u{255, 128, 64}
	BlendRGBHline[BlenderRGB8[color.Linear, order.RGB], color.Linear, order.RGB](
		dst, 0, 3, src, 255, covers, bl,
	)

	// First pixel should be fully red, others partially blended
	if dst[0] != 255 || dst[1] != 0 || dst[2] != 0 {
		t.Errorf("BlendRGBHline with coverage pixel 0 failed: got [%d, %d, %d], want [255, 0, 0]",
			dst[0], dst[1], dst[2])
	}

	// Second pixel should be partially blended (coverage=128)
	if dst[3] <= 100 {
		t.Errorf("BlendRGBHline with coverage pixel 1 red component should be greater than 100, got %d", dst[3])
	}
}

func TestCopyRGBHline(t *testing.T) {
	dst := make([]basics.Int8u, 9) // 3 pixels
	src := color.RGB8[color.Linear]{R: 255, G: 128, B: 64}

	CopyRGBHline[color.Linear, order.RGB](dst, 0, 3, src)

	// Check all pixels
	for i := 0; i < 3; i++ {
		offset := i * 3
		if dst[offset] != 255 || dst[offset+1] != 128 || dst[offset+2] != 64 {
			t.Errorf("CopyRGBHline pixel %d failed: got [%d, %d, %d], want [255, 128, 64]",
				i, dst[offset], dst[offset+1], dst[offset+2])
		}
	}
}

func TestConvertRGBAToRGB(t *testing.T) {
	rgba := color.RGBA8[color.Linear]{R: 255, G: 128, B: 64, A: 192}
	rgb := ConvertRGBAToRGB(rgba)

	if rgb.R != 255 || rgb.G != 128 || rgb.B != 64 {
		t.Errorf("ConvertRGBAToRGB failed: got {%d, %d, %d}, want {255, 128, 64}", rgb.R, rgb.G, rgb.B)
	}
}

func TestConvertRGBToRGBA(t *testing.T) {
	rgb := color.RGB8[color.Linear]{R: 255, G: 128, B: 64}
	rgba := ConvertRGBToRGBA(rgb)

	if rgba.R != 255 || rgba.G != 128 || rgba.B != 64 || rgba.A != 255 {
		t.Errorf("ConvertRGBToRGBA failed: got {%d, %d, %d, %d}, want {255, 128, 64, 255}",
			rgba.R, rgba.G, rgba.B, rgba.A)
	}
}

// Benchmarks
func BenchmarkBlenderRGB24(b *testing.B) {
	bl := BlenderRGB8[color.Linear, order.RGB]{}
	dst := []basics.Int8u{100, 150, 200}
	for i := 0; i < b.N; i++ {
		bl.BlendPix(dst, 255, 0, 0, 255, 255)
	}
}

func BenchmarkBlenderRGB24Pre(b *testing.B) {
	bl := BlenderRGBPre[color.Linear, order.RGB]{}
	dst := []basics.Int8u{100, 150, 200}
	for i := 0; i < b.N; i++ {
		bl.BlendPix(dst, 128, 0, 0, 128, 255)
	}
}

func BenchmarkBlendRGBHline(b *testing.B) {
	dst := make([]basics.Int8u, 300) // 100 pixels
	src := color.RGB8[color.Linear]{R: 255, G: 0, B: 0}
	bl := BlenderRGB8[color.Linear, order.RGB]{}
	for i := 0; i < b.N; i++ {
		BlendRGBHline[BlenderRGB8[color.Linear, order.RGB], color.Linear, order.RGB](dst, 0, 100, src, 255, nil, bl)
	}
}

func BenchmarkCopyRGBHline(b *testing.B) {
	dst := make([]basics.Int8u, 300) // 100 pixels
	src := color.RGB8[color.Linear]{R: 255, G: 0, B: 0}
	for i := 0; i < b.N; i++ {
		CopyRGBHline[color.Linear, order.RGB](dst, 0, 100, src)
	}
}
