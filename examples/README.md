# AGG Go Examples

This directory contains progressive examples demonstrating the AGG Go library capabilities.

## Structure

### Basic Examples (`basic/`)

- **hello_world/** - Simple context creation and basic drawing
- **shapes/** - Drawing basic geometric shapes
- **colors/** - Color manipulation and blending

### Intermediate Examples (`intermediate/`)

- **paths/** - Complex path operations and curves
- **transforms/** - Affine transformations and coordinate systems
- **gradients/** - Gradient fills and color interpolation

### Advanced Examples (`advanced/`)

- **image_filters/** - Image processing and filtering
- **custom_renderer/** - Custom rendering pipelines
- **performance/** - Optimization techniques and benchmarks

## Current Status

### ✅ Available

- Basic structure and hello_world skeleton

### 🚧 In Development

All examples require completion of core rendering pipeline:

- Pixel formats
- Scanline generation
- Rasterization
- Rendering

### Running Examples

Once the core is implemented:

```bash
# Basic examples
go run examples/basic/hello_world/main.go
go run examples/basic/shapes/main.go

# Intermediate examples
go run examples/intermediate/paths/main.go

# Advanced examples
go run examples/advanced/performance/main.go
```

## Development Notes

Examples serve as:

1. **API Validation** - Ensure the public API is intuitive
2. **Integration Tests** - Verify components work together
3. **Performance Benchmarks** - Measure rendering performance
4. **Documentation** - Show real-world usage patterns

Each example should include:

- Clear comments explaining the concepts
- Output verification (image generation or console output)
- Performance measurements where relevant
- Error handling demonstrations
