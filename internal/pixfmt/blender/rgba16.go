package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// Interfaces RGBA (16-bit)
////////////////////////////////////////////////////////////////////////////////

// RGBABlender16 defines the minimal interface used by RGBA16 pixfmt implementations.
// The blender handles color space interpretation and internal pixel ordering.
type RGBABlender16[S color.Space] interface {
	// GetPlain reads a pixel and returns plain RGBA components
	// interpreted according to color space S
	GetPlain(src []basics.Int8u) (r, g, b, a basics.Int16u)

	// SetPlain writes plain RGBA components to a pixel, mapping them to the
	// internal order of the blender
	SetPlain(dst []basics.Int8u, r, g, b, a basics.Int16u)

	// BlendPix blends plain RGBA source into the pixel with given coverage
	// r,g,b,a are interpreted according to S, and mapped to the order internal to the blender
	BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u)
}

// RawRGBA16Order provides optional fast path for zero-cost index access for RGBA16.
type RawRGBA16Order interface {
	IdxR() int
	IdxG() int
	IdxB() int
	IdxA() int
}

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Premultiplied destination (16-bit)
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA16 blends *plain* 16-bit source into a premultiplied destination buffer.
// Analogous to AGG's blender_rgba for 16-bit channels.
type BlenderRGBA16[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends a non-premultiplied RGBA16 source into a premultiplied buffer.
// Alpha is scaled by coverage; channels use lerp; alpha uses prelerp.
func (BlenderRGBA16[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O

	// Load dst components (little-endian) in order O
	dr := basics.Int16u(dst[o.IdxR()*2]) | basics.Int16u(dst[o.IdxR()*2+1])<<8
	dg := basics.Int16u(dst[o.IdxG()*2]) | basics.Int16u(dst[o.IdxG()*2+1])<<8
	db := basics.Int16u(dst[o.IdxB()*2]) | basics.Int16u(dst[o.IdxB()*2+1])<<8
	da := basics.Int16u(dst[o.IdxA()*2]) | basics.Int16u(dst[o.IdxA()*2+1])<<8

	// Blend
	dr = color.RGBA16Lerp(dr, r, alpha)
	dg = color.RGBA16Lerp(dg, g, alpha)
	db = color.RGBA16Lerp(db, b, alpha)
	da = color.RGBA16Prelerp(da, alpha, alpha)

	// Store back (little-endian)
	dst[o.IdxR()*2+0], dst[o.IdxR()*2+1] = basics.Int8u(dr), basics.Int8u(dr>>8)
	dst[o.IdxG()*2+0], dst[o.IdxG()*2+1] = basics.Int8u(dg), basics.Int8u(dg>>8)
	dst[o.IdxB()*2+0], dst[o.IdxB()*2+1] = basics.Int8u(db), basics.Int8u(db>>8)
	dst[o.IdxA()*2+0], dst[o.IdxA()*2+1] = basics.Int8u(da), basics.Int8u(da>>8)
}

func (BlenderRGBA16[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int16u) {
	// store PREMULTIPLIED to the framebuffer
	var o O
	pr := color.RGBA16Multiply(r, a)
	pg := color.RGBA16Multiply(g, a)
	pb := color.RGBA16Multiply(b, a)

	put16 := func(off int, v basics.Int16u) {
		dst[off+0] = basics.Int8u(v)
		dst[off+1] = basics.Int8u(v >> 8)
	}
	put16(o.IdxR()*2, pr)
	put16(o.IdxG()*2, pg)
	put16(o.IdxB()*2, pb)
	put16(o.IdxA()*2, a)
}

func (BlenderRGBA16[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int16u) {
	// read PREMULTIPLIED from the framebuffer, return PLAIN
	var o O
	get16 := func(off int) basics.Int16u {
		return basics.Int16u(src[off]) | basics.Int16u(src[off+1])<<8
	}
	pr := get16(o.IdxR() * 2)
	pg := get16(o.IdxG() * 2)
	pb := get16(o.IdxB() * 2)
	a = get16(o.IdxA() * 2)

	if a != 0 {
		r = demul16(pr, a)
		g = demul16(pg, a)
		b = demul16(pb, a)
	} else {
		r, g, b = 0, 0, 0
	}
	return
}

// RawRGBA16Order interface implementation for fast path access
func (BlenderRGBA16[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA16[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA16[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA16[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Premultiplied source -> Premultiplied destination (16-bit)
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA16Pre blends *premultiplied* 16-bit source into a premultiplied destination.
// Analogous to AGG's blender_rgba_pre for 16-bit channels.
type BlenderRGBA16Pre[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends premultiplied RGBA16 into premultiplied buffer.
// Coverage scales all premultiplied components; channels & alpha use prelerp.
func (BlenderRGBA16Pre[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	// Scale by coverage only when not full mask
	if cover != 0xFFFF {
		r = color.RGBA16MultCover(r, cover)
		g = color.RGBA16MultCover(g, cover)
		b = color.RGBA16MultCover(b, cover)
		a = color.RGBA16MultCover(a, cover)
	}
	// Early-out when the source contributes nothing
	if a|r|g|b == 0 {
		return
	}
	var o O

	// Load dst (little-endian)
	dr := basics.Int16u(dst[o.IdxR()*2]) | basics.Int16u(dst[o.IdxR()*2+1])<<8
	dg := basics.Int16u(dst[o.IdxG()*2]) | basics.Int16u(dst[o.IdxG()*2+1])<<8
	db := basics.Int16u(dst[o.IdxB()*2]) | basics.Int16u(dst[o.IdxB()*2+1])<<8
	da := basics.Int16u(dst[o.IdxA()*2]) | basics.Int16u(dst[o.IdxA()*2+1])<<8

	// Blend in premultiplied space
	dr = color.RGBA16Prelerp(dr, r, a)
	dg = color.RGBA16Prelerp(dg, g, a)
	db = color.RGBA16Prelerp(db, b, a)
	da = color.RGBA16Prelerp(da, a, a)

	// Store (little-endian)
	dst[o.IdxR()*2+0], dst[o.IdxR()*2+1] = basics.Int8u(dr), basics.Int8u(dr>>8)
	dst[o.IdxG()*2+0], dst[o.IdxG()*2+1] = basics.Int8u(dg), basics.Int8u(dg>>8)
	dst[o.IdxB()*2+0], dst[o.IdxB()*2+1] = basics.Int8u(db), basics.Int8u(db>>8)
	dst[o.IdxA()*2+0], dst[o.IdxA()*2+1] = basics.Int8u(da), basics.Int8u(da>>8)
}

func (BlenderRGBA16Pre[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int16u) {
	BlenderRGBA16[S, O]{}.SetPlain(dst, r, g, b, a) // premultiply on write
}

func (BlenderRGBA16Pre[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int16u) {
	return BlenderRGBA16[S, O]{}.GetPlain(src) // demultiply on read
}

// RawRGBA16Order interface implementation for fast path access
func (BlenderRGBA16Pre[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA16Pre[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA16Pre[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA16Pre[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Plain destination (16-bit)
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA16Plain blends *plain* 16-bit source into a *plain* destination.
// It mirrors AGG's blender_rgba_plain: premultiply dst → blend → demultiply.
type BlenderRGBA16Plain[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends plain src into plain dst using premultiplied math internally.
func (BlenderRGBA16Plain[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int16u) {
	alpha := color.RGBA16MultCover(a, cover)
	if alpha == 0 {
		return
	}
	var o O

	// Load dst (little-endian)
	dr := basics.Int16u(dst[o.IdxR()*2]) | basics.Int16u(dst[o.IdxR()*2+1])<<8
	dg := basics.Int16u(dst[o.IdxG()*2]) | basics.Int16u(dst[o.IdxG()*2+1])<<8
	db := basics.Int16u(dst[o.IdxB()*2]) | basics.Int16u(dst[o.IdxB()*2+1])<<8
	da := basics.Int16u(dst[o.IdxA()*2]) | basics.Int16u(dst[o.IdxA()*2+1])<<8

	// Premultiply destination by its alpha
	pdr := color.RGBA16Multiply(dr, da)
	pdg := color.RGBA16Multiply(dg, da)
	pdb := color.RGBA16Multiply(db, da)

	// Blend in premul space
	pdr = color.RGBA16Lerp(pdr, r, alpha)
	pdg = color.RGBA16Lerp(pdg, g, alpha)
	pdb = color.RGBA16Lerp(pdb, b, alpha)
	da = color.RGBA16Prelerp(da, alpha, alpha)

	// Demultiply back to plain
	if da != 0 {
		dr = demul16(pdr, da)
		dg = demul16(pdg, da)
		db = demul16(pdb, da)
	} else {
		dr, dg, db = 0, 0, 0
	}

	// Store (little-endian)
	dst[o.IdxR()*2+0], dst[o.IdxR()*2+1] = basics.Int8u(dr), basics.Int8u(dr>>8)
	dst[o.IdxG()*2+0], dst[o.IdxG()*2+1] = basics.Int8u(dg), basics.Int8u(dg>>8)
	dst[o.IdxB()*2+0], dst[o.IdxB()*2+1] = basics.Int8u(db), basics.Int8u(db>>8)
	dst[o.IdxA()*2+0], dst[o.IdxA()*2+1] = basics.Int8u(da), basics.Int8u(da>>8)
}

func (BlenderRGBA16Plain[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int16u) {
	var o O
	put16 := func(off int, v basics.Int16u) {
		dst[off+0] = basics.Int8u(v)
		dst[off+1] = basics.Int8u(v >> 8)
	}
	put16(o.IdxR()*2, r)
	put16(o.IdxG()*2, g)
	put16(o.IdxB()*2, b)
	put16(o.IdxA()*2, a)
}

func (BlenderRGBA16Plain[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int16u) {
	var o O
	get16 := func(off int) basics.Int16u {
		return basics.Int16u(src[off]) | basics.Int16u(src[off+1])<<8
	}
	r = get16(o.IdxR() * 2)
	g = get16(o.IdxG() * 2)
	b = get16(o.IdxB() * 2)
	a = get16(o.IdxA() * 2)
	return
}

// RawRGBA16Order interface implementation for fast path access
func (BlenderRGBA16Plain[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA16Plain[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA16Plain[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA16Plain[S, O]) IdxA() int { var o O; return o.IdxA() }

// demul16 converts a premultiplied component x back to straight: round(x * 65535 / a).
func demul16(x, a basics.Int16u) basics.Int16u {
	return basics.Int16u((uint32(x)*65535 + uint32(a)/2) / uint32(a))
}

////////////////////////////////////////////////////////////////////////////////
// Helpers & concrete aliases
////////////////////////////////////////////////////////////////////////////////

// BlendRGBA16Pixel is a typed helper like the 8-bit version.
func BlendRGBA16Pixel[B RGBABlender16[S], S color.Space](
	dst []basics.Int8u,
	src color.RGBA16[S],
	cover basics.Int16u,
	b B,
) {
	if src.IsTransparent() || cover == 0 {
		return
	}
	b.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
}

// Common aliases (Linear space + various byte orders)
type (
	BlenderRGBA16LinearRGBA = BlenderRGBA16[color.Linear, order.RGBA]
	BlenderRGBA16LinearBGRA = BlenderRGBA16[color.Linear, order.BGRA]
	BlenderRGBA16LinearARGB = BlenderRGBA16[color.Linear, order.ARGB]
	BlenderRGBA16LinearABGR = BlenderRGBA16[color.Linear, order.ABGR]

	BlenderRGBA16PreLinearRGBA = BlenderRGBA16Pre[color.Linear, order.RGBA]
	BlenderRGBA16PreLinearBGRA = BlenderRGBA16Pre[color.Linear, order.BGRA]
	BlenderRGBA16PreLinearARGB = BlenderRGBA16Pre[color.Linear, order.ARGB]
	BlenderRGBA16PreLinearABGR = BlenderRGBA16Pre[color.Linear, order.ABGR]

	BlenderRGBA16PlainLinearRGBA = BlenderRGBA16Plain[color.Linear, order.RGBA]
	BlenderRGBA16PlainLinearBGRA = BlenderRGBA16Plain[color.Linear, order.BGRA]
	BlenderRGBA16PlainLinearARGB = BlenderRGBA16Plain[color.Linear, order.ARGB]
	BlenderRGBA16PlainLinearABGR = BlenderRGBA16Plain[color.Linear, order.ABGR]
)

////////////////////////////////////////////////////////////////////////////////
// Aliases
////////////////////////////////////////////////////////////////////////////////

// Aliases (plain -> premul)
type (
	BlenderARGB16[S color.Space] = BlenderRGBA16[S, order.ARGB]
	BlenderBGRA16[S color.Space] = BlenderRGBA16[S, order.BGRA]
	BlenderABGR16[S color.Space] = BlenderRGBA16[S, order.ABGR]
)

// Premultiplied source -> premultiplied dst
type (
	BlenderARGB16Pre[S color.Space] = BlenderRGBA16Pre[S, order.ARGB]
	BlenderBGRA16Pre[S color.Space] = BlenderRGBA16Pre[S, order.BGRA]
	BlenderABGR16Pre[S color.Space] = BlenderRGBA16Pre[S, order.ABGR]
)

// Plain -> plain
type (
	BlenderARGB16Plain[S color.Space] = BlenderRGBA16Plain[S, order.ARGB]
	BlenderBGRA16Plain[S color.Space] = BlenderRGBA16Plain[S, order.BGRA]
	BlenderABGR16Plain[S color.Space] = BlenderRGBA16Plain[S, order.ABGR]
)
