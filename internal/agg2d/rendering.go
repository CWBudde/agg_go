// Package agg2d rendering pipeline for AGG2D high-level interface.
// This file contains rendering pipeline methods and functionality.
package agg2d

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	aggimage "agg_go/internal/image"
	"agg_go/internal/rasterizer"
	renscan "agg_go/internal/renderer/scanline"
	"agg_go/internal/scanline"
	"agg_go/internal/transform"
)

// gradientColorsToLUTInPlace updates a preallocated 256-entry LUT from [256]Color.
func gradientColorsToLUTInPlace(dst []color.RGBA8[color.Linear], gradientColors [256]Color) {
	for i, c := range gradientColors {
		dst[i] = color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}
	}
}

func (agg2d *Agg2D) refreshFillGradientLUTIfDirty() {
	if !agg2d.fillGradientLUTDirty {
		return
	}
	gradientColorsToLUTInPlace(agg2d.fillGradientLUT, agg2d.fillGradient)
	agg2d.fillGradientLUTDirty = false
}

func (agg2d *Agg2D) refreshLineGradientLUTIfDirty() {
	if !agg2d.lineGradientLUTDirty {
		return
	}
	gradientColorsToLUTInPlace(agg2d.lineGradientLUT, agg2d.lineGradient)
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
		agg2d.addStrokeToRasterizer(conv.NewConvStroke(agg2d.convCurve))
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

	// Apply master alpha to color
	masterAlpha := uint8(agg2d.masterAlpha * 255.0)
	adjustedAlpha := uint8((uint16(c[3]) * uint16(masterAlpha)) / 255)

	// Convert Color to internal color format with master alpha applied
	internalColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: adjustedAlpha}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
}

// renderSolidStroke renders solid stroke using current line color
func (agg2d *Agg2D) renderSolidStroke() {
	renderer := agg2d.currentRenderer()
	if renderer == nil {
		return
	}

	// Apply master alpha to line color
	masterAlpha := uint8(agg2d.masterAlpha * 255.0)
	adjustedAlpha := uint8((uint16(agg2d.lineColor[3]) * uint16(masterAlpha)) / 255)

	// Convert Color to internal color format with master alpha applied
	internalColor := color.RGBA8[color.Linear]{R: agg2d.lineColor[0], G: agg2d.lineColor[1], B: agg2d.lineColor[2], A: adjustedAlpha}

	// Create solid renderer
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
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
	rasAdapter := &rasterizerAdapter{ras: agg2d.rasterizer}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, renderer, agg2d.spanAllocator, spanGenerator)
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
	rasAdapter := &rasterizerAdapter{ras: agg2d.rasterizer}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, renderer, agg2d.spanAllocator, spanGenerator)
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
	rasAdapter := &rasterizerAdapter{ras: ras}
	slAdapter := &scanlineWrapper{sl: agg2d.scanline}
	renscan.RenderScanlinesAA(rasAdapter, slAdapter, renderer, agg2d.spanAllocator, spanGen)
}

// scanlineRender is a helper function to render scanlines using a renderer
func scanlineRender(ras *rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip], sl *scanline.ScanlineU8, renderer renscan.RendererInterface[color.RGBA8[color.Linear]]) {
	// Create adapters to bridge interface differences
	rasAdapter := &rasterizerAdapter{ras: ras}
	slAdapter := &scanlineWrapper{sl: sl}

	if !rasAdapter.RewindScanlines() {
		return
	}

	// Reset scanline for the rasterizer bounds
	slAdapter.Reset(rasAdapter.MinX(), rasAdapter.MaxX())

	// Prepare the renderer
	renderer.Prepare()

	// Sweep through all scanlines
	for rasAdapter.SweepScanline(slAdapter) {
		renderer.Render(slAdapter)
	}
}

// updateApproximationScales updates the approximation scale for curve converters
// based on the current transformation matrix scaling
func (agg2d *Agg2D) updateApproximationScales() {
	if agg2d.convCurve != nil {
		// Use the world-to-screen scaling factor with the global approximation scale
		scale := agg2d.WorldToScreenScalar(1.0) * ApproxScale

		// Update curve approximation scale
		agg2d.convCurve.SetApproximationScale(scale)
	}

	if agg2d.convStroke != nil {
		// Also update the stroke converter with the same scale for consistency
		scale := agg2d.WorldToScreenScalar(1.0) * ApproxScale
		agg2d.convStroke.SetApproximationScale(scale)
	}
}

