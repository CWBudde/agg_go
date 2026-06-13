// Package main ports AGG's flash_rasterizer2.cpp demo.
//
// Alternative Flash compound-shape rasterization: decomposes a compound shape
// into separate sub-shapes per fill style. For each style index, paths whose
// left-fill matches are added forward; paths whose right-fill matches are added
// reversed (inverted polygon winding). A clipping rasterizer is used so the
// spurious edge from the clipper origin is safely discarded.
//
// Keys: left/right arrows to cycle through the 24 shape frames.
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
	"github.com/cwbudde/agg_go/internal/demo/timing"
	"github.com/cwbudde/agg_go/internal/gsv"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
)

type demo struct {
	shapes   []shapesdata.RawShape
	colors   []color.RGBA8[color.Linear]
	shapeIdx int
}

func newDemo() *demo {
	shapes := shapesdata.LoadShapes()
	return &demo{shapes: shapes, colors: cppDemoColors()}
}

func (d *demo) Render(img *agg.Image) {
	if len(d.shapes) == 0 {
		return
	}
	idx := d.shapeIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(d.shapes) {
		idx = len(d.shapes) - 1
	}
	shape := &d.shapes[idx]

	if len(shape.Paths) == 0 {
		return
	}

	w, h := img.Width(), img.Height()

	// Viewport: fit shape bounding rect into canvas (aspect-ratio preserving, centred).
	bx1, by1, bx2, by2 := shape.BoundingRect()
	worldW := bx2 - bx1
	worldH := by2 - by1
	if worldW <= 0 || worldH <= 0 {
		return
	}
	cW, cH := float64(w), float64(h)
	sc := cW / worldW
	if sy := cH / worldH; sy < sc {
		sc = sy
	}
	tx := (cW-worldW*sc)/2 - bx1*sc
	ty := (cH-worldH*sc)/2 - by1*sc

	// Pre-flatten all paths in screen coordinates.
	flatPaths := make([][]shapesdata.FlatVertex, len(shape.Paths))
	for i := range shape.Paths {
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty, sc)
	}

	// Set up raw renderer pipeline.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, img.Stride())
	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, w, h)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255})

	// Clipping rasterizer.
	clipper := rasterizer.NewRasterizerSlClip[float64, rasterizer.DblConv](rasterizer.DblConv{})
	ras := rasterizer.NewRasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]](
		rasterizer.DblConv{}, clipper,
	)
	ras.ClipBox(0, 0, float64(w), float64(h))
	ras.AutoClose(false)

	sl := scanline.NewScanlineU8()

	// Fill pass (flash2 method).
	tFillStart := time.Now()
	for s := shape.MinStyle; s <= shape.MaxStyle; s++ {
		ras.Reset()
		for i, p := range shape.Paths {
			if p.LeftFill == p.RightFill {
				continue
			}
			flat := flatPaths[i]
			if len(flat) == 0 {
				continue
			}
			if p.LeftFill == s {
				vs := &flatVertexSource{verts: flat}
				ras.AddPath(vs, 0)
			}
			if p.RightFill == s {
				vs := &invertedFlatVS{verts: flat}
				ras.AddPath(vs, 0)
			}
		}
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		c := d.styleColor(s)
		for ras.SweepScanline(sl) {
			renscan.RenderScanlineAASolid(
				sl,
				renBase,
				c,
			)
		}
	}

	tFill := time.Since(tFillStart)

	// Stroke pass (using conv_stroke with round joins/caps, matching C++).
	tStrokeStart := time.Now()
	ras.AutoClose(true)
	strokeColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128}
	strokeW := math.Sqrt(sc)

	flatSrc := &flatConvVS{}
	stroke := conv.NewConvStroke(flatSrc)
	stroke.SetWidth(strokeW)
	stroke.SetLineJoin(basics.RoundJoin)
	stroke.SetLineCap(basics.RoundCap)

	strokeRasVS := &convStrokeRasVS{stroke: stroke}

	for i, p := range shape.Paths {
		if p.Line < 0 {
			continue
		}
		flat := flatPaths[i]
		if len(flat) == 0 {
			continue
		}
		ras.Reset()
		flatSrc.verts = flat
		ras.AddPath(strokeRasVS, 0)
		if !ras.RewindScanlines() {
			continue
		}
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			renscan.RenderScanlineAASolid(
				sl,
				renBase,
				strokeColor,
			)
		}
	}

	tStroke := time.Since(tStrokeStart)
	tTotal := tFill + tStroke

	// Text overlay (timing info, matching C++ gsv_text output).
	ras.AutoClose(true)
	tfillMs := float64(tFill.Microseconds()) / 1000.0
	tstrokeMs := float64(tStroke.Microseconds()) / 1000.0
	ttotalMs := float64(tTotal.Microseconds()) / 1000.0
	fillFPS, strokeFPS, totalFPS := 0, 0, 0
	if tfillMs > 0 {
		fillFPS = int(1000.0 / tfillMs)
	}
	if tstrokeMs > 0 {
		strokeFPS = int(1000.0 / tstrokeMs)
	}
	if ttotalMs > 0 {
		totalFPS = int(1000.0 / ttotalMs)
	}

	txt := "Space: Next Shape\n\n+/- : ZoomIn/ZoomOut (with respect to the mouse pointer)"
	if timing.ShowText() {
		txt = fmt.Sprintf("Fill=%.2fms (%dFPS) Stroke=%.2fms (%dFPS) Total=%.2fms (%dFPS)\n\n%s",
			tfillMs, fillFPS, tstrokeMs, strokeFPS, ttotalMs, totalFPS, txt)
	}

	gsvT := gsv.NewGSVText()
	gsvT.SetSize(8.0, 0)
	gsvT.SetFlip(true)
	gsvT.SetStartPoint(10.0, 20.0)
	gsvT.SetText(txt)

	gsvTS := gsv.NewGSVTextOutline(gsvT)
	gsvTS.SetWidth(1.6)

	textRasVS := &convVertexSourceRasVS{src: gsvTS}
	ras.Reset()
	ras.AddPath(textRasVS, 0)
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		textColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
		for ras.SweepScanline(sl) {
			renscan.RenderScanlineAASolid(
				sl,
				renBase,
				textColor,
			)
		}
	}

	_ = agg.RGBA(0, 0, 0, 0) // keep agg import live
}

