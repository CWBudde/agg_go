package blender

import (
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// RGB96 (32-bit float per channel)
////////////////////////////////////////////////////////////////////////////////

type RGB96Blender[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []float32, r, g, b, a, cover float32)
}

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

// 32-bit float
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
