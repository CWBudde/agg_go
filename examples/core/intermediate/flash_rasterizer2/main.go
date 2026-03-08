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
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/demo/shapesdata"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
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
		flatPaths[i] = shapesdata.FlattenPath(&shape.Paths[i], sc, sc, tx, ty)
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

	// Stroke pass.
	ras.AutoClose(true)
	strokeColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 128}
	strokeW := math.Sqrt(sc)
	if strokeW < 0.5 {
		strokeW = 0.5
	}

	for i, p := range shape.Paths {
		if p.Line < 0 {
			continue
		}
		flat := flatPaths[i]
		if len(flat) == 0 {
			continue
		}
		ras.Reset()
		strokeFlatPath(ras, flat, strokeW)
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
	i := n - 1 - v.pos
	fv := v.verts[i]
	*x, *y = fv.X, fv.Y
	var cmd uint32
	if v.pos == n-1 {
		cmd = shapesdata.PathCmdMoveTo
	} else {
		cmd = shapesdata.PathCmdLineTo
	}
	v.pos++
	return cmd
}

// --- Stroke helper ---

func strokeFlatPath(
	ras *rasterizer.RasterizerScanlineAA[float64, rasterizer.DblConv, *rasterizer.RasterizerSlClip[float64, rasterizer.DblConv]],
	flat []shapesdata.FlatVertex,
	w float64,
) {
	if len(flat) < 2 {
		return
	}
	hw := w * 0.5
	for i := 1; i < len(flat); i++ {
		if flat[i].Cmd != shapesdata.PathCmdLineTo {
			continue
		}
		x1, y1 := flat[i-1].X, flat[i-1].Y
		x2, y2 := flat[i].X, flat[i].Y
		dx, dy := x2-x1, y2-y1
		d := math.Sqrt(dx*dx + dy*dy)
		if d < 1e-6 {
			continue
		}
		nx, ny := -dy/d*hw, dx/d*hw
		ras.MoveToD(x1+nx, y1+ny)
		ras.LineToD(x2+nx, y2+ny)
		ras.LineToD(x2-nx, y2-ny)
		ras.LineToD(x1-nx, y1-ny)
	}
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
