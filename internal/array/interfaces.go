// Package array provides AGG-compatible array types and algorithms.
// This package implements POD (Plain Old Data) arrays that match the functionality
// of AGG's C++ array templates while being idiomatic Go code.
package array

// Fixed-width integers only (portable layouts).
type FixedInt interface {
	~int8 | ~int16 | ~int32 | ~int64 |
		~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Floats used in geometry.
type Float interface {
	~float32 | ~float64
}

// PlainScalarAGG are POD scalars safe for raw memcpy in AGG-style code.
// (No bool, no complex, no uintptr, no arch-dependent int/uint.)
type PlainScalarAGG interface {
	FixedInt | Float
}

// ArrayInterface defines the common interface for all array types in AGG.
// This interface allows algorithms to work with different array implementations.
type ArrayInterface[T any] interface {
	// Size returns the number of elements in the array
	Size() int

	// At returns the element at the specified index (bounds-checked)
	At(i int) T

	// Set sets the element at the specified index (bounds-checked)
	Set(i int, v T)

	// ValueAt returns the element at the specified index (may panic on out-of-bounds)
	ValueAt(i int) T
}

// GenericArrayInterface defines an unconstrained array interface for algorithms.
// This allows algorithms to work with any type, including structs.
type GenericArrayInterface[T any] interface {
	// Size returns the number of elements in the array
	Size() int

	// At returns the element at the specified index (bounds-checked)
	At(i int) T

	// Set sets the element at the specified index (bounds-checked)
	Set(i int, v T)

	// ValueAt returns the element at the specified index (may panic on out-of-bounds)
	ValueAt(i int) T
}

// MutableArrayInterface extends ArrayInterface with modification operations
type MutableArrayInterface[T any] interface {
	ArrayInterface[T]

	// Clear removes all elements but keeps capacity
	Clear()

	// RemoveAll removes all elements (equivalent to Clear)
	RemoveAll()
}

// GrowableArrayInterface extends MutableArrayInterface with growth operations
type GrowableArrayInterface[T any] interface {
	MutableArrayInterface[T]

	// Add appends an element to the end
	Add(v T)

	// PushBack appends an element to the end (equivalent to Add)
	PushBack(v T)

	// Capacity returns the current capacity
	Capacity() int
}

// SliceableArrayInterface provides access to underlying slice data
type SliceableArrayInterface[T any] interface {
	ArrayInterface[T]

	// Data returns the underlying slice
	Data() []T
}

// SerializableArrayInterface provides serialization capabilities
type SerializableArrayInterface[T any] interface {
	ArrayInterface[T]

	// ByteSize returns the size in bytes needed for serialization
	ByteSize() int

	// Serialize writes the array data to the provided byte slice
	Serialize(ptr []byte)

	// Deserialize reads array data from the provided byte slice
	Deserialize(data []byte)
}

// Ensure our main types will implement these interfaces
var (
	_ ArrayInterface[int8]          = (*PodArrayAdaptor[int8])(nil)
	_ ArrayInterface[int8]          = (*PodAutoArray[int8])(nil)
	_ MutableArrayInterface[int8]   = (*PodAutoVector[int8])(nil)
	_ SliceableArrayInterface[int8] = (*PodArray[int8])(nil)
	_ GrowableArrayInterface[int8]  = (*PodVector[int8])(nil)
	_ GrowableArrayInterface[int8]  = (*PodBVector[int8])(nil)
	_ ArrayInterface[int8]          = (*RangeAdaptor[int8])(nil)

	_ ArrayInterface[int16]          = (*PodArrayAdaptor[int16])(nil)
	_ ArrayInterface[int16]          = (*PodAutoArray[int16])(nil)
	_ MutableArrayInterface[int16]   = (*PodAutoVector[int16])(nil)
	_ SliceableArrayInterface[int16] = (*PodArray[int16])(nil)
	_ GrowableArrayInterface[int16]  = (*PodVector[int16])(nil)
	_ GrowableArrayInterface[int16]  = (*PodBVector[int16])(nil)
	_ ArrayInterface[int16]          = (*RangeAdaptor[int16])(nil)

	_ ArrayInterface[int32]          = (*PodArrayAdaptor[int32])(nil)
	_ ArrayInterface[int32]          = (*PodAutoArray[int32])(nil)
	_ MutableArrayInterface[int32]   = (*PodAutoVector[int32])(nil)
	_ SliceableArrayInterface[int32] = (*PodArray[int32])(nil)
	_ GrowableArrayInterface[int32]  = (*PodVector[int32])(nil)
	_ GrowableArrayInterface[int32]  = (*PodBVector[int32])(nil)
	_ GenericArrayInterface[int32]   = (*RangeAdaptor[int32])(nil)

	// Struct array interface compliance
	_ GenericArrayInterface[any] = (*PodStructArray[any])(nil)
)
