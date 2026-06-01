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

### 4.0 Current starting point (verified 2026-05-31)

A codebase audit established what float infrastructure already exists vs. what
must be built. The earlier draft of this phase assumed an "`rgba32`/`rgba128`-class
pixel-format stack" was available to wire in; for RGBA it is **not**. Reality:

**Already present (float-capable):**
- `internal/color/rgba32.go` — `color.RGBA32[CS]` with `float32` channels and a
  full method set (`Premultiply`, `Demultiply`, `Gradient`, `Scale`, `Add`,
  `Opacity`, …). This is the Go equivalent of C++ `agg::rgba32` (the float color
  selected by `AGG2D_USE_FLOAT_FORMAT`). `color.RGBA` (float64) also exists.
- `internal/buffer/rendering_buffer.go` — `RenderingBuffer[T]` is generic;
  `RenderingBufferF32` already exists and is used by the gray-float pixfmt.
- `internal/renderer/base.go` — `RendererBase[PF PixelFormat[C], C any]` is
  color-generic.
- `internal/span/*` — `SpanGradient[ColorT any, …]`, `SpanAllocator[C]`, and
  interpolators are color-generic.
- `internal/rasterizer/*` and `internal/scanline/*` are color-agnostic (coverage
  only) and need no changes.
- **Precedent to mirror:** a complete float **gray** stack exists —
  `internal/pixfmt/blender/gray32.go` (`Gray32Blender` over `[]float32`) and
  `internal/pixfmt/pixfmt_gray32.go` (`PixFmtAlphaBlendGray32` over
  `*buffer.RenderingBufferF32`). The float RGBA stack should be its structural twin.

**Missing (must be built):**
- A float **RGBA blender** (`[]float32`, order-aware). Only `rgba8`, `rgba16`,
  float-`rgb32`, and float-`gray32` blenders exist today.
- A float **RGBA pixfmt** (128-bit). Only `pixfmt_rgba8` (32-bit) and
  `pixfmt_rgba16` (64-bit) exist; the RGBA `PixFmtAlphaBlendRGBA` base is
  hardwired to `*buffer.RenderingBufferU8`.
- The `Agg2DFloat` twin (internal + public) and float `Image`/`Context` wiring.

**Naming decision (to avoid a real collision):** the existing 8-bit pixfmt is
*already* aliased `PixFmtRGBA32*` ("RGBA32" = 32-bit **pixel**, 8 bits/channel),
and a code comment equates C++ `blender_rgba32` with the 8-bit blender. To avoid
ambiguity the float stack is named by **total pixel width = 4 × float32 = 128
bits**, matching AGG's own `pixfmt_rgba128`:
- Blender: `RGBA128Blender[S]` interface + `BlenderRGBA128[S,O]` / `…Pre` / `…Plain`.
- Pixfmt: `PixFmtAlphaBlendRGBA128[…]` + aliases `PixFmtRGBA128`, `PixFmtRGBA128Pre`, `PixFmtRGBA128Plain`.
- The color these pair with is `color.RGBA32` (float). Document this pairing in
  `docs/AGG_DELTAS.md`.
- Public twin keeps the plan's working names: `Agg2DFloat`, `ContextFloat`, `ImageFloat`.

### 4.1 Goal

- [x] Add a dedicated float-backed `Agg2D` implementation path that mirrors the
      existing 8-bit `Agg2D` API closely enough for side-by-side parity work.
      (L4–L6: internal `Agg2DFloat` + public `Agg2DFloat`/`ContextFloat`.)
- [x] Keep the current 8-bit `Agg2D` behavior and public API stable. (No 8-bit
      files changed; the float path is purely additive.)
- [x] Make the float path explicit in naming and construction so callers opt in
      intentionally rather than changing global build configuration. (Selected by
      constructing `Agg2DFloat`/`ContextFloat`/`ImageFloat`; no build tags.)

