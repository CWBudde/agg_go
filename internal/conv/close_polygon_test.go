package conv

import (
	"testing"

	"agg_go/internal/basics"
)

func TestConvClosePolygon_NewConvClosePolygon(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	conv := NewConvClosePolygon(source)

	if conv == nil {
		t.Error("NewConvClosePolygon should return non-nil converter")
	}
}

func TestConvClosePolygon_EndPolyWithCloseFlag(t *testing.T) {
	// Create a polygon with EndPoly (no close flag initially)
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly}, // No close flag initially
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// The EndPoly command should have the close flag added
	endPolyFound := false
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) == 0 {
				t.Error("EndPoly command should have close flag set")
			}
			endPolyFound = true
			break
		}
	}

	if !endPolyFound {
		t.Error("Should find EndPoly command in output")
	}
}

func TestConvClosePolygon_PolygonWithStop(t *testing.T) {
	// Create a polygon that ends with Stop (no EndPoly)
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 0, basics.PathCmdLineTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 10, basics.PathCmdLineTo}, // Last vertex, no EndPoly
		// Stop command will be added by MockVertexSource
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should have EndPoly|Close inserted before Stop
	endPolyWithCloseFound := false
	stopFound := false
	endPolyIndex := -1
	stopIndex := -1

	for i, v := range output {
		if basics.IsEndPoly(v.Cmd) && (v.Cmd&basics.PathFlagClose) != 0 {
			endPolyWithCloseFound = true
			endPolyIndex = i
		}
		if basics.IsStop(v.Cmd) {
			stopFound = true
			stopIndex = i
		}
	}

	if !endPolyWithCloseFound {
		t.Error("Expected to find EndPoly command with close flag")
	}

	if !stopFound {
		t.Error("Expected to find Stop command")
	}

	if endPolyWithCloseFound && stopFound && endPolyIndex >= stopIndex {
		t.Error("EndPoly|Close should come before Stop")
	}
}

func TestConvClosePolygon_MultiplePolygonsWithMoveTo(t *testing.T) {
	// Create multiple polygons separated by MoveTo commands
	vertices := []Vertex{
		// First polygon - has line_to commands
		{0, 0, basics.PathCmdMoveTo},
		{5, 0, basics.PathCmdLineTo},
		{5, 5, basics.PathCmdLineTo},

		// Second polygon starts with MoveTo (should insert EndPoly|Close before it)
		{10, 10, basics.PathCmdMoveTo},
		{15, 10, basics.PathCmdLineTo},
		{15, 15, basics.PathCmdLineTo},
		// Ends with Stop (MockVertexSource adds this)
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should find EndPoly|Close inserted before the second MoveTo
	endPolyCloseCount := 0
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) && (v.Cmd&basics.PathFlagClose) != 0 {
			endPolyCloseCount++
		}
	}

	// Should have at least one EndPoly|Close (before second MoveTo, and potentially at end before Stop)
	if endPolyCloseCount < 1 {
		t.Errorf("Expected at least 1 EndPoly|Close command, got %d", endPolyCloseCount)
	}
}

func TestConvClosePolygon_NoLineToCommands(t *testing.T) {
	// Create a path with only MoveTo and no LineTo commands
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{5, 5, basics.PathCmdMoveTo}, // Another MoveTo, no LineTo commands
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// Should not insert any EndPoly|Close commands since no LineTo commands were seen
	endPolyCloseCount := 0
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) && (v.Cmd&basics.PathFlagClose) != 0 {
			endPolyCloseCount++
		}
	}

	if endPolyCloseCount > 0 {
		t.Error("Should not insert EndPoly|Close when no LineTo commands are present")
	}
}

func TestConvClosePolygon_EmptyPath(t *testing.T) {
	// Test with empty vertex source
	vertices := []Vertex{}
	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	x, y, cmd := conv.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v at (%.2f,%.2f)", cmd, x, y)
	}
}

func TestConvClosePolygon_SingleVertex(t *testing.T) {
	// Test with single vertex (degenerate case)
	vertices := []Vertex{
		{5, 5, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdEndPoly}, // No close flag initially
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

	// Collect all output vertices
	var output []Vertex
	for {
		x, y, cmd := conv.Vertex()
		output = append(output, Vertex{x, y, cmd})
		if cmd == basics.PathCmdStop {
			break
		}
	}

	// EndPoly should get close flag added
	endPolyFound := false
	for _, v := range output {
		if basics.IsEndPoly(v.Cmd) {
			if (v.Cmd & basics.PathFlagClose) == 0 {
				t.Error("EndPoly command should have close flag set")
			}
			endPolyFound = true
			break
		}
	}

	if !endPolyFound {
		t.Error("Should find EndPoly command in output")
	}
}

func TestConvClosePolygon_Rewind(t *testing.T) {
	vertices := []Vertex{
		{0, 0, basics.PathCmdMoveTo},
		{10, 10, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathFlagClose},
	}

	source := NewMockVertexSource(vertices)
	conv := NewConvClosePolygon(source)

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

func TestConvClosePolygon_Attach(t *testing.T) {
	vertices1 := []Vertex{{0, 0, basics.PathCmdMoveTo}}
	vertices2 := []Vertex{{5, 5, basics.PathCmdMoveTo}}

	source1 := NewMockVertexSource(vertices1)
	source2 := NewMockVertexSource(vertices2)

	conv := NewConvClosePolygon(source1)

	// Read from first source
	x, y, _ := conv.Vertex()
	if x != 0 || y != 0 {
		t.Error("Expected vertex from first source")
	}

	// Attach second source and rewind
	conv.Attach(source2)
	conv.Rewind(0)

	x, y, _ = conv.Vertex()
	if x != 5 || y != 5 {
		t.Error("Expected vertex from second source after attach")
	}
}
