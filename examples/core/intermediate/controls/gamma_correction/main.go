// Package main demonstrates the gamma correction control widget.
// This example shows how to use the GammaCtrl for interactive gamma curve editing.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

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

// saveImage saves an image to a PNG file.
func saveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// demonstrateGammaControl shows various gamma correction curves and their effects.
func demonstrateGammaControl() {
	fmt.Println("=== AGG Gamma Correction Control Demo ===")
	fmt.Println()

	// Create a gamma control widget
	gammaCtrl := gamma.NewGammaCtrl(10, 10, 300, 200, false)

	// Create sample image
	fmt.Println("Creating sample image...")
	sampleImg := createSampleImage(400, 300)

	// Save original image
	if err := saveImage(sampleImg, "original.png"); err != nil {
		log.Printf("Warning: Could not save original image: %v", err)
	} else {
		fmt.Println("Saved original.png")
	}

	// Test different gamma curves
	testCases := []struct {
		name               string
		kx1, ky1, kx2, ky2 float64
		description        string
		filename           string
	}{
		{
			"Identity",
			1.0, 1.0, 1.0, 1.0,
			"No gamma correction (identity curve)",
			"gamma_identity.png",
		},
		{
			"Brighten",
			0.5, 1.5, 0.5, 1.5,
			"Brightening curve (lifts shadows)",
			"gamma_bright.png",
		},
		{
			"Darken",
			1.5, 0.5, 1.5, 0.5,
			"Darkening curve (deepens shadows)",
			"gamma_dark.png",
		},
		{
			"High Contrast",
			0.3, 1.8, 0.3, 1.8,
			"High contrast curve (S-curve effect)",
			"gamma_contrast.png",
		},
		{
			"Low Contrast",
			1.8, 0.3, 1.8, 0.3,
			"Low contrast curve (flattened response)",
			"gamma_low_contrast.png",
		},
		{
			"Custom 1",
			0.8, 1.2, 1.2, 0.8,
			"Custom curve (mixed adjustment)",
			"gamma_custom1.png",
		},
		{
			"Extreme Bright",
			0.1, 1.9, 0.1, 1.9,
			"Extreme brightening (very aggressive)",
			"gamma_extreme_bright.png",
		},
		{
			"sRGB-like",
			1.1, 0.9, 0.9, 1.1,
			"sRGB-like gamma curve (standard display)",
			"gamma_srgb.png",
		},
	}

	fmt.Println()
	fmt.Println("Testing gamma correction curves:")
	fmt.Println("=================================")

	for i, test := range testCases {
		fmt.Printf("%d. %s\n", i+1, test.name)
		fmt.Printf("   Control points: kx1=%.3f, ky1=%.3f, kx2=%.3f, ky2=%.3f\n",
			test.kx1, test.ky1, test.kx2, test.ky2)
		fmt.Printf("   Description: %s\n", test.description)

		// Set gamma curve
		gammaCtrl.Values(test.kx1, test.ky1, test.kx2, test.ky2)

		// Verify the values were set correctly
		gotKx1, gotKy1, gotKx2, gotKy2 := gammaCtrl.GetValues()
		fmt.Printf("   Actual values: kx1=%.3f, ky1=%.3f, kx2=%.3f, ky2=%.3f\n",
			gotKx1, gotKy1, gotKx2, gotKy2)

		// Test gamma function at key points
		fmt.Print("   Gamma curve samples: ")
		for j := 0; j <= 4; j++ {
			x := float64(j) / 4.0
			y := gammaCtrl.Y(x)
			fmt.Printf("Y(%.2f)=%.3f ", x, y)
		}
		fmt.Println()

		// Apply gamma correction
		correctedImg := applyGammaCorrection(sampleImg, gammaCtrl)

		// Save corrected image
		if err := saveImage(correctedImg, test.filename); err != nil {
			log.Printf("Warning: Could not save %s: %v", test.filename, err)
		} else {
			fmt.Printf("   Saved: %s\n", test.filename)
		}

		fmt.Println()
	}
}

