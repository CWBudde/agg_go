// Package main demonstrates the basic usage of AGG platform support.
// This example shows how to create a platform support instance, set up event handlers,
// and perform basic rendering operations.
package main

import (
	"fmt"
	"time"

	"agg_go/internal/platform"
)

// Application represents our demo application
type Application struct {
	*platform.BaseEventHandler
	ps         *platform.PlatformSupport
	rc         *platform.RenderingContext
	frameCount int
}

// NewApplication creates a new demo application
func NewApplication() *Application {
	app := &Application{}

	// Create platform support with RGBA32 format
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG Platform Support Demo")

	// Create rendering context
	app.rc = platform.NewRenderingContext(app.ps)

	return app
}

// OnInit is called when the application is initialized
func (app *Application) OnInit() {
	fmt.Println("Application initialized")
	fmt.Printf("Pixel format: %s (%d bpp)\n",
		app.ps.Format().String(), app.ps.BPP())
	fmt.Printf("Window size: %dx%d\n",
		app.ps.Width(), app.ps.Height())
}

// OnResize is called when the window is resized
func (app *Application) OnResize(width, height int) {
	fmt.Printf("Window resized to %dx%d\n", width, height)

	// Update the resize transformation
	app.rc.SetupResizeTransform(width, height)

	// Force a redraw
	app.ps.ForceRedraw()
}

// OnMouseMove is called when the mouse is moved
func (app *Application) OnMouseMove(x, y int, flags platform.InputFlags) {
	// Transform screen coordinates to logical coordinates
	lx, ly := app.rc.InverseTransformPoint(float64(x), float64(y))

	if flags.HasMouseLeft() {
		// Draw a small square where the mouse is
		app.rc.FillRectangle(int(lx)-2, int(ly)-2, 4, 4, 255, 0, 0, 255)
		app.ps.UpdateWindow()
	}
}

// OnMouseButtonDown is called when a mouse button is pressed
func (app *Application) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	fmt.Printf("Mouse button down at (%d, %d), flags: %s\n", x, y, flags.String())

	if flags.HasMouseLeft() {
		// Draw a circle at the click position
		app.rc.DrawCircle(x, y, 10, 0, 255, 0, 255)
		app.ps.UpdateWindow()
	}
}

// OnKey is called when a key is pressed
func (app *Application) OnKey(x, y int, key platform.KeyCode, flags platform.InputFlags) {
	fmt.Printf("Key pressed: %s at (%d, %d), flags: %s\n",
		key.String(), x, y, flags.String())

	switch key {
	case platform.KeyEscape:
		fmt.Println("Escape pressed - would exit in a real application")
	case platform.KeyCode('c'), platform.KeyCode('C'):
		if flags.HasCtrl() {
			fmt.Println("Ctrl+C pressed - would copy in a real application")
		} else {
			// Clear the window
			app.rc.ClearWindow(0, 0, 0, 255)
			app.ps.UpdateWindow()
		}
	case platform.KeyF1:
		app.showStatistics()
	}
}

// OnDraw is called when the window needs to be redrawn
func (app *Application) OnDraw() {
	app.frameCount++

	// Clear background with a gradient-like effect
	bgColor := uint8((app.frameCount * 2) % 256)
	app.rc.ClearWindow(bgColor/4, bgColor/8, bgColor/2, 255)

	// Draw some test patterns
	app.drawTestPattern()

	// Update the window
	app.ps.UpdateWindow()
}

// OnIdle is called when the application is idle
func (app *Application) OnIdle() {
	// In a real application, you might perform background tasks here
	// For this demo, we'll just trigger a redraw occasionally
	if app.frameCount%60 == 0 {
		app.ps.ForceRedraw()
	}
}

// drawTestPattern draws a simple test pattern to demonstrate rendering
func (app *Application) drawTestPattern() {
	width := app.ps.Width()
	height := app.ps.Height()

	// Draw a grid pattern
	for x := 0; x < width; x += 50 {
		app.rc.DrawLine(x, 0, x, height-1, 128, 128, 128, 255)
	}
	for y := 0; y < height; y += 50 {
		app.rc.DrawLine(0, y, width-1, y, 128, 128, 128, 255)
	}

	// Draw some shapes
	centerX, centerY := width/2, height/2

	// Red rectangle
	app.rc.FillRectangle(centerX-100, centerY-50, 80, 40, 255, 0, 0, 255)

	// Green circle outline
	app.rc.DrawCircle(centerX, centerY, 60, 0, 255, 0, 255)

	// Blue filled rectangle with transparency
	for dy := 0; dy < 30; dy++ {
		for dx := 0; dx < 60; dx++ {
			app.rc.BlendPixel(centerX+20+dx, centerY+20+dy, 0, 0, 255, 128)
		}
	}

	// Draw frame counter
	app.drawText(fmt.Sprintf("Frame: %d", app.frameCount), 10, 10)
}

