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

RMSE values below are current as of 2026-06-13, measured by regenerating the Go
references (`UPDATE_VISUAL=1`) and comparing against the C++ references with
`cmd/visual-diff` (RMSE over all RGB channels). `[x]` marks demos that are
resolved: either pixel-exact (RMSE 0.0) or verified-faithful at the
floating-point noise floor — a small set of isolated sub-visual pixels whose
pipeline is a confirmed bit-faithful port of C++ (matrix, IRound, DDA,
filter/blender math all bit-identical) and whose residual is irreducible
libm-vs-Go float noise: AA-coverage LSB rounding (max channel diff ≤14, e.g.
`lion_outline`), or a single-subpixel coordinate flip at a grid-aligned sharp
edge that, for a bilinear image sample, can swing one pixel by a larger amount
(e.g. `image_alpha`, max diff 89 on 31 px). The defining test is "every integer
operation matches C++; only the transcendental/float inputs differ", not a fixed
RMSE/px cap.

- [x] `aa_demo` — pixel-exact (RMSE 0.0, 0/240000 px).
- [x] `alpha_mask` — pixel-exact (RMSE 0.0) via the lion srgba8-storage color roundtrip fix.
- [x] `alpha_mask2` — pixel-exact (RMSE 0.0): linear pipeline for all overlay passes, lion color roundtrip, gradient uround, and the line_interpolator_aa stale dist_start/dist_end fix.
- [x] `blend_color` — pixel-exact (RMSE 0.0, 0/145200 px).
- [x] `bspline` — pixel-exact (RMSE 0.0, 0/360000 px).
- [x] `circles` — pixel-exact (RMSE 0.0, 0/160000 px).
- [x] `component_rendering` — pixel-exact (RMSE 0.0, 0/102400 px).
- [x] `conv_contour` — pixel-exact (RMSE 0.0): rewritten from Agg2D to the linear pipeline (linear pixfmt, render_ctrl equivalent, FlipY + EncodeLinearRGBToSRGB).
- [x] `flash_rasterizer2` — pixel-exact (RMSE 0.0, 0/340600 px).
- [x] `gamma_correction` — pixel-exact (RMSE 0.0) after fixing C-sprintf label semantics in slider_ctrl.
- [x] `gamma_tuner` — pixel-exact (RMSE 0.0, 0/250000 px).
- [x] `gouraud_mesh` — pixel-exact (RMSE 0.0, 0/160000 px).
- [x] `gradient_focal` — pixel-exact (RMSE 0.0, 0/240000 px); the former timing-text residual is gone after deterministic reference regeneration. gradient_lut built/interpolated in sRGB space with the rgba8_gamma_dir roundtrip on stops, decoded to linear per entry; ellipse+conv_stroke boundary circle; linear pipeline + EncodeLinearRGBToSRGB.
- [x] `gradients_contour` — pixel-exact (RMSE 0.0): DT grayscale truncation (not +0.5), rbox defaults (text thickness 1.5, right edge 300), exact span_interpolator_trans; C++ reference recaptured after fixing the "Assymetric Conic" typo in the original demo.
- [x] `image_filters` — pixel-exact (RMSE 0.0): linear pipeline (sRGB-decoded PPM source, linear filtering, sRGB encode on save) + raw conv_stroke for the gsv status text.
- [x] `image_perspective` — pixel-exact (RMSE 0.0, 0/360000 px); former timing-text residual gone. Faithful-port rewrite: quad tool with handle circles, three modes (affine parl + NN, bilinear + 2x2, perspective + 2x2), linear pipeline + EncodeLinearRGBToSRGB.
- [x] `image_resample` — pixel-exact (RMSE 0.0, 0/360000 px); former timing-text residual gone. Direct faithful port: quad tool rendered like C++ interactive_polygon, all six transform modes via the real span generators, linear pipeline + EncodeLinearRGBToSRGB.
- [x] `image_transforms` — pixel-exact (RMSE 0.0, 0/96000 px).
- [x] `image1` — pixel-exact (RMSE 0.0, 0/122400 px).
- [x] `lion` — pixel-exact (RMSE 0.0): lion color roundtrip + C-truncation of the alpha slider byte.
- [x] `pattern_perspective` — pixel-exact (RMSE 0.0): quad tool + rbox rendered BEFORE the pattern, wrap-reflect accessor, normalized Hanning 2x2 filter, source rect ±150, linear_subdiv interpolator for perspective, sRGB-decoded agg.ppm, linear pipeline + EncodeLinearRGBToSRGB.
- [x] `pattern_resample` — pixel-exact (RMSE 0.0, 0/360000 px); former timing-text residual gone. Six resample modes + wrap-reflect pattern source, plus the demo's gamma_lut(2.0) (apply_gamma_dir on the pattern, apply_gamma_inv on the window before timing text and controls).
- [x] `perspective` — pixel-exact (RMSE 0.0) via the lion color roundtrip fix.
- [x] `raster_text` — pixel-exact (RMSE 0.0, 0/307200 px).
- [x] `rasterizer_compound` — pixel-exact (RMSE 0.0) after porting the linear-pipeline + sRGB-encode-on-save semantics of the C++ demo.
- [x] `rasterizers` — pixel-exact (RMSE 0.0, 0/165000 px).
- [x] `rounded_rect` — pixel-exact (RMSE 0.0, 0/240000 px).
- [x] `scanline_boolean` — pixel-exact (RMSE 0.0, 0/480000 px).
- [x] `lion_lens` — verified-faithful, float noise floor (RMSE 0.0015, 1/262144 px at ±1 LSB). conv_segmentator distortion pipeline is faithful; the lone pixel is sub-LSB sampling rounding.
- [x] `flash_rasterizer` — verified-faithful, float noise floor (RMSE 0.0031, 2/340600 px at ±1–2 LSB on one glyph edge).
- [x] `lion_outline` — verified-faithful, float noise floor (RMSE 0.0512, 18/262144 px in one isolated stroke segment, max channel diff 14). Confirmed bit-identical to C++: line_profile_aa (gamma_none), the Line0–3 / Pie dispatch in rasterizer_outline_aa, AddVertex close handling, and the IRound float→subpixel coordinate conversion. The single differing segment is a sub-ULP transcendental difference (rotation by π) flipping one vertex's subpixel coordinate — irreducible libm-vs-Go float noise, not a bug.
- [x] `simple_blur` — verified-faithful, float noise floor (RMSE 0.0579, 18/204800 px). Same root cause as `lion_outline`: identical color delta (255,251,244)→(255,246,230) on the same right-half lion outline-AA segment; the rasterizer_outline_aa + line_profile_aa pipeline is faithful.
- [x] `image_fltr_graph` — pixel-exact (RMSE 0.0, 0/234000 px). Fixed: the grid/axis lines were stroked through Agg2D (default `CapRound`), depositing AA coverage one row past each butt endpoint (rows y=9/290). C++ draws them with a raw `conv_stroke` (default butt cap); set `LineCap(CapButt)` in the demo's `strokeLine` to match.
- [x] `image_alpha` — verified-faithful, float noise floor (RMSE 0.4502, 31/96000 px, all gen-darker, on one diagonal blade in the spheres image). Confirmed bit-identical to C++: the matrix build (translate/rotate/translate, resizing=identity at initial size), `span_image_filter_rgb_bilinear` (fg=0 truncation, identical weights), the filter offset (dx_int=128, dx_dbl=0.5), `span_interpolator_linear` + `dda2_line_interpolator` (Init/Inc), and `IRound`. The residual is a single-subpixel sample flip (`x_lr` off by 1) at a sharp source-image edge that is geometrically near-aligned to the sample grid; a ~1-ULP `math.Sin/Cos(10°)` difference vs glibc tips ~31 consecutive samples the same way. Bilinear swings the flipped pixel by up to 89 (200→10 neighbor), unlike the AA-coverage demos — same irreducible float-noise class, larger per-pixel magnitude.

