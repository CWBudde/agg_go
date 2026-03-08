// Package main demonstrates the gamma correction control widget.
// This example shows how to use the GammaCtrl for interactive gamma curve editing.
package main

import (
	"fmt"
	"image"
	"image/color"

	agg "agg_go"
	"agg_go/examples/shared/demorunner"
	"agg_go/internal/ctrl/gamma"
)

// createSampleImage creates a test image with gradients for gamma correction demonstration.
func createSampleImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create different patterns in different sections
			section := x / (width / 4)

			switch section {
			case 0: // Horizontal gradient (grayscale)
				gray := uint8(255 * x / (width / 4))
				img.Set(x, y, color.RGBA{gray, gray, gray, 255})

			case 1: // Vertical gradient (red)
				red := uint8(255 * y / height)
				img.Set(x, y, color.RGBA{red, 0, 0, 255})

			case 2: // Diagonal gradient (green)
				green := uint8(255 * (x + y) / (width/4 + height))
				img.Set(x, y, color.RGBA{0, green, 0, 255})

			case 3: // Checkerboard pattern (blue variations)
				if ((x-3*width/4)/16+(y/16))%2 == 0 {
					img.Set(x, y, color.RGBA{0, 0, 255, 255})
				} else {
					img.Set(x, y, color.RGBA{128, 128, 255, 255})
				}

			default: // RGB gradient
				r := uint8(255 * x / width)
				g := uint8(255 * y / height)
				b := uint8(255 * (x + y) / (width + height))
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	return img
}

// applyGammaCorrection applies gamma correction to an image using the gamma control.
func applyGammaCorrection(img *image.RGBA, gammaCtrl *gamma.GammaCtrl) *image.RGBA {
	bounds := img.Bounds()
	corrected := image.NewRGBA(bounds)

	gammaTable := gammaCtrl.Gamma()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.RGBAAt(x, y)

			// Apply gamma correction to each color channel
			correctedColor := color.RGBA{
				R: gammaTable[originalColor.R],
				G: gammaTable[originalColor.G],
				B: gammaTable[originalColor.B],
				A: originalColor.A, // Keep alpha unchanged
			}

			corrected.Set(x, y, correctedColor)
		}
	}

	return corrected
}

type demo struct{}

func (d *demo) Render(ctx *agg.Context) {
	ctx.Clear(agg.White)

	// Create a gamma control widget
	gammaCtrl := gamma.NewGammaCtrl(10, 10, 300, 200, false)

	// Create sample image
	sampleImg := createSampleImage(400, 300)

	// Test different gamma curves
	testCases := []struct {
		name               string
		kx1, ky1, kx2, ky2 float64
	}{
		{"Identity", 1.0, 1.0, 1.0, 1.0},
		{"Brighten", 0.5, 1.5, 0.5, 1.5},
		{"Darken", 1.5, 0.5, 1.5, 0.5},
		{"High Contrast", 0.3, 1.8, 0.3, 1.8},
		{"Low Contrast", 1.8, 0.3, 1.8, 0.3},
		{"Custom 1", 0.8, 1.2, 1.2, 0.8},
		{"Extreme Bright", 0.1, 1.9, 0.1, 1.9},
		{"sRGB-like", 1.1, 0.9, 0.9, 1.1},
	}

	img := ctx.GetAgg2D()
	img.FillColor(agg.Black)
	img.TextAlignment(agg.AlignLeft, agg.AlignTop)

	for i, test := range testCases {
		gammaCtrl.Values(test.kx1, test.ky1, test.kx2, test.ky2)
		corrected := applyGammaCorrection(sampleImg, gammaCtrl)
		_ = corrected

		// Display test case info as text
		text := fmt.Sprintf("%d. %s (%.1f,%.1f,%.1f,%.1f)", i+1, test.name, test.kx1, test.ky1, test.kx2, test.ky2)
		img.Text(10, float64(20+i*18), text, false, 0, 0)
	}
}

func main() {
	demorunner.Run(demorunner.Config{
		Title:  "Gamma Correction Control",
		Width:  600,
		Height: 400,
	}, &demo{})
}
