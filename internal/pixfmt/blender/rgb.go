package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces (mirroring the RGBA pattern)
////////////////////////////////////////////////////////////////////////////////

type RGBBlender[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u)
}

type RGBBlenderSimple[S color.Space, O order.RGBOrder] interface {
	BlendPix(dst []basics.Int8u, r, g, b, a basics.Int8u)
}

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> RGB destination (no alpha stored)
////////////////////////////////////////////////////////////////////////////////

type BlenderRGB[S color.Space, O order.RGBOrder] struct{}

// Lerp by alpha*cover; destination stores only RGB (3 bytes).
func (BlenderRGB[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Lerp(dst[o.IdxR()], r, alpha)
	dst[o.IdxG()] = color.RGBA8Lerp(dst[o.IdxG()], g, alpha)
	dst[o.IdxB()] = color.RGBA8Lerp(dst[o.IdxB()], b, alpha)
}

func (BlenderRGB[S, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	if a == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Lerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = color.RGBA8Lerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = color.RGBA8Lerp(dst[o.IdxB()], b, a)
}

////////////////////////////////////////////////////////////////////////////////
/* Premultiplied source -> RGB destination (no alpha stored)

   Matches the RGBA "pre" semantics: channels use prelerp with an
   effective premultiplied coverage (scale r,g,b,a by cover first).
*/
////////////////////////////////////////////////////////////////////////////////

type BlenderRGBPre[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGBPre[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	if cover != 255 {
		r = color.RGBA8MultCover(r, cover)
		g = color.RGBA8MultCover(g, cover)
		b = color.RGBA8MultCover(b, cover)
		a = color.RGBA8MultCover(a, cover)
	}
	if a == 0 && r == 0 && g == 0 && b == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Prelerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = color.RGBA8Prelerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = color.RGBA8Prelerp(dst[o.IdxB()], b, a)
}

func (BlenderRGBPre[S, O]) BlendPixSimple(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	if a == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Prelerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = color.RGBA8Prelerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = color.RGBA8Prelerp(dst[o.IdxB()], b, a)
}

////////////////////////////////////////////////////////////////////////////////
// Gamma-corrected 8-bit RGB (no alpha stored)
////////////////////////////////////////////////////////////////////////////////

type GammaCorrector interface {
	Dir(v basics.Int8u) basics.Int8u // forward gamma
	Inv(v basics.Int8u) basics.Int8u // inverse gamma
}

type BlenderRGBGamma[S color.Space, O order.RGBOrder, G GammaCorrector] struct {
	gamma G
}

func NewBlenderRGBGamma[S color.Space, O order.RGBOrder, G GammaCorrector](g G) BlenderRGBGamma[S, O, G] {
	return BlenderRGBGamma[S, O, G]{gamma: g}
}

func (bl BlenderRGBGamma[S, O, G]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
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

	dst[o.IdxR()] = bl.gamma.Inv(color.RGBA8Lerp(dr, sr, alpha))
	dst[o.IdxG()] = bl.gamma.Inv(color.RGBA8Lerp(dg, sg, alpha))
	dst[o.IdxB()] = bl.gamma.Inv(color.RGBA8Lerp(db, sb, alpha))
}

////////////////////////////////////////////////////////////////////////////////
// Helpers for 8-bit RGB
////////////////////////////////////////////////////////////////////////////////

func BlendRGBPixel[B RGBBlender[S, O], S color.Space, O order.RGBOrder](
	dst []basics.Int8u,
	src color.RGB8[S],
	alpha, cover basics.Int8u,
	bl B,
) {
	if cover == 0 || alpha == 0 {
		return
	}
	bl.BlendPix(dst, src.R, src.G, src.B, alpha, cover)
}

func CopyRGBPixel[S color.Space, O order.RGBOrder](
	dst []basics.Int8u,
	src color.RGB8[S],
) {
	var o O
	dst[o.IdxR()] = src.R
	dst[o.IdxG()] = src.G
	dst[o.IdxB()] = src.B
}

func BlendRGBHline[B RGBBlender[S, O], S color.Space, O order.RGBOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGB8[S],
	alpha basics.Int8u,
	covers []basics.Int8u, // nil => full coverage
	bl B,
) {
	if length <= 0 || alpha == 0 {
		return
	}
	const pixStep = 3
	p := x * pixStep

	if covers == nil {
		for i := 0; i < length; i++ {
			bl.BlendPix(dst[p:p+3], src.R, src.G, src.B, alpha, 255)
			p += pixStep
		}
		return
	}
	for i := 0; i < length; i++ {
		if c := covers[i]; c != 0 {
			bl.BlendPix(dst[p:p+3], src.R, src.G, src.B, alpha, c)
		}
		p += pixStep
	}
}

func CopyRGBHline[S color.Space, O order.RGBOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGB8[S],
) {
	if length <= 0 {
		return
	}
	const pixStep = 3
	var o O
	p := x * pixStep
	for i := 0; i < length; i++ {
		dst[p+o.IdxR()] = src.R
		dst[p+o.IdxG()] = src.G
		dst[p+o.IdxB()] = src.B
		p += pixStep
	}
}

func FillRGBSpan[S color.Space, O order.RGBOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGB8[S],
) {
	CopyRGBHline[S, O](dst, x, length, src)
}

func ConvertRGBAToRGB[S color.Space](rgba color.RGBA8[S]) color.RGB8[S] {
	return color.RGB8[S]{R: rgba.R, G: rgba.G, B: rgba.B}
}

func ConvertRGBToRGBA[S color.Space](rgb color.RGB8[S]) color.RGBA8[S] {
	return color.RGBA8[S]{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (Linear / sRGB × RGB / BGR)
////////////////////////////////////////////////////////////////////////////////

// 8-bit
type (
	BlenderRGB24LinearRGB = BlenderRGB[color.Linear, order.RGB]
	BlenderRGB24LinearBGR = BlenderRGB[color.Linear, order.BGR]
	BlenderRGB24SRGBRGB   = BlenderRGB[color.SRGB, order.RGB]
	BlenderRGB24SRGBBGR   = BlenderRGB[color.SRGB, order.BGR]

	BlenderRGB24PreLinearRGB = BlenderRGBPre[color.Linear, order.RGB]
	BlenderRGB24PreLinearBGR = BlenderRGBPre[color.Linear, order.BGR]
	BlenderRGB24PreSRGBRGB   = BlenderRGBPre[color.SRGB, order.RGB]
	BlenderRGB24PreSRGBBGR   = BlenderRGBPre[color.SRGB, order.BGR]
)
