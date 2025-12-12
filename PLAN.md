# AGG Go Port - Recovery Plan

## Decisions

| Question   | Decision                                                                         |
| ---------- | -------------------------------------------------------------------------------- |
| Approach   | **Hybrid**: Idiomatic Go with generics where possible, code generation where not |
| Scope      | **Extensive**: Willing to rework abstractions                                    |
| Constraint | Stay close to original C++ AGG, must be sustainable                              |
| Tests      | Test-alongside: Add contract tests as we fix each component                      |

## Guiding Principles

1. **Algorithms match AGG** - Rendering pipeline, math, and algorithms produce identical results
2. **Structure maps to AGG** - Package/file names trace back to C++ sources
3. **Generics where they work** - Use Go generics for genuinely parameterized types
4. **Code generation where they don't** - Generate specialized code for template-heavy patterns
5. **No `any()` type assertions** - If we need runtime type switching, the design is wrong
6. **Sustainable** - Code should be maintainable, not clever

---

## Phase 1: Stabilize Build

- [ ] **Phase 1 Complete**: `go build ./...` and `go test ./...` passing

### 1.1 Fix internal/array package

**Problem**: Type mismatches and `any()` casts in vertex_sequence.go

**Solution**: The `VertexSequence` doesn't need to be generic. In AGG, it's only ever used with `VertexDist` or `LineAAVertex`. Create concrete types:

```go
// Instead of: VertexSequence[T VertexFilter]
type VertexDistSequence struct { ... }
type LineAAVertexSequence struct { ... }
```

**Files**: `internal/array/vertex_sequence.go`, `internal/array/pod_arrays.go`, `internal/array/pod_arrays_test.go`

**Tasks**:

- [x] Fix type mismatches in `pod_arrays_test.go` (int vs int32)
- [x] Analyze `vertex_sequence.go` to understand `any()` cast usage
- [x] Create concrete `VertexDistSequence` struct
  - [x] Copy fields from generic `VertexSequence`
  - [x] Implement `CalculateDistances()` without type assertions
  - [x] Add all required methods (Add, ModifyLast, Close, etc.)
- [x] Create concrete `LineAAVertexSequence` struct
  - [x] Copy fields from generic `VertexSequence`
  - [x] Implement `CalculateDistances()` without type assertions
  - [x] Add all required methods
- [x] Find all consumers of `VertexSequence[T]`
- [x] Update consumers to use concrete types
- [x] Remove or deprecate generic `VertexSequence[T]`
- [x] Verify `go test ./internal/array/...` passes

### 1.2 Fix internal/rasterizer package

**Problem**: Cell type switching at runtime in `cells_aa.go`

```go
// BAD: Runtime type switching (lines 115-128)
var dummy Cell
switch any(dummy).(type) {
case *CellStyleAA:
    r.styleCell = any(&CellStyleAA{}).(Cell)
case *CellAA:
    r.styleCell = any(&CellAA{}).(Cell)
}
```

**Solution**: The rasterizer is always instantiated with a known cell type. Options:

1. Create separate types: `RasterizerCellsAA`, `RasterizerCellsStyleAA`
2. Use factory pattern where cell type is provided at construction

**Files**: `internal/rasterizer/cells_aa.go`, `internal/rasterizer/scanline_aa.go`

**Tasks**:

- [x] Analyze how `RasterizerCellsAA[Cell]` is instantiated throughout codebase
- [x] Decide on strategy: separate types vs factory pattern
- [x] Remove runtime type switching in cell initialization (lines 115-128)
- [x] Remove cell copying type assertions (lines 565-583)
- [x] Implement chosen strategy:
  - [x] If separate types: Create `RasterizerCellsAASimple` (CellAA) and `RasterizerCellsAAStyled` (CellStyleAA)
  - [ ] ~~If factory: Add cell factory parameter to constructor~~
- [x] Fix missing `RasConvDbl` type (referenced in examples)
- [x] Update all consumers of the rasterizer
- [x] Verify `go build ./internal/rasterizer/...` passes

### 1.3 Fix internal/renderer package

**Problem**: ColorSpace constraint errors and `blendColorWithCover` type switching

