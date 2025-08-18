// Package transform provides coordinate transformation functionality for AGG.
// This implements transformation along a single curved path.
package transform

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Status represents the state of path construction.
type Status int

const (
	StatusInitial Status = iota
	StatusMakingPath
	StatusReady
)

// VertexSource interface for adding paths from external sources.
// This corresponds to AGG's template parameter VertexSource in add_path.
type VertexSource interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// TransSinglePath transforms coordinates along a single curved path.
// This corresponds to AGG's trans_single_path class.
type TransSinglePath struct {
	// Source vertices with distance information (exported for testing)
	SrcVertices *array.VertexSequence[array.VertexDist]

	// Base length for scaling (0.0 means use calculated length)
	baseLength float64

	// Scaling factor for fast indexing when preserve_x_scale is false
	kindex float64

	// Current state of path construction (exported for testing)
	Status Status

	// Whether to preserve X scale (affects interpolation method)
	preserveXScale bool
}

// NewTransSinglePath creates a new single path transformation.
func NewTransSinglePath() *TransSinglePath {
	return &TransSinglePath{
		SrcVertices:    array.NewVertexSequenceWithScale[array.VertexDist](array.NewBlockScale(6)),
		baseLength:     0.0,
		kindex:         0.0,
		Status:         StatusInitial,
		preserveXScale: true,
	}
}

// SetBaseLength sets the base length for scaling transformations.
func (t *TransSinglePath) SetBaseLength(v float64) {
	t.baseLength = v
}

// BaseLength returns the current base length.
func (t *TransSinglePath) BaseLength() float64 {
	return t.baseLength
}

// SetPreserveXScale sets whether to preserve X scale in transformations.
func (t *TransSinglePath) SetPreserveXScale(f bool) {
	t.preserveXScale = f
}

// PreserveXScale returns whether X scale is preserved.
func (t *TransSinglePath) PreserveXScale() bool {
	return t.preserveXScale
}

// Reset clears the path and resets to initial state.
func (t *TransSinglePath) Reset() {
	t.SrcVertices.RemoveAll()
	t.kindex = 0.0
	t.Status = StatusInitial
}

// MoveTo starts a new path or adds a line to the current position.
func (t *TransSinglePath) MoveTo(x, y float64) {
	if t.Status == StatusInitial {
		// Add first vertex
		t.SrcVertices.Add(array.NewVertexDist(x, y))
		t.Status = StatusMakingPath
	} else {
		t.LineTo(x, y)
	}
}

// LineTo adds a line segment to the current path.
func (t *TransSinglePath) LineTo(x, y float64) {
	if t.Status == StatusMakingPath {
		t.SrcVertices.Add(array.NewVertexDist(x, y))
	}
}

// FinalizePath completes path construction and prepares for transformation.
func (t *TransSinglePath) FinalizePath() {
	if t.Status == StatusMakingPath && t.SrcVertices.Size() > 1 {
		// Close the sequence (not as a closed path)
		t.SrcVertices.Close(false)

		// Remove degenerate final segments (like in AGG)
		if t.SrcVertices.Size() > 2 {
			lastIdx := t.SrcVertices.Size() - 1
			secondLastIdx := t.SrcVertices.Size() - 2
			thirdLastIdx := t.SrcVertices.Size() - 3

			// Calculate distances for the last segments
			lastVertex := t.SrcVertices.At(lastIdx)
			secondLastVertex := t.SrcVertices.At(secondLastIdx)
			thirdLastVertex := t.SrcVertices.At(thirdLastIdx)

			// Calculate distance from second-to-last to last
			secondToLastDist := basics.CalcDistance(secondLastVertex.X, secondLastVertex.Y, lastVertex.X, lastVertex.Y)
			// Calculate distance from third-to-last to second-to-last
			thirdToSecondDist := basics.CalcDistance(thirdLastVertex.X, thirdLastVertex.Y, secondLastVertex.X, secondLastVertex.Y)

			// If the second-to-last segment is very short compared to the third-to-last
			if secondToLastDist*10.0 < thirdToSecondDist {
				// Move the last vertex to second-to-last position
				t.SrcVertices.Set(secondLastIdx, lastVertex)

				// Remove the actual last vertex
				t.SrcVertices.RemoveLast()

				// The combined distance will be calculated in the cumulative distance loop
			}
		}

		// Calculate cumulative distances like in AGG
		// Each vertex already has its distance to the next vertex from Close()
		totalDist := 0.0
		for i := 0; i < t.SrcVertices.Size(); i++ {
			v := t.SrcVertices.At(i)
			// Save the existing distance (to next vertex)
			d := v.Dist
			// Set cumulative distance
			v.Dist = totalDist
			t.SrcVertices.Set(i, v)
			// Add current segment length to total
			totalDist += d
		}

		// Calculate index scaling factor for fast lookup
		if totalDist > 0 {
			t.kindex = float64(t.SrcVertices.Size()-1) / totalDist
		}
		t.Status = StatusReady
	}
}

