# TEST_TASKS.md

This document catalogs all failing tests in the AGG Go port with analysis and recommendations for fixes.

## Summary

**5 packages have failures:** ✅ **2 PACKAGES FIXED**

- 2 Build failures (missing dependencies)
- 8 Algorithm/logic failures and implementation deviations ✅ **2 FIXED**

**Total failures:** ~~13~~ ~~11~~ **11** specific test cases + 2 build failures + 5 implementation deviations ✅ **5 CASES FIXED**

## Build Failures (Priority: HIGH - Blocks development)

### 1. `internal/platform/sdl2` - Missing Dependency

**Error:** `no required module provides package github.com/veandco/go-sdl2/sdl`
**Status:** Build failed
**Root Cause:** Missing SDL2 Go bindings dependency
**Fix:** Run `go get github.com/veandco/go-sdl2/sdl` or make SDL2 backend optional
**Priority:** Medium (optional platform support)

### 2. `internal/platform/x11` - CGO/X11 Issues

**Error:** `could not determine what C.XDestroyImage refers to`
**Status:** Build failed  
**Root Cause:** CGO compilation issues with X11 headers
**Fix:** Install X11 development headers or fix CGO directives
**Priority:** Medium (optional platform support)

## Algorithm Failures (Priority: MEDIUM-HIGH)

### 3. `internal/pixfmt` - Premultiplied Alpha Blending

#### TestPixFmtRGB24Pre Failure:

**Error:** `Premultiplied blending failed: red component 63 should be greater than 128`
**Root Cause:** Incorrect premultiplied alpha blending mathematics
**File:** `internal/pixfmt/pixfmt_rgb.go` and various pixel format implementations
**Analysis:** Blending calculation produces unexpected low values for red component. The red component produces 63 instead of expected >128.

**Expected Behavior:** Should match AGG's premultiplied alpha blending formulas exactly.

**Impact:**

- Incorrect color blending in premultiplied alpha modes
- Visual artifacts in rendered output
- Compositing operations produce wrong results

**Fix:**

1. Review AGG's alpha premultiplication formulas in agg_pixfmt_rgba.h
2. Verify RGBA8Prelerp implementation and usage
3. Fix blending mathematics for all premultiplied formats
4. Ensure color component calculations match C++ behavior exactly
5. Add comprehensive test coverage for all blending modes

**Priority:** HIGH (affects visual correctness)

### 4. `internal/rasterizer` - Scanline Clipping

#### TestRasterizerSlClipWithClipping Failures:

- **fully_outside_left:** Expected 0 lines, got 1 line
- **crosses_boundary:** Expected 1 scanline, got 2

**Root Cause:** Line clipping boundary detection produces incorrect results for edge cases
**File:** `internal/rasterizer/clip.go` and related clipping code
**Analysis:** Rasterizer generates extra scanlines when crossing clip boundaries

**Expected Behavior:** Should match AGG's scanline clipping algorithm precisely.

**Impact:**

- Incorrect rendering of clipped geometry
- Lines may appear when they should be culled
- Clipping artifacts at viewport boundaries

**Fix:**

1. Review original AGG scanline clipping implementation
2. Compare with agg_rasterizer_sl_clip.h behavior
3. Fix boundary detection logic for all edge cases
4. Ensure proper handling of lines that touch but don't cross boundaries
5. Verify clipping coordinate system matches AGG conventions

**Priority:** MEDIUM-HIGH (affects geometry rendering)

### 5. `internal/vcgen` - Vertex Generator State Management

#### TestVCGenSmoothPoly1 Failures:

- **Basic:** Expected multiple Curve4 commands for smoothing, got 6 (unclear expectation)
- **InsufficientVertices:** Panic due to index bounds violation
- **CornerCalculation:** Control point calculation for smoothed polygon corners doesn't produce expected offset

**Panic:** `index 2 out of bounds [0, 2)`
**Root Cause:** Missing bounds checking in vertex sequence access and incorrect corner calculation
**File:** `internal/vcgen/smooth_poly1.go` (calculate method)

**Expected Behavior:**

- Control points should be properly offset from original corner vertices according to smoothing algorithm
- Should handle edge cases with insufficient vertices gracefully

**Impact:**

- Causes crashes (panic on bounds violation)
- Incorrect smooth polygon rendering
- Corner smoothing may not work as expected
- Generated bezier curves may have wrong curvature

**Fix:**

1. Add proper bounds checking and handle degenerate cases gracefully
2. Review AGG's agg_vcgen_smooth_poly1.cpp implementation
3. Verify corner angle and distance calculations
4. Check control point generation mathematics
5. Ensure smooth value scaling matches C++ behavior
6. Add comprehensive test cases for various polygon shapes

