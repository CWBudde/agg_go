// Package main demonstrates the spline control implementation.
// This example shows how to create and interact with a spline curve editor.
package main

import (
	"fmt"

	"agg_go/internal/color"
	"agg_go/internal/ctrl/spline"
)

func main() {
	// Create a spline control with RGBA colors
	ctrl := spline.NewSplineCtrlRGBA(10, 10, 300, 200, 6, false)

	// Set up some initial control points for a nice curve
	ctrl.SetPoint(1, 0.2, 0.8) // High peak near start
	ctrl.SetPoint(2, 0.4, 0.3) // Dip in middle
	ctrl.SetPoint(3, 0.6, 0.7) // Another peak
	ctrl.SetPoint(4, 0.8, 0.2) // Low end

	// Demonstrate spline value calculation
	fmt.Println("Spline Control Demo")
	fmt.Println("==================")
	fmt.Printf("Number of control points: %d\n", 6)
	fmt.Printf("Control bounds: (%.1f, %.1f) to (%.1f, %.1f)\n",
		ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())

	// Show some spline values at different X positions
	fmt.Println("\nSpline values at different X positions:")
	for x := 0.0; x <= 1.0; x += 0.2 {
		y := ctrl.Value(x)
		fmt.Printf("  x=%.1f  ->  y=%.3f\n", x, y)
	}

	// Show control point positions
	fmt.Println("\nControl point positions:")
	for i := uint(0); i < 6; i++ {
		x := ctrl.GetPointX(i)
		y := ctrl.GetPointY(i)
		fmt.Printf("  Point %d: (%.3f, %.3f)\n", i, x, y)
	}

	// Demonstrate mouse interaction simulation
	fmt.Println("\nSimulating mouse interaction...")

	// Simulate clicking on control point 2
	testX := ctrl.GetPointX(2)*(ctrl.X2()-ctrl.X1()) + ctrl.X1()
	testY := ctrl.GetPointY(2)*(ctrl.Y2()-ctrl.Y1()) + ctrl.Y1()
	fmt.Printf("Simulating click at control point 2 screen coordinates: (%.1f, %.1f)\n", testX, testY)

	redraw := ctrl.OnMouseButtonDown(testX, testY)
	fmt.Printf("Mouse down returned: %t (should be true for hit)\n", redraw)
	fmt.Printf("Active point is now: %d (should be 2)\n", ctrl.GetActivePoint())

	// Simulate dragging the point
	newX := testX + 10
	newY := testY - 15
	redraw = ctrl.OnMouseMove(newX, newY, true)
	fmt.Printf("Mouse move returned: %t (should be true for drag)\n", redraw)
	fmt.Printf("Point 2 new position: (%.3f, %.3f)\n", ctrl.GetPointX(2), ctrl.GetPointY(2))

	// Release mouse
	redraw = ctrl.OnMouseButtonUp(newX, newY)
	fmt.Printf("Mouse up returned: %t (should be true)\n", redraw)

	// Demonstrate keyboard navigation
	fmt.Println("\nTesting keyboard navigation...")
	redraw = ctrl.OnArrowKeys(false, true, false, false) // Right arrow
	fmt.Printf("Right arrow returned: %t (should be true)\n", redraw)
	fmt.Printf("Point 2 position after right arrow: (%.3f, %.3f)\n", ctrl.GetPointX(2), ctrl.GetPointY(2))

	// Demonstrate color customization
	fmt.Println("\nColor customization:")
	ctrl.SetCurveColor(color.NewRGBA(0.0, 0.8, 0.0, 1.0))       // Green curve
	ctrl.SetActivePointColor(color.NewRGBA(0.8, 0.4, 0.0, 1.0)) // Orange active point

	fmt.Println("Curve color set to green, active point color set to orange")

	// Show final spline values after modifications
	fmt.Println("\nFinal spline values after interaction:")
	for x := 0.0; x <= 1.0; x += 0.2 {
		y := ctrl.Value(x)
		fmt.Printf("  x=%.1f  ->  y=%.3f\n", x, y)
	}

	fmt.Println("\nDemo completed successfully!")
}
