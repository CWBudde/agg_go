package integration

import (
	"testing"

	agg "agg_go"
)

// TestContextAPIClear tests that Clear fills the entire buffer with the given color.
func TestContextAPIClear(t *testing.T) {
	ctx := agg.NewContext(20, 20)
	ctx.Clear(agg.Blue)

	img := ctx.GetImage()
	stride := 20 * 4
	pixel := getPixel(img.Data, stride, 10, 10)
	if pixel[0] != 0 || pixel[2] != 255 || pixel[3] != 255 {
		t.Errorf("Clear(Blue): expected RGBA(0,0,255,255), got RGBA(%d,%d,%d,%d)",
			pixel[0], pixel[1], pixel[2], pixel[3])
	}
}

// TestContextAPIFillRectangle tests that FillRectangle renders pixels at the correct
// screen position using the high-level public Context API.
func TestContextAPIFillRectangle(t *testing.T) {
	width, height := 20, 20
	ctx := agg.NewContext(width, height)
	ctx.Clear(agg.Blue)
	ctx.SetColor(agg.Red)
	ctx.FillRectangle(8, 8, 4, 4) // 4×4 rectangle at (8,8)

	img := ctx.GetImage()
	stride := width * 4
	center := getPixel(img.Data, stride, 10, 10)
	if center[0] < 200 || center[1] > 50 || center[2] > 50 {
		t.Errorf("FillRectangle center should be red, got RGBA(%d,%d,%d,%d)",
			center[0], center[1], center[2], center[3])
	}

	// Outside the rectangle should remain blue.
	outside := getPixel(img.Data, stride, 2, 2)
	if outside[0] > 50 || outside[2] < 200 {
		t.Errorf("Pixel outside rectangle should remain blue, got RGBA(%d,%d,%d,%d)",
			outside[0], outside[1], outside[2], outside[3])
	}
}
