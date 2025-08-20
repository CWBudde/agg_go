// Package gamma provides interactive gamma correction controls for AGG.
// This is a port of AGG's gamma_ctrl_impl and gamma_ctrl classes.
package gamma

import (
	"fmt"
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/conv"
	"agg_go/internal/ctrl"
	"agg_go/internal/ctrl/text"
	"agg_go/internal/shapes"
)

// GammaCtrlImpl implements the core gamma correction control functionality.
// This is a Go port of AGG's gamma_ctrl_impl class from agg_gamma_ctrl.h/cpp.
//
// The control provides:
// - Interactive gamma curve editing with spline-based control points
// - Real-time gamma preview functionality
// - Visual curve display with grid background
// - Mouse interaction for dragging control points
// - Keyboard support for fine adjustments
// - Text display of current control point values
type GammaCtrlImpl struct {
	*ctrl.BaseCtrl

	// Core gamma functionality
	gammaSpline *GammaSpline

	// Appearance settings
	borderWidth   float64
	borderExtra   float64
	curveWidth    float64
	gridWidth     float64
	textThickness float64
	pointSize     float64
	textHeight    float64
	textWidth     float64

	// Layout coordinates
	xc1, yc1, xc2, yc2 float64 // Control area bounds
	xs1, ys1, xs2, ys2 float64 // Spline area bounds
	xt1, yt1, xt2, yt2 float64 // Text area bounds

	// Control points in screen coordinates
	xp1, yp1, xp2, yp2 float64

	// Mouse interaction state
	p1Active   bool    // Which control point is active
	mousePoint uint    // Which point is being dragged (0=none, 1=point1, 2=point2)
	pdx, pdy   float64 // Mouse drag offset

	// Rendering components
	curveStroke  *conv.ConvStroke
	ellipse      *shapes.Ellipse
	textRenderer *text.SimpleText
	textStroke   *conv.ConvStroke

	// Vertex generation state
	currentPath uint
	vertexIndex uint
	vertices    [32]float64 // Pre-calculated vertices for current path
	vertexCount uint
}

// NewGammaCtrlImpl creates a new gamma control implementation.
// x1, y1, x2, y2: bounding rectangle
// flipY: whether to flip Y coordinates
func NewGammaCtrlImpl(x1, y1, x2, y2 float64, flipY bool) *GammaCtrlImpl {
	ctrl := &GammaCtrlImpl{
		BaseCtrl:      ctrl.NewBaseCtrl(x1, y1, x2, y2, flipY),
		gammaSpline:   NewGammaSpline(),
		borderWidth:   2.0,
		borderExtra:   0.0,
		curveWidth:    2.0,
		gridWidth:     0.2,
		textThickness: 1.5,
		pointSize:     5.0,
		textHeight:    9.0,
		textWidth:     0.0,
		xc1:           x1,
		yc1:           y1,
		xc2:           x2,
		yc2:           y2 - 9.0*2.0, // Reserve space for text
		xt1:           x1,
		yt1:           y2 - 9.0*2.0,
		xt2:           x2,
		yt2:           y2,
		p1Active:      true,
		mousePoint:    0,
		pdx:           0.0,
		pdy:           0.0,
		ellipse:       shapes.NewEllipse(),
		textRenderer:  text.NewSimpleText(),
		currentPath:   0,
		vertexIndex:   0,
		vertexCount:   0,
	}

	// Initialize curve stroke
	ctrl.curveStroke = conv.NewConvStroke(ctrl.gammaSpline)
	ctrl.curveStroke.SetWidth(ctrl.curveWidth)

	// Initialize text stroke
	ctrl.textStroke = conv.NewConvStroke(ctrl.textRenderer)
	ctrl.textStroke.SetWidth(ctrl.textThickness)
	ctrl.textStroke.SetLineCap(basics.RoundCap)
	ctrl.textStroke.SetLineJoin(basics.RoundJoin)

	ctrl.calcSplineBox()
	return ctrl
}

