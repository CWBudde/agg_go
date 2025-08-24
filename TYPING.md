# Type Safety Migration: Eliminating Duck-Typing and interface{} Usage

This document tracks the migration from `interface{}` and untyped `any` to properly typed generics throughout the AGG Go port.

## Overview

The goal is to eliminate all inappropriate usage of `interface{}` and replace with:

- Properly constrained generics using type parameters
- Well-defined interfaces with concrete method signatures
- Type-safe color and pixel format handling

## Migration Checklist

### Phase 1: Core Pixel Format Interfaces

#### [x] Update `[ ] internal/pixfmt/base.go`

- [x] Convert `PixelFormat` interface to use generic color type parameter `[C any]`
  - [x] Change `CopyPixel(x, y int, c interface{})` to `CopyPixel(x, y int, c C)`
  - [x] Change `BlendPixel(x, y int, c interface{}, cover basics.Int8u)` to `BlendPixel(x, y int, c C, cover basics.Int8u)`
  - [x] Change `CopyHline(x1, y, x2 int, c interface{})` to `CopyHline(x1, y, x2 int, c C)`
  - [x] Change `BlendHline(x1, y, x2 int, c interface{}, cover basics.Int8u)` to `BlendHline(x1, y, x2 int, c C, cover basics.Int8u)`
  - [x] Change `CopyVline(x, y1, y2 int, c interface{})` to `CopyVline(x, y1, y2 int, c C)`
  - [x] Change `BlendVline(x, y1, y2 int, c interface{}, cover basics.Int8u)` to `BlendVline(x, y1, y2 int, c C, cover basics.Int8u)`
  - [x] Change `CopyBar(x1, y1, x2, y2 int, c interface{})` to `CopyBar(x1, y1, x2, y2 int, c C)`
  - [x] Change `BlendBar(x1, y1, x2, y2 int, c interface{}, cover basics.Int8u)` to `BlendBar(x1, y1, x2, y2 int, c C, cover basics.Int8u)`
  - [x] Change `BlendSolidHspan(x, y, length int, c interface{}, covers []basics.Int8u)` to use `C`
  - [x] Change `BlendSolidVspan(x, y, length int, c interface{}, covers []basics.Int8u)` to use `C`
  - [x] Change `Clear(c interface{})` to `Clear(c C)`
  - [x] Change `Fill(c interface{})` to `Fill(c C)`
- [x] Update all pixel format implementations to match new interface:
  - [x] `[ ] internal/pixfmt/pixfmt_rgb.go` (already generic-compatible)
  - [x] `[ ] internal/pixfmt/pixfmt_rgba.go` (fully updated with all interface methods)
  - [x] `[ ] internal/pixfmt/pixfmt_gray.go` (fully updated with all interface methods)
  - [x] `[ ] internal/pixfmt/pixfmt_gray16.go` (fully updated with all interface methods)
  - [x] `[ ] internal/pixfmt/pixfmt_gray32.go` (partially updated - needs signature fixes for float32 vs Int8u)
  - [x] `[ ] internal/pixfmt/pixfmt_rgba64.go` (already compatible with generic interface)
  - [x] `[ ] internal/pixfmt/pixfmt_rgb_packed.go` (already compatible with generic interface)

#### [x] Update Blender Interfaces

- [x] Update `[ ] internal/pixfmt/blender_rgb.go` to remove `interface{}` usage
  - [x] Replace `BlenderRGBGamma[CS any, O any, G any]` with `BlenderRGBGamma[CS any, O any, G GammaCorrector]`
  - [x] Replace `BlenderRGB48Gamma[CS any, O any, G any]` with `BlenderRGB48Gamma[CS any, O any, G Gamma16Corrector]`
  - [x] Replace `BlenderRGB96Gamma[CS any, O any, G any]` with `BlenderRGB96Gamma[CS any, O any, G Gamma32Corrector]`
  - [x] Remove runtime type assertions `interface{}(bl.gamma).(GammaCorrector)`
  - [x] Use direct method calls `bl.gamma.Dir()` and `bl.gamma.Inv()`
