# Lion Parser C++ Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `internal/demo/lion/lion.go` a faithful 1:1 translation of `examples/parse_lion.cpp`, reducing unnecessary structural differences while keeping Go idioms.

**Architecture:** The key change is to use a single shared `PathStorageStl` (like C++) instead of one-per-color, and to delegate orientation normalization to `PathBase` methods (porting the missing C++ methods) rather than reimplementing it inline. The `ParseLion` function signature mirrors C++: it populates a path, colors slice, and path_idx slice.

**Tech Stack:** Go, `internal/path.PathBase`, `internal/color`, `internal/basics`

---

## Architectural Decisions

### D1: Single shared PathStorage vs per-color PathStorage

**C++:** One `path_storage` holds all lion sub-paths. Colors and path indices are separate parallel arrays. Callers use `path_idx[i]` to rewind to specific sub-paths.

**Current Go:** Each lion `Path` struct has its own `*PathStorageStl`. This diverges significantly from C++ and means callers iterate differently.

**Decision:** Switch to C++ model — one shared `PathStorageStl`, parallel `Colors` and `PathIdx` slices. This is the biggest change and affects all 22+ callers, but:

- Enables `render_all_paths` (task 10.3) which expects this exact structure
- Enables `bounding_rect` with path indices (used by C++ callers)
- Makes `conv_transform` wrapping natural (one path source, not N)
- The caller update is mechanical: replace `for _, lp := range paths` with index-based iteration using `PathIdx`

### D2: Port `ArrangeOrientationsAllPaths` to `PathBase`

**C++:** `path_base<VC>::arrange_orientations_all_paths()` is a method on path_base that delegates to `arrange_polygon_orientation` → `perceive_polygon_orientation` + `invert_polygon(start, end)`.

**Current Go:** Lion has a standalone `arrangeOrientationsCW()` that manually reimplements the shoelace formula and vertex reversal. This duplicates logic that belongs on `PathBase`.

**Decision:** Port the three missing C++ methods to `PathBase`:

- `PerceivePolygonOrientation(start, end uint) PathFlag`
- `ArrangePolygonOrientation(start uint, orientation PathFlag) uint`
- `ArrangeOrientationsAllPaths(orientation PathFlag)`

The two-argument `InvertPolygon(start, end)` also needs porting (the single-argument version may already exist or not — check). Then lion.go simply calls `path.ArrangeOrientationsAllPaths(PathFlagsCW)`.

### D3: Keep `ParseLion` return type Go-idiomatic

**C++:** Returns `unsigned npaths`, modifies output parameters by pointer.

**Go:** Return a result struct instead. This is a justified Go idiom difference:

```go
type LionData struct {
    Path    *path.PathStorageStl
    Colors  []color.RGBA8[color.Linear]
    PathIdx []uint
    NPaths  int
}
```

### D4: Parsing logic — match C++ flow exactly

The C++ parser calls `path.close_polygon()` before each new color AND before each `move_to`. The Go code only calls close_polygon before move_to. Match C++ exactly.

C++ uses `path.start_new_path()` to record path indices. Go should do the same.

---

## Tasks

### Task 1: Port missing `PathBase` orientation methods

**Files:**

- Modify: `internal/path/path_base.go` (add methods at end)
- Create: `internal/path/path_base_orientation_test.go`

**Step 1: Write tests for the three new methods**

Test `PerceivePolygonOrientation` with a known CW and CCW triangle.
Test `ArrangePolygonOrientation` flips CCW to CW.
Test `ArrangeOrientationsAllPaths` handles multiple sub-paths.

