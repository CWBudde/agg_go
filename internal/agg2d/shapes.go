// Basic shape drawing methods for AGG2D interface.
// This file contains all the basic shape drawing methods that match the C++ AGG2D interface.
package agg2d

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/shapes"
)

// Line draws a straight line between two points.
// This matches the C++ Agg2D::line method.
func (agg2d *Agg2D) Line(x1, y1, x2, y2 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.LineTo(x2, y2)
	agg2d.DrawPath(StrokeOnly)
}

// Triangle draws a triangle with three vertices.
// This matches the C++ Agg2D::triangle method.
func (agg2d *Agg2D) Triangle(x1, y1, x2, y2, x3, y3 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.LineTo(x2, y2)
	agg2d.LineTo(x3, y3)
	agg2d.ClosePolygon()
	agg2d.DrawPath(FillAndStroke)
}

// Rectangle draws a rectangle.
// This matches the C++ Agg2D::rectangle method.
func (agg2d *Agg2D) Rectangle(x1, y1, x2, y2 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.LineTo(x2, y1)
	agg2d.LineTo(x2, y2)
	agg2d.LineTo(x1, y2)
	agg2d.ClosePolygon()
	agg2d.DrawPath(FillAndStroke)
}

// RoundedRect draws a rounded rectangle with uniform corner radius.
// This matches the C++ Agg2D::roundedRect(x1, y1, x2, y2, r) method.
func (agg2d *Agg2D) RoundedRect(x1, y1, x2, y2, r float64) {
	agg2d.RoundedRectVariableRadii(x1, y1, x2, y2, r, r, r, r)
}

// RoundedRectXY draws a rounded rectangle with separate X and Y corner radii.
// This matches the C++ Agg2D::roundedRect(x1, y1, x2, y2, rx, ry) method.
func (agg2d *Agg2D) RoundedRectXY(x1, y1, x2, y2, rx, ry float64) {
	agg2d.RoundedRectVariableRadii(x1, y1, x2, y2, rx, ry, rx, ry)
}

// RoundedRectVariableRadii draws a rounded rectangle with different radii for each corner.
// This matches the C++ Agg2D::roundedRect(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop) method.
func (agg2d *Agg2D) RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop float64) {
	// Use proper rounded rectangle implementation from internal/shapes
	roundedRect := shapes.NewRoundedRectEmpty()
	roundedRect.SetRect(x1, y1, x2, y2)
	roundedRect.SetRadiusBottomTop(rxBottom, ryBottom, rxTop, ryTop)

	agg2d.ResetPath()

	// Rewind the rounded rectangle to start generating vertices
	roundedRect.Rewind(0)

	// Generate vertices and add to path
	first := true
	for {
		var x, y float64
		cmd := roundedRect.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		if first {
			agg2d.MoveTo(x, y)
			first = false
		} else if cmd == basics.PathCmdLineTo {
			agg2d.LineTo(x, y)
		} else if cmd&basics.PathCmdMask == basics.PathCmdEndPoly {
			agg2d.ClosePolygon()
		}
	}

	agg2d.DrawPath(FillAndStroke)
}

// Ellipse draws an ellipse.
// This matches the C++ Agg2D::ellipse method.
func (agg2d *Agg2D) Ellipse(cx, cy, rx, ry float64) {
	agg2d.ResetPath()
	agg2d.AddEllipse(cx, cy, rx, ry, CCW)
	agg2d.DrawPath(FillAndStroke)
}

// DrawCircle draws a circle with a stroke.
func (agg2d *Agg2D) DrawCircle(cx, cy, radius float64) {
	agg2d.ResetPath()
	agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	agg2d.DrawPath(StrokeOnly)
}

// FillCircle draws a filled circle.
func (agg2d *Agg2D) FillCircle(cx, cy, radius float64) {
	agg2d.ResetPath()
	agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	agg2d.DrawPath(FillOnly)
}

// Arc draws an elliptical arc.
// This matches the C++ Agg2D::arc method.
func (agg2d *Agg2D) Arc(cx, cy, rx, ry, start, sweep float64) {
	// Use proper arc implementation from internal/shapes
	arc := shapes.NewArcWithParams(cx, cy, rx, ry, start, start+sweep, true) // ccw=true for positive sweep

	agg2d.ResetPath()

	// Rewind the arc to start generating vertices
	arc.Rewind(0)

	// Generate vertices and add to path
	first := true
	for {
		var x, y float64
		cmd := arc.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		if first {
			agg2d.MoveTo(x, y)
			first = false
		} else if cmd == basics.PathCmdLineTo {
			agg2d.LineTo(x, y)
		}
	}

	agg2d.DrawPath(StrokeOnly)
}

// Star draws a star shape.
// This matches the C++ Agg2D::star method.
func (agg2d *Agg2D) Star(cx, cy, r1, r2, startAngle float64, numRays int) {
	agg2d.ResetPath()

	da := math.Pi / float64(numRays)
	a := startAngle

	for i := 0; i < numRays; i++ {
		// Outer point
		x := math.Cos(a)*r2 + cx
		y := math.Sin(a)*r2 + cy

		if i == 0 {
			agg2d.MoveTo(x, y)
		} else {
			agg2d.LineTo(x, y)
		}

		a += da

		// Inner point
		x = math.Cos(a)*r1 + cx
		y = math.Sin(a)*r1 + cy
		agg2d.LineTo(x, y)

		a += da
	}

	agg2d.ClosePolygon()
	agg2d.DrawPath(FillAndStroke)
}

// Curve draws a quadratic Bézier curve.
// This matches the C++ Agg2D::curve(x1, y1, x2, y2, x3, y3) method.
func (agg2d *Agg2D) Curve(x1, y1, x2, y2, x3, y3 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.QuadricCurveTo(x2, y2, x3, y3)
	agg2d.DrawPath(StrokeOnly)
}

// Curve4 draws a cubic Bézier curve.
// This matches the C++ Agg2D::curve(x1, y1, x2, y2, x3, y3, x4, y4) method.
func (agg2d *Agg2D) Curve4(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	agg2d.ResetPath()
	agg2d.MoveTo(x1, y1)
	agg2d.CubicCurveTo(x2, y2, x3, y3, x4, y4)
	agg2d.DrawPath(StrokeOnly)
}

// Polygon draws a polygon from an array of points.
// This matches the C++ Agg2D::polygon method.
func (agg2d *Agg2D) Polygon(xy []float64, numPoints int) {
	if len(xy) < numPoints*2 {
		return // Not enough points
	}

	agg2d.ResetPath()

	for i := 0; i < numPoints; i++ {
		x := xy[i*2]
		y := xy[i*2+1]

		if i == 0 {
			agg2d.MoveTo(x, y)
		} else {
			agg2d.LineTo(x, y)
		}
	}

	agg2d.ClosePolygon()
	agg2d.DrawPath(FillAndStroke)
}

// Polyline draws a polyline from an array of points.
// This matches the C++ Agg2D::polyline method.
func (agg2d *Agg2D) Polyline(xy []float64, numPoints int) {
	if len(xy) < numPoints*2 {
		return // Not enough points
	}

	agg2d.ResetPath()

	for i := 0; i < numPoints; i++ {
		x := xy[i*2]
		y := xy[i*2+1]

		if i == 0 {
			agg2d.MoveTo(x, y)
		} else {
			agg2d.LineTo(x, y)
		}
	}

	agg2d.DrawPath(StrokeOnly)
}
