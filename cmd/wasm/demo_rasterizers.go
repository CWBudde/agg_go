// Based on the original AGG examples: rasterizers.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/gamma"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/renderer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
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
	renscan.RenderScanlinesAASolid(ras, sl, renBase, cAA)

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
	renscan.RenderScanlinesAASolid(ras, sl, renBase, cAliased)

	// 3. Draw interactive handles
	for i := 0; i < 3; i++ {
		// Handles for AA triangle
		drawHandle(rasterizersX[i], rasterizersY[i])
		// Handles for aliased triangle
		drawHandle(rasterizersX[i]-200, rasterizersY[i])
	}
}

func handleRasterizersMouseDown(x, y float64) bool {
	rasterizersSelected = -1
	for i := 0; i < 3; i++ {
		// Check AA triangle handles
		dist := math.Sqrt(math.Pow(x-rasterizersX[i], 2) + math.Pow(y-rasterizersY[i], 2))
		if dist < 10 {
			rasterizersSelected = i
			rasterizersDragDX = x - rasterizersX[i]
			rasterizersDragDY = y - rasterizersY[i]
			return true
		}
		// Check aliased triangle handles
		dist = math.Sqrt(math.Pow(x-(rasterizersX[i]-200), 2) + math.Pow(y-rasterizersY[i], 2))
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
