package integration

import (
	"testing"

	"agg_go/internal/agg2d"
)

// TestBasicRendering tests the most basic rendering functionality
func TestBasicRendering(t *testing.T) {
	width, height := 10, 10
	stride := width * 4 // RGBA
	buffer := make([]uint8, height*stride)

	// Initialize buffer with known pattern
	for i := 0; i < len(buffer); i += 4 {
		buffer[i] = 128   // R
		buffer[i+1] = 128 // G
		buffer[i+2] = 128 // B
		buffer[i+3] = 255 // A
	}

	// Verify manual initialization worked
	pixel := getPixel(buffer, stride, 5, 5)
	t.Logf("Manual pixel: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	// Now test AGG2D
	ctx := agg2d.NewAgg2D()
	if ctx == nil {
		t.Fatal("Failed to create Agg2D context")
	}

	// Attach buffer
	ctx.Attach(buffer, width, height, stride)

	// Test clear
	ctx.ClearAll(agg2d.Color{255, 0, 0, 255}) // Red

	pixel = getPixel(buffer, stride, 5, 5)
	t.Logf("After clear: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	if pixel[0] != 255 || pixel[1] != 0 || pixel[2] != 0 || pixel[3] != 255 {
		t.Errorf("Clear failed: expected red (255,0,0,255), got RGBA(%d,%d,%d,%d)",
			pixel[0], pixel[1], pixel[2], pixel[3])
	}

	// Test simple rectangle
	ctx.FillColor(agg2d.Color{0, 255, 0, 255}) // Green
	ctx.ResetPath()
	ctx.Rectangle(2, 2, 8, 8)
	ctx.DrawPath(agg2d.FillOnly)

	pixel = getPixel(buffer, stride, 5, 5)
	t.Logf("After rectangle: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	// Dump entire buffer for debugging
	t.Logf("Buffer contents:")
	for y := 0; y < height; y++ {
		line := ""
		for x := 0; x < width; x++ {
			p := getPixel(buffer, stride, x, y)
			if p[0] > 200 && p[1] < 50 && p[2] < 50 {
				line += "R" // Red
			} else if p[1] > 200 && p[0] < 50 && p[2] < 50 {
				line += "G" // Green
			} else if p[0] > 200 && p[1] > 200 && p[2] > 200 {
				line += "W" // White
			} else {
				line += "?" // Other
			}
		}
		t.Logf("Row %d: %s", y, line)
	}
}
