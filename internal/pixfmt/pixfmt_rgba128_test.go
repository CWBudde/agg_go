package pixfmt

import (
	"math"
	"testing"

	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/renderer"
)

const rgba128PfEps = 1e-6

// Compile-time check: the float RGBA pixfmts satisfy renderer.PixelFormat.
var (
	_ renderer.PixelFormat[color.RGBA32[color.Linear]] = (*PixFmtRGBA128)(nil)
	_ renderer.PixelFormat[color.RGBA32[color.Linear]] = (*PixFmtRGBA128Plain)(nil)
	_ renderer.PixelFormat[color.RGBA32[color.Linear]] = (*PixFmtRGBA128Pre)(nil)
)

// newRGBA128Buf builds a float RGBA buffer of w x h pixels (zeroed).
func newRGBA128Buf(w, h int) *buffer.RenderingBufferF32 {
	data := make([]float32, w*h*4)
	rbuf := buffer.NewRenderingBufferF32()
	rbuf.Attach(data, w, h, w*4*4) // stride in bytes = w * 4 channels * 4 bytes
	return rbuf
}

func pxApproxEq(a, b float32) bool {
	return math.Abs(float64(a)-float64(b)) <= rgba128PfEps
}

func wantPixel(t *testing.T, pf *PixFmtRGBA128Plain, x, y int, r, g, b, a float32) {
	t.Helper()
	c := pf.Pixel(x, y)
	if !pxApproxEq(c.R, r) || !pxApproxEq(c.G, g) || !pxApproxEq(c.B, b) || !pxApproxEq(c.A, a) {
		t.Fatalf("pixel(%d,%d) = {%v,%v,%v,%v}, want {%v,%v,%v,%v}",
			x, y, c.R, c.G, c.B, c.A, r, g, b, a)
	}
}

func TestPixFmtRGBA128BasicProps(t *testing.T) {
	pf := NewPixFmtRGBA128(newRGBA128Buf(4, 3))
	if pf.Width() != 4 || pf.Height() != 3 {
		t.Fatalf("Width/Height = %d/%d, want 4/3", pf.Width(), pf.Height())
	}
	if pf.PixWidth() != 16 {
		t.Fatalf("PixWidth = %d, want 16 (4 x float32)", pf.PixWidth())
	}
}

func TestPixFmtRGBA128PlainCopyPixelRoundTrip(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(4, 3))
	pf.CopyPixel(1, 1, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8))
	wantPixel(t, pf, 1, 1, 0.2, 0.4, 0.6, 0.8)
	// neighbors untouched
	wantPixel(t, pf, 0, 1, 0, 0, 0, 0)
	wantPixel(t, pf, 2, 1, 0, 0, 0, 0)
}

func TestPixFmtRGBA128PlainBlendOpaqueReplaces(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(2, 2))
	pf.BlendPixel(0, 0, color.NewRGBA32[color.Linear](1.0, 0.5, 0.25, 1.0), 255)
	wantPixel(t, pf, 0, 0, 1.0, 0.5, 0.25, 1.0)
}

func TestPixFmtRGBA128PlainCopyHline(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(4, 2))
	pf.CopyHline(0, 0, 3, color.NewRGBA32[color.Linear](0.2, 0.4, 0.6, 0.8))
	for x := range 3 {
		wantPixel(t, pf, x, 0, 0.2, 0.4, 0.6, 0.8)
	}
	wantPixel(t, pf, 3, 0, 0, 0, 0, 0) // outside run
}

func TestPixFmtRGBA128PlainBlendSolidHspanWithCovers(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(3, 1))
	covers := []uint8{255, 128, 0}
	pf.BlendSolidHspan(0, 0, 3, color.NewRGBA32[color.Linear](0.0, 0.0, 0.0, 1.0), covers)
	wantPixel(t, pf, 0, 0, 0, 0, 0, 1.0)              // full cover -> opaque
	wantPixel(t, pf, 1, 0, 0, 0, 0, float32(128)/255) // half cover -> alpha ~0.5
	wantPixel(t, pf, 2, 0, 0, 0, 0, 0)                // zero cover -> untouched
}

func TestPixFmtRGBA128PlainClear(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(3, 2))
	pf.Clear(color.NewRGBA32[color.Linear](0.1, 0.2, 0.3, 1.0))
	for y := range 2 {
		for x := range 3 {
			wantPixel(t, pf, x, y, 0.1, 0.2, 0.3, 1.0)
		}
	}
}

func TestPixFmtRGBA128OutOfBoundsCopyIsNoOp(t *testing.T) {
	pf := NewPixFmtRGBA128Plain(newRGBA128Buf(2, 2))
	// must not panic
	pf.CopyPixel(-1, 0, color.NewRGBA32[color.Linear](1, 1, 1, 1))
	pf.CopyPixel(0, 5, color.NewRGBA32[color.Linear](1, 1, 1, 1))
	pf.BlendPixel(5, 5, color.NewRGBA32[color.Linear](1, 1, 1, 1), 255)
	wantPixel(t, pf, 0, 0, 0, 0, 0, 0)
}
