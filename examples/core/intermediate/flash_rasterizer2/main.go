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
	"math/rand"
	"time"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/demo/shapesdata"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

type demo struct {
	shapes   []shapesdata.RawShape
	colors   []color.RGBA8[color.Linear]
	shapeIdx int
}

func newDemo() *demo {
	shapes := shapesdata.LoadShapes()
	rng := rand.New(rand.NewSource(42))
	colors := make([]color.RGBA8[color.Linear], 100)
	for i := range colors {
		colors[i] = color.RGBA8[color.Linear]{
			R: uint8(rng.Intn(256)),
			G: uint8(rng.Intn(256)),
			B: uint8(rng.Intn(256)),
			A: 230,
		}
		colors[i].Premultiply()
	}
	return &demo{shapes: shapes, colors: colors}
}

func (d *demo) Render(ctx *agg.Context) {
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

	img := ctx.GetImage()
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
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty, 1.0)
	}

	// Set up raw renderer pipeline.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, w*4)
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
	slRas := &rasScanlineAdapter{sl: sl}

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
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
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
	if strokeW < 0.5 {
		strokeW = 0.5
	}

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
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
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

	txt := fmt.Sprintf("Fill=%.2fms (%dFPS) Stroke=%.2fms (%dFPS) Total=%.2fms (%dFPS)",
		tfillMs, fillFPS, tstrokeMs, strokeFPS, ttotalMs, totalFPS)

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
		for ras.SweepScanline(slRas) {
			renscan.RenderScanlineAASolid(
				&scanlineWrapperU8{sl: sl},
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

func (d *demo) OnKey(key rune) bool {
	switch key {
	case 'q', 'Q':
		return false
	}
	return false
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool { return false }
func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool { return false }
func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool   { return false }

func main() {
	d := newDemo()
	demorunner.Run(demorunner.Config{
		Title:  "Flash Rasterizer 2 (Style Decomposition)",
		Width:  800,
		Height: 600,
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

// invertedFlatVS iterates FlatVertex slices with polygon winding inverted,
// matching C++ path_storage::invert_polygon exactly:
// commands shifted left by one, original MoveTo goes to last, coordinates reversed.
// Result: LineTo(pN), LineTo(pN-1), …, LineTo(p1), MoveTo(p0).
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

	// Shifted command: cmd[pos] = original_cmd[pos+1], last gets original_cmd[0].
	var cmd uint32
	if v.pos < n-1 {
		cmd = v.verts[v.pos+1].Cmd
	} else {
		cmd = v.verts[0].Cmd // MoveTo
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

// --- Scanline adapters (mirrors cmd/wasm/adapter.go) ---

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

type scanlineWrapperU8 struct{ sl *scanline.ScanlineU8 }

func (w *scanlineWrapperU8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *scanlineWrapperU8) Y() int               { return w.sl.Y() }
func (w *scanlineWrapperU8) NumSpans() int        { return w.sl.NumSpans() }

type spanIterU8 struct {
	spans []scanline.Span
	idx   int
}

func (it *spanIterU8) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *spanIterU8) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *scanlineWrapperU8) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &spanIterU8{spans: nil, idx: 0}
	}
	return &spanIterU8{spans: spans, idx: 0}
}
