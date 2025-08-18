// Package shapes provides vector shape generators for the AGG library.
package shapes

import (
	"math"

	"agg_go/internal/basics"
)

// RoundedRect generates vertices for a rectangle with rounded corners.
// This is a direct port of AGG's rounded_rect class from agg_rounded_rect.h/cpp.
// It generates a closed path consisting of four arcs (one for each corner)
// connected by straight line segments.
//
// The corners are numbered as follows:
//
//	rx1,ry1 (bottom-left)    rx2,ry2 (bottom-right)
//	rx4,ry4 (top-left)       rx3,ry3 (top-right)
type RoundedRect struct {
	// Rectangle bounds
	x1, y1, x2, y2 float64

	// Corner radii (each corner has independent X and Y radius)
	rx1, ry1, rx2, ry2, rx3, ry3, rx4, ry4 float64

	// State machine for vertex generation
	status uint32
	arc    *Arc
}

// NewRoundedRect creates a new rounded rectangle with all corners having the same radius.
func NewRoundedRect(x1, y1, x2, y2, r float64) *RoundedRect {
	rr := &RoundedRect{
		arc: NewArc(),
	}
	rr.SetRect(x1, y1, x2, y2)
	rr.SetRadius(r)
	return rr
}

// NewRoundedRectEmpty creates a new empty rounded rectangle.
func NewRoundedRectEmpty() *RoundedRect {
	return &RoundedRect{
		arc: NewArc(),
	}
}

// SetRect sets the rectangle bounds.
// Ensures that x1,y1 is bottom-left and x2,y2 is top-right.
func (rr *RoundedRect) SetRect(x1, y1, x2, y2 float64) {
	rr.x1 = x1
	rr.y1 = y1
	rr.x2 = x2
	rr.y2 = y2

	// Normalize coordinates so that x1 <= x2 and y1 <= y2
	if x1 > x2 {
		rr.x1, rr.x2 = x2, x1
	}
	if y1 > y2 {
		rr.y1, rr.y2 = y2, y1
	}
}

// SetRadius sets all corners to have the same circular radius.
func (rr *RoundedRect) SetRadius(r float64) {
	rr.rx1 = r
	rr.ry1 = r
	rr.rx2 = r
	rr.ry2 = r
	rr.rx3 = r
	rr.ry3 = r
	rr.rx4 = r
	rr.ry4 = r
}

// SetRadiusXY sets all corners to have the same elliptical radius.
func (rr *RoundedRect) SetRadiusXY(rx, ry float64) {
	rr.rx1 = rx
	rr.rx2 = rx
	rr.rx3 = rx
	rr.rx4 = rx
	rr.ry1 = ry
	rr.ry2 = ry
	rr.ry3 = ry
	rr.ry4 = ry
}

// SetRadiusBottomTop sets bottom corners to one radius and top corners to another.
func (rr *RoundedRect) SetRadiusBottomTop(rxBottom, ryBottom, rxTop, ryTop float64) {
	rr.rx1 = rxBottom
	rr.rx2 = rxBottom
	rr.rx3 = rxTop
	rr.rx4 = rxTop
	rr.ry1 = ryBottom
	rr.ry2 = ryBottom
	rr.ry3 = ryTop
	rr.ry4 = ryTop
}

// SetRadiusAll sets each corner to have independent radii.
// The parameters correspond to corners: (x1,y1), (x2,y1), (x2,y2), (x1,y2)
func (rr *RoundedRect) SetRadiusAll(rx1, ry1, rx2, ry2, rx3, ry3, rx4, ry4 float64) {
	rr.rx1 = rx1
	rr.ry1 = ry1
	rr.rx2 = rx2
	rr.ry2 = ry2
	rr.rx3 = rx3
	rr.ry3 = ry3
	rr.rx4 = rx4
	rr.ry4 = ry4
}

