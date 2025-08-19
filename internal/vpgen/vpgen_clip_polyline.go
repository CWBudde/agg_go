package vpgen

import (
	"agg_go/internal/basics"
)

// VPGenClipPolyline clips polylines to a rectangular clipping region.
// Unlike polygon clipping, polylines are not closed and handle endpoints differently.
type VPGenClipPolyline struct {
	clipBox     basics.RectD
	x1, y1      float64
	x, y        [2]float64
	cmd         [2]basics.PathCommand
	numVertices uint32
	vertex      uint32
	moveTo      bool
}

// NewVPGenClipPolyline creates a new polyline clipping vertex processor.
func NewVPGenClipPolyline() *VPGenClipPolyline {
	return &VPGenClipPolyline{
		clipBox: basics.RectD{X1: 0, Y1: 0, X2: 1, Y2: 1},
	}
}

// SetClipBox sets the clipping rectangle.
func (v *VPGenClipPolyline) SetClipBox(x1, y1, x2, y2 float64) {
	v.clipBox.X1 = x1
	v.clipBox.Y1 = y1
	v.clipBox.X2 = x2
	v.clipBox.Y2 = y2
	v.clipBox.Normalize()
}

// X1 returns the left edge of the clip box.
func (v *VPGenClipPolyline) X1() float64 {
	return v.clipBox.X1
}

// Y1 returns the bottom edge of the clip box.
func (v *VPGenClipPolyline) Y1() float64 {
	return v.clipBox.Y1
}

// X2 returns the right edge of the clip box.
func (v *VPGenClipPolyline) X2() float64 {
	return v.clipBox.X2
}

// Y2 returns the top edge of the clip box.
func (v *VPGenClipPolyline) Y2() float64 {
	return v.clipBox.Y2
}

// AutoClose returns false indicating polylines should not be automatically closed.
func (v *VPGenClipPolyline) AutoClose() bool {
	return false
}

// AutoUnclose returns true indicating polylines can be unclosed.
func (v *VPGenClipPolyline) AutoUnclose() bool {
	return true
}

// Reset resets the vertex processor state.
func (v *VPGenClipPolyline) Reset() {
	v.vertex = 0
	v.numVertices = 0
	v.moveTo = false
}

// MoveTo processes a move-to command.
func (v *VPGenClipPolyline) MoveTo(x, y float64) {
	v.vertex = 0
	v.numVertices = 0
	v.x1 = x
	v.y1 = y
	v.moveTo = true
}

// LineTo processes a line-to command.
func (v *VPGenClipPolyline) LineTo(x, y float64) {
	x2 := x
	y2 := y
	flags := basics.ClipLineSegment(&v.x1, &v.y1, &x2, &y2, v.clipBox)

	v.vertex = 0
	v.numVertices = 0

	// Check if line is not completely outside (flag 4 means completely clipped)
	if (flags & 4) == 0 {
		// Check if we need to emit a MoveTo (flag 1 means start point was clipped, or first line)
		if (flags&1) != 0 || v.moveTo {
			v.x[0] = v.x1
			v.y[0] = v.y1
			v.cmd[0] = basics.PathCmdMoveTo
			v.numVertices = 1
		}
		// Always emit the LineTo
		v.x[v.numVertices] = x2
		v.y[v.numVertices] = y2
		v.cmd[v.numVertices] = basics.PathCmdLineTo
		v.numVertices++
		// Flag 2 means end point was clipped, so next line needs MoveTo
		v.moveTo = (flags & 2) != 0
	}

	// Update current position to original endpoint (before clipping)
	v.x1 = x
	v.y1 = y
}

// Vertex returns the next vertex in the clipped output.
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
