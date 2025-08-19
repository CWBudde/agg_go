package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvClipPolygon_Creation(t *testing.T) {
	vertices := []Vertex{}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)

	// Test default clipping box
	if conv.X1() != 0 || conv.Y1() != 0 || conv.X2() != 1 || conv.Y2() != 1 {
		t.Errorf("Expected default clip box (0,0,1,1), got (%f,%f,%f,%f)",
			conv.X1(), conv.Y1(), conv.X2(), conv.Y2())
	}
}

func TestConvClipPolygon_ClipBox(t *testing.T) {
	vertices := []Vertex{}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)

	// Set clip box
	conv.ClipBox(10, 20, 100, 200)

	// Verify clip box was set
	if conv.X1() != 10 || conv.Y1() != 20 || conv.X2() != 100 || conv.Y2() != 200 {
		t.Errorf("Expected clip box (10,20,100,200), got (%f,%f,%f,%f)",
			conv.X1(), conv.Y1(), conv.X2(), conv.Y2())
	}
}

func TestConvClipPolygon_SimpleRectangleInside(t *testing.T) {
	vertices := []Vertex{
		{X: 20, Y: 20, Cmd: basics.PathCmdMoveTo},
		{X: 80, Y: 20, Cmd: basics.PathCmdLineTo},
		{X: 80, Y: 80, Cmd: basics.PathCmdLineTo},
		{X: 20, Y: 80, Cmd: basics.PathCmdLineTo},
	}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)
	conv.ClipBox(0, 0, 100, 100)

	// Rewind and collect all vertices
	conv.Rewind(0)
	var resultVertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := conv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
		if len(resultVertices) > 20 { // Safety check
			t.Fatal("Too many vertices generated")
		}
	}

	// Should have all original vertices since rectangle is fully inside
	if len(resultVertices) < 4 {
		t.Errorf("Expected at least 4 vertices for rectangle, got %d", len(resultVertices))
	}

	// First vertex should be MoveTo(20,20)
	if resultVertices[0].cmd != basics.PathCmdMoveTo || resultVertices[0].x != 20 || resultVertices[0].y != 20 {
		t.Errorf("Expected MoveTo(20,20), got %v(%f,%f)",
			resultVertices[0].cmd, resultVertices[0].x, resultVertices[0].y)
	}
}

func TestConvClipPolygon_LineAcrossClipBox(t *testing.T) {
	vertices := []Vertex{
		{X: -50, Y: 50, Cmd: basics.PathCmdMoveTo}, // outside left
		{X: 150, Y: 50, Cmd: basics.PathCmdLineTo}, // outside right
	}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)
	conv.ClipBox(0, 0, 100, 100)

	// Rewind and collect vertices
	conv.Rewind(0)
	var resultVertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := conv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
		if len(resultVertices) > 10 { // Safety check
			break
		}
	}

	// Should generate clipped line segment
	t.Logf("Generated %d vertices for line across clip box", len(resultVertices))

	// All vertices should be within bounds
	for i, v := range resultVertices {
		if basics.IsVertex(v.cmd) {
			if v.x < 0 || v.x > 100 || v.y < 0 || v.y > 100 {
				t.Errorf("Vertex %d (%f,%f) is outside clip bounds", i, v.x, v.y)
			}
		}
	}
}

func TestConvClipPolygon_EmptyPath(t *testing.T) {
	vertices := []Vertex{}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)
	conv.ClipBox(0, 0, 100, 100)

	// Rewind and check
	conv.Rewind(0)
	_, _, cmd := conv.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v", cmd)
	}
}

func TestConvClipPolygon_MultipleSubpaths(t *testing.T) {
	vertices := []Vertex{
		// First subpath - inside
		{X: 20, Y: 20, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 30, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},

		// Second subpath - outside
		{X: 200, Y: 200, Cmd: basics.PathCmdMoveTo},
		{X: 300, Y: 300, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly},
	}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)
	conv.ClipBox(0, 0, 100, 100)

	// Rewind and collect vertices
	conv.Rewind(0)
	var resultVertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	for {
		x, y, cmd := conv.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		resultVertices = append(resultVertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
		if len(resultVertices) > 20 { // Safety check
			break
		}
	}

	t.Logf("Generated %d vertices for multiple subpaths", len(resultVertices))

	// Should have at least some vertices from the first subpath
	if len(resultVertices) < 2 {
		t.Errorf("Expected at least 2 vertices from inside subpath, got %d", len(resultVertices))
	}
}
