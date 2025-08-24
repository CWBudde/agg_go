// Simplified rounded rectangle example demonstrating AGG's rendering pipeline
// This is a non-interactive version of the original AGG rounded_rect.cpp demo
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	agg "agg_go"
)

func main() {
	// Canvas dimensions
	const width, height = 640, 480

	fmt.Println("AGG Go - Simplified Rounded Rectangle Demo")
	fmt.Printf("Creating %dx%d canvas with rounded rectangles...\n", width, height)

	// Create rendering context using the high-level API
	ctx := agg.NewContext(width, height)

	// Clear background to white
	ctx.Clear(agg.White)

	// Demo 1: Blue filled rounded rectangle
	ctx.SetColor(agg.Blue)
	ctx.FillRoundedRectangle(120, 100, 160, 100, 30)

	// Demo 2: Red outlined rounded rectangle
	ctx.SetColor(agg.Red)
	ctx.DrawRoundedRectangle(370, 100, 160, 100, 30)

	// Demo 3: Green filled rounded rectangle with different proportions
	ctx.SetColor(agg.Green)
	ctx.FillRoundedRectangle(120, 260, 160, 120, 40)

	// Demo 4: Purple outlined rounded rectangle
	ctx.SetColor(agg.RGB(0.6, 0.4, 0.8)) // Purple
	ctx.DrawRoundedRectangle(370, 260, 160, 120, 40)

	// Demo 5: Orange very rounded rectangle (pill shape)
	ctx.SetColor(agg.RGB(1.0, 0.6, 0.0)) // Orange
	ctx.FillRoundedRectangle(220, 400, 200, 70, 35)

	// Get the image and save as PNG
	img := ctx.GetImage()
	outputFile := "rounded_rect_demo.png"

	if err := saveAsPNG(img, outputFile); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Printf("Demo completed! Output saved to: %s\n", outputFile)
	fmt.Println("The demo shows:")
	fmt.Println("  - Blue filled rounded rectangle")
	fmt.Println("  - Red outlined rounded rectangle")
	fmt.Println("  - Green filled rounded rectangle")
	fmt.Println("  - Purple outlined rounded rectangle")
	fmt.Println("  - Orange filled pill-shaped rounded rectangle")

	fmt.Println("\nNote: This example uses pixel-level rendering for rounded rectangles.")
	fmt.Println("It does not yet use the full scanline/rasterizer pipeline.")
}

// saveAsPNG converts the AGG image to PNG format
func saveAsPNG(img *agg.Image, filename string) error {
	// Create Go image from AGG image
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))

	// Copy pixel data from AGG buffer to Go image
	copy(goImg.Pix, img.Data)

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Encode as PNG
	return png.Encode(file, goImg)
}
