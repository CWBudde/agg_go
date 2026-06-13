package agg

import (
	"errors"
	"image"

	"github.com/cwbudde/agg_go/internal/agg2d"
	"github.com/cwbudde/agg_go/internal/color"
)

// errNilImageFloat is returned by the float image-transform methods when the
// source image is nil.
var errNilImageFloat = errors.New("agg: image is nil")

// ImageFloat is the public float (128-bit, 4 x float32) raster image used as a
// rendering target or image source for Agg2DFloat. It stores straight
// (non-premultiplied) RGBA float32 data and exposes boundary conversions to
// standard Go image types.
type ImageFloat struct {
	impl *agg2d.ImageFloat
}

// NewImageFloat allocates a zeroed float image of the given dimensions.
func NewImageFloat(width, height int) *ImageFloat {
	return &ImageFloat{impl: agg2d.NewImageFloatEmpty(width, height)}
}

// Width returns the image width in pixels.
func (img *ImageFloat) Width() int { return img.impl.Width() }

// Height returns the image height in pixels.
func (img *ImageFloat) Height() int { return img.impl.Height() }

// GetPixelFloat returns the straight RGBA float components at (x,y); out of
// bounds returns zeros.
func (img *ImageFloat) GetPixelFloat(x, y int) (r, g, b, a float32) {
	c := img.impl.GetPixel(x, y)
	return c.R, c.G, c.B, c.A
}

// SetPixelFloat writes straight RGBA float components at (x,y).
func (img *ImageFloat) SetPixelFloat(x, y int, r, g, b, a float32) {
	img.impl.SetPixel(x, y, color.RGBA32[color.Linear]{R: r, G: g, B: b, A: a})
}

// Premultiply converts the image in place to premultiplied alpha.
func (img *ImageFloat) Premultiply() { img.impl.Premultiply() }

// Demultiply converts the image in place to straight alpha.
func (img *ImageFloat) Demultiply() { img.impl.Demultiply() }

// ToRGBA returns a Go image.RGBA (alpha-premultiplied, 8-bit) copy.
func (img *ImageFloat) ToRGBA() *image.RGBA { return img.impl.ToRGBA() }

// ToNRGBA64 returns a Go image.NRGBA64 (straight, 16-bit) copy.
func (img *ImageFloat) ToNRGBA64() *image.NRGBA64 { return img.impl.ToNRGBA64() }

// Agg2DFloat is the public float-precision twin of Agg2D. The public Color
// stays 8-bit; internally colors flow through the float RGBA32 pixel pipeline.
type Agg2DFloat struct {
	impl *agg2d.Agg2DFloat
}

// NewAgg2DFloat creates a new float-precision AGG2D rendering context.
func NewAgg2DFloat() *Agg2DFloat {
	return &Agg2DFloat{impl: agg2d.NewAgg2DFloat()}
}

// GetImpl exposes the internal float renderer for advanced operations.
func (a *Agg2DFloat) GetImpl() *agg2d.Agg2DFloat { return a.impl }

// Attach attaches a float rendering buffer. stride is in bytes per row.
func (a *Agg2DFloat) Attach(buf []float32, width, height, stride int) {
	a.impl.Attach(buf, width, height, stride)
}

// AttachImage attaches the rendering context to an existing float image.
func (a *Agg2DFloat) AttachImage(img *ImageFloat) {
	if img == nil {
		return
	}
	a.impl.AttachImageFloat(img.impl)
}

func toInternalColor(c Color) agg2d.Color { return agg2d.Color{c.R, c.G, c.B, c.A} }

// ClearAll fills the entire buffer with the specified color.
func (a *Agg2DFloat) ClearAll(c Color) { a.impl.ClearAll(toInternalColor(c)) }

// ClipBox sets the clipping rectangle.
func (a *Agg2DFloat) ClipBox(x1, y1, x2, y2 float64) { a.impl.ClipBox(x1, y1, x2, y2) }

