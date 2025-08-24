# Repository Guidelines

This document describes how to work in this repository efficiently and consistently.

## Project Structure & Module Organization

- Public API: `agg.go`, `types.go`.
- Implementation: `internal/<pkg>/` (e.g., `basics`, `color`, `buffer`, `pixfmt`, `scanline`, `rasterizer`, `renderer`).
- Examples: `examples/<group>/<name>/` (e.g., `examples/basic/hello_world`).
- Tests: `tests/{unit,integration,benchmark,visual}`.
- Docs: `docs/` for architecture and status notes.

## Build, Test, and Development Commands

- Build library and examples: `just build`.
- Run tests: `just test` | unit only: `just test-unit` | integration: `just test-integration`.
- Coverage report: `just test-coverage` (opens/produces `coverage.html`).
- Run an example: `just run-example basic/hello_world`.
- Quality checks: `just fmt` | `just vet` | `just lint` | `just tidy` | `just check`.
- Install pre-commit hook: `just init-hooks` (runs `just pre-commit` on commit).

## Coding Style & Naming Conventions

- Formatting: `gofumpt` + `gci` via `treefmt`. Enforce with `just lint` or `treefmt --fail-on-change`.
- Linting: `golangci-lint` (includes `misspell`, `gocritic`). Keep code idiomatic and simple.
- Packages: short, lowercase, no underscores.
- Files: `snake_case.go`; tests as `*_test.go`.
- Identifiers: Exported `PascalCase`, unexported `camelCase`. Avoid API stutter.

## Testing Guidelines

- Framework: standard `testing` (`TestXxx`, `BenchmarkXxx`).
- Location: unit tests beside code (`internal/<pkg>`) or cross-package in `tests/unit`.
- Other suites: `tests/{integration,visual,benchmark}`.
- Practices: deterministic tests; isolate state; avoid global mutation.
- Coverage: `just test-coverage` then open `coverage.html`.

## Commit & Pull Request Guidelines

- Commits: short, imperative subject (e.g., `rasterizer: fix cell merge`).
- PRs: focused scope; include description, linked issue (and `TASKS.md` items when relevant), and before/after images for visual changes.
- CI locally: ensure `just check` passes; update docs/examples when API changes.

## Security & Configuration Tips

- No manual memory management; avoid `unsafe` unless justified and reviewed.
- Keep `go.mod` tidy with `just tidy`.
- Use platform-specific builds only when necessary (`build-*-platforms`).
- Prefer proper interfaces and generics over duck typing with interface{}!

## C++ to Go Translation Strategy

The original code can be found in ../agg-2.6 and the source code in particular is located at ../agg-2.6/agg-src/. If possible always refer to the original C++ implementation for guidance. If not otherwise denoted below, try to be as close as possible to the original source code.

- **Templates → Generics**: C++ template classes become Go generic types (e.g., `Point[T]`, `Rect[T]`)
- **Manual Memory → GC**: Replaces C++ new/delete with Go's garbage collector
- **Inheritance → Interfaces**: C++ virtual methods become Go interfaces
- **Enums → Typed Constants**: C++ enums become Go typed constants
- **Avoid using interface{} (or any)**: Prefer constrained generics and explicit interfaces that model the required capabilities at compile time.

## Rendering Pipeline Flow

1. **Path Definition** (`types.go`: Path, MoveTo, LineTo)
2. **Transformation** (`internal/transform/`: affine matrices)
3. **Conversion** (`internal/conv/`: stroke, dash, contour)
4. **Rasterization** (`internal/rasterizer/`: vector → coverage data)
5. **Scanline Generation** (`internal/scanline/`: horizontal strips)
6. **Pixel Rendering** (`internal/renderer/` + `internal/pixfmt/`: final output)

## Memory Management

- Use Go slices instead of C++ raw pointers
- Rendering buffers wrap slices with bounds checking
- Reuse slices where possible for performance (scanlines, spans)
- No manual allocation/deallocation needed

## Error Handling

- Return errors for I/O operations and invalid parameters
- Panic for programmer errors (bounds violations, invalid state)
- Graceful degradation for edge cases in rendering

The codebase follows the detailed porting plan in docs/TASKS.md which lists every C++ file that needs Go implementation. Always mark completed tasks as done ("[x]") once completed.
