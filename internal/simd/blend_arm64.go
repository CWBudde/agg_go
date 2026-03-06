//go:build arm64 && !purego

package simd

//go:noescape
func fillRGBANEONAsm(dst []byte, pixel uint32, count int)

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
	pixel := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
	fillRGBANEONAsm(dst, pixel, count)
}

func blendSolidHspanRGBANEON(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	blendSolidHspanRGBAWithRunFill(dst, covers, r, g, b, a, premulSrc, fillRGBANEON)
}