// calcSplineBox calculates the inner bounds for the spline area.
func (gc *GammaCtrlImpl) calcSplineBox() {
	gc.xs1 = gc.xc1 + gc.borderWidth
	gc.ys1 = gc.yc1 + gc.borderWidth
	gc.xs2 = gc.xc2 - gc.borderWidth
	gc.ys2 = gc.yc2 - gc.borderWidth*0.5
}

// calcPoints calculates the screen coordinates of the control points.
func (gc *GammaCtrlImpl) calcPoints() {
	kx1, ky1, kx2, ky2 := gc.gammaSpline.GetValues()
	gc.xp1 = gc.xs1 + (gc.xs2-gc.xs1)*kx1*0.25
	gc.yp1 = gc.ys1 + (gc.ys2-gc.ys1)*ky1*0.25
	gc.xp2 = gc.xs2 - (gc.xs2-gc.xs1)*kx2*0.25
	gc.yp2 = gc.ys2 - (gc.ys2-gc.ys1)*ky2*0.25
}

// calcValues calculates the control point values from screen coordinates.
func (gc *GammaCtrlImpl) calcValues() {
	kx1 := (gc.xp1 - gc.xs1) * 4.0 / (gc.xs2 - gc.xs1)
	ky1 := (gc.yp1 - gc.ys1) * 4.0 / (gc.ys2 - gc.ys1)
	kx2 := (gc.xs2 - gc.xp2) * 4.0 / (gc.xs2 - gc.xs1)
	ky2 := (gc.ys2 - gc.yp2) * 4.0 / (gc.ys2 - gc.ys1)
	gc.gammaSpline.Values(kx1, ky1, kx2, ky2)
}

// SetBorderWidth sets the border width and extra border space.
func (gc *GammaCtrlImpl) SetBorderWidth(width, extra float64) {
	gc.borderWidth = width
	gc.borderExtra = extra
	gc.calcSplineBox()
}

// SetCurveWidth sets the width of the gamma curve line.
func (gc *GammaCtrlImpl) SetCurveWidth(width float64) {
	gc.curveWidth = width
	gc.curveStroke.SetWidth(width)
}

// SetGridWidth sets the width of the grid lines.
func (gc *GammaCtrlImpl) SetGridWidth(width float64) {
	gc.gridWidth = width
}

// SetTextThickness sets the thickness of text rendering.
func (gc *GammaCtrlImpl) SetTextThickness(thickness float64) {
	gc.textThickness = thickness
	gc.textStroke.SetWidth(thickness)
}

// SetTextSize sets the text size.
func (gc *GammaCtrlImpl) SetTextSize(height, width float64) {
	gc.textWidth = width
	gc.textHeight = height
	gc.yc2 = gc.Y2() - gc.textHeight*2.0
	gc.yt1 = gc.Y2() - gc.textHeight*2.0
	gc.calcSplineBox()
}

// SetPointSize sets the size of control point markers.
func (gc *GammaCtrlImpl) SetPointSize(size float64) {
	gc.pointSize = size
}

// Values sets the gamma curve control points.
func (gc *GammaCtrlImpl) Values(kx1, ky1, kx2, ky2 float64) {
	gc.gammaSpline.Values(kx1, ky1, kx2, ky2)
}

// GetValues returns the current gamma curve control points.
func (gc *GammaCtrlImpl) GetValues() (kx1, ky1, kx2, ky2 float64) {
	return gc.gammaSpline.GetValues()
}

// Gamma returns the gamma lookup table.
func (gc *GammaCtrlImpl) Gamma() []uint8 {
	return gc.gammaSpline.Gamma()
}

// Y calculates the gamma-corrected value for a given input.
func (gc *GammaCtrlImpl) Y(x float64) float64 {
	return gc.gammaSpline.Y(x)
}

// GetGammaSpline returns the underlying gamma spline.
func (gc *GammaCtrlImpl) GetGammaSpline() *GammaSpline {
	return gc.gammaSpline
}

