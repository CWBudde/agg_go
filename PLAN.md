# AGG Go Port - Fidelity-First Plan

## Objective

Port AGG 2.6 to Go so that:

1. Rendering behavior stays as close as possible to original AGG (`../agg-2.6/agg-src`).
2. Go code remains idiomatic, maintainable, and testable.
3. Deviations from AGG are explicit, justified, and tested.

This is the single authoritative project plan. SIMD optimization work is tracked here as later phases; there is no separate `docs/PLAN.md`.

## Non-Negotiables

- [ ] Every major behavior maps to a C++ source reference (file + method).
- [ ] No placeholder rendering paths in production-critical pipeline stages.
- [ ] Public API remains stable and idiomatic; internal architecture may change.
- [ ] Completed work is reflected in `docs/TASKS.md` with `[x]`.

## Porting Rules

1. Fidelity first for algorithms and numeric behavior.
2. Idiomatic Go for ownership, naming, package boundaries, and tests.
3. No silent fallbacks that change rendering semantics.
4. If behavior differs from AGG, document as an intentional delta.

---

## Phase 0 - Baseline and Traceability ✅

Project-wide tracking and auditability are in place:

- **Parity ledger**: One row per `Agg2D` method with C++ source reference, Go method mapping, status (`exact`/`close`/`placeholder`/`missing`), test reference, and notes. Key anchors are in `../agg-2.6/agg-src/agg2d/agg2d.h` and `../agg-2.6/agg-src/agg2d/agg2d.cpp`. Remaining open parity rows are tracked in Phase 4.
- **Placeholder inventory**: All simplified/placeholder paths in rendering-critical packages (`internal/agg2d`, `internal/rasterizer`, `internal/scanline`, `internal/renderer`, `internal/span`) are recorded and prioritized (`must-fix` / `acceptable temporary` / `low-priority`).

---

## Phase 1 - AGG2D Behavioral Parity

Primary target: `internal/agg2d/*` against `agg2d.cpp`.

Core `Agg2D` rendering behavior is aligned:

- **Image pipeline**: `renderImage*` uses the AGG-style scanline/span pipeline (interpolator sampling, filter LUT integration, resample-mode behavior, blend-color conversion) with correct transform flow and no nearest-neighbor-only fallback for transformed images.
- **Gradients**: Linear/radial matrix construction, transform/scalar conversion, and distance (`d1/d2`) handling match AGG ordering.
- **Text**: Glyph rendering runs through the real rasterizer/scanline pipeline (no rectangle fallback) and matches the vector vs raster cache contract. Kerning in `TextWidth`/`Text` uses glyph indices.
- **Clipping & state**: `ClipBox` propagation and clip-sensitive image ops match AGG semantics, covered by pixel-asserting tests. Attach-time and state-update behavior (fill rule, gamma, master alpha) is aligned.

Remaining:

- Visual tests for AGG2D demos still need to pass against reference thresholds.

---

## Phase 2 - Core Pipeline Parity ✅

Rasterizer → scanline → renderer → pixfmt behavior is aligned with AGG:

- **Rasterizer/scanline**: Fill rules, clipping edge cases, cell accumulation, sweep indexing, and duplicate-cell behavior match AGG expectations.
- **Renderer/pixfmt**: Copy/blend overlap and premultiplied vs straight-alpha behavior are aligned. The needed `copy_from` / `blend_from` helpers are ported across RGBA/RGB/Gray plus transposer/amask/composite pixfmts, with expanded Porter-Duff/composite coverage.
- **Converters**: Stroke/dash/transform ordering, line cap/join enum parity, and viewport/gradient/scalar propagation are aligned; key converter/vcgen/vpgen state machines are audited beyond just Agg2D call sites.

---

## Phase 3 - Font Subsystem Consolidation and Type Safety ✅

One coherent font stack with a tighter, type-safe surface:

- **Architecture**: A single authoritative font/cache architecture is used by Agg2D. `internal/font/freetype2` is reduced to minimal, justified wrappers (with documented Go-only deltas where applicable).
- **Type safety**: Broad runtime `interface{}` usage in text-critical paths is replaced with explicit interfaces; build-tag boundaries remain the only intentional runtime dispatch.
- **Lifecycle**: FreeType2 face/engine lifetime behavior is rechecked; multi-face close releases tracked faces correctly. The `maxFaces` cap is documented as an intentional Go-only policy delta.

---

## Phase 4 - Test Strategy for Port Fidelity

### 4.1 Contract tests (API behavior)

- [x] Expand AGG2D tests to assert outputs, not just `err == nil`, for the currently covered rendering paths.
      `internal/agg2d/rendering_test.go`, `internal/agg2d/image_test.go`, and `internal/agg2d/text_phase1_test.go` now use deterministic output assertions for solid fill, gradient fill, translated rendering output, clipped fill/stroke rendering, blend-mode compositing, transformed image placement/color coverage, and vector-text alignment/bounds.
- [x] Add deterministic checks for transform-image, clipping, blend modes, gradients, and text bounds.
- [x] Replace the remaining AGG2D smoke/integration tests in `internal/agg2d` with output or state assertions.
      `internal/agg2d/agg2d_test.go` now asserts concrete path storage results for path commands and ellipse generation, and `internal/agg2d/rendering_fixes_test.go` now verifies the rasterizer gamma mapping rather than only asserting "no panic".
- [x] Expand contract coverage for the currently identified weaker packages:
  - `internal/effects`
  - `internal/platform`
  - `internal/primitives`
  - `internal/pixfmt/blender`
  - targeted additions now cover platform backend-factory/image-format contracts plus packed-RGB and RGBA16 blender surfaces
- [x] Re-audit the remaining internal packages against the current coverage floor.
  - the next priority gaps are now `internal/pixfmt`, `internal/pixfmt/gamma`, and `internal/color`
  - support-only or demo-oriented low-coverage packages such as `internal/order`, `internal/gamma`, `internal/ctrl/text`, and `internal/demo/lion` are not Phase 4.1 blockers
- [ ] Raise the next priority coverage gaps identified by the re-audit:
  - `internal/pixfmt` (currently around 36%)
  - `internal/pixfmt/gamma` (currently around 65%)
  - `internal/color` (currently around 66%)
- [ ] Re-audit tests that primarily verify mocks or package-private state:
  - keep package-private assertions only where they are the clearest contract
  - prefer observable behavior or public API assertions where practical

### 4.2 Visual regression tests

- [ ] Generate canonical references from C++ AGG for core scenarios.
      `tests/visual/reference` already exists for the current Go-side visual suite, but the remaining parity step is to replace or supplement those images with canonical outputs generated from the original C++ AGG implementation.
- [x] Store references under `tests/visual/reference`.
- [x] Add automated diff thresholding and report generation.
      `tests/visual/framework` now supports diff-threshold pass/fail rules (`VISUAL_DIFF_TOLERANCE`, `VISUAL_MAX_DIFFERENT_PIXELS`, `VISUAL_MAX_DIFFERENT_RATIO`, `VISUAL_IGNORE_ALPHA`, `VISUAL_GENERATE_DIFFS`) and emits HTML reports with per-test diff statistics.
- [ ] Expand the C++-generated visual reference set to cover:
  - basic shapes and AA edge cases
  - gradients
  - text rendering
  - other parity-critical scenarios before less critical demos
- [ ] Expand visual coverage by category until parity-critical rendering areas are represented:
  - primitives
  - path stroke/fill variations
  - anti-aliasing quality cases
  - transformations
  - color/blend-mode cases
  - gradients and patterns
  - clipping
  - image operations
  - advanced and edge-case scenes
- [ ] Add reference-management workflow for the visual suite:
  - controlled regeneration and update flow for references
  - clear review surface for approving intended visual changes
  - keep reference categories organized under `tests/visual/reference`
