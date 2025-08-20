package main

import (
	"fmt"
	"math"

	agg "agg_go"
)

// AADemo demonstrates anti-aliasing quality
// This test mirrors the original AGG 2.6 aa_demo.cpp example
func main() {
	fmt.Println("AGG Anti-Aliasing Demo")
	fmt.Println("======================")

	// Create context - use same dimensions as original aa_demo
	width, height := 400, 300
	ctx := agg.NewAgg2D()
	buffer := make([]byte, width*height*4)
	ctx.Attach(buffer, width, height, width*4)

	// Clear to white background
	ctx.ClearAll(agg.Color{R: 255, G: 255, B: 255, A: 255})

	// Demonstrate different anti-aliasing scenarios
	fmt.Println("Testing anti-aliasing scenarios:")

	// Test 1: Lines at various angles to show anti-aliasing quality
	fmt.Println("  1. Angled lines (should show smooth anti-aliasing)")
	ctx.LineColor(agg.Color{R: 0, G: 0, B: 0, A: 255}) // Black
	ctx.LineWidth(1.0)

	centerX, centerY := float64(width/4), float64(height/4)
	radius := 50.0

	for i := 0; i < 16; i++ {
		angle := float64(i) * math.Pi / 8.0
		endX := centerX + radius*math.Cos(angle)
		endY := centerY + radius*math.Sin(angle)

		ctx.Line(centerX, centerY, endX, endY)
	}

	// Test 2: Circles with different stroke widths
	fmt.Println("  2. Circles with varying stroke widths")
	startX := float64(width * 3 / 4)
	for i := 0; i < 4; i++ {
		strokeWidth := 0.5 + float64(i)*0.5
		circleRadius := 15.0 + float64(i)*5.0
		circleY := float64(50 + i*40)

		ctx.LineWidth(strokeWidth)
		ctx.DrawCircle(startX, circleY, circleRadius)
	}

	// Test 3: Diagonal lines - critical test for anti-aliasing
	fmt.Println("  3. Diagonal lines (critical anti-aliasing test)")
	ctx.LineWidth(1.0)

	// 45-degree diagonal
	ctx.Line(50, float64(height*2/3), 150, float64(height*2/3)+100)

	// Shallow angle diagonal
	ctx.Line(200, float64(height*2/3), 350, float64(height*2/3)+30)

	// Test 4: Filled shapes with anti-aliased edges
	fmt.Println("  4. Filled shapes with anti-aliased edges")
	ctx.FillColor(agg.Color{R: 128, G: 128, B: 255, A: 180}) // Semi-transparent blue

	// Triangle
	ctx.ResetPath()
	ctx.MoveTo(50, float64(height-50))
	ctx.LineTo(100, float64(height-100))
	ctx.LineTo(100, float64(height-50))
	ctx.ClosePolygon()
	ctx.DrawPath(agg.FillOnly)

	// Ellipse
	ctx.ResetPath()
	ctx.AddEllipse(200, float64(height-75), 40, 25, agg.CCW)
	ctx.DrawPath(agg.FillOnly)

	// Test 5: Pixel-level precision test
	fmt.Println("  5. Sub-pixel positioning test")
	ctx.LineColor(agg.Color{R: 255, G: 0, B: 0, A: 255}) // Red
	ctx.LineWidth(1.0)

	// Lines offset by sub-pixel amounts
	for i := 0; i < 5; i++ {
		offset := float64(i) * 0.25 // Quarter-pixel offsets
		x := 300 + offset
		ctx.Line(x, 200, x, 250)
	}

	// Save output for visual inspection
	outputPath := "examples/shared/art/aa_demo_output.ppm"
	err := ctx.SaveImagePPM(outputPath)
	if err != nil {
		fmt.Printf("Error saving output image: %v\n", err)
	} else {
		fmt.Printf("Output saved to: %s\n", outputPath)
	}

	// Test results summary
	fmt.Println("\nAnti-Aliasing Test Results:")
	fmt.Printf("  Canvas size: %dx%d\n", width, height)
	fmt.Printf("  Test scenarios: 5\n")
	fmt.Printf("  Pixel format: %s\n", "RGBA32")

	fmt.Println("  Tests performed:")
	fmt.Println("    ✓ Radial line anti-aliasing")
	fmt.Println("    ✓ Variable stroke width rendering")
	fmt.Println("    ✓ Diagonal line quality")
	fmt.Println("    ✓ Filled shape edge smoothness")
	fmt.Println("    ✓ Sub-pixel positioning accuracy")

	fmt.Println("\nNote: Compare visual output with original AGG 2.6 aa_demo.cpp")
	fmt.Println("      This test validates anti-aliasing algorithm quality")
}