// ChangeActivePoint toggles which control point is active.
func (gc *GammaCtrlImpl) ChangeActivePoint() {
	gc.p1Active = !gc.p1Active
}

// InRect checks if a point is within the control bounds.
func (gc *GammaCtrlImpl) InRect(x, y float64) bool {
	gc.InverseTransformXY(&x, &y)
	return x >= gc.X1() && x <= gc.X2() && y >= gc.Y1() && y <= gc.Y2()
}

// OnMouseButtonDown handles mouse button press events.
func (gc *GammaCtrlImpl) OnMouseButtonDown(x, y float64) bool {
	gc.InverseTransformXY(&x, &y)
	gc.calcPoints()

	// Check if clicked on first control point
	if gc.calcDistance(x, y, gc.xp1, gc.yp1) <= gc.pointSize+1 {
		gc.mousePoint = 1
		gc.pdx = gc.xp1 - x
		gc.pdy = gc.yp1 - y
		gc.p1Active = true
		return true
	}

	// Check if clicked on second control point
	if gc.calcDistance(x, y, gc.xp2, gc.yp2) <= gc.pointSize+1 {
		gc.mousePoint = 2
		gc.pdx = gc.xp2 - x
		gc.pdy = gc.yp2 - y
		gc.p1Active = false
		return true
	}

	return false
}

// OnMouseButtonUp handles mouse button release events.
func (gc *GammaCtrlImpl) OnMouseButtonUp(x, y float64) bool {
	if gc.mousePoint != 0 {
		gc.mousePoint = 0
		return true
	}
	return false
}

// OnMouseMove handles mouse movement events.
func (gc *GammaCtrlImpl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	gc.InverseTransformXY(&x, &y)

	if !buttonPressed {
		return gc.OnMouseButtonUp(x, y)
	}

	if gc.mousePoint == 1 {
		gc.xp1 = x + gc.pdx
		gc.yp1 = y + gc.pdy
		gc.calcValues()
		return true
	}

	if gc.mousePoint == 2 {
		gc.xp2 = x + gc.pdx
		gc.yp2 = y + gc.pdy
		gc.calcValues()
		return true
	}

	return false
}

// OnArrowKeys handles keyboard arrow key events for fine adjustment.
func (gc *GammaCtrlImpl) OnArrowKeys(left, right, down, up bool) bool {
	kx1, ky1, kx2, ky2 := gc.gammaSpline.GetValues()
	ret := false

	if gc.p1Active {
		if left {
			kx1 -= 0.005
			ret = true
		}
		if right {
			kx1 += 0.005
			ret = true
		}
		if down {
			ky1 -= 0.005
			ret = true
		}
		if up {
			ky1 += 0.005
			ret = true
		}
	} else {
		if left {
			kx2 += 0.005
			ret = true
		}
		if right {
			kx2 -= 0.005
			ret = true
		}
		if down {
			ky2 += 0.005
			ret = true
		}
		if up {
			ky2 -= 0.005
			ret = true
		}
	}

	if ret {
		gc.gammaSpline.Values(kx1, ky1, kx2, ky2)
	}
	return ret
}

// calcDistance calculates the distance between two points.
func (gc *GammaCtrlImpl) calcDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// NumPaths returns the number of rendering paths.
func (gc *GammaCtrlImpl) NumPaths() uint {
	return 7
}

// Rewind prepares a specific path for vertex generation.
func (gc *GammaCtrlImpl) Rewind(pathID uint) {
	gc.currentPath = pathID
	gc.vertexIndex = 0
	gc.vertexCount = 0

	switch pathID {
	case 0: // Background
		gc.setupBackgroundPath()
	case 1: // Border
		gc.setupBorderPath()
	case 2: // Curve
		gc.setupCurvePath()
	case 3: // Grid
		gc.setupGridPath()
	case 4: // Point 1 (inactive)
		gc.setupPoint1Path()
	case 5: // Point 2 (active)
		gc.setupPoint2Path()
	case 6: // Text
		gc.setupTextPath()
	}
}

