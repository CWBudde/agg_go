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

## Phase 0 - Baseline and Traceability

### 0.1 Parity ledger

- [x] Create a parity ledger with one row per `Agg2D` method.
- [x] Columns: C++ source method, Go method, status (`exact`, `close`, `placeholder`, `missing`), test reference, notes.
- [x] Add source anchors for key AGG2D methods in:
  - `../agg-2.6/agg-src/agg2d/agg2d.h`
  - `../agg-2.6/agg-src/agg2d/agg2d.cpp`
- [x] Fold the remaining open parity rows into Phase 4 so parity tracking lives in the main plan rather than a separate ledger file.

### 0.2 Placeholder inventory

- [x] Record all placeholder/simplified paths in rendering-critical packages:
  - `internal/agg2d`
  - `internal/rasterizer`
  - `internal/scanline`
  - `internal/renderer`
  - `internal/span`
- [x] Classify each as `must-fix`, `acceptable temporary`, or `low-priority`.

### Exit criteria

- [x] Ledger exists and covers all `Agg2D` public operations.
- [x] Placeholder inventory is complete and prioritized.

---

## Phase 1 - AGG2D Behavioral Parity (Highest Priority)

Primary target: `internal/agg2d/*` against `agg2d.cpp`.

### 1.1 Image pipeline parity (critical)

- [x] Replace simplified `renderImage*` implementation with AGG-style scanline/span pipeline:
  - interpolator-based sampling
  - filter LUT integration
  - resample mode behavior (`NoResample`, `ResampleAlways`, `ResampleOnZoomOut`)
  - blend-color conversion path equivalent to AGG behavior
- [x] Remove nearest-neighbor-only fallback for transformed image rendering.
- [x] Align transform usage with AGG matrix flow (`parl->world->invert` then interpolator).

Files:

- `internal/agg2d/image.go`
- `internal/span/*` as needed
- `internal/renderer/scanline/*` as needed

### 1.2 Gradient parity

- [x] Ensure linear/radial gradient matrix construction matches AGG ordering.
- [x] Remove no-op world/screen helper placeholders and use real transform/scalar conversion.
- [x] Verify gradient distance (`d1/d2`) handling matches C++ path.

Files:

- `internal/agg2d/gradient.go`
- `internal/agg2d/utilities.go`

### 1.3 Text parity (minimum acceptable)

- [x] Remove rectangle fallback glyph rendering.
- [x] Render glyph scanlines/paths through real rasterizer/scanline pipeline.
- [x] Match vector vs raster cache behavior contract from AGG2D.

Files:

- `internal/agg2d/text.go`
- `internal/font/*` and/or `internal/fonts/*`

Notes:

- Raster glyph rendering now routes through scanline renderers (`RenderScanlinesAASolid`/`RenderScanlinesBinSolid`) via a glyph rasterizer adapter instead of direct per-row blending.
- Outline glyph positioning now follows AGG2D's embedded-adaptor contract (per-glyph translation + text-angle transform path).
- Kerning in `TextWidth`/`Text` now uses glyph indices, and outline cache hits refresh engine outline state before adaptor initialization.

### 1.4 Clipping and renderer-state parity

- [x] Ensure `ClipBox` updates all relevant renderer and rasterizer states consistently.
- [x] Verify `clearClipBox`, `copyImage`, `blendImage`, transformed image operations obey clip box identically to AGG semantics.
      Verified by dedicated pixel-asserting tests in `internal/agg2d/{agg2d,image,utilities}_test.go`.
- [x] Align `clearAll`, `inBox`, `alignPoint`, fill-rule updates, attach-time gamma reset, and master-alpha/gamma rasterizer behavior with `agg2d.cpp`.

Files:

- `internal/agg2d/buffer.go`
- `internal/agg2d/utilities.go`
- `internal/agg2d/image.go`
- `internal/agg2d/rendering.go`
- `internal/agg2d/fill_rules.go`

### Exit criteria

- [x] No `simplified`/`for now` rendering paths in `internal/agg2d` critical methods.
- [x] AGG2D image, gradient, text, clipping contract tests pass.
- [ ] Visual tests for AGG2D demos pass against reference thresholds.

---

## Phase 2 - Core Pipeline Parity (Rasterizer -> Scanline -> Renderer -> Pixfmt)

### 2.1 Rasterizer and scanline correctness

- [x] Align fill rules, clipping edge cases, cell accumulation, and sweep indexing with AGG reference in the core AA rasterizers.
- [x] Preserve AGG duplicate-cell behavior in sorted cell stores and compound scanline handling.
- [x] Resolve known integration inconsistencies caused by non-AGG test-driver assumptions (`Rectangle()+DrawPath()`, missing even-odd enablement, invalid star geometry).
- [x] Continue auditing `compound_aa` and related rasterizer paths for any remaining source-level edge cases not yet covered by direct parity tests.

