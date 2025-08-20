package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvMarkerAdaptor_Basic(t *testing.T) {
	// Create a simple path: line from (0,0) to (10,0)
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)

	// Test that it implements the basic vertex source interface
	adaptor.Rewind(0)

	vertexCount := 0
	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
	}

	// Should have at least the original vertices
	if vertexCount < 2 {
		t.Errorf("Expected at least 2 vertices, got %d", vertexCount)
	}
}

func TestConvMarkerAdaptor_WithMarkers(t *testing.T) {
	// Create a simple path
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{10, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	// Create a simple marker implementation
	markers := &NullMarkers{} // Use the existing null markers for simplicity

	adaptor := NewConvMarkerAdaptorWithMarkers(path, markers)

	adaptor.Rewind(0)

	vertexCount := 0
	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
	}

	// Should process the path vertices
	if vertexCount < 3 {
		t.Errorf("Expected at least 3 vertices, got %d", vertexCount)
	}
}

func TestConvMarkerAdaptor_Shortening(t *testing.T) {
	// Create a path
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)

	// Test shortening functionality
	adaptor.SetShorten(2.0)

	if adaptor.Shorten() != 2.0 {
		t.Errorf("Expected shorten value 2.0, got %f", adaptor.Shorten())
	}

	// The path should be shortened
	adaptor.Rewind(0)

	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should have vertices (exact behavior depends on shortening implementation)
	if len(vertices) == 0 {
		t.Error("Expected some vertices after shortening")
	}
}

func TestConvMarkerAdaptor_Attach(t *testing.T) {
	// Create initial path
	path1 := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path1)

	// Create second path
	path2 := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{1, 1, basics.PathCmdLineTo},
		{2, 2, basics.PathCmdLineTo},
		{3, 3, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	// Attach the new path
	adaptor.Attach(path2)
	adaptor.Rewind(0)

	vertexCount := 0
	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
	}

	// Should have vertices from the new path (4 vertices)
	if vertexCount < 4 {
		t.Errorf("Expected at least 4 vertices from new path, got %d", vertexCount)
	}
}

func TestConvMarkerAdaptor_Rewind(t *testing.T) {
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{1, 0, basics.PathCmdLineTo},
		{2, 0, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)

	// Read some vertices
	adaptor.Rewind(0)
	adaptor.Vertex()
	adaptor.Vertex()

	// Rewind and read again
	adaptor.Rewind(0)
	x1, y1, cmd1 := adaptor.Vertex()

	if cmd1 != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo after rewind, got %v", cmd1)
	}

	if x1 != 0 || y1 != 0 {
		t.Errorf("Expected first vertex at (0,0) after rewind, got (%.3f, %.3f)", x1, y1)
	}
}

func TestConvMarkerAdaptor_EmptyPath(t *testing.T) {
	// Empty path
	path := NewMockVertexSource([]Vertex{})

	adaptor := NewConvMarkerAdaptor(path)
	adaptor.Rewind(0)

	// Should immediately return stop
	_, _, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v", cmd)
	}
}

// TestMarkerImpl is a simple marker implementation for testing
type TestMarkerImpl struct {
	vertices []Vertex
	position int
}

func NewTestMarkerImpl() *TestMarkerImpl {
	return &TestMarkerImpl{
		vertices: make([]Vertex, 0),
		position: 0,
	}
}

func (m *TestMarkerImpl) RemoveAll() {
	m.vertices = m.vertices[:0]
	m.position = 0
}

func (m *TestMarkerImpl) AddVertex(x, y float64, cmd basics.PathCommand) {
	// Only add start and end markers for line segments
	if basics.IsMoveTo(cmd) || basics.IsLineTo(cmd) {
		// Add a simple marker shape (small cross) at this position
		m.vertices = append(m.vertices,
			Vertex{x - 1, y, basics.PathCmdMoveTo},
			Vertex{x + 1, y, basics.PathCmdLineTo},
			Vertex{x, y - 1, basics.PathCmdMoveTo},
			Vertex{x, y + 1, basics.PathCmdLineTo},
		)
	}
}

func (m *TestMarkerImpl) PrepareSrc() {
	m.position = 0
}

func (m *TestMarkerImpl) Rewind(pathID uint) {
	m.position = 0
}

func (m *TestMarkerImpl) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.position >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}

	v := m.vertices[m.position]
	m.position++
	return v.X, v.Y, v.Cmd
}

func TestConvMarkerAdaptor_CustomMarkers(t *testing.T) {
	// Create a simple path
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	markers := NewTestMarkerImpl()
	adaptor := NewConvMarkerAdaptorWithMarkers(path, markers)

	adaptor.Rewind(0)

	pathVertexCount := 0
	markerVertexCount := 0

	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}

		// Count vertices - first comes the path, then markers
		if pathVertexCount < 3 { // Original path has 3 vertices (move + 2 lines)
			pathVertexCount++
		} else {
			markerVertexCount++
		}
	}

	// Should have processed original path vertices
	if pathVertexCount != 3 {
		t.Errorf("Expected 3 path vertices, got %d", pathVertexCount)
	}

	// Should have marker vertices (4 vertices per marker point, 3 points = 12 vertices)
	if markerVertexCount != 12 {
		t.Errorf("Expected 12 marker vertices, got %d", markerVertexCount)
	}
}

