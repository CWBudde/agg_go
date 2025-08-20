package vpgen

import (
	"math"

	"agg_go/internal/basics"
)

// VPGenSegmentator breaks lines into uniform segments based on approximation scale.
// This is useful for converting long lines into multiple shorter segments for
// better rendering quality or processing.
type VPGenSegmentator struct {
	approximationScale float64
	x1, y1             float64
	dx, dy             float64
	dl, ddl            float64
	cmd                basics.PathCommand
}

// NewVPGenSegmentator creates a new line segmentation vertex processor.
func NewVPGenSegmentator() *VPGenSegmentator {
	return &VPGenSegmentator{
		approximationScale: 1.0,
	}
}

// SetApproximationScale sets the scale factor for segment approximation.
// Higher values create more segments (finer approximation).
func (v *VPGenSegmentator) SetApproximationScale(scale float64) {
	v.approximationScale = scale
}

// ApproximationScale returns the current approximation scale.
func (v *VPGenSegmentator) ApproximationScale() float64 {
	return v.approximationScale
}

// AutoClose returns false indicating lines should not be automatically closed.
func (v *VPGenSegmentator) AutoClose() bool {
	return false
}

// AutoUnclose returns false indicating lines should not be unclosed.
func (v *VPGenSegmentator) AutoUnclose() bool {
	return false
}

// Reset resets the vertex processor state.
func (v *VPGenSegmentator) Reset() {
	v.cmd = basics.PathCmdStop
}

// MoveTo processes a move-to command.
func (v *VPGenSegmentator) MoveTo(x, y float64) {
	v.x1 = x
	v.y1 = y
	v.dx = 0.0
	v.dy = 0.0
	v.dl = 2.0
	v.ddl = 2.0
	v.cmd = basics.PathCmdMoveTo
}

// LineTo processes a line-to command.
func (v *VPGenSegmentator) LineTo(x, y float64) {
	v.x1 += v.dx
	v.y1 += v.dy
	v.dx = x - v.x1
	v.dy = y - v.y1

	length := math.Sqrt(v.dx*v.dx+v.dy*v.dy) * v.approximationScale
	if length < 1e-30 {
		length = 1e-30
	}

	v.ddl = 1.0 / length
	if v.cmd == basics.PathCmdMoveTo {
		v.dl = 0.0
	} else {
		v.dl = v.ddl
	}

	if v.cmd == basics.PathCmdStop {
		v.cmd = basics.PathCmdLineTo
	}
}

// Vertex returns the next vertex in the segmented output.
func (v *VPGenSegmentator) Vertex() (x, y float64, cmd basics.PathCommand) {
	if v.cmd == basics.PathCmdStop {
		return 0, 0, basics.PathCmdStop
	}

	cmd = v.cmd
	v.cmd = basics.PathCmdLineTo

	// Always output the start point first when dl == 0
	if v.dl == 0.0 {
		x = v.x1
		y = v.y1
		v.dl += v.ddl
		return x, y, cmd
	}

	if v.dl >= 1.0-v.ddl {
		v.dl = 1.0
		v.cmd = basics.PathCmdStop
		x = v.x1 + v.dx
		y = v.y1 + v.dy
		return x, y, cmd
	}

	x = v.x1 + v.dx*v.dl
	y = v.y1 + v.dy*v.dl
	v.dl += v.ddl
	return x, y, cmd
}
