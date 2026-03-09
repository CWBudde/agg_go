// Package agg provides high-level convenience functions for 2D graphics rendering.
// This file implements a simplified Context API that wraps the lower-level Agg2D interface.
package agg

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// Context provides a high-level, user-friendly interface for 2D graphics rendering.
// It wraps the lower-level Agg2D interface with simplified method names and patterns.
type Context struct {
	agg2d     *Agg2D
	image     *Image
	width     int
	height    int
	lineWidth float64 // Track current line width
}

// NewContext creates a new rendering context with the specified dimensions.
// This provides a simplified interface compared to the lower-level Agg2D API.
func NewContext(width, height int) *Context {
	// Create AGG2D instance
	agg2d := NewAgg2D()

	// Create backing buffer (RGBA format, 4 bytes per pixel)
	stride := width * 4
	bufferSize := height * stride
	buffer := make([]uint8, bufferSize)

	// Attach buffer to AGG2D
	agg2d.Attach(buffer, width, height, stride)

	// Create image wrapper
	image := NewImage(buffer, width, height, stride)

	ctx := &Context{
		agg2d:     agg2d,
		image:     image,
		width:     width,
		height:    height,
		lineWidth: 1.0, // Default line width
	}

	// Set reasonable defaults
	ctx.SetColor(Black)
	ctx.agg2d.LineWidth(ctx.lineWidth)

	return ctx
}

// NewContextForImage creates a new rendering context for an existing image.
func NewContextForImage(img *Image) *Context {
	if img == nil {
		return nil
	}
	agg2d := NewAgg2D()
	agg2d.Attach(img.Data, img.width, img.height, img.renBuf.Stride())

	ctx := &Context{
		agg2d:     agg2d,
		image:     img,
		width:     img.Width(),
		height:    img.Height(),
		lineWidth: 1.0,
	}

	ctx.SetColor(Black)
	ctx.agg2d.LineWidth(ctx.lineWidth)

	return ctx
}

// Height returns the context height in pixels.
func (ctx *Context) Height() int {
	return ctx.height
}

// Width returns the context width in pixels.
func (ctx *Context) Width() int {
	return ctx.width
}

// Clear fills the entire context with the specified color.
func (ctx *Context) Clear(color Color) {
	ctx.agg2d.ClearAll(color)
}

// SetColor sets the current drawing color for both fill and stroke operations.
func (ctx *Context) SetColor(color Color) {
	ctx.agg2d.FillColor(color)
	ctx.agg2d.LineColor(color)
}

// DrawLine draws a line between two points using the current color.
func (ctx *Context) DrawLine(x1, y1, x2, y2 float64) {
	ctx.agg2d.Line(x1, y1, x2, y2)
}

// DrawThickLine draws a line with specified thickness.
func (ctx *Context) DrawThickLine(x1, y1, x2, y2, width float64) {
	oldWidth := ctx.lineWidth
	ctx.agg2d.LineWidth(width)
	ctx.agg2d.Line(x1, y1, x2, y2)
	ctx.agg2d.LineWidth(oldWidth)
}

// DrawRectangle draws a rectangle outline.
func (ctx *Context) DrawRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillRectangle fills a rectangle with the current color.
func (ctx *Context) FillRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawCircle draws a circle outline.
func (ctx *Context) DrawCircle(cx, cy, radius float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillCircle fills a circle with the current color.
func (ctx *Context) FillCircle(cx, cy, radius float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawEllipse draws an ellipse outline.
func (ctx *Context) DrawEllipse(cx, cy, rx, ry float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, rx, ry, CCW)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillEllipse fills an ellipse with the current color.
func (ctx *Context) FillEllipse(cx, cy, rx, ry float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, rx, ry, CCW)
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawRoundedRectangle draws a rounded rectangle outline.
func (ctx *Context) DrawRoundedRectangle(x, y, width, height, radius float64) {
	x2 := x + width
	y2 := y + height
	ctx.agg2d.ResetPath()
	ctx.drawRoundedRectPath(x, y, x2, y2, radius)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillRoundedRectangle fills a rounded rectangle with the current color.
func (ctx *Context) FillRoundedRectangle(x, y, width, height, radius float64) {
	x2 := x + width
	y2 := y + height
	ctx.agg2d.ResetPath()
	ctx.drawRoundedRectPath(x, y, x2, y2, radius)
	ctx.agg2d.DrawPath(FillOnly)
}

// Helper method to create a rounded rectangle path
func (ctx *Context) drawRoundedRectPath(x1, y1, x2, y2, radius float64) {
	roundedRect := shapes.NewRoundedRectEmpty()
	roundedRect.SetRect(x1, y1, x2, y2)
	roundedRect.SetRadius(radius)
	roundedRect.NormalizeRadius()
	roundedRect.Rewind(0)

	first := true
	for {
		var x, y float64
		cmd := roundedRect.Vertex(&x, &y)
		if cmd == basics.PathCmdStop {
			break
		}

		if first {
			ctx.agg2d.MoveTo(x, y)
			first = false
			continue
		}
		if cmd == basics.PathCmdLineTo {
			ctx.agg2d.LineTo(x, y)
			continue
		}
		if cmd&basics.PathCmdMask == basics.PathCmdEndPoly {
			ctx.agg2d.ClosePolygon()
		}
	}
}

// Fill fills the current path with the current color.
func (ctx *Context) Fill() {
	ctx.agg2d.DrawPath(FillOnly)
}

// Stroke strokes the current path with the current color.
func (ctx *Context) Stroke() {
	ctx.agg2d.DrawPath(StrokeOnly)
}

// GetImage returns the underlying image data.
func (ctx *Context) GetImage() *Image {
	return ctx.image
}

// GetAgg2D returns the underlying AGG2D instance for advanced operations.
func (ctx *Context) GetAgg2D() *Agg2D {
	return ctx.agg2d
}

// SetLineWidth sets the line width for stroke operations.
func (ctx *Context) SetLineWidth(width float64) {
	ctx.agg2d.LineWidth(width)
}

// BeginPath starts a new path.
func (ctx *Context) BeginPath() {
	ctx.agg2d.ResetPath()
}

// MoveTo moves to the specified point without drawing.
func (ctx *Context) MoveTo(x, y float64) {
	ctx.agg2d.MoveTo(x, y)
}

// LineTo draws a line to the specified point.
func (ctx *Context) LineTo(x, y float64) {
	ctx.agg2d.LineTo(x, y)
}

// ClosePath closes the current path.
func (ctx *Context) ClosePath() {
	ctx.agg2d.ClosePolygon()
}
