# AGoGo Audit

This document captures the current migration audit of `../AGoGo` so the
remaining absorption work can happen in this repository without relying on that
repository as a runtime dependency.

## Goal

`../AGoGo` is now a donor repository, not the long-term home of the project.
This audit identifies:

- what is still worth porting or reusing
- what is broken today
- what must be treated as unsupported instead of silently falling back
- which old wrappers and tests no longer fit the direct `engine`-local native
  design used in this repository

## Immediately Observed Breakages

The following concrete problems were found during the audit and explain why the
old repository cannot be treated as a trustworthy source of truth without
review:

| Problem | Concrete paths | Notes |
| --- | --- | --- |
| Duplicate `compareImages` helper | `../AGoGo/go/lib/agg_test.go`, `../AGoGo/go/lib/visual_regression_test.go` | Collides in the same Go test package. |
| Duplicate `abs` helper names | `../AGoGo/go/lib/paint_capi_test.go`, `../AGoGo/go/lib/svg_test.go` | Another same-package helper collision. |
| Missing `CAPIImageGetBuffer` bridge helper | references in `../AGoGo/go/lib/error_handling_test.go`, `performance_benchmark_test.go`, `visual_regression_test.go`, `path_capi_test.go`; only `CAPIImagePixelsPtr` / `CAPIImagePixelsLen` exist in `../AGoGo/go/lib/capi_bridge.go` | Tests depend on a helper that is not actually exported by the bridge. |
| Stale enum names | `../AGoGo/go/lib/error_handling_test.go`, `performance_benchmark_test.go`, `visual_regression_test.go`, `path_capi_test.go` | These tests still reference `LineCapRound` / `LineJoinRound`. |

## Reusable Material Worth Mining

These parts of `../AGoGo` still contain useful migration value:

- `../AGoGo/cpp/src/impl/` and `../AGoGo/cpp/src/ffi/`
  - useful for AGG-backed rendering, font, transform, image, and compositing
    implementation patterns
- `../AGoGo/go/lib/`
  - useful for facade-level edge cases, input validation expectations, and
    examples of text/image/path behavior that should either be reproduced or
    explicitly rejected here
- `../AGoGo/benchmarks/`
  - useful once the shared comparison and benchmark layer is added in this repo
- `../AGoGo/assets/`
  - candidate source for comparison fixtures if they still match the supported
    shared facade surface

## Stale or Misleading Material

Some donor-repo material should not be carried over verbatim:

- `../AGoGo/README.md`
- `../AGoGo/ROADMAP.md`

Those documents still position AGoGo as the future pure-Go home of the project.
That story is now obsolete. This repository is the canonical home, with the
pure-Go engine as the default implementation and the in-repo C++ engine as an
optional comparison/performance path.

## Stub, Fallback, and Partial Paths

The audit found several categories of behavior that must be classified during
migration rather than copied forward blindly.

### Explicit stub or placeholder rendering paths

| Path | Observation | Migration stance |
| --- | --- | --- |
| `../AGoGo/cpp/src/impl/render_path.cpp` | Contains explicit stub drawing and TODO bounds comments. | Do not treat as supported without replacement or explicit rejection. |
| `../AGoGo/cpp/src/impl/render_compositing.cpp` | Contains stub-oriented temp-buffer rendering paths. | Comparison-only until parity is verified. |
| `../AGoGo/cpp/src/impl/transform.cpp` | Stub mode copies paths instead of performing real transformation work. | Must become real support or explicit unavailability. |
| `../AGoGo/cpp/src/impl/render_triangle.cpp` | Uses precomputed bounds for stub behavior. | Do not use as evidence of final parity. |
| `../AGoGo/cpp/src/impl/path_boolean_scanline.cpp` | Empty in stub builds. | Keep out of supported surface until validated. |
| `../AGoGo/cpp/src/impl/text_on_curve.cpp` | Contains fallback paths for stub mode. | Keep comparison-only for now. |
| `../AGoGo/cpp/src/impl/image_filter.h` | Includes simple copy / nearest-neighbor fallback paths. | Replace or document as intentionally limited. |

### Default fallback mappings that should become explicit errors

| Path | Observation | Migration stance |
| --- | --- | --- |
| `../AGoGo/go/lib/paint.go` | Unknown values fall back to `PaintSolid`. | Reject unknown values explicitly. |
| `../AGoGo/go/lib/compositing.go` | Unknown values fall back to `CompOpSrcOver`. | Reject unknown values explicitly. |
| `../AGoGo/go/lib/image.go` | Unknown values fall back to `PixelFormatRGBA32`. | Reject unknown values explicitly. |
| `../AGoGo/cpp/src/ffi/compositing.cpp` | Unknown values fall back to `CompOpSrcOver`. | Reject unknown values explicitly. |
| `../AGoGo/cpp/src/ffi/paint_core.cpp` | Unknown values fall back to `PaintSolid`. | Reject unknown values explicitly. |

### Known unimplemented or partial features

| Path | Observation | Migration stance |
| --- | --- | --- |
| `../AGoGo/cpp/src/impl/lowlevel_rasterizer_scanline_aa.cpp` | Explicit not-implemented path. | Out of scope until intentionally ported. |
| `../AGoGo/cpp/src/impl/pattern.cpp`, `pattern.h` | Bounds and outputs still marked TODO or not implemented. | Keep out of shared facade until verified. |
| `../AGoGo/cpp/src/impl/gradient.h` | Output-bounds support marked not implemented. | Port only the verified subset. |
| `../AGoGo/cpp/src/impl/image_core.cpp` | Multiple TODOs for pixel formats and conversions. | Treat as partial support only. |
| `../AGoGo/cpp/src/impl/font.cpp` | Rotation path still marked as future enhancement. | Document current text limits. |
| `../AGoGo/go/lib/svg_parser.go` | Pattern, gradient text, `tspan`, and `textPath` still TODO or fallback-based. | Keep out of the shared facade. |
| `../AGoGo/go/lib/svg.go` | Path data handling still stores references with TODO. | Do not treat as a stable import path yet. |

## Old Wrapper/Test Shape To Keep or Drop

The direct `engine`-local native design in this repository changes which old
AGoGo assets are still worth carrying forward.

Keep or adapt:

- behavior-focused tests that validate rendering semantics, capability
  rejection, edge-case geometry, or text/image/path correctness
- benchmarks that can be rewritten on top of the shared `engine` facade
- fixtures that help compare the pure-Go and C++ engines through the same
  high-level scene description

Do not port as-is:

- standalone bridge APIs whose only purpose was to expose raw C handles to the
  old `go/lib` package
- tests that only validate donor-repo wrapper shape rather than actual engine
  behavior
- helpers that depend on missing bridge functions such as `CAPIImageGetBuffer`
  when the same behavior can be tested directly through the in-repo `engine`
  and package-private native helpers

## Migration Implications

The current repository should continue with these constraints:

- keep the pure-Go `agg` API as the default and compatibility anchor
- keep the C++ backend internal to the `engine` package
- reject stub builds and unsupported features explicitly
- do not import AGoGo fallback defaults as if they were correct semantics
- only migrate SVG, pattern, text-on-curve, boolean, and advanced image/filter
  behavior after each path is classified as supported, unsupported, or
  comparison-only
