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
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
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
	img := ctx.GetImage()
	rbuf := buffer.NewRenderingBufferU8()
	rbuf.Attach(img.Data, img.Width(), img.Height(), img.Stride())

	pixFmt := pixfmt.NewPixFmtRGBA32PreLinear(rbuf)
	renBase := renderer.NewRendererBaseWithPixfmt[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]](pixFmt)
	renBase.Clear(color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255})

	sl := scanline.NewScanlineP8()
	slBin := scanline.NewScanlineBin()

	ras := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)

	// 1. Draw anti-aliased triangle
	ps := path.NewPathStorageStl()
	ps.MoveTo(rasterizersX[0], rasterizersY[0])
	ps.LineTo(rasterizersX[1], rasterizersY[1])
	ps.LineTo(rasterizersX[2], rasterizersY[2])
	ps.ClosePolygon(basics.PathFlagsNone)

	cAA := color.RGBA8[color.Linear]{R: 178, G: 127, B: 25, A: uint8(255 * rasterizersAlpha)}

	// Set gamma for AA
	gPower := gamma.NewGammaPower(rasterizersGamma * 2.0)
	ras.SetGamma(gPower.Apply)

	ras.Reset()
	ras.AddPath(&pathSourceAdapter{ps: ps}, 0)

	if ras.RewindScanlines() {
		sl.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(sl) {
			y := sl.Y()
			for _, spanData := range sl.Spans() {
				if spanData.Len > 0 {
					renBase.BlendSolidHspan(int(spanData.X), y, int(spanData.Len), cAA, spanData.Covers)
				} else {
					renBase.BlendHline(int(spanData.X), y, int(spanData.X)-int(spanData.Len)-1, cAA, spanData.Covers[0])
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

	ras.AddPath(&pathSourceAdapter{ps: psAliased}, 0)

	if ras.RewindScanlines() {
		slBin.Reset(ras.MinX(), ras.MaxX())
		for ras.SweepScanline(slBin) {
			renscan.RenderScanlineBinSolid(slBin, renBase, cAliased)
		}
	}

	// 3. Draw interactive handles using low-level ellipse rendering
	handleRas := rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](
		rasterizer.RasConvInt{},
		rasterizer.NewRasterizerSlNoClip(),
	)
	handleSl := scanline.NewScanlineP8()
	fillColor := color.RGBA8[color.Linear]{R: 204, G: 51, B: 26, A: 153}
	outlineColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}
	for i := 0; i < 3; i++ {
		drawRasterizersHandle(handleRas, handleSl, renBase, rasterizersX[i], rasterizersY[i], fillColor, outlineColor)
		drawRasterizersHandle(handleRas, handleSl, renBase, rasterizersX[i]-200, rasterizersY[i], fillColor, outlineColor)
	}

	applyLinearToSRGB(img)
}

func drawRasterizersHandle(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	sl *scanline.ScanlineP8,
	renBase *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]],
	x, y float64,
	fillColor, outlineColor color.RGBA8[color.Linear],
) {
	// Fill the circle
	ell := shapes.NewEllipseWithParams(x, y, 5, 5, 20, false)
	ras.Reset()
	ras.AddPath(&ellipseVS{ell: ell}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, fillColor)

	// Stroke outline using a slightly larger ellipse
	ellOut := shapes.NewEllipseWithParams(x, y, 5.5, 5.5, 20, false)
	ras.Reset()
	ras.AddPath(&ellipseVS{ell: ellOut}, 0)
	renscan.RenderScanlinesAASolid(ras, sl, renBase, outlineColor)
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
