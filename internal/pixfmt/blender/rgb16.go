package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces (RGB 16-bit)
////////////////////////////////////////////////////////////////////////////////

type RGB48Blender[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []basics.Int16u, r, g, b, a, cover basics.Int16u)
}

// //////////////////////////////////////////////////////////////////////////////
// RGB48 (16-bit per channel)
// //////////////////////////////////////////////////////////////////////////////
type BlenderRGB48[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGB48[S, O]) BlendPix(dst []basics.Int16u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGB16MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGB16Lerp(dst[o.IdxR()], r, alpha)
	dst[o.IdxG()] = color.RGB16Lerp(dst[o.IdxG()], g, alpha)
	dst[o.IdxB()] = color.RGB16Lerp(dst[o.IdxB()], b, alpha)
}

type BlenderRGB48Pre[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGB48Pre[S, O]) BlendPix(dst []basics.Int16u, r, g, b, a, cover basics.Int16u) {
	if cover != 65535 {
		r = color.RGB16MultCover(r, cover)
		g = color.RGB16MultCover(g, cover)
		b = color.RGB16MultCover(b, cover)
		a = color.RGB16MultCover(a, cover)
	}
	if a == 0 && r == 0 && g == 0 && b == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGB16Prelerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = color.RGB16Prelerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = color.RGB16Prelerp(dst[o.IdxB()], b, a)
}

type Gamma16Corrector interface {
	Dir(v basics.Int16u) basics.Int16u
	Inv(v basics.Int16u) basics.Int16u
}

type BlenderRGB48Gamma[S color.Space, O order.RGBOrder, G Gamma16Corrector] struct {
	gamma G
}

func NewBlenderRGB48Gamma[S color.Space, O order.RGBOrder, G Gamma16Corrector](g G) BlenderRGB48Gamma[S, O, G] {
	return BlenderRGB48Gamma[S, O, G]{gamma: g}
}

func (bl BlenderRGB48Gamma[S, O, G]) BlendPix(dst []basics.Int16u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGB16MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O
	dr := bl.gamma.Dir(dst[o.IdxR()])
	dg := bl.gamma.Dir(dst[o.IdxG()])
	db := bl.gamma.Dir(dst[o.IdxB()])

	sr := bl.gamma.Dir(r)
	sg := bl.gamma.Dir(g)
	sb := bl.gamma.Dir(b)

	dst[o.IdxR()] = bl.gamma.Inv(color.RGB16Lerp(dr, sr, alpha))
	dst[o.IdxG()] = bl.gamma.Inv(color.RGB16Lerp(dg, sg, alpha))
	dst[o.IdxB()] = bl.gamma.Inv(color.RGB16Lerp(db, sb, alpha))
}

func (BlenderRGB48[S, O]) SetPlain(dst []basics.Int16u, r, g, b, a basics.Int16u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (BlenderRGB48[S, O]) GetPlain(src []basics.Int16u) (r, g, b, a basics.Int16u) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], 0xFFFF
}

func (BlenderRGB48Pre[S, O]) SetPlain(dst []basics.Int16u, r, g, b, a basics.Int16u) {
	var o O
	dst[o.IdxR()] = color.RGBA16Multiply(r, a)
	dst[o.IdxG()] = color.RGBA16Multiply(g, a)
	dst[o.IdxB()] = color.RGBA16Multiply(b, a)
}

func (BlenderRGB48Pre[S, O]) GetPlain(src []basics.Int16u) (r, g, b, a basics.Int16u) {
	var o O
	a = 0xFFFF
	if a == 0 {
		return 0, 0, 0, 0
	}
	// RGB has no alpha stored, so we can't demultiply - just return the stored values
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], a
}

func (bl BlenderRGB48Gamma[S, O, G]) SetPlain(dst []basics.Int16u, r, g, b, a basics.Int16u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (bl BlenderRGB48Gamma[S, O, G]) GetPlain(src []basics.Int16u) (r, g, b, a basics.Int16u) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], 0xFFFF
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (Linear / sRGB × RGB / BGR)
////////////////////////////////////////////////////////////////////////////////

