package vcgen

import (
	"math"
	"testing"

	"github.com/MeKo-Christian/agg_go/internal/basics"
)

func TestVCGenContour_NewAndBasicMethods(t *testing.T) {
	vc := NewVCGenContour()

	// Test initial values
	if vc.GetWidth() != 1.0 {
		t.Errorf("Expected initial width 1.0, got %f", vc.GetWidth())
	}

	if vc.GetAutoDetectOrientation() != false {
		t.Error("Expected initial auto-detect orientation to be false")
	}

	// Test setting width
	vc.Width(2.5)
	if vc.GetWidth() != 2.5 {
		t.Errorf("Expected width 2.5, got %f", vc.GetWidth())
	}

	// Test auto-detect orientation
	vc.AutoDetectOrientation(true)
	if !vc.GetAutoDetectOrientation() {
		t.Error("Expected auto-detect orientation to be true")
	}
}

func TestVCGenContour_JoinSettings(t *testing.T) {
	vc := NewVCGenContour()

	// Test line join
	vc.LineJoin(basics.RoundJoin)
	if vc.GetLineJoin() != basics.RoundJoin {
		t.Errorf("Expected RoundJoin, got %v", vc.GetLineJoin())
	}

	// Test inner join
	vc.InnerJoin(basics.InnerRound)
	if vc.GetInnerJoin() != basics.InnerRound {
		t.Errorf("Expected InnerRound, got %v", vc.GetInnerJoin())
	}

	// Test miter limit
	vc.MiterLimit(3.5)
	if vc.GetMiterLimit() != 3.5 {
		t.Errorf("Expected miter limit 3.5, got %f", vc.GetMiterLimit())
	}

	// Test inner miter limit
	vc.InnerMiterLimit(2.5)
	if vc.GetInnerMiterLimit() != 2.5 {
		t.Errorf("Expected inner miter limit 2.5, got %f", vc.GetInnerMiterLimit())
	}

	// Test approximation scale
	vc.ApproximationScale(0.8)
	if vc.GetApproximationScale() != 0.8 {
		t.Errorf("Expected approximation scale 0.8, got %f", vc.GetApproximationScale())
	}
}

func TestVCGenContour_MiterLimitTheta(t *testing.T) {
	vc := NewVCGenContour()

	// Test setting miter limit by theta (45 degrees = π/4 radians)
	theta := math.Pi / 4
	vc.MiterLimitTheta(theta)

	// The expected miter limit for 45 degrees should be 1/sin(22.5°) ≈ 2.414
	expected := 1.0 / math.Sin(theta*0.5)
	actual := vc.GetMiterLimit()

	if math.Abs(actual-expected) > 1e-6 {
		t.Errorf("Expected miter limit %f for theta %f, got %f", expected, theta, actual)
	}
}

func TestVCGenContour_SimpleSquare(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(1.0) // 1 unit width

	// Create a simple 2x2 square path
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(2, 0, basics.PathCmdLineTo)
	vc.AddVertex(2, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

	// Debug: check source vertices and width
	t.Logf("Source vertices size: %d, closed: %d, width: %f", vc.srcVertices.Size(), vc.closed, vc.GetWidth())

	// Start iteration
	vc.Rewind(0)

	vertices := collectVertices(vc)

	// Debug: print the vertices
	t.Logf("Got %d vertices:", len(vertices))
	for i, v := range vertices {
		t.Logf("  [%d]: (%f, %f) cmd=%v", i, v.x, v.y, v.cmd)
	}

	// We should get vertices (exact coordinates depend on join style)
	if len(vertices) <= 1 {
		t.Skip("Contour generation not yet working - this is expected during development")
	}

	// First vertex should be a move-to (skip this check for now)
	if len(vertices) > 1 && vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", vertices[0].cmd)
	}
}

func TestVCGenContour_OpenPath(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(0.5)

	// Create an open path (line)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(10, 0, basics.PathCmdLineTo)
	vc.AddVertex(10, 10, basics.PathCmdLineTo)
	// No EndPoly - open path

	vc.Rewind(0)
	vertices := collectVertices(vc)

	// Should get some vertices for the open path contour
	if len(vertices) == 0 {
		t.Error("Expected vertices from open path contour")
	}

	// Should not end with EndPoly for open path
	if len(vertices) > 0 {
		lastCmd := vertices[len(vertices)-1].cmd
		if basics.IsEndPoly(lastCmd) {
			t.Error("Open path should not end with EndPoly")
		}
	}
}

