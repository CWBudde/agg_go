// Package rasterizer provides anti-aliased outline rasterization functionality.
// This implements a port of AGG's rasterizer_outline_aa template class.
package rasterizer

import (
	"agg_go/internal/array"
	"agg_go/internal/basics"
	"agg_go/internal/primitives"
)

// OutlineAAJoin represents the line join style for anti-aliased outline rendering.
// This corresponds to AGG's outline_aa_join_e enum.
type OutlineAAJoin int

const (
	OutlineNoJoin            OutlineAAJoin = iota // outline_no_join
	OutlineMiterJoin                              // outline_miter_join
	OutlineRoundJoin                              // outline_round_join
	OutlineMiterAccurateJoin                      // outline_miter_accurate_join
)

// OutlineAARenderer defines the interface that a renderer must implement
// to be used with RasterizerOutlineAA.
// This corresponds to the Renderer template parameter in AGG.
type OutlineAARenderer[C any] interface {
	// AccurateJoinOnly returns true if the renderer only supports accurate joins
	AccurateJoinOnly() bool

	// Color sets the current color for rendering
	Color(c C)

	// Line rendering methods for different configurations
	// line0: render line without caps
	Line0(lp primitives.LineParameters)

	// line1: render line with start cap
	Line1(lp primitives.LineParameters, sx, sy int)

	// line2: render line with end cap
	Line2(lp primitives.LineParameters, ex, ey int)

	// line3: render line with both caps
	Line3(lp primitives.LineParameters, sx, sy, ex, ey int)

	// Pie renders a pie-shaped join (for round joins)
	Pie(x, y, x1, y1, x2, y2 int)

	// Semidot renders a semicircular cap
	Semidot(cmp func(int) bool, x, y, x1, y1 int)
}

// DrawVars represents the variables used during drawing.
// This corresponds to AGG's draw_vars struct.
type DrawVars struct {
	Idx                uint32                    // Current vertex index
	X1, Y1, X2, Y2     int                       // Current line coordinates
	Curr, Next         primitives.LineParameters // Current and next line parameters
	LCurr, LNext       int                       // Current and next line lengths
	XB1, YB1, XB2, YB2 int                       // Bisector coordinates
	Flags              uint32                    // Drawing flags
}

// RasterizerOutlineAA provides anti-aliased outline rasterization.
// This is a port of AGG's rasterizer_outline_aa<Renderer, Coord> template class.
type RasterizerOutlineAA[R OutlineAARenderer[C], C any] struct {
	renderer    R                              // The attached renderer
	srcVertices *array.LineAAVertexSequence    // Source vertex storage
	lineJoin    OutlineAAJoin                  // Line join style
	roundCap    bool                           // Whether to use round caps
	startX      int                            // Starting X coordinate
	startY      int                            // Starting Y coordinate
}

// NewRasterizerOutlineAA creates a new anti-aliased outline rasterizer.
func NewRasterizerOutlineAA[R OutlineAARenderer[C], C any](ren R) *RasterizerOutlineAA[R, C] {
	join := OutlineRoundJoin
	if ren.AccurateJoinOnly() {
		join = OutlineMiterAccurateJoin
	}

	return &RasterizerOutlineAA[R, C]{
		renderer:    ren,
		srcVertices: array.NewLineAAVertexSequence(),
		lineJoin:    join,
		roundCap:    false,
		startX:      0,
		startY:      0,
	}
}

// Attach attaches a new renderer to this rasterizer.
func (r *RasterizerOutlineAA[R, C]) Attach(ren R) {
	r.renderer = ren
}

// LineJoin sets the line join style.
func (r *RasterizerOutlineAA[R, C]) SetLineJoin(join OutlineAAJoin) {
	if r.renderer.AccurateJoinOnly() {
		r.lineJoin = OutlineMiterAccurateJoin
	} else {
		r.lineJoin = join
	}
}

// GetLineJoin returns the current line join style.
func (r *RasterizerOutlineAA[R, C]) GetLineJoin() OutlineAAJoin {
	return r.lineJoin
}

