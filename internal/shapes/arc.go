// Package shapes provides vector shape generators for the AGG library.
// This package implements various geometric shapes that can generate vertices
// for rendering, including arcs, ellipses, and other primitives.
package shapes

import (
	"math"

	"agg_go/internal/basics"
)

// Arc generates vertices along an elliptical arc.
// This is a direct port of AGG's arc class from agg_arc.h/cpp.
// It generates line segments that approximate an elliptical arc between
// two angles, with adaptive tessellation based on approximation scale.
type Arc struct {
	x           float64            // Center X coordinate
	y           float64            // Center Y coordinate
	rx          float64            // X radius
	ry          float64            // Y radius
	angle       float64            // Current angle during vertex generation
	start       float64            // Start angle (normalized)
	end         float64            // End angle (normalized)
	scale       float64            // Approximation scale factor
	da          float64            // Angular step size
	ccw         bool               // Counter-clockwise flag
	initialized bool               // Whether the arc has been initialized
	pathCmd     basics.PathCommand // Current path command
}

// NewArc creates a new arc with default values.
// The arc is not initialized until Init() is called.
func NewArc() *Arc {
	return &Arc{
		scale:       1.0,
		initialized: false,
	}
}

// NewArcWithParams creates and initializes a new arc with the specified parameters.
func NewArcWithParams(x, y, rx, ry, a1, a2 float64, ccw bool) *Arc {
	arc := &Arc{
		x:     x,
		y:     y,
		rx:    rx,
		ry:    ry,
		scale: 1.0,
	}
	arc.normalize(a1, a2, ccw)
	return arc
}

// Init initializes the arc with the specified parameters.
// x, y: center coordinates
// rx, ry: radii in X and Y directions
// a1, a2: start and end angles in radians
// ccw: true for counter-clockwise, false for clockwise
func (a *Arc) Init(x, y, rx, ry, a1, a2 float64, ccw bool) {
	a.x = x
	a.y = y
	a.rx = rx
	a.ry = ry
	a.normalize(a1, a2, ccw)
}

// SetApproximationScale sets the approximation scale factor.
// Larger values produce more line segments (higher quality but slower).
// If the arc is already initialized, it recalculates the angular step.
func (a *Arc) SetApproximationScale(s float64) {
	a.scale = s
	if a.initialized {
		a.normalize(a.start, a.end, a.ccw)
	}
}

// ApproximationScale returns the current approximation scale factor.
func (a *Arc) ApproximationScale() float64 {
	return a.scale
}

// Rewind resets the arc to its starting position for vertex generation.
// The pathId parameter is ignored (kept for interface compatibility).
func (a *Arc) Rewind(pathId uint32) {
	a.pathCmd = basics.PathCmdMoveTo
	a.angle = a.start
}

// Vertex generates the next vertex along the arc path.
// Returns the path command and updates x, y with the vertex coordinates.
// Returns PathCmdStop when the arc is complete.
func (a *Arc) Vertex(x, y *float64) basics.PathCommand {
	if basics.IsStop(a.pathCmd) {
		return basics.PathCmdStop
	}

	// Check if we've reached the end of the arc
	// The condition (a.angle < a.end-a.da/4) != a.ccw means:
	// - For CCW: if angle >= end-da/4, we're done
	// - For CW: if angle <= end-da/4, we're done
	if (a.angle < a.end-a.da/4) != a.ccw {
		// Generate the final vertex at the exact end angle
		*x = a.x + math.Cos(a.end)*a.rx
		*y = a.y + math.Sin(a.end)*a.ry
		a.pathCmd = basics.PathCmdStop
		return basics.PathCmdLineTo
	}

	// Generate vertex at current angle
	*x = a.x + math.Cos(a.angle)*a.rx
	*y = a.y + math.Sin(a.angle)*a.ry

	// Advance to next angle
	a.angle += a.da

	// Return appropriate command and update state
	pf := a.pathCmd
	a.pathCmd = basics.PathCmdLineTo
	return pf
}

// normalize calculates the angular step and normalizes angles for the arc.
// This is where the adaptive tessellation magic happens - the step size
// is calculated based on the approximation scale to ensure smooth curves.
func (a *Arc) normalize(a1, a2 float64, ccw bool) {
	// Calculate average radius for step size computation
	ra := (math.Abs(a.rx) + math.Abs(a.ry)) / 2.0

	// Calculate angular step based on approximation scale
	// This formula ensures that the chord error is proportional to 1/scale
	a.da = math.Acos(ra/(ra+0.125/a.scale)) * 2.0

	if ccw {
		// Counter-clockwise: ensure a2 >= a1
		for a2 < a1 {
			a2 += 2.0 * basics.Pi
		}
	} else {
		// Clockwise: ensure a1 >= a2, and negate step
		for a1 < a2 {
			a1 += 2.0 * basics.Pi
		}
		a.da = -a.da
	}

	a.ccw = ccw
	a.start = a1
	a.end = a2
	a.initialized = true
}
