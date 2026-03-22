package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
)

// PixFmtRGBARendererAdaptor exposes an RGB pixfmt through an RGBA8-based
// renderer surface, forwarding source alpha into the wrapped pixfmt's explicit
// alpha parameter. This matches how AGG's 24-bit RGB pixfmts still use rgba8 as
// color_type at the renderer boundary.
type PixFmtRGBARendererAdaptor[S color.Space, B blender.RGBBlender[S]] struct {
	pixfmt *PixFmtAlphaBlendRGB[S, B]
}

func NewPixFmtRGBARendererAdaptor[S color.Space, B blender.RGBBlender[S]](pixfmt *PixFmtAlphaBlendRGB[S, B]) *PixFmtRGBARendererAdaptor[S, B] {
	return &PixFmtRGBARendererAdaptor[S, B]{pixfmt: pixfmt}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) Width() int    { return pa.pixfmt.Width() }
func (pa *PixFmtRGBARendererAdaptor[S, B]) Height() int   { return pa.pixfmt.Height() }
func (pa *PixFmtRGBARendererAdaptor[S, B]) PixWidth() int { return pa.pixfmt.PixWidth() }

func (pa *PixFmtRGBARendererAdaptor[S, B]) Pixel(x, y int) color.RGBA8[S] {
	c := pa.pixfmt.Pixel(x, y)
	return color.RGBA8[S]{R: c.R, G: c.G, B: c.B, A: 255}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyPixel(x, y int, c color.RGBA8[S]) {
	if c.A == 0 {
		return
	}
	if c.A == 255 {
		pa.pixfmt.CopyPixel(x, y, color.RGB8[S]{R: c.R, G: c.G, B: c.B})
		return
	}
	pa.pixfmt.BlendPixel(x, y, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, basics.CoverFull)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendPixel(x, y int, c color.RGBA8[S], cover basics.Int8u) {
	pa.pixfmt.BlendPixel(x, y, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyHline(x, y, length int, c color.RGBA8[S]) {
	if c.A == 255 {
		pa.pixfmt.CopyHline(x, y, x+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B})
		return
	}
	pa.pixfmt.BlendHline(x, y, x+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, basics.CoverFull)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyVline(x, y, length int, c color.RGBA8[S]) {
	if c.A == 255 {
		pa.pixfmt.CopyVline(x, y, y+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B})
		return
	}
	pa.pixfmt.BlendVline(x, y, y+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, basics.CoverFull)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendHline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
	pa.pixfmt.BlendHline(x, y, x+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendVline(x, y, length int, c color.RGBA8[S], cover basics.Int8u) {
	pa.pixfmt.BlendVline(x, y, y+length-1, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendSolidHspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
	pa.pixfmt.BlendSolidHspan(x, y, length, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, covers)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendSolidVspan(x, y, length int, c color.RGBA8[S], covers []basics.Int8u) {
	pa.pixfmt.BlendSolidVspan(x, y, length, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, covers)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyBar(x1, y1, x2, y2 int, c color.RGBA8[S]) {
	if c.A == 255 {
		pa.pixfmt.CopyBar(x1, y1, x2, y2, color.RGB8[S]{R: c.R, G: c.G, B: c.B})
		return
	}
	pa.pixfmt.BlendBar(x1, y1, x2, y2, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, basics.CoverFull)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendBar(x1, y1, x2, y2 int, c color.RGBA8[S], cover basics.Int8u) {
	pa.pixfmt.BlendBar(x1, y1, x2, y2, color.RGB8[S]{R: c.R, G: c.G, B: c.B}, c.A, cover)
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyColorHspan(x, y, length int, colors []color.RGBA8[S]) {
	for i := 0; i < length && i < len(colors); i++ {
		pa.CopyPixel(x+i, y, colors[i])
	}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) CopyColorVspan(x, y, length int, colors []color.RGBA8[S]) {
	for i := 0; i < length && i < len(colors); i++ {
		pa.CopyPixel(x, y+i, colors[i])
	}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendColorHspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if covers != nil && i < len(covers) {
			actualCover = covers[i]
		}
		pa.BlendPixel(x+i, y, colors[i], actualCover)
	}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) BlendColorVspan(x, y, length int, colors []color.RGBA8[S], covers []basics.Int8u, cover basics.Int8u) {
	for i := 0; i < length && i < len(colors); i++ {
		actualCover := cover
		if covers != nil && i < len(covers) {
			actualCover = covers[i]
		}
		pa.BlendPixel(x, y+i, colors[i], actualCover)
	}
}

func (pa *PixFmtRGBARendererAdaptor[S, B]) Clear(c color.RGBA8[S]) {
	pa.CopyBar(0, 0, pa.Width()-1, pa.Height()-1, c)
}
func (pa *PixFmtRGBARendererAdaptor[S, B]) Fill(c color.RGBA8[S]) {
	pa.BlendBar(0, 0, pa.Width()-1, pa.Height()-1, c, basics.CoverFull)
}
