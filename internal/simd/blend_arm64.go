//go:build arm64 && !purego

package simd

func selectImplementationArch(features Features) implementation {
	if features.ForceGeneric {
		return genericImplementation()
	}
	if features.HasNEON {
		return implementation{
			name:                "neon",
			fillRGBA:            fillRGBANEON,
			blendSolidHspanRGBA: blendSolidHspanRGBANEON,
		}
	}
	return genericImplementation()
}

func fillRGBANEON(dst []byte, r, g, b, a uint8, count int) {
	fillRGBAGeneric(dst, r, g, b, a, count)
}

func blendSolidHspanRGBANEON(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
}
