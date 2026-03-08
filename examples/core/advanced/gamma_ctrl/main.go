// Port of AGG C++ gamma_ctrl.cpp – gamma curve control widget demo.
//
// Shows how the rasterizer's gamma curve affects anti-aliased ellipse rendering.
// Six rows of ellipse pairs are drawn at different opacities (0.1, 0.4, 1.0,
// 2.0×grey). The gamma control curve is displayed in the lower-left.
// Default gamma: identity (1.0, 1.0, 1.0, 1.0) control points.
package main

import (
	"math"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/ctrl/gamma"
)

const (
	width  = 800
	height = 600
)

type ellipseRow struct {
	cy    float64
	alpha float64
	color agg.Color
}

func drawEllipsePair(a *agg.Agg2D, cx, cy, wide, small, alpha float64, c agg.Color) {
	// Wide thin ellipse.
	ca := agg.NewColor(c.R, c.G, c.B, uint8(float64(c.A)*alpha))
	a.LineColor(ca)
	a.NoFill()
	a.LineWidth(1.0)
	a.ResetPath()
	const steps = 120
	for i := 0; i <= steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + wide*math.Cos(angle)
		y := cy + small*math.Sin(angle)
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)

	// Small circle.
	a.DrawCircle(cx, cy, small/2)
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	a := ctx.GetAgg2D()
	a.ResetTransformations()

	// Create gamma control (identity curve by default).
	gc := gamma.NewGammaCtrl(10, 340, 310, 585, false)
	gc.SetTextSize(10, 0)
	// Default identity values.
	gc.Values(1.0, 1.0, 1.0, 1.0)

	// Apply the gamma function to the rasterizer.
	ras := a.GetInternalRasterizer()
	ras.SetGamma(gc.Y)

	eWidth := float64(width)/2 - 10
	ecenter := float64(width) / 2

	rows := []ellipseRow{
		{45, 0.1, agg.NewColor(0, 0, 102, 255)},
		{95, 0.4, agg.NewColor(0, 0, 102, 255)},
		{145, 1.0, agg.NewColor(0, 0, 102, 255)},
		{195, 1.0, agg.NewColor(192, 192, 192, 255)},
		{245, 1.0, agg.NewColor(127, 127, 127, 255)},
		{295, 1.0, agg.NewColor(0, 0, 0, 255)},
	}

	for _, row := range rows {
		drawEllipsePair(a, ecenter, row.cy, eWidth, 15.5, row.alpha, row.color)
	}

	// Restore linear gamma for the control rendering.
	ras.SetGamma(func(x float64) float64 { return x })

	// Draw the gamma control widget manually using its vertex paths.
	// (Simplified: draw bounding box of where the widget would appear)
	a.LineColor(agg.NewColor(100, 100, 100, 150))
	a.NoFill()
	a.LineWidth(1.0)
	a.ResetPath()
	a.MoveTo(10, 340)
	a.LineTo(310, 340)
	a.LineTo(310, 585)
	a.LineTo(10, 585)
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Gamma Ctrl",
		Width:  width,
		Height: height,
	}, &demo{})
}
