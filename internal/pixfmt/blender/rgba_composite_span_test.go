package blender

import (
	"math/rand"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// allCompOps is every operator BlendSolidSpanStraight must handle.
var allCompOps = []CompOp{
	CompOpClear, CompOpSrc, CompOpDst, CompOpSrcOver, CompOpDstOver,
	CompOpSrcIn, CompOpDstIn, CompOpSrcOut, CompOpDstOut, CompOpSrcAtop,
	CompOpDstAtop, CompOpXor, CompOpPlus, CompOpMultiply, CompOpScreen,
	CompOpOverlay, CompOpDarken, CompOpLighten, CompOpColorDodge, CompOpColorBurn,
	CompOpHardLight, CompOpSoftLight, CompOpDifference, CompOpExclusion,
}

// TestBlendSolidSpanStraightMatchesBlendPix is the differential test required by
// PLAN.md Phase 6: the span fast path must be byte-for-byte identical to calling
// BlendPix per pixel, for EVERY operator, over randomised straight destinations
// (including translucent ones — the case the premult-dst SIMD kernels got wrong),
// translucent sources, and full / partial / zero-mixed coverage. It runs for both
// a standard RGBA order and a swapped (BGRA) order, so the span method's
// channel-index handling (o.IdxR()/…) is exercised, not just position 0..3.
func TestBlendSolidSpanStraightMatchesBlendPix(t *testing.T) {
	t.Run("RGBA", func(t *testing.T) {
		diffSpanVsBlendPix[order.RGBA](t, rand.New(rand.NewSource(0xC0FFEE)))
	})
	t.Run("BGRA", func(t *testing.T) {
		diffSpanVsBlendPix[order.BGRA](t, rand.New(rand.NewSource(0xB264A)))
	})
}

// diffSpanVsBlendPix asserts BlendSolidSpanStraight == per-pixel BlendPix for the
// blender's byte order O across all operators and coverage modes.
func diffSpanVsBlendPix[O order.RGBAOrder](t *testing.T, rng *rand.Rand) {
	t.Helper()
	const span = 48
	coverModes := []string{"full", "partial", "zero-mixed"}

	for _, op := range allCompOps {
		bl := NewCompositeBlenderPlain[color.Linear, O](op)
		for _, mode := range coverModes {
			for trial := 0; trial < 200; trial++ {
				dst := make([]basics.Int8u, span*4)
				for i := range dst {
					dst[i] = basics.Int8u(rng.Intn(256))
				}
				r := basics.Int8u(rng.Intn(256))
				g := basics.Int8u(rng.Intn(256))
				b := basics.Int8u(rng.Intn(256))
				a := basics.Int8u(rng.Intn(256))

				var covers []basics.Int8u
				switch mode {
				case "partial":
					covers = make([]basics.Int8u, span)
					for i := range covers {
						covers[i] = basics.Int8u(rng.Intn(256))
					}
				case "zero-mixed":
					covers = make([]basics.Int8u, span)
					for i := range covers {
						if rng.Intn(4) == 0 {
							covers[i] = 0
						} else {
							covers[i] = basics.Int8u(rng.Intn(256))
						}
					}
				}

				// Reference: BlendPix per pixel.
				want := append([]basics.Int8u(nil), dst...)
				for i := 0; i < span; i++ {
					cover := basics.Int8u(basics.CoverFull)
					if covers != nil {
						cover = covers[i]
					}
					bl.BlendPix(want[i*4:i*4+4], r, g, b, a, cover)
				}

				// Fast path.
				got := append([]basics.Int8u(nil), dst...)
				bl.BlendSolidSpanStraight(got, r, g, b, a, covers, span)

				for i := 0; i < span*4; i++ {
					if got[i] != want[i] {
						t.Fatalf("op=%d %s trial=%d px=%d byte=%d: span %d != per-pixel %d (src=%d,%d,%d,%d)",
							op, mode, trial, i/4, i%4, got[i], want[i], r, g, b, a)
					}
				}
			}
		}
	}
}

// TestBlendSolidSpanStraightFaithfulStraightOverOpaque value-pins the AGG-faithful
// result of a translucent source composited over an OPAQUE destination, for the
// operators whose result is translucent over opaque dst. This is the port-side
// regression anchor for the §5.5 premultiplied-storage bug (mirrors the CPP-side
// TestCPPXorBlendIsAGGFaithfulWithAggReal): the bug would leave PREMULTIPLIED data
// in the straight buffer — e.g. xor would read back ~(15,48,22,95) instead of the
// straight (40,130,60,95). Unlike the differential test (which only proves
// span == BlendPix), this pins the actual demultiplied values, so it also guards
// the scalar bridge itself.
func TestBlendSolidSpanStraightFaithfulStraightOverOpaque(t *testing.T) {
	// src translucent blue (40,60,220) a=160; dst opaque green (40,130,60) a=255.
	// xor over opaque: Dca'=Dca(1-Sa), Da'=1-Sa → straight color unchanged, a=95.
	// dst-out over opaque: same straight color, Da'=Da(1-Sa)=95.
	cases := []struct {
		name string
		op   CompOp
		want [4]basics.Int8u
	}{
		{"xor", CompOpXor, [4]basics.Int8u{40, 130, 60, 95}},
		{"dst_out", CompOpDstOut, [4]basics.Int8u{40, 130, 60, 95}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](c.op)
			dst := []basics.Int8u{40, 130, 60, 255}
			bl.BlendSolidSpanStraight(dst, 40, 60, 220, 160, nil, 1)
			got := [4]basics.Int8u{dst[0], dst[1], dst[2], dst[3]}
			if got != c.want {
				t.Errorf("%s over opaque green: got %v want %v (premultiplied-storage bug would give ~{15,48,22,95})",
					c.name, got, c.want)
			}
		})
	}
}

func benchSpanStraight(b *testing.B, op CompOp) {
	const span = 256
	bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](op)
	template := make([]basics.Int8u, span*4)
	for i := 0; i < span; i++ {
		p := i * 4
		template[p+0], template[p+1], template[p+2], template[p+3] = 40, 130, 60, 255
	}
	work := make([]basics.Int8u, span*4)
	b.Run("scalar", func(b *testing.B) {
		b.SetBytes(span * 4)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			for i := 0; i < span; i++ {
				bl.BlendPix(work[i*4:i*4+4], 40, 60, 220, 160, basics.CoverFull)
			}
		}
	})
	b.Run("span", func(b *testing.B) {
		b.SetBytes(span * 4)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			bl.BlendSolidSpanStraight(work, 40, 60, 220, 160, nil, span)
		}
	})
}

func BenchmarkBlendSolidSpanStraightXor(b *testing.B)      { benchSpanStraight(b, CompOpXor) }
func BenchmarkBlendSolidSpanStraightSrcOver(b *testing.B)  { benchSpanStraight(b, CompOpSrcOver) }
func BenchmarkBlendSolidSpanStraightMultiply(b *testing.B) { benchSpanStraight(b, CompOpMultiply) }
