package rasterizer

import (
	"agg_go/internal/basics"
	"testing"
)

// MockOutlineRenderer is a test implementation of OutlineRenderer
type MockOutlineRenderer struct {
	Operations []string    // Records all operations for verification
	X, Y       int         // Current position
	Color      interface{} // Current line color
}

func NewMockOutlineRenderer() *MockOutlineRenderer {
	return &MockOutlineRenderer{
		Operations: make([]string, 0),
		X:          0,
		Y:          0,
	}
}

func (m *MockOutlineRenderer) MoveTo(x, y int) {
	m.X = x
	m.Y = y
	m.Operations = append(m.Operations, "MoveTo")
}

func (m *MockOutlineRenderer) LineTo(x, y int) {
	m.X = x
	m.Y = y
	m.Operations = append(m.Operations, "LineTo")
}

func (m *MockOutlineRenderer) Coord(c float64) int {
	return int(c * 256) // Standard subpixel scale
}

func (m *MockOutlineRenderer) LineColor(c interface{}) {
	m.Color = c
	m.Operations = append(m.Operations, "LineColor")
}

// MockOutlineVertexSource for testing path operations
type MockOutlineVertexSource struct {
	vertices []Vertex
	index    int
}

type Vertex struct {
	X, Y float64
	Cmd  uint32
}

func NewMockOutlineVertexSource(vertices []Vertex) *MockOutlineVertexSource {
	return &MockOutlineVertexSource{
		vertices: vertices,
		index:    0,
	}
}

func (mvs *MockOutlineVertexSource) Rewind(pathID uint32) {
	mvs.index = 0
}

func (mvs *MockOutlineVertexSource) Vertex(x, y *float64) uint32 {
	if mvs.index >= len(mvs.vertices) {
		return uint32(basics.PathCmdStop)
	}

	v := mvs.vertices[mvs.index]
	*x = v.X
	*y = v.Y
	mvs.index++
	return v.Cmd
}

// MockColorStorage for testing multi-path rendering
type MockColorStorage struct {
	colors []interface{}
}

func (mcs *MockColorStorage) GetColor(index int) interface{} {
	if index >= 0 && index < len(mcs.colors) {
		return mcs.colors[index]
	}
	return nil
}

// MockPathIDStorage for testing multi-path rendering
type MockPathIDStorage struct {
	pathIDs []uint32
}

func (mps *MockPathIDStorage) GetPathID(index int) uint32 {
	if index >= 0 && index < len(mps.pathIDs) {
		return mps.pathIDs[index]
	}
	return 0
}

// MockController for testing control rendering
type MockController struct {
	paths  [][]Vertex
	colors []interface{}
}

func (mc *MockController) NumPaths() int {
	return len(mc.paths)
}

func (mc *MockController) Color(pathIndex int) interface{} {
	if pathIndex >= 0 && pathIndex < len(mc.colors) {
		return mc.colors[pathIndex]
	}
	return nil
}

func (mc *MockController) Rewind(pathID uint32) {
	// Implement based on pathID if needed
}

func (mc *MockController) Vertex(x, y *float64) uint32 {
	// Basic implementation - could be enhanced for specific test cases
	return uint32(basics.PathCmdStop)
}

func TestNewRasterizerOutline(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	if rasterizer.renderer != renderer {
		t.Error("Renderer not properly attached")
	}

	if rasterizer.vertices != 0 {
		t.Error("Initial vertex count should be 0")
	}
}

func TestAttach(t *testing.T) {
	renderer1 := NewMockOutlineRenderer()
	renderer2 := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer1)

	rasterizer.Attach(renderer2)

	if rasterizer.renderer != renderer2 {
		t.Error("Renderer not properly attached")
	}
}

func TestMoveTo(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	rasterizer.MoveTo(100, 200)

	if len(renderer.Operations) != 1 || renderer.Operations[0] != "MoveTo" {
		t.Error("MoveTo operation not recorded")
	}

	if renderer.X != 100 || renderer.Y != 200 {
		t.Error("Coordinates not properly set")
	}

	if rasterizer.vertices != 1 {
		t.Error("Vertex count should be 1 after MoveTo")
	}

	if rasterizer.startX != 100 || rasterizer.startY != 200 {
		t.Error("Start coordinates not properly recorded")
	}
}

func TestLineTo(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	rasterizer.MoveTo(0, 0) // Initialize path
	rasterizer.LineTo(100, 200)

	operations := renderer.Operations
	if len(operations) != 2 || operations[1] != "LineTo" {
		t.Error("LineTo operation not recorded")
	}

	if rasterizer.vertices != 2 {
		t.Error("Vertex count should be 2 after MoveTo + LineTo")
	}
}

