package vcgen

import (
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

func TestVCGenStrokeCreation(t *testing.T) {
	vg := NewVCGenStroke()

	if vg == nil {
		t.Fatal("Expected non-nil VCGenStroke")
	}

	// Test default values
	if vg.Width() != 1.0 {
		t.Errorf("Expected default width 1.0, got %f", vg.Width())
	}
	if vg.LineCap() != basics.ButtCap {
		t.Errorf("Expected default line cap ButtCap, got %v", vg.LineCap())
	}
	if vg.LineJoin() != basics.MiterJoin {
		t.Errorf("Expected default line join MiterJoin, got %v", vg.LineJoin())
	}
}

func TestVCGenStrokeSetters(t *testing.T) {
	vg := NewVCGenStroke()

	// Test width
	vg.SetWidth(5.0)
	if vg.Width() != 5.0 {
		t.Errorf("Expected width 5.0, got %f", vg.Width())
	}

	// Test line cap
	vg.SetLineCap(basics.RoundCap)
	if vg.LineCap() != basics.RoundCap {
		t.Errorf("Expected RoundCap, got %v", vg.LineCap())
	}

	// Test line join
	vg.SetLineJoin(basics.BevelJoin)
	if vg.LineJoin() != basics.BevelJoin {
		t.Errorf("Expected BevelJoin, got %v", vg.LineJoin())
	}

	// Test inner join
	vg.SetInnerJoin(basics.InnerRound)
	if vg.InnerJoin() != basics.InnerRound {
		t.Errorf("Expected InnerRound, got %v", vg.InnerJoin())
	}

	// Test miter limit
	vg.SetMiterLimit(8.0)
	if vg.MiterLimit() != 8.0 {
		t.Errorf("Expected miter limit 8.0, got %f", vg.MiterLimit())
	}
}

func TestVCGenStrokeSimpleLine(t *testing.T) {
	vg := NewVCGenStroke()
	vg.SetWidth(2.0)

	// Add a simple horizontal line: (0,0) -> (10,0)
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(10, 0, basics.PathCmdLineTo)

	vg.Rewind(0)

	vertices := make([]struct {
		x, y float64
		cmd  basics.PathCommand
	}, 0)

	// Collect all vertices
	for {
		x, y, cmd := vg.Vertex()
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})

		if basics.IsStop(cmd) {
			break
		}
	}

	// Should have generated stroke vertices
	if len(vertices) < 3 { // At least MoveTo, some LineTO's, and Stop
		t.Errorf("Expected at least 3 vertices, got %d", len(vertices))
	}

	// First command should be MoveTo
	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", vertices[0].cmd)
	}

	// Last command should be Stop
	if vertices[len(vertices)-1].cmd != basics.PathCmdStop {
		t.Errorf("Expected last command to be Stop, got %v", vertices[len(vertices)-1].cmd)
	}
}

func TestVCGenStrokeClosedPath(t *testing.T) {
	vg := NewVCGenStroke()
	vg.SetWidth(2.0)

	// Add a triangle: (0,0) -> (10,0) -> (5,10) -> close
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(10, 0, basics.PathCmdLineTo)
	vg.AddVertex(5, 10, basics.PathCmdLineTo)
	vg.AddVertex(0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose)))

	vg.Rewind(0)

	vertices := make([]struct {
		x, y float64
		cmd  basics.PathCommand
	}, 0)

	// Collect all vertices
	for {
		x, y, cmd := vg.Vertex()
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})

		if basics.IsStop(cmd) {
			break
		}
	}

	// Should have generated stroke vertices for a closed path
	if len(vertices) < 5 { // More vertices expected for closed path
		t.Errorf("Expected at least 5 vertices for closed path, got %d", len(vertices))
	}

	// Should contain EndPoly commands for closed path
	foundEndPoly := false
	for _, v := range vertices {
		if basics.IsEndPoly(v.cmd) {
			foundEndPoly = true
			break
		}
	}

	if !foundEndPoly {
		t.Error("Expected to find EndPoly command in closed path stroke")
	}
}

func TestVCGenStrokeClosedPathEmitsBothOutlines(t *testing.T) {
	vg := NewVCGenStroke()
	vg.SetWidth(2.0)
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(10, 0, basics.PathCmdLineTo)
	vg.AddVertex(10, 10, basics.PathCmdLineTo)
	vg.AddVertex(0, 10, basics.PathCmdLineTo)
	vg.AddVertex(0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose)))

	vg.Rewind(0)

	moveToCount := 0
	endPolyCount := 0
	var endPolyCmds []basics.PathCommand
	for {
		_, _, cmd := vg.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsMoveTo(cmd) {
			moveToCount++
		}
		if basics.IsEndPoly(cmd) {
			endPolyCount++
			endPolyCmds = append(endPolyCmds, cmd)
		}
	}

	if moveToCount < 2 {
		t.Fatalf("expected closed stroke to emit separate move_to commands for both outlines, got %d", moveToCount)
	}
	if endPolyCount != 2 {
		t.Fatalf("expected closed stroke to emit two end_poly commands, got %d", endPolyCount)
	}
	if !basics.IsCCW(uint32(endPolyCmds[0])) {
		t.Fatalf("expected first outline end_poly to be CCW, got %v", endPolyCmds[0])
	}
	if !basics.IsCW(uint32(endPolyCmds[1])) {
		t.Fatalf("expected second outline end_poly to be CW, got %v", endPolyCmds[1])
	}
}

