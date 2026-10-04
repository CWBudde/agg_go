package rasterizer

import (
	"testing"

	"github.com/cwbudde/agg_go/internal/basics"
)

// Golden values below come from AGG 2.6 C++ rasterizer_scanline_aa<>,
// rasterizer_scanline_aa_nogamma<> and rasterizer_compound_aa<> (default
// rasterizer_sl_clip_int, no clip box; clang++ -O0 -ffp-contract=off). All
// three variants share rasterizer_cells_aa::sort_cells() and agree.

// sortParityNoClip mirrors rasterizer_sl_clip_int without a clip box for the
// compound rasterizer's float clipper interface.
type sortParityNoClip struct{ x1, y1 int }

func (c *sortParityNoClip) ResetClipping()                 {}
func (c *sortParityNoClip) ClipBox(x1, y1, x2, y2 float64) {}
func (c *sortParityNoClip) MoveTo(x, y float64) {
	c.x1 = basics.IRound(x * basics.PolySubpixelScale)
	c.y1 = basics.IRound(y * basics.PolySubpixelScale)
}

func (c *sortParityNoClip) LineTo(outline *RasterizerCellsAAStyled, x, y float64) {
	x2 := basics.IRound(x * basics.PolySubpixelScale)
	y2 := basics.IRound(y * basics.PolySubpixelScale)
	outline.Line(c.x1, c.y1, x2, y2)
	c.x1, c.y1 = x2, y2
}

type sortParityRas struct {
	moveTo, lineTo func(x, y float64)
	rewind         func() bool
	sorted         func() bool
	totalCells     func() uint32
	bounds         func() [4]int
	cellSums       func() (cover, area int64)
}

func simpleCellSums(o *RasterizerCellsAASimple) func() (int64, int64) {
	return func() (cs, as int64) {
		for i := uint32(0); i < o.numCells; i++ {
			c := &o.cells[i>>CellBlockShift][i&CellBlockMask]
			cs += int64(c.Cover)
			as += int64(c.Area)
		}
		return cs, as
	}
}

func styledCellSums(o *RasterizerCellsAAStyled) func() (int64, int64) {
	return func() (cs, as int64) {
		for i := uint32(0); i < o.numCells; i++ {
			c := &o.cells[i>>CellBlockShift][i&CellBlockMask]
			cs += int64(c.Cover)
			as += int64(c.Area)
		}
		return cs, as
	}
}

func newSortParityRasterizers() map[string]func() sortParityRas {
	return map[string]func() sortParityRas{
		"gamma": func() sortParityRas {
			r := NewRasterizerScanlineAA[int, IntConv, *RasterizerSlClip[int, IntConv]](
				IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{}))
			return sortParityRas{
				r.MoveToD, r.LineToD, r.RewindScanlines, r.outline.Sorted, r.outline.TotalCells,
				func() [4]int { return [4]int{r.MinX(), r.MinY(), r.MaxX(), r.MaxY()} },
				simpleCellSums(r.outline),
			}
		},
		"nogamma": func() sortParityRas {
			r := NewRasterizerScanlineAANoGamma[int, IntConv, *RasterizerSlClip[int, IntConv]](
				IntConv{}, NewRasterizerSlClip[int, IntConv](IntConv{}))
			return sortParityRas{
				r.MoveToD, r.LineToD, r.RewindScanlines, r.outline.Sorted, r.outline.TotalCells,
				func() [4]int { return [4]int{r.MinX(), r.MinY(), r.MaxX(), r.MaxY()} },
				simpleCellSums(r.outline),
			}
		},
		"compound": func() sortParityRas {
			r := NewRasterizerCompoundAA(&sortParityNoClip{})
			r.Styles(0, -1)
			return sortParityRas{
				r.MoveToD, r.LineToD, r.RewindScanlines, r.outline.Sorted, r.outline.TotalCells,
				func() [4]int { return [4]int{r.MinX(), r.MinY(), r.MaxX(), r.MaxY()} },
				styledCellSums(r.outline),
			}
		},
	}
}

// An empty sort_cells() returns before setting m_sorted, so the next move_to
// does not reset() and the zero-area horizontal line's bounds survive.
func TestSortCellsEmptyKeepsBoundsMatchesCpp(t *testing.T) {
	for name, mk := range newSortParityRasterizers() {
		t.Run(name, func(t *testing.T) {
			r := mk()
			r.moveTo(5.5, 50.5)
			r.lineTo(100.5, 50.5)
			if r.rewind() {
				t.Fatal("empty rewind: got true, want false (C++)")
			}
			if r.sorted() {
				t.Fatal("empty sort: sorted = true, want false (C++)")
			}
			r.moveTo(10.25, 10.5)
			r.lineTo(20.75, 10.5)
			r.lineTo(20.75, 20.25)
			r.lineTo(10.25, 10.5)
			if !r.rewind() {
				t.Fatal("rewind: got false, want true")
			}
			if got, want := r.bounds(), [4]int{5, 10, 100, 50}; got != want {
				t.Errorf("bounds = %v, want %v (C++)", got, want)
			}
			if got := r.totalCells(); got != 31 {
				t.Errorf("total cells = %d, want 31 (C++)", got)
			}
		})
	}
}

// sort_cells() parks the current cell at (INT_MAX, INT_MAX) with zero
// cover/area, so lines appended after sorting do not re-add the last
// already-flushed cell.
func TestSortCellsParksCurrentCellMatchesCpp(t *testing.T) {
	for name, mk := range newSortParityRasterizers() {
		t.Run(name, func(t *testing.T) {
			r := mk()
			r.moveTo(10.25, 10.5)
			r.lineTo(20.75, 10.5)
			r.lineTo(20.75, 20.25)
			r.lineTo(10.25, 10.5)
			r.rewind()
			if got := r.totalCells(); got != 31 {
				t.Fatalf("cells before = %d, want 31 (C++)", got)
			}
			r.lineTo(30.5, 30.5)
			r.lineTo(35.5, 12.5)
			if got := r.totalCells(); got != 94 {
				t.Errorf("cells after = %d, want 94 (C++)", got)
			}
			if cs, as := r.cellSums(); cs != 640 || as != 471356 {
				t.Errorf("stored cell cover/area sums = %d/%d, want 640/471356 (C++)", cs, as)
			}
		})
	}
}
