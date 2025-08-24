package bezier

import (
	"fmt"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewBezierCtrl(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

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
		{0.0, 100.0},  // P3
	}

	points := []struct{ x, y float64 }{
		{ctrl.X1(), ctrl.Y1()},
		{ctrl.X2(), ctrl.Y2()},
		{ctrl.X3(), ctrl.Y3()},
		{ctrl.X4(), ctrl.Y4()},
	}

	for i, expected := range expectedPoints {
		if points[i].x != expected.x || points[i].y != expected.y {
			t.Errorf("Point %d: got (%f, %f), want (%f, %f)",
				i, points[i].x, points[i].y, expected.x, expected.y)
		}
	}
}

func TestBezierCtrlSetCurve(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Set new curve points
	ctrl.SetCurve(0.0, 0.0, 25.0, 75.0, 75.0, 25.0, 100.0, 100.0)

	// Verify points were set
	if ctrl.X1() != 0.0 || ctrl.Y1() != 0.0 {
		t.Errorf("P0: got (%f, %f), want (0.0, 0.0)", ctrl.X1(), ctrl.Y1())
	}
	if ctrl.X2() != 25.0 || ctrl.Y2() != 75.0 {
		t.Errorf("P1: got (%f, %f), want (25.0, 75.0)", ctrl.X2(), ctrl.Y2())
	}
	if ctrl.X3() != 75.0 || ctrl.Y3() != 25.0 {
		t.Errorf("P2: got (%f, %f), want (75.0, 25.0)", ctrl.X3(), ctrl.Y3())
	}
	if ctrl.X4() != 100.0 || ctrl.Y4() != 100.0 {
		t.Errorf("P3: got (%f, %f), want (100.0, 100.0)", ctrl.X4(), ctrl.Y4())
	}
}

func TestBezierCtrlIndividualPointSetters(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Test individual setters
	ctrl.SetX1(10.0)
	ctrl.SetY1(20.0)
	ctrl.SetX2(30.0)
	ctrl.SetY2(40.0)
	ctrl.SetX3(50.0)
	ctrl.SetY3(60.0)
	ctrl.SetX4(70.0)
	ctrl.SetY4(80.0)

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
	if ctrl.X4() != 70.0 || ctrl.Y4() != 80.0 {
		t.Errorf("P3: got (%f, %f), want (70.0, 80.0)", ctrl.X4(), ctrl.Y4())
	}
}

func TestBezierCtrlConfiguration(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

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

	// Test line color
	red := color.NewRGBA(1.0, 0.0, 0.0, 1.0)
	ctrl.SetLineColor(red)
	if clr := ctrl.LineColor(); clr != red {
		t.Errorf("LineColor() = %v, want %v", clr, red)
	}
}

func TestBezierCtrlCurveAccess(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Set known curve
	ctrl.SetCurve(0.0, 0.0, 33.33, 33.33, 66.66, 66.66, 100.0, 100.0)

	// Get curve object
	curve := ctrl.Curve()
	if curve == nil {
		t.Fatal("Curve() returned nil")
	}

	// The curve should be initialized with the control points
	// This is mainly testing that the integration works without errors
}

func TestBezierCtrlMouseInteraction(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

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
	result = ctrl.OnArrowKeys(true, false, false, false)
	if !result {
		t.Error("Should handle arrow keys when point is selected")
	}

	// Test clicking far from any control point
	result = ctrl.OnMouseButtonDown(200.0, 200.0)
	if result {
		t.Error("Should not detect click far from any control point")
	}
}

func TestBezierCtrlVertexGeneration(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Test path count
	if numPaths := ctrl.NumPaths(); numPaths != 7 {
		t.Errorf("NumPaths() = %d, want 7", numPaths)
	}

	// Test vertex generation for each path
	for pathID := uint(0); pathID < 7; pathID++ {
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

func TestBezierCtrlColorInterface(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Set a specific color
	blue := color.NewRGBA(0.0, 0.0, 1.0, 1.0)
	ctrl.SetLineColor(blue)

	// Test color interface for all paths
	for pathID := uint(0); pathID < 7; pathID++ {
		clr := ctrl.Color(pathID)
		if clr != blue {
			t.Errorf("Color(%d) = %v, want %v", pathID, clr, blue)
		}
	}
}

func TestBezierCtrlInRect(t *testing.T) {
	ctrl := NewDefaultBezierCtrl()

	// Test the inherited InRect method
	// Since BaseCtrl sets bounds to (0,0,1,1), points within should return true
	if !ctrl.InRect(0.5, 0.5) {
		t.Error("InRect(0.5, 0.5) should be true")
	}

	// Points outside should return false
	if ctrl.InRect(2.0, 2.0) {
		t.Error("InRect(2.0, 2.0) should be false")
	}
}
