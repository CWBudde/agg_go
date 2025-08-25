// Package primitives provides high-level primitive drawing operations.
// This package implements fast, integer-based drawing operations using
// Bresenham algorithms for lines and ellipses.
package primitives

import (
	"agg_go/internal/basics"
	"agg_go/internal/primitives"
)

// BaseRenderer defines the interface that a base renderer must implement
// to be used with RendererPrimitives.
type BaseRenderer[C any] interface {
	// Pixel operations
	BlendPixel(x, y int, c C, cover basics.Int8u)

	// Line operations
	BlendHline(x1, y, x2 int, c C, cover basics.Int8u)
	BlendVline(x, y1, y2 int, c C, cover basics.Int8u)

	// Area operations
	BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u)

	// Clipping box for visibility testing
	BoundingClipBox() basics.RectI
}

// RendererPrimitives provides primitive drawing operations on top of a base renderer.
// This is a port of AGG's renderer_primitives<BaseRenderer> template class.
type RendererPrimitives[BR BaseRenderer[C], C any] struct {
	ren       BR  // base renderer
	fillColor C   // current fill color
	lineColor C   // current line color
	currX     int // current x position
	currY     int // current y position
}

// NewRendererPrimitives creates a new primitive renderer with the given base renderer.
func NewRendererPrimitives[BR BaseRenderer[C], C any](ren BR) *RendererPrimitives[BR, C] {
	return &RendererPrimitives[BR, C]{
		ren:   ren,
		currX: 0,
		currY: 0,
	}
}

// Attach attaches a new base renderer.
func (rp *RendererPrimitives[BR, C]) Attach(ren BR) {
	rp.ren = ren
}

// Coord converts a floating point coordinate to the subpixel scale used by line_bresenham_interpolator.
func (rp *RendererPrimitives[BR, C]) Coord(c float64) int {
	return int(c * float64(1<<8)) // subpixel_scale = 1 << 8
}

// FillColor sets the fill color.
func (rp *RendererPrimitives[BR, C]) FillColor(c C) {
	rp.fillColor = c
}

// LineColor sets the line color.
func (rp *RendererPrimitives[BR, C]) LineColor(c C) {
	rp.lineColor = c
}

// GetFillColor returns the current fill color.
func (rp *RendererPrimitives[BR, C]) GetFillColor() C {
	return rp.fillColor
}

// GetLineColor returns the current line color.
func (rp *RendererPrimitives[BR, C]) GetLineColor() C {
	return rp.lineColor
}

// Rectangle draws a rectangle outline.
func (rp *RendererPrimitives[BR, C]) Rectangle(x1, y1, x2, y2 int) {
	rp.ren.BlendHline(x1, y1, x2-1, rp.lineColor, basics.CoverFull)
	rp.ren.BlendVline(x2, y1, y2-1, rp.lineColor, basics.CoverFull)
	rp.ren.BlendHline(x1+1, y2, x2, rp.lineColor, basics.CoverFull)
	rp.ren.BlendVline(x1, y1+1, y2, rp.lineColor, basics.CoverFull)
}

// SolidRectangle draws a filled rectangle.
func (rp *RendererPrimitives[BR, C]) SolidRectangle(x1, y1, x2, y2 int) {
	rp.ren.BlendBar(x1, y1, x2, y2, rp.fillColor, basics.CoverFull)
}

// OutlinedRectangle draws a filled rectangle with an outline.
func (rp *RendererPrimitives[BR, C]) OutlinedRectangle(x1, y1, x2, y2 int) {
	rp.Rectangle(x1, y1, x2, y2)
	rp.ren.BlendBar(x1+1, y1+1, x2-1, y2-1, rp.fillColor, basics.CoverFull)
}

