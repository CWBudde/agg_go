package polygon

import (
	"testing"

	"agg_go/internal/basics"
)

func TestNewSimplePolygonVertexSource(t *testing.T) {
	polygon := []float64{0.0, 0.0, 100.0, 0.0, 50.0, 86.6}
	numPoints := uint(3)

	vs := NewSimplePolygonVertexSource(polygon, numPoints, false, true)

	if vs.numPoints != numPoints {
		t.Errorf("numPoints = %d, want %d", vs.numPoints, numPoints)
	}

	if !vs.IsClose() {
		t.Error("IsClose() should be true")
	}

	if vs.roundoff {
		t.Error("roundoff should be false")
	}
}

func TestSimplePolygonVertexSourceCloseFlag(t *testing.T) {
	polygon := []float64{0.0, 0.0, 100.0, 0.0, 50.0, 86.6}
	vs := NewSimplePolygonVertexSource(polygon, 3, false, true)

	// Test initial state
	if !vs.IsClose() {
		t.Error("Should be closed initially")
	}

	// Test setting close to false
	vs.Close(false)
	if vs.IsClose() {
		t.Error("Should not be closed after setting to false")
	}

	// Test setting close to true
	vs.Close(true)
	if !vs.IsClose() {
		t.Error("Should be closed after setting to true")
	}
}

func TestSimplePolygonVertexGeneration(t *testing.T) {
	polygon := []float64{
		0.0, 0.0, // Point 0
		100.0, 0.0, // Point 1
		50.0, 86.6, // Point 2
	}
	vs := NewSimplePolygonVertexSource(polygon, 3, false, true)

	expectedVertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{
		{0.0, 0.0, basics.PathCmdMoveTo},
		{100.0, 0.0, basics.PathCmdLineTo},
		{50.0, 86.6, basics.PathCmdLineTo},
		{0.0, 0.0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose))},
	}

	vs.Rewind(0)

	for i, expected := range expectedVertices {
		x, y, cmd := vs.Vertex()

		if x != expected.x {
			t.Errorf("Vertex %d: x = %f, want %f", i, x, expected.x)
		}
		if y != expected.y {
			t.Errorf("Vertex %d: y = %f, want %f", i, y, expected.y)
		}
		if cmd != expected.cmd {
			t.Errorf("Vertex %d: cmd = %d, want %d", i, cmd, expected.cmd)
		}
	}

	// Next vertex should be stop
	_, _, cmd := vs.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Final vertex cmd = %d, want %d", cmd, basics.PathCmdStop)
	}
}

func TestSimplePolygonVertexGenerationOpen(t *testing.T) {
	polygon := []float64{
		0.0, 0.0,
		100.0, 0.0,
		50.0, 86.6,
	}
	vs := NewSimplePolygonVertexSource(polygon, 3, false, false) // open polygon

	expectedVertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{
		{0.0, 0.0, basics.PathCmdMoveTo},
		{100.0, 0.0, basics.PathCmdLineTo},
		{50.0, 86.6, basics.PathCmdLineTo},
		{0.0, 0.0, basics.PathCmdEndPoly}, // No close flag
	}

	vs.Rewind(0)

	for i, expected := range expectedVertices {
		x, y, cmd := vs.Vertex()

		if x != expected.x {
			t.Errorf("Vertex %d: x = %f, want %f", i, x, expected.x)
		}
		if y != expected.y {
			t.Errorf("Vertex %d: y = %f, want %f", i, y, expected.y)
		}
		if cmd != expected.cmd {
			t.Errorf("Vertex %d: cmd = %d, want %d", i, cmd, expected.cmd)
		}
	}
}

func TestSimplePolygonVertexGenerationRoundoff(t *testing.T) {
	polygon := []float64{
		0.3, 0.7,
		100.6, 0.2,
		50.8, 86.9,
	}
	vs := NewSimplePolygonVertexSource(polygon, 3, true, true) // with roundoff

	expectedVertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{
		{0.5, 0.5, basics.PathCmdMoveTo},   // floor(0.3) + 0.5, floor(0.7) + 0.5
		{100.5, 0.5, basics.PathCmdLineTo}, // floor(100.6) + 0.5, floor(0.2) + 0.5
		{50.5, 86.5, basics.PathCmdLineTo}, // floor(50.8) + 0.5, floor(86.9) + 0.5
		{0.0, 0.0, basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose))},
	}

	vs.Rewind(0)

	for i, expected := range expectedVertices {
		x, y, cmd := vs.Vertex()

		if i < 3 { // Only check coordinates for actual vertices
			if x != expected.x {
				t.Errorf("Vertex %d: x = %f, want %f", i, x, expected.x)
			}
			if y != expected.y {
				t.Errorf("Vertex %d: y = %f, want %f", i, y, expected.y)
			}
		}
		if cmd != expected.cmd {
			t.Errorf("Vertex %d: cmd = %d, want %d", i, cmd, expected.cmd)
		}
	}
}

func TestSimplePolygonVertexSourceRewind(t *testing.T) {
	polygon := []float64{0.0, 0.0, 100.0, 0.0, 50.0, 86.6}
	vs := NewSimplePolygonVertexSource(polygon, 3, false, true)

	// Generate some vertices
	vs.Rewind(0)
	vs.Vertex() // MoveTo
	vs.Vertex() // LineTo

	// Rewind and check we start from beginning
	vs.Rewind(0)
	x, y, cmd := vs.Vertex()

	if x != 0.0 || y != 0.0 {
		t.Errorf("After rewind, first vertex = (%f, %f), want (0.0, 0.0)", x, y)
	}
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("After rewind, first cmd = %d, want %d", cmd, basics.PathCmdMoveTo)
	}
}

func TestSimplePolygonVertexSourceEmptyPolygon(t *testing.T) {
	polygon := []float64{}
	vs := NewSimplePolygonVertexSource(polygon, 0, false, true)

	vs.Rewind(0)
	_, _, cmd := vs.Vertex()

	expectedCmd := basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose))
	if cmd != expectedCmd {
		t.Errorf("Empty polygon should return EndPoly+Close, got %d, want %d", cmd, expectedCmd)
	}

	// Next call should return stop
	_, _, cmd = vs.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("After EndPoly, should return Stop, got %d", cmd)
	}
}

func TestSimplePolygonVertexSourceSinglePoint(t *testing.T) {
	polygon := []float64{50.0, 75.0}
	vs := NewSimplePolygonVertexSource(polygon, 1, false, true)

	vs.Rewind(0)

	// Should get MoveTo
	x, y, cmd := vs.Vertex()
	if x != 50.0 || y != 75.0 || cmd != basics.PathCmdMoveTo {
		t.Errorf("Single point: got (%f, %f, %d), want (50.0, 75.0, %d)",
			x, y, cmd, basics.PathCmdMoveTo)
	}

	// Should get EndPoly
	_, _, cmd = vs.Vertex()
	if cmd != basics.PathCommand(uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose)) {
		t.Errorf("Single point EndPoly: got %d, want %d",
			cmd, basics.PathCommand(uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose)))
	}

	// Should get Stop
	_, _, cmd = vs.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Single point Stop: got %d, want %d", cmd, basics.PathCmdStop)
	}
}
