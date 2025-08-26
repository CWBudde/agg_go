package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Status represents the state of the stroke generator
type Status int

const (
	Initial Status = iota
	Ready
	Cap1
	Cap2
	Outline1
	CloseFirst
	Outline2
	OutVertices
	EndPoly1
	EndPoly2
	Stop
)

// VCGenStroke generates stroke vertices from input path vertices
type VCGenStroke struct {
	stroker     *basics.MathStroke
	srcVertices *array.VertexSequence[basics.VertexDist]
	outVertices *array.PodBVector[basics.PointD]
	shorten     float64
	closed      bool
	status      Status
	prevStatus  Status
	srcVertex   int
	outVertex   int
}

// NewVCGenStroke creates a new stroke vertex generator
func NewVCGenStroke() *VCGenStroke {
	return &VCGenStroke{
		stroker:     basics.NewMathStroke(),
		srcVertices: array.NewVertexSequence[basics.VertexDist](),
		outVertices: array.NewPodBVector[basics.PointD](),
		shorten:     0.0,
		closed:      false,
		status:      Initial,
		srcVertex:   0,
		outVertex:   0,
	}
}

// RemoveAll clears all vertices and resets the generator
func (vg *VCGenStroke) RemoveAll() {
	vg.srcVertices.RemoveAll()
	vg.closed = false
	vg.status = Initial
}

// AddVertex adds a vertex to the stroke path
func (vg *VCGenStroke) AddVertex(x, y float64, cmd basics.PathCommand) {
	vg.status = Initial

	if basics.IsMoveTo(cmd) {
		vg.srcVertices.ModifyLast(basics.VertexDist{X: x, Y: y})
	} else if basics.IsVertex(cmd) {
		vg.srcVertices.Add(basics.VertexDist{X: x, Y: y})
	} else {
		vg.closed = basics.GetCloseFlag(uint32(cmd)) != 0
	}
}

// PrepareSrc prepares the source vertices for stroke generation
func (vg *VCGenStroke) PrepareSrc() {
	// This method is called before starting vertex generation
	// Similar to what rewind() does in the original
	if vg.status == Initial {
		vg.srcVertices.Close(vg.closed)
		vg.shortenPath()
		if vg.srcVertices.Size() < 3 {
			vg.closed = false
		}
	}
	vg.status = Ready
	vg.srcVertex = 0
	vg.outVertex = 0
}

// shortenPath implements basic path shortening (simplified version)
func (vg *VCGenStroke) shortenPath() {
	// VERIFIED: This aligns with the original C++ implementation in agg_vcgen_stroke.cpp
	// The C++ version also calls shorten_path(m_src_vertices, m_shorten, m_closed) in rewind().
	// Use the shared shorten_path implementation to trim the polyline.
	// This matches AGG's agg_shorten_path behavior and mirrors vcgen_dash.
	if vg.shorten > 0.0 && vg.srcVertices.Size() > 1 {
		// Convert basics.VertexDist to array.VertexDist for ShortenPath
		convertedVertices := array.NewVertexSequence[array.VertexDist]()
		for i := 0; i < vg.srcVertices.Size(); i++ {
			v := vg.srcVertices.At(i)
			convertedVertices.Add(array.VertexDist{X: v.X, Y: v.Y, Dist: v.Dist})
		}

		array.ShortenPath(convertedVertices, vg.shorten, vg.closed)

		// Convert back to basics.VertexDist
		vg.srcVertices.RemoveAll()
		for i := 0; i < convertedVertices.Size(); i++ {
			v := convertedVertices.At(i)
			vg.srcVertices.Add(basics.VertexDist{X: v.X, Y: v.Y, Dist: v.Dist})
		}
	}
}

// Rewind resets the generator for vertex iteration
func (vg *VCGenStroke) Rewind(pathID uint) {
	vg.PrepareSrc()
}

