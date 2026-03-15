package main

import (
	"math"
	"strconv"
	"strings"

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

// Port of AGG C++ line_patterns_clip.cpp (web variant).
const (
	linePatternClipPadMin     = 36.0
	linePatternClipPadMax     = 84.0
	linePatternClipPointHit   = 20.0
	linePatternClipSegmentHit = 14.0
	linePatternClipBaseWidth  = 500.0
	linePatternClipBaseHeight = 500.0
)

const (
	linePatternClipDragNone = iota
	linePatternClipDragPoint
	linePatternClipDragChain
)

var (
	linePatternClipScaleX = 1.0
	linePatternClipStartX = 0.0

	linePatternClipDefaultPoints = [][2]float64{
		{20, 20},
		{480, 480},
		{440, 20},
		{40, 460},
		{100, 300},
	}

	linePatternClipPoints            [][2]float64
	linePatternClipPointsInitialized bool
	linePatternClipPointsCustomized  bool
	linePatternClipCanvasW           int
	linePatternClipCanvasH           int

	linePatternClipDragMode      = linePatternClipDragNone
	linePatternClipSelectedPoint = -1
	linePatternClipDragLastX     float64
	linePatternClipDragLastY     float64
)

func setLinePatternClipScaleX(v float64) {
	if v < 0.2 {
		v = 0.2
	}
	if v > 3.0 {
		v = 3.0
	}
	linePatternClipScaleX = v
}

func setLinePatternClipStartX(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 10 {
		v = 10
	}
	linePatternClipStartX = v
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
	c := color.NewRGBAFromRGBA8(
		uint8((p>>16)&0xFF),
		uint8((p>>8)&0xFF),
		uint8(p&0xFF),
		uint8((p>>24)&0xFF),
	)
	c.Premultiply()
	return c
}

type lineClipImageBaseAdapter struct {
	renBase *renderer.RendererBase[*pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8Pre[color.Linear, order.RGBA]], color.RGBA8[color.Linear]]
}

type lineClipSolidBaseAdapter struct {
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

func (a *lineClipImageBaseAdapter) BlendColorHSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorHspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *lineClipImageBaseAdapter) BlendColorVSpan(x, y, length int, colors []color.RGBA, covers []basics.CoverType) {
	buf := make([]color.RGBA8[color.Linear], len(colors))
	for i := range colors {
		buf[i] = rgbaToRGBA8(colors[i])
	}
	a.renBase.BlendColorVspan(x, y, length, buf, nil, basics.CoverFull)
}

func (a *lineClipSolidBaseAdapter) Width() int  { return a.renBase.Width() }
func (a *lineClipSolidBaseAdapter) Height() int { return a.renBase.Height() }
func (a *lineClipSolidBaseAdapter) BlendSolidHSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	a.renBase.BlendSolidHspan(x, y, length, c, covers)
}
func (a *lineClipSolidBaseAdapter) BlendSolidVSpan(x, y, length int, c color.RGBA8[color.Linear], covers []basics.CoverType) {
	a.renBase.BlendSolidVspan(x, y, length, c, covers)
}

type lineClipOutlineImageAdapter struct {
	ren *outline.RendererOutlineImage
}

func (a *lineClipOutlineImageAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *lineClipOutlineImageAdapter) Color(c color.RGBA8[color.Linear]) {}

func lineNormals(lp primitives.LineParameters) (sx, sy, ex, ey int) {
	sx = lp.X1 + (lp.Y2 - lp.Y1)
	sy = lp.Y1 - (lp.X2 - lp.X1)
	ex = lp.X2 + (lp.Y2 - lp.Y1)
	ey = lp.Y2 - (lp.X2 - lp.X1)
	return
}

