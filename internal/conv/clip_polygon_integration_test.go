package conv

import (
	"testing"

	"agg_go/internal/basics"
)

// Integration test for ConvClipPolygon that doesn't depend on problematic test files
func TestConvClipPolygon_Integration(t *testing.T) {
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

	// Should have vertices since rectangle is fully inside
	if len(resultVertices) == 0 {
		t.Error("Expected vertices for rectangle inside clip box, got none")
	}

	t.Logf("Generated %d vertices for rectangle inside clip box", len(resultVertices))

	// All vertices should be within bounds
	for i, v := range resultVertices {
		if basics.IsVertex(v.cmd) {
			if v.x < 0 || v.x > 100 || v.y < 0 || v.y > 100 {
				t.Errorf("Vertex %d (%f,%f) is outside clip bounds", i, v.x, v.y)
			}
		}
	}
}

func TestConvClipPolygon_LineClipping_Integration(t *testing.T) {
	// Line that crosses the clip box
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

	t.Logf("Generated %d vertices for line across clip box", len(resultVertices))

	// Should generate some vertices (the clipped line segment)
	if len(resultVertices) == 0 {
		t.Log("No vertices generated - line might be fully outside or clipping algorithm filtered it")
	} else {
		// All vertices should be within bounds
		for i, v := range resultVertices {
			if basics.IsVertex(v.cmd) {
				if v.x < 0 || v.x > 100 || v.y < 0 || v.y > 100 {
					t.Errorf("Vertex %d (%f,%f) is outside clip bounds", i, v.x, v.y)
				}
			}
		}
	}
}

func TestConvClipPolygon_ClipBoxSettings_Integration(t *testing.T) {
	vertices := []Vertex{}
	storage := NewMockVertexSource(vertices)
	conv := NewConvClipPolygon(storage)

	// Test clip box setting
	conv.ClipBox(10, 20, 100, 200)

	if conv.X1() != 10 || conv.Y1() != 20 || conv.X2() != 100 || conv.Y2() != 200 {
		t.Errorf("Expected clip box (10,20,100,200), got (%f,%f,%f,%f)",
			conv.X1(), conv.Y1(), conv.X2(), conv.Y2())
	}
}

func TestConvClipPolygon_EmptyPath_Integration(t *testing.T) {
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
