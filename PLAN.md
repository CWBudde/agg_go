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

## Phase 5 - Remaining Port Inventory from docs/TASKS.md

### 5.1 Transformation backlog

- [ ] Finish perspective and viewport follow-ups:
  - `TransPerspective`: 3D projection simulation, perspective correction, division optimization review, numerical stability review
  - `TransViewport`: multi-viewport support, zoom/pan flow, cached matrix handling, batch transforms
  - connect viewport transforms cleanly to renderer and path-processing call sites
- [ ] Finish path-based transform follow-ups:
  - `TransSinglePath`: preprocessing, quality control, modification hooks, caching
  - `TransSinglePath` integration: vertex-source compatibility, path-converter integration, animation workflow
  - `TransDoublePath`: width calculation, path-relationship metrics, mismatch handling, boundary conditions
  - `TransDoublePath` preprocessing helpers: dual-path preprocessing, batch transformation, later envelope-distortion and flow-field follow-ups
- [ ] Finish warp magnifier follow-ups:
  - lens shape control, dynamic properties, multiple zones
  - edge handling, renderer compatibility, AA integration
  - distortion-math review, coordinate handling, caching, real-time performance

### 5.2 Converter and generator backlog

- [x] Finish adaptor cleanup:
  - `ConvAdaptorVPGen` processor-conformance interface (VPGen interface defined, all vpgen components verified)
  - simpler state-model validation (basic/auto-close/auto-unclose/closed-polygon/empty-path/single-point/multi-rewind tests)
  - compatible-processor coverage (integration tests with ClipPolygon, ClipPolyline, Segmentator + interface-compliance compile check)
  - low-overhead and real-time behavior reviewed (no unnecessary allocations; state reset on Rewind)
- [x] Finish stroke and contour pipeline parity:
  - [x] `conv_stroke`: `InnerJoin` type and propagation through `conv_stroke`/`vcgen_stroke`, with miter-limit regression coverage
  - [x] `conv_stroke`: remaining integration work for rendering-quality params, complex stroke tests, direct access to underlying `vcgen_stroke`, and converter composability
  - [x] `vcgen_stroke`: core struct plus configuration API, source-vertex ingestion, rewind, and vertex output state machine
  - [x] `vcgen_stroke`: complete line-cap and line-join parity audit, inner-join corner handling review, and `math_stroke` integration audit
  - [x] `vcgen_contour`: core struct plus width/join and miter configuration surface
  - [x] `vcgen_contour`: finish vertex ingestion and rewind/vertex output parity, positive/negative/zero width edge cases, contour corner-join behavior, and `math_stroke` integration audit
  - [x] `conv_contour`: text outline generation use case, shape morphing and complex-path scenarios, converter chaining, efficiency and robustness tests
- [x] Finish dash, smoothing, and path-utility converters:
  - `conv_dash`: dash-then-stroke pipeline usage, dynamic dash pattern updates, `DashGenerator()` accessor, robustness (closed/angled/long/offset/shorten) tests
  - `conv_smooth_poly1`: corner detection, selective smoothing, open/closed polygon handling, curve-approximation quality, `ConvSmoothPoly1Curve` approximation methods/scale/tolerance
  - smaller path utility converters: all have implementation and test files (close_polygon, unclose_polygon, concat, shorten_path, segmentator, marker, marker_adaptor, transform, gpc)
- [x] Carry over unresolved placeholder-inventory items:
  - rasterizer cell-run compaction regression tests added for both `RasterizerCellsAASimple` (TestRasterizerCellsAA_RepeatedResetDoesNotDropCells) and `RasterizerCellsAAStyled` (TestRasterizerCellsAAStyled_RepeatedResetDoesNotDropCells)
  - `RenderAllPaths` typing tightened: `MultiPathRasterizerInterface` defined (embeds `RasterizerInterface` + `Reset()` + `AddPath()`), dynamic type assertions removed

### 5.3 Span, interpolator, and image-processing backlog

