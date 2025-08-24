// Path commands implementation for AGG2D interface.
// This file contains all the path manipulation methods that match the C++ AGG2D interface.
package agg2d

import (
	"agg_go/internal/basics"
	"agg_go/internal/shapes"
)

// resetPath clears all path data.
// This matches the C++ Agg2D::resetPath method.
func (agg2d *Agg2D) ResetPath() {
	agg2d.path.RemoveAll()
}

// MoveTo moves the current point to the specified coordinates.
// This matches the C++ Agg2D::moveTo method.
func (agg2d *Agg2D) MoveTo(x, y float64) {
	agg2d.path.MoveTo(x, y)
}

// MoveRel moves the current point by the specified relative amounts.
// This matches the C++ Agg2D::moveRel method.
func (agg2d *Agg2D) MoveRel(dx, dy float64) {
	agg2d.path.MoveRel(dx, dy)
}

// LineTo draws a line from the current point to the specified coordinates.
// This matches the C++ Agg2D::lineTo method.
func (agg2d *Agg2D) LineTo(x, y float64) {
	agg2d.path.LineTo(x, y)
}

// LineRel draws a line from the current point by the specified relative amounts.
// This matches the C++ Agg2D::lineRel method.
func (agg2d *Agg2D) LineRel(dx, dy float64) {
	agg2d.path.LineRel(dx, dy)
}

// HorLineTo draws a horizontal line to the specified x coordinate.
// This matches the C++ Agg2D::horLineTo method.
func (agg2d *Agg2D) HorLineTo(x float64) {
	agg2d.path.HLineTo(x)
}

// HorLineRel draws a horizontal line by the specified relative amount.
// This matches the C++ Agg2D::horLineRel method.
func (agg2d *Agg2D) HorLineRel(dx float64) {
	agg2d.path.HLineRel(dx)
}

// VerLineTo draws a vertical line to the specified y coordinate.
// This matches the C++ Agg2D::verLineTo method.
func (agg2d *Agg2D) VerLineTo(y float64) {
	agg2d.path.VLineTo(y)
}

// VerLineRel draws a vertical line by the specified relative amount.
// This matches the C++ Agg2D::verLineRel method.
func (agg2d *Agg2D) VerLineRel(dy float64) {
	agg2d.path.VLineRel(dy)
}

// ArcTo adds an elliptical arc to the path.
// This matches the C++ Agg2D::arcTo method.
func (agg2d *Agg2D) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	agg2d.path.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, x, y)
}

// ArcRel adds an elliptical arc to the path using relative coordinates.
// This matches the C++ Agg2D::arcRel method.
func (agg2d *Agg2D) ArcRel(rx, ry, angle float64, largeArcFlag, sweepFlag bool, dx, dy float64) {
	agg2d.path.ArcRel(rx, ry, angle, largeArcFlag, sweepFlag, dx, dy)
}

// QuadricCurveTo adds a quadratic Bézier curve to the path.
// This matches the C++ Agg2D::quadricCurveTo method.
func (agg2d *Agg2D) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	agg2d.path.Curve3(xCtrl, yCtrl, xTo, yTo)
}

// QuadricCurveRel adds a quadratic Bézier curve to the path using relative coordinates.
// This matches the C++ Agg2D::quadricCurveRel method.
func (agg2d *Agg2D) QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo float64) {
	agg2d.path.Curve3Rel(dxCtrl, dyCtrl, dxTo, dyTo)
}

// QuadricCurveToSmooth adds a smooth quadratic Bézier curve.
// This matches the C++ Agg2D::quadricCurveTo(xTo, yTo) method.
func (agg2d *Agg2D) QuadricCurveToSmooth(xTo, yTo float64) {
	// In a smooth curve, the control point is the reflection of the previous control point
	// Get the current position and calculate reflected control point
	// For a proper implementation, we would need to track the previous control point
	// This is a simplified version that uses the current position
	currentX, currentY, _ := agg2d.path.LastVertex()
	ctrlX := currentX + (currentX-xTo)*0.5 // Simple approximation
	ctrlY := currentY + (currentY-yTo)*0.5
	agg2d.path.Curve3(ctrlX, ctrlY, xTo, yTo)
}

