package pixfmt

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

// Focused microbenchmarks for the straight-alpha composite pixfmt's solid-span
// blend, isolated from render setup (no context/rasterizer). They measure the
// per-span cost of BlendSolidHspan, which now takes the blender's
// BlendSolidSpanStraight fast path (one concrete call per span instead of one
// interface dispatch per pixel; bit-exact with the per-pixel bridge). The
// scalar-vs-span comparison lives in blender/rgba_composite_span_test.go; here we
// track the end-to-end pixfmt cost. Subtract benchCompCopyOnly from each op to get
// the isolated blend cost, divide by benchCompSpan for ns/px.
//
// The destination row is reset to an opaque template each iteration (so iterative
// operators like dst-out don't decay alpha to zero and stop being representative);
// benchCompCopyOnly measures that reset alone so it can be subtracted out.

const benchCompSpan = 256

func benchCompSetup() (pf *PixFmtCompositeRGBA32, work, template []basics.Int8u) {
	work = make([]basics.Int8u, benchCompSpan*4)
	template = make([]basics.Int8u, benchCompSpan*4)
	// Opaque green destination, straight alpha.
	for i := 0; i < benchCompSpan; i++ {
		p := i * 4
		template[p+0], template[p+1], template[p+2], template[p+3] = 40, 130, 60, 255
	}
	rbuf := buffer.NewRenderingBufferU8WithData(work, benchCompSpan, 1, benchCompSpan*4)
	pf = NewPixFmtCompositeRGBA32(rbuf, blender.CompOpSrcOver)
	return pf, work, template
}

func benchCompBlend(b *testing.B, op blender.CompOp, covers []basics.Int8u) {
	pf, work, template := benchCompSetup()
	pf.SetCompOp(op)
	src := color.RGBA8[color.Linear]{R: 40, G: 60, B: 220, A: 160} // translucent source
	b.SetBytes(int64(benchCompSpan * 4))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		copy(work, template)
		pf.BlendSolidHspan(0, 0, benchCompSpan, src, covers)
	}
}

// benchCompCopyOnly measures the per-iteration destination reset (copy) so it can
// be subtracted from the op benchmarks to isolate the blend cost.
func BenchmarkCompCopyOnly(b *testing.B) {
	_, work, template := benchCompSetup()
	b.SetBytes(int64(benchCompSpan * 4))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		copy(work, template)
	}
}

func BenchmarkCompSolidHspanFullCover(b *testing.B) {
	ops := []struct {
		name string
		op   blender.CompOp
	}{
		{"src_over", blender.CompOpSrcOver},
		{"xor", blender.CompOpXor},
		{"dst_out", blender.CompOpDstOut},
		{"multiply", blender.CompOpMultiply},
	}
	for _, o := range ops {
		b.Run(o.name, func(b *testing.B) { benchCompBlend(b, o.op, nil) })
	}
}

func BenchmarkCompSolidHspanAACover(b *testing.B) {
	// Mostly-full span with anti-aliased ramps at both ends (the realistic shape
	// of a rasterized solid span), so the cover != 0/255 paths are exercised.
	covers := make([]basics.Int8u, benchCompSpan)
	for i := range covers {
		switch {
		case i < 8:
			covers[i] = basics.Int8u((i + 1) * 28)
		case i >= benchCompSpan-8:
			covers[i] = basics.Int8u((benchCompSpan - i) * 28)
		default:
			covers[i] = 255
		}
	}
	ops := []struct {
		name string
		op   blender.CompOp
	}{
		{"src_over", blender.CompOpSrcOver},
		{"xor", blender.CompOpXor},
		{"dst_out", blender.CompOpDstOut},
		{"multiply", blender.CompOpMultiply},
	}
	for _, o := range ops {
		b.Run(o.name, func(b *testing.B) { benchCompBlend(b, o.op, covers) })
	}
}
