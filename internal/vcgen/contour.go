package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// status_e represents the state machine states for contour generation
type contourStatus int

const (
	contourInitial contourStatus = iota
	contourReady
	contourOutline
	contourOutVertices
	contourEndPoly
	contourStop
)

// pointSliceConsumer implements VertexConsumer for []basics.Point[float64]
type pointSliceConsumer struct {
	points *[]basics.Point[float64]
}

func (p *pointSliceConsumer) Add(x, y float64) {
	*p.points = append(*p.points, basics.Point[float64]{X: x, Y: y})
}

func (p *pointSliceConsumer) RemoveAll() {
	*p.points = (*p.points)[:0]
}

// VCGenContour generates contour vertices from path data.
// This is a port of AGG's vcgen_contour class which creates offset curves
// (parallel lines) at a fixed distance from the original path.
type VCGenContour struct {
	stroker     *basics.MathStroke
	width       float64
	srcVertices *array.VertexDistSequence
	outVertices []basics.Point[float64]
	consumer    *pointSliceConsumer
	status      contourStatus
	srcVertex   int
	outVertex   int
	closed      uint32
	orientation uint32
	autoDetect  bool
}

// NewVCGenContour creates a new contour vertex generator
func NewVCGenContour() *VCGenContour {
	outVertices := make([]basics.Point[float64], 0, 32)
	vc := &VCGenContour{
		stroker:     basics.NewMathStroke(),
		width:       1.0,
		srcVertices: array.NewVertexDistSequence(),
		outVertices: outVertices,
		status:      contourInitial,
		srcVertex:   0,
		outVertex:   0,
		closed:      0,
		orientation: 0,
		autoDetect:  false,
	}
	vc.consumer = &pointSliceConsumer{points: &vc.outVertices}
	return vc
}

// RemoveAll clears all vertices and resets the state
func (vc *VCGenContour) RemoveAll() {
	vc.srcVertices.RemoveAll()
	vc.closed = 0
	vc.orientation = 0
	vc.status = contourInitial
}

// AddVertex adds a vertex to the contour generator
func (vc *VCGenContour) AddVertex(x, y float64, cmd basics.PathCommand) {
	vc.status = contourInitial

	if basics.IsMoveTo(cmd) {
		// For MoveTo, start a new path by adding the vertex
		vc.srcVertices.Add(basics.VertexDist{X: x, Y: y, Dist: 0})
	} else if basics.IsVertex(cmd) {
		vc.srcVertices.Add(basics.VertexDist{X: x, Y: y, Dist: 0})
	} else if basics.IsEndPoly(cmd) {
		vc.closed = basics.GetCloseFlag(uint32(cmd))
		if vc.orientation == uint32(basics.PathFlagsNone) {
			vc.orientation = basics.GetOrientation(uint32(cmd))
		}
	}
}

// PrepareSrc prepares the source vertices for processing
func (vc *VCGenContour) PrepareSrc() {
	// This method is for compatibility with VertexGenerator interface
	// Actual preparation happens in Rewind
}

// Rewind prepares for vertex iteration
func (vc *VCGenContour) Rewind(pathID uint) {
	if vc.status == contourInitial {
		vc.srcVertices.Close(true)

		if vc.autoDetect {
			if !basics.IsOriented(vc.orientation) {
				// Calculate polygon area to determine orientation
				vertices := make([]basics.Point[float64], vc.srcVertices.Size())
				for i := 0; i < vc.srcVertices.Size(); i++ {
					v := vc.srcVertices.Get(i)
					vertices[i] = basics.Point[float64]{X: v.X, Y: v.Y}
				}
				area := basics.CalcPolygonArea(vertices)
				if area > 0.0 {
					vc.orientation = uint32(basics.PathFlagsCCW)
				} else {
					vc.orientation = uint32(basics.PathFlagsCW)
				}
			}
		}

		if basics.IsOriented(vc.orientation) {
			if basics.IsCCW(vc.orientation) {
				vc.stroker.SetWidth(vc.width)
			} else {
				vc.stroker.SetWidth(-vc.width)
			}
		}
	}
	vc.status = contourReady
	vc.srcVertex = 0
}

