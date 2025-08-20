package polygon

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewPolygonCtrl(t *testing.T) {
	tests := []struct {
		name        string
		numPoints   uint
		pointRadius float64
		expected    float64
	}{
		{"default radius", 4, 5.0, 5.0},
		{"custom radius", 3, 7.5, 7.5},
		{"zero radius should default", 5, 0.0, 5.0},
		{"negative radius should default", 4, -1.0, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPolygonCtrl(tt.numPoints, tt.pointRadius)

			if ctrl.NumPoints() != tt.numPoints {
				t.Errorf("NumPoints() = %d, want %d", ctrl.NumPoints(), tt.numPoints)
			}

			if ctrl.PointRadius() != tt.expected {
				t.Errorf("PointRadius() = %f, want %f", ctrl.PointRadius(), tt.expected)
			}

			// Check default states
			if !ctrl.InPolygonCheck() {
				t.Error("InPolygonCheck() should be true by default")
			}

			if !ctrl.Close() {
				t.Error("Close() should be true by default")
			}

			if ctrl.LineWidth() != 1.0 {
				t.Errorf("LineWidth() = %f, want 1.0", ctrl.LineWidth())
			}
		})
	}
}

func TestPolygonCoordinates(t *testing.T) {
	ctrl := NewPolygonCtrl(3, 5.0)

	// Test setting coordinates
	coords := []struct{ x, y float64 }{
		{10.0, 20.0},
		{30.0, 40.0},
		{50.0, 60.0},
	}

	for i, coord := range coords {
		ctrl.SetXn(uint(i), coord.x)
		ctrl.SetYn(uint(i), coord.y)
	}

	// Test getting coordinates
	for i, coord := range coords {
		if x := ctrl.Xn(uint(i)); x != coord.x {
			t.Errorf("Xn(%d) = %f, want %f", i, x, coord.x)
		}
		if y := ctrl.Yn(uint(i)); y != coord.y {
			t.Errorf("Yn(%d) = %f, want %f", i, y, coord.y)
		}
	}

	// Test out of bounds access
	if x := ctrl.Xn(10); x != 0.0 {
		t.Errorf("Xn(10) = %f, want 0.0 for out of bounds", x)
	}
	if y := ctrl.Yn(10); y != 0.0 {
		t.Errorf("Yn(10) = %f, want 0.0 for out of bounds", y)
	}

	// Test out of bounds setting (should not panic)
	ctrl.SetXn(10, 100.0)
	ctrl.SetYn(10, 200.0)
}

func TestPolygonConfiguration(t *testing.T) {
	ctrl := NewPolygonCtrl(4, 5.0)

	// Test line width
	ctrl.SetLineWidth(2.5)
	if w := ctrl.LineWidth(); w != 2.5 {
		t.Errorf("LineWidth() = %f, want 2.5", w)
	}

	// Test point radius
	ctrl.SetPointRadius(7.0)
	if r := ctrl.PointRadius(); r != 7.0 {
		t.Errorf("PointRadius() = %f, want 7.0", r)
	}

	// Test polygon check
	ctrl.SetInPolygonCheck(false)
	if ctrl.InPolygonCheck() {
		t.Error("InPolygonCheck() should be false after setting")
	}

	// Test close
	ctrl.SetClose(false)
	if ctrl.Close() {
		t.Error("Close() should be false after setting")
	}

	// Test line color
	red := color.NewRGBA(1.0, 0.0, 0.0, 1.0)
	ctrl.SetLineColor(red)
	if clr := ctrl.LineColor(); clr != red {
		t.Errorf("LineColor() = %v, want %v", clr, red)
	}
}

func TestPolygonMouseInteraction(t *testing.T) {
	ctrl := NewPolygonCtrl(3, 10.0) // Large radius for easier testing

	// Set up triangle
	ctrl.SetXn(0, 0.0)
	ctrl.SetYn(0, 0.0)
	ctrl.SetXn(1, 100.0)
	ctrl.SetYn(1, 0.0)
	ctrl.SetXn(2, 50.0)
	ctrl.SetYn(2, 86.6) // Approximately equilateral triangle

	// Test clicking on a point
	if !ctrl.OnMouseButtonDown(5.0, 5.0) { // Near point 0
		t.Error("Should detect click on point 0")
	}

	// Test moving the point
	if !ctrl.OnMouseMove(15.0, 15.0, true) {
		t.Error("Should handle point movement")
	}

	// Check that point moved (accounting for drag offset)
	if x := ctrl.Xn(0); math.Abs(x-10.0) > 0.1 { // 15 - 5 (offset)
		t.Errorf("Point 0 X = %f, expected around 10.0", x)
	}
	if y := ctrl.Yn(0); math.Abs(y-10.0) > 0.1 { // 15 - 5 (offset)
		t.Errorf("Point 0 Y = %f, expected around 10.0", y)
	}

	// Test mouse up
	if !ctrl.OnMouseButtonUp(15.0, 15.0) {
		t.Error("Should handle mouse up after dragging")
	}

	// Test clicking on empty space
	if ctrl.OnMouseButtonDown(200.0, 200.0) {
		t.Error("Should not detect click on empty space")
	}
}

