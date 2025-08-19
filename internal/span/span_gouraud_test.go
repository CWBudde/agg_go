package span

import (
	"testing"

	"agg_go/internal/basics"
)

// TestColor is a simple color type for testing.
type TestColor struct {
	Value int
}

func TestSpanGouraudCreation(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()
	if sg == nil {
		t.Fatal("NewSpanGouraud returned nil")
	}

	// Test initial vertex iteration
	_, _, cmd := sg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop initially, got %v", cmd)
	}
}

func TestSpanGouraudWithTriangle(t *testing.T) {
	c1 := TestColor{100}
	c2 := TestColor{200}
	c3 := TestColor{50}

	sg := NewSpanGouraudWithTriangle(c1, c2, c3, 0, 0, 100, 0, 50, 100, 0)
	if sg == nil {
		t.Fatal("NewSpanGouraudWithTriangle returned nil")
	}

	coord := sg.Coord()
	if coord[0].Color.Value != 100 || coord[1].Color.Value != 200 || coord[2].Color.Value != 50 {
		t.Errorf("Colors not set correctly: got %v, %v, %v", coord[0].Color.Value, coord[1].Color.Value, coord[2].Color.Value)
	}
}

func TestSpanGouraudColors(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()
	c1 := TestColor{10}
	c2 := TestColor{20}
	c3 := TestColor{30}

	sg.Colors(c1, c2, c3)

	coord := sg.Coord()
	if coord[0].Color.Value != 10 {
		t.Errorf("Expected color 1 = 10, got %d", coord[0].Color.Value)
	}
	if coord[1].Color.Value != 20 {
		t.Errorf("Expected color 2 = 20, got %d", coord[1].Color.Value)
	}
	if coord[2].Color.Value != 30 {
		t.Errorf("Expected color 3 = 30, got %d", coord[2].Color.Value)
	}
}

func TestSpanGouraudTriangleBasic(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()

	// Simple triangle without dilation
	sg.Triangle(0, 0, 100, 0, 50, 100, 0)

	coord := sg.Coord()
	if coord[0].X != 0 || coord[0].Y != 0 {
		t.Errorf("Expected vertex 1 at (0,0), got (%.1f,%.1f)", coord[0].X, coord[0].Y)
	}
	if coord[1].X != 100 || coord[1].Y != 0 {
		t.Errorf("Expected vertex 2 at (100,0), got (%.1f,%.1f)", coord[1].X, coord[1].Y)
	}
	if coord[2].X != 50 || coord[2].Y != 100 {
		t.Errorf("Expected vertex 3 at (50,100), got (%.1f,%.1f)", coord[2].X, coord[2].Y)
	}
}

func TestSpanGouraudTriangleWithDilation(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()

	// Triangle with dilation
	sg.Triangle(0, 0, 100, 0, 50, 50, 2.0)

	// After dilation, the coordinates should be modified
	coord := sg.Coord()

	// The exact values depend on the mathematical calculations,
	// but we can verify they're not the same as input
	originalDistSq := (coord[0].X)*(coord[0].X) + (coord[0].Y)*(coord[0].Y)
	if originalDistSq == 0 {
		t.Error("Dilation should modify coordinates")
	}
}

func TestSpanGouraudVertexSource(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()
	sg.Triangle(0, 0, 100, 0, 50, 100, 0)

	// Test vertex source interface
	sg.Rewind(0)

	// First vertex - MoveTo
	x, y, cmd := sg.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("Expected first vertex at (0,0), got (%.1f,%.1f)", x, y)
	}

	// Second vertex - LineTo
	x, y, cmd = sg.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo, got %v", cmd)
	}
	if x != 100 || y != 0 {
		t.Errorf("Expected second vertex at (100,0), got (%.1f,%.1f)", x, y)
	}

	// Third vertex - LineTo
	x, y, cmd = sg.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo, got %v", cmd)
	}
	if x != 50 || y != 100 {
		t.Errorf("Expected third vertex at (50,100), got (%.1f,%.1f)", x, y)
	}

	// End - Stop
	x, y, cmd = sg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop, got %v", cmd)
	}
}

