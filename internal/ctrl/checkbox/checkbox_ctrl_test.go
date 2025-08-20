package checkbox

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/color"
)

func TestNewCheckboxCtrl(t *testing.T) {
	// Test basic constructor
	checkbox := NewCheckboxCtrl(10.0, 20.0, "Test Label", false)

	if checkbox == nil {
		t.Fatal("NewCheckboxCtrl returned nil")
	}

	// Test initial state
	if checkbox.IsChecked() {
		t.Error("New checkbox should be unchecked by default")
	}

	// Test bounds - checkbox should be 13.5 units square (9.0 * 1.5)
	expectedSize := 9.0 * 1.5
	if checkbox.X1() != 10.0 {
		t.Errorf("Expected X1=10.0, got %f", checkbox.X1())
	}
	if checkbox.Y1() != 20.0 {
		t.Errorf("Expected Y1=20.0, got %f", checkbox.Y1())
	}
	if checkbox.X2() != 10.0+expectedSize {
		t.Errorf("Expected X2=%f, got %f", 10.0+expectedSize, checkbox.X2())
	}
	if checkbox.Y2() != 20.0+expectedSize {
		t.Errorf("Expected Y2=%f, got %f", 20.0+expectedSize, checkbox.Y2())
	}

	// Test label
	if checkbox.Label() != "Test Label" {
		t.Errorf("Expected label 'Test Label', got '%s'", checkbox.Label())
	}

	// Test flip Y
	if checkbox.FlipY() {
		t.Error("FlipY should be false")
	}

	// Test flip Y = true
	checkboxFlipped := NewCheckboxCtrl(0.0, 0.0, "", true)
	if !checkboxFlipped.FlipY() {
		t.Error("FlipY should be true")
	}
}

func TestCheckboxState(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "", false)

	// Test initial state
	if checkbox.IsChecked() {
		t.Error("New checkbox should be unchecked")
	}

	// Test SetChecked
	checkbox.SetChecked(true)
	if !checkbox.IsChecked() {
		t.Error("Checkbox should be checked after SetChecked(true)")
	}

	checkbox.SetChecked(false)
	if checkbox.IsChecked() {
		t.Error("Checkbox should be unchecked after SetChecked(false)")
	}

	// Test Toggle
	checkbox.Toggle()
	if !checkbox.IsChecked() {
		t.Error("Checkbox should be checked after Toggle()")
	}

	checkbox.Toggle()
	if checkbox.IsChecked() {
		t.Error("Checkbox should be unchecked after second Toggle()")
	}
}

func TestCheckboxLabel(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Initial", false)

	// Test initial label
	if checkbox.Label() != "Initial" {
		t.Errorf("Expected label 'Initial', got '%s'", checkbox.Label())
	}

	// Test SetLabel
	checkbox.SetLabel("New Label")
	if checkbox.Label() != "New Label" {
		t.Errorf("Expected label 'New Label', got '%s'", checkbox.Label())
	}

	// Test empty label
	checkbox.SetLabel("")
	if checkbox.Label() != "" {
		t.Errorf("Expected empty label, got '%s'", checkbox.Label())
	}

	// Test long label (should be truncated to 127 characters)
	longLabel := make([]byte, 200)
	for i := range longLabel {
		longLabel[i] = 'A'
	}
	checkbox.SetLabel(string(longLabel))
	if len(checkbox.Label()) != 127 {
		t.Errorf("Expected label length 127, got %d", len(checkbox.Label()))
	}
}

func TestCheckboxTextSettings(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Test", false)

	// Test SetTextThickness
	checkbox.SetTextThickness(2.5)
	if checkbox.textThickness != 2.5 {
		t.Errorf("Expected text thickness 2.5, got %f", checkbox.textThickness)
	}

	// Test SetTextSize
	checkbox.SetTextSize(12.0, 8.0)
	if checkbox.textHeight != 12.0 {
		t.Errorf("Expected text height 12.0, got %f", checkbox.textHeight)
	}
	if checkbox.textWidth != 8.0 {
		t.Errorf("Expected text width 8.0, got %f", checkbox.textWidth)
	}

	// Test proportional width (0.0)
	checkbox.SetTextSize(10.0, 0.0)
	if checkbox.textWidth != 0.0 {
		t.Errorf("Expected text width 0.0 (proportional), got %f", checkbox.textWidth)
	}
}

