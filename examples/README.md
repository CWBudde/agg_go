# AGG Go Examples

This directory contains progressive examples demonstrating the AGG Go library capabilities, organized for clear learning progression and reliable builds.

## Structure

### Core Examples (`core/`)

Core examples use only the internal Go implementation and are guaranteed to build without external dependencies.

#### Basic Examples (`core/basic/`)
- **hello_world/** - Simple context creation and basic drawing
- **shapes/** - Drawing basic geometric shapes (circles, rectangles, lines)
- **colors_gray/** - Grayscale color handling
- **colors_rgba/** - RGBA color manipulation and blending
- **lines/** - Line drawing with different styles
- **rounded_rect/** - Rounded rectangle shapes
- **embedded_fonts_hello/** - Basic text rendering with embedded fonts

#### Intermediate Examples (`core/intermediate/`)
- **gradients/** - Gradient fills and color interpolation
- **controls/** - Interactive UI controls (sliders, checkboxes, etc.)
- **text_rendering/** - Advanced text rendering features
- **paths/** - Complex path operations and curves
- **transforms/** - Affine transformations and coordinate systems

#### Advanced Examples (`core/advanced/`)
- **advanced_rendering/** - Complex rendering techniques
- **image_filters/** - Image processing and filtering
- **custom_renderer/** - Custom rendering pipelines
- **performance/** - Optimization techniques and benchmarks

### Platform Examples (`platform/`)

Platform-specific backends with external dependencies. These examples use build tags for optional compilation.

#### SDL2 Backend (`platform/sdl2/`)
- Interactive graphics applications using SDL2
- Requires: `go get github.com/veandco/go-sdl2/sdl`
- Build: `go build -tags sdl2`

#### X11 Backend (`platform/x11/`)  
- Native X11 windowing system integration
- Requires: X11 development headers (libx11-dev)
- Build: `go build -tags x11`

### Test Examples (`tests/`)

Test examples that mirror the original AGG 2.6 C++ examples for compatibility verification.

- **circles/** - Circle rendering tests
- **aa_demo/** - Anti-aliasing demonstration  
- **rounded_rect/** - Rounded rectangle tests
- **lines/** - Line rendering tests
- **gradients/** - Gradient rendering tests
- **blur/** - Blur effect tests
- **alpha_mask/** - Alpha masking tests
- **perspective/** - Perspective transformation tests

### Shared Resources (`shared/`)

- **art/** - Images, fonts, and test data from AGG 2.6
- **utils/** - Common utilities for examples

## Current Status

### ✅ Available

- **Core examples structure**: All basic, intermediate, and advanced examples are organized and buildable
- **Platform examples**: SDL2 and X11 demos with proper build tags
- **Build system**: Updated Justfile with commands for new structure
- **Shared resources**: Art assets from original AGG 2.6

### 🚧 In Development

Many examples require completion of core rendering pipeline components. See `TEST_TASKS.md` for detailed status.

### Running Examples

Use the improved build system with `just`:

```bash
# Build all examples (core + platform with graceful dependency handling)
just build-examples

# Run core examples by category
just run-examples-basic          # All basic examples
just run-examples-intermediate   # Intermediate examples  
just run-examples-advanced       # Advanced examples

# Run specific examples
just run hello_world            # Basic example shortcut
just run-hello                  # Alias for hello_world
just run-example core/basic/shapes  # Full path

# Run platform examples (with dependencies)
just run-sdl2-demo             # SDL2 interactive demo
just run-x11-demo              # X11 interactive demo

# Run test examples (AGG 2.6 compatibility)
just run-tests                 # All test examples

# Development helpers
just build-example core/basic/hello_world  # Build specific example
just stats                     # Show project statistics
```

### Build Tags and Dependencies

Platform examples use Go build tags to handle optional dependencies gracefully:

- **SDL2**: `//go:build sdl2` - requires `github.com/veandco/go-sdl2/sdl`
- **X11**: `//go:build x11` - requires X11 development headers
- **Core**: No build tags - pure Go implementation

The build system automatically handles missing dependencies by showing informative messages rather than failing.

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