// Vertex returns the next vertex for the current path.
func (gc *GammaCtrlImpl) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch gc.currentPath {
	case 0, 1, 3: // Background, Border, Grid - use pre-calculated vertices
		return gc.getPreCalculatedVertex()
	case 2: // Curve - use stroke converter
		return gc.curveStroke.Vertex()
	case 4, 5: // Points - use ellipse
		var x, y float64
		cmd := gc.ellipse.Vertex(&x, &y)
		return x, y, cmd
	case 6: // Text - use text stroke
		return gc.textStroke.Vertex()
	default:
		return 0, 0, basics.PathCmdStop
	}
}

// Color returns the color for a specific path (placeholder for interface).
func (gc *GammaCtrlImpl) Color(pathID uint) interface{} {
	// This will be overridden by the templated GammaCtrl
	return nil
}

// setupBackgroundPath prepares vertices for the background rectangle.
func (gc *GammaCtrlImpl) setupBackgroundPath() {
	gc.vertices[0] = gc.X1() - gc.borderExtra
	gc.vertices[1] = gc.Y1() - gc.borderExtra
	gc.vertices[2] = gc.X2() + gc.borderExtra
	gc.vertices[3] = gc.Y1() - gc.borderExtra
	gc.vertices[4] = gc.X2() + gc.borderExtra
	gc.vertices[5] = gc.Y2() + gc.borderExtra
	gc.vertices[6] = gc.X1() - gc.borderExtra
	gc.vertices[7] = gc.Y2() + gc.borderExtra
	gc.vertexCount = 4
}

// setupBorderPath prepares vertices for the border rectangles.
func (gc *GammaCtrlImpl) setupBorderPath() {
	// Outer border
	gc.vertices[0] = gc.X1()
	gc.vertices[1] = gc.Y1()
	gc.vertices[2] = gc.X2()
	gc.vertices[3] = gc.Y1()
	gc.vertices[4] = gc.X2()
	gc.vertices[5] = gc.Y2()
	gc.vertices[6] = gc.X1()
	gc.vertices[7] = gc.Y2()

	// Inner border (spline area)
	gc.vertices[8] = gc.X1() + gc.borderWidth
	gc.vertices[9] = gc.Y1() + gc.borderWidth
	gc.vertices[10] = gc.X1() + gc.borderWidth
	gc.vertices[11] = gc.Y2() - gc.borderWidth
	gc.vertices[12] = gc.X2() - gc.borderWidth
	gc.vertices[13] = gc.Y2() - gc.borderWidth
	gc.vertices[14] = gc.X2() - gc.borderWidth
	gc.vertices[15] = gc.Y1() + gc.borderWidth

	// Text area separator
	gc.vertices[16] = gc.xc1 + gc.borderWidth
	gc.vertices[17] = gc.yc2 - gc.borderWidth*0.5
	gc.vertices[18] = gc.xc2 - gc.borderWidth
	gc.vertices[19] = gc.yc2 - gc.borderWidth*0.5
	gc.vertices[20] = gc.xc2 - gc.borderWidth
	gc.vertices[21] = gc.yc2 + gc.borderWidth*0.5
	gc.vertices[22] = gc.xc1 + gc.borderWidth
	gc.vertices[23] = gc.yc2 + gc.borderWidth*0.5

	gc.vertexCount = 12
}

// setupCurvePath prepares the gamma curve for rendering.
func (gc *GammaCtrlImpl) setupCurvePath() {
	// Ensure we have valid bounds before setting up the curve
	if gc.xs2 <= gc.xs1 || gc.ys2 <= gc.ys1 {
		return // Invalid bounds, skip setup
	}
	gc.gammaSpline.Box(gc.xs1, gc.ys1, gc.xs2, gc.ys2)
	gc.curveStroke.SetWidth(gc.curveWidth)
	gc.curveStroke.Rewind(0)
}

