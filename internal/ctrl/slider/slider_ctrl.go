// Package slider provides a slider control implementation for AGG.
// This is a port of AGG's slider_ctrl_impl and slider_ctrl classes.
package slider

import (
	"fmt"
	"math"

	"agg_go/internal/basics"
	"agg_go/internal/color"
	"agg_go/internal/ctrl"
	"agg_go/internal/ctrl/text"
	"agg_go/internal/path"
	"agg_go/internal/shapes"
)

// SliderCtrl implements a horizontal or vertical slider control.
// This corresponds to AGG's slider_ctrl_impl class.
type SliderCtrl struct {
	*ctrl.BaseCtrl

	// Appearance settings
	borderWidth   float64
	borderExtra   float64
	textThickness float64

	// Value settings
	value        float64 // Current normalized value [0, 1]
	previewValue float64 // Preview value during mouse interaction
	minValue     float64 // Minimum slider value
	maxValue     float64 // Maximum slider value
	numSteps     uint    // Number of discrete steps (0 = continuous)
	descending   bool    // True if slider decreases from left to right

	// Label settings
	label string

	// Internal layout
	xs1, ys1, xs2, ys2 float64 // Inner slider bounds

	// Mouse interaction
	mouseMove bool
	pointerDx float64 // Pointer drag delta from mouse to pointer center

	// Rendering components
	ellipse      *shapes.Ellipse
	textRenderer *text.SimpleText
	pathStorage  *path.PathStorage

	// Vertex generation state
	currentPath uint
	vertexIndex uint
	vertices    [32]float64 // Pre-calculated vertices for current path
	vertexCount uint

	// Colors for the 6 rendering paths
	colors [6]color.RGBA
}

// NewSliderCtrl creates a new slider control.
// x1, y1, x2, y2: bounding rectangle
// flipY: whether to flip Y coordinates
func NewSliderCtrl(x1, y1, x2, y2 float64, flipY bool) *SliderCtrl {
	slider := &SliderCtrl{
		BaseCtrl:      ctrl.NewBaseCtrl(x1, y1, x2, y2, flipY),
		borderWidth:   1.0,
		borderExtra:   (y2 - y1) / 2.0,
		textThickness: 1.0,
		value:         0.5,
		previewValue:  0.5,
		minValue:      0.0,
		maxValue:      1.0,
		numSteps:      0,
		descending:    false,
		label:         "",
		mouseMove:     false,
		pointerDx:     0.0,
		ellipse:       shapes.NewEllipse(),
		textRenderer:  text.NewSimpleText(),
		pathStorage:   path.NewPathStorage(),
		currentPath:   0,
		vertexIndex:   0,
		vertexCount:   0,
	}

	// Initialize default colors matching C++ AGG defaults
	slider.colors[0] = color.NewRGBA(1.0, 0.9, 0.8, 1.0) // Background - light beige
	slider.colors[1] = color.NewRGBA(0.7, 0.6, 0.6, 1.0) // Triangle - gray
	slider.colors[2] = color.NewRGBA(0.0, 0.0, 0.0, 1.0) // Text - black
	slider.colors[3] = color.NewRGBA(0.6, 0.4, 0.4, 0.4) // Pointer preview - translucent red
	slider.colors[4] = color.NewRGBA(0.8, 0.0, 0.0, 0.6) // Pointer - red
	slider.colors[5] = color.NewRGBA(0.0, 0.0, 0.0, 1.0) // Text (duplicate) - black

	slider.calcBox()
	return slider
}

// calcBox calculates the inner slider bounds based on border settings.
func (s *SliderCtrl) calcBox() {
	s.xs1 = s.X1() + s.borderWidth
	s.ys1 = s.Y1() + s.borderWidth
	s.xs2 = s.X2() - s.borderWidth
	s.ys2 = s.Y2() - s.borderWidth
}

// SetBorderWidth sets the border width and extra border space.
func (s *SliderCtrl) SetBorderWidth(width, extra float64) {
	s.borderWidth = width
	s.borderExtra = extra
	s.calcBox()
}

