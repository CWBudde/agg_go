# Text Rendering

This guide describes the current text-rendering surface of the Go port and how
it maps to the original AGG2D text API in
`../agg-2.6/agg-src/agg2d/agg2d.h`.

The canonical AGG2D text methods are still:

- `Font`
- `FontHeight`
- `FlipText`
- `TextAlignment`
- `TextHints`
- `TextWidth`
- `Text`

The root `Context` API wraps the same functionality with a few convenience
helpers such as `DrawText`, `MeasureText`, and `SetTextAlignment`.

## Text backends

There are currently two ways to render text:

### 1. FreeType backend

This is the closest match to the original AGG2D font workflow.

- Build with `-tags freetype`
- Load a font file with `Font(...)`
- Render text through glyph caches
- Use `RasterFontCache` or `VectorFontCache`

### 2. GSV stroke-font fallback

This is a Go-only fallback exposed as `FontGSV(height)`.

- no external font file required
- useful in WASM or no-FreeType builds
- intended for simple/demo text rendering
- not a full replacement for a real TrueType/OpenType font backend

## Build requirements

For the FreeType-backed path:

```bash
# Linux
sudo apt-get install libfreetype6-dev

# macOS
brew install freetype

# Build or run with FreeType support
go build -tags freetype
```

Without the `freetype` build tag, `Font(...)` will fail and the FreeType text
path is unavailable. In that case, use `FontGSV(...)` if you need a built-in
fallback.

## Agg2D usage

This is the direct AGG2D-style workflow:

```go
agg2d := agg.NewAgg2D()
agg2d.Attach(buf, width, height, stride)

if err := agg2d.Font("/path/to/font.ttf", 16.0, false, false, agg.RasterFontCache, 0.0); err != nil {
	log.Fatal(err)
}

agg2d.FillColor(agg.Black)
agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)
agg2d.Text(400, 300, "Hello, AGG2D", false, 0, 0)
```

That matches the original AGG2D model closely: load a font, configure fill
color and alignment, then call `Text`.

## Context usage

The high-level `Context` wrapper exposes the same pipeline in a more Go-shaped
form:

```go
ctx := agg.NewContext(800, 600)
ctx.Clear(agg.White)

if err := ctx.Font("/path/to/font.ttf", 18.0, false, false, agg.RasterFontCache, 0.0); err != nil {
	log.Fatal(err)
}

ctx.SetColor(agg.Black)
ctx.SetTextAlignment(agg.AlignCenter, agg.AlignCenter)

if err := ctx.DrawText("Hello, Context", 400, 300); err != nil {
	log.Fatal(err)
}
```

## Font loading semantics

### `Font(fileName, height, bold, italic, cacheType, angle)`

This is the direct counterpart of C++ `Agg2D::font(...)`.

- `fileName`: font path on disk
- `height`: requested font height in world units
- `bold`: currently not synthesized by the Go backend
- `italic`: currently not synthesized by the Go backend
- `cacheType`: `RasterFontCache` or `VectorFontCache`
- `angle`: stored as part of text state; the current public API typically uses transforms for rotation

Important current behavior:

- `RasterFontCache` configures the backend in screen-space units, matching AGG2D
  bitmap glyph behavior.
- `VectorFontCache` keeps glyph outlines in world-space units.
- If FreeType is unavailable, `Font(...)` returns an error.

## Alignment semantics

`TextAlignment(alignX, alignY)` matches the original AGG2D alignment model.

Horizontal alignment:

- `AlignLeft`
- `AlignCenter`
- `AlignRight`

Vertical alignment:

- `AlignBottom`
- `AlignCenter`
- `AlignTop`

Example:

```go
agg2d.TextAlignment(agg.AlignLeft, agg.AlignTop)
agg2d.Text(100, 100, "Top left", false, 0, 0)

agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)
agg2d.Text(400, 300, "Centered", false, 0, 0)

agg2d.TextAlignment(agg.AlignRight, agg.AlignBottom)
agg2d.Text(700, 500, "Bottom right", false, 0, 0)
```

## Measuring text

Use `TextWidth` or the `Context` helpers `MeasureText` / `GetTextWidth`.

```go
width := agg2d.TextWidth("Measured text")
```

Current measurement notes:

- kerning is applied using glyph indices, matching the AGG2D kerning model
- if no font backend is active, width is `0`
- actual coverage depends on the loaded font and backend

## Hinting and vertical flip

### `TextHints(bool)`

Enables or disables hinting for the FreeType path.

### `FlipText(bool)`

Controls the Y-direction convention used by the font engine. This exists in the
 original AGG2D API and is still relevant when matching AGG coordinate
 conventions.

## GSV fallback

If you cannot or do not want to depend on FreeType, use the built-in stroke-font
fallback:

```go
agg2d := agg.NewAgg2D()
agg2d.Attach(buf, width, height, stride)
agg2d.FontGSV(24)
agg2d.FillColor(agg.Black)
agg2d.Text(40, 80, "GSV fallback", false, 0, 0)
```

This path is useful for demos and constrained environments, but it should not be
documented as equivalent to a full TrueType/OpenType renderer.

## Current caveats

- `bold` and `italic` parameters are not currently synthesized by the FreeType
  backend.
- text-on-path is not implemented in the public `Context` helper surface
  (`DrawTextOnPath` returns an error).
- text coverage and Unicode support depend on the loaded font; the GSV fallback
  is much more limited than a real font backend.

## Related files and examples

- public wrapper: `text.go`
- AGG2D implementation: `internal/agg2d/text.go`
- example: `examples/core/intermediate/text_rendering/main.go`
- parity reference: [AGG2D_PARITY.md](./AGG2D_PARITY.md)
