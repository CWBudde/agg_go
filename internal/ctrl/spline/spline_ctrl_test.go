package spline

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewSplineCtrlImpl(t *testing.T) {
	// Test basic construction
	ctrl := NewSplineCtrlImpl[color.RGBA](10, 20, 100, 80, 6, false)

	if ctrl == nil {
		t.Fatal("NewSplineCtrlImpl returned nil")
	}

	if ctrl.numPnt != 6 {
		t.Errorf("Expected numPnt=6, got %d", ctrl.numPnt)
	}

	if ctrl.X1() != 10 || ctrl.Y1() != 20 || ctrl.X2() != 100 || ctrl.Y2() != 80 {
		t.Errorf("Incorrect bounds: got (%f,%f,%f,%f), expected (10,20,100,80)",
			ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}

	// Test clamping of point count
	ctrlLow := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 2, false)
	if ctrlLow.numPnt != 4 {
		t.Errorf("Expected numPnt clamped to 4, got %d", ctrlLow.numPnt)
	}

	ctrlHigh := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 50, false)
	if ctrlHigh.numPnt != maxControlPoints {
		t.Errorf("Expected numPnt clamped to %d, got %d", maxControlPoints, ctrlHigh.numPnt)
	}
}

func TestControlPointInitialization(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 8, false)

	// Check initial point distribution
	for i := uint(0); i < ctrl.numPnt; i++ {
		expectedX := float64(i) / float64(ctrl.numPnt-1)
		expectedY := 0.5

		if math.Abs(ctrl.xp[i]-expectedX) > 1e-6 {
			t.Errorf("Point %d X: expected %f, got %f", i, expectedX, ctrl.xp[i])
		}
		if math.Abs(ctrl.yp[i]-expectedY) > 1e-6 {
			t.Errorf("Point %d Y: expected %f, got %f", i, expectedY, ctrl.yp[i])
		}
	}
}

func TestControlPointConstraints(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 5, false)

	// Test first point constraint (should stay at x=0)
	ctrl.setXP(0, 0.5)
	if ctrl.xp[0] != 0.0 {
		t.Errorf("First point X should be constrained to 0, got %f", ctrl.xp[0])
	}

	// Test last point constraint (should stay at x=1)
	ctrl.setXP(4, 0.5)
	if ctrl.xp[4] != 1.0 {
		t.Errorf("Last point X should be constrained to 1, got %f", ctrl.xp[4])
	}

	// Test middle point ordering constraints
	ctrl.setXP(1, 0.1)  // Should work
	ctrl.setXP(2, 0.05) // Should be constrained by previous point
	if ctrl.xp[2] <= ctrl.xp[1] {
		t.Errorf("Point 2 should be constrained by point 1: p1=%f, p2=%f", ctrl.xp[1], ctrl.xp[2])
	}

	// Test Y coordinate clamping
	ctrl.setYP(1, -0.5)
	if ctrl.yp[1] != 0.0 {
		t.Errorf("Y coordinate should be clamped to 0, got %f", ctrl.yp[1])
	}

	ctrl.setYP(1, 1.5)
	if ctrl.yp[1] != 1.0 {
		t.Errorf("Y coordinate should be clamped to 1, got %f", ctrl.yp[1])
	}
}

func TestSplineCalculation(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 4, false)

	// Set up a simple curve: start low, peak in middle, end low
	ctrl.SetPoint(1, 0.33, 0.8)
	ctrl.SetPoint(2, 0.66, 0.2)

	// Test spline value calculation
	value := ctrl.Value(0.5)
	if value < 0.0 || value > 1.0 {
		t.Errorf("Spline value should be in [0,1], got %f", value)
	}

	// Test lookup table generation
	splineValues := ctrl.Spline()
	if len(splineValues) != splineValueCount {
		t.Errorf("Expected %d spline values, got %d", splineValueCount, len(splineValues))
	}

	splineValues8 := ctrl.Spline8()
	if len(splineValues8) != splineValueCount {
		t.Errorf("Expected %d 8-bit spline values, got %d", splineValueCount, len(splineValues8))
	}

	// Verify 8-bit values are correctly scaled
	for i := 0; i < splineValueCount; i++ {
		expected := uint8(splineValues[i] * 255.0)
		if splineValues8[i] != expected {
			t.Errorf("8-bit value %d: expected %d, got %d", i, expected, splineValues8[i])
		}
	}
}

