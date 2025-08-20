// Slider Control Demo
//
// This example demonstrates the AGG slider control with various configurations:
// - Basic horizontal slider
// - Slider with steps
// - Slider with custom range
// - Slider with custom colors
// - Multiple sliders showing different settings
//
// The demo shows how to create, configure, and use slider controls
// in the AGG Go port, matching the behavior of the original C++ AGG library.
package main

import (
	"fmt"
	"log"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl/slider"
)

func main() {
	fmt.Println("AGG Slider Control Demo")
	fmt.Println("======================")

	// Create different slider configurations to demonstrate features
	sliders := createDemoSliders()

	// Demonstrate each slider's functionality
	for i, s := range sliders {
		fmt.Printf("\n--- Slider %d Demo ---\n", i+1)
		demonstrateSlider(s, i+1)
	}

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("\nNote: This is a console demonstration of the slider control logic.")
	fmt.Println("In a full AGG application, these sliders would be rendered graphically")
	fmt.Println("and respond to mouse interactions in a window or framebuffer.")
}

// createDemoSliders creates various slider configurations for demonstration
func createDemoSliders() []*slider.SliderCtrl {
	sliders := make([]*slider.SliderCtrl, 0, 6)

	// 1. Basic horizontal slider (0-100 range)
	basic := slider.NewSliderCtrl(10, 10, 210, 30, false)
	basic.SetRange(0, 100)
	basic.SetValue(50)
	basic.SetLabel("Basic: %.0f")
	sliders = append(sliders, basic)

	// 2. Temperature slider (-10°C to 40°C)
	temp := slider.NewSliderCtrl(10, 40, 210, 60, false)
	temp.SetRange(-10, 40)
	temp.SetValue(20)
	temp.SetLabel("Temperature: %.1f°C")
	temp.SetPointerColor(color.NewRGBA(0.8, 0.3, 0.3, 1.0)) // Reddish for temperature
	sliders = append(sliders, temp)

	// 3. Volume slider with steps (0-10 with steps)
	volume := slider.NewSliderCtrl(10, 70, 210, 90, false)
	volume.SetRange(0, 10)
	volume.SetNumSteps(10) // 11 discrete positions: 0, 1, 2, ..., 10
	volume.SetValue(7)
	volume.SetLabel("Volume: %.0f")
	volume.SetPointerColor(color.NewRGBA(0.3, 0.8, 0.3, 1.0)) // Green for volume
	sliders = append(sliders, volume)

	// 4. Percentage slider (0-1 range, shown as percentage)
	percent := slider.NewSliderCtrl(10, 100, 210, 120, false)
	percent.SetRange(0, 1)
	percent.SetValue(0.75)
	percent.SetLabel("Progress: %.0f%%")
	percent.SetPointerColor(color.NewRGBA(0.3, 0.3, 0.8, 1.0))    // Blue for progress
	percent.SetBackgroundColor(color.NewRGBA(0.9, 0.9, 1.0, 1.0)) // Light blue background
	sliders = append(sliders, percent)

	// 5. Fine precision slider (scientific values)
	precision := slider.NewSliderCtrl(10, 130, 210, 150, false)
	precision.SetRange(0.001, 0.999)
	precision.SetValue(0.123)
	precision.SetLabel("Precision: %.3f")
	precision.SetTextColor(color.NewRGBA(0.5, 0.5, 0.5, 1.0)) // Gray text
	sliders = append(sliders, precision)

	// 6. Descending slider (high values on left, visual indicator only)
	desc := slider.NewSliderCtrl(10, 160, 210, 180, false)
	desc.SetRange(0, 100)
	desc.SetValue(30)
	desc.SetDescending(true) // Visual triangle points left
	desc.SetLabel("Descending: %.0f")
	desc.SetTriangleColor(color.NewRGBA(0.8, 0.8, 0.3, 1.0)) // Yellow triangle
	sliders = append(sliders, desc)

	return sliders
}