### 2.2 Renderer and pixfmt semantics

- [x] Confirm copy/blend overlap behavior aligns with `agg_renderer_base` semantics in renderer base and concrete RGB/RGBA pixfmts.
- [x] Align premultiplied vs straight-alpha behavior for Agg2D image rendering and composite image paths.
- [x] Port the core `copy_from` / `blend_from` helper surface needed by RGBA, RGB, Gray, transposer, amask, and composite pixfmts.
- [x] Expand parity coverage for Porter-Duff and non-`BlendAlpha` composite behavior against C++ reference outputs, especially outside the currently covered image-path cases.

### 2.3 Converters (conv/vcgen/vpgen) chain fidelity

- [x] Restore AGG2D stroke/dash/transform ordering and line cap/join enum parity.
- [x] Align viewport, gradient, and related transform/scalar propagation with AGG2D behavior.
- [x] Audit lower-level converter/vcgen/vpgen state machines beyond the Agg2D call sites, especially stroke/dash/curve/contour behavior and approximation-scale propagation.

### Exit criteria

- [x] Integration tests for full pipeline pass without known behavioral exceptions.
- [x] Golden image diffs are within agreed threshold.

---

## Phase 3 - Font Subsystem Consolidation and Type Safety

### 3.1 Consolidate `internal/font` vs `internal/fonts`

- [x] Define a single authoritative font/cache architecture.
- [x] Remove duplicated concepts and adapters where possible.
- [x] Audit remaining `internal/font/freetype2` convenience wrappers against `agg_font_freetype2.h/.cpp` and keep only the abstractions that are justified in Go.
      The former exported FontManager is now package-local, CacheManager2 has been reduced to a thin adaptor-facing wrapper over `internal/fonts.FmanCachedFont`, concrete gray8/mono adaptor wrapper types are no longer exposed, and test-only engine-selection/path-storage helpers are no longer exported. The remaining gray8/mono wrappers are retained as the minimal package boundary because `internal/scanline` exposes concrete serialized-scanline iteration APIs while `internal/fonts` expects generic adaptor/span interfaces.
- [x] Finish separating "Agg2D text path" vs "standalone `fman`/embedded-font support" in remaining docs/review notes.
- [x] Continue rechecking FreeType2 face/engine lifetime behavior against AGG, especially around multi-face ownership beyond the now-fixed unload/close ownership semantics.
      Engine-driven multi-face close now releases all tracked faces correctly. The explicit `maxFaces` cap is documented as an intentional Go-only policy delta, and `engine.Close()` actively closes tracked faces before freeing the library, which is safer than AGG's looser caller-owned loaded_face lifetime model but not a direct behavior match.

### 3.2 Replace runtime `interface{}` where feasible

- [x] Replace broad `interface{}` in AGG2D font fields with explicit interfaces.
- [x] Keep runtime dispatch only where build-tag boundaries require it, and document it.
- [x] Re-audit the generics-related review notes so they reflect the current typed state of the font subsystem.
- [x] Check FreeType2/CGO-adjacent font code for any remaining avoidable dynamic dispatch or stale comments claiming broader type erasure than the code now uses.
- [x] Keep the remaining signature mismatches between neighboring internal interfaces localized behind narrow adapters rather than widening the font API surface.
      The old FreeType2 adaptor bundle has been removed; CacheManager2 now depends on narrow per-adaptor methods plus package-local wrappers instead of a broader concrete-type aggregate.

### Exit criteria

- [x] One coherent font stack is used by AGG2D.
- [x] No avoidable runtime type assertions in text-critical path.
- [x] `internal/font/freetype2` is either brought closer to AGG's `fman` API surface or its remaining Go-only convenience APIs (`FontManager`, engine-selection helpers, thin adaptor wrappers) are explicitly documented as intentional deltas.
- [x] Embedded raster font data and cache behavior are rechecked against AGG/review notes so Phase 3 closes without known font-subsystem placeholders.
- [x] FreeType2 glyph-cache tests cover native and AGG gray/mono plus outline serialization paths for a real font when available.

---

## Phase 4 - Test Strategy for Port Fidelity

### 4.1 Contract tests (API behavior)

- [x] Expand AGG2D tests to assert outputs, not just `err == nil`, for the currently covered rendering paths.
      `internal/agg2d/rendering_test.go`, `internal/agg2d/image_test.go`, and `internal/agg2d/text_phase1_test.go` now use deterministic output assertions for solid fill, gradient fill, translated rendering output, clipped fill/stroke rendering, blend-mode compositing, transformed image placement/color coverage, and vector-text alignment/bounds.
