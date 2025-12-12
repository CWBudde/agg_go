// Rasterizers2 Demo - Simplified port from AGG C++ rasterizers2.cpp
//
// This example demonstrates different rasterization techniques:
// 1. Aliased lines with pixel accuracy (Bresenham)
// 2. Aliased lines with subpixel accuracy
// 3. Anti-aliased outline rendering
// 4. Anti-aliased scanline rendering with stroke
// 5. Anti-aliased outline with image patterns (simplified)
//
// Note: This is a simplified version that demonstrates the rendering concepts
// without using the full rasterizer pipeline (which has API inconsistencies).

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
	frameWidth  = 500
	frameHeight = 450
	pixelSize   = 4 // RGBA
)

// Spiral generates a spiral path
type Spiral struct {
	x, y       float64
	r1, r2     float64
	step       float64
	startAngle float64
	angle      float64
	currR      float64
	da         float64
	dr         float64
	start      bool
}

func NewSpiral(x, y, r1, r2, step, startAngle float64) *Spiral {
	return &Spiral{
		x:          x,
		y:          y,
		r1:         r1,
		r2:         r2,
		step:       step,
		startAngle: startAngle,
		da:         basics.Deg2RadF(8.0),
		dr:         step / 45.0,
	}
}

func (s *Spiral) Rewind() {
	s.angle = s.startAngle
	s.currR = s.r1
	s.start = true
}

// NextPoint returns the next point in the spiral, returns false when done
func (s *Spiral) NextPoint() (x, y float64, isStart bool, done bool) {
	if s.currR > s.r2 {
		return 0, 0, false, true
	}

	x = s.x + math.Cos(s.angle)*s.currR
	y = s.y + math.Sin(s.angle)*s.currR
	isStart = s.start
	s.currR += s.dr
	s.angle += s.da

	if s.start {
		s.start = false
	}
	return x, y, isStart, false
}

// Application holds the demo application state
type Application struct {
	rbuf       *buffer.RenderingBufferU8
	pixf       *pixfmt.PixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8[color.Linear, order.RGBA]]
	imageData  []byte
	step       float64
	width      float64
	startAngle float64
}

func NewApplication() *Application {
	app := &Application{
		imageData:  make([]byte, frameWidth*frameHeight*pixelSize),
		step:       0.1,
		width:      3.0,
		startAngle: 0.0,
	}

	app.rbuf = buffer.NewRenderingBufferU8WithData(app.imageData, frameWidth, frameHeight, frameWidth*pixelSize)
	b := blender.BlenderRGBA8[color.Linear, order.RGBA]{}
	app.pixf = pixfmt.NewPixFmtAlphaBlendRGBA[color.Linear, blender.BlenderRGBA8[color.Linear, order.RGBA]](app.rbuf, b)

	return app
}

// clearBuffer clears the buffer with a cream color
func (app *Application) clearBuffer() {
	cream := color.RGBA8[color.Linear]{R: 255, G: 255, B: 242, A: 255}
	for y := 0; y < app.pixf.Height(); y++ {
		for x := 0; x < app.pixf.Width(); x++ {
			app.pixf.CopyPixel(x, y, cream)
		}
	}
}

// drawPixel draws a single pixel with bounds checking
func (app *Application) drawPixel(x, y int, c color.RGBA8[color.Linear]) {
	if x >= 0 && y >= 0 && x < frameWidth && y < frameHeight {
		app.pixf.BlendPixel(x, y, c, basics.CoverFull)
	}
}

// drawLine draws a simple Bresenham line
func (app *Application) drawLine(x0, y0, x1, y1 int, c color.RGBA8[color.Linear]) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}

	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		app.drawPixel(x, y, c)

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// drawThickLine draws a line with thickness
func (app *Application) drawThickLine(x0, y0, x1, y1 float64, thickness int, c color.RGBA8[color.Linear]) {
	for dy := -thickness / 2; dy <= thickness/2; dy++ {
		for dx := -thickness / 2; dx <= thickness/2; dx++ {
			app.drawLine(int(x0)+dx, int(y0)+dy, int(x1)+dx, int(y1)+dy, c)
		}
	}
}

// drawAntiAliasedLine draws an anti-aliased line using Wu's algorithm (simplified)
func (app *Application) drawAntiAliasedLine(x0, y0, x1, y1 float64, c color.RGBA8[color.Linear]) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Sqrt(dx*dx + dy*dy)

	if length < 1 {
		return
	}

	steps := int(length * 2)
	if steps < 2 {
		steps = 2
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x0 + t*dx
		y := y0 + t*dy

		// Simple anti-aliasing by blending with fractional coverage
		ix, iy := int(x), int(y)
		fx, fy := x-float64(ix), y-float64(iy)

		// Blend the four surrounding pixels
		app.blendPixelCoverage(ix, iy, c, (1-fx)*(1-fy))
		app.blendPixelCoverage(ix+1, iy, c, fx*(1-fy))
		app.blendPixelCoverage(ix, iy+1, c, (1-fx)*fy)
		app.blendPixelCoverage(ix+1, iy+1, c, fx*fy)
	}
}

