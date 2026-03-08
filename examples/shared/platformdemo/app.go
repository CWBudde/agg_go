// Package platformdemo provides the shared interactive demo application logic
// used by all platform backends (X11, SDL2, etc.).
package platformdemo

import (
	"fmt"
	"math"
	"time"

	"agg_go/internal/platform"
)

// shape represents a single animated drawable shape.
type shape struct {
	x, y   float64
	vx, vy float64
	r, g, b, a uint8
	size   float64
	kind   string // "circle", "rect", "line", "triangle"
}

// App is the shared interactive demo application.
// Create it with New, then call Run.
type App struct {
	platform.BaseEventHandler
	backend     platform.PlatformBackend
	ps          *platform.PlatformSupport
	rc          *platform.RenderingContext
	frameCount  int
	mouseX      int
	mouseY      int
	isDragging  bool
	shapes      []shape
	currentTime time.Time
	running     bool
}

const (
	windowWidth  = 800
	windowHeight = 600
)

// New creates a new demo App around the given (already-created) backend.
func New(backend platform.PlatformBackend) *App {
	app := &App{
		backend:     backend,
		currentTime: time.Now(),
		shapes:      make([]shape, 0, 200),
		running:     true,
	}
	app.ps = platform.NewPlatformSupport(platform.PixelFormatRGBA32, false)
	app.ps.Caption("AGG Interactive Demo")
	app.rc = platform.NewRenderingContext(app.ps)
	app.initShapes()
	return app
}

func (app *App) initShapes() {
	app.shapes = app.shapes[:0]
	for i := range 25 {
		app.shapes = append(app.shapes, shape{
			x:    float64(150 + (i%5)*100),
			y:    float64(150 + (i/5)*80),
			vx:   float64((i%7 - 3) * 60),
			vy:   float64((i%5 - 2) * 40),
			r:    uint8((i * 60) % 256),
			g:    uint8((i * 95) % 256),
			b:    uint8((i * 140) % 256),
			a:    255,
			size: float64(8 + i%20),
			kind: []string{"circle", "rect", "line", "triangle"}[i%4],
		})
	}
}

// OnInit is called when the application is initialized.
func (app *App) OnInit() {
	fmt.Printf("Window size: %dx%d\n", app.ps.Width(), app.ps.Height())
	fmt.Println("Controls:")
	fmt.Println("  Mouse move          interact with shapes")
	fmt.Println("  Left click+drag     attract shapes, leave colored trail")
	fmt.Println("  Right click         remove nearby shapes")
	fmt.Println("  c                   clear screen")
	fmt.Println("  r                   reset shapes")
	fmt.Println("  ESC                 exit")
}

// OnDestroy is called when the application is shutting down.
func (app *App) OnDestroy() {
	app.running = false
}

// OnResize is called when the window is resized.
func (app *App) OnResize(width, height int) {
	app.rc.SetupResizeTransform(width, height)
	app.backend.ForceRedraw()
}

// OnMouseMove handles mouse movement.
func (app *App) OnMouseMove(x, y int, flags platform.InputFlags) {
	app.mouseX = x
	app.mouseY = y
	if app.isDragging || flags.HasMouseLeft() {
		tr := uint8((app.frameCount + x) % 256)
		tg := uint8((app.frameCount + y) % 256)
		tb := uint8((app.frameCount + x + y) % 256)
		app.rc.FillRectangle(x-3, y-3, 6, 6, tr, tg, tb, 180)
		_ = app.backend.UpdateWindow(app.ps.WindowBuffer())
	}
}

// OnMouseButtonDown handles mouse button presses.
func (app *App) OnMouseButtonDown(x, y int, flags platform.InputFlags) {
	if flags.HasMouseLeft() {
		app.isDragging = true
		s := shape{
			x:    float64(x),
			y:    float64(y),
			vx:   float64((x%300) - 150),
			vy:   float64((y%300) - 150),
			r:    uint8((app.frameCount + x) % 256),
			g:    uint8((app.frameCount + y) % 256),
			b:    uint8((app.frameCount + x + y) % 256),
			a:    255,
			size: float64(10 + (x+y)%25),
			kind: []string{"circle", "rect", "line", "triangle"}[(x+y)%4],
		}
		app.shapes = append(app.shapes, s)
	}
	if flags.HasMouseRight() {
		n := app.removeNearby(float64(x), float64(y), 60)
		if n > 0 {
			fmt.Printf("Removed %d shape(s)\n", n)
		}
	}
}

// OnMouseButtonUp handles mouse button releases.
func (app *App) OnMouseButtonUp(x, y int, flags platform.InputFlags) {
	app.isDragging = false
}

// OnKey handles key presses.
func (app *App) OnKey(x, y int, key platform.KeyCode, flags platform.InputFlags) {
	switch key {
	case platform.KeyEscape:
		app.running = false
	case platform.KeyCode('c'), platform.KeyCode('C'):
		app.rc.ClearWindow(0, 0, 0, 255)
		fmt.Println("Screen cleared")
	case platform.KeyCode('r'), platform.KeyCode('R'):
		app.initShapes()
		fmt.Println("Shapes reset")
	}
}