// AddPath adds vertices from a vertex source to the path.
// This is the template method equivalent from AGG.
func (t *TransSinglePath) AddPath(vs VertexSource, pathID uint) {
	vs.Rewind(pathID)

	for {
		x, y, cmd := vs.Vertex()
		if basics.IsStop(cmd) {
			break
		}

		if basics.IsMoveTo(cmd) {
			t.MoveTo(x, y)
		} else if basics.IsVertex(cmd) {
			t.LineTo(x, y)
		}
	}

	t.FinalizePath()
}

// TotalLength returns the total length of the path.
func (t *TransSinglePath) TotalLength() float64 {
	if t.baseLength >= 1e-10 {
		return t.baseLength
	}

	if t.Status == StatusReady && t.SrcVertices.Size() > 0 {
		lastIdx := t.SrcVertices.Size() - 1
		return t.SrcVertices.At(lastIdx).Dist
	}

	return 0.0
}

// Transform transforms coordinates along the path.
// The input x coordinate represents distance along the path,
// and y represents perpendicular offset from the path.
func (t *TransSinglePath) Transform(x, y *float64) {
	if t.Status != StatusReady {
		return
	}

	// Scale x coordinate if base length is set
	if t.baseLength > 1e-10 {
		pathLength := t.SrcVertices.At(t.SrcVertices.Size() - 1).Dist
		*x *= pathLength / t.baseLength
	}

	var x1, y1, dx, dy, d, dd float64

	switch {
	case *x < 0.0:
		// Extrapolation on the left
		v0 := t.SrcVertices.At(0)
		v1 := t.SrcVertices.At(1)

		x1 = v0.X
		y1 = v0.Y
		dx = v1.X - x1
		dy = v1.Y - y1
		dd = v1.Dist - v0.Dist
		d = *x
	case *x > t.SrcVertices.At(t.SrcVertices.Size()-1).Dist:
		// Extrapolation on the right
		lastIdx := t.SrcVertices.Size() - 1
		secondLastIdx := t.SrcVertices.Size() - 2

		vLast := t.SrcVertices.At(lastIdx)
		vSecondLast := t.SrcVertices.At(secondLastIdx)

		x1 = vLast.X
		y1 = vLast.Y
		dx = x1 - vSecondLast.X
		dy = y1 - vSecondLast.Y
		dd = vLast.Dist - vSecondLast.Dist
		d = *x - vLast.Dist
	default:
		// Interpolation
		var i, j int

		if t.preserveXScale {
			// Binary search for the segment
			i = 0
			j = t.SrcVertices.Size() - 1
			for (j - i) > 1 {
				k := (i + j) >> 1
				if *x < t.SrcVertices.At(k).Dist {
					j = k
				} else {
					i = k
				}
			}

			vi := t.SrcVertices.At(i)
			vj := t.SrcVertices.At(j)

			d = vi.Dist
			dd = vj.Dist - d
			d = *x - d
		} else {
			// Use uniform distribution with fast indexing
			i = int(*x * t.kindex)
			j = i + 1

			vi := t.SrcVertices.At(i)
			vj := t.SrcVertices.At(j)

			dd = vj.Dist - vi.Dist
			d = ((*x * t.kindex) - float64(i)) * dd
		}

		vi := t.SrcVertices.At(i)
		vj := t.SrcVertices.At(j)

		x1 = vi.X
		y1 = vi.Y
		dx = vj.X - x1
		dy = vj.Y - y1
	}

	// Calculate the point along the path
	x2 := x1 + dx*d/dd
	y2 := y1 + dy*d/dd

	// Apply perpendicular offset
	*x = x2 - *y*dy/dd
	*y = y2 + *y*dx/dd
}
