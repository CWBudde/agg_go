# AGG Go Generics Audit

This document catalogs every generic type in the codebase and classifies it into one of three categories:

- **Category A: True Generics** - Keep as generic (Go generics work well)
- **Category B: False Generics** - Make concrete (runtime type assertions needed)
- **Category C: Combinatorial Explosion** - Consider code generation

## Summary

| Category | Count | Description                                 |
| -------- | ----- | ------------------------------------------- |
| A        | ~40   | True generics - keep as-is                  |
| B        | 5     | False generics - refactor to concrete types |
| C        | 0     | No code generation needed                   |

**Key Finding**: Most generics are legitimate. Only a few use `any()` casts and should be refactored.

---

## Category A: True Generics (Keep)

These types use Go generics correctly without runtime type assertions.

### A.1 Geometric Types

| Go Type               | C++ Origin           | Constraint                               | Notes                  |
| --------------------- | -------------------- | ---------------------------------------- | ---------------------- |
| `Point[T CoordType]`  | `point_d`, `point_i` | `~int \| ~int32 \| ~float32 \| ~float64` | Clean parameterization |
| `Rect[T CoordType]`   | `rect_i`, `rect_d`   | `~int \| ~int32 \| ~float32 \| ~float64` | Clean parameterization |
| `Vertex[T CoordType]` | `vertex_d`           | `~int \| ~int32 \| ~float32 \| ~float64` | Clean parameterization |

### A.2 Container Types

| Go Type                | C++ Origin           | Constraint | Notes                  |
| ---------------------- | -------------------- | ---------- | ---------------------- |
| `PodArray[T]`          | `pod_array<T>`       | `any`      | Standard container     |
| `PodVector[T]`         | `pod_vector<T>`      | `any`      | Standard container     |
| `PodBVector[T]`        | `pod_bvector<T>`     | `any`      | Block-allocated vector |
| `PodAutoArray[T]`      | `pod_auto_array<T>`  | `any`      | Auto-sizing array      |
| `PodAutoVector[T]`     | `pod_auto_vector<T>` | `any`      | Auto-sizing vector     |
| `PodStructArray[T]`    | N/A                  | `any`      | Go-specific helper     |
| `PodArrayAdaptor[T]`   | N/A                  | `any`      | Adaptor pattern        |
| `RangeAdaptor[T]`      | N/A                  | `any`      | Algorithm helper       |
| `SliceRangeAdaptor[T]` | N/A                  | `any`      | Algorithm helper       |

### A.3 Color Types (Space Parameterization)

| Go Type            | C++ Origin | Constraint    | Notes                |
| ------------------ | ---------- | ------------- | -------------------- |
| `RGBA8[CS Space]`  | `rgba8`    | `color.Space` | Linear/SRGB variants |
| `RGBA16[CS Space]` | `rgba16`   | `color.Space` | Linear/SRGB variants |
| `RGBA32[CS Space]` | `rgba32`   | `color.Space` | Linear/SRGB variants |
| `RGB8[CS Space]`   | `rgb8`     | `color.Space` | Linear/SRGB variants |
| `RGB16[CS Space]`  | `rgb16`    | `color.Space` | Linear/SRGB variants |
| `RGB32[CS Space]`  | `rgb32`    | `color.Space` | Linear/SRGB variants |
| `Gray8[CS Space]`  | `gray8`    | `color.Space` | Linear/SRGB variants |
| `Gray16[CS Space]` | `gray16`   | `color.Space` | Linear/SRGB variants |
| `Gray32[CS Space]` | `gray32`   | `color.Space` | Linear/SRGB variants |

### A.4 Blender Types

| Go Type                                  | C++ Origin           | Constraint   | Notes      |
| ---------------------------------------- | -------------------- | ------------ | ---------- |
| `BlenderRGBA8[S, O]`                     | `blender_rgba`       | Space, Order | Well-typed |
| `BlenderRGBA8Pre[S, O]`                  | `blender_rgba_pre`   | Space, Order | Well-typed |
| `BlenderRGBA8Plain[S, O]`                | `blender_rgba_plain` | Space, Order | Well-typed |
| `BlenderRGB8[S, O]`                      | `blender_rgb`        | Space, Order | Well-typed |
| `BlenderRGBPre[S, O]`                    | `blender_rgb_pre`    | Space, Order | Well-typed |
| `BlenderGray8[S]`                        | `blender_gray`       | Space        | Well-typed |
| `BlenderGray8Pre[S]`                     | `blender_gray_pre`   | Space        | Well-typed |
| (similar for 16-bit and 32-bit variants) |                      |              |            |

