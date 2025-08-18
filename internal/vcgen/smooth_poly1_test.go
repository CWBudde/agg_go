package vcgen

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestVCGenSmoothPoly1_Basic(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Add a simple square
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)

	gen.Rewind(0)

	// First vertex should be MoveTo
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v at (%f,%f)", cmd, x, y)
	}

	// Collect all generated vertices
	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}
	vertices = append(vertices, struct {
		x, y float64
		cmd  basics.PathCommand
	}{x, y, cmd})

	for {
		x, y, cmd = gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should generate cubic Bezier curves for rounded corners
	// Expect: MoveTo, Curve4, Curve4, Curve4 (for each corner), then EndPoly
	curve4Count := 0
	for _, v := range vertices {
		if v.cmd == basics.PathCmdCurve4 {
			curve4Count++
		}
	}

	// Should have multiple Curve4 commands for cubic Bezier curves at corners
	if curve4Count < 8 { // At least 2 control points + end point for each corner
		t.Errorf("Expected multiple Curve4 commands for corner smoothing, got %d", curve4Count)
	}

	// Last vertex should be EndPoly
	if len(vertices) > 0 {
		lastVertex := vertices[len(vertices)-1]
		if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.cmd)
		}
	}
}

func TestVCGenSmoothPoly1_SmoothValue(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Test default smooth value
	defaultSmooth := gen.SmoothValue()
	if defaultSmooth <= 0 || defaultSmooth > 2.0 {
		t.Errorf("Default smooth value should be reasonable, got %f", defaultSmooth)
	}

	// Test setting custom smooth values
	testValues := []float64{0.0, 0.5, 1.0}

	for _, value := range testValues {
		gen.SetSmoothValue(value)
		if math.Abs(gen.SmoothValue()-value) > 1e-10 {
			t.Errorf("Expected smooth value %f, got %f", value, gen.SmoothValue())
		}
	}

	// Test with different smooth values on same polygon
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 0, basics.PathCmdLineTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(0, 50, basics.PathCmdLineTo)

	// Test minimum smoothing (0.0)
	gen.SetSmoothValue(0.0)
	gen.Rewind(0)

	zeroSmoothVertices := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		zeroSmoothVertices++
	}

	// Test maximum smoothing (1.0)
	gen.SetSmoothValue(1.0)
	gen.Rewind(0)

	maxSmoothVertices := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		maxSmoothVertices++
	}

	// Different smooth values may generate different number of vertices
	// (This is implementation dependent, but both should generate some vertices)
	if zeroSmoothVertices == 0 || maxSmoothVertices == 0 {
		t.Error("Both zero and max smoothing should generate vertices")
	}
}

func TestVCGenSmoothPoly1_Triangle(t *testing.T) {
	gen := NewVCGenSmoothPoly1()
	gen.SetSmoothValue(0.5)

	// Add a triangle
	gen.AddVertex(50, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)

	gen.Rewind(0)

	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should have smoothed the 3 corners of the triangle
	if len(vertices) == 0 {
		t.Error("Triangle smoothing should generate vertices")
	}

	// First vertex should be MoveTo
	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", vertices[0].cmd)
	}

	// Should have Curve4 commands for corner smoothing
	hasCurve4 := false
	for _, v := range vertices {
		if v.cmd == basics.PathCmdCurve4 {
			hasCurve4 = true
			break
		}
	}

	if !hasCurve4 {
		t.Error("Triangle smoothing should generate Curve4 commands")
	}
}

func TestVCGenSmoothPoly1_ClosedPolygon(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Add a closed pentagon
	gen.AddVertex(50, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 25, basics.PathCmdLineTo)
	gen.AddVertex(80, 75, basics.PathCmdLineTo)
	gen.AddVertex(20, 75, basics.PathCmdLineTo)
	gen.AddVertex(0, 25, basics.PathCmdLineTo)
	gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose) // Close the polygon

	gen.Rewind(0)

	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	if len(vertices) == 0 {
		t.Error("Closed pentagon should generate smooth vertices")
	}

	// Should end with EndPoly with close flag
	if len(vertices) > 0 {
		lastVertex := vertices[len(vertices)-1]
		if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Closed polygon should end with EndPoly, got %v", lastVertex.cmd)
		}

		if (lastVertex.cmd & basics.PathFlagClose) == 0 {
			t.Error("Closed polygon should have close flag")
		}
	}
}

