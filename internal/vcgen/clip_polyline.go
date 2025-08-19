package vcgen

import (
	"agg_go/internal/basics"
)

// VPGenClipPolyline generates vertices for polyline clipping using the Liang-Barsky algorithm.
// This is a direct translation of the AGG vpgen_clip_polyline class.
// Unlike polygon clipping, this handles line segments and never closes paths.
type VPGenClipPolyline struct {
	clipBox     basics.RectD
	x1, y1      float64
	x, y        [2]float64
	cmd         [2]basics.PathCommand
	numVertices uint32
	vertex      uint32
	moveTo      bool
}

// NewVPGenClipPolyline creates a new polyline clipping vertex processor generator
func NewVPGenClipPolyline() *VPGenClipPolyline {
	return &VPGenClipPolyline{
		clipBox: basics.RectD{X1: 0, Y1: 0, X2: 1, Y2: 1},
	}
}

// ClipBox sets the clipping rectangle
func (v *VPGenClipPolyline) ClipBox(x1, y1, x2, y2 float64) {
	v.clipBox.X1 = x1
	v.clipBox.Y1 = y1
	v.clipBox.X2 = x2
	v.clipBox.Y2 = y2
	v.clipBox.Normalize()
}

// X1 returns the left edge of the clipping box
func (v *VPGenClipPolyline) X1() float64 { return v.clipBox.X1 }

// Y1 returns the bottom edge of the clipping box
func (v *VPGenClipPolyline) Y1() float64 { return v.clipBox.Y1 }

// X2 returns the right edge of the clipping box
func (v *VPGenClipPolyline) X2() float64 { return v.clipBox.X2 }

// Y2 returns the top edge of the clipping box
func (v *VPGenClipPolyline) Y2() float64 { return v.clipBox.Y2 }

// AutoClose returns false since polylines should not be automatically closed
func (v *VPGenClipPolyline) AutoClose() bool { return false }

// AutoUnclose returns true since polylines should be automatically unclosed
func (v *VPGenClipPolyline) AutoUnclose() bool { return true }

// Reset resets the vertex processor state
func (v *VPGenClipPolyline) Reset() {
	v.vertex = 0
	v.numVertices = 0
	v.moveTo = false
}

// MoveTo starts a new path at the given coordinates
func (v *VPGenClipPolyline) MoveTo(x, y float64) {
	v.vertex = 0
	v.numVertices = 0
	v.x1 = x
	v.y1 = y
	v.moveTo = true
}

// LineTo adds a line segment to the current path
func (v *VPGenClipPolyline) LineTo(x, y float64) {
	x2 := x
	y2 := y

	// Use the existing ClipLineSegment function
	flags := basics.ClipLineSegment(&v.x1, &v.y1, &x2, &y2, v.clipBox)

	v.vertex = 0
	v.numVertices = 0

	// Check if line segment is not completely clipped (flag 4 means completely clipped)
	if (flags & 4) == 0 {
		// If first endpoint was clipped or we need a move_to due to disconnection, add move_to vertex
		if (flags&1) != 0 || v.moveTo {
			v.x[0] = v.x1
			v.y[0] = v.y1
			v.cmd[0] = basics.PathCmdMoveTo
			v.numVertices = 1
		}

		// Add the line_to vertex
		v.x[v.numVertices] = x2
		v.y[v.numVertices] = y2
		v.cmd[v.numVertices] = basics.PathCmdLineTo
		v.numVertices++

		// Update moveTo flag - set to true if second endpoint was clipped (creates disconnection)
		v.moveTo = (flags & 2) != 0
	}

	// Always update current position for next segment
	v.x1 = x
	v.y1 = y
}

// Vertex returns the next processed vertex
func (v *VPGenClipPolyline) Vertex() (x, y float64, cmd basics.PathCommand) {
	if v.vertex < v.numVertices {
		x = v.x[v.vertex]
		y = v.y[v.vertex]
		cmd = v.cmd[v.vertex]
		v.vertex++
		return x, y, cmd
	}
	return 0, 0, basics.PathCmdStop
}
