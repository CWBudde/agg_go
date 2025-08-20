package slider

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func TestNewSliderCtrl(t *testing.T) {
	x1, y1, x2, y2 := 10.0, 20.0, 210.0, 40.0
	flipY := false

	slider := NewSliderCtrl(x1, y1, x2, y2, flipY)

	if slider.X1() != x1 || slider.Y1() != y1 || slider.X2() != x2 || slider.Y2() != y2 {
		t.Errorf("Expected bounds (%.1f, %.1f, %.1f, %.1f), got (%.1f, %.1f, %.1f, %.1f)",
			x1, y1, x2, y2, slider.X1(), slider.Y1(), slider.X2(), slider.Y2())
	}

	if slider.FlipY() != flipY {
		t.Errorf("Expected FlipY %v, got %v", flipY, slider.FlipY())
	}

	// Test default values
	if slider.Value() != 0.5 {
		t.Errorf("Expected default value 0.5, got %.3f", slider.Value())
	}

	if slider.NumPaths() != 6 {
		t.Errorf("Expected 6 rendering paths, got %d", slider.NumPaths())
	}
}

func TestSliderCtrlRange(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test default range [0, 1]
	slider.SetValue(0.5)
	if slider.Value() != 0.5 {
		t.Errorf("Expected value 0.5 with default range, got %.3f", slider.Value())
	}

	// Test custom range [-10, 10]
	slider.SetRange(-10, 10)
	slider.SetValue(0) // Should be middle of range
	if slider.Value() != 0.0 {
		t.Errorf("Expected value 0.0 with range [-10, 10], got %.3f", slider.Value())
	}

	slider.SetValue(-10) // Minimum
	if slider.Value() != -10.0 {
		t.Errorf("Expected minimum value -10.0, got %.3f", slider.Value())
	}

	slider.SetValue(10) // Maximum
	if slider.Value() != 10.0 {
		t.Errorf("Expected maximum value 10.0, got %.3f", slider.Value())
	}

	// Test clamping
	slider.SetValue(-20) // Below minimum
	if slider.Value() != -10.0 {
		t.Errorf("Expected clamped minimum value -10.0, got %.3f", slider.Value())
	}

	slider.SetValue(20) // Above maximum
	if slider.Value() != 10.0 {
		t.Errorf("Expected clamped maximum value 10.0, got %.3f", slider.Value())
	}
}

func TestSliderCtrlSteps(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 10)
	slider.SetNumSteps(10) // 11 positions: 0, 1, 2, ..., 10

	// Test step quantization
	slider.SetValue(2.7) // Should snap to 3
	if slider.Value() != 3.0 {
		t.Errorf("Expected stepped value 3.0, got %.3f", slider.Value())
	}

	slider.SetValue(2.2) // Should snap to 2
	if slider.Value() != 2.0 {
		t.Errorf("Expected stepped value 2.0, got %.3f", slider.Value())
	}
}

func TestSliderCtrlMouseInteraction(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 100)

	// Test hit testing
	if !slider.InRect(100, 10) { // Center of slider
		t.Error("Expected center point to be in rect")
	}

	if slider.InRect(-50, 10) { // Far left (outside bounds)
		t.Error("Expected far left point to be outside rect")
	}

	// Set a known value and test mouse down on the pointer
	slider.SetValue(50) // Center value

	// Calculate expected pointer position
	pointerX := slider.xs1 + (slider.xs2-slider.xs1)*slider.value
	pointerY := (slider.ys1 + slider.ys2) / 2.0

	// Test mouse button down on pointer
	result := slider.OnMouseButtonDown(pointerX, pointerY)
	if !result {
		t.Error("Expected mouse button down to return true when clicking on pointer")
	}

	// Test mouse button down outside pointer
	result = slider.OnMouseButtonDown(-100, 10)
	if result {
		t.Error("Expected mouse button down to return false when outside pointer")
	}

	// Test mouse move while dragging
	result = slider.OnMouseMove(pointerX+20, pointerY, true) // Move right
	if !result {
		t.Error("Expected mouse move to return true when dragging")
	}

	// Test mouse button up
	result = slider.OnMouseButtonUp(pointerX+20, pointerY)
	if !result {
		t.Error("Expected mouse button up to return true when ending drag")
	}
}

