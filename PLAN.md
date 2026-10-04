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

> **Re-measured 2026-10-04 (after the Phase 8 P0/P1 fixes)**, fresh
> `UPDATE_VISUAL=1` render on **darwin/arm64** vs `reference/cpp/examples`, RGB pixel-diff count.
> The older rows were probably measured on linux/amd64 (no FMA), so small residuals in
> "exact" rows such as `bspline` may be FMA artefacts (see the 8.2 FMA sweep).
> - **Now 0 px, rows below are stale:** `bezier_div`, `image_filters2`, `multi_clip`,
>   `gouraud` (352 → 0 px from the rasterizer gamma `uround` fix).
> - **Much smaller than listed:** `conv_dash_marker` 34 px (max Δ2), `idea` 60 px.
> - **Unchanged:** the rest (e.g. `scanline_boolean2` 69206, `compositing` 95640,
>   `alpha_mask3` 69077, `pattern_fill` 64555, `line_thickness` 25756, `alpha_gradient`
>   26835, `aa_test` 10685, `graph_test` 5551).
> - **Not tracked as open rows but non-zero:** `bspline` 2, `flash_rasterizer` 5,
>   `flash_rasterizer2` 3, `lion_lens` 1, `lion_outline` 18, `simple_blur` 18,
>   `rasterizers2` 248.
>
> Re-baseline the rows when 8.3 (live parity rendering) lands.

- [ ] `distortions` — RMSE 0.0377 (53 px). A few isolated extreme pixels from distortion-resampling rounding in the lensed image/sphere; essentially done.
- [ ] `polymorphic_renderer` — RMSE 0.1239 (91 px). Float-vs-8bit AA edge-coverage rounding on the single triangle's anti-aliased edges only.
- [ ] `gradients` — RMSE 0.0335 (118 px). AA on the gradient-control spline curve lines and a couple of sphere-edge pixels; the gradient fill itself matches.
- [ ] `idea` — RMSE 0.7113 (171 px). A few extreme AA pixels on the tiny high-chroma lightbulb rays plus the step/degree label text.
- [x] `rasterizers2` — RMSE 0.8059→0.0376 (248 px, max channel diff 4). Real bug: `RendererPrimitives.Coord` truncated (`int(c*256)`) where C++ `renderer_primitives::coord` rounds (`iround`); the raw-`cos/sin` subpixel-Bresenham spiral landed a few arms one pixel off (brown↔cream swaps, max diff 159, dominating the RMSE). Fixed to `basics.IRound` (the aliased pixel-rounded spiral was already exact because `roundoff` floors first). Residual 248 px are ≤4 LSB image-pattern blend rounding on the arm edges — verified-faithful float-noise floor, same class as `lion_outline`/`image_alpha`. Shared `Coord` users (`gradients_contour`, `multi_clip`, `alpha_mask2`) confirmed still RMSE 0.0.
- [ ] `multi_clip` — RMSE 1.2046 (40 px). Only 40 px but a few are extreme: sub-pixel AA edge rounding on the dense thin random strokes/circles inside the clip cells.
- [x] `blur` — pixel-exact (RMSE 0.0, 0/145200 px). The shadow polygon control (`shadowCtrl`) was rendered after the "a" shape, placing it in front; C++ `on_draw()` renders it after the blur but before the shape. Moving the `renderCtrl(shadowCtrl)` call to that position fixed the z-order.
- [ ] `line_patterns` — RMSE 0.1216 (936 px). Image-pattern glyph sampling/positioning along each curved path plus a couple of saturated control-pin pixels.
- [ ] `gamma_ctrl` — RMSE 0.0644 (1378 px). Sub-pixel AA edge fringing on the green GSV "Text 2345" glyph outlines and the thin radial-spline lines; controls exact.
  → 2026-10-04: the rasterizer gamma-table truncation is fixed (§8.2, `SetGamma` now uses `uround`), but the re-measurement above still lists this row as unchanged. Still open: the glyph-outline and spline-line AA.
- [ ] `trans_polar` — RMSE 0.0506 (1628 px). Transform-resampling AA on the curved polar ring plus the control text and slider-knob X positions.
- [ ] `conv_stroke` — RMSE 0.0803 (1709 px). Faint float-vs-8bit AA edge fringing along the dashed-stroke borders and miter-join markers; near-exact.
- [ ] `mol_view` — RMSE 0.2857 (2138 px). Sub-pixel AA fringing on the green GSV title-text glyph edges and the thin atom-bond strokes; geometry/colors already corrected.
- [ ] `line_patterns_clip` — RMSE 0.1131 (3300 px). Patterned-stroke dash phase / clip-boundary sampling on the X-crossing lines and clipped line ends; edge AA plus stray control-pin pixels.
- [ ] `compositing2` — RMSE 0.0967 (5004 px). Comp-op blend rounding (8-bit vs float) on the edges of the four overlapping translucent circles; controls exact.
- [ ] `aa_test` — RMSE 0.1728 (10685 px). Float-vs-8bit AA fringing on the many thin anti-aliased lines/dashes in the radial sub-pixel line fans; no logic error.
- [ ] `alpha_gradient` — RMSE 0.4158 (26799 px). Accumulated 8-bit blend rounding (agg.RGBA truncates `uint8(v*255)` instead of round-to-nearest `*255+0.5`) across the whole alpha-blended gradient circle and translucent ellipses; the round-to-nearest fix is one-line but touches a shared blend path.
  → 2026-10-04: the rasterizer gamma-table rounding fix (§8.2) did not move this row (26835 px in the re-measurement above). The `agg.RGBA` truncation (`colors.go:153-156`) is still the lead.
- [ ] `line_thickness` — RMSE 0.3650 (25756 px). Uniform BGR96-float-vs-8bit edge-AA fringe along every diagonal line and radial spoke; essentially done pending a float renderer.
- [ ] `graph_test` — RMSE 0.7717 (37004 px). Sub-pixel AA on the grid of node-circle outlines plus glyph edges in the bottom timing/status text; residual after per-control text-height fixes.
- [ ] `pattern_fill` — RMSE 0.2836 (64555 px). Background tint off by integer-rounding the premultiplied RGBA8(102,0,26,26) instead of float premultiply-then-quantize (rgba_pre), spread across the pattern-filled star interior; controls/margins clean.
- [ ] `alpha_mask3` — RMSE 0.3438 (69120 px). Renders into 4-channel RGBA32 instead of the C++ opaque 3-channel BGR24 (pixfmt_rgb) buffer, so the layered low-alpha (25/127) over-blend rounds one LSB darker across the translucent shapes; controls/background identical.
- [ ] `conv_dash_marker` — RMSE 0.9998 (4672 px). Dash-phase / sub-pixel dash-segment positioning offset along the dashed line (every dash lands slightly shifted) plus the green smooth-outline edges.
  → 2026-10 audit: the geometry stages (smooth_poly1, curve3_div, vcgen_dash, markers_term) look numerically equivalent; the plain smooth outline differs too, so suspect thin-stroke AA downstream. Dump the stroke vertices from both sides to confirm.
- [ ] `bezier_div` — RMSE 1.1956 (2861 px). Stroke vertex generation at the Miter-Revert + Inner-Round join near the curve cusp differs slightly from C++ vcgen_stroke; diff concentrates at the inner-join triangle fan and dashed inner-stroke outline.
  → 2026-10-04: the swapped `curve4_div` case selector is fixed (§8.2), and the re-measurement above shows this row at 0 px, so the description is stale. Re-baseline it with 8.3 and tick it if 0 px is confirmed.
