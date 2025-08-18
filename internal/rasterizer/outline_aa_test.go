package rasterizer

import (
	"agg_go/internal/primitives"
	"testing"
)

// MockOutlineAARenderer implements OutlineAARenderer for testing
type MockOutlineAARenderer struct {
	AccurateJoinOnlyVal bool
	ColorCalls          []interface{}
	Line0Calls          []primitives.LineParameters
	Line1Calls          []struct {
		LP primitives.LineParameters
		SX int
		SY int
	}
	Line2Calls []struct {
		LP primitives.LineParameters
		EX int
		EY int
	}
	Line3Calls []struct {
		LP     primitives.LineParameters
		SX, SY int
		EX, EY int
	}
	PieCalls []struct {
		X, Y   int
		X1, Y1 int
		X2, Y2 int
	}
	SemidotCalls []struct {
		Cmp    func(int) bool
		X, Y   int
		X1, Y1 int
	}
}

func NewMockOutlineAARenderer() *MockOutlineAARenderer {
	return &MockOutlineAARenderer{
		AccurateJoinOnlyVal: false,
		ColorCalls:          make([]interface{}, 0),
		Line0Calls:          make([]primitives.LineParameters, 0),
		Line1Calls: make([]struct {
			LP primitives.LineParameters
			SX int
			SY int
		}, 0),
		Line2Calls: make([]struct {
			LP primitives.LineParameters
			EX int
			EY int
		}, 0),
		Line3Calls: make([]struct {
			LP     primitives.LineParameters
			SX, SY int
			EX, EY int
		}, 0),
		PieCalls: make([]struct {
			X, Y   int
			X1, Y1 int
			X2, Y2 int
		}, 0),
		SemidotCalls: make([]struct {
			Cmp    func(int) bool
			X, Y   int
			X1, Y1 int
		}, 0),
	}
}

func (m *MockOutlineAARenderer) AccurateJoinOnly() bool {
	return m.AccurateJoinOnlyVal
}

func (m *MockOutlineAARenderer) Color(c interface{}) {
	m.ColorCalls = append(m.ColorCalls, c)
}

func (m *MockOutlineAARenderer) Line0(lp primitives.LineParameters) {
	m.Line0Calls = append(m.Line0Calls, lp)
}

func (m *MockOutlineAARenderer) Line1(lp primitives.LineParameters, sx, sy int) {
	m.Line1Calls = append(m.Line1Calls, struct {
		LP primitives.LineParameters
		SX int
		SY int
	}{lp, sx, sy})
}

func (m *MockOutlineAARenderer) Line2(lp primitives.LineParameters, ex, ey int) {
	m.Line2Calls = append(m.Line2Calls, struct {
		LP primitives.LineParameters
		EX int
		EY int
	}{lp, ex, ey})
}

func (m *MockOutlineAARenderer) Line3(lp primitives.LineParameters, sx, sy, ex, ey int) {
	m.Line3Calls = append(m.Line3Calls, struct {
		LP     primitives.LineParameters
		SX, SY int
		EX, EY int
	}{lp, sx, sy, ex, ey})
}

func (m *MockOutlineAARenderer) Pie(x, y, x1, y1, x2, y2 int) {
	m.PieCalls = append(m.PieCalls, struct {
		X, Y   int
		X1, Y1 int
		X2, Y2 int
	}{x, y, x1, y1, x2, y2})
}

func (m *MockOutlineAARenderer) Semidot(cmp func(int) bool, x, y, x1, y1 int) {
	m.SemidotCalls = append(m.SemidotCalls, struct {
		Cmp    func(int) bool
		X, Y   int
		X1, Y1 int
	}{cmp, x, y, x1, y1})
}

func TestNewRasterizerOutlineAA(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	if rasterizer == nil {
		t.Fatalf("NewRasterizerOutlineAA returned nil")
	}

	// Should use round join by default for normal renderers
	if rasterizer.GetLineJoin() != OutlineRoundJoin {
		t.Errorf("Default line join = %v, want %v", rasterizer.GetLineJoin(), OutlineRoundJoin)
	}

	// Test with accurate join only renderer
	renderer.AccurateJoinOnlyVal = true
	rasterizer2 := NewRasterizerOutlineAA(renderer)
	if rasterizer2.GetLineJoin() != OutlineMiterAccurateJoin {
		t.Errorf("Line join for accurate-only renderer = %v, want %v",
			rasterizer2.GetLineJoin(), OutlineMiterAccurateJoin)
	}
}

func TestRasterizerOutlineAAAttach(t *testing.T) {
	renderer1 := NewMockOutlineAARenderer()
	renderer2 := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer1)

	rasterizer.Attach(renderer2)
	// We can't directly test the internal renderer, but we can test behavior
	// If we add functionality that exposes the renderer, we could test it
}

