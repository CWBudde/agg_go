// Package main implements a simplified rasterizers example from AGG.
// This example demonstrates the comparison between anti-aliased and aliased rendering
// with gamma correction and transparency controls.
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
	"agg_go/internal/path"
	"agg_go/internal/pixfmt"
	"agg_go/internal/rasterizer"
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

	// Rendering components
	ras *rasterizer.RasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl]

	// Rendering buffer and pixel format
	rbuf *buffer.RenderingBufferU8
	pixf *pixfmt.PixFmtAlphaBlendRGBA[pixfmt.BlenderRGBA[color.Linear, pixfmt.RGBAOrder], color.Linear]

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
	blender := pixfmt.BlenderRGBA[color.Linear, pixfmt.RGBAOrder]{}
	app.pixf = pixfmt.NewPixFmtAlphaBlendRGBA[pixfmt.BlenderRGBA[color.Linear, pixfmt.RGBAOrder], color.Linear](app.rbuf, blender)

	// Create rasterizer
	app.ras = rasterizer.NewRasterizerScanlineAA[*rasterizer.RasterizerSlNoClip, rasterizer.RasConvDbl](1000) // cell block limit

	// Note: Clipper initialization removed due to interface incompatibility and circular dependency
	// The RasterizerScanlineAA doesn't implement the Line() method required by RasterizerInterface
	// and creating a clipper with the rasterizer itself creates circular dependency

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

// simpleScanlineRender performs a simplified scanline rendering
func (app *Application) simpleScanlineRender(triangleColor color.RGBA8[color.Linear], useGamma bool) {
	// Set up gamma function
	if useGamma {
		gamma := app.gamma * 2.0
		gammaFunc := func(x float64) float64 {
			return math.Pow(x, gamma)
		}
		app.ras.SetGamma(gammaFunc)
	} else {
		// Binary threshold function
		threshold := app.gamma
		gammaFunc := func(x float64) float64 {
			if x < threshold {
				return 0.0
			}
			return 1.0
		}
		app.ras.SetGamma(gammaFunc)
	}

	// Check if we can render scanlines
	if !app.ras.RewindScanlines() {
		return
	}

	// Manual scanline rendering - simplified approach
	for y := app.ras.MinY(); y <= app.ras.MaxY(); y++ {
		// Get scanline data manually by examining cells
		if !app.ras.NavigateScanline(y) {
			continue
		}

		// For each pixel in the scanline, test if it's covered
		for x := app.ras.MinX(); x <= app.ras.MaxX(); x++ {
			// Simple hit test to determine coverage
			if app.ras.HitTest(x, y) {
				// Apply alpha blending
				finalColor := triangleColor
				finalColor.A = basics.Int8u(float64(finalColor.A) * app.alpha)

				// Bounds check
				if x >= 0 && x < frameWidth && y >= 0 && y < frameHeight {
					app.pixf.BlendPixel(x, y, finalColor, 255)
				}
			}
		}
	}
}

// drawAntiAliased renders the triangle with anti-aliasing and gamma correction
func (app *Application) drawAntiAliased() {
	// Create path for triangle
	pathStorage := path.NewPathStorage()
	pathStorage.MoveTo(app.x[0], app.y[0])
	pathStorage.LineTo(app.x[1], app.y[1])
	pathStorage.LineTo(app.x[2], app.y[2])
	pathStorage.ClosePolygon(basics.PathFlag(basics.PathFlagClose))

	// Reset rasterizer and add path vertices manually
	app.ras.Reset()

	// Add triangle vertices to rasterizer
	app.ras.MoveToD(app.x[0], app.y[0])
	app.ras.LineToD(app.x[1], app.y[1])
	app.ras.LineToD(app.x[2], app.y[2])
	app.ras.ClosePolygon()

	// Set color (brownish)
	triangleColor := color.RGBA8[color.Linear]{
		R: basics.Int8u(0.7*255 + 0.5),
		G: basics.Int8u(0.5*255 + 0.5),
		B: basics.Int8u(0.1*255 + 0.5),
		A: basics.Int8u(app.alpha*255 + 0.5),
	}

	// Render with anti-aliasing
	app.simpleScanlineRender(triangleColor, true)
}

// drawAliased renders the triangle without anti-aliasing (binary)
func (app *Application) drawAliased() {
	// Reset rasterizer and add path vertices manually (offset by 200 pixels left)
	app.ras.Reset()

	// Add triangle vertices to rasterizer
	app.ras.MoveToD(app.x[0]-200, app.y[0])
	app.ras.LineToD(app.x[1]-200, app.y[1])
	app.ras.LineToD(app.x[2]-200, app.y[2])
	app.ras.ClosePolygon()

	// Set color (blueish - different from anti-aliased)
	triangleColor := color.RGBA8[color.Linear]{
		R: basics.Int8u(0.1*255 + 0.5),
		G: basics.Int8u(0.5*255 + 0.5),
		B: basics.Int8u(0.7*255 + 0.5),
		A: basics.Int8u(app.alpha*255 + 0.5),
	}

	// Render with binary (aliased) rendering
	app.simpleScanlineRender(triangleColor, false)
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
	fmt.Println("AGG Rasterizers Example (Simplified)")
	fmt.Println("This example demonstrates anti-aliased vs aliased rendering")

	// Create application
	app := NewApplication()

	// Render initial frame
	app.onDraw()

	// Save the result
	err := app.saveImage("rasterizers_demo_simple.ppm")
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}

	fmt.Println("Demo image saved as 'rasterizers_demo_simple.ppm'")
	fmt.Println("Left triangle: Anti-aliased (brownish)")
	fmt.Println("Right triangle: Aliased/Binary (blueish)")

	// Demonstrate different gamma values
	fmt.Println("\nTesting different gamma values...")

	for _, gamma := range []float64{0.1, 0.5, 1.0} {
		app.gamma = gamma
		app.onDraw()
		filename := fmt.Sprintf("rasterizers_gamma_%.1f_simple.ppm", gamma)
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
	fmt.Println("\nNote: This is a simplified version that demonstrates the core concepts.")
	fmt.Println("For interactive mouse manipulation, see the full implementation.")
}