// GetBounds returns the current rendering bounds.
func (a *Agg2DFloat) GetBounds() (x1, y1, x2, y2 float64) {
	b := a.impl.GetBounds()
	return b.X1, b.Y1, b.X2, b.Y2
}

// FillColor sets the fill color.
func (a *Agg2DFloat) FillColor(c Color) { a.impl.FillColor(toInternalColor(c)) }

// LineColor sets the line color.
func (a *Agg2DFloat) LineColor(c Color) { a.impl.LineColor(toInternalColor(c)) }

// LineWidth sets the line width.
func (a *Agg2DFloat) LineWidth(w float64) { a.impl.LineWidth(w) }

// GetLineWidth returns the current line width.
func (a *Agg2DFloat) GetLineWidth() float64 { return a.impl.GetLineWidth() }

// LineCap sets the line cap style.
func (a *Agg2DFloat) LineCap(lineCap LineCap) { a.impl.LineCap(lineCap) }

// LineJoin sets the line join style.
func (a *Agg2DFloat) LineJoin(join LineJoin) { a.impl.LineJoin(join) }

// AddDash appends one dash-gap pair to the current dash pattern.
func (a *Agg2DFloat) AddDash(dashLen, gapLen float64) { a.impl.AddDash(dashLen, gapLen) }

// RemoveAllDashes clears every dash pattern segment.
func (a *Agg2DFloat) RemoveAllDashes() { a.impl.RemoveAllDashes() }

// DashStart sets the dash-phase offset.
func (a *Agg2DFloat) DashStart(offset float64) { a.impl.DashStart(offset) }

// GetDashStart returns the current dash-phase offset.
func (a *Agg2DFloat) GetDashStart() float64 { return a.impl.GetDashStart() }

// NoDashes disables dashed stroke rendering.
func (a *Agg2DFloat) NoDashes() { a.impl.NoDashes() }

// SetMasterAlpha sets the master alpha.
func (a *Agg2DFloat) SetMasterAlpha(alpha float64) { a.impl.SetMasterAlpha(alpha) }

// GetMasterAlpha returns the current master alpha.
func (a *Agg2DFloat) GetMasterAlpha() float64 { return a.impl.GetMasterAlpha() }

// SetAntiAliasGamma sets the anti-alias gamma.
func (a *Agg2DFloat) SetAntiAliasGamma(g float64) { a.impl.SetAntiAliasGamma(g) }

// TextAlignment sets text alignment state.
func (a *Agg2DFloat) TextAlignment(alignX, alignY TextAlignment) {
	a.impl.TextAlignment(alignX, alignY)
}

// FlipText sets the text flip state.
func (a *Agg2DFloat) FlipText(flip bool) { a.impl.FlipText(flip) }

// --- Text rendering ---

// Font loads and activates a font file for subsequent text rendering, mirroring
// the 8-bit Agg2D.Font. Requires cgo/FreeType; for WASM use FontGSV instead.
func (a *Agg2DFloat) Font(fontName string, height float64, bold, italic bool, cacheType FontCacheType, angle float64) error {
	return a.impl.Font(fontName, height, bold, italic, cacheType, angle)
}

// FontDefault loads a font using the upstream defaults: non-bold, non-italic,
// raster cache, zero angle.
func (a *Agg2DFloat) FontDefault(fontName string, height float64) error {
	return a.Font(fontName, height, false, false, RasterFontCache, 0.0)
}

// FontHeight returns the configured font height in world units.
func (a *Agg2DFloat) FontHeight() float64 { return a.impl.FontHeight() }

// FontGSV configures the built-in AGG GSV stroke-vector font as the active text
// backend. No external font file is required; it works in WASM builds.
func (a *Agg2DFloat) FontGSV(height float64) { a.impl.FontGSV(height) }

// SetResolution sets the font rendering resolution in DPI for FreeType text.
func (a *Agg2DFloat) SetResolution(dpi uint) { a.impl.SetResolution(dpi) }

