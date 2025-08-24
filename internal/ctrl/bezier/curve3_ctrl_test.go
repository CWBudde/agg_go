package bezier

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewCurve3Ctrl(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test default configuration
	if ctrl.LineWidth() != 1.0 {
		t.Errorf("LineWidth() = %f, want 1.0", ctrl.LineWidth())
	}

	if ctrl.PointRadius() != 5.0 {
		t.Errorf("PointRadius() = %f, want 5.0", ctrl.PointRadius())
	}

	// Test default control points
	expectedPoints := []struct{ x, y float64 }{
		{100.0, 0.0},  // P0
		{100.0, 50.0}, // P1
		{50.0, 100.0}, // P2
	}

	points := []struct{ x, y float64 }{
		{ctrl.X1(), ctrl.Y1()},
		{ctrl.X2(), ctrl.Y2()},
		{ctrl.X3(), ctrl.Y3()},
	}

	for i, expected := range expectedPoints {
		if points[i].x != expected.x || points[i].y != expected.y {
			t.Errorf("Point %d: got (%f, %f), want (%f, %f)",
				i, points[i].x, points[i].y, expected.x, expected.y)
		}
	}
}

func TestCurve3CtrlSetCurve(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Set new curve points
	ctrl.SetCurve(0.0, 0.0, 50.0, 75.0, 100.0, 50.0)

	// Verify points were set
	if ctrl.X1() != 0.0 || ctrl.Y1() != 0.0 {
		t.Errorf("P0: got (%f, %f), want (0.0, 0.0)", ctrl.X1(), ctrl.Y1())
	}
	if ctrl.X2() != 50.0 || ctrl.Y2() != 75.0 {
		t.Errorf("P1: got (%f, %f), want (50.0, 75.0)", ctrl.X2(), ctrl.Y2())
	}
	if ctrl.X3() != 100.0 || ctrl.Y3() != 50.0 {
		t.Errorf("P2: got (%f, %f), want (100.0, 50.0)", ctrl.X3(), ctrl.Y3())
	}
}

func TestCurve3CtrlIndividualPointSetters(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test individual setters
	ctrl.SetX1(10.0)
	ctrl.SetY1(20.0)
	ctrl.SetX2(30.0)
	ctrl.SetY2(40.0)
	ctrl.SetX3(50.0)
	ctrl.SetY3(60.0)

	// Verify all points
	if ctrl.X1() != 10.0 || ctrl.Y1() != 20.0 {
		t.Errorf("P0: got (%f, %f), want (10.0, 20.0)", ctrl.X1(), ctrl.Y1())
	}
	if ctrl.X2() != 30.0 || ctrl.Y2() != 40.0 {
		t.Errorf("P1: got (%f, %f), want (30.0, 40.0)", ctrl.X2(), ctrl.Y2())
	}
	if ctrl.X3() != 50.0 || ctrl.Y3() != 60.0 {
		t.Errorf("P2: got (%f, %f), want (50.0, 60.0)", ctrl.X3(), ctrl.Y3())
	}
}

func TestCurve3CtrlConfiguration(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test line width
	ctrl.SetLineWidth(3.0)
	if w := ctrl.LineWidth(); w != 3.0 {
		t.Errorf("LineWidth() = %f, want 3.0", w)
	}

	// Test point radius
	ctrl.SetPointRadius(8.0)
	if r := ctrl.PointRadius(); r != 8.0 {
		t.Errorf("PointRadius() = %f, want 8.0", r)
	}

	// Test line color
	green := color.NewRGBA(0.0, 1.0, 0.0, 1.0)
	ctrl.SetLineColor(green)
	if clr := ctrl.LineColor(); clr != green {
		t.Errorf("LineColor() = %v, want %v", clr, green)
	}
}

func TestCurve3CtrlCurveAccess(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Set known curve
	ctrl.SetCurve(0.0, 0.0, 50.0, 50.0, 100.0, 0.0)

	// Get curve object
	curve := ctrl.Curve()
	if curve == nil {
		t.Fatal("Curve() returned nil")
	}

	// The curve should be initialized with the control points
	// This is mainly testing that the integration works without errors
}

