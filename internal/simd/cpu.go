package simd

import (
	"sync"
)

// Features describes the CPU capabilities relevant to SIMD dispatch.
type Features struct {
	HasSSE    bool
	HasSSE2   bool
	HasSSE3   bool
	HasSSSE3  bool
	HasSSE41  bool
	HasAVX    bool
	HasAVX2   bool
	HasAVX512 bool
	HasNEON   bool

	ForceGeneric bool
	Architecture string
}

type implementation struct {
	name                 string
	fillRGBA             func(dst []byte, r, g, b, a uint8, count int)
	blendSolidHspanRGBA  func(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool)
	blendHlineRGBA       func(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool)
	blendColorHspanRGBA  func(dst, srcColors, covers []byte, count int, premulSrc bool)
	premultiplyRGBA      func(buf []byte, count int)
	demultiplyRGBA       func(buf []byte, count int)
	compSrcOverHspanRGBA func(dst, covers []byte, r, g, b, a uint8, count int)
}

var (
	detectMutex sync.Mutex
	detectOnce  sync.Once
	detected    Features

	forcedMutex    sync.RWMutex
	forcedFeatures *Features

	implMutex  sync.Mutex
	implOnce   sync.Once
	activeImpl implementation
)

// DetectFeatures returns the detected CPU features, cached across calls.
func DetectFeatures() Features {
	forcedMutex.RLock()
	forced := forcedFeatures
	forcedMutex.RUnlock()
	if forced != nil {
		return *forced
	}

	detectMutex.Lock()
	detectOnce.Do(func() {
		detected = detectFeaturesImpl()
	})
	features := detected
	detectMutex.Unlock()

	return features
}

// SetForcedFeatures overrides hardware detection for tests.
func SetForcedFeatures(f Features) {
	forcedMutex.Lock()
	forced := f
	forcedFeatures = &forced
	forcedMutex.Unlock()
	resetImplementation()
}

// ResetDetection clears forced features and cached detection/dispatch state.
func ResetDetection() {
	forcedMutex.Lock()
	forcedFeatures = nil
	forcedMutex.Unlock()

	detectMutex.Lock()
	detectOnce = sync.Once{}
	detected = Features{}
	detectMutex.Unlock()

	resetImplementation()
}

// ImplementationName returns the currently selected fill implementation name.
func ImplementationName() string {
	return currentImplementation().name
}

// FillRGBA fills count tightly-packed 4-byte pixels in dst with the provided byte pattern.
func FillRGBA(dst []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}

	maxCount := len(dst) / 4
	if count > maxCount {
		count = maxCount
	}
	if count <= 0 {
		return
	}

	currentImplementation().fillRGBA(dst, r, g, b, a, count)
}

func currentImplementation() implementation {
	implMutex.Lock()
	implOnce.Do(func() {
		activeImpl = selectImplementation(DetectFeatures())
	})
	impl := activeImpl
	implMutex.Unlock()
	return impl
}

func resetImplementation() {
	implMutex.Lock()
	implOnce = sync.Once{}
	activeImpl = implementation{}
	implMutex.Unlock()
}

func selectImplementation(features Features) implementation {
	return selectImplementationArch(features)
}

func genericImplementation() implementation {
	return implementation{
		name:                 "generic",
		fillRGBA:             fillRGBAGeneric,
		blendSolidHspanRGBA:  blendSolidHspanRGBAGeneric,
		blendHlineRGBA:       blendHlineRGBAGeneric,
		blendColorHspanRGBA:  blendColorHspanRGBAGeneric,
		premultiplyRGBA:      premultiplyRGBAGeneric,
		demultiplyRGBA:       demultiplyRGBAGeneric,
		compSrcOverHspanRGBA: compSrcOverHspanRGBAGeneric,
	}
}

func fillRGBAGeneric(dst []byte, r, g, b, a uint8, count int) {
	// Write the first pixel.
	dst[0] = r
	dst[1] = g
	dst[2] = b
	dst[3] = a

	// Use copy() doubling to fill the rest — the compiler maps copy() to
	// memcpy/memmove which is significantly faster than a scalar loop for
	// large spans.
	totalBytes := count * 4
	filled := 4
	for filled < totalBytes {
		filled += copy(dst[filled:totalBytes], dst[:filled])
	}
}

