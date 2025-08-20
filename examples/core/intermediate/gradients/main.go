// Package main demonstrates AGG2D gradient functionality.
// This example shows linear gradients, radial gradients, and multi-stop gradients.
package main

import (
	"fmt"

	agg "agg_go"
)

func main() {
	fmt.Println("AGG2D Gradient Demo")

	// Create rendering context
	agg2d := agg.NewAgg2D()

	// Setup rendering buffer (800x600, RGBA)
	width, height := 800, 600
	stride := width * 4
	buf := make([]uint8, height*stride)

	// Attach buffer to AGG2D context
	agg2d.Attach(buf, width, height, stride)

	// Clear background to white
	agg2d.ClearAll(agg.White)

	// Test 1: Linear Gradients
	fmt.Println("Creating linear gradients...")

	// Horizontal linear gradient (red to blue)
	agg2d.FillLinearGradient(50, 50, 250, 50, agg.Red, agg.Blue, 1.0)
	fmt.Printf("Set horizontal linear gradient: distance=%.1f\n", agg2d.FillGradientD2())

	// Vertical linear gradient (green to yellow)
	agg2d.FillLinearGradient(300, 50, 300, 150, agg.Green, agg.Yellow, 1.0)
	fmt.Printf("Set vertical linear gradient: distance=%.1f\n", agg2d.FillGradientD2())

	// Diagonal linear gradient (cyan to magenta)
	agg2d.FillLinearGradient(550, 50, 650, 150, agg.Cyan, agg.Magenta, 1.0)
	fmt.Printf("Set diagonal linear gradient: distance=%.1f\n", agg2d.FillGradientD2())

	// Test 2: Linear Gradient with Profile
	fmt.Println("Creating linear gradient with profile...")

	// Sharp profile gradient (profile = 0.5)
	agg2d.FillLinearGradient(50, 200, 250, 200, agg.Red, agg.Blue, 0.5)
	fmt.Printf("Sharp profile gradient: distance=%.1f\n", agg2d.FillGradientD2())

	// Normal profile gradient (profile = 1.0)
	agg2d.FillLinearGradient(300, 200, 500, 200, agg.Red, agg.Blue, 1.0)
	fmt.Printf("Normal profile gradient: distance=%.1f\n", agg2d.FillGradientD2())

	// Test 3: Radial Gradients
	fmt.Println("Creating radial gradients...")

	// Simple radial gradient (white center to black edge)
	agg2d.FillRadialGradient(150, 400, 75, agg.White, agg.Black, 1.0)
	fmt.Printf("White-to-black radial gradient: radius=%.1f, type=%v\n", agg2d.FillGradientD2(), agg2d.FillGradientFlag())

	// Colored radial gradient (red center to blue edge)
	agg2d.FillRadialGradient(400, 400, 75, agg.Red, agg.Blue, 1.0)
	fmt.Printf("Red-to-blue radial gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Radial gradient with sharp profile
	agg2d.FillRadialGradient(650, 400, 75, agg.Yellow, agg.Green, 0.3)
	fmt.Printf("Sharp radial gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Test 4: Multi-stop Radial Gradients
	fmt.Println("Creating multi-stop radial gradients...")

	// Three-color radial gradient (red -> green -> blue)
	agg2d.FillRadialGradientMultiStop(200, 530, 60, agg.Red, agg.Green, agg.Blue)
	fmt.Printf("RGB multi-stop gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Three-color radial gradient (yellow -> cyan -> magenta)
	agg2d.FillRadialGradientMultiStop(500, 530, 60, agg.Yellow, agg.Cyan, agg.Magenta)
	fmt.Printf("YCM multi-stop gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Test 5: Line Gradients
	fmt.Println("Creating line gradients...")

	// Linear line gradient
	agg2d.LineLinearGradient(50, 350, 250, 350, agg.Green, agg.Red, 1.0)
	fmt.Printf("Line linear gradient: distance=%.1f, type=%v\n", agg2d.LineGradientD2(), agg2d.LineGradientFlag())

	// Radial line gradient
	agg2d.LineRadialGradient(400, 350, 50, agg.Blue, agg.Yellow, 1.0)
	fmt.Printf("Line radial gradient: radius=%.1f\n", agg2d.LineGradientD2())

	// Test 6: Gradient Position Updates
	fmt.Println("Testing gradient position updates...")

	// Setup initial radial gradient
	agg2d.FillRadialGradient(100, 100, 30, agg.White, agg.Black, 1.0)
	fmt.Printf("Initial gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Update position and radius without changing colors
	agg2d.FillRadialGradientPos(700, 100, 50)
	fmt.Printf("Updated gradient: radius=%.1f\n", agg2d.FillGradientD2())

	// Test Color Interpolation
	fmt.Println("Testing color interpolation...")

	// Create a series of rectangles showing color interpolation
	red := agg.Red
	blue := agg.Blue

	for i := 0; i < 10; i++ {
		factor := float64(i) / 9.0
		interpolatedColor := red.Gradient(blue, factor)

		// In a complete implementation, we would set this interpolated color
		// and draw a rectangle. For now, we just demonstrate the calculation.
		_ = interpolatedColor

		// This would be: agg2d.FillColor(interpolatedColor)
		// agg2d.Rectangle(50 + i*20, 20, 70 + i*20, 40)
	}

	// Performance Test
	fmt.Println("Performance testing...")

	// Test gradient setup performance
	start := make(chan struct{})
	done := make(chan struct{})

	go func() {
		<-start
		for i := 0; i < 1000; i++ {
			agg2d.FillLinearGradient(0, 0, 100, 100, agg.Red, agg.Blue, 1.0)
		}
		done <- struct{}{}
	}()

	close(start)
	<-done
	fmt.Println("Performance test completed: 1000 linear gradient setups")

	// Summary
	fmt.Println("\nGradient Demo Summary:")
	fmt.Println("✓ Linear gradients (horizontal, vertical, diagonal)")
	fmt.Println("✓ Linear gradients with profile parameter")
	fmt.Println("✓ Radial gradients (simple and with profile)")
	fmt.Println("✓ Multi-stop radial gradients (3-color)")
	fmt.Println("✓ Line gradients (linear and radial)")
	fmt.Println("✓ Gradient position updates")
	fmt.Println("✓ Color interpolation")
	fmt.Println("✓ Performance testing")
	fmt.Println("")
	fmt.Printf("Rendering buffer created: %dx%d pixels (%d bytes)\n", width, height, len(buf))
	fmt.Println("Note: Actual rendering requires integration with the rendering pipeline.")
	fmt.Println("This demo shows gradient setup and configuration functionality.")
}