// TextHints enables or disables font hinting.
func (a *Agg2DFloat) TextHints(hints bool) { a.impl.TextHints(hints) }

// TextForceAutohint enables or disables FreeType's auto-hinter for raster text.
func (a *Agg2DFloat) TextForceAutohint(force bool) { a.impl.TextForceAutohint(force) }

// GetTextHints reports whether text hinting is enabled.
func (a *Agg2DFloat) GetTextHints() bool { return a.impl.GetTextHints() }

// GetAscender returns the configured font ascender in world units.
func (a *Agg2DFloat) GetAscender() float64 { return a.impl.GetAscender() }

// GetDescender returns the configured font descender in world units.
func (a *Agg2DFloat) GetDescender() float64 { return a.impl.GetDescender() }

// MeasureText returns the width and height of str for the current font settings.
func (a *Agg2DFloat) MeasureText(str string) (width, height float64) {
	return a.impl.MeasureText(str)
}

// GetTextHeight returns the nominal height of the current font.
func (a *Agg2DFloat) GetTextHeight() float64 { return a.impl.GetTextHeight() }

// Text renders str at x, y with optional rounding and offset adjustments.
func (a *Agg2DFloat) Text(x, y float64, str string, roundOff bool, dx, dy float64) {
	a.impl.Text(x, y, str, roundOff, dx, dy)
}

// TextDefault renders text using the upstream defaults: no round-off, zero offsets.
func (a *Agg2DFloat) TextDefault(x, y float64, str string) {
	a.Text(x, y, str, false, 0.0, 0.0)
}

// TextWidth measures str using the active font backend and cache mode.
func (a *Agg2DFloat) TextWidth(str string) float64 { return a.impl.TextWidth(str) }

// GetTextBounds returns the actual ink bounds of str relative to the baseline origin.
func (a *Agg2DFloat) GetTextBounds(str string) (x, y, width, height float64) {
	return a.impl.GetTextBounds(str)
}

// --- Blend modes ---

// SetBlendMode sets the general blend/composite mode for fills, strokes, and
// gradients. BlendAlpha (the default) uses the plain alpha pipeline; any other
// mode routes rendering through the float composite pixfmt.
func (a *Agg2DFloat) SetBlendMode(mode BlendMode) { a.impl.SetBlendMode(mode) }

// GetBlendMode returns the current general blend/composite mode.
func (a *Agg2DFloat) GetBlendMode() BlendMode { return a.impl.GetBlendMode() }

// SetImageBlendMode sets the blend mode used by image transfer operations.
func (a *Agg2DFloat) SetImageBlendMode(mode BlendMode) { a.impl.SetImageBlendMode(mode) }

// GetImageBlendMode returns the current image blend mode.
func (a *Agg2DFloat) GetImageBlendMode() BlendMode { return a.impl.GetImageBlendMode() }

// SetImageBlendColor sets the image blend color whose alpha scales sampled image
// spans before compositing.
func (a *Agg2DFloat) SetImageBlendColor(c Color) { a.impl.SetImageBlendColor(toInternalColor(c)) }

// GetImageBlendColor returns the current image blend color.
func (a *Agg2DFloat) GetImageBlendColor() Color {
	c := a.impl.GetImageBlendColor()
	return Color{R: c[0], G: c[1], B: c[2], A: c[3]}
}

// --- Path building ---

// ResetPath clears the current path.
func (a *Agg2DFloat) ResetPath() { a.impl.ResetPath() }

// MoveTo moves the current point to absolute coordinates.
func (a *Agg2DFloat) MoveTo(x, y float64) { a.impl.MoveTo(x, y) }

// MoveRel moves the current point by a relative amount.
func (a *Agg2DFloat) MoveRel(dx, dy float64) { a.impl.MoveRel(dx, dy) }

// LineTo draws a line to absolute coordinates.
func (a *Agg2DFloat) LineTo(x, y float64) { a.impl.LineTo(x, y) }

