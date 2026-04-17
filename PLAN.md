# AGG Go Port - Fidelity-First Plan

## Objective

Port AGG 2.6 to Go so that:

1. Rendering behavior stays as close as possible to original AGG (`../agg-2.6/agg-src`).
2. Go code remains idiomatic, maintainable, and testable.
3. Deviations from AGG are explicit, justified, and tested.

This document tracks only unresolved work. Completed work is intentionally omitted so the
remaining plan stays focused and actionable. Intentional deviations belong in
`docs/AGG_DELTAS.md`.

## Non-Negotiables

- Every remaining behavioral gap maps to a C++ source reference.
- No placeholder rendering paths in production-critical pipeline stages.
- Public API remains stable and idiomatic; internal architecture may change.

## Porting Rules

1. Fidelity first for algorithms and numeric behavior.
2. Idiomatic Go for ownership, naming, package boundaries, and tests.
3. No silent fallbacks that change rendering semantics.
4. If behavior differs from AGG, document it in `docs/AGG_DELTAS.md`.

---

## Phase 1 - Visual Regression and Demo Parity

This is the main remaining parity gate. The visual corpus is still the best way to catch
integration-level mismatches, especially where the code is already functionally correct but
still differs from upstream in positioning, orientation, clipping, or reference-frame setup.

### 1.1 Visual corpus and workflow

- [ ] Bring `tests/visual/demo_parity_test.go` to green against the C++ references.
- [ ] Add a controlled reference-regeneration and approval workflow under `tests/visual/`.
- [ ] Keep per-demo parity notes and a minimal verification path for every open demo.
- [ ] Add source-linked test coverage for every parity row marked `exact`.
- [ ] Add a documented rationale for every parity row marked `close`.
- [ ] Centralize visual references and the approval workflow under `tests/visual/`.

### 1.2 Remaining demo mismatches

Keep the remaining corpus of demo mismatches under active repair:

- [ ] `aa_demo`
- [ ] `aa_test`
- [ ] `alpha_gradient`
- [ ] `alpha_mask`
- [ ] `alpha_mask2`
- [ ] `alpha_mask3`
- [ ] `bezier_div`
- [ ] `blend_color`
- [ ] `blur`
- [ ] `bspline`
- [ ] `circles`
- [ ] `component_rendering`
- [ ] `compositing`
- [ ] `compositing2`
- [ ] `conv_contour`
- [ ] `conv_dash_marker`
- [ ] `conv_stroke`
- [ ] `distortions`
- [ ] `flash_rasterizer`
- [ ] `flash_rasterizer2`
- [ ] `gamma_correction`
- [ ] `gamma_ctrl`
- [ ] `gamma_tuner`
- [ ] `gouraud_mesh`
- [ ] `gradient_focal`
- [ ] `gradients_contour`
- [ ] `gradients`
- [ ] `graph_test`
- [ ] `idea`
- [ ] `image_alpha`
- [ ] `image_filters`
- [ ] `image_filters2`
- [ ] `image_fltr_graph`
- [ ] `image_perspective`
- [ ] `image_resample`
- [ ] `image_transforms`
- [ ] `image1`
- [ ] `line_patterns_clip`
- [ ] `line_patterns`
- [ ] `line_thickness`
- [ ] `lion_lens`
- [ ] `lion_outline`
- [ ] `lion`
- [ ] `mol_view`
- [ ] `multi_clip`
- [ ] `pattern_fill`
- [ ] `pattern_perspective`
- [ ] `pattern_resample`
- [ ] `perspective`
- [ ] `polymorphic_renderer`
- [ ] `raster_text`
- [ ] `rasterizer_compound`
- [ ] `rasterizers`
- [ ] `rasterizers2`
- [ ] `rounded_rect`
- [ ] `scanline_boolean`
- [ ] `scanline_boolean2`
- [ ] `simple_blur`
- [ ] `trans_polar`

### 1.3 Exit criteria

- [ ] Visual regression suite passes in CI.
- [ ] No AGG2D parity row remains untriaged or placeholder-level.
- [ ] Visual references and approval workflow are centralized under `tests/visual/`.

---

## Phase 2 - Demo-Specific Fixes

