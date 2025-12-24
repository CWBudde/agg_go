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

	// Update bounding box with start point (matches C++ AGG behavior)
	ex1 := x1 >> basics.PolySubpixelShift
	ey1 := y1 >> basics.PolySubpixelShift
	ex2 := x2 >> basics.PolySubpixelShift
	ey2 := y2 >> basics.PolySubpixelShift

	if ex1 < r.minX {
		r.minX = ex1
	}
	if ex1 > r.maxX {
		r.maxX = ex1
	}
	if ey1 < r.minY {
		r.minY = ey1
	}
	if ey1 > r.maxY {
		r.maxY = ey1
	}
	if ex2 < r.minX {
		r.minX = ex2
	}
	if ex2 > r.maxX {
		r.maxX = ex2
	}
	if ey2 < r.minY {
		r.minY = ey2
	}
	if ey2 > r.maxY {
		r.maxY = ey2
	}

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
	if r.sorted {
		return
	}

	// Flush any pending current cell before sorting
	r.addCurrCell()

	// Empty/degenerate?
	if r.numCells == 0 || r.minY > r.maxY {
		r.sortedY.Clear()
		r.sortedCells.Clear()
		r.sorted = true
		return
	}

	// 1) Count cells per Y
	h := r.maxY - r.minY + 1
	counts := make([]uint32, h)

	for b := uint32(0); b < r.numBlocks; b++ {
		block := r.cells[b]
		// compute number of valid entries in this block
		limit := CellBlockSize
		// if last (possibly partial) block:
		lastBlock := (r.numCells - 1) >> CellBlockShift
		if b == lastBlock {
			// entries in last block = (numCells - fullBlocks*BlockSize)
			used := int(r.numCells - (lastBlock << CellBlockShift))
			if used < limit {
				limit = used
			}
		}
		for i := 0; i < limit; i++ {
			c := block[i]
			if c == nil {
				continue
			}
			y := c.GetY()
			if y < r.minY || y > r.maxY {
				continue
			}
			counts[y-r.minY]++
		}
	}

	// 2) Build ranges (start, num) into sortedY
	r.sortedY.Resize(h)
	total := uint32(0)
	for i := 0; i < h; i++ {
		sy := r.sortedY.At(i) // copy
		sy.Start = total
		sy.Num = counts[i]
		r.sortedY.Set(i, sy) // write back
		total += counts[i]
	}

	// 3) Allocate sortedCells and fill per-Y runs
	r.sortedCells.Resize(int(total))

	// per-Y write cursors (offset inside each run)
	write := make([]uint32, h)

	for b := uint32(0); b < r.numBlocks; b++ {
		block := r.cells[b]
		limit := CellBlockSize
		lastBlock := (r.numCells - 1) >> CellBlockShift
		if b == lastBlock {
			used := int(r.numCells - (lastBlock << CellBlockShift))
			if used < limit {
				limit = used
			}
		}
		for i := 0; i < limit; i++ {
			c := block[i]
			if c == nil {
				continue
			}
			y := c.GetY()
			if y < r.minY || y > r.maxY {
				continue
			}
			yi := y - r.minY
			sy := r.sortedY.At(yi)
			off := write[yi]
			r.sortedCells.Set(int(sy.Start+off), c)
			write[yi] = off + 1
		}
	}

	// 4) Sort cells by X within each Y-run and consolidate identical X cells
	r.sortCellsByXAndConsolidate()

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
			if ey == ey1 {
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

// sortCellsByXAndConsolidate sorts cells by X coordinate within each Y-run and consolidates
// cells with identical coordinates by summing their area and cover values
func (r *RasterizerCellsAASimple) sortCellsByXAndConsolidate() {
	h := r.maxY - r.minY + 1

	for i := 0; i < h; i++ {
		sy := r.sortedY.At(i)
		if sy.Num == 0 {
			continue
		}

		// Get slice of cell pointers for this Y-scanline
		start := int(sy.Start)
		length := int(sy.Num)

		if length <= 1 {
			continue // Single cell or empty - no need to sort
		}

		// Extract cells into a slice for sorting
		cells := make([]*CellAA, length)
		for j := 0; j < length; j++ {
			cells[j] = r.sortedCells.At(start + j)
		}

		// Sort by X coordinate using Go's standard library
		r.quickSortCellsByX(cells)

		// Consolidate cells with identical X coordinates
		consolidated := r.consolidateCells(cells)

		// Update the counts and write back consolidated cells
		newSy := sy
		newSy.Num = uint32(len(consolidated))
		r.sortedY.Set(i, newSy)

		// Write consolidated cells back to sortedCells
		for j, cell := range consolidated {
			r.sortedCells.Set(start+j, cell)
		}

		// If we have fewer cells after consolidation, we need to compact
		// For simplicity, we'll leave gaps for now as this is implementation detail
	}
}

// quickSortCellsByX implements a quicksort algorithm for cell pointers by X coordinate
// This mirrors the C++ qsort_cells implementation
func (r *RasterizerCellsAASimple) quickSortCellsByX(cells []*CellAA) {
	const qsortThreshold = 9

	if len(cells) <= 1 {
		return
	}

	// For small arrays, use insertion sort
	if len(cells) <= qsortThreshold {
		r.insertionSortCellsByX(cells)
		return
	}

	// Quicksort partition
	pivot := r.partitionCells(cells)

	// Recursively sort partitions
	r.quickSortCellsByX(cells[:pivot])
	r.quickSortCellsByX(cells[pivot+1:])
}

// insertionSortCellsByX performs insertion sort on cells by X coordinate
func (r *RasterizerCellsAASimple) insertionSortCellsByX(cells []*CellAA) {
	for i := 1; i < len(cells); i++ {
		key := cells[i]
		keyX := key.GetX()
		j := i - 1

		for j >= 0 && cells[j].GetX() > keyX {
			cells[j+1] = cells[j]
			j--
		}
		cells[j+1] = key
	}
}

// partitionCells partitions the cell array for quicksort
func (r *RasterizerCellsAASimple) partitionCells(cells []*CellAA) int {
	// Use last element as pivot
	pivotIdx := len(cells) - 1
	pivot := cells[pivotIdx]
	pivotX := pivot.GetX()

	i := -1
	for j := 0; j < pivotIdx; j++ {
		if cells[j].GetX() <= pivotX {
			i++
			cells[i], cells[j] = cells[j], cells[i]
		}
	}

	cells[i+1], cells[pivotIdx] = cells[pivotIdx], cells[i+1]
	return i + 1
}

// consolidateCells consolidates cells with identical X coordinates by summing their area and cover
func (r *RasterizerCellsAASimple) consolidateCells(cells []*CellAA) []*CellAA {
	if len(cells) == 0 {
		return cells
	}

	consolidated := make([]*CellAA, 0, len(cells))

	i := 0
	for i < len(cells) {
		currentCell := *cells[i] // Make a copy
		currentX := currentCell.GetX()

		// Sum all cells with the same X coordinate
		j := i + 1
		for j < len(cells) && cells[j].GetX() == currentX {
			currentCell.AddArea(cells[j].GetArea())
			currentCell.AddCover(cells[j].GetCover())
			j++
		}

		consolidated = append(consolidated, &currentCell)
		i = j
	}

	return consolidated
}

// renderHLine renders a horizontal line segment within a single scanline
// This is a complete implementation of the AGG render_hline algorithm
func (r *RasterizerCellsAASimple) renderHLine(ey, x1, y1, x2, y2 int) {
	ex1 := x1 >> basics.PolySubpixelShift
	ex2 := x2 >> basics.PolySubpixelShift
	fx1 := x1 & basics.PolySubpixelMask
	fx2 := x2 & basics.PolySubpixelMask

	var delta, p, first int
	var dx int64
	var incr, lift, mod, rem int

	// Trivial case - happens often
	if y1 == y2 {
		r.setCurrCell(ex2, ey)
		return
	}

	// Everything is located in a single cell - that is easy!
	if ex1 == ex2 {
		delta = y2 - y1
		r.currCell.AddCover(delta)
		r.currCell.AddArea((fx1 + fx2) * delta)
		return
	}

	// Ok, we'll have to render a run of adjacent cells on the same hline...
	p = (basics.PolySubpixelScale - fx1) * (y2 - y1)
	first = basics.PolySubpixelScale
	incr = 1

	dx = int64(x2) - int64(x1)

	if dx < 0 {
		p = fx1 * (y2 - y1)
		first = 0
		incr = -1
		dx = -dx
	}

	delta = int(int64(p) / dx)
	mod = int(int64(p) % dx)

	if mod < 0 {
		delta--
		mod += int(dx)
	}

	r.currCell.AddCover(delta)
	r.currCell.AddArea((fx1 + first) * delta)

	ex1 += incr
	r.setCurrCell(ex1, ey)
	y1 += delta

	if ex1 != ex2 {
		p = basics.PolySubpixelScale * (y2 - y1 + delta)
		lift = int(int64(p) / dx)
		rem = int(int64(p) % dx)

		if rem < 0 {
			lift--
			rem += int(dx)
		}

		mod -= int(dx)

		for ex1 != ex2 {
			delta = lift
			mod += rem
			if mod >= 0 {
				mod -= int(dx)
				delta++
			}

			r.currCell.AddCover(delta)
			r.currCell.AddArea(basics.PolySubpixelScale * delta)
			y1 += delta
			ex1 += incr
			r.setCurrCell(ex1, ey)
		}
	}

	delta = y2 - y1
	r.currCell.AddCover(delta)
	r.currCell.AddArea((fx2 + basics.PolySubpixelScale - first) * delta)
}
