// Package rbox provides a radio button group control implementation for AGG.
// This is a port of AGG's rbox_ctrl_impl and rbox_ctrl classes from agg_rbox_ctrl.h/cpp.
package rbox

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl"
	"agg_go/internal/ctrl/text"
	"agg_go/internal/shapes"
)

// ellipseAdapter adapts shapes.Ellipse to conv.VertexSource interface
type ellipseAdapter struct {
	ellipse *shapes.Ellipse
}

func newEllipseAdapter(ellipse *shapes.Ellipse) *ellipseAdapter {
	return &ellipseAdapter{ellipse: ellipse}
}

func (ea *ellipseAdapter) Rewind(pathID uint) {
	ea.ellipse.Rewind(uint32(pathID))
}

func (ea *ellipseAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = ea.ellipse.Vertex(&x, &y)
	return x, y, cmd
}

// RboxCtrl implements a radio button group control.
// This corresponds to AGG's rbox_ctrl_impl class from agg_rbox_ctrl.h.
type RboxCtrl[C any] struct {
	*ctrl.BaseCtrl

	// Border styling
	borderWidth float64
	borderExtra float64

	// Text styling
	textThickness float64
	textHeight    float64
	textWidth     float64

	// Items and state
	items    [32]string // Radio button labels (max 32 items like C++)
	numItems uint32     // Number of items currently added
	curItem  int        // Currently selected item (-1 for none selected)

	// Calculated bounds
	xs1, ys1, xs2, ys2 float64 // Inner bounds after border

	// Rendering components
	ellipse       *shapes.Ellipse  // For radio button circles
	ellipseStroke *conv.ConvStroke // For stroked circles
	textRenderer  *text.SimpleText // For text labels

	// Vertex generation state
	vertices    [32]float64 // Pre-calculated vertices (x,y pairs)
	currentPath uint        // Current path being rendered
	drawItem    uint32      // Current item being drawn
	vertexIndex uint        // Current vertex index
	dy          float64     // Vertical spacing between items

	// Colors for the 5 rendering paths
	colors [5]C
}

// NewRboxCtrl creates a new radio button group control.
// x1, y1, x2, y2: bounding rectangle
// flipY: whether to flip Y coordinates
// backgroundColor, borderColor, textColor, inactiveColor, activeColor: colors for 5 rendering paths
func NewRboxCtrl[C any](x1, y1, x2, y2 float64, flipY bool, backgroundColor, borderColor, textColor, inactiveColor, activeColor C) *RboxCtrl[C] {
	rbox := &RboxCtrl[C]{
		BaseCtrl:      ctrl.NewBaseCtrl(x1, y1, x2, y2, flipY),
		borderWidth:   1.0,
		borderExtra:   0.0,
		textThickness: 1.5,
		textHeight:    9.0,
		textWidth:     0.0, // 0.0 means proportional width
		numItems:      0,
		curItem:       -1, // No item selected initially
		currentPath:   0,
		drawItem:      0,
		vertexIndex:   0,
		dy:            18.0, // textHeight * 2.0
	}

	// Initialize rendering components
	rbox.ellipse = shapes.NewEllipse()
	ellipseAdapter := newEllipseAdapter(rbox.ellipse)
	rbox.ellipseStroke = conv.NewConvStroke(ellipseAdapter)
	rbox.textRenderer = text.NewSimpleText()

	// Set colors with provided values
	rbox.colors[0] = backgroundColor // Background
	rbox.colors[1] = borderColor     // Border
	rbox.colors[2] = textColor       // Text
	rbox.colors[3] = inactiveColor   // Inactive
	rbox.colors[4] = activeColor     // Active

	rbox.calcRbox()
	return rbox
}

// NewDefaultRboxCtrl creates a radio button group control with default RGBA colors.
// This provides backward compatibility for existing code.
func NewDefaultRboxCtrl(x1, y1, x2, y2 float64, flipY bool) *RboxCtrl[color.RGBA] {
	// Default colors matching C++ AGG defaults
	backgroundColor := color.NewRGBA(1.0, 1.0, 0.9, 1.0) // Background - light yellow
	borderColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0)     // Border - black
	textColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0)       // Text - black
	inactiveColor := color.NewRGBA(0.0, 0.0, 0.0, 1.0)   // Inactive - black
	activeColor := color.NewRGBA(0.4, 0.0, 0.0, 1.0)     // Active - dark red

	return NewRboxCtrl[color.RGBA](x1, y1, x2, y2, flipY, backgroundColor, borderColor, textColor, inactiveColor, activeColor)
}

