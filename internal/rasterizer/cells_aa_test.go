package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

func TestRasterizerCellsAA_Basic(t *testing.T) {
	// Create a rasterizer for basic cells
	rasterizer := NewRasterizerCellsAA[*CellAA](1024)

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
		numCells := rasterizer.ScanlineNumCells(uint32(y))
		if numCells > 0 {
			cells := rasterizer.ScanlineCells(uint32(y))
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
	rasterizer := NewRasterizerCellsAA[*CellAA](1024)

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
	rasterizer := NewRasterizerCellsAA[*CellAA](1024)

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
		cells := rasterizer.ScanlineCells(uint32(y))
		for _, cell := range cells {
			if (*cell).GetX() != expectedX {
				t.Errorf("Vertical line: expected X=%d, got X=%d", expectedX, (*cell).GetX())
			}
		}
	}
}

func TestRasterizerCellsAA_HorizontalLine(t *testing.T) {
	rasterizer := NewRasterizerCellsAA[*CellAA](1024)

	// Test horizontal line
	x1 := 5 << basics.PolySubpixelShift
	x2 := 15 << basics.PolySubpixelShift
	y := 10 << basics.PolySubpixelShift

	rasterizer.Line(x1, y, x2, y)

	if rasterizer.TotalCells() == 0 {
		t.Error("Expected cells after horizontal line rasterization")
	}

	rasterizer.SortCells()

	// All cells should have the same Y coordinate
	expectedY := y >> basics.PolySubpixelShift
	if rasterizer.MinY() != expectedY || rasterizer.MaxY() != expectedY {
		t.Errorf("Horizontal line: expected Y range [%d,%d], got [%d,%d]",
			expectedY, expectedY, rasterizer.MinY(), rasterizer.MaxY())
	}
}