- [x] Update `[ ] internal/pixfmt/blender_rgba.go` to remove `interface{}` usage (already clean)
- [x] Update `[ ] internal/pixfmt/blender_gray.go` to remove `interface{}` usage (already clean)
- [x] Update `[ ] internal/pixfmt/blender_gray16.go` to remove `interface{}` usage (already clean)
- [x] Update `[ ] internal/pixfmt/blender_gray32.go` to remove `interface{}` usage (already clean)
- [x] Update `[ ] internal/pixfmt/blender_rgb_packed.go` to remove `interface{}` usage
  - [x] Replace `BlenderRGB555Gamma[G any]` with `BlenderRGB555Gamma[G GammaCorrector]`
  - [x] Replace `BlenderRGB565Gamma[G any]` with `BlenderRGB565Gamma[G GammaCorrector]`
  - [x] Replace `BlenderBGR555Gamma[G any]` with `BlenderBGR555Gamma[G GammaCorrector]`
  - [x] Replace `BlenderBGR565Gamma[G any]` with `BlenderBGR565Gamma[G GammaCorrector]`
  - [x] Remove runtime type assertions in all gamma blenders
- [x] Update `[ ] internal/pixfmt/blender_rgba_composite.go` to remove `interface{}` usage (already clean)
- [x] Update `[ ] internal/pixfmt/blender_base_test.go` to fix test struct using function-based approach

**Result:** All blender interfaces now use properly constrained generics instead of `interface{}` with runtime type assertions. Gamma correction functionality is fully compile-time type-safe while maintaining all existing behavior. All tests pass.

### Phase 2: Renderer Stack

#### [x] Fix `[ ] internal/agg2d/agg2d.go`

- [x] Replace `pixfmt interface{}` with properly typed pixel format
- [x] Replace `renBase interface{}` with typed base renderer
- [ ] Replace `renSolid interface{}` with typed solid renderer (deferred to later phase)
- [ ] Replace `fontEngine interface{}` with proper FontEngine interface (deferred to later phase)
- [ ] Replace `fontCacheManager interface{}` with proper cache manager type (deferred to later phase)
- [x] Fix `baseRendererAdapter` to use generics instead of type assertions:
  - [x] Remove runtime type checks like `if col, ok := c.(color.RGBA8[color.Linear]); ok`
  - [x] Make adapter generic over color type
  - [x] Update `BlendSolidHspan` method
  - [x] Update `BlendHline` method
  - [x] Update `BlendColorHspan` method

#### [x] Update Base Renderer (`[ ] internal/renderer/base.go`)

- [x] Already uses generics properly with `RendererBase[PF PixelFormat[C], C any]`
- [x] Ensure all implementations follow this pattern

#### [x] Update Multi-clip Renderer (`[ ] internal/renderer/mclip.go`)

- [x] Already uses generics properly with `RendererMClip[PF PixelFormat[C], C any]`
- [x] Verify all methods maintain type safety

#### [x] Update Enlarged Renderer (`[ ] internal/renderer/enlarged.go`)

- [x] Check for any `interface{}` usage and convert to generics (already clean - using proper generics)

#### [x] Fix Raster Text Renderers (`[ ] internal/renderer/raster_text.go`)

- [x] Update `BaseRendererInterface` to use generics: `BaseRendererInterface[C any]`
- [x] Update `RendererRasterHTextSolid` to use generics: `[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any]`
- [x] Update `RendererRasterVTextSolid` to use generics: `[BR BaseRendererInterface[C], GG glyph.GlyphGenerator, C any]`
- [x] Remove `interface{}` from color parameter/return types
- [x] Update all method signatures to use typed colors

#### [x] Fix Marker Renderer (`[ ] internal/renderer/markers/markers.go`)

- [x] Change `interface{}` type parameter to `any`: `RendererMarkers[BR primitives_pkg.BaseRenderer[C], C any]`
- [x] Update constructor and method signatures
- [x] Fix test mock implementations to use typed colors

#### [x] Fix Primitives Renderer (`[ ] internal/renderer/primitives/primitives.go`)

