// Package rasterizer converts vector paths into scanline-ready coverage data.
//
// In AGG terms, this package sits between path construction/conversion and the
// scanline renderers. It is responsible for:
//
// - clipping line segments against the active clip box
// - converting world/path coordinates into the internal subpixel domain
// - accumulating edge coverage into cell arrays
// - sweeping those cells back out as anti-aliased scanlines
//
// The core entry points mirror the original AGG rasterizer families:
//
// - RasterizerScanlineAA for standard anti-aliased polygon rasterization
// - RasterizerScanlineAANoGamma for the same sweep logic without a gamma table
// - RasterizerCompoundAA for styled/layered compound rasterization
// - RasterizerOutline and RasterizerOutlineAA for stroke/outline-oriented paths
//
// Fidelity notes:
//
//   - coverage uses AGG's subpixel/cell accumulation model rather than a simpler
//     edge-sampling fallback
//   - fill-rule handling follows AGG's non-zero and even-odd semantics
//   - clipping and coordinate conversion are separated into policy objects, just
//     like AGG's ras_conv_* and rasterizer_sl_clip helpers
//
// Most users interact with this package indirectly through internal/agg2d,
// internal/scanline, and internal/renderer. It remains documented because it is
// a key maintenance boundary for rendering correctness.
package rasterizer
