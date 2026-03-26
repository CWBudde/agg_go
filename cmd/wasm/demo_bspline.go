// Port of AGG C++ bspline.cpp – B-Spline Interpolation interactive demo.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	icolor "github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/polygon"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	isl "github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// --- State ---

var (
	bsplinePts = [6][2]float64{} // 6 draggable control points (initialized lazily)

	bsplineClosed    = false
	bsplineNumPoints = 20.0 // interpolation quality (1–40); same as m_num_points default

	bsplineSelected = -1
	bsplineDragDX   = 0.0
	bsplineDragDY   = 0.0
)

// bsplineInit seeds the control points matching the C++ on_init() (flip_y=false).
func bsplineInit() {
	w := float64(width)
	h := float64(height)
	bsplinePts[0] = [2]float64{100, 100}
	bsplinePts[1] = [2]float64{w - 100, 100}
	bsplinePts[2] = [2]float64{w - 100, h - 100}
	bsplinePts[3] = [2]float64{100, h - 100}
	bsplinePts[4] = [2]float64{w / 2, h / 2}
	bsplinePts[5] = [2]float64{w / 2, h / 3}
}

func init() {
	bsplineInit()
}

// --- Drawing ---

func drawBSplineDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32Linear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]](pf)
	ren.Clear(icolor.RGBA8[icolor.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := isl.NewScanlineU8()

	// Flat coordinate slice required by SimplePolygonVertexSource.
	coords := make([]float64, len(bsplinePts)*2)
	for i, p := range bsplinePts {
		coords[i*2] = p[0]
		coords[i*2+1] = p[1]
	}

	// 1. B-spline curve via conv_bspline, rendered as a 2px black stroke.
	src := polygon.NewSimplePolygonVertexSource(coords, uint(len(bsplinePts)), false, bsplineClosed)
	bspline := conv.NewConvBSpline(src)
	bspline.SetInterpolationStep(1.0 / bsplineNumPoints)

	stroke := conv.NewConvStroke(bspline)
	stroke.SetWidth(2.0)

	ras.Reset()
	ras.AddPath(&bsplineConvSource{src: stroke}, 0)
	renscan.RenderScanlinesAASolid(
		ras, sl, ren,
		icolor.RGBA8[icolor.Linear]{R: 0, G: 0, B: 0, A: 255},
	)

	// 2. Control polygon – translucent blue rgba(0, 0.3, 0.5, 0.6).
	ctrlSrc := polygon.NewSimplePolygonVertexSource(coords, uint(len(bsplinePts)), false, bsplineClosed)
	ctrlStroke := conv.NewConvStroke(ctrlSrc)
	ctrlStroke.SetWidth(1.0)

	ras.Reset()
	ras.AddPath(&bsplineConvSource{src: ctrlStroke}, 0)
	renscan.RenderScanlinesAASolid(
		ras, sl, ren,
		icolor.RGBA8[icolor.Linear]{R: 0, G: 76, B: 127, A: 153},
	)

	// 3. Draggable handle circles at each control point.
	for _, p := range bsplinePts {
		bsplineDrawHandle(ras, sl, ren, p[0], p[1])
	}

	applyLinearToSRGB(img)
}

// bsplineDrawHandle renders a filled + outlined circle handle at (x, y).
func bsplineDrawHandle(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *isl.ScanlineU8,
	ren *renderer.RendererBase[*pixfmt.PixFmtRGBA32[icolor.Linear], icolor.RGBA8[icolor.Linear]],
	x, y float64,
) {
	// Filled circle: rgba(0.8, 0.2, 0.1, 0.6)
	ell := shapes.NewEllipseWithParams(x, y, 5, 5, 32, false)
	ras.Reset()
	ras.AddPath(&ellipseVS{ell: ell}, 0)
	renscan.RenderScanlinesAASolid(
		ras, sl, ren,
		icolor.RGBA8[icolor.Linear]{R: 204, G: 51, B: 25, A: 153},
	)

	// Outline circle: black
	ellOut := shapes.NewEllipseWithParams(x, y, 5, 5, 32, false)
	outlineStroke := conv.NewConvStroke(&bsplineEllipseSource{ell: ellOut})
	outlineStroke.SetWidth(1.0)
	ras.Reset()
	ras.AddPath(&bsplineConvSource{src: outlineStroke}, 0)
	renscan.RenderScanlinesAASolid(
		ras, sl, ren,
		icolor.RGBA8[icolor.Linear]{R: 0, G: 0, B: 0, A: 255},
	)
}

// bsplineConvSource adapts a conv.VertexSource to rasterizer.VertexSource.
type bsplineConvSource struct {
	src conv.VertexSource
}

func (a *bsplineConvSource) Rewind(pathID uint32) { a.src.Rewind(uint(pathID)) }
func (a *bsplineConvSource) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.src.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// bsplineEllipseSource adapts shapes.Ellipse to conv.VertexSource.
type bsplineEllipseSource struct {
	ell *shapes.Ellipse
}

func (a *bsplineEllipseSource) Rewind(pathID uint) { a.ell.Rewind(uint32(pathID)) }
func (a *bsplineEllipseSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = a.ell.Vertex(&x, &y)
	return x, y, cmd
}

// --- Mouse handlers ---

func handleBSplineMouseDown(x, y float64) bool {
	bsplineSelected = -1
	for i, p := range bsplinePts {
		dx := x - p[0]
		dy := y - p[1]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			bsplineSelected = i
			bsplineDragDX = dx
			bsplineDragDY = dy
			return true
		}
	}
	return false
}

func handleBSplineMouseMove(x, y float64) bool {
	if bsplineSelected < 0 {
		return false
	}
	bsplinePts[bsplineSelected][0] = x - bsplineDragDX
	bsplinePts[bsplineSelected][1] = y - bsplineDragDY
	return true
}

func handleBSplineMouseUp() {
	bsplineSelected = -1
}
