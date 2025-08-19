package vcgen

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// TestVCGenDashBasic tests basic dash generator functionality
func TestVCGenDashBasic(t *testing.T) {
	dash := NewVCGenDash()

	// Add vertices to form a simple line
	dash.AddVertex(0, 0, basics.PathCmdMoveTo)
	dash.AddVertex(100, 0, basics.PathCmdLineTo)
	dash.AddVertex(0, 0, basics.PathCmdEndPoly) // End the path

	// Add dash pattern
	dash.AddDash(10.0, 5.0)

	// Prepare and get vertices
	dash.PrepareSrc()
	vertices := collectVCGenVertices(dash)

	// Should generate alternating dash/gap pattern
	if len(vertices) == 0 {
		t.Error("Expected vertices from dash generator")
	}

	// First vertex should be move_to
	if vertices[0].cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be move_to, got %v", vertices[0].cmd)
	}
}

// TestVCGenDashAddDash tests adding dash patterns
func TestVCGenDashAddDash(t *testing.T) {
	dash := NewVCGenDash()

	// Initially should have no dashes
	if dash.numDashes != 0 {
		t.Errorf("Expected 0 dashes initially, got %d", dash.numDashes)
	}

	// Add first dash pattern
	dash.AddDash(10.0, 5.0)
	if dash.numDashes != 2 {
		t.Errorf("Expected 2 dash elements after first AddDash, got %d", dash.numDashes)
	}

	// Check pattern values
	if dash.dashes[0] != 10.0 {
		t.Errorf("Expected first dash length of 10.0, got %f", dash.dashes[0])
	}
	if dash.dashes[1] != 5.0 {
		t.Errorf("Expected first gap length of 5.0, got %f", dash.dashes[1])
	}
	if math.Abs(dash.totalDashLen-15.0) > 1e-10 {
		t.Errorf("Expected total dash length of 15.0, got %f", dash.totalDashLen)
	}

	// Add second dash pattern
	dash.AddDash(8.0, 2.0)
	if dash.numDashes != 4 {
		t.Errorf("Expected 4 dash elements after second AddDash, got %d", dash.numDashes)
	}
	if math.Abs(dash.totalDashLen-25.0) > 1e-10 {
		t.Errorf("Expected total dash length of 25.0, got %f", dash.totalDashLen)
	}
}

// TestVCGenDashRemoveAllDashes tests clearing dash patterns
func TestVCGenDashRemoveAllDashes(t *testing.T) {
	dash := NewVCGenDash()

	// Add some dashes
	dash.AddDash(10.0, 5.0)
	dash.AddDash(8.0, 2.0)

	// Remove all
	dash.RemoveAllDashes()

	if dash.numDashes != 0 {
		t.Errorf("Expected 0 dashes after RemoveAllDashes, got %d", dash.numDashes)
	}
	if dash.totalDashLen != 0.0 {
		t.Errorf("Expected total dash length of 0.0 after RemoveAllDashes, got %f", dash.totalDashLen)
	}
	if dash.currDash != 0 {
		t.Errorf("Expected current dash index of 0 after RemoveAllDashes, got %d", dash.currDash)
	}
	if dash.currDashStart != 0.0 {
		t.Errorf("Expected current dash start of 0.0 after RemoveAllDashes, got %f", dash.currDashStart)
	}
}

// TestVCGenDashStart tests dash start offset
func TestVCGenDashStart(t *testing.T) {
	dash := NewVCGenDash()
	dash.AddDash(10.0, 5.0)
	dash.AddDash(8.0, 2.0)

	// Test positive offset
	dash.DashStart(7.0)
	if dash.dashStart != 7.0 {
		t.Errorf("Expected dash start of 7.0, got %f", dash.dashStart)
	}
	if dash.currDash != 0 {
		t.Errorf("Expected current dash index of 0 for offset 7.0, got %d", dash.currDash)
	}
	if dash.currDashStart != 7.0 {
		t.Errorf("Expected current dash start of 7.0, got %f", dash.currDashStart)
	}

	// Test offset that moves to next dash
	dash.DashStart(12.0) // Should move past first dash (10) into gap (5), with 2.0 remaining in gap
	if dash.currDash != 1 {
		t.Errorf("Expected current dash index of 1 for offset 12.0, got %d", dash.currDash)
	}
	if dash.currDashStart != 2.0 {
		t.Errorf("Expected current dash start of 2.0 for offset 12.0, got %f", dash.currDashStart)
	}

	// Test negative offset (should use absolute value)
	dash.DashStart(-3.0)
	if dash.dashStart != -3.0 {
		t.Errorf("Expected dash start of -3.0, got %f", dash.dashStart)
	}
	if dash.currDashStart != 3.0 {
		t.Errorf("Expected current dash start of 3.0 for negative offset, got %f", dash.currDashStart)
	}
}

