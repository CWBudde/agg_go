package conv

import (
	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

// ConvCurve converts Bezier curves in a path to line segments
// This is equivalent to agg::conv_curve in the original AGG library
type ConvCurve struct {
	source VertexSource
	lastX  float64
	lastY  float64
	curve3 *curves.Curve3
	curve4 *curves.Curve4
}

// NewConvCurve creates a new curve converter
func NewConvCurve(source VertexSource) *ConvCurve {
	return &ConvCurve{
		source: source,
		lastX:  0.0,
		lastY:  0.0,
		curve3: curves.NewCurve3(),
		curve4: curves.NewCurve4(),
	}
}

// Attach attaches a new vertex source
func (c *ConvCurve) Attach(source VertexSource) {
	c.source = source
}

// ApproximationMethod returns the current approximation method
func (c *ConvCurve) ApproximationMethod() curves.CurveApproximationMethod {
	return c.curve4.ApproximationMethod()
}

// SetApproximationMethod sets the approximation method for both curve types
func (c *ConvCurve) SetApproximationMethod(method curves.CurveApproximationMethod) {
	c.curve3.SetApproximationMethod(method)
	c.curve4.SetApproximationMethod(method)
}

// ApproximationScale returns the current approximation scale
func (c *ConvCurve) ApproximationScale() float64 {
	return c.curve4.ApproximationScale()
}

// SetApproximationScale sets the approximation scale for both curve types
func (c *ConvCurve) SetApproximationScale(scale float64) {
	c.curve3.SetApproximationScale(scale)
	c.curve4.SetApproximationScale(scale)
}

// AngleTolerance returns the current angle tolerance
func (c *ConvCurve) AngleTolerance() float64 {
	return c.curve4.AngleTolerance()
}

// SetAngleTolerance sets the angle tolerance for both curve types
func (c *ConvCurve) SetAngleTolerance(tolerance float64) {
	c.curve3.SetAngleTolerance(tolerance)
	c.curve4.SetAngleTolerance(tolerance)
}

// CuspLimit returns the current cusp limit
func (c *ConvCurve) CuspLimit() float64 {
	return c.curve4.CuspLimit()
}

// SetCuspLimit sets the cusp limit for both curve types
func (c *ConvCurve) SetCuspLimit(limit float64) {
	c.curve3.SetCuspLimit(limit)
	c.curve4.SetCuspLimit(limit)
}

// Rewind rewinds the curve converter
func (c *ConvCurve) Rewind(pathID uint) {
	c.source.Rewind(pathID)
	c.lastX = 0.0
	c.lastY = 0.0
	c.curve3.Reset()
	c.curve4.Reset()
}

// Vertex returns the next vertex, converting curves to line segments
func (c *ConvCurve) Vertex() (x, y float64, cmd basics.PathCommand) {
	// First check if we have vertices from curve3
	x, y, cmd = c.curve3.Vertex()
	if cmd != basics.PathCmdStop {
		c.lastX = x
		c.lastY = y
		if cmd == basics.PathCmdMoveTo {
			return x, y, cmd
		}
		return x, y, basics.PathCmdLineTo
	}

	// Then check if we have vertices from curve4
	x, y, cmd = c.curve4.Vertex()
	if cmd != basics.PathCmdStop {
		c.lastX = x
		c.lastY = y
		if cmd == basics.PathCmdMoveTo {
			return x, y, cmd
		}
		return x, y, basics.PathCmdLineTo
	}

	// Get the next command from the source
	x, y, cmd = c.source.Vertex()

	switch cmd {
	case basics.PathCmdCurve3:
		// Get the end point for quadratic Bezier (3 points total)
		endX, endY, _ := c.source.Vertex()

		// Initialize the curve with start, control, and end points
		c.curve3.Init(c.lastX, c.lastY, x, y, endX, endY)

		// Get the first vertex (should be move_to)
		x, y, _ = c.curve3.Vertex()
		// Get the second vertex (first line_to of the approximation)
		x, y, _ = c.curve3.Vertex()

		c.lastX = x
		c.lastY = y
		return x, y, basics.PathCmdLineTo

	case basics.PathCmdCurve4:
		// Get the second control point and end point for cubic Bezier (4 points total)
		ct2X, ct2Y, _ := c.source.Vertex()
		endX, endY, _ := c.source.Vertex()

		// Initialize the curve with start, control1, control2, and end points
		c.curve4.Init(c.lastX, c.lastY, x, y, ct2X, ct2Y, endX, endY)

		// Get the first vertex (should be move_to)
		x, y, _ = c.curve4.Vertex()
		// Get the second vertex (first line_to of the approximation)
		x, y, _ = c.curve4.Vertex()

		c.lastX = x
		c.lastY = y
		return x, y, basics.PathCmdLineTo

	default:
		// For all other commands, just pass through and update last position
		if basics.IsVertex(cmd) {
			c.lastX = x
			c.lastY = y
		}
		return x, y, cmd
	}
}
