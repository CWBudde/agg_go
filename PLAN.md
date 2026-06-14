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

All implementation work for the final library lives in this repository. The
opt-in `engine` facade now renders through either the pure-Go `Port` backend or
an in-repo C++ AGG-backed `CPP` backend (build tags `agogo aggreal`, linking
system `libagg`/`freetype2`); nothing imports the external `github.com/cwbudde/agogo`
module. `../AGoGo` remains only a read-only oracle for auditing edge cases. The
end state is a single repository that can be renamed back to `AGoGo`.

The foundation (5.1–5.4, 5.6–5.8) and most verification (5.9) are **done**; what
remains is closing the last C++ parity gaps, hardening the real-native build's
availability story, and the final rename. The sections below preserve their
numbers because `docs/BACKENDS.md` and `tests/conformance/` cross-reference §5.5
and §5.9 by number.

### 5.1–5.4 Facade, API boundary, and v1 scope — DONE

The backend-neutral facade is complete for its v1 surface and both engines
implement it.

- **API boundary (§5.2):** root `agg` package stays concrete and pure Go; the
  backend-selectable surface lives in a separate `engine` package; no cgo files
  in root `agg`; the native C++ layer is package-private inside `engine`; callers
  not opting into backend selection stay source-compatible. No bridge-plus-adapter
  architecture was reintroduced.
- **Facade shape (§5.3):** `engine.Kind` (`Port`/`CPP`), `engine.Config`,
  `engine.Available()`, `engine.NewContext`/`NewContextForImage`/`NewImage`/
  `NewImageFromGoImage`/`NewImageFromBuffer`; the narrow shared surface (clear,
  fill/stroke color, line width/cap/join, path construction, fill rules, clip box,
  transforms, basic image drawing, compositing, image export incl.
  `ToStandardImage`/JPEG) is exposed and no wider. Unsupported operations return
  typed capability errors; engine/resource-mismatch is a typed error; package
  docs and a runnable example exist.
- **v1 scope (§5.4):** all the shared high-level operations (shapes, path verbs,
  fill/stroke, affine transforms, clip box, solid fills, dashed strokes via AGG
  `conv_dash`, image copy/scale/quad, compositing-mode selection) work on both
  engines, with port coverage finished for clip/image-region/gradients/text.
  Getter contract decided **IN scope** and implemented symmetrically with the
  setters: fill-rule/blend-mode/gradient-type/clip-box/text-hint state, text
  metrics/bounds, `GetFillColor`/`GetStrokeColor`/`GetLineWidth`/`GetLineCap`/
  `GetLineJoin`, and `GetTransform()` (cumulative affine in AGG order; the C++
  backend reads its native matrix via the `agg_go_cpp_matrix_store` bridge).
  Round-trips exactly on both backends (`engine_test.go`/`engine_aggreal_test.go`).

### 5.5 In-repo C++ engine — DONE except the parity gaps below

The in-repo `agogo`-tagged native layer is self-contained (local header/source,
cgo config, probes, build-mode tests). The real AGG-backed build (`agogo aggreal`)
makes `engine.CPP` available and ports image scale/quad, clip box, the supported
compositing subset, gradients, dashed strokes, and a first text slice — all
package-private, behind availability/capability checks, never silently falling
back to the port or accepting a stub as valid. Compositing for **solid** fills/
strokes renders directly through a comp-op pixfmt with a straight-alpha adaptor
(`compositing_src`/`srcover`/`clear` byte-exact, strict); **gradient** fills/
strokes under every supported blend mode composite the recoloured layer through
the same operator using the shape's AA coverage as cover (`compositing_gradient`).

**Remaining C++ parity gaps** (each currently surfaces as a typed capability error
or a documented conformance skip, never a silent wrong render):

- [ ] **Blend modes beyond the supported five** (`alpha`/`clear`/`src`/`dst`/
      `srcover`). Other `agg.BlendMode` values fail with
      `ErrUnsupportedCapability`; extend the comp-op path to the full AGG set.
- [ ] **Transformed image draw.** `DrawImageRegion` under an active transform is
      rejected; the `image_affine` scene skips on `cpp`. Route image draw through
      AGG's affine image span path so it works under a non-identity CTM.
- [ ] **Image draw under a non-src-over blend.** Image paths still use the CPU
      composite helper rather than the AGG comp-op operator, so image draw under
      `src`/`clear`/etc. is not yet a guaranteed parity case. The
      `agg_go_cpp_image_composite_cover` primitive added for gradients can be
      reused here. (This subsumes the former "replace remaining CPU helper paths"
      item — it is specifically the image paths that still bypass AGG.)
