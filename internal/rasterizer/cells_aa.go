package rasterizer

import (
	"math"

	"agg_go/internal/array"
	"agg_go/internal/basics"
)

// Cell block scale constants for memory management
const (
	CellBlockShift = 12
	CellBlockSize  = 1 << CellBlockShift
	CellBlockMask  = CellBlockSize - 1
	CellBlockPool  = 256
)

// SortedY represents a range of cells for a specific Y coordinate
type SortedY struct {
	Start uint32 // Starting index in sorted cells array
	Num   uint32 // Number of cells for this Y
}

// RasterizerCellsAA implements the main rasterization algorithm.
// This is an internal class used by the rasterizer and should not be used directly.
// It's equivalent to AGG's rasterizer_cells_aa<Cell> template class.
type RasterizerCellsAA[Cell CellInterface] struct {
	numBlocks      uint32
	maxBlocks      uint32
	currBlock      uint32
	numCells       uint32
	cellBlockLimit uint32
	cells          [][]*Cell                 // 2D array of cell blocks
	currCellPtr    *Cell                     // Current cell pointer
	sortedCells    *array.PodVector[*Cell]   // Sorted cell pointers
	sortedY        *array.PodVector[SortedY] // Y-coordinate ranges
	currCell       Cell                      // Current working cell
	styleCell      Cell                      // Style/template cell
	minX, minY     int                       // Bounding box minimum
	maxX, maxY     int                       // Bounding box maximum
	sorted         bool                      // Whether cells are sorted
}

// CellInterface defines the interface that cell types must implement
type CellInterface interface {
	Initial()
	Style(styleCell CellInterface)
	NotEqual(ex, ey int, styleCell CellInterface) int
	GetX() int
	GetY() int
	GetCover() int
	GetArea() int
	SetX(x int)
	SetY(y int)
	SetCover(cover int)
	SetArea(area int)
	AddCover(cover int)
	AddArea(area int)
}

// Ensure CellAA implements CellInterface
var _ CellInterface = (*CellAA)(nil)

// GetX returns the X coordinate
func (c *CellAA) GetX() int { return c.X }

// GetY returns the Y coordinate
func (c *CellAA) GetY() int { return c.Y }

// GetCover returns the coverage value
func (c *CellAA) GetCover() int { return c.Cover }

// GetArea returns the area value
func (c *CellAA) GetArea() int { return c.Area }

// SetX sets the X coordinate
func (c *CellAA) SetX(x int) { c.X = x }

// SetY sets the Y coordinate
func (c *CellAA) SetY(y int) { c.Y = y }

// SetCover sets the coverage value
func (c *CellAA) SetCover(cover int) { c.Cover = cover }

// SetArea sets the area value
func (c *CellAA) SetArea(area int) { c.Area = area }

// AddCover adds to the coverage value
func (c *CellAA) AddCover(cover int) { c.Cover += cover }

// AddArea adds to the area value
func (c *CellAA) AddArea(area int) { c.Area += area }

// NewRasterizerCellsAA creates a new cell-based rasterizer with the specified cell block limit
func NewRasterizerCellsAA[Cell CellInterface](cellBlockLimit uint32) *RasterizerCellsAA[Cell] {
	r := &RasterizerCellsAA[Cell]{
		numBlocks:      0,
		maxBlocks:      0,
		currBlock:      0,
		numCells:       0,
		cellBlockLimit: cellBlockLimit,
		cells:          nil,
		currCellPtr:    nil,
		sortedCells:    array.NewPodVector[*Cell](),
		sortedY:        array.NewPodVector[SortedY](),
		minX:           math.MaxInt32,
		minY:           math.MaxInt32,
		maxX:           math.MinInt32,
		maxY:           math.MinInt32,
		sorted:         false,
	}

	// Initialize cells properly for both value and pointer types
	var dummy Cell
	switch any(dummy).(type) {
	case *CellStyleAA:
		// For CellStyleAA pointer types, create new instances
		r.styleCell = any(&CellStyleAA{}).(Cell)
		r.currCell = any(&CellStyleAA{}).(Cell)
	case *CellAA:
		// For CellAA pointer types, create new instances
		r.styleCell = any(&CellAA{}).(Cell)
		r.currCell = any(&CellAA{}).(Cell)
	default:
		// For value types, use zero value
		r.styleCell = *new(Cell)
		r.currCell = *new(Cell)
	}
	r.styleCell.Initial()
	r.currCell.Initial()
	return r
}

