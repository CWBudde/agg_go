// Package transform provides coordinate transformation functionality for AGG.
// This implements transformation between two curved paths.
package transform

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// TransDoublePath transforms coordinates between two curved paths.
// This corresponds to AGG's trans_double_path class.
type TransDoublePath struct {
	// Source vertices for the two paths
	srcVertices1 *array.VertexSequence[array.VertexDist]
	srcVertices2 *array.VertexSequence[array.VertexDist]

	// Base dimensions for scaling
	baseLength float64
	baseHeight float64

	// Scaling factors for fast indexing
	kindex1 float64
	kindex2 float64

	// Current state of path construction
	status1 Status
	status2 Status

	// Whether to preserve X scale (affects interpolation method)
	preserveXScale bool
}

// NewTransDoublePath creates a new double path transformation.
func NewTransDoublePath() *TransDoublePath {
	return &TransDoublePath{
		srcVertices1:   array.NewVertexSequenceWithScale[array.VertexDist](array.NewBlockScale(6)),
		srcVertices2:   array.NewVertexSequenceWithScale[array.VertexDist](array.NewBlockScale(6)),
		baseLength:     0.0,
		baseHeight:     1.0,
		kindex1:        0.0,
		kindex2:        0.0,
		status1:        StatusInitial,
		status2:        StatusInitial,
		preserveXScale: true,
	}
}

// SetBaseLength sets the base length for scaling transformations.
func (t *TransDoublePath) SetBaseLength(v float64) {
	t.baseLength = v
}

// BaseLength returns the current base length.
func (t *TransDoublePath) BaseLength() float64 {
	return t.baseLength
}

// SetBaseHeight sets the base height for scaling transformations.
func (t *TransDoublePath) SetBaseHeight(v float64) {
	t.baseHeight = v
}

// BaseHeight returns the current base height.
func (t *TransDoublePath) BaseHeight() float64 {
	return t.baseHeight
}

// SetPreserveXScale sets whether to preserve X scale in transformations.
func (t *TransDoublePath) SetPreserveXScale(f bool) {
	t.preserveXScale = f
}

// PreserveXScale returns whether X scale is preserved.
func (t *TransDoublePath) PreserveXScale() bool {
	return t.preserveXScale
}

// Reset clears both paths and resets to initial state.
func (t *TransDoublePath) Reset() {
	t.srcVertices1.RemoveAll()
	t.srcVertices2.RemoveAll()
	t.kindex1 = 0.0
	t.kindex2 = 0.0
	t.status1 = StatusInitial
	t.status2 = StatusInitial
}

// MoveTo1 starts a new path or adds a line to the current position for path 1.
func (t *TransDoublePath) MoveTo1(x, y float64) {
	if t.status1 == StatusInitial {
		t.srcVertices1.ModifyLast(array.NewVertexDist(x, y))
		t.status1 = StatusMakingPath
	} else {
		t.LineTo1(x, y)
	}
}

// LineTo1 adds a line segment to path 1.
func (t *TransDoublePath) LineTo1(x, y float64) {
	if t.status1 == StatusMakingPath {
		t.srcVertices1.Add(array.NewVertexDist(x, y))
	}
}

// MoveTo2 starts a new path or adds a line to the current position for path 2.
func (t *TransDoublePath) MoveTo2(x, y float64) {
	if t.status2 == StatusInitial {
		t.srcVertices2.ModifyLast(array.NewVertexDist(x, y))
		t.status2 = StatusMakingPath
	} else {
		t.LineTo2(x, y)
	}
}

// LineTo2 adds a line segment to path 2.
func (t *TransDoublePath) LineTo2(x, y float64) {
	if t.status2 == StatusMakingPath {
		t.srcVertices2.Add(array.NewVertexDist(x, y))
	}
}

// AddPaths adds paths from vertex sources.
// This corresponds to AGG's template add_paths method.
func (t *TransDoublePath) AddPaths(vs1, vs2 VertexSource, path1ID, path2ID uint) {
	// Add vertices from first path
	vs1.Rewind(path1ID)
	for {
		x, y, cmd := vs1.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			t.MoveTo1(x, y)
		} else if basics.IsVertex(cmd) {
			t.LineTo1(x, y)
		}
	}

	// Add vertices from second path
	vs2.Rewind(path2ID)
	for {
		x, y, cmd := vs2.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			t.MoveTo2(x, y)
		} else if basics.IsVertex(cmd) {
			t.LineTo2(x, y)
		}
	}

	t.FinalizePaths()
}