// drawText draws simple text using pixel manipulation (very basic)
func (app *Application) drawText(text string, x, y int) {
	// This is a very simple text rendering - just draw pixels for demonstration
	for i, char := range text {
		if char >= 32 && char <= 126 {
			app.drawChar(char, x+i*8, y)
		}
	}
}

// drawChar draws a single character using a simple bitmap font
func (app *Application) drawChar(char rune, x, y int) {
	// Very simple 8x8 bitmap for some characters
	// This is just for demonstration - a real implementation would use proper fonts
	var bitmap [][]bool

	switch char {
	case 'F':
		bitmap = [][]bool{
			{true, true, true, true, true, true, true, false},
			{true, false, false, false, false, false, false, false},
			{true, false, false, false, false, false, false, false},
			{true, true, true, true, true, false, false, false},
			{true, false, false, false, false, false, false, false},
			{true, false, false, false, false, false, false, false},
			{true, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case 'r':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, true, true, true, false, false, false, false},
			{false, true, false, false, true, false, false, false},
			{false, true, false, false, false, false, false, false},
			{false, true, false, false, false, false, false, false},
			{false, true, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case 'a':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, true, true, true, false, false, false},
			{false, false, false, false, false, true, false, false},
			{false, false, true, true, true, true, false, false},
			{false, true, false, false, false, true, false, false},
			{false, false, true, true, true, true, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case 'm':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{true, true, false, true, true, false, false, false},
			{true, false, true, false, false, true, false, false},
			{true, false, true, false, false, true, false, false},
			{true, false, true, false, false, true, false, false},
			{true, false, true, false, false, true, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case 'e':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, true, true, true, false, false, false},
			{false, true, false, false, false, true, false, false},
			{false, true, true, true, true, true, false, false},
			{false, true, false, false, false, false, false, false},
			{false, false, true, true, true, true, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case ':':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, true, true, false, false, false, false},
			{false, false, true, true, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, true, true, false, false, false, false},
			{false, false, true, true, false, false, false, false},
			{false, false, false, false, false, false, false, false},
		}
	case ' ':
		bitmap = [][]bool{
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false},
		}
	default:
		// For digits and other characters, draw a simple block
		if char >= '0' && char <= '9' {
			bitmap = [][]bool{
				{false, true, true, true, true, false, false, false},
				{true, false, false, false, false, true, false, false},
				{true, false, false, false, false, true, false, false},
				{true, false, false, false, false, true, false, false},
				{true, false, false, false, false, true, false, false},
				{true, false, false, false, false, true, false, false},
				{false, true, true, true, true, false, false, false},
				{false, false, false, false, false, false, false, false},
			}
		} else {
			// Default block for unknown characters
			bitmap = [][]bool{
				{true, true, true, true, true, true, true, true},
				{true, false, false, false, false, false, false, true},
				{true, false, false, false, false, false, false, true},
				{true, false, false, false, false, false, false, true},
				{true, false, false, false, false, false, false, true},
				{true, false, false, false, false, false, false, true},
				{true, false, false, false, false, false, false, true},
				{true, true, true, true, true, true, true, true},
			}
		}
	}

	// Draw the bitmap
	for row := 0; row < len(bitmap); row++ {
		for col := 0; col < len(bitmap[row]); col++ {
			if bitmap[row][col] {
				app.rc.SetPixel(x+col, y+row, 255, 255, 255, 255)
			}
		}
	}
}

// showStatistics displays platform and rendering statistics
func (app *Application) showStatistics() {
	fmt.Println("\n=== Platform Statistics ===")
	stats := app.rc.Statistics()

	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Printf("Current caption: %s\n", app.ps.GetCaption())
	fmt.Printf("Wait mode: %v\n", app.ps.WaitMode())
	fmt.Printf("Window flags: %d\n", app.ps.WindowFlags())
	fmt.Println("===========================")
}