// LineRel draws a line by a relative amount.
func (a *Agg2DFloat) LineRel(dx, dy float64) { a.impl.LineRel(dx, dy) }

// HorLineTo draws a horizontal line to the given x.
func (a *Agg2DFloat) HorLineTo(x float64) { a.impl.HorLineTo(x) }

// HorLineRel draws a horizontal line by a relative amount.
func (a *Agg2DFloat) HorLineRel(dx float64) { a.impl.HorLineRel(dx) }

// VerLineTo draws a vertical line to the given y.
func (a *Agg2DFloat) VerLineTo(y float64) { a.impl.VerLineTo(y) }

// VerLineRel draws a vertical line by a relative amount.
func (a *Agg2DFloat) VerLineRel(dy float64) { a.impl.VerLineRel(dy) }

// ArcTo adds an SVG-style elliptical arc to the path.
func (a *Agg2DFloat) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	a.impl.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, x, y)
}

// QuadricCurveTo adds a quadratic Bézier curve to the path.
func (a *Agg2DFloat) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	a.impl.QuadricCurveTo(xCtrl, yCtrl, xTo, yTo)
}

// QuadricCurveRel adds a quadratic Bézier curve using relative coordinates.
func (a *Agg2DFloat) QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo float64) {
	a.impl.QuadricCurveRel(dxCtrl, dyCtrl, dxTo, dyTo)
}

// QuadricCurveToSmooth adds a smooth quadratic Bézier curve (reflected control point).
func (a *Agg2DFloat) QuadricCurveToSmooth(xTo, yTo float64) {
	a.impl.QuadricCurveToSmooth(xTo, yTo)
}

// QuadricCurveRelSmooth adds a smooth quadratic Bézier curve using relative coordinates.
func (a *Agg2DFloat) QuadricCurveRelSmooth(dxTo, dyTo float64) {
	a.impl.QuadricCurveRelSmooth(dxTo, dyTo)
}

// CubicCurveTo adds a cubic Bézier curve to the path.
func (a *Agg2DFloat) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
}

// CubicCurveRel adds a cubic Bézier curve using relative coordinates.
func (a *Agg2DFloat) CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	a.impl.CubicCurveRel(dxCtrl1, dyCtrl1, dxCtrl2, dyCtrl2, dxTo, dyTo)
}

// CubicCurveToSmooth adds a smooth cubic Bézier curve (reflected first control point).
func (a *Agg2DFloat) CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveToSmooth(xCtrl2, yCtrl2, xTo, yTo)
}

// CubicCurveRelSmooth adds a smooth cubic Bézier curve using relative coordinates.
func (a *Agg2DFloat) CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo float64) {
	a.impl.CubicCurveRelSmooth(dxCtrl2, dyCtrl2, dxTo, dyTo)
}

// AddEllipse appends an ellipse contour to the current path.
func (a *Agg2DFloat) AddEllipse(cx, cy, rx, ry float64, dir Direction) {
	a.impl.AddEllipse(cx, cy, rx, ry, dir)
}

// ClosePolygon closes the current sub-path.
func (a *Agg2DFloat) ClosePolygon() { a.impl.ClosePolygon() }

// DrawPath renders the current path according to the given flag.
func (a *Agg2DFloat) DrawPath(flag DrawPathFlag) { a.impl.DrawPath(flag) }

// DrawPathNoTransform renders the current path with the identity transform.
func (a *Agg2DFloat) DrawPathNoTransform(flag DrawPathFlag) { a.impl.DrawPathNoTransform(flag) }

// --- Shapes ---

// Line strokes a single line segment.
func (a *Agg2DFloat) Line(x1, y1, x2, y2 float64) { a.impl.Line(x1, y1, x2, y2) }

// Triangle fills and strokes a triangle.
func (a *Agg2DFloat) Triangle(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Triangle(x1, y1, x2, y2, x3, y3)
}

// Rectangle fills and strokes a rectangle.
func (a *Agg2DFloat) Rectangle(x1, y1, x2, y2 float64) { a.impl.Rectangle(x1, y1, x2, y2) }