func TestPointInPolygon(t *testing.T) {
	ctrl := NewPolygonCtrl(4, 5.0)

	// Set up square
	ctrl.SetXn(0, 0.0)
	ctrl.SetYn(0, 0.0)
	ctrl.SetXn(1, 100.0)
	ctrl.SetYn(1, 0.0)
	ctrl.SetXn(2, 100.0)
	ctrl.SetYn(2, 100.0)
	ctrl.SetXn(3, 0.0)
	ctrl.SetYn(3, 100.0)

	// Test point inside polygon
	if !ctrl.pointInPolygon(50.0, 50.0) {
		t.Error("Point (50,50) should be inside square")
	}

	// Test point outside polygon
	if ctrl.pointInPolygon(150.0, 150.0) {
		t.Error("Point (150,150) should be outside square")
	}

	// Test point on edge (may vary depending on algorithm)
	// This test is less strict due to floating point precision
	inside := ctrl.pointInPolygon(0.0, 50.0)
	t.Logf("Point on edge result: %v", inside)
}

func TestVertexGeneration(t *testing.T) {
	ctrl := NewPolygonCtrl(3, 5.0)

	// Set up triangle
	ctrl.SetXn(0, 0.0)
	ctrl.SetYn(0, 0.0)
	ctrl.SetXn(1, 100.0)
	ctrl.SetYn(1, 0.0)
	ctrl.SetXn(2, 50.0)
	ctrl.SetYn(2, 86.6)

	// Test path count
	if numPaths := ctrl.NumPaths(); numPaths != 1 {
		t.Errorf("NumPaths() = %d, want 1", numPaths)
	}

	// Test vertex generation
	ctrl.Rewind(0)

	vertexCount := 0
	validVertices := 0
	controlPointVertices := 0

	for {
		x, y, cmd := ctrl.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}

		vertexCount++

		// Count valid (non-NaN, non-Inf) vertices
		if !math.IsNaN(x) && !math.IsNaN(y) && !math.IsInf(x, 0) && !math.IsInf(y, 0) {
			validVertices++

			// Control points start around vertex 15+ based on pattern
			if vertexCount > 10 && (cmd == basics.PathCmdMoveTo ||
				(cmd == basics.PathCmdLineTo && vertexCount > 14)) {
				controlPointVertices++
			}
		}

		// Prevent infinite loop
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	if vertexCount == 0 {
		t.Error("No vertices generated")
	}

	// We expect some valid vertices (control points)
	if validVertices == 0 {
		t.Error("No valid vertices generated")
	}

	// We should have control point vertices (ellipses for each polygon point)
	if controlPointVertices == 0 {
		t.Error("No control point vertices found")
	}

	t.Logf("Total vertices: %d, Valid vertices: %d, Control point vertices: %d",
		vertexCount, validVertices, controlPointVertices)
}

func TestCheckEdge(t *testing.T) {
	ctrl := NewPolygonCtrl(3, 10.0)

	// Set up triangle
	ctrl.SetXn(0, 0.0)
	ctrl.SetYn(0, 0.0)
	ctrl.SetXn(1, 100.0)
	ctrl.SetYn(1, 0.0)
	ctrl.SetXn(2, 50.0)
	ctrl.SetYn(2, 86.6)

	// Test point near edge 0 (from point 0 to point 2)
	if !ctrl.checkEdge(0, 25.0, 43.3) { // Midpoint of edge from (0,0) to (50,86.6)
		t.Error("Should detect point near edge 0")
	}

	// Test point far from edge
	if ctrl.checkEdge(0, 200.0, 200.0) {
		t.Error("Should not detect point far from edge")
	}
}