### 4.2 Layered build order (TDD, bottom-up)

Each layer is independently testable; build and green-test before moving up.
C++ references live under `../agg-2.6/agg-src/`.

- [x] **L1 — Float RGBA blender** `internal/pixfmt/blender/rgba128.go` (+ `_test.go`).
      Structural twin of `gray32.go`, with RGBA8's order/interface shape. Provides
      `RGBA128Blender[S]` and `BlenderRGBA128` (plain→premul), `BlenderRGBA128Pre`
      (premul→premul), `BlenderRGBA128Plain` (plain→plain), plus single-pixel/hline
      helpers (`BlendRGBA128Pixel`, `CopyRGBA128Pixel`, `BlendRGBA128Hline`,
      `CopyRGBA128Hline`, `FillRGBA128Span`) and order/space aliases. Float
      arithmetic: `lerp(p,q,a)=p+(q-p)*a`, `prelerp(p,q,a)=p+q-p*a`, cover ∈ [0,1].
      13 TDD unit tests green; golangci-lint clean.
      C++ ref: `agg_pixfmt_rgba.h` `blender_rgba{,_pre,_plain}` with float color.
- [x] **L2 — Float RGBA pixfmt** `internal/pixfmt/pixfmt_rgba128.go` (+ `_test.go`).
      Twin of `pixfmt_gray32.go` over `*buffer.RenderingBufferF32`; implements the
      full `renderer.PixelFormat[color.RGBA32[S]]` surface (compile-time asserted
      in the test) — pixel/hline/vline/bar/solid-span/color-span/clear — driven by
      the L1 blender (order + premul semantics) with 8-bit `cover` normalised to
      [0,1]. Length-based vline/bar semantics match `pixfmt_rgba8`/`RendererBase`
      (not gray32's end-coord variant). Concrete aliases `PixFmtRGBA128{,Pre,Plain}`
      + sRGB variants and constructors. 7 TDD tests green; gofmt/lint clean; whole
      pixfmt package passes. Composite (`Comp`/`CompPre`) variants deferred to L5.
      C++ ref: `pixfmt_rgba` family parameterised on `rgba128`.
- [x] **L3 — Float image + buffer wiring** `internal/agg2d/buffer_float.go` and a
      float `Image`. Define the boundary contract (see 4.3) and conversions
      to/from `color.RGBA32`, standard Go `image.RGBA`/`image.NRGBA64`, and the
      8-bit AGG image.
      DONE 2026-05-31 (TDD, 8 tests): added `ImageFloat` over `RenderingBufferF32`
      storing **straight** RGBA float32 (4/pixel, [0,1]), the float twin of
      `Image`. Constructors `NewImageFloat`/`NewImageFloatEmpty`; `Width/Height/
      Stride/IsAttached/Attach`; straight `GetPixel`/`SetPixel` over `color.RGBA32`;
      in-place `Premultiply`/`Demultiply`. Boundary conversions honoring each
      format's alpha convention: `ToNRGBA64`/`NewImageFloatFromNRGBA64` (straight
      ↔ straight 16-bit), `ToRGBA`/`NewImageFloatFromRGBA` (straight ↔ Go's
      premultiplied 8-bit, premul/demul at the boundary), `ToImage8`/
      `NewImageFloatFromImage8` (straight ↔ 8-bit AGG image, ×255). Contract
      documented in the file header. Verified: 8/8 tests, full agg2d package green,
      gofmt/vet/golangci-lint clean on new files, `go build ./...` OK.
