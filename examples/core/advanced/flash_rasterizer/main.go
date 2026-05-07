// Port of AGG C++ flash_rasterizer.cpp.
package main

import (
	"fmt"
	"math"
	"time"

	agg "github.com/cwbudde/agg_go"
	"github.com/cwbudde/agg_go/examples/shared/lowlevelrunner"
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/demo/shapesdata"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

const (
	width  = 655
	height = 520
)

type flashStyleHandler struct {
	colors []color.RGBA8[color.Linear]
}

func (h *flashStyleHandler) IsSolid(style int) bool { return true }
func (h *flashStyleHandler) Color(style int) color.RGBA8[color.Linear] {
	if style < 0 || style >= len(h.colors) {
		return color.RGBA8[color.Linear]{}
	}
	return h.colors[style]
}

func (h *flashStyleHandler) GenerateSpan(colors []color.RGBA8[color.Linear], x, y, l, style int) {
}

type flashSLAdapter struct{ sl *scanline.ScanlineU8 }

func (a *flashSLAdapter) ResetSpans()                      { a.sl.ResetSpans() }
func (a *flashSLAdapter) AddCell(x int, c basics.Int8u)    { a.sl.AddCell(x, uint(c)) }
func (a *flashSLAdapter) AddSpan(x, l int, c basics.Int8u) { a.sl.AddSpan(x, l, uint(c)) }
func (a *flashSLAdapter) Finalize(y int)                   { a.sl.Finalize(y) }
func (a *flashSLAdapter) NumSpans() int                    { return a.sl.NumSpans() }

type compoundClip struct {
	clip *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]
}

func newCompoundClip() *compoundClip {
	return &compoundClip{
		clip: rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{}),
	}
}

func (c *compoundClip) ResetClipping() { c.clip.ResetClipping() }
func (c *compoundClip) ClipBox(x1, y1, x2, y2 float64) {
	c.clip.ClipBox(x1, y1, x2, y2)
}
func (c *compoundClip) MoveTo(x, y float64) { c.clip.MoveTo(x, y) }
func (c *compoundClip) LineTo(outline *rasterizer.RasterizerCellsAAStyled, x, y float64) {
	c.clip.LineTo(outline, x, y)
}

type demo struct {
	shapes []shapesdata.RawShape
	colors []color.RGBA8[color.Linear]
}

func newDemo() *demo {
	return &demo{
		shapes: shapesdata.LoadShapes(),
		colors: cppDemoColors(),
	}
}

func (d *demo) Render(img *agg.Image) {
	if len(d.shapes) == 0 {
		return
	}
	shape := &d.shapes[0]
	if len(shape.Paths) == 0 {
		return
	}

	w, h := img.Width(), img.Height()
	bx1, by1, bx2, by2 := shape.BoundingRect()
	worldW, worldH := bx2-bx1, by2-by1
	if worldW <= 0 || worldH <= 0 {
		return
	}

	sc := float64(w) / worldW
	if sy := float64(h) / worldH; sy < sc {
		sc = sy
	}
	tx := (float64(w)-worldW*sc)/2 - bx1*sc
	ty := (float64(h)-worldH*sc)/2 - by1*sc

	flatPaths := make([][]shapesdata.FlatVertex, len(shape.Paths))
	for i := range shape.Paths {
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty, sc)
	}

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, w, h)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	clipper := newCompoundClip()
	rasc := rasterizer.NewRasterizerCompoundAA(clipper)
	rasc.ClipBox(0, 0, float64(w), float64(h))

	tFillStart := time.Now()
	for i, p := range shape.Paths {
		if p.LeftFill < 0 && p.RightFill < 0 {
			continue
		}
		rasc.Styles(p.LeftFill, p.RightFill)
		vs := &flatVertexSource{verts: flatPaths[i]}
		rasc.AddPath(vs, 0)
	}

	rasc.Sort()
	renderCompound(rasc, renBase, &flashStyleHandler{colors: d.colors})
	tFill := time.Since(tFillStart)

	tStrokeStart := time.Now()
	renderStrokes(shape, flatPaths, sc, renBase, w, h)
	tStroke := time.Since(tStrokeStart)

	renderInfoText(renBase, w, h, tFill, tStroke)
}