- [ ] Extend visual reporting where it adds signal:
  - retain diff images and HTML reports
  - add summary metrics only if they help triage regressions without hiding pixel-level failures
- [ ] Keep visual-suite runtime practical:
  - preserve parallel execution where safe
  - separate parity-critical coverage from exhaustive scenarios if the full suite becomes too slow

### 4.3 C++ parity checks

- [ ] For each parity row marked `exact`, include at least one source-linked test case.
- [ ] For rows marked `close`, include documented rationale.

### 4.4 Test-suite cleanup and failing-test closure

- [ ] Remove or convert debug-style integration tests that only log state:
  - `tests/integration/debug_test.go`
  - `tests/integration/debug2_test.go`
  - `tests/integration/debug3_test.go`
  - `tests/integration/minimal_debug_test.go`
  - `tests/integration/alternative_debug_test.go`
- [ ] Convert any useful debug coverage into proper contract or regression tests with assertions.
- [ ] Triage and close the currently known failing or build-broken test areas:
  - `agg2d`
  - `color`
  - `conv`
  - `fonts`
  - `pixfmt`
  - `pixfmt/blender`
  - `pixfmt/gamma`
  - `platform`
- [ ] Investigate unusually slow passing test suites and reduce runtime where possible, especially in rasterizer-heavy packages.

### 4.5 Remaining AGG2D parity rows

These items were previously tracked in a standalone parity ledger and now live directly in the phased plan.

- [ ] Audit `Attach` parity for the C++ `attach(Image&)` shape:
  - confirm behavior parity for image-backed attach flows
  - document any intentional Go API delta if the method shape differs
  - add a source-linked attach/image contract test
- [ ] Finish `TextWidth` parity:
  - close remaining kerning and metrics gaps vs `agg2d.cpp`
  - add deterministic width assertions tied to source-linked behavior
- [ ] Finish `Text` parity:
  - remove any remaining simplified raster-glyph behavior
  - verify glyph placement, raster cache behavior, and text output contracts
  - promote status to `close` only after deterministic output checks exist

### 4.6 Optional property tests

- [ ] Add property-style tests for transformations where invertibility and composition laws are stable enough to assert.
- [ ] Add property-style tests for color math where round-trip and monotonicity expectations are well-defined.
- [ ] Use `testing/quick` or equivalent lightweight property tooling only where it improves confidence without making failures opaque.

### Exit criteria

- [ ] `go test ./...` passes.
- [ ] Visual regression suite passes in CI.
- [ ] No AGG2D parity row remains untriaged or placeholder-level.
- [ ] The remaining debug-only tests are either deleted or converted to assertion-based coverage.
- [ ] Visual references and approval workflow are centralized under the existing `tests/visual/` framework rather than separate ad hoc docs.

---

## Phase 5 - Remaining Port Inventory from docs/TASKS.md ✅

All port inventory items from `docs/TASKS.md` are complete, explicitly deferred as out-of-C++-AGG-2.6-scope,
or documented as intentional Go deltas. `go test ./internal/...` passes (44 packages).

**5.1 Transformations**: All 7 types at C++ parity (TransAffine, TransPerspective, TransBilinear,
TransSinglePath, TransDoublePath, TransWarpMagnifier, TransViewport). `PerspectiveIteratorX` avoids
per-pixel divide; determinant check in `Invert()` for stability; `ViewportManager` handles multi-viewport;
`ConvTransform.Transformer()` getter added, duplicate setter removed. `SpanInterpolatorPerspectiveExact`
and `SpanInterpolatorPerspectiveLerp` verified against `agg_span_interpolator_persp.h` at full parity.
Out of scope: TransPolar (example-only in C++, no `agg_trans_polar.h`), WarpMagnifier multiple zones
(not in C++ AGG 2.6).