### A.5 Pixel Format Types

| Go Type                        | C++ Origin                | Constraint     | Notes                         |
| ------------------------------ | ------------------------- | -------------- | ----------------------------- |
| `PixFmtAlphaBlendRGBA[S, B]`   | `pixfmt_alpha_blend_rgba` | Space, Blender | Uses fast-path optimization\* |
| `PixFmtAlphaBlendRGB[S, B]`    | `pixfmt_alpha_blend_rgb`  | Space, Blender | Uses fast-path optimization\* |
| `PixFmtAlphaBlendGray[CS, B]`  | `pixfmt_alpha_blend_gray` | Space, Blender | Uses fast-path optimization\* |
| (similar for other bit depths) |                           |                |                               |

\*Fast-path optimization uses `any()` to check for `RawRGBAOrder` interface - this is a legitimate optimization pattern, not a false generic.

### A.6 Renderer Types

| Go Type                               | C++ Origin                    | Constraint         | Notes      |
| ------------------------------------- | ----------------------------- | ------------------ | ---------- |
| `RendererBase[PF, C]`                 | `renderer_base`               | PixelFormat, Color | Well-typed |
| `RendererMClip[PF, C]`                | `renderer_mclip`              | PixelFormat, Color | Well-typed |
| `RendererScanlineAA[BR, SA, SG, C]`   | `renderer_scanline_aa`        | Multiple           | Well-typed |
| `RendererScanlineBin[BR, SA, SG, C]`  | `renderer_scanline_bin`       | Multiple           | Well-typed |
| `RendererScanlineAASolid[BR, C]`      | `renderer_scanline_aa_solid`  | BaseRen, Color     | Well-typed |
| `RendererPrimitives[BR, C]`           | `renderer_primitives`         | BaseRen, Color     | Well-typed |
| `RendererOutlineAA[BR, C]`            | `renderer_outline_aa`         | BaseRen, Color     | Well-typed |
| `RendererMarkers[BR, C]`              | `renderer_markers`            | BaseRen, Color     | Well-typed |
| `RendererEnlargedT[Ren, C]`           | N/A                           | Renderer, Color    | Well-typed |
| `RendererRasterHTextSolid[BR, GG, C]` | `renderer_raster_htext_solid` | Multiple           | Well-typed |

### A.7 Span/Interpolator Types

| Go Type                             | C++ Origin                        | Constraint         | Notes                              |
| ----------------------------------- | --------------------------------- | ------------------ | ---------------------------------- |
| `SpanInterpolatorLinear[T]`         | `span_interpolator_linear`        | Transformer        | Well-typed                         |
| `SpanInterpolatorLinearSubdiv[T]`   | `span_interpolator_linear_subdiv` | Transformer        | Well-typed                         |
| `SpanInterpolatorTrans[T]`          | `span_interpolator_trans`         | Transformer        | Well-typed                         |
| `SpanInterpolatorAdaptor[I, D]`     | N/A                               | Interp, Distortion | Well-typed                         |
| `SpanSubdivAdaptor[I]`              | `span_subdiv_adaptor`             | Interpolator       | Uses `any()` for interface check\* |
| `SpanGradient[CT, IT, GT, CT2]`     | `span_gradient`                   | Multiple           | Well-typed                         |
| `SpanGradientAlpha[CT, IT, GT, AT]` | `span_gradient_alpha`             | Multiple           | Well-typed                         |
| `SpanPatternRGBA[S]`                | `span_pattern_rgba`               | Source             | Well-typed                         |
| `SpanPatternRGB[S]`                 | `span_pattern_rgb`                | Source             | Well-typed                         |
| `SpanPatternGray[S]`                | `span_pattern_gray`               | Source             | Well-typed                         |
| `SpanImageFilter*[S, I]`            | Various                           | Source, Interp     | Well-typed                         |
| `SpanConverter[C, SG, SC]`          | `span_converter`                  | Multiple           | Uses `any()` for color type\*      |
| `SpanAllocator[C]`                  | `span_allocator`                  | Color              | Well-typed                         |
| `SolidSpanGenerator[C]`             | N/A                               | Color              | Well-typed                         |
| `GradientLUT[T, CI]`                | `gradient_lut`                    | Color, Interp      | Well-typed                         |
| `ColorInterpolatorRGBA8[CS]`        | N/A                               | Space              | Well-typed                         |

