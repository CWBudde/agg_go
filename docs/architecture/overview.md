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
├── agg.go           # Core types, Agg2D wrapper, constants
├── colors.go        # Color types and management
├── context.go       # High-level Context API (primary user interface)
├── geometry.go      # Geometric primitives (rectangles, points)
├── transforms.go    # 2D transformations and viewport operations
├── gradients.go     # Gradient creation and management
├── images.go        # Image loading, manipulation, and rendering
├── text.go          # Text rendering and typography
├── stroke.go        # Stroke attributes and line styling
├── blending.go      # Blend modes and alpha compositing
├── fill_rules.go    # Fill rule constants
├── internal/        # All implementation details (hidden from users)
│   ├── agg2d/       # Core AGG2D rendering engine (main orchestrator)
│   ├── array/       # Dynamic arrays and block-allocated containers
│   ├── basics/      # Core types, path commands, math utilities
│   ├── bezierarc/   # Bezier arc approximation
│   ├── buffer/      # Rendering buffer management (RenderingBuffer, RowPtr)
│   ├── color/       # Color space handling and conversion utilities
│   ├── config/      # Build configuration constants
│   ├── conv/        # Path converters: stroke, dash, contour, transform
│   ├── ctrl/        # UI controls for examples (sliders, checkboxes)
│   ├── curves/      # Bezier curves, B-splines, path storage
│   ├── effects/     # High-level effect helpers
│   ├── font/        # Font loading: FreeType2 engine and cache
│   ├── gamma/       # Gamma LUT utilities
│   ├── geometry/    # Low-level geometry helpers
│   ├── glyph/       # Glyph raster and path cache
│   ├── gpc/         # General polygon clipping
│   ├── gsv/         # Built-in GSV stroke-vector font
│   ├── image/       # Image filter LUTs and wrap-mode accessors
│   ├── order/       # Pixel component order types (RGBA, BGRA, etc.)
│   ├── path/        # Path storage: vertex sequences, poly adaptors
│   ├── pixfmt/      # Pixel format implementations (RGBA, RGB, Gray, …)
│   ├── platform/    # Platform backends: SDL2, X11, mock
│   ├── primitives/  # Low-level AA primitive rendering (lines, ellipses)
│   ├── rasterizer/  # Vector → coverage data (cells, scanlines)
│   ├── renderer/    # Scanline and outline renderers
│   ├── scanline/    # Scanline storage, boolean algebra, bin storage
│   ├── shapes/      # Shape generators (arc, ellipse, rounded rect, …)
│   ├── simd/        # SIMD-accelerated pixel operations
│   ├── span/        # Span generators: gradients, image filters, Gouraud
│   ├── transform/   # Affine and non-linear transformations
│   ├── vcgen/       # Vertex generators: stroke, contour, dash, smooth
│   ├── vertex_source/ # Vertex source adaptor utilities
│   └── vpgen/       # Vertex pipeline generators: clip, segmentize
├── examples/        # Example applications
└── docs/            # Documentation
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
- **SIMD**: `internal/simd/` provides runtime-dispatched SSE2/AVX2/NEON paths
  for hot pixel operations (`CopyHline`, `Clear`, `BlendSolidHspan`)
- **Concurrency**: Rendering pipeline can be parallelized for large images
