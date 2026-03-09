package blender

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
)

// ──────────────────────────────────────────────────────────────────────────
// Gray8 — Pre variants & SetPlain/GetPlain
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderGray8SetGetPlain(t *testing.T) {
	b := BlenderGray8[color.Linear]{}
	var pix basics.Int8u = 42
	b.SetPlain(&pix, 200, 128)
	if pix != 200 {
		t.Errorf("Gray8 SetPlain: pix = %d, want 200", pix)
	}
	v, a := b.GetPlain(&pix)
	if v != 200 || a != 255 {
		t.Errorf("Gray8 GetPlain: v=%d a=%d, want 200 255", v, a)
	}
}

func TestBlenderGray8PreSetGetPlain(t *testing.T) {
	b := BlenderGray8Pre[color.Linear]{}
	var pix basics.Int8u = 0
	// SetPlain should premultiply: Gray8Multiply(200, 128) = 100..101
	b.SetPlain(&pix, 200, 128)
	if pix == 0 || pix > 101 {
		t.Errorf("Gray8Pre SetPlain: pix = %d, expected ~100", pix)
	}
	_, a := b.GetPlain(&pix)
	if a != 255 {
		t.Errorf("Gray8Pre GetPlain: a = %d, want 255", a)
	}
}

func TestBlendGrayPixelPre(t *testing.T) {
	var dst basics.Int8u = 0
	src := color.NewGray8WithAlpha[color.Linear](200, 255)
	cover := basics.Int8u(255)
	b := BlenderGray8Pre[color.Linear]{}
	BlendGrayPixelPre(&dst, src, cover, b)
	if dst == 0 {
		t.Error("BlendGrayPixelPre: dst should be non-zero after blending opaque src onto black")
	}
}

func TestBlendGrayPixelPreZeroAlpha(t *testing.T) {
	var dst basics.Int8u = 100
	src := color.NewGray8WithAlpha[color.Linear](200, 0) // alpha=0
	cover := basics.Int8u(255)
	b := BlenderGray8Pre[color.Linear]{}
	BlendGrayPixelPre(&dst, src, cover, b)
	if dst != 100 {
		t.Errorf("BlendGrayPixelPre with zero alpha: dst = %d, want 100 (unchanged)", dst)
	}
}

func TestBlendGrayHlinePre(t *testing.T) {
	dst := make([]basics.Int8u, 6)
	for i := range dst {
		dst[i] = 50
	}
	src := color.NewGray8WithAlpha[color.Linear](200, 255)
	b := BlenderGray8Pre[color.Linear]{}

	// nil covers → uniform full coverage
	BlendGrayHlinePre(dst, 1, 4, src, nil, b)
	for i := 1; i <= 4; i++ {
		if dst[i] <= 50 {
			t.Errorf("pixel %d should be > 50, got %d", i, dst[i])
		}
	}
	if dst[0] != 50 || dst[5] != 50 {
		t.Error("pixels outside range should be unchanged")
	}
}

func TestBlendGrayHlinePreWithCovers(t *testing.T) {
	dst := make([]basics.Int8u, 5)
	for i := range dst {
		dst[i] = 0
	}
	src := color.NewGray8WithAlpha[color.Linear](255, 255)
	covers := []basics.Int8u{255, 128, 0}
	b := BlenderGray8Pre[color.Linear]{}

	BlendGrayHlinePre(dst, 0, 3, src, covers, b)
	if dst[0] <= dst[1] {
		t.Error("full coverage should give brighter result than half coverage")
	}
	if dst[2] != 0 {
		t.Errorf("zero coverage pixel should be 0, got %d", dst[2])
	}
}

