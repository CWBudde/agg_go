package ctrl

import "agg_go/internal/basics"

// Common constants for control implementations

// DefaultTextThickness is the default thickness for text rendering in controls.
const DefaultTextThickness = 1.5

// DefaultBorderWidth is the default border width for controls.
const DefaultBorderWidth = 1.0

// DefaultExtraBorder is the default extra border space for controls.
const DefaultExtraBorder = 0.0

// DefaultTextSize is the default text size for control labels.
const DefaultTextSize = 10.0

// Control state constants
const (
	StateNormal = iota
	StateHover
	StatePressed
	StateFocused
	StateDisabled
)

// MouseButton represents different mouse buttons
type MouseButton int

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

// KeyState represents keyboard key states
type KeyState struct {
	Left  bool
	Right bool
	Down  bool
	Up    bool
}

// ControlEvent represents a control event with context
type ControlEvent struct {
	X, Y   float64
	Button MouseButton
	Keys   KeyState
}

// PathInfo represents information about a rendering path
type PathInfo[C any] struct {
	ID    uint
	Color C
}

// VertexIterator provides iteration over control vertices
type VertexIterator[C any] struct {
	pathID      uint
	vertexIndex uint
	ctrl        Ctrl[C]
}

// NewVertexIterator creates a new vertex iterator for a control
func NewVertexIterator[C any](ctrl Ctrl[C], pathID uint) *VertexIterator[C] {
	iter := &VertexIterator[C]{
		pathID:      pathID,
		vertexIndex: 0,
		ctrl:        ctrl,
	}
	ctrl.Rewind(pathID)
	return iter
}

// Next returns the next vertex in the path
func (vi *VertexIterator[C]) Next() (x, y float64, cmd basics.PathCommand, done bool) {
	x, y, cmd = vi.ctrl.Vertex()
	done = cmd == basics.PathCmdStop
	if !done {
		vi.vertexIndex++
	}
	return
}

// Reset rewinds the iterator to the beginning
func (vi *VertexIterator[C]) Reset() {
	vi.vertexIndex = 0
	vi.ctrl.Rewind(vi.pathID)
}

// PathID returns the current path ID
func (vi *VertexIterator[C]) PathID() uint {
	return vi.pathID
}

// VertexIndex returns the current vertex index
func (vi *VertexIterator[C]) VertexIndex() uint {
	return vi.vertexIndex
}