- [ ] **Dashed strokes under a non-identity transform.** The port dashes in user
      space; the C++ backend dashes the pre-transformed (device-space) path, so
      dash lengths diverge under a transform. The `dashed_stroke` scene therefore
      uses no active transform. Reconcile the dash-space convention.

### 5.6–5.8 AGoGo audit, trust boundaries, and comparison layer — DONE

- **Absorption + audit (§5.6/§5.7):** `../AGoGo/go` and `../AGoGo/cpp` were
  audited; reusable knowledge was carried over and the old standalone-bridge Go
  wrappers/tests were dropped in favour of the direct `engine`-local native
  design. Every stub/fallback/"not implemented" path is classified as supported,
  explicitly unavailable, or comparison-only, with a hard guard rejecting the C++
  engine when the build produced only a stub. `docs/BACKENDS.md` records the
  capability matrix, native dependencies, and gaps; stale AGoGo docs were
  reconciled to the "single repo, Go-first, optional in-repo C++ engine" story;
  partial SVG/text/pattern behavior is kept out of the facade.
- **Note — obviated items:** the donor-repo default-fallback enum behavior and the
  audit's build breakages (duplicate `abs`/`compareImages`, missing
  `CAPIImageGetBuffer`, stale exported `LineCapRound`/`LineJoinRound`) do **not**
  apply to the in-repo design: the native layer was written fresh with
  package-private constants and typed errors for unknown paint/compositing/pixel
  values, and never imports that glue. No such symbols exist in this repo.
- **Comparison & benchmark layer (§5.8):** the backend-neutral scene corpus
  (`engine/scene`: `Scene`/`All`/`Filter`, per-engine `BuildAssets`,
  capability-declared scenes) covers solid fill/stroke, dashed stroke, both fill
  rules, affine/scaled image, clip box, linear+radial gradients, the compositing
  subset (incl. `compositing_gradient`), and a font-skip-gated text scene.
  `cmd/engine-compare` emits port/cpp/diff PNGs; `tests/conformance/
BenchmarkCorpusRender` runs the corpus through every available engine.

### 5.9 Verification and exit criteria — mostly done

- [x] Unit tests for `engine.Available()`, default selection, unavailable/stub-
      rejected C++ requests, blank-image/caller-buffer/attached-context paths,
      examples, capability discovery, and typed engine-mismatch errors.
- [x] Cross-backend conformance (`tests/conformance/TestCrossBackendConformance`):
      per-class tolerance envelopes (documented in `docs/BACKENDS.md`) and
      capability-gap skips. Compositing scenes are strict/byte-exact; the
      `knownDivergence` mechanism is retained (currently empty) for the next
      partial feature.
- [ ] **CI build/test gate for the real C++ backend.** A green `agogo aggreal`
      (+`freetype`) build/test run must gate the backend being advertised as
      supported, so a broken native build can't ship as a working `cpp` engine.
- [ ] **Keep behavioral-difference docs current.** As the parity gaps in §5.5
      close, update `docs/BACKENDS.md`, and record any rendering-semantics
      deltas in `docs/AGG_DELTAS.md`. (Ongoing maintenance item, not a one-shot.)
- [ ] **Retire `../AGoGo` from normal workflows.** Confirm nothing in routine
      development, verification, or release validation still needs the external
      repo (no runtime dependency remains; this is the final sign-off before the
      rename).

### 5.10 Final rename and consolidation — open

- [ ] Rename this repository back to `AGoGo` once `../AGoGo` is redundant.
- [ ] Update `go.mod` from `github.com/cwbudde/agg_go` to the final module path,
      and fix every internal doc/example/CI/badge/link/generated reference that
      still says `agg_go`.
- [ ] Decide and document the importer-compatibility story: whether the package
      name stays `agg`, whether module redirects are relied on, and whether a
      temporary migration note or deprecated mirror is needed.

### 5.11 Non-goals (and deliberate v1 deferrals)

- Do not turn root `agg` into an interface-only abstraction layer.
- Do not abstract the low-level rasterizer/scanline/pixfmt internals or expose
  the full low-level AGG pipeline through the facade in this pass (deferred §5.4
  item — out of scope until a concrete need appears).
- Do not add demo-by-demo backend switching to the first public cut unless a demo
  already uses only the supported shared surface (deferred §5.4 item).
- Do not keep two actively developed repositories once migration completes.
- Do not rely on stub mode or undocumented fallbacks to claim engine support.
- Do not move the center of gravity for pure-Go rendering work out of this repo.

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
