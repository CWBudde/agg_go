// Port of AGG C++ gamma_correction.cpp – anti-aliasing gamma demonstration.
//
// Shows how the anti-aliasing gamma affects thin ellipse rendering on a split
// dark/light background. Two ellipses are drawn: one thin (1px) and one thick.
// The gamma value is fixed at 1.0 (linear); in the interactive version it's
// adjustable via a slider.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/demorunner"
)

const (
	width  = 800
	height = 600

	gammaValue    = 1.0
	gammaThick    = 1.0
	gammaContrast = 1.0
	gammaRX       = float64(width) / 3.0
	gammaRY       = float64(height) / 3.0
)

func drawEllipseStroked(a *agg.Agg2D, cx, cy, rx, ry, lw float64, c agg.Color) {
	a.LineColor(c)
	a.NoFill()
	a.LineWidth(lw)
	a.ResetPath()
	const steps = 120
	for i := 0; i <= steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + rx*math.Cos(angle)
		y := cy + ry*math.Sin(angle)
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}
	a.ClosePolygon()
	a.DrawPath(agg.StrokeOnly)
}

func drawGammaCurve(a *agg.Agg2D, xStart, yTop, gamma float64) {
	a.LineColor(agg.NewColor(0, 0, 0, 255))
	a.LineWidth(1.5)
	a.NoFill()
	a.ResetPath()
	for i := 0; i <= 256; i++ {
		x := xStart + float64(i)
		y := yTop + 100*(1.0-math.Pow(float64(i)/255.0, gamma))
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}
	a.DrawPath(agg.StrokeOnly)
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(width)
	h := float64(height)
	cx := w / 2
	cy := h / 2

	dark := 1.0 - gammaContrast
	light := gammaContrast
	f := func(v float64) uint8 { return uint8(v*255 + 0.5) }

	// Background: four quadrants.
	a.FillColor(agg.NewColor(f(dark), f(dark), f(dark), 255))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(cx, 0)
	a.LineTo(cx, cy)
	a.LineTo(0, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	a.FillColor(agg.NewColor(f(light), f(light), f(light), 255))
	a.ResetPath()
	a.MoveTo(cx, 0)
	a.LineTo(w, 0)
	a.LineTo(w, cy)
	a.LineTo(cx, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	a.FillColor(agg.NewColor(255, f(dark), f(dark), 255))
	a.ResetPath()
	a.MoveTo(0, cy)
	a.LineTo(w, cy)
	a.LineTo(w, h)
	a.LineTo(0, h)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Apply gamma to the rasterizer.
	a.AntiAliasGamma(gammaValue)

	// Gamma curve visualization in the top half.
	drawGammaCurve(a, cx-128, 20, gammaValue)

	// Thin red ellipse (1 pixel thick).
	drawEllipseStroked(a, cx, cy, gammaRX, gammaRY, gammaThick,
		agg.NewColor(255, 0, 0, 255))

	// Thicker ellipse offset slightly.
	drawEllipseStroked(a, cx, cy, gammaRX*0.8, gammaRY*0.8, gammaThick*3,
		agg.NewColor(255, 80, 80, 200))
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Gamma Correction",
		Width:  width,
		Height: height,
	}, &demo{})
}