// SetRoundCap sets whether to use round caps.
func (r *RasterizerOutlineAA[R, C]) SetRoundCap(v bool) {
	r.roundCap = v
}

// GetRoundCap returns the current round cap setting.
func (r *RasterizerOutlineAA[R, C]) GetRoundCap() bool {
	return r.roundCap
}

// MoveTo moves to the specified integer coordinates.
func (r *RasterizerOutlineAA[R, C]) MoveTo(x, y int) {
	r.startX = x
	r.startY = y
	r.srcVertices.ModifyLast(array.NewLineAAVertex(x, y))
}

// LineTo draws a line to the specified integer coordinates.
func (r *RasterizerOutlineAA[R, C]) LineTo(x, y int) {
	r.srcVertices.Add(array.NewLineAAVertex(x, y))
}

// MoveToD moves to the specified floating-point coordinates.
func (r *RasterizerOutlineAA[R, C]) MoveToD(x, y float64) {
	coord := primitives.LineCoord{}
	r.MoveTo(coord.Conv(x), coord.Conv(y))
}

// LineToD draws a line to the specified floating-point coordinates.
func (r *RasterizerOutlineAA[R, C]) LineToD(x, y float64) {
	coord := primitives.LineCoord{}
	r.LineTo(coord.Conv(x), coord.Conv(y))
}

// AddVertex adds a vertex with the specified command to the path.
func (r *RasterizerOutlineAA[R, C]) AddVertex(x, y float64, cmd uint32) {
	if basics.IsMoveTo(basics.PathCommand(cmd)) {
		r.Render(false)
		r.MoveToD(x, y)
	} else {
		if basics.IsEndPoly(basics.PathCommand(cmd)) {
			r.Render(basics.IsClosed(cmd))
			if basics.IsClosed(cmd) {
				r.MoveTo(r.startX, r.startY)
			}
		} else {
			r.LineToD(x, y)
		}
	}
}

// AddPath adds an entire path from a vertex source.
func (r *RasterizerOutlineAA[R, C]) AddPath(vs VertexSource, pathID uint32) {
	var x, y float64

	vs.Rewind(pathID)
	for {
		cmd := vs.Vertex(&x, &y)
		if basics.IsStop(basics.PathCommand(cmd)) {
			break
		}
		r.AddVertex(x, y, cmd)
	}
	r.Render(false)
}

// RenderAllPaths renders multiple paths using different colors.
func (r *RasterizerOutlineAA[R, C]) RenderAllPaths(vs VertexSource, colors ColorStorage[C], pathIDs PathIDStorage, numPaths int) {
	for i := 0; i < numPaths; i++ {
		r.renderer.Color(colors.GetColor(i))
		r.AddPath(vs, pathIDs.GetPathID(i))
	}
}

// RenderCtrl renders a UI control.
func (r *RasterizerOutlineAA[R, C]) RenderCtrl(ctrl Controller[C]) {
	for i := 0; i < ctrl.NumPaths(); i++ {
		r.renderer.Color(ctrl.Color(i))
		r.AddPath(ctrl, uint32(i))
	}
}

