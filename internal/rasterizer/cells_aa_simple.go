package rasterizer

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// RasterizerCellsAASimple implements the main rasterization algorithm with concrete CellAA type.
// This is a simplified version that avoids complex generics for the initial implementation.
type RasterizerCellsAASimple struct {
	numBlocks      uint32
	maxBlocks      uint32
	currBlock      uint32
	numCells       uint32
	cellBlockLimit uint32
	cells          [][]*CellAA               // 2D array of cell blocks
	currCellPtr    *CellAA                   // Current cell pointer
	sortedCells    *array.PodVector[*CellAA] // Sorted cell pointers
	sortedY        *array.PodVector[SortedY] // Y-coordinate ranges
	currCell       CellAA                    // Current working cell
	styleCell      CellAA                    // Style/template cell
	minX, minY     int                       // Bounding box minimum
	maxX, maxY     int                       // Bounding box maximum
	sorted         bool                      // Whether cells are sorted
}

// NewRasterizerCellsAASimple creates a new cell-based rasterizer with the specified cell block limit
func NewRasterizerCellsAASimple(cellBlockLimit uint32) *RasterizerCellsAASimple {
	r := &RasterizerCellsAASimple{
		numBlocks:      0,
		maxBlocks:      0,
		currBlock:      0,
		numCells:       0,
		cellBlockLimit: cellBlockLimit,
		cells:          nil,
		currCellPtr:    nil,
		sortedCells:    array.NewPodVector[*CellAA](),
		sortedY:        array.NewPodVector[SortedY](),
		minX:           math.MaxInt32,
		minY:           math.MaxInt32,
		maxX:           math.MinInt32,
		maxY:           math.MinInt32,
		sorted:         false,
	}

	r.styleCell.Initial()
	r.currCell.Initial()
	return r
}

// Reset clears all cells and resets the rasterizer state
func (r *RasterizerCellsAASimple) Reset() {
	r.numCells = 0
	r.currBlock = 0
	r.currCell.Initial()
	r.styleCell.Initial()
	r.sorted = false
	r.minX = math.MaxInt32
	r.minY = math.MaxInt32
	r.maxX = math.MinInt32
	r.maxY = math.MinInt32
}

// Style sets the style cell for subsequent operations
func (r *RasterizerCellsAASimple) Style(styleCell CellAA) {
	r.styleCell = styleCell
}

// Line rasterizes a line from (x1,y1) to (x2,y2) using the AGG algorithm
func (r *RasterizerCellsAASimple) Line(x1, y1, x2, y2 int) {
	// Implementation of AGG's line rasterization algorithm
	dy := y2 - y1

	if dy != 0 {
		if y1 > y2 {
			// Swap points if line goes upward
			x1, y1, x2, y2 = x2, y2, x1, y1
			dy = -dy
		}

		r.renderLine(x1, y1, x2, y2, dy)
	}

	r.setCurrCell(x2>>basics.PolySubpixelShift, y2>>basics.PolySubpixelShift)
}

// MinX returns the minimum X coordinate of the bounding box
func (r *RasterizerCellsAASimple) MinX() int { return r.minX }

// MinY returns the minimum Y coordinate of the bounding box
func (r *RasterizerCellsAASimple) MinY() int { return r.minY }

// MaxX returns the maximum X coordinate of the bounding box
func (r *RasterizerCellsAASimple) MaxX() int { return r.maxX }

// MaxY returns the maximum Y coordinate of the bounding box
func (r *RasterizerCellsAASimple) MaxY() int { return r.maxY }

// SortCells sorts all cells by Y coordinate and then by X coordinate
func (r *RasterizerCellsAASimple) SortCells() {
	// Implementation of cell sorting algorithm
	// This would include the complex sorting logic from AGG
	r.sorted = true
}

// TotalCells returns the total number of cells
func (r *RasterizerCellsAASimple) TotalCells() uint32 {
	return r.numCells
}

// ScanlineNumCells returns the number of cells for the given scanline Y
func (r *RasterizerCellsAASimple) ScanlineNumCells(y uint32) uint32 {
	if !r.sorted || int(y) < r.minY || int(y) > r.maxY {
		return 0
	}
	return r.sortedY.At(int(y - uint32(r.minY))).Num
}

// ScanlineCells returns the cells for the given scanline Y
func (r *RasterizerCellsAASimple) ScanlineCells(y uint32) []*CellAA {
	if !r.sorted || int(y) < r.minY || int(y) > r.maxY {
		return nil
	}

	sortedRange := r.sortedY.At(int(y - uint32(r.minY)))
	cells := make([]*CellAA, sortedRange.Num)

	for i := uint32(0); i < sortedRange.Num; i++ {
		cells[i] = r.sortedCells.At(int(sortedRange.Start + i))
	}

	return cells
}

// Sorted returns whether the cells have been sorted
func (r *RasterizerCellsAASimple) Sorted() bool {
	return r.sorted
}