// BlendSolidHspanRGBA blends a solid color into tightly packed RGBA pixels.
// dst must contain one 4-byte pixel per cover entry.
func BlendSolidHspanRGBA(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if len(covers) == 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if len(covers) > maxPixels {
		covers = covers[:maxPixels]
	}
	currentImplementation().blendSolidHspanRGBA(dst, covers, r, g, b, a, premulSrc)
}

func blendSolidHspanRGBAGeneric(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	for i, cv := range covers {
		if cv == 0 {
			continue
		}
		p := i * 4
		if a == 0 && !premulSrc {
			continue
		}
		if a == 255 && cv == 255 {
			dst[p+0] = r
			dst[p+1] = g
			dst[p+2] = b
			dst[p+3] = a
			continue
		}

		if premulSrc {
			rr, gg, bb, aa := r, g, b, a
			if cv != 255 {
				rr = rgba8Multiply(rr, cv)
				gg = rgba8Multiply(gg, cv)
				bb = rgba8Multiply(bb, cv)
				aa = rgba8Multiply(aa, cv)
			}
			if aa == 0 && rr == 0 && gg == 0 && bb == 0 {
				continue
			}
			dst[p+0] = rgba8Prelerp(dst[p+0], rr, aa)
			dst[p+1] = rgba8Prelerp(dst[p+1], gg, aa)
			dst[p+2] = rgba8Prelerp(dst[p+2], bb, aa)
			dst[p+3] = rgba8Prelerp(dst[p+3], aa, aa)
			continue
		}

		alpha := rgba8Multiply(a, cv)
		if alpha == 0 {
			continue
		}
		dst[p+0] = rgba8Lerp(dst[p+0], r, alpha)
		dst[p+1] = rgba8Lerp(dst[p+1], g, alpha)
		dst[p+2] = rgba8Lerp(dst[p+2], b, alpha)
		dst[p+3] = rgba8Prelerp(dst[p+3], alpha, alpha)
	}
}

func rgba8Multiply(a, b uint8) uint8 {
	t := uint32(a)*uint32(b) + 128
	return uint8(((t >> 8) + t) >> 8)
}

func rgba8Lerp(p, q, a uint8) uint8 {
	var greater int32
	if p > q {
		greater = 1
	}
	t := (int32(q)-int32(p))*int32(a) + 128 - greater
	return uint8(int32(p) + (((t >> 8) + t) >> 8))
}

func rgba8Prelerp(p, q, a uint8) uint8 {
	return p + q - rgba8Multiply(p, a)
}

// BlendHlineRGBA blends a solid color with uniform coverage into tightly packed RGBA pixels.
// dst must contain 4-byte pixels in R,G,B,A byte order.
func BlendHlineRGBA(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	currentImplementation().blendHlineRGBA(dst, r, g, b, a, cover, count, premulSrc)
}

func blendHlineRGBAGeneric(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if premulSrc {
		sr, sg, sb, sa := r, g, b, a
		if cover != 255 {
			sr = rgba8Multiply(sr, cover)
			sg = rgba8Multiply(sg, cover)
			sb = rgba8Multiply(sb, cover)
			sa = rgba8Multiply(sa, cover)
		}
		if sa == 0 && sr == 0 && sg == 0 && sb == 0 {
			return
		}
		for i := range count {
			p := i * 4
			dst[p+0] = rgba8Prelerp(dst[p+0], sr, sa)
			dst[p+1] = rgba8Prelerp(dst[p+1], sg, sa)
			dst[p+2] = rgba8Prelerp(dst[p+2], sb, sa)
			dst[p+3] = rgba8Prelerp(dst[p+3], sa, sa)
		}
		return
	}
	// Non-premul: precompute effective alpha once.
	alpha := rgba8Multiply(a, cover)
	if alpha == 0 {
		return
	}
	for i := range count {
		p := i * 4
		dst[p+0] = rgba8Lerp(dst[p+0], r, alpha)
		dst[p+1] = rgba8Lerp(dst[p+1], g, alpha)
		dst[p+2] = rgba8Lerp(dst[p+2], b, alpha)
		dst[p+3] = rgba8Prelerp(dst[p+3], alpha, alpha)
	}
}