func TestCurve3CtrlMouseInteraction(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test mouse interaction (delegates to polygon control)
	// This primarily tests that the methods exist and can be called

	// Test clicking near a control point
	result := ctrl.OnMouseButtonDown(100.0, 5.0) // Near P0
	if !result {
		t.Error("Should detect click near control point")
	}

	// Test mouse movement
	result = ctrl.OnMouseMove(105.0, 10.0, true)
	if !result {
		t.Error("Should handle mouse movement")
	}

	// Test mouse up
	result = ctrl.OnMouseButtonUp(105.0, 10.0)
	if !result {
		t.Error("Should handle mouse up")
	}

	// Test arrow keys (requires point selection)
	ctrl.OnMouseButtonDown(100.0, 5.0) // Select point
	result = ctrl.OnArrowKeys(false, true, false, false)
	if !result {
		t.Error("Should handle arrow keys when point is selected")
	}
}

func TestCurve3CtrlVertexGeneration(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test path count
	if numPaths := ctrl.NumPaths(); numPaths != 6 {
		t.Errorf("NumPaths() = %d, want 6", numPaths)
	}

	// Test vertex generation for each path
	for pathID := uint(0); pathID < 6; pathID++ {
		t.Run(fmt.Sprintf("Path %d", pathID), func(t *testing.T) {
			ctrl.Rewind(pathID)

			vertexCount := 0
			for {
				x, y, cmd := ctrl.Vertex()
				if cmd == basics.PathCmdStop {
					break
				}

				vertexCount++
				t.Logf("Path %d, Vertex %d: (%f, %f) cmd=%d", pathID, vertexCount, x, y, cmd)

				// Prevent infinite loop
				if vertexCount > 1000 {
					t.Fatalf("Path %d: Too many vertices generated", pathID)
				}
			}

			if vertexCount == 0 {
				t.Errorf("Path %d: No vertices generated", pathID)
			}

			t.Logf("Path %d: Total vertices generated: %d", pathID, vertexCount)
		})
	}
}

func TestCurve3CtrlColorInterface(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Set a specific color
	purple := color.NewRGBA(1.0, 0.0, 1.0, 1.0)
	ctrl.SetLineColor(purple)

	// Test color interface for all paths
	for pathID := uint(0); pathID < 6; pathID++ {
		clr := ctrl.Color(pathID)
		if clr != purple {
			t.Errorf("Color(%d) = %v, want %v", pathID, clr, purple)
		}
	}
}

func TestCurve3CtrlSpecificPaths(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Set a specific curve for testing
	ctrl.SetCurve(0.0, 0.0, 50.0, 100.0, 100.0, 0.0)

	pathDescriptions := []string{
		"Control line 1 (P0 to P1)",
		"Control line 2 (P1 to P2)",
		"Actual quadratic curve",
		"Point 1 (P0)",
		"Point 2 (P1)",
		"Point 3 (P2)",
	}

	for pathID := uint(0); pathID < 6; pathID++ {
		t.Run(pathDescriptions[pathID], func(t *testing.T) {
			ctrl.Rewind(pathID)

			// Just verify we can get at least one vertex without error
			x, y, cmd := ctrl.Vertex()
			if cmd == basics.PathCmdStop {
				t.Errorf("Path %d (%s): No vertices generated", pathID, pathDescriptions[pathID])
			} else {
				t.Logf("Path %d (%s): First vertex (%f, %f) cmd=%d",
					pathID, pathDescriptions[pathID], x, y, cmd)
			}
		})
	}
}

func TestCurve3CtrlInRect(t *testing.T) {
	ctrl := NewDefaultCurve3Ctrl()

	// Test the inherited InRect method
	// Since BaseCtrl sets bounds to (0,0,1,1), points within should return true
	if !ctrl.InRect(0.5, 0.5) {
		t.Error("InRect(0.5, 0.5) should be true")
	}

	// Points outside should return false
	if ctrl.InRect(-1.0, -1.0) {
		t.Error("InRect(-1.0, -1.0) should be false")
	}
}

func TestCurve3CtrlComparedToBezier(t *testing.T) {
	bezierCtrl := NewDefaultBezierCtrl()
	curve3Ctrl := NewDefaultCurve3Ctrl()

	// Curve3 should have fewer paths than Bezier
	if curve3Ctrl.NumPaths() >= bezierCtrl.NumPaths() {
		t.Errorf("Curve3 paths (%d) should be less than Bezier paths (%d)",
			curve3Ctrl.NumPaths(), bezierCtrl.NumPaths())
	}

	// Both should have similar basic functionality
	if curve3Ctrl.LineWidth() != bezierCtrl.LineWidth() {
		t.Error("Both controls should have same default line width")
	}

	if curve3Ctrl.PointRadius() != bezierCtrl.PointRadius() {
		t.Error("Both controls should have same default point radius")
	}
}
