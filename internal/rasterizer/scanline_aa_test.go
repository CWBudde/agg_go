package rasterizer

import (
	"testing"

	"agg_go/internal/basics"
)

// MockScanline implements ScanlineInterface for testing
type MockScanline struct {
	cells []MockCell
	spans []MockSpan
	y     int
}

type MockCell struct {
	x     int
	cover uint32
}

type MockSpan struct {
	x, len int
	cover  uint32
}

func (ms *MockScanline) ResetSpans() {
	ms.cells = ms.cells[:0]
	ms.spans = ms.spans[:0]
}

func (ms *MockScanline) AddCell(x int, cover uint32) {
	ms.cells = append(ms.cells, MockCell{x: x, cover: cover})
}

func (ms *MockScanline) AddSpan(x, len int, cover uint32) {
	ms.spans = append(ms.spans, MockSpan{x: x, len: len, cover: cover})
}

func (ms *MockScanline) Finalize(y int) {
	ms.y = y
}

func (ms *MockScanline) NumSpans() int {
	return len(ms.cells) + len(ms.spans)
}

// MockClip implements ClipInterface for testing
type MockClip struct {
	moveToX, moveToY float64
	lineToX, lineToY float64
	clipX1, clipY1   float64
	clipX2, clipY2   float64
	clipping         bool
}

func (mc *MockClip) ResetClipping() {
	mc.clipping = false
}

func (mc *MockClip) ClipBox(x1, y1, x2, y2 float64) {
	mc.clipX1, mc.clipY1 = x1, y1
	mc.clipX2, mc.clipY2 = x2, y2
	mc.clipping = true
}

func (mc *MockClip) MoveTo(x, y float64) {
	mc.moveToX, mc.moveToY = x, y
}

func (mc *MockClip) LineTo(outline RasterizerInterface, x, y float64) {
	mc.lineToX, mc.lineToY = x, y
}

func TestNewRasterizerScanlineAA(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

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

	// Check that gamma table is initialized linearly
	for i := 0; i < AAScale; i++ {
		if r.gamma[i] != i {
			t.Errorf("Expected gamma[%d] = %d, got %d", i, i, r.gamma[i])
		}
	}
}

func TestRasterizerScanlineAA_Reset(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)
	r.status = StatusLineTo

	r.Reset()

	if r.status != StatusInitial {
		t.Error("Expected status to be reset to StatusInitial")
	}
}

func TestRasterizerScanlineAA_FillingRule(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.FillingRule(basics.FillEvenOdd)

	if r.fillingRule != basics.FillEvenOdd {
		t.Error("Expected filling rule to be set to FillEvenOdd")
	}
}

func TestRasterizerScanlineAA_AutoClose(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.AutoClose(false)

	if r.autoClose {
		t.Error("Expected autoClose to be false")
	}
}

func TestRasterizerScanlineAA_SetGamma(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	// Set a simple gamma function (square root)
	r.SetGamma(func(x float64) float64 {
		return x * x // Actually x^2, not square root, but easier to test
	})

	// Check a few values
	expected0 := basics.URound(0.0 * AAMask)
	if r.gamma[0] != int(expected0) {
		t.Errorf("Expected gamma[0] = %d, got %d", expected0, r.gamma[0])
	}

	expected255 := basics.URound(1.0 * AAMask)
	if r.gamma[255] != int(expected255) {
		t.Errorf("Expected gamma[255] = %d, got %d", expected255, r.gamma[255])
	}
}

func TestRasterizerScanlineAA_ApplyGamma(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	// With linear gamma, ApplyGamma should return the same value
	result := r.ApplyGamma(128)
	if result != 128 {
		t.Errorf("Expected ApplyGamma(128) = 128, got %d", result)
	}

	// Test with custom gamma
	r.SetGamma(func(x float64) float64 { return x * 0.5 })
	result = r.ApplyGamma(128)
	expected := uint32(basics.URound(0.5 * 128))
	if result != expected {
		t.Errorf("Expected ApplyGamma(128) = %d, got %d", expected, result)
	}
}

func TestRasterizerScanlineAA_MoveTo(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.MoveTo(100, 200)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo")
	}

	if r.startX != 100 || r.startY != 200 {
		t.Errorf("Expected start position (100, 200), got (%d, %d)", r.startX, r.startY)
	}

	if clip.moveToX != 100.0 || clip.moveToY != 200.0 {
		t.Errorf("Expected clipper MoveTo called with (100, 200), got (%f, %f)", clip.moveToX, clip.moveToY)
	}
}

func TestRasterizerScanlineAA_LineTo(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.LineTo(300, 400)

	if r.status != StatusLineTo {
		t.Error("Expected status to be StatusLineTo")
	}

	if clip.lineToX != 300.0 || clip.lineToY != 400.0 {
		t.Errorf("Expected clipper LineTo called with (300, 400), got (%f, %f)", clip.lineToX, clip.lineToY)
	}
}

func TestRasterizerScanlineAA_MoveToD(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.MoveToD(10.5, 20.25)

	if r.status != StatusMoveTo {
		t.Error("Expected status to be StatusMoveTo")
	}

	expectedX := basics.IRound(10.5 * basics.PolySubpixelScale)
	expectedY := basics.IRound(20.25 * basics.PolySubpixelScale)

	if r.startX != expectedX || r.startY != expectedY {
		t.Errorf("Expected start position (%d, %d), got (%d, %d)", expectedX, expectedY, r.startX, r.startY)
	}

	if clip.moveToX != 10.5 || clip.moveToY != 20.25 {
		t.Errorf("Expected clipper MoveTo called with (10.5, 20.25), got (%f, %f)", clip.moveToX, clip.moveToY)
	}
}

func TestRasterizerScanlineAA_CalculateAlpha(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	// Test with filling rule FillNonZero
	r.FillingRule(basics.FillNonZero)

	// Area that would give full coverage
	area := AAScale << (basics.PolySubpixelShift*2 + 1 - AAShift)
	alpha := r.CalculateAlpha(area)

	if alpha == 0 {
		t.Error("Expected non-zero alpha for significant area")
	}

	// Test with filling rule FillEvenOdd
	r.FillingRule(basics.FillEvenOdd)
	alpha2 := r.CalculateAlpha(area)

	// The result may be different due to even-odd logic
	if alpha2 == 0 && alpha != 0 {
		t.Error("Even-odd filling rule should also produce non-zero alpha for this area")
	}
}

func TestRasterizerScanlineAA_ClipBox(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	r.ClipBox(10.0, 20.0, 100.0, 200.0)

	if !clip.clipping {
		t.Error("Expected clipping to be enabled")
	}

	if clip.clipX1 != 10.0 || clip.clipY1 != 20.0 || clip.clipX2 != 100.0 || clip.clipY2 != 200.0 {
		t.Errorf("Expected clip box (10, 20, 100, 200), got (%f, %f, %f, %f)",
			clip.clipX1, clip.clipY1, clip.clipX2, clip.clipY2)
	}
}

func TestRasterizerScanlineAA_ResetClipping(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

	// Enable clipping first
	clip.clipping = true

	r.ResetClipping()

	if clip.clipping {
		t.Error("Expected clipping to be disabled")
	}
}

func TestRasterizerScanlineAA_AddVertex(t *testing.T) {
	clip := &MockClip{}
	r := NewRasterizerScanlineAA[*MockClip, RasConvInt](1024, clip)

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
	r.AddVertex(0.0, 0.0, uint32(basics.PathCmdEndPoly)|uint32(basics.PathFlagsClose))
	if r.status != StatusClosed {
		t.Error("Expected status to be StatusClosed after Close command")
	}
}
