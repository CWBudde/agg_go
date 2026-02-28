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

- [ ] Replace simplified `renderImage*` implementation with AGG-style scanline/span pipeline:
  - interpolator-based sampling
  - filter LUT integration
  - resample mode behavior (`NoResample`, `ResampleAlways`, `ResampleOnZoomOut`)
  - blend-color conversion path equivalent to AGG behavior
- [ ] Remove nearest-neighbor-only fallback for transformed image rendering.
- [ ] Align transform usage with AGG matrix flow (`parl->world->invert` then interpolator).

Files:

- `internal/agg2d/image.go`
- `internal/span/*` as needed
- `internal/renderer/scanline/*` as needed

### 1.2 Gradient parity

- [ ] Ensure linear/radial gradient matrix construction matches AGG ordering.
- [ ] Remove no-op world/screen helper placeholders and use real transform/scalar conversion.
- [ ] Verify gradient distance (`d1/d2`) handling matches C++ path.

Files:

- `internal/agg2d/gradient.go`
- `internal/agg2d/utilities.go`

### 1.3 Text parity (minimum acceptable)

- [ ] Remove rectangle fallback glyph rendering.
- [ ] Render glyph scanlines/paths through real rasterizer/scanline pipeline.
- [ ] Match vector vs raster cache behavior contract from AGG2D.

Files:

- `internal/agg2d/text.go`
- `internal/font/*` and/or `internal/fonts/*`

### 1.4 Clipping and renderer-state parity

- [ ] Ensure `ClipBox` updates all relevant renderer and rasterizer states consistently.
- [ ] Verify `clearClipBox`, `copyImage`, `blendImage`, transformed image operations obey clip box identically to AGG semantics.

Files:

- `internal/agg2d/buffer.go`
- `internal/agg2d/utilities.go`
- `internal/agg2d/image.go`

### Exit criteria

- [ ] No `simplified`/`for now` rendering paths in `internal/agg2d` critical methods.
- [ ] AGG2D image, gradient, text, clipping contract tests pass.
- [ ] Visual tests for AGG2D demos pass against reference thresholds.

---

## Phase 2 - Core Pipeline Parity (Rasterizer -> Scanline -> Renderer -> Pixfmt)

### 2.1 Rasterizer and scanline correctness

- [ ] Verify fill rules, clipping, cell accumulation, and sweep behavior against AGG reference.
- [ ] Resolve known integration inconsistencies in rendering correctness tests.

### 2.2 Renderer and pixfmt semantics

- [ ] Verify compositing behavior for Porter-Duff modes and alpha paths.
- [ ] Ensure premultiplied vs straight-alpha behavior matches AGG where required.
- [ ] Confirm copy/blend operations align with `agg_renderer_base` semantics.

### 2.3 Converters (conv/vcgen/vpgen) chain fidelity

- [ ] Validate converter ordering and state-machine parity for stroke/dash/curve/contour.
- [ ] Ensure approximation scales are propagated exactly where AGG does.

### Exit criteria

- [ ] Integration tests for full pipeline pass without known behavioral exceptions.
- [ ] Golden image diffs are within agreed threshold.

---

## Phase 3 - Font Subsystem Consolidation and Type Safety

### 3.1 Consolidate `internal/font` vs `internal/fonts`

- [ ] Define a single authoritative font/cache architecture.
- [ ] Remove duplicated concepts and adapters where possible.

### 3.2 Replace runtime `interface{}` where feasible

- [ ] Replace broad `interface{}` in AGG2D font fields with explicit interfaces.
- [ ] Keep runtime dispatch only where build-tag boundaries require it, and document it.

### Exit criteria

- [ ] One coherent font stack is used by AGG2D.
- [ ] No avoidable runtime type assertions in text-critical path.

---

## Phase 4 - Test Strategy for Port Fidelity

### 4.1 Contract tests (API behavior)

- [ ] Expand AGG2D tests to assert outputs, not just `err == nil`.
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

1. [ ] Rebuild `internal/agg2d/image.go` around AGG span-interpolator pipeline.
2. [ ] Fix `internal/agg2d/gradient.go` transform/scalar parity (remove no-op helpers).
3. [x] Replace text rectangle fallback in `internal/agg2d/text.go`.
4. [ ] Align clip propagation across rasterizer and renderer bases in `internal/agg2d/buffer.go`.
5. [ ] Add pixel-asserting AGG2D tests for the above before moving to lower-priority items.