// Draw renders vertices from start to end index.
// This corresponds to AGG's draw method.
func (r *RasterizerOutlineAA[R, C]) draw(dv *DrawVars, start, end int) {
	for i := start; i < end; i++ {
		if r.lineJoin == OutlineRoundJoin {
			dv.XB1 = dv.Curr.X1 + (dv.Curr.Y2 - dv.Curr.Y1)
			dv.YB1 = dv.Curr.Y1 - (dv.Curr.X2 - dv.Curr.X1)
			dv.XB2 = dv.Curr.X2 + (dv.Curr.Y2 - dv.Curr.Y1)
			dv.YB2 = dv.Curr.Y2 - (dv.Curr.X2 - dv.Curr.X1)
		}

		switch dv.Flags {
		case 0:
			r.renderer.Line3(dv.Curr, dv.XB1, dv.YB1, dv.XB2, dv.YB2)
		case 1:
			r.renderer.Line2(dv.Curr, dv.XB2, dv.YB2)
		case 2:
			r.renderer.Line1(dv.Curr, dv.XB1, dv.YB1)
		case 3:
			r.renderer.Line0(dv.Curr)
		}

		if r.lineJoin == OutlineRoundJoin && (dv.Flags&2) == 0 {
			r.renderer.Pie(
				dv.Curr.X2, dv.Curr.Y2,
				dv.Curr.X2+(dv.Curr.Y2-dv.Curr.Y1),
				dv.Curr.Y2-(dv.Curr.X2-dv.Curr.X1),
				dv.Curr.X2+(dv.Next.Y2-dv.Next.Y1),
				dv.Curr.Y2-(dv.Next.X2-dv.Next.X1),
			)
		}

		dv.X1 = dv.X2
		dv.Y1 = dv.Y2
		dv.LCurr = dv.LNext
		dv.LNext = r.srcVertices.Get(int(dv.Idx)).Len

		dv.Idx++
		if int(dv.Idx) >= r.srcVertices.Size() {
			dv.Idx = 0
		}

		v := r.srcVertices.Get(int(dv.Idx))
		dv.X2 = v.X
		dv.Y2 = v.Y

		dv.Curr = dv.Next
		dv.Next = primitives.NewLineParameters(dv.X1, dv.Y1, dv.X2, dv.Y2, dv.LNext)
		dv.XB1 = dv.XB2
		dv.YB1 = dv.YB2

		switch r.lineJoin {
		case OutlineNoJoin:
			dv.Flags = 3

		case OutlineMiterJoin:
			dv.Flags >>= 1
			if dv.Curr.DiagonalQuadrant() == dv.Next.DiagonalQuadrant() {
				dv.Flags |= 2
			}
			if (dv.Flags & 2) == 0 {
				dv.XB2, dv.YB2 = primitives.Bisectrix(&dv.Curr, &dv.Next)
			}

		case OutlineRoundJoin:
			dv.Flags >>= 1
			if dv.Curr.DiagonalQuadrant() == dv.Next.DiagonalQuadrant() {
				dv.Flags |= 2
			}

		case OutlineMiterAccurateJoin:
			dv.Flags = 0
			dv.XB2, dv.YB2 = primitives.Bisectrix(&dv.Curr, &dv.Next)
		}
	}
}

// Render processes and renders the accumulated vertices.
// This corresponds to AGG's render method.
func (r *RasterizerOutlineAA[R, C]) Render(closePolygon bool) {
	r.srcVertices.Close(closePolygon)

	if closePolygon {
		r.renderClosed()
	} else {
		r.renderOpen()
	}

	r.srcVertices.RemoveAll()
}

// renderClosed renders a closed polygon.
func (r *RasterizerOutlineAA[R, C]) renderClosed() {
	if r.srcVertices.Size() < 3 {
		return
	}

	dv := &DrawVars{Idx: 2}

	// Get the last vertex
	v := r.srcVertices.Get(r.srcVertices.Size() - 1)
	x1 := v.X
	y1 := v.Y
	lprev := v.Len

	// Get the first vertex
	v = r.srcVertices.Get(0)
	x2 := v.X
	y2 := v.Y
	dv.LCurr = v.Len
	prev := primitives.NewLineParameters(x1, y1, x2, y2, lprev)

	// Get the second vertex
	v = r.srcVertices.Get(1)
	dv.X1 = v.X
	dv.Y1 = v.Y
	dv.LNext = v.Len
	dv.Curr = primitives.NewLineParameters(x2, y2, dv.X1, dv.Y1, dv.LCurr)

	// Get the third vertex
	v = r.srcVertices.Get(int(dv.Idx))
	dv.X2 = v.X
	dv.Y2 = v.Y
	dv.Next = primitives.NewLineParameters(dv.X1, dv.Y1, dv.X2, dv.Y2, dv.LNext)

	dv.XB1, dv.YB1, dv.XB2, dv.YB2 = 0, 0, 0, 0

	switch r.lineJoin {
	case OutlineNoJoin:
		dv.Flags = 3

	case OutlineMiterJoin, OutlineRoundJoin:
		dv.Flags = 0
		if prev.DiagonalQuadrant() == dv.Curr.DiagonalQuadrant() {
			dv.Flags |= 1
		}
		if dv.Curr.DiagonalQuadrant() == dv.Next.DiagonalQuadrant() {
			dv.Flags |= 2
		}

	case OutlineMiterAccurateJoin:
		dv.Flags = 0
	}

	if (dv.Flags&1) == 0 && r.lineJoin != OutlineRoundJoin {
		dv.XB1, dv.YB1 = primitives.Bisectrix(&prev, &dv.Curr)
	}

	if (dv.Flags&2) == 0 && r.lineJoin != OutlineRoundJoin {
		dv.XB2, dv.YB2 = primitives.Bisectrix(&dv.Curr, &dv.Next)
	}

	r.draw(dv, 0, r.srcVertices.Size())
}

