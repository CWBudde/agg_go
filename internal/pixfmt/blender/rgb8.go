package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// RGB Blender interface
////////////////////////////////////////////////////////////////////////////////

// RGBBlender defines the minimal interface used by RGB pixfmt implementations.
// The blender handles color space interpretation and internal pixel ordering.
type RGBBlender[S color.Space] interface {
	// GetPlain reads a pixel and returns plain RGB components
	// interpreted according to color space S
	GetPlain(px []byte) (r, g, b basics.Int8u)

	// SetPlain writes plain RGB components to a pixel, mapping them to the
	// internal order of the blender
	SetPlain(px []byte, r, g, b basics.Int8u)

	// BlendPix blends plain RGB source into the pixel with given alpha and coverage
	// r,g,b are interpreted according to S, and mapped to the order internal to the blender
	BlendPix(px []byte, r, g, b, a, cover basics.Int8u)
}

// RawRGBOrder provides optional fast path for zero-cost index access for RGB.
type RawRGBOrder interface {
	IdxR() int
	IdxG() int
	IdxB() int
}

////////////////////////////////////////////////////////////////////////////////
// RGBX Blender interface for 4-byte RGB with padding
////////////////////////////////////////////////////////////////////////////////

// RGBXBlender defines the interface for 4-byte RGB pixels with padding byte.
// The blender handles color space interpretation and internal pixel ordering.
type RGBXBlender[S color.Space] interface {
	// GetPlain reads a 4-byte pixel and returns plain RGB components
	// interpreted according to color space S (padding byte ignored)
	GetPlain(px []byte) (r, g, b basics.Int8u)

	// SetPlain writes plain RGB components to a 4-byte pixel, mapping them to the
	// internal order of the blender (padding byte unchanged)
	SetPlain(px []byte, r, g, b basics.Int8u)

	// BlendPix blends plain RGB source into the 4-byte pixel with given alpha and coverage
	// r,g,b are interpreted according to S, and mapped to the order internal to the blender
	BlendPix(px []byte, r, g, b, a, cover basics.Int8u)
}

// RawRGBXOrder provides optional fast path for zero-cost index access for RGBX.
type RawRGBXOrder interface {
	IdxR() int
	IdxG() int
	IdxB() int
}

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> RGB destination (no alpha stored)
////////////////////////////////////////////////////////////////////////////////

type BlenderRGB8[S color.Space, O order.RGBOrder] struct{}

