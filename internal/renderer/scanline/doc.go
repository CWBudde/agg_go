// Package scanline implements AGG-style scanline renderers.
//
// These renderers consume spans produced by a rasterizer/scanline pair and
// write them through a base renderer into a pixel format. The package provides:
//
//   - anti-aliased solid-color scanline renderers
//   - anti-aliased generated-span renderers for gradients and image filters
//   - binary scanline renderers
//   - helper functions such as RenderScanlines and RenderAllPaths that match the
//     structure of the original AGG render_scanlines helpers
//
// This package is the main meeting point between rasterizer output and span or
// solid-color painting logic.
package scanline
