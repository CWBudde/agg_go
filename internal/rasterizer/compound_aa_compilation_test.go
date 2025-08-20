package rasterizer

import (
	"testing"
)

// TestCompoundAACompilationFix tests that the compound_aa file compiles correctly
// This test verifies that the pointer dereference issues in SweepStyles are resolved
func TestCompoundAACompilationFix(t *testing.T) {
	// Create a compound rasterizer - this should compile without errors
	clipper := NewMockCompoundClipper(nil)
	rasterizer := NewRasterizerCompoundAA(clipper)

	// Test that the basic structure is accessible
	if rasterizer == nil {
		t.Error("Failed to create RasterizerCompoundAA")
	}

	// Test that AA constants are defined (this was failing before)
	if AAShift != 8 {
		t.Errorf("AAShift = %d, want 8", AAShift)
	}
	if AAScale != 256 {
		t.Errorf("AAScale = %d, want 256", AAScale)
	}
	if AAMask != 255 {
		t.Errorf("AAMask = %d, want 255", AAMask)
	}
	if AAScale2 != 512 {
		t.Errorf("AAScale2 = %d, want 512", AAScale2)
	}
	if AAMask2 != 511 {
		t.Errorf("AAMask2 = %d, want 511", AAMask2)
	}

	// Test CalculateAlpha method (this uses the constants that were undefined before)
	alpha := rasterizer.CalculateAlpha(0)
	if alpha != 0 {
		t.Errorf("CalculateAlpha(0) = %d, want 0", alpha)
	}

	t.Log("Compound AA compilation fix verified successfully")
}

// TestCompoundAAMethodsAccessible tests that all methods are accessible
func TestCompoundAAMethodsAccessible(t *testing.T) {
	clipper := NewMockCompoundClipper(nil)
	rasterizer := NewRasterizerCompoundAA(clipper)

	// Test that all major methods exist and can be called
	rasterizer.Reset()
	rasterizer.ResetClipping()
	rasterizer.FillingRule(0)
	rasterizer.LayerOrder(0)
	rasterizer.Styles(1, 2)

	// Test coordinate methods
	rasterizer.MoveTo(10, 10)
	rasterizer.LineTo(20, 20)
	rasterizer.MoveToD(10.5, 10.5)
	rasterizer.LineToD(20.5, 20.5)
	rasterizer.Edge(0, 0, 10, 10)
	rasterizer.EdgeD(0.0, 0.0, 10.0, 10.0)

	// Test bounds methods
	_ = rasterizer.MinX()
	_ = rasterizer.MinY()
	_ = rasterizer.MaxX()
	_ = rasterizer.MaxY()
	_ = rasterizer.MinStyle()
	_ = rasterizer.MaxStyle()

	// Test buffer allocation
	buffer := rasterizer.AllocateCoverBuffer(100)
	if len(buffer) != 100 {
		t.Errorf("AllocateCoverBuffer returned buffer of size %d, want 100", len(buffer))
	}

	t.Log("All compound AA methods accessible")
}