// Reset clears all cells and resets the rasterizer state
func (r *RasterizerCellsAA[Cell]) Reset() {
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
func (r *RasterizerCellsAA[Cell]) Style(styleCell Cell) {
	r.styleCell = styleCell
}

// Line rasterizes a line from (x1,y1) to (x2,y2) using the AGG algorithm
func (r *RasterizerCellsAA[Cell]) Line(x1, y1, x2, y2 int) {
	const dxLimit = 16384 << basics.PolySubpixelShift

	dx := int64(x2) - int64(x1)

	// Split long lines to avoid arithmetic overflow
	if dx >= dxLimit || dx <= -dxLimit {
		cx := int((int64(x1) + int64(x2)) >> 1)
		cy := int((int64(y1) + int64(y2)) >> 1)
		r.Line(x1, y1, cx, cy)
		r.Line(cx, cy, x2, y2)
		return
	}

	dy := int64(y2) - int64(y1)
	ex1 := x1 >> basics.PolySubpixelShift
	ex2 := x2 >> basics.PolySubpixelShift
	ey1 := y1 >> basics.PolySubpixelShift
	ey2 := y2 >> basics.PolySubpixelShift
	fy1 := y1 & basics.PolySubpixelMask
	fy2 := y2 & basics.PolySubpixelMask

	// Update bounding box
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

	r.setCurrCell(ex1, ey1)

	// Single horizontal line case
	if ey1 == ey2 {
		r.renderHLine(ey1, x1, fy1, x2, fy2)
		return
	}

	// Vertical line - special case
	incr := 1
	if dx == 0 {
		ex := x1 >> basics.PolySubpixelShift
		twoFx := (x1 - (ex << basics.PolySubpixelShift)) << 1
		var area int

		first := basics.PolySubpixelScale
		if dy < 0 {
			first = 0
			incr = -1
		}

		// First cell
		delta := first - fy1
		r.currCell.AddCover(delta)
		r.currCell.AddArea(twoFx * delta)

		ey1 += incr
		r.setCurrCell(ex, ey1)

		// Middle cells
		delta = first + first - basics.PolySubpixelScale
		area = twoFx * delta
		for ey1 != ey2 {
			r.currCell.SetCover(delta)
			r.currCell.SetArea(area)
			ey1 += incr
			r.setCurrCell(ex, ey1)
		}

		// Last cell
		delta = fy2 - basics.PolySubpixelScale + first
		r.currCell.AddCover(delta)
		r.currCell.AddArea(twoFx * delta)
		return
	}

	// General case - render multiple hlines
	p := (basics.PolySubpixelScale - fy1) * int(dx)
	first := basics.PolySubpixelScale

	if dy < 0 {
		p = fy1 * int(dx)
		first = 0
		incr = -1
		dy = -dy
	}

	delta := p / int(dy)
	mod := p % int(dy)

	if mod < 0 {
		delta--
		mod += int(dy)
	}

	xFrom := x1 + delta
	r.renderHLine(ey1, x1, fy1, xFrom, first)

	ey1 += incr
	r.setCurrCell(xFrom>>basics.PolySubpixelShift, ey1)

	if ey1 != ey2 {
		p = basics.PolySubpixelScale * int(dx)
		lift := p / int(dy)
		rem := p % int(dy)

		if rem < 0 {
			lift--
			rem += int(dy)
		}
		mod -= int(dy)

		for ey1 != ey2 {
			delta = lift
			mod += rem
			if mod >= 0 {
				mod -= int(dy)
				delta++
			}

			xTo := xFrom + delta
			r.renderHLine(ey1, xFrom, basics.PolySubpixelScale-first, xTo, first)
			xFrom = xTo

			ey1 += incr
			r.setCurrCell(xFrom>>basics.PolySubpixelShift, ey1)
		}
	}
	r.renderHLine(ey1, xFrom, basics.PolySubpixelScale-first, x2, fy2)
}

// MinX returns the minimum X coordinate of the bounding box
func (r *RasterizerCellsAA[Cell]) MinX() int { return r.minX }

// MinY returns the minimum Y coordinate of the bounding box
func (r *RasterizerCellsAA[Cell]) MinY() int { return r.minY }

// MaxX returns the maximum X coordinate of the bounding box
func (r *RasterizerCellsAA[Cell]) MaxX() int { return r.maxX }

// MaxY returns the maximum Y coordinate of the bounding box
func (r *RasterizerCellsAA[Cell]) MaxY() int { return r.maxY }

// swapCells swaps two cell pointers
func swapCells[T any](a, b *T) {
	temp := *a
	*a = *b
	*b = temp
}

// qsortCells implements quicksort for cell pointers, sorting by X coordinate
func qsortCells[Cell CellInterface](start []*Cell, num int) {
	const qsortThreshold = 9
	var stack [80][2]int
	top := 0
	limit := num
	base := 0

	for {
		length := limit - base

		if length > qsortThreshold {
			// Use base + len/2 as pivot
			pivot := base + length/2
			swapCells(&start[base], &start[pivot])

			i := base + 1
			j := limit - 1

			// Ensure *i <= *base <= *j
			if (*start[j]).GetX() < (*start[i]).GetX() {
				swapCells(&start[i], &start[j])
			}

			if (*start[base]).GetX() < (*start[i]).GetX() {
				swapCells(&start[base], &start[i])
			}

			if (*start[j]).GetX() < (*start[base]).GetX() {
				swapCells(&start[base], &start[j])
			}

			for {
				x := (*start[base]).GetX()
				for i++; (*start[i]).GetX() < x; i++ {
				}
				for j--; x < (*start[j]).GetX(); j-- {
				}

				if i > j {
					break
				}

				swapCells(&start[i], &start[j])
			}

			swapCells(&start[base], &start[j])

			// Push largest sub-array
			if j-base > limit-i {
				stack[top][0] = base
				stack[top][1] = j
				base = i
			} else {
				stack[top][0] = i
				stack[top][1] = limit
				limit = j
			}
			top++
		} else {
			// Insertion sort for small arrays
			for i := base + 1; i < limit; i++ {
				j := i
				for j > base && (*start[j]).GetX() < (*start[j-1]).GetX() {
					swapCells(&start[j], &start[j-1])
					j--
				}
			}

			if top > 0 {
				top--
				base = stack[top][0]
				limit = stack[top][1]
			} else {
				break
			}
		}
	}
}

// SortCells sorts all cells by Y coordinate and then by X coordinate
func (r *RasterizerCellsAA[Cell]) SortCells() {
	if r.sorted {
		return // Perform sort only the first time
	}

	r.addCurrCell()
	r.currCell.SetX(math.MaxInt32)
	r.currCell.SetY(math.MaxInt32)
	r.currCell.SetCover(0)
	r.currCell.SetArea(0)

	if r.numCells == 0 {
		return
	}

	// Allocate the array of cell pointers
	r.sortedCells.Allocate(int(r.numCells), 16)

	// Allocate and zero the Y array
	yRange := r.maxY - r.minY + 1
	r.sortedY.Allocate(yRange, 16)
	r.sortedY.Zero()

	// Create Y-histogram (count cells for each Y)
	nb := int(r.numCells)
	blockIdx := 0
	for nb > 0 {
		blockSize := CellBlockSize
		if nb < CellBlockSize {
			blockSize = nb
		}

		for i := 0; i < blockSize; i++ {
			cell := r.cells[blockIdx][i]
			yIdx := (*cell).GetY() - r.minY
			// Skip cells with invalid Y coordinates (uninitialized cells)
			if yIdx < 0 || yIdx >= yRange {
				continue
			}
			sortedRange := r.sortedY.At(yIdx)
			sortedRange.Start++
			r.sortedY.Set(yIdx, sortedRange)
		}
		blockIdx++
		nb -= blockSize
	}

	// Convert Y-histogram to starting indexes
	start := uint32(0)
	for i := 0; i < yRange; i++ {
		sortedRange := r.sortedY.At(i)
		v := sortedRange.Start
		sortedRange.Start = start
		start += v
		r.sortedY.Set(i, sortedRange)
	}

	// Fill cell pointer array sorted by Y
	nb = int(r.numCells)
	blockIdx = 0
	for nb > 0 {
		blockSize := CellBlockSize
		if nb < CellBlockSize {
			blockSize = nb
		}

		for i := 0; i < blockSize; i++ {
			cell := r.cells[blockIdx][i]
			yIdx := (*cell).GetY() - r.minY
			// Skip cells with invalid Y coordinates (uninitialized cells)
			if yIdx < 0 || yIdx >= yRange {
				continue
			}
			currY := r.sortedY.At(yIdx)
			r.sortedCells.Set(int(currY.Start+currY.Num), cell)
			currY.Num++
			r.sortedY.Set(yIdx, currY)
		}
		blockIdx++
		nb -= blockSize
	}

	// Sort cells by X within each Y
	for i := 0; i < yRange; i++ {
		currY := r.sortedY.At(i)
		if currY.Num > 0 {
			startIdx := int(currY.Start)
			cells := make([]*Cell, currY.Num)
			for j := uint32(0); j < currY.Num; j++ {
				cells[j] = r.sortedCells.At(startIdx + int(j))
			}
			qsortCells(cells, int(currY.Num))
			for j := uint32(0); j < currY.Num; j++ {
				r.sortedCells.Set(startIdx+int(j), cells[j])
			}
		}
	}

	r.sorted = true
}

// TotalCells returns the total number of cells
func (r *RasterizerCellsAA[Cell]) TotalCells() uint32 {
	return r.numCells
}

// ScanlineNumCells returns the number of cells for the given scanline Y
func (r *RasterizerCellsAA[Cell]) ScanlineNumCells(y uint32) uint32 {
	if !r.sorted || int(y) < r.minY || int(y) > r.maxY {
		return 0
	}
	return r.sortedY.At(int(y - uint32(r.minY))).Num
}

// ScanlineCells returns the cells for the given scanline Y
func (r *RasterizerCellsAA[Cell]) ScanlineCells(y uint32) []*Cell {
	if !r.sorted || int(y) < r.minY || int(y) > r.maxY {
		return nil
	}

	sortedRange := r.sortedY.At(int(y - uint32(r.minY)))
	cells := make([]*Cell, sortedRange.Num)

	for i := uint32(0); i < sortedRange.Num; i++ {
		cells[i] = r.sortedCells.At(int(sortedRange.Start + i))
	}

	return cells
}

// Sorted returns whether the cells have been sorted
func (r *RasterizerCellsAA[Cell]) Sorted() bool {
	return r.sorted
}

// setCurrCell sets the current cell coordinates, adding the previous cell if needed
func (r *RasterizerCellsAA[Cell]) setCurrCell(x, y int) {
	if r.currCell.NotEqual(x, y, r.styleCell) != 0 {
		r.addCurrCell()
		r.currCell.Style(r.styleCell)
		r.currCell.SetX(x)
		r.currCell.SetY(y)
		r.currCell.SetCover(0)
		r.currCell.SetArea(0)
	}
}

// addCurrCell adds the current cell to the cell list if it has coverage or area
func (r *RasterizerCellsAA[Cell]) addCurrCell() {
	if r.currCell.GetArea() != 0 || r.currCell.GetCover() != 0 {
		if (r.numCells & CellBlockMask) == 0 {
			if r.numBlocks >= r.cellBlockLimit {
				return
			}
			r.allocateBlock()
		}

		// Store cell in current block
		blockIndex := r.currBlock - 1
		cellIndex := r.numCells & CellBlockMask

		// Create a new cell instance and copy the current cell data
		var newCell Cell
		switch any(r.currCell).(type) {
		case *CellAA:
			cell := &CellAA{
				X:     r.currCell.GetX(),
				Y:     r.currCell.GetY(),
				Cover: r.currCell.GetCover(),
				Area:  r.currCell.GetArea(),
			}
			newCell = any(cell).(Cell)
		case *CellStyleAA:
			cell := &CellStyleAA{}
			cell.SetX(r.currCell.GetX())
			cell.SetY(r.currCell.GetY())
			cell.SetCover(r.currCell.GetCover())
			cell.SetArea(r.currCell.GetArea())
			newCell = any(cell).(Cell)
		default:
			// For value types, create a copy
			newCell = r.currCell
		}

		r.cells[blockIndex][cellIndex] = &newCell
		r.numCells++
	}
}

// allocateBlock allocates a new cell block
func (r *RasterizerCellsAA[Cell]) allocateBlock() {
	if r.currBlock >= r.numBlocks {
		if r.numBlocks >= r.maxBlocks {
			// Expand blocks array
			newMaxBlocks := r.maxBlocks + CellBlockPool
			newCells := make([][]*Cell, newMaxBlocks)

			if r.cells != nil {
				copy(newCells, r.cells)
			}

			r.cells = newCells
			r.maxBlocks = newMaxBlocks
		}

		// Allocate new block
		r.cells[r.numBlocks] = make([]*Cell, CellBlockSize)
		r.numBlocks++
	}

	r.currBlock++
}

// renderHLine renders a horizontal line segment within a single scanline
func (r *RasterizerCellsAA[Cell]) renderHLine(ey, x1, y1, x2, y2 int) {
	ex1 := x1 >> basics.PolySubpixelShift
	ex2 := x2 >> basics.PolySubpixelShift
	fx1 := x1 & basics.PolySubpixelMask
	fx2 := x2 & basics.PolySubpixelMask

	var delta, p, first int
	var dx int64
	var incr, lift, mod, rem int

	// Horizontal line at exact pixel boundary
	if y1 == y2 {
		if ex1 == ex2 {
			// Single pixel - no span, just set position
			r.setCurrCell(ex2, ey)
			return
		}
		// Multi-pixel horizontal span - should generate cells with minimal coverage
		// For lines at pixel boundaries, we give a small cover value to ensure cells are added
		incr := 1
		if ex1 > ex2 {
			incr = -1
		}
		for ex := ex1; ex != ex2; ex += incr {
			// Add minimal coverage so the cell gets registered
			r.currCell.AddCover(1)
			r.currCell.AddArea(0) // No area for horizontal line at boundary
			r.setCurrCell(ex+incr, ey)
		}
		return
	}

	// Single cell case
	if ex1 == ex2 {
		delta = y2 - y1
		r.currCell.AddCover(delta)
		r.currCell.AddArea((fx1 + fx2) * delta)
		return
	}

	// Multi-cell case - run of adjacent cells on same hline
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