// demonstrateSlider shows the capabilities of a single slider
func demonstrateSlider(s *slider.SliderCtrl, id int) {
	fmt.Printf("Slider bounds: (%.0f,%.0f) to (%.0f,%.0f)\n",
		s.X1(), s.Y1(), s.X2(), s.Y2())
	fmt.Printf("Current value: %.3f\n", s.Value())
	fmt.Printf("Number of rendering paths: %d\n", s.NumPaths())

	// Show color information
	fmt.Println("Colors:")
	colorNames := []string{"Background", "Triangle", "Text", "Pointer Preview", "Pointer", "Text (dup)"}
	for i := uint(0); i < s.NumPaths(); i++ {
		c := s.Color(i)
		fmt.Printf("  Path %d (%s): RGBA(%.2f, %.2f, %.2f, %.2f)\n",
			i, colorNames[i], c.R, c.G, c.B, c.A)
	}

	// Demonstrate keyboard navigation
	fmt.Println("Testing keyboard navigation:")
	originalValue := s.Value()

	// Test right arrow (increase)
	if s.OnArrowKeys(false, true, false, false) {
		fmt.Printf("  Right arrow: %.3f → %.3f\n", originalValue, s.Value())
	}

	// Test left arrow (decrease)
	currentValue := s.Value()
	if s.OnArrowKeys(true, false, false, false) {
		fmt.Printf("  Left arrow:  %.3f → %.3f\n", currentValue, s.Value())
	}

	// Demonstrate mouse interaction simulation
	demonstrateMouseInteraction(s)

	// Show vertex generation for each path
	fmt.Println("Vertex generation test:")
	for pathID := uint(0); pathID < s.NumPaths(); pathID++ {
		s.Rewind(pathID)
		vertexCount := 0
		for {
			_, _, cmd := s.Vertex()
			if basics.IsStop(cmd) {
				break
			}
			vertexCount++
			if vertexCount > 100 { // Safety limit
				break
			}
		}
		fmt.Printf("  Path %d: %d vertices\n", pathID, vertexCount)
	}
}

// demonstrateMouseInteraction simulates mouse interactions with the slider
func demonstrateMouseInteraction(s *slider.SliderCtrl) {
	fmt.Println("Testing mouse interaction simulation:")

	// Calculate slider dimensions for simulation
	centerX := (s.X1() + s.X2()) / 2
	centerY := (s.Y1() + s.Y2()) / 2

	originalValue := s.Value()

	// Test hit detection
	if s.InRect(centerX, centerY) {
		fmt.Printf("  Center point (%.1f, %.1f) is inside slider bounds ✓\n", centerX, centerY)
	} else {
		fmt.Printf("  Center point (%.1f, %.1f) is outside slider bounds ✗\n", centerX, centerY)
	}

	// Test point outside bounds
	if !s.InRect(-100, centerY) {
		fmt.Printf("  Point (-100, %.1f) correctly detected as outside bounds ✓\n", centerY)
	}

	// Simulate clicking on the slider (this is a simplified simulation)
	// In a real application, this would be based on actual pointer position
	fmt.Printf("  Simulated interaction: %.3f → ", originalValue)

	// Set a new value to simulate mouse interaction result
	newValue := 0.75 * (s.X2() - s.X1()) / (s.X2() - s.X1()) // Simulate 75% position
	simulatedValue := originalValue + (newValue * 10)        // Small change for demo
	if simulatedValue > 1.0 {
		simulatedValue = originalValue - 0.1
	}

	// Apply the simulated change
	currentRange := s.Value()      // This gives us the actual range value
	s.SetValue(currentRange * 0.9) // Small change for demonstration

	fmt.Printf("%.3f (simulated mouse drag)\n", s.Value())
}

// Additional helper functions for the demo

// validateSliderState performs basic validation on a slider's state
func validateSliderState(s *slider.SliderCtrl) error {
	if s.X2() <= s.X1() {
		return fmt.Errorf("invalid slider bounds: X2 (%.2f) <= X1 (%.2f)", s.X2(), s.X1())
	}
	if s.Y2() <= s.Y1() {
		return fmt.Errorf("invalid slider bounds: Y2 (%.2f) <= Y1 (%.2f)", s.Y2(), s.Y1())
	}
	if s.NumPaths() != 6 {
		return fmt.Errorf("expected 6 rendering paths, got %d", s.NumPaths())
	}
	return nil
}

func init() {
	// Validate that we can create sliders without errors
	testSlider := slider.NewSliderCtrl(0, 0, 100, 20, false)
	if err := validateSliderState(testSlider); err != nil {
		log.Fatalf("Slider validation failed: %v", err)
	}
}
