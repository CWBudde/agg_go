// Package agg2d float shape convenience methods (L5). Mirrors shapes.go; bodies
// are identical since shapes build paths and delegate to DrawPath. The shape
// builders in internal/shapes are color-agnostic and reused verbatim.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/shapes"
)

// Line strokes a single line segment.
func (a *Agg2DFloat) Line(x1, y1, x2, y2 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.LineTo(x2, y2)
	a.DrawPath(StrokeOnly)
}

// Triangle fills and strokes a triangle.
func (a *Agg2DFloat) Triangle(x1, y1, x2, y2, x3, y3 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.LineTo(x2, y2)
	a.LineTo(x3, y3)
	a.ClosePolygon()
	a.DrawPath(FillAndStroke)
}

// Rectangle fills and strokes an axis-aligned rectangle.
func (a *Agg2DFloat) Rectangle(x1, y1, x2, y2 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.LineTo(x2, y1)
	a.LineTo(x2, y2)
	a.LineTo(x1, y2)
	a.ClosePolygon()
	a.DrawPath(FillAndStroke)
}

// Ellipse fills and strokes an ellipse.
func (a *Agg2DFloat) Ellipse(cx, cy, rx, ry float64) {
	a.ResetPath()
	a.AddEllipse(cx, cy, rx, ry, CCW)
	a.DrawPath(FillAndStroke)
}

// DrawCircle strokes a circle.
func (a *Agg2DFloat) DrawCircle(cx, cy, radius float64) {
	a.ResetPath()
	a.AddEllipse(cx, cy, radius, radius, CCW)
	a.DrawPath(StrokeOnly)
}

// FillCircle fills a circle.
func (a *Agg2DFloat) FillCircle(cx, cy, radius float64) {
	a.ResetPath()
	a.AddEllipse(cx, cy, radius, radius, CCW)
	a.DrawPath(FillOnly)
}

// RoundedRect fills and strokes a rounded rectangle with a uniform corner radius.
func (a *Agg2DFloat) RoundedRect(x1, y1, x2, y2, r float64) {
	a.RoundedRectVariableRadii(x1, y1, x2, y2, r, r, r, r)
}

// RoundedRectXY fills and strokes a rounded rectangle with separate x and y radii.
func (a *Agg2DFloat) RoundedRectXY(x1, y1, x2, y2, rx, ry float64) {
	a.RoundedRectVariableRadii(x1, y1, x2, y2, rx, ry, rx, ry)
}

// RoundedRectVariableRadii fills and strokes a rounded rectangle with distinct
// bottom and top radii. The shape builder is color-agnostic and shared with the
// 8-bit path.
func (a *Agg2DFloat) RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop float64) {
	roundedRect := shapes.NewRoundedRectEmpty()
	roundedRect.SetRect(x1, y1, x2, y2)
	roundedRect.SetRadiusBottomTop(rxBottom, ryBottom, rxTop, ryTop)

	a.ResetPath()
	roundedRect.Rewind(0)

	first := true
	for {
		var x, y float64
		cmd := roundedRect.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		switch {
		case first:
			a.MoveTo(x, y)
			first = false
		case cmd == basics.PathCmdLineTo:
			a.LineTo(x, y)
		case cmd&basics.PathCmdMask == basics.PathCmdEndPoly:
			a.ClosePolygon()
		}
	}

	a.DrawPath(FillAndStroke)
}

// Arc strokes an elliptical arc described by center, radii, start, and sweep.
func (a *Agg2DFloat) Arc(cx, cy, rx, ry, start, sweep float64) {
	arc := shapes.NewArcWithParams(cx, cy, rx, ry, start, start+sweep, true)

	a.ResetPath()
	arc.Rewind(0)

	first := true
	for {
		var x, y float64
		cmd := arc.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		if first {
			a.MoveTo(x, y)
			first = false
		} else if cmd == basics.PathCmdLineTo {
			a.LineTo(x, y)
		}
	}

	a.DrawPath(StrokeOnly)
}

// Star fills and strokes a star polygon with numRays points.
func (a *Agg2DFloat) Star(cx, cy, r1, r2, startAngle float64, numRays int) {
	a.ResetPath()

	da := math.Pi / float64(numRays)
	angle := startAngle

	for i := 0; i < numRays; i++ {
		x := math.Cos(angle)*r2 + cx
		y := math.Sin(angle)*r2 + cy

		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}

		angle += da

		x = math.Cos(angle)*r1 + cx
		y = math.Sin(angle)*r1 + cy
		a.LineTo(x, y)

		angle += da
	}

	a.ClosePolygon()
	a.DrawPath(FillAndStroke)
}

// Curve strokes a quadratic Bézier convenience shape.
func (a *Agg2DFloat) Curve(x1, y1, x2, y2, x3, y3 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.QuadricCurveTo(x2, y2, x3, y3)
	a.DrawPath(StrokeOnly)
}

// Curve4 strokes a cubic Bézier convenience shape.
func (a *Agg2DFloat) Curve4(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	a.ResetPath()
	a.MoveTo(x1, y1)
	a.CubicCurveTo(x2, y2, x3, y3, x4, y4)
	a.DrawPath(StrokeOnly)
}

// Polygon fills and strokes a closed polygon from alternating x,y coordinates.
func (a *Agg2DFloat) Polygon(xy []float64, numPoints int) {
	if len(xy) < numPoints*2 {
		return
	}

	a.ResetPath()
	for i := 0; i < numPoints; i++ {
		x := xy[i*2]
		y := xy[i*2+1]
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}

	a.ClosePolygon()
	a.DrawPath(FillAndStroke)
}

// Polyline strokes an open polyline from alternating x,y coordinates.
func (a *Agg2DFloat) Polyline(xy []float64, numPoints int) {
	if len(xy) < numPoints*2 {
		return
	}

	a.ResetPath()
	for i := 0; i < numPoints; i++ {
		x := xy[i*2]
		y := xy[i*2+1]
		if i == 0 {
			a.MoveTo(x, y)
		} else {
			a.LineTo(x, y)
		}
	}

	a.DrawPath(StrokeOnly)
}