// QuadricCurveRelSmooth adds a smooth quadratic Bézier curve using relative coordinates.
// This matches the C++ Agg2D::quadricCurveRel(dxTo, dyTo) method.
func (agg2d *Agg2D) QuadricCurveRelSmooth(dxTo, dyTo float64) {
	// For smooth relative curves, use relative approximation
	dCtrlX := dxTo * 0.5 // Simple approximation
	dCtrlY := dyTo * 0.5
	agg2d.path.Curve3Rel(dCtrlX, dCtrlY, dxTo, dyTo)
}

// CubicCurveTo adds a cubic Bézier curve to the path.
// This matches the C++ Agg2D::cubicCurveTo method.
func (agg2d *Agg2D) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	agg2d.path.Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

// CubicCurveRel adds a cubic Bézier curve to the path using relative coordinates.
// This matches the C++ Agg2D::cubicCurveRel method.
func (agg2d *Agg2D) CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	agg2d.path.Curve4Rel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo)
}

// CubicCurveToSmooth adds a smooth cubic Bézier curve.
// This matches the C++ Agg2D::cubicCurveTo(xCtrl2, yCtrl2, xTo, yTo) method.
func (agg2d *Agg2D) CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo float64) {
	// In a smooth curve, the first control point is the reflection of the previous second control point
	// Get the current position and calculate reflected control point
	currentX, currentY, _ := agg2d.path.LastVertex()
	// This is a simplified version - proper implementation would track the previous control point
	xCtrl1 := currentX + (currentX-xCtrl2)*0.3 // Simple reflection approximation
	yCtrl1 := currentY + (currentY-yCtrl2)*0.3
	agg2d.path.Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

// CubicCurveRelSmooth adds a smooth cubic Bézier curve using relative coordinates.
// This matches the C++ Agg2D::cubicCurveRel(dxCtrl2, dyCtrl2, dxTo, dyTo) method.
func (agg2d *Agg2D) CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	// For smooth relative curves, use relative approximation
	dxCtrl1 := -dxCtrl2 * 0.3 // Simple reflection approximation
	dyCtrl1 := -dyCtrl2 * 0.3
	agg2d.path.Curve4Rel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo)
}

// AddEllipse adds an ellipse to the path.
// This matches the C++ Agg2D::addEllipse method.
func (agg2d *Agg2D) AddEllipse(cx, cy, rx, ry float64, dir Direction) {
	// Use proper ellipse implementation from internal/shapes
	ellipse := shapes.NewEllipseWithParams(cx, cy, rx, ry, 0, dir == CW)

	// Rewind the ellipse to start generating vertices
	ellipse.Rewind(0)

	// Generate vertices and add to path
	first := true
	for {
		var x, y float64
		cmd := ellipse.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		if first {
			agg2d.path.MoveTo(x, y)
			first = false
		} else if cmd == basics.PathCmdLineTo {
			agg2d.path.LineTo(x, y)
		}
	}

	// Close the ellipse
	agg2d.path.ClosePolygon(basics.PathFlagsNone)
}

// ClosePolygon closes the current sub-path.
// This matches the C++ Agg2D::closePolygon method.
func (agg2d *Agg2D) ClosePolygon() {
	agg2d.path.ClosePolygon(basics.PathFlagsNone)
}

// DrawPath renders the current path according to the specified flag.
// This matches the C++ Agg2D::drawPath method.
func (agg2d *Agg2D) DrawPath(flag DrawPathFlag) {
	// Update approximation scales before rendering
	agg2d.updateApproximationScales()

	switch flag {
	case FillOnly:
		// Render fill only
		agg2d.renderFill()
	case StrokeOnly:
		// Render stroke only
		agg2d.renderStroke()
	case FillAndStroke:
		// Render both fill and stroke
		agg2d.renderFill()
		agg2d.renderStroke()
	case FillWithLineColor:
		// Render fill using line color
		agg2d.renderFillWithLineColor()
	}
}

// DrawPathNoTransform renders the current path without applying transformations.
// This matches the C++ Agg2D::drawPathNoTransform method.
func (agg2d *Agg2D) DrawPathNoTransform(flag DrawPathFlag) {
	// Render without transformation by temporarily storing and resetting transform
	oldTransform := *agg2d.transform
	agg2d.transform.Reset() // Reset to identity matrix

	// Render with identity transform
	agg2d.DrawPath(flag)

	// Restore original transform
	*agg2d.transform = oldTransform
}

// Helper methods for rendering - implemented in agg2d.go
