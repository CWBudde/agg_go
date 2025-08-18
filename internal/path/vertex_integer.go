package path

import (
	"agg_go/internal/basics"
)

// VertexInteger represents a vertex with integer coordinates and embedded command bits.
// This corresponds to AGG's vertex_integer<T, CoordShift> template struct.
// Coordinates are stored as integers with the command bits packed into the low bits.
type VertexInteger[T ~int16 | ~int32 | ~int64] struct {
	X T // X coordinate with command bit in LSB
	Y T // Y coordinate with command bit in LSB
}

// CoordShift represents the coordinate scaling constants.
type CoordShift[CoordShiftBits int] struct{}

// Default coordinate shift constants
const (
	DefaultCoordShift = 6
	DefaultCoordScale = 64 // 1 << 6
)

// Path command enumeration for integer vertices
const (
	CmdMoveTo uint32 = 0
	CmdLineTo uint32 = 1
	CmdCurve3 uint32 = 2
	CmdCurve4 uint32 = 3
)

// NewVertexInteger creates a new integer vertex with the given coordinates and command.
// The coordinates are shifted left by 1 bit and the command is packed into the low bits.
func NewVertexInteger[T ~int16 | ~int32 | ~int64](x, y T, cmd uint32) VertexInteger[T] {
	return VertexInteger[T]{
		X: ((x << 1) & ^1) | T(cmd&1),
		Y: ((y << 1) & ^1) | T((cmd>>1)&1),
	}
}

// NewVertexIntegerFromFloat creates a new integer vertex from floating-point coordinates.
// The coordinates are scaled by coordScale before packing.
func NewVertexIntegerFromFloat[T ~int16 | ~int32 | ~int64](x, y float64, cmd uint32, coordShift int) VertexInteger[T] {
	if coordShift == 0 {
		coordShift = DefaultCoordShift
	}
	coordScale := float64(int(1) << uint(coordShift))

	scaledX := T(x * coordScale)
	scaledY := T(y * coordScale)

	return NewVertexInteger(scaledX, scaledY, cmd)
}

// Vertex extracts the coordinates and command from the integer vertex.
// Returns the floating-point coordinates with optional offset and scaling,
// and the corresponding path command.
func (v VertexInteger[T]) Vertex(dx, dy, scale float64, coordShift int) (float64, float64, basics.PathCommand) {
	if coordShift == 0 {
		coordShift = DefaultCoordShift
	}
	coordScale := float64(int(1) << uint(coordShift))

	x := dx + (float64(v.X>>1)/coordScale)*scale
	y := dy + (float64(v.Y>>1)/coordScale)*scale

	cmdBits := ((uint32(v.Y) & 1) << 1) | (uint32(v.X) & 1)

	switch cmdBits {
	case CmdMoveTo:
		return x, y, basics.PathCmdMoveTo
	case CmdLineTo:
		return x, y, basics.PathCmdLineTo
	case CmdCurve3:
		return x, y, basics.PathCmdCurve3
	case CmdCurve4:
		return x, y, basics.PathCmdCurve4
	default:
		return x, y, basics.PathCmdStop
	}
}

// VertexSimple extracts coordinates and command using default parameters.
func (v VertexInteger[T]) VertexSimple() (float64, float64, basics.PathCommand) {
	return v.Vertex(0, 0, 1.0, DefaultCoordShift)
}
