package x11

/*
#include <X11/keysym.h>
*/
import "C"

import (
	"agg_go/internal/platform/types"
)

// initKeymap initializes the X11 key symbol to AGG key code mapping
func (x *X11Backend) initKeymap() {
	// Initialize default mapping (for ASCII characters)
	for i := 0; i < 256; i++ {
		x.keymap[i] = types.KeyCode(i)
	}

	// Map special X11 key symbols to AGG key codes
	// This is based on the original AGG X11 implementation

	// Navigation keys
	x.setKeyMapping(C.XK_Up, types.KeyUp)
	x.setKeyMapping(C.XK_Down, types.KeyDown)
	x.setKeyMapping(C.XK_Left, types.KeyLeft)
	x.setKeyMapping(C.XK_Right, types.KeyRight)
	x.setKeyMapping(C.XK_Insert, types.KeyInsert)
	x.setKeyMapping(C.XK_Delete, types.KeyDelete)
	x.setKeyMapping(C.XK_Home, types.KeyHome)
	x.setKeyMapping(C.XK_End, types.KeyEnd)
	x.setKeyMapping(C.XK_Page_Up, types.KeyPageUp)
	x.setKeyMapping(C.XK_Page_Down, types.KeyPageDown)

	// Function keys
	x.setKeyMapping(C.XK_F1, types.KeyF1)
	x.setKeyMapping(C.XK_F2, types.KeyF2)
	x.setKeyMapping(C.XK_F3, types.KeyF3)
	x.setKeyMapping(C.XK_F4, types.KeyF4)
	x.setKeyMapping(C.XK_F5, types.KeyF5)
	x.setKeyMapping(C.XK_F6, types.KeyF6)
	x.setKeyMapping(C.XK_F7, types.KeyF7)
	x.setKeyMapping(C.XK_F8, types.KeyF8)
	x.setKeyMapping(C.XK_F9, types.KeyF9)
	x.setKeyMapping(C.XK_F10, types.KeyF10)
	x.setKeyMapping(C.XK_F11, types.KeyF11)
	x.setKeyMapping(C.XK_F12, types.KeyF12)
	x.setKeyMapping(C.XK_F13, types.KeyF13)
	x.setKeyMapping(C.XK_F14, types.KeyF14)
	x.setKeyMapping(C.XK_F15, types.KeyF15)

	// Keypad keys
	x.setKeyMapping(C.XK_KP_0, types.KeyKP0)
	x.setKeyMapping(C.XK_KP_1, types.KeyKP1)
	x.setKeyMapping(C.XK_KP_2, types.KeyKP2)
	x.setKeyMapping(C.XK_KP_3, types.KeyKP3)
	x.setKeyMapping(C.XK_KP_4, types.KeyKP4)
	x.setKeyMapping(C.XK_KP_5, types.KeyKP5)
	x.setKeyMapping(C.XK_KP_6, types.KeyKP6)
	x.setKeyMapping(C.XK_KP_7, types.KeyKP7)
	x.setKeyMapping(C.XK_KP_8, types.KeyKP8)
	x.setKeyMapping(C.XK_KP_9, types.KeyKP9)

	// Alternative keypad mappings (when NumLock is off)
	x.setKeyMapping(C.XK_KP_Insert, types.KeyKP0)
	x.setKeyMapping(C.XK_KP_End, types.KeyKP1)
	x.setKeyMapping(C.XK_KP_Down, types.KeyKP2)
	x.setKeyMapping(C.XK_KP_Page_Down, types.KeyKP3)
	x.setKeyMapping(C.XK_KP_Left, types.KeyKP4)
	x.setKeyMapping(C.XK_KP_Begin, types.KeyKP5)
	x.setKeyMapping(C.XK_KP_Right, types.KeyKP6)
	x.setKeyMapping(C.XK_KP_Home, types.KeyKP7)
	x.setKeyMapping(C.XK_KP_Up, types.KeyKP8)
	x.setKeyMapping(C.XK_KP_Page_Up, types.KeyKP9)

	// Keypad operators
	x.setKeyMapping(C.XK_KP_Delete, types.KeyKPPeriod)
	x.setKeyMapping(C.XK_KP_Decimal, types.KeyKPPeriod)
	x.setKeyMapping(C.XK_KP_Divide, types.KeyKPDivide)
	x.setKeyMapping(C.XK_KP_Multiply, types.KeyKPMultiply)
	x.setKeyMapping(C.XK_KP_Subtract, types.KeyKPMinus)
	x.setKeyMapping(C.XK_KP_Add, types.KeyKPPlus)
	x.setKeyMapping(C.XK_KP_Enter, types.KeyKPEnter)
	x.setKeyMapping(C.XK_KP_Equal, types.KeyKPEquals)

	// Lock keys
	x.setKeyMapping(C.XK_Num_Lock, types.KeyNumLock)
	x.setKeyMapping(C.XK_Caps_Lock, types.KeyCapsLock)
	x.setKeyMapping(C.XK_Scroll_Lock, types.KeyScrollLock)

	// Other special keys
	x.setKeyMapping(C.XK_Pause, types.KeyPause)
	x.setKeyMapping(C.XK_Clear, types.KeyClear)
}

// setKeyMapping sets a mapping from X11 KeySym to AGG KeyCode
func (x *X11Backend) setKeyMapping(keySym C.ulong, keyCode types.KeyCode) {
	// Use the lower 8 bits as index into keymap array
	index := int(keySym) & 0xFF
	if index < len(x.keymap) {
		x.keymap[index] = keyCode
	}
}

// x11KeyToAGG converts an X11 KeySym to AGG KeyCode
func (x *X11Backend) x11KeyToAGG(keySym C.ulong) types.KeyCode {
	// Handle ASCII range directly
	if keySym >= 32 && keySym <= 126 {
		return types.KeyCode(keySym)
	}

	// Handle special keys through keymap
	index := int(keySym) & 0xFF
	if index < len(x.keymap) {
		mappedKey := x.keymap[index]
		if mappedKey != 0 {
			return mappedKey
		}
	}

	// Handle some special cases that don't fit in 8-bit index
	switch keySym {
	case C.XK_BackSpace:
		return types.KeyBackspace
	case C.XK_Tab:
		return types.KeyTab
	case C.XK_Return:
		return types.KeyReturn
	case C.XK_Escape:
		return types.KeyEscape
	default:
		// For unknown keys, try to return the raw value if it's in ASCII range
		if keySym < 256 {
			return types.KeyCode(keySym)
		}
		// Return a safe default for unknown keys
		return types.KeyCode(0)
	}
}