// demonstrateGammaTable shows the gamma lookup table for different curves.
func demonstrateGammaTable() {
	fmt.Println("=== Gamma Lookup Table Analysis ===")
	fmt.Println()

	gammaCtrl := gamma.NewGammaCtrl(10, 10, 300, 200, false)

	curves := []struct {
		name               string
		kx1, ky1, kx2, ky2 float64
	}{
		{"Identity", 1.0, 1.0, 1.0, 1.0},
		{"Bright", 0.5, 1.5, 0.5, 1.5},
		{"Dark", 1.5, 0.5, 1.5, 0.5},
	}

	for _, curve := range curves {
		fmt.Printf("%s Curve (%.1f, %.1f, %.1f, %.1f):\n",
			curve.name, curve.kx1, curve.ky1, curve.kx2, curve.ky2)

		gammaCtrl.Values(curve.kx1, curve.ky1, curve.kx2, curve.ky2)
		gammaTable := gammaCtrl.Gamma()

		// Show sample values from the gamma table
		fmt.Print("Gamma table samples: ")
		for i := 0; i < 256; i += 32 {
			fmt.Printf("%d->%d ", i, gammaTable[i])
		}
		fmt.Printf("%d->%d\n", 255, gammaTable[255])

		// Calculate some statistics
		var sum int
		min, max := gammaTable[0], gammaTable[0]
		for _, val := range gammaTable {
			sum += int(val)
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
		avg := float64(sum) / 256.0

		fmt.Printf("Statistics: min=%d, max=%d, avg=%.1f\n", min, max, avg)
		fmt.Println()
	}
}

// demonstrateInteractivity simulates interactive control usage.
func demonstrateInteractivity() {
	fmt.Println("=== Interactive Control Simulation ===")
	fmt.Println()

	gammaCtrl := gamma.NewGammaCtrl(0, 0, 300, 200, false)

	// Simulate mouse interaction
	fmt.Println("Simulating mouse interactions:")

	// Set initial gamma curve
	gammaCtrl.Values(1.0, 1.0, 1.0, 1.0)
	kx1, ky1, kx2, ky2 := gammaCtrl.GetValues()
	fmt.Printf("Initial curve: %.3f, %.3f, %.3f, %.3f\n", kx1, ky1, kx2, ky2)

	// Simulate clicking on a control point and dragging
	fmt.Println("Simulating control point interaction...")

	// Simulate mouse interaction with control
	impl := gammaCtrl.GammaCtrlImpl

	// Simulate mouse down on a point (using approximate coordinates)
	handled := impl.OnMouseButtonDown(50, 50)
	fmt.Printf("Mouse down: handled=%t\n", handled)

	// Simulate dragging
	for i := 0; i < 5; i++ {
		dx := float64(i) * 2.0
		dy := float64(i) * 1.0
		handled = impl.OnMouseMove(50+dx, 50+dy, true)
		kx1, ky1, kx2, ky2 := impl.GetValues()
		fmt.Printf("Drag step %d: handled=%t, values=(%.3f, %.3f, %.3f, %.3f)\n",
			i+1, handled, kx1, ky1, kx2, ky2)
	}

	// Simulate mouse up
	handled = impl.OnMouseButtonUp(0, 0)
	fmt.Printf("Mouse up: handled=%t\n", handled)

	fmt.Println()

	// Simulate keyboard interaction
	fmt.Println("Simulating keyboard interactions:")

	// Test arrow keys
	keys := []struct {
		name                  string
		left, right, down, up bool
	}{
		{"Left arrow", true, false, false, false},
		{"Right arrow", false, true, false, false},
		{"Down arrow", false, false, true, false},
		{"Up arrow", false, false, false, true},
		{"No keys", false, false, false, false},
	}

	for _, key := range keys {
		oldKx1, oldKy1, oldKx2, oldKy2 := impl.GetValues()
		handled = impl.OnArrowKeys(key.left, key.right, key.down, key.up)
		newKx1, newKy1, newKx2, newKy2 := impl.GetValues()

		fmt.Printf("%s: handled=%t", key.name, handled)
		if handled {
			fmt.Printf(", change=(%.3f, %.3f, %.3f, %.3f) -> (%.3f, %.3f, %.3f, %.3f)",
				oldKx1, oldKy1, oldKx2, oldKy2, newKx1, newKy1, newKx2, newKy2)
		}
		fmt.Println()
	}

	// Test switching active point
	fmt.Println()
	fmt.Println("Testing active point switching...")
	impl.ChangeActivePoint()
	fmt.Println("Active point toggled")
}

// demonstrateVertexGeneration shows how to use the control as a vertex source.
func demonstrateVertexGeneration() {
	fmt.Println("=== Vertex Source Interface Demo ===")
	fmt.Println()

	gammaCtrl := gamma.NewGammaCtrl(50, 50, 250, 150, false)
	gammaCtrl.Values(0.8, 1.2, 1.2, 0.8) // Set interesting curve

	fmt.Printf("Control has %d rendering paths\n", gammaCtrl.NumPaths())
	fmt.Println()

	pathNames := []string{
		"Background",
		"Border",
		"Curve",
		"Grid",
		"Inactive Point",
		"Active Point",
		"Text",
	}

	// Generate vertices for each path
	for pathID := uint(0); pathID < gammaCtrl.NumPaths(); pathID++ {
		pathName := "Unknown"
		if int(pathID) < len(pathNames) {
			pathName = pathNames[pathID]
		}

		fmt.Printf("Path %d (%s):\n", pathID, pathName)

		// Get color for this path
		color := gammaCtrl.Color(pathID)
		fmt.Printf("  Color: %+v\n", color)

		gammaCtrl.Rewind(pathID)

		vertexCount := 0
		moveToCount := 0
		lineToCount := 0

		for {
			x, y, cmd := gammaCtrl.Vertex()
			if cmd == 0 { // PathCmdStop
				break
			}

			vertexCount++
			if vertexCount <= 5 { // Show first few vertices
				cmdName := "Unknown"
				switch cmd {
				case 1: // PathCmdMoveTo
					cmdName = "MoveTo"
					moveToCount++
				case 2: // PathCmdLineTo
					cmdName = "LineTo"
					lineToCount++
				case 3: // PathCmdCurve3
					cmdName = "Curve3"
				case 4: // PathCmdCurve4
					cmdName = "Curve4"
				}
				fmt.Printf("  Vertex %d: %s(%.2f, %.2f)\n", vertexCount, cmdName, x, y)
			}

			// Safety check to prevent infinite loops
			if vertexCount > 10000 {
				fmt.Printf("  ... (truncated - too many vertices)\n")
				break
			}
		}

		fmt.Printf("  Total: %d vertices (%d MoveTo, %d LineTo)\n",
			vertexCount, moveToCount, lineToCount)
		fmt.Println()
	}
}

func main() {
	fmt.Println("AGG Go - Gamma Correction Control Example")
	fmt.Println("=========================================")
	fmt.Println()

	// Run demonstrations
	demonstrateGammaControl()
	demonstrateGammaTable()
	demonstrateInteractivity()
	demonstrateVertexGeneration()

	fmt.Println("Demo completed! Check the generated PNG files to see gamma correction effects.")
	fmt.Println()
	fmt.Println("Files generated:")
	fmt.Println("  - original.png: Original test image")
	fmt.Println("  - gamma_*.png: Images with various gamma corrections applied")
	fmt.Println()
	fmt.Println("Note: This is a command-line demo. In a real application, you would:")
	fmt.Println("  1. Integrate the gamma control with a GUI framework")
	fmt.Println("  2. Render the control paths using AGG's rasterizer")
	fmt.Println("  3. Handle mouse and keyboard events from the GUI")
	fmt.Println("  4. Apply gamma correction to rendered graphics in real-time")
}
