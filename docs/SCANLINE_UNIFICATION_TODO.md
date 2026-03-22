# Scanline Interface Unification — Remaining Work

**Date**: 2026-03-21
**Status**: In progress — core infrastructure done, adapter removal in examples/wasm pending

## What Was Done

### 1. Unified Scanline Interface (`internal/scanline/interfaces.go`)

Created a single `Scanline` interface combining writer (rasterizer) and reader (renderer) methods:

- Writer: `ResetSpans()`, `AddCell(x int, cover uint)`, `AddSpan(x, length int, cover uint)`, `Finalize(y int)`
- Reader: `Y() int`, `NumSpans() int`, `BeginIterator() ScanlineIterator`
- Setup: `Reset(minX, maxX int)`

Added `BeginIterator()` method and `sliceIter*` adapter types to all concrete scanline types:

- `ScanlineP8`, `ScanlineU8`, `ScanlineBin`
- `Scanline32P8`, `Scanline32U8`, `Scanline32Bin`

### 2. Broke Import Cycle

- `internal/scanline/boolean_algebra.go`: removed import of `renderer/scanline`, now uses local `ScanlineIterator` (from `storage_aa.go`)
- `internal/scanline/boolean_algebra_test.go`: replaced `scanline.SpanData` → `SpanInfo`, `scanline.ScanlineIterator` → `ScanlineIterator`

### 3. Updated Rasterizer (`internal/rasterizer/`)

- Changed `ScanlineInterface.AddCell/AddSpan` from `uint32` to `uint` (matching concrete types)
- Changed `RasterizerScanlineAA.SweepScanline` to accept `scanline.Scanline` (the unified interface)
- Changed `RasterizerScanlineAANoGamma.SweepScanline` to accept `scanline.Scanline`
- Updated `scanline_aa.go` and `scanline_aa_nogamma.go`: `uint32(alpha)` → `uint(alpha)`
- Updated test mock in `scanline_aa_test.go` to satisfy unified interface

### 4. Updated Renderer (`internal/renderer/scanline/`)

- `interfaces.go`: `ScanlineInterface`, `ScanlineIterator`, `SpanData` are now type aliases for `scanline.Scanline`, `scanline.ScanlineIterator`, `scanline.SpanInfo`
- `RasterizerInterface.SweepScanline` accepts `ScanlineInterface` (= `scanline.Scanline`)
- Removed `ResettableScanline` (no longer needed — `Reset()` is in the unified interface)
- `render_functions.go`: all `sl.Begin()` → `sl.BeginIterator()`, removed `ResettableScanline` type assertions
- `helpers.go`: same changes
- `test_mocks.go`: updated mocks to satisfy unified interface (added `AddCell`, `AddSpan`, etc.)

### 5. Updated `internal/renderer/enlarged.go`

- `sl.Begin()` → `sl.BeginIterator()`

### 6. Removed Adapter Boilerplate from `internal/agg2d/`

- `adapters.go`: removed `scanlineWrapper`, `spanIter`, `rasterizerAdapter`, `rasScanlineAdapter`
- `agg2d.go`, `rendering.go`, `image.go`, `text.go`: use concrete types directly

### 7. Already Fixed by User

The user has already manually fixed these files (visible in system reminders):

- `cmd/aggtest/main.go` — adapter removed, uses `newRas()` returning `*rasType` directly
- `examples/core/intermediate/alpha_mask2/main.go` — adapter removed
- `examples/core/intermediate/alpha_mask/main.go` — adapter removed
- `examples/core/basic/multi_clip/main.go` — partially fixed (has `sl := sl` duplicate line that needs cleanup)

## What Remains — 19 Failing Packages

All failures follow the same pattern: old adapter types (`scanlineWrapper`, `rasScanlineAdapter/Adaptor`, `rasterizerAdapter`, `scanlineWrapperP8`, `scanlineWrapperU8`, etc.) don't satisfy the new unified `scanline.Scanline` interface.

### Fix Pattern (mechanical, same for every file)

1. **Delete adapter types**: Remove `scanlineWrapper`, `spanIter`, `rasterizerAdaptor`/`rasterizerAdapter`, `rasScanlineAdaptor`/`rasScanlineAdapter`, `scanlineWrapperP8`, `scanlineWrapperU8`, and all their methods.

2. **Replace scanline creation**:

   ```go
   // OLD:
   sl := &scanlineWrapper{sl: scanline.NewScanlineP8()}
   // NEW:
   sl := scanline.NewScanlineP8()
   ```

3. **Replace rasterizer creation** (where an adapter wraps it):

   ```go
   // OLD:
   ras := &rasterizerAdapter{ras: rasterizer.NewRasterizerScanlineAA[...](conv, clip)}
   // NEW:
   ras := rasterizer.NewRasterizerScanlineAA[...](conv, clip)
   ```

   Or use a type alias + helper:

   ```go
   type rasType = rasterizer.RasterizerScanlineAA[int, rasterizer.RasConvInt, *rasterizer.RasterizerSlNoClip]
   func newRas() *rasType { return rasterizer.NewRasterizerScanlineAA[...](...) }
   ```

