package blender

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
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

// RGBAFastBlender extends RawRGBAOrder with blend-mode information, enabling
// pixfmt methods to inline the blend math directly instead of dispatching
// through the generic BlendPix interface per pixel.
//
// This mirrors C++ AGG's template monomorphization: the blend formula is
// resolved once at the top of the span loop, then the inner loop uses
// concrete (inlinable) color arithmetic.
type RGBAFastBlender interface {
	RawRGBAOrder
	// PremulSrc returns true when the source color is premultiplied.
	//   false → plain→premul (blender_rgba):     lerp for RGB, prelerp for A
	//   true  → premul→premul (blender_rgba_pre): prelerp for all channels
	PremulSrc() bool
}

// RGBAFastPathDisabler marks blenders whose BlendPix has custom channel math
// that cannot be replaced by the standard RGBA fast paths.
type RGBAFastPathDisabler interface {
	DisableRGBAFastPath() bool
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
func (BlenderRGBA8[S, O]) IdxR() int       { var o O; return o.IdxR() }
func (BlenderRGBA8[S, O]) IdxG() int       { var o O; return o.IdxG() }
func (BlenderRGBA8[S, O]) IdxB() int       { var o O; return o.IdxB() }
func (BlenderRGBA8[S, O]) IdxA() int       { var o O; return o.IdxA() }
func (BlenderRGBA8[S, O]) PremulSrc() bool { return false }

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
func (BlenderRGBA8Pre[S, O]) IdxR() int       { var o O; return o.IdxR() }
func (BlenderRGBA8Pre[S, O]) IdxG() int       { var o O; return o.IdxG() }
func (BlenderRGBA8Pre[S, O]) IdxB() int       { var o O; return o.IdxB() }
func (BlenderRGBA8Pre[S, O]) IdxA() int       { var o O; return o.IdxA() }
func (BlenderRGBA8Pre[S, O]) PremulSrc() bool { return true }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Plain destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA8Plain blends *plain* source into a *plain* destination buffer.
// Matches AGG's blender_rgba_plain (plain → plain): it premultiplies dst on-the-fly,
// blends in premultiplied space, then demultiplies to store plain again.
type BlenderRGBA8Plain[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends non-premultiplied src into a non-premultiplied destination.
// It matches AGG's blender_rgba_plain: premultiply the destination RGB by its
// alpha, blend in premultiplied space, then demultiply the result for storage.
func (BlenderRGBA8Plain[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	a = color.RGBA8MultCover(a, cover)
	if a == 0 {
		return
	}
	var o O

	da := dst[o.IdxA()]
	pr := color.RGBA8Multiply(dst[o.IdxR()], da)
	pg := color.RGBA8Multiply(dst[o.IdxG()], da)
	pb := color.RGBA8Multiply(dst[o.IdxB()], da)

	dst[o.IdxR()] = color.RGBA8Lerp(pr, r, a)
	dst[o.IdxG()] = color.RGBA8Lerp(pg, g, a)
	dst[o.IdxB()] = color.RGBA8Lerp(pb, b, a)
	dst[o.IdxA()] = color.RGBA8Prelerp(da, a, a)
	demultiplyRGBA8Plain(dst, o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA())
}

func demultiplyRGBA8Plain(px []basics.Int8u, ir, ig, ib, ia int) {
	a := px[ia]
	if a >= color.RGBA8BaseMask {
		return
	}
	if a == 0 {
		px[ir], px[ig], px[ib] = 0, 0, 0
		return
	}
	px[ir] = demultiplyRGBA8PlainChannel(px[ir], a)
	px[ig] = demultiplyRGBA8PlainChannel(px[ig], a)
	px[ib] = demultiplyRGBA8PlainChannel(px[ib], a)
}

func demultiplyRGBA8PlainChannel(v, a basics.Int8u) basics.Int8u {
	x := uint32(v) * color.RGBA8BaseMask / uint32(a)
	if x > color.RGBA8BaseMask {
		return color.RGBA8BaseMask
	}
	return basics.Int8u(x)
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

////////////////////////////////////////////////////////////////////////////////
// Plain source -> Plain destination, high precision
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA8PlainFixed is AGG 2.4's original high-precision plain blender.
// BlenderRGBA8Plain ports the agg24-svn rewrite, which routes the blend through
// multiply -> lerp -> demultiply and loses a bit of precision doing so; this one
// keeps the composite in a single fixed-point expression and divides by the
// combined alpha directly. Matplotlib restores exactly this form for its Agg
// backend (src/agg_workaround.h, fixed_blender_rgba_plain), so a renderer that
// wants pixel parity with Matplotlib needs this blender rather than the SVN one.
type BlenderRGBA8PlainFixed[S color.Space, O order.RGBAOrder] struct{}

// BlendPix blends non-premultiplied src into a non-premultiplied destination.
// The arithmetic is deliberately identical to the C++ original:
//
//	a = ((alpha + a) << 8) - alpha * a
//	p[R] = (((cr << 8) - r) * alpha + (r << 8)) / a    // r = p[R] * p[A]
//
// Note the integer division truncates and the divisor is the *unshifted*
// combined alpha; both are load-bearing for parity.
func (BlenderRGBA8PlainFixed[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	var o O

	// C++ AGG never reaches blend_pix for an opaque source at full coverage --
	// copy_or_blend_pix and blend_hline short-circuit to a plain store first.
	// The formula below is not exact for that case (it would return 199 for a
	// source of 200), so the short-circuit is load-bearing, and this blender
	// carries it because the pixfmt's opaque fast path is reserved for
	// RGBAFastBlender implementations.
	if a == color.RGBA8BaseMask && cover == color.RGBA8BaseMask {
		dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = r, g, b, a
		return
	}

	alpha := int64(color.RGBA8MultCover(a, cover))
	if alpha == 0 {
		return
	}

	da := int64(dst[o.IdxA()])
	pr := int64(dst[o.IdxR()]) * da
	pg := int64(dst[o.IdxG()]) * da
	pb := int64(dst[o.IdxB()]) * da

	combined := ((alpha + da) << 8) - alpha*da
	if combined == 0 {
		dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = 0, 0, 0, 0
		return
	}

	dst[o.IdxA()] = basics.Int8u(combined >> 8)
	dst[o.IdxR()] = basics.Int8u(((int64(r)<<8-pr)*alpha + pr<<8) / combined)
	dst[o.IdxG()] = basics.Int8u(((int64(g)<<8-pg)*alpha + pg<<8) / combined)
	dst[o.IdxB()] = basics.Int8u(((int64(b)<<8-pb)*alpha + pb<<8) / combined)
}

func (BlenderRGBA8PlainFixed[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	var o O
	dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = r, g, b, a
}

func (BlenderRGBA8PlainFixed[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], src[o.IdxA()]
}

// RawRGBAOrder interface implementation for fast path access. PremulSrc is
// deliberately not implemented: the standard RGBA fast paths substitute lerp
// math for BlendPix, which is the very thing this blender exists to avoid.
func (BlenderRGBA8PlainFixed[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA8PlainFixed[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA8PlainFixed[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA8PlainFixed[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Gamma-correct (linearising) source -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

// GammaLUT8 is the minimal gamma lookup table interface for gamma-correct blending.
// Dir converts a display (sRGB-like) value to linear; Inv is the reverse.
// Both SimpleGammaLut and GammaLUT8 from the gamma package satisfy this interface.
type GammaLUT8 interface {
	Dir(basics.Int8u) basics.Int8u
	Inv(basics.Int8u) basics.Int8u
}

// BlenderRGBA8Gamma blends plain source into a premultiplied destination using a
// gamma lookup table, matching C++ AGG's blender_rgb_gamma.
//
// Copy operations (SetPlain/GetPlain) are raw — gamma only applies during blending.
// RGB is linearised with Dir, blended in linear space, then re-encoded with Inv.
// Alpha uses the standard prelerp formula.
type BlenderRGBA8Gamma[S color.Space, O order.RGBAOrder] struct {
	Lut GammaLUT8
}

// BlendPix blends a plain RGBA source into a premultiplied destination buffer,
// linearising RGB through the gamma LUT before blending.
// Matches C++ blender_rgb_gamma::blend_pix:
//
//	p[C] = inv( (dir(src[C]) - dir(dst[C])) * alpha >> 8 + dir(dst[C]) )
func (bl BlenderRGBA8Gamma[S, O]) BlendPix(dst []basics.Int8u, cr, cg, cb, a, cover basics.Int8u) {
	a = color.RGBA8MultCover(a, cover)
	if a == 0 {
		return
	}
	var o O
	dr := int(bl.Lut.Dir(dst[o.IdxR()]))
	dg := int(bl.Lut.Dir(dst[o.IdxG()]))
	db := int(bl.Lut.Dir(dst[o.IdxB()]))
	sr := int(bl.Lut.Dir(cr))
	sg := int(bl.Lut.Dir(cg))
	sb := int(bl.Lut.Dir(cb))
	ia := int(a)
	dst[o.IdxR()] = bl.Lut.Inv(basics.Int8u(((sr - dr) * ia >> 8) + dr))
	dst[o.IdxG()] = bl.Lut.Inv(basics.Int8u(((sg - dg) * ia >> 8) + dg))
	dst[o.IdxB()] = bl.Lut.Inv(basics.Int8u(((sb - db) * ia >> 8) + db))
	dst[o.IdxA()] = color.RGBA8Prelerp(dst[o.IdxA()], a, a)
}

func (bl BlenderRGBA8Gamma[S, O]) SetPlain(dst []basics.Int8u, r, g, b, a basics.Int8u) {
	BlenderRGBA8[S, O]{}.SetPlain(dst, r, g, b, a)
}

func (bl BlenderRGBA8Gamma[S, O]) GetPlain(src []basics.Int8u) (r, g, b, a basics.Int8u) {
	return BlenderRGBA8[S, O]{}.GetPlain(src)
}

// RawRGBAOrder interface implementation for fast path access
func (bl BlenderRGBA8Gamma[S, O]) IdxR() int       { var o O; return o.IdxR() }
func (bl BlenderRGBA8Gamma[S, O]) IdxG() int       { var o O; return o.IdxG() }
func (bl BlenderRGBA8Gamma[S, O]) IdxB() int       { var o O; return o.IdxB() }
func (bl BlenderRGBA8Gamma[S, O]) IdxA() int       { var o O; return o.IdxA() }
func (bl BlenderRGBA8Gamma[S, O]) PremulSrc() bool { return false }
func (bl BlenderRGBA8Gamma[S, O]) DisableRGBAFastPath() bool {
	return true
}

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
