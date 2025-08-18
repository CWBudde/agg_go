package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestConvSmoothPoly1_Basic(t *testing.T) {
	// Create a simple rectangular path
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	// Collect all vertices
	var resultVertices []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have generated smooth vertices
	if len(resultVertices) < 4 {
		t.Errorf("Smooth polygon should generate multiple vertices, got %d", len(resultVertices))
	}

	// First vertex should be MoveTo
	if resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}

	// Should have cubic Bezier curves (Curve4 commands) for smoothed corners
	curve4Count := 0
	for _, v := range resultVertices {
		if v.Cmd == basics.PathCmdCurve4 {
			curve4Count++
		}
	}

	// Should have multiple Curve4 commands for corner smoothing
	if curve4Count < 8 { // At least 2 control points + end point for multiple corners
		t.Errorf("Expected multiple Curve4 commands for corner smoothing, got %d", curve4Count)
	}

	// Last vertex should be EndPoly
	if len(resultVertices) > 0 {
		lastVertex := resultVertices[len(resultVertices)-1]
		if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.Cmd)
		}
	}
}

func TestConvSmoothPoly1_SmoothValue(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	// Test default smooth value
	defaultSmooth := smooth.SmoothValue()
	if defaultSmooth <= 0 || defaultSmooth > 2.0 {
		t.Errorf("Default smooth value should be reasonable, got %f", defaultSmooth)
	}

	// Test setting custom smooth values
	testValues := []float64{0.0, 0.5, 1.0}

	for _, value := range testValues {
		smooth.SetSmoothValue(value)
		if math.Abs(smooth.SmoothValue()-value) > 1e-10 {
			t.Errorf("Expected smooth value %f, got %f", value, smooth.SmoothValue())
		}
	}

	// Test with different smooth values on same polygon
	// Test minimum smoothing (0.0)
	smooth.SetSmoothValue(0.0)
	source.index = 0 // Reset source
	smooth.Rewind(0)

	zeroSmoothVertices := 0
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		zeroSmoothVertices++
	}

	// Test maximum smoothing (1.0)
	smooth.SetSmoothValue(1.0)
	source.index = 0
	smooth.Rewind(0)

	maxSmoothVertices := 0
	for {
		_, _, cmd := smooth.Vertex()
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

func TestConvSmoothPoly1_CurveApproximation(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	// Test without curve approximation
	smooth.SetCurveApproximation(false)
	if smooth.CurveApproximation() != false {
		t.Error("Expected curve approximation to be false")
	}

	source.index = 0
	smooth.Rewind(0)

	noCurveApproxVertices := 0
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		noCurveApproxVertices++
	}

	// Test with curve approximation
	smooth.SetCurveApproximation(true)
	if smooth.CurveApproximation() != true {
		t.Error("Expected curve approximation to be true")
	}

	source.index = 0
	smooth.Rewind(0)

	curveApproxVertices := 0
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		curveApproxVertices++
	}

	// Both modes should generate vertices, but possibly different counts
	if noCurveApproxVertices == 0 || curveApproxVertices == 0 {
		t.Error("Both approximation modes should generate vertices")
	}
}

