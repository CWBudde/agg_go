//go:build agg2d_demo_simple
// +build agg2d_demo_simple

// Package main demonstrates the AGG2D high-level interface.
// This is a simplified version that avoids text rendering to focus on graphics features.
package main

import (
	"fmt"
	"image/png"
	"os"

	agg "agg_go"
)

// createImageFromFile creates a test image (since we don't have spheres.bmp)
func createImageFromFile() *agg.Image {
	// Create a simple test pattern image
	width, height := 100, 100
	stride := width * 4
	buf := make([]uint8, height*stride)

	// Create a simple gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*stride + x*4
			buf[idx] = uint8((x * 255) / width)    // R
			buf[idx+1] = uint8((y * 255) / height) // G
			buf[idx+2] = 128                       // B
			buf[idx+3] = 255                       // A
		}
	}

	return agg.NewImage(buf, width, height, stride)
}

func main() {
	fmt.Println("AGG2D Demo - Go Port (Graphics-focused version)")
	fmt.Println("===============================================")

	// Create rendering context
	width, height := 600, 600
	ctx := agg.NewContext(width, height)
	agg2d := ctx.GetAgg2D()

	// Clear background to white
	ctx.Clear(agg.White)

	// Set viewport - scale 0,0,600,600 to the actual window size
	// preserving aspect ratio and placing the viewport in the center
	agg2d.Viewport(0, 0, 600, 600, 0, 0, float64(width), float64(height), agg.XMidYMid)

	// Rounded rectangle border
	agg2d.LineColor(agg.Black)
	agg2d.NoFill()
	agg2d.RoundedRect(0.5, 0.5, 599.5, 599.5, 20.0)

	// Draw informational text placeholder (without actual font rendering)
	agg2d.FillColor(agg.RGB(0.8, 0.8, 0.8))
	agg2d.NoLine()
	// Text placeholder rectangles
	agg2d.Rectangle(100, 15, 500, 25)     // "Regular Raster Text" placeholder
	agg2d.Rectangle(100.5, 45.5, 300, 65) // "Outlined Text" placeholder

	// Gradients (Aqua Buttons)
	// Aqua Button Normal
	xb1, yb1 := 400.0, 80.0
	xb2, yb2 := xb1+150, yb1+36

	agg2d.FillColor(agg.RGBA(0, 50, 180, 180))
	agg2d.LineColor(agg.RGBA(0, 0, 80, 255))
	agg2d.LineWidth(1.0)
	agg2d.RoundedRect(xb1, yb1, xb2, yb2, 12)

	agg2d.LineColor(agg.RGBA(0, 0, 0, 0))
	agg2d.FillLinearGradient(xb1, yb1, xb1, yb1+30,
		agg.RGBA(100, 200, 255, 255),
		agg.RGBA(255, 255, 255, 0), 1.0)
	agg2d.RoundedRect(xb1+3, yb1+2.5, xb2-3, yb1+30, 9)

	// Text placeholder for button
	agg2d.FillColor(agg.RGBA(0, 0, 50, 200))
	agg2d.NoLine()
	agg2d.Rectangle((xb1+xb2)/2.0-30, (yb1+yb2)/2.0-5, (xb1+xb2)/2.0+30, (yb1+yb2)/2.0+5)

	agg2d.FillLinearGradient(xb1, yb2-20, xb1, yb2-3,
		agg.RGBA(0, 0, 255, 0),
		agg.RGBA(100, 255, 255, 255), 1.0)
	agg2d.RoundedRect(xb1+3, yb2-20, xb2-3, yb2-2, 9)

	// Aqua Button Pressed
	xb1, yb1 = 400, 30
	xb2, yb2 = xb1+150, yb1+36

	agg2d.FillColor(agg.RGBA(0, 50, 180, 180))
	agg2d.LineColor(agg.RGBA(0, 0, 0, 255))
	agg2d.LineWidth(2.0)
	agg2d.RoundedRect(xb1, yb1, xb2, yb2, 12)

	agg2d.LineColor(agg.RGBA(0, 0, 0, 0))
	agg2d.FillLinearGradient(xb1, yb1+2, xb1, yb1+25,
		agg.RGBA(60, 160, 255, 255),
		agg.RGBA(100, 255, 255, 0), 1.0)
	agg2d.RoundedRect(xb1+3, yb1+2.5, xb2-3, yb1+30, 9)

	// Text placeholder for pressed button
	agg2d.FillColor(agg.RGBA(0, 0, 50, 255))
	agg2d.NoLine()
	agg2d.Rectangle((xb1+xb2)/2.0-35, (yb1+yb2)/2.0-5, (xb1+xb2)/2.0+35, (yb1+yb2)/2.0+5)

	agg2d.FillLinearGradient(xb1, yb2-25, xb1, yb2-5,
		agg.RGBA(0, 180, 255, 0),
		agg.RGBA(0, 200, 255, 255), 1.0)
	agg2d.RoundedRect(xb1+3, yb2-25, xb2-3, yb2-2, 9)

	// Basic Shapes -- Ellipse
	agg2d.LineWidth(3.5)
	agg2d.LineColor(agg.RGB(20.0/255, 80.0/255, 80.0/255))
	agg2d.FillColor(agg.RGBA(200, 255, 80, 200))
	agg2d.Ellipse(450, 200, 50, 90)

	// Paths - Arc demonstrations
	agg2d.ResetPath()
	agg2d.FillColor(agg.RGBA(255, 0, 0, 100))
	agg2d.LineColor(agg.RGBA(0, 0, 255, 100))
	agg2d.LineWidth(2)
	agg2d.MoveTo(300/2, 200/2)
	agg2d.HorLineRel(-150 / 2)
	agg2d.ArcRel(150/2, 150/2, 0, true, false, 150/2, -150/2)
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillAndStroke)

	agg2d.ResetPath()
	agg2d.FillColor(agg.RGBA(255, 255, 0, 100))
	agg2d.LineColor(agg.RGBA(0, 0, 255, 100))
	agg2d.LineWidth(2)
	agg2d.MoveTo(275/2, 175/2)
	agg2d.VerLineRel(-150 / 2)
	agg2d.ArcRel(150/2, 150/2, 0, false, false, -150/2, 150/2)
	agg2d.ClosePolygon()
	agg2d.DrawPath(agg.FillAndStroke)

	// Complex path with multiple arcs
	agg2d.ResetPath()
	agg2d.NoFill()
	agg2d.LineColor(agg.RGB(127.0/255, 0, 0))
	agg2d.LineWidth(5)
	agg2d.MoveTo(600/2, 350/2)
	agg2d.LineRel(50/2, -25/2)
	agg2d.ArcRel(25/2, 25/2, agg.Deg2RadFunc(-30), false, true, 50/2, -25/2)
	agg2d.LineRel(50/2, -25/2)
	agg2d.ArcRel(25/2, 50/2, agg.Deg2RadFunc(-30), false, true, 50/2, -25/2)
	agg2d.LineRel(50/2, -25/2)
	agg2d.ArcRel(25/2, 75/2, agg.Deg2RadFunc(-30), false, true, 50/2, -25/2)
	agg2d.LineRel(50, -25)
	agg2d.ArcRel(25/2, 100/2, agg.Deg2RadFunc(-30), false, true, 50/2, -25/2)
	agg2d.LineRel(50/2, -25/2)
	agg2d.DrawPath(agg.StrokeOnly)

	// Master Alpha - from now on everything will be translucent
	agg2d.MasterAlpha(0.85)

	// Create a test image for transformations
	img := createImageFromFile()

	// Transform image to destination path
	agg2d.ResetPath()
	agg2d.MoveTo(450, 200)
	agg2d.CubicCurveTo(595, 220, 575, 350, 595, 350)
	agg2d.LineTo(470, 340)
	if err := agg2d.TransformImagePath(img, 10, 10, img.Width()-10, img.Height()-10,
		450, 200, 595, 350); err != nil {
		fmt.Printf("Warning: Image transformation failed: %v\n", err)
	}

	// Add/Sub/Contrast Blending Modes
	agg2d.NoLine()
	agg2d.FillColor(agg.RGB(70.0/255, 70.0/255, 0))
	agg2d.BlendMode(agg.BlendAdd)
	agg2d.Ellipse(500, 280, 20, 40)

	agg2d.FillColor(agg.White)
	agg2d.BlendMode(agg.BlendOverlay)
	agg2d.Ellipse(500+40, 280, 20, 40)

	// Radial gradient
	agg2d.BlendMode(agg.BlendAlpha)
	agg2d.FillRadialGradient(400, 500, 40,
		agg.RGBA(255, 255, 0, 0),
		agg.RGBA(0, 0, 127, 255), 1.0)
	agg2d.Ellipse(400, 500, 40, 40)

	// Get the final image and save
	finalImg := ctx.GetImage()
	outputFile := "agg2d_demo_graphics.png"

	if err := saveAsPNG(finalImg, outputFile); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Printf("AGG2D Graphics Demo completed successfully!\n")
	fmt.Printf("Output saved to: %s\n", outputFile)
	fmt.Println()
	fmt.Println("This demo demonstrates:")
	fmt.Println("  ✓ Viewport transformations and coordinate mapping")
	fmt.Println("  ✓ Linear and radial gradients")
	fmt.Println("  ✓ Rounded rectangles and complex shapes")
	fmt.Println("  ✓ Path operations with arcs and curves")
	fmt.Println("  ✓ Blend modes (Add, Overlay, Alpha)")
	fmt.Println("  ✓ Master alpha for global transparency")
	fmt.Println("  ✓ Image transformations along paths")
	fmt.Println()
	fmt.Println("Note: Text rendering requires FreeType integration debugging.")
	fmt.Println("This version focuses on the graphics capabilities of AGG2D.")
}

// saveAsPNG saves an AGG image as PNG
func saveAsPNG(aggImg *agg.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Convert AGG image to Go image
	goImg := aggImg.ToGoImage()

	// Save as PNG
	return png.Encode(file, goImg)
}