**5.2 Converters and generators**: `ConvAdaptorVPGen` + all vpgen components verified. Stroke/contour
pipeline complete: `conv_stroke`, `vcgen_stroke` (InnerJoin, all cap/join types), `vcgen_contour`,
`conv_contour` all at C++ parity. Dash, smooth-poly, and all smaller path-utility converters complete
with tests. Rasterizer cell-run compaction regression tests added; `RenderAllPaths` typed via
`MultiPathRasterizerInterface`.

**5.3 Span / image-processing**: `GradientContour.Calculate()` formula fixed to match C++
(`buffer*(d2/256)+d1`, not linear lerp). Bilinear filter spurious premultiplied-alpha clamping removed
(not in C++ AGG); clip variant implements all three C++ boundary cases. Gouraud shading complete:
`SpanGouraudGray` and `SpanGouraudRGBA` at C++ parity. `BoundingRect`, `ConvShortenPath`,
`VCGenVertexSequence` all implemented and tested.

**5.4 Fonts and utilities**: `RowPtr` bridge resolved (direct slice for plain, on-demand cache for pre).
GSV embedded font complete (`GSVText`, `GSVTextOutline`, `font_data.go`). FreeType custom memory hook
unsupported by design (`FT_Init_FreeType` used; `ftMemory` ignored with `_ = ftMemory`). Color
conversion: all three C++ headers ported (`agg_color_conv.h`, `agg_color_conv_rgb8.h`,
`agg_color_conv_rgb16.h`) to `internal/color/conv/`.

**5.5 Generics / pixfmt**: RGBA16 pixfmt refactored to blender-interface pattern; tests restored. `any()`
assertions eliminated from Gray16/Gray32 pixfmts; `RawRGB/RGBAOrder` fast paths retained as legitimate.
`VertexFilter` shims removed from `array/vertex_sequence.go` and `basics/math_stroke.go`. Color space
kept at Linear + SRGB only.

**Note**: `TestTransformImageUsesPremultipliedRenderer` updated to use premultiplied source input —
C++ AGG routes image rendering through `m_renBasePre` (`agg2d.cpp:1738`) which expects premultiplied
values; no automatic straight→premultiplied conversion occurs in the span path.

---

## Phase 6 - API and Documentation Finalization ✅

- [x] `Context` / `Agg2D` separation documented; package doc in `agg.go` corrected.
- [x] `docs/TASKS.md` synchronized; architecture overview updated to 35 internal packages.
- [x] `docs/AGG_DELTAS.md` created — documents all intentional deviations from C++ AGG 2.6.

---

## Phase 7 - SIMD Infrastructure and Bulk Pixel Paths ✅

`internal/simd/` package with runtime CPU detection, build-tagged arch dispatch, and `purego` scalar baseline.
Four pixel operations each have generic, amd64 (SSE2/SSE4.1/AVX2), and arm64 (NEON/generic) paths.
Assembly in flat `internal/simd/*.s` files (idiomatic Go layout).

- [x] **FillRGBA** — packed-RGBA bulk fill; wired to `CopyHline` / `Clear`.
- [x] **BlendSolidHspanRGBA** — solid-color AA spans with per-pixel cover (SSE4.1 PMAXUW/PMINUW lerp).
- [x] **BlendHlineRGBA** — uniform-coverage hline blend; alpha==255 routes to FillRGBA.
- [x] **BlendColorHspanRGBA** — per-pixel color+cover (scalar IMULQ alpha, SIMD lerp for 8 channels).
- [x] `pixfmt_rgba8.go` fast paths wired for all four operations; RGBA byte order uses SIMD, others fall back to scalar.
- [x] Table-driven tests verify bit-identical output across all forced implementation paths.
- [x] QEMU arm64 correctness checks in regular workflow (`just test-arm64`).

---

## Phase 8 - SIMD Expansion Targets

Each section follows the same three-tier pattern as Phase 7: generic Go → SSE4.1 → AVX2 on amd64; generic fallback → NEON on arm64.

