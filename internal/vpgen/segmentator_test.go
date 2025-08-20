package vpgen

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestVPGenSegmentator_Basic(t *testing.T) {
	vpgen := NewVPGenSegmentator()

	// Test default approximation scale
	if vpgen.ApproximationScale() != 1.0 {
		t.Errorf("Default approximation scale should be 1.0, got %v", vpgen.ApproximationScale())
	}

	// Test setting approximation scale
	vpgen.SetApproximationScale(2.0)
	if vpgen.ApproximationScale() != 2.0 {
		t.Errorf("ApproximationScale should be 2.0, got %v", vpgen.ApproximationScale())
	}

	// Test AutoClose/AutoUnclose
	if vpgen.AutoClose() {
		t.Error("AutoClose should return false for segmentator")
	}
	if vpgen.AutoUnclose() {
		t.Error("AutoUnclose should return false for segmentator")
	}
}

func TestVPGenSegmentator_ShortLine(t *testing.T) {
	vpgen := NewVPGenSegmentator()
	vpgen.SetApproximationScale(1.0)

	// Very short line - should produce just start and end points
	vpgen.Reset()
	vpgen.MoveTo(0, 0)
	vpgen.LineTo(1, 0)

	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// For debugging - print all vertices
	t.Logf("Got %d vertices:", len(vertices))
	for i, v := range vertices {
		t.Logf("  %d: (%v, %v, %v)", i, v.x, v.y, v.cmd)
	}

	// Should have two vertices (start and end points)
	if len(vertices) != 2 {
		t.Errorf("Expected 2 vertices for short line, got %d", len(vertices))
	}

	if len(vertices) >= 1 {
		// First vertex should be MoveTo command at start point (0,0)
		firstVertex := vertices[0]
		if firstVertex.cmd != basics.PathCmdMoveTo {
			t.Errorf("First vertex should be MoveTo command, got %v", firstVertex.cmd)
		}
		if firstVertex.x != 0 || firstVertex.y != 0 {
			t.Errorf("First vertex should be at start point (0,0), got (%v, %v)", firstVertex.x, firstVertex.y)
		}
	}

	if len(vertices) >= 2 {
		// Second vertex should be LineTo command at end point (1,0)
		secondVertex := vertices[1]
		if secondVertex.cmd != basics.PathCmdLineTo {
			t.Errorf("Second vertex should be LineTo command, got %v", secondVertex.cmd)
		}
		if secondVertex.x != 1 || secondVertex.y != 0 {
			t.Errorf("Second vertex should be at end point (1,0), got (%v, %v)", secondVertex.x, secondVertex.y)
		}
	}
}

func TestVPGenSegmentator_LongLineWithSegmentation(t *testing.T) {
	vpgen := NewVPGenSegmentator()
	vpgen.SetApproximationScale(1.0) // 1 unit per segment

	// Long line that should be segmented
	vpgen.Reset()
	vpgen.MoveTo(0, 0)
	vpgen.LineTo(10, 0) // 10 units long, should create multiple segments

	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should have many vertices (segmented)
	if len(vertices) < 5 {
		t.Errorf("Expected many vertices for long line, got %d", len(vertices))
	}

	// First should be MoveTo at start
	if vertices[0].cmd != basics.PathCmdMoveTo || vertices[0].x != 0 || vertices[0].y != 0 {
		t.Errorf("First vertex should be MoveTo(0,0), got (%v, %v, %v)",
			vertices[0].x, vertices[0].y, vertices[0].cmd)
	}

	// Rest should be LineTo commands
	for i := 1; i < len(vertices); i++ {
		if vertices[i].cmd != basics.PathCmdLineTo {
			t.Errorf("Vertex %d should be LineTo, got %v", i, vertices[i].cmd)
		}
		if vertices[i].y != 0 {
			t.Errorf("All vertices should have y=0, vertex %d has y=%v", i, vertices[i].y)
		}
	}

	// Last vertex should be at end point
	lastVertex := vertices[len(vertices)-1]
	if math.Abs(lastVertex.x-10) > 1e-10 {
		t.Errorf("Last vertex should have x=10, got x=%v", lastVertex.x)
	}
}

