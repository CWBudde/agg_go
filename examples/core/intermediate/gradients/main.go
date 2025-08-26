// Package main demonstrates AGG gradients functionality with actual rendering.
// This example creates various gradient-filled shapes and saves them to PNG files.
package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	agg "agg_go"
)

const (
	width  = 800
	height = 600
)

// saveToImage saves an AGG image as a PNG file
func saveToImage(img *agg.Image, filename string) error {
	// Create Go image from AGG image
	goImg := image.NewRGBA(image.Rect(0, 0, img.Width(), img.Height()))

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

// renderGradientDemo renders various gradient examples
func renderGradientDemo() error {
	fmt.Println("AGG Gradient Rendering Demo")

	// Create rendering context using high-level API
	ctx := agg.NewContext(width, height)

	// Clear background to black
	ctx.Clear(agg.Black)

	// Demo 1: Linear gradients
	fmt.Println("Rendering linear gradients...")

	// Horizontal linear gradient (red to blue)
	ctx.SetLinearGradient(50, 75, 200, 75, agg.Red, agg.Blue)
	ctx.FillRectangle(50, 50, 150, 50)

	// Vertical linear gradient (green to yellow)
	ctx.SetLinearGradient(250, 50, 250, 150, agg.Green, agg.Yellow)
	ctx.FillRectangle(225, 50, 50, 100)

	// Diagonal linear gradient (cyan to magenta)
	ctx.SetLinearGradient(320, 50, 420, 150, agg.Cyan, agg.Magenta)
	ctx.FillRectangle(300, 50, 120, 100)

	// Demo 2: Linear Gradient with Profile
	fmt.Println("Rendering linear gradients with different profiles...")

	// Sharp profile gradient (profile = 0.5)
	ctx.SetLinearGradientWithProfile(480, 50, 630, 50, agg.Red, agg.Blue, 0.5)
	ctx.FillRectangle(450, 50, 180, 50)

	// Normal profile gradient (profile = 1.0)
	ctx.SetLinearGradientWithProfile(480, 120, 630, 120, agg.Red, agg.Blue, 1.0)
	ctx.FillRectangle(450, 100, 180, 50)

	// Demo 3: Radial gradients
	fmt.Println("Rendering radial gradients...")

	// Simple radial gradient (white center to black edge)
	ctx.SetRadialGradient(100, 250, 60, agg.White, agg.Black)
	ctx.FillEllipse(100, 250, 60, 60)

	// Colored radial gradient (red center to blue edge)
	ctx.SetRadialGradient(250, 250, 60, agg.Red, agg.Blue)
	ctx.FillEllipse(250, 250, 60, 60)

	// Radial gradient with sharp profile
	ctx.SetRadialGradientWithProfile(400, 250, 60, agg.Yellow, agg.Green, 0.3)
	ctx.FillEllipse(400, 250, 60, 60)

	// Demo 4: Multi-stop Radial Gradients
	fmt.Println("Rendering multi-stop radial gradients...")

	// Three-color radial gradient (red -> green -> blue)
	ctx.SetRadialGradientMultiStop(100, 400, 50, agg.Red, agg.Green, agg.Blue)
	ctx.FillEllipse(100, 400, 50, 50)

	// Three-color radial gradient (yellow -> cyan -> magenta)
	ctx.SetRadialGradientMultiStop(250, 400, 50, agg.Yellow, agg.Cyan, agg.Magenta)
	ctx.FillEllipse(250, 400, 50, 50)

	// Demo 5: Stroke/Line gradients
	fmt.Println("Rendering line gradients...")

	// Linear stroke gradient
	ctx.SetLineWidth(8.0)
	ctx.SetStrokeLinearGradient(450, 200, 650, 200, agg.Green, agg.Red)
	ctx.MoveTo(450, 200)
	ctx.LineTo(650, 200)
	ctx.Stroke()

	// Radial stroke gradient - draw a circle outline
	ctx.SetLineWidth(6.0)
	ctx.SetStrokeRadialGradient(550, 300, 40, agg.Blue, agg.Yellow)
	ctx.DrawEllipse(550, 300, 40, 40)

	// Demo 6: Mixed shapes with gradients
	fmt.Println("Rendering mixed shapes with gradients...")

	// Rounded rectangle with linear gradient
	ctx.SetLinearGradient(450, 380, 650, 480, agg.RGB(1.0, 0.5, 0.0), agg.RGB(0.5, 0.0, 1.0))
	ctx.FillRoundedRectangle(450, 380, 200, 100, 20)

	// Rectangle outline with gradient
	ctx.SetLineWidth(4.0)
	ctx.SetStrokeLinearGradient(50, 480, 200, 580, agg.RGB(0.0, 1.0, 0.5), agg.RGB(1.0, 0.0, 0.5))
	ctx.DrawRoundedRectangle(50, 480, 150, 100, 15)

	// Demo 7: Color interpolation test
	fmt.Println("Testing color interpolation...")

	// Create a series of rectangles showing color interpolation
	red := agg.Red
	blue := agg.Blue

	for i := 0; i < 10; i++ {
		factor := float64(i) / 9.0
		interpolatedColor := red.Gradient(blue, factor)

		ctx.SetColor(interpolatedColor)
		ctx.FillRectangle(50+float64(i)*20, 15, 18, 25)
	}

	// Get the image and save as PNG
	img := ctx.GetImage()
	fmt.Println("Saving gradient demo to gradient_demo.png...")
	return saveToImage(img, "gradient_demo.png")
}

func main() {
	if err := renderGradientDemo(); err != nil {
		fmt.Printf("Error rendering gradients: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Gradient demo completed successfully!")
	fmt.Println("Output saved to gradient_demo.png")
	fmt.Println()
	fmt.Println("Demo features:")
	fmt.Println("✓ Linear gradients (horizontal, vertical, diagonal)")
	fmt.Println("✓ Linear gradients with profile parameter")
	fmt.Println("✓ Radial gradients (simple and with profile)")
	fmt.Println("✓ Multi-stop radial gradients (3-color)")
	fmt.Println("✓ Stroke gradients (linear and radial)")
	fmt.Println("✓ Mixed shapes with gradients")
	fmt.Println("✓ Color interpolation")
}
