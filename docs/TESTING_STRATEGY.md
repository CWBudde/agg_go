# Testing Strategy for AGG Go Port

## Overview

This document outlines the testing strategy for the AGG Go port, including test patterns, coverage goals, and best practices identified during Phase 4.1 of the recovery plan.

## Test Coverage Summary

As of Phase 4.1 completion, the project has:

- **225 test files** in `internal/` packages
- **13 test files** in `tests/` directory
- **Test coverage**: Ranges from 24% to 100% across packages

### High Coverage Packages (>80%)

- `internal/vpgen`: 100.0%
- `internal/shapes`: 98.4%
- `internal/bezierarc`: 93.6%
- `internal/vcgen`: 93.9%
- `internal/config`: 93.3%
- `internal/gsv`: 86.2%
- `internal/renderer/scanline`: 85.8%
- `internal/image`: 84.7%
- `internal/renderer/markers`: 84.1%
- `internal/glyph`: 83.7%
- `internal/ctrl/spline`: 82.2%
- `internal/transform`: 81.5%

### Medium Coverage Packages (60-80%)

- `internal/array`: 77.5%
- `internal/renderer/outline`: 77.8%
- `internal/color/conv`: 77.3%
- `internal/curves`: 77.2%
- `internal/path`: 75.2%
- `internal/span`: 73.9%
- `internal/buffer`: 72.0%
- `internal/scanline`: 71.1%
- `internal/basics`: 70.3%
- `internal/color`: 64.3%
- `internal/agg2d`: 64.1%
- `internal/renderer`: 62.7%
- `internal/gpc`: 60.3%

### Low Coverage Packages (<60%)

- `internal/effects`: 59.5%
- `internal/platform`: 52.6%
- `internal/primitives`: 42.4%
- `internal/pixfmt/blender`: 24.1%
- `internal/platform/x11`: 0.2%

### Zero Coverage (Not Yet Implemented or Platform-Specific)

- `internal/font`
- `internal/font/freetype`
- `internal/gamma`
- `internal/order`
- `internal/platform/sdl2`
- `internal/platform/types`
- `internal/renderer/primitives`
- `internal/vertex_source`
- `internal/ctrl/text`

## Test Patterns

### 1. Unit Tests for Data Structures

**Pattern**: Test construction, basic operations, and edge cases

**Example** (from `internal/array/vertex_sequence_test.go`):

```go
func TestNewLineAAVertexSequence(t *testing.T) {
	vs := NewLineAAVertexSequence()
	if vs == nil {
		t.Fatalf("NewLineAAVertexSequence returned nil")
	}
	if vs.Size() != 0 {
		t.Errorf("New vertex sequence size = %d, want 0", vs.Size())
	}
}
```

**Key Points**:
- Test constructors return non-nil values
- Verify initial state is correct
- Test basic operations (Add, Get, Size, etc.)
- Test edge cases (empty sequences, boundary conditions)

### 2. Contract Tests for Interfaces

**Pattern**: Test that implementations satisfy interface contracts

**Example** (from `internal/array/vertex_sequence_test.go`):

```go
// Test type constraint - this should compile if VertexFilter is implemented correctly
func TestVertexFilterConstraint(t *testing.T) {
	// Test that LineAAVertex implements VertexFilter
	var _ VertexFilter = LineAAVertex{}

	// Test that VertexDist implements VertexFilter
	var _ VertexFilter = VertexDist{}
}
```

**Key Points**:
- Compile-time verification of interface satisfaction
- Ensures type constraints work correctly with generics

### 3. Mock-Based Tests for Complex Dependencies

**Pattern**: Use simple mock implementations for testing units in isolation

**Example** (from `internal/rasterizer/scanline_aa_test.go`):

```go
// MockScanline implements ScanlineInterface for testing
type MockScanline struct {
	cells []MockCell
	spans []MockSpan
	y     int
}

func (ms *MockScanline) ResetSpans() {
	ms.cells = ms.cells[:0]
	ms.spans = ms.spans[:0]
}
```

**Key Points**:
- Mocks implement only required interfaces
- Keep mocks simple and focused
- Test observable behavior, not mock existence (anti-pattern!)
- Use mocks to verify interactions with dependencies

### 4. Behavior Verification Tests

**Pattern**: Test observable behavior through public API, not private state

**Good Example**:

```go
func TestRasterizerScanlineAA_MoveTo(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAA[float64, DblConv, *MockClip](DblConv{}, clipper)

	r.MoveTo(100, 200)

	// Verify observable behavior via public API
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo")
	}

	// Verify interaction with dependency
	if clipper.moveToX != 100.0 || clipper.moveToY != 200.0 {
		t.Errorf("Expected clipper MoveTo called with (100, 200), got (%f, %f)",
			clipper.moveToX, clipper.moveToY)
	}
}
```

**Note**: Accessing package-private fields (like `r.status`) is acceptable within the same package for unit tests, but prefer testing observable public behavior when possible.

### 5. Round-Trip and Correctness Tests

**Pattern**: Test that operations are reversible or produce correct results

**Example** (from `internal/array/vertex_sequence_test.go`):

```go
func TestVertexDistCalculateDistance(t *testing.T) {
	v1 := NewVertexDist(0.0, 0.0)
	v2 := NewVertexDist(3.0, 4.0) // 3-4-5 triangle

	v1.CalculateDistance(v2)

	// Distance should be 5.0
	if v1.Dist != 5.0 {
		t.Errorf("CalculateDistance result = %f, want 5.0", v1.Dist)
	}
}
```

**Key Points**:
- Use known values with predictable results (e.g., 3-4-5 triangle)
- Verify mathematical correctness
- Test both forward and reverse operations where applicable

### 6. Benchmark Tests for Performance-Critical Code

