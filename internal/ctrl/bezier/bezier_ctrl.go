// Package bezier provides bezier curve control implementation for AGG.
// This is a port of AGG's bezier_ctrl_impl and bezier_ctrl classes.
package bezier

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl"
	"agg_go/internal/ctrl/polygon"
	"agg_go/internal/curves"
	"agg_go/internal/shapes"
)

// BezierCtrl implements an interactive cubic Bezier curve control.
// This corresponds to AGG's bezier_ctrl_impl class.
type BezierCtrl struct {
	*ctrl.BaseCtrl

	// Bezier curve and rendering components
	curve   *curves.Curve4
	ellipse *shapes.Ellipse
	stroke  *conv.ConvStroke
	poly    *polygon.PolygonCtrl

	// Rendering state
	idx uint

	// Color for rendering
	lineColor color.RGBA
}

// NewBezierCtrl creates a new cubic Bezier curve control.
// The control starts with default points that form a visible curve.
func NewBezierCtrl() *BezierCtrl {
	// Create a 4-point polygon for the Bezier control points
	poly := polygon.NewPolygonCtrl(4, 5.0)
	poly.SetInPolygonCheck(false)

	// Set default control points that create a nice visible curve
	poly.SetXn(0, 100.0) // P0 - start point
	poly.SetYn(0, 0.0)
	poly.SetXn(1, 100.0) // P1 - first control point
	poly.SetYn(1, 50.0)
	poly.SetXn(2, 50.0) // P2 - second control point
	poly.SetYn(2, 100.0)
	poly.SetXn(3, 0.0) // P3 - end point
	poly.SetYn(3, 100.0)

	curve := curves.NewCurve4()
	ellipse := shapes.NewEllipse()
	stroke := conv.NewConvStroke(curve)

	ctrl := &BezierCtrl{
		BaseCtrl:  ctrl.NewBaseCtrl(0, 0, 1, 1, false),
		curve:     curve,
		ellipse:   ellipse,
		stroke:    stroke,
		poly:      poly,
		idx:       0,
		lineColor: color.NewRGBA(0.0, 0.0, 0.0, 1.0),
	}

	// Initialize the curve with the default points
	ctrl.updateCurve()

	return ctrl
}

// Curve Management Methods

// SetCurve sets the Bezier curve control points.
func (b *BezierCtrl) SetCurve(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	b.poly.SetXn(0, x1)
	b.poly.SetYn(0, y1)
	b.poly.SetXn(1, x2)
	b.poly.SetYn(1, y2)
	b.poly.SetXn(2, x3)
	b.poly.SetYn(2, y3)
	b.poly.SetXn(3, x4)
	b.poly.SetYn(3, y4)
	b.updateCurve()
}

// Curve returns the internal curve object after updating it with current points.
func (b *BezierCtrl) Curve() *curves.Curve4 {
	b.updateCurve()
	return b.curve
}

// updateCurve updates the internal curve with current control points.
func (b *BezierCtrl) updateCurve() {
	b.curve.Init(
		b.poly.Xn(0), b.poly.Yn(0), // P0
		b.poly.Xn(1), b.poly.Yn(1), // P1
		b.poly.Xn(2), b.poly.Yn(2), // P2
		b.poly.Xn(3), b.poly.Yn(3), // P3
	)
}

// Control Point Access Methods

// X1 returns the X coordinate of the start point (P0).
func (b *BezierCtrl) X1() float64 { return b.poly.Xn(0) }

// Y1 returns the Y coordinate of the start point (P0).
func (b *BezierCtrl) Y1() float64 { return b.poly.Yn(0) }

// X2 returns the X coordinate of the first control point (P1).
func (b *BezierCtrl) X2() float64 { return b.poly.Xn(1) }

// Y2 returns the Y coordinate of the first control point (P1).
func (b *BezierCtrl) Y2() float64 { return b.poly.Yn(1) }

// X3 returns the X coordinate of the second control point (P2).
func (b *BezierCtrl) X3() float64 { return b.poly.Xn(2) }

// Y3 returns the Y coordinate of the second control point (P2).
func (b *BezierCtrl) Y3() float64 { return b.poly.Yn(2) }

// X4 returns the X coordinate of the end point (P3).
func (b *BezierCtrl) X4() float64 { return b.poly.Xn(3) }

// Y4 returns the Y coordinate of the end point (P3).
func (b *BezierCtrl) Y4() float64 { return b.poly.Yn(3) }

// SetX1 sets the X coordinate of the start point (P0).
func (b *BezierCtrl) SetX1(x float64) { b.poly.SetXn(0, x) }

// SetY1 sets the Y coordinate of the start point (P0).
func (b *BezierCtrl) SetY1(y float64) { b.poly.SetYn(0, y) }

// SetX2 sets the X coordinate of the first control point (P1).
func (b *BezierCtrl) SetX2(x float64) { b.poly.SetXn(1, x) }

// SetY2 sets the Y coordinate of the first control point (P1).
func (b *BezierCtrl) SetY2(y float64) { b.poly.SetYn(1, y) }

// SetX3 sets the X coordinate of the second control point (P2).
func (b *BezierCtrl) SetX3(x float64) { b.poly.SetXn(2, x) }

// SetY3 sets the Y coordinate of the second control point (P2).
func (b *BezierCtrl) SetY3(y float64) { b.poly.SetYn(2, y) }

// SetX4 sets the X coordinate of the end point (P3).
func (b *BezierCtrl) SetX4(x float64) { b.poly.SetXn(3, x) }

