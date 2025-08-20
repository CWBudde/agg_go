//go:build sdl2
// +build sdl2

// Package main demonstrates the SDL2 platform backend for AGG.
// This example shows how to create interactive graphics applications using SDL2.
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"agg_go/internal/platform"
)

// SDL2DemoApp represents our demo application
type SDL2DemoApp struct {
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
	running     bool
}

// Shape represents a drawable shape
type Shape struct {
	X, Y       float64
	VX, VY     float64
	R, G, B, A uint8
	Size       float64
	Type       string
}

// NewSDL2DemoApp creates a new demo application
func NewSDL2DemoApp() (*SDL2DemoApp, error) {
	app := &SDL2DemoApp{
		currentTime: time.Now(),
		shapes:      make([]Shape, 0, 100),
		running:     true,
	}

	// Create backend factory and get SDL2 backend
	factory := platform.GetBackendFactory()
	backend, err := factory.CreateBackend(platform.BackendSDL2, platform.PixelFormatRGBA32, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDL2 backend: %w", err)
	}
	app.backend = backend

	// Create platform support
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG SDL2 Demo - Interactive Graphics")

	// Create rendering context
	app.rc = platform.NewRenderingContext(app.ps)

	// Initialize some demo shapes
	app.initShapes()

	return app, nil
}

// initShapes initializes some animated shapes for the demo
func (app *SDL2DemoApp) initShapes() {
	// Create some bouncing shapes with different colors and behaviors
	for i := 0; i < 25; i++ {
		shape := Shape{
			X:    float64(150 + (i%5)*100),
			Y:    float64(150 + (i/5)*80),
			VX:   float64((i%7 - 3) * 60),
			VY:   float64((i%5 - 2) * 40),
			R:    uint8((i * 60) % 256),
			G:    uint8((i * 95) % 256),
			B:    uint8((i * 140) % 256),
			A:    255,
			Size: float64(8 + i%20),
			Type: []string{"circle", "rect", "line", "triangle"}[i%4],
		}
		app.shapes = append(app.shapes, shape)
	}
}

// OnInit is called when the application is initialized
func (app *SDL2DemoApp) OnInit() {
	fmt.Println("SDL2 Demo initialized")
	fmt.Printf("Window size: %dx%d\n", app.ps.Width(), app.ps.Height())
	fmt.Println("Controls:")
	fmt.Println("  - Move mouse to interact with shapes")
	fmt.Println("  - Left click and drag to create trails and add shapes")
	fmt.Println("  - Right click to remove nearby shapes")
	fmt.Println("  - Press 'c' to clear the screen")
	fmt.Println("  - Press 'r' to reset shapes")
	fmt.Println("  - Press 'ESC' to exit")
}

// OnDestroy is called when the application is being destroyed
func (app *SDL2DemoApp) OnDestroy() {
	fmt.Println("SDL2 Demo shutting down")
	app.running = false
}

// OnResize is called when the window is resized
func (app *SDL2DemoApp) OnResize(width, height int) {
	fmt.Printf("Window resized to %dx%d\n", width, height)
	app.rc.SetupResizeTransform(width, height)
	app.backend.ForceRedraw()
}

// OnMouseMove is called when the mouse is moved
func (app *SDL2DemoApp) OnMouseMove(x, y int, flags platform.InputFlags) {
	app.mouseX = x
	app.mouseY = y

	if app.isDragging || flags.HasMouseLeft() {
		// Create trail effect with varying colors
		trail_r := uint8((app.frameCount + x) % 256)
		trail_g := uint8((app.frameCount + y) % 256)
		trail_b := uint8((app.frameCount + x + y) % 256)
		app.rc.FillRectangle(x-3, y-3, 6, 6, trail_r, trail_g, trail_b, 180)
	}
}

// OnMouseButtonDown is called when a mouse button is pressed
func (app *SDL2DemoApp) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	if flags.HasMouseLeft() {
		app.isDragging = true
		
		// Add a new shape at mouse position with random properties
		shape := Shape{
			X:    float64(x),
			Y:    float64(y),
			VX:   float64((x % 300) - 150),
			VY:   float64((y % 300) - 150),
			R:    uint8((app.frameCount + x) % 256),
			G:    uint8((app.frameCount + y) % 256),
			B:    uint8((app.frameCount + x + y) % 256),
			A:    255,
			Size: float64(10 + (x+y)%25),
			Type: []string{"circle", "rect", "line", "triangle"}[(x+y)%4],
		}
		app.shapes = append(app.shapes, shape)
		fmt.Printf("Added shape at (%d, %d)\n", x, y)
	}

	if flags.HasMouseRight() {
		// Remove nearby shapes
		removed := app.removeNearbyShapes(float64(x), float64(y), 60)
		if removed > 0 {
			fmt.Printf("Removed %d shapes\n", removed)
		}
	}
}

// OnMouseButtonUp is called when a mouse button is released
func (app *SDL2DemoApp) OnMouseButtonUp(x, y int, flags platform.InputFlags) {
	app.isDragging = false
}

