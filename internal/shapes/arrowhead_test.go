package shapes

import (
	"testing"

	"agg_go/internal/basics"
)

func TestNewArrowhead(t *testing.T) {
	a := NewArrowhead()

	// Test default values
	if a.headD1 != 1.0 || a.headD2 != 1.0 || a.headD3 != 1.0 || a.headD4 != 0.0 {
		t.Error("Default head dimensions incorrect")
	}
	if a.tailD1 != 1.0 || a.tailD2 != 1.0 || a.tailD3 != 1.0 || a.tailD4 != 0.0 {
		t.Error("Default tail dimensions incorrect")
	}
	if a.headFlag != false || a.tailFlag != false {
		t.Error("Default flags should be false")
	}
	if a.currID != 0 || a.currCoord != 0 {
		t.Error("Default state should be zero")
	}
}

func TestArrowheadHead(t *testing.T) {
	a := NewArrowhead()
	a.Head(2.0, 1.5, 0.5, 0.25)

	if a.headD1 != 2.0 || a.headD2 != 1.5 || a.headD3 != 0.5 || a.headD4 != 0.25 {
		t.Error("Head dimensions not set correctly")
	}
	if !a.headFlag {
		t.Error("Head flag should be true after setting head")
	}
}

func TestArrowheadTail(t *testing.T) {
	a := NewArrowhead()
	a.Tail(3.0, 2.0, 1.0, 0.5)

	if a.tailD1 != 3.0 || a.tailD2 != 2.0 || a.tailD3 != 1.0 || a.tailD4 != 0.5 {
		t.Error("Tail dimensions not set correctly")
	}
	if !a.tailFlag {
		t.Error("Tail flag should be true after setting tail")
	}
}

func TestArrowheadEnableDisable(t *testing.T) {
	a := NewArrowhead()

	// Test enable/disable head
	a.EnableHead()
	if !a.headFlag {
		t.Error("EnableHead should set headFlag to true")
	}
	a.DisableHead()
	if a.headFlag {
		t.Error("DisableHead should set headFlag to false")
	}

	// Test enable/disable tail
	a.EnableTail()
	if !a.tailFlag {
		t.Error("EnableTail should set tailFlag to true")
	}
	a.DisableTail()
	if a.tailFlag {
		t.Error("DisableTail should set tailFlag to false")
	}
}

func TestArrowheadRewindDisabled(t *testing.T) {
	a := NewArrowhead()

	// Test tail disabled (pathID 0)
	a.Rewind(0)
	if a.cmd[0] != basics.PathCmdStop {
		t.Error("Disabled tail should start with PathCmdStop")
	}

	// Test head disabled (pathID 1)
	a.Rewind(1)
	if a.cmd[0] != basics.PathCmdStop {
		t.Error("Disabled head should start with PathCmdStop")
	}
}

func TestArrowheadRewindTail(t *testing.T) {
	a := NewArrowhead()
	a.Tail(2.0, 1.5, 1.0, 0.5)

	a.Rewind(0)

	// Check that currID and currCoord are reset
	if a.currID != 0 || a.currCoord != 0 {
		t.Error("Rewind should reset currID and currCoord")
	}

	// Verify tail coordinates are calculated correctly
	expectedCoords := [][2]float64{
		{2.0, 0.0},   // tailD1, 0
		{1.5, 1.0},   // tailD1 - tailD4, tailD3
		{-2.0, 1.0},  // -tailD2 - tailD4, tailD3
		{-1.5, 0.0},  // -tailD2, 0
		{-2.0, -1.0}, // -tailD2 - tailD4, -tailD3
		{1.5, -1.0},  // tailD1 - tailD4, -tailD3
	}

	for i, expected := range expectedCoords {
		if a.coord[i*2] != expected[0] || a.coord[i*2+1] != expected[1] {
			t.Errorf("Tail coordinate %d: expected (%.1f, %.1f), got (%.1f, %.1f)",
				i, expected[0], expected[1], a.coord[i*2], a.coord[i*2+1])
		}
	}

	// Verify tail commands
	expectedCmds := []basics.PathCommand{
		basics.PathCmdMoveTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW)),
		basics.PathCmdStop,
	}

	for i, expected := range expectedCmds {
		if a.cmd[i] != expected {
			t.Errorf("Tail command %d: expected %d, got %d", i, expected, a.cmd[i])
		}
	}
}

