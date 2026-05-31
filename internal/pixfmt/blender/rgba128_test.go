package blender

import (
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

const rgba128Eps = 1e-6

func approxEq(a, b float32) bool {
	return math.Abs(float64(a)-float64(b)) <= rgba128Eps
}

func assertPx(t *testing.T, got []float32, wantR, wantG, wantB, wantA float32) {
	t.Helper()
	// Tests always use RGBA order for assertions unless stated otherwise.
	if !approxEq(got[0], wantR) || !approxEq(got[1], wantG) || !approxEq(got[2], wantB) || !approxEq(got[3], wantA) {
		t.Fatalf("pixel = {%v, %v, %v, %v}, want {%v, %v, %v, %v}",
			got[0], got[1], got[2], got[3], wantR, wantG, wantB, wantA)
	}
}

// --- arithmetic helpers -----------------------------------------------------

func TestRGBA128LerpAndPrelerp(t *testing.T) {
	if got := RGBA128Lerp(0.2, 1.0, 0.5); !approxEq(got, 0.6) {
		t.Errorf("RGBA128Lerp(0.2,1.0,0.5) = %v, want 0.6", got)
	}
	// prelerp(p,q,a) = p + q - p*a
	if got := RGBA128Prelerp(0.5, 0.5, 0.5); !approxEq(got, 0.75) {
		t.Errorf("RGBA128Prelerp(0.5,0.5,0.5) = %v, want 0.75", got)
	}
}

// --- SetPlain / GetPlain round-trip & channel order -------------------------

func TestRGBA128SetGetPlainRGBA(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.RGBA]
	px := make([]float32, 4)
	bl.SetPlain(px, 0.1, 0.2, 0.3, 0.4)
	assertPx(t, px, 0.1, 0.2, 0.3, 0.4)

	r, g, b, a := bl.GetPlain(px)
	if !approxEq(r, 0.1) || !approxEq(g, 0.2) || !approxEq(b, 0.3) || !approxEq(a, 0.4) {
		t.Fatalf("GetPlain = (%v,%v,%v,%v), want (0.1,0.2,0.3,0.4)", r, g, b, a)
	}
}

func TestRGBA128SetPlainHonorsBGRAOrder(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.BGRA]
	px := make([]float32, 4)
	bl.SetPlain(px, 1.0, 0.5, 0.25, 1.0) // r,g,b,a
	// BGRA: B=0, G=1, R=2, A=3
	if !approxEq(px[0], 0.25) || !approxEq(px[1], 0.5) || !approxEq(px[2], 1.0) || !approxEq(px[3], 1.0) {
		t.Fatalf("BGRA pixel = {%v,%v,%v,%v}, want {0.25,0.5,1,1}", px[0], px[1], px[2], px[3])
	}
}

// --- BlenderRGBA128: plain source -> premultiplied destination ---------------

func TestRGBA128BlendPlainOverTransparentFullCover(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.RGBA]
	px := []float32{0, 0, 0, 0} // premultiplied transparent
	bl.BlendPix(px, 1.0, 0.5, 0.25, 1.0, 1.0)
	assertPx(t, px, 1.0, 0.5, 0.25, 1.0)
}

func TestRGBA128BlendPlainHalfCover(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.RGBA]
	px := []float32{0, 0, 0, 0}
	bl.BlendPix(px, 1.0, 0.5, 0.25, 1.0, 0.5)
	// a' = 1*0.5 = 0.5; channels lerp from 0; alpha prelerp from 0 -> 0.5
	assertPx(t, px, 0.5, 0.25, 0.125, 0.5)
}

func TestRGBA128BlendZeroAlphaIsNoOp(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.RGBA]
	px := []float32{0.3, 0.3, 0.3, 0.7}
	bl.BlendPix(px, 1.0, 1.0, 1.0, 0.0, 1.0)
	assertPx(t, px, 0.3, 0.3, 0.3, 0.7)
}

func TestRGBA128BlendZeroCoverIsNoOp(t *testing.T) {
	var bl BlenderRGBA128[color.Linear, order.RGBA]
	px := []float32{0.3, 0.3, 0.3, 0.7}
	bl.BlendPix(px, 1.0, 1.0, 1.0, 1.0, 0.0)
	assertPx(t, px, 0.3, 0.3, 0.3, 0.7)
}

// --- BlenderRGBA128Pre: premultiplied source -> premultiplied destination ----

func TestRGBA128PreBlendOverTransparent(t *testing.T) {
	var bl BlenderRGBA128Pre[color.Linear, order.RGBA]
	px := []float32{0, 0, 0, 0}
	// premultiplied source (r,g,b already * a)
	bl.BlendPix(px, 0.5, 0.25, 0.125, 0.5, 1.0)
	assertPx(t, px, 0.5, 0.25, 0.125, 0.5)
}

func TestRGBA128PreReportsPremulSrc(t *testing.T) {
	var pre BlenderRGBA128Pre[color.Linear, order.RGBA]
	if !pre.PremulSrc() {
		t.Error("BlenderRGBA128Pre.PremulSrc() = false, want true")
	}
	var plainToPre BlenderRGBA128[color.Linear, order.RGBA]
	if plainToPre.PremulSrc() {
		t.Error("BlenderRGBA128.PremulSrc() = true, want false")
	}
}

// --- BlenderRGBA128Plain: plain source -> plain destination ------------------

func TestRGBA128PlainOpaqueReplacesDst(t *testing.T) {
	var bl BlenderRGBA128Plain[color.Linear, order.RGBA]
	px := []float32{0.4, 0.4, 0.4, 0.5} // straight (non-premultiplied)
	// opaque black, full cover -> fully covers, alpha goes opaque
	bl.BlendPix(px, 0.0, 0.0, 0.0, 1.0, 1.0)
	assertPx(t, px, 0.0, 0.0, 0.0, 1.0)
}

// --- single-pixel + hline helpers over color.RGBA32 -------------------------

func TestBlendRGBA128PixelOverTransparent(t *testing.T) {
	px := []float32{0, 0, 0, 0}
	src := color.NewRGBA32[color.Linear](1.0, 0.5, 0.25, 1.0)
	BlendRGBA128Pixel[color.Linear, order.RGBA](px, src, 1.0, BlenderRGBA128[color.Linear, order.RGBA]{})
	assertPx(t, px, 1.0, 0.5, 0.25, 1.0)
}

func TestCopyRGBA128Hline(t *testing.T) {
	px := make([]float32, 4*3)
	src := color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8)
	CopyRGBA128Hline[color.Linear, order.RGBA](px, 0, 3, src)
	for i := range 3 {
		base := i * 4
		assertPx(t, px[base:base+4], 0.2, 0.4, 0.6, 0.8)
	}
}

func TestBlendRGBA128HlineFullCover(t *testing.T) {
	px := make([]float32, 4*2) // two transparent premul pixels
	src := color.NewRGBA32[color.Linear](1.0, 0.5, 0.25, 1.0)
	BlendRGBA128Hline[color.Linear, order.RGBA](px, 0, 2, src, nil, BlenderRGBA128[color.Linear, order.RGBA]{})
	assertPx(t, px[0:4], 1.0, 0.5, 0.25, 1.0)
	assertPx(t, px[4:8], 1.0, 0.5, 0.25, 1.0)
}