// BlendColorHspanRGBA blends per-pixel RGBA colors with per-pixel coverage into dst.
// srcColors must be a flat RGBA byte slice (4 bytes per pixel, same count as dst pixels).
// covers is per-pixel coverage (nil → uniform cover=255, alpha = src_a directly).
func BlendColorHspanRGBA(dst, srcColors, covers []byte, count int, premulSrc bool) {
	if count <= 0 || len(dst) < 4 || len(srcColors) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	currentImplementation().blendColorHspanRGBA(dst, srcColors, covers, count, premulSrc)
}

// PremultiplyRGBA converts count tightly-packed RGBA pixels in buf from
// straight-alpha to premultiplied-alpha in-place. buf must be R,G,B,A order.
func PremultiplyRGBA(buf []byte, count int) {
	if count <= 0 || len(buf) < 4 {
		return
	}
	maxCount := len(buf) / 4
	if count > maxCount {
		count = maxCount
	}
	currentImplementation().premultiplyRGBA(buf, count)
}

// DemultiplyRGBA converts count tightly-packed RGBA pixels in buf from
// premultiplied-alpha to straight-alpha in-place. buf must be R,G,B,A order.
func DemultiplyRGBA(buf []byte, count int) {
	if count <= 0 || len(buf) < 4 {
		return
	}
	maxCount := len(buf) / 4
	if count > maxCount {
		count = maxCount
	}
	currentImplementation().demultiplyRGBA(buf, count)
}

func premultiplyRGBAGeneric(buf []byte, count int) {
	for i := range count {
		p := i * 4
		a := buf[p+3]
		if a == 255 {
			continue
		}
		if a == 0 {
			buf[p+0] = 0
			buf[p+1] = 0
			buf[p+2] = 0
			continue
		}
		buf[p+0] = rgba8Multiply(buf[p+0], a)
		buf[p+1] = rgba8Multiply(buf[p+1], a)
		buf[p+2] = rgba8Multiply(buf[p+2], a)
	}
}

func demultiplyRGBAGeneric(buf []byte, count int) {
	for i := range count {
		p := i * 4
		a := buf[p+3]
		if a == 0 {
			buf[p+0] = 0
			buf[p+1] = 0
			buf[p+2] = 0
			continue
		}
		if a == 255 {
			continue
		}
		buf[p+0] = demul8u(buf[p+0], a)
		buf[p+1] = demul8u(buf[p+1], a)
		buf[p+2] = demul8u(buf[p+2], a)
	}
}

// demul8u reverses premultiplication: returns round(x * 255 / a), clamped to 255.
func demul8u(x, a uint8) uint8 {
	if x >= a {
		return 255
	}
	return uint8((uint32(x)*255 + uint32(a)/2) / uint32(a))
}

// ─────────────────────────────────────────────────────────────────────────────
// Composite Porter-Duff blend modes
// ─────────────────────────────────────────────────────────────────────────────
//
// All composite functions operate on premultiplied-alpha dst (Dca, Da in [0,255]).
// Source is straight-alpha (r, g, b, a), with per-pixel coverage in covers
// (nil → full coverage = 255). The effective premultiplied source is:
//
//	Sa  = mul(a, cover)
//	Sca = {mul(r,Sa), mul(g,Sa), mul(b,Sa)}
//
// Integer arithmetic uses the AGG rounding formula: mul(x,y) = (x*y+128+((x*y+128)>>8))>>8
// which matches rgba8Multiply defined above.

// CompSrcOverHspanRGBA blends straight-alpha source over premultiplied dst
// using Porter-Duff SrcOver. dst must be packed RGBA in R,G,B,A byte order.
// covers provides per-pixel coverage (nil → uniform full coverage).
func CompSrcOverHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	currentImplementation().compSrcOverHspanRGBA(dst, covers, r, g, b, a, count)
}

// compSrcOverHspanRGBAGeneric is the scalar baseline for SrcOver on premultiplied dst.
func compSrcOverHspanRGBAGeneric(dst, covers []byte, r, g, b, a uint8, count int) {
	if a == 0 && covers == nil {
		return
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		if sa == 0 {
			continue
		}
		scar := rgba8Multiply(r, sa)
		scag := rgba8Multiply(g, sa)
		scab := rgba8Multiply(b, sa)
		// SrcOver: Dca' = Sca + Dca - mul(Dca, Sa) = rgba8Prelerp(Dca, Sca, Sa)
		dst[p+0] = rgba8Prelerp(dst[p+0], scar, sa)
		dst[p+1] = rgba8Prelerp(dst[p+1], scag, sa)
		dst[p+2] = rgba8Prelerp(dst[p+2], scab, sa)
		dst[p+3] = rgba8Prelerp(dst[p+3], sa, sa)
	}
}

