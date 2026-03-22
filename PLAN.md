# AGG Go Port - Fidelity-First Plan

## Objective

Port AGG 2.6 to Go so that:

1. Rendering behavior stays as close as possible to original AGG (`../agg-2.6/agg-src`).
2. Go code remains idiomatic, maintainable, and testable.
3. Deviations from AGG are explicit, justified, and tested.

This is the single authoritative project plan. All task tracking, port inventory, and
documentation gaps live here. `docs/AGG_DELTAS.md` records intentional deviations.

## Non-Negotiables

- [ ] Every major behavior maps to a C++ source reference (file + method).
- [ ] No placeholder rendering paths in production-critical pipeline stages.
- [ ] Public API remains stable and idiomatic; internal architecture may change.

## Porting Rules

1. Fidelity first for algorithms and numeric behavior.
2. Idiomatic Go for ownership, naming, package boundaries, and tests.
3. No silent fallbacks that change rendering semantics.
4. If behavior differs from AGG, document as an intentional delta in `docs/AGG_DELTAS.md`.

---

## Phase 0 - Baseline and Traceability ✅

Project-wide tracking and auditability are in place:

- **Parity ledger**: One row per `Agg2D` method with C++ source reference, Go method mapping,
  status (`exact`/`close`/`placeholder`/`missing`), test reference, and notes. Key anchors are
  in `../agg-2.6/agg-src/agg2d/agg2d.h` and `../agg-2.6/agg-src/agg2d/agg2d.cpp`.
- **Placeholder inventory**: All simplified/placeholder paths in rendering-critical packages
  (`internal/agg2d`, `internal/rasterizer`, `internal/scanline`, `internal/renderer`,
  `internal/span`) are recorded and prioritized (`must-fix` / `acceptable temporary` /
  `low-priority`).

---

## Phase 1 - AGG2D Behavioral Parity

Primary target: `internal/agg2d/*` against `agg2d.cpp`.

Core `Agg2D` rendering behavior is aligned:

- **Image pipeline**: `renderImage*` uses the AGG-style scanline/span pipeline (interpolator
  sampling, filter LUT integration, resample-mode behavior, blend-color conversion) with correct
  transform flow and no nearest-neighbor-only fallback for transformed images.
- **Gradients**: Linear/radial matrix construction, transform/scalar conversion, and distance
  (`d1/d2`) handling match AGG ordering.
- **Text**: Glyph rendering runs through the real rasterizer/scanline pipeline (no rectangle
  fallback) and matches the vector vs raster cache contract. Kerning in `TextWidth`/`Text` uses
  glyph indices.
- **Clipping & state**: `ClipBox` propagation and clip-sensitive image ops match AGG semantics,
  covered by pixel-asserting tests. Attach-time and state-update behavior (fill rule, gamma,
  master alpha) is aligned.

Remaining:

- [ ] Visual tests for AGG2D demos still need to pass against reference thresholds.

---

## Phase 2 - Core Pipeline Parity ✅

### ⚠️ URGENT: PixFmtAMaskAdaptor hardcodes RGBA8 — RGB pixfmts cannot be wrapped

**Gap**: `PixFmtInterface` (in `internal/pixfmt/pixfmt_amask_adaptor.go`) requires `RGBA8[Linear]`
for every method. This means only RGBA pixfmts can be wrapped by `PixFmtAMaskAdaptor`. In C++,
`pixfmt_amask_adaptor<PixFmt, Mask>` is a template and wraps _any_ pixfmt regardless of color type
— including `pixfmt_bgr24`, `pixfmt_rgb24`, `pixfmt_gray8`, etc.

**Concrete impact**:

- `alpha_mask2.cpp` uses `pixfmt_amask_adaptor<pixfmt_bgr24, amask_no_clip_gray8>`.
- The Go port cannot replicate this: `PixFmtBGR24` has color type `RGB8[Linear]` and does not
  satisfy `PixFmtInterface`. The example works around this by using `PixFmtRGBA32[Linear]` instead
  of BGR24 — a silent behavioral deviation.
- For an opaque white background the RGB blending formulas are numerically equivalent (RGBA lerp ==
  RGB lerp when destination alpha is 255), so the deviation does not affect the alpha_mask2 output
  directly. However, it prevents faithful porting of any demo that uses a non-RGBA main buffer with
  an alpha mask, and masks any future bug in the RGBA path that the BGR24 path would not share.

**Fix required**:
Make `PixFmtAMaskAdaptor` generic over the wrapped pixfmt's color type, or introduce a second
adaptor variant (`PixFmtAMaskAdaptorRGB`) for RGB pixfmts. The internal `PixFmtInterface` should
either become generic (`PixFmtInterface[C]`) or the amask adaptor should bypass it and delegate
directly to a narrower "blend span" interface that both RGB and RGBA pixfmts satisfy.

**References**:

- C++: `agg_pixfmt_amask_adaptor.h`, `pixfmt_amask_adaptor<PixFmt, AlphaMask>`
- Go: `internal/pixfmt/pixfmt_amask_adaptor.go`, `PixFmtInterface`
- Go: `internal/pixfmt/pixfmt_rgb8.go`, `PixFmtBGR24` / `PixFmtAlphaBlendRGB`

**Tasks**:

- [ ] Redesign `PixFmtInterface` / `PixFmtAMaskAdaptor` to support RGB (and Gray) pixfmts.
- [ ] Port `alpha_mask2` main buffer to `PixFmtBGR24` once the adaptor supports it.
- [ ] Add a test that wraps `PixFmtBGR24` with the amask adaptor and asserts correct blending.

---

Rasterizer → scanline → renderer → pixfmt behavior is aligned with AGG:

- **Rasterizer/scanline**: Fill rules, clipping edge cases, cell accumulation, sweep indexing,
  and duplicate-cell behavior match AGG expectations.
