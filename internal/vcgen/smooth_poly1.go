package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// SmoothPolyStatus represents the state of smooth polygon generation
type SmoothPolyStatus int

const (
	SmoothPolyInitial SmoothPolyStatus = iota
	SmoothPolyReady
	SmoothPolyPolygon
	SmoothPolyCtrlB
	SmoothPolyCtrlE
	SmoothPolyCtrl1
	SmoothPolyCtrl2
	SmoothPolyEndPoly
	SmoothPolyStop
)

// VCGenSmoothPoly1 generates vertices for smooth polygon corners using cubic Bezier curves
type VCGenSmoothPoly1 struct {
	srcVertices *array.VertexCmdSequence
	smoothValue float64
	closed      bool
	status      SmoothPolyStatus
	srcVertex   int
	ctrl1X      float64
	ctrl1Y      float64
	ctrl2X      float64
	ctrl2Y      float64
}

// NewVCGenSmoothPoly1 creates a new smooth polygon vertex generator
func NewVCGenSmoothPoly1() *VCGenSmoothPoly1 {
	return &VCGenSmoothPoly1{
		srcVertices: array.NewVertexCmdSequence(),
		smoothValue: 0.5,
		closed:      false,
		status:      SmoothPolyInitial,
		srcVertex:   0,
	}
}

// SetSmoothValue sets the smoothing value (0.0 to 1.0)
// Higher values create more rounded corners
func (v *VCGenSmoothPoly1) SetSmoothValue(value float64) {
	v.smoothValue = value * 0.5
}

// SmoothValue returns the current smooth value
func (v *VCGenSmoothPoly1) SmoothValue() float64 {
	return v.smoothValue * 2.0
}

// RemoveAll clears all vertices
func (v *VCGenSmoothPoly1) RemoveAll() {
	v.srcVertices.RemoveAll()
	v.closed = false
	v.status = SmoothPolyInitial
}

// AddVertex adds a vertex to the polygon
func (v *VCGenSmoothPoly1) AddVertex(x, y float64, cmd basics.PathCommand) {
	v.status = SmoothPolyInitial

	if basics.IsMoveTo(cmd) {
		v.srcVertices.ModifyLast(array.VertexDistCmd{X: x, Y: y, Dist: 0, Cmd: cmd})
	} else {
		if basics.IsVertex(cmd) {
			v.srcVertices.Add(array.VertexDistCmd{X: x, Y: y, Dist: 0, Cmd: cmd})
		} else {
			v.closed = basics.IsClosed(uint32(cmd))
		}
	}
}

// PrepareSrc prepares the smooth polygon for vertex generation
func (v *VCGenSmoothPoly1) PrepareSrc() {
	// This method is called by conv_adaptor_vcgen
}

// Rewind rewinds the smooth polygon generator
func (v *VCGenSmoothPoly1) Rewind(pathID uint) {
	if v.status == SmoothPolyInitial {
		v.srcVertices.Close(v.closed)
	}
	v.status = SmoothPolyReady
	v.srcVertex = 0
}

