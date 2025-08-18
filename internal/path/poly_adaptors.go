package path

import (
	"agg_go/internal/basics"
)

// VertexSource interface represents a source of path vertices.
// This corresponds to AGG's vertex_source concept.
type VertexSource interface {
	// Rewind resets the vertex source to start from the beginning of the specified path.
	Rewind(pathID uint)

	// NextVertex returns the next vertex coordinates and command.
	// When the path is finished, it returns PathCmdStop.
	NextVertex() (x, y float64, cmd uint32)
}

// PolyPlainAdaptor adapts a plain array of coordinates to the VertexSource interface.
// This is a direct port of AGG's poly_plain_adaptor template class.
type PolyPlainAdaptor[T ~int | ~int32 | ~float32 | ~float64] struct {
	data   []T
	ptr    int
	end    int
	closed bool
	stop   bool
}

// NewPolyPlainAdaptor creates a new polygon plain adaptor.
func NewPolyPlainAdaptor[T ~int | ~int32 | ~float32 | ~float64]() *PolyPlainAdaptor[T] {
	return &PolyPlainAdaptor[T]{}
}

// NewPolyPlainAdaptorWithData creates a new polygon plain adaptor with data.
func NewPolyPlainAdaptorWithData[T ~int | ~int32 | ~float32 | ~float64](data []T, numPoints uint, closed bool) *PolyPlainAdaptor[T] {
	adaptor := &PolyPlainAdaptor[T]{}
	adaptor.Init(data, numPoints, closed)
	return adaptor
}

// Init initializes the adaptor with coordinate data.
// data should contain interleaved x,y coordinates, so numPoints*2 elements total.
func (ppa *PolyPlainAdaptor[T]) Init(data []T, numPoints uint, closed bool) {
	ppa.data = data
	ppa.ptr = 0
	ppa.end = int(numPoints * 2)
	ppa.closed = closed
	ppa.stop = false
}

// Rewind implements the VertexSource interface.
func (ppa *PolyPlainAdaptor[T]) Rewind(pathID uint) {
	ppa.ptr = 0
	ppa.stop = false
}

// NextVertex implements the VertexSource interface.
func (ppa *PolyPlainAdaptor[T]) NextVertex() (x, y float64, cmd uint32) {
	if ppa.ptr < ppa.end {
		first := ppa.ptr == 0
		x = float64(ppa.data[ppa.ptr])
		y = float64(ppa.data[ppa.ptr+1])
		ppa.ptr += 2

		if first {
			return x, y, uint32(basics.PathCmdMoveTo)
		}
		return x, y, uint32(basics.PathCmdLineTo)
	}

	if ppa.closed && !ppa.stop {
		ppa.stop = true
		return 0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)
	}

	return 0, 0, uint32(basics.PathCmdStop)
}

// VertexContainer interface represents a container that can store vertices.
type VertexContainer[T any] interface {
	Size() int
	At(index int) T
}

// VertexType interface represents a vertex type with X and Y coordinates.
type VertexType interface {
	GetX() float64
	GetY() float64
}

// PolyContainerAdaptor adapts a container of vertices to the VertexSource interface.
// This is a direct port of AGG's poly_container_adaptor template class.
type PolyContainerAdaptor[Container VertexContainer[V], V VertexType] struct {
	container Container
	index     int
	closed    bool
	stop      bool
}

// NewPolyContainerAdaptor creates a new polygon container adaptor.
func NewPolyContainerAdaptor[Container VertexContainer[V], V VertexType]() *PolyContainerAdaptor[Container, V] {
	return &PolyContainerAdaptor[Container, V]{}
}

// NewPolyContainerAdaptorWithData creates a new polygon container adaptor with data.
func NewPolyContainerAdaptorWithData[Container VertexContainer[V], V VertexType](data Container, closed bool) *PolyContainerAdaptor[Container, V] {
	adaptor := &PolyContainerAdaptor[Container, V]{}
	adaptor.Init(data, closed)
	return adaptor
}

// Init initializes the adaptor with container data.
func (pca *PolyContainerAdaptor[Container, V]) Init(data Container, closed bool) {
	pca.container = data
	pca.index = 0
	pca.closed = closed
	pca.stop = false
}

// Rewind implements the VertexSource interface.
func (pca *PolyContainerAdaptor[Container, V]) Rewind(pathID uint) {
	pca.index = 0
	pca.stop = false
}

// NextVertex implements the VertexSource interface.
func (pca *PolyContainerAdaptor[Container, V]) NextVertex() (x, y float64, cmd uint32) {
	if pca.index < pca.container.Size() {
		first := pca.index == 0
		v := pca.container.At(pca.index)
		pca.index++

		x = v.GetX()
		y = v.GetY()

		if first {
			return x, y, uint32(basics.PathCmdMoveTo)
		}
		return x, y, uint32(basics.PathCmdLineTo)
	}

	if pca.closed && !pca.stop {
		pca.stop = true
		return 0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)
	}

	return 0, 0, uint32(basics.PathCmdStop)
}