// Ellipse fills and strokes an ellipse.
func (a *Agg2DFloat) Ellipse(cx, cy, rx, ry float64) { a.impl.Ellipse(cx, cy, rx, ry) }

// DrawCircle strokes a circle.
func (a *Agg2DFloat) DrawCircle(cx, cy, radius float64) { a.impl.DrawCircle(cx, cy, radius) }

// FillCircle fills a circle.
func (a *Agg2DFloat) FillCircle(cx, cy, radius float64) { a.impl.FillCircle(cx, cy, radius) }

// RoundedRect fills and strokes a rounded rectangle with a uniform corner radius.
func (a *Agg2DFloat) RoundedRect(x1, y1, x2, y2, r float64) {
	a.impl.RoundedRect(x1, y1, x2, y2, r)
}

// RoundedRectXY fills and strokes a rounded rectangle with separate x and y radii.
func (a *Agg2DFloat) RoundedRectXY(x1, y1, x2, y2, rx, ry float64) {
	a.impl.RoundedRectXY(x1, y1, x2, y2, rx, ry)
}

// RoundedRectVariableRadii fills and strokes a rounded rectangle with distinct top and bottom radii.
func (a *Agg2DFloat) RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop float64) {
	a.impl.RoundedRectVariableRadii(x1, y1, x2, y2, rxBottom, ryBottom, rxTop, ryTop)
}

// Arc strokes an elliptical arc described by center, radii, start, and sweep angles.
func (a *Agg2DFloat) Arc(cx, cy, rx, ry, start, sweep float64) {
	a.impl.Arc(cx, cy, rx, ry, start, sweep)
}

// ArcRel appends an elliptical arc to the path using relative coordinates.
func (a *Agg2DFloat) ArcRel(rx, ry, angle float64, largeArcFlag, sweepFlag bool, dx, dy float64) {
	a.impl.ArcRel(rx, ry, angle, largeArcFlag, sweepFlag, dx, dy)
}

// Star fills and strokes a star polygon centered at cx, cy.
func (a *Agg2DFloat) Star(cx, cy, r1, r2, startAngle float64, numRays int) {
	a.impl.Star(cx, cy, r1, r2, startAngle, numRays)
}

// Curve strokes a quadratic Bézier convenience shape.
func (a *Agg2DFloat) Curve(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Curve(x1, y1, x2, y2, x3, y3)
}

// Curve4 strokes a cubic Bézier convenience shape.
func (a *Agg2DFloat) Curve4(x1, y1, x2, y2, x3, y3, x4, y4 float64) {
	a.impl.Curve4(x1, y1, x2, y2, x3, y3, x4, y4)
}

// Polygon fills and strokes a polygon from xy as alternating x,y coordinates.
func (a *Agg2DFloat) Polygon(xy []float64, numPoints int) { a.impl.Polygon(xy, numPoints) }

// Polyline strokes an open polyline from xy as alternating x,y coordinates.
func (a *Agg2DFloat) Polyline(xy []float64, numPoints int) { a.impl.Polyline(xy, numPoints) }

// Parallelogram applies a transform mapping the unit square to a parallelogram.
func (a *Agg2DFloat) Parallelogram(x1, y1, x2, y2, x3, y3 float64) {
	a.impl.Parallelogram(x1, y1, x2, y2, x3, y3)
}

// ParallelogramFromRect mirrors the upstream `parallelogram(x1, y1, x2, y2, para)`
// overload by mapping the source rectangle `(x1, y1)-(x2, y2)` into the
// destination parallelogram encoded as `{px1, py1, px2, py2, px3, py3}`.
func (a *Agg2DFloat) ParallelogramFromRect(x1, y1, x2, y2 float64, parallelogram []float64) {
	if len(parallelogram) != 6 {
		return
	}
	a.impl.ParallelogramFromRect(
		x1, y1, x2, y2,
		parallelogram[0], parallelogram[1],
		parallelogram[2], parallelogram[3],
		parallelogram[4], parallelogram[5],
	)
}

