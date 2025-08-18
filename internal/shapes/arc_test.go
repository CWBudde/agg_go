package shapes

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

const epsilon = 1e-10

func TestNewArc(t *testing.T) {
	arc := NewArc()

	if arc == nil {
		t.Fatal("NewArc() returned nil")
	}

	if arc.scale != 1.0 {
		t.Errorf("Expected default scale 1.0, got %f", arc.scale)
	}

	if arc.initialized {
		t.Error("Expected new arc to be uninitialized")
	}
}

func TestNewArcWithParams(t *testing.T) {
	x, y := 100.0, 50.0
	rx, ry := 30.0, 20.0
	a1, a2 := 0.0, math.Pi/2
	ccw := true

	arc := NewArcWithParams(x, y, rx, ry, a1, a2, ccw)

	if arc == nil {
		t.Fatal("NewArcWithParams() returned nil")
	}

	if arc.x != x || arc.y != y {
		t.Errorf("Expected center (%.1f, %.1f), got (%.1f, %.1f)", x, y, arc.x, arc.y)
	}

	if arc.rx != rx || arc.ry != ry {
		t.Errorf("Expected radii (%.1f, %.1f), got (%.1f, %.1f)", rx, ry, arc.rx, arc.ry)
	}

	if !arc.initialized {
		t.Error("Expected arc to be initialized")
	}

	if arc.ccw != ccw {
		t.Errorf("Expected ccw %v, got %v", ccw, arc.ccw)
	}
}

func TestArcInit(t *testing.T) {
	arc := NewArc()
	x, y := 100.0, 50.0
	rx, ry := 30.0, 20.0
	a1, a2 := 0.0, math.Pi
	ccw := false

	arc.Init(x, y, rx, ry, a1, a2, ccw)

	if arc.x != x || arc.y != y {
		t.Errorf("Expected center (%.1f, %.1f), got (%.1f, %.1f)", x, y, arc.x, arc.y)
	}

	if arc.rx != rx || arc.ry != ry {
		t.Errorf("Expected radii (%.1f, %.1f), got (%.1f, %.1f)", rx, ry, arc.rx, arc.ry)
	}

	if !arc.initialized {
		t.Error("Expected arc to be initialized after Init()")
	}

	if arc.ccw != ccw {
		t.Errorf("Expected ccw %v, got %v", ccw, arc.ccw)
	}
}

func TestApproximationScale(t *testing.T) {
	arc := NewArc()

	// Test getter
	if arc.ApproximationScale() != 1.0 {
		t.Errorf("Expected default approximation scale 1.0, got %f", arc.ApproximationScale())
	}

	// Test setter
	newScale := 2.5
	arc.SetApproximationScale(newScale)
	if arc.ApproximationScale() != newScale {
		t.Errorf("Expected approximation scale %f, got %f", newScale, arc.ApproximationScale())
	}
}

func TestRewind(t *testing.T) {
	arc := NewArcWithParams(0, 0, 10, 10, 0, math.Pi/2, true)

	arc.Rewind(0)

	if arc.pathCmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo after rewind, got %d", arc.pathCmd)
	}

	if arc.angle != arc.start {
		t.Errorf("Expected angle to be reset to start (%f), got %f", arc.start, arc.angle)
	}
}

