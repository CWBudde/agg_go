// Based on the original AGG examples: circles.cpp.
package main

import (
	"math"
	"math/rand"

	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/curves"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/renderer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/shapes"
)

type scatterPoint struct {
	x, y, z float64
	r, g, b float64
}

var (
	circlesPoints []scatterPoint
	splineR       *curves.BSpline
	splineG       *curves.BSpline
	splineB       *curves.BSpline
	numPoints     = 10000

	// Sliders
	selectivity = 0.1
	sizeScale   = 0.5
	zRangeLow   = 0.2
	zRangeHigh  = 0.8

	// Reusable components
	circlesEllipse     *shapes.Ellipse
	circlesInitialized bool
)

func circlesClampU8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return uint8(v*255 + 0.5)
}

func initCircles() {
	if circlesInitialized {
		return
	}

	splineRX := []float64{0.000000, 0.200000, 0.400000, 0.910484, 0.957258, 1.000000}
	splineRY := []float64{1.000000, 0.800000, 0.600000, 0.066667, 0.169697, 0.600000}
	splineGX := []float64{0.000000, 0.292244, 0.485655, 0.564859, 0.795607, 1.000000}
	splineGY := []float64{0.000000, 0.607260, 0.964065, 0.892558, 0.435571, 0.000000}
	splineBX := []float64{0.000000, 0.055045, 0.143034, 0.433082, 0.764859, 1.000000}
	splineBY := []float64{0.385480, 0.128493, 0.021416, 0.271507, 0.713974, 1.000000}

	splineR = curves.NewBSplineFromPoints(splineRX, splineRY)
	splineG = curves.NewBSplineFromPoints(splineGX, splineGY)
	splineB = curves.NewBSplineFromPoints(splineBX, splineBY)

	circlesEllipse = shapes.NewEllipse()

	generateCircles()
	circlesInitialized = true
}

func generateCircles() {
	circlesPoints = make([]scatterPoint, numPoints)
	rx := float64(width) / 3.5
	ry := float64(height) / 3.5

	const twoPi = 2.0 * math.Pi

	for i := 0; i < numPoints; i++ {
		z := rand.Float64()
		x := math.Cos(z*twoPi) * rx
		y := math.Sin(z*twoPi) * ry

		dist := rand.Float64() * (rx * 0.5)
		angle := rand.Float64() * (math.Pi * 2.0)

		circlesPoints[i].z = z
		circlesPoints[i].x = float64(width)*0.5 + x + math.Cos(angle)*dist
		circlesPoints[i].y = float64(height)*0.5 + y + math.Sin(angle)*dist

		circlesPoints[i].r = splineR.Get(z) * 0.8
		circlesPoints[i].g = splineG.Get(z) * 0.8
		circlesPoints[i].b = splineB.Get(z) * 0.8
	}
}

func drawCirclesScatterDemo() {
	initCircles()

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

	radius := sizeScale * 5.0

	for _, p := range circlesPoints {
		z := p.z
		alpha := 1.0

		if z < zRangeLow {
			alpha = 1.0 - (zRangeLow-z)*selectivity*100.0
		} else if z > zRangeHigh {
			alpha = 1.0 - (z-zRangeHigh)*selectivity*100.0
		}

		if alpha > 1.0 {
			alpha = 1.0
		} else if alpha <= 0.0 {
			continue
		}

		circlesEllipse.Init(p.x, p.y, radius, radius, 8, false)
		ras.Reset()
		ras.AddPath(&ellipseVS{ell: circlesEllipse}, 0)
		renscan.RenderScanlinesAASolid[color.RGBA8[color.Linear]](ras, sl, ren, color.RGBA8[color.Linear]{
			R: circlesClampU8(p.r),
			G: circlesClampU8(p.g),
			B: circlesClampU8(p.b),
			A: circlesClampU8(alpha),
		})
	}

	applyPremulLinearToSRGB(img)

	// Update for animation (idle loop in original)
	for i := range circlesPoints {
		circlesPoints[i].x += rand.Float64()*selectivity - selectivity*0.5
		circlesPoints[i].y += rand.Float64()*selectivity - selectivity*0.5
		circlesPoints[i].z += rand.Float64()*selectivity*0.01 - selectivity*0.005
		if circlesPoints[i].z < 0.0 {
			circlesPoints[i].z = 0.0
		} else if circlesPoints[i].z > 1.0 {
			circlesPoints[i].z = 1.0
		}
	}
}