// Lerp by alpha*cover; destination stores only RGB (3 bytes).
func (BlenderRGB8[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Lerp(dst[o.IdxR()], r, alpha)
	dst[o.IdxG()] = color.RGBA8Lerp(dst[o.IdxG()], g, alpha)
	dst[o.IdxB()] = color.RGBA8Lerp(dst[o.IdxB()], b, alpha)
}

func (BlenderRGB8[S, O]) SetPlain(dst []basics.Int8u, r, g, b basics.Int8u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (BlenderRGB8[S, O]) GetPlain(src []basics.Int8u) (r, g, b basics.Int8u) {
	var o O
	r = src[o.IdxR()]
	g = src[o.IdxG()]
	b = src[o.IdxB()]
	return
}

// RawRGBOrder interface implementation for fast path access
func (BlenderRGB8[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGB8[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGB8[S, O]) IdxB() int { var o O; return o.IdxB() }

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

func (BlenderRGBPre[S, O]) SetPlain(dst []basics.Int8u, r, g, b basics.Int8u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (BlenderRGBPre[S, O]) GetPlain(src []basics.Int8u) (r, g, b basics.Int8u) {
	var o O
	r = src[o.IdxR()]
	g = src[o.IdxG()]
	b = src[o.IdxB()]
	return
}

// RawRGBOrder interface implementation for fast path access
func (BlenderRGBPre[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBPre[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBPre[S, O]) IdxB() int { var o O; return o.IdxB() }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> RGBX destination (4-byte with padding)
////////////////////////////////////////////////////////////////////////////////

type BlenderRGBX8[S color.Space, O order.RGBOrder] struct{}

// Lerp by alpha*cover; destination stores RGB in 4-byte format with padding.
func (BlenderRGBX8[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	alpha := color.RGBA8MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Lerp(dst[o.IdxR()], r, alpha)
	dst[o.IdxG()] = color.RGBA8Lerp(dst[o.IdxG()], g, alpha)
	dst[o.IdxB()] = color.RGBA8Lerp(dst[o.IdxB()], b, alpha)
	// Padding byte at position 3 remains unchanged
}

func (BlenderRGBX8[S, O]) SetPlain(dst []basics.Int8u, r, g, b basics.Int8u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
	// Padding byte remains unchanged
}

func (BlenderRGBX8[S, O]) GetPlain(src []basics.Int8u) (r, g, b basics.Int8u) {
	var o O
	r = src[o.IdxR()]
	g = src[o.IdxG()]
	b = src[o.IdxB()]
	// Padding byte ignored
	return
}

// RawRGBXOrder interface implementation for fast path access
func (BlenderRGBX8[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBX8[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBX8[S, O]) IdxB() int { var o O; return o.IdxB() }

////////////////////////////////////////////////////////////////////////////////
// Premultiplied source -> RGBX destination (4-byte with padding)
////////////////////////////////////////////////////////////////////////////////

type BlenderRGBXPre[S color.Space, O order.RGBOrder] struct{}

func (BlenderRGBXPre[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
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
	// Padding byte remains unchanged
}

func (BlenderRGBXPre[S, O]) SetPlain(dst []basics.Int8u, r, g, b basics.Int8u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
	// Padding byte remains unchanged
}

func (BlenderRGBXPre[S, O]) GetPlain(src []basics.Int8u) (r, g, b basics.Int8u) {
	var o O
	r = src[o.IdxR()]
	g = src[o.IdxG()]
	b = src[o.IdxB()]
	// Padding byte ignored
	return
}

// RawRGBXOrder interface implementation for fast path access
func (BlenderRGBXPre[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBXPre[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBXPre[S, O]) IdxB() int { var o O; return o.IdxB() }

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

func (bl BlenderRGBGamma[S, O, G]) SetPlain(dst []basics.Int8u, r, g, b basics.Int8u) {
	var o O
	dst[o.IdxR()] = r
	dst[o.IdxG()] = g
	dst[o.IdxB()] = b
}

func (bl BlenderRGBGamma[S, O, G]) GetPlain(src []basics.Int8u) (r, g, b basics.Int8u) {
	var o O
	r = src[o.IdxR()]
	g = src[o.IdxG()]
	b = src[o.IdxB()]
	return
}

////////////////////////////////////////////////////////////////////////////////
// Helpers for 8-bit RGB
////////////////////////////////////////////////////////////////////////////////

func BlendRGBPixel[B RGBBlender[S], S color.Space, O order.RGBOrder](
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

func BlendRGBHline[B RGBBlender[S], S color.Space, O order.RGBOrder](
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
		for range length {
			bl.BlendPix(dst[p:p+3], src.R, src.G, src.B, alpha, 255)
			p += pixStep
		}
		return
	}
	for i := range length {
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
	for range length {
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

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (consistent with RGBA pattern)
////////////////////////////////////////////////////////////////////////////////

// Linear space
type (
	BlenderRGB8LinearRGB = BlenderRGB8[color.Linear, order.RGB]
	BlenderRGB8LinearBGR = BlenderRGB8[color.Linear, order.BGR]

	BlenderRGB8PreLinearRGB = BlenderRGBPre[color.Linear, order.RGB]
	BlenderRGB8PreLinearBGR = BlenderRGBPre[color.Linear, order.BGR]
)

// sRGB space
type (
	BlenderRGB8SRGBrgb = BlenderRGB8[color.SRGB, order.RGB]
	BlenderRGB8SRGBbgr = BlenderRGB8[color.SRGB, order.BGR]

	BlenderRGB8PreSRGBrgb = BlenderRGBPre[color.SRGB, order.RGB]
	BlenderRGB8PreSRGBbgr = BlenderRGBPre[color.SRGB, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Platform-specific aliases (matching RGBA pattern)
////////////////////////////////////////////////////////////////////////////////

type (
	// Standard RGB (most common web/OpenGL format)
	BlenderRGB8Standard    = BlenderRGB8[color.SRGB, order.RGB]
	BlenderRGB8PreStandard = BlenderRGBPre[color.SRGB, order.RGB]

	// Windows/DirectX common format (BGR)
	BlenderBGR8Windows    = BlenderRGB8[color.SRGB, order.BGR]
	BlenderBGR8PreWindows = BlenderRGBPre[color.SRGB, order.BGR]

	// Linear space variants for high-quality rendering
	BlenderRGB8Linear    = BlenderRGB8[color.Linear, order.RGB]
	BlenderRGB8PreLinear = BlenderRGBPre[color.Linear, order.RGB]
	BlenderBGR8Linear    = BlenderRGB8[color.Linear, order.BGR]
	BlenderBGR8PreLinear = BlenderRGBPre[color.Linear, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Short aliases for common usage (matching RGBA pattern)
////////////////////////////////////////////////////////////////////////////////

type (
	// Generic order-specific aliases (similar to RGBA pattern)
	BlenderRGBGeneric[S color.Space] = BlenderRGB8[S, order.RGB]
	BlenderBGRGeneric[S color.Space] = BlenderRGB8[S, order.BGR]

	BlenderRGBPreGeneric[S color.Space] = BlenderRGBPre[S, order.RGB]
	BlenderBGRPreGeneric[S color.Space] = BlenderRGBPre[S, order.BGR]
)

////////////////////////////////////////////////////////////////////////////////
// Compatibility aliases for pixfmt
////////////////////////////////////////////////////////////////////////////////

type (
	// Primary compatibility aliases matching pixfmt usage
	BlenderRGB24SRGB   = BlenderRGB8[color.SRGB, order.RGB]
	BlenderBGR24SRGB   = BlenderRGB8[color.SRGB, order.BGR]
	BlenderRGB24Linear = BlenderRGB8[color.Linear, order.RGB]
	BlenderBGR24Linear = BlenderRGB8[color.Linear, order.BGR]

	// Pre-multiplied variants
	BlenderRGB24PreSRGB   = BlenderRGBPre[color.SRGB, order.RGB]
	BlenderBGR24PreSRGB   = BlenderRGBPre[color.SRGB, order.BGR]
	BlenderRGB24PreLinear = BlenderRGBPre[color.Linear, order.RGB]
	BlenderBGR24PreLinear = BlenderRGBPre[color.Linear, order.BGR]

	// Simplified aliases used by pixfmt constructors
	BlenderRGB24Pre = BlenderRGBPre[color.Linear, order.RGB]
	BlenderBGR24Pre = BlenderRGBPre[color.Linear, order.BGR]
)
