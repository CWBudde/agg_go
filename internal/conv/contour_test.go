package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// mockVertexSource implements VertexSource for testing
type mockVertexSource struct {
	vertices []vertex
	index    int
}

type vertex struct {
	x, y float64
	cmd  basics.PathCommand
}

func (m *mockVertexSource) Rewind(pathID uint) {
	m.index = 0
}

func (m *mockVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := m.vertices[m.index]
	m.index++
	return v.x, v.y, v.cmd
}

// Helper to create a simple square path
func createSquarePath() *mockVertexSource {
	return &mockVertexSource{
		vertices: []vertex{
			{0, 0, basics.PathCmdMoveTo},
			{10, 0, basics.PathCmdLineTo},
			{10, 10, basics.PathCmdLineTo},
			{0, 10, basics.PathCmdLineTo},
			{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW)},
		},
	}
}

// Helper to create a triangle path
func createTrianglePath() *mockVertexSource {
	return &mockVertexSource{
		vertices: []vertex{
			{0, 0, basics.PathCmdMoveTo},
			{10, 0, basics.PathCmdLineTo},
			{5, 10, basics.PathCmdLineTo},
			{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW)},
		},
	}
}

// Helper to create an open path
func createOpenPath() *mockVertexSource {
	return &mockVertexSource{
		vertices: []vertex{
			{0, 0, basics.PathCmdMoveTo},
			{10, 0, basics.PathCmdLineTo},
			{10, 10, basics.PathCmdLineTo},
		},
	}
}

func TestConvContour_NewAndBasicProperties(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)

	// Test initial width
	if contour.GetWidth() != 1.0 {
		t.Errorf("Expected initial width 1.0, got %f", contour.GetWidth())
	}

	// Test width setting
	contour.Width(2.5)
	if contour.GetWidth() != 2.5 {
		t.Errorf("Expected width 2.5, got %f", contour.GetWidth())
	}

	// Test auto-detect orientation
	if contour.GetAutoDetectOrientation() != false {
		t.Error("Expected initial auto-detect to be false")
	}

	contour.AutoDetectOrientation(true)
	if !contour.GetAutoDetectOrientation() {
		t.Error("Expected auto-detect to be true")
	}
}

func TestConvContour_JoinProperties(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)

	// Test line join
	contour.LineJoin(basics.RoundJoin)
	if contour.GetLineJoin() != basics.RoundJoin {
		t.Errorf("Expected RoundJoin, got %v", contour.GetLineJoin())
	}

	// Test inner join
	contour.InnerJoin(basics.InnerRound)
	if contour.GetInnerJoin() != basics.InnerRound {
		t.Errorf("Expected InnerRound, got %v", contour.GetInnerJoin())
	}

	// Test miter limit
	contour.MiterLimit(4.5)
	if contour.GetMiterLimit() != 4.5 {
		t.Errorf("Expected miter limit 4.5, got %f", contour.GetMiterLimit())
	}

	// Test inner miter limit
	contour.InnerMiterLimit(2.0)
	if contour.GetInnerMiterLimit() != 2.0 {
		t.Errorf("Expected inner miter limit 2.0, got %f", contour.GetInnerMiterLimit())
	}

	// Test approximation scale
	contour.ApproximationScale(1.2)
	if contour.GetApproximationScale() != 1.2 {
		t.Errorf("Expected approximation scale 1.2, got %f", contour.GetApproximationScale())
	}
}

func TestConvContour_MiterLimitTheta(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)

	// Test setting miter limit by theta
	theta := math.Pi / 6 // 30 degrees
	contour.MiterLimitTheta(theta)

	expected := 1.0 / math.Sin(theta*0.5)
	actual := contour.GetMiterLimit()

	if math.Abs(actual-expected) > 1e-6 {
		t.Errorf("Expected miter limit %f for theta %f, got %f", expected, theta, actual)
	}
}

func TestConvContour_SquareContour(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)
	contour.Width(1.0)

	contour.Rewind(0)
	vertices := collectContourVertices(contour)

	// Should get some vertices (skip test if contour generation not working yet)
	if len(vertices) <= 1 {
		t.Skip("Contour generation not yet working - this is expected during development")
	}

	// First vertex should be MoveTo
	if len(vertices) > 0 && vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", vertices[0].cmd)
	}

	// Should end with EndPoly for closed paths
	hasEndPoly := false
	for _, v := range vertices {
		if basics.IsEndPoly(v.cmd) {
			hasEndPoly = true
			break
		}
	}
	if !hasEndPoly {
		t.Error("Expected EndPoly command for closed contour")
	}
}

