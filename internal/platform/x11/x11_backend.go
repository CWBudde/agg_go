package x11

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <X11/keysym.h>
#include <stdlib.h>
#include <string.h>

// Helper functions to avoid Go pointer issues
void copyBuffer(char* dest, char* src, int size) {
	memcpy(dest, src, size);
}

XImage* createXImage(Display* display, Visual* visual, unsigned int depth,
					 int format, int offset, char* data,
					 unsigned int width, unsigned int height,
					 int bitmap_pad, int bytes_per_line) {
	return XCreateImage(display, visual, depth, format, offset, data,
						width, height, bitmap_pad, bytes_per_line);
}

// Wrapper for XDestroyImage to ensure proper cleanup
int destroyXImage(XImage* image) {
	if (image) {
		// Don't let X11 free the data since Go manages it
		char* original_data = image->data;
		image->data = NULL;
		int result = XDestroyImage(image);
		return result;
	}
	return 1;
}
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"

	"agg_go/internal/buffer"
	"agg_go/internal/platform/types"
)

// X11Backend implements PlatformBackend for X11
type X11Backend struct {
	// Display connection
	display *C.Display
	screen  C.int
	depth   C.int
	visual  *C.Visual
	window  C.Window
	gc      C.GC

	// Window properties
	caption string
	width   int
	height  int
	format  types.PixelFormat
	flipY   bool
	bpp     int
	sysBpp  int

	// XImage for buffer display
	ximg      *C.XImage
	imgData   []byte
	imgStride int

	// Event handling
	eventCallback types.EventCallback
	wmDeleteAtom  C.Atom

	// State flags
	initialized bool
	shouldClose bool
	startTicks  uint32

	// Keymap for converting X11 keys to AGG keys
	keymap [256]types.KeyCode
}

// NewX11BackendImpl creates a new X11 backend implementation
func NewX11BackendImpl(format types.PixelFormat, flipY bool) (*X11Backend, error) {
	backend := &X11Backend{
		caption: "AGG X11 Window",
		format:  format,
		flipY:   flipY,
		bpp:     format.BPP(),
	}

	// Initialize keymap
	backend.initKeymap()

	return backend, nil
}

// Init initializes the X11 backend
func (x *X11Backend) Init(width, height int, flags types.WindowFlags) error {
	if x.initialized {
		return fmt.Errorf("X11 backend already initialized")
	}

	x.width = width
	x.height = height

	// Open display connection
	x.display = C.XOpenDisplay(nil)
	if x.display == nil {
		return fmt.Errorf("cannot open X11 display")
	}

	// Get default screen
	x.screen = C.XDefaultScreen(x.display)
	x.depth = C.XDefaultDepth(x.display, x.screen)
	x.visual = C.XDefaultVisual(x.display, x.screen)

	// Determine system pixel format and BPP
	x.sysBpp = int(x.depth)
	if x.sysBpp == 24 {
		x.sysBpp = 32 // X11 typically uses 32-bit for 24-bit color
	}

	// Create window
	rootWindow := C.XDefaultRootWindow(x.display)
	x.window = C.XCreateSimpleWindow(
		x.display, rootWindow,
		0, 0, C.uint(width), C.uint(height),
		1, C.XBlackPixel(x.display, x.screen),
		C.XWhitePixel(x.display, x.screen))

	if x.window == 0 {
		C.XCloseDisplay(x.display)
		return fmt.Errorf("failed to create X11 window")
	}

	// Set window properties
	x.setWindowProperties(flags)

	// Create graphics context
	x.gc = C.XCreateGC(x.display, x.window, 0, nil)

	// Set up WM_DELETE_WINDOW protocol
	x.wmDeleteAtom = C.XInternAtom(x.display, C.CString("WM_DELETE_WINDOW"), C.False)
	C.XSetWMProtocols(x.display, x.window, &x.wmDeleteAtom, 1)

	// Select input events
	eventMask := C.ExposureMask | C.KeyPressMask | C.KeyReleaseMask |
		C.ButtonPressMask | C.ButtonReleaseMask | C.PointerMotionMask |
		C.StructureNotifyMask

	C.XSelectInput(x.display, x.window, C.long(eventMask))

	// Create image buffer
	err := x.createImageBuffer()
	if err != nil {
		x.Destroy()
		return fmt.Errorf("failed to create image buffer: %w", err)
	}

	// Map window (make it visible)
	C.XMapWindow(x.display, x.window)
	C.XFlush(x.display)

	x.initialized = true
	x.startTicks = uint32(time.Now().UnixNano() / 1e6)

	// Trigger init callback
	if x.eventCallback != nil {
		x.eventCallback.OnInit()
	}

	return nil
}

