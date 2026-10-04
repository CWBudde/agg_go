// Package agg2d rendering pipeline for AGG2D high-level interface.
// This file contains rendering pipeline methods and functionality.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	aggimage "github.com/cwbudde/agg_go/internal/image"
	"github.com/cwbudde/agg_go/internal/rasterizer"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

// gradientColorsToLUTInPlace updates a preallocated 256-entry LUT from [256]Color.
func gradientColorsToLUTInPlace(dst []color.RGBA8[color.Linear], gradientColors *[256]Color) {
	for i, c := range gradientColors {
		dst[i] = color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}
	}
}

func (agg2d *Agg2D) refreshFillGradientLUTIfDirty() {
	if !agg2d.fillGradientLUTDirty {
		return
	}
	gradientColorsToLUTInPlace(agg2d.fillGradientLUT, &agg2d.fillGradient)
	agg2d.fillGradientLUTDirty = false
}

func (agg2d *Agg2D) refreshLineGradientLUTIfDirty() {
	if !agg2d.lineGradientLUTDirty {
		return
	}
	gradientColorsToLUTInPlace(agg2d.lineGradientLUT, &agg2d.lineGradient)
	agg2d.lineGradientLUTDirty = false
}

// Rendering methods

func (agg2d *Agg2D) currentRenderer() *baseRendererAdapter[color.RGBA8[color.Linear]] {
	if agg2d.blendMode != BlendAlpha && agg2d.renBaseComp != nil {
		return agg2d.renBaseComp
	}
	return agg2d.renBase
}

func (agg2d *Agg2D) currentImageRenderer() *baseRendererAdapter[color.RGBA8[color.Linear]] {
	if agg2d.blendMode != BlendAlpha && agg2d.renBaseCompPre != nil {
		return agg2d.renBaseCompPre
	}
	return agg2d.renBasePre
}

// renderFill renders the current path as a filled shape
func (agg2d *Agg2D) renderFill() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Apply fill rule (even-odd or non-zero winding)
	if agg2d.evenOddFlag {
		agg2d.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.FillingRule(basics.FillNonZero)
	}

	// Create transformed curve converter
	transformedPath := conv.NewConvTransform(agg2d.convCurve, agg2d.transform)

	// Add path vertices to rasterizer
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Render with appropriate color/gradient
	if agg2d.fillGradientFlag == Solid {
		agg2d.renderSolidFill()
	} else {
		agg2d.renderGradientFill()
	}
}

// renderStroke renders the current path as a stroked outline
func (agg2d *Agg2D) renderStroke() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.convStroke == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Always use non-zero fill rule for strokes
	agg2d.rasterizer.FillingRule(basics.FillNonZero)

	// When convDash is in the pipeline but has no active dashes, bypass it and
	// stroke convCurve directly. This matches AGG C++ which uses separate
	// conv_stroke and conv_stroke<conv_dash> pipelines: when no dashes are set,
	// the plain conv_stroke<conv_curve> is used rather than the dashed one.
	if agg2d.convDash != nil && agg2d.convDash.NumDashes() == 0 {
		agg2d.addStrokeToRasterizer(agg2d.undashedStroke())
	} else {
		agg2d.addStrokeToRasterizer(agg2d.convStroke)
	}

	// Render with appropriate color/gradient
	if agg2d.lineGradientFlag == Solid {
		agg2d.renderSolidStroke()
	} else {
		agg2d.renderGradientStroke()
	}
}

// undashedStroke returns a conv_stroke over the curve converter that carries
// every setting of the (dashed) main stroke converter -- in particular the
// approximation scale, which C++ only updates from the transform setters.
func (agg2d *Agg2D) undashedStroke() *conv.ConvStroke {
	src := agg2d.convStroke
	stroke := conv.NewConvStroke(agg2d.convCurve)
	stroke.SetMiterLimit(src.MiterLimit())
	stroke.SetInnerMiterLimit(src.InnerMiterLimit())
	stroke.SetInnerJoin(src.InnerJoin())
	stroke.SetApproximationScale(src.ApproximationScale())
	stroke.SetShorten(src.Shorten())
	return stroke
}

// addStrokeToRasterizer applies the given stroke converter (with current settings)
// through the world transform and feeds vertices into the rasterizer.
func (agg2d *Agg2D) addStrokeToRasterizer(stroke *conv.ConvStroke) {
	stroke.SetWidth(agg2d.lineWidth)
	stroke.SetLineCap(basics.LineCap(agg2d.lineCap))
	stroke.SetLineJoin(basics.LineJoin(agg2d.lineJoin))
	strokeSource := conv.NewConvTransform(stroke, agg2d.transform)
	strokeSource.Rewind(0)
	for {
		x, y, cmd := strokeSource.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}
}

