// Package ctrl provides the base control infrastructure for AGG interactive UI widgets.
// This is a port of AGG's agg_ctrl.h functionality.
package ctrl

import (
	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

// Ctrl defines the core control interface that all AGG controls must implement.
// This corresponds to the C++ ctrl base class from agg_ctrl.h.
type Ctrl interface {
	// Event handling methods
	InRect(x, y float64) bool
	OnMouseButtonDown(x, y float64) bool
	OnMouseButtonUp(x, y float64) bool
	OnMouseMove(x, y float64, buttonPressed bool) bool
	OnArrowKeys(left, right, down, up bool) bool

	// Transformation methods
	SetTransform(mtx *transform.TransAffine)
	ClearTransform()
	TransformXY(x, y *float64)
	InverseTransformXY(x, y *float64)
	Scale() float64

	// Bounds methods
	X1() float64
	Y1() float64
	X2() float64
	Y2() float64
	FlipY() bool

	// Vertex source interface for rendering
	NumPaths() uint
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)

	// Color management for multi-path rendering
	Color(pathID uint) interface{}
}

// BaseCtrl provides common control functionality that can be embedded by specific controls.
// This corresponds to the protected and private members of the C++ ctrl class.
type BaseCtrl struct {
	// Bounding rectangle
	x1, y1, x2, y2 float64

	// Coordinate system flags
	flipY bool

	// Transformation matrix
	mtx *transform.TransAffine
}

// NewBaseCtrl creates a new base control with the specified bounds and coordinate system.
// x1, y1: top-left corner coordinates
// x2, y2: bottom-right corner coordinates
// flipY: true if Y-axis should be flipped (common in window coordinate systems)
func NewBaseCtrl(x1, y1, x2, y2 float64, flipY bool) *BaseCtrl {
	return &BaseCtrl{
		x1:    x1,
		y1:    y1,
		x2:    x2,
		y2:    y2,
		flipY: flipY,
		mtx:   nil,
	}
}

// Bounds methods
func (bc *BaseCtrl) X1() float64 { return bc.x1 }
func (bc *BaseCtrl) Y1() float64 { return bc.y1 }
func (bc *BaseCtrl) X2() float64 { return bc.x2 }
func (bc *BaseCtrl) Y2() float64 { return bc.y2 }
func (bc *BaseCtrl) FlipY() bool { return bc.flipY }

// SetBounds updates the control's bounding rectangle.
func (bc *BaseCtrl) SetBounds(x1, y1, x2, y2 float64) {
	bc.x1, bc.y1, bc.x2, bc.y2 = x1, y1, x2, y2
}

// Transformation methods
func (bc *BaseCtrl) SetTransform(mtx *transform.TransAffine) {
	bc.mtx = mtx
}

func (bc *BaseCtrl) ClearTransform() {
	bc.mtx = nil
}

// TransformXY transforms coordinates for rendering.
// This applies Y-axis flipping first, then the transformation matrix.
func (bc *BaseCtrl) TransformXY(x, y *float64) {
	if bc.flipY {
		*y = bc.y1 + bc.y2 - *y
	}
	if bc.mtx != nil {
		bc.mtx.Transform(x, y)
	}
}

// InverseTransformXY applies inverse transformation for mouse coordinates.
// This applies inverse matrix transformation first, then Y-axis flipping.
func (bc *BaseCtrl) InverseTransformXY(x, y *float64) {
	if bc.mtx != nil {
		bc.mtx.InverseTransform(x, y)
	}
	if bc.flipY {
		*y = bc.y1 + bc.y2 - *y
	}
}

// Scale returns the current transformation scale factor.
// Returns 1.0 if no transformation is applied.
func (bc *BaseCtrl) Scale() float64 {
	if bc.mtx != nil {
		sx := bc.mtx.SX
		sy := bc.mtx.SY
		return (sx + sy) / 2.0 // Average scaling
	}
	return 1.0
}

// InRect checks if a point is within the control's bounds.
// This is a basic implementation that can be overridden by specific controls.
func (bc *BaseCtrl) InRect(x, y float64) bool {
	// Apply inverse transformation to convert screen coordinates to control coordinates
	bc.InverseTransformXY(&x, &y)

	return x >= bc.x1 && y >= bc.y1 && x <= bc.x2 && y <= bc.y2
}
