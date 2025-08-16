// Package basics provides allocators that match AGG's allocation strategy.
// These allocators provide the same interface as AGG's pod_allocator and obj_allocator
// but use Go's built-in memory management.
package basics

// PodAllocator provides allocation for Plain Old Data arrays.
// In AGG, this is used for data that doesn't require constructor/destructor calls.
// In Go, we leverage the garbage collector and use slices for dynamic arrays.
type PodAllocator[T any] struct{}

// NewPodAllocator creates a new pod allocator for type T.
func NewPodAllocator[T any]() PodAllocator[T] {
	return PodAllocator[T]{}
}

// Allocate creates a new slice of the specified size.
// This corresponds to AGG's pod_allocator<T>::allocate(unsigned num).
func (PodAllocator[T]) Allocate(num int) []T {
	if num <= 0 {
		return nil
	}
	return make([]T, num)
}

// Deallocate is provided for API compatibility with AGG.
// In Go, memory is managed by the garbage collector, so this is a no-op.
// This corresponds to AGG's pod_allocator<T>::deallocate(T* ptr, unsigned).
func (PodAllocator[T]) Deallocate(ptr []T, num int) {
	// No-op in Go - garbage collector handles memory management
}

// ObjAllocator provides allocation for single objects.
// In AGG, this is used for objects that require constructor/destructor calls.
// In Go, we use new() for single object allocation.
type ObjAllocator[T any] struct{}

// NewObjAllocator creates a new object allocator for type T.
func NewObjAllocator[T any]() ObjAllocator[T] {
	return ObjAllocator[T]{}
}

// Allocate creates a new instance of type T.
// This corresponds to AGG's obj_allocator<T>::allocate().
func (ObjAllocator[T]) Allocate() *T {
	return new(T)
}

// Deallocate is provided for API compatibility with AGG.
// In Go, memory is managed by the garbage collector, so this is a no-op.
// This corresponds to AGG's obj_allocator<T>::deallocate(T* ptr).
func (ObjAllocator[T]) Deallocate(ptr *T) {
	// No-op in Go - garbage collector handles memory management
}

// Global allocator instances for convenience.
// These can be used when you need a default allocator instance.
var (
	DefaultPodAllocator = NewPodAllocator[byte]()
	DefaultObjAllocator = NewObjAllocator[any]()
)

// AllocatorInterface defines the interface for pod allocators.
// This allows for dependency injection and testing with custom allocators.
type AllocatorInterface[T any] interface {
	Allocate(num int) []T
	Deallocate(ptr []T, num int)
}

// ObjectAllocatorInterface defines the interface for object allocators.
type ObjectAllocatorInterface[T any] interface {
	Allocate() *T
	Deallocate(ptr *T)
}

// Compile-time interface compliance checks
var (
	_ AllocatorInterface[int]       = (*PodAllocator[int])(nil)
	_ ObjectAllocatorInterface[int] = (*ObjAllocator[int])(nil)
)