The demos below are not just "visual noise"; each one still encodes a concrete porting issue:
asset selection, input mapping, coordinate-frame handling, canvas orientation, or demo-specific
state initialization. The goal here is to close them one by one and attach a minimal verification
path to each fix.

### 2.1 Open fixes

- [ ] `trans_curve`: evaluate the source bitmap choice and switch to a better upstream-compatible
      shared asset if one exists.
- [ ] `trans_curve2`: same asset/parity task as `trans_curve`.
- [ ] `image_resample`: restore draggable quad handles and ensure down/move/up handlers map to
      this demo correctly.
- [ ] `image_perspective`: add or fix draggable quad handles and mouse-interaction wiring.
- [ ] `gamma_correction`: correct the quadrant placement so the background matches the C++ frame.
- [ ] `compositing2`: expand rendering so the composited circles occupy the intended canvas region.
- [ ] `aatest`: restore the expected grey background.
- [ ] `gradients`: center the sphere instead of leaving it far to the right.
- [ ] `flash_rasterizer` and `flash_rasterizer2`: verify whether the blank C++ reference frame is
      intentional or whether the Go port still has a shape-index initialization bug.
- [ ] `gouraud_mesh`: restore the full-sized layout and the statistics text at the bottom.

### 2.2 Required follow-up for each fix

- [ ] Add a standalone-vs-web parity note.
- [ ] Add a minimal verification path such as a render smoke test or bounded image-hash check.
- [ ] Record the C++ source reference for the relevant behavior.

---

## Phase 3 - Font Fidelity

The font subsystem is mostly in good shape. The recent FreeType and low-level
runner fixes appear to have addressed the main baseline/orientation regression
path (notably the bitmap-bounds/row-order handling and the earlier double-flip
/ inverted-baseline behavior). The remaining work here is to lock that in with
explicit regression coverage and, if practical, a direct C++ comparison.

### 3.1 Open bug

- [x] Fix the FreeType raster glyph vertical baseline bug where short glyphs such as `.`, `,`, and
      `-` render at the wrong Y position when using `RasterFontCache`.
- [x] Re-check the glyph bitmap positioning pipeline around `InitEmbeddedAdaptors`,
      `NewSerializedScanlinesAdaptorAA`, and `glyphBitmapRasterizer.SweepScanline`.
- [ ] Add a pixel-level regression test for a mixed-height string such as `0.2 H,x-y`.
- [ ] Add a C++ comparison for the same string and font size if practical.

### 3.2 Acceptance criteria

- [x] Short glyphs land on the baseline band rather than above x-height.
- [ ] The regression is covered by a stable image- or pixel-level test.
- [ ] Any intentional deviation is documented in `docs/AGG_DELTAS.md`.

---

## Phase 4 - Explicit Float Agg2D Variant

AGG 2.6's `Agg2D` has a compile-time `AGG2D_USE_FLOAT_FORMAT` switch in
`../agg-2.6/agg-src/agg2d/agg2d.h` that swaps the internal `ColorType` from
`agg::rgba8` to `agg::rgba32`. The Go port should support the same mode as an
explicit, dedicated implementation path, not via build tags and not by mutating
the semantics of the current 8-bit `Agg2D`.

### 4.1 Goal

- [ ] Add a dedicated float-backed `Agg2D` implementation path that mirrors the
      existing 8-bit `Agg2D` API closely enough for side-by-side parity work.
- [ ] Keep the current 8-bit `Agg2D` behavior and public API stable.
- [ ] Make the float path explicit in naming and construction so callers opt in
      intentionally rather than changing global build configuration.

### 4.2 Shape of the work

- [ ] Introduce a separate implementation rather than a runtime mode flag on the
      existing `Agg2D`. The working assumption is a dedicated copy/twin such as
      `Agg2DFloat`, `ContextFloat`, and `ImageFloat`, with shared helpers only
      where behavior is identical.
- [ ] Mirror the C++ `AGG2D_USE_FLOAT_FORMAT` intent specifically at the
      Agg2D-layer pixel-format wiring: the float variant should use the
      `rgba32`/`rgba128`-class pixel-format stack for its internal renderers
      while preserving the current path semantics around blend/plain/pre modes.
