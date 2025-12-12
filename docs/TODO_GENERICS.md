# AGG Go Generics Refactoring Status

## ✅ Completed

### Core RGBA8 Format Refactoring

- [x] **Define slimmer blender contract**
  - [x] `RGBABlender[S color.Space]` interface with `GetPlain/SetPlain/BlendPix`
  - [x] `RawRGBAOrder` interface for fast path optimization
  - [x] All concrete blenders implement interfaces properly

- [x] **Simplify RGBA8 pixfmt type parameters**
  - [x] `PixFmtAlphaBlendRGBA[S color.Space, B blender.RGBABlender[S]]`
  - [x] Remove order parameter `O` from pixfmt
  - [x] Update constructor signatures
  - [x] Update all method receivers

- [x] **Remove direct dependence on order in RGBA8 pixfmt**
  - [x] Replace all `idxs[O]()` usage with blender methods
  - [x] Primary path via `GetPlain/SetPlain/BlendPix`
  - [x] Fast path via `RawRGBAOrder` type assertion (e.g., in `CopyHline`)
  - [x] All core methods migrated

- [x] **Update RGBA8 type aliases & constructors**
  - [x] Aliases choose order via blender: `PixFmtRGBA32[S] = PixFmtAlphaBlendRGBA[S, BlenderRGBA8[S, order.RGBA]]`
  - [x] Helper constructors updated with new signatures
  - [x] Linear convenience constructors preserved

### Gray8 Format Refactoring

- [x] **Update Gray8 generic constraints**
  - [x] `PixFmtAlphaBlendGray[CS color.Space, B blender.GrayBlender[CS]]`
  - [x] Remove `[B any, CS color.Space]` pattern
  - [x] All method signatures updated
  - [x] Type aliases and constructors fixed
  - [x] Remove type assertions to specific blenders

### RGB8 24-bit Format Core

- [x] **Basic RGB24 format refactoring**
  - [x] `RGBBlender[S color.Space]` interface defined
  - [x] Remove `idxsRGB[O]()` usage from core methods
  - [x] `GetPixel/CopyPixel` use blender interface
  - [x] Type aliases for RGB24/BGR24 normal and Pre variants

---

## 🚧 TODO

### Completed Items (Remaining Issues)

- [x] **RGB16 format (48-bit) - COMPLETED**
  - [x] Remove `idxsRGB[O]()` usage
  - [x] Update `RGB48Blender` interface from `[S, O]` to `[S]`
  - [x] Apply same pattern as RGB8

- [x] **RGBX32 format (32-bit with padding) - COMPLETED**
  - [x] Identified need for specialized RGBX blender interface
  - [x] Implemented `RGBXBlender[S]` interface for 4-byte RGB with padding
  - [x] Created `BlenderRGBX8[S, O]` and `BlenderRGBXPre[S, O]` implementations
  - [x] Updated `PixFmtAlphaBlendRGBX32` to use new blender interface
  - [x] Fixed type aliases and constructors for all RGBX32 variants
  - [x] Added premultiplied RGBX32 variants with proper constructors

### 16-bit Format Updates

- [ ] **RGBA16 formats - NEEDS EXTENSIVE REFACTORING**
  - [x] `RGBABlender16` interface updated to remove order parameter
  - [ ] Apply same refactoring pattern as RGBA8 to pixfmt_rgba16.go
  - ⚠️ **Note**: Requires complete rewrite of method signatures - currently needs major work

- [x] **Gray16/Gray32 formats - COMPLETED**
  - [x] Ensured interface alignment with Gray8 pattern
  - [x] Added missing GetPlain/SetPlain methods
  - [x] Generic constraints consistency verified

### Packed Format Alignment

- [x] **RGB packed formats (RGB555/565) - COMPLETED**
  - [x] Verified blender interfaces align with new pattern
  - [x] Confirmed no order parameter dependencies

### AGG Compatibility Features

- [x] **Whole-buffer utilities - COMPLETED**
  - [x] `Premultiply()` method on RGBA pixfmt types (no-op for RGB)
  - [x] `Demultiply()` method on RGBA pixfmt types (no-op for RGB)
  - [x] `ApplyGammaDir(g)` and `ApplyGammaInv(g)` methods
  - [x] Fast path implementation with fallback to `GetPlain/SetPlain`

### Extended Features (Optional)

- [ ] **Consider additional color spaces**
  - [ ] Evaluate need beyond `Linear/SRGB` (e.g., Lab, HSV)
  - [ ] Extend `color.Space` constraint if needed

- [ ] **Performance optimizations**
  - [ ] Profile blender interface vs direct access performance
  - [ ] Optimize hot paths identified in benchmarks

### Summary Status

**Core Work: ✅ COMPLETED**

- RGB16, RGB24, RGBX32, Gray16/Gray32, and RGB packed formats all successfully refactored
- Whole-buffer utility methods (Premultiply, Demultiply, ApplyGamma) added
- Core pixfmt library builds cleanly after refactoring

**Known Issues:**

- RGBA16 format needs extensive refactoring (currently commented out)
- Some examples need updates for new generic signatures
- Build succeeds for core library, examples have compatibility issues

**Next Steps (Optional):**

- Complete RGBA16 format refactoring following RGBA8 pattern
- Update example code to use new generic signatures

---

## Architecture Summary

**Design Principle:** Blenders encapsulate pixel ordering; pixfmt uses interface with optional fast paths.

**Pattern Established:**

```go
// Interface constrains behavior
type RGBABlender[S color.Space] interface {
    GetPlain(px []byte) (r, g, b, a basics.Int8u)
    SetPlain(px []byte, r, g, b, a basics.Int8u)
    BlendPix(px []byte, r, g, b, a, cover basics.Int8u)
}

// Pixfmt depends only on interface
type PixFmtAlphaBlendRGBA[S color.Space, B RGBABlender[S]] struct { ... }

// Fast path when available
if ro, ok := any(pf.blender).(RawRGBAOrder); ok { /* use indices */ }
```

This maintains AGG compatibility while providing Go type safety and performance.
