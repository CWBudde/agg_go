package blender

import (
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces (RGB 32-bit)
////////////////////////////////////////////////////////////////////////////////

type RGB96Blender[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []float32, r, g, b, a, cover float32)
}

////////////////////////////////////////////////////////////////////////////////
// RGB96 (32-bit float per channel)
////////////////////////////////////////////////////////////////////////////////

type BlenderRGB96[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGB96[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	alpha := a * cover
	if alpha <= 0 {
		return
	}
	var o O
	inv := float32(1.0) - alpha
	dst[o.IdxR()] = dst[o.IdxR()]*inv + r*alpha
	dst[o.IdxG()] = dst[o.IdxG()]*inv + g*alpha
	dst[o.IdxB()] = dst[o.IdxB()]*inv + b*alpha
}

type BlenderRGB96Pre[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGB96Pre[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	// Scale premultiplied source by coverage
	cr := r * cover
	cg := g * cover
	cb := b * cover
	ca := a * cover
	if ca <= 0 && cr == 0 && cg == 0 && cb == 0 {
		return
	}
	var o O
	inv := float32(1.0) - ca
	dst[o.IdxR()] = dst[o.IdxR()]*inv + cr
	dst[o.IdxG()] = dst[o.IdxG()]*inv + cg
	dst[o.IdxB()] = dst[o.IdxB()]*inv + cb
}

type Gamma32Corrector interface {
	Dir(v float32) float32
	Inv(v float32) float32
}

type BlenderRGB96Gamma[S color.Space, O order.RGBOrder, G Gamma32Corrector] struct {
	gamma G
}

func NewBlenderRGB96Gamma[S color.Space, O order.RGBOrder, G Gamma32Corrector](g G) BlenderRGB96Gamma[S, O, G] {
	return BlenderRGB96Gamma[S, O, G]{gamma: g}
}

func (bl BlenderRGB96Gamma[S, O, G]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	alpha := a * cover
	if alpha <= 0 {
		return
	}
	var o O
	dr := bl.gamma.Dir(dst[o.IdxR()])
	dg := bl.gamma.Dir(dst[o.IdxG()])
	db := bl.gamma.Dir(dst[o.IdxB()])

	sr := bl.gamma.Dir(r)
	sg := bl.gamma.Dir(g)
	sb := bl.gamma.Dir(b)

	inv := float32(1.0) - alpha
	dst[o.IdxR()] = bl.gamma.Inv(dr*inv + sr*alpha)
	dst[o.IdxG()] = bl.gamma.Inv(dg*inv + sg*alpha)
	dst[o.IdxB()] = bl.gamma.Inv(db*inv + sb*alpha)
}

func (BlenderRGB96[S, O]) SetPlain(dst []float32, r, g, b, a float32) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (BlenderRGB96[S, O]) GetPlain(src []float32) (r, g, b, a float32) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], 1.0
}

func (BlenderRGB96Pre[S, O]) SetPlain(dst []float32, r, g, b, a float32) {
	var o O
	dst[o.IdxR()] = r * a
	dst[o.IdxG()] = g * a
	dst[o.IdxB()] = b * a
}

func (BlenderRGB96Pre[S, O]) GetPlain(src []float32) (r, g, b, a float32) {
	var o O
	// RGB has no alpha stored, so we can't demultiply - just return the stored values
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], 1.0
}

func (bl BlenderRGB96Gamma[S, O, G]) SetPlain(dst []float32, r, g, b, a float32) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (bl BlenderRGB96Gamma[S, O, G]) GetPlain(src []float32) (r, g, b, a float32) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], 1.0
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases for 32-bit float RGB
////////////////////////////////////////////////////////////////////////////////

