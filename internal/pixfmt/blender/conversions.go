package blender

import "agg_go/internal/color"

// ConvertRGBAToRGB converts an RGBA8 color to RGB8 by dropping the alpha channel.
func ConvertRGBAToRGB[S color.Space](rgba color.RGBA8[S]) color.RGB8[S] {
	return color.RGB8[S]{R: rgba.R, G: rgba.G, B: rgba.B}
}

// ConvertRGBToRGBA converts an RGB8 color to RGBA8 by setting alpha to 255 (opaque).
func ConvertRGBToRGBA[S color.Space](rgb color.RGB8[S]) color.RGBA8[S] {
	return color.RGBA8[S]{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}
