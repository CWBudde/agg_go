package path

import (
	"bytes"
	"encoding/binary"

	"agg_go/internal/basics"
)

// SerializedIntegerPathAdaptor reads integer path data from a serialized byte buffer.
// This corresponds to AGG's serialized_integer_path_adaptor<T, CoordShift> template class.
// It allows reading paths that were previously serialized with PathStorageInteger.
type SerializedIntegerPathAdaptor[T ~int16 | ~int32 | ~int64] struct {
	data       []byte
	end        int
	ptr        int
	dx         float64
	dy         float64
	scale      float64
	vertices   uint32
	coordShift int
}

// NewSerializedIntegerPathAdaptor creates a new adaptor for reading serialized integer paths.
func NewSerializedIntegerPathAdaptor[T ~int16 | ~int32 | ~int64]() *SerializedIntegerPathAdaptor[T] {
	return &SerializedIntegerPathAdaptor[T]{
		data:       nil,
		end:        0,
		ptr:        0,
		dx:         0.0,
		dy:         0.0,
		scale:      1.0,
		vertices:   0,
		coordShift: DefaultCoordShift,
	}
}

// NewSerializedIntegerPathAdaptorWithData creates a new adaptor with initial data.
func NewSerializedIntegerPathAdaptorWithData[T ~int16 | ~int32 | ~int64](
	data []byte, dx, dy float64,
) *SerializedIntegerPathAdaptor[T] {
	adaptor := NewSerializedIntegerPathAdaptor[T]()
	adaptor.Init(data, dx, dy, 1.0, DefaultCoordShift)
	return adaptor
}

// Init initializes the adaptor with serialized data and transformation parameters.
func (sipa *SerializedIntegerPathAdaptor[T]) Init(
	data []byte, dx, dy, scale float64, coordShift int,
) {
	sipa.data = data
	sipa.end = len(data)
	sipa.ptr = 0
	sipa.dx = dx
	sipa.dy = dy
	sipa.scale = scale
	sipa.vertices = 0
	if coordShift == 0 {
		sipa.coordShift = DefaultCoordShift
	} else {
		sipa.coordShift = coordShift
	}
}

// Rewind resets the adaptor to the beginning of the data.
func (sipa *SerializedIntegerPathAdaptor[T]) Rewind(pathID uint32) {
	sipa.ptr = 0
	sipa.vertices = 0
}

// Vertex reads the next vertex from the serialized data.
func (sipa *SerializedIntegerPathAdaptor[T]) Vertex() (float64, float64, basics.PathCommand) {
	if sipa.data == nil || len(sipa.data) == 0 || sipa.ptr > sipa.end {
		return 0, 0, basics.PathCmdStop
	}

	if sipa.ptr == sipa.end {
		sipa.ptr += int(binary.Size(VertexInteger[T]{}))
		return 0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
	}

	// Check if we have enough data for a complete vertex
	vertexSize := int(binary.Size(VertexInteger[T]{}))
	if sipa.ptr+vertexSize > sipa.end {
		return 0, 0, basics.PathCmdStop
	}

	// Read the vertex data
	var vertex VertexInteger[T]
	reader := bytes.NewReader(sipa.data[sipa.ptr : sipa.ptr+vertexSize])
	if err := binary.Read(reader, binary.LittleEndian, &vertex); err != nil {
		return 0, 0, basics.PathCmdStop
	}

	x, y, cmd := vertex.Vertex(sipa.dx, sipa.dy, sipa.scale, sipa.coordShift)

	// Handle polygon closing for move_to commands after we've already processed vertices
	if basics.IsMoveTo(cmd) && sipa.vertices > 2 {
		sipa.vertices = 0
		return 0, 0, basics.PathCmdEndPoly | basics.PathCommand(basics.PathFlagsClose)
	}

	sipa.vertices++
	sipa.ptr += vertexSize
	return x, y, cmd
}

// Size returns the number of vertices in the serialized data.
func (sipa *SerializedIntegerPathAdaptor[T]) Size() uint32 {
	if sipa.data == nil {
		return 0
	}
	vertexSize := int(binary.Size(VertexInteger[T]{}))
	return uint32(sipa.end / vertexSize)
}

// IsEmpty returns true if there is no data to read.
func (sipa *SerializedIntegerPathAdaptor[T]) IsEmpty() bool {
	return sipa.data == nil || sipa.end == 0
}

// SetTransform updates the transformation parameters.
func (sipa *SerializedIntegerPathAdaptor[T]) SetTransform(dx, dy, scale float64) {
	sipa.dx = dx
	sipa.dy = dy
	sipa.scale = scale
}
