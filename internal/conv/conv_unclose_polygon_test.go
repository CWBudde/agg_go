package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvUnclosePolygon_NewConvUnclosePolygon(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	conv := NewConvUnclosePolygon(source)

	if conv == nil {
		t.Error("NewConvUnclosePolygon should return non-nil converter")
	}
}

func TestConvUnclosePolygon_RemoveCloseFlag(t *testing.T) {
	// Create a closed polygon with EndPoly|Close
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose}, // Has close flag
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// The EndPoly command should have the close flag removed
	endPolyFound := false
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) != 0 {
				t.Error("EndPoly command should NOT have close flag set")
			}
			endPolyFound = true
			break
		}
	}

	if !endPolyFound {
		t.Error("Should find EndPoly command in output")
	}
}

func TestConvUnclosePolygon_AlreadyOpenPolygon(t *testing.T) {
	// Create a polygon that's already open (no close flag)
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly}, // No close flag initially
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should find EndPoly without close flag (unchanged)
	endPolyFound := false
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) != 0 {
				t.Error("EndPoly command should remain without close flag")
			}
			endPolyFound = true
			break
		}
	}

	if !endPolyFound {
		t.Error("Should find EndPoly command in output")
	}
}

func TestConvUnclosePolygon_MultiplePolygons(t *testing.T) {
	// Create multiple polygons, some closed, some open
	vertices := []Vertex{
		// First polygon - closed
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose}, // Closed

		// Second polygon - open
		{10, 10, basics.PathCmdMoveTo},
		{15, 10, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdEndPoly}, // Open

		// Third polygon - closed
		{20, 20, basics.PathCmdMoveTo},
		{25, 20, basics.PathCmdLineTo},
		{25, 25, basics.PathCmdLineTo},
		{20, 20, basics.PathCmdEndPoly | basics.PathFlagClose}, // Closed
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Count EndPoly commands and verify none have close flags
	endPolyCount := 0
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) != 0 {
				t.Error("All EndPoly commands should have close flag removed")
			}
			endPolyCount++
		}
	}

	if endPolyCount != 3 {
		t.Errorf("Expected 3 EndPoly commands, got %d", endPolyCount)
	}
}

func TestConvUnclosePolygon_NoEndPolyCommands(t *testing.T) {
	// Create a path with no EndPoly commands
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		// No EndPoly - MockVertexSource will add Stop
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should not find any EndPoly commands to modify
	endPolyCount := 0
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			endPolyCount++
		}
	}

	if endPolyCount != 0 {
		t.Error("Should not find any EndPoly commands when none are present")
	}
}

func TestConvUnclosePolygon_OnlyMoveToCommands(t *testing.T) {
	// Create a path with only MoveTo commands (no polygons)
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 5, basics.PathCmdMoveTo},
		{10, 10, basics.PathCmdMoveTo},
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Verify all MoveTo commands pass through unchanged
	moveToCount := 0
	for _, v := range output {
		if basics.IsMoveTo(v.Cmd) {
			moveToCount++
		}
	}

	if moveToCount != 3 {
		t.Errorf("Expected 3 MoveTo commands, got %d", moveToCount)
	}
}

func TestConvUnclosePolygon_EmptyPath(t *testing.T) {
	// Test with empty vertex source
	vertices := []Vertex{}
	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v at (%.2f,%.2f)", cmd, x, y)
	}
}

func TestConvUnclosePolygon_SingleVertex(t *testing.T) {
	// Test with single vertex ending with EndPoly|Close
	vertices := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose}, // Has close flag
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// EndPoly should have close flag removed
	endPolyFound := false
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) != 0 {
				t.Error("EndPoly command should have close flag removed")
			}
			endPolyFound = true
			break
		}
	}

	if !endPolyFound {
		t.Error("Should find EndPoly command in output")
	}
}

func TestConvUnclosePolygon_OtherCommandsUnchanged(t *testing.T) {
	// Test that non-EndPoly commands pass through unchanged
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdLineTo},
		{0, 5, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose}, // EndPoly with close flag
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Verify commands pass through unchanged except for EndPoly
	lineToCount := 0
	foundEndPolyWithoutCloseFlag := false

	for _, v := range output {
		if basics.IsLineTo(v.Cmd) {
			lineToCount++
		}
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) != 0 {
				t.Error("EndPoly should have close flag removed")
			}
			foundEndPolyWithoutCloseFlag = true
		}
	}

	if lineToCount != 3 {
		t.Errorf("Expected 3 LineTo commands, got %d", lineToCount)
	}

	if !foundEndPolyWithoutCloseFlag {
		t.Error("Should find EndPoly without close flag")
	}
}

func TestConvUnclosePolygon_Rewind(t *testing.T) {
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvUnclosePolygon(source)

	// Read some vertices
	conv.Vertex()
	conv.Vertex()

	// Rewind and read again
	conv.Rewind(0)
	x, y, cmd := conv.Vertex()

	if cmd != basics.PathCmdMoveTo || x != 0 || y != 0 {
		t.Errorf("After rewind, expected MoveTo (0,0), got %v (%.2f,%.2f)", cmd, x, y)
	}
}

func TestConvUnclosePolygon_Attach(t *testing.T) {
	vertices1 := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}
	vertices2 := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdEndPoly}, // No close flag
	}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)

	conv := NewConvUnclosePolygon(source1)

	// Read from first source - EndPoly should have close flag removed
	conv.Vertex()              // MoveTo
	_, _, cmd := conv.Vertex() // EndPoly
	if basics.IsEndPoly(cmd) && (cmd&basics.PathFlagClose) != 0 {
		t.Error("Expected close flag to be removed from first source")
	}

	// Attach second source and rewind
	conv.Attach(source2)
	conv.Rewind(0)

	// Read from second source
	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 5 || y != 5 {
		t.Error("Expected MoveTo from second source after attach")
	}

	_, _, cmd = conv.Vertex() // EndPoly
	if basics.IsEndPoly(cmd) && (cmd&basics.PathFlagClose) != 0 {
		t.Error("Second source EndPoly should remain without close flag")
	}
}