```go
func TestPerceivePolygonOrientation(t *testing.T) {
    // CW triangle (area < 0 in y-up): (0,0) -> (1,0) -> (0,1)
    p := NewPathStorageStl()
    p.MoveTo(0, 0)
    p.LineTo(1, 0)
    p.LineTo(0, 1)
    // perceive on vertices [0..3)
    got := p.PerceivePolygonOrientation(0, 3)
    if got != basics.PathFlagsCW {
        t.Errorf("expected CW, got %v", got)
    }
}

func TestPerceivePolygonOrientationCCW(t *testing.T) {
    // CCW triangle: (0,0) -> (0,1) -> (1,0)
    p := NewPathStorageStl()
    p.MoveTo(0, 0)
    p.LineTo(0, 1)
    p.LineTo(1, 0)
    got := p.PerceivePolygonOrientation(0, 3)
    if got != basics.PathFlagsCCW {
        t.Errorf("expected CCW, got %v", got)
    }
}

func TestArrangeOrientationsAllPaths(t *testing.T) {
    p := NewPathStorageStl()
    // Sub-path 1: CCW triangle
    p.MoveTo(0, 0)
    p.LineTo(0, 1)
    p.LineTo(1, 0)
    p.ClosePolygon(basics.PathFlagsNone)
    // Sub-path 2: already CW
    p.MoveTo(10, 10)
    p.LineTo(11, 10)
    p.LineTo(10, 11)
    p.ClosePolygon(basics.PathFlagsNone)

    p.ArrangeOrientationsAllPaths(basics.PathFlagsCW)

    // Both should now be CW
    // Verify by checking perceive on each sub-path
}
```