- **Renderer/pixfmt**: Copy/blend overlap and premultiplied vs straight-alpha behavior are aligned.
  The needed `copy_from` / `blend_from` helpers are ported across RGBA/RGB/Gray plus
  transposer/amask/composite pixfmts, with expanded Porter-Duff/composite coverage.
- **Converters**: Stroke/dash/transform ordering, line cap/join enum parity, and
  viewport/gradient/scalar propagation are aligned; key converter/vcgen/vpgen state machines are
  audited beyond just Agg2D call sites.

---

## Phase 3 - Font Subsystem Consolidation and Type Safety ✅

One coherent font stack with a tighter, type-safe surface:

- **Architecture**: A single authoritative font/cache architecture is used by Agg2D.
  `internal/font/freetype2` is reduced to minimal, justified wrappers (with documented Go-only
  deltas where applicable).
- **Type safety**: Broad runtime `interface{}` usage in text-critical paths is replaced with
  explicit interfaces; build-tag boundaries remain the only intentional runtime dispatch.
- **Lifecycle**: FreeType2 face/engine lifetime behavior is rechecked; multi-face close releases
  tracked faces correctly. The `maxFaces` cap is documented as an intentional Go-only policy delta.

---

## Phase 4 - Remaining Port Inventory ✅

All port inventory items are complete, explicitly deferred as out-of-C++-AGG-2.6-scope,
or documented as intentional Go deltas. `go test ./internal/...` passes (44 packages).

**4.1 Transformations**: All 7 types at C++ parity (TransAffine, TransPerspective, TransBilinear,
TransSinglePath, TransDoublePath, TransWarpMagnifier, TransViewport). `PerspectiveIteratorX` avoids
per-pixel divide; determinant check in `Invert()` for stability; `ViewportManager` handles
multi-viewport; `ConvTransform.Transformer()` getter added, duplicate setter removed.
`SpanInterpolatorPerspectiveExact` and `SpanInterpolatorPerspectiveLerp` verified against
`agg_span_interpolator_persp.h` at full parity. Out of scope: TransPolar (example-only in C++),
WarpMagnifier multiple zones (not in C++ AGG 2.6).

**4.2 Converters and generators**: `ConvAdaptorVPGen` + all vpgen components verified. Stroke/contour
pipeline complete: `conv_stroke`, `vcgen_stroke` (InnerJoin, all cap/join types), `vcgen_contour`,
`conv_contour` all at C++ parity. Dash, smooth-poly, and all smaller path-utility converters
complete with tests. Rasterizer cell-run compaction regression tests added;
`RenderAllPaths` typed via `MultiPathRasterizerInterface`.

**4.3 Span / image-processing**: `GradientContour.Calculate()` formula fixed to match C++
(`buffer*(d2/256)+d1`, not linear lerp). Bilinear filter spurious premultiplied-alpha clamping
removed (not in C++ AGG); clip variant implements all three C++ boundary cases. Gouraud shading
complete: `SpanGouraudGray` and `SpanGouraudRGBA` at C++ parity. `BoundingRect`,
`ConvShortenPath`, `VCGenVertexSequence` all implemented and tested.

**4.4 Fonts and utilities**: `RowPtr` bridge resolved (direct slice for plain, on-demand cache for
pre). GSV embedded font complete (`GSVText`, `GSVTextOutline`, `font_data.go`). FreeType custom
memory hook unsupported by design (`FT_Init_FreeType` used; `ftMemory` ignored with
`_ = ftMemory`). Color conversion: all three C++ headers ported (`agg_color_conv.h`,
`agg_color_conv_rgb8.h`, `agg_color_conv_rgb16.h`) to `internal/color/conv/`.

**4.5 Generics / pixfmt**: RGBA16 pixfmt refactored to blender-interface pattern; tests restored.
`any()` assertions eliminated from Gray16/Gray32 pixfmts; `RawRGB/RGBAOrder` fast paths retained
as legitimate. `VertexFilter` shims removed from `array/vertex_sequence.go` and
`basics/math_stroke.go`. Color space kept at Linear + SRGB only.

**Note**: `TestTransformImageUsesPremultipliedRenderer` updated to use premultiplied source input —
C++ AGG routes image rendering through `m_renBasePre` (`agg2d.cpp:1738`) which expects
premultiplied values; no automatic straight→premultiplied conversion occurs in the span path.

---

## Phase 5 - API and Documentation Finalization ✅

- [x] `Context` / `Agg2D` separation documented; package doc in `agg.go` corrected.
- [x] Architecture overview updated to 35 internal packages.
- [x] `docs/AGG_DELTAS.md` created — documents all intentional deviations from C++ AGG 2.6.

---

## Phase 6 - SIMD Infrastructure and Bulk Pixel Paths ✅

`internal/simd/` package with runtime CPU detection, build-tagged arch dispatch, and `purego`
scalar baseline. Four pixel operations each have generic, amd64 (SSE2/SSE4.1/AVX2), and arm64
(NEON/generic) paths. Assembly in flat `internal/simd/*.s` files (idiomatic Go layout).

- [x] **FillRGBA** — packed-RGBA bulk fill; wired to `CopyHline` / `Clear`.
- [x] **BlendSolidHspanRGBA** — solid-color AA spans with per-pixel cover (SSE4.1 PMAXUW/PMINUW lerp).
- [x] **BlendHlineRGBA** — uniform-coverage hline blend; alpha==255 routes to FillRGBA.
- [x] **BlendColorHspanRGBA** — per-pixel color+cover (scalar IMULQ alpha, SIMD lerp for 8 channels).
- [x] `pixfmt_rgba8.go` fast paths wired for all four operations; RGBA byte order uses SIMD, others fall back to scalar.
- [x] Table-driven tests verify bit-identical output across all forced implementation paths.
- [x] QEMU arm64 correctness checks in regular workflow (`just test-arm64`).

