// Port of AGG C++ rounded_rect.cpp – interactive rounded rectangle.
//
// Two draggable control points define the opposite corners of a rectangle.
// Sliders control the corner radius and a sub-pixel offset.
// A checkbox toggles white-on-black rendering.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

// --- State ---

const (
	rrNativeWidth  = 600
	rrNativeHeight = 400
)

var (
	rrPts = [2][2]float64{
		{100, 100},
		{500, 350},
	}
	rrRadius   = 25.0
	rrOffset   = 0.0
	rrDarkBg   = false
	rrSelected = -1
	rrDragDX   = 0.0
	rrDragDY   = 0.0
)

// rrCanvasOffset centres the native 600x400 scene in the 800x600 canvas.
func rrCanvasOffset() (ox, oy float64) {
	return float64(width-rrNativeWidth) / 2, float64(height-rrNativeHeight) / 2
}

// --- Vertex source adapters ---

// rrRoundedRectVS adapts shapes.RoundedRect to conv.VertexSource.
type rrRoundedRectVS struct {
	rr *shapes.RoundedRect
}

func (s *rrRoundedRectVS) Rewind(pathID uint) { s.rr.Rewind(uint32(pathID)) }

func (s *rrRoundedRectVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = s.rr.Vertex(&x, &y)
	return
}

// rrStrokeVS adapts conv.ConvStroke to the rasterizer VertexSource interface.
type rrStrokeVS struct {
	cs *conv.ConvStroke
}

func (s *rrStrokeVS) Rewind(pathID uint32) { s.cs.Rewind(uint(pathID)) }

func (s *rrStrokeVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := s.cs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// rrEllipseVS adapts shapes.Ellipse to the rasterizer VertexSource interface.
type rrEllipseVS struct {
	e *shapes.Ellipse
}

func (s *rrEllipseVS) Rewind(pathID uint32) { s.e.Rewind(pathID) }

func (s *rrEllipseVS) Vertex(x, y *float64) uint32 {
	return uint32(s.e.Vertex(x, y))
}

// --- Drawing ---

func drawRoundedRectDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)

	// Background.
	if rrDarkBg {
		ren.Clear(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
	} else {
		ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})
	}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	ox, oy := rrCanvasOffset()

	// Small circles at the two control points.
	gray := color.RGBA8[color.Linear]{R: 127, G: 127, B: 127, A: 255}
	for _, pt := range rrPts {
		e := shapes.NewEllipseWithParams(pt[0]+ox, pt[1]+oy, 5, 5, 32, false)
		ras.Reset()
		ras.AddPath(&rrEllipseVS{e: e}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, ren, gray)
	}

	// Rounded rectangle with optional sub-pixel offset.
	off := rrOffset
	x1 := rrPts[0][0] + off + ox
	y1 := rrPts[0][1] + off + oy
	x2 := rrPts[1][0] + off + ox
	y2 := rrPts[1][1] + off + oy

	rr := shapes.NewRoundedRect(x1, y1, x2, y2, rrRadius)
	rr.NormalizeRadius()

	stroke := conv.NewConvStroke(&rrRoundedRectVS{rr: rr})
	stroke.SetWidth(1.0)

	var lineColor color.RGBA8[color.Linear]
	if rrDarkBg {
		lineColor = color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	} else {
		lineColor = color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	}

	ras.Reset()
	ras.AddPath(&rrStrokeVS{cs: stroke}, 0)
	renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, ren, lineColor)

	applyPremulLinearToSRGB(img)
}

// --- Mouse handlers ---

func handleRoundedRectMouseDown(x, y float64) bool {
	ox, oy := rrCanvasOffset()
	for i, pt := range rrPts {
		dx := x - (pt[0] + ox)
		dy := y - (pt[1] + oy)
		if math.Sqrt(dx*dx+dy*dy) < 8.0 {
			rrSelected = i
			rrDragDX = dx
			rrDragDY = dy
			return true
		}
	}
	return false
}

func handleRoundedRectMouseMove(x, y float64) bool {
	if rrSelected < 0 {
		return false
	}
	ox, oy := rrCanvasOffset()
	rrPts[rrSelected][0] = x - rrDragDX - ox
	rrPts[rrSelected][1] = y - rrDragDY - oy
	return true
}

func handleRoundedRectMouseUp() {
	rrSelected = -1
}