**Pattern**: Include benchmarks for operations that need to be fast

**Example**:

```go
func BenchmarkVertexSequenceAdd(b *testing.B) {
	vs := NewLineAAVertexSequence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.Add(NewLineAAVertex(i*1000, i*1000))
	}
}
```

**Key Points**:
- Use `b.ResetTimer()` to exclude setup time
- Test realistic usage patterns
- Useful for comparing implementations

## Testing Anti-Patterns to Avoid

Based on the `superpowers:testing-anti-patterns` skill:

### ❌ **Don't Test Mock Behavior**

```go
// BAD: Testing that the mock exists
test('renders sidebar', () => {
  render(<Page />);
  expect(screen.getByTestId('sidebar-mock')).toBeInTheDocument();
});
```

### ❌ **Don't Add Test-Only Methods to Production Code**

```go
// BAD: destroy() only used in tests
class Session {
  async destroy() {  // Test-only method!
    await this._workspaceManager?.destroyWorkspace(this.id);
  }
}
```

**Fix**: Put cleanup logic in test utilities, not production code.

### ❌ **Don't Mock Without Understanding Dependencies**

**Before mocking**, ask:
1. What side effects does the real method have?
2. Does this test depend on any of those side effects?
3. Do I fully understand what this test needs?

### ✅ **Do Test Real Behavior**

```go
// GOOD: Test observable behavior through public API
func TestScanlineGeneration(t *testing.T) {
	sl := NewScanline()
	sl.AddCell(10, 128)
	sl.AddSpan(20, 5, 255)
	sl.Finalize(100)

	// Verify state through public methods
	if sl.NumSpans() != 2 {
		t.Errorf("Expected 2 spans, got %d", sl.NumSpans())
	}
}
```

## Test Organization

### Directory Structure

```
agg_go/
├── internal/
│   ├── array/
│   │   ├── pod_arrays.go
│   │   ├── pod_arrays_test.go      # Unit tests alongside code
│   │   └── vertex_sequence_test.go
│   ├── rasterizer/
│   │   ├── scanline_aa.go
│   │   └── scanline_aa_test.go
│   └── ...
└── tests/
    ├── integration/                 # Integration tests
    │   ├── rendering_pipeline_test.go
    │   └── color_blending_test.go
    └── visual/                      # Visual regression tests
        ├── primitives/
        │   └── rectangle_test.go
        └── circle_component_test.go
```

### Naming Conventions

- Unit test files: `<file>_test.go` in same package
- Integration tests: In `tests/integration/`
- Visual tests: In `tests/visual/`
- Test functions: `Test<FunctionName>` or `Test<Type>_<Method>`
- Benchmark functions: `Benchmark<Operation>`

## Known Test Issues (To Be Fixed)

### Failing Tests to Address

1. **agg2d**: 1 failing test (Sign function edge case)
2. **color**: 7 failing tests (conversion, lerp, gradient issues)
3. **conv**: Multiple failures (B-spline, GPC polygon operations)
4. **fonts**: 5 failures (character count mismatches)
5. **pixfmt**: Build failures (RGBA16 not implemented)
6. **pixfmt/blender**: 3 failures (blending operations)
7. **pixfmt/gamma**: Panic (nil pointer in RGB gamma application)
8. **platform**: 2 failures (stub implementations)
9. **rasterizer**: Tests pass but take 76 seconds (performance issue?)

### Debug Test Files to Clean Up (Phase 4.3)

These files contain `t.Log()` instead of assertions and should be deleted or refactored:

- `tests/integration/debug_test.go`
- `tests/integration/debug2_test.go`
- `tests/integration/debug3_test.go`
- `tests/integration/minimal_debug_test.go`
- `tests/integration/alternative_debug_test.go`

## Visual Regression Testing (Phase 4.2)

**Status**: Not yet implemented

**Plan**:
1. Generate reference images from C++ AGG for:
   - Basic shapes (lines, rectangles, circles)
   - Anti-aliased edges
   - Gradients
   - Text rendering
2. Create image comparison utility
3. Store golden images in `tests/golden/`
4. Set acceptable pixel difference threshold

## Next Steps for Phase 4

### Phase 4.1: Contract Tests ✅ (In Progress)

- [x] Audit existing test coverage
- [x] Document test patterns
- [x] Fix build failures in test files
- [ ] Add contract tests for packages with <60% coverage
- [ ] Ensure all tests verify real behavior, not mocks

### Phase 4.2: Visual Regression Tests

- [ ] Generate C++ AGG reference images
- [ ] Implement image comparison tool
- [ ] Create golden image test suite
- [ ] Set pixel difference thresholds

### Phase 4.3: Clean Up Existing Tests

- [ ] Delete debug test files
- [ ] Remove tests accessing private fields unnecessarily
- [ ] Convert useful debug tests to proper contract tests
- [ ] Fix all failing tests

### Phase 4.4: Property Tests (Optional)

- [ ] Add property tests for transforms
- [ ] Add property tests for color math
- [ ] Use `testing/quick` or similar

## Principles

1. **Test Behavior, Not Implementation**: Focus on what the code does, not how it does it
2. **No Test-Only Production Code**: Keep production code clean; put test utilities in test files
3. **Understand Before Mocking**: Know why you're mocking and what side effects you need
4. **Complete Mocks**: Mock the full data structure as it exists in reality
5. **TDD When Possible**: Write tests first to ensure they test real behavior
6. **Contract Over Implementation**: Test public API contracts, not internal state
7. **Meaningful Assertions**: Every test should have clear pass/fail criteria

## References

- AGG C++ source: `../agg-2.6/agg-src/`
- Testing anti-patterns skill: `superpowers:testing-anti-patterns`
- Test-driven development skill: `superpowers:test-driven-development`
