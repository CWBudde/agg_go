package path

// VertexD represents a vertex with double precision coordinates and command.
// This corresponds to AGG's vertex_d struct used in STL storage.
type VertexD struct {
	X   float64
	Y   float64
	Cmd byte
}

// NewVertexD creates a new vertex with the given coordinates and command.
func NewVertexD(x, y float64, cmd uint32) VertexD {
	return VertexD{
		X:   x,
		Y:   y,
		Cmd: byte(cmd),
	}
}

// VertexStlStorage provides slice-based vertex storage similar to std::vector.
// This is a direct port of AGG's vertex_stl_storage template class.
// It's simpler than block-based storage but may be less memory efficient for very large paths.
type VertexStlStorage[T ~float32 | ~float64] struct {
	vertices []VertexD
}

// NewVertexStlStorage creates a new STL-based vertex storage.
func NewVertexStlStorage[T ~float32 | ~float64]() *VertexStlStorage[T] {
	return &VertexStlStorage[T]{
		vertices: make([]VertexD, 0),
	}
}

// NewVertexStlStorageWithCapacity creates a new STL-based vertex storage with initial capacity.
func NewVertexStlStorageWithCapacity[T ~float32 | ~float64](capacity int) *VertexStlStorage[T] {
	return &VertexStlStorage[T]{
		vertices: make([]VertexD, 0, capacity),
	}
}

// RemoveAll removes all vertices but keeps allocated memory.
func (vss *VertexStlStorage[T]) RemoveAll() {
	vss.vertices = vss.vertices[:0] // Keep capacity, reset length
}

// FreeAll removes all vertices and deallocates memory.
func (vss *VertexStlStorage[T]) FreeAll() {
	vss.vertices = nil
}

// AddVertex adds a new vertex with command to the storage.
func (vss *VertexStlStorage[T]) AddVertex(x, y float64, cmd uint32) {
	vss.vertices = append(vss.vertices, NewVertexD(x, y, cmd))
}

// ModifyVertex modifies the coordinates of an existing vertex.
func (vss *VertexStlStorage[T]) ModifyVertex(idx uint, x, y float64) {
	if int(idx) < len(vss.vertices) {
		v := &vss.vertices[idx]
		v.X = x
		v.Y = y
	}
}

// ModifyVertexAndCommand modifies both coordinates and command of an existing vertex.
func (vss *VertexStlStorage[T]) ModifyVertexAndCommand(idx uint, x, y float64, cmd uint32) {
	if int(idx) < len(vss.vertices) {
		v := &vss.vertices[idx]
		v.X = x
		v.Y = y
		v.Cmd = byte(cmd)
	}
}

// ModifyCommand modifies the command of an existing vertex.
func (vss *VertexStlStorage[T]) ModifyCommand(idx uint, cmd uint32) {
	if int(idx) < len(vss.vertices) {
		vss.vertices[idx].Cmd = byte(cmd)
	}
}

// SwapVertices swaps two vertices in the storage.
func (vss *VertexStlStorage[T]) SwapVertices(v1, v2 uint) {
	if int(v1) < len(vss.vertices) && int(v2) < len(vss.vertices) {
		vss.vertices[v1], vss.vertices[v2] = vss.vertices[v2], vss.vertices[v1]
	}
}

// LastCommand returns the command of the last vertex, or PathCmdStop if empty.
func (vss *VertexStlStorage[T]) LastCommand() uint32 {
	if len(vss.vertices) > 0 {
		return uint32(vss.vertices[len(vss.vertices)-1].Cmd)
	}
	return uint32(0) // PathCmdStop equivalent
}

// LastVertex returns the coordinates and command of the last vertex.
func (vss *VertexStlStorage[T]) LastVertex() (x, y float64, cmd uint32) {
	if len(vss.vertices) == 0 {
		return 0.0, 0.0, uint32(0) // PathCmdStop equivalent
	}
	v := vss.vertices[len(vss.vertices)-1]
	return v.X, v.Y, uint32(v.Cmd)
}

// PrevVertex returns the coordinates and command of the second-to-last vertex.
func (vss *VertexStlStorage[T]) PrevVertex() (x, y float64, cmd uint32) {
	if len(vss.vertices) < 2 {
		return 0.0, 0.0, uint32(0) // PathCmdStop equivalent
	}
	v := vss.vertices[len(vss.vertices)-2]
	return v.X, v.Y, uint32(v.Cmd)
}

// LastX returns the X coordinate of the last vertex, or 0.0 if empty.
func (vss *VertexStlStorage[T]) LastX() float64 {
	if len(vss.vertices) > 0 {
		return vss.vertices[len(vss.vertices)-1].X
	}
	return 0.0
}

// LastY returns the Y coordinate of the last vertex, or 0.0 if empty.
func (vss *VertexStlStorage[T]) LastY() float64 {
	if len(vss.vertices) > 0 {
		return vss.vertices[len(vss.vertices)-1].Y
	}
	return 0.0
}

// TotalVertices returns the total number of vertices in the storage.
func (vss *VertexStlStorage[T]) TotalVertices() uint {
	return uint(len(vss.vertices))
}

// Vertex returns the coordinates and command of the vertex at the given index.
func (vss *VertexStlStorage[T]) Vertex(idx uint) (x, y float64, cmd uint32) {
	if int(idx) < len(vss.vertices) {
		v := vss.vertices[idx]
		return v.X, v.Y, uint32(v.Cmd)
	}
	return 0.0, 0.0, uint32(0) // PathCmdStop equivalent
}

// Command returns the command of the vertex at the given index.
func (vss *VertexStlStorage[T]) Command(idx uint) uint32 {
	if int(idx) < len(vss.vertices) {
		return uint32(vss.vertices[idx].Cmd)
	}
	return uint32(0) // PathCmdStop equivalent
}

// Capacity returns the current capacity of the underlying slice.
func (vss *VertexStlStorage[T]) Capacity() int {
	return cap(vss.vertices)
}

// Reserve ensures the storage has at least the specified capacity.
func (vss *VertexStlStorage[T]) Reserve(capacity int) {
	if cap(vss.vertices) < capacity {
		newVertices := make([]VertexD, len(vss.vertices), capacity)
		copy(newVertices, vss.vertices)
		vss.vertices = newVertices
	}
}

// Shrink reduces the capacity to match the current size, freeing unused memory.
func (vss *VertexStlStorage[T]) Shrink() {
	if cap(vss.vertices) > len(vss.vertices) {
		newVertices := make([]VertexD, len(vss.vertices))
		copy(newVertices, vss.vertices)
		vss.vertices = newVertices
	}
}
