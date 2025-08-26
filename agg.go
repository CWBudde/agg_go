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
func (a *Agg2D) LineCap(capStyle LineCap) {
	a.impl.LineCap(int(capStyle))
}

// LineJoin sets the line join style.
func (a *Agg2D) LineJoin(join LineJoin) {
	a.impl.LineJoin(int(join))
}

// FillLinearGradient sets up a linear gradient for fill operations.
func (a *Agg2D) FillLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	a.impl.FillLinearGradient(x1, y1, x2, y2, internalC1, internalC2, profile)
}

// FillRadialGradient sets up a radial gradient for fill operations.
func (a *Agg2D) FillRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	a.impl.FillRadialGradient(x, y, r, internalC1, internalC2, profile)
}

// FillRadialGradientMultiStop sets up a radial gradient with three colors.
func (a *Agg2D) FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	internalC3 := [4]uint8{c3.R, c3.G, c3.B, c3.A}
	a.impl.FillRadialGradientMultiStop(x, y, r, internalC1, internalC2, internalC3)
}

// LineLinearGradient sets up a linear gradient for line/stroke operations.
func (a *Agg2D) LineLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	a.impl.LineLinearGradient(x1, y1, x2, y2, internalC1, internalC2, profile)
}

// LineRadialGradient sets up a radial gradient for line/stroke operations.
func (a *Agg2D) LineRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	a.impl.LineRadialGradient(x, y, r, internalC1, internalC2, profile)
}

// LineRadialGradientMultiStop sets up a radial gradient with three colors for line operations.
func (a *Agg2D) LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	internalC3 := [4]uint8{c3.R, c3.G, c3.B, c3.A}
	a.impl.LineRadialGradientMultiStop(x, y, r, internalC1, internalC2, internalC3)
}

// FillGradientFlag returns the current fill gradient type.
func (a *Agg2D) FillGradientFlag() int {
	return a.impl.FillGradientFlag()
}

// LineGradientFlag returns the current line gradient type.
func (a *Agg2D) LineGradientFlag() int {
	return a.impl.LineGradientFlag()
}

// FillGradientD1 returns the first bound of the current fill gradient.
func (a *Agg2D) FillGradientD1() float64 {
	return a.impl.FillGradientD1()
}

// FillGradientD2 returns the second bound of the current fill gradient.
func (a *Agg2D) FillGradientD2() float64 {
	return a.impl.FillGradientD2()
}

// LineGradientD1 returns the first bound of the current line gradient.
func (a *Agg2D) LineGradientD1() float64 {
	return a.impl.LineGradientD1()
}

// LineGradientD2 returns the second bound of the current line gradient.
func (a *Agg2D) LineGradientD2() float64 {
	return a.impl.LineGradientD2()
}

// TransformImage transforms and renders an image with source and destination rectangles.
func (a *Agg2D) TransformImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	return a.impl.TransformImage(img.ToInternalImage(), imgX1, imgY1, imgX2, imgY2, dstX1, dstY1, dstX2, dstY2)
}

// TransformImageSimple transforms and renders entire image to destination rectangle.
func (a *Agg2D) TransformImageSimple(img *Image, dstX1, dstY1, dstX2, dstY2 float64) error {
	return a.impl.TransformImageSimple(img.ToInternalImage(), dstX1, dstY1, dstX2, dstY2)
}

// TransformImageParallelogram transforms and renders an image using a parallelogram.
func (a *Agg2D) TransformImageParallelogram(img *Image, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	return a.impl.TransformImageParallelogram(img.ToInternalImage(), imgX1, imgY1, imgX2, imgY2, parallelogram)
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

// GetImageFilter returns the current image filtering method.
func (a *Agg2D) GetImageFilter() ImageFilter {
	return ImageFilter(a.impl.GetImageFilter())
}

// GetImageResample returns the current image resampling method.
func (a *Agg2D) GetImageResample() ImageResample {
	return ImageResample(a.impl.GetImageResample())
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

// Package-level initialization
func init() {
	// Any package-level initialization can go here
}
