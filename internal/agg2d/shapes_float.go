// Package agg2d float shape convenience methods (L5). Mirrors shapes.go; bodies
// are identical since shapes build paths and delegate to DrawPath.
package agg2d

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
