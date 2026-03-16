# Migrating From C++ AGG

This guide is for readers coming from the original C++ AGG 2.6 codebase in
`../agg-2.6/agg-src/`.

The Go port keeps the rendering pipeline and most of the high-level `Agg2D`
shape/state model close to upstream, but it exposes two public entry points:

- `agg.Agg2D` for the closest match to C++ `Agg2D`
- `agg.Context` for a more Go-idiomatic immediate-mode API

For exact `Agg2D` overload/name translation, also see
[../AGG2D_PARITY.md](../AGG2D_PARITY.md).

## Choosing The API

Use `agg.Agg2D` when:

- you are translating code from `agg2d.h` or `agg2d.cpp`
- you want upstream naming and state behavior
- you intend to adapt original AGG2D examples or documentation closely

Use `agg.Context` when:

- you are writing new Go code
- you want one owned RGBA image plus simple draw helpers
- you do not need direct parity with every `Agg2D` method family

## Type Mapping Reference

### High-level API

| C++ AGG                                 | Go port       | Notes                               |
| --------------------------------------- | ------------- | ----------------------------------- |
| `Agg2D`                                 | `agg.Agg2D`   | Closest public parity surface       |
| `Agg2D::Image`                          | `agg.Image`   | RGBA image wrapper                  |
| `Agg2D::Color` / `rgba8` usage in Agg2D | `agg.Color`   | Root-level public color type        |
| no direct equivalent                    | `agg.Context` | Go convenience wrapper over `Agg2D` |

### Core geometry and transforms

| C++ AGG                        | Go port                               | Notes                                                            |
| ------------------------------ | ------------------------------------- | ---------------------------------------------------------------- |
| `agg::rect_i` / `rect_d`       | `agg.Rect`                            | Root-level integer rectangle helper                              |
| point-like `double x, y` pairs | `agg.Point` / `agg.PointI`            | Explicit structs in public API                                   |
| `agg::trans_affine`            | `internal/transform.TransAffine`      | Public wrapper helpers in `agg.Transformations` and `agg.Affine` |
| `agg::trans_perspective`       | `internal/transform.TransPerspective` | Internal package, close header parity                            |
| `agg::trans_bilinear`          | `internal/transform.TransBilinear`    | Internal package, close header parity                            |
| `agg::trans_viewport`          | `internal/transform.TransViewport`    | Internal package, close header parity                            |

### Pipeline package mapping

| C++ header family               | Go package                                          |
| ------------------------------- | --------------------------------------------------- |
| `agg_rasterizer_scanline_aa*.h` | `internal/rasterizer`                               |
| `agg_scanline_*.h`              | `internal/scanline`                                 |
| `agg_renderer_*.h`              | `internal/renderer`                                 |
| `agg_pixfmt_*.h`                | `internal/pixfmt`                                   |
| `agg_span_*.h`                  | `internal/span`                                     |
| `agg_conv_*.h`                  | `internal/conv`                                     |
| `agg_trans_*.h`                 | `internal/transform`                                |
| `agg_font_*` / `agg_gsv_text*`  | `internal/font`, `internal/gsv`, `agg` text helpers |

## Method Translation Rules

The main migration issue is that Go has no overloads.

- C++ getter overloads become `Get...` methods.
  Example: `blendMode() const` -> `GetBlendMode()`
- Overloads that only differ by image destination form use explicit suffixes.
  Example: `transformImage(...)` whole-image form -> `TransformImageSimple(...)`
- Radial-gradient overload families use `...MultiStop` and `...Pos` where needed.
- Public methods take Go slices and structs instead of raw pointers.

## Common Pattern Translations

### 1. Create a render target and clear it

C++ AGG2D:

```cpp
Agg2D g;
g.attach(buf, width, height, stride);
g.clearAll(Agg2D::Color(255, 255, 255));
```

Go, close parity:

```go
buf := make([]byte, width*height*4)
g := agg.NewAgg2D()
g.Attach(buf, width, height, width*4)
g.ClearAll(agg.White)
```

Go, idiomatic:

```go
ctx := agg.NewContext(width, height)
ctx.Clear(agg.White)
```

### 2. Draw a filled and stroked path

C++ AGG2D:

```cpp
g.fillColor(Agg2D::Color(230, 80, 40));
g.lineColor(Agg2D::Color(20, 20, 20));
g.lineWidth(3);
g.resetPath();
g.moveTo(20, 20);
g.lineTo(140, 20);
g.lineTo(80, 110);
g.closePolygon();
g.drawPath(Agg2D::FillAndStroke);
```

