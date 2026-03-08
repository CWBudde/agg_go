package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// SmoothPolyStatus represents the state of smooth polygon generation.
// States mirror the C++ vcgen_smooth_poly1 exactly.
type SmoothPolyStatus int

const (
	SmoothPolyInitial SmoothPolyStatus = iota
	SmoothPolyReady
	SmoothPolyPolygon
	SmoothPolyCtrlB // beginning control point for open path (Curve3)
	SmoothPolyCtrlE // ending control point for open path (Curve3)
	SmoothPolyCtrl1 // first control point for interior segments (Curve4)
	SmoothPolyCtrl2 // second control point for interior segments (Curve4)
	SmoothPolyEndPoly
	SmoothPolyStop
)

// VCGenSmoothPoly1 generates smooth polygon corners using cubic/quadratic Bezier curves.
// Port of AGG's vcgen_smooth_poly1.
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

// NewVCGenSmoothPoly1 creates a new smooth polygon vertex generator.
func NewVCGenSmoothPoly1() *VCGenSmoothPoly1 {
	return &VCGenSmoothPoly1{
		srcVertices: array.NewVertexCmdSequence(),
		smoothValue: 0.5,
		closed:      false,
		status:      SmoothPolyInitial,
		srcVertex:   0,
	}
}

// SetSmoothValue sets the smoothing value.
// The C++ internal representation is value*0.5, so SetSmoothValue(1.0) → smoothValue=0.5.
func (v *VCGenSmoothPoly1) SetSmoothValue(value float64) {
	v.smoothValue = value * 0.5
}

// SmoothValue returns the external smooth value (multiply internal by 2).
func (v *VCGenSmoothPoly1) SmoothValue() float64 {
	return v.smoothValue * 2.0
}

// RemoveAll clears all vertices.
func (v *VCGenSmoothPoly1) RemoveAll() {
	v.srcVertices.RemoveAll()
	v.closed = false
	v.status = SmoothPolyInitial
}

// AddVertex adds a vertex to the polygon (called by ConvAdaptorVCGen).
func (v *VCGenSmoothPoly1) AddVertex(x, y float64, cmd basics.PathCommand) {
	v.status = SmoothPolyInitial

	switch {
	case basics.IsMoveTo(cmd):
		v.srcVertices.ModifyLast(array.VertexDistCmd{X: x, Y: y})
	case basics.IsVertex(cmd):
		v.srcVertices.Add(array.VertexDistCmd{X: x, Y: y})
	default:
		v.closed = basics.IsClosed(uint32(cmd))
	}
}

// PrepareSrc is called by ConvAdaptorVCGen before vertex generation starts.
func (v *VCGenSmoothPoly1) PrepareSrc() {}

// Rewind resets the generator. Distances are computed on the first rewind.
func (v *VCGenSmoothPoly1) Rewind(_ uint) {
	if v.status == SmoothPolyInitial {
		v.srcVertices.Close(v.closed)
	}
	v.status = SmoothPolyReady
	v.srcVertex = 0
}

// Vertex returns the next vertex, exactly matching the C++ state machine.
//
// For a closed polygon or an open path with ≥3 points, the emitted sequence is:
//
//	Open path (n points):
//	  MoveTo(P0),
//	  Curve3(ctrl2_for_P0_P1),          // ctrl_b: quadratic start
//	  Curve4(P1), Curve4(ctrl1), Curve4(ctrl2),   // interior cubics …
//	  …
//	  Curve3(Pk), Curve3(ctrl1_for_Pk_Pn-1),       // ctrl_e: quadratic end
//	  Curve3(Pn-1), EndPoly
//
//	Closed polygon:
//	  MoveTo(P0),
//	  Curve4(ctrl1), Curve4(ctrl2), Curve4(P1),
//	  …
//	  Curve4(ctrl1), Curve4(ctrl2), Curve4(P0), EndPoly|Close
func (v *VCGenSmoothPoly1) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for !basics.IsStop(cmd) {
		switch v.status {
		case SmoothPolyInitial:
			v.Rewind(0)
			continue // fall through to ready

		case SmoothPolyReady:
			n := v.srcVertices.Size()
			if n < 2 {
				return 0, 0, basics.PathCmdStop
			}
			if n == 2 {
				// Two-point degenerate: plain line segment
				if v.srcVertex >= n {
					return 0, 0, basics.PathCmdStop
				}
				vx := v.srcVertices.At(v.srcVertex)
				x, y = vx.X, vx.Y
				v.srcVertex++
				switch v.srcVertex {
				case 1:
					return x, y, basics.PathCmdMoveTo
				case 2:
					return x, y, basics.PathCmdLineTo
				}
				return 0, 0, basics.PathCmdStop
			}
			// ≥3 points: start the smooth polygon
			cmd = basics.PathCmdMoveTo
			v.status = SmoothPolyPolygon
			v.srcVertex = 0
			continue // fall through to polygon

		case SmoothPolyPolygon:
			n := v.srcVertices.Size()
			if v.closed {
				if v.srcVertex >= n {
					// Emit the closing vertex (back to P0) as Curve4 endpoint
					p0 := v.srcVertices.At(0)
					x, y = p0.X, p0.Y
					v.status = SmoothPolyEndPoly
					return x, y, basics.PathCmdCurve4
				}
			} else {
				if v.srcVertex >= n-1 {
					// Emit the final endpoint as Curve3 endpoint
					last := v.srcVertices.At(n - 1)
					x, y = last.X, last.Y
					v.status = SmoothPolyEndPoly
					return x, y, basics.PathCmdCurve3
				}
			}

			v.calculate(v.srcVertex)
			curr := v.srcVertices.At(v.srcVertex)
			x, y = curr.X, curr.Y
			v.srcVertex++

			if v.closed {
				v.status = SmoothPolyCtrl1
				if v.srcVertex == 1 {
					return x, y, basics.PathCmdMoveTo
				}
				return x, y, basics.PathCmdCurve4
			}
			// Open path boundary logic
			if v.srcVertex == 1 {
				v.status = SmoothPolyCtrlB
				return x, y, basics.PathCmdMoveTo
			}
			if v.srcVertex >= n-1 {
				v.status = SmoothPolyCtrlE
				return x, y, basics.PathCmdCurve3
			}
			v.status = SmoothPolyCtrl1
			return x, y, basics.PathCmdCurve4

		case SmoothPolyCtrlB:
			// Beginning quadratic control point (ctrl2 of first segment)
			x, y = v.ctrl2X, v.ctrl2Y
			v.status = SmoothPolyPolygon
			return x, y, basics.PathCmdCurve3

		case SmoothPolyCtrlE:
			// Ending quadratic control point (ctrl1 of last segment)
			x, y = v.ctrl1X, v.ctrl1Y
			v.status = SmoothPolyPolygon
			return x, y, basics.PathCmdCurve3

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
			}
			return 0, 0, basics.PathCmdEndPoly

		case SmoothPolyStop:
			return 0, 0, basics.PathCmdStop

		default:
			return 0, 0, basics.PathCmdStop
		}
	}
	return x, y, cmd
}

