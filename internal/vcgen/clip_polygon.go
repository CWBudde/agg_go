package vcgen

import (
	"agg_go/internal/basics"
)

// VPGenClipPolygon generates vertices for polygon clipping using the Liang-Barsky algorithm.
// This is a direct translation of the AGG vpgen_clip_polygon class.
type VPGenClipPolygon struct {
	clipBox     basics.RectD
	x1, y1      float64
	clipFlags   uint32
	x, y        [4]float64
	numVertices uint32
	vertex      uint32
	cmd         basics.PathCommand
}

// NewVPGenClipPolygon creates a new polygon clipping vertex processor generator
func NewVPGenClipPolygon() *VPGenClipPolygon {
	return &VPGenClipPolygon{
		clipBox: basics.RectD{X1: 0, Y1: 0, X2: 1, Y2: 1},
		cmd:     basics.PathCmdMoveTo,
	}
}

// ClipBox sets the clipping rectangle
func (v *VPGenClipPolygon) ClipBox(x1, y1, x2, y2 float64) {
	v.clipBox.X1 = x1
	v.clipBox.Y1 = y1
	v.clipBox.X2 = x2
	v.clipBox.Y2 = y2
	v.clipBox.Normalize()
}

// X1 returns the left edge of the clipping box
func (v *VPGenClipPolygon) X1() float64 { return v.clipBox.X1 }

// Y1 returns the bottom edge of the clipping box
func (v *VPGenClipPolygon) Y1() float64 { return v.clipBox.Y1 }

// X2 returns the right edge of the clipping box
func (v *VPGenClipPolygon) X2() float64 { return v.clipBox.X2 }

// Y2 returns the top edge of the clipping box
func (v *VPGenClipPolygon) Y2() float64 { return v.clipBox.Y2 }

// AutoClose returns true since polygons should be automatically closed
func (v *VPGenClipPolygon) AutoClose() bool { return true }

// AutoUnclose returns false since polygons should not be automatically unclosed
func (v *VPGenClipPolygon) AutoUnclose() bool { return false }

// Reset resets the vertex processor state
func (v *VPGenClipPolygon) Reset() {
	v.vertex = 0
	v.numVertices = 0
}

// MoveTo starts a new path at the given coordinates
func (v *VPGenClipPolygon) MoveTo(x, y float64) {
	v.vertex = 0
	v.numVertices = 0
	v.clipFlags = v.clippingFlags(x, y)
	if v.clipFlags == 0 {
		v.x[0] = x
		v.y[0] = y
		v.numVertices = 1
	}
	v.x1 = x
	v.y1 = y
	v.cmd = basics.PathCmdMoveTo
}

// LineTo adds a line segment to the current path
func (v *VPGenClipPolygon) LineTo(x, y float64) {
	v.vertex = 0
	v.numVertices = 0
	flags := v.clippingFlags(x, y)

	if v.clipFlags == flags {
		if flags == 0 {
			v.x[0] = x
			v.y[0] = y
			v.numVertices = 1
		}
	} else {
		v.numVertices = basics.ClipLiangBarsky(
			v.x1, v.y1,
			x, y,
			v.clipBox,
			v.x[:], v.y[:],
		)
	}

	v.clipFlags = flags
	v.x1 = x
	v.y1 = y
}

// Vertex returns the next processed vertex
func (v *VPGenClipPolygon) Vertex() (x, y float64, cmd basics.PathCommand) {
	if v.vertex < v.numVertices {
		x = v.x[v.vertex]
		y = v.y[v.vertex]
		v.vertex++
		cmd = v.cmd
		v.cmd = basics.PathCmdLineTo
		return x, y, cmd
	}
	return 0, 0, basics.PathCmdStop
}

// clippingFlags determines the clipping code of the vertex according to the
// Cyrus-Beck line clipping algorithm
//
//	      |        |
//	0110  |  0010  | 0011
//	      |        |
//
// -------+--------+-------- clip_box.y2
//
//	      |        |
//	0100  |  0000  | 0001
//	      |        |
//
// -------+--------+-------- clip_box.y1
//
//	      |        |
//	1100  |  1000  | 1001
//	      |        |
//	clip_box.x1  clip_box.x2
func (v *VPGenClipPolygon) clippingFlags(x, y float64) uint32 {
	if x < v.clipBox.X1 {
		if y > v.clipBox.Y2 {
			return 6
		}
		if y < v.clipBox.Y1 {
			return 12
		}
		return 4
	}

	if x > v.clipBox.X2 {
		if y > v.clipBox.Y2 {
			return 3
		}
		if y < v.clipBox.Y1 {
			return 9
		}
		return 1
	}

	if y > v.clipBox.Y2 {
		return 2
	}
	if y < v.clipBox.Y1 {
		return 8
	}

	return 0
}
