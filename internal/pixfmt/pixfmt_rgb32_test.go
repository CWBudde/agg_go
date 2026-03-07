package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
)

func newRGB96Buf(w, h int) ([]float32, *buffer.RenderingBufferF32) {
	buf := make([]float32, w*h*3)
	rbuf := buffer.NewRenderingBufferF32WithData(buf, w, h, w*3*4)
	return buf, rbuf
}

func TestPixFmtRGB96Basic(t *testing.T) {
	_, rbuf := newRGB96Buf(4, 4)
	pf := NewPixFmtRGB96Linear(rbuf)

	if pf.Width() != 4 {
		t.Errorf("Width() expected 4, got %d", pf.Width())
	}
	if pf.Height() != 4 {
		t.Errorf("Height() expected 4, got %d", pf.Height())
	}
	if pf.PixWidth() != 12 {
		t.Errorf("PixWidth() expected 12, got %d", pf.PixWidth())
	}
}

func TestPixFmtRGB96CopyAndGetPixel(t *testing.T) {
	_, rbuf := newRGB96Buf(4, 4)
	pf := NewPixFmtRGB96Linear(rbuf)

	c := color.RGB32[color.Linear]{R: 1.0, G: 0.5, B: 0.25}
	pf.CopyPixel(2, 2, c)

	got := pf.GetPixel(2, 2)
	if got.R != 1.0 || got.G != 0.5 || got.B != 0.25 {
		t.Errorf("CopyPixel/GetPixel failed: got %+v want %+v", got, c)
	}

	// out of bounds returns zero
	oob := pf.GetPixel(-1, 0)
	if oob.R != 0 || oob.G != 0 || oob.B != 0 {
		t.Errorf("out-of-bounds GetPixel should return zero, got %+v", oob)
	}
}

func TestPixFmtRGB96BlendPixel(t *testing.T) {
	_, rbuf := newRGB96Buf(2, 1)
	pf := NewPixFmtRGB96Linear(rbuf)

	// set background to gray
	pf.CopyPixel(0, 0, color.RGB32[color.Linear]{R: 0.5, G: 0.5, B: 0.5})

	// blend red with half alpha, full coverage
	red := color.RGB32[color.Linear]{R: 1.0, G: 0.0, B: 0.0}
	pf.BlendPixel(0, 0, red, 0.5, 1.0)

	got := pf.GetPixel(0, 0)
	// blended value should be somewhere between 0.5 and 1.0 for R
	if got.R <= 0.5 || got.R > 1.0 {
		t.Errorf("BlendPixel R expected between 0.5 and 1.0, got %f", got.R)
	}
}

func TestPixFmtRGB96Clear(t *testing.T) {
	_, rbuf := newRGB96Buf(3, 3)
	pf := NewPixFmtRGB96Linear(rbuf)

	c := color.RGB32[color.Linear]{R: 0.1, G: 0.2, B: 0.3}
	pf.Clear(c)

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			got := pf.GetPixel(x, y)
			if got.R != c.R || got.G != c.G || got.B != c.B {
				t.Errorf("Clear (%d,%d): got %+v want %+v", x, y, got, c)
			}
		}
	}
}

func TestPixFmtRGB96Constructors(t *testing.T) {
	_, rbuf := newRGB96Buf(1, 1)
	_ = NewPixFmtBGR96Linear(rbuf)
	_ = NewPixFmtRGB96SRGB(rbuf)
	_ = NewPixFmtBGR96SRGB(rbuf)
	_ = NewPixFmtRGB96Pre(rbuf)
	_ = NewPixFmtBGR96Pre(rbuf)
	_ = NewPixFmtRGB96PreSRGB(rbuf)
	_ = NewPixFmtBGR96PreSRGB(rbuf)
}

func TestPixFmtRGB96BGROrder(t *testing.T) {
	_, rbuf := newRGB96Buf(1, 1)
	pf := NewPixFmtBGR96Linear(rbuf)

	c := color.RGB32[color.Linear]{R: 0.9, G: 0.5, B: 0.1}
	pf.CopyPixel(0, 0, c)
	got := pf.GetPixel(0, 0)
	if got.R != c.R || got.G != c.G || got.B != c.B {
		t.Errorf("BGR96 GetPixel round-trip failed: got %+v want %+v", got, c)
	}
}

func TestPixFmtRGB32PreLinearConstructors(t *testing.T) {
	buf := make([]basics.Int8u, 4*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 1, 1, 4)
	_ = NewPixFmtRGBX32PreLinear(rbuf)
	_ = NewPixFmtXRGB32PreLinear(rbuf)
	_ = NewPixFmtBGRX32PreLinear(rbuf)
	_ = NewPixFmtXBGR32PreLinear(rbuf)
}