// renderFillWithLineColor renders the current path filled with line color
func (agg2d *Agg2D) renderFillWithLineColor() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Apply fill rule (even-odd or non-zero winding)
	if agg2d.evenOddFlag {
		agg2d.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.FillingRule(basics.FillNonZero)
	}

	// Create transformed curve converter
	transformedPath := conv.NewConvTransform(agg2d.convCurve, agg2d.transform)

	// Add path vertices to rasterizer
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	// Render using line color instead of fill color
	if agg2d.lineGradientFlag == Solid {
		agg2d.renderSolidFillWithColor(agg2d.lineColor)
	} else {
		agg2d.renderGradientFillWithLineGradient()
	}
}

// renderSolidFill renders solid fill using current fill color
func (agg2d *Agg2D) renderSolidFill() {
	agg2d.renderSolidFillWithColor(agg2d.fillColor)
}

// RenderRasterizerWithColor renders whatever is currently accumulated in the rasterizer
// using the provided solid color, without resetting it first.
// Use this after manually populating the rasterizer via GetInternalRasterizer().AddPath().
func (agg2d *Agg2D) RenderRasterizerWithColor(c Color) {
	agg2d.renderSolidFillWithColor(c)
}

// renderSolidFillWithColor renders solid fill using specified color
func (agg2d *Agg2D) renderSolidFillWithColor(c Color) {
	renderer := agg2d.currentRenderer()
	if renderer == nil {
		return
	}

	// Master alpha is not applied to the colour: C++ Agg2D folds it into the
	// rasterizer gamma (Agg2DRasterizerGamma, agg2d.cpp:1747) and passes the
	// unscaled colour to renSolid.color() (agg2d.cpp:1493).
	internalColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, internalColor)

	// Render scanlines
	agg2d.scanlineRender(renSolid)
}

// renderSolidStroke renders solid stroke using current line color
func (agg2d *Agg2D) renderSolidStroke() {
	renderer := agg2d.currentRenderer()
	if renderer == nil {
		return
	}

	// Unscaled colour; master alpha lives in the rasterizer gamma (see
	// renderSolidFillWithColor).
	c := agg2d.lineColor
	internalColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, internalColor)

	// Render scanlines
	agg2d.scanlineRender(renSolid)
}

// renderGradientFill renders gradient fill using the appropriate gradient type
func (agg2d *Agg2D) renderGradientFill() {
	switch agg2d.fillGradientFlag {
	case Linear:
		agg2d.renderLinearGradientFill(true) // true = use fill gradient settings
	case Radial:
		agg2d.renderRadialGradientFill(true) // true = use fill gradient settings
	default:
		// Solid fill fallback
		agg2d.renderSolidFill()
	}
}

// renderLinearGradientFill renders linear gradient fill
func (agg2d *Agg2D) renderLinearGradientFill(useFillGradient bool) {
	renderer := agg2d.currentRenderer()
	if renderer == nil || agg2d.spanAllocator == nil {
		return
	}

	// Choose the appropriate gradient settings
	var gradientMatrix *transform.TransAffine
	var d1, d2 float64

	if useFillGradient {
		gradientMatrix = agg2d.fillGradientMatrix
		d1 = agg2d.fillGradientD1
		d2 = agg2d.fillGradientD2
	} else {
		gradientMatrix = agg2d.lineGradientMatrix
		d1 = agg2d.lineGradientD1
		d2 = agg2d.lineGradientD2
	}

	var spanGenerator renscan.SpanGeneratorInterface[color.RGBA8[color.Linear]]
	if useFillGradient {
		agg2d.refreshFillGradientLUTIfDirty()
		agg2d.fillLinearSpanInterpolator.SetTransformer(gradientMatrix)
		agg2d.fillLinearSpanGenerator.SetD1(d1)
		agg2d.fillLinearSpanGenerator.SetD2(d2)
		spanGenerator = agg2d.fillLinearSpanGenerator
	} else {
		agg2d.refreshLineGradientLUTIfDirty()
		agg2d.lineLinearSpanInterpolator.SetTransformer(gradientMatrix)
		agg2d.lineLinearSpanGenerator.SetD1(d1)
		agg2d.lineLinearSpanGenerator.SetD2(d2)
		spanGenerator = agg2d.lineLinearSpanGenerator
	}

	// Render scanlines using the span generator directly
	renscan.RenderScanlinesAA(agg2d.rasterizer, agg2d.scanline, renderer, agg2d.spanAllocator, spanGenerator)
}

