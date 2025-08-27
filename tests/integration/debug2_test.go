package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestDrawingMethods tests different drawing methods to see which work
func TestDrawingMethods(t *testing.T) {
	width, height := 20, 20
	stride := width * 4
	buffer := make([]uint8, height*stride)

	ctx := agg2d.NewAgg2D()
	ctx.Attach(buffer, width, height, stride)

	// Test 1: Clear only
	t.Log("=== Test 1: Clear to blue ===")
	ctx.ClearAll(agg2d.Color{0, 0, 255, 255}) // Blue
	centerPixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Center pixel after clear: RGBA(%d,%d,%d,%d)", centerPixel[0], centerPixel[1], centerPixel[2], centerPixel[3])

	// Test 2: Try to draw a simple line
	t.Log("=== Test 2: Draw line ===")
	ctx.LineColor(agg2d.Color{255, 255, 0, 255}) // Yellow
	ctx.LineWidth(2.0)
	ctx.Line(5, 10, 15, 10) // Horizontal line

	linePixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Line pixel: RGBA(%d,%d,%d,%d)", linePixel[0], linePixel[1], linePixel[2], linePixel[3])

	// Test 3: Try rectangle using different method
	t.Log("=== Test 3: Rectangle using Rectangle() method ===")
	ctx.FillColor(agg2d.Color{255, 0, 255, 255}) // Magenta
	ctx.Rectangle(8, 8, 12, 12)                  // 4x4 rectangle

	rectPixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Rectangle pixel: RGBA(%d,%d,%d,%d)", rectPixel[0], rectPixel[1], rectPixel[2], rectPixel[3])

	// Test 4: Try path-based approach step by step
	t.Log("=== Test 4: Manual path construction ===")
	ctx.ClearAll(agg2d.Color{128, 128, 128, 255}) // Gray background

	ctx.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	t.Log("Set fill color to green")

	ctx.ResetPath()
	t.Log("Reset path")

	ctx.MoveTo(6, 6)
	t.Log("MoveTo(6,6)")

	ctx.LineTo(14, 6)
	t.Log("LineTo(14,6)")

	ctx.LineTo(14, 14)
	t.Log("LineTo(14,14)")

	ctx.LineTo(6, 14)
	t.Log("LineTo(6,14)")

	ctx.ClosePolygon()
	t.Log("ClosePolygon")

	// Check before drawing
	beforePixel := getPixel(buffer, stride, 10, 10)
	t.Logf("Before DrawPath: RGBA(%d,%d,%d,%d)", beforePixel[0], beforePixel[1], beforePixel[2], beforePixel[3])

	ctx.DrawPath(agg2d.FillOnly)
	t.Log("Called DrawPath(FillOnly)")

	afterPixel := getPixel(buffer, stride, 10, 10)
	t.Logf("After DrawPath: RGBA(%d,%d,%d,%d)", afterPixel[0], afterPixel[1], afterPixel[2], afterPixel[3])

	// Check if anything changed anywhere in the buffer
	changed := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := getPixel(buffer, stride, x, y)
			if pixel[0] != 128 || pixel[1] != 128 || pixel[2] != 128 {
				t.Logf("Found change at (%d,%d): RGBA(%d,%d,%d,%d)", x, y, pixel[0], pixel[1], pixel[2], pixel[3])
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Log("No pixels changed after DrawPath - path rendering may not be working")
	}
}
