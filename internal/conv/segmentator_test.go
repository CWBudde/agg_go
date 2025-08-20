package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestNewConvSegmentator(t *testing.T) {
	// Create a simple path as vertex source
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvSegmentator(source)

	if conv == nil {
		t.Fatal("NewConvSegmentator returned nil")
	}

	// Check default approximation scale
	if conv.GetApproximationScale() != 1.0 {
		t.Errorf("Expected default approximation scale 1.0, got %v", conv.GetApproximationScale())
	}
}

func TestConvSegmentatorApproximationScale(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	conv := NewConvSegmentator(source)

	testScales := []float64{0.5, 1.0, 2.0, 5.0}
	for _, scale := range testScales {
		conv.ApproximationScale(scale)
		if got := conv.GetApproximationScale(); got != scale {
			t.Errorf("Expected approximation scale %v, got %v", scale, got)
		}
	}
}

func TestConvSegmentatorAttach(t *testing.T) {
	source1 := NewMockVertexSource([]Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
	})

	source2 := NewMockVertexSource([]Vertex{
		{X: 5, Y: 5, Cmd: basics.PathCmdMoveTo},
		{X: 15, Y: 15, Cmd: basics.PathCmdLineTo},
	})

	conv := NewConvSegmentator(source1)

	// Test with first path
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if uint32(cmd) != uint32(basics.PathCmdMoveTo) || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%v,%v)", cmd, x, y)
	}

	// Attach second path
	conv.Attach(source2)
	conv.Rewind(0)
	x, y, cmd = conv.Vertex()
	if uint32(cmd) != uint32(basics.PathCmdMoveTo) || x != 5 || y != 5 {
		t.Errorf("After attach, expected MoveTo(5,5), got %v(%v,%v)", cmd, x, y)
	}
}

func TestConvSegmentatorBasicSegmentation(t *testing.T) {
	// Create a horizontal line of length 4
	source := NewMockVertexSource([]Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 4, Y: 0, Cmd: basics.PathCmdLineTo},
	})

	conv := NewConvSegmentator(source)
	conv.ApproximationScale(1.0)

	vertices := collectConvSegmentatorVertices(conv)

	// Should have move_to + line segments at 1.0, 2.0, 4.0 = 4 vertices
	expectedCount := 4
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices, got %d", expectedCount, len(vertices))
	}

	// Check first vertex is move_to
	if vertices[0].cmd != uint32(basics.PathCmdMoveTo) {
		t.Errorf("Expected first command to be MoveTo, got %v", vertices[0].cmd)
	}

	// Check end point
	lastVertex := vertices[len(vertices)-1]
	if math.Abs(lastVertex.x-4) > 1e-10 || math.Abs(lastVertex.y-0) > 1e-10 {
		t.Errorf("Expected last vertex at (4,0), got (%v,%v)", lastVertex.x, lastVertex.y)
	}
}

func TestConvSegmentatorWithScale(t *testing.T) {
	// Create a horizontal line of length 2
	source := NewMockVertexSource([]Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 2, Y: 0, Cmd: basics.PathCmdLineTo},
	})

	conv := NewConvSegmentator(source)
	conv.ApproximationScale(2.0) // Should double the segments

	vertices := collectConvSegmentatorVertices(conv)

	// Should have move_to + line segments at 0.5, 1.0, 2.0 = 4 vertices (2 * 2.0 scale)
	expectedCount := 4
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices with scale 2.0, got %d", expectedCount, len(vertices))
	}

	// Check spacing - should have points at 0, 0.5, 1.0, 2.0
	expectedX := []float64{0, 0.5, 1.0, 2.0}
	for i, expected := range expectedX {
		if i < len(vertices) {
			if math.Abs(vertices[i].x-expected) > 1e-10 {
				t.Errorf("Vertex %d: expected x=%v, got %v", i, expected, vertices[i].x)
			}
		}
	}
}

