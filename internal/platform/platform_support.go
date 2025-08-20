// Package platform provides the core platform support infrastructure for AGG applications.
// It's designed to create interactive demo examples with basic window management,
// event handling, and rendering capabilities.
//
// This is a Go port of the AGG 2.6 platform support system, adapted to be
// platform-agnostic and focused on testing and demonstration purposes.
package platform

import (
	"fmt"
	"time"

	"agg_go/internal/buffer"
	"agg_go/internal/platform/types"
)

// Type aliases to avoid breaking existing code while using shared types
type (
	WindowFlags   = types.WindowFlags
	PixelFormat   = types.PixelFormat
	InputFlags    = types.InputFlags
	KeyCode       = types.KeyCode
	EventCallback = types.EventCallback
)

// Re-export constants for backward compatibility
const (
	WindowResize          = types.WindowResize
	WindowHWBuffer        = types.WindowHWBuffer
	WindowKeepAspectRatio = types.WindowKeepAspectRatio
	WindowProcessAllKeys  = types.WindowProcessAllKeys
)

// Re-export pixel format constants for backward compatibility
const (
	PixelFormatUndefined = types.PixelFormatUndefined
	PixelFormatBW        = types.PixelFormatBW
	PixelFormatGray8     = types.PixelFormatGray8
	PixelFormatSGray8    = types.PixelFormatSGray8
	PixelFormatGray16    = types.PixelFormatGray16
	PixelFormatGray32    = types.PixelFormatGray32
	PixelFormatRGB555    = types.PixelFormatRGB555
	PixelFormatRGB565    = types.PixelFormatRGB565
	PixelFormatRGB24     = types.PixelFormatRGB24
	PixelFormatSRGB24    = types.PixelFormatSRGB24
	PixelFormatBGR24     = types.PixelFormatBGR24
	PixelFormatSBGR24    = types.PixelFormatSBGR24
	PixelFormatRGBA32    = types.PixelFormatRGBA32
	PixelFormatSRGBA32   = types.PixelFormatSRGBA32
	PixelFormatARGB32    = types.PixelFormatARGB32
	PixelFormatSARGB32   = types.PixelFormatSARGB32
	PixelFormatABGR32    = types.PixelFormatABGR32
	PixelFormatSABGR32   = types.PixelFormatSABGR32
	PixelFormatBGRA32    = types.PixelFormatBGRA32
	PixelFormatSBGRA32   = types.PixelFormatSBGRA32
	PixelFormatRGB48     = types.PixelFormatRGB48
	PixelFormatSRGB48    = types.PixelFormatSRGB48
	PixelFormatBGR48     = types.PixelFormatBGR48
	PixelFormatSBGR48    = types.PixelFormatSBGR48
	PixelFormatRGBA64    = types.PixelFormatRGBA64
	PixelFormatSRGBA64   = types.PixelFormatSRGBA64
	PixelFormatARGB64    = types.PixelFormatARGB64
	PixelFormatSARGB64   = types.PixelFormatSARGB64
	PixelFormatABGR64    = types.PixelFormatABGR64
	PixelFormatSABGR64   = types.PixelFormatSABGR64
	PixelFormatBGRA64    = types.PixelFormatBGRA64
	PixelFormatSBGRA64   = types.PixelFormatSBGRA64
	PixelFormatRGB96     = types.PixelFormatRGB96
	PixelFormatSRGB96    = types.PixelFormatSRGB96
	PixelFormatBGR96     = types.PixelFormatBGR96
	PixelFormatSBGR96    = types.PixelFormatSBGR96
	PixelFormatRGBA128   = types.PixelFormatRGBA128
	PixelFormatSRGBA128  = types.PixelFormatSRGBA128
	PixelFormatARGB128   = types.PixelFormatARGB128
	PixelFormatSARGB128  = types.PixelFormatSARGB128
	PixelFormatABGR128   = types.PixelFormatABGR128
	PixelFormatSABGR128  = types.PixelFormatSABGR128
	PixelFormatBGRA128   = types.PixelFormatBGRA128
	PixelFormatSBGRA128  = types.PixelFormatSBGRA128
)

// Re-export input flag constants for backward compatibility with examples
const (
	MouseLeft  = types.MouseLeft
	MouseRight = types.MouseRight
	KbdShift   = types.KbdShift
	KbdCtrl    = types.KbdCtrl
)

