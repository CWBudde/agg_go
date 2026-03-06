//go:build amd64 && !purego

package simd

//go:noescape
func fillRGBAAVX2Asm(dst []byte, pixel uint32, count int)

//go:noescape
func fillRGBASSE2Asm(dst []byte, pixel uint32, count int)

//go:noescape
func blendSolidHspanRGBASSE41Asm(dst []byte, covers []byte, pixelOpaque uint32, srcA uint8, count int)

//go:noescape
func blendSolidHspanRGBAAVX2Asm(dst []byte, covers []byte, pixelOpaque uint32, srcA uint8, count int)

func selectImplementationArch(features Features) implementation {
	if features.ForceGeneric {
		return genericImplementation()
	}
	if features.HasAVX2 {
		return implementation{
			name:                "avx2",
			fillRGBA:            fillRGBAAVX2,
			blendSolidHspanRGBA: blendSolidHspanRGBAAVX2,
		}
	}
	if features.HasSSE2 && features.HasSSSE3 && features.HasSSE41 {
		return implementation{
			name:                "sse41",
			fillRGBA:            fillRGBASSE2,
			blendSolidHspanRGBA: blendSolidHspanRGBASSE41,
		}
	}
	if features.HasSSE2 {
		return implementation{
			name:                "sse2",
			fillRGBA:            fillRGBASSE2,
			blendSolidHspanRGBA: blendSolidHspanRGBASSE2,
		}
	}
	return genericImplementation()
}

func fillRGBAAVX2(dst []byte, r, g, b, a uint8, count int) {
	pixel := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
	fillRGBAAVX2Asm(dst, pixel, count)
}

func fillRGBASSE2(dst []byte, r, g, b, a uint8, count int) {
	pixel := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(a)<<24
	fillRGBASSE2Asm(dst, pixel, count)
}

func blendSolidHspanRGBAAVX2(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if premulSrc {
		blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendSolidHspanRGBAAVX2Asm(dst, covers, pixelOpaque, a, len(covers))
}

func blendSolidHspanRGBASSE41(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if premulSrc {
		blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendSolidHspanRGBASSE41Asm(dst, covers, pixelOpaque, a, len(covers))
}

func blendSolidHspanRGBASSE2(dst []byte, covers []byte, r, g, b, a uint8, premulSrc bool) {
	blendSolidHspanRGBAWithRunFill(dst, covers, r, g, b, a, premulSrc, fillRGBASSE2)
}