func TestPointAccessors(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 5, false)

	// Test setting and getting points
	ctrl.SetPoint(2, 0.4, 0.7)

	if math.Abs(ctrl.GetPointX(2)-0.4) > 1e-6 {
		t.Errorf("Point X: expected 0.4, got %f", ctrl.GetPointX(2))
	}
	if math.Abs(ctrl.GetPointY(2)-0.7) > 1e-6 {
		t.Errorf("Point Y: expected 0.7, got %f", ctrl.GetPointY(2))
	}

	// Test setting just Y value
	ctrl.SetValue(3, 0.3)
	if math.Abs(ctrl.GetPointY(3)-0.3) > 1e-6 {
		t.Errorf("Point Y after SetValue: expected 0.3, got %f", ctrl.GetPointY(3))
	}

	// Test bounds checking
	if ctrl.GetPointX(100) != 0.0 {
		t.Errorf("Out-of-bounds X should return 0, got %f", ctrl.GetPointX(100))
	}
	if ctrl.GetPointY(100) != 0.0 {
		t.Errorf("Out-of-bounds Y should return 0, got %f", ctrl.GetPointY(100))
	}
}

func TestMouseInteraction(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 4, false)

	// Test mouse down on a control point
	// Point 1 should be at screen coordinates (33.33, 50) approximately
	x := ctrl.calcXP(1)
	y := ctrl.calcYP(1)

	result := ctrl.OnMouseButtonDown(x, y)
	if !result {
		t.Error("Mouse down on control point should return true")
	}
	if ctrl.activePnt != 1 {
		t.Errorf("Active point should be 1, got %d", ctrl.activePnt)
	}
	if ctrl.movePnt != 1 {
		t.Errorf("Move point should be 1, got %d", ctrl.movePnt)
	}

	// Test mouse move while dragging
	newX := x + 10
	newY := y - 10
	result = ctrl.OnMouseMove(newX, newY, true)
	if !result {
		t.Error("Mouse move while dragging should return true")
	}

	// Test mouse up
	result = ctrl.OnMouseButtonUp(newX, newY)
	if !result {
		t.Error("Mouse up after dragging should return true")
	}
	if ctrl.movePnt != -1 {
		t.Errorf("Move point should be -1 after mouse up, got %d", ctrl.movePnt)
	}

	// Test mouse down in empty area
	result = ctrl.OnMouseButtonDown(-50, -50)
	if result {
		t.Error("Mouse down in empty area should return false")
	}
}

func TestKeyboardNavigation(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 4, false)

	// Test with no active point
	result := ctrl.OnArrowKeys(true, false, false, false)
	if result {
		t.Error("Arrow keys with no active point should return false")
	}

	// Set active point and test movement
	ctrl.ActivePoint(2)
	originalX := ctrl.GetPointX(2)
	originalY := ctrl.GetPointY(2)

	// Test left arrow
	result = ctrl.OnArrowKeys(true, false, false, false)
	if !result {
		t.Error("Arrow key movement should return true")
	}
	if ctrl.GetPointX(2) >= originalX {
		t.Error("Left arrow should decrease X coordinate")
	}

	// Test up arrow
	result = ctrl.OnArrowKeys(false, false, false, true)
	if !result {
		t.Error("Arrow key movement should return true")
	}
	if ctrl.GetPointY(2) <= originalY {
		t.Error("Up arrow should increase Y coordinate")
	}
}

func TestVertexSource(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](10, 20, 90, 80, 4, false)

	// Test number of paths
	if ctrl.NumPaths() != numPaths {
		t.Errorf("Expected %d paths, got %d", numPaths, ctrl.NumPaths())
	}

	// Test background path (path 0)
	ctrl.Rewind(0)
	vertexCount := 0
	for {
		x, y, cmd := ctrl.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if vertexCount == 0 && cmd != basics.PathCmdMoveTo {
			t.Error("First vertex should be MoveTo")
		}
		if vertexCount > 0 && cmd != basics.PathCmdLineTo {
			t.Errorf("Non-first vertex should be LineTo, got %d", cmd)
		}
		// Ensure we have valid coordinates
		if x < -1000 || x > 1000 || y < -1000 || y > 1000 {
			t.Errorf("Invalid vertex coordinates: (%f, %f)", x, y)
		}
		vertexCount++
	}
	if vertexCount != 4 {
		t.Errorf("Background path should have 4 vertices, got %d", vertexCount)
	}

	// Test border path (path 1)
	ctrl.Rewind(1)
	vertexCount = 0
	moveToCount := 0
	for {
		x, y, cmd := ctrl.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		if cmd == basics.PathCmdMoveTo {
			moveToCount++
		}
		// Ensure we have valid coordinates
		if x < -1000 || x > 1000 || y < -1000 || y > 1000 {
			t.Errorf("Invalid vertex coordinates: (%f, %f)", x, y)
		}
		vertexCount++
	}
	if vertexCount != 8 {
		t.Errorf("Border path should have 8 vertices, got %d", vertexCount)
	}
	if moveToCount != 2 {
		t.Errorf("Border path should have 2 MoveTo commands, got %d", moveToCount)
	}

	// Test curve path (path 2)
	ctrl.Rewind(2)
	vertexCount = 0
	for {
		x, y, cmd := ctrl.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		// Use the coordinates to ensure they're not optimized away
		_ = x + y
		vertexCount++
		if vertexCount > 1000 { // Prevent infinite loop
			break
		}
	}
	if vertexCount == 0 {
		t.Error("Curve path should have vertices")
	}
}

