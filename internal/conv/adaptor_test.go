package conv

import (
	"testing"

	"agg_go/internal/basics"
)

// MockVertexSource for testing
type MockVertexSource struct {
	vertices []Vertex
	index    int
}

type Vertex struct {
	X, Y float64
	Cmd  basics.PathCommand
}

func NewMockVertexSource(vertices []Vertex) *MockVertexSource {
	return &MockVertexSource{vertices: vertices, index: 0}
}

func (m *MockVertexSource) Rewind(pathID uint) {
	m.index = 0
}

func (m *MockVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := m.vertices[m.index]
	m.index++
	return v.X, v.Y, v.Cmd
}

// MockVertexGenerator for testing
type MockVertexGenerator struct {
	vertices []Vertex
	index    int
}

func NewMockVertexGenerator() *MockVertexGenerator {
	return &MockVertexGenerator{
		vertices: make([]Vertex, 0),
		index:    0,
	}
}

func (m *MockVertexGenerator) RemoveAll() {
	m.vertices = m.vertices[:0]
	m.index = 0
}

func (m *MockVertexGenerator) AddVertex(x, y float64, cmd basics.PathCommand) {
	m.vertices = append(m.vertices, Vertex{X: x, Y: y, Cmd: cmd})
}

func (m *MockVertexGenerator) PrepareSrc() {
	// No preparation needed for mock
}

func (m *MockVertexGenerator) Rewind(pathID uint) {
	m.index = 0
}

func (m *MockVertexGenerator) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := m.vertices[m.index]
	m.index++
	return v.X, v.Y, v.Cmd
}

// MockMarkers for testing
type MockMarkers struct {
	vertices []Vertex
	index    int
}

func NewMockMarkers() *MockMarkers {
	return &MockMarkers{
		vertices: make([]Vertex, 0),
		index:    0,
	}
}

func (m *MockMarkers) RemoveAll() {
	m.vertices = m.vertices[:0]
	m.index = 0
}

func (m *MockMarkers) AddVertex(x, y float64, cmd basics.PathCommand) {
	m.vertices = append(m.vertices, Vertex{X: x, Y: y, Cmd: cmd})
}

func (m *MockMarkers) PrepareSrc() {
	// No preparation needed for mock
}

func (m *MockMarkers) Rewind(pathID uint) {
	m.index = 0
}

func (m *MockMarkers) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, basics.PathCmdStop
	}
	v := m.vertices[m.index]
	m.index++
	return v.X, v.Y, v.Cmd
}

func TestConvAdaptorVCGen_Basic(t *testing.T) {
	// Create a simple path
	sourceVertices := []Vertex{
		{X: 10, Y: 20, Cmd: basics.PathCmdMoveTo},
		{X: 30, Y: 40, Cmd: basics.PathCmdLineTo},
		{X: 50, Y: 60, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source, generator)

	// Test basic functionality
	adaptor.Rewind(0)

	// First vertex should trigger generator processing
	_, _, _ = adaptor.Vertex()

	// Generator should have received all vertices except stop
	expectedVertices := 3 // MoveTo, LineTo, LineTo
	if len(generator.vertices) != expectedVertices {
		t.Errorf("Expected generator to receive %d vertices, got %d", expectedVertices, len(generator.vertices))
	}

	// Check that generator received correct vertices
	if generator.vertices[0].X != 10 || generator.vertices[0].Y != 20 || generator.vertices[0].Cmd != basics.PathCmdMoveTo {
		t.Errorf("Generator received incorrect first vertex: %+v", generator.vertices[0])
	}
}

func TestConvAdaptorVCGen_WithMarkers(t *testing.T) {
	sourceVertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	markers := NewMockMarkers()
	adaptor := NewConvAdaptorVCGenWithMarkers(source, generator, markers)

	adaptor.Rewind(0)
	adaptor.Vertex() // Process the path

	// Both generator and markers should receive vertices
	if len(generator.vertices) == 0 {
		t.Error("Generator should have received vertices")
	}
	if len(markers.vertices) == 0 {
		t.Error("Markers should have received vertices")
	}

	// Should receive the same vertices
	if len(generator.vertices) != len(markers.vertices) {
		t.Errorf("Generator and markers should receive same number of vertices, got %d and %d",
			len(generator.vertices), len(markers.vertices))
	}
}

func TestConvAdaptorVCGen_EmptyPath(t *testing.T) {
	sourceVertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source, generator)

	adaptor.Rewind(0)
	x, y, cmd := adaptor.Vertex()

	// Should immediately return stop for empty path
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty path, got %v", cmd)
	}

	if x != 0 || y != 0 {
		t.Errorf("Expected (0,0) for empty path, got (%f,%f)", x, y)
	}
}

