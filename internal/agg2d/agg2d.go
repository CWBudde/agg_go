// Package agg2d provides the internal AGG2D high-level interface implementation.
// This is a Go port of the C++ Agg2D class from AGG 2.6.
package agg2d

import (
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
	"agg_go/internal/scanline"
	"agg_go/internal/span"
	"agg_go/internal/transform"
)

// Color represents a RGBA color with 8-bit components.
type Color [4]uint8

// Type aliases for the different enums and constants
type (
	BlendMode      = int
	Gradient       = int
	LineCap        = int
	LineJoin       = int
	TextAlignment  = int
	FontCacheType  = int
	ImageFilter    = int
	ImageResample  = int
	ViewportOption = int
)

// Core constants
const (
	// Gradients
	Solid  Gradient = 0
	Linear Gradient = 1
	Radial Gradient = 2

	// Line caps
	CapButt  LineCap = 0
	CapRound LineCap = 1

	// Line joins
	JoinMiter LineJoin = 0
	JoinRound LineJoin = 1

	// Text alignment
	AlignLeft   TextAlignment = 0
	AlignBottom TextAlignment = 0

	// Font cache
	RasterFontCache FontCacheType = 0

	// Image filter
	ImageFilterBilinear ImageFilter = 0

	// Image resample
	NoResample ImageResample = 0

	// ViewportOption constants
	Anisotropic ViewportOption = iota
	XMinYMin
	XMinYMid
	XMinYMax
	XMidYMin
	XMidYMid
	XMidYMax
	XMaxYMin
	XMaxYMid
	XMaxYMax
)

// Agg2D is the main high-level rendering interface.
// This matches the C++ Agg2D class from the original AGG library.
type Agg2D struct {
	// Rendering buffer
	rbuf *buffer.RenderingBuffer[uint8]

	// Clip box
	clipBox struct{ X1, Y1, X2, Y2 float64 } // RectD equivalent

	// Blend modes
	blendMode       BlendMode
	imageBlendMode  BlendMode
	imageBlendColor Color

	// Scanline and rasterizer
	scanline   *scanline.ScanlineU8
	rasterizer *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]

	// Rendering components (now properly typed)
	pixfmt  *pixfmt.PixFmtRGBA32
	renBase *baseRendererAdapter[color.RGBA8[color.Linear]]

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

	// Span rendering components for gradients and patterns
	spanAllocator *span.SpanAllocator[color.RGBA8[color.Linear]]

	// Control point tracking for smooth curves
	lastCtrlX, lastCtrlY float64
	hasLastCtrl          bool
}

// TransformStack is defined in transform.go

var (
	Black = Color{0, 0, 0, 255}
	White = Color{255, 255, 255, 255}
)

func NewColor(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// TransformStack manages a stack of transformation matrices
type TransformStack struct {
	stack []*transform.TransAffine
}

// Gradient creates a linear interpolation between two colors
func (c Color) Gradient(to Color, factor float64) Color {
	// Clamp factor to [0, 1]
	if factor < 0.0 {
		factor = 0.0
	}
	if factor > 1.0 {
		factor = 1.0
	}

	// Interpolate each component
	r := uint8(float64(c[0]) + factor*float64(int(to[0])-int(c[0])))
	g := uint8(float64(c[1]) + factor*float64(int(to[1])-int(c[1])))
	b := uint8(float64(c[2]) + factor*float64(int(to[2])-int(c[2])))
	a := uint8(float64(c[3]) + factor*float64(int(to[3])-int(c[3])))

	return Color{r, g, b, a}
}

// NewAgg2D creates a new AGG2D rendering context.
// This matches the C++ Agg2D constructor.
func NewAgg2D() *Agg2D {
	agg2d := &Agg2D{
		rbuf:               buffer.NewRenderingBuffer[uint8](),
		clipBox:            struct{ X1, Y1, X2, Y2 float64 }{0, 0, 0, 0},
		blendMode:          BlendAlpha,
		imageBlendMode:     BlendDst,
		imageBlendColor:    NewColor(0, 0, 0, 255),
		masterAlpha:        1.0,
		antiAliasGamma:     1.0,
		fillColor:          White,
		lineColor:          Black,
		fillGradientFlag:   Solid,
		lineGradientFlag:   Solid,
		fillGradientD1:     0.0,
		lineGradientD1:     0.0,
		fillGradientD2:     100.0,
		lineGradientD2:     100.0,
		textAngle:          0.0,
		textAlignX:         AlignLeft,
		textAlignY:         AlignBottom,
		textHints:          true,
		fontHeight:         0.0,
		fontAscent:         0.0,
		fontDescent:        0.0,
		fontCacheType:      RasterFontCache,
		imageFilter:        ImageFilterBilinear,
		imageResample:      NoResample,
		lineWidth:          1.0,
		lineCap:            CapRound,
		lineJoin:           JoinRound,
		evenOddFlag:        false,
		path:               path.NewPathStorageStl(),
		transform:          transform.NewTransAffine(),
		fillGradientMatrix: transform.NewTransAffine(),
		lineGradientMatrix: transform.NewTransAffine(),
		scanline:           scanline.NewScanlineU8(),
	}

	// Initialize converters
	pathAdapter := &pathVertexSourceAdapter{path: agg2d.path}
	agg2d.convCurve = conv.NewConvCurve(pathAdapter)
	agg2d.convStroke = conv.NewConvStroke(agg2d.convCurve)

	// Initialize rasterizer with default cell block limit and clipper
	clipper := &rasterizer.RasterizerSlNoClip{}
	agg2d.rasterizer = rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl](8192, clipper)

	// Initialize span allocator for gradient rendering
	agg2d.spanAllocator = span.NewSpanAllocator[color.RGBA8[color.Linear]]()

	// Set default line cap and join
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetLineCap(1)  // Default CapRound
		agg2d.convStroke.SetLineJoin(1) // Default JoinRound
	}

	return agg2d
}

// FillColor sets the fill color.
func (agg2d *Agg2D) FillColor(c Color) {
	agg2d.fillColor = c
	agg2d.fillGradientFlag = Solid
}

// LineColor sets the line color.
func (agg2d *Agg2D) LineColor(c Color) {
	agg2d.lineColor = c
	agg2d.lineGradientFlag = Solid
}
