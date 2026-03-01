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
	"fmt"
	"math"
	"os"

	"agg_go/internal/agg2d"
	"agg_go/internal/color"
	aggimage "agg_go/internal/image"
	"agg_go/internal/rasterizer"
	renscan "agg_go/internal/renderer/scanline"
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

// Deg2RadFunc converts degrees to radians.
func Deg2RadFunc(degrees float64) float64 {
	return degrees * Deg2Rad
}

// Public type definitions mirror the internal AGG2D enums so numeric values
// stay compatible with upstream when passed through the wrapper.
type (
	LineCap       = agg2d.LineCap
	LineJoin      = agg2d.LineJoin
	ImageResample = agg2d.ImageResample
	TextAlignment = agg2d.TextAlignment
)

// LineCap constants
const (
	CapButt   LineCap = agg2d.CapButt
	CapSquare LineCap = agg2d.CapSquare
	CapRound  LineCap = agg2d.CapRound
)

// LineJoin constants
const (
	JoinMiter LineJoin = agg2d.JoinMiter
	JoinRound LineJoin = agg2d.JoinRound
	JoinBevel LineJoin = agg2d.JoinBevel
)

// Backward-compatible aliases kept for the old agg.go API naming.
const (
	ResampleNearest  ImageResample = agg2d.NoResample
	ResampleBilinear ImageResample = agg2d.ResampleAlways
	ResampleBicubic  ImageResample = agg2d.ResampleOnZoomOut
)

// TextAlignment constants
const (
	AlignLeft   TextAlignment = agg2d.AlignLeft
	AlignRight  TextAlignment = agg2d.AlignRight
	AlignCenter TextAlignment = agg2d.AlignCenter
	AlignTop    TextAlignment = agg2d.AlignTop
	AlignBottom TextAlignment = agg2d.AlignBottom
)

// Agg2D provides the public interface to the AGG2D rendering engine.
// It wraps the internal implementation and exposes only the necessary methods.
type Agg2D struct {
	impl           *agg2d.Agg2D
	attachedBuffer []uint8
	attachedWidth  int
	attachedHeight int
	attachedStride int
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
	a.attachedBuffer = buf
	a.attachedWidth = width
	a.attachedHeight = height
	a.attachedStride = stride
}

// ClipBox sets the clipping rectangle.
func (a *Agg2D) ClipBox(x1, y1, x2, y2 float64) {
	a.impl.ClipBox(x1, y1, x2, y2)
}

func (a *Agg2D) GetClipBox() (x1, y1, x2, y2 float64) {
	return a.impl.GetClipBox()
}

// ClearAll fills the entire buffer with the specified color and resets the current path.
func (a *Agg2D) ClearAll(c Color) {
	// Convert public Color to internal color format
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.ClearAll(internalColor)
	a.impl.ResetPath()
}

