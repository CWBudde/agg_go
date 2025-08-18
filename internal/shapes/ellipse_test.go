package shapes

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

const ellipseEpsilon = 1e-10

func TestNewEllipse(t *testing.T) {
	ellipse := NewEllipse()

	if ellipse == nil {
		t.Fatal("NewEllipse() returned nil")
	}

	if ellipse.x != 0.0 || ellipse.y != 0.0 {
		t.Errorf("Expected default center (0, 0), got (%.1f, %.1f)", ellipse.x, ellipse.y)
	}

	if ellipse.rx != 1.0 || ellipse.ry != 1.0 {
		t.Errorf("Expected default radii (1, 1), got (%.1f, %.1f)", ellipse.rx, ellipse.ry)
	}

	if ellipse.scale != 1.0 {
		t.Errorf("Expected default scale 1.0, got %f", ellipse.scale)
	}

	if ellipse.cw {
		t.Error("Expected default counter-clockwise orientation")
	}

	if ellipse.step != 0 {
		t.Errorf("Expected initial step 0, got %d", ellipse.step)
	}
}

func TestNewEllipseWithParams(t *testing.T) {
	x, y := 100.0, 50.0
	rx, ry := 30.0, 20.0
	numSteps := uint32(8)
	cw := true

	ellipse := NewEllipseWithParams(x, y, rx, ry, numSteps, cw)

	if ellipse == nil {
		t.Fatal("NewEllipseWithParams() returned nil")
	}

	if ellipse.x != x || ellipse.y != y {
		t.Errorf("Expected center (%.1f, %.1f), got (%.1f, %.1f)", x, y, ellipse.x, ellipse.y)
	}

	if ellipse.rx != rx || ellipse.ry != ry {
		t.Errorf("Expected radii (%.1f, %.1f), got (%.1f, %.1f)", rx, ry, ellipse.rx, ellipse.ry)
	}

	if ellipse.num != numSteps {
		t.Errorf("Expected num steps %d, got %d", numSteps, ellipse.num)
	}

	if ellipse.cw != cw {
		t.Errorf("Expected cw %v, got %v", cw, ellipse.cw)
	}
}

func TestNewEllipseWithParamsAutoSteps(t *testing.T) {
	// Test auto-calculation of steps when numSteps = 0
	ellipse := NewEllipseWithParams(0, 0, 10, 10, 0, false)

	if ellipse.num == 0 {
		t.Error("Expected auto-calculated steps to be > 0")
	}

	// Should be reasonable number of steps for a circle of radius 10
	if ellipse.num < 4 || ellipse.num > 1000 {
		t.Errorf("Auto-calculated steps seem unreasonable: %d", ellipse.num)
	}
}

func TestEllipseInit(t *testing.T) {
	ellipse := NewEllipse()
	x, y := 50.0, 25.0
	rx, ry := 15.0, 10.0
	numSteps := uint32(12)
	cw := true

	ellipse.Init(x, y, rx, ry, numSteps, cw)

	if ellipse.x != x || ellipse.y != y {
		t.Errorf("Expected center (%.1f, %.1f), got (%.1f, %.1f)", x, y, ellipse.x, ellipse.y)
	}

	if ellipse.rx != rx || ellipse.ry != ry {
		t.Errorf("Expected radii (%.1f, %.1f), got (%.1f, %.1f)", rx, ry, ellipse.rx, ellipse.ry)
	}

	if ellipse.num != numSteps {
		t.Errorf("Expected num steps %d, got %d", numSteps, ellipse.num)
	}

	if ellipse.cw != cw {
		t.Errorf("Expected cw %v, got %v", cw, ellipse.cw)
	}

	if ellipse.step != 0 {
		t.Errorf("Expected step reset to 0, got %d", ellipse.step)
	}
}

func TestEllipseApproximationScale(t *testing.T) {
	ellipse := NewEllipse()

	// Test getter
	if ellipse.ApproximationScale() != 1.0 {
		t.Errorf("Expected default approximation scale 1.0, got %f", ellipse.ApproximationScale())
	}

	// Test setter
	newScale := 2.5
	ellipse.SetApproximationScale(newScale)
	if ellipse.ApproximationScale() != newScale {
		t.Errorf("Expected approximation scale %f, got %f", newScale, ellipse.ApproximationScale())
	}
}

