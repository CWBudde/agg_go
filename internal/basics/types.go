// Package basics provides core types and constants for the AGG rendering library.
// This package contains the fundamental building blocks used throughout the library.
package basics

// Basic integer types following AGG's naming convention
type (
	Int8   = int8
	Int8u  = uint8
	Int16  = int16
	Int16u = uint16
	Int32  = int32
	Int32u = uint32
	Int64  = int64
	Int64u = uint64
)

// CoverType represents coverage values for anti-aliasing
type CoverType = Int8u

// CoordType represents a coordinate type that can be either integer or floating-point.
// This corresponds to AGG's coord_type template parameter concept.
type CoordType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Basic geometric types
type Point[T CoordType] struct {
	X, Y T
}

type Rect[T CoordType] struct {
	X1, Y1, X2, Y2 T
}

// Normalize ensures that X1 <= X2 and Y1 <= Y2 by swapping values if needed
func (r *Rect[T]) Normalize() {
	if r.X1 > r.X2 {
		r.X1, r.X2 = r.X2, r.X1
	}
	if r.Y1 > r.Y2 {
		r.Y1, r.Y2 = r.Y2, r.Y1
	}
}

// Clip clips this rectangle against another rectangle, returning true if they intersect
func (r *Rect[T]) Clip(clipBox Rect[T]) bool {
	if r.X2 > clipBox.X1 && r.Y2 > clipBox.Y1 && r.X1 < clipBox.X2 && r.Y1 < clipBox.Y2 {
		if r.X1 < clipBox.X1 {
			r.X1 = clipBox.X1
		}
		if r.Y1 < clipBox.Y1 {
			r.Y1 = clipBox.Y1
		}
		if r.X2 > clipBox.X2 {
			r.X2 = clipBox.X2
		}
		if r.Y2 > clipBox.Y2 {
			r.Y2 = clipBox.Y2
		}
		return true
	}
	return false
}

type Vertex[T CoordType] struct {
	X, Y T
	Cmd  uint32
}

// RowInfo represents row information for rendering buffers
type RowInfo[T any] struct {
	X1, X2 int
	Ptr    []T
}

// NewRowInfo creates a new RowInfo with the specified parameters
func NewRowInfo[T any](x1, x2 int, ptr []T) RowInfo[T] {
	return RowInfo[T]{X1: x1, X2: x2, Ptr: ptr}
}

// ConstRowInfo represents read-only row information
type ConstRowInfo[T any] struct {
	X1, X2 int
	Ptr    []T
}

// NewConstRowInfo creates a new ConstRowInfo with the specified parameters
func NewConstRowInfo[T any](x1, x2 int, ptr []T) ConstRowInfo[T] {
	return ConstRowInfo[T]{X1: x1, X2: x2, Ptr: ptr}
}

// Utility functions for basic types
func IMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func IMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func UMin[T ~uint | ~uint32 | ~uint8 | ~uint16](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func UMax[T ~uint | ~uint32 | ~uint8 | ~uint16](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Rectangle utility functions
func IntersectRectangles[T CoordType](r1, r2 Rect[T]) (Rect[T], bool) {
	result := Rect[T]{
		X1: max(r1.X1, r2.X1),
		Y1: max(r1.Y1, r2.Y1),
		X2: min(r1.X2, r2.X2),
		Y2: min(r1.Y2, r2.Y2),
	}

	if result.X1 < result.X2 && result.Y1 < result.Y2 {
		return result, true
	}
	return result, false
}

func UniteRectangles[T CoordType](r1, r2 Rect[T]) Rect[T] {
	return Rect[T]{
		X1: min(r1.X1, r2.X1),
		Y1: min(r1.Y1, r2.Y1),
		X2: max(r1.X2, r2.X2),
		Y2: max(r1.Y2, r2.Y2),
	}
}

// IsEqualEps compares two floating point values with epsilon tolerance
func IsEqualEps(v1, v2, epsilon float64) bool {
	if v1 < v2 {
		return (v2 - v1) <= epsilon
	}
	return (v1 - v2) <= epsilon
}

// Commonly used type aliases matching AGG's conventions

// Point type aliases for common numeric types
type (
	PointI = Point[int]     // Integer point
	PointF = Point[float32] // Float32 point
	PointD = Point[float64] // Float64/double point
)

// Rect type aliases for common numeric types
type (
	RectI = Rect[int]     // Integer rectangle
	RectF = Rect[float32] // Float32 rectangle
	RectD = Rect[float64] // Float64/double rectangle
)

// Vertex type aliases for common numeric types
type (
	VertexI = Vertex[int]     // Integer vertex
	VertexF = Vertex[float32] // Float32 vertex
	VertexD = Vertex[float64] // Float64/double vertex
)
