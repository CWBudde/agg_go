// Package checkbox provides a checkbox control implementation for AGG.
// This is a port of AGG's cbox_ctrl_impl and cbox_ctrl classes.
package checkbox

import (
	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl"
	"agg_go/internal/ctrl/text"
)

// CheckboxCtrl implements a checkbox control with label support.
// This corresponds to AGG's cbox_ctrl template class with ColorT parameter.
type CheckboxCtrl[C any] struct {
	*ctrl.BaseCtrl

	// Checkbox state
	checked bool

	// Label settings
	label         string
	textThickness float64
	textHeight    float64
	textWidth     float64

	// Rendering components
	textRenderer *text.SimpleText

	// Vertex generation state - matching C++ implementation
	vertices    [32]float64 // Pre-calculated vertices (x,y pairs for max 16 points)
	currentPath uint
	vertexIndex uint

	// Colors for the 3 rendering paths (inactive, text, active)
	colors [3]C
}

// NewCheckboxCtrl creates a new checkbox control.
// x, y: position of the checkbox (top-left corner)
// label: text label to display next to the checkbox
// flipY: whether to flip Y coordinates
// inactiveColor, textColor, activeColor: colors for the 3 rendering paths
func NewCheckboxCtrl[C any](x, y float64, label string, flipY bool, inactiveColor, textColor, activeColor C) *CheckboxCtrl[C] {
	// Calculate bounds: checkbox is 9.0 * 1.5 units square, following C++ implementation
	checkboxSize := 9.0 * 1.5

	checkbox := &CheckboxCtrl[C]{
		BaseCtrl:      ctrl.NewBaseCtrl(x, y, x+checkboxSize, y+checkboxSize, flipY),
		checked:       false,
		label:         label,
		textThickness: 1.5,
		textHeight:    9.0,
		textWidth:     0.0, // 0.0 means proportional width
		textRenderer:  text.NewSimpleText(),
		currentPath:   0,
		vertexIndex:   0,
	}

	// Set initial label
	checkbox.setLabel(label)

	// Initialize colors with provided values
	checkbox.colors[0] = inactiveColor // Inactive (border)
	checkbox.colors[1] = textColor     // Text
	checkbox.colors[2] = activeColor   // Active (checkmark)

	return checkbox
}

// NewDefaultCheckboxCtrl creates a checkbox with default RGBA colors.
// This provides backward compatibility for existing code.
func NewDefaultCheckboxCtrl(x, y float64, label string, flipY bool) *CheckboxCtrl[color.RGBA] {
	// Default colors matching C++ AGG defaults
	inactiveColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0) // Inactive (border) - black
	textColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0)     // Text - black
	activeColor := color.NewRGBA(0.4, 0.0, 0.0, 1.0)   // Active (checkmark) - dark red

	return NewCheckboxCtrl[color.RGBA](x, y, label, flipY, inactiveColor, textColor, activeColor)
}

// State Management Methods

// IsChecked returns the current checkbox state.
func (c *CheckboxCtrl[C]) IsChecked() bool {
	return c.checked
}

// SetChecked sets the checkbox state.
func (c *CheckboxCtrl[C]) SetChecked(checked bool) {
	c.checked = checked
}

// Toggle toggles the checkbox state.
func (c *CheckboxCtrl[C]) Toggle() {
	c.checked = !c.checked
}

// Label and Text Configuration Methods

// Label returns the current label text.
func (c *CheckboxCtrl[C]) Label() string {
	return c.label
}

// SetLabel sets the label text.
func (c *CheckboxCtrl[C]) SetLabel(label string) {
	c.setLabel(label)
}

// setLabel is the internal implementation that copies the label string.
func (c *CheckboxCtrl[C]) setLabel(label string) {
	// Limit label length to 127 characters like C++ implementation
	if len(label) > 127 {
		label = label[:127]
	}
	c.label = label
}

