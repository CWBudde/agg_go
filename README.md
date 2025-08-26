# agg_go

Go port of the Anti-Grain Geometry (AGG) 2.6 library — a high‑quality 2D renderer with anti‑aliasing and sub‑pixel accuracy. The goal is a clean, idiomatic Go API over a faithful implementation of AGG’s core pipeline.

Important: Work in progress. This repository may be deleted or merged into the author’s AGoGo project.

## Overview

- Purpose: Provide an idiomatic Go implementation of AGG 2.6.
- Scope: Core rendering pipeline (paths → rasterizer → scanlines → pixel formats → renderers), examples, and tests.
- API: Minimal, user‑focused surface at the module root (e.g., `agg.NewContext`, color/geometry helpers) wrapping internal packages.

## Status

- In development: Most internals exist; examples and some APIs are stabilizing; minor inconsistencies/bugs expected.
- Public API: Being finalized; a redesign is planned.
- Docs: See `docs/` for architecture, roadmap, and areas in progress.

References:

- Implementation roadmap: `docs/TASKS.md`
- Completed items: `docs/TASKS-COMPLETED.md`
- Example tracking: `docs/TASKS-EXAMPLES.md`

## Repository Structure

- Public API: `agg.go`, `types.go`, plus high‑level helpers (colors, geometry, context).
- Internals (hidden): `internal/<pkg>/` (e.g., `basics`, `pixfmt`, `rasterizer`, `scanline`, `renderer`, `transform`, `conv`).
- Examples: `examples/<group>/<name>/` (e.g., `examples/core/basic/hello_world`).
- Tests: `tests/{unit,integration,benchmark,visual}`.
- Docs: `docs/` (architecture, tasks, status, tutorials).

Run `just stats` for a quick package overview.

## Build & Test

This repo uses Just for orchestration. Install with `cargo install just` or your package manager.

- List commands: `just --list`
- Build library and examples: `just build`
- Tests: `just test` (unit+integration) | unit: `just test-unit`
- Coverage: `just test-coverage` (writes `coverage.html`)
- Quality: `just fmt` | `just vet` | `just lint` | `just tidy` | `just check`

Platform/example tags:

- X11 examples: `go run -tags x11 examples/platform/x11/main.go`
- SDL2 examples: `go run -tags sdl2 examples/platform/sdl2/main.go`

## Quickstart

Minimal “hello world” using the high‑level Context API:

```go
package main

import (
    "fmt"
    agg "agg_go"
)

func main() {
    ctx := agg.NewContext(800, 600)
    ctx.Clear(agg.RGB(0.7, 0.8, 1.0))
    ctx.SetColor(agg.Red)
    ctx.DrawRectangle(100, 100, 200, 150)
    ctx.Fill()
    ctx.SetColor(agg.RGB(0, 0.8, 0))
    ctx.DrawCircle(400, 300, 80)
    ctx.Fill()
    img := ctx.GetImage()
    fmt.Printf("%dx%d image, %d bytes\n", img.Width, img.Height, len(img.Data))
}
```

Or run the example: `just run-hello` or `just run-example core/basic/hello_world`.

## Development Notes

- Coding style: `gofumpt` + `gci` via `treefmt`; check with `just lint`.
- Linting: `golangci-lint` (misspell, gocritic, etc.).
- Packages: short, lowercase; files `snake_case.go`; tests `*_test.go`.
- Errors: return errors for I/O/invalid params; panic on programmer errors (bounds/state).
- Memory: slices, no `unsafe` unless justified; reuse buffers for performance.

### C++ → Go Translation

- Templates → Generics (e.g., `Point[T]`, `Rect[T]`).
- Inheritance → Interfaces; virtuals → explicit interfaces.
- Enums → typed constants.
- Prefer constrained generics; avoid `any`/`interface{}` when possible.

### Rendering Pipeline

1. Path definition (`types.go`)
2. Transformation (`internal/transform`)
3. Conversion (`internal/conv`: stroke, dash, contour)
4. Rasterization (`internal/rasterizer`)
5. Scanlines (`internal/scanline`)
6. Pixel rendering (`internal/renderer` + `internal/pixfmt`)

## Examples

- Basic: `examples/core/basic/*` (hello world, shapes, lines, rounded rect, colors, embedded fonts)
- Intermediate: gradients, text rendering, controls (sliders, rbox, spline), transforms
- Advanced: custom renderers, image filters, performance

Build all: `just build-examples`
Run basic set: `just run-examples-basic`

Note: Some examples require optional dependencies (X11/SDL2, FreeType) and may be stubbed until those are installed.

## Documentation

- Start here: `docs/README.md`
- Architecture and status: `docs/architecture/`, `docs/REVIEW_STATUS.md`
- Text rendering and FreeType notes: `docs/TEXT_RENDERING.md`
- SIMD and performance notes: `docs/SIMD_OPTIMIZATIONS.md`

## Roadmap & Tasks

Follow `docs/TASKS.md` for the full porting plan, `docs/TASKS-COMPLETED.md` for what’s done, and `docs/TASKS-EXAMPLES.md` for example parity. Please keep these up to date during development.

## License

This is an in‑progress port of AGG 2.6 for Go. Licensing will follow the original AGG licensing terms; see upstream AGG for details. This repository may be consolidated into AGoGo in the future.