func renderCompound(
	rasc *rasterizer.RasterizerCompoundAA[*compoundClip],
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	styleHandler *flashStyleHandler,
) {
	if !rasc.RewindScanlines() {
		return
	}

	minX := rasc.MinX()
	maxX := rasc.MaxX()

	slAA := scanline.NewScanlineU8()
	slBin := scanline.NewScanlineU8()
	slAA.Reset(minX, maxX)
	slBin.Reset(minX, maxX)
	adAA := &flashSLAdapter{sl: slAA}
	adBin := &flashSLAdapter{sl: slBin}

	length := maxX - minX + 2
	if length < 0 {
		length = 0
	}
	colorSpan := make([]color.RGBA8[color.Linear], length*2)
	mixBuffer := colorSpan[length:]

	for {
		numStyles := rasc.SweepStyles()
		if numStyles == 0 {
			break
		}
		if numStyles == 1 {
			if rasc.SweepScanline(adAA, 0) {
				style := int(rasc.Style(0))
				c := styleHandler.Color(style)
				y := slAA.Y()
				for _, sp := range slAA.Spans() {
					if sp.Len > 0 {
						renBase.BlendSolidHspan(int(sp.X), y, int(sp.Len), c, sp.Covers)
					}
				}
			}
		} else {
			if rasc.SweepScanline(adBin, -1) {
				y := slBin.Y()
				for _, sp := range slBin.Spans() {
					for j := 0; j < int(sp.Len); j++ {
						mixBuffer[int(sp.X)-minX+j] = color.RGBA8[color.Linear]{}
					}
				}
				for i := uint32(0); i < numStyles; i++ {
					style := int(rasc.Style(i))
					if rasc.SweepScanline(adAA, int(i)) {
						for _, sp := range slAA.Spans() {
							c := styleHandler.Color(style)
							for j := 0; j < int(sp.Len); j++ {
								ptr := &mixBuffer[int(sp.X)-minX+j]
								cover := sp.Covers[j]
								ptr.AddWithCover(c, cover)
							}
						}
					}
				}
				for _, sp := range slBin.Spans() {
					renBase.BlendColorHspan(int(sp.X), y, int(sp.Len), mixBuffer[int(sp.X)-minX:], nil, basics.CoverFull)
				}
			}
		}
	}
}

func renderStrokes(
	shape *shapesdata.RawShape,
	flatPaths [][]shapesdata.FlatVertex,
	scale float64,
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	w, h int,
) {
	clipper := rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{})
	ras := rasterizer.NewRasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]](
		rasterizer.DblConv{}, clipper,
	)
	ras.ClipBox(0, 0, float64(w), float64(h))
	ras.AutoClose(true)

	sl := scanline.NewScanlineU8()
	strokeColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128}
	strokeW := math.Sqrt(scale)

	flatSrc := &flatConvVS{}
	stroke := conv.NewConvStroke(flatSrc)
	stroke.SetWidth(strokeW)
	stroke.SetLineJoin(basics.RoundJoin)
	stroke.SetLineCap(basics.RoundCap)

	strokeRasVS := &convStrokeRasVS{stroke: stroke}
	for i, p := range shape.Paths {
		if p.Line < 0 || len(flatPaths[i]) == 0 {
			continue
		}
		ras.Reset()
		flatSrc.verts = flatPaths[i]
		ras.AddPath(strokeRasVS, 0)
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			renscan.RenderScanlineAASolid(sl, renBase, strokeColor)
		}
	}
}

