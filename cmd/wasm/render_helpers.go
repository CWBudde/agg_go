package main

import (
	agg "github.com/MeKo-Christian/agg_go"
	icol "github.com/MeKo-Christian/agg_go/internal/color"
)

// applyLinearToSRGB gamma-encodes a linearly-rendered RGBA buffer to sRGB in-place.
// Call after rendering for demos whose standalone example uses EncodeLinearRGBToSRGB: true.
func applyLinearToSRGB(img *agg.Image) {
	d := img.Data
	for i := 0; i+3 < len(d); i += 4 {
		c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
			R: d[i], G: d[i+1], B: d[i+2], A: d[i+3],
		})
		d[i], d[i+1], d[i+2] = c.R, c.G, c.B
	}
}