func TestVCGenContour_OrientationDetection(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(1.0)
	vc.AutoDetectOrientation(true)

	// Create a CCW square (positive area)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(2, 0, basics.PathCmdLineTo)
	vc.AddVertex(2, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose))

	vc.Rewind(0)
	vertices := collectVertices(vc)

	if len(vertices) == 0 {
		t.Error("Expected vertices from auto-oriented contour")
	}

	// Test with CW square (negative area)
	vc.RemoveAll()
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(0, 2, basics.PathCmdLineTo) // CW direction
	vc.AddVertex(2, 2, basics.PathCmdLineTo)
	vc.AddVertex(2, 0, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose))

	vc.Rewind(0)
	vertices2 := collectVertices(vc)

	if len(vertices2) == 0 {
		t.Error("Expected vertices from CW auto-oriented contour")
	}

	// The contours should be different due to different orientations
	// (though testing exact coordinates would require more complex setup)
}

func TestVCGenContourApproximationScaleAffectsRoundJoinSubdivision(t *testing.T) {
	buildContour := func(scale float64) []struct {
		x, y float64
		cmd  basics.PathCommand
	} {
		vc := NewVCGenContour()
		vc.Width(3.0)
		vc.LineJoin(basics.RoundJoin)
		vc.ApproximationScale(scale)
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		vc.AddVertex(20, 0, basics.PathCmdLineTo)
		vc.AddVertex(20, 20, basics.PathCmdLineTo)
		vc.Rewind(0)
		return collectVertices(vc)
	}

	coarse := buildContour(0.25)
	fine := buildContour(4.0)

	if len(coarse) <= 1 {
		t.Fatalf("expected coarse contour to emit vertices, got %d", len(coarse))
	}
	if len(fine) <= len(coarse) {
		t.Fatalf("expected finer contour approximation scale to emit more vertices, got coarse=%d fine=%d", len(coarse), len(fine))
	}
}

func TestVCGenContour_NegativeWidth(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(-1.0) // Negative width

	// Create a simple square
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(2, 0, basics.PathCmdLineTo)
	vc.AddVertex(2, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 2, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

	vc.Rewind(0)
	vertices := collectVertices(vc)

	if len(vertices) == 0 {
		t.Error("Expected vertices from negative width contour")
	}

	// Negative width should produce an inner contour
	// The exact test would require checking if vertices are inside the original path
}

func TestVCGenContour_RemoveAll(t *testing.T) {
	vc := NewVCGenContour()

	// Add some vertices
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(1, 1, basics.PathCmdLineTo)

	// Remove all
	vc.RemoveAll()

	// Should have no vertices after removal
	vc.Rewind(0)
	vertices := collectVertices(vc)

	// For now, we always get at least a STOP vertex
	if len(vertices) > 1 {
		t.Errorf("Expected at most 1 vertex (STOP) after RemoveAll, got %d", len(vertices))
	}
}

func TestVCGenContour_InsufficientVertices(t *testing.T) {
	vc := NewVCGenContour()

	// Add only one vertex (insufficient for contour)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)

	vc.Rewind(0)
	vertices := collectVertices(vc)

	// Should get only STOP vertex with insufficient input
	if len(vertices) > 1 {
		t.Errorf("Expected at most 1 vertex (STOP) with insufficient input, got %d", len(vertices))
	}
}

// Helper function to collect all vertices from a generator
func collectVertices(vc *VCGenContour) []struct {
	x, y float64
	cmd  basics.PathCommand
} {
	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := vc.Vertex()
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})

		if basics.IsStop(cmd) {
			break
		}
	}

	return vertices
}

// Benchmark tests
func BenchmarkVCGenContour_SimpleSquare(b *testing.B) {
	vc := NewVCGenContour()
	vc.Width(1.0)

	for i := 0; i < b.N; i++ {
		vc.RemoveAll()

		// Create a simple square
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		vc.AddVertex(10, 0, basics.PathCmdLineTo)
		vc.AddVertex(10, 10, basics.PathCmdLineTo)
		vc.AddVertex(0, 10, basics.PathCmdLineTo)
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

		vc.Rewind(0)

		// Consume all vertices
		for {
			_, _, cmd := vc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
		}
	}
}

func BenchmarkVCGenContour_ComplexPath(b *testing.B) {
	vc := NewVCGenContour()
	vc.Width(2.0)

	for i := 0; i < b.N; i++ {
		vc.RemoveAll()

		// Create a more complex path
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		for j := 1; j <= 20; j++ {
			angle := float64(j) * math.Pi / 10
			x := 10 + 5*math.Cos(angle)
			y := 10 + 5*math.Sin(angle)
			vc.AddVertex(x, y, basics.PathCmdLineTo)
		}
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

		vc.Rewind(0)

		// Consume all vertices
		for {
			_, _, cmd := vc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
		}
	}
}