// SetRange sets the value range of the slider.
func (s *SliderCtrl) SetRange(min, max float64) {
	s.minValue = min
	s.maxValue = max
}

// SetNumSteps sets the number of discrete steps (0 for continuous).
func (s *SliderCtrl) SetNumSteps(steps uint) {
	s.numSteps = steps
}

// SetLabel sets the label format string (printf-style).
func (s *SliderCtrl) SetLabel(format string) {
	s.label = format
}

// SetTextThickness sets the thickness of text rendering.
func (s *SliderCtrl) SetTextThickness(thickness float64) {
	s.textThickness = thickness
	s.textRenderer.SetThickness(thickness)
}

// SetDescending sets whether the slider decreases from left to right.
func (s *SliderCtrl) SetDescending(descending bool) {
	s.descending = descending
}

// Value returns the current slider value in the specified range.
func (s *SliderCtrl) Value() float64 {
	return s.value*(s.maxValue-s.minValue) + s.minValue
}

// SetValue sets the current slider value.
func (s *SliderCtrl) SetValue(value float64) {
	s.previewValue = (value - s.minValue) / (s.maxValue - s.minValue)
	if s.previewValue > 1.0 {
		s.previewValue = 1.0
	}
	if s.previewValue < 0.0 {
		s.previewValue = 0.0
	}
	s.normalizeValue(true)
}

// normalizeValue applies step quantization and updates the actual value.
func (s *SliderCtrl) normalizeValue(previewFlag bool) bool {
	ret := true
	if s.numSteps > 0 {
		step := uint(s.previewValue*float64(s.numSteps) + 0.5)
		newValue := float64(step) / float64(s.numSteps)
		ret = s.value != newValue
		s.value = newValue
	} else {
		s.value = s.previewValue
	}

	if previewFlag {
		s.previewValue = s.value
	}
	return ret
}

// Event handling methods implementing the Ctrl interface

func (s *SliderCtrl) InRect(x, y float64) bool {
	// Apply inverse transformation
	s.InverseTransformXY(&x, &y)

	// Check if point is within slider bounds (matching C++ AGG behavior)
	return x >= s.X1() && x <= s.X2() && y >= s.Y1() && y <= s.Y2()
}

func (s *SliderCtrl) OnMouseButtonDown(x, y float64) bool {
	// Apply inverse transformation
	s.InverseTransformXY(&x, &y)

	// Calculate current pointer position
	pointerX := s.xs1 + (s.xs2-s.xs1)*s.value
	pointerY := (s.ys1 + s.ys2) / 2.0

	// Check if click is on the pointer (within radius)
	pointerRadius := s.Y2() - s.Y1()
	distance := math.Sqrt((x-pointerX)*(x-pointerX) + (y-pointerY)*(y-pointerY))

	if distance <= pointerRadius {
		// Store the delta between mouse position and pointer center
		s.pointerDx = pointerX - x
		s.mouseMove = true
		return true
	}
	return false
}

func (s *SliderCtrl) OnMouseButtonUp(x, y float64) bool {
	_ = x // Unused in AGG implementation
	_ = y // Unused in AGG implementation

	if s.mouseMove {
		s.mouseMove = false
		s.normalizeValue(true)
		return true
	}
	return false
}

func (s *SliderCtrl) OnMouseMove(x, y float64, buttonPressed bool) bool {
	// Apply inverse transformation
	s.InverseTransformXY(&x, &y)

	if !buttonPressed {
		s.OnMouseButtonUp(x, y)
		return false
	}

	if s.mouseMove {
		// Calculate new pointer position with delta correction
		pointerX := x + s.pointerDx
		normalizedPos := (pointerX - s.xs1) / (s.xs2 - s.xs1)

		// Clamp to valid range
		if normalizedPos < 0.0 {
			normalizedPos = 0.0
		}
		if normalizedPos > 1.0 {
			normalizedPos = 1.0
		}

		s.previewValue = normalizedPos
		return true
	}
	return false
}

