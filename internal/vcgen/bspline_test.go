package vcgen

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestVCGenBSpline_Basic(t *testing.T) {
	gen := NewVCGenBSpline()

	// Add a simple B-spline with 4 control points
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(25, 75, basics.PathCmdLineTo)
	gen.AddVertex(75, 75, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	// First vertex should be MoveTo
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v at (%f,%f)", cmd, x, y)
	}

	// Collect all spline points
	var splinePoints []struct{ x, y float64 }
	splinePoints = append(splinePoints, struct{ x, y float64 }{x, y})

	for {
		x, y, cmd = gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdLineTo {
			splinePoints = append(splinePoints, struct{ x, y float64 }{x, y})
		}
	}

	// Should have multiple interpolated points
	if len(splinePoints) < 3 {
		t.Errorf("B-spline should generate multiple points, got %d", len(splinePoints))
	}

	// Points should form a smooth curve (basic sanity check)
	// First point should be close to first control point
	if math.Abs(splinePoints[0].x-0) > 10 || math.Abs(splinePoints[0].y-0) > 10 {
		t.Errorf("First spline point should be close to first control point, got (%f,%f)",
			splinePoints[0].x, splinePoints[0].y)
	}
}

func TestVCGenBSpline_InterpolationStep(t *testing.T) {
	gen := NewVCGenBSpline()

	// Test default interpolation step
	defaultStep := gen.InterpolationStep()
	if defaultStep <= 0 {
		t.Errorf("Default interpolation step should be positive, got %f", defaultStep)
	}

	// Set a custom interpolation step
	customStep := 0.1
	gen.SetInterpolationStep(customStep)

	if math.Abs(gen.InterpolationStep()-customStep) > 1e-10 {
		t.Errorf("Expected interpolation step %f, got %f", customStep, gen.InterpolationStep())
	}

	// Add control points
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	// Smaller step should generate more points
	gen.Rewind(0)
	smallStepCount := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		smallStepCount++
	}

	// Test with larger step
	gen.SetInterpolationStep(0.5)
	gen.RemoveAll()
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)
	largeStepCount := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		largeStepCount++
	}

	// Smaller step should generate more points
	if smallStepCount <= largeStepCount {
		t.Errorf("Smaller interpolation step should generate more points: %d vs %d",
			smallStepCount, largeStepCount)
	}
}

func TestVCGenBSpline_MoveTo(t *testing.T) {
	gen := NewVCGenBSpline()

	// MoveTo should modify the last vertex
	gen.AddVertex(10, 10, basics.PathCmdMoveTo)
	gen.AddVertex(20, 20, basics.PathCmdMoveTo) // Should replace previous
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	// First point should start from the final MoveTo position (approximately)
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}

	// The spline should incorporate the final MoveTo position in its calculation
	// (exact position depends on B-spline mathematics, but should be reasonable)
	if x < 10 || x > 30 || y < 0 || y > 30 {
		t.Errorf("First spline point should be influenced by MoveTo position, got (%f,%f)", x, y)
	}
}

func TestVCGenBSpline_ClosedPath(t *testing.T) {
	gen := NewVCGenBSpline()

	// Add control points for a closed path
	gen.AddVertex(0, 50, basics.PathCmdMoveTo)
	gen.AddVertex(50, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 50, basics.PathCmdLineTo)
	gen.AddVertex(50, 0, basics.PathCmdLineTo)
	gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose) // Close the path

	gen.Rewind(0)

	// Collect all vertices
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

	// Should have MoveTo, multiple LineTo, and EndPoly
	if len(vertices) < 3 {
		t.Errorf("Closed B-spline should have multiple vertices, got %d", len(vertices))
	}

	// First vertex should be MoveTo
	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", vertices[0].cmd)
	}

	// Last vertex should be EndPoly with close flag
	lastVertex := vertices[len(vertices)-1]
	if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.cmd)
	}

	// Should have close flag
	if (lastVertex.cmd & basics.PathFlagClose) == 0 {
		t.Error("Closed path should have close flag")
	}
}

