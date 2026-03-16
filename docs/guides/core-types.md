# Core Types

This guide documents the main public value types exposed at the module root.

The original project plan still refers to `types.go`, but the current public API
spreads those types across focused files such as `colors.go`, `geometry.go`,
`images.go`, and `transforms.go`.

## Public concrete types

At the root package, the API is intentionally concrete and Go-friendly:

- `Color` in `colors.go`
- `Point`, `PointI`, `Rect`, `Size`, `SizeI`, `Affine` in `geometry.go`
- `Image` in `images.go`
- `Transformations` in `transforms.go`

Unlike the internal packages, these root types are not generic. That is
intentional: the public API aims to be simple to consume, while the fidelity and
generic abstractions live behind the `internal/` boundary.

## `Color`

`Color` is the main root-level color type:

```go
c := agg.NewColor(255, 128, 64, 255)
```

Useful helpers:

- `NewColor(r, g, b, a uint8)`
- `NewColorRGB(r, g, b uint8)`
- `RGB(r, g, b float64)`
- `RGBA(r, g, b, a float64)`
- predefined colors like `Black`, `White`, `Red`, `Green`, `Blue`

`Color` is the type used by both `Context` and `Agg2D`.

## `Point` and `PointI`

Use `Point` for floating-point geometry:

```go
p := agg.NewPoint(10.5, 20.25)
```

Use `PointI` when you need integer coordinates:

```go
pi := agg.NewPointI(10, 20)
pf := pi.ToFloat()
```

## `Rect`

`Rect` is the root-level integer rectangle:

```go
r := agg.NewRect(10, 20, 110, 70)
width := r.Width()
height := r.Height()
```

This is a convenience value type for public code. The lower-level rasterization
pipeline uses more specialized rectangle forms internally.

## `Affine` and `Transformations`

There are two related transformation-facing types:

- `Affine` in `geometry.go`
- `Transformations` in `transforms.go`

`Transformations` is the type used by the public `Context` and `Agg2D`
transformation APIs. It corresponds to the AGG2D `Transformations` struct from
the original C++ interface.

Example:

```go
tr := agg.Translation(40, 20)
ctx.SetTransform(tr)
```

`Affine` is a lighter helper type that mirrors AGG `trans_affine` more directly.

## `Image`

`Image` is the public raster buffer wrapper used for:

- rendering destinations
- loading from files
- image transforms and compositing
- conversion to standard Go `image.Image`

Common entry points:

```go
img, err := agg.LoadImageFromFile("input.png")
if err != nil {
	log.Fatal(err)
}

ctx := agg.NewContextForImage(img)
_ = ctx
```

## Paths in the public API

There is no root-level exported `Path` type today.

Instead, the public API exposes path construction procedurally:

- `Context.BeginPath`
- `Context.MoveTo`
- `Context.LineTo`
- `Context.ClosePath`
- `Context.Fill`
- `Context.Stroke`

For closer parity with the original AGG2D path model, use `ctx.GetAgg2D()` and
the lower-level `Agg2D` path commands.

## Why generics are not in the root API

The internal port uses generics heavily to model AGG templates safely and
idiomatically. The root package intentionally hides most of that complexity.

That means:

- public code uses straightforward concrete types
- internal packages keep the template-to-generics fidelity
- callers get a simpler API surface without losing implementation fidelity

This split is deliberate and matches the project goal of a faithful AGG port
with an idiomatic Go entry point.