// Vertex returns the next vertex in the stroke outline
func (vg *VCGenStroke) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	for !basics.IsStop(cmd) {
		switch vg.status {
		case Initial:
			vg.PrepareSrc()

		case Ready:
			minVertices := 2
			if vg.closed {
				minVertices = 3
			}
			if vg.srcVertices.Size() < minVertices {
				cmd = basics.PathCmdStop
				break
			}

			if vg.closed {
				vg.status = Outline1
			} else {
				vg.status = Cap1
			}
			cmd = basics.PathCmdMoveTo
			vg.srcVertex = 0
			vg.outVertex = 0

		case Cap1:
			// Calculate start cap
			v0 := vg.srcVertices.At(0)
			v1 := vg.srcVertices.At(1)
			consumer := array.NewPodBVectorConsumer(vg.outVertices)
			vg.stroker.CalcCap(consumer, v0, v1, v0.Dist)
			vg.srcVertex = 1
			vg.prevStatus = Outline1
			vg.status = OutVertices
			vg.outVertex = 0

		case Cap2:
			// Calculate end cap
			size := vg.srcVertices.Size()
			v0 := vg.srcVertices.At(size - 1)
			v1 := vg.srcVertices.At(size - 2)
			consumer := array.NewPodBVectorConsumer(vg.outVertices)
			vg.stroker.CalcCap(consumer, v0, v1, v1.Dist)
			vg.prevStatus = Outline2
			vg.status = OutVertices
			vg.outVertex = 0

		case Outline1:
			if vg.closed {
				if vg.srcVertex >= vg.srcVertices.Size() {
					vg.prevStatus = CloseFirst
					vg.status = EndPoly1
					break
				}
			} else {
				if vg.srcVertex >= vg.srcVertices.Size()-1 {
					vg.status = Cap2
					break
				}
			}

			// Calculate join
			v0 := vg.getPrev(vg.srcVertex)
			v1 := vg.getCurr(vg.srcVertex)
			v2 := vg.getNext(vg.srcVertex)
			consumer := array.NewPodBVectorConsumer(vg.outVertices)
			vg.stroker.CalcJoin(consumer, v0, v1, v2, v0.Dist, v1.Dist)

			vg.srcVertex++
			vg.prevStatus = vg.status
			vg.status = OutVertices
			vg.outVertex = 0

		case CloseFirst:
			vg.status = Outline2
			cmd = basics.PathCmdMoveTo

		case Outline2:
			if vg.srcVertex <= 0 || (!vg.closed && vg.srcVertex <= 1) {
				vg.status = EndPoly2
				vg.prevStatus = Stop
				break
			}

			vg.srcVertex--
			v0 := vg.getNext(vg.srcVertex)
			v1 := vg.getCurr(vg.srcVertex)
			v2 := vg.getPrev(vg.srcVertex)
			consumer := array.NewPodBVectorConsumer(vg.outVertices)
			vg.stroker.CalcJoin(consumer, v0, v1, v2, v1.Dist, v2.Dist)

			vg.prevStatus = vg.status
			vg.status = OutVertices
			vg.outVertex = 0

		case OutVertices:
			if vg.outVertex >= vg.outVertices.Size() {
				vg.status = vg.prevStatus
			} else {
				point := vg.outVertices.At(vg.outVertex)
				vg.outVertex++
				x = point.X
				y = point.Y
				return x, y, cmd
			}

		case EndPoly1:
			vg.status = vg.prevStatus
			return 0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))

		case EndPoly2:
			vg.status = vg.prevStatus
			return 0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCW))

		case Stop:
			cmd = basics.PathCmdStop
		}
	}

	return 0, 0, cmd
}

// Helper methods for vertex access with wrapping
func (vg *VCGenStroke) getPrev(idx int) basics.VertexDist {
	if idx == 0 {
		if vg.closed && vg.srcVertices.Size() > 1 {
			return vg.srcVertices.At(vg.srcVertices.Size() - 1)
		}
		return vg.srcVertices.At(0)
	}
	return vg.srcVertices.At(idx - 1)
}

func (vg *VCGenStroke) getCurr(idx int) basics.VertexDist {
	return vg.srcVertices.At(idx)
}

func (vg *VCGenStroke) getNext(idx int) basics.VertexDist {
	if idx >= vg.srcVertices.Size()-1 {
		if vg.closed && vg.srcVertices.Size() > 1 {
			return vg.srcVertices.At(0)
		}
		return vg.srcVertices.At(vg.srcVertices.Size() - 1)
	}
	return vg.srcVertices.At(idx + 1)
}

// Stroke parameter delegation methods
func (vg *VCGenStroke) SetLineCap(lc basics.LineCap) {
	vg.stroker.SetLineCap(lc)
}

func (vg *VCGenStroke) LineCap() basics.LineCap {
	return vg.stroker.LineCap()
}

func (vg *VCGenStroke) SetLineJoin(lj basics.LineJoin) {
	vg.stroker.SetLineJoin(lj)
}

func (vg *VCGenStroke) LineJoin() basics.LineJoin {
	return vg.stroker.LineJoin()
}

func (vg *VCGenStroke) SetInnerJoin(ij basics.InnerJoin) {
	vg.stroker.SetInnerJoin(ij)
}

func (vg *VCGenStroke) InnerJoin() basics.InnerJoin {
	return vg.stroker.InnerJoin()
}

func (vg *VCGenStroke) SetWidth(w float64) {
	vg.stroker.SetWidth(w)
}

func (vg *VCGenStroke) Width() float64 {
	return vg.stroker.Width()
}

func (vg *VCGenStroke) SetMiterLimit(ml float64) {
	vg.stroker.SetMiterLimit(ml)
}

func (vg *VCGenStroke) MiterLimit() float64 {
	return vg.stroker.MiterLimit()
}

func (vg *VCGenStroke) SetMiterLimitTheta(t float64) {
	vg.stroker.SetMiterLimitTheta(t)
}

func (vg *VCGenStroke) SetInnerMiterLimit(ml float64) {
	vg.stroker.SetInnerMiterLimit(ml)
}

func (vg *VCGenStroke) InnerMiterLimit() float64 {
	return vg.stroker.InnerMiterLimit()
}

func (vg *VCGenStroke) SetApproximationScale(as float64) {
	vg.stroker.SetApproximationScale(as)
}

func (vg *VCGenStroke) ApproximationScale() float64 {
	return vg.stroker.ApproximationScale()
}

func (vg *VCGenStroke) SetShorten(s float64) {
	vg.shorten = s
}

func (vg *VCGenStroke) Shorten() float64 {
	return vg.shorten
}
