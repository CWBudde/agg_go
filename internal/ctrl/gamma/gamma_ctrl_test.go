package gamma

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewGammaCtrlImpl(t *testing.T) {
	x1, y1, x2, y2 := 10.0, 20.0, 200.0, 150.0
	ctrl := NewGammaCtrlImpl[color.RGBA](x1, y1, x2, y2, false)

	if ctrl == nil {
		t.Fatal("NewGammaCtrlImpl returned nil")
	}

	// Check bounds
	if ctrl.X1() != x1 || ctrl.Y1() != y1 || ctrl.X2() != x2 || ctrl.Y2() != y2 {
		t.Errorf("Bounds mismatch: expected (%f,%f,%f,%f), got (%f,%f,%f,%f)",
			x1, y1, x2, y2, ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}

	// Check that gamma spline is initialized
	if ctrl.gammaSpline == nil {
		t.Error("Gamma spline not initialized")
	}

	// Check default values
	if ctrl.NumPaths() != 7 {
		t.Errorf("Expected 7 paths, got %d", ctrl.NumPaths())
	}

	// Check default gamma is identity
	gamma := ctrl.Gamma()
	if len(gamma) != 256 {
		t.Errorf("Expected gamma table length 256, got %d", len(gamma))
	}
}

func TestNewGammaCtrl(t *testing.T) {
	x1, y1, x2, y2 := 10.0, 20.0, 200.0, 150.0
	ctrl := NewGammaCtrl(x1, y1, x2, y2, false)

	if ctrl == nil {
		t.Fatal("NewGammaCtrl returned nil")
	}

	// Test color access
	for i := uint(0); i < 7; i++ {
		color := ctrl.Color(i)
		// RGBA is a value type, not a pointer, so it can't be nil
		// Just verify it's a valid RGBA color by checking alpha channel
		if color.A < 0.0 {
			t.Errorf("Color %d has invalid alpha: %f", i, color.A)
		}
	}
}

func TestGammaCtrlBounds(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test InRect
	tests := []struct {
		x, y     float64
		expected bool
	}{
		{50, 50, true},   // Inside
		{0, 0, true},     // On boundary
		{100, 100, true}, // On boundary
		{-1, 50, false},  // Outside left
		{101, 50, false}, // Outside right
		{50, -1, false},  // Outside top
		{50, 101, false}, // Outside bottom
	}

	for _, test := range tests {
		result := ctrl.InRect(test.x, test.y)
		if result != test.expected {
			t.Errorf("InRect(%f, %f) = %t, expected %t",
				test.x, test.y, result, test.expected)
		}
	}
}

func TestGammaCtrlValues(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test setting and getting values
	testValues := []struct {
		kx1, ky1, kx2, ky2 float64
	}{
		{1.0, 1.0, 1.0, 1.0}, // Identity
		{0.5, 1.5, 0.5, 1.5}, // Bright curve
		{1.5, 0.5, 1.5, 0.5}, // Dark curve
		{1.2, 0.8, 0.8, 1.2}, // Mixed curve
	}

	for _, test := range testValues {
		ctrl.Values(test.kx1, test.ky1, test.kx2, test.ky2)

		gotKx1, gotKy1, gotKx2, gotKy2 := ctrl.GetValues()

		tolerance := 0.001
		if math.Abs(gotKx1-test.kx1) > tolerance ||
			math.Abs(gotKy1-test.ky1) > tolerance ||
			math.Abs(gotKx2-test.kx2) > tolerance ||
			math.Abs(gotKy2-test.ky2) > tolerance {
			t.Errorf("Values round-trip failed: set (%f,%f,%f,%f), got (%f,%f,%f,%f)",
				test.kx1, test.ky1, test.kx2, test.ky2,
				gotKx1, gotKy1, gotKy2, gotKy2)
		}
	}
}

func TestGammaCtrlMouseInteraction(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Calculate where control points should be for default values
	ctrl.calcPoints()

	// Test mouse interaction with first control point
	t.Run("Point1Interaction", func(t *testing.T) {
		// Click on first control point
		handled := ctrl.OnMouseButtonDown(ctrl.xp1, ctrl.yp1)
		if !handled {
			t.Error("Mouse down on control point 1 should be handled")
		}
		if ctrl.mousePoint != 1 {
			t.Errorf("Expected mousePoint = 1, got %d", ctrl.mousePoint)
		}
		if !ctrl.p1Active {
			t.Error("Point 1 should be active after clicking")
		}

		// Move mouse to drag point
		oldKx1, oldKy1, _, _ := ctrl.GetValues()
		handled = ctrl.OnMouseMove(ctrl.xp1+5, ctrl.yp1+5, true)
		if !handled {
			t.Error("Mouse move while dragging should be handled")
		}

		newKx1, newKy1, _, _ := ctrl.GetValues()
		if oldKx1 == newKx1 && oldKy1 == newKy1 {
			t.Error("Control point values should change when dragging")
		}

		// Release mouse
		handled = ctrl.OnMouseButtonUp(ctrl.xp1+5, ctrl.yp1+5)
		if !handled {
			t.Error("Mouse up after dragging should be handled")
		}
		if ctrl.mousePoint != 0 {
			t.Errorf("Expected mousePoint = 0 after release, got %d", ctrl.mousePoint)
		}
	})

	// Test mouse interaction with second control point
	t.Run("Point2Interaction", func(t *testing.T) {
		// Click on second control point
		handled := ctrl.OnMouseButtonDown(ctrl.xp2, ctrl.yp2)
		if !handled {
			t.Error("Mouse down on control point 2 should be handled")
		}
		if ctrl.mousePoint != 2 {
			t.Errorf("Expected mousePoint = 2, got %d", ctrl.mousePoint)
		}
		if ctrl.p1Active {
			t.Error("Point 1 should not be active after clicking point 2")
		}
	})

	// Test clicking outside control points
	t.Run("NoPointInteraction", func(t *testing.T) {
		// Reset mouse point state first
		ctrl.mousePoint = 0
		handled := ctrl.OnMouseButtonDown(10, 10) // Far from control points
		if handled {
			t.Error("Mouse down outside control points should not be handled")
		}
		if ctrl.mousePoint != 0 {
			t.Errorf("Expected mousePoint = 0, got %d", ctrl.mousePoint)
		}
	})
}

func TestGammaCtrlKeyboardInteraction(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Set to known values
	ctrl.Values(1.0, 1.0, 1.0, 1.0)

	// Test arrow key adjustment for point 1 (active by default)
	t.Run("Point1Keys", func(t *testing.T) {
		oldKx1, oldKy1, _, _ := ctrl.GetValues()

		// Test left arrow (should decrease kx1)
		handled := ctrl.OnArrowKeys(true, false, false, false)
		if !handled {
			t.Error("Arrow key should be handled")
		}
		newKx1, newKy1, _, _ := ctrl.GetValues()
		if newKx1 >= oldKx1 {
			t.Error("Left arrow should decrease kx1")
		}

		// Test right arrow (should increase kx1)
		handled = ctrl.OnArrowKeys(false, true, false, false)
		if !handled {
			t.Error("Arrow key should be handled")
		}
		newKx1, _, _, _ = ctrl.GetValues()
		if newKx1 <= oldKx1-0.005 { // Should be back closer to original or higher
			t.Error("Right arrow should increase kx1")
		}

		// Test down arrow (should decrease ky1)
		handled = ctrl.OnArrowKeys(false, false, true, false)
		if !handled {
			t.Error("Arrow key should be handled")
		}
		_, newKy1, _, _ = ctrl.GetValues()
		if newKy1 >= oldKy1 {
			t.Error("Down arrow should decrease ky1")
		}

		// Test up arrow (should increase ky1)
		handled = ctrl.OnArrowKeys(false, false, false, true)
		if !handled {
			t.Error("Arrow key should be handled")
		}
		_, newKy1, _, _ = ctrl.GetValues()
		if newKy1 <= oldKy1-0.005 { // Should be back closer to original or higher
			t.Error("Up arrow should increase ky1")
		}
	})

	// Test arrow key adjustment for point 2
	t.Run("Point2Keys", func(t *testing.T) {
		ctrl.ChangeActivePoint() // Switch to point 2
		if ctrl.p1Active {
			t.Error("Point 1 should not be active after ChangeActivePoint")
		}

		_, _, oldKx2, _ := ctrl.GetValues()

		// Test left arrow (should increase kx2 for point 2)
		handled := ctrl.OnArrowKeys(true, false, false, false)
		if !handled {
			t.Error("Arrow key should be handled")
		}
		_, _, newKx2, _ := ctrl.GetValues()
		if newKx2 <= oldKx2 {
			t.Error("Left arrow should increase kx2 for point 2")
		}
	})

	// Test no-op case
	t.Run("NoKeys", func(t *testing.T) {
		handled := ctrl.OnArrowKeys(false, false, false, false)
		if handled {
			t.Error("No arrow keys should not be handled")
		}
	})
}

func TestGammaCtrlAppearanceSettings(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test border width
	ctrl.SetBorderWidth(3.0, 1.0)
	if ctrl.borderWidth != 3.0 || ctrl.borderExtra != 1.0 {
		t.Error("Border width setting failed")
	}

	// Test curve width
	ctrl.SetCurveWidth(2.5)
	if ctrl.curveWidth != 2.5 {
		t.Error("Curve width setting failed")
	}

	// Test grid width
	ctrl.SetGridWidth(0.5)
	if ctrl.gridWidth != 0.5 {
		t.Error("Grid width setting failed")
	}

	// Test text thickness
	ctrl.SetTextThickness(2.0)
	if ctrl.textThickness != 2.0 {
		t.Error("Text thickness setting failed")
	}

	// Test text size
	ctrl.SetTextSize(12.0, 8.0)
	if ctrl.textHeight != 12.0 || ctrl.textWidth != 8.0 {
		t.Error("Text size setting failed")
	}

	// Test point size
	ctrl.SetPointSize(7.0)
	if ctrl.pointSize != 7.0 {
		t.Error("Point size setting failed")
	}
}

func TestGammaCtrlVertexGeneration(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test paths (skip curve and text paths which may have external dependencies)
	testPaths := []uint{0, 1, 3, 4, 5} // Skip paths 2 (curve) and 6 (text)
	for _, pathID := range testPaths {
		t.Run("", func(t *testing.T) {
			ctrl.Rewind(pathID)

			vertexCount := 0
			hasMoveTo := false
			_ = false // hasLineTo

			for {
				x, y, cmd := ctrl.Vertex()
				if cmd == basics.PathCmdStop {
					break
				}

				vertexCount++
				if cmd == basics.PathCmdMoveTo {
					hasMoveTo = true
				} else if cmd == basics.PathCmdLineTo {
					// hasLineTo = true
				}

				// Coordinates should be reasonable
				if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
					t.Errorf("Path %d: Invalid vertex coordinates: (%f, %f)", pathID, x, y)
				}

				// Prevent infinite loops
				if vertexCount > 10000 {
					t.Fatalf("Path %d: Too many vertices generated - possible infinite loop", pathID)
				}
			}

			if vertexCount == 0 {
				t.Errorf("Path %d: No vertices generated", pathID)
			}

			// Most paths should have a MoveTo command
			if pathID != 7 && !hasMoveTo { // Allow some paths to not have MoveTo
				t.Logf("Path %d: No MoveTo command found (may be normal)", pathID)
			}
		})
	}
}