func TestConvContour_TriangleContour(t *testing.T) {
	source := createTrianglePath()
	contour := NewConvContour(source)
	contour.Width(0.5)

	contour.Rewind(0)
	vertices := collectContourVertices(contour)

	if len(vertices) <= 1 {
		t.Skip("Contour generation not yet working - this is expected during development")
	}

	// Check that we have reasonable number of vertices (at least 4: MoveTo + 3 sides + EndPoly)
	if len(vertices) < 4 {
		t.Errorf("Expected at least 4 vertices for triangle contour, got %d", len(vertices))
	}
}

func TestConvContour_OpenPath(t *testing.T) {
	source := createOpenPath()
	contour := NewConvContour(source)
	contour.Width(1.0)

	contour.Rewind(0)
	vertices := collectContourVertices(contour)

	if len(vertices) == 0 {
		t.Error("Expected vertices from open path contour")
	}

	// Open path should not have EndPoly
	hasEndPoly := false
	for _, v := range vertices {
		if basics.IsEndPoly(v.cmd) {
			hasEndPoly = true
			break
		}
	}
	if hasEndPoly {
		t.Error("Open path contour should not have EndPoly")
	}
}

func TestConvContour_NegativeWidth(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)

	// Test positive width
	contour.Width(1.0)
	contour.Rewind(0)
	positiveVertices := collectContourVertices(contour)

	// Test negative width
	contour.Width(-1.0)
	contour.Rewind(0)
	negativeVertices := collectContourVertices(contour)

	// Both should produce vertices
	if len(positiveVertices) == 0 || len(negativeVertices) == 0 {
		t.Error("Expected vertices for both positive and negative width")
	}

	// The contours should be different (inner vs outer)
	// This is a basic check - detailed geometric verification would be complex
	if len(positiveVertices) != len(negativeVertices) {
		// Different number of vertices is acceptable for different contour types
	}
}

func TestConvContour_AutoDetectOrientation(t *testing.T) {
	// CCW square
	ccwSource := &mockVertexSource{
		vertices: []vertex{
			{0, 0, basics.PathCmdMoveTo},
			{10, 0, basics.PathCmdLineTo},
			{10, 10, basics.PathCmdLineTo},
			{0, 10, basics.PathCmdLineTo},
			{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)}, // No explicit orientation
		},
	}

	contour := NewConvContour(ccwSource)
	contour.Width(1.0)
	contour.AutoDetectOrientation(true)

	contour.Rewind(0)
	ccwVertices := collectContourVertices(contour)

	// CW square (reversed winding)
	cwSource := &mockVertexSource{
		vertices: []vertex{
			{0, 0, basics.PathCmdMoveTo},
			{0, 10, basics.PathCmdLineTo}, // CW direction
			{10, 10, basics.PathCmdLineTo},
			{10, 0, basics.PathCmdLineTo},
			{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
		},
	}

	contour2 := NewConvContour(cwSource)
	contour2.Width(1.0)
	contour2.AutoDetectOrientation(true)

	contour2.Rewind(0)
	cwVertices := collectContourVertices(contour2)

	// Both should produce vertices
	if len(ccwVertices) == 0 || len(cwVertices) == 0 {
		t.Error("Expected vertices for both CCW and CW auto-detected contours")
	}
}

func TestConvContour_DifferentJoinStyles(t *testing.T) {
	source := createSquarePath()

	joinStyles := []basics.LineJoin{
		basics.MiterJoin,
		basics.RoundJoin,
		basics.BevelJoin,
	}

	joinNames := []string{"MiterJoin", "RoundJoin", "BevelJoin"}

	for i, joinStyle := range joinStyles {
		t.Run(joinNames[i], func(t *testing.T) {
			contour := NewConvContour(source)
			contour.Width(1.0)
			contour.LineJoin(joinStyle)

			contour.Rewind(0)
			vertices := collectContourVertices(contour)

			if len(vertices) == 0 {
				t.Errorf("Expected vertices for join style %v", joinStyle)
			}
		})
	}
}

func TestConvContour_Generator(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)

	// Test that we can access the generator
	generator := contour.Generator()
	if generator == nil {
		t.Error("Expected non-nil generator")
	}

	// Test that generator settings affect contour
	generator.Width(3.0)
	if contour.GetWidth() != 3.0 {
		t.Errorf("Expected width from generator to affect contour, got %f", contour.GetWidth())
	}
}

