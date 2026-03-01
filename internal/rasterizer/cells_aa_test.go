package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

func TestRasterizerCellsAA_Basic(t *testing.T) {
	// Create a rasterizer for basic cells
	rasterizer := NewRasterizerCellsAASimple(1024)

	// Test initial state
	if rasterizer.TotalCells() != 0 {
		t.Errorf("Expected 0 initial cells, got %d", rasterizer.TotalCells())
	}

	if rasterizer.Sorted() {
		t.Error("Expected rasterizer to not be sorted initially")
	}

	// Test line rasterization
	x1, y1 := 10<<basics.PolySubpixelShift, 10<<basics.PolySubpixelShift
	x2, y2 := 20<<basics.PolySubpixelShift, 20<<basics.PolySubpixelShift

	rasterizer.Line(x1, y1, x2, y2)

	// Should have some cells now
	if rasterizer.TotalCells() == 0 {
		t.Error("Expected cells after line rasterization")
	}

	// Test bounding box
	minX, minY := rasterizer.MinX(), rasterizer.MinY()
	maxX, maxY := rasterizer.MaxX(), rasterizer.MaxY()

	if minX > maxX || minY > maxY {
		t.Errorf("Invalid bounding box: (%d,%d) to (%d,%d)", minX, minY, maxX, maxY)
	}

	// Test sorting
	rasterizer.SortCells()
	if !rasterizer.Sorted() {
		t.Error("Expected rasterizer to be sorted after SortCells()")
	}

	// Test scanline access
	for y := minY; y <= maxY; y++ {
		numCells := rasterizer.ScanlineNumCells(y)
		if numCells > 0 {
			cells := rasterizer.ScanlineCells(y)
			if len(cells) != int(numCells) {
				t.Errorf("Scanline %d: expected %d cells, got %d", y, numCells, len(cells))
			}

			// Verify cells are sorted by X
			for i := 1; i < len(cells); i++ {
				if (*cells[i]).GetX() < (*cells[i-1]).GetX() {
					t.Errorf("Scanline %d: cells not sorted by X", y)
					break
				}
			}
		}
	}
}

func TestRasterizerCellsAA_Reset(t *testing.T) {
	rasterizer := NewRasterizerCellsAASimple(1024)

	// Add some data
	x1, y1 := 5<<basics.PolySubpixelShift, 5<<basics.PolySubpixelShift
	x2, y2 := 15<<basics.PolySubpixelShift, 15<<basics.PolySubpixelShift
	rasterizer.Line(x1, y1, x2, y2)

	// Reset should clear everything
	rasterizer.Reset()

	if rasterizer.TotalCells() != 0 {
		t.Errorf("Expected 0 cells after reset, got %d", rasterizer.TotalCells())
	}

	if rasterizer.Sorted() {
		t.Error("Expected rasterizer to not be sorted after reset")
	}
}

func TestRasterizerCellsAA_VerticalLine(t *testing.T) {
	rasterizer := NewRasterizerCellsAASimple(1024)

	// Test vertical line (special case in Line method)
	x := 10 << basics.PolySubpixelShift
	y1 := 5 << basics.PolySubpixelShift
	y2 := 15 << basics.PolySubpixelShift

	rasterizer.Line(x, y1, x, y2)

	if rasterizer.TotalCells() == 0 {
		t.Error("Expected cells after vertical line rasterization")
	}

	rasterizer.SortCells()

	// All cells should have the same X coordinate
	expectedX := x >> basics.PolySubpixelShift
	for y := rasterizer.MinY(); y <= rasterizer.MaxY(); y++ {
		cells := rasterizer.ScanlineCells(y)
		for _, cell := range cells {
			if (*cell).GetX() != expectedX {
				t.Errorf("Vertical line: expected X=%d, got X=%d", expectedX, (*cell).GetX())
			}
		}
	}
}

func TestRasterizerCellsAA_HorizontalLine(t *testing.T) {
	rasterizer := NewRasterizerCellsAASimple(1024)

	// Test horizontal line
	x1 := 5 << basics.PolySubpixelShift
	x2 := 15 << basics.PolySubpixelShift
	y := 10 << basics.PolySubpixelShift

	rasterizer.Line(x1, y, x2, y)
	rasterizer.SortCells()

	if rasterizer.TotalCells() != 0 {
		t.Errorf("Expected no accumulated cells for an exact horizontal edge, got %d", rasterizer.TotalCells())
	}
}

func TestRasterizerCellsAA_SortPreservesDuplicateXCells(t *testing.T) {
	rasterizer := NewRasterizerCellsAASimple(1024)

	rasterizer.currCell = CellAA{X: 7, Y: 3, Cover: 1, Area: 2}
	rasterizer.addCurrCell()
	rasterizer.currCell = CellAA{X: 7, Y: 3, Cover: 3, Area: 4}
	rasterizer.addCurrCell()
	rasterizer.currCell.Initial()

	rasterizer.SortCells()

	if got := rasterizer.ScanlineNumCells(3); got != 2 {
		t.Fatalf("Expected duplicate X cells to remain after sorting, got %d cells", got)
	}

	cells := rasterizer.ScanlineCells(3)
	if len(cells) != 2 {
		t.Fatalf("Expected 2 scanline cells, got %d", len(cells))
	}
	if cells[0].X != 7 || cells[1].X != 7 {
		t.Fatalf("Expected both sorted cells to keep X=7, got %d and %d", cells[0].X, cells[1].X)
	}
}
