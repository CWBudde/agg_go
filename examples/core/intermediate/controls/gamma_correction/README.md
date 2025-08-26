# Gamma Correction Example

This example demonstrates AGG's gamma correction control widget and its effects on image rendering.

## What This Example Does

- Creates test images with various patterns for gamma correction demonstration:
  - **Grayscale gradient**: Horizontal gradient to show gamma curve effects
  - **Red gradient**: Vertical red channel gradient
  - **Green diagonal**: Diagonal green gradient pattern
  - **Blue checkerboard**: Alternating blue pattern for edge analysis
- Applies different gamma correction curves to demonstrate visual effects
- Generates multiple output images showing gamma correction results
- Shows the `GammaCtrl` widget functionality for interactive gamma adjustment

## AGG Concepts Demonstrated

- **Gamma Correction**: Mathematical curves for display calibration
- **Color Channel Processing**: Individual RGB channel gamma application
- **Interactive Controls**: Using `GammaCtrl` widget for curve editing
- **Image Processing Pipeline**: Applying corrections to rendered images
- **Visual Calibration**: Understanding gamma effects on different patterns

## Gamma Curves Applied

The example generates images with different gamma settings:

1. **Linear/Identity**: No gamma correction (gamma = 1.0)
2. **Dark/Low Gamma**: Gamma < 1.0 for darker midtones
3. **Bright/High Gamma**: Gamma > 1.0 for brighter midtones
4. **sRGB Standard**: Standard sRGB gamma curve
5. **Custom Curves**: User-defined gamma curves

## Output Files Generated

- `gamma_identity.png` - No correction (linear)
- `gamma_dark.png` - Low gamma (darker)
- `gamma_bright.png` - High gamma (brighter)
- `gamma_srgb.png` - sRGB standard curve
- `gamma_custom1.png` - Custom curve example
- `original.png` - Original test pattern

## How to Run

```bash
cd examples/core/intermediate/controls/gamma_correction
go run main.go
```

## Expected Output

- Console output showing gamma correction process
- Multiple PNG files demonstrating different gamma curves
- Test pattern images showing gamma effects on:
  - Gradient smoothness
  - Color channel balance
  - Edge definition in patterns

## Technical Details

The example demonstrates:

- **Gamma mathematics**: Power law functions for color correction
- **Lookup tables**: Efficient gamma correction implementation
- **Color space conversion**: Linear to gamma-corrected space
- **Pattern generation**: Test images designed to reveal gamma effects

## Visual Assessment

When viewing the output images, observe:

- **Gradient smoothness**: How gamma affects gradient transitions
- **Midtone brightness**: Changes in mid-gray values
- **Color balance**: Effects on individual RGB channels
- **Pattern clarity**: Impact on checkerboard and geometric patterns

## Educational Value

Perfect for understanding:

- Display gamma correction principles
- Visual effects of different gamma curves
- Color space and calibration concepts
- Interactive control widget usage
- Image processing pipeline integration

## Relationship to Original AGG

Corresponds to `examples/gamma_correction.cpp` in the original AGG 2.6 C++ library. The C++ version includes:

- Real-time interactive gamma curve editing
- Mouse-based control point manipulation
- Live preview of gamma effects
- Advanced gamma curve shapes (S-curves, custom points)

## Related Examples

- [gamma_ctrl](../gamma_ctrl/) - Interactive gamma control widget
- [rasterizers](../../rasterizers/) - Gamma in anti-aliasing pipeline
- [aa_demo](../../../tests/aa_demo/) - Anti-aliasing with gamma correction
