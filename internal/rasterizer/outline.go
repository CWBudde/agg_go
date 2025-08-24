// Package rasterizer provides outline rasterization functionality.
// This implements a port of AGG's rasterizer_outline template class.
package rasterizer

import (
	"agg_go/internal/basics"
)

// OutlineRenderer defines the interface that a renderer must implement
// to be used with RasterizerOutline.
type OutlineRenderer[C any] interface {
	// MoveTo moves the current drawing position to the specified coordinates
	MoveTo(x, y int)

	// LineTo draws a line from the current position to the specified coordinates
	LineTo(x, y int)

	// Coord converts a floating-point coordinate to the subpixel scale
	// used by the renderer (typically multiplied by 256 for subpixel precision)
	Coord(c float64) int

	// LineColor sets the current line color for drawing operations
	LineColor(c C)
}

// RasterizerOutline provides outline rasterization functionality.
// This is a port of AGG's rasterizer_outline<Renderer> template class.
type RasterizerOutline[R OutlineRenderer[C], C any] struct {
	renderer R    // the attached renderer
	startX   int  // starting X coordinate for the current path
	startY   int  // starting Y coordinate for the current path
	vertices uint // number of vertices in the current path
}

// NewRasterizerOutline creates a new outline rasterizer with the given renderer.
func NewRasterizerOutline[R OutlineRenderer[C], C any](ren R) *RasterizerOutline[R, C] {
	return &RasterizerOutline[R, C]{
		renderer: ren,
		startX:   0,
		startY:   0,
		vertices: 0,
	}
}

// Attach attaches a new renderer to this rasterizer.
func (r *RasterizerOutline[R, C]) Attach(ren R) {
	r.renderer = ren
}

// MoveTo moves to the specified integer coordinates and starts a new path.
func (r *RasterizerOutline[R, C]) MoveTo(x, y int) {
	r.vertices = 1
	r.startX = x
	r.startY = y
	r.renderer.MoveTo(x, y)
}

// LineTo draws a line to the specified integer coordinates.
func (r *RasterizerOutline[R, C]) LineTo(x, y int) {
	r.vertices++
	r.renderer.LineTo(x, y)
}

// MoveToD moves to the specified floating-point coordinates.
// Coordinates are converted to integer subpixel coordinates using the renderer's Coord method.
func (r *RasterizerOutline[R, C]) MoveToD(x, y float64) {
	r.MoveTo(r.renderer.Coord(x), r.renderer.Coord(y))
}

// LineToD draws a line to the specified floating-point coordinates.
// Coordinates are converted to integer subpixel coordinates using the renderer's Coord method.
func (r *RasterizerOutline[R, C]) LineToD(x, y float64) {
	r.LineTo(r.renderer.Coord(x), r.renderer.Coord(y))
}

// Close closes the current path by drawing a line back to the starting point.
// Only closes the path if there are more than 2 vertices.
func (r *RasterizerOutline[R, C]) Close() {
	if r.vertices > 2 {
		r.LineTo(r.startX, r.startY)
	}
	r.vertices = 0
}

// AddVertex adds a vertex with the specified command to the path.
// This method interprets path commands and calls the appropriate MoveTo/LineTo/Close methods.
func (r *RasterizerOutline[R, C]) AddVertex(x, y float64, cmd uint32) {
	if basics.IsMoveTo(basics.PathCommand(cmd)) {
		r.MoveToD(x, y)
	} else {
		if basics.IsEndPoly(basics.PathCommand(cmd)) {
			if basics.IsClosed(cmd) {
				r.Close()
			}
		} else {
			r.LineToD(x, y)
		}
	}
}

// AddPath adds an entire path from a vertex source to the rasterizer.
// The vertex source is rewound to the specified path ID before processing.
func (r *RasterizerOutline[R, C]) AddPath(vs VertexSource, pathID uint32) {
	var x, y float64

	vs.Rewind(pathID)
	for {
		cmd := vs.Vertex(&x, &y)
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		r.AddVertex(x, y, cmd)
	}
}

// ColorStorage represents a collection of colors that can be indexed.
type ColorStorage[C any] interface {
	GetColor(index int) C
}

// PathIDStorage represents a collection of path IDs that can be indexed.
type PathIDStorage interface {
	GetPathID(index int) uint32
}

// RenderAllPaths renders multiple paths from a vertex source using different colors.
// Each path is rendered with its corresponding color from the color storage.
func (r *RasterizerOutline[R, C]) RenderAllPaths(vs VertexSource, colors ColorStorage[C], pathIDs PathIDStorage, numPaths int) {
	for i := 0; i < numPaths; i++ {
		r.renderer.LineColor(colors.GetColor(i))
		r.AddPath(vs, pathIDs.GetPathID(i))
	}
}

// Controller represents a UI control that can provide multiple colored paths.
type Controller[C any] interface {
	NumPaths() int
	Color(pathIndex int) C
	VertexSource
}

// RenderCtrl renders a UI control with multiple colored paths.
func (r *RasterizerOutline[R, C]) RenderCtrl(ctrl Controller[C]) {
	for i := 0; i < ctrl.NumPaths(); i++ {
		r.renderer.LineColor(ctrl.Color(i))
		r.AddPath(ctrl, uint32(i))
	}
}
