package platform

import "fmt"

// InputFlags represents mouse and keyboard flags for event handling.
// These flags indicate the state of input devices at the time an event occurs.
type InputFlags uint32

const (
	MouseLeft  InputFlags = 1 << 0 // Left mouse button is pressed
	MouseRight InputFlags = 1 << 1 // Right mouse button is pressed
	KbdShift   InputFlags = 1 << 2 // Shift key is pressed
	KbdCtrl    InputFlags = 1 << 3 // Ctrl key is pressed
)

// String returns the string representation of the input flags.
func (f InputFlags) String() string {
	var flags []string
	if f&MouseLeft != 0 {
		flags = append(flags, "MouseLeft")
	}
	if f&MouseRight != 0 {
		flags = append(flags, "MouseRight")
	}
	if f&KbdShift != 0 {
		flags = append(flags, "KbdShift")
	}
	if f&KbdCtrl != 0 {
		flags = append(flags, "KbdCtrl")
	}
	if len(flags) == 0 {
		return "None"
	}
	result := ""
	for i, flag := range flags {
		if i > 0 {
			result += "|"
		}
		result += flag
	}
	return result
}

// HasMouseLeft returns true if the left mouse button flag is set.
func (f InputFlags) HasMouseLeft() bool {
	return f&MouseLeft != 0
}

// HasMouseRight returns true if the right mouse button flag is set.
func (f InputFlags) HasMouseRight() bool {
	return f&MouseRight != 0
}

// HasShift returns true if the shift key flag is set.
func (f InputFlags) HasShift() bool {
	return f&KbdShift != 0
}

// HasCtrl returns true if the ctrl key flag is set.
func (f InputFlags) HasCtrl() bool {
	return f&KbdCtrl != 0
}

// KeyCode represents keyboard key codes.
// These are platform-independent key codes that should be mapped from
// platform-specific key codes.
type KeyCode int

const (
	// ASCII set - should be supported everywhere
	KeyBackspace KeyCode = 8
	KeyTab       KeyCode = 9
	KeyClear     KeyCode = 12
	KeyReturn    KeyCode = 13
	KeyPause     KeyCode = 19
	KeyEscape    KeyCode = 27

	// Extended keys
	KeyDelete KeyCode = 127

	// Keypad
	KeyKP0       KeyCode = 256
	KeyKP1       KeyCode = 257
	KeyKP2       KeyCode = 258
	KeyKP3       KeyCode = 259
	KeyKP4       KeyCode = 260
	KeyKP5       KeyCode = 261
	KeyKP6       KeyCode = 262
	KeyKP7       KeyCode = 263
	KeyKP8       KeyCode = 264
	KeyKP9       KeyCode = 265
	KeyKPPeriod  KeyCode = 266
	KeyKPDivide  KeyCode = 267
	KeyKPMultiply KeyCode = 268
	KeyKPMinus   KeyCode = 269
	KeyKPPlus    KeyCode = 270
	KeyKPEnter   KeyCode = 271
	KeyKPEquals  KeyCode = 272

	// Arrow keys and navigation
	KeyUp       KeyCode = 273
	KeyDown     KeyCode = 274
	KeyRight    KeyCode = 275
	KeyLeft     KeyCode = 276
	KeyInsert   KeyCode = 277
	KeyHome     KeyCode = 278
	KeyEnd      KeyCode = 279
	KeyPageUp   KeyCode = 280
	KeyPageDown KeyCode = 281

	// Function keys
	KeyF1  KeyCode = 282
	KeyF2  KeyCode = 283
	KeyF3  KeyCode = 284
	KeyF4  KeyCode = 285
	KeyF5  KeyCode = 286
	KeyF6  KeyCode = 287
	KeyF7  KeyCode = 288
	KeyF8  KeyCode = 289
	KeyF9  KeyCode = 290
	KeyF10 KeyCode = 291
	KeyF11 KeyCode = 292
	KeyF12 KeyCode = 293
	KeyF13 KeyCode = 294
	KeyF14 KeyCode = 295
	KeyF15 KeyCode = 296

	// Lock keys (platform dependent)
	KeyNumLock    KeyCode = 300
	KeyCapsLock   KeyCode = 301
	KeyScrollLock KeyCode = 302
)