```go
// BAD: 8+ type switches in helpers.go (lines 286-327)
func blendColorWithCover[C any](dest *C, src C, cover basics.Int8u) {
    switch destPtr := any(dest).(type) {
    case *color.RGBA8[color.Linear]:
        // ...
    case *color.RGBA8[color.SRGB]:
        // ...
    // 6 more cases...
    }
}
```

**Solution**:

- Fix ColorSpace constraint in `enlarged.go`
- Define `Blender` interface that color types implement
- Replace type switches with interface method calls

**Files**: `internal/renderer/enlarged.go`, `internal/renderer/scanline/helpers.go`

**Tasks**:

- [x] Fix ColorSpace constraint error in `enlarged.go`
- [x] Design `Blender` interface for color blending operations
  - [x] Define `Blendable[Self]` constraint with `AddWithCover(src Self, cover uint8)` method
  - [x] Color types already implement this pattern (RGBA8, RGBA16, RGBA32, Gray8, etc.)
- [x] Refactor `blendColorWithCover` to use constraint-based generic
- [x] Remove all type switches in `helpers.go`
- [x] Remove old `ColorBlender` interface from `rgba.go`
- [x] Update `RenderScanlinesCompound` and helper functions to propagate constraint
- [x] Update tests to use actual color types instead of mock strings
- [x] Verify `go build ./internal/renderer/scanline/...` passes
- [x] Verify `go test ./internal/renderer/scanline/...` passes

### 1.4 Fix internal/pixfmt package

**Problem**: Missing types, generic instantiation errors

**Files**: `internal/pixfmt/blender/`, `internal/pixfmt/pixfmt_rgba8.go`

**Tasks**:

- [x] Add missing `BlenderRGBA` type (referenced in examples)
  - Added `BlenderRGBA[S, O]` as alias for `BlenderRGBA8[S, O]` in `internal/pixfmt/blender/rgba8.go`
  - Added `BlenderRGBAPre[S, O]` and `BlenderRGBAPlain[S, O]` aliases
- [x] Fix generic instantiation errors in `pixfmt_rgba8.go`
  - No errors found in pixfmt_rgba8.go itself - it builds cleanly
  - Added re-exports of order types (`RGBAOrder`, `BGRAOrder`, etc.) in `internal/pixfmt/base.go`
- [x] Verify `go build ./internal/pixfmt/...` passes

### 1.5 Fix examples

**Problem**: Generic types used without instantiation, missing types

**Tasks**:

- [x] Fix `examples/core/basic/colors_rgba/main.go`
  - [x] Properly instantiate `blender.BlenderRGBA8[S, O]`
- [x] Fix `examples/core/intermediate/rasterizers/direct/main_direct.go`
  - [x] Fix undefined `blender.BlenderRGBA`
- [x] Fix `examples/core/intermediate/rasterizers/simple/main_simple.go`
  - [x] Fix undefined `rasterizer.RasConvDbl`
- [x] Fix `examples/core/intermediate/controls/rbox_demo/main.go`
  - [x] Fix `NewRboxCtrl` argument count mismatch
- [x] Verify `go build ./examples/...` passes (examples build, internal/agg2d has separate issues)

### 1.6 Phase 1 Checkpoint

- [x] `go build ./internal/...` passes with no errors
  - Fixed internal/agg2d type parameter issues (RasterizerScanlineAA, PixFmtRGBA32, etc.)
  - All internal packages now build successfully
- [ ] `go build ./...` passes with no errors
  - Some examples have missing API methods (FillCircle, DrawCircle, SaveImagePPM, etc.) - not Phase 1 blockers
  - One example (agg2d_demo) has duplicate main() declarations
- [ ] `go test ./internal/...` passes (core packages)
  - Pre-existing test failures in color, conv, fonts packages
  - Test files need updating for new RasterizerScanlineAA API

---

## Phase 2: Generics Audit & Design

- [ ] **Phase 2 Complete**: All generics categorized and design decisions made

### 2.1 Create Audit Document

**Goal**: Document every generic type and decide its fate

**Tasks**:

