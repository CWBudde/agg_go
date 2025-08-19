package platform

import (
	"strings"
	"testing"
)

func TestInputFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    InputFlags
		expected string
	}{
		{"No flags", 0, "None"},
		{"Mouse left", MouseLeft, "MouseLeft"},
		{"Mouse right", MouseRight, "MouseRight"},
		{"Keyboard shift", KbdShift, "KbdShift"},
		{"Keyboard ctrl", KbdCtrl, "KbdCtrl"},
		{"Multiple flags", MouseLeft | KbdShift, "MouseLeft|KbdShift"},
		{"All flags", MouseLeft | MouseRight | KbdShift | KbdCtrl, "MouseLeft|MouseRight|KbdShift|KbdCtrl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flags.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestInputFlagsHelpers(t *testing.T) {
	flags := MouseLeft | KbdShift

	if !flags.HasMouseLeft() {
		t.Error("Expected HasMouseLeft() to be true")
	}

	if flags.HasMouseRight() {
		t.Error("Expected HasMouseRight() to be false")
	}

	if !flags.HasShift() {
		t.Error("Expected HasShift() to be true")
	}

	if flags.HasCtrl() {
		t.Error("Expected HasCtrl() to be false")
	}
}

func TestKeyCodeString(t *testing.T) {
	tests := []struct {
		key      KeyCode
		expected string
	}{
		{KeyBackspace, "Backspace"},
		{KeyTab, "Tab"},
		{KeyReturn, "Return"},
		{KeyEscape, "Escape"},
		{KeyDelete, "Delete"},
		{KeyKP0, "KP0"},
		{KeyKP9, "KP9"},
		{KeyKPPlus, "KP+"},
		{KeyKPMinus, "KP-"},
		{KeyUp, "Up"},
		{KeyDown, "Down"},
		{KeyLeft, "Left"},
		{KeyRight, "Right"},
		{KeyF1, "F1"},
		{KeyF12, "F12"},
		{KeyNumLock, "NumLock"},
		{KeyCapsLock, "CapsLock"},
		{KeyCode(65), "'A'"}, // ASCII 'A'
		{KeyCode(32), "' '"}, // ASCII space
		{KeyCode(999), "Unknown(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.key.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestKeyCodeHelpers(t *testing.T) {
	tests := []struct {
		name             string
		key              KeyCode
		expectASCII      bool
		expectFunction   bool
		expectArrow      bool
		expectKeypad     bool
		expectNavigation bool
	}{
		{"ASCII letter", KeyCode(65), true, false, false, false, false}, // 'A'
		{"ASCII space", KeyCode(32), true, false, false, false, false},  // ' '
		{"ASCII tilde", KeyCode(126), true, false, false, false, false}, // '~'
		{"Control char", KeyCode(1), false, false, false, false, false}, // Control character
		{"Backspace", KeyBackspace, false, false, false, false, false},
		{"Function key F1", KeyF1, false, true, false, false, false},
		{"Function key F12", KeyF12, false, true, false, false, false},
		{"Arrow up", KeyUp, false, false, true, false, true},
		{"Arrow down", KeyDown, false, false, true, false, true},
		{"Keypad 0", KeyKP0, false, false, false, true, false},
		{"Keypad plus", KeyKPPlus, false, false, false, true, false},
		{"Home", KeyHome, false, false, false, false, true},
		{"End", KeyEnd, false, false, false, false, true},
		{"Page up", KeyPageUp, false, false, false, false, true},
		{"Delete", KeyDelete, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key.IsASCII() != tt.expectASCII {
				t.Errorf("IsASCII(): expected %v, got %v", tt.expectASCII, tt.key.IsASCII())
			}

			if tt.key.IsFunctionKey() != tt.expectFunction {
				t.Errorf("IsFunctionKey(): expected %v, got %v", tt.expectFunction, tt.key.IsFunctionKey())
			}

			if tt.key.IsArrowKey() != tt.expectArrow {
				t.Errorf("IsArrowKey(): expected %v, got %v", tt.expectArrow, tt.key.IsArrowKey())
			}

			if tt.key.IsKeypadKey() != tt.expectKeypad {
				t.Errorf("IsKeypadKey(): expected %v, got %v", tt.expectKeypad, tt.key.IsKeypadKey())
			}

			if tt.key.IsNavigationKey() != tt.expectNavigation {
				t.Errorf("IsNavigationKey(): expected %v, got %v", tt.expectNavigation, tt.key.IsNavigationKey())
			}
		})
	}
}

func TestMouseEvent(t *testing.T) {
	event := MouseEvent{
		X:     100,
		Y:     200,
		Flags: MouseLeft | KbdShift,
	}

	if event.X != 100 {
		t.Errorf("Expected X=100, got %d", event.X)
	}

	if event.Y != 200 {
		t.Errorf("Expected Y=200, got %d", event.Y)
	}

	if !event.Flags.HasMouseLeft() {
		t.Error("Expected mouse left flag to be set")
	}

	if !event.Flags.HasShift() {
		t.Error("Expected shift flag to be set")
	}
}

func TestKeyboardEvent(t *testing.T) {
	event := KeyboardEvent{
		X:     50,
		Y:     75,
		Key:   KeyEscape,
		Flags: KbdCtrl,
	}

	if event.X != 50 {
		t.Errorf("Expected X=50, got %d", event.X)
	}

	if event.Y != 75 {
		t.Errorf("Expected Y=75, got %d", event.Y)
	}

	if event.Key != KeyEscape {
		t.Errorf("Expected key=KeyEscape, got %v", event.Key)
	}

	if !event.Flags.HasCtrl() {
		t.Error("Expected ctrl flag to be set")
	}
}

func TestResizeEvent(t *testing.T) {
	event := ResizeEvent{
		Width:  800,
		Height: 600,
	}

	if event.Width != 800 {
		t.Errorf("Expected Width=800, got %d", event.Width)
	}

	if event.Height != 600 {
		t.Errorf("Expected Height=600, got %d", event.Height)
	}
}

func TestBaseEventHandler(t *testing.T) {
	handler := &BaseEventHandler{}

	// All these should not panic and should do nothing
	handler.OnInit()
	handler.OnResize(100, 100)
	handler.OnIdle()
	handler.OnMouseMove(50, 50, MouseLeft)
	handler.OnMouseButtonDown(50, 50, MouseLeft)
	handler.OnMouseButtonUp(50, 50, MouseLeft)
	handler.OnKey(50, 50, KeyEscape, KbdCtrl)
	handler.OnCtrlChange()
	handler.OnDraw()
	handler.OnPostDraw(nil)
}

func TestEventHandlerInterface(t *testing.T) {
	// Test that BaseEventHandler implements EventHandler interface
	var handler EventHandler = &BaseEventHandler{}

	// This should compile without errors, proving the interface is implemented
	handler.OnInit()
	handler.OnDraw()
}

// Custom event handler for testing
type testEventHandler struct {
	BaseEventHandler
	initCalled       bool
	resizeCalled     bool
	idleCalled       bool
	mouseMoveCalled  bool
	mouseDownCalled  bool
	mouseUpCalled    bool
	keyCalled        bool
	ctrlChangeCalled bool
	drawCalled       bool
	postDrawCalled   bool
	lastResize       ResizeEvent
	lastMouse        MouseEvent
	lastKey          KeyboardEvent
}

func (h *testEventHandler) OnInit() {
	h.initCalled = true
}

func (h *testEventHandler) OnResize(width, height int) {
	h.resizeCalled = true
	h.lastResize = ResizeEvent{Width: width, Height: height}
}

func (h *testEventHandler) OnIdle() {
	h.idleCalled = true
}

func (h *testEventHandler) OnMouseMove(x, y int, flags InputFlags) {
	h.mouseMoveCalled = true
	h.lastMouse = MouseEvent{X: x, Y: y, Flags: flags}
}

func (h *testEventHandler) OnMouseButtonDown(x, y int, flags InputFlags) {
	h.mouseDownCalled = true
	h.lastMouse = MouseEvent{X: x, Y: y, Flags: flags}
}

func (h *testEventHandler) OnMouseButtonUp(x, y int, flags InputFlags) {
	h.mouseUpCalled = true
	h.lastMouse = MouseEvent{X: x, Y: y, Flags: flags}
}

func (h *testEventHandler) OnKey(x, y int, key KeyCode, flags InputFlags) {
	h.keyCalled = true
	h.lastKey = KeyboardEvent{X: x, Y: y, Key: key, Flags: flags}
}

func (h *testEventHandler) OnCtrlChange() {
	h.ctrlChangeCalled = true
}

func (h *testEventHandler) OnDraw() {
	h.drawCalled = true
}

func (h *testEventHandler) OnPostDraw(rawHandler interface{}) {
	h.postDrawCalled = true
}

func TestCustomEventHandler(t *testing.T) {
	handler := &testEventHandler{}

	// Test that all methods can be called
	handler.OnInit()
	if !handler.initCalled {
		t.Error("OnInit was not recorded")
	}

	handler.OnResize(800, 600)
	if !handler.resizeCalled {
		t.Error("OnResize was not recorded")
	}
	if handler.lastResize.Width != 800 || handler.lastResize.Height != 600 {
		t.Errorf("Resize event not recorded correctly: got %+v", handler.lastResize)
	}

	handler.OnMouseMove(100, 200, MouseLeft|KbdShift)
	if !handler.mouseMoveCalled {
		t.Error("OnMouseMove was not recorded")
	}
	if handler.lastMouse.X != 100 || handler.lastMouse.Y != 200 {
		t.Errorf("Mouse event coordinates not recorded correctly: got %+v", handler.lastMouse)
	}
	if !handler.lastMouse.Flags.HasMouseLeft() || !handler.lastMouse.Flags.HasShift() {
		t.Errorf("Mouse event flags not recorded correctly: got %+v", handler.lastMouse.Flags)
	}

	handler.OnKey(50, 75, KeyF5, KbdCtrl)
	if !handler.keyCalled {
		t.Error("OnKey was not recorded")
	}
	if handler.lastKey.Key != KeyF5 {
		t.Errorf("Key event key not recorded correctly: got %+v", handler.lastKey.Key)
	}
	if !handler.lastKey.Flags.HasCtrl() {
		t.Errorf("Key event flags not recorded correctly: got %+v", handler.lastKey.Flags)
	}

	handler.OnDraw()
	if !handler.drawCalled {
		t.Error("OnDraw was not recorded")
	}
}

func TestKeyCodeConstants(t *testing.T) {
	// Test that key code constants have expected values (matching SDL constants)
	expectedValues := map[KeyCode]int{
		KeyBackspace: 8,
		KeyTab:       9,
		KeyReturn:    13,
		KeyEscape:    27,
		KeyDelete:    127,
		KeyKP0:       256,
		KeyKP9:       265,
		KeyUp:        273,
		KeyDown:      274,
		KeyRight:     275,
		KeyLeft:      276,
		KeyF1:        282,
		KeyF12:       293,
		KeyNumLock:   300,
		KeyCapsLock:  301,
	}

	for key, expectedValue := range expectedValues {
		if int(key) != expectedValue {
			t.Errorf("Key %s: expected value %d, got %d", key.String(), expectedValue, int(key))
		}
	}
}

func TestInputFlagsConstants(t *testing.T) {
	// Test that input flag constants have expected bit values
	expectedValues := map[InputFlags]uint32{
		MouseLeft:  1,
		MouseRight: 2,
		KbdShift:   4,
		KbdCtrl:    8,
	}

	for flag, expectedValue := range expectedValues {
		if uint32(flag) != expectedValue {
			flagName := strings.Split(flag.String(), "|")[0] // Get first flag name
			t.Errorf("Flag %s: expected value %d, got %d", flagName, expectedValue, uint32(flag))
		}
	}
}

func TestInputFlagsCombinations(t *testing.T) {
	// Test various flag combinations
	tests := []struct {
		name  string
		flags InputFlags
		parts []string
	}{
		{"Mouse buttons", MouseLeft | MouseRight, []string{"MouseLeft", "MouseRight"}},
		{"Keyboard modifiers", KbdShift | KbdCtrl, []string{"KbdShift", "KbdCtrl"}},
		{"Mixed input", MouseLeft | KbdCtrl, []string{"MouseLeft", "KbdCtrl"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flags.String()
			for _, part := range tt.parts {
				if !strings.Contains(result, part) {
					t.Errorf("Expected '%s' to contain '%s'", result, part)
				}
			}
		})
	}
}
