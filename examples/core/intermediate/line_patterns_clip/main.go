// Package main ports AGG's line_patterns_clip.cpp demo (core patterned-line clipping).
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/order"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt/blender"
	"github.com/MeKo-Christian/agg_go/internal/primitives"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	outline "github.com/MeKo-Christian/agg_go/internal/renderer/outline"
)

type pathSourceAdapter struct {
	ps *path.PathStorageStl
}

func (a *pathSourceAdapter) Rewind(pathID uint32) {
	a.ps.Rewind(uint(pathID))
}

func (a *pathSourceAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

var linePatternChain = []uint32{
	16, 7,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff975252, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xa4c7a2a2, 0x3affff00, 0x3affff00, 0xff975151, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff954f4f, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
}

type lineChainPatternSource struct {
	data []uint32
}

func (s *lineChainPatternSource) Width() float64  { return float64(s.data[0]) }
func (s *lineChainPatternSource) Height() float64 { return float64(s.data[1]) }
func (s *lineChainPatternSource) Pixel(x, y int) color.RGBA {
	w := int(s.data[0])
	idx := y*w + x + 2
	if idx < 2 || idx >= len(s.data) {
		return color.NewRGBA(0, 0, 0, 0)
	}
	p := s.data[idx]
	c := color.NewRGBA(
		float64((p>>16)&0xFF)/255.0,
		float64((p>>8)&0xFF)/255.0,
		float64(p&0xFF)/255.0,
		float64((p>>24)&0xFF)/255.0,
	)
	c.Premultiply()
	return c
}

type lineImageBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]]
}

func rgbaToRGBA8(c color.RGBA) color.RGBA8[color.Linear] {
	clamp := func(v float64) uint8 {
		if v <= 0 {
			return 0
		}
		if v >= 1 {
			return 255
		}
		return uint8(v*255 + 0.5)
	}
	return color.RGBA8[color.Linear]{R: clamp(c.R), G: clamp(c.G), B: clamp(c.B), A: clamp(c.A)}
}

func (a *lineImageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *lineImageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

type lineOutlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *lineOutlineImageAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *lineOutlineImageAdapter) Color(c color.RGBA8[color.Linear]) {}

func lineNormals(lp primitives.LineParameters) (sx, sy, ex, ey int) {
	sx = lp.X1 + (lp.Y2 - lp.Y1)
	sy = lp.Y1 - (lp.X2 - lp.X1)
	ex = lp.X2 + (lp.Y2 - lp.Y1)
	ey = lp.Y2 - (lp.X2 - lp.X1)
	return
}

func (a *lineOutlineImageAdapter) Line0(lp primitives.LineParameters) { //nolint:gocritic // Interface requires value parameter.
	sx, sy, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineOutlineImageAdapter) Line1(lp primitives.LineParameters, sx, sy int) { //nolint:gocritic // Interface requires value parameter.
	_, _, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineOutlineImageAdapter) Line2(lp primitives.LineParameters, ex, ey int) { //nolint:gocritic // Interface requires value parameter.
	sx, sy, _, _ := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *lineOutlineImageAdapter) Pie(x, y, x1, y1, x2, y2 int)                 {}
func (a *lineOutlineImageAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {}
func (a *lineOutlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) { //nolint:gocritic // Interface requires value parameter.
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	const (
		scaleX = 1.0
		startX = 0.0
	)

	w := ctx.GetImage().Width()
	h := ctx.GetImage().Height()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 128, G: 191, B: 217, A: 255})

	patternSource := &lineChainPatternSource{data: linePatternChain}
	filter := outline.NewPatternFilterRGBAAdapter()
	scaledSrc := outline.NewLineImageScale(patternSource, 3.0)
	pattern := outline.NewLineImagePatternPow2(filter)
	pattern.Create(scaledSrc)

	renImg := outline.NewRendererOutlineImage(&lineImageBaseAdapter{renBase: renBase}, pattern)
	renImg.SetScaleX(scaleX)
	renImg.SetStartX(startX)
	rasImg := rasterizer.NewRasterizerOutlineAA[*lineOutlineImageAdapter, color.RGBA8[color.Linear]](&lineOutlineImageAdapter{ren: renImg})

	clipPad := 9.0
	renImg.ClipBox(50-clipPad, 50-clipPad, float64(w)-50+clipPad, float64(h)-50+clipPad)
	renBase.ClipBox(50, 50, w-50, h-50)

	ps := path.NewPathStorageStl()
	ps.MoveTo(20, 20)
	ps.LineTo(float64(w)-20, float64(h)-20)
	ps.LineTo(float64(w)-60, 20)
	ps.LineTo(40, float64(h)-40)
	ps.LineTo(100, 300)
	rasImg.AddPath(&pathSourceAdapter{ps: ps}, 0)

	renBase.ResetClipping(true)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Line Patterns Clip",
		Width:  500,
		Height: 500,
	}, &demo{})
}
