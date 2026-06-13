// Package agg2d float state-accessor / RGBA-setter / alias methods (L5/breadth).
// Color-agnostic style readbacks and convenience setters mirrored onto
// Agg2DFloat. The bodies match the 8-bit twin (colors.go, fill_rules.go,
// rendering.go, stroke.go, utilities.go) because they only touch style state
// that lives on the float struct in the same shape. The image-filter LUT logic
// is byte-identical to rendering.go since both twins share *aggimage.
package agg2d

import (
	aggimage "github.com/cwbudde/agg_go/internal/image"
)

// FillColorRGBA sets the fill color using RGBA components. Mirrors colors.go.
func (a *Agg2DFloat) FillColorRGBA(r, g, b, alpha uint8) {
	a.FillColor(NewColor(r, g, b, alpha))
}

// LineColorRGBA sets the line color using RGBA components. Mirrors colors.go.
func (a *Agg2DFloat) LineColorRGBA(r, g, b, alpha uint8) {
	a.LineColor(NewColor(r, g, b, alpha))
}

// ClearAllRGBA fills the entire buffer with the specified RGBA color. Mirrors
// colors.go.
func (a *Agg2DFloat) ClearAllRGBA(r, g, b, alpha uint8) {
	a.ClearAll(NewColor(r, g, b, alpha))
}

// GetFillColor returns the current fill color. Mirrors colors.go.
func (a *Agg2DFloat) GetFillColor() Color { return a.fillColor }

// GetLineColor returns the current line color. Mirrors colors.go.
func (a *Agg2DFloat) GetLineColor() Color { return a.lineColor }

// GetClipBox returns the current clipping rectangle. Mirrors colors.go.
func (a *Agg2DFloat) GetClipBox() (x1, y1, x2, y2 float64) {
	return a.clipBox.X1, a.clipBox.Y1, a.clipBox.X2, a.clipBox.Y2
}

// GetLineCap returns the current line cap style. Mirrors rendering.go.
func (a *Agg2DFloat) GetLineCap() LineCap { return a.lineCap }

// GetLineJoin returns the current line join style. Mirrors rendering.go.
func (a *Agg2DFloat) GetLineJoin() LineJoin { return a.lineJoin }

// GetImageFilter returns the current image filter. Mirrors rendering.go.
func (a *Agg2DFloat) GetImageFilter() ImageFilter { return a.imageFilter }

// GetImageResample returns the current image resampling method. Mirrors
// rendering.go.
func (a *Agg2DFloat) GetImageResample() ImageResample { return a.imageResample }

// GetAntiAliasGamma returns the current anti-alias gamma value. Mirrors
// rendering.go.
func (a *Agg2DFloat) GetAntiAliasGamma() float64 { return a.antiAliasGamma }

// MiterLimit sets the stroke miter limit. Mirrors stroke.go.
func (a *Agg2DFloat) MiterLimit(ml float64) {
	if a.convStroke != nil {
		a.convStroke.SetMiterLimit(ml)
	}
}

// GetMiterLimit returns the current miter limit. Mirrors stroke.go.
func (a *Agg2DFloat) GetMiterLimit() float64 {
	if a.convStroke != nil {
		return a.convStroke.MiterLimit()
	}
	return 4.0 // Default AGG miter limit
}

