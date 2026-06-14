package blender

import (
	"github.com/cwbudde/agg_go/internal/basics"
	"github.com/cwbudde/agg_go/internal/color"
	"github.com/cwbudde/agg_go/internal/order"
)

// Float (128-bit, 4 x float32) composite blenders.
//
// These are the float twins of the 8-bit CompositeBlender / CompositeBlenderPre
// in rgba_composite.go, and the structural analogue of AGG's comp_op_adaptor_rgba
// instantiated with the float rgba32 color type. They evaluate the same
// Porter-Duff / SVG composite algebra in premultiplied space; the per-operator
// equations are shared verbatim with the 8-bit path via the unexported
// CompositeBlender.blendOperation, so the math cannot drift between the two.
//
// Channels live in [0,1] (no /255 normalization). Results are clamped to [0,1]
// for storage, matching the 8-bit path which saturates through its to8 helper.

// clampF01 clamps a float64 composite result to the [0,1] storage range and
// narrows it to float32, mirroring the 8-bit to8 saturation behavior.
func clampF01(v float64) float32 {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}
	return float32(v)
}

// CompositeBlenderRGBA128 composites a *straight*-alpha source into a
// premultiplied destination buffer (Dca/Da). It is the float analogue of the
// 8-bit CompositeBlender and matches AGG's comp_op_adaptor_rgba.
type CompositeBlenderRGBA128[S color.Space, O order.RGBAOrder] struct {
	op CompOp
}

func NewCompositeBlenderRGBA128[S color.Space, O order.RGBAOrder](op CompOp) CompositeBlenderRGBA128[S, O] {
	return CompositeBlenderRGBA128[S, O]{op: op}
}

func (bl CompositeBlenderRGBA128[S, O]) GetOp() CompOp { return bl.op }

// GetPlain reads a premultiplied pixel and returns straight (demultiplied) RGBA.
func (bl CompositeBlenderRGBA128[S, O]) GetPlain(px []float32) (r, g, b, a float32) {
	var o O
	a = px[o.IdxA()]
	if a <= 0 {
		return 0, 0, 0, 0
	}
	inv := float32(1) / a
	return px[o.IdxR()] * inv, px[o.IdxG()] * inv, px[o.IdxB()] * inv, a
}

// SetPlain stores straight RGBA as premultiplied components in channel order O.
func (bl CompositeBlenderRGBA128[S, O]) SetPlain(px []float32, r, g, b, a float32) {
	var o O
	px[o.IdxR()] = r * a
	px[o.IdxG()] = g * a
	px[o.IdxB()] = b * a
	px[o.IdxA()] = a
}