func TestMoveToD(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	rasterizer.MoveToD(1.5, 2.5)

	expectedX := int(1.5 * 256) // Using mock renderer's Coord method
	expectedY := int(2.5 * 256)

	if renderer.X != expectedX || renderer.Y != expectedY {
		t.Errorf("MoveToD coordinates not properly converted. Expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, renderer.X, renderer.Y)
	}
}

func TestLineToD(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	rasterizer.MoveTo(0, 0) // Initialize path
	rasterizer.LineToD(1.5, 2.5)

	expectedX := int(1.5 * 256)
	expectedY := int(2.5 * 256)

	if renderer.X != expectedX || renderer.Y != expectedY {
		t.Errorf("LineToD coordinates not properly converted. Expected (%d, %d), got (%d, %d)",
			expectedX, expectedY, renderer.X, renderer.Y)
	}
}

func TestClose(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create a path with more than 2 vertices
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(100, 0)
	rasterizer.LineTo(100, 100)
	rasterizer.Close()

	// Should have: MoveTo, LineTo, LineTo, LineTo (for close)
	if len(renderer.Operations) != 4 {
		t.Errorf("Expected 4 operations, got %d", len(renderer.Operations))
	}

	// Final position should be back at start
	if renderer.X != 0 || renderer.Y != 0 {
		t.Error("Close() should return to starting position")
	}

	if rasterizer.vertices != 0 {
		t.Error("Vertex count should be reset to 0 after Close()")
	}
}

func TestCloseWithTwoVertices(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create a path with only 2 vertices
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(100, 0)
	rasterizer.Close()

	// Should have: MoveTo, LineTo (no additional LineTo for close)
	if len(renderer.Operations) != 2 {
		t.Errorf("Expected 2 operations for 2-vertex path, got %d", len(renderer.Operations))
	}
}

func TestAddVertex(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Test MoveTo command
	rasterizer.AddVertex(100.0, 200.0, uint32(basics.PathCmdMoveTo))

	expectedX := int(100.0 * 256)
	expectedY := int(200.0 * 256)

	if renderer.X != expectedX || renderer.Y != expectedY {
		t.Error("AddVertex MoveTo not properly handled")
	}

	// Test LineTo command
	rasterizer.AddVertex(150.0, 250.0, uint32(basics.PathCmdLineTo))

	expectedX = int(150.0 * 256)
	expectedY = int(250.0 * 256)

	if renderer.X != expectedX || renderer.Y != expectedY {
		t.Error("AddVertex LineTo not properly handled")
	}
}

func TestAddVertexEndPoly(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create a path
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(100, 0)
	rasterizer.LineTo(100, 100)

	// Test EndPoly with close flag
	cmd := uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)
	rasterizer.AddVertex(0.0, 0.0, cmd)

	// Should have added a LineTo back to start
	if len(renderer.Operations) != 4 {
		t.Errorf("Expected 4 operations after EndPoly close, got %d", len(renderer.Operations))
	}
}

func TestAddPath(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create a mock vertex source with a simple path
	vertices := []Vertex{
		{10.0, 20.0, uint32(basics.PathCmdMoveTo)},
		{30.0, 40.0, uint32(basics.PathCmdLineTo)},
		{50.0, 60.0, uint32(basics.PathCmdLineTo)},
		{0.0, 0.0, uint32(basics.PathCmdStop)},
	}

	vs := NewMockOutlineVertexSource(vertices)
	rasterizer.AddPath(vs, 0)

	// Should have MoveTo + 2 LineTo operations
	if len(renderer.Operations) != 3 {
		t.Errorf("Expected 3 operations from AddPath, got %d", len(renderer.Operations))
	}

	expectedOps := []string{"MoveTo", "LineTo", "LineTo"}
	for i, op := range expectedOps {
		if i >= len(renderer.Operations) || renderer.Operations[i] != op {
			t.Errorf("Expected operation %d to be %s, got %s", i, op, renderer.Operations[i])
		}
	}
}

func TestRenderAllPaths(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create mock vertex source (simple implementation for this test)
	vertices := []Vertex{
		{10.0, 20.0, uint32(basics.PathCmdMoveTo)},
		{30.0, 40.0, uint32(basics.PathCmdLineTo)},
		{0.0, 0.0, uint32(basics.PathCmdStop)},
	}
	vs := NewMockOutlineVertexSource(vertices)

	// Create color and path ID storage
	colors := &MockColorStorage{colors: []interface{}{"red", "blue"}}
	pathIDs := &MockPathIDStorage{pathIDs: []uint32{0, 1}}

	rasterizer.RenderAllPaths(vs, colors, pathIDs, 2)

	// Should have 2 LineColor operations
	colorCount := 0
	for _, op := range renderer.Operations {
		if op == "LineColor" {
			colorCount++
		}
	}

	if colorCount != 2 {
		t.Errorf("Expected 2 LineColor operations, got %d", colorCount)
	}
}

func TestRenderCtrl(t *testing.T) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	// Create a mock controller with 2 paths
	paths := [][]Vertex{
		{{10.0, 20.0, uint32(basics.PathCmdMoveTo)}, {30.0, 40.0, uint32(basics.PathCmdLineTo)}},
		{{50.0, 60.0, uint32(basics.PathCmdMoveTo)}, {70.0, 80.0, uint32(basics.PathCmdLineTo)}},
	}
	colors := []interface{}{"red", "blue"}

	ctrl := &MockController{paths: paths, colors: colors}
	rasterizer.RenderCtrl(ctrl)

	// Should have 2 LineColor operations (one for each path)
	colorCount := 0
	for _, op := range renderer.Operations {
		if op == "LineColor" {
			colorCount++
		}
	}

	if colorCount != 2 {
		t.Errorf("Expected 2 LineColor operations for 2 paths, got %d", colorCount)
	}
}

// Benchmark tests
func BenchmarkMoveTo(b *testing.B) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rasterizer.MoveTo(i%1000, i%1000)
	}
}

func BenchmarkLineTo(b *testing.B) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	rasterizer.MoveTo(0, 0) // Initialize path

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rasterizer.LineTo(i%1000, i%1000)
	}
}

func BenchmarkAddVertex(b *testing.B) {
	renderer := NewMockOutlineRenderer()
	rasterizer := NewRasterizerOutline(renderer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := uint32(basics.PathCmdLineTo)
		if i%100 == 0 {
			cmd = uint32(basics.PathCmdMoveTo)
		}
		rasterizer.AddVertex(float64(i%1000), float64(i%1000), cmd)
	}
}