func TestEllipseApproximationScaleEffect(t *testing.T) {
	// Test that higher approximation scale generates more steps
	ellipse1 := NewEllipseWithParams(0, 0, 10, 10, 0, false)
	ellipse1.SetApproximationScale(1.0)

	ellipse2 := NewEllipseWithParams(0, 0, 10, 10, 0, false)
	ellipse2.SetApproximationScale(4.0)

	if ellipse2.num <= ellipse1.num {
		t.Errorf("Higher approximation scale should generate more steps: scale 1.0=%d, scale 4.0=%d",
			ellipse1.num, ellipse2.num)
	}

	t.Logf("Scale 1.0: %d steps, Scale 4.0: %d steps", ellipse1.num, ellipse2.num)
}

func TestEllipseRewind(t *testing.T) {
	ellipse := NewEllipseWithParams(0, 0, 10, 10, 8, false)
	ellipse.step = 5 // Advance step counter

	ellipse.Rewind(0)

	if ellipse.step != 0 {
		t.Errorf("Expected step to be reset to 0, got %d", ellipse.step)
	}
}

func TestVertexBasicCircle(t *testing.T) {
	// Create a simple circle with 8 steps
	ellipse := NewEllipseWithParams(0, 0, 10, 10, 8, false)
	ellipse.Rewind(0)

	var x, y float64
	vertexCount := 0

	// First vertex should be MoveTo at (10, 0)
	cmd := ellipse.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be PathCmdMoveTo, got %d", cmd)
	}

	expectedX, expectedY := 10.0, 0.0
	if math.Abs(x-expectedX) > ellipseEpsilon || math.Abs(y-expectedY) > ellipseEpsilon {
		t.Errorf("Expected first vertex (%.6f, %.6f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}
	vertexCount++

	// Generate LineTo vertices
	for vertexCount < int(ellipse.num) {
		cmd = ellipse.Vertex(&x, &y)
		if cmd != basics.PathCmdLineTo {
			t.Errorf("Expected PathCmdLineTo at step %d, got %d", vertexCount, cmd)
		}

		// Verify vertex is on the circle
		dist := math.Sqrt(x*x + y*y)
		if math.Abs(dist-10.0) > ellipseEpsilon {
			t.Errorf("Vertex (%.6f, %.6f) not on circle, distance=%f", x, y, dist)
		}

		vertexCount++
	}

	// Next vertex should be EndPoly with close flags
	cmd = ellipse.Vertex(&x, &y)
	if (cmd & basics.PathCommand(basics.PathCmdMask)) != basics.PathCmdEndPoly {
		t.Errorf("Expected PathCmdEndPoly, got %d", cmd)
	}

	// Check that close flag is set
	if (uint32(cmd) & uint32(basics.PathFlagsClose)) == 0 {
		t.Error("Expected PathFlagsClose to be set")
	}

	// Final call should return Stop
	cmd = ellipse.Vertex(&x, &y)
	if cmd != basics.PathCmdStop {
		t.Errorf("Expected PathCmdStop, got %d", cmd)
	}

	t.Logf("Generated %d vertices for circle", vertexCount)
}

func TestVertexClockwiseCircle(t *testing.T) {
	// Create a clockwise circle with 8 steps
	ellipse := NewEllipseWithParams(0, 0, 10, 10, 8, true)
	ellipse.Rewind(0)

	var x, y float64
	vertices := make([][2]float64, 0, 8)

	// Collect all vertices
	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) {
			break
		}
		if basics.IsStop(cmd) {
			break
		}
		vertices = append(vertices, [2]float64{x, y})
	}

	// For clockwise, angles should decrease
	// Check that we move clockwise by comparing consecutive angles
	for i := 1; i < len(vertices); i++ {
		angle1 := math.Atan2(vertices[i-1][1], vertices[i-1][0])
		angle2 := math.Atan2(vertices[i][1], vertices[i][0])

		// Normalize angles to [0, 2π]
		if angle1 < 0 {
			angle1 += 2 * math.Pi
		}
		if angle2 < 0 {
			angle2 += 2 * math.Pi
		}

		// For clockwise, angle should decrease (with wraparound)
		diff := angle1 - angle2
		if diff < 0 {
			diff += 2 * math.Pi
		}

		// Should be positive and reasonable (not too big due to wraparound)
		if diff <= 0 || diff > math.Pi {
			t.Errorf("Angles not decreasing clockwise at step %d: %.6f -> %.6f (diff=%.6f)",
				i, angle1, angle2, diff)
		}
	}
}