- [x] Update `BaseRenderer` interface to use generics: `BaseRenderer[C any]`
- [x] Update `RendererPrimitives` to use proper generic constraints: `[BR BaseRenderer[C], C any]`
- [x] Remove all `interface{}` usage from color parameters
- [x] Update all method signatures and color handling
- [x] Fix test mock implementations

#### [x] Fix Outline Anti-Aliased Renderer (`[ ] internal/renderer/outline/outline_aa.go`)

- [x] Update `BaseRendererInterface` to use generics: `BaseRendererInterface[C any]`
- [x] Update `RendererOutlineAA` to use proper generics: `[BaseRenderer BaseRendererInterface[C], C any]`
- [x] Remove `interface{}` from color parameters and return types
- [x] Update all method signatures to use typed colors
- [x] Fix test mock implementations

### Phase 3: Scanline and Span Processing

#### [x] Fix Scanline Helpers (`[ ] internal/renderer/scanline/helpers.go`)

- [x] Convert `PathColorStorage` interface:
  - [x] Change `GetColor(index int) interface{}` to generic `GetColor(index int) C`
  - [x] Create `PathColorStorage[C any]` interface
- [x] Update `RenderAllPaths` function to use typed color storage
- [x] Remove type assertions in color handling

#### [x] Update Scanline Render Functions (`[ ] internal/renderer/scanline/render_functions.go`)

- [x] Ensure all render functions use typed generics
- [x] Remove any remaining `interface{}` parameters

#### [x] Fix Span Converter (`[ ] internal/span/converter.go`)

- [x] Replace `interface{}` nil checks (lines 60, 63, 73, 78) with proper zero value detection:
  - [x] Use reflection or type-specific zero value checks
  - [x] Consider adding an `IsZero()` method to the interface
- [x] Ensure all span generators use proper type parameters

#### [x] Update Span Allocator (`[ ] internal/span/allocator.go`)

- [x] Verify generic type usage throughout
- [x] Convert SpanAllocator from `[]interface{}` to generic `SpanAllocator[C SpanColorType]`
- [x] Update all method signatures to use typed colors `[]C` instead of `[]interface{}`
- [x] Replace nil checks with zero value initialization
- [x] Ensure compliance with `SpanAllocatorInterface[C any]`

#### [x] Update Span Generators

- [x] `[ ] internal/span/span_gradient.go` - check for `interface{}` usage (clean - uses proper generics)
- [x] `[ ] internal/span/span_gradient_alpha.go` - check for `interface{}` usage (clean - uses proper generics)
- [x] `[ ] internal/span/span_gradient_image.go` - check for `interface{}` usage (fixed - eliminated runtime type assertions)
- [x] `[ ] internal/span/span_gouraud.go` - check for `interface{}` usage (clean - uses proper generics)
- [x] `[ ] internal/span/span_image_filter_rgb.go` - check for `interface{}` usage (fixed - uses RGBSourceInterface directly)
- [x] `[ ] internal/span/span_image_filter_rgba.go` - check for `interface{}` usage (partially fixed - main interface{} eliminated, some types need RGBASourceInterface constraint)

**Result:** All major span generator `interface{}` usage has been eliminated. The core gradient generators already used proper generics. Image filter generators now use typed source interfaces instead of runtime type assertions. Some RGBA filter types still need conversion from `SourceInterface` to `RGBASourceInterface` for complete type safety.

### Phase 4: Font and Text Rendering

#### [x] Define Proper Font Interfaces (`[ ] internal/fonts/`)

- [x] Create proper interface types for:
  - [x] `PathAdaptorType` interface with required methods
  - [x] `Gray8AdaptorType` interface with required methods
  - [x] `Gray8ScanlineType` interface with required methods
  - [x] `MonoAdaptorType` interface with required methods
  - [x] `MonoScanlineType` interface with required methods

#### [x] Update Font Engine Interface (`[ ] internal/fonts/cache_manager2.go`)

- [x] Replace in `FontEngine2` interface:
  - [x] Change `PathAdaptor() interface{}` to `PathAdaptor() PathAdaptorType`
  - [x] Change `Gray8Adaptor() interface{}` to `Gray8Adaptor() Gray8AdaptorType`
  - [x] Change `Gray8Scanline() interface{}` to `Gray8Scanline() Gray8ScanlineType`
  - [x] Change `MonoAdaptor() interface{}` to `MonoAdaptor() MonoAdaptorType`
  - [x] Change `MonoScanline() interface{}` to `MonoScanline() MonoScanlineType`

