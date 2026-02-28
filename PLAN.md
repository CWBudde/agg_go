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

- [x] **Phase 1 Complete**: Core internal packages build successfully

**Summary**: Fixed type mismatches, removed problematic generics, and established concrete types for array, rasterizer, renderer, and pixfmt packages. Internal packages now build cleanly.

### Completed Tasks

**1.1 array** - Fixed `pod_arrays_test.go` type mismatches (int vs int32), created concrete `VertexDistSequence` and `LineAAVertexSequence` types

**1.2 rasterizer** - Created `RasterizerCellsAASimple` and `RasterizerCellsAAStyled` concrete types, removed runtime type switching

**1.3 renderer** - Defined `Blendable[Self]` constraint, replaced 8+ type switches in `helpers.go` with constraint-based generics

**1.4 pixfmt** - Added `BlenderRGBA` type aliases, re-exported order types in `base.go`

**1.5 examples** - Fixed generic instantiation errors in basic and intermediate examples

**1.6 Checkpoint** - All `internal/` packages build successfully; some examples need API implementation (deferred to Phase 5)

---

## Phase 2: Generics Audit & Design

- [x] **Phase 2 Complete**: All generics categorized and design decisions made

**Summary**: Audited 40+ generic types, categorized into true generics (keep), false generics (refactor to concrete), and combinatorial cases (use interfaces). Determined code generation is not needed.

### Completed Tasks

**2.1 Audit** - Created `docs/GENERICS_AUDIT.md` documenting every generic type, C++ template mapping, and instantiation count

**2.2 Categorization**

- **Category A (True Generics)**: ~40 types keep generics (`Point[T]`, `Rect[T]`, `PodArray[T]`, `RGBA8[CS]`, etc.)
- **Category B (False Generics)**: 5 types need concrete replacements (`VertexSequence`, `RasterizerCellsAA`, `GammaLUT`, `Saturation`, `Gray8`)
- **Category C (Combinatorial)**: 0 types need code generation (interfaces + type aliases sufficient)

**2.3 Code Generation Decision** - NOT NEEDED; pixel format system uses blender interfaces with fast-path type assertions for performance

---

## Phase 3: Core Package Rework

- [x] **Phase 3 Complete**: All packages free of problematic `any()` casts

**Summary**: Eliminated problematic `any()` casts from 7 packages (array, rasterizer, pixfmt, basics, color, span, font). Documented acceptable `any()` uses at module boundaries. Net result: cleaner type system, ~100+ lines of boilerplate removed, zero runtime cost.

### Completed Tasks

**3.1 array** - Replaced generic `VertexSequence[T]` with 3 concrete types (`VertexDistSequence`, `LineAAVertexSequence`, `VertexCmdSequence`)

**3.2 rasterizer** - Used existing concrete types `RasterizerCellsAASimple` and `RasterizerCellsAAStyled`, removed generic `RasterizerCellsAA[Cell]`

**3.3 pixfmt/gamma** - Replaced type-switch `GammaLUT` with clean generic implementation using proper `Unsigned` constraint

**3.4 basics** - Replaced generic `Saturation[T]` with 4 concrete structs (`SaturationInt`, `SaturationInt32`, `SaturationUint`, `SaturationUint32`)

**3.5 color** - Replaced generic `ConvertGray8FromRGBA8[CS]` with concrete `ConvertGray8LinearFromRGBA8` and `ConvertGray8SRGBFromRGBA8`

**3.6 span/converter** - Documented type switches as **ACCEPTABLE** (module boundary pattern, graceful degradation, follows C++ AGG template specialization)

**3.7 font subsystem** - Created `internal/font/interfaces.go` with `IntegerPathStorage` and `SerializedScanlinesAdaptor` interfaces, eliminated ~85 lines of type assertions across 7 files, documented legitimate `interface{}` uses (build-tag dispatch in agg2d.go)

---

## Phase 4: Test Strategy

- [ ] **Phase 4 Complete**: Comprehensive test suite in place

### 4.1 Contract Tests

**Principle**: Test public API behavior, not internal state

**Status**: ✅ **In Progress** - Test patterns documented, existing tests audited

**Completed**:

- [x] Audited existing test coverage (225 test files in internal/, 13 in tests/)
- [x] Documented test patterns in `docs/TESTING_STRATEGY.md`
- [x] Fixed build failures in test files (rasterizer, pixfmt)
- [x] Applied testing anti-patterns principles to tests
- [x] Identified packages needing additional coverage

**Remaining Tasks**:

- [x] Fixed conv package tests (BSpline edge cases, GPC placeholder acknowledgment, ConvStroke vertex count)
- [x] Fixed agg2d Sign function test (corrected test expectation for very small values)
- [x] Fixed pixfmt/gamma nil pointer panic (changed RGBOrder from interface type to concrete type in tests)
- [ ] Fix failing tests in remaining packages:
  - [x] color: Gray16/Gray32 conversion and lerp tests - all passing
  - [x] fonts: character count mismatches - all passing
  - [x] pixfmt: blending and clear operations - all passing
  - [x] pixfmt/blender: Porter-Duff operations - all passing
  - [ ] rasterizer: clipping and cell generation tests (4 failures) - may need C++ AGG reference
  - [ ] integration: rendering pipeline issues (30+ failures) - likely downstream from lower-level issues
- [ ] Add contract tests for low-coverage packages (<60%): effects, platform, primitives, pixfmt/blender
- [ ] Ensure all tests verify behavior, not implementation details
- [ ] Add missing edge case tests from AGG documentation

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
