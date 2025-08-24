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
	SmoothPolyVertex
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
					// Generate control points for the final segment back to start
					v.calculate(v.srcVertex - 1) // Calculate for the last actual vertex
					vertex := v.srcVertices.At(0)
					x, y = vertex.X, vertex.Y
					v.srcVertex++              // Increment to prevent re-entry
					v.status = SmoothPolyCtrl1 // Go through control point states
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
				// For subsequent vertices, first output the control points, then the vertex as LineTo
				v.status = SmoothPolyCtrl1
				// Don't return the vertex as Curve4, return it after control points
				continue
			}

		case SmoothPolyCtrl1:
			x, y = v.ctrl1X, v.ctrl1Y
			v.status = SmoothPolyCtrl2
			return x, y, basics.PathCmdCurve4

		case SmoothPolyCtrl2:
			x, y = v.ctrl2X, v.ctrl2Y
			v.status = SmoothPolyVertex
			return x, y, basics.PathCmdCurve4

		case SmoothPolyVertex:
			// Return the vertex position that was calculated in SmoothPolyPolygon
			if v.closed && v.srcVertex > v.srcVertices.Size() {
				// This is the final segment return, go to EndPoly
				vertex := v.srcVertices.At(0) // Return to start vertex
				x, y = vertex.X, vertex.Y
				v.status = SmoothPolyEndPoly
				return x, y, basics.PathCmdCurve4
			} else {
				vertex := v.srcVertices.At(v.srcVertex - 1) // srcVertex was already incremented
				x, y = vertex.X, vertex.Y
				v.status = SmoothPolyPolygon
				return x, y, basics.PathCmdLineTo
			}

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

	if size < 3 {
		// Not enough points for proper smoothing
		v.ctrl1X = 0
		v.ctrl1Y = 0
		v.ctrl2X = 0
		v.ctrl2Y = 0
		return
	}

	// Get four consecutive vertices for smoothing calculation
	// matching the C++ prev/curr/next/next logic
	var v0, v1, v2, v3 array.VertexDistCmd

	// v0 = prev(idx)
	if idx == 0 {
		if v.closed && size > 1 {
			v0 = v.srcVertices.At(size - 1)
		} else {
			v0 = v.srcVertices.At(0) // Use same vertex if no previous
		}
	} else {
		v0 = v.srcVertices.At(idx - 1)
	}

	// v1 = curr(idx)
	v1 = v.srcVertices.At(idx)

	// v2 = next(idx)
	if idx >= size-1 {
		if v.closed {
			v2 = v.srcVertices.At(0)
		} else {
			v2 = v.srcVertices.At(size - 1) // Use same vertex if no next
		}
	} else {
		v2 = v.srcVertices.At(idx + 1)
	}

	// v3 = next(idx + 1)
	if idx >= size-2 {
		if v.closed {
			if size > 1 {
				v3 = v.srcVertices.At((idx + 2) % size)
			} else {
				v3 = v2 // Fallback
			}
		} else {
			v3 = v2 // Use same as v2 if no next-next
		}
	} else {
		v3 = v.srcVertices.At(idx + 2)
	}

	// Use the distances from the vertex_dist structures (C++ version uses v0.dist, v1.dist, v2.dist)
	// The C++ algorithm uses the distance to the NEXT vertex, stored in the vertex_dist.dist field
	// Ensure distances are valid and non-zero
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