#### [x] Update Cached Glyph Structure

- [x] Change `FmanCachedGlyph.CachedFont interface{}` to proper type (`*FmanCachedFont`)
- [x] Update `CacheGlyph` method signature to use typed parameter

#### [x] Update Font Engine Implementations

- [x] `[ ] internal/fonts/cache_manager2.go` - implement typed interfaces in `FmanFontCacheManager2`
- [x] Update mock implementations in tests to match
- [x] Remove all runtime type assertions from `InitEmbeddedAdaptors`

**Result:** All font `interface{}` usage has been eliminated. The font system now uses properly constrained generics with compile-time type safety. All adaptors and scanlines have well-defined interfaces. Runtime type assertions have been removed from the font cache manager. All tests pass.

### Phase 5: Rasterizer Components

#### [x] Update Rasterizer Interfaces

- [x] Check `[ ] internal/rasterizer/clip.go` for `interface{}` usage (clean - no interface{} usage found)
- [x] Check `[ ] internal/rasterizer/cells_aa.go` for `interface{}` usage (clean - no interface{} usage found)
- [x] Check `[ ] internal/rasterizer/compound_aa.go` for `interface{}` usage (clean - no interface{} usage found)
- [x] Check `[ ] internal/rasterizer/outline.go` for `interface{}` usage (fixed - converted to generic color type)
- [x] Check `[ ] internal/rasterizer/outline_aa.go` for `interface{}` usage (fixed - converted to generic color type)

#### [x] Update Renderer Primitives

- [x] Check `[ ] internal/renderer/primitives/primitives.go` for `interface{}` usage (already fixed in Phase 2)
- [x] Check `[ ] internal/renderer/markers/markers.go` for `interface{}` usage (already fixed in Phase 2)
- [x] Check `[ ] internal/renderer/outline/outline_aa.go` for `interface{}` usage (already fixed in Phase 2)

**Result:** All `interface{}` usage in rasterizer components has been successfully eliminated. The main changes were:

1. **`[ ] internal/rasterizer/outline.go`**:

   - Converted `OutlineRenderer` interface to `OutlineRenderer[C any]` with typed `LineColor(c C)` method
   - Converted `ColorStorage` interface to `ColorStorage[C any]` with typed `GetColor(index int) C` method
   - Converted `Controller` interface to `Controller[C any]` with typed `Color(pathIndex int) C` method
   - Updated `RasterizerOutline` to `RasterizerOutline[R OutlineRenderer[C], C any]` with all method receivers updated

2. **`[ ] internal/rasterizer/outline_aa.go`**:

   - Converted `OutlineAARenderer` interface to `OutlineAARenderer[C any]` with typed `Color(c C)` method
   - Updated `RasterizerOutlineAA` to `RasterizerOutlineAA[R OutlineAARenderer[C], C any]` with all method receivers updated
   - Updated calls to use the new generic `ColorStorage[C]` and `Controller[C]` interfaces

3. **Test file compatibility**: The legitimate use of `interface{}` in `[ ] internal/rasterizer/clip_test.go` was left unchanged as it's required for testing different generic instantiations.

All rasterizer tests pass, confirming that the type safety improvements maintain backward compatibility while eliminating runtime type assertions.

### Phase 6: Platform and Control Components

#### [x] Update Platform Support

- [x] `[ ] internal/platform/platform_support.go` - remove `interface{}` usage
- [x] `[ ] internal/platform/backend.go` - remove `interface{}` usage
- [x] `[ ] internal/platform/events.go` - remove `interface{}` usage
- [x] `[ ] internal/platform/rendering_context.go` - remove `interface{}` usage
- [x] `[ ] internal/platform/sdl2/sdl2_display.go` - remove `interface{}` usage
- [x] `[ ] internal/platform/x11/x11_display.go` - remove `interface{}` usage

#### [x] Update Control Components (Core Interface)