func (s *SliderCtrl) OnArrowKeys(left, right, down, up bool) bool {
	step := 0.005 // Default small step like in C++ AGG
	if s.numSteps > 0 {
		step = 1.0 / float64(s.numSteps)
	}

	// AGG treats right/up as increment, left/down as decrement
	if right || up {
		s.previewValue += step
		if s.previewValue > 1.0 {
			s.previewValue = 1.0
		}
		s.normalizeValue(true)
		return true
	}

	if left || down {
		s.previewValue -= step
		if s.previewValue < 0.0 {
			s.previewValue = 0.0
		}
		s.normalizeValue(true)
		return true
	}

	return false
}

// Vertex source interface implementation

func (s *SliderCtrl) NumPaths() uint {
	return 6 // Background, Triangle, Text, Pointer Preview, Pointer, Steps
}

func (s *SliderCtrl) Rewind(pathID uint) {
	s.currentPath = pathID
	s.vertexIndex = 0
	s.vertexCount = 0

	switch pathID {
	case 0: // Background
		s.generateBackgroundPath()
	case 1: // Triangle
		s.generateTrianglePath()
	case 2: // Text
		s.generateTextPath()
	case 3: // Pointer preview
		s.generatePointerPreviewPath()
	case 4: // Pointer
		s.generatePointerPath()
	case 5: // Steps
		s.generateStepsPath()
	}
}

func (s *SliderCtrl) Vertex() (x, y float64, cmd basics.PathCommand) {
	switch s.currentPath {
	case 2: // Text path
		return s.textRenderer.Vertex()

	case 3, 4: // Ellipse paths (pointer preview and pointer)
		cmd = s.ellipse.Vertex(&x, &y)
		if !basics.IsStop(cmd) {
			s.TransformXY(&x, &y)
		}
		return x, y, cmd

	case 5: // Steps path
		x, y, cmdUint := s.pathStorage.NextVertex()
		pathCmd := basics.PathCommand(cmdUint)
		if !basics.IsStop(pathCmd) {
			s.TransformXY(&x, &y)
		}
		return x, y, pathCmd

	default: // Paths 0 and 1 (background and triangle)
		if s.vertexIndex >= s.vertexCount {
			return 0, 0, basics.PathCmdStop
		}

		if s.vertexIndex == 0 {
			x = s.vertices[0]
			y = s.vertices[1]
			s.vertexIndex++
			s.TransformXY(&x, &y)
			return x, y, basics.PathCmdMoveTo
		}

		idx := s.vertexIndex * 2
		if idx+1 < uint(len(s.vertices)) && idx < s.vertexCount*2 {
			x = s.vertices[idx]
			y = s.vertices[idx+1]
			s.vertexIndex++
			s.TransformXY(&x, &y)

			// Close the path for filled shapes
			if s.vertexIndex >= s.vertexCount {
				return x, y, basics.PathCmdLineTo | basics.PathFlagClose
			}
			return x, y, basics.PathCmdLineTo
		}

		return 0, 0, basics.PathCmdStop
	}
}

func (s *SliderCtrl) Color(pathID uint) color.RGBA {
	if pathID < uint(len(s.colors)) {
		return s.colors[pathID]
	}
	return s.colors[0] // Default color
}

// Color customization methods

// SetBackgroundColor sets the background color (path 0)
func (s *SliderCtrl) SetBackgroundColor(c color.RGBA) {
	s.colors[0] = c
}

// SetTriangleColor sets the triangle color (path 1)
func (s *SliderCtrl) SetTriangleColor(c color.RGBA) {
	s.colors[1] = c
}

// SetTextColor sets the text color (paths 2 and 5)
func (s *SliderCtrl) SetTextColor(c color.RGBA) {
	s.colors[2] = c
	s.colors[5] = c
}

// SetPointerPreviewColor sets the pointer preview color (path 3)
func (s *SliderCtrl) SetPointerPreviewColor(c color.RGBA) {
	s.colors[3] = c
}

