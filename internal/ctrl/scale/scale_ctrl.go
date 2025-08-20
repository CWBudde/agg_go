// Package scale provides a scale/range control implementation for AGG.
// This is a port of AGG's scale_ctrl_impl and scale_ctrl classes.
package scale

import (
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl"
	"agg_go/internal/shapes"
)

// MoveType represents what is being moved during mouse interaction
type MoveType int

const (
	MoveNothing MoveType = iota
	MoveValue1           // Moving the first value pointer
	MoveValue2           // Moving the second value pointer
	MoveSlider           // Moving the entire slider range
)

// ScaleCtrl implements a two-value range control (like a range slider).
// This corresponds to AGG's scale_ctrl_impl class.
type ScaleCtrl struct {
	*ctrl.BaseCtrl

	// Appearance settings
	borderThickness float64
	borderExtra     float64

	// Value settings - both values are normalized to [0, 1]
	value1 float64 // First value (typically the minimum of the range)
	value2 float64 // Second value (typically the maximum of the range)
	minD   float64 // Minimum distance between value1 and value2

	// Internal layout - inner slider bounds
	xs1, ys1, xs2, ys2 float64

	// Mouse interaction state
	moveWhat MoveType // What is currently being moved
	pdx, pdy float64  // Pointer deltas for mouse dragging

	// Pre-calculated vertex arrays for rendering
	vx [32]float64
	vy [32]float64

	// Rendering components
	ellipse *shapes.Ellipse

	// Vertex generation state
	idx    uint // Current path index being generated
	vertex uint // Current vertex index within path

	// Colors for the 5 rendering paths
	colors [5]color.RGBA
}

// NewScaleCtrl creates a new scale control.
// x1, y1, x2, y2: bounding rectangle
// flipY: whether to flip Y coordinates
func NewScaleCtrl(x1, y1, x2, y2 float64, flipY bool) *ScaleCtrl {
	scale := &ScaleCtrl{
		BaseCtrl:        ctrl.NewBaseCtrl(x1, y1, x2, y2, flipY),
		borderThickness: 1.0,
		borderExtra:     0.0,
		value1:          0.3, // Default range: 30% to 70%
		value2:          0.7,
		minD:            0.01, // Minimum 1% distance between values
		moveWhat:        MoveNothing,
		pdx:             0.0,
		pdy:             0.0,
		ellipse:         shapes.NewEllipse(),
		idx:             0,
		vertex:          0,
	}

	// Set border extra based on control orientation
	if math.Abs(x2-x1) > math.Abs(y2-y1) {
		// Horizontal orientation
		scale.borderExtra = (y2 - y1) / 2.0
	} else {
		// Vertical orientation
		scale.borderExtra = (x2 - x1) / 2.0
	}

	// Initialize default colors matching C++ AGG defaults
	scale.colors[0] = color.NewRGBA(1.0, 0.9, 0.8, 1.0) // Background - light beige
	scale.colors[1] = color.NewRGBA(0.0, 0.0, 0.0, 1.0) // Border - black
	scale.colors[2] = color.NewRGBA(0.8, 0.0, 0.0, 0.8) // Pointer1 - red
	scale.colors[3] = color.NewRGBA(0.8, 0.0, 0.0, 0.8) // Pointer2 - red
	scale.colors[4] = color.NewRGBA(0.2, 0.1, 0.0, 0.6) // Slider - dark brown

	scale.calcBox()
	return scale
}

// calcBox calculates the inner slider bounds based on border settings.
func (s *ScaleCtrl) calcBox() {
	s.xs1 = s.X1() + s.borderThickness
	s.ys1 = s.Y1() + s.borderThickness
	s.xs2 = s.X2() - s.borderThickness
	s.ys2 = s.Y2() - s.borderThickness
}

// BorderThickness sets the border width and optional extra margin.
func (s *ScaleCtrl) BorderThickness(t, extra float64) {
	s.borderThickness = t
	s.borderExtra = extra
	s.calcBox()
}

// Resize updates the control's bounding rectangle.
func (s *ScaleCtrl) Resize(x1, y1, x2, y2 float64) {
	s.SetBounds(x1, y1, x2, y2)
	s.calcBox()

	// Recalculate border extra based on new orientation
	if math.Abs(x2-x1) > math.Abs(y2-y1) {
		s.borderExtra = (y2 - y1) / 2.0
	} else {
		s.borderExtra = (x2 - x1) / 2.0
	}
}

