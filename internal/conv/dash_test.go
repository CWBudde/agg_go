package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// TestConvDashBasic tests basic dash functionality
func TestConvDashBasic(t *testing.T) {
	// Create a simple line path
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	// Create dash converter
	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0) // 10 unit dash, 5 unit gap

	// Collect vertices
	outputVertices := collectVertices(dash)

	// Should have alternating move_to and line_to commands for dashes
	if len(outputVertices) < 2 {
		t.Errorf("Expected at least 2 vertices for dashed line, got %d", len(outputVertices))
	}

	// First vertex should be move_to at start
	if outputVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be move_to, got %v", outputVertices[0].Cmd)
	}

	if outputVertices[0].X != 0 || outputVertices[0].Y != 0 {
		t.Errorf("Expected first vertex at (0,0), got (%f,%f)", outputVertices[0].X, outputVertices[0].Y)
	}
}

// TestConvDashMultipleDashes tests multiple dash patterns
func TestConvDashMultipleDashes(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 60, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0) // First pattern: 10 dash, 5 gap
	dash.AddDash(5.0, 10.0) // Second pattern: 5 dash, 10 gap

	outputVertices := collectVertices(dash)

	// Should have multiple segments
	if len(outputVertices) < 4 {
		t.Errorf("Expected at least 4 vertices for multi-dash pattern, got %d", len(outputVertices))
	}

	// Verify alternating pattern
	moveToCount := 0
	lineToCount := 0
	for _, v := range outputVertices {
		switch v.Cmd {
		case basics.PathCmdMoveTo:
			moveToCount++
		case basics.PathCmdLineTo:
			lineToCount++
		}
	}

	if moveToCount == 0 {
		t.Error("Expected at least one move_to command")
	}
	if lineToCount == 0 {
		t.Error("Expected at least one line_to command")
	}
}

// TestConvDashStartOffset tests dash start offset
func TestConvDashStartOffset(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0)
	dash.DashStart(2.5) // Start 2.5 units into the first dash

	outputVertices := collectVertices(dash)

	// Should still generate dashed line but with offset
	if len(outputVertices) == 0 {
		t.Error("Expected vertices from dashed line with offset")
	}

	// First vertex should still be move_to
	if outputVertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be move_to, got %v", outputVertices[0].Cmd)
	}
}

// TestConvDashShorten tests path shortening
func TestConvDashShorten(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0)
	dash.Shorten(10.0) // Shorten by 10 units

	outputVertices := collectVertices(dash)

	// Should still generate vertices but path should be shorter
	if len(outputVertices) == 0 {
		t.Error("Expected vertices from shortened dashed line")
	}

	// Check that shortening worked by verifying no vertex extends to x=100
	for _, v := range outputVertices {
		if v.X > 90.1 { // Allow small floating point error
			t.Errorf("Expected shortened path not to extend beyond ~90, found vertex at x=%f", v.X)
		}
	}
}

// TestConvDashClosedPath tests dashing on closed paths
func TestConvDashClosedPath(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 50, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 50, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(15.0, 5.0)

	outputVertices := collectVertices(dash)

	// Should handle closed path correctly
	if len(outputVertices) == 0 {
		t.Error("Expected vertices from dashed closed path")
	}

	// Should have move_to and line_to commands
	hasMoveTo := false
	hasLineTo := false
	for _, v := range outputVertices {
		switch v.Cmd {
		case basics.PathCmdMoveTo:
			hasMoveTo = true
		case basics.PathCmdLineTo:
			hasLineTo = true
		}
	}

	if !hasMoveTo {
		t.Error("Expected at least one move_to command for closed path")
	}
	if !hasLineTo {
		t.Error("Expected at least one line_to command for closed path")
	}
}

// TestConvDashRemoveAllDashes tests clearing dash patterns
func TestConvDashRemoveAllDashes(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0)
	dash.RemoveAllDashes()

	outputVertices := collectVertices(dash)

	// Should not generate any vertices when no dashes are defined
	if len(outputVertices) != 0 {
		t.Errorf("Expected no vertices after removing all dashes, got %d", len(outputVertices))
	}
}

// TestConvDashGetShorten tests getter for shorten value
func TestConvDashGetShorten(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 100, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)
	dash := NewConvDash(source)

	// Test default value
	if dash.GetShorten() != 0.0 {
		t.Errorf("Expected default shorten value of 0.0, got %f", dash.GetShorten())
	}

	// Test setting and getting
	dash.Shorten(5.5)
	if math.Abs(dash.GetShorten()-5.5) > 1e-10 {
		t.Errorf("Expected shorten value of 5.5, got %f", dash.GetShorten())
	}
}

// TestConvDashLongLine tests dashing on a longer line to verify pattern repetition
func TestConvDashLongLine(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 200, Y: 0, Cmd: basics.PathCmdLineTo},
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0) // Total pattern length: 15

	outputVertices := collectVertices(dash)

	// Should repeat the pattern multiple times
	if len(outputVertices) < 8 {
		t.Errorf("Expected many vertices for long dashed line, got %d", len(outputVertices))
	}

	// Verify we have alternating dash/gap pattern
	prevWasDash := false
	for i, v := range outputVertices {
		if i == 0 {
			// First vertex should be move_to (start of first dash)
			if v.Cmd != basics.PathCmdMoveTo {
				t.Errorf("Expected first vertex to be move_to, got %v", v.Cmd)
			}
			prevWasDash = true
		} else {
			if prevWasDash {
				// After a dash, should have either another line_to (continuing dash) or move_to (gap)
				if v.Cmd != basics.PathCmdLineTo && v.Cmd != basics.PathCmdMoveTo {
					t.Errorf("Unexpected command after dash: %v", v.Cmd)
				}
				if v.Cmd == basics.PathCmdMoveTo {
					prevWasDash = false
				}
			} else {
				// After a gap, should have move_to (start of next dash)
				if v.Cmd == basics.PathCmdMoveTo {
					prevWasDash = true
				}
			}
		}
	}
}

// TestConvDashAngledLine tests dashing on angled lines
func TestConvDashAngledLine(t *testing.T) {
	vertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 40, Cmd: basics.PathCmdLineTo}, // 3-4-5 triangle, length = 50
	}
	source := NewMockVertexSource(vertices)

	dash := NewConvDash(source)
	dash.AddDash(10.0, 5.0)

	outputVertices := collectVertices(dash)

	// Should generate dashed line along the angle
	if len(outputVertices) < 4 {
		t.Errorf("Expected multiple vertices for angled dashed line, got %d", len(outputVertices))
	}

	// Check that vertices lie along the expected line (slope = 4/3)
	for _, v := range outputVertices[1:] { // Skip first vertex at origin
		if v.X > 0 { // Avoid division by zero
			expectedY := (4.0 / 3.0) * v.X
			if math.Abs(v.Y-expectedY) > 0.1 {
				t.Errorf("Vertex (%f,%f) not on expected line, expected y=%f", v.X, v.Y, expectedY)
			}
		}
	}
}

// OutputVertex represents a vertex with coordinates and command for output
type OutputVertex struct {
	X, Y float64
	Cmd  basics.PathCommand
}

// collectVertices collects all vertices from a vertex source
func collectVertices(vs VertexSource) []OutputVertex {
	var vertices []OutputVertex
	vs.Rewind(0)

	for {
		x, y, cmd := vs.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, OutputVertex{X: x, Y: y, Cmd: cmd})
	}

	return vertices
}
