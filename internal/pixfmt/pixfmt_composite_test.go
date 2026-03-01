package pixfmt

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt/blender"
)

type rgbaRowSource struct {
	pixels []color.RGBA8[color.Linear]
}

func (s *rgbaRowSource) GetPixel(x, y int) color.RGBA8[color.Linear] {
	if y != 0 || x < 0 || x >= len(s.pixels) {
		return color.RGBA8[color.Linear]{}
	}
	return s.pixels[x]
}

func (s *rgbaRowSource) Width() int  { return len(s.pixels) }
func (s *rgbaRowSource) Height() int { return 1 }

func TestPixFmtCompositeRGBA32PreBlendFrom(t *testing.T) {
	buf := make([]basics.Int8u, 3*4)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, 3, 1, 3*4)
	pf := NewPixFmtCompositeRGBA32Pre(rbuf, blender.CompOpSrcOver)

	src := &rgbaRowSource{
		pixels: []color.RGBA8[color.Linear]{
			{R: 128, G: 0, B: 0, A: 128},
			{R: 0, G: 128, B: 0, A: 128},
			{R: 0, G: 0, B: 128, A: 128},
		},
	}

	pf.BlendFrom(src, 0, 0, 0, 0, 3, basics.CoverFull)

	for x, want := range src.pixels {
		if got := pf.GetPixel(x, 0); got != want {
			t.Fatalf("pixel %d: got %+v want %+v", x, got, want)
		}
	}
}
