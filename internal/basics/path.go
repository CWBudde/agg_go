package basics

// Path command enumeration
type PathCommand uint32

const (
	PathCmdStop PathCommand = iota
	PathCmdMoveTo
	PathCmdLineTo
	PathCmdCurve3
	PathCmdCurve4
	PathCmdCurveN
	PathCmdCatrom
	PathCmdUbspline
	PathCmdEndPoly
	PathCmdMask = 0x0F
)

// Path flags enumeration
type PathFlag uint32

const (
	PathFlagsNone  PathFlag = 0
	PathFlagsCCW   PathFlag = 0x10
	PathFlagsCW    PathFlag = 0x20
	PathFlagsClose PathFlag = 0x40
	PathFlagsMask  PathFlag = 0xF0
)

// PathCommand constants with flags
const (
	PathFlagClose = PathCommand(PathFlagsClose)
)

// Path utility functions
func IsVertex(c PathCommand) bool {
	return c >= PathCmdMoveTo && c < PathCmdEndPoly
}

func IsDrawing(c PathCommand) bool {
	return c >= PathCmdLineTo && c < PathCmdEndPoly
}

func IsStop(c PathCommand) bool {
	return c == PathCmdStop
}

func IsMoveTo(c PathCommand) bool {
	return c == PathCmdMoveTo
}

func IsLineTo(c PathCommand) bool {
	return c == PathCmdLineTo
}

func IsCurve(c PathCommand) bool {
	return c == PathCmdCurve3 || c == PathCmdCurve4
}

func IsCurve3(c PathCommand) bool {
	return c == PathCmdCurve3
}

func IsCurve4(c PathCommand) bool {
	return c == PathCmdCurve4
}

func IsEndPoly(c PathCommand) bool {
	return (c & PathCommand(PathCmdMask)) == PathCmdEndPoly
}

func IsClose(cmd uint32) bool {
	return (cmd & uint32(PathFlagsClose)) == uint32(PathFlagsClose)
}

func IsNextPoly(cmd uint32) bool {
	return IsStop(PathCommand(cmd)) || IsMoveTo(PathCommand(cmd)) || IsEndPoly(PathCommand(cmd))
}

func IsCW(cmd uint32) bool {
	return (cmd & uint32(PathFlagsCW)) != 0
}

func IsCCW(cmd uint32) bool {
	return (cmd & uint32(PathFlagsCCW)) != 0
}

func IsOriented(cmd uint32) bool {
	return (cmd & uint32(PathFlagsCW|PathFlagsCCW)) != 0
}

func IsClosed(cmd uint32) bool {
	return (cmd & uint32(PathFlagsClose)) != 0
}

func GetCloseFlag(cmd uint32) uint32 {
	return cmd & uint32(PathFlagsClose)
}

func ClearOrientation(cmd uint32) uint32 {
	return cmd & ^uint32(PathFlagsCW|PathFlagsCCW)
}

func GetOrientation(cmd uint32) uint32 {
	return cmd & uint32(PathFlagsCW|PathFlagsCCW)
}

func SetOrientation(cmd uint32, orientation PathFlag) uint32 {
	return ClearOrientation(cmd) | uint32(orientation)
}