func (a *Agg2D) ClearClipBox(c Color) {
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.ClearClipBox(internalColor)
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (a *Agg2D) WorldToScreen(x, y *float64) {
	a.impl.WorldToScreen(x, y)
}

// ScreenToWorld transforms screen coordinates to world coordinates.
func (a *Agg2D) ScreenToWorld(x, y *float64) {
	a.impl.ScreenToWorld(x, y)
}

func (a *Agg2D) WorldToScreenScalar(scalar float64) float64 {
	return a.impl.WorldToScreenScalar(scalar)
}

func (a *Agg2D) ScreenToWorldScalar(scalar float64) float64 {
	return a.impl.ScreenToWorldScalar(scalar)
}

func (a *Agg2D) AlignPoint(x, y *float64) {
	a.impl.AlignPoint(x, y)
}

func (a *Agg2D) InBox(worldX, worldY float64) bool {
	return a.impl.InBox(worldX, worldY)
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

func (a *Agg2D) GetLineWidth() float64 {
	return a.impl.GetLineWidth()
}

func (a *Agg2D) GetLineCap() LineCap {
	return a.impl.GetLineCap()
}

func (a *Agg2D) GetLineJoin() LineJoin {
	return a.impl.GetLineJoin()
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

func (a *Agg2D) TransformImageParallelogramSimple(img *Image, parallelogram []float64) error {
	return a.impl.TransformImageParallelogramSimple(img.ToInternalImage(), parallelogram)
}

// ResetTransformations resets the transformation matrix to identity.
func (a *Agg2D) ResetTransformations() {
	a.impl.ResetTransformations()
}

func (a *Agg2D) GetTransformations() *Transformations {
	return fromInternalTransformations(a.impl.GetTransformations())
}

func (a *Agg2D) SetTransformations(tr *Transformations) {
	a.impl.SetTransformations(toInternalTransformations(tr))
}

func (a *Agg2D) Affine(tr *Transformations) {
	a.impl.AffineFromMatrix(toInternalTransformations(tr))
}

func (a *Agg2D) Rotate(angle float64) {
	a.impl.Rotate(angle)
}

func (a *Agg2D) Scale(sx, sy float64) {
	a.impl.Scale(sx, sy)
}

func (a *Agg2D) Skew(sx, sy float64) {
	a.impl.Skew(sx, sy)
}

func (a *Agg2D) Translate(x, y float64) {
	a.impl.Translate(x, y)
}

func (a *Agg2D) PushTransform() {
	a.impl.PushTransform()
}

func (a *Agg2D) PopTransform() bool {
	return a.impl.PopTransform()
}

func (a *Agg2D) WorldToScreenDistance(worldDistance float64) float64 {
	return a.impl.WorldToScreenDistance(worldDistance)
}

func (a *Agg2D) ScreenToWorldDistance(screenDistance float64) (float64, bool) {
	return a.impl.ScreenToWorldDistance(screenDistance)
}

// ImageFilter type
type ImageFilter = agg2d.ImageFilter

// ImageFilter constants for image interpolation.
const (
	FilterNoFilter ImageFilter = agg2d.NoFilter
	FilterBilinear ImageFilter = agg2d.Bilinear
	FilterHanning  ImageFilter = agg2d.Hanning
	FilterHermite  ImageFilter = agg2d.Hermite
	FilterQuadric  ImageFilter = agg2d.Quadric
	FilterBicubic  ImageFilter = agg2d.Bicubic
	FilterCatrom   ImageFilter = agg2d.Catrom
	FilterSpline16 ImageFilter = agg2d.Spline16
	FilterSpline36 ImageFilter = agg2d.Spline36
	FilterBlackman ImageFilter = agg2d.Blackman
	FilterHamming  ImageFilter = agg2d.Blackman + 1
	FilterMitchell ImageFilter = agg2d.Blackman + 2
	FilterGaussian ImageFilter = agg2d.Blackman + 3
	FilterBessel   ImageFilter = agg2d.Blackman + 4
	FilterSinc     ImageFilter = agg2d.Blackman + 5
	FilterLanczos  ImageFilter = agg2d.Blackman + 6
)

// ImageFilter sets the image filtering method using a predefined filter type.
func (a *Agg2D) ImageFilter(ft ImageFilter) {
	var f aggimage.FilterFunction
	switch ft {
	case FilterBilinear:
		f = aggimage.BilinearFilter{}
	case FilterHanning:
		f = aggimage.HanningFilter{}
	case FilterHamming:
		f = aggimage.HammingFilter{}
	case FilterHermite:
		f = aggimage.HermiteFilter{}
	case FilterQuadric:
		f = aggimage.QuadricFilter{}
	case FilterBicubic:
		f = aggimage.BicubicFilter{}
	case FilterCatrom:
		f = aggimage.CatromFilter{}
	case FilterMitchell:
		f = aggimage.NewMitchellFilter(1.0/3.0, 1.0/3.0)
	case FilterSpline16:
		f = aggimage.Spline16Filter{}
	case FilterSpline36:
		f = aggimage.Spline36Filter{}
	case FilterGaussian:
		f = aggimage.GaussianFilter{}
	case FilterBessel:
		f = aggimage.BesselFilter{}
	case FilterSinc:
		f = aggimage.NewSincFilter(4.0)
	case FilterLanczos:
		f = aggimage.NewLanczosFilter(4.0)
	case FilterBlackman:
		f = aggimage.NewBlackmanFilter(4.0)
	default:
		f = aggimage.BilinearFilter{}
	}
	a.impl.SetImageFilterLUT(aggimage.NewImageFilterLUTWithFilter(f, true))
}

// SetImageFilterRadius sets the image filtering method with a custom radius for supported filters.
func (a *Agg2D) SetImageFilterRadius(ft ImageFilter, radius float64) {
	var f aggimage.FilterFunction
	switch ft {
	case FilterSinc:
		f = aggimage.NewSincFilter(radius)
	case FilterLanczos:
		f = aggimage.NewLanczosFilter(radius)
	case FilterBlackman:
		f = aggimage.NewBlackmanFilter(radius)
	default:
		a.ImageFilter(ft)
		return
	}
	a.impl.SetImageFilterLUT(aggimage.NewImageFilterLUTWithFilter(f, true))
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

func (a *Agg2D) MoveRel(dx, dy float64) {
	a.impl.MoveRel(dx, dy)
}

func (a *Agg2D) LineTo(x, y float64) {
	a.impl.LineTo(x, y)
}

func (a *Agg2D) HorLineTo(x float64) {
	a.impl.HorLineTo(x)
}

func (a *Agg2D) AddEllipse(cx, cy, rx, ry float64, dir Direction) {
	a.impl.AddEllipse(cx, cy, rx, ry, dir)
}

func (a *Agg2D) VerLineTo(y float64) {
	a.impl.VerLineTo(y)
}

func (a *Agg2D) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	a.impl.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, x, y)
}

func (a *Agg2D) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	a.impl.QuadricCurveTo(xCtrl, yCtrl, xTo, yTo)
}

func (a *Agg2D) QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo float64) {
	a.impl.QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo)
}