// AddItem adds a radio button item with the specified label.
// Returns true if the item was added, false if the maximum number (32) was reached.
func (r *RboxCtrl[C]) AddItem(text string) bool {
	if r.numItems < 32 {
		r.items[r.numItems] = text
		r.numItems++
		return true
	}
	return false
}

// CurItem returns the index of the currently selected item (-1 if none).
func (r *RboxCtrl[C]) CurItem() int {
	return r.curItem
}

// SetCurItem sets the currently selected item.
// Use -1 to deselect all items.
func (r *RboxCtrl[C]) SetCurItem(item int) {
	if item >= -1 && item < int(r.numItems) {
		r.curItem = item
	}
}

// NumItems returns the number of items in the radio button group.
func (r *RboxCtrl[C]) NumItems() uint32 {
	return r.numItems
}

// ItemText returns the text for the specified item index.
// Returns empty string if index is out of range.
func (r *RboxCtrl[C]) ItemText(index int) string {
	if index >= 0 && index < int(r.numItems) {
		return r.items[index]
	}
	return ""
}

// Border styling methods

// SetBorderWidth sets the border width and optional extra border space.
func (r *RboxCtrl[C]) SetBorderWidth(width, extra float64) {
	r.borderWidth = width
	r.borderExtra = extra
	r.calcRbox()
}

// BorderWidth returns the current border width.
func (r *RboxCtrl[C]) BorderWidth() float64 {
	return r.borderWidth
}

// Text styling methods

// SetTextThickness sets the text stroke thickness.
func (r *RboxCtrl[C]) SetTextThickness(thickness float64) {
	r.textThickness = thickness
}

// TextThickness returns the current text thickness.
func (r *RboxCtrl[C]) TextThickness() float64 {
	return r.textThickness
}

// SetTextSize sets the text height and optional width.
// If width is 0.0, proportional width is used.
func (r *RboxCtrl[C]) SetTextSize(height, width float64) {
	r.textHeight = height
	r.textWidth = width
	r.dy = height * 2.0 // Update vertical spacing
}

// TextHeight returns the current text height.
func (r *RboxCtrl[C]) TextHeight() float64 {
	return r.textHeight
}

// TextWidth returns the current text width (0.0 means proportional).
func (r *RboxCtrl[C]) TextWidth() float64 {
	return r.textWidth
}

// Color management methods

// SetBackgroundColor sets the background color.
func (r *RboxCtrl[C]) SetBackgroundColor(c C) {
	r.colors[0] = c
}

// SetBorderColor sets the border color.
func (r *RboxCtrl[C]) SetBorderColor(c C) {
	r.colors[1] = c
}

// SetTextColor sets the text color.
func (r *RboxCtrl[C]) SetTextColor(c C) {
	r.colors[2] = c
}

// SetInactiveColor sets the inactive radio button color.
func (r *RboxCtrl[C]) SetInactiveColor(c C) {
	r.colors[3] = c
}

// SetActiveColor sets the active radio button color.
func (r *RboxCtrl[C]) SetActiveColor(c C) {
	r.colors[4] = c
}

// Color returns the color for the specified path.
func (r *RboxCtrl[C]) Color(pathID uint) C {
	if pathID < 5 {
		return r.colors[pathID]
	}
	return r.colors[0] // Default to background color
}

// Event handling methods

// OnMouseButtonDown handles mouse button down events.
// Returns true if the event was handled.
func (r *RboxCtrl[C]) OnMouseButtonDown(x, y float64) bool {
	// Convert to control coordinates
	r.InverseTransformXY(&x, &y)

	// Check each radio button for clicks
	for i := uint32(0); i < r.numItems; i++ {
		xp := r.xs1 + r.dy/1.3
		yp := r.ys1 + r.dy*float64(i) + r.dy/1.3

		// Calculate distance from click to radio button center
		distance := math.Sqrt((x-xp)*(x-xp) + (y-yp)*(y-yp))

		// Check if click is within radio button circle
		if distance <= r.textHeight/1.5 {
			r.curItem = int(i)
			return true
		}
	}

	return false
}

// OnMouseButtonUp handles mouse button up events.
func (r *RboxCtrl[C]) OnMouseButtonUp(x, y float64) bool {
	return false // No special handling needed
}

