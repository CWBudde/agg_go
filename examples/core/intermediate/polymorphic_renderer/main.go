// Package main ports AGG's polymorphic_renderer.cpp demo.
//
// In the original C++ demo, a virtual base class (polymorphic_renderer_solid_rgba8_base)
// and a template adaptor let one rendering routine work with any pixel-format backend.
// Go interfaces provide this naturally: the same draw call works through any
// implementation of the SolidFiller interface below, without virtual keyword,
// explicit factory, or heap allocation of C++ base classes.
//
// Visual: a filled triangle on a white background.
// Drag the three vertex handles to reshape it.
package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

// SolidFiller is the Go equivalent of C++'s polymorphic_renderer_solid_rgba8_base.
// Any pixel-format backend that implements these methods is a valid renderer.
type SolidFiller interface {
	Clear(c color.RGBA8[color.Linear])
	SetColor(c color.RGBA8[color.Linear])
	RenderTriangle(x, y [3]float64)
}

// rgba32Filler is one concrete implementation backed by PixFmtRGBA32.
type rgba32Filler struct {
	renBase   *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
	fillColor color.RGBA8[color.Linear]
}

func newRGBA32Filler(rbuf *buffer.RenderingBufferU8, w, h int) *rgba32Filler {
	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	rb.ClipBox(0, 0, w, h)
	return &rgba32Filler{renBase: rb}
}

func (r *rgba32Filler) Clear(c color.RGBA8[color.Linear])    { r.renBase.Clear(c) }
func (r *rgba32Filler) SetColor(c color.RGBA8[color.Linear]) { r.fillColor = c }

// RenderTriangle is the rendering routine — it operates identically regardless
// of which concrete SolidFiller implementation is in use.
func (r *rgba32Filler) RenderTriangle(x, y [3]float64) {
	ps := path.NewPathStorageStl()
	ps.MoveTo(x[0], y[0])
	ps.LineTo(x[1], y[1])
	ps.LineTo(x[2], y[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&psAdapter{ps: ps}, 0)

	sl := scanline.NewScanlineU8()
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&slAdapter{sl: sl}) {
			y := sl.Y()
			for _, sp := range sl.Spans() {
				if sp.Len > 0 {
					r.renBase.BlendSolidHspan(int(sp.X), y, int(sp.Len), r.fillColor, sp.Covers)
				}
			}
		}
	}
}

// drawFilled demonstrates the polymorphic dispatch: the same code works with
// any SolidFiller, just as the C++ version worked with any PixFmt.
func drawFilled(ren SolidFiller, x, y [3]float64, bg, fg color.RGBA8[color.Linear]) {
	ren.Clear(bg)
	ren.SetColor(fg)
	ren.RenderTriangle(x, y)
}

// --- Demo ---

type demo struct {
	x, y     [3]float64
	selected int
	dragDX   float64
	dragDY   float64
}

func newDemo() *demo {
	return &demo{
		x:        [3]float64{100, 369, 143},
		y:        [3]float64{60, 170, 310},
		selected: -1,
	}
}

func (d *demo) Render(ctx *agg.Context) {
	img := ctx.GetImage()
	w, h := img.Width(), img.Height()

	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, w*4)

	// Swap out rgba32Filler for any other SolidFiller without changing drawFilled.
	var ren SolidFiller = newRGBA32Filler(rbuf, w, h)
	drawFilled(
		ren,
		d.x, d.y,
		color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}, // white bg
		color.RGBA8[color.Linear]{R: 80, G: 30, B: 20, A: 255},    // dark red fill
	)
}

func (d *demo) OnMouseDown(x, y int, btn demorunner.Buttons) bool {
	d.selected = -1
	for i := 0; i < 3; i++ {
		dx := float64(x) - d.x[i]
		dy := float64(y) - d.y[i]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			d.selected = i
			d.dragDX = dx
			d.dragDY = dy
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn demorunner.Buttons) bool {
	if d.selected < 0 {
		return false
	}
	d.x[d.selected] = float64(x) - d.dragDX
	d.y[d.selected] = float64(y) - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn demorunner.Buttons) bool {
	d.selected = -1
	return false
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Polymorphic Renderer",
		Width:  400,
		Height: 330,
	}, newDemo())
}

// --- Minimal adapters ---

type psAdapter struct{ ps *path.PathStorageStl }

func (a *psAdapter) Rewind(id uint32) { a.ps.Rewind(uint(id)) }
func (a *psAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ps.NextVertex()
	*x, *y = vx, vy
	return cmd
}

type slAdapter struct{ sl *scanline.ScanlineU8 }

func (a *slAdapter) ResetSpans()                 { a.sl.ResetSpans() }
func (a *slAdapter) AddCell(x int, cover uint32) { a.sl.AddCell(x, uint(cover)) }
func (a *slAdapter) AddSpan(x, length int, cover uint32) {
	a.sl.AddSpan(x, length, uint(cover))
}
func (a *slAdapter) Finalize(y int) { a.sl.Finalize(y) }
func (a *slAdapter) NumSpans() int  { return a.sl.NumSpans() }
