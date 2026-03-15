package vcgen

import (
	"github.com/MeKo-Christian/agg_go/internal/array"
	"github.com/MeKo-Christian/agg_go/internal/basics"
)

// CoordType represents a coordinate pair for marker positions
type CoordType struct {
	X, Y float64
}

// NewCoordType creates a new coordinate pair
func NewCoordType(x, y float64) CoordType {
	return CoordType{X: x, Y: y}
}

// VCGenMarkersTerm is a terminal markers generator (arrowhead/arrowtail)
// This class generates markers at the start and end points of paths.
// It stores coordinate pairs and provides them as move_to/line_to commands.
type VCGenMarkersTerm struct {
	markers array.PodBVector[CoordType] // Storage for marker coordinates
	currID  uint                        // Current path ID for rewind
	currIdx uint                        // 0=move_to, 1=line_to, 2=stop
}

// NewVCGenMarkersTerm creates a new terminal markers generator
func NewVCGenMarkersTerm() *VCGenMarkersTerm {
	return &VCGenMarkersTerm{
		markers: *array.NewPodBVector[CoordType](), // Use default capacity
		currID:  0,
		currIdx: 0,
	}
}

// RemoveAll removes all markers (Vertex Generator Interface)
func (m *VCGenMarkersTerm) RemoveAll() {
	m.markers.RemoveAll()
}

// AddVertex adds a vertex to the marker system (Vertex Generator Interface)
// This processes path commands to determine marker positions:
// - MoveTo commands define potential marker start positions
// - Vertex commands (LineTo, etc.) define marker end positions
func (m *VCGenMarkersTerm) AddVertex(x, y float64, cmd basics.PathCommand) {
	if basics.IsMoveTo(cmd) {
		if m.markers.Size()&1 != 0 {
			// Initial state, the first coordinate was added.
			// If two or more calls of start_vertex() occur
			// we just modify the last one.
			m.markers.ModifyLast(NewCoordType(x, y))
		} else {
			m.markers.Add(NewCoordType(x, y))
		}
	} else if basics.IsVertex(cmd) {
		if m.markers.Size()&1 != 0 {
			// Initial state, the first coordinate was added.
			// Add three more points: 0,1,1,0
			m.markers.Add(NewCoordType(x, y))
			last := m.markers.At(m.markers.Size() - 1)
			m.markers.Add(last)
			third := m.markers.At(m.markers.Size() - 3)
			m.markers.Add(third)
		} else if m.markers.Size() > 0 {
			// Replace two last points: 0,1,1,0 -> 0,1,2,1
			prev := m.markers.At(m.markers.Size() - 2)
			m.markers.Set(m.markers.Size()-1, prev)
			m.markers.Set(m.markers.Size()-2, NewCoordType(x, y))
		}
	}
}

// PrepareSrc prepares the source for processing (required for some interfaces)
func (m *VCGenMarkersTerm) PrepareSrc() {
	// No preparation needed for terminal markers
}

// Rewind rewinds the marker generator to the beginning (Vertex Source Interface)
func (m *VCGenMarkersTerm) Rewind(pathID uint) {
	m.currID = pathID * 2
	m.currIdx = 0
}

// Vertex returns the next marker vertex (Vertex Source Interface)
// Returns marker coordinates as move_to/line_to commands for rendering
func (m *VCGenMarkersTerm) Vertex() (x, y float64, cmd basics.PathCommand) {
	// Each path contributes one marker pair: move_to at index pathID*2,
	// line_to at index pathID*2+1.
	if m.markers.Size() < 2 || int(m.currID+1) >= m.markers.Size() {
		return 0, 0, basics.PathCmdStop
	}

	switch m.currIdx {
	case 0:
		coord := m.markers.At(int(m.currID))
		x, y = coord.X, coord.Y
		m.currIdx = 1
		return x, y, basics.PathCmdMoveTo
	case 1:
		coord := m.markers.At(int(m.currID + 1))
		x, y = coord.X, coord.Y
		m.currIdx = 2
		return x, y, basics.PathCmdLineTo
	default:
		return 0, 0, basics.PathCmdStop
	}
}
