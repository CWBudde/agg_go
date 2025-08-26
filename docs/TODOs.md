# Code TODOs

This checklist tracks remaining TODO items and missing features in the AGG Go port. Regenerate TODO comments with:

`rg -n "TODO|FIXME|XXX|HACK" --glob "**/*.go" -S --sort path`

## Core Library TODOs

- [ ] **Agg2D API Enhancements**
  - [ ] `internal/agg2d/blend_modes.go:102,116,145`: Implement proper blend mode support for pixel formats and renderers
  - [ ] `internal/agg2d/image.go:179,196`: Implement path-based image transformations
  - [ ] `internal/agg2d/rendering.go:349,453`: Add gamma correction and master alpha support to rasterizer
  - [ ] `internal/agg2d/utilities.go:98,122`: Implement clip box clearing and rectangular region clearing with color

- [ ] **Font System**
  - [ ] `internal/font/freetype2/engine.go:83`: Support custom memory management (optional enhancement)
  - [ ] `internal/font/freetype2/types.go:170,171`: Add conv_curve wrapper for int16/int32 paths (optional enhancement)
  - [ ] `internal/fonts/embedded_fonts.go`: Complete implementation of 6 missing embedded font datasets (MCS7x12_mono_high/low, MCS11_prop variants, MCS12_prop, MCS13_prop)

- [ ] **Rendering Pipeline**
  - [ ] `gradients.go:51,177`: Complete gradient rendering implementation
  - [ ] `images.go:91,264,324`: Finish image rendering and pattern support
  - [ ] `internal/vcgen/stroke.go:90`: Verify stroke generation aligns with original C++ implementation

## Examples TODOs

- [ ] **Interactive Controls**
  - [ ] `examples/core/intermediate/rasterizers2/main.go:687,778`: Implement full control rendering interface and timing support

## Development Notes

- Most core rendering functionality is implemented; remaining TODOs are primarily enhancements and edge cases
- Font system is functional but could benefit from curve wrapper optimizations
- Blend modes and advanced image operations are the largest remaining gaps