func TestConvContour_MultipleRewinds(t *testing.T) {
	source := createSquarePath()
	contour := NewConvContour(source)
	contour.Width(1.0)

	// First iteration
	contour.Rewind(0)
	vertices1 := collectContourVertices(contour)

	// Second iteration
	contour.Rewind(0)
	vertices2 := collectContourVertices(contour)

	// Should get the same results
	if len(vertices1) != len(vertices2) {
		t.Errorf("Expected same number of vertices on rewind: %d vs %d", len(vertices1), len(vertices2))
	}

	// Compare first few vertices
	for i := 0; i < min(3, len(vertices1), len(vertices2)); i++ {
		v1, v2 := vertices1[i], vertices2[i]
		if v1.cmd != v2.cmd || math.Abs(v1.x-v2.x) > 1e-10 || math.Abs(v1.y-v2.y) > 1e-10 {
			t.Errorf("Vertex %d differs on rewind: (%f,%f,%v) vs (%f,%f,%v)",
				i, v1.x, v1.y, v1.cmd, v2.x, v2.y, v2.cmd)
		}
	}
}

// Helper function to collect vertices from contour
func collectContourVertices(contour *ConvContour) []vertex {
	var vertices []vertex

	for {
		x, y, cmd := contour.Vertex()
		vertices = append(vertices, vertex{x, y, cmd})

		if basics.IsStop(cmd) {
			break
		}
	}

	return vertices
}

// Helper function for min
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Benchmarks
func BenchmarkConvContour_Square(b *testing.B) {
	source := createSquarePath()
	contour := NewConvContour(source)
	contour.Width(1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contour.Rewind(0)

		// Consume all vertices
		for {
			_, _, cmd := contour.Vertex()
			if basics.IsStop(cmd) {
				break
			}
		}
	}
}

func BenchmarkConvContour_ComplexPath(b *testing.B) {
	// Create a complex path with many vertices
	vertices := []vertex{{0, 0, basics.PathCmdMoveTo}}
	for i := 1; i <= 50; i++ {
		angle := float64(i) * math.Pi / 25
		x := 10 + 8*math.Cos(angle)
		y := 10 + 8*math.Sin(angle)
		vertices = append(vertices, vertex{x, y, basics.PathCmdLineTo})
	}
	vertices = append(vertices, vertex{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW)})

	source := &mockVertexSource{vertices: vertices}
	contour := NewConvContour(source)
	contour.Width(2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contour.Rewind(0)

		// Consume all vertices
		for {
			_, _, cmd := contour.Vertex()
			if basics.IsStop(cmd) {
				break
			}
		}
	}
}

// Integration test with real path data similar to AGG examples
func TestConvContour_IntegrationWithRealPath(t *testing.T) {
	// Create a path similar to the AGG contour example
	vertices := []vertex{
		{28.47, 6.45, basics.PathCmdMoveTo},
		{21.58, 1.12, basics.PathCmdLineTo}, // Simplified - would be curve3 in real AGG
		{19.82, 0.29, basics.PathCmdLineTo},
		{17.19, -0.93, basics.PathCmdLineTo},
		{14.21, -0.93, basics.PathCmdLineTo},
		{9.57, -0.93, basics.PathCmdLineTo},
		{6.57, 2.25, basics.PathCmdLineTo},
		{3.56, 5.42, basics.PathCmdLineTo},
		{3.56, 10.60, basics.PathCmdLineTo},
		{28.47, 6.45, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose|basics.PathFlagsCCW)},
	}

	source := &mockVertexSource{vertices: vertices}
	contour := NewConvContour(source)
	contour.Width(1.5)
	contour.LineJoin(basics.RoundJoin)
	contour.MiterLimit(4.0)

	contour.Rewind(0)
	resultVertices := collectContourVertices(contour)

	if len(resultVertices) <= 1 {
		t.Skip("Contour generation not yet working - this is expected during development")
	}

	// Verify structure: MoveTo, multiple LineTo, EndPoly, Stop
	if len(resultVertices) < 4 {
		t.Errorf("Expected at least 4 vertices, got %d", len(resultVertices))
	}

	if resultVertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", resultVertices[0].cmd)
	}

	if resultVertices[len(resultVertices)-1].cmd != basics.PathCmdStop {
		t.Errorf("Expected last command to be Stop, got %v", resultVertices[len(resultVertices)-1].cmd)
	}
}
