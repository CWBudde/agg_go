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

The font subsystem is mostly in good shape, but one visible regression remains in the FreeType
baseline positioning path. This is a good candidate for a narrow fix plus a pixel-level regression
test rather than a larger architecture change.

### 3.1 Open bug

- [ ] Fix the FreeType raster glyph vertical baseline bug where short glyphs such as `.`, `,`, and
      `-` render at the wrong Y position when using `RasterFontCache`.
- [ ] Re-check the glyph bitmap positioning pipeline around `InitEmbeddedAdaptors`,
      `NewSerializedScanlinesAdaptorAA`, and `glyphBitmapRasterizer.SweepScanline`.
- [ ] Add a pixel-level regression test for a mixed-height string such as `0.2 H,x-y`.
- [ ] Add a C++ comparison for the same string and font size if practical.

### 3.2 Acceptance criteria

- [ ] Short glyphs land on the baseline band rather than above x-height.
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

## Phase 5 - Exit Checklist

The plan is complete when the remaining open items below are all closed or explicitly deferred
with rationale:

- [ ] Visual regression CI is green.
- [ ] Every remaining demo mismatch has either been fixed or documented as intentionally deferred.
- [ ] The FreeType baseline issue is fixed or has a documented, tested workaround.
- [ ] The pixfmt generics decision has been made and reflected in code and docs.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement or fix the Go behavior.
3. Add or update contract tests.
4. Add or update visual regression tests if the behavior is rendering-visible.
5. Update this plan.
