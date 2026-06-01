# Known Deltas from C++ AGG 2.6

This document records intentional deviations of the Go port from the original
C++ AGG 2.6 implementation. Each entry states what differs, why, and where the
relevant Go code lives.

---

## Language and Runtime Differences

### Memory management

**C++**: Manual `new`/`delete`, custom pod allocators, raw pointer arithmetic.
**Go**: Garbage-collected slices replace all manual allocation. No semantic
difference for the rendering pipeline; GC pause impact is acceptable for
typical usage.

### Templates → Generics

**C++**: Heavily templated (e.g., `agg::renderer_scanline_aa_solid<PixFmt>`).
**Go**: Go generics with explicit type constraints. The public `agg` package
exposes concrete types; generics remain internal.

### Virtual dispatch → Interfaces

**C++**: Virtual methods for polymorphism.
**Go**: Go interfaces. Behavior is identical; struct layout and dispatch
mechanism differ.

---

## Intentional Feature Deltas

### FreeType custom memory hook not supported

**C++ source**: `agg_font_freetype.h` — `FT_New_Library` + custom `FT_Memory`.
**Go**: `FT_Init_FreeType` is used instead; any `ftMemory` parameter is
ignored (`_ = ftMemory`). FreeType manages its own heap via the system
allocator. Custom memory hooks offer no practical benefit in a GC environment.
**File**: `internal/font/freetype2/engine.go`

### `maxFaces` cap (font engine)

**C++**: No fixed cap on simultaneously open FreeType faces.
**Go**: A configurable `maxFaces` limit is enforced to bound goroutine/GC
pressure from open cgo handles. Documented as a Go-only policy delta.
**File**: `internal/font/freetype2/engine.go`

### TransPolar not implemented

**C++**: `agg_trans_polar.h` exists as a standalone header for polar coordinate
mapping, but it is used only in one example (`polar_transformer.cpp`) and is
not part of the core AGG library.
**Go**: Not ported. The transformation is example-only in C++ AGG 2.6 with no
corresponding `.cpp` implementation file.

### TransWarpMagnifier — single zone only

**C++**: `agg_trans_warp_magnifier.h` supports a single circular magnification
zone. The multiple-zone variant discussed in some AGG forks is not in AGG 2.6.
**Go**: Single-zone warp magnifier implemented at full parity with the C++ 2.6
header. Multi-zone support is out of scope.
**File**: `internal/transform/trans_warp_magnifier.go`

### TransViewport — multi-viewport not implemented

**C++**: `trans_viewport` provides a single viewport mapping.
**Go**: `TransViewport` implements the same single-viewport contract. A
multi-viewport manager (`ViewportManager`) is provided as a Go addition, but
the underlying per-viewport semantics match C++. Batch transformations and
zoom/pan integration are not in scope.
**File**: `internal/transform/trans_viewport.go`

---

## Rendering Behavior Deltas

### Bilinear filter: no premultiplied-alpha clamping

**C++ source**: `agg_span_image_filter_rgba.h` — bilinear RGBA filter does not
clamp `r/g/b ≤ a` after sampling. Some earlier ports added such a clamp.
**Go**: The clamp is absent, matching C++ AGG 2.6 exactly. Callers that supply
premultiplied source data will get correct output; straight-alpha sources may
produce values `> a` after bilinear interpolation (same as C++).
**File**: `internal/span/span_image_filter_rgba.go`

### Image rendering uses premultiplied renderer

**C++ source**: `agg2d.cpp:1738` — `renderImage` routes through `m_renBasePre`
(the premultiplied base renderer). No automatic straight→premultiplied
conversion occurs in the span path.
**Go**: Same behavior: image transforms use the premultiplied renderer. Source
images must already be in premultiplied form for correct output. This is
consistent with C++ but is called out here because it affects test data and
image loading behavior.
**File**: `internal/agg2d/agg2d.go`, `internal/agg2d/image_test.go`

### GradientContour scaling formula

**C++ source**: `agg_gradient_lut.h` / `agg_span_gradient_contour.h` —
calculate uses `buffer * (d2 / 256) + d1`.
**Go**: Matches this formula exactly after a bug fix in Phase 5.3. Earlier Go
versions used a linear lerp, which was incorrect.
**File**: `internal/span/span_gradient_contour.go`

---

## Color Space

### Linear + sRGB only

**C++**: AGG 2.6 supports a range of color spaces via template parameters.
**Go**: Only `color.Linear` and `color.SRGB` color spaces are implemented. This
covers the primary use cases. Additional color spaces (e.g., wide-gamut) are
not in scope for the current port.
**Files**: `internal/color/`

### Output buffer byte encoding — linear, same as C++

