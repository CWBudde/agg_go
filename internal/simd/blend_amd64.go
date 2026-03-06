//go:build amd64 && !purego

package simd

//go:noescape
func fillRGBAAVX2Asm(dst []byte, pixel uint32, count int)

func selectImplementationArch(features Features) implementation {
	if features.ForceGeneric {
		return genericImplementation()
	}
	if features.HasAVX2 {
		return implementation{
			name:     "avx2",
			fillRGBA: fillRGBAAVX2,
		}
	}
	if features.HasSSE2 {
		return implementation{
			name:     "sse2",
			fillRGBA: fillRGBASSE2,
		}
	}
	return genericImplementation()
}

func fillRGBAAVX2(dst []byte, r, g, b, a uint8, count int) {
	pixel := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
	fillRGBAAVX2Asm(dst, pixel, count)
}

func fillRGBASSE2(dst []byte, r, g, b, a uint8, count int) {
	fillRGBAGeneric(dst, r, g, b, a, count)
}
