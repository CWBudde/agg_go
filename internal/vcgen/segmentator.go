package vcgen

import (
	"math"

	"agg_go/internal/basics"
)

// VPGenSegmentator is the segmentator vertex processor generator.
// This is a port of AGG's vpgen_segmentator class.
//
// The segmentator divides line segments into equal-length pieces
// based on the approximation scale. It's useful for creating
// evenly-spaced points along paths.
type VPGenSegmentator struct {
	approximationScale float64
	x1                 float64
	y1                 float64
	dx                 float64
	dy                 float64
	dl                 float64 // current position along segment (0.0 to 1.0)
	ddl                float64 // step size for each segment
	cmd                basics.PathCommand
}

// NewVPGenSegmentator creates a new segmentator vertex processor generator
func NewVPGenSegmentator() *VPGenSegmentator {
	return &VPGenSegmentator{
		approximationScale: 1.0,
		cmd:                basics.PathCmdStop,
	}
}

// ApproximationScale sets the approximation scale factor.
// Higher values create more segments per unit length.
func (s *VPGenSegmentator) ApproximationScale(scale float64) {
	s.approximationScale = scale
}

// GetApproximationScale returns the current approximation scale
func (s *VPGenSegmentator) GetApproximationScale() float64 {
	return s.approximationScale
}

// AutoClose returns false - segmentator doesn't auto-close polygons
func (s *VPGenSegmentator) AutoClose() bool {
	return false
}

// AutoUnclose returns false - segmentator doesn't auto-unclose polygons
func (s *VPGenSegmentator) AutoUnclose() bool {
	return false
}

// Reset resets the segmentator to initial state
func (s *VPGenSegmentator) Reset() {
	s.cmd = basics.PathCmdStop
}

// MoveTo starts a new path at the given coordinates
func (s *VPGenSegmentator) MoveTo(x, y float64) {
	s.x1 = x
	s.y1 = y
	s.dx = 0.0
	s.dy = 0.0
	s.dl = 2.0 // Start beyond range to trigger initial vertex output
	s.ddl = 2.0
	s.cmd = basics.PathCmdMoveTo
}

// LineTo adds a line segment to be processed
func (s *VPGenSegmentator) LineTo(x, y float64) {
	// Update current position and calculate vector to new point
	s.x1 += s.dx
	s.y1 += s.dy
	s.dx = x - s.x1
	s.dy = y - s.y1

	// Calculate length and determine step size
	length := math.Sqrt(s.dx*s.dx+s.dy*s.dy) * s.approximationScale
	if length < 1e-30 {
		length = 1e-30 // Prevent division by zero
	}
	s.ddl = 1.0 / length

	// Set initial position: 0.0 for move_to, ddl for continuing lines
	if s.cmd == basics.PathCmdMoveTo {
		s.dl = 0.0
	} else {
		s.dl = s.ddl
	}

	// Transition from stop to line_to if needed
	if s.cmd == basics.PathCmdStop {
		s.cmd = basics.PathCmdLineTo
	}
}

// Vertex returns the next segmented vertex
func (s *VPGenSegmentator) Vertex() (x, y float64, cmd basics.PathCommand) {
	if s.cmd == basics.PathCmdStop {
		return 0, 0, basics.PathCmdStop
	}

	cmd = s.cmd
	s.cmd = basics.PathCmdLineTo

	// Check if we've reached the end of the segment
	if s.dl >= 1.0-s.ddl {
		s.dl = 1.0
		s.cmd = basics.PathCmdStop
		x = s.x1 + s.dx
		y = s.y1 + s.dy
		return x, y, cmd
	}

	// Calculate intermediate point along the segment
	x = s.x1 + s.dx*s.dl
	y = s.y1 + s.dy*s.dl
	s.dl += s.ddl
	return x, y, cmd
}