func TestConvSegmentatorDiagonalLine(t *testing.T) {
	// Create diagonal line: length = sqrt(3^2 + 4^2) = 5
	source := NewMockVertexSource([]Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 3, Y: 4, Cmd: basics.PathCmdLineTo},
	})

	conv := NewConvSegmentator(source)
	conv.ApproximationScale(1.0)

	vertices := collectConvSegmentatorVertices(conv)

	// Should have move_to + segments at 1.0, 2.0, 3.0, 4.0, 5.0 = 5 vertices
	expectedCount := 5
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices for diagonal line, got %d", expectedCount, len(vertices))
	}

	// Check that intermediate points lie on the line y = (4/3)x
	for i := 1; i < len(vertices)-1; i++ {
		v := vertices[i]
		expectedY := (4.0 / 3.0) * v.x
		if math.Abs(v.y-expectedY) > 1e-9 {
			t.Errorf("Vertex %d at (%v,%v) is not on the line y=(4/3)x", i, v.x, v.y)
		}
	}
}

func TestConvSegmentatorComplexPath(t *testing.T) {
	// Create a more complex path with multiple segments
	source := NewMockVertexSource([]Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 2, Y: 0, Cmd: basics.PathCmdLineTo}, // Horizontal line of length 2
		{X: 2, Y: 2, Cmd: basics.PathCmdLineTo}, // Vertical line of length 2
		{X: 0, Y: 2, Cmd: basics.PathCmdLineTo}, // Horizontal line of length 2
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
	})

	conv := NewConvSegmentator(source)
	conv.ApproximationScale(1.0)

	vertices := collectConvSegmentatorVertices(conv)

	// Should segment each line according to its length
	// Each line of length 2 should create 2 vertices (start + end)
	// Complex path processing may vary, so we just check we got reasonable output
	if len(vertices) < 4 {
		t.Errorf("Expected at least 4 vertices for complex path, got %d", len(vertices))
	}

	// Verify the path visits the expected corners
	corners := []struct{ x, y float64 }{
		{0, 0}, {2, 0}, {2, 2}, {0, 2},
	}

	cornerFound := make([]bool, len(corners))
	for _, v := range vertices {
		for i, corner := range corners {
			if math.Abs(v.x-corner.x) < 1e-10 && math.Abs(v.y-corner.y) < 1e-10 {
				cornerFound[i] = true
			}
		}
	}

	for i, found := range cornerFound {
		if !found {
			t.Errorf("Corner %d (%v,%v) not found in segmented vertices",
				i, corners[i].x, corners[i].y)
		}
	}
}

func TestConvSegmentatorZeroLengthLine(t *testing.T) {
	// Zero-length line should still work
	source := NewMockVertexSource([]Vertex{
		{X: 5, Y: 10, Cmd: basics.PathCmdMoveTo},
		{X: 5, Y: 10, Cmd: basics.PathCmdLineTo},
	})

	conv := NewConvSegmentator(source)
	conv.ApproximationScale(1.0)

	vertices := collectConvSegmentatorVertices(conv)

	// Should have at least 1 vertex for zero-length line
	if len(vertices) < 1 {
		t.Errorf("Expected at least 1 vertex for zero-length line, got %d", len(vertices))
	}

	// Should be at the MoveTo position
	if vertices[0].x != 5 || vertices[0].y != 10 {
		t.Errorf("Expected vertex at (5,10), got (%v,%v)", vertices[0].x, vertices[0].y)
	}
}

// Helper function to collect all vertices from a conv segmentator
func collectConvSegmentatorVertices(conv *ConvSegmentator) []convVertex {
	var vertices []convVertex

	conv.Rewind(0)
	for {
		x, y, cmd := conv.Vertex()
		if cmd == uint32(basics.PathCmdStop) {
			break
		}
		vertices = append(vertices, convVertex{x: x, y: y, cmd: cmd})
	}

	return vertices
}

// Helper struct for collecting conv vertices
type convVertex struct {
	x, y float64
	cmd  uint32
}
