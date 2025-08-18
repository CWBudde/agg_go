package path

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// MockVertexSource implements VertexSource for testing
type MockVertexSource struct {
	vertices []vertex
	index    int
}

type vertex struct {
	x, y float64
	cmd  uint32
}

func NewMockVertexSource(vertices []vertex) *MockVertexSource {
	return &MockVertexSource{
		vertices: vertices,
		index:    0,
	}
}

func (mvs *MockVertexSource) Rewind(pathID uint) {
	mvs.index = 0
}

func (mvs *MockVertexSource) NextVertex() (x, y float64, cmd uint32) {
	if mvs.index >= len(mvs.vertices) {
		return 0, 0, uint32(basics.PathCmdStop)
	}
	v := mvs.vertices[mvs.index]
	mvs.index++
	return v.x, v.y, v.cmd
}

func TestPathLength(t *testing.T) {
	const epsilon = 1e-10

	t.Run("EmptyPath", func(t *testing.T) {
		vs := NewMockVertexSource([]vertex{})
		length := PathLength(vs, 0)
		if length != 0 {
			t.Errorf("Expected length 0 for empty path, got %f", length)
		}
	})

	t.Run("SinglePoint", func(t *testing.T) {
		vs := NewMockVertexSource([]vertex{
			{10, 20, uint32(basics.PathCmdMoveTo)},
		})
		length := PathLength(vs, 0)
		if length != 0 {
			t.Errorf("Expected length 0 for single point, got %f", length)
		}
	})

	t.Run("SimpleLine", func(t *testing.T) {
		// Line from (0,0) to (3,4) - length should be 5
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{3, 4, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		expected := 5.0
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("Rectangle", func(t *testing.T) {
		// Rectangle 10x8 units - perimeter should be 36
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{10, 0, uint32(basics.PathCmdLineTo)},
			{10, 8, uint32(basics.PathCmdLineTo)},
			{0, 8, uint32(basics.PathCmdLineTo)},
			{0, 0, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		expected := 36.0
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("ClosedRectangle", func(t *testing.T) {
		// Rectangle 10x8 units with close command - perimeter should be 36
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{10, 0, uint32(basics.PathCmdLineTo)},
			{10, 8, uint32(basics.PathCmdLineTo)},
			{0, 8, uint32(basics.PathCmdLineTo)},
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
		})
		length := PathLength(vs, 0)
		expected := 36.0
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("Triangle", func(t *testing.T) {
		// Right triangle with sides 3, 4, 5
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{3, 0, uint32(basics.PathCmdLineTo)},
			{0, 4, uint32(basics.PathCmdLineTo)},
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
		})
		length := PathLength(vs, 0)
		expected := 3.0 + 5.0 + 4.0 // sides: 3 + sqrt(3²+4²) + 4
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("MultipleSubpaths", func(t *testing.T) {
		// Two separate line segments
		vs := NewMockVertexSource([]vertex{
			// First subpath: (0,0) to (3,4) - length 5
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{3, 4, uint32(basics.PathCmdLineTo)},
			// Second subpath: (10,0) to (13,4) - length 5
			{10, 0, uint32(basics.PathCmdMoveTo)},
			{13, 4, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		expected := 10.0 // 5 + 5
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("PathWithCurveCommands", func(t *testing.T) {
		// Path with curve commands (should be treated as vertices)
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{1, 1, uint32(basics.PathCmdCurve3)}, // Treated as vertex
			{2, 0, uint32(basics.PathCmdCurve3)}, // Treated as vertex
			{3, 1, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		// Distance calculations:
		// (0,0) to (1,1) = sqrt(2) ≈ 1.414
		// (1,1) to (2,0) = sqrt(2) ≈ 1.414
		// (2,0) to (3,1) = sqrt(2) ≈ 1.414
		expected := 3 * math.Sqrt(2)
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("ComplexPolygon", func(t *testing.T) {
		// Hexagon with radius 10 (all sides equal, length = 10 each)
		// For a regular hexagon, each side length equals the radius
		vs := NewMockVertexSource([]vertex{
			{10, 0, uint32(basics.PathCmdMoveTo)},                                  // Right
			{5, 8.660254, uint32(basics.PathCmdLineTo)},                            // Top right (10*sin(60°) ≈ 8.66)
			{-5, 8.660254, uint32(basics.PathCmdLineTo)},                           // Top left
			{-10, 0, uint32(basics.PathCmdLineTo)},                                 // Left
			{-5, -8.660254, uint32(basics.PathCmdLineTo)},                          // Bottom left
			{5, -8.660254, uint32(basics.PathCmdLineTo)},                           // Bottom right
			{10, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)}, // Close
		})
		length := PathLength(vs, 0)
		expected := 60.0                       // 6 sides of length 10 each
		if math.Abs(length-expected) > 0.001 { // Slightly larger epsilon for floating point
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("PathWithNonVertexCommands", func(t *testing.T) {
		// Test that non-vertex commands are properly ignored
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{5, 0, uint32(basics.PathCmdLineTo)},
			// Add some non-vertex command (like EndPoly without close)
			{0, 0, uint32(basics.PathCmdEndPoly)},
			{10, 0, uint32(basics.PathCmdMoveTo)}, // New subpath
			{15, 0, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		expected := 10.0 // 5 + 5 (two line segments)
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})
}

func TestPathLengthEdgeCases(t *testing.T) {
	const epsilon = 1e-10

	t.Run("PathStartingWithClose", func(t *testing.T) {
		// Path that starts with a close command (should be ignored)
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{5, 0, uint32(basics.PathCmdLineTo)},
		})
		length := PathLength(vs, 0)
		expected := 5.0
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("MultipleMoveTo", func(t *testing.T) {
		// Multiple MoveTo commands in sequence
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{1, 1, uint32(basics.PathCmdMoveTo)}, // This resets the start point
			{2, 2, uint32(basics.PathCmdMoveTo)}, // This resets again
			{5, 2, uint32(basics.PathCmdLineTo)}, // Line from (2,2) to (5,2)
		})
		length := PathLength(vs, 0)
		expected := 3.0 // Distance from (2,2) to (5,2)
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})

	t.Run("CloseWithoutPrecedingVertices", func(t *testing.T) {
		// Close command without any preceding vertices
		vs := NewMockVertexSource([]vertex{
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
		})
		length := PathLength(vs, 0)
		expected := 0.0
		if math.Abs(length-expected) > epsilon {
			t.Errorf("Expected length %f, got %f", expected, length)
		}
	})
}

// BenchmarkPathLength provides performance benchmarks for the PathLength function
func BenchmarkPathLength(b *testing.B) {
	// Create a complex polygon for benchmarking
	vertices := make([]vertex, 1000)
	for i := 0; i < 999; i++ {
		angle := float64(i) * 2 * math.Pi / 999
		cmd := uint32(basics.PathCmdLineTo)
		if i == 0 {
			cmd = uint32(basics.PathCmdMoveTo)
		}
		vertices[i] = vertex{
			x:   math.Cos(angle) * 100,
			y:   math.Sin(angle) * 100,
			cmd: cmd,
		}
	}
	vertices[999] = vertex{
		x:   vertices[0].x,
		y:   vertices[0].y,
		cmd: uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose),
	}

	vs := NewMockVertexSource(vertices)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PathLength(vs, 0)
	}
}