---

## Phase 7 - SIMD Expansion Targets

Each section follows the same three-tier pattern: generic Go → SSE4.1 → AVX2 on amd64;
generic fallback → NEON on arm64.

### 7.1 Premultiply / Demultiply ✅

- [x] Generic — correct scalar baseline with zero-alpha guard on demultiply.
- [x] SSE4.1 (amd64) — process 4 pixels/iter: PMULLW × α/255 (AGG rounding); PACKUSWB clamp.
- [x] AVX2 (amd64) — delegates to SSE4.1 kernel (bottleneck is memory bandwidth).
- [x] NEON (arm64) — generic fallback (correct and tested via QEMU).
- [x] Wired into `pixfmt_rgba8.go`; SIMD fast path for standard RGBA byte order.
- [x] Table-driven tests: bit-identical output vs. scalar, zero-alpha row, boundary alphas, round-trip.

### 7.2 Composite Blend Modes ✅

- [x] Generic — integer-arithmetic scalar for `SrcOver`, `DstOver`, `SrcIn`, `DstIn`, `SrcOut`,
      `DstOut`, `Xor`, `Clear` in `internal/simd/cpu.go`.
- [x] SSE4.1 (amd64) — `SrcOver` 2 pixels/iter via `compSrcOverHspanRGBASSE41Asm`.
- [x] AVX2 (amd64) — delegates to SSE4.1 kernel.
- [x] NEON (arm64) — generic integer-arithmetic fallback.
- [x] Wired into `pixfmt_composite.go` `BlendHline` and `BlendSolidHspan` fast paths.
- [x] Tests: bit-exact (±1) output vs. float64 reference across all paths.

### 7.3 Gradient and Image Span Generation

Span generators feed pixel data into `BlendColorHspan`; profile before committing to SIMD.

- [x] Profile baseline in `internal/span/` before writing any SIMD code.
      Added length-scaled span benchmarks for `BenchmarkSpanGradientGenerate` and
      `BenchmarkSpanImageFilterRGBAGenerate` on 2026-03-14.
- [x] SSE4.1 (amd64) — linear gradient: PADDD step accumulation + PSHUFB color lookup.
      Profiled and skipped on 2026-03-14. Baseline throughput was already ~180-245 MB/s
      for linear gradients, and the representative `BenchmarkAgg2DSceneGradientClip/800x600`
      run still spent ~24.3 ms/op outside any demonstrated span-generation hotspot.
- [x] AVX2 (amd64) — double-width linear interpolation if SSE4.1 proves worthwhile.
      Skipped on 2026-03-14 because the SSE4.1 path was not justified by profiling.
- [x] NEON (arm64) — `vaddq_s32` step accumulation; skip if not hot.
      Skipped on 2026-03-14 because the generic path is not yet a demonstrated hotspot.
- [x] Image-filter / resampling kernels: SSE4.1 `PMADDUBSW` for bilinear tap accumulation.
      Profiled and skipped on 2026-03-14. Bilinear RGBA generation measured ~155-206 MB/s
      with zero allocations in the focused benchmark, which was not enough evidence to
      justify assembly without a profile showing it dominates scene time.
- [x] Only implement tiers that show measurable gain in profiling.

### 7.4 Alpha-Mask Helpers

- [x] Generic — correct scalar baseline for mask fill and RGB-to-gray conversion.
      Added shared horizontal-span helpers on 2026-03-14 for contiguous one-component
      masks, stepped component extraction, and RGB24-to-gray conversion. Current
      microbenchmarks: `BenchmarkAlphaMaskU8FillHspan` 11.4 ns/op and
      `BenchmarkAlphaMaskU8FillHspanRGBToGray` 227.3 ns/op, both with 0 allocs/op.
- [x] SSE4.1 (amd64) — mask fill: 16 bytes/iter; RGB→gray: `PMADDUBSW` with BT.601 weights.
      Added `internal/simd` SSE4.1 kernels on 2026-03-14 for short one-byte mask copies
      and exact RGB24→gray conversion. Gray conversion uses three `PMADDUBSW` passes to
      preserve the scalar `(77*r + 150*g + 29*b) >> 8` result without saturation.
      Current microbenchmarks: `RGB24ToGrayU8` improved from ~4.0 GB/s generic to
      ~15.4 GB/s SSE4.1 at 1024 pixels; one-byte mask fill uses SSE4.1 for short spans
      and falls back to `copy()` on longer spans where the runtime memmove path is faster.
- [x] AVX2 (amd64) — 32 bytes/iter mask fill; 256-bit RGB→gray.
      Added AVX2 kernels on 2026-03-14 for 32-byte mask copies and 8-pixel RGB24→gray
      conversion using two 128-bit lane-aligned loads per block. Current microbenchmarks:
      `CopyMask1U8` at 256 pixels improved from 22.60 ns/op (SSE4.1) to 21.86 ns/op
      (AVX2), and `RGB24ToGrayU8` at 4096 pixels improved from 1093 ns/op (SSE4.1) to
      673.1 ns/op (AVX2).
- [x] NEON (arm64) — `vst1q_u8` mask fill; `vmull`/`vadd` for RGB→gray.
      Added NEON kernels on 2026-03-14 for one-byte mask copy and 8-pixel RGB24→gray
      conversion using `VTBL` channel extraction plus `VPMULL`/`VADD` accumulation.
      Verified via `just test-arm64` under QEMU for `internal/simd`, plus
      `GOOS=linux GOARCH=arm64 go build ./internal/pixfmt`.
- [x] Wire into alpha-mask call sites in `internal/pixfmt/`.
      `AlphaMaskU8` and `AMaskNoClipU8` horizontal span paths now dispatch through
      the shared helpers instead of per-pixel `RowPtr` lookups.
