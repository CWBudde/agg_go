// Package shapes provides vector shape generators for the AGG library.
// This file implements the arrowhead vertex source for generating arrow heads and tails.
package shapes

import (
	"agg_go/internal/basics"
)

// Arrowhead generates vertices for arrowhead and arrowtail shapes.
// This is a direct port of AGG's arrowhead class from agg_arrowhead.h/cpp.
// It provides a simple way to generate polygonal arrowheads and tails
// that can be attached to line endpoints.
type Arrowhead struct {
	// Head dimensions
	headD1 float64 // Head length
	headD2 float64 // Head width at base
	headD3 float64 // Head width at tip
	headD4 float64 // Head offset

	// Tail dimensions
	tailD1 float64 // Tail length
	tailD2 float64 // Tail width at base
	tailD3 float64 // Tail width at tip
	tailD4 float64 // Tail offset

	// Flags
	headFlag bool // Whether to generate head
	tailFlag bool // Whether to generate tail

	// Vertex generation state
	coord     [16]float64           // Coordinate buffer (8 vertices * 2 coords)
	cmd       [8]basics.PathCommand // Command buffer
	currID    uint32                // Current path ID
	currCoord uint32                // Current coordinate index
}

// NewArrowhead creates a new arrowhead with default values.
func NewArrowhead() *Arrowhead {
	return &Arrowhead{
		headD1:    1.0,
		headD2:    1.0,
		headD3:    1.0,
		headD4:    0.0,
		tailD1:    1.0,
		tailD2:    1.0,
		tailD3:    1.0,
		tailD4:    0.0,
		headFlag:  false,
		tailFlag:  false,
		currID:    0,
		currCoord: 0,
	}
}

// Head sets the head dimensions and enables head generation.
// d1: head length, d2: head width at base, d3: head width at tip, d4: head offset
func (a *Arrowhead) Head(d1, d2, d3, d4 float64) {
	a.headD1 = d1
	a.headD2 = d2
	a.headD3 = d3
	a.headD4 = d4
	a.headFlag = true
}

// EnableHead enables head generation with current dimensions.
func (a *Arrowhead) EnableHead() {
	a.headFlag = true
}

// DisableHead disables head generation.
func (a *Arrowhead) DisableHead() {
	a.headFlag = false
}

// Tail sets the tail dimensions and enables tail generation.
// d1: tail length, d2: tail width at base, d3: tail width at tip, d4: tail offset
func (a *Arrowhead) Tail(d1, d2, d3, d4 float64) {
	a.tailD1 = d1
	a.tailD2 = d2
	a.tailD3 = d3
	a.tailD4 = d4
	a.tailFlag = true
}

// EnableTail enables tail generation with current dimensions.
func (a *Arrowhead) EnableTail() {
	a.tailFlag = true
}

// DisableTail disables tail generation.
func (a *Arrowhead) DisableTail() {
	a.tailFlag = false
}

// Rewind prepares the arrowhead for vertex generation.
// pathID 0 = tail, pathID 1 = head
func (a *Arrowhead) Rewind(pathID uint32) {
	a.currID = pathID
	a.currCoord = 0

	if pathID == 0 {
		// Generate tail coordinates
		if !a.tailFlag {
			a.cmd[0] = basics.PathCmdStop
			return
		}

		// Tail vertices (hexagon-like shape)
		a.coord[0] = a.tailD1 // Top right
		a.coord[1] = 0.0
		a.coord[2] = a.tailD1 - a.tailD4 // Top right inner
		a.coord[3] = a.tailD3
		a.coord[4] = -a.tailD2 - a.tailD4 // Top left inner
		a.coord[5] = a.tailD3
		a.coord[6] = -a.tailD2 // Top left
		a.coord[7] = 0.0
		a.coord[8] = -a.tailD2 - a.tailD4 // Bottom left inner
		a.coord[9] = -a.tailD3
		a.coord[10] = a.tailD1 - a.tailD4 // Bottom right inner
		a.coord[11] = -a.tailD3

		// Tail commands
		a.cmd[0] = basics.PathCmdMoveTo
		a.cmd[1] = basics.PathCmdLineTo
		a.cmd[2] = basics.PathCmdLineTo
		a.cmd[3] = basics.PathCmdLineTo
		a.cmd[4] = basics.PathCmdLineTo
		a.cmd[5] = basics.PathCmdLineTo
		a.cmd[6] = basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))
		a.cmd[7] = basics.PathCmdStop
		return
	}

	if pathID == 1 {
		// Generate head coordinates
		if !a.headFlag {
			a.cmd[0] = basics.PathCmdStop
			return
		}

		// Head vertices (triangular shape)
		a.coord[0] = -a.headD1 // Point
		a.coord[1] = 0.0
		a.coord[2] = a.headD2 + a.headD4 // Upper base
		a.coord[3] = -a.headD3
		a.coord[4] = a.headD2 // Base center
		a.coord[5] = 0.0
		a.coord[6] = a.headD2 + a.headD4 // Lower base
		a.coord[7] = a.headD3

		// Head commands
		a.cmd[0] = basics.PathCmdMoveTo
		a.cmd[1] = basics.PathCmdLineTo
		a.cmd[2] = basics.PathCmdLineTo
		a.cmd[3] = basics.PathCmdLineTo
		a.cmd[4] = basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))
		a.cmd[5] = basics.PathCmdStop
		return
	}
}

// Vertex generates the next vertex in the current path.
// Returns the path command and updates x, y with the vertex coordinates.
func (a *Arrowhead) Vertex(x, y *float64) basics.PathCommand {
	if a.currID < 2 {
		currIdx := a.currCoord * 2
		*x = a.coord[currIdx]
		*y = a.coord[currIdx+1]
		cmd := a.cmd[a.currCoord]
		a.currCoord++
		return cmd
	}
	return basics.PathCmdStop
}