func TestConvAdaptorVCGen_MultipleRewinds(t *testing.T) {
	sourceVertices := []Vertex{
		{X: 5, Y: 10, Cmd: basics.PathCmdMoveTo},
		{X: 15, Y: 20, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source, generator)

	// First rewind and process
	adaptor.Rewind(0)
	adaptor.Vertex()
	firstVertexCount := len(generator.vertices)

	// Second rewind should reset and process again
	adaptor.Rewind(0)
	adaptor.Vertex()

	// Generator should be reset and receive vertices again
	if len(generator.vertices) != firstVertexCount {
		t.Errorf("Expected generator to be reset on second rewind, got %d vertices after first, %d after second",
			firstVertexCount, len(generator.vertices))
	}
}

func TestNullMarkers(t *testing.T) {
	markers := &NullMarkers{}

	// All operations should be no-ops and not panic
	markers.RemoveAll()
	markers.AddVertex(10, 20, basics.PathCmdMoveTo)
	markers.PrepareSrc()
	markers.Rewind(0)

	x, y, cmd := markers.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("NullMarkers should always return PathCmdStop, got %v", cmd)
	}
	if x != 0 || y != 0 {
		t.Errorf("NullMarkers should always return (0,0), got (%f,%f)", x, y)
	}
}

func TestConvAdaptorVCGen_AttachNewSource(t *testing.T) {
	sourceVertices1 := []Vertex{
		{X: 1, Y: 1, Cmd: basics.PathCmdMoveTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	sourceVertices2 := []Vertex{
		{X: 2, Y: 2, Cmd: basics.PathCmdMoveTo},
		{X: 4, Y: 4, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source1 := NewMockVertexSource(sourceVertices1)
	source2 := NewMockVertexSource(sourceVertices2)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source1, generator)

	// Process first source
	adaptor.Rewind(0)
	adaptor.Vertex()
	_ = len(generator.vertices)

	// Attach new source
	adaptor.Attach(source2)
	adaptor.Rewind(0)
	adaptor.Vertex()

	// Should have processed the new source
	if len(generator.vertices) != 2 { // MoveTo + LineTo from second source
		t.Errorf("Expected 2 vertices from second source, got %d", len(generator.vertices))
	}

	// Check first vertex from new source
	if generator.vertices[0].X != 2 || generator.vertices[0].Y != 2 {
		t.Errorf("Expected (2,2) from new source, got (%f,%f)", generator.vertices[0].X, generator.vertices[0].Y)
	}
}

func TestConvAdaptorVCGen_AccessorMethods(t *testing.T) {
	source := NewMockVertexSource([]Vertex{})
	generator := NewMockVertexGenerator()
	markers := NewMockMarkers()
	adaptor := NewConvAdaptorVCGenWithMarkers(source, generator, markers)

	// Test accessor methods
	if adaptor.Generator() != generator {
		t.Error("Generator() should return the same generator instance")
	}

	if adaptor.Markers() != markers {
		t.Error("Markers() should return the same markers instance")
	}
}

// Benchmark tests
func BenchmarkConvAdaptorVCGen_SmallPath(b *testing.B) {
	sourceVertices := []Vertex{
		{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo},
		{X: 10, Y: 10, Cmd: basics.PathCmdLineTo},
		{X: 20, Y: 0, Cmd: basics.PathCmdLineTo},
		{X: 0, Y: 0, Cmd: basics.PathCmdStop},
	}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source, generator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0 // Reset source manually for benchmark
		generator.RemoveAll()
		adaptor.Rewind(0)

		for {
			_, _, cmd := adaptor.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkConvAdaptorVCGen_LargePath(b *testing.B) {
	// Create a path with 100 vertices
	sourceVertices := make([]Vertex, 101)
	sourceVertices[0] = Vertex{X: 0, Y: 0, Cmd: basics.PathCmdMoveTo}

	for i := 1; i < 100; i++ {
		sourceVertices[i] = Vertex{X: float64(i), Y: float64(i * i), Cmd: basics.PathCmdLineTo}
	}
	sourceVertices[100] = Vertex{X: 0, Y: 0, Cmd: basics.PathCmdStop}

	source := NewMockVertexSource(sourceVertices)
	generator := NewMockVertexGenerator()
	adaptor := NewConvAdaptorVCGen(source, generator)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.index = 0 // Reset source manually for benchmark
		generator.RemoveAll()
		adaptor.Rewind(0)

		for {
			_, _, cmd := adaptor.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}