// setupGridPath prepares vertices for the grid lines.
func (gc *GammaCtrlImpl) setupGridPath() {
	gc.calcPoints()

	// Horizontal center line
	gc.vertices[0] = gc.xs1
	gc.vertices[1] = (gc.ys1+gc.ys2)*0.5 - gc.gridWidth*0.5
	gc.vertices[2] = gc.xs2
	gc.vertices[3] = (gc.ys1+gc.ys2)*0.5 - gc.gridWidth*0.5
	gc.vertices[4] = gc.xs2
	gc.vertices[5] = (gc.ys1+gc.ys2)*0.5 + gc.gridWidth*0.5
	gc.vertices[6] = gc.xs1
	gc.vertices[7] = (gc.ys1+gc.ys2)*0.5 + gc.gridWidth*0.5

	// Vertical center line
	gc.vertices[8] = (gc.xs1+gc.xs2)*0.5 - gc.gridWidth*0.5
	gc.vertices[9] = gc.ys1
	gc.vertices[10] = (gc.xs1+gc.xs2)*0.5 - gc.gridWidth*0.5
	gc.vertices[11] = gc.ys2
	gc.vertices[12] = (gc.xs1+gc.xs2)*0.5 + gc.gridWidth*0.5
	gc.vertices[13] = gc.ys2
	gc.vertices[14] = (gc.xs1+gc.xs2)*0.5 + gc.gridWidth*0.5
	gc.vertices[15] = gc.ys1

	// Control point guide lines (simplified)
	gc.vertices[16] = gc.xs1
	gc.vertices[17] = gc.yp1 - gc.gridWidth*0.5
	gc.vertices[18] = gc.xp1 - gc.gridWidth*0.5
	gc.vertices[19] = gc.yp1 - gc.gridWidth*0.5
	gc.vertices[20] = gc.xp1 - gc.gridWidth*0.5
	gc.vertices[21] = gc.ys1
	gc.vertices[22] = gc.xp1 + gc.gridWidth*0.5
	gc.vertices[23] = gc.ys1

	gc.vertexCount = 12
}

// setupPoint1Path prepares the inactive control point.
func (gc *GammaCtrlImpl) setupPoint1Path() {
	gc.calcPoints()
	if gc.p1Active {
		gc.ellipse.Init(gc.xp2, gc.yp2, gc.pointSize, gc.pointSize, 32, false)
	} else {
		gc.ellipse.Init(gc.xp1, gc.yp1, gc.pointSize, gc.pointSize, 32, false)
	}
}

// setupPoint2Path prepares the active control point.
func (gc *GammaCtrlImpl) setupPoint2Path() {
	gc.calcPoints()
	if gc.p1Active {
		gc.ellipse.Init(gc.xp1, gc.yp1, gc.pointSize, gc.pointSize, 32, false)
	} else {
		gc.ellipse.Init(gc.xp2, gc.yp2, gc.pointSize, gc.pointSize, 32, false)
	}
}

// setupTextPath prepares the text display.
func (gc *GammaCtrlImpl) setupTextPath() {
	kx1, ky1, kx2, ky2 := gc.gammaSpline.GetValues()
	text := fmt.Sprintf("%5.3f %5.3f %5.3f %5.3f", kx1, ky1, kx2, ky2)

	gc.textRenderer.SetText(text)
	gc.textRenderer.SetSize(gc.textHeight)
	gc.textRenderer.SetPosition(gc.xt1+gc.borderWidth*2.0, (gc.yt1+gc.yt2)*0.5-gc.textHeight*0.5)

	gc.textStroke.SetWidth(gc.textThickness)
	gc.textStroke.SetLineCap(basics.RoundCap)
	gc.textStroke.SetLineJoin(basics.RoundJoin)
	gc.textStroke.Rewind(0)
}

