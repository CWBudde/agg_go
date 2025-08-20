// Package types contains the shared type definitions needed by both platform backends
// and the main platform package. This package breaks import cycles by providing
// a common location for type definitions that multiple packages need.
package types

import "fmt"

// WindowFlags represents window configuration flags.
type WindowFlags uint32

const (
	WindowResize          WindowFlags = 1 << 0 // Window can be resized
	WindowHWBuffer        WindowFlags = 1 << 1 // Use hardware buffer (platform dependent)
	WindowKeepAspectRatio WindowFlags = 1 << 2 // Maintain aspect ratio during resize
	WindowProcessAllKeys  WindowFlags = 1 << 3 // Process all keyboard events
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
	PixelFormatSRGB48                       // R-G-B, 16 bits per color component (sRGB)
	PixelFormatBGR48                        // B-G-R, 16 bits per color component
	PixelFormatSBGR48                       // B-G-R, 16 bits per color component (sRGB)
	PixelFormatRGBA64                       // R-G-B-A, 16 bits per color component
	PixelFormatSRGBA64                      // R-G-B-A, 16 bits per color component (sRGB)
	PixelFormatARGB64                       // A-R-G-B, 16 bits per color component
	PixelFormatSARGB64                      // A-R-G-B, 16 bits per color component (sRGB)
	PixelFormatABGR64                       // A-B-G-R, 16 bits per color component
	PixelFormatSABGR64                      // A-B-G-R, 16 bits per color component (sRGB)
	PixelFormatBGRA64                       // B-G-R-A, 16 bits per color component
	PixelFormatSBGRA64                      // B-G-R-A, 16 bits per color component (sRGB)
	PixelFormatRGB96                        // R-G-B, one 32-bit float per color component
	PixelFormatSRGB96                       // R-G-B, one 32-bit float per color component (sRGB)
	PixelFormatBGR96                        // B-G-R, one 32-bit float per color component
	PixelFormatSBGR96                       // B-G-R, one 32-bit float per color component (sRGB)
	PixelFormatRGBA128                      // R-G-B-A, one 32-bit float per color component
	PixelFormatSRGBA128                     // R-G-B-A, one 32-bit float per color component (sRGB)
	PixelFormatARGB128                      // A-R-G-B, one 32-bit float per color component
	PixelFormatSARGB128                     // A-R-G-B, one 32-bit float per color component (sRGB)
	PixelFormatABGR128                      // A-B-G-R, one 32-bit float per color component
	PixelFormatSABGR128                     // A-B-G-R, one 32-bit float per color component (sRGB)
	PixelFormatBGRA128                      // B-G-R-A, one 32-bit float per color component
	PixelFormatSBGRA128                     // B-G-R-A, one 32-bit float per color component (sRGB)
)