// MinDelta returns the minimum distance between values.
func (s *ScaleCtrl) MinDelta() float64 {
	return s.minD
}

// SetMinDelta sets the minimum distance between values.
func (s *ScaleCtrl) SetMinDelta(d float64) {
	s.minD = d
}

// Value1 returns the first value (typically minimum of range).
func (s *ScaleCtrl) Value1() float64 {
	return s.value1
}

// SetValue1 sets the first value with validation.
func (s *ScaleCtrl) SetValue1(value float64) {
	if value < 0.0 {
		value = 0.0
	}
	if value > 1.0 {
		value = 1.0
	}
	if s.value2-value < s.minD {
		value = s.value2 - s.minD
	}
	if value < 0.0 {
		value = 0.0
	}
	s.value1 = value
}

// Value2 returns the second value (typically maximum of range).
func (s *ScaleCtrl) Value2() float64 {
	return s.value2
}

// SetValue2 sets the second value with validation.
func (s *ScaleCtrl) SetValue2(value float64) {
	if value < 0.0 {
		value = 0.0
	}
	if value > 1.0 {
		value = 1.0
	}
	if value-s.value1 < s.minD {
		value = s.value1 + s.minD
	}
	s.value2 = value
}

// Move shifts both values by the specified delta, maintaining the range.
func (s *ScaleCtrl) Move(d float64) {
	s.value1 += d
	s.value2 += d

	if s.value1 < 0.0 {
		s.value2 -= s.value1
		s.value1 = 0.0
	}
	if s.value2 > 1.0 {
		s.value1 -= s.value2 - 1.0
		s.value2 = 1.0
	}
}

// Color management methods

// BackgroundColor sets the background color.
func (s *ScaleCtrl) BackgroundColor(c color.RGBA) {
	s.colors[0] = c
}

// BorderColor sets the border color.
func (s *ScaleCtrl) BorderColor(c color.RGBA) {
	s.colors[1] = c
}

// PointersColor sets the color for both pointer circles.
func (s *ScaleCtrl) PointersColor(c color.RGBA) {
	s.colors[2] = c
	s.colors[3] = c
}

// SliderColor sets the color for the slider bar.
func (s *ScaleCtrl) SliderColor(c color.RGBA) {
	s.colors[4] = c
}

// Color returns the color for the specified path.
func (s *ScaleCtrl) Color(pathID uint) interface{} {
	if pathID < uint(len(s.colors)) {
		return s.colors[pathID]
	}
	return color.NewRGBA(0, 0, 0, 1) // Default black
}

// Mouse interaction methods

// InRect checks if a point is within the control's bounds.
func (s *ScaleCtrl) InRect(x, y float64) bool {
	s.InverseTransformXY(&x, &y)
	return x >= s.X1() && x <= s.X2() && y >= s.Y1() && y <= s.Y2()
}

// OnMouseButtonDown handles mouse button press events.
func (s *ScaleCtrl) OnMouseButtonDown(x, y float64) bool {
	s.InverseTransformXY(&x, &y)

	var xp1, xp2, ys1, ys2, xp, yp float64

	if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
		// Horizontal orientation
		xp1 = s.xs1 + (s.xs2-s.xs1)*s.value1
		xp2 = s.xs1 + (s.xs2-s.xs1)*s.value2
		ys1 = s.Y1() - s.borderExtra/2.0
		ys2 = s.Y2() + s.borderExtra/2.0
		yp = (s.ys1 + s.ys2) / 2.0

		// Check if clicking on the slider bar (between pointers)
		if x > xp1 && y > ys1 && x < xp2 && y < ys2 {
			s.pdx = xp1 - x
			s.moveWhat = MoveSlider
			return true
		}

		// Check if clicking on pointer 1
		if basics.CalcDistance(x, y, xp1, yp) <= s.Y2()-s.Y1() {
			s.pdx = xp1 - x
			s.moveWhat = MoveValue1
			return true
		}

		// Check if clicking on pointer 2
		if basics.CalcDistance(x, y, xp2, yp) <= s.Y2()-s.Y1() {
			s.pdx = xp2 - x
			s.moveWhat = MoveValue2
			return true
		}
	} else {
		// Vertical orientation
		xp1 = s.X1() - s.borderExtra/2.0
		xp2 = s.X2() + s.borderExtra/2.0
		ys1 = s.ys1 + (s.ys2-s.ys1)*s.value1
		ys2 = s.ys1 + (s.ys2-s.ys1)*s.value2
		xp = (s.xs1 + s.xs2) / 2.0

		// Check if clicking on the slider bar (between pointers)
		if x > xp1 && y > ys1 && x < xp2 && y < ys2 {
			s.pdy = ys1 - y
			s.moveWhat = MoveSlider
			return true
		}

		// Check if clicking on pointer 1
		if basics.CalcDistance(x, y, xp, ys1) <= s.X2()-s.X1() {
			s.pdy = ys1 - y
			s.moveWhat = MoveValue1
			return true
		}

		// Check if clicking on pointer 2
		if basics.CalcDistance(x, y, xp, ys2) <= s.X2()-s.X1() {
			s.pdy = ys2 - y
			s.moveWhat = MoveValue2
			return true
		}
	}

	return false
}