// renderRadialGradientFill renders radial gradient fill
func (agg2d *Agg2D) renderRadialGradientFill(useFillGradient bool) {
	renderer := agg2d.currentRenderer()
	if renderer == nil || agg2d.spanAllocator == nil {
		return
	}

	// Choose the appropriate gradient settings
	var gradientMatrix *transform.TransAffine
	var d1, d2 float64

	if useFillGradient {
		gradientMatrix = agg2d.fillGradientMatrix
		d1 = agg2d.fillGradientD1
		d2 = agg2d.fillGradientD2
	} else {
		gradientMatrix = agg2d.lineGradientMatrix
		d1 = agg2d.lineGradientD1
		d2 = agg2d.lineGradientD2
	}

	var spanGenerator renscan.SpanGeneratorInterface[color.RGBA8[color.Linear]]
	if useFillGradient {
		agg2d.refreshFillGradientLUTIfDirty()
		agg2d.fillRadialSpanInterpolator.SetTransformer(gradientMatrix)
		agg2d.fillRadialSpanGenerator.SetD1(d1)
		agg2d.fillRadialSpanGenerator.SetD2(d2)
		spanGenerator = agg2d.fillRadialSpanGenerator
	} else {
		agg2d.refreshLineGradientLUTIfDirty()
		agg2d.lineRadialSpanInterpolator.SetTransformer(gradientMatrix)
		agg2d.lineRadialSpanGenerator.SetD1(d1)
		agg2d.lineRadialSpanGenerator.SetD2(d2)
		spanGenerator = agg2d.lineRadialSpanGenerator
	}

	// Render scanlines using the span generator directly
	renscan.RenderScanlinesAA(agg2d.rasterizer, agg2d.scanline, renderer, agg2d.spanAllocator, spanGenerator)
}

// renderGradientStroke renders gradient stroke using line gradient settings
func (agg2d *Agg2D) renderGradientStroke() {
	switch agg2d.lineGradientFlag {
	case Linear:
		agg2d.renderLinearGradientFill(false) // false = use line gradient settings
	case Radial:
		agg2d.renderRadialGradientFill(false) // false = use line gradient settings
	default:
		// Solid stroke fallback
		agg2d.renderSolidStroke()
	}
}

// renderGradientFillWithLineGradient renders fill using line gradient settings
func (agg2d *Agg2D) renderGradientFillWithLineGradient() {
	switch agg2d.lineGradientFlag {
	case Linear:
		agg2d.renderLinearGradientFill(false) // false = use line gradient settings
	case Radial:
		agg2d.renderRadialGradientFill(false) // false = use line gradient settings
	default:
		// Solid fill fallback using line color
		agg2d.renderSolidFillWithColor(agg2d.lineColor)
	}
}

// RenderScanlinesAAWithSpanGen renders the rasterizer using a custom span generator.
// This enables advanced effects like combining color gradients with alpha gradients.
func (agg2d *Agg2D) RenderScanlinesAAWithSpanGen(
	ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip],
	spanGen renscan.SpanGeneratorInterface[color.RGBA8[color.Linear]],
) {
	renderer := agg2d.currentRenderer()
	if renderer == nil || agg2d.spanAllocator == nil {
		return
	}
	renscan.RenderScanlinesAA(ras, agg2d.scanline, renderer, agg2d.spanAllocator, spanGen)
}

// scanlineRender renders scanlines from the rasterizer using the cached adapters.
func (agg2d *Agg2D) scanlineRender(renderer renscan.RendererInterface[color.RGBA8[color.Linear]]) {
	ras := agg2d.rasterizer
	sl := agg2d.scanline

	if !ras.RewindScanlines() {
		return
	}

	sl.Reset(ras.MinX(), ras.MaxX())
	renderer.Prepare()

	for ras.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

// updateApproximationScales mirrors the C++ idiom
//
//	m_convCurve.approximation_scale(worldToScreen(1.0) * g_approxScale);
//	m_convStroke.approximation_scale(worldToScreen(1.0) * g_approxScale);
//
// which agg2d.cpp runs only from transformations(), affine(), scale(),
// parallelogram() and viewport() -- not from rotate/skew/translate,
// resetTransformations or drawPath.
func (agg2d *Agg2D) updateApproximationScales() {
	scale := agg2d.WorldToScreenScalar(1.0) * ApproxScale
	if agg2d.convCurve != nil {
		agg2d.convCurve.SetApproximationScale(scale)
	}
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetApproximationScale(scale)
	}
}

