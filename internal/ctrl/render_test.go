package ctrl

import (
	"testing"

	"agg_go/internal/basics"
)

// Mock implementations for testing

type mockRasterizer struct {
	resetCalled  bool
	addPathCalls []uint
}

func (mr *mockRasterizer) Reset() {
	mr.resetCalled = true
}

func (mr *mockRasterizer) AddPath(vs VertexSourceInterface, pathID uint) {
	mr.addPathCalls = append(mr.addPathCalls, pathID)

	// Consume the vertex source to test the adapter
	vs.Rewind(pathID)
	for {
		x, y, cmd := vs.Vertex()
		_ = x
		_ = y
		if cmd == 0 { // Stop command
			break
		}
	}
}

type mockScanline struct{}

type mockRenderer struct {
	colorCalls []interface{}
}

func (mr *mockRenderer) SetColor(color interface{}) {
	mr.colorCalls = append(mr.colorCalls, color)
}

type mockControl struct {
	*BaseCtrl
	rewindCalls []uint
	vertexIndex int
	vertices    [][3]float64 // x, y, cmd
	colors      []interface{}
}

func newMockControl() *mockControl {
	return &mockControl{
		BaseCtrl: NewBaseCtrl(0, 0, 100, 100, false),
		vertices: [][3]float64{
			{0, 0, float64(basics.PathCmdMoveTo)},
			{100, 0, float64(basics.PathCmdLineTo)},
			{100, 100, float64(basics.PathCmdLineTo)},
			{0, 100, float64(basics.PathCmdLineTo)},
			{0, 0, float64(basics.PathCmdStop)},
		},
		colors: []interface{}{"red", "blue"},
	}
}

func (mc *mockControl) OnMouseButtonDown(x, y float64) bool         { return false }
func (mc *mockControl) OnMouseButtonUp(x, y float64) bool           { return false }
func (mc *mockControl) OnMouseMove(x, y float64, pressed bool) bool { return false }
func (mc *mockControl) OnArrowKeys(left, right, down, up bool) bool { return false }

func (mc *mockControl) NumPaths() uint {
	return uint(len(mc.colors))
}

func (mc *mockControl) Rewind(pathID uint) {
	mc.rewindCalls = append(mc.rewindCalls, pathID)
	mc.vertexIndex = 0
}

func (mc *mockControl) Vertex() (x, y float64, cmd basics.PathCommand) {
	if mc.vertexIndex >= len(mc.vertices) {
		return 0, 0, basics.PathCmdStop
	}

	v := mc.vertices[mc.vertexIndex]
	mc.vertexIndex++
	return v[0], v[1], basics.PathCommand(v[2])
}

func (mc *mockControl) Color(pathID uint) interface{} {
	if pathID < uint(len(mc.colors)) {
		return mc.colors[pathID]
	}
	return "default"
}

func TestCtrlVertexSourceAdapter(t *testing.T) {
	ctrl := newMockControl()
	adapter := &ctrlVertexSourceAdapter{ctrl: ctrl, pathID: 0}

	// Test Rewind
	adapter.Rewind(0)
	if len(ctrl.rewindCalls) != 1 || ctrl.rewindCalls[0] != 0 {
		t.Error("Expected Rewind to be called on control")
	}

	// Test Vertex iteration
	expectedVertices := [][3]float64{
		{0, 0, float64(basics.PathCmdMoveTo)},
		{100, 0, float64(basics.PathCmdLineTo)},
		{100, 100, float64(basics.PathCmdLineTo)},
		{0, 100, float64(basics.PathCmdLineTo)},
		{0, 0, float64(basics.PathCmdStop)},
	}

	for i, expected := range expectedVertices {
		x, y, cmd := adapter.Vertex()
		if x != expected[0] || y != expected[1] || float64(cmd) != expected[2] {
			t.Errorf("Vertex %d: expected (%.1f, %.1f, %.0f), got (%.1f, %.1f, %d)",
				i, expected[0], expected[1], expected[2], x, y, cmd)
		}

		if cmd == 0 { // Stop command
			break
		}
	}
}