- [x] **L4 — `Agg2DFloat` internal twin** `internal/agg2d/agg2d_float.go`. Mirror
      the internal `Agg2D` struct field-for-field, swapping the pixfmt/renderer/
      color/gradient-LUT/span types to the float ones. Keep a one-to-one method
      mapping; share only behavior-identical helpers (transform, path, converters,
      rasterizer, scanline are reused as-is).
      DONE 2026-05-31 (TDD, 4 + 3 tests). Two parts:
      (1) Float gradient span support in `internal/span/span_gradient.go`:
      `GradientPrebuiltColorRGBA32` + `NewLinearGradientFromLUT32` /
      `NewRadialGradientFromLUT32` (float twins of the RGBA8 LUT helpers) — 3 tests.
      (2) `Agg2DFloat` struct mirroring `Agg2D` field-for-field with float types:
      `rbuf *RenderingBufferF32`, `pixfmt *PixFmtRGBA128Plain`/`pixfmtPre`,
      `renBase`/`renBasePre baseRendererAdapter[RGBA32[Linear]]`, gradient array
      `[256]RGBA32`, `spanAllocator`/`*GradientLUT []RGBA32`, and the four float
      gradient span generators. Per C++ `AGG2D_USE_FLOAT_FORMAT`: the public
      `Color` stays 8-bit (srgba8); `ColorType`→rgba32 drives pixfmt/blender/
      span/gradient. Constructor `newAgg2DFloat` mirrors `NewAgg2D`; render wiring
      `Attach`/`AttachImageFloat`/`initializeRendering`/`ClipBox`/`ClearAll`/
      `GetBounds`/`WorldToScreen`/`ScreenToWorld`/`FillColor`/`LineColor` +
      `colorToRGBA32` boundary helper. Composite pixfmt fields deferred to L5
      (no `PixFmtCompositeRGBA128` yet). Fields declared ahead of their L5/L6
      wiring (gradient arrays, font, dash, curve-ctrl, transform stack) carry
      documented `//nolint:unused`. Verified: 7/7 new tests, full agg2d + span +
      pixfmt packages green, gofmt/vet clean, golangci-lint clean (only the
      pre-existing `infertypeargs` style hints shared with the 8-bit `buffer.go`),
      `go build ./...` OK.
- [x] **L5 — Subsystem coverage** in the float twin: clear/fill/stroke, path
      rendering, gradients (linear/radial), image copy/blend/transform, and text
      state plumbing. Reuse color-generic span/renderer code; only the pixfmt and
      color LUTs differ.
      DONE 2026-05-31 (TDD, 5 + 2 end-to-end pixel-asserting tests). Five new files,
      all rendering into a real float buffer and verified at pixel level:
      • `rendering_float.go` — float render core: `renderFill`/`renderStroke`/
        `renderFillWithLineColor`, `addStrokeToRasterizer`, `renderSolidFillWithColor`
        (+ master-alpha), `renderSolidStroke`, `renderGradientFill`/`Stroke`,
        `renderLinearGradientFill`/`renderRadialGradientFill`, `scanlineRender`,
        `updateApproximationScales`/`updateRasterizerGamma`, `WorldToScreenScalar`,
        `LineWidth/Cap/Join`, `Set/GetMasterAlpha`, `SetAntiAliasGamma`,
        `TextAlignment`/`FlipText` (text state plumbing). The render helpers
        (`RenderScanlinesAA`, `NewRendererScanlineAASolidWithColor`) are color-generic
        and instantiate over `RGBA32`; only the LUT refresh (`copy` of `[256]RGBA32`)
        and `colorToRGBA32`+master-alpha differ from 8-bit.
      • `paths_float.go` — `ResetPath`/`MoveTo`/`LineTo`/(rel)/`HorLineTo`/`VerLineTo`/
        `ArcTo`/`QuadricCurveTo`/`CubicCurveTo`/`AddEllipse`/`ClosePolygon`/`DrawPath`/
        `DrawPathNoTransform` (bodies identical to paths.go — shared color-agnostic state).
      • `shapes_float.go` — `Line`/`Triangle`/`Rectangle`/`Ellipse`/`DrawCircle`/`FillCircle`.
      • `gradient_float.go` — float gradient builders (`buildProfileGradient32`/
        `buildThreeColorGradient32`, interpolating in RGBA32 space) + `FillLinearGradient`/
        `LineLinearGradient`/`FillRadialGradient`/`LineRadialGradient`/`FillRadialGradientMultiStop`.
      • `image_float.go` — `CopyImageFloat`/`BlendImageFloat` via the base renderer's
        CopyFrom/BlendFrom with a float source pixfmt over the `ImageFloat` buffer.
      Carried forward (documented in files): composite (Comp/CompPre) blend modes;
      full affine/perspective image transforms (`TransformImage*`); the remaining
      shape helpers (Star, RoundedRect variants, Arc, smooth curve variants); and
      actual glyph rasterization (only text *state* is plumbed). Verified: 15/15
      agg2d float tests, full agg2d package green, gofmt/vet clean, golangci-lint
      clean on all float files, `go build ./...` OK.