- [ ] Keep the 8-bit and float implementations source-comparable. If code is
      factored, the factoring must preserve a clear one-to-one mapping back to
      C++ Agg2D methods and state.
- [ ] Avoid build tags for selecting the float path. Selection must happen via
      explicit constructors/types in normal Go code.

### 4.3 Required scope

- [ ] Provide dedicated float image and context types with attach/create APIs,
      not just a hidden internal renderer.
- [ ] Cover at least these Agg2D subsystems in the float variant:
      clear/fill/stroke, path rendering, image copy/blend/transform, gradients,
      and text state plumbing.
- [ ] Define and document the boundary conversions between float Agg2D images
      and standard Go image types or 8-bit AGG images.
- [ ] Add an explicit contract for premultiply/demultiply behavior in the float
      variant, including whether exported helper APIs expose straight or premultiplied
      data at the boundary.

### 4.4 Non-goals

- [ ] Do not replace the existing 8-bit `Agg2D`.
- [ ] Do not hide both precisions behind one opaque constructor unless parity
      and debugging remain straightforward.
- [ ] Do not treat this as a generic whole-library precision rewrite. This task
      is specifically about an explicit Agg2D-level float twin first.
- [ ] Do not add a 16-bit Agg2D variant in the same change unless it falls out
      naturally from shared lower-level abstractions after the float path is in
      place and tested.

### 4.5 Verification and exit criteria

- [ ] Add side-by-side tests that render the same scene through 8-bit Agg2D and
      float Agg2D, then compare the quantized float output against expected
      tolerance envelopes.
- [ ] Add source-linked tests for premultiply/demultiply and transformed-image
      behavior in the float path.
- [ ] Add at least one visual regression/demo hook that can run a demo via the
      float Agg2D path without disturbing the existing 8-bit baseline.
- [ ] Document the intentional API and behavioral differences between the 8-bit
      and float variants in `docs/AGG_DELTAS.md` or a dedicated companion note.

---

## Phase 5 - In-Repo Dual Engine Integration and AGoGo Absorption

All implementation work for the final library should happen in this repository.
`../AGoGo` is now only a migration source and temporary oracle: useful for
auditing existing C++ FFI glue, tests, docs, benchmarks, and edge cases, but
not a long-term runtime dependency. The end state is:

1. This repository contains the pure-Go port and any optional C++ reference
   engine support needed for comparison or performance work.
2. `../AGoGo` is redundant.
3. This repository is renamed back to `AGoGo` once the migration is complete.

The goal of this phase is therefore twofold: keep a high-level way to choose
either the Go port or a C++ AGG-backed path when useful, while ensuring that
all such code, tooling, and maintenance live here rather than being split
across two active repositories.

### 5.1 Goal

- [x] Create a first-cut in-repo `engine` facade with a working port-backed
      implementation and an explicit typed unavailable path for the future C++
      engine.
- [ ] Add an opt-in high-level engine-selection layer in this repository that
      can render through either the native `agg_go` port or an in-repo C++ AGG
      reference/backend path.
- [ ] Keep `agg_go` as the canonical, default, dependency-light engine.
- [ ] Use `../AGoGo` only as a migration source until equivalent or better
      functionality, tests, and docs live here.
- [ ] End with a single maintained repository that can be renamed back to
      `AGoGo` without losing functionality.

### 5.2 Canonical API boundary

- [x] Keep the existing root `agg` package concrete and pure Go.
- [x] Introduce a separate backend-selectable facade package, tentatively
      `engine`, instead of mutating `agg.Agg2D`, `agg.Context`, or `agg.Image`
      into interface-first types.
- [x] Keep current `agg_go` callers source-compatible unless they explicitly opt
      into backend selection.
- [x] Do not add direct cgo-backed implementation files to the root `agg`
      package.
- [x] Do not require the final public API to depend on the external
      `../AGoGo` repository or module path.

### 5.3 Required facade shape

- [x] Define an explicit backend enum such as `engine.Kind` with at least
      `Port` and `AGoGo`.
- [x] Define an explicit config/construction API such as `engine.Config`,
      `engine.Available()`, and `engine.NewContext(...)`.
