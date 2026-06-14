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

## Phase 4 - Explicit Float Agg2D Variant — DONE

AGG 2.6's `Agg2D` has a compile-time `AGG2D_USE_FLOAT_FORMAT` switch
(`../agg-2.6/agg-src/agg2d/agg2d.h`) that swaps the internal `ColorType` from
`agg::rgba8` to `agg::rgba32`. The Go port now provides this as an explicit,
additive float twin — `Agg2DFloat`/`ContextFloat`/`ImageFloat` — selected purely
by construction (no build tags) and with the 8-bit `Agg2D` untouched. The twin
mirrors the full 8-bit public surface and is verified for cross-precision parity.

- **Float pixel stack (named by 128-bit pixel width to avoid the 8-bit
  `PixFmtRGBA32` = 32-bit-pixel alias collision; pairs with `color.RGBA32`
  float):** blender `internal/pixfmt/blender/rgba128.go`
  (`BlenderRGBA128{,Pre,Plain}`, `lerp`/`prelerp`, cover ∈ [0,1]) and pixfmt
  `internal/pixfmt/pixfmt_rgba128.go` (`PixFmtRGBA128{,Pre,Plain}` over
  `*buffer.RenderingBufferF32`), structural twins of the float `gray32` stack.
  Composite variants in `blender/rgba128_composite.go` +
  `pixfmt_composite_rgba128.go` reuse the 8-bit `CompositeBlender.blendOperation`.
- **Internal + public twin:** `internal/agg2d/agg2d_float.go` mirrors `Agg2D`
  field-for-field with float types; root `agg2d_float.go` + `context_float.go`
  expose the public API. The public `Color` stays 8-bit (srgba8); `colorToRGBA32`
  bridges at the boundary. Rasterizer, scanline, transform, path, curve/stroke/
  dash converters, font/glyph cache, gradient/image-filter/Gouraud span bases are
  color-agnostic and reused as-is — only the pixfmt/blender/color LUTs differ.
- **Boundary contract** (`internal/agg2d/buffer_float.go`): `ImageFloat` stores
  **straight** RGBA float32 ([0,1], 4/pixel); premul/demul happens inside the
  pixfmt blenders, identical to 8-bit. Conversions honor each format's alpha
  convention: `ToNRGBA64`/`ToRGBA`/`ToImage8` (+ inverses).
- **Full surface coverage:** clear/fill/stroke, paths (incl. relative + smooth
  curves), shapes (Arc/RoundedRect\*/Star/Polygon/Polyline/Curve/Parallelogram),
  dashed strokes, gradients (linear/radial + D1/D2 + N-stop multi-stop),
  affine/perspective image transforms + copy/blend/PPM export, viewport &
  coordinate mapping, transform stack & affine matrix, state accessors + C++-style
  alias setters, composite blend modes, text glyph rendering, DrawPath escape
  hatches (`GetInternalRasterizer`/`RenderRasterizerWithColor`/`ScanlineRender`/
  `RenderScanlinesAAWithSpanGen`), and Gouraud shading. The one genuinely new
  piece was the float Gouraud span generator
  `internal/span/span_gouraud_rgba128.go` (color-agnostic `SpanGouraud[C]` base
  reused; per-edge calc + horizontal Generate reimplemented in straight float
  space). Each subsystem has its own `*_float.go` builder + root wrapper.
- **Verification:** cross-precision parity tests render the same scene through
  both pipelines and compare quantized output (solid tol 1, gradient/Gouraud/
  transform tol ≤ 3, AA ≤ 4); source-linked premul/demul tests; a visual hook
  (`tests/visual/float_path_test.go`, `float_image_transform_test.go`). All float
  files are gofmt/vet/golangci-lint clean.
- **Documented deviations** (`docs/AGG_DELTAS.md` "Float Agg2D Variant"):
  no-build-tag selection, `RGBA128`/`color.RGBA32` naming, 8-bit public `Color`,
  the straight-data boundary contract, the float bilinear's omission of AGG's
  +0.5 integer rounding bias, and the whole-image (not region-cropped)
  `BlendImageDefaultAlpha`.

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