// String returns the string representation of the pixel format.
func (pf PixelFormat) String() string {
	switch pf {
	case PixelFormatUndefined:
		return "Undefined"
	case PixelFormatBW:
		return "BW"
	case PixelFormatGray8:
		return "Gray8"
	case PixelFormatSGray8:
		return "SGray8"
	case PixelFormatGray16:
		return "Gray16"
	case PixelFormatGray32:
		return "Gray32"
	case PixelFormatRGB555:
		return "RGB555"
	case PixelFormatRGB565:
		return "RGB565"
	case PixelFormatRGB24:
		return "RGB24"
	case PixelFormatSRGB24:
		return "SRGB24"
	case PixelFormatBGR24:
		return "BGR24"
	case PixelFormatSBGR24:
		return "SBGR24"
	case PixelFormatRGBA32:
		return "RGBA32"
	case PixelFormatSRGBA32:
		return "SRGBA32"
	case PixelFormatARGB32:
		return "ARGB32"
	case PixelFormatSARGB32:
		return "SARGB32"
	case PixelFormatABGR32:
		return "ABGR32"
	case PixelFormatSABGR32:
		return "SABGR32"
	case PixelFormatBGRA32:
		return "BGRA32"
	case PixelFormatSBGRA32:
		return "SBGRA32"
	case PixelFormatRGB48:
		return "RGB48"
	case PixelFormatSRGB48:
		return "SRGB48"
	case PixelFormatBGR48:
		return "BGR48"
	case PixelFormatSBGR48:
		return "SBGR48"
	case PixelFormatRGBA64:
		return "RGBA64"
	case PixelFormatSRGBA64:
		return "SRGBA64"
	case PixelFormatARGB64:
		return "ARGB64"
	case PixelFormatSARGB64:
		return "SARGB64"
	case PixelFormatABGR64:
		return "ABGR64"
	case PixelFormatSABGR64:
		return "SABGR64"
	case PixelFormatBGRA64:
		return "BGRA64"
	case PixelFormatSBGRA64:
		return "SBGRA64"
	case PixelFormatRGB96:
		return "RGB96"
	case PixelFormatSRGB96:
		return "SRGB96"
	case PixelFormatBGR96:
		return "BGR96"
	case PixelFormatSBGR96:
		return "SBGR96"
	case PixelFormatRGBA128:
		return "RGBA128"
	case PixelFormatSRGBA128:
		return "SRGBA128"
	case PixelFormatARGB128:
		return "ARGB128"
	case PixelFormatSARGB128:
		return "SARGB128"
	case PixelFormatABGR128:
		return "ABGR128"
	case PixelFormatSABGR128:
		return "SABGR128"
	case PixelFormatBGRA128:
		return "BGRA128"
	case PixelFormatSBGRA128:
		return "SBGRA128"
	default:
		return "Unknown"
	}
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
	case PixelFormatRGB48, PixelFormatSRGB48, PixelFormatBGR48, PixelFormatSBGR48:
		return 48
	case PixelFormatRGBA64, PixelFormatSRGBA64, PixelFormatARGB64, PixelFormatSARGB64,
		PixelFormatABGR64, PixelFormatSABGR64, PixelFormatBGRA64, PixelFormatSBGRA64:
		return 64
	case PixelFormatRGB96, PixelFormatSRGB96, PixelFormatBGR96, PixelFormatSBGR96:
		return 96
	case PixelFormatRGBA128, PixelFormatSRGBA128, PixelFormatARGB128, PixelFormatSARGB128,
		PixelFormatABGR128, PixelFormatSABGR128, PixelFormatBGRA128, PixelFormatSBGRA128:
		return 128
	default:
		return 0
	}
}

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
	KeyKP0        KeyCode = 256
	KeyKP1        KeyCode = 257
	KeyKP2        KeyCode = 258
	KeyKP3        KeyCode = 259
	KeyKP4        KeyCode = 260
	KeyKP5        KeyCode = 261
	KeyKP6        KeyCode = 262
	KeyKP7        KeyCode = 263
	KeyKP8        KeyCode = 264
	KeyKP9        KeyCode = 265
	KeyKPPeriod   KeyCode = 266
	KeyKPDivide   KeyCode = 267
	KeyKPMultiply KeyCode = 268
	KeyKPMinus    KeyCode = 269
	KeyKPPlus     KeyCode = 270
	KeyKPEnter    KeyCode = 271
	KeyKPEquals   KeyCode = 272

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

	// Modifier keys
	KeyNumLock    KeyCode = 300
	KeyCapsLock   KeyCode = 301
	KeyScrollLock KeyCode = 302
	KeyRShift     KeyCode = 303
	KeyLShift     KeyCode = 304
	KeyRCtrl      KeyCode = 305
	KeyLCtrl      KeyCode = 306
	KeyRAlt       KeyCode = 307
	KeyLAlt       KeyCode = 308
	KeyRMeta      KeyCode = 309
	KeyLMeta      KeyCode = 310
	KeyLSuper     KeyCode = 311 // Left "Windows" key
	KeyRSuper     KeyCode = 312 // Right "Windows" key
	KeyMode       KeyCode = 313 // AltGr key
	KeyCompose    KeyCode = 314 // Multi-key compose key
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
		return "KPPeriod"
	case KeyKPDivide:
		return "KPDivide"
	case KeyKPMultiply:
		return "KPMultiply"
	case KeyKPMinus:
		return "KP-"
	case KeyKPPlus:
		return "KP+"
	case KeyKPEnter:
		return "KPEnter"
	case KeyKPEquals:
		return "KPEquals"
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
	case KeyRShift:
		return "RShift"
	case KeyLShift:
		return "LShift"
	case KeyRCtrl:
		return "RCtrl"
	case KeyLCtrl:
		return "LCtrl"
	case KeyRAlt:
		return "RAlt"
	case KeyLAlt:
		return "LAlt"
	case KeyRMeta:
		return "RMeta"
	case KeyLMeta:
		return "LMeta"
	case KeyLSuper:
		return "LSuper"
	case KeyRSuper:
		return "RSuper"
	case KeyMode:
		return "Mode"
	case KeyCompose:
		return "Compose"
	default:
		if k >= 32 && k < 127 {
			return fmt.Sprintf("'%c'", k)
		}
		return fmt.Sprintf("Unknown(%d)", k)
	}
}

// IsASCII returns true if the key code represents a printable ASCII character.
func (k KeyCode) IsASCII() bool {
	return k >= 32 && k < 127
}

// IsFunctionKey returns true if the key code is a function key (F1-F12).
func (k KeyCode) IsFunctionKey() bool {
	return k >= KeyF1 && k <= KeyF12
}

// IsArrowKey returns true if the key code is an arrow key (Up, Down, Left, Right).
func (k KeyCode) IsArrowKey() bool {
	return k == KeyUp || k == KeyDown || k == KeyLeft || k == KeyRight
}

// IsKeypadKey returns true if the key code is a keypad key.
func (k KeyCode) IsKeypadKey() bool {
	return k >= KeyKP0 && k <= KeyKPEquals
}

// IsNavigationKey returns true if the key code is a navigation key (Home, End, PageUp, PageDown, Insert, or an arrow key).
func (k KeyCode) IsNavigationKey() bool {
	return k.IsArrowKey() || k == KeyInsert || k == KeyHome || k == KeyEnd || k == KeyPageUp || k == KeyPageDown
}

// EventCallback defines the interface for handling platform events.
// Applications should implement this interface to handle user input and system events.
type EventCallback interface {
	OnInit()
	OnDestroy()
	OnResize(width, height int)
	OnIdle()
	OnMouseMove(x, y int, flags InputFlags)
	OnMouseButtonDown(x, y int, flags InputFlags)
	OnMouseButtonUp(x, y int, flags InputFlags)
	OnKey(x, y int, key KeyCode, flags InputFlags)
	OnDraw()
}