// Ellipse draws an ellipse outline using Bresenham algorithm.
func (rp *RendererPrimitives[BR, C]) Ellipse(x, y, rx, ry int) {
	ei := primitives.NewEllipseBresenhamInterpolator(rx, ry)
	dx := 0
	dy := -ry

	for dy < 0 {
		dx += ei.Dx()
		dy += ei.Dy()

		rp.ren.BlendPixel(x+dx, y+dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x+dx, y-dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x-dx, y-dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x-dx, y+dy, rp.lineColor, basics.CoverFull)

		ei.Inc()
	}
}

// SolidEllipse draws a filled ellipse using Bresenham algorithm.
func (rp *RendererPrimitives[BR, C]) SolidEllipse(x, y, rx, ry int) {
	ei := primitives.NewEllipseBresenhamInterpolator(rx, ry)
	dx := 0
	dy := -ry
	dy0 := dy
	dx0 := dx

	for dy < 0 {
		dx += ei.Dx()
		dy += ei.Dy()

		if dy != dy0 {
			rp.ren.BlendHline(x-dx0, y+dy0, x+dx0, rp.fillColor, basics.CoverFull)
			rp.ren.BlendHline(x-dx0, y-dy0, x+dx0, rp.fillColor, basics.CoverFull)
		}
		dx0 = dx
		dy0 = dy
		ei.Inc()
	}
	// Draw the final horizontal line
	rp.ren.BlendHline(x-dx0, y+dy0, x+dx0, rp.fillColor, basics.CoverFull)
}

// OutlinedEllipse draws a filled ellipse with an outline.
func (rp *RendererPrimitives[BR, C]) OutlinedEllipse(x, y, rx, ry int) {
	ei := primitives.NewEllipseBresenhamInterpolator(rx, ry)
	dx := 0
	dy := -ry

	for dy < 0 {
		dx += ei.Dx()
		dy += ei.Dy()

		rp.ren.BlendPixel(x+dx, y+dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x+dx, y-dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x-dx, y-dy, rp.lineColor, basics.CoverFull)
		rp.ren.BlendPixel(x-dx, y+dy, rp.lineColor, basics.CoverFull)

		if ei.Dy() != 0 && dx > 0 {
			rp.ren.BlendHline(x-dx+1, y+dy, x+dx-1, rp.fillColor, basics.CoverFull)
			rp.ren.BlendHline(x-dx+1, y-dy, x+dx-1, rp.fillColor, basics.CoverFull)
		}
		ei.Inc()
	}
}

// Line draws a line from (x1,y1) to (x2,y2).
func (rp *RendererPrimitives[BR, C]) Line(x1, y1, x2, y2 int, last bool) {
	li := primitives.NewLineBresenhamInterpolator(x1, y1, x2, y2)

	length := li.Len()
	if length == 0 {
		if last {
			rp.ren.BlendPixel(li.LineLr(x1), li.LineLr(y1), rp.lineColor, basics.CoverFull)
		}
		return
	}

	if last {
		length++
	}

	if li.IsVer() {
		for length > 0 {
			rp.ren.BlendPixel(li.X2(), li.Y1(), rp.lineColor, basics.CoverFull)
			li.VStep()
			length--
		}
	} else {
		for length > 0 {
			rp.ren.BlendPixel(li.X1(), li.Y2(), rp.lineColor, basics.CoverFull)
			li.HStep()
			length--
		}
	}
}

// MoveTo moves the current position to (x, y).
func (rp *RendererPrimitives[BR, C]) MoveTo(x, y int) {
	rp.currX = x
	rp.currY = y
}

// LineTo draws a line from the current position to (x, y).
func (rp *RendererPrimitives[BR, C]) LineTo(x, y int) {
	rp.Line(rp.currX, rp.currY, x, y, false)
	rp.currX = x
	rp.currY = y
}

// LineToLast draws a line from the current position to (x, y) with last pixel.
func (rp *RendererPrimitives[BR, C]) LineToLast(x, y int, last bool) {
	rp.Line(rp.currX, rp.currY, x, y, last)
	rp.currX = x
	rp.currY = y
}

// Ren returns the base renderer.
func (rp *RendererPrimitives[BR, C]) Ren() BR {
	return rp.ren
}
