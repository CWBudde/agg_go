package path

// PathStorageStl is an alternative path storage using slice-based storage.
// This is equivalent to AGG's stl_path_storage typedef.
// Use this for smaller paths or when you need simpler memory management.
type PathStorageStl = PathBase[*VertexStlStorage[float64]]

// PathStorageF32 is a path storage using 32-bit floating point coordinates.
// Use this when memory is limited and reduced precision is acceptable.
type PathStorageF32 = PathBase[*VertexBlockStorage[float32]]

// PathStorageStlF32 combines STL storage with 32-bit precision.
type PathStorageStlF32 = PathBase[*VertexStlStorage[float32]]

// NewPathStorageStl creates a new path storage using slice-based storage.
// This is simpler but may be less memory efficient for very large paths.
func NewPathStorageStl() *PathStorageStl {
	return NewPathBase(
		NewVertexStlStorage[float64](),
	)
}

// NewPathStorageStlWithCapacity creates a new STL path storage with initial capacity.
func NewPathStorageStlWithCapacity(capacity int) *PathStorageStl {
	return NewPathBase(
		NewVertexStlStorageWithCapacity[float64](capacity),
	)
}

// NewPathStorageF32 creates a new path storage using 32-bit coordinates.
// Use this when memory is limited and reduced precision is acceptable.
func NewPathStorageF32() *PathStorageF32 {
	return NewPathBase(
		NewVertexBlockStorage[float32](),
	)
}

// NewPathStorageStlF32 creates a new STL path storage with 32-bit coordinates.
func NewPathStorageStlF32() *PathStorageStlF32 {
	return NewPathBase(
		NewVertexStlStorage[float32](),
	)
}

// Storage type enumeration for documentation and benchmarking
type StorageType int

const (
	// StorageBlock uses block-based allocation, efficient for large paths
	StorageBlock StorageType = iota
	// StorageSlice uses slice-based allocation, simpler but potentially less efficient
	StorageSlice
)

// StorageInfo provides information about different storage types
type StorageInfo struct {
	Type        StorageType
	Description string
	BestUseCase string
}

// GetStorageInfo returns information about available storage types
func GetStorageInfo() []StorageInfo {
	return []StorageInfo{
		{
			Type:        StorageBlock,
			Description: "Block-based storage with efficient memory allocation",
			BestUseCase: "Large paths with many vertices, general purpose",
		},
		{
			Type:        StorageSlice,
			Description: "Slice-based storage using Go slices",
			BestUseCase: "Smaller paths, simple memory management",
		},
	}
}