func TestVCGenBSpline_OpenPath(t *testing.T) {
	gen := NewVCGenBSpline()

	// Add control points for an open path
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(25, 50, basics.PathCmdLineTo)
	gen.AddVertex(75, 50, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Collect all vertices
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

	// Should have MoveTo and multiple LineTo
	if len(vertices) < 2 {
		t.Errorf("Open B-spline should have multiple vertices, got %d", len(vertices))
	}

	// First vertex should be MoveTo
	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", vertices[0].cmd)
	}

	// Last vertex should be EndPoly without close flag
	lastVertex := vertices[len(vertices)-1]
	if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.cmd)
	}

	// Should not have close flag
	if (lastVertex.cmd & basics.PathFlagClose) != 0 {
		t.Error("Open path should not have close flag")
	}
}

func TestVCGenBSpline_InsufficientPoints(t *testing.T) {
	gen := NewVCGenBSpline()

	// Test with only one point
	gen.AddVertex(50, 50, basics.PathCmdMoveTo)
	gen.Rewind(0)

	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single point should result in Stop, got %v at (%f,%f)", cmd, x, y)
	}

	// Test with two points - C++ outputs them directly, not through spline
	gen.RemoveAll()
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.Rewind(0)

	// Should output the two points directly
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("First of two points should be MoveTo, got %v at (%f,%f)", cmd, x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Second of two points should be LineTo, got %v at (%f,%f)", cmd, x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After two points should be Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestVCGenBSpline_RemoveAll(t *testing.T) {
	gen := NewVCGenBSpline()

	// Add some vertices
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

func TestVCGenBSpline_MultipleRewinds(t *testing.T) {
	gen := NewVCGenBSpline()

	// Add control points
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(33, 100, basics.PathCmdLineTo)
	gen.AddVertex(66, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	// First iteration
	gen.Rewind(0)
	var firstIteration []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo || cmd == basics.PathCmdLineTo {
			firstIteration = append(firstIteration, struct{ x, y float64 }{x, y})
		}
	}

	// Second iteration should produce same results
	gen.Rewind(0)
	var secondIteration []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo || cmd == basics.PathCmdLineTo {
			secondIteration = append(secondIteration, struct{ x, y float64 }{x, y})
		}
	}

	if len(firstIteration) != len(secondIteration) {
		t.Errorf("Multiple rewinds should produce same number of points: %d vs %d",
			len(firstIteration), len(secondIteration))
	}

	for i := 0; i < len(firstIteration) && i < len(secondIteration); i++ {
		if math.Abs(firstIteration[i].x-secondIteration[i].x) > 1e-10 ||
			math.Abs(firstIteration[i].y-secondIteration[i].y) > 1e-10 {
			t.Errorf("Point %d differs between iterations: (%f,%f) vs (%f,%f)",
				i, firstIteration[i].x, firstIteration[i].y,
				secondIteration[i].x, secondIteration[i].y)
		}
	}
}

func TestVCGenBSpline_NonVertexCommands(t *testing.T) {
	gen := NewVCGenBSpline()

	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)
	// Add non-vertex command
	gen.AddVertex(0, 0, basics.PathCmdEndPoly)

	gen.Rewind(0)

	// Should still process the B-spline normally
	vertexCount := 0
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	if vertexCount == 0 {
		t.Error("Non-vertex commands shouldn't prevent B-spline generation")
	}
}

func TestVCGenBSpline_SmoothCurve(t *testing.T) {
	gen := NewVCGenBSpline()
	gen.SetInterpolationStep(0.1) // Fine interpolation

	// Create a control polygon that should result in a smooth curve
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(0, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	var points []struct{ x, y float64 }
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo || cmd == basics.PathCmdLineTo {
			points = append(points, struct{ x, y float64 }{x, y})
		}
	}

	// Should have many interpolated points
	if len(points) < 5 {
		t.Errorf("Fine interpolation should produce many points, got %d", len(points))
	}

	// Check that curve is reasonably smooth (no huge jumps)
	for i := 1; i < len(points); i++ {
		dx := points[i].x - points[i-1].x
		dy := points[i].y - points[i-1].y
		distance := math.Sqrt(dx*dx + dy*dy)

		// With fine interpolation, consecutive points should be close
		if distance > 20.0 { // Reasonable threshold for smooth curve
			t.Errorf("Consecutive points too far apart at %d: distance %f", i, distance)
		}
	}
}