// 16-bit RGB48 format aliases (explicit naming)
type (
	BlenderRGB48LinearRGB    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB48LinearBGR    = BlenderRGB48[color.Linear, order.BGR]
	BlenderRGB48SRGBRGB      = BlenderRGB48[color.SRGB, order.RGB]
	BlenderRGB48SRGBBGR      = BlenderRGB48[color.SRGB, order.BGR]
	BlenderRGB48PreLinearRGB = BlenderRGB48Pre[color.Linear, order.RGB]
	BlenderRGB48PreLinearBGR = BlenderRGB48Pre[color.Linear, order.BGR]
	BlenderRGB48PreSRGBRGB   = BlenderRGB48Pre[color.SRGB, order.RGB]
	BlenderRGB48PreSRGBBGR   = BlenderRGB48Pre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Generic RGB16 aliases (more flexible)
////////////////////////////////////////////////////////////////////////////////

type (
	BlenderRGB16LinearRGB    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB16LinearBGR    = BlenderRGB48[color.Linear, order.BGR]
	BlenderRGB16SRGBRGB      = BlenderRGB48[color.SRGB, order.RGB]
	BlenderRGB16SRGBBGR      = BlenderRGB48[color.SRGB, order.BGR]
	BlenderRGB16PreLinearRGB = BlenderRGB48Pre[color.Linear, order.RGB]
	BlenderRGB16PreLinearBGR = BlenderRGB48Pre[color.Linear, order.BGR]
	BlenderRGB16PreSRGBRGB   = BlenderRGB48Pre[color.SRGB, order.RGB]
	BlenderRGB16PreSRGBBGR   = BlenderRGB48Pre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Platform-specific RGB16 aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Standard RGB16 (most common high-precision format)
	BlenderRGB16Standard    = BlenderRGB48[color.SRGB, order.RGB]
	BlenderRGB16PreStandard = BlenderRGB48Pre[color.SRGB, order.RGB]

	// High-precision BGR (less common but useful for some platforms)
	BlenderBGR16Standard    = BlenderRGB48[color.SRGB, order.BGR]
	BlenderBGR16PreStandard = BlenderRGB48Pre[color.SRGB, order.BGR]

	// Linear space variants for high-quality rendering
	BlenderRGB16Linear    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB16PreLinear = BlenderRGB48Pre[color.Linear, order.RGB]
	BlenderBGR16Linear    = BlenderRGB48[color.Linear, order.BGR]
	BlenderBGR16PreLinear = BlenderRGB48Pre[color.Linear, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Ultra-short aliases for common usage
////////////////////////////////////////////////////////////////////////////////

type (
	// Short aliases for the most common cases
	RGB16Blender = BlenderRGB48[color.SRGB, order.RGB]
	BGR16Blender = BlenderRGB48[color.SRGB, order.BGR]

	RGB16PreBlender = BlenderRGB48Pre[color.SRGB, order.RGB]
	BGR16PreBlender = BlenderRGB48Pre[color.SRGB, order.BGR]

	// Alternative naming (RGB48 for 6-byte format)
	RGB48BlenderStd = BlenderRGB48[color.SRGB, order.RGB]
	BGR48BlenderStd = BlenderRGB48[color.SRGB, order.BGR]

	RGB48PreBlenderStd = BlenderRGB48Pre[color.SRGB, order.RGB]
	BGR48PreBlenderStd = BlenderRGB48Pre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// High-precision and gamma-corrected aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// High-precision blending (emphasizes 16-bit precision)
	BlenderRGB16HighPrecision    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB16PreHighPrecision = BlenderRGB48Pre[color.Linear, order.RGB]

	// HDR/wide-gamut aliases
	BlenderRGB16HDR    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB16PreHDR = BlenderRGB48Pre[color.Linear, order.RGB]

	// Professional imaging aliases
	BlenderRGB16Pro    = BlenderRGB48[color.Linear, order.RGB]
	BlenderRGB16PrePro = BlenderRGB48Pre[color.Linear, order.RGB]
)

////////////////////////////////////////////////////////////////////////////////
// Simple order-specific aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Simple RGB/BGR distinction with generic space parameter
	BlenderRGB16Simple[S color.Space] = BlenderRGB48[S, order.RGB]
	BlenderBGR16Simple[S color.Space] = BlenderRGB48[S, order.BGR]

	BlenderRGB16PreSimple[S color.Space] = BlenderRGB48Pre[S, order.RGB]
	BlenderBGR16PreSimple[S color.Space] = BlenderRGB48Pre[S, order.BGR]
)