func TestVCGenSmoothPoly1_OpenPolygon(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Add an open polygon (no close command)
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 0, basics.PathCmdLineTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(0, 50, basics.PathCmdLineTo)

	gen.Rewind(0)

	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	if len(vertices) == 0 {
		t.Error("Open polygon should generate smooth vertices")
	}

	// Should end with EndPoly without close flag
	if len(vertices) > 0 {
		lastVertex := vertices[len(vertices)-1]
		if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Open polygon should end with EndPoly, got %v", lastVertex.cmd)
		}

		if (lastVertex.cmd & basics.PathFlagClose) != 0 {
			t.Error("Open polygon should not have close flag")
		}
	}
}

func TestVCGenSmoothPoly1_InsufficientVertices(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Test with single vertex
	gen.AddVertex(50, 50, basics.PathCmdMoveTo)
	gen.Rewind(0)

	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single vertex should result in Stop, got %v at (%f,%f)", cmd, x, y)
	}

	// Test with two vertices (line segment)
	gen.RemoveAll()
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.Rewind(0)

	var twoVertexCommands []basics.PathCommand
	for {
		_, _, cmd = gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		twoVertexCommands = append(twoVertexCommands, cmd)
	}

	// Two vertices should still be processed (as a line)
	if len(twoVertexCommands) != 2 { // MoveTo + LineTo
		t.Errorf("Two vertices should produce line, got %d commands", len(twoVertexCommands))
	}

	if twoVertexCommands[0] != basics.PathCmdMoveTo || twoVertexCommands[1] != basics.PathCmdLineTo {
		t.Errorf("Two vertices should produce MoveTo+LineTo, got %v,%v",
			twoVertexCommands[0], twoVertexCommands[1])
	}
}

func TestVCGenSmoothPoly1_MoveTo(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// MoveTo should modify the last vertex
	gen.AddVertex(10, 10, basics.PathCmdMoveTo)
	gen.AddVertex(20, 20, basics.PathCmdMoveTo) // Should replace previous
	gen.AddVertex(50, 20, basics.PathCmdLineTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(20, 50, basics.PathCmdLineTo)

	gen.Rewind(0)

	// First vertex should start from the final MoveTo position
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}

	// Should be close to final MoveTo position (smoothing algorithm may adjust slightly)
	if math.Abs(x-20) > 30 || math.Abs(y-20) > 30 {
		t.Errorf("First vertex should be near final MoveTo position (20,20), got (%f,%f)", x, y)
	}
}

func TestVCGenSmoothPoly1_NonVertexCommands(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 0, basics.PathCmdLineTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(0, 50, basics.PathCmdLineTo)
	// Add non-vertex command (should set closed flag)
	gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose)

	gen.Rewind(0)

	// Should still process the smoothing
	hasVertices := false
	var lastCmd basics.PathCommand

	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		hasVertices = true
		lastCmd = cmd
	}

	if !hasVertices {
		t.Error("Non-vertex commands shouldn't prevent smoothing")
	}

	// Should respect the close flag
	if (lastCmd & basics.PathFlagClose) == 0 {
		t.Error("Should respect close flag from non-vertex command")
	}
}

