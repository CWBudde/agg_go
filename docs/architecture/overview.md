# AGG Go Architecture Overview

## Design Philosophy

The AGG Go port maintains the original AGG's core architecture while adapting to Go's idioms and best practices.

### Key Design Decisions

1. **Clean Public API**: All implementation details are hidden in `internal/` packages
2. **Go Generics**: Used to replace C++ templates where appropriate
3. **Interface-Based Design**: Polymorphism through interfaces rather than template specialization
4. **Memory Safety**: Leverages Go's garbage collector instead of manual memory management

## Package Structure

```
agg_go/
├── agg.go, types.go          # Public API (what users import)
├── internal/                 # All implementation details
│   ├── basics/              # Core types, constants, math utilities
│   ├── color/               # Color space handling
│   ├── buffer/              # Memory management for pixel data
│   ├── pixfmt/              # Pixel format implementations
│   ├── scanline/            # Scanline generation and storage
│   ├── rasterizer/          # Path rasterization
│   ├── renderer/            # Final rendering pipeline
│   ├── geometry/            # Primitive shapes
│   ├── curves/              # Bezier curves and path storage
│   ├── transform/           # Affine transformations
│   ├── conv/                # Path converters (stroke, dash, etc.)
│   ├── span/                # Span generation and gradients
│   └── image/               # Image processing
├── examples/                # Progressive examples
└── docs/                    # Documentation
```

## Rendering Pipeline

The AGG rendering pipeline follows this flow:

1. **Path Definition**: User creates paths with MoveTo, LineTo, curves
2. **Transformation**: Apply affine transformations
3. **Conversion**: Convert paths (stroke, dash, contour)
4. **Rasterization**: Convert vector paths to coverage data
5. **Scanline Generation**: Create horizontal scanlines
6. **Rendering**: Apply colors and blend to pixel buffer

## Template → Generic Translation

AGG's heavy use of C++ templates is translated to Go using:

- **Generics** for type-parameterized containers and algorithms
- **Interfaces** for polymorphic behavior
- **Concrete types** for common use cases to avoid generic overhead

## Memory Management

- **Buffer Management**: Rendering buffers use slices with bounds checking
- **Path Storage**: Dynamic arrays for path commands and vertices
- **Scanline Storage**: Pooled allocators for temporary scanline data
- **Garbage Collection**: No manual memory management required

## Error Handling

- Go's standard error handling for I/O operations
- Panic for programmer errors (bounds violations, invalid state)
- Graceful degradation for rendering edge cases

## Performance Considerations

- **Hot Path Optimization**: Critical rendering loops avoid allocations
- **Slice Reuse**: Scanline and span slices are reused where possible
- **SIMD-Ready**: Structure allows for future SIMD optimizations
- **Concurrency**: Rendering pipeline can be parallelized for large images
