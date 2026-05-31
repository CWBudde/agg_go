package blender

import (
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// Float (128-bit, 4 x float32) RGBA blenders.
//
// This is the structural twin of the 8-bit blenders in rgba8.go, operating on
// []float32 pixels with channels in [0,1] and coverage in [0,1]. It pairs with
// the float color color.RGBA32 and mirrors C++ AGG's blender_rgba{,_pre,_plain}
// instantiated with the float rgba32 color type (agg_pixfmt_rgba.h). The "128"
// suffix denotes the total pixel width (4 x float32 = 128 bits) and avoids the
// naming collision with the existing 8-bit PixFmtRGBA32/BlenderRGBA8 aliases,
// matching AGG's own pixfmt_rgba128 convention.

////////////////////////////////////////////////////////////////////////////////
// Interface
////////////////////////////////////////////////////////////////////////////////

// RGBA128Blender blends 4 x float32 RGBA pixels in color space S.
// The blender owns channel ordering and premultiply interpretation.
type RGBA128Blender[S color.Space] interface {
	// GetPlain reads a pixel and returns plain (non-premultiplied) RGBA components.
	GetPlain(px []float32) (r, g, b, a float32)

	// SetPlain writes plain RGBA components, mapping to the internal channel order.
	SetPlain(px []float32, r, g, b, a float32)

	// BlendPix blends plain RGBA source into the pixel with the given coverage.
	BlendPix(px []float32, r, g, b, a, cover float32)
}

////////////////////////////////////////////////////////////////////////////////
// Arithmetic helpers (float)
////////////////////////////////////////////////////////////////////////////////

// RGBA128Lerp: straight alpha interpolation, dst + (src-dst)*a.
func RGBA128Lerp(p, q, a float32) float32 { return p + (q-p)*a }

// RGBA128Prelerp: premultiplied interpolation, p + q - p*a.
func RGBA128Prelerp(p, q, a float32) float32 { return p + q - p*a }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA128 blends a *plain* source into a premultiplied destination buffer.
// Matches AGG's blender_rgba (plain -> premultiplied).
type BlenderRGBA128[S color.Space, O order.RGBAOrder] struct{}

func (BlenderRGBA128[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	a *= cover
	if a <= 0 {
		return
	}
	var o O
	dst[o.IdxR()] = RGBA128Lerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = RGBA128Lerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = RGBA128Lerp(dst[o.IdxB()], b, a)
	dst[o.IdxA()] = RGBA128Prelerp(dst[o.IdxA()], a, a)
}

func (BlenderRGBA128[S, O]) SetPlain(dst []float32, r, g, b, a float32) {
	var o O
	dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = r, g, b, a
}

func (BlenderRGBA128[S, O]) GetPlain(src []float32) (r, g, b, a float32) {
	var o O
	return src[o.IdxR()], src[o.IdxG()], src[o.IdxB()], src[o.IdxA()]
}

func (BlenderRGBA128[S, O]) IdxR() int       { var o O; return o.IdxR() }
func (BlenderRGBA128[S, O]) IdxG() int       { var o O; return o.IdxG() }
func (BlenderRGBA128[S, O]) IdxB() int       { var o O; return o.IdxB() }
func (BlenderRGBA128[S, O]) IdxA() int       { var o O; return o.IdxA() }
func (BlenderRGBA128[S, O]) PremulSrc() bool { return false }

////////////////////////////////////////////////////////////////////////////////
// Premultiplied source -> Premultiplied destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA128Pre blends a *premultiplied* source into a premultiplied
// destination buffer. Matches AGG's blender_rgba_pre.
type BlenderRGBA128Pre[S color.Space, O order.RGBAOrder] struct{}

func (BlenderRGBA128Pre[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	if cover != 1.0 {
		r *= cover
		g *= cover
		b *= cover
		a *= cover
	}
	if a <= 0 && r <= 0 && g <= 0 && b <= 0 {
		return
	}
	var o O
	dst[o.IdxR()] = RGBA128Prelerp(dst[o.IdxR()], r, a)
	dst[o.IdxG()] = RGBA128Prelerp(dst[o.IdxG()], g, a)
	dst[o.IdxB()] = RGBA128Prelerp(dst[o.IdxB()], b, a)
	dst[o.IdxA()] = RGBA128Prelerp(dst[o.IdxA()], a, a)
}

func (BlenderRGBA128Pre[S, O]) SetPlain(dst []float32, r, g, b, a float32) {
	BlenderRGBA128[S, O]{}.SetPlain(dst, r, g, b, a)
}

func (BlenderRGBA128Pre[S, O]) GetPlain(src []float32) (r, g, b, a float32) {
	return BlenderRGBA128[S, O]{}.GetPlain(src)
}

func (BlenderRGBA128Pre[S, O]) IdxR() int       { var o O; return o.IdxR() }
func (BlenderRGBA128Pre[S, O]) IdxG() int       { var o O; return o.IdxG() }
func (BlenderRGBA128Pre[S, O]) IdxB() int       { var o O; return o.IdxB() }
func (BlenderRGBA128Pre[S, O]) IdxA() int       { var o O; return o.IdxA() }
func (BlenderRGBA128Pre[S, O]) PremulSrc() bool { return true }

////////////////////////////////////////////////////////////////////////////////
// Plain (non-premultiplied) source -> Plain destination
////////////////////////////////////////////////////////////////////////////////

// BlenderRGBA128Plain blends a *plain* source into a *plain* destination buffer.
// Matches AGG's blender_rgba_plain: premultiply the destination by its alpha,
// blend in premultiplied space, then demultiply for storage.
type BlenderRGBA128Plain[S color.Space, O order.RGBAOrder] struct{}

func (BlenderRGBA128Plain[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	a *= cover
	if a <= 0 {
		return
	}
	var o O
	da := dst[o.IdxA()]
	pr := dst[o.IdxR()] * da
	pg := dst[o.IdxG()] * da
	pb := dst[o.IdxB()] * da

	pr = RGBA128Lerp(pr, r, a)
	pg = RGBA128Lerp(pg, g, a)
	pb = RGBA128Lerp(pb, b, a)
	na := RGBA128Prelerp(da, a, a)

	dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()] = demultiplyRGBA128(pr, pg, pb, na)
	dst[o.IdxA()] = na
}

// demultiplyRGBA128 converts premultiplied r,g,b back to straight given alpha.
func demultiplyRGBA128(r, g, b, a float32) (float32, float32, float32) {
	if a <= 0 {
		return 0, 0, 0
	}
	if a >= 1.0 {
		return r, g, b
	}
	inv := 1.0 / a
	return r * inv, g * inv, b * inv
}

func (BlenderRGBA128Plain[S, O]) SetPlain(dst []float32, r, g, b, a float32) {
	BlenderRGBA128[S, O]{}.SetPlain(dst, r, g, b, a)
}

func (BlenderRGBA128Plain[S, O]) GetPlain(src []float32) (r, g, b, a float32) {
	return BlenderRGBA128[S, O]{}.GetPlain(src)
}

func (BlenderRGBA128Plain[S, O]) IdxR() int { var o O; return o.IdxR() }
func (BlenderRGBA128Plain[S, O]) IdxG() int { var o O; return o.IdxG() }
func (BlenderRGBA128Plain[S, O]) IdxB() int { var o O; return o.IdxB() }
func (BlenderRGBA128Plain[S, O]) IdxA() int { var o O; return o.IdxA() }

////////////////////////////////////////////////////////////////////////////////
// Single-pixel and span helpers over color.RGBA32
////////////////////////////////////////////////////////////////////////////////

// BlendRGBA128Pixel blends a single pixel using the provided blender.
func BlendRGBA128Pixel[S color.Space, O order.RGBAOrder](
	dst []float32,
	src color.RGBA32[S],
	cover float32,
	b RGBA128Blender[S],
) {
	if src.A <= 0 || cover <= 0 {
		return
	}
	b.BlendPix(dst, src.R, src.G, src.B, src.A, cover)
}

// CopyRGBA128Pixel writes plain RGBA components to dst in order O (no blending).
func CopyRGBA128Pixel[S color.Space, O order.RGBAOrder](dst []float32, src color.RGBA32[S]) {
	var o O
	dst[o.IdxR()] = src.R
	dst[o.IdxG()] = src.G
	dst[o.IdxB()] = src.B
	dst[o.IdxA()] = src.A
}

// BlendRGBA128Hline blends a horizontal run with optional per-pixel coverage
// (covers == nil means full cover).
func BlendRGBA128Hline[S color.Space, O order.RGBAOrder](
	dst []float32,
	x, length int,
	src color.RGBA32[S],
	covers []float32,
	b RGBA128Blender[S],
) {
	if length <= 0 || src.A <= 0 {
		return
	}
	const pixStep = 4
	p := x * pixStep
	if covers == nil {
		for range length {
			b.BlendPix(dst[p:p+4], src.R, src.G, src.B, src.A, 1.0)
			p += pixStep
		}
		return
	}
	for i := range length {
		if c := covers[i]; c > 0 {
			b.BlendPix(dst[p:p+4], src.R, src.G, src.B, src.A, c)
		}
		p += pixStep
	}
}

// CopyRGBA128Hline copies a horizontal run of one plain color into dst in order O.
func CopyRGBA128Hline[S color.Space, O order.RGBAOrder](
	dst []float32,
	x, length int,
	src color.RGBA32[S],
) {
	if length <= 0 {
		return
	}
	var o O
	const pixStep = 4
	p := x * pixStep
	for range length {
		dst[p+o.IdxR()] = src.R
		dst[p+o.IdxG()] = src.G
		dst[p+o.IdxB()] = src.B
		dst[p+o.IdxA()] = src.A
		p += pixStep
	}
}

// FillRGBA128Span is a synonym of CopyRGBA128Hline (explicit name for intent).
func FillRGBA128Span[S color.Space, O order.RGBAOrder](dst []float32, x, length int, src color.RGBA32[S]) {
	CopyRGBA128Hline[S, O](dst, x, length, src)
}

////////////////////////////////////////////////////////////////////////////////
// Convenience aliases (order/space), mirroring rgba8.go
////////////////////////////////////////////////////////////////////////////////

type (
	BlenderRGBA128Linear      = BlenderRGBA128[color.Linear, order.RGBA]
	BlenderRGBA128PreLinear   = BlenderRGBA128Pre[color.Linear, order.RGBA]
	BlenderRGBA128PlainLinear = BlenderRGBA128Plain[color.Linear, order.RGBA]

	BlenderRGBA128SRGB      = BlenderRGBA128[color.SRGB, order.RGBA]
	BlenderRGBA128PreSRGB   = BlenderRGBA128Pre[color.SRGB, order.RGBA]
	BlenderRGBA128PlainSRGB = BlenderRGBA128Plain[color.SRGB, order.RGBA]
)