// rasterizerGamma is C++ Agg2DRasterizerGamma (agg2d.cpp:1747):
// gamma_multiply(alpha)(gamma_power(gamma)(x)) = min(alpha * pow(x, gamma), 1).
func rasterizerGamma(alpha, gamma float64) func(float64) float64 {
	return func(x float64) float64 {
		y := math.Pow(x, gamma) * alpha
		if y > 1.0 {
			y = 1.0
		}
		return y
	}
}

// binaryRasterizerGamma is the aliased (SetAntiAliased(false)) Go extension:
// AGG's scanline_bin path emits a fully covered pixel for every touched cell,
// regardless of partial coverage; master alpha still applies like
// gamma_multiply.
func binaryRasterizerGamma(alpha float64) func(float64) float64 {
	return func(x float64) float64 {
		if x <= 0.0 {
			return 0.0
		}
		if alpha > 1.0 {
			return 1.0
		}
		return alpha
	}
}

// updateRasterizerGamma mirrors C++ Agg2D::updateRasterizerGamma
// (agg2d.cpp:1762), called from masterAlpha() and antiAliasGamma().
func (agg2d *Agg2D) updateRasterizerGamma() {
	if agg2d.rasterizer == nil {
		return
	}
	if agg2d.antiAliasOff {
		agg2d.rasterizer.SetGamma(binaryRasterizerGamma(agg2d.masterAlpha))
		return
	}
	agg2d.rasterizer.SetGamma(rasterizerGamma(agg2d.masterAlpha, agg2d.antiAliasGamma))
}

// SetAntiAliased toggles between the anti-aliased scanline pipeline and the
// binary-coverage behavior of AGG's scanline_bin renderer. It mirrors how
// renderers built on AGG (e.g. matplotlib's Agg backend) switch to
// renderer_scanline_bin_solid when a graphics context requests
// antialiased=false.
func (agg2d *Agg2D) SetAntiAliased(enabled bool) {
	agg2d.antiAliasOff = !enabled
	agg2d.updateRasterizerGamma()
}

// GetAntiAliased reports whether the anti-aliased pipeline is active.
func (agg2d *Agg2D) GetAntiAliased() bool {
	return !agg2d.antiAliasOff
}

// LineWidth sets the line width.
func (agg2d *Agg2D) LineWidth(w float64) {
	agg2d.lineWidth = w
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetWidth(w)
	}
}

// LineCap sets the line cap style.
func (agg2d *Agg2D) LineCap(lineCap LineCap) {
	agg2d.lineCap = lineCap
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetLineCap(basics.LineCap(lineCap))
	}
}

// LineJoin sets the line join style.
func (agg2d *Agg2D) LineJoin(join LineJoin) {
	agg2d.lineJoin = join
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetLineJoin(basics.LineJoin(join))
	}
}

// ResetTransformations resets the transformation matrix to identity.
func (agg2d *Agg2D) ResetTransformations() {
	if agg2d.transform != nil {
		agg2d.transform.Reset()
	}
}

// ImageFilter sets the image filtering method.
// This matches the C++ Agg2D::imageFilter method (agg2d.cpp:1241).
func (agg2d *Agg2D) ImageFilter(f ImageFilter) {
	agg2d.imageFilter = f
	if agg2d.imageFilterLUT == nil {
		agg2d.imageFilterLUT = aggimage.NewImageFilterLUT()
	}
	calculateImageFilterLUT(agg2d.imageFilterLUT, f)
}

// calculateImageFilterLUT recomputes lut for f exactly like the switch in C++
// Agg2D::imageFilter (agg2d.cpp:1241-1258) for the C++ enum members: NoFilter
// leaves the LUT untouched; every other filter is normalized. The members
// after Spline36 other than Blackman144 are Go extensions.
func calculateImageFilterLUT(lut *aggimage.ImageFilterLUT, f ImageFilter) {
	var filter aggimage.FilterFunction
	switch f {
	case NoFilter:
		return
	case Bilinear:
		filter = aggimage.BilinearFilter{}
	case Hanning:
		filter = aggimage.HanningFilter{}
	case Hermite:
		filter = aggimage.HermiteFilter{}
	case Quadric:
		filter = aggimage.QuadricFilter{}
	case Bicubic:
		filter = aggimage.BicubicFilter{}
	case Catrom:
		filter = aggimage.CatromFilter{}
	case Spline16:
		filter = aggimage.Spline16Filter{}
	case Spline36:
		filter = aggimage.Spline36Filter{}
	case Blackman144:
		filter = aggimage.NewBlackman144Filter()
	// Go extensions (not in C++ Agg2D::ImageFilter).
	case Hamming:
		filter = aggimage.HammingFilter{}
	case Blackman:
		filter = aggimage.NewBlackmanFilter(4.0)
	case Kaiser:
		filter = aggimage.NewKaiserFilter(0)
	case Gaussian:
		filter = aggimage.GaussianFilter{}
	case Bessel:
		filter = aggimage.BesselFilter{}
	case Mitchell:
		filter = aggimage.NewMitchellFilter(0, 0)
	case Sinc:
		filter = aggimage.NewSincFilter(4.0)
	case Lanczos:
		filter = aggimage.NewLanczosFilter(4.0)
	default:
		filter = aggimage.BilinearFilter{}
	}
	lut.Calculate(filter, true)
}

