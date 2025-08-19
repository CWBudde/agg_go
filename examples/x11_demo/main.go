//go:build x11
// +build x11

// Package main demonstrates the X11 platform backend for AGG.
// This example shows how to create interactive graphics applications using the X11 windowing system.
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"agg_go/internal/platform"
)

// X11DemoApp represents our demo application
type X11DemoApp struct {
	platform.BaseEventHandler
	backend     platform.PlatformBackend
	ps          *platform.PlatformSupport
	rc          *platform.RenderingContext
	frameCount  int
	mouseX      int
	mouseY      int
	isDragging  bool
	shapes      []Shape
	currentTime time.Time
}

// Shape represents a drawable shape
type Shape struct {
	X, Y       float64
	VX, VY     float64
	R, G, B, A uint8
	Size       float64
	Type       string
}

// NewX11DemoApp creates a new demo application
func NewX11DemoApp() (*X11DemoApp, error) {
	app := &X11DemoApp{
		currentTime: time.Now(),
		shapes:      make([]Shape, 0, 100),
	}

	// Create backend factory and get X11 backend
	factory := platform.GetBackendFactory()
	backend, err := factory.CreateBackend(platform.BackendX11, platform.PixelFormatRGBA32, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create X11 backend: %w", err)
	}
	app.backend = backend

	// Create platform support
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG X11 Demo - Interactive Graphics")

	// Create rendering context
	app.rc = platform.NewRenderingContext(app.ps)

	// Initialize some demo shapes
	app.initShapes()

	return app, nil
}

// initShapes initializes some animated shapes for the demo
func (app *X11DemoApp) initShapes() {
	// Create some bouncing shapes
	for i := 0; i < 20; i++ {
		shape := Shape{
			X:    float64(100 + i*20),
			Y:    float64(100 + i*10),
			VX:   float64((i%5 - 2) * 50),
			VY:   float64((i%3 - 1) * 30),
			R:    uint8((i * 50) % 256),
			G:    uint8((i * 80) % 256),
			B:    uint8((i * 120) % 256),
			A:    255,
			Size: float64(10 + i%15),
			Type: []string{"circle", "rect", "line"}[i%3],
		}
		app.shapes = append(app.shapes, shape)
	}
}

// OnInit is called when the application is initialized
func (app *X11DemoApp) OnInit() {
	fmt.Println("X11 Demo initialized")
	fmt.Printf("Window size: %dx%d\n", app.ps.Width(), app.ps.Height())
	fmt.Println("Controls:")
	fmt.Println("  - Move mouse to interact with shapes")
	fmt.Println("  - Left click and drag to create trails")
	fmt.Println("  - Press 'c' to clear the screen")
	fmt.Println("  - Press 'r' to reset shapes")
	fmt.Println("  - Press 'ESC' to exit")
}

// OnDestroy is called when the application is being destroyed
func (app *X11DemoApp) OnDestroy() {
	fmt.Println("X11 Demo destroyed")
}

// OnResize is called when the window is resized
func (app *X11DemoApp) OnResize(width, height int) {
	fmt.Printf("Window resized to %dx%d\n", width, height)
	app.rc.SetupResizeTransform(width, height)
	app.backend.ForceRedraw()
}

// OnMouseMove is called when the mouse is moved
func (app *X11DemoApp) OnMouseMove(x, y int, flags platform.InputFlags) {
	app.mouseX = x
	app.mouseY = y

	if app.isDragging || flags.HasMouseLeft() {
		// Create trail effect
		app.rc.FillRectangle(x-2, y-2, 4, 4, 255, 255, 255, 128)
		app.backend.UpdateWindow(app.ps.WindowBuffer())
	}
}

// OnMouseButtonDown is called when a mouse button is pressed
func (app *X11DemoApp) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	app.isDragging = flags.HasMouseLeft()

	if flags.HasMouseLeft() {
		// Add a new shape at mouse position
		shape := Shape{
			X:    float64(x),
			Y:    float64(y),
			VX:   float64((x % 200) - 100),
			VY:   float64((y % 200) - 100),
			R:    uint8(app.frameCount % 256),
			G:    uint8((app.frameCount * 2) % 256),
			B:    uint8((app.frameCount * 3) % 256),
			A:    255,
			Size: 15,
			Type: "circle",
		}
		app.shapes = append(app.shapes, shape)
	}

	if flags.HasMouseRight() {
		// Remove nearby shapes
		app.removeNearbyShapes(float64(x), float64(y), 50)
	}
}

// OnMouseButtonUp is called when a mouse button is released
func (app *X11DemoApp) OnMouseButtonUp(x, y int, flags platform.InputFlags) {
	app.isDragging = false
}