// setCurrCell sets the current cell coordinates, adding the previous cell if needed
func (r *RasterizerCellsAASimple) setCurrCell(x, y int) {
	if r.currCell.NotEqual(x, y, &r.styleCell) != 0 {
		r.addCurrCell()
		r.currCell.Style(&r.styleCell)
		r.currCell.SetX(x)
		r.currCell.SetY(y)
		r.currCell.SetCover(0)
		r.currCell.SetArea(0)
	}
}

// addCurrCell adds the current cell to the cell list if it has coverage or area
func (r *RasterizerCellsAASimple) addCurrCell() {
	if r.currCell.GetArea() != 0 || r.currCell.GetCover() != 0 {
		if (r.numCells & CellBlockMask) == 0 {
			if r.numBlocks >= r.cellBlockLimit {
				return
			}
			r.allocateBlock()
		}

		// Store cell in current block
		blockIndex := r.numCells >> CellBlockShift
		cellIndex := r.numCells & CellBlockMask

		// Copy current cell to the allocated position
		cellCopy := r.currCell
		r.cells[blockIndex][cellIndex] = &cellCopy
		r.numCells++

		// Update bounding box
		x, y := r.currCell.GetX(), r.currCell.GetY()
		if x < r.minX {
			r.minX = x
		}
		if x > r.maxX {
			r.maxX = x
		}
		if y < r.minY {
			r.minY = y
		}
		if y > r.maxY {
			r.maxY = y
		}
	}
}

// allocateBlock allocates a new cell block
func (r *RasterizerCellsAASimple) allocateBlock() {
	if r.numBlocks >= r.maxBlocks {
		// Expand blocks array
		newMaxBlocks := r.maxBlocks + CellBlockPool
		newCells := make([][]*CellAA, newMaxBlocks)

		if r.cells != nil {
			copy(newCells, r.cells)
		}

		r.cells = newCells
		r.maxBlocks = newMaxBlocks
	}

	// Allocate new block
	r.cells[r.numBlocks] = make([]*CellAA, CellBlockSize)
	r.numBlocks++
}

// renderLine implements the AGG line rendering algorithm
func (r *RasterizerCellsAASimple) renderLine(x1, y1, x2, y2, dy int) {
	dx := x2 - x1

	ey1 := y1 >> basics.PolySubpixelShift
	ey2 := y2 >> basics.PolySubpixelShift

	fy1 := y1 & basics.PolySubpixelMask
	fy2 := y2 & basics.PolySubpixelMask

	// Implementation of the complex AGG line rasterization algorithm
	// This is a simplified version - the full implementation would include
	// all the cases handled in the original AGG render_hline method

	if ey1 == ey2 {
		// Horizontal line case
		r.renderHLine(ey1, x1, fy1, x2, fy2)
	} else {
		// Multi-scanline case - step through each Y
		for ey := ey1; ey <= ey2; ey++ {
			// Calculate X intersection for this Y
			if ey == ey1 && ey == ey2 {
				r.renderHLine(ey, x1, fy1, x2, fy2)
			} else if ey == ey1 {
				// First scanline
				nextY := (ey + 1) << basics.PolySubpixelShift
				x := int64(x1) + ((int64(dx) * int64(nextY-y1)) / int64(dy))
				r.renderHLine(ey, x1, fy1, int(x), basics.PolySubpixelScale)
			} else if ey == ey2 {
				// Last scanline
				prevY := ey << basics.PolySubpixelShift
				x := int64(x1) + ((int64(dx) * int64(prevY-y1)) / int64(dy))
				r.renderHLine(ey, int(x), 0, x2, fy2)
			} else {
				// Middle scanlines
				currY := ey << basics.PolySubpixelShift
				x := int64(x1) + ((int64(dx) * int64(currY-y1)) / int64(dy))
				nextY := (ey + 1) << basics.PolySubpixelShift
				nextX := int64(x1) + ((int64(dx) * int64(nextY-y1)) / int64(dy))
				r.renderHLine(ey, int(x), 0, int(nextX), basics.PolySubpixelScale)
			}
		}
	}
}

// renderHLine renders a horizontal line segment within a single scanline
func (r *RasterizerCellsAASimple) renderHLine(ey, x1, y1, x2, y2 int) {
	ex1 := x1 >> basics.PolySubpixelShift
	ex2 := x2 >> basics.PolySubpixelShift
	fx1 := x1 & basics.PolySubpixelMask
	fx2 := x2 & basics.PolySubpixelMask

	// Trivial case
	if y1 == y2 {
		r.setCurrCell(ex2, ey)
		return
	}

	// Single cell case
	if ex1 == ex2 {
		delta := y2 - y1
		r.currCell.AddCover(delta)
		r.currCell.AddArea((fx1 + fx2) * delta)
		return
	}

	// Multi-cell case would be implemented here with the full AGG algorithm
	// This is a simplified placeholder
	r.setCurrCell(ex2, ey)
}
