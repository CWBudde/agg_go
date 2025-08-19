package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// TestVertexSource implements a simple vertex source for testing
type TestVertexSource struct {
	vertices []TestVertex
	index    int
}

type TestVertex struct {
	x, y float64
	cmd  basics.PathCommand
}

func NewTestVertexSource(vertices []TestVertex) *TestVertexSource {
	return &TestVertexSource{vertices: vertices, index: 0}
}

func (t *TestVertexSource) Rewind(pathID uint) {
	t.index = 0
}

func (t *TestVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if t.index >= len(t.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := t.vertices[t.index]
	t.index++
	return v.x, v.y, v.cmd
}

func TestConvShortenPath_Basic(t *testing.T) {
	// Create a simple line path: (0,0) -> (10,0) -> (20,0)
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{20, 0, basics.PathCmdLineTo},
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	// Test with no shortening
	converter.SetShorten(0)
	converter.Rewind(0)

	// Should get all vertices back
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 0 {
		t.Errorf("Expected LineTo(10,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo || x != 20 || y != 0 {
		t.Errorf("Expected LineTo(20,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop, got %v", cmd)
	}
}

func TestConvShortenPath_ShortenFromEnd(t *testing.T) {
	// Create a simple line path: (0,0) -> (10,0) -> (20,0)
	// Total length is 20, each segment is 10
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{20, 0, basics.PathCmdLineTo},
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	// Shorten by 5 units from the end
	converter.SetShorten(5)
	if converter.Shorten() != 5 {
		t.Errorf("Expected shorten value 5, got %.1f", converter.Shorten())
	}

	converter.Rewind(0)

	// Should get first vertex
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Should get second vertex
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 0 {
		t.Errorf("Expected LineTo(10,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Should get shortened third vertex at (15,0) instead of (20,0)
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected LineTo, got %v", cmd)
	}
	if math.Abs(x-15.0) > 0.001 || math.Abs(y-0.0) > 0.001 {
		t.Errorf("Expected LineTo(15,0), got LineTo(%.1f,%.1f)", x, y)
	}

	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop, got %v", cmd)
	}
}

func TestConvShortenPath_ShortenWholeSegment(t *testing.T) {
	// Create a path where shortening removes a whole segment
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{20, 0, basics.PathCmdLineTo},
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	// Shorten by 10 units (exactly one segment)
	converter.SetShorten(10)
	converter.Rewind(0)

	// Should get first vertex
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Should get second vertex (unchanged)
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 0 {
		t.Errorf("Expected LineTo(10,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Third vertex should be removed completely
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop, got %v", cmd)
	}
}

func TestConvShortenPath_ShortenMoreThanPath(t *testing.T) {
	// Test shortening more than the total path length
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	// Shorten by 20 units (more than the 10 unit path)
	converter.SetShorten(20)
	converter.Rewind(0)

	// Should get no vertices (path completely removed)
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop for over-shortened path, got %v(%.1f,%.1f)", cmd, x, y)
	}
}

func TestConvShortenPath_EmptyPath(t *testing.T) {
	// Test with empty path
	vertices := []TestVertex{}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	converter.SetShorten(5)
	converter.Rewind(0)

	_, _, cmd := converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop for empty path, got %v", cmd)
	}
}

func TestConvShortenPath_SingleVertex(t *testing.T) {
	// Test with single vertex
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	converter.SetShorten(5)
	converter.Rewind(0)

	// Should still get the single vertex
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop, got %v", cmd)
	}
}

func TestConvShortenPath_DiagonalLine(t *testing.T) {
	// Test with diagonal line to verify distance calculation
	vertices := []TestVertex{
		{0, 0, basics.PathCmdMoveTo},
		{3, 4, basics.PathCmdLineTo}, // Distance = 5
		{6, 8, basics.PathCmdLineTo}, // Another 5 units
	}

	source := NewTestVertexSource(vertices)
	converter := NewConvShortenPath(source)

	// Shorten by 2.5 units (half the last segment)
	converter.SetShorten(2.5)
	converter.Rewind(0)

	// First vertex unchanged
	x, y, cmd := converter.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("Expected MoveTo(0,0), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Second vertex unchanged
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo || x != 3 || y != 4 {
		t.Errorf("Expected LineTo(3,4), got %v(%.1f,%.1f)", cmd, x, y)
	}

	// Third vertex should be shortened to midpoint between (3,4) and (6,8)
	x, y, cmd = converter.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected LineTo, got %v", cmd)
	}
	expectedX := 3 + (6-3)*0.5 // 4.5
	expectedY := 4 + (8-4)*0.5 // 6.0
	if math.Abs(x-expectedX) > 0.001 || math.Abs(y-expectedY) > 0.001 {
		t.Errorf("Expected LineTo(%.1f,%.1f), got LineTo(%.1f,%.1f)", expectedX, expectedY, x, y)
	}
}
