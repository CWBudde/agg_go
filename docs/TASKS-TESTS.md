# Visual Testing Framework for AGG Go Library

## Overview

This document outlines a comprehensive visual testing strategy for the AGG Go library. Visual tests validate rendering accuracy by comparing generated images against golden reference images, ensuring pixel-perfect fidelity and preventing visual regressions.

## How Visual Tests Work

Visual tests follow a **golden reference approach**:

1. **Reference Generation**: Known-good images are created manually or from the original C++ AGG library
2. **Test Execution**: Go code generates test images using current implementation
3. **Comparison**: Pixel-by-pixel comparison between generated and reference images
4. **Validation**: Tests pass if images match within acceptable tolerance (typically exact match for most cases, fuzzy matching for anti-aliased edges)
5. **Reporting**: Failed tests generate diff images highlighting discrepancies

## Test Categories and Specifications

### 1. Basic Primitives Testing

**Purpose**: Validate fundamental shape rendering accuracy

**Test Cases**:

- **Lines**: Horizontal/vertical/diagonal, various slopes, sub-pixel positioning, pixel-perfect alignment
- **Circles**: Perfect circles at various radii (1px to 200px), sub-pixel centers, edge anti-aliasing
- **Ellipses**: Various aspect ratios, rotation angles, degenerate cases (very flat/tall)
- **Rectangles**: Axis-aligned, pixel boundaries, sub-pixel positioning, zero-width/height edge cases
- **Rounded Rectangles**: Various corner radii, asymmetric corners, extreme cases (radius > width/height)
- **Polygons**: Regular (triangle to dodecagon), irregular, self-intersecting, concave/convex
- **Arcs**: Full circles, partial arcs, various sweep angles (15° to 359°), start/end alignment

**Expected Output**: 50+ reference images covering all primitive variations

### 2. Path Operations Testing

**Purpose**: Validate path stroking, filling, and conversion accuracy

**Test Cases**:

- **Stroke Styles**: Line widths 0.5px to 20px, sub-pixel widths
- **Line Caps**: Butt, round, square caps on various line angles
- **Line Joins**: Miter, round, bevel joins at various angles, miter limit testing
- **Fill Rules**: Even-odd vs non-zero winding for complex paths, nested shapes
- **Dash Patterns**: Simple (5-5), complex (10-5-2-5), phase offsets, continuation across curves
- **Path Conversion**: B-spline to polyline conversion, curve subdivision accuracy
- **Contours**: Outline generation, offset curves, expansion/contraction

**Expected Output**: 75+ reference images covering stroke and fill variations

### 3. Anti-Aliasing Quality Testing

**Purpose**: Validate sub-pixel accuracy and coverage calculation

**Test Cases**:

- **Sub-pixel Positioning**: Lines/shapes at 0.25, 0.5, 0.75 pixel offsets
- **Coverage Accuracy**: Diagonal lines, circle edges, curve smoothness
- **Gamma Correction**: Linear vs sRGB, gamma values 0.5, 1.0, 1.8, 2.2
- **Pixel Grid Alignment**: On-pixel vs between-pixel positioning
- **Anti-aliased vs Aliased**: Side-by-side comparisons for quality validation
- **Thin Line Quality**: 1px, 0.5px, 0.25px line rendering

**Expected Output**: 30+ reference images focusing on anti-aliasing quality

### 4. Transformations Testing

**Purpose**: Validate coordinate transformation accuracy

**Test Cases**:

- **Affine Transformations**:
  - Translation: Integer and fractional offsets
  - Rotation: 0°, 15°, 30°, 45°, 90°, arbitrary angles
  - Scaling: Uniform (0.5x, 2x, 10x), non-uniform (2x, 0.5x)
  - Skewing: X-axis, Y-axis, combined transformations
- **Perspective**: 3D projection, quadrilateral mapping, keystoning effects
- **Bilinear**: Four-point image distortion
- **Viewport Mapping**: World-to-device coordinate conversion, aspect ratio preservation
- **Combined Transforms**: Rotation+scaling+translation, transform concatenation
- **Transform Precision**: Very small and very large transformation values

**Expected Output**: 40+ reference images covering transformation accuracy

### 5. Color Operations and Blending

**Purpose**: Validate color accuracy and blending mode correctness

**Test Cases**:

- **Color Spaces**: RGB24, RGB32, RGBA32, Gray8, Gray16, various bit depths
- **Blending Modes**: All 16 AGG blend modes:
  - Normal, Multiply, Screen, Overlay, Soft-light, Hard-light
  - Color-dodge, Color-burn, Darken, Lighten, Difference, Exclusion
  - Contrast, Invert, Invert-rgb