// Vertex returns the next vertex in the contour
func (vc *VCGenContour) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for !basics.IsStop(cmd) {
		switch vc.status {
		case contourInitial:
			vc.Rewind(0)

		case contourReady:
			if vc.srcVertices.Size() < 2+int(vc.closed) {
				cmd = basics.PathCmdStop
				break
			}
			vc.status = contourOutline
			cmd = basics.PathCmdMoveTo
			vc.srcVertex = 0
			vc.outVertex = 0

		case contourOutline:
			if vc.srcVertex >= vc.srcVertices.Size() {
				vc.status = contourEndPoly
				break
			}

			// Calculate join for current vertex
			vc.outVertices = vc.outVertices[:0] // Clear the slice but keep capacity
			prev := vc.prev(vc.srcVertex)
			curr := vc.curr(vc.srcVertex)
			next := vc.next(vc.srcVertex)

			// Ensure distances are calculated
			vc.srcVertices.CalculateDistances()

			// Update prev/curr/next after distance calculation
			prev = vc.prev(vc.srcVertex)
			curr = vc.curr(vc.srcVertex)
			next = vc.next(vc.srcVertex)

			vc.stroker.CalcJoin(
				vc.consumer,
				prev,
				curr,
				next,
				prev.Dist,
				curr.Dist,
			)
			vc.srcVertex++
			vc.status = contourOutVertices
			vc.outVertex = 0

		case contourOutVertices:
			if vc.outVertex >= len(vc.outVertices) {
				vc.status = contourOutline
			} else {
				pt := vc.outVertices[vc.outVertex]
				vc.outVertex++
				x = pt.X
				y = pt.Y
				return x, y, cmd
			}

		case contourEndPoly:
			if vc.closed == 0 {
				return 0, 0, basics.PathCmdStop
			}
			vc.status = contourStop
			return 0, 0, basics.PathCmdEndPoly | basics.PathFlagClose | basics.PathCommand(basics.PathFlagsCCW)

		case contourStop:
			return 0, 0, basics.PathCmdStop
		}
	}
	return 0, 0, cmd
}

// Width sets the contour width
func (vc *VCGenContour) Width(w float64) {
	vc.width = w
	vc.stroker.SetWidth(w)
}

// GetWidth returns the current contour width
func (vc *VCGenContour) GetWidth() float64 {
	return vc.width
}

// LineJoin sets the line join style
func (vc *VCGenContour) LineJoin(lj basics.LineJoin) {
	vc.stroker.SetLineJoin(lj)
}

// GetLineJoin returns the current line join style
func (vc *VCGenContour) GetLineJoin() basics.LineJoin {
	return vc.stroker.LineJoin()
}

// InnerJoin sets the inner join style
func (vc *VCGenContour) InnerJoin(ij basics.InnerJoin) {
	vc.stroker.SetInnerJoin(ij)
}

// GetInnerJoin returns the current inner join style
func (vc *VCGenContour) GetInnerJoin() basics.InnerJoin {
	return vc.stroker.InnerJoin()
}

// MiterLimit sets the miter limit
func (vc *VCGenContour) MiterLimit(ml float64) {
	vc.stroker.SetMiterLimit(ml)
}

// GetMiterLimit returns the current miter limit
func (vc *VCGenContour) GetMiterLimit() float64 {
	return vc.stroker.MiterLimit()
}

// MiterLimitTheta sets the miter limit by angle
func (vc *VCGenContour) MiterLimitTheta(t float64) {
	vc.stroker.SetMiterLimitTheta(t)
}

// InnerMiterLimit sets the inner miter limit
func (vc *VCGenContour) InnerMiterLimit(ml float64) {
	vc.stroker.SetInnerMiterLimit(ml)
}

// GetInnerMiterLimit returns the current inner miter limit
func (vc *VCGenContour) GetInnerMiterLimit() float64 {
	return vc.stroker.InnerMiterLimit()
}

// ApproximationScale sets the approximation scale
func (vc *VCGenContour) ApproximationScale(as float64) {
	vc.stroker.SetApproximationScale(as)
}

// GetApproximationScale returns the current approximation scale
func (vc *VCGenContour) GetApproximationScale() float64 {
	return vc.stroker.ApproximationScale()
}

// AutoDetectOrientation sets whether to automatically detect polygon orientation
func (vc *VCGenContour) AutoDetectOrientation(v bool) {
	vc.autoDetect = v
}

// GetAutoDetectOrientation returns the current auto-detect setting
func (vc *VCGenContour) GetAutoDetectOrientation() bool {
	return vc.autoDetect
}

// Helper methods for vertex access
func (vc *VCGenContour) prev(idx int) basics.VertexDist {
	size := vc.srcVertices.Size()
	if idx == 0 {
		return vc.srcVertices.At(size - 1)
	}
	return vc.srcVertices.At(idx - 1)
}

func (vc *VCGenContour) curr(idx int) basics.VertexDist {
	return vc.srcVertices.At(idx)
}

func (vc *VCGenContour) next(idx int) basics.VertexDist {
	size := vc.srcVertices.Size()
	if idx == size-1 {
		return vc.srcVertices.At(0)
	}
	return vc.srcVertices.At(idx + 1)
}
