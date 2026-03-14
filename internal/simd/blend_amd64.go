//go:build amd64 && !purego

package simd

//go:noescape
func fillRGBAAVX2Asm(dst []byte, pixel uint32, count int)

//go:noescape
func fillRGBASSE2Asm(dst []byte, pixel uint32, count int)

//go:noescape
func copyMask1U8SSE41Asm(dst, src []byte, count int)

//go:noescape
func copyMask1U8AVX2Asm(dst, src []byte, count int)

//go:noescape
func rgb24ToGrayU8SSE41Asm(dst, src []byte, blocks int)

//go:noescape
func rgb24ToGrayU8AVX2Asm(dst, src []byte, blocks int)

//go:noescape
func blendSolidHspanRGBASSE41Asm(dst, covers []byte, pixelOpaque uint32, srcA uint8, count int)

//go:noescape
func blendSolidHspanRGBAAVX2Asm(dst, covers []byte, pixelOpaque uint32, srcA uint8, count int)

//go:noescape
func blendHlineRGBASSE41Asm(dst []byte, pixelOpaque uint32, alpha uint8, count int)

//go:noescape
func blendColorHspanRGBASSE41Asm(dst, srcColors, covers []byte, count int)

//go:noescape
func premultiplyRGBASSE41Asm(buf []byte, count int)

//go:noescape
func compSrcOverHspanRGBASSE41Asm(dst []byte, sca uint32, sa uint8, count int)