// Re-export key constants for backward compatibility with examples
const (
	// ASCII set
	KeyBackspace = types.KeyBackspace
	KeyTab       = types.KeyTab
	KeyClear     = types.KeyClear
	KeyReturn    = types.KeyReturn
	KeyPause     = types.KeyPause
	KeyEscape    = types.KeyEscape
	KeyDelete    = types.KeyDelete

	// Keypad
	KeyKP0        = types.KeyKP0
	KeyKP1        = types.KeyKP1
	KeyKP2        = types.KeyKP2
	KeyKP3        = types.KeyKP3
	KeyKP4        = types.KeyKP4
	KeyKP5        = types.KeyKP5
	KeyKP6        = types.KeyKP6
	KeyKP7        = types.KeyKP7
	KeyKP8        = types.KeyKP8
	KeyKP9        = types.KeyKP9
	KeyKPPeriod   = types.KeyKPPeriod
	KeyKPDivide   = types.KeyKPDivide
	KeyKPMultiply = types.KeyKPMultiply
	KeyKPMinus    = types.KeyKPMinus
	KeyKPPlus     = types.KeyKPPlus
	KeyKPEnter    = types.KeyKPEnter
	KeyKPEquals   = types.KeyKPEquals

	// Arrow keys and navigation
	KeyUp       = types.KeyUp
	KeyDown     = types.KeyDown
	KeyRight    = types.KeyRight
	KeyLeft     = types.KeyLeft
	KeyInsert   = types.KeyInsert
	KeyHome     = types.KeyHome
	KeyEnd      = types.KeyEnd
	KeyPageUp   = types.KeyPageUp
	KeyPageDown = types.KeyPageDown

	// Function keys
	KeyF1  = types.KeyF1
	KeyF2  = types.KeyF2
	KeyF3  = types.KeyF3
	KeyF4  = types.KeyF4
	KeyF5  = types.KeyF5
	KeyF6  = types.KeyF6
	KeyF7  = types.KeyF7
	KeyF8  = types.KeyF8
	KeyF9  = types.KeyF9
	KeyF10 = types.KeyF10
	KeyF11 = types.KeyF11
	KeyF12 = types.KeyF12

	// Modifier keys
	KeyNumLock    = types.KeyNumLock
	KeyCapsLock   = types.KeyCapsLock
	KeyScrollLock = types.KeyScrollLock
	KeyRShift     = types.KeyRShift
	KeyLShift     = types.KeyLShift
	KeyRCtrl      = types.KeyRCtrl
	KeyLCtrl      = types.KeyLCtrl
	KeyRAlt       = types.KeyRAlt
	KeyLAlt       = types.KeyLAlt
)

// PlatformSupport provides the core platform support functionality for AGG applications.
// It manages rendering buffers, handles events, and provides basic window operations.
type PlatformSupport struct {
	// Window configuration
	format      PixelFormat
	flipY       bool
	bpp         int
	windowFlags WindowFlags
	caption     string
	waitMode    bool

	// Window dimensions
	initialWidth  int
	initialHeight int
	currentWidth  int
	currentHeight int

	// Rendering buffers
	windowBuffer buffer.RenderingBuffer[uint8]
	imageBuffers [maxImages]buffer.RenderingBuffer[uint8]

	// Timer
	startTime time.Time

	// Event handlers
	onInitHandler       func()
	onResizeHandler     func(width, height int)
	onIdleHandler       func()
	onMouseMoveHandler  func(x, y int, flags InputFlags)
	onMouseDownHandler  func(x, y int, flags InputFlags)
	onMouseUpHandler    func(x, y int, flags InputFlags)
	onKeyHandler        func(x, y int, key KeyCode, flags InputFlags)
	onCtrlChangeHandler func()
	onDrawHandler       func()
	onPostDrawHandler   func(rawHandler interface{})
}

const (
	maxImages = 16 // Maximum number of image buffers
)

// NewPlatformSupport creates a new platform support instance with the specified pixel format and Y-axis orientation.
func NewPlatformSupport(format PixelFormat, flipY bool) *PlatformSupport {
	ps := &PlatformSupport{
		format:   format,
		flipY:    flipY,
		bpp:      format.BPP(),
		waitMode: false,
		caption:  "AGG Application",
	}

	// Initialize buffers
	ps.windowBuffer = *buffer.NewRenderingBuffer[uint8]()
	for i := range ps.imageBuffers {
		ps.imageBuffers[i] = *buffer.NewRenderingBuffer[uint8]()
	}

	return ps
}

// Caption sets the window caption (title).
func (ps *PlatformSupport) Caption(caption string) {
	ps.caption = caption
}

// GetCaption returns the current window caption.
func (ps *PlatformSupport) GetCaption() string {
	return ps.caption
}

// Format returns the pixel format.
func (ps *PlatformSupport) Format() PixelFormat {
	return ps.format
}

// FlipY returns whether the Y-axis is flipped.
func (ps *PlatformSupport) FlipY() bool {
	return ps.flipY
}

// BPP returns the bits per pixel.
func (ps *PlatformSupport) BPP() int {
	return ps.bpp
}

// WaitMode returns the current wait mode setting.
func (ps *PlatformSupport) WaitMode() bool {
	return ps.waitMode
}

