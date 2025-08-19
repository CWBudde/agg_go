package x11

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/keysym.h>
*/
import "C"

import (
	"unsafe"

	"agg_go/internal/platform/types"
)

// PollEvents polls for X11 events and processes them
func (x *X11Backend) PollEvents() bool {
	if !x.initialized {
		return false
	}

	var event C.XEvent

	// Process all pending events
	for C.XPending(x.display) > 0 {
		C.XNextEvent(x.display, &event)
		x.handleEvent(&event)
	}

	return !x.shouldClose
}

// WaitEvent waits for an X11 event and processes it
func (x *X11Backend) WaitEvent() bool {
	if !x.initialized {
		return false
	}

	var event C.XEvent
	C.XNextEvent(x.display, &event)
	x.handleEvent(&event)

	return !x.shouldClose
}

// ForceRedraw forces a window redraw
func (x *X11Backend) ForceRedraw() {
	if x.eventCallback != nil {
		x.eventCallback.OnDraw()
	}
}

// handleEvent processes individual X11 events
func (x *X11Backend) handleEvent(event *C.XEvent) {
	switch (*C.XAnyEvent)(unsafe.Pointer(event))._type {
	case C.Expose:
		x.handleExposeEvent(event)
	case C.ConfigureNotify:
		x.handleConfigureEvent(event)
	case C.KeyPress:
		x.handleKeyPressEvent(event)
	case C.KeyRelease:
		x.handleKeyReleaseEvent(event)
	case C.ButtonPress:
		x.handleButtonPressEvent(event)
	case C.ButtonRelease:
		x.handleButtonReleaseEvent(event)
	case C.MotionNotify:
		x.handleMotionEvent(event)
	case C.ClientMessage:
		x.handleClientMessageEvent(event)
	}
}

// handleExposeEvent handles window expose events
func (x *X11Backend) handleExposeEvent(event *C.XEvent) {
	if x.eventCallback != nil {
		x.eventCallback.OnDraw()
	}
}

// handleConfigureEvent handles window configuration changes (resize, move, etc.)
func (x *X11Backend) handleConfigureEvent(event *C.XEvent) {
	configEvent := (*C.XConfigureEvent)(unsafe.Pointer(event))

	newWidth := int(configEvent.width)
	newHeight := int(configEvent.height)

	if newWidth != x.width || newHeight != x.height {
		x.width = newWidth
		x.height = newHeight

		// Recreate image buffer for new size
		x.createImageBuffer()

		if x.eventCallback != nil {
			x.eventCallback.OnResize(newWidth, newHeight)
		}
	}
}

// handleKeyPressEvent handles key press events
func (x *X11Backend) handleKeyPressEvent(event *C.XEvent) {
	keyEvent := (*C.XKeyEvent)(unsafe.Pointer(event))

	// Get key symbol
	keySym := C.XLookupKeysym(keyEvent, 0)

	// Convert to AGG key code
	keyCode := x.x11KeyToAGG(keySym)

	// Get input flags
	flags := x.getInputFlags(keyEvent.state)

	if x.eventCallback != nil {
		x.eventCallback.OnKey(int(keyEvent.x), int(keyEvent.y), keyCode, flags)
	}
}

// handleKeyReleaseEvent handles key release events
func (x *X11Backend) handleKeyReleaseEvent(event *C.XEvent) {
	// For now, we treat key release the same as key press
	// A more sophisticated implementation might track key states
	x.handleKeyPressEvent(event)
}

// handleButtonPressEvent handles mouse button press events
func (x *X11Backend) handleButtonPressEvent(event *C.XEvent) {
	buttonEvent := (*C.XButtonEvent)(unsafe.Pointer(event))

	// Convert X11 button to input flags
	flags := x.getInputFlags(buttonEvent.state)

	// Add the pressed button to flags
	switch buttonEvent.button {
	case C.Button1: // Left button
		flags |= types.MouseLeft
	case C.Button3: // Right button
		flags |= types.MouseRight
	}

	if x.eventCallback != nil {
		x.eventCallback.OnMouseButtonDown(int(buttonEvent.x), int(buttonEvent.y), flags)
	}
}

// handleButtonReleaseEvent handles mouse button release events
func (x *X11Backend) handleButtonReleaseEvent(event *C.XEvent) {
	buttonEvent := (*C.XButtonEvent)(unsafe.Pointer(event))

	// Convert X11 button to input flags (without the released button)
	flags := x.getInputFlags(buttonEvent.state)

	// The released button is NOT included in flags for button up events
	if x.eventCallback != nil {
		x.eventCallback.OnMouseButtonUp(int(buttonEvent.x), int(buttonEvent.y), flags)
	}
}

// handleMotionEvent handles mouse motion events
func (x *X11Backend) handleMotionEvent(event *C.XEvent) {
	motionEvent := (*C.XMotionEvent)(unsafe.Pointer(event))

	flags := x.getInputFlags(motionEvent.state)

	if x.eventCallback != nil {
		x.eventCallback.OnMouseMove(int(motionEvent.x), int(motionEvent.y), flags)
	}
}

// handleClientMessageEvent handles client messages (like window close)
func (x *X11Backend) handleClientMessageEvent(event *C.XEvent) {
	clientEvent := (*C.XClientMessageEvent)(unsafe.Pointer(event))

	// Access the first long value from the data union
	dataPtr := (*C.long)(unsafe.Pointer(&clientEvent.data[0]))
	if C.Atom(*dataPtr) == x.wmDeleteAtom {
		x.shouldClose = true
	}
}

// getInputFlags converts X11 modifier state to AGG input flags
func (x *X11Backend) getInputFlags(state C.uint) types.InputFlags {
	var flags types.InputFlags

	if state&C.Button1Mask != 0 {
		flags |= types.MouseLeft
	}
	if state&C.Button3Mask != 0 {
		flags |= types.MouseRight
	}
	if state&C.ShiftMask != 0 {
		flags |= types.KbdShift
	}
	if state&C.ControlMask != 0 {
		flags |= types.KbdCtrl
	}

	return flags
}
