// Radio Button Group (Rbox) Control Demo
//
// This example demonstrates the AGG radio button group control with various configurations:
// - Basic radio button group
// - Radio button group with different colors
// - Multiple radio button groups showing different settings
// - Mouse interaction simulation
// - Arrow key navigation simulation
//
// The demo shows how to create, configure, and use radio button controls
// in the AGG Go port, matching the behavior of the original C++ AGG library.
package main

import (
	"fmt"
	"log"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl/rbox"
)

func main() {
	fmt.Println("AGG Radio Button Group (Rbox) Control Demo")
	fmt.Println("==========================================")

	// Create different rbox configurations to demonstrate features
	rboxes := createDemoRboxes()

	// Demonstrate each rbox's functionality
	for i, r := range rboxes {
		fmt.Printf("\n--- Radio Button Group %d Demo ---\n", i+1)
		demonstrateRbox(r, i+1)
	}

	fmt.Println("\n--- Mouse Interaction Demo ---")
	demonstrateMouseInteraction(rboxes[0])

	fmt.Println("\n--- Arrow Key Navigation Demo ---")
	demonstrateArrowKeyNavigation(rboxes[1])

	fmt.Println("\n--- Vertex Generation Demo ---")
	demonstrateVertexGeneration(rboxes[2])

	fmt.Println("\nDemo completed successfully!")
	fmt.Println("\nNote: This is a console demonstration of the radio button control logic.")
	fmt.Println("In a full AGG application, these radio buttons would be rendered graphically")
	fmt.Println("and respond to mouse interactions in a window or framebuffer.")
}

// createDemoRboxes creates various rbox configurations for demonstration
func createDemoRboxes() []*rbox.RboxCtrl {
	rboxes := make([]*rbox.RboxCtrl, 0, 4)

	// 1. Basic color selection radio button group
	colorRbox := rbox.NewRboxCtrl(10, 10, 150, 120, false)
	colorRbox.AddItem("Red")
	colorRbox.AddItem("Green")
	colorRbox.AddItem("Blue")
	colorRbox.AddItem("Yellow")
	colorRbox.SetCurItem(0) // Select red by default
	rboxes = append(rboxes, colorRbox)

	// 2. Graphics quality radio button group with custom colors
	qualityRbox := rbox.NewRboxCtrl(170, 10, 300, 100, false)
	qualityRbox.AddItem("Low Quality")
	qualityRbox.AddItem("Medium Quality")
	qualityRbox.AddItem("High Quality")
	qualityRbox.SetCurItem(1) // Select medium by default

	// Customize colors
	qualityRbox.SetBackgroundColor(color.NewRGBA(0.95, 0.95, 1.0, 1.0)) // Light blue
	qualityRbox.SetBorderColor(color.NewRGBA(0.2, 0.2, 0.6, 1.0))       // Dark blue
	qualityRbox.SetTextColor(color.NewRGBA(0.1, 0.1, 0.4, 1.0))         // Navy
	qualityRbox.SetInactiveColor(color.NewRGBA(0.5, 0.5, 0.7, 1.0))     // Gray-blue
	qualityRbox.SetActiveColor(color.NewRGBA(0.8, 0.2, 0.2, 1.0))       // Red
	rboxes = append(rboxes, qualityRbox)

	// 3. File format radio button group with custom text sizing
	formatRbox := rbox.NewRboxCtrl(10, 140, 200, 250, false)
	formatRbox.AddItem("PNG")
	formatRbox.AddItem("JPEG")
	formatRbox.AddItem("BMP")
	formatRbox.AddItem("TIFF")
	formatRbox.AddItem("SVG")
	formatRbox.SetTextSize(12.0, 8.0) // Larger text
	formatRbox.SetTextThickness(2.0)  // Thicker text
	formatRbox.SetCurItem(2)          // Select BMP
	rboxes = append(rboxes, formatRbox)

	// 4. Single item radio button (edge case)
	singleRbox := rbox.NewRboxCtrl(220, 140, 320, 170, false)
	singleRbox.AddItem("Only Option")
	singleRbox.SetCurItem(0)
	rboxes = append(rboxes, singleRbox)

	return rboxes
}

