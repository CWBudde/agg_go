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
)

// WindowFlags represents window configuration flags.
type WindowFlags uint32

const (
	WindowResize           WindowFlags = 1 << 0 // Window can be resized
	WindowHWBuffer         WindowFlags = 1 << 1 // Use hardware buffer (platform dependent)
	WindowKeepAspectRatio  WindowFlags = 1 << 2 // Maintain aspect ratio during resize
	WindowProcessAllKeys   WindowFlags = 1 << 3 // Process all keyboard events
)

// PixelFormat represents the pixel format of the rendering buffer.
type PixelFormat int

const (
	PixelFormatUndefined PixelFormat = iota // No conversions applied
	PixelFormatBW                           // 1 bit per color B/W
	PixelFormatGray8                        // Simple 256 level grayscale
	PixelFormatSGray8                       // Simple 256 level grayscale (sRGB)
	PixelFormatGray16                       // Simple 65535 level grayscale
	PixelFormatGray32                       // Grayscale, one 32-bit float per pixel
	PixelFormatRGB555                       // 15 bit rgb
	PixelFormatRGB565                       // 16 bit rgb
	PixelFormatRGB24                        // R-G-B, one byte per color component
	PixelFormatSRGB24                       // R-G-B, one byte per color component (sRGB)
	PixelFormatBGR24                        // B-G-R, one byte per color component
	PixelFormatSBGR24                       // B-G-R, native win32 BMP format (sRGB)
	PixelFormatRGBA32                       // R-G-B-A, one byte per color component
	PixelFormatSRGBA32                      // R-G-B-A, one byte per color component (sRGB)
	PixelFormatARGB32                       // A-R-G-B, native MAC format
	PixelFormatSARGB32                      // A-R-G-B, native MAC format (sRGB)
	PixelFormatABGR32                       // A-B-G-R, one byte per color component
	PixelFormatSABGR32                      // A-B-G-R, one byte per color component (sRGB)
	PixelFormatBGRA32                       // B-G-R-A, native win32 BMP format
	PixelFormatSBGRA32                      // B-G-R-A, native win32 BMP format (sRGB)
	PixelFormatRGB48                        // R-G-B, 16 bits per color component
	PixelFormatBGR48                        // B-G-R, native win32 BMP format
	PixelFormatRGB96                        // R-G-B, one 32-bit float per color component
	PixelFormatBGR96                        // B-G-R, one 32-bit float per color component
	PixelFormatRGBA64                       // R-G-B-A, 16 bits per color component
	PixelFormatARGB64                       // A-R-G-B, native MAC format
	PixelFormatABGR64                       // A-B-G-R, one byte per color component
	PixelFormatBGRA64                       // B-G-R-A, native win32 BMP format
	PixelFormatRGBA128                      // R-G-B-A, one 32-bit float per color component
	PixelFormatARGB128                      // A-R-G-B, one 32-bit float per color component
	PixelFormatABGR128                      // A-B-G-R, one 32-bit float per color component
	PixelFormatBGRA128                      // B-G-R-A, one 32-bit float per color component
)

// String returns the string representation of the pixel format.
func (pf PixelFormat) String() string {
	formats := []string{
		"Undefined", "BW", "Gray8", "SGray8", "Gray16", "Gray32",
		"RGB555", "RGB565", "RGB24", "SRGB24", "BGR24", "SBGR24",
		"RGBA32", "SRGBA32", "ARGB32", "SARGB32", "ABGR32", "SABGR32",
		"BGRA32", "SBGRA32", "RGB48", "BGR48", "RGB96", "BGR96",
		"RGBA64", "ARGB64", "ABGR64", "BGRA64", "RGBA128", "ARGB128",
		"ABGR128", "BGRA128",
	}
	if int(pf) < len(formats) {
		return formats[pf]
	}
	return fmt.Sprintf("Unknown(%d)", int(pf))
}

// BPP returns the bits per pixel for the pixel format.
func (pf PixelFormat) BPP() int {
	switch pf {
	case PixelFormatBW:
		return 1
	case PixelFormatGray8, PixelFormatSGray8:
		return 8
	case PixelFormatGray16, PixelFormatRGB555, PixelFormatRGB565:
		return 16
	case PixelFormatRGB24, PixelFormatSRGB24, PixelFormatBGR24, PixelFormatSBGR24:
		return 24
	case PixelFormatRGBA32, PixelFormatSRGBA32, PixelFormatARGB32, PixelFormatSARGB32,
		 PixelFormatABGR32, PixelFormatSABGR32, PixelFormatBGRA32, PixelFormatSBGRA32,
		 PixelFormatGray32:
		return 32
	case PixelFormatRGB48, PixelFormatBGR48:
		return 48
	case PixelFormatRGBA64, PixelFormatARGB64, PixelFormatABGR64, PixelFormatBGRA64:
		return 64
	case PixelFormatRGB96, PixelFormatBGR96:
		return 96
	case PixelFormatRGBA128, PixelFormatARGB128, PixelFormatABGR128, PixelFormatBGRA128:
		return 128
	default:
		return 0
	}
}

// PlatformSupport provides the core platform support functionality for AGG applications.
// It manages rendering buffers, handles events, and provides basic window operations.
type PlatformSupport struct {
	// Window configuration
	format       PixelFormat
	flipY        bool
	bpp          int
	windowFlags  WindowFlags
	caption      string
	waitMode     bool

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