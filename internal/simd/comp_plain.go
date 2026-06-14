package simd

// Straight-alpha composite span kernels ("plain").
//
// Unlike the Comp*HspanRGBA kernels in cpu.go — which operate on a
// *premultiplied* destination and leave a premultiplied result — these operate
// on a *straight*-alpha destination, exactly like blender.CompositeBlenderPlain:
// premultiply the destination on read, evaluate the operator in premultiplied
// space, then demultiply the result back to straight alpha for storage. They are
// the fast path for the straight composite pixfmt (internal/pixfmt) and must stay
// bit-for-bit identical to CompositeBlenderPlain[Linear, RGBA] — locked by the
// differential test in comp_plain_test.go.
//
// dst is packed RGBA in R,G,B,A byte order; src (r,g,b,a) is straight alpha;
// covers gives per-pixel coverage (nil → uniform full coverage = 255). The
// per-pixel coverage multiply uses rgba8Multiply, which is bit-identical to
// color.RGBA8MultCover, so the source-alpha computation matches the scalar path
// exactly. The float math mirrors CompositeBlenderPlain.BlendPix and the
// per-operator equations in blender/rgba_composite.go verbatim.
//
// This is currently a scalar (non-vectorised) tier registered as the generic
// implementation; an AVX2/SSE tier can replace the inner loop later without
// changing callers (PLAN.md "Composite SIMD Fast Path").

// to8Plain clamps a [0,1] float to uint8 with round-half-up, matching blender.to8.
func to8Plain(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 255
	}
	return uint8(v*255.0 + 0.5)
}

// srcAlpha returns the straight source alpha in [0,1] for pixel i, applying
// coverage with the AGG multiply (bit-identical to color.RGBA8MultCover).
func srcAlpha(a uint8, covers []byte, i int) float64 {
	if covers != nil {
		return float64(rgba8Multiply(a, covers[i])) / 255.0
	}
	return float64(a) / 255.0
}

// CompSrcOverPlainHspanRGBA blends a straight-alpha source over a straight-alpha
// destination using Porter-Duff SrcOver, storing a straight-alpha result.
func CompSrcOverPlainHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	if maxPixels := len(dst) / 4; count > maxPixels {
		count = maxPixels
	}
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0
	for i := 0; i < count; i++ {
		sa := srcAlpha(a, covers, i)
		if sa <= 0 {
			continue
		}
		p := i * 4
		// Premultiplied source.
		scr, scg, scb := rf*sa, gf*sa, bf*sa
		// Straight destination -> premultiplied.
		da := float64(dst[p+3]) / 255.0
		dcr := (float64(dst[p+0]) / 255.0) * da
		dcg := (float64(dst[p+1]) / 255.0) * da
		dcb := (float64(dst[p+2]) / 255.0) * da
		// src-over: Dca' = Sca + Dca(1-Sa); Da' = Sa + Da(1-Sa).
		is1 := 1.0 - sa
		rr := scr + dcr*is1
		rg := scg + dcg*is1
		rb := scb + dcb*is1
		ra := sa + da*is1
		if ra <= 0 {
			dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
			continue
		}
		dst[p+0] = to8Plain(rr / ra)
		dst[p+1] = to8Plain(rg / ra)
		dst[p+2] = to8Plain(rb / ra)
		dst[p+3] = to8Plain(ra)
	}
}

// CompXorPlainHspanRGBA blends a straight-alpha source with a straight-alpha
// destination using Porter-Duff Xor, storing a straight-alpha result.
func CompXorPlainHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	if maxPixels := len(dst) / 4; count > maxPixels {
		count = maxPixels
	}
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0
	for i := 0; i < count; i++ {
		sa := srcAlpha(a, covers, i)
		if sa <= 0 {
			continue
		}
		p := i * 4
		// Premultiplied source.
		scr, scg, scb := rf*sa, gf*sa, bf*sa
		// Straight destination -> premultiplied.
		da := float64(dst[p+3]) / 255.0
		dcr := (float64(dst[p+0]) / 255.0) * da
		dcg := (float64(dst[p+1]) / 255.0) * da
		dcb := (float64(dst[p+2]) / 255.0) * da
		// xor: Dca' = Sca(1-Da) + Dca(1-Sa); Da' = Sa + Da - 2·Sa·Da.
		is := 1.0 - sa
		id := 1.0 - da
		rr := scr*id + dcr*is
		rg := scg*id + dcg*is
		rb := scb*id + dcb*is
		ra := sa + da - 2*sa*da
		if ra <= 0 {
			dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
			continue
		}
		dst[p+0] = to8Plain(rr / ra)
		dst[p+1] = to8Plain(rg / ra)
		dst[p+2] = to8Plain(rb / ra)
		dst[p+3] = to8Plain(ra)
	}
}