- [x] **L6 — Public surface** `agg2d_float.go` (root) + `context_float.go` +
      float `Image` API: `NewAgg2DFloat`, `Attach`, `AttachImage`, and the mirror
      of the 8-bit public methods that the float subsystems support.
      DONE 2026-05-31 (TDD, 4 public end-to-end tests). Exported the internal
      constructor (`newAgg2DFloat`→`NewAgg2DFloat`). Root `agg2d_float.go`:
      • public `ImageFloat` (thin wrapper over `agg2d.ImageFloat`) — `NewImageFloat`,
        `Width`/`Height`, `Get/SetPixelFloat` (float tuples, no internal-color leak in
        signatures), `Premultiply`/`Demultiply`, and boundary `ToRGBA`/`ToNRGBA64`
        returning standard Go image types.
      • public `Agg2DFloat` (wraps `*agg2d.Agg2DFloat`) with `NewAgg2DFloat`, `Attach`
        (`[]float32`), `AttachImage`, `GetImpl`, and the full mirror of the supported
        8-bit methods: ClearAll/ClipBox/GetBounds, FillColor/LineColor, LineWidth/
        Cap/Join, master-alpha + AA gamma, path (ResetPath/MoveTo/LineTo/rel/Hor/Ver/
        ArcTo/Quadric/Cubic/AddEllipse/ClosePolygon/DrawPath/NoTransform), shapes
        (Line/Triangle/Rectangle/Ellipse/Draw+FillCircle), gradients (Fill/Line ×
        Linear/Radial + MultiStop), image transfer (CopyImage/BlendImage), transforms
        (WorldToScreen/ScreenToWorld/WorldToScreenScalar), and text state (TextAlignment/
        FlipText). Public `Color` stays 8-bit; `toInternalColor` bridges to `agg2d.Color`.
      `context_float.go`: high-level `ContextFloat` (`NewContextFloat`/`ForImage`,
      Clear/SetColor/SetLineWidth, DrawLine/Draw+FillRectangle/Draw+FillCircle,
      GetImage/GetAgg2D), mirroring the 8-bit `Context`. No build tags — float path
      is selected purely by constructing `Agg2DFloat`/`ContextFloat`/`ImageFloat`.
      Verified: 4/4 public tests (solid fill, gradient+CopyImage, ToRGBA boundary,
      ContextFloat) + full root and internal agg2d suites green, gofmt/vet clean,
      golangci-lint clean on all float files, `go build ./...` OK.

Do not introduce build tags for selection; the float path is chosen purely by
constructing `Agg2DFloat`/`ContextFloat`/`ImageFloat`.

### 4.3 Required scope and boundary contract

- [x] Provide dedicated float image and context types with attach/create APIs,
      not just a hidden internal renderer. (L6 — public `ImageFloat`, `Agg2DFloat`,
      `ContextFloat` with `NewImageFloat`/`NewAgg2DFloat`/`NewContextFloat`/Attach.)
- [x] Cover at least these Agg2D subsystems in the float variant:
      clear/fill/stroke, path rendering, image copy/blend/transform, gradients,
      and text state plumbing. (L5 — full affine/perspective image transform
      carried forward; copy/blend covered.)
