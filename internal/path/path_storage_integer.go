package path

import (
	"bytes"
	"encoding/binary"
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// PathStorageInteger is a compact integer-based path storage container.
// This corresponds to AGG's path_storage_integer<T, CoordShift> template class.
// It provides space-efficient storage for paths using integer coordinates.
type PathStorageInteger[T ~int16 | ~int32 | ~int64] struct {
	storage    *array.PodBVector[VertexInteger[T]]
	vertexIdx  uint32
	closed     bool
	coordShift int
}

// NewPathStorageInteger creates a new integer path storage with default coordinate shift.
func NewPathStorageInteger[T ~int16 | ~int32 | ~int64]() *PathStorageInteger[T] {
	return &PathStorageInteger[T]{
		storage:    array.NewPodBVector[VertexInteger[T]](),
		vertexIdx:  0,
		closed:     true,
		coordShift: DefaultCoordShift,
	}
}

// NewPathStorageIntegerWithShift creates a new integer path storage with custom coordinate shift.
func NewPathStorageIntegerWithShift[T ~int16 | ~int32 | ~int64](shift int) *PathStorageInteger[T] {
	return &PathStorageInteger[T]{
		storage:    array.NewPodBVector[VertexInteger[T]](),
		vertexIdx:  0,
		closed:     true,
		coordShift: shift,
	}
}

// RemoveAll removes all vertices from the path.
func (psi *PathStorageInteger[T]) RemoveAll() {
	psi.storage.RemoveAll()
}

// MoveTo adds a move_to command with the given coordinates.
func (psi *PathStorageInteger[T]) MoveTo(x, y T) {
	vertex := NewVertexInteger(x, y, CmdMoveTo)
	psi.storage.Add(vertex)
}

// LineTo adds a line_to command with the given coordinates.
func (psi *PathStorageInteger[T]) LineTo(x, y T) {
	vertex := NewVertexInteger(x, y, CmdLineTo)
	psi.storage.Add(vertex)
}

// Curve3 adds a quadratic Bézier curve with control point and end point.
func (psi *PathStorageInteger[T]) Curve3(xCtrl, yCtrl, xTo, yTo T) {
	ctrl := NewVertexInteger(xCtrl, yCtrl, CmdCurve3)
	to := NewVertexInteger(xTo, yTo, CmdCurve3)
	psi.storage.Add(ctrl)
	psi.storage.Add(to)
}

// Curve4 adds a cubic Bézier curve with two control points and end point.
func (psi *PathStorageInteger[T]) Curve4(xCtrl1, yCtrl1, xCtrl2, yCtrl2, xTo, yTo T) {
	ctrl1 := NewVertexInteger(xCtrl1, yCtrl1, CmdCurve4)
	ctrl2 := NewVertexInteger(xCtrl2, yCtrl2, CmdCurve4)
	to := NewVertexInteger(xTo, yTo, CmdCurve4)
	psi.storage.Add(ctrl1)
	psi.storage.Add(ctrl2)
	psi.storage.Add(to)
}

// ClosePolygon closes the current polygon (no-op for integer storage).
func (psi *PathStorageInteger[T]) ClosePolygon() {
	// No operation needed
}

// Size returns the number of vertices in the path.
func (psi *PathStorageInteger[T]) Size() uint32 {
	return uint32(psi.storage.Size())
}

// Vertex returns the vertex at the given index.
func (psi *PathStorageInteger[T]) Vertex(idx uint32) (float64, float64, basics.PathCommand) {
	if idx >= psi.Size() {
		return 0, 0, basics.PathCmdStop
	}
	vertex := psi.storage.At(int(idx))
	return vertex.Vertex(0, 0, 1.0, psi.coordShift)
}

// ByteSize returns the number of bytes required for serialization.
func (psi *PathStorageInteger[T]) ByteSize() uint32 {
	return psi.Size() * uint32(binary.Size(VertexInteger[T]{}))
}

// Serialize writes the path data to a byte slice.
func (psi *PathStorageInteger[T]) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	for i := 0; i < psi.storage.Size(); i++ {
		vertex := psi.storage.At(i)
		if err := binary.Write(buf, binary.LittleEndian, vertex); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Rewind resets the vertex iterator to the beginning.
func (psi *PathStorageInteger[T]) Rewind(pathID uint32) {
	psi.vertexIdx = 0
	psi.closed = true
}

// Vertex iteration method for path traversal.
func (psi *PathStorageInteger[T]) VertexIterate() (float64, float64, basics.PathCommand) {
	if psi.storage.Size() < 2 || psi.vertexIdx > psi.Size() {
		return 0, 0, basics.PathCmdStop
	}

	if psi.vertexIdx == psi.Size() {
		psi.vertexIdx++
		return 0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
	}

	vertex := psi.storage.At(int(psi.vertexIdx))
	x, y, cmd := vertex.Vertex(0, 0, 1.0, psi.coordShift)

	if basics.IsMoveTo(cmd) && !psi.closed {
		psi.closed = true
		return 0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
	}

	psi.closed = false
	psi.vertexIdx++
	return x, y, cmd
}

// BoundingRect calculates the bounding rectangle of the path.
func (psi *PathStorageInteger[T]) BoundingRect() basics.Rect[float64] {
	if psi.storage.Size() == 0 {
		return basics.Rect[float64]{X1: 0, Y1: 0, X2: 0, Y2: 0}
	}

	bounds := basics.Rect[float64]{
		X1: math.Inf(1), Y1: math.Inf(1),
		X2: math.Inf(-1), Y2: math.Inf(-1),
	}

	for i := 0; i < psi.storage.Size(); i++ {
		vertex := psi.storage.At(i)
		x, y, _ := vertex.Vertex(0, 0, 1.0, psi.coordShift)

		if x < bounds.X1 {
			bounds.X1 = x
		}
		if y < bounds.Y1 {
			bounds.Y1 = y
		}
		if x > bounds.X2 {
			bounds.X2 = x
		}
		if y > bounds.Y2 {
			bounds.Y2 = y
		}
	}

	return bounds
}
