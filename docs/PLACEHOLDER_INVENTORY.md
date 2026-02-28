# Placeholder Inventory (Rendering-Critical Packages)

Scope: `internal/agg2d`, `internal/rasterizer`, `internal/scanline`, `internal/renderer`, `internal/span`.

Classification policy:
- `must-fix`: currently changes rendering semantics vs AGG or risks incorrect output.
- `acceptable temporary`: not ideal, but either AGG-equivalent or not currently fidelity-breaking.
- `low-priority`: non-critical divergence; keep tracked, fix after parity-critical paths.

## Inventory

| Package | Location | Placeholder / Simplification | Class | AGG Reference | Target Phase |
|---|---|---|---|---|---|
| `internal/agg2d` | `internal/agg2d/image.go:42` | `renderImageWithPath` uses a custom pixel loop + inverse mapping, bypassing AGG span/interpolator/filter pipeline; nearest-neighbor note at `:149`. | `must-fix` | `../agg-2.6/agg-src/agg2d/agg2d.cpp:1600`, `:1718`, `:1282` | Phase `1.1` |
| `internal/agg2d` | `internal/agg2d/gradient.go:351` | `worldToScreen`/`worldToScreenPoint` helper is a no-op/1:1 placeholder, affecting radial gradient transform setup. | `must-fix` | `../agg-2.6/agg-src/agg2d/agg2d.cpp:276`, `:289`, `:540` | Phase `1.2` |
| `internal/rasterizer` | `internal/rasterizer/cells_aa_simple.go:497` | After consolidation, Y-runs are not compacted ("leave gaps for now"), risking run-index mismatch. | `must-fix` | `../agg-2.6/agg-src/include/agg_rasterizer_cells_aa.h:627` | Phase `2.1` |
| `internal/rasterizer` | `internal/rasterizer/cells_aa_styled.go:483` | Same non-compaction behavior in styled rasterizer path. | `must-fix` | `../agg-2.6/agg-src/include/agg_rasterizer_cells_aa.h:627` | Phase `2.1` |
| `internal/span` | `internal/span/span_image_filter_rgb.go:316` | Bilinear-clip partial-overlap path falls back to background instead of weighted edge sampling. | `must-fix` | `../agg-2.6/agg-src/include/agg_span_image_filter_rgb.h:170`, `:270` | Phase `1.1` |
| `internal/renderer` | `internal/renderer/scanline/helpers.go:46` | `RenderAllPaths` uses minimal placeholder-type interfaces (dynamic add-path/color-set). Functionally close, but weakly typed vs AGG template contract. | `acceptable temporary` | `../agg-2.6/agg-src/include/agg_renderer_scanline.h:454` | Phase `2.3` |
| `internal/renderer` | `internal/renderer/outline/outline_image.go:1071` | Methods marked "not implemented" (`Semidot/Pie/Line0/Line1/Line2`), but AGG reference methods are also empty stubs. | `acceptable temporary` | `../agg-2.6/agg-src/include/agg_renderer_outline_image.h:903` | Keep as-is |
| `internal/agg2d` | `internal/agg2d/adapters.go:22` | `RowPtr` adapter returns `nil` (simplified bridge). Not currently parity-critical, but blocks some row-pointer-style optimizations/compat paths. | `low-priority` | `../agg-2.6/agg-src/include/agg_rendering_buffer.h:86` | Phase `2.2` |
| `internal/span` | `internal/span/span_gradient_contour.go:123` | Contour creation rasterizes with manual Bresenham path simplification instead of AGG outline rasterization stage. | `low-priority` | `../agg-2.6/agg-src/include/agg_span_gradient_contour.h:171` | Phase `2.x` |

## Priority Summary

- `must-fix`: 5
- `acceptable temporary`: 2
- `low-priority`: 2

## Recently Resolved

- `internal/scanline/storage_aa_serialized.go`: removed placeholder embedded iterator and `any`-based cover decoding; now parses serialized AA spans per AGG layout.
- `internal/agg2d/text.go`: removed rectangle glyph fallback; gray8/mono glyphs now render using bitmap coverage data per pixel.

## Execution Order (Parity-First)

1. Phase `1.1`: image transform/render + span image filter clipping parity.
2. Phase `1.2`: gradient world/screen transform parity.
3. Phase `1.3`: text glyph scanline decoding/render parity.
4. Phase `2.1`: rasterizer cell run compaction correctness.
5. Phase `2.2`/`2.3`: adapter and renderer interface cleanup.
