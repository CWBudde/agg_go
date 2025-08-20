# **TEST_TASKS.md**

This document tracks all failing tests in the AGG Go port, grouped by category, with analysis and actionable tasks.

---

## **Summary**

* **10 failing test cases** across **6 packages**
* **Several algorithmic issues** (color conversion, alpha blending, distance transform, vertex smoothing)
* **Implementation gaps** (missing types, undefined methods, zero outputs)
* Some previously listed issues are **resolved** ✅

---

## **1. Algorithm & Logic Failures**  *(Priority: MEDIUM-HIGH)*

### [ ] **Color Conversion Produces Zeroed Data**

**Package:** `internal/color/conv`
**Test:** `TestConvert`
**Errors (excerpt):**

```
At byte index 0 (0,0): expected 255, got 0
At byte index 2 (2,0): expected 128, got 0
At byte index 7 (1,1): expected 240, got 0
```

**Root Cause:**
Color conversion functions produce zeroed outputs instead of expected transformed bytes.

**Tasks:**

* [ ] Inspect conversion implementation in `color_conv.go`.
* [ ] Compare against original AGG reference behavior.
* [ ] Add intermediate debug dumps to locate where values are zeroed.
* [ ] Verify byte order and channel mapping.

---

### [ ] **Premultiplied Alpha Blending Incorrect**

**Package:** `internal/pixfmt`
**Test:** `TestPixFmtRGB24Pre`
**Error:**

```
Premultiplied blending failed: red component 63 should be greater than 128
```

**Root Cause:**
Blending formula diverges from AGG’s expected premultiplied alpha math.

**Tasks:**

* [ ] Revisit AGG formulas in `agg_pixfmt_rgba.h`.
* [ ] Validate Go’s implementation for all premultiplied pixel formats.
* [ ] Add detailed test coverage for blending edge cases.

---

### [ ] **Distance Transform Produces Wrong Values**

**Package:** `internal/span`
**Test:** `TestDistanceTransformAlgorithm`
**Error:**

```
DT result[2]: expected 4.000000, got 1.000000
DT result[5]: expected 4.000000, got 1.000000
```

**Root Cause:**
Distance field propagation likely fails; incorrect neighborhood or step calculations.

**Tasks:**

* [ ] Compare implementation with C++ `agg_span_gradient_contour`.
* [ ] Check whether initialization or accumulation order differs.
* [ ] Add visual regression tests for distance transform results.

---

### [ ] **VCGen Smooth Poly1 Corner Calculation Incorrect**

**Package:** `internal/vcgen`
**Test:** `TestVCGenSmoothPoly1_CornerCalculation`
**Error:**

```
Control points should be offset from original corner
```

**Root Cause:**
Control point offsets for smoothed polygon corners diverge from AGG’s math.

**Tasks:**

* [ ] Cross-check `smooth_poly1.go` with `agg_vcgen_smooth_poly1.cpp`.
* [ ] Verify angle-based offset calculation.
* [ ] Add tests for triangles, rectangles, and concave polygons.

---

## **2. Implementation Deviations**  *(Priority: MEDIUM)*

### [ ] **Conv Marker Adaptor Produces Zero Vertices**

**Package:** `internal/conv`
**Test:** `TestConvMarkerAdaptor_CustomMarkers`
**Error:**

```
Expected 12 marker vertices, got 0
```

**Root Cause:**
Marker adaptor pipeline likely disconnected from upstream vertex sources.

**Tasks:**

* [ ] Check how markers are passed to the adaptor.
* [ ] Ensure correct initialization of marker generators.
* [ ] Compare against AGG’s `conv_marker_adaptor` behavior.

---

### [ ] **GPC Union Produces Empty Results**

**Package:** `internal/conv`
**Test:** `TestExampleUsage/Union`
**Errors:**

```
Union operation produced 0 vertices
Union operation should always produce some result
```

**Root Cause:**
Polygon clipping engine integration is incomplete or miswired.

**Tasks:**

* [ ] Verify GPC adapter initialization.
* [ ] Compare union test with AGG’s reference outputs.
* [ ] Check winding rules and polygon orientation.

---

## **3. Cleaned-Up / Resolved Issues** ✅

* SDL2 missing dependency
* X11 CGO compilation
* Rasterizer scanline clipping
* VPGen segmentator interpolation behavior
* ConvSmoothPoly1Curve approximation control
* Example linting issues