// OnMouseButtonUp handles mouse button release events.
func (s *ScaleCtrl) OnMouseButtonUp(x, y float64) bool {
	s.moveWhat = MoveNothing
	return false
}

// OnMouseMove handles mouse movement events.
func (s *ScaleCtrl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	s.InverseTransformXY(&x, &y)

	if !buttonPressed {
		return s.OnMouseButtonUp(x, y)
	}

	xp := x + s.pdx
	yp := y + s.pdy
	var dv float64

	switch s.moveWhat {
	case MoveValue1:
		if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
			// Horizontal orientation
			s.value1 = (xp - s.xs1) / (s.xs2 - s.xs1)
		} else {
			// Vertical orientation
			s.value1 = (yp - s.ys1) / (s.ys2 - s.ys1)
		}
		if s.value1 < 0.0 {
			s.value1 = 0.0
		}
		if s.value1 > s.value2-s.minD {
			s.value1 = s.value2 - s.minD
		}
		return true

	case MoveValue2:
		if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
			// Horizontal orientation
			s.value2 = (xp - s.xs1) / (s.xs2 - s.xs1)
		} else {
			// Vertical orientation
			s.value2 = (yp - s.ys1) / (s.ys2 - s.ys1)
		}
		if s.value2 > 1.0 {
			s.value2 = 1.0
		}
		if s.value2 < s.value1+s.minD {
			s.value2 = s.value1 + s.minD
		}
		return true

	case MoveSlider:
		dv = s.value2 - s.value1
		if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
			// Horizontal orientation
			s.value1 = (xp - s.xs1) / (s.xs2 - s.xs1)
		} else {
			// Vertical orientation
			s.value1 = (yp - s.ys1) / (s.ys2 - s.ys1)
		}
		s.value2 = s.value1 + dv

		if s.value1 < 0.0 {
			dv = s.value2 - s.value1
			s.value1 = 0.0
			s.value2 = s.value1 + dv
		}
		if s.value2 > 1.0 {
			dv = s.value2 - s.value1
			s.value2 = 1.0
			s.value1 = s.value2 - dv
		}
		return true

	default:
		return false
	}
}

// OnArrowKeys handles arrow key events (currently not implemented).
func (s *ScaleCtrl) OnArrowKeys(left, right, down, up bool) bool {
	// Arrow key support could be added here for keyboard navigation
	return false
}

// Vertex source interface methods

// NumPaths returns the number of rendering paths.
func (s *ScaleCtrl) NumPaths() uint {
	return 5
}

