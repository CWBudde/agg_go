// Port of AGG C++ rounded_rect.cpp – interactive rounded rectangle with controls.
//
// Renders a rounded rectangle defined by two draggable corner points, with
// adjustable radius and subpixel offset. Matches the C++ original's rendering
// pipeline: renderer_base + rasterizer_scanline_aa + scanline_p8 + conv_stroke.
// Default: corners at (100,100)-(500,350), radius=25, offset=0.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

const (
	demoWidth  = 600
	demoHeight = 400

	defaultRadius = 25.0
	defaultOffset = 0.0
)

// rrVertexSource adapts shapes.RoundedRect to conv.VertexSource.
type rrVertexSource struct {
	rr *shapes.RoundedRect
}

func (s *rrVertexSource) Rewind(pathID uint) { s.rr.Rewind(uint32(pathID)) }

func (s *rrVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.rr.Vertex(&x, &y)
	return
}

// strokeVertexSource adapts conv.ConvStroke to the rasterizer VertexSource.
type strokeVertexSource struct {
	cs *conv.ConvStroke
}

func (s *strokeVertexSource) Rewind(pathID uint32) { s.cs.Rewind(uint(pathID)) }

func (s *strokeVertexSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := s.cs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// ellipseVertexSource adapts shapes.Ellipse to the rasterizer VertexSource.
type ellipseVertexSource struct {
	e *shapes.Ellipse
}

func (s *ellipseVertexSource) Rewind(pathID uint32) { s.e.Rewind(pathID) }

func (s *ellipseVertexSource) Vertex(x, y *float64) uint32 {
	cmd := s.e.Vertex(x, y)
	return uint32(cmd)
}

// Scanline/rasterizer adapters to bridge rasterizer ↔ renscan interfaces.
type demo struct {
	x      [2]float64
	y      [2]float64
	dx, dy float64
	idx    int

	radius float64
	offset float64
}

func newDemo() *demo {
	return &demo{
		x:      [2]float64{100, 500},
		y:      [2]float64{100, 350},
		idx:    -1,
		radius: defaultRadius,
		offset: defaultOffset,
	}
}

func (d *demo) Render(img *agg.Image) {
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)
	pixFmt := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	rb := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[color.Linear], color.RGBA8[color.Linear]](pixFmt)
	rb.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineP8()

	// Render two "control" circles.
	gray := color.RGBA8[color.Linear]{R: 127, G: 127, B: 127, A: 255}
	for i := 0; i < 2; i++ {
		e := shapes.NewEllipseWithParams(d.x[i], d.y[i], 3, 3, 16, false)
		ras.Reset()
		ras.AddPath(&ellipseVertexSource{e: e}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, rb, gray)
	}

	// Create rounded rectangle.
	off := d.offset
	rr := shapes.NewRoundedRect(d.x[0]+off, d.y[0]+off, d.x[1]+off, d.y[1]+off, d.radius)
	rr.NormalizeRadius()

	// Draw as outline.
	stroke := conv.NewConvStroke(&rrVertexSource{rr: rr})
	stroke.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&strokeVertexSource{cs: stroke}, 0)
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, rb, black)
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if !btn.Left {
		return false
	}
	fx, fy := float64(x), float64(y)
	for i := 0; i < 2; i++ {
		if math.Sqrt((fx-d.x[i])*(fx-d.x[i])+(fy-d.y[i])*(fy-d.y[i])) < 5.0 {
			d.dx = fx - d.x[i]
			d.dy = fy - d.y[i]
			d.idx = i
			return true
		}
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left && d.idx >= 0 {
		d.x[d.idx] = float64(x) - d.dx
		d.y[d.idx] = float64(y) - d.dy
		return true
	}
	if !btn.Left {
		d.idx = -1
	}
	return false
}

func (d *demo) OnMouseUp(_, _ int, _ lowlevelrunner.Buttons) bool {
	d.idx = -1
	return false
}

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Rounded Rectangle",
		Width:  demoWidth,
		Height: demoHeight,
	}, newDemo())
}