func (a *Agg2D) QuadricCurveToSmooth(xTo, yTo float64) {
	a.impl.QuadricCurveToSmooth(xTo, yTo)
}

func (a *Agg2D) QuadricCurveRelSmooth(dxTo, dyTo float64) {
	a.impl.QuadricCurveRelSmooth(dxTo, dyTo)
}

func (a *Agg2D) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

func (a *Agg2D) CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	a.impl.CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo)
}

func (a *Agg2D) CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo)
}

func (a *Agg2D) CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	a.impl.CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo)
}

func (a *Agg2D) ClosePolygon() {
	a.impl.ClosePolygon()
}

func (a *Agg2D) DrawPath(flag DrawPathFlag) {
	a.impl.DrawPath(flag)
}

func (a *Agg2D) DrawPathNoTransform(flag DrawPathFlag) {
	a.impl.DrawPathNoTransform(flag)
}

// Shape methods
func (a *Agg2D) Line(x1, y1, x2, y2 float64) {
	a.impl.Line(x1, y1, x2, y2)
}

func (a *Agg2D) Triangle(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Triangle(x1, y1, x2, y2, x3, y3)
}

func (a *Agg2D) Rectangle(x1, y1, x2, y2 float64) {
	a.impl.Rectangle(x1, y1, x2, y2)
}

func (a *Agg2D) Ellipse(cx, cy, rx, ry float64) {
	a.impl.Ellipse(cx, cy, rx, ry)
}

// RoundedRect draws a rounded rectangle.
func (a *Agg2D) RoundedRect(x1, y1, x2, y2, r float64) {
	a.impl.RoundedRect(x1, y1, x2, y2, r)
}

func (a *Agg2D) RoundedRectXY(x1, y1, x2, y2, rx, ry float64) {
	a.impl.RoundedRectXY(x1, y1, x2, y2, rx, ry)
}

func (a *Agg2D) RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop float64) {
	a.impl.RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop)
}

func (a *Agg2D) Arc(cx, cy, rx, ry, start, sweep float64) {
	a.impl.Arc(cx, cy, rx, ry, start, sweep)
}

func (a *Agg2D) Star(cx, cy, r1, r2, startAngle float64, numRays int) {
	a.impl.Star(cx, cy, r1, r2, startAngle, numRays)
}

func (a *Agg2D) Curve(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Curve(x1, y1, x2, y2, x3, y3)
}

func (a *Agg2D) Curve4(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	a.impl.Curve4(x1, y1, x2, y2, x3, y3, x4, y4)
}

func (a *Agg2D) Polygon(xy []float64, numPoints int) {
	a.impl.Polygon(xy, numPoints)
}

func (a *Agg2D) Polyline(xy []float64, numPoints int) {
	a.impl.Polyline(xy, numPoints)
}

func (a *Agg2D) DrawCircle(cx, cy, radius float64) {
	a.impl.DrawCircle(cx, cy, radius)
}

func (a *Agg2D) FillCircle(cx, cy, radius float64) {
	a.impl.FillCircle(cx, cy, radius)
}

