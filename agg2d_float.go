package agg

import (
	"image"

	"github.com/cwbudde/agg_go/internal/agg2d"
	"github.com/cwbudde/agg_go/internal/color"
)

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

// VerLineTo draws a vertical line to the given y.
func (a *Agg2DFloat) VerLineTo(y float64) { a.impl.VerLineTo(y) }

// ArcTo adds an SVG-style elliptical arc to the path.
func (a *Agg2DFloat) ArcTo(rx, ry, angle float64, largeArcFlag, sweepFlag bool, x, y float64) {
	a.impl.ArcTo(rx, ry, angle, largeArcFlag, sweepFlag, x, y)
}

// QuadricCurveTo adds a quadratic Bézier curve to the path.
func (a *Agg2DFloat) QuadricCurveTo(xCtrl, yCtrl, xTo, yTo float64) {
	a.impl.QuadricCurveTo(xCtrl, yCtrl, xTo, yTo)
}

// CubicCurveTo adds a cubic Bézier curve to the path.
func (a *Agg2DFloat) CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo float64) {
	a.impl.CubicCurveTo(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo)
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

// --- Transforms ---

// WorldToScreen transforms world coordinates to screen coordinates.
func (a *Agg2DFloat) WorldToScreen(x, y *float64) { a.impl.WorldToScreen(x, y) }

// ScreenToWorld transforms screen coordinates to world coordinates.
func (a *Agg2DFloat) ScreenToWorld(x, y *float64) { a.impl.ScreenToWorld(x, y) }

// WorldToScreenScalar converts a world-space distance using the current transform.
func (a *Agg2DFloat) WorldToScreenScalar(scalar float64) float64 {
	return a.impl.WorldToScreenScalar(scalar)
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
