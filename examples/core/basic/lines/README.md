# Lines Example

This example demonstrates AGG's line rendering capabilities with various orientations and patterns.

## What This Example Does

- Creates a 640x480 rendering canvas
- Draws a comprehensive set of line demonstrations:
  - **Grid pattern**: Light gray horizontal and vertical guide lines
  - **Diagonal lines**: Lines at various angles radiating from center
  - **Radial pattern**: Multiple lines forming a sun-burst pattern
  - **Coordinate axes**: Reference lines for visual guidance
- Saves the output as PNG file

## AGG Concepts Demonstrated

- **Line Geometry**: Drawing lines between two points with pixel precision
- **Color Management**: Using both RGB values and predefined colors
- **Geometric Calculations**: Using trigonometry for radial line patterns
- **Grid Systems**: Creating regular patterns for visual reference
- **High-Level API**: Utilizing `DrawLine()` method from context API

## Line Patterns Shown

1. **Reference Grid**: 40-pixel spaced horizontal and vertical lines
2. **Diagonal Lines**: Lines at different angles from center
3. **Radial Pattern**: Sun-burst effect with lines radiating outward
4. **Various Orientations**: Horizontal, vertical, and diagonal demonstrations

## How to Run

```bash
cd examples/core/basic/lines
go run main.go
```

## Expected Output

- Console output showing line drawing progress
- PNG file `lines_demo.png` with comprehensive line demonstrations
- Visual patterns showing AGG's line rendering quality

## Technical Details

The example demonstrates:

- **Precision**: Sub-pixel accurate line positioning
- **Anti-aliasing**: Smooth line edges without jagged artifacts
- **Performance**: Efficient rendering of multiple line segments
- **Mathematical accuracy**: Correct trigonometric calculations for angles

Uses the high-level context API:

- `DrawLine(x1, y1, x2, y2)` for line segments
- `SetColor()` for line color control
- Coordinate system with (0,0) at top-left

## Mathematical Components

- **Grid generation**: Regular spacing calculations
- **Radial patterns**: Using `math.Sin()` and `math.Cos()` for circular arrangements
- **Angle calculations**: Converting degrees to radians for proper trigonometry

## Educational Value

Perfect for learning:

- Basic line rendering in AGG
- Understanding coordinate systems
- Mathematical pattern generation
- Anti-aliasing quality assessment
- High-level API usage patterns

## Relationship to Original AGG

While not directly corresponding to a specific C++ example, this demonstrates line rendering capabilities that are fundamental to many AGG examples, particularly those involving geometric patterns and precision line work.

## Related Examples

- [aa_demo](../../tests/aa_demo/) - Anti-aliasing focus with line quality
- [rasterizers](../../intermediate/rasterizers/) - Low-level line rendering
- [shapes](../shapes/) - Other geometric primitives