- [x] `[ ] internal/ctrl/ctrl.go` - remove `interface{}` usage (converted to `Ctrl[C any]`)
- [x] `[ ] internal/ctrl/types.go` - remove `interface{}` usage (converted `PathInfo[C]` and `VertexIterator[C]`)
- [x] `[ ] internal/ctrl/render.go` - remove `interface{}` usage (updated all render functions)
- [x] Update all control implementations:
  - [x] `[ ] internal/ctrl/bezier/bezier_ctrl.go`
  - [x] `[ ] internal/ctrl/checkbox/checkbox_ctrl.go`
  - [x] `[ ] internal/ctrl/gamma/gamma_ctrl.go`
  - [x] `[ ] internal/ctrl/polygon/polygon_ctrl.go`
  - [x] `[ ] internal/ctrl/rbox/rbox_ctrl.go`
  - [x] `[ ] internal/ctrl/scale/scale_ctrl.go`
  - [x] `[ ] internal/ctrl/spline/spline_ctrl.go`

### Phase 7: Example Applications

#### [x] Update Example Code

- [x] `examples/core/basic/embedded_fonts_hello/main.go`
- [x] `examples/core/basic/colors_gray/main.go`
- [x] `examples/core/basic/colors_rgba/main.go`
- [x] `examples/core/basic/shapes/main.go`
- [x] `examples/core/intermediate/rasterizers/main.go`
- [x] Any other examples using `interface{}`

### Phase 8: Utility and Helper Components

#### [x] Update Effects

- [x] `[ ] internal/effects/slight_blur.go` - removed `interface{}` usage
  - [x] Updated `PixelIterator` interface to `PixelIterator[T any]` with typed `Value() T`
  - [x] Updated `PixFmtInterface` to `PixFmtInterface[T any]` with typed pixel iterator
  - [x] Updated `SlightBlur` to `SlightBlur[PixFmt PixFmtInterface[T], T comparable]`
  - [x] Updated all function signatures to use typed generics
- [x] `[ ] internal/effects/stack_blur_optimized.go` - removed `interface{}` usage
  - [x] Updated `RGBImageInterface` to `RGBImageInterface[PtrType any]` with typed pointers
  - [x] Updated `RGBAImageInterface` to `RGBAImageInterface[PtrType any]` with typed pointers
  - [x] Updated `StackBlurRGB24` and `StackBlurRGBA32` function signatures to use generic constraints

#### [x] Update GPC (General Polygon Clipper)

- [x] `[ ] internal/gpc/gpc.go` - checked for `any` usage (only found in comments, no interface{} issues)

#### [x] Update Configuration

- [x] `[ ] internal/config/config.go` - removed `interface{}` usage
  - [x] Added `RenderingBufferInterface[T any]` common interface
  - [x] Updated `NewRenderingBuffer[T any]() interface{}` to `NewRenderingBuffer[T any]() RenderingBufferInterface[T]`
  - [x] Updated `NewRenderingBufferWithData[T any](...) interface{}` to return `RenderingBufferInterface[T]`

**Result:** All `interface{}` usage in utility and helper components has been successfully eliminated. The effects system now uses properly constrained generics with compile-time type safety. The configuration system uses a common rendering buffer interface instead of returning `interface{}`. All changes maintain backward compatibility while improving type safety.

### Phase 9: Testing and Validation

#### [x] Update Control Implementations to Use Generics

- [x] Convert CheckboxCtrl to generic `CheckboxCtrl[C any]` with typed Color method
- [x] Convert PolygonCtrl to generic `PolygonCtrl[C any]` with typed Color method
- [x] Convert BezierCtrl to generic `BezierCtrl[C any]` with typed Color method
- [x] Convert Curve3Ctrl to generic `Curve3Ctrl[C any]` with typed Color method
- [x] Convert RboxCtrl to generic `RboxCtrl[C any]` with typed Color method
- [x] Fix SplineCtrl Color method to return `ColorT` instead of `interface{}`
- [x] Create backward-compatible default constructors for all controls

#### [x] Update Test Files