func TestSliderCtrlKeyboardNavigation(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 100)
	slider.SetValue(50) // Start at middle

	originalValue := slider.Value()

	// Test right arrow (should increase value)
	result := slider.OnArrowKeys(false, true, false, false)
	if !result {
		t.Error("Expected arrow key to return true")
	}
	if slider.Value() <= originalValue {
		t.Errorf("Expected value to increase after right arrow, was %.3f now %.3f",
			originalValue, slider.Value())
	}

	// Test left arrow (should decrease value)
	currentValue := slider.Value()
	result = slider.OnArrowKeys(true, false, false, false)
	if !result {
		t.Error("Expected arrow key to return true")
	}
	if slider.Value() >= currentValue {
		t.Errorf("Expected value to decrease after left arrow, was %.3f now %.3f",
			currentValue, slider.Value())
	}
}

func TestSliderCtrlDescending(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 100)
	slider.SetDescending(true)

	// In AGG, descending mode only affects visual triangle direction,
	// not the actual value mapping. Mouse and keyboard work the same.
	slider.SetValue(50) // Start at middle

	// Test keyboard navigation - should work same as non-descending
	originalValue := slider.Value()
	t.Logf("Debug: Original value before arrow key: %.3f", originalValue)
	result := slider.OnArrowKeys(false, true, false, false) // Right arrow
	newValue := slider.Value()
	t.Logf("Debug: Value after right arrow: %.3f (returned %v)", newValue, result)
	if newValue <= originalValue {
		t.Errorf("Expected right arrow to increase value, was %.3f now %.3f", originalValue, newValue)
	}

	// Test left arrow decreases
	originalValue = slider.Value()
	result = slider.OnArrowKeys(true, false, false, false) // Left arrow
	newValue = slider.Value()
	if newValue >= originalValue {
		t.Errorf("Expected left arrow to decrease value, was %.3f now %.3f", originalValue, newValue)
	}
}

func TestSliderCtrlVertexGeneration(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test that all paths can be rewound and generate vertices
	for pathID := uint(0); pathID < slider.NumPaths(); pathID++ {
		slider.Rewind(pathID)

		// Try to get at least one vertex from each path
		x, y, cmd := slider.Vertex()
		_ = x // Vertex coordinates
		_ = y

		// Should get either a valid command or stop
		if cmd != basics.PathCmdMoveTo && cmd != basics.PathCmdLineTo && cmd != basics.PathCmdStop {
			t.Errorf("Path %d: Expected valid path command, got %v", pathID, cmd)
		}

		// Test color access
		color := slider.Color(pathID)
		// Colors are RGBA structs, so they can't be nil
		// Check if it's a valid color (non-negative components)
		if color.R < 0 || color.G < 0 || color.B < 0 || color.A < 0 {
			t.Errorf("Path %d: Expected valid color, got %+v", pathID, color)
		}
	}
}

func TestSliderCtrlBorderSettings(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test border width setting
	slider.SetBorderWidth(2.0, 5.0)

	// Border settings should affect internal layout calculations
	// This is tested indirectly through the calcBox() method

	// The actual effect would be visible in the vertex generation
	// but we can at least verify the method doesn't panic
	slider.Rewind(0) // Background path
}

func TestSliderCtrlLabel(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test setting a label format
	slider.SetLabel("Value: %.1f")

	// Test text thickness
	slider.SetTextThickness(2.0)

	// Rewind text path to generate text
	slider.Rewind(2) // Text path

	// The text should be generated, but we can't easily test the content
	// without more complex text rendering verification
	x, y, cmd := slider.Vertex()
	_ = x
	_ = y
	_ = cmd
}

func TestSliderCtrlEdgeCases(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test zero-width slider
	slider.SetBounds(100, 100, 100, 120)
	result := slider.OnMouseButtonDown(100, 110)
	// Should handle gracefully without crashing
	_ = result

	// Test invalid range
	slider.SetRange(10, 0) // max < min
	slider.SetValue(5)
	// Should handle gracefully

	// Test very large number of steps
	slider.SetNumSteps(1000000)
	slider.SetValue(50)
	// Should handle gracefully

	// Test vertex generation with extreme values
	for pathID := uint(0); pathID < slider.NumPaths(); pathID++ {
		slider.Rewind(pathID)

		// Try to read many vertices to test bounds
		for i := 0; i < 100; i++ {
			x, y, cmd := slider.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			// Verify coordinates aren't NaN or infinite
			if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
				t.Errorf("Path %d vertex %d: Invalid coordinates (%.3f, %.3f)", pathID, i, x, y)
			}
		}
	}
}
