# Code TODOs

This checklist is generated from TODO-like comments found in Go source files (`TODO`, `FIXME`, `XXX`, `HACK`). Each file groups its items with line numbers and short descriptions so you can track progress. Regenerate with:

`rg -n "TODO|FIXME|XXX|HACK" --glob "**/*.go" -S --sort path`

- [x] internal/conv/marker_test.go

  - [x] L170: Investigate multi-marker processing to match C++ behavior

- [x] internal/conv/smooth_poly1.go

  - [x] L92: Implement proper curve approximation control

- [x] internal/ctrl/render.go

  - [x] L50: Use color with appropriate renderer
  - [x] L52: Call appropriate render function per renderer type
  - [x] L74: Implement SetColor based on renderer interface
  - [x] L75: Implement scanline rendering invocation

- [ ] internal/pixfmt/blender/base_test.go

  - [ ] L269: Move this test to the parent pixfmt package

- [ ] internal/pixfmt/pixfmt_rgb_test.go

  - [ ] L584: Fix premultiplied blending mathematics

- [ ] internal/pixfmt/pixfmt_rgba64.go

  - [ ] L214: Implement ARGB order for PixFmtARGB64Linear
  - [ ] L215: Implement ABGR order for PixFmtABGR64Linear
  - [ ] L216: Implement BGRA order for PixFmtBGRA64Linear

- [ ] internal/platform/platform_support.go

  - [ ] L339: Implement image loading (BMP/PPM format)
  - [ ] L346: Implement image saving (BMP/PPM format)

- [ ] internal/platform/x11/x11_display.go

  - [ ] L309: Implement actual image saving

- [ ] internal/font/freetype2/cache_integration.go

  - [ ] L47: Convert to Fman adaptors when available (optional enhancement)
  - [ ] L58: Convert to Fman adaptors when available (optional enhancement)

- [ ] internal/font/freetype2/engine.go

  - [ ] L83: Support custom memory management if needed (optional enhancement)

- [ ] internal/font/freetype2/types.go

  - [ ] L168: Add conv_curve wrapper for int16 paths (optional enhancement)
  - [ ] L169: Add conv_curve wrapper for int32 paths (optional enhancement)

- [ ] internal/fonts/embedded_fonts.go

  - [ ] L1056: Implement GSE4x8 font data
  - [ ] L1063: Implement GSE5x9 font data
  - [ ] L1070: Implement GSE6x9 font data
  - [ ] L1077: Implement GSE6x12 font data
  - [ ] L1084: Implement GSE7x11 font data
  - [ ] L1091: Implement GSE7x11_bold font data
  - [ ] L1098: Implement GSE7x15 font data
  - [ ] L1105: Implement GSE7x15_bold font data
  - [ ] L1112: Implement GSE8x16 font data
  - [ ] L1119: Implement GSE8x16_bold font data
  - [ ] L1134: Implement MCS5x11_mono font data
  - [ ] L1141: Implement MCS6x10_mono font data
  - [ ] L1148: Implement MCS6x11_mono font data
  - [ ] L1155: Implement MCS7x12_mono_high font data
  - [ ] L1162: Implement MCS7x12_mono_low font data
  - [ ] L1169: Implement MCS11_prop font data
  - [ ] L1176: Implement MCS11_prop_condensed font data
  - [ ] L1183: Implement MCS12_prop font data
  - [ ] L1190: Implement MCS13_prop font data
  - [ ] L1205: Implement Verdana12_bold font data
  - [ ] L1212: Implement Verdana13 font data
  - [ ] L1219: Implement Verdana13_bold font data
  - [ ] L1226: Implement Verdana14 font data
  - [ ] L1233: Implement Verdana14_bold font data
  - [ ] L1240: Implement Verdana16 font data
  - [ ] L1247: Implement Verdana16_bold font data
  - [ ] L1254: Implement Verdana17 font data
  - [ ] L1261: Implement Verdana17_bold font data
  - [ ] L1268: Implement Verdana18 font data
  - [ ] L1275: Implement Verdana18_bold font data

- [ ] internal/gpc/gpc.go

  - [ ] L352: Implement proper intersection detection and complex scanline algorithm
  - [ ] L391: Fix the complete scanline algorithm for intersection, difference, XOR
  - [ ] L396: Complete GPC algorithm implementation

- [ ] internal/rasterizer/clip.go

  - [ ] L521: Fix clipping for boundary-crossing lines
  - [ ] L553: Fix rasterizer clipping boundary detection logic

- [ ] internal/vcgen/bspline.go

  - [ ] L50: Fix edge cases with very small interpolation steps
  - [ ] L98: Fix multiple rewinds state management
  - [ ] L178: Fix B-spline generator state management
  - [ ] L217: Handle not-ready cases (insufficient points, RemoveAll, etc)

- [ ] internal/vcgen/stroke.go

  - [ ] L90: Implement proper path shortening when agg_shorten_path is ported (uses array.ShortenPath) -> Done, but confirm that this aligns with the original C++ implementation

- [ ] examples/core/basic/basic_demo/main.go

  - [ ] L292: FIX stats in demo ("TODO FIX stats!!!")

- [ ] internal/agg2d/image.go

  - [ ] L33: Complete integration with rendering pipeline
  - [ ] L161: Implement actual blending
  - [ ] L186: Implement actual copying
  - [ ] L208: Implement premultiplication
  - [ ] L220: Implement demultiplication

- [ ] internal/agg2d/rendering.go

  - [ ] L14: Implement full fill rendering with gradients and patterns
  - [ ] L33: Implement full stroke rendering with dashes and line styles
  - [ ] L73: Implement proper curve approximation scale setting

- [ ] internal/agg2d/text.go

  - [ ] L191: Apply rotation transformation
  - [ ] L225: Add glyph outline path to current path
  - [ ] L232: Render the glyph using the scanline renderer
  - [ ] L240: Render the monochrome glyph
  - [ ] L254: Implement scanline rendering for glyphs
  - [ ] L268: Use the AGG renderer to draw the scanline data
  - [ ] L285: Implement binary (1-bit) rendering path similar to AA

- [ ] internal/agg2d/text_test.go

  - [ ] L97: Add tests with actual font loading when FreeType is available

- [ ] internal/agg2d/paths.go

  - [ ] L93: Implement proper smooth curve calculations by tracking previous control points

- [ ] internal/effects/slight_blur.go

  - [ ] L83: Implement full pixel format integration for slight_blur

- [ ] internal/gpc/gpc.go

  - [ ] L928: Implement triangulation for polygons with holes in PolygonToTristrip

- [ ] internal/pixfmt/pixfmt_transposer.go

  - [ ] L177: Implement proper pixel format handling in CopyFrom

- [ ] internal/platform/x11/x11_display.go

  - [ ] L299: Implement delay using system calls

- [x] internal/renderer/scanline/helpers.go

  - [x] L245: Implement proper color blending in renderCompoundSolidStyle
  - [x] L270: Implement proper color blending in renderCompoundGeneratedStyle