// 32-bit float RGB96 format aliases (explicit naming)
type (
	BlenderRGB96LinearRGB    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGB96LinearBGR    = BlenderRGB96[color.Linear, order.BGR]
	BlenderRGB96SRGBRGB      = BlenderRGB96[color.SRGB, order.RGB]
	BlenderRGB96SRGBBGR      = BlenderRGB96[color.SRGB, order.BGR]
	BlenderRGB96PreLinearRGB = BlenderRGB96Pre[color.Linear, order.RGB]
	BlenderRGB96PreLinearBGR = BlenderRGB96Pre[color.Linear, order.BGR]
	BlenderRGB96PreSRGBRGB   = BlenderRGB96Pre[color.SRGB, order.RGB]
	BlenderRGB96PreSRGBBGR   = BlenderRGB96Pre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Generic RGB32 aliases (more intuitive naming)
////////////////////////////////////////////////////////////////////////////////

type (
	BlenderRGB32LinearRGB    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGB32LinearBGR    = BlenderRGB96[color.Linear, order.BGR]
	BlenderRGB32SRGBRGB      = BlenderRGB96[color.SRGB, order.RGB]
	BlenderRGB32SRGBBGR      = BlenderRGB96[color.SRGB, order.BGR]
	BlenderRGB32PreLinearRGB = BlenderRGB96Pre[color.Linear, order.RGB]
	BlenderRGB32PreLinearBGR = BlenderRGB96Pre[color.Linear, order.BGR]
	BlenderRGB32PreSRGBRGB   = BlenderRGB96Pre[color.SRGB, order.RGB]
	BlenderRGB32PreSRGBBGR   = BlenderRGB96Pre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Float-specific aliases for high-precision rendering
////////////////////////////////////////////////////////////////////////////////

type (
	// Standard float RGB (most common high-precision format)
	BlenderRGBFloat    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGBFloatPre = BlenderRGB96Pre[color.Linear, order.RGB]
	BlenderBGRFloat    = BlenderRGB96[color.Linear, order.BGR]
	BlenderBGRFloatPre = BlenderRGB96Pre[color.Linear, order.BGR]

	// sRGB float variants (less common but sometimes needed)
	BlenderRGBFloatSRGB    = BlenderRGB96[color.SRGB, order.RGB]
	BlenderRGBFloatPreSRGB = BlenderRGB96Pre[color.SRGB, order.RGB]
)

////////////////////////////////////////////////////////////////////////////////
// Ultra-short aliases for common usage
////////////////////////////////////////////////////////////////////////////////

type (
	// Short aliases for the most common cases
	RGB32Blender = BlenderRGB96[color.Linear, order.RGB]
	BGR32Blender = BlenderRGB96[color.Linear, order.BGR]

	RGB32PreBlender = BlenderRGB96Pre[color.Linear, order.RGB]
	BGR32PreBlender = BlenderRGB96Pre[color.Linear, order.BGR]

	// Float-specific short names
	RGBFloatBlender = BlenderRGB96[color.Linear, order.RGB]
	BGRFloatBlender = BlenderRGB96[color.Linear, order.BGR]

	RGBFloatPreBlender = BlenderRGB96Pre[color.Linear, order.RGB]
	BGRFloatPreBlender = BlenderRGB96Pre[color.Linear, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Professional and HDR aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// HDR/high dynamic range blending
	BlenderRGBHDR    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGBPreHDR = BlenderRGB96Pre[color.Linear, order.RGB]

	// Professional imaging (always linear for accuracy)
	BlenderRGBPro    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGBPrePro = BlenderRGB96Pre[color.Linear, order.RGB]

	// Scientific/engineering precision
	BlenderRGBPrecision    = BlenderRGB96[color.Linear, order.RGB]
	BlenderRGBPrePrecision = BlenderRGB96Pre[color.Linear, order.RGB]
)

////////////////////////////////////////////////////////////////////////////////
// Simple order-specific aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Simple RGB/BGR distinction with generic space parameter
	BlenderRGB32Simple[S color.Space] = BlenderRGB96[S, order.RGB]
	BlenderBGR32Simple[S color.Space] = BlenderRGB96[S, order.BGR]

	BlenderRGB32PreSimple[S color.Space] = BlenderRGB96Pre[S, order.RGB]
	BlenderBGR32PreSimple[S color.Space] = BlenderRGB96Pre[S, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Specific space and order aliases
////////////////////////////////////////////////////////////////////////////////

type (
	// Standard linear space aliases
	BlenderRGB96Linear = BlenderRGB96[color.Linear, order.RGB]
	BlenderBGR96Linear = BlenderRGB96[color.Linear, order.BGR]

	// SRGB space aliases
	BlenderRGB96SRGB = BlenderRGB96[color.SRGB, order.RGB]
	BlenderBGR96SRGB = BlenderRGB96[color.SRGB, order.BGR]

	// Premultiplied linear space aliases
	BlenderRGB96PreLinear = BlenderRGB96Pre[color.Linear, order.RGB]
	BlenderBGR96PreLinear = BlenderRGB96Pre[color.Linear, order.BGR]

	// Premultiplied SRGB space aliases
	BlenderRGB96PreSRGB = BlenderRGB96Pre[color.SRGB, order.RGB]
	BlenderBGR96PreSRGB = BlenderRGB96Pre[color.SRGB, order.BGR]
)
