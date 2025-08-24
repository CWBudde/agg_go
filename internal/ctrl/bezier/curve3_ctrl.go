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

// Curve3Ctrl implements an interactive quadratic Bezier curve control.
// This corresponds to AGG's curve3_ctrl_impl class.
type Curve3Ctrl[C any] struct {
	*ctrl.BaseCtrl

	// Quadratic curve and rendering components
	curve   *curves.Curve3
	ellipse *shapes.Ellipse
	stroke  *conv.ConvStroke
	poly    *polygon.PolygonCtrl[C]

	// Rendering state
	idx uint

	// Color for rendering
	lineColor C
}

// NewCurve3Ctrl creates a new quadratic Bezier curve control.
// The control starts with default points that form a visible curve.
func NewCurve3Ctrl[C any](lineColor C) *Curve3Ctrl[C] {
	// Create a 3-point polygon for the quadratic Bezier control points
	poly := polygon.NewPolygonCtrl[C](3, 5.0, lineColor)
	poly.SetInPolygonCheck(false)

	// Set default control points that create a nice visible curve
	poly.SetXn(0, 100.0) // P0 - start point
	poly.SetYn(0, 0.0)
	poly.SetXn(1, 100.0) // P1 - control point
	poly.SetYn(1, 50.0)
	poly.SetXn(2, 50.0) // P2 - end point
	poly.SetYn(2, 100.0)

	curve := curves.NewCurve3()
	ellipse := shapes.NewEllipse()
	stroke := conv.NewConvStroke(curve)

	ctrl := &Curve3Ctrl[C]{
		BaseCtrl:  ctrl.NewBaseCtrl(0, 0, 1, 1, false),
		curve:     curve,
		ellipse:   ellipse,
		stroke:    stroke,
		poly:      poly,
		idx:       0,
		lineColor: lineColor,
	}

	// Initialize the curve with the default points
	ctrl.updateCurve()

	return ctrl
}

// NewDefaultCurve3Ctrl creates a Curve3 control with default RGBA color (black).
// This provides backward compatibility for existing code.
func NewDefaultCurve3Ctrl() *Curve3Ctrl[color.RGBA] {
	defaultColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0) // default black
	return NewCurve3Ctrl[color.RGBA](defaultColor)
}

// Curve Management Methods

// SetCurve sets the quadratic Bezier curve control points.
func (c *Curve3Ctrl[C]) SetCurve(x1, y1, x2, y2, x3, y3 float64) {
	c.poly.SetXn(0, x1)
	c.poly.SetYn(0, y1)
	c.poly.SetXn(1, x2)
	c.poly.SetYn(1, y2)
	c.poly.SetXn(2, x3)
	c.poly.SetYn(2, y3)
	c.updateCurve()
}

// Curve returns the internal curve object after updating it with current points.
func (c *Curve3Ctrl[C]) Curve() *curves.Curve3 {
	c.updateCurve()
	return c.curve
}

// updateCurve updates the internal curve with current control points.
func (c *Curve3Ctrl[C]) updateCurve() {
	c.curve.Init(
		c.poly.Xn(0), c.poly.Yn(0), // P0
		c.poly.Xn(1), c.poly.Yn(1), // P1
		c.poly.Xn(2), c.poly.Yn(2), // P2
	)
}

// Control Point Access Methods

// X1 returns the X coordinate of the start point (P0).
func (c *Curve3Ctrl[C]) X1() float64 { return c.poly.Xn(0) }

// Y1 returns the Y coordinate of the start point (P0).
func (c *Curve3Ctrl[C]) Y1() float64 { return c.poly.Yn(0) }

// X2 returns the X coordinate of the control point (P1).
func (c *Curve3Ctrl[C]) X2() float64 { return c.poly.Xn(1) }

// Y2 returns the Y coordinate of the control point (P1).
func (c *Curve3Ctrl[C]) Y2() float64 { return c.poly.Yn(1) }

// X3 returns the X coordinate of the end point (P2).
func (c *Curve3Ctrl[C]) X3() float64 { return c.poly.Xn(2) }

// Y3 returns the Y coordinate of the end point (P2).
func (c *Curve3Ctrl[C]) Y3() float64 { return c.poly.Yn(2) }

// SetX1 sets the X coordinate of the start point (P0).
func (c *Curve3Ctrl[C]) SetX1(x float64) { c.poly.SetXn(0, x) }

// SetY1 sets the Y coordinate of the start point (P0).
func (c *Curve3Ctrl[C]) SetY1(y float64) { c.poly.SetYn(0, y) }

// SetX2 sets the X coordinate of the control point (P1).
func (c *Curve3Ctrl[C]) SetX2(x float64) { c.poly.SetXn(1, x) }

// SetY2 sets the Y coordinate of the control point (P1).
func (c *Curve3Ctrl[C]) SetY2(y float64) { c.poly.SetYn(1, y) }

// SetX3 sets the X coordinate of the end point (P2).
func (c *Curve3Ctrl[C]) SetX3(x float64) { c.poly.SetXn(2, x) }

