package blender

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/order"
)

// Composite blend operation types (SVG/Porter-Duff + extras)
type CompOp int

const (
	CompOpClear CompOp = iota
	CompOpSrc
	CompOpDst
	CompOpSrcOver
	CompOpDstOver
	CompOpSrcIn
	CompOpDstIn
	CompOpSrcOut
	CompOpDstOut
	CompOpSrcAtop
	CompOpDstAtop
	CompOpXor
	CompOpPlus
	CompOpMultiply
	CompOpScreen
	CompOpOverlay
	CompOpDarken
	CompOpLighten
	CompOpColorDodge
	CompOpColorBurn
	CompOpHardLight
	CompOpSoftLight
	CompOpDifference
	CompOpExclusion
)

// CompositeBlender operates in premultiplied space (Sca/Dca, Sa/Da).
// S is the color space tag, O is the byte/channel order (compile-time).
type CompositeBlender[S color.Space, O order.RGBAOrder] struct {
	op CompOp
}

func NewCompositeBlender[S color.Space, O order.RGBAOrder](op CompOp) CompositeBlender[S, O] {
	return CompositeBlender[S, O]{op: op}
}

func (bl CompositeBlender[S, O]) GetOp() CompOp { return bl.op }

// BlendPix:
//   - Interprets dst as *premultiplied* RGBA (Dca, Da) in order O
//   - Builds premultiplied source Sca/Sa from (r,g,b,a) and coverage
//   - Evaluates the chosen composite op in premultiplied algebra
//   - Writes premultiplied result back to dst
func (bl CompositeBlender[S, O]) BlendPix(dst []basics.Int8u, r, g, b, a, cover basics.Int8u) {
	var o O

	// Sa with coverage in [0,1]
	sa := float64(color.RGBA8MultCover(a, cover)) / 255.0
	if sa <= 0 {
		return
	}

	// Sca (premultiplied source)
	s := normalizedRGBA{
		r: (float64(r) / 255.0) * sa,
		g: (float64(g) / 255.0) * sa,
		b: (float64(b) / 255.0) * sa,
		a: sa,
	}

	// Dca/Da (premultiplied destination), read as-is
	d := normalizedRGBA{
		r: float64(dst[o.IdxR()]) / 255.0,
		g: float64(dst[o.IdxG()]) / 255.0,
		b: float64(dst[o.IdxB()]) / 255.0,
		a: float64(dst[o.IdxA()]) / 255.0,
	}

	res := bl.blendOperation(d, s)

	// Store premultiplied result
	dst[o.IdxR()] = to8(res.r)
	dst[o.IdxG()] = to8(res.g)
	dst[o.IdxB()] = to8(res.b)
	dst[o.IdxA()] = to8(res.a)
}

// normalizedRGBA holds premultiplied color components in [0,1]
type normalizedRGBA struct{ r, g, b, a float64 }

// Utility: clamp [0,1] -> uint8 with rounding
func to8(v float64) basics.Int8u {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 255
	}
	return basics.Int8u(v*255.0 + 0.5)
}

// Select the composite equation
func (bl CompositeBlender[S, O]) blendOperation(d, s normalizedRGBA) normalizedRGBA {
	switch bl.op {
	case CompOpClear:
		return bl.clear(d, s)
	case CompOpSrc:
		return bl.src(d, s)
	case CompOpDst:
		return bl.dst(d, s)
	case CompOpSrcOver:
		return bl.sourceOver(d, s)
	case CompOpDstOver:
		return bl.dstOver(d, s)
	case CompOpSrcIn:
		return bl.srcIn(d, s)
	case CompOpDstIn:
		return bl.dstIn(d, s)
	case CompOpSrcOut:
		return bl.srcOut(d, s)
	case CompOpDstOut:
		return bl.dstOut(d, s)
	case CompOpSrcAtop:
		return bl.srcAtop(d, s)
	case CompOpDstAtop:
		return bl.dstAtop(d, s)
	case CompOpXor:
		return bl.xor(d, s)
	case CompOpPlus:
		return bl.plus(d, s)
	case CompOpMultiply:
		return bl.multiply(d, s)
	case CompOpScreen:
		return bl.screen(d, s)
	case CompOpOverlay:
		return bl.overlay(d, s)
	case CompOpDarken:
		return bl.darken(d, s)
	case CompOpLighten:
		return bl.lighten(d, s)
	case CompOpColorDodge:
		return bl.colorDodge(d, s)
	case CompOpColorBurn:
		return bl.colorBurn(d, s)
	case CompOpHardLight:
		return bl.hardLight(d, s)
	case CompOpSoftLight:
		return bl.softLight(d, s)
	case CompOpDifference:
		return bl.difference(d, s)
	case CompOpExclusion:
		return bl.exclusion(d, s)
	default:
		return bl.sourceOver(d, s)
	}
}