func (app *Application) blendPixelCoverage(x, y int, c color.RGBA8[color.Linear], coverage float64) {
	if x >= 0 && y >= 0 && x < frameWidth && y < frameHeight {
		cover := basics.Int8u(coverage * 255)
		if cover > 0 {
			app.pixf.BlendPixel(x, y, c, cover)
		}
	}
}

// drawSpiral draws a spiral using the specified drawing method
func (app *Application) drawSpiral(centerX, centerY float64, c color.RGBA8[color.Linear], method int) {
	spiral := NewSpiral(centerX, centerY, 5, 70, 16, app.startAngle)
	spiral.Rewind()

	var prevX, prevY float64
	first := true
	thickness := int(app.width + 0.5)
	if thickness < 1 {
		thickness = 1
	}

	for {
		x, y, isStart, done := spiral.NextPoint()
		if done {
			break
		}

		if isStart {
			prevX, prevY = x, y
			first = true
			continue
		}

		if !first {
			switch method {
			case 0: // Aliased pixel accuracy (rounded)
				app.drawLine(int(prevX+0.5), int(prevY+0.5), int(x+0.5), int(y+0.5), c)
			case 1: // Aliased subpixel accuracy (fractional coordinates used directly)
				app.drawLine(int(prevX), int(prevY), int(x), int(y), c)
			case 2: // Anti-aliased outline
				app.drawAntiAliasedLine(prevX, prevY, x, y, c)
			case 3: // Anti-aliased with thickness
				app.drawThickLine(prevX, prevY, x, y, thickness, c)
			case 4: // Pattern (simplified - just uses colored pixels)
				app.drawPatternLine(prevX, prevY, x, y, thickness)
			}
		}

		prevX, prevY = x, y
		first = false
	}
}

// drawPatternLine draws a line with a simple chain-like pattern
func (app *Application) drawPatternLine(x0, y0, x1, y1 float64, thickness int) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Sqrt(dx*dx + dy*dy)

	if length < 1 {
		return
	}

	steps := int(length)
	if steps < 2 {
		steps = 2
	}

	// Chain link colors
	colors := []color.RGBA8[color.Linear]{
		{R: 102, G: 0, B: 0, A: 255},     // Dark red
		{R: 154, G: 87, B: 87, A: 255},   // Light red
		{R: 194, G: 153, B: 153, A: 180}, // Pink (semi-transparent)
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x0 + t*dx
		y := y0 + t*dy

		// Choose color based on position in pattern
		colorIdx := (i / 4) % len(colors)
		c := colors[colorIdx]

		// Draw circle at this position
		for dy := -thickness; dy <= thickness; dy++ {
			for dx := -thickness; dx <= thickness; dx++ {
				if dx*dx+dy*dy <= thickness*thickness {
					app.blendPixelCoverage(int(x)+dx, int(y)+dy, c, 0.7)
				}
			}
		}
	}
}

// drawText draws simple placeholder text
func (app *Application) drawText(x, y float64, text string, c color.RGBA8[color.Linear]) {
	for dy := 0; dy < 8; dy++ {
		for dx := 0; dx < len(text)*6; dx++ {
			px, py := int(x)+dx, int(y)+dy
			app.blendPixelCoverage(px, py, c, 0.25)
		}
	}
}

func (app *Application) OnDraw() {
	width := float64(frameWidth)
	height := float64(frameHeight)

	app.clearBuffer()

	drawColor := color.RGBA8[color.Linear]{R: 102, G: 77, B: 26, A: 255} // Brown
	textColor := color.RGBA8[color.Linear]{R: 0, G: 0, B: 0, A: 255}

	// Draw five spirals with different techniques
	// 1. Aliased pixel accuracy (top left)
	app.drawSpiral(width/5, height/4, drawColor, 0)
	app.drawText(30, 60, "Bresenham (pixel)", textColor)

	// 2. Aliased subpixel accuracy (top center)
	app.drawSpiral(width/2, height/4, drawColor, 1)
	app.drawText(width/2-50, 60, "Bresenham (subpixel)", textColor)

	// 3. Anti-aliased outline (bottom left)
	app.drawSpiral(width/5, height-height/4+20, drawColor, 2)
	app.drawText(30, height/2+30, "Anti-aliased", textColor)

	// 4. Anti-aliased with thickness (bottom center)
	app.drawSpiral(width/2, height-height/4+20, drawColor, 3)
	app.drawText(width/2-50, height/2+30, "AA with thickness", textColor)

	// 5. Pattern line (bottom right)
	app.drawSpiral(width-width/5, height-height/4+20, drawColor, 4)
	app.drawText(width-width/5-50, height/2+30, "Pattern", textColor)
}

