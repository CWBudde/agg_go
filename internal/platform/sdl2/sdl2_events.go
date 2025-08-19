package sdl2

import (
	"agg_go/internal/platform/types"
	"github.com/veandco/go-sdl2/sdl"
)

// PollEvents polls for SDL2 events and processes them
func (s *SDL2Backend) PollEvents() bool {
	if !s.initialized {
		return false
	}

	for {
		event := sdl.PollEvent()
		if event == nil {
			break
		}
		s.handleEvent(event)
	}

	return !s.shouldClose
}

// WaitEvent waits for an SDL2 event and processes it
func (s *SDL2Backend) WaitEvent() bool {
	if !s.initialized {
		return false
	}

	event := sdl.WaitEvent()
	if event != nil {
		s.handleEvent(event)
	}

	return !s.shouldClose
}

// ForceRedraw forces a window redraw
func (s *SDL2Backend) ForceRedraw() {
	if s.eventCallback != nil {
		s.eventCallback.OnDraw()
	}
}

// handleEvent processes individual SDL2 events
func (s *SDL2Backend) handleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.QuitEvent:
		s.handleQuitEvent(e)
	case *sdl.WindowEvent:
		s.handleWindowEvent(e)
	case *sdl.KeyboardEvent:
		s.handleKeyboardEvent(e)
	case *sdl.MouseButtonEvent:
		s.handleMouseButtonEvent(e)
	case *sdl.MouseMotionEvent:
		s.handleMouseMotionEvent(e)
	}
}

// handleQuitEvent handles application quit events
func (s *SDL2Backend) handleQuitEvent(event *sdl.QuitEvent) {
	s.shouldClose = true
}

// handleWindowEvent handles window events (resize, expose, etc.)
func (s *SDL2Backend) handleWindowEvent(event *sdl.WindowEvent) {
	switch event.Event {
	case sdl.WINDOWEVENT_EXPOSED:
		if s.eventCallback != nil {
			s.eventCallback.OnDraw()
		}
	case sdl.WINDOWEVENT_RESIZED:
		newWidth := int(event.Data1)
		newHeight := int(event.Data2)

		if newWidth != s.width || newHeight != s.height {
			s.width = newWidth
			s.height = newHeight

			// Recreate texture and surface for new size
			if s.texture != nil {
				s.texture.Destroy()
			}
			if s.surface != nil {
				s.surface.Free()
			}

			// Recreate with new dimensions
			var err error
			s.texture, err = s.renderer.CreateTexture(
				s.pixelFormat,
				sdl.TEXTUREACCESS_STREAMING,
				int32(newWidth), int32(newHeight))
			if err == nil {
				s.surface, err = sdl.CreateRGBSurface(
					0, int32(newWidth), int32(newHeight), int32(s.bpp),
					s.rmask, s.gmask, s.bmask, s.amask)
			}

			if s.eventCallback != nil && err == nil {
				s.eventCallback.OnResize(newWidth, newHeight)
			}
		}
	}
}

// handleKeyboardEvent handles keyboard events
func (s *SDL2Backend) handleKeyboardEvent(event *sdl.KeyboardEvent) {
	if event.Type != sdl.KEYDOWN {
		return // Only handle key down events for now
	}

	// Get mouse position for key events
	mouseX, mouseY, _ := sdl.GetMouseState()

	// Convert SDL2 key to AGG key code
	keyCode := s.sdlKeyToAGG(event.Keysym.Sym)

	// Get input flags
	flags := s.getInputFlags()

	// Add keyboard modifiers
	if event.Keysym.Mod&sdl.KMOD_SHIFT != 0 {
		flags |= types.KbdShift
	}
	if event.Keysym.Mod&sdl.KMOD_CTRL != 0 {
		flags |= types.KbdCtrl
	}

	if s.eventCallback != nil {
		s.eventCallback.OnKey(int(mouseX), int(mouseY), keyCode, flags)
	}
}

// handleMouseButtonEvent handles mouse button events
func (s *SDL2Backend) handleMouseButtonEvent(event *sdl.MouseButtonEvent) {
	flags := s.getInputFlags()

	// Add the button that was pressed/released
	switch event.Button {
	case sdl.BUTTON_LEFT:
		if event.Type == sdl.MOUSEBUTTONDOWN {
			flags |= types.MouseLeft
		}
	case sdl.BUTTON_RIGHT:
		if event.Type == sdl.MOUSEBUTTONDOWN {
			flags |= types.MouseRight
		}
	}

	if s.eventCallback != nil {
		if event.Type == sdl.MOUSEBUTTONDOWN {
			s.eventCallback.OnMouseButtonDown(int(event.X), int(event.Y), flags)
		} else {
			s.eventCallback.OnMouseButtonUp(int(event.X), int(event.Y), flags)
		}
	}
}

