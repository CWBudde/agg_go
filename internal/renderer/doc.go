// Package renderer provides the bridge from scanline/outline coverage data to
// concrete pixel-format writes.
//
// In the AGG pipeline, rasterizers decide which pixels or spans are covered;
// renderer packages decide how those covered spans are written into a pixel
// buffer. This subtree therefore contains:
//
//   - RendererBase: clipped primitive and span writes over a pixel format
//   - scanline/: AGG-style scanline renderers for AA, binary, solid, and
//     generated-span rendering
//   - outline/: anti-aliased outline renderers and pattern-based stroke helpers
//   - raster text helpers that turn glyph spans into renderer operations
//   - small helper subpackages such as primitives and markers
//
// The package structure follows original AGG responsibilities rather than a
// single monolithic renderer abstraction: rasterizer computes coverage,
// scanlines store spans, and renderer applies those spans to a pixel format.
package renderer