**C++**: The rendering pipeline operates in linear space. Output buffers
(e.g., BGR24, RGBA32) contain linear-encoded byte values. When written to PNG
or displayed via platform screenshot, no gamma or sRGB encoding is applied —
the raw linear bytes are stored as-is. PNG viewers interpret these bytes as
sRGB (since no ICC profile is embedded), which makes the output appear
pale/washed-out. This is the standard AGG 2.6 behavior.

**Go**: Same behavior — output buffers contain linear-encoded bytes, identical
to C++. PNG files written from these buffers have no color-space tag and will
be interpreted as sRGB by viewers, producing the same pale appearance.

**Implications for visual comparison**: Because both C++ and Go write the same
linear bytes, pixel-level parity tests (`tests/integration/cpp_parity_test.go`)
compare identical byte encodings. If the Go port ever switches to writing
sRGB-encoded bytes to the output buffer (e.g., for display correctness), all
reference images and pixel-level comparisons would need regeneration.

**Policy**: The Go port maintains linear byte encoding in output buffers to
match C++ AGG 2.6. Any change to sRGB output encoding would be a breaking
change requiring coordinated reference image updates.

**C++ reference**: `agg_pixfmt_rgba.h` — blending always operates on the raw
buffer bytes; no gamma/sRGB encoding pass exists between the blender and the
buffer write.

---

## Float Agg2D Variant (`AGG2D_USE_FLOAT_FORMAT`)

C++ AGG 2.6's `Agg2D` has a compile-time `AGG2D_USE_FLOAT_FORMAT` switch
(`agg2d/agg2d.h`) that swaps the internal `ColorType` from `agg::rgba8` to the
float color `agg::rgba32`, dragging the pixel format, blender, span, and
gradient-LUT types along with it. The Go port provides the same capability as an
explicit, dedicated path rather than a build-time switch.

### Selection is explicit, not a build tag

**C++**: Precision is chosen at compile time via `#define AGG2D_USE_FLOAT_FORMAT`;
a single translation unit is either 8-bit or float.
**Go**: There is no build tag. Callers opt into the float path by constructing
`Agg2DFloat` / `ContextFloat` / `ImageFloat`. The existing 8-bit `Agg2D`,
`Context`, and `Image` are untouched and remain the default. This lets both
precisions coexist in one binary for side-by-side parity and debugging.
**Files**: `agg2d_float.go`, `context_float.go` (root), `internal/agg2d/agg2d_float.go`.

### Naming: `RGBA128` = 128-bit pixel (4 × float32)

**C++**: The float pixel format is `pixfmt_rgba` parameterised on `rgba32`
(`agg_pixfmt_rgba.h`); AGG also names a 4×float32 format `pixfmt_rgba128`.
**Go**: The existing 8-bit pixfmt is _already_ aliased `PixFmtRGBA32*` ("RGBA32"
= 32-bit **pixel**, 8 bits/channel), so the float stack is named by **total
pixel width = 4 × float32 = 128 bits** to avoid a real collision, matching AGG's
own `pixfmt_rgba128`. The float color it pairs with is `color.RGBA32` (float32
channels, the Go equivalent of `agg::rgba32`).
**Files**: `internal/pixfmt/blender/rgba128.go` (`BlenderRGBA128{,Pre,Plain}`),
`internal/pixfmt/pixfmt_rgba128.go` (`PixFmtRGBA128{,Pre,Plain}`).

### Public `Color` stays 8-bit

**C++**: `Agg2D::Color` is `srgba8` regardless of the format switch; only the
internal `ColorType` becomes float.
**Go**: Same — the public `Color` passed to `Agg2DFloat`/`ContextFloat` is still
the 8-bit `agg.Color`. Internally it is bridged to `color.RGBA32` and flows
through the float pixel pipeline.
**File**: `internal/agg2d/agg2d_float.go` (`colorToRGBA32` boundary helper).

### Premultiply/demultiply boundary contract

**C++**: `rgba32T::premultiply()`/`demultiply()` (`agg_color_rgba.h`, ~line 1243)
use the float formulas `r*=a; g*=a; b*=a` (and the inverse), guarded by
`if (a < 1)` so opaque pixels are left untouched, with `a <= 0` zeroing RGB.
**Go**: Internal storage and the `Plain`/`Pre` blender split mirror the 8-bit
semantics exactly. Exported helper APIs expose **straight (non-premultiplied)**
float data at the boundary; conversion to/from premultiplied happens inside the
pixfmt blenders, identical to the 8-bit path. The `color.RGBA32` and
`ImageFloat` premultiply/demultiply reproduce the C++ float formulas including
the `a < 1` opaque-no-op guard. Boundary conversions
(`ToRGBA`/`ToNRGBA64`/`ToImage8` and their `From*` inverses) honor each format's
alpha convention (`ToRGBA` premultiplies for Go's `image.RGBA`; `ToNRGBA64`
stays straight). Source-linked tests:
`internal/agg2d/premultiply_float_test.go`.
**Files**: `internal/agg2d/buffer_float.go`, `internal/pixfmt/blender/rgba128.go`,
`internal/color/rgba32.go`.

