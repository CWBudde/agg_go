package ctrl

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/renderer/scanline"
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
	colorCalls []string
}

func (mr *mockRenderer) SetColor(color string) {
	mr.colorCalls = append(mr.colorCalls, color)
}

type mockControl struct {
	*BaseCtrl
	rewindCalls []uint
	vertexIndex int
	vertices    [][3]float64 // x, y, cmd
	colors      []string
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
		colors: []string{"red", "blue"},
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

func (mc *mockControl) Color(pathID uint) string {
	if pathID < uint(len(mc.colors)) {
		return mc.colors[pathID]
	}
	return "default"
}

func TestCtrlVertexSourceAdapter(t *testing.T) {
	ctrl := newMockControl()
	adapter := &ctrlVertexSourceAdapter[string]{ctrl: ctrl, pathID: 0}

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

		adapter := &ctrlVertexSourceAdapter[string]{ctrl: ctrl, pathID: i}
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
	colorCalls := make(map[uint]string)

	renderFunc := func(pathID uint, vertices []Vertex, color string) {
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
	renderFunc := CreateTestRenderFunc[string]()

	// Test that it doesn't panic with various inputs
	renderFunc(0, []Vertex{{X: 10, Y: 20, Cmd: uint32(basics.PathCmdMoveTo)}}, "red")
	renderFunc(1, []Vertex{}, "")
	renderFunc(999, nil, "blue")

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

	renderFunc := func(pathID uint, vertices []Vertex, color string) {
		totalPaths++
		totalVertices += len(vertices)

		// Verify that all vertices have valid coordinates
		for i, v := range vertices {
			if v.X < 0 || v.X > 100 || v.Y < 0 || v.Y > 100 {
				t.Errorf("Path %d vertex %d has coordinates outside expected bounds: (%.1f, %.1f)",
					pathID, i, v.X, v.Y)
			}
		}

		// Verify color is not empty
		if color == "" {
			t.Errorf("Path %d has empty color", pathID)
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

// Full scanline interfaces mocks for comprehensive testing

type mockScanlineRasterizer struct {
	resetCalled          bool
	addPathCalls         []uint
	rewindScanlinesCount int
	sweepScanlineCalls   int
	currentSweepCount    int
	minX, maxX           int
}

func (mr *mockScanlineRasterizer) Reset() {
	mr.resetCalled = true
	// Reset the sweep count for the new path
	mr.currentSweepCount = 0
}

func (mr *mockScanlineRasterizer) AddPath(vs VertexSourceInterface, pathID uint) {
	mr.addPathCalls = append(mr.addPathCalls, pathID)
	// Consume the vertex source
	vs.Rewind(pathID)
	for {
		x, y, cmd := vs.Vertex()
		_, _, _ = x, y, cmd
		if cmd == 0 {
			break
		}
	}
}

func (mr *mockScanlineRasterizer) RewindScanlines() bool {
	mr.rewindScanlinesCount++
	mr.currentSweepCount = 0
	return true
}

func (mr *mockScanlineRasterizer) SweepScanline(sl scanline.ScanlineInterface) bool {
	mr.sweepScanlineCalls++
	mr.currentSweepCount++
	// Return false after processing some scanlines to end the loop for this path
	return mr.currentSweepCount <= 2
}

func (mr *mockScanlineRasterizer) MinX() int {
	return mr.minX
}

func (mr *mockScanlineRasterizer) MaxX() int {
	return mr.maxX
}

// Mock scanline that returns a simple span
type mockScanlineFull struct {
	y        int
	numSpans int
}

func (ms *mockScanlineFull) Y() int {
	return ms.y
}

func (ms *mockScanlineFull) NumSpans() int {
	return ms.numSpans
}

func (ms *mockScanlineFull) Begin() scanline.ScanlineIterator {
	return &mockScanlineIterator{
		span: scanline.SpanData{
			X:      10,
			Len:    20,
			Covers: []basics.Int8u{basics.CoverFull, basics.CoverFull},
		},
	}
}

type mockScanlineIterator struct {
	span    scanline.SpanData
	hasNext bool
}

func (msi *mockScanlineIterator) GetSpan() scanline.SpanData {
	return msi.span
}

func (msi *mockScanlineIterator) Next() bool {
	if msi.hasNext {
		msi.hasNext = false
		return true
	}
	return false
}

// Mock base renderer
type mockBaseRenderer struct {
	blendSolidHspanCalls []blendSolidHspanCall
	blendHlineCalls      []blendHlineCall
	blendColorHspanCalls []blendColorHspanCall
}

type blendSolidHspanCall struct {
	x, y, len int
	color     string
	covers    []basics.Int8u
}

type blendHlineCall struct {
	x, y, x2 int
	color    string
	cover    basics.Int8u
}

type blendColorHspanCall struct {
	x, y, len int
	colors    []string
	covers    []basics.Int8u
	cover     basics.Int8u
}

func (mbr *mockBaseRenderer) BlendSolidHspan(x, y, len int, color string, covers []basics.Int8u) {
	mbr.blendSolidHspanCalls = append(mbr.blendSolidHspanCalls, blendSolidHspanCall{
		x: x, y: y, len: len, color: color, covers: covers,
	})
}

func (mbr *mockBaseRenderer) BlendHline(x, y, x2 int, color string, cover basics.Int8u) {
	mbr.blendHlineCalls = append(mbr.blendHlineCalls, blendHlineCall{
		x: x, y: y, x2: x2, color: color, cover: cover,
	})
}

func (mbr *mockBaseRenderer) BlendColorHspan(x, y, len int, colors []string, covers []basics.Int8u, cover basics.Int8u) {
	mbr.blendColorHspanCalls = append(mbr.blendColorHspanCalls, blendColorHspanCall{
		x: x, y: y, len: len, colors: colors, covers: covers, cover: cover,
	})
}

// Mock renderer with color setting capability
type mockFullRenderer struct {
	currentColor        string
	setColorCalls       []string
	prepareCalled       bool
	renderScanlineCalls int
}

func (mfr *mockFullRenderer) SetColor(color string) {
	mfr.currentColor = color
	mfr.setColorCalls = append(mfr.setColorCalls, color)
}

func (mfr *mockFullRenderer) Prepare() {
	mfr.prepareCalled = true
}

func (mfr *mockFullRenderer) Render(sl scanline.ScanlineInterface) {
	mfr.renderScanlineCalls++
	// Just count the render calls for testing
}

// Test RenderCtrl function with full scanline interfaces
func TestRenderCtrlFull(t *testing.T) {
	ctrl := newMockControl()
	ras := &mockScanlineRasterizer{minX: 0, maxX: 100}
	sl := &mockScanlineFull{y: 50, numSpans: 1}
	ren := &mockBaseRenderer{}

	// Call the RenderCtrl function
	RenderCtrl(ras, sl, ren, ctrl)

	// Verify rasterizer was called correctly for each path
	expectedPaths := int(ctrl.NumPaths())
	if len(ras.addPathCalls) != expectedPaths {
		t.Errorf("Expected %d AddPath calls, got %d", expectedPaths, len(ras.addPathCalls))
	}

	// Verify RewindScanlines was called for each path
	if ras.rewindScanlinesCount != expectedPaths {
		t.Errorf("Expected %d RewindScanlines calls, got %d", expectedPaths, ras.rewindScanlinesCount)
	}

	// Verify AddPath was called with correct path IDs
	for i := 0; i < expectedPaths; i++ {
		if int(ras.addPathCalls[i]) != i {
			t.Errorf("Expected AddPath call %d to have pathID %d, got %d", i, i, ras.addPathCalls[i])
		}
	}

	// Verify renderer was called - it should have been called via RenderScanlinesAASolid
	// The number of calls depends on how many paths were processed and how many scanlines each generated
	if len(ren.blendSolidHspanCalls) == 0 {
		t.Error("Expected BlendSolidHspan to be called from RenderScanlinesAASolid")
	}

	// Since each path generates separate RenderScanlinesAASolid calls,
	// we should see calls for each color. The exact order and number depends on
	// the mock rasterizer implementation, so we just verify that both colors appear.
	colorsUsed := make(map[string]bool)
	for _, call := range ren.blendSolidHspanCalls {
		colorsUsed[call.color] = true
	}

	expectedColors := []string{"red", "blue"}
	for _, expectedColor := range expectedColors {
		if !colorsUsed[expectedColor] {
			t.Errorf("Expected to see color %s in render calls, but it was not found", expectedColor)
		}
	}
}

// Test RenderCtrlRS function with full scanline interfaces
func TestRenderCtrlRSFull(t *testing.T) {
	ctrl := newMockControl()
	ras := &mockScanlineRasterizer{minX: 0, maxX: 100}
	sl := &mockScanlineFull{y: 50, numSpans: 1}
	ren := &mockFullRenderer{}

	// Call the RenderCtrlRS function
	RenderCtrlRS(ras, sl, ren, ctrl)

	expectedPaths := int(ctrl.NumPaths())

	// Verify rasterizer was called correctly
	if ras.rewindScanlinesCount == 0 {
		t.Error("Expected RewindScanlines to be called")
	}

	if ras.sweepScanlineCalls == 0 {
		t.Error("Expected SweepScanline to be called")
	}

	if len(ras.addPathCalls) != expectedPaths {
		t.Errorf("Expected %d AddPath calls, got %d", expectedPaths, len(ras.addPathCalls))
	}

	// Verify renderer SetColor was called correctly
	if len(ren.setColorCalls) != expectedPaths {
		t.Errorf("Expected %d SetColor calls, got %d", expectedPaths, len(ren.setColorCalls))
	}

	// Verify Prepare was called
	if !ren.prepareCalled {
		t.Error("Expected Prepare to be called from RenderScanlines")
	}

	// Verify Render was called
	if ren.renderScanlineCalls == 0 {
		t.Error("Expected Render to be called from RenderScanlines")
	}

	// Verify correct colors were set
	expectedColors := []string{"red", "blue"}
	for i, color := range ren.setColorCalls {
		if color != expectedColors[i] {
			t.Errorf("Expected SetColor(%s) for path %d, got SetColor(%s)", expectedColors[i], i, color)
		}
	}
}