- [x] Define and document the boundary conversions between float Agg2D images
      and standard Go image types or 8-bit AGG images. (L3 — `buffer_float.go`.)
- [x] Add an explicit contract for premultiply/demultiply behavior in the float
      variant. Working contract (to confirm during L2/L3): internal storage and
      the `Plain`/`Pre` blender split mirror the 8-bit semantics exactly; exported
      helper APIs expose **straight (non-premultiplied)** float data at the
      boundary, with conversion to/from premultiplied happening inside the pixfmt
      blenders, identical to the 8-bit path.

### 4.4 Non-goals

- [x] Do not replace the existing 8-bit `Agg2D`. (Untouched; float path is additive.)
- [x] Do not hide both precisions behind one opaque constructor unless parity
      and debugging remain straightforward. (Separate `NewAgg2D` / `NewAgg2DFloat`.)
- [x] Do not treat this as a generic whole-library precision rewrite. This task
      is specifically about an explicit Agg2D-level float twin first. (Scope held
      to the Agg2D twin + the float pixfmt/blender it needs.)
- [x] Do not add a 16-bit Agg2D variant in the same change unless it falls out
      naturally from shared lower-level abstractions after the float path is in
      place and tested. (No 16-bit Agg2D added.)

### 4.5 Verification and exit criteria

- [x] Add side-by-side tests that render the same scene through 8-bit Agg2D and
      float Agg2D, then compare the quantized float output against expected
      tolerance envelopes.
      DONE 2026-05-31 — `internal/agg2d/parity_float_test.go`: a `parityTarget`
      interface (the method subset shared verbatim by `*Agg2D` and `*Agg2DFloat`)
      drives one scene through both pipelines; solid fill is pixel-identical
      (tol 1), linear gradient within tol 3, stroke within tol 2.
- [x] Add source-linked tests for premultiply/demultiply in the float path.
      DONE 2026-05-31 — `internal/agg2d/premultiply_float_test.go` ties
      `color.RGBA32` and `ImageFloat` premul/demul to `agg_color_rgba.h`
      `rgba32T::premultiply()`/`demultiply()` (~L1243), including the `a < 1`
      opaque-no-op guard and `a <= 0` zeroing. The **transformed-image** half of
      this item stays deferred: `TransformImage*` is not yet in the float path
      (see §4.7 and `docs/AGG_DELTAS.md` "Float Agg2D Variant").
- [x] Add at least one visual regression/demo hook that can run a demo via the
      float Agg2D path without disturbing the existing 8-bit baseline.
      DONE 2026-05-31 — `tests/visual/float_path_test.go` renders one opaque
      scene through the public float path (`ContextFloat` → `ImageFloat.ToRGBA`)
      and the same scene through 8-bit `Context` as the oracle, asserting parity
      within documented tolerances (solid ~1, gradient/AA ~4). No new reference
      PNGs; the float render is saved to `tests/visual/output/float_path.png`.
- [x] Document the intentional API and behavioral differences between the 8-bit
      and float variants in `docs/AGG_DELTAS.md` or a dedicated companion note.
      DONE 2026-05-31 — `docs/AGG_DELTAS.md` "Float Agg2D Variant" section:
      explicit (no-build-tag) selection, `RGBA128`/`color.RGBA32` naming,
      8-bit public `Color`, the premul/demul boundary contract, and the deferred
      capability gaps.

### 4.6 Post-completion hardening (2026-05-31)

Follow-up after the L1–L6 review:
- [x] **Parity bug fixed**: float `Attach` now calls `updateRasterizerGamma()`
      (matching 8-bit `buffer.go`). Without it, a master alpha set before a
      re-`Attach` leaked into later rendering as a stale coverage scale (caught by
      `TestAgg2DFloatAttachResetsRasterizerGamma`: interior alpha was 0.498 → now 1.0).