// NormalizeRadius ensures that corner radii don't exceed the rectangle dimensions.
// This prevents overlapping arcs by scaling down all radii proportionally if needed.
func (rr *RoundedRect) NormalizeRadius() {
	dx := math.Abs(rr.y2 - rr.y1)
	dy := math.Abs(rr.x2 - rr.x1)

	k := 1.0

	// Check each side to ensure radii don't exceed half the side length
	if t := dx / (rr.rx1 + rr.rx2); t < k {
		k = t
	}
	if t := dx / (rr.rx3 + rr.rx4); t < k {
		k = t
	}
	if t := dy / (rr.ry1 + rr.ry2); t < k {
		k = t
	}
	if t := dy / (rr.ry3 + rr.ry4); t < k {
		k = t
	}

	// Scale down all radii if necessary
	if k < 1.0 {
		rr.rx1 *= k
		rr.ry1 *= k
		rr.rx2 *= k
		rr.ry2 *= k
		rr.rx3 *= k
		rr.ry3 *= k
		rr.rx4 *= k
		rr.ry4 *= k
	}
}

// SetApproximationScale sets the approximation scale for arc generation.
func (rr *RoundedRect) SetApproximationScale(s float64) {
	rr.arc.SetApproximationScale(s)
}

// ApproximationScale returns the current approximation scale.
func (rr *RoundedRect) ApproximationScale() float64 {
	return rr.arc.ApproximationScale()
}

// Rewind resets the rounded rectangle to its starting position for vertex generation.
// The pathId parameter is ignored (kept for interface compatibility).
func (rr *RoundedRect) Rewind(pathId uint32) {
	rr.status = 0
}

// Vertex generates the next vertex along the rounded rectangle path.
// Returns the path command and updates x, y with the vertex coordinates.
// The path consists of 4 arcs connected by line segments, forming a closed polygon.
//
// State machine progression:
// 0-1: Bottom-left arc (from π to 3π/2)
// 2-3: Bottom-right arc (from 3π/2 to 0)
// 4-5: Top-right arc (from 0 to π/2)
// 6-7: Top-left arc (from π/2 to π)
// 8: End polygon with close flag
func (rr *RoundedRect) Vertex(x, y *float64) basics.PathCommand {
	cmd := basics.PathCmdStop

	switch rr.status {
	case 0:
		// Initialize bottom-left arc
		rr.arc.Init(rr.x1+rr.rx1, rr.y1+rr.ry1, rr.rx1, rr.ry1,
			basics.Pi, basics.Pi+basics.Pi*0.5, true)
		rr.arc.Rewind(0)
		rr.status++
		fallthrough

	case 1:
		// Generate bottom-left arc vertices
		cmd = rr.arc.Vertex(x, y)
		if basics.IsStop(cmd) {
			rr.status++
		} else {
			return cmd
		}
		fallthrough

	case 2:
		// Initialize bottom-right arc
		rr.arc.Init(rr.x2-rr.rx2, rr.y1+rr.ry2, rr.rx2, rr.ry2,
			basics.Pi+basics.Pi*0.5, 0.0, true)
		rr.arc.Rewind(0)
		rr.status++
		fallthrough

	case 3:
		// Generate bottom-right arc vertices
		cmd = rr.arc.Vertex(x, y)
		if basics.IsStop(cmd) {
			rr.status++
		} else {
			return basics.PathCmdLineTo
		}
		fallthrough

	case 4:
		// Initialize top-right arc
		rr.arc.Init(rr.x2-rr.rx3, rr.y2-rr.ry3, rr.rx3, rr.ry3,
			0.0, basics.Pi*0.5, true)
		rr.arc.Rewind(0)
		rr.status++
		fallthrough

	case 5:
		// Generate top-right arc vertices
		cmd = rr.arc.Vertex(x, y)
		if basics.IsStop(cmd) {
			rr.status++
		} else {
			return basics.PathCmdLineTo
		}
		fallthrough

	case 6:
		// Initialize top-left arc
		rr.arc.Init(rr.x1+rr.rx4, rr.y2-rr.ry4, rr.rx4, rr.ry4,
			basics.Pi*0.5, basics.Pi, true)
		rr.arc.Rewind(0)
		rr.status++
		fallthrough

	case 7:
		// Generate top-left arc vertices
		cmd = rr.arc.Vertex(x, y)
		if basics.IsStop(cmd) {
			rr.status++
		} else {
			return basics.PathCmdLineTo
		}
		fallthrough

	case 8:
		// End the polygon with close flag
		cmd = basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))
		rr.status++
	}

	return cmd
}