func (d *demo) styleColor(s int) color.RGBA8[color.Linear] {
	if s < 0 || s >= len(d.colors) {
		return color.RGBA8[color.Linear]{R: 200, G: 200, B: 200, A: 200}
	}
	return d.colors[s]
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

func (d *demo) OnKey(key rune) bool {
	switch key {
	case 'q', 'Q':
		return false
	}
	return false
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool { return false }
func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool { return false }
func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool   { return false }

func main() {
	d := newDemo()
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:                 "Flash Rasterizer 2 (Style Decomposition)",
		Width:                 655,
		Height:                520,
		EncodeLinearRGBToSRGB: true,
	}, d)
}

// --- Vertex sources ---

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

// invertedFlatVS iterates FlatVertex slices with polygon winding inverted.
// Mirrors C++ path_storage::invert_polygon: shift commands left, then
// reverse all vertices (coordinates AND commands together).
// Result: MoveTo(pN), LineTo(pN-1), …, LineTo(p1), LineTo(p0).
type invertedFlatVS struct {
	verts []shapesdata.FlatVertex
	pos   int
}

func (v *invertedFlatVS) Rewind(_ uint32) { v.pos = 0 }
func (v *invertedFlatVS) Vertex(x, y *float64) uint32 {
	n := len(v.verts)
	if v.pos >= n {
		return uint32(basics.PathCmdStop)
	}
	fv := v.verts[n-1-v.pos]
	*x, *y = fv.X, fv.Y

	// After shift-left then full reversal:
	//   pos 0   → original cmd[0] (MoveTo)
	//   pos i>0 → original cmd[n-i]
	var cmd uint32
	if v.pos == 0 {
		cmd = v.verts[0].Cmd
	} else {
		cmd = v.verts[n-v.pos].Cmd
	}
	v.pos++
	return cmd
}

// --- conv.VertexSource adapter for flat vertices (feeds into ConvStroke) ---

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

// convStrokeRasVS adapts conv.ConvStroke to the rasterizer's VertexSource interface.
type convStrokeRasVS struct {
	stroke *conv.ConvStroke
}

func (a *convStrokeRasVS) Rewind(pathID uint32) { a.stroke.Rewind(uint(pathID)) }
func (a *convStrokeRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.stroke.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}

// convVertexSourceRasVS adapts any conv.VertexSource to the rasterizer's VertexSource interface.
type convVertexSourceRasVS struct {
	src conv.VertexSource
}

func (a *convVertexSourceRasVS) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *convVertexSourceRasVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x, *y = vx, vy
	return uint32(cmd)
}
