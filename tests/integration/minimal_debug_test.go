package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestMinimalRendering - The absolute simplest test to debug rendering
func TestMinimalRendering(t *testing.T) {
	// Step 1: Test buffer creation and manual manipulation
	width, height := 10, 10
	stride := width * 4
	buffer := make([]uint8, height*stride)

	// Manually set one pixel to red to verify buffer access works
	setPixel(buffer, stride, 5, 5, [4]uint8{255, 0, 0, 255})
	pixel := getPixel(buffer, stride, 5, 5)
	t.Logf("Manual pixel set: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	if pixel[0] != 255 || pixel[1] != 0 || pixel[2] != 0 || pixel[3] != 255 {
		t.Fatal("Basic buffer manipulation failed - fundamental issue")
	}

	// Step 2: Test AGG2D creation and buffer attachment
	ctx := agg2d.NewAgg2D()
	if ctx == nil {
		t.Fatal("AGG2D creation failed")
	}
	t.Log("✓ AGG2D created successfully")

	// Clear buffer first
	for i := range buffer {
		buffer[i] = 0
	}

	ctx.Attach(buffer, width, height, stride)
	t.Log("✓ Buffer attached successfully")

	// Step 3: Test clear operation (this works in our tests)
	ctx.ClearAll(agg2d.Color{0, 255, 0, 255}) // Green
	pixel = getPixel(buffer, stride, 5, 5)
	t.Logf("After clear: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	if pixel[1] != 255 {
		t.Fatal("Clear operation failed - AGG2D buffer attachment issue")
	}
	t.Log("✓ Clear operation works")

	// Step 4: Test the minimal path rendering
	ctx.FillColor(agg2d.Color{255, 0, 0, 255}) // Red
	t.Log("✓ Fill color set")

	ctx.ResetPath()
	t.Log("✓ Path reset")

	// Use the simplest possible path - single pixel rectangle
	ctx.MoveTo(5, 5)
	ctx.LineTo(6, 5)
	ctx.LineTo(6, 6)
	ctx.LineTo(5, 6)
	ctx.ClosePolygon()
	t.Log("✓ Simple 1x1 rectangle path created")

	// Check buffer BEFORE DrawPath
	beforePixel := getPixel(buffer, stride, 5, 5)
	t.Logf("Before DrawPath: RGBA(%d,%d,%d,%d)", beforePixel[0], beforePixel[1], beforePixel[2], beforePixel[3])

	// This is where it fails
	ctx.DrawPath(agg2d.FillOnly)
	t.Log("✓ DrawPath called")

	// Check buffer AFTER DrawPath
	afterPixel := getPixel(buffer, stride, 5, 5)
	t.Logf("After DrawPath: RGBA(%d,%d,%d,%d)", afterPixel[0], afterPixel[1], afterPixel[2], afterPixel[3])

	if afterPixel[0] < 200 { // Should be red
		t.Error("❌ DrawPath failed - no change in buffer")

		// Dump entire buffer to see if ANYTHING changed
		t.Log("Full buffer dump:")
		changed := 0
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				p := getPixel(buffer, stride, x, y)
				if p[0] != 0 || p[1] != 255 || p[2] != 0 { // Not the green background
					t.Logf("  Pixel (%d,%d): RGBA(%d,%d,%d,%d)", x, y, p[0], p[1], p[2], p[3])
					changed++
				}
			}
		}
		t.Logf("Total pixels changed from background: %d", changed)

		if changed == 0 {
			t.Error("No pixels changed at all - DrawPath is completely non-functional")
		}
	} else {
		t.Log("✓ SUCCESS: DrawPath works!")
	}
}