func TestCheckboxColors(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Test", false)

	// Test default colors
	expectedInactive := color.NewRGBA(0.0, 0.0, 0.0, 1.0)
	expectedText := color.NewRGBA(0.0, 0.0, 0.0, 1.0)
	expectedActive := color.NewRGBA(0.4, 0.0, 0.0, 1.0)

	if checkbox.colors[0] != expectedInactive {
		t.Errorf("Expected default inactive color %v, got %v", expectedInactive, checkbox.colors[0])
	}
	if checkbox.colors[1] != expectedText {
		t.Errorf("Expected default text color %v, got %v", expectedText, checkbox.colors[1])
	}
	if checkbox.colors[2] != expectedActive {
		t.Errorf("Expected default active color %v, got %v", expectedActive, checkbox.colors[2])
	}

	// Test color setters
	redColor := color.NewRGBA(1.0, 0.0, 0.0, 1.0)
	greenColor := color.NewRGBA(0.0, 1.0, 0.0, 1.0)
	blueColor := color.NewRGBA(0.0, 0.0, 1.0, 1.0)

	checkbox.SetInactiveColor(redColor)
	checkbox.SetTextColor(greenColor)
	checkbox.SetActiveColor(blueColor)

	if checkbox.colors[0] != redColor {
		t.Errorf("Expected inactive color %v, got %v", redColor, checkbox.colors[0])
	}
	if checkbox.colors[1] != greenColor {
		t.Errorf("Expected text color %v, got %v", greenColor, checkbox.colors[1])
	}
	if checkbox.colors[2] != blueColor {
		t.Errorf("Expected active color %v, got %v", blueColor, checkbox.colors[2])
	}
}

func TestCheckboxMouseInteraction(t *testing.T) {
	checkbox := NewCheckboxCtrl(10.0, 20.0, "Test", false)

	// Test click inside bounds - should toggle
	if !checkbox.OnMouseButtonDown(12.0, 22.0) {
		t.Error("Mouse click inside bounds should return true")
	}
	if !checkbox.IsChecked() {
		t.Error("Mouse click inside bounds should toggle checkbox to checked")
	}

	// Click again - should toggle back
	if !checkbox.OnMouseButtonDown(12.0, 22.0) {
		t.Error("Second mouse click inside bounds should return true")
	}
	if checkbox.IsChecked() {
		t.Error("Second mouse click inside bounds should toggle checkbox to unchecked")
	}

	// Test click outside bounds - should not toggle
	if checkbox.OnMouseButtonDown(5.0, 5.0) {
		t.Error("Mouse click outside bounds should return false")
	}
	if checkbox.IsChecked() {
		t.Error("Mouse click outside bounds should not change checkbox state")
	}

	// Test boundary clicks (at edges)
	checkboxSize := 9.0 * 1.5

	// Top-left corner (should be inside)
	if !checkbox.OnMouseButtonDown(10.0, 20.0) {
		t.Error("Click at top-left corner should be inside bounds")
	}

	// Bottom-right corner (should be inside)
	checkbox.SetChecked(false) // Reset state
	if !checkbox.OnMouseButtonDown(10.0+checkboxSize, 20.0+checkboxSize) {
		t.Error("Click at bottom-right corner should be inside bounds")
	}

	// Just outside bounds
	checkbox.SetChecked(false) // Reset state
	if checkbox.OnMouseButtonDown(10.0+checkboxSize+0.1, 20.0) {
		t.Error("Click just outside right edge should be outside bounds")
	}
	if checkbox.IsChecked() {
		t.Error("Click outside bounds should not change state")
	}
}

func TestCheckboxMouseOtherEvents(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Test", false)

	// Test OnMouseButtonUp (should always return false)
	if checkbox.OnMouseButtonUp(5.0, 5.0) {
		t.Error("OnMouseButtonUp should always return false")
	}

	// Test OnMouseMove (should always return false)
	if checkbox.OnMouseMove(5.0, 5.0, true) {
		t.Error("OnMouseMove should always return false")
	}
	if checkbox.OnMouseMove(5.0, 5.0, false) {
		t.Error("OnMouseMove should always return false")
	}

	// Test OnArrowKeys (should always return false)
	if checkbox.OnArrowKeys(true, false, false, false) {
		t.Error("OnArrowKeys should always return false")
	}
	if checkbox.OnArrowKeys(false, true, true, true) {
		t.Error("OnArrowKeys should always return false")
	}
}

