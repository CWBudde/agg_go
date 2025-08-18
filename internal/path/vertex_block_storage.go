// Package path provides path storage functionality for AGG.
// This is a port of AGG's agg_path_storage.h which implements efficient
// vertex storage using block-based memory allocation.
package path

import (
	"agg_go/internal/basics"
)

// VertexBlockStorage provides efficient storage for vertices using block-based allocation.
// This is a direct port of AGG's vertex_block_storage template class.
// T is the coordinate type (typically float64), BlockShift determines block size (2^BlockShift),
// and BlockPool determines how many blocks to allocate at once.
type VertexBlockStorage[T ~int | ~int32 | ~float32 | ~float64] struct {
	totalVertices uint
	totalBlocks   uint
	maxBlocks     uint
	coordBlocks   [][]T    // Each block contains 2*blockSize coordinates (x,y pairs)
	cmdBlocks     [][]byte // Each block contains blockSize command bytes

	// Block configuration constants
	blockShift uint
	blockSize  uint
	blockMask  uint
	blockPool  uint
}

// NewVertexBlockStorage creates a new vertex block storage with default parameters.
// Default block shift is 8 (256 vertices per block), block pool is 256 blocks.
func NewVertexBlockStorage[T ~int | ~int32 | ~float32 | ~float64]() *VertexBlockStorage[T] {
	return NewVertexBlockStorageWithParams[T](8, 256)
}

// NewVertexBlockStorageWithParams creates a new vertex block storage with custom parameters.
func NewVertexBlockStorageWithParams[T ~int | ~int32 | ~float32 | ~float64](blockShift, blockPool uint) *VertexBlockStorage[T] {
	blockSize := uint(1) << blockShift
	return &VertexBlockStorage[T]{
		blockShift: blockShift,
		blockSize:  blockSize,
		blockMask:  blockSize - 1,
		blockPool:  blockPool,
	}
}

// Copy constructor equivalent
func NewVertexBlockStorageFromCopy[T ~int | ~int32 | ~float32 | ~float64](other *VertexBlockStorage[T]) *VertexBlockStorage[T] {
	vbs := NewVertexBlockStorageWithParams[T](other.blockShift, other.blockPool)
	// Copy all vertices
	for i := uint(0); i < other.totalVertices; i++ {
		x, y, cmd := other.vertex(i)
		vbs.AddVertex(float64(x), float64(y), cmd)
	}
	return vbs
}

// RemoveAll removes all vertices but keeps allocated memory.
func (vbs *VertexBlockStorage[T]) RemoveAll() {
	vbs.totalVertices = 0
}

// FreeAll removes all vertices and deallocates memory.
func (vbs *VertexBlockStorage[T]) FreeAll() {
	vbs.coordBlocks = nil
	vbs.cmdBlocks = nil
	vbs.totalVertices = 0
	vbs.totalBlocks = 0
	vbs.maxBlocks = 0
}

// AddVertex adds a new vertex with command to the storage.
func (vbs *VertexBlockStorage[T]) AddVertex(x, y float64, cmd uint32) {
	coordPtr, cmdPtr := vbs.storagePointers()
	*cmdPtr = byte(cmd)
	coordPtr[0] = T(x)
	coordPtr[1] = T(y)
	vbs.totalVertices++
}

// ModifyVertex modifies the coordinates of an existing vertex.
func (vbs *VertexBlockStorage[T]) ModifyVertex(idx uint, x, y float64) {
	blockIdx := idx >> vbs.blockShift
	offset := (idx & vbs.blockMask) << 1
	pv := vbs.coordBlocks[blockIdx][offset:]
	pv[0] = T(x)
	pv[1] = T(y)
}

// ModifyVertexAndCommand modifies both coordinates and command of an existing vertex.
func (vbs *VertexBlockStorage[T]) ModifyVertexAndCommand(idx uint, x, y float64, cmd uint32) {
	blockIdx := idx >> vbs.blockShift
	offset := idx & vbs.blockMask
	coordOffset := offset << 1

	// Modify coordinates
	pv := vbs.coordBlocks[blockIdx][coordOffset:]
	pv[0] = T(x)
	pv[1] = T(y)

	// Modify command
	vbs.cmdBlocks[blockIdx][offset] = byte(cmd)
}

// ModifyCommand modifies the command of an existing vertex.
func (vbs *VertexBlockStorage[T]) ModifyCommand(idx uint, cmd uint32) {
	blockIdx := idx >> vbs.blockShift
	offset := idx & vbs.blockMask
	vbs.cmdBlocks[blockIdx][offset] = byte(cmd)
}

// SwapVertices swaps two vertices (both coordinates and commands).
func (vbs *VertexBlockStorage[T]) SwapVertices(v1, v2 uint) {
	b1 := v1 >> vbs.blockShift
	b2 := v2 >> vbs.blockShift
	o1 := v1 & vbs.blockMask
	o2 := v2 & vbs.blockMask

	// Swap coordinates
	pv1 := vbs.coordBlocks[b1][(o1 << 1):]
	pv2 := vbs.coordBlocks[b2][(o2 << 1):]
	pv1[0], pv2[0] = pv2[0], pv1[0]
	pv1[1], pv2[1] = pv2[1], pv1[1]

	// Swap commands
	cmd1 := vbs.cmdBlocks[b1][o1]
	vbs.cmdBlocks[b1][o1] = vbs.cmdBlocks[b2][o2]
	vbs.cmdBlocks[b2][o2] = cmd1
}