// Porter–Duff/SVG ops (all in premultiplied space)

// clear: D' = 0
func (bl CompositeBlender[S, O]) clear(d, s normalizedRGBA) normalizedRGBA {
	return normalizedRGBA{}
}

// src: D' = S
func (bl CompositeBlender[S, O]) src(d, s normalizedRGBA) normalizedRGBA {
	return s
}

// dst: D' = D
func (bl CompositeBlender[S, O]) dst(d, s normalizedRGBA) normalizedRGBA {
	return d
}

// src-over: Dca' = Sca + Dca(1 - Sa); Da' = Sa + Da(1 - Sa)
func (bl CompositeBlender[S, O]) sourceOver(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	is1 := 1.0 - s.a
	return normalizedRGBA{
		r: s.r + d.r*is1,
		g: s.g + d.g*is1,
		b: s.b + d.b*is1,
		a: s.a + d.a*is1,
	}
}

// dst-over: Dca' = Dca + Sca(1 - Da); Da' = Da + Sa(1 - Da)
func (bl CompositeBlender[S, O]) dstOver(d, s normalizedRGBA) normalizedRGBA {
	if d.a >= 1.0 {
		return d
	}
	id1 := 1.0 - d.a
	return normalizedRGBA{
		r: d.r + s.r*id1,
		g: d.g + s.g*id1,
		b: d.b + s.b*id1,
		a: d.a + s.a*id1,
	}
}

// src-in: Dca' = Sca*Da; Da' = Sa*Da
func (bl CompositeBlender[S, O]) srcIn(d, s normalizedRGBA) normalizedRGBA {
	return normalizedRGBA{r: s.r * d.a, g: s.g * d.a, b: s.b * d.a, a: s.a * d.a}
}

// dst-in: Dca' = Dca*Sa; Da' = Da*Sa
func (bl CompositeBlender[S, O]) dstIn(d, s normalizedRGBA) normalizedRGBA {
	return normalizedRGBA{r: d.r * s.a, g: d.g * s.a, b: d.b * s.a, a: d.a * s.a}
}

// src-out: Dca' = Sca(1 - Da); Da' = Sa(1 - Da)
func (bl CompositeBlender[S, O]) srcOut(d, s normalizedRGBA) normalizedRGBA {
	id := 1.0 - d.a
	return normalizedRGBA{r: s.r * id, g: s.g * id, b: s.b * id, a: s.a * id}
}

// dst-out: Dca' = Dca(1 - Sa); Da' = Da(1 - Sa)
func (bl CompositeBlender[S, O]) dstOut(d, s normalizedRGBA) normalizedRGBA {
	is := 1.0 - s.a
	return normalizedRGBA{r: d.r * is, g: d.g * is, b: d.b * is, a: d.a * is}
}

// src-atop: Dca' = Sca*Da + Dca(1 - Sa); Da' = Da
func (bl CompositeBlender[S, O]) srcAtop(d, s normalizedRGBA) normalizedRGBA {
	is := 1.0 - s.a
	return normalizedRGBA{
		r: s.r*d.a + d.r*is,
		g: s.g*d.a + d.g*is,
		b: s.b*d.a + d.b*is,
		a: d.a,
	}
}

