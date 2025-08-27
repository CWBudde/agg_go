package integration

import (
	"testing"

	agg "agg_go"
)

// TestHighLevelAPI tests the higher-level Context API
func TestHighLevelAPI(t *testing.T) {
	width, height := 20, 20

	// Use the higher-level Context API that actually works
	ctx := agg.NewContext(width, height)

	// Clear to blue
	ctx.Clear(agg.Blue)

	// Get the underlying buffer to inspect
	img := ctx.GetImage()

	// Check the pixel manually
	stride := width * 4
	pixel := getPixel(img.Data, stride, 10, 10)
	t.Logf("After clear with Context API: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	// Draw a red rectangle
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(8, 8, 4, 4) // 4x4 rectangle at (8,8)

	img = ctx.GetImage()
	pixel = getPixel(img.Data, stride, 10, 10)
	t.Logf("After rectangle with Context API: RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])

	if pixel[0] < 200 || pixel[1] > 50 || pixel[2] > 50 {
		t.Errorf("Rectangle should be red, got RGBA(%d,%d,%d,%d)", pixel[0], pixel[1], pixel[2], pixel[3])
	} else {
		t.Log("SUCCESS: Rectangle rendering works with Context API!")
	}
}
