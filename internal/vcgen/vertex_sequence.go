package vcgen

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// VCGenVertexSequence is a vertex generator that stores and replays a sequence of vertices
type VCGenVertexSequence struct {
	srcVertices *array.VertexCmdSequence
	flags       basics.PathCommand
	curVertex   int
	shorten     float64
	ready       bool
}

// NewVCGenVertexSequence creates a new vertex sequence generator
func NewVCGenVertexSequence() *VCGenVertexSequence {
	return &VCGenVertexSequence{
		srcVertices: array.NewVertexCmdSequence(),
		flags:       0,
		curVertex:   0,
		shorten:     0.0,
		ready:       false,
	}
}

// RemoveAll clears all vertices from the sequence
func (v *VCGenVertexSequence) RemoveAll() {
	v.ready = false
	v.srcVertices.RemoveAll()
	v.curVertex = 0
	v.flags = 0
}

// AddVertex adds a vertex to the sequence
func (v *VCGenVertexSequence) AddVertex(x, y float64, cmd basics.PathCommand) {
	v.ready = false

	if basics.IsMoveTo(cmd) {
		v.srcVertices.ModifyLast(array.VertexDistCmd{X: x, Y: y, Dist: 0, Cmd: cmd})
	} else {
		if basics.IsVertex(cmd) {
			v.srcVertices.Add(array.VertexDistCmd{X: x, Y: y, Dist: 0, Cmd: cmd})
		} else {
			v.flags = basics.PathCommand(uint32(cmd) & uint32(basics.PathFlagsMask))
		}
	}
}

// PrepareSrc prepares the vertex sequence for output
func (v *VCGenVertexSequence) PrepareSrc() {
	if !v.ready {
		v.srcVertices.Close(basics.IsClosed(uint32(v.flags)))
		if v.shorten > 0.0 {
			v.shortenPath()
		}
		v.ready = true
	}
}

// Rewind rewinds the vertex sequence iterator
func (v *VCGenVertexSequence) Rewind(pathID uint) {
	v.PrepareSrc()
	v.curVertex = 0
}

// Vertex returns the next vertex in the sequence
func (v *VCGenVertexSequence) Vertex() (x, y float64, cmd basics.PathCommand) {
	if !v.ready {
		v.PrepareSrc()
	}

	if v.curVertex >= v.srcVertices.Size() {
		return 0, 0, basics.PathCmdStop
	}

	vertex := v.srcVertices.At(v.curVertex)
	v.curVertex++

	return vertex.X, vertex.Y, vertex.Cmd
}

// SetShorten sets the path shortening distance
func (v *VCGenVertexSequence) SetShorten(s float64) {
	v.shorten = s
	v.ready = false
}

// Shorten returns the current path shortening distance
func (v *VCGenVertexSequence) Shorten() float64 {
	return v.shorten
}

// shortenPath shortens the path by the specified amount
func (v *VCGenVertexSequence) shortenPath() {
	if v.srcVertices.Size() <= 1 || v.shorten <= 0.0 {
		return
	}

	// Calculate total path length
	totalLen := 0.0
	for i := 1; i < v.srcVertices.Size(); i++ {
		v1 := v.srcVertices.At(i - 1)
		v2 := v.srcVertices.At(i)
		dx := v2.X - v1.X
		dy := v2.Y - v1.Y
		dist := basics.Sqrt(dx*dx + dy*dy)
		v.srcVertices.ModifyAt(i, array.VertexDistCmd{
			X:    v2.X,
			Y:    v2.Y,
			Dist: dist,
			Cmd:  v2.Cmd,
		})
		totalLen += dist
	}

	if totalLen < v.shorten {
		v.srcVertices.RemoveAll()
		return
	}

	// Shorten from both ends
	shortenPerEnd := v.shorten / 2.0

	// Shorten from start
	accLen := 0.0
	startIdx := 1
	for i := 1; i < v.srcVertices.Size() && accLen < shortenPerEnd; i++ {
		vertex := v.srcVertices.At(i)
		accLen += vertex.Dist
		startIdx = i
	}

	if accLen > shortenPerEnd && startIdx > 1 {
		// Interpolate the start vertex
		vertex := v.srcVertices.At(startIdx)
		prevVertex := v.srcVertices.At(startIdx - 1)

		excess := accLen - shortenPerEnd
		ratio := excess / vertex.Dist

		newX := vertex.X - (vertex.X-prevVertex.X)*ratio
		newY := vertex.Y - (vertex.Y-prevVertex.Y)*ratio

		v.srcVertices.ModifyAt(startIdx, array.VertexDistCmd{
			X:    newX,
			Y:    newY,
			Dist: excess,
			Cmd:  vertex.Cmd,
		})
	}

	// Remove vertices from start
	for i := 1; i < startIdx; i++ {
		v.srcVertices.RemoveAt(1)
	}

	if v.srcVertices.Size() <= 1 {
		return
	}

	// Shorten from end
	accLen = 0.0
	endIdx := v.srcVertices.Size() - 1
	for i := v.srcVertices.Size() - 1; i > 0 && accLen < shortenPerEnd; i-- {
		vertex := v.srcVertices.At(i)
		accLen += vertex.Dist
		endIdx = i
	}

	if accLen > shortenPerEnd && endIdx < v.srcVertices.Size()-1 {
		// Interpolate the end vertex
		vertex := v.srcVertices.At(endIdx)
		nextVertex := v.srcVertices.At(endIdx + 1)

		excess := accLen - shortenPerEnd
		ratio := excess / nextVertex.Dist

		newX := nextVertex.X - (nextVertex.X-vertex.X)*ratio
		newY := nextVertex.Y - (nextVertex.Y-vertex.Y)*ratio

		v.srcVertices.ModifyAt(endIdx, array.VertexDistCmd{
			X:    newX,
			Y:    newY,
			Dist: vertex.Dist,
			Cmd:  vertex.Cmd,
		})
	}

	// Remove vertices from end
	for i := endIdx + 1; i < v.srcVertices.Size(); {
		v.srcVertices.RemoveAt(i)
	}
}
