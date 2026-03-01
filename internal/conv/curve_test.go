package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/curves"
)

// Test vertex source that can generate curve commands
type CurveVertexSource struct {
	vertices []CurveVertex
	index    int
}

type CurveVertex struct {
	X, Y float64
	Cmd  basics.PathCommand
}

func NewCurveVertexSource(vertices []CurveVertex) *CurveVertexSource {
	return &CurveVertexSource{vertices: vertices, index: 0}
}

func (c *CurveVertexSource) Rewind(pathID uint) {
	c.index = 0
}

func (c *CurveVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if c.index >= len(c.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := c.vertices[c.index]
	c.index++
	return v.X, v.Y, v.Cmd
}

func TestConvCurve_BasicLinearPath(t *testing.T) {
	// Test that linear paths pass through unchanged
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 20, Y: 20, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	curve.Rewind(0)

	// First vertex
	x, y, cmd := curve.Vertex()
	if x != 0 || y != 0 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected (0,0,MoveTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Second vertex
	x, y, cmd = curve.Vertex()
	if x != 10 || y != 10 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (10,10,LineTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Third vertex
	x, y, cmd = curve.Vertex()
	if x != 20 || y != 20 || cmd != basics.PathCmdLineTo {
		t.Errorf("Expected (20,20,LineTo), got (%f,%f,%v)", x, y, cmd)
	}

	// Stop
	x, y, cmd = curve.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop, got %v", cmd)
	}
}

func TestConvCurve_QuadraticBezier(t *testing.T) {
	// Test curve3 (quadratic Bezier) conversion
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve3}, // Control point
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},  // End point (will be consumed by curve)
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	curve.Rewind(0)

	// First vertex should be MoveTo
	x, y, cmd := curve.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}

	// Next vertex should be first approximated point from the curve
	x, y, cmd = curve.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for curve approximation, got %v", cmd)
	}

	// Collect all curve approximation points
	curvePoints := []CurveVertex{{X: x, Y: y, Cmd: cmd}}
	for {
		x, y, cmd = curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		curvePoints = append(curvePoints, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have multiple points approximating the curve
	if len(curvePoints) < 2 {
		t.Errorf("Expected multiple curve approximation points, got %d", len(curvePoints))
	}

	// All approximation points should be LineTo commands
	for i, pt := range curvePoints {
		if pt.Cmd != basics.PathCmdLineTo {
			t.Errorf("Curve approximation point %d should be LineTo, got %v", i, pt.Cmd)
		}
	}
}

func TestConvCurve_CubicBezier(t *testing.T) {
	// Test curve4 (cubic Bezier) conversion
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 100, Cmd: basics.PathCmdCurve4}, // First control point
		{X: 70, Y: 100, Cmd: basics.PathCmdLineTo}, // Second control point (will be consumed)
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},  // End point (will be consumed)
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	curve.Rewind(0)

	// First vertex should be MoveTo
	x, y, cmd := curve.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}

	// Collect all curve approximation points
	var curvePoints []CurveVertex
	for {
		x, y, cmd = curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		curvePoints = append(curvePoints, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have multiple points approximating the cubic curve
	if len(curvePoints) < 2 {
		t.Errorf("Expected multiple cubic curve approximation points, got %d", len(curvePoints))
	}

	// All approximation points should be LineTo commands
	for i, pt := range curvePoints {
		if pt.Cmd != basics.PathCmdLineTo {
			t.Errorf("Cubic curve approximation point %d should be LineTo, got %v", i, pt.Cmd)
		}
	}
}

func TestConvCurve_MixedPath(t *testing.T) {
	// Test path with both linear segments and curves
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 25, Cmd: basics.PathCmdLineTo},  // Linear segment
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve3}, // Start quadratic curve
		{X: 75, Y: 25, Cmd: basics.PathCmdLineTo},  // End quadratic curve
		{X: 100, Y: 25, Cmd: basics.PathCmdLineTo}, // Another linear segment
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	curve.Rewind(0)

	var allVertices []CurveVertex
	for {
		x, y, cmd := curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		allVertices = append(allVertices, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Should have MoveTo as first command
	if len(allVertices) == 0 || allVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Error("Path should start with MoveTo")
	}

	// Should have multiple vertices (linear + curve approximation)
	if len(allVertices) < 4 {
		t.Errorf("Mixed path should have multiple vertices, got %d", len(allVertices))
	}

	// Find linear segment (should be unchanged)
	found := false
	for _, v := range allVertices {
		if v.X == 25 && v.Y == 25 && v.Cmd == basics.PathCmdLineTo {
			found = true
			break
		}
	}
	if !found {
		t.Error("Linear segment should be preserved unchanged")
	}
}

func TestConvCurve_ApproximationMethods(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve3},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	// Test different approximation methods
	methods := []curves.CurveApproximationMethod{curves.CurveInc, curves.CurveDiv}

	for _, method := range methods {
		curve.SetApproximationMethod(method)
		if curve.ApproximationMethod() != method {
			t.Errorf("Expected approximation method %v, got %v", method, curve.ApproximationMethod())
		}

		// Both methods should produce some approximation
		source.index = 0 // Reset source
		curve.Rewind(0)

		vertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		if vertexCount == 0 {
			t.Errorf("Approximation method %v should produce vertices", method)
		}
	}
}

