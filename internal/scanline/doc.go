// Package scanline implements AGG's scanline containers, scanline storage, and
// scanline-combination helpers.
//
// Rasterizers accumulate coverage into scanline objects one row at a time.
// Renderers then iterate the produced spans and forward them to pixfmt methods.
//
// The package keeps AGG's main scanline families:
//
// - unpacked anti-aliased scanlines (`scanline_u8`, `scanline32_u8`)
// - packed anti-aliased scanlines (`scanline_p8`, `scanline32_p8`)
// - binary scanlines without coverage (`scanline_bin`, `scanline32_bin`)
// - storage/adaptor types for recording and replaying scanlines
// - boolean algebra helpers for combining scanline streams
//
// Upstream references include agg_scanline_u.h, agg_scanline_p.h,
// agg_scanline_bin.h, agg_scanline_storage_aa.h, agg_scanline_storage_bin.h,
// and agg_scanline_boolean_algebra.h.
package scanline