// BlendPix builds a premultiplied source from the straight (r,g,b,a) and the
// coverage, reads the destination as premultiplied (Dca/Da), evaluates the
// selected composite operator, and writes the premultiplied result back.
func (bl CompositeBlenderRGBA128[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	var o O

	sa := float64(a) * float64(cover)
	if sa <= 0 {
		return
	}

	s := normalizedRGBA{
		r: float64(r) * sa,
		g: float64(g) * sa,
		b: float64(b) * sa,
		a: sa,
	}
	d := normalizedRGBA{
		r: float64(dst[o.IdxR()]),
		g: float64(dst[o.IdxG()]),
		b: float64(dst[o.IdxB()]),
		a: float64(dst[o.IdxA()]),
	}

	res := CompositeBlender[S, O](bl).blendOperation(d, s)

	dst[o.IdxR()] = clampF01(res.r)
	dst[o.IdxG()] = clampF01(res.g)
	dst[o.IdxB()] = clampF01(res.b)
	dst[o.IdxA()] = clampF01(res.a)
}

// CompositeBlenderRGBA128Pre composites a *premultiplied* source into a
// premultiplied destination buffer. It is the float analogue of the 8-bit
// CompositeBlenderPre and matches AGG's comp_op_adaptor_rgba_pre.
type CompositeBlenderRGBA128Pre[S color.Space, O order.RGBAOrder] struct {
	op CompOp
}

func NewCompositeBlenderRGBA128Pre[S color.Space, O order.RGBAOrder](op CompOp) CompositeBlenderRGBA128Pre[S, O] {
	return CompositeBlenderRGBA128Pre[S, O]{op: op}
}

func (bl CompositeBlenderRGBA128Pre[S, O]) GetOp() CompOp { return bl.op }

func (bl CompositeBlenderRGBA128Pre[S, O]) GetPlain(px []float32) (r, g, b, a float32) {
	return CompositeBlenderRGBA128[S, O](bl).GetPlain(px)
}

func (bl CompositeBlenderRGBA128Pre[S, O]) SetPlain(px []float32, r, g, b, a float32) {
	CompositeBlenderRGBA128[S, O](bl).SetPlain(px, r, g, b, a)
}

func (bl CompositeBlenderRGBA128Pre[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
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
	d := normalizedRGBA{
		r: float64(dst[o.IdxR()]),
		g: float64(dst[o.IdxG()]),
		b: float64(dst[o.IdxB()]),
		a: float64(dst[o.IdxA()]),
	}
	s := normalizedRGBA{
		r: float64(r),
		g: float64(g),
		b: float64(b),
		a: float64(a),
	}

	res := CompositeBlender[S, O](bl).blendOperation(d, s)

	dst[o.IdxR()] = clampF01(res.r)
	dst[o.IdxG()] = clampF01(res.g)
	dst[o.IdxB()] = clampF01(res.b)
	dst[o.IdxA()] = clampF01(res.a)
}

// CompositeBlenderRGBA128Plain bridges AGG's premultiplied composite math to a
// *straight* (non-premultiplied) destination buffer — the storage convention
// used by the float Agg2D path (PixFmtRGBA128Plain). It is the float analogue of
// CompositeBlenderPlain: it premultiplies the destination on read, evaluates the
// operator in premultiplied space, then demultiplies the result back to straight
// alpha for storage. With an opaque destination (Da == 1) it is identical to
// CompositeBlenderRGBA128, since straight and premultiplied representations
// coincide there; the difference only matters when compositing over partially
// transparent destination content, where reading straight values as
// premultiplied would over-contribute the destination colour (a washed-out halo).
type CompositeBlenderRGBA128Plain[S color.Space, O order.RGBAOrder] struct {
	op CompOp
}

func NewCompositeBlenderRGBA128Plain[S color.Space, O order.RGBAOrder](op CompOp) CompositeBlenderRGBA128Plain[S, O] {
	return CompositeBlenderRGBA128Plain[S, O]{op: op}
}

func (bl CompositeBlenderRGBA128Plain[S, O]) GetOp() CompOp { return bl.op }

// GetPlain/SetPlain operate on a straight-alpha buffer, so they pass through.
func (bl CompositeBlenderRGBA128Plain[S, O]) GetPlain(px []float32) (r, g, b, a float32) {
	var o O
	return px[o.IdxR()], px[o.IdxG()], px[o.IdxB()], px[o.IdxA()]
}

func (bl CompositeBlenderRGBA128Plain[S, O]) SetPlain(px []float32, r, g, b, a float32) {
	var o O
	px[o.IdxR()], px[o.IdxG()], px[o.IdxB()], px[o.IdxA()] = r, g, b, a
}

func (bl CompositeBlenderRGBA128Plain[S, O]) BlendPix(dst []float32, r, g, b, a, cover float32) {
	var o O

	sa := float64(a) * float64(cover)
	if sa <= 0 {
		return
	}

	// Premultiplied source.
	s := normalizedRGBA{
		r: float64(r) * sa,
		g: float64(g) * sa,
		b: float64(b) * sa,
		a: sa,
	}

	// Straight destination -> premultiply (Dca/Da) for the composite math.
	da := float64(dst[o.IdxA()])
	d := normalizedRGBA{
		r: float64(dst[o.IdxR()]) * da,
		g: float64(dst[o.IdxG()]) * da,
		b: float64(dst[o.IdxB()]) * da,
		a: da,
	}

	res := CompositeBlender[S, O](bl).blendOperation(d, s)

	// Demultiply the premultiplied result back to straight alpha for storage.
	if res.a <= 0 {
		dst[o.IdxR()], dst[o.IdxG()], dst[o.IdxB()], dst[o.IdxA()] = 0, 0, 0, 0
		return
	}
	inv := 1.0 / res.a
	dst[o.IdxR()] = clampF01(res.r * inv)
	dst[o.IdxG()] = clampF01(res.g * inv)
	dst[o.IdxB()] = clampF01(res.b * inv)
	dst[o.IdxA()] = clampF01(res.a)
}

// BlendSolidSpanStraight applies the active operator across count straight-alpha
// destination pixels in one call — the float twin of
// CompositeBlenderPlain.BlendSolidSpanStraight. It performs the identical
// premultiply-on-read / op / demultiply-on-write bridge as BlendPix, so its
// output is bit-for-bit identical to calling BlendPix per pixel; the win is one
// concrete call per span instead of one interface dispatch per pixel (the comp
// pixfmt holds the blender behind an interface). dst is packed in channel order O
// (4 float32 per pixel); covers may be nil (uniform full coverage) or hold one
// basics.Int8u cover per pixel, normalised exactly as the pixfmt's coverToF32
// (float32(cover)/255) before widening, matching the per-pixel BlendPix calls.
func (bl CompositeBlenderRGBA128Plain[S, O]) BlendSolidSpanStraight(dst []float32, r, g, b, a float32, covers []basics.Int8u, count int) {
	var o O
	cb := CompositeBlender[S, O](bl)
	ir, ig, ib, ia := o.IdxR(), o.IdxG(), o.IdxB(), o.IdxA()
	rf := float64(r)
	gf := float64(g)
	bf := float64(b)
	af := float64(a)
	for i := 0; i < count; i++ {
		// Coverage, matching BlendPix(coverToF32(covers[i])): float32(cover)/255
		// computed in float32, then widened to float64.
		cover := 1.0
		if covers != nil {
			cover = float64(float32(covers[i]) / 255.0)
		}
		sa := af * cover
		if sa <= 0 {
			continue
		}
		p := i * 4
		// Premultiplied source.
		s := normalizedRGBA{r: rf * sa, g: gf * sa, b: bf * sa, a: sa}
		// Straight destination -> premultiplied.
		da := float64(dst[p+ia])
		d := normalizedRGBA{
			r: float64(dst[p+ir]) * da,
			g: float64(dst[p+ig]) * da,
			b: float64(dst[p+ib]) * da,
			a: da,
		}
		res := cb.blendOperation(d, s)
		// Demultiply back to straight alpha for storage.
		if res.a <= 0 {
			dst[p+ir], dst[p+ig], dst[p+ib], dst[p+ia] = 0, 0, 0, 0
			continue
		}
		inv := 1.0 / res.a
		dst[p+ir] = clampF01(res.r * inv)
		dst[p+ig] = clampF01(res.g * inv)
		dst[p+ib] = clampF01(res.b * inv)
		dst[p+ia] = clampF01(res.a)
	}
}

// Convenience aliases (Linear space, RGBA order), mirroring the 8-bit set.
type (
	CompositeBlenderRGBA128Linear      = CompositeBlenderRGBA128[color.Linear, order.RGBA]
	CompositeBlenderRGBA128PreLinear   = CompositeBlenderRGBA128Pre[color.Linear, order.RGBA]
	CompositeBlenderRGBA128PlainLinear = CompositeBlenderRGBA128Plain[color.Linear, order.RGBA]
	CompositeBlenderRGBA128SRGB        = CompositeBlenderRGBA128[color.SRGB, order.RGBA]
	CompositeBlenderRGBA128PreSRGB     = CompositeBlenderRGBA128Pre[color.SRGB, order.RGBA]
	CompositeBlenderRGBA128PlainSRGB   = CompositeBlenderRGBA128Plain[color.SRGB, order.RGBA]
)