- [ ] Finish contour and image span generators:
  - `span_gradient_contour`: distance-field preprocessing, contour input methods, multi-contour support, edge handling
  - `span_gradient_image`: pixel-sampling interface, image-coordinate mapping, image-transformation path, caching and memory-management strategy
- [ ] Recheck image-filter edge behavior against AGG:
  - bilinear clip partial-overlap weighted edge sampling
  - background fallback only where AGG does so
- [ ] Finish Gouraud shading completeness:
  - grayscale setup
  - RGBA alpha interpolation and compositing
  - integration coverage
- [ ] Complete contour-gradient parity review:
  - confirm the current outline-rasterizer path matches AGG intent closely enough
  - document or fix remaining deltas in contour creation and distance mapping
- [ ] Finish image-filter and interpolator follow-ups:
  - remaining RGBA image-filter work: four-channel review, alpha optimization, RGBA pixfmt coverage, memory-access-pattern audit
  - perspective interpolator follow-ups: flexible mapping, high-accuracy mode, adaptive accuracy, complex projection coverage
  - transform interpolator follow-ups: transformation-overhead management, non-linear transformation support
- [ ] Finish utility math follow-ups:
  - transformed bounding rectangle
  - `shorten_path` edge cases, vertex-sequence compatibility, stroke integration

### 5.4 Font and utility backlog

- [ ] Resolve the `RowPtr` bridge decision in `internal/agg2d/adapters.go`.
- [ ] Finish embedded raster-font integration:
  - document font-data format
  - simplify glyph-access interfaces where justified
  - recheck rendering integration
- [ ] Decide whether to implement the FreeType custom memory-management hook or document it as unsupported.
- [ ] Port remaining color-conversion surfaces:
  - `agg_color_conv.h`
  - `agg_color_conv_rgb8.h`
  - `agg_color_conv_rgb16.h`

### 5.5 Generics and pixfmt refactoring backlog

These items were previously tracked in separate generics TODO and audit notes and now live directly in the phased plan.

- [ ] Complete the RGBA16 pixfmt generics refactor:
  - apply the RGBA8 blender-interface pattern to `pixfmt_rgba16.go`
  - finish method-signature cleanup
  - restore deterministic tests
- [ ] Resolve generics-related example compatibility:
  - update examples using older pixfmt generic signatures
  - verify constructors and type aliases compile cleanly
- [ ] Decide whether to expand `color.Space` beyond `Linear` and `SRGB`.
- [ ] Validate generics fast-path performance claims:
  - profile interface-based blender dispatch vs direct access
  - keep `RawRGBAOrder` only where measurable
- [ ] Retire deprecated generic-era sequence shims:
  - ensure callers use `VertexDistSequence` / `LineAAVertexSequence`
  - remove compatibility-only paths once no callers remain
- [ ] Re-audit remaining `any()`-based generic dispatch and classify each site:
  - keep legitimate optimization or boundary-adaptation checks
  - convert true type-dispatch cases into concrete or typed APIs
  - explicitly review serialized scanline storage and span color-conversion boundaries

### Exit criteria

- [ ] Every unfinished `docs/TASKS.md` item is either completed, explicitly deferred, or linked to one of the grouped tasks above.
- [ ] Remaining generator and converter gaps no longer block AGG feature parity.
- [ ] Utility and font leftovers are either implemented or documented as intentional deltas.

---

## Phase 6 - API and Documentation Finalization

### 6.1 API cleanup

- [ ] Keep high-level `Context` ergonomic and clearly separated from low-level AGG2D behavior.
- [ ] Ensure naming is idiomatic Go without losing AGG traceability in docs.

### 6.2 Documentation hygiene

- [ ] Keep `docs/TASKS.md` and `docs/TASKS-COMPLETED.md` synchronized with completed items.
- [ ] Update architecture docs after each completed phase.
- [ ] Add explicit “known deltas from AGG” section.

### Exit criteria

- [ ] Public API docs and examples align with actual behavior.
- [ ] All completed tasks marked in `docs/TASKS.md`.

---

