package rasterizer

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

// CellInterface defines the interface that cell types must implement
// This interface is kept for documentation purposes, but concrete types should be used
// directly (RasterizerCellsAASimple for CellAA, RasterizerCellsAAStyled for CellStyleAA).
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

// The following generic type and methods have been removed in favor of concrete implementations:
// - RasterizerCellsAA[Cell] generic type (had problematic type switches at lines 118-126, 568-583)
// - NewRasterizerCellsAA[Cell] constructor
// - All generic methods (Reset, Style, Line, SortCells, etc.)
//
// Use these concrete replacements instead:
// - RasterizerCellsAASimple for CellAA (see cells_aa_simple.go)
// - RasterizerCellsAAStyled for CellStyleAA (see cells_aa_styled.go)
//
// This removal eliminates problematic runtime type assertions (any() casts) that
// violated the project's design principles.

// qsortCells implements quicksort for cell pointers, sorting by X coordinate
// This is a shared helper function used by both RasterizerCellsAASimple and RasterizerCellsAAStyled
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
			start[base], start[pivot] = start[pivot], start[base]

			i := base + 1
			j := limit - 1

			// Ensure *i <= *base <= *j
			if (*start[j]).GetX() < (*start[i]).GetX() {
				start[i], start[j] = start[j], start[i]
			}

			if (*start[base]).GetX() < (*start[i]).GetX() {
				start[base], start[i] = start[i], start[base]
			}

			if (*start[j]).GetX() < (*start[base]).GetX() {
				start[base], start[j] = start[j], start[base]
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

				start[i], start[j] = start[j], start[i]
			}

			start[base], start[j] = start[j], start[base]

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
					start[j], start[j-1] = start[j-1], start[j]
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
