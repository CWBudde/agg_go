// Package color implements AGG's color model layer.
//
// It provides the concrete color value types used throughout the rendering
// pipeline, along with colorspace conversion, luminance/gamma helpers, and
// cover-aware blending utilities.
//
// The package keeps the same broad split as C++ AGG:
//
// - floating-point base colors (`RGBA`)
// - fixed-width RGB/RGBA/Gray families (`*8`, `*16`, `*32`)
// - colorspace markers (`Linear`, `SRGB`)
// - helper math for premultiply/demultiply, interpolation, and blending
// - row/pixel format conversion helpers in `internal/color/conv`
//
// Upstream references include agg_color_rgba.h, agg_color_gray.h,
// agg_gamma_lut.h, agg_color_conv.h, agg_color_conv_rgb8.h, and
// agg_color_conv_rgb16.h.
package color
