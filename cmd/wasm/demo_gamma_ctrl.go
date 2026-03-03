// Based on the original AGG examples: gamma_ctrl.cpp.
package main

import (
	agg "agg_go"
	"agg_go/internal/ctrl/gamma"
)

var gammaControl *gamma.GammaCtrl

type GammaControl = gamma.GammaCtrl

func initGammaCtrlDemo() {
	if gammaControl == nil {
		gammaControl = gamma.NewGammaCtrl(10, 10, 300, 200, false)
	}
}

func drawGammaCtrlDemo() {
	initGammaCtrlDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	ewidth := float64(width)/2 - 10
	ecenter := float64(width) / 2

	// Render shapes using gamma from control
	ras := agg2d.GetInternalRasterizer()
	ras.Reset()
	ras.SetGamma(gammaControl.Y)

	// Draw ellipses with different colors and line widths
	drawGammaEllipse(agg2d, ecenter, 220, ewidth, 15, 2.0, agg.NewColor(0, 0, 0, 255))
	drawGammaEllipse(agg2d, ecenter, 220, 11, 11, 2.0, agg.NewColor(0, 0, 0, 255))

	drawGammaEllipse(agg2d, ecenter, 260, ewidth, 15, 2.0, agg.NewColor(127, 127, 127, 255))
	drawGammaEllipse(agg2d, ecenter, 260, 11, 11, 2.0, agg.NewColor(127, 127, 127, 255))

	drawGammaEllipse(agg2d, ecenter, 300, ewidth, 15, 2.0, agg.NewColor(192, 192, 192, 255))
	drawGammaEllipse(agg2d, ecenter, 300, 11, 11, 2.0, agg.NewColor(192, 192, 192, 255))

	drawGammaEllipse(agg2d, ecenter, 340, ewidth, 15.5, 1.0, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 340, 10.5, 10.5, 1.0, agg.NewColor(0, 0, 102, 255))

	drawGammaEllipse(agg2d, ecenter, 380, ewidth, 15.5, 0.4, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 380, 10.5, 10.5, 0.4, agg.NewColor(0, 0, 102, 255))

	drawGammaEllipse(agg2d, ecenter, 420, ewidth, 15.5, 0.1, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 420, 10.5, 10.5, 0.1, agg.NewColor(0, 0, 102, 255))

	// Draw some text
	agg2d.FontGSV(50)
	agg2d.FillColor(agg.NewColor(0, 127, 0, 255))
	agg2d.Text(320, 10, "Text 2345", false, 0, 0)

	// Draw rotating arrows
	drawRotatingArrows(agg2d, 400, 130, agg.NewColor(127, 0, 0, 255))

	// Finally, render the control itself without gamma correction (or with identity)
	ras.SetGamma(func(x float64) float64 { return x })
	renderControl(agg2d, gammaControl)
}

func drawGammaEllipse(agg2d *agg.Agg2D, cx, cy, rx, ry, strokeWidth float64, c agg.Color) {
	agg2d.NoFill()
	agg2d.LineColor(c)
	agg2d.LineWidth(strokeWidth)
	agg2d.Ellipse(cx, cy, rx, ry)
}

func drawRotatingArrows(agg2d *agg.Agg2D, cx, cy float64, c agg.Color) {
	agg2d.FillColor(c)
	agg2d.NoLine()

	for i := 0; i < 35; i++ {
		agg2d.PushTransform()
		agg2d.Translate(-cx, -cy)
		agg2d.Rotate(float64(i) / 35.0 * 2.0 * agg.Pi)
		agg2d.Translate(cx, cy)

		// Create arrow path
		agg2d.ResetPath()
		agg2d.MoveTo(cx+30, cy-1.0)
		agg2d.LineTo(cx+60, cy+0.0)
		agg2d.LineTo(cx+30, cy+1.0)
		agg2d.MoveTo(cx+27, cy-1.0)
		agg2d.LineTo(cx+10, cy+0.0)
		agg2d.LineTo(cx+27, cy+1.0)
		agg2d.DrawPath(agg.FillOnly)

		agg2d.PopTransform()
	}
}

func renderControl(agg2d *agg.Agg2D, ctrl *gamma.GammaCtrl) {
	ras := agg2d.GetInternalRasterizer()
	numPaths := ctrl.NumPaths()

	for i := uint(0); i < numPaths; i++ {
		ras.Reset()
		ctrl.Rewind(i)
		adapter := &gammaCtrlAdapter{ctrl: ctrl}
		ras.AddPath(adapter, 0)
		c := ctrl.Color(i)
		agg2d.FillColor(agg.RGBA(c.R, c.G, c.B, c.A))
		agg2d.DrawPath(agg.FillOnly)
	}
}

type gammaCtrlAdapter struct {
	ctrl *gamma.GammaCtrl
}

func (a *gammaCtrlAdapter) Rewind(pathID uint32) {
	a.ctrl.Rewind(uint(pathID))
}

func (a *gammaCtrlAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.ctrl.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

func handleGammaCtrlMouseDown(x, y float64) bool {
	if gammaControl == nil {
		return false
	}
	return gammaControl.OnMouseButtonDown(x, y)
}

func handleGammaCtrlMouseMove(x, y float64) bool {
	if gammaControl == nil {
		return false
	}
	return gammaControl.OnMouseMove(x, y, true)
}

func handleGammaCtrlMouseUp() {
	if gammaControl != nil {
		gammaControl.OnMouseButtonUp(0, 0)
	}
}