- [x] Update all control `*_test.go` files to use default constructors
- [x] Remove mock implementations using `interface{}` type assertions
- [x] Fix tests that compared colors to nil or used type assertions
- [x] Ensure test coverage for generic implementations

#### [x] Phase 9 Core Fixes Completed

- [x] Convert GammaCtrl to generic `GammaCtrl[C any]` with typed `Color()` method
- [x] Convert SplineCtrl to generic `SplineCtrlImpl[C any]` with typed `Color()` method
- [x] Fix span_image_filter_rgba type assertions by updating constraints to `RGBASourceInterface`
- [x] Update `SpanImageFilterRGBA[Source RGBASourceInterface, ...]` constraints
- [x] Update `SpanImageFilterRGBABilinearClip[Source RGBASourceInterface, ...]` constraints
- [x] Eliminate runtime type assertions in RGBA span filters
- [x] All control interfaces now use compile-time type safety

#### [x] Validation Steps

- [x] Reduced `interface{}` usage in core control and span systems
- [x] All span filters compile successfully
- [x] All control Color() methods return typed colors instead of `interface{}`
- [x] Eliminated 3 runtime type assertions in span_image_filter_rgba.go
- [x] All control types properly generic with backward-compatible constructors

#### [x] Test File Interface Cleanup (Latest)

**Successfully Completed:**

- [x] **Rasterizer converter interface{} return types** - Fixed test files to handle typed returns from Upscale/Downscale methods properly (maintaining backward compatibility with existing interface{} API)
- [x] **Mock test renderer color types** - Updated all mock renderers to use generic color types `[C any]` instead of `interface{}`:
  - `MockBaseRenderer[C any]` in `internal/renderer/raster_text_test.go`
  - `MockOutlineRenderer[C any]` in `internal/renderer/outline_test.go`
  - `MockOutlineAARenderer[C any]` in `internal/renderer/outline_aa_test.go`
  - `MockColorStorage[C any]` and `MockColorStorageAA[C any]` for color storage tests
- [x] **Polymorphic test collections** - Replaced `[]interface{}` collections with type-safe generic helpers
- [x] **Test factory functions** - Updated factory functions to return `any` instead of `interface{}` (modern Go idiom)
- [x] **Platform interface compatibility** - Updated `OnPostDraw(rawHandler any)` in platform tests

**Results:**

- All test files compile and pass with full type safety
- Eliminated interface{} usage from all test mock implementations while maintaining test coverage
- Test code now uses compile-time type checking instead of runtime type assertions
- Backward compatibility maintained for production code while improving test code quality

#### [x] Phase 9 Implementation Completed

**Successfully Completed:**

- [x] **Fill rules rasterizer interface{}** - Converted `applyFillRuleToRasterizer(rasterizer interface{})` to use proper `FillingRuleSetter` interface in `internal/agg2d/fill_rules.go`
- [x] **Font system DataType() and Bounds() methods** - Updated `FontEngine` interface to return typed `GlyphDataType` and `basics.Rect[int]` instead of `interface{}` in `internal/font/cache_manager.go`
- [x] **FreeType font engine interface{} methods** - Fixed both stub and real implementations to use proper return types in `internal/font/freetype/`
- [x] **Pixfmt RGB blender type assertions** - Converted all RGB pixel formats to use proper `blender.RGBBlender`, `blender.RGB48Blender`, `blender.RGB96Blender` constraints, eliminating runtime type assertions
- [x] **Pixfmt RGBA blender type assertions** - Converted RGBA pixel formats to use proper `blender.RGBABlender` constraint, eliminating runtime type assertions
- [x] **Pixfmt RGB packed blender type assertions (partial)** - Replaced `interface{}` type aliases with concrete `NoBlender` type for plain formats

**Results:**

- Eliminated all remaining actionable `interface{}` usage from Phase 9 scope
- All pixfmt types now use compile-time type safety with proper generic constraints
- Font system uses typed interfaces instead of `interface{}` placeholders
- Fill rule system uses proper interface instead of runtime type assertions

#### [x] Test File Interface{} Cleanup (Phase 9 Final)

**Successfully Completed:**