func TestConvMarkerAdaptor_ShorteningEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		pathLength      float64
		shortenAmount   float64
		shouldHaveVerts bool
		expectedLastX   float64
	}{
		{
			name:            "Zero shortening",
			pathLength:      10,
			shortenAmount:   0,
			shouldHaveVerts: true,
			expectedLastX:   10,
		},
		{
			name:            "Partial shortening",
			pathLength:      10,
			shortenAmount:   3,
			shouldHaveVerts: true,
			expectedLastX:   7, // Should stop at 7 instead of 10
		},
		{
			name:            "Complete shortening",
			pathLength:      10,
			shortenAmount:   10,
			shouldHaveVerts: false, // Path completely shortened
		},
		{
			name:            "Over-shortening",
			pathLength:      10,
			shortenAmount:   15,
			shouldHaveVerts: false, // Path completely shortened
		},
		{
			name:            "Negative shortening",
			pathLength:      10,
			shortenAmount:   -1,
			shouldHaveVerts: true,
			expectedLastX:   10, // Should ignore negative values
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := NewMockVertexSource([]Vertex{
				{0, 0, basics.PathCmdMoveTo},
				{test.pathLength, 0, basics.PathCmdLineTo},
				{0, 0, basics.PathCmdStop},
			})

			adaptor := NewConvMarkerAdaptor(path)
			adaptor.SetShorten(test.shortenAmount)
			adaptor.Rewind(0)

			vertices := []Vertex{}
			for {
				x, y, cmd := adaptor.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertices = append(vertices, Vertex{x, y, cmd})
			}

			if test.shouldHaveVerts {
				if len(vertices) == 0 {
					t.Errorf("Expected vertices but got none")
				} else {
					// Check last vertex position for partial shortening
					if test.shortenAmount > 0 && test.shortenAmount < test.pathLength {
						lastVertex := vertices[len(vertices)-1]
						tolerance := 0.001
						if abs(lastVertex.X-test.expectedLastX) > tolerance {
							t.Errorf("Expected last vertex X to be %.3f, got %.3f",
								test.expectedLastX, lastVertex.X)
						}
					}
				}
			} else {
				if len(vertices) > 0 {
					t.Errorf("Expected no vertices but got %d", len(vertices))
				}
			}
		})
	}
}

func TestConvMarkerAdaptor_ComplexPaths(t *testing.T) {
	// Test with a multi-segment path
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{10, 5, basics.PathCmdLineTo},
		{15, 5, basics.PathCmdLineTo},
		{20, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)
	adaptor.Rewind(0)

	vertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, Vertex{x, y, cmd})
	}

	// Should have all 5 vertices
	if len(vertices) != 5 {
		t.Errorf("Expected 5 vertices, got %d", len(vertices))
	}

	// First vertex should be MoveTo
	if vertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be MoveTo, got %v", vertices[0].Cmd)
	}

	// Rest should be LineTo
	for i := 1; i < len(vertices); i++ {
		if vertices[i].Cmd != basics.PathCmdLineTo {
			t.Errorf("Expected vertex %d to be LineTo, got %v", i, vertices[i].Cmd)
		}
	}
}

func TestConvMarkerAdaptor_ClosedPolygon(t *testing.T) {
	// Test with a closed polygon
	path := NewMockVertexSource([]Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)
	adaptor.Rewind(0)

	vertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, Vertex{x, y, cmd})
	}

	// Should process the closed polygon correctly
	if len(vertices) < 4 {
		t.Errorf("Expected at least 4 vertices for closed polygon, got %d", len(vertices))
	}

	// Check that we get the expected sequence
	if vertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be MoveTo, got %v", vertices[0].Cmd)
	}
}

func TestConvMarkerAdaptor_RewindBehavior(t *testing.T) {
	path := NewMockVertexSource([]Vertex{
		{1, 2, basics.PathCmdMoveTo},
		{3, 4, basics.PathCmdLineTo},
		{5, 6, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	})

	adaptor := NewConvMarkerAdaptor(path)

	// First pass
	adaptor.Rewind(0)
	firstPassVertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		firstPassVertices = append(firstPassVertices, Vertex{x, y, cmd})
	}

	// Second pass after rewind
	adaptor.Rewind(0)
	secondPassVertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		secondPassVertices = append(secondPassVertices, Vertex{x, y, cmd})
	}

	// Both passes should yield the same results
	if len(firstPassVertices) != len(secondPassVertices) {
		t.Errorf("Different vertex counts between passes: %d vs %d",
			len(firstPassVertices), len(secondPassVertices))
	}

	for i := 0; i < len(firstPassVertices) && i < len(secondPassVertices); i++ {
		v1, v2 := firstPassVertices[i], secondPassVertices[i]
		if v1.X != v2.X || v1.Y != v2.Y || v1.Cmd != v2.Cmd {
			t.Errorf("Vertex %d differs between passes: (%.3f,%.3f,%v) vs (%.3f,%.3f,%v)",
				i, v1.X, v1.Y, v1.Cmd, v2.X, v2.Y, v2.Cmd)
		}
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