// SetTextThickness sets the thickness of text rendering.
func (c *CheckboxCtrl[C]) SetTextThickness(thickness float64) {
	c.textThickness = thickness
	c.textRenderer.SetThickness(thickness)
}

// SetTextSize sets the text height and width.
// width of 0.0 means proportional width.
func (c *CheckboxCtrl[C]) SetTextSize(height, width float64) {
	c.textHeight = height
	c.textWidth = width
	c.textRenderer.SetSize(height)
}

// Color Management Methods

// SetTextColor sets the text color.
func (c *CheckboxCtrl[C]) SetTextColor(clr C) {
	c.colors[1] = clr
}

// SetInactiveColor sets the inactive (border) color.
func (c *CheckboxCtrl[C]) SetInactiveColor(clr C) {
	c.colors[0] = clr
}

// SetActiveColor sets the active (checkmark) color.
func (c *CheckboxCtrl[C]) SetActiveColor(clr C) {
	c.colors[2] = clr
}

// Mouse Interaction Methods

// OnMouseButtonDown handles mouse button press events.
func (c *CheckboxCtrl[C]) OnMouseButtonDown(x, y float64) bool {
	// Transform screen coordinates to control coordinates
	c.InverseTransformXY(&x, &y)

	// Check if click is within control bounds
	if x >= c.X1() && y >= c.Y1() && x <= c.X2() && y <= c.Y2() {
		c.Toggle()
		return true
	}
	return false
}

// OnMouseButtonUp handles mouse button release events.
func (c *CheckboxCtrl[C]) OnMouseButtonUp(x, y float64) bool {
	return false
}

// OnMouseMove handles mouse move events.
func (c *CheckboxCtrl[C]) OnMouseMove(x, y float64, buttonPressed bool) bool {
	return false
}

// OnArrowKeys handles arrow key events.
func (c *CheckboxCtrl[C]) OnArrowKeys(left, right, down, up bool) bool {
	return false
}

// Vertex Source Interface Implementation

// NumPaths returns the number of rendering paths (3: border, text, checkmark).
func (c *CheckboxCtrl[C]) NumPaths() uint {
	return 3
}

// Color returns the color for the specified path.
func (c *CheckboxCtrl[C]) Color(pathID uint) C {
	if pathID < uint(len(c.colors)) {
		return c.colors[pathID]
	}
	return c.colors[0] // Default to inactive color
}

// Rewind prepares the specified path for vertex generation.
func (c *CheckboxCtrl[C]) Rewind(pathID uint) {
	c.currentPath = pathID
	c.vertexIndex = 0

	switch pathID {
	case 0: // Border path
		c.generateBorderVertices()
	case 1: // Text path
		c.generateTextVertices()
	case 2: // Checkmark path (only if checked)
		if c.checked {
			c.generateCheckmarkVertices()
		}
	}
}

// Vertex returns the next vertex in the current path.
func (c *CheckboxCtrl[C]) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch c.currentPath {
	case 0: // Border path
		return c.getBorderVertex()
	case 1: // Text path
		return c.getTextVertex()
	case 2: // Checkmark path
		if c.checked {
			return c.getCheckmarkVertex()
		}
		return 0, 0, basics.PathCmdStop
	default:
		return 0, 0, basics.PathCmdStop
	}
}

// Border vertex generation (path 0)
func (c *CheckboxCtrl[C]) generateBorderVertices() {
	x1, y1, x2, y2 := c.X1(), c.Y1(), c.X2(), c.Y2()
	t := c.textThickness

	// Outer rectangle vertices
	c.vertices[0], c.vertices[1] = x1, y1 // 0: top-left
	c.vertices[2], c.vertices[3] = x2, y1 // 1: top-right
	c.vertices[4], c.vertices[5] = x2, y2 // 2: bottom-right
	c.vertices[6], c.vertices[7] = x1, y2 // 3: bottom-left

	// Inner rectangle vertices (for hollow effect)
	c.vertices[8], c.vertices[9] = x1+t, y1+t   // 4: inner top-left
	c.vertices[10], c.vertices[11] = x1+t, y2-t // 5: inner bottom-left
	c.vertices[12], c.vertices[13] = x2-t, y2-t // 6: inner bottom-right
	c.vertices[14], c.vertices[15] = x2-t, y1+t // 7: inner top-right
}

