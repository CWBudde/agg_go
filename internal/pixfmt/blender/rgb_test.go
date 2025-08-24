package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestBlenderRGB(t *testing.T) {
	blender := BlenderRGB24{}

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
	if dst[0] <= 100 || dst[0] >= 255 {
		t.Errorf("BlendPix half opacity failed: red component %d should be between 100 and 255", dst[0])
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
	if dst[0] <= 100 || dst[0] >= 255 {
		t.Errorf("BlendPix half coverage failed: red component %d should be between 100 and 255", dst[0])
	}
}

func TestBlenderRGBPre(t *testing.T) {
	blender := BlenderRGB24Pre{}

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
	blender := BlenderBGR24{}

	// Test BGR color order
	dst := []basics.Int8u{100, 150, 200}       // BGR destination (B=100, G=150, R=200)
	blender.BlendPix(dst, 255, 0, 0, 255, 255) // Blend red with full alpha

	// In BGR order, red goes to index 2
	if dst[2] != 255 || dst[1] != 0 || dst[0] != 0 {
		t.Errorf("BlenderBGR failed: got [%d, %d, %d], want [0, 0, 255] (BGR order)", dst[0], dst[1], dst[2])
	}
}

func TestGetRGBColorOrder(t *testing.T) {
	// Test RGB order
	order := GetRGBColorOrder[color.RGB24Order]()
	if order.R != 0 || order.G != 1 || order.B != 2 {
		t.Errorf("RGB order failed: got R=%d, G=%d, B=%d, want R=0, G=1, B=2", order.R, order.G, order.B)
	}

	// Test BGR order
	order = GetRGBColorOrder[color.BGR24Order]()
	if order.R != 2 || order.G != 1 || order.B != 0 {
		t.Errorf("BGR order failed: got R=%d, G=%d, B=%d, want R=2, G=1, B=0", order.R, order.G, order.B)
	}
}

func TestBlendRGBPixel(t *testing.T) {
	dst := []basics.Int8u{100, 150, 200}
	src := color.RGB8Linear{R: 255, G: 0, B: 0}
	blender := BlenderRGB24{}

	BlendRGBPixel(dst, src, 255, 255, blender)

	// Should blend to red
	if dst[0] != 255 || dst[1] != 0 || dst[2] != 0 {
		t.Errorf("BlendRGBPixel failed: got [%d, %d, %d], want [255, 0, 0]", dst[0], dst[1], dst[2])
	}
}

func TestCopyRGBPixel(t *testing.T) {
	dst := []basics.Int8u{100, 150, 200}
	src := color.RGB8Linear{R: 255, G: 128, B: 64}

	CopyRGBPixel[color.RGB24Order](dst, src)

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

	src := color.RGB8Linear{R: 255, G: 0, B: 0}
	blender := BlenderRGB24{}

	// Blend 3 pixels starting at x=0
	BlendRGBHline(dst, 0, 3, src, 255, nil, blender)

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
	BlendRGBHline(dst, 0, 3, src, 255, covers, blender)

	// First pixel should be fully red, others partially blended
	if dst[0] != 255 || dst[1] != 0 || dst[2] != 0 {
		t.Errorf("BlendRGBHline with coverage pixel 0 failed: got [%d, %d, %d], want [255, 0, 0]",
			dst[0], dst[1], dst[2])
	}

	// Second pixel should be partially blended (coverage=128)
	if dst[3] <= 100 || dst[3] >= 255 {
		t.Errorf("BlendRGBHline with coverage pixel 1 red component should be between 100 and 255, got %d", dst[3])
	}
}

func TestCopyRGBHline(t *testing.T) {
	dst := make([]basics.Int8u, 9) // 3 pixels
	src := color.RGB8Linear{R: 255, G: 128, B: 64}

	CopyRGBHline[color.RGB24Order](dst, 0, 3, src)

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
	rgba := color.RGBA8Linear{R: 255, G: 128, B: 64, A: 192}
	rgb := ConvertRGBAToRGB(rgba)

	if rgb.R != 255 || rgb.G != 128 || rgb.B != 64 {
		t.Errorf("ConvertRGBAToRGB failed: got {%d, %d, %d}, want {255, 128, 64}", rgb.R, rgb.G, rgb.B)
	}
}

func TestConvertRGBToRGBA(t *testing.T) {
	rgb := color.RGB8Linear{R: 255, G: 128, B: 64}
	rgba := ConvertRGBToRGBA(rgb)

	if rgba.R != 255 || rgba.G != 128 || rgba.B != 64 || rgba.A != 255 {
		t.Errorf("ConvertRGBToRGBA failed: got {%d, %d, %d, %d}, want {255, 128, 64, 255}",
			rgba.R, rgba.G, rgba.B, rgba.A)
	}
}

// Test that blenders implement the interface
func TestBlenderInterfaces(t *testing.T) {
	var _ RGBBlender = BlenderRGB24{}
	var _ RGBBlender = BlenderRGB24Pre{}
	var _ RGBBlender = BlenderBGR24{}
	var _ RGBBlender = BlenderBGR24Pre{}
}

// Benchmark tests
func BenchmarkBlenderRGB24(b *testing.B) {
	blender := BlenderRGB24{}
	dst := []basics.Int8u{100, 150, 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blender.BlendPix(dst, 255, 0, 0, 255, 255)
	}
}

func BenchmarkBlenderRGB24Pre(b *testing.B) {
	blender := BlenderRGB24Pre{}
	dst := []basics.Int8u{100, 150, 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blender.BlendPix(dst, 128, 0, 0, 128, 255)
	}
}

func BenchmarkBlendRGBHline(b *testing.B) {
	dst := make([]basics.Int8u, 300) // 100 pixels
	src := color.RGB8Linear{R: 255, G: 0, B: 0}
	blender := BlenderRGB24{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BlendRGBHline(dst, 0, 100, src, 255, nil, blender)
	}
}

func BenchmarkCopyRGBHline(b *testing.B) {
	dst := make([]basics.Int8u, 300) // 100 pixels
	src := color.RGB8Linear{R: 255, G: 0, B: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CopyRGBHline[color.RGB24Order](dst, 0, 100, src)
	}
}