// getPreCalculatedVertex returns the next pre-calculated vertex.
func (gc *GammaCtrlImpl) getPreCalculatedVertex() (x, y float64, cmd basics.PathCommand) {
	if gc.vertexIndex >= gc.vertexCount {
		return 0, 0, basics.PathCmdStop
	}

	cmd = basics.PathCmdLineTo
	if gc.vertexIndex == 0 {
		cmd = basics.PathCmdMoveTo
	}

	// Handle multi-path border case
	if gc.currentPath == 1 && (gc.vertexIndex == 4 || gc.vertexIndex == 8) {
		cmd = basics.PathCmdMoveTo
	}

	// Handle multi-path grid case
	if gc.currentPath == 3 && (gc.vertexIndex == 4 || gc.vertexIndex == 8) {
		cmd = basics.PathCmdMoveTo
	}

	x = gc.vertices[gc.vertexIndex*2]
	y = gc.vertices[gc.vertexIndex*2+1]
	gc.vertexIndex++

	// Apply transformation
	gc.TransformXY(&x, &y)
	return x, y, cmd
}

// GammaCtrl is a generic gamma control with customizable colors.
// This corresponds to AGG's templated gamma_ctrl class.
type GammaCtrl struct {
	*GammaCtrlImpl

	// Color scheme for the 7 rendering paths
	backgroundColor  color.RGBA
	borderColor      color.RGBA
	curveColor       color.RGBA
	gridColor        color.RGBA
	inactivePntColor color.RGBA
	activePntColor   color.RGBA
	textColor        color.RGBA

	// Color pointers for easy access
	colors [7]*color.RGBA
}

// NewGammaCtrl creates a new gamma control with default colors.
func NewGammaCtrl(x1, y1, x2, y2 float64, flipY bool) *GammaCtrl {
	ctrl := &GammaCtrl{
		GammaCtrlImpl:    NewGammaCtrlImpl(x1, y1, x2, y2, flipY),
		backgroundColor:  color.NewRGBA(1.0, 1.0, 0.9, 1.0), // Light yellow
		borderColor:      color.NewRGBA(0.0, 0.0, 0.0, 1.0), // Black
		curveColor:       color.NewRGBA(0.0, 0.0, 0.0, 1.0), // Black
		gridColor:        color.NewRGBA(0.2, 0.2, 0.0, 1.0), // Dark yellow
		inactivePntColor: color.NewRGBA(0.0, 0.0, 0.0, 1.0), // Black
		activePntColor:   color.NewRGBA(1.0, 0.0, 0.0, 1.0), // Red
		textColor:        color.NewRGBA(0.0, 0.0, 0.0, 1.0), // Black
	}

	// Set up color pointer array
	ctrl.colors[0] = &ctrl.backgroundColor
	ctrl.colors[1] = &ctrl.borderColor
	ctrl.colors[2] = &ctrl.curveColor
	ctrl.colors[3] = &ctrl.gridColor
	ctrl.colors[4] = &ctrl.inactivePntColor
	ctrl.colors[5] = &ctrl.activePntColor
	ctrl.colors[6] = &ctrl.textColor

	return ctrl
}

// Color setter methods
func (gc *GammaCtrl) SetBackgroundColor(c color.RGBA)  { gc.backgroundColor = c }
func (gc *GammaCtrl) SetBorderColor(c color.RGBA)      { gc.borderColor = c }
func (gc *GammaCtrl) SetCurveColor(c color.RGBA)       { gc.curveColor = c }
func (gc *GammaCtrl) SetGridColor(c color.RGBA)        { gc.gridColor = c }
func (gc *GammaCtrl) SetInactivePntColor(c color.RGBA) { gc.inactivePntColor = c }
func (gc *GammaCtrl) SetActivePntColor(c color.RGBA)   { gc.activePntColor = c }
func (gc *GammaCtrl) SetTextColor(c color.RGBA)        { gc.textColor = c }

// Color returns the color for a specific path.
func (gc *GammaCtrl) Color(pathID uint) interface{} {
	if pathID < 7 {
		return *gc.colors[pathID]
	}
	return color.NewRGBA(0, 0, 0, 1) // Default black
}
