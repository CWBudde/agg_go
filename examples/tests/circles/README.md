# Circles Example

This example demonstrates high-performance circle rendering, corresponding to the original AGG `circles.cpp` performance and quality test.

## What This Example Does

- Creates a 320x200 rendering canvas
- Tests multiple circle rendering scenarios:
  - Solid filled circles in different colors and sizes
  - Outlined circles with stroke width
  - Overlapping translucent circles with alpha blending
- Saves output as PPM image file
- Provides test result summary

## AGG Concepts Demonstrated

- **Ellipse Rendering**: Using AGG's ellipse algorithms for perfect circles
- **Multiple Drawing Modes**: Both filled (`FillCircle`) and outlined (`DrawCircle`) rendering
- **Alpha Blending**: Semi-transparent circles with proper color composition
- **Performance Testing**: Batch rendering of multiple geometric primitives
- **Color Management**: RGB color specification and alpha channel handling
- **Image Output**: Saving rendered graphics in PPM format

## How to Run

```bash
cd examples/tests/circles
go run main.go
```

## Expected Output

- Console output showing test progress through three scenarios
- PPM image file `examples/shared/art/circles_test_output.ppm`
- Test completion summary with success/failure status

## Performance Focus

This example corresponds to AGG's performance-oriented circle demo that typically renders thousands of circles to test:

- Rendering speed and throughput
- Memory efficiency
- Anti-aliasing quality at scale
- Viewport culling optimization

The current Go version demonstrates the basic functionality with a smaller set of circles.

## Technical Details

Uses the AGG2D high-level interface:

- `NewAgg2D()` for context creation
- `Attach()` for buffer management
- `FillCircle()` / `DrawCircle()` for shape rendering
- `SaveImagePPM()` for output

## Relationship to Original AGG

Corresponds to `examples/circles.cpp` in the original AGG 2.6 C++ library. The C++ version includes:

- Performance benchmarking with thousands of circles
- Frame rate measurement and statistics
- Interactive controls for circle count and animation
- Advanced optimization techniques

## Related Examples

- [hello_world](../core/basic/hello_world/) - Basic shape rendering
- [aa_demo](../aa_demo/) - Anti-aliasing quality focus
- [rasterizers](../core/intermediate/rasterizers/) - Low-level rendering pipeline