func TestConvSmoothPoly1_Triangle(t *testing.T) {
	// Test with a triangle
	vertices := []CurveVertex{
		{X: 50, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)
	smooth.SetSmoothValue(0.5)

	smooth.Rewind(0)

	var vertices_result []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices_result = append(vertices_result, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have smoothed the 3 corners of the triangle
	if len(vertices_result) == 0 {
		t.Error("Triangle smoothing should generate vertices")
	}

	// First vertex should be MoveTo
	if vertices_result[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", vertices_result[0].Cmd)
	}

	// Should have Curve4 commands for corner smoothing
	hasCurve4 := false
	for _, v := range vertices_result {
		if v.Cmd == basics.PathCmdCurve4 {
			hasCurve4 = true
			break
		}
	}

	if !hasCurve4 {
		t.Error("Triangle smoothing should generate Curve4 commands")
	}
}

func TestConvSmoothPoly1_ClosedPolygon(t *testing.T) {
	// Create a closed pentagon
	vertices := []CurveVertex{
		{X: 50, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 25, Cmd: basics.PathCmdLineTo},
		{X: 80, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 20, Y: 75, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 25, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathFlagClose}, // Close the polygon
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	var vertices_result []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices_result = append(vertices_result, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	if len(vertices_result) == 0 {
		t.Error("Closed pentagon should generate smooth vertices")
	}

	// Should end with EndPoly with close flag
	if len(vertices_result) > 0 {
		lastVertex := vertices_result[len(vertices_result)-1]
		if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Closed polygon should end with EndPoly, got %v", lastVertex.Cmd)
		}

		if (lastVertex.Cmd & basics.PathFlagClose) == 0 {
			t.Error("Closed polygon should have close flag")
		}
	}
}

func TestConvSmoothPoly1_OpenPolygon(t *testing.T) {
	// Create an open polygon (no close command)
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	var vertices_result []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices_result = append(vertices_result, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	if len(vertices_result) == 0 {
		t.Error("Open polygon should generate smooth vertices")
	}

	// Should end with EndPoly without close flag
	if len(vertices_result) > 0 {
		lastVertex := vertices_result[len(vertices_result)-1]
		if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Open polygon should end with EndPoly, got %v", lastVertex.Cmd)
		}

		if (lastVertex.Cmd & basics.PathFlagClose) != 0 {
			t.Error("Open polygon should not have close flag")
		}
	}
}

func TestConvSmoothPoly1_InsufficientVertices(t *testing.T) {
	// Test with single vertex
	vertices := []CurveVertex{
		{X: 50, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	x, y, cmd := smooth.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single vertex should result in Stop, got %v at (%f,%f)", cmd, x, y)
	}

	// Test with two vertices (line segment)
	vertices = []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(vertices)
	smooth = NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	var twoVertexCommands []basics.PathCommand
	for {
		_, _, cmd = smooth.Vertex()
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

func TestConvSmoothPoly1_Attach(t *testing.T) {
	vertices1 := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	vertices2 := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source1 := NewCurveVertexSource(vertices1)
	source2 := NewCurveVertexSource(vertices2)
	smooth := NewConvSmoothPoly1(source1)

	// Process first source
	smooth.Rewind(0)
	firstVertexCount := 0
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstVertexCount++
	}

	// Attach second source
	smooth.Attach(source2)
	smooth.Rewind(0)
	secondVertexCount := 0
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondVertexCount++
	}

	// Second source should have more vertices (more corners to smooth)
	if secondVertexCount <= firstVertexCount {
		t.Error("Second source with more corners should produce more vertices")
	}
}

func TestConvSmoothPoly1_EmptyPath(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)
	x, y, cmd := smooth.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestConvSmoothPoly1_MultipleRewinds(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	// First iteration
	smooth.Rewind(0)
	var firstIteration []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstIteration = append(firstIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Second iteration should produce same results
	source.index = 0 // Reset source
	smooth.Rewind(0)
	var secondIteration []CurveVertex
	for {
		x, y, cmd := smooth.Vertex()
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

func TestConvSmoothPoly1_ComplexPath(t *testing.T) {
	// Test with a more complex path having multiple sub-paths
	vertices := []CurveVertex{
		// First sub-path
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},

		// Second sub-path
		{X: 100, Y: 100, Cmd: basics.PathCmdMoveTo},
		{X: 125, Y: 150, Cmd: basics.PathCmdLineTo},
		{X: 150, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},

		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	var resultVertices []CurveVertex
	var moveToCount int
	var endPolyCount int

	for {
		x, y, cmd := smooth.Vertex()
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

	// Should have processed multiple sub-paths
	if moveToCount < 2 {
		t.Errorf("Expected at least 2 MoveTo commands for multiple sub-paths, got %d", moveToCount)
	}

	if endPolyCount < 2 {
		t.Errorf("Expected at least 2 EndPoly commands for multiple sub-paths, got %d", endPolyCount)
	}
}

func TestConvSmoothPoly1_EdgeCases(t *testing.T) {
	// Test with coincident vertices
	vertices := []CurveVertex{
		{X: 50, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo}, // Same point
		{X: 100, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	// Should handle degenerate case gracefully
	hasVertices := false
	for {
		_, _, cmd := smooth.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		hasVertices = true
	}

	if !hasVertices {
		t.Error("Coincident vertices should still allow some smoothing")
	}

	// Test with very small polygon
	vertices = []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 1, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 1, Y: 1, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 1, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(vertices)
	smooth = NewConvSmoothPoly1(source)

	smooth.Rewind(0)

	// Should handle very small polygons
	smallPolyVertices := 0
	for {
		_, _, cmd := smooth.Vertex()
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
func BenchmarkConvSmoothPoly1_Rectangle(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)
	smooth.SetSmoothValue(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smooth.Rewind(0)
		for {
			_, _, cmd := smooth.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvSmoothPoly1_ComplexPolygon(b *testing.B) {
	// Setup complex polygon (octagon)
	vertices := []CurveVertex{
		{X: 30, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 70, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 30, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 70, Cmd: basics.PathCmdLineTo},
		{X: 70, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 30, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 70, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 30, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)
	smooth.SetSmoothValue(0.7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smooth.Rewind(0)
		for {
			_, _, cmd := smooth.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvSmoothPoly1_WithCurveApproximation(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)
	smooth.SetSmoothValue(0.8)
	smooth.SetCurveApproximation(true) // Enable curve approximation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smooth.Rewind(0)
		for {
			_, _, cmd := smooth.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvSmoothPoly1_MaxSmoothing(b *testing.B) {
	// Setup triangle with sharp corners
	vertices := []CurveVertex{
		{X: 50, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smooth := NewConvSmoothPoly1(source)
	smooth.SetSmoothValue(1.0) // Maximum smoothing

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smooth.Rewind(0)
		for {
			_, _, cmd := smooth.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
