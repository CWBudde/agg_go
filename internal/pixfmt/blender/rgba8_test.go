package blender

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

func TestBlenderRGBA(t *testing.T) {
	bl := BlenderRGBA8[color.Linear, order.RGBA]{}

	// Gray background in RGBA layout
	dst := []basics.Int8u{100, 100, 100, 255}

	// Source ~50% alpha
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128)
	cover := basics.Int8u(255)

	bl.BlendPix(dst, r, g, b, a, cover)

	// Result should move toward source
	if dst[0] <= 100 || dst[0] >= 200 {
		t.Errorf("BlendPix red result %d should be between 100 and 200", dst[0])
	}
	if dst[1] <= 100 || dst[1] >= 150 {
		t.Errorf("BlendPix green result %d should be between 100 and 150", dst[1])
	}
	// Blue 100 -> 50 at ~50% alpha ≈ 75; allow some tolerance
	if dst[2] < 70 || dst[2] > 90 {
		t.Errorf("BlendPix blue result %d should be in [70,90]", dst[2])
	}
}

func TestBlenderRGBAPre(t *testing.T) {
	bl := BlenderRGBA8Pre[color.Linear, order.RGBA]{}

	dst := []basics.Int8u{100, 100, 100, 255}
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128)
	cover := basics.Int8u(255)

	orig := append([]basics.Int8u(nil), dst...)
	bl.BlendPix(dst, r, g, b, a, cover)

	changed := false
	for i := 0; i < 4; i++ {
		if dst[i] != orig[i] {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("BlendPix should have modified the destination")
	}
}

func TestBlenderRGBAPlain(t *testing.T) {
	bl := BlenderRGBA8Plain[color.Linear, order.RGBA]{}

	dst := []basics.Int8u{100, 100, 100, 200}
	r, g, b, a := basics.Int8u(200), basics.Int8u(150), basics.Int8u(50), basics.Int8u(128)
	cover := basics.Int8u(255)

	bl.BlendPix(dst, r, g, b, a, cover)

	if dst[0] == 100 && dst[1] == 100 && dst[2] == 100 {
		t.Error("BlendPix should have changed the destination values")
	}
}

func TestBlendRGBAPixel(t *testing.T) {
	dst := []basics.Int8u{50, 50, 50, 255}
	src := color.NewRGBA8[color.Linear](150, 200, 100, 200)
	cover := basics.Int8u(255)

	var bl BlenderRGBA8[color.Linear, order.RGBA]
	BlendRGBAPixel[color.Linear, order.RGBA](dst, src, cover, bl)

	if dst[0] <= 50 || dst[1] <= 50 {
		t.Error("BlendRGBAPixel should blend towards brighter colors")
	}
}

func TestBlendRGBAPixelZeroAlpha(t *testing.T) {
	orig := []basics.Int8u{100, 100, 100, 255}
	dst := append([]basics.Int8u(nil), orig...)

	src := color.NewRGBA8[color.Linear](200, 200, 200, 0) // Zero alpha
	cover := basics.Int8u(255)

	var bl BlenderRGBA8[color.Linear, order.RGBA]
	BlendRGBAPixel[color.Linear, order.RGBA](dst, src, cover, bl)

	for i := 0; i < 4; i++ {
		if dst[i] != orig[i] {
			t.Errorf("Zero-alpha blend should not change dst[%d]: expected %d, got %d",
				i, orig[i], dst[i])
		}
	}
}

func TestCopyRGBAPixel(t *testing.T) {
	dst := []basics.Int8u{0, 0, 0, 0}
	src := color.NewRGBA8[color.Linear](100, 150, 200, 255)

	// Provide S and O explicitly (S inferred from src too)
	CopyRGBAPixel[color.Linear, order.RGBA](dst, src)

	if dst[0] != 100 || dst[1] != 150 || dst[2] != 200 || dst[3] != 255 {
		t.Errorf("CopyRGBAPixel failed: got (%d,%d,%d,%d), expected (100,150,200,255)",
			dst[0], dst[1], dst[2], dst[3])
	}
}

func TestBlendRGBAHline(t *testing.T) {
	// 10 pixels row in RGBA layout
	dst := make([]basics.Int8u, 40)
	for i := range dst {
		dst[i] = 50
	}
	src := color.NewRGBA8[color.Linear](150, 200, 100, 200)

	var bl BlenderRGBA8[color.Linear, order.RGBA]

	x := 2
	length := 5
	BlendRGBAHline[color.Linear, order.RGBA](dst, x, length, src, nil, bl)

	// Affected pixels [2..6)
	for i := x; i < x+length; i++ {
		o := i * 4
		if dst[o] <= 50 || dst[o+1] <= 50 {
			t.Errorf("BlendRGBAHline pixel %d should be brighter than background", i)
		}
	}
	// Unaffected pixels (spot-check)
	if dst[0] != 50 || dst[4] != 50 || dst[28] != 50 {
		t.Error("BlendRGBAHline should not affect pixels outside the range")
	}
}

func TestBlendRGBAHlineWithCovers(t *testing.T) {
	dst := make([]basics.Int8u, 40)
	for i := range dst {
		dst[i] = 50
	}
	src := color.NewRGBA8[color.Linear](200, 200, 200, 255)
	covers := []basics.Int8u{255, 200, 100, 50, 0}

	var bl BlenderRGBA8[color.Linear, order.RGBA]

	x := 2
	length := len(covers)
	BlendRGBAHline[color.Linear, order.RGBA](dst, x, length, src, covers, bl)

	p0 := dst[x*4]     // full cover
	p1 := dst[(x+1)*4] // partial cover
	p4 := dst[(x+4)*4] // zero cover

	if p0 <= p1 {
		t.Error("Higher coverage should result in values closer to source")
	}
	if p4 != 50 {
		t.Errorf("Zero coverage should leave pixel unchanged: expected 50, got %d", p4)
	}
}

func TestCopyRGBAHline(t *testing.T) {
	dst := make([]basics.Int8u, 40)
	src := color.NewRGBA8[color.Linear](100, 150, 200, 255)

	x := 2
	length := 5
	CopyRGBAHline[color.Linear, order.RGBA](dst, x, length, src)

	for i := x; i < x+length; i++ {
		o := i * 4
		if dst[o] != 100 || dst[o+1] != 150 || dst[o+2] != 200 || dst[o+3] != 255 {
			t.Errorf("CopyRGBAHline pixel %d should match source color", i)
		}
	}
}

func TestDifferentOrdersCompile(t *testing.T) {
	// This test intentionally just ensures different instantiations compile.
	_ = BlenderRGBA8[color.Linear, order.RGBA]{}
	_ = BlenderRGBA8[color.Linear, order.BGRA]{}
	_ = BlenderRGBA8[color.Linear, order.ARGB]{}
	_ = BlenderRGBA8[color.Linear, order.ABGR]{}

	_ = BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	_ = BlenderRGBA8Plain[color.Linear, order.RGBA]{}
}