// OnKey is called when a key is pressed
func (app *X11DemoApp) OnKey(x, y int, key platform.KeyCode, flags platform.InputFlags) {
	switch key {
	case platform.KeyEscape:
		fmt.Println("Escape pressed - exiting")
		// In a real implementation, you would set a quit flag
		app.backend.Destroy()

	case platform.KeyCode('c'), platform.KeyCode('C'):
		// Clear screen
		app.rc.ClearWindow(0, 0, 0, 255)
		app.backend.UpdateWindow(app.ps.WindowBuffer())
		fmt.Println("Screen cleared")

	case platform.KeyCode('r'), platform.KeyCode('R'):
		// Reset shapes
		app.shapes = app.shapes[:0]
		app.initShapes()
		fmt.Println("Shapes reset")

	case platform.KeyCode('f'), platform.KeyCode('F'):
		// Toggle fullscreen effect by maximizing shapes
		for i := range app.shapes {
			app.shapes[i].Size *= 1.5
		}

	case platform.KeyF1:
		app.showInfo()
	}
}

// OnDraw is called when the window needs to be redrawn
func (app *X11DemoApp) OnDraw() {
	app.frameCount++

	// Clear background with animated color
	bgIntensity := uint8(math.Sin(float64(app.frameCount)*0.01)*30 + 40)
	app.rc.ClearWindow(bgIntensity/4, bgIntensity/6, bgIntensity/2, 255)

	// Update and draw shapes
	app.updateShapes()
	app.drawShapes()

	// Draw UI elements
	app.drawUI()

	// Update window
	app.backend.UpdateWindow(app.ps.WindowBuffer())
}

// OnIdle is called when the application is idle
func (app *X11DemoApp) OnIdle() {
	// Trigger redraw for animation
	if app.frameCount%2 == 0 { // Limit frame rate
		app.backend.ForceRedraw()
	}
}

// updateShapes updates the positions and properties of animated shapes
func (app *X11DemoApp) updateShapes() {
	dt := time.Since(app.currentTime).Seconds()
	app.currentTime = time.Now()

	width := float64(app.ps.Width())
	height := float64(app.ps.Height())

	for i := range app.shapes {
		shape := &app.shapes[i]

		// Update position
		shape.X += shape.VX * dt
		shape.Y += shape.VY * dt

		// Bounce off walls
		if shape.X < 0 || shape.X > width {
			shape.VX = -shape.VX
			shape.X = math.Max(0, math.Min(width, shape.X))
		}
		if shape.Y < 0 || shape.Y > height {
			shape.VY = -shape.VY
			shape.Y = math.Max(0, math.Min(height, shape.Y))
		}

		// Add attraction to mouse cursor
		if app.isDragging {
			dx := float64(app.mouseX) - shape.X
			dy := float64(app.mouseY) - shape.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 && dist < 200 {
				force := 100.0 / (dist + 1)
				shape.VX += dx / dist * force * dt
				shape.VY += dy / dist * force * dt
			}
		}

		// Apply some damping
		shape.VX *= 0.99
		shape.VY *= 0.99

		// Update color cycling
		shape.R = uint8((int(shape.R) + 1) % 256)
	}

	// Remove shapes that are too far out of bounds
	newShapes := app.shapes[:0]
	for _, shape := range app.shapes {
		if shape.X > -100 && shape.X < width+100 && shape.Y > -100 && shape.Y < height+100 {
			newShapes = append(newShapes, shape)
		}
	}
	app.shapes = newShapes
}

// drawShapes renders all the animated shapes
func (app *X11DemoApp) drawShapes() {
	for _, shape := range app.shapes {
		switch shape.Type {
		case "circle":
			app.rc.DrawCircle(int(shape.X), int(shape.Y), int(shape.Size),
				shape.R, shape.G, shape.B, shape.A)
		case "rect":
			size := int(shape.Size)
			app.rc.FillRectangle(int(shape.X)-size/2, int(shape.Y)-size/2, size, size,
				shape.R, shape.G, shape.B, shape.A)
		case "line":
			size := int(shape.Size)
			app.rc.DrawLine(int(shape.X)-size, int(shape.Y)-size,
				int(shape.X)+size, int(shape.Y)+size,
				shape.R, shape.G, shape.B, shape.A)
		}
	}
}

