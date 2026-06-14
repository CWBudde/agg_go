//go:build !amd64 || purego

package simd

// CompSrcOverPlainStraightHspanRGBA has no SIMD implementation on this
// platform/build; it always reports "not handled" so the caller uses the scalar
// bridge. See the amd64 build for the documented contract.
func CompSrcOverPlainStraightHspanRGBA(dst []byte, src, is1 *[4]float64, count int) bool {
	return false
}