// finalizePath prepares a single path for transformation.
// This corresponds to AGG's finalize_path method.
func (t *TransDoublePath) finalizePath(vertices *array.VertexSequence[array.VertexDist]) float64 {
	vertices.Close(false)

	// Remove degenerate final segments (like in AGG)
	if vertices.Size() > 2 {
		lastIdx := vertices.Size() - 1
		secondLastIdx := vertices.Size() - 2
		thirdLastIdx := vertices.Size() - 3

		// Get the vertices
		lastVertex := vertices.At(lastIdx)
		secondLastVertex := vertices.At(secondLastIdx)
		thirdLastVertex := vertices.At(thirdLastIdx)

		// Calculate distances for the last segments
		secondToLastDist := basics.CalcDistance(secondLastVertex.X, secondLastVertex.Y, lastVertex.X, lastVertex.Y)
		thirdToSecondDist := basics.CalcDistance(thirdLastVertex.X, thirdLastVertex.Y, secondLastVertex.X, secondLastVertex.Y)

		// If the second-to-last segment is very short compared to the third-to-last
		if secondToLastDist*10.0 < thirdToSecondDist {
			// Combine distances and move last vertex to second-to-last position
			combinedDist := thirdToSecondDist + secondToLastDist
			vertices.Set(secondLastIdx, lastVertex)
			vertices.RemoveLast()

			// Update the distance for the combined segment
			secondLastVertex = vertices.At(secondLastIdx)
			secondLastVertex.Dist = combinedDist
			vertices.Set(secondLastIdx, secondLastVertex)
		}
	}

	// Calculate cumulative distances
	totalDist := 0.0
	for i := 0; i < vertices.Size(); i++ {
		v := vertices.At(i)
		d := v.Dist
		v.Dist = totalDist
		vertices.Set(i, v)
		totalDist += d
	}

	// Return scaling factor for fast indexing
	if totalDist > 0 {
		return float64(vertices.Size()-1) / totalDist
	}
	return 0.0
}

// FinalizePaths completes path construction and prepares for transformation.
func (t *TransDoublePath) FinalizePaths() {
	if t.status1 == StatusMakingPath && t.srcVertices1.Size() > 1 &&
		t.status2 == StatusMakingPath && t.srcVertices2.Size() > 1 {

		t.kindex1 = t.finalizePath(t.srcVertices1)
		t.kindex2 = t.finalizePath(t.srcVertices2)
		t.status1 = StatusReady
		t.status2 = StatusReady
	}
}

// TotalLength1 returns the total length of path 1.
func (t *TransDoublePath) TotalLength1() float64 {
	if t.baseLength >= 1e-10 {
		return t.baseLength
	}
	if t.status1 == StatusReady {
		return t.srcVertices1.At(t.srcVertices1.Size() - 1).Dist
	}
	return 0.0
}

// TotalLength2 returns the total length of path 2.
func (t *TransDoublePath) TotalLength2() float64 {
	if t.baseLength >= 1e-10 {
		return t.baseLength
	}
	if t.status2 == StatusReady {
		return t.srcVertices2.At(t.srcVertices2.Size() - 1).Dist
	}
	return 0.0
}

// transform1 transforms a point along a single path.
// This corresponds to AGG's transform1 method.
func (t *TransDoublePath) transform1(vertices *array.VertexSequence[array.VertexDist],
	kindex, kx float64, x, y *float64) {

	var x1, y1, dx, dy, d, dd float64

	*x *= kx

	switch {
	case *x < 0.0:
		// Extrapolation on the left
		v0 := vertices.At(0)
		v1 := vertices.At(1)

		x1 = v0.X
		y1 = v0.Y
		dx = v1.X - x1
		dy = v1.Y - y1
		dd = v1.Dist - v0.Dist
		d = *x

	case *x > vertices.At(vertices.Size()-1).Dist:
		// Extrapolation on the right
		lastIdx := vertices.Size() - 1
		secondLastIdx := vertices.Size() - 2

		vLast := vertices.At(lastIdx)
		vSecondLast := vertices.At(secondLastIdx)

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
			j = vertices.Size() - 1
			for (j - i) > 1 {
				k := (i + j) >> 1
				if *x < vertices.At(k).Dist {
					j = k
				} else {
					i = k
				}
			}

			vi := vertices.At(i)
			vj := vertices.At(j)

			d = vi.Dist
			dd = vj.Dist - d
			d = *x - d
		} else {
			// Use uniform distribution with fast indexing
			i = int(*x * kindex)
			j = i + 1

			vi := vertices.At(i)
			vj := vertices.At(j)

			dd = vj.Dist - vi.Dist
			d = ((*x * kindex) - float64(i)) * dd
		}

		vi := vertices.At(i)
		vj := vertices.At(j)

		x1 = vi.X
		y1 = vi.Y
		dx = vj.X - x1
		dy = vj.Y - y1
	}

	// Calculate the interpolated point
	*x = x1 + dx*d/dd
	*y = y1 + dy*d/dd
}

// Transform transforms coordinates between the two paths.
// This corresponds to AGG's transform method.
func (t *TransDoublePath) Transform(x, y *float64) {
	if t.status1 == StatusReady && t.status2 == StatusReady {
		// Scale x coordinate if base length is set
		if t.baseLength > 1e-10 {
			*x *= t.srcVertices1.At(t.srcVertices1.Size()-1).Dist / t.baseLength
		}

		x1 := *x
		y1 := *y
		x2 := *x
		y2 := *y

		// Calculate scaling ratio between the two paths
		dd := t.srcVertices2.At(t.srcVertices2.Size()-1).Dist /
			t.srcVertices1.At(t.srcVertices1.Size()-1).Dist

		// Transform along both paths
		t.transform1(t.srcVertices1, t.kindex1, 1.0, &x1, &y1)
		t.transform1(t.srcVertices2, t.kindex2, dd, &x2, &y2)

		// Interpolate between the two paths based on y coordinate
		*x = x1 + *y*(x2-x1)/t.baseHeight
		*y = y1 + *y*(y2-y1)/t.baseHeight
	}
}