func TestVCGenBSpline_EdgeCases(t *testing.T) {
	gen := NewVCGenBSpline()

	// Test with coincident points
	gen.AddVertex(50, 50, basics.PathCmdMoveTo)
	gen.AddVertex(50, 50, basics.PathCmdLineTo)  // Same point
	gen.AddVertex(50, 50, basics.PathCmdLineTo)  // Same point again
	gen.AddVertex(100, 50, basics.PathCmdLineTo) // Different point

	gen.Rewind(0)

	// Should still generate some spline (might be degenerate but shouldn't crash)
	hasVertices := false
	for {
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		hasVertices = true
	}

	// Might or might not produce vertices depending on B-spline implementation,
	// but shouldn't crash
	_ = hasVertices

	// Test with very small interpolation step (should be clamped)
	gen.RemoveAll()
	gen.SetInterpolationStep(1e-10) // Very small, should be clamped to 1e-3
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(1, 1, basics.PathCmdLineTo)
	gen.AddVertex(2, 0, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Should not cause infinite loops or excessive memory usage due to clamping
	count := 0
	for count < 5000 { // Safety limit
		_, _, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		count++
	}

	// With clamping to 1e-3, should have reasonable number of vertices
	if count >= 5000 {
		t.Errorf("Very small interpolation step should be clamped to prevent excessive vertex generation, got %d vertices", count)
	}

	// Should be less than 3000 vertices for this simple 3-point spline with min step
	if count > 3000 {
		t.Errorf("Expected reasonable number of vertices, got %d", count)
	}
}

// Benchmark tests
func BenchmarkVCGenBSpline_Generation(b *testing.B) {
	gen := NewVCGenBSpline()
	gen.SetInterpolationStep(0.05) // Moderate interpolation

	// Setup B-spline with several control points
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(25, 100, basics.PathCmdLineTo)
	gen.AddVertex(75, 100, basics.PathCmdLineTo)
	gen.AddVertex(100, 0, basics.PathCmdLineTo)

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

func BenchmarkVCGenBSpline_FineInterpolation(b *testing.B) {
	gen := NewVCGenBSpline()
	gen.SetInterpolationStep(0.01) // Fine interpolation

	// Setup complex B-spline
	gen.AddVertex(0, 50, basics.PathCmdMoveTo)
	for i := 1; i < 10; i++ {
		x := float64(i * 10)
		y := 50 + 30*math.Sin(float64(i)*0.5)
		gen.AddVertex(x, y, basics.PathCmdLineTo)
	}

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

func BenchmarkVCGenBSpline_ManyControlPoints(b *testing.B) {
	gen := NewVCGenBSpline()
	gen.SetInterpolationStep(0.1)

	// Setup B-spline with many control points
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	for i := 1; i < 50; i++ {
		x := float64(i * 2)
		y := 50 * math.Sin(float64(i)*0.2)
		gen.AddVertex(x, y, basics.PathCmdLineTo)
	}

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

// Test fixes for the TODO items
func TestVCGenBSpline_FixedTODOs(t *testing.T) {
	t.Run("VerySmallInterpolationStep", func(t *testing.T) {
		gen := NewVCGenBSpline()

		// Test very small step (should be clamped to minimum)
		gen.SetInterpolationStep(1e-10)
		step := gen.InterpolationStep()
		if step < 1e-3 {
			t.Errorf("Expected step to be clamped to minimum 1e-3, got %e", step)
		}

		// Test very large step (should be clamped to maximum)
		gen.SetInterpolationStep(2.0)
		step = gen.InterpolationStep()
		if step > 1.0 {
			t.Errorf("Expected step to be clamped to maximum 1.0, got %f", step)
		}
	})

	t.Run("MultipleRewindsConsistency", func(t *testing.T) {
		gen := NewVCGenBSpline()
		gen.SetInterpolationStep(0.1)

		// Add control points
		gen.AddVertex(0, 0, basics.PathCmdMoveTo)
		gen.AddVertex(50, 100, basics.PathCmdLineTo)
		gen.AddVertex(100, 50, basics.PathCmdLineTo)
		gen.AddVertex(150, 0, basics.PathCmdLineTo)

		// First iteration
		gen.Rewind(0)
		var firstRun []struct {
			x, y float64
			cmd  basics.PathCommand
		}
		for {
			x, y, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			firstRun = append(firstRun, struct {
				x, y float64
				cmd  basics.PathCommand
			}{x, y, cmd})
		}

		// Second iteration should be identical
		gen.Rewind(0)
		var secondRun []struct {
			x, y float64
			cmd  basics.PathCommand
		}
		for {
			x, y, cmd := gen.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			secondRun = append(secondRun, struct {
				x, y float64
				cmd  basics.PathCommand
			}{x, y, cmd})
		}

		if len(firstRun) != len(secondRun) {
			t.Errorf("Multiple rewinds produced different vertex counts: %d vs %d",
				len(firstRun), len(secondRun))
		}

		for i := 0; i < len(firstRun) && i < len(secondRun); i++ {
			if math.Abs(firstRun[i].x-secondRun[i].x) > 1e-10 ||
				math.Abs(firstRun[i].y-secondRun[i].y) > 1e-10 ||
				firstRun[i].cmd != secondRun[i].cmd {
				t.Errorf("Multiple rewinds produced different vertex %d: "+
					"(%f,%f,%v) vs (%f,%f,%v)", i,
					firstRun[i].x, firstRun[i].y, firstRun[i].cmd,
					secondRun[i].x, secondRun[i].y, secondRun[i].cmd)
			}
		}
	})

	t.Run("InsufficientPointsHandling", func(t *testing.T) {
		gen := NewVCGenBSpline()

		// Test with 0 points
		gen.Rewind(0)
		_, _, cmd := gen.Vertex()
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected Stop for 0 points, got %v", cmd)
		}

		// Test with 1 point
		gen.AddVertex(50, 50, basics.PathCmdMoveTo)
		gen.Rewind(0)
		_, _, cmd = gen.Vertex()
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected Stop for 1 point, got %v", cmd)
		}

		// Test with 2 points (should output them directly)
		gen.AddVertex(100, 100, basics.PathCmdLineTo)
		gen.Rewind(0)

		// Should get MoveTo
		_, _, cmd = gen.Vertex()
		if cmd != basics.PathCmdMoveTo {
			t.Errorf("Expected MoveTo for first of 2 points, got %v", cmd)
		}

		// Should get LineTo
		_, _, cmd = gen.Vertex()
		if cmd != basics.PathCmdLineTo {
			t.Errorf("Expected LineTo for second of 2 points, got %v", cmd)
		}

		// Should get Stop
		_, _, cmd = gen.Vertex()
		if cmd != basics.PathCmdStop {
			t.Errorf("Expected Stop after 2 points, got %v", cmd)
		}
	})

	t.Run("MoveToModifyLastBehavior", func(t *testing.T) {
		gen := NewVCGenBSpline()

		// First MoveTo should add a point
		gen.AddVertex(10, 10, basics.PathCmdMoveTo)
		if gen.srcVertices.Size() != 1 {
			t.Errorf("Expected 1 vertex after first MoveTo, got %d", gen.srcVertices.Size())
		}

		// Second MoveTo should modify (replace) the last point
		gen.AddVertex(20, 20, basics.PathCmdMoveTo)
		if gen.srcVertices.Size() != 1 {
			t.Errorf("Expected 1 vertex after second MoveTo, got %d", gen.srcVertices.Size())
		}

		point := gen.srcVertices.At(0)
		if point.X != 20 || point.Y != 20 {
			t.Errorf("Expected last point to be (20,20), got (%f,%f)", point.X, point.Y)
		}
	})

	t.Run("ClosedPathHandling", func(t *testing.T) {
		gen := NewVCGenBSpline()
		gen.SetInterpolationStep(0.5)

		// Add closed path
		gen.AddVertex(0, 0, basics.PathCmdMoveTo)
		gen.AddVertex(50, 100, basics.PathCmdLineTo)
		gen.AddVertex(100, 50, basics.PathCmdLineTo)
		gen.AddVertex(50, 0, basics.PathCmdLineTo)
		gen.AddVertex(0, 0, basics.PathCmdEndPoly|basics.PathFlagClose)

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

		// Should have vertices including EndPoly with close flag
		if len(vertices) < 2 {
			t.Errorf("Expected multiple vertices for closed path, got %d", len(vertices))
		}

		// Last vertex should be EndPoly with close flag
		lastVertex := vertices[len(vertices)-1]
		if (lastVertex.cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Expected last command to be EndPoly, got %v", lastVertex.cmd)
		}
		if (lastVertex.cmd & basics.PathFlagClose) == 0 {
			t.Error("Expected close flag on closed path")
		}
	})
}
