// Package pixfmt implements AGG pixel formats and pixel-format adaptors.
//
// Pixel formats are the concrete write/read backend beneath RendererBase. They
// define:
//
// - byte layout and channel order (RGBA, BGRA, RGB, Gray, packed RGB, etc.)
// - whether source data is straight alpha, premultiplied alpha, or composite-op aware
// - how primitive, line, bar, and span writes are applied to the buffer
//
// The package follows AGG's original separation of concerns:
//
//   - blenders define per-pixel math and channel ordering
//   - pixel formats expose span-oriented drawing operations over a rendering buffer
//   - adaptors extend an existing pixel format with masking, transposition, or
//     alternate blending semantics
//
// Supported format families in this package include:
//
// - RGBA/RGB 8-bit formats
// - RGBA/RGB 16-bit and float-backed variants
// - Gray 8/16/32-bit formats
// - packed 16-bit RGB formats such as 555/565
// - composite and alpha-mask adaptors
// - transposed pixel-format views
//
// In the full pipeline, renderer packages call into pixfmt operations after
// rasterizer and scanline packages have already resolved coverage.
package pixfmt
