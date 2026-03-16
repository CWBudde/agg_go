package agg

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/shapes"
)

// Context provides the main high-level drawing API for typical Go callers.
//
// It manages an RGBA backing buffer, exposes immediate-mode helpers such as
// DrawRectangle and FillCircle, and also allows explicit path construction via
// BeginPath, MoveTo, LineTo, Fill, and Stroke.
//
// The Context keeps the underlying Agg2D instance available through GetAgg2D
// for advanced use cases that need closer parity with the original C++ AGG2D
// interface.
type Context struct {
	agg2d     *Agg2D
	image     *Image
	width     int
	height    int
	lineWidth float64 // Default stroke width used by convenience helpers.
}

// NewContext allocates a new RGBA image buffer and attaches a fresh Agg2D
// renderer to it.
//
// The returned Context owns its backing image. Call GetImage to access the
// pixels or save them with Image helpers such as SaveToPNG.
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

// NewContextForImage creates a Context that renders into an existing Image.
//
// Use this when image allocation is managed elsewhere but you still want the
// higher-level Context API on top of that buffer.
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

// Clear fills the entire attached image with color and clears any current path
// state inside Agg2D.
func (ctx *Context) Clear(color Color) {
	ctx.agg2d.ClearAll(color)
}

// SetColor sets both the fill and stroke colors at once.
//
// This is the simplest way to keep immediate-mode shape helpers rendering with
// the same color for outlines and fills.
func (ctx *Context) SetColor(color Color) {
	ctx.agg2d.FillColor(color)
	ctx.agg2d.LineColor(color)
}

// DrawLine renders a stroked line immediately using the current stroke state.
func (ctx *Context) DrawLine(x1, y1, x2, y2 float64) {
	ctx.agg2d.Line(x1, y1, x2, y2)
}

// DrawThickLine renders a line immediately with a temporary stroke width.
//
// The previous Context stroke width is restored after rendering.
func (ctx *Context) DrawThickLine(x1, y1, x2, y2, width float64) {
	oldWidth := ctx.lineWidth
	ctx.agg2d.LineWidth(width)
	ctx.agg2d.Line(x1, y1, x2, y2)
	ctx.agg2d.LineWidth(oldWidth)
}

// DrawRectangle renders a stroked rectangle immediately.
//
// Unlike the path API, this helper does not require a later Stroke call.
func (ctx *Context) DrawRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillRectangle renders a filled rectangle immediately.
//
// Unlike the path API, this helper does not require a later Fill call.
func (ctx *Context) FillRectangle(x, y, width, height float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.MoveTo(x, y)
	ctx.agg2d.LineTo(x+width, y)
	ctx.agg2d.LineTo(x+width, y+height)
	ctx.agg2d.LineTo(x, y+height)
	ctx.agg2d.ClosePolygon()
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawCircle renders a stroked circle immediately.
func (ctx *Context) DrawCircle(cx, cy, radius float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillCircle renders a filled circle immediately.
func (ctx *Context) FillCircle(cx, cy, radius float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, radius, radius, CCW)
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawEllipse renders a stroked ellipse immediately.
func (ctx *Context) DrawEllipse(cx, cy, rx, ry float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, rx, ry, CCW)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillEllipse renders a filled ellipse immediately.
func (ctx *Context) FillEllipse(cx, cy, rx, ry float64) {
	ctx.agg2d.ResetPath()
	ctx.agg2d.AddEllipse(cx, cy, rx, ry, CCW)
	ctx.agg2d.DrawPath(FillOnly)
}

// DrawRoundedRectangle renders a stroked rounded rectangle immediately.
func (ctx *Context) DrawRoundedRectangle(x, y, width, height, radius float64) {
	x2 := x + width
	y2 := y + height
	ctx.agg2d.ResetPath()
	ctx.drawRoundedRectPath(x, y, x2, y2, radius)
	ctx.agg2d.DrawPath(StrokeOnly)
}

// FillRoundedRectangle renders a filled rounded rectangle immediately.
func (ctx *Context) FillRoundedRectangle(x, y, width, height, radius float64) {
	x2 := x + width
	y2 := y + height
	ctx.agg2d.ResetPath()
	ctx.drawRoundedRectPath(x, y, x2, y2, radius)
	ctx.agg2d.DrawPath(FillOnly)
}

// drawRoundedRectPath appends a rounded-rectangle outline to the current path.
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

// Fill rasterizes the current path using the current fill state.
//
// Call BeginPath first when constructing geometry manually with MoveTo, LineTo,
// and ClosePath.
func (ctx *Context) Fill() {
	ctx.agg2d.DrawPath(FillOnly)
}

// Stroke rasterizes the current path using the current stroke state.
//
// Call BeginPath first when constructing geometry manually with MoveTo, LineTo,
// and ClosePath.
func (ctx *Context) Stroke() {
	ctx.agg2d.DrawPath(StrokeOnly)
}

// GetImage returns the backing image owned or attached by the Context.
//
// The returned image shares memory with the Context, so subsequent drawing
// operations update the same pixel buffer.
func (ctx *Context) GetImage() *Image {
	return ctx.image
}

// GetAgg2D exposes the underlying Agg2D renderer for advanced operations that
// are not surfaced directly on Context.
func (ctx *Context) GetAgg2D() *Agg2D {
	return ctx.agg2d
}

// SetLineWidth sets the stroke width used for subsequent stroked operations.
func (ctx *Context) SetLineWidth(width float64) {
	ctx.agg2d.LineWidth(width)
}

// BeginPath clears the current path so new path commands start from an empty
// shape.
func (ctx *Context) BeginPath() {
	ctx.agg2d.ResetPath()
}

// MoveTo starts a new subpath at x, y without drawing a segment.
func (ctx *Context) MoveTo(x, y float64) {
	ctx.agg2d.MoveTo(x, y)
}

// LineTo appends a straight segment from the current point to x, y.
func (ctx *Context) LineTo(x, y float64) {
	ctx.agg2d.LineTo(x, y)
}

// ClosePath closes the current contour by connecting the last point back to the
// first point in the subpath.
func (ctx *Context) ClosePath() {
	ctx.agg2d.ClosePolygon()
}
