// Package array provides vertex sequence functionality for AGG.
// This implements a port of AGG's vertex_sequence template class.
package array

import (
	"math"

	"agg_go/internal/basics"
)

// Use VertexFilter from basics package
type VertexFilter = basics.VertexFilter

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

// VertexDistCmd represents a vertex with distance information and command.
// This combines vertex coordinate, distance, and path command.
type VertexDistCmd struct {
	X, Y float64            // Vertex coordinates
	Dist float64            // Distance to the next vertex
	Cmd  basics.PathCommand // Path command for this vertex
}

// NewVertexDistCmd creates a new vertex with distance and command.
func NewVertexDistCmd(x, y, dist float64, cmd basics.PathCommand) VertexDistCmd {
	return VertexDistCmd{X: x, Y: y, Dist: dist, Cmd: cmd}
}

// Validate implements the VertexFilter interface for VertexDistCmd.
func (v VertexDistCmd) Validate(val VertexFilter) bool {
	other, ok := val.(VertexDistCmd)
	if !ok {
		return false
	}

	distance := basics.CalcDistance(v.X, v.Y, other.X, other.Y)
	return distance > basics.VertexDistEpsilon
}

// CalculateDistance calculates and sets the distance to another vertex.
func (v *VertexDistCmd) CalculateDistance(other VertexDistCmd) {
	v.Dist = basics.CalcDistance(v.X, v.Y, other.X, other.Y)
	if v.Dist <= basics.VertexDistEpsilon {
		v.Dist = 1.0 / basics.VertexDistEpsilon
	}
}

// VertexCmdSequence is a concrete vertex sequence for VertexDistCmd.
type VertexCmdSequence struct {
	storage *PodBVector[VertexDistCmd]
}

// NewVertexCmdSequence creates a new vertex command sequence
func NewVertexCmdSequence() *VertexCmdSequence {
	return &VertexCmdSequence{
		storage: NewPodBVectorWithScale[VertexDistCmd](NewBlockScale(6)),
	}
}

// Size returns the number of vertices in the sequence.
func (vs *VertexCmdSequence) Size() int {
	return vs.storage.Size()
}

// Get returns the vertex at the specified index.
func (vs *VertexCmdSequence) Get(index int) VertexDistCmd {
	return vs.storage.At(index)
}