- **Alpha Compositing**: Pre-multiplied vs straight alpha, transparency gradients
- **Color Precision**: 8-bit, 16-bit, 32-bit, floating-point color accuracy
- **Alpha Channel**: Opacity levels 0%, 25%, 50%, 75%, 100%

**Expected Output**: 60+ reference images for each supported pixel format

### 6. Gradients and Patterns

**Purpose**: Validate gradient rendering and pattern filling

**Test Cases**:

- **Linear Gradients**: Horizontal, vertical, diagonal, arbitrary angles
- **Radial Gradients**: Circular, elliptical, focal point variations
- **Color Stops**: 2-stop, multi-stop, irregular spacing, color interpolation
- **Gradient Shapes**: Contour-based, diamond, cone, custom gradient functions
- **Pattern Fills**: Image patterns, geometric patterns, tiling modes (repeat, reflect, pad)
- **Pattern Transforms**: Scaling, rotation, translation of pattern fills
- **Gradient Quality**: Smooth transitions, banding prevention

**Expected Output**: 35+ reference images covering gradient varieties

### 7. Text Rendering Testing

**Purpose**: Validate text rendering accuracy and font handling

**Test Cases**:

- **Raster Fonts**: Embedded GSV fonts, bitmap fonts, various sizes (6pt to 72pt)
- **Vector Text**: GSV text rendering, path conversion accuracy
- **Font Metrics**: Character spacing, line height, baseline alignment
- **Text Effects**: Rotation (0°, 45°, 90°, arbitrary), scaling, path-following text
- **Unicode Support**: ASCII, extended ASCII, UTF-8 characters (if supported)
- **Anti-aliasing**: Text smoothing quality, sub-pixel rendering
- **Font Rendering**: Different font sizes, edge sharpness

**Expected Output**: 25+ reference images covering text rendering scenarios

### 8. Clipping Operations

**Purpose**: Validate clipping accuracy and boundary handling

**Test Cases**:

- **Rectangular Clipping**: Various clip bounds, partial shapes, edge alignment
- **Polygon Clipping**: Complex clip shapes, self-intersecting polygons
- **Path Clipping**: Clipping polylines vs polygons, open vs closed paths
- **Multi-level Clipping**: Nested clip regions, boolean operations
- **Clip Precision**: Sub-pixel clip boundaries, anti-aliasing at edges
- **Performance Clipping**: Large scenes with complex clip regions

**Expected Output**: 20+ reference images demonstrating clipping accuracy

### 9. Image Operations

**Purpose**: Validate image filtering and transformation accuracy

**Test Cases**:

- **Image Filtering**: All filter types (nearest, bilinear, bicubic, Lanczos, etc.)
- **Image Transformations**: Rotation, scaling, perspective distortion
- **Image Patterns**: Tiling modes, wrap modes, boundary handling
- **Image Blending**: Alpha channels, color space conversions
- **Filter Quality**: Edge handling, interpolation accuracy, artifact prevention
- **Performance**: Large image processing, memory efficiency

**Expected Output**: 30+ reference images covering image operations

### 10. Advanced Features

**Purpose**: Validate complex mathematical operations and advanced rendering

**Test Cases**:

- **Curve Mathematics**:
  - Bézier curves: Quadratic, cubic, rational
  - B-splines: Various degrees, knot vectors
  - Curve subdivision: Accuracy, smoothness
- **Rasterization Quality**:
  - Complex paths, self-intersecting paths
  - Very large coordinate values
  - Compound paths with multiple sub-paths
- **Mathematical Precision**: Floating-point edge cases, extreme values

**Expected Output**: 25+ reference images for advanced features

### 11. Edge Cases and Stress Tests

**Purpose**: Validate robustness and error handling

**Test Cases**:

- **Degenerate Geometry**: Zero-width/height shapes, coincident points, NaN/Inf values
- **Extreme Values**: Very large/small coordinates (±1e6), extreme transformations
- **Complex Scenes**: Thousands of overlapping objects, deep nesting levels
- **Memory Stress**: Large images (4K, 8K), complex paths with many vertices
- **Precision Edge Cases**: Sub-pixel accuracy at various zoom levels
- **Invalid Inputs**: Malformed paths, invalid colors, out-of-bounds operations

**Expected Output**: 15+ reference images testing edge cases

### 12. Cross-Platform Consistency