## Phase 7 - SIMD Infrastructure and Bulk Pixel Paths

Primary goal: accelerate hot rendering paths without changing rendering semantics.

### 7.0 Architecture and constraints

- [x] Add a separate `internal/simd/` package with runtime CPU detection and scalar fallback.
- [x] Reuse the `algo-fft` pattern: cached feature detection, build-tagged arch dispatch, and test overrides.
- [x] Respect `purego` as a reliable scalar baseline.
- [ ] Add real Plan 9 assembly entry points under `internal/simd/asm_amd64/` and `internal/simd/asm_arm64/`.

Planned package layout:

- `cpu.go`
- `detect_amd64.go`
- `detect_arm64.go`
- `detect_generic.go`
- `blend_amd64.go`
- `blend_arm64.go`
- `blend_generic.go`
- `asm_amd64/`
- `asm_arm64/`

### 7.1 Phase 1a - CopyHline / Clear

Profile and intent: `CopyHline` and `Clear` are the safest first SIMD target because they are bulk fill operations with straightforward byte-exact validation.

- [x] Introduce a packed-RGBA fill primitive.
- [x] Dispatch across generic scalar, amd64 (`sse2` / `avx2`), and arm64 (`neon`) paths.
- [x] Wire `PixFmtAlphaBlendRGBA.CopyHline` and `Clear` through the new primitive.
- [x] Add deterministic tests for CPU detection and forced-feature dispatch, byte-exact fill behavior, and arm64/QEMU execution.
- [x] Replace the current arch wrappers with real assembly implementations for SSE2, AVX2, and NEON.

### 7.2 Phase 1b - BlendSolidHspan

Why next: this is one of the hottest AA rendering paths and gives useful SIMD coverage beyond pure fill loops.

- [x] SIMD-optimize solid-color horizontal spans with per-pixel cover.
- [x] Preserve bit-identical output with the scalar implementation.
- [x] Verify against direct byte and pixel assertions, plus the current focused SIMD validation suite.

### 7.3 Phase 1c - BlendHline

- [ ] SIMD-optimize uniform-coverage horizontal blending.
- [ ] Share structure with the `BlendSolidHspan` implementation where it stays readable.

### 7.4 Phase 1d - BlendColorHspan

- [ ] SIMD-optimize per-pixel color plus coverage blending for gradients and image rendering.

### Exit criteria

- [ ] Generic, amd64, and arm64 paths are all present.
- [ ] `purego` remains a reliable scalar baseline.
- [ ] QEMU-backed arm64 correctness checks are part of the regular workflow.
- [ ] SIMD paths have deterministic correctness coverage before any benchmarking claims.

---

## Phase 8 - SIMD Expansion Targets

### 8.1 Premultiply / Demultiply

- [ ] SIMD-optimize whole-buffer RGBA premultiplication.
- [ ] SIMD-optimize demultiplication with correct zero-alpha handling.

### 8.2 Composite blend modes

- [ ] Optimize the important Porter-Duff and SVG composite operators.
- [ ] Prioritize modes used by compositing demos and common `SrcOver`-adjacent paths.

### 8.3 Gradient and image span generation

- [ ] Optimize gradient span generation where interpolation and LUT access make this worthwhile.
- [ ] Optimize image-filter and resampling kernels after the earlier bulk pixel paths land.

### 8.4 Alpha-mask helpers

- [ ] SIMD-optimize alpha-mask fill operations.
- [ ] SIMD-optimize RGB-to-gray mask conversion helpers where they remain hot.

### 8.5 Gamma / LUT application

- [ ] Only pursue gamma and LUT SIMD work if profiling still shows it matters after earlier SIMD phases.

### Recommended SIMD implementation order

1. AVX2 on amd64
2. SSE2 on amd64
3. NEON on arm64

Rationale:

- AVX2 is the most practical development target and usually offers the highest immediate payoff.
- SSE2 is the amd64 baseline that keeps older CPUs covered.
- NEON matters for Apple Silicon and Linux arm64; QEMU or native validation keeps that path honest.

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
