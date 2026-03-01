package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

func TestNewRasterizerScanlineAANoGamma(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	if r == nil {
		t.Fatal("Expected non-nil rasterizer")
	}

	if r.fillingRule != basics.FillNonZero {
		t.Error("Expected default filling rule to be FillNonZero")
	}

	if !r.autoClose {
		t.Error("Expected autoClose to be true by default")
	}

	if r.status != StatusInitial {
		t.Error("Expected initial status to be StatusInitial")
	}
}

func TestRasterizerScanlineAANoGammaApplyGamma(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	// Test that ApplyGamma returns the input unchanged (no gamma correction)
	testValues := []int{0, 50, 100, 128, 200, 255}
	for _, val := range testValues {
		result := r.ApplyGamma(val)
		if result != uint8(val) {
			t.Errorf("Expected ApplyGamma(%d) = %d, got %d", val, val, result)
		}
	}
}

func TestRasterizerScanlineAANoGammaReset(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	// Change state
	r.status = StatusMoveTo
	r.startX = 100
	r.startY = 200

	// Reset should restore initial state
	r.Reset()

	if r.status != StatusInitial {
		t.Error("Expected status to be StatusInitial after reset")
	}
}

func TestRasterizerScanlineAANoGammaFillingRule(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	// Test setting different filling rules
	r.FillingRule(basics.FillEvenOdd)
	if r.fillingRule != basics.FillEvenOdd {
		t.Error("Expected filling rule to be FillEvenOdd")
	}

	r.FillingRule(basics.FillNonZero)
	if r.fillingRule != basics.FillNonZero {
		t.Error("Expected filling rule to be FillNonZero")
	}
}

func TestRasterizerScanlineAANoGammaAutoClose(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	// Test auto close setting
	r.AutoClose(false)
	if r.autoClose {
		t.Error("Expected autoClose to be false")
	}

	r.AutoClose(true)
	if !r.autoClose {
		t.Error("Expected autoClose to be true")
	}
}

func TestRasterizerScanlineAANoGammaMoveTo(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, clipper)

	// Test integer MoveTo
	x, y := 100, 200
	r.MoveTo(x, y)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveTo")
	}

	// Verify the clipper was called with correct coordinates
	if clipper.moveToX != float64(x) || clipper.moveToY != float64(y) {
		t.Errorf("Expected clipper MoveTo called with (%d, %d), got (%f, %f)", x, y, clipper.moveToX, clipper.moveToY)
	}
}

func TestRasterizerScanlineAANoGammaMoveToD(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, clipper)

	// Test floating point MoveTo
	x, y := 10.5, 20.7
	r.MoveToD(x, y)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveToD")
	}

	// Verify the clipper was called with correct coordinates
	if clipper.moveToX != x || clipper.moveToY != y {
		t.Errorf("Expected clipper MoveTo called with (%f, %f), got (%f, %f)", x, y, clipper.moveToX, clipper.moveToY)
	}
}

func TestRasterizerScanlineAANoGammaLineTo(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, clipper)

	// Must start with MoveTo
	r.MoveTo(0, 0)

	// Test LineTo
	r.LineTo(100, 200)

	if r.status != StatusLineTo {
		t.Error("Expected status to be StatusLineTo after LineTo")
	}
}

func TestRasterizerScanlineAANoGammaClosePolygon(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Create a simple path
	r.MoveTo(0, 0)
	r.LineTo(100, 0)
	r.LineTo(100, 100)

	// Close polygon
	r.ClosePolygon()

	if r.status != StatusClosed {
		t.Error("Expected status to be StatusClosed after ClosePolygon")
	}
}

func TestRasterizerScanlineAANoGammaAddVertex(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Test MoveTo command
	r.AddVertex(10.0, 20.0, uint32(basics.PathCmdMoveTo))
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after MoveTo command")
	}

	// Test LineTo command
	r.AddVertex(30.0, 40.0, uint32(basics.PathCmdLineTo))
	if r.status != StatusLineTo {
		t.Error("Expected status to be StatusLineTo after LineTo command")
	}

	// Test Close command
	r.AddVertex(0.0, 0.0, uint32(basics.PathFlagsClose))
	if r.status != StatusClosed {
		t.Error("Expected status to be StatusClosed after Close command")
	}
}

func TestRasterizerScanlineAANoGammaEdge(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Test Edge with integer coordinates
	r.Edge(0, 0, 100, 100)
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after Edge")
	}

	// Test EdgeD with floating point coordinates
	r.EdgeD(10.5, 20.5, 30.7, 40.9)
	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo after EdgeD")
	}
}

func TestRasterizerScanlineAANoGammaCalculateAlpha(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	tests := []struct {
		area     int
		expected uint32
	}{
		{0, 0},
		{AAScale << (basics.PolySubpixelShift*2 + 1 - AAShift), AAScale - 1}, // Max value
		{-100, 0}, // Negative area
	}

	for _, test := range tests {
		result := r.CalculateAlpha(test.area)
		if result > AAMask {
			t.Errorf("CalculateAlpha(%d) returned %d, which exceeds AAMask (%d)", test.area, result, AAMask)
		}
	}
}