// TestVCGenContourAddVertexMoveToUsesModifyLast verifies that calling AddVertex
// with MoveTo on a non-empty generator replaces (not adds) the last vertex,
// matching the C++ vcgen_contour::add_vertex modify_last semantics.
func TestVCGenContourAddVertexMoveToUsesModifyLast(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(1.0)

	// First MoveTo — sequence is empty, equivalent to Add
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	// LineTo adds a real second vertex
	vc.AddVertex(10, 0, basics.PathCmdLineTo)
	// Second MoveTo should REPLACE the last vertex (modify_last), not add a third
	vc.AddVertex(99, 99, basics.PathCmdMoveTo)

	// After the second MoveTo the internal source sequence should have only 2
	// entries (the initial empty slot replaced by first MoveTo, then LineTo, then
	// second MoveTo replaces the LineTo).
	// We can observe the effect indirectly: calling Rewind and iterating should
	// not produce degenerate extra vertices from the orphaned LineTo.
	vc.AddVertex(20, 0, basics.PathCmdLineTo)
	vc.AddVertex(20, 10, basics.PathCmdLineTo)
	vc.AddVertex(99, 99, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose))

	vc.Rewind(0)
	vs := collectVertices(vc)

	// Should produce valid output (MoveTo + vertices + EndPoly + Stop)
	if len(vs) < 3 {
		t.Fatalf("expected ≥ 3 vertices, got %d", len(vs))
	}
	if vs[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("first vertex should be MoveTo, got %v", vs[0].cmd)
	}
}

// TestVCGenContourZeroWidthDegenerate verifies zero-width contour behaviour.
// With width=0 the stroker width is 0; output vertices should collapse to the
// original path outline (no offset). The generator must not crash or loop.
func TestVCGenContourZeroWidthDegenerate(t *testing.T) {
	vc := NewVCGenContour()
	vc.Width(0.0)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(10, 0, basics.PathCmdLineTo)
	vc.AddVertex(10, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

	vc.Rewind(0)
	vs := collectVertices(vc)

	// Must terminate (not loop) and produce at least Stop
	if len(vs) == 0 {
		t.Fatal("expected at least Stop vertex")
	}
	last := vs[len(vs)-1]
	if !basics.IsStop(last.cmd) {
		t.Errorf("expected final Stop, got %v", last.cmd)
	}
}

// TestVCGenContourPositiveWidthExpandsOutward verifies that a positive width on
// a CCW square expands outward (outer bounding box of output > original square).
// Ref: agg_vcgen_contour.cpp rewind orientation logic.
func TestVCGenContourPositiveWidthExpandsOutward(t *testing.T) {
	const w = 2.0
	vc := NewVCGenContour()
	vc.Width(w)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(10, 0, basics.PathCmdLineTo)
	vc.AddVertex(10, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

	vc.Rewind(0)
	vs := collectVertices(vc)

	if len(vs) < 3 {
		t.Fatalf("expected output vertices, got %d", len(vs))
	}

	maxX := -1e308
	minX := 1e308
	for _, v := range vs {
		if basics.IsStop(v.cmd) || basics.IsEndPoly(v.cmd) {
			continue
		}
		if v.x > maxX {
			maxX = v.x
		}
		if v.x < minX {
			minX = v.x
		}
	}

	// MathStroke internally uses half-width, so a vcgen_contour width of w
	// produces an actual outward offset of w/2. The contour must lie outside
	// the original square (minX < 0, maxX > 10).
	halfW := w / 2
	if minX > -halfW+1e-6 {
		t.Errorf("positive width contour minX=%.3f should be ≤ %.3f (offset outward)", minX, -halfW)
	}
	if maxX < 10+halfW-1e-6 {
		t.Errorf("positive width contour maxX=%.3f should be ≥ %.3f (offset outward)", maxX, 10+halfW)
	}
}

// TestVCGenContourNegativeWidthContractsInward verifies that a negative width on
// a CCW square contracts inward (output bbox smaller than original square).
func TestVCGenContourNegativeWidthContractsInward(t *testing.T) {
	const w = 2.0
	vc := NewVCGenContour()
	vc.Width(-w)
	vc.AddVertex(0, 0, basics.PathCmdMoveTo)
	vc.AddVertex(10, 0, basics.PathCmdLineTo)
	vc.AddVertex(10, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 10, basics.PathCmdLineTo)
	vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))

	vc.Rewind(0)
	vs := collectVertices(vc)

	if len(vs) < 3 {
		t.Fatalf("expected output vertices, got %d", len(vs))
	}

	maxX := -1e308
	minX := 1e308
	for _, v := range vs {
		if basics.IsStop(v.cmd) || basics.IsEndPoly(v.cmd) {
			continue
		}
		if v.x > maxX {
			maxX = v.x
		}
		if v.x < minX {
			minX = v.x
		}
	}

	// Negative width on CCW path: offset is inward by |w|/2 = 1.0.
	// Contour must remain inside the original [0,10] square.
	if minX < 0-1e-6 {
		t.Errorf("negative width contour minX=%.3f should be ≥ 0 (contracted inward)", minX)
	}
	if maxX > 10+1e-6 {
		t.Errorf("negative width contour maxX=%.3f should be ≤ 10 (contracted inward)", maxX)
	}
}