### 8.1 Premultiply / Demultiply

Whole-buffer pre/demultiplication is called on every image load and compositing operation.

- [x] **Generic** — correct scalar baseline with zero-alpha guard on demultiply.
- [x] **SSE4.1 (amd64)** — process 4 pixels/iter: PMULLW × α/255 (AGG rounding); PACKUSWB clamp; alpha channel restored from saved original.
- [x] **AVX2 (amd64)** — delegates to SSE4.1 kernel (bottleneck is memory bandwidth, not arithmetic).
- [x] **NEON (arm64)** — generic fallback (NEON assembly deferred; generic path is correct and tested via QEMU).
- [x] Wire into `pixfmt_rgba8.go` premultiply / demultiply call sites; SIMD fast path for standard RGBA byte order, scalar for BGRA/ARGB/ABGR.
- [x] Table-driven tests: bit-identical output vs. scalar across all paths, including zero-alpha row, boundary alphas, and round-trip precision check.

### 8.2 Composite Blend Modes ✅

Porter-Duff operators beyond `SrcOver` (used by the compositing demos and advanced rendering paths).

- [x] **Generic** — integer-arithmetic scalar for `SrcOver`, `DstOver`, `SrcIn`, `DstIn`, `SrcOut`, `DstOut`, `Xor`, `Clear` in `internal/simd/cpu.go`.
- [x] **SSE4.1 (amd64)** — `SrcOver` 2 pixels/iter via `compSrcOverHspanRGBASSE41Asm` (`comp_src_over_sse41_amd64.s`); formula: `Dca' = Dca + Sca - mul(Dca, Sa)` using PMOVZXBW/PMULLW/PADDW/PSUBW.
- [x] **AVX2 (amd64)** — delegates to SSE4.1 kernel (memory-bandwidth bound, not arithmetic).
- [x] **NEON (arm64)** — generic integer-arithmetic fallback (NEON assembly deferred; correct and tested via QEMU).
- [x] Wire into `pixfmt_composite.go` `BlendHline` and `BlendSolidHspan` fast paths for SrcOver, DstOver, SrcIn, DstIn, SrcOut, DstOut, Xor, Clear with standard RGBA byte order.
- [x] Tests: `TestCompSrcOverHspanRGBAComprehensive` verifies bit-exact (±1) output vs. float64 reference across all forced paths; `TestCompOtherOpsGeneric` covers all 6 non-SIMD ops; `TestCompClearHspanRGBA` covers Clear.

### 8.3 Gradient and Image Span Generation

Span generators feed pixel data into `BlendColorHspan`; their inner loops can be hot for complex scenes.

- [ ] **Generic** — baseline already exists in `internal/span/`; profile before committing to SIMD.
- [ ] **SSE4.1 (amd64)** — linear gradient interpolation: PADDD step accumulation + PSHUFB color lookup if LUT stays hot.
- [ ] **AVX2 (amd64)** — double-width linear interpolation if the SSE4.1 path proves worthwhile.
- [ ] **NEON (arm64)** — `vaddq_s32` step accumulation; generic LUT access (cache-miss bound, may not benefit).
- [ ] Image-filter / resampling kernels: SSE4.1 `PMADDUBSW` dot-product for bilinear tap accumulation.
- [ ] Only implement tiers that show measurable gain in profiling; skip otherwise.

### 8.4 Alpha-Mask Helpers

Alpha-mask operations sit on the scanline-rendering hot path when masks are active.

- [ ] **Generic** — correct scalar baseline for mask fill and RGB-to-gray conversion.
- [ ] **SSE4.1 (amd64)** — mask fill: 16 bytes/iter with MOVDQU store; RGB→gray: `PMADDUBSW` with `[77,150,29,0]` weights (BT.601, scaled).
- [ ] **AVX2 (amd64)** — 32 bytes/iter mask fill; 256-bit RGB→gray with same weight vector.
- [ ] **NEON (arm64)** — `vst1q_u8` mask fill; `vdotq_u32` or manual `vmull`/`vadd` for RGB→gray.
- [ ] Wire into alpha-mask fill and conversion call sites in `internal/pixfmt/`.
- [ ] Tests: byte-exact mask fill across sizes; gray values within ±1 of scalar (rounding may differ by ISA).

