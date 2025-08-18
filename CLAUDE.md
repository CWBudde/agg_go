# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go port of the Anti-Grain Geometry (AGG) 2.6 C++ library - a high-quality 2D graphics rendering library with anti-aliasing capabilities. The project is in early development with foundational components implemented.

## Development Commands

This project uses [Just](https://github.com/casey/just) for build orchestration. Install with `cargo install just` or your package manager.

```bash
# Show all available commands
just --list

# Essential development workflow
just check          # Run fmt, vet, lint, tidy, and tests
just quick           # Fast feedback: fmt + vet only
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

**Common Workflows:**

- Development: `just quick && just test-unit` for fast feedback
- Pre-commit: `just check` runs full validation
- Testing: `just test-coverage` for detailed test analysis
- Examples: `just build-examples && just run-hello` to verify basic functionality

## Architecture & Key Concepts

### Public API Design

- **Clean Interface**: Only `agg.go` and `types.go` are exposed to users
- **Hidden Implementation**: All complexity lives in `internal/` packages
- **User-Focused**: Simple API like `ctx := agg.NewContext(800, 600); ctx.SetColor(agg.Red); ctx.DrawCircle(x, y, r)`

### C++ to Go Translation Strategy

- **Templates → Generics**: C++ template classes become Go generic types (e.g., `Point[T]`, `Rect[T]`)
- **Manual Memory → GC**: Replaces C++ new/delete with Go's garbage collector
- **Inheritance → Interfaces**: C++ virtual methods become Go interfaces
- **Enums → Typed Constants**: C++ enums become Go typed constants

### Rendering Pipeline Flow

1. **Path Definition** (`types.go`: Path, MoveTo, LineTo)
2. **Transformation** (`internal/transform/`: affine matrices)
3. **Conversion** (`internal/conv/`: stroke, dash, contour)
4. **Rasterization** (`internal/rasterizer/`: vector → coverage data)
5. **Scanline Generation** (`internal/scanline/`: horizontal strips)
6. **Pixel Rendering** (`internal/renderer/` + `internal/pixfmt/`: final output)

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

The codebase follows the detailed porting plan in TASKS.md which lists every C++ file that needs Go implementation. Always mark completed tasks as done ("[x]") once completed.

## Current Test Failures (with TODO comments added)

As of the latest test run, several issues remain that require further investigation and fixes:

### Liang-Barsky Clipping Algorithm (internal/basics/)
- **Issue**: Algorithm behavior doesn't match test expectations for edge cases
- **Status**: TODO comments added documenting the discrepancy
- **Next**: Compare with original AGG C++ implementation for expected behavior

### Pixel Blending Mathematics (internal/pixfmt/)
- **Issue**: Premultiplied alpha blending calculations incorrect
- **Status**: Fixed zero-alpha case, but complex blending still fails
- **Next**: Review alpha premultiplication and RGBA8Prelerp implementation

### Rasterizer Clipping Boundary Detection (internal/rasterizer/)
- **Issue**: Boundary detection for lines crossing clip regions
- **Status**: TODO comments added for Cases 8 and 12
- **Next**: Review AGG scanline clipping algorithm implementation

### VCGen State Management (internal/vcgen/)
- **Issue**: B-spline and smooth polygon generators have edge case failures
- **Status**: Fixed several state management issues, added bounds checking
- **Next**: Remaining vertex sequence access panics need fixes

### Converter Integration (internal/conv/)
- **Issue**: Adaptor integration between path converters and vertex generators
- **Status**: Missing integration causing test failures
- **Next**: Implement proper ConvAdaptorVCGen integration

### Missing Dependencies from TASKS.md
Many AGG components are not yet implemented. Key missing dependencies include:
- Advanced curve approximation algorithms
- Complete scanline rendering pipeline components
- Full rasterizer cell storage and sorting
- Gamma correction and color management
- Advanced path stroking and dashing algorithms

For new features, always check TASKS.md for dependency requirements before implementation.