// Vertex returns the next vertex in the smoothed polygon
func (v *VCGenSmoothPoly1) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for {
		switch v.status {
		case SmoothPolyInitial:
			v.Rewind(0)
			continue

		case SmoothPolyReady:
			if v.srcVertices.Size() < 2 {
				return 0, 0, basics.PathCmdStop
			}

			if v.srcVertices.Size() == 2 {
				// Handle line segments (2 vertices) properly
				if v.srcVertex >= 2 {
					return 0, 0, basics.PathCmdStop
				}
				vertex := v.srcVertices.At(v.srcVertex)
				x, y = vertex.X, vertex.Y
				v.srcVertex++
				if v.srcVertex == 1 {
					return x, y, basics.PathCmdMoveTo
				}
				if v.srcVertex == 2 {
					return x, y, basics.PathCmdLineTo
				}
				return 0, 0, basics.PathCmdStop
			}

			cmd = basics.PathCmdMoveTo
			v.status = SmoothPolyPolygon
			v.srcVertex = 0
			continue

		case SmoothPolyPolygon:
			if v.closed {
				if v.srcVertex >= v.srcVertices.Size() {
					vertex := v.srcVertices.At(0)
					x, y = vertex.X, vertex.Y
					v.status = SmoothPolyEndPoly
					return x, y, basics.PathCmdCurve4
				}
			} else {
				if v.srcVertex >= v.srcVertices.Size()-1 {
					vertex := v.srcVertices.At(v.srcVertices.Size() - 1)
					x, y = vertex.X, vertex.Y
					v.status = SmoothPolyEndPoly
					return x, y, basics.PathCmdLineTo
				}
			}

			v.calculate(v.srcVertex)
			vertex := v.srcVertices.At(v.srcVertex)
			x, y = vertex.X, vertex.Y
			v.srcVertex++

			if v.srcVertex == 1 {
				return x, y, basics.PathCmdMoveTo
			} else {
				v.status = SmoothPolyCtrl1
				return x, y, basics.PathCmdCurve4
			}

		case SmoothPolyCtrl1:
			x, y = v.ctrl1X, v.ctrl1Y
			v.status = SmoothPolyCtrl2
			return x, y, basics.PathCmdCurve4

		case SmoothPolyCtrl2:
			x, y = v.ctrl2X, v.ctrl2Y
			v.status = SmoothPolyPolygon
			return x, y, basics.PathCmdCurve4

		case SmoothPolyEndPoly:
			v.status = SmoothPolyStop
			if v.closed {
				return 0, 0, basics.PathCmdEndPoly | basics.PathFlagClose
			} else {
				return 0, 0, basics.PathCmdEndPoly
			}

		case SmoothPolyStop:
			return 0, 0, basics.PathCmdStop

		default:
			return 0, 0, basics.PathCmdStop
		}
	}
}

// calculate computes the control points for smooth corners
func (v *VCGenSmoothPoly1) calculate(idx int) {
	size := v.srcVertices.Size()

	// TODO: Fix smooth poly generator bounds checking and vertex sequence access
	// Issue: When accessing vertices with modulo arithmetic, need to ensure we have enough points
	// for smoothing calculation (need at least 4 points for proper cubic Bezier)
	if size < 3 {
		// Not enough points for proper smoothing - use simple control points
		v.ctrl1X = 0
		v.ctrl1Y = 0
		v.ctrl2X = 0
		v.ctrl2Y = 0
		return
	}

	// Get four consecutive vertices for smoothing calculation
	v0 := v.srcVertices.At((idx - 1 + size) % size)
	v1 := v.srcVertices.At(idx)
	v2 := v.srcVertices.At((idx + 1) % size)
	v3 := v.srcVertices.At((idx + 2) % size)

	// Calculate distances
	dist01 := basics.Sqrt((v1.X-v0.X)*(v1.X-v0.X) + (v1.Y-v0.Y)*(v1.Y-v0.Y))
	dist12 := basics.Sqrt((v2.X-v1.X)*(v2.X-v1.X) + (v2.Y-v1.Y)*(v2.Y-v1.Y))
	dist23 := basics.Sqrt((v3.X-v2.X)*(v3.X-v2.X) + (v3.Y-v2.Y)*(v3.Y-v2.Y))

	if dist01 < basics.VertexDistEpsilon {
		dist01 = 1.0
	}
	if dist12 < basics.VertexDistEpsilon {
		dist12 = 1.0
	}
	if dist23 < basics.VertexDistEpsilon {
		dist23 = 1.0
	}

	k1 := dist01 / (dist01 + dist12)
	k2 := dist12 / (dist12 + dist23)

	xm1 := v0.X + (v2.X-v0.X)*k1
	ym1 := v0.Y + (v2.Y-v0.Y)*k1
	xm2 := v1.X + (v3.X-v1.X)*k2
	ym2 := v1.Y + (v3.Y-v1.Y)*k2

	v.ctrl1X = v1.X + v.smoothValue*(v2.X-xm1)
	v.ctrl1Y = v1.Y + v.smoothValue*(v2.Y-ym1)
	v.ctrl2X = v2.X + v.smoothValue*(v1.X-xm2)
	v.ctrl2Y = v2.Y + v.smoothValue*(v1.Y-ym2)
}