func TestRasterizerScanlineAANoGammaEvenOddFilling(t *testing.T) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.FillingRule(basics.FillEvenOdd)

	// Test even-odd specific calculation
	area := AAScale2 << (basics.PolySubpixelShift*2 + 1 - AAShift)
	alpha := r.CalculateAlpha(area)

	// For even-odd, should handle wrap-around
	if alpha > AAMask {
		t.Errorf("Even-odd alpha calculation failed: got %d, max should be %d", alpha, AAMask)
	}
}

func TestRasterizerScanlineAANoGammaRewindScanlines(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Empty rasterizer should return false
	if r.RewindScanlines() {
		t.Error("Expected RewindScanlines to return false for empty rasterizer")
	}

	// Add some geometry
	r.MoveTo(0, 0)
	r.LineTo(100*basics.PolySubpixelScale, 0)
	r.LineTo(100*basics.PolySubpixelScale, 100*basics.PolySubpixelScale)
	r.ClosePolygon()

	// Now should return true (though we can't easily test the actual rasterization without more setup)
	// The test here is mainly to ensure the method doesn't panic and handles the basic case
}

func TestRasterizerScanlineAANoGammaNavigateScanline(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Empty rasterizer should return false
	if r.NavigateScanline(100) {
		t.Error("Expected NavigateScanline to return false for empty rasterizer")
	}
}

func TestRasterizerScanlineAANoGammaHitTest(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Empty rasterizer should return false for any hit test
	if r.HitTest(50, 50) {
		t.Error("Expected HitTest to return false for empty rasterizer")
	}
}

func TestRasterizerScanlineAANoGammaSweepScanlineWithPositiveMinY(t *testing.T) {
	ras := NewRasterizerScanlineAANoGamma[int, IntConv, *RasterizerSlNoClip](IntConv{}, NewRasterizerSlNoClip())
	sl := &MockScanline{}

	ras.MoveToD(10, 10)
	ras.LineToD(20, 10)
	ras.LineToD(15, 20)
	ras.ClosePolygon()

	if !ras.RewindScanlines() {
		t.Fatal("Expected rasterizer to contain scanlines")
	}

	if !ras.SweepScanline(sl) {
		t.Fatal("Expected at least one swept scanline")
	}

	if sl.y < 10 {
		t.Fatalf("Expected finalized scanline y to be within rasterized bounds, got %d", sl.y)
	}
	if sl.NumSpans() == 0 {
		t.Fatal("Expected swept scanline to contain spans or cells")
	}
}

// MockVertexSource for testing AddPath
type MockVertexSource struct {
	vertices []vertex
	index    int
}

type vertex struct {
	x, y float64
	cmd  uint32
}

func (mvs *MockVertexSource) Rewind(pathID uint32) {
	mvs.index = 0
}

func (mvs *MockVertexSource) Vertex(x, y *float64) uint32 {
	if mvs.index >= len(mvs.vertices) {
		return uint32(basics.PathCmdStop)
	}

	v := mvs.vertices[mvs.index]
	*x = v.x
	*y = v.y
	mvs.index++
	return v.cmd
}

func TestRasterizerScanlineAANoGammaAddPath(t *testing.T) {
	clipper := &MockClip{}
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
	r.clipper = clipper

	// Create a simple path: MoveTo, LineTo, LineTo, EndPoly+Close
	vs := &MockVertexSource{
		vertices: []vertex{
			{0, 0, uint32(basics.PathCmdMoveTo)},
			{100, 0, uint32(basics.PathCmdLineTo)},
			{100, 100, uint32(basics.PathCmdLineTo)},
			{0, 0, uint32(basics.PathCmdEndPoly) | uint32(basics.PathFlagsClose)},
		},
	}

	r.AddPath(vs, 0)

	// The path should have been processed successfully
	// The final status depends on the last command, but we mainly test that it doesn't panic
	// and processes all commands without error
}

// Benchmark comparison with gamma version to ensure performance benefit
func BenchmarkRasterizerScanlineAANoGamma_Triangle(b *testing.B) {
	clipper := &MockClip{}

	for i := 0; i < b.N; i++ {
		r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})
		r.clipper = clipper

		// Draw a triangle
		r.MoveTo(0, 0)
		r.LineTo(100*basics.PolySubpixelScale, 0)
		r.LineTo(50*basics.PolySubpixelScale, 100*basics.PolySubpixelScale)
		r.ClosePolygon()

		// Test alpha calculation (main performance difference)
		r.CalculateAlpha(1000)
	}
}

func BenchmarkRasterizerScanlineAANoGamma_ApplyGamma(b *testing.B) {
	r := NewRasterizerScanlineAANoGamma[float64, DblConv, *MockClip](DblConv{}, &MockClip{})

	for i := 0; i < b.N; i++ {
		r.ApplyGamma(i % 256)
	}
}

// TODO: Add tests for HitTestScanline if/when it's implemented
