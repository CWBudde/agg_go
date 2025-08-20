package ctrl

import (
	"testing"

	"agg_go/internal/basics"
	"agg_go/internal/transform"
)

func TestNewBaseCtrl(t *testing.T) {
	x1, y1, x2, y2 := 10.0, 20.0, 110.0, 60.0
	flipY := true

	ctrl := NewBaseCtrl(x1, y1, x2, y2, flipY)

	if ctrl.X1() != x1 || ctrl.Y1() != y1 || ctrl.X2() != x2 || ctrl.Y2() != y2 {
		t.Errorf("Expected bounds (%.1f, %.1f, %.1f, %.1f), got (%.1f, %.1f, %.1f, %.1f)",
			x1, y1, x2, y2, ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}

	if ctrl.FlipY() != flipY {
		t.Errorf("Expected FlipY %v, got %v", flipY, ctrl.FlipY())
	}
}

func TestBaseCtrlSetBounds(t *testing.T) {
	ctrl := NewBaseCtrl(0, 0, 100, 100, false)

	newX1, newY1, newX2, newY2 := 50.0, 60.0, 150.0, 160.0
	ctrl.SetBounds(newX1, newY1, newX2, newY2)

	if ctrl.X1() != newX1 || ctrl.Y1() != newY1 || ctrl.X2() != newX2 || ctrl.Y2() != newY2 {
		t.Errorf("Expected bounds (%.1f, %.1f, %.1f, %.1f), got (%.1f, %.1f, %.1f, %.1f)",
			newX1, newY1, newX2, newY2, ctrl.X1(), ctrl.Y1(), ctrl.X2(), ctrl.Y2())
	}
}

func TestBaseCtrlTransformation(t *testing.T) {
	ctrl := NewBaseCtrl(0, 0, 100, 100, false)

	// Test without transformation
	if ctrl.Scale() != 1.0 {
		t.Errorf("Expected scale 1.0 without transformation, got %.3f", ctrl.Scale())
	}

	// Test with transformation
	mtx := transform.NewTransAffine()
	mtx.ScaleXY(2.0, 2.0)
	ctrl.SetTransform(mtx)

	if ctrl.Scale() != 2.0 {
		t.Errorf("Expected scale 2.0 with 2x scale transformation, got %.3f", ctrl.Scale())
	}

	// Test coordinate transformation
	x, y := 50.0, 50.0
	origX, origY := x, y
	ctrl.TransformXY(&x, &y)

	expectedX, expectedY := 100.0, 100.0 // 2x scaling
	if x != expectedX || y != expectedY {
		t.Errorf("Expected transformed coordinates (%.1f, %.1f), got (%.1f, %.1f)",
			expectedX, expectedY, x, y)
	}

	// Test inverse transformation
	ctrl.InverseTransformXY(&x, &y)
	if x != origX || y != origY {
		t.Errorf("Expected inverse transform to restore (%.1f, %.1f), got (%.1f, %.1f)",
			origX, origY, x, y)
	}

	// Test clearing transformation
	ctrl.ClearTransform()
	if ctrl.Scale() != 1.0 {
		t.Errorf("Expected scale 1.0 after clearing transformation, got %.3f", ctrl.Scale())
	}
}

func TestBaseCtrlYFlipping(t *testing.T) {
	ctrl := NewBaseCtrl(0, 0, 100, 100, true) // Enable Y flipping

	x, y := 50.0, 25.0
	ctrl.TransformXY(&x, &y)

	// With Y flipping: y' = y1 + y2 - y = 0 + 100 - 25 = 75
	expectedY := 75.0
	if y != expectedY {
		t.Errorf("Expected Y flipped coordinate %.1f, got %.1f", expectedY, y)
	}

	// Test inverse transformation
	ctrl.InverseTransformXY(&x, &y)
	expectedY = 25.0 // Should restore original
	if y != expectedY {
		t.Errorf("Expected inverse Y flip to restore %.1f, got %.1f", expectedY, y)
	}
}

func TestBaseCtrlInRect(t *testing.T) {
	ctrl := NewBaseCtrl(10, 20, 110, 120, false)

	testCases := []struct {
		x, y     float64
		expected bool
		name     string
	}{
		{50, 70, true, "center point"},
		{10, 20, true, "top-left corner"},
		{110, 120, true, "bottom-right corner"},
		{5, 70, false, "left of bounds"},
		{115, 70, false, "right of bounds"},
		{50, 15, false, "above bounds"},
		{50, 125, false, "below bounds"},
	}

	for _, tc := range testCases {
		result := ctrl.InRect(tc.x, tc.y)
		if result != tc.expected {
			t.Errorf("InRect(%.1f, %.1f) for %s: expected %v, got %v",
				tc.x, tc.y, tc.name, tc.expected, result)
		}
	}
}

func TestBaseCtrlInRectWithTransformation(t *testing.T) {
	ctrl := NewBaseCtrl(0, 0, 100, 100, false)

	// Apply translation transformation
	mtx := transform.NewTransAffine()
	mtx.Translate(50, 50)
	ctrl.SetTransform(mtx)

	// Point that should be inside after inverse transformation
	// Screen coordinate (100, 100) should map to control coordinate (50, 50)
	result := ctrl.InRect(100, 100)
	if !result {
		t.Error("Expected point (100, 100) to be inside bounds after translation")
	}

	// Point that should be outside
	result = ctrl.InRect(25, 25)
	if result {
		t.Error("Expected point (25, 25) to be outside bounds after translation")
	}
}

// Mock control implementation for testing the Ctrl interface
type mockCtrl struct {
	*BaseCtrl
	numPaths uint
	colors   []interface{}
}

func newMockCtrl(x1, y1, x2, y2 float64) *mockCtrl {
	return &mockCtrl{
		BaseCtrl: NewBaseCtrl(x1, y1, x2, y2, false),
		numPaths: 2,
		colors:   []interface{}{"red", "blue"},
	}
}

func (mc *mockCtrl) OnMouseButtonDown(x, y float64) bool         { return false }
func (mc *mockCtrl) OnMouseButtonUp(x, y float64) bool           { return false }
func (mc *mockCtrl) OnMouseMove(x, y float64, pressed bool) bool { return false }
func (mc *mockCtrl) OnArrowKeys(left, right, down, up bool) bool { return false }

func (mc *mockCtrl) NumPaths() uint     { return mc.numPaths }
func (mc *mockCtrl) Rewind(pathID uint) {}
func (mc *mockCtrl) Vertex() (x, y float64, cmd basics.PathCommand) {
	return 0, 0, basics.PathCmdStop
}
func (mc *mockCtrl) Color(pathID uint) interface{} {
	if pathID < uint(len(mc.colors)) {
		return mc.colors[pathID]
	}
	return "default"
}

func TestCtrlInterface(t *testing.T) {
	// Test that mockCtrl implements Ctrl interface
	var ctrl Ctrl = newMockCtrl(0, 0, 100, 100)

	// Test basic interface methods
	if ctrl.NumPaths() != 2 {
		t.Errorf("Expected 2 paths, got %d", ctrl.NumPaths())
	}

	if ctrl.Color(0) != "red" {
		t.Errorf("Expected color 'red', got %v", ctrl.Color(0))
	}

	if ctrl.Color(1) != "blue" {
		t.Errorf("Expected color 'blue', got %v", ctrl.Color(1))
	}

	if ctrl.Color(2) != "default" {
		t.Errorf("Expected default color, got %v", ctrl.Color(2))
	}
}

func TestVertexIterator(t *testing.T) {
	ctrl := newMockCtrl(0, 0, 100, 100)
	iter := NewVertexIterator(ctrl, 0)

	if iter.PathID() != 0 {
		t.Errorf("Expected path ID 0, got %d", iter.PathID())
	}

	if iter.VertexIndex() != 0 {
		t.Errorf("Expected vertex index 0, got %d", iter.VertexIndex())
	}

	// Test iteration (will immediately return stop command from mock)
	_, _, cmd, done := iter.Next()
	if !done {
		t.Error("Expected iteration to be done immediately with mock control")
	}
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop, got %v", cmd)
	}

	// Test reset
	iter.Reset()
	if iter.VertexIndex() != 0 {
		t.Errorf("Expected vertex index 0 after reset, got %d", iter.VertexIndex())
	}
}
