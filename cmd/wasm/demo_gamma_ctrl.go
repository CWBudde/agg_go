// Based on the original AGG examples: gamma_ctrl.cpp.
package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	"github.com/MeKo-Christian/agg_go/internal/ctrl/gamma"
)

var gammaControl *gamma.GammaCtrl

type GammaControl = gamma.GammaCtrl

func initGammaCtrlDemo() {
	if gammaControl == nil {
		// Position control in the lower-left area of the 800x600 canvas.
		// Original C++ used (10,10,300,200) with flip_y; here Y increases downward.
		gammaControl = gamma.NewGammaCtrl(10, 340, 310, 585, false)
		gammaControl.SetTextSize(10, 0)
	}
}

func drawGammaCtrlDemo() {
	initGammaCtrlDemo()

	agg2d := ctx.GetAgg2D()
	agg2d.ResetTransformations()
	agg2d.ClearAll(agg.White)

	ewidth := float64(width)/2 - 10
	ecenter := float64(width) / 2

	// Apply gamma from control to the rasterizer before drawing shapes.
	ras := agg2d.GetInternalRasterizer()
	ras.SetGamma(gammaControl.Y)

	// Six ellipse pairs ordered top-to-bottom (least visible → most visible),
	// matching the original flip_y layout from gamma_ctrl.cpp.
	drawGammaEllipse(agg2d, ecenter, 45, ewidth, 15.5, 0.1, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 45, 10.5, 10.5, 0.1, agg.NewColor(0, 0, 102, 255))

	drawGammaEllipse(agg2d, ecenter, 95, ewidth, 15.5, 0.4, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 95, 10.5, 10.5, 0.4, agg.NewColor(0, 0, 102, 255))

	drawGammaEllipse(agg2d, ecenter, 145, ewidth, 15.5, 1.0, agg.NewColor(0, 0, 102, 255))
	drawGammaEllipse(agg2d, ecenter, 145, 10.5, 10.5, 1.0, agg.NewColor(0, 0, 102, 255))

	drawGammaEllipse(agg2d, ecenter, 195, ewidth, 15, 2.0, agg.NewColor(192, 192, 192, 255))
	drawGammaEllipse(agg2d, ecenter, 195, 11, 11, 2.0, agg.NewColor(192, 192, 192, 255))

	drawGammaEllipse(agg2d, ecenter, 245, ewidth, 15, 2.0, agg.NewColor(127, 127, 127, 255))
	drawGammaEllipse(agg2d, ecenter, 245, 11, 11, 2.0, agg.NewColor(127, 127, 127, 255))

	drawGammaEllipse(agg2d, ecenter, 295, ewidth, 15, 2.0, agg.NewColor(0, 0, 0, 255))
	drawGammaEllipse(agg2d, ecenter, 295, 11, 11, 2.0, agg.NewColor(0, 0, 0, 255))

	// Render without gamma correction for the control and decorative elements.
	ras.SetGamma(func(x float64) float64 { return x })

	// Draw text in lower-right, matching original start_point(320,10) after flip_y.
	agg2d.FontGSV(50)
	agg2d.FillColor(agg.NewColor(0, 127, 0, 255))
	agg2d.Text(370, 555, "Text 2345", false, 0, 0)

	// Rotating arrows to the right of the gamma control.
	drawRotatingArrows(agg2d, 490, 415, agg.NewColor(127, 0, 0, 255))

	// Render the gamma control itself last (no gamma applied).
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
		adapter := &gammaCtrlAdapter{ctrl: ctrl}
		ras.AddPath(adapter, uint32(i)) // AddPath calls adapter.Rewind(i) → ctrl.Rewind(i)
		c := ctrl.Color(i)
		agg2d.RenderRasterizerWithColor(agg.RGBA(c.R, c.G, c.B, c.A))
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
