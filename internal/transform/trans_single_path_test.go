package transform

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

// Mock vertex source for testing AddPath functionality
type mockVertexSource struct {
	vertices []vertex
	index    int
}

type vertex struct {
	x, y float64
	cmd  basics.PathCommand
}

func newMockVertexSource(vertices []vertex) *mockVertexSource {
	return &mockVertexSource{
		vertices: vertices,
		index:    0,
	}
}

func (mvs *mockVertexSource) Rewind(pathID uint) {
	mvs.index = 0
}

func (mvs *mockVertexSource) Vertex() (x, y float64, cmd basics.PathCommand) {
	if mvs.index >= len(mvs.vertices) {
		return 0, 0, basics.PathCmdStop
	}

	v := mvs.vertices[mvs.index]
	mvs.index++
	return v.x, v.y, v.cmd
}

func TestTransSinglePath_NewTransSinglePath(t *testing.T) {
	tsp := NewTransSinglePath()

	if tsp == nil {
		t.Fatal("NewTransSinglePath returned nil")
	}

	if tsp.BaseLength() != 0.0 {
		t.Errorf("Expected base length 0.0, got %f", tsp.BaseLength())
	}

	if !tsp.PreserveXScale() {
		t.Error("Expected preserve X scale to be true by default")
	}

	if tsp.TotalLength() != 0.0 {
		t.Errorf("Expected total length 0.0 for empty path, got %f", tsp.TotalLength())
	}
}

func TestTransSinglePath_BaseLength(t *testing.T) {
	tsp := NewTransSinglePath()

	// Test setting and getting base length
	tsp.SetBaseLength(100.0)
	if tsp.BaseLength() != 100.0 {
		t.Errorf("Expected base length 100.0, got %f", tsp.BaseLength())
	}

	// Test that total length returns base length when set
	if tsp.TotalLength() != 100.0 {
		t.Errorf("Expected total length to return base length 100.0, got %f", tsp.TotalLength())
	}
}

func TestTransSinglePath_PreserveXScale(t *testing.T) {
	tsp := NewTransSinglePath()

	// Test default value
	if !tsp.PreserveXScale() {
		t.Error("Expected preserve X scale to be true by default")
	}

	// Test setting to false
	tsp.SetPreserveXScale(false)
	if tsp.PreserveXScale() {
		t.Error("Expected preserve X scale to be false after setting")
	}

	// Test setting back to true
	tsp.SetPreserveXScale(true)
	if !tsp.PreserveXScale() {
		t.Error("Expected preserve X scale to be true after setting")
	}
}

func TestTransSinglePath_BasicPath(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a simple horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test total length
	expectedLength := 100.0
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected total length %f, got %f", expectedLength, tsp.TotalLength())
	}

	// Test transformation at the start
	x, y := 0.0, 0.0
	tsp.Transform(&x, &y)
	if math.Abs(x-0.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (0,0) at start, got (%f,%f)", x, y)
	}

	// Test transformation at the end
	x, y = 100.0, 0.0
	tsp.Transform(&x, &y)
	if math.Abs(x-100.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (100,0) at end, got (%f,%f)", x, y)
	}

	// Test transformation in the middle
	x, y = 50.0, 0.0
	tsp.Transform(&x, &y)
	if math.Abs(x-50.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (50,0) at middle, got (%f,%f)", x, y)
	}
}