// OnMouseMove handles mouse move events.
func (r *RboxCtrl[C]) OnMouseMove(x, y float64, buttonPressed bool) bool {
	return false // No special handling needed
}

// OnArrowKeys handles arrow key events for keyboard navigation.
func (r *RboxCtrl[C]) OnArrowKeys(left, right, down, up bool) bool {
	if r.curItem >= 0 {
		if up || right {
			r.curItem++
			if r.curItem >= int(r.numItems) {
				r.curItem = 0
			}
			return true
		}

		if down || left {
			r.curItem--
			if r.curItem < 0 {
				r.curItem = int(r.numItems - 1)
			}
			return true
		}
	}
	return false
}

// Vertex source interface methods

// NumPaths returns the number of rendering paths.
func (r *RboxCtrl[C]) NumPaths() uint {
	return 5 // Background, Border, Text, Inactive circles, Active circle
}

// Rewind starts rendering a specific path.
func (r *RboxCtrl[C]) Rewind(pathID uint) {
	r.currentPath = pathID
	r.drawItem = 0
	r.vertexIndex = 0

	switch pathID {
	case 0: // Background
		r.setupBackgroundPath()
	case 1: // Border
		r.setupBorderPath()
	case 2: // Text
		r.setupTextPath()
	case 3: // Inactive circles
		r.setupInactiveCirclesPath()
	case 4: // Active circle
		r.setupActiveCirclePath()
	}
}

// Vertex generates the next vertex for the current path.
func (r *RboxCtrl[C]) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch r.currentPath {
	case 0: // Background
		return r.generateBackgroundVertex()
	case 1: // Border
		return r.generateBorderVertex()
	case 2: // Text
		return r.generateTextVertex()
	case 3: // Inactive circles
		return r.generateInactiveCirclesVertex()
	case 4: // Active circle
		return r.generateActiveCircleVertex()
	}
	return 0, 0, basics.PathCmdStop
}

// Internal helper methods

// calcRbox calculates the inner bounds after accounting for border.
func (r *RboxCtrl[C]) calcRbox() {
	r.xs1 = r.X1() + r.borderWidth
	r.ys1 = r.Y1() + r.borderWidth
	r.xs2 = r.X2() - r.borderWidth
	r.ys2 = r.Y2() - r.borderWidth
}

// setupBackgroundPath prepares the background rectangle vertices.
func (r *RboxCtrl[C]) setupBackgroundPath() {
	r.vertices[0] = r.X1() - r.borderExtra
	r.vertices[1] = r.Y1() - r.borderExtra
	r.vertices[2] = r.X2() + r.borderExtra
	r.vertices[3] = r.Y1() - r.borderExtra
	r.vertices[4] = r.X2() + r.borderExtra
	r.vertices[5] = r.Y2() + r.borderExtra
	r.vertices[6] = r.X1() - r.borderExtra
	r.vertices[7] = r.Y2() + r.borderExtra
}

// setupBorderPath prepares the border rectangle vertices.
func (r *RboxCtrl[C]) setupBorderPath() {
	// Outer rectangle
	r.vertices[0] = r.X1()
	r.vertices[1] = r.Y1()
	r.vertices[2] = r.X2()
	r.vertices[3] = r.Y1()
	r.vertices[4] = r.X2()
	r.vertices[5] = r.Y2()
	r.vertices[6] = r.X1()
	r.vertices[7] = r.Y2()
	// Inner rectangle
	r.vertices[8] = r.X1() + r.borderWidth
	r.vertices[9] = r.Y1() + r.borderWidth
	r.vertices[10] = r.X1() + r.borderWidth
	r.vertices[11] = r.Y2() - r.borderWidth
	r.vertices[12] = r.X2() - r.borderWidth
	r.vertices[13] = r.Y2() - r.borderWidth
	r.vertices[14] = r.X2() - r.borderWidth
	r.vertices[15] = r.Y1() + r.borderWidth
}

// setupTextPath prepares the first text label.
func (r *RboxCtrl[C]) setupTextPath() {
	if r.numItems > 0 {
		r.textRenderer.SetText(r.items[0])
		r.textRenderer.SetPosition(r.xs1+r.dy*1.5, r.ys1+r.dy/2.0)
		r.textRenderer.SetSize(r.textHeight)
		r.textRenderer.Rewind(0)
	}
}

