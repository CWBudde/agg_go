package color

import (
	"agg_go/internal/basics"
)

// Luminance calculates luminance from RGBA using ITU-R BT.709
func LuminanceFromRGBA(c RGBA) float64 {
	return 0.299*c.R + 0.587*c.G + 0.114*c.B
}

// LuminanceFromRGBA8 calculates luminance from 8-bit RGBA (optimized)
func LuminanceFromRGBA8[CS ColorSpace](c RGBA8[CS]) basics.Int8u {
	// Using integer arithmetic: 0.299*77 + 0.587*150 + 0.114*29 ≈ 256
	return basics.Int8u((uint32(c.R)*77 + uint32(c.G)*150 + uint32(c.B)*29) >> 8)
}