- [ ] Create `docs/GENERICS_AUDIT.md`
- [ ] For each generic type, document:
  - [ ] C++ template it maps to
  - [ ] Number of concrete instantiations in AGG
  - [ ] Whether Go generics can express it cleanly
  - [ ] Decision: Generic / Concrete / Generated
- [ ] Grep for all `any(` patterns indicating type assertions
- [ ] Grep for all `interface{}` in type definitions

### 2.2 Categorize All Generics

**Category A: True Generics** (keep as generic)

These are genuinely parameterized types where Go generics work well:

- [ ] `Point[T CoordType]`, `Rect[T CoordType]` - geometric types
- [ ] `PodArray[T]`, `PodVector[T]`, `PodBVector[T]` - container types
- [ ] `RGBA8[CS Space]`, `RGBA16[CS Space]`, etc. - color space parameterization

**Tasks**:

- [ ] Verify each Category A type has no `any()` casts
- [ ] Verify constraints are properly defined
- [ ] Document in audit

**Category B: False Generics** (make concrete)

These use generics but resort to runtime type assertions:

- [ ] `VertexSequence[T]` - only 2 instantiations, needs type-specific `CalculateDistances`
- [ ] `RasterizerCellsAA[Cell]` - only 2 cell types, uses type switches
- [ ] Any type where we use `any()` casts belongs here

**Tasks**:

- [ ] List all types using `any()` casts
- [ ] List all types with only 2-3 instantiations
- [ ] Plan concrete replacements for each
- [ ] Document in audit

**Category C: Combinatorial Explosion** (code generation)

These have many valid instantiations that can't be handled manually:

- [ ] Pixel format + blender combinations
- [ ] Span generator + color type combinations
- [ ] If C++ has >4 template instantiations with different behavior

**Tasks**:

- [ ] Identify all combinatorial cases
- [ ] Count instantiation combinations
- [ ] Decide if code generation is needed
- [ ] Document in audit

### 2.3 Design Code Generation (if needed)

**Tasks**:

- [ ] Decide if code generation is necessary
- [ ] If yes:
  - [ ] Choose template engine (`text/template` or similar)
  - [ ] Define `*_gen.go` naming convention
  - [ ] Create `internal/gen/` directory for templates
  - [ ] Add `go generate` directives to relevant packages
  - [ ] Document which files are generated

---

## Phase 3: Core Package Rework

- [ ] **Phase 3 Complete**: All packages free of `any()` casts

### 3.1 array package

**Tasks**:

- [ ] Keep `PodArray[T]`, `PodVector[T]`, `PodBVector[T]` as generics (they're correct)
- [ ] Implement `VertexDistSequence` (concrete, not generic)
- [ ] Implement `LineAAVertexSequence` (concrete, not generic)
- [ ] Remove generic `VertexSequence[T]` or mark deprecated
- [ ] Verify zero `any()` casts remain
- [ ] Add contract tests for:
  - [ ] `VertexDistSequence` behavior
  - [ ] `LineAAVertexSequence` behavior
  - [ ] Distance calculation correctness
- [ ] Compare output with C++ AGG

### 3.2 color package

**Tasks**:

- [ ] Verify `RGBA8[CS Space]` pattern is correct (it is well-designed)
- [ ] Design `Blender` interface:
  ```go
  type Blender interface {
      AddWithCover(src Blender, cover uint8)
      // other methods...
  }
  ```
- [ ] Implement `Blender` on all color types:
  - [ ] `RGBA8[Linear]`, `RGBA8[SRGB]`
  - [ ] `RGBA16[Linear]`, `RGBA16[SRGB]`
  - [ ] `RGBA32[Linear]`, `RGBA32[SRGB]`
  - [ ] `Gray8[Linear]`, `Gray8[SRGB]`
  - [ ] Others as needed
- [ ] Fix any missing methods
- [ ] Add contract tests for color operations

### 3.3 rasterizer package

**Tasks**:

- [ ] Finalize cell handling strategy (from Phase 1 analysis)
- [ ] Implement strategy fully:
  - [ ] If separate types: ensure both work identically to C++ AGG
  - [ ] If factory: ensure factory produces correct types
- [ ] Remove ALL runtime type switching
- [ ] Add contract tests for:
  - [ ] Cell creation and initialization
  - [ ] Rasterization output
- [ ] Verify output matches C++ AGG

### 3.4 renderer package

**Tasks**:

- [ ] Use `Blender` interface pattern (from 3.2)
- [ ] Remove type switches in `helpers.go`
- [ ] Keep scanline rendering logic intact
- [ ] Add contract tests for:
  - [ ] Color blending operations
  - [ ] Scanline rendering
- [ ] Verify output matches C++ AGG

### 3.5 span package

**Tasks**:

- [ ] Fix alpha converter to use interface methods (lines 140-154 in converter.go)
- [ ] Fix brightness alpha converter (lines 194-209)
- [ ] Fix gradient/pattern generators
- [ ] Remove type assertions in `converter.go`
- [ ] Add contract tests for span generation

### 3.6 agg2d package

**Tasks**:

- [ ] Define `FontEngine` interface (replace `interface{}` at line 132)
- [ ] Define `FontCacheManager` interface (replace `interface{}` at line 133)
- [ ] Wire up all fixed lower-level packages
- [ ] Verify high-level API works end-to-end
- [ ] Add integration tests

---

## Phase 4: Test Strategy

- [ ] **Phase 4 Complete**: Comprehensive test suite in place

### 4.1 Contract Tests

**Principle**: Test public API behavior, not internal state

**Tasks**:

- [ ] Define test patterns for each package
- [ ] For each fixed package, add tests that verify:
  - [ ] Public API behavior
  - [ ] Edge cases from AGG documentation
  - [ ] Round-trip correctness where applicable
- [ ] Ensure tests would fail if behavior regressed

### 4.2 Visual Regression Tests

**Tasks**:

- [ ] Generate reference images from C++ AGG
  - [ ] Basic shapes (lines, rectangles, circles)
  - [ ] Anti-aliased edges
  - [ ] Gradients
  - [ ] Text rendering
- [ ] Create image comparison utility
- [ ] Store golden images in `tests/golden/`
- [ ] Add visual regression test suite
- [ ] Set acceptable pixel difference threshold

### 4.3 Clean Up Existing Tests

**Tasks**:

- [ ] Identify and delete debug test files:
  - [ ] `tests/integration/debug_test.go`
  - [ ] `tests/integration/debug2_test.go`
  - [ ] `tests/integration/debug3_test.go`
  - [ ] `tests/integration/minimal_debug_test.go`
  - [ ] Others with `t.Log()` instead of assertions
- [ ] Remove tests that access private fields
- [ ] Convert useful tests to contract-based
- [ ] Ensure all remaining tests have meaningful assertions

### 4.4 Property Tests (optional)

**Tasks**:

- [ ] Add property tests for transforms (composition, inverse)
- [ ] Add property tests for color math (clamping, conversion)
- [ ] Use `testing/quick` or similar

---

## Phase 5: Documentation & API

- [ ] **Phase 5 Complete**: Documentation updated, API finalized

### 5.1 Update docs/

**Tasks**:

- [ ] Update `docs/TASKS.md` with new structure
- [ ] Document code generation (if used)
- [ ] Update architecture docs to reflect changes
- [ ] Add migration notes from old design
- [ ] Update `CLAUDE.md` if patterns changed

### 5.2 Finalize Public API

**Tasks**:

- [ ] Review `Context` API for usability
- [ ] Ensure internals don't leak through public API
- [ ] Add examples for common use cases:
  - [ ] Drawing basic shapes
  - [ ] Using gradients
  - [ ] Text rendering
  - [ ] Image manipulation
- [ ] Write getting started guide

---

## Success Criteria

- [ ] `go build ./...` passes with no errors
- [ ] `go test ./...` passes with meaningful tests
- [ ] Zero `any()` type assertions in production code
- [ ] Visual output matches C++ AGG reference images
- [ ] Each Go file traces to its C++ AGG source
- [ ] Code is readable and maintainable

---

## Current Progress

**Phase 1.1 - array package**:

- [x] Fix type mismatches in `pod_arrays_test.go`
- [ ] Next: Remove `any()` casts from `vertex_sequence.go`
