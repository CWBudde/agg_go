package shapes

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

const roundedRectEpsilon = 1e-10

func TestNewRoundedRect(t *testing.T) {
	rr := NewRoundedRect(10, 20, 100, 80, 5)

	// Check bounds
	if rr.x1 != 10 || rr.y1 != 20 || rr.x2 != 100 || rr.y2 != 80 {
		t.Errorf("Expected bounds (10,20,100,80), got (%g,%g,%g,%g)",
			rr.x1, rr.y1, rr.x2, rr.y2)
	}

	// Check that all radii are set to the same value
	if rr.rx1 != 5 || rr.ry1 != 5 || rr.rx2 != 5 || rr.ry2 != 5 ||
		rr.rx3 != 5 || rr.ry3 != 5 || rr.rx4 != 5 || rr.ry4 != 5 {
		t.Errorf("Expected all radii to be 5")
	}
}

func TestSetRect(t *testing.T) {
	rr := NewRoundedRectEmpty()

	// Test normal coordinates
	rr.SetRect(10, 20, 100, 80)
	if rr.x1 != 10 || rr.y1 != 20 || rr.x2 != 100 || rr.y2 != 80 {
		t.Errorf("Expected bounds (10,20,100,80), got (%g,%g,%g,%g)",
			rr.x1, rr.y1, rr.x2, rr.y2)
	}

	// Test coordinate normalization (swapped coordinates)
	rr.SetRect(100, 80, 10, 20)
	if rr.x1 != 10 || rr.y1 != 20 || rr.x2 != 100 || rr.y2 != 80 {
		t.Errorf("Expected normalized bounds (10,20,100,80), got (%g,%g,%g,%g)",
			rr.x1, rr.y1, rr.x2, rr.y2)
	}
}

func TestSetRadius(t *testing.T) {
	rr := NewRoundedRectEmpty()

	// Test single radius
	rr.SetRadius(15)
	expected := 15.0
	radii := []float64{rr.rx1, rr.ry1, rr.rx2, rr.ry2, rr.rx3, rr.ry3, rr.rx4, rr.ry4}
	for i, r := range radii {
		if r != expected {
			t.Errorf("Radius %d: expected %g, got %g", i, expected, r)
		}
	}
}

func TestSetRadiusXY(t *testing.T) {
	rr := NewRoundedRectEmpty()

	rr.SetRadiusXY(10, 20)
	expectedRX := 10.0
	expectedRY := 20.0

	rxValues := []float64{rr.rx1, rr.rx2, rr.rx3, rr.rx4}
	ryValues := []float64{rr.ry1, rr.ry2, rr.ry3, rr.ry4}

	for i, rx := range rxValues {
		if rx != expectedRX {
			t.Errorf("RX %d: expected %g, got %g", i, expectedRX, rx)
		}
	}
	for i, ry := range ryValues {
		if ry != expectedRY {
			t.Errorf("RY %d: expected %g, got %g", i, expectedRY, ry)
		}
	}
}

func TestSetRadiusBottomTop(t *testing.T) {
	rr := NewRoundedRectEmpty()

	rr.SetRadiusBottomTop(5, 10, 15, 20)

	// Check bottom corners (1,2)
	if rr.rx1 != 5 || rr.ry1 != 10 || rr.rx2 != 5 || rr.ry2 != 10 {
		t.Errorf("Bottom radii: expected (5,10), got (%g,%g) and (%g,%g)",
			rr.rx1, rr.ry1, rr.rx2, rr.ry2)
	}

	// Check top corners (3,4)
	if rr.rx3 != 15 || rr.ry3 != 20 || rr.rx4 != 15 || rr.ry4 != 20 {
		t.Errorf("Top radii: expected (15,20), got (%g,%g) and (%g,%g)",
			rr.rx3, rr.ry3, rr.rx4, rr.ry4)
	}
}

func TestSetRadiusAll(t *testing.T) {
	rr := NewRoundedRectEmpty()

	rr.SetRadiusAll(1, 2, 3, 4, 5, 6, 7, 8)

	expected := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	actual := []float64{rr.rx1, rr.ry1, rr.rx2, rr.ry2, rr.rx3, rr.ry3, rr.rx4, rr.ry4}

	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("Radius %d: expected %g, got %g", i, exp, actual[i])
		}
	}
}

func TestNormalizeRadius(t *testing.T) {
	rr := NewRoundedRect(0, 0, 100, 60, 0) // 100x60 rectangle

	// Set radii that would overlap
	rr.SetRadiusAll(50, 30, 60, 40, 40, 20, 30, 35) // rx1+rx2=110 > 100 (width)

	rr.NormalizeRadius()

	// Check that no two opposing radii sum to more than the corresponding dimension
	width := rr.x2 - rr.x1  // 100
	height := rr.y2 - rr.y1 // 60

	if rr.rx1+rr.rx2 > width+roundedRectEpsilon {
		t.Errorf("Bottom radii sum (%g) exceeds width (%g)", rr.rx1+rr.rx2, width)
	}
	if rr.rx3+rr.rx4 > width+roundedRectEpsilon {
		t.Errorf("Top radii sum (%g) exceeds width (%g)", rr.rx3+rr.rx4, width)
	}
	if rr.ry1+rr.ry2 > height+roundedRectEpsilon {
		t.Errorf("Left radii sum (%g) exceeds height (%g)", rr.ry1+rr.ry2, height)
	}
	if rr.ry3+rr.ry4 > height+roundedRectEpsilon {
		t.Errorf("Right radii sum (%g) exceeds height (%g)", rr.ry3+rr.ry4, height)
	}
}