// Path relative methods
func (a *Agg2D) LineRel(dx, dy float64) {
	a.impl.LineRel(dx, dy)
}

func (a *Agg2D) HorLineRel(dx float64) {
	a.impl.HorLineRel(dx)
}

func (a *Agg2D) VerLineRel(dy float64) {
	a.impl.VerLineRel(dy)
}

func (a *Agg2D) ArcRel(rx, ry, angle float64, largeArcFlag, sweepFlag bool, dx, dy float64) {
	a.impl.ArcRel(rx, ry, angle, largeArcFlag, sweepFlag, dx, dy)
}

// Viewport sets the viewport transformation.
func (a *Agg2D) Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption) {
	a.impl.Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, int(opt))
}

// Text methods
func (a *Agg2D) Font(fontName string, height float64, bold, italic bool, cacheType FontCacheType, angle float64) error {
	return a.impl.Font(fontName, height, bold, italic, cacheType, angle)
}

func (a *Agg2D) FontHeight() float64 {
	return a.impl.FontHeight()
}

func (a *Agg2D) FlipText(flip bool) {
	a.impl.FlipText(flip)
}

func (a *Agg2D) TextHints(hints bool) {
	a.impl.TextHints(hints)
}

func (a *Agg2D) GetTextHints() bool {
	return a.impl.GetTextHints()
}

func (a *Agg2D) MiterLimit(ml float64) {
	a.impl.MiterLimit(ml)
}

func (a *Agg2D) GetMiterLimit() float64 {
	return a.impl.GetMiterLimit()
}

func (a *Agg2D) AddDash(dashLen, gapLen float64) {
	a.impl.AddDash(dashLen, gapLen)
}

func (a *Agg2D) RemoveAllDashes() {
	a.impl.RemoveAllDashes()
}

func (a *Agg2D) DashStart(offset float64) {
	a.impl.DashStart(offset)
}

func (a *Agg2D) GetDashStart() float64 {
	return a.impl.GetDashStart()
}

func (a *Agg2D) NoDashes() {
	a.impl.NoDashes()
}

func (a *Agg2D) FillEvenOdd(evenOddFlag bool) {
	a.impl.FillEvenOdd(evenOddFlag)
}

func (a *Agg2D) GetFillEvenOdd() bool {
	return a.impl.GetFillEvenOdd()
}

func (a *Agg2D) IsEvenOddFillRule() bool {
	return a.impl.IsEvenOddFillRule()
}

func (a *Agg2D) IsNonZeroFillRule() bool {
	return a.impl.IsNonZeroFillRule()
}

func (a *Agg2D) FillRuleDescription() string {
	return a.impl.FillRuleDescription()
}

func (a *Agg2D) Text(x, y float64, str string, roundOff bool, dx, dy float64) {
	a.impl.Text(x, y, str, roundOff, dx, dy)
}

func (a *Agg2D) TextWidth(str string) float64 {
	return a.impl.TextWidth(str)
}

// Blend mode methods
func (a *Agg2D) BlendMode(mode BlendMode) {
	a.impl.SetBlendMode(mode)
}

func (a *Agg2D) GetBlendMode() BlendMode {
	return a.impl.GetBlendMode()
}

func (a *Agg2D) ImageBlendMode(mode BlendMode) {
	a.impl.SetImageBlendMode(mode)
}

func (a *Agg2D) GetImageBlendMode() BlendMode {
	return a.impl.GetImageBlendMode()
}

func (a *Agg2D) ImageBlendColor(c Color) {
	internalColor := [4]uint8{c.R, c.G, c.B, c.A}
	a.impl.SetImageBlendColor(internalColor)
}

func (a *Agg2D) GetImageBlendColor() Color {
	c := a.impl.GetImageBlendColor()
	return Color{R: c[0], G: c[1], B: c[2], A: c[3]}
}

// Master alpha methods
func (a *Agg2D) MasterAlpha(alpha float64) {
	a.impl.SetMasterAlpha(alpha)
}

func (a *Agg2D) GetMasterAlpha() float64 {
	return a.impl.GetMasterAlpha()
}

func (a *Agg2D) AntiAliasGamma(gamma float64) {
	a.impl.SetAntiAliasGamma(gamma)
}

func (a *Agg2D) GetAntiAliasGamma() float64 {
	return a.impl.GetAntiAliasGamma()
}

