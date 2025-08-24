package platform

import "agg_go/internal/platform/types"

// MouseEvent represents a mouse event.
type MouseEvent struct {
	X     int              // Mouse X coordinate
	Y     int              // Mouse Y coordinate
	Flags types.InputFlags // Input flags at the time of the event
}

// KeyboardEvent represents a keyboard event.
type KeyboardEvent struct {
	X     int              // Mouse X coordinate at the time of key event
	Y     int              // Mouse Y coordinate at the time of key event
	Key   types.KeyCode    // The key that was pressed/released
	Flags types.InputFlags // Input flags at the time of the event
}

// ResizeEvent represents a window resize event.
type ResizeEvent struct {
	Width  int // New window width
	Height int // New window height
}

// EventHandler is an interface for handling platform events.
type EventHandler interface {
	OnInit()
	OnResize(width, height int)
	OnIdle()
	OnMouseMove(x, y int, flags types.InputFlags)
	OnMouseButtonDown(x, y int, flags types.InputFlags)
	OnMouseButtonUp(x, y int, flags types.InputFlags)
	OnKey(x, y int, key types.KeyCode, flags types.InputFlags)
	OnCtrlChange()
	OnDraw()
	OnPostDraw(rawHandler RawEventHandler)
}

// BaseEventHandler provides default implementations for all event handler methods.
// Applications can embed this struct and override only the methods they need.
type BaseEventHandler struct{}

// OnInit is called when the application is initialized.
func (h *BaseEventHandler) OnInit() {}

// OnResize is called when the window is resized.
func (h *BaseEventHandler) OnResize(width, height int) {}

// OnIdle is called when the application is idle (no events to process).
func (h *BaseEventHandler) OnIdle() {}

// OnMouseMove is called when the mouse is moved.
func (h *BaseEventHandler) OnMouseMove(x, y int, flags types.InputFlags) {}

// OnMouseButtonDown is called when a mouse button is pressed.
func (h *BaseEventHandler) OnMouseButtonDown(x, y int, flags types.InputFlags) {}

// OnMouseButtonUp is called when a mouse button is released.
func (h *BaseEventHandler) OnMouseButtonUp(x, y int, flags types.InputFlags) {}

// OnKey is called when a key is pressed or released.
func (h *BaseEventHandler) OnKey(x, y int, key types.KeyCode, flags types.InputFlags) {}

// OnCtrlChange is called when a control's state changes.
func (h *BaseEventHandler) OnCtrlChange() {}

// OnDraw is called when the window needs to be redrawn.
func (h *BaseEventHandler) OnDraw() {}

// OnPostDraw is called after drawing is complete.
func (h *BaseEventHandler) OnPostDraw(rawHandler RawEventHandler) {}
