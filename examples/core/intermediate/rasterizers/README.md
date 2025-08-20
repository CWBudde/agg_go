# Rasterizers Example

This example demonstrates the comparison between anti-aliased and aliased (binary) rendering of triangles with gamma correction and transparency controls.

It is a port of the `rasterizers.cpp` example from the original AGG 2.6 library.

## Implementations

This example provides multiple implementations to demonstrate different approaches:

1. **main_direct.go** - A working direct implementation that demonstrates core concepts
2. **main_simple.go** - An attempt to use AGG's scanline system (has interface issues)
3. **main.go** - Full implementation with complete AGG interfaces (has compatibility issues)

**Recommended**: Use `main_direct.go` for a working demonstration of the concepts.

## Features

- **Anti-aliased vs Aliased Rendering**: Shows side-by-side comparison of smooth anti-aliased rendering (left triangle) versus sharp aliased rendering (right triangle)
- **Gamma Correction**: Demonstrates how gamma correction affects the appearance of anti-aliased edges
- **Interactive Controls**: 
  - Gamma slider: Controls the gamma correction value (0.0 to 1.0)
  - Alpha slider: Controls the transparency of both triangles (0.0 to 1.0)
  - Performance test checkbox: Runs a performance comparison between rendering modes
- **Interactive Triangle Manipulation**:
  - Click and drag individual vertices to reshape the triangles
  - Click inside a triangle to move the entire shape
  - Both triangles move together to maintain the comparison

## Key Concepts Demonstrated

1. **Anti-aliased Rendering**: Uses coverage-based anti-aliasing with gamma correction
2. **Aliased Rendering**: Uses binary (on/off) pixel coverage for sharp edges
3. **Gamma Correction**: Shows how different gamma values affect edge appearance
4. **Performance Comparison**: Demonstrates the computational difference between rendering modes

## Usage

### Direct Implementation (Recommended)

```bash
go run main_direct.go
```

This will:
1. Render two triangles side by side using direct rasterization
2. Save demonstration images showing different gamma values
3. Run a performance test comparing aliased vs anti-aliased rendering
4. Output timing information to the console

### AGG Scanline Implementation (Advanced)

```bash
go run main_simple.go  # ✅ Compiles, ❌ Runtime crash (nil pointer dereference)
go run main.go         # ✅ Compiles, ❌ Runtime crash (nil pointer dereference) 
```

**Status Update**: Both implementations now compile successfully after interface compatibility fixes, but both crash at runtime due to uninitialized clipper in the rasterizer constructor.

**Issues Fixed**:
- ✅ Compilation errors in both files resolved
- ✅ Interface incompatibilities bridged with adapter pattern
- ✅ Checkbox method naming corrected

**Remaining Issues**:
- ❌ Runtime crash: `RasterizerScanlineAA` constructor doesn't initialize the `clipper` field
- ⚠️ Architectural: Interface design issues between packages

See [COMPATIBILITY_ISSUES.md](./COMPATIBILITY_ISSUES.md) for detailed analysis and fix recommendations.

## Output Files

### Direct Implementation
- `rasterizers_demo_direct.ppm`: Basic demonstration with default settings
- `rasterizers_gamma_0.1_direct.ppm`: Low gamma (darker edges)
- `rasterizers_gamma_0.5_direct.ppm`: Medium gamma (balanced)
- `rasterizers_gamma_1.0_direct.ppm`: High gamma (lighter edges)

### Other Implementations
- `rasterizers_demo_simple.ppm`: From simple implementation (if it works)
- `rasterizers_demo.ppm`: From full implementation (if it works)

## Technical Details

The example directly implements core AGG concepts:

- **Rasterization**: Converting vector paths to coverage data
- **Scanline Rendering**: Processing horizontal strips of pixels
- **Pixel Format Management**: Handling RGBA pixel data with blending
- **Path Storage**: Managing vector path data and vertex sequences

This is a simplified implementation that focuses on demonstrating the core rendering differences rather than providing a full interactive application.

## Original AGG Correspondence

This Go example corresponds to:
- Original file: `agg-2.6/agg-src/examples/rasterizers.cpp`
- Shows the same visual comparison between anti-aliased and aliased rendering
- Demonstrates identical gamma correction effects
- Maintains the same triangle positioning and color scheme