// LastCommand returns the command of the last vertex, or PathCmdStop if empty.
func (vbs *VertexBlockStorage[T]) LastCommand() uint32 {
	if vbs.totalVertices > 0 {
		return vbs.command(vbs.totalVertices - 1)
	}
	return uint32(basics.PathCmdStop)
}

// LastVertex returns the coordinates and command of the last vertex.
func (vbs *VertexBlockStorage[T]) LastVertex() (x, y float64, cmd uint32) {
	if vbs.totalVertices > 0 {
		return vbs.vertex(vbs.totalVertices - 1)
	}
	return 0, 0, uint32(basics.PathCmdStop)
}

// PrevVertex returns the coordinates and command of the second-to-last vertex.
func (vbs *VertexBlockStorage[T]) PrevVertex() (x, y float64, cmd uint32) {
	if vbs.totalVertices > 1 {
		return vbs.vertex(vbs.totalVertices - 2)
	}
	return 0, 0, uint32(basics.PathCmdStop)
}

// LastX returns the X coordinate of the last vertex.
func (vbs *VertexBlockStorage[T]) LastX() float64 {
	if vbs.totalVertices > 0 {
		idx := vbs.totalVertices - 1
		blockIdx := idx >> vbs.blockShift
		offset := (idx & vbs.blockMask) << 1
		return float64(vbs.coordBlocks[blockIdx][offset])
	}
	return 0.0
}

// LastY returns the Y coordinate of the last vertex.
func (vbs *VertexBlockStorage[T]) LastY() float64 {
	if vbs.totalVertices > 0 {
		idx := vbs.totalVertices - 1
		blockIdx := idx >> vbs.blockShift
		offset := (idx & vbs.blockMask) << 1
		return float64(vbs.coordBlocks[blockIdx][offset+1])
	}
	return 0.0
}

// TotalVertices returns the total number of vertices in storage.
func (vbs *VertexBlockStorage[T]) TotalVertices() uint {
	return vbs.totalVertices
}

// vertex returns the coordinates and command of the vertex at the given index.
func (vbs *VertexBlockStorage[T]) vertex(idx uint) (x, y float64, cmd uint32) {
	blockIdx := idx >> vbs.blockShift
	offset := (idx & vbs.blockMask) << 1
	pv := vbs.coordBlocks[blockIdx][offset:]
	x = float64(pv[0])
	y = float64(pv[1])
	cmd = uint32(vbs.cmdBlocks[blockIdx][idx&vbs.blockMask])
	return
}

// Vertex returns the coordinates and command of the vertex at the given index (public version).
func (vbs *VertexBlockStorage[T]) Vertex(idx uint) (x, y float64, cmd uint32) {
	return vbs.vertex(idx)
}

// command returns the command of the vertex at the given index.
func (vbs *VertexBlockStorage[T]) command(idx uint) uint32 {
	blockIdx := idx >> vbs.blockShift
	offset := idx & vbs.blockMask
	return uint32(vbs.cmdBlocks[blockIdx][offset])
}

// Command returns the command of the vertex at the given index (public version).
func (vbs *VertexBlockStorage[T]) Command(idx uint) uint32 {
	return vbs.command(idx)
}

// allocateBlock allocates a new block of memory for vertices.
func (vbs *VertexBlockStorage[T]) allocateBlock(nb uint) {
	if nb >= vbs.maxBlocks {
		// Expand block pointer arrays
		newMaxBlocks := vbs.maxBlocks + vbs.blockPool

		newCoordBlocks := make([][]T, newMaxBlocks)
		newCmdBlocks := make([][]byte, newMaxBlocks)

		// Copy existing pointers
		copy(newCoordBlocks, vbs.coordBlocks)
		copy(newCmdBlocks, vbs.cmdBlocks)

		vbs.coordBlocks = newCoordBlocks
		vbs.cmdBlocks = newCmdBlocks
		vbs.maxBlocks = newMaxBlocks
	}

	// Allocate coordinate block (2 coordinates per vertex)
	vbs.coordBlocks[nb] = make([]T, vbs.blockSize*2)

	// Allocate command block (1 command per vertex)
	vbs.cmdBlocks[nb] = make([]byte, vbs.blockSize)

	vbs.totalBlocks++
}

// storagePointers returns pointers to the coordinate and command storage for the next vertex.
func (vbs *VertexBlockStorage[T]) storagePointers() (coordPtr []T, cmdPtr *byte) {
	nb := vbs.totalVertices >> vbs.blockShift
	if nb >= vbs.totalBlocks {
		vbs.allocateBlock(nb)
	}

	offset := vbs.totalVertices & vbs.blockMask
	coordOffset := offset << 1
	coordPtr = vbs.coordBlocks[nb][coordOffset:]
	cmdPtr = &vbs.cmdBlocks[nb][offset]
	return
}
