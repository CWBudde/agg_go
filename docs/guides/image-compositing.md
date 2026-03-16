# Image Compositing

This guide shows the image compositing workflow exposed by the public API and
how it maps to the original AGG2D image functions.

At the `Context` level, the usual flow is:

1. load or create an `Image`
2. configure filtering or blend state
3. draw or transform the image
4. save the target image

## Draw an image at a position

```go
ctx := agg.NewContext(800, 600)
ctx.Clear(agg.White)

img, err := agg.LoadImageFromFile("input.png")
if err != nil {
	log.Fatal(err)
}

if err := ctx.DrawImage(img, 80, 60); err != nil {
	log.Fatal(err)
}
```

`DrawImage` is the high-level wrapper over the AGG2D whole-image transform
overload.

## Scale an image

```go
ctx.SetImageFilter(agg.Bilinear)
ctx.SetImageResample(agg.ResampleAlways)

if err := ctx.DrawImageScaled(img, 300, 60, 220, 160); err != nil {
	log.Fatal(err)
}
```

Filtering and resampling affect subsequent transformed image draws.

## Draw a source region into a destination rectangle

```go
if err := ctx.DrawImageRegion(img, 20, 20, 120, 90, 80, 260, 240, 180); err != nil {
	log.Fatal(err)
}
```

This maps to the AGG2D `transformImage` overload that takes both source and
destination rectangles.

## Blend mode control

The high-level `Context` exposes blend-mode control through the underlying
compositing API:

```go
ctx.SetBlendMode(agg.BlendMultiply)
if err := ctx.DrawImageScaled(img, 420, 240, 220, 160); err != nil {
	log.Fatal(err)
}

ctx.SetBlendNormal()
```

For image-specific blend state closer to the C++ AGG2D API, use `Agg2D`
directly:

```go
a := ctx.GetAgg2D()
a.ImageBlendMode(agg.BlendScreen)
a.ImageBlendColor(agg.NewColor(255, 255, 255, 192))
if err := a.BlendImageSimple(img, 520, 80, 255); err != nil {
	log.Fatal(err)
}
```

That corresponds to the original `imageBlendMode`, `imageBlendColor`, and
`blendImage` method family from `agg2d.h`.

## Transform into a parallelogram

The lower-level `Agg2D` API gives direct access to AGG’s parallelogram image
mapping interface:

```go
para := []float64{
	120, 420,
	320, 390,
	140, 560,
}

if err := ctx.GetAgg2D().TransformImageParallelogramSimple(img, para); err != nil {
	log.Fatal(err)
}
```

This matches the C++ overload family that maps an image to a destination
parallelogram.

## Save the composited result

```go
if err := ctx.GetImage().SaveToPNG("composited.png"); err != nil {
	log.Fatal(err)
}
```

## Related example

For a more AGG-style example built around the lower-level API, see
`examples/core/intermediate/compositing`, which follows the original C++
compositing demo structure more closely.