### 8.5 Gamma / LUT Application

256-entry LUT lookups are inherently gather-bound; SIMD benefit is limited but worth one profiling pass.

- [ ] Profile gamma application in a representative scene before writing any SIMD code.
- [ ] **SSE4.1 (amd64)** — `PSHUFB`-based 16-entry partial LUT or scalar gather; implement only if profiling justifies.
- [ ] **AVX2 (amd64)** — `VPGATHERDD` gather if available and beneficial; otherwise skip.
- [ ] **NEON (arm64)** — `vtbl` / `vqtbl1q_u8` for 16-entry segments; skip if not hot.
- [ ] If none of the tiers show meaningful gain, mark 8.5 as "profiled, skipped" and close.

---

## Phase 9 - Example and Demo Parity

Primary goal: keep the example surface close to the upstream AGG demo set while remaining idiomatic in Go and supporting both standalone examples and the web demo where it makes sense.

The remaining backlog is smaller than the old example ledger suggested because several demos have already been ported since that file was written.

### 9.1 Example parity infrastructure

- [ ] Keep one authoritative example-parity list in `PLAN.md`.
- [ ] For each newly ported upstream demo:
  - record the C++ source
  - decide whether it belongs in standalone examples, the web demo, or both
  - add a minimal verification path so the demo does not silently rot
- [ ] Reuse shared helpers and assets where possible so new examples do not fragment the example surface.

### 9.2 High-priority remaining demo ports

- [ ] Port the remaining high-value text, image, and stroke demos:
  - `raster_text.cpp`
  - `image_resample.cpp`
  - `gradient_focal.cpp`
  - `line_patterns.cpp`
  - `line_patterns_clip.cpp`
  - `line_thickness.cpp`
- [ ] Port the remaining high-value rendering-pipeline demos:
  - `rasterizer_compound.cpp`
  - `scanline_boolean2.cpp`
  - `pattern_perspective.cpp`
  - `pattern_resample.cpp`
  - `image_perspective.cpp`

### 9.3 Medium-priority interactive and advanced demo ports

- [ ] Port the remaining interactive and geometry-heavy demos:
  - `interactive_polygon.cpp`
  - `graph_test.cpp`
  - `gpc_test.cpp`
  - `gradients_contour.cpp`
- [ ] Port the remaining advanced rendering and math demos:
  - `flash_rasterizer2.cpp`
  - `polymorphic_renderer.cpp`
  - `blend_color.cpp`
  - `image_filters2.cpp`
  - `image_fltr_graph.cpp`

### 9.4 Lower-priority or support-heavy upstream demos

- [ ] Triage the remaining support-heavy demos case by case:
  - `freetype_test.cpp`
  - `truetype_test.cpp`
  - `trans_curve1.cpp`
  - `trans_curve1_ft.cpp`
  - `trans_curve2_ft.cpp`
  - `make_arrows.cpp`
  - `make_gb_poly.cpp`
  - `mol_view.cpp`
  - `idea.cpp`
- [ ] Decide for each whether it should be fully ported, replaced by a Go-idiomatic equivalent, or explicitly deferred with rationale.

### 9.5 Exit criteria

- [ ] Every remaining upstream demo is either ported as a standalone example or web demo, replaced by a documented Go-idiomatic equivalent, or explicitly deferred with rationale.
- [ ] Example coverage reflects current repository reality rather than a stale external ledger.
- [ ] Newly added demos build and run through the existing example workflows.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement/fix Go behavior.
3. Add or update contract tests.
4. Add/update visual regression if rendering-visible.
5. Mark ledger status and `docs/TASKS.md`.