// TestVCGenDashShorten tests path shortening
func TestVCGenDashShorten(t *testing.T) {
	dash := NewVCGenDash()

	// Test default shorten value
	if dash.GetShorten() != 0.0 {
		t.Errorf("Expected default shorten value of 0.0, got %f", dash.GetShorten())
	}

	// Test setting shorten value
	dash.Shorten(5.5)
	if math.Abs(dash.GetShorten()-5.5) > 1e-10 {
		t.Errorf("Expected shorten value of 5.5, got %f", dash.GetShorten())
	}
}

// TestVCGenDashMaxDashes tests maximum dash limit
func TestVCGenDashMaxDashes(t *testing.T) {
	dash := NewVCGenDash()

	// Add maximum number of dashes
	for i := 0; i < MaxDashes/2; i++ {
		dash.AddDash(1.0, 1.0)
	}

	if dash.numDashes != MaxDashes {
		t.Errorf("Expected %d dash elements, got %d", MaxDashes, dash.numDashes)
	}

	// Try to add one more (should be ignored)
	dash.AddDash(1.0, 1.0)
	if dash.numDashes != MaxDashes {
		t.Errorf("Expected dash count to remain at %d after exceeding limit, got %d", MaxDashes, dash.numDashes)
	}
}

// TestVCGenDashRemoveAll tests clearing all vertices
func TestVCGenDashRemoveAll(t *testing.T) {
	dash := NewVCGenDash()

	// Add some vertices
	dash.AddVertex(0, 0, basics.PathCmdMoveTo)
	dash.AddVertex(100, 0, basics.PathCmdLineTo)

	// Remove all
	dash.RemoveAll()

	if dash.status != DashStatusInitial {
		t.Errorf("Expected status to be Initial after RemoveAll, got %v", dash.status)
	}
	if dash.srcVertices.Size() != 0 {
		t.Errorf("Expected 0 source vertices after RemoveAll, got %d", dash.srcVertices.Size())
	}
	if dash.closed != 0 {
		t.Errorf("Expected closed flag to be 0 after RemoveAll, got %d", dash.closed)
	}
}

// TestVCGenDashNoVertices tests behavior with no vertices
func TestVCGenDashNoVertices(t *testing.T) {
	dash := NewVCGenDash()
	dash.AddDash(10.0, 5.0)

	// Try to get vertices without adding any
	dash.PrepareSrc()
	_, _, cmd := dash.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop with no vertices, got %v", cmd)
	}
}

// TestVCGenDashNoDashes tests behavior with no dash patterns
func TestVCGenDashNoDashes(t *testing.T) {
	dash := NewVCGenDash()

	// Add vertices but no dash patterns
	dash.AddVertex(0, 0, basics.PathCmdMoveTo)
	dash.AddVertex(100, 0, basics.PathCmdLineTo)

	dash.PrepareSrc()
	_, _, cmd := dash.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop with no dash patterns, got %v", cmd)
	}
}

// TestVCGenDashSingleVertex tests behavior with only one vertex
func TestVCGenDashSingleVertex(t *testing.T) {
	dash := NewVCGenDash()
	dash.AddDash(10.0, 5.0)

	// Add only one vertex
	dash.AddVertex(0, 0, basics.PathCmdMoveTo)

	dash.PrepareSrc()
	_, _, cmd := dash.Vertex()

	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop with single vertex, got %v", cmd)
	}
}

// TestVCGenDashClosedPath tests closed path handling
func TestVCGenDashClosedPath(t *testing.T) {
	dash := NewVCGenDash()
	dash.AddDash(15.0, 5.0)

	// Create a closed square path
	dash.AddVertex(0, 0, basics.PathCmdMoveTo)
	dash.AddVertex(50, 0, basics.PathCmdLineTo)
	dash.AddVertex(50, 50, basics.PathCmdLineTo)
	dash.AddVertex(0, 50, basics.PathCmdLineTo)
	dash.AddVertex(0, 0, basics.PathCmdEndPoly)

	dash.PrepareSrc()
	vertices := collectVCGenVertices(dash)

	// Should generate vertices for the closed path
	if len(vertices) == 0 {
		t.Error("Expected vertices for closed path")
	}

	// Should have alternating move_to and line_to commands
	hasMoveTo := false
	hasLineTo := false
	for _, v := range vertices {
		if v.cmd == basics.PathCmdMoveTo {
			hasMoveTo = true
		} else if v.cmd == basics.PathCmdLineTo {
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

// VCGenVertex represents a vertex from VCGenDash
type VCGenVertex struct {
	x, y float64
	cmd  basics.PathCommand
}

// collectVCGenVertices collects all vertices from a VCGenDash
func collectVCGenVertices(dash *VCGenDash) []VCGenVertex {
	var vertices []VCGenVertex
	dash.Rewind(0)

	for {
		x, y, cmd := dash.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertices = append(vertices, VCGenVertex{x: x, y: y, cmd: cmd})
	}

	return vertices
}