// SetPointerColor sets the pointer color (path 4)
func (s *SliderCtrl) SetPointerColor(c color.RGBA) {
	s.colors[4] = c
}

// Path generation methods

func (s *SliderCtrl) generateBackgroundPath() {
	// Background rectangle with border extra
	s.vertices[0] = s.X1() - s.borderExtra
	s.vertices[1] = s.Y1() - s.borderExtra
	s.vertices[2] = s.X2() + s.borderExtra
	s.vertices[3] = s.Y1() - s.borderExtra
	s.vertices[4] = s.X2() + s.borderExtra
	s.vertices[5] = s.Y2() + s.borderExtra
	s.vertices[6] = s.X1() - s.borderExtra
	s.vertices[7] = s.Y2() + s.borderExtra
	s.vertexCount = 4
}

func (s *SliderCtrl) generateTrianglePath() {
	// Triangle shape indicating slider direction
	if s.descending {
		s.vertices[0] = s.X1()
		s.vertices[1] = s.Y1()
		s.vertices[2] = s.X2()
		s.vertices[3] = s.Y1()
		s.vertices[4] = s.X1()
		s.vertices[5] = s.Y2()
		s.vertices[6] = s.X1()
		s.vertices[7] = s.Y1()
	} else {
		s.vertices[0] = s.X1()
		s.vertices[1] = s.Y1()
		s.vertices[2] = s.X2()
		s.vertices[3] = s.Y1()
		s.vertices[4] = s.X2()
		s.vertices[5] = s.Y2()
		s.vertices[6] = s.X1()
		s.vertices[7] = s.Y1()
	}
	s.vertexCount = 4
}

func (s *SliderCtrl) generateTextPath() {
	if s.label != "" {
		// Format the label with current value (matching C++ behavior)
		text := fmt.Sprintf(s.label, s.Value())
		s.textRenderer.SetText(text)
		// Position text at start of slider with proper size
		s.textRenderer.SetPosition(s.X1(), s.Y1())
		s.textRenderer.SetSize((s.Y2() - s.Y1()) * 1.2)
		s.textRenderer.Rewind(0)
	} else {
		// Clear text if no label
		s.textRenderer.SetText("")
		s.textRenderer.Rewind(0)
	}
}

func (s *SliderCtrl) generatePointerPreviewPath() {
	// Ellipse at preview position
	centerX := s.xs1 + (s.xs2-s.xs1)*s.previewValue
	centerY := (s.ys1 + s.ys2) / 2.0
	radius := s.Y2() - s.Y1()

	s.ellipse.Init(centerX, centerY, radius, radius, 32, false)
	s.ellipse.Rewind(0)
}

func (s *SliderCtrl) generatePointerPath() {
	// Ellipse at actual value position
	s.normalizeValue(false)
	centerX := s.xs1 + (s.xs2-s.xs1)*s.value
	centerY := (s.ys1 + s.ys2) / 2.0
	radius := s.Y2() - s.Y1()

	s.ellipse.Init(centerX, centerY, radius, radius, 32, false)
	s.ellipse.Rewind(0)
}

func (s *SliderCtrl) generateStepsPath() {
	s.pathStorage.RemoveAll()

	if s.numSteps > 0 {
		// Generate tick marks like in C++ AGG
		for i := uint(0); i <= s.numSteps; i++ {
			x := s.xs1 + (s.xs2-s.xs1)*float64(i)/float64(s.numSteps)

			// Calculate tick width (adaptive based on slider width)
			d := (s.xs2 - s.xs1) / float64(s.numSteps)
			if d > 0.004 {
				d = 0.004
			}

			// Create tick mark with small triangle shape
			s.pathStorage.MoveTo(x, s.Y1())
			s.pathStorage.LineTo(x-d*(s.X2()-s.X1()), s.Y1()-s.borderExtra)
			s.pathStorage.LineTo(x+d*(s.X2()-s.X1()), s.Y1()-s.borderExtra)
		}
	}
	s.pathStorage.Rewind(0)
}
