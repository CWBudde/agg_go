package agg2d

import (
	"github.com/MeKo-Christian/agg_go/internal/basics"
	"github.com/MeKo-Christian/agg_go/internal/buffer"
	"github.com/MeKo-Christian/agg_go/internal/color"
	"github.com/MeKo-Christian/agg_go/internal/conv"
	"github.com/MeKo-Christian/agg_go/internal/font"
	"github.com/MeKo-Christian/agg_go/internal/font/freetype"
	"github.com/MeKo-Christian/agg_go/internal/gsv"
	aggimage "github.com/MeKo-Christian/agg_go/internal/image"
	"github.com/MeKo-Christian/agg_go/internal/path"
	"github.com/MeKo-Christian/agg_go/internal/pixfmt"
	"github.com/MeKo-Christian/agg_go/internal/rasterizer"
	"github.com/MeKo-Christian/agg_go/internal/renderer"
	renscan "github.com/MeKo-Christian/agg_go/internal/renderer/scanline"
	"github.com/MeKo-Christian/agg_go/internal/scanline"
	"github.com/MeKo-Christian/agg_go/internal/span"
	"github.com/MeKo-Christian/agg_go/internal/transform"
)

// Color is the user-facing 8-bit RGBA color carried through the internal Agg2D
// state machine, matching the C++ Agg2D::Color role.
type Color [4]uint8

// Type aliases for the Agg2D-style enums exposed across the package.
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

