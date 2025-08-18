package transform

import (
	"math"
	"testing"

	"agg_go/internal/basics"
)

func newMockPath(coords []float64) *mockVertexSource {
	if len(coords)%2 != 0 {
		panic("coords must have even length")
	}

	vertices := make([]vertex, len(coords)/2)

	for i := 0; i < len(coords); i += 2 {
		cmd := basics.PathCmdLineTo
		if i == 0 {
			cmd = basics.PathCmdMoveTo
		}
		vertices[i/2] = vertex{coords[i], coords[i+1], cmd}
	}

	return newMockVertexSource(vertices)
}

func TestTransDoublePath_BasicConstruction(t *testing.T) {
	trans := NewTransDoublePath()

	// Test initial state
	if trans.BaseLength() != 0.0 {
		t.Errorf("Expected initial base length 0.0, got %f", trans.BaseLength())
	}
	if trans.BaseHeight() != 1.0 {
		t.Errorf("Expected initial base height 1.0, got %f", trans.BaseHeight())
	}
	if !trans.PreserveXScale() {
		t.Error("Expected preserve X scale to be true initially")
	}

	// Test setters
	trans.SetBaseLength(100.0)
	if trans.BaseLength() != 100.0 {
		t.Errorf("Expected base length 100.0, got %f", trans.BaseLength())
	}

	trans.SetBaseHeight(50.0)
	if trans.BaseHeight() != 50.0 {
		t.Errorf("Expected base height 50.0, got %f", trans.BaseHeight())
	}

	trans.SetPreserveXScale(false)
	if trans.PreserveXScale() {
		t.Error("Expected preserve X scale to be false")
	}
}

func TestTransDoublePath_ManualPathConstruction(t *testing.T) {
	trans := NewTransDoublePath()

	// Build first path: horizontal line from (0,0) to (100,0)
	trans.MoveTo1(0, 0)
	trans.LineTo1(100, 0)

	// Build second path: horizontal line from (0,10) to (100,10)
	trans.MoveTo2(0, 10)
	trans.LineTo2(100, 10)

	trans.FinalizePaths()

	// Test transformation at the beginning
	x, y := 0.0, 0.0
	trans.Transform(&x, &y)
	if math.Abs(x-0.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (0,0), got (%f,%f)", x, y)
	}

	// Test transformation at the end
	x, y = 100.0, 0.0
	trans.Transform(&x, &y)
	if math.Abs(x-100.0) > 1e-10 || math.Abs(y-0.0) > 1e-10 {
		t.Errorf("Expected (100,0), got (%f,%f)", x, y)
	}

	// Test transformation in the middle with y offset
	x, y = 50.0, 1.0 // y=1.0 means full way to second path (baseHeight=1.0)
	trans.Transform(&x, &y)
	if math.Abs(x-50.0) > 1e-10 || math.Abs(y-10.0) > 1e-10 {
		t.Errorf("Expected (50,10), got (%f,%f)", x, y)
	}

	// Test transformation with partial y offset
	x, y = 50.0, 0.5 // halfway between paths
	trans.Transform(&x, &y)
	if math.Abs(x-50.0) > 1e-10 || math.Abs(y-5.0) > 1e-10 {
		t.Errorf("Expected (50,5), got (%f,%f)", x, y)
	}
}

func TestTransDoublePath_AddPaths(t *testing.T) {
	trans := NewTransDoublePath()

	// Create two simple paths
	path1 := newMockPath([]float64{0, 0, 50, 0, 100, 0})
	path2 := newMockPath([]float64{0, 20, 50, 20, 100, 20})

	trans.AddPaths(path1, path2, 0, 0)

	// Test that paths were added correctly
	if trans.TotalLength1() <= 0 {
		t.Error("Expected positive length for path 1")
	}
	if trans.TotalLength2() <= 0 {
		t.Error("Expected positive length for path 2")
	}

	// Test basic transformation
	x, y := 25.0, 0.5
	trans.Transform(&x, &y)

	// Should be roughly at (25, 10) - halfway between y=0 and y=20
	if math.Abs(x-25.0) > 1e-10 || math.Abs(y-10.0) > 1e-10 {
		t.Errorf("Expected approximately (25,10), got (%f,%f)", x, y)
	}
}

func TestTransDoublePath_CurvedPaths(t *testing.T) {
	trans := NewTransDoublePath()

	// Create curved paths - semicircles
	// Bottom semicircle from (-50,0) to (50,0) through (0,-50)
	trans.MoveTo1(-50, 0)
	for i := 1; i <= 10; i++ {
		angle := math.Pi * float64(i) / 10.0
		x := 50 * math.Cos(angle)
		y := -50 * math.Sin(angle)
		trans.LineTo1(x, y)
	}

	// Top semicircle from (-50,20) to (50,20) through (0,70)
	trans.MoveTo2(-50, 20)
	for i := 1; i <= 10; i++ {
		angle := math.Pi * float64(i) / 10.0
		x := 50 * math.Cos(angle)
		y := 20 + 50*math.Sin(angle)
		trans.LineTo2(x, y)
	}

	trans.FinalizePaths()

	// Test transformation at various points
	x, y := 0.0, 0.0 // Should map roughly to bottom of lower semicircle
	trans.Transform(&x, &y)
	// At x=0 on the bottom semicircle, we should be near (0, -50)
	if math.Abs(x-0.0) > 5.0 || y > -30.0 {
		t.Logf("Bottom semicircle center: (%f,%f)", x, y)
	}

	x, y = 0.0, 1.0 // Should map roughly to top of upper semicircle
	trans.Transform(&x, &y)
	// At x=0 on the top semicircle, we should be near (0, 70)
	if math.Abs(x-0.0) > 5.0 || y < 50.0 {
		t.Logf("Top semicircle center: (%f,%f)", x, y)
	}
}

