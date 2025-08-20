// Package agg provides a Go port of the Anti-Grain Geometry (AGG) rendering library.
//
// AGG is a high-quality 2D graphics library that provides anti-aliased rendering
// with subpixel accuracy. This Go port maintains the core functionality while
// providing a Go-idiomatic API.
//
// Basic usage:
//
//	agg2d := agg.NewAgg2D()
//	agg2d.Attach(buffer, width, height, stride)
//	agg2d.FillColor(agg.Color{255, 0, 0, 255}) // Red
//	agg2d.Ellipse(400, 300, 100, 100)
//	agg2d.DrawPath(agg.FillAndStroke)
package agg

import (
	"math"

	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/path"
	"agg_go/internal/rasterizer"
	"agg_go/internal/scanline"
	"agg_go/internal/transform"
)

// Forward declaration for transform stack (actual definition in agg2d_transform.go)

// Color represents an SRGBA8 color (0-255 range) matching AGG's srgba8.
// This is the primary color type used throughout the AGG2D interface.
type Color struct {
	R, G, B, A uint8
}

// NewColor creates a new Color with the specified RGBA components.
func NewColor(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// NewColorRGB creates a new opaque Color with alpha = 255.
func NewColorRGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: 255}
}

// ConvertToRGBA converts the color to floating-point RGBA (0.0-1.0 range).
func (c Color) ConvertToRGBA() color.RGBA {
	const scale = 1.0 / 255.0
	return color.RGBA{
		R: float64(c.R) * scale,
		G: float64(c.G) * scale,
		B: float64(c.B) * scale,
		A: float64(c.A) * scale,
	}
}

// Gradient performs linear interpolation between this color and another color.
// Parameter k should be in the range [0.0, 1.0] where:
// - k = 0.0 returns this color (c)
// - k = 1.0 returns the target color (c2)
// - k = 0.5 returns the midpoint between both colors
// This matches the C++ AGG srgba8::gradient method.
func (c Color) Gradient(c2 Color, k float64) Color {
	if k <= 0.0 {
		return c
	}
	if k >= 1.0 {
		return c2
	}

	// Linear interpolation: c + (c2 - c) * k
	return Color{
		R: uint8(float64(c.R) + (float64(c2.R)-float64(c.R))*k),
		G: uint8(float64(c.G) + (float64(c2.G)-float64(c.G))*k),
		B: uint8(float64(c.B) + (float64(c2.B)-float64(c.B))*k),
		A: uint8(float64(c.A) + (float64(c2.A)-float64(c.A))*k),
	}
}

// Predefined colors for convenience
var (
	Black       = NewColorRGB(0, 0, 0)
	White       = NewColorRGB(255, 255, 255)
	Red         = NewColorRGB(255, 0, 0)
	Green       = NewColorRGB(0, 255, 0)
	Blue        = NewColorRGB(0, 0, 255)
	Yellow      = NewColorRGB(255, 255, 0)
	Cyan        = NewColorRGB(0, 255, 255)
	Magenta     = NewColorRGB(255, 0, 255)
	Transparent = NewColor(0, 0, 0, 0)
)

