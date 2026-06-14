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

// applyPremulLinearToSRGB converts a premultiplied linear RGBA buffer to straight sRGB in-place.
// Used for demos that render into a PixFmtRGBA32Pre (premultiplied) buffer.
// The canvas putImageData expects straight sRGB, so we un-premultiply before gamma encoding
// and do not re-premultiply afterwards.
func applyPremulLinearToSRGB(img *agg.Image) {
	d := img.Data
	for i := 0; i+3 < len(d); i += 4 {
		a := d[i+3]
		if a == 0 {
			d[i], d[i+1], d[i+2] = 0, 0, 0
			continue
		}
		if a == 255 {
			// Fully opaque: no un-premultiply needed, just gamma encode.
			c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{
				R: d[i], G: d[i+1], B: d[i+2], A: 255,
			})
			d[i], d[i+1], d[i+2] = c.R, c.G, c.B
			continue
		}
		// Un-premultiply: straight = premul * 255 / alpha (rounded).
		inv := uint32(255)
		r := uint8((uint32(d[i])*inv + uint32(a)/2) / uint32(a))
		g := uint8((uint32(d[i+1])*inv + uint32(a)/2) / uint32(a))
		b := uint8((uint32(d[i+2])*inv + uint32(a)/2) / uint32(a))
		c := icol.ConvertToSRGBFromLinear(icol.RGBA8[icol.Linear]{R: r, G: g, B: b, A: a})
		d[i], d[i+1], d[i+2] = c.R, c.G, c.B
	}
}