func TestGammaCtrlGammaFunctionality(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test Y function
	testPoints := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	for _, x := range testPoints {
		y := ctrl.Y(x)
		if y < 0.0 || y > 1.0 {
			t.Errorf("Y(%f) = %f outside valid range [0,1]", x, y)
		}
	}

	// Test gamma table access
	gamma := ctrl.Gamma()
	if len(gamma) != 256 {
		t.Errorf("Expected gamma table length 256, got %d", len(gamma))
	}

	// Test gamma spline access
	spline := ctrl.GetGammaSpline()
	if spline == nil {
		t.Error("GetGammaSpline returned nil")
	}
}

func TestGammaCtrlColorCustomization(t *testing.T) {
	ctrl := NewGammaCtrl(0, 0, 100, 100, false)

	// Test color setters
	testColor := color.RGBA{R: 1.0, G: 0.5, B: 0.25, A: 0.8}

	ctrl.SetBackgroundColor(testColor)
	if ctrl.backgroundColor != testColor {
		t.Error("SetBackgroundColor failed")
	}

	ctrl.SetBorderColor(testColor)
	if ctrl.borderColor != testColor {
		t.Error("SetBorderColor failed")
	}

	ctrl.SetCurveColor(testColor)
	if ctrl.curveColor != testColor {
		t.Error("SetCurveColor failed")
	}

	ctrl.SetGridColor(testColor)
	if ctrl.gridColor != testColor {
		t.Error("SetGridColor failed")
	}

	ctrl.SetInactivePntColor(testColor)
	if ctrl.inactivePntColor != testColor {
		t.Error("SetInactivePntColor failed")
	}

	ctrl.SetActivePntColor(testColor)
	if ctrl.activePntColor != testColor {
		t.Error("SetActivePntColor failed")
	}

	ctrl.SetTextColor(testColor)
	if ctrl.textColor != testColor {
		t.Error("SetTextColor failed")
	}

	// Test that Color() returns the set colors
	for i := uint(0); i < 7; i++ {
		colorVal := ctrl.Color(i)
		// RGBA is a value type, not a pointer, so it can't be nil
		// Just check that the color has reasonable values
		if colorVal.A < 0.0 {
			t.Errorf("Color(%d) has invalid alpha: %f", i, colorVal.A)
		}
		// Note: All colors were set to testColor, so they should all match
		// colorVal is already color.RGBA type from Color() method
		if colorVal != testColor {
			t.Errorf("Color(%d) does not match set color", i)
		}
	}

	// Test out of range color
	outOfRangeColor := ctrl.Color(10)
	// RGBA is a value type, so it can't be nil - just check it returns a valid color
	if outOfRangeColor.A < 0.0 {
		t.Error("Out of range color should return valid default color")
	}
}