// RGB creates a Color from floating-point RGB values (0.0-1.0 range).
// This is a convenience function for creating colors from normalized values.
func RGB(r, g, b float64) Color {
	return Color{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// RGBA creates a Color from floating-point RGBA values (0.0-1.0 range).
// This is a convenience function for creating colors from normalized values.
func RGBA(r, g, b, a float64) Color {
	return Color{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: uint8(a * 255),
	}
}

// RGBA8 type for compatibility with examples that might use it
type RGBA8 struct {
	R, G, B, A uint8
}

// NewRGBA8 creates a new RGBA8 color.
func NewRGBA8(r, g, b, a uint8) RGBA8 {
	return RGBA8{R: r, G: g, B: b, A: a}
}

// RGBAColor represents a floating-point RGBA color.
type RGBAColor struct {
	R, G, B, A float64
}

// NewRGBAColor creates a new RGBAColor with floating-point values.
func NewRGBAColor(r, g, b, a float64) RGBAColor {
	return RGBAColor{R: r, G: g, B: b, A: a}
}

// LineJoin defines the style of line joins.
type LineJoin int

const (
	JoinMiter LineJoin = iota // Miter join - sharp corner
	JoinRound                 // Round join - rounded corner
	JoinBevel                 // Bevel join - flat corner
)

// LineCap defines the style of line caps.
type LineCap int

const (
	CapButt   LineCap = iota // Butt cap - flat end
	CapSquare                // Square cap - flat end extending beyond endpoint
	CapRound                 // Round cap - rounded end
)

// TextAlignment defines text alignment options.
type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignRight
	AlignCenter
	AlignTop    = AlignRight // For vertical alignment
	AlignBottom = AlignLeft  // For vertical alignment
)

// DrawPathFlag defines how paths should be rendered.
// Note: These constants are defined in agg2d.go to avoid duplication.

// ViewportOption defines viewport scaling behavior.
type ViewportOption int

const (
	Anisotropic ViewportOption = iota // Stretch to fit, ignoring aspect ratio
	XMinYMin                          // Align to top-left, preserve aspect ratio
	XMidYMin                          // Align to top-center, preserve aspect ratio
	XMaxYMin                          // Align to top-right, preserve aspect ratio
	XMinYMid                          // Align to middle-left, preserve aspect ratio
	XMidYMid                          // Align to center, preserve aspect ratio
	XMaxYMid                          // Align to middle-right, preserve aspect ratio
	XMinYMax                          // Align to bottom-left, preserve aspect ratio
	XMidYMax                          // Align to bottom-center, preserve aspect ratio
	XMaxYMax                          // Align to bottom-right, preserve aspect ratio
)

// BlendMode defines pixel blending operations.
// Note: These constants are defined in agg2d.go to avoid duplication.

// ImageFilter defines image filtering modes.
type ImageFilter int

const (
	NoFilter    ImageFilter = iota // No filtering
	Bilinear                       // Bilinear interpolation
	Hanning                        // Hanning filter
	Hermite                        // Hermite filter
	Quadric                        // Quadric filter
	Bicubic                        // Bicubic filter
	Catrom                         // Catmull-Rom filter
	Spline16                       // Spline16 filter
	Spline36                       // Spline36 filter
	Blackman144                    // Blackman 144 filter
)

// ImageResample defines image resampling modes.
type ImageResample int

const (
	NoResample        ImageResample = iota // No resampling
	ResampleAlways                         // Always resample
	ResampleOnZoomOut                      // Resample only when zooming out
)

// FontCacheType defines font caching modes.
type FontCacheType int

const (
	RasterFontCache FontCacheType = iota // Raster font cache
	VectorFontCache                      // Vector font cache
)

// Direction defines path winding direction.
// Note: These constants are defined in agg2d.go to avoid duplication.

// Gradient defines gradient types.
type Gradient int

const (
	Solid  Gradient = iota // Solid color
	Linear                 // Linear gradient
	Radial                 // Radial gradient
)

// Rect represents an integer rectangle.
type Rect struct {
	X1, Y1, X2, Y2 int
}

// RectD represents a floating-point rectangle.
// Note: This type is defined in agg2d.go to avoid duplication.

// Affine represents a 2D affine transformation matrix.
// This matches AGG's trans_affine structure.
type Affine struct {
	matrix [6]float64 // [sx, shy, shx, sy, tx, ty]
}

// NewAffine creates an identity affine transformation.
func NewAffine() *Affine {
	a := &Affine{}
	a.Reset()
	return a
}

// Reset sets the matrix to identity.
func (a *Affine) Reset() {
	a.matrix[0] = 1.0 // sx
	a.matrix[1] = 0.0 // shy
	a.matrix[2] = 0.0 // shx
	a.matrix[3] = 1.0 // sy
	a.matrix[4] = 0.0 // tx
	a.matrix[5] = 0.0 // ty
}

// Transformations represents the transformation state (actual definition in agg2d_transform.go).

// Image represents a raster image that can be used as a rendering target.
// This matches the C++ Agg2D::Image structure.
type Image struct {
	renBuf *buffer.RenderingBuffer[uint8]
	Data   []uint8 // Raw pixel data (RGBA format)
	width  int     // Width in pixels
	height int     // Height in pixels
}

// NewImage creates a new image with the specified buffer.
func NewImage(buf []uint8, width, height, stride int) *Image {
	img := &Image{
		renBuf: buffer.NewRenderingBuffer[uint8](),
		Data:   buf,
		width:  width,
		height: height,
	}
	img.renBuf.Attach(buf, width, height, stride)
	return img
}

// Width returns the image width.
func (img *Image) Width() int {
	return img.width
}

// Height returns the image height.
func (img *Image) Height() int {
	return img.height
}

// Attach attaches a buffer to the image.
func (img *Image) Attach(buf []uint8, width, height, stride int) {
	img.renBuf.Attach(buf, width, height, stride)
	img.Data = buf
	img.width = width
	img.height = height
}

// Agg2D represents the main rendering context.
// This is the primary interface that matches the C++ Agg2D class.
type Agg2D struct {
	// Rendering buffer
	rbuf *buffer.RenderingBuffer[uint8]

	// Clip box
	clipBox struct{ X1, Y1, X2, Y2 float64 } // RectD equivalent

	// Blend modes
	blendMode       int // BlendMode defined in agg2d.go
	imageBlendMode  int // BlendMode defined in agg2d.go
	imageBlendColor Color

	// Scanline and rasterizer
	scanline   *scanline.ScanlineU8
	rasterizer *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]

	// Rendering components (using interfaces for compatibility)
	pixfmt   interface{} // RGBA pixel format - will be *pixfmt.PixFmtRGBA32
	renBase  interface{} // Base renderer
	renSolid interface{} // Solid color renderer

	// Master alpha and anti-aliasing gamma
	masterAlpha    float64
	antiAliasGamma float64

	// Fill and line colors
	fillColor Color
	lineColor Color

	// Gradients
	fillGradient       [256]Color
	lineGradient       [256]Color
	fillGradientFlag   Gradient
	lineGradientFlag   Gradient
	fillGradientMatrix *transform.TransAffine
	lineGradientMatrix *transform.TransAffine
	fillGradientD1     float64
	lineGradientD1     float64
	fillGradientD2     float64
	lineGradientD2     float64

	// Line attributes
	lineCap   LineCap
	lineJoin  LineJoin
	lineWidth float64

	// Text attributes
	textAngle        float64
	textAlignX       TextAlignment
	textAlignY       TextAlignment
	textHints        bool
	fontHeight       float64
	fontAscent       float64
	fontDescent      float64
	fontCacheType    FontCacheType
	fontEngine       interface{} // FontEngine interface - actual type depends on build tags
	fontCacheManager interface{} // FontCacheManager - manages glyph caching

	// Image filtering
	imageFilter   ImageFilter
	imageResample ImageResample

	// Fill mode
	evenOddFlag bool

	// Path and transformation
	path           *path.PathStorageStl
	transform      *transform.TransAffine
	transformStack *TransformStack // Optional transform stack for push/pop operations

	// Converters
	convCurve  *conv.ConvCurve
	convDash   *conv.ConvDash // Optional dash converter (nil when not using dashes)
	convStroke *conv.ConvStroke
}

// Constants
const (
	// Pi constant for convenience
	Pi = math.Pi
)

// Deg2Rad and Rad2Deg functions are defined in agg2d_transform.go
