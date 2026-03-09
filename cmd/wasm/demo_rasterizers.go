// Based on the original AGG examples: rasterizers.cpp.
package main

import (
	"math"

	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/gamma"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
)

var (
	rasterizersX        = [3]float64{100 + 120, 369 + 120, 143 + 120}
	rasterizersY        = [3]float64{60, 170, 310}
	rasterizersGamma    = 0.5
	rasterizersAlpha    = 1.0
	rasterizersSelected = -1
	rasterizersDragDX   = 0.0
	rasterizersDragDY   = 0.0
)

func drawRasterizersDemo() {
	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()

	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Width()*4)

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	sl := scanline.NewScanlineU8()

	// 1. Draw anti-aliased triangle
	ps := path.NewPathStorageStl()
	ps.MoveTo(rasterizersX[0], rasterizersY[0])
	ps.LineTo(rasterizersX[1], rasterizersY[1])
	ps.LineTo(rasterizersX[2], rasterizersY[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	cAA := color.RGBA8[color.Linear]{R: 178, G: 127, B: 25, A: uint8(255 * rasterizersAlpha)}

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)

	// Set gamma for AA
	gPower := gamma.NewGammaPower(rasterizersGamma * 2.0)
	ras.SetGamma(gPower.Apply)

	adapter := &pathSourceAdapter{ps: ps}
	ras.AddPath(adapter, 0)

	// Use manual sweep loop to avoid interface mismatches
	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), cAA, spanData.Covers)
				}
			}
		}
	}

	// 2. Draw aliased triangle (shifted by -200)
	psAliased := path.NewPathStorageStl()
	psAliased.MoveTo(rasterizersX[0]-200, rasterizersY[0])
	psAliased.LineTo(rasterizersX[1]-200, rasterizersY[1])
	psAliased.LineTo(rasterizersX[2]-200, rasterizersY[2])
	psAliased.ClosePolygon(basics.PathFlagsNone)

	cAliased := color.RGBA8[color.Linear]{R: 25, G: 127, B: 178, A: uint8(255 * rasterizersAlpha)}

	ras.Reset()
	// Set gamma threshold for aliased rendering
	gThreshold := gamma.NewGammaThreshold(rasterizersGamma)
	ras.SetGamma(gThreshold.Apply)

	adapterAliased := &pathSourceAdapter{ps: psAliased}
	ras.AddPath(adapterAliased, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(&rasScanlineAdapter{sl: sl}) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), cAliased, spanData.Covers)
				}
			}
		}
	}

	// 3. Draw interactive handles
	for i := 0; i < 3; i++ {
		drawHandle(rasterizersX[i], rasterizersY[i])
		drawHandle(rasterizersX[i]-200, rasterizersY[i])
	}
}

func handleRasterizersMouseDown(x, y float64) bool {
	rasterizersSelected = -1
	for i := 0; i < 3; i++ {
		dist := math.Sqrt((x-rasterizersX[i])*(x-rasterizersX[i]) + (y-rasterizersY[i])*(y-rasterizersY[i]))
		if dist < 10 {
			rasterizersSelected = i
			rasterizersDragDX = x - rasterizersX[i]
			rasterizersDragDY = y - rasterizersY[i]
			return true
		}
		dist = math.Sqrt((x-rasterizersX[i]-200)*(x-rasterizersX[i]-200) + (y-rasterizersY[i])*(y-rasterizersY[i]))
		if dist < 10 {
			rasterizersSelected = i
			rasterizersDragDX = x - (rasterizersX[i] - 200)
			rasterizersDragDY = y - rasterizersY[i]
			return true
		}
	}
	return false
}

func handleRasterizersMouseMove(x, y float64) bool {
	if rasterizersSelected != -1 {
		rasterizersX[rasterizersSelected] = x - rasterizersDragDX
		rasterizersY[rasterizersSelected] = y - rasterizersDragDY
		return true
	}
	return false
}

func handleRasterizersMouseUp() {
	rasterizersSelected = -1
}
