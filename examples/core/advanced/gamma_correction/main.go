// Port of AGG C++ gamma_correction.cpp – anti-aliasing gamma demonstration.
//
// Shows how the anti-aliasing gamma affects thin ellipse rendering on a split
// dark/light background. Multiple concentric ellipses at different stroke
// widths are drawn, matching the C++ original's rendering.
// The gamma/thickness/contrast values are fixed defaults; in the interactive
// version they are adjustable via sliders.
package main

import (
	"math"

	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/examples/shared/lowlevelrunner"
)

const (
	width  = 400
	height = 320

	defaultThickness = 1.0
	defaultContrast  = 1.0
	defaultGamma     = 1.0
)

func drawEllipseStroked(a *agg.Agg2D, cx, cy, rx, ry, lw float64, c agg.Color) {
	a.LineColor(c)
	a.NoFill()
	a.LineWidth(lw)
	a.ResetPath()
	a.AddEllipse(cx, cy, rx, ry, agg.CCW)
	a.DrawPath(agg.StrokeOnly)
}

type demo struct {
	rx, ry float64
}

func (d *demo) Render(img *agg.Image) {
	ctx := agg.NewContextForImage(img)
	a := ctx.GetAgg2D()
	a.ResetTransformations()

	w := float64(width)
	h := float64(height)
	cx := w / 2
	cy := h / 2

	dark := 1.0 - defaultContrast
	light := defaultContrast
	f := func(v float64) uint8 { return uint8(v*255 + 0.5) }

	// Background: split dark/light with red bottom half.
	// Top-left quadrant: dark
	a.FillColor(agg.NewColor(f(dark), f(dark), f(dark), 255))
	a.NoLine()
	a.ResetPath()
	a.MoveTo(0, 0)
	a.LineTo(cx, 0)
	a.LineTo(cx, cy)
	a.LineTo(0, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Top-right quadrant: light
	a.FillColor(agg.NewColor(f(light), f(light), f(light), 255))
	a.ResetPath()
	a.MoveTo(cx, 0)
	a.LineTo(w, 0)
	a.LineTo(w, cy)
	a.LineTo(cx, cy)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Bottom half: reddish dark
	a.FillColor(agg.NewColor(255, f(dark), f(dark), 255))
	a.ResetPath()
	a.MoveTo(0, cy)
	a.LineTo(w, cy)
	a.LineTo(w, h)
	a.LineTo(0, h)
	a.ClosePolygon()
	a.DrawPath(agg.FillOnly)

	// Apply gamma to the rasterizer.
	a.AntiAliasGamma(defaultGamma)

	// Gamma power curve visualization.
	a.LineColor(agg.NewColor(80, 127, 80, 255))
	a.NoFill()
	a.LineWidth(2.0)
	a.ResetPath()
	xStart := (w - 256.0) / 2.0
	yStart := 50.0
	for i := 0; i <= 255; i++ {
		v := float64(i) / 255.0
		gval := math.Pow(v, defaultGamma)
		px := xStart + float64(i)
		py := yStart + gval*255.0
		if i == 0 {
			a.MoveTo(px, py)
		} else {
			a.LineTo(px, py)
		}
	}
	a.DrawPath(agg.StrokeOnly)

	// Concentric ellipses at the same thickness, matching C++ original.
	// Red ellipse at full rx, ry
	drawEllipseStroked(a, cx, cy, d.rx, d.ry, defaultThickness, agg.NewColor(255, 0, 0, 255))

	// Green ellipse at rx-5, ry-5
	drawEllipseStroked(a, cx, cy, d.rx-5.0, d.ry-5.0, defaultThickness, agg.NewColor(0, 255, 0, 255))

	// Blue ellipse at rx-10, ry-10
	drawEllipseStroked(a, cx, cy, d.rx-10.0, d.ry-10.0, defaultThickness, agg.NewColor(0, 0, 255, 255))

	// Black ellipse at rx-15, ry-15
	drawEllipseStroked(a, cx, cy, d.rx-15.0, d.ry-15.0, defaultThickness, agg.NewColor(0, 0, 0, 255))

	// White ellipse at rx-20, ry-20
	drawEllipseStroked(a, cx, cy, d.rx-20.0, d.ry-20.0, defaultThickness, agg.NewColor(255, 255, 255, 255))
}

func (d *demo) OnMouseDown(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.rx = math.Abs(float64(width)/2 - float64(x))
		d.ry = math.Abs(float64(height)/2 - float64(y))
		return true
	}
	return false
}

func (d *demo) OnMouseMove(x, y int, btn lowlevelrunner.Buttons) bool {
	if btn.Left {
		d.rx = math.Abs(float64(width)/2 - float64(x))
		d.ry = math.Abs(float64(height)/2 - float64(y))
		return true
	}
	return false
}

func (d *demo) OnMouseUp(_, _ int, _ lowlevelrunner.Buttons) bool { return false }

func main() {
	lowlevelrunner.Run(lowlevelrunner.Config{
		Title:  "Gamma Correction",
		Width:  width,
		Height: height,
	}, &demo{
		rx: float64(width) / 3.0,
		ry: float64(height) / 3.0,
	})
}