- [x] Support these common high-level operations through both engines:
      rectangle/circle helpers, `MoveTo`/`LineTo`/`QuadTo`/`CubicTo`/`Close`,
      fill/stroke rendering, affine transforms, clip box control, solid fills,
      dashed strokes (`AddDash`/`RemoveAllDashes`/`DashStart`/`GetDashStart`,
      driving AGG `conv_dash` on both backends), basic image copy/scale/quad
      mapping, and compositing mode selection.
- [x] Finish the port-backed facade coverage for operations already present on
      the root `agg.Context`, especially clip box, image-region helpers,
      gradients, and text-related APIs.
- [x] Decide and implement the first getter/readback subset exposed by the
      facade v1 contract: fill-rule state, blend mode state, gradient type
      state, clip-box readback, text-hint state, and text metrics/bounds.
- [x] Decide whether backend-neutral transform-matrix readback belongs in the
      facade v1 contract or remains intentionally out of scope. Decision: IN
      scope. `GetTransform()` returns the cumulative affine as an
      `agg.Transformations` value (AGG order); the port delegates to
      `agg.Context.GetTransform`, the C++ backend reads its native matrix back
      via an `agg_go_cpp_matrix_store` bridge call. Round-trips exactly on both
      backends (tagged tests in `engine_test.go`/`engine_aggreal_test.go`).
- [x] Decide whether additional style/state getters from the root API, such as
      current colors, line width, line cap, and line join, belong in the facade
      contract or should stay write-only for v1. Decision: IN scope. Added
      `GetFillColor`/`GetStrokeColor`/`GetLineWidth`/`GetLineCap`/`GetLineJoin`,
      symmetric with the setters and the existing getter subset; the port
      delegates to the root getters and the C++ backend returns stored Go-side
      style state.
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
- [ ] Extend that partial C++ parity work to the remaining shared-facade gaps.
      Done: dashed strokes (real-native `agg::conv_dash` → `agg::conv_stroke`);
      the supported compositing set (`src`/`srcover`/`clear`/`alpha`/`dst`) now
      renders solid fills/strokes directly through a comp-op pixfmt with a
      straight-alpha adaptor, so `compositing_src`/`srcover`/`clear` are
      byte-exact (promoted from logged divergences to strict). Remaining: blend
      modes beyond the supported five, gradient fills under a non-src-over blend
      (still CPU-layer composited), transformed image-region behavior, dashed
      strokes under a non-identity transform, and image paths that still bypass
      AGG in the real-native build.
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

- [x] Add a backend-neutral scene corpus in `agg_go` that renders the same
      operations through both engines (`engine/scene`: `Scene`/`All`/`Filter`,
      per-engine `BuildAssets`, capability-declared scenes).
- [x] Cover at least these comparison scenes: solid fill/stroke, dashed stroke,
      self-intersecting paths with both fill rules, affine/scaled image
      transforms, clip boxes, gradients (linear+radial), compositing, and one
      text scene. The `dashed_stroke` scene compares within the path envelope on
      both backends (AGG `conv_dash`); see `docs/BACKENDS.md` "Dashed strokes".
- [x] Add a render workflow that emits `port`, `cpp`, and diff outputs for
      manual inspection and parity triage (`cmd/engine-compare`).
- [x] Add a benchmark workflow that runs the same scene corpus through both
      engines and records runtime/allocation data from the same high-level
      description (`tests/conformance/BenchmarkCorpusRender`).

### 5.9 Verification and exit criteria

- [x] Add first-cut unit tests for `engine.Available()`, default engine
      selection, explicit unavailable C++ requests, and port-backed image
      drawing through the facade.
- [x] Add tests covering blank-image creation, caller-managed buffers, attached
      contexts, examples, and typed engine-mismatch errors.
- [x] Add unit tests for backend selection, capability discovery, and explicit
      unavailable/stub-rejected error paths.
- [x] Add cross-backend conformance tests that render the same scene through
      both engines and compare outputs with exact-match or documented tolerance
      envelopes (`tests/conformance/TestCrossBackendConformance`: per-class
      tolerance envelopes and capability-gap skips, envelopes documented in
      `docs/BACKENDS.md`). The compositing scenes are now strict and byte-exact;
      the `knownDivergence` logging mechanism is retained (currently empty) for
      future partial features.
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
