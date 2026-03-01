package pixfmt

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func detectBytesPerPixel(src interface {
	RowData(y int) []basics.Int8u
	Width() int
	Height() int
}, sampleY int,
) int {
	if pixWidthProvider, ok := src.(interface{ PixWidth() int }); ok {
		return pixWidthProvider.PixWidth()
	}
	if sampleY >= 0 && sampleY < src.Height() {
		srcRowData := src.RowData(sampleY)
		if srcRowData != nil && src.Width() > 0 {
			detectedBPP := len(srcRowData) / src.Width()
			if detectedBPP > 0 && detectedBPP <= 8 {
				return detectedBPP
			}
		}
	}
	return 4
}

func decodeRGBA8FromRowData(row []basics.Int8u, bytesPerPixel, pixelX int) (color.RGBA8[color.Linear], bool) {
	if pixelX < 0 {
		return color.RGBA8[color.Linear]{}, false
	}
	srcOffset := pixelX * bytesPerPixel
	if srcOffset < 0 || srcOffset+bytesPerPixel-1 >= len(row) {
		return color.RGBA8[color.Linear]{}, false
	}

	switch bytesPerPixel {
	case 1:
		gray := basics.Int8u(row[srcOffset])
		return color.RGBA8[color.Linear]{R: gray, G: gray, B: gray, A: 255}, true
	case 2:
		gray := basics.Int8u(row[srcOffset])
		alpha := basics.Int8u(row[srcOffset+1])
		return color.RGBA8[color.Linear]{R: gray, G: gray, B: gray, A: alpha}, true
	case 3:
		return color.RGBA8[color.Linear]{
			R: basics.Int8u(row[srcOffset]),
			G: basics.Int8u(row[srcOffset+1]),
			B: basics.Int8u(row[srcOffset+2]),
			A: 255,
		}, true
	default:
		return color.RGBA8[color.Linear]{
			R: basics.Int8u(row[srcOffset]),
			G: basics.Int8u(row[srcOffset+1]),
			B: basics.Int8u(row[srcOffset+2]),
			A: basics.Int8u(row[srcOffset+3]),
		}, true
	}
}
