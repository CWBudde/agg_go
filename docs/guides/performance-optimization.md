# Performance Optimization

This guide focuses on practical performance advice for users of the public API.

For internal SIMD and low-level optimization notes, see
[../SIMD_OPTIMIZATIONS.md](../SIMD_OPTIMIZATIONS.md).

## Start with the right API level

Use the highest-level API that still matches your workload:

- `Context` for straightforward scene rendering
- `Agg2D` when you need closer control over transforms, gradients, image
  mapping, blend modes, or path lifetime

The easiest performance mistake is rebuilding more state than necessary with the
high-level wrappers when the lower-level `Agg2D` surface would let you reuse it.

## Reuse rendering state

AGG-style rendering benefits from reusing stateful objects and avoiding repeated
setup:

- reuse a `Context` or `Agg2D` instance across frames when possible
- reuse the same backing `Image` instead of reallocating every draw
- keep fonts loaded rather than re-calling `Font(...)` for every string
- reuse dash, gradient, and transform state when rendering many similar shapes

## Prefer immediate helpers for simple shapes

For common shapes, the `Context` helpers are already efficient enough and reduce
path-management overhead in calling code:

- `FillRectangle`
- `DrawRectangle`
- `FillCircle`
- `DrawCircle`

Use explicit path mode when:

- the geometry is custom
- multiple segments belong to one logical fill/stroke operation
- you want closer parity with the original AGG2D path workflow

## Minimize image resampling cost

Transformed image rendering is one of the more expensive public operations.

Practical guidance:

- use `NoResample` when nearest behavior is acceptable
- use `ResampleOnZoomOut` when quality matters mainly for downscaling
- use `ResampleAlways` only when you really need higher-quality transformed image sampling
- keep the image filter simple unless you have a visible quality reason to change it

Example:

```go
ctx.SetImageFilter(agg.Bilinear)
ctx.SetImageResample(agg.ResampleOnZoomOut)
```

## Batch transform changes

Changing transforms is cheap, but excessive churn still adds up in interactive
rendering.

- use `PushTransform` / `PopTransform` around local transforms
- avoid rebuilding equivalent transforms repeatedly
- prefer one transform around a batch of related drawing calls instead of
  transforming every item independently in user code

## Fonts and text

Text performance depends heavily on backend and cache mode.

- `RasterFontCache` is the faster default for normal screen rendering
- `VectorFontCache` is more flexible when scaling or transforming text
- `TextWidth` and `Text` benefit from glyph cache reuse once a font is loaded
- repeated font loading is expensive; keep font setup out of inner loops

If you are running without FreeType, `FontGSV(...)` is available as a fallback,
but it is a feature/portability path, not a general high-performance text path.

## Images and compositing

For image-heavy scenes:

- avoid unnecessary conversion between `Image` and standard-library images
- keep source images loaded and reused
- choose blend modes deliberately; more complex compositing costs more than
  plain source-over style drawing
- use `BlendImageSimple` or `CopyImageSimple` when you do not need full geometric transforms

## Avoid avoidable allocations

At the public API level, the easiest allocation wins come from:

- reusing `Context` and `Image` objects
- reusing parallelogram slices when repeatedly calling transformed image helpers
- keeping reusable geometry data in slices instead of rebuilding it per frame

## Measure before going lower

Before dropping into custom rasterizer/span work:

1. benchmark the current code path
2. identify whether shape generation, text, image transforms, or output dominates
3. only then switch from `Context` to `Agg2D` or lower-level internal APIs

The port already contains SIMD and lower-level optimizations in important pixel
paths, so many workloads will benefit more from better scene/state organization
than from immediately bypassing the public API.

## Related material

- [getting-started.md](./getting-started.md)
- [basic-shapes.md](./basic-shapes.md)
- [image-compositing.md](./image-compositing.md)
- [../SIMD_OPTIMIZATIONS.md](../SIMD_OPTIMIZATIONS.md)
