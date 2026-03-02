// Based on the original AGG examples: circles.cpp.
package main

import (
	"math"
	"math/rand"

	agg "agg_go"
	"agg_go/internal/curves"
)

type scatterPoint struct {
	x, y, z float64
	color   agg.Color
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
)

func initCircles() {
	if len(circlesPoints) > 0 {
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

	generateCircles()
}

func generateCircles() {
	circlesPoints = make([]scatterPoint, numPoints)
	rx := float64(width) / 3.5
	ry := float64(height) / 3.5

	for i := 0; i < numPoints; i++ {
		z := rand.Float64()
		x := math.Cos(z*2.0*math.Pi) * rx
		y := math.Sin(z*2.0*math.Pi) * ry

		dist := rand.Float64() * (rx / 2.0)
		angle := rand.Float64() * (math.Pi * 2.0)

		circlesPoints[i].z = z
		circlesPoints[i].x = float64(width)/2.0 + x + math.Cos(angle)*dist
		circlesPoints[i].y = float64(height)/2.0 + y + math.Sin(angle)*dist

		r := splineR.Get(z) * 0.8
		g := splineG.Get(z) * 0.8
		b := splineB.Get(z) * 0.8
		circlesPoints[i].color = agg.NewColor(uint8(r*255), uint8(g*255), uint8(b*255), 255)
	}
}

func drawCirclesScatterDemo() {
	initCircles()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.NoLine()

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
		}
		if alpha < 0.0 {
			alpha = 0.0
		}

		if alpha > 0.0 {
			color := p.color.WithAlphaF(alpha)
			agg2d.FillColor(color)
			radius := sizeScale * 5.0
			agg2d.FillCircle(p.x, p.y, radius)
		}
	}

	// Update for animation (idle loop in original)
	for i := range circlesPoints {
		circlesPoints[i].x += rand.Float64()*selectivity - selectivity*0.5
		circlesPoints[i].y += rand.Float64()*selectivity - selectivity*0.5
		circlesPoints[i].z += rand.Float64()*selectivity*0.01 - selectivity*0.005
		if circlesPoints[i].z < 0.0 {
			circlesPoints[i].z = 0.0
		}
		if circlesPoints[i].z > 1.0 {
			circlesPoints[i].z = 1.0
		}
	}
}