Go, close parity:

```go
g.FillColor(agg.NewColorRGB(230, 80, 40))
g.LineColor(agg.NewColorRGB(20, 20, 20))
g.LineWidth(3)
g.ResetPath()
g.MoveTo(20, 20)
g.LineTo(140, 20)
g.LineTo(80, 110)
g.ClosePolygon()
g.DrawPath(agg.FillAndStroke)
```

Go, idiomatic:

```go
ctx.SetColor(agg.NewColorRGB(230, 80, 40))
ctx.BeginPath()
ctx.MoveTo(20, 20)
ctx.LineTo(140, 20)
ctx.LineTo(80, 110)
ctx.ClosePath()
ctx.Fill()

ctx.SetColor(agg.Black)
ctx.SetLineWidth(3)
ctx.Stroke()
```

Note:
`Context` immediate helpers such as `FillRectangle` and `DrawCircle` render
immediately. They are not path-building calls.

### 3. Use transforms

C++:

```cpp
g.resetTransformations();
g.translate(100, 80);
g.rotate(agg::deg2rad(30.0));
g.scale(1.5);
```

Go:

```go
g.ResetTransformations()
g.Translate(100, 80)
g.Rotate(30 * agg.Deg2Rad)
g.Scale(1.5)
```

For reusable transform values in root-level Go code:

```go
tr := agg.Translation(100, 80)
tr.Multiply(agg.RotationDegrees(30))
tr.Multiply(agg.UniformScaling(1.5))
```

### 4. Draw images

C++ AGG2D:

```cpp
g.blendImage(img, x, y, alpha);
g.transformImage(img, dstX1, dstY1, dstX2, dstY2);
```

Go:

```go
g.BlendImage(img, x, y, alpha)
g.TransformImageSimple(img, dstX1, dstY1, dstX2, dstY2)
```

### 5. Text in cgo and no-cgo environments

C++ AGG2D:

- FreeType-backed text is the normal path.

Go:

- `Agg2D.Font(...)` is the closest equivalent when built with FreeType support
- `Agg2D.FontGSV(...)` is the no-cgo / WASM-safe vector-font fallback
- `Context.DrawString(...)` style convenience should be treated as wrapper API,
  not one-to-one C++ parity

## Build And Integration Notes

### Module dependency

Add the module in the usual Go way:

```bash
go get github.com/MeKo-Christian/agg_go
```

Basic library build:

```bash
go build ./...
```

Repo-local orchestration:

```bash
just build
just test
```

### Optional build tags

The core library is usable without platform-window tags.

- `-tags freetype`
  Enables the FreeType-backed font engine in `internal/font/freetype` and
  `internal/font/freetype2`
- `-tags sdl2`
  Enables the SDL2 platform backend and related examples
- `-tags x11`
  Enables the X11 platform backend and related examples

Examples:

```bash
go test -tags freetype ./...
go run -tags sdl2 examples/platform/sdl2/main.go
go run -tags x11 examples/platform/x11/main.go
```

### WASM

The repository contains a dedicated WASM demo under `cmd/wasm/`.

Preferred repo workflow:

```bash
just build-wasm
just serve-web
```

The WASM environment does not use cgo/FreeType. For text or demo code that
must stay portable there, prefer the built-in GSV/vector-font path.

### Cross-compilation

The rendering core is Go code and cross-compiles normally for non-tagged
library use. Optional backends are tag-gated:

- avoid `sdl2` and `x11` tags unless those target dependencies exist
- enable `freetype` only when the target build environment provides FreeType

### Internal package usage

If you are porting low-level AGG code rather than `Agg2D`, map the original
headers directly to `internal/*` packages and keep the original pipeline order:

1. `internal/path` or source vertices
2. `internal/transform`
3. `internal/conv`
4. `internal/rasterizer`
5. `internal/scanline`
6. `internal/renderer`
7. `internal/pixfmt`

That route is appropriate for faithful ports of AGG examples and algorithms,
but not for ordinary application code.

## Practical Recommendation

For straight ports from `agg2d.h`, start with `agg.Agg2D` and the
[../AGG2D_PARITY.md](../AGG2D_PARITY.md) mapping.

For new Go applications, start with `agg.Context`, and only drop to `Agg2D` or
`internal/*` packages when you need the original AGG structure or specific
pipeline control.