- [x] **Cross-precision parity test** added (see §4.5).
- [x] **Breadth increment** (trivial color-agnostic delegations): world transforms
      (`Rotate`/`Scale`/`UniformScale`/`Skew`/`Translate`/`ResetTransformations`),
      fill mode (`FillEvenOdd`/`GetFillEvenOdd`, `NoFill`/`NoLine`) on both the
      internal twin (`transform_float.go`) and the public surface, TDD'd incl. an
      even-odd donut-hole test.

### 4.7 Deferred to a future change (not blocking)

The L1–L6 float twin is complete, tested, and usable today (clear/fill/stroke,
paths, shapes, image copy/blend, linear/radial gradients, world transforms, fill
rules; cross-precision parity + visual hook green). The following are genuinely
new work, deferred with rationale and recorded in `docs/AGG_DELTAS.md`
("Float Agg2D Variant" → capability gaps). None block the float twin from being
usable.

- [ ] Full affine/perspective **image transform** (`TransformImage*`). Only
      rectangle-aligned `CopyImage`/`BlendImage` exist. The span generators and
      interpolators are color-generic and reusable; only float image
      pixel-format adapters need wiring. *Highest priority.*
- [ ] **Composite blend modes**: build `PixFmtCompositeRGBA128` and wire
      `BlendMode`/`imageBlendMode`. Today only src-over is effective in the float
      path (state is stored but inert).
- [ ] **Text glyph rendering**: mirror `Text()` glyph rasterization. Only text
      *state* (alignment/flip/hints/height) is plumbed.
- [ ] Remaining **~100 public-method delegations**: Viewport/Parallelogram,
      Arc/RoundedRect/Polygon/Star, curve variants, dashes, positioned/multi-stop
      gradient variants, `Get*` accessors, `GouraudTriangle`, transform stack.
      *Mostly mechanical; path/transform building is color-agnostic and reusable.*

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
- [x] Keep the optional native C++ support layer internal to the `engine`
      package rather than growing a second public or semi-public bridge API.
- [ ] Avoid re-introducing a separate bridge-plus-adapter architecture unless a
      concrete reuse case appears that justifies the extra abstraction.

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
- [x] Create an in-repo `agogo`-tagged native C++ backend support layer inside
      the `engine` package with local header/source files, cgo build
      configuration, native metadata/probe helpers, and build-mode-specific
      tests, so the native boundary now lives in this repository rather than in
      the external AGoGo module.
- [x] Migrate the first actual native primitive slice directly into the
      `engine` package: local image allocation/clear/readback, region blit,
      local path allocation/editing, affine matrix/path transforms, and minimal
      native path fill/stroke primitives with tagged tests.
- [x] Convert the current native helper primitives from internal migration
      scaffolding into actual package-private `engine` backend types
      implementing `engine.Context` and `engine.Image` for the currently
      supported subset.
- [x] Keep the native helper surface package-private while the C++ backend is
      still partial, so users only interact through the public `engine`
      interfaces and availability/capability checks.
- [ ] Move or reimplement the remaining C++ FFI glue, build configuration,
      wrappers, and test helpers needed by the actual engine adapter inside
      this repository rather than depending on `github.com/cwbudde/agogo` at
      runtime.
- [x] Add a first real AGG-backed in-repo build path for the current direct
      `engine` C++ subset, so `engine.CPP` becomes available in a local
      `agogo`+real-native build instead of only through the stub native layer.
- [x] Port the first real direct C++ backend primitives needed for parity with
      the current facade subset: image scaling/quad mapping, clip box support,
      the current compositing subset, gradients, and a first text slice.
- [ ] Extend that partial C++ parity work to the remaining shared-facade gaps,
      especially dashed strokes, broader compositing coverage, transformed
      image-region behavior, and any image paths that still bypass AGG in the
      real-native build.