// SetY4 sets the Y coordinate of the end point (P3).
func (b *BezierCtrl) SetY4(y float64) { b.poly.SetYn(3, y) }

// Stroke Configuration Methods

// SetLineWidth sets the width of the curve and control line strokes.
func (b *BezierCtrl) SetLineWidth(w float64) {
	b.stroke.SetWidth(w)
}

// LineWidth returns the current stroke width.
func (b *BezierCtrl) LineWidth() float64 {
	return b.stroke.Width()
}

// SetPointRadius sets the radius for control point rendering and hit testing.
func (b *BezierCtrl) SetPointRadius(r float64) {
	b.poly.SetPointRadius(r)
}

// PointRadius returns the current point radius.
func (b *BezierCtrl) PointRadius() float64 {
	return b.poly.PointRadius()
}

// Color Management

// SetLineColor sets the line color for rendering.
func (b *BezierCtrl) SetLineColor(c color.RGBA) {
	b.lineColor = c
	b.poly.SetLineColor(c)
}

// LineColor returns the current line color.
func (b *BezierCtrl) LineColor() color.RGBA {
	return b.lineColor
}

// Mouse Interaction Methods (delegate to polygon control)

// OnMouseButtonDown handles mouse button press events.
func (b *BezierCtrl) OnMouseButtonDown(x, y float64) bool {
	return b.poly.OnMouseButtonDown(x, y)
}

// OnMouseButtonUp handles mouse button release events.
func (b *BezierCtrl) OnMouseButtonUp(x, y float64) bool {
	return b.poly.OnMouseButtonUp(x, y)
}

// OnMouseMove handles mouse move events.
func (b *BezierCtrl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	return b.poly.OnMouseMove(x, y, buttonPressed)
}

// OnArrowKeys handles arrow key events.
func (b *BezierCtrl) OnArrowKeys(left, right, down, up bool) bool {
	return b.poly.OnArrowKeys(left, right, down, up)
}

// Vertex Source Interface

// NumPaths returns the number of rendering paths.
// Returns 7: control line 1, control line 2, curve, point 1, point 2, point 3, point 4
func (b *BezierCtrl) NumPaths() uint {
	return 7
}

// Rewind resets the vertex generation state.
func (b *BezierCtrl) Rewind(pathID uint) {
	b.idx = pathID

	// Set approximation scale based on current transformation
	b.curve.SetApproximationScale(b.Scale())

	switch pathID {
	case 0: // Control line 1 (P0 to P1)
		b.curve.Init(
			b.poly.Xn(0), b.poly.Yn(0), // P0
			(b.poly.Xn(0)+b.poly.Xn(1))*0.5, (b.poly.Yn(0)+b.poly.Yn(1))*0.5, // midpoint
			(b.poly.Xn(0)+b.poly.Xn(1))*0.5, (b.poly.Yn(0)+b.poly.Yn(1))*0.5, // midpoint
			b.poly.Xn(1), b.poly.Yn(1), // P1
		)
		b.stroke.Rewind(0)

	case 1: // Control line 2 (P2 to P3)
		b.curve.Init(
			b.poly.Xn(2), b.poly.Yn(2), // P2
			(b.poly.Xn(2)+b.poly.Xn(3))*0.5, (b.poly.Yn(2)+b.poly.Yn(3))*0.5, // midpoint
			(b.poly.Xn(2)+b.poly.Xn(3))*0.5, (b.poly.Yn(2)+b.poly.Yn(3))*0.5, // midpoint
			b.poly.Xn(3), b.poly.Yn(3), // P3
		)
		b.stroke.Rewind(0)

	case 2: // Actual Bezier curve
		b.curve.Init(
			b.poly.Xn(0), b.poly.Yn(0), // P0
			b.poly.Xn(1), b.poly.Yn(1), // P1
			b.poly.Xn(2), b.poly.Yn(2), // P2
			b.poly.Xn(3), b.poly.Yn(3), // P3
		)
		b.stroke.Rewind(0)

	case 3: // Point 1 (P0)
		r := b.poly.PointRadius()
		b.ellipse.Init(b.poly.Xn(0), b.poly.Yn(0), r, r, 20, false)
		b.ellipse.Rewind(0)

	case 4: // Point 2 (P1)
		r := b.poly.PointRadius()
		b.ellipse.Init(b.poly.Xn(1), b.poly.Yn(1), r, r, 20, false)
		b.ellipse.Rewind(0)

	case 5: // Point 3 (P2)
		r := b.poly.PointRadius()
		b.ellipse.Init(b.poly.Xn(2), b.poly.Yn(2), r, r, 20, false)
		b.ellipse.Rewind(0)

	case 6: // Point 4 (P3)
		r := b.poly.PointRadius()
		b.ellipse.Init(b.poly.Xn(3), b.poly.Yn(3), r, r, 20, false)
		b.ellipse.Rewind(0)
	}
}

// Vertex returns the next vertex for rendering.
func (b *BezierCtrl) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch b.idx {
	case 0, 1, 2: // Lines and curve - use stroke
		x, y, cmd = b.stroke.Vertex()
	case 3, 4, 5, 6: // Points - use ellipse
		cmd = b.ellipse.Vertex(&x, &y)
	default:
		return 0, 0, basics.PathCmdStop
	}

	if cmd != basics.PathCmdStop {
		b.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// Color returns the color for the specified path.
func (b *BezierCtrl) Color(pathID uint) interface{} {
	return b.lineColor
}