func TestVPGenSegmentator_ApproximationScale(t *testing.T) {
	// Test with different approximation scales
	testCases := []struct {
		name  string
		scale float64
	}{
		{"fine", 2.0},   // More segments
		{"coarse", 0.5}, // Fewer segments
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vpgen := NewVPGenSegmentator()
			vpgen.SetApproximationScale(tc.scale)

			vpgen.Reset()
			vpgen.MoveTo(0, 0)
			vpgen.LineTo(5, 0)

			vertexCount := 0
			for {
				_, _, cmd := vpgen.Vertex()
				if cmd == basics.PathCmdStop {
					break
				}
				vertexCount++
			}

			// Higher scale should produce more vertices
			if tc.scale > 1.0 && vertexCount < 3 {
				t.Errorf("High approximation scale should produce more vertices, got %d", vertexCount)
			}
		})
	}
}

func TestVPGenSegmentator_Reset(t *testing.T) {
	vpgen := NewVPGenSegmentator()

	// Setup some state
	vpgen.MoveTo(10, 10)
	vpgen.LineTo(20, 20)

	// Reset should clear state
	vpgen.Reset()

	// Should produce no vertices after reset
	x, y, cmd := vpgen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After reset should produce Stop, got (%v, %v, %v)", x, y, cmd)
	}
}

func TestVPGenSegmentator_ZeroLengthLine(t *testing.T) {
	vpgen := NewVPGenSegmentator()

	// Line with zero length
	vpgen.Reset()
	vpgen.MoveTo(5, 5)
	vpgen.LineTo(5, 5) // Same point

	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should produce at least one vertex (the endpoint)
	if len(vertices) < 1 {
		t.Errorf("Expected at least 1 vertex for zero-length line, got %d", len(vertices))
	}

	// All vertices should be at same location
	for i, v := range vertices {
		if v.x != 5 || v.y != 5 {
			t.Errorf("Vertex %d should be at (5,5), got (%v, %v)", i, v.x, v.y)
		}
	}
}

func TestVPGenSegmentator_MultipleLines(t *testing.T) {
	vpgen := NewVPGenSegmentator()
	vpgen.SetApproximationScale(1.0)

	// First line (length 3)
	vpgen.Reset()
	vpgen.MoveTo(0, 0)
	vpgen.LineTo(3, 0)

	// Consume first line vertices
	firstLineVertices := 0
	for {
		_, _, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		firstLineVertices++
	}

	// Second line from (3,0) to (3,4), length 4
	vpgen.LineTo(3, 4)

	// Consume second line vertices
	secondLineVertices := 0
	for {
		_, _, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		secondLineVertices++
	}

	// Log for debugging
	t.Logf("First line (length 3): %d vertices", firstLineVertices)
	t.Logf("Second line (length 4): %d vertices", secondLineVertices)

	// Both lines should produce at least one vertex each
	if firstLineVertices < 1 {
		t.Errorf("First line should produce at least 1 vertex, got %d", firstLineVertices)
	}
	if secondLineVertices < 1 {
		t.Errorf("Second line should produce at least 1 vertex, got %d", secondLineVertices)
	}
}

func TestVPGenSegmentator_DiagonalLine(t *testing.T) {
	vpgen := NewVPGenSegmentator()
	vpgen.SetApproximationScale(1.0)

	// Diagonal line
	vpgen.Reset()
	vpgen.MoveTo(0, 0)
	vpgen.LineTo(3, 4) // Length 5

	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for {
		x, y, cmd := vpgen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Check that segments maintain proper direction
	if len(vertices) >= 3 {
		// Check that x increases and y increases along the line
		for i := 1; i < len(vertices); i++ {
			if vertices[i].x < vertices[i-1].x || vertices[i].y < vertices[i-1].y {
				t.Errorf("Vertices should progress in diagonal direction, vertex %d regressed", i)
			}
		}
	}

	// Last vertex should be at end point
	if len(vertices) > 0 {
		lastVertex := vertices[len(vertices)-1]
		if math.Abs(lastVertex.x-3) > 1e-10 || math.Abs(lastVertex.y-4) > 1e-10 {
			t.Errorf("Last vertex should be at (3,4), got (%v, %v)", lastVertex.x, lastVertex.y)
		}
	}
}
