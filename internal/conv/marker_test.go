package conv

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// MockMarkerLocator provides test marker positions
type MockMarkerLocator struct {
	markers       []struct{ x1, y1, x2, y2 float64 } // Each marker has a position and direction point
	currentMarker int
	vertexCount   int
}

func NewMockMarkerLocator(positions [][4]float64) *MockMarkerLocator {
	markers := make([]struct{ x1, y1, x2, y2 float64 }, len(positions))
	for i, pos := range positions {
		markers[i] = struct{ x1, y1, x2, y2 float64 }{pos[0], pos[1], pos[2], pos[3]}
	}
	return &MockMarkerLocator{
		markers:       markers,
		currentMarker: 0,
		vertexCount:   0,
	}
}

func (m *MockMarkerLocator) Rewind(markerIndex uint) {
	m.currentMarker = int(markerIndex)
	m.vertexCount = 0
}

func (m *MockMarkerLocator) Vertex() (x, y float64, cmd basics.PathCommand) {
	// Process all markers sequentially within a single rewind
	for m.currentMarker < len(m.markers) {
		marker := m.markers[m.currentMarker]
		switch m.vertexCount {
		case 0:
			// Return first point (marker position)
			m.vertexCount++
			return marker.x1, marker.y1, basics.PathCmdMoveTo
		case 1:
			// Return second point (direction)
			m.vertexCount++
			return marker.x2, marker.y2, basics.PathCmdLineTo
		default:
			// After providing two vertices for this marker, move to next marker
			m.currentMarker++
			m.vertexCount = 0
			// Continue loop to process next marker if available
		}
	}
	// No more markers available
	return 0, 0, basics.PathCmdStop
}

// MockMarkerShapes provides test marker geometry
type MockMarkerShapes struct {
	shapes     [][]basics.Point[float64] // Each shape is a slice of points
	shapeIndex int
	pointIndex int
}

func NewMockMarkerShapes(shapes [][]basics.Point[float64]) *MockMarkerShapes {
	return &MockMarkerShapes{
		shapes:     shapes,
		shapeIndex: 0,
		pointIndex: 0,
	}
}

func (m *MockMarkerShapes) Rewind(shapeIndex uint) {
	m.shapeIndex = int(shapeIndex)
	if m.shapeIndex >= len(m.shapes) {
		m.shapeIndex = 0 // Wrap around or use first shape
	}
	m.pointIndex = 0
}

func (m *MockMarkerShapes) Vertex() (x, y float64, cmd basics.PathCommand) {
	if m.shapeIndex >= len(m.shapes) || m.pointIndex >= len(m.shapes[m.shapeIndex]) {
		return 0, 0, basics.PathCmdStop
	}

	point := m.shapes[m.shapeIndex][m.pointIndex]
	m.pointIndex++

	// First point is move_to, rest are line_to
	if m.pointIndex == 1 {
		return point.X, point.Y, basics.PathCmdMoveTo
	}
	return point.X, point.Y, basics.PathCmdLineTo
}

func TestConvMarker_Basic(t *testing.T) {
	// Create a simple marker locator with one marker at origin pointing right
	locator := NewMockMarkerLocator([][4]float64{
		{0, 0, 1, 0}, // Marker at (0,0) pointing to (1,0)
	})

	// Create a simple triangle marker shape
	triangle := []basics.Point[float64]{
		{X: 0, Y: 0},    // Center
		{X: 1, Y: 0.5},  // Right point
		{X: 1, Y: -0.5}, // Right point below
	}
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{triangle})

	// Create the marker converter
	marker := NewConvMarker(locator, shapes)

	// Test that Transform() returns the transformation matrix
	transform := marker.Transform()
	if transform == nil {
		t.Fatal("Transform() returned nil")
	}

	// Test basic rewind and vertex generation
	marker.Rewind(0)

	// Should get the triangle vertices, transformed
	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}
	for i := 0; i < 10; i++ { // Limit to avoid infinite loop
		x, y, cmd := marker.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}
	if len(vertices) < 3 {
		t.Errorf("Expected at least 3 vertices for triangle, got %d", len(vertices))
	}
}