func (c *CheckboxCtrl[C]) getBorderVertex() (x, y float64, cmd basics.PathCommand) {
	if c.vertexIndex >= 8 {
		return 0, 0, basics.PathCmdStop
	}

	var command basics.PathCommand
	if c.vertexIndex == 0 || c.vertexIndex == 4 {
		command = basics.PathCmdMoveTo
	} else {
		command = basics.PathCmdLineTo
	}

	x = c.vertices[c.vertexIndex*2]
	y = c.vertices[c.vertexIndex*2+1]
	c.vertexIndex++

	// Transform coordinates
	c.TransformXY(&x, &y)

	return x, y, command
}

// Text vertex generation (path 1)
func (c *CheckboxCtrl[C]) generateTextVertices() {
	if c.label == "" {
		return
	}

	// Position text to the right of the checkbox with some spacing
	textX := c.X1() + c.textHeight*2.0
	textY := c.Y1() + c.textHeight/5.0

	// Configure text renderer
	c.textRenderer.SetText(c.label)
	c.textRenderer.SetPosition(textX, textY)
	c.textRenderer.SetSize(c.textHeight)
	c.textRenderer.SetThickness(c.textThickness)
}

func (c *CheckboxCtrl[C]) getTextVertex() (x, y float64, cmd basics.PathCommand) {
	if c.label == "" {
		return 0, 0, basics.PathCmdStop
	}

	// Get vertex from text renderer
	x, y, cmd = c.textRenderer.Vertex()

	// Transform coordinates if not stop command
	if cmd != basics.PathCmdStop {
		c.TransformXY(&x, &y)
	}

	return x, y, cmd
}

// Checkmark vertex generation (path 2) - X-shaped checkmark
func (c *CheckboxCtrl[C]) generateCheckmarkVertices() {
	x1, y1, x2, y2 := c.X1(), c.Y1(), c.X2(), c.Y2()
	t := c.textThickness * 1.5
	d2 := (y2 - y1) / 2.0 // Half height for center calculation

	// Generate X-shaped checkmark vertices following C++ implementation
	c.vertices[0], c.vertices[1] = x1+c.textThickness, y1+c.textThickness   // 0
	c.vertices[2], c.vertices[3] = x1+d2, y1+d2-t                           // 1
	c.vertices[4], c.vertices[5] = x2-c.textThickness, y1+c.textThickness   // 2
	c.vertices[6], c.vertices[7] = x1+d2+t, y1+d2                           // 3
	c.vertices[8], c.vertices[9] = x2-c.textThickness, y2-c.textThickness   // 4
	c.vertices[10], c.vertices[11] = x1+d2, y1+d2+t                         // 5
	c.vertices[12], c.vertices[13] = x1+c.textThickness, y2-c.textThickness // 6
	c.vertices[14], c.vertices[15] = x1+d2-t, y1+d2                         // 7
}

func (c *CheckboxCtrl[C]) getCheckmarkVertex() (x, y float64, cmd basics.PathCommand) {
	if c.vertexIndex >= 8 {
		return 0, 0, basics.PathCmdStop
	}

	var command basics.PathCommand
	if c.vertexIndex == 0 {
		command = basics.PathCmdMoveTo
	} else {
		command = basics.PathCmdLineTo
	}

	x = c.vertices[c.vertexIndex*2]
	y = c.vertices[c.vertexIndex*2+1]
	c.vertexIndex++

	// Transform coordinates
	c.TransformXY(&x, &y)

	return x, y, command
}
