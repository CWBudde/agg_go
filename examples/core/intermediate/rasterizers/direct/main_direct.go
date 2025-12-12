// Package main implements a direct rasterizers example from AGG.
// This example demonstrates the comparison between anti-aliased and aliased rendering
// by implementing simple triangle rasterization directly.
// Simplified port of ../agg-2.6/agg-src/examples/rasterizers.cpp
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"agg_go/internal/basics"
	"agg_go/internal/buffer"
	"agg_go/internal/color"
	"agg_go/internal/order"
	"agg_go/internal/pixfmt"
	"agg_go/internal/pixfmt/blender"
)

const (
	// Image dimensions
	frameWidth  = 500
	frameHeight = 330

	// Pixel format - we'll use RGBA for simplicity
	pixelSize = 4 // RGBA
)

// Application holds the main application state
type Application struct {
	// Triangle vertices (3 points)
	x, y [3]float64

	// Control values
	gamma float64
	alpha float64

	// Rendering buffer and pixel format
	rbuf *buffer.RenderingBufferU8
	pixf *pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8[color.Linear, order.RGBA]]

	// Image buffer
	imageData []byte
}

// NewApplication creates a new rasterizers application
func NewApplication() *Application {
	app := &Application{
		gamma:     0.5,
		alpha:     1.0,
		imageData: make([]byte, frameWidth*frameHeight*pixelSize),
	}

	// Initialize triangle vertices (matching C++ original positions)
	app.x[0] = 100 + 120
	app.y[0] = 60
	app.x[1] = 369 + 120
	app.y[1] = 170
	app.x[2] = 143 + 120
	app.y[2] = 310

	// Initialize rendering components
	app.rbuf = buffer.NewRenderingBufferU8WithData(app.imageData, frameWidth, frameHeight, frameWidth*pixelSize)

	// Create blender and pixel format
	b := blender.BlenderRGBA8[color.Linear, order.RGBA]{}
	app.pixf = pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8[color.Linear, order.RGBA]](app.rbuf, b)

	return app
}