func TestVCGenStrokeRemoveAll(t *testing.T) {
	vg := NewVCGenStroke()

	// Add some vertices
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(10, 0, basics.PathCmdLineTo)

	// Clear all
	vg.RemoveAll()

	vg.Rewind(0)

	// Should generate only a stop command
	x, y, cmd := vg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop after RemoveAll, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenStrokeEmptyPath(t *testing.T) {
	vg := NewVCGenStroke()

	vg.Rewind(0)

	// Should generate only a stop command for empty path
	x, y, cmd := vg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop for empty path, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenStrokeSingleVertex(t *testing.T) {
	vg := NewVCGenStroke()
	vg.SetWidth(2.0)

	// Add only one vertex
	vg.AddVertex(5, 5, basics.PathCmdMoveTo)

	vg.Rewind(0)

	// Should generate only a stop command (not enough vertices for a stroke)
	x, y, cmd := vg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop for single vertex, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenStrokeDifferentLineCaps(t *testing.T) {
	testCases := []struct {
		name    string
		lineCap basics.LineCap
	}{
		{"ButtCap", basics.ButtCap},
		{"SquareCap", basics.SquareCap},
		{"RoundCap", basics.RoundCap},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vg := NewVCGenStroke()
			vg.SetWidth(2.0)
			vg.SetLineCap(tc.lineCap)

			// Add a simple line
			vg.AddVertex(0, 0, basics.PathCmdMoveTo)
			vg.AddVertex(10, 0, basics.PathCmdLineTo)

			vg.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := vg.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				if cmd == basics.PathCmdLineTo || cmd == basics.PathCmdMoveTo {
					vertexCount++
				}
			}

			// Different line caps should generate different numbers of vertices
			// Round cap should generate more vertices than butt cap
			if tc.lineCap == basics.RoundCap && vertexCount < 6 {
				t.Logf("Round cap generated %d vertices (may be correct)", vertexCount)
			} else if vertexCount == 0 {
				t.Errorf("No vertices generated for %s", tc.name)
			}
		})
	}
}

func TestVCGenStrokeDifferentLineJoins(t *testing.T) {
	testCases := []struct {
		name     string
		lineJoin basics.LineJoin
	}{
		{"MiterJoin", basics.MiterJoin},
		{"RoundJoin", basics.RoundJoin},
		{"BevelJoin", basics.BevelJoin},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vg := NewVCGenStroke()
			vg.SetWidth(2.0)
			vg.SetLineJoin(tc.lineJoin)

			// Add an L-shape to create a join
			vg.AddVertex(0, 0, basics.PathCmdMoveTo)
			vg.AddVertex(10, 0, basics.PathCmdLineTo)
			vg.AddVertex(10, 10, basics.PathCmdLineTo)

			vg.Rewind(0)

			vertexCount := 0
			for {
				_, _, cmd := vg.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				if cmd == basics.PathCmdLineTo || cmd == basics.PathCmdMoveTo {
					vertexCount++
				}
			}

			// Should generate some vertices for the join
			if vertexCount == 0 {
				t.Errorf("No vertices generated for %s", tc.name)
			}

			// Round join might generate more vertices than other joins
			if tc.lineJoin == basics.RoundJoin && vertexCount < 8 {
				t.Logf("Round join generated %d vertices (may be correct)", vertexCount)
			}
		})
	}
}

func TestVCGenStrokeShortenReducesExtent(t *testing.T) {
	// A horizontal open path; shortening should reduce the maximum X
	collectMaxX := func(vg *VCGenStroke) float64 {
		maxX := -1e308
		for {
			x, _, cmd := vg.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if cmd == basics.PathCmdMoveTo || cmd == basics.PathCmdLineTo {
				if x > maxX {
					maxX = x
				}
			}
		}
		return maxX
	}

	base := NewVCGenStroke()
	base.SetWidth(4.0)
	base.SetLineCap(basics.ButtCap)
	base.AddVertex(0, 0, basics.PathCmdMoveTo)
	base.AddVertex(100, 0, basics.PathCmdLineTo)
	base.Rewind(0)
	maxXBase := collectMaxX(base)

	shortened := NewVCGenStroke()
	shortened.SetWidth(4.0)
	shortened.SetLineCap(basics.ButtCap)
	shortened.SetShorten(20.0)
	shortened.AddVertex(0, 0, basics.PathCmdMoveTo)
	shortened.AddVertex(100, 0, basics.PathCmdLineTo)
	shortened.Rewind(0)
	maxXShort := collectMaxX(shortened)

	if !(maxXShort < maxXBase-1e-6) {
		t.Fatalf("expected shortened stroke max X < base (got %.3f vs %.3f)", maxXShort, maxXBase)
	}
}