// SetY3 sets the Y coordinate of the end point (P2).
func (c *Curve3Ctrl[C]) SetY3(y float64) { c.poly.SetYn(2, y) }

// Stroke Configuration Methods

// SetLineWidth sets the width of the curve and control line strokes.
func (c *Curve3Ctrl[C]) SetLineWidth(w float64) {
	c.stroke.SetWidth(w)
}

// LineWidth returns the current stroke width.
func (c *Curve3Ctrl[C]) LineWidth() float64 {
	return c.stroke.Width()
}

// SetPointRadius sets the radius for control point rendering and hit testing.
func (c *Curve3Ctrl[C]) SetPointRadius(r float64) {
	c.poly.SetPointRadius(r)
}

// PointRadius returns the current point radius.
func (c *Curve3Ctrl[C]) PointRadius() float64 {
	return c.poly.PointRadius()
}

// Color Management

// SetLineColor sets the line color for rendering.
func (c *Curve3Ctrl[C]) SetLineColor(clr C) {
	c.lineColor = clr
	c.poly.SetLineColor(clr)
}

// LineColor returns the current line color.
func (c *Curve3Ctrl[C]) LineColor() C {
	return c.lineColor
}

// Mouse Interaction Methods (delegate to polygon control)

// OnMouseButtonDown handles mouse button press events.
func (c *Curve3Ctrl[C]) OnMouseButtonDown(x, y float64) bool {
	return c.poly.OnMouseButtonDown(x, y)
}

// OnMouseButtonUp handles mouse button release events.
func (c *Curve3Ctrl[C]) OnMouseButtonUp(x, y float64) bool {
	return c.poly.OnMouseButtonUp(x, y)
}

// OnMouseMove handles mouse move events.
func (c *Curve3Ctrl[C]) OnMouseMove(x, y float64, buttonPressed bool) bool {
	return c.poly.OnMouseMove(x, y, buttonPressed)
}

// OnArrowKeys handles arrow key events.
func (c *Curve3Ctrl[C]) OnArrowKeys(left, right, down, up bool) bool {
	return c.poly.OnArrowKeys(left, right, down, up)
}

// Vertex Source Interface

// NumPaths returns the number of rendering paths.
// Returns 6: control line 1, control line 2, curve, point 1, point 2, point 3
func (c *Curve3Ctrl[C]) NumPaths() uint {
	return 6
}

// Rewind resets the vertex generation state.
func (c *Curve3Ctrl[C]) Rewind(pathID uint) {
	c.idx = pathID

	switch pathID {
	case 0: // Control line 1 (P0 to P1)
		c.curve.Init(
			c.poly.Xn(0), c.poly.Yn(0), // P0
			(c.poly.Xn(0)+c.poly.Xn(1))*0.5, (c.poly.Yn(0)+c.poly.Yn(1))*0.5, // midpoint
			c.poly.Xn(1), c.poly.Yn(1), // P1
		)
		c.stroke.Rewind(0)

	case 1: // Control line 2 (P1 to P2)
		c.curve.Init(
			c.poly.Xn(1), c.poly.Yn(1), // P1
			(c.poly.Xn(1)+c.poly.Xn(2))*0.5, (c.poly.Yn(1)+c.poly.Yn(2))*0.5, // midpoint
			c.poly.Xn(2), c.poly.Yn(2), // P2
		)
		c.stroke.Rewind(0)

	case 2: // Actual quadratic Bezier curve
		c.curve.Init(
			c.poly.Xn(0), c.poly.Yn(0), // P0
			c.poly.Xn(1), c.poly.Yn(1), // P1
			c.poly.Xn(2), c.poly.Yn(2), // P2
		)
		c.stroke.Rewind(0)

	case 3: // Point 1 (P0)
		r := c.poly.PointRadius()
		c.ellipse.Init(c.poly.Xn(0), c.poly.Yn(0), r, r, 20, false)
		c.ellipse.Rewind(0)

	case 4: // Point 2 (P1)
		r := c.poly.PointRadius()
		c.ellipse.Init(c.poly.Xn(1), c.poly.Yn(1), r, r, 20, false)
		c.ellipse.Rewind(0)

	case 5: // Point 3 (P2)
		r := c.poly.PointRadius()
		c.ellipse.Init(c.poly.Xn(2), c.poly.Yn(2), r, r, 20, false)
		c.ellipse.Rewind(0)
	}
}

// Vertex returns the next vertex for rendering.
func (c *Curve3Ctrl[C]) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch c.idx {
	case 0, 1, 2: // Lines and curve - use stroke
		x, y, cmd = c.stroke.Vertex()
	case 3, 4, 5: // Points - use ellipse
		cmd = c.ellipse.Vertex(&x, &y)
	default:
		return 0, 0, basics.PathCmdStop
	}

	if cmd != basics.PathCmdStop {
		c.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// Color returns the color for the specified path.
func (c *Curve3Ctrl[C]) Color(pathID uint) C {
	return c.lineColor
}