- [x] Tests: byte-exact mask fill; gray values within ±1 of scalar.
      Added exact-output tests for contiguous one-component fill plus RGB→gray fill
      and combine paths.

### 7.5 Gamma / LUT Application

- [x] Profile gamma application in a representative scene before writing any SIMD code.
      Added focused benchmarks on 2026-03-14 for whole-buffer RGBA gamma application
      and RGBA `BlendFromLUT`, plus a representative `blend_color` demo benchmark for
      the LUT path. Current microbenchmarks: `BenchmarkPixFmtRGBA32ApplyGammaDir`
      measured ~338-341 MB/s, and `BenchmarkPixFmtRGBA32BlendFromLUT` measured
      ~425-489 MB/s with 0 allocs/op. Representative scene:
      `BenchmarkBlendColorLUT/800x600` measured 7536388 ns/op with 680156 B/op and
      1029 allocs/op, which does not support gamma/LUT SIMD as the next clear
      bottleneck.
- [x] SSE4.1 (amd64) — `PSHUFB`-based partial LUT; implement only if profiling justifies.
      Profiled and skipped on 2026-03-14. The measured baseline did not justify a
      partial-table SIMD path over the existing scalar LUT walk.
- [x] AVX2 (amd64) — `VPGATHERDD` gather if available and beneficial; otherwise skip.
      Skipped on 2026-03-14 because profiling did not justify gather-based LUT work.
- [x] NEON (arm64) — `vtbl`/`vqtbl1q_u8` for 16-entry segments; skip if not hot.
      Skipped on 2026-03-14 because the generic path is not yet a demonstrated hotspot.
- [x] If none of the tiers show meaningful gain, mark as "profiled, skipped" and close.

---

## Phase 8 - Test Strategy for Port Fidelity

### 8.1 Contract tests (API behavior)

- [x] Expand AGG2D tests to assert outputs for covered rendering paths.
- [x] Add deterministic checks for transform-image, clipping, blend modes, gradients, text bounds.
- [x] Replace remaining AGG2D smoke/integration tests with output or state assertions.
- [x] Expand contract coverage for weaker packages: `effects`, `platform`, `primitives`,
      `pixfmt/blender`.
- [x] Re-audit internal packages against the current coverage floor.
- [x] Raise coverage for the next priority gaps:
  - `internal/pixfmt` improved to 57.5% with additional constructor/accessor and generic-mask-path tests.
  - `internal/pixfmt/gamma` improved to 77.1% with RGBA gamma/multiplier behavior tests.
  - `internal/color` improved to 76.6% with floating-point RGBA math and conversion-helper tests.
- [x] Re-audit tests that verify mocks or package-private state; prefer observable behavior.

### 8.2 Visual regression tests

- [x] Store references under `tests/visual/reference`.
- [x] Automated diff thresholding and HTML report generation in `tests/visual/framework`.
- [x] Extract and centralize the current Go golden-test screenshots as a bootstrap reference corpus.
      Snapshot stored under `tests/visual/reference/bootstrap/go-golden/` on 2026-03-16
      (56 PNGs copied from `tests/visual/reference/primitives/`).
- [x] Import an initial canonical C++ screenshot corpus from precompiled AGG demos.
      Stored under `tests/visual/reference/cpp/examples/` on 2026-03-16
      (60 PNGs captured from `../agg-2.6/build/examples/` via the upstream X11
      `F2` screenshot path).
- [x] Generate the matching Go-port screenshot corpus for direct demo-level comparison.
      Stored under `tests/visual/reference/go/examples/` on 2026-03-16
      (60 PNGs rendered through the default headless demo runner path).
- [x] Generate canonical references from C++ AGG for core scenarios and replace Go-side
      references where C++ output is the ground truth.
- [ ] Drive the new `tests/visual/demo_parity_test.go` corpus to green.
  - Fix demo frame mismatches against C++ output:
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
    - [x] `gouraud`
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
- [x] Expand C++-generated visual reference set:
  - basic shapes and AA edge cases
  - gradients, text rendering, other parity-critical scenarios
- [x] Expand visual coverage by category (partial — parity-critical areas done):
  - primitives ✓ (`shapes_test.go`, `rectangle_test.go`)
  - path stroke/fill ✓ (`stroke_test.go`)
  - transformations ✓ (rotate, scale, nested-transform cases in `rectangle_test.go`)
  - color/blend-mode ✓ (`blend_test.go`: SrcOver, Multiply, Screen, Overlay, Darken, Lighten, Difference, Xor, Add, global alpha)
  - gradients ✓ (`gradient_test.go`: linear H/V/diagonal/profile, radial centered/off-center/multi-stop, gradient on path, transparency compositing)
  - clipping ✓ (clipped rectangles in `rectangle_test.go`)
  - anti-aliasing quality cases (pending)
  - image operations (pending)
  - advanced and edge-case scenes (pending)
- [ ] Add reference-management workflow: controlled regeneration, approval surface.
      Bootstrap note recorded in `tests/visual/reference/bootstrap/README.md`, but no
      canonical C++ import/regeneration workflow exists yet.

### 8.3 C++ parity checks

- [ ] For each parity row marked `exact`, include at least one source-linked test case.
- [ ] For rows marked `close`, include documented rationale.

### 8.4 Test-suite cleanup ✅

- [x] Removed debug-style integration tests that only log state.
- [x] Converted useful debug coverage into proper contract tests with assertions.
- [x] All packages pass `go test`.

### 8.5 Remaining AGG2D parity rows ✅

- [x] `Attach` / `AttachImage` parity verified; contract tests in `attach_test.go`.
- [x] `TextWidth` parity: kerning, glyph-index lookup, missing glyphs, empty string, idempotency.
- [x] `Text` parity: `GlyphDataMono` documented as Go extension; kerning placement and
      world-to-screen conversion verified; raster glyph placement (gray8/mono) covered.