func TestArrowheadRewindHead(t *testing.T) {
	a := NewArrowhead()
	a.Head(3.0, 2.0, 1.5, 0.5)

	a.Rewind(1)

	// Check that currID and currCoord are reset
	if a.currID != 1 || a.currCoord != 0 {
		t.Error("Rewind should reset currCoord and set currID to 1")
	}

	// Verify head coordinates are calculated correctly
	expectedCoords := [][2]float64{
		{-3.0, 0.0}, // -headD1, 0
		{2.5, -1.5}, // headD2 + headD4, -headD3
		{2.0, 0.0},  // headD2, 0
		{2.5, 1.5},  // headD2 + headD4, headD3
	}

	for i, expected := range expectedCoords {
		if a.coord[i*2] != expected[0] || a.coord[i*2+1] != expected[1] {
			t.Errorf("Head coordinate %d: expected (%.1f, %.1f), got (%.1f, %.1f)",
				i, expected[0], expected[1], a.coord[i*2], a.coord[i*2+1])
		}
	}

	// Verify head commands
	expectedCmds := []basics.PathCommand{
		basics.PathCmdMoveTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW)),
		basics.PathCmdStop,
	}

	for i, expected := range expectedCmds[:6] {
		if a.cmd[i] != expected {
			t.Errorf("Head command %d: expected %d, got %d", i, expected, a.cmd[i])
		}
	}
}

func TestArrowheadVertexTail(t *testing.T) {
	a := NewArrowhead()
	a.Tail(1.0, 1.0, 1.0, 0.0)
	a.Rewind(0)

	var x, y float64
	vertices := []struct {
		expectedX   float64
		expectedY   float64
		expectedCmd basics.PathCommand
	}{
		{1.0, 0.0, basics.PathCmdMoveTo},
		{1.0, 1.0, basics.PathCmdLineTo},
		{-1.0, 1.0, basics.PathCmdLineTo},
		{-1.0, 0.0, basics.PathCmdLineTo},
		{-1.0, -1.0, basics.PathCmdLineTo},
		{1.0, -1.0, basics.PathCmdLineTo},
		{0.0, 0.0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))},
		{0.0, 0.0, basics.PathCmdStop},
	}

	for i, expected := range vertices {
		cmd := a.Vertex(&x, &y)
		if cmd != expected.expectedCmd {
			t.Errorf("Vertex %d: expected command %d, got %d", i, expected.expectedCmd, cmd)
		}
		if i < 6 { // Only check coordinates for actual vertices, not end/stop commands
			if x != expected.expectedX || y != expected.expectedY {
				t.Errorf("Vertex %d: expected (%.1f, %.1f), got (%.1f, %.1f)",
					i, expected.expectedX, expected.expectedY, x, y)
			}
		}
	}
}

func TestArrowheadVertexHead(t *testing.T) {
	a := NewArrowhead()
	a.Head(2.0, 1.0, 0.5, 0.0)
	a.Rewind(1)

	var x, y float64
	vertices := []struct {
		expectedX   float64
		expectedY   float64
		expectedCmd basics.PathCommand
	}{
		{-2.0, 0.0, basics.PathCmdMoveTo},
		{1.0, -0.5, basics.PathCmdLineTo},
		{1.0, 0.0, basics.PathCmdLineTo},
		{1.0, 0.5, basics.PathCmdLineTo},
		{0.0, 0.0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))},
		{0.0, 0.0, basics.PathCmdStop},
	}

	for i, expected := range vertices {
		cmd := a.Vertex(&x, &y)
		if cmd != expected.expectedCmd {
			t.Errorf("Vertex %d: expected command %d, got %d", i, expected.expectedCmd, cmd)
		}
		if i < 4 { // Only check coordinates for actual vertices, not end/stop commands
			if x != expected.expectedX || y != expected.expectedY {
				t.Errorf("Vertex %d: expected (%.1f, %.1f), got (%.1f, %.1f)",
					i, expected.expectedX, expected.expectedY, x, y)
			}
		}
	}
}

func TestArrowheadVertexInvalidPath(t *testing.T) {
	a := NewArrowhead()
	a.Head(1.0, 1.0, 1.0, 0.0)
	a.Rewind(2) // Invalid path ID

	var x, y float64
	cmd := a.Vertex(&x, &y)
	if cmd != basics.PathCmdStop {
		t.Error("Invalid path ID should return PathCmdStop")
	}
}

func TestArrowheadBothHeadAndTail(t *testing.T) {
	a := NewArrowhead()
	a.Head(2.0, 1.5, 1.0, 0.25)
	a.Tail(3.0, 2.0, 1.5, 0.5)

	// Test that both are enabled
	if !a.headFlag || !a.tailFlag {
		t.Error("Both head and tail should be enabled")
	}

	// Test tail generation
	a.Rewind(0)
	var x, y float64
	cmd := a.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Error("Tail should start with MoveTo")
	}
	if x != 3.0 || y != 0.0 {
		t.Errorf("First tail vertex should be (3.0, 0.0), got (%.1f, %.1f)", x, y)
	}

	// Test head generation
	a.Rewind(1)
	cmd = a.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Error("Head should start with MoveTo")
	}
	if x != -2.0 || y != 0.0 {
		t.Errorf("First head vertex should be (-2.0, 0.0), got (%.1f, %.1f)", x, y)
	}
}

func TestArrowheadVertexSourceInterface(t *testing.T) {
	// Ensure Arrowhead implements the VertexSource interface
	var _ interface {
		Rewind(uint32)
		Vertex(*float64, *float64) basics.PathCommand
	} = (*Arrowhead)(nil)
}