- [ ] `compositing` — RMSE 0.5101 (98059 px). ±1-LSB gradient/composite interpolation rounding in the 8-bit-linear scene path (the known RGBA128 float comp-op residual) spread across the gradient-filled shapes; controls/text exact.
- [ ] `scanline_boolean2` — RMSE 1.2047 (69301 px). Sub-pixel cover/span-boundary discrepancy in the scanline boolean AND-combine path (num_spans 1033 vs C++ 1031) on the intersection-shape AA edges; GSV text stroke already corrected.
  → 2026-10 audit: the sbool AND helpers match C++; diff the input storages first (the demo's contour round-trip is the likely cause, §8.2).
- [ ] `image_filters2` — RMSE 1.2622 (53494 px). Largest real gap: the scaled right-side image is rendered via Agg2D's dedicated bilinear resampler instead of the C++ LUT-based span_image_filter_rgba general filter, so every fractional sample blends source texels differently across the whole image; control panel clean.
  → 2026-10 audit: this description is stale — the example already uses `SpanImageFilterRGBA`. The 2026-10-04 re-measurement above shows 0 px, so re-baseline it with 8.3 and close the matching §8.2 `image_filters2` item if 0 px is confirmed.

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

The foundation (5.1–5.4, 5.6–5.8) and most verification (5.9) are **done**, and
the C++ backend's parity gaps plus the one port-side comp-op bug they surfaced
(§5.5) are all closed. What remains is keeping the behavioural-difference docs
current (§5.9) and the final rename (§5.10). The sections below preserve their
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

### 5.5 In-repo C++ engine — DONE (all parity gaps closed)

The in-repo `agogo`-tagged native layer is self-contained (local header/source,
cgo config, probes, build-mode tests). The real AGG-backed build (`agogo aggreal`)
makes `engine.CPP` available and ports image scale/quad, clip box, compositing,
gradients, dashed strokes, and a first text slice — all package-private, behind
availability/capability checks, never silently falling back to the port or
accepting a stub as valid. Compositing renders through a comp-op pixfmt with a
straight-alpha adaptor that mirrors the port's `CompositeBlenderPlain`: **solid**
fills/strokes are byte-exact (`compositing_src`/`srcover`/`clear`, strict);
**gradient** fills/strokes composite the recoloured layer through the same
operator using the shape's AA coverage as cover (`compositing_gradient`). All
paths honour the **full AGG operator set** — every `agg.BlendMode` (Porter-Duff +
separable) maps 1:1 onto `comp_op_e` via `map_comp_op`, dispatched through AGG's
`g_comp_op_func`; the single `requireBlendMode` / `supported_comp_op_mode` gate
accepts the whole enum.

**Closed parity gaps** (each formerly a typed capability error or documented
conformance skip — never a silent wrong render; retained as a reconciliation
record, with the parity-relevant deviation each fix uncovered):

- [x] **Full blend-mode set (vector + gradient + image + text).** Vector/gradient
      paths dispatch every `comp_op_e` (`compositing_multiply` byte-exact). Image
      (scaled/quad) blits composite per-pixel via `comp_op_adaptor_rgba_plain`
      (`blend_image_pixel`, full cover on covered pixels only); text routes every
      mode through `compositeCoverFrom` with layer alpha as per-pixel cover, so
      clear/src cannot wipe the background. Mirrors the port's comp-op base
      renderer (`renBaseCompPre`). The single `requireBlendMode` gate replaced the
      former `requireImageBlendMode`. Locked by `TestCPPExtendedBlendModesRenderWithAggReal`,
      `TestCPPXorBlendIsAGGFaithfulWithAggReal`,
      `TestCPPImageDrawUnderExtendedBlendModeIsFaithfulWithAggReal`,
      `TestCPPTextUnderExtendedBlendModePreservesBackgroundWithAggReal`,
      `TestCPPBackendExtendedBlendModeOnDrawImageQuad`; strict `image_blend` scene
      (image over a colour field under multiply) agrees at ~0.053 (Tol 4 / ratio
      0.08, image-sampler-noise class).
- [x] **Transformed image & vector draw.** `DrawImageRegion` (and the `DrawImage`/
      `DrawImageScaled` delegating to it) map the dest-rect corners through the
      active matrix and blit via the quad path, mirroring the port's `renderImage`.
      Strict `image_affine` scene (Tol 4 / ratio 0.10, CPP nearest-neighbour vs
      Port bilinear). _Deviation fixed:_ the native matrix composed `Translate`/
      `Rotate`/`Scale` in reverse of `agg::trans_affine`; corrected via
      `matrix_premultiply` (primitive pre-multiplied in output space), which also
      fixes transformed vector rendering. Locked by
      `TestCPPTransformedImageDrawRendersWithAggReal`,
      `TestCPPTransformComposeOrderMatchesPortWithAggReal`,
      `TestCPPNativeMatrixTransformPointTranslateRotateScale`.
- [x] **Dashed/plain strokes under a non-identity transform.** The native stroke
      functions take a trailing `const AggGoCPPMatrix*` and apply it to the stroked
      outline via `agg::conv_transform` after dash+stroke (`add_stroke_to_ras`):
      `path -> dash -> stroke -> transform`, identical to Agg2D and the port's
      `addStrokeToRasterizer`. `Stroke()` passes the **user-space** path + matrix,
      so dash period and line width scale with the transform; a null/identity matrix
      keeps the direct-rasterize path (no-transform scenes byte-identical). Strict
      `dashed_stroke_transform` scene is **byte-exact (0/65536)**; locked by
      `TestCPPDashedStrokeUnderTransformDashesInUserSpaceWithAggReal`. (Stub build,
      never advertised, ignores the matrix.)

**Port-side comp-op bug (surfaced by the CPP work) — fixed:**

- [x] **Port stored premultiplied data in its straight buffer** for comp-ops whose
      result is _translucent_ over an opaque destination (`xor`, `dst-out`, and the
      `src-in`/`dst-in` family). Root cause: `internal/pixfmt/pixfmt_composite.go`'s
      `BlendHline`/`BlendSolidHspan` took a "SIMD fast path" through the
      `simd.Comp*HspanRGBA` kernels, which operate on a **premultiplied** destination
      and leave a premultiplied result — but this pixfmt stores **straight** alpha,
      so the per-pixel premultiply-on-read / demultiply-on-write bridge that
      `blender.CompositeBlenderPlain` performs was skipped. It only showed when the
      result alpha < 255 (src-over/clear stayed correct: opaque result ⇒ premult ==
      straight, hence those scenes were byte-exact while xor read back too dark).
      Fix: drop the premult-dst SIMD fast path from the straight composite pixfmt and
      always route through the scalar `CompositeBlenderPlain` (the SIMD kernels were
      only ever wired to this straight pixfmt and only `SrcOver` had a real vector
      kernel, so the cost is limited to explicit non-default blend modes; the comp
      pixfmt is bypassed entirely for the default `BlendAlpha`). The float path
      (`pixfmt_composite_rgba128.go`) was already correct (no SIMD, scalar
      `CompositeBlenderRGBA128Plain`). New strict `compositing_xor` and
      `compositing_dstout` corpus scenes now agree cross-backend (0 px over tolerance
      2; max 1 LSB from float-demul vs CPP integer-demul rounding); the CPP side
      stays locked by `TestCPPXorBlendIsAGGFaithfulWithAggReal`.

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

## Phase 6 - Composite Fast Path (performance) — DONE

Accelerated the straight-alpha composite pixfmt (`internal/pixfmt/pixfmt_composite.go`
and the float `pixfmt_composite_rgba128.go`) — reached only for explicit blend modes
(the default `BlendAlpha` uses the already-SIMD `renBase` and was unaffected) —
without reintroducing the §5.5 premultiplied-storage bug. Every fast path holds the
three fidelity constraints: storage stays **straight alpha** (the deliberate
deviation from stock AGG's premultiplied `comp_op_adaptor_rgba`; both backends agree
on it, locked by `TestCPPXorBlendIsAGGFaithfulWithAggReal`); results reproduce the
scalar `to8(res.r/res.a)` demultiply rounding (1-LSB cross-backend envelope); and
they are correct over **translucent** — not just opaque — destinations.

- **Hoisted span methods (all 24 operators; 8-bit + float; solid + per-pixel-colour).**
  `CompositeBlenderPlain.BlendSolidSpanStraight` / `BlendColorSpanStraight` (8-bit) and
  the `CompositeBlenderRGBA128Plain` twins (float) replace per-pixel interface dispatch
  — and, for colour hspans, a per-pixel row refetch + bounds check — with one concrete
  call per span, reusing the shared `blendOperation` so they are **byte-for-byte
  identical** to per-pixel `BlendPix`. Wired into `BlendHline`/`BlendSolidHspan`/
  `BlendColorHspan` via the `straightSpanBlender[…]` interfaces (the premultiplied-source
  Pre blender doesn't implement them → per-pixel fallback). **~1.7–2.0×, 0 allocs**,
  conformance byte-unchanged. Locked by the differential + pixfmt-wiring tests in
  `internal/pixfmt/blender/` and `pixfmt_composite_test.go`. `BlendColorVspan` stays
  per-pixel (each pixel a different row → no contiguous-span hoist).

- **True SIMD tier for uniform-coverage SrcOver** (the common large-solid-span case).
  Bit-exact float64 kernels in `internal/simd/comp_plain_*`: AVX2 (**2.28×**, one pixel
  per 256-bit register) with an SSE2 fallback (**1.9×**, two 128-bit registers/pixel)
  for pre-AVX2 amd64; dispatch is AVX2 → SSE2 → scalar. float64 throughout, no FMA,
  `VCVTTPD2DQ` truncation matching `uint8(v*255+0.5)`. Other operators / AA edges /
  ARGB-ABGR fall through to the (already ~2×) scalar bridge. An integer fixed-point
  kernel (fails constraint 3 — max 125 LSB over a translucent dst) and a generic
  method-expression dispatch (allocates) were measured and rejected; the float divide
  is the correctness-mandated speed floor.

- **Option B (AGG-native premultiplied comp buffer) — rejected.** Storing premultiplied
  data so the premult-dst kernels apply with no per-pixel divide would reverse the
  straight-alpha deviation, pushing a demultiply onto every read path (`GetPixel`/
  `ToGoImage`/alpha masks/conformance comparator), reintroducing the §5.5 bug class, and
  breaking the locked port↔CPP straight-storage agreement — all for an already-
  accelerated non-default path. Revisit only if AGG-native premultiplied storage is
  adopted repo-wide (CPP side + all straight readers moving together).

---

## Phase 7 - Image Resample Performance

Downscaling a photograph through `DrawImageAffine` is currently ~50× slower than
libvips and the gap widens as the reduction ratio *falls*. Benchmarked in
`../rasterbench` (12th Gen i7-1255U, Go 1.26.5, libvips 8.15.1, agg_go v0.5.0)
on a 3024² centre crop, decode excluded:

| output | reduction | libvips | agg_go |
| ---: | ---: | ---: | ---: |
| 48 px | 63× | ~0 | 21 ms |
| 128 px | 24× | ~0 | 624 ms |
| 192 px | 16× | ~0 | 3.3 s |
| 512 px | 5.9× | ~0 | 5.0 s |
| 1024 px | 3.0× | ~0 | 5.9 s |

Full pipeline including codecs: 574 ms for nine renditions vs 29.6 s. Pinning
libvips to one thread only moves it to 726 ms, so this is **not** a
single-threading problem — Go-vs-C accounts for a 2.2× decode gap at most.
Output quality is not the differentiator either: the two resamplers agree to
50 dB PSNR at 512 px and above.

Two mechanisms explain the shape of the agg_go column, both confirmed against
C++ AGG:

1. **`span_image_resample_rgba` convolves non-separably.** The inner loops in
   `agg_span_image_filter_rgba.h:643-760` (`span_image_resample_rgba_affine`)
   walk a `(2·radius_x)×(2·radius_y)` window per destination pixel, and
   `radius_x = (diameter · m_rx) >> 1` grows with the reduction ratio. Cost per
   destination pixel is therefore O(r²), and since r ∝ reduction while the
   destination area ∝ 1/reduction², total work is *constant* in output size
   once unclamped — which is the ~5 s plateau at 512 and 1024 px.
2. **`m_scale_limit` clamps the footprint.** *(Corrected 2026-10-04.)* The
   `DrawImageAffine` resample path (`internal/agg2d/image.go:119`) uses
   `SpanImageResampleRGBAAffine`, whose base `SpanImageResampleAffine` defaults
   to `scaleLimit: 200.0` (`internal/span/span_image_filter.go:143,166`), as C++
   `span_image_resample_affine` does (`agg_span_image_filter.h:104,114`). The limit
   of 20 at `agg_span_image_filter.h:191`, asserted in
   `span_image_filter_test.go:299`, belongs to the generic `span_image_resample`,
   which this path does not use. `prepare()` (`agg_span_image_filter.h:140-150`)
   compares the **product** `scale_x·scale_y` with the limit. For a uniform
   reduction s the clamp therefore starts at s > √200 ≈ 14.1, and from there the
   per-axis scale is `200/s`: the footprint *shrinks* as the reduction grows,
   down to the floor of 1. That is why 48 px is cheap, and also why 48 px is where
   agg_go and libvips disagree most (RMSE 6.7 vs 0.8 at 512 px): the footprint is
   far too narrow to band-limit the source.

All of 48/96/128 px (63×/31.5×/23.6×) are in the clamped regime. They are also
above 20×, so they cannot show a "below-20×" anomaly, and 192 px (15.75×) is
clamped too. The earlier "predicted 0.5/2.0/3.6 s" assumed a constant footprint
of 20 above 20:1, which this path does not have. With the actual `200/s` clamp, a
work ∝ (output px)²·scale² model scaled from 5.0 s at 512 px predicts about
13/203/642 ms and 3.25 s for 48/96/128/192 px, against the measured
21/220/624 ms and 3.3 s. That is a plausible explanation, not a confirmed one:
no output size between the clamp threshold and 512 px (14.1:1 to 5.9:1) has been
measured yet. Task 7.3 confirms or rejects it.

### 7.1 Separable two-pass for axis-aligned transforms

The highest-fidelity win, and the one to do first. AGG already computes its
weight as a **product** of two 1-D LUT lookups —
`weight = (filter[x_hr] * filter[y_hr] + image_filter_scale/2) >> image_filter_shift`
(`agg_span_image_filter_rgba.h:715`) — so the kernel is separable in all but
evaluation order. Convolving horizontally into an intermediate buffer and then
vertically turns O(r²) taps per destination pixel into O(r), which at 5.9:1
reduction with a diameter-6 Lanczos is ~35² = 1225 taps down to 2×35 = 70.

- Detect the axis-aligned case in `DrawImageAffine`: `shx == 0 && shy == 0` and
  no rotation, i.e. a pure scale plus translate. This is the case every
  thumbnailing caller hits. General affine (rotation, skew, perspective) keeps
  the existing single-pass path unchanged — a two-pass decomposition is not
  equivalent there.
- Intermediate buffer is `dstWidth × srcHeight` in the source's channel format.
  Size it from the clipped destination rect, not the full source.
- **Fidelity constraint:** the result must differ from the single-pass path only
  by intermediate rounding. Fix the intermediate to a wider accumulator
  (int32/int64 fixed point, or float64 in the RGBA128 variant) so the two-pass
  result stays inside a 1-LSB envelope of single-pass, and lock that with a
  differential test in the style of `TestCPPXorBlendIsAGGFaithfulWithAggReal`.
  If the envelope cannot be held, this becomes a documented deviation in
  `docs/AGG_DELTAS.md` and must be opt-in — not a silent fallback (Porting
  Rule 3).
- Applies to the 8-bit and float span families alike; wire through
  `internal/agg2d/image.go:118` and `image_transform_float.go:104`.

### 7.2 Integer box pre-shrink — opt-in deviation

What libvips does: `vips_resize` performs an integer box shrink first and
applies the expensive kernel only to the small residual factor. Stock AGG has no
such stage, so adding one **changes rendered output** and cannot be the default.

- New field on `ImageTransformOptions`, e.g. `PreShrink`, defaulting to off so
  existing callers are bit-unchanged.
- When on and the reduction exceeds a threshold, box-average by
  `floor(reduction / t)` and let the filtered path handle the remainder.
- Record in `docs/AGG_DELTAS.md` as an explicit, tested deviation with the
  measured quality delta against the single-pass path (PSNR per ratio), not just
  the speedup.
- Only worth doing **after** 7.1: a separable pass may already close enough of
  the gap that carrying a deviation is not worth it. Measure before committing
  to it.

### 7.3 Confirm the clamped-regime cost model

The 48–192 px measurements are all in the clamped regime (see the corrected
mechanism 2 above), so they cannot show a "sub-20×" anomaly. Their fall-off fits
the shrinking `200/s` footprint, but that is a model, not a measurement. Settle it
before optimising anything in that range, because an optimisation aimed at the
wrong mechanism is worse than none:

- Measure at least two output sizes in the unclamped range between 192 px and
  512 px (reduction below ≈14.1:1, for example 256 px and 384 px) in
  `../rasterbench`'s `cmd/curve` sweep. No such measurement exists yet. Check
  that the cost stays at the ~5 s plateau there.
- Profile or count taps per destination pixel at 48/96/128/192 px, and confirm
  that they follow `rx = uround(200/s · image_subpixel_scale)`.
- Confirm that the `m_rx`/`m_ry` derivation in `SpanImageResampleAffine`
  matches `agg_span_image_filter.h:135-155` exactly, including the
  `scale_xy > m_scale_limit` proportional rescale on line 140.
- Record whether the shrinking footprint at high reductions (the 48 px aliasing)
  is AGG-faithful behaviour to keep, or a reason to set a lower limit via 7.4.

### 7.4 Expose `scale_limit` and `blur` on the public affine API

`SpanImageResampleAffine` has `ScaleLimit`/`SetScaleLimit` and blur internally,
mirroring C++ (`agg_span_image_filter.h:120-128`), but neither is reachable
from `ImageTransformOptions`. A caller therefore cannot trade the aliasing above
the clamp threshold (≈14.1:1 with the default limit of 200) for a wider footprint,
nor tighten it for speed. This is a genuine API-surface gap
against C++ AGG rather than a performance change; it is cheap and it makes 7.3
measurable from outside the module.

### 7.5 Exit criteria

- [ ] Axis-aligned downscale is within 5× of libvips at 5.9:1 reduction on the
      `../rasterbench` workload (from ~94× today).
- [ ] Single-pass output is byte-unchanged for every general-affine case, and
      the axis-aligned path is within the documented 1-LSB envelope.
- [ ] Visual regression corpus green; any deviation documented in
      `docs/AGG_DELTAS.md` and opt-in.
- [ ] `scale_limit` / `blur` reachable from `ImageTransformOptions`.
- [ ] The clamped-regime cost model (7.3) is confirmed with tap counts and with
      unclamped measurements between 192 and 512 px, in this plan or in `docs/`,
      not merely optimised around.

### 7.6 Non-goals

- **Matching libvips overall.** libvips wins the full-pipeline benchmark largely
  on codecs (WebP encode, which neither AGG nor the Go standard library has at
  all) and on a threaded, streaming architecture. This phase targets the
  resampling stage only.
- **Multithreading the resampler.** Worth 1.3× at most on this workload, against
  the ~17× the separable pass is worth, and it would put a concurrency contract
  on a currently-serial public API. Revisit only after 7.1 lands.
- **Bundling image codecs.** Out of scope for a rasteriser; callers bring their
  own.

---

## Phase 8 - Review Remediation (2026-10-04 audit)

Source: an 8-area review against `../agg-2.6/agg-src`. Items marked ✔ were re-verified
against the C++ source by hand. Several existing tests *encode* the bugs below, so fix
the test together with the code and derive expected values from C++, not from Go.

Ratings at the time of the audit (0–10): fidelity — vertex 7, rasterizer 6, renderer/span 7,
pixfmt/color 4, Agg2D 5, fonts/ctrl 6; parity verification 4; test quality 4; simplicity 3;
duplication 3; architecture 5; Go idiom 5; public API 3; performance 5; tooling/CI 3;
docs 5; hygiene 4; demo organization 3. **Overall 4.5.**

Priorities: **P0** = blocks everything, **P1** = parity bug or parity-gate gap,
**P2** = simplification/API, **P3** = performance/polish.

### 8.1 P0 — Green tree and honest CI — ✅ DONE (2026-10-04)

- [x] Fix the broken build. The uncommitted deletion of `internal/font/interfaces.go`
      removed `font.SerializedScanlinesAdaptor`, which `internal/agg2d/text.go:829` and
      `text_float.go:648` still use. Restore it, or move it next to its users.
      `internal/font/freetype2` also needs `font.IntegerPathStorage` (see 8.4 item 12 for
      whether freetype2 survives). The deleted `internal/color/blender.go` is safe to drop.
      → 2026-10-04: Done: consumer-side interfaces (`renscan.RasterizerInterface` in agg2d, local `integerPathStorage` in freetype2); the deletions were kept.
- [x] `TestPixFmtCompositeRGBA32BlendPixelCompositeOps/xor` fails on arm64 (128 vs 127).
      Go fuses `x*y+z` into FMA on arm64. Make the float blend/composite expressions
      FMA-proof (`float64(x*y) + z`) so results are architecture-independent.
      → 2026-10-04: Done: `float64()` guards in `blender/rgba_composite.go` (0 FMADD/FMSUB left there; C++ oracle confirms 127) and `internal/image/filters.go`. Remaining FMA sites tracked under 8.2.
- [x] Add `//go:build` tags to `internal/platform/x11` and `internal/platform/sdl2`, so that
      `go build/vet/test ./...` and lint work without X11/SDL headers (lint is dead on macOS today).
      → 2026-10-04: Done: `x11` / `sdl2` tags, backend selection fixed for `-tags x11,sdl2`, `just build-platform`.
- [x] CI runs `go test ./...` (root, `tests/*`, `engine`, examples), not just `./internal/...`.
      Also: add a `pull_request` trigger, pin the golangci-lint/gofumpt/gci versions, read
      the Go version from `go.mod` in `deploy-wasm.yml`, add an arm64 job (NEON + FMA) and a
      `-tags freetype` job.
      → 2026-10-04: Done for the CI wiring: amd64+arm64 matrix, `pull_request` trigger, pinned linters, a `freetype` job, `go-version-file` in deploy-wasm. The matrix runs `go vet ./...`, but **not** the full `go test ./...`: `unit-tests.yml` runs `go test $(go list ./... | grep -v /tests/visual/primitives)`, because that package still fails (next item). Re-adding it is tracked there.

- [x] `tests/visual/primitives` fails at HEAD (`TestBlendModes`: blend_xor 9000 px and
      blend_src_over 6 px; `TestGradients`: 10 cases; the thin_line references are missing).
      Decide per case whether the Go golden image or the code is wrong,
      checking against C++ where it has an equivalent.
      → 2026-10-04: Done, decided per case:
      - **Gradients (10 cases):** stale goldens. They changed with 7d9e45a, which ports
        `rgba8T::gradient`; the new `TestCPPOracleColorGradient` checks `Color.Gradient` against
        a C++ `Agg2D::Color::gradient` hash.
      - **blend_src_over:** stale golden. It changed with 4e431e6, which removed the premultiplied
        SIMD comp kernels from the straight-alpha pixfmt. The new
        `TestCompositeBlenderPlainMatchesCppOracle` checks that bridge against stock AGG
        premultiply → comp_op → demultiply, within ±2 in premultiplied space.
      - **blend_xor:** the code was wrong. `Image.ToGoImage`/`ToStandardImage` put straight bytes
        into a premultiplied `*image.RGBA`. They now return `*image.NRGBA`: a breaking signature
        change, the user's choice, also applied to the `engine.Image` interface. The golden only
        moves by ±1 quantisation, 0 in premultiplied space.
      - **thin_line:** the missing references only log a warning; that test does not fail.
  - [x] Then drop the `grep -v /tests/visual/primitives` filter from `unit-tests.yml`, so CI
        really runs `go test ./...`. Until then this package is not covered by CI.
        → 2026-10-04: Done; `unit-tests.yml` and `just test-all` run `go test ./...`.
- [x] Clear the golangci-lint backlog (55 pre-existing findings in a local
      `golangci-lint run ./...` on darwin: gocritic 23, staticcheck 15, unused 10, revive 4,
      ineffassign 3). Then remove the `only-new-issues: true` setting that PR #5 adds to the
      golangci-lint step in `lint.yml`, so lint checks the whole tree again, not just new code.
      → 2026-10-04: Done. The real backlog was 149 on darwin and 153 with `x11,sdl2` on Linux.
      The 55 was golangci-lint's capped default output (3 identical issues, 50 per linter).
      Fixes, with no new `//nolint`:
      - `OutlineAARenderer.Line0-3` take `*LineParameters`, as C++ `const&` does. This also
        removed 32 old `//nolint`.
      - `CurrentBitmap` returns a `GlyphBitmap`.
      - Four files over 1500 lines are split by pure moves (`fonts`, `gpc`, `pixfmt_rgb_packed`,
        `simd/cpu_test`); `cmd/wasm/main.go` is split too.
      - Tests now use `GetGSE4x6` instead of the deprecated `GetSimple4x6Font`.
      - SDL uses `GetTicks64`.
      `only-new-issues` is removed.

### 8.2 P1 — Confirmed numeric parity bugs

Cross-cutting
- [ ] FMA sweep. Go fuses `x*y±z` on arm64, but the C++ references are x86 builds without
      contraction. `go build -gcflags=-S ./... 2>&1 | grep -E 'F(N?M(ADD|SUB))D'` found about
      1300 fused sites (distinct source lines) in 221 files at audit time; after the guards
      below it is 1228 lines in 210 files on darwin/arm64. The compiler writes the `-S` listing
      to stderr, so the `2>&1` is required; `-gcflags=all=-S` would also list the standard
      library. Already guarded: `curve3_div`/`curve4_div` in `curves.go`, `CalcPolygonArea`
      in `basics/math.go`, `blender/rgba_composite.go`, `image/filters.go`.
      Next, by parity impact:
  - `basics/math_stroke.go`, `CalcDistance`/`CalcSqDistance` (stroke/contour have 1-ulp drift)
  - `transform/*` (affine, perspective, bilinear)
  - `span/interpolator_persp`, gouraud, gradients
  - `renderer/outline*`, `vcgen/smooth_poly1`, `color/rgba8` / `rgba32`

  Add the `-S` grep (with `2>&1`) as an arm64 CI regression check for the guarded code.
  Only `rgba_composite.go` and `filters.go` have 0 fused sites. `curves.go` (38) and
  `basics/math.go` (21) are guarded only in the functions named above, so check those per
  function, not per file. The div recursion still has 8 fused `(a+b)/2` midpoint lines
  (`curves.go:319-320, 711-716`); confirm that they are exact, or guard them, before the check
  goes in.

Vertex pipeline

- [x] ✔ `internal/curves/curves.go:727-736` — the `curve4_div` case selector is swapped
      (+1 for d2, +2 for d3; C++ `agg_curves.cpp:426` uses `(d2>eps)<<1 | (d3>eps)`), so
      cases 1 and 2 run each other's bodies. This is the prime suspect for the `bezier_div`
      residual. Add golden point dumps for the 15 `bezier_div` test curves at several
      tolerances.
      → 2026-10-04: Done: the selector now mirrors C++, plus FMA guards in curve3/4_div; golden test `curve_div_golden_test.go` (27 cubics, 11 quads, C++ dumps, bit-exact). The `bezier_div` demo was already exact, so the bug was masked.
- [x] ✔ `internal/basics/math.go:140` `CalcPolygonArea` returns `math.Abs`, so the
      `vcgen_contour` auto-orientation (`internal/vcgen/contour.go:113`) always picks CCW.
      Port the signed `calc_polygon_area` (`agg_math.h:230`) and keep its summation order.
      Also drop the per-rewind slice allocation.
      → 2026-10-04: Done: signed area in the C++ summation order; contour computes it without allocating; C++ golden test for contour auto-orientation.
- [ ] `internal/vcgen/vertex_sequence.go:69` never emits `end_poly | flags` and uses a
      hand-rolled `shortenPath`. Closed paths through `ConvShortenPath` /
      `ConvMarkerAdaptor` therefore turn into open ones. Reuse a generic `array.ShortenPath`.
- [ ] `internal/transform/trans_single_path.go:124` does not write back the merged last
      segment distance, and `trans_double_path.go:177` writes it to the wrong index.
- [ ] `internal/conv/curve.go` `ConvCurve.Vertex` updates `lastX`/`lastY` only for vertex
      commands; C++ `conv_curve::vertex` sets `m_last_x`/`m_last_y` after every command
      (found 2026-10-04 during the lint cleanup).
- [ ] `internal/basics/types.go:148` `IsEqualEps` uses an absolute difference. Port AGG's
      frexp/ldexp relative comparison (used by `IsIdentity` / `IsEqual`).
- [ ] `internal/path/path_base.go:136` `ArcTo` emits a duplicate start vertex. Route it
      through `JoinPath` as C++ does.
- [ ] `internal/gpc/gpc.go:30` `Epsilon = 1e-15`; C uses `DBL_EPSILON`.

Rasterizer / scanline

- [x] ✔ `internal/rasterizer/scanline_aa.go:118` — the gamma table truncates (`uint8(val)`);
      C++ uses `uround`. This affects every `SetGamma` user (gamma_ctrl, idea, sbool, Agg2D).
      The test only checks indices 0 and 255.
      → 2026-10-04: Done: `uround`, full 256-entry C++ LUT test. `gouraud` went from 352 px to 0 px.
- [ ] The default cell block limit is 256 (`scanline_aa.go:51`, `scanline_aa_nogamma.go:38`);
      C++ uses 1024. Expose the limit as a constructor option like C++.
      → 2026-10-04: Default is now 1024. Still open: exposing the limit as a constructor option.
- [x] `ClosePolygon` also closes from the move-to state, and `Edge`/`EdgeD` don't set the
      status to move-to (C++ `agg_rasterizer_scanline_aa.h:331,409,422`). The gamma and
      nogamma copies disagree here.
      → 2026-10-04: Done for both rasterizers, with C++ span/bounds parity tests.
- [ ] Integer `MoveTo`/`LineTo` units: `Downscale(x*256)` makes `DblConv` treat ints as
      pixels (`clip.go:46,135`); C++ `ras_conv_dbl` treats them as 1/256 subpixels.
- [x] `SortCells` doesn't reset the current cell after flushing (`cells_aa_simple.go:263`
      vs `agg_rasterizer_cells_aa.h:632`).
      → 2026-10-04: Done for the simple and styled cells, with C++ parity tests.
- [ ] The boolean shapes never call `sl.reset(ir.x1, ir.x2)` / `reset(ur...)` and skip the
      invalid-box early return (`boolean_algebra.go:982,1044`). `XorFormulaSaddle` adds
      early returns (C++ gives 3 for (0,0)), and `boolean_algebra_test.go:181` locks in
      the Go behaviour.
- [ ] `scanline_boolean2` (1033 vs 1031 spans): the AND helpers match C++, so first dump
      and diff the input storages. Likely cause: the demo's contour round-trip
      (`internal/demo/scanlineboolean2/draw.go:1004-1049`) instead of feeding
      `conv_transform` / `conv_stroke` straight into `add_path`.
- [ ] Port `scanline_u8_am` / `scanline32_u8_am` (missing).

Pixfmt / color / blenders

- [x] ✔ `internal/color/rgba16.go:36` `RGBA16Prelerp` is a lerp. C++ uses
      `p + q - multiply(p, a)`. It is used by the 16-bit blenders and `rgb16`;
      `rgba16_test.go:47` asserts the bug. Also port the shift-based lerp and
      `mult_cover(a, cover<<8|cover)`.
      → 2026-10-04: Done, together with an audit of all 8/16-bit fixed-point helpers (lerp, demultiply, from_double, gray gradient, luminance) and C++ golden hashes. RGB48 got `copyOrBlendPix`.
- [ ] Comp-op coverage semantics. C++ blends clear/src/src_in/dst_in/src_out/dst_out/dst_atop
      as `d·(1−cover) + op·cover` (`agg_pixfmt_rgba.h:297-560`); Go only scales source
      alpha (`blender/rgba_composite.go:587-675`). Also port `clip()` and the `plus`
      formula, the integer source premultiply in the adaptor (`:1256`), and the
      `clip_to_dst` adaptors.
      Measured 2026-10-04: the straight-alpha bridge is within ±2 of AGG in premultiplied
      space (`comp_plain_oracle_test.go`); the residual is this integer source premultiply.
- [ ] `pixfmt_rgba8.go:114,409,422` take an opaque+full-cover copy shortcut for *every*
      blender, so comp-op blenders wrapped in `PixFmtAlphaBlendRGBA` (the compositing demos)
      skip the operator. Restrict the shortcut to `RGBAFastBlender`.
- [ ] Unify the 6 demultiply/premultiply variants into one AGG-exact `multiplier_rgba`
      (rounding + clamp, like `simd/cpu.go:408`). Make `SetPlain`/`GetPlain` mean exactly
      `conv_rgba_pre` / `conv_rgba_plain` at every bit depth.
- [ ] `PixFmtAlphaBlendGray16.BlendColorHspan/Vspan` pass 8-bit covers as `Int16u(cover)`,
      but the gray16 blender treats cover as 16-bit (255 ≈ 0.4 % coverage). Expand them with `*257`.
- [ ] Packed RGB 555/565 blend math (`blender/rgb_packed.go:72,185`) must use C++'s shift form.
- [ ] `BlendFrom` ignores the source order (`pixfmt_rgba8.go:976`).
- [ ] The float blender formulas use `(1−a)p+aq` / `(1−a)p+q` in C++
      (`agg_color_rgba.h:1197`); Go uses `p+(q−p)a` (`blender/rgba128.go:40`).
- [ ] sRGB: `ConvertFromRGBA[SRGB]` ignores the colour space (`color/rgba8.go:48`), and
      `color/conversion.go:66` uses pow+round with a wrong breakpoint instead of AGG's
      table search.
- [ ] The "Pre" blenders return early when a==0 and drop additive colour
      (`rgb8.go:118`, `gray8.go:57`, `rgba16.go:161`).

Renderer / span / image / blur

- [ ] `internal/image/image_accessors.go:78` — `ImageAccessorClip.Span` returns the
      background for partly-outside spans; C++ returns `pixel()`.
- [ ] `renderer/base.go:300,317` — `CopyBar`/`BlendBar` drop 1-px bars because
      `IntersectRectangles` uses a strict `<`. Use an inclusive clip like `RectI.Clip`.
- [ ] `span/gradient_lut.go:212-233` — `BuildLUT` invents a last-stop special case
      (C++ gives lut[255]=253; Go gives 255), and `gradient_lut_test.go:145` locks it in.
      Port it verbatim, use the generic interpolator for sRGB, and delete the local copy in
      `examples/.../gradient_focal/main.go:137`.
- [ ] The RGBA image filters ignore `order_type` on output
      (`span_image_filter_rgba.go:164,288,548,682`), and the bilinear_clip fast path adds
      clamps C++ lacks (`:376-411`, contradicting `AGG_DELTAS.md:78`).
- [ ] `effects/stack_blur_optimized.go:275,281`: exported `StackBlurRGB24`/`StackBlurRGBA32`
      are silent no-ops. Recursive blur is horizontal-only (`effects/blur.go:625`).
- [ ] Zero-value-means-default sentinels diverge from C++: `NewKaiserFilter(0)`,
      `NewMitchellFilter(0,0)`, `subpixelShift==0`.
- [ ] Smaller items:
  - outline `ClipBox` should use `line_coord_sat::conv` (`outline_aa.go:81`, `outline_image.go:1011`)
  - the primitives ellipse needs do-while semantics for `ry=0` (`primitives.go:100`)
  - `Dda2LineInterpolator.Init` must use the raw count (`interpolator_linear.go:55`)
- [ ] `image_filters2`: the §1.2 description is stale — the example already uses
      `SpanImageFilterRGBA`. The real gap is that `internal/demo/imagefilters2/draw.go:130`
      leaves `imageFilter=Bilinear`, so `SetImageFilterLUT` is ignored
      (`internal/agg2d/image.go:126`), and it renders premultiplied where C++ uses a plain
      bgra32. Re-measure.

Agg2D

- [x] ✔ Master alpha is applied twice to solid fill/stroke: the colour alpha
      (`rendering.go:196,217`) *and* the rasterizer gamma. C++ applies it only via
      `Agg2DRasterizerGamma` (`agg2d.cpp:1747`). Same in `rendering_float.go:46`.
      `color_blending_test.go:153` is too weak to notice.
      → 2026-10-04: Done in both twins, including float Gouraud and raster text; `cpp_oracle_test.go` checks 14 scenes byte-exact against agg2d.cpp.
- [x] ✔ The AA gamma uses `pow(x, 1/g)` (`rendering.go:425`); C++ `gamma_power` is
      `pow(x, g)`. Also remove the gamma [0.1,3] and master-alpha clamps
      (`rendering.go:600-622`), and make `SetGamma` round.
      → 2026-10-04: Done: `alpha·pow(x,g)` in a shared helper, clamps removed, installed only from SetMasterAlpha/SetAntiAliasGamma/Attach.
- [x] ✔ `ApproxScale = 1.0` (`constants.go:20`, whose comment falsely claims parity);
      C++ `g_approxScale = 2.0`. Update the scale only where C++ does
      (scale/affine/parallelogram/viewport/transformations), not on every `DrawPath`
      (`paths.go:318`), which also overrides the user's `ApproximationScale()`.
      → 2026-10-04: Done: 2.0, updated only where agg2d.cpp updates it; scalar conversion uses `0.7071068`.
- [ ] `Arc` with a negative sweep draws the complementary arc (`shapes.go:118`, `shapes_float.go:108`).
      Arc/Ellipse/AddEllipse must emit `bezier_arc` curves flattened through `conv_curve`
      (`agg2d.cpp:820-839`), not fixed-scale polygons.
- [ ] `RoundedRect*` lacks `normalize_radius()` (`shapes.go:58-91`).
- [ ] `ParallelogramFromRect` maps the wrong corner (`transform.go:454`); it should match
      AGG `rect_to_parl` (`agg2d.cpp:388`). Fix `agg_public_test.go:148` and
      `agg2d_parity_example_test.go:44`, which lock in the bug.
- [x] `DrawPath` lacks the C++ guards: skip the fill if `fillColor.a == 0`, and skip the
      stroke if `lineColor.a == 0 || lineWidth <= 0` (`agg2d.cpp:1373-1407`).
      → 2026-10-04: Done in both twins.
- [x] `Blackman144` is aliased to `ImageFilterBlackman` (`constants.go:85`); C++ uses
      `image_filter_blackman144` (radius 6).
      → 2026-10-04: Done: own enum value plus one shared C++ filter switch for both twins.
- [ ] Strokes always use non-zero winding (`rendering.go:104`); C++ shares the even-odd flag.
- [ ] The no-dash stroke path rebuilds `ConvStroke` each call and loses the miter/inner
      miter/approx scale/shorten settings (`rendering.go:111`).
      → 2026-10-04: Partly done: the miter, inner miter, inner join, approximation scale and shorten settings are now copied. The converter is still rebuilt on each call.
- [ ] Text:
  - `Font()` never clears `gsvFontMode` (`text.go:95`)
  - raster glyphs ignore fill gradients and wrongly apply master alpha (`text.go:837-858`)
  - shaped-raster alignment uses ink bounds instead of advance width + 'H' height
- [ ] Context:
  - `DrawThickLine` restores a never-updated `lineWidth` (`context.go:131`)
  - `SetGlobalAlpha` cancels gradients (`blending.go:68`)
  - `SetItalic(false)` doesn't undo the skew
- [ ] The `ViewportOption`/`BlendMode` numeric values differ from upstream despite
      `agg.go:96` claiming compatibility.
- [ ] `Agg2D::Color` is `srgba8` in C++: colours are converted sRGB→linear on input, while Go
      treats `Color` as linear. The oracle scenes therefore use only 0/255 channels. Decide on
      parity vs compatibility (see AGG_DELTAS).
- [ ] Also from the 2026-10-04 Agg2D pass: C++ `blendImage(img, x, y, alpha)` blends twice
      (agg2d.cpp ~1795). Check Go and record it as a delta.
- [ ] Image blend mode/colour: C++ effectively ignores both (`agg2d.cpp:1615`), 8-bit Go
      applies both, float Go applies only alpha. Pick one, document it, and fix the twin drift.

Fonts / controls / platform

- [ ] Re-port `decompose_ft_outline` line by line (`internal/font/freetype/engine.go:1164-1328`):
  - off-curve first point (`point--` / `limit--`)
  - no extra closing curve when a conic run ends on the last point
  - `goto Close` instead of an unconditional `ClosePolygon`

  The same bug is in `freetype2/engine.go:238`. Test against C++ vertex dumps of glyphs
  whose contours start off-curve.
- [ ] Agg2D raster text must use `glyph_ren_agg_gray8` (AGG rasterizer), not
      FreeType's `FT_Render_Glyph`. Adopt C++'s five-value `glyph_rendering` enum.
- [ ] Outline glyphs: serialize them into the cache as int 26.6 (`dbl_to_int26p6`)
      instead of re-decomposing on every hit (`engine.go:929`, `cache_manager.go:268`).
      Add `change_stamp`, `synchronize`, and prev/last-glyph kerning.
- [ ] Controls:
  - slider `OnMouseButtonUp` (`slider_ctrl.go:206`)
  - polygon ctrl arrow keys, `OnMouseMove` flag, `<` vs `<=`, crossings test
    (`polygon_ctrl.go:186-421`)
  - all-zero colours from the generic gamma/spline ctrl constructors
  - `GSVText.TextWidth` must leave the cursor moved
- [ ] Runner: mouse Y must be `height - y`, not `h-1-y` (`lowlevelrunner/runner_platform.go:102`),
      and the event loop needs a wait mode instead of a 100%-CPU poll.

### 8.3 P1 — Parity verification infrastructure

- [ ] Demo parity must render **live**. `tests/visual/demo_parity_test.go` currently compares
      gitignored, possibly stale `reference/go/examples/*.png` and skips on fresh clones.
      Render each demo in-process, as `examples/.../trans_curve/main_test.go` does.
- [ ] Per-demo thresholds: exact (0 px) for every RMSE-0 row in §1.2, and an explicit budget
      for the rest. Replace the global `Tolerance:10 / 1%`. Delete the tautological
      `TestDemoParityThresholdIsTight` (`:385`).
- [ ] Reproducible C++ references. Commit the `agg-2.6` patch set (`--screenshot` flag,
      `demo_timing.h`, example edits) as `tests/visual/reference/cpp/patches/`, and record the
      upstream SHA plus build/capture steps. Add an explicit approval step for
      `UPDATE_VISUAL` / `GENERATE_REFERENCES` (closes §1.1).
- [ ] A **C++ oracle dump tool**: a small C++ program built against `../agg-2.6`, plus
      Go table tests. It should cover:
  - cells/spans for fill rules, gamma, clipping, compound and all sbool ops
  - vertex streams for `math_stroke` joins, `vcgen_stroke` / `dash` / `contour`,
    `curve3/4_div`, `smooth_poly1`, `trans_*_path`
  - golden vectors for rgba8/rgba16 multiply/lerp/prelerp/demultiply (exhaustive 256²)
    and every comp-op × cover
- [ ] An **`agg2d.cpp` golden harness**: about 40 targeted scenes (master alpha, gamma,
      scaled ellipse, negative arc, oversized rounded-rect radii, parallelogram, every image
      filter / resample mode, text alignment, `NoFill`+`BlendSrc`), diffed byte-for-byte.
      The current `engine` CPP backend is a hand-written shim, not `agg2d.cpp`, so it is not
      an Agg2D oracle. Consider backing it with the real `agg2d.cpp`.
- [ ] Replace bug-enshrining tests:
  - `color/rgba16_test.go:47`
  - `span/gradient_lut_test.go:145`
  - `scanline/boolean_algebra_test.go:181`
  - `agg_public_test.go:148`
  - `agg2d_parity_example_test.go:44`
- [ ] Turn assertion-free "should not panic" / vertex-count tests in `vcgen`, `conv` and
      `span` into golden comparisons, and remove the stale skip in `vcgen/contour_test.go:112`.
- [ ] Add tests for the untested packages: `internal/blur`, `internal/gamma`,
      `internal/renderer/primitives`, `internal/font` (default build), `internal/ctrl/text`.
- [ ] Fix the C++ backend's portability: hardcoded `-I/usr/include/agg2 -L/usr/lib/x86_64-linux-gnu`
      in `engine/cpp_native_aggreal_flags.go` should become pkg-config / `CGO_*`. Then gate a
      CI job on `-tags "agogo aggreal"` (closes §5.9).

### 8.4 P2 — Simplification (ordered by value ÷ effort, never at the cost of parity)

1. [ ] Delete dead code:
   - **Packages:** `internal/config`, `internal/vertex_source`. Fold `internal/vpgen`: move
     `vcgen/{segmentator,clip_polygon,clip_polyline}.go` into `vpgen` (where AGG keeps them)
     and delete the duplicates.
   - **Unused helpers:** `internal/array/interfaces.go`, most of `array/comparators.go`
     (use `cmp`/`slices`), and the unused `simd.Comp*HspanRGBA` + `comp_src_over_sse41_amd64.s`.
   - **Effects/examples/cmd leftovers:** the `effects` blur stubs and unused image
     interfaces, `examples/shared/renderutil`, `cmd/checkconv`, `cmd/lion_bounds`.
   - **Transform extras:** `transform/viewport_manager.go`, `viewport_integration.go`.
   - **Decide on:** `internal/color/conv` (920 LOC, zero importers) — wire it in or mark it parity-only.
2. [ ] Move test doubles out of the library:
   - `renderer/scanline/test_mocks.go`, `platform/backend_mock.go`, the `scanline.Mock*`
     helpers and `image.MockPixelFormat` → `_test.go` or a `testutil` package
   - rename `scanline/scanline_hit_test.go`: production code is hidden in a test-named file
3. [ ] Remove `reflect.ValueOf(...).IsZero()` from the span hot path (`span/converter.go:62-81`).
4. [ ] One gamma/sRGB implementation. Today it exists in `internal/gamma`, `internal/pixfmt/gamma`
      (plus its `init()`), and `color/conversion.go`. Keep one LUT behind a lazy `sync.Once`.
5. [ ] **One vertex-source protocol** (the biggest structural win; today there are 4
      protocols with 145 adapters). Decision: `Rewind(pathID uint)` +
      `Vertex() (x, y float64, cmd basics.PathCommand)`. This is idiomatic, and the
      pointer-out form makes x/y escape to the heap through interface calls. Define it
      once and alias it in `conv`, `rasterizer`, `path`, `shapes`, `bezierarc`, `ctrl`,
      `transform`, then delete the adapters. Record the signature delta in `AGG_DELTAS.md`.
6. [ ] Canonical contracts in one package (aliased everywhere else): PixelFormat (5 copies),
      BaseRenderer (6), Scanline (10, incl. the deprecated `rasterizer.ScanlineInterface`),
      Rasterizer (7). The scanline interface takes C++ signatures plus `Reset`, so
      rasterizers/storages plug straight into `sbool_*` without the ~5 demo adapters.
      One span-generator contract: `Generate(colors []C, x, y, n int)` + `Prepare()` (3
      signatures today).
7. [ ] Instantiate `RendererBase` with **concrete** pixfmt types in agg2d instead of the
      interface `renderer.PixelFormat[C]` (31 sites of generics over an interface:
      dictionary *and* dynamic dispatch). Drop the phantom `Clip any` parameter on
      `RasterizerScanlineAA`.
8. [ ] Collapse the remaining duplicates:
   - small types: one `VertexDist`, generic `VertexSequence[T]`, one dda/dda2 (4 copies today)
   - one blur package mirroring `agg_blur.h` with a full recursive blur (`blur` currently imports `effects`)
   - drop `color.ColorOrder` in favour of `internal/order`
   - duplicated code paths: cells simple/styled sharing one core, gamma/nogamma
     rasterizers sharing one core, 16/32-bit scanlines generic over the cover type
9. [ ] Pixfmt: map the 9 hand-written RGBA/RGB/gray structs back onto C++'s 3 templates where
      a benchmark confirms no regression. Collapse packed 555/565 into one `PixFmtRGBPacked[B]`.
      Prune the blender alias sprawl.
10. [ ] Agg2D float twin (~4.8k LOC across 19 `*_float.go` + root wrappers): one generic core
       over colour/pixfmt, instantiated twice (C++ does this with one typedef). It has already
       drifted: image blend mode, `Context` line width, `BlendImage` renderer, missing rect overloads.
11. [ ] Demos:
   - one implementation per demo under the existing repository-level `internal/demo/<name>`
     behind a single Demo interface. Do not move them to `examples/internal/`: Go's
     internal-package rule would then block `cmd/wasm` from importing them
   - `examples/*/main.go` and `cmd/wasm` become thin adapters with a registry (34 wasm demos
     are re-implementations today; `cmd/wasm/main.go` is a 1945-line if-chain)
   - merge `demorunner` and `lowlevelrunner`
12. [ ] One FreeType engine: keep the `freetype2` design (correct enum, agg_gray8 path) after
       fixing decomposition, then delete the other.

### 8.5 P2 — Public API

- [ ] Typed enums instead of `= int` aliases (`agg.go:98-104`, `blending.go:8`, `transforms.go:14`).
- [ ] Prune `Context`: 17 `SetBlend*` aliases, `SetCompositeOperation`, `SetStrokeWidth`,
      six dash presets, the no-op `SetBold` / `SetUnderline` / `DrawTextOnPath` stubs.
      Stop `Context` reaching into `ctx.agg2d.impl`.
- [ ] Public `FilterMitchell` / `FilterSinc` / `FilterLanczos` (`agg.go:501-505`, `Blackman + 2/5/6`)
      select Kaiser / Mitchell / Sinc, because the internal enum values are 11/14/15/16.
      `surface_api_test.go` locks the wrong values. Fix both, and note it in the changelog.
- [ ] Record behaviour changes from the 2026-10-04 parity pass in a changelog:
  - `AntiAliasGamma` is now `pow(x, g)` (it was `pow(x, 1/g)`), and it and master alpha are no longer clamped
  - `Blackman144` now uses radius 6
  - `DrawPath` no longer resets the approximation scale
  - 16-bit lerp/prelerp now match C++
- [ ] Fix the misleading `ResampleNearest/Bilinear/Bicubic` aliases (`agg.go:117`). Fix or
      remove the stale `Version` / `BuildDate`.
- [ ] Move the escape hatches (`GetInternalRasterizer`, `ScanlineRender`, `Agg2DFloat.GetImpl`)
      into an `advanced` sub-package, or drop them. Stop exposing `internal/color` types
      in exported signatures.
- [ ] Keep the non-AGG extras (Photoshop blend modes, N-stop gradients, transform stack,
      quad transforms) out of the parity core in `internal/agg2d`, or mark them clearly.
- [ ] `engine/`: one `agogo` tag (drop the always-rejected stub tier), pkg-config flags.

### 8.6 P3 — Performance (benchmark before and after; keep output byte-identical)

- [ ] `scanline/storage_aa.go:154` `getBlockSlice` copies to the end of the storage on every
      `Get` (O(N²)). Its `Render`/`SweepScanline` do byte-wise `unsafe` copies, and the 16/32-bit
      storages silently store zeros. Return sub-slices and use `copy`.
- [ ] Cells use 64-bit `int` (32/40 bytes vs C++ 16/20) → `int32`. The bounding box is
      updated per cell instead of per line.
- [ ] Remove per-join/per-vertex allocations in `vcgen/stroke.go:139-198` and
      `vcgen/dash.go:178-229`, plus the 8-float alloc in `bezierarc`. Also remove the
      per-hline `make` in `outline_aa.go:113,186` and the per-line interpolator heap
      allocations, the per-span `simpleIterator` in `boolean_algebra.go`, the
      `sort.Slice` per scanline in the compound rasterizer, and the per-span alloc in the
      Agg2D Gouraud renderer.
- [ ] Hoist the per-pixel type assertions (`span_image_filter_rgba.go:902`,
      `pixfmt_rgba8.go:1374`, `cell_style_aa.go:31`). Drop the zero-fill in
      `SpanAllocator.Allocate`. Store outline-image patterns as rgba8, not float `color.RGBA`.
- [ ] Unsafe: the `*[1024]byte` casts in `array/pod_arrays.go:532,551` and
      `pod_bvector.go:308-360` panic for elements over 1 KiB. Use `unsafe.Slice`.
      Un-export `NextPixPtr` / `PixPtrOffset`.

### 8.7 P2 — Tooling, docs, hygiene

- [ ] `AGG_DELTAS.md` "Composite blend modes" still says the 8-bit composite pixfmt treats
      the buffer as premultiplied; since 4e431e6 it is straight (`CompositeBlenderPlain`).
      Re-check the float twin and fix the text.
- [ ] Justfile:
  - fix or delete the broken recipes: `serve-web` (`go run -e`), `profile-mem/cpu`,
    `run-examples-*`, `run-tests`, `stats` (the `$$` escapes), `docs` (writes to the
    missing `docs/api/`), `update-tasks`, `build-windows`, `list-examples`
  - `lint` must not use `--new`
  - `check` must not run the mutating `fmt`/`tidy`
- [ ] Untrack the committed artifacts: `examples/platform/sdl2/sdl2` (ELF),
      `examples/core/basic/shapes/ellipse_test.ppm`, `.claude/agents/*`.
- [ ] Remove the duplicate `tests/visual/reference/bootstrap/go-golden/primitives`, the
      orphaned `reference/wasm/*.png` (5.3 MB, read by no test), and the duplicate `spheres.ppm`.
- [ ] Docs:
  - remove references to the missing `docs/TASKS.md`, `TASKS-COMPLETED.md`, `REVIEW_STATUS.md`,
    `types.go` and `tests/unit` (README, AGENTS.md, `.github/copilot-instructions.md`,
    `docs/concepts/*`)
  - fix the file names in `AGG_DELTAS.md:63,72` and `translation-guide.md:1048`
  - update the stale "Phase 8.2/8.3" references in `tests/visual/reference/{bootstrap,cpp}/README.md`
- [ ] Record the silent C++-bug fixes in `docs/AGG_DELTAS.md` (or revert them to match C++):
  - `renderer_primitives::visible` `x+y`
  - `bezier_arc_svg` `ry=-rx`
  - the cells line-split `return`
  - `src_atop` blue channel
  - the Agg2D line-gradient matrix
  - `blendImage` double blend
- [ ] Add a LICENSE file. Consider moving the SDL2 demos into a nested module so `go-sdl2`
      leaves the library's `go.mod`.
- [ ] Trim this file. About 43% of PLAN.md is DONE narratives (Phases 2–6, the `[x]` rows in
      §1.2). Move them to `docs/AGG_DELTAS.md` or a changelog, and keep one line per open item.

---

## Phase 9 - Exit Checklist

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
- [ ] Phase 7's resample exit criteria are met, or the phase is explicitly
      deferred with rationale (it is a performance phase, not a fidelity gap).
- [ ] Phase 8 P0/P1 items are closed; P2/P3 items are explicitly scheduled or deferred.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement or fix the Go behavior.
3. Add or update contract tests.
4. Add or update visual regression tests if the behavior is rendering-visible.
5. Update this plan.
