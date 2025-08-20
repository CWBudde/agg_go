// Package spline provides an interactive spline curve editor control.
// This is a port of AGG's spline_ctrl_impl and spline_ctrl classes from agg_spline_ctrl.h/cpp.
package spline

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl"
	"agg_go/internal/curves"
	"agg_go/internal/path"
	"agg_go/internal/shapes"
)

const (
	// Maximum number of control points supported
	maxControlPoints = 32
	// Number of pre-calculated spline values for fast lookup
	splineValueCount = 256
	// Number of rendering paths
	numPaths = 5
)

// SplineCtrlImpl implements an interactive spline curve editor control.
// This corresponds to AGG's spline_ctrl_impl class.
//
// The control allows users to interactively modify a smooth curve by dragging
// control points. It's commonly used for gamma correction, color transfer functions,
// and other curve-based parameter adjustments.
type SplineCtrlImpl struct {
	*ctrl.BaseCtrl

	// Configuration
	numPnt      uint    // Number of control points (4-32)
	borderWidth float64 // Border thickness around control
	borderExtra float64 // Extra space outside border
	curveWidth  float64 // Width of the curve line
	pointSize   float64 // Size of control points

	// Control point data
	xp [maxControlPoints]float64 // X coordinates of control points (normalized 0-1)
	yp [maxControlPoints]float64 // Y coordinates of control points (normalized 0-1)

	// Spline computation
	spline        *curves.BSpline           // B-spline calculator
	splineValues  [splineValueCount]float64 // Pre-calculated spline values
	splineValues8 [splineValueCount]uint8   // 8-bit versions of spline values

	// Layout
	xs1, ys1, xs2, ys2 float64 // Inner spline area bounds

	// Rendering components
	curvePath *path.PathStorage // Path for the spline curve
	ellipse   *shapes.Ellipse   // For drawing control points

	// Vertex generation state
	currentPath uint        // Current path being rendered (0-4)
	vertexIndex uint        // Current vertex in path
	vertices    [32]float64 // Pre-calculated vertices for rectangular paths
	vertexCount uint        // Number of vertices in current path

	// Mouse interaction state
	activePnt int     // Index of currently selected control point (-1 if none)
	movePnt   int     // Index of control point being dragged (-1 if none)
	pdx, pdy  float64 // Offset from mouse to control point center during drag
}

// NewSplineCtrlImpl creates a new spline control implementation.
// x1, y1, x2, y2: bounding rectangle for the control
// numPnt: number of control points (clamped to 4-32 range)
// flipY: whether to flip Y coordinates for rendering
func NewSplineCtrlImpl(x1, y1, x2, y2 float64, numPnt uint, flipY bool) *SplineCtrlImpl {
	// Clamp number of points to valid range
	if numPnt < 4 {
		numPnt = 4
	}
	if numPnt > maxControlPoints {
		numPnt = maxControlPoints
	}

	ctrl := &SplineCtrlImpl{
		BaseCtrl:    ctrl.NewBaseCtrl(x1, y1, x2, y2, flipY),
		numPnt:      numPnt,
		borderWidth: 1.0,
		borderExtra: 0.0,
		curveWidth:  1.0,
		pointSize:   3.0,
		spline:      curves.NewBSpline(),
		curvePath:   path.NewPathStorage(),
		ellipse:     shapes.NewEllipse(),
		activePnt:   -1,
		movePnt:     -1,
		pdx:         0.0,
		pdy:         0.0,
	}

	// Initialize stroke converter for curve rendering
	// Note: We'll initialize this later when needed since we need to handle interface properly

	// Initialize control points with default values
	// Points are evenly spaced horizontally at y=0.5
	for i := uint(0); i < ctrl.numPnt; i++ {
		ctrl.xp[i] = float64(i) / float64(ctrl.numPnt-1)
		ctrl.yp[i] = 0.5
	}

	ctrl.calcSplineBox()
	ctrl.updateSpline()

	return ctrl
}

// calcSplineBox calculates the inner bounds of the spline area.
func (s *SplineCtrlImpl) calcSplineBox() {
	s.xs1 = s.X1() + s.borderWidth
	s.ys1 = s.Y1() + s.borderWidth
	s.xs2 = s.X2() - s.borderWidth
	s.ys2 = s.Y2() - s.borderWidth
}

