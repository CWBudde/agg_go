// Package main demonstrates the AGG2D high-level interface.
// This is a Go port of the original C++ agg2d_demo.cpp that showcases
// the complete feature set of the AGG2D high-level rendering API.
package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	agg "agg_go"
)

// findSystemFont attempts to locate a usable system font
func findSystemFont() string {
	// Common font paths on Linux systems
	fontPaths := []string{
		// Ubuntu/Debian paths
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf",

		// Fedora/Red Hat paths
		"/usr/share/fonts/liberation-sans/LiberationSans-Regular.ttf",
		"/usr/share/fonts/dejavu-sans-fonts/DejaVuSans.ttf",

		// Arch Linux paths
		"/usr/share/fonts/TTF/liberation/LiberationSans-Regular.ttf",

		// Generic fallbacks
		"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",

		// Look for any TTF font as last resort
	}

	// Check if any of the known paths exist
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Using font: %s\n", path)
			return path
		}
	}

	// Try to find any TTF font in common directories
	fontDirs := []string{
		"/usr/share/fonts/",
		"/usr/local/share/fonts/",
		"/System/Library/Fonts/", // macOS
		"C:/Windows/Fonts/",      // Windows
	}

	for _, dir := range fontDirs {
		if fonts := findTTFFonts(dir); len(fonts) > 0 {
			fmt.Printf("Using font: %s\n", fonts[0])
			return fonts[0]
		}
	}

	// Fallback - this will likely fail but provides a clear error message
	fmt.Println("Warning: No system fonts found. Install liberation-fonts or dejavu-fonts package.")
	return "Arial" // Will fail, but with a clear error
}

// findTTFFonts searches for TTF fonts in a directory
func findTTFFonts(dir string) []string {
	var fonts []string

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fonts
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking despite errors
		}

		if filepath.Ext(path) == ".ttf" && !info.IsDir() {
			fonts = append(fonts, path)
			if len(fonts) >= 5 { // Limit search to avoid scanning too many fonts
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return nil
	}

	return fonts
}

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
	fmt.Println("AGG2D Demo - Go Port")
	fmt.Println("====================")

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

	// Regular Text - try to find a system font
	fontPath := findSystemFont()
	if err := agg2d.Font(fontPath, 14.0, false, false, agg.RasterFontCache, 0.0); err != nil {
		fmt.Printf("Warning: Could not load font %s: %v\n", fontPath, err)
	}
	agg2d.FillColor(agg.Black)
	agg2d.NoLine()
	agg2d.Text(100, 20, "Regular Raster Text -- Fast, but can't be rotated", false, 0, 0)

	// Outlined Text
	if err := agg2d.Font(fontPath, 50.0, false, false, agg.VectorFontCache, 0.0); err != nil {
		fmt.Printf("Warning: Could not load font %s: %v\n", fontPath, err)
	}
	agg2d.LineColor(agg.RGB(50.0/255, 0, 0))
	agg2d.FillColor(agg.RGB(180.0/255, 200.0/255, 100.0/255))
	agg2d.LineWidth(1.0)
	agg2d.Text(100.5, 50.5, "Outlined Text", false, 0, 0)

	// Text Alignment demonstration with reference lines
	drawAlignmentLines := func(x, y float64) {
		agg2d.LineColor(agg.RGB(0.7, 0.7, 0.7))
		agg2d.LineWidth(0.5)
		agg2d.Line(x-150, y, x+150, y)
		agg2d.Line(x, y-20, x, y+20)
	}

	// Set font for alignment demos
	if err := agg2d.Font(fontPath, 40.0, false, false, agg.VectorFontCache, 0.0); err != nil {
		fmt.Printf("Warning: Could not load font %s: %v\n", fontPath, err)
	}
	agg2d.FillColor(agg.RGB(100.0/255, 50.0/255, 50.0/255))
	agg2d.NoLine()

	positions := []struct {
		x, y           float64
		text           string
		alignX, alignY agg.TextAlignment
	}{
		{250, 150, "Left-Bottom", agg.AlignLeft, agg.AlignBottom},
		{250, 200, "Center-Bottom", agg.AlignCenter, agg.AlignBottom},
		{250, 250, "Right-Bottom", agg.AlignRight, agg.AlignBottom},
		{250, 300, "Left-Center", agg.AlignLeft, agg.AlignCenter},
		{250, 350, "Center-Center", agg.AlignCenter, agg.AlignCenter},
		{250, 400, "Right-Center", agg.AlignRight, agg.AlignCenter},
		{250, 450, "Left-Top", agg.AlignLeft, agg.AlignTop},
		{250, 500, "Center-Top", agg.AlignCenter, agg.AlignTop},
		{250, 550, "Right-Top", agg.AlignRight, agg.AlignTop},
	}

	for _, pos := range positions {
		drawAlignmentLines(pos.x, pos.y)
		agg2d.TextAlignment(pos.alignX, pos.alignY)
		agg2d.Text(pos.x, pos.y, pos.text, true, 0, 0)
	}

	// Gradients (Aqua Buttons)
	if err := agg2d.Font(fontPath, 20.0, false, false, agg.VectorFontCache, 0.0); err != nil {
		fmt.Printf("Warning: Could not load font %s: %v\n", fontPath, err)
	}

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

	agg2d.FillColor(agg.RGBA(0, 0, 50, 200))
	agg2d.NoLine()
	agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)
	agg2d.Text((xb1+xb2)/2.0, (yb1+yb2)/2.0, "Aqua Button", true, 0.0, 0.0)

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

	agg2d.FillColor(agg.RGBA(0, 0, 50, 255))
	agg2d.NoLine()
	agg2d.TextAlignment(agg.AlignCenter, agg.AlignCenter)
	agg2d.Text((xb1+xb2)/2.0, (yb1+yb2)/2.0, "Aqua Pressed", false, 0.0, 0.0)

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
	outputFile := "agg2d_demo.png"

	if err := saveAsPNG(finalImg, outputFile); err != nil {
		fmt.Printf("Error saving PNG: %v\n", err)
		return
	}

	fmt.Printf("AGG2D Demo completed successfully!\n")
	fmt.Printf("Output saved to: %s\n", outputFile)
	fmt.Println()
	fmt.Println("This demo demonstrates:")
	fmt.Println("  ✓ Viewport transformations and coordinate mapping")
	fmt.Println("  ✓ Text rendering with different alignments")
	fmt.Println("  ✓ Linear and radial gradients")
	fmt.Println("  ✓ Rounded rectangles and complex shapes")
	fmt.Println("  ✓ Path operations with arcs and curves")
	fmt.Println("  ✓ Blend modes (Add, Overlay, Alpha)")
	fmt.Println("  ✓ Master alpha for global transparency")
	fmt.Println("  ✓ Image transformations along paths")
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