func TestVCGenSmoothPoly1_RemoveAll(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Add vertices
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	// Remove all
	gen.RemoveAll()

	gen.Rewind(0)
	x, y, cmd := gen.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("After RemoveAll, should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenSmoothPoly1_MultipleRewinds(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Add a rectangle
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)

	// First iteration
	gen.Rewind(0)
	var firstIteration []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstIteration = append(firstIteration, struct{ x, y float64 }{x, y})
	}

	// Second iteration should produce same results
	gen.Rewind(0)
	var secondIteration []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondIteration = append(secondIteration, struct{ x, y float64 }{x, y})
	}

	if len(firstIteration) != len(secondIteration) {
		t.Errorf("Multiple rewinds should produce same number of vertices: %d vs %d",
			len(firstIteration), len(secondIteration))
	}

	for i := 0; i < len(firstIteration) && i < len(secondIteration); i++ {
		if math.Abs(firstIteration[i].x-secondIteration[i].x) > 1e-10 ||
			math.Abs(firstIteration[i].y-secondIteration[i].y) > 1e-10 {
			t.Errorf("Vertex %d differs between iterations: (%f,%f) vs (%f,%f)",
				i, firstIteration[i].x, firstIteration[i].y,
				secondIteration[i].x, secondIteration[i].y)
		}
	}
}

func TestVCGenSmoothPoly1_CornerCalculation(t *testing.T) {
	gen := NewVCGenSmoothPoly1()
	gen.SetSmoothValue(0.5)

	// Create a right angle corner that should be smoothed
	gen.AddVertex(0, 50, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo) // Horizontal segment
	gen.AddVertex(50, 0, basics.PathCmdLineTo)  // Vertical segment (90-degree corner)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	var curve4Points []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdCurve4 {
			curve4Points = append(curve4Points, struct{ x, y float64 }{x, y})
		}
	}

	// Should have generated control points for the corner
	if len(curve4Points) == 0 {
		t.Error("Right angle corner should generate Curve4 control points")
	}

	// Control points should be reasonable (not at original corner position)
	for _, pt := range curve4Points {
		if pt.x == 50 && pt.y == 50 {
			t.Error("Control points should be offset from original corner")
		}
	}
}

func TestVCGenSmoothPoly1_EdgeCases(t *testing.T) {
	gen := NewVCGenSmoothPoly1()

	// Test with coincident vertices
	gen.AddVertex(50, 50, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo) // Same point
	gen.AddVertex(100, 50, basics.PathCmdLineTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(50, 100, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Should handle degenerate case gracefully
	hasVertices := false
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		hasVertices = true
	}

	if !hasVertices {
		t.Error("Coincident vertices should still allow some smoothing")
	}

	// Test with very small polygon
	gen.RemoveAll()
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(1, 0, basics.PathCmdLineTo)
	gen.AddVertex(1, 1, basics.PathCmdLineTo)
	gen.AddVertex(0, 1, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Should handle very small polygons
	smallPolyVertices := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		smallPolyVertices++
	}

	if smallPolyVertices == 0 {
		t.Error("Very small polygon should still be processed")
	}
}

// Benchmark tests
func BenchmarkVCGenSmoothPoly1_Rectangle(b *testing.B) {
	gen := NewVCGenSmoothPoly1()
	gen.SetSmoothValue(0.5)

	// Setup rectangle
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Rewind(0)
		for {
			_, _, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkVCGenSmoothPoly1_ComplexPolygon(b *testing.B) {
	gen := NewVCGenSmoothPoly1()
	gen.SetSmoothValue(0.7)

	// Setup complex polygon (octagon)
	gen.AddVertex(30, 0, basics.PathCmdMoveTo)
	gen.AddVertex(70, 0, basics.PathCmdLineTo)
	gen.AddVertex(100, 30, basics.PathCmdLineTo)
	gen.AddVertex(100, 70, basics.PathCmdLineTo)
	gen.AddVertex(70, 100, basics.PathCmdLineTo)
	gen.AddVertex(30, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 70, basics.PathCmdLineTo)
	gen.AddVertex(0, 30, basics.PathCmdLineTo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Rewind(0)
		for {
			_, _, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkVCGenSmoothPoly1_MaxSmoothing(b *testing.B) {
	gen := NewVCGenSmoothPoly1()
	gen.SetSmoothValue(1.0) // Maximum smoothing

	// Setup triangle with sharp corners
	gen.AddVertex(50, 0, basics.PathCmdMoveTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Rewind(0)
		for {
			_, _, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
