// Package agg provides high-level convenience functions for 2D graphics rendering.
// This file implements a simplified Context API that wraps the lower-level Agg2D interface.
package agg

import (
	"math"
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

// Width returns the context width in pixels.
func (ctx *Context) Width() int {
	return ctx.width
}

// Height returns the context height in pixels.
func (ctx *Context) Height() int {
	return ctx.height
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
func (ctx *Context) DrawThickLine(x1, y1, x2, y2 float64, width float64) {
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
	// Ensure proper ordering
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Clamp radius to half of width/height
	w := x2 - x1
	h := y2 - y1
	radius = math.Min(radius, math.Min(w/2, h/2))

	if radius <= 0 {
		// No rounding, draw regular rectangle
		ctx.agg2d.MoveTo(x1, y1)
		ctx.agg2d.LineTo(x2, y1)
		ctx.agg2d.LineTo(x2, y2)
		ctx.agg2d.LineTo(x1, y2)
		ctx.agg2d.ClosePolygon()
		return
	}

	// Start from top-left, going clockwise
	ctx.agg2d.MoveTo(x1+radius, y1)
	ctx.agg2d.LineTo(x2-radius, y1)                     // Top edge
	ctx.addCornerArc(x2-radius, y1+radius, radius, 0)   // Top-right corner
	ctx.agg2d.LineTo(x2, y2-radius)                     // Right edge
	ctx.addCornerArc(x2-radius, y2-radius, radius, 90)  // Bottom-right corner
	ctx.agg2d.LineTo(x1+radius, y2)                     // Bottom edge
	ctx.addCornerArc(x1+radius, y2-radius, radius, 180) // Bottom-left corner
	ctx.agg2d.LineTo(x1, y1+radius)                     // Left edge
	ctx.addCornerArc(x1+radius, y1+radius, radius, 270) // Top-left corner
	ctx.agg2d.ClosePolygon()
}

// Helper method to add a 90-degree corner arc
func (ctx *Context) addCornerArc(cx, cy, radius float64, startAngle float64) {
	// Use bezier curve to approximate 90-degree arc
	const kappa = 0.5522847498307936 // (4/3)*tan(pi/8)

	startRad := startAngle * math.Pi / 180
	endRad := (startAngle + 90) * math.Pi / 180

	x1 := cx + radius*math.Cos(startRad)
	y1 := cy + radius*math.Sin(startRad)
	x4 := cx + radius*math.Cos(endRad)
	y4 := cy + radius*math.Sin(endRad)

	// Calculate control points
	x2 := x1 - kappa*radius*math.Sin(startRad)
	y2 := y1 + kappa*radius*math.Cos(startRad)
	x3 := x4 + kappa*radius*math.Sin(endRad)
	y3 := y4 - kappa*radius*math.Cos(endRad)

	ctx.agg2d.CubicCurveTo(x2, y2, x3, y3, x4, y4)
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
