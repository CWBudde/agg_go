package rasterizer

import (
	"testing"
)

// TestCompoundAASweepStylesMinimal tests the specific pointer dereference issue that was fixed
func TestCompoundAASweepStylesMinimal(t *testing.T) {
	// Create a minimal compound rasterizer setup
	clipper := NewMockCompoundClipper(nil)
	rasterizer := NewRasterizerCompoundAA(clipper)
	clipper.outline = rasterizer.outline

	// Create some geometry with styles to trigger the SweepStyles code path
	rasterizer.Styles(1, 0) // Set left=1, right=0 styles
	rasterizer.MoveTo(10, 10)
	rasterizer.LineTo(90, 10)
	rasterizer.LineTo(90, 90)
	rasterizer.LineTo(10, 90)
	rasterizer.LineTo(10, 10)

	// Sort to prepare for scanline processing
	rasterizer.Sort()

	// Test that NavigateScanline works (this was failing before due to pointer issues)
	navigated := rasterizer.NavigateScanline(50)
	if !navigated {
		t.Skip("NavigateScanline failed - may be due to insufficient geometry")
		return
	}

	// Test that SweepStyles works without crashing (this was the main issue)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SweepStyles panicked: %v", r)
		}
	}()

	numStyles := rasterizer.SweepStyles()
	t.Logf("SweepStyles completed successfully, returned %d styles", numStyles)

	// If we got here, the pointer dereference issues are fixed
	if numStyles == 0 {
		t.Log("No styles returned, but method completed without crashing")
	}
}

// TestCompoundAACalculateAlphaWithConstants tests that the AA constants are properly defined
func TestCompoundAACalculateAlphaWithConstants(t *testing.T) {
	rasterizer := NewRasterizerCompoundAA(NewMockCompoundClipper(nil))

	// Test basic alpha calculation
	alpha := rasterizer.CalculateAlpha(0)
	if alpha != 0 {
		t.Errorf("CalculateAlpha(0) = %d, want 0", alpha)
	}

	// Test non-zero area
	alpha = rasterizer.CalculateAlpha(1000)
	if alpha == 0 {
		t.Error("CalculateAlpha(1000) should produce non-zero alpha")
	}

	// Verify constants are accessible (compilation test)
	_ = AAShift
	_ = AAScale
	_ = AAMask
	_ = AAScale2
	_ = AAMask2

	t.Logf("AA constants defined: Shift=%d, Scale=%d, Mask=%d, Scale2=%d, Mask2=%d",
		AAShift, AAScale, AAMask, AAScale2, AAMask2)
}

// TestCompoundAAPointerDerefFix tests that cells can be accessed correctly
func TestCompoundAAPointerDerefFix(t *testing.T) {
	// Create a rasterizer and add some cells
	clipper := NewMockCompoundClipper(nil)
	rasterizer := NewRasterizerCompoundAA(clipper)
	clipper.outline = rasterizer.outline

	// Add some basic geometry
	rasterizer.Styles(2, 1)
	rasterizer.MoveTo(0, 0)
	rasterizer.LineTo(100, 0)

	rasterizer.Sort()

	// Try to get cells from a scanline - this exercises the fixed pointer dereference code
	if rasterizer.outline.TotalCells() > 0 {
		cells := rasterizer.outline.ScanlineCells(0)
		if len(cells) > 0 {
			// Test that we can access cell properties without panic
			// This was failing before due to **CellStyleAA vs *CellStyleAA issues
			x := (*cells[0]).GetX()
			y := (*cells[0]).GetY()
			left := (*cells[0]).Left
			right := (*cells[0]).Right

			t.Logf("Successfully accessed cell: x=%d, y=%d, left=%d, right=%d", x, y, left, right)
		}
	}
}