- [x] **internal/ctrl/bezier/bezier_ctrl_test.go** - Replaced `interface{}` helper function with `fmt.Sprintf`
- [x] **internal/ctrl/bezier/curve3_ctrl_test.go** - Replaced `interface{}` helper function with `fmt.Sprintf`
- [x] **internal/transform/warp_magnifier_test.go** - Replaced `interface{}` with `any` for interface type assertion testing

**Results:**

- All test files compile and pass with full type safety
- Eliminated unnecessary `interface{}` helper functions in control tests
- Used modern Go `any` alias for legitimate polymorphic interface testing
- All Phase 9 test file objectives completed successfully

#### [ ] Remaining for Future Phases (Architectural)

- [x] **Pixfmt RGB packed interface{} usage** - Converted all RGB555/565/BGR555/BGR565 formats to use proper `blender.RGB16PackedBlender` constraints, eliminating runtime type assertions
- [ ] agg2d interface{} fields (marked as "to be updated in later phase")
- [x] **FreeType2 interface{} elimination** (COMPLETED) - Successfully replaced all `interface{}` usage in font/freetype2 code:
  - [x] FontEngineBase struct fields converted from `interface{}` to concrete scanline types
  - [x] FontEngineAdaptorTypes struct fields typed with proper generic adaptors
  - [x] CacheManager2 struct fields and methods now use typed scanline adaptors
  - [x] PathStorage interface created and properly used in engine.go
  - [x] All type assertions replaced with proper type checking
  - [x] All scanline components properly initialized in constructors
  - [x] Package compiles successfully with build tag `freetype`

- Go files containing interface{} (line numbers):
  - [ ] internal/agg2d/agg2d.go: line 98, 134, 135
  - [ ] internal/agg2d/text.go
  - [x] internal/ctrl/bezier/bezier_ctrl_test.go: line 241 (COMPLETED)
  - [x] internal/ctrl/bezier/curve3_ctrl_test.go: line 276 (COMPLETED)
  - [ ] internal/fonts/interfaces.go: line 2
  - [ ] internal/platform/interfaces.go: line 15, 31
  - [ ] internal/renderer/scanline/interfaces.go: line 60
  - [x] internal/transform/warp_magnifier_test.go: line 320 (COMPLETED)

## Implementation Notes

### Pattern to Follow

**Bad (current):**

```go
type PixelFormat interface {
    CopyPixel(x, y int, c interface{})
}

func (p *SomePixFmt) CopyPixel(x, y int, c interface{}) {
    if color, ok := c.(color.RGBA8[color.Linear]); ok {
        // actual implementation
    }
}
```

**Good (target):**

```go
type PixelFormat[C any] interface {
    CopyPixel(x, y int, c C)
}

type SomePixFmt[C any] struct {
    // fields
}

func (p *SomePixFmt[C]) CopyPixel(x, y int, c C) {
    // direct implementation, no type assertion needed
}
```

### Common Color Types to Support

- `color.RGBA8[CS any]` - 8-bit RGBA with colorspace
- `color.Gray8[CS any]` - 8-bit grayscale with colorspace
- `color.RGBA` - floating-point RGBA
- Custom color types as needed

### Zero Value Detection Pattern

Instead of:

```go
if spanGen := interface{}(sc.spanGen); spanGen != nil {
    sc.spanGen.Prepare()
}
```

Use:

```go
// Option 1: Add IsZero method to interface
if !sc.spanGen.IsZero() {
    sc.spanGen.Prepare()
}

// Option 2: Use reflection (less preferred)
if !reflect.ValueOf(sc.spanGen).IsZero() {
    sc.spanGen.Prepare()
}
```

## Success Criteria

1. No `interface{}` usage except where absolutely necessary for true dynamic typing
2. All `any` type parameters have appropriate constraints
3. No runtime type assertions or type switches for color/pixel operations
4. Full type safety at compile time
5. All tests pass
6. All examples compile and run correctly
7. Performance equal or better than before (no boxing overhead)

## References

- Original C++ AGG source: `/home/christian/Code/agg-2.6/agg-src/`
- Go generics documentation: <https://go.dev/doc/tutorial/generics>
- Project guidelines: `/home/christian/Code/agg_go/CLAUDE.md`