// ImageFilter selects the image filtering method and recalculates the shared
// filter LUT. Mirrors rendering.go.
func (a *Agg2DFloat) ImageFilter(f ImageFilter) {
	a.imageFilter = f
	if a.imageFilterLUT == nil {
		a.imageFilterLUT = aggimage.NewImageFilterLUT()
	}

	switch f {
	case NoFilter:
		// AGG keeps the LUT unchanged for NoFilter.
		return
	case Bilinear:
		a.imageFilterLUT.Calculate(aggimage.BilinearFilter{}, true)
	case Hanning:
		a.imageFilterLUT.Calculate(aggimage.HanningFilter{}, true)
	case Hamming:
		a.imageFilterLUT.Calculate(aggimage.HammingFilter{}, true)
	case Hermite:
		a.imageFilterLUT.Calculate(aggimage.HermiteFilter{}, true)
	case Quadric:
		a.imageFilterLUT.Calculate(aggimage.QuadricFilter{}, true)
	case Bicubic:
		a.imageFilterLUT.Calculate(aggimage.BicubicFilter{}, true)
	case Catrom:
		a.imageFilterLUT.Calculate(aggimage.CatromFilter{}, true)
	case Spline16:
		a.imageFilterLUT.Calculate(aggimage.Spline16Filter{}, true)
	case Spline36:
		a.imageFilterLUT.Calculate(aggimage.Spline36Filter{}, true)
	case Blackman:
		a.imageFilterLUT.Calculate(aggimage.NewBlackmanFilter(4.0), true)
	case Kaiser:
		a.imageFilterLUT.Calculate(aggimage.NewKaiserFilter(0), true)
	case Gaussian:
		a.imageFilterLUT.Calculate(aggimage.GaussianFilter{}, true)
	case Bessel:
		a.imageFilterLUT.Calculate(aggimage.BesselFilter{}, true)
	case Mitchell:
		a.imageFilterLUT.Calculate(aggimage.NewMitchellFilter(0, 0), true)
	case Sinc:
		a.imageFilterLUT.Calculate(aggimage.NewSincFilter(4.0), true)
	case Lanczos:
		a.imageFilterLUT.Calculate(aggimage.NewLanczosFilter(4.0), true)
	default:
		a.imageFilterLUT.Calculate(aggimage.BilinearFilter{}, true)
	}
}

// SetImageFilterRadius sets the image filtering method with a custom radius for
// supported filters. Mirrors rendering.go.
func (a *Agg2DFloat) SetImageFilterRadius(f ImageFilter, radius float64) {
	a.imageFilter = f
	if a.imageFilterLUT == nil {
		a.imageFilterLUT = aggimage.NewImageFilterLUT()
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
		a.ImageFilter(f)
		return
	}
	a.imageFilterLUT.Calculate(funcObj, true)
}

// ImageResample sets the image resampling method. Mirrors rendering.go.
func (a *Agg2DFloat) ImageResample(r ImageResample) { a.imageResample = r }

// IsEvenOddFillRule reports whether the even-odd fill rule is active. Mirrors
// fill_rules.go.
func (a *Agg2DFloat) IsEvenOddFillRule() bool { return a.GetFillEvenOdd() }

// IsNonZeroFillRule reports whether the non-zero winding fill rule is active.
// Mirrors fill_rules.go.
func (a *Agg2DFloat) IsNonZeroFillRule() bool { return !a.GetFillEvenOdd() }

// FillRuleDescription returns a human-readable description of the current fill
// rule. Mirrors fill_rules.go.
func (a *Agg2DFloat) FillRuleDescription() string {
	if a.evenOddFlag {
		return "Even-Odd (XOR-based filling)"
	}
	return "Non-Zero Winding (direction-based filling)"
}

// ResetStyle resets all style settings to their default values. Mirrors
// utilities.go.
func (a *Agg2DFloat) ResetStyle() {
	a.fillColor = White
	a.fillGradientFlag = Solid
	a.lineColor = Black
	a.lineGradientFlag = Solid
	a.lineWidth = 1.0
	a.lineCap = CapRound
	a.lineJoin = JoinRound
	a.masterAlpha = 1.0
	a.evenOddFlag = false
	if a.convStroke != nil {
		a.convStroke.SetWidth(1.0)
		a.convStroke.SetLineCap(basicsLineCap(CapRound))
		a.convStroke.SetLineJoin(basicsLineJoin(JoinRound))
	}
	a.NoDashes()
	a.textAlignX = AlignLeft
	a.textAlignY = AlignBottom
}

// ClearClipBox clears the current clipping box with the specified color.
// Mirrors utilities.go.
func (a *Agg2DFloat) ClearClipBox(c Color) {
	a.ClearClipBoxRGBA(c[0], c[1], c[2], c[3])
}

// ClearClipBoxRGBA clears the current clipping box with the specified RGBA
// values. Mirrors utilities.go, building the float color via colorToRGBA32.
func (a *Agg2DFloat) ClearClipBoxRGBA(r, g, b, alpha uint8) {
	if a.renBase == nil {
		return
	}
	clearColor := colorToRGBA32(Color{r, g, b, alpha})
	a.renBase.rendererBase().CopyBar(0, 0, a.renBase.Width(), a.renBase.Height(), clearColor)
}
