//go:build arm64 && !purego

package simd

//go:noescape
func fillRGBANEONAsm(dst []byte, pixel uint32, count int)

//go:noescape
func copyMask1UNEONAsm(dst, src []byte, count int)

//go:noescape
func rgb24ToGrayU8NEONAsm(dst, src []byte, blocks int)

func selectImplementationArch(features Features) implementation {
	if features.ForceGeneric {
		return genericImplementation()
	}
	if features.HasNEON {
		return implementation{
			name:                 "neon",
			fillRGBA:             fillRGBANEON,
			copyMask1U8:          copyMask1UNEON,
			rgb24ToGrayU8:        rgb24ToGrayU8NEON,
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

func copyMask1UNEON(dst, src []byte, count int) {
	copyMask1UNEONAsm(dst, src, count)
}

func rgb24ToGrayU8NEON(dst, src []byte, count int) {
	blocks := count / 8
	if maxBlocks := len(src) / 24; blocks > maxBlocks {
		blocks = maxBlocks
	}
	if blocks > 0 {
		rgb24ToGrayU8NEONAsm(dst, src, blocks)
	}
	processed := blocks * 8
	if processed < count {
		rgb24ToGrayU8Generic(dst[processed:], src[processed*3:], count-processed)
	}
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