Ordered easiest → hardest to close (near-exact AA residuals first, localized
single-cause bugs next, then broad rounding/format fixes that touch shared paths,
and finally the genuinely algorithmic/architectural gaps).

- [ ] `distortions` — RMSE 0.0377 (53 px). A few isolated extreme pixels from distortion-resampling rounding in the lensed image/sphere; essentially done.
- [ ] `polymorphic_renderer` — RMSE 0.1239 (91 px). Float-vs-8bit AA edge-coverage rounding on the single triangle's anti-aliased edges only.
- [ ] `gradients` — RMSE 0.0335 (118 px). AA on the gradient-control spline curve lines and a couple of sphere-edge pixels; the gradient fill itself matches.
- [ ] `idea` — RMSE 0.7113 (171 px). A few extreme AA pixels on the tiny high-chroma lightbulb rays plus the step/degree label text.
- [ ] `rasterizers2` — RMSE 0.8059 (258 px). 258 stray white pixels: off-by-one Bresenham / arbitrary-image-pattern marker placement in the lower spirals, not a colorspace issue.
- [ ] `multi_clip` — RMSE 1.2046 (40 px). Only 40 px but a few are extreme: sub-pixel AA edge rounding on the dense thin random strokes/circles inside the clip cells.
- [ ] `blur` — RMSE 0.4550 (347 px). Stack-blur intermediate-precision rounding across the whole blurred glyph plus a few saturated selection-handle pixels; slider/shadow geometry already fixed.
- [ ] `line_patterns` — RMSE 0.1216 (936 px). Image-pattern glyph sampling/positioning along each curved path plus a couple of saturated control-pin pixels.
- [ ] `gamma_ctrl` — RMSE 0.0644 (1378 px). Sub-pixel AA edge fringing on the green GSV "Text 2345" glyph outlines and the thin radial-spline lines; controls exact.
- [ ] `trans_polar` — RMSE 0.0506 (1628 px). Transform-resampling AA on the curved polar ring plus the control text and slider-knob X positions.
- [ ] `conv_stroke` — RMSE 0.0803 (1709 px). Faint float-vs-8bit AA edge fringing along the dashed-stroke borders and miter-join markers; near-exact.
- [ ] `mol_view` — RMSE 0.2857 (2138 px). Sub-pixel AA fringing on the green GSV title-text glyph edges and the thin atom-bond strokes; geometry/colors already corrected.
- [ ] `line_patterns_clip` — RMSE 0.1131 (3300 px). Patterned-stroke dash phase / clip-boundary sampling on the X-crossing lines and clipped line ends; edge AA plus stray control-pin pixels.
- [ ] `compositing2` — RMSE 0.0967 (5004 px). Comp-op blend rounding (8-bit vs float) on the edges of the four overlapping translucent circles; controls exact.
- [ ] `aa_test` — RMSE 0.1728 (10685 px). Float-vs-8bit AA fringing on the many thin anti-aliased lines/dashes in the radial sub-pixel line fans; no logic error.
- [ ] `alpha_gradient` — RMSE 0.4158 (26799 px). Accumulated 8-bit blend rounding (agg.RGBA truncates `uint8(v*255)` instead of round-to-nearest `*255+0.5`) across the whole alpha-blended gradient circle and translucent ellipses; the round-to-nearest fix is one-line but touches a shared blend path.
- [ ] `line_thickness` — RMSE 0.3650 (25756 px). Uniform BGR96-float-vs-8bit edge-AA fringe along every diagonal line and radial spoke; essentially done pending a float renderer.
- [ ] `graph_test` — RMSE 0.7717 (37004 px). Sub-pixel AA on the grid of node-circle outlines plus glyph edges in the bottom timing/status text; residual after per-control text-height fixes.
- [ ] `pattern_fill` — RMSE 0.2836 (64555 px). Background tint off by integer-rounding the premultiplied RGBA8(102,0,26,26) instead of float premultiply-then-quantize (rgba_pre), spread across the pattern-filled star interior; controls/margins clean.
- [ ] `alpha_mask3` — RMSE 0.3438 (69120 px). Renders into 4-channel RGBA32 instead of the C++ opaque 3-channel BGR24 (pixfmt_rgb) buffer, so the layered low-alpha (25/127) over-blend rounds one LSB darker across the translucent shapes; controls/background identical.
- [ ] `conv_dash_marker` — RMSE 0.9998 (4672 px). Dash-phase / sub-pixel dash-segment positioning offset along the dashed line (every dash lands slightly shifted) plus the green smooth-outline edges.
- [ ] `bezier_div` — RMSE 1.1956 (2861 px). Stroke vertex generation at the Miter-Revert + Inner-Round join near the curve cusp differs slightly from C++ vcgen_stroke; diff concentrates at the inner-join triangle fan and dashed inner-stroke outline.
- [ ] `compositing` — RMSE 0.5101 (98059 px). ±1-LSB gradient/composite interpolation rounding in the 8-bit-linear scene path (the known RGBA128 float comp-op residual) spread across the gradient-filled shapes; controls/text exact.
- [ ] `scanline_boolean2` — RMSE 1.2047 (69301 px). Sub-pixel cover/span-boundary discrepancy in the scanline boolean AND-combine path (num_spans 1033 vs C++ 1031) on the intersection-shape AA edges; GSV text stroke already corrected.
- [ ] `image_filters2` — RMSE 1.2622 (53494 px). Largest real gap: the scaled right-side image is rendered via Agg2D's dedicated bilinear resampler instead of the C++ LUT-based span_image_filter_rgba general filter, so every fractional sample blends source texels differently across the whole image; control panel clean.

