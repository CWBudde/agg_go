package vcgen

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestNewVPGenSegmentator(t *testing.T) {
	seg := NewVPGenSegmentator()
	if seg == nil {
		t.Fatal("NewVPGenSegmentator returned nil")
	}

	// Check default values
	if seg.GetApproximationScale() != 1.0 {
		t.Errorf("Expected default approximation scale 1.0, got %v", seg.GetApproximationScale())
	}

	if seg.AutoClose() {
		t.Error("Expected AutoClose() to return false")
	}

	if seg.AutoUnclose() {
		t.Error("Expected AutoUnclose() to return false")
	}
}

func TestVPGenSegmentatorApproximationScale(t *testing.T) {
	seg := NewVPGenSegmentator()

	testScales := []float64{0.5, 1.0, 2.0, 10.0}
	for _, scale := range testScales {
		seg.ApproximationScale(scale)
		if got := seg.GetApproximationScale(); got != scale {
			t.Errorf("Expected approximation scale %v, got %v", scale, got)
		}
	}
}

func TestVPGenSegmentatorReset(t *testing.T) {
	seg := NewVPGenSegmentator()

	// Set some state
	seg.MoveTo(10, 20)
	seg.LineTo(30, 40)

	// Reset
	seg.Reset()

	// Should return stop command
	x, y, cmd := seg.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop after reset, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("Expected (0,0) after reset, got (%v,%v)", x, y)
	}
}

func TestVPGenSegmentatorBasicSegmentation(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(1.0)

	// Create a horizontal line of length 4 (should create 4 segments)
	seg.MoveTo(0, 0)
	seg.LineTo(4, 0)

	vertices := collectSegmentatorVertices(seg)

	// First vertex should be the move_to
	if len(vertices) < 1 {
		t.Fatal("Expected at least 1 vertex")
	}

	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be MoveTo, got %v", vertices[0].cmd)
	}

	if vertices[0].x != 0 || vertices[0].y != 0 {
		t.Errorf("Expected first vertex at (0,0), got (%v,%v)", vertices[0].x, vertices[0].y)
	}

	// Last vertex should be the end point
	lastVertex := vertices[len(vertices)-1]
	if math.Abs(lastVertex.x-4) > 1e-10 || math.Abs(lastVertex.y-0) > 1e-10 {
		t.Errorf("Expected last vertex at (4,0), got (%v,%v)", lastVertex.x, lastVertex.y)
	}

	// Check that we have evenly spaced vertices
	expectedCount := 4 // move_to + line segments at 1.0, 2.0, 4.0
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices, got %d", expectedCount, len(vertices))
	}
}

func TestVPGenSegmentatorWithScale(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(2.0) // Double the segments

	// Create a horizontal line of length 2 (should create 4 segments with scale 2.0)
	seg.MoveTo(0, 0)
	seg.LineTo(2, 0)

	vertices := collectSegmentatorVertices(seg)

	// Should have move_to + line segments at 0.5, 1.0, 2.0 = 4 vertices
	expectedCount := 4
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices with scale 2.0, got %d", expectedCount, len(vertices))
	}

	// Check intermediate points
	if len(vertices) >= 3 {
		// Second vertex should be at 0.5
		if math.Abs(vertices[1].x-0.5) > 1e-10 {
			t.Errorf("Expected second vertex x=0.5, got %v", vertices[1].x)
		}

		// Third vertex should be at 1.0
		if math.Abs(vertices[2].x-1.0) > 1e-10 {
			t.Errorf("Expected third vertex x=1.0, got %v", vertices[2].x)
		}
	}
}

func TestVPGenSegmentatorZeroLengthLine(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(1.0)

	// Zero-length line
	seg.MoveTo(5, 10)
	seg.LineTo(5, 10)

	vertices := collectSegmentatorVertices(seg)

	// Should produce only the move_to vertex for zero-length line
	if len(vertices) != 1 {
		t.Errorf("Expected 1 vertex for zero-length line, got %d", len(vertices))
	}

	// The vertex should be the MoveTo point
	if vertices[0].x != 5 || vertices[0].y != 10 {
		t.Errorf("Expected vertex at (5,10), got (%v,%v)", vertices[0].x, vertices[0].y)
	}
}

func TestVPGenSegmentatorDiagonalLine(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(1.0)

	// Diagonal line: length = sqrt(3^2 + 4^2) = 5
	seg.MoveTo(0, 0)
	seg.LineTo(3, 4)

	vertices := collectSegmentatorVertices(seg)

	// Should segment based on actual length (5), so 5 vertices total
	expectedCount := 5
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertices for diagonal line, got %d", expectedCount, len(vertices))
	}

	// Check that intermediate points lie on the line
	for i := 1; i < len(vertices)-1; i++ {
		v := vertices[i]
		// Point should satisfy: y = (4/3) * x
		expectedY := (4.0 / 3.0) * v.x
		if math.Abs(v.y-expectedY) > 1e-10 {
			t.Errorf("Vertex %d at (%v,%v) is not on the line y=(4/3)x", i, v.x, v.y)
		}
	}
}

func TestVPGenSegmentatorMultipleSegments(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(1.0)

	// Create a path with one line segment of length 2
	seg.MoveTo(0, 0)
	seg.LineTo(2, 0)

	vertices1 := collectSegmentatorVertices(seg)

	// Should give us move_to + end point = 2 vertices
	if len(vertices1) != 2 {
		t.Errorf("Expected 2 vertices for segment of length 2, got %d", len(vertices1))
	}

	// Reset and create a different segment
	seg.Reset()
	seg.MoveTo(2, 0)
	seg.LineTo(2, 2) // Vertical segment of length 2

	vertices2 := collectSegmentatorVertices(seg)

	// Should also give us 2 vertices
	if len(vertices2) != 2 {
		t.Errorf("Expected 2 vertices for second segment, got %d", len(vertices2))
	}
}

func TestVPGenSegmentatorSmallScale(t *testing.T) {
	seg := NewVPGenSegmentator()
	seg.ApproximationScale(0.1) // Very small scale

	// Line of length 10 with scale 0.1 should create just 1 segment
	seg.MoveTo(0, 0)
	seg.LineTo(10, 0)

	vertices := collectSegmentatorVertices(seg)

	// Should have only 1 vertex (move_to becomes the end point)
	expectedCount := 1
	if len(vertices) != expectedCount {
		t.Errorf("Expected %d vertex with small scale, got %d", expectedCount, len(vertices))
	}
}

// Helper function to collect all vertices from a segmentator
func collectSegmentatorVertices(seg *VPGenSegmentator) []vertex {
	var vertices []vertex

	for {
		x, y, cmd := seg.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, vertex{x: x, y: y, cmd: cmd})
	}

	return vertices
}

// Helper struct for collecting vertices
type vertex struct {
	x, y float64
	cmd  basics.PathCommand
}
