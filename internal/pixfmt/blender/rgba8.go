package blender

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

////////////////////////////////////////////////////////////////////////////////
// RGBA Blender interface
////////////////////////////////////////////////////////////////////////////////

// RGBABlender defines the minimal interface used by pixfmt implementations.
// The blender handles color space interpretation and internal pixel ordering.
type RGBABlender[S color.Space] interface {
	// GetPlain reads a pixel and returns plain (non-premultiplied) RGBA components
	// interpreted according to color space S
	GetPlain(px []byte) (r, g, b, a basics.Int8u)

	// SetPlain writes plain RGBA components to a pixel, mapping them to the
	// internal order and storage format of the blender
	SetPlain(px []byte, r, g, b, a basics.Int8u)

	// BlendPix blends plain RGBA source into the pixel with given coverage
	// r,g,b,a are interpreted according to S, and mapped to the order internal to the blender
	BlendPix(px []byte, r, g, b, a, cover basics.Int8u)
}

// RawRGBAOrder provides optional fast path for zero-cost index access.
// Blenders that expose direct index access should implement this interface
// to allow optimized operations when order-specific code is needed.
type RawRGBAOrder interface {
	IdxR() int
	IdxG() int
	IdxB() int
	IdxA() int
}

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA8 blends *plain* source into a premultiplied destination buffer.
// Matches AGG's blender_rgba (plain → premultiplied).
type BlenderRGBA8[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends a non-premultiplied RGBA source into a premultiplied buffer.
// Alpha is scaled by coverage; channels use lerp; alpha uses prelerp.
func (BlenderRGBA8[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	a = color.RGBA8MultCover(a, cover)
	if a == 0 {
		return
	}
	var o O
	dst[o.IdxR()] = color.RGBA8Lerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = color.RGBA8Lerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = color.RGBA8Lerp(dst[o.IdxB()], b, a)
	dst[o.IdxA()] = color.RGBA8Prelerp(dst[o.IdxA()], a, a)
}

func (BlenderRGBA8[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	var o O
	// SetPlain should set the exact plain/straight alpha values without premultiplying
	// The blending operations (BlendPix, etc.) handle premultiplication as needed
	dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = r, g, b, a
}

func (BlenderRGBA8[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u) {
	var o O
	// GetPlain returns the exact stored values without demultiplying
	// This matches SetPlain which stores plain/straight alpha values
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], src[o.IdxA()]
}

// RawRGBAOrder interface implementation for fast path access
func (BlenderRGBA8[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA8[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA8[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA8[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Premultiplied source -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA8Pre blends *premultiplied* source into a premultiplied destination buffer.
// Matches AGG's blender_rgba_pre (premultiplied → premultiplied).
type BlenderRGBA8Pre[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends a premultiplied RGBA source into a premultiplied buffer.
// Channels and alpha use prelerp. Coverage scales all premultiplied components.
func (BlenderRGBA8Pre[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
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
	dst[o.IdxA()] = color.RGBA8Prelerp(dst[o.IdxA()], a, a)
}

func (BlenderRGBA8Pre[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	BlenderRGBA8[S, O]{}.SetPlain(dst, r, g, b, a)
}

func (BlenderRGBA8Pre[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u) {
	return BlenderRGBA8[S, O]{}.GetPlain(src)
}

// RawRGBAOrder interface implementation for fast path access
func (BlenderRGBA8Pre[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA8Pre[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA8Pre[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA8Pre[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Plain destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA8Plain blends *plain* source into a *plain* destination buffer.
// Matches AGG's blender_rgba_plain (plain → plain): it premultiplies dst on-the-fly,
// blends in premultiplied space, then demultiplies to store plain again.
type BlenderRGBA8Plain[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends non-premultiplied src into non-premultiplied dst using the classic
// “premultiply → blend in premul → demultiply” approach.
func (BlenderRGBA8Plain[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	a = color.RGBA8MultCover(a, cover)
	if a == 0 {
		return
	}
	var o O

	da := dst[o.IdxA()]
	// premultiply dst on the fly
	dr := color.RGBA8Multiply(dst[o.IdxR()], da)
	dg := color.RGBA8Multiply(dst[o.IdxG()], da)
	db := color.RGBA8Multiply(dst[o.IdxB()], da)

	dr = color.RGBA8Lerp(dr, r, a)
	dg = color.RGBA8Lerp(dg, g, a)
	db = color.RGBA8Lerp(db, b, a)
	da = color.RGBA8Prelerp(da, a, a)

	if da > 0 {
		dst[o.IdxR()] = demul8(dr, da)
		dst[o.IdxG()] = demul8(dg, da)
		dst[o.IdxB()] = demul8(db, da)
		dst[o.IdxA()] = da
	} else {
		dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = 0, 0, 0, 0
	}
}

func (BlenderRGBA8Plain[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	var o O
	dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = r, g, b, a
}

func (BlenderRGBA8Plain[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], src[o.IdxA()]
}

// RawRGBAOrder interface implementation for fast path access
func (BlenderRGBA8Plain[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA8Plain[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA8Plain[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA8Plain[S, O]) IdxA() int { var o O; return o.IdxA() }

// BlendRGBAPixel blends a single pixel using the provided blender B.
// Works for any Space S and Order O, and never branches on order at runtime.
func BlendRGBAPixel[S color.Space, O order.RGBAOrder](
	dst []basics.Int8u,
	src color.RGBA8[S],
	cover basics.Int8u,
	b RGBABlender[S],
) {
	if src.IsTransparent() || cover == 0 {
		return
	}
	b.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
}

// CopyRGBAPixel writes the *plain* RGBA components to dst in order O.
// (Use this when you want a raw copy with no blending.)
func CopyRGBAPixel[S color.Space, O order.RGBAOrder](
	dst []basics.Int8u,
	src color.RGBA8[S],
) {
	var o O
	dst[o.IdxR()] = src.R
	dst[o.IdxG()] = src.G
	dst[o.IdxB()] = src.B
	dst[o.IdxA()] = src.A
}

// Blend a horizontal span
func BlendRGBAHline[S color.Space, O order.RGBAOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGBA8[S],
	covers []basics.Int8u, // nil => full cover
	b RGBABlender[S],
) {
	if length <= 0 || src.IsTransparent() {
		return
	}
	const pixStep = 4
	p := x * pixStep

	if covers == nil {
		for i := 0; i < length; i++ {
			b.BlendPix(dst[p:p+4], src.R, src.G, src.B, src.A, 255)
			p += pixStep
		}
		return
	}
	for i := 0; i < length; i++ {
		if c := covers[i]; c != 0 {
			b.BlendPix(dst[p:p+4], src.R, src.G, src.B, src.A, c)
		}
		p += pixStep
	}
}

// CopyRGBAHline copies a horizontal run of the same plain color into dst in order O.
func CopyRGBAHline[S color.Space, O order.RGBAOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGBA8[S],
) {
	if length <= 0 {
		return
	}
	var o O
	const pixStep = 4
	p := x * pixStep
	for i := 0; i < length; i++ {
		dst[p+o.IdxR()] = src.R
		dst[p+o.IdxG()] = src.G
		dst[p+o.IdxB()] = src.B
		dst[p+o.IdxA()] = src.A
		p += pixStep
	}
}

// FillRGBASpan is a synonym of CopyRGBAHline (explicit name for intent).
func FillRGBASpan[S color.Space, O order.RGBAOrder](
	dst []basics.Int8u,
	x, length int,
	src color.RGBA8[S],
) {
	CopyRGBAHline[S, O](dst, x, length, src)
}

// demul8 converts a premultiplied component x back to straight by x * 255 / a with rounding.
func demul8(x, a basics.Int8u) basics.Int8u {
	// (x*255 + a/2) / a  — classic rounded divide
	return basics.Int8u((uint32(x)*255 + uint32(a)/2) / uint32(a))
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases for common Order/Space combinations
////////////////////////////////////////////////////////////////////////////////

// Linear space
type (
	BlenderRGBA8LinearRGBA = BlenderRGBA8[color.Linear, order.RGBA]
	BlenderRGBA8LinearBGRA = BlenderRGBA8[color.Linear, order.BGRA]
	BlenderRGBA8LinearARGB = BlenderRGBA8[color.Linear, order.ARGB]
	BlenderRGBA8LinearABGR = BlenderRGBA8[color.Linear, order.ABGR]

	BlenderRGBA8PreLinearRGBA = BlenderRGBA8Pre[color.Linear, order.RGBA]
	BlenderRGBA8PreLinearBGRA = BlenderRGBA8Pre[color.Linear, order.BGRA]
	BlenderRGBA8PreLinearARGB = BlenderRGBA8Pre[color.Linear, order.ARGB]
	BlenderRGBA8PreLinearABGR = BlenderRGBA8Pre[color.Linear, order.ABGR]

	BlenderRGBA8PlainLinearRGBA = BlenderRGBA8Plain[color.Linear, order.RGBA]
	BlenderRGBA8PlainLinearBGRA = BlenderRGBA8Plain[color.Linear, order.BGRA]
	BlenderRGBA8PlainLinearARGB = BlenderRGBA8Plain[color.Linear, order.ARGB]
	BlenderRGBA8PlainLinearABGR = BlenderRGBA8Plain[color.Linear, order.ABGR]
)

// sRGB space
type (
	BlenderRGBA8SRGBrgba = BlenderRGBA8[color.SRGB, order.RGBA]
	BlenderRGBA8SRGBbgra = BlenderRGBA8[color.SRGB, order.BGRA]
	BlenderRGBA8SRGBargb = BlenderRGBA8[color.SRGB, order.ARGB]
	BlenderRGBA8SRGBabgr = BlenderRGBA8[color.SRGB, order.ABGR]

	BlenderRGBA8PreSRGBrgba = BlenderRGBA8Pre[color.SRGB, order.RGBA]
	BlenderRGBA8PreSRGBbgra = BlenderRGBA8Pre[color.SRGB, order.BGRA]
	BlenderRGBA8PreSRGBargb = BlenderRGBA8Pre[color.SRGB, order.ARGB]
	BlenderRGBA8PreSRGBabgr = BlenderRGBA8Pre[color.SRGB, order.ABGR]

	BlenderRGBA8PlainSRGBrgba = BlenderRGBA8Plain[color.SRGB, order.RGBA]
	BlenderRGBA8PlainSRGBbgra = BlenderRGBA8Plain[color.SRGB, order.BGRA]
	BlenderRGBA8PlainSRGBargb = BlenderRGBA8Plain[color.SRGB, order.ARGB]
	BlenderRGBA8PlainSRGBabgr = BlenderRGBA8Plain[color.SRGB, order.ABGR]
)

////////////////////////////////////////////////////////////////////////////////
// Aliases
////////////////////////////////////////////////////////////////////////////////

// Aliases (plain -> premul)
type (
	BlenderARGB8[S color.Space] = BlenderRGBA8[S, order.ARGB]
	BlenderBGRA8[S color.Space] = BlenderRGBA8[S, order.BGRA]
	BlenderABGR8[S color.Space] = BlenderRGBA8[S, order.ABGR]
)

// Premultiplied source -> premultiplied dst
type (
	BlenderARGB8Pre[S color.Space] = BlenderRGBA8Pre[S, order.ARGB]
	BlenderBGRA8Pre[S color.Space] = BlenderRGBA8Pre[S, order.BGRA]
	BlenderABGR8Pre[S color.Space] = BlenderRGBA8Pre[S, order.ABGR]
)

// Plain -> plain
type (
	BlenderARGB8Plain[S color.Space] = BlenderRGBA8Plain[S, order.ARGB]
	BlenderBGRA8Plain[S color.Space] = BlenderRGBA8Plain[S, order.BGRA]
	BlenderABGR8Plain[S color.Space] = BlenderRGBA8Plain[S, order.ABGR]
)

////////////////////////////////////////////////////////////////////////////////
// Common platform-specific aliases
////////////////////////////////////////////////////////////////////////////////

// Most common combinations for various platforms
type (
	// Standard RGBA (most common)
	BlenderRGBA8Standard      = BlenderRGBA8[color.SRGB, order.RGBA]
	BlenderRGBA8PreStandard   = BlenderRGBA8Pre[color.SRGB, order.RGBA]
	BlenderRGBA8PlainStandard = BlenderRGBA8Plain[color.SRGB, order.RGBA]

	// Windows/DirectX common format (BGRA)
	BlenderBGRA8Windows      = BlenderRGBA8[color.SRGB, order.BGRA]
	BlenderBGRA8PreWindows   = BlenderRGBA8Pre[color.SRGB, order.BGRA]
	BlenderBGRA8PlainWindows = BlenderRGBA8Plain[color.SRGB, order.BGRA]

	// Mac/iOS common format (ARGB)
	BlenderARGB8Mac      = BlenderRGBA8[color.SRGB, order.ARGB]
	BlenderARGB8PreMac   = BlenderRGBA8Pre[color.SRGB, order.ARGB]
	BlenderARGB8PlainMac = BlenderRGBA8Plain[color.SRGB, order.ARGB]

	// Android common format (ABGR)
	BlenderABGR8Android      = BlenderRGBA8[color.SRGB, order.ABGR]
	BlenderABGR8PreAndroid   = BlenderRGBA8Pre[color.SRGB, order.ABGR]
	BlenderABGR8PlainAndroid = BlenderRGBA8Plain[color.SRGB, order.ABGR]
)

////////////////////////////////////////////////////////////////////////////////
// Linear space aliases for high-quality rendering
////////////////////////////////////////////////////////////////////////////////

type (
	// Linear space variants (better for blending quality)
	BlenderRGBA8Linear      = BlenderRGBA8[color.Linear, order.RGBA]
	BlenderRGBA8PreLinear   = BlenderRGBA8Pre[color.Linear, order.RGBA]
	BlenderRGBA8PlainLinear = BlenderRGBA8Plain[color.Linear, order.RGBA]

	BlenderBGRA8Linear      = BlenderRGBA8[color.Linear, order.BGRA]
	BlenderBGRA8PreLinear   = BlenderRGBA8Pre[color.Linear, order.BGRA]
	BlenderBGRA8PlainLinear = BlenderRGBA8Plain[color.Linear, order.BGRA]

	BlenderARGB8Linear      = BlenderRGBA8[color.Linear, order.ARGB]
	BlenderARGB8PreLinear   = BlenderRGBA8Pre[color.Linear, order.ARGB]
	BlenderARGB8PlainLinear = BlenderRGBA8Plain[color.Linear, order.ARGB]

	BlenderABGR8Linear      = BlenderRGBA8[color.Linear, order.ABGR]
	BlenderABGR8PreLinear   = BlenderRGBA8Pre[color.Linear, order.ABGR]
	BlenderABGR8PlainLinear = BlenderRGBA8Plain[color.Linear, order.ABGR]
)

////////////////////////////////////////////////////////////////////////////////
// Generic aliases matching C++ AGG naming
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA is the generic 8-bit RGBA blender matching C++ blender_rgba<ColorT, Order>.
// This is an alias for BlenderRGBA8 to match the C++ naming convention where
// blender_rgba<rgba8, order_rgba> is equivalent to blender_rgba32.
type BlenderRGBA[S color.Space, O order.RGBAOrder] = BlenderRGBA8[S, O]

// BlenderRGBAPre is the generic 8-bit premultiplied RGBA blender matching C++ blender_rgba_pre<ColorT, Order>.
type BlenderRGBAPre[S color.Space, O order.RGBAOrder] = BlenderRGBA8Pre[S, O]

// BlenderRGBAPlain is the generic 8-bit plain RGBA blender matching C++ blender_rgba_plain<ColorT, Order>.
type BlenderRGBAPlain[S color.Space, O order.RGBAOrder] = BlenderRGBA8Plain[S, O]

////////////////////////////////////////////////////////////////////////////////
// Short aliases for common usage
////////////////////////////////////////////////////////////////////////////////

type (
	// Ultra-short aliases for the most common cases
	RGBA8Blender = BlenderRGBA8[color.SRGB, order.RGBA]
	BGRA8Blender = BlenderRGBA8[color.SRGB, order.BGRA]
	ARGB8Blender = BlenderRGBA8[color.SRGB, order.ARGB]
	ABGR8Blender = BlenderRGBA8[color.SRGB, order.ABGR]

	RGBA8PreBlender = BlenderRGBA8Pre[color.SRGB, order.RGBA]
	BGRA8PreBlender = BlenderRGBA8Pre[color.SRGB, order.BGRA]
	ARGB8PreBlender = BlenderRGBA8Pre[color.SRGB, order.ARGB]
	ABGR8PreBlender = BlenderRGBA8Pre[color.SRGB, order.ABGR]

	RGBA8PlainBlender = BlenderRGBA8Plain[color.SRGB, order.RGBA]
	BGRA8PlainBlender = BlenderRGBA8Plain[color.SRGB, order.BGRA]
	ARGB8PlainBlender = BlenderRGBA8Plain[color.SRGB, order.ARGB]
	ABGR8PlainBlender = BlenderRGBA8Plain[color.SRGB, order.ABGR]
)