func TestSpanGouraudArrangeVertices(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()
	sg.Colors(TestColor{1}, TestColor{2}, TestColor{3})
	sg.Triangle(50, 100, 0, 0, 100, 50, 0) // Intentionally unsorted

	arranged := sg.ArrangeVertices()

	// Should be sorted by Y coordinate: (0,0), (100,50), (50,100)
	if arranged[0].Y != 0 {
		t.Errorf("Expected first vertex Y=0, got Y=%.1f", arranged[0].Y)
	}
	if arranged[1].Y != 50 {
		t.Errorf("Expected second vertex Y=50, got Y=%.1f", arranged[1].Y)
	}
	if arranged[2].Y != 100 {
		t.Errorf("Expected third vertex Y=100, got Y=%.1f", arranged[2].Y)
	}

	// Colors should follow the vertices
	if arranged[0].X != 0 || arranged[0].Color.Value != 2 {
		t.Errorf("Expected first arranged vertex: (0,0) with color 2, got (%.1f,%.1f) with color %d",
			arranged[0].X, arranged[0].Y, arranged[0].Color.Value)
	}
}

func TestSpanGouraudArrangeVerticesEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		vertices [3][2]float64 // x, y pairs
		expected [3]float64    // expected Y order
	}{
		{
			name:     "Already sorted",
			vertices: [3][2]float64{{0, 0}, {50, 50}, {100, 100}},
			expected: [3]float64{0, 50, 100},
		},
		{
			name:     "Reverse sorted",
			vertices: [3][2]float64{{0, 100}, {50, 50}, {100, 0}},
			expected: [3]float64{0, 50, 100},
		},
		{
			name:     "Same Y coordinates",
			vertices: [3][2]float64{{0, 50}, {50, 50}, {100, 50}},
			expected: [3]float64{50, 50, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := NewSpanGouraud[TestColor]()
			sg.Colors(TestColor{1}, TestColor{2}, TestColor{3})
			sg.Triangle(tt.vertices[0][0], tt.vertices[0][1],
				tt.vertices[1][0], tt.vertices[1][1],
				tt.vertices[2][0], tt.vertices[2][1], 0)

			arranged := sg.ArrangeVertices()

			for i, expectedY := range tt.expected {
				if arranged[i].Y != expectedY {
					t.Errorf("Vertex %d: expected Y=%.1f, got Y=%.1f", i, expectedY, arranged[i].Y)
				}
			}
		})
	}
}

func TestSpanGouraudRewind(t *testing.T) {
	sg := NewSpanGouraud[TestColor]()
	sg.Triangle(0, 0, 100, 0, 50, 100, 0)

	// Read some vertices
	sg.Vertex()
	sg.Vertex()

	// Rewind
	sg.Rewind(0)

	// Should start from beginning again
	x, y, cmd := sg.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("After rewind, expected PathCmdMoveTo, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("After rewind, expected first vertex at (0,0), got (%.1f,%.1f)", x, y)
	}
}

func BenchmarkSpanGouraudTriangle(b *testing.B) {
	sg := NewSpanGouraud[TestColor]()
	c1, c2, c3 := TestColor{100}, TestColor{200}, TestColor{50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sg.Colors(c1, c2, c3)
		sg.Triangle(0, 0, 100, 0, 50, 100, 1.0)
	}
}

func BenchmarkSpanGouraudArrangeVertices(b *testing.B) {
	sg := NewSpanGouraud[TestColor]()
	sg.Colors(TestColor{1}, TestColor{2}, TestColor{3})
	sg.Triangle(50, 100, 0, 0, 100, 50, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sg.ArrangeVertices()
	}
}