- [x] Make the current package-private C++ backend gradient setters affect
      actual fill/stroke rendering for the migrated subset, with tagged tests
      covering at least one fill gradient and one stroke gradient case.
- [x] Make the current package-private C++ backend honor clip-box state during
      fill, stroke, and image operations by compositing through clip-aware
      native helper paths instead of storing clip state only.
- [x] Add scaled image-region drawing to the current package-private C++ backend
      subset, with internal tests covering scaled copy plus clip/blend
      interaction.
- [x] Add package-private C++ image quad drawing for the current migrated
      subset, with internal tests covering full-image quads, source-region
      quads, clip interaction, and typed unsupported blend-mode rejection.
- [x] Add a first real-text slice to the AGG-backed in-repo C++ engine path:
      font loading, hinting configuration, text drawing, and basic text
      measurement/bounds for the current facade contract.
- [x] Return concrete typed unavailable errors for the currently known C++
      migration prerequisites and build modes: missing `agogo` build tag,
      `agogo` builds without cgo, and `agogo` builds where the in-repo C++
      backend support is still only a stub.
- [ ] Extend the unavailable-path checks once the real in-repo C++ backend
      exists so
      missing native AGG libs, pkg-config failures, or other link/runtime
      prerequisites also surface as concrete unavailable errors.
- [ ] Collapse the temporary `aggreal` build split back into the primary
      `agogo` build once native dependency probing can replace compile-time
      system-library assumptions cleanly.
- [ ] Replace the remaining CPU helper paths in the real C++ build with direct
      AGG-backed implementations where that materially affects parity or
      performance, especially for image operations that are still not using AGG
      internally.
- [x] Never silently fall back from the C++ engine to the native port, and
      never silently accept a stub implementation as a valid backend.

### 5.6 AGoGo absorption gate

Before exposing the in-repo C++ engine outside comparison tooling, mine
`../AGoGo` for reusable assets and make them trustworthy here:

- [x] Audit `../AGoGo/go` and `../AGoGo/cpp` to identify what should be ported
      into this repository: C++ wrapper code, tests, benchmarks, fixtures,
      docs, and feature-specific edge-case knowledge.
- [ ] Eliminate donor-repo default-fallback enum behavior during migration:
      unknown paint/compositing/pixel-format values must fail explicitly rather
      than silently mapping to solid color, src-over, or RGBA32 defaults.
- [ ] Fix the currently observed test/build breakages from the audit:
      duplicate helper symbols such as `abs` and `compareImages`, missing test
      bridge functions such as `CAPIImageGetBuffer`, and stale enum names such
      as `LineCapRound` / `LineJoinRound`.
- [x] Review and classify all current stub, fallback, and "not implemented"
      paths found in AGoGo-derived code: each one must become either
      fully supported, explicitly unavailable, or comparison-only.
- [x] Add a hard guard that rejects the C++ engine when the migrated build has
      produced a stub implementation instead of a real AGG-backed library.

### 5.7 AGoGo feature audit and trust boundaries

- [x] Audit the AGoGo surface area against what the in-repo facade intends to
      expose, focusing on image, path, transform, stroke, paint, compositing,
      text, and scanline/boolean behavior.
- [x] Audit which of the old AGoGo Go wrapper types and tests are still useful
      after the direct `engine`-local native design, and drop anything that only
      existed to support the old standalone bridge API shape.
- [x] Record engine support status in a new `docs/BACKENDS.md` capability
      matrix, including required native dependencies, migrated pieces, and known
      unsupported operations.
- [x] Reconcile or absorb stale AGoGo documentation that still positions it as
      the future pure-Go destination; update this repo's docs so the final story
      is "single repo, Go-first implementation, optional in-repo C++ reference
      engine".
- [x] Keep any AGoGo-derived but still-partial SVG/text/pattern behavior out of
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
- [x] Add unit tests for backend selection, capability discovery, and explicit
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