// dst-atop: Dca' = Dca*Sa + Sca(1 - Da); Da' = Sa
func (bl CompositeBlender[S, O]) dstAtop(d, s normalizedRGBA) normalizedRGBA {
	id := 1.0 - d.a
	return normalizedRGBA{
		r: d.r*s.a + s.r*id,
		g: d.g*s.a + s.g*id,
		b: d.b*s.a + s.b*id,
		a: s.a,
	}
}

// xor: Dca' = Sca(1 - Da) + Dca(1 - Sa); Da' = Sa + Da - 2SaDa
func (bl CompositeBlender[S, O]) xor(d, s normalizedRGBA) normalizedRGBA {
	is := 1.0 - s.a
	id := 1.0 - d.a
	return normalizedRGBA{
		r: s.r*id + d.r*is,
		g: s.g*id + d.g*is,
		b: s.b*id + d.b*is,
		a: s.a + d.a - 2*s.a*d.a,
	}
}

// plus (linear dodge): Dca' = Sca + Dca; Da' = Sa + Da - SaDa
func (bl CompositeBlender[S, O]) plus(d, s normalizedRGBA) normalizedRGBA {
	return normalizedRGBA{
		r: d.r + s.r,
		g: d.g + s.g,
		b: d.b + s.b,
		a: s.a + d.a - s.a*d.a,
	}
}

// multiply: Dca' = ScaDca + Sca(1 - Da) + Dca(1 - Sa); Da' = Sa + Da - SaDa
func (bl CompositeBlender[S, O]) multiply(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	is := 1.0 - s.a
	id := 1.0 - d.a
	return normalizedRGBA{
		r: s.r*d.r + s.r*id + d.r*is,
		g: s.g*d.g + s.g*id + d.g*is,
		b: s.b*d.b + s.b*id + d.b*is,
		a: d.a + s.a - s.a*d.a,
	}
}

// screen: Dca' = Sca + Dca - ScaDca; Da' = Sa + Da - SaDa
func (bl CompositeBlender[S, O]) screen(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	return normalizedRGBA{
		r: s.r + d.r - s.r*d.r,
		g: s.g + d.g - s.g*d.g,
		b: s.b + d.b - s.b*d.b,
		a: d.a + s.a - s.a*d.a,
	}
}

// overlay
func (bl CompositeBlender[S, O]) overlay(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	id := 1.0 - d.a
	is := 1.0 - s.a
	sada := s.a * d.a

	calc := func(dca, sca, da, sa, sada, id, is float64) float64 {
		if 2*dca <= da {
			return 2*sca*dca + sca*id + dca*is
		}
		return sada - 2*(da-dca)*(sa-sca) + sca*id + dca*is
	}

	return normalizedRGBA{
		r: calc(d.r, s.r, d.a, s.a, sada, id, is),
		g: calc(d.g, s.g, d.a, s.a, sada, id, is),
		b: calc(d.b, s.b, d.a, s.a, sada, id, is),
		a: d.a + s.a - s.a*d.a,
	}
}

// darken
func (bl CompositeBlender[S, O]) darken(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	id := 1.0 - d.a
	is := 1.0 - s.a
	return normalizedRGBA{
		r: math.Min(s.r*d.a, d.r*s.a) + s.r*id + d.r*is,
		g: math.Min(s.g*d.a, d.g*s.a) + s.g*id + d.g*is,
		b: math.Min(s.b*d.a, d.b*s.a) + s.b*id + d.b*is,
		a: d.a + s.a - s.a*d.a,
	}
}

// lighten
func (bl CompositeBlender[S, O]) lighten(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	id := 1.0 - d.a
	is := 1.0 - s.a
	return normalizedRGBA{
		r: math.Max(s.r*d.a, d.r*s.a) + s.r*id + d.r*is,
		g: math.Max(s.g*d.a, d.g*s.a) + s.g*id + d.g*is,
		b: math.Max(s.b*d.a, d.b*s.a) + s.b*id + d.b*is,
		a: d.a + s.a - s.a*d.a,
	}
}

