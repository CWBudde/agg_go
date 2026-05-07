// Based on the original AGG examples: gouraud.cpp.
package main

import (
	"fmt"
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
	"github.com/cwbudde/agg_go/internal/span"
)

var (
	gouraudX        = [3]float64{100, 500, 300}
	gouraudY        = [3]float64{100, 150, 500}
	gouraudDilation = 0.5
	gouraudOpacity  = 1.0
	gouraudSelected = -1
	gouraudDragDX   = 0.0
	gouraudDragDY   = 0.0
)

// gouraudEllipseVS adapts shapes.Ellipse to the rasterizer VertexSource interface.
type gouraudEllipseVS struct {
	e *shapes.Ellipse
}

func (s *gouraudEllipseVS) Rewind(pathID uint32) { s.e.Rewind(pathID) }

func (s *gouraudEllipseVS) Vertex(x, y *float64) uint32 {
	return uint32(s.e.Vertex(x, y))
}

// gouraudStrokeVS adapts conv.ConvStroke to the rasterizer VertexSource interface.
type gouraudStrokeVS struct {
	cs *conv.ConvStroke
}

func (s *gouraudStrokeVS) Rewind(pathID uint32) { s.cs.Rewind(uint(pathID)) }

func (s *gouraudStrokeVS) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := s.cs.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// gouraudRasAdapter adapts SpanGouraudRGBA to the rasterizer VertexSource interface.
type gouraudRasAdapter struct {
	sg interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}
}

func (a *gouraudRasAdapter) Rewind(pathID uint32) {
	a.sg.Rewind(uint(pathID))
}

func (a *gouraudRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.sg.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// gouraudSpanRenderer renders Gouraud-shaded spans into a RendererBase.
type gouraudSpanRenderer struct {
	ren   *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
	span  *span.SpanGouraudRGBA
	alloc *span.SpanAllocator[span.RGBAColor]
}

func (r *gouraudSpanRenderer) Prepare() {
	r.span.Prepare()
}

func (r *gouraudSpanRenderer) SetColor(_ color.RGBA8[color.Linear]) {}

func (r *gouraudSpanRenderer) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.BeginIterator()
	for {
		spanData := it.GetSpan()
		x := spanData.X
		length := spanData.Len

		colors := r.alloc.Allocate(length)
		r.span.Generate(colors, x, y, uint(length))

		baseColors := make([]color.RGBA8[color.Linear], length)
		for i := 0; i < length; i++ {
			baseColors[i] = color.RGBA8[color.Linear]{
				R: uint8(colors[i].R),
				G: uint8(colors[i].G),
				B: uint8(colors[i].B),
				A: uint8(colors[i].A),
			}
		}

		r.ren.BlendColorHspan(x, y, length, baseColors, spanData.Covers, 255)

		if !it.Next() {
			break
		}
	}
}

func renderGouraudTriangle(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineU8,
	ren *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	x1, y1, x2, y2, x3, y3 float64,
	c1, c2, c3 color.RGBA8[color.Linear],
	dilation float64,
) {
	gc1 := span.RGBAColor{R: int(c1.R), G: int(c1.G), B: int(c1.B), A: int(c1.A)}
	gc2 := span.RGBAColor{R: int(c2.R), G: int(c2.G), B: int(c2.B), A: int(c2.A)}
	gc3 := span.RGBAColor{R: int(c3.R), G: int(c3.G), B: int(c3.B), A: int(c3.A)}

	spanGen := span.NewSpanGouraudRGBAWithTriangle(gc1, gc2, gc3, x1, y1, x2, y2, x3, y3, dilation)

	spanRen := &gouraudSpanRenderer{
		ren:   ren,
		span:  spanGen,
		alloc: span.NewSpanAllocator[span.RGBAColor](),
	}

	adapter := &gouraudRasAdapter{sg: spanGen}

	ras.Reset()
	ras.AddPath(adapter, 0)

	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	spanRen.Prepare()
	for ras.SweepScanline(sl) {
		spanRen.Render(sl)
	}
}

func drawGouraudDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{}, rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	alpha := uint8(gouraudOpacity * 255)
	logStatus(fmt.Sprintf("Gouraud Dilation: %.2f  Opacity: %.2f", gouraudDilation, gouraudOpacity))

	// Subdivision into 6 triangles as in original gouraud.cpp
	xc := (gouraudX[0] + gouraudX[1] + gouraudX[2]) / 3.0
	yc := (gouraudY[0] + gouraudY[1] + gouraudY[2]) / 3.0

	x1 := (gouraudX[1]+gouraudX[0])*0.5 - (xc - (gouraudX[1]+gouraudX[0])*0.5)
	y1 := (gouraudY[1]+gouraudY[0])*0.5 - (yc - (gouraudY[1]+gouraudY[0])*0.5)

	x2 := (gouraudX[2]+gouraudX[1])*0.5 - (xc - (gouraudX[2]+gouraudX[1])*0.5)
	y2 := (gouraudY[2]+gouraudY[1])*0.5 - (yc - (gouraudY[2]+gouraudY[1])*0.5)

	x3 := (gouraudX[0]+gouraudX[2])*0.5 - (xc - (gouraudX[0]+gouraudX[2])*0.5)
	y3 := (gouraudY[0]+gouraudY[2])*0.5 - (yc - (gouraudY[0]+gouraudY[2])*0.5)

	cRed := color.RGBA8[color.Linear]{R: 255, G: 0, B: 0, A: alpha}
	cGreen := color.RGBA8[color.Linear]{R: 0, G: 255, B: 0, A: alpha}
	cBlue := color.RGBA8[color.Linear]{R: 0, G: 0, B: 255, A: alpha}
	cWhite := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: alpha}
	cBlack := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: alpha}

	d := gouraudDilation

	// First three triangles (center-based, white center vertex)
	renderGouraudTriangle(ras, sl, ren, gouraudX[0], gouraudY[0], gouraudX[1], gouraudY[1], xc, yc, cRed, cGreen, cWhite, d)
	renderGouraudTriangle(ras, sl, ren, gouraudX[1], gouraudY[1], gouraudX[2], gouraudY[2], xc, yc, cGreen, cBlue, cWhite, d)
	renderGouraudTriangle(ras, sl, ren, gouraudX[2], gouraudY[2], gouraudX[0], gouraudY[0], xc, yc, cBlue, cRed, cWhite, d)

	// Next three triangles (edge-based, black outer vertex)
	renderGouraudTriangle(ras, sl, ren, gouraudX[0], gouraudY[0], gouraudX[1], gouraudY[1], x1, y1, cRed, cGreen, cBlack, d)
	renderGouraudTriangle(ras, sl, ren, gouraudX[1], gouraudY[1], gouraudX[2], gouraudY[2], x2, y2, cGreen, cBlue, cBlack, d)
	renderGouraudTriangle(ras, sl, ren, gouraudX[2], gouraudY[2], gouraudX[0], gouraudY[0], x3, y3, cBlue, cRed, cBlack, d)

	// Draw interactive handles
	handleFill := color.RGBA8[color.Linear]{R: 200, G: 50, B: 20, A: 150}
	handleLine := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	for i := 0; i < 3; i++ {
		hx, hy := gouraudX[i], gouraudY[i]

		// Filled circle
		ell := shapes.NewEllipseWithParams(hx, hy, 8, 8, 32, false)
		ras.Reset()
		ras.AddPath(&gouraudEllipseVS{e: ell}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, ren, handleFill)

		// Circle outline via stroke
		ellSrc := newGouraudEllipseConvSrc(ell)
		stroke := conv.NewConvStroke(ellSrc)
		stroke.SetWidth(1.0)
		ras.Reset()
		ras.AddPath(&gouraudStrokeVS{cs: stroke}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, ren, handleLine)
	}

	applyLinearToSRGB(img)
}

// gouraudEllipseConvSrc adapts shapes.Ellipse to conv.VertexSource.
type gouraudEllipseConvSrc struct {
	e *shapes.Ellipse
}

func newGouraudEllipseConvSrc(e *shapes.Ellipse) *gouraudEllipseConvSrc {
	return &gouraudEllipseConvSrc{e: e}
}

func (s *gouraudEllipseConvSrc) Rewind(pathID uint) { s.e.Rewind(uint32(pathID)) }

func (s *gouraudEllipseConvSrc) Vertex() (float64, float64, basics.PathCommand) {
	var x, y float64
	cmd := s.e.Vertex(&x, &y)
	return x, y, basics.PathCommand(cmd)
}

func handleGouraudMouseDown(x, y float64) bool {
	gouraudSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt((x-gouraudX[i])*(x-gouraudX[i]) + (y-gouraudY[i])*(y-gouraudY[i]))
		if dist < 20 {
			gouraudSelected = i
			gouraudDragDX = x - gouraudX[i]
			gouraudDragDY = y - gouraudY[i]
			return true
		}
	}
	return false
}

func handleGouraudMouseMove(x, y float64) bool {
	if gouraudSelected != -1 {
		gouraudX[gouraudSelected] = x - gouraudDragDX
		gouraudY[gouraudSelected] = y - gouraudDragDY
		return true
	}
	return false
}

func handleGouraudMouseUp() {
	gouraudSelected = -1
}