func TestRasterizerOutlineAALineJoin(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Test setting different join types
	rasterizer.SetLineJoin(OutlineMiterJoin)
	if rasterizer.GetLineJoin() != OutlineMiterJoin {
		t.Errorf("SetLineJoin(OutlineMiterJoin) result = %v, want %v",
			rasterizer.GetLineJoin(), OutlineMiterJoin)
	}

	rasterizer.SetLineJoin(OutlineNoJoin)
	if rasterizer.GetLineJoin() != OutlineNoJoin {
		t.Errorf("SetLineJoin(OutlineNoJoin) result = %v, want %v",
			rasterizer.GetLineJoin(), OutlineNoJoin)
	}

	// Test with accurate join only renderer
	renderer.AccurateJoinOnlyVal = true
	rasterizer.SetLineJoin(OutlineRoundJoin)
	if rasterizer.GetLineJoin() != OutlineMiterAccurateJoin {
		t.Errorf("SetLineJoin with accurate-only renderer = %v, want %v",
			rasterizer.GetLineJoin(), OutlineMiterAccurateJoin)
	}
}

func TestRasterizerOutlineAARoundCap(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Default should be false
	if rasterizer.GetRoundCap() {
		t.Errorf("Default round cap = true, want false")
	}

	rasterizer.SetRoundCap(true)
	if !rasterizer.GetRoundCap() {
		t.Errorf("SetRoundCap(true) result = false, want true")
	}

	rasterizer.SetRoundCap(false)
	if rasterizer.GetRoundCap() {
		t.Errorf("SetRoundCap(false) result = true, want false")
	}
}

func TestRasterizerOutlineAAMoveTo(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	rasterizer.MoveTo(100, 200)

	// We can't directly inspect internal state, but we can test that it doesn't crash
	// and that subsequent operations work correctly
}

func TestRasterizerOutlineAALineTo(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(100, 100)

	// Test that no errors occur during line addition
}

func TestRasterizerOutlineAAMoveToD(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	rasterizer.MoveToD(1.5, 2.5)

	// Should convert to subpixel coordinates internally
}

func TestRasterizerOutlineAALineToD(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	rasterizer.MoveToD(0.0, 0.0)
	rasterizer.LineToD(1.5, 2.5)

	// Should convert to subpixel coordinates internally
}

func TestRasterizerOutlineAARenderTwoVertices(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Add two vertices and render
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 1000)
	rasterizer.Render(false)

	// Should call Line3 method
	if len(renderer.Line3Calls) == 0 {
		t.Errorf("Expected Line3 to be called for two-vertex path")
	}
}

func TestRasterizerOutlineAARenderTwoVerticesWithRoundCap(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)
	rasterizer.SetRoundCap(true)

	// Add two vertices and render
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 1000)
	rasterizer.Render(false)

	// Should call Line3 and Semidot methods
	if len(renderer.Line3Calls) == 0 {
		t.Errorf("Expected Line3 to be called for two-vertex path with round caps")
	}
	if len(renderer.SemidotCalls) != 2 {
		t.Errorf("Expected 2 Semidot calls for round caps, got %d", len(renderer.SemidotCalls))
	}
}

func TestRasterizerOutlineAARenderThreeVertices(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Add three vertices and render
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 0)
	rasterizer.LineTo(1000, 1000)
	rasterizer.Render(false)

	// Should call Line3 method multiple times
	if len(renderer.Line3Calls) < 2 {
		t.Errorf("Expected at least 2 Line3 calls for three-vertex path, got %d", len(renderer.Line3Calls))
	}
}

func TestRasterizerOutlineAARenderThreeVerticesRoundJoin(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)
	rasterizer.SetLineJoin(OutlineRoundJoin)

	// Add three vertices and render
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 0)
	rasterizer.LineTo(1000, 1000)
	rasterizer.Render(false)

	// Should call Line3 and Pie methods for round join
	if len(renderer.Line3Calls) < 2 {
		t.Errorf("Expected at least 2 Line3 calls, got %d", len(renderer.Line3Calls))
	}
	if len(renderer.PieCalls) == 0 {
		t.Errorf("Expected Pie calls for round join")
	}
}

func TestRasterizerOutlineAARenderMultipleVertices(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Add multiple vertices forming a rectangle
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 0)
	rasterizer.LineTo(1000, 1000)
	rasterizer.LineTo(0, 1000)
	rasterizer.Render(false)

	// Should make multiple rendering calls
	totalCalls := len(renderer.Line0Calls) + len(renderer.Line1Calls) +
		len(renderer.Line2Calls) + len(renderer.Line3Calls)
	if totalCalls == 0 {
		t.Errorf("Expected rendering calls for multiple-vertex path")
	}
}

