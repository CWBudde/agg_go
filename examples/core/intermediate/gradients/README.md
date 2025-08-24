# AGG2D Gradient Support

This directory demonstrates the Phase 4 implementation of AGG2D gradient support, which provides high-level gradient functionality similar to HTML5 Canvas or Cairo graphics libraries.

## Features Implemented

### 1. Color Interpolation

- `Color.Gradient(c2, k)` method for smooth color transitions
- Linear interpolation between two colors with factor k (0.0-1.0)
- Supports RGBA channels including alpha blending

### 2. Linear Gradients

- `FillLinearGradient(x1, y1, x2, y2, c1, c2, profile)` - Fill gradients
- `LineLinearGradient(x1, y1, x2, y2, c1, c2, profile)` - Stroke gradients
- Supports arbitrary direction (horizontal, vertical, diagonal)
- Profile parameter controls transition sharpness (0.0-1.0)

### 3. Radial Gradients

- `FillRadialGradient(x, y, r, c1, c2, profile)` - Fill gradients
- `LineRadialGradient(x, y, r, c1, c2, profile)` - Stroke gradients
- Center-to-edge color transitions
- Variable radius support

### 4. Multi-Stop Radial Gradients

- `FillRadialGradientMultiStop(x, y, r, c1, c2, c3)` - Three-color gradients
- `LineRadialGradientMultiStop(x, y, r, c1, c2, c3)` - Three-color stroke gradients
- Fixed transition points at 50% intervals

### 5. Gradient Position Updates

- `FillRadialGradientPos(x, y, r)` - Update position without changing colors
- `LineRadialGradientPos(x, y, r)` - Update stroke position without changing colors
- Efficient for animations and dynamic positioning

## Usage Examples

### Basic Linear Gradient

```go
agg2d := agg.NewAgg2D()
agg2d.FillLinearGradient(0, 0, 100, 0, agg.Red, agg.Blue, 1.0) // Horizontal red-to-blue
```

### Radial Gradient with Profile

```go
agg2d.FillRadialGradient(50, 50, 25, agg.White, agg.Black, 0.5) // Sharp transition
```

### Multi-Stop Gradient

```go
agg2d.FillRadialGradientMultiStop(50, 50, 30, agg.Red, agg.Green, agg.Blue)
```

## Performance

Based on benchmarks:

- Color interpolation: ~0.13 ns per operation (extremely fast)
- Gradient setup: ~1000 ns per operation (1 microsecond)
- Suitable for real-time applications and animations

## Implementation Notes

### C++ Compatibility

This implementation closely follows the original AGG 2.6 C++ code:

- Same gradient array structure (256-element lookup table)
- Identical profile parameter behavior
- Compatible transformation matrices
- Matching method signatures and semantics

### Gradient Array Structure

- 256-element color lookup table
- Profile parameter controls the distribution of colors within the array
- startGradient = 128 - int(profile \* 127)
- endGradient = 128 + int(profile \* 127)
- Colors outside the profile range are solid (no interpolation)

### Coordinate System

- Uses world coordinates for input
- Automatically converts to screen coordinates for rendering
- Transformation matrices handle rotation and scaling
- Currently assumes 1:1 world-to-screen mapping (will be updated with full rendering pipeline)

## Testing

Comprehensive test suite covers:

- Color interpolation accuracy
- Linear gradient setup and configuration
- Radial gradient setup and configuration
- Multi-stop gradient functionality
- Position update methods
- Profile parameter effects
- Performance benchmarks

Run tests with:

```bash
go test -v -run ".*Gradient.*"
```

Run benchmarks with:

```bash
go test -bench=".*Gradient.*"
```

## Integration Status

✅ **Completed:**

- Core gradient mathematics and color interpolation
- All gradient setup methods
- Comprehensive test coverage
- Performance optimization
- Example demonstration

⏳ **Pending:**

- Integration with the rendering pipeline
- Actual gradient span generation during rendering
- Connection to rasterizer and scanline renderer

## Future Work

The gradient functionality is ready for integration with the rendering pipeline. The next step is to:

1. Connect gradient arrays to span generators in `internal/span/`
2. Integrate with the rasterizer for actual gradient rendering
3. Add gradient-aware rendering paths to the main drawing methods
4. Implement gradient caching for performance optimization

This implementation provides the foundation for full gradient rendering in AGG2D, matching the capabilities of the original C++ library.