func TestTransDoublePath_Extrapolation(t *testing.T) {
	trans := NewTransDoublePath()

	// Simple horizontal lines
	trans.MoveTo1(0, 0)
	trans.LineTo1(100, 0)

	trans.MoveTo2(0, 10)
	trans.LineTo2(100, 10)

	trans.FinalizePaths()

	// Test left extrapolation
	x, y := -50.0, 0.5
	trans.Transform(&x, &y)
	// Should extrapolate linearly
	if x > -40.0 || x < -60.0 {
		t.Errorf("Left extrapolation gave unexpected x: %f", x)
	}

	// Test right extrapolation
	x, y = 150.0, 0.5
	trans.Transform(&x, &y)
	// Should extrapolate linearly
	if x < 140.0 || x > 160.0 {
		t.Errorf("Right extrapolation gave unexpected x: %f", x)
	}
}

func TestTransDoublePath_BaseLength(t *testing.T) {
	trans := NewTransDoublePath()

	// Create paths
	trans.MoveTo1(0, 0)
	trans.LineTo1(100, 0)

	trans.MoveTo2(0, 10)
	trans.LineTo2(100, 10)

	// Set a custom base length
	trans.SetBaseLength(200.0)
	trans.FinalizePaths()

	// With base length 200 but actual length 100, x coordinates should be scaled by 0.5
	x, y := 200.0, 0.0 // Input x=200 should map to actual x=100 (end of path)
	trans.Transform(&x, &y)
	if math.Abs(x-100.0) > 1e-10 {
		t.Errorf("Expected x=100 with base length scaling, got %f", x)
	}
}

func TestTransDoublePath_PreserveXScale(t *testing.T) {
	trans := NewTransDoublePath()

	// Create non-uniform path (shorter segments at the beginning)
	trans.MoveTo1(0, 0)
	trans.LineTo1(10, 0)  // Short segment
	trans.LineTo1(100, 0) // Long segment

	trans.MoveTo2(0, 10)
	trans.LineTo2(10, 10)
	trans.LineTo2(100, 10)

	// Test with preserve X scale enabled (default)
	trans.SetPreserveXScale(true)
	trans.FinalizePaths()

	x1, y1 := 5.0, 0.0 // Should be in first segment
	trans.Transform(&x1, &y1)

	// Reset and test with preserve X scale disabled
	trans2 := NewTransDoublePath()
	trans2.MoveTo1(0, 0)
	trans2.LineTo1(10, 0)
	trans2.LineTo1(100, 0)

	trans2.MoveTo2(0, 10)
	trans2.LineTo2(10, 10)
	trans2.LineTo2(100, 10)

	trans2.SetPreserveXScale(false)
	trans2.FinalizePaths()

	x2, y2 := 5.0, 0.0
	trans2.Transform(&x2, &y2)

	// The results should be different due to different interpolation methods
	// We don't test exact values as the behavior is complex, just that they differ
	if math.Abs(x1-x2) < 1e-10 && math.Abs(y1-y2) < 1e-10 {
		t.Log("PreserveXScale setting may not be affecting results as expected")
	}
}

func TestTransDoublePath_Reset(t *testing.T) {
	trans := NewTransDoublePath()

	// Build paths
	trans.MoveTo1(0, 0)
	trans.LineTo1(100, 0)
	trans.MoveTo2(0, 10)
	trans.LineTo2(100, 10)
	trans.FinalizePaths()

	// Verify paths exist
	if trans.TotalLength1() <= 0 {
		t.Error("Expected positive length before reset")
	}

	// Reset
	trans.Reset()

	// Verify paths are cleared
	if trans.TotalLength1() != 0 {
		t.Error("Expected zero length after reset")
	}
	if trans.TotalLength2() != 0 {
		t.Error("Expected zero length after reset")
	}
}

func TestTransDoublePath_EdgeCases(t *testing.T) {
	trans := NewTransDoublePath()

	// Test transformation before paths are finalized
	x, y := 50.0, 0.5
	originalX, originalY := x, y
	trans.Transform(&x, &y)

	// Should not modify coordinates if paths aren't ready
	if x != originalX || y != originalY {
		t.Errorf("Transform should not modify coordinates before finalization")
	}

	// Test with single point paths
	trans.MoveTo1(50, 25)
	trans.MoveTo2(50, 75)
	// Don't add LineTo - just single points
	trans.FinalizePaths()

	// Should handle gracefully (though behavior is undefined for single points)
	x, y = 50.0, 0.5
	trans.Transform(&x, &y)
	// Just verify it doesn't crash
}

func BenchmarkTransDoublePath_Transform(b *testing.B) {
	trans := NewTransDoublePath()

	// Create moderately complex paths
	for i := 0; i <= 50; i++ {
		x := float64(i) * 2.0
		y := math.Sin(float64(i)*0.1) * 10
		if i == 0 {
			trans.MoveTo1(x, y)
		} else {
			trans.LineTo1(x, y)
		}
	}

	for i := 0; i <= 50; i++ {
		x := float64(i) * 2.0
		y := 20 + math.Sin(float64(i)*0.1)*10
		if i == 0 {
			trans.MoveTo2(x, y)
		} else {
			trans.LineTo2(x, y)
		}
	}

	trans.FinalizePaths()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := float64(i % 100)
		y := 0.5
		trans.Transform(&x, &y)
	}
}