// --- Gradients ---

// FillLinearGradient sets up a linear gradient for fill operations.
func (a *Agg2DFloat) FillLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	a.impl.FillLinearGradient(x1, y1, x2, y2, toInternalColor(c1), toInternalColor(c2), profile)
}

// LineLinearGradient sets up a linear gradient for stroke operations.
func (a *Agg2DFloat) LineLinearGradient(x1, y1, x2, y2 float64, c1, c2 Color, profile float64) {
	a.impl.LineLinearGradient(x1, y1, x2, y2, toInternalColor(c1), toInternalColor(c2), profile)
}

// FillRadialGradient sets up a two-color radial gradient for fill operations.
func (a *Agg2DFloat) FillRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	a.impl.FillRadialGradient(x, y, r, toInternalColor(c1), toInternalColor(c2), profile)
}

// LineRadialGradient sets up a two-color radial gradient for stroke operations.
func (a *Agg2DFloat) LineRadialGradient(x, y, r float64, c1, c2 Color, profile float64) {
	a.impl.LineRadialGradient(x, y, r, toInternalColor(c1), toInternalColor(c2), profile)
}

// FillRadialGradientMultiStop sets up a three-color radial gradient for fill.
func (a *Agg2DFloat) FillRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	a.impl.FillRadialGradientMultiStop(x, y, r, toInternalColor(c1), toInternalColor(c2), toInternalColor(c3))
}

// LineRadialGradientMultiStop sets up a three-color radial gradient for strokes.
func (a *Agg2DFloat) LineRadialGradientMultiStop(x, y, r float64, c1, c2, c3 Color) {
	a.impl.LineRadialGradientMultiStop(x, y, r, toInternalColor(c1), toInternalColor(c2), toInternalColor(c3))
}

// FillRadialGradientStops sets up a radial fill gradient from an arbitrary sorted
// slice of GradientStops (Position 0 = centre, Position 1 = edge). Stops must be
// in ascending Position order; positions outside [0, 1] are clamped.
func (a *Agg2DFloat) FillRadialGradientStops(x, y, r float64, stops []GradientStop) {
	internalStops := make([]agg2d.ColorStop, len(stops))
	for i, s := range stops {
		internalStops[i] = agg2d.ColorStop{
			Position: s.Position,
			Color:    toInternalColor(s.Color),
		}
	}
	a.impl.FillRadialGradientStops(x, y, r, internalStops)
}

// FillRadialGradientPos repositions the fill radial gradient without changing colors.
func (a *Agg2DFloat) FillRadialGradientPos(x, y, r float64) { a.impl.FillRadialGradientPos(x, y, r) }

// LineRadialGradientPos repositions the line radial gradient without changing colors.
func (a *Agg2DFloat) LineRadialGradientPos(x, y, r float64) { a.impl.LineRadialGradientPos(x, y, r) }

// FillGradientFlag returns the current fill gradient type.
func (a *Agg2DFloat) FillGradientFlag() int { return int(a.impl.FillGradientFlag()) }

// LineGradientFlag returns the current line gradient type.
func (a *Agg2DFloat) LineGradientFlag() int { return int(a.impl.LineGradientFlag()) }

// FillGradientD1 returns the first bound of the current fill gradient.
func (a *Agg2DFloat) FillGradientD1() float64 { return a.impl.FillGradientD1() }

// FillGradientD2 returns the second bound of the current fill gradient.
func (a *Agg2DFloat) FillGradientD2() float64 { return a.impl.FillGradientD2() }

// LineGradientD1 returns the first bound of the current line gradient.
func (a *Agg2DFloat) LineGradientD1() float64 { return a.impl.LineGradientD1() }

// LineGradientD2 returns the second bound of the current line gradient.
func (a *Agg2DFloat) LineGradientD2() float64 { return a.impl.LineGradientD2() }