func TestVertexEllipse(t *testing.T) {
	// Create an ellipse with different X and Y radii
	ellipse := NewEllipseWithParams(5, 3, 20, 10, 0, false)
	ellipse.Rewind(0)

	var x, y float64
	vertexCount := 0

	// Generate all vertices
	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
			break
		}

		// Verify vertex is on the ellipse: ((x-cx)/rx)² + ((y-cy)/ry)² = 1
		dx := (x - 5.0) / 20.0
		dy := (y - 3.0) / 10.0
		ellipseValue := dx*dx + dy*dy

		if math.Abs(ellipseValue-1.0) > ellipseEpsilon {
			t.Errorf("Vertex (%.6f, %.6f) not on ellipse, ellipse value=%f", x, y, ellipseValue)
		}

		vertexCount++
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	t.Logf("Generated %d vertices for ellipse", vertexCount)
}

func TestVertexZeroRadius(t *testing.T) {
	// Test ellipse with zero X radius
	ellipse := NewEllipseWithParams(5, 5, 0, 10, 8, false)
	ellipse.Rewind(0)

	var x, y float64
	vertexCount := 0

	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
			break
		}

		// All X coordinates should be 5 (center X) since rx=0
		if math.Abs(x-5.0) > ellipseEpsilon {
			t.Errorf("Expected x=5.0 for zero X radius, got %f", x)
		}

		vertexCount++
		if vertexCount > 100 {
			t.Fatal("Too many vertices generated")
		}
	}
}

func TestVertexSinglePoint(t *testing.T) {
	// Test ellipse with both radii zero (degenerate case)
	ellipse := NewEllipseWithParams(7, 3, 0, 0, 4, false)
	ellipse.Rewind(0)

	var x, y float64
	vertexCount := 0

	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
			break
		}

		// All coordinates should be at center (7, 3)
		if math.Abs(x-7.0) > ellipseEpsilon || math.Abs(y-3.0) > ellipseEpsilon {
			t.Errorf("Expected vertex (7.0, 3.0) for zero radii, got (%.6f, %.6f)", x, y)
		}

		vertexCount++
		if vertexCount > 100 {
			t.Fatal("Too many vertices generated")
		}
	}
}

func TestVertexLargeEllipse(t *testing.T) {
	// Test with large radii to ensure no overflow issues
	ellipse := NewEllipseWithParams(0, 0, 1000, 500, 0, false)
	ellipse.Rewind(0)

	var x, y float64
	vertexCount := 0
	maxDist := 0.0

	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
			break
		}

		// Track maximum distance from center
		dist := math.Sqrt(x*x + y*y)
		if dist > maxDist {
			maxDist = dist
		}

		// Verify vertex is on the ellipse
		dx := x / 1000.0
		dy := y / 500.0
		ellipseValue := dx*dx + dy*dy

		if math.Abs(ellipseValue-1.0) > 1e-6 {
			t.Errorf("Vertex (%.6f, %.6f) not on large ellipse, ellipse value=%f", x, y, ellipseValue)
		}

		vertexCount++
		if vertexCount > 10000 {
			t.Fatal("Too many vertices generated")
		}
	}

	// Maximum distance should be close to major radius
	expectedMaxDist := 1000.0
	if math.Abs(maxDist-expectedMaxDist) > 1.0 {
		t.Errorf("Expected max distance ~%.1f, got %.6f", expectedMaxDist, maxDist)
	}

	t.Logf("Generated %d vertices for large ellipse, max distance=%.6f", vertexCount, maxDist)
}

