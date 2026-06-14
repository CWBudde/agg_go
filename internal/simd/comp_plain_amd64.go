//go:build amd64 && !purego

package simd

//go:noescape
func compSrcOverPlainStraightHspanRGBAAVX2Asm(dst []byte, src, is1 *[4]float64, count int)

// CompSrcOverPlainStraightHspanRGBA composites a solid premultiplied source over
// count STRAIGHT-alpha RGBA/BGRA destination pixels using Porter-Duff SrcOver,
// writing straight-alpha results. It is the SIMD (AVX2, float64) tier for the
// straight composite blender's uniform-coverage SrcOver span and is bit-for-bit
// identical to the scalar premultiply -> op -> demultiply bridge.
//
// src holds the premultiplied source {Sca,Sca,Sca,Sa} in DST byte order (lane k
// maps to dst byte k); is1 holds {1-Sa, ...}. Alpha must be at byte 3 (RGBA or
// BGRA) — the caller is responsible for that gate and for building src in byte
// order. Returns true when the SIMD kernel handled the span; returns false when
// AVX2 is unavailable (or forced off) so the caller falls back to the scalar
// bridge — never silently skipping work.
func CompSrcOverPlainStraightHspanRGBA(dst []byte, src, is1 *[4]float64, count int) bool {
	f := DetectFeatures()
	if f.ForceGeneric || !f.HasAVX2 {
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
	compSrcOverPlainStraightHspanRGBAAVX2Asm(dst, src, is1, count)
	return true
}