func TestInRect(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](10, 20, 90, 80, 4, false)

	// Test points inside rectangle
	if !ctrl.InRect(50, 50) {
		t.Error("Point (50,50) should be inside rectangle")
	}

	// Test boundary points
	if !ctrl.InRect(10, 20) {
		t.Error("Top-left corner should be inside rectangle")
	}
	if !ctrl.InRect(90, 80) {
		t.Error("Bottom-right corner should be inside rectangle")
	}

	// Test points outside rectangle
	if ctrl.InRect(5, 50) {
		t.Error("Point (5,50) should be outside rectangle")
	}
	if ctrl.InRect(50, 15) {
		t.Error("Point (50,15) should be outside rectangle")
	}
}

func TestSplineCtrlGeneric(t *testing.T) {
	ctrl := NewSplineCtrl[color.RGBA](0, 0, 100, 100, 5, false)

	if ctrl == nil {
		t.Fatal("NewSplineCtrl returned nil")
	}

	// Test color setting and retrieval
	testColor := color.NewRGBA(0.5, 0.3, 0.8, 1.0)
	ctrl.SetBackgroundColor(testColor)

	retrievedColor := ctrl.Color(0)
	if retrievedColor != testColor {
		t.Errorf("Retrieved color doesn't match set color")
	}
}

func TestSplineCtrlRGBA(t *testing.T) {
	ctrl := NewSplineCtrlRGBA(0, 0, 100, 100, 6, false)

	if ctrl == nil {
		t.Fatal("NewSplineCtrlRGBA returned nil")
	}

	// Test that default colors are set
	// Colors are now direct values, not pointers, so they can't be nil
	// Instead check that the colors have valid values (non-zero alpha)
	bgColor := ctrl.Color(0)
	if bgColor.A == 0 {
		t.Error("Background color should have non-zero alpha")
	}

	borderColor := ctrl.Color(1)
	if borderColor.A == 0 {
		t.Error("Border color should have non-zero alpha")
	}

	// Test all paths have colors with non-zero alpha
	for i := uint(0); i < ctrl.NumPaths(); i++ {
		color := ctrl.Color(i)
		if color.A == 0 {
			t.Errorf("Path %d should have a color with non-zero alpha", i)
		}
	}
}

func TestActivePointManagement(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 5, false)

	// Test initial state
	if ctrl.GetActivePoint() != -1 {
		t.Errorf("Initial active point should be -1, got %d", ctrl.GetActivePoint())
	}

	// Test setting valid active point
	ctrl.ActivePoint(2)
	if ctrl.GetActivePoint() != 2 {
		t.Errorf("Active point should be 2, got %d", ctrl.GetActivePoint())
	}

	// Test setting invalid active point (should be ignored)
	ctrl.ActivePoint(10)
	if ctrl.GetActivePoint() != 2 {
		t.Errorf("Invalid active point should be ignored, still should be 2, got %d", ctrl.GetActivePoint())
	}

	ctrl.ActivePoint(-2)
	if ctrl.GetActivePoint() != 2 {
		t.Errorf("Invalid active point should be ignored, still should be 2, got %d", ctrl.GetActivePoint())
	}
}

func TestBorderAndSizeSettings(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](0, 0, 100, 100, 4, false)

	// Test border width setting
	ctrl.BorderWidth(2.0, 1.0)
	if ctrl.borderWidth != 2.0 {
		t.Errorf("Border width should be 2.0, got %f", ctrl.borderWidth)
	}
	if ctrl.borderExtra != 1.0 {
		t.Errorf("Border extra should be 1.0, got %f", ctrl.borderExtra)
	}

	// Test curve width setting
	ctrl.CurveWidth(1.5)
	if ctrl.curveWidth != 1.5 {
		t.Errorf("Curve width should be 1.5, got %f", ctrl.curveWidth)
	}

	// Test point size setting
	ctrl.PointSize(4.0)
	if ctrl.pointSize != 4.0 {
		t.Errorf("Point size should be 4.0, got %f", ctrl.pointSize)
	}
}

func TestCoordinateConversion(t *testing.T) {
	ctrl := NewSplineCtrlImpl[color.RGBA](10, 20, 90, 80, 4, false)

	// Test coordinate conversion for first point (should be at left edge)
	screenX := ctrl.calcXP(0)
	expectedX := 10 + ctrl.borderWidth // xs1
	if math.Abs(screenX-expectedX) > 1e-6 {
		t.Errorf("First point screen X: expected %f, got %f", expectedX, screenX)
	}

	// Test coordinate conversion for last point (should be at right edge)
	screenX = ctrl.calcXP(ctrl.numPnt - 1)
	expectedX = 90 - ctrl.borderWidth // xs2
	if math.Abs(screenX-expectedX) > 1e-6 {
		t.Errorf("Last point screen X: expected %f, got %f", expectedX, screenX)
	}
}