\*Uses `any()` to check for optional interfaces - legitimate optimization pattern.

### A.8 Rasterizer Types

| Go Type                                   | C++ Origin               | Constraint        | Notes      |
| ----------------------------------------- | ------------------------ | ----------------- | ---------- |
| `RasterizerScanlineAA[C, V, Clip]`        | `rasterizer_scanline_aa` | Coord, Conv, Clip | Well-typed |
| `RasterizerScanlineAANoGamma[C, V, Clip]` | N/A                      | Coord, Conv, Clip | Well-typed |
| `RasterizerCompoundAA[Clip]`              | `rasterizer_compound_aa` | Clip              | Well-typed |
| `RasterizerOutline[R, C]`                 | `rasterizer_outline`     | Renderer, Color   | Well-typed |
| `RasterizerOutlineAA[R, C]`               | `rasterizer_outline_aa`  | Renderer, Color   | Well-typed |
| `RasterizerSlClip[C, V]`                  | `rasterizer_sl_clip`     | Coord, Conv       | Well-typed |

### A.9 Path Storage Types

| Go Type                           | C++ Origin               | Constraint                   | Notes      |
| --------------------------------- | ------------------------ | ---------------------------- | ---------- |
| `PathBase[VC]`                    | `path_base`              | VertexContainer              | Well-typed |
| `PathStorageInteger[T]`           | `path_storage_integer`   | `~int16 \| ~int32 \| ~int64` | Well-typed |
| `VertexBlockStorage[T]`           | `vertex_block_storage`   | Coord constraint             | Well-typed |
| `VertexStlStorage[T]`             | `vertex_stl_storage`     | `~float32 \| ~float64`       | Well-typed |
| `VertexInteger[T]`                | `vertex_integer`         | `~int16 \| ~int32 \| ~int64` | Well-typed |
| `SerializedIntegerPathAdaptor[T]` | N/A                      | Integer constraint           | Well-typed |
| `PolyPlainAdaptor[T]`             | `poly_plain_adaptor`     | Coord constraint             | Well-typed |
| `PolyContainerAdaptor[C, V]`      | `poly_container_adaptor` | Container, Vertex            | Well-typed |

### A.10 Converter Types

| Go Type                 | C++ Origin           | Constraint              | Notes      |
| ----------------------- | -------------------- | ----------------------- | ---------- |
| `ConvTransform[VS, T]`  | `conv_transform`     | VertexSource, Transform | Well-typed |
| `ConvConcat[VS1, VS2]`  | `conv_concat`        | VertexSource x2         | Well-typed |
| `ConvAdaptorVPGen[VPG]` | `conv_adaptor_vpgen` | VPGen                   | Well-typed |
| `ConvCurveInteger[T]`   | N/A                  | Integer constraint      | Well-typed |
| `ConvGPC[VSA, VSB]`     | `conv_gpc`           | VertexSource x2         | Well-typed |

### A.11 Image/Buffer Types

| Go Type                         | C++ Origin               | Constraint     | Notes      |
| ------------------------------- | ------------------------ | -------------- | ---------- |
| `RenderingBuffer[T]`            | `rendering_buffer`       | `any`          | Well-typed |
| `RenderingBufferCache[T]`       | N/A                      | `any`          | Well-typed |
| `RowPtrCache[T]`                | `row_ptr_cache`          | `any`          | Well-typed |
| `ImageAccessorClip[PF]`         | `image_accessor_clip`    | PixelFormat    | Well-typed |
| `ImageAccessorNoClip[PF]`       | `image_accessor_no_clip` | PixelFormat    | Well-typed |
| `ImageAccessorClone[PF]`        | `image_accessor_clone`   | PixelFormat    | Well-typed |
| `ImageAccessorWrap[PF, WX, WY]` | `image_accessor_wrap`    | Multiple       | Well-typed |
| `ImageFilter[F]`                | `image_filter`           | FilterFunction | Well-typed |

