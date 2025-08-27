package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// RGB48 (16-bit per channel)
////////////////////////////////////////////////////////////////////////////////

type RGB48Blender[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []basics.Int16u, r, g, b, a, cover basics.Int16u)
}

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

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (Linear / sRGB × RGB / BGR)
////////////////////////////////////////////////////////////////////////////////

// 16-bit
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
