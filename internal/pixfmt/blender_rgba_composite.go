package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"math"
)

// Composite blend operation types
type CompOp int

const (
	CompOpClear CompOp = iota
	CompOpSrc
	CompOpDst
	CompOpSrcOver
	CompOpDstOver
	CompOpSrcIn
	CompOpDstIn
	CompOpSrcOut
	CompOpDstOut
	CompOpSrcAtop
	CompOpDstAtop
	CompOpXor
	CompOpPlus
	CompOpMultiply
	CompOpScreen
	CompOpOverlay
	CompOpDarken
	CompOpLighten
	CompOpColorDodge
	CompOpColorBurn
	CompOpHardLight
	CompOpSoftLight
	CompOpDifference
	CompOpExclusion
)

// CompositeBlender implements composite blending operations for RGBA
type CompositeBlender[CS any, O any] struct {
	op CompOp
}

// NewCompositeBlender creates a new composite blender with the specified operation
func NewCompositeBlender[CS any, O any](op CompOp) CompositeBlender[CS, O] {
	return CompositeBlender[CS, O]{op: op}
}

// BlendPix blends a pixel using the specified composite operation
func (bl CompositeBlender[CS, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha == 0 {
		return
	}

	order := getColorOrder[O]()

	// Convert to normalized floating point for calculations
	s := normalizedRGBA{
		r: float64(r) / 255.0,
		g: float64(g) / 255.0,
		b: float64(b) / 255.0,
		a: float64(alpha) / 255.0,
	}

	d := normalizedRGBA{
		r: float64(dst[order.R]) / 255.0,
		g: float64(dst[order.G]) / 255.0,
		b: float64(dst[order.B]) / 255.0,
		a: float64(dst[order.A]) / 255.0,
	}

	result := bl.blendOperation(d, s)

	// Clamp and convert back to 8-bit
	dst[order.R] = basics.Int8u(clamp(result.r * 255.0))
	dst[order.G] = basics.Int8u(clamp(result.g * 255.0))
	dst[order.B] = basics.Int8u(clamp(result.b * 255.0))
	dst[order.A] = basics.Int8u(clamp(result.a * 255.0))
}

// normalizedRGBA represents RGBA values in normalized floating point [0, 1]
type normalizedRGBA struct {
	r, g, b, a float64
}

// blendOperation performs the actual blending calculation
func (bl CompositeBlender[CS, O]) blendOperation(d, s normalizedRGBA) normalizedRGBA {
	switch bl.op {
	case CompOpMultiply:
		return bl.multiply(d, s)
	case CompOpScreen:
		return bl.screen(d, s)
	case CompOpOverlay:
		return bl.overlay(d, s)
	case CompOpDarken:
		return bl.darken(d, s)
	case CompOpLighten:
		return bl.lighten(d, s)
	case CompOpColorDodge:
		return bl.colorDodge(d, s)
	case CompOpColorBurn:
		return bl.colorBurn(d, s)
	case CompOpHardLight:
		return bl.hardLight(d, s)
	case CompOpSoftLight:
		return bl.softLight(d, s)
	case CompOpDifference:
		return bl.difference(d, s)
	case CompOpExclusion:
		return bl.exclusion(d, s)
	case CompOpPlus:
		return bl.plus(d, s)
	default:
		// Default to source-over blending
		return bl.sourceOver(d, s)
	}
}

// multiply blend mode: Dca' = Sca.Dca + Sca.(1 - Da) + Dca.(1 - Sa)
func (bl CompositeBlender[CS, O]) multiply(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	s1a := 1.0 - s.a
	d1a := 1.0 - d.a

	return normalizedRGBA{
		r: s.r*d.r + s.r*d1a + d.r*s1a,
		g: s.g*d.g + s.g*d1a + d.g*s1a,
		b: s.b*d.b + s.b*d1a + d.b*s1a,
		a: d.a + s.a - s.a*d.a,
	}
}

// screen blend mode: Dca' = Sca + Dca - Sca.Dca
func (bl CompositeBlender[CS, O]) screen(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	return normalizedRGBA{
		r: d.r + s.r - s.r*d.r,
		g: d.g + s.g - s.g*d.g,
		b: d.b + s.b - s.b*d.b,
		a: d.a + s.a - s.a*d.a,
	}
}

// overlay blend mode
func (bl CompositeBlender[CS, O]) overlay(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a
	sada := s.a * d.a

	calcOverlay := func(dca, sca, da, sa, sada, d1a, s1a float64) float64 {
		if 2*dca <= da {
			return 2*sca*dca + sca*d1a + dca*s1a
		}
		return sada - 2*(da-dca)*(sa-sca) + sca*d1a + dca*s1a
	}

	return normalizedRGBA{
		r: calcOverlay(d.r, s.r, d.a, s.a, sada, d1a, s1a),
		g: calcOverlay(d.g, s.g, d.a, s.a, sada, d1a, s1a),
		b: calcOverlay(d.b, s.b, d.a, s.a, sada, d1a, s1a),
		a: d.a + s.a - s.a*d.a,
	}
}

// darken blend mode: Dca' = min(Sca.Da, Dca.Sa) + Sca.(1 - Da) + Dca.(1 - Sa)
func (bl CompositeBlender[CS, O]) darken(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a

	return normalizedRGBA{
		r: math.Min(s.r*d.a, d.r*s.a) + s.r*d1a + d.r*s1a,
		g: math.Min(s.g*d.a, d.g*s.a) + s.g*d1a + d.g*s1a,
		b: math.Min(s.b*d.a, d.b*s.a) + s.b*d1a + d.b*s1a,
		a: d.a + s.a - s.a*d.a,
	}
}

// lighten blend mode: Dca' = max(Sca.Da, Dca.Sa) + Sca.(1 - Da) + Dca.(1 - Sa)
func (bl CompositeBlender[CS, O]) lighten(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a

	return normalizedRGBA{
		r: math.Max(s.r*d.a, d.r*s.a) + s.r*d1a + d.r*s1a,
		g: math.Max(s.g*d.a, d.g*s.a) + s.g*d1a + d.g*s1a,
		b: math.Max(s.b*d.a, d.b*s.a) + s.b*d1a + d.b*s1a,
		a: d.a + s.a - s.a*d.a,
	}
}

// colorDodge blend mode
func (bl CompositeBlender[CS, O]) colorDodge(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	if d.a <= 0 {
		return normalizedRGBA{
			r: s.r * (1.0 - d.a),
			g: s.g * (1.0 - d.a),
			b: s.b * (1.0 - d.a),
			a: s.a,
		}
	}

	sada := s.a * d.a
	s1a := 1.0 - s.a
	d1a := 1.0 - d.a

	calcDodge := func(dca, sca, da, sa, sada, d1a, s1a float64) float64 {
		if sca < sa {
			return sada*math.Min(1.0, (dca/da)*sa/(sa-sca)) + sca*d1a + dca*s1a
		}
		if dca > 0 {
			return sada + sca*d1a + dca*s1a
		}
		return sca * d1a
	}

	return normalizedRGBA{
		r: calcDodge(d.r, s.r, d.a, s.a, sada, d1a, s1a),
		g: calcDodge(d.g, s.g, d.a, s.a, sada, d1a, s1a),
		b: calcDodge(d.b, s.b, d.a, s.a, sada, d1a, s1a),
		a: d.a + s.a - s.a*d.a,
	}
}

// colorBurn blend mode
func (bl CompositeBlender[CS, O]) colorBurn(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	if d.a <= 0 {
		return normalizedRGBA{
			r: s.r * (1.0 - d.a),
			g: s.g * (1.0 - d.a),
			b: s.b * (1.0 - d.a),
			a: s.a,
		}
	}

	sada := s.a * d.a
	s1a := 1.0 - s.a
	d1a := 1.0 - d.a

	calcBurn := func(dca, sca, da, sa, sada, d1a, s1a float64) float64 {
		if sca > 0 {
			return sada*(1.0-math.Min(1.0, (1.0-dca/da)*sa/sca)) + sca*d1a + dca*s1a
		}
		if dca < da {
			return dca * s1a
		}
		return sada + sca*d1a + dca*s1a
	}

	return normalizedRGBA{
		r: calcBurn(d.r, s.r, d.a, s.a, sada, d1a, s1a),
		g: calcBurn(d.g, s.g, d.a, s.a, sada, d1a, s1a),
		b: calcBurn(d.b, s.b, d.a, s.a, sada, d1a, s1a),
		a: d.a + s.a - s.a*d.a,
	}
}

// hardLight blend mode (overlay with source and destination swapped)
func (bl CompositeBlender[CS, O]) hardLight(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a
	sada := s.a * d.a

	calcHardLight := func(dca, sca, da, sa, sada, d1a, s1a float64) float64 {
		if 2*sca <= sa {
			return 2*sca*dca + sca*d1a + dca*s1a
		}
		return sada - 2*(da-dca)*(sa-sca) + sca*d1a + dca*s1a
	}

	return normalizedRGBA{
		r: calcHardLight(d.r, s.r, d.a, s.a, sada, d1a, s1a),
		g: calcHardLight(d.g, s.g, d.a, s.a, sada, d1a, s1a),
		b: calcHardLight(d.b, s.b, d.a, s.a, sada, d1a, s1a),
		a: d.a + s.a - s.a*d.a,
	}
}

// softLight blend mode
func (bl CompositeBlender[CS, O]) softLight(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a
	sada := s.a * d.a

	calcSoftLight := func(dca, sca, da, sa, sada, d1a, s1a float64) float64 {
		if 2*sca <= sa {
			return dca*(sa+(2*sca-sa)*(1-dca/da)) + sca*d1a + dca*s1a
		}
		if 4*dca <= da {
			return sada*((2*sca-sa)*((16*dca/da-12)*dca/da+3)) + sca*d1a + dca*s1a
		}
		return sada*((2*sca-sa)*(math.Sqrt(dca/da)-dca/da)) + sca*d1a + dca*s1a
	}

	return normalizedRGBA{
		r: calcSoftLight(d.r, s.r, d.a, s.a, sada, d1a, s1a),
		g: calcSoftLight(d.g, s.g, d.a, s.a, sada, d1a, s1a),
		b: calcSoftLight(d.b, s.b, d.a, s.a, sada, d1a, s1a),
		a: d.a + s.a - s.a*d.a,
	}
}

// difference blend mode: Dca' = |Dca.Sa - Sca.Da| + Sca.(1 - Da) + Dca.(1 - Sa)
func (bl CompositeBlender[CS, O]) difference(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a

	return normalizedRGBA{
		r: math.Abs(d.r*s.a-s.r*d.a) + s.r*d1a + d.r*s1a,
		g: math.Abs(d.g*s.a-s.g*d.a) + s.g*d1a + d.g*s1a,
		b: math.Abs(d.b*s.a-s.b*d.a) + s.b*d1a + d.b*s1a,
		a: d.a + s.a - s.a*d.a,
	}
}

// exclusion blend mode: Dca' = (Sca.Da + Dca.Sa - 2.Sca.Dca) + Sca.(1 - Da) + Dca.(1 - Sa)
func (bl CompositeBlender[CS, O]) exclusion(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	d1a := 1.0 - d.a
	s1a := 1.0 - s.a

	return normalizedRGBA{
		r: s.r*d.a + d.r*s.a - 2*s.r*d.r + s.r*d1a + d.r*s1a,
		g: s.g*d.a + d.g*s.a - 2*s.g*d.g + s.g*d1a + d.g*s1a,
		b: s.b*d.a + d.b*s.a - 2*s.b*d.b + s.b*d1a + d.b*s1a,
		a: d.a + s.a - s.a*d.a,
	}
}

// plus blend mode: Dca' = Sca + Dca
func (bl CompositeBlender[CS, O]) plus(d, s normalizedRGBA) normalizedRGBA {
	return normalizedRGBA{
		r: d.r + s.r,
		g: d.g + s.g,
		b: d.b + s.b,
		a: d.a + s.a - s.a*d.a,
	}
}

// sourceOver blend mode (standard alpha blending)
func (bl CompositeBlender[CS, O]) sourceOver(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}

	invSrcAlpha := 1.0 - s.a
	return normalizedRGBA{
		r: s.r + d.r*invSrcAlpha,
		g: s.g + d.g*invSrcAlpha,
		b: s.b + d.b*invSrcAlpha,
		a: s.a + d.a*invSrcAlpha,
	}
}

// clamp clamps a value to [0, 255]
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// Concrete composite blender types for convenience
type (
	CompositeBlenderRGBA8Multiply   = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Screen     = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Overlay    = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Darken     = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Lighten    = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8ColorDodge = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8ColorBurn  = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8HardLight  = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8SoftLight  = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Difference = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Exclusion  = CompositeBlender[color.Linear, RGBAOrder]
	CompositeBlenderRGBA8Plus       = CompositeBlender[color.Linear, RGBAOrder]
)

// Helper functions to create specific composite blenders
func NewMultiplyBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpMultiply)
}

func NewScreenBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpScreen)
}

func NewOverlayBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpOverlay)
}

func NewDarkenBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpDarken)
}

func NewLightenBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpLighten)
}

func NewColorDodgeBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpColorDodge)
}

func NewColorBurnBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpColorBurn)
}

func NewHardLightBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpHardLight)
}

func NewSoftLightBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpSoftLight)
}

func NewDifferenceBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpDifference)
}

func NewExclusionBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpExclusion)
}

func NewPlusBlender[CS any, O any]() CompositeBlender[CS, O] {
	return NewCompositeBlender[CS, O](CompOpPlus)
}