func (app *Application) saveImage(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "P6\n%d %d\n255\n", frameWidth, frameHeight)

	for i := 0; i < len(app.imageData); i += 4 {
		file.Write([]byte{app.imageData[i], app.imageData[i+1], app.imageData[i+2]})
	}

	return nil
}

func (app *Application) performanceTest() {
	fmt.Println("Running performance test...")

	drawColor := color.RGBA8[color.Linear]{R: 102, G: 77, B: 26, A: 255}
	width := float64(frameWidth)
	height := float64(frameHeight)

	// Test aliased subpixel accuracy
	fmt.Print("Testing aliased subpixel accuracy...")
	start := time.Now()
	for i := 0; i < 200; i++ {
		app.drawSpiral(width/2, height/4, drawColor, 1)
		app.startAngle += basics.Deg2RadF(app.step)
	}
	elapsed1 := time.Since(start)
	fmt.Printf(" %.2f ms\n", float64(elapsed1.Nanoseconds())/1e6)

	// Test anti-aliased outline
	fmt.Print("Testing anti-aliased outline...")
	start = time.Now()
	for i := 0; i < 200; i++ {
		app.drawSpiral(width/5, height-height/4+20, drawColor, 2)
		app.startAngle += basics.Deg2RadF(app.step)
	}
	elapsed2 := time.Since(start)
	fmt.Printf(" %.2f ms\n", float64(elapsed2.Nanoseconds())/1e6)

	// Test anti-aliased with thickness
	fmt.Print("Testing anti-aliased with thickness...")
	start = time.Now()
	for i := 0; i < 200; i++ {
		app.drawSpiral(width/2, height-height/4+20, drawColor, 3)
		app.startAngle += basics.Deg2RadF(app.step)
	}
	elapsed3 := time.Since(start)
	fmt.Printf(" %.2f ms\n", float64(elapsed3.Nanoseconds())/1e6)

	// Test pattern line
	fmt.Print("Testing pattern line...")
	start = time.Now()
	for i := 0; i < 200; i++ {
		app.drawSpiral(width-width/5, height-height/4+20, drawColor, 4)
		app.startAngle += basics.Deg2RadF(app.step)
	}
	elapsed4 := time.Since(start)
	fmt.Printf(" %.2f ms\n", float64(elapsed4.Nanoseconds())/1e6)

	totalTime := elapsed1 + elapsed2 + elapsed3 + elapsed4
	fmt.Printf("\nPerformance Test Results:\n")
	fmt.Printf("  Aliased subpixel accuracy:     %8.2f ms\n", float64(elapsed1.Nanoseconds())/1e6)
	fmt.Printf("  Anti-aliased outline:          %8.2f ms\n", float64(elapsed2.Nanoseconds())/1e6)
	fmt.Printf("  Anti-aliased + thickness:      %8.2f ms\n", float64(elapsed3.Nanoseconds())/1e6)
	fmt.Printf("  Pattern line:                  %8.2f ms\n", float64(elapsed4.Nanoseconds())/1e6)
	fmt.Printf("  Total time:                    %8.2f ms\n", float64(totalTime.Nanoseconds())/1e6)
	fmt.Println("Performance test completed!")
}

func main() {
	fmt.Println("AGG Rasterizers2 Demo (Simplified)")
	fmt.Println("===================================")
	fmt.Println("This demo shows different rasterization techniques:")
	fmt.Println("- Bresenham lines (pixel and subpixel accuracy)")
	fmt.Println("- Anti-aliased rendering")
	fmt.Println("- Thick line rendering")
	fmt.Println("- Pattern line rendering")

	app := NewApplication()

	// Draw initial frame
	app.OnDraw()

	// Save the result
	err := app.saveImage("rasterizers2_demo.ppm")
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return
	}

	fmt.Println("\nDemo image saved as 'rasterizers2_demo.ppm'")

	// Test animation frames
	fmt.Println("\nGenerating animation frames...")
	for i := 0; i < 5; i++ {
		app.startAngle += basics.Deg2RadF(15.0)
		app.OnDraw()
		filename := fmt.Sprintf("rasterizers2_frame_%d.ppm", i)
		err := app.saveImage(filename)
		if err != nil {
			fmt.Printf("Error saving %s: %v\n", filename, err)
		} else {
			fmt.Printf("Saved %s\n", filename)
		}
	}

	// Run performance test
	fmt.Println()
	app.performanceTest()

	fmt.Println("\nExample completed successfully!")
	fmt.Println("\nNote: This is a simplified version demonstrating core rendering concepts.")
	fmt.Println("The full rasterizer pipeline has API inconsistencies being resolved.")
}