func TestConvCurve_ApproximationScale(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve3},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	// Test different scales
	scales := []float64{0.1, 1.0, 10.0}

	for _, scale := range scales {
		curve.SetApproximationScale(scale)
		if math.Abs(curve.ApproximationScale()-scale) > 1e-10 {
			t.Errorf("Expected approximation scale %f, got %f", scale, curve.ApproximationScale())
		}

		source.index = 0
		curve.Rewind(0)

		// Higher scales should generally produce more vertices
		vertexCount := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		if vertexCount == 0 {
			t.Errorf("Scale %f should produce vertices", scale)
		}
	}
}

func TestConvCurve_ApproximationScaleAffectsSubdivision(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 120, Cmd: basics.PathCmdCurve4},
		{X: 75, Y: 120, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	countVertices := func(scale float64) int {
		source := NewCurveVertexSource(vertices)
		curve := NewConvCurve(source)
		curve.SetApproximationScale(scale)
		curve.Rewind(0)

		count := 0
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			count++
		}
		return count
	}

	coarse := countVertices(0.1)
	fine := countVertices(10.0)

	if coarse < 2 {
		t.Fatalf("expected coarse curve approximation to emit multiple vertices, got %d", coarse)
	}
	if fine <= coarse {
		t.Fatalf("expected finer approximation scale to emit more vertices, got coarse=%d fine=%d", coarse, fine)
	}
}

func TestConvCurve_AngleTolerance(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve4},
		{X: 75, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	// Test angle tolerance
	tolerance := 0.1
	curve.SetAngleTolerance(tolerance)

	if math.Abs(curve.AngleTolerance()-tolerance) > 1e-10 {
		t.Errorf("Expected angle tolerance %f, got %f", tolerance, curve.AngleTolerance())
	}
}

func TestConvCurve_CuspLimit(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 100, Cmd: basics.PathCmdCurve4},
		{X: 75, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	// Test cusp limit
	limit := 0.5
	curve.SetCuspLimit(limit)

	if math.Abs(curve.CuspLimit()-limit) > 1e-10 {
		t.Errorf("Expected cusp limit %f, got %f", limit, curve.CuspLimit())
	}
}

func TestConvCurve_Attach(t *testing.T) {
	vertices1 := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	vertices2 := []CurveVertex{
		{X: 5, Y: 5, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdCurve3},
		{X: 50, Y: 5, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source1 := NewCurveVertexSource(vertices1)
	source2 := NewCurveVertexSource(vertices2)
	curve := NewConvCurve(source1)

	// Process first source
	curve.Rewind(0)
	firstVertex := make([]CurveVertex, 0)
	for {
		x, y, cmd := curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstVertex = append(firstVertex, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Attach second source
	curve.Attach(source2)
	curve.Rewind(0)
	secondVertex := make([]CurveVertex, 0)
	for {
		x, y, cmd := curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondVertex = append(secondVertex, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Second source should have more vertices (due to curve)
	if len(secondVertex) <= len(firstVertex) {
		t.Error("Second source with curve should produce more vertices than first linear source")
	}

	// First vertex should match the MoveTo from second source
	if len(secondVertex) > 0 && (secondVertex[0].X != 5 || secondVertex[0].Y != 5) {
		t.Errorf("First vertex should be from new source (5,5), got (%f,%f)",
			secondVertex[0].X, secondVertex[0].Y)
	}
}

func TestConvCurve_EmptyPath(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	curve.Rewind(0)
	x, y, cmd := curve.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Empty path should return Stop, got %v at (%f,%f)", cmd, x, y)
	}
}

func TestConvCurve_MultipleRewinds(t *testing.T) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 50, Cmd: basics.PathCmdCurve3},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	// First iteration
	curve.Rewind(0)
	var firstIteration []CurveVertex
	for {
		x, y, cmd := curve.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstIteration = append(firstIteration, CurveVertex{X: x, Y: y, Cmd: cmd})
	}

	// Second iteration should produce same results
	source.index = 0 // Reset source manually
	curve.Rewind(0)
	var secondIteration []CurveVertex
	for {
		x, y, cmd := curve.Vertex()
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

// Benchmark tests
func BenchmarkConvCurve_LinearPath(b *testing.B) {
	vertices := make([]CurveVertex, 101)
	vertices[0] = CurveVertex{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo}
	for i := 1; i < 100; i++ {
		vertices[i] = CurveVertex{X: float64(i), Y: float64(i), Cmd: basics.PathCmdLineTo}
	}
	vertices[100] = CurveVertex{X: 0, Y: 0, Cmd: basics.PathCmdStop}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		curve.Rewind(0)
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvCurve_QuadraticCurves(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 100, Cmd: basics.PathCmdCurve3},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 75, Y: 100, Cmd: basics.PathCmdCurve3},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		curve.Rewind(0)
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvCurve_CubicCurves(b *testing.B) {
	vertices := []CurveVertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 25, Y: 100, Cmd: basics.PathCmdCurve4},
		{X: 75, Y: 100, Cmd: basics.PathCmdLineTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewCurveVertexSource(vertices)
	curve := NewConvCurve(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0
		curve.Rewind(0)
		for {
			_, _, cmd := curve.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
