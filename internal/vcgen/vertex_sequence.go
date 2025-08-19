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

// shortenPath shortens the path by the specified amount from the END only
// This matches the C++ AGG implementation exactly
func (v *VCGenVertexSequence) shortenPath() {
	if v.shorten <= 0.0 || v.srcVertices.Size() <= 1 {
		return
	}

	// First, calculate distances between consecutive vertices
	// This stores the distance from the previous vertex to the current one
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
	}

	// Shorten from the end by removing segments
	s := v.shorten
	n := v.srcVertices.Size() - 2 // Start from second-to-last vertex

	// Remove complete segments from the end
	for n >= 0 && v.srcVertices.Size() > 1 {
		vertex := v.srcVertices.At(n + 1) // The vertex at position n+1
		d := vertex.Dist
		if d > s {
			break // This segment is longer than remaining shortening distance
		}
		v.srcVertices.RemoveLast()
		s -= d
		n--
	}

	// Check if we removed everything except the first vertex
	if v.srcVertices.Size() < 2 {
		v.srcVertices.RemoveAll()
		return
	}

	// Interpolate the final vertex position if we have remaining shortening distance
	if s > 0 {
		n = v.srcVertices.Size() - 1
		prev := v.srcVertices.At(n - 1)
		last := v.srcVertices.At(n)

		// Calculate the interpolation factor
		// s is the remaining distance to shorten
		// last.Dist is the total distance of the last segment
		d := (last.Dist - s) / last.Dist

		// Interpolate position along the segment
		x := prev.X + (last.X-prev.X)*d
		y := prev.Y + (last.Y-prev.Y)*d

		// Update the last vertex position
		v.srcVertices.ModifyAt(n, array.VertexDistCmd{
			X:    x,
			Y:    y,
			Dist: last.Dist,
			Cmd:  last.Cmd,
		})
	}
}
