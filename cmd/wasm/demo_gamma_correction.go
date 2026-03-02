// Port of AGG C++ gamma_correction.cpp – "Thin red ellipse / gamma correction".
//
// Shows how the anti-aliasing gamma affects the visual quality of thin
// colored lines rendered over a split dark/light background. The demo
// uses AntiAliasGamma to control the rasterizer's coverage-to-alpha mapping,
// which is the closest equivalent to the C++ pixfmt_gamma approach.
package main

import (
	"math"

	agg "agg_go"
)

// --- State ---

var (
	gammaValue    = 1.0
	gammaThick    = 1.0 // line thickness
	gammaContrast = 1.0 // 0=no contrast, 1=full

	// Ellipse radii – updated by mouse drag.
	gammaRX = float64(width) / 3.0
	gammaRY = float64(height) / 3.0
)

// --- Drawing ---

func drawGammaCorrectionDemo() {
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(width)
	h := float64(height)
	cx := w / 2
	cy := h / 2

	dark := 1.0 - gammaContrast
	light := gammaContrast
	f := func(v float64) uint8 { return uint8(v*255 + 0.5) }

	// Background: four quadrants, matching the C++ copy_bar sequence.
	// Top-left: dark gray
	a.FillColor(agg.NewColor(f(dark), f(dark), f(dark), 255))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(cx, 0)
	a.LineTo(cx, cy)
	a.LineTo(0, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Top-right: light gray
	a.FillColor(agg.NewColor(f(light), f(light), f(light), 255))
	a.ResetPath()
	a.MoveTo(cx, 0)
	a.LineTo(w, 0)
	a.LineTo(w, cy)
	a.LineTo(cx, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Bottom half: reddish (1, dark, dark) – overwrites both sides as in C++.
	a.FillColor(agg.NewColor(255, f(dark), f(dark), 255))
	a.ResetPath()
	a.MoveTo(0, cy)
	a.LineTo(w, cy)
	a.LineTo(w, h)
	a.LineTo(0, h)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Apply the gamma to the rasterizer's anti-aliasing coverage mapping.
	a.AntiAliasGamma(gammaValue)

	// Draw the gamma power curve in the top area (like the C++ demo).
	// The curve maps [0,255] through gamma power.
	drawGammaCurve(a, cx-128, 20, gammaValue)

	// 5 concentric stroked ellipses: Red, Green, Blue, Black, White.
	type ellipseSpec struct {
		dr    float64
		color agg.Color
	}
	specs := []ellipseSpec{
		{0, agg.NewColor(255, 0, 0, 255)},
		{5, agg.NewColor(0, 200, 0, 255)},
		{10, agg.NewColor(0, 0, 255, 255)},
		{15, agg.NewColor(0, 0, 0, 255)},
		{20, agg.NewColor(255, 255, 255, 255)},
	}
	a.NoFill()
	a.LineWidth(gammaThick)
	for _, s := range specs {
		a.LineColor(s.color)
		a.ResetPath()
		a.Ellipse(cx, cy, gammaRX-s.dr, gammaRY-s.dr)
		a.DrawPath(agg.StrokeOnly)
	}

	// Reset gamma to default so other demos are not affected.
	a.AntiAliasGamma(1.0)
}

// drawGammaCurve draws the gamma power curve as a thin green polyline.
func drawGammaCurve(a *agg.Agg2D, startX, startY, gamma float64) {
	const npts = 256
	a.LineColor(agg.NewColor(80, 160, 80, 255))
	a.LineWidth(2.0)
	a.NoFill()
	a.ResetPath()
	for i := 0; i < npts; i++ {
		v := float64(i) / float64(npts-1)
		gv := math.Pow(v, gamma)
		px := startX + float64(i)
		py := startY + gv*80 // 80px height for the curve
		if i == 0 {
			a.MoveTo(px, py)
		} else {
			a.LineTo(px, py)
		}
	}
	a.DrawPath(agg.StrokeOnly)
}

// --- Mouse handlers ---

func handleGammaCorrectionMouseDown(x, y float64) bool {
	handleGammaCorrectionMouseMove(x, y)
	return true
}

func handleGammaCorrectionMouseMove(x, y float64) bool {
	gammaRX = math.Abs(float64(width)/2 - x)
	gammaRY = math.Abs(float64(height)/2 - y)
	if gammaRX < 5 {
		gammaRX = 5
	}
	if gammaRY < 5 {
		gammaRY = 5
	}
	return true
}
