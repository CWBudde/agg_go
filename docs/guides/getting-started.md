# Getting Started

This guide uses the high-level `agg.Context` API to draw a few basic shapes and
save the result as a PNG.

## 1. Create a module and add the dependency

```bash
go mod init example.com/agg-demo
go get github.com/MeKo-Christian/agg_go
```

## 2. Draw into a Context

Create `main.go`:

```go
package main

import (
	"log"

	agg "github.com/MeKo-Christian/agg_go"
)

func main() {
	ctx := agg.NewContext(800, 600)
	ctx.Clear(agg.White)

	// Immediate-mode helpers render right away.
	ctx.SetColor(agg.NewColor(220, 70, 50, 255))
	ctx.FillRectangle(80, 80, 220, 140)

	ctx.SetColor(agg.NewColor(30, 90, 180, 255))
	ctx.SetLineWidth(6)
	ctx.DrawCircle(500, 280, 90)

	// Explicit path mode lets you build a custom shape before filling it.
	ctx.SetColor(agg.Black)
	ctx.BeginPath()
	ctx.MoveTo(120, 320)
	ctx.LineTo(260, 500)
	ctx.LineTo(60, 500)
	ctx.ClosePath()
	ctx.Fill()

	if err := ctx.GetImage().SaveToPNG("output.png"); err != nil {
		log.Fatal(err)
	}
}
```

Run it:

```bash
go run .
```

You should get `output.png` in the working directory.

## 3. Understand the two drawing styles

The `Context` API exposes two complementary styles:

- Immediate mode: helpers like `FillRectangle`, `DrawCircle`, and `DrawLine`
  build and render their geometry immediately.
- Path mode: `BeginPath`, `MoveTo`, `LineTo`, `ClosePath`, `Fill`, and
  `Stroke` let you build one or more contours explicitly before rendering.

Use immediate mode for common shapes and path mode when you need custom
geometry.

## 4. When to drop down to Agg2D

`Context` is the recommended entry point for simple drawing, image output, and
basic transforms. If you need closer parity with the original C++ AGG2D API,
use:

```go
agg2d := ctx.GetAgg2D()
```

That gives you access to gradients, image transforms, blend modes, text APIs,
and lower-level AGG2D path operations.

## 5. Next steps

- See the public parity notes in [../AGG2D_PARITY.md](../AGG2D_PARITY.md).
- Browse `examples/core/basic/` for small runnable examples.
- Use `just run-example basic/hello_world` to run the starter example already in
  this repository.