// clearBuffer clears the entire buffer to white
func (app *Application) clearBuffer() {
	whiteColor := color.RGBA8[color.Linear]{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < app.pixf.Height(); y++ {
		for x := 0; x < app.pixf.Width(); x++ {
			app.pixf.CopyPixel(x, y, whiteColor)
		}
	}
}

// isPointInTriangle checks if a point is inside a triangle using barycentric coordinates
func (app *Application) isPointInTriangle(px, py float64, offsetX float64) bool {
	// Adjust triangle coordinates with offset
	x1, y1 := app.x[0]+offsetX, app.y[0]
	x2, y2 := app.x[1]+offsetX, app.y[1]
	x3, y3 := app.x[2]+offsetX, app.y[2]

	// Barycentric coordinate calculation
	denom := (y2-y3)*(x1-x3) + (x3-x2)*(y1-y3)
	if math.Abs(denom) < 1e-10 {
		return false
	}

	a := ((y2-y3)*(px-x3) + (x3-x2)*(py-y3)) / denom
	b := ((y3-y1)*(px-x3) + (x1-x3)*(py-y3)) / denom
	c := 1 - a - b

	return a >= 0 && b >= 0 && c >= 0
}

// getEdgeDistance calculates the distance from a point to the nearest triangle edge
func (app *Application) getEdgeDistance(px, py float64, offsetX float64) float64 {
	// Adjust triangle coordinates with offset
	x1, y1 := app.x[0]+offsetX, app.y[0]
	x2, y2 := app.x[1]+offsetX, app.y[1]
	x3, y3 := app.x[2]+offsetX, app.y[2]

	// Calculate distance to each edge and return minimum
	edges := [][4]float64{
		{x1, y1, x2, y2},
		{x2, y2, x3, y3},
		{x3, y3, x1, y1},
	}

	minDist := math.Inf(1)
	for _, edge := range edges {
		ex1, ey1, ex2, ey2 := edge[0], edge[1], edge[2], edge[3]

		// Distance from point to line segment
		A := px - ex1
		B := py - ey1
		C := ex2 - ex1
		D := ey2 - ey1

		dot := A*C + B*D
		lenSq := C*C + D*D

		var param float64 = -1
		if lenSq != 0 {
			param = dot / lenSq
		}

		var xx, yy float64
		if param < 0 {
			xx, yy = ex1, ey1
		} else if param > 1 {
			xx, yy = ex2, ey2
		} else {
			xx, yy = ex1+param*C, ey1+param*D
		}

		dx := px - xx
		dy := py - yy
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

// calculateCoverage calculates anti-aliased coverage for a pixel
func (app *Application) calculateCoverage(px, py float64, offsetX float64, useAntiAliasing bool) float64 {
	if !useAntiAliasing {
		// Binary coverage - either fully inside or outside
		if app.isPointInTriangle(px+0.5, py+0.5, offsetX) {
			return 1.0
		}
		return 0.0
	}

	// Anti-aliased coverage - sample multiple points within the pixel
	samples := 4 // 4x4 supersampling
	sampleSize := 1.0 / float64(samples)
	totalCoverage := 0.0

	for sy := 0; sy < samples; sy++ {
		for sx := 0; sx < samples; sx++ {
			sampleX := px + (float64(sx)+0.5)*sampleSize
			sampleY := py + (float64(sy)+0.5)*sampleSize

			if app.isPointInTriangle(sampleX, sampleY, offsetX) {
				totalCoverage += 1.0
			}
		}
	}

	return totalCoverage / float64(samples*samples)
}

// applyGamma applies gamma correction to coverage value
func (app *Application) applyGamma(coverage float64, useAntiAliasing bool) float64 {
	if !useAntiAliasing {
		// Binary threshold for aliased rendering
		if coverage >= app.gamma {
			return 1.0
		}
		return 0.0
	}

	// Power gamma for anti-aliased rendering
	gamma := app.gamma * 2.0
	return math.Pow(coverage, gamma)
}

// renderTriangle renders a triangle with the specified parameters
func (app *Application) renderTriangle(triangleColor color.RGBA8[color.Linear], offsetX float64, useAntiAliasing bool) {
	// Calculate bounding box
	minX := int(math.Floor(math.Min(math.Min(app.x[0], app.x[1]), app.x[2]) + offsetX))
	maxX := int(math.Ceil(math.Max(math.Max(app.x[0], app.x[1]), app.x[2]) + offsetX))
	minY := int(math.Floor(math.Min(math.Min(app.y[0], app.y[1]), app.y[2])))
	maxY := int(math.Ceil(math.Max(math.Max(app.y[0], app.y[1]), app.y[2])))

	// Clamp to image bounds
	if minX < 0 {
		minX = 0
	}
	if maxX >= frameWidth {
		maxX = frameWidth - 1
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= frameHeight {
		maxY = frameHeight - 1
	}

	// Rasterize each pixel in the bounding box
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			coverage := app.calculateCoverage(float64(x), float64(y), offsetX, useAntiAliasing)
			if coverage > 0 {
				// Apply gamma correction
				correctedCoverage := app.applyGamma(coverage, useAntiAliasing)

				if correctedCoverage > 0 {
					// Apply alpha and coverage
					finalColor := triangleColor
					finalColor.A = basics.Int8u(float64(finalColor.A)*app.alpha*correctedCoverage + 0.5)

					if finalColor.A > 0 {
						// Blend with existing pixel
						cover := basics.Int8u(correctedCoverage*255 + 0.5)
						app.pixf.BlendPixel(x, y, finalColor, cover)
					}
				}
			}
		}
	}
}

// drawAntiAliased renders the triangle with anti-aliasing and gamma correction
func (app *Application) drawAntiAliased() {
	// Set color (brownish)
	triangleColor := color.RGBA8[color.Linear]{
		R: basics.Int8u(0.7*255 + 0.5),
		G: basics.Int8u(0.5*255 + 0.5),
		B: basics.Int8u(0.1*255 + 0.5),
		A: 255,
	}

	// Render with anti-aliasing (no offset)
	app.renderTriangle(triangleColor, 0, true)
}

// drawAliased renders the triangle without anti-aliasing (binary)
func (app *Application) drawAliased() {
	// Set color (blueish - different from anti-aliased)
	triangleColor := color.RGBA8[color.Linear]{
		R: basics.Int8u(0.1*255 + 0.5),
		G: basics.Int8u(0.5*255 + 0.5),
		B: basics.Int8u(0.7*255 + 0.5),
		A: 255,
	}

	// Render without anti-aliasing (offset by -200 pixels)
	app.renderTriangle(triangleColor, -200, false)
}

// onDraw renders the complete frame
func (app *Application) onDraw() {
	// Clear background to white
	app.clearBuffer()

	// Draw both triangles
	app.drawAntiAliased()
	app.drawAliased()
}

// saveImage saves the current frame as a PPM file
func (app *Application) saveImage(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write PPM header
	fmt.Fprintf(file, "P6\n%d %d\n255\n", frameWidth, frameHeight)

	// Write pixel data (convert RGBA to RGB)
	for i := 0; i < len(app.imageData); i += 4 {
		file.Write([]byte{app.imageData[i], app.imageData[i+1], app.imageData[i+2]})
	}

	return nil
}

func main() {
	fmt.Println("AGG Rasterizers Example (Direct Implementation)")
	fmt.Println("This example demonstrates anti-aliased vs aliased rendering")

	// Create application
	app := NewApplication()

	// Render initial frame
	app.onDraw()

	// Save the result
	err := app.saveImage("rasterizers_demo_direct.ppm")
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}

	fmt.Println("Demo image saved as 'rasterizers_demo_direct.ppm'")
	fmt.Println("Left triangle: Anti-aliased (brownish)")
	fmt.Println("Right triangle: Aliased/Binary (blueish)")

	// Demonstrate different gamma values
	fmt.Println("\nTesting different gamma values...")

	for _, gamma := range []float64{0.1, 0.5, 1.0} {
		app.gamma = gamma
		app.onDraw()
		filename := fmt.Sprintf("rasterizers_gamma_%.1f_direct.ppm", gamma)
		err := app.saveImage(filename)
		if err != nil {
			fmt.Printf("Error saving %s: %v\n", filename, err)
		} else {
			fmt.Printf("Saved %s (gamma=%.1f)\n", filename, gamma)
		}
	}

	// Performance test
	fmt.Println("\nRunning performance test...")

	// Test aliased rendering
	start := time.Now()
	for i := 0; i < 1000; i++ {
		app.drawAliased()
	}
	t1 := time.Since(start)

	// Test anti-aliased rendering
	start = time.Now()
	for i := 0; i < 1000; i++ {
		app.drawAntiAliased()
	}
	t2 := time.Since(start)

	fmt.Printf("Time Aliased=%.2fms Time Anti-Aliased=%.2fms\n",
		float64(t1.Nanoseconds())/1e6, float64(t2.Nanoseconds())/1e6)

	fmt.Println("\nExample completed successfully!")
	fmt.Println("\nNote: This implementation uses direct triangle rasterization")
	fmt.Println("to demonstrate the core concepts without complex AGG interfaces.")
}
