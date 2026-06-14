package simd_test

import (
	"math/rand"
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/simd"
)

// kernel is one of the straight-alpha "plain" composite span kernels under test,
// paired with the scalar reference op it must reproduce bit-for-bit.
type plainKernel struct {
	name string
	op   blender.CompOp
	fn   func(dst, covers []byte, r, g, b, a uint8, count int)
}

var plainKernels = []plainKernel{
	{"src_over", blender.CompOpSrcOver, simd.CompSrcOverPlainHspanRGBA},
	{"xor", blender.CompOpXor, simd.CompXorPlainHspanRGBA},
}

// scalarReference runs CompositeBlenderPlain[Linear, RGBA] pixel-by-pixel over a
// copy of dst, the path the kernel must match exactly.
func scalarReference(op blender.CompOp, dst, covers []byte, r, g, b, a uint8, count int) []byte {
	out := append([]byte(nil), dst...)
	bl := blender.NewCompositeBlenderPlain[color.Linear, order.RGBA](op)
	for i := 0; i < count; i++ {
		cover := basics.Int8u(255)
		if covers != nil {
			cover = covers[i]
		}
		bl.BlendPix(out[i*4:i*4+4], r, g, b, a, cover)
	}
	return out
}

// TestCompPlainKernelsMatchScalarBlender is the differential test required by
// PLAN.md ("Composite SIMD Fast Path"): the fast-path kernels must be bit-for-bit
// identical to the scalar CompositeBlenderPlain over randomised straight
// destinations (including translucent ones), translucent sources, and every
// coverage mode (full, none, and AA partial covers).
func TestCompPlainKernelsMatchScalarBlender(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5eed))
	const span = 64

	coverModes := []string{"full", "partial", "zero-mixed"}

	for _, k := range plainKernels {
		for _, mode := range coverModes {
			for trial := 0; trial < 400; trial++ {
				// Randomised straight-alpha destination, including fully and
				// partially transparent pixels (the case the premult-dst SIMD
				// kernels got wrong).
				dst := make([]byte, span*4)
				for i := range dst {
					dst[i] = byte(rng.Intn(256))
				}
				r := byte(rng.Intn(256))
				g := byte(rng.Intn(256))
				b := byte(rng.Intn(256))
				a := byte(rng.Intn(256))

				var covers []byte
				switch mode {
				case "full":
					covers = nil
				case "partial":
					covers = make([]byte, span)
					for i := range covers {
						covers[i] = byte(rng.Intn(256))
					}
				case "zero-mixed":
					covers = make([]byte, span)
					for i := range covers {
						if rng.Intn(4) == 0 {
							covers[i] = 0
						} else {
							covers[i] = byte(rng.Intn(256))
						}
					}
				}

				got := append([]byte(nil), dst...)
				k.fn(got, covers, r, g, b, a, span)
				want := scalarReference(k.op, dst, covers, r, g, b, a, span)

				for i := 0; i < span*4; i++ {
					if got[i] != want[i] {
						t.Fatalf("%s/%s trial=%d px=%d byte=%d: got %d want %d (src=%d,%d,%d,%d dst=%d,%d,%d,%d)",
							k.name, mode, trial, i/4, i%4, got[i], want[i],
							r, g, b, a, dst[i/4*4], dst[i/4*4+1], dst[i/4*4+2], dst[i/4*4+3])
					}
				}
			}
		}
	}
}

// scalarBlenderHspan is the per-pixel scalar path (what pixfmt's BlendSolidHspan
// does today), used as the benchmark baseline for the fast-path kernels.
func scalarBlenderHspan(op blender.CompOp, dst, covers []byte, r, g, b, a uint8, count int) {
	bl := blender.NewCompositeBlenderPlain[color.Linear, order.RGBA](op)
	for i := 0; i < count; i++ {
		cover := basics.Int8u(255)
		if covers != nil {
			cover = covers[i]
		}
		bl.BlendPix(dst[i*4:i*4+4], r, g, b, a, cover)
	}
}

func benchPlainCompare(b *testing.B, op blender.CompOp, kernel func(dst, covers []byte, r, g, b, a uint8, count int)) {
	const span = 256
	template := make([]byte, span*4)
	for i := 0; i < span; i++ {
		p := i * 4
		template[p+0], template[p+1], template[p+2], template[p+3] = 40, 130, 60, 255
	}
	work := make([]byte, span*4)
	b.Run("scalar", func(b *testing.B) {
		b.SetBytes(span * 4)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			scalarBlenderHspan(op, work, nil, 40, 60, 220, 160, span)
		}
	})
	b.Run("kernel", func(b *testing.B) {
		b.SetBytes(span * 4)
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			copy(work, template)
			kernel(work, nil, 40, 60, 220, 160, span)
		}
	})
}

func BenchmarkCompPlainSrcOver(b *testing.B) {
	benchPlainCompare(b, blender.CompOpSrcOver, simd.CompSrcOverPlainHspanRGBA)
}

func BenchmarkCompPlainXor(b *testing.B) {
	benchPlainCompare(b, blender.CompOpXor, simd.CompXorPlainHspanRGBA)
}