- [x] Add engine-level image loading/conversion helpers so callers can create
      engine images from Go images or files without going through the root API.
- [x] Add engine-level blank-image/create/attach APIs for caller-managed buffers
      instead of supporting only file/image conversion inputs.
- [ ] Expose a narrow backend-neutral surface covering the shared high-level
      operations only: clear, fill/stroke color, line width/cap/join, path
      construction, fill rules, clip box, transforms, basic image drawing,
      compositing, and image export.
- [x] Treat unsupported operations as explicit capability errors rather than
      silent no-op or backend switching.
- [ ] Wire capability checks into any future partial or comparison-only backend
      implementations so unsupported features fail with typed capability errors
      instead of ad hoc messages.
- [x] Define and test the error contract for engine/resource mismatches, such as
      passing an image created by one engine implementation into another.
- [x] Add package-level docs and a small end-to-end example for first-time
      `engine` usage.
- [x] Add engine-level image export/interop helpers beyond PNG and `ToGoImage`,
      such as `ToStandardImage` and JPEG export, or explicitly document why the
      facade stops short of those conversions.

### 5.4 Scope for v1

- [ ] Support these common high-level operations through both engines:
      rectangle/circle helpers, `MoveTo`/`LineTo`/`QuadTo`/`CubicTo`/`Close`,
      fill/stroke rendering, affine transforms, clip box control, solid fills,
      dashed strokes, basic image copy/scale/quad mapping, and compositing mode
      selection.
- [x] Finish the port-backed facade coverage for operations already present on
      the root `agg.Context`, especially clip box, image-region helpers,
      gradients, and text-related APIs.
- [x] Decide and implement the first getter/readback subset exposed by the
      facade v1 contract: fill-rule state, blend mode state, gradient type
      state, clip-box readback, text-hint state, and text metrics/bounds.
- [ ] Decide whether backend-neutral transform-matrix readback belongs in the
      facade v1 contract or remains intentionally out of scope.
- [ ] Decide whether additional style/state getters from the root API, such as
      current colors, line width, line cap, and line join, belong in the facade
      contract or should stay write-only for v1.
- [ ] Delay full abstraction of low-level rasterizer/scanline/pixfmt internals.
- [ ] Keep demo-by-demo backend switching out of the first public cut unless the
      demo already uses only the supported shared surface.

### 5.5 In-repo C++ engine migration constraints

- [x] Keep the optional C++-backed engine behind an explicit build tag such as
      `agogo` or `cppref`, not in the default build.
- [ ] Move or reimplement any required C++ FFI glue, build configuration,
      wrappers, and test helpers inside this repository rather than depending on
      `github.com/cwbudde/agogo` at runtime.
- [x] Return concrete typed unavailable errors for the currently known C++
      migration prerequisites and build modes: missing `agogo` build tag,
      `agogo` builds without cgo, and `agogo` builds where the in-repo bridge is
      still only a stub.
- [ ] Extend the unavailable-path checks once the real C++ bridge exists so
      missing native AGG libs, pkg-config failures, or other link/runtime
      prerequisites also surface as concrete unavailable errors.
- [x] Never silently fall back from the C++ engine to the native port, and
      never silently accept a stub implementation as a valid backend.

### 5.6 AGoGo absorption gate

Before exposing the in-repo C++ engine outside comparison tooling, mine
`../AGoGo` for reusable assets and make them trustworthy here:

- [ ] Audit `../AGoGo/go` and `../AGoGo/cpp` to identify what should be ported
      into this repository: C++ wrapper code, tests, benchmarks, fixtures,
      docs, and feature-specific edge-case knowledge.
- [ ] Fix the currently observed test/build breakages from the audit:
      duplicate helper symbols such as `abs` and `compareImages`, missing test
      bridge functions such as `CAPIImageGetBuffer`, and stale enum names such
      as `LineCapRound` / `LineJoinRound`.
- [ ] Review and classify all current stub, fallback, and "not implemented"
      paths found in AGoGo-derived code: each one must become either
      fully supported, explicitly unavailable, or comparison-only.
- [ ] Add a hard guard that rejects the C++ engine when the migrated build has
      produced a stub implementation instead of a real AGG-backed library.