// SetImageFilterRadius sets the image filtering method with a custom radius for supported filters.
func (agg2d *Agg2D) SetImageFilterRadius(f ImageFilter, radius float64) {
	agg2d.imageFilter = f
	if agg2d.imageFilterLUT == nil {
		agg2d.imageFilterLUT = aggimage.NewImageFilterLUT()
	}

	var funcObj aggimage.FilterFunction
	switch f {
	case Blackman:
		funcObj = aggimage.NewBlackmanFilter(radius)
	case Sinc:
		funcObj = aggimage.NewSincFilter(radius)
	case Lanczos:
		funcObj = aggimage.NewLanczosFilter(radius)
	default:
		agg2d.ImageFilter(f)
		return
	}
	agg2d.imageFilterLUT.Calculate(funcObj, true)
}

// ImageResample sets the image resampling method.
func (agg2d *Agg2D) ImageResample(r ImageResample) {
	agg2d.imageResample = r
}

// AffineImageResamplePolicy controls how affine image transforms choose
// between direct filtered spans and the affine resampler when ImageResample is
// NoResample. The default preserves Agg2D behavior.
func (agg2d *Agg2D) AffineImageResamplePolicy(policy AffineImageResamplePolicy) {
	agg2d.affineImageResamplePolicy = policy
}

// TextAlignment sets text alignment.
func (agg2d *Agg2D) TextAlignment(alignX, alignY TextAlignment) {
	agg2d.textAlignX = alignX
	agg2d.textAlignY = alignY
}

// GetLineWidth returns the current line width
func (agg2d *Agg2D) GetLineWidth() float64 {
	return agg2d.lineWidth
}

// GetLineCap returns the current line cap style
func (agg2d *Agg2D) GetLineCap() LineCap {
	return agg2d.lineCap
}

// GetLineJoin returns the current line join style
func (agg2d *Agg2D) GetLineJoin() LineJoin {
	return agg2d.lineJoin
}

// GetImageFilter returns the current image filter
func (agg2d *Agg2D) GetImageFilter() ImageFilter {
	return agg2d.imageFilter
}

// GetImageResample returns the current image resampling method
func (agg2d *Agg2D) GetImageResample() ImageResample {
	return agg2d.imageResample
}

// GetAffineImageResamplePolicy returns the current affine image resample policy.
func (agg2d *Agg2D) GetAffineImageResamplePolicy() AffineImageResamplePolicy {
	return agg2d.affineImageResamplePolicy
}

// GetMasterAlpha returns the current master alpha value
func (agg2d *Agg2D) GetMasterAlpha() float64 {
	return agg2d.masterAlpha
}

// SetMasterAlpha sets the master alpha value. Like C++ Agg2D::masterAlpha
// (agg2d.cpp:220) the value is not clamped; it is folded into the rasterizer
// gamma, where gamma_multiply clamps the product to 1.
func (agg2d *Agg2D) SetMasterAlpha(alpha float64) {
	agg2d.masterAlpha = alpha
	agg2d.updateRasterizerGamma()
}

// GetAntiAliasGamma returns the current anti-alias gamma value
func (agg2d *Agg2D) GetAntiAliasGamma() float64 {
	return agg2d.antiAliasGamma
}

// SetAntiAliasGamma sets the anti-alias gamma value. Like C++
// Agg2D::antiAliasGamma (agg2d.cpp:233) the value is not clamped; coverage is
// mapped through gamma_power, i.e. pow(cover, gamma).
func (agg2d *Agg2D) SetAntiAliasGamma(gamma float64) {
	agg2d.antiAliasGamma = gamma
	agg2d.updateRasterizerGamma()
}
