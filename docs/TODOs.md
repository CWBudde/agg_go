# Code TODOs

This checklist tracks remaining TODO items and missing features in the AGG Go port. Regenerate TODO comments with:

`rg -n "TODO|FIXME|XXX|HACK" --glob "**/*.go" -S --sort path`

## Core Library TODOs

- [x] **Agg2D API Enhancements**

  - [x] `internal/agg2d/image.go:179,196`: Implement path-based image transformations

- [ ] **Font System**

  - [ ] `internal/font/freetype2/engine.go:83`: Support custom memory management (optional enhancement)
  - [x] `internal/font/freetype2/types.go:170,171`: Add conv_curve wrapper for int16/int32 paths (optional enhancement)
  - [x] `internal/fonts/embedded_fonts.go`: Complete implementation of remaining embedded font datasets:
    - [x] MCS11_prop - implemented with actual font data
    - [x] MCS11_prop_condensed - implemented with actual font data
    - [x] MCS12_prop - implemented with actual font data
    - [x] MCS13_prop - implemented with actual font data

- [x] **Rendering Pipeline**
  - [x] `gradients.go:51,177`: Complete gradient rendering implementation
  - [x] `images.go:91,264,324`: Finish image rendering and pattern support
  - [x] `internal/vcgen/stroke.go:90`: Verify stroke generation aligns with original C++ implementation

## Examples TODOs

- [x] **Interactive Controls**
  - [x] `examples/core/intermediate/rasterizers2/main.go:687,778`: Implement full control rendering interface and timing support

## Development Notes

- Most core rendering functionality is implemented; remaining TODOs are primarily enhancements and edge cases
- Font system is functional but could benefit from curve wrapper optimizations
- Blend modes and advanced image operations are the largest remaining gaps
