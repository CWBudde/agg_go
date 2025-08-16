package rasterizer

import (
	"math"
	"testing"
)

func TestCellAA_Initial(t *testing.T) {
	var cell CellAA
	cell.Initial()

	if cell.X != math.MaxInt32 {
		t.Errorf("Expected X to be MaxInt32, got %d", cell.X)
	}
	if cell.Y != math.MaxInt32 {
		t.Errorf("Expected Y to be MaxInt32, got %d", cell.Y)
	}
	if cell.Cover != 0 {
		t.Errorf("Expected Cover to be 0, got %d", cell.Cover)
	}
	if cell.Area != 0 {
		t.Errorf("Expected Area to be 0, got %d", cell.Area)
	}
}

func TestCellAA_Style(t *testing.T) {
	var cell, styleCell CellAA

	// Style method should be a no-op for basic CellAA
	cell.Style(&styleCell)

	// Since it's a no-op, cell should remain unchanged
	if cell.X != 0 || cell.Y != 0 || cell.Cover != 0 || cell.Area != 0 {
		t.Error("Style method should not modify the cell")
	}
}

func TestCellAA_NotEqual(t *testing.T) {
	var cell CellAA
	cell.X = 10
	cell.Y = 20

	var styleCell CellAA

	// Same coordinates should return 0
	result := cell.NotEqual(10, 20, &styleCell)
	if result != 0 {
		t.Errorf("Expected 0 for same coordinates, got %d", result)
	}

	// Different X should return non-zero
	result = cell.NotEqual(11, 20, &styleCell)
	if result == 0 {
		t.Error("Expected non-zero for different X coordinate")
	}

	// Different Y should return non-zero
	result = cell.NotEqual(10, 21, &styleCell)
	if result == 0 {
		t.Error("Expected non-zero for different Y coordinate")
	}

	// Both different should return non-zero
	result = cell.NotEqual(11, 21, &styleCell)
	if result == 0 {
		t.Error("Expected non-zero for different coordinates")
	}
}

func TestCellAA_GettersSetters(t *testing.T) {
	var cell CellAA

	// Test setters
	cell.SetX(100)
	cell.SetY(200)
	cell.SetCover(128)
	cell.SetArea(256)

	// Test getters
	if cell.GetX() != 100 {
		t.Errorf("Expected X=100, got %d", cell.GetX())
	}
	if cell.GetY() != 200 {
		t.Errorf("Expected Y=200, got %d", cell.GetY())
	}
	if cell.GetCover() != 128 {
		t.Errorf("Expected Cover=128, got %d", cell.GetCover())
	}
	if cell.GetArea() != 256 {
		t.Errorf("Expected Area=256, got %d", cell.GetArea())
	}
}

func TestCellAA_AddMethods(t *testing.T) {
	var cell CellAA
	cell.SetCover(50)
	cell.SetArea(100)

	// Test AddCover
	cell.AddCover(25)
	if cell.GetCover() != 75 {
		t.Errorf("Expected Cover=75 after adding 25, got %d", cell.GetCover())
	}

	// Test AddArea
	cell.AddArea(50)
	if cell.GetArea() != 150 {
		t.Errorf("Expected Area=150 after adding 50, got %d", cell.GetArea())
	}

	// Test negative additions
	cell.AddCover(-30)
	if cell.GetCover() != 45 {
		t.Errorf("Expected Cover=45 after subtracting 30, got %d", cell.GetCover())
	}
}

func TestCellAA_Interface(t *testing.T) {
	var cell CellAA

	// Verify that CellAA implements CellInterface
	var _ CellInterface = &cell

	// Test interface methods
	cell.Initial()
	cell.Style(&cell)

	// After Initial(), coordinates are MaxInt32, so comparing with them should return 0
	result := cell.NotEqual(cell.X, cell.Y, &cell)

	if result != 0 {
		t.Error("Interface methods should work correctly")
	}
}