func TestConvMarker_MultipleMarkers(t *testing.T) {
	// Create multiple marker positions
	locator := NewMockMarkerLocator([][4]float64{
		{0, 0, 1, 0},   // First marker at origin pointing right
		{10, 5, 11, 5}, // Second marker at (10,5) pointing right
	})

	// Simple line marker
	line := []basics.Point[float64]{
		{X: 0, Y: 0},
		{X: 2, Y: 0},
	}
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{line, line})

	marker := NewConvMarker(locator, shapes)
	marker.Rewind(0)

	// Collect all vertices to analyze the pattern
	vertices := []struct {
		x, y float64
		cmd  basics.PathCommand
	}{}

	for i := 0; i < 20; i++ { // Limit to avoid infinite loop
		x, y, cmd := marker.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})
	}

	// Should have vertices for both markers
	if len(vertices) < 4 { // At least 2 vertices per marker * 2 markers
		t.Errorf("Expected at least 4 vertices for both markers, got %d", len(vertices))

		// Debug output to understand current behavior
		t.Logf("Vertices received:")
		for i, v := range vertices {
			t.Logf("  %d: (%.2f, %.2f) cmd=%v", i, v.x, v.y, v.cmd)
		}
		return
	}

	// Analyze the pattern - should see position jumps indicating multiple markers
	markerPositions := []struct{ x, y float64 }{}
	for i, v := range vertices {
		if basics.IsMoveTo(v.cmd) {
			// Check if this is a significant position change (new marker)
			if len(markerPositions) == 0 {
				markerPositions = append(markerPositions, struct{ x, y float64 }{v.x, v.y})
			} else {
				// Check distance from last marker
				lastPos := markerPositions[len(markerPositions)-1]
				dist := (v.x-lastPos.x)*(v.x-lastPos.x) + (v.y-lastPos.y)*(v.y-lastPos.y)
				if dist > 1.0 { // Significant distance indicates new marker
					markerPositions = append(markerPositions, struct{ x, y float64 }{v.x, v.y})
				}
			}
		}

		// Debug: log first few vertices
		if i < 6 {
			t.Logf("Vertex %d: (%.2f, %.2f) cmd=%v", i, v.x, v.y, v.cmd)
		}
	}

	// Should detect both marker positions
	expectedMarkers := 2
	if len(markerPositions) != expectedMarkers {
		t.Errorf("Expected %d distinct markers, found %d", expectedMarkers, len(markerPositions))
		t.Logf("Marker positions found:")
		for i, pos := range markerPositions {
			t.Logf("  Marker %d: (%.2f, %.2f)", i+1, pos.x, pos.y)
		}
	}
}

func TestConvMarker_MarkerOrientation(t *testing.T) {
	// Create a marker pointing upward (90 degrees)
	locator := NewMockMarkerLocator([][4]float64{
		{0, 0, 0, 1}, // Marker at origin pointing up
	})

	// Simple horizontal line that should be rotated to vertical
	line := []basics.Point[float64]{
		{X: 0, Y: 0},
		{X: 1, Y: 0}, // Originally horizontal
	}
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{line})

	marker := NewConvMarker(locator, shapes)
	marker.Rewind(0)

	// Skip the first vertex (move_to at origin)
	marker.Vertex()

	// Get the second vertex which should be rotated 90 degrees
	x, y, cmd := marker.Vertex()
	if basics.IsStop(cmd) {
		t.Fatal("Expected a vertex, got stop")
	}

	// The point (1,0) rotated 90 degrees should be approximately (0,1)
	expectedX, expectedY := 0.0, 1.0
	epsilon := 1e-6

	if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
		t.Errorf("Expected marker to be rotated to (~%.3f, ~%.3f), got (%.3f, %.3f)",
			expectedX, expectedY, x, y)
	}
}

func TestConvMarker_TransformMatrix(t *testing.T) {
	locator := NewMockMarkerLocator([][4]float64{
		{0, 0, 1, 0}, // Marker at origin pointing right
	})

	line := []basics.Point[float64]{
		{X: 0, Y: 0}, // Start at origin
		{X: 1, Y: 0}, // Line to (1,0)
	}
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{line})

	marker := NewConvMarker(locator, shapes)

	// Apply a scaling transformation
	marker.Transform().Scale(2.0)

	marker.Rewind(0)

	// Get vertices and test scaling
	marker.Vertex() // Skip first vertex (move_to)
	x, y, cmd := marker.Vertex()
	if basics.IsStop(cmd) {
		t.Fatal("Expected a vertex, got stop")
	}

	// The point should be scaled by 2
	expectedX, expectedY := 2.0, 0.0
	epsilon := 1e-6

	if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
		t.Errorf("Expected scaled point (%.3f, %.3f), got (%.3f, %.3f)",
			expectedX, expectedY, x, y)
	}
}

func TestConvMarker_EmptyMarkers(t *testing.T) {
	// Empty marker locator
	locator := NewMockMarkerLocator([][4]float64{})
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{})

	marker := NewConvMarker(locator, shapes)
	marker.Rewind(0)

	// Should immediately return stop
	_, _, cmd := marker.Vertex()
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop for empty markers, got %v", cmd)
	}
}

func TestConvMarker_StateReset(t *testing.T) {
	locator := NewMockMarkerLocator([][4]float64{
		{0, 0, 1, 0},
	})

	line := []basics.Point[float64]{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
	}
	shapes := NewMockMarkerShapes([][]basics.Point[float64]{line})

	marker := NewConvMarker(locator, shapes)

	// Read some vertices
	marker.Rewind(0)
	marker.Vertex()
	marker.Vertex()

	// Rewind and read again - should get the same sequence
	marker.Rewind(0)
	x1, y1, cmd1 := marker.Vertex()
	_, _, cmd2 := marker.Vertex()

	if cmd1 != basics.PathCmdMoveTo {
		t.Errorf("After rewind, expected PathCmdMoveTo, got %v", cmd1)
	}

	if cmd2 == basics.PathCmdStop {
		t.Error("After rewind, expected vertex but got stop")
	}

	// Values should be consistent
	if x1 != 0 || y1 != 0 {
		t.Errorf("Expected first vertex at origin after rewind, got (%.3f, %.3f)", x1, y1)
	}
}