func TestTransSinglePath_PerpendicularOffset(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test perpendicular offset at middle of path
	x, y := 50.0, 10.0 // 10 units perpendicular to horizontal line
	tsp.Transform(&x, &y)

	// For a horizontal line, perpendicular offset should be in Y direction
	expectedX := 50.0
	expectedY := 10.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) with perpendicular offset, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_VerticalLine(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a vertical line from (0,0) to (0,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(0, 100)
	tsp.FinalizePath()

	// Test total length
	expectedLength := 100.0
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected total length %f, got %f", expectedLength, tsp.TotalLength())
	}

	// Test transformation at middle with perpendicular offset
	x, y := 50.0, 10.0 // 10 units perpendicular to vertical line
	tsp.Transform(&x, &y)

	// For a vertical line going up, perpendicular offset goes to the left (negative X)
	expectedX := -10.0
	expectedY := 50.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) with perpendicular offset on vertical line, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_DiagonalLine(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a diagonal line from (0,0) to (100,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 100)
	tsp.FinalizePath()

	// Test total length (should be sqrt(2) * 100)
	expectedLength := math.Sqrt(2) * 100.0
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected total length %f, got %f", expectedLength, tsp.TotalLength())
	}

	// Test transformation at middle point
	x, y := expectedLength/2, 0.0
	tsp.Transform(&x, &y)

	// Should be at (50, 50)
	expectedX := 50.0
	expectedY := 50.0
	if math.Abs(x-expectedX) > 1e-9 || math.Abs(y-expectedY) > 1e-9 {
		t.Errorf("Expected (%f,%f) at middle of diagonal, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_MultiSegmentPath(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create an L-shaped path: (0,0) -> (100,0) -> (100,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.LineTo(100, 100)
	tsp.FinalizePath()

	// Test total length (should be 200)
	expectedLength := 200.0
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected total length %f, got %f", expectedLength, tsp.TotalLength())
	}

	// Test transformation at the corner (distance 100)
	x, y := 100.0, 0.0
	tsp.Transform(&x, &y)

	// Should be at (100, 0)
	expectedX := 100.0
	expectedY := 0.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) at corner, got (%f,%f)", expectedX, expectedY, x, y)
	}

	// Test transformation at 3/4 of the path (distance 150)
	x, y = 150.0, 0.0
	tsp.Transform(&x, &y)

	// Should be at (100, 50)
	expectedX = 100.0
	expectedY = 50.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) at 3/4 of path, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_Extrapolation(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test extrapolation before the start (negative distance)
	x, y := -50.0, 0.0
	tsp.Transform(&x, &y)

	// Should extrapolate backwards along the line
	expectedX := -50.0
	expectedY := 0.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) for backward extrapolation, got (%f,%f)", expectedX, expectedY, x, y)
	}

	// Test extrapolation after the end
	x, y = 150.0, 0.0
	tsp.Transform(&x, &y)

	// Should extrapolate forward along the line
	expectedX = 150.0
	expectedY = 0.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) for forward extrapolation, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_PreserveXScaleModes(t *testing.T) {
	// Test both preserve X scale modes with the same path
	tsp1 := NewTransSinglePath()
	tsp1.SetPreserveXScale(true)

	tsp2 := NewTransSinglePath()
	tsp2.SetPreserveXScale(false)

	// Create identical paths
	for _, tsp := range []*TransSinglePath{tsp1, tsp2} {
		tsp.MoveTo(0, 0)
		tsp.LineTo(50, 0)
		tsp.LineTo(100, 50)
		tsp.FinalizePath()
	}

	// Test transformation at the same point
	x1, y1 := 25.0, 0.0
	x2, y2 := 25.0, 0.0

	tsp1.Transform(&x1, &y1)
	tsp2.Transform(&x2, &y2)

	// Results should be similar (within reasonable tolerance)
	// The exact values may differ due to different interpolation methods
	// (binary search vs uniform distribution)
	if math.Abs(x1-x2) > 5.0 || math.Abs(y1-y2) > 5.0 {
		t.Errorf("Results too different between preserve X scale modes: (%f,%f) vs (%f,%f)", x1, y1, x2, y2)
	}
}

