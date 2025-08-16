// Package main demonstrates basic grayscale color handling in AGG Go.
package main

import (
	"fmt"
	"os"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/pixfmt"
)

func main() {
	fmt.Println("AGG Go - Grayscale Color Example")
	fmt.Println("=================================")

	// Create a rendering buffer
	width, height := 100, 50
	buf := make([]basics.Int8u, width*height)
	rbuf := buffer.NewRenderingBufferU8WithData(buf, width, height, width)

	// Create a grayscale pixel format
	pf := pixfmt.NewPixFmtGray8(rbuf)

	// Clear the buffer with a medium gray
	clearColor := color.NewGray8WithAlpha[color.Linear](128, 255)
	pf.Clear(clearColor)

	// Draw some patterns to demonstrate grayscale operations
	drawGrayscalePatterns(pf)

	// Output a text representation
	printGrayscaleBuffer(pf, width, height)

	// Demonstrate color conversions
	demonstrateColorConversions()

	// Demonstrate blending
	demonstrateBlendingMath()

	// Demonstrate visual blending in the buffer
	demonstrateBlending(pf)

	fmt.Println("\nGrayscale example completed successfully!")
}

func drawGrayscalePatterns(pf *pixfmt.PixFmtGray8) {
	width, height := pf.Width(), pf.Height()

	// Draw a gradient from left to right
	for x := 0; x < width/2; x++ {
		// Calculate gradient value (0-255)
		value := basics.Int8u(float64(x) / float64(width/2) * 255)
		grayColor := color.NewGray8WithAlpha[color.Linear](value, 255)
		pf.CopyVline(x, 0, height/2-1, grayColor)
	}

	// Draw some rectangles with different gray levels
	colors := []basics.Int8u{64, 128, 192, 255}
	rectWidth := width / len(colors)

	for i, grayValue := range colors {
		x1 := i * rectWidth
		x2 := x1 + rectWidth - 1
		y1 := height/2 + 5
		y2 := height - 5

		grayColor := color.NewGray8WithAlpha[color.Linear](grayValue, 255)
		pf.CopyBar(x1, y1, x2, y2, grayColor)
	}

	// Draw some blended pixels
	blendColor := color.NewGray8WithAlpha[color.Linear](200, 128) // 50% alpha
	for x := width / 2; x < width; x += 2 {
		for y := 0; y < height/2; y += 2 {
			pf.BlendPixel(x, y, blendColor, 255)
		}
	}
}

func printGrayscaleBuffer(pf *pixfmt.PixFmtGray8, width, height int) {
	fmt.Println("\nGrayscale Buffer Contents (scaled to ASCII):")
	fmt.Println("============================================")

	// ASCII characters representing different gray levels
	chars := " .:-=+*#%@"

	for y := 0; y < height; y += 2 { // Skip every other row for readability
		for x := 0; x < width; x += 2 { // Skip every other column
			pixel := pf.GetPixel(x, y)
			// Convert grayscale value to character index
			charIndex := int(pixel.V) * (len(chars) - 1) / 255
			if charIndex >= len(chars) {
				charIndex = len(chars) - 1
			}
			fmt.Print(string(chars[charIndex]))
		}
		fmt.Println()
	}
}

func demonstrateColorConversions() {
	fmt.Println("\nColor Conversion Examples:")
	fmt.Println("==========================")

	// Create an RGBA color
	rgba := color.NewRGBA(0.6, 0.3, 0.1, 0.8) // Brownish color
	fmt.Printf("Original RGBA: R=%.2f, G=%.2f, B=%.2f, A=%.2f\n", rgba.R, rgba.G, rgba.B, rgba.A)

	// Convert to grayscale
	gray := color.ConvertGray8FromRGBA[color.Linear](rgba)
	fmt.Printf("Converted to Gray8: V=%d, A=%d\n", gray.V, gray.A)

	// Convert back to RGBA
	rgbaBack := gray.ConvertToRGBA()
	fmt.Printf("Converted back to RGBA: R=%.2f, G=%.2f, B=%.2f, A=%.2f\n",
		rgbaBack.R, rgbaBack.G, rgbaBack.B, rgbaBack.A)

	// Demonstrate different gray types
	gray16 := color.ConvertGray16FromRGBA[color.Linear](rgba)
	fmt.Printf("As Gray16: V=%d, A=%d\n", gray16.V, gray16.A)

	gray32 := color.ConvertGray32FromRGBA[color.Linear](rgba)
	fmt.Printf("As Gray32: V=%.3f, A=%.3f\n", gray32.V, gray32.A)

	// Demonstrate colorspace conversion
	fmt.Println("\nColorspace Conversions:")
	linearGray := color.NewGray8WithAlpha[color.Linear](128, 255)
	srgbGray := color.ConvertGray8LinearToSRGB(linearGray)
	fmt.Printf("Linear Gray8(128) -> sRGB Gray8(%d)\n", srgbGray.V)

	backToLinear := color.ConvertGray8SRGBToLinear(srgbGray)
	fmt.Printf("sRGB Gray8(%d) -> Linear Gray8(%d)\n", srgbGray.V, backToLinear.V)
}

func demonstrateBlendingMath() {
	fmt.Println("\nBlending Examples:")
	fmt.Println("==================")

	// Create some test colors
	background := color.NewGray8WithAlpha[color.Linear](100, 255)
	foreground := color.NewGray8WithAlpha[color.Linear](200, 128) // 50% alpha

	fmt.Printf("Background: V=%d, A=%d\n", background.V, background.A)
	fmt.Printf("Foreground: V=%d, A=%d (50%% alpha)\n", foreground.V, foreground.A)

	// Manual blending calculation
	blended := color.Gray8Lerp(background.V, foreground.V, foreground.A)
	fmt.Printf("Blended result: V=%d\n", blended)

	// Demonstrate premultiplication
	premult := foreground
	premult.Premultiply()
	fmt.Printf("Premultiplied foreground: V=%d, A=%d\n", premult.V, premult.A)

	premult.Demultiply()
	fmt.Printf("After demultiply: V=%d, A=%d\n", premult.V, premult.A)

	// Demonstrate gradient
	color1 := color.NewGray8WithAlpha[color.Linear](0, 255)
	color2 := color.NewGray8WithAlpha[color.Linear](255, 255)
	gradient := color1.Gradient(color2, 0.3) // 30% towards color2
	fmt.Printf("Gradient (30%% from black to white): V=%d, A=%d\n", gradient.V, gradient.A)
}

func demonstrateBlending(pf *pixfmt.PixFmtGray8) {
	width := pf.Width()

	// Create a section for blending demonstration
	startY := pf.Height() - 10
	if startY < 0 {
		return
	}

	// Fill background
	bgColor := color.NewGray8WithAlpha[color.Linear](80, 255)
	pf.CopyBar(0, startY, width-1, pf.Height()-1, bgColor)

	// Blend some shapes with different alpha values
	alphas := []basics.Int8u{64, 128, 192, 255}
	rectWidth := width / len(alphas)

	for i, alpha := range alphas {
		x1 := i * rectWidth
		x2 := x1 + rectWidth - 1
		y1 := startY + 2
		y2 := pf.Height() - 3

		blendColor := color.NewGray8WithAlpha[color.Linear](220, alpha)
		pf.BlendBar(x1, y1, x2, y2, blendColor, 255)
	}
}

// Check if this is being run as main
func init() {
	if len(os.Args) > 0 && os.Args[0] != "go" {
		// This is likely being run directly, not through go test
		// You can add any initialization here if needed
	}
}