- [x] Add deterministic checks for transform-image, clipping, blend modes, gradients, and text bounds.
- [ ] Continue replacing remaining AGG2D smoke/integration tests with output assertions where they still only verify "no crash" behavior.
      Remaining likely targets include broader path/image integration cases in `internal/agg2d/rendering_fixes_test.go` and any text rendering paths that still depend on loose FreeType/system-font checks rather than deterministic bounds or pixel contracts.
- [ ] Expand contract coverage for currently weaker packages:
  - `internal/effects`
  - `internal/platform`
  - `internal/primitives`
  - `internal/pixfmt/blender`
  - any other package below the desired coverage floor after parity-critical work
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

- [ ] Finish adaptor cleanup:
  - `ConvAdaptorVPGen` processor-conformance interface
  - simpler state-model validation
  - compatible-processor coverage
  - low-overhead and real-time behavior review
- [ ] Finish stroke and contour pipeline parity:
  - [x] `conv_stroke`: `InnerJoin` type and propagation through `conv_stroke`/`vcgen_stroke`, with miter-limit regression coverage
  - [ ] `conv_stroke`: remaining integration work for rendering-quality params, complex stroke tests, direct access to underlying `vcgen_stroke`, and converter composability
  - [x] `vcgen_stroke`: core struct plus configuration API, source-vertex ingestion, rewind, and vertex output state machine
  - [ ] `vcgen_stroke`: complete line-cap and line-join parity audit, inner-join corner handling review, and `math_stroke` integration audit
  - [x] `vcgen_contour`: core struct plus width/join and miter configuration surface
  - [ ] `vcgen_contour`: finish vertex ingestion and rewind/vertex output parity, positive/negative/zero width edge cases, contour corner-join behavior, and `math_stroke` integration audit
  - [ ] `conv_contour`: text outline generation use case, shape morphing and complex-path scenarios, converter chaining, efficiency and robustness tests
- [ ] Finish dash, smoothing, and path-utility converters:
  - `conv_dash`: explicit dash-then-stroke usage coverage, dynamic dash updates, robustness tests, underlying `vcgen_dash` access
  - `conv_smooth_poly1`: corner detection and selective smoothing, polygon-path review, curve-approximation quality checks, converter-pipeline integration
  - smaller path utility converters:
    - `conv_close_polygon`: path analysis, path modification, efficient closure
    - `conv_unclose_polygon`: path integrity, open-path creation, usage coverage
    - `conv_concat`: complex path construction, concatenation efficiency
    - `conv_shorten_path`: arc-length handling, boundary conditions, efficient shortening
    - `conv_segmentator`: segment control, uniform output, quality control
    - `conv_marker` and `conv_marker_adaptor`: path processing, marker extensibility, efficiency
    - `conv_transform`: streaming-transform behavior, command integrity, renderer compatibility
  - lower-priority converter items: `conv_gpc` advanced polygon features and compatibility notes
- [ ] Carry over unresolved placeholder-inventory items:
  - recheck rasterizer cell-run compaction behavior in `RasterizerCellsAASimple` and `RasterizerCellsAAStyled`, then add regression tests
  - tighten `RenderAllPaths` typing in `internal/renderer/scanline/helpers.go`

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

---

## Immediate Execution Queue (Start Here)

1. [x] Rebuild `internal/agg2d/image.go` around AGG span-interpolator pipeline.
2. [x] Fix `internal/agg2d/gradient.go` transform/scalar parity (remove no-op helpers).
3. [x] Replace text rectangle fallback in `internal/agg2d/text.go`.
4. [x] Align clip propagation across rasterizer and renderer bases in `internal/agg2d/buffer.go`.
5. [x] Add pixel-asserting AGG2D tests for the above before moving to lower-priority items.
       Current coverage includes clip/copy assertions and deterministic image sampling assertions; broader end-to-end render-output parity remains tracked in Phase `4.1`.
6. [x] Implement `vcgen_stroke` core struct plus configuration API from Phase 5.2.
7. [x] Port `InnerJoin` through `conv_stroke` and `vcgen_stroke`, then add miter-limit regression tests.
8. [x] Implement `vcgen_contour` core struct plus width/join handling from Phase 5.2.
9. [ ] Close the remaining AGG2D smoke tests from Phase 4.1 that still only assert "no crash".
10. [x] Replace the Phase 7.1 arch wrappers in `internal/simd/` with real Plan 9 asm for `fillRGBA`.
11. [ ] Implement Phase 7.2 SIMD `BlendSolidHspan` and validate against scalar output plus visual tests.
        NEON arm64 now uses run-fill hybrid (NEON fill for solid-coverage runs, generic for partial).
        Comprehensive validation test suite added covering 15 scenarios across all implementations.
        Found and fixed AVX2 register-clobber bug in 8-pixel loop (X4/Y4 aliasing in both alpha and opaque paths).
12. [ ] Port one high-value missing demo from Phase 9.2, preferring `raster_text.cpp` or `image_resample.cpp`.