// SetWaitMode sets the wait mode. When true, the application waits for events
// and doesn't call OnIdle(). When false, it calls OnIdle() when the event queue is empty.
func (ps *PlatformSupport) SetWaitMode(waitMode bool) {
	ps.waitMode = waitMode
}

// Init initializes the platform support with the specified window dimensions and flags.
func (ps *PlatformSupport) Init(width, height int, flags WindowFlags) error {
	ps.initialWidth = width
	ps.initialHeight = height
	ps.currentWidth = width
	ps.currentHeight = height
	ps.windowFlags = flags

	// Calculate stride based on pixel format
	stride := width * ps.bpp / 8
	bufferSize := stride * height

	// Initialize window buffer
	windowData := make([]uint8, bufferSize)
	ps.windowBuffer.Attach(windowData, width, height, stride)

	// Call initialization handler
	if ps.onInitHandler != nil {
		ps.onInitHandler()
	}

	return nil
}

// WindowFlags returns the current window flags.
func (ps *PlatformSupport) WindowFlags() WindowFlags {
	return ps.windowFlags
}

// Width returns the current window width.
func (ps *PlatformSupport) Width() int {
	return ps.currentWidth
}

// Height returns the current window height.
func (ps *PlatformSupport) Height() int {
	return ps.currentHeight
}

// InitialWidth returns the initial window width.
func (ps *PlatformSupport) InitialWidth() int {
	return ps.initialWidth
}

// InitialHeight returns the initial window height.
func (ps *PlatformSupport) InitialHeight() int {
	return ps.initialHeight
}

// WindowBuffer returns a reference to the main rendering buffer.
func (ps *PlatformSupport) WindowBuffer() *buffer.RenderingBuffer[uint8] {
	return &ps.windowBuffer
}

// ImageBuffer returns a reference to the specified image buffer.
func (ps *PlatformSupport) ImageBuffer(idx int) *buffer.RenderingBuffer[uint8] {
	if idx >= 0 && idx < maxImages {
		return &ps.imageBuffers[idx]
	}
	return nil
}

// CreateImage creates an image buffer with the specified dimensions.
// If width or height is 0, uses the current window dimensions.
func (ps *PlatformSupport) CreateImage(idx int, width, height int) bool {
	if idx < 0 || idx >= maxImages {
		return false
	}

	if width == 0 {
		width = ps.currentWidth
	}
	if height == 0 {
		height = ps.currentHeight
	}

	stride := width * ps.bpp / 8
	bufferSize := stride * height
	imageData := make([]uint8, bufferSize)

	ps.imageBuffers[idx].Attach(imageData, width, height, stride)
	return true
}

// LoadImage loads an image from file (stub implementation).
func (ps *PlatformSupport) LoadImage(idx int, filename string) bool {
	// TODO: Implement image loading (BMP/PPM format)
	// For now, just create an empty image
	return ps.CreateImage(idx, 0, 0)
}

// SaveImage saves an image to file (stub implementation).
func (ps *PlatformSupport) SaveImage(idx int, filename string) bool {
	// TODO: Implement image saving (BMP/PPM format)
	return true
}

// CopyImageToWindow copies the specified image buffer to the window buffer.
func (ps *PlatformSupport) CopyImageToWindow(idx int) {
	if idx >= 0 && idx < maxImages && ps.imageBuffers[idx].Buf() != nil {
		ps.windowBuffer.CopyFrom(&ps.imageBuffers[idx])
	}
}

// CopyWindowToImage copies the window buffer to the specified image buffer.
func (ps *PlatformSupport) CopyWindowToImage(idx int) {
	if idx >= 0 && idx < maxImages {
		ps.CreateImage(idx, ps.windowBuffer.Width(), ps.windowBuffer.Height())
		ps.imageBuffers[idx].CopyFrom(&ps.windowBuffer)
	}
}

// CopyImageToImage copies one image buffer to another.
func (ps *PlatformSupport) CopyImageToImage(idxTo, idxFrom int) {
	if idxFrom >= 0 && idxFrom < maxImages &&
		idxTo >= 0 && idxTo < maxImages &&
		ps.imageBuffers[idxFrom].Buf() != nil {
		fromBuffer := &ps.imageBuffers[idxFrom]
		ps.CreateImage(idxTo, fromBuffer.Width(), fromBuffer.Height())
		ps.imageBuffers[idxTo].CopyFrom(fromBuffer)
	}
}

// ForceRedraw sets a flag to redraw the window on the next event cycle.
func (ps *PlatformSupport) ForceRedraw() {
	// In a real implementation, this would set a redraw flag or send a message
	// For now, we just call the draw handler immediately
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
}

// UpdateWindow immediately updates the window with the current buffer content.
func (ps *PlatformSupport) UpdateWindow() {
	// In a real implementation, this would copy the buffer to the actual window
	// For now, this is a no-op since we don't have actual window display
}

