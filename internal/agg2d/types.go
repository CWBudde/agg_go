// Types for AGG2D interface.
// This file contains the type definitions that match the C++ AGG2D interface.
package agg2d

// Direction defines the direction for path operations (clockwise or counter-clockwise).
// This matches the C++ Agg2D::Direction enum.
type Direction int

const (
	// CW represents clockwise direction
	CW Direction = iota
	// CCW represents counter-clockwise direction
	CCW
)

// DrawPathFlag defines how paths should be rendered.
// This matches the C++ Agg2D::DrawPathFlag enum.
type DrawPathFlag int

const (
	// FillOnly renders only the fill of the path
	FillOnly DrawPathFlag = iota
	// StrokeOnly renders only the stroke of the path
	StrokeOnly
	// FillAndStroke renders both fill and stroke of the path
	FillAndStroke
	// FillWithLineColor renders fill using the line color
	FillWithLineColor
)
