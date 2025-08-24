// Package agg provides a Go port of the Anti-Grain Geometry (AGG) rendering library.
//
// AGG is a high-quality 2D graphics library that provides anti-aliased rendering
// with subpixel accuracy. This Go port maintains the core functionality while
// providing a Go-idiomatic API.
//
// The library is organized into focused, domain-specific modules:
//
//   - colors.go      - Color types and color management
//   - geometry.go    - Geometric primitives (rectangles, points)
//   - transforms.go  - 2D transformations and viewport operations
//   - gradients.go   - Gradient creation and management
//   - images.go      - Image loading, manipulation, and rendering
//   - text.go        - Text rendering and typography
//   - stroke.go      - Stroke attributes and line styling
//   - blending.go    - Blend modes and alpha compositing
//   - paths.go       - Path operations and manipulation
//   - shapes.go      - Shape drawing primitives
//   - context.go     - Main rendering context (primary interface)
//
// Basic usage:
//
//	ctx := agg.NewContext(800, 600)
//	ctx.SetColor(agg.Red)
//	ctx.FillRectangle(100, 100, 200, 150)
//	ctx.SetStrokeWidth(2.0)
//	ctx.DrawCircle(400, 300, 50)
//	ctx.SaveToPNG("output.png")
//
// The Context interface provides a high-level, user-friendly API that wraps
// the lower-level AGG2D functionality. For advanced usage, you can access
// the underlying AGG2D instance through the Context.
package agg

import (
	"math"

	"agg_go/internal/agg2d"
)

// Version information
const (
	Version    = "0.1.0"
	AGGVersion = "2.6"
	BuildDate  = "2024"
)

// Mathematical constants for convenience
const (
	Pi      = math.Pi
	Pi2     = math.Pi * 2
	PiHalf  = math.Pi / 2
	Deg2Rad = math.Pi / 180.0
	Rad2Deg = 180.0 / math.Pi
)

// Path directions
const (
	CW  = agg2d.CW  // Clockwise direction
	CCW = agg2d.CCW // Counter-clockwise direction
)

// Draw path flags
const (
	FillOnly          = agg2d.FillOnly          // Render only fill
	StrokeOnly        = agg2d.StrokeOnly        // Render only stroke
	FillAndStroke     = agg2d.FillAndStroke     // Render both fill and stroke
	FillWithLineColor = agg2d.FillWithLineColor // Fill using line color
)

// Direction type for path operations
type Direction = agg2d.Direction

// DrawPathFlag type for rendering operations
type DrawPathFlag = agg2d.DrawPathFlag

// Clamp constrains a value to a specified range.
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Lerp performs linear interpolation between two values.
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Public type definitions that wrap internal types
type (
	LineCap       int
	LineJoin      int
	ImageFilter   int
	ImageResample int
	TextAlignment int
)

// LineCap constants
const (
	CapButt LineCap = iota
	CapRound
	CapSquare
)

// LineJoin constants
const (
	JoinMiter LineJoin = iota
	JoinRound
	JoinBevel
)

// ImageFilter constants
const (
	FilterNearest ImageFilter = iota
	FilterBilinear
	FilterBicubic
)

// ImageResample constants
const (
	ResampleNearest ImageResample = iota
	ResampleBilinear
	ResampleBicubic
)

// TextAlignment constants
const (
	AlignLeft TextAlignment = iota
	AlignRight
	AlignCenter
	AlignTop    = AlignRight
	AlignBottom = AlignLeft
)

// Agg2D provides the public interface to the AGG2D rendering engine.
// It wraps the internal implementation and exposes only the necessary methods.
type Agg2D struct {
	impl *agg2d.Agg2D
}

// NewAgg2D creates a new AGG2D rendering context.
func NewAgg2D() *Agg2D {
	return &Agg2D{
		impl: agg2d.NewAgg2D(),
	}
}

// Attach attaches a rendering buffer to the AGG2D context.
func (a *Agg2D) Attach(buf []uint8, width, height, stride int) {
	a.impl.Attach(buf, width, height, stride)
}

// ClipBox sets the clipping rectangle.
func (a *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	a.impl.ClipBox(x1, y1, x2, y2)
}

// ClearAll fills the entire buffer with the specified color.
func (a *Agg2D) ClearAll(c Color) {
	// Convert public Color to internal color format
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.ClearAll(internalColor)
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (a *Agg2D) WorldToScreen(x, y *float64) {
	a.impl.WorldToScreen(x, y)
}

// ScreenToWorld transforms screen coordinates to world coordinates.
func (a *Agg2D) ScreenToWorld(x, y *float64) {
	a.impl.ScreenToWorld(x, y)
}

// FillColor sets the fill color.
func (a *Agg2D) FillColor(c Color) {
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.FillColor(internalColor)
}

// LineColor sets the line color.
func (a *Agg2D) LineColor(c Color) {
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.LineColor(internalColor)
}

// LineWidth sets the line width.
func (a *Agg2D) LineWidth(w float64) {
	a.impl.LineWidth(w)
}

// LineCap sets the line cap style.
func (a *Agg2D) LineCap(cap LineCap) {
	a.impl.LineCap(int(cap))
}

// LineJoin sets the line join style.
func (a *Agg2D) LineJoin(join LineJoin) {
	a.impl.LineJoin(int(join))
}

// ResetTransformations resets the transformation matrix to identity.
func (a *Agg2D) ResetTransformations() {
	a.impl.ResetTransformations()
}

// ImageFilter sets the image filtering method.
func (a *Agg2D) ImageFilter(f ImageFilter) {
	a.impl.ImageFilter(int(f))
}

// ImageResample sets the image resampling method.
func (a *Agg2D) ImageResample(r ImageResample) {
	a.impl.ImageResample(int(r))
}

// TextAlignment sets text alignment.
func (a *Agg2D) TextAlignment(alignX, alignY TextAlignment) {
	a.impl.TextAlignment(int(alignX), int(alignY))
}

// Path methods
func (a *Agg2D) ResetPath() {
	a.impl.ResetPath()
}

func (a *Agg2D) MoveTo(x, y float64) {
	a.impl.MoveTo(x, y)
}

func (a *Agg2D) LineTo(x, y float64) {
	a.impl.LineTo(x, y)
}

func (a *Agg2D) AddEllipse(cx, cy, rx, ry float64, dir Direction) {
	a.impl.AddEllipse(cx, cy, rx, ry, dir)
}

func (a *Agg2D) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

func (a *Agg2D) ClosePolygon() {
	a.impl.ClosePolygon()
}

func (a *Agg2D) DrawPath(flag DrawPathFlag) {
	a.impl.DrawPath(flag)
}

// Shape methods
func (a *Agg2D) Line(x1, y1, x2, y2 float64) {
	a.impl.Line(x1, y1, x2, y2)
}

func (a *Agg2D) Rectangle(x1, y1, x2, y2 float64) {
	a.impl.Rectangle(x1, y1, x2, y2)
}

func (a *Agg2D) Ellipse(cx, cy, rx, ry float64) {
	a.impl.Ellipse(cx, cy, rx, ry)
}

// Image type - simplified for now
type Image struct {
	Data   []uint8
	Width  int
	Height int
	Stride int
}

// NewImage creates a new Image wrapper
func NewImage(data []uint8, width, height, stride int) *Image {
	return &Image{
		Data:   data,
		Width:  width,
		Height: height,
		Stride: stride,
	}
}

// Package-level initialization
func init() {
	// Any package-level initialization can go here
}