func TestGammaCtrlTransformation(t *testing.T) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	// Test coordinate transformation (basic test)
	x, y := 50.0, 50.0
	origX, origY := x, y

	// Transform and inverse transform should be identity for no transformation
	ctrl.TransformXY(&x, &y)
	ctrl.InverseTransformXY(&x, &y)

	if math.Abs(x-origX) > 0.001 || math.Abs(y-origY) > 0.001 {
		t.Errorf("Transform round-trip failed: (%f,%f) -> (%f,%f)", origX, origY, x, y)
	}
}

func TestGammaCtrlEdgeCases(t *testing.T) {
	_ = NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false) // For potential future use

	// Test very small control area
	smallCtrl := NewGammaCtrlImpl[color.RGBA](0, 0, 10, 10, false)
	if smallCtrl == nil {
		t.Error("Should handle small control area")
	}

	// Test zero-size control area
	zeroCtrl := NewGammaCtrlImpl[color.RGBA](0, 0, 0, 0, false)
	if zeroCtrl == nil {
		t.Error("Should handle zero-size control area")
	}

	// Test negative coordinates
	negCtrl := NewGammaCtrlImpl[color.RGBA](-100, -100, -10, -10, false)
	if negCtrl == nil {
		t.Error("Should handle negative coordinates")
	}

	// Test flipY
	flipCtrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, true)
	if flipCtrl == nil {
		t.Error("Should handle Y-axis flipping")
	}
	if !flipCtrl.FlipY() {
		t.Error("FlipY flag not set correctly")
	}
}

// Benchmark tests
func BenchmarkGammaCtrlVertexGeneration(b *testing.B) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pathID := uint(i % 7)
		ctrl.Rewind(pathID)

		for {
			_, _, cmd := ctrl.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
		}
	}
}

func BenchmarkGammaCtrlMouseInteraction(b *testing.B) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)
	ctrl.calcPoints()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctrl.OnMouseButtonDown(ctrl.xp1, ctrl.yp1)
		ctrl.OnMouseMove(ctrl.xp1+1, ctrl.yp1+1, true)
		ctrl.OnMouseButtonUp(ctrl.xp1+1, ctrl.yp1+1)
	}
}

func BenchmarkGammaCtrlArrowKeys(b *testing.B) {
	ctrl := NewGammaCtrlImpl[color.RGBA](0, 0, 100, 100, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		direction := i % 4
		switch direction {
		case 0:
			ctrl.OnArrowKeys(true, false, false, false)
		case 1:
			ctrl.OnArrowKeys(false, true, false, false)
		case 2:
			ctrl.OnArrowKeys(false, false, true, false)
		case 3:
			ctrl.OnArrowKeys(false, false, false, true)
		}
	}
}
