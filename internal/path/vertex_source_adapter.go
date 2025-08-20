package path

import (
	"agg_go/internal/basics"
)

// VertexSourceInterface defines the interface for vertex sources
// This mirrors the interface from internal/conv but avoids circular imports
type VertexSourceInterface interface {
	Rewind(pathID uint)
	Vertex() (x, y float64, cmd basics.PathCommand)
}

// PathStorageVertexSourceAdapter adapts PathStorage to the VertexSourceInterface
// required by converters like ConvGPC. This bridges the gap between PathStorage's
// NextVertex method and the VertexSourceInterface's Vertex method.
type PathStorageVertexSourceAdapter struct {
	pathStorage *PathStorage
}

// NewPathStorageVertexSourceAdapter creates a new adapter for PathStorage
func NewPathStorageVertexSourceAdapter(pathStorage *PathStorage) *PathStorageVertexSourceAdapter {
	return &PathStorageVertexSourceAdapter{
		pathStorage: pathStorage,
	}
}

// Rewind resets the vertex iterator to the beginning
func (adapter *PathStorageVertexSourceAdapter) Rewind(pathID uint) {
	adapter.pathStorage.Rewind(pathID)
}

// Vertex returns the next vertex from the path storage
func (adapter *PathStorageVertexSourceAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmdRaw := adapter.pathStorage.NextVertex()
	return x, y, basics.PathCommand(cmdRaw)
}

// PathStorageStlVertexSourceAdapter adapts PathStorageStl to the VertexSourceInterface
type PathStorageStlVertexSourceAdapter struct {
	pathStorage *PathStorageStl
}

// NewPathStorageStlVertexSourceAdapter creates a new adapter for PathStorageStl
func NewPathStorageStlVertexSourceAdapter(pathStorage *PathStorageStl) *PathStorageStlVertexSourceAdapter {
	return &PathStorageStlVertexSourceAdapter{
		pathStorage: pathStorage,
	}
}

// Rewind resets the vertex iterator to the beginning
func (adapter *PathStorageStlVertexSourceAdapter) Rewind(pathID uint) {
	adapter.pathStorage.Rewind(pathID)
}

// Vertex returns the next vertex from the path storage
func (adapter *PathStorageStlVertexSourceAdapter) Vertex() (x, y float64, cmd basics.PathCommand) {
	x, y, cmdRaw := adapter.pathStorage.NextVertex()
	return x, y, basics.PathCommand(cmdRaw)
}
