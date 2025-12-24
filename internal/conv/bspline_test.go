package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestConvBSpline_Basic(t *testing.T) {
	// Create a simple path with control points
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)

	// Collect all vertices
	var resultVertices []CurveVertex
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have generated smooth B-spline curve
	if len(resultVertices) < 4 {
		t.Errorf("B-spline should generate multiple vertices, got %d", len(resultVertices))
	}

	// First vertex should be MoveTo
	if resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}

	// Rest should be LineTo (from B-spline approximation), except the last should be EndPoly
	for i := 1; i < len(resultVertices)-1; i++ {
		if resultVertices[i].Cmd != basics.PathCmdLineTo {
			t.Errorf("B-spline vertex %d should be LineTo, got %v", i, resultVertices[i].Cmd)
		}
	}

	// Last vertex should be EndPoly
	if len(resultVertices) > 1 {
		lastVertex := resultVertices[len(resultVertices)-1]
		if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.Cmd)
		}
	}
}

func TestConvBSpline_InterpolationStep(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	// Test default interpolation step
	defaultStep := bspline.InterpolationStep()
	if defaultStep <= 0 {
		t.Errorf("Default interpolation step should be positive, got %f", defaultStep)
	}

	// Test setting custom interpolation step
	customStep := 0.05
	bspline.SetInterpolationStep(customStep)

	if math.Abs(bspline.InterpolationStep()-customStep) > 1e-10 {
		t.Errorf("Expected interpolation step %f, got %f", customStep, bspline.InterpolationStep())
	}

	// Smaller step should generate more vertices
	source.index = 0 // Reset source
	bspline.Rewind(0)

	smallStepCount := 0
	for {
		_, _, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		smallStepCount++
	}

	// Test with larger step
	bspline.SetInterpolationStep(0.5)
	source.index = 0
	bspline.Rewind(0)

	largeStepCount := 0
	for {
		_, _, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		largeStepCount++
	}

	if smallStepCount <= largeStepCount {
		t.Errorf("Smaller step should generate more vertices: %d vs %d",
			smallStepCount, largeStepCount)
	}
}

func TestConvBSpline_EmptyPath(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)
	x, y, cmd := bspline.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestConvBSpline_InsufficientControlPoints(t *testing.T) {
	// Test with only one control point
	vertices := []CurveVertex{
		{X: 50, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)
	x, y, cmd := bspline.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Single control point should result in Stop, got %v at (%f,%f)", cmd, x, y)
	}

	// Test with two control points - BSpline can still produce output
	// The implementation may generate a degenerate curve or simple line
	vertices = []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(vertices)
	bspline = NewConvBSpline(source)

	bspline.Rewind(0)
	x, y, cmd = bspline.Vertex()

	// With 2 control points, BSpline may produce output (degenerate case)
	// Just verify it doesn't panic and returns valid command
	if cmd != basics.PathCmdMoveTo && cmd != basics.PathCmdLineTo && cmd != basics.PathCmdStop {
		t.Errorf("Expected valid path command, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestConvBSpline_ClosedPath(t *testing.T) {
	// Create a closed quadrilateral
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathFlagClose},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)

	var resultVertices []CurveVertex
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	if len(resultVertices) == 0 {
		t.Error("Closed path should generate B-spline vertices")
	}

	// Should have EndPoly command at the end
	lastVertex := resultVertices[len(resultVertices)-1]
	if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Closed path should end with EndPoly, got %v", lastVertex.Cmd)
	}

	// Should have close flag
	if (lastVertex.Cmd & basics.PathFlagClose) == 0 {
		t.Error("Closed path should have close flag")
	}
}

func TestConvBSpline_OpenPath(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 33, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 66, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)

	var resultVertices []CurveVertex
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	if len(resultVertices) == 0 {
		t.Error("Open path should generate B-spline vertices")
	}

	// Should have EndPoly command at the end
	lastVertex := resultVertices[len(resultVertices)-1]
	if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Open path should end with EndPoly, got %v", lastVertex.Cmd)
	}

	// Should not have close flag
	if (lastVertex.Cmd & basics.PathFlagClose) != 0 {
		t.Error("Open path should not have close flag")
	}
}

func TestConvBSpline_SmoothCurve(t *testing.T) {
	// Create a path that should result in a smooth B-spline
	vertices := []CurveVertex{
		{X: 0, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)
	bspline.SetInterpolationStep(0.1) // Fine interpolation for smooth curve

	bspline.Rewind(0)

	var points []struct{ x, y float64 }
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo || cmd == basics.PathCmdLineTo {
			points = append(points, struct{ x, y float64 }{x, y})
		}
	}

	// Should generate many smooth points
	if len(points) < 10 {
		t.Errorf("Fine interpolation should generate many points, got %d", len(points))
	}

	// Check smoothness by verifying no large jumps between consecutive points
	for i := 1; i < len(points); i++ {
		dx := points[i].x - points[i-1].x
		dy := points[i].y - points[i-1].y
		distance := math.Sqrt(dx*dx + dy*dy)

		// With fine interpolation, consecutive points should be close
		if distance > 15.0 {
			t.Errorf("Large jump between consecutive points at %d: distance %f", i, distance)
		}
	}
}

