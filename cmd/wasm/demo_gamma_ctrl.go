// Based on the original AGG examples: gamma_ctrl.cpp.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	gammactrl "github.com/MeKo-Christian/agg_go/internal/ctrl/gamma"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

var gammaControl *gammactrl.GammaCtrl

type GammaControl = gammactrl.GammaCtrl

func initGammaCtrlDemo() {
	if gammaControl == nil {
		// Position control in the lower-left area of the 800x600 canvas.
		// Original C++ used (10,10,300,200) with flip_y; here Y increases downward.
		gammaControl = gammactrl.NewGammaCtrl(10, 340, 310, 585, false)
		gammaControl.SetTextSize(10, 0)
	}
}

// gcPixFmt is the concrete pixfmt type for this demo (non-premultiplied, linear, no sRGB).
type gcPixFmt = pixfmt.PixFmtRGBA32[color.Linear]

// gcRendererBase is the concrete renderer base type for this demo.
type gcRendererBase = renderer.RendererBase[*gcPixFmt, color.RGBA8[color.Linear]]

// gcRasType is the concrete rasterizer type for this demo.
type gcRasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

func newGCRasterizer() *gcRasType {
	return rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
}

// ellipseToConvVS adapts shapes.Ellipse (pointer-based Vertex) to conv.VertexSource.
type ellipseToConvVS struct {
	e *shapes.Ellipse
}

func (a *ellipseToConvVS) Rewind(pathID uint) {
	a.e.Rewind(uint32(pathID))
}

func (a *ellipseToConvVS) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.e.Vertex(&x, &y)
	return x, y, cmd
}

// renderStrokedEllipseGC renders a stroked ellipse directly into the rasterizer/renderer.
func renderStrokedEllipseGC(
	ras *gcRasType,
	sl *scanline.ScanlineU8,
	ren *gcRendererBase,
	cx, cy, rx, ry, strokeWidth float64,
	c color.RGBA8[color.Linear],
) {
	e := shapes.NewEllipseWithParams(cx, cy, rx, ry, 100, false)
	stroke := conv.NewConvStroke(&ellipseToConvVS{e: e})
	stroke.SetWidth(strokeWidth)
	adapter := conv.NewRasterizerVertexSourceAdapter(stroke)
	ras.Reset()
	ras.AddPath(adapter, 0)
	renscan.RenderScanlinesAASolid(ras, sl, ren, c)
}

// renderTextGC renders stroked GSV text at the given position.
func renderTextGC(
	ras *gcRasType,
	sl *scanline.ScanlineU8,
	ren *gcRendererBase,
	x, y, size float64,
	text string,
	c color.RGBA8[color.Linear],
) {
	t := gsv.NewGSVText()
	t.SetText(text)
	t.SetSize(size, 0)
	t.SetStartPoint(x, y)
	outline := gsv.NewGSVTextOutlineWithTransform(t, transform.NewTransAffineSkewing(0.15, 0.0))
	outline.SetWidth(2.0)
	adapter := conv.NewRasterizerVertexSourceAdapter(outline)
	ras.Reset()
	ras.AddPath(adapter, 0)
	renscan.RenderScanlinesAASolid(ras, sl, ren, c)
}

// drawArrowPairGC renders one pair of filled arrow triangles rotated by angle around (cx, cy).
func drawArrowPairGC(
	ras *gcRasType,
	sl *scanline.ScanlineU8,
	ren *gcRendererBase,
	cx, cy, angle float64,
	c color.RGBA8[color.Linear],
) {
	rotate := func(dx, dy float64) (float64, float64) {
		s, cs := math.Sincos(angle)
		return cx + dx*cs - dy*s, cy + dx*s + dy*cs
	}

	p0x, p0y := rotate(30, -1)
	p1x, p1y := rotate(60, 0)
	p2x, p2y := rotate(30, 1)

	p3x, p3y := rotate(27, -1)
	p4x, p4y := rotate(10, 0)
	p5x, p5y := rotate(27, 1)

	ras.Reset()
	ras.MoveToD(p0x, p0y)
	ras.LineToD(p1x, p1y)
	ras.LineToD(p2x, p2y)
	ras.MoveToD(p3x, p3y)
	ras.LineToD(p4x, p4y)
	ras.LineToD(p5x, p5y)
	renscan.RenderScanlinesAASolid(ras, sl, ren, c)
}

func drawGammaCtrlDemo() {
	initGammaCtrlDemo()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32[color.Linear](rbuf)
	ren := renderer.NewRendererBaseWithPixfmt(pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := newGCRasterizer()
	sl := scanline.NewScanlineU8()

	ewidth := float64(width)/2 - 10
	ecenter := float64(width) / 2

	// Apply gamma from the control to the rasterizer before drawing shapes.
	ras.SetGamma(gammaControl.Y)

	darkBlue := color.RGBA8[color.Linear]{R: 0, G: 0, B: 102, A: 255}
	lightGray := color.RGBA8[color.Linear]{R: 192, G: 192, B: 192, A: 255}
	midGray := color.RGBA8[color.Linear]{R: 127, G: 127, B: 127, A: 255}
	black := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	// Six ellipse pairs ordered top-to-bottom, matching the original flip_y layout.
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 45, ewidth, 15.5, 0.1, darkBlue)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 45, 10.5, 10.5, 0.1, darkBlue)

	renderStrokedEllipseGC(ras, sl, ren, ecenter, 95, ewidth, 15.5, 0.4, darkBlue)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 95, 10.5, 10.5, 0.4, darkBlue)

	renderStrokedEllipseGC(ras, sl, ren, ecenter, 145, ewidth, 15.5, 1.0, darkBlue)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 145, 10.5, 10.5, 1.0, darkBlue)

	renderStrokedEllipseGC(ras, sl, ren, ecenter, 195, ewidth, 15, 2.0, lightGray)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 195, 11, 11, 2.0, lightGray)

	renderStrokedEllipseGC(ras, sl, ren, ecenter, 245, ewidth, 15, 2.0, midGray)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 245, 11, 11, 2.0, midGray)

	renderStrokedEllipseGC(ras, sl, ren, ecenter, 295, ewidth, 15, 2.0, black)
	renderStrokedEllipseGC(ras, sl, ren, ecenter, 295, 11, 11, 2.0, black)

	// Render text and arrows without gamma correction.
	ras.SetGamma(func(x float64) float64 { return x })

	// Draw text in lower-right, matching original start_point(320,10) after flip_y.
	renderTextGC(ras, sl, ren, 370, 555, 50, "Text 2345", color.RGBA8[color.Linear]{R: 0, G: 127, B: 0, A: 255})

	// Rotating arrows to the right of the gamma control.
	arrowColor := color.RGBA8[color.Linear]{R: 127, G: 0, B: 0, A: 255}
	for i := 0; i < 35; i++ {
		angle := float64(i) / 35.0 * 2.0 * math.Pi
		drawArrowPairGC(ras, sl, ren, 490, 415, angle, arrowColor)
	}
}

func handleGammaCtrlMouseDown(x, y float64) bool {
	if gammaControl == nil {
		return false
	}
	return gammaControl.OnMouseButtonDown(x, y)
}

func handleGammaCtrlMouseMove(x, y float64) bool {
	if gammaControl == nil {
		return false
	}
	return gammaControl.OnMouseMove(x, y, true)
}

func handleGammaCtrlMouseUp() {
	if gammaControl != nil {
		gammaControl.OnMouseButtonUp(0, 0)
	}
}
