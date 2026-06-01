// Package agg2d float path building (L5). Color-agnostic path methods mirrored
// onto Agg2DFloat; bodies are identical to paths.go since they only manipulate
// the shared path/transform/converter state.
package agg2d

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/shapes"
)

// ResetPath clears all path data.
func (a *Agg2DFloat) ResetPath() {
	a.path.RemoveAll()
	a.hasLastCtrl = false
}

// MoveTo moves the current point to absolute coordinates.
func (a *Agg2DFloat) MoveTo(x, y float64) {
	a.path.MoveTo(x, y)
	a.hasLastCtrl = false
}

// MoveRel moves the current point by a relative amount.
func (a *Agg2DFloat) MoveRel(dx, dy float64) {
	a.path.MoveRel(dx, dy)
	a.hasLastCtrl = false
}

// LineTo draws a line to absolute coordinates.
func (a *Agg2DFloat) LineTo(x, y float64) {
	a.path.LineTo(x, y)
	a.hasLastCtrl = false
}

// LineRel draws a line by a relative amount.
func (a *Agg2DFloat) LineRel(dx, dy float64) {
	a.path.LineRel(dx, dy)
	a.hasLastCtrl = false
}

// HorLineTo draws a horizontal line to the given x.
func (a *Agg2DFloat) HorLineTo(x float64) { a.path.HLineTo(x) }

// VerLineTo draws a vertical line to the given y.
func (a *Agg2DFloat) VerLineTo(y float64) { a.path.VLineTo(y) }

// ArcTo adds an SVG-style elliptical arc to the path.
func (a *Agg2DFloat) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	a.path.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, x, y)
}

// ArcRel adds an elliptical arc to the path using relative coordinates.
func (a *Agg2DFloat) ArcRel(rx, ry, angle float64, largeArcFlag, sweepFlag bool, dx, dy float64) {
	a.path.ArcRel(rx, ry, angle, largeArcFlag, sweepFlag, dx, dy)
}

// QuadricCurveTo adds a quadratic Bézier curve to the path.
func (a *Agg2DFloat) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	a.path.Curve3(xCtrl, yCtrl, xTo, yTo)
	a.lastCtrlX = xCtrl
	a.lastCtrlY = yCtrl
	a.hasLastCtrl = true
}

// CubicCurveTo adds a cubic Bézier curve to the path.
func (a *Agg2DFloat) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	a.path.Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
	a.lastCtrlX = xCtrl2
	a.lastCtrlY = yCtrl2
	a.hasLastCtrl = true
}

// AddEllipse appends an ellipse contour to the current path.
func (a *Agg2DFloat) AddEllipse(cx, cy, rx, ry float64, dir Direction) {
	ellipse := shapes.NewEllipseWithParams(cx, cy, rx, ry, 0, dir == CW)
	ellipse.Rewind(0)
	first := true
	for {
		var x, y float64
		cmd := ellipse.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}
		if first {
			a.path.MoveTo(x, y)
			first = false
		} else if cmd == basics.PathCmdLineTo {
			a.path.LineTo(x, y)
		}
	}
	a.path.ClosePolygon(basics.PathFlagsNone)
}

// ClosePolygon closes the current sub-path.
func (a *Agg2DFloat) ClosePolygon() {
	a.path.ClosePolygon(basics.PathFlagsNone)
}

// DrawPath renders the current path according to the given flag.
func (a *Agg2DFloat) DrawPath(flag DrawPathFlag) {
	a.updateApproximationScales()
	switch flag {
	case FillOnly:
		a.renderFill()
	case StrokeOnly:
		a.renderStroke()
	case FillAndStroke:
		a.renderFill()
		a.renderStroke()
	case FillWithLineColor:
		a.renderFillWithLineColor()
	}
}

// DrawPathNoTransform renders the current path with the identity transform.
func (a *Agg2DFloat) DrawPathNoTransform(flag DrawPathFlag) {
	oldTransform := *a.transform
	a.transform.Reset()
	a.DrawPath(flag)
	*a.transform = oldTransform
}