func (a *lineClipOutlineImageAdapter) Line0(lp primitives.LineParameters) {
	sx, sy, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineClipOutlineImageAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	_, _, ex, ey := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func (a *lineClipOutlineImageAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	sx, sy, _, _ := lineNormals(lp)
	a.ren.Line3(&lp, sx, sy, ex, ey)
}
func (a *lineClipOutlineImageAdapter) Pie(x, y, x1, y1, x2, y2 int)                 {}
func (a *lineClipOutlineImageAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {}
func (a *lineClipOutlineImageAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

type lineClipOutlineAAAdapter struct {
	ren *outline.RendererOutlineAA[*lineClipSolidBaseAdapter, color.RGBA8[color.Linear]]
}

func (a *lineClipOutlineAAAdapter) AccurateJoinOnly() bool            { return a.ren.AccurateJoinOnly() }
func (a *lineClipOutlineAAAdapter) Color(c color.RGBA8[color.Linear]) { a.ren.Color(c) }
func (a *lineClipOutlineAAAdapter) Pie(x, y, x1, y1, x2, y2 int)      { a.ren.Pie(x, y, x1, y1, x2, y2) }
func (a *lineClipOutlineAAAdapter) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	a.ren.Semidot(cmp, x, y, x1, y1)
}
func (a *lineClipOutlineAAAdapter) Line0(lp primitives.LineParameters) { a.ren.Line0(&lp) }
func (a *lineClipOutlineAAAdapter) Line1(lp primitives.LineParameters, sx, sy int) {
	a.ren.Line1(&lp, sx, sy)
}
func (a *lineClipOutlineAAAdapter) Line2(lp primitives.LineParameters, ex, ey int) {
	a.ren.Line2(&lp, ex, ey)
}
func (a *lineClipOutlineAAAdapter) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	a.ren.Line3(&lp, sx, sy, ex, ey)
}

func linePatternClipPadding(w, h int) float64 {
	pad := math.Min(float64(w), float64(h)) * 0.1
	if pad < linePatternClipPadMin {
		return linePatternClipPadMin
	}
	if pad > linePatternClipPadMax {
		return linePatternClipPadMax
	}
	return pad
}

func stretchLinePatternClipPoints(w, h int) [][2]float64 {
	pad := linePatternClipPadding(w, h)
	availW := math.Max(1, float64(w)-2*pad)
	availH := math.Max(1, float64(h)-2*pad)
	scale := math.Min(availW/linePatternClipBaseWidth, availH/linePatternClipBaseHeight)
	offX := (float64(w) - linePatternClipBaseWidth*scale) * 0.5
	offY := (float64(h) - linePatternClipBaseHeight*scale) * 0.5

	points := make([][2]float64, len(linePatternClipDefaultPoints))
	for i, pt := range linePatternClipDefaultPoints {
		points[i] = [2]float64{
			offX + pt[0]*scale,
			offY + pt[1]*scale,
		}
	}
	return points
}

func ensureLinePatternClipPoints() {
	w, h := ctx.Width(), ctx.Height()
	if !linePatternClipPointsInitialized ||
		(!linePatternClipPointsCustomized && (linePatternClipCanvasW != w || linePatternClipCanvasH != h)) {
		linePatternClipPoints = stretchLinePatternClipPoints(w, h)
		linePatternClipPointsInitialized = true
		linePatternClipCanvasW = w
		linePatternClipCanvasH = h
	}
}

func linePatternClipClamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampLinePatternClipPoint(x, y float64) (float64, float64) {
	if ctx == nil {
		return x, y
	}
	pad := math.Max(8, linePatternClipPadding(ctx.Width(), ctx.Height())*0.2)
	maxX := math.Max(pad, float64(ctx.Width())-pad)
	maxY := math.Max(pad, float64(ctx.Height())-pad)
	return linePatternClipClamp(x, pad, maxX), linePatternClipClamp(y, pad, maxY)
}

func linePatternClipDistanceSquared(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

func linePatternClipSegmentDistanceSquared(px, py, ax, ay, bx, by float64) float64 {
	abx := bx - ax
	aby := by - ay
	den := abx*abx + aby*aby
	if den <= 0 {
		return linePatternClipDistanceSquared(px, py, ax, ay)
	}
	t := ((px-ax)*abx + (py-ay)*aby) / den
	t = linePatternClipClamp(t, 0, 1)
	cx := ax + abx*t
	cy := ay + aby*t
	return linePatternClipDistanceSquared(px, py, cx, cy)
}

func linePatternClipNearestSegment(x, y float64) (int, float64) {
	bestIdx := -1
	bestDist2 := math.MaxFloat64
	for i := 0; i+1 < len(linePatternClipPoints); i++ {
		a := linePatternClipPoints[i]
		b := linePatternClipPoints[i+1]
		d2 := linePatternClipSegmentDistanceSquared(x, y, a[0], a[1], b[0], b[1])
		if d2 < bestDist2 {
			bestDist2 = d2
			bestIdx = i
		}
	}
	return bestIdx, bestDist2
}

func encodeLinePatternClipPoints() string {
	ensureLinePatternClipPoints()
	parts := make([]string, len(linePatternClipPoints))
	for i, pt := range linePatternClipPoints {
		parts[i] = strings.Join([]string{
			strconv.FormatFloat(pt[0], 'f', 1, 64),
			strconv.FormatFloat(pt[1], 'f', 1, 64),
		}, ",")
	}
	return strings.Join(parts, ";")
}

func setLinePatternClipPointsEncoded(encoded string) bool {
	chunks := strings.Split(encoded, ";")
	if len(chunks) != len(linePatternClipDefaultPoints) {
		return false
	}

	points := make([][2]float64, len(linePatternClipDefaultPoints))
	for i, chunk := range chunks {
		fields := strings.Split(chunk, ",")
		if len(fields) != 2 {
			return false
		}
		x, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return false
		}
		y, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return false
		}
		points[i] = [2]float64{x, y}
	}

	linePatternClipPoints = points
	linePatternClipPointsInitialized = true
	linePatternClipPointsCustomized = true
	if ctx != nil {
		linePatternClipCanvasW = ctx.Width()
		linePatternClipCanvasH = ctx.Height()
	}
	return true
}