// updateSpline recalculates the spline curve and lookup table.
func (s *SplineCtrlImpl) updateSpline() {
	// Create slices from the arrays for the spline initialization
	xPoints := make([]float64, s.numPnt)
	yPoints := make([]float64, s.numPnt)

	for i := uint(0); i < s.numPnt; i++ {
		xPoints[i] = s.xp[i]
		yPoints[i] = s.yp[i]
	}

	// Initialize the B-spline with control points
	s.spline.InitFromPoints(xPoints, yPoints)

	// Pre-calculate spline values for fast lookup
	for i := 0; i < splineValueCount; i++ {
		x := float64(i) / float64(splineValueCount-1)
		value := s.spline.Get(x)

		// Clamp value to [0, 1] range
		if value < 0.0 {
			value = 0.0
		}
		if value > 1.0 {
			value = 1.0
		}

		s.splineValues[i] = value
		s.splineValues8[i] = uint8(value * 255.0)
	}
}

// BorderWidth sets the border width and extra space around the control.
func (s *SplineCtrlImpl) BorderWidth(width, extra float64) {
	s.borderWidth = width
	s.borderExtra = extra
	s.calcSplineBox()
}

// CurveWidth sets the width of the curve line.
func (s *SplineCtrlImpl) CurveWidth(width float64) {
	s.curveWidth = width
}

// PointSize sets the size of the control points.
func (s *SplineCtrlImpl) PointSize(size float64) {
	s.pointSize = size
}

// ActivePoint sets the currently active control point.
func (s *SplineCtrlImpl) ActivePoint(index int) {
	if index >= -1 && index < int(s.numPnt) {
		s.activePnt = index
	}
}

// ActivePoint returns the index of the currently active control point.
func (s *SplineCtrlImpl) GetActivePoint() int {
	return s.activePnt
}

// Spline returns the pre-calculated spline values as a slice.
func (s *SplineCtrlImpl) Spline() []float64 {
	return s.splineValues[:]
}

// Spline8 returns the pre-calculated 8-bit spline values as a slice.
func (s *SplineCtrlImpl) Spline8() []uint8 {
	return s.splineValues8[:]
}

// Value returns the spline value for a given X coordinate.
func (s *SplineCtrlImpl) Value(x float64) float64 {
	value := s.spline.Get(x)
	if value < 0.0 {
		value = 0.0
	}
	if value > 1.0 {
		value = 1.0
	}
	return value
}

// SetValue sets the Y coordinate of a control point by index.
func (s *SplineCtrlImpl) SetValue(idx uint, y float64) {
	if idx < s.numPnt {
		s.setYP(idx, y)
		s.updateSpline()
	}
}

// SetPoint sets both X and Y coordinates of a control point by index.
func (s *SplineCtrlImpl) SetPoint(idx uint, x, y float64) {
	if idx < s.numPnt {
		s.setXP(idx, x)
		s.setYP(idx, y)
		s.updateSpline()
	}
}

// GetPointX returns the X coordinate of a control point.
func (s *SplineCtrlImpl) GetPointX(idx uint) float64 {
	if idx < s.numPnt {
		return s.xp[idx]
	}
	return 0.0
}

// GetPointY returns the Y coordinate of a control point.
func (s *SplineCtrlImpl) GetPointY(idx uint) float64 {
	if idx < s.numPnt {
		return s.yp[idx]
	}
	return 0.0
}

// setXP sets the X coordinate of a control point with constraints.
// The first and last points are constrained to x=0 and x=1 respectively.
// Interior points cannot cross their neighbors.
func (s *SplineCtrlImpl) setXP(idx uint, val float64) {
	// Clamp to [0, 1] range
	if val < 0.0 {
		val = 0.0
	}
	if val > 1.0 {
		val = 1.0
	}

	if idx == 0 {
		// First point is fixed at x=0
		val = 0.0
	} else if idx == s.numPnt-1 {
		// Last point is fixed at x=1
		val = 1.0
	} else {
		// Interior points cannot cross neighbors
		if val < s.xp[idx-1]+0.001 {
			val = s.xp[idx-1] + 0.001
		}
		if val > s.xp[idx+1]-0.001 {
			val = s.xp[idx+1] - 0.001
		}
	}

	s.xp[idx] = val
}

// setYP sets the Y coordinate of a control point with constraints.
func (s *SplineCtrlImpl) setYP(idx uint, val float64) {
	// Clamp to [0, 1] range
	if val < 0.0 {
		val = 0.0
	}
	if val > 1.0 {
		val = 1.0
	}
	s.yp[idx] = val
}

// calcXP converts normalized X coordinate to screen coordinates.
func (s *SplineCtrlImpl) calcXP(idx uint) float64 {
	return s.xs1 + (s.xs2-s.xs1)*s.xp[idx]
}

