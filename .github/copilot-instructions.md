# GitHub Copilot Instructions for AGG Go Port

## Big Picture Architecture

This is a Go port of the Anti-Grain Geometry (AGG) 2.6 C++ library. It implements a high-quality 2D graphics rendering pipeline with sub-pixel accuracy and anti-aliasing.

### Rendering Pipeline Flow

1.  **Path Definition**: `types.go` (Path, MoveTo, LineTo)
2.  **Transformation**: `internal/transform/` (Affine matrices)
3.  **Conversion**: `internal/conv/` (Stroke, dash, contour)
4.  **Rasterization**: `internal/rasterizer/` (Vector to coverage data)
5.  **Scanline Generation**: `internal/scanline/` (Horizontal strips)
6.  **Pixel Rendering**: `internal/renderer/` + `internal/pixfmt/` (Final output)

### Structural Decisions

- **Public API**: Located at the root (`agg.go`, `types.go`, `context.go`). This is the only user-facing surface.
- **Internal Implementation**: All complexity is hidden in `internal/` packages.
- **Translation Strategy**:
  - C++ Templates → Go Generics (e.g., `Point[T]`, `Rect[T]`, `PodArray[T]`).
  - C++ Inheritance → Go Interfaces.
  - C++ Enums → Typed Constants.
  - Manual Memory → Go Garbage Collector (use slices, avoid `unsafe`).

## Critical Developer Workflows

Use `just` (Justfile) for all common tasks:

- `just quick`: Fast feedback (fmt + vet). Run this frequently.
- `just check`: Full validation (fmt, vet, lint, tidy, tests). Run before committing.
- `just test`: Run all unit and integration tests.
- `just test-coverage`: Generate and view coverage reports.
- `just run-hello`: Run the basic "Hello World" example.
- `just run-example <group>/<name>`: Run a specific example (e.g., `just run-example core/basic/shapes`).

## Project-Specific Conventions

- **Naming**: Files use `snake_case.go`. Exported identifiers use `PascalCase`, unexported use `camelCase`.
- **Type Safety**: Avoid `any` or `interface{}`. Prefer constrained generics and explicit interfaces.
- **Error Handling**: Return `error` for I/O or invalid parameters. Use `panic` for programmer errors (e.g., out-of-bounds, invalid state).
- **Performance**: Reuse slices and buffers where possible (e.g., scanlines, spans) to minimize allocations.
- **C++ Reference**: Always refer to the original C++ source in `../agg-2.6/agg-src/` when implementing or fixing algorithms to ensure parity.

## Key Files & Patterns

- **Entry Point**: `agg2d.go` provides the high-level `Agg2D` context.
- **Core Types**: `types.go` defines fundamental types like `Point`, `Rect`, and `Color`.
- **Interface Contracts**: `internal/array/interfaces.go` and `internal/font/interfaces.go` define key abstractions.
- **Testing**: Unit tests should be placed alongside the code in `*_test.go` files. Use `tests/` for integration and visual tests.
