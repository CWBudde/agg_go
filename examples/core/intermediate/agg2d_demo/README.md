# AGG2D Demo

This example demonstrates the high-level AGG2D interface in Go, ported from the original C++ `agg2d_demo.cpp`.

## Overview

The AGG2D interface provides a simplified, high-level API for 2D graphics rendering that abstracts away the complexity of the lower-level AGG components. This demo showcases the complete feature set of the AGG2D API, including:

- **Viewport Transformations**: Coordinate mapping and scaling
- **Text Rendering**: With different alignments (left/center/right, top/middle/bottom)
- **Gradient Effects**: Linear and radial gradients for complex fills
- **Shape Drawing**: Rounded rectangles, ellipses, and complex paths
- **Path Operations**: Arc commands, relative path movements
- **Blend Modes**: Add, Overlay, Alpha blending for visual effects
- **Master Alpha**: Global transparency control
- **Image Transformations**: Transforming images along arbitrary paths

## Features Demonstrated

### 1. Viewport and Coordinate Systems

```go
agg2d.Viewport(0, 0, 600, 600, 0, 0, float64(width), float64(height), agg.XMidYMid)
```

Maps world coordinates (0-600) to screen coordinates while preserving aspect ratio.

### 2. Text Rendering and Alignment

The demo shows text rendering with both raster and vector fonts:

- **Raster fonts**: Fast rendering but cannot be rotated
- **Vector fonts**: Can be rotated and scaled, support outlines

All nine text alignment combinations are demonstrated:

- Left/Center/Right horizontal alignment
- Top/Middle/Bottom vertical alignment

### 3. Gradient-Filled UI Elements

Creates realistic "Aqua" buttons using linear gradients:

- Multi-layer gradients for depth effects
- Normal and pressed button states
- Rounded rectangle shapes with different corner radii

### 4. Complex Path Operations

Demonstrates advanced path construction with:

- Relative arc commands (`ArcRel`)
- Horizontal/vertical line segments (`HorLineRel`, `VerLineRel`)
- Cubic Bézier curves (`CubicCurveTo`)
- Path closing and filling/stroking

### 5. Blend Modes

Shows different compositing operations:

- **Add**: Brightens overlapping areas
- **Overlay**: Complex color blending
- **Alpha**: Standard alpha blending

### 6. Image Transformations

Transforms raster images along arbitrary vector paths, demonstrating the integration between vector graphics and raster image processing.

## Building and Running

```bash
# From the agg2d_demo directory
go build .
./agg2d_demo
```

This will generate `agg2d_demo.png` showing all the demonstrated features.

## Code Structure

The demo is organized into logical sections:

1. **Setup**: Context creation and viewport configuration
2. **Text Examples**: Different fonts and alignment options
3. **Gradient Buttons**: Multi-layer gradient effects
4. **Shape Drawing**: Basic geometric shapes
5. **Path Operations**: Complex vector paths with arcs
6. **Blend Mode Examples**: Different compositing modes
7. **Image Processing**: Raster image transformations

## Technical Notes

### Font Handling

The demo attempts to load Arial font but gracefully handles font loading failures by printing warnings and continuing execution. In production, you would typically:

1. Bundle specific font files with your application
2. Use system font discovery mechanisms
3. Provide fallback fonts

### Color Management

The demo uses various color specifications:

- Named colors (`agg.Black`, `agg.White`)
- RGB values (`agg.RGB(r, g, b)`) with 0-1 range
- RGBA values (`agg.RGBA(r, g, b, a)`) with 0-255 range

### Performance Considerations

- Raster fonts are faster for static text
- Vector fonts enable rotation and scaling effects
- Master alpha affects all subsequent drawing operations
- Complex paths with many arc segments can be expensive

## Original Source

This demo is ported from the C++ AGG library's `agg2d_demo.cpp`, maintaining functional equivalence while adapting to Go's idioms and type system.

## See Also

- **Basic Examples**: Start with simpler examples in `examples/core/basic/`
- **AGG Documentation**: Refer to `docs/` for detailed API documentation
- **Original AGG**: C++ source available in the parent AGG 2.6 distribution
