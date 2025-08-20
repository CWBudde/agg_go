package slider

import (
	"math"
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestSliderCtrlStepsRendering(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetNumSteps(5) // Should create 6 tick marks (0, 1, 2, 3, 4, 5)

	// Rewind steps path (path 5)
	slider.Rewind(5)

	vertexCount := 0
	for {
		x, y, cmd := slider.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// Vertices should be within bounds
		if x < slider.X1()-10 || x > slider.X2()+10 {
			t.Errorf("Step vertex X coordinate %.2f outside reasonable bounds [%.2f, %.2f]", x, slider.X1()-10, slider.X2()+10)
		}
		if y < slider.Y1()-10 || y > slider.Y2()+10 {
			t.Errorf("Step vertex Y coordinate %.2f outside reasonable bounds [%.2f, %.2f]", y, slider.Y1()-10, slider.Y2()+10)
		}
	}

	// Should have generated vertices for steps (each step has 3 vertices: move, line, line)
	if vertexCount == 0 {
		t.Error("Expected step vertices to be generated")
	}
}

func TestSliderCtrlTransformation(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 100, 20, false)

	// Test basic vertex generation without transformation
	slider.Rewind(0)                // Background path
	x1, y1, cmd1 := slider.Vertex() // First vertex

	if cmd1 != basics.PathCmdMoveTo {
		t.Errorf("Expected first vertex to be MoveTo, got %v", cmd1)
	}

	// Coordinates should be at expected positions
	expectedX := slider.X1() - slider.borderExtra
	expectedY := slider.Y1() - slider.borderExtra

	// Allow small tolerance for floating point comparison
	tolerance := 0.001
	if math.Abs(x1-expectedX) > tolerance || math.Abs(y1-expectedY) > tolerance {
		t.Errorf("Expected first vertex at (%.3f, %.3f), got (%.3f, %.3f)", expectedX, expectedY, x1, y1)
	}
}

func TestSliderCtrlAdvancedMouseDrag(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 100)
	slider.SetValue(50) // Start at middle

	// Calculate pointer position
	pointerX := slider.xs1 + (slider.xs2-slider.xs1)*slider.value
	pointerY := (slider.ys1 + slider.ys2) / 2.0

	// Start drag
	result := slider.OnMouseButtonDown(pointerX, pointerY)
	if !result {
		t.Fatal("Mouse button down should succeed on pointer")
	}

	// Test basic drag behavior

	// Test dragging to end position
	result = slider.OnMouseMove(slider.xs2, pointerY, true)
	if !result {
		t.Error("Mouse move should return true when dragging")
	}

	// End drag
	result = slider.OnMouseButtonUp(pointerX, pointerY)
	if !result {
		t.Error("Mouse button up should return true when ending drag")
	}

	// After drag ends, mouse move without button should not affect value
	lastValue := slider.Value()
	result = slider.OnMouseMove(slider.xs1, pointerY, false)
	if result {
		t.Error("Mouse move without button should return false")
	}
	if slider.Value() != lastValue {
		t.Error("Value should not change after drag ends")
	}
}

func TestSliderCtrlColorCustomization(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)

	// Test setting custom colors
	customRed := color.NewRGBA(1.0, 0.0, 0.0, 1.0)
	customBlue := color.NewRGBA(0.0, 0.0, 1.0, 1.0)

	slider.SetPointerColor(customRed)
	slider.SetBackgroundColor(customBlue)

	// Verify colors are set correctly
	if slider.Color(4) != customRed {
		t.Errorf("Expected pointer color to be %+v, got %+v", customRed, slider.Color(4))
	}
	if slider.Color(0) != customBlue {
		t.Errorf("Expected background color to be %+v, got %+v", customBlue, slider.Color(0))
	}

	// Test text color affects both text paths
	customGreen := color.NewRGBA(0.0, 1.0, 0.0, 1.0)
	slider.SetTextColor(customGreen)

	if slider.Color(2) != customGreen {
		t.Errorf("Expected text color (path 2) to be %+v, got %+v", customGreen, slider.Color(2))
	}
	if slider.Color(5) != customGreen {
		t.Errorf("Expected text color (path 5) to be %+v, got %+v", customGreen, slider.Color(5))
	}
}