// color-dodge
func (bl CompositeBlender[S, O]) colorDodge(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	if d.a <= 0 {
		// Sca*(1 - Da)
		id := 1.0 - d.a
		return normalizedRGBA{r: s.r * id, g: s.g * id, b: s.b * id, a: s.a}
	}
	sada := s.a * d.a
	is := 1.0 - s.a
	id := 1.0 - d.a

	calc := func(dca, sca, da, sa, sada, id, is float64) float64 {
		if sca < sa {
			return sada*math.Min(1.0, (dca/da)*sa/(sa-sca)) + sca*id + dca*is
		}
		if dca > 0 {
			return sada + sca*id + dca*is
		}
		return sca * id
	}

	return normalizedRGBA{
		r: calc(d.r, s.r, d.a, s.a, sada, id, is),
		g: calc(d.g, s.g, d.a, s.a, sada, id, is),
		b: calc(d.b, s.b, d.a, s.a, sada, id, is),
		a: d.a + s.a - s.a*d.a,
	}
}

// color-burn
func (bl CompositeBlender[S, O]) colorBurn(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	if d.a <= 0 {
		id := 1.0 - d.a
		return normalizedRGBA{r: s.r * id, g: s.g * id, b: s.b * id, a: s.a}
	}
	sada := s.a * d.a
	is := 1.0 - s.a
	id := 1.0 - d.a

	calc := func(dca, sca, da, sa, sada, id, is float64) float64 {
		if sca > 0 {
			return sada*(1.0-math.Min(1.0, (1.0-dca/da)*sa/sca)) + sca*id + dca*is
		}
		if dca > da {
			return sada + dca*is
		}
		return dca * is
	}

	return normalizedRGBA{
		r: calc(d.r, s.r, d.a, s.a, sada, id, is),
		g: calc(d.g, s.g, d.a, s.a, sada, id, is),
		b: calc(d.b, s.b, d.a, s.a, sada, id, is),
		a: d.a + s.a - s.a*d.a,
	}
}

// hard-light (overlay with src/dst swapped)
func (bl CompositeBlender[S, O]) hardLight(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	id := 1.0 - d.a
	is := 1.0 - s.a
	sada := s.a * d.a

	calc := func(dca, sca, da, sa, sada, id, is float64) float64 {
		if 2*sca <= sa {
			return 2*sca*dca + sca*id + dca*is
		}
		return sada - 2*(da-dca)*(sa-sca) + sca*id + dca*is
	}

	return normalizedRGBA{
		r: calc(d.r, s.r, d.a, s.a, sada, id, is),
		g: calc(d.g, s.g, d.a, s.a, sada, id, is),
		b: calc(d.b, s.b, d.a, s.a, sada, id, is),
		a: d.a + s.a - s.a*d.a,
	}
}

// soft-light (matches AGG’s form)
func (bl CompositeBlender[S, O]) softLight(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	if d.a <= 0 {
		id := 1.0 - d.a
		return normalizedRGBA{r: s.r * id, g: s.g * id, b: s.b * id, a: s.a}
	}

	sada := s.a * d.a
	is := 1.0 - s.a
	id := 1.0 - d.a

	calc := func(dca, sca, da, sa, sada, id, is float64) float64 {
		dcasa := dca * sa
		if 2*sca <= sa {
			return dcasa - (sada-2*sca*da)*dcasa*(sada-dcasa) + sca*id + dca*is
		}
		if 4*dca <= da {
			return dcasa + (2*sca*da-sada)*((((16*dcasa-12)*dcasa+4)*dca*da)-dca*da) + sca*id + dca*is
		}
		return dcasa + (2*sca*da-sada)*(math.Sqrt(dcasa)-dcasa) + sca*id + dca*is
	}

	return normalizedRGBA{
		r: calc(d.r, s.r, d.a, s.a, sada, id, is),
		g: calc(d.g, s.g, d.a, s.a, sada, id, is),
		b: calc(d.b, s.b, d.a, s.a, sada, id, is),
		a: s.a + d.a - sada,
	}
}

