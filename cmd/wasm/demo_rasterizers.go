// Based on the original AGG examples: rasterizers.cpp.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/internal/basics"
	"agg_go/internal/gamma"
	"agg_go/internal/path"
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

	// 1. Draw anti-aliased triangle
	ps := path.NewPathStorageStl()
	ps.MoveTo(rasterizersX[0], rasterizersY[0])
	ps.LineTo(rasterizersX[1], rasterizersY[1])
	ps.LineTo(rasterizersX[2], rasterizersY[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	agg2d.SetFillColor(agg.NewColorRGBA8(agg.SRGBA8(178, 127, 25, uint8(255*rasterizersAlpha))))
	
	ras := agg2d.GetInternalRasterizer()
	ras.Reset()
	
	// Set gamma for AA
	gPower := gamma.NewGammaPower(rasterizersGamma * 2.0)
	ras.SetGamma(&gPower)
	
	adapter := &pathSourceAdapter{ps: ps}
	ras.AddPath(adapter, 0)
	agg2d.DrawPath(ras)

	// 2. Draw aliased triangle (shifted by -200)
	psAliased := path.NewPathStorageStl()
	psAliased.MoveTo(rasterizersX[0]-200, rasterizersY[0])
	psAliased.LineTo(rasterizersX[1]-200, rasterizersY[1])
	psAliased.LineTo(rasterizersX[2]-200, rasterizersY[2])
	psAliased.ClosePolygon(basics.PathFlagsNone)

	agg2d.SetFillColor(agg.NewColorRGBA8(agg.SRGBA8(25, 127, 178, uint8(255*rasterizersAlpha))))
	
	ras.Reset()
	// Set gamma threshold for aliased rendering
	gThreshold := gamma.NewGammaThreshold(rasterizersGamma)
	ras.SetGamma(&gThreshold)
	
	adapterAliased := &pathSourceAdapter{ps: psAliased}
	ras.AddPath(adapterAliased, 0)
	agg2d.DrawPath(ras)

	// 3. Draw interactive handles
	for i := 0; i < 3; i++ {
		// Handles for AA triangle
		drawHandle(rasterizersX[i], rasterizersY[i])
		// Handles for aliased triangle
		drawHandle(rasterizersX[i]-200, rasterizersY[i])
	}
}

func drawHandle(x, y float64) {
	ctx.SetColor(agg.RGBA(0.8, 0.2, 0.1, 0.6))
	ctx.FillCircle(x, y, 5)
	ctx.SetColor(agg.Black)
	ctx.DrawCircle(x, y, 5)
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