// --- Image transfer ---

// CopyImage copies a float source image into the target at (dstX, dstY).
func (a *Agg2DFloat) CopyImage(img *ImageFloat, dstX, dstY int) {
	if img == nil {
		return
	}
	a.impl.CopyImageFloat(img.impl, dstX, dstY)
}

// BlendImage blends a float source image into the target at (dstX, dstY) with
// the given uniform coverage (0..255).
func (a *Agg2DFloat) BlendImage(img *ImageFloat, dstX, dstY int, cover uint8) {
	if img == nil {
		return
	}
	a.impl.BlendImageFloat(img.impl, dstX, dstY, cover)
}

// --- Image transforms (affine / perspective) ---

// TransformImage maps a source rectangle of img to a destination rectangle using
// the current image filter and world transform.
func (a *Agg2DFloat) TransformImage(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloat(img.impl, imgX1, imgY1, imgX2, imgY2, dstX1, dstY1, dstX2, dstY2)
}

// TransformImageSimple maps the whole image to a destination rectangle.
func (a *Agg2DFloat) TransformImageSimple(img *ImageFloat, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatSimple(img.impl, dstX1, dstY1, dstX2, dstY2)
}

// TransformImageParallelogram maps a source rectangle to a destination
// parallelogram (6 floats: 3 corners).
func (a *Agg2DFloat) TransformImageParallelogram(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatParallelogram(img.impl, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImageParallelogramSimple maps the whole image to a destination
// parallelogram.
func (a *Agg2DFloat) TransformImageParallelogramSimple(img *ImageFloat, parallelogram []float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatParallelogramSimple(img.impl, parallelogram)
}

// TransformImagePath maps a source rectangle along the current path, clipping the
// image to the path shape.
func (a *Agg2DFloat) TransformImagePath(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatPath(img.impl, imgX1, imgY1, imgX2, imgY2, dstX1, dstY1, dstX2, dstY2)
}

// TransformImagePathSimple maps the whole image along the current path.
func (a *Agg2DFloat) TransformImagePathSimple(img *ImageFloat, dstX1, dstY1, dstX2, dstY2 float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatPathSimple(img.impl, dstX1, dstY1, dstX2, dstY2)
}

// TransformImagePathParallelogram maps a source rectangle along the current path
// to a destination parallelogram.
func (a *Agg2DFloat) TransformImagePathParallelogram(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, parallelogram []float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatPathParallelogram(img.impl, imgX1, imgY1, imgX2, imgY2, parallelogram)
}

// TransformImagePathParallelogramSimple maps the whole image along the current
// path to a destination parallelogram.
func (a *Agg2DFloat) TransformImagePathParallelogramSimple(img *ImageFloat, parallelogram []float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatPathParallelogramSimple(img.impl, parallelogram)
}

// TransformImageQuad maps a source rectangle to an arbitrary destination
// quadrangle using perspective interpolation. quad holds [x0,y0,...,x3,y3] for
// the TL, TR, BR, BL corners.
func (a *Agg2DFloat) TransformImageQuad(img *ImageFloat, imgX1, imgY1, imgX2, imgY2 int, quad [8]float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatQuad(img.impl, imgX1, imgY1, imgX2, imgY2, quad)
}

// TransformImageQuadSimple maps the whole image to an arbitrary destination
// quadrangle using perspective interpolation.
func (a *Agg2DFloat) TransformImageQuadSimple(img *ImageFloat, quad [8]float64) error {
	if img == nil {
		return errNilImageFloat
	}
	return a.impl.TransformImageFloatQuadSimple(img.impl, quad)
}

// --- Transforms ---

// WorldToScreen transforms world coordinates to screen coordinates.
func (a *Agg2DFloat) WorldToScreen(x, y *float64) { a.impl.WorldToScreen(x, y) }

// ScreenToWorld transforms screen coordinates to world coordinates.
func (a *Agg2DFloat) ScreenToWorld(x, y *float64) { a.impl.ScreenToWorld(x, y) }

// WorldToScreenScalar converts a world-space distance using the current transform.
func (a *Agg2DFloat) WorldToScreenScalar(scalar float64) float64 {
	return a.impl.WorldToScreenScalar(scalar)
}

// ScreenToWorldScalar converts a screen-space distance using the current transform.
func (a *Agg2DFloat) ScreenToWorldScalar(scalar float64) float64 {
	return a.impl.ScreenToWorldScalar(scalar)
}

// Viewport sets up a viewport transformation mapping world coordinates onto a
// screen rectangle.
func (a *Agg2DFloat) Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2 float64, opt ViewportOption) {
	a.impl.Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, opt)
}

