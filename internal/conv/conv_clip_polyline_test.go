package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvClipPolyline_BasicLine(t *testing.T) {
	// Create a simple polyline: move to (2,2), line to (8,8)
	vertices := []Vertex{
		{2, 2, basics.PathCmdMoveTo},
		{8, 8, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should get MoveTo(2,2) then LineTo(8,8) for line inside clip box
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 8 {
		t.Errorf("Expected line_to (8,8), got %v (%f,%f)", cmd, x, y)
	}

	// Should be done
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestConvClipPolyline_ClippedLine(t *testing.T) {
	// Create a line that extends outside: move to (5,5), line to (15,5)
	vertices := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should get MoveTo(5,5) then clipped LineTo(10,5)
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 5 || y != 5 {
		t.Errorf("Expected move_to (5,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 5 {
		t.Errorf("Expected clipped line_to (10,5), got %v (%f,%f)", cmd, x, y)
	}

	// Should be done
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestConvClipPolyline_CompletelyClippedLine(t *testing.T) {
	// Create a line completely outside: move to (20,20), line to (30,30)
	vertices := []Vertex{
		{20, 20, basics.PathCmdMoveTo},
		{30, 30, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should get no vertices
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop for completely clipped line, got %v (%f,%f)", cmd, x, y)
	}
}

func TestConvClipPolyline_MultipleSegments(t *testing.T) {
	// Create a polyline with multiple segments
	vertices := []Vertex{
		{2, 2, basics.PathCmdMoveTo},
		{8, 2, basics.PathCmdLineTo},  // inside -> inside
		{15, 2, basics.PathCmdLineTo}, // inside -> outside
		{5, 2, basics.PathCmdLineTo},  // outside -> inside
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should get MoveTo(2,2) from original MoveTo
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2), got %v (%f,%f)", cmd, x, y)
	}

	// First segment: should get LineTo(8,2)
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 2 {
		t.Errorf("Expected line_to (8,2), got %v (%f,%f)", cmd, x, y)
	}

	// Second segment: should get clipped LineTo(10,2)
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 2 {
		t.Errorf("Expected clipped line_to (10,2), got %v (%f,%f)", cmd, x, y)
	}

	// Third segment: should get MoveTo(10,2) due to disconnection
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 2 {
		t.Errorf("Expected move_to (10,2), got %v (%f,%f)", cmd, x, y)
	}

	// Third segment: should get LineTo(5,2)
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 5 || y != 2 {
		t.Errorf("Expected line_to (5,2), got %v (%f,%f)", cmd, x, y)
	}

	// Should be done
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestConvClipPolyline_LineAcrossBoundary(t *testing.T) {
	// Create a line that crosses from outside to outside through the clip box
	vertices := []Vertex{
		{-5, 5, basics.PathCmdMoveTo},
		{15, 5, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should get move_to and line_to for the clipped segment
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 0 || y != 5 {
		t.Errorf("Expected move_to (0,5), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 5 {
		t.Errorf("Expected line_to (10,5), got %v (%f,%f)", cmd, x, y)
	}

	// Should be done
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop, got %v", cmd)
	}
}

func TestConvClipPolyline_ClipBoxAccessors(t *testing.T) {
	vertices := []Vertex{{0, 0, basics.PathCmdStop}}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)

	conv.ClipBox(1, 2, 8, 9)

	if conv.X1() != 1 {
		t.Errorf("Expected X1=1, got %f", conv.X1())
	}
	if conv.Y1() != 2 {
		t.Errorf("Expected Y1=2, got %f", conv.Y1())
	}
	if conv.X2() != 8 {
		t.Errorf("Expected X2=8, got %f", conv.X2())
	}
	if conv.Y2() != 9 {
		t.Errorf("Expected Y2=9, got %f", conv.Y2())
	}
}

func TestConvClipPolyline_EmptyPath(t *testing.T) {
	// Empty path
	vertices := []Vertex{
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	conv.Rewind(0)

	// Should immediately stop
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected stop for empty path, got %v (%f,%f)", cmd, x, y)
	}
}

func TestConvClipPolyline_Rewind(t *testing.T) {
	// Test that rewind properly resets the converter
	vertices := []Vertex{
		{2, 2, basics.PathCmdMoveTo},
		{8, 8, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}
	source := NewMockVertexSource(vertices)
	conv := NewConvClipPolyline(source)
	conv.ClipBox(0, 0, 10, 10)

	// First iteration - should get MoveTo then LineTo
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2), got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 8 {
		t.Errorf("Expected line_to (8,8), got %v (%f,%f)", cmd, x, y)
	}

	// Rewind and iterate again - should get same sequence
	conv.Rewind(0)
	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 2 || y != 2 {
		t.Errorf("Expected move_to (2,2) after rewind, got %v (%f,%f)", cmd, x, y)
	}

	x, y, cmd = conv.Vertex()
	if cmd != basics.PathCmdLineTo || x != 8 || y != 8 {
		t.Errorf("Expected line_to (8,8) after rewind, got %v (%f,%f)", cmd, x, y)
	}
}