### 8.6 Optional property tests ✅

- [x] Property-style tests for transformations: translate/rotate round-trip, identity multiply no-op, composition associativity, inverse gives identity — `internal/transform/affine_property_test.go` using `testing/quick`.
- [x] Property-style tests for color math: sRGB↔linear scalar round-trips, monotonicity of both conversion directions, RGBA8 Gradient endpoints, LUT monotonicity, alpha preservation — `internal/color/property_test.go` using `testing/quick`.
- [x] `testing/quick` used throughout with bounded-float generators to avoid overflow; failures surface as concrete counterexamples.

### 8.7 Exit criteria

- [x] `go test ./...` passes (51 packages, 0 failures).
- [ ] Visual regression suite passes in CI.
- [ ] No AGG2D parity row remains untriaged or placeholder-level.
- [ ] Visual references and approval workflow centralized under `tests/visual/`.

---

## Phase 9 - Example and Demo Parity

Primary goal: keep the example surface close to the upstream AGG demo set while remaining
idiomatic in Go.

### 9.1 Example parity infrastructure

- [x] For each newly ported upstream demo: record the C++ source, decide placement (standalone vs web demo), and add a minimal verification path so the demo does not silently rot.
- [x] Reuse shared helpers and assets where possible.
- [x] Recorded and wired: `gradient_focal.cpp`, `line_thickness.cpp`, `rasterizer_compound.cpp`, `image_resample.cpp`, `line_patterns_clip.cpp`, `line_patterns.cpp`, `scanline_boolean2.cpp`.
      Sources: `../agg-2.6/agg-src/examples/{gradient_focal.cpp,line_thickness.cpp,rasterizer_compound.cpp,image_resample.cpp,line_patterns_clip.cpp,line_patterns.cpp,scanline_boolean2.cpp}`.
      Standalone: `examples/core/intermediate/{gradient_focal,line_thickness,rasterizer_compound,image_resample,line_patterns_clip,line_patterns,scanline_boolean2}/main.go`.
      Web: `cmd/wasm/{demo_gradient_focal.go,demo_line_thickness.go,demo_rasterizer_compound.go,demo_image_resample.go,demo_line_patterns_clip.go,demo_line_patterns.go,demo_scanline_boolean2.go}` + `web/index.html`.
      Verification: `cmd/wasm/{main_stub.go,render_test.go}` demo switch paths.
      Extras: `line_patterns.cpp` publishes `examples/shared/art/1.bmp..9.bmp` from `line_patterns.bmp.zip`; `scanline_boolean2.cpp` reuses `internal/demo/aggshapes/shapes.go`.

### 9.2 High-priority remaining demo ports

- [x] Completed: `raster_text.cpp`, `image_resample.cpp`, `gradient_focal.cpp`, `line_patterns.cpp`, `line_patterns_clip.cpp`, `line_thickness.cpp`, `rasterizer_compound.cpp`, `scanline_boolean2.cpp`.
- [x] Quad-warp cluster: `pattern_perspective.cpp`, `pattern_resample.cpp`, `image_perspective.cpp`.
      Sources `../agg-2.6/agg-src/examples/{pattern_perspective.cpp,pattern_resample.cpp,image_perspective.cpp}`; standalone `examples/core/intermediate/{pattern_perspective,pattern_resample,image_perspective}/main.go`; web `cmd/wasm/{demo_pattern_perspective.go,demo_pattern_resample.go,demo_image_perspective.go}` + `web/index.html`; shared rendering `internal/demo/quadwarp/draw.go`; embedded assets `examples/shared/art/{agg.ppm,spheres.ppm}` via `internal/demo/imageassets/assets.go`; verification `cmd/wasm/{main_stub.go,render_test.go}`.

### 9.3 Medium-priority demo ports

- [x] Completed: `interactive_polygon.cpp`, `graph_test.cpp`, `gpc_test.cpp`, `gradients_contour.cpp`, `flash_rasterizer2.cpp`, `polymorphic_renderer.cpp`.
- [x] `image_fltr_graph.cpp`: source `../agg-2.6/agg-src/examples/image_fltr_graph.cpp`; standalone `examples/core/intermediate/image_fltr_graph/main.go`; web `cmd/wasm/demo_image_fltr_graph.go` + `web/index.html`; shared rendering `internal/demo/imagefltrgraph/draw.go`; URL/HTML controls `web/{event-handlers.js,demo-state.js,url-state.js}`; verification `cmd/wasm/{main.go,main_stub.go,render_test.go}`.
- [x] `blend_color.cpp`: shared draw `internal/demo/blendcolor/draw.go`; standalone `examples/core/intermediate/blend_color/main.go`; web `cmd/wasm/demo_blend_color.go` + `web/index.html`; infrastructure `RendererBase.BlendFromColor`/`BlendFromLUT` + gray8 `GrayImageInterface` compliance.
- [x] `image_filters2.cpp`: shared renderer `internal/demo/imagefilters2/draw.go`; standalone `examples/core/intermediate/image_filters2/main.go`; web `cmd/wasm/demo_image_filters2.go` + `web/index.html`; verification `cmd/wasm/{main.go,main_stub.go,render_test.go}`.

### 9.4 Lower-priority or support-heavy demos

Triage each: fully port, replace with Go-idiomatic equivalent, or defer with rationale:

- [x] `trans_curve1.cpp`, `trans_curve1_ft.cpp`, `trans_curve2_ft.cpp`: Go-idiomatic/vector-text and FreeType-outline variants wired via `internal/demo/transcurve/draw.go` and `examples/core/intermediate/{trans_curve,trans_curve1_ft,trans_curve2_ft}/main.go`, with runtime fallback where FreeType or fonts are unavailable.
- [x] Shared shape assets: `make_arrows.cpp` and `make_gb_poly.cpp` live in `internal/demo/aggshapes/shapes.go` (`MakeArrows`, `MakeGBPoly`), are reused across demos, and are covered by `internal/demo/aggshapes/shapes_test.go`.
- [x] `mol_view.cpp`: shared renderer `internal/demo/molview/draw.go`; standalone `examples/core/intermediate/mol_view/main.go`; web `cmd/wasm/demo_mol_view.go` + `web/index.html`; embedded original `1.sdf` dataset and SDF parser.
- [x] `idea.cpp`: shared renderer `internal/demo/idea/draw.go`; standalone `examples/core/intermediate/idea/main.go`; web `cmd/wasm/demo_idea.go` + `web/index.html`; verification `cmd/wasm/{main.go,main_stub.go,render_test.go}`.
- [x] `truetype_test.cpp`: standalone FreeType showcase in `examples/core/intermediate/truetype_test/main.go` with gray8, outline, and mono panels plus runtime fallback when FreeType/font files are unavailable.

### 9.5 Bug fixing

- [x] `line_thickness` (web): aligned framing/centering against standalone render and C++ intent.
- [x] `line_patterns` (web): fixed empty output; verified pattern asset decode/load and span generation; added a non-empty render check in the demo benchmark/smoke path.
- [x] `line_patterns_clip` (web): fixed empty output; verified clip-box/path and pattern source wiring; added a non-empty render check.
- [x] `scanline_boolean2` (web): corrected map orientation, centered the original 655x520 reference frame, and fixed mouse Y mapping while preserving upstream `flip_y` parity.
- [ ] `trans_curve` (web): evaluate source bitmap choice and switch to a better upstream-compatible shared asset if available; centering parity fix already applied for the original 600x600 frame.
- [ ] `trans_curve2` (web): same bitmap/parity task as `trans_curve`; centering parity fix already applied for the original 600x600 frame.
- [x] `distortions` (web): fixed mostly-empty output by correcting wave-distortion amplitude math, switching to spheres source image, and matching upstream-style default center init.
- [x] `image1` (web): aligned transform/math to the upstream `image1.cpp` reference frame (`initial_w=src_w+20`, `initial_h=src_h+60`), centered that frame, switched to embedded `spheres.ppm` with procedural fallback, and sanitized invalid scale input.
- [ ] `image_resample` (web): restore draggable quad handles and ensure down/move/up handlers map to this demo as they do for perspective demos.
- [ ] `image_perspective` (web): add or fix draggable quad handles and mouse interaction wiring.
- [x] `image_transforms` (web): fixed near-empty output by mapping screen sampling into a centered source-image reference frame, sizing the star from source dimensions, and switching to embedded `spheres.ppm` with finite-scale guards.
- [ ] `pattern_fill` (web): fix empty output; verify offscreen pattern generation and final blend spans.
- [x] `pattern_perspective` (web): added or fixed draggable quad handles and mouse interaction wiring.
- [x] `pattern_resample` (web): added or fixed draggable quad handles and mouse interaction wiring.
- [x] `rasterizer_compound` (web): fixed upside-down/odd glyph rendering by applying upstream `flip_y` parity and centering the original 440x330 scene for closer standalone/C++ alignment.
- [ ] For all above: add per-demo parity notes (standalone vs web) plus a minimal verification path (render smoke and, where practical, non-empty or image-hash threshold checks).
- [ ] **FreeType raster glyph vertical baseline bug**: When using `RasterFontCache` with FreeType,
      glyphs with small vertical extent (e.g. `.`, `,`, `-`) are rendered at the wrong Y position —
      the period in "0.2" appears at cap-height instead of the baseline. The issue is in the glyph
      bitmap positioning pipeline: `InitEmbeddedAdaptors` passes `glyph.Bounds` to
      `NewSerializedScanlinesAdaptorAA`, and `glyphBitmapRasterizer.SweepScanline` computes
      `scanY = bounds.Y1 + offsetY + row`. If `Bounds.Y1` does not correctly encode the glyph's
      Y-bearing relative to the baseline (as set by the FreeType engine in
      `internal/font/freetype/engine.go` during glyph caching), short glyphs get vertically
      misplaced. Compare with how C++ AGG's `font_cache_manager::init_embedded_adaptors` uses
      `glyph->bounds` for gray8 data — the Go port may be computing or storing `Bounds.Y1/Y2`
      differently from the C++ original.
      **Tests needed**: Add a pixel-level regression test that renders a string containing mixed-height
      glyphs (e.g. "0.2 H,x-y") with `RasterFontCache` and asserts that the period/comma/hyphen
      pixels fall within the baseline-to-descender band, not above x-height. A visual golden-image
      test comparing against C++ AGG output for the same string and font size would be ideal.

### 9.6 Exit criteria

- [ ] Every remaining upstream demo is ported, replaced by a documented equivalent, or deferred.
- [ ] Newly added demos build and run through the existing example workflows.

---

## Phase 10 - Rendering Fidelity: Architectural Gaps

This phase captures all architectural gaps and subtle behavioral differences discovered through
systematic layer-by-layer comparison between Go and C++ AGG output (March 2026 investigation).
Each gap is concrete, traced to a C++ source reference, and has a verifiable fix path.

---

### 10.1 ✅ sRGB Color Handling in Lion/Demo Rendering

**Status**: DONE — lion hex colors are LINEAR values, not sRGB. No conversion needed.
Lion parser restructured to match C++ `parse_lion()` exactly: single shared PathStorage
with parallel Colors/PathIdx arrays, `ArrangeOrientationsAllPaths(PathFlagsCW)`.

**Root cause (revised)**: C++ `rgb8_packed()` returns `rgba8` (linear type). When `parse_lion`
stores into `srgba8[]`, the C++ type system roundtrips linear→sRGB→linear via converting
constructors — but the net result is identity (±1 rounding). The hex values represent the
actual linear colors to blend. Applying `ConvertRGBA8SRGBToLinear` would double-convert.