func TestCheckboxVertexSource(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Test", false)

	// Test NumPaths
	if checkbox.NumPaths() != 3 {
		t.Errorf("Expected 3 paths, got %d", checkbox.NumPaths())
	}

	// Test Color method
	redColor := color.NewRGBA(1.0, 0.0, 0.0, 1.0)
	greenColor := color.NewRGBA(0.0, 1.0, 0.0, 1.0)
	blueColor := color.NewRGBA(0.0, 0.0, 1.0, 1.0)

	checkbox.SetInactiveColor(redColor)
	checkbox.SetTextColor(greenColor)
	checkbox.SetActiveColor(blueColor)

	if checkbox.Color(0) != redColor {
		t.Errorf("Expected color for path 0 to be %v, got %v", redColor, checkbox.Color(0))
	}
	if checkbox.Color(1) != greenColor {
		t.Errorf("Expected color for path 1 to be %v, got %v", greenColor, checkbox.Color(1))
	}
	if checkbox.Color(2) != blueColor {
		t.Errorf("Expected color for path 2 to be %v, got %v", blueColor, checkbox.Color(2))
	}

	// Test invalid path ID (should return default color)
	if checkbox.Color(999) != redColor {
		t.Errorf("Expected color for invalid path ID to be default (inactive) color %v, got %v", redColor, checkbox.Color(999))
	}
}

func TestCheckboxBorderVertices(t *testing.T) {
	checkbox := NewCheckboxCtrl(10.0, 20.0, "", false)

	// Test border path (path 0)
	checkbox.Rewind(0)

	// Should generate 8 vertices (outer + inner rectangle)
	vertexCount := 0
	for {
		x, y, cmd := checkbox.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// Verify commands
		if vertexCount == 1 || vertexCount == 5 {
			if cmd != basics.PathCmdMoveTo {
				t.Errorf("Vertex %d should be MoveTo, got %v", vertexCount, cmd)
			}
		} else {
			if cmd != basics.PathCmdLineTo {
				t.Errorf("Vertex %d should be LineTo, got %v", vertexCount, cmd)
			}
		}

		// Verify coordinates are within reasonable bounds
		if x < 0.0 || x > 50.0 || y < 0.0 || y > 50.0 {
			t.Errorf("Vertex %d coordinates (%f, %f) seem out of bounds", vertexCount, x, y)
		}
	}

	if vertexCount != 8 {
		t.Errorf("Expected 8 border vertices, got %d", vertexCount)
	}
}

func TestCheckboxTextVertices(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "A", false)

	// Test text path (path 1)
	checkbox.Rewind(1)

	// Should generate some vertices for the letter "A"
	vertexCount := 0
	for {
		_, _, cmd := checkbox.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// Verify command types
		if cmd != basics.PathCmdMoveTo && cmd != basics.PathCmdLineTo {
			t.Errorf("Text vertex %d has invalid command %v", vertexCount, cmd)
		}
	}

	if vertexCount == 0 {
		t.Error("Expected some text vertices for letter 'A', got none")
	}

	// Test empty label
	checkbox.SetLabel("")
	checkbox.Rewind(1)

	_, _, cmd := checkbox.Vertex()
	if cmd != basics.PathCmdStop {
		t.Error("Empty label should immediately return PathCmdStop")
	}
}

func TestCheckboxCheckmarkVertices(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "", false)

	// Test checkmark path when unchecked (path 2)
	checkbox.Rewind(2)
	_, _, cmd := checkbox.Vertex()
	if cmd != basics.PathCmdStop {
		t.Error("Unchecked checkbox should not generate checkmark vertices")
	}

	// Test checkmark path when checked
	checkbox.SetChecked(true)
	checkbox.Rewind(2)

	vertexCount := 0
	firstVertexSeen := false
	for {
		x, y, cmd := checkbox.Vertex()
		if cmd == basics.PathCmdStop {
			break
		}
		vertexCount++

		// First vertex should be MoveTo
		if !firstVertexSeen {
			if cmd != basics.PathCmdMoveTo {
				t.Errorf("First checkmark vertex should be MoveTo, got %v", cmd)
			}
			firstVertexSeen = true
		} else {
			if cmd != basics.PathCmdLineTo {
				t.Errorf("Checkmark vertex %d should be LineTo, got %v", vertexCount, cmd)
			}
		}

		// Verify coordinates are reasonable
		if x < -50.0 || x > 50.0 || y < -50.0 || y > 50.0 {
			t.Errorf("Checkmark vertex %d coordinates (%f, %f) seem out of bounds", vertexCount, x, y)
		}
	}

	if vertexCount != 8 {
		t.Errorf("Expected 8 checkmark vertices, got %d", vertexCount)
	}
}

