// Package main demonstrates the basic usage of the AGG Go library.
// This example creates a simple image with a colored background.
package main

import (
	"fmt"

	agg "agg_go"
)

func main() {
	// Create a new rendering context
	width, height := 800, 600
	ctx := agg.NewContext(width, height)

	fmt.Printf("Created %dx%d rendering context\n", ctx.Width(), ctx.Height())

	// Clear the background to a light blue color
	ctx.Clear(agg.RGB(0.7, 0.8, 1.0))

	// Set drawing color to red
	ctx.SetColor(agg.Red)

	// Draw a simple rectangle
	ctx.DrawRectangle(100, 100, 200, 150)
	ctx.Fill()

	// Draw a circle
	ctx.SetColor(agg.RGB(0, 0.8, 0)) // Green
	ctx.DrawCircle(400, 300, 80)
	ctx.Fill()

	// Get the final image
	img := ctx.GetImage()

	fmt.Printf("Generated image: %dx%d, %d bytes\n",
		img.Width, img.Height, len(img.Data))

	// For now, just print some pixel values to verify it's working
	if len(img.Data) >= 16 {
		fmt.Printf("First few pixels: R=%d G=%d B=%d A=%d\n",
			img.Data[0], img.Data[1], img.Data[2], img.Data[3])
	}

	fmt.Println("Hello World example completed successfully!")
}