// Utility methods
func (a *Agg2D) NoFill() {
	a.impl.NoFill()
}

func (a *Agg2D) NoLine() {
	a.impl.NoLine()
}

func (a *Agg2D) SaveImagePPM(filename string) error {
	if a.attachedBuffer == nil || a.attachedWidth <= 0 || a.attachedHeight <= 0 || a.attachedStride <= 0 {
		return fmt.Errorf("no attached RGBA buffer")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "P6\n%d %d\n255\n", a.attachedWidth, a.attachedHeight); err != nil {
		return err
	}

	rgbRow := make([]byte, a.attachedWidth*3)
	for y := 0; y < a.attachedHeight; y++ {
		rowStart := y * a.attachedStride
		for x := 0; x < a.attachedWidth; x++ {
			src := rowStart + x*4
			dst := x * 3
			if src+2 >= len(a.attachedBuffer) {
				return fmt.Errorf("attached buffer too small for %dx%d image", a.attachedWidth, a.attachedHeight)
			}
			rgbRow[dst] = a.attachedBuffer[src]
			rgbRow[dst+1] = a.attachedBuffer[src+1]
			rgbRow[dst+2] = a.attachedBuffer[src+2]
		}
		if _, err := file.Write(rgbRow); err != nil {
			return err
		}
	}

	return nil
}

// Image transformation methods
func (a *Agg2D) TransformImagePath(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	internalImg := img.ToInternalImage()
	return a.impl.TransformImagePath(internalImg, imgX1, imgY1, imgX2, imgY2, dstX1, dstY1, dstX2, dstY2)
}

func (a *Agg2D) TransformImagePathSimple(img *Image, dstX1, dstY1, dstX2, dstY2 float64) error {
	return a.impl.TransformImagePathSimple(img.ToInternalImage(), dstX1, dstY1, dstX2, dstY2)
}

func (a *Agg2D) TransformImagePathParallelogram(img *Image, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	return a.impl.TransformImagePathParallelogram(img.ToInternalImage(), imgX1, imgY1, imgX2, imgY2, parallelogram)
}

func (a *Agg2D) TransformImagePathParallelogramSimple(img *Image, parallelogram []float64) error {
	return a.impl.TransformImagePathParallelogramSimple(img.ToInternalImage(), parallelogram)
}

func (a *Agg2D) BlendImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX, dstY float64, alpha uint) error {
	return a.impl.BlendImage(img.ToInternalImage(), imgX1, imgY1, imgX2, imgY2, dstX, dstY, alpha)
}

func (a *Agg2D) BlendImageSimple(img *Image, dstX, dstY float64, alpha uint) error {
	return a.impl.BlendImageSimple(img.ToInternalImage(), dstX, dstY, alpha)
}

func (a *Agg2D) CopyImage(img *Image, imgX1, imgY1, imgX2, imgY2 int, dstX, dstY float64) error {
	return a.impl.CopyImage(img.ToInternalImage(), imgX1, imgY1, imgX2, imgY2, dstX, dstY)
}

func (a *Agg2D) CopyImageSimple(img *Image, dstX, dstY float64) error {
	return a.impl.CopyImageSimple(img.ToInternalImage(), dstX, dstY)
}

// GetInternalRasterizer returns the underlying rasterizer for advanced usage.
func (a *Agg2D) GetInternalRasterizer() *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip] {
	return a.impl.GetInternalRasterizer()
}

// ScanlineRender renders the current rasterizer data using a custom renderer.
func (a *Agg2D) ScanlineRender(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], renderer renscan.RendererInterface[color.RGBA8[color.Linear]]) {
	a.impl.ScanlineRender(ras, renderer)
}

// GouraudTriangle renders a Gouraud-shaded triangle.
func (a *Agg2D) GouraudTriangle(x1, y1, x2, y2, x3, y3 float64, c1, c2, c3 Color, d float64) {
	internalC1 := [4]uint8{c1.R, c1.G, c1.B, c1.A}
	internalC2 := [4]uint8{c2.R, c2.G, c2.B, c2.A}
	internalC3 := [4]uint8{c3.R, c3.G, c3.B, c3.A}
	a.impl.GouraudTriangle(x1, y1, x2, y2, x3, y3, internalC1, internalC2, internalC3, d)
}

// Package-level initialization
func init() {
	// Any package-level initialization can go here
}