func TestBlendGrayHlinePreZeroAlpha(t *testing.T) {
	dst := make([]basics.Int8u, 4)
	for i := range dst {
		dst[i] = 77
	}
	src := color.NewGray8WithAlpha[color.Linear](200, 0) // alpha=0
	b := BlenderGray8Pre[color.Linear]{}
	BlendGrayHlinePre(dst, 0, 4, src, nil, b)
	for i, v := range dst {
		if v != 77 {
			t.Errorf("pixel %d: want 77, got %d (zero-alpha should leave unchanged)", i, v)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Gray16 — SetPlain/GetPlain / Pre variants / FillGray16Span
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderGray16SetGetPlain(t *testing.T) {
	b := BlenderGray16[color.Linear]{}
	var pix basics.Int16u = 0
	b.SetPlain(&pix, 0xABCD, 0xFFFF)

	v, a := b.GetPlain(&pix)
	if a != 0xFFFF {
		t.Errorf("Gray16 GetPlain alpha = %d, want 65535", a)
	}
	if v != 0xABCD {
		t.Errorf("Gray16 GetPlain value = %d, want 0xABCD", v)
	}
}

func TestBlenderGray16PreSetGetPlain(t *testing.T) {
	b := BlenderGray16Pre[color.Linear]{}
	var pix basics.Int16u = 0
	b.SetPlain(&pix, 60000, 32768)
	_, a := b.GetPlain(&pix)
	if a != 0xFFFF {
		t.Errorf("Gray16Pre GetPlain alpha = %d, want 65535", a)
	}
}

func TestBlendGray16PixelPre(t *testing.T) {
	b := BlenderGray16Pre[color.Linear]{}
	var pix basics.Int16u = 0
	src := color.Gray16[color.Linear]{V: 60000, A: 65535}
	BlendGray16PixelPre(&pix, src, 65535, b)
	if pix == 0 {
		t.Error("BlendGray16PixelPre: pixel should be non-zero after opaque blend")
	}
}

func TestBlendGray16HlinePre(t *testing.T) {
	dst := make([]basics.Int16u, 6)
	src := color.Gray16[color.Linear]{V: 60000, A: 65535}
	b := BlenderGray16Pre[color.Linear]{}
	BlendGray16HlinePre(dst, 0, 3, src, nil, b)
	if dst[0] == 0 {
		t.Error("BlendGray16HlinePre: first pixel should be non-zero")
	}
}

func TestFillGray16Span(t *testing.T) {
	dst := make([]basics.Int16u, 5)
	src := color.Gray16[color.Linear]{V: 0xBEEF, A: 65535}
	FillGray16Span(dst, 0, 5, src)
	if dst[0] != 0xBEEF {
		t.Errorf("FillGray16Span: pixel = %d, want 0xBEEF", dst[0])
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Gray32 — SetPlain/GetPlain / Pre variants / FillGray32Span
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderGray32SetGetPlain(t *testing.T) {
	b := BlenderGray32[color.Linear]{}
	var pix float32 = 0
	b.SetPlain(&pix, 0.75, 1.0)
	v, a := b.GetPlain(&pix)
	if a != 1.0 {
		t.Errorf("Gray32 GetPlain alpha = %g, want 1.0", a)
	}
	if v != 0.75 {
		t.Errorf("Gray32 GetPlain value = %g, want 0.75", v)
	}
}

func TestBlenderGray32PreSetGetPlain(t *testing.T) {
	b := BlenderGray32Pre[color.Linear]{}
	var pix float32 = 0
	b.SetPlain(&pix, 0.8, 0.5)
	_, a := b.GetPlain(&pix)
	if a != 1.0 {
		t.Errorf("Gray32Pre GetPlain alpha = %g, want 1.0", a)
	}
}

func TestBlendGray32PixelPre(t *testing.T) {
	b := BlenderGray32Pre[color.Linear]{}
	var pix float32 = 0
	src := color.Gray32[color.Linear]{V: 0.8, A: 1.0}
	BlendGray32PixelPre(&pix, src, 1.0, b)
	if pix == 0 {
		t.Error("BlendGray32PixelPre: pixel should be non-zero after opaque blend")
	}
}

func TestFillGray32Span(t *testing.T) {
	dst := make([]float32, 5)
	src := color.Gray32[color.Linear]{V: 0.5, A: 1.0}
	FillGray32Span(dst, 0, 5, src)
	if dst[0] != 0.5 {
		t.Errorf("FillGray32Span: pixel = %g, want 0.5", dst[0])
	}
}

// ──────────────────────────────────────────────────────────────────────────
// RGB8 — SetPlain/GetPlain/BlendPix for BlenderRGB8 and BlenderRGBPre
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderRGB8SetGetPlain(t *testing.T) {
	b := BlenderRGB8[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 3)
	b.SetPlain(pix, 100, 150, 200)

	r, g, bb := b.GetPlain(pix)
	if r != 100 || g != 150 || bb != 200 {
		t.Errorf("RGB8 SetPlain/GetPlain: got (%d,%d,%d), want (100,150,200)", r, g, bb)
	}
}

func TestBlenderRGBPreSetGetPlain(t *testing.T) {
	b := BlenderRGBPre[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 3)
	b.SetPlain(pix, 80, 160, 240)

	r, g, bb := b.GetPlain(pix)
	if r != 80 || g != 160 || bb != 240 {
		t.Errorf("RGBPre SetPlain/GetPlain: got (%d,%d,%d), want (80,160,240)", r, g, bb)
	}
}

func TestBlenderRGB8IdxFunctions(t *testing.T) {
	b := BlenderRGB8[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGB8 idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

func TestBlenderRGBPreIdxFunctions(t *testing.T) {
	b := BlenderRGBPre[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGBPre idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

func TestBlenderRGBBGRSetGetPlain(t *testing.T) {
	b := BlenderRGB8[color.Linear, order.BGR]{}
	pix := make([]basics.Int8u, 3)
	b.SetPlain(pix, 100, 150, 200)
	r, g, bb := b.GetPlain(pix)
	if r != 100 || g != 150 || bb != 200 {
		t.Errorf("RGB8 BGR SetPlain/GetPlain: got (%d,%d,%d), want (100,150,200)", r, g, bb)
	}
}

// ──────────────────────────────────────────────────────────────────────────
// RGBA8 — SetPlain/GetPlain/PremulSrc / IdxR/G/B/A
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderRGBA8SetGetPlain(t *testing.T) {
	b := BlenderRGBA8[color.Linear, order.RGBA]{}
	pix := make([]basics.Int8u, 4)
	b.SetPlain(pix, 100, 150, 200, 255)
	r, g, bb, a := b.GetPlain(pix)
	if r != 100 || g != 150 || bb != 200 || a != 255 {
		t.Errorf("RGBA8 SetPlain/GetPlain: got (%d,%d,%d,%d), want (100,150,200,255)", r, g, bb, a)
	}
}

func TestBlenderRGBA8PremulSrc(t *testing.T) {
	b := BlenderRGBA8[color.Linear, order.RGBA]{}
	_ = b.PremulSrc()
}

func TestBlenderRGBA8PreSetGetPlain(t *testing.T) {
	b := BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	pix := make([]basics.Int8u, 4)
	b.SetPlain(pix, 100, 150, 200, 255)
	r, g, bb, a := b.GetPlain(pix)
	if r != 100 || g != 150 || bb != 200 || a != 255 {
		t.Errorf("RGBA8Pre SetPlain/GetPlain: got (%d,%d,%d,%d)", r, g, bb, a)
	}
}

func TestBlenderRGBA8PrePremulSrc(t *testing.T) {
	b := BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	_ = b.PremulSrc()
}

func TestBlenderRGBA8IdxFunctions(t *testing.T) {
	b := BlenderRGBA8[color.Linear, order.RGBA]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 || b.IdxA() != 3 {
		t.Errorf("RGBA8 idx: R=%d G=%d B=%d A=%d, want 0 1 2 3", b.IdxR(), b.IdxG(), b.IdxB(), b.IdxA())
	}
}

func TestBlenderRGBA8PreIdxFunctions(t *testing.T) {
	b := BlenderRGBA8Pre[color.Linear, order.RGBA]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 || b.IdxA() != 3 {
		t.Errorf("RGBA8Pre idx: R=%d G=%d B=%d A=%d, want 0 1 2 3", b.IdxR(), b.IdxG(), b.IdxB(), b.IdxA())
	}
}

func TestFillRGBASpan(t *testing.T) {
	dst := make([]basics.Int8u, 5*4)
	src := color.RGBA8[color.Linear]{R: 200, G: 100, B: 50, A: 255}
	FillRGBASpan[color.Linear, order.RGBA](dst, 0, 5, src)
	for i := 0; i < 5; i++ {
		if dst[i*4+0] != 200 {
			t.Errorf("pixel %d R = %d, want 200", i, dst[i*4+0])
		}
	}
}

func TestFillRGBSpan(t *testing.T) {
	dst := make([]basics.Int8u, 5*3)
	src := color.RGB8[color.Linear]{R: 200, G: 100, B: 50}
	FillRGBSpan[color.Linear, order.RGB](dst, 0, 5, src)
	for i := 0; i < 5; i++ {
		if dst[i*3+0] != 200 {
			t.Errorf("pixel %d R = %d, want 200", i, dst[i*3+0])
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Composite blender — constructors & GetOp / Porter-Duff operations
// ──────────────────────────────────────────────────────────────────────────

func blendComposite(t *testing.T, name string, bl CompositeBlender[color.Linear, order.RGBA]) {
	t.Helper()
	dst := []basics.Int8u{128, 64, 32, 200}
	bl.BlendPix(dst, 255, 0, 0, 255, 255)
	// Just verify it runs without panic; exact values are op-specific
	t.Logf("%s result: R=%d G=%d B=%d A=%d", name, dst[0], dst[1], dst[2], dst[3])
}

func TestCompositeBlenderConstructors(t *testing.T) {
	blendComposite(t, "DstOver", NewDstOverBlender[color.Linear, order.RGBA]())
	blendComposite(t, "DstIn", NewDstInBlender[color.Linear, order.RGBA]())
	blendComposite(t, "SrcOut", NewSrcOutBlender[color.Linear, order.RGBA]())
	blendComposite(t, "DstOut", NewDstOutBlender[color.Linear, order.RGBA]())
	blendComposite(t, "SrcAtop", NewSrcAtopBlender[color.Linear, order.RGBA]())
	blendComposite(t, "DstAtop", NewDstAtopBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Darken", NewDarkenBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Lighten", NewLightenBlender[color.Linear, order.RGBA]())
	blendComposite(t, "ColorDodge", NewColorDodgeBlender[color.Linear, order.RGBA]())
	blendComposite(t, "ColorBurn", NewColorBurnBlender[color.Linear, order.RGBA]())
	blendComposite(t, "HardLight", NewHardLightBlender[color.Linear, order.RGBA]())
	blendComposite(t, "SoftLight", NewSoftLightBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Difference", NewDifferenceBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Exclusion", NewExclusionBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Plus", NewPlusBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Clear", NewClearBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Src", NewSrcBlender[color.Linear, order.RGBA]())
	blendComposite(t, "Dst", NewDstBlender[color.Linear, order.RGBA]())
	blendComposite(t, "SrcOver", NewSrcOverBlender[color.Linear, order.RGBA]())
	blendComposite(t, "SrcIn", NewSrcInBlender[color.Linear, order.RGBA]())
}

func TestCompositeBlenderGetOp(t *testing.T) {
	bl := NewDarkenBlender[color.Linear, order.RGBA]()
	if bl.GetOp() != CompOpDarken {
		t.Errorf("GetOp = %d, want CompOpDarken (%d)", bl.GetOp(), CompOpDarken)
	}
}

func TestCompositeBlenderPreGetOp(t *testing.T) {
	bl := NewCompositeBlenderPre[color.Linear, order.RGBA](CompOpSrcOver)
	if bl.GetOp() != CompOpSrcOver {
		t.Errorf("CompositeBlenderPre.GetOp = %d, want CompOpSrcOver (%d)", bl.GetOp(), CompOpSrcOver)
	}
}

func TestCompositeBlenderPreGetSetPlain(t *testing.T) {
	bl := NewCompositeBlenderPre[color.Linear, order.RGBA](CompOpSrcOver)
	pix := make([]basics.Int8u, 4)
	bl.SetPlain(pix, 100, 150, 200, 255)
	r, g, b, a := bl.GetPlain(pix)
	// GetPlain demultiplies premultiplied dst; with alpha=255 values should be ~same
	if a != 255 {
		t.Errorf("CompositeBlenderPre GetPlain alpha = %d, want 255", a)
	}
	_ = r
	_ = g
	_ = b
}

func TestCompositeBlenderBlendColorHspan(t *testing.T) {
	// 1 row × 3 pixels, stride = 3*4 = 12
	stride := 12
	dst := make([]basics.Int8u, stride)
	for i := 0; i < 3; i++ {
		dst[i*4+3] = 200 // set some alpha
	}
	covers := []basics.Int8u{255, 255, 255}
	bl := NewSrcOverBlender[color.Linear, order.RGBA]()
	bl.BlendColorHspan(dst, 0, 0, 3, stride, 0, 255, 0, 255, covers)
}

func TestCompositeBlenderBlendColorVspan(t *testing.T) {
	// 3 rows × 1 pixel, stride = 4
	stride := 4
	dst := make([]basics.Int8u, 3*stride)
	for i := 0; i < 3; i++ {
		dst[i*4+3] = 255
	}
	covers := []basics.Int8u{255, 255, 255}
	bl := NewSrcOverBlender[color.Linear, order.RGBA]()
	bl.BlendColorVspan(dst, 0, 0, 3, stride, 0, 0, 255, 255, covers)
}

// ──────────────────────────────────────────────────────────────────────────
// RGB8 X-variants (4-byte padded)
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderRGBX8SetGetPlain(t *testing.T) {
	b := BlenderRGBX8[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 4) // 3 RGB + 1 padding
	b.SetPlain(pix, 100, 150, 200)
	r, g, bb := b.GetPlain(pix)
	if r != 100 || g != 150 || bb != 200 {
		t.Errorf("RGBX8 SetPlain/GetPlain: got (%d,%d,%d), want (100,150,200)", r, g, bb)
	}
}

func TestBlenderRGBX8BlendPix(t *testing.T) {
	b := BlenderRGBX8[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 4)
	b.BlendPix(pix, 255, 0, 0, 255, 255)
	if pix[0] == 0 {
		t.Error("RGBX8 BlendPix: R channel should be non-zero after blending red")
	}
}

func TestBlenderRGBX8IdxFunctions(t *testing.T) {
	b := BlenderRGBX8[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGBX8 idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

func TestBlenderRGBXPreSetGetPlain(t *testing.T) {
	b := BlenderRGBXPre[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 4)
	b.SetPlain(pix, 80, 160, 240)
	r, g, bb := b.GetPlain(pix)
	if r != 80 || g != 160 || bb != 240 {
		t.Errorf("RGBXPre SetPlain/GetPlain: got (%d,%d,%d), want (80,160,240)", r, g, bb)
	}
}

func TestBlenderRGBXPreBlendPix(t *testing.T) {
	b := BlenderRGBXPre[color.Linear, order.RGB]{}
	pix := make([]basics.Int8u, 4)
	b.BlendPix(pix, 200, 0, 0, 255, 255)
	if pix[0] == 0 {
		t.Error("RGBXPre BlendPix: R channel should be non-zero")
	}
}

func TestBlenderRGBXPreIdxFunctions(t *testing.T) {
	b := BlenderRGBXPre[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGBXPre idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

// ──────────────────────────────────────────────────────────────────────────
// RGB16 (RGB48) blenders
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderRGB48SetGetPlain(t *testing.T) {
	b := BlenderRGB48[color.Linear, order.RGB]{}
	pix := make([]basics.Int16u, 3)
	b.SetPlain(pix, 10000, 20000, 30000)
	r, g, bb := b.GetPlain(pix)
	if r != 10000 || g != 20000 || bb != 30000 {
		t.Errorf("RGB48 SetPlain/GetPlain: got (%d,%d,%d), want (10000,20000,30000)", r, g, bb)
	}
}

func TestBlenderRGB48BlendPix(t *testing.T) {
	b := BlenderRGB48[color.Linear, order.RGB]{}
	pix := make([]basics.Int16u, 3)
	b.BlendPix(pix, 60000, 0, 0, 65535, 65535)
	if pix[0] == 0 {
		t.Error("RGB48 BlendPix: R channel should be non-zero after blending red")
	}
}

func TestBlenderRGB48IdxFunctions(t *testing.T) {
	b := BlenderRGB48[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGB48 idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

func TestBlenderRGB48PreSetGetPlain(t *testing.T) {
	b := BlenderRGB48Pre[color.Linear, order.RGB]{}
	pix := make([]basics.Int16u, 3)
	b.SetPlain(pix, 10000, 20000, 30000)
	r, g, bb := b.GetPlain(pix)
	if r != 10000 || g != 20000 || bb != 30000 {
		t.Errorf("RGB48Pre SetPlain/GetPlain: got (%d,%d,%d)", r, g, bb)
	}
}

func TestBlenderRGB48PreBlendPix(t *testing.T) {
	b := BlenderRGB48Pre[color.Linear, order.RGB]{}
	pix := make([]basics.Int16u, 3)
	b.BlendPix(pix, 60000, 0, 0, 65535, 65535)
	if pix[0] == 0 {
		t.Error("RGB48Pre BlendPix: R channel should be non-zero")
	}
}

func TestBlenderRGB48PreIdxFunctions(t *testing.T) {
	b := BlenderRGB48Pre[color.Linear, order.RGB]{}
	if b.IdxR() != 0 || b.IdxG() != 1 || b.IdxB() != 2 {
		t.Errorf("RGB48Pre idx: R=%d G=%d B=%d, want 0 1 2", b.IdxR(), b.IdxG(), b.IdxB())
	}
}

// ──────────────────────────────────────────────────────────────────────────
// RGB32 (RGB96) blenders
// ──────────────────────────────────────────────────────────────────────────

func TestBlenderRGB96SetGetPlain(t *testing.T) {
	b := BlenderRGB96[color.Linear, order.RGB]{}
	pix := make([]float32, 3)
	b.SetPlain(pix, 0.5, 0.7, 0.9, 1.0)
	r, g, bb, a := b.GetPlain(pix)
	if r != 0.5 || g != 0.7 || bb != 0.9 {
		t.Errorf("RGB96 SetPlain/GetPlain: got (%g,%g,%g), want (0.5,0.7,0.9)", r, g, bb)
	}
	if a != 1.0 {
		t.Errorf("RGB96 GetPlain alpha = %g, want 1.0", a)
	}
}

func TestBlenderRGB96BlendPix(t *testing.T) {
	b := BlenderRGB96[color.Linear, order.RGB]{}
	pix := make([]float32, 3)
	b.BlendPix(pix, 1.0, 0.0, 0.0, 1.0, 1.0)
	if pix[0] == 0 {
		t.Error("RGB96 BlendPix: R channel should be non-zero after blending red")
	}
}

func TestBlenderRGB96PreSetGetPlain(t *testing.T) {
	b := BlenderRGB96Pre[color.Linear, order.RGB]{}
	pix := make([]float32, 3)
	b.SetPlain(pix, 0.5, 0.7, 0.9, 1.0)
	r, g, bb, _ := b.GetPlain(pix)
	// SetPlain premultiplies: r*a = 0.5, g*a=0.7, b*a=0.9
	if r != 0.5 || g != 0.7 || bb != 0.9 {
		t.Errorf("RGB96Pre SetPlain/GetPlain: got (%g,%g,%g)", r, g, bb)
	}
}

func TestBlenderRGB96PreBlendPix(t *testing.T) {
	b := BlenderRGB96Pre[color.Linear, order.RGB]{}
	pix := make([]float32, 3)
	b.BlendPix(pix, 0.8, 0.0, 0.0, 1.0, 1.0)
	if pix[0] == 0 {
		t.Error("RGB96Pre BlendPix: R channel should be non-zero")
	}
}

// ──────────────────────────────────────────────────────────────────────────
// Gamma blender constructors (just verify they compile & return non-zero)
// ──────────────────────────────────────────────────────────────────────────

// trivialGamma is a no-op gamma corrector for testing purposes.
type trivialGamma struct{}

func (trivialGamma) Dir(v basics.Int8u) basics.Int8u { return v }
func (trivialGamma) Inv(v basics.Int8u) basics.Int8u { return v }

func TestNewBlenderRGBGamma(t *testing.T) {
	g := NewBlenderRGBGamma[color.Linear, order.RGB](trivialGamma{})
	pix := make([]basics.Int8u, 3)
	g.SetPlain(pix, 100, 150, 200)
	r, gg, b := g.GetPlain(pix)
	if r == 0 && gg == 0 && b == 0 {
		t.Error("BlenderRGBGamma GetPlain returned all zeros")
	}
}

func TestNewBlenderRGBGammaBGR(t *testing.T) {
	g := NewBlenderRGBGamma[color.Linear, order.BGR](trivialGamma{})
	pix := make([]basics.Int8u, 3)
	g.SetPlain(pix, 100, 150, 200)
	r, gg, b := g.GetPlain(pix)
	if r == 0 && gg == 0 && b == 0 {
		t.Error("BlenderRGBGamma BGR GetPlain returned all zeros")
	}
}

func TestBlenderRGBGammaBlendPix(t *testing.T) {
	// Must exercise BlendPix with non-zero alpha to actually enter the blend path
	g := NewBlenderRGBGamma[color.Linear, order.RGB](trivialGamma{})
	pix := make([]basics.Int8u, 3)
	g.BlendPix(pix, 200, 0, 0, 255, 255)
	if pix[0] == 0 {
		t.Error("BlenderRGBGamma BlendPix: R should be non-zero")
	}
}

// trivialGamma16 is a no-op gamma corrector for 16-bit testing.
type trivialGamma16 struct{}

func (trivialGamma16) Dir(v basics.Int16u) basics.Int16u { return v }
func (trivialGamma16) Inv(v basics.Int16u) basics.Int16u { return v }

func TestNewBlenderRGB48Gamma(t *testing.T) {
	g := NewBlenderRGB48Gamma[color.Linear, order.RGB](trivialGamma16{})
	pix := make([]basics.Int16u, 3)
	g.SetPlain(pix, 10000, 20000, 30000)
	r, gg, b := g.GetPlain(pix)
	if r != 10000 || gg != 20000 || b != 30000 {
		t.Errorf("RGB48Gamma SetPlain/GetPlain: got (%d,%d,%d)", r, gg, b)
	}
	_ = g.IdxR()
	_ = g.IdxG()
	_ = g.IdxB()
}

func TestBlenderRGB48GammaBlendPix(t *testing.T) {
	g := NewBlenderRGB48Gamma[color.Linear, order.RGB](trivialGamma16{})
	pix := make([]basics.Int16u, 3)
	g.BlendPix(pix, 60000, 0, 0, 65535, 65535)
	if pix[0] == 0 {
		t.Error("RGB48Gamma BlendPix: R should be non-zero")
	}
}

// trivialGamma32 is a no-op gamma corrector for float32 testing.
type trivialGamma32 struct{}

func (trivialGamma32) Dir(v float32) float32 { return v }
func (trivialGamma32) Inv(v float32) float32 { return v }

func TestNewBlenderRGB96Gamma(t *testing.T) {
	g := NewBlenderRGB96Gamma[color.Linear, order.RGB](trivialGamma32{})
	pix := make([]float32, 3)
	g.SetPlain(pix, 0.5, 0.6, 0.7, 1.0)
	r, gg, b, _ := g.GetPlain(pix)
	if r != 0.5 || gg != 0.6 || b != 0.7 {
		t.Errorf("RGB96Gamma SetPlain/GetPlain: got (%g,%g,%g)", r, gg, b)
	}
}

func TestBlenderRGB96GammaBlendPix(t *testing.T) {
	g := NewBlenderRGB96Gamma[color.Linear, order.RGB](trivialGamma32{})
	pix := make([]float32, 3)
	g.BlendPix(pix, 0.8, 0.0, 0.0, 1.0, 1.0)
	if pix[0] == 0 {
		t.Error("RGB96Gamma BlendPix: R should be non-zero")
	}
}