// calculate computes the Bezier control points for the segment ending at srcVertices[idx+1].
// Matches the C++ calculate(prev, curr, next, next+1).
func (v *VCGenSmoothPoly1) calculate(idx int) {
	n := v.srcVertices.Size()
	if n < 3 {
		v.ctrl1X, v.ctrl1Y = 0, 0
		v.ctrl2X, v.ctrl2Y = 0, 0
		return
	}

	var v0, v1, v2, v3 array.VertexDistCmd

	// v0 = prev(idx): clamp to 0 for open, wrap for closed
	if idx == 0 {
		if v.closed {
			v0 = v.srcVertices.At(n - 1)
		} else {
			v0 = v.srcVertices.At(0)
		}
	} else {
		v0 = v.srcVertices.At(idx - 1)
	}

	v1 = v.srcVertices.At(idx)

	// v2 = next(idx)
	if idx >= n-1 {
		if v.closed {
			v2 = v.srcVertices.At(0)
		} else {
			v2 = v.srcVertices.At(n - 1)
		}
	} else {
		v2 = v.srcVertices.At(idx + 1)
	}

	// v3 = next(idx+1)
	if idx >= n-2 {
		if v.closed {
			v3 = v.srcVertices.At((idx + 2) % n)
		} else {
			v3 = v.srcVertices.At(n - 1)
		}
	} else {
		v3 = v.srcVertices.At(idx + 2)
	}

	// Use stored distances (set by VertexCmdSequence.Close); fall back to computed.
	dist0 := v0.Dist
	dist1 := v1.Dist
	dist2 := v2.Dist

	if dist0 <= 0 {
		dist0 = basics.Sqrt((v1.X-v0.X)*(v1.X-v0.X) + (v1.Y-v0.Y)*(v1.Y-v0.Y))
		if dist0 < basics.VertexDistEpsilon {
			dist0 = 1.0
		}
	}
	if dist1 <= 0 {
		dist1 = basics.Sqrt((v2.X-v1.X)*(v2.X-v1.X) + (v2.Y-v1.Y)*(v2.Y-v1.Y))
		if dist1 < basics.VertexDistEpsilon {
			dist1 = 1.0
		}
	}
	if dist2 <= 0 {
		dist2 = basics.Sqrt((v3.X-v2.X)*(v3.X-v2.X) + (v3.Y-v2.Y)*(v3.Y-v2.Y))
		if dist2 < basics.VertexDistEpsilon {
			dist2 = 1.0
		}
	}

	k1 := dist0 / (dist0 + dist1)
	k2 := dist1 / (dist1 + dist2)

	xm1 := v0.X + (v2.X-v0.X)*k1
	ym1 := v0.Y + (v2.Y-v0.Y)*k1
	xm2 := v1.X + (v3.X-v1.X)*k2
	ym2 := v1.Y + (v3.Y-v1.Y)*k2

	v.ctrl1X = v1.X + v.smoothValue*(v2.X-xm1)
	v.ctrl1Y = v1.Y + v.smoothValue*(v2.Y-ym1)
	v.ctrl2X = v2.X + v.smoothValue*(v1.X-xm2)
	v.ctrl2Y = v2.Y + v.smoothValue*(v1.Y-ym2)
}
