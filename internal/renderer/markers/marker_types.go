// Package markers provides marker shape rendering functionality.
// This package implements fast marker drawing operations using integer coordinates
// and built-in shape algorithms.
package markers

// MarkerType represents the type of marker to draw
type MarkerType int

// Marker type constants matching AGG's marker_e enum
const (
	MarkerSquare MarkerType = iota
	MarkerDiamond
	MarkerCircle
	MarkerCrossedCircle
	MarkerSemiEllipseLeft
	MarkerSemiEllipseRight
	MarkerSemiEllipseUp
	MarkerSemiEllipseDown
	MarkerTriangleLeft
	MarkerTriangleRight
	MarkerTriangleUp
	MarkerTriangleDown
	MarkerFourRays
	MarkerCross
	MarkerX
	MarkerDash
	MarkerDot
	MarkerPixel

	EndOfMarkers
)

// String returns the name of the marker type
func (m MarkerType) String() string {
	switch m {
	case MarkerSquare:
		return "square"
	case MarkerDiamond:
		return "diamond"
	case MarkerCircle:
		return "circle"
	case MarkerCrossedCircle:
		return "crossed_circle"
	case MarkerSemiEllipseLeft:
		return "semiellipse_left"
	case MarkerSemiEllipseRight:
		return "semiellipse_right"
	case MarkerSemiEllipseUp:
		return "semiellipse_up"
	case MarkerSemiEllipseDown:
		return "semiellipse_down"
	case MarkerTriangleLeft:
		return "triangle_left"
	case MarkerTriangleRight:
		return "triangle_right"
	case MarkerTriangleUp:
		return "triangle_up"
	case MarkerTriangleDown:
		return "triangle_down"
	case MarkerFourRays:
		return "four_rays"
	case MarkerCross:
		return "cross"
	case MarkerX:
		return "x"
	case MarkerDash:
		return "dash"
	case MarkerDot:
		return "dot"
	case MarkerPixel:
		return "pixel"
	default:
		return "unknown"
	}
}