func TestCheckboxInvalidPath(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "Test", false)

	// Test invalid path ID
	checkbox.Rewind(999)
	_, _, cmd := checkbox.Vertex()
	if cmd != basics.PathCmdStop {
		t.Error("Invalid path ID should immediately return PathCmdStop")
	}
}

func TestCheckboxWithTransformation(t *testing.T) {
	checkbox := NewCheckboxCtrl(0.0, 0.0, "", false)

	// Test that clicks work with coordinate transformation
	// The InverseTransformXY should be called in OnMouseButtonDown

	// Click at origin
	if !checkbox.OnMouseButtonDown(0.0, 0.0) {
		t.Error("Click at origin should be inside bounds")
	}
	if !checkbox.IsChecked() {
		t.Error("Click should have toggled checkbox")
	}

	// Test boundary detection with InRect inherited from BaseCtrl
	if !checkbox.InRect(0.0, 0.0) {
		t.Error("Origin should be inside checkbox bounds")
	}

	checkboxSize := 9.0 * 1.5
	if checkbox.InRect(-1.0, 0.0) {
		t.Error("Point outside left boundary should not be inside bounds")
	}
	if checkbox.InRect(checkboxSize+1.0, 0.0) {
		t.Error("Point outside right boundary should not be inside bounds")
	}
}

// Integration test to verify the checkbox works as expected in a complete scenario
func TestCheckboxIntegration(t *testing.T) {
	// Create checkbox with all features
	checkbox := NewCheckboxCtrl(50.0, 100.0, "Enable Feature", true)

	// Configure appearance
	checkbox.SetTextSize(12.0, 0.0)
	checkbox.SetTextThickness(2.0)
	checkbox.SetTextColor(color.NewRGBA(0.2, 0.2, 0.8, 1.0))     // Blue text
	checkbox.SetInactiveColor(color.NewRGBA(0.5, 0.5, 0.5, 1.0)) // Gray border
	checkbox.SetActiveColor(color.NewRGBA(0.0, 0.8, 0.0, 1.0))   // Green checkmark

	// Verify initial state
	if checkbox.IsChecked() {
		t.Error("Checkbox should start unchecked")
	}
	if checkbox.Label() != "Enable Feature" {
		t.Error("Label should be set correctly")
	}
	if !checkbox.FlipY() {
		t.Error("FlipY should be true")
	}

	// Test interaction
	if !checkbox.OnMouseButtonDown(55.0, 105.0) {
		t.Error("Click inside bounds should work")
	}
	if !checkbox.IsChecked() {
		t.Error("Checkbox should be checked after click")
	}

	// Test rendering paths
	if checkbox.NumPaths() != 3 {
		t.Error("Should have 3 rendering paths")
	}

	// Verify all paths generate vertices
	for pathID := uint(0); pathID < checkbox.NumPaths(); pathID++ {
		checkbox.Rewind(pathID)

		vertexCount := 0
		for {
			_, _, cmd := checkbox.Vertex()
			if cmd == basics.PathCmdStop {
				break
			}
			vertexCount++
		}

		if pathID == 1 && vertexCount == 0 {
			t.Error("Text path should generate vertices for non-empty label")
		}
		if pathID == 2 && checkbox.IsChecked() && vertexCount == 0 {
			t.Error("Checkmark path should generate vertices when checked")
		}
	}

	// Test color retrieval
	if checkbox.Color(0) == nil {
		t.Error("Should return valid color for border path")
	}
	if checkbox.Color(1) == nil {
		t.Error("Should return valid color for text path")
	}
	if checkbox.Color(2) == nil {
		t.Error("Should return valid color for checkmark path")
	}
}