// renderOpen renders an open path.
func (r *RasterizerOutlineAA[R, C]) renderOpen() {
	switch r.srcVertices.Size() {
	case 0, 1:
		return

	case 2:
		r.renderTwoVertices()

	case 3:
		r.renderThreeVertices()

	default:
		r.renderMultipleVertices()
	}
}

// renderTwoVertices renders a path with exactly two vertices.
func (r *RasterizerOutlineAA[R, C]) renderTwoVertices() {
	v1 := r.srcVertices.Get(0)
	v2 := r.srcVertices.Get(1)

	lp := primitives.NewLineParameters(v1.X, v1.Y, v2.X, v2.Y, v1.Len)

	if r.roundCap {
		r.renderer.Semidot(primitives.CmpDistStart, v1.X, v1.Y,
			v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X))
	}

	r.renderer.Line3(lp,
		v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X),
		v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X))

	if r.roundCap {
		r.renderer.Semidot(primitives.CmpDistEnd, v2.X, v2.Y,
			v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X))
	}
}

// renderThreeVertices renders a path with exactly three vertices.
func (r *RasterizerOutlineAA[R, C]) renderThreeVertices() {
	v1 := r.srcVertices.Get(0)
	v2 := r.srcVertices.Get(1)
	v3 := r.srcVertices.Get(2)

	lp1 := primitives.NewLineParameters(v1.X, v1.Y, v2.X, v2.Y, v1.Len)
	lp2 := primitives.NewLineParameters(v2.X, v2.Y, v3.X, v3.Y, v2.Len)

	if r.roundCap {
		r.renderer.Semidot(primitives.CmpDistStart, v1.X, v1.Y,
			v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X))
	}

	if r.lineJoin == OutlineRoundJoin {
		r.renderer.Line3(lp1, v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X),
			v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X))

		r.renderer.Pie(v2.X, v2.Y, v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X),
			v2.X+(v3.Y-v2.Y), v2.Y-(v3.X-v2.X))

		r.renderer.Line3(lp2, v2.X+(v3.Y-v2.Y), v2.Y-(v3.X-v2.X),
			v3.X+(v3.Y-v2.Y), v3.Y-(v3.X-v2.X))
	} else {
		xb, yb := primitives.Bisectrix(&lp1, &lp2)
		r.renderer.Line3(lp1, v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X), xb, yb)
		r.renderer.Line3(lp2, xb, yb, v3.X+(v3.Y-v2.Y), v3.Y-(v3.X-v2.X))
	}

	if r.roundCap {
		r.renderer.Semidot(primitives.CmpDistEnd, v3.X, v3.Y,
			v3.X+(v3.Y-v2.Y), v3.Y-(v3.X-v2.X))
	}
}