// calcYP converts normalized Y coordinate to screen coordinates.
func (s *SplineCtrlImpl) calcYP(idx uint) float64 {
	return s.ys1 + (s.ys2-s.ys1)*s.yp[idx]
}

// NumPaths returns the number of rendering paths (always 5).
func (s *SplineCtrlImpl) NumPaths() uint {
	return numPaths
}

// calcCurve generates the spline curve path for rendering.
func (s *SplineCtrlImpl) calcCurve() {
	s.curvePath.RemoveAll()

	// Start the curve at the first point
	x := s.xs1
	y := s.ys1 + (s.ys2-s.ys1)*s.splineValues[0]
	s.curvePath.MoveTo(x, y)

	// Add line segments for the rest of the curve
	for i := 1; i < splineValueCount; i++ {
		x = s.xs1 + (s.xs2-s.xs1)*float64(i)/float64(splineValueCount-1)
		y = s.ys1 + (s.ys2-s.ys1)*s.splineValues[i]
		s.curvePath.LineTo(x, y)
	}
}

// Rewind prepares a specific rendering path for vertex generation.
// The spline control has 5 paths:
// 0: Background rectangle
// 1: Border rectangle
// 2: Spline curve
// 3: Inactive control points
// 4: Active control point
func (s *SplineCtrlImpl) Rewind(pathID uint) {
	s.currentPath = pathID
	s.vertexIndex = 0
	s.vertexCount = 0

	switch pathID {
	case 0: // Background rectangle
		s.vertexCount = 4
		s.vertices[0] = s.X1() - s.borderExtra
		s.vertices[1] = s.Y1() - s.borderExtra
		s.vertices[2] = s.X2() + s.borderExtra
		s.vertices[3] = s.Y1() - s.borderExtra
		s.vertices[4] = s.X2() + s.borderExtra
		s.vertices[5] = s.Y2() + s.borderExtra
		s.vertices[6] = s.X1() - s.borderExtra
		s.vertices[7] = s.Y2() + s.borderExtra

	case 1: // Border rectangle (with hole)
		s.vertexCount = 8
		// Outer rectangle
		s.vertices[0] = s.X1()
		s.vertices[1] = s.Y1()
		s.vertices[2] = s.X2()
		s.vertices[3] = s.Y1()
		s.vertices[4] = s.X2()
		s.vertices[5] = s.Y2()
		s.vertices[6] = s.X1()
		s.vertices[7] = s.Y2()
		// Inner rectangle (hole)
		s.vertices[8] = s.X1() + s.borderWidth
		s.vertices[9] = s.Y1() + s.borderWidth
		s.vertices[10] = s.X1() + s.borderWidth
		s.vertices[11] = s.Y2() - s.borderWidth
		s.vertices[12] = s.X2() - s.borderWidth
		s.vertices[13] = s.Y2() - s.borderWidth
		s.vertices[14] = s.X2() - s.borderWidth
		s.vertices[15] = s.Y1() + s.borderWidth

	case 2: // Spline curve
		s.calcCurve()
		s.curvePath.Rewind(0)

	case 3: // Inactive control points
		if int(s.vertexIndex) < int(s.numPnt) && int(s.vertexIndex) != s.activePnt {
			cx := s.calcXP(s.vertexIndex)
			cy := s.calcYP(s.vertexIndex)
			s.ellipse.Init(cx, cy, s.pointSize, s.pointSize, 32, false)
			s.ellipse.Rewind(0)
		} else {
			s.vertexIndex = s.numPnt // No more points
		}

	case 4: // Active control point
		if s.activePnt >= 0 && s.activePnt < int(s.numPnt) {
			cx := s.calcXP(uint(s.activePnt))
			cy := s.calcYP(uint(s.activePnt))
			s.ellipse.Init(cx, cy, s.pointSize, s.pointSize, 32, false)
			s.ellipse.Rewind(0)
		}
	}
}

