package span

import (
	"math"

	"agg_go/internal/basics"
)

// CoordType represents a coordinate with an associated color.
// This is equivalent to AGG's coord_type struct in span_gouraud.
type CoordType[C any] struct {
	X     float64
	Y     float64
	Color C
}

// SpanGouraud implements Gouraud shading for triangles.
// This is the base class for smooth color interpolation across triangular regions.
// It's equivalent to AGG's span_gouraud template class.
type SpanGouraud[C any] struct {
	coord  [3]CoordType[C]       // Triangle vertices with colors
	x      [8]float64            // Dilated triangle x coordinates
	y      [8]float64            // Dilated triangle y coordinates
	cmd    [8]basics.PathCommand // Path commands for vertex source
	vertex int                   // Current vertex index for path iteration
}

// NewSpanGouraud creates a new Gouraud span generator.
func NewSpanGouraud[C any]() *SpanGouraud[C] {
	sg := &SpanGouraud[C]{
		vertex: 0,
	}
	sg.cmd[0] = basics.PathCmdStop
	return sg
}

// NewSpanGouraudWithTriangle creates a new Gouraud span generator with initial triangle.
func NewSpanGouraudWithTriangle[C any](c1, c2, c3 C, x1, y1, x2, y2, x3, y3, d float64) *SpanGouraud[C] {
	sg := NewSpanGouraud[C]()
	sg.Colors(c1, c2, c3)
	sg.Triangle(x1, y1, x2, y2, x3, y3, d)
	return sg
}

// Colors sets the colors for the three triangle vertices.
func (sg *SpanGouraud[C]) Colors(c1, c2, c3 C) {
	sg.coord[0].Color = c1
	sg.coord[1].Color = c2
	sg.coord[2].Color = c3
}

// Triangle sets the triangle coordinates and optionally dilates it.
// The dilation parameter d is used to achieve numerical stability by creating
// beveled joins in the vertices. The actual color interpolation coordinates
// are calculated as miter joins.
func (sg *SpanGouraud[C]) Triangle(x1, y1, x2, y2, x3, y3, d float64) {
	sg.coord[0].X = x1
	sg.coord[0].Y = y1
	sg.coord[1].X = x2
	sg.coord[1].Y = y2
	sg.coord[2].X = x3
	sg.coord[2].Y = y3

	sg.x[0] = x1
	sg.y[0] = y1
	sg.x[1] = x2
	sg.y[1] = y2
	sg.x[2] = x3
	sg.y[2] = y3

	sg.cmd[0] = basics.PathCmdMoveTo
	sg.cmd[1] = basics.PathCmdLineTo
	sg.cmd[2] = basics.PathCmdLineTo
	sg.cmd[3] = basics.PathCmdStop

	if d != 0.0 {
		// Dilate triangle for numerical stability
		sg.dilateTriangle(x1, y1, x2, y2, x3, y3, d)

		// Calculate miter join intersections for color interpolation
		x, y, ok := basics.CalcIntersection(sg.x[4], sg.y[4], sg.x[5], sg.y[5],
			sg.x[0], sg.y[0], sg.x[1], sg.y[1])
		if ok {
			sg.coord[0].X = x
			sg.coord[0].Y = y
		}

		x, y, ok = basics.CalcIntersection(sg.x[0], sg.y[0], sg.x[1], sg.y[1],
			sg.x[2], sg.y[2], sg.x[3], sg.y[3])
		if ok {
			sg.coord[1].X = x
			sg.coord[1].Y = y
		}

		x, y, ok = basics.CalcIntersection(sg.x[2], sg.y[2], sg.x[3], sg.y[3],
			sg.x[4], sg.y[4], sg.x[5], sg.y[5])
		if ok {
			sg.coord[2].X = x
			sg.coord[2].Y = y
		}

		sg.cmd[3] = basics.PathCmdLineTo
		sg.cmd[4] = basics.PathCmdLineTo
		sg.cmd[5] = basics.PathCmdLineTo
		sg.cmd[6] = basics.PathCmdStop
	}
}

// dilateTriangle performs triangle dilation for numerical stability.
// This creates a 6-vertex polygon with beveled joins.
func (sg *SpanGouraud[C]) dilateTriangle(x1, y1, x2, y2, x3, y3, d float64) {
	loc := basics.CrossProduct(x1, y1, x2, y2, x3, y3)

	if math.Abs(loc) > basics.IntersectionEpsilon {
		if loc > 0.0 {
			d = -d
		}

		// Calculate orthogonal vectors for each edge
		dx1, dy1 := basics.CalcOrthogonal(d, x1, y1, x2, y2)
		dx2, dy2 := basics.CalcOrthogonal(d, x2, y2, x3, y3)
		dx3, dy3 := basics.CalcOrthogonal(d, x3, y3, x1, y1)

		// Create dilated 6-vertex polygon
		sg.x[0] = x1 + dx1
		sg.y[0] = y1 + dy1
		sg.x[1] = x2 + dx1
		sg.y[1] = y2 + dy1
		sg.x[2] = x2 + dx2
		sg.y[2] = y2 + dy2
		sg.x[3] = x3 + dx2
		sg.y[3] = y3 + dy2
		sg.x[4] = x3 + dx3
		sg.y[4] = y3 + dy3
		sg.x[5] = x1 + dx3
		sg.y[5] = y1 + dy3
	}
}

// Rewind resets the vertex iterator (implements vertex source interface).
func (sg *SpanGouraud[C]) Rewind(pathID uint) {
	sg.vertex = 0
}

// Vertex returns the next vertex in the path (implements vertex source interface).
func (sg *SpanGouraud[C]) Vertex() (x, y float64, cmd basics.PathCommand) {
	if sg.vertex >= len(sg.cmd) {
		return 0, 0, basics.PathCmdStop
	}

	x = sg.x[sg.vertex]
	y = sg.y[sg.vertex]
	cmd = sg.cmd[sg.vertex]
	sg.vertex++

	return x, y, cmd
}

// ArrangeVertices sorts the triangle vertices by Y coordinate (bottom to top).
// This is used by derived classes for proper scan-line generation.
func (sg *SpanGouraud[C]) ArrangeVertices() [3]CoordType[C] {
	coord := [3]CoordType[C]{
		sg.coord[0],
		sg.coord[1],
		sg.coord[2],
	}

	// Sort by Y coordinate (bubble sort for 3 elements)
	if coord[0].Y > coord[2].Y {
		coord[0], coord[2] = coord[2], coord[0]
	}

	if coord[0].Y > coord[1].Y {
		coord[0], coord[1] = coord[1], coord[0]
	}

	if coord[1].Y > coord[2].Y {
		coord[1], coord[2] = coord[2], coord[1]
	}

	return coord
}

// Coord returns the triangle coordinates for direct access.
func (sg *SpanGouraud[C]) Coord() [3]CoordType[C] {
	return sg.coord
}