// render is the main rendering method that handles both fill and stroke colors
func (agg2d *Agg2D) render(fillColor bool) {
	if fillColor {
		agg2d.renderFill()
	} else {
		agg2d.renderStroke()
	}
}

// updateRasterizerGamma updates the rasterizer gamma correction
func (agg2d *Agg2D) updateRasterizerGamma() {
	if agg2d.rasterizer == nil {
		return
	}

	gamma := agg2d.antiAliasGamma
	alpha := agg2d.masterAlpha
	gammaFunc := func(x float64) float64 {
		if x <= 0.0 {
			return 0.0
		}
		if x >= 1.0 {
			return alpha
		}
		return alpha * math.Pow(x, 1.0/gamma)
	}
	agg2d.rasterizer.SetGamma(gammaFunc)
}

// LineWidth sets the line width.
func (agg2d *Agg2D) LineWidth(w float64) {
	agg2d.lineWidth = w
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetWidth(w)
	}
}

// LineCap sets the line cap style.
func (agg2d *Agg2D) LineCap(cap LineCap) {
	agg2d.lineCap = cap
	if agg2d.convStroke != nil {
		agg2d.convStroke.SetLineCap(basics.LineCap(cap))
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
func (agg2d *Agg2D) ImageFilter(f ImageFilter) {
	agg2d.imageFilter = f
	if agg2d.imageFilterLUT == nil {
		agg2d.imageFilterLUT = aggimage.NewImageFilterLUT()
	}

	switch f {
	case NoFilter:
		// AGG keeps the LUT unchanged for NoFilter.
		return
	case Bilinear:
		agg2d.imageFilterLUT.Calculate(aggimage.BilinearFilter{}, true)
	case Hanning:
		agg2d.imageFilterLUT.Calculate(aggimage.HanningFilter{}, true)
	case Hamming:
		agg2d.imageFilterLUT.Calculate(aggimage.HammingFilter{}, true)
	case Hermite:
		agg2d.imageFilterLUT.Calculate(aggimage.HermiteFilter{}, true)
	case Quadric:
		agg2d.imageFilterLUT.Calculate(aggimage.QuadricFilter{}, true)
	case Bicubic:
		agg2d.imageFilterLUT.Calculate(aggimage.BicubicFilter{}, true)
	case Catrom:
		agg2d.imageFilterLUT.Calculate(aggimage.CatromFilter{}, true)
	case Spline16:
		agg2d.imageFilterLUT.Calculate(aggimage.Spline16Filter{}, true)
	case Spline36:
		agg2d.imageFilterLUT.Calculate(aggimage.Spline36Filter{}, true)
	case Blackman:
		agg2d.imageFilterLUT.Calculate(aggimage.NewBlackmanFilter(4.0), true)
	case Kaiser:
		agg2d.imageFilterLUT.Calculate(aggimage.NewKaiserFilter(0), true)
	case Gaussian:
		agg2d.imageFilterLUT.Calculate(aggimage.GaussianFilter{}, true)
	case Bessel:
		agg2d.imageFilterLUT.Calculate(aggimage.BesselFilter{}, true)
	case Mitchell:
		agg2d.imageFilterLUT.Calculate(aggimage.NewMitchellFilter(0, 0), true)
	case Sinc:
		agg2d.imageFilterLUT.Calculate(aggimage.NewSincFilter(4.0), true)
	case Lanczos:
		agg2d.imageFilterLUT.Calculate(aggimage.NewLanczosFilter(4.0), true)
	default:
		agg2d.imageFilterLUT.Calculate(aggimage.BilinearFilter{}, true)
	}
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

// GetMasterAlpha returns the current master alpha value
func (agg2d *Agg2D) GetMasterAlpha() float64 {
	return agg2d.masterAlpha
}

// SetMasterAlpha sets the master alpha value
func (agg2d *Agg2D) SetMasterAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	} else if alpha > 1.0 {
		alpha = 1.0
	}
	agg2d.masterAlpha = alpha
	agg2d.updateRasterizerGamma()
}

// GetAntiAliasGamma returns the current anti-alias gamma value
func (agg2d *Agg2D) GetAntiAliasGamma() float64 {
	return agg2d.antiAliasGamma
}

// SetAntiAliasGamma sets the anti-alias gamma value
func (agg2d *Agg2D) SetAntiAliasGamma(gamma float64) {
	if gamma < 0.1 {
		gamma = 0.1
	} else if gamma > 3.0 {
		gamma = 3.0
	}
	agg2d.antiAliasGamma = gamma
	agg2d.updateRasterizerGamma()
}

// Math helpers local to the rendering package.
func cos(x float64) float64 {
	return math.Cos(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}