**Purpose**: Ensure consistent rendering across platforms and pixel formats

**Test Cases**:

- **Platform Backends**: SDL2 vs mock backend comparisons, X11 consistency
- **Pixel Formats**: RGB24, RGB32, RGBA32, BGR variants, endianness testing
- **Operating Systems**: Linux consistency, behavior validation
- **Compiler Differences**: Go version compatibility, optimization level effects
- **Hardware Variations**: Different graphics hardware, GPU vs CPU rendering

**Expected Output**: 20+ reference images ensuring cross-platform consistency

## Implementation Phases

### Phase 1: Test Infrastructure (Weeks 1-2)

- Visual test framework in `/tests/visual/`
- Image comparison utilities (pixel-perfect and fuzzy matching)
- Reference image management system
- Test result reporting (HTML with side-by-side image comparisons)
- CI integration for automated visual validation

### Phase 2: Core Primitive Tests (Weeks 3-4)

- Basic shapes: lines, circles, rectangles, polygons
- Anti-aliasing quality validation
- Color accuracy and pixel format tests
- Transform correctness verification
- **Target**: 150+ test images covering fundamentals

### Phase 3: Advanced Feature Tests (Weeks 5-6)

- Path operations and stroke styles
- Gradient and pattern rendering
- Text rendering validation
- Image processing operations
- **Target**: 100+ test images covering advanced features

### Phase 4: Integration and Quality Assurance (Weeks 7-8)

- Complex scene rendering tests
- Edge case and stress testing
- Cross-platform validation
- Performance benchmarks with visual output
- Documentation and maintenance procedures
- **Target**: 75+ test images covering integration scenarios

## Test Infrastructure Design

### Directory Structure

```text
tests/visual/
├── framework/           # Test infrastructure
│   ├── compare.go      # Image comparison utilities
│   ├── runner.go       # Test execution framework
│   └── report.go       # HTML report generation
├── reference/          # Golden reference images
│   ├── primitives/     # Basic shapes
│   ├── paths/          # Path operations
│   ├── transforms/     # Transformations
│   └── ...            # Other categories
├── output/             # Generated test images
├── diffs/              # Difference images for failures
└── reports/            # HTML test reports
```

### Test Framework Features

- **Automated Execution**: Run all visual tests with single command
- **Parallel Testing**: Multi-threaded test execution for speed
- **Fuzzy Matching**: Configurable tolerance for anti-aliased edges
- **Diff Generation**: Highlight pixel differences in failed tests
- **Performance Metrics**: Track rendering speed alongside visual accuracy
- **CI Integration**: Automatic visual validation on commits
- **Reference Management**: Tools for updating reference images

## Expected Deliverables

### 1. Visual Test Suite

- **400+ Reference Images**: Comprehensive coverage of all rendering features
- **Automated Test Runner**: Execute full test suite in under 5 minutes
- **Category Organization**: Tests grouped by functionality for targeted testing

### 2. Comparison and Validation Tools

- **Pixel-Perfect Comparison**: Exact matching for crisp graphics
- **Fuzzy Matching**: Tolerance-based comparison for anti-aliased content
- **Statistical Analysis**: Pixel difference histograms, PSNR metrics
- **Visual Diff Reports**: Side-by-side comparison with highlighted differences

### 3. Documentation and Reporting

- **HTML Test Gallery**: Visual browser for all test results
- **Regression Reports**: Historical tracking of visual changes
- **Performance Metrics**: Rendering time and memory usage tracking
- **Maintenance Guide**: Procedures for updating and maintaining tests

### 4. Continuous Integration

- **Automated Validation**: Visual tests run on every commit
- **Reference Image Management**: Tools for approving visual changes
- **Performance Monitoring**: Track rendering performance over time
- **Cross-Platform Testing**: Validate consistency across environments

### 5. Quality Assurance

- **100% Feature Coverage**: Every rendering capability has visual tests
- **Regression Prevention**: Catch visual bugs before they reach production
- **Performance Validation**: Ensure rendering speed meets requirements
- **Documentation**: Comprehensive test documentation and examples

## Success Metrics

- **Coverage**: 100% of public API methods have corresponding visual tests
- **Accuracy**: 99.9% pixel-perfect accuracy for non-anti-aliased content
- **Performance**: Test suite completes in under 10 minutes
- **Reliability**: Zero false positives in visual comparisons
- **Maintainability**: Easy to add new tests and update references

This comprehensive visual testing framework will ensure the AGG Go library maintains pixel-perfect rendering accuracy and prevents visual regressions throughout development.