// handleMouseMotionEvent handles mouse motion events
func (s *SDL2Backend) handleMouseMotionEvent(event *sdl.MouseMotionEvent) {
	flags := s.getInputFlags()

	if s.eventCallback != nil {
		s.eventCallback.OnMouseMove(int(event.X), int(event.Y), flags)
	}
}

// getInputFlags gets the current input state flags
func (s *SDL2Backend) getInputFlags() types.InputFlags {
	var flags types.InputFlags

	// Get mouse state
	_, _, mouseState := sdl.GetMouseState()
	if mouseState&sdl.ButtonLMask() != 0 {
		flags |= types.MouseLeft
	}
	if mouseState&sdl.ButtonRMask() != 0 {
		flags |= types.MouseRight
	}

	// Get keyboard state
	keyState := sdl.GetKeyboardState()
	if keyState[sdl.SCANCODE_LSHIFT] != 0 || keyState[sdl.SCANCODE_RSHIFT] != 0 {
		flags |= types.KbdShift
	}
	if keyState[sdl.SCANCODE_LCTRL] != 0 || keyState[sdl.SCANCODE_RCTRL] != 0 {
		flags |= types.KbdCtrl
	}

	return flags
}

// sdlKeyToAGG converts an SDL2 key code to AGG key code
func (s *SDL2Backend) sdlKeyToAGG(key sdl.Keycode) types.KeyCode {
	// SDL2 uses the same key codes as AGG for many keys, which is convenient
	// since AGG's key codes were originally based on SDL

	// Handle ASCII range directly
	if key >= 32 && key <= 126 {
		return types.KeyCode(key)
	}

	// Handle special keys
	switch key {
	case sdl.K_BACKSPACE:
		return types.KeyBackspace
	case sdl.K_TAB:
		return types.KeyTab
	case sdl.K_CLEAR:
		return types.KeyClear
	case sdl.K_RETURN:
		return types.KeyReturn
	case sdl.K_PAUSE:
		return types.KeyPause
	case sdl.K_ESCAPE:
		return types.KeyEscape
	case sdl.K_DELETE:
		return types.KeyDelete

	// Keypad keys
	case sdl.K_KP_0:
		return types.KeyKP0
	case sdl.K_KP_1:
		return types.KeyKP1
	case sdl.K_KP_2:
		return types.KeyKP2
	case sdl.K_KP_3:
		return types.KeyKP3
	case sdl.K_KP_4:
		return types.KeyKP4
	case sdl.K_KP_5:
		return types.KeyKP5
	case sdl.K_KP_6:
		return types.KeyKP6
	case sdl.K_KP_7:
		return types.KeyKP7
	case sdl.K_KP_8:
		return types.KeyKP8
	case sdl.K_KP_9:
		return types.KeyKP9
	case sdl.K_KP_PERIOD:
		return types.KeyKPPeriod
	case sdl.K_KP_DIVIDE:
		return types.KeyKPDivide
	case sdl.K_KP_MULTIPLY:
		return types.KeyKPMultiply
	case sdl.K_KP_MINUS:
		return types.KeyKPMinus
	case sdl.K_KP_PLUS:
		return types.KeyKPPlus
	case sdl.K_KP_ENTER:
		return types.KeyKPEnter
	case sdl.K_KP_EQUALS:
		return types.KeyKPEquals

	// Arrow keys
	case sdl.K_UP:
		return types.KeyUp
	case sdl.K_DOWN:
		return types.KeyDown
	case sdl.K_RIGHT:
		return types.KeyRight
	case sdl.K_LEFT:
		return types.KeyLeft
	case sdl.K_INSERT:
		return types.KeyInsert
	case sdl.K_HOME:
		return types.KeyHome
	case sdl.K_END:
		return types.KeyEnd
	case sdl.K_PAGEUP:
		return types.KeyPageUp
	case sdl.K_PAGEDOWN:
		return types.KeyPageDown

	// Function keys
	case sdl.K_F1:
		return types.KeyF1
	case sdl.K_F2:
		return types.KeyF2
	case sdl.K_F3:
		return types.KeyF3
	case sdl.K_F4:
		return types.KeyF4
	case sdl.K_F5:
		return types.KeyF5
	case sdl.K_F6:
		return types.KeyF6
	case sdl.K_F7:
		return types.KeyF7
	case sdl.K_F8:
		return types.KeyF8
	case sdl.K_F9:
		return types.KeyF9
	case sdl.K_F10:
		return types.KeyF10
	case sdl.K_F11:
		return types.KeyF11
	case sdl.K_F12:
		return types.KeyF12
	case sdl.K_F13:
		return types.KeyF13
	case sdl.K_F14:
		return types.KeyF14
	case sdl.K_F15:
		return types.KeyF15

	// Lock keys
	case sdl.K_NUMLOCKCLEAR:
		return types.KeyNumLock
	case sdl.K_CAPSLOCK:
		return types.KeyCapsLock
	case sdl.K_SCROLLLOCK:
		return types.KeyScrollLock

	default:
		// Return the raw key code for unknown keys
		return types.KeyCode(key)
	}
}