// Vertex returns the next vertex in the current rendering path.
func (s *SplineCtrlImpl) Vertex() (x, y float64, cmd basics.PathCommand) {
	var cmdUint uint32

	switch s.currentPath {
	case 0: // Background rectangle
		if s.vertexIndex == 0 {
			cmd = basics.PathCmdMoveTo
		} else if s.vertexIndex >= s.vertexCount {
			cmd = basics.PathCmdStop
		} else {
			cmd = basics.PathCmdLineTo
		}

		if s.vertexIndex < s.vertexCount {
			x = s.vertices[s.vertexIndex*2]
			y = s.vertices[s.vertexIndex*2+1]
			s.vertexIndex++
		}

	case 1: // Border rectangle
		if s.vertexIndex == 0 || s.vertexIndex == 4 {
			cmd = basics.PathCmdMoveTo
		} else if s.vertexIndex >= s.vertexCount {
			cmd = basics.PathCmdStop
		} else {
			cmd = basics.PathCmdLineTo
		}

		if s.vertexIndex < s.vertexCount {
			x = s.vertices[s.vertexIndex*2]
			y = s.vertices[s.vertexIndex*2+1]
			s.vertexIndex++
		}

	case 2: // Spline curve
		x, y, cmdUint = s.curvePath.NextVertex()
		cmd = basics.PathCommand(cmdUint)

	case 3: // Inactive control points
		if int(s.vertexIndex) < int(s.numPnt) {
			// Find next inactive point
			for int(s.vertexIndex) < int(s.numPnt) && int(s.vertexIndex) == s.activePnt {
				s.vertexIndex++
			}

			if int(s.vertexIndex) < int(s.numPnt) {
				cx := s.calcXP(s.vertexIndex)
				cy := s.calcYP(s.vertexIndex)
				s.ellipse.Init(cx, cy, s.pointSize, s.pointSize, 32, false)
				s.ellipse.Rewind(0)
				s.vertexIndex++
			}
		}

		if int(s.vertexIndex) <= int(s.numPnt) {
			cmd = s.ellipse.Vertex(&x, &y)
		} else {
			cmd = basics.PathCmdStop
		}

	case 4: // Active control point
		if s.activePnt >= 0 && s.activePnt < int(s.numPnt) {
			cmd = s.ellipse.Vertex(&x, &y)
		} else {
			cmd = basics.PathCmdStop
		}

	default:
		cmd = basics.PathCmdStop
	}

	// Apply transformation if the command is not stop
	if !basics.IsStop(cmd) {
		s.TransformXY(&x, &y)
	}

	return x, y, cmd
}

// calcDistance calculates the Euclidean distance between two points.
func calcDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// InRect checks if a point is within the control's bounds.
func (s *SplineCtrlImpl) InRect(x, y float64) bool {
	// Apply inverse transformation to convert screen coordinates to control coordinates
	s.InverseTransformXY(&x, &y)
	return x >= s.X1() && x <= s.X2() && y >= s.Y1() && y <= s.Y2()
}

// OnMouseButtonDown handles mouse press events.
// Returns true if the control needs to be redrawn.
func (s *SplineCtrlImpl) OnMouseButtonDown(x, y float64) bool {
	// Convert screen coordinates to control coordinates
	s.InverseTransformXY(&x, &y)

	// Check each control point for hit detection
	for i := uint(0); i < s.numPnt; i++ {
		xp := s.calcXP(i)
		yp := s.calcYP(i)

		// Check if mouse is within point size + 1 pixel tolerance
		if calcDistance(x, y, xp, yp) <= s.pointSize+1.0 {
			// Calculate drag offset from point center to mouse position
			s.pdx = xp - x
			s.pdy = yp - y

			// Set this point as both active and being moved
			s.activePnt = int(i)
			s.movePnt = int(i)

			return true // Redraw needed
		}
	}

	return false // No hit, no redraw needed
}

// OnMouseButtonUp handles mouse release events.
// Returns true if the control needs to be redrawn.
func (s *SplineCtrlImpl) OnMouseButtonUp(x, y float64) bool {
	if s.movePnt >= 0 {
		s.movePnt = -1
		return true // Redraw needed to show final position
	}
	return false
}

// OnMouseMove handles mouse movement events.
// buttonPressed indicates if the mouse button is still pressed.
// Returns true if the control needs to be redrawn.
func (s *SplineCtrlImpl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	// Convert screen coordinates to control coordinates
	s.InverseTransformXY(&x, &y)

	if !buttonPressed {
		// If button is released, treat as mouse up
		return s.OnMouseButtonUp(x, y)
	}

	if s.movePnt >= 0 {
		// Calculate new position with drag offset
		xp := x + s.pdx
		yp := y + s.pdy

		// Convert screen coordinates to normalized coordinates
		normalizedX := (xp - s.xs1) / (s.xs2 - s.xs1)
		normalizedY := (yp - s.ys1) / (s.ys2 - s.ys1)

		// Update the control point position with constraints
		s.setXP(uint(s.movePnt), normalizedX)
		s.setYP(uint(s.movePnt), normalizedY)

		// Recalculate the spline curve
		s.updateSpline()

		return true // Redraw needed
	}

	return false
}

