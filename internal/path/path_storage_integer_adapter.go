package path

import (
	"agg_go/internal/basics"
)

// PathStorageIntegerAdapter adapts PathStorageInteger to implement the VertexSource interface
// This is needed because PathStorageInteger has VertexIterate() and Rewind(uint32) methods,
// but VertexSource requires Vertex() and Rewind(uint) methods.
type PathStorageIntegerAdapter[T ~int16 | ~int32 | ~int64] struct {
	storage *PathStorageInteger[T]
}

// NewPathStorageIntegerAdapter creates a new adapter for PathStorageInteger
func NewPathStorageIntegerAdapter[T ~int16 | ~int32 | ~int64](storage *PathStorageInteger[T]) *PathStorageIntegerAdapter[T] {
	return &PathStorageIntegerAdapter[T]{
		storage: storage,
	}
}

// Rewind rewinds the path storage to start iteration from the beginning
// Adapts uint to uint32 for the underlying PathStorageInteger
func (psia *PathStorageIntegerAdapter[T]) Rewind(pathID uint) {
	psia.storage.Rewind(uint32(pathID))
}

// Vertex returns the next vertex in the path iteration
// Adapts VertexIterate() to the VertexSource interface
func (psia *PathStorageIntegerAdapter[T]) Vertex() (x, y float64, cmd basics.PathCommand) {
	return psia.storage.VertexIterate()
}
