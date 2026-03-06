package main

import (
	"fmt"
	"math"
	"os"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	ctrltext "agg_go/internal/ctrl/text"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
	"agg_go/internal/primitives"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	outline "agg_go/internal/renderer/outline"
	rprimitives "agg_go/internal/renderer/primitives"
	"agg_go/internal/scanline"
)

const (
	frameWidth  = 500
	frameHeight = 450
)

var pixmapChain = []uint32{
	16, 7,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff975252, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xa4c7a2a2, 0x3affff00, 0x3affff00, 0xff975151, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000,
	0x00ffffff, 0x5ae0cccc, 0xffa46767, 0xff660000, 0xff954f4f, 0x7ed4b8b8, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0x5ae0cccc, 0xa8c6a0a0, 0xff7f2929, 0xff670202, 0x9ecaa6a6, 0x5ae0cccc, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x0cfbf9f9, 0xff9a5757, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xff660000, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff,
	0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0xb4c29999, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xff9a5757, 0xb4c29999, 0x00ffffff, 0x00ffffff, 0x00ffffff, 0x00ffffff,
}

type convVertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

type convToRasAdapter struct {
	src convVertexSource
}

func (a *convToRasAdapter) Rewind(pathID uint32) {
	a.src.Rewind(uint(pathID))
}

func (a *convToRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

type rasScanlineAdapter struct {
	sl *scanline.ScanlineU8
}

func (a *rasScanlineAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *rasScanlineAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *rasScanlineAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *rasScanlineAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *rasScanlineAdapter) NumSpans() int  { return a.sl.NumSpans() }

type spiral struct {
	x, y                 float64
	r1, r2               float64
	step, startAngle     float64
	angle, currR, da, dr float64
	start                bool
}

func newSpiral(x, y, r1, r2, step, startAngle float64) *spiral {
	return &spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         basics.Deg2RadF(8.0),
		dr:         step / 45.0,
	}
}

func (s *spiral) Rewind(pathID uint) {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

func (s *spiral) Vertex() (x, y float64, cmd basics.PathCommand) {
	if s.currR > s.r2 {
		return 0, 0, basics.PathCmdStop
	}

	x = s.x + math.Cos(s.angle)*s.currR
	y = s.y + math.Sin(s.angle)*s.currR
	s.currR += s.dr
	s.angle += s.da
	if s.start {
		s.start = false
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}

type roundoffSource struct {
	src convVertexSource
}

func (r *roundoffSource) Rewind(pathID uint) {
	r.src.Rewind(pathID)
}

func (r *roundoffSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = r.src.Vertex()
	if basics.IsVertex(cmd) {
		x = math.Floor(x)
		y = math.Floor(y)
	}
	return x, y, cmd
}

type chainPatternSource struct {
	data []uint32
}

func (s *chainPatternSource) Width() float64  { return float64(s.data[0]) }
func (s *chainPatternSource) Height() float64 { return float64(s.data[1]) }

func (s *chainPatternSource) Pixel(x, y int) color.RGBA {
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

type outlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *outlineAAAdapter) AccurateJoinOnly() bool             { return a.ren.AccurateJoinOnly() }
func (a *outlineAAAdapter) Color(c color.RGBA8[color.Linear])  { a.ren.Color(c) }
func (a *outlineAAAdapter) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *outlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *outlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *outlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *outlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int) { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *outlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}

type imageBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]]
}

type outlineBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]]
}

func (a *outlineBaseAdapter) Width() int { return a.renBase.Width() }

func (a *outlineBaseAdapter) Height() int { return a.renBase.Height() }

func (a *outlineBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidHspan(x, y, length, c, convCovers)
}

func (a *outlineBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	convCovers := make([]basics.Int8u, len(covers))
	for i := range covers {
		convCovers[i] = basics.Int8u(covers[i])
	}
	a.renBase.BlendSolidVspan(x, y, length, c, convCovers)
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

func (a *imageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *imageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

type outlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *outlineImageAdapter) AccurateJoinOnly() bool                         { return a.ren.AccurateJoinOnly() }
func (a *outlineImageAdapter) Color(c color.RGBA8[color.Linear])              {}
func (a *outlineImageAdapter) Line0(lp primitives.LineParameters)             {}
func (a *outlineImageAdapter) Line1(lp primitives.LineParameters, sx, sy int) {}
func (a *outlineImageAdapter) Line2(lp primitives.LineParameters, ex, ey int) {}
func (a *outlineImageAdapter) Pie(x, y, x1, y1, x2, y2 int)                   {}
func (a *outlineImageAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int)   {}
func (a *outlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func renderSolidPath(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	vs rasterizer.VertexSource,
	col color.RGBA8[color.Linear],
) {
	ras.Reset()
	ras.AddPath(vs, 0)
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
		y := sl.Y()
		for _, spanData := range sl.Spans() {
			if spanData.Len > 0 {
				renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), col, spanData.Covers)
			}
		}
	}
}