// OnArrowKeys handles arrow key events for precise control point adjustment.
// Returns true if the control needs to be redrawn.
func (s *SplineCtrlImpl) OnArrowKeys(left, right, down, up bool) bool {
	if s.activePnt < 0 || s.activePnt >= int(s.numPnt) {
		return false // No active point to move
	}

	kx := s.xp[s.activePnt]
	ky := s.yp[s.activePnt]
	changed := false

	// Small increment for precise adjustment
	const increment = 0.001

	if left {
		kx -= increment
		changed = true
	}
	if right {
		kx += increment
		changed = true
	}
	if down {
		ky -= increment
		changed = true
	}
	if up {
		ky += increment
		changed = true
	}

	if changed {
		// Apply constraints and update the point
		s.setXP(uint(s.activePnt), kx)
		s.setYP(uint(s.activePnt), ky)

		// Recalculate the spline curve
		s.updateSpline()

		return true // Redraw needed
	}

	return false
}

// Color returns the color for a specific rendering path.
// This is a placeholder that returns nil - actual colors should be managed
// by the SplineCtrl wrapper.
func (s *SplineCtrlImpl) Color(pathID uint) interface{} {
	return nil
}

// SplineCtrl is a generic wrapper around SplineCtrlImpl that provides color management.
// This corresponds to AGG's spline_ctrl template class.
type SplineCtrl[ColorT any] struct {
	*SplineCtrlImpl

	// Colors for the 5 rendering paths
	backgroundColor    ColorT
	borderColor        ColorT
	curveColor         ColorT
	inactivePointColor ColorT
	activePointColor   ColorT

	// Color array for easy indexing
	colors [5]*ColorT
}

// NewSplineCtrl creates a new spline control with color management.
// x1, y1, x2, y2: bounding rectangle for the control
// numPnt: number of control points (clamped to 4-32 range)
// flipY: whether to flip Y coordinates for rendering
func NewSplineCtrl[ColorT any](x1, y1, x2, y2 float64, numPnt uint, flipY bool) *SplineCtrl[ColorT] {
	impl := NewSplineCtrlImpl(x1, y1, x2, y2, numPnt, flipY)

	ctrl := &SplineCtrl[ColorT]{
		SplineCtrlImpl: impl,
	}

	// Set up color array for easy access
	ctrl.colors[0] = &ctrl.backgroundColor
	ctrl.colors[1] = &ctrl.borderColor
	ctrl.colors[2] = &ctrl.curveColor
	ctrl.colors[3] = &ctrl.inactivePointColor
	ctrl.colors[4] = &ctrl.activePointColor

	return ctrl
}

// SetBackgroundColor sets the background color for path 0.
func (s *SplineCtrl[ColorT]) SetBackgroundColor(c ColorT) {
	s.backgroundColor = c
}

// SetBorderColor sets the border color for path 1.
func (s *SplineCtrl[ColorT]) SetBorderColor(c ColorT) {
	s.borderColor = c
}

// SetCurveColor sets the curve color for path 2.
func (s *SplineCtrl[ColorT]) SetCurveColor(c ColorT) {
	s.curveColor = c
}

// SetInactivePointColor sets the color for inactive control points (path 3).
func (s *SplineCtrl[ColorT]) SetInactivePointColor(c ColorT) {
	s.inactivePointColor = c
}

// SetActivePointColor sets the color for the active control point (path 4).
func (s *SplineCtrl[ColorT]) SetActivePointColor(c ColorT) {
	s.activePointColor = c
}

// Color returns the color for a specific rendering path.
func (s *SplineCtrl[ColorT]) Color(pathID uint) interface{} {
	if pathID < uint(len(s.colors)) {
		return *s.colors[pathID]
	}
	return s.backgroundColor
}

// Convenience constructors for common color types

// NewSplineCtrlRGBA creates a spline control with RGBA color management.
func NewSplineCtrlRGBA(x1, y1, x2, y2 float64, numPnt uint, flipY bool) *SplineCtrl[color.RGBA] {
	ctrl := NewSplineCtrl[color.RGBA](x1, y1, x2, y2, numPnt, flipY)

	// Set default colors similar to AGG's defaults
	ctrl.SetBackgroundColor(color.NewRGBA(1.0, 1.0, 0.9, 1.0))    // Light yellow background
	ctrl.SetBorderColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0))        // Black border
	ctrl.SetCurveColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0))         // Black curve
	ctrl.SetInactivePointColor(color.NewRGBA(0.0, 0.0, 0.0, 1.0)) // Black inactive points
	ctrl.SetActivePointColor(color.NewRGBA(1.0, 0.0, 0.0, 1.0))   // Red active point

	return ctrl
}
