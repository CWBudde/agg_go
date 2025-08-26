# Shapes Example

This example demonstrates AGG's basic shape rendering capabilities using the high-level context API.

## What This Example Does

- Creates a 400x300 rendering canvas
- Demonstrates various ellipse rendering techniques:
  - **Filled ellipses** with different colors and dimensions
  - **Outlined ellipses** showing stroke rendering
  - **Multiple shapes** with different aspect ratios
- Saves the output as a PNG image file

## AGG Concepts Demonstrated

- **Ellipse Geometry**: Creating ellipses with configurable width and height
- **Fill vs Stroke**: Comparing filled shapes (`FillEllipse`) vs outlined shapes (`DrawEllipse`)
- **Color Management**: Using predefined colors (Red, Blue, Green, Yellow)
- **Shape Composition**: Combining multiple shapes in a single scene
- **High-Level API**: Utilizing AGG's simplified context interface

## Shapes Rendered

1. **Large Red Ellipse**: Filled ellipse (80x60) at center
2. **Blue Outline**: Ellipse outline (100x80) around the red ellipse
3. **Small Green Ellipse**: Filled ellipse (30x20) in upper left
4. **Yellow Ellipse**: Filled ellipse (25x40) in lower right area

## How to Run

```bash
cd examples/core/basic/shapes
go run main.go
```

## Expected Output

- Console output showing shape creation progress
- PNG file with the rendered shapes
- Demonstration of filled vs outlined rendering

## Technical Details

This example uses AGG's high-level context API:

- `NewContext(width, height)` for canvas creation
- `SetColor()` for color selection
- `FillEllipse()` for filled shape rendering
- `DrawEllipse()` for outline shape rendering

The ellipse rendering uses AGG's geometric ellipse algorithms that provide:

- Perfect mathematical accuracy
- Anti-aliased edges for smooth appearance
- Configurable aspect ratios (width ≠ height)

## Educational Value

This example is ideal for:

- Learning basic AGG shape rendering
- Understanding the difference between fill and stroke operations
- Exploring color management in AGG
- Getting familiar with the high-level context API

## Relationship to Original AGG

While not directly corresponding to a specific original AGG C++ example, this demonstrates the same ellipse rendering capabilities found throughout the AGG library, particularly in interactive demos and shape manipulation examples.

## Related Examples

- [hello_world](../hello_world/) - Even more basic AGG usage
- [rounded_rect](../rounded_rect/) - Rounded rectangle shapes
- [circles](../../tests/circles/) - Circle-specific performance testing
- [aa_demo](../../tests/aa_demo/) - Anti-aliasing quality focus