// String returns the string representation of the key code.
func (k KeyCode) String() string {
	switch k {
	case KeyBackspace:
		return "Backspace"
	case KeyTab:
		return "Tab"
	case KeyClear:
		return "Clear"
	case KeyReturn:
		return "Return"
	case KeyPause:
		return "Pause"
	case KeyEscape:
		return "Escape"
	case KeyDelete:
		return "Delete"
	case KeyKP0:
		return "KP0"
	case KeyKP1:
		return "KP1"
	case KeyKP2:
		return "KP2"
	case KeyKP3:
		return "KP3"
	case KeyKP4:
		return "KP4"
	case KeyKP5:
		return "KP5"
	case KeyKP6:
		return "KP6"
	case KeyKP7:
		return "KP7"
	case KeyKP8:
		return "KP8"
	case KeyKP9:
		return "KP9"
	case KeyKPPeriod:
		return "KP."
	case KeyKPDivide:
		return "KP/"
	case KeyKPMultiply:
		return "KP*"
	case KeyKPMinus:
		return "KP-"
	case KeyKPPlus:
		return "KP+"
	case KeyKPEnter:
		return "KPEnter"
	case KeyKPEquals:
		return "KP="
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyRight:
		return "Right"
	case KeyLeft:
		return "Left"
	case KeyInsert:
		return "Insert"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	case KeyF13:
		return "F13"
	case KeyF14:
		return "F14"
	case KeyF15:
		return "F15"
	case KeyNumLock:
		return "NumLock"
	case KeyCapsLock:
		return "CapsLock"
	case KeyScrollLock:
		return "ScrollLock"
	default:
		// For ASCII printable characters
		if k >= 32 && k <= 126 {
			return fmt.Sprintf("'%c'", rune(k))
		}
		return fmt.Sprintf("Unknown(%d)", int(k))
	}
}

// IsASCII returns true if the key code represents a printable ASCII character.
func (k KeyCode) IsASCII() bool {
	return k >= 32 && k <= 126
}

// IsFunctionKey returns true if the key code represents a function key (F1-F15).
func (k KeyCode) IsFunctionKey() bool {
	return k >= KeyF1 && k <= KeyF15
}

// IsArrowKey returns true if the key code represents an arrow key.
func (k KeyCode) IsArrowKey() bool {
	return k >= KeyUp && k <= KeyLeft
}

// IsKeypadKey returns true if the key code represents a keypad key.
func (k KeyCode) IsKeypadKey() bool {
	return k >= KeyKP0 && k <= KeyKPEquals
}

// IsNavigationKey returns true if the key code represents a navigation key.
func (k KeyCode) IsNavigationKey() bool {
	return (k >= KeyInsert && k <= KeyPageDown) || k.IsArrowKey()
}

// MouseEvent represents a mouse event.
type MouseEvent struct {
	X     int        // Mouse X coordinate
	Y     int        // Mouse Y coordinate
	Flags InputFlags // Input flags at the time of the event
}

// KeyboardEvent represents a keyboard event.
type KeyboardEvent struct {
	X     int        // Mouse X coordinate at the time of key event
	Y     int        // Mouse Y coordinate at the time of key event
	Key   KeyCode    // The key that was pressed/released
	Flags InputFlags // Input flags at the time of the event
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
	OnMouseMove(x, y int, flags InputFlags)
	OnMouseButtonDown(x, y int, flags InputFlags)
	OnMouseButtonUp(x, y int, flags InputFlags)
	OnKey(x, y int, key KeyCode, flags InputFlags)
	OnCtrlChange()
	OnDraw()
	OnPostDraw(rawHandler interface{})
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
func (h *BaseEventHandler) OnMouseMove(x, y int, flags InputFlags) {}

// OnMouseButtonDown is called when a mouse button is pressed.
func (h *BaseEventHandler) OnMouseButtonDown(x, y int, flags InputFlags) {}

// OnMouseButtonUp is called when a mouse button is released.
func (h *BaseEventHandler) OnMouseButtonUp(x, y int, flags InputFlags) {}

// OnKey is called when a key is pressed or released.
func (h *BaseEventHandler) OnKey(x, y int, key KeyCode, flags InputFlags) {}

// OnCtrlChange is called when a control's state changes.
func (h *BaseEventHandler) OnCtrlChange() {}

// OnDraw is called when the window needs to be redrawn.
func (h *BaseEventHandler) OnDraw() {}

// OnPostDraw is called after drawing is complete.
func (h *BaseEventHandler) OnPostDraw(rawHandler interface{}) {}