### 1.3 Exit criteria

- [ ] Visual regression suite passes in CI.
- [ ] No AGG2D parity row remains untriaged or placeholder-level.
- [ ] Visual references and approval workflow are centralized under `tests/visual/`.

---

## Phase 2 - Demo-Specific Fixes — DONE

All demo-specific porting issues (asset selection, input mapping, coordinate-frame
handling, canvas orientation, state init) are resolved, each with a verification path.

- **`trans_curve` / `trans_curve2`**: the embedded GSV vector font is the portable,
  deterministic stand-in for C++'s Win32 "Times New Roman" (TrueType intent covered by
  the `_ft` variants). `trans_curve2` was rewritten from the lion to the faithful
  text-along-a-double-path; its render core (`internal/demo/transcurve.DrawDouble`) is
  shared verbatim with the WASM demo, so standalone/web output is identical.
- **`image_resample` / `image_perspective`**: draggable quad handles + mouse wiring
  restored in the faithful-port rewrites.
- **`gamma_correction`, `gouraud_mesh`**: pixel-exact (RMSE 0.0) — earlier layout/quadrant/
  text reports were stale.
- **`compositing2`, `gradients`, `aatest`, `flash_rasterizer{,2}`**: layout/region/background/
  shape-index issues were stale; residuals are float-vs-8bit AA only (see §1.2).
