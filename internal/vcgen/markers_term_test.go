package vcgen

import (
	"testing"

	"agg_go/internal/basics"
)

func TestVCGenMarkersTerm_Basic(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Add a simple line segment
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(30, 40, basics.PathCmdLineTo)

	// Test start marker (path_id = 0)
	gen.Rewind(0)

	// Should get start marker (MoveTo -> LineTo)
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo for start marker, got %v", cmd)
	}
	if x != 10 || y != 20 {
		t.Errorf("Expected start marker at (10,20), got (%f,%f)", x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for start marker, got %v", cmd)
	}
	if x != 30 || y != 40 {
		t.Errorf("Expected start marker end at (30,40), got (%f,%f)", x, y)
	}

	// Should stop after start marker
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop after start marker, got %v", cmd)
	}

	// Test end marker (path_id = 1)
	gen.Rewind(1)

	// Should get end marker (MoveTo -> LineTo) - reversed direction
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo for end marker, got %v", cmd)
	}
	if x != 30 || y != 40 {
		t.Errorf("Expected end marker at (30,40), got (%f,%f)", x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for end marker, got %v", cmd)
	}
	if x != 10 || y != 20 {
		t.Errorf("Expected end marker end at (10,20), got (%f,%f)", x, y)
	}

	// Should stop after end marker
	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop after end marker, got %v", cmd)
	}
}

func TestVCGenMarkersTerm_MultipleMoveTo(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Multiple MoveTo calls should modify the last marker
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(15, 25, basics.PathCmdMoveTo) // Should replace previous
	gen.AddVertex(30, 40, basics.PathCmdLineTo)

	gen.Rewind(0)

	// Should get updated start marker
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %v", cmd)
	}
	if x != 15 || y != 25 {
		t.Errorf("Expected updated start marker at (15,25), got (%f,%f)", x, y)
	}
}

func TestVCGenMarkersTerm_MultipleLineTo(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Multiple LineTo calls should update the end marker
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(30, 40, basics.PathCmdLineTo)
	gen.AddVertex(50, 60, basics.PathCmdLineTo) // Should update end marker

	// Test start marker (path_id = 0)
	gen.Rewind(0)

	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo for start marker, got %v", cmd)
	}
	if x != 10 || y != 20 {
		t.Errorf("Expected start marker at (10,20), got (%f,%f)", x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for start marker, got %v", cmd)
	}
	if x != 30 || y != 40 {
		t.Errorf("Expected start marker end at (30,40), got (%f,%f)", x, y)
	}

	// Test end marker (path_id = 1) - should reflect last LineTo position
	gen.Rewind(1)

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo for end marker, got %v", cmd)
	}
	if x != 50 || y != 60 {
		t.Errorf("Expected end marker at (50,60), got (%f,%f)", x, y)
	}

	x, y, cmd = gen.Vertex()
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for end marker, got %v", cmd)
	}
	if x != 30 || y != 40 {
		t.Errorf("Expected end marker end at (30,40), got (%f,%f)", x, y)
	}
}

func TestVCGenMarkersTerm_RemoveAll(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Add some markers
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(30, 40, basics.PathCmdLineTo)

	// Remove all
	gen.RemoveAll()
	gen.Rewind(0)

	// Should immediately stop
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop after RemoveAll, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("Expected (0,0) after RemoveAll, got (%f,%f)", x, y)
	}
}

func TestVCGenMarkersTerm_EmptyPath(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	gen.Rewind(0)

	// Should immediately stop for empty path
	x, y, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("Expected (0,0) for empty path, got (%f,%f)", x, y)
	}
}

func TestVCGenMarkersTerm_OnlyMoveTo(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Only MoveTo, no LineTo
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)

	gen.Rewind(0)

	// Should stop because there's no complete line segment
	_, _, cmd := gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for incomplete path, got %v", cmd)
	}
}

func TestVCGenMarkersTerm_PathID(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Add markers
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(30, 40, basics.PathCmdLineTo)

	// Test different path IDs
	gen.Rewind(0)
	_, _, cmd := gen.Vertex()
	if cmd == basics.PathCmdStop {
		t.Error("Path ID 0 should have markers")
	}

	gen.Rewind(1)
	_, _, cmd = gen.Vertex()
	if cmd == basics.PathCmdStop {
		t.Error("Path ID 1 should have markers")
	}

	gen.Rewind(2)
	_, _, cmd = gen.Vertex()
	if cmd != basics.PathCmdStop {
		t.Error("Path ID 2 should be beyond available markers")
	}
}

func TestVCGenMarkersTerm_ComplexPath(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// Complex path with multiple segments
	gen.AddVertex(0, 0, basics.PathCmdMoveTo)
	gen.AddVertex(10, 10, basics.PathCmdLineTo)
	gen.AddVertex(20, 10, basics.PathCmdLineTo)
	gen.AddVertex(30, 0, basics.PathCmdLineTo)

	// Test start marker (path_id = 0)
	gen.Rewind(0)

	var startCommands []basics.PathCommand
	var startCoords [][2]float64

	// Collect start marker vertices
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		startCommands = append(startCommands, cmd)
		startCoords = append(startCoords, [2]float64{x, y})
	}

	// Should have exactly 2 commands for start marker: MoveTo, LineTo
	if len(startCommands) != 2 {
		t.Errorf("Expected 2 start marker commands, got %d", len(startCommands))
	}

	// Start marker should be from start (0,0) to first LineTo (10,10)
	if len(startCoords) >= 2 {
		if startCoords[0][0] != 0 || startCoords[0][1] != 0 {
			t.Errorf("Start marker should begin at (0,0), got (%f,%f)",
				startCoords[0][0], startCoords[0][1])
		}
		if startCoords[1][0] != 10 || startCoords[1][1] != 10 {
			t.Errorf("Start marker should end at (10,10), got (%f,%f)",
				startCoords[1][0], startCoords[1][1])
		}
	}

	// Test end marker (path_id = 1)
	gen.Rewind(1)

	var endCommands []basics.PathCommand
	var endCoords [][2]float64

	// Collect end marker vertices
	for {
		x, y, cmd := gen.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		endCommands = append(endCommands, cmd)
		endCoords = append(endCoords, [2]float64{x, y})
	}

	// Should have exactly 2 commands for end marker: MoveTo, LineTo
	if len(endCommands) != 2 {
		t.Errorf("Expected 2 end marker commands, got %d", len(endCommands))
	}

	// End marker should be from final point back to second-to-last
	if len(endCoords) >= 2 {
		if endCoords[0][0] != 30 || endCoords[0][1] != 0 {
			t.Errorf("End marker should begin at (30,0), got (%f,%f)",
				endCoords[0][0], endCoords[0][1])
		}
		if endCoords[1][0] != 20 || endCoords[1][1] != 10 {
			t.Errorf("End marker should end at (20,10), got (%f,%f)",
				endCoords[1][0], endCoords[1][1])
		}
	}
}

func TestVCGenMarkersTerm_PrepareSrc(t *testing.T) {
	gen := NewVCGenMarkersTerm()

	// PrepareSrc should not panic and should be a no-op
	gen.PrepareSrc()

	// Should still be able to add vertices and get results
	gen.AddVertex(10, 20, basics.PathCmdMoveTo)
	gen.AddVertex(30, 40, basics.PathCmdLineTo)

	gen.Rewind(0)
	_, _, cmd := gen.Vertex()
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("PrepareSrc should not affect vertex generation, got %v", cmd)
	}
}