func selectImplementationArch(features Features) implementation {
	if features.ForceGeneric {
		return genericImplementation()
	}
	if features.HasAVX2 {
		return implementation{
			name:                 "avx2",
			fillRGBA:             fillRGBAAVX2,
			copyMask1U8:          copyMask1U8AVX2,
			rgb24ToGrayU8:        rgb24ToGrayU8AVX2,
			blendSolidHspanRGBA:  blendSolidHspanRGBAAVX2,
			blendHlineRGBA:       blendHlineRGBAAVX2,
			blendColorHspanRGBA:  blendColorHspanRGBAAVX2,
			premultiplyRGBA:      premultiplyRGBASSE41,
			demultiplyRGBA:       demultiplyRGBAGeneric,
			compSrcOverHspanRGBA: compSrcOverHspanRGBAAVX2,
		}
	}
	if features.HasSSE2 && features.HasSSSE3 && features.HasSSE41 {
		return implementation{
			name:                 "sse41",
			fillRGBA:             fillRGBASSE2,
			copyMask1U8:          copyMask1U8SSE41,
			rgb24ToGrayU8:        rgb24ToGrayU8SSE41,
			blendSolidHspanRGBA:  blendSolidHspanRGBASSE41,
			blendHlineRGBA:       blendHlineRGBASSE41,
			blendColorHspanRGBA:  blendColorHspanRGBASSE41,
			premultiplyRGBA:      premultiplyRGBASSE41,
			demultiplyRGBA:       demultiplyRGBAGeneric,
			compSrcOverHspanRGBA: compSrcOverHspanRGBASSE41,
		}
	}
	if features.HasSSE2 {
		return implementation{
			name:                 "sse2",
			fillRGBA:             fillRGBASSE2,
			copyMask1U8:          copyMask1U8Generic,
			rgb24ToGrayU8:        rgb24ToGrayU8Generic,
			blendSolidHspanRGBA:  blendSolidHspanRGBASSE2,
			blendHlineRGBA:       blendHlineRGBASSE2,
			blendColorHspanRGBA:  blendColorHspanRGBASSE2,
			premultiplyRGBA:      premultiplyRGBAGeneric,
			demultiplyRGBA:       demultiplyRGBAGeneric,
			compSrcOverHspanRGBA: compSrcOverHspanRGBAGeneric,
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

func copyMask1U8SSE41(dst, src []byte, count int) {
	if count > 128 {
		copyMask1U8Generic(dst, src, count)
		return
	}
	copyMask1U8SSE41Asm(dst, src, count)
}

func copyMask1U8AVX2(dst, src []byte, count int) {
	if count > 128 {
		copyMask1U8Generic(dst, src, count)
		return
	}
	copyMask1U8AVX2Asm(dst, src, count)
}

func rgb24ToGrayU8SSE41(dst, src []byte, count int) {
	blocks := 0
	if len(src) >= 16 {
		blocks = (len(src)-16)/12 + 1
		maxBlocks := count / 4
		if blocks > maxBlocks {
			blocks = maxBlocks
		}
	}
	if blocks > 0 {
		rgb24ToGrayU8SSE41Asm(dst, src, blocks)
	}
	processed := blocks * 4
	if processed < count {
		rgb24ToGrayU8Generic(dst[processed:], src[processed*3:], count-processed)
	}
}

func rgb24ToGrayU8AVX2(dst, src []byte, count int) {
	blocks := 0
	if len(src) >= 28 {
		blocks = (len(src)-28)/24 + 1
		maxBlocks := count / 8
		if blocks > maxBlocks {
			blocks = maxBlocks
		}
	}
	if blocks > 0 {
		rgb24ToGrayU8AVX2Asm(dst, src, blocks)
	}
	processed := blocks * 8
	if processed < count {
		rgb24ToGrayU8SSE41(dst[processed:], src[processed*3:], count-processed)
	}
}

func blendSolidHspanRGBAAVX2(dst, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if premulSrc {
		blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendSolidHspanRGBAAVX2Asm(dst, covers, pixelOpaque, a, len(covers))
}

func blendSolidHspanRGBASSE41(dst, covers []byte, r, g, b, a uint8, premulSrc bool) {
	if premulSrc {
		blendSolidHspanRGBAGeneric(dst, covers, r, g, b, a, premulSrc)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendSolidHspanRGBASSE41Asm(dst, covers, pixelOpaque, a, len(covers))
}

func blendSolidHspanRGBASSE2(dst, covers []byte, r, g, b, a uint8, premulSrc bool) {
	blendSolidHspanRGBAWithRunFill(dst, covers, r, g, b, a, premulSrc, fillRGBASSE2)
}

func blendHlineRGBAAVX2(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if premulSrc {
		blendHlineRGBAGeneric(dst, r, g, b, a, cover, count, premulSrc)
		return
	}
	alpha := rgba8Multiply(a, cover)
	if alpha == 0 {
		return
	}
	if alpha == 255 {
		fillRGBAAVX2(dst, r, g, b, a, count)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendHlineRGBASSE41Asm(dst, pixelOpaque, alpha, count)
}

func blendHlineRGBASSE41(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if premulSrc {
		blendHlineRGBAGeneric(dst, r, g, b, a, cover, count, premulSrc)
		return
	}
	alpha := rgba8Multiply(a, cover)
	if alpha == 0 {
		return
	}
	if alpha == 255 {
		fillRGBASSE2(dst, r, g, b, a, count)
		return
	}
	pixelOpaque := uint32(r) | uint32(g)<<8 | uint32(b)<<16 | uint32(0xFF)<<24
	blendHlineRGBASSE41Asm(dst, pixelOpaque, alpha, count)
}

func blendHlineRGBASSE2(dst []byte, r, g, b, a, cover uint8, count int, premulSrc bool) {
	if !premulSrc {
		alpha := rgba8Multiply(a, cover)
		if alpha == 255 {
			fillRGBASSE2(dst, r, g, b, a, count)
			return
		}
	}
	blendHlineRGBAGeneric(dst, r, g, b, a, cover, count, premulSrc)
}

// blendColorHspanRGBAAVX2 and blendColorHspanRGBASSE41 both use the SSE4.1 asm
// (PSHUFB / PMOVZXBW / PMULLW path). AVX2 has no separate BlendColorHspan kernel
// because the bottleneck is scalar alpha computation, not the arithmetic bandwidth.
func blendColorHspanRGBAAVX2(dst, srcColors, covers []byte, count int, premulSrc bool) {
	if premulSrc || covers == nil {
		blendColorHspanRGBAGeneric(dst, srcColors, covers, count, premulSrc)
		return
	}
	blendColorHspanRGBASSE41Asm(dst, srcColors, covers, count)
}

func blendColorHspanRGBASSE41(dst, srcColors, covers []byte, count int, premulSrc bool) {
	if premulSrc || covers == nil {
		blendColorHspanRGBAGeneric(dst, srcColors, covers, count, premulSrc)
		return
	}
	blendColorHspanRGBASSE41Asm(dst, srcColors, covers, count)
}

func blendColorHspanRGBASSE2(dst, srcColors, covers []byte, count int, premulSrc bool) {
	blendColorHspanRGBAGeneric(dst, srcColors, covers, count, premulSrc)
}

func premultiplyRGBASSE41(buf []byte, count int) {
	premultiplyRGBASSE41Asm(buf, count)
}

// compSrcOverHspanRGBASSE41 blends a solid straight-alpha src over premultiplied dst
// using Porter-Duff SrcOver. The SSE4.1 path handles the uniform-coverage case
// (covers == nil); variable coverage falls back to the generic scalar loop.
func compSrcOverHspanRGBASSE41(dst, covers []byte, r, g, b, a uint8, count int) {
	if a == 0 && covers == nil {
		return
	}
	if covers != nil {
		compSrcOverHspanRGBAGeneric(dst, covers, r, g, b, a, count)
		return
	}
	sca := uint32(rgba8Multiply(r, a)) |
		uint32(rgba8Multiply(g, a))<<8 |
		uint32(rgba8Multiply(b, a))<<16 |
		uint32(a)<<24
	compSrcOverHspanRGBASSE41Asm(dst, sca, a, count)
}

// compSrcOverHspanRGBAAVX2 delegates to the SSE4.1 kernel (the bottleneck is
// memory bandwidth, not arithmetic throughput).
func compSrcOverHspanRGBAAVX2(dst, covers []byte, r, g, b, a uint8, count int) {
	compSrcOverHspanRGBASSE41(dst, covers, r, g, b, a, count)
}