// StartTimer starts the timer for elapsed time measurement.
func (ps *PlatformSupport) StartTimer() {
	ps.startTime = time.Now()
}

// ElapsedTime returns the time elapsed since the last StartTimer() call in milliseconds.
func (ps *PlatformSupport) ElapsedTime() float64 {
	return float64(time.Since(ps.startTime).Nanoseconds()) / 1e6
}

// Message displays a message (stub implementation).
func (ps *PlatformSupport) Message(msg string) {
	// In a real implementation, this would show a message box
	fmt.Println("Platform Message:", msg)
}

// ImageExtension returns the default image file extension for this platform.
func (ps *PlatformSupport) ImageExtension() string {
	return ".bmp" // Default to BMP format
}

// Run starts the main event loop (stub implementation).
func (ps *PlatformSupport) Run() int {
	// In a real implementation, this would start the platform-specific event loop
	// For now, just call the draw handler once
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
	return 0
}

// Event handler setters

// SetOnInit sets the initialization event handler.
func (ps *PlatformSupport) SetOnInit(handler func()) {
	ps.onInitHandler = handler
}

// SetOnResize sets the resize event handler.
func (ps *PlatformSupport) SetOnResize(handler func(width, height int)) {
	ps.onResizeHandler = handler
}

// SetOnIdle sets the idle event handler.
func (ps *PlatformSupport) SetOnIdle(handler func()) {
	ps.onIdleHandler = handler
}

// SetOnMouseMove sets the mouse move event handler.
func (ps *PlatformSupport) SetOnMouseMove(handler func(x, y int, flags InputFlags)) {
	ps.onMouseMoveHandler = handler
}

// SetOnMouseDown sets the mouse button down event handler.
func (ps *PlatformSupport) SetOnMouseDown(handler func(x, y int, flags InputFlags)) {
	ps.onMouseDownHandler = handler
}

// SetOnMouseUp sets the mouse button up event handler.
func (ps *PlatformSupport) SetOnMouseUp(handler func(x, y int, flags InputFlags)) {
	ps.onMouseUpHandler = handler
}

// SetOnKey sets the keyboard event handler.
func (ps *PlatformSupport) SetOnKey(handler func(x, y int, key KeyCode, flags InputFlags)) {
	ps.onKeyHandler = handler
}

// SetOnCtrlChange sets the control change event handler.
func (ps *PlatformSupport) SetOnCtrlChange(handler func()) {
	ps.onCtrlChangeHandler = handler
}

// SetOnDraw sets the draw event handler.
func (ps *PlatformSupport) SetOnDraw(handler func()) {
	ps.onDrawHandler = handler
}

// SetOnPostDraw sets the post-draw event handler.
func (ps *PlatformSupport) SetOnPostDraw(handler func(rawHandler interface{})) {
	ps.onPostDrawHandler = handler
}

// Trigger event handlers for testing purposes

// TriggerResize triggers a resize event.
func (ps *PlatformSupport) TriggerResize(width, height int) {
	ps.currentWidth = width
	ps.currentHeight = height

	// Update window buffer
	stride := width * ps.bpp / 8
	bufferSize := stride * height
	windowData := make([]uint8, bufferSize)
	ps.windowBuffer.Attach(windowData, width, height, stride)

	if ps.onResizeHandler != nil {
		ps.onResizeHandler(width, height)
	}
}

// TriggerMouseMove triggers a mouse move event.
func (ps *PlatformSupport) TriggerMouseMove(x, y int, flags InputFlags) {
	if ps.onMouseMoveHandler != nil {
		ps.onMouseMoveHandler(x, y, flags)
	}
}

// TriggerMouseDown triggers a mouse button down event.
func (ps *PlatformSupport) TriggerMouseDown(x, y int, flags InputFlags) {
	if ps.onMouseDownHandler != nil {
		ps.onMouseDownHandler(x, y, flags)
	}
}

// TriggerMouseUp triggers a mouse button up event.
func (ps *PlatformSupport) TriggerMouseUp(x, y int, flags InputFlags) {
	if ps.onMouseUpHandler != nil {
		ps.onMouseUpHandler(x, y, flags)
	}
}

// TriggerKey triggers a keyboard event.
func (ps *PlatformSupport) TriggerKey(x, y int, key KeyCode, flags InputFlags) {
	if ps.onKeyHandler != nil {
		ps.onKeyHandler(x, y, key, flags)
	}
}

// TriggerIdle triggers an idle event.
func (ps *PlatformSupport) TriggerIdle() {
	if ps.onIdleHandler != nil {
		ps.onIdleHandler()
	}
}

// TriggerDraw triggers a draw event.
func (ps *PlatformSupport) TriggerDraw() {
	if ps.onDrawHandler != nil {
		ps.onDrawHandler()
	}
}
