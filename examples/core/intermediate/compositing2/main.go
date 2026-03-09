// Port of AGG C++ compositing2.cpp – compositing with color ramps and blend modes.
//
// Shows source-over compositing of radial gradient circles using a variety of
// Porter-Duff composition operators. Default: alpha_src=1.0, alpha_dst=1.0,
// operator=SrcOver (displayed as a grid of all operators).
package main

import (
	"math"

	agg "agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

const (
	width  = 800
	height = 600
)

// drawRadialGradientCircle draws a circle filled with a radial gradient from c1 to c2.
func drawRadialGradientCircle(a *agg.Agg2D, cx, cy, r float64, c1, c2 agg.Color) {
	steps := 64
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		angle1 := 2 * math.Pi * float64(i) / float64(steps)
		angle2 := 2 * math.Pi * float64(i+1) / float64(steps)

		// Interpolate color.
		rc := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
		gc := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
		bc := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
		ac := uint8(float64(c1.A)*(1-t) + float64(c2.A)*t)

		a.FillColor(agg.NewColor(rc, gc, bc, ac))
		a.NoLine()
		a.ResetPath()
		a.MoveTo(cx, cy)
		a.LineTo(cx+r*math.Cos(angle1), cy+r*math.Sin(angle1))
		a.LineTo(cx+r*math.Cos(angle2), cy+r*math.Sin(angle2))
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)
	}
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.RGBA(0.5, 0.5, 0.5, 1.0))

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// All compositing operators to demonstrate.
	type opEntry struct {
		name string
		mode agg.BlendMode
	}
	ops := []opEntry{
		{"Alpha", agg.BlendAlpha},
		{"Multiply", agg.BlendMultiply},
		{"Screen", agg.BlendScreen},
		{"Overlay", agg.BlendOverlay},
		{"Darken", agg.BlendDarken},
		{"Lighten", agg.BlendLighten},
		{"ColorDodge", agg.BlendColorDodge},
		{"ColorBurn", agg.BlendColorBurn},
		{"HardLight", agg.BlendHardLight},
		{"SoftLight", agg.BlendSoftLight},
		{"Difference", agg.BlendDifference},
		{"Exclusion", agg.BlendExclusion},
	}

	const (
		cols   = 4
		cellW  = 200.0
		cellH  = 150.0
		radius = 50.0
	)

	for i, op := range ops {
		col := i % cols
		row := i / cols
		cx := float64(col)*cellW + cellW/2
		cy := float64(row)*cellH + cellH/2

		// Background cell.
		a.BlendMode(agg.BlendAlpha)
		a.FillColor(agg.RGBA(0.8, 0.8, 0.8, 0.5))
		a.NoLine()
		a.ResetPath()
		a.MoveTo(cx-cellW/2+3, cy-cellH/2+3)
		a.LineTo(cx+cellW/2-3, cy-cellH/2+3)
		a.LineTo(cx+cellW/2-3, cy+cellH/2-3)
		a.LineTo(cx-cellW/2+3, cy+cellH/2-3)
		a.ClosePolygon()
		a.DrawPath(agg.FillOnly)

		// Source circle (magenta→yellow radial gradient).
		a.BlendMode(agg.BlendAlpha)
		drawRadialGradientCircle(a, cx-15, cy, radius,
			agg.NewColor(255, 0, 128, 255),
			agg.NewColor(255, 255, 0, 64))

		// Destination circle with the blend mode (cyan→transparent).
		a.BlendMode(op.mode)
		drawRadialGradientCircle(a, cx+15, cy, radius,
			agg.NewColor(0, 128, 255, 255),
			agg.NewColor(0, 255, 128, 64))

		_ = op.name
	}

	a.BlendMode(agg.BlendAlpha)
}

func main() {
	demorunner.Run(demorunner.Config{Title: "Compositing 2", Width: width, Height: height}, &demo{})
}