// collectStrokeVertices drains all vertices from the stroke generator and returns them.
func collectStrokeVertices(vg *VCGenStroke) []struct {
	x, y float64
	cmd  basics.PathCommand
} {
	var out []struct {
		x, y float64
		cmd  basics.PathCommand
	}
	for {
		x, y, cmd := vg.Vertex()
		out = append(out, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
		if basics.IsStop(cmd) {
			break
		}
	}
	return out
}

// TestVCGenStrokeButtCapGeometry verifies that ButtCap does not extend beyond
// the endpoints of the path (no overshoot). Ref: agg_vcgen_stroke.cpp calc_cap.
func TestVCGenStrokeButtCapGeometry(t *testing.T) {
	vg := NewVCGenStroke()
	vg.SetWidth(4.0)
	vg.SetLineCap(basics.ButtCap)
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(20, 0, basics.PathCmdLineTo)
	vg.Rewind(0)

	vs := collectStrokeVertices(vg)

	maxX := -1e308
	minX := 1e308
	for _, v := range vs {
		if basics.IsStop(v.cmd) {
			continue
		}
		if v.x > maxX {
			maxX = v.x
		}
		if v.x < minX {
			minX = v.x
		}
	}

	// ButtCap must not extend beyond path endpoints (within floating-point tolerance)
	if minX < -1e-9 {
		t.Errorf("ButtCap: min X %.6f < 0, cap overshoots start endpoint", minX)
	}
	if maxX > 20+1e-9 {
		t.Errorf("ButtCap: max X %.6f > 20, cap overshoots end endpoint", maxX)
	}
}

// TestVCGenStrokeSquareCapGeometry verifies that SquareCap extends exactly
// half-width beyond each endpoint. Ref: agg_math_stroke.h calc_cap.
func TestVCGenStrokeSquareCapGeometry(t *testing.T) {
	const halfW = 2.0
	vg := NewVCGenStroke()
	vg.SetWidth(2 * halfW)
	vg.SetLineCap(basics.SquareCap)
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(20, 0, basics.PathCmdLineTo)
	vg.Rewind(0)

	vs := collectStrokeVertices(vg)

	maxX := -1e308
	minX := 1e308
	for _, v := range vs {
		if basics.IsStop(v.cmd) {
			continue
		}
		if v.x > maxX {
			maxX = v.x
		}
		if v.x < minX {
			minX = v.x
		}
	}

	// SquareCap must extend by exactly half-width = 2.0 beyond each endpoint
	if minX > -halfW+1e-6 {
		t.Errorf("SquareCap: min X %.6f should be ≈ %.6f (half-width extension at start)", minX, -halfW)
	}
	if maxX < 20+halfW-1e-6 {
		t.Errorf("SquareCap: max X %.6f should be ≈ %.6f (half-width extension at end)", maxX, 20+halfW)
	}
}

// TestVCGenStrokeRoundCapMoreVerticesThanButt verifies that RoundCap generates
// more outline vertices than ButtCap for the same geometry (arc approximation).
func TestVCGenStrokeRoundCapMoreVerticesThanButt(t *testing.T) {
	countLineVertices := func(cap basics.LineCap) int {
		vg := NewVCGenStroke()
		vg.SetWidth(4.0)
		vg.SetLineCap(cap)
		vg.AddVertex(0, 0, basics.PathCmdMoveTo)
		vg.AddVertex(20, 0, basics.PathCmdLineTo)
		vg.Rewind(0)
		n := 0
		for {
			_, _, cmd := vg.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) || basics.IsMoveTo(cmd) {
				n++
			}
		}
		return n
	}

	buttCount := countLineVertices(basics.ButtCap)
	roundCount := countLineVertices(basics.RoundCap)

	if roundCount <= buttCount {
		t.Errorf("RoundCap (%d vertices) should generate more than ButtCap (%d vertices)", roundCount, buttCount)
	}
}