// drawUI draws the user interface elements
func (app *X11DemoApp) drawUI() {
	// Draw simple text info (using pixel manipulation)
	info := fmt.Sprintf("Frame: %d Shapes: %d Mouse: (%d,%d)",
		app.frameCount, len(app.shapes), app.mouseX, app.mouseY)
	app.drawText(info, 10, 10, 255, 255, 255, 255)

	// Draw crosshair at mouse position
	if app.mouseX >= 0 && app.mouseY >= 0 {
		app.rc.DrawLine(app.mouseX-10, app.mouseY, app.mouseX+10, app.mouseY, 255, 255, 255, 128)
		app.rc.DrawLine(app.mouseX, app.mouseY-10, app.mouseX, app.mouseY+10, 255, 255, 255, 128)
	}
}

// drawText draws simple text using pixel manipulation (basic bitmap font)
func (app *X11DemoApp) drawText(text string, x, y int, r, g, b, a uint8) {
	// Very basic character rendering - 8x8 pixels per character
	for i, char := range text {
		if char >= 32 && char <= 126 {
			app.drawChar(char, x+i*8, y, r, g, b, a)
		}
	}
}

// drawChar draws a single character (simplified bitmap font)
func (app *X11DemoApp) drawChar(char rune, x, y int, r, g, b, a uint8) {
	// Very simple 8x8 bitmap for basic characters
	// This would normally use a proper font rendering system
	switch char {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// Draw a simple box for digits
		app.rc.FillRectangle(x, y, 6, 8, r/2, g/2, b/2, a)
		app.rc.FillRectangle(x+1, y+1, 4, 6, r, g, b, a)
	case ' ':
		// Space - do nothing
	default:
		// Default character representation
		app.rc.FillRectangle(x, y, 6, 8, r, g, b, a/2)
	}
}

// removeNearbyShapes removes shapes within the specified radius
func (app *X11DemoApp) removeNearbyShapes(x, y, radius float64) {
	newShapes := app.shapes[:0]
	removed := 0

	for _, shape := range app.shapes {
		dx := shape.X - x
		dy := shape.Y - y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > radius {
			newShapes = append(newShapes, shape)
		} else {
			removed++
		}
	}

	app.shapes = newShapes
	if removed > 0 {
		fmt.Printf("Removed %d shapes\n", removed)
	}
}

// showInfo displays application information
func (app *X11DemoApp) showInfo() {
	fmt.Println("\n=== X11 Demo Information ===")

	// Get backend info
	if nativeHandle := app.backend.GetNativeHandle(); nativeHandle != nil {
		fmt.Printf("Backend: X11 (native handle available)\n")
	}

	fmt.Printf("Window size: %dx%d\n", app.ps.Width(), app.ps.Height())
	fmt.Printf("Pixel format: %s (%d bpp)\n", app.ps.Format().String(), app.ps.BPP())
	fmt.Printf("Frame count: %d\n", app.frameCount)
	fmt.Printf("Active shapes: %d\n", len(app.shapes))
	fmt.Printf("Mouse position: (%d, %d)\n", app.mouseX, app.mouseY)

	// Get rendering context statistics
	if stats := app.rc.Statistics(); stats != nil {
		fmt.Printf("Rendering statistics:\n")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	fmt.Printf("Ticks: %d\n", app.backend.GetTicks())
	fmt.Println("============================\n")
}

// Run runs the demo application
func (app *X11DemoApp) Run() error {
	// Set event callbacks
	if eventCallback, ok := app.backend.(interface{ SetEventCallback(platform.EventCallback) }); ok {
		eventCallback.SetEventCallback(app)
	}

	// Initialize the backend
	err := app.backend.Init(800, 600, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}
	defer app.backend.Destroy()

	// Initialize platform support
	err = app.ps.Init(800, 600, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize platform support: %w", err)
	}

	fmt.Println("X11 Demo starting...")
	fmt.Println("Use mouse and keyboard to interact with the demo")

	// Initial draw
	app.backend.ForceRedraw()

	// Run the event loop
	return nil
}

func main() {
	fmt.Println("AGG X11 Platform Demo")
	fmt.Println("====================")

	// Create the demo application
	app, err := NewX11DemoApp()
	if err != nil {
		log.Fatalf("Failed to create demo application: %v", err)
	}

	// Run the application
	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run demo application: %v", err)
	}

	// For this demo, we'll run a simple event loop simulation
	// In a real application, app.backend.Run() would handle this
	fmt.Println("Running event loop simulation...")

	// Simulate some events for demonstration
	for i := 0; i < 100; i++ {
		app.OnDraw()
		time.Sleep(50 * time.Millisecond)

		// Simulate some mouse movements
		if i%10 == 0 {
			app.OnMouseMove(100+i*5, 100+i*3, 0)
		}

		// Simulate some key presses
		if i == 50 {
			app.OnKey(0, 0, platform.KeyCode('r'), 0)
		}
	}

	fmt.Println("Demo completed successfully!")
}