// Add adds a vertex to the sequence after validation.
func (vs *VertexCmdSequence) Add(val VertexDistCmd) {
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
func (vs *VertexCmdSequence) ModifyLast(val VertexDistCmd) {
	vs.storage.RemoveLast()
	vs.Add(val)
}

// Close processes the sequence for closure and removes invalid vertices.
func (vs *VertexCmdSequence) Close(closed bool) {
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
func (vs *VertexCmdSequence) RemoveAll() {
	vs.storage.RemoveAll()
}

// At returns the vertex at the specified index
func (vs *VertexCmdSequence) At(index int) VertexDistCmd {
	return vs.storage.At(index)
}

// Set modifies the vertex at the specified index.
func (vs *VertexCmdSequence) Set(index int, val VertexDistCmd) {
	vs.storage.Set(index, val)
}

// RemoveLast removes the last vertex from the sequence.
func (vs *VertexCmdSequence) RemoveLast() {
	vs.storage.RemoveLast()
}

// ModifyAt modifies the vertex at the specified index
func (vs *VertexCmdSequence) ModifyAt(index int, val VertexDistCmd) {
	if index < vs.storage.Size() && index >= 0 {
		vs.storage.Set(index, val)
	}
}

// RemoveAt removes the vertex at the specified index
// This shifts elements to fill the gap.
func (vs *VertexCmdSequence) RemoveAt(index int) {
	if index < 0 || index >= vs.storage.Size() {
		return
	}

	// Shift elements down
	for i := index; i < vs.storage.Size()-1; i++ {
		vs.storage.Set(i, vs.storage.At(i+1))
	}
	vs.storage.RemoveLast()
}

// VertexDistSequence is a concrete vertex sequence for VertexDist.
// This eliminates the need for any() casts in CalculateDistances.
type VertexDistSequence struct {
	storage *PodBVector[VertexDist]
}

// NewVertexDistSequence creates a new vertex distance sequence with default block scale.
func NewVertexDistSequence() *VertexDistSequence {
	return &VertexDistSequence{
		storage: NewPodBVectorWithScale[VertexDist](NewBlockScale(6)),
	}
}

// NewVertexDistSequenceWithScale creates a new vertex distance sequence with specified block scale.
func NewVertexDistSequenceWithScale(scale BlockScale) *VertexDistSequence {
	return &VertexDistSequence{
		storage: NewPodBVectorWithScale[VertexDist](scale),
	}
}

// Size returns the number of vertices in the sequence.
func (vs *VertexDistSequence) Size() int {
	return vs.storage.Size()
}

// Get returns the vertex at the specified index.
func (vs *VertexDistSequence) Get(index int) VertexDist {
	return vs.storage.At(index)
}

// Add adds a vertex to the sequence after validation.
func (vs *VertexDistSequence) Add(val VertexDist) {
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
func (vs *VertexDistSequence) ModifyLast(val VertexDist) {
	vs.storage.RemoveLast()
	vs.Add(val)
}

// Close processes the sequence for closure and removes invalid vertices.
func (vs *VertexDistSequence) Close(closed bool) {
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

	vs.CalculateDistances()
}

// CalculateDistances calculates and stores distances between consecutive vertices.
// No any() casts needed - we know we're working with VertexDist.
func (vs *VertexDistSequence) CalculateDistances() {
	for i := 0; i < vs.storage.Size()-1; i++ {
		curr := vs.storage.At(i)
		next := vs.storage.At(i + 1)
		curr.CalculateDistance(next)
		vs.storage.Set(i, curr)
	}
}

// RemoveAll clears all vertices from the sequence.
func (vs *VertexDistSequence) RemoveAll() {
	vs.storage.RemoveAll()
}

// At returns the vertex at the specified index.
func (vs *VertexDistSequence) At(index int) VertexDist {
	return vs.storage.At(index)
}

// Set modifies the vertex at the specified index.
func (vs *VertexDistSequence) Set(index int, val VertexDist) {
	vs.storage.Set(index, val)
}

// RemoveLast removes the last vertex from the sequence.
func (vs *VertexDistSequence) RemoveLast() {
	vs.storage.RemoveLast()
}

// LineAAVertexSequence is a concrete vertex sequence for LineAAVertex.
// This eliminates the need for any() casts in CalculateDistances.
type LineAAVertexSequence struct {
	storage *PodBVector[LineAAVertex]
}

// NewLineAAVertexSequence creates a new line AA vertex sequence with default block scale.
func NewLineAAVertexSequence() *LineAAVertexSequence {
	return &LineAAVertexSequence{
		storage: NewPodBVectorWithScale[LineAAVertex](NewBlockScale(6)),
	}
}

// NewLineAAVertexSequenceWithScale creates a new line AA vertex sequence with specified block scale.
func NewLineAAVertexSequenceWithScale(scale BlockScale) *LineAAVertexSequence {
	return &LineAAVertexSequence{
		storage: NewPodBVectorWithScale[LineAAVertex](scale),
	}
}

// Size returns the number of vertices in the sequence.
func (vs *LineAAVertexSequence) Size() int {
	return vs.storage.Size()
}

// Get returns the vertex at the specified index.
func (vs *LineAAVertexSequence) Get(index int) LineAAVertex {
	return vs.storage.At(index)
}

// Add adds a vertex to the sequence after validation.
func (vs *LineAAVertexSequence) Add(val LineAAVertex) {
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
func (vs *LineAAVertexSequence) ModifyLast(val LineAAVertex) {
	vs.storage.RemoveLast()
	vs.Add(val)
}

// Close processes the sequence for closure and removes invalid vertices.
func (vs *LineAAVertexSequence) Close(closed bool) {
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

	vs.CalculateDistances()
}

// CalculateDistances calculates and stores distances between consecutive vertices.
// No any() casts needed - we know we're working with LineAAVertex.
func (vs *LineAAVertexSequence) CalculateDistances() {
	for i := 0; i < vs.storage.Size()-1; i++ {
		curr := vs.storage.At(i)
		next := vs.storage.At(i + 1)
		curr.CalculateDistance(next)
		vs.storage.Set(i, curr)
	}
}

// RemoveAll clears all vertices from the sequence.
func (vs *LineAAVertexSequence) RemoveAll() {
	vs.storage.RemoveAll()
}

// At returns the vertex at the specified index.
func (vs *LineAAVertexSequence) At(index int) LineAAVertex {
	return vs.storage.At(index)
}

// Set modifies the vertex at the specified index.
func (vs *LineAAVertexSequence) Set(index int, val LineAAVertex) {
	vs.storage.Set(index, val)
}

// RemoveLast removes the last vertex from the sequence.
func (vs *LineAAVertexSequence) RemoveLast() {
	vs.storage.RemoveLast()
}

// ShortenPath shortens a vertex sequence from the end by the specified distance.
// This is a port of AGG's shorten_path template function.
func ShortenPath(vs *VertexDistSequence, s float64, closed bool) {
	if s > 0.0 && vs.Size() > 1 {
		var d float64
		n := vs.Size() - 2

		// Remove vertices from the end while their distance is less than s
		// Note: In C++, the loop condition is "while(n)" which stops when n becomes 0
		// This means we don't process the vertex at index 0 (the first vertex)
		for n > 0 {
			d = vs.Get(n).Dist
			if d > s {
				break
			}
			vs.RemoveLast()
			s -= d
			n--
		}

		if vs.Size() < 2 {
			vs.RemoveAll()
		} else {
			// Adjust the last vertex position
			n = vs.Size() - 1
			prev := vs.Get(n - 1)
			last := vs.Get(n)

			d = (prev.Dist - s) / prev.Dist
			x := prev.X + (last.X-prev.X)*d
			y := prev.Y + (last.Y-prev.Y)*d

			// Create new vertex with adjusted position
			newLast := VertexDist{X: x, Y: y, Dist: 0.0}

			// Replace the last vertex
			vs.ModifyLast(newLast)

			// Validate the vertex - if it fails validation, remove it
			if !prev.Validate(newLast) {
				vs.RemoveLast()
			}

			vs.Close(closed)
		}
	}
}