### A.12 Gamma/LUT Types

| Go Type            | C++ Origin  | Constraint  | Notes      |
| ------------------ | ----------- | ----------- | ---------- |
| `GammaLUT[Lo, Hi]` | `gamma_lut` | Unsigned x2 | Well-typed |
| `SRGBLUT[T]`       | `srgb_lut`  | Numeric     | Well-typed |
| `SRGBConv[T]`      | `srgb_conv` | Numeric     | Well-typed |

### A.13 Effect Types

| Go Type                    | C++ Origin                 | Constraint           | Notes      |
| -------------------------- | -------------------------- | -------------------- | ---------- |
| `StackBlurCalcRGBA[T]`     | `stack_blur_calc_rgba`     | `~uint32 \| ~uint64` | Well-typed |
| `RecursiveBlurCalcRGBA[T]` | `recursive_blur_calc_rgba` | `~float64`           | Well-typed |

### A.14 Font Cache Types

| Go Type                    | C++ Origin           | Constraint  | Notes      |
| -------------------------- | -------------------- | ----------- | ---------- |
| `FontCacheManager[T]`      | `font_cache_manager` | FontEngine  | Well-typed |
| `FmanFontCacheManager2[T]` | N/A                  | FontEngine2 | Well-typed |

### A.15 Control Types (UI)

| Go Type           | C++ Origin      | Constraint | Notes      |
| ----------------- | --------------- | ---------- | ---------- |
| `RboxCtrl[C]`     | `rbox_ctrl`     | Color      | Well-typed |
| `SplineCtrl[C]`   | `spline_ctrl`   | Color      | Well-typed |
| `GammaCtrl[C]`    | `gamma_ctrl`    | Color      | Well-typed |
| `CheckboxCtrl[C]` | `checkbox_ctrl` | Color      | Well-typed |
| `BezierCtrl[C]`   | `bezier_ctrl`   | Color      | Well-typed |
| `Curve3Ctrl[C]`   | `curve3_ctrl`   | Color      | Well-typed |
| `PolygonCtrl[C]`  | `polygon_ctrl`  | Color      | Well-typed |

### A.16 Scanline Storage Types

| Go Type                           | C++ Origin              | Constraint | Notes      |
| --------------------------------- | ----------------------- | ---------- | ---------- |
| `ScanlineStorageAA[T]`            | `scanline_storage_aa`   | `any`      | Well-typed |
| `ScanlineCellStorage[T]`          | `scanline_cell_storage` | `any`      | Well-typed |
| `EmbeddedScanline[T]`             | N/A                     | `any`      | Well-typed |
| `SerializedScanlinesAdaptorAA[T]` | N/A                     | `any`      | Well-typed |

### A.17 Miscellaneous

| Go Type                  | C++ Origin        | Constraint         | Notes                            |
| ------------------------ | ----------------- | ------------------ | -------------------------------- |
| `PodAllocator[T]`        | `pod_allocator`   | `any`              | Memory helper                    |
| `ObjAllocator[T]`        | `obj_allocator`   | `any`              | Memory helper                    |
| `Saturation[T]`          | N/A               | Integer constraint | Uses `any()` for type dispatch\* |
| `MulOne[T]`              | N/A               | Integer constraint | Well-typed                       |
| `Comparator[T]`          | N/A               | `any`              | Algorithm helper                 |
| `ReverseComparator[T]`   | N/A               | `any`              | Algorithm helper                 |
| `CompositeBlender[S, O]` | `comp_op_adaptor` | Space, Order       | Well-typed                       |

---

## Category B: False Generics (Refactor)

These types use `any()` casts because generic constraints cannot express the required behavior.

### B.1 VertexSequence[T]

**File**: `internal/array/vertex_sequence.go`

**Problem**: Uses `any()` casts to determine concrete type for distance calculations:

