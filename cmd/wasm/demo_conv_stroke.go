// Port of AGG C++ conv_stroke.cpp – "Line Join" interactive demo.
package main

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	"github.com/cwbudde/agg_go/internal/scanline"
)

// --- State ---

var (
	strokePts = [3][2]float64{
		{157, 160},
		{469, 270},
		{243, 410},
	}
	strokeJoin       = 0 // 0=miter, 1=round, 2=bevel
	strokeCap        = 0 // 0=butt, 1=square, 2=round
	strokeWidth      = 20.0
	strokeMiterLimit = 4.0

	strokeSelected = -1
	strokeDragDX   = 0.0
	strokeDragDY   = 0.0
)

var (
	strokeJoins = []basics.LineJoin{basics.MiterJoin, basics.RoundJoin, basics.BevelJoin}
	strokeCaps  = []basics.LineCap{basics.ButtCap, basics.SquareCap, basics.RoundCap}
)

// --- Drawing ---

func drawConvStrokeDemo() {
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pf := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	ren := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pf)
	ren.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	sl := scanline.NewScanlineU8()

	x := [3]float64{strokePts[0][0], strokePts[1][0], strokePts[2][0]}
	y := [3]float64{strokePts[0][1], strokePts[1][1], strokePts[2][1]}

	lj := strokeJoins[strokeJoin]
	lc := strokeCaps[strokeCap]

	buildPS := func() *path.PathStorageStl {
		ps := path.NewPathStorageStl()
		ps.MoveTo(x[0], y[0])
		ps.LineTo((x[0]+x[1])/2, (y[0]+y[1])/2)
		ps.LineTo(x[1], y[1])
		ps.LineTo(x[2], y[2])
		ps.LineTo(x[2], y[2])

		ps.MoveTo((x[0]+x[1])/2, (y[0]+y[1])/2)
		ps.LineTo((x[1]+x[2])/2, (y[1]+y[2])/2)
		ps.LineTo((x[2]+x[0])/2, (y[2]+y[0])/2)
		ps.ClosePolygon(0)
		return ps
	}

	renderSolid := func(c color.RGBA8[color.Linear]) {
		if ras.RewindScanlines() {
			sl.Reset(ras.MinX(), ras.MaxX())
			for ras.SweepScanline(sl) {
				yy := sl.Y()
				for _, span := range sl.Spans() {
					if span.Len > 0 {
						ren.BlendSolidHspan(int(span.X), yy, int(span.Len), c, span.Covers)
					}
				}
			}
		}
	}

	addPathToRas := func(src conv.VertexSource) {
		src.Rewind(0)
		for {
			vx, vy, cmd := src.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			ras.AddVertex(vx, vy, uint32(cmd))
		}
	}

	// (1) Wide stroked path with selected join/cap.
	{
		ps := buildPS()
		psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
		wideStroke := conv.NewConvStroke(psAdapter)
		wideStroke.SetLineJoin(lj)
		wideStroke.SetLineCap(lc)
		wideStroke.SetMiterLimit(strokeMiterLimit)
		wideStroke.SetWidth(strokeWidth)
		ras.Reset()
		addPathToRas(wideStroke)
		renderSolid(color.RGBA8[color.Linear]{R: 204, G: 178, B: 153, A: 255})
	}

	// (2) Thin outline of the raw path in black.
	{
		ps := buildPS()
		psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
		thinStroke := conv.NewConvStroke(psAdapter)
		thinStroke.SetLineJoin(basics.MiterJoin)
		thinStroke.SetLineCap(basics.ButtCap)
		thinStroke.SetWidth(1.5)
		ras.Reset()
		addPathToRas(thinStroke)
		renderSolid(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255})
	}

	// (3) Dashed thin overlay on the wide stroke.
	// C++ pipeline: path → conv_stroke(wide) → conv_dash → conv_stroke(thin).
	{
		ps := buildPS()
		psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)

		wideStroke := conv.NewConvStroke(psAdapter)
		wideStroke.SetLineJoin(lj)
		wideStroke.SetLineCap(lc)
		wideStroke.SetMiterLimit(strokeMiterLimit)
		wideStroke.SetWidth(strokeWidth)

		dashedStroke := conv.NewConvDash(wideStroke)
		dashedStroke.AddDash(20.0, strokeWidth/2.5)

		thinStroke := conv.NewConvStroke(dashedStroke)
		thinStroke.SetMiterLimit(4.0)
		thinStroke.SetWidth(strokeWidth / 5.0)
		thinStroke.SetLineCap(lc)
		thinStroke.SetLineJoin(lj)

		ras.Reset()
		addPathToRas(thinStroke)
		renderSolid(color.RGBA8[color.Linear]{R: 0, G: 0, B: 77, A: 255})
	}

	// (4) Semi-transparent fill of the raw path.
	{
		ps := buildPS()
		psAdapter := path.NewPathStorageStlVertexSourceAdapter(ps)
		ras.Reset()
		addPathToRas(psAdapter)
		renderSolid(color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 51})
	}

	applyPremulLinearToSRGB(img)

	// Interactive handles (drawn on top via ctx after sRGB encoding).
	for i := 0; i < 3; i++ {
		drawHandle(x[i], y[i])
	}
}

// --- Mouse handlers ---

func handleConvStrokeMouseDown(x, y float64) bool {
	strokeSelected = -1
	for i := 0; i < 3; i++ {
		dx := x - strokePts[i][0]
		dy := y - strokePts[i][1]
		if math.Sqrt(dx*dx+dy*dy) < 15 {
			strokeSelected = i
			strokeDragDX = dx
			strokeDragDY = dy
			return true
		}
	}
	return false
}

func handleConvStrokeMouseMove(x, y float64) bool {
	if strokeSelected < 0 {
		return false
	}
	strokePts[strokeSelected][0] = x - strokeDragDX
	strokePts[strokeSelected][1] = y - strokeDragDY
	return true
}

func handleConvStrokeMouseUp() {
	strokeSelected = -1
}
