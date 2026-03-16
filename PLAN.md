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
      (40 PNGs captured from `../agg-2.6/build/examples/` via the upstream X11
      `F2` screenshot path).
- [x] Generate the matching Go-port screenshot corpus for direct demo-level comparison.
      Stored under `tests/visual/reference/go/examples/` on 2026-03-16
      (40 PNGs rendered through the default headless demo runner path).
- [ ] Generate canonical references from C++ AGG for core scenarios and replace Go-side
      references where C++ output is the ground truth.
- [ ] Drive the new `tests/visual/demo_parity_test.go` corpus to green.
      Current status on 2026-03-16: all 40 imported demo pairs fail parity, split into
      dimension mismatches and same-size visual mismatches.
  - [ ] Fix demo frame/dimension mismatches against C++ output:
        `aa_demo`, `alpha_mask`, `alpha_mask2`, `alpha_mask3`, `bezier_div`, `blur`,
        `bspline`, `circles`, `component_rendering`, `compositing`, `compositing2`,
        `conv_contour`, `conv_stroke`, `distortions`, `flash_rasterizer`,
        `flash_rasterizer2`, `gamma_correction`, `gamma_ctrl`, `gouraud`,
        `gouraud_mesh`, `gradients`, `gradients_contour`, `image1`, `image_alpha`,
        `image_filters`, `image_filters2`, `image_fltr_graph`, `image_perspective`,
        `image_resample`, `line_patterns`, `lion`, `lion_lens`, `lion_outline`.
  - [ ] Fix same-size but visually divergent demo outputs:
        `blend_color`, `conv_dash_marker`, `gamma_tuner`, `gradient_focal`,
        `graph_test`, `idea`, `line_patterns_clip`.
- [ ] Expand C++-generated visual reference set:
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

### 9.6 Exit criteria

- [ ] Every remaining upstream demo is ported, replaced by a documented equivalent, or deferred.
- [ ] Newly added demos build and run through the existing example workflows.

---

## Working Cadence

For each task:

1. Link C++ source method(s).
2. Implement/fix Go behavior.
3. Add or update contract tests.
4. Add/update visual regression if rendering-visible.
5. Update this plan.
