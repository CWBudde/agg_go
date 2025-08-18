package conv

import (
	"agg_go/internal/basics"
	"testing"
)

// MockVPGen implements VPGen for testing
type MockVPGen struct {
	vertices    []Vertex
	index       int
	autoClose   bool
	autoUnclose bool
	resetCalled bool
}

func NewMockVPGen(autoClose, autoUnclose bool) *MockVPGen {
	return &MockVPGen{
		autoClose:   autoClose,
		autoUnclose: autoUnclose,
	}
}

func (m *MockVPGen) Reset() {
	m.vertices = nil
	m.index = 0
	m.resetCalled = true
}

func (m *MockVPGen) MoveTo(x, y float64) {
	m.vertices = append(m.vertices, Vertex{x, y, basics.PathCmdMoveTo})
}

func (m *MockVPGen) LineTo(x, y float64) {
	m.vertices = append(m.vertices, Vertex{x, y, basics.PathCmdLineTo})
}

func (m *MockVPGen) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := m.vertices[m.index]
	m.index++
	return v.X, v.Y, v.Cmd
}

func (m *MockVPGen) AutoClose() bool {
	return m.autoClose
}

func (m *MockVPGen) AutoUnclose() bool {
	return m.autoUnclose
}

// Helper function to create a simple path
func createSimplePath() *MockVertexSource {
	return NewMockVertexSource([]Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{50, 60, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly},
	})
}

func TestConvAdaptorVPGen_Basic(t *testing.T) {
	source := createSimplePath()
	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	// Test rewind
	adaptor.Rewind(0)
	if !vpgen.resetCalled {
		t.Error("VPGen.Reset() should have been called")
	}

	// Process vertices
	x, y, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 20 {
		t.Errorf("Expected MoveTo(10, 20), got %v(%f, %f)", cmd, x, y)
	}

	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdLineTo || x != 30 || y != 40 {
		t.Errorf("Expected LineTo(30, 40), got %v(%f, %f)", cmd, x, y)
	}

	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdLineTo || x != 50 || y != 60 {
		t.Errorf("Expected LineTo(50, 60), got %v(%f, %f)", cmd, x, y)
	}

	// Should process EndPoly
	_, _, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdEndPoly {
		t.Errorf("Expected EndPoly, got %v", cmd)
	}

	// Should stop
	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop, got %v", cmd)
	}
}

func TestConvAdaptorVPGen_AutoClose(t *testing.T) {
	source := NewMockVertexSource([]Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{50, 60, basics.PathCmdLineTo},
		{70, 80, basics.PathCmdMoveTo}, // New path triggers auto-close
		{90, 100, basics.PathCmdLineTo},
	})

	vpgen := NewMockVPGen(true, false) // Auto-close enabled
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	adaptor.Rewind(0)

	// Process first path
	x, y, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 20 {
		t.Errorf("Expected MoveTo(10, 20), got %v(%f, %f)", cmd, x, y)
	}

	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdLineTo || x != 30 || y != 40 {
		t.Errorf("Expected LineTo(30, 40), got %v(%f, %f)", cmd, x, y)
	}

	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdLineTo || x != 50 || y != 60 {
		t.Errorf("Expected LineTo(50, 60), got %v(%f, %f)", cmd, x, y)
	}

	// Auto-close should generate line back to start
	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdLineTo || x != 10 || y != 20 {
		t.Errorf("Expected auto-close LineTo(10, 20), got %v(%f, %f)", cmd, x, y)
	}

	// Should get EndPoly with close flag
	x, y, cmd = adaptor.Vertex()
	expectedCmd := basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
	if cmd != expectedCmd {
		t.Errorf("Expected EndPoly with close flag %v, got %v", expectedCmd, cmd)
	}
}

func TestConvAdaptorVPGen_ClosedPolygon(t *testing.T) {
	source := NewMockVertexSource([]Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{50, 60, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)},
	})

	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	adaptor.Rewind(0)

	// Process vertices
	vertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, Vertex{x, y, cmd})
	}

	// Should include the closing line
	if len(vertices) < 4 {
		t.Errorf("Expected at least 4 vertices for closed polygon, got %d", len(vertices))
	}

	// Last processed vertex should be LineTo back to start
	lastLine := vertices[len(vertices)-2] // Before EndPoly
	if lastLine.Cmd != basics.PathCmdLineTo || lastLine.X != 10 || lastLine.Y != 20 {
		t.Errorf("Expected closing LineTo(10, 20), got %v(%f, %f)", lastLine.Cmd, lastLine.X, lastLine.Y)
	}
}

func TestConvAdaptorVPGen_EmptyPath(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	adaptor.Rewind(0)

	x, y, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected Stop for empty path, got %v(%f, %f)", cmd, x, y)
	}
}

func TestConvAdaptorVPGen_SinglePoint(t *testing.T) {
	source := NewMockVertexSource([]Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{0, 0, basics.PathCmdEndPoly},
	})

	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	adaptor.Rewind(0)

	// Should process MoveTo
	x, y, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 10 || y != 20 {
		t.Errorf("Expected MoveTo(10, 20), got %v(%f, %f)", cmd, x, y)
	}

	// Should process EndPoly (no auto-close for single point)
	x, y, cmd = adaptor.Vertex()
	if cmd != basics.PathCmdEndPoly {
		t.Errorf("Expected EndPoly, got %v", cmd)
	}
}

func TestConvAdaptorVPGen_AutoUnclose(t *testing.T) {
	source := NewMockVertexSource([]Vertex{
		{10, 20, basics.PathCmdMoveTo},
		{30, 40, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdEndPoly},
	})

	vpgen := NewMockVPGen(false, true) // Auto-unclose enabled
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	adaptor.Rewind(0)

	// Process vertices until stop
	vertices := []Vertex{}
	for {
		x, y, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, Vertex{x, y, cmd})
	}

	// With auto-unclose, EndPoly should be processed differently
	if len(vertices) == 0 {
		t.Error("Expected some vertices to be processed")
	}
}

func TestConvAdaptorVPGen_Attach(t *testing.T) {
	source1 := createSimplePath()
	source2 := NewMockVertexSource([]Vertex{
		{100, 200, basics.PathCmdMoveTo},
		{300, 400, basics.PathCmdLineTo},
	})

	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source1, vpgen)

	// Attach new source
	adaptor.Attach(source2)
	adaptor.Rewind(0)

	// Should process from new source
	x, y, cmd := adaptor.Vertex()
	if cmd != basics.PathCmdMoveTo || x != 100 || y != 200 {
		t.Errorf("Expected MoveTo(100, 200) from new source, got %v(%f, %f)", cmd, x, y)
	}
}

func TestConvAdaptorVPGen_VPGenAccess(t *testing.T) {
	source := createSimplePath()
	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	// Test VPGen access
	retrievedVPGen := adaptor.VPGen()
	if retrievedVPGen != vpgen {
		t.Error("VPGen() should return the same instance")
	}
}

func TestConvAdaptorVPGen_MultipleRewinds(t *testing.T) {
	source := createSimplePath()
	vpgen := NewMockVPGen(false, false)
	adaptor := NewConvAdaptorVPGen(source, vpgen)

	// First pass
	adaptor.Rewind(0)
	count1 := 0
	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		count1++
	}

	// Second pass should yield same results
	adaptor.Rewind(0)
	count2 := 0
	for {
		_, _, cmd := adaptor.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		count2++
	}

	if count1 != count2 {
		t.Errorf("Rewind should reset state: first pass %d vertices, second pass %d vertices", count1, count2)
	}
}
