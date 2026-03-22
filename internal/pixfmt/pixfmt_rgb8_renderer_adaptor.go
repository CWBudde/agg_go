package pixfmt

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
)

// PixFmtRGBRendererAdaptor exposes RGB24-style pixfmts through the
// renderer.PixelFormat-compatible surface expected by RendererBase and the
// alpha-mask adaptor. It uses opaque source alpha, which matches the current
// RGB8 color model in this Go port.
type PixFmtRGBRendererAdaptor[S color.Space, B blender.RGBBlender[S]] struct {
	pixfmt *PixFmtAlphaBlendRGB[S, B]
}

func NewPixFmtRGBRendererAdaptor[S color.Space, B blender.RGBBlender[S]](pixfmt *PixFmtAlphaBlendRGB[S, B]) *PixFmtRGBRendererAdaptor[S, B] {
	return &PixFmtRGBRendererAdaptor[S, B]{pixfmt: pixfmt}
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) Width() int    { return pa.pixfmt.Width() }
func (pa *PixFmtRGBRendererAdaptor[S, B]) Height() int   { return pa.pixfmt.Height() }
func (pa *PixFmtRGBRendererAdaptor[S, B]) PixWidth() int { return pa.pixfmt.PixWidth() }
func (pa *PixFmtRGBRendererAdaptor[S, B]) Pixel(x, y int) color.RGB8[S] {
	return pa.pixfmt.Pixel(x, y)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyPixel(x, y int, c color.RGB8[S]) {
	pa.pixfmt.CopyPixel(x, y, c)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendPixel(x, y int, c color.RGB8[S], cover basics.Int8u) {
	pa.pixfmt.BlendPixel(x, y, c, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyHline(x, y, length int, c color.RGB8[S]) {
	pa.pixfmt.CopyHline(x, y, x+length-1, c)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyVline(x, y, length int, c color.RGB8[S]) {
	pa.pixfmt.CopyVline(x, y, y+length-1, c)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendHline(x, y, length int, c color.RGB8[S], cover basics.Int8u) {
	pa.pixfmt.BlendHline(x, y, x+length-1, c, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendVline(x, y, length int, c color.RGB8[S], cover basics.Int8u) {
	pa.pixfmt.BlendVline(x, y, y+length-1, c, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendSolidHspan(x, y, length int, c color.RGB8[S], covers []basics.Int8u) {
	pa.pixfmt.BlendSolidHspan(x, y, length, c, basics.CoverFull, covers)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendSolidVspan(x, y, length int, c color.RGB8[S], covers []basics.Int8u) {
	pa.pixfmt.BlendSolidVspan(x, y, length, c, basics.CoverFull, covers)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyBar(x1, y1, x2, y2 int, c color.RGB8[S]) {
	pa.pixfmt.CopyBar(x1, y1, x2, y2, c)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendBar(x1, y1, x2, y2 int, c color.RGB8[S], cover basics.Int8u) {
	pa.pixfmt.BlendBar(x1, y1, x2, y2, c, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyColorHspan(x, y, length int, colors []color.RGB8[S]) {
	pa.pixfmt.CopyColorHspan(x, y, length, colors)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) CopyColorVspan(x, y, length int, colors []color.RGB8[S]) {
	pa.pixfmt.CopyColorVspan(x, y, length, colors)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendColorHspan(x, y, length int, colors []color.RGB8[S], covers []basics.Int8u, cover basics.Int8u) {
	pa.pixfmt.BlendColorHspan(x, y, length, colors, covers, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) BlendColorVspan(x, y, length int, colors []color.RGB8[S], covers []basics.Int8u, cover basics.Int8u) {
	pa.pixfmt.BlendColorVspan(x, y, length, colors, covers, basics.CoverFull, cover)
}

func (pa *PixFmtRGBRendererAdaptor[S, B]) Clear(c color.RGB8[S]) { pa.pixfmt.Clear(c) }
func (pa *PixFmtRGBRendererAdaptor[S, B]) Fill(c color.RGB8[S])  { pa.pixfmt.Fill(c) }