// renderMultipleVertices renders a path with more than three vertices.
func (r *RasterizerOutlineAA[R, C]) renderMultipleVertices() {
	dv := &DrawVars{Idx: 3}

	// Setup initial vertices
	v1 := r.srcVertices.Get(0)
	v2 := r.srcVertices.Get(1)
	v3 := r.srcVertices.Get(2)
	v4 := r.srcVertices.Get(int(dv.Idx))

	prev := primitives.NewLineParameters(v1.X, v1.Y, v2.X, v2.Y, v1.Len)
	dv.Curr = primitives.NewLineParameters(v2.X, v2.Y, v3.X, v3.Y, v2.Len)
	dv.Next = primitives.NewLineParameters(v3.X, v3.Y, v4.X, v4.Y, v3.Len)

	dv.X1, dv.Y1 = v3.X, v3.Y
	dv.X2, dv.Y2 = v4.X, v4.Y
	dv.LCurr = v2.Len
	dv.LNext = v3.Len

	dv.XB1, dv.YB1, dv.XB2, dv.YB2 = 0, 0, 0, 0

	switch r.lineJoin {
	case OutlineNoJoin:
		dv.Flags = 3

	case OutlineMiterJoin, OutlineRoundJoin:
		dv.Flags = 0
		if prev.DiagonalQuadrant() == dv.Curr.DiagonalQuadrant() {
			dv.Flags |= 1
		}
		if dv.Curr.DiagonalQuadrant() == dv.Next.DiagonalQuadrant() {
			dv.Flags |= 2
		}

	case OutlineMiterAccurateJoin:
		dv.Flags = 0
	}

	if r.roundCap {
		r.renderer.Semidot(primitives.CmpDistStart, v1.X, v1.Y,
			v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X))
	}

	// Render first segment
	if (dv.Flags & 1) == 0 {
		if r.lineJoin == OutlineRoundJoin {
			r.renderer.Line3(prev, v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X),
				v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X))
			r.renderer.Pie(prev.X2, prev.Y2,
				v2.X+(v2.Y-v1.Y), v2.Y-(v2.X-v1.X),
				dv.Curr.X1+(dv.Curr.Y2-dv.Curr.Y1),
				dv.Curr.Y1-(dv.Curr.X2-dv.Curr.X1))
		} else {
			dv.XB1, dv.YB1 = primitives.Bisectrix(&prev, &dv.Curr)
			r.renderer.Line3(prev, v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X), dv.XB1, dv.YB1)
		}
	} else {
		r.renderer.Line1(prev, v1.X+(v2.Y-v1.Y), v1.Y-(v2.X-v1.X))
	}

	if (dv.Flags&2) == 0 && r.lineJoin != OutlineRoundJoin {
		dv.XB2, dv.YB2 = primitives.Bisectrix(&dv.Curr, &dv.Next)
	}

	// Render middle segments
	r.draw(dv, 1, r.srcVertices.Size()-2)

	// Render last segment
	if (dv.Flags & 1) == 0 {
		if r.lineJoin == OutlineRoundJoin {
			r.renderer.Line3(dv.Curr,
				dv.Curr.X1+(dv.Curr.Y2-dv.Curr.Y1),
				dv.Curr.Y1-(dv.Curr.X2-dv.Curr.X1),
				dv.Curr.X2+(dv.Curr.Y2-dv.Curr.Y1),
				dv.Curr.Y2-(dv.Curr.X2-dv.Curr.X1))
		} else {
			r.renderer.Line3(dv.Curr, dv.XB1, dv.YB1,
				dv.Curr.X2+(dv.Curr.Y2-dv.Curr.Y1),
				dv.Curr.Y2-(dv.Curr.X2-dv.Curr.X1))
		}
	} else {
		r.renderer.Line2(dv.Curr,
			dv.Curr.X2+(dv.Curr.Y2-dv.Curr.Y1),
			dv.Curr.Y2-(dv.Curr.X2-dv.Curr.X1))
	}

	if r.roundCap {
		lastVertex := r.srcVertices.Get(r.srcVertices.Size() - 1)
		r.renderer.Semidot(primitives.CmpDistEnd, dv.Curr.X2, dv.Curr.Y2,
			dv.Curr.X2+(dv.Curr.Y2-dv.Curr.Y1),
			dv.Curr.Y2-(dv.Curr.X2-dv.Curr.X1))
		_ = lastVertex // Suppress unused variable warning
	}
}