func TestArrowKeys(t *testing.T) {
	ctrl := NewPolygonCtrl(3, 5.0)

	// Set initial position
	ctrl.SetXn(0, 50.0)
	ctrl.SetYn(0, 50.0)

	// Simulate mouse click to select point 0
	ctrl.OnMouseButtonDown(50.0, 50.0)

	// Test arrow key movement
	if !ctrl.OnArrowKeys(true, false, false, false) { // left
		t.Error("Should handle left arrow key")
	}

	if x := ctrl.Xn(0); x != 49.0 {
		t.Errorf("After left arrow, X = %f, want 49.0", x)
	}

	if !ctrl.OnArrowKeys(false, true, false, false) { // right
		t.Error("Should handle right arrow key")
	}

	if x := ctrl.Xn(0); x != 50.0 {
		t.Errorf("After right arrow, X = %f, want 50.0", x)
	}

	if !ctrl.OnArrowKeys(false, false, true, false) { // down
		t.Error("Should handle down arrow key")
	}

	if y := ctrl.Yn(0); y != 51.0 {
		t.Errorf("After down arrow, Y = %f, want 51.0", y)
	}

	if !ctrl.OnArrowKeys(false, false, false, true) { // up
		t.Error("Should handle up arrow key")
	}

	if y := ctrl.Yn(0); y != 50.0 {
		t.Errorf("After up arrow, Y = %f, want 50.0", y)
	}

	// Test with no point selected
	ctrl.OnMouseButtonUp(50.0, 50.0) // Release selection
	if ctrl.OnArrowKeys(true, false, false, false) {
		t.Error("Should not handle arrow keys with no selection")
	}
}

func TestPolygonControlAgainstCPPBehavior(t *testing.T) {
	// Test that our implementation matches the C++ AGG behavior for core functionality
	ctrl := NewPolygonCtrl(4, 10.0)

	// Set up a square like in C++ tests
	ctrl.SetXn(0, 100.0)
	ctrl.SetYn(0, 100.0)
	ctrl.SetXn(1, 200.0)
	ctrl.SetYn(1, 100.0)
	ctrl.SetXn(2, 200.0)
	ctrl.SetYn(2, 200.0)
	ctrl.SetXn(3, 100.0)
	ctrl.SetYn(3, 200.0)

	// Test point-in-polygon (matches C++ crossings multiply algorithm)
	if !ctrl.pointInPolygon(150.0, 150.0) {
		t.Error("Point (150,150) should be inside square")
	}

	if ctrl.pointInPolygon(50.0, 50.0) {
		t.Error("Point (50,50) should be outside square")
	}

	// Test edge detection (matches C++ check_edge implementation)
	// Edge 1 goes from point 0 to point 1 (bottom edge of square)
	if !ctrl.checkEdge(1, 150.0, 100.0) { // Point on bottom edge
		t.Error("Should detect point on bottom edge")
	}

	// Test mouse interaction follows C++ behavior
	// Mouse down on point 0
	if !ctrl.OnMouseButtonDown(105.0, 105.0) { // Near point 0 with some tolerance
		t.Error("Should detect mouse down on point 0")
	}

	// Drag the point
	if !ctrl.OnMouseMove(115.0, 115.0, true) {
		t.Error("Should handle point dragging")
	}

	// Check point moved correctly (accounting for drag offset)
	if math.Abs(ctrl.Xn(0)-110.0) > 1.0 || math.Abs(ctrl.Yn(0)-110.0) > 1.0 {
		t.Errorf("Point 0 moved to (%f, %f), expected around (110, 110)",
			ctrl.Xn(0), ctrl.Yn(0))
	}

	// Test whole polygon movement (node == num_points)
	ctrl.OnMouseButtonUp(115.0, 115.0) // Release point

	// Click inside polygon to select whole polygon
	if !ctrl.OnMouseButtonDown(150.0, 150.0) {
		t.Error("Should detect click inside polygon for whole polygon selection")
	}

	originalPositions := make([]float64, ctrl.NumPoints()*2)
	for i := uint(0); i < ctrl.NumPoints(); i++ {
		originalPositions[i*2] = ctrl.Xn(i)
		originalPositions[i*2+1] = ctrl.Yn(i)
	}

	// Move whole polygon
	if !ctrl.OnMouseMove(160.0, 160.0, true) {
		t.Error("Should handle whole polygon movement")
	}

	// Check all points moved by the same amount
	dx, dy := 10.0, 10.0 // Expected movement
	for i := uint(0); i < ctrl.NumPoints(); i++ {
		expectedX := originalPositions[i*2] + dx
		expectedY := originalPositions[i*2+1] + dy

		if math.Abs(ctrl.Xn(i)-expectedX) > 1.0 ||
			math.Abs(ctrl.Yn(i)-expectedY) > 1.0 {
			t.Errorf("Point %d moved to (%f, %f), expected (%f, %f)",
				i, ctrl.Xn(i), ctrl.Yn(i), expectedX, expectedY)
		}
	}
}
