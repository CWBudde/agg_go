# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go port of the Anti-Grain Geometry (AGG) 2.6 C++ library - a high-quality 2D graphics rendering library with anti-aliasing capabilities. The project is in intermediate development most internal parts are implemented. There are only slight inconsistencies and bugs, which needs fixing. The examples are yet to be fully ported and tested. The public API is still being finalized and a redesign is planned.

## Development Commands

This project uses [Just](https://github.com/casey/just) for build orchestration. Install with `cargo install just` or your package manager.

```bash
# Show all available commands
just --list

# Essential development workflow
just check           # Run fmt, vet, lint, tidy, and tests
just quick           # Fast feedback: fmt + vet only -> call often!
just build           # Build library and examples
just test            # Run all tests (unit + integration)

# Specific operations
just build-lib       # Build library only
just test-coverage   # Generate coverage report
just run-hello       # Run hello world example
just run-example basic/shapes  # Run specific example

# Development helpers
just fmt             # Format all Go code
just lint            # Run golangci-lint
just tidy            # Clean up dependencies
just clean           # Remove build artifacts

# Advanced workflows
just dev             # Watch files and run checks (requires watchexec)
just ci              # Full CI pipeline (build + test with race detection)
just docs            # Generate API documentation
just stats           # Show project statistics
just todo            # Find TODO/FIXME comments
```

Use the MCP `mcp__codanna*` commands for code navigation and manipulation for Go, wherever this comes handy.

**Common Workflows:**

- Development: `just quick && just test-unit` for fast feedback
- Pre-commit: `just check` runs full validation
- Testing: `just test-coverage` for detailed test analysis
- Examples: `just build-examples && just run-hello` to verify basic functionality

## Architecture & Key Concepts

### Public API Design (yet to be finalized or redesigned)

- **Clean Interface**: Only `agg2d.go` and `types.go` (and other files on root) are exposed to users
- **Hidden Implementation**: All complexity lives in `internal/` packages
- **User-Focused**: Simple API like `ctx := agg.NewContext(800, 600); ctx.SetColor(agg.Red); ctx.DrawCircle(x, y, r)`

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

## Repository Structure

```
agg_go/
├── agg2d*.go              # Public API files (user-facing interface)
├── types.go, context.go   # Core types and context definitions
├── internal/              # Implementation packages (hidden from users)
│   ├── array/            # Dynamic arrays and containers
│   ├── basics/           # Math, clipping, fundamental operations
│   ├── buffer/           # Rendering buffer management
│   ├── color/            # Color space handling and conversions
│   ├── conv/             # Path converters (stroke, dash, contour, etc.)
│   ├── ctrl/             # UI controls for examples
│   ├── curves/           # Curve mathematics (B-splines, etc.)
│   ├── font/             # Font loading and management
│   ├── path/             # Path storage and manipulation
│   ├── pixfmt/           # Pixel format implementations
│   ├── rasterizer/       # Vector to pixel conversion
│   ├── renderer/         # Final pixel rendering
│   ├── scanline/         # Scanline generation and storage
│   ├── span/             # Span generation for gradients/patterns
│   ├── transform/        # Affine transformations
│   ├── vcgen/            # Vertex generators
│   └── ???/              # far more like vpgen, primitives, gsv, gpc, and such (not everything is still in use though)
├── examples/             # Example applications
│   ├── core/basic/       # Simple demos
│   ├── core/intermediate/ # Advanced features
│   └── platform/         # Platform-specific backends
└── tests/                # Test suites
```

### File Naming Scheme

The Go port follows a systematic naming convention derived from the original C++ AGG library:

- **C++ Header → Go Package/File**: `rasterizer_scanline_aa.h` becomes `internal/rasterizer/scanline_aa.go`
- **Package Name Provides Context**: Since Go packages provide namespace, files omit redundant prefixes
  - `internal/conv/stroke.go` not `conv_stroke.go` (package already indicates conversion)
  - `internal/renderer/base.go` not `renderer_base.go` (package already indicates renderer)
  - `internal/vpgen/clip_polygon.go` not `vpgen_clip_polygon.go`
- **Template Suffixes**: C++ templates like `pixfmt_rgba32` become `internal/pixfmt/pixfmt_rgba.go` with generics
- **Public API**: Root-level files like `agg2d.go` provide the clean user interface, hiding internal complexity

This naming scheme maintains traceability to the original C++ source while following Go conventions for package organization and file naming, avoiding redundant prefixes within package directories.

## Development Patterns

### Adding New Internal Components

1. Create package in `internal/componentname/`
2. Use generics for type parameters where AGG used C++ templates
3. Follow Go naming conventions (e.g., `NewRenderingBuffer` not `rendering_buffer_new`)
4. Keep public API surface minimal - implementation details stay internal

### Template Translation Examples

- `pod_array<T>` → `PodArray[T]` with Go generics
- `rgba8T<Colorspace>` → `RGBA8[CS any]` where CS is Linear or SRGB type
- C++ enum → Go typed constants with methods

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
