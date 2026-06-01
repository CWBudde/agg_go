// Package agg2d float rendering pipeline (L5). Float twin of rendering.go: the
// genuinely color-dependent render path for Agg2DFloat, instantiated over the
// float RGBA32 color type. The rasterizer, scanline, converters, and transform
// are color-agnostic and reused from the shared subsystems.
package agg2d

import (
	"math"

	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/conv"
	renscan "github.com/cwbudde/agg_go/internal/renderer/scanline"
	"github.com/cwbudde/agg_go/internal/transform"
)

// refreshFillGradientLUTIfDirty copies the float gradient array into the LUT.
func (a *Agg2DFloat) refreshFillGradientLUTIfDirty() {
	if !a.fillGradientLUTDirty {
		return
	}
	copy(a.fillGradientLUT, a.fillGradient[:])
	a.fillGradientLUTDirty = false
}

func (a *Agg2DFloat) refreshLineGradientLUTIfDirty() {
	if !a.lineGradientLUTDirty {
		return
	}
	copy(a.lineGradientLUT, a.lineGradient[:])
	a.lineGradientLUTDirty = false
}

// currentRenderer returns the active base renderer: the composite renderer when
// a Porter-Duff / SVG blend mode is selected, otherwise the plain alpha
// renderer. Mirrors the 8-bit Agg2D.currentRenderer.
func (a *Agg2DFloat) currentRenderer() *baseRendererAdapter[color.RGBA32[color.Linear]] {
	if a.blendMode != BlendAlpha && a.renBaseComp != nil {
		return a.renBaseComp
	}
	return a.renBase
}

// applyMasterAlpha scales a straight float color's alpha by master alpha.
func (a *Agg2DFloat) applyMasterAlpha(c color.RGBA32[color.Linear]) color.RGBA32[color.Linear] {
	c.A *= float32(a.masterAlpha)
	return c
}

// renderFill renders the current path as a filled shape.
func (a *Agg2DFloat) renderFill() {
	if a.rasterizer == nil || a.path == nil || a.scanline == nil {
		return
	}
	a.rasterizer.Reset()
	if a.evenOddFlag {
		a.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		a.rasterizer.FillingRule(basics.FillNonZero)
	}

	transformedPath := conv.NewConvTransform(a.convCurve, a.transform)
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		a.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	if a.fillGradientFlag == Solid {
		a.renderSolidFillWithColor(a.fillColor)
	} else {
		a.renderGradientFill()
	}
}

// renderStroke renders the current path as a stroked outline.
func (a *Agg2DFloat) renderStroke() {
	if a.rasterizer == nil || a.path == nil || a.convStroke == nil || a.scanline == nil {
		return
	}
	a.rasterizer.Reset()
	a.rasterizer.FillingRule(basics.FillNonZero)

	if a.convDash != nil && a.convDash.NumDashes() == 0 {
		a.addStrokeToRasterizer(conv.NewConvStroke(a.convCurve))
	} else {
		a.addStrokeToRasterizer(a.convStroke)
	}

	if a.lineGradientFlag == Solid {
		a.renderSolidStroke()
	} else {
		a.renderGradientStroke()
	}
}

func (a *Agg2DFloat) addStrokeToRasterizer(stroke *conv.ConvStroke) {
	stroke.SetWidth(a.lineWidth)
	stroke.SetLineCap(basics.LineCap(a.lineCap))
	stroke.SetLineJoin(basics.LineJoin(a.lineJoin))
	strokeSource := conv.NewConvTransform(stroke, a.transform)
	strokeSource.Rewind(0)
	for {
		x, y, cmd := strokeSource.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		a.rasterizer.AddVertex(x, y, uint32(cmd))
	}
}

// renderFillWithLineColor fills the current path using the line color/gradient.
func (a *Agg2DFloat) renderFillWithLineColor() {
	if a.rasterizer == nil || a.path == nil || a.scanline == nil {
		return
	}
	a.rasterizer.Reset()
	if a.evenOddFlag {
		a.rasterizer.FillingRule(basics.FillEvenOdd)
	} else {
		a.rasterizer.FillingRule(basics.FillNonZero)
	}

	transformedPath := conv.NewConvTransform(a.convCurve, a.transform)
	transformedPath.Rewind(0)
	for {
		x, y, cmd := transformedPath.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		a.rasterizer.AddVertex(x, y, uint32(cmd))
	}

	if a.lineGradientFlag == Solid {
		a.renderSolidFillWithColor(a.lineColor)
	} else {
		a.renderGradientFillWithLineGradient()
	}
}