func TestTransSinglePath_AddPath(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a mock vertex source with a simple path
	vertices := []vertex{
		{0, 0, basics.PathCmdMoveTo},
		{100, 0, basics.PathCmdLineTo},
		{100, 100, basics.PathCmdLineTo},
		{0, 0, basics.PathCmdStop},
	}

	mvs := newMockVertexSource(vertices)
	tsp.AddPath(mvs, 0)

	// Test that the path was added correctly
	expectedLength := 200.0 // 100 + 100
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected total length %f after AddPath, got %f", expectedLength, tsp.TotalLength())
	}

	// Test transformation at a known point
	x, y := 150.0, 0.0 // 3/4 along the path
	tsp.Transform(&x, &y)

	// Should be at (100, 50)
	expectedX := 100.0
	expectedY := 50.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) at 3/4 of AddPath result, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_Reset(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a path
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Verify path exists
	if tsp.TotalLength() == 0 {
		t.Error("Expected non-zero length before reset")
	}

	// Reset the path
	tsp.Reset()

	// Verify path is cleared
	if tsp.TotalLength() != 0 {
		t.Errorf("Expected zero length after reset, got %f", tsp.TotalLength())
	}

	// Verify we can create a new path
	tsp.MoveTo(50, 50)
	tsp.LineTo(150, 150)
	tsp.FinalizePath()

	expectedLength := math.Sqrt(2) * 100.0
	if math.Abs(tsp.TotalLength()-expectedLength) > 1e-10 {
		t.Errorf("Expected length %f after reset and new path, got %f", expectedLength, tsp.TotalLength())
	}
}

func TestTransSinglePath_BaseLength_Scaling(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a path with length 100
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Set base length to 200 (2x scaling)
	tsp.SetBaseLength(200.0)

	// Test that coordinates are scaled appropriately
	x, y := 100.0, 0.0 // Half of base length
	tsp.Transform(&x, &y)

	// Should map to middle of actual path
	expectedX := 50.0
	expectedY := 0.0
	if math.Abs(x-expectedX) > 1e-10 || math.Abs(y-expectedY) > 1e-10 {
		t.Errorf("Expected (%f,%f) with base length scaling, got (%f,%f)", expectedX, expectedY, x, y)
	}
}

func TestTransSinglePath_EmptyPath(t *testing.T) {
	tsp := NewTransSinglePath()

	// Test transformation on empty path (should not crash)
	x, y := 50.0, 10.0
	tsp.Transform(&x, &y)

	// Coordinates should remain unchanged
	if x != 50.0 || y != 10.0 {
		t.Errorf("Expected coordinates unchanged for empty path, got (%f,%f)", x, y)
	}
}

func TestTransSinglePath_SinglePoint(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a path with only one point
	tsp.MoveTo(50, 50)
	tsp.FinalizePath()

	// Should have zero length
	if tsp.TotalLength() != 0 {
		t.Errorf("Expected zero length for single point, got %f", tsp.TotalLength())
	}

	// Test transformation (should not crash)
	x, y := 10.0, 5.0
	tsp.Transform(&x, &y)

	// Behavior may vary, but should not crash
}

func TestTransSinglePath_GetPositionAt(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test position at start
	x, y := tsp.GetPositionAt(0.0)
	if math.Abs(x-0.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (0,0) at start, got (%f,%f)", x, y)
	}

	// Test position at middle
	x, y = tsp.GetPositionAt(50.0)
	if math.Abs(x-50.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (50,0) at middle, got (%f,%f)", x, y)
	}

	// Test position at end
	x, y = tsp.GetPositionAt(100.0)
	if math.Abs(x-100.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (100,0) at end, got (%f,%f)", x, y)
	}
}

func TestTransSinglePath_GetTangentAt(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test tangent at middle - should be (1,0) for horizontal line
	dx, dy := tsp.GetTangentAt(50.0)
	expectedDx, expectedDy := 1.0, 0.0
	if math.Abs(dx-expectedDx) > 1e-10 || math.Abs(dy-expectedDy) > 1e-10 {
		t.Errorf("Expected tangent (%f,%f) for horizontal line, got (%f,%f)", expectedDx, expectedDy, dx, dy)
	}

	// Test that tangent is normalized
	length := math.Sqrt(dx*dx + dy*dy)
	if math.Abs(length-1.0) > 1e-10 {
		t.Errorf("Expected normalized tangent (length=1), got length=%f", length)
	}
}

