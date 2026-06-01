package pixfmt

import (
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
)

func newCompF32Buf(w, h int) *buffer.RenderingBufferF32 {
	rbuf := buffer.NewRenderingBufferF32()
	rbuf.Attach(make([]float32, w*h*4), w, h, w*4*4)
	return rbuf
}

// TestPixFmtCompositeRGBA128Multiply checks BlendSolidHspan composites with the
// selected operator (multiply over an opaque destination).
func TestPixFmtCompositeRGBA128Multiply(t *testing.T) {
	rbuf := newCompF32Buf(4, 1)
	pf := NewPixFmtCompositeRGBA128Linear(rbuf, blender.CompOpMultiply)
	// Opaque destination color (premultiplied == straight since a=1).
	pf.Clear(color.RGBA32[color.Linear]{R: 0.4, G: 0.5, B: 0.6, A: 1})

	pf.BlendSolidHspan(0, 0, 4, color.RGBA32[color.Linear]{R: 0.5, G: 0.5, B: 0.5, A: 1}, nil)

	got := pf.GetPixel(0, 0)
	if math.Abs(float64(got.R-0.2)) > 1e-4 || math.Abs(float64(got.G-0.25)) > 1e-4 || math.Abs(float64(got.B-0.3)) > 1e-4 {
		t.Fatalf("multiply span = %v, want {0.2,0.25,0.3,1}", got)
	}
}

// TestPixFmtCompositeRGBA128SetCompOp verifies the operator can be swapped while
// preserving the premultiplied-vs-straight source convention.
func TestPixFmtCompositeRGBA128SetCompOp(t *testing.T) {
	rbuf := newCompF32Buf(2, 1)
	pf := NewPixFmtCompositeRGBA128Linear(rbuf, blender.CompOpSrcOver)
	if pf.GetCompOp() != blender.CompOpSrcOver {
		t.Fatalf("initial op = %v, want SrcOver", pf.GetCompOp())
	}
	pf.SetCompOp(blender.CompOpScreen)
	if pf.GetCompOp() != blender.CompOpScreen {
		t.Fatalf("op after SetCompOp = %v, want Screen", pf.GetCompOp())
	}
}

// TestPixFmtCompositeRGBA128Clear fills without compositing.
func TestPixFmtCompositeRGBA128Clear(t *testing.T) {
	rbuf := newCompF32Buf(3, 2)
	pf := NewPixFmtCompositeRGBA128Linear(rbuf, blender.CompOpSrcOver)
	pf.Clear(color.RGBA32[color.Linear]{R: 0.1, G: 0.2, B: 0.3, A: 0.4})
	for y := range 2 {
		for x := range 3 {
			c := pf.GetPixel(x, y)
			if c.R != 0.1 || c.A != 0.4 {
				t.Fatalf("clear pixel (%d,%d) = %v", x, y, c)
			}
		}
	}
}