// renderSolidFillWithColor renders solid fill using the specified 8-bit color.
func (a *Agg2DFloat) renderSolidFillWithColor(c Color) {
	renderer := a.currentRenderer()
	if renderer == nil {
		return
	}
	internalColor := a.applyMasterAlpha(colorToRGBA32(c))
	renSolid := renscan.NewRendererScanlineAASolidWithColor(renderer, internalColor)
	a.scanlineRender(renSolid)
}

// RenderRasterizerWithColor renders the currently accumulated rasterizer content
// using the provided solid color, without resetting it first.
func (a *Agg2DFloat) RenderRasterizerWithColor(c Color) {
	a.renderSolidFillWithColor(c)
}

func (a *Agg2DFloat) renderSolidStroke() {
	a.renderSolidFillWithColor(a.lineColor)
}

func (a *Agg2DFloat) renderGradientFill() {
	switch a.fillGradientFlag {
	case Linear:
		a.renderLinearGradientFill(true)
	case Radial:
		a.renderRadialGradientFill(true)
	default:
		a.renderSolidFillWithColor(a.fillColor)
	}
}

func (a *Agg2DFloat) renderGradientStroke() {
	switch a.lineGradientFlag {
	case Linear:
		a.renderLinearGradientFill(false)
	case Radial:
		a.renderRadialGradientFill(false)
	default:
		a.renderSolidStroke()
	}
}

func (a *Agg2DFloat) renderGradientFillWithLineGradient() {
	switch a.lineGradientFlag {
	case Linear:
		a.renderLinearGradientFill(false)
	case Radial:
		a.renderRadialGradientFill(false)
	default:
		a.renderSolidFillWithColor(a.lineColor)
	}
}

func (a *Agg2DFloat) renderLinearGradientFill(useFillGradient bool) {
	renderer := a.currentRenderer()
	if renderer == nil || a.spanAllocator == nil {
		return
	}

	var gradientMatrix *transform.TransAffine
	var d1, d2 float64
	var spanGenerator renscan.SpanGeneratorInterface[color.RGBA32[color.Linear]]

	if useFillGradient {
		gradientMatrix, d1, d2 = a.fillGradientMatrix, a.fillGradientD1, a.fillGradientD2
		a.refreshFillGradientLUTIfDirty()
		a.fillLinearSpanInterpolator.SetTransformer(gradientMatrix)
		a.fillLinearSpanGenerator.SetD1(d1)
		a.fillLinearSpanGenerator.SetD2(d2)
		spanGenerator = a.fillLinearSpanGenerator
	} else {
		gradientMatrix, d1, d2 = a.lineGradientMatrix, a.lineGradientD1, a.lineGradientD2
		a.refreshLineGradientLUTIfDirty()
		a.lineLinearSpanInterpolator.SetTransformer(gradientMatrix)
		a.lineLinearSpanGenerator.SetD1(d1)
		a.lineLinearSpanGenerator.SetD2(d2)
		spanGenerator = a.lineLinearSpanGenerator
	}

	renscan.RenderScanlinesAA(a.rasterizer, a.scanline, renderer, a.spanAllocator, spanGenerator)
}

func (a *Agg2DFloat) renderRadialGradientFill(useFillGradient bool) {
	renderer := a.currentRenderer()
	if renderer == nil || a.spanAllocator == nil {
		return
	}

	var gradientMatrix *transform.TransAffine
	var d1, d2 float64
	var spanGenerator renscan.SpanGeneratorInterface[color.RGBA32[color.Linear]]

	if useFillGradient {
		gradientMatrix, d1, d2 = a.fillGradientMatrix, a.fillGradientD1, a.fillGradientD2
		a.refreshFillGradientLUTIfDirty()
		a.fillRadialSpanInterpolator.SetTransformer(gradientMatrix)
		a.fillRadialSpanGenerator.SetD1(d1)
		a.fillRadialSpanGenerator.SetD2(d2)
		spanGenerator = a.fillRadialSpanGenerator
	} else {
		gradientMatrix, d1, d2 = a.lineGradientMatrix, a.lineGradientD1, a.lineGradientD2
		a.refreshLineGradientLUTIfDirty()
		a.lineRadialSpanInterpolator.SetTransformer(gradientMatrix)
		a.lineRadialSpanGenerator.SetD1(d1)
		a.lineRadialSpanGenerator.SetD2(d2)
		spanGenerator = a.lineRadialSpanGenerator
	}

	renscan.RenderScanlinesAA(a.rasterizer, a.scanline, renderer, a.spanAllocator, spanGenerator)
}

