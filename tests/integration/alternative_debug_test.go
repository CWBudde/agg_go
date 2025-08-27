package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestAlternativeDrawingMethods - Test different drawing methods to find what works
func TestAlternativeDrawingMethods(t *testing.T) {
	width, height := 20, 20
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)
	ctx.ClearAll(agg2d.Color{100, 100, 100, 255}) // Gray background

	// Test 1: Direct Line method (bypasses paths)
	t.Log("=== Testing Line() method ===")
	ctx.LineColor(agg2d.Color{255, 0, 0, 255})
	ctx.LineWidth(2.0)
	ctx.Line(5, 10, 15, 10) // Direct line drawing

	linePixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Line pixel: RGBA(%d,%d,%d,%d)", linePixel[0], linePixel[1], linePixel[2], linePixel[3])

	// Test 2: Rectangle method (might bypass paths)
	ctx.ClearAll(agg2d.Color{100, 100, 100, 255})
	t.Log("=== Testing Rectangle() method ===")
	ctx.FillColor(agg2d.Color{0, 255, 0, 255})
	ctx.Rectangle(8, 8, 12, 12) // Direct rectangle
	// Note: Rectangle() might still need DrawPath()

	rectPixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Rectangle pixel: RGBA(%d,%d,%d,%d)", rectPixel[0], rectPixel[1], rectPixel[2], rectPixel[3])

	// Test 3: AddEllipse method
	ctx.ClearAll(agg2d.Color{100, 100, 100, 255})
	t.Log("=== Testing AddEllipse() + DrawPath ===")
	ctx.FillColor(agg2d.Color{0, 0, 255, 255})
	ctx.ResetPath()
	ctx.AddEllipse(10, 10, 5, 5, agg2d.CCW)
	ctx.DrawPath(agg2d.FillOnly)

	ellipsePixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Ellipse pixel: RGBA(%d,%d,%d,%d)", ellipsePixel[0], ellipsePixel[1], ellipsePixel[2], ellipsePixel[3])

	// Summary
	if linePixel[0] > 200 {
		t.Log("✓ Line() method works")
	} else {
		t.Log("❌ Line() method failed")
	}

	if rectPixel[1] > 200 {
		t.Log("✓ Rectangle() method works")
	} else {
		t.Log("❌ Rectangle() method failed")
	}

	if ellipsePixel[2] > 200 {
		t.Log("✓ AddEllipse() + DrawPath works")
	} else {
		t.Log("❌ AddEllipse() + DrawPath failed")
	}
}