// Core constants mirror the enum values exposed by the C++ Agg2D interface.
const (
	// Gradients
	Solid  Gradient = 0
	Linear Gradient = 1
	Radial Gradient = 2

	// Line caps
	CapButt   LineCap = 0
	CapSquare LineCap = 1
	CapRound  LineCap = 2

	// Line joins
	JoinMiter LineJoin = 0
	JoinRound LineJoin = 2
	JoinBevel LineJoin = 3

	// Text alignment
	AlignLeft   TextAlignment = 0
	AlignBottom TextAlignment = 0

	// Font cache
	RasterFontCache FontCacheType = 0

	// Image filter
	ImageFilterBilinear ImageFilter = 1

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

// Agg2D is the internal stateful rendering facade that coordinates the AGG
// path, transform, rasterizer, span, image, and text subsystems.
//
// Like the original C++ Agg2D class, it is configured through mutable drawing
// state. Operations update that state, populate the current path or image
// parameters, and then render through the scanline pipeline into the currently
// attached buffer.
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
	rasterizer *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

	// Cached adapters that bridge the internal scanline/rasterizer types to
	// the renderer/scanline interfaces. Stored as fields (like C++ stores
	// m_renSolid, m_scanline, m_rasterizer by value) to avoid per-call
	// heap allocations in the hot rendering path.
	rasAdapter rasterizerAdapter
	slAdapter  scanlineWrapper

	// Rendering components (now properly typed)
	pixfmt         *pixfmt.PixFmtRGBA32[color.Linear]
	pixfmtPre      *pixfmt.PixFmtRGBA32Pre[color.Linear]
	pixfmtComp     *pixfmt.PixFmtCompositeRGBA32
	pixfmtCompPre  *pixfmt.PixFmtCompositeRGBA32Pre
	renBase        *baseRendererAdapter[color.RGBA8[color.Linear]]
	renBasePre     *baseRendererAdapter[color.RGBA8[color.Linear]]
	renBaseComp    *baseRendererAdapter[color.RGBA8[color.Linear]]
	renBaseCompPre *baseRendererAdapter[color.RGBA8[color.Linear]]

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
	textAngle     float64
	textAlignX    TextAlignment
	textAlignY    TextAlignment
	textHints     bool
	flipText      bool
	fontHeight    float64
	fontAscent    float64
	fontDescent   float64
	fontCacheType FontCacheType

	// AGG's agg2d.h wires Agg2D through font_cache_manager<FontEngine>.
	// Keep that stack authoritative here; the fman/font_cache_manager2 path
	// remains separate for lower-level FreeType2 experiments and examples.
	//
	// Both build-tag variants expose the same concrete engine type, so Agg2D
	// does not need runtime type assertions on the text path.
	fontEngine       *freetype.FontEngineFreetype
	fontCacheManager *font.FontCacheManager

	// TODO(Path B): Temporary GSV stroke-font fallback — replace with a proper
	// pure-Go TTF engine (Path A) once one is available.
	// GSV is used in WASM builds where cgo/FreeType is unavailable.
	gsvText     *gsv.GSVText // Non-nil when FontGSV() has been called
	gsvFontMode bool         // True when the active font backend is GSV

	// Image filtering
	imageFilter    ImageFilter
	imageResample  ImageResample
	imageFilterLUT *aggimage.ImageFilterLUT

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
	spanAllocator   *span.SpanAllocator[color.RGBA8[color.Linear]]
	fillGradientLUT []color.RGBA8[color.Linear]
	lineGradientLUT []color.RGBA8[color.Linear]

	fillLinearSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	lineLinearSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	fillRadialSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	lineRadialSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]

	fillLinearSpanGenerator *span.SpanGradient[
		color.RGBA8[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientLinearX,
		*span.GradientPrebuiltColorRGBA8[color.Linear],
	]
	lineLinearSpanGenerator *span.SpanGradient[
		color.RGBA8[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientLinearX,
		*span.GradientPrebuiltColorRGBA8[color.Linear],
	]
	fillRadialSpanGenerator *span.SpanGradient[
		color.RGBA8[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientRadial,
		*span.GradientPrebuiltColorRGBA8[color.Linear],
	]
	lineRadialSpanGenerator *span.SpanGradient[
		color.RGBA8[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientRadial,
		*span.GradientPrebuiltColorRGBA8[color.Linear],
	]

	fillGradientLUTDirty bool
	lineGradientLUTDirty bool

	// Control point tracking for smooth curves
	lastCtrlX, lastCtrlY float64
	hasLastCtrl          bool
}

// TransformStack is defined in transform.go

var (
	Black = Color{0, 0, 0, 255}
	White = Color{255, 255, 255, 255}
)

// NewColor constructs an explicit RGBA color value.
func NewColor(r, g, b, a uint8) Color {
	return Color{r, g, b, a}
}

// TransformStack stores saved world transforms for PushTransform/PopTransform-
// style workflows.
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
		imageFilterLUT:     aggimage.NewImageFilterLUTWithFilter(aggimage.BilinearFilter{}, true),
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
	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(agg2d.path)
	agg2d.convCurve = conv.NewConvCurve(pathAdapter)
	agg2d.convStroke = conv.NewConvStroke(agg2d.convCurve)

	// Initialize rasterizer with default cell block limit and clipper
	clipper := rasterizer.NewRasterizerSlNoClip()
	conv := rasterizer.RasConvInt{}
	agg2d.rasterizer = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](conv, clipper)

	// Initialize cached adapters
	agg2d.rasAdapter.ras = agg2d.rasterizer
	agg2d.slAdapter.sl = agg2d.scanline

	// Initialize span allocator for gradient rendering
	agg2d.spanAllocator = span.NewSpanAllocator[color.RGBA8[color.Linear]]()
	agg2d.fillGradientLUT = make([]color.RGBA8[color.Linear], 256)
	agg2d.lineGradientLUT = make([]color.RGBA8[color.Linear], 256)

	agg2d.fillLinearSpanInterpolator = span.NewSpanInterpolatorLinearDefault(agg2d.fillGradientMatrix)
	agg2d.lineLinearSpanInterpolator = span.NewSpanInterpolatorLinearDefault(agg2d.lineGradientMatrix)
	agg2d.fillRadialSpanInterpolator = span.NewSpanInterpolatorLinearDefault(agg2d.fillGradientMatrix)
	agg2d.lineRadialSpanInterpolator = span.NewSpanInterpolatorLinearDefault(agg2d.lineGradientMatrix)

	agg2d.fillLinearSpanGenerator = span.NewLinearGradientFromLUT(
		agg2d.fillLinearSpanInterpolator,
		agg2d.fillGradientLUT,
		agg2d.fillGradientD1,
		agg2d.fillGradientD2,
	)
	agg2d.lineLinearSpanGenerator = span.NewLinearGradientFromLUT(
		agg2d.lineLinearSpanInterpolator,
		agg2d.lineGradientLUT,
		agg2d.lineGradientD1,
		agg2d.lineGradientD2,
	)
	agg2d.fillRadialSpanGenerator = span.NewRadialGradientFromLUT(
		agg2d.fillRadialSpanInterpolator,
		agg2d.fillGradientLUT,
		agg2d.fillGradientD1,
		agg2d.fillGradientD2,
	)
	agg2d.lineRadialSpanGenerator = span.NewRadialGradientFromLUT(
		agg2d.lineRadialSpanInterpolator,
		agg2d.lineGradientLUT,
		agg2d.lineGradientD1,
		agg2d.lineGradientD2,
	)

	agg2d.fillGradientLUTDirty = true
	agg2d.lineGradientLUTDirty = true

	// Set default line cap and join
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetLineCap(basics.LineCap(CapRound))
		agg2d.convStroke.SetLineJoin(basics.LineJoin(JoinRound))
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

// GetInternalRasterizer returns the underlying rasterizer.
func (agg2d *Agg2D) GetInternalRasterizer() *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip] {
	return agg2d.rasterizer
}

// ScanlineRender renders the given rasterizer data using a custom renderer.
func (agg2d *Agg2D) ScanlineRender(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], renderer renscan.RendererInterface[color.RGBA8[color.Linear]]) {
	ra := rasterizerAdapter{ras: ras}
	sl := &agg2d.slAdapter

	if !ra.RewindScanlines() {
		return
	}

	sl.Reset(ra.MinX(), ra.MaxX())
	renderer.Prepare()

	for ra.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

// GouraudTriangle renders a Gouraud-shaded triangle.
func (agg2d *Agg2D) GouraudTriangle(x1, y1, x2, y2, x3, y3 float64, c1, c2, c3 Color, d float64) {
	agg2d.rasterizer.Reset()

	// Convert colors to SpanGouraudRGBA expected format (0-255 range, matching C++ rgba8)
	gc1 := span.RGBAColor{R: int(c1[0]), G: int(c1[1]), B: int(c1[2]), A: int(c1[3])}
	gc2 := span.RGBAColor{R: int(c2[0]), G: int(c2[1]), B: int(c2[2]), A: int(c2[3])}
	gc3 := span.RGBAColor{R: int(c3[0]), G: int(c3[1]), B: int(c3[2]), A: int(c3[3])}

	spanGen := span.NewSpanGouraudRGBAWithTriangle(gc1, gc2, gc3, x1, y1, x2, y2, x3, y3, d)

	// Use a custom renderer that doesn't rely on the broken interfaces
	renderer := &gouraudRenderer{
		ren:   agg2d.renBase.rendererBase(),
		span:  spanGen,
		alloc: span.NewSpanAllocator[span.RGBAColor](),
	}

	// We need an adapter here too because of circular dependency or internal types
	// But Agg2D can see spanGen and its Rewind(uint).
	// Let's use a local adapter or just fix the interface if possible.
	// For now, let's use a simple anonymous adapter.
	adapter := &gouraudRasAdapter{sg: spanGen}
	agg2d.rasterizer.AddPath(adapter, 0)
	agg2d.scanlineRender(renderer)
}

type gouraudRenderer struct {
	ren   *renderer.RendererBase[renderer.PixelFormat[color.RGBA8[color.Linear]], color.RGBA8[color.Linear]]
	span  *span.SpanGouraudRGBA
	alloc *span.SpanAllocator[span.RGBAColor]
}

func (r *gouraudRenderer) Prepare() {
	r.span.Prepare()
}

func (r *gouraudRenderer) SetColor(c color.RGBA8[color.Linear]) {}

func (r *gouraudRenderer) Render(sl renscan.ScanlineInterface) {
	y := sl.Y()
	it := sl.Begin()
	for {
		spanData := it.GetSpan()
		x := spanData.X
		length := spanData.Len

		colors := r.alloc.Allocate(length)
		r.span.Generate(colors, x, y, uint(length))

		// Convert back to base renderer colors and blend
		baseColors := make([]color.RGBA8[color.Linear], length)
		for i := 0; i < length; i++ {
			baseColors[i] = color.RGBA8[color.Linear]{
				R: uint8(colors[i].R),
				G: uint8(colors[i].G),
				B: uint8(colors[i].B),
				A: uint8(colors[i].A),
			}
		}

		r.ren.BlendColorHspan(x, y, length, baseColors, spanData.Covers, 255)

		if !it.Next() {
			break
		}
	}
}

type gouraudRasAdapter struct {
	sg interface {
		Rewind(uint)
		Vertex() (float64, float64, basics.PathCommand)
	}
}

func (a *gouraudRasAdapter) Rewind(pathID uint32) {
	a.sg.Rewind(uint(pathID))
}

func (a *gouraudRasAdapter) Vertex(x, y *float64) uint32 {
	vx, vy, cmd := a.sg.Vertex()
	*x = vx
	*y = vy
	return uint32(cmd)
}

// SetImageFilterLUT sets the image filter lookup table.
func (agg2d *Agg2D) SetImageFilterLUT(lut *aggimage.ImageFilterLUT) {
	agg2d.imageFilterLUT = lut
}