func handleLinePatternsClipMouseDown(x, y float64) bool {
	ensureLinePatternClipPoints()

	bestIdx := -1
	bestDist2 := linePatternClipPointHit * linePatternClipPointHit
	for i, pt := range linePatternClipPoints {
		d2 := linePatternClipDistanceSquared(x, y, pt[0], pt[1])
		if d2 <= bestDist2 {
			bestDist2 = d2
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		linePatternClipDragMode = linePatternClipDragPoint
		linePatternClipSelectedPoint = bestIdx
		linePatternClipPointsCustomized = true
		return true
	}

	if _, dist2 := linePatternClipNearestSegment(x, y); dist2 <= linePatternClipSegmentHit*linePatternClipSegmentHit {
		linePatternClipDragMode = linePatternClipDragChain
		linePatternClipSelectedPoint = -1
		linePatternClipDragLastX = x
		linePatternClipDragLastY = y
		linePatternClipPointsCustomized = true
		return true
	}

	return false
}

func handleLinePatternsClipMouseMove(x, y float64) bool {
	switch linePatternClipDragMode {
	case linePatternClipDragPoint:
		if linePatternClipSelectedPoint < 0 {
			return false
		}
		linePatternClipPoints[linePatternClipSelectedPoint][0], linePatternClipPoints[linePatternClipSelectedPoint][1] =
			clampLinePatternClipPoint(x, y)
		return true
	case linePatternClipDragChain:
		dx := x - linePatternClipDragLastX
		dy := y - linePatternClipDragLastY
		for i := range linePatternClipPoints {
			linePatternClipPoints[i][0] += dx
			linePatternClipPoints[i][1] += dy
			linePatternClipPoints[i][0], linePatternClipPoints[i][1] =
				clampLinePatternClipPoint(linePatternClipPoints[i][0], linePatternClipPoints[i][1])
		}
		linePatternClipDragLastX = x
		linePatternClipDragLastY = y
		return true
	default:
		return false
	}
}

func handleLinePatternsClipMouseUp() {
	linePatternClipDragMode = linePatternClipDragNone
	linePatternClipSelectedPoint = -1
}

func buildLinePatternsClipPath() *path.PathStorageStl {
	ps := path.NewPathStorageStl()
	ps.MoveTo(linePatternClipPoints[0][0], linePatternClipPoints[0][1])
	for i := 1; i < len(linePatternClipPoints); i++ {
		ps.LineTo(linePatternClipPoints[i][0], linePatternClipPoints[i][1])
	}
	return ps
}

func drawLinePatternsClipDemo() {
	ensureLinePatternClipPoints()

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

	renImg := outline.NewRendererOutlineImage(&lineClipImageBaseAdapter{renBase: renBase}, pattern)
	renImg.SetScaleX(linePatternClipScaleX)
	renImg.SetStartX(linePatternClipStartX)
	rasImg := rasterizer.NewRasterizerOutlineAA[*lineClipOutlineImageAdapter, color.RGBA8[color.Linear]](&lineClipOutlineImageAdapter{ren: renImg})
	rasImg.SetRoundCap(true)

	profile := outline.NewLineProfileAA()
	profile.SmootherWidth(10.0)
	profile.Width(8.0)
	renLine := outline.NewRendererOutlineAA[*lineClipSolidBaseAdapter, color.RGBA8[color.Linear]](&lineClipSolidBaseAdapter{renBase: renBase}, profile)
	renLine.Color(color.RGBA8[color.Linear]{R: 0, G: 0, B: 127, A: 255})
	rasLine := rasterizer.NewRasterizerOutlineAA[*lineClipOutlineAAAdapter, color.RGBA8[color.Linear]](&lineClipOutlineAAAdapter{ren: renLine})
	rasLine.SetRoundCap(true)

	clipInset := int(math.Round(linePatternClipPadding(width, height)))
	clipPad := 9.0
	renImg.ClipBox(
		float64(clipInset)-clipPad,
		float64(clipInset)-clipPad,
		float64(width-clipInset)+clipPad,
		float64(height-clipInset)+clipPad,
	)
	renLine.ClipBox(
		float64(clipInset)-clipPad,
		float64(clipInset)-clipPad,
		float64(width-clipInset)+clipPad,
		float64(height-clipInset)+clipPad,
	)

	ps := buildLinePatternsClipPath()
	rasLine.AddPath(&pathSourceAdapter{ps: ps}, 0)
	rasImg.AddPath(&pathSourceAdapter{ps: ps}, 0)

	renBase.BlendBar(
		0,
		0,
		width-1,
		height-1,
		color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255},
		200,
	)

	renBase.ClipBox(clipInset, clipInset, width-clipInset, height-clipInset)
	ps = buildLinePatternsClipPath()
	rasLine.AddPath(&pathSourceAdapter{ps: ps}, 0)
	rasImg.AddPath(&pathSourceAdapter{ps: ps}, 0)

	renBase.ResetClipping(true)
}