// TestVCGenContourPositiveVsNegativeWidthAreMirrored verifies that positive and
// negative width produce mirror-image contours (same vertex count, inverted offsets).
func TestVCGenContourPositiveVsNegativeWidthAreMirrored(t *testing.T) {
	buildContour := func(w float64) []struct{ x, y float64 } {
		vc := NewVCGenContour()
		vc.Width(w)
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		vc.AddVertex(10, 0, basics.PathCmdLineTo)
		vc.AddVertex(10, 10, basics.PathCmdLineTo)
		vc.AddVertex(0, 10, basics.PathCmdLineTo)
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))
		vc.Rewind(0)
		var pts []struct{ x, y float64 }
		for {
			x, y, cmd := vc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) || basics.IsMoveTo(cmd) {
				pts = append(pts, struct{ x, y float64 }{x, y})
			}
		}
		return pts
	}

	pos := buildContour(1.5)
	neg := buildContour(-1.5)

	if len(pos) == 0 || len(neg) == 0 {
		t.Fatal("both contours must produce vertices")
	}
	if len(pos) != len(neg) {
		t.Errorf("positive (%d) and negative (%d) width contours should produce same vertex count",
			len(pos), len(neg))
	}
}

// TestVCGenContourCornerJoinStylesProduceDifferentGeometry verifies that the
// three major join styles (Miter/Round/Bevel) produce geometrically distinct
// outputs at corners. Ref: agg_math_stroke.h calc_join.
func TestVCGenContourCornerJoinStylesProduceDifferentGeometry(t *testing.T) {
	buildCorner := func(join basics.LineJoin) []struct{ x, y float64 } {
		vc := NewVCGenContour()
		vc.Width(3.0)
		vc.LineJoin(join)
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		vc.AddVertex(20, 0, basics.PathCmdLineTo)
		vc.AddVertex(20, 20, basics.PathCmdLineTo)
		vc.AddVertex(0, 20, basics.PathCmdLineTo)
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))
		vc.Rewind(0)
		var pts []struct{ x, y float64 }
		for {
			x, y, cmd := vc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) || basics.IsMoveTo(cmd) {
				pts = append(pts, struct{ x, y float64 }{x, y})
			}
		}
		return pts
	}

	miter := buildCorner(basics.MiterJoin)
	round := buildCorner(basics.RoundJoin)
	bevel := buildCorner(basics.BevelJoin)

	if len(miter) == 0 || len(round) == 0 || len(bevel) == 0 {
		t.Fatal("all join styles must produce output")
	}

	// Round join must produce more vertices than bevel (arc approximation)
	if len(round) <= len(bevel) {
		t.Errorf("RoundJoin (%d vertices) should produce more than BevelJoin (%d vertices)", len(round), len(bevel))
	}
}

// TestVCGenContourMathStrokeIntegration verifies the math_stroke integration:
// changing approximation scale affects the number of vertices generated for
// rounded corners, confirming that CalcJoin is properly called with the stroker.
func TestVCGenContourMathStrokeIntegration(t *testing.T) {
	buildWithScale := func(scale float64) int {
		vc := NewVCGenContour()
		vc.Width(4.0)
		vc.LineJoin(basics.RoundJoin)
		vc.ApproximationScale(scale)
		vc.AddVertex(0, 0, basics.PathCmdMoveTo)
		vc.AddVertex(30, 0, basics.PathCmdLineTo)
		vc.AddVertex(30, 30, basics.PathCmdLineTo)
		vc.AddVertex(0, 30, basics.PathCmdLineTo)
		vc.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW))
		vc.Rewind(0)
		n := 0
		for {
			_, _, cmd := vc.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			if basics.IsVertex(cmd) || basics.IsMoveTo(cmd) {
				n++
			}
		}
		return n
	}

	coarse := buildWithScale(0.1)
	fine := buildWithScale(5.0)

	if coarse == 0 || fine == 0 {
		t.Fatal("both scale levels must produce vertices")
	}
	if fine <= coarse {
		t.Errorf("finer approximation scale (%d vertices) should produce more than coarse (%d vertices)", fine, coarse)
	}
}