func TestRoundedRectApproximationScale(t *testing.T) {
	rr := NewRoundedRect(0, 0, 100, 100, 10)

	// Test default scale
	if rr.ApproximationScale() != 1.0 {
		t.Errorf("Expected default scale 1.0, got %g", rr.ApproximationScale())
	}

	// Test setting scale
	rr.SetApproximationScale(2.5)
	if math.Abs(rr.ApproximationScale()-2.5) > roundedRectEpsilon {
		t.Errorf("Expected scale 2.5, got %g", rr.ApproximationScale())
	}
}

func TestRoundedRectRewind(t *testing.T) {
	rr := NewRoundedRect(0, 0, 100, 100, 10)

	// Advance state by calling Vertex once
	var x, y float64
	rr.Vertex(&x, &y)

	// Rewind should reset status to 0
	rr.Rewind(0)
	if rr.status != 0 {
		t.Errorf("Expected status 0 after rewind, got %d", rr.status)
	}
}

func TestVertexGeneration(t *testing.T) {
	// Create a simple square with small radius for predictable testing
	rr := NewRoundedRect(0, 0, 100, 100, 10)
	rr.NormalizeRadius()

	var x, y float64
	var vertices []struct {
		x, y float64
		cmd  basics.PathCommand
	}

	// Collect all vertices
	rr.Rewind(0)
	for {
		cmd := rr.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, struct {
			x, y float64
			cmd  basics.PathCommand
		}{x, y, cmd})

		// Safety check to prevent infinite loops
		if len(vertices) > 1000 {
			t.Fatal("Too many vertices generated, possible infinite loop")
		}
	}

	if len(vertices) == 0 {
		t.Fatal("No vertices generated")
	}

	// First vertex should be a MoveTo command
	if !basics.IsVertex(vertices[0].cmd) {
		t.Errorf("First vertex should have a vertex command, got %d", vertices[0].cmd)
	}

	// Check that we have some LineTo commands
	hasLineTo := false
	for _, v := range vertices {
		if basics.IsVertex(v.cmd) && v.cmd != basics.PathCmdMoveTo {
			hasLineTo = true
			break
		}
	}
	if !hasLineTo {
		t.Error("Should have LineTo commands in the path")
	}

	// Last command should be end polygon
	rr.Rewind(0)
	var lastCmd basics.PathCommand
	for {
		cmd := rr.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		lastCmd = cmd
	}

	expectedEndCmd := basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW))
	if lastCmd != expectedEndCmd {
		t.Errorf("Expected end command %d, got %d", expectedEndCmd, lastCmd)
	}
}

func TestVertexCoordinateRanges(t *testing.T) {
	// Test that generated vertices are within reasonable bounds
	rr := NewRoundedRect(10, 20, 90, 80, 5)

	var x, y float64
	minX, maxX := math.Inf(1), math.Inf(-1)
	minY, maxY := math.Inf(1), math.Inf(-1)

	rr.Rewind(0)
	for {
		cmd := rr.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsVertex(cmd) {
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	// Check that vertices are within the bounding box (allowing for radius)
	tolerance := 0.1 // Small tolerance for floating point precision
	if minX < rr.x1-tolerance || maxX > rr.x2+tolerance {
		t.Errorf("X coordinates out of bounds: got [%g, %g], expected [%g, %g]",
			minX, maxX, rr.x1, rr.x2)
	}
	if minY < rr.y1-tolerance || maxY > rr.y2+tolerance {
		t.Errorf("Y coordinates out of bounds: got [%g, %g], expected [%g, %g]",
			minY, maxY, rr.y1, rr.y2)
	}
}

func TestZeroRadius(t *testing.T) {
	// Test behavior with zero radius (should be a regular rectangle)
	rr := NewRoundedRect(0, 0, 100, 100, 0)

	var x, y float64
	var vertices []struct{ x, y float64 }

	rr.Rewind(0)
	for {
		cmd := rr.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		if basics.IsVertex(cmd) {
			vertices = append(vertices, struct{ x, y float64 }{x, y})
		}
	}

	// With zero radius, we should still get vertices (even if arcs are degenerate)
	if len(vertices) == 0 {
		t.Error("Zero radius should still generate vertices")
	}
}

func TestNegativeRadius(t *testing.T) {
	// Test behavior with negative radius
	rr := NewRoundedRect(0, 0, 100, 100, -10)
	rr.NormalizeRadius()

	// After normalization, radii should not be negative
	radii := []float64{rr.rx1, rr.ry1, rr.rx2, rr.ry2, rr.rx3, rr.ry3, rr.rx4, rr.ry4}
	for i, r := range radii {
		if r < 0 {
			t.Errorf("Radius %d should not be negative after normalization: %g", i, r)
		}
	}
}