// CompDstOverHspanRGBA blends using Porter-Duff DstOver on premultiplied dst.
// Dca' = Dca + Sca*(1-Da); Da' = Da + Sa*(1-Da)
func CompDstOverHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		if sa == 0 {
			continue
		}
		da := dst[p+3]
		scar := rgba8Multiply(r, sa)
		scag := rgba8Multiply(g, sa)
		scab := rgba8Multiply(b, sa)
		// DstOver: Dca' = rgba8Prelerp(Sca, Dca, Da) = Sca + Dca - mul(Sca, Da)
		dst[p+0] = rgba8Prelerp(scar, dst[p+0], da)
		dst[p+1] = rgba8Prelerp(scag, dst[p+1], da)
		dst[p+2] = rgba8Prelerp(scab, dst[p+2], da)
		dst[p+3] = rgba8Prelerp(sa, da, da)
	}
}

// CompSrcInHspanRGBA blends using Porter-Duff SrcIn on premultiplied dst.
// Dca' = Sca*Da; Da' = Sa*Da
func CompSrcInHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		da := dst[p+3]
		scar := rgba8Multiply(rgba8Multiply(r, sa), da)
		scag := rgba8Multiply(rgba8Multiply(g, sa), da)
		scab := rgba8Multiply(rgba8Multiply(b, sa), da)
		dst[p+0] = scar
		dst[p+1] = scag
		dst[p+2] = scab
		dst[p+3] = rgba8Multiply(sa, da)
	}
}

// CompDstInHspanRGBA blends using Porter-Duff DstIn on premultiplied dst.
// Dca' = Dca*Sa; Da' = Da*Sa
func CompDstInHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		dst[p+0] = rgba8Multiply(dst[p+0], sa)
		dst[p+1] = rgba8Multiply(dst[p+1], sa)
		dst[p+2] = rgba8Multiply(dst[p+2], sa)
		dst[p+3] = rgba8Multiply(dst[p+3], sa)
	}
}

// CompSrcOutHspanRGBA blends using Porter-Duff SrcOut on premultiplied dst.
// Dca' = Sca*(1-Da); Da' = Sa*(1-Da)
func CompSrcOutHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		da := dst[p+3]
		// Sca*(1-Da) = Sca - mul(Sca, Da)
		scar := rgba8Multiply(r, sa)
		scag := rgba8Multiply(g, sa)
		scab := rgba8Multiply(b, sa)
		dst[p+0] = scar - rgba8Multiply(scar, da)
		dst[p+1] = scag - rgba8Multiply(scag, da)
		dst[p+2] = scab - rgba8Multiply(scab, da)
		dst[p+3] = sa - rgba8Multiply(sa, da)
	}
}

// CompDstOutHspanRGBA blends using Porter-Duff DstOut on premultiplied dst.
// Dca' = Dca*(1-Sa); Da' = Da*(1-Sa)
func CompDstOutHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		// Dca' = Dca - mul(Dca, Sa)
		dst[p+0] -= rgba8Multiply(dst[p+0], sa)
		dst[p+1] -= rgba8Multiply(dst[p+1], sa)
		dst[p+2] -= rgba8Multiply(dst[p+2], sa)
		dst[p+3] -= rgba8Multiply(dst[p+3], sa)
	}
}

