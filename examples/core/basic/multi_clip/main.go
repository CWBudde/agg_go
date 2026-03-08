// Port of AGG C++ multi_clip.cpp – multi-clip region rendering.
//
// Renders the lion through a grid of N×N inset clip rectangles using
// RendererMClip. Default: N=3 (3×3 grid of clip boxes).
package main

import (
	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	liondemo "agg_go/internal/demo/lion"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
)

const clipN = 3

// Scanline/rasterizer adapters.

type mcRasAdp struct {
	ras interface {
		RewindScanlines() bool
		SweepScanline(sl rasterizer.ScanlineInterface) bool
		MinX() int
		MaxX() int
	}
}

func (r *mcRasAdp) RewindScanlines() bool { return r.ras.RewindScanlines() }
func (r *mcRasAdp) MinX() int             { return r.ras.MinX() }
func (r *mcRasAdp) MaxX() int             { return r.ras.MaxX() }

type mcSlAdpP8 struct{ sl *scanline.ScanlineP8 }

func (a *mcSlAdpP8) ResetSpans()                 { a.sl.ResetSpans() }
func (a *mcSlAdpP8) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *mcSlAdpP8) AddSpan(x, l int, c uint32)  { a.sl.AddSpan(x, l, uint(c)) }
func (a *mcSlAdpP8) Finalize(y int)              { a.sl.Finalize(y) }
func (a *mcSlAdpP8) NumSpans() int               { return a.sl.NumSpans() }

func (r *mcRasAdp) SweepScanline(sl renscan.ScanlineInterface) bool {
	if w, ok := sl.(*mcSlWrapP8); ok {
		return r.ras.SweepScanline(&mcSlAdpP8{sl: w.sl})
	}
	return false
}

type mcSlWrapP8 struct{ sl *scanline.ScanlineP8 }

func (w *mcSlWrapP8) Reset(minX, maxX int) { w.sl.Reset(minX, maxX) }
func (w *mcSlWrapP8) Y() int               { return w.sl.Y() }
func (w *mcSlWrapP8) NumSpans() int        { return w.sl.NumSpans() }

type mcSpanIter struct {
	spans []scanline.SpanP8
	idx   int
}

func (it *mcSpanIter) GetSpan() renscan.SpanData {
	s := it.spans[it.idx]
	return renscan.SpanData{X: int(s.X), Len: int(s.Len), Covers: s.Covers}
}
func (it *mcSpanIter) Next() bool { it.idx++; return it.idx < len(it.spans) }

func (w *mcSlWrapP8) Begin() renscan.ScanlineIterator {
	spans := w.sl.Spans()
	if len(spans) == 0 {
		return &mcSpanIter{nil, 0}
	}
	return &mcSpanIter{spans, 0}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	width := ctx.Width()
	height := ctx.Height()
	img := ctx.GetImage()
	agg2d := ctx.GetAgg2D()

	// White background.
	agg2d.ClearAll(agg.White)

	// Setup lion transform: centred, facing right.
	agg2d.ResetTransformations()
	agg2d.Translate(-250, -250)
	agg2d.Rotate(basics.Pi)
	agg2d.Translate(float64(width)/2, float64(height)/2)
	mtx := agg2d.GetTransformations()

	// Setup multi-clip renderer.
	mainBuf := buffer.NewRenderingBufferWithData[uint8](img.Data, width, height, width*4)
	mainPixf := pixfmt.NewPixFmtRGBA32[color.Linear](mainBuf)
	mclip := renderer.NewRendererMClip(mainPixf)

	mclip.ResetClipping(false) // start with no visible regions
	n := clipN
	for xi := 0; xi < n; xi++ {
		for yi := 0; yi < n; yi++ {
			x1 := int(float64(width) * float64(xi) / float64(n))
			y1 := int(float64(height) * float64(yi) / float64(n))
			x2 := int(float64(width) * float64(xi+1) / float64(n))
			y2 := int(float64(height) * float64(yi+1) / float64(n))
			mclip.AddClipBox(x1+5, y1+5, x2-5, y2-5)
		}
	}

	ras := agg2d.GetInternalRasterizer()
	rasAdp := &mcRasAdp{ras: ras}
	sl := scanline.NewScanlineP8()
	slAdp := &mcSlWrapP8{sl: sl}

	lionPaths := liondemo.Parse()
	for _, lp := range lionPaths {
		c := color.RGBA8[color.Linear]{R: lp.Color[0], G: lp.Color[1], B: lp.Color[2], A: 255}
		ras.Reset()
		lp.Path.Rewind(0)
		for {
			x, y, cmd := lp.Path.NextVertex()
			if basics.IsStop(basics.PathCommand(cmd)) {
				break
			}
			tx, ty := mtx.Transform(x, y)
			if basics.IsMoveTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdMoveTo))
			} else if basics.IsLineTo(basics.PathCommand(cmd)) {
				ras.AddVertex(tx, ty, uint32(basics.PathCmdLineTo))
			}
		}
		renscan.RenderScanlinesAASolid(rasAdp, slAdp, mclip, c)
	}
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Multi Clip", Width: 800, Height: 600}, &demo{})
}