func renderInfoText(
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	w, h int,
	tFill, tStroke time.Duration,
) {
	tfillMs := float64(tFill.Microseconds()) / 1000.0
	tstrokeMs := float64(tStroke.Microseconds()) / 1000.0
	ttotalMs := tfillMs + tstrokeMs
	fps := func(ms float64) int {
		if ms <= 0 {
			return 0
		}
		return int(1000.0 / ms)
	}

	txt := fmt.Sprintf("Fill=%.2fms (%dFPS) Stroke=%.2fms (%dFPS) Total=%.2fms (%dFPS)\n\nSpace: Next Shape\n\n+/- : ZoomIn/ZoomOut (with respect to the mouse pointer)",
		tfillMs, fps(tfillMs),
		tstrokeMs, fps(tstrokeMs),
		ttotalMs, fps(ttotalMs),
	)

	clipper := rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{})
	ras := rasterizer.NewRasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]](
		rasterizer.DblConv{}, clipper,
	)
	ras.ClipBox(0, 0, float64(w), float64(h))

	t := gsv.NewGSVText()
	t.SetSize(8.0, 0)
	t.SetFlip(true)
	t.SetStartPoint(10.0, 20.0)
	t.SetText(txt)

	ts := gsv.NewGSVTextOutline(t)
	ts.SetWidth(1.6)

	ras.AddPath(&convVertexSourceRasVS{src: ts}, 0)
	if !ras.RewindScanlines() {
		return
	}

	sl := scanline.NewScanlineU8()
	sl.Reset(ras.MinX(), ras.MaxX())
	for ras.SweepScanline(sl) {
		renscan.RenderScanlineAASolid(sl, renBase, color.RGBA8[color.Linear]{A: 255})
	}
}

func cppDemoColors() []color.RGBA8[color.Linear] {
	rng := newGlibcRand(1)
	colors := make([]color.RGBA8[color.Linear], 100)
	for i := range colors {
		// GCC evaluates the original srgba8(rand(), rand(), rand(), 230)
		// arguments right-to-left for this demo build.
		b := basics.Int8u(rng.next() & 0xFF)
		g := basics.Int8u(rng.next() & 0xFF)
		r := basics.Int8u(rng.next() & 0xFF)
		c := color.ConvertToLinear(color.RGBA8[color.SRGB]{
			R: r,
			G: g,
			B: b,
			A: 230,
		})
		c.Premultiply()
		colors[i] = c
	}
	return colors
}

type glibcRand struct {
	state []uint32
	idx   int
}

func newGlibcRand(seed uint32) *glibcRand {
	if seed == 0 {
		seed = 1
	}
	state := make([]uint32, 344)
	state[0] = seed
	for i := 1; i < 31; i++ {
		state[i] = uint32((16807 * uint64(state[i-1])) % 2147483647)
	}
	for i := 31; i < 34; i++ {
		state[i] = state[i-31]
	}
	for i := 34; i < len(state); i++ {
		state[i] = state[i-31] + state[i-3]
	}
	return &glibcRand{state: state, idx: 344}
}

func (r *glibcRand) next() uint32 {
	r.state = append(r.state, r.state[r.idx-31]+r.state[r.idx-3])
	v := r.state[r.idx] >> 1
	r.idx++
	return v
}

type flatVertexSource struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *flatVertexSource) Rewind(_ uint32) { v.pos = 0 }
func (v *flatVertexSource) Vertex(x, y *float64) uint32 {
	if v.pos >= len(v.verts) {
		return uint32(basics.PathCmdStop)
	}
	fv := v.verts[v.pos]
	v.pos++
	*x, *y = fv.X, fv.Y
	return fv.Cmd
}

type flatConvVS struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *flatConvVS) Rewind(_ uint) { v.pos = 0 }
func (v *flatConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	if v.pos >= len(v.verts) {
		return 0, 0, basics.PathCmdStop
	}
	fv := v.verts[v.pos]
	v.pos++
	return fv.X, fv.Y, basics.PathCommand(fv.Cmd)
}

type convStrokeRasVS struct {
	stroke *conv.ConvStroke
}

func (a *convStrokeRasVS) Rewind(pathID uint32) { a.stroke.Rewind(uint(pathID)) }
func (a *convStrokeRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.stroke.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

type convVertexSourceRasVS struct {
	src conv.VertexSource
}

func (a *convVertexSourceRasVS) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convVertexSourceRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

func main() {
	d := newDemo()
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Flash Rasterizer",
		Width:                 width,
		Height:                height,
		EncodeLinearRGBToSRGB: true,
	}, d)
}