// demonstrateRbox shows basic rbox functionality
func demonstrateRbox(r *rbox.RboxCtrl, groupNum int) {
	fmt.Printf("Radio Button Group %d:\n", groupNum)
	fmt.Printf("  Bounds: (%.0f,%.0f) to (%.0f,%.0f)\n", r.X1(), r.Y1(), r.X2(), r.Y2())
	fmt.Printf("  Number of items: %d\n", r.NumItems())
	fmt.Printf("  Currently selected: %d", r.CurItem())
	if r.CurItem() >= 0 {
		fmt.Printf(" ('%s')", r.ItemText(r.CurItem()))
	}
	fmt.Printf("\n")

	fmt.Printf("  Items:\n")
	for i := uint32(0); i < r.NumItems(); i++ {
		marker := "  "
		if int(i) == r.CurItem() {
			marker = "● " // Selected
		} else {
			marker = "○ " // Unselected
		}
		fmt.Printf("    %s %d: %s\n", marker, i, r.ItemText(int(i)))
	}

	fmt.Printf("  Text settings: height=%.1f, width=%.1f, thickness=%.1f\n",
		r.TextHeight(), r.TextWidth(), r.TextThickness())
	fmt.Printf("  Border width: %.1f\n", r.BorderWidth())
}

// demonstrateMouseInteraction simulates mouse clicks on radio buttons
func demonstrateMouseInteraction(r *rbox.RboxCtrl) {
	fmt.Printf("Testing mouse interaction on first radio button group:\n")

	originalSelection := r.CurItem()
	fmt.Printf("Original selection: %d\n", originalSelection)

	// Simulate clicking on different radio buttons
	// The positions are calculated based on the rbox's internal layout
	xs1 := r.X1() + r.BorderWidth()
	ys1 := r.Y1() + r.BorderWidth()
	dy := r.TextHeight() * 2.0

	testClicks := []struct {
		x, y float64
		desc string
	}{
		{xs1 + dy/1.3, ys1 + dy*0 + dy/1.3, "first radio button"},
		{xs1 + dy/1.3, ys1 + dy*1 + dy/1.3, "second radio button"},
		{xs1 + dy/1.3, ys1 + dy*2 + dy/1.3, "third radio button"},
		{xs1 + 100, ys1 + dy/1.3, "empty area (should not change selection)"},
	}

	for i, click := range testClicks {
		fmt.Printf("  Click %d at (%.1f, %.1f) - %s: ", i+1, click.x, click.y, click.desc)
		handled := r.OnMouseButtonDown(click.x, click.y)
		if handled {
			fmt.Printf("handled, new selection: %d ('%s')\n", r.CurItem(), r.ItemText(r.CurItem()))
		} else {
			fmt.Printf("not handled, selection remains: %d\n", r.CurItem())
		}
	}
}

// demonstrateArrowKeyNavigation simulates arrow key navigation
func demonstrateArrowKeyNavigation(r *rbox.RboxCtrl) {
	fmt.Printf("Testing arrow key navigation on second radio button group:\n")

	// Start with item 0 selected
	r.SetCurItem(0)
	fmt.Printf("Starting with item %d selected\n", r.CurItem())

	arrowTests := []struct {
		left, right, down, up bool
		desc                  string
	}{
		{false, true, false, false, "RIGHT arrow"},
		{false, false, false, true, "UP arrow"},
		{false, true, false, false, "RIGHT arrow (should wrap to beginning)"},
		{true, false, false, false, "LEFT arrow"},
		{false, false, true, false, "DOWN arrow"},
		{true, false, false, false, "LEFT arrow (should wrap to end)"},
	}

	for i, test := range arrowTests {
		oldSelection := r.CurItem()
		handled := r.OnArrowKeys(test.left, test.right, test.down, test.up)
		fmt.Printf("  Test %d - %s: ", i+1, test.desc)
		if handled {
			fmt.Printf("handled, selection changed from %d to %d ('%s')\n",
				oldSelection, r.CurItem(), r.ItemText(r.CurItem()))
		} else {
			fmt.Printf("not handled, selection remains %d\n", r.CurItem())
		}
	}

	// Test arrow keys with no item selected
	r.SetCurItem(-1)
	fmt.Printf("  Testing with no item selected: ")
	handled := r.OnArrowKeys(false, true, false, false)
	if handled {
		fmt.Printf("handled (unexpected)\n")
	} else {
		fmt.Printf("not handled (expected)\n")
	}
}