func TestRasterizerOutlineAARenderClosed(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Add vertices forming a triangle
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(1000, 0)
	rasterizer.LineTo(500, 1000)
	rasterizer.Render(true) // Close the polygon

	// Should make rendering calls for closed polygon
	totalCalls := len(renderer.Line0Calls) + len(renderer.Line1Calls) +
		len(renderer.Line2Calls) + len(renderer.Line3Calls)
	if totalCalls == 0 {
		t.Errorf("Expected rendering calls for closed polygon")
	}
}

func TestOutlineAAJoinConstants(t *testing.T) {
	// Test that join constants have expected values
	if OutlineNoJoin != 0 {
		t.Errorf("OutlineNoJoin = %d, want 0", OutlineNoJoin)
	}
	if OutlineMiterJoin != 1 {
		t.Errorf("OutlineMiterJoin = %d, want 1", OutlineMiterJoin)
	}
	if OutlineRoundJoin != 2 {
		t.Errorf("OutlineRoundJoin = %d, want 2", OutlineRoundJoin)
	}
	if OutlineMiterAccurateJoin != 3 {
		t.Errorf("OutlineMiterAccurateJoin = %d, want 3", OutlineMiterAccurateJoin)
	}
}

// Mock vertex source for testing AddPath
type MockVertexSourceAA struct {
	Vertices []struct {
		X, Y float64
		Cmd  uint32
	}
	Index int
}

func NewMockVertexSourceAA() *MockVertexSourceAA {
	return &MockVertexSourceAA{
		Vertices: make([]struct {
			X, Y float64
			Cmd  uint32
		}, 0),
		Index: 0,
	}
}

func (m *MockVertexSourceAA) AddVertex(x, y float64, cmd uint32) {
	m.Vertices = append(m.Vertices, struct {
		X, Y float64
		Cmd  uint32
	}{x, y, cmd})
}

func (m *MockVertexSourceAA) Rewind(pathID uint32) {
	m.Index = 0
}

func (m *MockVertexSourceAA) Vertex(x, y *float64) uint32 {
	if m.Index >= len(m.Vertices) {
		return 0 // Stop command
	}

	vertex := m.Vertices[m.Index]
	*x = vertex.X
	*y = vertex.Y
	cmd := vertex.Cmd
	m.Index++

	return cmd
}

func TestRasterizerOutlineAAAddPath(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	// Create a mock vertex source with a simple path
	vs := NewMockVertexSourceAA()
	vs.AddVertex(0, 0, 1)     // MoveTo
	vs.AddVertex(100, 0, 2)   // LineTo
	vs.AddVertex(100, 100, 2) // LineTo
	vs.AddVertex(0, 0, 0)     // Stop

	rasterizer.AddPath(vs, 0)

	// Should have processed the path without errors
	// The exact rendering calls depend on the path complexity
}

// Mock color storage for testing
type MockColorStorageAA struct {
	Colors []interface{}
}

func (m *MockColorStorageAA) GetColor(index int) interface{} {
	if index >= 0 && index < len(m.Colors) {
		return m.Colors[index]
	}
	return "default"
}

// Mock path ID storage for testing
type MockPathIDStorageAA struct {
	PathIDs []uint32
}

func (m *MockPathIDStorageAA) GetPathID(index int) uint32 {
	if index >= 0 && index < len(m.PathIDs) {
		return m.PathIDs[index]
	}
	return 0
}

func TestRasterizerOutlineAARenderAllPaths(t *testing.T) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	vs := NewMockVertexSourceAA()
	vs.AddVertex(0, 0, 1)     // MoveTo
	vs.AddVertex(100, 100, 2) // LineTo
	vs.AddVertex(0, 0, 0)     // Stop

	colors := &MockColorStorageAA{Colors: []interface{}{"red", "blue"}}
	pathIDs := &MockPathIDStorageAA{PathIDs: []uint32{0, 1}}

	rasterizer.RenderAllPaths(vs, colors, pathIDs, 2)

	// Should have called Color twice
	if len(renderer.ColorCalls) != 2 {
		t.Errorf("Expected 2 color calls, got %d", len(renderer.ColorCalls))
	}
}

func BenchmarkRasterizerOutlineAAMoveTo(b *testing.B) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rasterizer.MoveTo(i, i)
	}
}

func BenchmarkRasterizerOutlineAALineTo(b *testing.B) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	rasterizer.MoveTo(0, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rasterizer.LineTo(i, i)
	}
}

func BenchmarkRasterizerOutlineAARenderTwoVertices(b *testing.B) {
	renderer := NewMockOutlineAARenderer()
	rasterizer := NewRasterizerOutlineAA(renderer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rasterizer.MoveTo(0, 0)
		rasterizer.LineTo(100, 100)
		rasterizer.Render(false)
	}
}