**Priority:** HIGH (causes crashes)

## Implementation Deviations from C++ AGG

These issues represent deviations from the original C++ AGG 2.6 implementation that need to be addressed for full compatibility.

### 6. `internal/vpgen` - VPGen Segmentator Interpolation Behavior

**Location:** `internal/vpgen/vpgen_segmentator.go`

**Problem:** For lines where length ≤ approximation scale (e.g., 1-unit line with scale 1.0), the segmentator produces only the endpoint with MoveTo command instead of producing both start and end points as separate vertices.

**Expected Behavior:** Should likely produce the start point as MoveTo and end point as LineTo, even for very short segments.

**Impact:**

- TestVPGenSegmentator_ShortLine had to be adjusted to expect different output
- May affect rendering quality for short line segments
- Differs from C++ vpgen_segmentator behavior

**Required Fix:**

1. Compare behavior with original C++ agg_vpgen_segmentator.cpp
2. Verify correct interpolation algorithm for dl/ddl calculations
3. Ensure proper vertex generation for all line lengths
4. Update tests to match correct C++ behavior

**Priority:** LOW (minor visual impact)

### 7. `internal/conv` - ConvSmoothPoly1Curve Integration

**Location:** `internal/conv/conv_smooth_poly1.go` (ConvSmoothPoly1Curve)

**Problem:** The curve approximation control is not fully implemented. SetCurveApproximation() and CurveApproximation() methods have placeholder implementations.

**Expected Behavior:** Should properly integrate with ConvCurve for smooth curve-to-line approximation control.

**Impact:**

- No runtime switching between raw curves and approximated segments
- Missing fine control over curve approximation quality
- Not compatible with AGG's curve approximation framework

**Required Fix:**

1. Implement proper ConvCurve integration
2. Add approximation scale and tolerance controls
3. Enable runtime switching between curve modes
4. Match C++ conv_smooth_poly1 + conv_curve composition behavior

**Priority:** LOW (feature completeness)

## Example Build Failures

### 8. `examples/basic/colors` - Multiple main Functions

**Error:** `main redeclared in this block`
**Root Cause:** Multiple files in same package declare main()
**Fix:** Restructure example to have single main or separate packages
**Priority:** LOW (examples only)

### 9. `examples/platform/basic_demo` - Code Quality

**Error:** `fmt.Println arg list ends with redundant newline`
**Root Cause:** Linting issues in example code
**Fix:** Remove redundant newlines from print statements
**Priority:** LOW (examples only)

## Dependency Analysis

### Missing AGG Components (from TASKS.md)

Many test failures stem from incomplete implementations. Key missing dependencies:

- Advanced curve approximation algorithms
- Complete scanline rendering pipeline
- Full rasterizer cell storage and sorting
- Advanced path stroking algorithms

### Fix Order Recommendations

1. **IMMEDIATE (Blocks development):**

   - Fix `internal/vcgen` panic (safety issue)

2. **HIGH PRIORITY (Core functionality):**

   - Fix premultiplied alpha blending (affects visual correctness)
   - Fix rasterizer scanline clipping

3. **MEDIUM PRIORITY (Quality/accuracy):**

   - Fix VCGen smooth poly corner calculation (affects shape quality)
   - Clean up example code

4. **LOW PRIORITY (Optional features):**
   - Fix platform backend dependencies (SDL2, X11)
   - Fix VPGen segmentator behavior (minor visual impact)
   - Fix ConvSmoothPoly1Curve integration (feature completeness)

## Testing Strategy

### After Algorithm Fixes:

1. All core algorithm tests pass
2. Integration tests work with fixed components
3. No regressions in currently passing tests
4. Performance benchmarks show expected behavior

### Comprehensive Testing Strategy for Implementation Issues:

1. **Create C++ Reference Tests**: Build minimal C++ AGG programs that demonstrate correct behavior for each issue
2. **Comparative Analysis**: Compare Go output with C++ output for identical inputs
3. **Mathematical Verification**: Verify algorithms match published AGG mathematics
4. **Edge Case Coverage**: Ensure all boundary conditions are tested
5. **Performance Impact**: Measure any performance impact of correctness fixes

## Notes

- Many failures indicate systematic issues (coordinate orientation, bounds checking)
- Some tests may need adjustment if AGG behavior differs from assumptions
- Consider comparing with original AGG C++ test suite for expected behavior
- TODO comments already added to some failing areas per CLAUDE.md