func TestTransSinglePath_GetTangentAt_VerticalLine(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a vertical line from (0,0) to (0,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(0, 100)
	tsp.FinalizePath()

	// Test tangent at middle - should be (0,1) for vertical line going up
	dx, dy := tsp.GetTangentAt(50.0)
	expectedDx, expectedDy := 0.0, 1.0
	if math.Abs(dx-expectedDx) > 1e-10 || math.Abs(dy-expectedDy) > 1e-10 {
		t.Errorf("Expected tangent (%f,%f) for vertical line, got (%f,%f)", expectedDx, expectedDy, dx, dy)
	}
}

func TestTransSinglePath_GetNormalAt(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test normal at middle - should be (0,1) for horizontal line (90° CCW from tangent)
	nx, ny := tsp.GetNormalAt(50.0)
	expectedNx, expectedNy := 0.0, 1.0
	if math.Abs(nx-expectedNx) > 1e-10 || math.Abs(ny-expectedNy) > 1e-10 {
		t.Errorf("Expected normal (%f,%f) for horizontal line, got (%f,%f)", expectedNx, expectedNy, nx, ny)
	}

	// Test that normal is normalized
	length := math.Sqrt(nx*nx + ny*ny)
	if math.Abs(length-1.0) > 1e-10 {
		t.Errorf("Expected normalized normal (length=1), got length=%f", length)
	}
}

func TestTransSinglePath_GetAngleAt(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a horizontal line from (0,0) to (100,0)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 0)
	tsp.FinalizePath()

	// Test angle at middle - should be 0 for horizontal line pointing right
	angle := tsp.GetAngleAt(50.0)
	expectedAngle := 0.0
	if math.Abs(angle-expectedAngle) > 1e-10 {
		t.Errorf("Expected angle %f for horizontal line, got %f", expectedAngle, angle)
	}
}

func TestTransSinglePath_GetAngleAt_VerticalLine(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a vertical line from (0,0) to (0,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(0, 100)
	tsp.FinalizePath()

	// Test angle at middle - should be π/2 for vertical line pointing up
	angle := tsp.GetAngleAt(50.0)
	expectedAngle := math.Pi / 2
	if math.Abs(angle-expectedAngle) > 1e-10 {
		t.Errorf("Expected angle %f for vertical line, got %f", expectedAngle, angle)
	}
}

func TestTransSinglePath_GetAngleAt_DiagonalLine(t *testing.T) {
	tsp := NewTransSinglePath()

	// Create a 45-degree diagonal line from (0,0) to (100,100)
	tsp.MoveTo(0, 0)
	tsp.LineTo(100, 100)
	tsp.FinalizePath()

	// Test angle at middle - should be π/4 for 45-degree line
	angle := tsp.GetAngleAt(tsp.TotalLength() / 2)
	expectedAngle := math.Pi / 4
	if math.Abs(angle-expectedAngle) > 1e-10 {
		t.Errorf("Expected angle %f for diagonal line, got %f", expectedAngle, angle)
	}
}

func TestTransSinglePath_HelperMethods_EmptyPath(t *testing.T) {
	tsp := NewTransSinglePath()

	// Test helper methods on empty path (should not crash)
	x, y := tsp.GetPositionAt(10.0)
	if x != 10.0 || y != 0.0 {
		t.Errorf("Expected (10,0) for empty path, got (%f,%f)", x, y)
	}

	dx, dy := tsp.GetTangentAt(10.0)
	if dx != 0.0 || dy != 0.0 {
		t.Errorf("Expected (0,0) tangent for empty path, got (%f,%f)", dx, dy)
	}

	nx, ny := tsp.GetNormalAt(10.0)
	if nx != 0.0 || ny != 0.0 {
		t.Errorf("Expected (0,0) normal for empty path, got (%f,%f)", nx, ny)
	}

	angle := tsp.GetAngleAt(10.0)
	if angle != 0.0 {
		t.Errorf("Expected 0 angle for empty path, got %f", angle)
	}
}
