# Basic Shapes

This guide covers the two shape-building styles exposed by the public
`agg.Context` API:

- immediate-mode helpers for common shapes
- explicit path construction for custom geometry

Both styles ultimately map to the AGG2D path and rasterization pipeline from
the original C++ library, but the `Context` wrapper keeps the common cases
short.

## Immediate-mode helpers

Immediate-mode helpers build and render the shape in one call.

```go
ctx := agg.NewContext(640, 480)
ctx.Clear(agg.White)

ctx.SetColor(agg.Red)
ctx.FillRectangle(40, 40, 180, 100)

ctx.SetColor(agg.Blue)
ctx.SetLineWidth(4)
ctx.DrawRectangle(260, 40, 180, 100)

ctx.SetColor(agg.Green)
ctx.FillCircle(130, 240, 60)

ctx.SetColor(agg.Black)
ctx.DrawEllipse(360, 240, 90, 50)
```

These helpers are the easiest way to render:

- rectangles: `DrawRectangle`, `FillRectangle`
- circles: `DrawCircle`, `FillCircle`
- ellipses: `DrawEllipse`, `FillEllipse`
- rounded rectangles: `DrawRoundedRectangle`, `FillRoundedRectangle`
- lines: `DrawLine`, `DrawThickLine`

## Path mode

For custom shapes, build the path explicitly and then call `Fill` or `Stroke`.

```go
ctx.SetColor(agg.Black)
ctx.BeginPath()
ctx.MoveTo(500, 120)
ctx.LineTo(580, 280)
ctx.LineTo(420, 280)
ctx.ClosePath()
ctx.Fill()
```

This matches the AGG2D model more closely:

- `BeginPath` clears the current path
- `MoveTo` starts a contour
- `LineTo` appends segments
- `ClosePath` closes the contour
- `Fill` or `Stroke` rasterizes the finished path

## Rounded rectangles

The `Context` wrapper exposes a simple single-radius rounded rectangle helper:

```go
ctx.SetColor(agg.Orange)
ctx.FillRoundedRectangle(60, 320, 220, 90, 18)
```

If you need closer parity with the original AGG2D overloads, drop down to
`Agg2D`:

```go
a := ctx.GetAgg2D()
a.FillColor(agg.Purple)
a.RoundedRectXY(320, 320, 520, 410, 24, 12)
a.DrawPath(agg.FillOnly)
```

That maps directly to the C++ `roundedRect(x1, y1, x2, y2, rx, ry)` family.

## Fill and stroke state

The `Context` convenience API uses the current fill and stroke state held inside
the underlying `Agg2D` renderer.

- `SetColor` updates both fill and stroke colors at once.
- `SetLineWidth` affects subsequent stroked operations.
- Immediate-mode helpers render immediately.
- Path mode waits until `Fill` or `Stroke`.

One common mistake is calling `DrawRectangle` and then `Fill()`. That is not
necessary: `DrawRectangle` already renders. Use `FillRectangle` for a filled
rect, or use path mode if you want to build the geometry yourself.

## Saving the result

```go
if err := ctx.GetImage().SaveToPNG("basic-shapes.png"); err != nil {
	log.Fatal(err)
}
```

## Related APIs

- `Context` for the high-level workflow
- `Agg2D` for closer C++ parity and advanced shape/path operations
- `examples/core/basic/hello_world` for a minimal runnable example