func TestVertexBasicArc(t *testing.T) {
	// Create a quarter circle (90 degrees) from 0 to π/2
	arc := NewArcWithParams(0, 0, 10, 10, 0, math.Pi/2, true)
	arc.Rewind(0)

	var x, y float64
	vertexCount := 0

	// First vertex should be MoveTo at (10, 0)
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be PathCmdMoveTo, got %d", cmd)
	}

	expectedX, expectedY := 10.0, 0.0
	if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
		t.Errorf("Expected first vertex (%.6f, %.6f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}
	vertexCount++

	// Generate remaining vertices until stop
	for {
		cmd = arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}

		if cmd != basics.PathCmdLineTo {
			t.Errorf("Expected PathCmdLineTo, got %d", cmd)
		}

		// Verify vertex is approximately on the circle
		dist := math.Sqrt(x*x + y*y)
		if math.Abs(dist-10.0) > 0.1 { // Allow some tolerance for tessellation
			t.Errorf("Vertex (%.6f, %.6f) not on circle, distance=%f", x, y, dist)
		}

		vertexCount++

		// Safety check to prevent infinite loop
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	// Final vertex should be approximately at (0, 10)
	expectedX, expectedY = 0.0, 10.0
	if math.Abs(x-expectedX) > 0.01 || math.Abs(y-expectedY) > 0.01 {
		t.Errorf("Expected final vertex (%.1f, %.1f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}

	t.Logf("Generated %d vertices for quarter circle", vertexCount)
}

func TestVertexClockwiseArc(t *testing.T) {
	// Create a clockwise quarter circle from π/2 to 0
	arc := NewArcWithParams(0, 0, 10, 10, math.Pi/2, 0, false)
	arc.Rewind(0)

	var x, y float64

	// First vertex should be MoveTo at (0, 10)
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected first command to be PathCmdMoveTo, got %d", cmd)
	}

	expectedX, expectedY := 0.0, 10.0
	if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
		t.Errorf("Expected first vertex (%.1f, %.1f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}

	// Skip to final vertex
	vertexCount := 1
	for {
		cmd = arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	// Final vertex should be approximately at (10, 0)
	expectedX, expectedY = 10.0, 0.0
	if math.Abs(x-expectedX) > 0.01 || math.Abs(y-expectedY) > 0.01 {
		t.Errorf("Expected final vertex (%.1f, %.1f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}
}

func TestVertexEllipticalArc(t *testing.T) {
	// Create an elliptical arc with different X and Y radii
	arc := NewArcWithParams(5, 3, 20, 10, 0, math.Pi, true)
	arc.Rewind(0)

	var x, y float64
	vertexCount := 0

	// Generate all vertices
	for {
		cmd := arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}

		// Verify vertex is on the ellipse: ((x-cx)/rx)² + ((y-cy)/ry)² = 1
		dx := (x - 5.0) / 20.0
		dy := (y - 3.0) / 10.0
		ellipseValue := dx*dx + dy*dy

		if math.Abs(ellipseValue-1.0) > 0.01 { // Allow tolerance for tessellation
			t.Errorf("Vertex (%.6f, %.6f) not on ellipse, ellipse value=%f", x, y, ellipseValue)
		}

		vertexCount++
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	t.Logf("Generated %d vertices for elliptical arc", vertexCount)
}

func TestVertexFullCircle(t *testing.T) {
	// Create a full circle (2π radians)
	arc := NewArcWithParams(0, 0, 5, 5, 0, 2*math.Pi, true)
	arc.Rewind(0)

	var x, y float64
	var firstX, firstY float64
	vertexCount := 0

	// Get first vertex
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdMoveTo {
		t.Errorf("Expected PathCmdMoveTo, got %d", cmd)
	}
	firstX, firstY = x, y
	vertexCount++

	// Generate remaining vertices
	for {
		cmd = arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		vertexCount++
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}

	// Final vertex should be close to first vertex (full circle)
	if math.Abs(x-firstX) > 0.01 || math.Abs(y-firstY) > 0.01 {
		t.Errorf("Full circle didn't close properly: first(%.6f, %.6f) final(%.6f, %.6f)",
			firstX, firstY, x, y)
	}

	t.Logf("Generated %d vertices for full circle", vertexCount)
}

func TestApproximationScaleEffect(t *testing.T) {
	// Test that higher approximation scale generates more vertices
	arc1 := NewArcWithParams(0, 0, 10, 10, 0, math.Pi/2, true)
	arc1.SetApproximationScale(1.0)
	arc1.Rewind(0)

	arc2 := NewArcWithParams(0, 0, 10, 10, 0, math.Pi/2, true)
	arc2.SetApproximationScale(4.0)
	arc2.Rewind(0)

	count1 := countVertices(arc1)
	count2 := countVertices(arc2)

	if count2 <= count1 {
		t.Errorf("Higher approximation scale should generate more vertices: scale 1.0=%d, scale 4.0=%d",
			count1, count2)
	}

	t.Logf("Scale 1.0: %d vertices, Scale 4.0: %d vertices", count1, count2)
}

func TestZeroRadiusArc(t *testing.T) {
	// Test arc with zero radius - should still work
	arc := NewArcWithParams(5, 5, 0, 10, 0, math.Pi, true)
	arc.Rewind(0)

	var x, y float64
	vertexCount := 0

	for {
		cmd := arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}

		// All X coordinates should be 5 (center X) since rx=0
		if math.Abs(x-5.0) > epsilon {
			t.Errorf("Expected x=5.0 for zero X radius, got %f", x)
		}

		vertexCount++
		if vertexCount > 1000 {
			t.Fatal("Too many vertices generated")
		}
	}
}

func TestSameStartEndAngles(t *testing.T) {
	// Test arc where start and end angles are the same
	arc := NewArcWithParams(0, 0, 10, 10, math.Pi/4, math.Pi/4, true)
	arc.Rewind(0)

	var x, y float64

	// With equal start and end, normalize doesn't add 2π
	if arc.start != arc.end {
		t.Errorf("Expected start==end for same angles, got start=%.6f, end=%.6f", arc.start, arc.end)
	}

	// For same start/end angles, the termination condition is met immediately
	// since angle < end-da/4 for reasonable da values, so we get LineTo directly
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for same start/end angles, got %d", cmd)
	}

	// Position should be at the angle (start==end)
	expectedX := math.Cos(arc.start) * arc.rx
	expectedY := math.Sin(arc.start) * arc.ry
	if math.Abs(x-expectedX) > epsilon || math.Abs(y-expectedY) > epsilon {
		t.Errorf("Expected position (%.6f, %.6f), got (%.6f, %.6f)", expectedX, expectedY, x, y)
	}

	// Next call should return Stop
	cmd = arc.Vertex(&x, &y)
	if !basics.IsStop(cmd) {
		t.Errorf("Expected PathCmdStop after same angle arc, got %d", cmd)
	}
}

func TestSameStartEndAnglesCW(t *testing.T) {
	// Test arc where start and end angles are the same with clockwise
	arc := NewArcWithParams(0, 0, 10, 10, math.Pi/4, math.Pi/4, false)
	arc.Rewind(0)

	var x, y float64

	// With equal start and end for clockwise, normalize doesn't change them
	if arc.start != arc.end {
		t.Errorf("Expected start==end for same angles CW, got start=%.6f, end=%.6f", arc.start, arc.end)
	}

	// For CW with same angles, should behave similarly - immediate termination
	cmd := arc.Vertex(&x, &y)
	if cmd != basics.PathCmdLineTo {
		t.Errorf("Expected PathCmdLineTo for same start/end angles CW, got %d", cmd)
	}

	// Next call should return Stop
	cmd = arc.Vertex(&x, &y)
	if !basics.IsStop(cmd) {
		t.Errorf("Expected PathCmdStop after same angle CW arc, got %d", cmd)
	}
}

// Helper function to count vertices generated by an arc
func countVertices(arc *Arc) int {
	var x, y float64
	count := 0

	for {
		cmd := arc.Vertex(&x, &y)
		if basics.IsStop(cmd) {
			break
		}
		count++
		if count > 1000 {
			break // Safety check
		}
	}

	return count
}

func TestNormalizeCCW(t *testing.T) {
	arc := NewArc()
	arc.rx = 10.0
	arc.ry = 10.0
	arc.scale = 1.0

	// Test counter-clockwise normalization
	arc.normalize(0, math.Pi/2, true)

	if !arc.ccw {
		t.Error("Expected CCW flag to be true")
	}

	if arc.start != 0 {
		t.Errorf("Expected start angle 0, got %f", arc.start)
	}

	if arc.end != math.Pi/2 {
		t.Errorf("Expected end angle π/2, got %f", arc.end)
	}

	if arc.da <= 0 {
		t.Errorf("Expected positive da for CCW, got %f", arc.da)
	}
}

func TestNormalizeCW(t *testing.T) {
	arc := NewArc()
	arc.rx = 10.0
	arc.ry = 10.0
	arc.scale = 1.0

	// Test clockwise normalization
	arc.normalize(math.Pi/2, 0, false)

	if arc.ccw {
		t.Error("Expected CCW flag to be false")
	}

	if arc.da >= 0 {
		t.Errorf("Expected negative da for CW, got %f", arc.da)
	}
}

func BenchmarkArcVertex(b *testing.B) {
	arc := NewArcWithParams(0, 0, 100, 100, 0, 2*math.Pi, true)
	var x, y float64

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		arc.Rewind(0)
		for {
			cmd := arc.Vertex(&x, &y)
			if basics.IsStop(cmd) {
				break
			}
		}
	}
}
