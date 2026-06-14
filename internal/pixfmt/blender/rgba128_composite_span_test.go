package blender

import (
	"math/rand"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// TestRGBA128BlendSolidSpanStraightMatchesBlendPix is the float twin of
// TestBlendSolidSpanStraightMatchesBlendPix: the hoisted span fast path on the
// straight float comp blender must be EXACTLY equal (not merely approximately —
// both paths run the identical float64 premultiply->op->demultiply bridge and
// store through clampF01) to calling BlendPix per pixel, for every operator over
// randomised straight + translucent float destinations, translucent sources, and
// full / partial / zero-mixed coverage, for both RGBA and BGRA byte order.
func TestRGBA128BlendSolidSpanStraightMatchesBlendPix(t *testing.T) {
	t.Run("RGBA", func(t *testing.T) {
		diffF32SpanVsBlendPix[order.RGBA](t, rand.New(rand.NewSource(0xF10A7)))
	})
	t.Run("BGRA", func(t *testing.T) {
		diffF32SpanVsBlendPix[order.BGRA](t, rand.New(rand.NewSource(0xB6F32)))
	})
}

// benchF32SpanStraight compares the float span fast path against per-pixel
// BlendPix for one operator (both bit-exact); reports the dispatch-elimination
// win and confirms 0 allocs.
func benchF32SpanStraight(b *testing.B, op CompOp) {
	const span = 256
	bl := NewCompositeBlenderRGBA128Plain[color.Linear, order.RGBA](op)
	template := make([]float32, span*4)
	for i := 0; i < span; i++ {
		p := i * 4
		template[p+0], template[p+1], template[p+2], template[p+3] = 0.16, 0.51, 0.24, 1.0
	}
	work := make([]float32, span*4)
	b.Run("scalar", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			for i := 0; i < span; i++ {
				bl.BlendPix(work[i*4:i*4+4], 0.16, 0.24, 0.86, 0.63, 1.0)
			}
		}
	})
	b.Run("span", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			bl.BlendSolidSpanStraight(work, 0.16, 0.24, 0.86, 0.63, nil, span)
		}
	})
}

func BenchmarkRGBA128SpanStraightSrcOver(b *testing.B)  { benchF32SpanStraight(b, CompOpSrcOver) }
func BenchmarkRGBA128SpanStraightMultiply(b *testing.B) { benchF32SpanStraight(b, CompOpMultiply) }

func diffF32SpanVsBlendPix[O order.RGBAOrder](t *testing.T, rng *rand.Rand) {
	t.Helper()
	const span = 48
	coverModes := []string{"full", "partial", "zero-mixed"}

	for _, op := range allCompOps {
		bl := NewCompositeBlenderRGBA128Plain[color.Linear, O](op)
		for _, mode := range coverModes {
			for trial := 0; trial < 100; trial++ {
				dst := make([]float32, span*4)
				for i := range dst {
					dst[i] = rng.Float32()
				}
				r := rng.Float32()
				g := rng.Float32()
				b := rng.Float32()
				a := rng.Float32()

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

				// Reference: BlendPix per pixel, with the cover normalised exactly
				// as the pixfmt's coverToF32 (float32(cover)/255) before BlendPix
				// widens it.
				want := append([]float32(nil), dst...)
				for i := 0; i < span; i++ {
					cover := float32(1.0)
					if covers != nil {
						cover = float32(covers[i]) / 255.0
					}
					bl.BlendPix(want[i*4:i*4+4], r, g, b, a, cover)
				}

				// Fast path.
				got := append([]float32(nil), dst...)
				bl.BlendSolidSpanStraight(got, r, g, b, a, covers, span)

				for i := 0; i < span*4; i++ {
					if got[i] != want[i] {
						t.Fatalf("op=%d %s trial=%d px=%d ch=%d: span %v != per-pixel %v (src=%v,%v,%v,%v)",
							op, mode, trial, i/4, i%4, got[i], want[i], r, g, b, a)
					}
				}
			}
		}
	}
}
