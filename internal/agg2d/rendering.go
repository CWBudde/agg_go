// Package agg2d rendering pipeline for AGG2D high-level interface.
// This file contains rendering pipeline methods and functionality.
package agg2d

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/renderer/scanline"
)

// Rendering methods

// renderFill renders the current path as a filled shape
func (agg2d *Agg2D) renderFill() {
	if agg2d.rasterizer == nil || agg2d.path == nil || agg2d.scanline == nil {
		return
	}

	// Reset rasterizer for new path
	agg2d.rasterizer.Reset()

	// Apply fill rule (even-odd or non-zero winding)
	if agg2d.evenOddFlag {
		agg2d.rasterizer.SetFillRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.SetFillRule(basics.FillNonZero)
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
		agg2d.rasterizer.AddVertex(x, y, cmd)
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
	agg2d.rasterizer.SetFillRule(basics.FillNonZero)

	// Create stroke path (potentially with dashes)
	var strokeSource conv.VertexSource
	if agg2d.convDash != nil {
		// Use dashed stroke
		strokeSource = conv.NewConvTransform(agg2d.convDash, agg2d.transform)
	} else {
		// Use regular stroke
		strokeSource = conv.NewConvTransform(agg2d.convStroke, agg2d.transform)
	}

	// Add stroked path vertices to rasterizer
	strokeSource.Rewind(0)
	for {
		x, y, cmd := strokeSource.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		agg2d.rasterizer.AddVertex(x, y, cmd)
	}

	// Render with appropriate color/gradient
	if agg2d.lineGradientFlag == Solid {
		agg2d.renderSolidStroke()
	} else {
		agg2d.renderGradientStroke()
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
		agg2d.rasterizer.SetFillRule(basics.FillEvenOdd)
	} else {
		agg2d.rasterizer.SetFillRule(basics.FillNonZero)
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
		agg2d.rasterizer.AddVertex(x, y, cmd)
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

// renderSolidFillWithColor renders solid fill using specified color
func (agg2d *Agg2D) renderSolidFillWithColor(c Color) {
	if agg2d.renBase == nil {
		return
	}

	// Convert Color to internal color format
	internalColor := color.RGBA8[color.Linear]{R: c[0], G: c[1], B: c[2], A: c[3]}

	// Create solid renderer
	renSolid := scanline.NewRendererScanlineAASolidWithColor(agg2d.renBase, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
}

// renderSolidStroke renders solid stroke using current line color
func (agg2d *Agg2D) renderSolidStroke() {
	if agg2d.renBase == nil {
		return
	}

	// Convert Color to internal color format
	internalColor := color.RGBA8[color.Linear]{R: agg2d.lineColor[0], G: agg2d.lineColor[1], B: agg2d.lineColor[2], A: agg2d.lineColor[3]}

	// Create solid renderer
	renSolid := scanline.NewRendererScanlineAASolidWithColor(agg2d.renBase, internalColor)

	// Render scanlines
	scanlineRender(agg2d.rasterizer, agg2d.scanline, renSolid)
}

// renderGradientFill renders gradient fill (placeholder for now)
func (agg2d *Agg2D) renderGradientFill() {
	// TODO: Implement gradient fill rendering
	// For now, fall back to solid fill
	agg2d.renderSolidFill()
}

// renderGradientStroke renders gradient stroke (placeholder for now)
func (agg2d *Agg2D) renderGradientStroke() {
	// TODO: Implement gradient stroke rendering
	// For now, fall back to solid stroke
	agg2d.renderSolidStroke()
}

// renderGradientFillWithLineGradient renders fill using line gradient (placeholder for now)
func (agg2d *Agg2D) renderGradientFillWithLineGradient() {
	// TODO: Implement gradient fill using line gradient
	// For now, fall back to solid fill with line color
	agg2d.renderSolidFillWithColor(agg2d.lineColor)
}

// scanlineRender is a helper function to render scanlines
func scanlineRender[R any, S any](rasterizer interface{}, scanline S, renderer R) {
	// TODO: Implement proper scanline rendering
	// This is a placeholder that needs to be implemented based on the actual
	// interfaces provided by the rasterizer and scanline types
}

// updateApproximationScales updates the approximation scale for curve converters
// based on the current transformation matrix scaling
func (agg2d *Agg2D) updateApproximationScales() {
	if agg2d.convCurve != nil {
		// Calculate overall scaling factor from transformation matrix
		scaleX := math.Sqrt(agg2d.transform.SX*agg2d.transform.SX + agg2d.transform.SHY*agg2d.transform.SHY)
		scaleY := math.Sqrt(agg2d.transform.SHX*agg2d.transform.SHX + agg2d.transform.SY*agg2d.transform.SY)
		scale := (scaleX + scaleY) / 2.0

		// Update curve approximation scale
		// TODO: Implement proper curve approximation scale setting
		// agg2d.convCurve.SetApproximationScale(scale)
		_ = scale // Avoid unused variable warning for now
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
	if agg2d.rasterizer != nil && agg2d.antiAliasGamma != 1.0 {
		// TODO: Implement gamma correction in rasterizer
		// agg2d.rasterizer.SetGamma(agg2d.antiAliasGamma)
	}
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
		switch cap {
		case 0: // CapButt
			agg2d.convStroke.SetLineCap(0) // basics.ButtCap
		case 2: // CapSquare
			agg2d.convStroke.SetLineCap(2) // basics.SquareCap
		case 1: // CapRound
			agg2d.convStroke.SetLineCap(1) // basics.RoundCap
		}
	}
}

// LineJoin sets the line join style.
func (agg2d *Agg2D) LineJoin(join LineJoin) {
	agg2d.lineJoin = join
	if agg2d.convStroke != nil {
		switch join {
		case 0: // JoinMiter
			agg2d.convStroke.SetLineJoin(0) // basics.MiterJoin
		case 1: // JoinRound
			agg2d.convStroke.SetLineJoin(1) // basics.RoundJoin
		case 2: // JoinBevel
			agg2d.convStroke.SetLineJoin(2) // basics.BevelJoin
		}
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
	// TODO: Apply master alpha to renderers
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

// Math helper functions (simplified)
func cos(x float64) float64 {
	return math.Cos(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}