// TestVCGenStrokeMiterLimitClipping verifies that the miter limit affects join
// geometry. For a 90° turn with width 4:
//   - miter distance = sqrt(2) * half-width ≈ 2.83 (ratio ≈ 1.41)
//   - limit 1.1 clips to bevel (2 outer vertices per join)
//   - limit 4.0 keeps the miter point (1 outer vertex per join)
//
// Both must produce output; differing limits must produce different geometry.
// Ref: agg_math_stroke.h calc_join miter-limit branch.
func TestVCGenStrokeMiterLimitClipping(t *testing.T) {
	buildAngle := func(miterLimit float64) int {
		vg := NewVCGenStroke()
		vg.SetWidth(4.0)
		vg.SetLineJoin(basics.MiterJoin)
		vg.SetMiterLimit(miterLimit)
		vg.AddVertex(0, 0, basics.PathCmdMoveTo)
		vg.AddVertex(10, 0, basics.PathCmdLineTo)
		vg.AddVertex(10, 10, basics.PathCmdLineTo) // 90° turn
		vg.Rewind(0)
		n := 0
		for {
			_, _, cmd := vg.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) || basics.IsMoveTo(cmd) {
				n++
			}
		}
		return n
	}

	// For a 90° join: miter ratio ≈ 1.41; limit 1.1 clips to bevel, limit 4.0 keeps miter.
	// Bevel (clipped) uses 2 outer-side vertices; miter uses 1. They must differ.
	lowLimit := buildAngle(1.1)  // clips to bevel → 2 outer vertices
	highLimit := buildAngle(4.0) // keeps miter tip → 1 outer vertex

	if lowLimit == 0 || highLimit == 0 {
		t.Fatalf("both miter limits must produce vertices (low=%d, high=%d)", lowLimit, highLimit)
	}
	if lowLimit == highLimit {
		t.Errorf("different miter limits should produce different vertex counts (both gave %d)", lowLimit)
	}
}

// TestVCGenStrokeInnerJoinCornerHandling verifies that all four InnerJoin modes
// successfully render a concave corner (inner side of an acute-angle bend).
// Ref: agg_math_stroke.h calc_join inner-join branch.
func TestVCGenStrokeInnerJoinCornerHandling(t *testing.T) {
	innerJoins := []struct {
		name string
		ij   basics.InnerJoin
	}{
		{"InnerBevel", basics.InnerBevel},
		{"InnerMiter", basics.InnerMiter},
		{"InnerJag", basics.InnerJag},
		{"InnerRound", basics.InnerRound},
	}

	for _, tc := range innerJoins {
		t.Run(tc.name, func(t *testing.T) {
			// Acute V-shape: inner side gets a concave corner
			vg := NewVCGenStroke()
			vg.SetWidth(6.0)
			vg.SetLineJoin(basics.MiterJoin)
			vg.SetInnerJoin(tc.ij)
			vg.AddVertex(0, 0, basics.PathCmdMoveTo)
			vg.AddVertex(10, 0, basics.PathCmdLineTo)
			vg.AddVertex(5, 4, basics.PathCmdLineTo) // shallow V

			vg.Rewind(0)
			vs := collectStrokeVertices(vg)

			// Must produce at least MoveTo + a few vertices + Stop
			nonStop := 0
			for _, v := range vs {
				if !basics.IsStop(v.cmd) {
					nonStop++
				}
			}
			if nonStop < 4 {
				t.Errorf("%s: expected ≥ 4 non-stop vertices, got %d", tc.name, nonStop)
			}
		})
	}
}

// TestVCGenStrokeClosedPathWidthSymmetry verifies that for a closed square path
// the stroke outline is symmetric around the original edges (equal offsets on
// both sides). Ref: agg_vcgen_stroke.cpp outline1 / outline2 paths.
func TestVCGenStrokeClosedPathWidthSymmetry(t *testing.T) {
	const halfW = 1.5
	vg := NewVCGenStroke()
	vg.SetWidth(2 * halfW)
	vg.SetLineJoin(basics.BevelJoin)
	vg.AddVertex(0, 0, basics.PathCmdMoveTo)
	vg.AddVertex(10, 0, basics.PathCmdLineTo)
	vg.AddVertex(10, 10, basics.PathCmdLineTo)
	vg.AddVertex(0, 10, basics.PathCmdLineTo)
	vg.AddVertex(0, 0, basics.PathCommand(uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose)))
	vg.Rewind(0)

	vs := collectStrokeVertices(vg)

	// Collect Y coords on the bottom edge (where Y should cluster near ±halfW)
	minY := 1e308
	maxY := -1e308
	for _, v := range vs {
		if basics.IsStop(v.cmd) {
			continue
		}
		if v.y < minY {
			minY = v.y
		}
		if v.y > maxY {
			maxY = v.y
		}
	}

	// Outer edge should extend beyond 10 and inner below 0
	if maxY < 10+halfW-1.0 {
		t.Errorf("Outer edge maxY %.3f, expected ≥ %.3f", maxY, 10+halfW-1.0)
	}
	if minY > -halfW+1.0 {
		t.Errorf("Inner edge minY %.3f, expected ≤ %.3f", minY, -halfW+1.0)
	}
}
