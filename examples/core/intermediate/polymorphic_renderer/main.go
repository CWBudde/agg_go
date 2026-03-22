// Package main ports AGG's polymorphic_renderer.cpp demo.
//
// In the original C++ demo, a virtual base class (polymorphic_renderer_solid_rgba8_base)
// and a template adaptor let one rendering routine work with any pixel-format backend.
// Go interfaces provide this naturally: the same draw call works through any
// implementation of the SolidRenderer interface below, without virtual keyword,
// explicit factory, or heap allocation of C++ base classes.
//
// Visual: a filled triangle on a white background.
// Drag the three vertex handles to reshape it.
//
// NOTE: The C++ original uses pix_format_rgb555 (15-bit packed pixels).
// This port uses RGBA32 until PixFmtRGB555 is completed (PLAN.md 10.10).
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	rendsl "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

// SolidRenderer is the Go equivalent of C++'s polymorphic_renderer_solid_rgba8_base.
// Any pixel-format backend that implements these methods is a valid renderer.
type SolidRenderer interface {
	Clear(c color.RGBA8[color.SRGB])
	SetColor(c color.RGBA8[color.SRGB])
	Prepare()
	Render(sl rendsl.ScanlineInterface)
}

// rgba32Renderer is one concrete implementation backed by PixFmtRGBA32 (sRGB).
// This mirrors C++'s polymorphic_renderer_solid_rgba8_adaptor<PixFmt>.
type rgba32Renderer struct {
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]]
	ren     *rendsl.RendererScanlineAASolid[*renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]]
}

func newRGBA32Renderer(rbuf *buffer.RenderingBufferU8) *rgba32Renderer {
	pf := pixfmt.NewPixFmtRGBA32[color.SRGB](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.SRGB]], color.RGBA8[color.SRGB]](pf)
	ren := rendsl.NewRendererScanlineAASolidWithRenderer(rb)
	return &rgba32Renderer{renBase: rb, ren: ren}
}

func (r *rgba32Renderer) Clear(c color.RGBA8[color.SRGB]) { r.renBase.Clear(c) }
func (r *rgba32Renderer) SetColor(c color.RGBA8[color.SRGB]) {
	r.ren.SetColor(c)
}
func (r *rgba32Renderer) Prepare()                          { r.ren.Prepare() }
func (r *rgba32Renderer) Render(sl rendsl.ScanlineInterface) { r.ren.Render(sl) }

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

func (d *demo) Render(img *agg.Image) {
	w, h := img.Width(), img.Height()

	// Negative stride gives flip_y=true (Y=0 at bottom), matching C++.
	// The runner's FlipY config handles flipping the output and mouse coords.
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, -w*4)

	var ren SolidRenderer = newRGBA32Renderer(rbuf)

	// Build the triangle path.
	ps := path.NewPathStorageStl()
	ps.MoveTo(d.x[0], d.y[0])
	ps.LineTo(d.x[1], d.y[1])
	ps.LineTo(d.x[2], d.y[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&psAdapter{ps: ps}, 0)

	sl := scanline.NewScanlineP8()

	// Polymorphic dispatch: same code works with any SolidRenderer,
	// just as the C++ version works with any PixFmt.
	ren.Clear(color.RGBA8[color.SRGB]{R: 255, G: 255, B: 255, A: 255})
	ren.SetColor(color.RGBA8[color.SRGB]{R: 80, G: 30, B: 20, A: 255})
	rendsl.RenderScanlines[color.RGBA8[color.SRGB]](ras, sl, ren)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	d.selected = -1
	for i := range 3 {
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

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if d.selected < 0 {
		return false
	}
	d.x[d.selected] = float64(x) - d.dragDX
	d.y[d.selected] = float64(y) - d.dragDY
	return true
}

func (d *demo) OnMouseUp(x, y int, btn lowlevelrunner.Buttons) bool {
	d.selected = -1
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Polymorphic Renderer",
		Width:  400,
		Height: 330,
		FlipY:  true,
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
