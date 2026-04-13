package agg

import "github.com/cwbudde/agg_go/internal/basics"

// PathCommand is the public AGG path verb type.
type PathCommand = basics.PathCommand

const (
	PathCmdStop    PathCommand = basics.PathCmdStop
	PathCmdMoveTo  PathCommand = basics.PathCmdMoveTo
	PathCmdLineTo  PathCommand = basics.PathCmdLineTo
	PathCmdCurve3  PathCommand = basics.PathCmdCurve3
	PathCmdCurve4  PathCommand = basics.PathCmdCurve4
	PathCmdEndPoly PathCommand = basics.PathCmdEndPoly
)

// IsPathVertex reports whether cmd carries vertex coordinates.
func IsPathVertex(cmd PathCommand) bool {
	return basics.IsVertex(cmd)
}

// IsPathCurve3 reports whether cmd is a quadratic Bezier vertex.
func IsPathCurve3(cmd PathCommand) bool {
	return basics.IsCurve3(cmd)
}

// IsPathCurve4 reports whether cmd is a cubic Bezier vertex.
func IsPathCurve4(cmd PathCommand) bool {
	return basics.IsCurve4(cmd)
}

// IsPathEndPoly reports whether cmd terminates the current contour.
func IsPathEndPoly(cmd PathCommand) bool {
	return basics.IsEndPoly(cmd)
}

// IsPathClose reports whether cmd closes the current contour.
func IsPathClose(cmd PathCommand) bool {
	return basics.IsClose(uint32(cmd))
}
