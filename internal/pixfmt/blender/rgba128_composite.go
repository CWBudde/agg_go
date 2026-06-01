package blender

import (
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

// Convenience aliases (Linear space, RGBA order), mirroring the 8-bit set.
type (
	CompositeBlenderRGBA128Linear    = CompositeBlenderRGBA128[color.Linear, order.RGBA]
	CompositeBlenderRGBA128PreLinear = CompositeBlenderRGBA128Pre[color.Linear, order.RGBA]
	CompositeBlenderRGBA128SRGB      = CompositeBlenderRGBA128[color.SRGB, order.RGBA]
	CompositeBlenderRGBA128PreSRGB   = CompositeBlenderRGBA128Pre[color.SRGB, order.RGBA]
)
