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
- Docs: See `PLAN.md` for the unified roadmap and `docs/` for architecture and supporting notes.

References:

- Unified roadmap and phase plan: `PLAN.md`
- Implementation roadmap: `docs/TASKS.md`
- Completed items: `docs/TASKS-COMPLETED.md`

## Web Demo

A live web demo of AGG Go, compiled to WebAssembly (WASM), is available. It showcases various rendering features directly in your browser.

- **Live Demo:** [https://christian-schlichtherle.github.io/agg-go/](https://christian-schlichtherle.github.io/agg-go/)
- **Source:** `cmd/wasm/main.go` and `web/`

You can also run the demo locally:

```bash
just serve-web
```

## Repository Structure

- Public API: `agg.go` plus high-level helpers such as `colors.go`, `geometry.go`, `context.go`, `images.go`, and `transforms.go`.
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

Minimal example using the high-level `Context` API:

```go
package main

import (
	"log"

	agg "github.com/MeKo-Christian/agg_go"
)

func main() {
	ctx := agg.NewContext(800, 600)
	ctx.Clear(agg.White)

	ctx.SetColor(agg.NewColor(220, 70, 50, 255))
	ctx.FillRectangle(80, 80, 220, 140)

	ctx.SetColor(agg.NewColor(30, 90, 180, 255))
	ctx.SetLineWidth(6)
	ctx.DrawCircle(500, 280, 90)

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

For a fuller walkthrough, see [docs/guides/getting-started.md](docs/guides/getting-started.md).
You can also run `just run-example basic/hello_world`.

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

1. Path definition (`context.go` + `Agg2D` path methods)
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

Follow `PLAN.md` for the consolidated roadmap and phased backlog, `docs/TASKS.md` for the detailed file-level implementation inventory, and `docs/TASKS-COMPLETED.md` for completed items. Please keep these up to date during development.

## License

This is an in‑progress port of AGG 2.6 for Go. Licensing will follow the original AGG licensing terms; see upstream AGG for details. This repository may be consolidated into AGoGo in the future.
