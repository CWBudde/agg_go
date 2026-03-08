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