// ViewportDefault applies the upstream default viewport option: XMidYMid.
func (a *Agg2DFloat) ViewportDefault(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2 float64) {
	a.Viewport(worldX1, worldY1, worldX2, worldY2, screenX1, screenY1, screenX2, screenY2, XMidYMid)
}

// WorldToScreenDistance transforms a distance from world to screen units.
func (a *Agg2DFloat) WorldToScreenDistance(worldDistance float64) float64 {
	return a.impl.WorldToScreenDistance(worldDistance)
}

// ScreenToWorldDistance transforms a distance from screen to world units; the
// bool is false when the transform is not invertible.
func (a *Agg2DFloat) ScreenToWorldDistance(screenDistance float64) (float64, bool) {
	return a.impl.ScreenToWorldDistance(screenDistance)
}

// AlignPoint snaps a world point to pixel-centre boundaries for crisp rendering.
func (a *Agg2DFloat) AlignPoint(x, y *float64) { a.impl.AlignPoint(x, y) }

// InBox reports whether a world point lies inside the current clip box.
func (a *Agg2DFloat) InBox(worldX, worldY float64) bool { return a.impl.InBox(worldX, worldY) }

// AffineImageResamplePolicy controls how affine image transforms choose between
// direct filtered spans and the affine resampler.
func (a *Agg2DFloat) AffineImageResamplePolicy(policy AffineImageResamplePolicy) {
	a.impl.AffineImageResamplePolicy(policy)
}

// GetAffineImageResamplePolicy returns the current affine image resample policy.
func (a *Agg2DFloat) GetAffineImageResamplePolicy() AffineImageResamplePolicy {
	return a.impl.GetAffineImageResamplePolicy()
}

// ResetTransformations resets the world transform to identity.
func (a *Agg2DFloat) ResetTransformations() { a.impl.ResetTransformations() }

// Rotate applies a rotation (radians) to the world transform.
func (a *Agg2DFloat) Rotate(angle float64) { a.impl.Rotate(angle) }

// Scale applies a scaling transformation.
func (a *Agg2DFloat) Scale(sx, sy float64) { a.impl.Scale(sx, sy) }

// UniformScale applies uniform scaling on both axes.
func (a *Agg2DFloat) UniformScale(s float64) { a.impl.UniformScale(s) }

// Skew applies a skewing transformation (radians).
func (a *Agg2DFloat) Skew(sx, sy float64) { a.impl.Skew(sx, sy) }

// Translate applies a translation.
func (a *Agg2DFloat) Translate(x, y float64) { a.impl.Translate(x, y) }

// --- Fill mode ---

// NoFill disables fill (transparent fill color).
func (a *Agg2DFloat) NoFill() { a.impl.NoFill() }

// NoLine disables stroke (transparent line color).
func (a *Agg2DFloat) NoLine() { a.impl.NoLine() }

// FillEvenOdd selects the even-odd (true) or non-zero (false) fill rule.
func (a *Agg2DFloat) FillEvenOdd(evenOddFlag bool) { a.impl.FillEvenOdd(evenOddFlag) }

// GetFillEvenOdd reports whether the even-odd fill rule is active.
func (a *Agg2DFloat) GetFillEvenOdd() bool { return a.impl.GetFillEvenOdd() }
