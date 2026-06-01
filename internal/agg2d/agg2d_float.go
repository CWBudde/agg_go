package agg2d

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/buffer"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	"github.com/cwbudde/agg_go/internal/font"
	"github.com/cwbudde/agg_go/internal/font/freetype"
	"github.com/cwbudde/agg_go/internal/gsv"
	aggimage "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/path"
	"github.com/cwbudde/agg_go/internal/pixfmt"
	"github.com/cwbudde/agg_go/internal/pixfmt/blender"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	"github.com/cwbudde/agg_go/internal/scanline"
	"github.com/cwbudde/agg_go/internal/span"
	"github.com/cwbudde/agg_go/internal/transform"
)

// Agg2DFloat is the float (128-bit, 4 x float32) twin of Agg2D. It mirrors the
// 8-bit internal facade field-for-field, swapping the pixfmt, base renderer,
// gradient LUT, span allocator, and gradient span generator types to their
// float (RGBA32 / RGBA128) counterparts.
//
// Following AGG's agg2d.h with AGG2D_USE_FLOAT_FORMAT defined: the public Color
// stays 8-bit (srgba8), while ColorType — used by the pixfmt, blenders, span
// allocator, gradient array, and span gradients — becomes the float rgba32.
// Color-agnostic helpers (transform, path, converters, rasterizer, scanline,
// image filter LUT) are reused as-is.
//
// Composite blend modes use the float composite pixfmts (PixFmtCompositeRGBA128
// / ...Pre) wrapped in dedicated base renderers, mirroring the 8-bit Comp /
// CompPre pipeline; currentRenderer / currentImageRenderer switch to them
// whenever blendMode != BlendAlpha.
type Agg2DFloat struct {
	// Rendering buffer (float)
	rbuf *buffer.RenderingBufferF32

	// Clip box
	clipBox struct{ X1, Y1, X2, Y2 float64 }

	// Blend modes
	blendMode       BlendMode
	imageBlendMode  BlendMode
	imageBlendColor Color

	// Scanline and rasterizer (color-agnostic, reused as-is)
	scanline   *scanline.ScanlineU8
	rasterizer *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]

	// Rendering components (float)
	pixfmt     *pixfmt.PixFmtRGBA128Plain
	pixfmtPre  *pixfmt.PixFmtRGBA128Pre
	renBase    *baseRendererAdapter[color.RGBA32[color.Linear]]
	renBasePre *baseRendererAdapter[color.RGBA32[color.Linear]]

	// Composite (Porter-Duff / SVG blend mode) rendering components (float).
	// renBaseComp drives solid/gradient fills and strokes; renBaseCompPre drives
	// premultiplied image transfers, mirroring the 8-bit Comp / CompPre split.
	pixfmtComp     *pixfmt.PixFmtCompositeRGBA128Linear
	pixfmtCompPre  *pixfmt.PixFmtCompositeRGBA128PreLinear
	renBaseComp    *baseRendererAdapter[color.RGBA32[color.Linear]]
	renBaseCompPre *baseRendererAdapter[color.RGBA32[color.Linear]]

	// Master alpha and anti-aliasing gamma
	masterAlpha    float64
	antiAliasGamma float64

	// Fill and line colors (public Color stays 8-bit, per C++ srgba8)
	fillColor Color
	lineColor Color

	// Gradients (the gradient array becomes float ColorType)
	fillGradient       [256]color.RGBA32[color.Linear] //nolint:unused // wired in L5 (gradient building)
	lineGradient       [256]color.RGBA32[color.Linear] //nolint:unused // wired in L5 (gradient building)
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
	textAngle         float64
	textAlignX        TextAlignment
	textAlignY        TextAlignment
	textHints         bool
	textForceAutohint bool
	flipText          bool
	resolution        uint
	fontHeight        float64
	fontAscent        float64
	fontDescent       float64
	fontCacheType     FontCacheType

	fontEngine       *freetype.FontEngineFreetype //nolint:unused // wired in L5 (text state plumbing)
	fontCacheManager *font.FontCacheManager       //nolint:unused // wired in L5 (text state plumbing)

	gsvText     *gsv.GSVText //nolint:unused // wired in L5 (text state plumbing)
	gsvFontMode bool         //nolint:unused // wired in L5 (text state plumbing)

	// Image filtering (color-agnostic)
	imageFilter               ImageFilter
	imageResample             ImageResample
	affineImageResamplePolicy AffineImageResamplePolicy
	imageFilterLUT            *aggimage.ImageFilterLUT

	// Fill mode
	evenOddFlag bool

	// Path and transformation (reused as-is)
	path           *path.PathStorageStl
	transform      *transform.TransAffine
	transformStack *TransformStack //nolint:unused // reserved for push/pop transform (not yet mirrored)

	// Converters (reused as-is)
	convCurve  *conv.ConvCurve
	convDash   *conv.ConvDash //nolint:unused // wired in L5 (dashed strokes)
	convStroke *conv.ConvStroke

	// Span rendering components for gradients (float)
	spanAllocator   *span.SpanAllocator[color.RGBA32[color.Linear]]
	fillGradientLUT []color.RGBA32[color.Linear]
	lineGradientLUT []color.RGBA32[color.Linear]

	fillLinearSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	lineLinearSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	fillRadialSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]
	lineRadialSpanInterpolator *span.SpanInterpolatorLinear[*transform.TransAffine]

	fillLinearSpanGenerator *span.SpanGradient[
		color.RGBA32[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientLinearX,
		*span.GradientPrebuiltColorRGBA32[color.Linear],
	]
	lineLinearSpanGenerator *span.SpanGradient[
		color.RGBA32[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientLinearX,
		*span.GradientPrebuiltColorRGBA32[color.Linear],
	]
	fillRadialSpanGenerator *span.SpanGradient[
		color.RGBA32[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientRadial,
		*span.GradientPrebuiltColorRGBA32[color.Linear],
	]
	lineRadialSpanGenerator *span.SpanGradient[
		color.RGBA32[color.Linear],
		*span.SpanInterpolatorLinear[*transform.TransAffine],
		span.GradientRadial,
		*span.GradientPrebuiltColorRGBA32[color.Linear],
	]

	fillGradientLUTDirty bool
	lineGradientLUTDirty bool

	// Control point tracking for smooth curves
	lastCtrlX, lastCtrlY float64 //nolint:unused // wired in L5 (smooth curve path methods)
	hasLastCtrl          bool    //nolint:unused // wired in L5 (smooth curve path methods)
}

// NewAgg2DFloat creates a new float AGG2D rendering context, mirroring NewAgg2D.
func NewAgg2DFloat() *Agg2DFloat {
	a := &Agg2DFloat{
		rbuf:                      buffer.NewRenderingBufferF32(),
		clipBox:                   struct{ X1, Y1, X2, Y2 float64 }{0, 0, 0, 0},
		blendMode:                 BlendAlpha,
		imageBlendMode:            BlendDst,
		imageBlendColor:           NewColor(0, 0, 0, 255),
		masterAlpha:               1.0,
		antiAliasGamma:            1.0,
		fillColor:                 White,
		lineColor:                 Black,
		fillGradientFlag:          Solid,
		lineGradientFlag:          Solid,
		fillGradientD1:            0.0,
		lineGradientD1:            0.0,
		fillGradientD2:            100.0,
		lineGradientD2:            100.0,
		textAngle:                 0.0,
		textAlignX:                AlignLeft,
		textAlignY:                AlignBottom,
		textHints:                 true,
		textForceAutohint:         false,
		resolution:                72,
		fontHeight:                0.0,
		fontAscent:                0.0,
		fontDescent:               0.0,
		fontCacheType:             RasterFontCache,
		imageFilter:               ImageFilterBilinear,
		imageResample:             NoResample,
		affineImageResamplePolicy: AffineImageResampleAgg2D,
		imageFilterLUT:            aggimage.NewImageFilterLUTWithFilter(aggimage.BilinearFilter{}, true),
		lineWidth:                 1.0,
		lineCap:                   CapRound,
		lineJoin:                  JoinRound,
		evenOddFlag:               false,
		path:                      path.NewPathStorageStl(),
		transform:                 transform.NewTransAffine(),
		fillGradientMatrix:        transform.NewTransAffine(),
		lineGradientMatrix:        transform.NewTransAffine(),
		scanline:                  scanline.NewScanlineU8(),
	}

	// Initialize converters (reused as-is)
	pathAdapter := path.NewPathStorageStlVertexSourceAdapter(a.path)
	a.convCurve = conv.NewConvCurve(pathAdapter)
	a.convStroke = conv.NewConvStroke(a.convCurve)

	// Initialize rasterizer with default cell block limit and clipper
	clipper := rasterizer.NewRasterizerSlNoClip()
	rconv := rasterizer.RasConvInt{}
	a.rasterizer = rasterizer.NewRasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip](rconv, clipper)

	// Initialize float span allocator and gradient LUTs
	a.spanAllocator = span.NewSpanAllocator[color.RGBA32[color.Linear]]()
	a.fillGradientLUT = make([]color.RGBA32[color.Linear], 256)
	a.lineGradientLUT = make([]color.RGBA32[color.Linear], 256)

	a.fillLinearSpanInterpolator = span.NewSpanInterpolatorLinearDefault(a.fillGradientMatrix)
	a.lineLinearSpanInterpolator = span.NewSpanInterpolatorLinearDefault(a.lineGradientMatrix)
	a.fillRadialSpanInterpolator = span.NewSpanInterpolatorLinearDefault(a.fillGradientMatrix)
	a.lineRadialSpanInterpolator = span.NewSpanInterpolatorLinearDefault(a.lineGradientMatrix)

	a.fillLinearSpanGenerator = span.NewLinearGradientFromLUT32(
		a.fillLinearSpanInterpolator, a.fillGradientLUT, a.fillGradientD1, a.fillGradientD2,
	)
	a.lineLinearSpanGenerator = span.NewLinearGradientFromLUT32(
		a.lineLinearSpanInterpolator, a.lineGradientLUT, a.lineGradientD1, a.lineGradientD2,
	)
	a.fillRadialSpanGenerator = span.NewRadialGradientFromLUT32(
		a.fillRadialSpanInterpolator, a.fillGradientLUT, a.fillGradientD1, a.fillGradientD2,
	)
	a.lineRadialSpanGenerator = span.NewRadialGradientFromLUT32(
		a.lineRadialSpanInterpolator, a.lineGradientLUT, a.lineGradientD1, a.lineGradientD2,
	)

	a.fillGradientLUTDirty = true
	a.lineGradientLUTDirty = true

	if a.convStroke != nil {
		a.convStroke.SetLineCap(basics.LineCap(CapRound))
		a.convStroke.SetLineJoin(basics.LineJoin(JoinRound))
	}

	return a
}

// colorToRGBA32 converts the 8-bit public Color to the straight float color used
// by the float pixfmt and renderers.
func colorToRGBA32(c Color) color.RGBA32[color.Linear] {
	return color.RGBA32[color.Linear]{
		R: float32(c[0]) / 255,
		G: float32(c[1]) / 255,
		B: float32(c[2]) / 255,
		A: float32(c[3]) / 255,
	}
}

// Attach attaches a float rendering buffer. stride is in bytes per row.
func (a *Agg2DFloat) Attach(buf []float32, width, height, stride int) {
	a.rbuf.Attach(buf, width, height, stride)

	a.ResetTransformations()
	a.lineWidth = 1.0
	a.lineColor = Black
	a.fillColor = White
	a.textAlignX = AlignLeft
	a.textAlignY = AlignBottom
	a.ClipBox(0, 0, float64(width), float64(height))
	a.lineCap = CapRound
	a.lineJoin = JoinRound
	a.flipText = false
	a.imageFilter = ImageFilterBilinear
	a.imageResample = NoResample
	a.affineImageResamplePolicy = AffineImageResampleAgg2D
	a.masterAlpha = 1.0
	a.antiAliasGamma = 1.0
	a.blendMode = BlendAlpha

	a.initializeRendering()
	a.updateRasterizerGamma()
}

// AttachImageFloat attaches the rendering context to an existing float image.
func (a *Agg2DFloat) AttachImageFloat(img *ImageFloat) {
	if img == nil || img.renBuf == nil {
		return
	}
	a.Attach(img.renBuf.Buf(), img.renBuf.Width(), img.renBuf.Height(), img.renBuf.Stride())
}

// ResetTransformations resets the world transform to identity.
func (a *Agg2DFloat) ResetTransformations() {
	a.transform.Reset()
}

// initializeRendering sets up the float pixel formats and base renderers.
func (a *Agg2DFloat) initializeRendering() {
	width := a.rbuf.Width()
	height := a.rbuf.Height()
	if width <= 0 || height <= 0 {
		return
	}

	a.pixfmt = pixfmt.NewPixFmtRGBA128Plain(a.rbuf)
	a.pixfmtPre = pixfmt.NewPixFmtRGBA128Pre(a.rbuf)
	a.renBase = newBaseRendererAdapter[color.RGBA32[color.Linear]](a.pixfmt)
	a.renBasePre = newBaseRendererAdapter[color.RGBA32[color.Linear]](a.pixfmtPre)

	// Composite pixfmts default to source-over; SetBlendMode swaps the operator.
	a.pixfmtComp = pixfmt.NewPixFmtCompositeRGBA128Linear(a.rbuf, blender.CompOpSrcOver)
	a.pixfmtCompPre = pixfmt.NewPixFmtCompositeRGBA128PreLinear(a.rbuf, blender.CompOpSrcOver)
	a.renBaseComp = newBaseRendererAdapter[color.RGBA32[color.Linear]](a.pixfmtComp)
	a.renBaseCompPre = newBaseRendererAdapter[color.RGBA32[color.Linear]](a.pixfmtCompPre)

	a.ClipBox(a.clipBox.X1, a.clipBox.Y1, a.clipBox.X2, a.clipBox.Y2)

	if a.rasterizer != nil {
		a.rasterizer.Reset()
		a.rasterizer.ClipBox(0, 0, float64(width), float64(height))
	}
}

// ClipBox sets the clipping rectangle.
func (a *Agg2DFloat) ClipBox(x1, y1, x2, y2 float64) {
	a.clipBox.X1, a.clipBox.Y1, a.clipBox.X2, a.clipBox.Y2 = x1, y1, x2, y2

	rx1, ry1, rx2, ry2 := int(x1), int(y1), int(x2), int(y2)
	if a.renBase != nil {
		a.renBase.ClipBox(rx1, ry1, rx2, ry2)
	}
	if a.renBasePre != nil {
		a.renBasePre.ClipBox(rx1, ry1, rx2, ry2)
	}
	if a.renBaseComp != nil {
		a.renBaseComp.ClipBox(rx1, ry1, rx2, ry2)
	}
	if a.renBaseCompPre != nil {
		a.renBaseCompPre.ClipBox(rx1, ry1, rx2, ry2)
	}
	if a.rasterizer != nil {
		a.rasterizer.ClipBox(x1, y1, x2, y2)
	}
}

// GetBounds returns the current rendering bounds.
func (a *Agg2DFloat) GetBounds() struct{ X1, Y1, X2, Y2 float64 } {
	return a.clipBox
}

// WorldToScreen transforms world coordinates to screen coordinates.
func (a *Agg2DFloat) WorldToScreen(x, y *float64) { a.transform.Transform(x, y) }

// ScreenToWorld transforms screen coordinates to world coordinates.
func (a *Agg2DFloat) ScreenToWorld(x, y *float64) { a.transform.InverseTransform(x, y) }

// ClearAll fills the entire buffer with the specified color.
func (a *Agg2DFloat) ClearAll(c Color) {
	if a.pixfmt == nil {
		return
	}
	a.pixfmt.Clear(colorToRGBA32(c))
}

// FillColor sets the fill color.
func (a *Agg2DFloat) FillColor(c Color) {
	a.fillColor = c
	a.fillGradientFlag = Solid
}

// LineColor sets the line color.
func (a *Agg2DFloat) LineColor(c Color) {
	a.lineColor = c
	a.lineGradientFlag = Solid
}