// Rewind initializes vertex generation for the specified path.
func (s *ScaleCtrl) Rewind(pathID uint) {
	s.idx = pathID

	switch pathID {
	case 0: // Background
		s.vertex = 0
		s.vx[0] = s.X1() - s.borderExtra
		s.vy[0] = s.Y1() - s.borderExtra
		s.vx[1] = s.X2() + s.borderExtra
		s.vy[1] = s.Y1() - s.borderExtra
		s.vx[2] = s.X2() + s.borderExtra
		s.vy[2] = s.Y2() + s.borderExtra
		s.vx[3] = s.X1() - s.borderExtra
		s.vy[3] = s.Y2() + s.borderExtra

	case 1: // Border
		s.vertex = 0
		s.vx[0] = s.X1()
		s.vy[0] = s.Y1()
		s.vx[1] = s.X2()
		s.vy[1] = s.Y1()
		s.vx[2] = s.X2()
		s.vy[2] = s.Y2()
		s.vx[3] = s.X1()
		s.vy[3] = s.Y2()
		s.vx[4] = s.X1() + s.borderThickness
		s.vy[4] = s.Y1() + s.borderThickness
		s.vx[5] = s.X1() + s.borderThickness
		s.vy[5] = s.Y2() - s.borderThickness
		s.vx[6] = s.X2() - s.borderThickness
		s.vy[6] = s.Y2() - s.borderThickness
		s.vx[7] = s.X2() - s.borderThickness
		s.vy[7] = s.Y1() + s.borderThickness

	case 2: // Pointer 1
		s.setupPointerEllipse(s.value1)

	case 3: // Pointer 2
		s.setupPointerEllipse(s.value2)

	case 4: // Slider
		s.vertex = 0
		if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
			// Horizontal orientation
			s.vx[0] = s.xs1 + (s.xs2-s.xs1)*s.value1
			s.vy[0] = s.Y1() - s.borderExtra/2.0
			s.vx[1] = s.xs1 + (s.xs2-s.xs1)*s.value2
			s.vy[1] = s.vy[0]
			s.vx[2] = s.vx[1]
			s.vy[2] = s.Y2() + s.borderExtra/2.0
			s.vx[3] = s.vx[0]
			s.vy[3] = s.vy[2]
		} else {
			// Vertical orientation
			s.vx[0] = s.X1() - s.borderExtra/2.0
			s.vy[0] = s.ys1 + (s.ys2-s.ys1)*s.value1
			s.vx[1] = s.vx[0]
			s.vy[1] = s.ys1 + (s.ys2-s.ys1)*s.value2
			s.vx[2] = s.X2() + s.borderExtra/2.0
			s.vy[2] = s.vy[1]
			s.vx[3] = s.vx[2]
			s.vy[3] = s.vy[0]
		}
	}
}

// setupPointerEllipse configures the ellipse for a pointer at the given value.
func (s *ScaleCtrl) setupPointerEllipse(value float64) {
	if math.Abs(s.X2()-s.X1()) > math.Abs(s.Y2()-s.Y1()) {
		// Horizontal orientation
		centerX := s.xs1 + (s.xs2-s.xs1)*value
		centerY := (s.ys1 + s.ys2) / 2.0
		radius := s.Y2() - s.Y1()
		s.ellipse.Init(centerX, centerY, radius, radius, 32, false)
	} else {
		// Vertical orientation
		centerX := (s.xs1 + s.xs2) / 2.0
		centerY := s.ys1 + (s.ys2-s.ys1)*value
		radius := s.X2() - s.X1()
		s.ellipse.Init(centerX, centerY, radius, radius, 32, false)
	}
	s.ellipse.Rewind(0)
}

// Vertex returns the next vertex for the current path.
func (s *ScaleCtrl) Vertex() (x, y float64, cmd basics.PathCommand) {
	cmd = basics.PathCmdLineTo

	switch s.idx {
	case 0, 4: // Background and Slider (rectangles)
		if s.vertex == 0 {
			cmd = basics.PathCmdMoveTo
		}
		if s.vertex >= 4 {
			cmd = basics.PathCmdStop
		} else {
			x = s.vx[s.vertex]
			y = s.vy[s.vertex]
			s.vertex++
		}

	case 1: // Border (double rectangle)
		if s.vertex == 0 || s.vertex == 4 {
			cmd = basics.PathCmdMoveTo
		}
		if s.vertex >= 8 {
			cmd = basics.PathCmdStop
		} else {
			x = s.vx[s.vertex]
			y = s.vy[s.vertex]
			s.vertex++
		}

	case 2, 3: // Pointers (ellipses)
		cmd = s.ellipse.Vertex(&x, &y)

	default:
		cmd = basics.PathCmdStop
	}

	if !basics.IsStop(cmd) {
		s.TransformXY(&x, &y)
	}

	return x, y, cmd
}