func TestCalcNumSteps(t *testing.T) {
	ellipse := NewEllipse()
	ellipse.rx = 10.0
	ellipse.ry = 5.0
	ellipse.scale = 1.0

	ellipse.calcNumSteps()

	// Should generate reasonable number of steps
	if ellipse.num < 4 {
		t.Errorf("Too few steps calculated: %d", ellipse.num)
	}

	if ellipse.num > 1000 {
		t.Errorf("Too many steps calculated: %d", ellipse.num)
	}

	// Higher scale should give more steps
	oldNum := ellipse.num
	ellipse.scale = 4.0
	ellipse.calcNumSteps()

	if ellipse.num <= oldNum {
		t.Errorf("Higher scale didn't increase steps: %d -> %d", oldNum, ellipse.num)
	}
}

func TestVertexPathCommands(t *testing.T) {
	// Test the exact sequence of path commands
	ellipse := NewEllipseWithParams(0, 0, 5, 5, 4, false) // Simple 4-step circle
	ellipse.Rewind(0)

	var x, y float64
	commands := make([]basics.PathCommand, 0, 6)

	// Collect all commands
	for {
		cmd := ellipse.Vertex(&x, &y)
		commands = append(commands, cmd)
		if basics.IsStop(cmd) {
			break
		}
	}

	// Should have: MoveTo, LineTo, LineTo, LineTo, EndPoly, Stop
	expectedCmds := []basics.PathCommand{
		basics.PathCmdMoveTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCmdLineTo,
		basics.PathCommand(uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose) | uint32(basics.PathFlagsCCW)),
		basics.PathCmdStop,
	}

	if len(commands) != len(expectedCmds) {
		t.Errorf("Expected %d commands, got %d", len(expectedCmds), len(commands))
	}

	for i, expected := range expectedCmds {
		if i >= len(commands) {
			break
		}
		if commands[i] != expected {
			t.Errorf("Command %d: expected %d, got %d", i, expected, commands[i])
		}
	}
}

func TestVertexRepeatedRewind(t *testing.T) {
	// Test that Rewind can be called multiple times
	ellipse := NewEllipseWithParams(0, 0, 5, 5, 4, false)

	var x, y float64

	// First generation
	ellipse.Rewind(0)
	cmd1 := ellipse.Vertex(&x, &y)
	x1, y1 := x, y

	// Partial generation
	ellipse.Vertex(&x, &y)

	// Rewind and generate again
	ellipse.Rewind(0)
	cmd2 := ellipse.Vertex(&x, &y)
	x2, y2 := x, y

	if cmd1 != cmd2 {
		t.Errorf("First command after rewind differs: %d vs %d", cmd1, cmd2)
	}

	if math.Abs(x1-x2) > ellipseEpsilon || math.Abs(y1-y2) > ellipseEpsilon {
		t.Errorf("First vertex after rewind differs: (%.6f,%.6f) vs (%.6f,%.6f)", x1, y1, x2, y2)
	}
}

// Helper function to count vertices generated by an ellipse
func countEllipseVertices(ellipse *Ellipse) int {
	var x, y float64
	count := 0

	ellipse.Rewind(0)
	for {
		cmd := ellipse.Vertex(&x, &y)
		if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
			break
		}
		count++
		if count > 10000 {
			break // Safety check
		}
	}

	return count
}

func BenchmarkEllipseVertex(b *testing.B) {
	ellipse := NewEllipseWithParams(0, 0, 100, 100, 0, false)
	var x, y float64

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ellipse.Rewind(0)
		for {
			cmd := ellipse.Vertex(&x, &y)
			if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
				break
			}
		}
	}
}

func BenchmarkEllipseVertexHighQuality(b *testing.B) {
	ellipse := NewEllipseWithParams(0, 0, 100, 100, 0, false)
	ellipse.SetApproximationScale(4.0) // High quality
	var x, y float64

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ellipse.Rewind(0)
		for {
			cmd := ellipse.Vertex(&x, &y)
			if basics.IsEndPoly(cmd) || basics.IsStop(cmd) {
				break
			}
		}
	}
}

func BenchmarkCalcNumSteps(b *testing.B) {
	ellipse := NewEllipse()
	ellipse.rx = 100.0
	ellipse.ry = 50.0

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ellipse.scale = float64(i%10) + 1.0
		ellipse.calcNumSteps()
	}
}
