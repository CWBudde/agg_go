package blender

import (
	"math/rand"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/simd"
)

// allCompOps is every operator BlendSolidSpanStraight must handle.
var allCompOps = []CompOp{
	CompOpClear, CompOpSrc, CompOpDst, CompOpSrcOver, CompOpDstOver,
	CompOpSrcIn, CompOpDstIn, CompOpSrcOut, CompOpDstOut, CompOpSrcAtop,
	CompOpDstAtop, CompOpXor, CompOpPlus, CompOpMultiply, CompOpScreen,
	CompOpOverlay, CompOpDarken, CompOpLighten, CompOpColorDodge, CompOpColorBurn,
	CompOpHardLight, CompOpSoftLight, CompOpDifference, CompOpExclusion,
	CompOpDissolve, CompOpLinearBurn, CompOpDarkerColor, CompOpLighterColor,
	CompOpVividLight, CompOpLinearLight, CompOpPinLight, CompOpHardMix,
	CompOpSubtract, CompOpDivide, CompOpHue, CompOpSaturation, CompOpColor,
	CompOpLuminosity, CompOpColorBurnPhotoshop,
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

// TestBlendColorSpanStraightMatchesBlendPix is the per-pixel-colour twin of
// TestBlendSolidSpanStraightMatchesBlendPix: the gradient/image colour-span fast
// path must be byte-for-byte identical to calling BlendPix per pixel, for EVERY
// operator, over randomised straight + translucent destinations, a per-pixel
// translucent source colour array, and full / partial / zero-mixed coverage, for
// both RGBA and BGRA byte order.
func TestBlendColorSpanStraightMatchesBlendPix(t *testing.T) {
	t.Run("RGBA", func(t *testing.T) {
		diffColorSpanVsBlendPix[order.RGBA](t, rand.New(rand.NewSource(0xC010A)))
	})
	t.Run("BGRA", func(t *testing.T) {
		diffColorSpanVsBlendPix[order.BGRA](t, rand.New(rand.NewSource(0xB6C01)))
	})
}

func diffColorSpanVsBlendPix[O order.RGBAOrder](t *testing.T, rng *rand.Rand) {
	t.Helper()
	const span = 48
	coverModes := []string{"full", "partial", "zero-mixed"}
	const uniformCover = basics.Int8u(200) // the cover used when covers == nil

	for _, op := range allCompOps {
		bl := NewCompositeBlenderPlain[color.Linear, O](op)
		for _, mode := range coverModes {
			for trial := 0; trial < 200; trial++ {
				dst := make([]basics.Int8u, span*4)
				for i := range dst {
					dst[i] = basics.Int8u(rng.Intn(256))
				}
				colors := make([]color.RGBA8[color.Linear], span)
				for i := range colors {
					colors[i] = color.RGBA8[color.Linear]{
						R: basics.Int8u(rng.Intn(256)),
						G: basics.Int8u(rng.Intn(256)),
						B: basics.Int8u(rng.Intn(256)),
						A: basics.Int8u(rng.Intn(256)),
					}
				}

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

				// Reference: BlendPix per pixel, with the same actualCover selection
				// the pixfmt's BlendColorHspan performs (covers[i] when present, else
				// the uniform cover).
				want := append([]basics.Int8u(nil), dst...)
				for i := 0; i < span; i++ {
					cover := uniformCover
					if i < len(covers) {
						cover = covers[i]
					}
					c := colors[i]
					bl.BlendPix(want[i*4:i*4+4], c.R, c.G, c.B, c.A, cover)
				}

				// Fast path.
				got := append([]basics.Int8u(nil), dst...)
				bl.BlendColorSpanStraight(got, colors, covers, uniformCover, span)

				for i := 0; i < span*4; i++ {
					if got[i] != want[i] {
						t.Fatalf("op=%d %s trial=%d px=%d byte=%d: span %d != per-pixel %d (color=%v)",
							op, mode, trial, i/4, i%4, got[i], want[i], colors[i/4])
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

// TestBlendSolidSpanStraightSrcOverSIMDMatchesScalar locks the optional AVX2
// SIMD tier (PLAN.md Phase 6) bit-for-bit against the scalar bridge through the
// real wired entry point. It runs the SAME BlendSolidSpanStraight call twice —
// once with SIMD forced off (the scalar premultiply->op->demultiply loop) and
// once with AVX2 forced on (the float64 kernel) — and asserts byte-identical
// output. The SIMD path covers only uniform coverage (covers == nil) SrcOver on
// an alpha-at-byte-3 buffer (RGBA/BGRA), so those are exactly the cases tested,
// across pixel counts that exercise the per-pixel kernel including short spans.
func TestBlendSolidSpanStraightSrcOverSIMDMatchesScalar(t *testing.T) {
	if !simd.DetectFeatures().HasAVX2 {
		t.Skip("AVX2 unavailable; SIMD tier not exercised on this CPU")
	}
	t.Cleanup(simd.ResetDetection)

	t.Run("RGBA", func(t *testing.T) {
		srcOverSIMDvsScalar[order.RGBA](t, rand.New(rand.NewSource(0x5114D)), simd.Features{HasAVX2: true})
	})
	t.Run("BGRA", func(t *testing.T) {
		srcOverSIMDvsScalar[order.BGRA](t, rand.New(rand.NewSource(0xB6A5A)), simd.Features{HasAVX2: true})
	})
}

// TestBlendSolidSpanStraightSrcOverSSE2MatchesScalar locks the SSE2 fallback tier
// (the pre-AVX2 amd64 path) against the scalar bridge, exactly as the AVX2 test
// does for its tier. Forcing {HasSSE2:true} with AVX2 absent selects the SSE2
// kernel in CompSrcOverPlainStraightHspanRGBA. SSE2 is architecturally guaranteed
// on amd64, so DetectFeatures().HasSSE2 gates this to the platforms where the
// kernel actually exists.
func TestBlendSolidSpanStraightSrcOverSSE2MatchesScalar(t *testing.T) {
	if !simd.DetectFeatures().HasSSE2 {
		t.Skip("SSE2 unavailable; SSE2 tier not exercised on this platform")
	}
	t.Cleanup(simd.ResetDetection)

	t.Run("RGBA", func(t *testing.T) {
		srcOverSIMDvsScalar[order.RGBA](t, rand.New(rand.NewSource(0x55E20)), simd.Features{HasSSE2: true})
	})
	t.Run("BGRA", func(t *testing.T) {
		srcOverSIMDvsScalar[order.BGRA](t, rand.New(rand.NewSource(0xB55E2)), simd.Features{HasSSE2: true})
	})
}

func srcOverSIMDvsScalar[O order.RGBAOrder](t *testing.T, rng *rand.Rand, simdFeatures simd.Features) {
	t.Helper()
	bl := NewCompositeBlenderPlain[color.Linear, O](CompOpSrcOver)
	counts := []int{1, 2, 3, 4, 5, 7, 8, 15, 16, 17, 31, 33, 64, 255, 256}

	for _, count := range counts {
		for trial := 0; trial < 64; trial++ {
			base := make([]basics.Int8u, count*4)
			for i := range base {
				base[i] = basics.Int8u(rng.Intn(256)) // straight, incl. translucent dst
			}
			r := basics.Int8u(rng.Intn(256))
			g := basics.Int8u(rng.Intn(256))
			b := basics.Int8u(rng.Intn(256))
			a := basics.Int8u(1 + rng.Intn(255)) // a>0 so the SIMD path engages

			simd.SetForcedFeatures(simd.Features{ForceGeneric: true})
			scalar := append([]basics.Int8u(nil), base...)
			bl.BlendSolidSpanStraight(scalar, r, g, b, a, nil, count)

			simd.SetForcedFeatures(simdFeatures)
			got := append([]basics.Int8u(nil), base...)
			bl.BlendSolidSpanStraight(got, r, g, b, a, nil, count)

			for i := range got {
				if got[i] != scalar[i] {
					t.Fatalf("count=%d trial=%d byte=%d (px=%d ch=%d): simd %d != scalar %d (src=%d,%d,%d,%d)",
						count, trial, i, i/4, i%4, got[i], scalar[i], r, g, b, a)
				}
			}
		}
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

// BenchmarkSrcOverSpanSIMDvsScalar measures the optional AVX2 SIMD tier against
// the scalar bridge for the same BlendSolidSpanStraight SrcOver call: the
// "scalar" subtest forces SIMD off, "simd" forces AVX2 on. Both are bit-exact
// (locked by TestBlendSolidSpanStraightSrcOverSIMDMatchesScalar); this just
// reports the speedup.
func BenchmarkSrcOverSpanSIMDvsScalar(b *testing.B) {
	feats := simd.DetectFeatures() // capture real CPU before any forcing below
	if !feats.HasAVX2 {
		b.Skip("AVX2 unavailable")
	}
	b.Cleanup(simd.ResetDetection)
	const span = 256
	bl := NewCompositeBlenderPlain[color.Linear, order.RGBA](CompOpSrcOver)
	template := make([]basics.Int8u, span*4)
	for i := 0; i < span; i++ {
		p := i * 4
		template[p+0], template[p+1], template[p+2], template[p+3] = 40, 130, 60, 255
	}
	work := make([]basics.Int8u, span*4)
	run := func(b *testing.B) {
		b.SetBytes(span * 4)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			bl.BlendSolidSpanStraight(work, 40, 60, 220, 160, nil, span)
		}
	}
	b.Run("scalar", func(b *testing.B) {
		simd.SetForcedFeatures(simd.Features{ForceGeneric: true})
		run(b)
	})
	if feats.HasSSE2 {
		b.Run("sse2", func(b *testing.B) {
			simd.SetForcedFeatures(simd.Features{HasSSE2: true})
			run(b)
		})
	}
	b.Run("simd", func(b *testing.B) {
		simd.SetForcedFeatures(simd.Features{HasAVX2: true})
		run(b)
	})
}

func BenchmarkBlendSolidSpanStraightXor(b *testing.B)      { benchSpanStraight(b, CompOpXor) }
func BenchmarkBlendSolidSpanStraightSrcOver(b *testing.B)  { benchSpanStraight(b, CompOpSrcOver) }
func BenchmarkBlendSolidSpanStraightMultiply(b *testing.B) { benchSpanStraight(b, CompOpMultiply) }
