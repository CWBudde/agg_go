package basics

import (
	"testing"
)

// MockVertexSource implements VertexSource for testing
type MockVertexSource struct {
	vertices []Vertex[float64]
	index    int
}

func NewMockVertexSource(vertices []Vertex[float64]) *MockVertexSource {
	return &MockVertexSource{
		vertices: vertices,
		index:    0,
	}
}

func (m *MockVertexSource) Rewind(pathID uint) {
	m.index = 0
}

func (m *MockVertexSource) Vertex() (x, y float64, cmd PathCommand) {
	if m.index >= len(m.vertices) {
		return 0, 0, PathCmdStop
	}

	vertex := m.vertices[m.index]
	m.index++
	return vertex.X, vertex.Y, PathCommand(vertex.Cmd)
}

func TestBoundingRectSingle_SimpleRectangle(t *testing.T) {
	// Create a simple rectangle path
	vertices := []Vertex[float64]{
		{X: 10, Y: 20, Cmd: uint32(PathCmdMoveTo)},
		{X: 30, Y: 20, Cmd: uint32(PathCmdLineTo)},
		{X: 30, Y: 40, Cmd: uint32(PathCmdLineTo)},
		{X: 10, Y: 40, Cmd: uint32(PathCmdLineTo)},
		{X: 10, Y: 20, Cmd: uint32(PathCmdLineTo)}, // Close the rectangle
	}

	vs := NewMockVertexSource(vertices)
	rect, valid := BoundingRectSingle[float64](vs, 0)

	if !valid {
		t.Fatal("Expected valid bounding rectangle")
	}

	expectedRect := Rect[float64]{X1: 10, Y1: 20, X2: 30, Y2: 40}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRectSingle_SinglePoint(t *testing.T) {
	vertices := []Vertex[float64]{
		{X: 15, Y: 25, Cmd: uint32(PathCmdMoveTo)},
	}

	vs := NewMockVertexSource(vertices)
	rect, valid := BoundingRectSingle[float64](vs, 0)

	if !valid {
		t.Fatal("Expected valid bounding rectangle for single point")
	}

	expectedRect := Rect[float64]{X1: 15, Y1: 25, X2: 15, Y2: 25}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRectSingle_EmptyPath(t *testing.T) {
	vertices := []Vertex[float64]{}

	vs := NewMockVertexSource(vertices)
	rect, valid := BoundingRectSingle[float64](vs, 0)

	if valid {
		t.Fatal("Expected invalid bounding rectangle for empty path")
	}

	// Should return initial invalid rectangle (X1 > X2, Y1 > Y2)
	if rect.X1 <= rect.X2 || rect.Y1 <= rect.Y2 {
		t.Errorf("Expected invalid rectangle, got %v", rect)
	}
}

func TestBoundingRectSingle_IgnoreNonVertexCommands(t *testing.T) {
	vertices := []Vertex[float64]{
		{X: 10, Y: 20, Cmd: uint32(PathCmdMoveTo)},
		{X: 30, Y: 40, Cmd: uint32(PathCmdLineTo)},
		{X: 0, Y: 0, Cmd: uint32(PathCmdEndPoly)}, // Should be ignored
	}

	vs := NewMockVertexSource(vertices)
	rect, valid := BoundingRectSingle[float64](vs, 0)

	if !valid {
		t.Fatal("Expected valid bounding rectangle")
	}

	expectedRect := Rect[float64]{X1: 10, Y1: 20, X2: 30, Y2: 40}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRectSingle_IntegerType(t *testing.T) {
	vertices := []Vertex[float64]{
		{X: 10.7, Y: 20.3, Cmd: uint32(PathCmdMoveTo)},
		{X: 30.1, Y: 40.9, Cmd: uint32(PathCmdLineTo)},
	}

	vs := NewMockVertexSource(vertices)
	rect, valid := BoundingRectSingle[int](vs, 0)

	if !valid {
		t.Fatal("Expected valid bounding rectangle")
	}

	expectedRect := Rect[int]{X1: 10, Y1: 20, X2: 30, Y2: 40}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRect_MultiplePaths(t *testing.T) {
	// Create a mock vertex source that returns different paths based on path ID
	vs := &MultiPathVertexSource{
		paths: map[uint][]Vertex[float64]{
			0: {
				{X: 10, Y: 10, Cmd: uint32(PathCmdMoveTo)},
				{X: 20, Y: 20, Cmd: uint32(PathCmdLineTo)},
			},
			1: {
				{X: 30, Y: 30, Cmd: uint32(PathCmdMoveTo)},
				{X: 40, Y: 40, Cmd: uint32(PathCmdLineTo)},
			},
			2: {
				{X: 5, Y: 5, Cmd: uint32(PathCmdMoveTo)},
				{X: 15, Y: 15, Cmd: uint32(PathCmdLineTo)},
			},
		},
	}

	pathIDs := SliceGetID{0, 1, 2}
	rect, valid := BoundingRect[float64](vs, pathIDs, 0, 3)

	if !valid {
		t.Fatal("Expected valid bounding rectangle")
	}

	// Should encompass all three paths: (5,5) to (40,40)
	expectedRect := Rect[float64]{X1: 5, Y1: 5, X2: 40, Y2: 40}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRect_PartialPathRange(t *testing.T) {
	vs := &MultiPathVertexSource{
		paths: map[uint][]Vertex[float64]{
			0: {
				{X: 10, Y: 10, Cmd: uint32(PathCmdMoveTo)},
				{X: 20, Y: 20, Cmd: uint32(PathCmdLineTo)},
			},
			1: {
				{X: 30, Y: 30, Cmd: uint32(PathCmdMoveTo)},
				{X: 40, Y: 40, Cmd: uint32(PathCmdLineTo)},
			},
			2: {
				{X: 5, Y: 5, Cmd: uint32(PathCmdMoveTo)},
				{X: 15, Y: 15, Cmd: uint32(PathCmdLineTo)},
			},
		},
	}

	// Only use paths 1 and 2 (start=1, num=2)
	pathIDs := SliceGetID{0, 1, 2}
	rect, valid := BoundingRect[float64](vs, pathIDs, 1, 2)

	if !valid {
		t.Fatal("Expected valid bounding rectangle")
	}

	// Should encompass paths 1 and 2: (5,5) to (40,40)
	expectedRect := Rect[float64]{X1: 5, Y1: 5, X2: 40, Y2: 40}
	if rect != expectedRect {
		t.Errorf("Expected %v, got %v", expectedRect, rect)
	}
}

func TestBoundingRect_EmptyPathSet(t *testing.T) {
	vs := &MultiPathVertexSource{
		paths: map[uint][]Vertex[float64]{},
	}

	pathIDs := SliceGetID{0, 1}
	rect, valid := BoundingRect[float64](vs, pathIDs, 0, 2)

	if valid {
		t.Fatal("Expected invalid bounding rectangle for empty paths")
	}

	// Should return initial invalid rectangle
	if rect.X1 <= rect.X2 || rect.Y1 <= rect.Y2 {
		t.Errorf("Expected invalid rectangle, got %v", rect)
	}
}

func TestSliceGetID(t *testing.T) {
	pathIDs := SliceGetID{10, 20, 30}

	if pathIDs.Get(0) != 10 {
		t.Errorf("Expected 10, got %v", pathIDs.Get(0))
	}
	if pathIDs.Get(1) != 20 {
		t.Errorf("Expected 20, got %v", pathIDs.Get(1))
	}
	if pathIDs.Get(2) != 30 {
		t.Errorf("Expected 30, got %v", pathIDs.Get(2))
	}

	// Test out of bounds
	if pathIDs.Get(3) != 0 {
		t.Errorf("Expected 0 for out of bounds, got %v", pathIDs.Get(3))
	}
}

// MultiPathVertexSource is a more complex mock for testing multiple paths
type MultiPathVertexSource struct {
	paths       map[uint][]Vertex[float64]
	currentPath uint
	index       int
}

func (m *MultiPathVertexSource) Rewind(pathID uint) {
	m.currentPath = pathID
	m.index = 0
}

func (m *MultiPathVertexSource) Vertex() (x, y float64, cmd PathCommand) {
	vertices, exists := m.paths[m.currentPath]
	if !exists || m.index >= len(vertices) {
		return 0, 0, PathCmdStop
	}

	vertex := vertices[m.index]
	m.index++
	return vertex.X, vertex.Y, PathCommand(vertex.Cmd)
}

// Benchmark tests
func BenchmarkBoundingRectSingle(b *testing.B) {
	vertices := make([]Vertex[float64], 1000)
	for i := range vertices {
		vertices[i] = Vertex[float64]{
			X:   float64(i),
			Y:   float64(i * 2),
			Cmd: uint32(PathCmdLineTo),
		}
	}
	vertices[0].Cmd = uint32(PathCmdMoveTo)

	vs := NewMockVertexSource(vertices)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BoundingRectSingle[float64](vs, 0)
	}
}

func BenchmarkBoundingRect(b *testing.B) {
	vs := &MultiPathVertexSource{
		paths: make(map[uint][]Vertex[float64]),
	}

	// Create 10 paths with 100 vertices each
	for pathID := uint(0); pathID < 10; pathID++ {
		vertices := make([]Vertex[float64], 100)
		for i := range vertices {
			vertices[i] = Vertex[float64]{
				X:   float64(i + int(pathID)*100),
				Y:   float64(i*2 + int(pathID)*200),
				Cmd: uint32(PathCmdLineTo),
			}
		}
		vertices[0].Cmd = uint32(PathCmdMoveTo)
		vs.paths[pathID] = vertices
	}

	pathIDs := make(SliceGetID, 10)
	for i := range pathIDs {
		pathIDs[i] = uint(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BoundingRect[float64](vs, pathIDs, 0, 10)
	}
}
