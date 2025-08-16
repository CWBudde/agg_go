// Package rasterizer provides rasterization functionality for anti-aliased rendering.
// This includes cell-based rasterization and scanline generation algorithms.
package rasterizer

import "math"

// CellAA represents a pixel cell for anti-aliased rasterization.
// There are no constructors defined intentionally to avoid extra overhead
// when allocating arrays of cells, following the original AGG design.
type CellAA struct {
	X     int // X coordinate of the cell
	Y     int // Y coordinate of the cell
	Cover int // Coverage value for anti-aliasing
	Area  int // Area value for anti-aliasing
}

// Initial initializes the cell to its default state.
// Sets coordinates to maximum values and coverage/area to zero.
func (c *CellAA) Initial() {
	c.X = math.MaxInt32
	c.Y = math.MaxInt32
	c.Cover = 0
	c.Area = 0
}

// Style sets the style for the cell (no-op in basic implementation).
// This method exists for compatibility with the template interface.
func (c *CellAA) Style(styleCell CellInterface) {
	// No operation in basic cell implementation
}

// NotEqual checks if the cell position differs from the given coordinates.
// Returns non-zero if the position is different, zero if same.
// Uses unsigned arithmetic as in the original AGG implementation.
func (c *CellAA) NotEqual(ex, ey int, styleCell CellInterface) int {
	return int(uint32(ex)-uint32(c.X)) | int(uint32(ey)-uint32(c.Y))
}
