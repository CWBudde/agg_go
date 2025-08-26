# Rounded Rectangle Example

This example demonstrates AGG's rounded rectangle rendering capabilities, corresponding to the original AGG `rounded_rect.cpp` demo.

## What This Example Does

- Creates a 640x480 canvas with white background
- Renders five different rounded rectangles with varying:
  - Fill vs outline styles
  - Colors (blue, red, green, purple, orange)
  - Dimensions and corner radius values
  - Shape proportions (including pill-shaped)
- Saves output as PNG file

## AGG Concepts Demonstrated

- **Rounded Rectangle Geometry**: Creating rounded rectangles with configurable corner radius
- **Fill vs Stroke Rendering**: Comparing filled shapes (`FillRoundedRectangle`) vs outlined shapes (`DrawRoundedRectangle`)
- **Color Management**: Using both predefined colors and custom RGB values
- **High-Level API**: Utilizing AGG's simplified context-based interface
- **Image Output**: Converting AGG image data to standard PNG format

## Output

The example generates `rounded_rect_demo.png` showing:

- Blue filled rounded rectangle
- Red outlined rounded rectangle
- Green filled rounded rectangle (different proportions)
- Purple outlined rounded rectangle
- Orange filled pill-shaped rounded rectangle

## How to Run

```bash
cd examples/core/basic/rounded_rect
go run main.go
```

## Expected Output

- Console output describing the rendering process
- PNG file `rounded_rect_demo.png` showing five rounded rectangles
- Summary of what each rectangle demonstrates

## Technical Notes

This is currently a simplified version using the high-level context API. The original AGG `rounded_rect.cpp` includes:

- Interactive mouse controls for rectangle manipulation
- Real-time radius adjustment with sliders
- Subpixel positioning demonstrations
- Full scanline rasterizer pipeline usage

## Relationship to Original AGG

Corresponds to `examples/rounded_rect.cpp` in the original AGG 2.6 C++ library. The C++ version includes interactive controls and demonstrates more advanced rasterization techniques.

## Related Examples

- [shapes](../shapes/) - More shape types
- [aa_demo](../../tests/aa_demo/) - Anti-aliasing demonstration
- [rasterizers](../../intermediate/rasterizers/) - Low-level rasterization
