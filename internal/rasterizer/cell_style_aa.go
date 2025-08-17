package rasterizer

import "math"

// CellStyleAA represents a pixel cell with style information for compound rasterization.
// This extends the basic cell concept to include left and right style identifiers,
// enabling multi-style rendering in a single pass. There are no constructors defined
// intentionally to avoid extra overhead when allocating arrays of cells.
type CellStyleAA struct {
	X     int   // X coordinate of the cell
	Y     int   // Y coordinate of the cell
	Cover int   // Coverage value for anti-aliasing
	Area  int   // Area value for anti-aliasing
	Left  int16 // Left style identifier
	Right int16 // Right style identifier
}

// Initial initializes the cell to its default state.
// Sets coordinates to maximum values, coverage/area to zero, and styles to -1.
func (c *CellStyleAA) Initial() {
	c.X = math.MaxInt32
	c.Y = math.MaxInt32
	c.Cover = 0
	c.Area = 0
	c.Left = -1
	c.Right = -1
}

// Style sets the style information for the cell from another style cell.
// Only copies the Left and Right style identifiers.
func (c *CellStyleAA) Style(styleCell CellInterface) {
	if styleCell != nil {
		if sc, ok := styleCell.(*CellStyleAA); ok {
			c.Left = sc.Left
			c.Right = sc.Right
		}
	}
}

// NotEqual checks if the cell position or style differs from the given coordinates and style.
// Returns non-zero if position or style is different, zero if same.
// Uses unsigned arithmetic for position comparison and includes style comparison.
func (c *CellStyleAA) NotEqual(ex, ey int, styleCell CellInterface) int {
	positionDiff := int(uint32(ex)-uint32(c.X)) | int(uint32(ey)-uint32(c.Y))

	if styleCell == nil {
		return positionDiff
	}

	if sc, ok := styleCell.(*CellStyleAA); ok {
		styleDiff := int(uint32(c.Left)-uint32(sc.Left)) | int(uint32(c.Right)-uint32(sc.Right))
		return positionDiff | styleDiff
	}

	return positionDiff
}

// GetX returns the X coordinate of the cell
func (c *CellStyleAA) GetX() int {
	return c.X
}

// GetY returns the Y coordinate of the cell
func (c *CellStyleAA) GetY() int {
	return c.Y
}

// GetCover returns the coverage value of the cell
func (c *CellStyleAA) GetCover() int {
	return c.Cover
}

// GetArea returns the area value of the cell
func (c *CellStyleAA) GetArea() int {
	return c.Area
}

// SetPosition sets the X and Y coordinates of the cell
func (c *CellStyleAA) SetPosition(x, y int) {
	c.X = x
	c.Y = y
}

// SetCoverage sets the coverage and area values of the cell
func (c *CellStyleAA) SetCoverage(cover, area int) {
	c.Cover = cover
	c.Area = area
}

// SetStyles sets the left and right style identifiers
func (c *CellStyleAA) SetStyles(left, right int16) {
	c.Left = left
	c.Right = right
}

// SetX sets the X coordinate of the cell
func (c *CellStyleAA) SetX(x int) {
	c.X = x
}

// SetY sets the Y coordinate of the cell
func (c *CellStyleAA) SetY(y int) {
	c.Y = y
}

// SetCover sets the coverage value of the cell
func (c *CellStyleAA) SetCover(cover int) {
	c.Cover = cover
}

// SetArea sets the area value of the cell
func (c *CellStyleAA) SetArea(area int) {
	c.Area = area
}

// AddCover adds to the coverage value of the cell
func (c *CellStyleAA) AddCover(cover int) {
	c.Cover += cover
}

// AddArea adds to the area value of the cell
func (c *CellStyleAA) AddArea(area int) {
	c.Area += area
}

// Ensure CellStyleAA implements CellInterface
var _ CellInterface = (*CellStyleAA)(nil)