// OnDraw renders one frame.
func (app *App) OnDraw() {
	app.frameCount++

	// Animated RGB background.
	fc := float64(app.frameCount)
	bgR := uint8(math.Sin(fc*0.005)*20 + 30)
	bgG := uint8(math.Sin(fc*0.007)*15 + 20)
	bgB := uint8(math.Sin(fc*0.009)*25 + 35)
	app.rc.ClearWindow(bgR, bgG, bgB, 255)

	app.updateShapes()
	app.drawShapes()
	app.drawCrosshair()

	_ = app.backend.UpdateWindow(app.ps.WindowBuffer())
}

// OnIdle drives animation at ~60 fps.
func (app *App) OnIdle() {
	if app.running {
		app.backend.ForceRedraw()
		app.backend.Delay(16)
	}
}

func (app *App) updateShapes() {
	dt := time.Since(app.currentTime).Seconds()
	if dt > 0.1 {
		dt = 0.016
	}
	app.currentTime = time.Now()

	w := float64(app.ps.Width())
	h := float64(app.ps.Height())

	for i := range app.shapes {
		s := &app.shapes[i]

		s.x += s.vx * dt
		s.y += s.vy * dt

		// Mouse attraction while dragging.
		if app.isDragging {
			dx := float64(app.mouseX) - s.x
			dy := float64(app.mouseY) - s.y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 && dist < 150 {
				force := 150.0 / (dist + 1)
				s.vx += dx / dist * force * dt
				s.vy += dy / dist * force * dt
			}
		}

		// Gravity and air resistance.
		s.vy += 100 * dt
		s.vx *= 0.999
		s.vy *= 0.999

		// Bounce off walls with slight energy loss.
		if s.x < s.size || s.x > w-s.size {
			s.vx = -s.vx * 0.95
			s.x = math.Max(s.size, math.Min(w-s.size, s.x))
		}
		if s.y < s.size || s.y > h-s.size {
			s.vy = -s.vy * 0.95
			s.y = math.Max(s.size, math.Min(h-s.size, s.y))
		}
	}

	// Cap total shape count.
	if len(app.shapes) > 200 {
		app.shapes = app.shapes[len(app.shapes)-150:]
	}
}

func (app *App) drawShapes() {
	for _, s := range app.shapes {
		xi, yi, sz := int(s.x), int(s.y), int(s.size)
		switch s.kind {
		case "circle":
			app.rc.DrawCircle(xi, yi, sz, s.r, s.g, s.b, s.a)
		case "rect":
			app.rc.FillRectangle(xi-sz/2, yi-sz/2, sz, sz, s.r, s.g, s.b, s.a)
		case "line":
			app.rc.DrawLine(xi-sz, yi-sz, xi+sz, yi+sz, s.r, s.g, s.b, s.a)
		case "triangle":
			app.rc.DrawLine(xi, yi-sz, xi-sz, yi+sz, s.r, s.g, s.b, s.a)
			app.rc.DrawLine(xi-sz, yi+sz, xi+sz, yi+sz, s.r, s.g, s.b, s.a)
			app.rc.DrawLine(xi+sz, yi+sz, xi, yi-sz, s.r, s.g, s.b, s.a)
		}
	}
}

func (app *App) drawCrosshair() {
	if app.mouseX < 0 && app.mouseY < 0 {
		return
	}
	app.rc.DrawLine(app.mouseX-15, app.mouseY, app.mouseX+15, app.mouseY, 255, 255, 255, 100)
	app.rc.DrawLine(app.mouseX, app.mouseY-15, app.mouseX, app.mouseY+15, 255, 255, 255, 100)
}

func (app *App) removeNearby(x, y, radius float64) int {
	out := app.shapes[:0]
	removed := 0
	for _, s := range app.shapes {
		dx, dy := s.x-x, s.y-y
		if math.Sqrt(dx*dx+dy*dy) > radius {
			out = append(out, s)
		} else {
			removed++
		}
	}
	app.shapes = out
	return removed
}

// Run initializes the window and starts the event loop.
func (app *App) Run() error {
	if setter, ok := app.backend.(platform.EventCallbackSetter); ok {
		setter.SetEventCallback(app)
	}

	if err := app.ps.Init(windowWidth, windowHeight, platform.WindowResize); err != nil {
		return fmt.Errorf("platform support init: %w", err)
	}
	if err := app.backend.Init(windowWidth, windowHeight, platform.WindowResize); err != nil {
		return fmt.Errorf("backend init: %w", err)
	}
	defer app.backend.Destroy()

	for app.running {
		if !app.backend.PollEvents() {
			break
		}
		app.OnIdle()
	}
	return nil
}
