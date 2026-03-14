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
			name:                 "neon",
			fillRGBA:             fillRGBANEON,
			copyMask1U8:          copyMask1U8Generic,
			rgb24ToGrayU8:        rgb24ToGrayU8Generic,
			blendSolidHspanRGBA:  blendSolidHspanRGBANEON,
			blendHlineRGBA:       blendHlineRGBANEON,
			blendColorHspanRGBA:  blendColorHspanRGBANEON,
			premultiplyRGBA:      premultiplyRGBAGeneric,
			demultiplyRGBA:       demultiplyRGBAGeneric,
			compSrcOverHspanRGBA: compSrcOverHspanRGBAGeneric,
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

func blendHlineRGBANEON(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if !premulSrc {
		alpha := rgba8Multiply(a, cover)
		if alpha == 255 {
			fillRGBANEON(dst, r, g, b, a, count)
			return
		}
	}
	blendHlineRGBAGeneric(dst, r, g, b, a, cover, count, premulSrc)
}

func blendColorHspanRGBANEON(dst, srcColors, covers []byte, count int, premulSrc bool) {
	blendColorHspanRGBAGeneric(dst, srcColors, covers, count, premulSrc)
}