- **Follow-ups**: standalone-vs-web parity notes, render smoke tests
  (`examples/core/intermediate/trans_curve{,2}/main_test.go`), and C++ source references
  added across demo headers, the `transcurve` package doc, and `docs/AGG_DELTAS.md`.

---

## Phase 3 - Font Fidelity — DONE

The FreeType raster-glyph vertical-baseline bug (short glyphs `.`/`,`/`-` drifting above
x-height under `RasterFontCache`) is fixed and locked in.

- Root fix: y-up Y sub-pixel phase quantization + `dstY = baseY - top + 1` bitmap placement
  (`internal/agg2d/text.go`); net baseline matches AGG.
- Regression coverage: `internal/agg2d/text_baseline_regression_test.go` renders `0.2 H,x-y`
  and asserts font-relative inked geometry (short marks land in the baseline band, comma
  descends below it); proven to fail under a simulated inverted-baseline mutation. Sub-pixel
  phase pinned by `TestRasterTextYPhaseMatchesYUpQuantization`. Both run under `-tags freetype`.
- A pixel-exact C++ comparison was evaluated and rejected as non-deterministic (FreeType
  version/hinting dependent); the font-relative invariants are the durable equivalent.
- Deviation documented in `docs/AGG_DELTAS.md` ("FreeType raster-text vertical baseline &
  Y sub-pixel phase").

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
_already_ aliased `PixFmtRGBA32*` ("RGBA32" = 32-bit **pixel**, 8 bits/channel),
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
      (not gray32's end-coord variant). Concrete aliases `PixFmtRGBA128{,Pre,Plain}` + sRGB variants and constructors. 7 TDD tests green; gofmt/lint clean; whole
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
      actual glyph rasterization (only text _state_ is plumbed). Verified: 15/15
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
      this item is now covered too: `TransformImage*` is in the float path as of
      2026-06-01 (see §4.7 and `docs/AGG_DELTAS.md` "Float image transforms"),
      with parity tests over affine/parallelogram/perspective transforms.
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

- [x] Full affine/perspective **image transform** (`TransformImage*`).
      DONE 2026-06-01 — float twins of the NN/bilinear/2×2/general/affine-resample
      RGBA filters in `internal/span/span_image_filter_rgba32.go` (reusing the
      color-agnostic filter/resample bases), a clone-clamped float image source
      `internal/agg2d/adapters_float.go`, and the full
      `renderImage`/`newImageFilterGenerator`/`renderImagePerspective` mirror plus
      `TransformImage*`/`…Parallelogram*`/`…Path*`/`…Quad*` surface in
      `internal/agg2d/image_transform_float.go` (public wrappers in
      `agg2d_float.go`). One documented deviation: the float bilinear omits AGG's
      integer rounding bias (which would shift float channels by +0.5); see
      `docs/AGG_DELTAS.md` "Float image transforms". Parity vs the 8-bit path
      verified in `internal/agg2d/image_transform_float_test.go` (affine tol 3,
      parallelogram/quad tol 4) + a visual hook
      `tests/visual/float_image_transform_test.go`.
- [x] **Composite blend modes**. DONE 2026-06-01 — float composite blenders
      `CompositeBlenderRGBA128`/`...Pre` in
      `internal/pixfmt/blender/rgba128_composite.go` (reusing the 8-bit
      `CompositeBlender.blendOperation` so the per-operator algebra is shared
      verbatim), the float composite pixfmt `PixFmtCompositeRGBA128` in
      `internal/pixfmt/pixfmt_composite_rgba128.go`, and the `renBaseComp`/
      `renBaseCompPre` wiring in `Agg2DFloat` with `SetBlendMode`/
      `updateBlendMode` (`internal/agg2d/blend_modes_float.go`) plus
      `currentRenderer`/`currentImageRenderer` switching on `blendMode`. Public
      `SetBlendMode`/`GetBlendMode`/image-blend accessors in `agg2d_float.go`.
      Parity vs the 8-bit path for ten operators (Multiply/Screen/Darken/
      Lighten/Difference/Exclusion/Overlay/HardLight/Plus/SrcOver) verified in
      `internal/agg2d/composite_float_test.go` (tol 2) plus a public-API test in
      `agg2d_float_test.go`; see `docs/AGG_DELTAS.md` "Composite blend modes".
- [x] **Text glyph rendering**. DONE 2026-06-01 — float twin of the 8-bit text
      pipeline in `internal/agg2d/text_float.go` (`Font`/`FontGSV`/`Text`/
      `TextWidth`/`MeasureText`/`GetTextBounds`/metrics + float `renderScanlines`/
      `renderGlyphScanlines`/`renderShapedRasterMask`/`textGSV`). The font engine,
      glyph cache, GSV font, and layout/metrics/bounds math are color-agnostic and
      reused verbatim; the raster-bitmap blend helper `blendRasterGlyphBitmap` was
      made generic over the color type (`text.go`). Solid glyph fills flow through
      `color.RGBA32[color.Linear]` and the float base renderer, honoring the active
      blend mode. Public wrappers (`Font`/`FontDefault`/`FontGSV`/`SetResolution`/
      `TextHints`/`TextForceAutohint`/`GetTextHints`/`GetAscender`/`GetDescender`/
      `MeasureText`/`GetTextHeight`/`Text`/`TextDefault`/`TextWidth`/`GetTextBounds`)
      in root `agg2d_float.go`. Parity vs the 8-bit path (GSV stroke + FreeType
      outline/raster caches, max channel diff ≤ 2) in
      `internal/agg2d/text_float_test.go` plus a public-API test in
      `agg2d_float_test.go`; see `docs/AGG_DELTAS.md` "Text glyph rendering".
- [ ] **Remaining ~90 public-method delegations.** The float twin still lacks the
      bulk of the 8-bit surface. An audit (root `Agg2D` vs `Agg2DFloat`) shows ~90
      missing public methods; most also need their underlying internal
      `Agg2DFloat` builder. Path-, transform-, and state-building is color-agnostic
      and largely reusable from the 8-bit `Agg2D`, so the work is mostly mechanical —
      the one exception (Gouraud) needs a genuinely new float span generator. Each
      group below is one reviewable slice (internal builder + root wrapper + a small
      parity test vs the 8-bit oracle):
  - [x] **Shapes** — DONE 2026-06-01. `Arc`, `ArcRel`, `RoundedRect`,
        `RoundedRectXY`, `RoundedRectVariableRadii`, `Polygon`, `Polyline`, `Star`,
        `Curve`, `Curve4`, `Parallelogram`, `ParallelogramFromRect`. Pure path
        construction → existing float `DrawPath`; the shape builders
        (`internal/shapes/rounded_rect.go`, arc/curve converters) are color-agnostic
        and reused as-is. Implemented in `internal/agg2d/shapes_float.go`
        (RoundedRect\*/Arc/Star/Curve/Curve4/Polygon/Polyline),
        `internal/agg2d/paths_float.go` (`ArcRel`), and
        `internal/agg2d/transform_float.go` (`Parallelogram`/`ParallelogramFromRect`),
        with root wrappers in `agg2d_float.go`. Parity vs the 8-bit oracle covered by
        `internal/agg2d/shapes_float_test.go` (12 shape tests, tol ≤ 2) and the
        public-surface `TestPublicAgg2DFloatShapes` in `agg2d_float_test.go`.
        (Convenience `Curve`/`Curve4` are the shapes-group methods and live here, not
        in the curve-command group below.)
  - [x] **Curve & relative path commands** — DONE 2026-06-01. `CubicCurveToSmooth`,
        `CubicCurveRel`, `CubicCurveRelSmooth`, `QuadricCurveToSmooth`,
        `QuadricCurveRel`, `QuadricCurveRelSmooth`, `HorLineRel`, `VerLineRel`.
        Implemented in `internal/agg2d/paths_float.go` with root wrappers in
        `agg2d_float.go`; the smooth-curve reflection uses the now-wired
        `lastCtrlX/Y`/`hasLastCtrl` fields (stale `//nolint:unused` markers removed).
        Parity vs the 8-bit oracle in `internal/agg2d/curves_float_test.go` (7 tests,
        tol ≤ 2) covering rel/smooth quadric & cubic curves and Hor/VerLineRel.
        (Convenience `Curve`/`Curve4` are done — see the Shapes group above.)
  - [x] **Dashed strokes** — DONE 2026-06-13. `AddDash`, `RemoveAllDashes`,
        `DashStart`, `GetDashStart`, `NoDashes`. The `conv_dash` converter and all
        dash math are color-agnostic and reused verbatim; the internal float twin
        `internal/agg2d/dash_float.go` mirrors the 8-bit `stroke.go` methods plus
        `initializeDashing` (rebuilds the pipeline Path→Curve→Dash→Stroke,
        preserving stroke state read off the existing `convStroke`). The existing
        `rendering_float.go` solid-fallback branch (`convDash.NumDashes() == 0`)
        now fires for real once dashes are installed. Root wrappers in
        `agg2d_float.go`; the stale `//nolint:unused` marker on the `convDash`
        field was removed. Parity vs the 8-bit oracle in
        `internal/agg2d/dash_float_test.go` (6 scene tests — single/multi-segment
        patterns, phase offset, polyline, RemoveAllDashes/NoDashes solid fallback,
        tol 2 — plus a `GetDashStart` round-trip) and a public-surface test
        `TestPublicAgg2DFloatDashedStrokes` in `agg2d_float_test.go`.
  - [x] **Gradient variants** — DONE 2026-06-13. `FillGradientD1/D2`,
        `LineGradientD1/D2`, `FillGradientFlag`, `LineGradientFlag`,
        `FillRadialGradientPos`, `FillRadialGradientStops`, `LineRadialGradientPos`,
        `LineRadialGradientMultiStop`. The gradient LUT/span pipeline and the
        world-radial setup already existed; added the float N-stop builder
        `buildNStopGradient32` (float twin of `buildNStopGradient`, interpolating
        stops in RGBA32 space) plus the setters/accessors in
        `internal/agg2d/gradient_float.go`. Root wrappers in `agg2d_float.go`
        (the flag accessors return `int` to match the 8-bit public surface;
        `FillRadialGradientStops` converts root `[]GradientStop` → internal
        `[]ColorStop`). Parity vs the 8-bit oracle in
        `internal/agg2d/gradient_variants_float_test.go` (5 scene tests —
        N-stop radial, fill/line multi-stop, fill/line reposition, tol 3 — plus a
        D1/D2/flag accessor-parity test) and a public-surface test
        `TestPublicAgg2DFloatGradientVariants` in `agg2d_float_test.go`.
  - [x] **Viewport & coordinate mapping** — DONE 2026-06-13. `Viewport`,
        `GetViewportTransform`, `GetScaling`, `WorldToScreenDistance`,
        `ScreenToWorldDistance`, `AlignPoint`, `InBox`, `AffineImageResamplePolicy`/
        `GetAffineImageResamplePolicy` on the internal float twin in
        `internal/agg2d/viewport_float.go` (bodies mirror transform.go/utilities.go/
        rendering.go; the package-level `viewportTransform` helper and the
        `baseRendererAdapter.rendererBase().InBox` path are reused verbatim — color-
        agnostic). Root wrappers in `agg2d_float.go` add `ScreenToWorldScalar`,
        `Viewport`, `ViewportDefault` (XMidYMid), the distance/align/inbox methods,
        and the resample-policy setter/getter (root `ViewportOption`/
        `AffineImageResamplePolicy` aliases). `WorldToScreen`/`ScreenToWorld`/
        `WorldToScreenScalar`/`ScreenToWorldScalar` already existed on the float
        twin. Parity vs the 8-bit oracle in
        `internal/agg2d/viewport_float_test.go` (a viewport render-parity test, tol
        2, plus a scalar/bool accessor-parity test for distance/align/inbox/resample
        across an identical viewport setup) and a public-surface test
        `TestPublicAgg2DFloatViewportCoordinateMapping` in `agg2d_float_test.go`.
  - [x] **Transform stack & affine matrix** — DONE 2026-06-13. `GetTransformations`,
        `SetTransformations`, `AffineFromMatrix`, `PushTransform`, `PopTransform`,
        `PushTransformations`/`PopTransformations` (aliases), `GetTransformStackDepth`
        on the internal float twin in `internal/agg2d/transform_stack_float.go`
        (bodies mirror transform.go; reuse the shared `TransformStack` type and the
        existing `Affine(*transform.TransAffine)` from transform_float.go — color-
        agnostic). The stale `//nolint:unused` marker on the `transformStack` field
        was removed. Root wrappers in `agg2d_float.go`: `GetTransformations`,
        `SetTransformations` (via `to/fromInternalTransformations`), `Affine`
        (public root form takes `*Transformations` → `impl.AffineFromMatrix`, matching
        the 8-bit surface), `PushTransform`, `PopTransform`. Parity vs the 8-bit
        oracle in `internal/agg2d/transform_stack_float_test.go` (a balanced
        push/pop/affine-sequence matrix+depth parity test, a Get/Set round-trip, and
        an empty-stack PopTransform→false test) and a public-surface test
        `TestPublicAgg2DFloatTransformStack` in `agg2d_float_test.go`.
  - [x] **State accessors & RGBA/alias setters** — DONE 2026-06-13. Internal float
        twin in `internal/agg2d/state_accessors_float.go`: `Get*` readbacks
        (`GetFillColor`, `GetLineColor`, `GetLineCap`, `GetLineJoin`, `GetMiterLimit`,
        `GetClipBox`, `GetImageFilter`, `GetImageResample`, `GetAntiAliasGamma`),
        `*RGBA` setters (`FillColorRGBA`, `LineColorRGBA`, `ClearAllRGBA`,
        `ClearClipBoxRGBA`), `MiterLimit`/`ImageFilter`/`ImageResample`/
        `SetImageFilterRadius` setters, fill-rule queries (`IsEvenOddFillRule`,
        `IsNonZeroFillRule`, `FillRuleDescription`), `ResetStyle`, `ClearClipBox`.
        Bodies mirror colors.go/fill_rules.go/rendering.go/stroke.go/utilities.go;
        the image-filter LUT switch is byte-identical (both twins share `*aggimage`),
        `ClearClipBoxRGBA` builds the float color via `colorToRGBA32` + renderer
        `CopyBar`. Root wrappers in `agg2d_float.go`: all the above plus
        `GetClipBoxRect` (composed from `GetClipBox`) and the C++-style accessor
        aliases that the 8-bit public surface exposes — `AntiAliasGamma`,
        `MasterAlpha`, `BlendMode`, `ImageBlendMode`, `ImageBlendColor`,
        `ImageBlendColorRGBA` (the `Set*`/`Get*` forms already lived on the float
        twin; these add the C++ alias spellings). Parity vs the 8-bit oracle in
        `internal/agg2d/state_accessors_float_test.go` (full style-snapshot readback,
        float value round-trips, `ResetStyle` defaults, `ClearClipBoxRGBA` render
        parity) and public-surface tests `TestPublicAgg2DFloatStateAccessors`,
        `TestPublicAgg2DFloatClearClipBoxRGBA`, `TestPublicAgg2DFloatAccessorAliases`
        in `agg2d_float_test.go`. (Deferred from this item: `GetClipBoxRect` exists at
        the root only, as on 8-bit; `ImageFilter`/`ImageResample` listed twice in the
        original spec — implemented once as setters.)
  - [x] **Image convenience + export** — DONE 2026-06-13. `CopyImageSimple`,
        `BlendImageSimple`, `BlendImageDefaultAlpha`, `BlendImageSimpleDefaultAlpha`,
        `SaveImagePPM`. Internal float twin in
        `internal/agg2d/image_convenience_float.go`: the float-dst `*Simple` forms
        mirror the 8-bit `image.go` semantics (WorldToScreen + integer truncation of
        the destination) and delegate to the existing whole-image transfer
        primitives `CopyImageFloat`/`BlendImageFloat`; `BlendImageDefaultAlpha`/
        `BlendImageSimpleDefaultAlpha` add the upstream default-alpha-255 spellings;
        `SaveImagePPM` is the float twin of the 8-bit `agg.go` exporter, writing the
        attached `rbuf`'s straight RGB channels through `roundToU8` (alpha dropped, as
        PPM has none). Root wrappers in `agg2d_float.go`. Parity/behavior vs the 8-bit
        oracle in `internal/agg2d/image_convenience_float_test.go` (copy/blend land at
        the rounded dst, default-alpha blends, nil-source errors, and a PPM
        header+body round-trip incl. the no-buffer error path) and public-surface
        tests `TestPublicAgg2DFloatImageConvenience`, `TestPublicAgg2DFloatSaveImagePPM`
        in `agg2d_float_test.go`. (Deviation from the 8-bit region-blend surface: the
        float twin has no region-cropped `BlendImage`/`CopyImage` — its transfer
        primitives are whole-image — so `BlendImageDefaultAlpha` is the whole-image
        integer-dst form rather than the 8-bit region form.)
  - [x] **DrawPath defaults & escape hatches** — DONE 2026-06-14.
        `DrawPathDefault`, `DrawPathNoTransformDefault`, `GetInternalRasterizer`,
        `RenderRasterizerWithColor`, `ScanlineRender`, `RenderScanlinesAAWithSpanGen`.
        The rasterizer/scanline/span-allocator are color-agnostic and shared
        verbatim with the 8-bit twin; only the renderer/span-generator color type
        differs (`RGBA32`). Internal float twin: `RenderRasterizerWithColor` already
        existed in `rendering_float.go`; added `GetInternalRasterizer`/
        `ScanlineRender`/`RenderScanlinesAAWithSpanGen` in
        `internal/agg2d/drawpath_escape_float.go` (bodies mirror agg2d.go/
        rendering.go, reusing `a.scanline`/`a.spanAllocator`/`currentRenderer` and
        `renscan.RenderScanlinesAA`). `DrawPath`/`DrawPathNoTransform` already
        existed. Root wrappers in `agg2d_float.go`: the two `*Default` convenience
        forms delegate to `impl.DrawPath(FillAndStroke)` /
        `impl.DrawPathNoTransform(FillAndStroke)` (matching the 8-bit root), plus the
        four rasterizer escape hatches (root now imports `internal/rasterizer` +
        `renscan`). Behavior vs the 8-bit semantics in
        `internal/agg2d/drawpath_escape_float_test.go` (raw-rasterizer triangle
        painted via RenderRasterizerWithColor / a caller solid renderer via
        ScanlineRender / a custom constant-color span generator via
        RenderScanlinesAAWithSpanGen, plus the live-rasterizer accessor identity) and
        public-surface tests `TestPublicAgg2DFloatDrawPathDefaults`,
        `TestPublicAgg2DFloatRasterizerEscapeHatches` in `agg2d_float_test.go`
        (DrawPathNoTransformDefault verified to ignore an off-buffer world translate).
  - [ ] **Gouraud shading** (`GouraudTriangle`): the only non-mechanical item —
        requires a float Gouraud span generator (`span_gouraud_rgba128`, the float
        twin of the 8-bit `span_gouraud_rgba`) before the public method can delegate,
        analogous to the gradient/image-filter float twins already built in L5.

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