func TestConvBSpline_MultipleRewinds(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	// First iteration
	bspline.Rewind(0)
	var firstIteration []CurveVertex
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstIteration = append(firstIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Second iteration should produce same results
	source.index = 0 // Reset source
	bspline.Rewind(0)
	var secondIteration []CurveVertex
	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondIteration = append(secondIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	if len(firstIteration) != len(secondIteration) {
		t.Errorf("Multiple rewinds should produce same number of vertices: %d vs %d",
			len(firstIteration), len(secondIteration))
	}

	for i := 0; i < len(firstIteration) && i < len(secondIteration); i++ {
		if firstIteration[i] != secondIteration[i] {
			t.Errorf("Vertex %d differs between iterations: %+v vs %+v",
				i, firstIteration[i], secondIteration[i])
		}
	}
}

func TestConvBSpline_LinearPath(t *testing.T) {
	// Test that linear paths are handled correctly (might produce a linear B-spline)
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)

	// Should still process the path (B-spline can handle linear control points)
	hasVertices := false
	for {
		_, _, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		hasVertices = true
	}

	// Linear paths should still be processed by B-spline generator
	if !hasVertices {
		t.Error("Linear path should still be processed by B-spline")
	}
}

func TestConvBSpline_ComplexPath(t *testing.T) {
	// Test with a more complex path having multiple sub-paths
	// Note: B-spline converter processes the entire input as a single path,
	// so we expect one continuous B-spline curve through all control points
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 125, Y: 150, Cmd: basics.PathCmdLineTo},
		{X: 150, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	bspline.Rewind(0)

	var resultVertices []CurveVertex
	var moveToCount int
	var endPolyCount int

	for {
		x, y, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})

		if cmd == basics.PathCmdMoveTo {
			moveToCount++
		}
		if (cmd & basics.PathCmdMask) == basics.PathCmdEndPoly {
			endPolyCount++
		}
	}

	// Should have processed all points as a single B-spline path
	if moveToCount != 1 {
		t.Errorf("Expected 1 MoveTo command for single B-spline path, got %d", moveToCount)
	}

	if endPolyCount != 1 {
		t.Errorf("Expected 1 EndPoly command for single B-spline path, got %d", endPolyCount)
	}

	// Should generate vertices for the smooth curve
	if len(resultVertices) < 10 {
		t.Errorf("Complex path should generate many vertices, got %d", len(resultVertices))
	}
}

func TestConvBSpline_EdgeCases(t *testing.T) {
	// Test with very small interpolation step
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 1, Y: 1, Cmd: basics.PathCmdLineTo},
		{X: 2, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)
	bspline.SetInterpolationStep(1e-6)

	bspline.Rewind(0)

	// Should not cause excessive vertex generation or infinite loops
	vertexCount := 0
	maxVertices := 10000 // Generous limit for very small interpolation steps
	for vertexCount < maxVertices {
		_, _, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	if vertexCount >= maxVertices {
		t.Errorf("Very small interpolation step caused excessive vertex generation: %d vertices", vertexCount)
	}

	// Verify the interpolation step was actually set (may be very small)
	actualStep := bspline.InterpolationStep()
	if actualStep > 1e-3 {
		t.Errorf("Expected small interpolation step, got %f", actualStep)
	}

	// Test with coincident control points
	vertices = []CurveVertex{
		{X: 50, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo}, // Same point
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo}, // Same point again
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(vertices)
	bspline = NewConvBSpline(source)

	bspline.Rewind(0)

	// Should handle degenerate case gracefully (shouldn't crash)
	for {
		_, _, cmd := bspline.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
	}
}

// Benchmark tests
func BenchmarkConvBSpline_SimpleSpline(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		bspline.Rewind(0)
		for {
			_, _, cmd := bspline.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvBSpline_ComplexSpline(b *testing.B) {
	// Create complex B-spline with many control points
	vertices := make([]CurveVertex, 22)
	vertices[0] = CurveVertex{X: 0, Y: 50, Cmd: basics.PathCmdMoveTo}

	for i := 1; i < 21; i++ {
		x := float64(i * 5)
		y := 50 + 30*math.Sin(float64(i)*0.5)
		vertices[i] = CurveVertex{X: x, Y: y, Cmd: basics.PathCmdLineTo}
	}
	vertices[21] = CurveVertex{X: 0, Y: 0, Cmd: basics.PathCmdStop}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)
	bspline.SetInterpolationStep(0.05) // Fine interpolation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		bspline.Rewind(0)
		for {
			_, _, cmd := bspline.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvBSpline_FineInterpolation(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 33, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 66, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	bspline := NewConvBSpline(source)
	bspline.SetInterpolationStep(0.01) // Very fine interpolation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		bspline.Rewind(0)
		for {
			_, _, cmd := bspline.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