// scanlineRender sweeps the rasterizer into the given renderer.
func (a *Agg2DFloat) scanlineRender(renderer renscan.RendererInterface[color.RGBA32[color.Linear]]) {
	ras := a.rasterizer
	sl := a.scanline
	if !ras.RewindScanlines() {
		return
	}
	sl.Reset(ras.MinX(), ras.MaxX())
	renderer.Prepare()
	for ras.SweepScanline(sl) {
		renderer.Render(sl)
	}
}

func (a *Agg2DFloat) updateApproximationScales() {
	scale := a.WorldToScreenScalar(1.0) * ApproxScale
	if a.convCurve != nil {
		a.convCurve.SetApproximationScale(scale)
	}
	if a.convStroke != nil {
		a.convStroke.SetApproximationScale(scale)
	}
}

func (a *Agg2DFloat) updateRasterizerGamma() {
	if a.rasterizer == nil {
		return
	}
	gamma := a.antiAliasGamma
	alpha := a.masterAlpha
	a.rasterizer.SetGamma(func(x float64) float64 {
		if x <= 0.0 {
			return 0.0
		}
		if x >= 1.0 {
			return alpha
		}
		return alpha * math.Pow(x, 1.0/gamma)
	})
}

// WorldToScreenScalar converts a world scalar to screen units.
func (a *Agg2DFloat) WorldToScreenScalar(scalar float64) float64 {
	x1, y1 := 0.0, 0.0
	x2, y2 := scalar, scalar
	a.WorldToScreen(&x1, &y1)
	a.WorldToScreen(&x2, &y2)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx+dy*dy) / math.Sqrt(2.0)
}

// LineWidth sets the line width.
func (a *Agg2DFloat) LineWidth(w float64) {
	a.lineWidth = w
	if a.convStroke != nil {
		a.convStroke.SetWidth(w)
	}
}

// LineCap sets the line cap style.
func (a *Agg2DFloat) LineCap(lineCap LineCap) {
	a.lineCap = lineCap
	if a.convStroke != nil {
		a.convStroke.SetLineCap(basics.LineCap(lineCap))
	}
}

// LineJoin sets the line join style.
func (a *Agg2DFloat) LineJoin(join LineJoin) {
	a.lineJoin = join
	if a.convStroke != nil {
		a.convStroke.SetLineJoin(basics.LineJoin(join))
	}
}

// GetLineWidth returns the current line width.
func (a *Agg2DFloat) GetLineWidth() float64 { return a.lineWidth }

// SetMasterAlpha sets the master alpha and updates the rasterizer gamma.
func (a *Agg2DFloat) SetMasterAlpha(alpha float64) {
	if alpha < 0.0 {
		alpha = 0.0
	} else if alpha > 1.0 {
		alpha = 1.0
	}
	a.masterAlpha = alpha
	a.updateRasterizerGamma()
}

// GetMasterAlpha returns the current master alpha.
func (a *Agg2DFloat) GetMasterAlpha() float64 { return a.masterAlpha }

// SetAntiAliasGamma sets the anti-alias gamma and updates the rasterizer gamma.
func (a *Agg2DFloat) SetAntiAliasGamma(gamma float64) {
	if gamma < 0.1 {
		gamma = 0.1
	} else if gamma > 3.0 {
		gamma = 3.0
	}
	a.antiAliasGamma = gamma
	a.updateRasterizerGamma()
}

// TextAlignment sets text alignment state.
func (a *Agg2DFloat) TextAlignment(alignX, alignY TextAlignment) {
	a.textAlignX = alignX
	a.textAlignY = alignY
}

// FlipText sets the text flip state.
func (a *Agg2DFloat) FlipText(flip bool) {
	a.flipText = flip
	if a.fontEngine != nil {
		a.fontEngine.SetFlipY(flip)
	}
}
