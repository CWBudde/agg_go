// Package main demonstrates the basic usage of the AGG Go library.
// This example creates a simple image with a colored background.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	agg "agg_go"
)

func savePNG(img *agg.Image, filename string) error {
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))
	copy(goImg.Pix, img.Data)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, goImg)
}

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
		img.Width(), img.Height(), len(img.Data))
	if err := savePNG(img, "hello_world.png"); err != nil {
		fmt.Printf("Failed to save hello_world.png: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Hello World example completed successfully!")
	fmt.Println("Output saved to hello_world.png")
}