func TestRenderCtrlGeneric(t *testing.T) {
	// This test verifies the generic RenderCtrl function structure
	// The actual rendering would require real rasterizer/scanline/renderer implementations

	ctrl := newMockControl()
	ras := &mockRasterizer{}
	_ = &mockScanline{} // sl
	_ = &mockRenderer{} // ren

	// The generic function would be called like this:
	// RenderCtrl(ras, sl, ren, ctrl)

	// For now, manually verify the logic that would be in RenderCtrl
	numPaths := ctrl.NumPaths()

	for i := uint(0); i < numPaths; i++ {
		ras.Reset()
		if !ras.resetCalled {
			t.Error("Expected rasterizer Reset to be called")
		}
		ras.resetCalled = false // Reset for next iteration

		adapter := &ctrlVertexSourceAdapter{ctrl: ctrl, pathID: i}
		ras.AddPath(adapter, i)

		if len(ras.addPathCalls) != int(i+1) {
			t.Errorf("Expected %d AddPath calls, got %d", i+1, len(ras.addPathCalls))
		}

		if ras.addPathCalls[i] != i {
			t.Errorf("Expected AddPath to be called with path %d, got %d", i, ras.addPathCalls[i])
		}

		color := ctrl.Color(i)
		expectedColor := ctrl.colors[i]
		if color != expectedColor {
			t.Errorf("Expected color %v for path %d, got %v", expectedColor, i, color)
		}
	}
}

func TestSimpleRenderCtrl(t *testing.T) {
	ctrl := newMockControl()

	renderCalls := make(map[uint][]Vertex)
	colorCalls := make(map[uint]interface{})

	renderFunc := func(pathID uint, vertices []Vertex, color interface{}) {
		renderCalls[pathID] = vertices
		colorCalls[pathID] = color
	}

	SimpleRenderCtrl(ctrl, renderFunc)

	// Verify that render function was called for each path
	expectedPaths := ctrl.NumPaths()
	if uint(len(renderCalls)) != expectedPaths {
		t.Errorf("Expected render calls for %d paths, got %d", expectedPaths, len(renderCalls))
	}

	// Verify path 0 vertices
	vertices0, ok := renderCalls[0]
	if !ok {
		t.Fatal("Expected render call for path 0")
	}

	if len(vertices0) != 4 { // Should exclude the stop vertex
		t.Errorf("Expected 4 vertices for path 0, got %d", len(vertices0))
	}

	// Check first vertex
	if vertices0[0].X != 0 || vertices0[0].Y != 0 || vertices0[0].Cmd != uint32(basics.PathCmdMoveTo) {
		t.Errorf("Expected first vertex (0, 0, MoveTo), got (%.1f, %.1f, %d)",
			vertices0[0].X, vertices0[0].Y, vertices0[0].Cmd)
	}

	// Verify colors
	if colorCalls[0] != "red" {
		t.Errorf("Expected color 'red' for path 0, got %v", colorCalls[0])
	}
	if colorCalls[1] != "blue" {
		t.Errorf("Expected color 'blue' for path 1, got %v", colorCalls[1])
	}
}

func TestCreateTestRenderFunc(t *testing.T) {
	renderFunc := CreateTestRenderFunc()

	// Test that it doesn't panic with various inputs
	renderFunc(0, []Vertex{{X: 10, Y: 20, Cmd: uint32(basics.PathCmdMoveTo)}}, "red")
	renderFunc(1, []Vertex{}, nil)
	renderFunc(999, nil, struct{}{})

	// If we get here without panicking, the test passes
}

func TestVertexStruct(t *testing.T) {
	vertex := Vertex{
		X:   123.45,
		Y:   678.90,
		Cmd: uint32(basics.PathCmdLineTo),
	}

	if vertex.X != 123.45 {
		t.Errorf("Expected X 123.45, got %.2f", vertex.X)
	}
	if vertex.Y != 678.90 {
		t.Errorf("Expected Y 678.90, got %.2f", vertex.Y)
	}
	if vertex.Cmd != uint32(basics.PathCmdLineTo) {
		t.Errorf("Expected Cmd %d, got %d", uint32(basics.PathCmdLineTo), vertex.Cmd)
	}
}

// Integration test using the mock control
func TestRenderControlIntegration(t *testing.T) {
	ctrl := newMockControl()

	// Test complete rendering pipeline with simple render function
	var totalVertices int
	var totalPaths int

	renderFunc := func(pathID uint, vertices []Vertex, color interface{}) {
		totalPaths++
		totalVertices += len(vertices)

		// Verify that all vertices have valid coordinates
		for i, v := range vertices {
			if v.X < 0 || v.X > 100 || v.Y < 0 || v.Y > 100 {
				t.Errorf("Path %d vertex %d has coordinates outside expected bounds: (%.1f, %.1f)",
					pathID, i, v.X, v.Y)
			}
		}

		// Verify color is not nil
		if color == nil {
			t.Errorf("Path %d has nil color", pathID)
		}
	}

	SimpleRenderCtrl(ctrl, renderFunc)

	if totalPaths != int(ctrl.NumPaths()) {
		t.Errorf("Expected %d paths to be rendered, got %d", ctrl.NumPaths(), totalPaths)
	}

	if totalVertices == 0 {
		t.Error("Expected some vertices to be rendered")
	}
}
