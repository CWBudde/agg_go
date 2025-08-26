# Anti-Aliasing Demo

This example demonstrates AGG's anti-aliasing quality and capabilities, corresponding to the original AGG `aa_demo.cpp` educational demo.

## What This Example Does

- Creates a 400x300 test canvas for anti-aliasing quality assessment
- Runs five comprehensive anti-aliasing test scenarios:
  1. **Radial Lines**: 16 lines at different angles from center point
  2. **Variable Stroke Width**: Circles with increasing stroke thickness
  3. **Diagonal Lines**: Critical 45° and shallow angle line tests
  4. **Filled Shapes**: Triangle and ellipse with smooth edges
  5. **Sub-pixel Positioning**: Lines offset by quarter-pixel increments
- Saves output as PPM image for visual inspection
- Provides detailed test result summary

## AGG Concepts Demonstrated

- **Anti-Aliasing Quality**: Smooth edge rendering without jagged pixels
- **Sub-pixel Accuracy**: Positioning precision beyond pixel boundaries
- **Stroke Width Handling**: Variable line thickness with consistent quality
- **Filled Shape Edges**: Smooth boundaries for complex polygons
- **Alpha Blending**: Semi-transparent shapes with proper edge composition
- **Critical Angle Testing**: Diagonal lines that reveal aliasing artifacts

## How to Run

```bash
cd examples/tests/aa_demo
go run main.go
```

## Expected Output

- Console output showing progression through five test scenarios
- PPM image file `examples/shared/art/aa_demo_output.ppm`
- Comprehensive test result summary with checkmarks for each scenario

## Visual Assessment Criteria

When examining the output image, look for:

- **Smooth line edges** without stair-step artifacts
- **Consistent stroke quality** across different widths
- **Clean diagonal lines** especially at 45° angles
- **Smooth filled shape boundaries** without rough edges
- **Sub-pixel precision** in closely spaced vertical lines

## Technical Details

This example demonstrates the core of AGG's value proposition - high-quality anti-aliasing through:

- **Scanline rasterization** with coverage calculation
- **Sub-pixel sampling** for edge smoothness
- **Alpha blending** for smooth color transitions
- **Geometric precision** in coordinate handling

## Relationship to Original AGG

Corresponds to `examples/aa_demo.cpp` in the original AGG 2.6 C++ library. The C++ version includes:

- Interactive pixel magnification for close inspection
- Side-by-side aliased vs anti-aliased comparison
- Custom enlarged renderer for educational visualization
- Mouse interaction for detailed examination

## Comparison with Other Libraries

AGG's anti-aliasing is designed to be superior to:

- Simple pixel-based rendering (aliased)
- Basic multi-sampling approaches
- Software renderers without geometric precision

## Related Examples

- [circles](../circles/) - Circle rendering quality
- [rasterizers](../core/intermediate/rasterizers/) - Low-level rasterization control
- [line_thickness](../../core/basic/lines/) - Line rendering precision
