// Package array provides vertex sequence functionality for AGG.
// This implements a port of AGG's vertex_sequence template class.
package array

import (
	"agg_go/internal/basics"
	"math"
)

// VertexSequence is a specialized vector that automatically filters vertices
// based on a validation function. This is equivalent to AGG's vertex_sequence<T, S>.
// The type T must implement a callable operator that returns true if the vertex
// should be kept or false if it should be filtered out.
type VertexSequence[T VertexFilter] struct {
	storage *PodBVector[T]
}

// VertexFilter represents a vertex type that can validate itself against another vertex.
// This corresponds to AGG's requirement that T must expose bool T::operator() (const T& val).
type VertexFilter interface {
	// Validate checks if this vertex should be kept when the given vertex is being added.
	// Returns true if the vertex meets the criteria, false if it should be filtered.
	Validate(val VertexFilter) bool
}

// NewVertexSequence creates a new vertex sequence with default block scale.
func NewVertexSequence[T VertexFilter]() *VertexSequence[T] {
	return &VertexSequence[T]{
		storage: NewPodBVectorWithScale[T](NewBlockScale(6)), // Default S=6 like AGG
	}
}

// NewVertexSequenceWithScale creates a new vertex sequence with specified block scale.
func NewVertexSequenceWithScale[T VertexFilter](scale BlockScale) *VertexSequence[T] {
	return &VertexSequence[T]{
		storage: NewPodBVectorWithScale[T](scale),
	}
}

// Size returns the number of vertices in the sequence.
func (vs *VertexSequence[T]) Size() int {
	return vs.storage.Size()
}

// Get returns the vertex at the specified index.
func (vs *VertexSequence[T]) Get(index int) T {
	return vs.storage.At(index)
}

// Add adds a vertex to the sequence after validation.
// This corresponds to AGG's vertex_sequence::add method.
func (vs *VertexSequence[T]) Add(val T) {
	// If we have more than 1 vertex, validate the previous vertex against the current one
	if vs.storage.Size() > 1 {
		prev := vs.storage.At(vs.storage.Size() - 2)
		curr := vs.storage.At(vs.storage.Size() - 1)
		if !prev.Validate(curr) {
			vs.storage.RemoveLast()
		}
	}
	vs.storage.Add(val)
}

// ModifyLast replaces the last vertex with a new one after validation.
// This corresponds to AGG's vertex_sequence::modify_last method.
func (vs *VertexSequence[T]) ModifyLast(val T) {
	vs.storage.RemoveLast()
	vs.Add(val)
}

// Close processes the sequence for closure and removes invalid vertices.
// This corresponds to AGG's vertex_sequence::close method.
func (vs *VertexSequence[T]) Close(closed bool) {
	// Remove trailing vertices that don't validate
	for vs.storage.Size() > 1 {
		prev := vs.storage.At(vs.storage.Size() - 2)
		curr := vs.storage.At(vs.storage.Size() - 1)
		if prev.Validate(curr) {
			break
		}
		last := vs.storage.At(vs.storage.Size() - 1)
		vs.storage.RemoveLast()
		vs.ModifyLast(last)
	}

	// If closed, validate the last vertex against the first
	if closed {
		for vs.storage.Size() > 1 {
			last := vs.storage.At(vs.storage.Size() - 1)
			first := vs.storage.At(0)
			if last.Validate(first) {
				break
			}
			vs.storage.RemoveLast()
		}
	}
}

// RemoveAll clears all vertices from the sequence.
func (vs *VertexSequence[T]) RemoveAll() {
	vs.storage.RemoveAll()
}

// LineAAVertex represents a vertex for anti-aliased line rendering.
// This corresponds to AGG's line_aa_vertex struct.
type LineAAVertex struct {
	X, Y int // Vertex coordinates
	Len  int // Distance to the next vertex
}

// NewLineAAVertex creates a new line AA vertex.
func NewLineAAVertex(x, y int) LineAAVertex {
	return LineAAVertex{X: x, Y: y, Len: 0}
}

// Validate implements the VertexFilter interface for LineAAVertex.
// This corresponds to AGG's line_aa_vertex::operator() method.
// It calculates the distance to the next vertex and returns true if the distance
// is greater than the minimum threshold (line_subpixel_scale + line_subpixel_scale/2).
func (v LineAAVertex) Validate(val VertexFilter) bool {
	other, ok := val.(LineAAVertex)
	if !ok {
		return false
	}

	dx := float64(other.X - v.X)
	dy := float64(other.Y - v.Y)

	// Calculate distance and store in len (modifying the vertex would require a pointer)
	// For validation, we just check the threshold
	distance := basics.URound(math.Sqrt(dx*dx + dy*dy))

	// Line subpixel scale constants (from line_aa_basics)
	const lineSubpixelScale = 256
	threshold := lineSubpixelScale + lineSubpixelScale/2

	return int(distance) > threshold
}

// CalculateDistance calculates and sets the distance to another vertex.
// This should be called to update the Len field after validation.
func (v *LineAAVertex) CalculateDistance(other LineAAVertex) {
	dx := float64(other.X - v.X)
	dy := float64(other.Y - v.Y)
	v.Len = int(basics.URound(math.Sqrt(dx*dx + dy*dy)))
}

// VertexDist represents a vertex with distance information for general use.
// This corresponds to AGG's vertex_dist struct.
type VertexDist struct {
	X, Y float64 // Vertex coordinates
	Dist float64 // Distance to the next vertex
}

// NewVertexDist creates a new vertex with distance.
func NewVertexDist(x, y float64) VertexDist {
	return VertexDist{X: x, Y: y, Dist: 0.0}
}

// Validate implements the VertexFilter interface for VertexDist.
// This corresponds to AGG's vertex_dist::operator() method.
func (v VertexDist) Validate(val VertexFilter) bool {
	other, ok := val.(VertexDist)
	if !ok {
		return false
	}

	distance := basics.CalcDistance(v.X, v.Y, other.X, other.Y)

	// Use vertex distance epsilon for comparison
	ret := distance > basics.VertexDistEpsilon
	if !ret {
		// Set a small positive distance if below epsilon
		// Note: In Go we can't modify the distance directly in validation
		// This would need to be handled by the caller
	}
	return ret
}

// CalculateDistance calculates and sets the distance to another vertex.
func (v *VertexDist) CalculateDistance(other VertexDist) {
	v.Dist = basics.CalcDistance(v.X, v.Y, other.X, other.Y)
	if v.Dist <= basics.VertexDistEpsilon {
		v.Dist = 1.0 / basics.VertexDistEpsilon
	}
}