```go
if vd, ok := any(curr).(VertexDist); ok { ... }
if lv, ok := any(curr).(LineAAVertex); ok { ... }
```

**C++ Instantiations**: Only 2 types

- `vertex_sequence<vertex_dist>`
- `vertex_sequence<line_aa_vertex>`

**Status**: Already has concrete replacements:

- `VertexDistSequence` - eliminates `any()` casts
- `LineAAVertexSequence` - eliminates `any()` casts

**Decision**: **KEEP CONCRETE** - The generic `VertexSequence[T]` is deprecated. Use concrete types.

---

### B.2 RasterizerCellsAA[Cell]

**File**: `internal/rasterizer/cells_aa.go`

**Problem**: Uses type switch on `any(dummy).(type)` to create cells:

```go
switch any(dummy).(type) {
case *CellStyleAA:
    r.styleCell = any(&CellStyleAA{}).(Cell)
case *CellAA:
    r.styleCell = any(&CellAA{}).(Cell)
}
```

**C++ Instantiations**: Only 2 types

- `rasterizer_cells_aa<cell_aa>`
- `rasterizer_cells_aa<cell_style_aa>`

**Recommendation**: Create concrete types:

- `RasterizerCellsAA` (for `CellAA`)
- `RasterizerCellsStyleAA` (for `CellStyleAA`)

Or use a factory interface:

```go
type CellFactory interface {
    NewCell() Cell
}
```

**Decision**: **REFACTOR** - Create two concrete types to eliminate type switches.

---

### B.3 GammaLUT (internal/pixfmt/gamma/lut.go)

**File**: `internal/pixfmt/gamma/lut.go`

**Problem**: Uses extensive `any()` type switches:

```go
switch any(&lut.dirGamma[i]).(type) {
case *basics.Int8u:
    *any(&lut.dirGamma[i]).(*basics.Int8u) = basics.Int8u(scaled)
case *basics.Int16u:
    // etc.
}
```

**C++ Instantiations**: ~5 combinations (8u/8u, 8u/16u, 16u/8u, 16u/16u, 32u/32u)

**Recommendation**: Since the number of instantiations is manageable, create concrete types:

- `GammaLUT8` (8-bit in, 8-bit out)
- `GammaLUT16` (16-bit in, 16-bit out)
- `GammaLUT8to16`, `GammaLUT16to8` if needed

Or use code generation if more combinations are needed.

**Decision**: **REFACTOR** - Create concrete implementations for common LUT sizes.

---

### B.4 Saturation[T] (internal/basics/constants.go)

**File**: `internal/basics/constants.go`

**Problem**: Uses `any()` type switch for bit-depth detection:

```go
switch any(zero).(type) {
case int:   return 0x7FFFFFFF
case int32: return 0x7FFFFFFF
// etc.
}
```

**Recommendation**: This is a utility for constant values. Replace with explicit functions:

- `SaturationInt() int`
- `SaturationInt32() int32`
- etc.

Or use a constraint-based approach with compile-time type specialization.

**Decision**: **REFACTOR** - Replace with explicit typed functions.

---

### B.5 Gray8[CS] (partial)

**File**: `internal/color/gray8.go`

**Problem**: One method uses type switch:

```go
switch any(c).(type) {
case Gray8[color.Linear]:
    // linear handling
case Gray8[color.SRGB]:
    // sRGB handling
}
```

**Recommendation**: This is a minor case. The color space should drive behavior through the `CS` type parameter, not runtime checks.

**Decision**: **REFACTOR** - Move space-specific logic to type methods or blender.

---

## Category C: Combinatorial Explosion (Code Generation)

**Finding**: No types require code generation.

The pixel format + blender combinations are handled through the existing generic system with interface-based fast paths. The C++ template instantiation explosion is avoided through:

1. **Blender interfaces** - `RGBABlender[S]`, `RGBBlender[S]`, etc.
2. **Fast-path type assertions** - `if ro, ok := any(pf.blender).(RawRGBAOrder); ok`
3. **Type aliases** - Pre-defined combinations like `PixFmtRGBA32`, `PixFmtBGRA32`

This pattern is correct and doesn't need code generation.

---

## Interface{} Usage Analysis