// OnKey is called when a key is pressed
func (app *SDL2DemoApp) OnKey(x, y int, key platform.KeyCode, flags platform.InputFlags) {
	switch key {
	case platform.KeyEscape:
		fmt.Println("Escape pressed - exiting")
		app.running = false

	case platform.KeyCode('c'), platform.KeyCode('C'):
		// Clear screen
		app.rc.ClearWindow(0, 0, 0, 255)
		fmt.Println("Screen cleared")

	case platform.KeyCode('r'), platform.KeyCode('R'):
		// Reset shapes
		app.shapes = app.shapes[:0]
		app.initShapes()
		fmt.Println("Shapes reset")
	}
}

// OnDraw is called when the window needs to be redrawn
func (app *SDL2DemoApp) OnDraw() {
	app.frameCount++

	// Clear background with animated gradient
	bgR := uint8(math.Sin(float64(app.frameCount)*0.005)*20 + 30)
	bgG := uint8(math.Sin(float64(app.frameCount)*0.007)*15 + 20)
	bgB := uint8(math.Sin(float64(app.frameCount)*0.009)*25 + 35)
	app.rc.ClearWindow(bgR, bgG, bgB, 255)

	// Update and draw shapes
	app.updateShapes()
	app.drawShapes()

	// Draw UI elements
	app.drawUI()
}

// OnIdle is called when the application is idle
func (app *SDL2DemoApp) OnIdle() {
	// Continuous redraw for smooth animation
	if app.running {
		app.backend.ForceRedraw()
		app.backend.Delay(16) // ~60 FPS
	}
}

// updateShapes updates the positions and properties of animated shapes
func (app *SDL2DemoApp) updateShapes() {
	dt := time.Since(app.currentTime).Seconds()
	if dt > 0.1 { // Cap delta time to prevent large jumps
		dt = 0.016 // ~60 FPS equivalent
	}
	app.currentTime = time.Now()

	width := float64(app.ps.Width())
	height := float64(app.ps.Height())

	for i := range app.shapes {
		shape := &app.shapes[i]

		// Update position
		shape.X += shape.VX * dt
		shape.Y += shape.VY * dt

		// Bounce off walls with slight energy loss
		if shape.X < shape.Size || shape.X > width-shape.Size {
			shape.VX = -shape.VX * 0.95
			shape.X = math.Max(shape.Size, math.Min(width-shape.Size, shape.X))
		}
		if shape.Y < shape.Size || shape.Y > height-shape.Size {
			shape.VY = -shape.VY * 0.95
			shape.Y = math.Max(shape.Size, math.Min(height-shape.Size, shape.Y))
		}

		// Add attraction/repulsion to mouse cursor
		if app.isDragging {
			dx := float64(app.mouseX) - shape.X
			dy := float64(app.mouseY) - shape.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 && dist < 150 {
				force := 150.0 / (dist + 1)
				shape.VX += dx / dist * force * dt
				shape.VY += dy / dist * force * dt
			}
		}

		// Add gravity and air resistance
		shape.VY += 100 * dt // Gravity
		shape.VX *= 0.999    // Air resistance
		shape.VY *= 0.999
	}

	// Remove shapes that are too far out of bounds
	if len(app.shapes) > 200 {
		app.shapes = app.shapes[:150] // Keep only recent shapes
	}
}

// drawShapes renders all the animated shapes
func (app *SDL2DemoApp) drawShapes() {
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
		case "triangle":
			// Draw a simple triangle using lines
			size := int(shape.Size)
			x, y := int(shape.X), int(shape.Y)
			app.rc.DrawLine(x, y-size, x-size, y+size, shape.R, shape.G, shape.B, shape.A)
			app.rc.DrawLine(x-size, y+size, x+size, y+size, shape.R, shape.G, shape.B, shape.A)
			app.rc.DrawLine(x+size, y+size, x, y-size, shape.R, shape.G, shape.B, shape.A)
		}
	}
}

// drawUI draws the user interface elements
func (app *SDL2DemoApp) drawUI() {
	// Draw crosshair at mouse position
	if app.mouseX >= 0 && app.mouseY >= 0 {
		app.rc.DrawLine(app.mouseX-15, app.mouseY, app.mouseX+15, app.mouseY, 255, 255, 255, 100)
		app.rc.DrawLine(app.mouseX, app.mouseY-15, app.mouseX, app.mouseY+15, 255, 255, 255, 100)
	}
}

// removeNearbyShapes removes shapes within the specified radius
func (app *SDL2DemoApp) removeNearbyShapes(x, y, radius float64) int {
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
	return removed
}

// Run runs the demo application
func (app *SDL2DemoApp) Run() error {
	// Set event callbacks
	if eventCallbackSetter, ok := app.backend.(platform.EventCallbackSetter); ok {
		eventCallbackSetter.SetEventCallback(app)
	}

	// Initialize platform support first
	err := app.ps.Init(900, 700, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize platform support: %w", err)
	}

	// Initialize the backend
	err = app.backend.Init(900, 700, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}
	defer app.backend.Destroy()

	fmt.Println("SDL2 Demo starting...")

	// Run the event loop
	for app.running {
		if !app.backend.PollEvents() {
			break
		}
		app.OnIdle()
	}

	return nil
}

func main() {
	fmt.Println("AGG SDL2 Platform Demo")
	fmt.Println("=====================")

	app, err := NewSDL2DemoApp()
	if err != nil {
		log.Fatalf("Failed to create demo application: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run demo application: %v", err)
	}

	fmt.Println("Demo completed successfully!")
}