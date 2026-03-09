// Port of AGG's polymorphic_renderer.cpp.
//
// Demonstrates the Go equivalent of C++ virtual-dispatch polymorphism:
// the same rendering code operates uniformly on different pixel-format
// backends through a Go interface, without virtual keyword or base class.
//
// Visual: a single filled triangle on a white background. Drag the three
// corner handles to reshape it.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

// --- State ---

var (
	polyRenX        = [3]float64{100, 369, 143}
	polyRenY        = [3]float64{60, 170, 310}
	polyRenSelected = -1
	polyRenDragDX   = 0.0
	polyRenDragDY   = 0.0
)

// --- Rendering ---

func drawPolymorphicRendererDemo() {
	ctx.GetAgg2D().ResetTransformations()

	img := ctx.GetImage()
	w, h := img.Width(), img.Height()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, w, h, w*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.ClipBox(0, 0, w, h)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	ps := path.NewPathStorageStl()
	ps.MoveTo(polyRenX[0], polyRenY[0])
	ps.LineTo(polyRenX[1], polyRenY[1])
	ps.LineTo(polyRenX[2], polyRenY[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	fillColor := color.RGBA8[color.Linear]{R: 80, G: 30, B: 20, A: 255}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	sl := scanline.NewScanlineU8()
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, sp := range sl.Spans() {
				if sp.Len > 0 {
					renBase.BlendSolidHspan(int(sp.X), y, int(sp.Len), fillColor, sp.Covers)
				}
			}
		}
	}

	// Draw interactive vertex handles.
	for i := 0; i < 3; i++ {
		drawHandle(polyRenX[i], polyRenY[i])
	}
}

// --- Mouse handlers ---

func handlePolyRenMouseDown(x, y float64) bool {
	polyRenSelected = -1
	for i := 0; i < 3; i++ {
		dx := x - polyRenX[i]
		dy := y - polyRenY[i]
		if math.Sqrt(dx*dx+dy*dy) < 10 {
			polyRenSelected = i
			polyRenDragDX = dx
			polyRenDragDY = dy
			return true
		}
	}
	return false
}

func handlePolyRenMouseMove(x, y float64) bool {
	if polyRenSelected < 0 {
		return false
	}
	polyRenX[polyRenSelected] = x - polyRenDragDX
	polyRenY[polyRenSelected] = y - polyRenDragDY
	return true
}

func handlePolyRenMouseUp() {
	polyRenSelected = -1
}