func drawText(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]],
	x, y float64,
	lines []string,
) {
	for i, line := range lines {
		if line == "" {
			continue
		}
		t := ctrltext.NewSimpleText()
		t.SetSize(8)
		t.SetThickness(0.7)
		t.SetText(line)
		t.SetPosition(x, y+float64(i)*12)
		renderSolidPath(ras, sl, renBase, &convToRasAdapter{src: t}, color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
	}
}

func savePPM(filename string, imgData []uint8, width, height int) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "P6\n%d %d\n255\n", width, height); err != nil {
		return err
	}

	for i := 0; i < len(imgData); i += 4 {
		if _, err := f.Write([]byte{imgData[i], imgData[i+1], imgData[i+2]}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	imgData := make([]uint8, frameWidth*frameHeight*4)
	rbuf := buffer.NewRenderingBufferU8WithData(imgData, frameWidth, frameHeight, frameWidth*4)

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]](pf)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	rasAA := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	renPrim := rprimitives.NewRendererPrimitives[*renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](renBase)
	rasAliased := rasterizer.NewRasterizerOutline[*rprimitives.RendererPrimitives[*renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](renPrim)

	profile := outline.NewLineProfileAA()
	profile.Width(3.0)
	renOutlineAA := outline.NewRendererOutlineAA[*outlineBaseAdapter, color.RGBA8[color.Linear]](&outlineBaseAdapter{renBase: renBase}, profile)
	rasOutlineAA := rasterizer.NewRasterizerOutlineAA[*outlineAAAdapter, color.RGBA8[color.Linear]](&outlineAAAdapter{ren: renOutlineAA})
	rasOutlineAA.SetRoundCap(true)
	rasOutlineAA.SetLineJoin(rasterizer.OutlineRoundJoin)

	patternSource := &chainPatternSource{data: pixmapChain}
	filter := outline.NewPatternFilterRGBAAdapter()
	scaledSrc := outline.NewLineImageScale(patternSource, 3.0)
	pattern := outline.NewLineImagePatternPow2(filter)
	pattern.Create(scaledSrc)
	renImg := outline.NewRendererOutlineImage(&imageBaseAdapter{renBase: renBase}, pattern)
	renImg.SetScaleX(3.0 / patternSource.Height())
	rasImg := rasterizer.NewRasterizerOutlineAA[*outlineImageAdapter, color.RGBA8[color.Linear]](&outlineImageAdapter{ren: renImg})

	brown := color.RGBA8[color.Linear]{R: 102, G: 77, B: 26, A: 255}

	s1 := &roundoffSource{src: newSpiral(float64(frameWidth)/5, float64(frameHeight)/4+50, 5, 70, 16, 0)}
	renPrim.LineColor(brown)
	rasAliased.AddPath(&convToRasAdapter{src: s1}, 0)

	s2 := newSpiral(float64(frameWidth)/2, float64(frameHeight)/4+50, 5, 70, 16, 0)
	renPrim.LineColor(brown)
	rasAliased.AddPath(&convToRasAdapter{src: s2}, 0)

	s3 := newSpiral(float64(frameWidth)/5, float64(frameHeight)-float64(frameHeight)/4+20, 5, 70, 16, 0)
	renOutlineAA.Color(brown)
	rasOutlineAA.AddPath(&convToRasAdapter{src: s3}, 0)

	s4 := newSpiral(float64(frameWidth)/2, float64(frameHeight)-float64(frameHeight)/4+20, 5, 70, 16, 0)
	stroke := conv.NewConvStroke(s4)
	stroke.SetWidth(3.0)
	stroke.SetLineCap(basics.RoundCap)
	renderSolidPath(rasAA, sl, renBase, &convToRasAdapter{src: stroke}, brown)

	s5 := newSpiral(float64(frameWidth)-float64(frameWidth)/5, float64(frameHeight)-float64(frameHeight)/4+20, 5, 70, 16, 0)
	rasImg.AddPath(&convToRasAdapter{src: s5}, 0)

	drawText(rasAA, sl, renBase, 50, 80, []string{"Bresenham lines,", "regular accuracy"})
	drawText(rasAA, sl, renBase, float64(frameWidth)/2-50, 80, []string{"Bresenham lines,", "subpixel accuracy"})
	drawText(rasAA, sl, renBase, 50, float64(frameHeight)/2+50, []string{"Anti-aliased lines"})
	drawText(rasAA, sl, renBase, float64(frameWidth)/2-50, float64(frameHeight)/2+50, []string{"Scanline rasterizer"})
	drawText(rasAA, sl, renBase, float64(frameWidth)-float64(frameWidth)/5-50, float64(frameHeight)/2+50, []string{"Arbitrary Image Pattern"})

	if err := savePPM("rasterizers2_demo.ppm", imgData, frameWidth, frameHeight); err != nil {
		panic(err)
	}

	fmt.Println("rasterizers2_demo.ppm")
}
