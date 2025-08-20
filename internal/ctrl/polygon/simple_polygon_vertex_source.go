// Package polygon provides polygon control implementation for AGG.
// This is a port of AGG's polygon_ctrl_impl and simple_polygon_vertex_source classes.
package polygon

import (
	"math"

	"agg_go/internal/basics"
)

// SimplePolygonVertexSource generates vertices for a simple polygon.
// This corresponds to AGG's simple_polygon_vertex_source class.
type SimplePolygonVertexSource struct {
	polygon   []float64 // coordinate array [x1, y1, x2, y2, ...]
	numPoints uint
	vertex    uint
	roundoff  bool
	close     bool
}

// NewSimplePolygonVertexSource creates a new simple polygon vertex source.
// polygon: coordinate array where each pair represents (x, y)
// numPoints: number of points in the polygon
// roundoff: whether to apply floor+0.5 rounding to coordinates
// close: whether to close the polygon path
func NewSimplePolygonVertexSource(polygon []float64, numPoints uint, roundoff, close bool) *SimplePolygonVertexSource {
	return &SimplePolygonVertexSource{
		polygon:   polygon,
		numPoints: numPoints,
		vertex:    0,
		roundoff:  roundoff,
		close:     close,
	}
}

// Close sets whether the polygon should be closed.
func (s *SimplePolygonVertexSource) Close(f bool) {
	s.close = f
}

// IsClose returns whether the polygon is closed.
func (s *SimplePolygonVertexSource) IsClose() bool {
	return s.close
}

// Rewind resets the vertex iterator to the beginning.
func (s *SimplePolygonVertexSource) Rewind(pathID uint) {
	s.vertex = 0
}

// Vertex returns the next vertex in the polygon.
// Returns coordinates and path command following AGG vertex source protocol.
func (s *SimplePolygonVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if s.vertex > s.numPoints {
		return 0, 0, basics.PathCmdStop
	}

	if s.vertex == s.numPoints {
		s.vertex++
		if s.close {
			return 0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose))
		}
		return 0, 0, basics.PathCmdEndPoly
	}

	x = s.polygon[s.vertex*2]
	y = s.polygon[s.vertex*2+1]

	if s.roundoff {
		x = math.Floor(x) + 0.5
		y = math.Floor(y) + 0.5
	}

	s.vertex++

	if s.vertex == 1 {
		return x, y, basics.PathCmdMoveTo
	}
	return x, y, basics.PathCmdLineTo
}