// setupInactiveCirclesPath prepares the first inactive circle.
// Using smaller filled circles instead of stroked circles as workaround
func (r *RboxCtrl[C]) setupInactiveCirclesPath() {
	if r.numItems > 0 {
		radius := r.textHeight / 2.5 // Smaller than active circles
		r.ellipse.Init(
			r.xs1+r.dy/1.3,
			r.ys1+r.dy/1.3,
			radius,
			radius,
			32, false)
		r.ellipse.Rewind(0)
	}
	// If no items, the ellipse won't be initialized and will return stop immediately
}

// setupActiveCirclePath prepares the active circle (filled).
func (r *RboxCtrl[C]) setupActiveCirclePath() {
	if r.curItem >= 0 {
		r.ellipse.Init(
			r.xs1+r.dy/1.3,
			r.ys1+r.dy*float64(r.curItem)+r.dy/1.3,
			r.textHeight/2.0,
			r.textHeight/2.0,
			32, false)
		r.ellipse.Rewind(0)
	}
}

// generateBackgroundVertex generates vertices for the background rectangle.
func (r *RboxCtrl[C]) generateBackgroundVertex() (x, y float64, cmd basics.PathCommand) {
	if r.vertexIndex >= 4 {
		return 0, 0, basics.PathCmdStop
	}

	cmd = basics.PathCmdLineTo
	if r.vertexIndex == 0 {
		cmd = basics.PathCmdMoveTo
	}

	x = r.vertices[r.vertexIndex*2]
	y = r.vertices[r.vertexIndex*2+1]
	r.vertexIndex++

	if r.vertexIndex < 4 {
		r.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// generateBorderVertex generates vertices for the border.
func (r *RboxCtrl[C]) generateBorderVertex() (x, y float64, cmd basics.PathCommand) {
	if r.vertexIndex >= 8 {
		return 0, 0, basics.PathCmdStop
	}

	cmd = basics.PathCmdLineTo
	if r.vertexIndex == 0 || r.vertexIndex == 4 {
		cmd = basics.PathCmdMoveTo
	}

	x = r.vertices[r.vertexIndex*2]
	y = r.vertices[r.vertexIndex*2+1]
	r.vertexIndex++

	if r.vertexIndex < 8 {
		r.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// generateTextVertex generates vertices for text labels.
func (r *RboxCtrl[C]) generateTextVertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmd = r.textRenderer.Vertex()

	if cmd == basics.PathCmdStop {
		r.drawItem++
		if r.drawItem >= r.numItems {
			r.TransformXY(&x, &y)
			return x, y, basics.PathCmdStop
		}

		// Setup next text label
		r.textRenderer.SetText(r.items[r.drawItem])
		r.textRenderer.SetPosition(
			r.xs1+r.dy*1.5,
			r.ys1+r.dy*float64(r.drawItem+1)-r.dy/2.0)
		r.textRenderer.Rewind(0)
		x, y, cmd = r.textRenderer.Vertex()
	}

	if cmd != basics.PathCmdStop {
		r.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// generateInactiveCirclesVertex generates vertices for inactive radio circles.
// Note: Using filled circles instead of stroked circles due to ConvStroke issue
func (r *RboxCtrl[C]) generateInactiveCirclesVertex() (x, y float64, cmd basics.PathCommand) {
	// If no items, return stop immediately
	if r.numItems == 0 {
		return 0, 0, basics.PathCmdStop
	}

	cmd = r.ellipse.Vertex(&x, &y)

	if cmd == basics.PathCmdStop {
		r.drawItem++
		if r.drawItem >= r.numItems {
			return 0, 0, basics.PathCmdStop
		}

		// Setup next inactive circle - using smaller radius to show as outline
		radius := r.textHeight / 2.5 // Smaller than active circles
		r.ellipse.Init(
			r.xs1+r.dy/1.3,
			r.ys1+r.dy*float64(r.drawItem)+r.dy/1.3,
			radius,
			radius,
			32, false)
		r.ellipse.Rewind(0)
		cmd = r.ellipse.Vertex(&x, &y)
	}

	if cmd != basics.PathCmdStop {
		r.TransformXY(&x, &y)
	}
	return x, y, cmd
}

// generateActiveCircleVertex generates vertices for the active (filled) circle.
func (r *RboxCtrl[C]) generateActiveCircleVertex() (x, y float64, cmd basics.PathCommand) {
	if r.curItem < 0 {
		return 0, 0, basics.PathCmdStop
	}

	cmd = r.ellipse.Vertex(&x, &y)
	if cmd != basics.PathCmdStop {
		r.TransformXY(&x, &y)
	}
	return x, y, cmd
}