4. **Fix method calls** that went through the adapter wrapper:

   ```go
   // OLD: ras.ras.Reset() / ras.ras.AddPath(...) / ras.ras.AddVertex(...)
   // NEW: ras.Reset()    / ras.AddPath(...)      / ras.AddVertex(...)
   ```

5. **Pass scanline/rasterizer directly** to render functions:

   ```go
   renscan.RenderScanlinesAASolid(ras, sl, rb, color)  // both satisfy interfaces directly
   ```

6. **Remove unused imports** (`rasterizer` package if only used by adapter).

7. **Keep `ellipseVS`** adapter — this wraps `shapes.Ellipse` to `rasterizer.VertexSource` (different concern, still needed).

### Files to Fix

#### `cmd/wasm/` (3 files)

- `cmd/wasm/adapter.go` — main shared adapter file; delete `rasScanlineAdapter`, `scanlineWrapperP8`/`scanlineWrapperU8`, `rasterizerAdapter` types and the `RasterizerInterface` / `SweepScanline` boilerplate
- `cmd/wasm/demo_alpha_mask.go` — uses adapters from adapter.go
- `cmd/wasm/demo_alpha_mask2.go` — uses adapters from adapter.go

#### `examples/core/basic/` (3 files)

- `examples/core/basic/circles/main.go`
- `examples/core/basic/multi_clip/main.go` — partially fixed, has duplicate `sl := sl` line to clean up
- `examples/core/basic/rounded_rect/main.go`

#### `examples/core/intermediate/` (13 files)

- `examples/core/intermediate/bspline/main.go`
- `examples/core/intermediate/compositing2/main.go`
- `examples/core/intermediate/conv_dash_marker/main.go`
- `examples/core/intermediate/flash_rasterizer2/main.go`
- `examples/core/intermediate/image1/main.go`
- `examples/core/intermediate/image_alpha/main.go`
- `examples/core/intermediate/image_transforms/main.go`
- `examples/core/intermediate/line_thickness/main.go`
- `examples/core/intermediate/pattern_fill/main.go`
- `examples/core/intermediate/polymorphic_renderer/main.go`
- `examples/core/intermediate/rasterizers/main.go`
- `examples/core/intermediate/rasterizers2/main.go`
- `examples/core/intermediate/truetype_test/main.go`

#### `examples/core/advanced/` (2 files)

- `examples/core/advanced/distortions/main.go`
- `examples/core/advanced/gamma_correction/main.go`

### Special Cases

#### `cmd/wasm/adapter.go`

This is a shared adapter file used by multiple wasm demos. It defines adapter types used across `demo_*.go` files. After removing the adapters, each demo that needs a rasterizer should create one directly (or use a shared helper function in the file).

The file also defines a `RasterizerInterface` local interface — this should be replaced with `renscan.RasterizerInterface`.

#### `internal/demo/blendcolor/draw.go` and `internal/demo/quadwarp/draw.go`

These have `scanlineAdapter` types with `AddCell(int, uint32)` — change to `uint` or delete the adapter entirely and use concrete scanline types.

#### `internal/demo/scanlineboolean2/draw.go`

Has `rasterScanlineBinAdapter` — this adapts for `ScanlineBin` types. Same pattern: delete adapter, use concrete types directly.

#### `examples/core/intermediate/polymorphic_renderer/main.go`

May use a custom renderer that wraps scanlines — check carefully before removing adapters.

#### `examples/core/intermediate/rasterizers/main.go` and `rasterizers2/main.go`

May use `RasterizerScanlineAANoGamma` in addition to `RasterizerScanlineAA` — both now accept `scanline.Scanline`, so the fix is the same.

#### `examples/core/intermediate/flash_rasterizer2/main.go`

Uses compound rasterizer (`RasterizerCompoundAA`). The compound rasterizer has its own `CompoundScanlineInterface` with `basics.Int8u` covers — this is SEPARATE from the unified interface and should not be changed. Only the non-compound adapter boilerplate needs removal.

## After Fixing All Files

1. Run `go build ./...` — should produce zero errors
2. Run `go test ./...` — check for test regressions
3. Run `go run ./cmd/aggtest/` — verify pixel-level output still matches C++
4. Run `just check` — full validation

## Related: PLAN.md 10.7 Status

Once all files compile and tests pass, update PLAN.md section 10.7:

- [x] Unified scanline interface defined in `internal/scanline/interfaces.go`
- [x] Rasterizer and renderer use the same interface
- [ ] All adapter boilerplate removed from examples and demos (19 packages remaining)
- [ ] All example and demo files compile without manual adapter types
