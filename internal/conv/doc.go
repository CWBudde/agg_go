// Package conv implements AGG's vertex-converter layer.
//
// Converters sit between a vertex source and the rasterizer. They consume path
// commands, transform or expand them, and expose a new vertex source with the
// converted geometry.
//
// The package follows AGG's original split:
//
// - conv_adaptor_vcgen wraps stateful vertex generators such as stroke, dash,
//   contour, B-spline, curve, and marker processors
// - conv_adaptor_vpgen wraps streaming processors such as polygon/polyline
//   clipping and segmentators
// - concrete conv_* types mirror AGG's public converter families while the
//   underlying vcgen/vpgen packages hold the geometry algorithms
//
// Upstream references include agg_conv_adaptor_vcgen.h,
// agg_conv_adaptor_vpgen.h, agg_conv_stroke.h, agg_conv_dash.h,
// agg_conv_contour.h, agg_conv_bspline.h, and the clipping converter headers.
package conv
