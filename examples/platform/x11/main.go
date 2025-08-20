//go:build x11
// +build x11

// Package main demonstrates the X11 platform backend for AGG.
// This example shows how to create interactive graphics applications using X11.
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

// NewX11DemoApp creates a new demo application
func NewX11DemoApp() (*X11DemoApp, error) {
	app := &X11DemoApp{
		currentTime: time.Now(),
		shapes:      make([]Shape, 0, 100),
		running:     true,
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
		app.running = false

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

		// Apply some damping
		shape.VX *= 0.99
		shape.VY *= 0.99

		// Update color cycling
		shape.R = uint8((int(shape.R) + 1) % 256)
	}
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
	// Draw crosshair at mouse position
	if app.mouseX >= 0 && app.mouseY >= 0 {
		app.rc.DrawLine(app.mouseX-10, app.mouseY, app.mouseX+10, app.mouseY, 255, 255, 255, 128)
		app.rc.DrawLine(app.mouseX, app.mouseY-10, app.mouseX, app.mouseY+10, 255, 255, 255, 128)
	}
}

// Run runs the demo application
func (app *X11DemoApp) Run() error {
	// Set event callbacks
	if eventCallbackSetter, ok := app.backend.(platform.EventCallbackSetter); ok {
		eventCallbackSetter.SetEventCallback(app)
	}

	// Initialize platform support first
	err := app.ps.Init(800, 600, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize platform support: %w", err)
	}

	// Initialize the backend
	err = app.backend.Init(800, 600, platform.WindowResize)
	if err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}
	defer app.backend.Destroy()

	fmt.Println("X11 Demo starting...")

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
	fmt.Println("AGG X11 Platform Demo")
	fmt.Println("====================")

	app, err := NewX11DemoApp()
	if err != nil {
		log.Fatalf("Failed to create demo application: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("Failed to run demo application: %v", err)
	}

	fmt.Println("Demo completed successfully!")
}