// CompXorHspanRGBA blends using Porter-Duff Xor on premultiplied dst.
// Dca' = Sca*(1-Da) + Dca*(1-Sa); Da' = Sa + Da - 2*Sa*Da
func CompXorHspanRGBA(dst, covers []byte, r, g, b, a uint8, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	maxPixels := len(dst) / 4
	if count > maxPixels {
		count = maxPixels
	}
	for i := range count {
		p := i * 4
		var sa uint8
		if covers != nil {
			cv := covers[i]
			if cv == 0 {
				dst[p+0], dst[p+1], dst[p+2], dst[p+3] = 0, 0, 0, 0
				continue
			}
			sa = rgba8Multiply(a, cv)
		} else {
			sa = a
		}
		da := dst[p+3]
		scar := rgba8Multiply(r, sa)
		scag := rgba8Multiply(g, sa)
		scab := rgba8Multiply(b, sa)
		// Sca*(1-Da)
		scar2 := scar - rgba8Multiply(scar, da)
		scag2 := scag - rgba8Multiply(scag, da)
		scab2 := scab - rgba8Multiply(scab, da)
		// Dca*(1-Sa)
		dst[p+0] = scar2 + (dst[p+0] - rgba8Multiply(dst[p+0], sa))
		dst[p+1] = scag2 + (dst[p+1] - rgba8Multiply(dst[p+1], sa))
		dst[p+2] = scab2 + (dst[p+2] - rgba8Multiply(dst[p+2], sa))
		// Da' = Sa + Da - 2*mul(Sa,Da): use uint16 to avoid wrap
		sada := min(uint16(sa)+uint16(da)-2*uint16(rgba8Multiply(sa, da)), 255)
		dst[p+3] = uint8(sada)
	}
}

// CompClearHspanRGBA zeroes pixels in the given span.
func CompClearHspanRGBA(dst []byte, count int) {
	if count <= 0 || len(dst) < 4 {
		return
	}
	total := count * 4
	if total > len(dst) {
		total = len(dst) &^ 3
	}
	for i := range dst[:total] {
		dst[i] = 0
	}
}

func blendColorHspanRGBAGeneric(dst, srcColors, covers []byte, count int, premulSrc bool) {
	for i := range count {
		sp := i * 4
		sr, sg, sb, sa := srcColors[sp], srcColors[sp+1], srcColors[sp+2], srcColors[sp+3]
		if sa == 0 {
			continue
		}
		var cv uint8
		if covers != nil {
			cv = covers[i]
		} else {
			cv = 255
		}
		if cv == 0 {
			continue
		}
		dp := i * 4
		if premulSrc {
			rr := rgba8Multiply(sr, cv)
			gg := rgba8Multiply(sg, cv)
			bb := rgba8Multiply(sb, cv)
			aa := rgba8Multiply(sa, cv)
			if aa == 0 && rr == 0 && gg == 0 && bb == 0 {
				continue
			}
			dst[dp+0] = rgba8Prelerp(dst[dp+0], rr, aa)
			dst[dp+1] = rgba8Prelerp(dst[dp+1], gg, aa)
			dst[dp+2] = rgba8Prelerp(dst[dp+2], bb, aa)
			dst[dp+3] = rgba8Prelerp(dst[dp+3], aa, aa)
		} else {
			alpha := rgba8Multiply(sa, cv)
			if alpha == 0 {
				continue
			}
			dst[dp+0] = rgba8Lerp(dst[dp+0], sr, alpha)
			dst[dp+1] = rgba8Lerp(dst[dp+1], sg, alpha)
			dst[dp+2] = rgba8Lerp(dst[dp+2], sb, alpha)
			dst[dp+3] = rgba8Prelerp(dst[dp+3], alpha, alpha)
		}
	}
}

// blendSolidHspanRGBAWithRunFill is a hybrid blend strategy: it detects
// runs of full-coverage (255) pixels and fills them with the provided SIMD
// fill function, falling back to the generic scalar blend for partial
// coverage. This is used by SSE2 and NEON paths where a full vectorized
// per-pixel blend is not available.
func blendSolidHspanRGBAWithRunFill(
	dst []byte,
	covers []byte,
	r, g, b, a uint8,
	premulSrc bool,
	fill func(dst []byte, r, g, b, a uint8, count int),
) {
	if len(covers) == 0 {
		return
	}

	if a != 255 {
		blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
		return
	}

	for i := 0; i < len(covers); {
		if covers[i] == 255 {
			start := i
			for i < len(covers) && covers[i] == 255 {
				i++
			}
			fill(dst[start*4:], r, g, b, a, i-start)
			continue
		}

		start := i
		for i < len(covers) && covers[i] != 255 {
			i++
		}
		blendSolidHspanRGBAGeneric(dst[start*4:i*4], covers[start:i], r, g, b, a, premulSrc)
	}
}
