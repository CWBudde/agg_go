# AGG Go Port - Fidelity-First Plan

## Objective

Port AGG 2.6 to Go so that:

1. Rendering behavior stays as close as possible to original AGG (`../agg-2.6/agg-src`).
2. Go code remains idiomatic, maintainable, and testable.
3. Deviations from AGG are explicit, justified, and tested.

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

- [x] Create `docs/PARITY_LEDGER.md` with one row per `Agg2D` method.
- [x] Columns: C++ source method, Go method, status (`exact`, `close`, `placeholder`, `missing`), test reference, notes.
- [x] Add source anchors for key AGG2D methods in:
  - `../agg-2.6/agg-src/agg2d/agg2d.h`
  - `../agg-2.6/agg-src/agg2d/agg2d.cpp`

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
- [ ] Continue auditing `compound_aa` and related rasterizer paths for any remaining source-level edge cases not yet covered by direct parity tests.

### 2.2 Renderer and pixfmt semantics

- [x] Confirm copy/blend overlap behavior aligns with `agg_renderer_base` semantics in renderer base and concrete RGB/RGBA pixfmts.
- [x] Align premultiplied vs straight-alpha behavior for Agg2D image rendering and composite image paths.
- [x] Port the core `copy_from` / `blend_from` helper surface needed by RGBA, RGB, Gray, transposer, amask, and composite pixfmts.
- [ ] Expand parity coverage for Porter-Duff and non-`BlendAlpha` composite behavior against C++ reference outputs, especially outside the currently covered image-path cases.

### 2.3 Converters (conv/vcgen/vpgen) chain fidelity

- [x] Restore AGG2D stroke/dash/transform ordering and line cap/join enum parity.
- [x] Align viewport, gradient, and related transform/scalar propagation with AGG2D behavior.
- [ ] Audit lower-level converter/vcgen/vpgen state machines beyond the Agg2D call sites, especially stroke/dash/curve/contour behavior and approximation-scale propagation.

### Exit criteria

- [ ] Integration tests for full pipeline pass without known behavioral exceptions.
- [ ] Golden image diffs are within agreed threshold.

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
- [x] Re-audit `docs/GENERICS_AUDIT.md` and related notes so they reflect the current typed state of the font subsystem.
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

- [ ] Expand AGG2D tests to assert outputs, not just `err == nil`.
      Started by converting `internal/agg2d/rendering_test.go`, `internal/agg2d/image_test.go`, and `internal/agg2d/text_phase1_test.go` from smoke-style "no crash" checks to deterministic output assertions for solid fill, gradient fill, translated rendering output, clipped fill/stroke rendering, blend-mode compositing, transformed image placement/color coverage, and vector-text alignment/bounds.
- [ ] Add deterministic checks for transform-image, clipping, blend modes, gradients, text bounds.

### 4.2 Visual regression tests

- [ ] Generate canonical references from C++ AGG for core scenarios.
- [ ] Store references under `tests/visual/reference`.
- [ ] Add automated diff thresholding and report generation.

### 4.3 C++ parity checks

- [ ] For each parity-ledger row marked `exact`, include at least one source-linked test case.
- [ ] For rows marked `close`, include documented rationale.

### Exit criteria

- [ ] `go test ./...` passes.
- [ ] Visual regression suite passes in CI.
- [ ] Parity ledger has no untriaged `placeholder` entries.

---

## Phase 5 - API and Documentation Finalization

### 5.1 API cleanup

- [ ] Keep high-level `Context` ergonomic and clearly separated from low-level AGG2D behavior.
- [ ] Ensure naming is idiomatic Go without losing AGG traceability in docs.

### 5.2 Documentation hygiene

- [ ] Keep `docs/TASKS.md` and `docs/TASKS-COMPLETED.md` synchronized with completed items.
- [ ] Update architecture docs after each completed phase.
- [ ] Add explicit “known deltas from AGG” section.

### Exit criteria

- [ ] Public API docs and examples align with actual behavior.
- [ ] All completed tasks marked in `docs/TASKS.md`.

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
