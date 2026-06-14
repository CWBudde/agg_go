package main

import (
	agg "github.com/cwbudde/agg_go"
	icol "github.com/cwbudde/agg_go/internal/color"
)

// applyLinearToSRGB gamma-encodes a straight linearly-rendered RGBA buffer to sRGB in-place.
// Call after rendering for demos whose standalone example uses EncodeLinearRGBToSRGB: true.
// The buffer must contain straight (non-premultiplied) alpha; use applyPremulLinearToSRGB
// when the pixfmt stores premultiplied values (e.g. PixFmtRGBA32Pre).
func applyLinearToSRGB(img *agg.Image) {
	d := img.Data
	for i := 0; i+3 < len(d); i += 4 {
		c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
			R: d[i], G: d[i+1], B: d[i+2], A: d[i+3],
		})
		d[i], d[i+1], d[i+2] = c.R, c.G, c.B
	}
}

// applyPremulLinearToSRGB converts a premultiplied linear RGBA buffer to opaque sRGB in-place.
// Used for demos that render into a PixFmtRGBA32Pre (premultiplied) buffer.
//
// C++ AGG blits premultiplied buffers to screen via platform_support which ignores alpha
// (the platform renders to BGR, not RGBA). In the web demo we must do the equivalent:
// un-premultiply the RGB channels, gamma-encode to sRGB, then force alpha=255 so the
// canvas putImageData treats every pixel as fully opaque and does not composite against
// the page background.
func applyPremulLinearToSRGB(img *agg.Image) {
	d := img.Data
	for i := 0; i+3 < len(d); i += 4 {
		a := d[i+3]
		var r, g, b uint8
		if a == 0 {
			// Fully transparent premul pixel — treat as white (the standard clear color).
			r, g, b = 255, 255, 255
		} else if a == 255 {
			c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
				R: d[i], G: d[i+1], B: d[i+2], A: 255,
			})
			r, g, b = c.R, c.G, c.B
		} else {
			// Un-premultiply: straight = premul * 255 / alpha (rounded).
			sr := uint8((uint32(d[i])*255 + uint32(a)/2) / uint32(a))
			sg := uint8((uint32(d[i+1])*255 + uint32(a)/2) / uint32(a))
			sb := uint8((uint32(d[i+2])*255 + uint32(a)/2) / uint32(a))
			c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{R: sr, G: sg, B: sb, A: a})
			r, g, b = c.R, c.G, c.B
		}
		d[i], d[i+1], d[i+2], d[i+3] = r, g, b, 255
	}
}
