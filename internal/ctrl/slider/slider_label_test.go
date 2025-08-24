package slider

import (
	"testing"

	"agg_go/internal/basics"
)

func TestSliderCtrlLabelRendering(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetRange(0, 100)
	slider.SetValue(75.5)

	// Test various label formats that use the new character support
	testCases := []struct {
		label    string
		expected string
	}{
		{"Value: %.1f", "Value: 75.5"},
		{"%.0f%%", "76%"},
		{"Level: %.2f", "Level: 75.50"},
		{"Test: %.0f", "Test: 76"},
		{"Config: %.1f", "Config: 75.5"},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			slider.SetLabel(tc.label)

			// Rewind text path to generate the text
			slider.Rewind(2)

			// Check that vertices are generated (indicates text is being rendered)
			vertexCount := 0
			for i := 0; i < 1000; i++ { // Limit iterations to prevent infinite loops
				_, _, cmd := slider.Vertex()
				if basics.IsStop(cmd) {
					break
				}
				vertexCount++
			}

			// Should generate some vertices for non-empty labels
			if vertexCount == 0 {
				t.Errorf("Expected text vertices for label %q, got none", tc.label)
			}
		})
	}
}

func TestSliderCtrlTextCharacters(t *testing.T) {
	slider := NewSliderCtrl(0, 0, 200, 20, false)
	slider.SetLabel("Value: 123.45%")

	// Test that the text renderer can handle various characters
	// This test verifies that the expanded character set doesn't crash
	slider.Rewind(2)

	vertexCount := 0
	for i := 0; i < 1000; i++ {
		_, _, cmd := slider.Vertex()
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
	}

	// Should generate vertices for the label
	if vertexCount == 0 {
		t.Error("Expected text vertices for complex label, got none")
	}
}