**Go implementation**: `lion.Parse()` returns `LionData` with single shared `*PathStorageStl`,
`Colors []RGBA8[Linear]`, `PathIdx []uint`, and `NPaths int`. All callers updated.

**Verification**: `TestCPPParity_Step6_LionThroughMask` confirms pixel(300,100) = (245, 217, 177).

**Tasks**:

- [x] Audit every lion/demo caller — all use colors as LINEAR, no spurious conversions.
- [x] Verify output PNG color values against C++ step6 reference: `out(300,100)` = (245,217,177) ✓
- [x] Restructure lion.go to single shared PathStorage matching C++ parse_lion.
- [x] Port `ArrangeOrientationsAllPaths` to `PathBase` (5 methods from C++ `agg_path_storage.h`).
- [x] Update all 26 callers to new `LionData` API.
- [ ] Update visual regression references (deferred — visual tests have broader pre-existing failures).

---

### 10.2 ⚠️ `PixFmtAMaskAdaptor` Cannot Wrap Non-RGBA Pixfmts

**Status**: OPEN — documented in Phase 2, tasks not yet started.

**Root cause**: `PixFmtInterface` (used by `PixFmtAMaskAdaptor`) hardcodes `RGBA8[Linear]` for
all blend methods. C++ uses templates, so `pixfmt_amask_adaptor<pixfmt_bgr24, mask>` compiles
fine. In Go, `PixFmtBGR24` (color type `RGB8[Linear]`) does not satisfy `PixFmtInterface`.

**C++ reference**: `agg_pixfmt_amask_adaptor.h` — `pixfmt_amask_adaptor<PixFmt, AlphaMask>` is a
template that delegates all blend calls to the wrapped `PixFmt` regardless of its color type.

**Go files**:

- `internal/pixfmt/pixfmt_amask_adaptor.go`: `PixFmtInterface`, `PixFmtAMaskAdaptor`
- `internal/pixfmt/pixfmt_rgb8.go`: `PixFmtBGR24` / `PixFmtAlphaBlendRGB`

**Tasks**:

- [ ] Introduce a generic `PixFmtBlendInterface[C color.ColorType]` that abstracts over color type.
- [ ] Make `PixFmtAMaskAdaptor` generic: `PixFmtAMaskAdaptor[C color.ColorType]`.
- [ ] Ensure `PixFmtBGR24` satisfies the new generic interface.
- [ ] Port `alpha_mask2` main buffer to `PixFmtBGR24` once the adaptor is generic.
- [ ] Add a test: wrap `PixFmtBGR24` with amask adaptor, assert correct blending output.

---

### 10.3 ⚠️ `render_all_paths` Equivalent Missing for Multi-Path Lion Rendering

**Status**: OPEN — Go examples iterate paths manually with `lp.Path.Rewind(0)` + `NextVertex()`,
but C++ uses `render_all_paths` with `conv_transform` which handles full AGG vertex-source
semantics including all command types (curves, end-poly, etc.).

**Root cause**: The Go examples only forward `MoveTo` and `LineTo` from `NextVertex()` and
silently drop `EndPoly`/`ClosePolygon` and any curve commands. `EndPoly` commands carry the
`PathFlagsClose` flag that tells the rasterizer to auto-close the polygon — dropping them may
cause unclosed paths or missed fills in some lion sub-paths.

**C++ reference**:

- `agg_renderer_scanline.h`: `render_all_paths(ras, sl, ren, vs, colors, path_idx, num_paths)`
- `agg_conv_transform.h`: `conv_transform<path_storage, trans_affine>` — wraps a vertex source
  with a transform, forwarding ALL commands (including EndPoly) unchanged.

**Go files**:

- `examples/core/intermediate/alpha_mask2/main.go` lines 260–279: manual vertex loop
- `cmd/aggtest/main.go` step6: same pattern

**Tasks**:

- [ ] Port `render_all_paths` to Go as `RenderAllPaths(ras, sl, ren, vs, colors, pathIdx)`.
      C++ source: `agg_renderer_scanline.h` `render_all_paths` template.
- [ ] Implement or reuse `ConvTransform` to wrap a path storage with an affine transform,
      forwarding all commands including `EndPoly|Close`.
- [ ] Update `alpha_mask2`, `alpha_mask`, `multi_clip`, and wasm demo equivalents to use
      the proper `RenderAllPaths` + `ConvTransform` pattern.
- [ ] Add a pixel-level test comparing `RenderAllPaths` output against C++ step6 reference.

---

### 10.4 ✅ `sgray8` vs `gray8` for Alpha Mask Generation

**Status**: DONE — All alpha mask examples switched to `PixFmtSGray8` with `Gray8[SRGB]` colors
to match C++ `pixfmt_sgray8` / `sgray8`. Output is byte-identical because the color space
parameter only affects floating-point conversions (which never occur in the mask blending path).

**C++ reference**:

- `agg_color_gray.h`: `gray8T<sRGB>` vs `gray8T<linear>` — `mult_cover` is the same 8-bit op
- `examples/alpha_mask2.cpp` line 197: `pixfmt_sgray8 pixf(m_alpha_mask_rbuf)`

**Tasks**:

- [x] Confirm whether `PixFmtSGray8` exists in Go. Yes: `internal/pixfmt/pixfmt_gray8.go`.
- [x] Switch `alpha_mask`, `alpha_mask2`, `alpha_mask3` mask pixfmt to `PixFmtSGray8`.
      Also switched: `cmd/aggtest`, `tests/integration/cpp_parity_test.go`, and 3 wasm demos.
- [x] Add a unit test comparing mask output byte-for-byte between `PixFmtGray8` and `PixFmtSGray8`.
      `TestGray8LinearSRGBMaskEquivalence` in `internal/pixfmt/additional_contract_test.go` —
      6 representative input combinations, all produce identical bytes.

