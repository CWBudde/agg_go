// Package transform implements AGG's geometric transformation layer.
//
// These types convert coordinates between spaces before rasterization or image
// sampling. The package covers both matrix-based transforms and AGG's
// path-/viewport-based mappers.
//
// The main families mirror the original C++ `trans_*` classes:
//
// - affine transforms for translation, scale, rotation, skew, and composition
// - perspective and bilinear quadrilateral mappings
// - viewport transforms for world/device mapping with aspect-ratio policies
// - path-based transforms (`trans_single_path`, `trans_double_path`)
// - effect transforms such as `trans_warp_magnifier`
//
// Upstream references include agg_trans_affine.h, agg_trans_perspective.h,
// agg_trans_bilinear.h, agg_trans_viewport.h, agg_trans_single_path.h,
// agg_trans_double_path.h, and agg_trans_warp_magnifier.h.
package transform