// simulateInteraction simulates user interaction for demonstration
func (app *Application) simulateInteraction() {
	fmt.Println("Starting simulated interaction...")

	// Simulate initialization
	app.ps.TriggerDraw()

	// Simulate some mouse movements
	app.ps.TriggerMouseMove(100, 100, platform.MouseLeft)
	app.ps.TriggerMouseMove(150, 120, platform.MouseLeft)
	app.ps.TriggerMouseMove(200, 140, 0)

	// Simulate mouse clicks
	app.ps.TriggerMouseDown(250, 200, platform.MouseLeft)
	app.ps.TriggerMouseUp(250, 200, platform.MouseLeft)

	// Simulate keyboard input
	app.ps.TriggerKey(300, 250, platform.KeyCode('c'), platform.KbdCtrl)
	app.ps.TriggerKey(300, 250, platform.KeyF1, 0)
	app.ps.TriggerKey(300, 250, platform.KeyEscape, 0)

	// Simulate a resize
	app.ps.TriggerResize(900, 700)

	// Simulate some idle cycles
	for i := 0; i < 5; i++ {
		app.ps.TriggerIdle()
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	fmt.Println("AGG Platform Support Demo")
	fmt.Println("========================")

	// Create the application
	app := NewApplication()

	// Set up event handlers
	app.ps.SetOnInit(app.OnInit)
	app.ps.SetOnResize(app.OnResize)
	app.ps.SetOnMouseMove(app.OnMouseMove)
	app.ps.SetOnMouseDown(app.OnMouseButtonDown)
	app.ps.SetOnKey(app.OnKey)
	app.ps.SetOnDraw(app.OnDraw)
	app.ps.SetOnIdle(app.OnIdle)

	// Initialize the platform support
	err := app.ps.Init(800, 600, platform.WindowResize|platform.WindowKeepAspectRatio)
	if err != nil {
		fmt.Printf("Failed to initialize platform support: %v\n", err)
		return
	}

	fmt.Println("Platform support initialized successfully")

	// Start the timer for performance measurement
	app.ps.StartTimer()

	// Create and test some image buffers
	fmt.Println("Testing image buffer operations...")
	app.ps.CreateImage(0, 200, 150)
	app.ps.CreateImage(1, 100, 100)
	app.rc.ClearImage(0, 128, 64, 192, 255)
	app.rc.ClearImage(1, 64, 128, 64, 255)

	// Test image copy operations
	app.ps.CopyImageToWindow(0)
	app.ps.CopyWindowToImage(2)
	app.ps.CopyImageToImage(3, 1)

	fmt.Printf("Elapsed time after setup: %.2f ms\n", app.ps.ElapsedTime())

	// Demonstrate event handling
	app.simulateInteraction()

	// Show final statistics
	app.showStatistics()

	// Test different pixel formats
	fmt.Println("Testing different pixel formats...")
	formats := []platform.PixelFormat{
		platform.PixelFormatRGB24,
		platform.PixelFormatBGRA32,
		platform.PixelFormatGray8,
	}

	for _, format := range formats {
		testPS := platform.NewPlatformSupport(format, false)
		err := testPS.Init(100, 100, 0)
		if err != nil {
			fmt.Printf("Failed to test format %s: %v\n", format.String(), err)
			continue
		}

		testRC := platform.NewRenderingContext(testPS)
		testRC.ClearWindow(255, 128, 64, 255)

		fmt.Printf("Successfully tested format: %s (%d bpp)\n",
			format.String(), format.BPP())
	}

	fmt.Printf("Final elapsed time: %.2f ms\n", app.ps.ElapsedTime())
	fmt.Println("Demo completed successfully!")

	// Display usage instructions
	fmt.Println("\nIn a real windowing application, you would:")
	fmt.Println("- Use actual window creation APIs (SDL2, GLFW, etc.)")
	fmt.Println("- Connect real mouse and keyboard events")
	fmt.Println("- Display the rendered buffer to an actual window")
	fmt.Println("- Handle window close events and cleanup")
	fmt.Println("\nThis demo shows the core AGG platform support functionality")
	fmt.Println("that would be used as a foundation for such applications.")
}
