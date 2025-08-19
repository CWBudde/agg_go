package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// Integration test: ConvTransform with ConvStroke
func TestConvTransform_IntegrationWithStroke(t *testing.T) {
	// Create a simple line path
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(2, 2).Translate(5, 5)

	// Apply transformation first, then stroke
	conv1 := NewConvTransform(source, transformer)
	conv2 := NewConvStroke(conv1)
	conv2.SetWidth(2.0)

	conv2.Rewind(0)

	// Should get stroked geometry of the transformed line
	vertexCount := 0
	for {
		_, _, cmd := conv2.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	// Stroked line should produce multiple vertices (outline)
	if vertexCount < 4 {
		t.Errorf("Integration with stroke: expected at least 4 vertices for stroked line, got %d", vertexCount)
	}
}

// Integration test: ConvTransform with ConvClosePolygon
func TestConvTransform_IntegrationWithClosePolygon(t *testing.T) {
	// Create an open polygon
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(100, 100)

	// Apply transformation first, then close polygon
	conv1 := NewConvTransform(source, transformer)
	conv2 := NewConvClosePolygon(conv1)

	conv2.Rewind(0)

	vertices_out := []Vertex{}
	for {
		x, y, cmd := conv2.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices_out = append(vertices_out, Vertex{x, y, cmd})
	}

	// Should have the transformed vertices with close command
	if len(vertices_out) < 5 { // 4 original + close
		t.Errorf("Integration with close polygon: expected at least 5 vertices, got %d", len(vertices_out))
	}

	// Check that first vertex is properly transformed
	if vertices_out[0].X != 100 || vertices_out[0].Y != 100 {
		t.Errorf("Integration: first vertex should be (100,100), got (%.1f,%.1f)",
			vertices_out[0].X, vertices_out[0].Y)
	}

	// Last vertex should be EndPoly with Close flag
	lastVertex := vertices_out[len(vertices_out)-1]
	if lastVertex.Cmd != (basics.PathCmdEndPoly | basics.PathFlagClose) {
		t.Errorf("Integration: expected close command, got %v", lastVertex.Cmd)
	}
}

// Integration test: Multiple ConvTransform in sequence
func TestConvTransform_IntegrationChain(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)

	// Chain multiple transformations
	trans1 := transform.NewTransAffine().ScaleXY(2, 2)       // Scale by 2
	trans2 := transform.NewTransAffine().Translate(10, 20)   // Translate
	trans3 := transform.NewTransAffine().Rotate(math.Pi / 2) // Rotate 90 degrees

	conv1 := NewConvTransform(source, trans1)
	conv2 := NewConvTransform(conv1, trans2)
	conv3 := NewConvTransform(conv2, trans3)

	conv3.Rewind(0)

	// First vertex: (1,1) -> scale -> (2,2) -> translate -> (12,22) -> rotate -> (-22,12)
	x, y, cmd := conv3.Vertex()
	expectedX, expectedY := -22.0, 12.0
	tolerance := 1e-10

	if math.Abs(x-expectedX) > tolerance || math.Abs(y-expectedY) > tolerance || cmd != basics.PathCmdMoveTo {
		t.Errorf("Transform chain: expected (%v, %v, MoveTo), got (%v, %v, %v)",
			expectedX, expectedY, x, y, cmd)
	}

	// Second vertex: (2,2) -> scale -> (4,4) -> translate -> (14,24) -> rotate -> (-24,14)
	x, y, cmd = conv3.Vertex()
	expectedX, expectedY = -24.0, 14.0

	if math.Abs(x-expectedX) > tolerance || math.Abs(y-expectedY) > tolerance || cmd != basics.PathCmdLineTo {
		t.Errorf("Transform chain line: expected (%v, %v, LineTo), got (%v, %v, %v)",
			expectedX, expectedY, x, y, cmd)
	}
}

// Integration test: ConvTransform preserving source interface
func TestConvTransform_IntegrationAsVertexSource(t *testing.T) {
	vertices := []Vertex{
		{1, 1, basics.PathCmdMoveTo},
		{2, 2, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().Translate(5, 5)
	conv := NewConvTransform(source, transformer)

	// Use the transform converter as a vertex source for another converter
	strokeConv := NewConvStroke(conv)
	strokeConv.SetWidth(1.0)

	// Should be able to iterate through vertices
	strokeConv.Rewind(0)
	vertexCount := 0
	for {
		_, _, cmd := strokeConv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++
	}

	// Should produce stroked output
	if vertexCount == 0 {
		t.Error("Integration as vertex source: should produce vertices")
	}
}

// Performance test with many transformations
func TestConvTransform_PerformanceStress(t *testing.T) {
	// Create a path with many vertices
	vertices := make([]Vertex, 1001)
	vertices[0] = Vertex{0, 0, basics.PathCmdMoveTo}
	for i := 1; i < 1000; i++ {
		vertices[i] = Vertex{float64(i), float64(i), basics.PathCmdLineTo}
	}
	vertices[1000] = Vertex{0, 0, basics.PathCmdStop}

	source := NewMockVertexSource(vertices)
	transformer := transform.NewTransAffine().ScaleXY(1.1, 1.1).Translate(0.5, 0.5)
	conv := NewConvTransform(source, transformer)

	// Time the transformation
	conv.Rewind(0)

	transformedCount := 0
	for {
		_, _, cmd := conv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		transformedCount++
	}

	if transformedCount != 1000 {
		t.Errorf("Performance test: expected 1000 transformed vertices, got %d", transformedCount)
	}
}
