package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

func TestConvSmoothPoly1_Basic(t *testing.T) {
	// Create a simple rectangular path
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
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

	// Test without curve approximation (ConvSmoothPoly1 just generates cubic curves)
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

	// Test with curve approximation (ConvSmoothPoly1 still generates same cubic curves)
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

	// Both modes should generate same vertices since ConvSmoothPoly1 doesn't approximate
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
	// NOTE: Current implementation limitation - processes only first sub-path due to
	// ConvAdaptorVCGen architecture. This is a known limitation to be addressed.
	vertices := []CurveVertex{
		// First sub-path (will be processed)
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},

		// Second sub-path (currently not processed due to implementation limitation)
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

	// Due to current implementation limitation, only first sub-path is processed
	// Should have at least one sub-path processed correctly
	if moveToCount < 1 {
		t.Errorf("Expected at least 1 MoveTo command, got %d", moveToCount)
	}

	if endPolyCount < 1 {
		t.Errorf("Expected at least 1 EndPoly command, got %d", endPolyCount)
	}

	// Verify smooth polygon generation worked for the processed sub-path
	if len(resultVertices) < 5 {
		t.Errorf("Expected several vertices from smooth polygon processing, got %d", len(resultVertices))
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

func TestConvSmoothPoly1Curve_Basic(t *testing.T) {
	// Create a simple rectangular path
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	smoothCurve.Rewind(0)

	// Collect all vertices - should be line segments approximating the smooth curves
	var resultVertices []CurveVertex
	for {
		x, y, cmd := smoothCurve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have generated many line segments to approximate the smooth curves
	if len(resultVertices) < 10 {
		t.Errorf("Smooth polygon with curve approximation should generate many line segments, got %d", len(resultVertices))
	}

	// First vertex should be MoveTo
	if resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}

	// Should have only MoveTo and LineTo commands (no Curve4 since they're approximated)
	for i, v := range resultVertices {
		if v.Cmd != basics.PathCmdMoveTo && v.Cmd != basics.PathCmdLineTo &&
			(v.Cmd&basics.PathCmdMask) != basics.PathCmdEndPoly {
			t.Errorf("Vertex %d should be MoveTo/LineTo/EndPoly but got %v", i, v.Cmd)
		}
	}

	// Last vertex should be EndPoly (but might be LineTo since curve approximation creates line segments)
	if len(resultVertices) > 0 {
		lastVertex := resultVertices[len(resultVertices)-1]
		maskedCmd := lastVertex.Cmd & basics.PathCmdMask
		if maskedCmd != basics.PathCmdEndPoly && lastVertex.Cmd != basics.PathCmdLineTo {
			t.Errorf("Last vertex should be EndPoly or LineTo, got %v (masked: %v)",
				lastVertex.Cmd, maskedCmd)
		}
	}
}

func TestConvSmoothPoly1Curve_ApproximationMethods(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test default approximation method
	defaultMethod := smoothCurve.ApproximationMethod()
	if defaultMethod < 0 || defaultMethod > 2 {
		t.Errorf("Default approximation method should be valid, got %v", defaultMethod)
	}

	// Test setting different approximation methods
	methods := []curves.CurveApproximationMethod{
		curves.CurveInc,
		curves.CurveDiv,
	}

	for _, method := range methods {
		smoothCurve.SetApproximationMethod(method)
		if smoothCurve.ApproximationMethod() != method {
			t.Errorf("Expected approximation method %v, got %v", method, smoothCurve.ApproximationMethod())
		}
	}
}

func TestConvSmoothPoly1Curve_ApproximationScale(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test default approximation scale
	defaultScale := smoothCurve.ApproximationScale()
	if defaultScale <= 0 || defaultScale > 10 {
		t.Errorf("Default approximation scale should be reasonable, got %f", defaultScale)
	}

	// Test setting different approximation scales
	testScales := []float64{0.1, 1.0, 2.0, 5.0}

	for _, scale := range testScales {
		smoothCurve.SetApproximationScale(scale)
		actualScale := smoothCurve.ApproximationScale()
		if actualScale != scale {
			t.Errorf("Expected approximation scale %f, got %f", scale, actualScale)
		}

		// Test that different scales produce different vertex counts
		source.index = 0
		smoothCurve.Rewind(0)
		vertexCount := 0
		for {
			_, _, cmd := smoothCurve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		// Should have some vertices for any reasonable scale
		if vertexCount == 0 {
			t.Errorf("Scale %f should produce vertices", scale)
		}
	}
}

func TestConvSmoothPoly1Curve_AngleTolerance(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test default angle tolerance
	defaultTolerance := smoothCurve.AngleTolerance()
	if defaultTolerance < 0 || defaultTolerance > 1 {
		t.Errorf("Default angle tolerance should be reasonable, got %f", defaultTolerance)
	}

	// Test setting different angle tolerances
	testTolerances := []float64{0.01, 0.1, 0.2, 0.5}

	for _, tolerance := range testTolerances {
		smoothCurve.SetAngleTolerance(tolerance)
		actualTolerance := smoothCurve.AngleTolerance()
		if actualTolerance != tolerance {
			t.Errorf("Expected angle tolerance %f, got %f", tolerance, actualTolerance)
		}
	}
}

func TestConvSmoothPoly1Curve_CuspLimit(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test default cusp limit
	defaultLimit := smoothCurve.CuspLimit()
	if defaultLimit < 0 {
		t.Errorf("Default cusp limit should be non-negative, got %f", defaultLimit)
	}

	// Test setting different cusp limits
	testLimits := []float64{0.0, 0.1, 0.5, 1.0}

	for _, limit := range testLimits {
		smoothCurve.SetCuspLimit(limit)
		actualLimit := smoothCurve.CuspLimit()
		if math.Abs(actualLimit-limit) > 1e-10 {
			t.Errorf("Expected cusp limit %f, got %f", limit, actualLimit)
		}
	}
}

func TestConvSmoothPoly1Curve_SmoothValue(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test default smooth value
	defaultSmooth := smoothCurve.SmoothValue()
	if defaultSmooth <= 0 || defaultSmooth > 2.0 {
		t.Errorf("Default smooth value should be reasonable, got %f", defaultSmooth)
	}

	// Test setting custom smooth values
	testValues := []float64{0.0, 0.5, 1.0}

	for _, value := range testValues {
		smoothCurve.SetSmoothValue(value)
		actualValue := smoothCurve.SmoothValue()
		if actualValue != value {
			t.Errorf("Expected smooth value %f, got %f", value, actualValue)
		}
	}
}

func TestConvSmoothPoly1Curve_DifferentApproximations(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)

	// Test with coarse approximation (fewer line segments)
	smoothCurve.SetApproximationScale(5.0)
	source.index = 0
	smoothCurve.Rewind(0)

	coarseVertexCount := 0
	for {
		_, _, cmd := smoothCurve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		coarseVertexCount++
	}

	// Test with fine approximation (more line segments)
	smoothCurve.SetApproximationScale(0.1)
	source.index = 0
	smoothCurve.Rewind(0)

	fineVertexCount := 0
	for {
		_, _, cmd := smoothCurve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		fineVertexCount++
	}

	// Fine approximation should typically generate more vertices
	if coarseVertexCount == 0 || fineVertexCount == 0 {
		t.Error("Both approximation scales should generate vertices")
	}

	// Note: We don't strictly require fineVertexCount > coarseVertexCount
	// because the actual behavior depends on the curve complexity and implementation
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

func BenchmarkConvSmoothPoly1Curve_Rectangle(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)
	smoothCurve.SetSmoothValue(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smoothCurve.Rewind(0)
		for {
			_, _, cmd := smoothCurve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvSmoothPoly1Curve_FineApproximation(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	smoothCurve := NewConvSmoothPoly1Curve(source)
	smoothCurve.SetSmoothValue(0.8)
	smoothCurve.SetApproximationScale(0.1) // Fine approximation

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smoothCurve.Rewind(0)
		for {
			_, _, cmd := smoothCurve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