The following files use `interface{}` in type definitions:

| File                                           | Usage                                                    | Assessment                                 |
| ---------------------------------------------- | -------------------------------------------------------- | ------------------------------------------ |
| `internal/agg2d/agg2d.go`                      | `fontEngine interface{}`, `fontCacheManager interface{}` | **Needs typing** - CGO boundary workaround |
| `internal/agg2d/text.go`                       | `adaptor interface{}`                                    | **Needs typing** - CGO boundary            |
| `internal/font/freetype2/types.go`             | `rasterizer interface{}`                                 | **Needs typing** - Complex generics        |
| `internal/font/freetype2/engine.go`            | `pathStorage interface{}`                                | **Needs typing** - CGO boundary            |
| `internal/font/freetype2/cache_integration.go` | `pathAdaptor interface{}`                                | **Needs typing** - CGO boundary            |

**Root Cause**: The font subsystem uses CGO and has complex generic dependencies that are difficult to express in Go. These `interface{}` usages are isolated to the font rendering path.

**Recommendation**: Document as technical debt. Consider:

1. Creating a dedicated `FontTypes` interface package
2. Using internal type assertions in a single location
3. Eventually introducing proper interfaces as the font API stabilizes

---

## Any() Usage Patterns

### Legitimate Optimization Patterns

These uses of `any()` are correct and should be kept:

1. **Fast-path detection** (pixfmt files):

   ```go
   if ro, ok := any(pf.blender).(blender.RawRGBAOrder); ok {
       // Use raw indices for performance
   }
   ```

   This checks for an optional interface to enable optimized code paths.

2. **Optional interface checks** (span files):

   ```go
   if transformerGetter, ok := any(s.interpolator).(transform.TransformerGetter); ok {
       // Use getter interface if available
   }
   ```

   This implements optional behavior extensions.

3. **Color type adaptation** (span/converter.go):
   ```go
   switch color := any(c).(type) {
   case color.RGBA8[color.SRGB]:
       // Handle specific color type
   }
   ```
   This handles color type conversion at boundaries.

### Problematic Patterns (to refactor)

1. **Type dispatch in generic code** (vertex_sequence.go, cells_aa.go)
2. **Numeric type switches** (gamma/lut.go, constants.go)

---

## Relationship to TODO_GENERICS.md

The existing `docs/TODO_GENERICS.md` documents the **pixel format generics refactoring** specifically. It tracks:

- Blender interface cleanup
- Removal of order parameter from pixfmt types
- Addition of utility methods (Premultiply, Demultiply, ApplyGamma)

**Assessment**: TODO_GENERICS.md is **complementary** to this audit:

- TODO_GENERICS.md = Implementation tracking for pixfmt refactoring
- GENERICS_AUDIT.md = Comprehensive catalog of all generics

**Recommendation**: Keep both documents:

- Mark TODO_GENERICS.md as "Pixfmt Generics Status" (specific tracking)
- This document serves as the comprehensive audit

---

## Action Items

### High Priority

1. [ ] **Refactor RasterizerCellsAA** - Create `RasterizerCellsAA` and `RasterizerCellsStyleAA` concrete types
2. [ ] **Remove deprecated VertexSequence[T]** - Ensure all code uses `VertexDistSequence` or `LineAAVertexSequence`

### Medium Priority

3. [ ] **Refactor GammaLUT** - Create concrete implementations for common LUT configurations
4. [ ] **Refactor Saturation** - Replace with explicit typed functions
5. [ ] **Fix Gray8 type switch** - Move space-specific logic to appropriate location

### Low Priority (Technical Debt)

6. [ ] **Document font subsystem interface{}** - Accept as CGO boundary limitation
7. [ ] **Profile fast-path patterns** - Verify `any()` optimizations provide measurable benefit

---

## Verification Commands

To find all `any()` casts:

```bash
grep -rn 'any(' internal/ --include='*.go' | grep -v '_test.go'
```

To find all generic type definitions:

```bash
grep -rn 'type \w\+\[' internal/ --include='*.go'
```

To find all `interface{}` usage:

```bash
grep -rn 'interface{}' internal/ --include='*.go' | grep -v '_test.go'
```
