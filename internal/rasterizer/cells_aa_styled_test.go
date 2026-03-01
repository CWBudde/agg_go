package rasterizer

import "testing"

func TestRasterizerCellsAAStyled_SortPreservesDuplicateStyledCells(t *testing.T) {
	rasterizer := NewRasterizerCellsAAStyled(1024)

	rasterizer.currCell = CellStyleAA{X: 9, Y: 4, Cover: 1, Area: 2, Left: 2, Right: 1}
	rasterizer.addCurrCell()
	rasterizer.currCell = CellStyleAA{X: 9, Y: 4, Cover: 3, Area: 5, Left: 2, Right: 1}
	rasterizer.addCurrCell()
	rasterizer.currCell.Initial()

	rasterizer.SortCells()

	if got := rasterizer.ScanlineNumCells(4); got != 2 {
		t.Fatalf("Expected duplicate styled cells to remain after sorting, got %d cells", got)
	}

	cells := rasterizer.ScanlineCells(4)
	if len(cells) != 2 {
		t.Fatalf("Expected 2 scanline cells, got %d", len(cells))
	}
	if cells[0].X != 9 || cells[1].X != 9 {
		t.Fatalf("Expected both sorted cells to keep X=9, got %d and %d", cells[0].X, cells[1].X)
	}
	if cells[0].Left != 2 || cells[1].Left != 2 || cells[0].Right != 1 || cells[1].Right != 1 {
		t.Fatal("Expected style metadata to remain unchanged after sorting")
	}
}