### Float image transforms (`TransformImage*`)

**C++**: `Agg2D::renderImage` (`agg2d/agg2d.cpp`) instantiates
`span_image_filter_rgba_{nn,bilinear,2x2}` / `span_image_filter_rgba` /
`span_image_resample_rgba_affine` over `image_accessor_clone<PixFormat>`, where
`PixFormat` is the float `pixfmt_rgba128` under `AGG2D_USE_FLOAT_FORMAT`; the
quad overloads use `span_interpolator_persp_lerp`.
**Go**: The float path mirrors this completely. `internal/span/span_image_filter_rgba32.go`
holds float twins of the NN / bilinear / 2×2 / general / affine-resample RGBA
filters (reusing the color-agnostic `SpanImageFilter`/`SpanImageResampleAffine`
bases), producing `color.RGBA32` from straight `[]float32` source rows.
`internal/agg2d/adapters_float.go` provides the clone-clamped `imagePixelFormatFloat`
source, and `internal/agg2d/image_transform_float.go` reproduces
`renderImage`/`newImageFilterGenerator`/`renderImagePerspective` plus the full
`TransformImage*` / `…Parallelogram*` / `…Path*` / `…Quad*` surface (public
wrappers in `agg2d_float.go`). Like the 8-bit path it transfers through the
premultiplied base renderer, so source images should be premultiplied first
(opaque sources are unaffected). Parity is covered by
`internal/agg2d/image_transform_float_test.go` (affine/parallelogram/quad vs the
8-bit path) and a visual hook (`tests/visual/float_image_transform_test.go`).

> **Bilinear bias deviation.** AGG's shared `span_image_filter_rgba_bilinear`
> template seeds its accumulator with `image_subpixel_scale²/2` as an integer
> rounding bias applied before the final `downshift`. For 8-bit channels that is
> ≈0.5 of one 0..255 unit (harmless rounding); for `rgba32` the channels are
> floats in [0,1], so the same bias would add a full **+0.5** to every channel
> and corrupt the result. Float color does not quantize, so the float bilinear
> generator omits the bias and computes a true weighted average — which is what
> actually matches the 8-bit output within tolerance. The LUT-based generators
> (2×2/general/resample) clamp against `full_value()` = 1.0, mirroring AGG.

### Capability gaps vs the 8-bit variant (deferred)

The float twin is usable today for clear/fill/stroke, path rendering, image
copy/blend, **affine/perspective image transforms**, gradients (linear/radial),
world transforms, and fill rules, and is covered by a cross-precision parity
test (`internal/agg2d/parity_float_test.go`), image-transform parity tests
(`internal/agg2d/image_transform_float_test.go`), and visual hooks
(`tests/visual/float_path_test.go`, `tests/visual/float_image_transform_test.go`).
The following 8-bit features are **not yet ported** to the float path and are
intentionally deferred (tracked in PLAN.md §4.7):

- **Composite blend modes** — there is no `PixFmtCompositeRGBA128`; `BlendMode`/
  `imageBlendMode` state exists but does not yet affect rendering. Src-over is
  the only effective mode.
- **Text glyph rendering** — only text _state_ (alignment, flip, hints, font
  height) is plumbed; `Text()` glyph rasterization is not yet mirrored.
- **Remaining ~100 public-method delegations** — e.g. Viewport, Parallelogram,
  Arc, RoundedRect, Polygon, Star, dashes, positioned/multi-stop gradient
  variants, `Get*` accessors, `GouraudTriangle`, and the transform stack.

---

## Public API Additions (Go-only, no C++ equivalent)

These additions have no C++ counterpart and are Go-specific conveniences:

- `Context` — high-level ergonomic API wrapping `Agg2D` (similar to HTML5
  Canvas API).
- `ContextStrokeAttributes` — snapshot/restore of all stroke state.
- `ViewportManager` — manages multiple named viewports.
- `FontGSV` — exposes the built-in GSV stroke-vector font directly on `Agg2D`
  without requiring an external font file (useful in WASM/no-cgo builds).
- `GouraudTriangle` — exposes Gouraud shading as a single method on `Agg2D`
  for convenience.
- `RenderRasterizerWithColor`, `ScanlineRender`, `RenderScanlinesAAWithSpanGen`
  — advanced escape hatches for direct rasterizer/renderer access.