### 5.7 AGoGo feature audit and trust boundaries

- [ ] Audit the AGoGo surface area against what the in-repo facade intends to
      expose, focusing on image, path, transform, stroke, paint, compositing,
      text, and scanline/boolean behavior.
- [ ] Record engine support status in a new `docs/BACKENDS.md` capability
      matrix, including required native dependencies, migrated pieces, and known
      unsupported operations.
- [ ] Reconcile or absorb stale AGoGo documentation that still positions it as
      the future pure-Go destination; update this repo's docs so the final story
      is "single repo, Go-first implementation, optional in-repo C++ reference
      engine".
- [ ] Keep any AGoGo-derived but still-partial SVG/text/pattern behavior out of
      the shared facade until it is verified and documented.

### 5.8 Shared comparison and benchmark layer

- [ ] Add a backend-neutral scene corpus in `agg_go` that renders the same
      operations through both engines.
- [ ] Cover at least these comparison scenes: solid fill/stroke, dashed stroke,
      self-intersecting paths with both fill rules, affine image transforms,
      clip boxes, gradients, compositing, and one text scene once the migrated
      C++ text path is inside the supported matrix.
- [ ] Add a render workflow that emits `port`, `cpp`, and diff outputs for
      manual inspection and parity triage.
- [ ] Add a benchmark workflow that runs the same scene corpus through both
      engines and records runtime/allocation data from the same high-level
      description.

### 5.9 Verification and exit criteria

- [x] Add first-cut unit tests for `engine.Available()`, default engine
      selection, explicit unavailable C++ requests, and port-backed image
      drawing through the facade.
- [x] Add tests covering blank-image creation, caller-managed buffers, attached
      contexts, examples, and typed engine-mismatch errors.
- [ ] Add unit tests for backend selection, capability discovery, and explicit
      unavailable/stub-rejected error paths.
- [ ] Add cross-backend conformance tests that render the same scene through
      both engines and compare outputs with exact-match or documented tolerance
      envelopes.
- [ ] Require the migrated in-repo C++ engine to pass its own build/test gate
      before it is considered a supported backend.
- [ ] Document all intentional behavioral differences and capability gaps in
      `docs/BACKENDS.md` and, when rendering semantics differ, in
      `docs/AGG_DELTAS.md`.
- [ ] Remove the need for `../AGoGo` to perform normal development,
      verification, or release validation.

### 5.10 Final rename and consolidation

- [ ] Once the external AGoGo repository is redundant, rename this repository
      back to `AGoGo`.
- [ ] Update `go.mod` from `github.com/cwbudde/agg_go` to the final module path
      chosen for the renamed repository.
- [ ] Update internal docs, examples, CI, badges, links, and generated
      references that still mention `agg_go`.
- [ ] Decide and document the compatibility story for existing importers:
      whether package name stays `agg`, whether module redirects are relied on,
      and whether a temporary migration note or deprecated mirror is needed.

### 5.11 Non-goals

- [ ] Do not turn the existing root `agg` package into an interface-only
      abstraction layer.
- [ ] Do not expose the full low-level AGG pipeline through the shared facade in
      the first pass.
- [ ] Do not keep two actively developed repositories with overlapping
      responsibilities once the migration is complete.
- [ ] Do not rely on AGoGo stub mode or undocumented fallbacks to claim engine
      support.
- [ ] Do not move the center of gravity for pure-Go rendering work back out of
      this repository.

---

## Phase 6 - Exit Checklist

The plan is complete when the remaining open items below are all closed or explicitly deferred
with rationale:

- [ ] Visual regression CI is green.
- [ ] Every remaining demo mismatch has either been fixed or documented as intentionally deferred.
- [ ] The FreeType baseline issue is fixed or has a documented, tested workaround.
- [ ] The pixfmt generics decision has been made and reflected in code and docs.
- [ ] All required AGoGo functionality has been migrated or superseded here, and
      the external AGoGo repo is no longer needed for normal work.
- [ ] The repository/module rename plan is complete or explicitly deferred with
      rationale and migration notes.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement or fix the Go behavior.
3. Add or update contract tests.
4. Add or update visual regression tests if the behavior is rendering-visible.
5. Update this plan.
