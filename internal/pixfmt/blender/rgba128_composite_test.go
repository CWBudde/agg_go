package blender

import (
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

const compF32Eps = 1e-5

func approxF32(a, b float32) bool {
	return math.Abs(float64(a-b)) <= compF32Eps
}

// TestCompositeBlenderRGBA128SrcOverOpaque: opaque source over opaque dest is
// a straight replace (src wins).
func TestCompositeBlenderRGBA128SrcOverOpaque(t *testing.T) {
	bl := NewCompositeBlenderRGBA128[color.Linear, order.RGBA](CompOpSrcOver)
	dst := []float32{0.1, 0.2, 0.3, 1.0}
	bl.BlendPix(dst, 0.7, 0.5, 0.2, 1.0, 1.0)
	if !approxF32(dst[0], 0.7) || !approxF32(dst[1], 0.5) || !approxF32(dst[2], 0.2) || !approxF32(dst[3], 1.0) {
		t.Fatalf("src-over opaque = %v, want {0.7,0.5,0.2,1}", dst)
	}
}

// TestCompositeBlenderRGBA128Multiply: multiply of two opaque colors yields the
// component-wise product (premultiplied algebra with Sa=Da=1).
func TestCompositeBlenderRGBA128Multiply(t *testing.T) {
	bl := NewCompositeBlenderRGBA128[color.Linear, order.RGBA](CompOpMultiply)
	dst := []float32{0.4, 0.5, 0.6, 1.0}
	bl.BlendPix(dst, 0.5, 0.5, 0.5, 1.0, 1.0)
	if !approxF32(dst[0], 0.2) || !approxF32(dst[1], 0.25) || !approxF32(dst[2], 0.3) || !approxF32(dst[3], 1.0) {
		t.Fatalf("multiply = %v, want {0.2,0.25,0.3,1}", dst)
	}
}

// TestCompositeBlenderRGBA128Clear writes zero regardless of source.
func TestCompositeBlenderRGBA128Clear(t *testing.T) {
	bl := NewCompositeBlenderRGBA128[color.Linear, order.RGBA](CompOpClear)
	dst := []float32{0.4, 0.5, 0.6, 1.0}
	bl.BlendPix(dst, 0.5, 0.5, 0.5, 1.0, 1.0)
	for i, v := range dst {
		if !approxF32(v, 0) {
			t.Fatalf("clear[%d] = %v, want 0", i, v)
		}
	}
}

// TestCompositeBlenderRGBA128MatchesByteBlender checks the float blender agrees
// with the 8-bit composite blender within byte-quantization tolerance for a
// range of ops and opaque inputs (premultiplied == straight when alpha == 1).
func TestCompositeBlenderRGBA128MatchesByteBlender(t *testing.T) {
	ops := []CompOp{
		CompOpSrcOver, CompOpDstOver, CompOpMultiply, CompOpScreen,
		CompOpDarken, CompOpLighten, CompOpDifference, CompOpExclusion,
		CompOpPlus, CompOpXor,
	}
	// Opaque source and destination colors.
	sr, sg, sb := 0.7, 0.4, 0.2
	dr, dg, db := 0.3, 0.6, 0.5

	for _, op := range ops {
		// 8-bit reference.
		bbl := NewCompositeBlender[color.Linear, order.RGBA](op)
		bdst := []uint8{
			uint8(dr*255 + 0.5), uint8(dg*255 + 0.5), uint8(db*255 + 0.5), 255,
		}
		bbl.BlendPix(bdst, uint8(sr*255+0.5), uint8(sg*255+0.5), uint8(sb*255+0.5), 255, 255)

		// Float.
		fbl := NewCompositeBlenderRGBA128[color.Linear, order.RGBA](op)
		fdst := []float32{float32(dr), float32(dg), float32(db), 1.0}
		fbl.BlendPix(fdst, float32(sr), float32(sg), float32(sb), 1.0, 1.0)

		for c := range 4 {
			want := float32(bdst[c]) / 255.0
			if math.Abs(float64(fdst[c]-want)) > 0.02 {
				t.Errorf("op=%d chan=%d float=%.4f byte=%.4f (diff>0.02)", op, c, fdst[c], want)
			}
		}
	}
}

// TestCompositeBlenderRGBA128PreSrcOver verifies the premultiplied-source variant
// honors coverage scaling and composites in premultiplied space.
func TestCompositeBlenderRGBA128PreSrcOver(t *testing.T) {
	bl := NewCompositeBlenderRGBA128Pre[color.Linear, order.RGBA](CompOpSrcOver)
	// Premultiplied half-alpha gray source over opaque black.
	dst := []float32{0, 0, 0, 1.0}
	bl.BlendPix(dst, 0.25, 0.25, 0.25, 0.5, 1.0)
	// Dca' = Sca + Dca(1-Sa) = 0.25 + 0*0.5 = 0.25; Da' = 0.5 + 1*0.5 = 1.0
	if !approxF32(dst[0], 0.25) || !approxF32(dst[3], 1.0) {
		t.Fatalf("pre src-over = %v, want r=0.25 a=1", dst)
	}
}

// TestCompositeBlenderRGBA128GetSetPlain round-trips plain colors through
// premultiplied storage.
func TestCompositeBlenderRGBA128GetSetPlain(t *testing.T) {
	bl := NewCompositeBlenderRGBA128[color.Linear, order.RGBA](CompOpSrcOver)
	px := make([]float32, 4)
	bl.SetPlain(px, 0.8, 0.4, 0.2, 0.5)
	// Stored premultiplied: r = 0.8*0.5 = 0.4
	if !approxF32(px[0], 0.4) || !approxF32(px[3], 0.5) {
		t.Fatalf("SetPlain stored %v, want premultiplied r=0.4 a=0.5", px)
	}
	r, g, b, a := bl.GetPlain(px)
	if !approxF32(r, 0.8) || !approxF32(g, 0.4) || !approxF32(b, 0.2) || !approxF32(a, 0.5) {
		t.Fatalf("GetPlain = {%v,%v,%v,%v}, want straight {0.8,0.4,0.2,0.5}", r, g, b, a)
	}
}