**Step 2: Run tests — expect FAIL (methods don't exist)**

```bash
go test ./internal/path/ -run TestPerceivePolygon -v
go test ./internal/path/ -run TestArrangeOrientations -v
```

**Step 3: Implement the methods on `PathBase`**

Port directly from C++ `agg_path_storage.h` lines 1251-1385:

```go
// PerceivePolygonOrientation returns PathFlagsCW or PathFlagsCCW based on
// the signed area of vertices in range [start, end).
// C++ ref: agg_path_storage.h path_base::perceive_polygon_orientation
func (pb *PathBase[VC]) PerceivePolygonOrientation(start, end uint) basics.PathFlag {
    np := end - start
    area := 0.0
    for i := uint(0); i < np; i++ {
        x1, y1, _ := pb.Vertices().Vertex(start + i)
        x2, y2, _ := pb.Vertices().Vertex(start + (i+1)%np)
        area += x1*y2 - y1*x2
    }
    if area < 0.0 {
        return basics.PathFlagsCW
    }
    return basics.PathFlagsCCW
}

// InvertPolygonRange inverts vertex order in [start, end).
// C++ ref: agg_path_storage.h path_base::invert_polygon(start, end)
func (pb *PathBase[VC]) InvertPolygonRange(start, end uint) {
    tmpCmd := pb.Vertices().Command(start)
    end-- // make end inclusive
    for i := start; i < end; i++ {
        pb.Vertices().ModifyCommand(i, pb.Vertices().Command(i+1))
    }
    pb.Vertices().ModifyCommand(end, tmpCmd)
    for end > start {
        pb.Vertices().SwapVertices(start, end)
        start++
        end--
    }
}

// ArrangePolygonOrientation arranges one polygon starting at 'start' to the
// given orientation. Returns the index past the end of this polygon.
// C++ ref: agg_path_storage.h path_base::arrange_polygon_orientation
func (pb *PathBase[VC]) ArrangePolygonOrientation(start uint, orientation basics.PathFlag) uint {
    if orientation == basics.PathFlagsNone {
        return start
    }
    total := pb.TotalVertices()
    // Skip non-vertices
    for start < total && !basics.IsVertex(basics.PathCommand(pb.Command(start))) {
        start++
    }
    // Skip consecutive move_to
    for start+1 < total &&
        basics.IsMoveTo(basics.PathCommand(pb.Command(start))) &&
        basics.IsMoveTo(basics.PathCommand(pb.Command(start+1))) {
        start++
    }
    // Find end of polygon
    end := start + 1
    for end < total && !basics.IsNextPoly(basics.PathCommand(pb.Command(end))) {
        end++
    }
    if end-start > 2 {
        if basics.PathFlag(pb.PerceivePolygonOrientation(start, end)) != orientation {
            pb.InvertPolygonRange(start, end)
            for end < total && basics.IsEndPoly(basics.PathCommand(pb.Command(end))) {
                cmd := pb.Command(end)
                pb.ModifyCommand(end, basics.SetOrientation(cmd, uint32(orientation)))
                end++
            }
        }
    }
    return end
}

// ArrangeOrientationsAllPaths arranges all polygons in all paths to the
// given orientation.
// C++ ref: agg_path_storage.h path_base::arrange_orientations_all_paths
func (pb *PathBase[VC]) ArrangeOrientationsAllPaths(orientation basics.PathFlag) {
    if orientation == basics.PathFlagsNone {
        return
    }
    start := uint(0)
    for start < pb.TotalVertices() {
        start = pb.ArrangePolygonOrientation(start, orientation)
    }
}
```

Note: Check if `IsNextPoly` and `SetOrientation` exist in `basics`. Port them if not.

**Step 4: Run tests — expect PASS**

```bash
go test ./internal/path/ -run TestPerceivePolygon -v
go test ./internal/path/ -run TestArrangeOrientations -v
```

**Step 5: Commit**

```bash
git add internal/path/path_base.go internal/path/path_base_orientation_test.go
git commit -m "path: port ArrangeOrientationsAllPaths from C++ path_base"
```

---

### Task 2: Restructure lion.go to match C++ parse_lion

**Files:**

- Modify: `internal/demo/lion/lion.go`

**Step 1: Replace the `Path` struct and `Parse` function**

New API:

```go
// LionData holds the parsed lion artwork, structured like C++ parse_lion output:
// a single shared path storage with parallel color and path-index arrays.
type LionData struct {
    Path    *path.PathStorageStl
    Colors  []color.RGBA8[color.Linear]
    PathIdx []uint
    NPaths  int
}

// Parse parses the embedded lion data, returning a single shared path storage
// with per-sub-path colors and indices. Mirrors C++ parse_lion() exactly.
func Parse() LionData {
    // ... match C++ flow
}
```

The parsing loop must match C++ exactly:

1. On color line: `path.ClosePolygon()`, store color, `path.StartNewPath()` → pathIdx
2. On M command: `path.ClosePolygon()`, `path.MoveTo(x, y)`
3. On L command: `path.LineTo(x, y)`
4. After parsing: `path.ArrangeOrientationsAllPaths(basics.PathFlagsCW)`

Remove the old `arrangeOrientationsCW` function entirely.

**Step 2: Build — expect compile errors in callers (that's expected)**

```bash
go build ./internal/demo/lion/
```

**Step 3: Commit**

```bash
git add internal/demo/lion/lion.go
git commit -m "lion: restructure to single shared PathStorage matching C++ parse_lion"
```

---

### Task 3: Update all callers to use new LionData API

**Files:** All 22+ files that import `internal/demo/lion`

This is a mechanical transformation. The old pattern:

```go
lionPaths := liondemo.Parse()
for _, lp := range lionPaths {
    lp.Path.Rewind(0)
    for {
        x, y, cmd := lp.Path.NextVertex()
        // ... use lp.Color
    }
}
```

Becomes:

```go
ld := liondemo.Parse()
for i := 0; i < ld.NPaths; i++ {
    ld.Path.Rewind(ld.PathIdx[i])
    for {
        x, y, cmd := ld.Path.NextVertex()
        // ... use ld.Colors[i]
    }
}
```

For callers that use the high-level agg2d API and iterate manually, the same pattern applies but with `ld.Colors[i].R` etc instead of `lp.Color.R`.

Bounding box computation changes from iterating per-path to iterating the single shared path once.

Update files in groups:

1. Examples (core/basic, core/intermediate, core/advanced)
2. WASM demos (cmd/wasm/demo\_\*.go)
3. Tests (tests/benchmark, tests/integration, tests/visual)
4. Commands (cmd/aggtest, cmd/lion_bounds)

**Step 1: Update all callers**

**Step 2: Build — expect PASS**

```bash
go build ./...
```

**Step 3: Run tests**

```bash
go test ./tests/integration/ -run TestCPPParity -v
```

**Step 4: Commit**

```bash
git add -A
git commit -m "lion: update all callers to use shared PathStorage API"
```

---

### Task 4: Verify parity and clean up

**Step 1: Run full test suite**

```bash
just test
```

**Step 2: Run pixel-level parity tests**

```bash
go test ./tests/integration/ -run TestCPPParity -v -count=1
```

Expected: all pass, pixel(300,100) still = (245, 217, 177).

**Step 3: Commit any fixes if needed**

---

## Pre-flight Checklist

Before starting, verify these helpers exist in `internal/basics/`:

- [ ] `IsNextPoly(cmd) bool` — C++: `is_next_poly`
- [ ] `SetOrientation(cmd uint32, orientation uint32) uint32` — C++: `set_orientation`
- [ ] `PathFlagsCW` and `PathFlagsCCW` constants
- [ ] `IsEndPoly(cmd) bool`

If any are missing, port them first as part of Task 1.
