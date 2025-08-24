# GEMINI.md

## Project Overview

This project is a Go port of the Anti-Grain Geometry (AGG) library, a high-quality 2D graphics library. The goal of this project is to provide a minimal, idiomatic Go implementation of AGG 2.6, maintaining the core functionality and API structure of the original C++ codebase.

The main package `agg` provides the primary API for creating and drawing on a rendering context. The core data structures include `Context`, `Image`, `Path`, and various `Color` types. The internal packages, such as `internal/basics` and `internal/buffer`, provide the low-level implementations for the public API.

## Building and Running

The project uses a `Justfile` for build orchestration.

### Building

- **Build the library:** `just build-lib`
- **Build all examples:** `just build-examples`
- **Build a specific example:** `just build-example EXAMPLE=<example_name>`

### Running

- **Run a specific example:** `just run-example EXAMPLE=<example_name>`
- **Run the "hello world" example:** `just run-hello`

### Testing

- **Run all tests:** `just test`
- **Run unit tests:** `just test-unit`
- **Run integration tests:** `just test-integration`
- **Run benchmark tests:** `just test-bench`
- **Run visual regression tests:** `just test-visual`
- **Run tests with coverage:** `just test-coverage`

## Development Conventions

### Formatting

- **Format all Go code:** `just fmt`

### Linting

- **Run linters:** `just lint`
- **Fix linting issues:** `just lint-fix`

### Dependencies

- **Tidy dependencies:** `just tidy`

### All Checks

- **Run all checks (fmt, vet, lint, test):** `just check`

### C++ to Go Translation Strategy

The original code can be found in ../agg-2.6 and the source code in particular is located at ../agg-2.6/agg-src/. If possible always refer to the original C++ implementation for guidance. If not otherwise denoted below, try to be as close as possible to the original source code.

- **Templates → Generics**: C++ template classes become Go generic types (e.g., `Point[T]`, `Rect[T]`)
- **Manual Memory → GC**: Replaces C++ new/delete with Go's garbage collector
- **Inheritance → Interfaces**: C++ virtual methods become Go interfaces
- **Enums → Typed Constants**: C++ enums become Go typed constants
- **Avoid using interface{} (or any)**: Prefer constrained generics and explicit interfaces that model the required capabilities at compile time.

### Rendering Pipeline Flow

1. **Path Definition** (`types.go`: Path, MoveTo, LineTo)
2. **Transformation** (`internal/transform/`: affine matrices)
3. **Conversion** (`internal/conv/`: stroke, dash, contour)
4. **Rasterization** (`internal/rasterizer/`: vector → coverage data)
5. **Scanline Generation** (`internal/scanline/`: horizontal strips)
6. **Pixel Rendering** (`internal/renderer/` + `internal/pixfmt/`: final output)

### Memory Management

- Use Go slices instead of C++ raw pointers
- Rendering buffers wrap slices with bounds checking
- Reuse slices where possible for performance (scanlines, spans)
- No manual allocation/deallocation needed

### Error Handling

- Return errors for I/O operations and invalid parameters
- Panic for programmer errors (bounds violations, invalid state)
- Graceful degradation for edge cases in rendering

The codebase follows the detailed porting plan in docs/TASKS.md which lists every C++ file that needs Go implementation. Always mark completed tasks as done ("[x]") once completed.