// setWindowProperties sets various window properties based on flags
func (x *X11Backend) setWindowProperties(flags types.WindowFlags) {
	// Set window title
	cCaption := C.CString(x.caption)
	defer C.free(unsafe.Pointer(cCaption))
	C.XStoreName(x.display, x.window, cCaption)

	// Handle resize flag
	if flags&types.WindowResize == 0 {
		// Make window non-resizable
		var hints C.XSizeHints
		hints.flags = C.PMinSize | C.PMaxSize
		hints.min_width = C.int(x.width)
		hints.min_height = C.int(x.height)
		hints.max_width = C.int(x.width)
		hints.max_height = C.int(x.height)
		C.XSetWMNormalHints(x.display, x.window, &hints)
	}
}

// createImageBuffer creates the XImage for displaying the rendering buffer
func (x *X11Backend) createImageBuffer() error {
	// Calculate buffer size
	x.imgStride = x.width * x.bpp / 8
	bufferSize := x.imgStride * x.height

	// Allocate image data
	x.imgData = make([]byte, bufferSize)

	// Create XImage
	x.ximg = C.createXImage(
		x.display, x.visual, C.uint(x.depth),
		C.ZPixmap, 0,
		(*C.char)(unsafe.Pointer(&x.imgData[0])),
		C.uint(x.width), C.uint(x.height),
		32, C.int(x.imgStride))

	if x.ximg == nil {
		return fmt.Errorf("failed to create XImage")
	}

	return nil
}

// Destroy cleans up X11 resources
func (x *X11Backend) Destroy() error {
	if !x.initialized {
		return nil
	}

	if x.eventCallback != nil {
		x.eventCallback.OnDestroy()
	}

	if x.ximg != nil {
		C.destroyXImage(x.ximg)
		x.ximg = nil
	}

	if x.gc != nil {
		C.XFreeGC(x.display, x.gc)
		x.gc = nil
	}

	if x.window != 0 {
		C.XDestroyWindow(x.display, x.window)
		x.window = 0
	}

	if x.display != nil {
		C.XCloseDisplay(x.display)
		x.display = nil
	}

	x.initialized = false
	return nil
}

// Run starts the X11 event loop
func (x *X11Backend) Run() int {
	if !x.initialized {
		return 1
	}

	// Main event loop
	for !x.shouldClose {
		// Check for events
		if C.XPending(x.display) > 0 {
			// Process all pending events
			if !x.PollEvents() {
				break
			}
		} else {
			// No events pending, trigger idle callback
			if x.eventCallback != nil {
				x.eventCallback.OnIdle()
			}

			// Small delay to prevent high CPU usage
			C.XFlush(x.display)
			// Use usleep for a small delay (1ms)
			// time.Sleep would be better but keeping C-style for consistency
		}
	}

	return 0
}

// SetCaption sets the window caption
func (x *X11Backend) SetCaption(caption string) {
	x.caption = caption
	if x.initialized {
		cCaption := C.CString(caption)
		defer C.free(unsafe.Pointer(cCaption))
		C.XStoreName(x.display, x.window, cCaption)
		C.XFlush(x.display)
	}
}

// GetCaption returns the window caption
func (x *X11Backend) GetCaption() string {
	return x.caption
}

// SetWindowSize sets the window size
func (x *X11Backend) SetWindowSize(width, height int) error {
	if !x.initialized {
		return fmt.Errorf("X11 backend not initialized")
	}

	oldWidth, oldHeight := x.width, x.height
	x.width = width
	x.height = height

	// Resize window
	C.XResizeWindow(x.display, x.window, C.uint(width), C.uint(height))

	// Recreate image buffer
	err := x.createImageBuffer()
	if err != nil {
		return fmt.Errorf("failed to recreate image buffer: %w", err)
	}

	C.XFlush(x.display)

	// Trigger resize callback
	if x.eventCallback != nil && (width != oldWidth || height != oldHeight) {
		x.eventCallback.OnResize(width, height)
	}

	return nil
}

// GetWindowSize returns the current window size
func (x *X11Backend) GetWindowSize() (width, height int) {
	return x.width, x.height
}

// UpdateWindow updates the window display with the rendering buffer
func (x *X11Backend) UpdateWindow(buffer *buffer.RenderingBuffer[uint8]) error {
	if !x.initialized || x.ximg == nil {
		return fmt.Errorf("X11 backend not properly initialized")
	}

	// Copy buffer data to XImage, handling pixel format conversion if needed
	err := x.copyBufferToXImage(buffer)
	if err != nil {
		return fmt.Errorf("failed to copy buffer to XImage: %w", err)
	}

	// Put image to window
	C.XPutImage(x.display, x.window, x.gc, x.ximg,
		0, 0, 0, 0, C.uint(x.width), C.uint(x.height))
	C.XFlush(x.display)

	return nil
}

// SetEventCallback sets the event callback handler
func (x *X11Backend) SetEventCallback(callback types.EventCallback) {
	x.eventCallback = callback
}

// Additional methods will be implemented in separate files for better organization
// (x11_events.go, x11_keymap.go, x11_display.go)
