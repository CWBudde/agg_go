package simd

func copyMask1U8Generic(dst, src []byte, count int) {
	copy(dst[:count], src[:count])
}

func rgb24ToGrayU8Generic(dst, src []byte, count int) {
	for i := 0; i < count; i++ {
		base := i * 3
		dst[i] = uint8((int(src[base])*77 + int(src[base+1])*150 + int(src[base+2])*29) >> 8)
	}
}