// demonstrateVertexGeneration shows vertex generation for all paths
func demonstrateVertexGeneration(r *rbox.RboxCtrl) {
	fmt.Printf("Testing vertex generation for third radio button group:\n")
	fmt.Printf("Number of rendering paths: %d\n", r.NumPaths())

	pathNames := []string{"Background", "Border", "Text", "Inactive Circles", "Active Circle"}

	for pathID := uint(0); pathID < r.NumPaths(); pathID++ {
		fmt.Printf("\n  Path %d (%s):\n", pathID, pathNames[pathID])

		r.Rewind(pathID)
		vertexCount := 0
		maxVertices := 50 // Limit output to prevent spam

		for vertexCount < maxVertices {
			x, y, cmd := r.Vertex()

			if cmd == basics.PathCmdStop {
				fmt.Printf("    [%d vertices] STOP\n", vertexCount)
				break
			}

			cmdName := getPathCommandName(cmd)
			if vertexCount < 10 { // Only show first 10 vertices
				fmt.Printf("    Vertex %d: (%.2f, %.2f) %s\n", vertexCount, x, y, cmdName)
			} else if vertexCount == 10 {
				fmt.Printf("    ... (showing first 10 vertices only)\n")
			}

			vertexCount++
		}

		if vertexCount >= maxVertices {
			fmt.Printf("    [Stopped at %d vertices to prevent excessive output]\n", maxVertices)
		}
	}
}

// getPathCommandName returns a human-readable name for a path command
func getPathCommandName(cmd basics.PathCommand) string {
	switch cmd & 0x0F {
	case basics.PathCmdStop:
		return "STOP"
	case basics.PathCmdMoveTo:
		return "MOVE_TO"
	case basics.PathCmdLineTo:
		return "LINE_TO"
	case basics.PathCmdCurve3:
		return "CURVE3"
	case basics.PathCmdCurve4:
		return "CURVE4"
	case basics.PathCmdCurveN:
		return "CURVE_N"
	case basics.PathCmdCatrom:
		return "CATROM"
	case basics.PathCmdUbspline:
		return "UBSPLINE"
	case basics.PathCmdEndPoly:
		flags := ""
		if uint32(cmd)&uint32(basics.PathFlagsClose) != 0 {
			flags += " CLOSE"
		}
		if uint32(cmd)&uint32(basics.PathFlagsCCW) != 0 {
			flags += " CCW"
		}
		if uint32(cmd)&uint32(basics.PathFlagsCW) != 0 {
			flags += " CW"
		}
		return "END_POLY" + flags
	default:
		return fmt.Sprintf("UNKNOWN(0x%02X)", uint32(cmd))
	}
}

// Example of how the rbox control would be used in a real application
func demonstrateRealWorldUsage() {
	fmt.Println("\n--- Real World Usage Example ---")

	// Create a settings dialog-style radio button group
	antiAliasingRbox := rbox.NewRboxCtrl(50, 50, 250, 150, false)

	// Add anti-aliasing options
	antiAliasingRbox.AddItem("No Anti-aliasing")
	antiAliasingRbox.AddItem("2x MSAA")
	antiAliasingRbox.AddItem("4x MSAA")
	antiAliasingRbox.AddItem("8x MSAA")

	// Set default to 4x MSAA
	antiAliasingRbox.SetCurItem(2)

	// Customize appearance
	antiAliasingRbox.SetBorderWidth(2.0, 1.0)
	antiAliasingRbox.SetTextSize(10.0, 0.0)
	antiAliasingRbox.SetTextThickness(1.2)

	// Set professional colors
	antiAliasingRbox.SetBackgroundColor(color.NewRGBA(0.96, 0.96, 0.96, 1.0)) // Light gray
	antiAliasingRbox.SetBorderColor(color.NewRGBA(0.4, 0.4, 0.4, 1.0))        // Medium gray
	antiAliasingRbox.SetTextColor(color.NewRGBA(0.2, 0.2, 0.2, 1.0))          // Dark gray
	antiAliasingRbox.SetInactiveColor(color.NewRGBA(0.6, 0.6, 0.6, 1.0))      // Gray
	antiAliasingRbox.SetActiveColor(color.NewRGBA(0.2, 0.6, 0.9, 1.0))        // Blue

	fmt.Printf("Anti-aliasing Settings:\n")
	fmt.Printf("Current selection: %s\n", antiAliasingRbox.ItemText(antiAliasingRbox.CurItem()))

	// In a real application, you would:
	// 1. Render this rbox control in a window or dialog
	// 2. Handle mouse events by calling OnMouseButtonDown/Up/Move
	// 3. Handle keyboard events by calling OnArrowKeys
	// 4. Use the vertex generation (Rewind/Vertex) to draw the control
	// 5. Use Color() to get colors for each rendering path
	// 6. Read the current selection to apply the chosen setting

	fmt.Println("In a real application, this would be rendered graphically and be interactive.")
}

func init() {
	// Ensure we can create controls without panicking
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("Demo failed with panic: %v", r)
		}
	}()
}