---

### 10.5 ✅ `amask_no_clip_gray8` vs `AlphaMaskU8` — Clip vs No-Clip Variants

**Status**: DONE — All alpha mask examples and demos switched to `AMaskNoClipU8` to match C++.

**C++ reference**: `agg_alpha_mask_u8.h`: `amask_no_clip_gray8` = `amask_no_clip_u8<1, 0>`.

**Tasks**:

- [x] Switch `alpha_mask2`, `alpha_mask`, `alpha_mask3` examples to `AMaskNoClipU8` to match C++.
      Also switched: `cmd/aggtest`, `tests/integration/cpp_parity_test.go`, and 3 wasm demos.
- [x] Verify `AMaskNoClipU8.CombineHspan` produces identical output to `AlphaMaskU8.CombineHspan`
      for in-bounds coordinates. Table-driven equivalence test added in
      `internal/pixfmt/additional_contract_test.go:TestAlphaMaskClipNoClipEquivalence`.
      Covers CombinePixel (6 values × 16 positions), CombineHspan (4 rows), CombineVspan (4 cols).

---

### 10.6 ✅ C++ Rasterizer Uses `scanline_u8` Not `scanline_p8` in Some Examples

**Status**: DONE — Audited all C++ examples and aligned Go ports to use the matching scanline type.

**C++ reference**: `agg_scanline_u.h`: `scanline_u8` — unpacked; `agg_scanline_p.h`: `scanline_p8` — packed.

**Tasks**:

- [x] Audit each C++ example to identify which scanline type it uses.
      Cross-referenced all 60+ C++ examples against Go ports and wasm demos.
- [x] For demos that use `scanline_u8`, switch Go port to use `ScanlineU8`.
      Fixed 11 standalone examples and 5 wasm demos. Changed both directions:
      P8→U8 (aa_demo, aa_test, alpha_gradient, alpha_mask2, bezier_div, gamma_correction, multi_clip)
      and U8→P8 (pattern_fill, polymorphic_renderer, rasterizers, rasterizers2).
- [x] Ensure `ScanlineU8` is implemented in `internal/scanline/` and satisfies `ScanlineInterface`.
      Confirmed: `ScanlineU8` in `internal/scanline/scanline_u8.go` satisfies `scanline.Scanline`.
      Also fixed `internal/ctrl/render_test.go` mock to satisfy the unified interface.

---

### 10.7 ✅ Missing `scanlineWrapper` / Rasterizer Adaptor — Architecture Smell

**Status**: DONE — Unified scanline interface defined in `internal/scanline/interfaces.go`.
All adapter boilerplate removed from examples, wasm demos, internal demos, and tests.

**Root cause** (resolved): Two different `ScanlineInterface` types existed — one in
`internal/renderer/scanline` and one in `internal/rasterizer`. They had incompatible method sets,
forcing every example that called `RenderScanlinesAASolid` + `RasterizerScanlineAA` to write
4 adapter types.

**Tasks**:

- [x] Unify the scanline interface: single `scanline.Scanline` interface in `internal/scanline/interfaces.go`
      combining writer (rasterizer) and reader (renderer) methods. Rasterizer and renderer packages
      both use this interface via type aliases.
- [x] Remove the per-example adapter boilerplate (`rasterizerAdaptor`, `scanlineWrapper`,
      `rasScanlineAdapter`, `scanlineWrapperU8`, `scanlineWrapperP8`, `spanIter*`, etc.)
      from all example, wasm demo, internal demo, and test files.
- [x] Ensure all example and demo files compile without manual adapter types.
      `go build ./...` passes with zero errors.

---

### 10.8 ✅ `cmd/aggtest` Regression Suite — Make Permanent

**Status**: DONE — All 6 cmd/aggtest steps converted to proper integration tests in
`tests/integration/cpp_parity_test.go`. `cmd/aggtest` kept as a human-readable developer tool.

**Tasks**:

- [x] Convert `cmd/aggtest` steps into proper `testing.T`-based tests under `tests/integration/`.
      File: `tests/integration/cpp_parity_test.go` — 6 sub-tests matching cmd/aggtest steps 1–6.
- [x] Each step becomes a sub-test with `t.Errorf` on pixel mismatch, not just `fmt.Printf`.
- [x] Keep `cmd/aggtest` as a human-readable tool; the test suite is the ground truth.
- [x] Add the C++ reference pixel values as test constants with source references.
      Each test documents the C++ source file (step3_rgba.cpp, step4_lion.cpp, step6_lion_full.cpp).

---

### 10.9 ✅ Output PNG Byte Encoding — Linear vs sRGB

**Status**: DONE — Policy documented in `docs/AGG_DELTAS.md` under "Color Space".

**Tasks**:

- [x] Document in `docs/AGG_DELTAS.md`: Go port writes linear bytes to output buffers, same as C++.
      PNG files are not color-space tagged; viewers will interpret as sRGB (matches C++ behavior).
- [x] Add a note that reference PNG comparison is valid only when both sides use the same encoding.
      Documented: any switch to sRGB output encoding would be a breaking change requiring
      coordinated reference image updates.

---

### 10.10 Exit Criteria

- [ ] All items in 10.1–10.8 are resolved or explicitly deferred with rationale.
- [x] `cmd/aggtest` pixel values match C++ for all 6 steps.
      Proper integration tests in `tests/integration/cpp_parity_test.go`.
- [ ] `alpha_mask2`, `alpha_mask`, `alpha_mask3` visual output matches C++ reference images within
      the visual regression threshold.
- [x] The `rasterizerAdaptor`/`scanlineWrapper` boilerplate is removed from all example files.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement/fix Go behavior.
3. Add or update contract tests.
4. Add/update visual regression if rendering-visible.
5. Update this plan.
