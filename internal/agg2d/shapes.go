// Basic shape drawing methods for AGG2D interface.
// This file contains all the basic shape drawing methods that match the C++ AGG2D interface.
package agg2d

import (
	"math"
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
	// TODO: Use internal/shapes/rounded_rect for proper implementation
	// For now, implement a simplified version using bezier curves

	agg2d.ResetPath()

	// Ensure proper ordering
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Clamp radii to half of width/height
	w := x2 - x1
	h := y2 - y1

	rxTop = math.Min(rxTop, w/2)
	rxBottom = math.Min(rxBottom, w/2)
	ryTop = math.Min(ryTop, h/2)
	ryBottom = math.Min(ryBottom, h/2)

	// Top edge
	agg2d.MoveTo(x1+rxTop, y1)
	agg2d.LineTo(x2-rxTop, y1)

	// Top-right corner
	if rxTop > 0 && ryTop > 0 {
		agg2d.addCornerArc(x2-rxTop, y1+ryTop, rxTop, ryTop, 0) // 0 degrees to 90 degrees
	}

	// Right edge
	agg2d.LineTo(x2, y2-ryBottom)

	// Bottom-right corner
	if rxBottom > 0 && ryBottom > 0 {
		agg2d.addCornerArc(x2-rxBottom, y2-ryBottom, rxBottom, ryBottom, 90) // 90 degrees to 180 degrees
	}

	// Bottom edge
	agg2d.LineTo(x1+rxBottom, y2)

	// Bottom-left corner
	if rxBottom > 0 && ryBottom > 0 {
		agg2d.addCornerArc(x1+rxBottom, y2-ryBottom, rxBottom, ryBottom, 180) // 180 degrees to 270 degrees
	}

	// Left edge
	agg2d.LineTo(x1, y1+ryTop)

	// Top-left corner
	if rxTop > 0 && ryTop > 0 {
		agg2d.addCornerArc(x1+rxTop, y1+ryTop, rxTop, ryTop, 270) // 270 degrees to 360 degrees
	}

	agg2d.ClosePolygon()
	agg2d.DrawPath(FillAndStroke)
}

// Helper function to add a corner arc
func (agg2d *Agg2D) addCornerArc(cx, cy, rx, ry float64, startAngle float64) {
	// Use bezier curve to approximate 90-degree arc
	const kappa = 0.5522847498307936 // (4/3)*tan(pi/8)

	startRad := startAngle * math.Pi / 180
	endRad := (startAngle + 90) * math.Pi / 180

	x1 := cx + rx*math.Cos(startRad)
	y1 := cy + ry*math.Sin(startRad)
	x4 := cx + rx*math.Cos(endRad)
	y4 := cy + ry*math.Sin(endRad)

	// Calculate control points
	x2 := x1 - kappa*rx*math.Sin(startRad)
	y2 := y1 + kappa*ry*math.Cos(startRad)
	x3 := x4 + kappa*rx*math.Sin(endRad)
	y3 := y4 - kappa*ry*math.Cos(endRad)

	agg2d.CubicCurveTo(x2, y2, x3, y3, x4, y4)
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
	agg2d.ResetPath()

	// Convert angles to radians
	startRad := start
	sweepRad := sweep

	// For now, approximate with line segments
	// TODO: Use proper bezier arc implementation
	segments := int(math.Max(4, math.Abs(sweepRad)*180/math.Pi/15)) // ~15 degrees per segment

	for i := 0; i <= segments; i++ {
		angle := startRad + sweepRad*float64(i)/float64(segments)
		x := cx + rx*math.Cos(angle)
		y := cy + ry*math.Sin(angle)

		if i == 0 {
			agg2d.MoveTo(x, y)
		} else {
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