// PolyContainerReverseAdaptor adapts a container of vertices to the VertexSource interface with reverse iteration.
// This is a direct port of AGG's poly_container_reverse_adaptor template class.
type PolyContainerReverseAdaptor[Container VertexContainer[V], V VertexType] struct {
	container Container
	index     int
	closed    bool
	stop      bool
}

// NewPolyContainerReverseAdaptor creates a new reverse polygon container adaptor.
func NewPolyContainerReverseAdaptor[Container VertexContainer[V], V VertexType]() *PolyContainerReverseAdaptor[Container, V] {
	return &PolyContainerReverseAdaptor[Container, V]{
		index: -1,
	}
}

// NewPolyContainerReverseAdaptorWithData creates a new reverse polygon container adaptor with data.
func NewPolyContainerReverseAdaptorWithData[Container VertexContainer[V], V VertexType](data Container, closed bool) *PolyContainerReverseAdaptor[Container, V] {
	adaptor := &PolyContainerReverseAdaptor[Container, V]{}
	adaptor.Init(data, closed)
	return adaptor
}

// Init initializes the adaptor with container data.
func (pcra *PolyContainerReverseAdaptor[Container, V]) Init(data Container, closed bool) {
	pcra.container = data
	pcra.index = data.Size() - 1
	pcra.closed = closed
	pcra.stop = false
}

// Rewind implements the VertexSource interface.
func (pcra *PolyContainerReverseAdaptor[Container, V]) Rewind(pathID uint) {
	pcra.index = pcra.container.Size() - 1
	pcra.stop = false
}

// NextVertex implements the VertexSource interface.
func (pcra *PolyContainerReverseAdaptor[Container, V]) NextVertex() (x, y float64, cmd uint32) {
	if pcra.index >= 0 {
		first := pcra.index == pcra.container.Size()-1
		v := pcra.container.At(pcra.index)
		pcra.index--

		x = v.GetX()
		y = v.GetY()

		if first {
			return x, y, uint32(basics.PathCmdMoveTo)
		}
		return x, y, uint32(basics.PathCmdLineTo)
	}

	if pcra.closed && !pcra.stop {
		pcra.stop = true
		return 0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)
	}

	return 0, 0, uint32(basics.PathCmdStop)
}

// LineAdaptor adapts a simple line segment to the VertexSource interface.
// This is a direct port of AGG's line_adaptor class.
type LineAdaptor struct {
	coord [4]float64 // x1, y1, x2, y2
	line  *PolyPlainAdaptor[float64]
}

// NewLineAdaptor creates a new line adaptor.
func NewLineAdaptor() *LineAdaptor {
	la := &LineAdaptor{
		line: NewPolyPlainAdaptor[float64](),
	}
	la.line.Init(la.coord[:], 2, false)
	return la
}

// NewLineAdaptorWithCoords creates a new line adaptor with coordinates.
func NewLineAdaptorWithCoords(x1, y1, x2, y2 float64) *LineAdaptor {
	la := NewLineAdaptor()
	la.Init(x1, y1, x2, y2)
	return la
}

// Init initializes the line adaptor with coordinates.
func (la *LineAdaptor) Init(x1, y1, x2, y2 float64) {
	la.coord[0] = x1
	la.coord[1] = y1
	la.coord[2] = x2
	la.coord[3] = y2
	la.line.Rewind(0)
}

// Rewind implements the VertexSource interface.
func (la *LineAdaptor) Rewind(pathID uint) {
	la.line.Rewind(pathID)
}

// NextVertex implements the VertexSource interface.
func (la *LineAdaptor) NextVertex() (x, y float64, cmd uint32) {
	return la.line.NextVertex()
}

// SimpleVertex is a simple vertex implementation for use with container adaptors.
type SimpleVertex struct {
	X, Y float64
}

// NewSimpleVertex creates a new simple vertex.
func NewSimpleVertex(x, y float64) SimpleVertex {
	return SimpleVertex{X: x, Y: y}
}

// GetX returns the X coordinate.
func (v SimpleVertex) GetX() float64 {
	return v.X
}

// GetY returns the Y coordinate.
func (v SimpleVertex) GetY() float64 {
	return v.Y
}

// SimpleVertexContainer is a simple container for vertices.
type SimpleVertexContainer []SimpleVertex

// Size returns the number of vertices in the container.
func (svc SimpleVertexContainer) Size() int {
	return len(svc)
}

// At returns the vertex at the specified index.
func (svc SimpleVertexContainer) At(index int) SimpleVertex {
	return svc[index]
}
