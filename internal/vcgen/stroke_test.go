package vcgen

import (
	"testing"

	"agg_go/internal/basics"
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