// difference: Dca' = Sca + Dca - 2 min(Sca Da, Dca Sa)
func (bl CompositeBlender[S, O]) difference(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	return normalizedRGBA{
		r: s.r + d.r - 2*math.Min(s.r*d.a, d.r*s.a),
		g: s.g + d.g - 2*math.Min(s.g*d.a, d.g*s.a),
		b: s.b + d.b - 2*math.Min(s.b*d.a, d.b*s.a),
		a: s.a + d.a - s.a*d.a,
	}
}

// exclusion
func (bl CompositeBlender[S, O]) exclusion(d, s normalizedRGBA) normalizedRGBA {
	if s.a <= 0 {
		return d
	}
	id := 1.0 - d.a
	is := 1.0 - s.a
	return normalizedRGBA{
		r: s.r*d.a + d.r*s.a - 2*s.r*d.r + s.r*id + d.r*is,
		g: s.g*d.a + d.g*s.a - 2*s.g*d.g + s.g*id + d.g*is,
		b: s.b*d.a + d.b*s.a - 2*s.b*d.b + s.b*id + d.b*is,
		a: d.a + s.a - s.a*d.a,
	}
}

// ---------- Helpers / convenience ----------

// Common aliases (Linear space, different orders)
type (
	CompositeBlenderRGBA8LinearRGBA = CompositeBlender[color.Linear, order.RGBA]
	CompositeBlenderRGBA8LinearBGRA = CompositeBlender[color.Linear, order.BGRA]
	CompositeBlenderRGBA8LinearARGB = CompositeBlender[color.Linear, order.ARGB]
	CompositeBlenderRGBA8LinearABGR = CompositeBlender[color.Linear, order.ABGR]
)

// Helper constructors
func NewMultiplyBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpMultiply)
}
func NewScreenBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpScreen)
}
func NewOverlayBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpOverlay)
}
func NewDarkenBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDarken)
}
func NewLightenBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpLighten)
}
func NewColorDodgeBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpColorDodge)
}
func NewColorBurnBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpColorBurn)
}
func NewHardLightBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpHardLight)
}
func NewSoftLightBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSoftLight)
}
func NewDifferenceBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDifference)
}
func NewExclusionBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpExclusion)
}
func NewPlusBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpPlus)
}
func NewClearBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpClear)
}
func NewSrcBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSrc)
}
func NewDstBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDst)
}
func NewSrcOverBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSrcOver)
}
func NewDstOverBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDstOver)
}
func NewSrcInBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSrcIn)
}
func NewDstInBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDstIn)
}
func NewSrcOutBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSrcOut)
}
func NewDstOutBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDstOut)
}
func NewSrcAtopBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpSrcAtop)
}
func NewDstAtopBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpDstAtop)
}

func NewXorBlender[S color.Space, O order.RGBAOrder]() CompositeBlender[S, O] {
	return NewCompositeBlender[S, O](CompOpXor)
}

// BlendColorHspan blends a horizontal span (with stride in bytes).
// dst points to the beginning of the *whole image* buffer.
func (bl CompositeBlender[S, O]) BlendColorHspan(
	dst []basics.Int8u, x, y, length, stride int,
	r, g, b, a basics.Int8u, covers []basics.Int8u,
) {
	if length <= 0 {
		return
	}
	const pix = 4
	offset := y*stride + x*pix
	for i := 0; i < length; i++ {
		c := basics.Int8u(255)
		if covers != nil && i < len(covers) {
			c = covers[i]
		}
		if c != 0 {
			bl.BlendPix(dst[offset:], r, g, b, a, c)
		}
		offset += pix
	}
}

// BlendColorVspan blends a vertical span (with stride in bytes).
func (bl CompositeBlender[S, O]) BlendColorVspan(
	dst []basics.Int8u, x, y, length, stride int,
	r, g, b, a basics.Int8u, covers []basics.Int8u,
) {
	if length <= 0 {
		return
	}
	const pix = 4
	offset := y*stride + x*pix
	for i := 0; i < length; i++ {
		c := basics.Int8u(255)
		if covers != nil && i < len(covers) {
			c = covers[i]
		}
		if c != 0 {
			bl.BlendPix(dst[offset:], r, g, b, a, c)
		}
		offset += stride
	}
}
