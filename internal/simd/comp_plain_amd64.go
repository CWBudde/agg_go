//go:build amd64 && !purego

package simd

//go:noescape
func compSrcOverPlainStraightHspanRGBAAVX2Asm(dst []byte, src, is1 *[4]float64, count int)

//go:noescape
func compSrcOverPlainStraightHspanRGBASSE2Asm(dst []byte, src, is1 *[4]float64, count int)

// CompSrcOverPlainStraightHspanRGBA composites a solid premultiplied source over
// count STRAIGHT-alpha RGBA/BGRA destination pixels using Porter-Duff SrcOver,
// writing straight-alpha results. It is the SIMD tier for the straight composite
// blender's uniform-coverage SrcOver span and is bit-for-bit identical to the
// scalar premultiply -> op -> demultiply bridge.
//
// Two tiers, both float64 and byte-identical to the scalar path and to each
// other: AVX2 (4 channels in one 256-bit register) when available, otherwise the
// SSE2 fallback (each pixel split across two 128-bit registers) for pre-AVX2
// amd64. SSE2 is architecturally guaranteed on amd64, so the only non-SIMD
// outcome here is ForceGeneric (test hook).
//
// src holds the premultiplied source {Sca,Sca,Sca,Sa} in DST byte order (lane k
// maps to dst byte k); is1 holds {1-Sa, ...}. Alpha must be at byte 3 (RGBA or
// BGRA) — the caller is responsible for that gate and for building src in byte
// order. Returns true when a SIMD kernel handled the span; returns false when no
// SIMD tier is usable (forced generic) so the caller falls back to the scalar
// bridge — never silently skipping work.
func CompSrcOverPlainStraightHspanRGBA(dst []byte, src, is1 *[4]float64, count int) bool {
	f := DetectFeatures()
	if f.ForceGeneric {
		return false
	}
	if !f.HasAVX2 && !f.HasSSE2 {
		return false
	}
	if count <= 0 {
		return true
	}
	if maxPixels := len(dst) / 4; count > maxPixels {
		count = maxPixels
	}
	if count <= 0 {
		return true
	}
	if f.HasAVX2 {
		compSrcOverPlainStraightHspanRGBAAVX2Asm(dst, src, is1, count)
	} else {
		compSrcOverPlainStraightHspanRGBASSE2Asm(dst, src, is1, count)
	}
	return true
}
