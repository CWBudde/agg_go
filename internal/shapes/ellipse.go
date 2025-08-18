package shapes

import (
	"math"

	"agg_go/internal/basics"
)

// Ellipse generates vertices for a complete ellipse.
// This is a direct port of AGG's ellipse class from agg_ellipse.h.
// It generates line segments that approximate an ellipse with adaptive
// tessellation based on approximation scale. Unlike Arc, Ellipse always
// generates a complete closed ellipse.
type Ellipse struct {
	x     float64 // Center X coordinate
	y     float64 // Center Y coordinate
	rx    float64 // X radius
	ry    float64 // Y radius
	scale float64 // Approximation scale factor
	num   uint32  // Number of steps for tessellation
	step  uint32  // Current step during vertex generation
	cw    bool    // Clockwise flag (false = counter-clockwise)
}

// NewEllipse creates a new ellipse with default values.
// Default center is (0,0), radii are (1,1), and it's counter-clockwise.
func NewEllipse() *Ellipse {
	return &Ellipse{
		x:     0.0,
		y:     0.0,
		rx:    1.0,
		ry:    1.0,
		scale: 1.0,
		num:   4,
		step:  0,
		cw:    false,
	}
}

// NewEllipseWithParams creates and initializes a new ellipse with the specified parameters.
// x, y: center coordinates
// rx, ry: radii in X and Y directions
// numSteps: number of tessellation steps (0 = auto-calculate)
// cw: true for clockwise, false for counter-clockwise
func NewEllipseWithParams(x, y, rx, ry float64, numSteps uint32, cw bool) *Ellipse {
	ellipse := &Ellipse{
		x:     x,
		y:     y,
		rx:    rx,
		ry:    ry,
		scale: 1.0,
		num:   numSteps,
		step:  0,
		cw:    cw,
	}
	if ellipse.num == 0 {
		ellipse.calcNumSteps()
	}
	return ellipse
}

// Init initializes the ellipse with the specified parameters.
// x, y: center coordinates
// rx, ry: radii in X and Y directions
// numSteps: number of tessellation steps (0 = auto-calculate)
// cw: true for clockwise, false for counter-clockwise
func (e *Ellipse) Init(x, y, rx, ry float64, numSteps uint32, cw bool) {
	e.x = x
	e.y = y
	e.rx = rx
	e.ry = ry
	e.num = numSteps
	e.step = 0
	e.cw = cw
	if e.num == 0 {
		e.calcNumSteps()
	}
}

// SetApproximationScale sets the approximation scale factor.
// Larger values produce more line segments (higher quality but slower).
// Automatically recalculates the number of tessellation steps.
func (e *Ellipse) SetApproximationScale(scale float64) {
	e.scale = scale
	e.calcNumSteps()
}

// ApproximationScale returns the current approximation scale factor.
func (e *Ellipse) ApproximationScale() float64 {
	return e.scale
}

// Rewind resets the ellipse to its starting position for vertex generation.
// The pathId parameter is ignored (kept for interface compatibility).
func (e *Ellipse) Rewind(pathId uint32) {
	e.step = 0
}

// Vertex generates the next vertex along the ellipse path.
// Returns the path command and updates x, y with the vertex coordinates.
// The first vertex is MoveTo, subsequent vertices are LineTo,
// and finally EndPoly with close flags when the ellipse is complete.
func (e *Ellipse) Vertex(x, y *float64) basics.PathCommand {
	if e.step == e.num {
		e.step++
		return basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))
	}
	if e.step > e.num {
		return basics.PathCmdStop
	}

	// Calculate angle for current step
	angle := float64(e.step) / float64(e.num) * 2.0 * basics.Pi
	if e.cw {
		angle = 2.0*basics.Pi - angle
	}

	// Calculate vertex position
	*x = e.x + math.Cos(angle)*e.rx
	*y = e.y + math.Sin(angle)*e.ry

	e.step++
	if e.step == 1 {
		return basics.PathCmdMoveTo
	}
	return basics.PathCmdLineTo
}

// calcNumSteps calculates the optimal number of tessellation steps
// based on the ellipse radii and approximation scale.
// This uses the same adaptive algorithm as AGG to ensure smooth curves
// with minimal vertex count.
func (e *Ellipse) calcNumSteps() {
	// Use average radius for step calculation
	ra := (math.Abs(e.rx) + math.Abs(e.ry)) / 2.0

	// Calculate angular step based on chord error tolerance
	// The formula ensures the approximation error is proportional to 1/scale
	da := math.Acos(ra/(ra+0.125/e.scale)) * 2.0

	// Calculate number of steps for full circle
	e.num = basics.URound(2.0 * basics.Pi / da)
}
